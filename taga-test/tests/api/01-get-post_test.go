package api_test

import (
	"fmt"
	"testing"

	"e2e-template/tests"
)

type PublicGetTestCase struct {
	Name        string
	Path        string
	Description string
	Expected    string
	ValidateFn  func(tc *tests.TestContext)
}

// TestAPI_01_GetPost_TableDriven is a parameterized table-driven test verifying public GET endpoints.
func TestAPI_01_GetPost_TableDriven(t *testing.T) {
	testCases := []PublicGetTestCase{
		{
			Name:        "Get About Details",
			Path:        "/api/public/about",
			Description: "Verifies retrieving organization metadata from /api/public/about.",
			Expected:    "HTTP 200 OK with valid non-empty AboutResponse ID and Name",
			ValidateFn: func(tc *tests.TestContext) {
				var about AboutResponse
				err := tc.Client.SendHttpRequest("GET", "/api/public/about", nil, nil, &about, nil)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
					tc.Fatalf("Failed to retrieve about info: %v", err)
				}
				tc.Actual = fmt.Sprintf("HTTP 200 OK, ID=%d, Name='%s'", about.ID, about.Name)
				if about.ID == 0 {
					tc.FailureReason = "Expected non-zero ID in AboutResponse, got 0"
					tc.Errorf("Expected non-zero ID in AboutResponse, got 0")
				}
				if about.Name == "" {
					tc.FailureReason = "Expected non-empty Name in AboutResponse"
					tc.Errorf("Expected non-empty Name in AboutResponse")
				}
			},
		},
		{
			Name:        "Get About Stats",
			Path:        "/api/public/about/stats",
			Description: "Verifies retrieving association statistics from /api/public/about/stats.",
			Expected:    "HTTP 200 OK with valid stats array",
			ValidateFn: func(tc *tests.TestContext) {
				var stats []StatsResponse
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats", nil, nil, &stats, nil)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("HTTP Request failed: %v", err)
					tc.Fatalf("Failed to retrieve stats: %v", err)
				}
				tc.Actual = fmt.Sprintf("HTTP 200 OK, Count=%d", len(stats))
				if len(stats) == 0 {
					tc.FailureReason = "Expected non-empty stats array"
					tc.Errorf("Expected non-empty stats array")
				}
			},
		},
		{
			Name:        "Invalid Route /api/public/about/invalid",
			Path:        "/api/public/about/invalid",
			Description: "Verifies 404 Not Found on invalid endpoint.",
			Expected:    "HTTP 404 Not Found",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/invalid", nil, nil, &dummy, nil)
				if err == nil {
					tc.FailureReason = "Expected 404 error, got 200 OK"
					tc.Errorf("Expected 404 error, got 200 OK")
				} else if err.StatusCode() != 404 {
					tc.FailureReason = fmt.Sprintf("Expected 404, got %d", err.StatusCode())
					tc.Errorf("Expected 404, got %d", err.StatusCode())
				} else {
					tc.Actual = "HTTP 404 Not Found as expected"
				}
			},
		},
		{
			Name:        "Wrong HTTP Method POST on /api/public/about",
			Path:        "/api/public/about",
			Description: "Verifies POST on GET endpoint is rejected.",
			Expected:    "HTTP 405 Method Not Allowed or 404 Not Found",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/public/about", nil, nil, &dummy, nil)
				if err == nil {
					tc.FailureReason = "Expected HTTP error, got 200 OK"
					tc.Errorf("Expected HTTP error, got 200 OK")
				} else if err.StatusCode() != 405 && err.StatusCode() != 404 {
					tc.FailureReason = fmt.Sprintf("Expected 405/404, got %d", err.StatusCode())
					tc.Errorf("Expected 405/404, got %d", err.StatusCode())
				} else {
					tc.Actual = fmt.Sprintf("Rejected with %d as expected", err.StatusCode())
				}
			},
		},
		{
			Name:        "SQL Injection Query Parameter on /api/public/about",
			Path:        "/api/public/about?id=1' OR '1'='1",
			Description: "Verifies SQL injection payload does not cause a server crash (500).",
			Expected:    "HTTP 200 OK or 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/public/about?id=1' OR '1'='1", nil, nil, &dummy, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on SQL injection payload"
					tc.Errorf("Server crashed on SQL injection payload")
				} else {
					tc.Actual = "Handled safely without crashing"
				}
			},
		},
		{
			Name:        "Business Logic - Negative Page/Limit for Stats",
			Path:        "/api/public/about/stats?page=-1&limit=-100",
			Description: "Verifies negative numbers for pagination are handled gracefully.",
			Expected:    "HTTP 400 Bad Request or default pagination",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats?page=-1&limit=-100", nil, nil, &dummy, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on negative pagination parameters"
					tc.Errorf("Server crashed on negative pagination parameters")
				} else if err != nil && err.StatusCode() == 400 {
					tc.Actual = "Handled gracefully with 400 Bad Request"
				} else {
					tc.Actual = "Handled safely with default values or 200 OK"
				}
			},
		},
		{
			Name:        "Extreme Boundary - Max Integer Pagination Limit",
			Path:        "/api/public/about/stats?page=1&limit=9223372036854775807",
			Description: "Verifies extreme integer bounds on pagination do not crash the server.",
			Expected:    "HTTP 400 Bad Request or default pagination",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats?page=1&limit=9223372036854775807", nil, nil, &dummy, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on max integer bounds"
					tc.Errorf("Server crashed on max integer bounds")
				} else {
					tc.Actual = "Handled gracefully"
				}
			},
		},
		{
			Name:        "Logical Paradox - Contradictory Date Range with Negative Capacity",
			Path:        "/api/public/about/stats?start_date=2100-01-01&end_date=1900-01-01&capacity=-5",
			Description: "Verifies that dates contradicting each other with negative capacities are handled gracefully.",
			Expected:    "HTTP 400 Bad Request or 200 OK without crashing",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats?start_date=2100-01-01&end_date=1900-01-01&capacity=-5", nil, nil, &dummy, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on logical paradox payload"
					tc.Errorf("Server crashed on logical paradox payload")
				} else {
					tc.Actual = "Handled gracefully"
				}
			},
		},
		{
			Name:        "State Machine Violation - Cancel Already Completed Stats Task via GET",
			Path:        "/api/public/about/stats?action=cancel&status=completed",
			Description: "Verifies state machine violation attempt via query parameters on GET.",
			Expected:    "HTTP 400 Bad Request or safely ignored",
			ValidateFn: func(tc *tests.TestContext) {
				var dummy map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats?action=cancel&status=completed", nil, nil, &dummy, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on state machine violation"
					tc.Errorf("Server crashed on state machine violation")
				} else {
					tc.Actual = "Handled gracefully"
				}
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(
			t,
			testCase.Name,
			testCase.Description,
			testCase.Expected,
			testCase.ValidateFn,
		)
	}
}
