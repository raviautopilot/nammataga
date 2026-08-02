package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_11_ManageEvents(t *testing.T) {
	tests.RunUITest(t, "Admin Manage Events Workflow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Admin Persona and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_07_ManageEvents")

		// Calculate tomorrow's date to ensure valid upcoming event date (YYYY-MM-DD)
		tomorrowDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")

		// Declarative Persona Action Flow
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.ManageEventAction(
			admin,
			cfg,
			"AA Test Annual Conference 2026",
			tomorrowDate,
			"10:00",
			"Chennai Convention Hall",
			"Annual state level agriculture officers conference and workshop.",
			result,
		)
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
