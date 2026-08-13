package pages

import (
	"fmt"
	"strings"
	"time"

	"e2e-template/pkg/config"
	"e2e-template/pkg/ui"
)

// AdminDashboardPage handles actions inside the Admin Panel / Dashboard.
type AdminDashboardPage struct {
	*ui.Page
}

// NewAdminDashboardPage creates a new AdminDashboardPage instance.
func NewAdminDashboardPage(page *ui.Page) *AdminDashboardPage {
	return &AdminDashboardPage{Page: page}
}

// OpenAdminPanel clicks the "Admin Panel" navigation button in the navbar.
func (a *AdminDashboardPage) OpenAdminPanel(adminPanelBtnTestID string, timeout time.Duration) error {
	return a.ClickByTestID(adminPanelBtnTestID, timeout)
}

// OpenAddMemberModal clicks the "Add Member" button on the admin dashboard.
func (a *AdminDashboardPage) OpenAddMemberModal(btnTestID string, timeout time.Duration) error {
	return a.ClickByTestID(btnTestID, timeout)
}

// FillAddMemberForm populates all form fields from cfg without submitting.
func (a *AdminDashboardPage) FillAddMemberForm(cfg *config.Config, timeout time.Duration) error {
	d := cfg.NewMemberFormData

	// Text Inputs
	fields := []struct {
		testID string
		val    string
		label  string
	}{
		{cfg.AdminAddMemberTagaIdInputTestID, d.TagaID, "Taga ID"},
		{cfg.AdminAddMemberNameInputTestID, d.Name, "Name"},
		{cfg.AdminAddMemberInitialInputTestID, d.Initial, "Initial"},
		{cfg.AdminAddMemberFatherNameInputTestID, d.FatherName, "Father Name"},
		{cfg.AdminAddMemberMotherNameInputTestID, d.MotherName, "Mother Name"},
		{cfg.AdminAddMemberEduQualInputTestID, d.EducationalQualification, "Educational Qualification"},
		{cfg.AdminAddMemberDesignationInputTestID, d.Designation, "Designation"},
		{cfg.AdminAddMemberRecruitmentBatchInputTestID, d.RecruitmentBatch, "Recruitment Batch"},
		{cfg.AdminAddMemberSeniorityNumInputTestID, d.SeniorityNumber, "Seniority Number"},
		{cfg.AdminAddMemberDobInputTestID, d.DateOfBirth, "Date of Birth"},
		{cfg.AdminAddMemberMobileInputTestID, d.MobileNumber, "Mobile Number"},
		{cfg.AdminAddMemberEmailInputTestID, d.Email, "Email"},
		{cfg.AdminAddMemberTbfNumInputTestID, d.TbfNumber, "TBF Number"},
		{cfg.AdminAddMemberCpsGpfNumInputTestID, d.CpsGpfNumber, "CPS/GPF Number"},
		{cfg.AdminAddMemberResAddressInputTestID, d.ResidentialAddress, "Residential Address"},
		{cfg.AdminAddMemberPermAddressInputTestID, d.PermanentAddress, "Permanent Address"},
	}

	for _, f := range fields {
		if f.testID != "" && f.val != "" {
			if err := a.SendKeysByTestID(f.testID, f.val, timeout); err != nil {
				return fmt.Errorf("failed to enter %s: %w", f.label, err)
			}
		}
	}

	// Handle Custom Select Dropdowns (Gender, Working District, Native District, Payment Status)
	payStatusTestID := cfg.AdminAddMemberPaymentStatusSelectTestID
	if payStatusTestID == "" {
		payStatusTestID = "testid-add-member-payment-status-select"
	}
	payStatusVal := d.PaymentStatus
	if payStatusVal == "" {
		payStatusVal = "Unpaid"
	}

	dropdowns := []struct {
		triggerTestID string
		optionVal     string
		label         string
	}{
		{cfg.AdminAddMemberGenderSelectTestID, d.Gender, "Gender"},
		{cfg.AdminAddMemberWorkingDistrictSelectTestID, d.WorkingDistrict, "Working District"},
		{cfg.AdminAddMemberNativeDistrictSelectTestID, d.NativeDistrict, "Native District"},
		{payStatusTestID, payStatusVal, "Payment Status"},
	}

	for _, drop := range dropdowns {
		if drop.triggerTestID != "" && drop.optionVal != "" {
			formattedVal := strings.Title(strings.ToLower(drop.optionVal))
			if err := a.SelectCustomDropdownByText(drop.triggerTestID, formattedVal, timeout); err != nil {
				if errRaw := a.SelectCustomDropdownByText(drop.triggerTestID, drop.optionVal, timeout); errRaw != nil {
					return fmt.Errorf("failed to select %s (%s): %w", drop.label, drop.optionVal, err)
				}
			}
		}
	}

	return nil
}

// SubmitAddMemberForm clicks the Add Member submit button.
func (a *AdminDashboardPage) SubmitAddMemberForm(submitTestID string, timeout time.Duration) error {
	return a.ClickByTestID(submitTestID, timeout)
}

// SearchMember types the email into the member search input field.
func (a *AdminDashboardPage) SearchMember(searchInputTestID, email string, timeout time.Duration) error {
	if err := a.SendKeysByTestID(searchInputTestID, email, timeout); err != nil {
		return err
	}
	// Dispatch native JavaScript input event to ensure React updates searchQuery state
	if el, err := a.FindElementByTestID(searchInputTestID, timeout); err == nil {
		_, _ = a.Driver.ExecuteScript("arguments[0].dispatchEvent(new Event('input', { bubbles: true }));", []interface{}{el})
	}
	return nil
}

// ClickFirstTableRowViewButton clicks the "View" button in the member management table for the filtered search.
func (a *AdminDashboardPage) ClickFirstTableRowViewButton(timeout time.Duration) error {
	// Try multiple XPath expressions for the View button in the table row
	xpaths := []string{
		"//table//tbody//tr[1]//button[contains(@data-testid, '-view-button')]",
		"//table//tbody//tr[1]//button[contains(., 'View')]",
		"//table//tbody//tr//button[contains(@data-testid, '-view-button')]",
		"//button[contains(@data-testid, '-view-button')]",
	}

	var lastErr error
	for _, xpath := range xpaths {
		el, err := a.WaitUntilVisible(xpath, 2*time.Second)
		if err == nil {
			// Scroll into view and click via JS to bypass any overlay blocking
			_, _ = a.Driver.ExecuteScript("arguments[0].scrollIntoView({block: 'center'}); arguments[0].click();", []interface{}{el})
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("view button not visible in table row: %w", lastErr)
}

// OpenMemberViewDetails clicks the "View" button for a given member in the management table.
func (a *AdminDashboardPage) OpenMemberViewDetails(memberID string, timeout time.Duration) error {
	viewBtnTestID := fmt.Sprintf("testid-member-%s-view-button", memberID)
	return a.ClickByTestID(viewBtnTestID, timeout)
}

// DeleteMember clicks the Delete button inside the View Details modal and confirms deletion in the dialog.
func (a *AdminDashboardPage) DeleteMember(deleteBtnTestID, confirmDeleteBtnTestID string, timeout time.Duration) error {
	if err := a.ClickByTestID(deleteBtnTestID, timeout); err != nil {
		return fmt.Errorf("failed to click member delete button: %w", err)
	}
	if err := a.ClickByTestID(confirmDeleteBtnTestID, timeout); err != nil {
		return fmt.Errorf("failed to click confirm delete button: %w", err)
	}
	time.Sleep(1 * time.Second)
	if err := a.ClickByTestID("testid-delete-success-ok-button", timeout); err != nil {
		return fmt.Errorf("failed to click delete success OK button: %w", err)
	}
	return nil
}

// OpenBulkUploadModal clicks the "Bulk Member Upload" button.
func (a *AdminDashboardPage) OpenBulkUploadModal(btnTestID string, timeout time.Duration) error {
	return a.ClickByTestID(btnTestID, timeout)
}

// UploadBulkFile sends the absolute file path to the file input element.
func (a *AdminDashboardPage) UploadBulkFile(fileInputTestID, filePath string, timeout time.Duration) error {
	return a.SendKeysByTestID(fileInputTestID, filePath, timeout)
}

// SubmitBulkUpload clicks the bulk upload submit button.
func (a *AdminDashboardPage) SubmitBulkUpload(submitTestID string, timeout time.Duration) error {
	return a.ClickByTestID(submitTestID, timeout)
}

// RefreshMemberTable clicks the table refresh button to reload the latest member records.
func (a *AdminDashboardPage) RefreshMemberTable(refreshBtnTestID string, timeout time.Duration) error {
	return a.ClickByTestID(refreshBtnTestID, timeout)
}

// DeleteMemberByEmail searches for a specific member by email, clicks its specific table row View button, and confirms deletion.
func (a *AdminDashboardPage) DeleteMemberByEmail(searchInputTestID, email, deleteBtnTestID, confirmDeleteBtnTestID string, timeout time.Duration) error {
	// Sanitize email by removing leading, trailing, and internal whitespace
	cleanEmail := strings.TrimSpace(email)
	cleanEmail = strings.ReplaceAll(cleanEmail, " ", "")

	// 1. Clear search input box cleanly
	_ = a.SendKeysByTestID(searchInputTestID, "", timeout)
	time.Sleep(500 * time.Millisecond)

	// 2. Type search email
	if err := a.SearchMember(searchInputTestID, cleanEmail, timeout); err != nil {
		return fmt.Errorf("failed to search email '%s': %w", cleanEmail, err)
	}
	
	// 3. Wait for debounced search filter & React table re-render
	time.Sleep(2 * time.Second)

	// 4. Click View button on the filtered row with retries
	var lastClickErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := a.ClickFirstTableRowViewButton(timeout); err == nil {
			lastClickErr = nil
			break
		} else {
			lastClickErr = err
			time.Sleep(1 * time.Second)
		}
	}
	if lastClickErr != nil {
		return fmt.Errorf("failed to click View button for email '%s': %w", email, lastClickErr)
	}
	time.Sleep(1 * time.Second)

	// 5. Delete & Confirm
	if err := a.DeleteMember(deleteBtnTestID, confirmDeleteBtnTestID, timeout); err != nil {
		return err
	}

	// 6. Wait for deletion API to resolve and refresh table filters
	time.Sleep(2 * time.Second)
	_ = a.RefreshMemberTable("testid-member-refresh-button", timeout)
	time.Sleep(1 * time.Second)
	return nil
}

// DeleteMemberByMobile searches for a specific member by mobile number, clicks its specific table row View button, and confirms deletion.
func (a *AdminDashboardPage) DeleteMemberByMobile(searchInputTestID string, mobile string, deleteBtnTestID string, confirmDeleteBtnTestID string, timeout time.Duration) error {
	// Sanitize mobile by removing leading, trailing, and internal whitespace
	cleanMobile := strings.TrimSpace(mobile)
	cleanMobile = strings.ReplaceAll(cleanMobile, " ", "")

	// 1. Clear search input box cleanly
	_ = a.SendKeysByTestID(searchInputTestID, "", timeout)
	time.Sleep(500 * time.Millisecond)

	// 2. Type search mobile number
	if err := a.SearchMember(searchInputTestID, cleanMobile, timeout); err != nil {
		return fmt.Errorf("failed to search mobile '%s': %w", cleanMobile, err)
	}
	
	// 3. Wait for debounced search filter & React table re-render
	time.Sleep(2 * time.Second)

	// 4. Click View button on the filtered row matching the mobile number exactly
	rowViewBtnXPath := fmt.Sprintf("//table//tbody//tr[contains(., '%s')]//button[contains(@data-testid, '-view-button') or contains(text(), 'View')]", cleanMobile)
	var lastClickErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := a.Click(rowViewBtnXPath, timeout); err == nil {
			lastClickErr = nil
			break
		} else {
			lastClickErr = err
			time.Sleep(1 * time.Second)
		}
	}
	if lastClickErr != nil {
		return fmt.Errorf("failed to click View button for mobile '%s': %w", cleanMobile, lastClickErr)
	}
	time.Sleep(1 * time.Second)

	// 5. Delete & Confirm
	if err := a.DeleteMember(deleteBtnTestID, confirmDeleteBtnTestID, timeout); err != nil {
		return err
	}

	// 6. Wait for deletion API to resolve and refresh table filters
	time.Sleep(2 * time.Second)
	_ = a.RefreshMemberTable("testid-member-refresh-button", timeout)
	time.Sleep(1 * time.Second)
	return nil
}

// DeleteMemberByName searches for a specific member by name, clicks its specific table row View button, and confirms deletion.
func (a *AdminDashboardPage) DeleteMemberByName(searchInputTestID, name, deleteBtnTestID, confirmDeleteBtnTestID string, timeout time.Duration) error {
	// 1. Search name
	if err := a.SearchMember(searchInputTestID, name, timeout); err != nil {
		return fmt.Errorf("failed to search name '%s': %w", name, err)
	}
	time.Sleep(1 * time.Second)

	// 2. Click View button for the row matching this name exactly
	rowViewBtnXPath := fmt.Sprintf("//table//tbody//tr[contains(., '%s')]//button[contains(@data-testid, '-view-button') or contains(text(), 'View')]", name)
	if err := a.Click(rowViewBtnXPath, timeout); err != nil {
		return fmt.Errorf("failed to click View button for name '%s': %w", name, err)
	}
	time.Sleep(1 * time.Second)

	// 3. Delete & Confirm
	return a.DeleteMember(deleteBtnTestID, confirmDeleteBtnTestID, timeout)
}
