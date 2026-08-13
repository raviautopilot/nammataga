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
	"taga-api/service/audit"
	"taga-api/service/auth"
	"taga-api/service/jwt"
	"taga-api/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type MemberLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type MemberLoginResponse struct {
	Token               string      `json:"token"`
	Role                string      `json:"role"`
	ExpiresIn           int64       `json:"expires_in"`
	User                interface{} `json:"user"`
	ForceChangePassword bool        `json:"forceChangePassword"`
}

func checkAnnualSubscriptionStatus(memberID string) bool {
	now := time.Now()


	tagaId := getMemberTagaIdByUUID(memberID)
	cfg := config.GetConfig()
	subsFile := filepath.Join(filepath.Dir(cfg.MembersFile), "..", "subscriptions", "member_subscriptions.json")

	data, err := os.ReadFile(subsFile)
	if err != nil {
		config.Logger.Warn("Failed to read subscriptions file", zap.Error(err))
		return false
	}

	var subscriptions []model.MemberSubscription
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		config.Logger.Warn("Failed to parse subscriptions", zap.Error(err))
		return false
	}

	for _, sub := range subscriptions {
		if sub.MemberID == tagaId &&
			sub.SubscriptionID == "annual-subscription" &&
			sub.Status == "active" {
			graceEnd := sub.EndDate.AddDate(0, 2, 0)
			if now.Before(graceEnd) || now.Equal(graceEnd) {
				return true
			}
		}
	}
	return false
}

// ResetPasswordHandler handles password reset
// @Summary Reset Password
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param resetRequest body model.ResetPasswordRequest true "Reset details"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/reset-password [post]
func ResetPasswordHandler(c *gin.Context) {
	var resetRequest model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&resetRequest); err != nil {
		config.Logger.Error("Invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := auth.ResetPassword(resetRequest.Email, resetRequest.OldPassword, resetRequest.NewPassword)
	if err != nil {
		if err.Error() == "invalid or expired token" {
			respondError(c, http.StatusUnauthorized, "Invalid or expired token")
			return
		}
		config.Logger.Error("Password reset failed", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Password reset failed")
		return
	}

	respondMessage(c, "Password reset successful")
}

// ForgotPasswordHandler handles password reset requests
// @Summary Forgot Password
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param forgotPasswordRequest body model.ForgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/forgot-password [post]
func ForgotPasswordHandler(c *gin.Context) {
	var forgotPasswordRequest model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		config.Logger.Error("Invalid request body", zap.Error(err))
		respondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := auth.ForgotPassword(forgotPasswordRequest.Email)
	if err != nil {
		if err.Error() == "email not found" {
			respondError(c, http.StatusNotFound, "Email not found")
			return
		}
		config.Logger.Error("Forgot password failed", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to process request")
		return
	}

	respondMessage(c, "If the email exists, a password reset link has been sent")
}

// MemberForgotPasswordHandler godoc
// @Summary Member Forgot Password
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param request body model.ForgotPasswordRequest true "Forgot Password"
// @Success 200 {object} map[string]string
// @Router /api/auth/member-forgot-password [post]
func MemberForgotPasswordHandler(c *gin.Context) {
	var data model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := auth.ForgotPassword(data.Email)
	if err != nil {
		respondError(c, http.StatusNotFound, err.Error())
		return
	}

	respondMessage(c, "Password reset email sent")
}

// MemberLoginHandler - JWT based login for members
// @Summary Member Login
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param credentials body MemberLoginRequest true "Member credentials"
// @Success 200 {object} MemberLoginResponse
// @Router /api/member/login [post]
func MemberLoginHandler(c *gin.Context) {
	var req MemberLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read members", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Login failed")
		return
	}

	var foundMember map[string]interface{}
	for _, member := range members {
		email, ok := member["emailId"].(string)
		if ok && strings.EqualFold(email, req.Email) {
			foundMember = member
			break
		}
	}

	if foundMember == nil {
		respondError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	storedPassword, ok := foundMember["password"].(string)
	if !ok {
		respondError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		config.Logger.Warn("Invalid password for member", zap.String("email", req.Email))
		// Audit failed login — use tagaId if resolvable, otherwise anonymous
		tagaID := getMemberTagaIdByEmail(req.Email)
		resID := tagaID
		if tagaID == "" {
			tagaID = "anonymous"
			resID = req.Email
		}
		_ = audit.Log(c, tagaID, req.Email,
			audit.ActionLoginFailed, audit.ModuleAuth,
			"member", resID, fmt.Sprintf("Member login failed for %s: invalid password", req.Email),
			nil, nil)
		respondError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	memberID := utils.GetStringFromMap(foundMember, "id")
	memberEmail := utils.GetStringFromMap(foundMember, "emailId")
	memberName := utils.GetStringFromMap(foundMember, "name")

	firstLogin := false
	if val, ok := foundMember["first_login"].(bool); ok {
		firstLogin = val
	}

	if firstLogin {
		c.JSON(http.StatusForbidden, gin.H{
			"message":             "Password change required",
			"forceChangePassword": true,
			"email":               memberEmail,
		})
		return
	}

	token, expiresIn, err := jwt.GenerateMemberToken(memberID, memberEmail, memberName)
	if err != nil {
		config.Logger.Error("Failed to generate member token", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	isPaid := checkAnnualSubscriptionStatus(memberID)

	userResponse := map[string]interface{}{
		"id":                       foundMember["id"],
		"tagaId":                   foundMember["tagaId"],
		"name":                     foundMember["name"],
		"initial":                  foundMember["initial"],
		"emailId":                  foundMember["emailId"],
		"mobileNumber":             foundMember["mobile_number"],
		"designation":              foundMember["designation"],
		"workingDistrict":          foundMember["working_district"],
		"nativeDistrict":           foundMember["native_district"],
		"recruitmentBatch":         foundMember["recruitment_batch"],
		"seniorityNumber":          foundMember["seniority_number"],
		"dateOfBirth":              foundMember["date_of_birth"],
		"fatherName":               foundMember["father_name"],
		"motherName":               foundMember["mother_name"],
		"educationalQualification": foundMember["educational_qualification"],
		"residentialAddress":       foundMember["residential_address"],
		"permanentAddress":         foundMember["permanent_address"],
		"tbfNumber":                foundMember["tbf_number"],
		"cpsGpfNumber":             foundMember["cps_gpf_number"],
		"firstLogin":               foundMember["first_login"],
		"isPaid":                   isPaid,
		"subscription_active":      isPaid,
		"payment_status":           foundMember["payment_status"],
	}

	respondOK(c, MemberLoginResponse{
		Token:               token,
		Role:                "member",
		ExpiresIn:           expiresIn,
		User:                userResponse,
		ForceChangePassword: false,
	})

	// Audit successful member login (after response so it does not delay the user)
	tagaID := getMemberTagaIdByEmail(req.Email)
	if tagaID == "" {
		tagaID = memberID
	}
	_ = audit.Log(c, tagaID, memberEmail,
		audit.ActionLogin, audit.ModuleAuth,
		"member", tagaID, fmt.Sprintf("Member %s logged in", memberEmail),
		nil, nil)
}

// MemberLogoutHandler
// @Summary Member logout
// @Tags Member Auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/member/logout [post]
func MemberLogoutHandler(c *gin.Context) {
	// Audit logout — resolve tagaId from JWT context
	memberUUID, _ := c.Get("member_id")
	tagaID := "anonymous"
	username := ""
	if memberUUID != nil {
		tagaID = getMemberTagaIdByUUID(fmt.Sprintf("%v", memberUUID))
		if val, ok := c.Get("member_email"); ok {
			if s, ok := val.(string); ok {
				username = s
			}
		}
	}

	_ = audit.Log(c, tagaID, fmt.Sprintf("%v", username),
		audit.ActionLogout, audit.ModuleAuth,
		"member", tagaID, fmt.Sprintf("Member %v logged out", username),
		nil, nil)
	respondMessage(c, "Logged out successfully")
}

// ChangeMemberPasswordHandler
// @Summary Change member password
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param request body object true "Change Password Request"
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/change-password [post]
func ChangeMemberPasswordHandler(c *gin.Context) {
	var req struct {
		Email           string `json:"email" binding:"required"`
		OldPassword     string `json:"oldPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		respondError(c, http.StatusBadRequest, "Passwords do not match")
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read members")
		return
	}

	found := false
	for i, member := range members {
		email := utils.GetStringFromMap(member, "emailId")
		if strings.EqualFold(email, req.Email) {
			storedPassword := utils.GetStringFromMap(member, "password")
			if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.OldPassword)); err != nil {
				respondError(c, http.StatusBadRequest, "Old password is incorrect")
				return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			if err != nil {
				respondError(c, http.StatusInternalServerError, "Failed to hash password")
				return
			}

			members[i]["password"] = string(hashedPassword)
			members[i]["first_login"] = false
			found = true
			break
		}
	}

	if !found {
		respondError(c, http.StatusNotFound, "User not found")
		return
	}

	cfg := config.GetConfig()
	data, _ := json.MarshalIndent(members, "", "  ")
	if err := os.WriteFile(cfg.MembersFile, data, 0644); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to save")
		return
	}

	// Audit password change — never store old/new password
	tagaID := getMemberTagaIdByEmail(req.Email)
	if tagaID == "" {
		tagaID = "anonymous"
	}
	_ = audit.Log(c, tagaID, req.Email,
		audit.ActionPasswordChanged, audit.ModuleAuth,
		"member", tagaID, fmt.Sprintf("Member %s changed their password", req.Email),
		nil, nil)

	respondMessage(c, "Password has been successfully changed")
}
