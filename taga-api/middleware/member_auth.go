package middleware

import (
	"net/http"
	"strings"
	"taga-api/config"
	"taga-api/service"
	"taga-api/service/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MemberAuthMiddleware validates JWT token for member routes
func MemberAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use Bearer token"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Use unified JWT service
		claims, err := jwt.ValidateMemberToken(tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired. Please login again"})
				c.Abort()
				return
			}
			config.Logger.Error("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		var memberIDStr, emailStr string
		// Set user info in context
		if memberID, ok := claims["member_id"]; ok {
			c.Set("member_id", memberID)
			if s, ok := memberID.(string); ok {
				memberIDStr = s
			}
		}
		if email, ok := claims["email"]; ok {
			c.Set("member_email", email)
			if s, ok := email.(string); ok {
				emailStr = s
			}
		}
		if name, ok := claims["name"]; ok {
			c.Set("member_name", name)
		}
		if username, ok := claims["username"]; ok {
			c.Set("username", username)
		}
		c.Set("role", claims["role"])

		// Administrators have global access — skip member deletion checks
		roleStr, _ := claims["role"].(string)
		if roleStr == "admin" || roleStr == jwt.RoleAdmin {
			c.Next()
			return
		}

		// Check if account has been deleted
		if service.IsMemberDeletedOrNotFound(memberIDStr, emailStr) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Account has been deleted"})
			c.Abort()
			return
		}

		// Check if account deletion is pending
		// Allow checking deletion status and eligibility even when pending
		isStatusEndpoint := strings.HasSuffix(c.Request.URL.Path, "/account-deletion/status") ||
			strings.HasSuffix(c.Request.URL.Path, "/account-deletion/eligibility")
		if !isStatusEndpoint && service.IsMemberDeletionPending(memberIDStr, emailStr) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   service.PendingAccountMessage,
				"message": service.PendingAccountMessage,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
