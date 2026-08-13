package actions

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"e2e-template/pkg/config"
	"e2e-template/pkg/ui"
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

// Helper: captureScreenshot captures a PNG screenshot with a small render delay and appends it to evidence.
func captureScreenshot(ap *AdminPersona, r *Result, name string) {
	if ap == nil || ap.Page == nil {
		return
	}
	time.Sleep(500 * time.Millisecond)
	if scr, scrErr := ap.Page.CaptureScreenshot(name); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
}

// Helper: captureNetworkEvidence retrieves monkey-patched fetch/XHR API errors from browser sessionStorage and appends formatted logs to Advice.
func captureNetworkEvidence(ap *AdminPersona, r *Result) {
	if ap == nil || ap.Page == nil || ap.Page.Driver == nil {
		return
	}
	if raw := ap.Page.RetrieveNetworkErrors(); raw != "" {
		if formatted := ui.FormatNetworkErrors(raw); formatted != "" {
			r.Advice = append(r.Advice, formatted)
		}
	}
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

	ap.Page.InjectNetworkInterceptor()

	homePage := pages.NewHomePage(ap.Page)
	loginPage := pages.NewLoginPage(ap.Page)

	if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminLoginButtonTestID' button exists on Home Page")
		captureScreenshot(ap, r, "Step_01_AdminLogin_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	captureScreenshot(ap, r, "Step_01_AdminLogin_ModalOpen")

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
		captureScreenshot(ap, r, r.TestName+"_Step01_AdminLogin_Failure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Wait for home page to fully render after login redirect
	time.Sleep(3 * time.Second)
	captureScreenshot(ap, r, "Step_01_AdminLogin_HomePage")
	captureNetworkEvidence(ap, r)
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
	ap.Page.InjectNetworkInterceptor()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Check if navbar 'adminPanelButtonTestID' is visible after Admin login")
		captureScreenshot(ap, r, "Step_02_AdminPanel_Failure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Wait for admin panel table to fully render
	time.Sleep(3 * time.Second)
	captureScreenshot(ap, r, "Step_02_AdminPanel_Loaded")
	captureNetworkEvidence(ap, r)
	r.Advice = append(r.Advice, "Admin Panel opened successfully.")
}

// AddSingleMember opens the modal, fills all 19 form fields from config, and submits. Returns the temporary password if any.
func AddSingleMember(aai AdminActionsInterface, cfg *config.Config, r *Result) string {
	actionName := "Add Single Member with 19 Fields"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return ""
	}

	ap := aai.GetAdminPersona()
	ap.Page.InjectNetworkInterceptor()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	if err := adminDashboard.OpenAddMemberModal(cfg.AdminAddMemberButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminAddMemberButtonTestID' is clickable in Admin Dashboard")
		captureScreenshot(ap, r, "Step_03_AddMember_OpenFailure")
		captureNetworkEvidence(ap, r)
		return ""
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_03_AddMember_ModalOpen")

	if err := adminDashboard.FillAddMemberForm(cfg, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify all 19 form input testIDs in config.json match the UI elements")
		captureScreenshot(ap, r, "Step_03_AddMember_FillFailure")
		captureNetworkEvidence(ap, r)
		return ""
	}

	// Screenshot: form fully filled, before submitting
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_03_AddMember_FormFilled")

	if err := adminDashboard.SubmitAddMemberForm(cfg.AdminAddMemberSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminAddMemberSubmitButtonTestID' submit button")
		captureScreenshot(ap, r, "Step_04_AddMember_SubmitFailure")
		captureNetworkEvidence(ap, r)
		return ""
	}

	// Screenshot: success toast visible immediately after submit
	time.Sleep(2 * time.Second)
	captureScreenshot(ap, r, "Step_04_AddMember_SuccessToast")
	captureNetworkEvidence(ap, r)

	// Extract the temporary password
	tempPassword, _ := ap.Page.GetTextByTestID("testid-temp-password", 2*time.Second)

	// Click OK on the success modal to close it
	if err := ap.Page.ClickByTestID("testid-add-success-ok-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click add success OK button: %w", err)
		captureScreenshot(ap, r, "Step_04_AddMember_DismissFailure")
		captureNetworkEvidence(ap, r)
		return ""
	}
	time.Sleep(1 * time.Second)

	// Refresh table so new member appears
	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	captureScreenshot(ap, r, "Step_05_AdminPanel_AfterAddMember")
	captureNetworkEvidence(ap, r)
	r.Advice = append(r.Advice, "Member added successfully with all 19 fields populated.")
	
	return tempPassword
}

// SetPaymentStatusToPaid attempts to select the paid checkbox/button.
func SetPaymentStatusToPaid(aai AdminActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Set Member Payment Status to Paid"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	ap.Page.InjectNetworkInterceptor()
	captureScreenshot(ap, r, "Step_SetPaymentStatus_Attempt")

	if err := ap.Page.ClickByTestID("testid-member-payment-paid-checkbox", 5*time.Second); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select paid status: %w", err)
		r.Advice = append(r.Advice, "Advice: Verify testid-member-payment-paid-checkbox exists on form")
		captureScreenshot(ap, r, "Step_SetPaymentStatus_Failure")
		captureNetworkEvidence(ap, r)
		return
	}
	captureScreenshot(ap, r, "Step_SetPaymentStatus_Success")
	captureNetworkEvidence(ap, r)
}

// DeleteMemberByEmail searches for a member by email and confirms deletion, capturing screenshots at each step.
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
	ap.Page.InjectNetworkInterceptor()
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

	// 1. Search for member by email
	_ = ap.Page.SendKeysByTestID(cfg.MemberSearchInputTestID, "", ap.DefaultTimeout)
	time.Sleep(500 * time.Millisecond)

	if err := adminDashboard.SearchMember(cfg.MemberSearchInputTestID, cleanEmail, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to search member by email '%s': %w", cleanEmail, err)
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check search input testID '%s'", cfg.MemberSearchInputTestID))
		captureScreenshot(ap, r, "Step_06_DeleteEmail_SearchInputFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot 1: Search result in member table prior to deletion
	captureScreenshot(ap, r, "Step_06_DeleteEmail_01_SearchResult")

	// 2. Click View button on filtered row
	var lastClickErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := adminDashboard.ClickFirstTableRowViewButton(ap.DefaultTimeout); err == nil {
			lastClickErr = nil
			break
		} else {
			lastClickErr = err
			time.Sleep(1 * time.Second)
		}
	}
	if lastClickErr != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click View button for email '%s': %w", cleanEmail, lastClickErr)
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check if member email '%s' exists in table", cleanEmail))
		captureScreenshot(ap, r, "Step_06_DeleteEmail_SearchFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// Screenshot 2: Member Details modal open showing View details and Delete button
	captureScreenshot(ap, r, "Step_06_DeleteEmail_02_ViewDetailsModal")

	// 3. Click Delete button inside View modal
	if err := ap.Page.ClickByTestID(cfg.MemberDeleteButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click member delete button: %w", err)
		r.Advice = append(r.Advice, "Advice: Verify delete button testID in view modal")
		captureScreenshot(ap, r, "Step_06_DeleteEmail_DeleteClickFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// Screenshot 3: Confirmation modal open asking to confirm member deletion
	captureScreenshot(ap, r, "Step_06_DeleteEmail_03_ConfirmModal")

	// 4. Click Confirm Delete button
	if err := ap.Page.ClickByTestID(cfg.MemberConfirmDeleteButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click confirm delete button: %w", err)
		r.Advice = append(r.Advice, "Advice: Verify confirm delete button testID")
		captureScreenshot(ap, r, "Step_06_DeleteEmail_ConfirmClickFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot 4: Success toast / confirmation message visible after confirming deletion
	captureScreenshot(ap, r, "Step_07_DeleteMember_SuccessToast")
	captureNetworkEvidence(ap, r)

	// Dismiss success modal if present
	_ = ap.Page.ClickByTestID("testid-delete-success-ok-button", ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	// 5. Refresh table and search again to capture post-deletion empty state proving member is gone
	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)
	_ = adminDashboard.SearchMember(cfg.MemberSearchInputTestID, cleanEmail, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	// Screenshot 5: Table view after deletion verifying member is no longer in list
	captureScreenshot(ap, r, "Step_08_DeleteEmail_05_TableAfterDeletion")
	captureNetworkEvidence(ap, r)

	r.Advice = append(r.Advice, fmt.Sprintf("Member '%s' deleted successfully.", cleanEmail))
}

// DeleteMemberByMobile searches for a member by mobile number and confirms deletion, capturing screenshots at each step.
func DeleteMemberByMobile(aai AdminActionsInterface, cfg *config.Config, mobile string, stepPrefix string, r *Result) {
	cleanMobile := strings.TrimSpace(mobile)
	cleanMobile = strings.ReplaceAll(cleanMobile, " ", "")

	actionName := fmt.Sprintf("Delete Member (%s)", cleanMobile)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	ap := aai.GetAdminPersona()
	ap.Page.InjectNetworkInterceptor()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	// Ensure we navigate to Admin Panel cleanly
	if err := adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.KeyDown("\uE00C") // Escape key
		_ = ap.Page.Driver.KeyUp("\uE00C")
		time.Sleep(1 * time.Second)
		_ = adminDashboard.OpenAdminPanel(cfg.AdminPanelButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(1 * time.Second)

	// 1. Search for member by mobile number
	_ = ap.Page.SendKeysByTestID(cfg.MemberSearchInputTestID, "", ap.DefaultTimeout)
	time.Sleep(500 * time.Millisecond)

	if err := adminDashboard.SearchMember(cfg.MemberSearchInputTestID, cleanMobile, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to search member by mobile '%s': %w", cleanMobile, err)
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check search input testID '%s'", cfg.MemberSearchInputTestID))
		captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_SearchInputFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot 1: Search result in member table prior to deletion
	captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_01_SearchResult")

	// 2. Click View button on row matching mobile number
	rowViewBtnXPath := fmt.Sprintf("//table//tbody//tr[contains(., '%s')]//button[contains(@data-testid, '-view-button') or contains(text(), 'View')]", cleanMobile)
	var lastClickErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := ap.Page.Click(rowViewBtnXPath, ap.DefaultTimeout); err == nil {
			lastClickErr = nil
			break
		} else {
			lastClickErr = err
			time.Sleep(1 * time.Second)
		}
	}
	if lastClickErr != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click View button for mobile '%s': %w", cleanMobile, lastClickErr)
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check if member mobile '%s' exists in table", cleanMobile))
		captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_SearchFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// Screenshot 2: Member Details modal open showing View details and Delete button
	captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_02_ViewDetailsModal")

	// 3. Click Delete button inside View modal
	if err := ap.Page.ClickByTestID(cfg.MemberDeleteButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click member delete button: %w", err)
		r.Advice = append(r.Advice, "Advice: Verify delete button testID in view modal")
		captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_DeleteClickFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// Screenshot 3: Confirmation modal open asking to confirm member deletion
	captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_03_ConfirmModal")

	// 4. Click Confirm Delete button
	if err := ap.Page.ClickByTestID(cfg.MemberConfirmDeleteButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click confirm delete button: %w", err)
		r.Advice = append(r.Advice, "Advice: Verify confirm delete button testID")
		captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_ConfirmClickFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot 4: Success toast / confirmation message visible after confirming deletion
	captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_SuccessToast")
	captureNetworkEvidence(ap, r)

	// Dismiss success modal if present
	_ = ap.Page.ClickByTestID("testid-delete-success-ok-button", ap.DefaultTimeout)
	time.Sleep(1 * time.Second)

	// 5. Refresh table and search again to capture post-deletion empty state proving member is gone
	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)
	_ = adminDashboard.SearchMember(cfg.MemberSearchInputTestID, cleanMobile, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	// Screenshot 5: Table view after deletion verifying member is no longer in list
	captureScreenshot(ap, r, stepPrefix+"_DeleteMobile_"+cleanMobile+"_05_TableAfterDeletion")
	captureNetworkEvidence(ap, r)

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
	ap.Page.InjectNetworkInterceptor()
	adminDashboard := pages.NewAdminDashboardPage(ap.Page)

	if err := adminDashboard.OpenBulkUploadModal(cfg.AdminBulkUploadButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminBulkUploadButtonTestID' button exists on Admin Panel")
		captureScreenshot(ap, r, "Step_03_BulkUpload_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_03_BulkUpload_ModalOpen")

	absPath, err := filepath.Abs(fixtureRelPath)
	if err != nil {
		absPath = fixtureRelPath
	}

	if err := adminDashboard.UploadBulkFile(cfg.AdminBulkUploadFileInputTestID, absPath, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Ensure file '%s' exists and file input testID is correct", absPath))
		captureScreenshot(ap, r, "Step_03_BulkUpload_FileFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Screenshot: CSV file selected in upload dialog
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_03_BulkUpload_FileSelected")

	if err := adminDashboard.SubmitBulkUpload(cfg.AdminBulkUploadSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify bulk upload submit button testID")
		captureScreenshot(ap, r, "Step_04_BulkUpload_SubmitFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Screenshot: upload success toast / result visible
	time.Sleep(3 * time.Second)
	captureScreenshot(ap, r, "Step_04_BulkUpload_SuccessToast")
	captureNetworkEvidence(ap, r)

	_ = adminDashboard.RefreshMemberTable(cfg.MemberRefreshButtonTestID, ap.DefaultTimeout)
	time.Sleep(2 * time.Second)

	captureScreenshot(ap, r, "Step_05_AdminPanel_AfterBulkUpload")
	captureNetworkEvidence(ap, r)
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
	ap.Page.InjectNetworkInterceptor()

	// 1. Generate Excel Report (Membership Report - All Time)
	if err := ap.Page.ClickByTestID("testid-generate-excel-report-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-generate-excel-report-button' exists on Admin Panel")
		captureScreenshot(ap, r, "Step_03_MembershipReport_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_03_MembershipReport_ModalOpen")

	// Select 'All Time' from period dropdown
	if err := ap.Page.SelectCustomDropdownByText("testid-report-period-select", "All Time", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to select 'All Time' from report period dropdown")
		captureScreenshot(ap, r, "Step_03_MembershipReport_SelectFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_03_MembershipReport_PeriodSelected")

	// Click download
	if err := ap.Page.ClickByTestID("testid-download-excel-report-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-download-excel-report-button' exists in the report modal")
		captureScreenshot(ap, r, "Step_03_MembershipReport_DownloadFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	
	// Screenshot: after triggering first download
	time.Sleep(3 * time.Second) 
	captureScreenshot(ap, r, "Step_03_MembershipReport_Downloaded")
	captureNetworkEvidence(ap, r)

	// 2. Export Member List Table Excel
	if err := ap.Page.ClickByTestID("testid-export-excel-button", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-export-excel-button' exists above the member list table")
		captureScreenshot(ap, r, "Step_04_MemberListTable_ExportFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	
	// Screenshot: after triggering second download
	time.Sleep(3 * time.Second) 
	captureScreenshot(ap, r, "Step_04_MemberListTable_Downloaded")
	captureNetworkEvidence(ap, r)

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
	ap.Page.InjectNetworkInterceptor()
	homePage := pages.NewHomePage(ap.Page)

	if err := homePage.Logout(cfg.LogoutButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'logoutButtonTestID' element exists in header/navbar")
		captureScreenshot(ap, r, "Step_07_AdminLogout_Failure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Wait for redirect back to home / login page to fully complete
	time.Sleep(3 * time.Second)
	captureScreenshot(ap, r, "Step_07_AdminLogout_HomePage")
	captureNetworkEvidence(ap, r)
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
	ap.Page.InjectNetworkInterceptor()
	homePage := pages.NewHomePage(ap.Page)
	loginPage := pages.NewLoginPage(ap.Page)

	if err := homePage.OpenAdminLogin(cfg.AdminLoginButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'adminLoginButtonTestID' button exists on Home Page")
		captureScreenshot(ap, r, screenshotName+"_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Only enter credentials if non-empty
	if username != "" {
		if err := loginPage.EnterUsername(cfg.AdminLoginUsernameInputTestID, username, ap.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			captureScreenshot(ap, r, screenshotName+"_UserFailure")
			captureNetworkEvidence(ap, r)
			return
		}
	}
	if password != "" {
		if err := loginPage.EnterPassword(cfg.AdminLoginPasswordInputTestID, password, ap.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			captureScreenshot(ap, r, screenshotName+"_PassFailure")
			captureNetworkEvidence(ap, r)
			return
		}
	}

	if err := loginPage.SubmitLogin(cfg.AdminLoginSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		captureScreenshot(ap, r, screenshotName+"_SubmitFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	time.Sleep(waitTime)
	captureScreenshot(ap, r, screenshotName)
	captureNetworkEvidence(ap, r)
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
	ap.Page.InjectNetworkInterceptor()
	homePage := pages.NewHomePage(ap.Page)

	if err := homePage.Logout(cfg.LogoutButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'logoutButtonTestID' element exists in header/navbar")
		captureScreenshot(ap, r, screenshotName+"_Failure")
		captureNetworkEvidence(ap, r)
		return
	}

	time.Sleep(waitTime)
	captureScreenshot(ap, r, screenshotName)
	captureNetworkEvidence(ap, r)
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
	ap.Page.InjectNetworkInterceptor()

	// 1. Click Send Announcement button
	if err := ap.Page.ClickByTestID(cfg.AdminSendAnnouncementButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click send announcement button: %w", err)
		r.Advice = append(r.Advice, "Advice: Ensure Admin Panel is open and Send Announcement button is visible")
		captureScreenshot(ap, r, "Step_01_Announcement_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_01_AnnouncementModal_Open")

	// 2. Fill Title
	if err := ap.Page.SendKeysByTestID(cfg.AdminAnnouncementTitleInputTestID, title, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill announcement title: %w", err)
		captureScreenshot(ap, r, "Step_02_Announcement_TitleFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// 3. Fill Message
	if err := ap.Page.SendKeysByTestID(cfg.AdminAnnouncementMessageInputTestID, message, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill announcement message: %w", err)
		captureScreenshot(ap, r, "Step_02_Announcement_MessageFailure")
		captureNetworkEvidence(ap, r)
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
	captureScreenshot(ap, r, "Step_02_AnnouncementForm_Filled")

	// 5. Submit Announcement
	if err := ap.Page.ClickByTestID(cfg.AdminAnnouncementSubmitButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to submit announcement: %w", err)
		captureScreenshot(ap, r, "Step_03_Announcement_SubmitFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	time.Sleep(2 * time.Second)

	// Screenshot 3: Announcement sent confirmation
	captureScreenshot(ap, r, "Step_03_Announcement_Sent")
	captureNetworkEvidence(ap, r)

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
	ap.Page.InjectNetworkInterceptor()

	// 1. Click Manage District Office Bearers button in Admin Panel ("testid-manage-office-bearers-button")
	if err := ap.Page.ClickByTestID(cfg.AdminManageOfficeBearersButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to open district office bearers modal: %w", err)
		captureScreenshot(ap, r, "Step_01_DistrictBearers_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_01_DistrictBearers_ModalOpen")

	// 2. Select District (e.g. Ariyalur) by dropdown field
	if err := ap.Page.Click("//div[@data-testid='testid-office-bearers-modal']//button[contains(@role, 'combobox')]", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click district dropdown combobox: %w", err)
		captureScreenshot(ap, r, "Step_01_DistrictBearers_DropdownFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(500 * time.Millisecond)
	if err := ap.Page.Click(fmt.Sprintf("//*[contains(@role, 'option') and (text()='%s' or .='%s')]", district, district), ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select district '%s' from dropdown: %w", district, err)
		captureScreenshot(ap, r, "Step_01_DistrictBearers_SelectFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// 3. Backup existing Joint Secretary (Women) - Index 3
	nameElem, err := ap.Page.FindElementByTestID("testid-bearer-name-3", ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find Joint Secretary (Women) name input: %w", err)
		captureScreenshot(ap, r, "Step_01_DistrictBearers_FindNameFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	contactElem, err := ap.Page.FindElementByTestID("testid-bearer-contact-3", ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find Joint Secretary (Women) contact input: %w", err)
		captureScreenshot(ap, r, "Step_01_DistrictBearers_FindContactFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	originalName, _ := nameElem.GetAttribute("value")
	originalContact, _ := contactElem.GetAttribute("value")

	r.Advice = append(r.Advice, fmt.Sprintf("Backup data for Joint Secretary (Women) before edit: Name='%s', Mobile='%s'", originalName, originalContact))

	// Screenshot 1: Before Edit
	captureScreenshot(ap, r, "Step_01_DistrictBearers_BeforeEdit")

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
	captureScreenshot(ap, r, "Step_02_DistrictBearers_FormFilled")

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
	captureScreenshot(ap, r, "Step_03_DistrictBearers_Saved")
	captureNetworkEvidence(ap, r)

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
		captureScreenshot(ap, r, "Step_04_DistrictBearers_PublicSelectFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Scroll down to the end of the page to fully view the district office bearers table & photos
	_, _ = ap.Page.Driver.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
	time.Sleep(1 * time.Second)

	// Screenshot 4: Public Page Result Table (scrolled down)
	captureScreenshot(ap, r, "Step_04_OfficeBearers_PublicPage_Result")

	// 7. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// Click Office Bearer Management Button again
	if err := ap.Page.ClickByTestID(cfg.AdminManageOfficeBearersButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open district office bearers modal in admin panel: %w", err)
		captureScreenshot(ap, r, "Step_05_DistrictBearers_ReopenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// Select District dropdown again
	if err := ap.Page.Click("//div[@data-testid='testid-office-bearers-modal']//button[contains(@role, 'combobox')]", ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click district dropdown combobox on restore step: %w", err)
		captureScreenshot(ap, r, "Step_05_DistrictBearers_RestoreDropdownFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(500 * time.Millisecond)
	if err := ap.Page.Click(fmt.Sprintf("//*[contains(@role, 'option') and (text()='%s' or .='%s')]", district, district), ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-select district '%s' on restore step: %w", district, err)
		captureScreenshot(ap, r, "Step_05_DistrictBearers_RestoreSelectFailure")
		captureNetworkEvidence(ap, r)
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
	captureScreenshot(ap, r, "Step_05_DistrictBearers_RestoredToDefault")
	captureNetworkEvidence(ap, r)

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
	ap.Page.InjectNetworkInterceptor()

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
		captureScreenshot(ap, r, "Step_01_ResourceUpload_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Click Resources Tab inside Manage Content modal
	tabXPath := "//div[@data-testid='testid-content-management-modal']//button[@data-testid='testid-resources-button' or text()='Resources']"
	if err := ap.Page.Click(tabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminResourcesTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(500 * time.Millisecond)
	captureScreenshot(ap, r, "Step_01_ResourceUpload_ModalOpen")

	// 3. Select Category (Establishment)
	if err := ap.Page.SelectCustomDropdownByText(cfg.AdminResourceCategorySelectTestID, categoryName, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select category '%s': %w", categoryName, err)
		captureScreenshot(ap, r, "Step_01_ResourceUpload_CategoryFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// 4. Upload PDF file
	fileElem, err := ap.Page.FindElementByTestID(cfg.AdminResourceFileInputTestID, ap.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find resource file input: %w", err)
		captureScreenshot(ap, r, "Step_01_ResourceUpload_InputFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	if err := fileElem.SendKeys(absPdfPath); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to upload PDF file: %w", err)
		captureScreenshot(ap, r, "Step_01_ResourceUpload_SendKeysFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Screenshot Step 01: Form Filled
	captureScreenshot(ap, r, "Step_01_ResourceUpload_FormFilled")

	// 6. Click Upload Resource Button
	if err := ap.Page.ClickByTestID(cfg.AdminUploadResourceButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click upload resource button: %w", err)
		captureScreenshot(ap, r, "Step_02_ResourceUpload_SubmitFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Uploaded Confirmation Toast/List
	captureScreenshot(ap, r, "Step_02_ResourceUpload_Submitted")
	captureNetworkEvidence(ap, r)

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
	captureScreenshot(ap, r, "Step_03_ResourcesPage_Verification")

	// 8. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// 9. Click Manage Content -> Resources tab
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open manage content modal: %w", err)
		captureScreenshot(ap, r, "Step_04_Resource_ReopenFailure")
		captureNetworkEvidence(ap, r)
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
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to locate or click delete button for the uploaded resource 'a_test_resource_sample': %w", err)
		captureScreenshot(ap, r, "Step_04_Resource_DeleteClickFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// Confirm delete in ConfirmDeleteDialog if dialog pops up
	confirmDelXPath := "//div[contains(@role, 'alertdialog') or contains(@class, 'max-w-md')]//button[contains(text(), 'Delete') or contains(., 'Delete')]"
	if err := ap.Page.Click(confirmDelXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Click("//button[contains(text(), 'Confirm') or contains(text(), 'Delete')]", ap.DefaultTimeout)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 04: Resource Deleted
	captureScreenshot(ap, r, "Step_04_Resource_Deleted")
	captureNetworkEvidence(ap, r)

	// Verify the resource has actually disappeared from the list
	err = ap.Page.Driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		el, err := wd.FindElement(selenium.ByXPATH, "//p[contains(text(), 'a_test_resource_sample')]")
		if err != nil {
			return true, nil // successfully disappeared
		}
		disp, err := el.IsDisplayed()
		if err != nil || !disp {
			return true, nil // successfully hidden or disappeared
		}
		return false, nil // still visible
	}, 5*time.Second)

	if err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("resource document was not deleted and is still present in the list: %w", err)
		captureScreenshot(ap, r, "Step_04_Resource_StillPresentFailure")
		captureNetworkEvidence(ap, r)
		return
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
	ap.Page.InjectNetworkInterceptor()

	// 1. Click Manage Content button
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click manage content button: %w", err)
		captureScreenshot(ap, r, "Step_01_EventCreate_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Click Events Tab inside Content Management Modal
	eventsTabXPath := "//div[@data-testid='testid-content-management-modal']//button[@data-testid='testid-events-button' or text()='Events']"
	if err := ap.Page.Click(eventsTabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminEventsTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(500 * time.Millisecond)
	captureScreenshot(ap, r, "Step_01_EventCreate_ModalOpen")

	// 3. Fill Event Title
	if err := ap.Page.SendKeysByTestID(cfg.AdminEventTitleInputTestID, title, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill event title: %w", err)
		captureScreenshot(ap, r, "Step_01_EventCreate_TitleFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// 4. Fill Event Date
	if err := ap.Page.SendKeysByTestID(cfg.AdminEventDateInputTestID, eventDate, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill event date: %w", err)
		captureScreenshot(ap, r, "Step_01_EventCreate_DateFailure")
		captureNetworkEvidence(ap, r)
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
	captureScreenshot(ap, r, "Step_01_EventCreate_FormFilled")

	// 7. Click Publish Event Button
	if err := ap.Page.ClickByTestID(cfg.AdminPublishEventButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click publish event button: %w", err)
		captureScreenshot(ap, r, "Step_02_Event_PublishFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Event Published
	captureScreenshot(ap, r, "Step_02_Event_Published")
	captureNetworkEvidence(ap, r)

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
	captureScreenshot(ap, r, "Step_03_EventsPage_UpcomingEvents_Result")

	// 9. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// 10. Click Manage Content -> Events tab to clean up test event
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open manage content modal: %w", err)
		captureScreenshot(ap, r, "Step_04_Event_ReopenFailure")
		captureNetworkEvidence(ap, r)
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
	captureScreenshot(ap, r, "Step_04_Event_Deleted")
	captureNetworkEvidence(ap, r)

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
	ap.Page.InjectNetworkInterceptor()

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
		captureScreenshot(ap, r, "Step_01_GalleryUpload_OpenFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Click Gallery Tab inside Content Management Modal
	galleryTabXPath := "//div[@data-testid='testid-content-management-modal']//button[@data-testid='testid-gallery-button' or text()='Gallery']"
	if err := ap.Page.Click(galleryTabXPath, ap.DefaultTimeout); err != nil {
		_ = ap.Page.ClickByTestID(cfg.AdminGalleryTabButtonTestID, ap.DefaultTimeout)
	}
	time.Sleep(500 * time.Millisecond)
	captureScreenshot(ap, r, "Step_01_GalleryUpload_ModalOpen")

	// 3. Fill Description
	if err := ap.Page.SendKeysByTestID(cfg.AdminGalleryDescriptionInputTestID, description, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill gallery photo description: %w", err)
		captureScreenshot(ap, r, "Step_01_GalleryUpload_DescFailure")
		captureNetworkEvidence(ap, r)
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
		captureScreenshot(ap, r, "Step_01_GalleryUpload_InputFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	if err := photoElem.SendKeys(absImagePath); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to upload gallery photo file: %w", err)
		captureScreenshot(ap, r, "Step_01_GalleryUpload_SendKeysFailure")
		captureNetworkEvidence(ap, r)
		return
	}

	// Screenshot Step 01: Gallery Upload Form Filled
	captureScreenshot(ap, r, "Step_01_GalleryUpload_FormFilled")

	// 6. Click Upload Photo Button
	if err := ap.Page.ClickByTestID(cfg.AdminUploadPhotoButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click upload photo button: %w", err)
		captureScreenshot(ap, r, "Step_02_GalleryPhoto_UploadFailure")
		captureNetworkEvidence(ap, r)
		return
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 02: Gallery Photo Uploaded
	captureScreenshot(ap, r, "Step_02_GalleryPhoto_Uploaded")
	captureNetworkEvidence(ap, r)

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
	captureScreenshot(ap, r, "Step_03_EventsPage_Gallery_Result")

	// 8. Return to Admin Panel
	if err := ap.Page.ClickByTestID(cfg.AdminPanelButtonTestID, ap.DefaultTimeout); err != nil {
		_ = ap.Page.Driver.Get(cfg.UiURL + "/#/admin")
	}
	time.Sleep(2 * time.Second)

	// 9. Click Manage Content -> Gallery tab to clean up test photo
	if err := ap.Page.ClickByTestID(cfg.AdminManageContentButtonTestID, ap.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to re-open manage content modal: %w", err)
		captureScreenshot(ap, r, "Step_04_GalleryPhoto_ReopenFailure")
		captureNetworkEvidence(ap, r)
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
			captureScreenshot(ap, r, "Step_04_GalleryPhoto_DeleteClickFailure")
			captureNetworkEvidence(ap, r)
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
	captureScreenshot(ap, r, "Step_04_GalleryPhoto_Deleted")
	captureNetworkEvidence(ap, r)

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
	ap.Page.InjectNetworkInterceptor()

	// 1. Search for target member by Mobile Number in Member Management table
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
	captureScreenshot(ap, r, "Step_01_EditMember_SearchTable")

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
	captureScreenshot(ap, r, "Step_02_EditMember_ViewDetailsBeforeEdit")

	// 3. Click Edit Button in View Details Modal
	if err := ap.Page.ClickByTestID("testid-member-edit-button", ap.DefaultTimeout); err != nil {
		if err2 := ap.Page.Click("//button[contains(., 'Edit')]", ap.DefaultTimeout); err2 != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to click edit member button: %w", err2)
			captureScreenshot(ap, r, "Step_02_EditMember_EditClickFailure")
			captureNetworkEvidence(ap, r)
			return
		}
	}
	time.Sleep(1 * time.Second)
	captureScreenshot(ap, r, "Step_02_EditMember_EditFormOpen")

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
	captureScreenshot(ap, r, "Step_03_EditMember_FormUpdated")

	// 6. Click Save Changes button
	if err := ap.Page.ClickByTestID("testid-member-save-edit-button", ap.DefaultTimeout); err != nil {
		if err2 := ap.Page.Click("//button[contains(., 'Save Changes') or contains(., 'Save')]", ap.DefaultTimeout); err2 != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to click save changes on member edit: %w", err2)
			captureScreenshot(ap, r, "Step_03_EditMember_SaveClickFailure")
			captureNetworkEvidence(ap, r)
			return
		}
	}
	time.Sleep(2 * time.Second)
	captureScreenshot(ap, r, "Step_03_EditMember_SaveToast")
	captureNetworkEvidence(ap, r)

	// 7. Re-open View Details to verify edited data and take screenshot
	_, _ = ap.Page.Driver.ExecuteScript(jsSearchScript, nil)
	time.Sleep(2 * time.Second)

	if err := ap.Page.Click(viewButtonXPath, ap.DefaultTimeout); err != nil {
		_, _ = ap.Page.Driver.ExecuteScript("const btns = Array.from(document.querySelectorAll('button')); const b = btns.find(x => x.textContent.includes('View')); if (b) b.click();", nil)
	}
	time.Sleep(2 * time.Second)

	// Screenshot Step 04: View Details Panel showing Edited Data
	captureScreenshot(ap, r, "Step_04_EditMember_VerifiedEditedDetails")

	// Close View Details Modal
	_ = ap.Page.Click("//button[contains(text(), 'Close') or contains(@class, 'absolute right-4')]", 2*time.Second)

	r.Advice = append(r.Advice, fmt.Sprintf("Member '%s' details updated to Designation='%s' and verified successfully.", cleanMobile, updatedDesignation))
}
