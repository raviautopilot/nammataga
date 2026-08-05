package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_15_TAGATower_AllRoomBooking(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - All Room Booking", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_15_TAGATower_AllRoomBooking")

		// 1. Login as Member
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)

		// 2. Navigate to TAGA Towers
		actions.NavigateToTAGATower(member, cfg, result)

		// 3. Book all beds in Apex Room (capacity 3) for guests
		actions.BookAllBedsInRoom(member, cfg, result, "apex-1", 3)

		// 4. Cancel the booking (Cleanup)
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
