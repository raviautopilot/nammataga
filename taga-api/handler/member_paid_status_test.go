package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
