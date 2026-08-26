package actions

import (
	"math/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"e2e-template/pkg/config"
	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/pages"
)

// MemberActionsInterface allows MemberPersona to execute member-specific actions.
type MemberActionsInterface interface {
	PublicActionsInterface
	GetMemberPersona() *MemberPersona
}

// GetMemberPersona implements MemberActionsInterface for MemberPersona.
func (p *MemberPersona) GetMemberPersona() *MemberPersona {
	return p
}

// MemberLoginAttempt attempts member login with the given username, password, and custom timeout/screenshot.
func MemberLoginAttempt(mai MemberActionsInterface, cfg *config.Config, username, password string, waitTime time.Duration, screenshotName string, r *Result) {
	actionName := fmt.Sprintf("Member Login Attempt (User: %q)", username)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()
	homePage := pages.NewHomePage(mp.Page)
	loginPage := pages.NewLoginPage(mp.Page)

	if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Only enter credentials if non-empty
	if username != "" {
		if err := loginPage.EnterUsername(cfg.MemberLoginUsernameInputTestID, username, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			return
		}
	}
	if password != "" {
		if err := loginPage.EnterPassword(cfg.MemberLoginPasswordInputTestID, password, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			return
		}
	}

	if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-profile-card'], [data-testid='testid-member-subscriptions-button']", 5 * time.Second, screenshotName)
}

// LogoutMemberCustom logouts the member persona with customized wait time and screenshot.
func LogoutMemberCustom(mai MemberActionsInterface, cfg *config.Config, waitTime time.Duration, screenshotName string, r *Result) {
	actionName := "Logout Member"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()
	homePage := pages.NewHomePage(mp.Page)

	if err := homePage.Logout(cfg.LogoutButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	time.Sleep(waitTime)
	_ = homePage.GoToHome(mp.BaseURL)
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-profile-card'], [data-testid='testid-member-subscriptions-button']", 5 * time.Second, screenshotName)
}

// LoginAsMember logs in the member persona using credentials from the config.
func LoginAsMember(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Login As Member"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()
	homePage := pages.NewHomePage(mp.Page)
	loginPage := pages.NewLoginPage(mp.Page)

	if err := homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'memberLoginButtonTestID' element exists in header/navbar")
		return
	}

	if err := loginPage.EnterUsername(cfg.MemberLoginUsernameInputTestID, cfg.MemberCredentials.Username, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to enter member username")
		return
	}

	if err := loginPage.EnterPassword(cfg.MemberLoginPasswordInputTestID, cfg.MemberCredentials.Password, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to enter member password")
		return
	}

	// Take screenshot before submitting the form
	r.CaptureScreenshot(mp.Page, "MemberLogin_Filled_Form")

	if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to click submit login button")
		return
	}

	// Wait for dashboard to fully render after login redirect
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-logout-button']", 5 * time.Second, "MemberLogin_Dashboard")
	r.Advice = append(r.Advice, "Member logged in successfully.")
}

// ForceChangePassword handles the first-time login forced password change flow.
func ForceChangePassword(mai MemberActionsInterface, cfg *config.Config, email, tempPassword, newPassword string, r *Result) {
	actionName := "Force Change Password on First Login"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Open member login modal if not already open (assumes we are on home page)
	homePage := pages.NewHomePage(mp.Page)
	_ = homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, mp.DefaultTimeout)

	// Click the explicit "Change Password" button on the login form
	if err := mp.Page.ClickByTestID("testid-change-password-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Wait for change password dialog to appear
	time.Sleep(1 * time.Second)

	if err := mp.Page.SendKeysByTestID("testid-change-password-email-input", email, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-change-password-old-input", tempPassword, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-change-password-new-input", newPassword, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-change-password-confirm-input", newPassword, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.CaptureScreenshot(mp.Page, "ChangePassword_Filled")

	if err := mp.Page.ClickByTestID("testid-change-password-submit-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, "ChangePassword_Success")
	r.Advice = append(r.Advice, "Password changed successfully.")
}

// ValidateUnpaidMemberAccess verifies that an unpaid member can only access specific pages.
func ValidateUnpaidMemberAccess(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Validate Unpaid Member Access"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// 1. Visit allowed pages
	pages := []struct {
		TestID         string
		ScreenshotName string
		Locator        string
	}{
		{"testid-home-button", "Home_Page", "css:[data-testid='home-link']"},
		{"testid-office-bearers-button", "Office_Bearers_Page", "css:[data-testid='testid-office-bearers-district-select']"},
		{"testid-membership-button", "Member_Profile_Page", "css:body"},
		{"testid-member-subscriptions-button", "Subscriptions_Page", "css:body"},
		{"testid-member-announcements-button", "Announcements_Page", "css:body"},
		{"testid-events-button", "Events_Gallery_Page", "css:body"},
	}

	for _, page := range pages {
		if err := mp.Page.ClickByTestID(page.TestID, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify '%s' element exists", page.TestID))
			return
		}

		r.WaitForElementAndCapture(mp.Page, page.Locator, 5 * time.Second, page.ScreenshotName)
	}

	// 2. Verify restricted pages are NOT accessible (buttons should not be in DOM)
	restrictedButtons := []string{
		"testid-resources-button",
		"testid-taga-towers-button",
		"testid-grievance-button",
		"testid-admin-panel-button",
	}

	for _, btn := range restrictedButtons {
		// Attempt to find it, should timeout because it doesn't exist
		_, err := mp.Page.WaitUntilClickable(fmt.Sprintf("css:[data-testid='%s']", btn), 2*time.Second)
		if err == nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("unpaid member should NOT see the restricted button: %s", btn)
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Ensure unpaid members cannot access %s", btn))
			return
		}
	}
	r.CaptureScreenshot(mp.Page, "UnpaidMember_RestrictedAccess_Validated")

	r.Advice = append(r.Advice, "Successfully validated unpaid member access.")
}

// ValidatePaidMemberAccess verifies that a paid subscriber member can access all member-exclusive pages.
func ValidatePaidMemberAccess(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Validate Paid Member Access"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// 1. Visit accessible pages for paid member
	pages := []struct {
		TestID         string
		ScreenshotName string
		Locator        string
	}{
		{"testid-home-button", "Home_Page", "css:[data-testid='home-link']"},
		{"testid-office-bearers-button", "Office_Bearers_Page", "css:[data-testid='testid-office-bearers-district-select']"},
		{"testid-membership-button", "Member_Profile_Page", "css:body"},
		{"testid-resources-button", "Resources_Page", "css:body"},
		{"testid-taga-towers-button", "TAGA_Towers_Page", "css:body"},
		{"testid-events-button", "Events_Gallery_Page", "css:body"},
		{"testid-grievance-button", "Grievance_Page", "css:body"},
	}

	for _, page := range pages {
		if err := mp.Page.ClickByTestID(page.TestID, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify '%s' element exists for paid member", page.TestID))
			return
		}

		r.WaitForElementAndCapture(mp.Page, page.Locator, 5*time.Second, page.ScreenshotName)
	}

	// 2. Verify admin restricted pages are NOT accessible to regular paid member
	restrictedButtons := []string{
		"testid-members-button",
		"testid-member-login-button",
	}

	for _, btn := range restrictedButtons {
		_, err := mp.Page.WaitUntilClickable(fmt.Sprintf("css:[data-testid='%s']", btn), 2*time.Second)
		if err == nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("paid member should NOT see the admin/login button: %s", btn)
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Ensure regular paid members cannot access %s", btn))
			return
		}
	}
	r.CaptureScreenshot(mp.Page, "PaidMember_Access_Validated")

	r.Advice = append(r.Advice, "Successfully validated paid member access.")
}

// PayAnnualSubscription navigates to the subscription page and mocks an annual subscription payment.
func PayAnnualSubscription(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Pay Annual Subscription"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// 1. Navigate to Profile / Membership Page
	if err := mp.Page.ClickByTestID("testid-membership-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-membership-button' element exists")
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Switch to Subscriptions Tab
	if err := mp.Page.ClickByTestID("testid-member-subscriptions-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-member-subscriptions-button' exists in Membership tabs")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-subscriptions-button']", 5 * time.Second, "Subscription_Tab_Opened")

	// 3. Click 'Pay Now' for Annual Subscription
	if err := mp.Page.ClickByTestID("testid-pay-now-annual-subscription-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-pay-now-annual-subscription-button' exists. It might already be paid.")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:[role='dialog']", 5 * time.Second, "Payment_Modal_Opened")

	// 4. Inject Mock Razorpay
	mp.Page.InjectMockRazorpay()

	// 5. Click Proceed to Pay
	if err := mp.Page.ClickByTestID("testid-membership-payment-submit-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-membership-payment-submit-button' exists in modal")
		return
	}

	// Wait for processing and UI refresh (e.g., toast and modal close)
	time.Sleep(3 * time.Second)

	// 6. Check that UI updated to 'Paid'
	if err := mp.Page.VerifyFormElements([]string{"testid-paid-badge-annual-subscription"}, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-paid-badge-annual-subscription' is visible after payment")
		return
	}

	r.CaptureScreenshot(mp.Page, "Annual_Subscription_Paid")

	r.Advice = append(r.Advice, "Successfully completed mock annual subscription payment.")
}

// LogoutMember logs out the member persona.
func LogoutMember(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Logout Member"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()
	homePage := pages.NewHomePage(mp.Page)

	if err := homePage.Logout(cfg.LogoutButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'logoutButtonTestID' element exists in header/navbar")
		return
	}

	time.Sleep(2 * time.Second)
	_ = homePage.GoToHome(mp.BaseURL)
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-login-button']", 5 * time.Second, "MemberLogout_HomePage")
	r.Advice = append(r.Advice, "Member logged out successfully.")
}

// NavigateToTAGATower navigates a logged-in member to the TAGA Towers page.
func NavigateToTAGATower(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Navigate to TAGA Towers"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	if err := mp.Page.ClickByTestID("testid-taga-towers-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-taga-towers-button' exists on the sidebar")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-taga-towers-page']", 5 * time.Second, "TAGATower_Dashboard")
}

// BookSimpleRoom performs a straightforward booking for the logged-in member.
func BookSimpleRoom(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string) {
	actionName := "Book Simple TAGA Tower Room: " + roomID
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		r.Advice = append(r.Advice, "Advice: Room availability might not be loaded or room does not exist")
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify %s exists and room is available", bookBtnID))
		return
	}
	time.Sleep(2 * time.Second) // wait for modal to open

	// Fill Modal: Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-booker-phone-input' exists")
		return
	}

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Booking_Modal_%s", roomID))

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify payment button exists")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "xpath://li[@data-sonner-toast and contains(., 'Booking Successful')]", 5 * time.Second, fmt.Sprintf("TAGATower_Booking_Complete_%s", roomID))
}

// CancelLatestBooking cancels the most recently created booking by clicking the first cancel button.
func CancelLatestBooking(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Cancel Latest Booking"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Locate and click the first cancel button matching the partial data-testid
	clickCancelScript := `
	const cancelBtns = document.querySelectorAll('[data-testid$="-cancel-button"]');
	for (let btn of cancelBtns) {
		if (btn.getAttribute('data-testid').startsWith('testid-booking-')) {
			btn.click();
			return true;
		}
	}
	return false;
	`
	clicked, err := mp.Page.Driver.ExecuteScript(clickCancelScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click a booking cancel button: %v", err)
		r.Advice = append(r.Advice, "Advice: Make sure a booking exists in 'My Bookings'")
		return
	}
	time.Sleep(1 * time.Second)

	// Confirm cancellation in the modal
	if err := mp.Page.ClickByTestID("testid-confirm-booking-cancel-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-confirm-booking-cancel-button' exists in modal")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, "TAGATower_Booking_Cancelled")
}

// BookRoomForGuest performs a booking for a guest rather than the logged-in member.
func BookRoomForGuest(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, guestName string, guestAge string, guestContact string) {
	actionName := "Book TAGA Tower Room for Guest: " + roomID
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		r.Advice = append(r.Advice, "Advice: Room availability might not be loaded or room does not exist")
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify %s exists and room is available", bookBtnID))
		return
	}
	time.Sleep(2 * time.Second) // wait for modal to open

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-booker-phone-input' exists")
		return
	}

	// Select "Guest" / "Others"
	if _, err := mp.Page.Driver.ExecuteScript("document.getElementById('guest').click();", nil); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to select 'Guest' radio option")
		return
	}
	time.Sleep(1 * time.Second) // wait for guest input fields to appear

	// Fill Guest 1 details
	if err := mp.Page.SendKeysByTestID("testid-guest-1-name-input", guestName, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-guest-1-age-input", guestAge, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-guest-1-contact-input", guestContact, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	// Select Guest 1 Gender (Male)
	if _, err := mp.Page.Driver.ExecuteScript("const g = document.getElementById('guest-1-male'); if(g) g.click();", nil); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select guest 1 gender: %v", err)
		return
	}
	time.Sleep(500 * time.Millisecond)

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_GuestBooking_Modal_%s", roomID))

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify payment button exists")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "xpath://li[@data-sonner-toast and contains(., 'Booking Successful')]", 5 * time.Second, fmt.Sprintf("TAGATower_Guest_Booking_Complete_%s", roomID))
}

// BookAllBedsInRoom performs a booking for all beds in a dormitory or regular room.
func BookAllBedsInRoom(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, capacity int) {
	actionName := "Book All Beds in Room: " + roomID
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		r.Advice = append(r.Advice, "Advice: Room availability might not be loaded or room does not exist")
		return
	}

	// Click Book on the specific room (try direct click or JS click)
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn && !btn.disabled) {
			btn.scrollIntoView({behavior: 'instant', block: 'center'});
			btn.click();
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		if clickErr := mp.Page.ClickByTestID(bookBtnID, mp.DefaultTimeout); clickErr != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, clickErr)
			return
		}
	}

	// Wait for booking modal and booker phone input to be visible
	if _, err := mp.Page.WaitUntilVisible("css:[data-testid='testid-booker-phone-input']", 10*time.Second); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("booking modal / phone input did not appear: %w", err)
		r.Advice = append(r.Advice, "Advice: Verify room booking button triggers dialog and 'testid-booker-phone-input' exists")
		return
	}

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Select "Guest" / "Others"
	if _, err := mp.Page.Driver.ExecuteScript("const g = document.getElementById('guest'); if(g) { g.click(); }", nil); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	time.Sleep(1 * time.Second)

	// Check if Bed Count Select is visible in the UI (only for rooms allowing single beds)
	checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
	hasSelectObj, err := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
	if err == nil && hasSelectObj == true {
		// Select Bed Count (Capacity)
		if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err == nil {
			time.Sleep(1 * time.Second)
			selectOptionScript := fmt.Sprintf(`
				const options = document.querySelectorAll('[role="option"]');
				for(let opt of options) {
					const val = opt.getAttribute('data-value') || opt.getAttribute('value') || '';
					const txt = opt.textContent || '';
					if(val === '%d' || txt.includes('%d bed') || txt.includes('%d beds')) {
						opt.click();
						return true;
					}
				}
				return false;
			`, capacity, capacity, capacity)
			_, _ = mp.Page.Driver.ExecuteScript(selectOptionScript, nil)
			time.Sleep(1 * time.Second)
		}
	}

	// Fill guest details one by one and click "Add Guest" if there are more
	for i := 0; i < capacity; i++ {
		idx := i + 1
		nameID := fmt.Sprintf("testid-guest-%d-name-input", idx)
		ageID := fmt.Sprintf("testid-guest-%d-age-input", idx)
		contactID := fmt.Sprintf("testid-guest-%d-contact-input", idx)

		// Fill name
		nameVal := fmt.Sprintf("Male Guest %d", idx)
		if err := mp.Page.SendKeysByTestID(nameID, nameVal, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to fill guest %d name: %v", idx, err)
			return
		}

		// Fill age
		ageVal := fmt.Sprintf("%d", 20+i)
		if err := mp.Page.SendKeysByTestID(ageID, ageVal, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to fill guest %d age: %v", idx, err)
			return
		}

		// Fill contact
		contactVal := fmt.Sprintf("98765432%02d", idx)
		if err := mp.Page.SendKeysByTestID(contactID, contactVal, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to fill guest %d contact: %v", idx, err)
			return
		}

		// Select gender (Male)
		genderScript := fmt.Sprintf("const g = document.getElementById('guest-%d-male'); if(g) g.click();", idx)
		if _, err := mp.Page.Driver.ExecuteScript(genderScript, nil); err != nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("failed to select guest %d gender: %v", idx, err)
			return
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_AllRoomBooking_Modal_%s", roomID))

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, fmt.Sprintf("TAGATower_AllRoomBooking_Complete_%s", roomID))
}

// TryBookRoomAsSelfMultibooking tries to book multiple beds for Self, which should be disallowed.
// If it succeeds in proceeding to payment/mocking Razorpay (or doesn't block it), the action fails,
// because only guest bookings should allow multibed booking.
func TryBookRoomAsSelfMultibooking(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, beds int) {
	actionName := "Try Book Room as Self Multibooking: " + roomID
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		r.Advice = append(r.Advice, "Advice: Room availability might not be loaded or room does not exist")
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		return
	}
	time.Sleep(2 * time.Second) // wait for modal to open

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Select "Self" explicitly
	if _, err := mp.Page.Driver.ExecuteScript("document.getElementById('self').click();", nil); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	time.Sleep(1 * time.Second)

	// Check if Bed Count Select is visible in the UI
	// Since Self bookings are strictly 1 bed, the UI must hide the bed-count selector entirely.
	checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
	hasSelectObj, err := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
	if err == nil && hasSelectObj == false {
		// Bed count selector is properly hidden for Self booking! This is the expected secure behavior.
		r.Advice = append(r.Advice, "UI correctly hid the bed count selector for Self booking (locked to 1 bed). Secure behavior verified.")
		return
	}

	// If selector is present, try to click Bed Count Select
	if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "Bed count selector cannot be clicked for Self booking. Secure behavior verified.")
		return
	}
	time.Sleep(1 * time.Second)

	// Try to select beds > 1 (e.g. 2 beds)
	bedText := fmt.Sprintf("%d bed", beds)
	selectOptionScript := fmt.Sprintf(`
		const options = document.querySelectorAll('[role="option"]');
		for(let opt of options) {
			if(opt.textContent.includes('%s')) {
				opt.click();
				return true;
			}
		}
		return false;
	`, bedText)
	optionClicked, err := mp.Page.Driver.ExecuteScript(selectOptionScript, nil)
	if err != nil || optionClicked != true {
		r.Advice = append(r.Advice, "System prevented selecting multiple beds for Self booking (Option not selectable). This is the expected secure behavior.")
		return
	}
	time.Sleep(1 * time.Second)

	// Capture modal state before proceeding
	r.CaptureScreenshot(mp.Page, "TAGATower_SelfMultibooking_Attempt")

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed to payment
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "System prevented proceeding to payment for self multibooking.")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, "TAGATower_SelfMultibooking_Allowed_Failure")

	// Clean up the created booking immediately
	cleanActionName := "Cleanup Stale Booking"
	r.Actions = append(r.Actions, cleanActionName)

	// Click the cancel button in My Bookings list
	clickCancelScript := `
	const cancelBtns = document.querySelectorAll('[data-testid$="-cancel-button"]');
	for (let btn of cancelBtns) {
		if (btn.getAttribute('data-testid').startsWith('testid-booking-')) {
			btn.click();
			return true;
		}
	}
	return false;
	`
	clickedCancel, cErr := mp.Page.Driver.ExecuteScript(clickCancelScript, nil)
	if cErr == nil && clickedCancel == true {
		time.Sleep(1 * time.Second)
		if err := mp.Page.ClickByTestID("testid-confirm-booking-cancel-button", mp.DefaultTimeout); err == nil {
			r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, "TAGATower_SelfMultibooking_Cancelled_Cleanup")
			r.Advice = append(r.Advice, "Cleanup: Cancelled the allowed self-multibooking.")
		}
	}

	// Mark the result as failed
	r.Status = "failed"
	r.Error = fmt.Errorf("security vulnerability: member was allowed to book %d beds for Self (multibooking). Only guest bookings should allow multibed bookings", beds)
	mp.Page.LastError = r.Error
	r.Advice = append(r.Advice, "Advice: Update frontend and backend validation to disallow BedCount > 1 when booking for Self")
}

// TryGuestBookingWithIncompleteGuestDetails tries to book N beds in guest mode
// but only fills in the details for the first guest, then attempts to proceed.
// The action EXPECTS the system to block the booking (toast error), and FAILS if
// the payment flow is reached — proving the "fill all N guests" gate works.
func TryGuestBookingWithIncompleteGuestDetails(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, beds int) {
	actionName := fmt.Sprintf("Try Guest Booking with Incomplete Details (%d beds, 1 filled): %s", beds, roomID)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for and click the room book button
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		r.Advice = append(r.Advice, "Advice: Room availability might not be loaded or room does not exist")
		return
	}
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)
	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		return
	}
	time.Sleep(2 * time.Second)

	// Fill booker phone
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-booker-phone-input' exists")
		return
	}

	// Switch to Guest mode
	if _, err := mp.Page.Driver.ExecuteScript("document.getElementById('guest').click();", nil); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to select 'Guest' radio option")
		return
	}
	time.Sleep(1 * time.Second)

	// Select N beds (only valid when room allowsSingleBed)
	if beds > 1 {
		checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
		hasSelectObj, selectErr := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
		if selectErr == nil && hasSelectObj == true {
			if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err == nil {
				time.Sleep(1 * time.Second)
				bedText := fmt.Sprintf("%d bed", beds)
				selectOptionScript := fmt.Sprintf(`
					const options = document.querySelectorAll('[role="option"]');
					for(let opt of options) {
						if(opt.textContent.includes('%s')) {
							opt.click();
							return true;
						}
					}
					return false;
				`, bedText)
				_, _ = mp.Page.Driver.ExecuteScript(selectOptionScript, nil)
				time.Sleep(1 * time.Second)
			}
		}
	}

	// Fill ONLY Guest 1 — intentionally leave guests 2..N blank
	if err := mp.Page.SendKeysByTestID("testid-guest-1-name-input", "Partial Guest", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill guest 1 name: %v", err)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-guest-1-age-input", "30", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill guest 1 age: %v", err)
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-guest-1-contact-input", "9876543299", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fill guest 1 contact: %v", err)
		return
	}
	// Select Guest 1 Gender (Male)
	if _, err := mp.Page.Driver.ExecuteScript("const g = document.getElementById('guest-1-male'); if(g) g.click();", nil); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select guest 1 gender: %v", err)
		return
	}
	time.Sleep(300 * time.Millisecond)

	// Capture state before attempting to proceed
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_IncompleteGuest_Before_%s", roomID))

	// Attempt to proceed to payment — the system MUST block this
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("proceed-to-payment button not found: %v", err)
		return
	}

	// Wait briefly for either a toast error OR the Razorpay modal
	time.Sleep(2 * time.Second)
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_IncompleteGuest_After_%s", roomID))

	// Check whether a toast error appeared (expected) rather than the payment proceeding
	toastCheckScript := `
		const toasts = document.querySelectorAll('li[data-sonner-toast]');
		for (let t of toasts) {
			const txt = t.textContent || '';
			if (txt.includes('Guest') || txt.includes('guest') || txt.includes('details') || txt.includes('fill')) {
				return txt;
			}
		}
		return null;
	`
	toastVal, toastErr := mp.Page.Driver.ExecuteScript(toastCheckScript, nil)
	if toastErr != nil || toastVal == nil {
		// Toast not found — the system let the user through without all guest details filled
		r.Status = "failed"
		r.Error = fmt.Errorf("validation failure: system did not block booking when only 1/%d guest details were filled", beds)
		mp.Page.LastError = r.Error
		r.Advice = append(r.Advice, "Advice: Frontend must require all N guest detail forms to be fully filled before proceeding to payment")
		return
	}

	// Toast appeared — system correctly blocked the booking
	r.Advice = append(r.Advice, fmt.Sprintf("System correctly blocked incomplete guest booking with toast: %v", toastVal))
}


// SelectFutureDates selects a 1-day date range (e.g. Day 1 to Day 2) using synthetic MouseEvents like in BookRoomForTenDays
func SelectFutureDates(mai MemberActionsInterface, r *Result) {
	actionName := "Select Future Dates in Calendar"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	select1DayScript := `
	const calendar = document.querySelector('[data-testid="testid-room-date-range-calendar"]');
	if (!calendar) return false;
	const buttons = Array.from(calendar.querySelectorAll('button'));
	const days = buttons.filter(btn => {
		const text = btn.textContent.trim();
		const num = Number(text);
		return !isNaN(num) && num >= 1 && num <= 31 && 
		       !btn.disabled && 
		       btn.getAttribute('aria-disabled') !== 'true' &&
		       !btn.classList.contains('day-outside') &&
		       !btn.classList.contains('rdp-day_outside');
	});
	if (days.length >= 3) {
		const clickEl = (el) => {
			['mousedown', 'mouseup', 'click'].forEach(type => {
				el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
			});
		};
		// Click Day 1 (check-in) twice, then Day 2 (check-out)
		clickEl(days[1]);
		setTimeout(() => {
			clickEl(days[1]);
			setTimeout(() => {
				clickEl(days[2]);
			}, 300);
		}, 300);
		return true;
	}
	return false;
	`
	selected, err := mp.Page.Driver.ExecuteScript(select1DayScript, nil)
	if err != nil || selected == false {
		r.Advice = append(r.Advice, "Note: Could not select 1-day date range via script, proceeding with default dates")
	} else {
		time.Sleep(3 * time.Second) // wait for availability to refresh
	}
}

// SelectThreeDaysOverlappingDates selects a 3-day range in the calendar (Day 0 to Day 3) using synthetic MouseEvents
func SelectThreeDaysOverlappingDates(mai MemberActionsInterface, r *Result) {
	actionName := "Select 3 Consecutive Days Spanning Booked Date"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	select3DaysScript := `
	const calendar = document.querySelector('[data-testid="testid-room-date-range-calendar"]');
	if (!calendar) return false;
	const buttons = Array.from(calendar.querySelectorAll('button'));
	const days = buttons.filter(btn => {
		const text = btn.textContent.trim();
		const num = Number(text);
		return !isNaN(num) && num >= 1 && num <= 31 && 
		       !btn.disabled && 
		       btn.getAttribute('aria-disabled') !== 'true' &&
		       !btn.classList.contains('day-outside') &&
		       !btn.classList.contains('rdp-day_outside');
	});
	if (days.length >= 4) {
		const clickEl = (el) => {
			['mousedown', 'mouseup', 'click'].forEach(type => {
				el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
			});
		};
		// Click Day 0 (check-in) twice, then Day 3 (check-out) spanning Day 1-2
		clickEl(days[0]);
		setTimeout(() => {
			clickEl(days[0]);
			setTimeout(() => {
				clickEl(days[3]);
			}, 300);
		}, 300);
		return true;
	}
	return false;
	`
	selected, err := mp.Page.Driver.ExecuteScript(select3DaysScript, nil)
	if err != nil || selected == false {
		r.Advice = append(r.Advice, "Note: Could not select 3-day range via script, proceeding with current selection")
	}

	time.Sleep(3 * time.Second) // wait for availability to refresh
	r.CaptureScreenshot(mp.Page, "TAGATower_3Days_DateRange_Selected")
}

// BookSingleBedWithGender books 1 bed in a room with a specific gender.
func BookSingleBedWithGender(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, gender string) {
	actionName := fmt.Sprintf("Book Single Bed in Room %s as %s", roomID, gender)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		return
	}
	time.Sleep(2 * time.Second) // wait for modal to open

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Ensure Bed Count is 1
	checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
	hasSelectObj, err := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
	if err == nil && hasSelectObj == true {
		if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err == nil {
			time.Sleep(1 * time.Second)
			selectOptionScript := `
				const options = document.querySelectorAll('[role="option"]');
				for(let opt of options) {
					if(opt.textContent.includes('1 bed')) {
						opt.click();
						return true;
					}
				}
				return false;
			`
			_, _ = mp.Page.Driver.ExecuteScript(selectOptionScript, nil)
			time.Sleep(1 * time.Second)
		}
	}

	// Click the gender radio button
	genderScript := fmt.Sprintf("document.getElementById('%s').click();", gender)
	if _, err := mp.Page.Driver.ExecuteScript(genderScript, nil); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select gender %s: %v", gender, err)
		return
	}
	time.Sleep(1 * time.Second)

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_GenderBooking_%s_%s", roomID, gender))

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	r.WaitForElementAndCapture(mp.Page, "xpath://li[@data-sonner-toast and contains(., 'Booking Successful')]", 5 * time.Second, fmt.Sprintf("TAGATower_GenderBooking_Complete_%s_%s", roomID, gender))
}

// TryBookSingleBedWithGenderOpposite tries to book 1 bed in a room with a specific gender, expecting it to be disallowed.
// If it is allowed (proceeds to payment/Razorpay), we fail the test and then clean it up.
func TryBookSingleBedWithGenderOpposite(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, gender string) {
	actionName := fmt.Sprintf("Try Book Single Bed in Room %s as opposite gender %s", roomID, gender)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("room button %s did not become visible: %v", bookBtnID, err)
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		return
	}
	time.Sleep(2 * time.Second) // wait for modal to open

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Ensure Bed Count is 1
	checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
	hasSelectObj, err := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
	if err == nil && hasSelectObj == true {
		if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err == nil {
			time.Sleep(1 * time.Second)
			selectOptionScript := `
				const options = document.querySelectorAll('[role="option"]');
				for(let opt of options) {
					if(opt.textContent.includes('1 bed')) {
						opt.click();
						return true;
					}
				}
				return false;
			`
			_, _ = mp.Page.Driver.ExecuteScript(selectOptionScript, nil)
			time.Sleep(1 * time.Second)
		}
	}

	// Click the opposite gender radio button
	genderScript := fmt.Sprintf("document.getElementById('%s').click();", gender)
	if _, err := mp.Page.Driver.ExecuteScript(genderScript, nil); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select gender %s: %v", gender, err)
		return
	}
	time.Sleep(1 * time.Second)

	// Capture modal state before clicking proceed
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_OppositeGender_Attempt_%s_%s", roomID, gender))

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		// If it failed to click (or showed validation error), that is expected behavior!
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_OppositeGender_Blocked_%s_%s", roomID, gender))
		r.Advice = append(r.Advice, "System prevented proceeding to payment for opposite gender booking.")
		return
	}

	time.Sleep(1 * time.Second)

	// Check if a toast error appeared or if Razorpay modal was NOT opened
	toastCheckScript := `
		const toasts = document.querySelectorAll('li[data-sonner-toast]');
		for (let t of toasts) {
			const txt = t.textContent || '';
			if (txt.includes('partially occupied') || txt.includes('only') || txt.includes('guests') || txt.includes('gender')) {
				return txt;
			}
		}
		return null;
	`
	toastVal, _ := mp.Page.Driver.ExecuteScript(toastCheckScript, nil)

	// Check if modal is still open (payment was blocked)
	checkModalScript := `return !!document.querySelector('[data-testid="testid-room-booking-modal"]');`
	modalOpenObj, _ := mp.Page.Driver.ExecuteScript(checkModalScript, nil)
	modalOpen, _ := modalOpenObj.(bool)

	if toastVal != nil || modalOpen {
		// Capture blocked screenshot
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_OppositeGender_Blocked_%s_%s", roomID, gender))
		r.Advice = append(r.Advice, fmt.Sprintf("System correctly blocked booking for opposite gender '%s' on partially occupied room %s (Toast/Block verified).", gender, roomID))

		// Close modal if open
		closeModalScript := `
			const closeBtn = document.querySelector('[data-testid="testid-booking-modal-cancel-button"]');
			if (closeBtn) closeBtn.click();
		`
		_, _ = mp.Page.Driver.ExecuteScript(closeModalScript, nil)
		time.Sleep(1 * time.Second)
		return
	}

	// Capture modal state of the allowed (but should be disallowed) booking
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_GenderMix_Allowed_Failure_%s_%s", roomID, gender))

	// Clean up both the allowed opposite-gender booking and the original booking
	clickCancelScript := `
	const cancelBtns = document.querySelectorAll('[data-testid$="-cancel-button"]');
	for (let btn of cancelBtns) {
		if (btn.getAttribute('data-testid').startsWith('testid-booking-')) {
			btn.click();
			return true;
		}
	}
	return false;
	`
	for i := 0; i < 2; i++ {
		clickedCancel, cErr := mp.Page.Driver.ExecuteScript(clickCancelScript, nil)
		if cErr == nil && clickedCancel == true {
			time.Sleep(1 * time.Second)
			if err := mp.Page.ClickByTestID("testid-confirm-booking-cancel-button", mp.DefaultTimeout); err == nil {
				time.Sleep(3 * time.Second)
			}
		}
	}

	// Mark test as failed since opposite gender booking was allowed
	r.Status = "failed"
	r.Error = fmt.Errorf("security vulnerability: room %s allowed booking for both male and female concurrently. Gender mixed dorm bookings should be disallowed", roomID)
	mp.Page.LastError = r.Error
	r.Advice = append(r.Advice, "Advice: Update frontend and backend to block opposite gender bookings for partially occupied single-bed rooms")
}

// VerifyDormitoryGenderRestrictionUI verifies that the gender selection is locked to a specific gender
// via a read-only badge and that no radio buttons for gender selection are present in the DOM.
func VerifyDormitoryGenderRestrictionUI(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, expectedBadgeText string) {
	actionName := fmt.Sprintf("Verify Dormitory %s Restricts Gender to %s", roomID, expectedBadgeText)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room to load before clicking
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	if _, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", bookBtnID), mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("dormitory button %s did not become visible: %v", bookBtnID, err)
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)

	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to find or click %s: %v", bookBtnID, err)
		return
	}
	time.Sleep(2 * time.Second) // wait for modal to open

	// Ensure Bed Count is 1 to reveal the gender restriction UI
	checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
	hasSelectObj, err := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
	if err == nil && hasSelectObj == true {
		if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err == nil {
			time.Sleep(1 * time.Second)
			selectOptionScript := `
				const options = document.querySelectorAll('[role="option"]');
				for(let opt of options) {
					if(opt.textContent.includes('1 bed')) {
						opt.click();
						return true;
					}
				}
				return false;
			`
			_, _ = mp.Page.Driver.ExecuteScript(selectOptionScript, nil)
			time.Sleep(1 * time.Second)
		}
	}

	// Verify the radio group is ABSENT
	hasRadioGroupScript := `return !!document.querySelector('[data-testid="testid-booking-gender-radio-group"]');`
	hasRadioGroup, err := mp.Page.Driver.ExecuteScript(hasRadioGroupScript, nil)
	if err == nil && hasRadioGroup == true {
		r.Status = "failed"
		r.Error = fmt.Errorf("gender selection radio group was found in %s, but it should be a read-only badge", roomID)
		return
	}

	// Verify the static badge text is PRESENT
	hasBadgeScript := fmt.Sprintf(`
		const elements = Array.from(document.querySelectorAll('*'));
		return elements.some(el => el.textContent === '%s' && el.classList.contains('inline-flex'));
	`, expectedBadgeText)
	hasBadge, err := mp.Page.Driver.ExecuteScript(hasBadgeScript, nil)
	if err != nil || hasBadge == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("expected read-only badge '%s' was not found in %s modal", expectedBadgeText, roomID)
		return
	}

	// Capture modal state of the correct UI
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Dormitory_%s_Restriction_Verified", roomID))

	// Close the modal
	closeScript := `
		const closeBtn = document.querySelector('button[aria-label="Close"], dialog form[method="dialog"] button, .lucide-x');
		if(closeBtn) closeBtn.click();
		else {
			const closeBtnFallback = document.querySelector('[role="dialog"] button');
			if(closeBtnFallback) closeBtnFallback.click();
		}
	`
	_, _ = mp.Page.Driver.ExecuteScript(closeScript, nil)
	time.Sleep(1 * time.Second)
}

// BookRoomForTenDays books a room for a duration of 10 days (from today to today + 9 days).
func BookRoomForTenDays(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string) {
	actionName := "Book Room for 10 Days: " + roomID
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Select start date (today) twice, then select 10th day (today + 9 days) once
	selectDatesScript := `
	const calendar = document.querySelector('[data-testid="testid-room-date-range-calendar"]');
	if (!calendar) return false;
	const buttons = Array.from(calendar.querySelectorAll('button'));
	const days = buttons.filter(btn => {
		const text = btn.textContent.trim();
		const num = Number(text);
		return !isNaN(num) && num >= 1 && num <= 31 && 
		       !btn.disabled && 
		       btn.getAttribute('aria-disabled') !== 'true' &&
		       !btn.classList.contains('day-outside') &&
		       !btn.classList.contains('rdp-day_outside');
	});
	if (days.length >= 10) {
		const clickEl = (el) => {
			['mousedown', 'mouseup', 'click'].forEach(type => {
				el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
			});
		};
		// Click check-in date (today) twice
		clickEl(days[0]);
		setTimeout(() => {
			clickEl(days[0]);
			// Click check-out date (10th day) after another delay
			setTimeout(() => {
				clickEl(days[9]);
			}, 300);
		}, 300);
		return true;
	}
	return false;
	`
	selected, err := mp.Page.Driver.ExecuteScript(selectDatesScript, nil)
	if err != nil || selected == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select 10 days range on calendar: %v", err)
		return
	}
	time.Sleep(3 * time.Second) // wait for UI/room list to refresh

	// Verify if the room book button is visible.
	// If it's not visible or full, this is the bug! The test should fail here.
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	checkVisibleScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		return !!btn && btn.offsetParent !== null && !btn.disabled;
	`, bookBtnID)
	visibleObj, err := mp.Page.Driver.ExecuteScript(checkVisibleScript, nil)
	visible, _ := visibleObj.(bool)

	if err != nil || !visible {
		r.Status = "failed"
		r.Error = fmt.Errorf("bug detected: room %s is not available or shown as full for a 10-day booking", roomID)
		mp.Page.LastError = r.Error
		r.Advice = append(r.Advice, "Advice: Fix the frontend room availability filter logic when handling 10-day bookings")

		// Capture failure screenshot
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_10DaysBooking_FullBug_%s", roomID))
		return
	}

	// Click Book on the specific room using JS to bypass sticky navbar
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)
	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click book button for %s: %v", roomID, err)
		return
	}
	time.Sleep(2 * time.Second)

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_10DaysBooking_Modal_%s", roomID))

	// Mocking Razorpay
	mp.Page.InjectMockRazorpay()

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, fmt.Sprintf("TAGATower_10DaysBooking_Complete_%s", roomID))
}

// TryBookOverlappingRoom attempts to book the same room for overlapping dates and expects it to be blocked.
func TryBookOverlappingRoom(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string) {
	actionName := "Try Book Overlapping Room: " + roomID
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Wait for room availability and cards to update
	time.Sleep(3 * time.Second)

	// Verify if the room book button is visible and active (not disabled/Full).
	// If it is disabled or has offsetParent == null (or doesn't exist), that means it's correctly blocked!
	bookBtnID := fmt.Sprintf("testid-room-%s-book-button", roomID)
	checkVisibleScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		return !!btn && btn.offsetParent !== null && !btn.disabled;
	`, bookBtnID)
	visibleObj, err := mp.Page.Driver.ExecuteScript(checkVisibleScript, nil)
	visible, _ := visibleObj.(bool)

	if err != nil || !visible {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_3Days_Blocked_%s", roomID))
		r.Advice = append(r.Advice, fmt.Sprintf("System correctly blocked 3-day overlapping booking for room %s (Room button is disabled/Full).", roomID))
		return
	}

	// If button was active, try clicking Book to check if payment proceed is blocked
	jsClickScript := fmt.Sprintf(`
		const btn = document.querySelector('[data-testid="%s"]');
		if(btn) {
			btn.scrollIntoView({behavior: 'smooth', block: 'center'});
			setTimeout(() => btn.click(), 500);
			return true;
		}
		return false;
	`, bookBtnID)
	clicked, err := mp.Page.Driver.ExecuteScript(jsClickScript, nil)
	if err != nil || clicked == false {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_3Days_Blocked_%s", roomID))
		r.Advice = append(r.Advice, "System correctly blocked overlapping booking (could not click book button).")
		return
	}
	time.Sleep(2 * time.Second)

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "System correctly prevented proceeding with overlapping booking.")
		return
	}

	// Capture modal state before clicking proceed
	time.Sleep(1 * time.Second)
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_Attempt_%s", roomID))

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_3Days_Blocked_%s", roomID))
		r.Advice = append(r.Advice, "System correctly blocked proceeding to payment for overlapping booking.")
		return
	}
	time.Sleep(2 * time.Second)

	// Check if a toast error appeared or if modal stayed open without creating Razorpay payment
	toastCheckScript := `
		const toasts = document.querySelectorAll('li[data-sonner-toast]');
		for (let t of toasts) {
			const txt = t.textContent || '';
			if (txt.includes('not enough beds') || txt.includes('full') || txt.includes('already booked') || txt.includes('no longer available')) {
				return txt;
			}
		}
		return null;
	`
	toastVal, _ := mp.Page.Driver.ExecuteScript(toastCheckScript, nil)

	checkModalScript := `return !!document.querySelector('[data-testid="testid-room-booking-modal"]');`
	modalOpenObj, _ := mp.Page.Driver.ExecuteScript(checkModalScript, nil)
	modalOpen, _ := modalOpenObj.(bool)

	if toastVal != nil || modalOpen {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_3Days_Blocked_%s", roomID))
		r.Advice = append(r.Advice, fmt.Sprintf("System correctly blocked 3-day overlapping booking for %s (Toast/Modal block verified).", roomID))

		// Close modal if open
		closeModalScript := `
			const closeBtn = document.querySelector('[data-testid="testid-booking-modal-cancel-button"]');
			if (closeBtn) closeBtn.click();
		`
		_, _ = mp.Page.Driver.ExecuteScript(closeModalScript, nil)
		time.Sleep(1 * time.Second)
		return
	}

	// Capture modal state of the allowed (but should be disallowed) booking
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_Allowed_Failure_%s", roomID))

	// Clean up the overlapping booking if it got created
	clickCancelScript := `
	const cancelBtns = document.querySelectorAll('[data-testid$="-cancel-button"]');
	for (let btn of cancelBtns) {
		if (btn.getAttribute('data-testid').startsWith('testid-booking-')) {
			btn.click();
			return true;
		}
	}
	return false;
	`
	clickedCancel, cErr := mp.Page.Driver.ExecuteScript(clickCancelScript, nil)
	if cErr == nil && clickedCancel == true {
		time.Sleep(1 * time.Second)
		if err := mp.Page.ClickByTestID("testid-confirm-booking-cancel-button", mp.DefaultTimeout); err == nil {
			time.Sleep(3 * time.Second)
		}
	}

	// Mark test as failed since overlapping booking was allowed
	r.Status = "failed"
	r.Error = fmt.Errorf("security vulnerability: room %s allowed booking for overlapping dates when it was already occupied", roomID)
	mp.Page.LastError = r.Error
	r.Advice = append(r.Advice, "Advice: Enforce strict backend validation to prevent overlapping bookings for the same room")
}

// ViewMemberSubscriptions navigates to the subscription page and takes a screenshot.
func ViewMemberSubscriptions(mai MemberActionsInterface, cfg *config.Config, screenshotName string, r *Result) {
	actionName := "View Member Subscriptions"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// 1. Navigate to Profile / Membership Page
	if err := mp.Page.ClickByTestID("testid-membership-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-membership-button' element exists")
		return
	}
	time.Sleep(1 * time.Second)

	// 2. Switch to Subscriptions Tab
	if err := mp.Page.ClickByTestID("testid-member-subscriptions-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-member-subscriptions-button' exists in Membership tabs")
		return
	}
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-profile-card'], [data-testid='testid-member-subscriptions-button']", 5 * time.Second, screenshotName)
}

// ForceMemberPasswords is a test utility that bypasses the lack of API exposure for temp passwords.
// It directly overrides the password hash in the members database file and unsets the first_login flag.
func ForceMemberPasswords(emails []string, newPassword string) {
	membersFile := "/home/sudhan_dev/Downloads/code/nammataga/taga-api/data/member/members.json"
	data, err := os.ReadFile(membersFile)
	if err != nil {
		fmt.Printf("WARNING: Failed to read members file %s: %v\n", membersFile, err)
		return
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		fmt.Printf("WARNING: Failed to parse members file: %v\n", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("WARNING: Failed to hash password: %v\n", err)
		return
	}

	emailMap := make(map[string]bool)
	for _, e := range emails {
		emailMap[strings.ToLower(strings.TrimSpace(e))] = true
	}

	changed := false
	for i, m := range members {
		if mEmail, ok := m["emailId"].(string); ok {
			if emailMap[strings.ToLower(strings.TrimSpace(mEmail))] {
				members[i]["password"] = string(hashedPassword)
				members[i]["first_login"] = false
				changed = true
			}
		}
	}

	if changed {
		updatedData, _ := json.MarshalIndent(members, "", "  ")
		_ = os.WriteFile(membersFile, updatedData, 0644)
	}
}

// VerifyMemberSubscriptionStatus logs in as a member, checks their subscription status, and logs out.
func VerifyMemberSubscriptionStatus(mai MemberActionsInterface, cfg *config.Config, username, password, stepPrefix string, r *Result) {
	mp := mai.GetMemberPersona()
	actionName := fmt.Sprintf("Verify Subscription Status (%s)", username)
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	MemberLoginAttempt(mai, cfg, username, password, 4*time.Second, stepPrefix+"_Login", r)
	if r.Failed() {
		return
	}

	GoToHome(mai, r)
	if r.Failed() {
		return
	}

	// Wait for home page to render, click Membership, then Subscriptions
	if err := mp.Page.ClickByTestID("testid-membership-button", 5*time.Second); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-membership-button' exists on member dashboard")
		return
	}
	time.Sleep(1 * time.Second)

	if err := mp.Page.ClickByTestID("testid-member-subscriptions-button", 5*time.Second); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-member-subscriptions-button' exists in Membership dropdown")
		return
	}

	// Capture the subscription status
	r.WaitForElementAndCapture(mp.Page, "css:[data-testid='testid-member-subscriptions-button']", 5*time.Second, stepPrefix+"_Subscription_Status")

	// Log out safely
	LogoutMember(mai, cfg, r)
}

// ClearMockEmails sends a request to the API backend to clear the mock emails file.
func ClearMockEmails(cfg *config.Config) {
	req, err := http.NewRequest(http.MethodDelete, cfg.BaseURL+"api/admin/mock-emails", nil)
	if err == nil {
		client := &http.Client{Timeout: 5 * time.Second}
		client.Do(req)
	}
}

// FetchTempPasswordFromMockEmail reads the mock_emails.json file via the API backend to get the temporary password.
func FetchTempPasswordFromMockEmail(cfg *config.Config, email string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		resp, err := client.Get(cfg.BaseURL + "api/admin/mock-emails")
		if err != nil || resp.StatusCode != http.StatusOK {
			continue
		}
		
		var mockEmails map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&mockEmails); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		
		if pass, ok := mockEmails[email]; ok && pass != "" {
			return pass
		}
	}
	return ""
}

// VerifyBulkUploadedMember performs the full password change and subscription validation for a bulk-uploaded member.
func VerifyBulkUploadedMember(page *ui.Page, cfg *config.Config, email, stepPrefix string, result *Result) {
	member := NewMemberPersona(page, cfg.UiURL, 5*time.Second)
	tempPass := FetchTempPasswordFromMockEmail(cfg, email)
	
	GoToHome(member, result)
	ForceChangePassword(member, cfg, email, tempPass, cfg.MemberCredentials.Password, result)
	VerifyMemberSubscriptionStatus(member, cfg, email, cfg.MemberCredentials.Password, stepPrefix, result)
}

// ForgotPasswordByEmail triggers forgot password using the email option
func ForgotPasswordByEmail(mai MemberActionsInterface, cfg *config.Config, email string, r *Result) {
	mp := mai.GetMemberPersona()
	actionName := "Forgot Password by Email"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		return
	}

	homePage := pages.NewHomePage(mp.Page)
	_ = homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, mp.DefaultTimeout)

	// Click forgot password button to open the modal
	if err := mp.Page.ClickByTestID("testid-forgot-password-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	time.Sleep(1 * time.Second)

	// Enter email into forgot password email input
	if err := mp.Page.SendKeysByTestID("testid-forgot-password-email-input", email, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Click reset button
	if err := mp.Page.ClickByTestID("testid-reset-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, "ForgotPasswordByEmail_Success")
}

// ForgotPasswordByTbf triggers forgot password using the TBF number option
func ForgotPasswordByTbf(mai MemberActionsInterface, cfg *config.Config, tbfNumber, email string, r *Result) {
	mp := mai.GetMemberPersona()
	actionName := "Forgot Password by TBF Number"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		return
	}

	homePage := pages.NewHomePage(mp.Page)
	_ = homePage.OpenMemberLogin(cfg.MemberLoginButtonTestID, mp.DefaultTimeout)

	// Click forgot password button to open the modal
	if err := mp.Page.ClickByTestID("testid-forgot-password-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	time.Sleep(1 * time.Second)

	// Click "Try other way" button
	if err := mp.Page.ClickByTestID("testid-try-other-way-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	time.Sleep(500 * time.Millisecond)

	// Enter TBF Number into input
	if err := mp.Page.SendKeysByTestID("testid-forgot-password-tbf-input", tbfNumber, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Click reset button
	if err := mp.Page.ClickByTestID("testid-reset-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.WaitForElementAndCapture(mp.Page, "css:li[data-sonner-toast]", 5 * time.Second, "ForgotPasswordByTbf_Success")
}

// ResetForgotPasswordWithTemporaryPassword retrieves the temporary password from mock emails and performs the password change
func ResetForgotPasswordWithTemporaryPassword(mai MemberActionsInterface, cfg *config.Config, email, newPassword string, r *Result) {
	actionName := "Reset Forgot Password with Temporary Password"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		return
	}

	tempPass := FetchTempPasswordFromMockEmail(cfg, email)
	if tempPass == "" {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to fetch temporary password from mock email")
		return
	}

	ForceChangePassword(mai, cfg, email, tempPass, newPassword, r)
}


// MemberSubmitsEditRequest submits a profile edit request as a logged-in member.
func MemberSubmitsEditRequest(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Member Submits Edit Request"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		return
	}
	mp := mai.GetMemberPersona()
	
	rand.Seed(time.Now().UnixNano())
	randomMobile := fmt.Sprintf("9%09d", rand.Intn(1000000000))
	designations := []string{"Agricultural Officer", "Assistant Director of Agriculture", "Deputy Director of Agriculture", "Joint Director of Agriculture", "Additional Director of Agriculture"}
	randomDesignation := designations[rand.Intn(len(designations))]


	// Navigate to Membership Page
	if err := mp.Page.ClickByTestID("testid-membership-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Click "Edit Profile" button
	if err := mp.Page.ClickByTestID("testid-request-profile-edit-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Fill mobile number
	if err := mp.Page.SendKeysByTestID("testid-profile-edit-mobile-input", randomMobile, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	// Fill working district
	if err := mp.Page.SendKeysByTestID("testid-profile-edit-designation-input", randomDesignation, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}
	// Fill remarks
	if err := mp.Page.SendKeysByTestID("testid-profile-edit-remarks-input", "E2E Test Member Remarks", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.CaptureScreenshot(mp.Page, "Member_Edit_Request_Form_Filled")

	// Submit
	if err := mp.Page.ClickByTestID("testid-profile-edit-submit-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	time.Sleep(2 * time.Second) // wait for modal to close
}

func MemberSubmitsGrievance(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Member Submits Grievance"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		return
	}
	mp := mai.GetMemberPersona()

	// Navigate to Grievance Page
	if err := mp.Page.ClickByTestID("testid-grievance-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	time.Sleep(1 * time.Second)

	// Fill subject
	if err := mp.Page.SendKeysByTestID("testid-grievance-subject-input", "E2E Test Grievance Subject", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Wait for category select to be clickable (finish loading dropdowns from API)
	if _, err := mp.Page.WaitUntilClickable("css:[data-testid='testid-grievance-category-select']", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Select Category
	if err := mp.Page.SelectCustomDropdownByText("testid-grievance-category-select", "Pay & Allowances", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Select Priority
	if err := mp.Page.SelectCustomDropdownByText("testid-grievance-priority-select", "High", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Fill phone
	if err := mp.Page.SendKeysByTestID("testid-grievance-contact-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Fill description (must be > 50 characters)
	if err := mp.Page.SendKeysByTestID("testid-grievance-description-input", "This is a detailed description of the grievance submitted by E2E test. It has to be more than fifty characters long to pass frontend validation.", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.CaptureScreenshot(mp.Page, "Member_Grievance_Form_Filled")

	// Submit
	if err := mp.Page.ClickByTestID("testid-submit-grievance-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	// Wait for success message
	if _, err := mp.Page.WaitUntilVisible("css:[data-testid='testid-success-message']", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		mai.GetMemberPersona().Page.CaptureScreenshot("Failure_" + actionName)
		return
	}

	r.CaptureScreenshot(mp.Page, "Member_Grievance_Submitted_Successfully")
	time.Sleep(2 * time.Second)
}
