package tests

import (
	"encoding/json"
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

	"golang.org/x/crypto/bcrypt"

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
		if err := os.MkdirAll(ExecutionLogDir, 0777); err != nil {
			fmt.Printf("WARNING: Failed to create log directory: %v\n", err)
		}
		if err := os.MkdirAll(ExecutionReportDir, 0777); err != nil {
			fmt.Printf("WARNING: Failed to create report directory: %v\n", err)
		}
		if err := os.MkdirAll(ExecutionScreenshotDir, 0777); err != nil {
			fmt.Printf("WARNING: Failed to create screenshot directory: %v\n", err)
		}

		seedMemberUser()
		mergeOfficeData()
		logger.SetLevel(logger.INFO)
	})
}

func mergeOfficeData() {
	stateFile := "/home/sudhan_dev/Downloads/code/nammataga/taga-api/data/office_bearers/state-executive.json"
	districtFile := "/home/sudhan_dev/Downloads/code/nammataga/taga-api/data/office_bearers/district_office_bearers.json"
	targetDir := "/home/sudhan_dev/Downloads/code/nammataga/taga-api/data/office"
	targetFile := filepath.Join(targetDir, "taga-office.json")

	stateData, err := os.ReadFile(stateFile)
	if err != nil {
		fmt.Printf("WARNING: Failed to read state officers: %v\n", err)
		return
	}
	var stateOfficers []interface{}
	if err := json.Unmarshal(stateData, &stateOfficers); err != nil {
		fmt.Printf("WARNING: Failed to parse state officers: %v\n", err)
		return
	}

	districtData, err := os.ReadFile(districtFile)
	if err != nil {
		fmt.Printf("WARNING: Failed to read district officers: %v\n", err)
		return
	}
	var districtOfficers map[string]interface{}
	if err := json.Unmarshal(districtData, &districtOfficers); err != nil {
		fmt.Printf("WARNING: Failed to parse district officers: %v\n", err)
		return
	}

	combined := map[string]interface{}{
		"state_officers":    stateOfficers,
		"district_officers": districtOfficers,
	}

	combinedJSON, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		fmt.Printf("WARNING: Failed to marshal merged office data: %v\n", err)
		return
	}

	if err := os.MkdirAll(targetDir, 0777); err != nil {
		fmt.Printf("WARNING: Failed to create target office dir: %v\n", err)
		return
	}

	if err := os.WriteFile(targetFile, combinedJSON, 0777); err != nil {
		fmt.Printf("WARNING: Failed to write merged office data file: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully merged state & district office bearers into taga-office.json")
}

func seedMemberUser() {
	membersFile := "/home/sudhan_dev/Downloads/code/nammataga/taga-api/data/member/members.json"

	data, err := os.ReadFile(membersFile)
	if err != nil {
		fmt.Printf("WARNING: Failed to read members file %s: %v\n", membersFile, err)
		return
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		fmt.Printf("WARNING: Failed to parse members file: %v\n", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("WARNING: Failed to hash password: %v\n", err)
		return
	}

	email := "sudhantest08@gmail.com"
	found := false
	for i, m := range members {
		if mEmail, ok := m["emailId"].(string); ok && strings.EqualFold(mEmail, email) {
			members[i]["password"] = string(hashedPassword)
			members[i]["first_login"] = false
			found = true
			break
		}
	}

	if !found {
		newM := map[string]interface{}{
			"id":                       "d11348e1-9a65-4945-bb1b-f100a5df15cg",
			"emailId":                  email,
			"username":                 email,
			"password":                 string(hashedPassword),
			"first_login":              false,
			"name":                     "Test Member Sudhan",
			"initial":                  "S",
			"gender":                   "Male",
			"mobile_number":            "9944637254",
			"tagaId":                   "TAGA005",
			"payment_status":           "Paid",
			"subscription_active":      true,
			"cps_gpf_number":           "GPF001",
			"date_of_birth":            "1988-05-15",
			"designation":              "Agriculture Officer",
			"educational_qualification": "M.Sc Agriculture",
			"father_name":              "Rajendran",
			"mother_name":              "Sita Lakshmi",
			"native_district":          "Salem",
			"working_district":         "Salem",
			"residential_address":      "12 Gandhi Nagar",
			"permanent_address":        "12 Gandhi Nagar",
			"tbf_number":               "TBF001",
			"sno":                      1,
			"login_count":              6,
			"created_at":               time.Now().Format(time.RFC3339),
		}
		members = append(members, newM)
	}

	updatedData, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		fmt.Printf("WARNING: Failed to marshal seeded members: %v\n", err)
		return
	}

	if err := os.WriteFile(membersFile, updatedData, 0644); err != nil {
		fmt.Printf("WARNING: Failed to write seeded members file: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully seeded member user 'sudhantest08@gmail.com' in database")
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

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}
	
	c := exec.Command(cmd, args...)
	if err := c.Start(); err == nil {
		return nil
	}
	
	return exec.Command("python3", "-m", "webbrowser", url).Start()
}

// TeardownSuite generates test execution reports.
func TeardownSuite() {
	stopChromeDriver()

	rep := report.GetGlobalReporter()
	if err := rep.GenerateReports(ExecutionReportDir); err != nil {
		logger.Error("Failed to compile test reports: %v", err)
	} else {
		logger.Info("Test reports generated in '%s' directory.", ExecutionReportDir)

		// Automatically open report in browser if not running in headless mode
		// and only trigger it during the final 'ui.test' package run.
		if GlobalConfig != nil && !GlobalConfig.Headless && os.Getenv("E2E_HEADLESS") != "true" && strings.Contains(os.Args[0], "ui.test") {
			htmlPath := filepath.Join(ExecutionReportDir, "report.html")
			if absPath, err := filepath.Abs(htmlPath); err == nil {
				logger.Info("Opening test report: %s", absPath)
				if err := openBrowser(absPath); err != nil {
					logger.Error("Failed to open report in browser: %v", err)
				}
			}
		}
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

		// Create a separate subdirectory for this specific test's screenshots
		sanitizedTestDir := strings.ReplaceAll(name, "/", "_")
		sanitizedTestDir = strings.ReplaceAll(sanitizedTestDir, " ", "_")
		testScreenshotDir := filepath.Join(ExecutionScreenshotDir, sanitizedTestDir)

		page := ui.NewPage(driver, testScreenshotDir)

		defer func() {
			duration := time.Since(startTime)
			status := "passed"
			errStr := ""
			screenshotPath := ""

			// Retrieve and save all intercepted network requests per test
			if netReqs := page.RetrieveNetworkRequests(); len(netReqs) > 0 {
				testRequestDir := filepath.Join(ExecutionLogDir, sanitizedTestDir)
				saveTestNetworkLogs(testRequestDir, netReqs)
			}

			if subT.Failed() {
				status = "failed"
				errStr = "UI interaction or page assertion failure."
				if page.LastError != nil {
					errStr = page.LastError.Error()
				}
				if netErrors := page.RetrieveNetworkErrors(); netErrors != "" {
					formatted := ui.FormatNetworkErrors(netErrors)
					if formatted != "" {
						errStr = fmt.Sprintf("%s\n%s", errStr, formatted)
						logger.Error("Test failed with network errors:\n%s", formatted)
						subT.Logf("Intercepted Network Errors:\n%s", formatted)
					}
				}
				if path, sErr := page.CaptureScreenshot(name); sErr == nil {
					screenshotPath = "../screenshots/" + sanitizedTestDir + "/" + filepath.Base(path)
				} else {
					logger.Error("Failed to write failure screenshot: %v", sErr)
				}
			}

			// Collect all screenshots from this test's specific directory (sorted by step order)
			var screenshots []string
			if entries, err := os.ReadDir(testScreenshotDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() {
						screenshots = append(screenshots, "../screenshots/"+sanitizedTestDir+"/"+entry.Name())
					}
				}
			}

			rep.RecordWithCategoryAndScreenshots(name, "UI", category, status, duration, errStr, screenshotPath, screenshots)
		}()


		fn(subT, page)
	})
}

// CleanupMemberByEmailOrMobile ensures any existing member matching email or mobile is deleted via API.
func CleanupMemberByEmailOrMobile(cfg *config.Config, identifiers ...string) {
	if cfg == nil || cfg.BaseURL == "" {
		logger.Info("[API Cleanup] Skipping: cfg or BaseURL is empty")
		return
	}

	c := client.NewClient(cfg.BaseURL, time.Duration(cfg.Timeout)*time.Second, ExecutionLogDir)

	loginReq := map[string]string{
		"username": cfg.AdminCredentials.Username,
		"password": cfg.AdminCredentials.Password,
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := c.SendHttpRequest("POST", "/api/admin/login", nil, &loginReq, &loginResp, nil); err != nil || loginResp.Token == "" {
		logger.Error("[API Cleanup] Failed to login admin via API: %v", err)
		return
	}

	auth := &client.BearerTokenAuth{Token: loginResp.Token}
	deletedIDs := make(map[string]bool)

	for _, idStr := range identifiers {
		clean := strings.TrimSpace(idStr)
		if clean == "" {
			continue
		}

		var searchResp struct {
			Members []struct {
				ID string `json:"id"`
			} `json:"members"`
		}

		err := c.SendHttpRequest("GET", "/api/admin/members?search="+url.QueryEscape(clean), nil, nil, &searchResp, auth)
		if err != nil {
			logger.Error("[API Cleanup] Search failed for '%s': %v", clean, err)
			continue
		}

		logger.Info("[API Cleanup] Search for '%s' returned %d member(s)", clean, len(searchResp.Members))
		for _, m := range searchResp.Members {
			if m.ID != "" && !deletedIDs[m.ID] {
				var delResp map[string]interface{}
				if err := c.SendHttpRequest("DELETE", "/api/admin/members/"+m.ID, nil, nil, &delResp, auth); err != nil {
					logger.Error("[API Cleanup] Failed to delete member ID %s: %v", m.ID, err)
				} else {
					logger.Info("[API Cleanup] Successfully deleted member ID %s via API", m.ID)
					deletedIDs[m.ID] = true
				}
			}
		}
	}
}

func saveTestNetworkLogs(testRequestDir string, requests []ui.InterceptedNetworkRequest) {
	if len(requests) == 0 {
		return
	}
	if err := os.MkdirAll(testRequestDir, 0777); err != nil {
		logger.Error("Failed to create test request directory %s: %v", testRequestDir, err)
		return
	}

	// Save consolidated JSON trace file
	tracePath := filepath.Join(testRequestDir, "requests.json")
	if data, err := json.MarshalIndent(requests, "", "  "); err == nil {
		_ = os.WriteFile(tracePath, data, 0644)
	}

	// Save individual request/response JSON files
	for i, req := range requests {
		seq := i + 1

		var reqBodyJSON interface{}
		if req.RequestBody != "" {
			_ = json.Unmarshal([]byte(req.RequestBody), &reqBodyJSON)
			if reqBodyJSON == nil {
				reqBodyJSON = req.RequestBody
			}
		}
		reqLog := map[string]interface{}{
			"timestamp": req.Timestamp,
			"method":    req.Method,
			"url":       req.URL,
			"body":      reqBodyJSON,
		}
		reqFilePath := filepath.Join(testRequestDir, fmt.Sprintf("%03d-request.json", seq))
		if data, err := json.MarshalIndent(reqLog, "", "  "); err == nil {
			_ = os.WriteFile(reqFilePath, data, 0644)
		}

		var respBodyJSON interface{}
		if req.ResponseBody != "" {
			_ = json.Unmarshal([]byte(req.ResponseBody), &respBodyJSON)
			if respBodyJSON == nil {
				respBodyJSON = req.ResponseBody
			}
		}
		respLog := map[string]interface{}{
			"timestamp":   req.Timestamp,
			"status_code": req.Status,
			"latency_ms":  req.LatencyMs,
			"body":        respBodyJSON,
		}
		respFilePath := filepath.Join(testRequestDir, fmt.Sprintf("%03d-response.json", seq))
		if data, err := json.MarshalIndent(respLog, "", "  "); err == nil {
			_ = os.WriteFile(respFilePath, data, 0644)
		}
	}
}



