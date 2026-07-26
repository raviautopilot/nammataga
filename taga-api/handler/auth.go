package handler

import (
	"net/http"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResetPasswordHandler handles password reset
// @Summary Reset Password
// @Description Resets the password using a valid reset token
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param resetRequest body model.ResetPasswordRequest true "Reset details"
// @Success 200 {object} map[string]interface{} "Password reset successful"
// @Failure 400 {object} map[string]interface{} "Invalid request payload"
// @Failure 401 {object} map[string]interface{} "Invalid or expired token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/auth/reset-password [post]
func ResetPasswordHandler(c *gin.Context) {
	var resetRequest model.ResetPasswordRequest

	if err := c.ShouldBindJSON(&resetRequest); err != nil {
		config.Logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := auth.ResetPassword(resetRequest.Email, resetRequest.OldPassword, resetRequest.NewPassword)
	if err != nil {
		if err.Error() == "invalid or expired token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}
		config.Logger.Error("Password reset failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Password reset failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successful",
	})
}

// ForgotPasswordHandler handles password reset requests
// @Summary Forgot Password
// @Description Initiates a password reset process
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param forgotPasswordRequest body model.ForgotPasswordRequest true "Email address"
// @Success 200 {object} map[string]interface{} "Password reset email sent"
// @Failure 400 {object} map[string]interface{} "Invalid request payload"
// @Failure 404 {object} map[string]interface{} "Email not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/auth/forgot-password [post]
func ForgotPasswordHandler(c *gin.Context) {
	var forgotPasswordRequest model.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		config.Logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := auth.ForgotPassword(forgotPasswordRequest.Email)
	if err != nil {
		if err.Error() == "email not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Email not found",
			})
			return
		}
		config.Logger.Error("Forgot password failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process request",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// MemberForgotPasswordHandler godoc
// @Summary Member Forgot Password
// @Description Send reset email
// @Tags Member Auth
// @Accept json
// @Produce json
// @Param request body model.ForgotPasswordRequest true "Forgot Password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/auth/member-forgot-password [post]
func MemberForgotPasswordHandler(c *gin.Context) {
	var data model.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// ✅ USE auth service
	err := auth.ForgotPassword(data.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent",
	})
}


