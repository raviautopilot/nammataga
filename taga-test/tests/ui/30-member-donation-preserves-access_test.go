package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_30_MemberDonationPreservesPaidAccess(t *testing.T) {
	tests.RunUITest(t, "Member Donation Preserves Paid Access Status", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Personas and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_30_MemberDonationPreservesPaidAccess")

		// --- 1. Admin Flow: Add a new member (Unpaid) ---
		creds, _ := actions.InitializeMemberTest(cfg)
		defer actions.CleanupMemberTest(cfg, creds)

		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.AddSingleMember(admin, cfg, creds, result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_01_AdminLogout", result)

		// --- 2. Member Flow: Login & Pay Annual Subscription ---
		actions.GoToHome(member, result)
		actions.ForceChangePassword(member, cfg, creds.Email, creds.TempPassword, creds.NewPassword, result)
		actions.MemberLoginAttempt(member, cfg, creds.Email, creds.NewPassword, 4*time.Second, "Step_02_MemberLogin", result)
		
		// Pay annual subscription to get active paid status (isPaid = true)
		actions.PayAnnualSubscription(member, cfg, result)
		
		// Verify that member has active paid access (TAGA Towers, Resources, etc. accessible)
		actions.ValidatePaidMemberAccess(member, cfg, result)

		// --- 3. Member Flow: Make a Donation / Legal Fund contribution ---
		actions.PayNonAnnualContribution(member, cfg, "donation", result)

		// --- 4. Verification: Verify active membership access is STILL PRESERVED ---
		actions.ValidatePaidMemberAccess(member, cfg, result)
		
		actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_03_MemberLogout", result)

		// --- 5. Admin Flow: Login and delete the member ---
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
