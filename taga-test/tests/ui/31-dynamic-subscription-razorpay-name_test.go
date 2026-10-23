package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_31_DynamicSubscriptionRazorpayName(t *testing.T) {
	tests.RunUITest(t, "Dynamic Subscription Name on Razorpay Checkout Modal", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		// Initialize Personas and Result collector
		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_31_DynamicSubscriptionRazorpayName")

		// --- 1. Admin Flow: Add a new member (Unpaid) ---
		creds, _ := actions.InitializeMemberTest(cfg)
		defer actions.CleanupMemberTest(cfg, creds)

		actions.GoToHome(admin, result)
		actions.LoginAsAdmin(admin, cfg, result)
		actions.OpenAdminPanel(admin, cfg, result)
		actions.AddSingleMember(admin, cfg, creds, result)
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_01_AdminLogout", result)

		// --- 2. Member Flow: Change password, login ---
		actions.GoToHome(member, result)
		actions.ForceChangePassword(member, cfg, creds.Email, creds.TempPassword, creds.NewPassword, result)
		actions.MemberLoginAttempt(member, cfg, creds.Email, creds.NewPassword, 4*time.Second, "Step_02_MemberLogin", result)

		// --- 3. Verify Dynamic Razorpay Modal Options across different subscription types ---

		// 3a. Test TBF New Registration Fee
		actions.VerifySubscriptionRazorpayOptions(
			member,
			cfg,
			"tbf-new-registration",
			"TBF New Registration Fee",
			"One-time registration fee for TBF",
			result,
		)

		// 3b. Test TBF Additional Amount (Need-based / custom amount)
		actions.VerifySubscriptionRazorpayOptions(
			member,
			cfg,
			"tbf-additional",
			"TBF Additional Amount",
			"Additional contribution to TBF (custom amount)",
			result,
		)

		// 3c. Test Annual Subscription
		actions.VerifySubscriptionRazorpayOptions(
			member,
			cfg,
			"annual-subscription",
			"Annual Subscription",
			"Yearly membership fee (April 1 - March 31)",
			result,
		)

		// --- 4. Member Flow: Logout ---
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
