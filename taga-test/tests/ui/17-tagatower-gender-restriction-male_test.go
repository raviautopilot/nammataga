package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_17_TAGATower_GenderRestriction_Male(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Gender Restriction Male First", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_17_TAGATower_GenderRestriction_Male")

		// 1. Login as Member
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)

		// 2. Navigate to TAGA Towers
		actions.NavigateToTAGATower(member, cfg, result)

		// 3. Select future dates on the calendar to ensure all beds are available
		actions.SelectFutureDates(member, result)

		// 4. Book a single bed as Male
		actions.BookSingleBedWithGender(member, cfg, result, "apex-1", "male")

		// 5. Try to book a bed in the same room as Female (should fail/disallowed)
		actions.TryBookSingleBedWithGenderOpposite(member, cfg, result, "apex-1", "female")

		// 6. Cleanup: Cancel the original Male booking
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
