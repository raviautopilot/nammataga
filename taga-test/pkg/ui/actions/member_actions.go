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

	if scr, scrErr := mp.Page.CaptureScreenshot(screenshotName); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
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

	if scr, scrErr := mp.Page.CaptureScreenshot(screenshotName); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
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
	if scr, scrErr := mp.Page.CaptureScreenshot("Step_01_MemberLogin_Filled_Form"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	if err := loginPage.SubmitLogin(cfg.MemberLoginSubmitButtonTestID, mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Failed to click submit login button")
		return
	}

	// Wait for dashboard to fully render after login redirect
	time.Sleep(3 * time.Second)
	if scr, scrErr := mp.Page.CaptureScreenshot("Step_02_MemberLogin_Dashboard"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Member logged in successfully.")
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
		{"testid-home-button", "Step_03_Home_Page", 2 * time.Second},
		{"testid-office-bearers-button", "Step_04_Office_Bearers_Page", 2 * time.Second},
		{"testid-resources-button", "Step_05_Resources_Page", 2 * time.Second},
		{"testid-events-button", "Step_06_Events_Gallery_Page", 2 * time.Second},
		{"testid-upcoming-events-button", "Step_07_Upcoming_Events_Page", 2 * time.Second}, // Sub-tab of Events
		{"testid-taga-towers-button", "Step_08_TAGATowers_Page", 2 * time.Second},
		{"testid-grievance-button", "Step_09_Grievance_Page", 2 * time.Second},
		{"testid-membership-button", "Step_10_Member_Profile_Page", 2 * time.Second}, // Membership default is Profile
		{"testid-member-subscriptions-button", "Step_11_Subscriptions_Page", 2 * time.Second}, // Sub-tab of Membership
		{"testid-member-announcements-button", "Step_12_Announcements_Page", 2 * time.Second}, // Sub-tab of Membership
	}

	for _, page := range pages {
		if err := mp.Page.ClickByTestID(page.TestID, mp.DefaultTimeout); err != nil {
			r.Status = "failed"
			r.Error = err
			r.Advice = append(r.Advice, fmt.Sprintf("Advice: Verify '%s' element exists", page.TestID))
			return
		}
		
		time.Sleep(page.WaitTime)
		
		if scr, scrErr := mp.Page.CaptureScreenshot(page.ScreenshotName); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
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
	
	if scr, scrErr := mp.Page.CaptureScreenshot("Step_03_Subscription_Tab_Opened"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

	// 3. Click 'Pay Now' for Annual Subscription
	if err := mp.Page.ClickByTestID("testid-pay-now-annual-subscription-button", mp.DefaultTimeout); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify 'testid-pay-now-annual-subscription-button' exists. It might already be paid.")
		return
	}
	time.Sleep(1 * time.Second)
	
	if scr, scrErr := mp.Page.CaptureScreenshot("Step_04_Payment_Modal_Opened"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

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

	if scr, scrErr := mp.Page.CaptureScreenshot("Step_05_Annual_Subscription_Paid"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}

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

	if scr, scrErr := mp.Page.CaptureScreenshot("Step_13_MemberLogout_HomePage"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Member logged out successfully.")
}

