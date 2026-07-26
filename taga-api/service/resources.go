package service

import (
	"encoding/json"
	"fmt"
	"os"
)

type ResourceCategory struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Documents     []Document `json:"documents"`
	Subcategories []string   `json:"subcategories,omitempty"`
}

type Document struct {
	Title       string `json:"title"`
	Year        string `json:"year"`
	Subcategory string `json:"subcategory,omitempty"`
	URL         string `json:"url,omitempty"`
}

// LoadResources loads all resource categories from the JSON file
func LoadResources() ([]ResourceCategory, error) {
	file, err := os.ReadFile("data/resources.json")
	if err != nil {
		fmt.Println("ERROR in GetResourceCategories:", err)
		return nil, fmt.Errorf("cannot read resources.json: %w", err)
	}

	// Debug: print file content to verify it's correct
	fmt.Println("Loaded JSON file content:", string(file))

	var data []ResourceCategory
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	return data, nil
}
