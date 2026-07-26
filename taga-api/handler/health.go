package handler

import (
	"net/http"
	"taga-api/config"

	"github.com/gin-gonic/gin"
)

// rootHandler handles requests to the root endpoint
// @Summary Get welcome message
// @Description Returns a welcome message from the API
// @Tags general
// @Produce json
// @Success 200 {object} map[string]interface{} "success response"
// @Router / [get]
func RootHandler(c *gin.Context) {
	config.Logger.Info("Root endpoint accessed")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello from Gin + Zap!",
		"status":  "success",
	})
}

// healthHandler handles health check requests
// @Summary Health check
// @Description Returns the health status of the API
// @Tags general
// @Produce json
// @Success 200 {object} map[string]interface{} "health status"
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}


