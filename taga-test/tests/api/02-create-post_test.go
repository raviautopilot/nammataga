package api_tests

import (
	"fmt"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// GrievancePayload defines the payload for creating a grievance
type GrievancePayload struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

// TestAPI_02_CreatePost verifies creation of a grievance payload.
func TestAPI_02_CreatePost(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Create Grievance Payload",
		"Verifies creating a new grievance payload with custom bearer authentication token.",
		"HTTP 200 OK or 201 Created with successful status message",
		func(tc *tests.TestContext) {
			payload := &GrievancePayload{
				Name:        "Automated Test Member",
				Email:       "testmember@taga-tn.org",
				Phone:       "9876543210",
				Category:    "General",
				Priority:    "Medium",
				Subject:     "E2E Automation Test Grievance",
				Description: "Verifying API framework creation wrapper with custom auth headers",
			}
			var response map[string]interface{}
			auth := &client.BearerTokenAuth{Token: "super-secret-e2e-token"}

			err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, auth)
			if err != nil {
				tc.FailureReason = fmt.Sprintf("Failed to create grievance: %v", err)
				tc.Fatalf("Failed to create grievance: %v", err)
			}

			tc.Actual = fmt.Sprintf("Successfully processed with response: %v", response)
		},
	)
}
