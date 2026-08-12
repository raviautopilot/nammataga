package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_12_MemberAnnualSubscriptionPayment(t *testing.T) {
	tests.RunUITest(t, "Member Annual Subscription Payment", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_12_MemberPayment")

		// 1. Go to Home & Login
		actions.GoToHome(member, result)
		actions.LoginAsMember(member, cfg, result) // Captures Step 01 & 02

		// 2. Perform Mock Payment Flow
		actions.PayAnnualSubscription(member, cfg, result) // Captures Step 03, 04, 05

		// 3. Logout
		actions.LogoutMember(member, cfg, result) // Captures Step 13 (by default in member_actions.go)

		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
