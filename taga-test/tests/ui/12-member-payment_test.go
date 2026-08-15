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

		// Initialize Personas and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_12_MemberPayment")

		// --- 1. Admin Flow: Add a new member (Unpaid) ---
		creds, _ := actions.InitializeMemberTest(cfg)
		defer actions.CleanupMemberTest(cfg, creds)
		
		// Ensure the new member is unpaid initially so the pay button appears
		// (This is already handled inside InitializeMemberTest by mutating cfg)

		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.AddSingleMember(admin, cfg, creds, result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_01_AdminLogout", result)

		// --- 2. Member Flow: Change password, login, pay subscription, logout ---
		actions.GoToHome(member, result)
		actions.ForceChangePassword(member, cfg, creds.Email, creds.TempPassword, creds.NewPassword, result)
		actions.MemberLoginAttempt(member, cfg, creds.Email, creds.NewPassword, 4*time.Second, "Step_02_MemberLogin", result)
		
		// Execute the payment flow UI steps
		actions.PayAnnualSubscription(member, cfg, result)
		
		actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_03_MemberLogout", result)

		// --- 3. Admin Flow: Login and delete the member ---
		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.DeleteMemberByMobile(admin, cfg, creds.MobileNumber, "Step_04_AdminDelete", result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_05_AdminLogout", result)

		// Assert Result
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}
	})
}
