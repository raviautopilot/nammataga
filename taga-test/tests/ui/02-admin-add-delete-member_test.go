package ui_tests

import (
	"strings"
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_06_AdminAddDeleteMember(t *testing.T) {
	tests.RunUITest(t, "Admin Add and Delete Member Workflow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Admin Persona and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_06_AdminAddDeleteMember")

		// Declarative Persona Action Flow
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.AddSingleMember(admin, cfg, result)

		// Defer cleanup if member was added successfully
		if !result.Failed() {
			defer func() {
				cleanupResult := actions.NewResult("Cleanup")
				cleanMobile := strings.TrimSpace(cfg.NewMemberFormData.MobileNumber)
				cleanMobile = strings.ReplaceAll(cleanMobile, " ", "")
				actions.DeleteMemberByMobile(admin, cfg, cleanMobile, cleanupResult)
			}()
		}

		actions.SetPaymentStatusToPaid(admin, cfg, result)
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
