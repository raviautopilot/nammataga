package api_test

import (
	"fmt"
	"testing"

	"e2e-template/tests"
)

type PointerValidationTestCase struct {
	Name        string
	TargetURL   string
	Body        interface{}
	RespObj     interface{}
	Description string
	Expected    string
	ValidateFn  func(tc *tests.TestContext)
}

// TestAPI_03_PointerValidation_TableDriven enforces client-side pointer type safety validation via parameterized test cases.
func TestAPI_03_PointerValidation_TableDriven(t *testing.T) {
	testCases := []PointerValidationTestCase{
		{
			Name:        "Enforce Non-Pointer Request Body Safety Check",
			TargetURL:   "/api/grievances",
			Body:        GrievancePayload{Name: "Passed by value struct"},
			RespObj:     &map[string]interface{}{},
			Description: "Verifies that passing a value struct instead of a pointer to SendHttpRequest returns a validation error.",
			Expected:    "Validation error: 'HTTP Error: status=0, err=request body must be a pointer to a struct/value'",
			ValidateFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				valuePayload := GrievancePayload{Name: "Passed by value struct"}

				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, valuePayload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected SendHttpRequest to fail when passed a value struct argument"
					tc.Fatalf("Expected SendHttpRequest to fail on value struct argument")
				}

				expectedErr := "HTTP Error: status=0, err=request body must be a pointer to a struct/value"
				tc.Actual = err.Error()

				if err.Error() != expectedErr {
					tc.FailureReason = fmt.Sprintf("Expected error '%s', got '%s'", expectedErr, err.Error())
					tc.Errorf("Expected validation error '%s', got '%s'", expectedErr, err.Error())
				}
			},
		},
		{
			Name:        "Enforce Non-Pointer Response Object Safety Check",
			TargetURL:   "/api/public/about",
			Body:        nil,
			RespObj:     AboutResponse{}, // Passed by value instead of &AboutResponse{}!
			Description: "Verifies that passing a non-pointer response object to SendHttpRequest returns a JSON unmarshal error.",
			Expected:    "Validation or unmarshal error when response object is not a pointer",
			ValidateFn: func(tc *tests.TestContext) {
				var valueResp AboutResponse
				err := tc.Client.SendHttpRequest("GET", "/api/public/about", nil, nil, valueResp, nil)
				if err == nil {
					tc.FailureReason = "Expected SendHttpRequest to fail when response object is passed by value"
					tc.Fatalf("Expected SendHttpRequest to fail when response object is passed by value")
				}
				tc.Actual = fmt.Sprintf("Intercepted expected error: %v", err)
			},
		},
		{
			Name:        "Nil Request Body Safety Check",
			TargetURL:   "/api/grievances",
			Body:        nil,
			RespObj:     &map[string]interface{}{},
			Description: "Verifies handling of nil request body when endpoint expects payload.",
			Expected:    "Validation error or Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, nil, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected SendHttpRequest to fail or return 400 when body is nil"
					tc.Fatalf("Expected SendHttpRequest to fail or return 400 when body is nil")
				}
				tc.Actual = fmt.Sprintf("Intercepted expected error: %v", err)
			},
		},
		{
			Name:        "Nil Response Object Safety Check",
			TargetURL:   "/api/public/about",
			Body:        nil,
			RespObj:     nil,
			Description: "Verifies handling of nil response object.",
			Expected:    "Validation error or success if response body ignored",
			ValidateFn: func(tc *tests.TestContext) {
				err := tc.Client.SendHttpRequest("GET", "/api/public/about", nil, nil, nil, nil)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("GET request failed: %v", err)
					tc.Fatalf("GET request failed: %v", err)
				}
				tc.Actual = "Successfully ignored response body with nil response object"
			},
		},
		{
			Name:        "Business Logic - IDOR Pointer Validation Safety",
			TargetURL:   "/api/member/profile?user_id=9999", // Trying to access another user's profile
			Body:        nil,
			RespObj:     &map[string]interface{}{},
			Description: "Verifies that passing another user's ID (IDOR attempt) is rejected and doesn't crash pointer logic.",
			Expected:    "HTTP 401/403 or pointer safely ignored",
			ValidateFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				// We do an unauthenticated call to a protected route with an IDOR payload
				err := tc.Client.SendHttpRequest("GET", "/api/member/profile?user_id=9999", nil, nil, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected error for unauthenticated/IDOR request, got success"
					tc.Fatalf("Expected error for unauthenticated/IDOR request, got success")
				}
				tc.Actual = fmt.Sprintf("Handled IDOR gracefully with error: %v", err)
			},
		},
		{
			Name:        "State Machine Violation - Update Cancelled Pointer",
			TargetURL:   "/api/grievances/123/update?status=cancelled",
			Body:        &map[string]interface{}{"status": "processing"},
			RespObj:     &map[string]interface{}{},
			Description: "Verifies that updating a logically cancelled pointer state machine safely fails without panic.",
			Expected:    "HTTP 400/404 or cleanly rejected",
			ValidateFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				body := map[string]interface{}{"status": "processing"}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances/123/update?status=cancelled", nil, &body, &resp, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Expected non-500 error for state machine violation, got 500"
					tc.Fatalf("Expected non-500 error for state machine violation, got 500")
				}
				tc.Actual = fmt.Sprintf("Handled gracefully with error: %v", err)
			},
		},
		{
			Name:        "Extreme Boundary - Pointer Max Integer Slice Allocation",
			TargetURL:   "/api/public/about/stats",
			Body:        nil,
			RespObj:     &[]StatsResponse{},
			Description: "Verifies extreme integer boundaries for pointer slice allocation limit does not crash.",
			Expected:    "HTTP 400/200 safely handled",
			ValidateFn: func(tc *tests.TestContext) {
				var resp []StatsResponse
				err := tc.Client.SendHttpRequest("GET", "/api/public/about/stats?limit=9223372036854775807", nil, nil, &resp, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on extreme pointer slice allocation boundary"
					tc.Fatalf("Server crashed on extreme pointer slice allocation boundary")
				}
				tc.Actual = "Safely handled pointer allocation"
			},
		},
		{
			Name:        "Logical Paradox - Nested Pointer with Contradicting Flags",
			TargetURL:   "/api/grievances",
			Body:        &map[string]interface{}{"is_active": true, "is_deleted": true},
			RespObj:     &map[string]interface{}{},
			Description: "Verifies logical paradox in payload pointer fields.",
			Expected:    "Validation error or safely ignored",
			ValidateFn: func(tc *tests.TestContext) {
				var resp map[string]interface{}
				body := map[string]interface{}{"is_active": true, "is_deleted": true}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, &body, &resp, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on logical paradox pointer validation"
					tc.Fatalf("Server crashed on logical paradox pointer validation")
				}
				tc.Actual = fmt.Sprintf("Handled paradox gracefully: %v", err)
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
