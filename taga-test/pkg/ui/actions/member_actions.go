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

	if scr, scrErr := mp.Page.CaptureScreenshot("Step_03_MemberLogout_HomePage"); scrErr == nil {
		r.Evidence = append(r.Evidence, scr)
	}
	r.Advice = append(r.Advice, "Member logged out successfully.")
}

