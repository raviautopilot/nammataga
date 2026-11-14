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
	"github.com/stretchr/testify/assert"
	"taga-api/model"
)

func TestGetMemberNotificationsAndUnreadCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempNotifDir := "data/notifications"
	tempAnnDir := "data/announcements"
	_ = os.MkdirAll(tempNotifDir, 0755)
	_ = os.MkdirAll(tempAnnDir, 0755)

	notifFile := filepath.Join(tempNotifDir, "member_notifications.json")
	annFile := filepath.Join(tempAnnDir, "announcements.json")

	origNotif, _ := os.ReadFile(notifFile)
	origAnn, _ := os.ReadFile(annFile)

	t.Cleanup(func() {
		if origNotif != nil {
			_ = os.WriteFile(notifFile, origNotif, 0644)
		}
		if origAnn != nil {
			_ = os.WriteFile(annFile, origAnn, 0644)
		}
	})

	testMemberID := "test-member-uuid-123"
	testEmail := "testmember@example.com"
	testTagaID := "TAGA-TEST-001"

	// 1. Direct member notification
	notifList := []model.MemberNotification{
		{
			ID:          "notif-1",
			MemberID:    testMemberID,
			MemberEmail: testEmail,
			TagaID:      testTagaID,
			Title:       "Personal Direct Notice",
			Message:     "Your subscription renewal is coming up.",
			Priority:    "urgent",
			IsRead:      false,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "notif-2",
			MemberID:    "other-member-uuid",
			MemberEmail: "other@example.com",
			Title:       "Other Notice",
			Message:     "Notice for someone else.",
			Priority:    "normal",
			IsRead:      false,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}
	notifData, _ := json.MarshalIndent(notifList, "", "  ")
	_ = os.WriteFile(notifFile, notifData, 0644)

	// 2. Broadcast announcement with send_to == "all"
	annList := []model.Notification{
		{
			ID:       "broadcast-ann-1",
			Title:    "Annual General Body Meeting",
			Message:  "AGM will take place at TAGA auditorium.",
			Priority: "urgent",
			SendTo:   "all",
			SentAt:   time.Now(),
		},
	}
	annData, _ := json.MarshalIndent(annList, "", "  ")
	_ = os.WriteFile(annFile, annData, 0644)

	// Test readMemberNotificationsFromFile matches on MemberID, MemberEmail, TagaID & includes broadcast
	notifs, err := readMemberNotificationsFromFile(testMemberID)
	assert.NoError(t, err)
	assert.NotEmpty(t, notifs)

	var foundDirect, foundBroadcast, foundOther bool
	for _, n := range notifs {
		if n.ID == "notif-1" {
			foundDirect = true
		}
		if n.ID == "broadcast-ann-1" {
			foundBroadcast = true
		}
		if n.ID == "notif-2" {
			foundOther = true
		}
	}
	assert.True(t, foundDirect, "Direct member notification must be present")
	assert.True(t, foundBroadcast, "Broadcast announcement must be automatically included")
	assert.False(t, foundOther, "Other member notification must NOT be included")

	// Test fallback to Gin context when query param is missing
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/member/notifications", nil)
	c.Set("member_id", testMemberID)

	GetMemberNotifications(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var respNotifs []model.MemberNotification
	err = json.Unmarshal(w.Body.Bytes(), &respNotifs)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(respNotifs), 2)

	// Test GetUnreadCount with JWT context fallback
	wUnread := httptest.NewRecorder()
	cUnread, _ := gin.CreateTestContext(wUnread)
	cUnread.Request, _ = http.NewRequest("GET", "/api/member/notifications/unread/count", nil)
	cUnread.Set("member_email", testEmail)

	GetUnreadCount(cUnread)
	assert.Equal(t, http.StatusOK, wUnread.Code)

	var unreadResp map[string]interface{}
	err = json.Unmarshal(wUnread.Body.Bytes(), &unreadResp)
	assert.NoError(t, err)
	assert.NotNil(t, unreadResp["unread_count"])
}
