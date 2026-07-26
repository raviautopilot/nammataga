package middleware

import (
	"net/http"
	"strings"
	"taga-api/config"
	"taga-api/service/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminAuthMiddleware validates JWT token and checks if user is admin
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			config.Logger.Warn("No authorization header provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			config.Logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use Bearer token"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Use unified JWT service
		claims, err := jwt.ValidateAdminToken(tokenString)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				config.Logger.Warn("Token expired")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired. Please login again"})
				c.Abort()
				return
			}
			config.Logger.Error("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		if username, ok := claims["username"]; ok {
			c.Set("username", username)
		}
		if userID, ok := claims["userID"]; ok {
			c.Set("userID", userID)
		}
		c.Set("role", claims["role"])
		c.Next()
	}
}


