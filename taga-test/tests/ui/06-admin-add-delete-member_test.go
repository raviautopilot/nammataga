package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
	"e2e-template/tests"
)

const afterActionWaitAddMember = 2 * time.Second
const testName06 = "Admin Add and Delete Member Workflow"

func TestUI_06_AdminAddDeleteMember(t *testing.T) {
	tests.RunUITest(t, testName06, func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		url := cfg.UiURL
		timeout := 5 * time.Second

		homePage := pages.NewHomePage(page)
		loginPage := pages.NewLoginPage(page)
		adminDashboard := pages.NewAdminDashboardPage(page)

		// ── Step 1: Login as Admin ────────────────────────────────────────
		if err := homePage.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}
		if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Login: %v", err)
		}
		if err := loginPage.FillAndSubmitLogin(
			cfg.AdminLoginUsernameInputTestID,
			cfg.AdminLoginPasswordInputTestID,
			cfg.AdminLoginSubmitButtonTestID,
			cfg.AdminCredentials.Username,
			cfg.AdminCredentials.Password,
			timeout,
		); err != nil {
			t.Fatalf("Failed to login as Admin: %v", err)
		}
		time.Sleep(afterActionWaitAddMember)
		_, _ = page.CaptureScreenshot("Step_01_Admin_Login_Submitted")

		// ── Step 2: Navigate to Admin Panel ──────────────────────────────
		if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Panel: %v", err)
		}
		time.Sleep(afterActionWaitAddMember)
		_, _ = page.CaptureScreenshot("Step_02_Admin_Panel_Opened")

		// ── Step 3: Open Add Member Form ──────────────────────────────────
		if err := adminDashboard.OpenAddMemberModal(cfg.AdminAddMemberButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Add Member modal: %v", err)
		}
		time.Sleep(1 * time.Second)

		// ── Step 4: Fill ALL Fields in Add Member Form ────────────────────
		if err := adminDashboard.FillAddMemberForm(cfg, timeout); err != nil {
			t.Fatalf("Failed to fill Add Member form: %v", err)
		}
		time.Sleep(1 * time.Second)
		_, _ = page.CaptureScreenshot("Step_03_Add_Member_Form_Filled")

		// ── Step 5: Submit Add Member Form ────────────────────────────────
		if err := adminDashboard.SubmitAddMemberForm(cfg.AdminAddMemberSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to click submit Add Member button: %v", err)
		}
		time.Sleep(afterActionWaitAddMember)
		_, _ = page.CaptureScreenshot("Step_04_Member_Added_Success")

		// ── Step 6: Search New Member in Management Table by Name ─────────
		if err := adminDashboard.SearchMember(cfg.MemberSearchInputTestID, cfg.NewMemberFormData.Name, timeout); err != nil {
			t.Fatalf("Failed to search new member by name: %v", err)
		}
		time.Sleep(afterActionWaitAddMember)
		_, _ = page.CaptureScreenshot("Step_05_Member_Found_In_Table")

		// ── Step 7: Click View Details Button in Table ─────────────────────
		if err := adminDashboard.ClickFirstTableRowViewButton(timeout); err != nil {
			t.Fatalf("Failed to click View Details button in table row: %v", err)
		}
		time.Sleep(1 * time.Second)
		_, _ = page.CaptureScreenshot("Step_06_Member_View_Details_Modal_Opened")

		// ── Step 8: Delete Member (Clean Up Test Data) ───────────────────
		if err := adminDashboard.DeleteMember(
			cfg.MemberDeleteButtonTestID,
			cfg.MemberConfirmDeleteButtonTestID,
			timeout,
		); err != nil {
			t.Fatalf("Failed to delete test member: %v", err)
		}
		time.Sleep(afterActionWaitAddMember)
		_, _ = page.CaptureScreenshot("Step_07_Member_Deleted_Successfully")

		// ── Step 9: Logout Admin ─────────────────────────────────────────
		if err := homePage.Logout(cfg.LogoutButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to logout Admin: %v", err)
		}
		time.Sleep(afterActionWaitAddMember)
		_, _ = page.CaptureScreenshot("Step_08_Admin_After_Logout")
	})
}
