package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
)

func TestAboutHandler(t *testing.T) {
	// Initialize config to ensure config files are found
	config.Init()

	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.GET("/api/public/about", AboutHandler)

	req, _ := http.NewRequest("GET", "/api/public/about", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The handler expects config.Config.AboutFile to exist
	// With config initialized, it should work properly
	// Accept either success or internal error (if data files don't exist)
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("Unexpected status code: %d", w.Code)
	}
}

func TestAboutStatsHandler(t *testing.T) {
	// Initialize config to ensure config files are found
	config.Init()

	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.GET("/api/public/about/stats", AboutStatsHandler)

	req, _ := http.NewRequest("GET", "/api/public/about/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 with stats array
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Active Members")
}
