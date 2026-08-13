package ui_test

import (
	"strings"
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_02_AdminAddDeleteMember(t *testing.T) {
	tests.RunUITest(t, "Admin Add and Delete Member Workflow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Admin Persona and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_02_AdminAddDeleteMember")

		// Pre-test setup: Ensure target test member is deleted via API so test runs with a clean state
		cfg.NewMemberFormData.PaymentStatus = "Unpaid"
		cleanEmail := strings.TrimSpace(cfg.NewMemberFormData.Email)
		cleanMobile := strings.TrimSpace(cfg.NewMemberFormData.MobileNumber)
		cleanMobileNoSpace := strings.ReplaceAll(cleanMobile, " ", "")
		tests.CleanupMemberByEmailOrMobile(cfg, cleanEmail, cleanMobile)

		// Defer fallback API cleanup to guarantee clean DB state if any step fails
		defer func() {
			tests.CleanupMemberByEmailOrMobile(cfg, cleanEmail, cleanMobile)
		}()

		// Declarative Persona Action Flow
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		tempPassword := actions.AddSingleMember(admin, cfg, result)

		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_05b_AdminLogout", result)

		// Member Flow: Click change password, login with new password, and logout
		if tempPassword != "" {
			actions.GoToHome(member, result)

			// Open change password dialog explicitly and change password
			actions.ForceChangePassword(member, cfg, cleanEmail, tempPassword, "test123", result)

			// Login with new password
			actions.MemberLoginAttempt(member, cfg, cleanEmail, "test123", 4*time.Second, "Step_08_MemberLogin_Success", result)

			// View subscriptions
			actions.ViewMemberSubscriptions(member, cfg, "Step_08b_ViewSubscriptions", result)

			// Logout member
			actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_09_MemberLogout", result)
		}

		// Back to Admin: Login and delete the member
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.DeleteMemberByMobile(admin, cfg, cleanMobileNoSpace, "Step_10", result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_11_AdminLogout", result)

		// Assert Result
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
