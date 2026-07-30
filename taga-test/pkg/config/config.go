package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// Config holds the configuration values for the testing framework.
type Config struct {
	BaseURL             string   `json:"baseUrl"`
	UiURL               string   `json:"uiUrl"`
	SeleniumURL         string   `json:"seleniumUrl"`
	Headless            bool     `json:"headless"`
	Timeout             int      `json:"timeout"`
	AdminLoginTestIDs  []string `json:"adminLoginTestIDs"`
	MemberLoginTestIDs []string `json:"memberLoginTestIDs"`
}

// LoadConfig reads the configuration file from path and applies environment overrides.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		BaseURL:     "https://api.nammataga.com",
		UiURL:       "https://nammataga.com",
		SeleniumURL: "http://localhost:9515",
		Headless:    false,
		Timeout:     10,
	}

	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
	}

	// Environment overrides
	if val := os.Getenv("E2E_BASE_URL"); val != "" {
		cfg.BaseURL = val
	}
	if val := os.Getenv("E2E_UI_URL"); val != "" {
		cfg.UiURL = val
	}
	if val := os.Getenv("E2E_SELENIUM_URL"); val != "" {
		cfg.SeleniumURL = val
	}
	if val := os.Getenv("E2E_HEADLESS"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			cfg.Headless = boolVal
		}
	}
	if val := os.Getenv("E2E_TIMEOUT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.Timeout = intVal
		}
	}

	return cfg, nil
}
