package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_18_TAGATower_GenderRestriction_Female(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Gender Restriction Female First", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_18_TAGATower_GenderRestriction_Female")

		// Declarative Persona Action Flow
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)
		actions.NavigateToTAGATower(member, cfg, result)
		actions.SelectFutureDates(member, result)
		actions.BookSingleBedWithGender(member, cfg, result, "apex-1", "female")
		actions.TryBookSingleBedWithGenderOpposite(member, cfg, result, "apex-1", "male")
		actions.CancelLatestBooking(member, cfg, result)
		actions.LogoutMember(member, cfg, result)

		// Assert Result
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
