package ui_tests

import (
	"path/filepath"
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
	"e2e-template/tests"
)

const afterActionWaitBulkUpload = 2 * time.Second
const testName07 = "Admin Bulk Member Upload and Automated Cleanup Workflow"

func TestUI_07_BulkUploadMembers(t *testing.T) {
	tests.RunUITest(t, testName07, func(t *testing.T, page *ui.Page) {
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
		time.Sleep(afterActionWaitBulkUpload)
		_, _ = page.CaptureScreenshot("Step_01_Admin_Login_Submitted")

		// ── Step 2: Navigate to Admin Panel ──────────────────────────────
		if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Admin Panel: %v", err)
		}
		time.Sleep(afterActionWaitBulkUpload)
		_, _ = page.CaptureScreenshot("Step_02_Admin_Panel_Opened")

		// ── Step 3: Open Bulk Upload Modal ────────────────────────────────
		if err := adminDashboard.OpenBulkUploadModal(cfg.AdminBulkUploadButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to open Bulk Upload modal: %v", err)
		}
		time.Sleep(1 * time.Second)
		_, _ = page.CaptureScreenshot("Step_03_Bulk_Upload_Modal_Opened")

		// ── Step 4: Attach CSV File from Fixtures ─────────────────────────
		sampleCSVPath, err := filepath.Abs("../../fixtures/bulk_members_sample.csv")
		if err != nil {
			sampleCSVPath = "fixtures/bulk_members_sample.csv"
		}

		if err := adminDashboard.UploadBulkFile(cfg.AdminBulkUploadFileInputTestID, sampleCSVPath, timeout); err != nil {
			t.Fatalf("Failed to attach bulk CSV file (%s): %v", sampleCSVPath, err)
		}
		time.Sleep(1 * time.Second)
		_, _ = page.CaptureScreenshot("Step_04_Bulk_Upload_File_Selected")

		// ── Step 5: Submit Bulk Upload ────────────────────────────────────
		if err := adminDashboard.SubmitBulkUpload(cfg.AdminBulkUploadSubmitButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to click Submit Bulk Upload button: %v", err)
		}
		time.Sleep(4 * time.Second) // allow server time to process upload & modal to close
		_, _ = page.CaptureScreenshot("Step_05_Bulk_Upload_Submitted_Successfully")

		// ── Step 6: Verify and Delete All 5 Uploaded Test Members ──────────
		for i, email := range cfg.BulkMemberEmails {
			if err := adminDashboard.DeleteMemberByEmail(
				cfg.MemberSearchInputTestID,
				email,
				cfg.MemberDeleteButtonTestID,
				cfg.MemberConfirmDeleteButtonTestID,
				timeout,
			); err != nil {
				t.Fatalf("Failed to verify and delete bulk uploaded member %d (%s): %v", i+1, email, err)
			}
			time.Sleep(1 * time.Second)
		}
		time.Sleep(afterActionWaitBulkUpload)
		_, _ = page.CaptureScreenshot("Step_06_All_5_Bulk_Members_Cleaned_Up")

		// ── Step 7: Logout Admin ─────────────────────────────────────────
		if err := homePage.Logout(cfg.LogoutButtonTestID, timeout); err != nil {
			t.Fatalf("Failed to logout Admin: %v", err)
		}
		time.Sleep(afterActionWaitBulkUpload)
		_, _ = page.CaptureScreenshot("Step_07_Admin_After_Logout")
	})
}
