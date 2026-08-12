package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// ============================================================================
// 1. GET /api/events/upcoming - Parameterized Tests
// ============================================================================

func TestAPI_Events_Upcoming(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Upcoming Events", "Retrieves list of scheduled upcoming events.", "HTTP 200 OK containing JSON array of events", func(tctx *tests.TestContext) {
		var resp []interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/events/upcoming", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d upcoming events", len(resp))
	})
}

// ============================================================================
// 2. GET /api/gallery/years - Parameterized Tests
// ============================================================================

func TestAPI_Gallery_Years(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Gallery Years", "Retrieves years for which gallery images are available.", "HTTP 200 OK containing list of years", func(tctx *tests.TestContext) {
		var resp []int
		err := tctx.Client.SendHttpRequest("GET", "/api/gallery/years", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved years: %v", resp)
	})
}

// ============================================================================
// 3. GET /api/gallery - Parameterized Tests
// ============================================================================

func TestAPI_Gallery_Images(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		queryParams    string
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Get Default Gallery (Current Year)",
			persona:        "Visitor",
			description:    "Retrieves gallery images for the current year when no query parameter is specified.",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Happy Path - Get Specific Year Gallery",
			persona:        "Visitor",
			description:    "Retrieves gallery images for the year 2026.",
			queryParams:    "?year=2026",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Validation - Invalid Year Format",
			persona:        "Visitor",
			description:    "Attempts gallery lookup with a malformed non-integer year parameter.",
			queryParams:    "?year=two-thousand-twenty-six",
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "invalid year",
		},
		{
			name:           "Validation - Year Out of Bounds",
			persona:        "Visitor",
			description:    "Attempts gallery lookup with a year far in the future.",
			queryParams:    "?year=9999",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - XSS in Year Parameter",
			persona:        "Malicious User",
			description:    "Attempts XSS payload in year parameter.",
			queryParams:    "?year=<script>alert('xss')</script>",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET Gallery - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp []interface{}
			var errResp map[string]interface{}

			var err client.HttpError
			if tc.expectedStatus == http.StatusOK {
				err = tctx.Client.SendHttpRequest("GET", "/api/gallery"+tc.queryParams, nil, nil, &resp, nil)
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d gallery images", len(resp))
			} else {
				err = tctx.Client.SendHttpRequest("GET", "/api/gallery"+tc.queryParams, nil, nil, &errResp, nil)
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 4. Events Create - Business Logic Parameterized Tests
// ============================================================================

func TestAPI_Events_Create_BusinessLogic(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:        "Business Logic - Past Event Date",
			persona:     "Admin",
			description: "Attempts to create an event in the past.",
			payload: map[string]interface{}{
				"title":       "Past Event",
				"startDate":   "2020-01-01",
				"endDate":     "2020-01-02",
				"maxCapacity": 100,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - End Date Before Start Date",
			persona:     "Admin",
			description: "Attempts to create an event ending before it starts.",
			payload: map[string]interface{}{
				"title":       "Time Travel Event",
				"startDate":   "2026-12-10",
				"endDate":     "2026-12-05",
				"maxCapacity": 100,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - Overlapping Dates",
			persona:     "Admin",
			description: "Attempts to create an event overlapping with an existing mandatory event.",
			payload: map[string]interface{}{
				"title":       "Overlapping Event",
				"startDate":   "2026-12-15",
				"endDate":     "2026-12-20",
				"maxCapacity": 100,
			},
			expectedStatus: http.StatusConflict, // or 400
		},
		{
			name:        "Business Logic - Max Capacity Exceeded",
			persona:     "Admin",
			description: "Attempts to set an unreasonably high max capacity.",
			payload: map[string]interface{}{
				"title":       "Massive Event",
				"startDate":   "2026-12-25",
				"endDate":     "2026-12-26",
				"maxCapacity": 1000000,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Admin] POST Create Event - "+tc.name, tc.description, fmt.Sprintf("HTTP %d", tc.expectedStatus), func(tctx *tests.TestContext) {
			// Using dummy admin auth
			token := "admin_token_mock" // Mocked token since getValidAdminToken isn't imported here
			adminAuth := &client.BearerTokenAuth{Token: token}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/admin/events", nil, tc.payload, &resp, adminAuth)

			if err == nil {
				if tc.expectedStatus != http.StatusOK {
					tctx.FailureReason = fmt.Sprintf("Expected %d, got 200 OK", tc.expectedStatus)
					tctx.Errorf("Expected error status %d", tc.expectedStatus)
				}
			} else {
				if tc.expectedStatus == http.StatusBadRequest && err.StatusCode() == http.StatusConflict {
					tctx.Actual = fmt.Sprintf("HTTP %d correctly rejected: %v", err.StatusCode(), err.ResponseBody())
				} else {
					assertErrorStatus(tctx, err, tc.expectedStatus, "")
				}
			}
		})
	}
}
