package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_11_MemberLoginHappyPath(t *testing.T) {
	tests.RunUITest(t, "Member Login Happy Path", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Admin & Member Personas and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_11_MemberLogin")

		// 1. Initialize dynamic test credentials & cleanup defer
		creds, _ := actions.InitializeMemberTest(cfg)
		defer actions.CleanupMemberTest(cfg, creds)

		// 2. Admin Flow: Login as admin and add a new single unpaid member
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.AddSingleMember(admin, cfg, creds, result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_01_AdminLogout", result)

		// 3. Member Flow: Change password, login as unpaid member, validate restricted access, logout
		actions.GoToHome(member, result)
		actions.ForceChangePassword(member, cfg, creds.Email, creds.TempPassword, creds.NewPassword, result)
		actions.MemberLoginAttempt(member, cfg, creds.Email, creds.NewPassword, 4*time.Second, "Step_02_MemberLogin_Success", result)
		actions.ValidateUnpaidMemberAccess(member, cfg, result)
		actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_03_MemberLogout", result)

		// 4. Admin Flow: Login as admin and delete the created test member
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.DeleteMemberByMobile(admin, cfg, creds.MobileNumber, "Step_04", result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_05_AdminLogout", result)

		// Assert Result
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
