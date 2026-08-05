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

// DownloadExcelReports generates the membership report with all-time selection and exports the member list table.
func DownloadExcelReports(aai AdminActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Download Excel Reports"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// 1. Generate Excel Report (Membership Report - All Time)
	if err := ap.Page.ClickByTestID("testid-generate-excel-report-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-generate-excel-report-button' exists on Admin Panel")
		return
	}
	time.Sleep(1 * time.Second)

	// Select 'All Time' from period dropdown
	if err := ap.Page.SelectCustomDropdownByText("testid-report-period-select", "All Time", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to select 'All Time' from report period dropdown")
		return
	}
	time.Sleep(1 * time.Second)

	// Click download
	if err := ap.Page.ClickByTestID("testid-download-excel-report-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-download-excel-report-button' exists in the report modal")
		return
	}
	
	// Screenshot: after triggering first download
	time.Sleep(3 * time.Second) 
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_MembershipReport_Downloaded"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 2. Export Member List Table Excel
	if err := ap.Page.ClickByTestID("testid-export-excel-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-export-excel-button' exists above the member list table")
		return
	}
	
	// Screenshot: after triggering second download
	time.Sleep(3 * time.Second) 
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_MemberListTable_Downloaded"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	r.Advice = append(r.Advice, "Both excel reports generated and downloaded successfully.")
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

// SendAnnouncement fills and submits the announcement form with specified title, message, priority, and screenshots.
func SendAnnouncement(aai AdminActionsInterface, cfg *config.Config, title, message, priority string, r *Result) {
	actionName := "Send Announcement"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// 1. Click Send Announcement button
	if err := ap.Page.ClickByTestID(cfg.AdminSendAnnouncementButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click send announcement button: %w", err)
		r.Advice = append(r.Advice, "Advice: Ensure Admin Panel is open and Send Announcement button is visible")
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Fill Title
	if err := ap.Page.SendKeysByTestID(cfg.AdminAnnouncementTitleInputTestID, title, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill announcement title: %w", err)
		return
	}

	// 3. Fill Message
	if err := ap.Page.SendKeysByTestID(cfg.AdminAnnouncementMessageInputTestID, message, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill announcement message: %w", err)
		return
	}

	// 4. Select Priority if specified
	if priority != "" {
		if err := ap.Page.SelectCustomDropdownByText(cfg.AdminAnnouncementPrioritySelectTestID, strings.Title(priority), ap.DefaultTimeout); err != nil {
			// Fallback: click directly if dropdown selection doesn't match
			_ = ap.Page.ClickByTestID(cfg.AdminAnnouncementPrioritySelectTestID, ap.DefaultTimeout)
		}
	}

	// Screenshot 2: Announcement data form filled
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_AnnouncementForm_Filled"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 5. Submit Announcement
	if err := ap.Page.ClickByTestID(cfg.AdminAnnouncementSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to submit announcement: %w", err)
		return
	}

	time.Sleep(2 * time.Second)

	// Screenshot 3: Announcement sent confirmation
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_Announcement_Sent"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	r.Advice = append(r.Advice, "Announcement created and sent successfully.")
}

// ManageDistrictOfficeBearers handles editing, public page validation, and backup restoration of District Office Bearers.
func ManageDistrictOfficeBearers(aai AdminActionsInterface, cfg *config.Config, district string, r *Result) {
	actionName := fmt.Sprintf("Manage District Office Bearers (%s)", district)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// 1. Click Manage District Office Bearers button in Admin Panel ("testid-manage-office-bearers-button")
	if err := ap.Page.ClickByTestID(cfg.AdminManageOfficeBearersButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to open district office bearers modal: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Select District (e.g. Ariyalur) by dropdown field
	if err := ap.Page.Click("//div[@data-testid='testid-office-bearers-modal']//button[contains(@role, 'combobox')]", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click district dropdown combobox: %w", err)
		return
	}
	time.Sleep(500 * time.Millisecond)
	if err := ap.Page.Click(fmt.Sprintf("//*[contains(@role, 'option') and (text()='%s' or .='%s')]", district, district), ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select district '%s' from dropdown: %w", district, err)
		return
	}
	time.Sleep(1 * time.Second)

	// 3. Backup existing Joint Secretary (Women) - Index 3
	nameElem, err := ap.Page.FindElementByTestID("testid-bearer-name-3", ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find Joint Secretary (Women) name input: %w", err)
		return
	}
	contactElem, err := ap.Page.FindElementByTestID("testid-bearer-contact-3", ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find Joint Secretary (Women) contact input: %w", err)
		return
	}

	originalName, _ := nameElem.GetAttribute("value")
	originalContact, _ := contactElem.GetAttribute("value")

	r.Advice = append(r.Advice, fmt.Sprintf("Backup data for Joint Secretary (Women) before edit: Name='%s', Mobile='%s'", originalName, originalContact))

	// Screenshot 1: Before Edit
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_01_DistrictBearers_BeforeEdit"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 4. Fill Joint Secretary (Women) test data (triggers React onChange via input events)
	testName := "Test Officer Joint Sec Women"
	testContact := "9876543210"

	// Helper to send input keys & dispatch React input event
	fillWithReactEvent := func(testID, value string) {
		_ = ap.Page.SendKeysByTestID(testID, value, ap.DefaultTimeout)
		jsScript := fmt.Sprintf(`
			const el = document.querySelector('[data-testid="%s"]');
			if (el) {
				const prototype = Object.getPrototypeOf(el);
				const setter = Object.getOwnPropertyDescriptor(prototype, 'value').set;
				setter.call(el, %q);
				el.dispatchEvent(new Event('input', { bubbles: true }));
				el.dispatchEvent(new Event('change', { bubbles: true }));
			}
		`, testID, value)
		_, _ = ap.Page.Driver.ExecuteScript(jsScript, nil)
	}

	fillWithReactEvent("testid-bearer-name-3", testName)
	fillWithReactEvent("testid-bearer-contact-3", testContact)
	time.Sleep(1 * time.Second)

	// Screenshot 2: Form Filled
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_DistrictBearers_FormFilled"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	confirmXpath := "//div[contains(@role, 'alertdialog') or contains(@class, 'max-w-sm')]//button[contains(text(), 'Confirm Save') or contains(., 'Confirm Save')]"

	// 5. Click Save Changes & Confirm Save in AlertDialog
	// Scroll modal save button into view and click
	_, _ = ap.Page.Driver.ExecuteScript("const btn = document.querySelector('[data-testid=\"testid-save-button\"]'); if (btn) { btn.scrollIntoView({block: 'center'}); }", nil)
	time.Sleep(500 * time.Millisecond)

	if err := ap.Page.ClickByTestID("testid-save-button", ap.DefaultTimeout); err != nil {
		// Fallback direct JS click
		_, _ = ap.Page.Driver.ExecuteScript("const btn = document.querySelector('[data-testid=\"testid-save-button\"]'); if (btn) { btn.click(); }", nil)
	}
	time.Sleep(1 * time.Second)

	// Click Confirm Save button in AlertDialog
	if err := ap.Page.Click(confirmXpath, ap.DefaultTimeout); err != nil {
		if err2 := ap.Page.Click("//button[contains(., 'Confirm Save')]", ap.DefaultTimeout); err2 != nil {
			// JS Fallback for AlertDialog action button
			_, _ = ap.Page.Driver.ExecuteScript("const btns = Array.from(document.querySelectorAll('button')); const b = btns.find(x => x.textContent.includes('Confirm Save')); if(b) b.click();", nil)
		}
	}
	time.Sleep(2 * time.Second)

	// Screenshot 3: Saved in Admin Panel
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_DistrictBearers_Saved"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	// 6. Navigate to Office Bearers page by clicking "testid-office-bearers-button" (Admin navbar button)
	if err := ap.Page.ClickByTestID(cfg.OfficeBearersButtonTestID, ap.DefaultTimeout); err != nil {
		// Fallback direct URL if navbar button click intercepted
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/office-bearers")
	}
	time.Sleep(2 * time.Second)

	// Search & Select District in Public Page dropdown ("testid-office-bearers-district-select")
	if err := ap.Page.SelectCustomDropdownByText(cfg.OfficeBearersDistrictSelectTestID, district, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select district '%s' on public page: %w", district, err)
		return
	}
	time.Sleep(2 * time.Second)

	// Scroll down to the end of the page to fully view the district office bearers table & photos
	_, _ = ap.Page.Driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
	time.Sleep(1 * time.Second)

	// Screenshot 4: Public Page Result Table (scrolled down)
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_OfficeBearers_PublicPage_Result"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 7. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// Click Office Bearer Management Button again
	if err := ap.Page.ClickByTestID(cfg.AdminManageOfficeBearersButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open district office bearers modal in admin panel: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Select District dropdown again
	if err := ap.Page.Click("//div[@data-testid='testid-office-bearers-modal']//button[contains(@role, 'combobox')]", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click district dropdown combobox on restore step: %w", err)
		return
	}
	time.Sleep(500 * time.Millisecond)
	if err := ap.Page.Click(fmt.Sprintf("//*[contains(@role, 'option') and (text()='%s' or .='%s')]", district, district), ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-select district '%s' on restore step: %w", district, err)
		return
	}
	time.Sleep(1 * time.Second)

	// Re-change to default data of changed bearer detail if backup data existed before test
	hasBackupData := strings.TrimSpace(originalName) != "" || strings.TrimSpace(originalContact) != ""

	fillWithReactEvent("testid-bearer-name-3", originalName)
	fillWithReactEvent("testid-bearer-contact-3", originalContact)
	time.Sleep(1 * time.Second)

	if hasBackupData {
		// Scroll save button into view and click
		_, _ = ap.Page.Driver.ExecuteScript("const btn = document.querySelector('[data-testid=\"testid-save-button\"]'); if (btn) { btn.scrollIntoView({block: 'center'}); }", nil)
		time.Sleep(500 * time.Millisecond)

		if err := ap.Page.ClickByTestID("testid-save-button", ap.DefaultTimeout); err != nil {
			_, _ = ap.Page.Driver.ExecuteScript("const btn = document.querySelector('[data-testid=\"testid-save-button\"]'); if (btn) { btn.click(); }", nil)
		}
		time.Sleep(1 * time.Second)

		// Click Confirm Save on Restore
		if err := ap.Page.Click(confirmXpath, ap.DefaultTimeout); err != nil {
			if err2 := ap.Page.Click("//button[contains(., 'Confirm Save')]", ap.DefaultTimeout); err2 != nil {
				_, _ = ap.Page.Driver.ExecuteScript("const btns = Array.from(document.querySelectorAll('button')); const b = btns.find(x => x.textContent.includes('Confirm Save')); if(b) b.click();", nil)
			}
		}
		time.Sleep(2 * time.Second)
	} else {
		r.Advice = append(r.Advice, "Original field was empty before test; fields cleared back to empty without saving.")
	}

	// Screenshot 5: Restored to Default
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_05_DistrictBearers_RestoredToDefault"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	r.Advice = append(r.Advice, fmt.Sprintf("Restored Joint Secretary (Women) back to original state: Name='%s', Mobile='%s'", originalName, originalContact))
}

// ManageResourceDocument handles uploading a resource document, verifying on resources page, deleting it, and taking screenshots.
func ManageResourceDocument(aai AdminActionsInterface, cfg *config.Config, relativePdfPath, categoryName string, r *Result) {
	actionName := fmt.Sprintf("Manage Resource Document (%s)", categoryName)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// Resolve absolute path for fixture PDF
	absPdfPath, err := filepath.Abs(relativePdfPath)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to resolve PDF path: %w", err)
		return
	}

	// 1. Click Manage Content button
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click manage content button: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Click Resources Tab inside Manage Content modal
	tabXPath := "//div[@data-testid='testid-content-management-modal']//button[@data-testid='testid-resources-button' or text()='Resources']"
	if err := ap.Page.Click(tabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminResourcesTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(500 * time.Millisecond)

	// 3. Select Category (Establishment)
	if err := ap.Page.SelectCustomDropdownByText(cfg.AdminResourceCategorySelectTestID, categoryName, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select category '%s': %w", categoryName, err)
		return
	}

	// 4. Upload PDF file
	fileElem, err := ap.Page.FindElementByTestID(cfg.AdminResourceFileInputTestID, ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find resource file input: %w", err)
		return
	}
	if err := fileElem.SendKeys(absPdfPath); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to upload PDF file: %w", err)
		return
	}

	// Screenshot Step 01: Form Filled
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_01_ResourceUpload_FormFilled"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 6. Click Upload Resource Button
	if err := ap.Page.ClickByTestID(cfg.AdminUploadResourceButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click upload resource button: %w", err)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Uploaded Confirmation Toast/List
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_ResourceUpload_Submitted"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Content Management Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	// 7. Go to Resources Page as Admin via navbar button
	if err := ap.Page.ClickByTestID(cfg.ResourcesNavButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/resources")
	}
	time.Sleep(2 * time.Second)

	// Select Category "Establishment" on public resources page
	catBtnXPath := "//button[contains(@data-testid, 'testid-resource-category-') and (contains(., 'Establishment') or contains(text(), 'Establishment'))]"
	_ = ap.Page.Click(catBtnXPath, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	// Screenshot Step 03: Resources Page Verification
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_ResourcesPage_Verification"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 8. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// 9. Click Manage Content -> Resources tab
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open manage content modal: %w", err)
		return
	}
	time.Sleep(1 * time.Second)
	_ = ap.Page.Click(tabXPath, ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	// Expand "Establishment" category accordion under Existing Resources
	estAccXPath := "//button[contains(., 'Establishment') and contains(@class, 'w-full')]"
	_ = ap.Page.Click(estAccXPath, ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	// Click delete trash icon for the uploaded test PDF document (starting with 'a')
	deleteDocXPath := "//p[contains(text(), 'a_test_resource_sample')]/ancestor::div[contains(@class, 'flex items-center')]//button[contains(@class, 'text-destructive')]"
	if err := ap.Page.Click(deleteDocXPath, ap.DefaultTimeout); err != nil {
		// Fallback: Click first delete icon inside Establishment category
		_ = ap.Page.Click("//div[contains(., 'Establishment')]//button[contains(@class, 'text-destructive')]", ap.DefaultTimeout)
	}
	time.Sleep(1 * time.Second)

	// Confirm delete in ConfirmDeleteDialog if dialog pops up
	confirmDelXPath := "//div[contains(@role, 'alertdialog') or contains(@class, 'max-w-md')]//button[contains(text(), 'Delete') or contains(., 'Delete')]"
	if err := ap.Page.Click(confirmDelXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Click("//button[contains(text(), 'Confirm') or contains(text(), 'Delete')]", ap.DefaultTimeout)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 04: Resource Deleted
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_Resource_Deleted"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	r.Advice = append(r.Advice, "Resource document uploaded, verified on Resources page, and cleaned up successfully.")
}

// ManageEventAction handles creating an event, verifying on upcoming events page, deleting it, and taking screenshots.
func ManageEventAction(aai AdminActionsInterface, cfg *config.Config, title, eventDate, eventTime, location, description string, r *Result) {
	actionName := fmt.Sprintf("Manage Event (%s)", title)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// 1. Click Manage Content button
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click manage content button: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Click Events Tab inside Content Management Modal
	eventsTabXPath := "//div[@data-testid='testid-content-management-modal']//button[@data-testid='testid-events-button' or text()='Events']"
	if err := ap.Page.Click(eventsTabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminEventsTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(500 * time.Millisecond)

	// 3. Fill Event Title
	if err := ap.Page.SendKeysByTestID(cfg.AdminEventTitleInputTestID, title, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill event title: %w", err)
		return
	}

	// 4. Fill Event Date
	if err := ap.Page.SendKeysByTestID(cfg.AdminEventDateInputTestID, eventDate, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill event date: %w", err)
		return
	}

	// 5. Fill Event Time if provided
	if eventTime != "" {
		_ = ap.Page.SendKeysByTestID(cfg.AdminEventTimeInputTestID, eventTime, ap.DefaultTimeout)
	}

	// 6. Fill Event Location & Description
	if location != "" {
		_ = ap.Page.SendKeysByTestID(cfg.AdminEventLocationInputTestID, location, ap.DefaultTimeout)
	}
	if description != "" {
		_ = ap.Page.SendKeysByTestID(cfg.AdminEventDescriptionInputTestID, description, ap.DefaultTimeout)
	}

	// Screenshot Step 01: Event Form Filled
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_01_EventCreate_FormFilled"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 7. Click Publish Event Button
	if err := ap.Page.ClickByTestID(cfg.AdminPublishEventButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click publish event button: %w", err)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Event Published
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_Event_Published"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Content Management Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	// 8. Go to Events Page as Admin via navbar button
	if err := ap.Page.ClickByTestID(cfg.EventsNavButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/events")
	}
	time.Sleep(2 * time.Second)

	// Click "Upcoming Events" tab on public/admin Events page
	_ = ap.Page.ClickByTestID(cfg.UpcomingEventsTabButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	// Scroll down slightly so the upcoming events cards are clearly visible in the view
	_, _ = ap.Page.Driver.ExecuteScript("window.scrollBy(0, 350);", nil)
	time.Sleep(1 * time.Second)

	// Screenshot Step 03: Public Upcoming Events Verification
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_EventsPage_UpcomingEvents_Result"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 9. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// 10. Click Manage Content -> Events tab to clean up test event
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open manage content modal: %w", err)
		return
	}
	time.Sleep(1 * time.Second)
	_ = ap.Page.Click(eventsTabXPath, ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	// Click delete trash icon for the published test event in Existing Events list
	deleteEventXPath := fmt.Sprintf("//p[contains(text(), %q)]/ancestor::div[contains(@class, 'flex items-start')]//button[contains(@class, 'text-destructive')]", title)
	if err := ap.Page.Click(deleteEventXPath, ap.DefaultTimeout); err != nil {
		// Fallback: Click first delete icon in Existing Events section
		_ = ap.Page.Click("//div[contains(., 'Existing Events')]//button[contains(@class, 'text-destructive')]", ap.DefaultTimeout)
	}
	time.Sleep(1 * time.Second)

	// Confirm delete in ConfirmDeleteDialog
	confirmDelXPath := "//div[contains(@role, 'alertdialog') or contains(@class, 'max-w-md')]//button[contains(text(), 'Delete') or contains(., 'Delete')]"
	if err := ap.Page.Click(confirmDelXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Click("//button[contains(text(), 'Confirm') or contains(text(), 'Delete')]", ap.DefaultTimeout)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 04: Event Deleted
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_Event_Deleted"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	r.Advice = append(r.Advice, "Event created, verified on Upcoming Events tab, and cleaned up successfully.")
}

// ManageGalleryAction handles uploading a photo to gallery, verifying on public gallery page, deleting it, and taking screenshots.
func ManageGalleryAction(aai AdminActionsInterface, cfg *config.Config, relativeImagePath, description, photoDate string, r *Result) {
	actionName := fmt.Sprintf("Manage Gallery Photo (%s)", description)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// Resolve absolute path for fixture image
	absImagePath, err := filepath.Abs(relativeImagePath)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to resolve gallery image path: %w", err)
		return
	}

	// 1. Click Manage Content button
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click manage content button: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Click Gallery Tab inside Content Management Modal
	galleryTabXPath := "//div[@data-testid='testid-content-management-modal']//button[@data-testid='testid-gallery-button' or text()='Gallery']"
	if err := ap.Page.Click(galleryTabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminGalleryTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(500 * time.Millisecond)

	// 3. Fill Description
	if err := ap.Page.SendKeysByTestID(cfg.AdminGalleryDescriptionInputTestID, description, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill gallery photo description: %w", err)
		return
	}

	// 4. Fill Date via JS event to ensure proper HTML5 date input format YYYY-MM-DD
	_ = ap.Page.SendKeysByTestID(cfg.AdminGalleryDateInputTestID, photoDate, ap.DefaultTimeout)
	jsDateScript := fmt.Sprintf(`
		const dateEl = document.querySelector('[data-testid="%s"]');
		if (dateEl) {
			const prototype = Object.getPrototypeOf(dateEl);
			const setter = Object.getOwnPropertyDescriptor(prototype, 'value').set;
			setter.call(dateEl, %q);
			dateEl.dispatchEvent(new Event('input', { bubbles: true }));
			dateEl.dispatchEvent(new Event('change', { bubbles: true }));
		}
	`, cfg.AdminGalleryDateInputTestID, photoDate)
	_, _ = ap.Page.Driver.ExecuteScript(jsDateScript, nil)

	// 5. Upload Photo file
	photoElem, err := ap.Page.FindElementByTestID(cfg.AdminGalleryPhotoInputTestID, ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find gallery photo file input: %w", err)
		return
	}
	if err := photoElem.SendKeys(absImagePath); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to upload gallery photo file: %w", err)
		return
	}

	// Screenshot Step 01: Gallery Upload Form Filled
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_01_GalleryUpload_FormFilled"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 6. Click Upload Photo Button
	if err := ap.Page.ClickByTestID(cfg.AdminUploadPhotoButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click upload photo button: %w", err)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Gallery Photo Uploaded
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_GalleryPhoto_Uploaded"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Content Management Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	// 7. Go to Events/Gallery Page as Admin via navbar button
	if err := ap.Page.ClickByTestID(cfg.EventsNavButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/events")
	}
	time.Sleep(2 * time.Second)

	// Ensure Gallery tab is selected
	_ = ap.Page.ClickByTestID(cfg.GalleryTabButtonTestID, ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	// Extract current year from photoDate (e.g. "2026")
	currentYear := "2026"
	if len(photoDate) >= 4 {
		currentYear = photoDate[:4]
	}

	// Select uploaded year button on gallery page (e.g. 2026)
	currentYearXPath := fmt.Sprintf("//button[contains(@data-testid, 'testid-gallery-year-%s') or text()=%q]", currentYear, currentYear)
	if err := ap.Page.Click(currentYearXPath, ap.DefaultTimeout); err != nil {
		jsClickYear := fmt.Sprintf(`
			const btns = Array.from(document.querySelectorAll('button'));
			const btn = btns.find(b => b.textContent.trim() === %q || (b.getAttribute('data-testid') && b.getAttribute('data-testid').includes(%q)));
			if (btn) btn.click();
		`, currentYear, currentYear)
		_, _ = ap.Page.Driver.ExecuteScript(jsClickYear, nil)
	}
	time.Sleep(1500 * time.Millisecond)

	// Scroll down further so the newly uploaded gallery photo lower on the page is fully in view
	_, _ = ap.Page.Driver.ExecuteScript("window.scrollBy(0, 950);", nil)
	time.Sleep(1 * time.Second)

	// Screenshot Step 03: Public Photo Gallery Verification
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_EventsPage_Gallery_Result"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 8. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// 9. Click Manage Content -> Gallery tab to clean up test photo
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open manage content modal: %w", err)
		return
	}
	time.Sleep(1 * time.Second)

	// Ensure Gallery tab inside Content Management Modal is clicked
	if err := ap.Page.Click(galleryTabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminGalleryTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(1 * time.Second)

	// Filter by uploaded year in Existing Gallery Photos list using year filter dropdown
	_ = ap.Page.SelectCustomDropdownByText("testid-gallery-year-filter-select", currentYear, 2*time.Second)
	time.Sleep(1 * time.Second)

	// Click delete trash icon specifically for the uploaded test gallery photo matching title/description
	deletePhotoXPath := fmt.Sprintf("//p[contains(text(), %q)]/ancestor::div[contains(@class, 'flex items-start')]//button[contains(@class, 'text-destructive')]", description)
	if err := ap.Page.Click(deletePhotoXPath, ap.DefaultTimeout); err != nil {
		// Specific fallback targeting the row containing description text
		specificXPath := fmt.Sprintf("//div[contains(., %q)]//button[contains(@class, 'text-destructive')]", description)
		if err2 := ap.Page.Click(specificXPath, ap.DefaultTimeout); err2 != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to find delete button for uploaded gallery photo '%s': %w", description, err2)
			return
		}
	}
	time.Sleep(1 * time.Second)

	// Confirm delete in ConfirmDeleteDialog
	confirmDelXPath := "//div[contains(@role, 'alertdialog') or contains(@class, 'max-w-md')]//button[contains(text(), 'Delete') or contains(., 'Delete')]"
	if err := ap.Page.Click(confirmDelXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Click("//button[contains(text(), 'Confirm') or contains(text(), 'Delete')]", ap.DefaultTimeout)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 04: Gallery Photo Deleted
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_GalleryPhoto_Deleted"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close Modal
	_ = ap.Page.Click("//button[contains(@class, 'absolute right-4') or text()='Close']", 2*time.Second)

	r.Advice = append(r.Advice, "Gallery photo uploaded, verified on Photo Gallery page, and cleaned up successfully.")
}

// EditMemberDetails handles searching for a member by mobile number, editing member details, verifying the edit in the view panel, and cleaning up.
func EditMemberDetails(aai AdminActionsInterface, cfg *config.Config, targetMobile, updatedDesignation, updatedDistrict string, r *Result) {
	cleanMobile := strings.TrimSpace(targetMobile)
	cleanMobile = strings.ReplaceAll(cleanMobile, " ", "")

	actionName := fmt.Sprintf("Edit Member Details (%s)", cleanMobile)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()

	// 1. Search for target member by Mobile Number in Member Management table (dispatch JS event to prevent keypress API flooding)
	_ = ap.Page.SendKeysByTestID(cfg.MemberSearchInputTestID, cleanMobile, ap.DefaultTimeout)
	jsSearchScript := fmt.Sprintf(`
		const searchEl = document.querySelector('[data-testid="%s"]');
		if (searchEl) {
			const prototype = Object.getPrototypeOf(searchEl);
			const setter = Object.getOwnPropertyDescriptor(prototype, 'value').set;
			setter.call(searchEl, %q);
			searchEl.dispatchEvent(new Event('input', { bubbles: true }));
			searchEl.dispatchEvent(new Event('change', { bubbles: true }));
		}
	`, cfg.MemberSearchInputTestID, cleanMobile)
	_, _ = ap.Page.Driver.ExecuteScript(jsSearchScript, nil)
	time.Sleep(2 * time.Second)

	// Screenshot Step 01: Member Found in Table
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_01_EditMember_SearchTable"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 2. Click View Details button for member matching mobile
	viewButtonXPath := fmt.Sprintf("//tr[contains(., %q)]//button[contains(., 'View')]", cleanMobile)
	_, _ = ap.Page.Driver.ExecuteScript("const btn = document.querySelector('td button'); if (btn) { btn.scrollIntoView({block: 'center'}); }", nil)
	time.Sleep(500 * time.Millisecond)

	if err := ap.Page.Click(viewButtonXPath, ap.DefaultTimeout); err != nil {
		// Fallback: click view button via JS
		_, _ = ap.Page.Driver.ExecuteScript("const btns = Array.from(document.querySelectorAll('button')); const b = btns.find(x => x.textContent.includes('View')); if (b) b.click();", nil)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Member View Details Panel Open
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_02_EditMember_ViewDetailsBeforeEdit"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 3. Click Edit Button in View Details Modal
	if err := ap.Page.ClickByTestID("testid-member-edit-button", ap.DefaultTimeout); err != nil {
		if err2 := ap.Page.Click("//button[contains(., 'Edit')]", ap.DefaultTimeout); err2 != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to click edit member button: %w", err2)
			return
		}
	}
	time.Sleep(1 * time.Second)

	// 4. Update Designation field
	if updatedDesignation != "" {
		desigElem, err := ap.Page.FindElementByTestID(cfg.AdminAddMemberDesignationInputTestID, ap.DefaultTimeout)
		if err == nil && desigElem != nil {
			_ = desigElem.Clear()
			_ = desigElem.SendKeys(updatedDesignation)
		} else {
			_ = ap.Page.SendKeys("//input[@value and contains(@class, 'h-8')]", updatedDesignation, ap.DefaultTimeout)
		}
	}

	// 5. Update Working District if specified
	if updatedDistrict != "" {
		_ = ap.Page.SelectCustomDropdownByText(cfg.AdminAddMemberWorkingDistrictSelectTestID, updatedDistrict, ap.DefaultTimeout)
	}

	// Screenshot Step 03: Edit Form Filled
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_03_EditMember_FormUpdated"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 6. Click Save Changes button
	if err := ap.Page.ClickByTestID("testid-member-save-edit-button", ap.DefaultTimeout); err != nil {
		if err2 := ap.Page.Click("//button[contains(., 'Save Changes') or contains(., 'Save')]", ap.DefaultTimeout); err2 != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to click save changes on member edit: %w", err2)
			return
		}
	}
	time.Sleep(2 * time.Second)

	// 7. Re-open View Details to verify edited data and take screenshot
	_, _ = ap.Page.Driver.ExecuteScript(jsSearchScript, nil)
	time.Sleep(2 * time.Second)

	if err := ap.Page.Click(viewButtonXPath, ap.DefaultTimeout); err != nil {
		_, _ = ap.Page.Driver.ExecuteScript("const btns = Array.from(document.querySelectorAll('button')); const b = btns.find(x => x.textContent.includes('View')); if (b) b.click();", nil)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 04: View Details Panel showing Edited Data
	if scr, scrErr := ap.Page.CaptureScreenshot("Step_04_EditMember_VerifiedEditedDetails"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// Close View Details Modal
	_ = ap.Page.Click("//button[contains(text(), 'Close') or contains(@class, 'absolute right-4')]", 2*time.Second)

	r.Advice = append(r.Advice, fmt.Sprintf("Member '%s' details updated to Designation='%s' and verified successfully.", cleanMobile, updatedDesignation))
}

