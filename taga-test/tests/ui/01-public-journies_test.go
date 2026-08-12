package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_01_PublicJourneys(t *testing.T) {
	tests.RunUITest(t, "Verify Public User Journeys", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Setup the persona and the result collector in test setup
		pubPersona := actions.NewPublicPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_01_PublicJourneys")

		// Run simple sequential action calls passing pointers
		actions.GoToHome(pubPersona, result)
		actions.GoToOfficeBeaers(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToEvents(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToMemberLogin(pubPersona, result)
		actions.GoToHome(pubPersona, result)
		actions.GoToAdminLogin(pubPersona, result)

		// Assert on captured results
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
