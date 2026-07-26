package ui_tests

import (
	"testing"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

func TestUI_01_ServiceRunning(t *testing.T) {
	tests.RunUITest(t, "01_Service_Running", func(t *testing.T, page *ui.Page) {
		targetURL := tests.GlobalConfig.UiURL
		if targetURL == "" {
			t.Fatal("GlobalConfig.UiURL is empty")
		}

		t.Logf("Navigating to target URL: %s", targetURL)
		if err := page.Driver.Get(targetURL); err != nil {
			t.Fatalf("Failed to navigate to %s: %v", targetURL, err)
		}

		// Take a manual screenshot
		screenshotPath, err := page.CaptureScreenshot("01-service-running")
		if err != nil {
			t.Logf("Warning: failed to capture screenshot: %v", err)
		} else {
			t.Logf("Screenshot captured successfully at: %s", screenshotPath)
		}

		// Verify page loaded successfully by retrieving title or checking body element
		title, err := page.Driver.Title()
		if err != nil {
			t.Fatalf("Failed to get page title: %v", err)
		}
		t.Logf("Successfully loaded page. Title: %s", title)

		body, err := page.Driver.FindElement("css selector", "body")
		if err != nil {
			t.Fatalf("Failed to find body element: %v", err)
		}
		if body == nil {
			t.Fatal("Body element is nil")
		}
	})
}
