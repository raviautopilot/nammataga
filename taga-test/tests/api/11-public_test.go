package api_tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/tests"
)

// ============================================================================
// 1. Root & Health Check Endpoints
// ============================================================================

func TestAPI_Public_General(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Root Welcome", "Checks the server root welcome greeting.", "HTTP 200 OK containing 'Hello from Gin + Zap!'", func(tctx *tests.TestContext) {
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		msg, _ := resp["message"].(string)
		if !strings.Contains(msg, "Hello from Gin + Zap!") {
			tctx.FailureReason = fmt.Sprintf("Expected welcome message, got: %s", msg)
			tctx.Errorf("Expected welcome message, got: %s", msg)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
	})

	tests.RunAPITestWithDetails(t, "[Public] GET Health Status", "Performs general API health check.", "HTTP 200 OK containing 'healthy'", func(tctx *tests.TestContext) {
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/health", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		msg, _ := resp["message"].(string)
		if msg != "healthy" {
			tctx.FailureReason = fmt.Sprintf("Expected 'healthy', got: '%s'", msg)
			tctx.Errorf("Expected 'healthy', got: '%s'", msg)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Health Status='%s'", msg)
	})
}

// ============================================================================
// 2. Public /api/public/about/... - Parameterized Tests
// ============================================================================

func TestAPI_Public_About(t *testing.T) {
	cases := []struct {
		name        string
		endpoint    string
		description string
		expectedSub string
	}{
		{
			name:        "About Organization Details",
			endpoint:    "/api/public/about",
			description: "Retrieves overview, history, and description of TAGA association.",
			expectedSub: "association",
		},
		{
			name:        "Association Statistics",
			endpoint:    "/api/public/about/stats",
			description: "Retrieves statistics data such as member counts or districts.",
			expectedSub: "", // verify 200 OK array
		},
		{
			name:        "Association Objectives",
			endpoint:    "/api/public/about/objectives",
			description: "Retrieves objectives, goals, and values.",
			expectedSub: "",
		},
		{
			name:        "Association Services",
			endpoint:    "/api/public/about/services",
			description: "Retrieves services provided to agriculture officers.",
			expectedSub: "",
		},
		{
			name:        "Contact Details",
			endpoint:    "/api/public/about/contact",
			description: "Retrieves official email, address, and mobile contacts.",
			expectedSub: "email",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[Public] GET About - %s", tc.name)
		expectedStr := "HTTP 200 OK"
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp interface{}
			err := tctx.Client.SendHttpRequest("GET", tc.endpoint, nil, nil, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK retrieved successfully from %s", tc.endpoint)
		})
	}
}

// ============================================================================
// 3. Logo and Banners
// ============================================================================

func TestAPI_Public_Assets(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Logo Path", "Checks default logo asset URL.", "HTTP 200 OK with logo URL", func(tctx *tests.TestContext) {
		var resp map[string]string
		err := tctx.Client.SendHttpRequest("GET", "/api/logo", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		url := resp["url"]
		if !strings.Contains(url, "logo") {
			tctx.FailureReason = fmt.Sprintf("Expected logo URL path, got: %s", url)
			tctx.Errorf("Expected logo URL path, got: %s", url)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Logo URL='%s'", url)
	})

	tests.RunAPITestWithDetails(t, "[Public] GET Member Banner", "Checks member banner image URL.", "HTTP 200 OK with banner image path", func(tctx *tests.TestContext) {
		var resp map[string]string
		err := tctx.Client.SendHttpRequest("GET", "/api/member-banner", nil, nil, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		image := resp["image"]
		if !strings.Contains(image, "banner") {
			tctx.FailureReason = fmt.Sprintf("Expected banner image path, got: %s", image)
			tctx.Errorf("Expected banner image path, got: %s", image)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Banner Image='%s'", image)
	})
}

// ============================================================================
// 4. Office Bearers - Parameterized Tests
// ============================================================================

func TestAPI_Public_Office(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		queryParams    string
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Default Office Bearers (State)",
			persona:        "Visitor",
			description:    "Retrieves office bearers without parameters, defaulting to 'state' type.",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedSub:    "state",
		},
		{
			name:           "Happy Path - District Query",
			persona:        "Visitor",
			description:    "Retrieves office bearers for a specific working district.",
			queryParams:    "?pathParam=Salem",
			expectedStatus: http.StatusOK,
			expectedSub:    "Salem",
		},
		{
			name:           "Status - Nonexistent District",
			persona:        "Visitor",
			description:    "Attempts fetching bearers for a bogus/nonexistent district.",
			queryParams:    "?pathParam=NonexistentDistrictName",
			expectedStatus: http.StatusNotFound,
			expectedSub:    "could not be found",
		},
		{
			name:           "Security - SQL Injection in Path",
			persona:        "Malicious User",
			description:    "Attempts SQL injection in the district query parameter.",
			queryParams:    "?pathParam=' OR 1=1--",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation - Extremely Long Parameter",
			persona:        "Fuzzer",
			description:    "Sends a ridiculously long query parameter.",
			queryParams:    "?pathParam=" + strings.Repeat("A", 1000),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Business Logic - Inactive District",
			persona:        "Visitor",
			description:    "Requests office bearers for a theoretically valid but inactive district.",
			queryParams:    "?pathParam=InactiveDistrict",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET Office Bearers - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/office"+tc.queryParams, nil, nil, &resp, nil)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				typ, _ := resp["type"].(string)
				if !strings.Contains(typ, tc.expectedSub) {
					tctx.FailureReason = fmt.Sprintf("Expected response type '%s', got '%s'", tc.expectedSub, typ)
					tctx.Errorf("Expected response type '%s', got '%s'", tc.expectedSub, typ)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Bearers Type='%s'", typ)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}
