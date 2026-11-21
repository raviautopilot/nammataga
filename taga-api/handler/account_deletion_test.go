package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"taga-api/config"
	"taga-api/model"
	"taga-api/router"
	"taga-api/service"
	"taga-api/service/jwt"
)

func setupHandlerTestEnv(t *testing.T) (string, func()) {
	gin.SetMode(gin.TestMode)
	tmpDir, err := os.MkdirTemp("", "handler_test_*")
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "member"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "towers"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "subscriptions"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "data", "payments"), 0755))

	origConfig := config.Config
	config.Config.MembersFile = filepath.Join(tmpDir, "data", "member", "members.json")
	config.Config.DeletedMembersFile = filepath.Join(tmpDir, "data", "member", "deleted_member.json")
	config.Config.DeletionRequestsFile = filepath.Join(tmpDir, "data", "member", "deletion_requests.json")
	config.Config.Data.Tower.BookingsFile = filepath.Join(tmpDir, "data", "towers", "bookings.json")
	config.Config.JwtSecret = "test-jwt-secret-key-1234567890"

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

func setupTestRouter() *gin.Engine {
	r := gin.New()
	router.RegisterMemberRoutes(r)
	router.RegisterAdminRoutes(r)
	return r
}

func createTestMember(t *testing.T, id, tagaID, email, rawPassword string) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	member := map[string]interface{}{
		"id":            id,
		"tagaId":        tagaID,
		"emailId":       email,
		"username":      email,
		"name":          "Test User",
		"password":      string(hashed),
		"first_login":   false,
		"payment_status": "Paid",
	}

	var members []map[string]interface{}
	data, _ := os.ReadFile(config.Config.MembersFile)
	if len(data) > 0 {
		_ = json.Unmarshal(data, &members)
	}
	members = append(members, member)

	mBytes, _ := json.MarshalIndent(members, "", "  ")
	require.NoError(t, os.WriteFile(config.Config.MembersFile, mBytes, 0644))
}

func TestAccountDeletion_EligibilityAndRequest_PasswordAndActiveStayValidation(t *testing.T) {
	_, cleanup := setupHandlerTestEnv(t)
	defer cleanup()

	r := setupTestRouter()

	memberID := "mem-uuid-100"
	tagaID := "TAGA100"
	email := "member100@nammataga.com"
	rawPassword := "SecurePass123!"

	createTestMember(t, memberID, tagaID, email, rawPassword)

	token, _, err := jwt.GenerateMemberToken(memberID, email, "Test User")
	require.NoError(t, err)

	// 1. Check eligibility when no active stay -> ELIGIBLE
	reqEl, _ := http.NewRequest(http.MethodGet, "/api/member/account-deletion/eligibility", nil)
	reqEl.Header.Set("Authorization", "Bearer "+token)
	wEl := httptest.NewRecorder()
	r.ServeHTTP(wEl, reqEl)
	assert.Equal(t, http.StatusOK, wEl.Code)

	var respEl model.AccountDeletionEligibilityResponse
	_ = json.Unmarshal(wEl.Body.Bytes(), &respEl)
	assert.True(t, respEl.Eligible)

	// 2. Request deletion with wrong password -> 400 Bad Request
	wrongPayload := map[string]string{"password": "wrongpassword"}
	bPayload, _ := json.Marshal(wrongPayload)
	reqWrong, _ := http.NewRequest(http.MethodPost, "/api/member/account-deletion", bytes.NewBuffer(bPayload))
	reqWrong.Header.Set("Authorization", "Bearer "+token)
	reqWrong.Header.Set("Content-Type", "application/json")
	wWrong := httptest.NewRecorder()
	r.ServeHTTP(wWrong, reqWrong)
	assert.Equal(t, http.StatusBadRequest, wWrong.Code)

	// 3. Set up active stay at TAGA Towers
	now := time.Now()
	activeBooking := []model.Booking{
		{
			ID:            "BK_ACTIVE_STAY",
			BookerID:      tagaID,
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   now.AddDate(0, 0, -1),
			CheckOutDate:  now.AddDate(0, 0, 1),
		},
	}
	bData, _ := json.Marshal(activeBooking)
	require.NoError(t, os.WriteFile(config.Config.Data.Tower.BookingsFile, bData, 0644))

	// 3a. Check eligibility now -> NOT ELIGIBLE due to active stay
	wEl2 := httptest.NewRecorder()
	r.ServeHTTP(wEl2, reqEl)
	assert.Equal(t, http.StatusOK, wEl2.Code)
	var respEl2 model.AccountDeletionEligibilityResponse
	_ = json.Unmarshal(wEl2.Body.Bytes(), &respEl2)
	assert.False(t, respEl2.Eligible)
	assert.Equal(t, "active_stay", respEl2.Reason)
	assert.Equal(t, "You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.", respEl2.Message)

	// 3b. Request deletion with correct password while active stay exists -> MUST BE BLOCKED
	validPayload := map[string]string{
		"password": rawPassword,
		"reason":   "Relocating to another state",
	}
	bValid, _ := json.Marshal(validPayload)
	reqBlocked, _ := http.NewRequest(http.MethodPost, "/api/member/account-deletion", bytes.NewBuffer(bValid))
	reqBlocked.Header.Set("Authorization", "Bearer "+token)
	reqBlocked.Header.Set("Content-Type", "application/json")
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)
	assert.Equal(t, http.StatusBadRequest, wBlocked.Code)
	var blockedMap map[string]interface{}
	_ = json.Unmarshal(wBlocked.Body.Bytes(), &blockedMap)
	assert.Equal(t, "You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.", blockedMap["error"])

	// 4. Clear active stay (checkout completed)
	require.NoError(t, os.WriteFile(config.Config.Data.Tower.BookingsFile, []byte("[]"), 0644))

	// 5. Submit deletion request with valid password -> SUCCESS
	reqSuccess, _ := http.NewRequest(http.MethodPost, "/api/member/account-deletion", bytes.NewBuffer(bValid))
	reqSuccess.Header.Set("Authorization", "Bearer "+token)
	reqSuccess.Header.Set("Content-Type", "application/json")
	wSuccess := httptest.NewRecorder()
	r.ServeHTTP(wSuccess, reqSuccess)
	assert.Equal(t, http.StatusOK, wSuccess.Code)

	var successMap map[string]interface{}
	_ = json.Unmarshal(wSuccess.Body.Bytes(), &successMap)
	assert.Equal(t, "pending", successMap["status"])
	requestID := successMap["request_id"].(string)
	assert.NotEmpty(t, requestID)

	// 6. Check status endpoint -> Returns pending
	reqStatus, _ := http.NewRequest(http.MethodGet, "/api/member/account-deletion/status", nil)
	reqStatus.Header.Set("Authorization", "Bearer "+token)
	wStatus := httptest.NewRecorder()
	r.ServeHTTP(wStatus, reqStatus)
	assert.Equal(t, http.StatusOK, wStatus.Code)
	var statusMap map[string]interface{}
	_ = json.Unmarshal(wStatus.Body.Bytes(), &statusMap)
	assert.Equal(t, "pending", statusMap["status"])

	// 7. Attempt login as pending member:
	// Returns 403 Forbidden with exact professional message!
	loginPayload := map[string]string{
		"email":    email,
		"password": rawPassword,
	}
	bLogin, _ := json.Marshal(loginPayload)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/api/member/login", bytes.NewBuffer(bLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	r.ServeHTTP(wLogin, reqLogin)
	assert.Equal(t, http.StatusForbidden, wLogin.Code)

	var loginErrMap map[string]interface{}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginErrMap)
	assert.Equal(t, service.PendingAccountMessage, loginErrMap["error"])

	// 8. Attempt login as pending member with INVALID password -> Also returns 403 Forbidden because deletion review blocks login!
	loginWrongPayload := map[string]string{
		"email":    email,
		"password": "wrongpassword123",
	}
	bLoginWrong, _ := json.Marshal(loginWrongPayload)
	reqLoginWrong, _ := http.NewRequest(http.MethodPost, "/api/member/login", bytes.NewBuffer(bLoginWrong))
	reqLoginWrong.Header.Set("Content-Type", "application/json")
	wLoginWrong := httptest.NewRecorder()
	r.ServeHTTP(wLoginWrong, reqLoginWrong)
	assert.Equal(t, http.StatusForbidden, wLoginWrong.Code)
}

func TestAdminDirectDeletion_ActiveStayCheck(t *testing.T) {
	_, cleanup := setupHandlerTestEnv(t)
	defer cleanup()

	r := setupTestRouter()

	memberID := "mem-uuid-200"
	tagaID := "TAGA200"
	email := "member200@nammataga.com"

	createTestMember(t, memberID, tagaID, email, "Password123!")

	// Generate admin token
	adminToken, _, err := jwt.GenerateAdminToken("admin@nammataga.com")
	require.NoError(t, err)

	// Set active stay for TAGA200
	now := time.Now()
	activeBooking := []model.Booking{
		{
			ID:            "BK_ACTIVE_STAY_ADMIN",
			BookerID:      tagaID,
			PaymentStatus: model.PaymentConfirmed,
			CheckInDate:   now.AddDate(0, 0, -1),
			CheckOutDate:  now.AddDate(0, 0, 2),
		},
	}
	bData, _ := json.Marshal(activeBooking)
	require.NoError(t, os.WriteFile(config.Config.Data.Tower.BookingsFile, bData, 0644))

	// Admin attempts DELETE /api/admin/members/:id -> MUST BE BLOCKED BY ACTIVE STAY
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/admin/members/"+memberID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+adminToken)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)

	assert.Equal(t, http.StatusBadRequest, wDel.Code)
	var delResp map[string]interface{}
	_ = json.Unmarshal(wDel.Body.Bytes(), &delResp)
	assert.Equal(t, "You are currently staying at TAGA Towers. You can request account deletion after completing your checkout.", delResp["error"])

	// Clear active stay
	require.NoError(t, os.WriteFile(config.Config.Data.Tower.BookingsFile, []byte("[]"), 0644))

	// Admin executes DELETE /api/admin/members/:id -> SUCCESS
	wDelSuccess := httptest.NewRecorder()
	r.ServeHTTP(wDelSuccess, reqDel)
	assert.Equal(t, http.StatusOK, wDelSuccess.Code)
}

func TestAdminReview_ListApproveReject(t *testing.T) {
	_, cleanup := setupHandlerTestEnv(t)
	defer cleanup()

	r := setupTestRouter()

	memberID := "mem-uuid-300"
	tagaID := "TAGA300"
	email := "member300@nammataga.com"
	rawPassword := "Password300!"

	createTestMember(t, memberID, tagaID, email, rawPassword)

	memberToken, _, err := jwt.GenerateMemberToken(memberID, email, "Member 300")
	require.NoError(t, err)

	adminToken, _, err := jwt.GenerateAdminToken("admin@nammataga.com")
	require.NoError(t, err)

	// Member submits deletion request
	subPayload := map[string]string{
		"password": rawPassword,
		"reason":   "Retiring from service",
	}
	bSub, _ := json.Marshal(subPayload)
	reqSub, _ := http.NewRequest(http.MethodPost, "/api/member/account-deletion", bytes.NewBuffer(bSub))
	reqSub.Header.Set("Authorization", "Bearer "+memberToken)
	reqSub.Header.Set("Content-Type", "application/json")
	wSub := httptest.NewRecorder()
	r.ServeHTTP(wSub, reqSub)
	assert.Equal(t, http.StatusOK, wSub.Code)

	var subResp map[string]interface{}
	_ = json.Unmarshal(wSub.Body.Bytes(), &subResp)
	requestID := subResp["request_id"].(string)

	// 1. Admin lists deletion requests
	reqList, _ := http.NewRequest(http.MethodGet, "/api/admin/deletion-requests?status=pending", nil)
	reqList.Header.Set("Authorization", "Bearer "+adminToken)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	assert.Equal(t, http.StatusOK, wList.Code)

	var listResp map[string]interface{}
	_ = json.Unmarshal(wList.Body.Bytes(), &listResp)
	requestsList := listResp["requests"].([]interface{})
	assert.Len(t, requestsList, 1)

	// 2. Admin rejects without remarks -> 400 Bad Request
	emptyRemarks := map[string]string{"remarks": "   "}
	bEmpty, _ := json.Marshal(emptyRemarks)
	reqRejectEmpty, _ := http.NewRequest(http.MethodPost, "/api/admin/deletion-requests/"+requestID+"/reject", bytes.NewBuffer(bEmpty))
	reqRejectEmpty.Header.Set("Authorization", "Bearer "+adminToken)
	reqRejectEmpty.Header.Set("Content-Type", "application/json")
	wRejectEmpty := httptest.NewRecorder()
	r.ServeHTTP(wRejectEmpty, reqRejectEmpty)
	assert.Equal(t, http.StatusBadRequest, wRejectEmpty.Code)

	// 3. Admin rejects with valid remarks -> SUCCESS (Unlocks account)
	validRemarks := map[string]string{"remarks": "Please settle pending welfare society dues before deleting account."}
	bValidRem, _ := json.Marshal(validRemarks)
	reqRejectValid, _ := http.NewRequest(http.MethodPost, "/api/admin/deletion-requests/"+requestID+"/reject", bytes.NewBuffer(bValidRem))
	reqRejectValid.Header.Set("Authorization", "Bearer "+adminToken)
	reqRejectValid.Header.Set("Content-Type", "application/json")
	wRejectValid := httptest.NewRecorder()
	r.ServeHTTP(wRejectValid, reqRejectValid)
	assert.Equal(t, http.StatusOK, wRejectValid.Code)

	// 4. Verify member can now log in normally because account is unlocked!
	loginPayload := map[string]string{
		"email":    email,
		"password": rawPassword,
	}
	bLogin, _ := json.Marshal(loginPayload)
	reqLogin, _ := http.NewRequest(http.MethodPost, "/api/member/login", bytes.NewBuffer(bLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	r.ServeHTTP(wLogin, reqLogin)
	assert.Equal(t, http.StatusOK, wLogin.Code)

	// 5. Submit another request to test Approve flow
	reqSub2, _ := http.NewRequest(http.MethodPost, "/api/member/account-deletion", bytes.NewBuffer(bSub))
	reqSub2.Header.Set("Authorization", "Bearer "+memberToken)
	reqSub2.Header.Set("Content-Type", "application/json")
	wSub2 := httptest.NewRecorder()
	r.ServeHTTP(wSub2, reqSub2)
	assert.Equal(t, http.StatusOK, wSub2.Code)

	var subResp2 map[string]interface{}
	_ = json.Unmarshal(wSub2.Body.Bytes(), &subResp2)
	requestID2 := subResp2["request_id"].(string)

	// 6. Admin approves request -> SUCCESS
	reqApprove, _ := http.NewRequest(http.MethodPost, "/api/admin/deletion-requests/"+requestID2+"/approve", nil)
	reqApprove.Header.Set("Authorization", "Bearer "+adminToken)
	wApprove := httptest.NewRecorder()
	r.ServeHTTP(wApprove, reqApprove)
	assert.Equal(t, http.StatusOK, wApprove.Code)

	// 7. Verify member can no longer log in (Member not found / invalid credentials)
	wLoginAfterDel := httptest.NewRecorder()
	reqLoginAfterDel, _ := http.NewRequest(http.MethodPost, "/api/member/login", bytes.NewBuffer(bLogin))
	reqLoginAfterDel.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wLoginAfterDel, reqLoginAfterDel)
	assert.Equal(t, http.StatusUnauthorized, wLoginAfterDel.Code)
}

func TestMemberAuthMiddleware_PendingAndDeletedRestrictions(t *testing.T) {
	_, cleanup := setupHandlerTestEnv(t)
	defer cleanup()

	r := setupTestRouter()

	memberID := "mem-uuid-400"
	tagaID := "TAGA400"
	email := "member400@nammataga.com"
	rawPassword := "Password400!"

	createTestMember(t, memberID, tagaID, email, rawPassword)

	token, _, err := jwt.GenerateMemberToken(memberID, email, "Member 400")
	require.NoError(t, err)

	// Accessing profile before deletion request -> 200 OK
	reqProf, _ := http.NewRequest(http.MethodGet, "/api/member/profile", nil)
	reqProf.Header.Set("Authorization", "Bearer "+token)
	wProf := httptest.NewRecorder()
	r.ServeHTTP(wProf, reqProf)
	assert.Equal(t, http.StatusOK, wProf.Code)

	// Submit deletion request
	subPayload := map[string]string{
		"password": rawPassword,
		"reason":   "Leaving",
	}
	bSub, _ := json.Marshal(subPayload)
	reqSub, _ := http.NewRequest(http.MethodPost, "/api/member/account-deletion", bytes.NewBuffer(bSub))
	reqSub.Header.Set("Authorization", "Bearer "+token)
	reqSub.Header.Set("Content-Type", "application/json")
	wSub := httptest.NewRecorder()
	r.ServeHTTP(wSub, reqSub)
	assert.Equal(t, http.StatusOK, wSub.Code)

	// Now member is PENDING
	// Accessing /api/member/profile -> MUST ABORT WITH 403 FORBIDDEN and exact message
	wProfPending := httptest.NewRecorder()
	r.ServeHTTP(wProfPending, reqProf)
	assert.Equal(t, http.StatusForbidden, wProfPending.Code)
	var profErrMap map[string]interface{}
	_ = json.Unmarshal(wProfPending.Body.Bytes(), &profErrMap)
	exactMsg := "Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending."
	assert.Equal(t, exactMsg, profErrMap["error"])

	// But accessing /api/member/account-deletion/status -> ALLOWED (200 OK)
	reqStat, _ := http.NewRequest(http.MethodGet, "/api/member/account-deletion/status", nil)
	reqStat.Header.Set("Authorization", "Bearer "+token)
	wStat := httptest.NewRecorder()
	r.ServeHTTP(wStat, reqStat)
	assert.Equal(t, http.StatusOK, wStat.Code)

	// Admin deletes the member
	adminToken, _, _ := jwt.GenerateAdminToken("admin@nammataga.com")
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/admin/members/"+memberID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+adminToken)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	assert.Equal(t, http.StatusOK, wDel.Code)

	// Member is now DELETED
	// Token is still valid cryptographically, but account is deleted -> MUST ABORT WITH 401 UNAUTHORIZED
	wProfDeleted := httptest.NewRecorder()
	r.ServeHTTP(wProfDeleted, reqProf)
	assert.Equal(t, http.StatusUnauthorized, wProfDeleted.Code)
	var delErrMap map[string]interface{}
	_ = json.Unmarshal(wProfDeleted.Body.Bytes(), &delErrMap)
	assert.Equal(t, "Account has been deleted", delErrMap["error"])
}

func TestUIDReuse_DuringAddMember(t *testing.T) {
	_, cleanup := setupHandlerTestEnv(t)
	defer cleanup()

	r := setupTestRouter()
	adminToken, _, _ := jwt.GenerateAdminToken("admin@nammataga.com")

	// Seed deleted_member.json with an archived member
	archived := []map[string]interface{}{
		{
			"id":       "archived-uid-777",
			"tagaId":   "TAGA777",
			"emailId":  "reuseme@nammataga.com",
			"username": "reuseme@nammataga.com",
		},
	}
	dData, _ := json.Marshal(archived)
	require.NoError(t, os.WriteFile(config.Config.DeletedMembersFile, dData, 0644))

	// Admin registers new member with SAME TAGA ID and SAME Email
	addReq := map[string]interface{}{
		"tagaId":                   "TAGA777",
		"name":                     "Reused User",
		"initial":                  "R",
		"gender":                   "Male",
		"educationalQualification": "B.Sc",
		"designation":              "Officer",
		"workingDistrict":          "Chennai",
		"nativeDistrict":           "Salem",
		"dateOfBirth":              "1990-01-01",
		"mobileNumber":             "9876543210",
		"email":                    "reuseme@nammataga.com",
		"paymentStatus":            "Paid",
	}
	bAdd, _ := json.Marshal(addReq)
	reqAdd, _ := http.NewRequest(http.MethodPost, "/api/admin/members/add", bytes.NewBuffer(bAdd))
	reqAdd.Header.Set("Authorization", "Bearer "+adminToken)
	reqAdd.Header.Set("Content-Type", "application/json")
	wAdd := httptest.NewRecorder()
	r.ServeHTTP(wAdd, reqAdd)
	assert.Equal(t, http.StatusOK, wAdd.Code)

	// Check members.json: member UID must be "archived-uid-777"!
	mData, _ := os.ReadFile(config.Config.MembersFile)
	var members []map[string]interface{}
	_ = json.Unmarshal(mData, &members)
	require.Len(t, members, 1)
	assert.Equal(t, "archived-uid-777", members[0]["id"], "Should reuse original archived UID")
}

func TestAccountingFix_PaymentRecordsPreservedOnDeletion(t *testing.T) {
	tmpDir, cleanup := setupHandlerTestEnv(t)
	defer cleanup()

	r := setupTestRouter()
	adminToken, _, _ := jwt.GenerateAdminToken("admin@nammataga.com")

	memberID := "mem-uuid-500"
	tagaID := "TAGA500"
	email := "paytest@nammataga.com"

	createTestMember(t, memberID, tagaID, email, "Password500!")

	// Create payment transactions
	txnsPath := filepath.Join("data", "subscriptions", "payment_transactions.json")
	_ = os.MkdirAll(filepath.Dir(txnsPath), 0755)
	txns := []map[string]interface{}{
		{
			"id":        "txn-001",
			"member_id": tagaID,
			"amount":    50000,
		},
	}
	tData, _ := json.Marshal(txns)
	_ = os.WriteFile(txnsPath, tData, 0644)

	// Create processed payments
	processedPath := config.Config.ProcessedPaymentsFile
	_ = os.MkdirAll(filepath.Dir(processedPath), 0755)
	processed := []map[string]interface{}{
		{
			"payment_id":   "pay_123",
			"member_email": email,
			"amount":       500,
		},
	}
	pData, _ := json.Marshal(processed)
	_ = os.WriteFile(processedPath, pData, 0644)

	// Admin deletes member
	reqDel, _ := http.NewRequest(http.MethodDelete, "/api/admin/members/"+memberID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+adminToken)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	assert.Equal(t, http.StatusOK, wDel.Code)

	// CRITICAL ACCOUNTING FIX: Verify payment records are NOT deleted!
	currTxnsData, _ := os.ReadFile(txnsPath)
	var remainingTxns []map[string]interface{}
	_ = json.Unmarshal(currTxnsData, &remainingTxns)
	assert.Len(t, remainingTxns, 1, "Payment transactions must NOT be deleted")

	currProcData, _ := os.ReadFile(processedPath)
	var remainingProc []map[string]interface{}
	_ = json.Unmarshal(currProcData, &remainingProc)
	assert.Len(t, remainingProc, 1, "Processed payments must NOT be deleted")

	_ = tmpDir
}
