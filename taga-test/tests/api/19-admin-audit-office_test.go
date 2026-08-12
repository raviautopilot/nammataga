package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type BearerInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Contact string `json:"contact"`
}

// ============================================================================
// 1. Audit Logs Queries - Parameterized Tests
// ============================================================================

func TestAPI_AdminAudit_Queries(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		endpoint       string
		authType       string // "none", "member", "admin"
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Fetch Audit Logs",
			persona:        "Legitimate Admin",
			description:    "Retrieves full administrative audit log records.",
			endpoint:       "/api/admin/audit?page=1&limit=20",
			authType:       "admin",
			expectedStatus: http.StatusOK,
			expectedSub:    "data",
		},
		{
			name:           "Happy Path - Fetch Audit Users dropdown",
			persona:        "Legitimate Admin",
			description:    "Retrieves list of active administrative users for dropdown.",
			endpoint:       "/api/admin/audit/users?year=2026&month=08",
			authType:       "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Security - Get Audit Logs Anonymous",
			persona:        "Anonymous Guest",
			description:    "Attempts querying audit logs without Bearer JWT.",
			endpoint:       "/api/admin/audit",
			authType:       "none",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET %s - %s", tc.persona, tc.endpoint, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			switch tc.authType {
			case "admin":
				token := getValidAdminToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "member":
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "none":
				auth = nil
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", tc.endpoint, nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = "HTTP 200 OK retrieved successfully"
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 2. Office Bearers Management - Parameterized Tests
// ============================================================================

func TestAPI_AdminOffice_BearersManagement(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] GET Office Districts List", "Admin fetches all districts with office bearer info.", "HTTP 200 OK containing districts", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/office-bearers/districts", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK: %v", resp)
	})

	tests.RunAPITestWithDetails(t, "[Admin] GET Salem Bearers List", "Admin fetches office bearers list for Salem district.", "HTTP 200 OK containing bearers list", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/office-bearers/district/salem", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK: %v", resp)
	})

	tests.RunAPITestWithDetails(t, "[Admin] PUT Update Salem Bearers List - Happy Path", "Admin updates the district office bearers details for Salem.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		payload := []BearerInfo{
			{Name: "Salem President", Title: "President", Contact: "9944637251"},
			{Name: "Salem Secretary", Title: "Secretary", Contact: "9944637252"},
			{Name: "Salem Treasurer", Title: "Treasurer", Contact: "9944637253"},
			{Name: "Salem Vice President", Title: "Vice President", Contact: "9944637254"},
			{Name: "Salem Joint Secretary", Title: "Joint Secretary", Contact: "9944637255"},
			{Name: "Salem Executive Member", Title: "Executive Member", Contact: "9944637256"},
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("PUT", "/api/admin/office-bearers/district/salem", nil, &payload, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK updated Salem district bearers successfully"
	})

	tests.RunAPITestWithDetails(t, "[Admin] GET Office Bearers Backups", "Admin lists active district bearers database backup files.", "HTTP 200 OK containing backup paths list", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/admin/office-bearers/backups", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK backups: %v", resp)
	})
}

// ============================================================================
// 3. Reminders, Password Init, and Legacy Uploads
// ============================================================================

func TestAPI_AdminAudit_UtilityHandlers(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] POST Send Renewal Reminders - Manual Trigger", "Admin manually initiates subscription renewal reminders.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/send-renewal-reminders", nil, nil, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK reminders processing completed"
	})

	tests.RunAPITestWithDetails(t, "[Admin] POST Reset Passwords - Validation Error", "Admin submits request to reset passwords without specifying member ID parameter.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/init-password", nil, nil, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusBadRequest, "Please specify memberId parameter")
	})

	tests.RunAPITestWithDetails(t, "[Admin] POST Reset Passwords - Not Implemented", "Admin requests resetting password for a specific member ID.", "HTTP 501 Not Implemented", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/init-password?memberId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, nil, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusNotImplemented, "Reset password for specific member is not yet implemented")
	})

	tests.RunAPITestWithDetails(t, "[Admin] POST Legacy Registration Upload - Empty Payload", "Admin triggers legacy registration upload without form file.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/admin/upload-registration", nil, nil, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusBadRequest, "File upload failed")
	})
}

func TestAPI_NegativeScenarios_AdminAuditOffice(t *testing.T) {
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
			Endpoint:       "/api/admin/audit",
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
			Persona:        "User",
			Description:    "Use wrong method for endpoint",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/districts",
			AuthType:       "admin",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Attacker",
			Description:    "SQLi attempt in URL",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?page=1' OR '1'='1",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "XSS Payload in Path - Error Expected",
			Persona:        "Attacker",
			Description:    "XSS script injection in endpoint",
			Method:         "GET",
			Endpoint:       "/api/admin/audit/<script>alert(1)</script>",
			AuthType:       "admin",
			ExpectedStatus: http.StatusNotFound,
		},
		{
			Name:           "Malformed JSON Payload - Error Expected",
			Persona:        "User",
			Description:    "Send non-JSON data where expected",
			Method:         "PUT",
			Endpoint:       "/api/admin/office-bearers/district/salem",
			AuthType:       "admin",
			Payload:        "not-a-json-payload-string",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Boundary Values / Missing Fields - Error Expected",
			Persona:        "User",
			Description:    "Send empty object",
			Method:         "PUT",
			Endpoint:       "/api/admin/office-bearers/district/salem",
			AuthType:       "admin",
			Payload:        map[string]interface{}{},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Role Context Switching - Member Accessing Admin Audit",
			Persona:        "Member",
			Description:    "A normal member tries to fetch admin audit logs",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?page=1&limit=20",
			AuthType:       "member",
			ExpectedStatus: http.StatusForbidden,
		},
		{
			Name:           "Extreme Boundary - Max Integer Page Limit",
			Persona:        "Admin",
			Description:    "Requesting audit logs with max int64 page limit",
			Method:         "GET",
			Endpoint:       "/api/admin/audit?page=1&limit=9223372036854775807",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Logical Paradox - Nullified Office Bearers",
			Persona:        "Admin",
			Description:    "Updating bearers with negative contacts and empty names",
			Method:         "PUT",
			Endpoint:       "/api/admin/office-bearers/district/salem",
			AuthType:       "admin",
			Payload:        []map[string]string{{"name": "", "title": "President", "contact": "-999999999"}},
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
