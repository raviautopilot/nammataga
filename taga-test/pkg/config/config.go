package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// Credentials holds a username/password pair.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config holds the configuration values for the testing framework.
type Config struct {
	BaseURL                        string      `json:"baseUrl"`
	UiURL                          string      `json:"uiUrl"`
	SeleniumURL                    string      `json:"seleniumUrl"`
	Headless                       bool        `json:"headless"`
	Timeout                        int         `json:"timeout"`
	AdminLoginButtonTestID         string      `json:"adminLoginButtonTestID"`
	MemberLoginButtonTestID        string      `json:"memberLoginButtonTestID"`
	AdminLoginTestIDs              []string    `json:"adminLoginTestIDs"`
	MemberLoginTestIDs             []string    `json:"memberLoginTestIDs"`
	AdminCredentials               Credentials `json:"adminCredentials"`
	MemberCredentials              Credentials `json:"memberCredentials"`
	AdminLoginUsernameInputTestID  string      `json:"adminLoginUsernameInputTestID"`
	AdminLoginPasswordInputTestID  string      `json:"adminLoginPasswordInputTestID"`
	AdminLoginSubmitButtonTestID   string      `json:"adminLoginSubmitButtonTestID"`
	MemberLoginUsernameInputTestID string      `json:"memberLoginUsernameInputTestID"`
	MemberLoginPasswordInputTestID string      `json:"memberLoginPasswordInputTestID"`
	MemberLoginSubmitButtonTestID  string      `json:"memberLoginSubmitButtonTestID"`
	LogoutButtonTestID             string      `json:"logoutButtonTestID"`
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
