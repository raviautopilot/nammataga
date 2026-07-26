package handler

import (
	"net/http"

	"taga-api/model"
	"taga-api/service"

	"github.com/gin-gonic/gin"
)

// ApplyMembershipHandler godoc
// @Summary Apply for Membership
// @Description Submit a new TAGA membership application
// @Tags Membership
// @Accept json
// @Produce json
// @Param request body model.Membership true "Membership Application Data"
// @Success 200 {object} map[string]string "Application submitted successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/membership/apply [post]
func ApplyMembershipHandler(c *gin.Context) {
	var data model.Membership

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	err := service.ApplyMembership(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Application submitted",
	})
}

// GetMembershipListHandler godoc
// @Summary Get Membership List
// @Description Retrieve all membership applications (Admin purpose)
// @Tags Membership
// @Produce json
// @Success 200 {array} model.Membership "List of memberships"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/membership/list [get]
func GetMembershipListHandler(c *gin.Context) {
	data, err := service.GetMembershipList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetMembershipDistricts godoc
// @Summary Get Districts
// @Description Retrieve list of districts for membership form
// @Tags Membership
// @Produce json
// @Success 200 {array} string "List of districts"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/membership/districts [get]
func GetMembershipDistricts(c *gin.Context) {
	data, err := service.GetMembershipDistricts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, data)
}
