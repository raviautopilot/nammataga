package pages

import (
	"time"

	"e2e-template/pkg/ui"
)

// LoginPage handles interactions with login forms (Admin and Member modals/pages).
type LoginPage struct {
	*ui.Page
}

// NewLoginPage creates a new LoginPage instance wrapping the base Page.
func NewLoginPage(page *ui.Page) *LoginPage {
	return &LoginPage{Page: page}
}

// VerifyFormElements verifies that all required data-testid elements exist on the login form.
func (l *LoginPage) VerifyFormElements(testIDs []string, timeout time.Duration) error {
	return l.VerifyElementsPresentByTestIDs(testIDs, timeout)
}

// EnterUsername types the username into the specified input field.
func (l *LoginPage) EnterUsername(inputTestID, username string, timeout time.Duration) error {
	return l.SendKeysByTestID(inputTestID, username, timeout)
}

// EnterPassword types the password into the specified input field.
func (l *LoginPage) EnterPassword(inputTestID, password string, timeout time.Duration) error {
	return l.SendKeysByTestID(inputTestID, password, timeout)
}

// SubmitLogin clicks the submit button for the login form.
func (l *LoginPage) SubmitLogin(submitTestID string, timeout time.Duration) error {
	return l.ClickByTestID(submitTestID, timeout)
}

// FillAndSubmitLogin enters credentials and clicks submit in one step.
func (l *LoginPage) FillAndSubmitLogin(usernameTestID, passwordTestID, submitTestID, username, password string, timeout time.Duration) error {
	if err := l.EnterUsername(usernameTestID, username, timeout); err != nil {
		return err
	}
	if err := l.EnterPassword(passwordTestID, password, timeout); err != nil {
		return err
	}
	return l.SubmitLogin(submitTestID, timeout)
}
