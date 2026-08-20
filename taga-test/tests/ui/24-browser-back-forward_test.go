package ui_test

import (
	"fmt"
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/pkg/ui/actions"
	"e2e-template/tests"
)

// TestUI_24_BrowserBackForward verifies that the browser Back and Forward buttons
// navigate between previously visited in-app pages instead of leaving the website.
//
// Journey:
//   1. Home → Office Bearers → Events → Member Login
//   2. Press Back → should be on Events
//   3. Press Back → should be on Office Bearers
//   4. Press Forward → should be on Events
//   5. Press Forward → should be on Member Login
func TestUI_24_BrowserBackForward(t *testing.T) {
	tests.RunUITest(t, "Verify Browser Back/Forward Navigation", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		pubPersona := actions.NewPublicPersona(page, cfg.UiURL, 5*time.Second)
		result := actions.NewResult("TestUI_24_BrowserBackForward")

		// Step 1: Navigate through several pages to build browser history
		t.Log("Step 1: Building navigation history — Home → Office Bearers → Events → Member Login")

		actions.GoToHome(pubPersona, result)
		if result.Failed() {
			t.Fatalf("Failed to navigate to Home: %v", result.Error)
		}

		actions.GoToOfficeBeaers(pubPersona, result)
		if result.Failed() {
			t.Fatalf("Failed to navigate to Office Bearers: %v", result.Error)
		}

		actions.GoToEvents(pubPersona, result)
		if result.Failed() {
			t.Fatalf("Failed to navigate to Events: %v", result.Error)
		}

		actions.GoToMemberLogin(pubPersona, result)
		if result.Failed() {
			t.Fatalf("Failed to navigate to Member Login: %v", result.Error)
		}

		// Allow time for history state to settle
		time.Sleep(500 * time.Millisecond)

		// Step 2: Press Back — should go to Events (not leave the website)
		t.Log("Step 2: Pressing browser Back button — expecting Events page")
		if err := page.Driver.Back(); err != nil {
			t.Fatalf("Browser Back() failed: %v", err)
		}
		time.Sleep(1 * time.Second) // wait for popstate + React re-render

		result.CaptureScreenshot(page, "After_Back_1_Expect_Events")

		// Verify we're on the Events page by checking the URL hash and/or the events container
		currentURL, err := page.Driver.CurrentURL()
		if err != nil {
			t.Fatalf("Failed to get current URL after Back: %v", err)
		}
		t.Logf("After 1st Back: URL = %s", currentURL)

		eventsEl, eventsErr := page.WaitUntilVisible("css:[data-testid='testid-events-container']", 5*time.Second)
		if eventsErr != nil || eventsEl == nil {
			// Also accept hash-based verification as fallback
			if !containsHash(currentURL, "events") {
				t.Errorf("Expected Events page after 1st Back, but got URL: %s", currentURL)
			}
		} else {
			t.Log("✅ Events page confirmed after 1st Back")
		}

		// Step 3: Press Back again — should go to Office Bearers
		t.Log("Step 3: Pressing browser Back button again — expecting Office Bearers page")
		if err := page.Driver.Back(); err != nil {
			t.Fatalf("Browser Back() failed: %v", err)
		}
		time.Sleep(1 * time.Second)

		result.CaptureScreenshot(page, "After_Back_2_Expect_OfficeBearers")

		currentURL, err = page.Driver.CurrentURL()
		if err != nil {
			t.Fatalf("Failed to get current URL after 2nd Back: %v", err)
		}
		t.Logf("After 2nd Back: URL = %s", currentURL)

		officeBearersEl, obErr := page.WaitUntilVisible("css:[data-testid='testid-office-bearers-district-select']", 5*time.Second)
		if obErr != nil || officeBearersEl == nil {
			if !containsHash(currentURL, "office-bearers") {
				t.Errorf("Expected Office Bearers page after 2nd Back, but got URL: %s", currentURL)
			}
		} else {
			t.Log("✅ Office Bearers page confirmed after 2nd Back")
		}

		// Step 4: Press Forward — should go to Events
		t.Log("Step 4: Pressing browser Forward button — expecting Events page")
		if err := page.Driver.Forward(); err != nil {
			t.Fatalf("Browser Forward() failed: %v", err)
		}
		time.Sleep(1 * time.Second)

		result.CaptureScreenshot(page, "After_Forward_1_Expect_Events")

		currentURL, err = page.Driver.CurrentURL()
		if err != nil {
			t.Fatalf("Failed to get current URL after Forward: %v", err)
		}
		t.Logf("After 1st Forward: URL = %s", currentURL)

		eventsEl2, eventsErr2 := page.WaitUntilVisible("css:[data-testid='testid-events-container']", 5*time.Second)
		if eventsErr2 != nil || eventsEl2 == nil {
			if !containsHash(currentURL, "events") {
				t.Errorf("Expected Events page after Forward, but got URL: %s", currentURL)
			}
		} else {
			t.Log("✅ Events page confirmed after Forward")
		}

		// Step 5: Press Forward again — should go to Member Login
		t.Log("Step 5: Pressing browser Forward button again — expecting Member Login page")
		if err := page.Driver.Forward(); err != nil {
			t.Fatalf("Browser Forward() failed: %v", err)
		}
		time.Sleep(1 * time.Second)

		result.CaptureScreenshot(page, "After_Forward_2_Expect_MemberLogin")

		currentURL, err = page.Driver.CurrentURL()
		if err != nil {
			t.Fatalf("Failed to get current URL after 2nd Forward: %v", err)
		}
		t.Logf("After 2nd Forward: URL = %s", currentURL)

		loginEl, loginErr := page.WaitUntilVisible("css:[data-testid='testid-login-form']", 5*time.Second)
		if loginErr != nil || loginEl == nil {
			if !containsHash(currentURL, "member-login") {
				t.Errorf("Expected Member Login page after 2nd Forward, but got URL: %s", currentURL)
			}
		} else {
			t.Log("✅ Member Login page confirmed after 2nd Forward")
		}

		// Final summary
		if result.Failed() {
			t.Errorf("Test Journey Failed: %v", result.Error)
			t.Errorf("Actions Attempted: %v", result.Actions)
			t.Errorf("Evidence Captured: %v", result.Evidence)
			t.Fatalf("Advice / Remediation: %v", result.Advice)
		}

		t.Log("✅ Browser Back/Forward navigation test passed — all pages navigated correctly")
	})
}

// containsHash checks if the URL contains the expected hash fragment.
func containsHash(url, expectedPage string) bool {
	expected := fmt.Sprintf("#%s", expectedPage)
	return len(url) > 0 && contains(url, expected)
}

// contains is a simple substring check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
