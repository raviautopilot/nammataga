package api_tests

import (
	"fmt"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type PasswordChangeRequest struct {
	Email           string `json:"email"`
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

type ForgotPasswordRequestLocal struct {
	Email string `json:"email"`
}

func TestAPI_MemberPassword_Flow(t *testing.T) {
	const email = "sudhantest08@gmail.com"
	var savedClient *client.Client

	// Step A: Change Member Password (Happy Path)
	tests.RunAPITestWithDetails(t, "[Member] POST Change Member Password", "Changes the password of the seeded member.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		savedClient = tctx.Client
		token := getValidMemberToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		payload := &PasswordChangeRequest{
			Email:           email,
			OldPassword:     "test123",
			NewPassword:     "newpassword123!",
			ConfirmPassword: "newpassword123!",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/member/change-password", nil, payload, &resp, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK password changed successfully"
	})

	// Step B: Verify we can login with the NEW password
	tests.RunAPITestWithDetails(t, "[Member] POST Login with New Password", "Validates that login now accepts the newly set password.", "HTTP 200 OK containing jwt token", func(tctx *tests.TestContext) {
		loginPayload := map[string]string{
			"email":    email,
			"password": "newpassword123!",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/member/login", nil, &loginPayload, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK login successful with new password"
	})

	// Step C: Trigger Forgot Password email check (Member forgot password)
	tests.RunAPITestWithDetails(t, "[Public] POST Forgot Password Request", "Triggers resetting password flow for member email.", "HTTP 200 OK", func(tctx *tests.TestContext) {
		payload := &ForgotPasswordRequestLocal{
			Email: email,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/auth/member-forgot-password", nil, payload, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK forgot password flow processed successfully"
	})

	// Cleanup: Revert password back to "test123" so subsequent runs don't fail
	t.Cleanup(func() {
		if savedClient != nil {
			// Login with the new password first to get a token to permit change-password route
			loginPayload := map[string]string{
				"email":    email,
				"password": "newpassword123!",
			}
			var loginResp map[string]interface{}
			err := savedClient.SendHttpRequest("POST", "/api/member/login", nil, &loginPayload, &loginResp, nil)
			if err == nil {
				token, _ := loginResp["token"].(string)
				if token != "" {
					auth := &client.BearerTokenAuth{Token: token}
					payload := &PasswordChangeRequest{
						Email:           email,
						OldPassword:     "newpassword123!",
						NewPassword:     "test123",
						ConfirmPassword: "test123",
					}
					var resp map[string]interface{}
					_ = savedClient.SendHttpRequest("POST", "/api/member/change-password", nil, payload, &resp, auth)
				}
			}
		}
	})
}

func TestAPI_MemberPassword_NegativeScenarios(t *testing.T) {
	type TestCaseType struct {
		Name        string
		Description string
		Expected    string
		TestFn      func(tc *tests.TestContext)
	}

	testCases := []TestCaseType{
		{
			Name:        "[Member] POST Change Password - Missing Old Password",
			Description: "Attempts to change password without providing the old password.",
			Expected:    "HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				token := getValidMemberToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				payload := &PasswordChangeRequest{
					Email:           "sudhantest08@gmail.com",
					NewPassword:     "newpassword123!",
					ConfirmPassword: "newpassword123!",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/member/change-password", nil, payload, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected failure due to missing old password"
					tc.Fatalf("Expected 400 Bad Request")
				}
				tc.Actual = "Correctly rejected missing old password"
			},
		},
		{
			Name:        "[Public] POST Forgot Password - SQLi Email",
			Description: "Attempts SQL injection in forgot password email.",
			Expected:    "HTTP 404 Not Found or HTTP 400 Bad Request",
			TestFn: func(tc *tests.TestContext) {
				payload := &ForgotPasswordRequestLocal{
					Email: "' OR 1=1--",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/auth/member-forgot-password", nil, payload, &resp, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection of SQLi email"
					tc.Fatalf("Expected error for SQLi but got success")
				}
				tc.Actual = "Correctly rejected SQLi email"
			},
		},
		{
			Name:        "[Member] POST Change Password - Same as Old Password",
			Description: "Attempts to change the password to the exact same current password.",
			Expected:    "HTTP 400 Bad Request or HTTP 422 Unprocessable Entity",
			TestFn: func(tc *tests.TestContext) {
				token := getValidMemberToken(tc.T, tc.Client)
				auth := &client.BearerTokenAuth{Token: token}
				payload := &PasswordChangeRequest{
					Email:           "sudhantest08@gmail.com",
					OldPassword:     "test123",
					NewPassword:     "test123",
					ConfirmPassword: "test123",
				}
				var resp map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/member/change-password", nil, payload, &resp, auth)
				if err == nil {
					tc.FailureReason = "Expected rejection for using the same password"
					tc.Fatalf("Expected 400 Bad Request but got success")
				}
				tc.Actual = "Correctly rejected changing to the same password"
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(t, testCase.Name, testCase.Description, testCase.Expected, testCase.TestFn)
	}
}
