package handler

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"os"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin@taga.com"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

type AdminLoginResponse struct {
	Token     string `json:"token"`
	Role      string `json:"role"`
	ExpiresIn int64  `json:"expires_in"`
}

// AdminLoginHandler godoc
// @Summary Admin Login
// @Tags Admin Authentication
// @Accept json
// @Produce json
// @Param credentials body AdminLoginRequest true "Admin Login Credentials"
// @Success 200 {object} AdminLoginResponse
// @Router /api/admin/login [post]
func AdminLoginHandler(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	cfg := config.GetConfig()
	adminUsername := cfg.AdminEmail
	adminPassword := cfg.AdminPassword

	config.Logger.Info("Admin login attempt",
		zap.String("provided_username", req.Username),
		zap.String("expected_username", adminUsername))

	if req.Username != adminUsername || req.Password != adminPassword {
		config.Logger.Warn("Failed admin login - invalid credentials")
		respondError(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, expiresIn, err := jwt.GenerateAdminToken(req.Username)
	if err != nil {
		config.Logger.Error("Failed to generate admin token", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	config.Logger.Info("Admin login successful",
		zap.String("username", req.Username),
		zap.Int64("expires_in", expiresIn))

	respondOK(c, AdminLoginResponse{
		Token:     token,
		Role:      "admin",
		ExpiresIn: expiresIn,
	})
}

// InitPassword godoc
// @Summary Initialize member password(s)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param memberId query string false "Member ID or keyword: all | none (default)"
// @Success 200 {string} string
// @Router /api/admin/init-password [post]
func InitPassword(c *gin.Context) {
	memberId := c.DefaultQuery("memberId", "none")

	if memberId == "none" {
		respondError(c, http.StatusBadRequest, "Please specify memberId parameter. Use 'all' to reset all passwords or a specific member ID")
		return
	}

	if memberId == "all" {
		go resetPasswordForAll()
		respondMessage(c, "Password reset process started for all members. This may take a moment.")
		return
	}

	respondError(c, http.StatusNotImplemented, "Reset password for specific member is not yet implemented")
}

func resetPasswordForAll() {
	const filePath = "data/member/members.json"

	b, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	var members []model.Member
	if err := json.Unmarshal(b, &members); err != nil {
		panic(err)
	}

	for i := range members {
		members[i].Username = members[i].MobileNumber
		plain := genRandomPassword(8)
		members[i].Password = plain
		members[i].FirstLogin = true
	}

	out, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(filePath, out, 0644); err != nil {
		panic(err)
	}
}

func genRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "Default123!"
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}
