# E2E UI Test Suite Execution Guide

This document outlines how to compile, execute, configure, and expand the UI automation tests. It focuses particularly on the human-readable public user journey test structure.

---

## 1. Prerequisites & Environment Setup

Before running tests, ensure that `chromedriver` is installed on your system and running on port `9515`.

### Linux/Ubuntu Installation:
```bash
sudo apt-get update
sudo apt-get install -y chromium-browser chromium-chromedriver
```

### Start Chromedriver:
Ensure it runs in the background or a separate shell:
```bash
chromedriver --port=9515
```

---

## 2. Test Configuration

The test suite reads configuration from [config.json](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/config.json). You can modify this file directly or override values via environment variables:

| config.json Field | Environment Variable | Default Value | Description |
|---|---|---|---|
| `uiUrl` | `E2E_UI_URL` | `https://nammataga.com` | Target URL of the web frontend application |
| `baseUrl` | `E2E_BASE_URL` | `https://api.nammataga.com` | Target URL of the backend API |
| `seleniumUrl` | `E2E_SELENIUM_URL` | `http://localhost:9515` | Selenium WebDriver address |
| `headless` | `E2E_HEADLESS` | `false` (headed) | Toggle headless Chrome UI mode |
| `timeout` | `E2E_TIMEOUT` | `10` | Timeout in seconds for page element waits |

---

## 3. How to Run the Tests

You can run UI tests using standard `go test` commands or by executing the pre-compiled binary directly.

### Option A: Using the `run-tests.sh` Script (Recommended)
This script automatically starts `chromedriver` in the background, executes the tests in headless mode, and cleans up chromedriver processes on exit:
```bash
# Run all tests (API and UI)
./run-tests.sh

# Run only the Public Journeys UI Test
./run-tests.sh -run TestUI_02_PublicJourneys
```

### Option B: Using the Precompiled Binary (`ui.test`)
For optimal speed in local executions and CI/CD pipelines (avoiding recompiling test files every run), you can execute the precompiled binary directly. 

> [!IMPORTANT]
> When executing the binary directly, all standard `go test` arguments must be prefixed with `test.`.

```bash
# Execute only the Public Journeys UI test using the precompiled binary
E2E_HEADLESS=true ./ui.test -test.v -test.run TestUI_02_PublicJourneys
```

---

## 4. Pointer-Based Public Journeys Test Architecture

To make UI test scripts easily readable and editable for non-technical members, we implement public journeys using a **declarative, pointer-passing style** in [02-public-journies_test.go](file:///home/ubuntu/code/github/raviautopilot/nammataga/taga-test/tests/ui/02-public-journies_test.go).

### Step-by-Step Scenario Example:
```go
func TestUI_02_PublicJourneys(t *testing.T) {
	tests.RunUITest(t, "Verify Public User Journeys", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		
		// 1. Initialize Persona & Result pointers
		pubPersona := actions.NewPublicPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("Public User Journeys Test")

		// 2. Simple sequential action flows
		actions.GoToHome(pubPersona, result)
		actions.GoToOfficeBeaers(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToEvents(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToMemberLogin(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToAdminLogin(pubPersona, result)

		// 3. Evaluate results
		if result.Failed() {
			t.Fatalf("Journey failed! Error: %v. Advice: %v", result.Error, result.Advice)
		}
	})
}
```

### How Evidence & Advice are Captured:
Each action function (e.g. `GoToOfficeBeaers`) performs the following under the hood:
1. **Actions History**: Appends the action step name to `result.Actions`.
2. **Execution Check**: Instantly skips execution if a previous step in the journey has already failed.
3. **Defensive Validation**: Ensures the page driver and elements are ready.
4. **Remediation & Advice**: If a step fails, it appends friendly debugging tips to `result.Advice` (e.g., verifying if the button's testID is present).
5. **Screenshots & Evidence**: Captures and saves screenshots to the active evidence path, attaching them to `result.Evidence` to document successes or failures.

---

## 5. Test Reports & Artifacts

Every test execution generates reports and collects evidence in the `evidence/run-YYYY-MM-DD_HH-MM-SS/` directory.

* **HTML Visual Dashboard**: Open `evidence/run-.../reports/report.html` in any web browser to see detailed pass/fail status, execution duration, and embedded screenshots.
* **JSON Results**: Parsed test data is stored in `evidence/run-.../reports/report.json`.
* **Markdown Summary**: A quick test report summary is written to `evidence/test-report.md`.
* **Screenshots Directory**: All screenshots captured by tests (successes and failures) are located under `evidence/run-.../screenshots/`.
