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

		// 1. Login as Member
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)

		// 2. Navigate to TAGA Towers
		actions.NavigateToTAGATower(member, cfg, result)

		// 3. Select a future date range (tomorrow and day after)
		actions.SelectFutureDates(member, result)

		// 4. Book suite room "pavalam" (capacity 1) for this date range
		actions.BookSimpleRoom(member, cfg, result, "pavalam")

		// 5. Try to book the same room for the same date range (should be blocked)
		actions.TryBookOverlappingRoom(member, cfg, result, "pavalam")

		// 6. Cancel the first booking (Cleanup)
		actions.CancelLatestBooking(member, cfg, result)

		// 7. Logout
		actions.LogoutMember(member, cfg, result)

		if result.Failed() {
			t.Errorf("=================================================================")
			t.Errorf("🛑 TEST JOURNEY FAILED: %v", result.Error)
			t.Errorf("🛑 Actions Attempted: %v", result.Actions)
			t.Errorf("🛑 Evidence Captured: %v", result.Evidence)
			t.Fatalf("🛑 Advice / Remediation: %v", result.Advice)
			t.Errorf("=================================================================")
		}
	})
}
