package api_tests

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
