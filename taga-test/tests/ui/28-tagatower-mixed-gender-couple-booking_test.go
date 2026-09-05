package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_28_TAGATower_MixedGenderCoupleBooking(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Mixed Gender Couple Booking & Apex 3rd Bed Block", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_28_TAGATower_MixedGenderCoupleBooking")

		// Declarative Persona Action Flow
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)
		actions.NavigateToTAGATower(member, cfg, result)
		actions.SelectFutureDates(member, result)
		// 1. Mixed couple (1 Male + 1 Female) books 2 beds in Apex Suite -> PASS
		actions.BookMixedCoupleInRoom(member, cfg, result, "apex-1")
		// 2. Single person tries to book 3rd bed in Apex Suite -> FAILS / BLOCKED (Apex is fully booked)
		actions.TryBookSingleBedInMixedApex(member, cfg, result, "apex-1")
		// 3. Clean up couple booking
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
