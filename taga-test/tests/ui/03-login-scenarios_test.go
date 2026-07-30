package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

const afterActionWait = 2 * time.Second // wait for page to settle after each action

func TestUI_03_LoginScenarios(t *testing.T) {
	tests.RunUITest(t, "Admin and Member Login Scenarios", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		url := cfg.UiURL
		timeout := 5 * time.Second

		// ══════════════════════════════════════════════════════════════════
		// ADMIN LOGIN FLOW
		// ══════════════════════════════════════════════════════════════════

		// ── Attempt 1: Admin Empty Credentials ───────────────────────────

		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}
		if err := page.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := page.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to click Admin submit: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Admin__01_Empty_Credentials_Result")

		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home after empty attempt: %v", err)
		}

		// ── Attempt 2: Admin Wrong Credentials ───────────────────────────

		if err := page.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := page.EnterUsername(cfg.AdminLoginUsernameInputTestID, cfg.AdminCredentials.Username, timeout); err != nil {
			t.Fatalf("Failed to enter Admin username: %v", err)
		}
		if err := page.EnterPassword(cfg.AdminLoginPasswordInputTestID, "wrongpassword", timeout); err != nil {
			t.Fatalf("Failed to enter wrong password: %v", err)
		}
		if err := page.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Admin login: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Admin__02_Wrong_Credentials_Result")

		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home after wrong attempt: %v", err)
		}

		// ── Attempt 3: Admin Correct Credentials ─────────────────────────

		if err := page.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := page.EnterUsername(cfg.AdminLoginUsernameInputTestID, cfg.AdminCredentials.Username, timeout); err != nil {
			t.Fatalf("Failed to enter Admin username: %v", err)
		}
		if err := page.EnterPassword(cfg.AdminLoginPasswordInputTestID, cfg.AdminCredentials.Password, timeout); err != nil {
			t.Fatalf("Failed to enter Admin password: %v", err)
		}
		if err := page.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Admin login: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Admin__03_Correct_LoggedIn_Home")

		// Logout before Member flow
		if err := page.Logout(cfg.LogoutButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to logout Admin: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Admin__04_After_Logout")

		// ══════════════════════════════════════════════════════════════════
		// MEMBER LOGIN FLOW
		// ══════════════════════════════════════════════════════════════════

		// ── Attempt 1: Member Empty Credentials ──────────────────────────

		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}
		if err := page.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := page.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to click Member submit: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Member__01_Empty_Credentials_Result")

		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home after empty attempt: %v", err)
		}

		// ── Attempt 2: Member Wrong Credentials ──────────────────────────

		if err := page.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := page.EnterUsername(cfg.MemberLoginUsernameInputTestID, cfg.MemberCredentials.Username, timeout); err != nil {
			t.Fatalf("Failed to enter Member username: %v", err)
		}
		if err := page.EnterPassword(cfg.MemberLoginPasswordInputTestID, "wrongpassword", timeout); err != nil {
			t.Fatalf("Failed to enter wrong password: %v", err)
		}
		if err := page.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Member login: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Member__02_Wrong_Credentials_Result")

		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home after wrong attempt: %v", err)
		}

		// ── Attempt 3: Member Correct Credentials ────────────────────────

		if err := page.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := page.EnterUsername(cfg.MemberLoginUsernameInputTestID, cfg.MemberCredentials.Username, timeout); err != nil {
			t.Fatalf("Failed to enter Member username: %v", err)
		}
		if err := page.EnterPassword(cfg.MemberLoginPasswordInputTestID, cfg.MemberCredentials.Password, timeout); err != nil {
			t.Fatalf("Failed to enter Member password: %v", err)
		}
		if err := page.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Member login: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Member__03_Correct_LoggedIn_Result")

		// Logout after Member login
		if err := page.Logout(cfg.LogoutButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to logout Member: %v", err)
		}
		time.Sleep(afterActionWait)
		_, _ = page.CaptureScreenshot("Member__04_After_Logout")
	})
}
