package api_tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// AdminLoginRequest defines the payload sent to POST /api/admin/login
type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AdminLoginResponse defines the structure returned by POST /api/admin/login
type AdminLoginResponse struct {
	Token     string `json:"token"`
	Role      string `json:"role"`
	ExpiresIn int64  `json:"expires_in"`
}

// CommonAPIErrorResponse captures standard error and message payloads
type CommonAPIErrorResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Helper function to obtain a valid admin token for authenticated tests
func getValidAdminToken(t *testing.T, c *client.Client) string {
	adminCreds := tests.GlobalConfig.AdminCredentials
	req := &AdminLoginRequest{
		Username: adminCreds.Username,
		Password: adminCreds.Password,
	}
	var resp AdminLoginResponse
	err := c.SendHttpRequest("POST", "/api/admin/login", nil, req, &resp, nil)
	if err != nil {
		t.Fatalf("Setup failure: unable to obtain admin token for test: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("Setup failure: admin login returned empty token")
	}
	return resp.Token
}

// ============================================================================
// 1. POST /api/admin/login - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_AdminLogin_TableDriven(t *testing.T) {
	adminCreds := tests.GlobalConfig.AdminCredentials

	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
		expectToken    bool
		expectedErrSub string
	}{
		{
			name:           "Happy Path - Valid Admin Credentials",
			persona:        "Legitimate Admin User",
			description:    "Submits correct admin credentials configured in system to obtain JWT.",
			payload:        &AdminLoginRequest{Username: adminCreds.Username, Password: adminCreds.Password},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "Auth Failure - Invalid Password",
			persona:        "Intruder with Wrong Password",
			description:    "Submits valid admin username with an incorrect password.",
			payload:        &AdminLoginRequest{Username: adminCreds.Username, Password: "IncorrectPassword123!"},
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid credentials",
		},
		{
			name:           "Auth Failure - Nonexistent Username",
			persona:        "Intruder with Unknown Account",
			description:    "Submits unknown email address with password.",
			payload:        &AdminLoginRequest{Username: "unknown_admin@test.com", Password: "anyPassword"},
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid credentials",
		},
		{
			name:           "Validation - Empty Payload",
			persona:        "Buggy Client",
			description:    "Submits empty JSON body violating required field bindings.",
			payload:        &AdminLoginRequest{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation - Missing Password Field",
			persona:        "Malformed Client",
			description:    "Submits username only, omitting password.",
			payload:        &map[string]string{"username": adminCreds.Username},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation - Missing Username Field",
			persona:        "Malformed Client",
			description:    "Submits password only, omitting username.",
			payload:        &map[string]string{"password": adminCreds.Password},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - SQL Injection in Username",
			persona:        "Malicious Security Auditor",
			description:    "Attempts SQL injection payload in username field.",
			payload:        &AdminLoginRequest{Username: "' OR '1'='1' --", Password: "any"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Security - XSS Script Tag in Username",
			persona:        "Malicious Security Auditor",
			description:    "Attempts XSS payload in username field.",
			payload:        &AdminLoginRequest{Username: "<script>alert('xss')</script>", Password: "any"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Boundary - Max Length Password",
			persona:        "Attacker with Long Password",
			description:    "Submits extremely long password string.",
			payload:        &AdminLoginRequest{Username: adminCreds.Username, Password: strings.Repeat("A", 1024)},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Security - XSS in Password",
			persona:        "Malicious Security Auditor",
			description:    "Attempts XSS payload in password field.",
			payload:        &AdminLoginRequest{Username: adminCreds.Username, Password: "<script>alert('xss')</script>"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Business Logic - Admin Login IDOR/Role Escalation",
			persona:        "Member Trying to be Admin",
			description:    "Submits valid member credentials to admin login endpoint.",
			payload:        &AdminLoginRequest{Username: "member@taga-tn.org", Password: "memberPassword123"},
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid credentials",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectToken {
			expectedStr += " with valid JWT token and role 'admin'"
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp AdminLoginResponse
			var errResp CommonAPIErrorResponse

			var rawPtr interface{}
			if tc.expectedStatus == http.StatusOK {
				rawPtr = &resp
			} else {
				rawPtr = &errResp
			}

			err := tctx.Client.SendHttpRequest("POST", "/api/admin/login", nil, tc.payload, rawPtr, nil)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected HTTP 200, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got error: %v", err)
				}
				if resp.Token == "" {
					tctx.FailureReason = "Expected non-empty JWT token"
					tctx.Errorf("Admin login response missing token")
				}
				if resp.Role != "admin" {
					tctx.FailureReason = fmt.Sprintf("Expected role 'admin', got '%s'", resp.Role)
					tctx.Errorf("Expected role 'admin', got '%s'", resp.Role)
				}
				if resp.ExpiresIn <= 0 {
					tctx.FailureReason = fmt.Sprintf("Expected positive expiresIn, got %d", resp.ExpiresIn)
					tctx.Errorf("Expected positive expiresIn, got %d", resp.ExpiresIn)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Role='%s', TokenLength=%d", resp.Role, len(resp.Token))
			} else {
				if err == nil {
					tctx.FailureReason = fmt.Sprintf("Expected HTTP %d, but request succeeded with 200 OK", tc.expectedStatus)
					tctx.Fatalf("Expected error status %d, got 200 OK", tc.expectedStatus)
				}
				if err.StatusCode() != tc.expectedStatus {
					tctx.FailureReason = fmt.Sprintf("Expected HTTP %d, got HTTP %d", tc.expectedStatus, err.StatusCode())
					tctx.Errorf("Expected HTTP %d, got %d", tc.expectedStatus, err.StatusCode())
				}
				if tc.expectedErrSub != "" && !strings.Contains(err.ResponseBody(), tc.expectedErrSub) {
					tctx.FailureReason = fmt.Sprintf("Expected error substring '%s', got response: %s", tc.expectedErrSub, err.ResponseBody())
					tctx.Errorf("Expected error substring '%s', got response: %s", tc.expectedErrSub, err.ResponseBody())
				}
				tctx.Actual = fmt.Sprintf("HTTP %d correctly returned with response: %s", err.StatusCode(), err.ResponseBody())
			}
		})
	}
}

// TestAPI_AdminLogin_HTTPMethods verifies method rejection for /api/admin/login.
func TestAPI_AdminLogin_HTTPMethods(t *testing.T) {
	unsupportedMethods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range unsupportedMethods {
		method := method
		testName := fmt.Sprintf("Reject Unsupported Method %s on /api/admin/login", method)
		desc := fmt.Sprintf("Ensures calling %s on POST-only endpoint /api/admin/login returns 404/405.", method)
		expected := "HTTP 404 Not Found or 405 Method Not Allowed"

		tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
			err := tc.Client.SendHttpRequest(method, "/api/admin/login", nil, nil, nil, nil)
			if err == nil {
				tc.FailureReason = fmt.Sprintf("Expected rejection for %s /api/admin/login, got 200 OK", method)
				tc.Fatalf("Expected rejection for %s /api/admin/login, got 200 OK", method)
			}
			statusCode := err.StatusCode()
			if statusCode != http.StatusNotFound && statusCode != http.StatusMethodNotAllowed {
				tc.FailureReason = fmt.Sprintf("Expected 404 or 405, got HTTP %d", statusCode)
				tc.Errorf("Expected 404 or 405, got HTTP %d", statusCode)
			}
			tc.Actual = fmt.Sprintf("Rejected cleanly with HTTP status %d", statusCode)
		})
	}
}

// TestAPI_AdminLogin_LatencySLA verifies login endpoint responds within acceptable SLA (< 1500ms).
func TestAPI_AdminLogin_LatencySLA(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Admin Login Performance SLA (< 1500ms)",
		"Validates that admin credential verification and JWT generation respond under 1500ms.",
		"HTTP 200 OK within 1500ms",
		func(tc *tests.TestContext) {
			adminCreds := tests.GlobalConfig.AdminCredentials
			req := &AdminLoginRequest{
				Username: adminCreds.Username,
				Password: adminCreds.Password,
			}
			var resp AdminLoginResponse

			start := time.Now()
			err := tc.Client.SendHttpRequest("POST", "/api/admin/login", nil, req, &resp, nil)
			latency := time.Since(start)

			if err != nil {
				tc.FailureReason = fmt.Sprintf("Admin login request failed: %v", err)
				tc.Fatalf("Admin login request failed: %v", err)
			}

			tc.Actual = fmt.Sprintf("HTTP 200 OK in %v", latency)
			if latency > 1500*time.Millisecond {
				tc.FailureReason = fmt.Sprintf("Latency SLA breach: took %v (max allowed: 1500ms)", latency)
				tc.Errorf("Latency SLA breach: took %v (max allowed: 1500ms)", latency)
			}
		},
	)
}

// ============================================================================
// 2. POST /api/admin/init-password - Parameterized Tests
// ============================================================================

func TestAPI_AdminInitPassword_TableDriven(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "invalid", "malformed", "admin"
		queryParams    string
		expectedStatus int
		expectedErrSub string
	}{
		{
			name:           "Security - No Authorization Header",
			persona:        "Anonymous Guest",
			description:    "Attempts to call init-password without any Authorization header.",
			authType:       "none",
			queryParams:    "?memberId=TAGA123",
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Authorization header required",
		},
		{
			name:           "Security - Malformed Authorization Format",
			persona:        "Attacker with Bad Auth Header",
			description:    "Sends Authorization header without 'Bearer ' prefix.",
			authType:       "malformed",
			queryParams:    "?memberId=TAGA123",
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid authorization format",
		},
		{
			name:           "Security - Invalid JWT Token",
			persona:        "Attacker with Forged Token",
			description:    "Sends Bearer header with bogus/tampered JWT.",
			authType:       "invalid",
			queryParams:    "?memberId=TAGA123",
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid or expired token",
		},
		{
			name:           "Validation - Missing memberId Parameter",
			persona:        "Authenticated Admin",
			description:    "Calls init-password with no query parameter (defaults to memberId=none).",
			authType:       "admin",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedErrSub: "Please specify memberId parameter",
		},
		{
			name:           "Validation - Explicit memberId=none",
			persona:        "Authenticated Admin",
			description:    "Calls init-password with explicit memberId=none query parameter.",
			authType:       "admin",
			queryParams:    "?memberId=none",
			expectedStatus: http.StatusBadRequest,
			expectedErrSub: "Please specify memberId parameter",
		},
		{
			name:           "Status - Specific Member Password Init Not Implemented",
			persona:        "Authenticated Admin",
			description:    "Calls init-password for a specific member ID (feature currently not implemented).",
			authType:       "admin",
			queryParams:    "?memberId=TAGA001",
			expectedStatus: http.StatusNotImplemented,
			expectedErrSub: "Reset password for specific member is not yet implemented",
		},
		{
			name:           "Security - SQL Injection in memberId",
			persona:        "Authenticated Admin",
			description:    "SQL injection attempt in memberId parameter.",
			authType:       "admin",
			queryParams:    "?memberId=1' OR '1'='1",
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "Business Logic - Negative ID for init-password",
			persona:        "Authenticated Admin",
			description:    "Submits negative memberId where positive is expected.",
			authType:       "admin",
			queryParams:    "?memberId=-100",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedErrSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedErrSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator

			switch tc.authType {
			case "admin":
				adminToken := getValidAdminToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: adminToken}
			case "invalid":
				auth = &client.BearerTokenAuth{Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.bogus.token"}
			case "malformed":
				headers := map[string]string{"Authorization": "Token 12345"}
				err := tctx.Client.SendHttpRequest("POST", "/api/admin/init-password"+tc.queryParams, headers, nil, nil, nil)
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedErrSub)
				return
			case "none":
				auth = nil
			}

			err := tctx.Client.SendHttpRequest("POST", "/api/admin/init-password"+tc.queryParams, nil, nil, nil, auth)
			assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedErrSub)
		})
	}
}

// TestAPI_AdminInitPassword_HTTPMethods verifies method rejection for /api/admin/init-password.
func TestAPI_AdminInitPassword_HTTPMethods(t *testing.T) {
	unsupportedMethods := []string{"GET", "PUT", "DELETE"}

	for _, method := range unsupportedMethods {
		method := method
		testName := fmt.Sprintf("Reject Unsupported Method %s on /api/admin/init-password", method)
		desc := fmt.Sprintf("Ensures calling %s on POST-only endpoint /api/admin/init-password returns 404/405.", method)
		expected := "HTTP 404 Not Found or 405 Method Not Allowed"

		tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
			err := tc.Client.SendHttpRequest(method, "/api/admin/init-password", nil, nil, nil, nil)
			if err == nil {
				tc.FailureReason = fmt.Sprintf("Expected rejection for %s /api/admin/init-password, got 200 OK", method)
				tc.Fatalf("Expected rejection for %s /api/admin/init-password, got 200 OK", method)
			}
			statusCode := err.StatusCode()
			if statusCode != http.StatusNotFound && statusCode != http.StatusMethodNotAllowed && statusCode != http.StatusUnauthorized {
				tc.FailureReason = fmt.Sprintf("Expected 404, 405, or 401, got HTTP %d", statusCode)
				tc.Errorf("Expected 404, 405, or 401, got HTTP %d", statusCode)
			}
			tc.Actual = fmt.Sprintf("Rejected cleanly with HTTP status %d", statusCode)
		})
	}
}

// Helper function to assert HTTP errors cleanly
func assertErrorStatus(tctx *tests.TestContext, err client.HttpError, expectedStatus int, expectedErrSub string) {
	if err == nil {
		tctx.FailureReason = fmt.Sprintf("Expected HTTP %d, but request succeeded with 200 OK", expectedStatus)
		tctx.Fatalf("Expected HTTP %d, but request succeeded with 200 OK", expectedStatus)
	}

	if err.StatusCode() != expectedStatus {
		tctx.FailureReason = fmt.Sprintf("Expected HTTP %d, got HTTP %d", expectedStatus, err.StatusCode())
		tctx.Errorf("Expected HTTP %d, got %d", expectedStatus, err.StatusCode())
	}

	if expectedErrSub != "" && !strings.Contains(err.ResponseBody(), expectedErrSub) {
		tctx.FailureReason = fmt.Sprintf("Expected error body containing '%s', got: %s", expectedErrSub, err.ResponseBody())
		tctx.Errorf("Expected error body containing '%s', got: %s", expectedErrSub, err.ResponseBody())
	}

	tctx.Actual = fmt.Sprintf("HTTP %d correctly returned with response: %s", err.StatusCode(), err.ResponseBody())
}
