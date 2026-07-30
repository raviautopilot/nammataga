package tests

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

var (
	chromeDriverCmd *exec.Cmd
	chromeDriverMu  sync.Mutex
)

func startChromeDriverIfNeeded() {
	chromeDriverMu.Lock()
	defer chromeDriverMu.Unlock()

	if chromeDriverCmd != nil {
		return
	}

	seleniumURL := "http://localhost:9515"
	if GlobalConfig != nil && GlobalConfig.SeleniumURL != "" {
		seleniumURL = GlobalConfig.SeleniumURL
	}

	u, err := url.Parse(seleniumURL)
	var addr string
	var port string = "9515"
	if err == nil {
		addr = u.Host
		if _, p, err := net.SplitHostPort(u.Host); err == nil {
			port = p
		} else if !strings.Contains(u.Host, ":") {
			addr = u.Host + ":80"
		}
	} else {
		addr = "127.0.0.1:9515"
	}

	// Check if port is already listening
	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err == nil {
		conn.Close()
		// Chromedriver or another service is already running on this port, nothing to do
		return
	}

	// Start chromedriver in the background
	logger.Info("Auto-starting chromedriver on port %s...", port)
	cmd := exec.Command("chromedriver", "--port="+port)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		logger.Warn("Failed to auto-start chromedriver: %v. Tests might fail if webdriver is not running.", err)
		return
	}
	chromeDriverCmd = cmd

	// Wait up to 3 seconds for it to start
	for i := 0; i < 15; i++ {
		c, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			c.Close()
			logger.Info("Chromedriver started successfully on port %s.", port)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func stopChromeDriver() {
	chromeDriverMu.Lock()
	defer chromeDriverMu.Unlock()

	if chromeDriverCmd != nil && chromeDriverCmd.Process != nil {
		logger.Info("Stopping auto-started chromedriver...")
		_ = chromeDriverCmd.Process.Kill()
		_ = chromeDriverCmd.Wait()
		chromeDriverCmd = nil
	}
}

// TeardownSuite generates test execution reports.
func TeardownSuite() {
	stopChromeDriver()

	rep := report.GetGlobalReporter()
	if err := rep.GenerateReports(ExecutionReportDir); err != nil {
		logger.Error("Failed to compile test reports: %v", err)
	} else {
		logger.Info("Test reports generated in '%s' directory.", ExecutionReportDir)
	}
}

// TestContext wraps testing.T with rich metadata for expected vs actual reporting.
type TestContext struct {
	*testing.T
	Client        *client.Client
	Description   string
	Expected      string
	Actual        string
	FailureReason string
}

// RunAPITestWithDetails executes an API test with rich description, expected, actual, and failure tracking.
func RunAPITestWithDetails(t *testing.T, name string, description string, expected string, fn func(tc *TestContext)) {
	rep := report.GetGlobalReporter()
	startTime := time.Now()

	category := "General"
	if _, file, _, ok := runtime.Caller(1); ok {
		category = filepath.Base(file)
	}

	t.Run(name, func(subT *testing.T) {
		c := client.NewClient(GlobalConfig.BaseURL, time.Duration(GlobalConfig.Timeout)*time.Second, ExecutionLogDir)
		tc := &TestContext{
			T:           subT,
			Client:      c,
			Description: description,
			Expected:    expected,
		}

		defer func() {
			duration := time.Since(startTime)
			status := "passed"
			errStr := ""

			if subT.Failed() {
				status = "failed"
				errStr = "API assertion or validation error."
				if c.LastError != nil {
					errStr = c.LastError.Error()
				}
				if tc.FailureReason == "" {
					tc.FailureReason = errStr
				}
			}

			if tc.Actual == "" {
				if status == "passed" {
					tc.Actual = "Matched expected behavior: " + tc.Expected
				} else {
					tc.Actual = "Assertion failed: " + tc.FailureReason
				}
			}

			rep.RecordDetailed(name, "API", category, tc.Description, tc.Expected, tc.Actual, status, duration, errStr, tc.FailureReason, "", nil)
		}()

		fn(tc)
	})
}

// RunAPITest is a wrapper executing an API test case, injecting a custom Client and logging results.
func RunAPITest(t *testing.T, name string, fn func(t *testing.T, c *client.Client)) {
	rep := report.GetGlobalReporter()
	startTime := time.Now()

	category := "General"
	if _, file, _, ok := runtime.Caller(1); ok {
		category = filepath.Base(file)
	}

	t.Run(name, func(subT *testing.T) {
		c := client.NewClient(GlobalConfig.BaseURL, time.Duration(GlobalConfig.Timeout)*time.Second, ExecutionLogDir)

		defer func() {
			duration := time.Since(startTime)
			status := "passed"
			errStr := ""

			if subT.Failed() {
				status = "failed"
				errStr = "API assertion or validation error."
				if c.LastError != nil {
					errStr = c.LastError.Error()
				}
			}

			actual := "200 OK Response"
			if status == "failed" {
				actual = errStr
			}

			rep.RecordDetailed(name, "API", category, name, "Valid 200 OK API Response", actual, status, duration, errStr, "", "", nil)
		}()

		fn(subT, c)
	})
}


// RunUITest is a wrapper executing a UI test case, injecting a Page Object model, managing ChromeDriver, and capturing screenshots on failure.
func RunUITest(t *testing.T, name string, fn func(t *testing.T, page *ui.Page)) {
	startChromeDriverIfNeeded()

	rep := report.GetGlobalReporter()
	startTime := time.Now()

	category := "General"
	if _, file, _, ok := runtime.Caller(1); ok {
		category = filepath.Base(file)
	}

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
				if page.LastError != nil {
					errStr = page.LastError.Error()
				}
				if path, sErr := page.CaptureScreenshot(name); sErr == nil {
					screenshotPath = "../screenshots/" + filepath.Base(path)
				} else {
					logger.Error("Failed to write failure screenshot: %v", sErr)
				}
			}

			// Collect all screenshots from this test run (sorted by filename = step order)
			var screenshots []string
			if entries, err := os.ReadDir(ExecutionScreenshotDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						screenshots = append(screenshots, "../screenshots/"+entry.Name())
					}
				}
			}

			rep.RecordWithCategoryAndScreenshots(name, "UI", category, status, duration, errStr, screenshotPath, screenshots)
		}()

		fn(subT, page)
	})
}
