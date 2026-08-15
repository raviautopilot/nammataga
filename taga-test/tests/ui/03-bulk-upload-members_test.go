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
		
		// Clear mock emails to prevent reading stale temporary passwords
		actions.ClearMockEmails(cfg)

		// Declarative Persona Action Flow
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.BulkUploadMembers(admin, cfg, result)
		actions.GoToHome(admin, result)
		actions.LogoutAdmin(admin, cfg, result)

		// Wait a brief moment for background emails to be processed and written
		time.Sleep(2 * time.Second)

		// Test Paid Member
		actions.VerifyBulkUploadedMember(page, cfg, cfg.BulkMemberEmails[0], "Step_Member1_Paid", result)

		// Test Unpaid Member
		actions.VerifyBulkUploadedMember(page, cfg, cfg.BulkMemberEmails[1], "Step_Member2_Unpaid", result)

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
