package ui

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"e2e-template/pkg/logger"
)

// InitWebDriver connects to a running Selenium WebDriver / Chromedriver instance.
func InitWebDriver(seleniumURL string, headless bool) (selenium.WebDriver, error) {
	caps := selenium.Capabilities{"browserName": "chrome"}

	var args []string
	if headless {
		args = append(args, "--headless", "--disable-gpu")
	}
	// Standard arguments for stable docker/linux execution
	args = append(args, "--no-sandbox", "--disable-dev-shm-usage", "--window-size=1920,1080")

	chromeCaps := chrome.Capabilities{
		Args: args,
	}

	caps.AddChrome(chromeCaps)

	logger.Info("Connecting to Selenium WebDriver at %s (headless=%v)...", seleniumURL, headless)

	var driver selenium.WebDriver
	var err error

	// Retry connection up to 3 times
	for i := 1; i <= 3; i++ {
		driver, err = selenium.NewRemote(caps, seleniumURL)
		if err == nil {
			break
		}
		logger.Warn("Failed to connect to WebDriver (attempt %d/3): %v. Retrying in 2 seconds...", i, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("web driver connection failed: %w", err)
	}

	return driver, nil
}
