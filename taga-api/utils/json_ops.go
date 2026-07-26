package utils

import (
	"encoding/json"
	"os"
	"taga-api/config"

	"go.uber.org/zap"
)

// readJSONFile reads and parses a JSON file into the target interface
func ReadJSONFile(filename string, target interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		config.Logger.Error("Failed to read data file", zap.String("filename", filename), zap.Error(err))
		return err
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		config.Logger.Error("Failed to parse data", zap.String("filename", filename), zap.Error(err))
		return err
	}

	return nil
}
