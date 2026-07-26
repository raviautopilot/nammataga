package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetLogo godoc
// @Summary Get application logo
// @Description Returns the URL of the application logo
// @Tags office
// @Produce json
// @Success 200 {object} map[string]string "Logo URL response"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/logo [get]
func GetLogo(c *gin.Context) {

	logoURL := "/images/logo/logo1.jpg"

	c.JSON(http.StatusOK, gin.H{
		"url": logoURL,
	})
}
