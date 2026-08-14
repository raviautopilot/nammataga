package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_09_EditMemberDetails(t *testing.T) {
	tests.RunUITest(t, "Admin Edit Member Details Workflow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Admin Persona and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_09_EditMemberDetails")

		// --- 1. Setup ---
		creds, _ := actions.InitializeMemberTest(cfg)
		defer actions.CleanupMemberTest(cfg, creds)

		// --- 2. Action Flow ---
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)

		// 1. Add Member
		actions.AddSingleMember(admin, cfg, creds, result)

		// 2. Edit Member Details & Verify in View panel
		actions.EditMemberDetails(admin, cfg, creds, result)

		// 3. Delete Member Cleanup
		actions.DeleteMemberByMobile(admin, cfg, creds.MobileNumber, "Step_06", result)

		// 4. Logout Admin
		actions.LogoutAdmin(admin, cfg, result)

		// Assert Result
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
