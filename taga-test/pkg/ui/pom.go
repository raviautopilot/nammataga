package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"e2e-template/pkg/logger"
)

// Page serves as the base Page Object wrapper.
type Page struct {
	Driver        selenium.WebDriver
	ScreenshotDir string
}

// NewPage initializes a base Page.
func NewPage(driver selenium.WebDriver, screenshotDir string) *Page {
	return &Page{Driver: driver, ScreenshotDir: screenshotDir}
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
		return nil, fmt.Errorf("element '%s' was not visible after %v: %w", locator, timeout, err)
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
		return nil, fmt.Errorf("element '%s' was not clickable after %v: %w", locator, timeout, err)
	}
	return elem, nil
}

// Click waits for the element to be clickable and performs a click.
func (p *Page) Click(locator string, timeout time.Duration) error {
	el, err := p.WaitUntilClickable(locator, timeout)
	if err != nil {
		return err
	}
	return el.Click()
}

// SendKeys waits for the element to be visible, clears it, and types text.
func (p *Page) SendKeys(locator string, text string, timeout time.Duration) error {
	el, err := p.WaitUntilVisible(locator, timeout)
	if err != nil {
		return err
	}
	_ = el.Clear()
	return el.SendKeys(text)
}

// GetText waits for the element to be visible and returns its text value.
func (p *Page) GetText(locator string, timeout time.Duration) (string, error) {
	el, err := p.WaitUntilVisible(locator, timeout)
	if err != nil {
		return "", err
	}
	return el.Text()
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
