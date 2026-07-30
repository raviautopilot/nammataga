package ui_tests

import (
	"testing"
	"time"

	"e2e-template/pkg/ui"
	"e2e-template/tests"
)

// TestUI_02_LoginValidation executes an enterprise table-driven E2E test verifying Admin and Member login flows.
func TestUI_02_LoginValidation(t *testing.T) {
	tests.RunUITest(t, "Verify Admin and Member Login Element Availability", func(t *testing.T, page *ui.Page) {
		cfg := tests.GlobalConfig
		if cfg.UiURL == "" {
			t.Fatal("GlobalConfig.UiURL is empty")
		}

		// Table-driven test definitions loaded entirely from config.json
		loginFlows := []struct {
			Role       string
			ButtonID   string
			ElementIDs []string
		}{
			{
				Role:       "Admin",
				ButtonID:   cfg.AdminLoginButtonTestID,
				ElementIDs: cfg.AdminLoginTestIDs,
			},
			{
				Role:       "Member",
				ButtonID:   cfg.MemberLoginButtonTestID,
				ElementIDs: cfg.MemberLoginTestIDs,
			},
		}

		// STEP 1: INITIAL LANDING & HOME PAGE VERIFICATION
		if err := page.GoToHomePage(cfg.UiURL); err != nil {
			t.Fatalf("Initial landing - GoToHomePage failed: %v", err)
		}

		// Verify both landing login trigger buttons concurrently
		landingButtons := []string{cfg.AdminLoginButtonTestID, cfg.MemberLoginButtonTestID}
		if err := page.VerifyElementsPresentByTestIDs(landingButtons, 5*time.Second); err != nil {
			t.Fatalf("Landing page login options check failed: %v", err)
		}
		t.Log("✅ Home page verified: Both Admin and Member login options are present")

		// STEP 2: TABLE-DRIVEN FLOW EXECUTION
		for _, flow := range loginFlows {
			flow := flow // capture loop variable for sub-test isolation
			t.Run(flow.Role+"_Login_Availability", func(subT *testing.T) {
				if flow.ButtonID == "" {
					subT.Fatalf("%s login button test ID is empty in config.json", flow.Role)
				}
				if len(flow.ElementIDs) == 0 {
					subT.Fatalf("%s login test IDs array is empty in config.json", flow.Role)
				}

				// 1. Reset state by navigating to Home Page
				if err := page.GoToHomePage(cfg.UiURL); err != nil {
					subT.Fatalf("[%s Flow] GoToHomePage failed: %v", flow.Role, err)
				}

				// 2. Click target login button
				if err := page.ClickByTestID(flow.ButtonID, 5*time.Second); err != nil {
					subT.Fatalf("[%s Flow] Failed to click login button (%s): %v", flow.Role, flow.ButtonID, err)
				}

				// 3. Verify form elements in parallel
				if err := page.VerifyElementsPresentByTestIDs(flow.ElementIDs, 5*time.Second); err != nil {
					subT.Fatalf("[%s Flow] Form elements validation failed: %v", flow.Role, err)
				}

				// 4. Return back to Home Page
				if err := page.GoToHomePage(cfg.UiURL); err != nil {
					subT.Fatalf("[%s Flow] Return to GoToHomePage failed: %v", flow.Role, err)
				}

				subT.Logf("✅ %s login flow validated successfully", flow.Role)
			})
		}
	})
}
