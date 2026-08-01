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

	// Handle Custom Select Dropdowns (Gender, Working District, Native District)
	dropdowns := []struct {
		triggerTestID string
		optionVal     string
		label         string
	}{
		{cfg.AdminAddMemberGenderSelectTestID, d.Gender, "Gender"},
		{cfg.AdminAddMemberWorkingDistrictSelectTestID, d.WorkingDistrict, "Working District"},
		{cfg.AdminAddMemberNativeDistrictSelectTestID, d.NativeDistrict, "Native District"},
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
	return a.SendKeysByTestID(searchInputTestID, email, timeout)
}

// ClickFirstTableRowViewButton clicks the "View" button in the first row of the member management table.
func (a *AdminDashboardPage) ClickFirstTableRowViewButton(timeout time.Duration) error {
	viewBtnXPath := "//table//tbody//tr[1]//button[contains(@data-testid, '-view-button') or contains(text(), 'View')]"
	return a.Click(viewBtnXPath, timeout)
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

// DeleteMemberByEmail searches for a specific member by email, clicks its specific table row View button, and confirms deletion.
func (a *AdminDashboardPage) DeleteMemberByEmail(searchInputTestID, email, deleteBtnTestID, confirmDeleteBtnTestID string, timeout time.Duration) error {
	// 1. Search email
	if err := a.SearchMember(searchInputTestID, email, timeout); err != nil {
		return fmt.Errorf("failed to search email '%s': %w", email, err)
	}
	time.Sleep(1 * time.Second)

	// 2. Click View button for the row matching this email exactly
	rowViewBtnXPath := fmt.Sprintf("//table//tbody//tr[contains(., '%s')]//button[contains(@data-testid, '-view-button') or contains(text(), 'View')]", email)
	if err := a.Click(rowViewBtnXPath, timeout); err != nil {
		return fmt.Errorf("failed to click View button for email '%s': %w", email, err)
	}
	time.Sleep(1 * time.Second)

	// 3. Delete & Confirm
	return a.DeleteMember(deleteBtnTestID, confirmDeleteBtnTestID, timeout)
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
