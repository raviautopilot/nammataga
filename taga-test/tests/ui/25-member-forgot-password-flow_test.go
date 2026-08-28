package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_25_MemberForgotPasswordFlow(t *testing.T) {
	tests.RunUITest(t, "Member Forgot Password Flow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_25_MemberForgotPasswordFlow")

		creds, _ := actions.InitializeMemberTest(cfg)
		defer actions.CleanupMemberTest(cfg, creds)

		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.AddSingleMember(admin, cfg, creds, result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_01_AdminLogout", result)

		actions.ClearMockEmails(cfg)

		actions.GoToHome(member, result)
		actions.ForgotPasswordByEmail(member, cfg, creds.Email, result)
		actions.ResetForgotPasswordWithTemporaryPassword(member, cfg, creds.Email, "NewSecurePassword123!", result)
		actions.MemberLoginAttempt(member, cfg, creds.Email, "NewSecurePassword123!", 4*time.Second, "Step_02_MemberLogin_Success", result)
		actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_03_MemberLogout", result)

		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.DeleteMemberByMobile(admin, cfg, creds.MobileNumber, "Step_04", result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_05_AdminLogout", result)

		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
