package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

// TestUI_23_TAGATower_IncompleteGuestDetails verifies that when a member books
// in Guest mode with multiple beds selected, the system BLOCKS the payment flow
// if not all guest detail forms have been filled in.
//
// Scenario:
//   - Room: apex-1 (allowSingleBed = true, capacity > 1)
//   - Beds selected: 3
//   - Guest forms filled: only 1 out of 3
//
// Expected outcome:
//   - A validation toast error appears (e.g., "Please fill in all details for Guest 2")
//   - Payment flow is NOT reached
func TestUI_23_TAGATower_IncompleteGuestDetails(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Incomplete Guest Details Blocked", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_23_TAGATower_IncompleteGuestDetails")

		// Declarative Persona Action Flow
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)
		actions.NavigateToTAGATower(member, cfg, result)
		// Try to book 3 beds as guest but only fill 1 guest form — must be blocked
		actions.TryGuestBookingWithIncompleteGuestDetails(member, cfg, result, "apex-1", 3)
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
