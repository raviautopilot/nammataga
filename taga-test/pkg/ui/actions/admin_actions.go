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
		if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_LoginAsAdmin_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_LoginAsAdmin_Success"); scrErr == nil {
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
		if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_OpenAdminPanel_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_OpenAdminPanel_Success"); scrErr == nil {
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
	time.Sleep(1 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_Form_Filled_Before_Submit"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	if err := adminDashboard.SubmitAddMemberForm(cfg.AdminAddMemberSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminAddMemberSubmitButtonTestID' submit button")
		return
	}

	time.Sleep(2 * time.Second) // allow modal to close
	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_AddSingleMember_Success"); scrErr == nil {
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
		if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_DeleteMember_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_DeleteMember_Success"); scrErr == nil {
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
		if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_DeleteMember_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_DeleteMember_Success"); scrErr == nil {
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
	time.Sleep(1 * time.Second)

	if err := adminDashboard.SubmitBulkUpload(cfg.AdminBulkUploadSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify bulk upload submit button testID")
		return
	}

	time.Sleep(3 * time.Second) // allow server time to process upload
	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_BulkUpload_Success"); scrErr == nil {
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

	time.Sleep(2 * time.Second)
	if scr, scrErr := ap.Page.CaptureScreenshot(r.TestName + "_Logout_Success"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Admin logged out successfully.")
}
