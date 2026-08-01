package actions

import (
	"fmt"
	"time"

	"e2e-template/pkg/ui"
)

// Result captures test journey execution details, actions, evidence, and advice.
type Result struct {
	TestName string
	Actions  []string
	Evidence []string // File paths of captured screenshots
	Advice   []string // Suggestive messages or remediation advice
	Status   string   // "passed" or "failed"
	Error    error    // The last encountered error, if any
}

// NewResult initializes a test journey result.
func NewResult(testName string) *Result {
	return &Result{
		TestName: testName,
		Status:   "passed",
		Actions:  make([]string, 0),
		Evidence: make([]string, 0),
		Advice:   make([]string, 0),
	}
}

// Failed returns true if any action in the journey failed.
func (r *Result) Failed() bool {
	return r.Status == "failed"
}

// PublicActionsInterface allows different personas to share public actions.
type PublicActionsInterface interface {
	GetPublicPersona() *PublicPersona
}

// PublicPersona defines the credentials and capabilities of a general visitor.
type PublicPersona struct {
	*ui.Page
	BaseURL        string
	DefaultTimeout time.Duration
}

// NewPublicPersona creates a new PublicPersona.
func NewPublicPersona(page *ui.Page, baseURL string, defaultTimeout time.Duration) *PublicPersona {
	return &PublicPersona{
		Page:           page,
		BaseURL:        baseURL,
		DefaultTimeout: defaultTimeout,
	}
}

// GetPublicPersona implements PublicActionsInterface for PublicPersona.
func (p *PublicPersona) GetPublicPersona() *PublicPersona {
	return p
}

// ensurePage verifies the wrapper and its internal fields are not nil.
func ensurePage(p *PublicPersona) error {
	if p == nil || p.Page == nil {
		return fmt.Errorf("PublicPersona or Page is not initialized")
	}
	if p.Page.Driver == nil {
		return fmt.Errorf("WebDriver is not initialized")
	}
	return nil
}

// GoToHome navigates to the Home page and records the outcome.
func GoToHome(pai PublicActionsInterface, r *Result) {
	actionName := "Navigate to Home Page"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	p := pai.GetPublicPersona()
	if err := ensurePage(p); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Prerequisite Check: Ensure the Selenium driver is started and passed correctly to the persona")
		return
	}

	err := p.Page.GoToHome(p.BaseURL)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, fmt.Sprintf("Advice: Check if the application server is running at %s. Ensure the URL is reachable.", p.BaseURL))
		if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToHome_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToHome_Success"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Home page loaded successfully.")
}

// GoToOfficeBeaers navigates to the Office Bearers page and records the outcome.
func GoToOfficeBeaers(pai PublicActionsInterface, r *Result) {
	actionName := "Navigate to Office Bearers Page"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	p := pai.GetPublicPersona()
	if err := ensurePage(p); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Prerequisite Check: Ensure the Selenium driver is started and passed correctly to the persona")
		return
	}

	err := p.Page.ClickByTestID("testid-office-bearers-button", p.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify that the 'testid-office-bearers-button' element exists on the page and is clickable.")
		if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToOfficeBeaers_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToOfficeBeaers_Success"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Office Bearers page loaded successfully.")
}

// GoToEvents navigates to the Events page and records the outcome.
func GoToEvents(pai PublicActionsInterface, r *Result) {
	actionName := "Navigate to Events Page"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	p := pai.GetPublicPersona()
	if err := ensurePage(p); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Prerequisite Check: Ensure the Selenium driver is started and passed correctly to the persona")
		return
	}

	err := p.Page.ClickByTestID("testid-events-button", p.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify that the 'testid-events-button' element exists on the page and is clickable.")
		if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToEvents_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToEvents_Success"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Events page loaded successfully.")
}

// GoToMemberLogin navigates to the Member Login page and records the outcome.
func GoToMemberLogin(pai PublicActionsInterface, r *Result) {
	actionName := "Navigate to Member Login Page"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	p := pai.GetPublicPersona()
	if err := ensurePage(p); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Prerequisite Check: Ensure the Selenium driver is started and passed correctly to the persona")
		return
	}

	err := p.Page.ClickByTestID("testid-member-login-button", p.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify that the 'testid-member-login-button' element exists on the page and is clickable.")
		if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToMemberLogin_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToMemberLogin_Success"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Member Login page loaded successfully.")
}

// GoToAdminLogin navigates to the Admin Login page and records the outcome.
func GoToAdminLogin(pai PublicActionsInterface, r *Result) {
	actionName := "Navigate to Admin Login Page"
	r.Actions = append(r.Actions, actionName)
	if r.Failed() {
		r.Advice = append(r.Advice, fmt.Sprintf("Skipped '%s' because a previous step failed", actionName))
		return
	}

	p := pai.GetPublicPersona()
	if err := ensurePage(p); err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Prerequisite Check: Ensure the Selenium driver is started and passed correctly to the persona")
		return
	}

	err := p.Page.ClickByTestID("testid-admin-login-button", p.DefaultTimeout)
	if err != nil {
		r.Status = "failed"
		r.Error = err
		r.Advice = append(r.Advice, "Advice: Verify that the 'testid-admin-login-button' element exists in the footer and is clickable.")
		if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToAdminLogin_Failure"); scrErr == nil {
			r.Evidence = append(r.Evidence, scr)
		}
		return
	}

	if scr, scrErr := p.Page.CaptureScreenshot(r.TestName + "_GoToAdminLogin_Success"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Admin Login page loaded successfully.")
}

// ==================== PERSONA WRAPPERS ====================

// MemberPersona embeds PublicPersona to inherit all public capabilities.
type MemberPersona struct {
	*PublicPersona
}

// NewMemberPersona creates a new MemberPersona.
func NewMemberPersona(page *ui.Page, baseURL string, defaultTimeout time.Duration) *MemberPersona {
	return &MemberPersona{
		PublicPersona: NewPublicPersona(page, baseURL, defaultTimeout),
	}
}

// SubscriberPersona embeds MemberPersona.
type SubscriberPersona struct {
	*MemberPersona
}

// NewSubscriberPersona creates a new SubscriberPersona.
func NewSubscriberPersona(page *ui.Page, baseURL string, defaultTimeout time.Duration) *SubscriberPersona {
	return &SubscriberPersona{
		MemberPersona: NewMemberPersona(page, baseURL, defaultTimeout),
	}
}

// AdminPersona embeds PublicPersona.
type AdminPersona struct {
	*PublicPersona
}

// NewAdminPersona creates a new AdminPersona.
func NewAdminPersona(page *ui.Page, baseURL string, defaultTimeout time.Duration) *AdminPersona {
	return &AdminPersona{
		PublicPersona: NewPublicPersona(page, baseURL, defaultTimeout),
	}
}
