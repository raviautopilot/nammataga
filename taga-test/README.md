# Go E2E Testing Framework

A ready-to-use, modular End-to-End (E2E) testing framework built in Go. It supports automated API testing via a type-safe HTTP client wrapper, UI automation using Selenium WebDriver following the Page Object Model (POM) pattern, custom HTML/JSON reporting, and granular request/response logging.

---

## Features

- **Extensible API Client**: Auto-marshaling, struct pointer safety checks, and an interface-driven Authentication manager (Bearer, Basic, API Key, mTLS, and SSH signing).
- **Selenium UI Integration**: Base Page Object wrappers handling dynamic CSS/XPath element selection, waiting hooks, interaction wrappers, and automated screenshots on test failure.
- **Observability Logging**: Automatic date-organized file logging of request and response payloads at `tests/requests/YYYY-MM-DD/HH-MM-SS-<suffix>-(request|response).json`.
- **Beautiful HTML/JSON Reporting**: Interactive responsive test results dashboard compiled automatically after each execution.

---

## Directory Structure

```
├── go.mod                     # Go module definitions
├── Makefile                   # Execution commands
├── config.json                # Environment configs (dev/staging/prod)
├── pkg/
│   ├── config/
│   │   └── config.go          # Config loader and environment overrides
│   ├── logger/
│   │   └── logger.go          # Thread-safe logging levels (INFO, DEBUG, etc.)
│   ├── client/
│   │   ├── auth.go            # Authentication interface and sub-types
│   │   └── client.go          # Custom HTTP client wrapper and JSON logger
│   ├── ui/
│   │   ├── driver.go          # Selenium WebDriver connection & option manager
│   │   └── pom.go             # Page Object Model helper wrappers
│   │   └── pages/
│   │       └── login_page.go  # Example Page Object form
│   └── report/
│       ├── report.go          # Result collector and HTML compiler
│       └── template.html      # Visual dashboard layout template
└── tests/
    ├── main_test.go           # Suite bootstrap (TestMain, RunUITest, RunAPITest)
    ├── api_test.go            # API example test cases
    └── ui_test.go             # UI example test cases
```

---

## Prerequisites

1. **Golang**: Ensure Go 1.18+ is installed.
2. **Google Chrome & Chromedriver**: Install Chrome and Chromedriver on your local machine or server.
   - For Ubuntu/Linux VM:
     ```bash
     sudo apt-get update
     sudo apt-get install -y chromium-browser chromium-chromedriver
     ```
   - Ensure `chromedriver` is available in your PATH or running on port `9515`.

---

## Quick Start

1. **Start Chromedriver**:
   ```bash
   chromedriver --port=9515
   ```

2. **Run All Tests (Headed UI & API)**:
   ```bash
   make test-all
   ```

3. **Run in Headless Mode (CI / VMs)**:
   ```bash
   E2E_HEADLESS=true make test-all
   ```

4. **Run API Tests Only**:
   ```bash
   make test-api
   ```

5. **Clean Reports and Logs**:
   ```bash
   make clean
   ```

---

## Writing Tests

### 1. API Test Cases

Create tests using `RunAPITest`, passing your structures by pointer:

```go
type Post struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
}

func TestMyAPI(t *testing.T) {
    RunAPITest(t, "Fetch Post from API", func(t *testing.T, c *client.Client) {
        var post Post
        auth := &client.BearerTokenAuth{Token: "my-token"}
        
        // Headers (optional)
        headers := map[string]string{"X-Custom-Header": "Value"}

        // SendHttpRequest: checks structure pointers and auto-marshals JSON
        err := c.SendHttpRequest("GET", "/posts/1", headers, nil, &post, auth)
        if err != nil {
            t.Fatalf("API call failed: %v", err)
        }

        if post.ID != 1 {
            t.Errorf("Expected ID 1, got %d", post.ID)
        }
    })
}
```

### 2. UI Test Cases

Create page objects inside `pkg/ui/pages/` by wrapping `*ui.Page`:

```go
type GoogleSearchPage struct {
    *ui.Page
    SearchBox string
}

func NewGoogleSearchPage(driver selenium.WebDriver) *GoogleSearchPage {
    return &GoogleSearchPage{
        Page:      ui.NewPage(driver),
        SearchBox: "css:input[name='q']",
    }
}
```

Run them using `RunUITest`:

```go
func TestGoogleSearch(t *testing.T) {
    RunUITest(t, "Search Query", func(t *testing.T, page *ui.Page) {
        if err := page.Driver.Get("https://google.com"); err != nil {
            t.Fatal(err)
        }
        
        searchPage := NewGoogleSearchPage(page.Driver)
        err := searchPage.SendKeys(searchPage.SearchBox, "Golang Selenium", 5*time.Second)
        if err != nil {
            t.Fatal(err)
        }
    })
}
```

*Note: If a UI test fails, the framework automatically writes a screenshot to `tests/screenshots/` and includes an inline link/preview in the HTML report.*

---

## Configuration Overrides

The default settings reside in `config.json`. You can override them via environment variables:

| Config JSON | Env Variable | Default | Description |
|---|---|---|---|
| `baseUrl` | `E2E_BASE_URL` | `https://jsonplaceholder.typicode.com` | Host API for test suite |
| `seleniumUrl` | `E2E_SELENIUM_URL` | `http://localhost:9515` | Endpoint address of WebDriver |
| `headless` | `E2E_HEADLESS` | `false` | Toggle headless Chrome UI mode |
| `timeout` | `E2E_TIMEOUT` | `10` | Default timeout for waits/requests (secs) |
