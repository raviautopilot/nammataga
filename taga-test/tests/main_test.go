package tests

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

func TestMain(m *testing.M) {
	configPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	// Locate config.json dynamically based on current working directory
	var path string
	if _, err := os.Stat(*configPath); err == nil {
		path = *configPath
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

	// Setup execution timestamped folders under evidence sibling directory
	RunTimestamp = time.Now().Format("2006-01-02_15-04-05")
	EvidenceDir = "../evidence/run-" + RunTimestamp
	ExecutionLogDir = EvidenceDir + "/requests"
	ExecutionReportDir = EvidenceDir + "/reports"
	ExecutionScreenshotDir = EvidenceDir + "/screenshots"

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
	logger.Info("Starting E2E Test Suite...")
	logger.Info("Test evidence will be compiled inside: %s", EvidenceDir)

	// Run all tests in the package
	exitCode := m.Run()

	// Compile HTML and JSON reports
	rep := report.GetGlobalReporter()
	if err := rep.GenerateReports(ExecutionReportDir); err != nil {
		logger.Error("Failed to compile test reports: %v", err)
	} else {
		logger.Info("Test reports generated in '%s' directory.", ExecutionReportDir)
	}

	os.Exit(exitCode)
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
