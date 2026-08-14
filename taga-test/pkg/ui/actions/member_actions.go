package actions

import (
	"fmt"
	"time"

	"e2e-template/pkg/config"
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
		return
	}

	time.Sleep(waitTime)

	r.CaptureScreenshot(mp.Page, screenshotName)
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
		return
	}

	time.Sleep(waitTime)
	_ = homePage.GoToHome(mp.BaseURL)
	time.Sleep(waitTime)

	r.CaptureScreenshot(mp.Page, screenshotName)
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
	time.Sleep(1 * time.Second)
	r.CaptureScreenshot(mp.Page, "MemberLogin_Filled_Form")

	if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to click submit login button")
		return
	}

	// Wait for dashboard to fully render after login redirect
	time.Sleep(3 * time.Second)
	r.CaptureScreenshot(mp.Page, "MemberLogin_Dashboard")
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
		return
	}

	// Wait for change password dialog to appear
	time.Sleep(1 * time.Second)

	if err := mp.Page.SendKeysByTestID("testid-change-password-email-input", email, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-change-password-old-input", tempPassword, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-change-password-new-input", newPassword, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-change-password-confirm-input", newPassword, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}

	r.CaptureScreenshot(mp.Page, "ChangePassword_Filled")

	if err := mp.Page.ClickByTestID("testid-change-password-submit-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}

	time.Sleep(3 * time.Second)
	r.CaptureScreenshot(mp.Page, "ChangePassword_Success")
	r.Advice = append(r.Advice, "Password changed successfully.")
}

// VisitAllMemberPages navigates through all member-accessible pages and captures screenshots.
func VisitAllMemberPages(mai MemberActionsInterface, cfg *config.Config, r *Result) {
	actionName := "Visit All Member Pages"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	// Define navigation sequence
	pages := []struct {
		TestID         string
		ScreenshotName string
		WaitTime       time.Duration
	}{
		{"testid-home-button", "Home_Page", 2 * time.Second},
		{"testid-office-bearers-button", "Office_Bearers_Page", 2 * time.Second},
		{"testid-resources-button", "Resources_Page", 2 * time.Second},
		{"testid-events-button", "Events_Gallery_Page", 2 * time.Second},
		{"testid-upcoming-events-button", "Upcoming_Events_Page", 2 * time.Second}, // Sub-tab of Events
		{"testid-taga-towers-button", "TAGATowers_Page", 2 * time.Second},
		{"testid-grievance-button", "Grievance_Page", 2 * time.Second},
		{"testid-membership-button", "Member_Profile_Page", 2 * time.Second},          // Membership default is Profile
		{"testid-member-subscriptions-button", "Subscriptions_Page", 2 * time.Second}, // Sub-tab of Membership
		{"testid-member-announcements-button", "Announcements_Page", 2 * time.Second}, // Sub-tab of Membership
	}

	for _, page := range pages {
		if err := mp.Page.ClickByTestID(page.TestID, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify '%s' element exists", page.TestID))
			return
		}

		time.Sleep(page.WaitTime)

		r.CaptureScreenshot(mp.Page, page.ScreenshotName)
	}

	r.Advice = append(r.Advice, "Successfully visited all member pages.")
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
	time.Sleep(1 * time.Second)

	r.CaptureScreenshot(mp.Page, "Subscription_Tab_Opened")

	// 3. Click 'Pay Now' for Annual Subscription
	if err := mp.Page.ClickByTestID("testid-pay-now-annual-subscription-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-pay-now-annual-subscription-button' exists. It might already be paid.")
		return
	}
	time.Sleep(1 * time.Second)

	r.CaptureScreenshot(mp.Page, "Payment_Modal_Opened")

	// 4. Inject Mock Razorpay
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_test_123",
				razorpay_payment_id: "pay_mock_test_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	if _, err := mp.Page.Driver.ExecuteScript(mockScript, nil); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to inject mock Razorpay script")
		return
	}

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
	time.Sleep(2 * time.Second)

	r.CaptureScreenshot(mp.Page, "MemberLogout_HomePage")
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
	time.Sleep(2 * time.Second)

	r.CaptureScreenshot(mp.Page, "TAGATower_Dashboard")
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
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_123",
				razorpay_payment_id: "pay_mock_tower_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify payment button exists")
		return
	}
	time.Sleep(3 * time.Second)

	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Booking_Complete_%s", roomID))
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
	time.Sleep(3 * time.Second)

	r.CaptureScreenshot(mp.Page, "TAGATower_Booking_Cancelled")
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
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-guest-1-age-input", guestAge, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	if err := mp.Page.SendKeysByTestID("testid-guest-1-contact-input", guestContact, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_GuestBooking_Modal_%s", roomID))

	// Mocking Razorpay
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_guest_123",
				razorpay_payment_id: "pay_mock_tower_guest_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify payment button exists")
		return
	}
	time.Sleep(3 * time.Second)

	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_GuestBooking_Complete_%s", roomID))
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
		return
	}

	// Select "Guest" / "Others"
	if _, err := mp.Page.Driver.ExecuteScript("document.getElementById('guest').click();", nil); err != nil {
		r.Status = "failed"
		r.Error = err
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
			bedText := fmt.Sprintf("%d bed", capacity)
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

		// Click "Add Guest" if we still have more guests to fill
		if idx < capacity {
			if err := mp.Page.ClickByTestID("testid-add-guest-button", mp.DefaultTimeout); err != nil {
				r.Status = "failed"
				r.Error = fmt.Errorf("failed to click testid-add-guest-button for guest %d: %v", idx+1, err)
				return
			}
			time.Sleep(500 * time.Millisecond) // wait for new inputs to render
		}
	}

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_AllRoomBooking_Modal_%s", roomID))

	// Mocking Razorpay
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_allroom_123",
				razorpay_payment_id: "pay_mock_tower_allroom_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	time.Sleep(3 * time.Second)

	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_AllRoomBooking_Complete_%s", roomID))
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
		return
	}

	// Select "Self" explicitly
	if _, err := mp.Page.Driver.ExecuteScript("document.getElementById('self').click();", nil); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	time.Sleep(1 * time.Second)

	// Check if Bed Count Select is visible in the UI
	checkSelectScript := `return !!document.querySelector('[data-testid="testid-bed-count-select"]');`
	hasSelectObj, err := mp.Page.Driver.ExecuteScript(checkSelectScript, nil)
	if err != nil || hasSelectObj != true {
		r.Status = "failed"
		r.Error = fmt.Errorf("bed count select not visible or room does not support single bed bookings")
		return
	}

	// Try to click Bed Count Select
	if err := mp.Page.ClickByTestID("testid-bed-count-select", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to click bed count select: %v", err)
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
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_self_multi_123",
				razorpay_payment_id: "pay_mock_tower_self_multi_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed to payment
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "System prevented proceeding to payment for self multibooking.")
		return
	}
	time.Sleep(3 * time.Second)

	r.CaptureScreenshot(mp.Page, "TAGATower_SelfMultibooking_Allowed_Failure")

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
			time.Sleep(3 * time.Second)
			r.CaptureScreenshot(mp.Page, "TAGATower_SelfMultibooking_Cancelled_Cleanup")
			r.Advice = append(r.Advice, "Cleanup: Cancelled the allowed self-multibooking.")
		}
	}

	// Mark the result as failed
	r.Status = "failed"
	r.Error = fmt.Errorf("security vulnerability: member was allowed to book %d beds for Self (multibooking). Only guest bookings should allow multibed bookings", beds)
	mp.Page.LastError = r.Error
	r.Advice = append(r.Advice, "Advice: Update frontend and backend validation to disallow BedCount > 1 when booking for Self")
}

// SelectFutureDates clicks the next month button on the calendar and selects a free date range.
func SelectFutureDates(mai MemberActionsInterface, r *Result) {
	actionName := "Select Future Dates in Calendar"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	mp := mai.GetMemberPersona()

	selectFutureDatesScript := `
	// Click Next Month button
	const nextBtn = document.querySelector('button.rdp-nav_button_next') || 
	                document.querySelector('button[name="next-button"]') ||
	                document.querySelector('.rdp-nav_button_next button');
	if (nextBtn) {
		nextBtn.click();
		return true;
	}
	return false;
	`

	clicked, err := mp.Page.Driver.ExecuteScript(selectFutureDatesScript, nil)
	if err != nil || clicked == false {
		r.Advice = append(r.Advice, "Note: Next month button not found or clicked, proceeding with default dates")
		return
	}
	time.Sleep(1 * time.Second)

	// Select two active days in the next month
	clickDaysScript := `
	const days = Array.from(document.querySelectorAll('button.rdp-day')).filter(btn => {
		return !btn.disabled && 
		       !btn.classList.contains('rdp-day_outside') && 
		       btn.getAttribute('aria-disabled') !== 'true';
	});
	if (days.length >= 5) {
		days[2].click();
		setTimeout(() => {
			days[3].click();
		}, 300);
		return true;
	}
	return false;
	`
	selectedDays, err := mp.Page.Driver.ExecuteScript(clickDaysScript, nil)
	if err != nil || selectedDays == false {
		r.Advice = append(r.Advice, "Note: Could not select custom future dates, using default dates")
	} else {
		time.Sleep(2 * time.Second) // wait for availability to refresh
	}
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
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_gender_123",
				razorpay_payment_id: "pay_mock_tower_gender_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	time.Sleep(3 * time.Second)
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

	// Mocking Razorpay
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_gender_opp_123",
				razorpay_payment_id: "pay_mock_tower_gender_opp_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		// If it failed to click (or showed validation error), that is expected behavior!
		r.Advice = append(r.Advice, "System prevented proceeding to payment for opposite gender booking.")
		return
	}
	time.Sleep(3 * time.Second)

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
	r.Advice = append(r.Advice, "Advice: Update frontend to send gender for all single-bed rooms and ensure backend checks are enforced")
}

// TryBookDormitoryWithOppositeGender tries to book a bed in a gender-specific dormitory (gents-dorm or ladies-dorm)
// with the prohibited gender, expecting it to be disallowed.
func TryBookDormitoryWithOppositeGender(mai MemberActionsInterface, cfg *config.Config, r *Result, roomID string, prohibitedGender string) {
	actionName := fmt.Sprintf("Try Book Dormitory %s with Prohibited Gender %s", roomID, prohibitedGender)
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

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
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

	// Click the prohibited gender radio button
	genderScript := fmt.Sprintf("document.getElementById('%s').click();", prohibitedGender)
	if _, err := mp.Page.Driver.ExecuteScript(genderScript, nil); err != nil {
		r.Status = "failed"
		r.Error = fmt.Errorf("failed to select prohibited gender %s: %v", prohibitedGender, err)
		return
	}
	time.Sleep(1 * time.Second)

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Dormitory_%s_%s_Attempt", roomID, prohibitedGender))

	// Mocking Razorpay
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_dorm_opp_123",
				razorpay_payment_id: "pay_mock_dorm_opp_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed to payment
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "System prevented proceeding to payment for prohibited gender booking in dormitory.")
		return
	}
	time.Sleep(3 * time.Second)

	// Capture modal state of the allowed (but should be disallowed) booking
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Dormitory_%s_%s_Allowed_Failure", roomID, prohibitedGender))

	// Clean up the booking immediately
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

	// Mark test as failed since prohibited gender booking was allowed
	r.Status = "failed"
	r.Error = fmt.Errorf("security vulnerability: dormitory %s allowed booking for prohibited gender %s", roomID, prohibitedGender)
	mp.Page.LastError = r.Error
	r.Advice = append(r.Advice, fmt.Sprintf("Advice: Enforce gender checks in frontend and backend for gender-specific dormitory: %s", roomID))
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
		return
	}

	// Capture modal state
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_10DaysBooking_Modal_%s", roomID))

	// Mocking Razorpay
	mockScript := `
	window.Razorpay = function(options) {
		this.open = function() {
			options.handler({
				razorpay_order_id: "mock_order_tower_10days_123",
				razorpay_payment_id: "pay_mock_tower_10days_123",
				razorpay_signature: "mock_signature"
			});
		};
	};`
	mp.Page.Driver.ExecuteScript(mockScript, nil)

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		return
	}
	time.Sleep(3 * time.Second)

	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_10DaysBooking_Complete_%s", roomID))
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
		r.Advice = append(r.Advice, "System correctly blocked overlapping booking (room button is disabled/unavailable).")
		return
	}

	// Capture modal state of the allowed (but should be disallowed) booking
	r.CaptureScreenshot(mp.Page, fmt.Sprintf("TAGATower_Overlapping_Allowed_Failure_%s", roomID))

	// Try clicking Book just to see if we can open payment modal
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
		r.Advice = append(r.Advice, "System correctly blocked overlapping booking (could not click book button).")
		return
	}
	time.Sleep(2 * time.Second)

	// Fill Modal: Booker Phone Number
	if err := mp.Page.SendKeysByTestID("testid-booker-phone-input", "9876543210", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "System correctly prevented proceeding with overlapping booking.")
		return
	}

	// Click proceed
	if err := mp.Page.ClickByTestID("testid-booking-proceed-payment-button", mp.DefaultTimeout); err != nil {
		r.Advice = append(r.Advice, "System correctly blocked proceeding to payment for overlapping booking.")
		return
	}
	time.Sleep(3 * time.Second)

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
	time.Sleep(1 * time.Second)

	r.CaptureScreenshot(mp.Page, screenshotName)
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

	// 1. Verify Home is accessible
	_ = mp.Page.GoToHome(mp.BaseURL)
	time.Sleep(1 * time.Second)
	r.CaptureScreenshot(mp.Page, "Access_Home")

	// 2. Verify Office Bearers is accessible
	if err := mp.Page.ClickByTestID("testid-office-bearers-button", mp.DefaultTimeout); err == nil {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, "Access_OfficeBearers")
	} else {
		r.Status = "failed"
		r.Error = fmt.Errorf("office bearers should be accessible: %v", err)
		return
	}

	// 3. Verify Events is accessible
	if err := mp.Page.ClickByTestID("testid-events-button", mp.DefaultTimeout); err == nil {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, "Access_Events")
	} else {
		r.Status = "failed"
		r.Error = fmt.Errorf("events should be accessible: %v", err)
		return
	}

	// 4. Verify Profile (Subscriptions/Announcements) is accessible
	if err := mp.Page.ClickByTestID("testid-membership-button", mp.DefaultTimeout); err == nil {
		time.Sleep(1 * time.Second)
		r.CaptureScreenshot(mp.Page, "Access_Profile")
		// Check Subscriptions tab
		_ = mp.Page.ClickByTestID("testid-member-subscriptions-button", mp.DefaultTimeout)
		time.Sleep(1 * time.Second)
	} else {
		r.Status = "failed"
		r.Error = fmt.Errorf("profile should be accessible: %v", err)
		return
	}

	// 5. Verify Non-accessible pages are not in DOM
	nonAccessible := []string{
		"testid-resources-button",
		"testid-taga-towers-button",
		"testid-grievance-button",
		"testid-members-button",
	}

	for _, testID := range nonAccessible {
		_, err := mp.Page.WaitUntilVisible(fmt.Sprintf("css:[data-testid='%s']", testID), 1*time.Second)
		if err == nil {
			r.Status = "failed"
			r.Error = fmt.Errorf("page link %s should not be accessible for unpaid member", testID)
			return
		}
	}
	r.Advice = append(r.Advice, "Successfully validated unpaid member access constraints.")
}
