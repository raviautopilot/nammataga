package pages

import (
	"time"

	"github.com/tebeka/selenium"
	"e2e-template/pkg/ui"
)

// LoginPage implements the Page Object Model pattern for a standard Login page.
type LoginPage struct {
	*ui.Page
	UsernameField string
	PasswordField string
	SubmitButton  string
	ErrorMessage  string
}

// NewLoginPage creates a page object with pre-defined selectors.
func NewLoginPage(driver selenium.WebDriver, screenshotDir string) *LoginPage {
	return &LoginPage{
		Page:          ui.NewPage(driver, screenshotDir),
		UsernameField: "css:#username",
		PasswordField: "css:#password",
		SubmitButton:  "css:button[type='submit']",
		ErrorMessage:  "css:.error-message",
	}
}

// Login performs typing user credentials and clicking the submit button.
func (l *LoginPage) Login(username, password string, timeout time.Duration) error {
	if err := l.SendKeys(l.UsernameField, username, timeout); err != nil {
		return err
	}
	if err := l.SendKeys(l.PasswordField, password, timeout); err != nil {
		return err
	}
	return l.Click(l.SubmitButton, timeout)
}

// GetAlertText extracts any login validation error text display.
func (l *LoginPage) GetAlertText(timeout time.Duration) (string, error) {
	return l.GetText(l.ErrorMessage, timeout)
}
