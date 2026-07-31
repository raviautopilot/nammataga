package actions

import (
	"fmt"
	"time"

	"e2e-template/pkg/ui"
)

// PublicActions defines actions a general public user can perform on the UI.
type PublicActions struct {
	*ui.Page
}

// NewPublicActions initializes a new PublicActions wrapper and validates it.
func NewPublicActions(page *ui.Page) (*PublicActions, error) {
	if page == nil {
		return nil, fmt.Errorf("provided page cannot be nil")
	}
	if page.Driver == nil {
		return nil, fmt.Errorf("provided page's selenium WebDriver cannot be nil")
	}
	return &PublicActions{Page: page}, nil
}

// ensurePage verifies the wrapper and its internal fields are not nil.
func (a *PublicActions) ensurePage() error {
	if a == nil || a.Page == nil {
		return fmt.Errorf("PublicActions or Page is not initialized")
	}
	if a.Page.Driver == nil {
		return fmt.Errorf("WebDriver is not initialized")
	}
	return nil
}

// GoToHome navigates to the configured target URL (Home page).
func (a *PublicActions) GoToHome(targetURL string) error {
	if err := a.ensurePage(); err != nil {
		return err
	}
	return a.Page.GoToHome(targetURL)
}

// GoToOfficeBeaers clicks the Office Bearers link in the navigation menu.
func (a *PublicActions) GoToOfficeBeaers(timeout time.Duration) error {
	if err := a.ensurePage(); err != nil {
		return err
	}
	return a.Page.ClickByTestID("testid-office-bearers-button", timeout)
}

// GoToEvents clicks the Events link in the navigation menu.
func (a *PublicActions) GoToEvents(timeout time.Duration) error {
	if err := a.ensurePage(); err != nil {
		return err
	}
	return a.Page.ClickByTestID("testid-events-button", timeout)
}

// GoToMemberLogin clicks the Member Login link in the navigation menu.
func (a *PublicActions) GoToMemberLogin(timeout time.Duration) error {
	if err := a.ensurePage(); err != nil {
		return err
	}
	return a.Page.ClickByTestID("testid-member-login-button", timeout)
}

// GoToAdminLogin clicks the Administrative Access link in the footer.
func (a *PublicActions) GoToAdminLogin(timeout time.Duration) error {
	if err := a.ensurePage(); err != nil {
		return err
	}
	return a.Page.ClickByTestID("testid-admin-login-button", timeout)
}

// ==================== PERSONA WRAPPERS ====================

// MemberActions embeds PublicActions to inherit all public navigation,
// and adds member-specific actions.
type MemberActions struct {
	*PublicActions
}

// NewMemberActions initializes a new MemberActions wrapper.
func NewMemberActions(page *ui.Page) (*MemberActions, error) {
	pub, err := NewPublicActions(page)
	if err != nil {
		return nil, err
	}
	return &MemberActions{PublicActions: pub}, nil
}

// SubscriberActions embeds MemberActions to inherit all member and public navigation,
// and adds subscriber-specific actions.
type SubscriberActions struct {
	*MemberActions
}

// NewSubscriberActions initializes a new SubscriberActions wrapper.
func NewSubscriberActions(page *ui.Page) (*SubscriberActions, error) {
	mem, err := NewMemberActions(page)
	if err != nil {
		return nil, err
	}
	return &SubscriberActions{MemberActions: mem}, nil
}

// AdminActions embeds PublicActions to inherit all public navigation,
// and adds admin-specific actions.
type AdminActions struct {
	*PublicActions
}

// NewAdminActions initializes a new AdminActions wrapper.
func NewAdminActions(page *ui.Page) (*AdminActions, error) {
	pub, err := NewPublicActions(page)
	if err != nil {
		return nil, err
	}
	return &AdminActions{PublicActions: pub}, nil
}
