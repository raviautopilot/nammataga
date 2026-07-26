package tests

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"e2e-template/pkg/client"
	"e2e-template/pkg/config"
	"e2e-template/pkg/logger"
	"e2e-template/pkg/report"
	"e2e-template/pkg/ui"
)

// GlobalConfig holds the environment configuration used across tests.
var GlobalConfig *config.Config

// Global evidence directories for the current test run
var (
	RunTimestamp           string
	EvidenceDir            string
	ExecutionLogDir        string
	ExecutionReportDir     string
	ExecutionScreenshotDir string
)

var setupOnce sync.Once

func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

// SetupSuite initializes the test suite configuration and directory structure.
func SetupSuite() {
	setupOnce.Do(func() {
		var configPath string
		if !flag.Parsed() {
			configPathPtr := flag.String("config", "config.json", "path to config.json")
			flag.Parse()
			configPath = *configPathPtr
		} else {
			configPath = "config.json"
		}

		// Locate config.json dynamically based on current working directory
		var path string
		if _, err := os.Stat(configPath); err == nil {
			path = configPath
		} else if _, err := os.Stat("../config.json"); err == nil {
			path = "../config.json"
		} else if _, err := os.Stat("../../config.json"); err == nil {
			path = "../../config.json"
		} else {
			path = "config.json"
		}

		cfg, err := config.LoadConfig(path)
		if err != nil {
			fmt.Printf("CRITICAL: Failed to load config: %v\n", err)
			os.Exit(1)
		}
		GlobalConfig = cfg

		RunTimestamp = os.Getenv("E2E_RUN_TIMESTAMP")
		if RunTimestamp == "" {
			RunTimestamp = time.Now().Format("2006-01-02_15-04-05")
		}

		moduleRoot, err := findModuleRoot()
		if err != nil {
			EvidenceDir = filepath.Join("..", "evidence", "run-"+RunTimestamp)
		} else {
			EvidenceDir = filepath.Join(moduleRoot, "evidence", "run-"+RunTimestamp)
		}

		ExecutionLogDir = filepath.Join(EvidenceDir, "requests")
		ExecutionReportDir = filepath.Join(EvidenceDir, "reports")
		ExecutionScreenshotDir = filepath.Join(EvidenceDir, "screenshots")

		// Create directories for the current run
		if err := os.MkdirAll(ExecutionLogDir, 0755); err != nil {
			fmt.Printf("WARNING: Failed to create log directory: %v\n", err)
		}
		if err := os.MkdirAll(ExecutionReportDir, 0755); err != nil {
			fmt.Printf("WARNING: Failed to create report directory: %v\n", err)
		}
		if err := os.MkdirAll(ExecutionScreenshotDir, 0755); err != nil {
			fmt.Printf("WARNING: Failed to create screenshot directory: %v\n", err)
		}

		logger.SetLevel(logger.INFO)
	})
}

// TeardownSuite generates test execution reports.
func TeardownSuite() {
	rep := report.GetGlobalReporter()
	if err := rep.GenerateReports(ExecutionReportDir); err != nil {
		logger.Error("Failed to compile test reports: %v", err)
	} else {
		logger.Info("Test reports generated in '%s' directory.", ExecutionReportDir)
	}
}

// RunAPITest is a wrapper executing an API test case, injecting a custom Client and logging results.
func RunAPITest(t *testing.T, name string, fn func(t *testing.T, c *client.Client)) {
	rep := report.GetGlobalReporter()
	startTime := time.Now()

	t.Run(name, func(subT *testing.T) {
		c := client.NewClient(GlobalConfig.BaseURL, time.Duration(GlobalConfig.Timeout)*time.Second, ExecutionLogDir)

		defer func() {
			duration := time.Since(startTime)
			status := "passed"
			errStr := ""

			if subT.Failed() {
				status = "failed"
				errStr = "API assertion or validation error."
			}

			rep.Record(name, "API", status, duration, errStr, "")
		}()

		fn(subT, c)
	})
}

// RunUITest is a wrapper executing a UI test case, injecting a Page Object model, managing ChromeDriver, and capturing screenshots on failure.
func RunUITest(t *testing.T, name string, fn func(t *testing.T, page *ui.Page)) {
	rep := report.GetGlobalReporter()
	startTime := time.Now()

	t.Run(name, func(subT *testing.T) {
		driver, err := ui.InitWebDriver(GlobalConfig.SeleniumURL, GlobalConfig.Headless)
		if err != nil {
			subT.Fatalf("Failed to initialize Selenium driver: %v", err)
		}
		defer driver.Quit()

		page := ui.NewPage(driver, ExecutionScreenshotDir)

		defer func() {
			duration := time.Since(startTime)
			status := "passed"
			errStr := ""
			screenshotPath := ""

			if subT.Failed() {
				status = "failed"
				errStr = "UI interaction or page assertion failure."
				if path, sErr := page.CaptureScreenshot(name); sErr == nil {
					// Save just the filename to make it relative to the report file
					screenshotPath = "screenshots/" + filepath.Base(path)
				} else {
					logger.Error("Failed to write failure screenshot: %v", sErr)
				}
			}

			rep.Record(name, "UI", status, duration, errStr, screenshotPath)
		}()

		fn(subT, page)
	})
}
