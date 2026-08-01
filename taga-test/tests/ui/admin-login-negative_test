package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
	"e2e-template/tests"
)

const afterActionWaitAdmin = 2 * time.Second
const testNameAdminNegative = "Admin Negative Login Scenarios"

func TestUI_04_AdminLoginNegative(t *testing.T) {
	tests.RunUITest(t, testNameAdminNegative, func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		url := cfg.UiURL
		timeout := 5 * time.Second

		homePage := pages.NewHomePage(page)
		loginPage := pages.NewLoginPage(page)

		// ── Step 1: Empty Username and Empty Password ─────────────────────
		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}
		if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to click Admin submit: %v", err)
		}
		time.Sleep(afterActionWaitAdmin)
		_, _ = page.CaptureScreenshot("Step_01_Admin_Empty_Credentials_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 2: Valid Username with Invalid Password ──────────────────
		if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := loginPage.EnterUsername(cfg.AdminLoginUsernameInputTestID, cfg.AdminCredentials.Username, timeout); err != nil {
			t.Fatalf("Failed to enter Admin username: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.AdminLoginPasswordInputTestID, "wrongpassword123", timeout); err != nil {
			t.Fatalf("Failed to enter wrong password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Admin login: %v", err)
		}
		time.Sleep(afterActionWaitAdmin)
		_, _ = page.CaptureScreenshot("Step_02_Admin_Wrong_Password_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 3: Non-Existent Admin Username ───────────────────────────
		if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := loginPage.EnterUsername(cfg.AdminLoginUsernameInputTestID, "nonexistentadmin@gmail.com", timeout); err != nil {
			t.Fatalf("Failed to enter non-existent username: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.AdminLoginPasswordInputTestID, cfg.AdminCredentials.Password, timeout); err != nil {
			t.Fatalf("Failed to enter password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Admin login: %v", err)
		}
		time.Sleep(afterActionWaitAdmin)
		_, _ = page.CaptureScreenshot("Step_03_Admin_NonExistent_User_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 4: Invalid Email Format ──────────────────────────────────
		if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := loginPage.EnterUsername(cfg.AdminLoginUsernameInputTestID, "invalid-email-format", timeout); err != nil {
			t.Fatalf("Failed to enter invalid email format: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.AdminLoginPasswordInputTestID, "password123", timeout); err != nil {
			t.Fatalf("Failed to enter password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Admin login: %v", err)
		}
		time.Sleep(afterActionWaitAdmin)
		_, _ = page.CaptureScreenshot("Step_04_Admin_Invalid_Email_Format_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 5: Password Only (Username Empty) ────────────────────────
		if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.AdminLoginPasswordInputTestID, cfg.AdminCredentials.Password, timeout); err != nil {
			t.Fatalf("Failed to enter password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Admin login: %v", err)
		}
		time.Sleep(afterActionWaitAdmin)
		_, _ = page.CaptureScreenshot("Step_05_Admin_Password_Only_Empty_Username_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}
	})
}
