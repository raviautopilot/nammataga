package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
)

func TestRootHandler(t *testing.T) {
	// Initialize config and logger
	config.Init()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a router
	router := gin.Default()
	router.GET("/", RootHandler)

	// Create a test request
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response body contains expected fields
	assert.Contains(t, w.Body.String(), "message")
	assert.Contains(t, w.Body.String(), "status")
}

func TestHealthHandler(t *testing.T) {
	// Initialize config and logger
	config.Init()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a router
	router := gin.Default()
	router.GET("/health", HealthHandler)

	// Create a test request
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response body
	assert.Contains(t, w.Body.String(), "healthy")
}
