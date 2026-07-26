# E2E Framework Configuration & Customization Guide

This guide walks you through cloning, configuring, and extending the E2E testing framework for a new website and API target. We will use the target URLs `https://nammataga.com` (Frontend UI) and `https://api.nammataga.com` (Backend API) as a real-world example.

---

## 1. Cloning and Initial Setup

Clone this repository and verify dependencies:

```bash
# Clone the repository (replace with actual git URL)
git clone <repository-url> e2e-tests
cd e2e-tests

# Install dependencies (Selenium bindings and configuration packages)
make deps
```

---

## 2. Configuring Targets

To configure the framework to run against `nammataga.com`, update `config.json` in the root of the repository:

```json
{
  "baseUrl": "https://api.nammataga.com",
  "uiUrl": "https://nammataga.com",
  "seleniumUrl": "http://localhost:9515",
  "headless": false,
  "timeout": 10
}
```

*Note: You can override these variables on the fly in CI environments using environment variables (e.g. `E2E_BASE_URL=https://api.nammataga.com E2E_UI_URL=https://nammataga.com E2E_HEADLESS=true make test-all`).*

---

## 3. UI Element Capturing Techniques

We use the **Page Object Model (POM)** pattern. To automate pages, you must identify stable element locators.

### Finding Elements via Chrome DevTools
1. Open Chrome, navigate to `https://nammataga.com`, right-click on any element (e.g. a Sign In button), and select **Inspect**.
2. In the **Elements** panel of DevTools:
   - Press `Ctrl + F` (or `Cmd + F` on macOS) to open the search bar.
   - Test your selector candidates (CSS or XPath) to ensure they match exactly **1 element**.

### Selector Strategy Guidelines
* **Prefer IDs**: Selectors like `css:#login-email` or `css:#submit-btn` are fast and highly stable.
* **Use Clean CSS Classes**: Look for semantic class names like `css:.navbar-brand` or `css:.btn-primary`.
* **Avoid Auto-Generated Classes**: Avoid dynamic hashes (e.g. `css:.StyledButton-sc-1234a-0`). These change with every frontend build.
* **Use Attributes**: Elements without IDs often have descriptive attributes.
  - CSS: `css:input[name='email']` or `css:button[type='submit']`
* **XPath for Text/Hierarchies**: Use XPath if you need to locate elements by visible text or complex sibling hierarchies:
  - Text Match: `xpath://button[contains(text(), 'Sign In')]`
  - Sibling/Parent navigation: `xpath://div[@class='card-body']/following-sibling::div/button`

---

## 4. Coding Page Objects (UI Testing)

Create a new file in `pkg/ui/pages/` representing the web page. For example, let's create a login and user profile page for `nammataga.com`.

### Create Page Object: `pkg/ui/pages/profile_page.go`
```go
package pages

import (
	"time"

	"github.com/tebeka/selenium"
	"e2e-template/pkg/ui"
)

type ProfilePage struct {
	*ui.Page
	AvatarIcon    string
	ProfileName   string
	LogoutButton  string
}

func NewProfilePage(driver selenium.WebDriver, screenshotDir string) *ProfilePage {
	return &ProfilePage{
		Page:         ui.NewPage(driver, screenshotDir),
		AvatarIcon:   "css:.user-avatar",
		ProfileName:  "css:#profile-header-name",
		LogoutButton: "xpath://button[text()='Logout']",
	}
}

func (p *ProfilePage) GetUsername(timeout time.Duration) (string, error) {
	return p.GetText(p.ProfileName, timeout)
}
```

### Write UI E2E Test: `tests/profile_ui_test.go`
```go
package tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
)

func TestUI_NammatagaLogin(t *testing.T) {
	RunUITest(t, "Nammataga Authentication Flow", func(t *testing.T, page *ui.Page) {
		// Navigate to website
		if err := page.Driver.Get("https://nammataga.com/login"); err != nil {
			t.Fatalf("Failed to load page: %v", err)
		}

		loginPage := pages.NewLoginPage(page.Driver, page.ScreenshotDir)
		profilePage := pages.NewProfilePage(page.Driver, page.ScreenshotDir)

		// Perform login
		err := loginPage.Login("user@nammataga.com", "secure-password", 5*time.Second)
		if err != nil {
			t.Fatalf("Login attempt failed: %v", err)
		}

		// Verify redirected page username display matches
		name, err := profilePage.GetUsername(5*time.Second)
		if err != nil {
			t.Fatalf("Failed to fetch profile username: %v", err)
		}

		if name != "John Doe" {
			t.Errorf("Expected profile username 'John Doe', got '%s'", name)
		}
	})
}
```

---

## 5. Structuring API Models (API Testing)

API models map backend JSON keys to Go structs. These should be placed under `pkg/models/` or inside your API test suite.

### Where to put models?
* **Shared Models**: Place in a new folder `pkg/models/` (e.g. `pkg/models/user.go`) if they are reused across multiple test suites.
* **Test-Specific Models**: Place them directly inside the test file (e.g. `tests/api_auth_test.go`) if they are only used in a single validation suite.

### Example: API Authenticate Structs

Let's model the Login endpoint of `https://api.nammataga.com/v1/auth/login`.

#### Struct Design:
```go
package models

// LoginRequest defines the payload structure sent to POST /v1/auth/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserProfile models user metadata returned inside the login response.
type UserProfile struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// LoginResponse defines the payload returned from POST /v1/auth/login
type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int         `json:"expires_in"`
	User        UserProfile `json:"user"`
}
```

#### Writing API E2E Test: `tests/auth_api_test.go`
```go
package tests

import (
	"testing"
	"time"

	"e2e-template/pkg/client"
)

func TestAPI_AuthenticateUser(t *testing.T) {
	RunAPITest(t, "Nammataga Login Endpoint Validation", func(t *testing.T, c *client.Client) {
		reqPayload := &LoginRequest{
			Email:    "test@nammataga.com",
			Password: "my-super-secret-password",
		}
		var respPayload LoginResponse

		// Send request: Note we pass structs as pointers to c.SendHttpRequest
		err := c.SendHttpRequest("POST", "/v1/auth/login", nil, reqPayload, &respPayload, nil)
		if err != nil {
			t.Fatalf("API authentication failed: %v", err)
		}

		// Assertions
		if respPayload.AccessToken == "" {
			t.Errorf("Authentication failed: AccessToken was not returned")
		}
		if respPayload.User.Email != reqPayload.Email {
			t.Errorf("Expected authenticated user email to be %s, got %s", reqPayload.Email, respPayload.User.Email)
		}
	})
}
```

---

## 6. Execution Flow and Verification

1. Start your local Chromedriver:
   ```bash
   chromedriver --port=9515
   ```
2. Run your specific Nammataga tests:
   ```bash
   # Run all test suites
   make test-all

   # Or run specific test case matching pattern
   go test -v ./tests/... -run=TestAPI_AuthenticateUser
   ```
3. Open the newly generated HTML dashboard to verify logs and API exchanges:
   ```bash
   # View report (locate path in terminal logs)
   open evidence/run-<timestamp>/reports/report.html
   ```
