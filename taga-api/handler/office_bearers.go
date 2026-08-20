package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"taga-api/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

/*
   Structs
*/

type StateExecutive struct {
	Sno           int    `json:"sno"`
	Name          string `json:"name"`
	Designation   string `json:"designation"`
	Phone         string `json:"phone"`
	Location      string `json:"location"`
	Qualification string `json:"qualification"`
	Experience    int    `json:"experience"`
	Description   string `json:"description"`
	Image         string `json:"image,omitempty"` // optional
}

type DistrictBearer struct {
	Name       string `json:"name"`
	Title      string `json:"title"`
	Department string `json:"department"`
	Contact    string `json:"contact"`
}

func readJSONFile(path string, target interface{}) error {

	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, target)
}

// GetStateExecutive godoc
// @Summary Get State Executive Committee
// @Tags OfficeBearers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} StateExecutive
// @Router /api/office-bearers/state-executive [get]
func GetStateExecutive(c *gin.Context) {
	config.Logger.Info("GetStateExecutive HIT")
	var data []StateExecutive

	filePath := "data/office_bearers/state-executive.json"

	// Check file exists
	file, err := os.ReadFile(filePath)
	if err != nil {
		config.Logger.Error("FILE READ ERROR", zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Debug file content
	config.Logger.Debug("FILE READ SUCCESS")

	// Unmarshal JSON
	err = json.Unmarshal(file, &data)
	if err != nil {
		config.Logger.Error("JSON PARSE ERROR", zap.Error(err))

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	config.Logger.Debug("DATA LOADED", zap.Any("data", data))

	c.JSON(http.StatusOK, data)
}

// GetDistricts godoc
// @Summary Get district list
// @Tags OfficeBearers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} string
// @Router /api/office-bearers/districts [get]
func GetDistricts(c *gin.Context) {

	var districts []string

	err := readJSONFile("data/office_bearers/districts.json", &districts)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load districts",
		})
		return
	}

	c.JSON(http.StatusOK, districts)
}

// GetDistrictOfficeBearers godoc
// @Summary Get district office bearers
// @Tags OfficeBearers
// @Produce json
// @Security BearerAuth
// @Param district query string true "District name"
// @Success 200 {array} DistrictBearer
// @Router /api/office-bearers/district-office-bearers [get]
func GetDistrictOfficeBearers(c *gin.Context) {

	district := c.Query("district")

	var data map[string][]DistrictBearer

	err := readJSONFile("data/office_bearers/district_office_bearers.json", &data)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load data",
		})
		return
	}

	bearers := data[district]

	c.JSON(http.StatusOK, bearers)
}

