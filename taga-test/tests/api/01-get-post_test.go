package api_tests

import (
	"fmt"
	"testing"

	"e2e-template/tests"
)

// TestAPI_01_GetPost verifies retrieval of organization info using the rich test wrapper framework.
func TestAPI_01_GetPost(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Get About Details",
		"Verifies retrieving organization metadata from /api/public/about.",
		"HTTP 200 OK with valid non-empty AboutResponse ID and Name",
		func(tc *tests.TestContext) {
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
	)
}
