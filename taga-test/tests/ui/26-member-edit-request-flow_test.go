package ui_test

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

func TestUI_26_MemberEditRequestFlow(t *testing.T) {
	tests.RunUITest(t, "Member Edit Request Flow", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig

		admin := actions.NewAdminPersona(page, cfg.UiURL, 5*time.Second)
		member := actions.NewMemberPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_26_MemberEditRequest")

		// 1. Member Logs In (using the seeded subscriber)
		actions.GoToHome(member, result)
		actions.MemberLoginAttempt(member, cfg, cfg.MemberCredentials.Username, cfg.MemberCredentials.Password, 5*time.Second, "member_login", result)

		// 2. Member Submits Edit Request
		actions.MemberSubmitsEditRequest(member, cfg, result)

		// 3. Member Logs Out
		actions.LogoutMemberCustom(member, cfg, 2*time.Second, "Step_MemberLogout", result)
		
		// 4. Admin Logs In
		actions.GoToHome(admin, result)
		actions.AdminLoginAttempt(admin, cfg, cfg.AdminCredentials.Username, cfg.AdminCredentials.Password, 5*time.Second, "admin_login", result)
		actions.OpenAdminPanel(admin, cfg, result)

		// 5. Admin Processes Edit Request
		actions.AdminProcessesEditRequest(admin, cfg, result)

		// 6. Admin Logs Out
		actions.LogoutAdminCustom(admin, cfg, 2*time.Second, "Step_AdminLogout", result)
	})
}
