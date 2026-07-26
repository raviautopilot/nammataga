package ui_tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
	"e2e-template/tests"
)

func TestUI_02_LoginValidation(t *testing.T) {
	// Start a local HTTP test server containing the target UI components
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Mock Login Portal</title></head>
			<body>
				<h2>Login</h2>
				<form onsubmit="event.preventDefault(); document.getElementById('error-lbl').innerText = 'Invalid username or password';">
					<input type="text" id="username" placeholder="Username"><br/>
					<input type="password" id="password" placeholder="Password"><br/>
					<button type="submit">Submit</button>
				</form>
				<div class="error-message" id="error-lbl"></div>
			</body>
			</html>
		`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tests.RunUITest(t, "Verify Login Screen Error Message", func(t *testing.T, page *ui.Page) {
		// 1. Navigate to target mock web app
		if err := page.Driver.Get(server.URL); err != nil {
			t.Fatalf("Failed to load page URL: %v", err)
		}

		loginPage := pages.NewLoginPage(page.Driver, page.ScreenshotDir)

		// 2. Perform credentials login sequence
		err := loginPage.Login("test-user", "incorrect-pwd", 3*time.Second)
		if err != nil {
			t.Fatalf("Form submission action failed: %v", err)
		}

		// 3. Verify screen renders the error validation message
		errMsg, err := loginPage.GetAlertText(3*time.Second)
		if err != nil {
			t.Fatalf("Failed to retrieve error text from label: %v", err)
		}

		if errMsg != "Invalid username or password" {
			t.Errorf("Expected validation text 'Invalid username or password', got '%s'", errMsg)
		}
	})
}
