package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
)

func TestOfflineSubscriptionValidity_ExpiryAndStatusResolution(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// 1. Member A: Paid for current year (2026-2027, valid until 2027-03-31)
	memberA := map[string]interface{}{
		"id":                       uuid.New().String(),
		"tagaId":                   "TAGA_ACTIVE",
		"emailId":                  "active@taga.org",
		"name":                     "Active Member",
		"payment_status":           "Paid",
		"subscription_active":      true,
		"subscription_year":        "2026-2027",
		"subscription_valid_until": "2027-03-31T23:59:59Z",
	}

	// 2. Member B: Paid for old year (2024-2025, expired on 2025-03-31)
	memberB := map[string]interface{}{
		"id":                       uuid.New().String(),
		"tagaId":                   "TAGA_EXPIRED",
		"emailId":                  "expired@taga.org",
		"name":                     "Expired Member",
		"payment_status":           "Paid",
		"subscription_active":      true,
		"subscription_year":        "2024-2025",
		"subscription_valid_until": "2025-03-31T23:59:59Z",
	}

	// 3. Member C: Paid for 2025-2026 (ended 2026-03-31, so expired as of Nov 2026)
	memberC := map[string]interface{}{
		"id":                       uuid.New().String(),
		"tagaId":                   "TAGA_EXPIRED_2",
		"emailId":                  "expired2@taga.org",
		"name":                     "Expired Member 2",
		"payment_status":           "Paid",
		"subscription_active":      true,
		"subscription_year":        "2025-2026",
		"subscription_valid_until": "2026-03-31T23:59:59Z",
	}

	// 4. Member D: Unpaid member
	memberD := map[string]interface{}{
		"id":                  uuid.New().String(),
		"tagaId":              "TAGA_UNPAID",
		"emailId":             "unpaid@taga.org",
		"name":                "Unpaid Member",
		"payment_status":      "Unpaid",
		"subscription_active": false,
	}

	emptySubMap := make(map[string]bool)

	// Test Status Resolution:
	// Member A -> "Paid" and "Active"
	assert.Equal(t, "Paid", getPaymentStatusFromMember(memberA, emptySubMap))
	assert.Equal(t, "Active", getMembershipStatus(memberA, emptySubMap))

	// Member B -> "Expired" and "Inactive"
	assert.Equal(t, "Expired", getPaymentStatusFromMember(memberB, emptySubMap))
	assert.Equal(t, "Inactive", getMembershipStatus(memberB, emptySubMap))

	// Member C -> "Expired" and "Inactive"
	assert.Equal(t, "Expired", getPaymentStatusFromMember(memberC, emptySubMap))
	assert.Equal(t, "Inactive", getMembershipStatus(memberC, emptySubMap))

	// Member D -> "Unpaid" and "Inactive"
	assert.Equal(t, "Unpaid", getPaymentStatusFromMember(memberD, emptySubMap))
	assert.Equal(t, "Inactive", getMembershipStatus(memberD, emptySubMap))
}

func TestOfflineExpiredMember_WithOnlineRenewal_BecomesPaid(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Member with expired offline status in members.json, but has active subscription in subscriptionMap
	member := map[string]interface{}{
		"id":                       uuid.New().String(),
		"tagaId":                   "TAGA_RENEWED",
		"emailId":                  "renewed@taga.org",
		"name":                     "Renewed Member",
		"payment_status":           "Paid",
		"subscription_active":      false,
		"subscription_year":        "2024-2025",
		"subscription_valid_until": "2025-03-31T23:59:59Z",
	}

	subMap := map[string]bool{
		"renewed@taga.org": true,
	}

	assert.Equal(t, "Paid", getPaymentStatusFromMember(member, subMap))
	assert.Equal(t, "Active", getMembershipStatus(member, subMap))
}

func TestAdminUpdateMember_TogglesPaymentStatusAndSetsYear(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg := config.GetConfig()
	memberID := uuid.New().String()
	initialMember := map[string]interface{}{
		"id":                  memberID,
		"tagaId":              "TAGA_TOGGLE",
		"emailId":             "toggle@taga.org",
		"name":                "Toggle Member",
		"payment_status":      "Unpaid",
		"subscription_active": false,
	}

	initialData, _ := json.MarshalIndent([]map[string]interface{}{initialMember}, "", "  ")
	os.WriteFile(cfg.MembersFile, initialData, 0644)

	r := gin.New()
	r.PUT("/api/admin/members/:id", UpdateMember)

	// Admin sets member to "Paid"
	updatePayload := map[string]interface{}{
		"payment_status": "Paid",
	}
	bodyBytes, _ := json.Marshal(updatePayload)
	req, _ := http.NewRequest("PUT", "/api/admin/members/"+memberID, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Read back members.json and verify subscription_year and subscription_valid_until
	data, err := os.ReadFile(cfg.MembersFile)
	assert.NoError(t, err)

	var members []map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &members))
	assert.Len(t, members, 1)

	updated := members[0]
	assert.Equal(t, "Paid", updated["payment_status"])
	assert.Equal(t, true, updated["subscription_active"])
	assert.NotEmpty(t, updated["subscription_year"], "subscription_year should be recorded")
	assert.NotEmpty(t, updated["subscription_valid_until"], "subscription_valid_until should be recorded")

	// Now admin toggles back to "Unpaid"
	unpaidPayload := map[string]interface{}{
		"payment_status": "Unpaid",
	}
	unpaidBytes, _ := json.Marshal(unpaidPayload)
	req2, _ := http.NewRequest("PUT", "/api/admin/members/"+memberID, bytes.NewBuffer(unpaidBytes))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	data2, err := os.ReadFile(cfg.MembersFile)
	assert.NoError(t, err)
	var members2 []map[string]interface{}
	assert.NoError(t, json.Unmarshal(data2, &members2))
	assert.Equal(t, "Unpaid", members2[0]["payment_status"])
	assert.Equal(t, false, members2[0]["subscription_active"])
}

func TestExpiredMemberProfileAndPaidSubscriptions_EvaluatesCorrectly(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg := config.GetConfig()

	// Member with expired offline payment
	expiredUUID := uuid.New().String()
	expiredMember := map[string]interface{}{
		"id":                       expiredUUID,
		"tagaId":                   "TAGA_EXPIRED_MEMBER",
		"emailId":                  "expired_profile@test.com",
		"name":                     "Expired Profile User",
		"payment_status":           "Paid",
		"subscription_active":      true,
		"subscription_year":        "2024-2025",
		"subscription_valid_until": "2025-03-31T23:59:59Z",
	}

	membersData, _ := json.MarshalIndent([]map[string]interface{}{expiredMember}, "", "  ")
	os.WriteFile(cfg.MembersFile, membersData, 0644)

	// Test Profile endpoint: must return isPaid: false
	r := gin.New()
	r.GET("/api/member/profile", func(c *gin.Context) {
		c.Set("member_id", expiredUUID)
		c.Set("member_email", "expired_profile@test.com")
		c.Next()
	}, GetMemberProfileByToken)

	req, _ := http.NewRequest("GET", "/api/member/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var res map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))

	user := res["user"].(map[string]interface{})
	assert.Equal(t, false, user["isPaid"], "Expired offline paid member must have isPaid: false")
	assert.Equal(t, false, user["subscription_active"], "Expired offline paid member must have subscription_active: false")

	// Test GetMemberPaidSubscriptions: should NOT contain annual-subscription
	rSub := gin.New()
	rSub.GET("/api/subscriptions/member-paid", GetMemberPaidSubscriptions)

	reqSub, _ := http.NewRequest("GET", "/api/subscriptions/member-paid?email=expired_profile@test.com", nil)
	wSub := httptest.NewRecorder()
	rSub.ServeHTTP(wSub, reqSub)

	assert.Equal(t, http.StatusOK, wSub.Code)
	var paidIDs []string
	assert.NoError(t, json.Unmarshal(wSub.Body.Bytes(), &paidIDs))
	assert.NotContains(t, paidIDs, "annual-subscription", "Expired offline member must not have active annual-subscription")
}
