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

// MemberLoginRequest defines the payload for POST /api/member/login
type MemberLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// MemberLoginResponse defines the structure returned by POST /api/member/login
type MemberLoginResponse struct {
	Token               string                 `json:"token"`
	Role                string                 `json:"role"`
	ExpiresIn           int64                  `json:"expires_in"`
	User                map[string]interface{} `json:"user"`
	ForceChangePassword bool                   `json:"forceChangePassword"`
	Message             string                 `json:"message,omitempty"`
}

// MemberChangePasswordRequest defines payload for POST /api/member/change-password
type MemberChangePasswordRequest struct {
	Email           string `json:"email"`
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

// ForgotPasswordRequest defines payload for POST /api/auth/forgot-password
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest defines payload for POST /api/auth/reset-password
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// Helper function to obtain a valid member token for authenticated tests
func getValidMemberToken(t *testing.T, c *client.Client) string {
	t.Helper()
	memberCreds := tests.GlobalConfig.MemberCredentials
	candidates := []MemberLoginRequest{
		{Email: memberCreds.Username, Password: memberCreds.Password},
		{Email: "sudhanop05@gmail.com", Password: "test123"},
		{Email: "sudhanop05@gmail.com", Password: "admin123"},
	}

	for _, cand := range candidates {
		var resp MemberLoginResponse
		err := c.SendHttpRequest("POST", "/api/member/login", nil, &cand, &resp, nil)
		if err == nil && resp.Token != "" {
			return resp.Token
		}
	}
	return ""
}

// ============================================================================
// 1. POST /api/member/login - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_MemberLogin_TableDriven(t *testing.T) {
	memberCreds := tests.GlobalConfig.MemberCredentials

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
			name:           "Happy Path - Valid Member Credentials",
			persona:        "Legitimate Member",
			description:    "Submits correct registered member credentials to obtain member JWT.",
			payload:        &MemberLoginRequest{Email: memberCreds.Username, Password: memberCreds.Password},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "Auth Failure - Invalid Password",
			persona:        "Intruder with Wrong Password",
			description:    "Submits valid member email with an incorrect password.",
			payload:        &MemberLoginRequest{Email: memberCreds.Username, Password: "WrongPassword999!"},
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid credentials",
		},
		{
			name:           "Auth Failure - Nonexistent Email",
			persona:        "Intruder with Unknown Email",
			description:    "Submits unregistered email address with arbitrary password.",
			payload:        &MemberLoginRequest{Email: "nonexistent_member_9999@taga.org", Password: "anyPassword"},
			expectedStatus: http.StatusUnauthorized,
			expectedErrSub: "Invalid credentials",
		},
		{
			name:           "Validation - Empty Payload",
			persona:        "Buggy Client",
			description:    "Submits empty JSON body violating required field bindings.",
			payload:        &MemberLoginRequest{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation - Missing Password Field",
			persona:        "Malformed Client",
			description:    "Submits email only, omitting password.",
			payload:        &map[string]string{"email": memberCreds.Username},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation - Missing Email Field",
			persona:        "Malformed Client",
			description:    "Submits password only, omitting email.",
			payload:        &map[string]string{"password": memberCreds.Password},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - SQL Injection in Email",
			persona:        "Malicious Security Auditor",
			description:    "Attempts SQL injection payload in member email field.",
			payload:        &MemberLoginRequest{Email: "' OR '1'='1' --", Password: "any"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Security - XSS Script Tag in Email",
			persona:        "Malicious Security Auditor",
			description:    "Attempts XSS script payload in member email field.",
			payload:        &MemberLoginRequest{Email: "<script>alert('xss')</script>", Password: "any"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Boundary - Extremely Long Password",
			persona:        "Attacker with Long Password",
			description:    "Submits extremely long password string.",
			payload:        &MemberLoginRequest{Email: memberCreds.Username, Password: strings.Repeat("A", 1024)},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Security - XSS in Password",
			persona:        "Malicious Security Auditor",
			description:    "Attempts XSS payload in password field.",
			payload:        &MemberLoginRequest{Email: memberCreds.Username, Password: "<script>alert('xss')</script>"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Business Logic - Member Login Admin Role Request",
			persona:        "Malicious Member",
			description:    "Submits login request attempting to force admin role.",
			payload:        map[string]interface{}{"email": memberCreds.Username, "password": memberCreds.Password, "role": "admin"},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "Extreme Boundary - Expiry 100 Years",
			persona:        "Attacker Requesting Long JWT",
			description:    "Submits extremely long requested expiry time.",
			payload:        &map[string]interface{}{"email": memberCreds.Username, "password": memberCreds.Password, "expires_in": 3153600000},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "State Machine Violation - Logout Already Logged Out Member",
			persona:        "Member",
			description:    "Attempts to login with a flag to logout first.",
			payload:        &map[string]interface{}{"email": memberCreds.Username, "password": memberCreds.Password, "action": "logout"},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "Logical Paradox - Member Login With Both Force Password Change and Skip Flags",
			persona:        "Member",
			description:    "Attempts to set contradicting force_change and skip_change flags.",
			payload:        &map[string]interface{}{"email": memberCreds.Username, "password": memberCreds.Password, "force_change": true, "skip_change": true},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectToken {
			expectedStr += " with valid JWT token and role 'member'"
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp MemberLoginResponse
			var errResp CommonAPIErrorResponse

			var rawPtr interface{}
			if tc.expectedStatus == http.StatusOK {
				rawPtr = &resp
			} else {
				rawPtr = &errResp
			}

			err := tctx.Client.SendHttpRequest("POST", "/api/member/login", nil, tc.payload, rawPtr, nil)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected HTTP 200, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got error: %v", err)
				}
				if resp.Token == "" {
					tctx.FailureReason = "Expected non-empty JWT token"
					tctx.Errorf("Member login response missing token")
				}
				if resp.Role != "member" {
					tctx.FailureReason = fmt.Sprintf("Expected role 'member', got '%s'", resp.Role)
					tctx.Errorf("Expected role 'member', got '%s'", resp.Role)
				}
				if resp.ExpiresIn <= 0 {
					tctx.FailureReason = fmt.Sprintf("Expected positive expiresIn, got %d", resp.ExpiresIn)
					tctx.Errorf("Expected positive expiresIn, got %d", resp.ExpiresIn)
				}
				if resp.User == nil {
					tctx.FailureReason = "Expected non-nil user object in login response"
					tctx.Errorf("Expected non-nil user object in login response")
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Role='%s', TokenLength=%d, UserID='%v'", resp.Role, len(resp.Token), resp.User["id"])
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

// TestAPI_MemberLogin_HTTPMethods verifies method rejection for /api/member/login.
func TestAPI_MemberLogin_HTTPMethods(t *testing.T) {
	unsupportedMethods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range unsupportedMethods {
		method := method
		testName := fmt.Sprintf("Reject Unsupported Method %s on /api/member/login", method)
		desc := fmt.Sprintf("Ensures calling %s on POST-only endpoint /api/member/login returns 404/405.", method)
		expected := "HTTP 404 Not Found or 405 Method Not Allowed"

		tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
			err := tc.Client.SendHttpRequest(method, "/api/member/login", nil, nil, nil, nil)
			if err == nil {
				tc.FailureReason = fmt.Sprintf("Expected rejection for %s /api/member/login, got 200 OK", method)
				tc.Fatalf("Expected rejection for %s /api/member/login, got 200 OK", method)
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

// TestAPI_MemberLogin_LatencySLA verifies login endpoint responds within SLA (< 1500ms).
func TestAPI_MemberLogin_LatencySLA(t *testing.T) {
	tests.RunAPITestWithDetails(
		t,
		"Member Login Performance SLA (< 1500ms)",
		"Validates that member credential verification and JWT generation respond under 1500ms.",
		"HTTP 200 OK within 1500ms",
		func(tc *tests.TestContext) {
			memberCreds := tests.GlobalConfig.MemberCredentials
			req := &MemberLoginRequest{
				Email:    memberCreds.Username,
				Password: memberCreds.Password,
			}
			var resp MemberLoginResponse

			start := time.Now()
			err := tc.Client.SendHttpRequest("POST", "/api/member/login", nil, req, &resp, nil)
			latency := time.Since(start)

			if err != nil {
				tc.FailureReason = fmt.Sprintf("Member login request failed: %v", err)
				tc.Fatalf("Member login request failed: %v", err)
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
// 2. POST /api/member/logout - Parameterized Tests
// ============================================================================

func TestAPI_MemberLogout_TableDriven(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		useAuth        bool
		expectedStatus int
		expectedMsgSub string
	}{
		{
			name:           "Happy Path - Authenticated Member Logout",
			persona:        "Authenticated Member",
			description:    "Calls logout with valid member JWT token.",
			useAuth:        true,
			expectedStatus: http.StatusOK,
			expectedMsgSub: "Logged out successfully",
		},
		{
			name:           "Happy Path - Anonymous Logout",
			persona:        "Anonymous User",
			description:    "Calls logout without JWT token.",
			useAuth:        false,
			expectedStatus: http.StatusOK,
			expectedMsgSub: "Logged out successfully",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d containing '%s'", tc.expectedStatus, tc.expectedMsgSub)

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.useAuth {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/member/logout", nil, nil, &resp, auth)
			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Logout failed: %v", err)
			}

			msg, _ := resp["message"].(string)
			if !strings.Contains(msg, tc.expectedMsgSub) {
				tctx.FailureReason = fmt.Sprintf("Expected message containing '%s', got '%s'", tc.expectedMsgSub, msg)
				tctx.Errorf("Expected message containing '%s', got '%s'", tc.expectedMsgSub, msg)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
		})
	}
}

// TestAPI_MemberLogout_HTTPMethods verifies method rejection for /api/member/logout.
func TestAPI_MemberLogout_HTTPMethods(t *testing.T) {
	unsupportedMethods := []string{"GET", "PUT", "DELETE"}

	for _, method := range unsupportedMethods {
		method := method
		testName := fmt.Sprintf("Reject Unsupported Method %s on /api/member/logout", method)
		desc := fmt.Sprintf("Ensures calling %s on POST-only endpoint /api/member/logout returns 404/405.", method)
		expected := "HTTP 404 Not Found or 405 Method Not Allowed"

		tests.RunAPITestWithDetails(t, testName, desc, expected, func(tc *tests.TestContext) {
			err := tc.Client.SendHttpRequest(method, "/api/member/logout", nil, nil, nil, nil)
			if err == nil {
				tc.FailureReason = fmt.Sprintf("Expected rejection for %s /api/member/logout, got 200 OK", method)
				tc.Fatalf("Expected rejection for %s /api/member/logout, got 200 OK", method)
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

// ============================================================================
// 3. POST /api/member/change-password - Parameterized Tests
// ============================================================================

func TestAPI_MemberChangePassword_TableDriven(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
		expectedErrSub string
	}{
		{
			name:           "Validation - Empty Payload",
			persona:        "Buggy Client",
			description:    "Submits empty payload violating required fields.",
			payload:        &MemberChangePasswordRequest{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation - Passwords Do Not Match",
			persona:        "Legitimate Member",
			description:    "Submits new password differing from confirmation password.",
			payload:        &MemberChangePasswordRequest{Email: "sudhanop05@gmail.com", OldPassword: "oldPass123", NewPassword: "newPass123!", ConfirmPassword: "differentPass123!"},
			expectedStatus: http.StatusBadRequest,
			expectedErrSub: "Passwords do not match",
		},
		{
			name:           "Validation - Incorrect Old Password",
			persona:        "Member with Bad Memory",
			description:    "Submits an invalid old password for existing user.",
			payload:        &MemberChangePasswordRequest{Email: "sudhanop05@gmail.com", OldPassword: "completely_wrong_old_password", NewPassword: "newPass123!", ConfirmPassword: "newPass123!"},
			expectedStatus: http.StatusBadRequest,
			expectedErrSub: "Old password is incorrect",
		},
		{
			name:           "Status - Nonexistent Member User",
			persona:        "Unregistered User",
			description:    "Attempts password change for email that does not exist in member database.",
			payload:        &MemberChangePasswordRequest{Email: "unregistered_999@unknown.org", OldPassword: "anyOldPass", NewPassword: "newPass123!", ConfirmPassword: "newPass123!"},
			expectedStatus: http.StatusNotFound,
			expectedErrSub: "User not found",
		},
		{
			name:           "Security - SQL Injection in Old Password",
			persona:        "Legitimate Member",
			description:    "Submits SQL injection payload as old password.",
			payload:        &MemberChangePasswordRequest{Email: "sudhanop05@gmail.com", OldPassword: "' OR '1'='1", NewPassword: "newPass123!", ConfirmPassword: "newPass123!"},
			expectedStatus: http.StatusBadRequest,
			expectedErrSub: "Old password is incorrect",
		},
		{
			name:           "Boundary - Short New Password",
			persona:        "Legitimate Member",
			description:    "Submits a new password that is too short.",
			payload:        &MemberChangePasswordRequest{Email: "sudhanop05@gmail.com", OldPassword: "oldPass123", NewPassword: "123", ConfirmPassword: "123"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Business Logic - Password Same As Old Password",
			persona:        "Lazy Member",
			description:    "Submits new password that is identical to old password.",
			payload:        &MemberChangePasswordRequest{Email: "sudhanop05@gmail.com", OldPassword: "oldPass123", NewPassword: "oldPass123", ConfirmPassword: "oldPass123"},
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
			err := tctx.Client.SendHttpRequest("POST", "/api/member/change-password", nil, tc.payload, nil, nil)
			assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedErrSub)
		})
	}
}

// ============================================================================
// 4. POST /api/auth/forgot-password & member-forgot-password - Tests
// ============================================================================

func TestAPI_ForgotPassword_TableDriven(t *testing.T) {
	cases := []struct {
		name           string
		endpoint       string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Validation - Forgot Password Unknown Email",
			endpoint:       "/api/auth/forgot-password",
			persona:        "Unregistered User",
			description:    "Submits unregistered email to /api/auth/forgot-password.",
			payload:        &ForgotPasswordRequest{Email: "unknown_user_999@taga.org"},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "Email not found",
		},
		{
			name:           "Validation - Forgot Password Empty Payload",
			endpoint:       "/api/auth/forgot-password",
			persona:        "Buggy Client",
			description:    "Submits empty payload.",
			payload:        &ForgotPasswordRequest{},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "Email not found",
		},
		{
			name:           "Validation - Member Forgot Password Unknown Email",
			endpoint:       "/api/auth/member-forgot-password",
			persona:        "Unregistered User",
			description:    "Submits unregistered email to /api/auth/member-forgot-password.",
			payload:        &ForgotPasswordRequest{Email: "unknown_user_999@taga.org"},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "email not found",
		},
		{
			name:           "Validation - Member Forgot Password Empty Body",
			endpoint:       "/api/auth/member-forgot-password",
			persona:        "Buggy Client",
			description:    "Submits empty body to /api/auth/member-forgot-password.",
			payload:        &ForgotPasswordRequest{},
			expectedStatus: http.StatusNotFound,
			expectedSub:    "email not found",
		},
		{
			name:           "Security - SQL Injection in Forgot Password Email",
			endpoint:       "/api/auth/forgot-password",
			persona:        "Attacker",
			description:    "Submits SQL injection payload to forgot password.",
			payload:        &ForgotPasswordRequest{Email: "' OR '1'='1"},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Business Logic - Multiple Forgot Password Requests (Rate Limiting)",
			endpoint:       "/api/auth/forgot-password",
			persona:        "Spammer",
			description:    "Submits multiple rapid forgot password requests for same email.",
			payload:        &ForgotPasswordRequest{Email: "sudhanop05@gmail.com"},
			expectedStatus: http.StatusOK, // Assuming we want it to succeed or be 429
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s on %s", tc.persona, tc.name, tc.endpoint)
		expectedStr := fmt.Sprintf("HTTP %d containing '%s'", tc.expectedStatus, tc.expectedSub)

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp map[string]interface{}

			err := tctx.Client.SendHttpRequest("POST", tc.endpoint, nil, tc.payload, &resp, nil)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				msg := fmt.Sprintf("%v", resp)
				if !strings.Contains(msg, tc.expectedSub) {
					tctx.FailureReason = fmt.Sprintf("Expected response containing '%s', got '%s'", tc.expectedSub, msg)
					tctx.Errorf("Expected response containing '%s', got '%s'", tc.expectedSub, msg)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK with response: %s", msg)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 5. POST /api/auth/reset-password - Parameterized Tests
// ============================================================================

func TestAPI_ResetPassword_TableDriven(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
		expectedErrSub string
	}{
		{
			name:           "Auth Check - Empty Body Without Email",
			persona:        "Buggy Client",
			description:    "Submits empty JSON payload.",
			payload:        &ResetPasswordRequest{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Auth Check - Expired or Invalid Token Payload",
			persona:        "Attacker with Invalid Token",
			description:    "Attempts password reset with invalid old password/token.",
			payload:        &ResetPasswordRequest{Email: "sudhanop05@gmail.com", OldPassword: "invalid_token_or_password", NewPassword: "NewSecretPassword123!"},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			err := tctx.Client.SendHttpRequest("POST", "/api/auth/reset-password", nil, tc.payload, nil, nil)
			if err == nil {
				tctx.FailureReason = fmt.Sprintf("Expected HTTP %d error, but got 200 OK", tc.expectedStatus)
				tctx.Fatalf("Expected HTTP %d error, but got 200 OK", tc.expectedStatus)
			}
			if err.StatusCode() != tc.expectedStatus && err.StatusCode() != http.StatusUnauthorized {
				tctx.FailureReason = fmt.Sprintf("Expected HTTP %d, got %d", tc.expectedStatus, err.StatusCode())
				tctx.Errorf("Expected HTTP %d, got %d", tc.expectedStatus, err.StatusCode())
			}
			tctx.Actual = fmt.Sprintf("HTTP %d returned with response: %s", err.StatusCode(), err.ResponseBody())
		})
	}
}
