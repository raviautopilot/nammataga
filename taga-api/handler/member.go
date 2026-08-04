package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"taga-api/config"
	"taga-api/model"
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

// checkAnnualSubscriptionStatus determines if the member has an active annual subscription
// or is within the renewal grace period (April 1 - May 31).
// Special rule: All members are considered paid for 2026-2027. Real checks begin April 1, 2027.
func checkAnnualSubscriptionStatus(memberID string) bool {
	now := time.Now()

	// 2026-2027 waiver: everyone is paid until May 31, 2027
	firstYearGraceEnd := time.Date(2027, 5, 31, 23, 59, 59, 0, now.Location())
	if now.Before(firstYearGraceEnd) || now.Equal(firstYearGraceEnd) {
		return true
	}

	// After May 31, 2027: normal grace period (April 1 – May 31 each year)
	if isInGracePeriod(now) {
		return true
	}

	// 🔥 IMPORTANT: Convert the incoming UUID to tagaId for subscription lookup
	// Because subscriptions now store tagaId, not UUID
	tagaId := getMemberTagaIdByUUID(memberID)

	// Outside grace period: check for a valid active annual subscription
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
		// Compare tagaId with tagaId (both now use the same identifier)
		if sub.MemberID == tagaId &&
			sub.SubscriptionID == "annual-subscription" &&
			sub.Status == "active" &&
			now.Before(sub.EndDate) {
			return true
		}
	}
	return false
}

// isInGracePeriod checks if the current date falls within the renewal grace period (April 1 – May 31).
func isInGracePeriod(now time.Time) bool {
	year := now.Year()
	graceStart := time.Date(year, 4, 1, 0, 0, 0, 0, now.Location())
	graceEnd := time.Date(year, 5, 31, 23, 59, 59, 0, now.Location())
	return (now.Equal(graceStart) || now.After(graceStart)) && (now.Equal(graceEnd) || now.Before(graceEnd))
}

// MemberLoginHandler - JWT based login for members
// @Summary Member Login
// @Description Authenticates member and returns JWT token
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param credentials body MemberLoginRequest true "Member credentials"
// @Success 200 {object} MemberLoginResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/member/login [post]
func MemberLoginHandler(c *gin.Context) {
	var req MemberLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Read members
	members, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	// Find member by email
	var foundMember map[string]interface{}
	for _, member := range members {
		email, ok := member["emailId"].(string)
		if ok && strings.EqualFold(email, req.Email) {
			foundMember = member
			break
		}
	}

	if foundMember == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password using bcrypt
	storedPassword, ok := foundMember["password"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		config.Logger.Warn("Invalid password for member", zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Extract member details for token using utils helper
	memberID := utils.GetStringFromMap(foundMember, "id")
	memberEmail := utils.GetStringFromMap(foundMember, "emailId")
	memberName := utils.GetStringFromMap(foundMember, "name")
	// Get first_login flag
	firstLogin := false

	if val, ok := foundMember["first_login"].(bool); ok {
		firstLogin = val
	}

	// BLOCK FIRST LOGIN
	if firstLogin {

		c.JSON(http.StatusForbidden, gin.H{
			"message":             "Password change required",
			"forceChangePassword": true,
			"email":               memberEmail,
		})

		return
	}
	// Generate JWT token using unified service
	token, expiresIn, err := jwt.GenerateMemberToken(memberID, memberEmail, memberName)
	if err != nil {
		config.Logger.Error("Failed to generate member token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 🔥 Check if member has active Annual Subscription (with grace period)
	isPaid := checkAnnualSubscriptionStatus(memberID)
	subscriptionActive := isPaid

	// Prepare user response (excluding sensitive data)
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
		"subscription_active":      subscriptionActive,
		"payment_status":           foundMember["payment_status"],
	}

	// Get first_login flag
	// firstLogin := false
	// if val, ok := foundMember["first_login"].(bool); ok {
	// 	firstLogin = val
	// }

	// // Get login_count
	// loginCount := 0
	// if val, ok := foundMember["login_count"].(float64); ok {
	// 	loginCount = int(val)
	// }

	// // Increment login count
	// loginCount++
	// foundMember["login_count"] = loginCount

	// // Force change logic
	// forceChange := false
	// if firstLogin && loginCount >= 2 {
	// 	forceChange = true
	// }

	// // ✅ SAVE updated login_count to JSON file
	// cfg := config.GetConfig()
	// data, _ := json.MarshalIndent(members, "", "  ")
	// _ = os.WriteFile(cfg.MembersFile, data, 0644)

	c.JSON(http.StatusOK, MemberLoginResponse{
		Token:               token,
		Role:                "member",
		ExpiresIn:           expiresIn,
		User:                userResponse,
		ForceChangePassword: false,
	})
}

// GetMemberProfileByToken - JWT based profile fetch
// @Summary Get member profile
// @Description Fetch logged-in member profile using JWT token
// @Tags Member Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/member/profile [get]
func GetMemberProfileByToken(c *gin.Context) {
	memberID, exists := c.Get("member_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		config.Logger.Error("Failed to read members", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	}

	for _, member := range members {
		if id, ok := member["id"].(string); ok && id == memberID {
			delete(member, "password")
			member["isPaid"] = checkAnnualSubscriptionStatus(memberID.(string))
			member["subscription_active"] = member["isPaid"]
			c.JSON(http.StatusOK, gin.H{"user": member})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
}

// GetMemberProfile
// @Summary Get member profile by email
// @Description Fetch member profile using email
// @Tags Member Profile
// @Accept json
// @Produce json
// @Param email query string true "Member Email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/member/profile-by-email [get]


// MemberLogoutHandler
// @Summary Member logout
// @Description Logout member successfully
// @Tags Member Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/member/logout [post]
func MemberLogoutHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}



// UpdateMemberProfileHandler
// @Summary Update member profile
// @Description Update logged-in member profile
// @Tags Member Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Profile Update Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/member/profile [put]
func UpdateMemberProfileHandler(c *gin.Context) {
	memberID, exists := c.Get("member_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	found := false
	for i, member := range members {
		if id, ok := member["id"].(string); ok && id == memberID {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	cfg := config.GetConfig()
	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save"})
		return
	}

	if err := os.WriteFile(cfg.MembersFile, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// ChangeMemberPasswordHandler
// @Summary Change member password
// @Description Change member password using old password verification
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param request body object true "Change Password Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/auth/change-password [post]
func ChangeMemberPasswordHandler(c *gin.Context) {
	var req struct {
		Email           string `json:"email" binding:"required"`
		OldPassword     string `json:"oldPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	members, err := readExistingMembers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read members"})
		return
	}

	found := false
	for i, member := range members {
		email := utils.GetStringFromMap(member, "emailId")
		if strings.EqualFold(email, req.Email) {
			storedPassword := utils.GetStringFromMap(member, "password")
			if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.OldPassword)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Old password is incorrect"})
				return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}

			members[i]["password"] = string(hashedPassword)
			members[i]["first_login"] = false
			// members[i]["login_count"] = 0
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	cfg := config.GetConfig()
	data, _ := json.MarshalIndent(members, "", "  ")
	if err := os.WriteFile(cfg.MembersFile, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been successfully changed"})
}
