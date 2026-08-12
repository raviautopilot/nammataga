package api_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_DatabaseResiliency_CorruptInputs(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Member] PUT Profile - Corrupt JSON schema", "Submits raw malformed JSON string payload to profile update endpoint.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidMemberToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		corruptPayload := []byte(`{"address": "Salem", "contactPhone": }`) // Corrupt JSON

		reqUrl := tctx.Client.BaseURL + "/api/member/profile"
		req, err := http.NewRequest("PUT", reqUrl, bytes.NewReader(corruptPayload))
		if err != nil {
			tctx.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if err := auth.Apply(req); err != nil {
			tctx.Fatalf("Failed to apply auth: %v", err)
		}

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			tctx.FailureReason = fmt.Sprintf("Expected 400 Bad Request, got: %d", resp.StatusCode)
			tctx.Fatalf("Expected 400 Bad Request, got: %d", resp.StatusCode)
		}

		tctx.Actual = "HTTP 400 Bad Request correctly returned on corrupt JSON"
	})

	tests.RunAPITestWithDetails(t, "[Public] POST Grievance - Corrupt JSON schema", "Submits malformed JSON schema to grievance submission endpoint.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		corruptPayload := []byte(`{"subject": "Delay", `) // Malformed JSON

		reqUrl := tctx.Client.BaseURL + "/api/grievances"
		req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(corruptPayload))
		if err != nil {
			tctx.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			tctx.FailureReason = fmt.Sprintf("Expected 400 Bad Request, got: %d", resp.StatusCode)
			tctx.Fatalf("Expected 400 Bad Request, got: %d", resp.StatusCode)
		}

		tctx.Actual = "HTTP 400 Bad Request correctly returned on corrupt grievance input"
	})
}

func TestAPI_DatabaseResiliency_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Member] PUT Profile - Invalid Token",
			Description: "Attempts to update profile with an invalid JWT token.",
			Expected:    "HTTP 401 Unauthorized",
			TestFn: func(tc *tests.TestContext) {
				auth := &client.BearerTokenAuth{Token: "invalid_token_123"}
				payload := map[string]string{"address": "Chennai"}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("PUT", "/api/member/profile", nil, &payload, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected error with invalid token"
					tc.Fatalf("Expected 401 but got success")
				}
				tc.Actual = "Correctly rejected invalid token"
			},
		},
		{
			Name:        "[Public] POST Grievance - XSS Payload",
			Description: "Submits a grievance containing malicious XSS in the subject.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				payload := map[string]string{
					"subject": "<script>alert('XSS')</script>",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, &payload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected error for XSS payload"
					tc.Fatalf("Expected rejection but got success")
				}
				tc.Actual = "Correctly rejected XSS payload"
			},
		},
		{
			Name:        "[Public] POST Grievance - Future Date Business Logic Contradiction",
			Description: "Submits a grievance with an incident date set in the future.",
			Expected:    "HTTP 400 Bad Request or HTTP 422 Unprocessable Entity",
			TestFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"subject": "Future Incident",
					"incidentDate": "3000-01-01", // 1 year in future
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, &payload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for future incident date"
					tc.Fatalf("Expected error for future date but got success")
				}
				tc.Actual = "Correctly rejected future incident date"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
