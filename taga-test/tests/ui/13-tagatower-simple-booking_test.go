package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_13_TAGATower_SimpleBooking(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Simple Booking", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_13_TAGATower_SimpleBooking")

		// 1. Standard Login
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)

		// 2. Navigate to the Towers dashboard
		actions.NavigateToTAGATower(member, cfg, result)

		// 3. Perform a simple booking for the Apex Suite
		actions.BookSimpleRoom(member, cfg, result, "apex-1")

		// 4. Cancel the booking we just made (Cleanup)
		actions.CancelLatestBooking(member, cfg, result)

		// 5. Logout
		actions.LogoutMember(member, cfg, result)

		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
