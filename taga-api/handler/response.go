package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError sends a JSON error response with the given status code and message.
func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// respondOK sends a HTTP 200 JSON response with the provided payload.
func respondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// respondMessage sends a HTTP 200 JSON response with a success message.
func respondMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}
