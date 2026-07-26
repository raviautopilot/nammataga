package utils

import (
	"encoding/json"
	"os"
)

// ReadJSON reads data from a JSON file into a struct
func ReadJSON(filePath string, data interface{}) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if len(file) == 0 {
		return nil // empty file is okay
	}

	return json.Unmarshal(file, data)
}


