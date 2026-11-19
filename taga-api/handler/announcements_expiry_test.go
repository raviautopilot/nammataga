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

func setupTestAnnouncementsDir(t *testing.T) (string, string) {
	annDir := filepath.Join("data", "announcements")
	notifDir := filepath.Join("data", "notifications")
	_ = os.MkdirAll(annDir, 0755)
	_ = os.MkdirAll(notifDir, 0755)

	annFile := filepath.Join(annDir, "announcements.json")
	archFile := filepath.Join(annDir, "archived_announcements.json")
	notifFile := filepath.Join(notifDir, "member_notifications.json")

	origAnn, _ := os.ReadFile(annFile)
	origArch, _ := os.ReadFile(archFile)
	origNotif, _ := os.ReadFile(notifFile)

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
		if origNotif != nil {
			_ = os.WriteFile(notifFile, origNotif, 0644)
		} else {
			_ = os.Remove(notifFile)
		}
	})

	_ = os.Remove(annFile)
	_ = os.Remove(archFile)
	_ = os.Remove(notifFile)

	return annFile, archFile
}

func TestAnnouncementExpiryEngine(t *testing.T) {
	annFile, _ := setupTestAnnouncementsDir(t)

	now := time.Now().UTC()

	// Initial dataset:
	// a1: 2 hours ago (active)
	// a2: 10 days ago (active)
	// a3: 14 days and 5 minutes ago (expired)
	// a4: 25 days ago (expired)
	// a5: legacy record with no published_at, sent_at is 1 hour ago (active)
	// a6: legacy record with no published_at, sent_at is 15 days ago (expired)
	initialAnnouncements := []model.Notification{
		{
			ID:          "ann-active-1",
			Title:       "Active Notice 1",
			Message:     "This notice was posted 2 hours ago",
			Priority:    "urgent",
			SendTo:      "all",
			PublishedAt: now.Add(-2 * time.Hour),
			SentAt:      now.Add(-2 * time.Hour),
			CreatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ID:          "ann-active-2",
			Title:       "Active Notice 2",
			Message:     "This notice was posted 10 days ago",
			Priority:    "normal",
			SendTo:      "all",
			PublishedAt: now.Add(-10 * 24 * time.Hour),
			SentAt:      now.Add(-10 * 24 * time.Hour),
			CreatedAt:   now.Add(-10 * 24 * time.Hour),
		},
		{
			ID:          "ann-expired-1",
			Title:       "Expired Notice 1",
			Message:     "This notice was posted 14 days and 5 minutes ago",
			Priority:    "high",
			SendTo:      "all",
			PublishedAt: now.Add(-14*24*time.Hour - 5*time.Minute),
			SentAt:      now.Add(-14*24*time.Hour - 5*time.Minute),
			CreatedAt:   now.Add(-14*24*time.Hour - 5*time.Minute),
		},
		{
			ID:          "ann-expired-2",
			Title:       "Expired Notice 2",
			Message:     "This notice was posted 25 days ago",
			Priority:    "normal",
			SendTo:      "all",
			PublishedAt: now.Add(-25 * 24 * time.Hour),
			SentAt:      now.Add(-25 * 24 * time.Hour),
			CreatedAt:   now.Add(-25 * 24 * time.Hour),
		},
		{
			ID:       "ann-legacy-active",
			Title:    "Legacy Active Notice",
			Message:  "Legacy notice without published_at, sent 1 hour ago",
			Priority: "normal",
			SendTo:   "all",
			SentAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:       "ann-legacy-expired",
			Title:    "Legacy Expired Notice",
			Message:  "Legacy notice without published_at, sent 15 days ago",
			Priority: "normal",
			SendTo:   "all",
			SentAt:   now.Add(-15 * 24 * time.Hour),
		},
	}

	data, err := json.MarshalIndent(initialAnnouncements, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(annFile, data, 0644)
	require.NoError(t, err)

	// Step 1: Run GetActiveAnnouncements()
	active, err := GetActiveAnnouncements()
	require.NoError(t, err)

	// Verify active items: should have 3 items (ann-active-1, ann-active-2, ann-legacy-active)
	assert.Len(t, active, 3, "Only active announcements (< 14 days) should be returned")

	activeIDs := make(map[string]bool)
	for _, a := range active {
		activeIDs[a.ID] = true
	}
	assert.True(t, activeIDs["ann-active-1"], "ann-active-1 should be active")
	assert.True(t, activeIDs["ann-active-2"], "ann-active-2 should be active")
	assert.True(t, activeIDs["ann-legacy-active"], "ann-legacy-active should be active")
	assert.False(t, activeIDs["ann-expired-1"], "ann-expired-1 must not be in active")
	assert.False(t, activeIDs["ann-expired-2"], "ann-expired-2 must not be in active")
	assert.False(t, activeIDs["ann-legacy-expired"], "ann-legacy-expired must not be in active")

	// Step 2: Verify active announcements are sorted newest first
	assert.True(t, active[0].PublishedAt.After(active[1].PublishedAt) || active[0].PublishedAt.Equal(active[1].PublishedAt))
	assert.True(t, active[1].PublishedAt.After(active[2].PublishedAt) || active[1].PublishedAt.Equal(active[2].PublishedAt))

	// Step 3: Verify archived announcements in archived_announcements.json
	archived, err := GetArchivedAnnouncements()
	require.NoError(t, err)
	assert.Len(t, archived, 3, "Expired announcements should be moved to archive")

	archivedIDs := make(map[string]bool)
	for _, a := range archived {
		archivedIDs[a.ID] = true
	}
	assert.True(t, archivedIDs["ann-expired-1"])
	assert.True(t, archivedIDs["ann-expired-2"])
	assert.True(t, archivedIDs["ann-legacy-expired"])

	// Step 4: Verify repeated calls do not duplicate archives
	activeSecondCall, err := GetActiveAnnouncements()
	require.NoError(t, err)
	assert.Len(t, activeSecondCall, 3)

	archivedSecondCall, err := GetArchivedAnnouncements()
	require.NoError(t, err)
	assert.Len(t, archivedSecondCall, 3, "Repeated calls must NOT duplicate archived announcements")

	// Step 5: Verify no dependency on member_notifications.json
	notifFile := filepath.Join("data", "notifications", "member_notifications.json")
	_, statErr := os.Stat(notifFile)
	assert.True(t, os.IsNotExist(statErr), "member_notifications.json should NOT be created or used")

	// Step 6: Verify SaveAnnouncement prepends and keeps active
	newAnn := model.Notification{
		ID:       "ann-new-post",
		Title:    "Brand New Notice",
		Message:  "Just posted now",
		Priority: "urgent",
		SendTo:   "all",
	}
	err = SaveAnnouncement(newAnn)
	require.NoError(t, err)

	activeAfterSave, err := GetActiveAnnouncements()
	require.NoError(t, err)
	assert.Len(t, activeAfterSave, 4)
	assert.Equal(t, "ann-new-post", activeAfterSave[0].ID, "Newest announcement should be first")

	// Ensure archives still have only 3
	archivedAfterSave, err := GetArchivedAnnouncements()
	require.NoError(t, err)
	assert.Len(t, archivedAfterSave, 3)
}

func TestMemberNotificationEndpoints(t *testing.T) {
	annFile, _ := setupTestAnnouncementsDir(t)
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	testAnn := []model.Notification{
		{
			ID:          "endpoint-ann-1",
			Title:       "Endpoint Test Announcement",
			Message:     "Testing member notifications endpoint",
			Priority:    "high",
			SendTo:      "all",
			PublishedAt: now,
			SentAt:      now,
			CreatedAt:   now,
		},
	}
	data, _ := json.MarshalIndent(testAnn, "", "  ")
	_ = os.WriteFile(annFile, data, 0644)

	// 1. GET /api/member/notifications WITHOUT member_id query param
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/member/notifications", nil)

		GetMemberNotifications(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp []model.Notification
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "endpoint-ann-1", resp[0].ID)
		assert.False(t, resp[0].IsRead)
		assert.False(t, resp[0].PublishedAt.IsZero())
	}

	// 2. GET /api/member/notifications WITH member_id query param
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/member/notifications?member_id=some-member-id", nil)

		GetMemberNotifications(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp []model.Notification
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "endpoint-ann-1", resp[0].ID)
	}

	// 3. PUT /api/member/notifications/:id/read - Safe 200 OK no-op
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("PUT", "/api/member/notifications/endpoint-ann-1/read", nil)
		c.Params = gin.Params{{Key: "id", Value: "endpoint-ann-1"}}

		MarkNotificationRead(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "success", resp["status"])
	}

	// 4. GET /api/member/notifications/unread/count - Returns count 0
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/member/notifications/unread/count", nil)

		GetUnreadCount(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, float64(0), resp["count"])
		assert.Equal(t, float64(0), resp["unread_count"])
	}
}
