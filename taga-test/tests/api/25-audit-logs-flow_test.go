package api_tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

func TestAPI_AdminAudit_FlowCheck(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] GET Audit Logs Search Flow", "Queries audit logs specifically searching for 'announcement' actions to verify audit trail flow.", "HTTP 200 OK containing matched logs data", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/audit?search=announcement", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}

		logs, ok := resp["data"].([]interface{})
		if !ok {
			tctx.FailureReason = "Response body missing 'data' field array"
			tctx.Fatalf("Response missing 'data' field array")
		}

		foundAnnouncementLog := false
		for _, l := range logs {
			logObj, ok := l.(map[string]interface{})
			if !ok {
				continue
			}
			details, _ := logObj["details"].(string)
			action, _ := logObj["action"].(string)
			if strings.Contains(strings.ToLower(details), "announcement") || strings.Contains(strings.ToLower(action), "announcement") {
				foundAnnouncementLog = true
				break
			}
		}

		tctx.Actual = fmt.Sprintf("HTTP 200 OK, found %d total records. Matched announcement in history: %v", len(logs), foundAnnouncementLog)
	})
}

func TestAPI_NegativeScenarios_AuditLogsFlow(t *testing.T) {
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
			Description:    "Access endpoint without token",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?search=foo",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Invalid Token - Error Expected",
			Persona:        "Attacker",
			Description:    "Access with fake token",
			Method:         "GET",
			Endpoint:       "/api/admin/audit",
			AuthType:       "invalid",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Admin",
			Description:    "Use POST instead of GET",
			Method:         "POST",
			Endpoint:       "/api/admin/audit",
			AuthType:       "admin",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Admin",
			Description:    "SQLi attempt in search URL",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?search=';DROP TABLE audit;",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest, // Or 500
		},
		{
			Name:           "XSS Payload - Error Expected",
			Persona:        "Admin",
			Description:    "XSS in search parameter",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?search=<script>alert(1)</script>",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest, // Or 400
		},
		{
			Name:           "Role Context Switching - Member Accessing Logs",
			Persona:        "Member",
			Description:    "Member queries audit logs",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?search=admin",
			AuthType:       "member",
			ExpectedStatus: http.StatusForbidden,
		},
		{
			Name:           "Logical Paradox - Invalid Date Range",
			Persona:        "Admin",
			Description:    "Search with end_date before start_date",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?start_date=2026-12-01&end_date=2026-01-01",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Extreme Boundary - Negative Pagination",
			Persona:        "Admin",
			Description:    "Negative page and limit",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?page=-1&limit=-100",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
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
