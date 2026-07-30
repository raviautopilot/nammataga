package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

// TestUI_02_LoginValidation executes a single combined E2E test verifying both Admin and Member login pages/modals.
func TestUI_02_LoginValidation(t *testing.T) {
	tests.RunUITest(t, "Verify Admin and Member Login Element Availability", func(t *testing.T, page *ui.Page) {
		targetURL := tests.GlobalConfig.UiURL
		if targetURL == "" {
			t.Fatal("GlobalConfig.UiURL is empty")
		}

		// ==========================================
		// 1. ADMIN LOGIN VALIDATION
		// ==========================================
		// Navigate to root home page using Page receiver method
		if err := page.GoToHomePage(targetURL); err != nil {
			t.Fatalf("Admin flow - GoToHomePage failed: %v", err)
		}

		// Verify admin login button on landing page
		if err := page.VerifyElementsPresentByTestIDs([]string{"testid-admin-login-button"}, 5*time.Second); err != nil {
			t.Fatalf("Admin login button validation failed: %v", err)
		}

		// Open Admin login modal/page
		if err := page.ClickByTestID("testid-admin-login-button", 5*time.Second); err != nil {
			t.Fatalf("Failed to click admin login button: %v", err)
		}

		// Verify admin form elements from config.json (adminLoginTestIDs)
		adminIDs := tests.GlobalConfig.AdminLoginTestIDs
		if len(adminIDs) == 0 {
			t.Fatal("GlobalConfig.AdminLoginTestIDs is empty")
		}

		if err := page.VerifyElementsPresentByTestIDs(adminIDs, 5*time.Second); err != nil {
			t.Fatalf("Admin login form fields validation failed: %v", err)
		}

		t.Log("✅ Admin login elements validated successfully")

		// ==========================================
		// 2. MEMBER LOGIN VALIDATION
		// ==========================================
		// Reset state by navigating back to root home page
		if err := page.GoToHomePage(targetURL); err != nil {
			t.Fatalf("Member flow - GoToHomePage failed: %v", err)
		}

		// Verify member login button on landing page
		if err := page.VerifyElementsPresentByTestIDs([]string{"testid-member-login-button"}, 5*time.Second); err != nil {
			t.Fatalf("Member login button validation failed: %v", err)
		}

		// Open Member login modal/page
		if err := page.ClickByTestID("testid-member-login-button", 5*time.Second); err != nil {
			t.Fatalf("Failed to click member member button: %v", err)
		}

		// Verify member form elements from config.json (memberLoginTestIDs)
		memberIDs := tests.GlobalConfig.MemberLoginTestIDs
		if len(memberIDs) == 0 {
			t.Fatal("GlobalConfig.MemberLoginTestIDs is empty")
		}

		if err := page.VerifyElementsPresentByTestIDs(memberIDs, 5*time.Second); err != nil {
			t.Fatalf("Member login form fields validation failed: %v", err)
		}

		t.Log("✅ Member login elements validated successfully")
	})
}
