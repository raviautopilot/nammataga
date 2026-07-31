package pages

import (
	"time"

	"e2e-template/pkg/ui"
)

// HomePage handles high-level navigation for the home/landing page.
type HomePage struct {
	*ui.Page
}

// NewHomePage creates a new HomePage instance.
func NewHomePage(page *ui.Page) *HomePage {
	return &HomePage{Page: page}
}

// GoToHome navigates to the root application URL.
func (h *HomePage) GoToHome(targetURL string) error {
	return h.GoToHomePage(targetURL)
}

// OpenAdminLogin clicks the admin login button on the home page.
func (h *HomePage) OpenAdminLogin(btnTestID string, timeout time.Duration) error {
	return h.ClickByTestID(btnTestID, timeout)
}

// OpenMemberLogin clicks the member login button on the home page.
func (h *HomePage) OpenMemberLogin(btnTestID string, timeout time.Duration) error {
	return h.ClickByTestID(btnTestID, timeout)
}

// Logout clicks the logout button on the home page to end the current session.
func (h *HomePage) Logout(logoutTestID string, timeout time.Duration) error {
	return h.ClickByTestID(logoutTestID, timeout)
}
