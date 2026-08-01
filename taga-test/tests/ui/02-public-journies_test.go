package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

func TestUI_02_PublicJourneys(t *testing.T) {
	tests.RunUITest(t, "Verify Public User Journeys", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		url := cfg.UiURL
		timeout := 5 * time.Second

		// ── Step 1: Home Page ──────────────────────────────────────────────
		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to open Home Page: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_01_Public_Home_Page")

		// ── Step 2: Office Bearers Page ───────────────────────────────────
		if err := page.ClickByTestID("testid-office-bearers-button", timeout); err != nil {
			t.Fatalf("Failed to navigate to Office Bearers: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_02_Public_Office_Bearers")

		// ── Step 3: Events Page — Gallery Tab (default) ───────────────────
		if err := page.ClickByTestID("testid-events-button", timeout); err != nil {
			t.Fatalf("Failed to navigate to Events: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_03_Public_Events_Gallery")

		// ── Step 4: Events Page — Upcoming Events Tab ─────────────────────
		if err := page.ClickByTestID("testid-upcoming-events-button", timeout); err != nil {
			t.Fatalf("Failed to click Upcoming Events tab: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_04_Public_Events_Upcoming")

		// ── Step 5: Member Login Page ─────────────────────────────────────
		if err := page.GoToHome(url); err != nil {
			t.Fatalf("Failed to return to Home for Member Login: %v", err)
		}
		if err := page.ClickByTestID("testid-member-login-button", timeout); err != nil {
			t.Fatalf("Failed to open Member Login: %v", err)
		}
		time.Sleep(2 * time.Second)
		_, _ = page.CaptureScreenshot("Step_05_Public_Member_Login")
	})
}
