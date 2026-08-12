package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_16_TAGATower_SelfMultibookingNegative(t *testing.T) {
	tests.RunUITest(t, "TAGA Tower - Self Multibooking Negative", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_16_TAGATower_SelfMultibookingNegative")

		// 1. Login as Member
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result)

		// 2. Navigate to TAGA Towers
		actions.NavigateToTAGATower(member, cfg, result)

		// 3. Select future dates on the calendar to ensure all beds are available (0 occupied beds)
		actions.SelectFutureDates(member, result)

		// 4. Try to book multiple beds (2 beds) in Apex Room for Self (should fail/disallowed)
		actions.TryBookRoomAsSelfMultibooking(member, cfg, result, "apex-1", 2)

		// 5. Logout (skipped if previous step failed, which is expected here)
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
