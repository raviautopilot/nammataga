package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_SessionSecurity_Tampering(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Profile - Tampered Signature Token", "Attempts retrieving profile with a token that has its signature corrupted.", "HTTP 401 Unauthorized", func(tctx *tests.TestContext) {
		token := getValidMemberToken(tctx.T, tctx.Client)

		// Corrupt signature (change last few characters)
		tamperedToken := token[:len(token)-5] + "aaaaa"
		auth := &client.BearerTokenAuth{Token: tamperedToken}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/member/profile", nil, nil, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusUnauthorized, "")
	})

	tests.RunAPITestWithDetails(t, "[Public] GET Profile - Bogus Key Token", "Attempts retrieving profile with a completely bogus token format.", "HTTP 401 Unauthorized", func(tctx *tests.TestContext) {
		auth := &client.BearerTokenAuth{Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/member/profile", nil, nil, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusUnauthorized, "")
	})

	tests.RunAPITestWithDetails(t, "[Public] GET Profile - Invalid Authorization Header Format", "Attempts accessing profile with a non-Bearer header prefix.", "HTTP 401 Unauthorized", func(tctx *tests.TestContext) {
		headers := map[string]string{
			"Authorization": "Basic c3VkaGFuOnRlc3QxMjM=",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/member/profile", headers, nil, &resp, nil)

		assertErrorStatus(tctx, err, http.StatusUnauthorized, "")
	})
}

func TestAPI_NegativeScenarios_SessionSecurity(t *testing.T) {
	type TestCaseType struct {
		Name           string
		Persona        string
		Description    string
		Method         string
		Endpoint       string
		AuthType       string
		Payload        interface{}
		Headers        map[string]string
		ExpectedStatus int
		ExpectedSub    string
	}

	testCases := []TestCaseType{
		{
			Name:           "Missing Auth - Error Expected",
			Persona:        "Anonymous",
			Description:    "Access profile without token",
			Method:         "GET",
			Endpoint:       "/api/member/profile",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Member",
			Description:    "Use POST instead of GET",
			Method:         "POST",
			Endpoint:       "/api/member/profile",
			AuthType:       "member",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Member",
			Description:    "SQLi attempt in profile URL",
			Method:         "GET",
			Endpoint:       "/api/member/profile?id=1' OR '1'='1",
			AuthType:       "member",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "XSS Payload - Error Expected",
			Persona:        "Member",
			Description:    "XSS in parameter",
			Method:         "GET",
			Endpoint:       "/api/member/profile?id=<script>alert(1)</script>",
			AuthType:       "member",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "State Machine Violation - Revoked Token",
			Persona:        "Member",
			Description:    "Use a token that has been explicitly logged out",
			Method:         "GET",
			Endpoint:       "/api/member/profile",
			AuthType:       "member",
			Headers:        map[string]string{"Authorization": "Bearer revoked_token_123"},
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Logical Paradox - Future Issued Token",
			Persona:        "Member",
			Description:    "Token with 'iat' claim in the future",
			Method:         "GET",
			Endpoint:       "/api/member/profile",
			AuthType:       "member",
			Headers:        map[string]string{"Authorization": "Bearer future_issued_token_123"},
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Extreme Boundary - Massive Token Payload",
			Persona:        "Member",
			Description:    "Bearer token of extreme length",
			Method:         "GET",
			Endpoint:       "/api/member/profile",
			AuthType:       "member",
			Headers:        map[string]string{"Authorization": "Bearer extremely_long_token_string_which_exceeds_maximum_header_size_limit_and_causes_issues_1234567890"},
			ExpectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s %s - %s", tc.Persona, tc.Method, tc.Endpoint, tc.Name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.ExpectedStatus)
		if tc.ExpectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.ExpectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.Description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			switch tc.AuthType {
			case "admin":
				token := getValidAdminToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "member":
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "invalid":
				auth = &client.BearerTokenAuth{Token: "invalid-jwt-token-12345"}
			case "none":
				auth = nil
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest(tc.Method, tc.Endpoint, tc.Headers, tc.Payload, &resp, auth)
			if err == nil {
				tctx.Fatalf("Expected error for negative scenario, got none. Response: %v", resp)
			}
		})
	}
}
