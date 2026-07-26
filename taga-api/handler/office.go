package handler

import (
	"net/http"
	"strings"
	of "taga-api/service/office"

	"github.com/gin-gonic/gin"
)

// OfficeHandler godoc
// @Summary Get office bearer information
// @Description Retrieve office bearer data based on the path parameter. Use 'state' for state officers or a district name for district officers
// @Tags office
// @Produce json
// @Param pathParam query string true "Office type (state or district name)" default(state)
// @Success 200 {object} interface{} "Office data"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Office information not found"
// @Failure 500 {object} map[string]string "Failed to load office information"
// @Router /api/office [get]
func OfficeHandler(c *gin.Context) {
	pathParam := c.DefaultQuery("pathParam", "state")

	data, err := of.GetOfficeData(pathParam)
	if err != nil {
		// Check error type to determine appropriate status code
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "is required") ||
			strings.Contains(errorMsg, "Invalid") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   errorMsg,
				"message": "Please provide a valid office type. Use 'state' for state officers or a district name for district officers",
			})
			return
		}

		if strings.Contains(errorMsg, "not found") ||
			strings.Contains(errorMsg, "no officers found") ||
			strings.Contains(errorMsg, "district") && strings.Contains(errorMsg, "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   errorMsg,
				"message": "The requested office data could not be found. Please check the office type or district name",
			})
			return
		}

		// For other errors, return 500
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errorMsg,
			"message": "An internal server error occurred while processing your request",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"type": pathParam,
	})
}
