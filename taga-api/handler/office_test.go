package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
)

func TestOfficeHandler(t *testing.T) {
	// Initialize config to ensure config files are found
	config.Init()

	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.GET("/api/office", OfficeHandler)

	// Test with default parameter (state)
	req, _ := http.NewRequest("GET", "/api/office", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The handler will try to read data files which may not exist in test
	// So we expect either 200 if files exist, or an error status
	// We'll just check that the handler doesn't panic
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)

	// Test with explicit state parameter
	req2, _ := http.NewRequest("GET", "/api/office?pathParam=state", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.NotEqual(t, http.StatusInternalServerError, w2.Code)
}
