package office

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"taga-api/config"
	"taga-api/model"

	"go.uber.org/zap"
)

// GetOfficeData retrieves office data based on the type requested
// pathParam can be "state" for state officers or a district name for district officers
func GetOfficeData(pathParam string) (interface{}, error) {
	// Validate path parameter
	if pathParam == "" {
		config.Logger.Error("Empty path parameter provided")
		return nil, fmt.Errorf("office type is required")
	}

	// Construct the file path to the main office data file
	filename := filepath.Join("data", "office", "taga-office.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		config.Logger.Error("Office data file not found",
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("office data file not found")
	}

	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		config.Logger.Error("Failed to read office data file",
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read office data")
	}

	// Check if file is empty
	if len(data) == 0 {
		config.Logger.Error("Office data file is empty",
			zap.String("filename", filename))
		return nil, fmt.Errorf("office data file is empty")
	}

	// Parse the JSON data
	var officeData model.OfficeData
	err = json.Unmarshal(data, &officeData)
	if err != nil {
		config.Logger.Error("Failed to parse office data",
			zap.String("filename", filename),
			zap.Error(err))
		return nil, fmt.Errorf("failed to parse office data")
	}

	// Determine what data to return based on pathParam
	normalizedParam := strings.ToLower(strings.TrimSpace(pathParam))

	if normalizedParam == "state" {
		if len(officeData.StateOfficers) == 0 {
			return nil, fmt.Errorf("no state officers found")
		}
		return officeData.StateOfficers, nil
	} else {
		// Check if it's a valid district
		districtOfficers, exists := officeData.DistrictOfficers[pathParam]
		if !exists {
			// Try to find a case-insensitive match
			for districtName := range officeData.DistrictOfficers {
				if strings.EqualFold(districtName, pathParam) {
					districtOfficers = officeData.DistrictOfficers[districtName]
					exists = true
					break
				}
			}
		}

		if !exists {
			return nil, fmt.Errorf("district '%s' not found. Available districts: %v",
				pathParam, getAvailableDistricts(officeData.DistrictOfficers))
		}

		if len(districtOfficers) == 0 {
			return nil, fmt.Errorf("no officers found for district '%s'", pathParam)
		}
		return districtOfficers, nil
	}
}

// Helper function to get list of available districts
func getAvailableDistricts(districts map[string][]model.DistrictOfficer) []string {
	available := make([]string, 0, len(districts))
	for district := range districts {
		available = append(available, district)
	}
	return available
}
