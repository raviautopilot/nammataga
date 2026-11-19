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
	"github.com/stretchr/testify/require"
	"taga-api/model"
)

func TestGetMemberNotificationsAndUnreadCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempAnnDir := filepath.Join("data", "announcements")
	_ = os.MkdirAll(tempAnnDir, 0755)

	annFile := filepath.Join(tempAnnDir, "announcements.json")
	archFile := filepath.Join(tempAnnDir, "archived_announcements.json")

	origAnn, _ := os.ReadFile(annFile)
	origArch, _ := os.ReadFile(archFile)

	t.Cleanup(func() {
		if origAnn != nil {
			_ = os.WriteFile(annFile, origAnn, 0644)
		} else {
			_ = os.Remove(annFile)
		}
		if origArch != nil {
			_ = os.WriteFile(archFile, origArch, 0644)
		} else {
			_ = os.Remove(archFile)
		}
	})

	testMemberID := "test-member-uuid-123"
	testEmail := "testmember@example.com"

	// Broadcast announcements with send_to == "all"
	annList := []model.Notification{
		{
			ID:          "broadcast-ann-1",
			Title:       "Annual General Body Meeting",
			Message:     "AGM will take place at TAGA auditorium.",
			Priority:    "urgent",
			SendTo:      "all",
			PublishedAt: time.Now().UTC(),
			SentAt:      time.Now().UTC(),
			CreatedAt:   time.Now().UTC(),
		},
	}
	annData, _ := json.MarshalIndent(annList, "", "  ")
	err := os.WriteFile(annFile, annData, 0644)
	require.NoError(t, err)

	// Test GetActiveAnnouncements includes broadcast announcement
	notifs, err := GetActiveAnnouncements()
	assert.NoError(t, err)
	assert.NotEmpty(t, notifs)

	var foundBroadcast bool
	for _, n := range notifs {
		if n.ID == "broadcast-ann-1" {
			foundBroadcast = true
		}
	}
	assert.True(t, foundBroadcast, "Broadcast announcement must be automatically included")

	// Test GET /api/member/notifications with member_id in context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/member/notifications", nil)
	c.Set("member_id", testMemberID)

	GetMemberNotifications(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var respNotifs []model.Notification
	err = json.Unmarshal(w.Body.Bytes(), &respNotifs)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(respNotifs), 1)

	// Test GET /api/member/notifications/unread/count with member_email in context
	wUnread := httptest.NewRecorder()
	cUnread, _ := gin.CreateTestContext(wUnread)
	cUnread.Request, _ = http.NewRequest("GET", "/api/member/notifications/unread/count", nil)
	cUnread.Set("member_email", testEmail)

	GetUnreadCount(cUnread)
	assert.Equal(t, http.StatusOK, wUnread.Code)

	var unreadResp map[string]interface{}
	err = json.Unmarshal(wUnread.Body.Bytes(), &unreadResp)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), unreadResp["count"])
	assert.Equal(t, float64(0), unreadResp["unread_count"])
}
