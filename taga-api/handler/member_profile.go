package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/audit"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const editRequestsFilePath = "data/edit_requests.json"

// GetMemberProfileByToken - JWT based profile fetch
// @Summary Get member profile
// @Tags Member Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/member/profile [get]
func GetMemberProfileByToken(c *gin.Context) {
	memberID, exists := c.Get("member_id")
	if !exists {
		respondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read members", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to get profile")
		return
	}

	for _, member := range members {
		if id, ok := member["id"].(string); ok && id == memberID {
			delete(member, "password")
			member["isPaid"] = checkAnnualSubscriptionStatus(memberID.(string))
			member["subscription_active"] = member["isPaid"]
			respondOK(c, gin.H{"user": member})
			return
		}
	}

	respondError(c, http.StatusNotFound, "Member not found")
}

// UpdateMemberProfileHandler updates member details
// @Summary Update member profile
// @Tags Member Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Profile Update Request"
// @Success 200 {object} map[string]interface{}
// @Router /api/member/profile [put]
func UpdateMemberProfileHandler(c *gin.Context) {
	memberID, exists := c.Get("member_id")
	if !exists {
		respondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	found := false
	var oldState map[string]interface{}
	for i, member := range members {
		if id, ok := member["id"].(string); ok && id == memberID {
			// Capture old state for audit
			oldCopy := make(map[string]interface{}, len(member))
			for k, v := range member {
				oldCopy[k] = v
			}
			oldState = oldCopy
			if name, ok := updates["name"].(string); ok {
				members[i]["name"] = name
			}
			if mobile, ok := updates["mobile_number"].(string); ok {
				members[i]["mobile_number"] = mobile
			}
			if designation, ok := updates["designation"].(string); ok {
				members[i]["designation"] = designation
			}
			if residentialAddress, ok := updates["residential_address"].(string); ok {
				members[i]["residential_address"] = residentialAddress
			}
			if permanentAddress, ok := updates["permanent_address"].(string); ok {
				members[i]["permanent_address"] = permanentAddress
			}
			found = true
			break
		}
	}

	if !found {
		respondError(c, http.StatusNotFound, "Member not found")
		return
	}

	cfg := config.GetConfig()
	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save")
		return
	}

	if err := os.WriteFile(cfg.MembersFile, data, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save")
		return
	}

	// Audit member self-update
	tagaID := getMemberTagaIdByUUID(fmt.Sprintf("%v", memberID))
	memberEmail := ""
	if val, ok := c.Get("member_email"); ok {
		memberEmail, _ = val.(string)
	}
	var newState map[string]interface{}
	for _, m := range members {
		if id, _ := m["id"].(string); id == fmt.Sprintf("%v", memberID) {
			newState = m
			break
		}
	}
	_ = audit.Log(c, tagaID, memberEmail,
		audit.ActionUpdate, audit.ModuleMember,
		"member", tagaID,
		fmt.Sprintf("Member %s updated their profile", memberEmail),
		audit.Sanitize(oldState), audit.Sanitize(newState))

	respondMessage(c, "Profile updated successfully")
}

// CreateEditRequest godoc
// @Summary Create an edit request
// @Tags Member Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.EditRequest true "Edit Request Details"
// @Success 200 {object} map[string]interface{}
// @Router /api/member/edit-request [post]
func CreateEditRequest(c *gin.Context) {
	var req model.EditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to read members")
		return
	}

	var currentMember map[string]interface{}
	for _, m := range members {
		if email, ok := m["emailId"].(string); ok && email == req.Email {
			currentMember = m
			break
		}
	}

	if currentMember == nil {
		respondError(c, http.StatusNotFound, "member not found")
		return
	}

	memberID, _ := currentMember["id"].(string)
	memberName, _ := currentMember["name"].(string)

	requestGroupId := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	var newFields []model.FieldEditRequest

	// Compare fields
	compareAndAdd := func(fieldKey, frontendValue string) {
		if frontendValue == "" {
			return
		}
		oldValue := ""
		if val, ok := currentMember[fieldKey].(string); ok {
			oldValue = val
		}
		if oldValue != frontendValue {
			newFields = append(newFields, model.FieldEditRequest{
				ID:             uuid.New().String(),
				RequestGroupID: requestGroupId,
				MemberID:       memberID,
				Email:          req.Email,
				MemberName:     memberName,
				Field:          fieldKey,
				OldValue:       oldValue,
				NewValue:       frontendValue,
				MemberRemarks:  req.Remarks,
				Status:         "pending",
				CreatedAt:      now,
			})
		}
	}

	compareAndAdd("mobile_number", req.MobileNumber)
	compareAndAdd("emailId", req.MailId)
	compareAndAdd("designation", req.Designation)
	compareAndAdd("working_district", req.WorkingDistrict)
	compareAndAdd("residential_address", req.ResidentialAddress)
	compareAndAdd("permanent_address", req.PermanentAddress)

	if len(newFields) == 0 {
		respondError(c, http.StatusBadRequest, "No changes detected between form and current profile")
		return
	}

	// Read pending requests
	pendingFile := "data/requests/pending_requests.json"
	var pendingRequests []model.FieldEditRequest
	file, _ := os.ReadFile(pendingFile)
	if len(file) > 0 {
		_ = json.Unmarshal(file, &pendingRequests)
	}

	pendingRequests = append(pendingRequests, newFields...)
	data, err := json.MarshalIndent(pendingRequests, "", "  ")
	if err != nil || os.WriteFile(pendingFile, data, 0644) != nil {
		respondError(c, http.StatusInternalServerError, "failed to save request")
		return
	}

	// Send Admin Email (We will write SendAdminEditRequestEmail in mail_service next)
	err = service.SendAdminEditRequestEmail(req.Email, memberName, newFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Request saved but email failed",
			"error":   err.Error(),
		})
		return
	}

	tagaID, _ := currentMember["tagaId"].(string)
	_ = audit.Log(c, tagaID, req.Email,
		audit.ActionCreate, audit.ModuleMember,
		"edit_request", requestGroupId,
		fmt.Sprintf("Member submitted profile edit request: %s", req.Remarks),
		nil, newFields)

	respondMessage(c, "Edit request submitted successfully")
}

// GetMemberNotifications godoc
// @Summary Get notifications for a member
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
		respondError(c, http.StatusBadRequest, "member_id is required")
		return
	}

	notifications, err := readMemberNotificationsFromFile(memberID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read notifications")
		return
	}

	respondOK(c, notifications)
}

// MarkNotificationRead godoc
// @Summary Mark notification as read
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
		respondError(c, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	respondMessage(c, "Notification marked as read")
}

// GetUnreadCount godoc
// @Summary Get unread notification count for a member
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
		respondError(c, http.StatusBadRequest, "member_id is required")
		return
	}

	count, err := getUnreadNotificationCountFromFile(memberID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	respondOK(c, gin.H{"unread_count": count})
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
