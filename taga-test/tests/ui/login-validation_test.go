package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

func TestUI_02_LoginValidation(t *testing.T) {
	tests.RunUITest(t, "Verify Admin and Member Login Element Availability", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		url := cfg.UiURL
		timeout := 5 * time.Second

		// 1. Go to Home Page
		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}

		// 2. Open Admin Login & capture screenshot of Admin Login Modal
		if err := page.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := page.VerifyFormElements(cfg.AdminLoginTestIDs, timeout); err != nil {
			t.Fatalf("Failed to verify Admin Form Elements: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_01_Admin_Login_Modal")

		// 3. Return to Home Page
		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// 4. Open Member Login & capture screenshot of Member Login Modal
		if err := page.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := page.VerifyFormElements(cfg.MemberLoginTestIDs, timeout); err != nil {
			t.Fatalf("Failed to verify Member Form Elements: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_02_Member_Login_Modal")

		// 5. Return to Home Page
		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home at end: %v", err)
		}
	})
}
