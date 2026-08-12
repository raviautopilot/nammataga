package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
	"e2e-template/tests"
)

const afterActionWaitMember = 2 * time.Second
const testNameMemberNegative = "Member Negative Login Scenarios"

func TestUI_05_MemberLoginNegative(t *testing.T) {
	tests.RunUITest(t, testNameMemberNegative, func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		url := cfg.UiURL
		timeout := 5 * time.Second

		homePage := pages.NewHomePage(page)
		loginPage := pages.NewLoginPage(page)

		// ── Step 1: Empty Username and Empty Password ─────────────────────
		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}
		if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to click Member submit: %v", err)
		}
		time.Sleep(afterActionWaitMember)
		_, _ = page.CaptureScreenshot("Step_01_Member_Empty_Credentials_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 2: Valid Username with Invalid Password ──────────────────
		if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := loginPage.EnterUsername(cfg.MemberLoginUsernameInputTestID, cfg.MemberCredentials.Username, timeout); err != nil {
			t.Fatalf("Failed to enter Member username: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.MemberLoginPasswordInputTestID, "wrongpassword123", timeout); err != nil {
			t.Fatalf("Failed to enter wrong password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Member login: %v", err)
		}
		time.Sleep(afterActionWaitMember)
		_, _ = page.CaptureScreenshot("Step_02_Member_Wrong_Password_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 3: Non-Existent Member Username ──────────────────────────
		if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := loginPage.EnterUsername(cfg.MemberLoginUsernameInputTestID, "nonexistentmember@gmail.com", timeout); err != nil {
			t.Fatalf("Failed to enter non-existent username: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.MemberLoginPasswordInputTestID, cfg.MemberCredentials.Password, timeout); err != nil {
			t.Fatalf("Failed to enter password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Member login: %v", err)
		}
		time.Sleep(afterActionWaitMember)
		_, _ = page.CaptureScreenshot("Step_03_Member_NonExistent_User_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 4: Invalid Email Format ──────────────────────────────────
		if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := loginPage.EnterUsername(cfg.MemberLoginUsernameInputTestID, "invalid-email-format", timeout); err != nil {
			t.Fatalf("Failed to enter invalid email format: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.MemberLoginPasswordInputTestID, "password123", timeout); err != nil {
			t.Fatalf("Failed to enter password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Member login: %v", err)
		}
		time.Sleep(afterActionWaitMember)
		_, _ = page.CaptureScreenshot("Step_04_Member_Invalid_Email_Format_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}

		// ── Step 5: Password Only (Username Empty) ────────────────────────
		if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		if err := loginPage.EnterPassword(cfg.MemberLoginPasswordInputTestID, cfg.MemberCredentials.Password, timeout); err != nil {
			t.Fatalf("Failed to enter password: %v", err)
		}
		if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to submit Member login: %v", err)
		}
		time.Sleep(afterActionWaitMember)
		_, _ = page.CaptureScreenshot("Step_05_Member_Password_Only_Empty_Username_Result")

		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to return Home: %v", err)
		}
	})
}
