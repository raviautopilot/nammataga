package office_bearers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"taga-api/config"

	"go.uber.org/zap"
)

// DistrictBearer represents a single office bearer in a district
type DistrictBearer struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Contact string `json:"contact"`
}

// DistrictBearersData represents the entire JSON file structure
type DistrictBearersData map[string][]DistrictBearer

var (
	mu        sync.Mutex
	dataFile  = "data/office_bearers/district_office_bearers.json"
	backupDir = "data/office_bearers/backups"
)

// GetAllDistricts returns a sorted list of all district names
func GetAllDistricts() ([]string, error) {
	mu.Lock()
	defer mu.Unlock()

	data, err := readDataFile()
	if err != nil {
		return nil, err
	}

	districts := make([]string, 0, len(data))
	for district := range data {
		districts = append(districts, district)
	}

	sort.Strings(districts)
	return districts, nil
}

// GetDistrictBearers returns bearers for a specific district, ensuring exactly 6 bearers
func GetDistrictBearers(districtName string) ([]DistrictBearer, error) {
	mu.Lock()
	defer mu.Unlock()

	if districtName == "" {
		return nil, fmt.Errorf("district name cannot be empty")
	}

	data, err := readDataFile()
	if err != nil {
		return nil, err
	}

	bearers, exists := data[districtName]
	if !exists {
		// Return empty 6-position structure for new district
		return createEmptyBearers(), nil
	}

	// Ensure exactly 6 bearers
	return normalizeBearers(bearers), nil
}

// UpdateDistrictBearers updates all bearers for a specific district
func UpdateDistrictBearers(districtName string, bearers []DistrictBearer) error {
	mu.Lock()
	defer mu.Unlock()

	if districtName == "" {
		return fmt.Errorf("district name cannot be empty")
	}

	if len(bearers) != 6 {
		return fmt.Errorf("district must have exactly 6 bearers, got %d", len(bearers))
	}

	// Create backup before making changes
	backupPath, err := createBackupLocked()
	if err != nil {
		config.Logger.Error("Failed to create backup before update",
			zap.String("district", districtName),
			zap.Error(err))
		return fmt.Errorf("failed to create backup: %w", err)
	}
	config.Logger.Info("Backup created",
		zap.String("district", districtName),
		zap.String("backup_path", backupPath))

	// Read current data
	data, err := readDataFileLocked()
	if err != nil {
		return err
	}

	// Update or add district
	data[districtName] = bearers

	// Write back to file
	if err := writeDataFileLocked(data); err != nil {
		// Attempt to restore from backup on write failure
		config.Logger.Error("Write failed, attempting restore",
			zap.String("district", districtName),
			zap.Error(err))
		restoreFromBackupLocked(backupPath)
		return fmt.Errorf("failed to write data: %w", err)
	}

	config.Logger.Info("District bearers updated successfully",
		zap.String("district", districtName),
		zap.String("backup_path", backupPath))

	return nil
}

// RestoreBackupHandler triggers this to restore data from a backup file
func RestoreFromBackup(backupFile string) error {
	mu.Lock()
	defer mu.Unlock()
	return restoreFromBackupLocked(backupFile)
}

// ==================== PRIVATE HELPER FUNCTIONS ====================

// createBackupLocked creates backup (must be called with lock held)
func createBackupLocked() (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Check if source file exists
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		// No file to backup, create empty data structure
		emptyData := make(DistrictBearersData)
		if err := writeDataFileLocked(emptyData); err != nil {
			return "", fmt.Errorf("failed to create empty data file: %w", err)
		}
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("district_office_bearers_backup_%s.json", timestamp))

	// Read source file
	sourceData, err := os.ReadFile(dataFile)
	if err != nil {
		return "", fmt.Errorf("failed to read source file: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupFile, sourceData, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	return backupFile, nil
}

// restoreFromBackupLocked restores from backup (must be called with lock held)
func restoreFromBackupLocked(backupFile string) error {
	// Read backup file
	backupData, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Validate JSON structure
	var testData DistrictBearersData
	if err := json.Unmarshal(backupData, &testData); err != nil {
		return fmt.Errorf("backup file contains invalid JSON: %w", err)
	}

	// Create backup of current state before restore (just in case)
	currentBackup, err := createBackupLocked()
	if err != nil {
		config.Logger.Warn("Failed to create pre-restore backup", zap.Error(err))
	}

	// Write backup data to main file
	if err := os.WriteFile(dataFile, backupData, 0644); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	config.Logger.Info("Backup restored successfully",
		zap.String("backup_file", backupFile),
		zap.String("pre_restore_backup", currentBackup))

	return nil
}

// readDataFile reads and parses the JSON file (must be called with lock held)
func readDataFileLocked() (DistrictBearersData, error) {
	var data DistrictBearersData

	// Check if file exists
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		// Create empty data structure
		data = make(DistrictBearersData)
		if err := writeDataFileLocked(data); err != nil {
			return nil, err
		}
		return data, nil
	}

	// Read file
	fileData, err := os.ReadFile(dataFile)
	if err != nil {
		config.Logger.Error("Failed to read district bearers file", zap.Error(err))
		return nil, fmt.Errorf("failed to read data file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(fileData, &data); err != nil {
		config.Logger.Error("Failed to parse district bearers JSON", zap.Error(err))
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return data, nil
}

// readDataFile wrapper for external calls
func readDataFile() (DistrictBearersData, error) {
	return readDataFileLocked()
}

// writeDataFileLocked writes data to JSON file (must be called with lock held)
func writeDataFileLocked(data DistrictBearersData) error {
	// Ensure directory exists
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write file
	if err := os.WriteFile(dataFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// createEmptyBearers returns 6 empty bearer objects
func createEmptyBearers() []DistrictBearer {
	positions := []string{
		"President",
		"Secretary",
		"Treasurer",
		"Joint Secretary (Women)",
		"Joint Secretary (Seed)",
		"Joint Secretary (Marketing)",
	}

	bearers := make([]DistrictBearer, 6)
	for i, pos := range positions {
		bearers[i] = DistrictBearer{
			Name:    "",
			Title:   pos,
			Contact: "",
		}
	}
	return bearers
}

// normalizeBearers ensures exactly 6 bearers in the correct order
func normalizeBearers(bearers []DistrictBearer) []DistrictBearer {
	positions := []string{
		"President",
		"Secretary",
		"Treasurer",
		"Joint Secretary (Women)",
		"Joint Secretary (Seed)",
		"Joint Secretary (Marketing)",
	}

	result := make([]DistrictBearer, 6)

	// Create a map for quick lookup by title
	bearerMap := make(map[string]DistrictBearer)
	for _, b := range bearers {
		bearerMap[b.Title] = b
	}

	// Fill in order of positions
	for i, pos := range positions {
		if bearer, exists := bearerMap[pos]; exists {
			result[i] = bearer
		} else {
			result[i] = DistrictBearer{
				Name:    "",
				Title:   pos,
				Contact: "",
			}
		}
	}

	return result
}
