package actions

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"e2e-template/pkg/config"
	"e2e-template/pkg/ui/pages"
)

// AdminActionsInterface allows AdminPersona to execute admin-specific actions.
type AdminActionsInterface interface {
	PublicActionsInterface
	GetAdminPersona() *AdminPersona
}

// GetAdminPersona implements AdminActionsInterface for AdminPersona.
func (p *AdminPersona) GetAdminPersona() *AdminPersona {
	return p
}

// LoginAsAdmin opens Admin login, populates credentials, submits, and verifies login.
func LoginAsAdmin(aai AdminActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Login as Admin"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	if err := ensurePage(ap.PublicPersona); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Prerequisite Check: Ensure Webdriver and AdminPersona are properly initialized")
		return
	}

	homePage := pages.NewHomePage(ap.Page)
	loginPage := pages.NewLoginPage(ap.Page)

	if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminLoginButtonTestID' button exists on Home Page")
		return
	}

	err := loginPage.FillAndSubmitLogin(
		cfg.AdminLoginUsernameInputTestID,
		cfg.AdminLoginPasswordInputTestID,
		cfg.AdminLoginSubmitButtonTestID,
		cfg.AdminCredentials.Username,
		cfg.AdminCredentials.Password,
		ap.DefaultTimeout,
	)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify Admin credentials and submit button testIDs in config.json")
		time.Sleep(1 * time.Second)
		if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_Step01_AdminLogin_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	// Wait for home page to fully render after login redirect
	time.Sleep(3 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_01_AdminLogin_HomePage"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Admin logged in successfully.")
}

// OpenAdminPanel navigates to the Admin Panel dashboard.
func OpenAdminPanel(aai AdminActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Open Admin Panel"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Check if navbar 'adminPanelButtonTestID' is visible after Admin login")
		time.Sleep(1 * time.Second)
		if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_AdminPanel_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	// Wait for admin panel table to fully render
	time.Sleep(3 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_AdminPanel_Loaded"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Admin Panel opened successfully.")
}

// AddSingleMember opens the modal, fills all 19 form fields from config, and submits.
func AddSingleMember(aai AdminActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Add Single Member with 19 Fields"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	if err := adminDashboard.OpenAddMemberModal(cfg.AdminAddMemberButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminAddMemberButtonTestID' is clickable in Admin Dashboard")
		return
	}
	time.Sleep(1 * time.Second)

	if err := adminDashboard.FillAddMemberForm(cfg, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify all 19 form input testIDs in config.json match the UI elements")
		return
	}

	// Screenshot: form fully filled, before submitting
	time.Sleep(1 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_AddMember_FormFilled"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	if err := adminDashboard.SubmitAddMemberForm(cfg.AdminAddMemberSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminAddMemberSubmitButtonTestID' submit button")
		return
	}

	// Screenshot: success toast visible immediately after submit
	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_AddMember_SuccessToast"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Click OK on the success modal to close it
	if err := ap.Page.ClickByTestID("testid-add-success-ok-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click add success OK button: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Refresh table so new member appears
	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	if scr, scrErr := ap.Page.CaptureScreenshot("Step_05_AdminPanel_AfterAddMember"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Member added successfully with all 19 fields populated.")
}

// DeleteMemberByEmail searches for a member by email and confirms deletion.
func DeleteMemberByEmail(aai AdminActionsInterface, cfg *config.Config, email string, r *Result) {
	cleanEmail := strings.TrimSpace(email)
	cleanEmail = strings.ReplaceAll(cleanEmail, " ", "")

	actionName := fmt.Sprintf("Delete Member (%s)", cleanEmail)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	// Ensure we navigate to Admin Panel cleanly
	if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		// If Admin Panel button is covered by modal backdrop, press Escape key to close modal
		_ = ap.Page.Driver.KeyDown("\uE00C") // Escape key
		_ = ap.Page.Driver.KeyUp("\uE00C")
		time.Sleep(1 * time.Second)
		_ = adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(1 * time.Second)

	err := adminDashboard.DeleteMemberByEmail(
		cfg.MemberSearchInputTestID,
		cleanEmail,
		cfg.MemberDeleteButtonTestID,
		cfg.MemberConfirmDeleteButtonTestID,
		ap.DefaultTimeout,
	)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check if member email '%s' exists in table and search input testID is correct", cleanEmail))
		time.Sleep(1 * time.Second)
		if scr, scrErr := ap.Page.CaptureScreenshot("Step_06_SearchMember_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	// Screenshot: delete toast confirmation visible
	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_07_DeleteMember_SuccessToast"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, fmt.Sprintf("Member '%s' deleted successfully.", cleanEmail))
}

// DeleteMemberByMobile searches for a member by mobile number and confirms deletion.
func DeleteMemberByMobile(aai AdminActionsInterface, cfg *config.Config, mobile string, r *Result) {
	cleanMobile := strings.TrimSpace(mobile)
	cleanMobile = strings.ReplaceAll(cleanMobile, " ", "")

	actionName := fmt.Sprintf("Delete Member (%s)", cleanMobile)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	// Ensure we navigate to Admin Panel cleanly
	if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.KeyDown("\uE00C") // Escape key
		_ = ap.Page.Driver.KeyUp("\uE00C")
		time.Sleep(1 * time.Second)
		_ = adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(1 * time.Second)

	err := adminDashboard.DeleteMemberByMobile(
		cfg.MemberSearchInputTestID,
		cleanMobile,
		cfg.MemberDeleteButtonTestID,
		cfg.MemberConfirmDeleteButtonTestID,
		ap.DefaultTimeout,
	)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check if member mobile '%s' exists in table and search input testID is correct", cleanMobile))
		time.Sleep(1 * time.Second)
		if scr, scrErr := ap.Page.CaptureScreenshot("Step_06_SearchByMobile_" + cleanMobile + "_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	// Screenshot: delete toast visible right after deletion
	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_06_DeleteMobile_" + cleanMobile + "_SuccessToast"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, fmt.Sprintf("Member with mobile '%s' deleted successfully.", cleanMobile))
}

// BulkUploadMembers uploads CSV from fixtures and submits.
func BulkUploadMembers(aai AdminActionsInterface, cfg *config.Config, fixtureRelPath string, r *Result) {
	actionName := "Bulk Member Upload from CSV"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	if err := adminDashboard.OpenBulkUploadModal(cfg.AdminBulkUploadButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminBulkUploadButtonTestID' button exists on Admin Panel")
		return
	}
	time.Sleep(1 * time.Second)

	absPath, err := filepath.Abs(fixtureRelPath)
	if err != nil {
		absPath = fixtureRelPath
	}

	if err := adminDashboard.UploadBulkFile(cfg.AdminBulkUploadFileInputTestID, absPath, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Ensure file '%s' exists and file input testID is correct", absPath))
		return
	}

	// Screenshot: CSV file selected in upload dialog
	time.Sleep(1 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_BulkUpload_FileSelected"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	if err := adminDashboard.SubmitBulkUpload(cfg.AdminBulkUploadSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify bulk upload submit button testID")
		return
	}

	// Screenshot: upload success toast / result visible
	time.Sleep(3 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_BulkUpload_SuccessToast"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	if scr, scrErr := ap.Page.CaptureScreenshot("Step_05_AdminPanel_AfterBulkUpload"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Bulk CSV uploaded successfully.")
}

// Logout Admin persona.
func LogoutAdmin(aai AdminActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Logout Admin"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	homePage := pages.NewHomePage(ap.Page)

	if err := homePage.Logout(cfg.LogoutButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'logoutButtonTestID' element exists in header/navbar")
		return
	}

	// Wait for redirect back to home / login page to fully complete
	time.Sleep(3 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_07_AdminLogout_HomePage"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Admin logged out successfully.")
}

// AdminLoginAttempt attempts admin login with the given username, password, and custom timeout/screenshot.
func AdminLoginAttempt(aai AdminActionsInterface, cfg *config.Config, username, password string, waitTime time.Duration, screenshotName string, r *Result) {
	actionName := fmt.Sprintf("Admin Login Attempt (User: %q)", username)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	homePage := pages.NewHomePage(ap.Page)
	loginPage := pages.NewLoginPage(ap.Page)

	if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminLoginButtonTestID' button exists on Home Page")
		return
	}

	// Only enter credentials if non-empty
	if username != "" {
		if err := loginPage.EnterUsername(cfg.AdminLoginUsernameInputTestID, username, ap.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			return
		}
	}
	if password != "" {
		if err := loginPage.EnterPassword(cfg.AdminLoginPasswordInputTestID, password, ap.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			return
		}
	}

	if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}

	time.Sleep(waitTime)

	if scr, scrErr := ap.Page.CaptureScreenshot(screenshotName); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
}

// LogoutAdminCustom performs logout with a customized wait time and screenshot name.
func LogoutAdminCustom(aai AdminActionsInterface, cfg *config.Config, waitTime time.Duration, screenshotName string, r *Result) {
	actionName := "Logout Admin"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	homePage := pages.NewHomePage(ap.Page)

	if err := homePage.Logout(cfg.LogoutButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'logoutButtonTestID' element exists in header/navbar")
		return
	}

	time.Sleep(waitTime)
	if scr, scrErr := ap.Page.CaptureScreenshot(screenshotName); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Admin logged out successfully.")
}

