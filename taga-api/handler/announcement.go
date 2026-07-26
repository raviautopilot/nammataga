package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"taga-api/config"
	"taga-api/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AnnouncementRequest struct {
	Title    string `json:"title" binding:"required"`
	Message  string `json:"message" binding:"required"`
	Priority string `json:"priority"`
	SendTo   string `json:"sendTo"`
	District string `json:"district,omitempty"`
}

type AnnouncementResponse struct {
	Message        string `json:"message"`
	Recipients     int    `json:"recipients"`
	SendTo         string `json:"send_to"`
	AnnouncementID string `json:"announcement_id"`
}

// SendAnnouncement godoc
// @Summary Send announcement to members
// @Description Send email announcement to members and store in database
// @Tags Admin Announcements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param announcement body AnnouncementRequest true "Announcement details"
// @Success 200 {object} AnnouncementResponse
// @Router /api/admin/announcements/send [post]
func SendAnnouncement(c *gin.Context) {
	var req AnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Priority == "" {
		req.Priority = "normal"
	}
	if req.SendTo == "" {
		req.SendTo = "all"
	}

	// Get admin email from context
	adminEmail, exists := c.Get("username")
	if !exists {
		adminEmail = config.GetConfig().AdminEmail
	}

	// Read members
	members, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	if len(members) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":    "No members found in the system",
			"recipients": 0,
			"send_to":    req.SendTo,
		})
		return
	}

	// Filter recipients based on sendTo
	var recipients []map[string]interface{}
	sendToLower := strings.ToLower(strings.TrimSpace(req.SendTo))

	for _, member := range members {
		email, ok := member["emailId"].(string)
		if !ok || email == "" {
			continue
		}

		memberID, _ := member["id"].(string)
		workingDistrict := getString(member, "working_district")

		// Get payment status - you can update this when you add payment tracking
		// For now, you can add a "payment_status" field to members.json
		paymentStatus := getString(member, "payment_status")

		include := false

		switch sendToLower {
		case "all", "all members":
			// Include all members
			include = true

		case "paid", "paid members", "paid members only":
			// Include only paid members
			// You need to add "payment_status": "paid" to members.json
			// For now, it checks if payment_status exists and equals "paid"
			// To enable: add "payment_status": "paid" to your member entries
			if paymentStatus == "paid" || paymentStatus == "Paid" || paymentStatus == "PAID" {
				include = true
			}
			// If you don't have payment_status field yet, this will return 0 recipients
			// Once you add the field, it will work automatically

		case "unpaid", "unpaid members":
			// Include only unpaid members
			if paymentStatus == "unpaid" || paymentStatus == "Unpaid" || paymentStatus == "UNPAID" {
				include = true
			}

		case "district", "district members":
			// Filter by district
			if req.District != "" && strings.EqualFold(workingDistrict, req.District) {
				include = true
			}

		default:
			include = true
		}

		if include {
			recipients = append(recipients, map[string]interface{}{
				"email":            email,
				"id":               memberID,
				"working_district": workingDistrict,
				"payment_status":   paymentStatus,
			})
		}
	}

	if len(recipients) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":    fmt.Sprintf("No recipients found for filter: %s", req.SendTo),
			"recipients": 0,
			"send_to":    req.SendTo,
		})
		return
	}

	// Create announcement record
	announcement := model.Notification{
		ID:         uuid.New().String(),
		Title:      req.Title,
		Message:    req.Message,
		Priority:   req.Priority,
		SendTo:     req.SendTo,
		SentBy:     adminEmail.(string),
		SentAt:     time.Now(),
		Recipients: len(recipients),
		District:   req.District,
	}

	// Save announcement
	if err := saveAnnouncementToFile(announcement); err != nil {
		config.Logger.Error("Failed to save announcement", zap.Error(err))
	}

	// Save member notifications
	for _, recipient := range recipients {
		memberNotification := model.MemberNotification{
			ID:             uuid.New().String(),
			MemberID:       recipient["id"].(string),
			MemberEmail:    recipient["email"].(string),
			NotificationID: announcement.ID,
			Title:          req.Title,
			Message:        req.Message,
			Priority:       req.Priority,
			IsRead:         false,
			CreatedAt:      time.Now(),
		}
		if err := saveMemberNotificationToFile(memberNotification); err != nil {
			config.Logger.Error("Failed to save member notification", zap.Error(err))
		}
	}

	// Send emails asynchronously
	go func() {
		config.Logger.Info("Starting to send announcement emails",
			zap.Int("recipient_count", len(recipients)),
			zap.String("title", req.Title))

		for _, recipient := range recipients {
			subject := formatAnnouncementSubject(req.Title, req.Priority)
			body := buildAnnouncementEmailContent(req.Title, req.Message, req.Priority)

			if err := sendEmail(recipient["email"].(string), subject, body); err != nil {
				config.Logger.Error("Failed to send announcement email",
					zap.String("to", recipient["email"].(string)),
					zap.Error(err))
			}
		}
	}()

	c.JSON(http.StatusOK, AnnouncementResponse{
		Message:        fmt.Sprintf("Announcement sent successfully to %d recipients", len(recipients)),
		Recipients:     len(recipients),
		SendTo:         req.SendTo,
		AnnouncementID: announcement.ID,
	})
}

// GetMemberNotifications godoc
// @Summary Get notifications for a member
// @Description Get all notifications for a specific member
// @Tags Member Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member_id query string true "Member ID"
// @Success 200 {array} model.MemberNotification
// @Router /api/member/notifications [get]
func GetMemberNotifications(c *gin.Context) {
	memberID := c.Query("member_id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member_id is required"})
		return
	}

	notifications, err := readMemberNotificationsFromFile(memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkNotificationRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags Member Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/member/notifications/{id}/read [put]
func MarkNotificationRead(c *gin.Context) {
	notificationID := c.Param("id")

	if err := markNotificationAsReadInFile(notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// GetUnreadCount godoc
// @Summary Get unread notification count for a member
// @Description Get count of unread notifications for a specific member
// @Tags Member Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param member_id query string true "Member ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/member/notifications/unread/count [get]
func GetUnreadCount(c *gin.Context) {
	memberID := c.Query("member_id")
	if memberID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member_id is required"})
		return
	}

	count, err := getUnreadNotificationCountFromFile(memberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// DebugMembers godoc
// @Summary Debug member list
// @Description Check how many members are loaded
// @Tags Admin Announcements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}


// ==================== FILE STORAGE FUNCTIONS ====================

func getAnnouncementsFilePath() string {
	dir := "data/announcements"
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "announcements.json")
}

func getMemberNotificationsFilePath() string {
	dir := "data/notifications"
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "member_notifications.json")
}

func saveAnnouncementToFile(announcement model.Notification) error {
	filePath := getAnnouncementsFilePath()

	var announcements []model.Notification
	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		json.Unmarshal(data, &announcements)
	}

	announcements = append([]model.Notification{announcement}, announcements...)

	updatedData, err := json.MarshalIndent(announcements, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, updatedData, 0644)
}

func saveMemberNotificationToFile(notification model.MemberNotification) error {
	filePath := getMemberNotificationsFilePath()

	var notifications []model.MemberNotification
	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		json.Unmarshal(data, &notifications)
	}

	notifications = append(notifications, notification)

	updatedData, err := json.MarshalIndent(notifications, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, updatedData, 0644)
}

func readMemberNotificationsFromFile(memberID string) ([]model.MemberNotification, error) {
	filePath := getMemberNotificationsFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.MemberNotification{}, nil
		}
		return nil, err
	}

	var notifications []model.MemberNotification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return nil, err
	}

	var memberNotifications []model.MemberNotification
	for _, n := range notifications {
		if n.MemberID == memberID {
			memberNotifications = append(memberNotifications, n)
		}
	}

	return memberNotifications, nil
}

func markNotificationAsReadInFile(notificationID string) error {
	filePath := getMemberNotificationsFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var notifications []model.MemberNotification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return err
	}

	now := time.Now()
	for i := range notifications {
		if notifications[i].ID == notificationID {
			notifications[i].IsRead = true
			notifications[i].ReadAt = &now
			break
		}
	}

	updatedData, err := json.MarshalIndent(notifications, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, updatedData, 0644)
}

func getUnreadNotificationCountFromFile(memberID string) (int, error) {
	notifications, err := readMemberNotificationsFromFile(memberID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, n := range notifications {
		if !n.IsRead {
			count++
		}
	}
	return count, nil
}

// ==================== HELPER FUNCTIONS ====================

func formatAnnouncementSubject(title, priority string) string {
	switch priority {
	case "urgent":
		return "🚨 URGENT: " + title
	case "high":
		return "⚠️ IMPORTANT: " + title
	default:
		return "📢 " + title
	}
}

func buildAnnouncementEmailContent(title, message, priority string) string {
	var priorityColor, priorityBadge string

	switch priority {
	case "urgent":
		priorityColor = "#ff4444"
		priorityBadge = "URGENT"
	case "high":
		priorityColor = "#ff8800"
		priorityBadge = "HIGH PRIORITY"
	default:
		priorityColor = "#44aa00"
		priorityBadge = "NORMAL"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: %s; color: white; padding: 10px 20px; border-radius: 5px; }
        .badge { display: inline-block; padding: 3px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; background: %s; color: white; }
        .content { margin: 20px 0; padding: 20px; background: #f9f9f9; border-radius: 5px; }
        .footer { margin-top: 20px; padding-top: 10px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header" style="background-color: %s;">
            <h2>Tamil Nadu Agriculture Graduates Association (TAGA)</h2>
        </div>
        <div class="content">
            <span class="badge" style="background-color: %s;">%s</span>
            <h3>%s</h3>
            <p>%s</p>
        </div>
        <div class="footer">
            <p>This is an official announcement from TAGA administration.</p>
            <p>Please login to your dashboard to view all announcements.</p>
        </div>
    </div>
</body>
</html>
`, priorityColor, priorityColor, priorityColor, priorityColor, priorityBadge, title, message)
}
