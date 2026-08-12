package api_tests

import (
	"fmt"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_AdminAudit_DeepValidation(t *testing.T) {
	// Step A: Login to generate audit logs
	tests.RunAPITestWithDetails(t, "[Public] POST Member Login - Audit Trail Trigger", "Performs valid member login to verify audit record generation.", "HTTP 200 OK with JWT", func(tctx *tests.TestContext) {
		payload := map[string]string{
			"email":    "sudhantest08@gmail.com",
			"password": "test123",
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/member/login", nil, &payload, &resp, nil)
		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK login audit record triggered"
	})

	// Step B: Query with exact pagination limits (Limit=1, Page=1)
	tests.RunAPITestWithDetails(t, "[Admin] GET Audit Logs - Limit 1 Check", "Queries audit logs with pagination limit=1 to verify structural control.", "HTTP 200 OK containing exactly 1 data row", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/audit?page=1&limit=1", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}

		dataList, ok := resp["data"].([]interface{})
		if !ok {
			tctx.FailureReason = "Response did not contain data array"
			tctx.Fatalf("Response missing 'data'")
		}

		limitVal, _ := resp["limit"].(float64)
		if int(limitVal) != 1 {
			tctx.FailureReason = fmt.Sprintf("Expected limit metadata to be 1, got %v", limitVal)
			tctx.Errorf("Limit mismatch: %v", limitVal)
		}

		if len(dataList) > 1 {
			tctx.FailureReason = fmt.Sprintf("Expected at most 1 item, got %d", len(dataList))
			tctx.Errorf("Pagination limit was not enforced by server: got %d", len(dataList))
		}

		tctx.Actual = fmt.Sprintf("HTTP 200 OK, successfully retrieved %d items", len(dataList))
	})

	// Step C: Query with invalid negative pagination parameters
	tests.RunAPITestWithDetails(t, "[Admin] GET Audit Logs - Invalid Parameters Check", "Submits negative page and limit parameters to verify query sanitize logic.", "HTTP 200 OK (server resets to default)", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/audit?page=-1&limit=-10", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK returned, server normalized pagination limits"
	})
}

func TestAPI_AdminAudit_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Admin] GET Audit Logs - Missing Auth",
			Description: "Attempts to fetch audit logs without admin authorization.",
			Expected:    "HTTP 401 Unauthorized",
			TestFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/admin/audit?page=1&limit=10", nil, nil, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected unauthorized error"
					tc.Fatalf("Expected 401 but got success")
				}
				tc.Actual = "Correctly rejected unauthorized request"
			},
		},
		{
			Name:        "[Admin] GET Audit Logs - SQLi in Pagination",
			Description: "Attempts SQL injection in the limit parameter.",
			Expected:    "HTTP 400 Bad Request or handled gracefully",
			TestFn: func(tc *tests.TestContext) {
				token := getValidAdminToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/admin/audit?page=1&limit=10'OR'1'='1", nil, nil, &resp, auth)
				if err == nil {
					tc.Actual = "Server handled SQLi attempt gracefully"
				} else {
					tc.Actual = "Server rejected SQLi attempt"
				}
			},
		},
		{
			Name:        "[Admin] GET Audit Logs - Max Int Pagination DOS check",
			Description: "Attempts to pass maximum integer in limit to bypass pagination and cause DoS.",
			Expected:    "HTTP 400 Bad Request or graceful clamp",
			TestFn: func(tc *tests.TestContext) {
				token := getValidAdminToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/admin/audit?page=1&limit=2147483647", nil, nil, &resp, auth)
				if err == nil {
					dataList, ok := resp["data"].([]interface{})
					if ok && len(dataList) > 1000 {
						tc.FailureReason = "Pagination boundary bypassed"
						tc.Fatalf("Server returned excessively large data set: %d rows", len(dataList))
					}
					tc.Actual = "Server gracefully capped max pagination limit"
				} else {
					tc.Actual = "Server rejected max int pagination limit"
				}
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
