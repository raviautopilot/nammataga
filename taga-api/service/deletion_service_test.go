package service_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/email"
)

func setupTestEnvironment(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "deletion_test_*")
	require.NoError(t, err)

	// Subdirectories
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "member"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "towers"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "subscriptions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "payments"), 0755))

	origConfig := config.Config
	config.Config.MembersFile = filepath.Join(tmpDir, "data", "member", "members.json")
	config.Config.DeletedMembersFile = filepath.Join(tmpDir, "data", "member", "deleted_member.json")
	config.Config.DeletionRequestsFile = filepath.Join(tmpDir, "data", "member", "deletion_requests.json")
	config.Config.Data.Tower.BookingsFile = filepath.Join(tmpDir, "data", "towers", "bookings.json")

	// Initialize empty files
	require.NoError(t, os.WriteFile(config.Config.MembersFile, []byte("[]"), 0644))
	require.NoError(t, os.WriteFile(config.Config.DeletedMembersFile, []byte("[]"), 0644))
	require.NoError(t, os.WriteFile(config.Config.DeletionRequestsFile, []byte("[]"), 0644))
	require.NoError(t, os.WriteFile(config.Config.Data.Tower.BookingsFile, []byte("[]"), 0644))

	cleanup := func() {
		config.Config = origConfig
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestHasActiveStayAtTagaTowers(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)
	lastWeek := now.AddDate(0, 0, -7)

	bookings := []model.Booking{
		// 1. Confirmed active stay (yesterday to tomorrow) -> ACTIVE STAY
		{
			ID:            "BK001",
			BookerID:      "TAGA100",
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   yesterday,
			CheckOutDate:  tomorrow,
		},
		// 2. Pending active dates (yesterday to tomorrow) -> NOT confirmed -> NOT active stay
		{
			ID:            "BK002",
			BookerID:      "TAGA200",
			PaymentStatus: model.PaymentPending,
			CheckInDate:   yesterday,
			CheckOutDate:  tomorrow,
		},
		// 3. Confirmed completed stay (last week to yesterday) -> NOT active stay
		{
			ID:            "BK003",
			BookerID:      "TAGA300",
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   lastWeek,
			CheckOutDate:  yesterday,
		},
		// 4. Confirmed upcoming stay (tomorrow to next week) -> NOT active stay
		{
			ID:            "BK004",
			BookerID:      "TAGA400",
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   tomorrow,
			CheckOutDate:  nextWeek,
		},
	}

	bData, err := json.MarshalIndent(bookings, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(config.Config.Data.Tower.BookingsFile, bData, 0644))

	// Test TAGA100 has active stay
	hasStay, err := service.HasActiveStayAtTagaTowers("", "TAGA100")
	assert.NoError(t, err)
	assert.True(t, hasStay, "TAGA100 should have active stay")

	// Test TAGA200 has no active stay (pending payment)
	hasStay, err = service.HasActiveStayAtTagaTowers("", "TAGA200")
	assert.NoError(t, err)
	assert.False(t, hasStay, "TAGA200 should not have active stay (pending)")

	// Test TAGA300 has no active stay (completed)
	hasStay, err = service.HasActiveStayAtTagaTowers("", "TAGA300")
	assert.NoError(t, err)
	assert.False(t, hasStay, "TAGA300 should not have active stay (past)")

	// Test TAGA400 has no active stay (upcoming)
	hasStay, err = service.HasActiveStayAtTagaTowers("", "TAGA400")
	assert.NoError(t, err)
	assert.False(t, hasStay, "TAGA400 should not have active stay (upcoming)")
}

func TestExecuteAccountDeletion_ActiveStayBlocksDeletion(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	now := time.Now()
	activeBooking := []model.Booking{
		{
			ID:            "BK_ACTIVE",
			BookerID:      "TAGA001",
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   now.AddDate(0, 0, -1),
			CheckOutDate:  now.AddDate(0, 0, 2),
		},
	}
	bData, _ := json.Marshal(activeBooking)
	_ = os.WriteFile(config.Config.Data.Tower.BookingsFile, bData, 0644)

	member := map[string]interface{}{
		"id":       "user-uuid-1",
		"tagaId":   "TAGA001",
		"name":     "Active Stay Member",
		"emailId":  "stay@nammataga.com",
		"password": "hashed_secret_password",
	}
	mData, _ := json.Marshal([]map[string]interface{}{member})
	_ = os.WriteFile(config.Config.MembersFile, mData, 0644)

	// Attempt deletion -> MUST BE BLOCKED BY ACTIVE STAY
	err := service.ExecuteAccountDeletion(nil, "user-uuid-1", "TAGA001", "admin", "test", "")
	assert.Error(t, err)
	assert.Equal(t, service.ErrActiveStay, err)
	assert.Equal(t, service.ActiveStayMessage, err.Error())
}

func TestExecuteAccountDeletion_SecuritySecretsStrippedAndSubscriptionsCancelled(t *testing.T) {
	tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Setup member with secrets
	member := map[string]interface{}{
		"id":             "user-uuid-2",
		"tagaId":         "TAGA002",
		"name":           "Test Member",
		"emailId":        "test@nammataga.com",
		"password":       "$2a$10$encrypted_password",
		"salt":           "super_secret_salt",
		"token":          "jwt_token_secret",
		"otp":            "123456",
		"temp_password":  "tempPass123",
		"session_secret": "sessionKeySecret",
		"designation":    "Agriculture Officer",
	}
	mData, _ := json.Marshal([]map[string]interface{}{member})
	_ = os.WriteFile(config.Config.MembersFile, mData, 0644)

	// Setup bookings: one completed in past, one upcoming
	lastMonth := time.Now().AddDate(0, -1, 0)
	lastMonthEnd := time.Now().AddDate(0, -1, 5)
	tomorrow := time.Now().AddDate(0, 0, 1)
	nextWeek := time.Now().AddDate(0, 0, 5)
	booking := []model.Booking{
		{
			ID:            "BK_COMPLETED_PAST",
			BookerID:      "TAGA002",
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   lastMonth,
			CheckOutDate:  lastMonthEnd,
		},
		{
			ID:            "BK_UPCOMING",
			BookerID:      "TAGA002",
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   tomorrow,
			CheckOutDate:  nextWeek,
		},
	}
	bData, _ := json.Marshal(booking)
	_ = os.WriteFile(config.Config.Data.Tower.BookingsFile, bData, 0644)

	// Setup subscription
	subsFile := filepath.Join(tmpDir, "data", "subscriptions", "member_subscriptions.json")
	subs := []map[string]interface{}{
		{
			"id":           "sub-1",
			"member_id":    "TAGA002",
			"member_email": "test@nammataga.com",
			"status":       "active",
		},
	}
	sData, _ := json.Marshal(subs)
	_ = os.WriteFile(subsFile, sData, 0644)

	// Execute deletion
	err := service.ExecuteAccountDeletion(nil, "user-uuid-2", "TAGA002", "admin1", "Member requested deletion", "req-123")
	require.NoError(t, err)

	// 1. Verify removed from members.json
	currMembersData, _ := os.ReadFile(config.Config.MembersFile)
	var remainingMembers []map[string]interface{}
	_ = json.Unmarshal(currMembersData, &remainingMembers)
	assert.Empty(t, remainingMembers, "Member should be removed from members.json")

	// 2. Verify saved to deleted_member.json with stripped secrets
	deletedData, _ := os.ReadFile(config.Config.DeletedMembersFile)
	var deletedMembers []map[string]interface{}
	_ = json.Unmarshal(deletedData, &deletedMembers)
	require.Len(t, deletedMembers, 1)

	archived := deletedMembers[0]
	assert.Equal(t, "user-uuid-2", archived["id"])
	assert.Equal(t, "TAGA002", archived["tagaId"])
	assert.Equal(t, "admin1", archived["deleted_by"])
	assert.Equal(t, "Member requested deletion", archived["deletion_reason"])
	assert.NotEmpty(t, archived["deleted_at"])

	// CRITICAL SECURITY ASSERTIONS: Secrets MUST be stripped!
	assert.Nil(t, archived["password"], "password must be stripped from deleted_member.json")
	assert.Nil(t, archived["salt"], "salt must be stripped from deleted_member.json")
	assert.Nil(t, archived["token"], "token must be stripped from deleted_member.json")
	assert.Nil(t, archived["otp"], "otp must be stripped from deleted_member.json")
	assert.Nil(t, archived["temp_password"], "temp_password must be stripped from deleted_member.json")
	assert.Nil(t, archived["session_secret"], "session_secret must be stripped from deleted_member.json")

	// 3. Verify subscription is cancelled_by_account_deletion
	currSubsData, _ := os.ReadFile(subsFile)
	var updatedSubs []map[string]interface{}
	_ = json.Unmarshal(currSubsData, &updatedSubs)
	require.Len(t, updatedSubs, 1)
	assert.Equal(t, "cancelled_by_account_deletion", updatedSubs[0]["status"])

	// 4. Verify upcoming booking is cancelled and past booking preserved
	currBookingsData, _ := os.ReadFile(config.Config.Data.Tower.BookingsFile)
	var updatedBookings []model.Booking
	_ = json.Unmarshal(currBookingsData, &updatedBookings)
	require.Len(t, updatedBookings, 2)
	assert.Equal(t, model.PaymentConfirmed, updatedBookings[0].PaymentStatus, "Completed past booking must remain confirmed/preserved")
	assert.Equal(t, model.PaymentCancelled, updatedBookings[1].PaymentStatus, "Upcoming booking must be cancelled")
}

func TestResolveMemberUID_UIDReuse(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Seed deleted_member.json with an archived member
	archived := []map[string]interface{}{
		{
			"id":       "original-uuid-12345",
			"tagaId":   "TAGA555",
			"emailId":  "reuser@nammataga.com",
			"username": "reuser@nammataga.com",
		},
		{
			"id":       "conflict-uuid-1",
			"tagaId":   "TAGA666",
			"emailId":  "multi@nammataga.com",
			"username": "multi@nammataga.com",
		},
		{
			"id":       "conflict-uuid-2",
			"tagaId":   "TAGA666",
			"emailId":  "multi@nammataga.com",
			"username": "multi@nammataga.com",
		},
	}
	dData, _ := json.Marshal(archived)
	_ = os.WriteFile(config.Config.DeletedMembersFile, dData, 0644)

	activeMembers := []map[string]interface{}{
		{
			"id":     "active-uuid-999",
			"tagaId": "TAGA999",
			"email":  "active@nammataga.com",
		},
	}

	// Case 1: Both TAGA ID and email match exactly -> REUSE ORIGINAL UID
	reusedUID := service.ResolveMemberUID("TAGA555", "reuser@nammataga.com", activeMembers)
	assert.Equal(t, "original-uuid-12345", reusedUID, "Should reuse original UID when both TAGA ID and email match")

	// Case 2: Matching case-insensitive
	reusedUIDCase := service.ResolveMemberUID("taga555", "ReUser@Nammataga.Com", activeMembers)
	assert.Equal(t, "original-uuid-12345", reusedUIDCase, "Should reuse original UID case-insensitively")

	// Case 3: TAGA ID matches but email does NOT match -> FRESH UUID
	freshUID1 := service.ResolveMemberUID("TAGA555", "different@nammataga.com", activeMembers)
	assert.NotEqual(t, "original-uuid-12345", freshUID1)
	assert.NotEmpty(t, freshUID1)

	// Case 4: Email matches but TAGA ID does NOT match -> FRESH UUID
	freshUID2 := service.ResolveMemberUID("TAGA9999", "reuser@nammataga.com", activeMembers)
	assert.NotEqual(t, "original-uuid-12345", freshUID2)
	assert.NotEmpty(t, freshUID2)

	// Case 5: Multiple matches in archive -> LOGS WARNING & GENERATES FRESH UUID
	freshUIDMulti := service.ResolveMemberUID("TAGA666", "multi@nammataga.com", activeMembers)
	assert.NotEqual(t, "conflict-uuid-1", freshUIDMulti)
	assert.NotEqual(t, "conflict-uuid-2", freshUIDMulti)

	// Case 6: Candidate UID already exists in active members -> CONFLICT -> FRESH UUID
	conflictingActiveMembers := []map[string]interface{}{
		{
			"id": "original-uuid-12345",
		},
	}
	freshUIDConflict := service.ResolveMemberUID("TAGA555", "reuser@nammataga.com", conflictingActiveMembers)
	assert.NotEqual(t, "original-uuid-12345", freshUIDConflict)

	// Case 7: Neither matches -> FRESH UUID
	freshUIDNeither := service.ResolveMemberUID("TAGA888", "unrelated@nammataga.com", activeMembers)
	assert.NotEqual(t, "original-uuid-12345", freshUIDNeither)
	assert.NotEmpty(t, freshUIDNeither)
}

func TestDeletionRequests_CreateListReject(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// 1. Create a request
	req, err := service.CreateDeletionRequest(nil, "uuid-1", "TAGA001", "Member One", "one@nammataga.com", "Moving away")
	require.NoError(t, err)
	assert.Equal(t, model.DeletionStatusPending, req.Status)
	assert.NotEmpty(t, req.ID)

	// 2. Prevent duplicate pending request
	_, err = service.CreateDeletionRequest(nil, "uuid-1", "TAGA001", "Member One", "one@nammataga.com", "Duplicate")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "already pending"))

	// 3. List requests
	list, err := service.ListDeletionRequests("pending")
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// 4. Reject request without remarks -> FAIL
	_, err = service.RejectDeletionRequest(nil, req.ID, "admin", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remarks are required")

	// 5. Reject request with remarks -> SUCCESS
	rejected, err := service.RejectDeletionRequest(nil, req.ID, "admin", "Pending documentation required")
	require.NoError(t, err)
	assert.Equal(t, model.DeletionStatusRejected, rejected.Status)
	assert.Equal(t, "Pending documentation required", rejected.Remarks)

	// 6. Verify pending status is now cleared
	isPending := service.IsMemberDeletionPending("uuid-1", "one@nammataga.com")
	assert.False(t, isPending, "Member account should be unlocked (not pending)")
}

func TestDeletionRequests_MobileNumberAndSubscriptionEnrichment(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Seed members: one Paid, one Unpaid
	members := []map[string]interface{}{
		{
			"id":             "mem-paid-1",
			"tagaId":         "TAGAPAID1",
			"name":           "Paid Member",
			"emailId":        "paid@nammataga.com",
			"mobile_number":  "9876543210",
			"payment_status": "Paid",
		},
		{
			"id":             "mem-unpaid-2",
			"tagaId":         "TAGAUNPAID2",
			"name":           "Unpaid Member",
			"emailId":        "unpaid@nammataga.com",
			"mobile_number":  "9123456789",
			"payment_status": "Unpaid",
		},
	}
	mData, err := json.Marshal(members)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(config.Config.MembersFile, mData, 0644))

	// 1. Create request for Paid member without passing mobile explicitly
	req1, err := service.CreateDeletionRequest(nil, "mem-paid-1", "TAGAPAID1", "Paid Member", "paid@nammataga.com", "Moving abroad")
	require.NoError(t, err)
	assert.Equal(t, "9876543210", req1.MobileNumber, "Mobile number should be enriched from members.json on creation")
	assert.True(t, req1.IsPaid, "Should be marked as Paid")
	assert.True(t, req1.SubscriptionActive, "Subscription should be active")
	assert.Equal(t, "Paid", req1.SubscriptionStatus)

	// 2. Create request for Unpaid member with explicit mobile
	req2, err := service.CreateDeletionRequest(nil, "mem-unpaid-2", "TAGAUNPAID2", "Unpaid Member", "unpaid@nammataga.com", "Financial reasons", "9998887776")
	require.NoError(t, err)
	assert.Equal(t, "9998887776", req2.MobileNumber)
	assert.False(t, req2.IsPaid, "Should be marked as Unpaid")
	assert.False(t, req2.SubscriptionActive)
	assert.Equal(t, "Unpaid", req2.SubscriptionStatus)

	// 3. List requests and verify both are correctly enriched
	list, err := service.ListDeletionRequests("")
	require.NoError(t, err)
	require.Len(t, list, 2)

	for _, item := range list {
		if item.MemberID == "mem-paid-1" {
			assert.Equal(t, "9876543210", item.MobileNumber)
			assert.True(t, item.IsPaid)
			assert.True(t, item.SubscriptionActive)
			assert.Equal(t, "Paid", item.SubscriptionStatus)
		} else if item.MemberID == "mem-unpaid-2" {
			assert.Equal(t, "9998887776", item.MobileNumber)
			assert.False(t, item.IsPaid)
			assert.False(t, item.SubscriptionActive)
			assert.Equal(t, "Unpaid", item.SubscriptionStatus)
		}
	}
}

func TestDeletionRequests_EmailsDispatchedForAllSteps(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	config.Config.AdminEmail = "nammataga@gmail.com"
	config.Config.ResetPasswordURL = "https://www.nammataga.com"

	member := map[string]interface{}{
		"id":             "mem-email-test",
		"tagaId":         "TAGAEMAIL",
		"name":           "Email Tester",
		"emailId":        "emailtest@nammataga.com",
		"mobile_number":  "9876500000",
		"payment_status": "Paid",
	}
	mData, _ := json.Marshal([]map[string]interface{}{member})
	_ = os.WriteFile(config.Config.MembersFile, mData, 0644)

	email.ClearLastSentEmails()

	// 1. Step 1: Create Deletion Request -> Must send Ack to member and Alert to admin
	req, err := service.CreateDeletionRequest(nil, "mem-email-test", "TAGAEMAIL", "Email Tester", "emailtest@nammataga.com", "Retirement", "9876500000")
	require.NoError(t, err)

	// Wait briefly for asynchronous goroutines
	time.Sleep(150 * time.Millisecond)

	sentAfterCreate := email.GetLastSentEmails()
	require.GreaterOrEqual(t, len(sentAfterCreate), 2, "Both acknowledgement and admin alert emails should be dispatched")

	var memberAckFound, adminAlertFound bool
	for _, m := range sentAfterCreate {
		if m.To == "emailtest@nammataga.com" && strings.Contains(m.Subject, "Account Deletion Request Received") {
			memberAckFound = true
			assert.Contains(t, m.Body, "Email Tester")
			assert.Contains(t, m.Body, "TAGAEMAIL")
		}
		if m.To == "nammataga@gmail.com" && strings.Contains(m.Subject, "[ADMIN] New Account Deletion Request") {
			adminAlertFound = true
			assert.Contains(t, m.Body, "Email Tester")
			assert.Contains(t, m.Body, "9876500000")
		}
	}
	assert.True(t, memberAckFound, "Member acknowledgement email must be sent")
	assert.True(t, adminAlertFound, "Admin alert email must be sent")

	// 2. Step 2: Reject Deletion Request -> Must send rejection email to member with remarks
	email.ClearLastSentEmails()
	_, err = service.RejectDeletionRequest(nil, req.ID, "admin@nammataga.com", "Pending loan clearance needed")
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	sentAfterReject := email.GetLastSentEmails()
	require.Len(t, sentAfterReject, 1, "Rejection email should be dispatched to member")
	assert.Equal(t, "emailtest@nammataga.com", sentAfterReject[0].To)
	assert.Contains(t, sentAfterReject[0].Subject, "Account Deletion Request Update")
	assert.Contains(t, sentAfterReject[0].Body, "Pending loan clearance needed")
	assert.Contains(t, sentAfterReject[0].Body, "Your account remains fully active")

	// 3. Step 3: Approve Deletion Request (Create fresh request first)
	req2, err := service.CreateDeletionRequest(nil, "mem-email-test", "TAGAEMAIL", "Email Tester", "emailtest@nammataga.com", "Retirement 2", "9876500000")
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)
	email.ClearLastSentEmails()

	err = service.ExecuteAccountDeletion(nil, "mem-email-test", "TAGAEMAIL", "admin@nammataga.com", "Approved retirement", req2.ID)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	sentAfterApprove := email.GetLastSentEmails()
	require.Len(t, sentAfterApprove, 1, "Confirmation email should be dispatched to member on deletion approval")
	assert.Equal(t, "emailtest@nammataga.com", sentAfterApprove[0].To)
	assert.Contains(t, sentAfterApprove[0].Subject, "Account Deletion Confirmation")
	assert.Contains(t, sentAfterApprove[0].Body, "Summary of Actions Taken")
}
