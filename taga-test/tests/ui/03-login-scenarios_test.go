package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_03_LoginScenarios(t *testing.T) {
	tests.RunUITest(t, "Admin and Member Login Scenarios", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Setup the persona and the result collector in test setup
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("Admin and Member Login Scenarios")

		// ══════════════════════════════════════════════════════════════════
		// ADMIN LOGIN FLOW
		// ══════════════════════════════════════════════════════════════════

		// ── Attempt 1: Admin Empty Credentials ───────────────────────────
		actions.GoToHome(admin, result)
		actions.AdminLoginAttempt(admin, cfg, "", "", 2*time.Second, "Step_01_Admin_Empty_Credentials_Result", result)
		actions.GoToHome(admin, result)

		// ── Attempt 2: Admin Wrong Credentials ───────────────────────────
		actions.AdminLoginAttempt(admin, cfg, cfg.AdminCredentials.Username, "wrongpassword", 2*time.Second, "Step_02_Admin_Wrong_Credentials_Result", result)
		actions.GoToHome(admin, result)

		// ── Attempt 3: Admin Correct Credentials ─────────────────────────
		actions.AdminLoginAttempt(admin, cfg, cfg.AdminCredentials.Username, cfg.AdminCredentials.Password, 3*time.Second, "Step_03_Admin_Correct_LoggedIn_Home", result)

		// Logout before Member flow
		actions.LogoutAdminCustom(admin, cfg, 3*time.Second, "Step_04_Admin_After_Logout", result)

		// ══════════════════════════════════════════════════════════════════
		// MEMBER LOGIN FLOW
		// ══════════════════════════════════════════════════════════════════

		// ── Attempt 1: Member Empty Credentials ──────────────────────────
		actions.GoToHome(member, result)
		actions.MemberLoginAttempt(member, cfg, "", "", 2*time.Second, "Step_05_Member_Empty_Credentials_Result", result)
		actions.GoToHome(member, result)

		// ── Attempt 2: Member Wrong Credentials ──────────────────────────
		actions.MemberLoginAttempt(member, cfg, cfg.MemberCredentials.Username, "wrongpassword", 2*time.Second, "Step_06_Member_Wrong_Credentials_Result", result)
		actions.GoToHome(member, result)

		// ── Attempt 3: Member Correct Credentials ────────────────────────
		actions.MemberLoginAttempt(member, cfg, cfg.MemberCredentials.Username, cfg.MemberCredentials.Password, 3*time.Second, "Step_07_Member_Correct_LoggedIn_Dashboard", result)

		// Logout after Member login
		actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_08_Member_After_Logout", result)

		// Assert on captured results
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
