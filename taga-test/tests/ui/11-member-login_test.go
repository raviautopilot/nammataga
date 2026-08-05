package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_11_MemberLoginHappyPath(t *testing.T) {
	tests.RunUITest(t, "Member Login Happy Path", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_11_MemberLogin")

		// Declarative Persona Action Flow
		actions.GoToHome(member, result)
		
		// Login Member (using new persona action)
		actions.LoginAsMember(member, cfg, result)

		// Visit all accessible pages for member
		actions.VisitAllMemberPages(member, cfg, result)

		// Logout Member (using new persona action)
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
