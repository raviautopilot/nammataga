package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"taga-api/model"
	"taga-api/service"
	"taga-api/utils"
)

// GetAccountDeletionEligibility godoc
// @Summary Check account deletion eligibility
// @Tags Member
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.AccountDeletionEligibilityResponse
// @Router /api/member/account-deletion/eligibility [get]
func GetAccountDeletionEligibility(c *gin.Context) {
	memberID, _ := c.Get("member_id")
	memberEmail, _ := c.Get("member_email")

	memberIDStr := ""
	if memberID != nil {
		memberIDStr = memberID.(string)
	}
	emailStr := ""
	if memberEmail != nil {
		emailStr = memberEmail.(string)
	}

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load member profile")
		return
	}

	var foundMember map[string]interface{}
	for _, m := range members {
		id := utils.GetStringFromMap(m, "id")
		em := utils.GetStringFromMap(m, "emailId")
		if em == "" {
			em = utils.GetStringFromMap(m, "email")
		}
		if (memberIDStr != "" && id == memberIDStr) || (emailStr != "" && strings.EqualFold(em, emailStr)) {
			foundMember = m
			break
		}
	}

	if foundMember == nil {
		respondError(c, http.StatusNotFound, "Member not found")
		return
	}

	tagaID := utils.GetStringFromMap(foundMember, "tagaId")

	// 1. Active stay check
	hasActiveStay, err := service.HasActiveStayAtTagaTowers(memberIDStr, tagaID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to verify stay status")
		return
	}
	if hasActiveStay {
		c.JSON(http.StatusOK, model.AccountDeletionEligibilityResponse{
			Eligible: false,
			Reason:   "active_stay",
			Message:  service.ActiveStayMessage,
		})
		return
	}

	// 2. Pending request check
	hasPending, _, err := service.GetPendingDeletionRequest(memberIDStr, tagaID, emailStr)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to check deletion request status")
		return
	}
	if hasPending {
		c.JSON(http.StatusOK, model.AccountDeletionEligibilityResponse{
			Eligible: false,
			Reason:   "request_pending",
			Message:  service.PendingAccountMessage,
		})
		return
	}

	c.JSON(http.StatusOK, model.AccountDeletionEligibilityResponse{
		Eligible: true,
		Message:  "Eligible for account deletion",
	})
}

// RequestAccountDeletion godoc
// @Summary Request account deletion with password verification
// @Tags Member
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.AccountDeletionRequestPayload true "Account deletion payload"
// @Success 200 {object} map[string]interface{}
// @Router /api/member/account-deletion [post]
func RequestAccountDeletion(c *gin.Context) {
	var payload model.AccountDeletionRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "Password is required for account deletion")
		return
	}

	if strings.TrimSpace(payload.Password) == "" {
		respondError(c, http.StatusBadRequest, "Password cannot be empty")
		return
	}

	memberID, _ := c.Get("member_id")
	memberEmail, _ := c.Get("member_email")

	memberIDStr := ""
	if memberID != nil {
		memberIDStr = memberID.(string)
	}
	emailStr := ""
	if memberEmail != nil {
		emailStr = memberEmail.(string)
	}

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load member profile")
		return
	}

	var foundMember map[string]interface{}
	for _, m := range members {
		id := utils.GetStringFromMap(m, "id")
		em := utils.GetStringFromMap(m, "emailId")
		if em == "" {
			em = utils.GetStringFromMap(m, "email")
		}
		if (memberIDStr != "" && id == memberIDStr) || (emailStr != "" && strings.EqualFold(em, emailStr)) {
			foundMember = m
			break
		}
	}

	if foundMember == nil {
		respondError(c, http.StatusNotFound, "Member not found")
		return
	}

	// 1. Mandatory password confirmation
	storedPassword := utils.GetStringFromMap(foundMember, "password")
	if storedPassword == "" {
		respondError(c, http.StatusBadRequest, "Invalid password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(payload.Password)); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid password")
		return
	}

	tagaID := utils.GetStringFromMap(foundMember, "tagaId")
	memberName := utils.GetStringFromMap(foundMember, "name")
	memberEmailFromProfile := utils.GetStringFromMap(foundMember, "emailId")
	if memberEmailFromProfile == "" {
		memberEmailFromProfile = utils.GetStringFromMap(foundMember, "email")
	}
	if emailStr == "" {
		emailStr = memberEmailFromProfile
	}

	mobileNumber := utils.GetStringFromMap(foundMember, "mobile_number")
	if mobileNumber == "" {
		mobileNumber = utils.GetStringFromMap(foundMember, "mobileNumber")
	}
	if mobileNumber == "" {
		mobileNumber = utils.GetStringFromMap(foundMember, "phone")
	}
	if mobileNumber == "" {
		mobileNumber = utils.GetStringFromMap(foundMember, "mobile")
	}

	// 2. Active Stay Check
	hasActiveStay, err := service.HasActiveStayAtTagaTowers(memberIDStr, tagaID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to verify stay status")
		return
	}
	if hasActiveStay {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   service.ActiveStayMessage,
			"message": service.ActiveStayMessage,
		})
		return
	}

	// 3. Create deletion request
	req, err := service.CreateDeletionRequest(c, memberIDStr, tagaID, memberName, emailStr, payload.Reason, mobileNumber)
	if err != nil {
		if strings.Contains(err.Error(), "already pending") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Account deletion request is already pending",
				"message": "Account deletion request is already pending",
			})
			return
		}
		if err == service.ErrActiveStay {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   service.ActiveStayMessage,
				"message": service.ActiveStayMessage,
			})
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Account deletion request submitted successfully",
		"status":     string(req.Status),
		"request_id": req.ID,
	})
}

// GetAccountDeletionStatus godoc
// @Summary Get account deletion status for current member
// @Tags Member
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/member/account-deletion/status [get]
func GetAccountDeletionStatus(c *gin.Context) {
	memberID, _ := c.Get("member_id")
	memberEmail, _ := c.Get("member_email")

	memberIDStr := ""
	if memberID != nil {
		memberIDStr = memberID.(string)
	}
	emailStr := ""
	if memberEmail != nil {
		emailStr = memberEmail.(string)
	}

	tagaID := getMemberTagaIdByUUID(memberIDStr)
	if tagaID == "" && emailStr != "" {
		tagaID = getMemberTagaIdByEmail(emailStr)
	}

	latestReq, err := service.GetLatestDeletionRequest(memberIDStr, tagaID, emailStr)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to check deletion request status")
		return
	}

	if latestReq == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "none",
		})
		return
	}

	if latestReq.Status == model.DeletionStatusPending {
		c.JSON(http.StatusOK, gin.H{
			"status":       string(latestReq.Status),
			"request_id":   latestReq.ID,
			"requested_at": latestReq.RequestedAt,
			"reason":       latestReq.Reason,
		})
		return
	}

	if latestReq.Status == model.DeletionStatusRejected {
		c.JSON(http.StatusOK, gin.H{
			"status":       string(latestReq.Status),
			"request_id":   latestReq.ID,
			"requested_at": latestReq.RequestedAt,
			"processed_at": latestReq.ProcessedAt,
			"remarks":      latestReq.Remarks,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       string(latestReq.Status),
		"request_id":   latestReq.ID,
		"requested_at": latestReq.RequestedAt,
		"processed_at": latestReq.ProcessedAt,
	})
}

// ListDeletionRequests godoc
// @Summary List account deletion requests for admin review
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (e.g. pending, approved, rejected)"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/deletion-requests [get]
func ListDeletionRequests(c *gin.Context) {
	statusFilter := c.Query("status")
	requests, err := service.ListDeletionRequests(statusFilter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to retrieve deletion requests")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": requests,
		"total":    len(requests),
	})
}

// ApproveDeletionRequest godoc
// @Summary Approve an account deletion request
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Deletion Request ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/deletion-requests/{id}/approve [post]
func ApproveDeletionRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		respondError(c, http.StatusBadRequest, "Request ID is required")
		return
	}

	req, err := service.GetDeletionRequestByID(requestID)
	if err != nil {
		respondError(c, http.StatusNotFound, "Deletion request not found")
		return
	}

	if req.Status != model.DeletionStatusPending {
		respondError(c, http.StatusBadRequest, "Deletion request is already processed")
		return
	}

	adminUsername := getAdminUsername(c)
	if adminUsername == "" {
		adminUsername = "admin"
	}

	err = service.ExecuteAccountDeletion(c, req.MemberID, req.TagaID, adminUsername, req.Reason, req.ID)
	if err != nil {
		if err == service.ErrActiveStay {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   service.ActiveStayMessage,
				"message": service.ActiveStayMessage,
			})
			return
		}
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to execute deletion: %s", err.Error()))
		return
	}

	respondMessage(c, "Account deletion approved and processed successfully")
}

// RejectDeletionRequest godoc
// @Summary Reject an account deletion request with mandatory remarks
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Deletion Request ID"
// @Param payload body model.RejectDeletionRequestPayload true "Rejection remarks"
// @Success 200 {object} map[string]interface{}
// @Router /api/admin/deletion-requests/{id}/reject [post]
func RejectDeletionRequest(c *gin.Context) {
	requestID := c.Param("id")
	if requestID == "" {
		respondError(c, http.StatusBadRequest, "Request ID is required")
		return
	}

	var payload model.RejectDeletionRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "Remarks are required for rejecting deletion request")
		return
	}

	trimmedRemarks := strings.TrimSpace(payload.Remarks)
	if trimmedRemarks == "" {
		respondError(c, http.StatusBadRequest, "Remarks are required for rejecting deletion request")
		return
	}

	adminUsername := getAdminUsername(c)
	if adminUsername == "" {
		adminUsername = "admin"
	}

	_, err := service.RejectDeletionRequest(c, requestID, adminUsername, trimmedRemarks)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(c, http.StatusNotFound, "Deletion request not found")
			return
		}
		if strings.Contains(err.Error(), "already processed") {
			respondError(c, http.StatusBadRequest, "Deletion request is already processed")
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	respondMessage(c, "Account deletion request rejected")
}
