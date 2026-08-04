package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"taga-api/config"
	"taga-api/model"
	of "taga-api/service/office"
	"taga-api/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RootHandler handles requests to the root endpoint
// @Summary Get welcome message
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

// HealthHandler handles health check requests
// @Summary Health check
// @Tags general
// @Produce json
// @Success 200 {object} map[string]interface{} "health status"
// @Router /health [get]
func HealthHandler(c *gin.Context) {
	respondMessage(c, "healthy")
}

// AboutHandler handles requests to the about endpoint
// @Summary Get organization information
// @Tags about
// @Produce json
// @Success 200 {object} model.AboutResponse
// @Router /api/public/about [get]
func AboutHandler(c *gin.Context) {
	data, err := os.ReadFile(config.Config.AboutFile)
	if err != nil {
		config.Logger.Error("Failed to read about data file", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to load organization information")
		return
	}

	var aboutData model.AboutResponse
	if err = json.Unmarshal(data, &aboutData); err != nil {
		config.Logger.Error("Failed to parse about data", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "Failed to parse organization information")
		return
	}

	config.Logger.Info("About endpoint accessed")
	respondOK(c, aboutData)
}

// AboutStatsHandler retrieves organization statistics
// @Summary Get organization statistics
// @Tags about
// @Produce json
// @Success 200 {array} model.StatsResponse
// @Router /api/public/about/stats [get]
func AboutStatsHandler(c *gin.Context) {
	var stats []model.StatsResponse
	if err := utils.ReadJSONFile(config.Config.StatsFile, &stats); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load stats")
		return
	}
	respondOK(c, stats)
}

// AboutObjectivesHandler handles requests to the about objectives endpoint
// @Summary Get organization objectives
// @Tags about
// @Produce json
// @Success 200 {array} model.Objective
// @Router /api/public/about/objectives [get]
func AboutObjectivesHandler(c *gin.Context) {
	var objectives []model.Objective
	if err := utils.ReadJSONFile(config.Config.ObjectivesFile, &objectives); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load objectives information")
		return
	}
	config.Logger.Info("About objectives endpoint accessed", zap.Int("objectives_count", len(objectives)))
	respondOK(c, objectives)
}

// AboutServicesHandler retrieves the list of services offered by the organization.
// @Summary Get organization services
// @Tags about
// @Produce json
// @Success 200 {array} model.ServiceResponse
// @Router /api/public/about/services [get]
func AboutServicesHandler(c *gin.Context) {
	var services []model.ServiceResponse
	if err := utils.ReadJSONFile(config.Config.ServicesFile, &services); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load services information")
		return
	}
	config.Logger.Info("About services endpoint accessed", zap.Int("services_count", len(services)))
	respondOK(c, services)
}

// AboutContactHandler retrieves all contact information for the association.
// @Summary Get contact information
// @Tags about
// @Produce json
// @Success 200 {object} model.ContactResponse
// @Router /api/public/about/contact [get]
func AboutContactHandler(c *gin.Context) {
	var contactInfo model.ContactResponse
	if err := utils.ReadJSONFile(config.Config.ContactFile, &contactInfo); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load contact information")
		return
	}
	if contactInfo.PrimaryEmail == "" {
		config.Logger.Warn("Email missing from contact information")
	}
	config.Logger.Info("About contact endpoint accessed")
	respondOK(c, contactInfo)
}

// GetLogo returns application logo path
// @Summary Get application logo
// @Tags office
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/logo [get]
func GetLogo(c *gin.Context) {
	respondOK(c, gin.H{"url": "/images/logo/logo1.jpg"})
}

// OfficeHandler retrieves office bearer information
// @Summary Get office bearer information
// @Tags office
// @Produce json
// @Param pathParam query string true "Office type (state or district name)" default(state)
// @Success 200 {object} interface{}
// @Router /api/office [get]
func OfficeHandler(c *gin.Context) {
	pathParam := c.DefaultQuery("pathParam", "state")

	data, err := of.GetOfficeData(pathParam)
	if err != nil {
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "is required") || strings.Contains(errorMsg, "Invalid") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   errorMsg,
				"message": "Please provide a valid office type. Use 'state' for state officers or a district name for district officers",
			})
			return
		}

		if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "no officers found") || (strings.Contains(errorMsg, "district") && strings.Contains(errorMsg, "not found")) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   errorMsg,
				"message": "The requested office data could not be found. Please check the office type or district name",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   errorMsg,
			"message": "An internal server error occurred while processing your request",
		})
		return
	}

	respondOK(c, gin.H{
		"data": data,
		"type": pathParam,
	})
}
