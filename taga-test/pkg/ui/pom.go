package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"e2e-template/pkg/logger"
)

// Page serves as the base Page Object wrapper.
type Page struct {
	Driver        selenium.WebDriver
	ScreenshotDir string
	LastError     error
}

// NewPage initializes a base Page.
func NewPage(driver selenium.WebDriver, screenshotDir string) *Page {
	return &Page{Driver: driver, ScreenshotDir: screenshotDir}
}

// GoToHomePage navigates the browser to the root application URL specified in targetURL.
func (p *Page) GoToHomePage(targetURL string) error {
	if targetURL == "" {
		return fmt.Errorf("target URL is empty")
	}
	if err := p.Driver.Get(targetURL); err != nil {
		p.LastError = err
		return fmt.Errorf("failed to navigate to home page (%s): %w", targetURL, err)
	}
	p.InjectNetworkInterceptor()
	return nil
}

// GoToHome navigates to the configured UiURL.
func (p *Page) GoToHome(targetURL string) error {
	return p.GoToHomePage(targetURL)
}

// OpenAdminLogin clicks the admin login button.
func (p *Page) OpenAdminLogin(btnTestID string, timeout time.Duration) error {
	return p.ClickByTestID(btnTestID, timeout)
}

// OpenMemberLogin clicks the member login button.
func (p *Page) OpenMemberLogin(btnTestID string, timeout time.Duration) error {
	return p.ClickByTestID(btnTestID, timeout)
}

// VerifyFormElements verifies present testID elements.
func (p *Page) VerifyFormElements(testIDs []string, timeout time.Duration) error {
	return p.VerifyElementsPresentByTestIDs(testIDs, timeout)
}

// EnterUsername types the username into the specified input testID field.
func (p *Page) EnterUsername(inputTestID, username string, timeout time.Duration) error {
	return p.SendKeysByTestID(inputTestID, username, timeout)
}

// EnterPassword types the password into the specified input testID field.
func (p *Page) EnterPassword(inputTestID, password string, timeout time.Duration) error {
	return p.SendKeysByTestID(inputTestID, password, timeout)
}

// SubmitLogin clicks the submit button for the specified testID.
func (p *Page) SubmitLogin(submitTestID string, timeout time.Duration) error {
	return p.ClickByTestID(submitTestID, timeout)
}

// FillAndSubmitLogin enters credentials and clicks submit in one call.
func (p *Page) FillAndSubmitLogin(usernameTestID, passwordTestID, submitTestID, username, password string, timeout time.Duration) error {
	if err := p.EnterUsername(usernameTestID, username, timeout); err != nil {
		return err
	}
	if err := p.EnterPassword(passwordTestID, password, timeout); err != nil {
		return err
	}
	return p.SubmitLogin(submitTestID, timeout)
}

// Logout clicks the logout button to end the current session.
func (p *Page) Logout(logoutTestID string, timeout time.Duration) error {
	return p.ClickByTestID(logoutTestID, timeout)
}

// parseLocator routes selectors by prefix (e.g. xpath://... or css:#...) or defaults to CSS.
func parseLocator(locator string) (string, string) {
	if strings.HasPrefix(locator, "xpath:") {
		return selenium.ByXPATH, locator[6:]
	}
	if strings.HasPrefix(locator, "css:") {
		return selenium.ByCSSSelector, locator[4:]
	}
	if strings.HasPrefix(locator, "//") {
		return selenium.ByXPATH, locator
	}
	return selenium.ByCSSSelector, locator
}

// WaitUntilVisible blocks until the element is located and visible.
func (p *Page) WaitUntilVisible(locator string, timeout time.Duration) (selenium.WebElement, error) {
	by, val := parseLocator(locator)
	var elem selenium.WebElement
	err := p.Driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		el, err := wd.FindElement(by, val)
		if err != nil {
			return false, nil
		}
		disp, err := el.IsDisplayed()
		if err != nil {
			return false, nil
		}
		if disp {
			elem = el
			return true, nil
		}
		return false, nil
	}, timeout)

	if err != nil {
		wrappedErr := fmt.Errorf("element '%s' was not visible after %v: %w", locator, timeout, err)
		p.LastError = wrappedErr
		return nil, wrappedErr
	}
	return elem, nil
}

// WaitUntilClickable blocks until the element is located, visible, and enabled.
func (p *Page) WaitUntilClickable(locator string, timeout time.Duration) (selenium.WebElement, error) {
	by, val := parseLocator(locator)
	var elem selenium.WebElement
	err := p.Driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		el, err := wd.FindElement(by, val)
		if err != nil {
			return false, nil
		}
		disp, err := el.IsDisplayed()
		if err != nil || !disp {
			return false, nil
		}
		enabled, err := el.IsEnabled()
		if err != nil || !enabled {
			return false, nil
		}
		elem = el
		return true, nil
	}, timeout)

	if err != nil {
		wrappedErr := fmt.Errorf("element '%s' was not clickable after %v: %w", locator, timeout, err)
		p.LastError = wrappedErr
		return nil, wrappedErr
	}
	return elem, nil
}

// Click waits for the element to be clickable and performs a click.
func (p *Page) Click(locator string, timeout time.Duration) error {
	el, err := p.WaitUntilClickable(locator, timeout)
	if err != nil {
		p.LastError = err
		return err
	}
	err = el.Click()
	if err != nil {
		// Fallback for element click interception / overlapping navbar: scroll into view and click via JS
		_, jsErr := p.Driver.ExecuteScript("arguments[0].scrollIntoView({block: 'center'}); arguments[0].click();", []interface{}{el})
		if jsErr != nil {
			p.LastError = err
			return err
		}
	}
	return nil
}

// SendKeys waits for the element to be visible, clears it thoroughly, and types text.
func (p *Page) SendKeys(locator string, text string, timeout time.Duration) error {
	el, err := p.WaitUntilVisible(locator, timeout)
	if err != nil {
		p.LastError = err
		return err
	}
	_ = el.Clear()
	if text == "" {
		return nil
	}
	err = el.SendKeys(text)
	if err != nil {
		p.LastError = err
		return err
	}
	return nil
}

// GetText waits for the element to be visible and returns its text value.
func (p *Page) GetText(locator string, timeout time.Duration) (string, error) {
	el, err := p.WaitUntilVisible(locator, timeout)
	if err != nil {
		p.LastError = err
		return "", err
	}
	txt, err := el.Text()
	if err != nil {
		p.LastError = err
		return "", err
	}
	return txt, nil
}

// CaptureScreenshot takes a PNG screenshot and writes it to the local screenshots/ directory.
func (p *Page) CaptureScreenshot(testName string) (string, error) {
	screenshotData, err := p.Driver.Screenshot()
	if err != nil {
		return "", fmt.Errorf("failed to capture screenshot bytes: %w", err)
	}

	dir := p.ScreenshotDir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	sanitizedName := strings.ReplaceAll(testName, "/", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, " ", "_")
	filename := fmt.Sprintf("%s-%s.png", sanitizedName, time.Now().Format("15-04-05"))
	fullPath := filepath.Join(dir, filename)

	if err := os.WriteFile(fullPath, screenshotData, 0644); err != nil {
		return "", fmt.Errorf("failed to write screenshot file: %w", err)
	}

	logger.Info("Screenshot captured: %s", fullPath)
	return fullPath, nil
}

// FindElementByTestID finds a web element by data-testid.
func (p *Page) FindElementByTestID(testID string, timeout time.Duration) (selenium.WebElement, error) {
	return p.WaitUntilVisible(fmt.Sprintf("css:[data-testid=\"%s\"]", testID), timeout)
}

// ClickByTestID finds an element by data-testid and clicks it.
func (p *Page) ClickByTestID(testID string, timeout time.Duration) error {
	return p.Click(fmt.Sprintf("css:[data-testid=\"%s\"]", testID), timeout)
}

// SendKeysByTestID finds an element by data-testid, clears it, and types text.
func (p *Page) SendKeysByTestID(testID string, text string, timeout time.Duration) error {
	return p.SendKeys(fmt.Sprintf("css:[data-testid=\"%s\"]", testID), text, timeout)
}

// SelectCustomDropdownByText opens a Radix UI dropdown by trigger testID and clicks the item matching optionText.
func (p *Page) SelectCustomDropdownByText(triggerTestID, optionText string, timeout time.Duration) error {
	if err := p.ClickByTestID(triggerTestID, timeout); err != nil {
		return fmt.Errorf("failed to click dropdown trigger '%s': %w", triggerTestID, err)
	}
	time.Sleep(300 * time.Millisecond) // short pause for menu render

	// Find and click item matching exact text or containing text
	itemXPath := fmt.Sprintf("//*[contains(@role, 'option') and (text()='%s' or .='%s')]", optionText, optionText)
	return p.Click(itemXPath, timeout)
}

// GetTextByTestID finds an element by data-testid and retrieves its text.
func (p *Page) GetTextByTestID(testID string, timeout time.Duration) (string, error) {
	return p.GetText(fmt.Sprintf("css:[data-testid=\"%s\"]", testID), timeout)
}

// VerifyElementsPresentByTestIDs checks if all specified testID elements are visible on the page.
// The first check uses the full timeout to allow components to load, while subsequent checks use a short 200ms timeout
// to prevent cumulative delays.
func (p *Page) VerifyElementsPresentByTestIDs(testIDs []string, timeout time.Duration) error {
	var missing []string
	for i, id := range testIDs {
		t := timeout
		if i > 0 {
			t = 200 * time.Millisecond
		}
		_, err := p.FindElementByTestID(id, t)
		if err != nil {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing expected elements on page: %v", missing)
	}
	return nil
}

// VerifyElementsPresentConcurrently validates that all specified data-testid elements exist on page in parallel.
func (p *Page) VerifyElementsPresentConcurrently(testIDs []string, timeout time.Duration) error {
	if len(testIDs) == 0 {
		return nil
	}

	errChan := make(chan string, len(testIDs))
	var wg sync.WaitGroup

	for _, id := range testIDs {
		wg.Add(1)
		go func(testID string) {
			defer wg.Done()
			if _, err := p.FindElementByTestID(testID, timeout); err != nil {
				errChan <- testID
			}
		}(id)
	}

	wg.Wait()
	close(errChan)

	var missing []string
	for id := range errChan {
		missing = append(missing, id)
	}

	if len(missing) > 0 {
		return fmt.Errorf("concurrent element verification failed, missing: %v", missing)
	}
	return nil
}

type InterceptedNetworkRequest struct {
	Timestamp    string `json:"timestamp"`
	URL          string `json:"url"`
	Method       string `json:"method"`
	Status       int    `json:"status"`
	RequestBody  string `json:"requestBody"`
	ResponseBody string `json:"responseBody"`
	LatencyMs    int64  `json:"latencyMs"`
}

const NetworkInterceptorScript = `
(function() {
    if (window.__network_monkey_patched) return;
    window.__network_monkey_patched = true;

    function logRequest(url, method, status, requestBody, responseBody, latencyMs) {
        var reqStr = typeof requestBody === 'string' ? requestBody : JSON.stringify(requestBody);
        var respStr = typeof responseBody === 'string' ? responseBody : JSON.stringify(responseBody);
        var entry = {
            timestamp: new Date().toISOString(),
            url: url,
            method: method,
            status: status,
            requestBody: reqStr,
            responseBody: respStr,
            latencyMs: latencyMs || 0
        };

        var requests = [];
        try {
            requests = JSON.parse(sessionStorage.getItem('__networkRequests') || '[]');
        } catch(e) {}
        requests.push(entry);
        try {
            sessionStorage.setItem('__networkRequests', JSON.stringify(requests));
        } catch(e) {}

        if (status >= 400 || status === 0) {
            var errors = [];
            try {
                errors = JSON.parse(sessionStorage.getItem('__networkErrors') || '[]');
            } catch(e) {}
            errors.push(entry);
            try {
                sessionStorage.setItem('__networkErrors', JSON.stringify(errors));
            } catch(e) {}
        }
    }

    var originalFetch = window.fetch;
    window.fetch = function(input, init) {
        var startTime = Date.now();
        var url = typeof input === 'string' ? input : (input && input.url) || '';
        var method = (init && init.method) || 'GET';
        var reqBody = (init && init.body) || '';
        return originalFetch.apply(this, arguments).then(function(response) {
            response.clone().text().then(function(text) {
                logRequest(url, method, response.status, reqBody, text, Date.now() - startTime);
            }).catch(function() {
                logRequest(url, method, response.status, reqBody, '[failed to parse response]', Date.now() - startTime);
            });
            return response;
        }).catch(function(error) {
            logRequest(url, method, 0, reqBody, error.message || error.toString(), Date.now() - startTime);
            throw error;
        });
    };

    var originalOpen = XMLHttpRequest.prototype.open;
    var originalSend = XMLHttpRequest.prototype.send;

    XMLHttpRequest.prototype.open = function(method, url) {
        this.__method = method;
        this.__url = url;
        return originalOpen.apply(this, arguments);
    };

    XMLHttpRequest.prototype.send = function(body) {
        var xhr = this;
        var startTime = Date.now();
        xhr.addEventListener('load', function() {
            logRequest(xhr.__url, xhr.__method, xhr.status, body || '', xhr.responseText || '', Date.now() - startTime);
        });
        xhr.addEventListener('error', function() {
            logRequest(xhr.__url, xhr.__method, xhr.status, body || '', '[network error]', Date.now() - startTime);
        });
        return originalSend.apply(this, arguments);
    };
})();
`

// InjectNetworkInterceptor injects the fetch/XHR interceptor script into the browser page.
func (p *Page) InjectNetworkInterceptor() {
	_, _ = p.Driver.ExecuteScript(NetworkInterceptorScript, nil)
}

// RetrieveNetworkRequestsRaw retrieves the captured network requests JSON string from browser sessionStorage.
func (p *Page) RetrieveNetworkRequestsRaw() string {
	res, err := p.Driver.ExecuteScript("return sessionStorage.getItem('__networkRequests');", nil)
	if err != nil || res == nil {
		return ""
	}
	val, ok := res.(string)
	if !ok {
		return ""
	}
	return val
}

// RetrieveNetworkRequests retrieves and parses all captured network requests from browser sessionStorage.
func (p *Page) RetrieveNetworkRequests() []InterceptedNetworkRequest {
	jsonStr := p.RetrieveNetworkRequestsRaw()
	if jsonStr == "" {
		return nil
	}
	var requests []InterceptedNetworkRequest
	if err := json.Unmarshal([]byte(jsonStr), &requests); err != nil {
		return nil
	}
	return requests
}

// RetrieveNetworkErrors retrieves the captured network errors from the browser sessionStorage.
func (p *Page) RetrieveNetworkErrors() string {
	res, err := p.Driver.ExecuteScript("return sessionStorage.getItem('__networkErrors');", nil)
	if err != nil || res == nil {
		return ""
	}
	val, ok := res.(string)
	if !ok {
		return ""
	}
	return val
}

// FormatNetworkErrors parses the JSON error string and formats it into a human-readable summary.
func FormatNetworkErrors(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}
	var errors []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &errors); err != nil {
		return fmt.Sprintf("[Error parsing network logs: %v]\nRaw Log: %s", err, jsonStr)
	}
	if len(errors) == 0 {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString("==================================================\n")
	buf.WriteString("🚨 INTERCEPTED BACKEND NETWORK ERRORS / FAILURES:\n")
	buf.WriteString("==================================================\n")
	for i, e := range errors {
		buf.WriteString(fmt.Sprintf("[%d] Timestamp: %v\n", i+1, e["timestamp"]))
		buf.WriteString(fmt.Sprintf("    Request  : %v %v\n", e["method"], e["url"]))
		if reqBody := e["requestBody"]; reqBody != "" {
			buf.WriteString(fmt.Sprintf("    Payload  : %v\n", reqBody))
		}
		buf.WriteString(fmt.Sprintf("    Status   : %v\n", e["status"]))
		buf.WriteString(fmt.Sprintf("    Response : %v\n", e["responseBody"]))
		buf.WriteString("--------------------------------------------------\n")
	}
	return buf.String()
}


