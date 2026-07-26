package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"taga-api/config"
	"taga-api/model"
	"taga-api/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AboutHandler handles requests to the about endpoint
// @Summary Get organization information
// @Description Returns detailed information about TAGA organization
// @Tags about
// @Produce json
// @Success 200 {object} model.AboutResponse "Organization information"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/about [get]
func AboutHandler(c *gin.Context) {
	// Read the about data from external file
	data, err := os.ReadFile(config.Config.AboutFile)
	if err != nil {
		config.Logger.Error("Failed to read about data file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load organization information",
		})
		return
	}

	var aboutData model.AboutResponse
	err = json.Unmarshal(data, &aboutData)
	if err != nil {
		config.Logger.Error("Failed to parse about data", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse organization information",
		})
		return
	}

	config.Logger.Info("About endpoint accessed")
	c.JSON(http.StatusOK, aboutData)
}

// GetStats retrieves organization statistics
// @Summary Get organization statistics
// @Description Returns key statistics and metrics about TAGA organization including member count, years of service, districts covered, and programs/events
// @Tags about
// @Produce json
// @Success 200 {array} model.StatsResponse "List of organization statistics"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/about/stats [get]
func AboutStatsHandler(c *gin.Context) {
	var stats []model.StatsResponse

	err := utils.ReadJSONFile(config.Config.StatsFile, &stats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AboutObjectivesHandler handles requests to the about objectives endpoint
// @Summary Get organization objectives
// @Description Returns the list of TAGA organization objectives
// @Tags about
// @Produce json
// @Success 200 {array} model.Objective "List of organization objectives"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/about/objectives [get]
func AboutObjectivesHandler(c *gin.Context) {
	var objectives []model.Objective
	err := utils.ReadJSONFile(config.Config.ObjectivesFile, &objectives)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load objectives information",
		})
		return
	}

	config.Logger.Info("About objectives endpoint accessed",
		zap.Int("objectives_count", len(objectives)))
	c.JSON(http.StatusOK, objectives)
}

// AboutServicesHandler retrieves the list of services offered by the organization.
// @Summary Get organization services
// @Description Returns a list of services provided by the organization, including their title, description, and category.
// @Tags about
// @Produce json
// @Success 200 {array} model.ServiceResponse "List of organization services"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/about/services [get]
func AboutServicesHandler(c *gin.Context) {
	var services []model.ServiceResponse
	err := utils.ReadJSONFile(config.Config.ServicesFile, &services)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load services information",
		})
		return
	}

	config.Logger.Info("About services endpoint accessed",
		zap.Int("services_count", len(services)))
	c.JSON(http.StatusOK, services)
}

// AboutContactHandler retrieves all contact information for the association.
// @Summary Get contact information
// @Description Fetches all contact details including headquarters, office hours, and regional offices.
// @Tags about
// @Produce json
// @Success 200 {object} model.ContactResponse "Successful response with contact information"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/public/about/contact [get]
func AboutContactHandler(c *gin.Context) {
	var contactInfo model.ContactResponse
	err := utils.ReadJSONFile(config.Config.ContactFile, &contactInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load contact information",
		})
		return
	}

	if contactInfo.PrimaryEmail == "" {
		config.Logger.Warn("Email missing from contact information")
	}
	config.Logger.Info("About contact endpoint accessed")
	c.JSON(http.StatusOK, contactInfo)
}

