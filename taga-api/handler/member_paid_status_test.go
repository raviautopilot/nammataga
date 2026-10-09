package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
	"taga-api/model"
)

func setupTestEnvironment(t *testing.T) (tempDir string, cleanup func()) {
	gin.SetMode(gin.TestMode)
	config.Init()

	tempDir, err := os.MkdirTemp("", "taga_paid_test_*")
	assert.NoError(t, err)

	membersDir := filepath.Join(tempDir, "member")
	subsDir := filepath.Join(tempDir, "subscriptions")
	os.MkdirAll(membersDir, 0755)
	os.MkdirAll(subsDir, 0755)

	membersFile := filepath.Join(membersDir, "members.json")

	oldMembersFile := config.Config.MembersFile
	config.Config.MembersFile = membersFile

	cleanup = func() {
		config.Config.MembersFile = oldMembersFile
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestCheckAnnualSubscriptionStatus_MatchesTagaIdAndUUIDAndEmail(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg := config.GetConfig()
	subsFile := filepath.Join(filepath.Dir(cfg.MembersFile), "..", "subscriptions", "member_subscriptions.json")

	// Member A: UUID uuidA, TAGA ID TAGA001, Email member1@test.com
	uuidA := uuid.New().String()
	memberA := map[string]interface{}{
		"id":                  uuidA,
		"tagaId":              "TAGA001",
		"emailId":             "member1@test.com",
		"name":                "Member One",
		"payment_status":      "Paid",
		"subscription_active": true,
	}

	// Member B: UUID uuidB, TAGA ID TAGA002, Email member2@test.com
	uuidB := uuid.New().String()
	memberB := map[string]interface{}{
		"id":                  uuidB,
		"tagaId":              "TAGA002",
		"emailId":             "member2@test.com",
		"name":                "Member Two",
		"payment_status":      "Paid",
		"subscription_active": true,
	}

	// Member C: Unpaid member
	uuidC := uuid.New().String()
	memberC := map[string]interface{}{
		"id":                  uuidC,
		"tagaId":              "TAGA003",
		"emailId":             "member3@test.com",
		"name":                "Member Three",
		"payment_status":      "Unpaid",
		"subscription_active": false,
	}

	membersData, _ := json.MarshalIndent([]map[string]interface{}{memberA, memberB, memberC}, "", "  ")
	os.WriteFile(cfg.MembersFile, membersData, 0644)

	// In member_subscriptions.json:
	// Sub 1: stored with TAGA ID "TAGA001"
	// Sub 2: stored with UUID uuidB (simulating old manual admin fake payment)
	now := time.Now()
	nextYearEnd := getMembershipYearEnd(now)
	subs := []model.MemberSubscription{
		{
			ID:               uuid.New().String(),
			MemberID:         "TAGA001",
			MemberEmail:      "member1@test.com",
			MemberName:       "Member One",
			SubscriptionID:   "annual-subscription",
			SubscriptionName: "Annual Subscription",
			Status:           "active",
			StartDate:        now,
			EndDate:          nextYearEnd,
		},
		{
			ID:               uuid.New().String(),
			MemberID:         uuidB, // UUID stored directly
			MemberEmail:      "MEMBER2@TEST.COM",
			MemberName:       "Member Two",
			SubscriptionID:   "annual-subscription",
			SubscriptionName: "Annual Subscription",
			Status:           "active",
			StartDate:        now,
			EndDate:          nextYearEnd,
		},
	}
	subsData, _ := json.MarshalIndent(subs, "", "  ")
	os.WriteFile(subsFile, subsData, 0644)

	// Member A (TAGA ID match)
	assert.True(t, checkAnnualSubscriptionStatus(uuidA), "Member A should be active via TAGA ID")

	// Member B (UUID / Email match for old manual payments)
	assert.True(t, checkAnnualSubscriptionStatus(uuidB), "Member B should be active via UUID / Email fallback")

	// Member C (Unpaid)
	assert.False(t, checkAnnualSubscriptionStatus(uuidC), "Member C should not be active")
}

func TestGetMemberProfileByToken_PreservesPaidStatusWithFallback(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg := config.GetConfig()

	// Member with Paid status in members.json, but NO record in member_subscriptions.json
	memberUUID := uuid.New().String()
	offlineMember := map[string]interface{}{
		"id":                  memberUUID,
		"tagaId":              "TAGA_OFFLINE",
		"emailId":             "offline@test.com",
		"name":                "Offline Paid Member",
		"payment_status":      "Paid",
		"subscription_active": true,
	}

	membersData, _ := json.MarshalIndent([]map[string]interface{}{offlineMember}, "", "  ")
	os.WriteFile(cfg.MembersFile, membersData, 0644)

	r := gin.New()
	r.GET("/api/member/profile", func(c *gin.Context) {
		c.Set("member_id", memberUUID)
		c.Set("member_email", "offline@test.com")
		c.Next()
	}, GetMemberProfileByToken)

	req, _ := http.NewRequest("GET", "/api/member/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)

	user, ok := res["user"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, user["isPaid"], "Profile endpoint must return isPaid: true for offline paid member")
	assert.Equal(t, true, user["subscription_active"], "Profile endpoint must return subscription_active: true for offline paid member")
}

func TestGetMemberPaidSubscriptions_WithEmailChangeAndOfflineStatus(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg := config.GetConfig()
	subsFile := filepath.Join(filepath.Dir(cfg.MembersFile), "..", "subscriptions", "member_subscriptions.json")

	// Case 1: Member originally had sudhanop04@gmail.com, then changed email to sudhanop05@gmail.com
	memberUUID := uuid.New().String()
	memberChanged := map[string]interface{}{
		"id":                  memberUUID,
		"tagaId":              "TAGA_007",
		"emailId":             "sudhanop05@gmail.com",
		"name":                "Sudhan",
		"payment_status":      "Paid",
		"subscription_active": true,
	}

	// Case 2: Offline member with Paid status but no record in subscriptions
	offlineUUID := uuid.New().String()
	offlineMember := map[string]interface{}{
		"id":                  offlineUUID,
		"tagaId":              "TAGA_OFFLINE_2",
		"emailId":             "offline2@test.com",
		"name":                "Offline Payer 2",
		"payment_status":      "Paid",
		"subscription_active": true,
	}

	// Case 3: Truly unpaid member
	unpaidUUID := uuid.New().String()
	unpaidMember := map[string]interface{}{
		"id":                  unpaidUUID,
		"tagaId":              "TAGA_UNPAID",
		"emailId":             "unpaid@test.com",
		"name":                "Unpaid User",
		"payment_status":      "Unpaid",
		"subscription_active": false,
	}

	membersData, _ := json.MarshalIndent([]map[string]interface{}{memberChanged, offlineMember, unpaidMember}, "", "  ")
	os.WriteFile(cfg.MembersFile, membersData, 0644)

	// Save subscription with old email sudhanop04@gmail.com, but member_id = TAGA_007
	now := time.Now()
	nextYearEnd := getMembershipYearEnd(now)
	subs := []model.MemberSubscription{
		{
			ID:               uuid.New().String(),
			MemberID:         "TAGA_007",
			MemberEmail:      "sudhanop04@gmail.com",
			MemberName:       "Sudhan",
			SubscriptionID:   "annual-subscription",
			SubscriptionName: "Annual Subscription",
			Status:           "active",
			StartDate:        now,
			EndDate:          nextYearEnd,
		},
	}
	subsData, _ := json.MarshalIndent(subs, "", "  ")
	os.WriteFile(subsFile, subsData, 0644)

	r := gin.New()
	r.GET("/api/subscriptions/member-paid", GetMemberPaidSubscriptions)

	// Test 1: Query with new email sudhanop05@gmail.com -> MUST return annual-subscription
	req1, _ := http.NewRequest("GET", "/api/subscriptions/member-paid?email=sudhanop05@gmail.com", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	var paidIDs1 []string
	json.Unmarshal(w1.Body.Bytes(), &paidIDs1)
	assert.Contains(t, paidIDs1, "annual-subscription", "Member with changed email must be recognized as paid")

	// Test 2: Query offline member -> MUST return annual-subscription
	req2, _ := http.NewRequest("GET", "/api/subscriptions/member-paid?email=offline2@test.com", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var paidIDs2 []string
	json.Unmarshal(w2.Body.Bytes(), &paidIDs2)
	assert.Contains(t, paidIDs2, "annual-subscription", "Offline paid member must be recognized as paid")

	// Test 3: Query unpaid member -> MUST NOT contain annual-subscription
	req3, _ := http.NewRequest("GET", "/api/subscriptions/member-paid?email=unpaid@test.com", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	var paidIDs3 []string
	json.Unmarshal(w3.Body.Bytes(), &paidIDs3)
	assert.NotContains(t, paidIDs3, "annual-subscription", "Unpaid member must NOT have annual-subscription")
}

func TestSyncMemberEmailInSubscriptions(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg := config.GetConfig()
	subsFile := filepath.Join(filepath.Dir(cfg.MembersFile), "..", "subscriptions", "member_subscriptions.json")

	memberUUID := uuid.New().String()
	tagaID := "TAGA_SYNC_TEST"
	oldEmail := "oldemail@test.com"
	newEmail := "newemail@test.com"

	now := time.Now()
	nextYearEnd := getMembershipYearEnd(now)
	subs := []model.MemberSubscription{
		{
			ID:               uuid.New().String(),
			MemberID:         tagaID,
			MemberEmail:      oldEmail,
			MemberName:       "Sync Tester",
			SubscriptionID:   "annual-subscription",
			SubscriptionName: "Annual Subscription",
			Status:           "active",
			StartDate:        now,
			EndDate:          nextYearEnd,
		},
	}
	subsData, _ := json.MarshalIndent(subs, "", "  ")
	os.WriteFile(subsFile, subsData, 0644)

	// Call sync helper
	syncMemberEmailInSubscriptions(memberUUID, tagaID, oldEmail, newEmail)

	// Verify file on disk
	data, err := os.ReadFile(subsFile)
	assert.NoError(t, err)

	var updatedSubs []model.MemberSubscription
	err = json.Unmarshal(data, &updatedSubs)
	assert.NoError(t, err)
	assert.Len(t, updatedSubs, 1)
	assert.Equal(t, newEmail, updatedSubs[0].MemberEmail, "Subscription email must be updated to new email")
}

func TestCreateSubscriptionOrder_NegativeOrZeroAmount(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Send negative amount payload
	c.Request = httptest.NewRequest("POST", "/api/subscriptions/create-order", strings.NewReader(`{"subscriptionId":"donation","amount":-500,"email":"test@nammataga.com"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateSubscriptionOrder(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "Payment amount must be greater than zero")
}

