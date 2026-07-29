package api_tests

import (
	"fmt"
	"testing"

	"e2e-template/tests"
)

// TestAPI_03_PointerValidation enforces client-side pointer type safety validation.
func TestAPI_03_PointerValidation(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Enforce Pointer Type Safety Check",
		"Verifies that passing a value struct instead of a pointer to SendHttpRequest returns a validation error.",
		"Validation error: 'HTTP Error: status=0, err=request body must be a pointer to a struct/value'",
		func(tc *tests.TestContext) {
			valuePayload := GrievancePayload{
				Name: "Passed by value struct",
			}
			var resp map[string]interface{}

			// Intentionally pass a value struct instead of pointer for request body
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
	)
}
