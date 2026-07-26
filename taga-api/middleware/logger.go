package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"taga-api/config"
)

// Custom Gin middleware for Zap logging
func GinZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request entry
		config.Logger.Info("Request accessed",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		// Retrieve authenticated user info from context if present
		who := "anonymous"
		if username, ok := c.Get("username"); ok {
			who = fmt.Sprintf("admin:%v", username)
		} else if memberEmail, ok := c.Get("member_email"); ok {
			who = fmt.Sprintf("member:%v", memberEmail)
		} else if memberID, ok := c.Get("member_id"); ok {
			who = fmt.Sprintf("member_id:%v", memberID)
		}

		roleStr := "none"
		if role, ok := c.Get("role"); ok {
			roleStr = fmt.Sprintf("%v", role)
		}

		latency := time.Since(startTime)

		// Log request exit with status, duration, and caller details
		config.Logger.Info("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("who", who),
			zap.String("role", roleStr),
			zap.Duration("latency", latency),
		)
	}
}
