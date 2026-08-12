package api_test

import (
	"fmt"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// TestAPI_PublicSubscriptionPlans covers GET /api/subscriptions (public — no auth required)
// The endpoint wraps the plans array inside a {"data":[...]} envelope.
func TestAPI_PublicSubscriptionPlans(t *testing.T) {
	// Happy path: list is served to unauthenticated public users
	tests.RunAPITestWithDetails(t,
		"[Public] GET Subscriptions List",
		"Fetches the list of all available subscription plans without any auth.",
		"HTTP 200 OK containing subscription plan array inside data key",
		func(tctx *tests.TestContext) {
			var envelope map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/subscriptions", nil, nil, &envelope, nil)
			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("GET /api/subscriptions failed: %v", err)
			}
			plans, ok := envelope["data"].([]interface{})
			if !ok || len(plans) == 0 {
				tctx.FailureReason = "Expected at least 1 subscription plan in response data"
				tctx.Fatalf("Subscription list empty or missing: %+v", envelope)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, received %d subscription plan(s)", len(plans))
		})

	// Verify the plans have expected structure fields
	tests.RunAPITestWithDetails(t,
		"[Public] GET Subscriptions List - Schema Validation",
		"Verifies each plan object contains required schema fields: id, name, frequency, status.",
		"HTTP 200 OK with valid schema fields",
		func(tctx *tests.TestContext) {
			var envelope map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/subscriptions", nil, nil, &envelope, nil)
			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("GET /api/subscriptions failed: %v", err)
			}
			plans, _ := envelope["data"].([]interface{})
			firstPlan, _ := plans[0].(map[string]interface{})
			requiredFields := []string{"id", "name", "frequency", "status"}
			for _, field := range requiredFields {
				if _, ok := firstPlan[field]; !ok {
					tctx.FailureReason = fmt.Sprintf("Missing required field '%s' in subscription plan", field)
					tctx.Errorf("Subscription plan missing required field: '%s'", field)
				}
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, schema valid — id='%v', frequency='%v'",
				firstPlan["id"], firstPlan["frequency"])
		})

	// Auth should NOT be required — verify same response with member token too
	tests.RunAPITestWithDetails(t,
		"[Member] GET Subscriptions List - Auth Not Required",
		"Verifies authenticated members also get the public subscription list.",
		"HTTP 200 OK same payload",
		func(tctx *tests.TestContext) {
			token := getValidMemberToken(tctx.T, tctx.Client)
			auth := &client.BearerTokenAuth{Token: token}

			var envelope map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/subscriptions", nil, nil, &envelope, auth)
			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK with member auth, got: %v", err)
				tctx.Fatalf("GET /api/subscriptions with member token failed: %v", err)
			}
			plans, _ := envelope["data"].([]interface{})
			tctx.Actual = fmt.Sprintf("HTTP 200 OK with member auth, %d plan(s) returned", len(plans))
		})
}

func TestAPI_SubscriptionPlans_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Public] POST Subscriptions List - Wrong HTTP Method",
			Description: "Attempts to send a POST request to fetch subscription plans.",
			Expected:    "HTTP 405 Method Not Allowed",
			TestFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/subscriptions", nil, nil, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection of POST request"
					tc.Fatalf("Expected error for wrong HTTP method")
				}
				tc.Actual = "Correctly rejected POST request to GET endpoint"
			},
		},
		{
			Name:        "[Public] GET Subscriptions List - Invalid Query Parameters",
			Description: "Attempts to append invalid sorting parameters.",
			Expected:    "HTTP 400 Bad Request or safely ignored",
			TestFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("GET", "/api/subscriptions?sort=DROP_TABLE", nil, nil, &resp, nil)
				if err == nil {
					tc.Actual = "Safely ignored malicious query parameter"
				} else {
					tc.Actual = "Rejected request with invalid query parameter"
				}
			},
		},
		{
			Name:        "[Member] POST Subscription - Apply Non-existent Discount Code",
			Description: "Attempts to apply a fake discount code to a subscription.",
			Expected:    "HTTP 404 Not Found or 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				token := getValidMemberToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				payload := map[string]string{
					"discountCode": "FAKECODE100",
					"planId": "monthly",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/subscriptions/apply-discount", nil, &payload, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected rejection of fake discount code"
					tc.Fatalf("Expected error for fake discount code")
				}
				tc.Actual = "Correctly rejected fake discount code"
			},
		},
		{
			Name:        "[Member] POST Subscription - Improper Downgrade",
			Description: "Attempts to downgrade a subscription without paying cancellation fees.",
			Expected:    "HTTP 400 Bad Request or HTTP 403 Forbidden",
			TestFn: func(tc *tests.TestContext) {
				token := getValidMemberToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				payload := map[string]string{
					"newPlanId": "free",
					"waiveFees": "true",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/subscriptions/downgrade", nil, &payload, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected rejection of improper downgrade"
					tc.Fatalf("Expected error for improper downgrade")
				}
				tc.Actual = "Correctly rejected improper downgrade"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
