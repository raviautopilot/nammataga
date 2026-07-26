package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"taga-api/config"
)

// Custom Gin middleware for Zap logging
func GinZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		config.Logger.Info("Request processed",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}
