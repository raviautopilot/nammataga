package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

func TestUI_02_LoginValidation(t *testing.T) {
	tests.RunUITest(t, "Verify Nammataga Login Validation", func(t *testing.T, page *ui.Page) {
		targetURL := tests.GlobalConfig.UiURL
		if targetURL == "" {
			t.Fatal("GlobalConfig.UiURL is empty")
		}

		// 1. Navigate to target web app
		if err := page.Driver.Get(targetURL); err != nil {
			t.Fatalf("Failed to load page URL: %v", err)
		}

		// Verify member login button is present on landing page first
		if err := page.VerifyElementsPresentByTestIDs([]string{"testid-member-login-button"}, 5*time.Second); err != nil {
			t.Fatalf("Landing page validation failed: %v", err)
		}

		// 2. Click: testid-member-login-button
		err := page.ClickByTestID("testid-member-login-button", 5*time.Second)
		if err != nil {
			t.Fatalf("Failed to click member login button: %v", err)
		}

		// Verify all login form elements are loaded on the login page/modal before proceeding
		loginPageElements := []string{
			"testid-login-identifier-input",
			"testid-login-password-input",
			"testid-login-submit-button",
		}
		if err := page.VerifyElementsPresentByTestIDs(loginPageElements, 5*time.Second); err != nil {
			t.Fatalf("Login page validation failed: %v", err)
		}

		// 3. Send value: testid-login-identifier-input (testuser@gmail.com)
		err = page.SendKeysByTestID("testid-login-identifier-input", "testuser@gmail.com", 5*time.Second)
		if err != nil {
			t.Fatalf("Failed to send login identifier: %v", err)
		}

		// 4. Send value: testid-login-password-input (test123)
		err = page.SendKeysByTestID("testid-login-password-input", "test123", 5*time.Second)
		if err != nil {
			t.Fatalf("Failed to send login password: %v", err)
		}

		// 5. Click: testid-login-submit-button
		err = page.ClickByTestID("testid-login-submit-button", 5*time.Second)
		if err != nil {
			t.Fatalf("Failed to click login submit button: %v", err)
		}
	})
}
