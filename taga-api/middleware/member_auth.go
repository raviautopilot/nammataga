package middleware

import (
	"net/http"
	"strings"
	"taga-api/config"
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

		// Set user info in context
		if memberID, ok := claims["member_id"]; ok {
			c.Set("member_id", memberID)
		}
		if email, ok := claims["email"]; ok {
			c.Set("member_email", email)
		}
		if name, ok := claims["name"]; ok {
			c.Set("member_name", name)
		}
		c.Set("role", claims["role"])
		c.Next()
	}
}
