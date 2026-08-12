package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// TestAPI_AdminInitPassword covers the POST /api/admin/init-password endpoint
func TestAPI_AdminInitPassword_Flows(t *testing.T) {
	// No memberId → 400
	tests.RunAPITestWithDetails(t, "[Admin] POST Init Password - Missing memberId", "Calls init-password without memberId query param.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/init-password", nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "Please specify memberId")
	})

	// Specific (non-all) memberId → 501 Not Implemented
	tests.RunAPITestWithDetails(t, "[Admin] POST Init Password - Specific Member Not Implemented", "Calls init-password with a specific memberId, expects 501.", "HTTP 501 Not Implemented", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/init-password?memberId=some-specific-member-id", nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusNotImplemented, "not yet implemented")
	})
}

// TestAPI_AdminAuditUsers covers GET /api/admin/audit/users
func TestAPI_AdminAuditUsers_Flows(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] GET Audit Users - Valid Year/Month", "Fetches unique user IDs that have audit data for 2026-08.", "HTTP 200 OK with users list", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/audit/users?year=2026&month=08", nil, nil, &resp, auth)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		users, _ := resp["users"].([]interface{})
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, %d unique users found in audit", len(users))
	})

	tests.RunAPITestWithDetails(t, "[Admin] GET Audit Users - Invalid Month", "Passes month=99 to verify query validation.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/audit/users?year=2026&month=99", nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "Invalid month")
	})

	tests.RunAPITestWithDetails(t, "[Admin] GET Audit Users - Invalid Year", "Passes a non-numeric year to verify query validation.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/audit/users?year=abcd", nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "Invalid year")
	})
}

// TestAPI_LegacyAdminUpload covers the /admin/upload-registration legacy route
func TestAPI_LegacyAdminUpload(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] POST Legacy Upload Registration - Empty Payload", "Calls the legacy /admin/upload-registration route without any file.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/admin/upload-registration", nil, nil, &resp, auth)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})
}

// TestAPI_AdminMemberReports covers GET /api/admin/reports/members
func TestAPI_AdminMemberReports(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] GET Member Reports", "Fetches the member summary report data.", "HTTP 200 OK with report data", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		// Report returns CSV/text, so pass nil to skip JSON unmarshalling
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/reports/members", nil, nil, nil, auth)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK member report data retrieved"
	})
}

func TestAPI_AdminRemaining_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Public] GET Admin Reports - Unauthenticated",
			Description: "Attempts to download member reports without an admin token.",
			Expected:    "HTTP 401 Unauthorized",
			TestFn: func(tc *tests.TestContext) {
				err := tc.Client.SendHttpRequest("GET", "/api/admin/reports/members", nil, nil, nil, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for missing token"
					tc.Fatalf("Expected 401 Unauthorized")
				}
				tc.Actual = "Correctly blocked unauthenticated access to reports"
			},
		},
		{
			Name:        "[Admin] GET Audit Users - Wrong HTTP Method",
			Description: "Sends POST request to the GET audit users endpoint.",
			Expected:    "HTTP 405 Method Not Allowed",
			TestFn: func(tc *tests.TestContext) {
				token := getValidAdminToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/admin/audit/users", nil, nil, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected failure on incorrect HTTP method"
					tc.Fatalf("Expected error for wrong method")
				}
				tc.Actual = "Correctly rejected wrong HTTP method"
			},
		},
		{
			Name:        "[Admin] POST Init Password - Type Juggling",
			Description: "Attempts to pass an array instead of string for memberId.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				token := getValidAdminToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/admin/init-password?memberId[]=1&memberId[]=2", nil, nil, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected rejection for type juggling"
					tc.Fatalf("Expected error for type juggling but got success")
				}
				tc.Actual = "Correctly rejected type juggling payload"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
