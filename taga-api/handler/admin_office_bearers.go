package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"taga-api/config"
	"taga-api/service/office_bearers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetAllDistrictsHandler godoc
// @Summary Get all districts for office bearers management
// @Description Returns a list of all districts that have office bearers data
// @Tags Admin Office Bearers
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "districts list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/office-bearers/districts [get]
func GetAllDistrictsHandler(c *gin.Context) {
	config.Logger.Info("Fetching all districts for office bearers management")

	districts, err := office_bearers.GetAllDistricts()
	if err != nil {
		config.Logger.Error("Failed to fetch districts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch districts: " + err.Error(),
		})
		return
	}

	config.Logger.Info("Successfully fetched districts", zap.Int("count", len(districts)))
	c.JSON(http.StatusOK, gin.H{
		"districts": districts,
	})
}

// GetDistrictBearersHandler godoc
// @Summary Get office bearers for a specific district
// @Description Returns all 6 office bearers for the specified district
// @Tags Admin Office Bearers
// @Produce json
// @Security BearerAuth
// @Param district path string true "District name"
// @Success 200 {object} map[string]interface{} "bearers list"
// @Failure 400 {object} map[string]string "Invalid district name"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/office-bearers/district/{district} [get]
func GetDistrictBearersHandler(c *gin.Context) {
	district := c.Param("district")

	// URL decode and clean the district name
	district = strings.TrimSpace(district)

	if district == "" {
		config.Logger.Warn("Empty district name provided")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "District name cannot be empty",
		})
		return
	}

	config.Logger.Info("Fetching district bearers",
		zap.String("district", district))

	bearers, err := office_bearers.GetDistrictBearers(district)
	if err != nil {
		config.Logger.Error("Failed to fetch district bearers",
			zap.String("district", district),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch district bearers: " + err.Error(),
		})
		return
	}

	// Validate we have exactly 6 bearers
	if len(bearers) != 6 {
		config.Logger.Warn("District bearers count mismatch",
			zap.String("district", district),
			zap.Int("count", len(bearers)))
	}

	config.Logger.Info("Successfully fetched district bearers",
		zap.String("district", district),
		zap.Int("count", len(bearers)))

	c.JSON(http.StatusOK, gin.H{
		"bearers": bearers,
	})
}

// UpdateDistrictBearersHandler godoc
// @Summary Update all office bearers for a specific district
// @Description Updates all 6 office bearers for the specified district. Creates backup before update.
// @Tags Admin Office Bearers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param district path string true "District name"
// @Param bearers body []office_bearers.DistrictBearer true "Array of 6 bearers"
// @Success 200 {object} map[string]interface{} "success message with backup path"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "District not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/office-bearers/district/{district} [put]
func UpdateDistrictBearersHandler(c *gin.Context) {
	district := c.Param("district")
	district = strings.TrimSpace(district)

	if district == "" {
		config.Logger.Warn("Empty district name provided for update")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "District name cannot be empty",
		})
		return
	}

	// Parse request body
	var bearers []office_bearers.DistrictBearer
	if err := c.ShouldBindJSON(&bearers); err != nil {
		config.Logger.Error("Failed to parse request body",
			zap.String("district", district),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate exactly 6 bearers
	if len(bearers) != 6 {
		config.Logger.Warn("Invalid number of bearers",
			zap.String("district", district),
			zap.Int("received", len(bearers)),
			zap.Int("expected", 6))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Exactly 6 bearers are required for a district",
		})
		return
	}

	// Validate each bearer has required fields
	for i, bearer := range bearers {
		// Name and Contact can be empty, but Title must be present
		if bearer.Title == "" {
			config.Logger.Warn("Bearer missing title",
				zap.String("district", district),
				zap.Int("index", i))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "All bearers must have a title",
			})
			return
		}
	}

	config.Logger.Info("Updating district bearers",
		zap.String("district", district),
		zap.Int("bearers_count", len(bearers)))

	// Perform update
	err := office_bearers.UpdateDistrictBearers(district, bearers)
	if err != nil {
		config.Logger.Error("Failed to update district bearers",
			zap.String("district", district),
			zap.Error(err))

		// Handle specific error types
		if strings.Contains(err.Error(), "exactly 6 bearers") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update district bearers: " + err.Error(),
		})
		return
	}

	// Get backup path from the last backup (indicated by success)
	config.Logger.Info("Successfully updated district bearers",
		zap.String("district", district))

	c.JSON(http.StatusOK, gin.H{
		"message":  "District bearers updated successfully",
		"district": district,
		"bearers":  bearers,
	})
}

// RestoreBackupHandler godoc
// @Summary Restore district bearers from a backup file
// @Description Restores the district office bearers data from a specified backup file
// @Tags Admin Office Bearers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Backup file path"
// @Success 200 {object} map[string]string "success message"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/office-bearers/backup/restore [post]
func RestoreBackupHandler(c *gin.Context) {
	var request struct {
		BackupFile string `json:"backup_file" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		config.Logger.Error("Failed to parse restore request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: backup_file is required",
		})
		return
	}

	if request.BackupFile == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "backup_file cannot be empty",
		})
		return
	}

	config.Logger.Info("Restoring from backup",
		zap.String("backup_file", request.BackupFile))

	err := office_bearers.RestoreFromBackup(request.BackupFile)
	if err != nil {
		config.Logger.Error("Failed to restore from backup",
			zap.String("backup_file", request.BackupFile),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restore from backup: " + err.Error(),
		})
		return
	}

	config.Logger.Info("Successfully restored from backup",
		zap.String("backup_file", request.BackupFile))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Backup restored successfully",
		"backup_file": request.BackupFile,
	})
}

// ListBackupsHandler godoc
// @Summary List all available backup files
// @Description Returns a list of all backup files for district office bearers
// @Tags Admin Office Bearers
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "list of backup files"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/admin/office-bearers/backups [get]
func ListBackupsHandler(c *gin.Context) {
	backupDir := "data/office_bearers/backups"

	// Check if directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"backups": []string{},
		})
		return
	}

	// Read directory
	files, err := os.ReadDir(backupDir)
	if err != nil {
		config.Logger.Error("Failed to list backups", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list backups: " + err.Error(),
		})
		return
	}

	// Filter backup files
	backups := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "district_office_bearers_backup_") && strings.HasSuffix(file.Name(), ".json") {
			backups = append(backups, filepath.Join(backupDir, file.Name()))
		}
	}

	// Sort by name (which includes timestamp) in descending order (newest first)
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	config.Logger.Info("Listed backups", zap.Int("count", len(backups)))
	c.JSON(http.StatusOK, gin.H{
		"backups": backups,
	})
}
