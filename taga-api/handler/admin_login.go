package handler

import (
	"net/http"
	"taga-api/config"
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
// @Summary      Admin Login
// @Description  Authenticates admin user and returns JWT token for subsequent API calls
// @Tags         Admin Authentication
// @Accept       json
// @Produce      json
// @Param        credentials body AdminLoginRequest true "Admin Login Credentials"
// @Success      200  {object}  AdminLoginResponse
// @Failure      401  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /api/admin/login [post]
func AdminLoginHandler(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := config.GetConfig()
	adminUsername := cfg.AdminEmail
	adminPassword := cfg.AdminPassword

	config.Logger.Info("Admin login attempt",
		zap.String("provided_username", req.Username),
		zap.String("expected_username", adminUsername))

	if req.Username != adminUsername {
		config.Logger.Warn("Failed admin login - invalid username")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if req.Password != adminPassword {
		config.Logger.Warn("Failed admin login - invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token using unified service
	token, expiresIn, err := jwt.GenerateAdminToken(req.Username)
	if err != nil {
		config.Logger.Error("Failed to generate admin token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	config.Logger.Info("Admin login successful",
		zap.String("username", req.Username),
		zap.Int64("expires_in", expiresIn))

	c.JSON(http.StatusOK, AdminLoginResponse{
		Token:     token,
		Role:      "admin",
		ExpiresIn: expiresIn,
	})
}
