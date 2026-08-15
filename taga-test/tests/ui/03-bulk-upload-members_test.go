package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_03_BulkUploadMembers(t *testing.T) {
	tests.RunUITest(t, "Admin Bulk Member Upload and Automated Cleanup Workflow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Admin Persona and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_03_BulkUploadMembers")

		// Declarative Persona Action Flow
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.BulkUploadMembers(admin, cfg, result)
		actions.GoToHome(admin, result)
		actions.LogoutAdmin(admin, cfg, result)

		// Overriding the hashed DB passwords directly because fetching them from the API is impossible (BUG).
		actions.ForceMemberPasswords(cfg.BulkMemberEmails, cfg.MemberCredentials.Password)

		// Test Paid Member
		member1 := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		actions.VerifyMemberSubscriptionStatus(member1, cfg, cfg.BulkMemberEmails[0], cfg.MemberCredentials.Password, "Step_Member1_Paid", result)

		// Test Unpaid Member
		member2 := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		actions.VerifyMemberSubscriptionStatus(member2, cfg, cfg.BulkMemberEmails[1], cfg.MemberCredentials.Password, "Step_Member2_Unpaid", result)

		// Run automated UI cleanup of uploaded members synchronously before logging out
		actions.LoginAsAdmin(admin, cfg, result)
		actions.GoToHome(admin, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.BulkCleanupMembers(admin, cfg, result)
		actions.GoToHome(admin, result)
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
