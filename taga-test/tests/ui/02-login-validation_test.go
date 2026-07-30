package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

// TestUI_02_MemberLoginValidation verifies that the member login button and form fields are present.
func TestUI_02_MemberLoginValidation(t *testing.T) {
	tests.RunUITest(t, "Verify Member Login Available", func(t *testing.T, page *ui.Page) {
		targetURL := tests.GlobalConfig.UiURL
		if targetURL == "" {
			t.Fatal("GlobalConfig.UiURL is empty")
		}

		// 1. Navigate to target web app
		if err := page.Driver.Get(targetURL); err != nil {
			t.Fatalf("Failed to load page URL: %v", err)
		}

		// 2. Verify member login button is present on landing page
		if err := page.VerifyElementsPresentByTestIDs([]string{"testid-member-login-button"}, 5*time.Second); err != nil {
			t.Fatalf("Member login button validation failed: %v", err)
		}

		// 3. Click member login button
		if err := page.ClickByTestID("testid-member-login-button", 5*time.Second); err != nil {
			t.Fatalf("Failed to click member login button: %v", err)
		}

		// 4. Verify member login form fields, submit button, forgot password, and change password buttons are present
		memberPageElements := []string{
			"testid-login-identifier-input",
			"testid-login-password-input",
			"testid-login-submit-button",
			"testid-forgot-password-button",
			"testid-change-password-button",
		}
		if err := page.VerifyElementsPresentByTestIDs(memberPageElements, 5*time.Second); err != nil {
			t.Fatalf("Member login form fields validation failed: %v", err)
		}
	})
}

// TestUI_02_AdminLoginValidation verifies that the admin login button and form fields are present.
func TestUI_02_AdminLoginValidation(t *testing.T) {
	tests.RunUITest(t, "Verify Admin Login Available", func(t *testing.T, page *ui.Page) {
		targetURL := tests.GlobalConfig.UiURL
		if targetURL == "" {
			t.Fatal("GlobalConfig.UiURL is empty")
		}

		// 1. Navigate to target web app
		if err := page.Driver.Get(targetURL); err != nil {
			t.Fatalf("Failed to load page URL: %v", err)
		}

		// 2. Verify admin login button is present on landing page
		if err := page.VerifyElementsPresentByTestIDs([]string{"testid-admin-login-button"}, 5*time.Second); err != nil {
			t.Fatalf("Admin login button validation failed: %v", err)
		}

		// 3. Click admin login button
		if err := page.ClickByTestID("testid-admin-login-button", 5*time.Second); err != nil {
			t.Fatalf("Failed to click admin login button: %v", err)
		}

		// 4. Verify admin login form fields and submit button are present
		adminPageElements := []string{
			"testid-admin-login-username-input",
			"testid-admin-login-password-input",
			"testid-admin-login-submit-button",
		}
		if err := page.VerifyElementsPresentByTestIDs(adminPageElements, 5*time.Second); err != nil {
			t.Fatalf("Admin login form fields validation failed: %v", err)
		}
	})
}
