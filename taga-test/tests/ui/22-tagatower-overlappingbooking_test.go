package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_22_TAGATower_OverlappingBooking(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Overlapping Booking", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_22_TAGATower_OverlappingBooking")

		// Declarative Persona Action Flow
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)
		actions.NavigateToTAGATower(member, cfg, result)
		// Step 1: Select future date and book all beds in apex-1 (capacity 3) for 1 day
		actions.SelectFutureDates(member, result)
		actions.BookAllBedsInRoom(member, cfg, result, "apex-1", 3)
		// Step 2: Select a 3-day consecutive range spanning across that booked date
		actions.SelectThreeDaysOverlappingDates(member, result)
		// Step 3: Verify that booking the room for the 3-day range is blocked because the room is occupied on the middle date
		actions.TryBookOverlappingRoom(member, cfg, result, "apex-1")
		// Step 4: Cancel the original booking and clean up
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
