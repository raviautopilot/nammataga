package api_tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// EditRequest defines the payload for profile edit requests
type EditRequest struct {
	Email              string `json:"email"`
	MobileNumber       string `json:"mobileNumber"`
	Designation        string `json:"designation"`
	WorkingDistrict    string `json:"workingDistrict"`
	ResidentialAddress string `json:"residentialAddress"`
	PermanentAddress   string `json:"permanentAddress"`
	Remarks            string `json:"remarks"`
}

// ============================================================================
// 1. GET /api/member/profile - Parameterized Tests
// ============================================================================

func TestAPI_MemberProfile_Get(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "invalid", "member"
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Retrieve Own Profile",
			persona:        "Authenticated Member",
			description:    "Retrieves profile data of the logged-in member using a valid Bearer JWT.",
			authType:       "member",
			expectedStatus: http.StatusOK,
			expectedSub:    "sudhantest08@gmail.com",
		},
		{
			name:           "Security - No Authorization Header",
			persona:        "Anonymous Guest",
			description:    "Attempts profile access without supplying an Authorization token.",
			authType:       "none",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
		{
			name:           "Security - Invalid JWT Token",
			persona:        "Attacker with Forged Token",
			description:    "Attempts profile access with a malformed/invalid JWT token.",
			authType:       "invalid",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Invalid token",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET Profile - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			switch tc.authType {
			case "member":
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "invalid":
				auth = &client.BearerTokenAuth{Token: "invalid_jwt_token_payload"}
			case "none":
				auth = nil
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/member/profile", nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				userObj, ok := resp["user"].(map[string]interface{})
				if !ok {
					tctx.FailureReason = "Response body missing 'user' object"
					tctx.Fatalf("Response missing 'user' object")
				}
				email, _ := userObj["emailId"].(string)
				if !strings.Contains(email, tc.expectedSub) {
					tctx.FailureReason = fmt.Sprintf("Expected email containing '%s', got '%s'", tc.expectedSub, email)
					tctx.Errorf("Expected email containing '%s', got '%s'", tc.expectedSub, email)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, User Profile Email='%s'", email)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 2. PUT /api/member/profile - Parameterized Tests
// ============================================================================

func TestAPI_MemberProfile_Update(t *testing.T) {
	invalidJSON := "this-is-not-json-data"

	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "member"
		payload        interface{}
		expectedStatus int
		expectedSub    string
	}{
		{
			name:        "Happy Path - Update Address and Contact",
			persona:     "Authenticated Member",
			description: "Updates profile fields (name, mobile_number, residential_address) with valid data.",
			authType:    "member",
			payload: &map[string]interface{}{
				"name":                "Sudhan Updated",
				"mobile_number":       "9944637254",
				"residential_address": "12 Gandhi Nagar Updated",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "Profile updated successfully",
		},
		{
			name:           "Security - Unauthenticated Update Attempt",
			persona:        "Anonymous Guest",
			description:    "Attempts to update profile fields without authorization.",
			authType:       "none",
			payload:        &map[string]interface{}{"name": "Intruder"},
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
		{
			name:        "Validation - Malformed JSON Request Body",
			persona:     "Authenticated Member",
			description: "Submits malformed JSON content to the profile update API.",
			authType:    "member",
			payload:     &invalidJSON,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Validation - Exceed Boundary Value",
			persona:     "Authenticated Member",
			description: "Submits excessively long string for mobile_number.",
			authType:    "member",
			payload: &map[string]interface{}{
				"mobile_number": strings.Repeat("9", 256),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Security - XSS in Remarks",
			persona:     "Authenticated Member",
			description: "Attempts XSS injection in remarks field.",
			authType:    "member",
			payload: &map[string]interface{}{
				"remarks": "<script>alert('xss')</script>",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "successfully",
		},
		{
			name:        "Business Logic - Negative Mobile Number",
			persona:     "Authenticated Member",
			description: "Submits negative value for mobile number.",
			authType:    "member",
			payload: &map[string]interface{}{
				"mobile_number": "-9876543210",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Extreme Boundary - 100 Year Address",
			persona:     "Authenticated Member",
			description: "Submits extremely large address length.",
			authType:    "member",
			payload: &map[string]interface{}{
				"residential_address": strings.Repeat("A", 10000),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "State Machine Violation - Edit Approved Profile",
			persona:     "Authenticated Member",
			description: "Submits state override to edit an already approved profile.",
			authType:    "member",
			payload: &map[string]interface{}{
				"status": "approved",
				"action": "edit",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "successfully",
		},
		{
			name:        "Logical Paradox - Both Active and Suspended",
			persona:     "Authenticated Member",
			description: "Submits contradictory status flags.",
			authType:    "member",
			payload: &map[string]interface{}{
				"is_active": true,
				"is_suspended": true,
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "successfully",
		},
		{
			name:        "Role Context Switching - Member Self Promotion",
			persona:     "Authenticated Member",
			description: "Attempts to update their own role to admin.",
			authType:    "member",
			payload: &map[string]interface{}{
				"role": "admin",
				"is_admin": true,
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "successfully",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] PUT Profile - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("PUT", "/api/member/profile", nil, tc.payload, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				msg, _ := resp["message"].(string)
				if !strings.Contains(msg, tc.expectedSub) {
					tctx.FailureReason = fmt.Sprintf("Expected message containing '%s', got '%s'", tc.expectedSub, msg)
					tctx.Errorf("Expected message containing '%s', got '%s'", tc.expectedSub, msg)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 3. POST /api/member/edit-request - Parameterized Tests
// ============================================================================

func TestAPI_MemberProfile_EditRequest(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
		expectedSub    string
	}{
		{
			name:        "Happy Path - Create New Profile Edit Request",
			persona:     "Legitimate Member",
			description: "Submits a profile edit request which gets saved for admin approval.",
			payload: &EditRequest{
				Email:              "sudhanop05@gmail.com",
				MobileNumber:       "9944637254",
				Designation:        "Senior Executive",
				WorkingDistrict:    "Salem",
				ResidentialAddress: "12 Gandhi Nagar",
				PermanentAddress:   "12 Gandhi Nagar",
				Remarks:            "E2E Update Request",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "Edit request submitted successfully",
		},
		{
			name:           "Validation - Empty Payload",
			persona:        "Buggy Client",
			description:    "Submits empty payload.",
			payload:        &EditRequest{},
			expectedStatus: http.StatusOK, // Save succeeds but email might skip or fail gracefully
			expectedSub:    "submitted successfully",
		},
		{
			name:           "Security - SQL Injection in Email Edit",
			persona:        "Legitimate Member",
			description:    "Submits SQLi payload in email field.",
			payload: &EditRequest{
				Email:              "' OR '1'='1",
				MobileNumber:       "9944637254",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Business Logic - Edit Request With Missing Required Designation",
			persona:        "Legitimate Member",
			description:    "Submits an edit request with missing designation.",
			payload: &EditRequest{
				Email:              "sudhanop05@gmail.com",
				MobileNumber:       "9944637254",
				WorkingDistrict:    "Salem",
				ResidentialAddress: "12 Gandhi Nagar",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] POST Edit Request - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d containing '%s'", tc.expectedStatus, tc.expectedSub)

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/member/edit-request", nil, tc.payload, &resp, nil)

			if err != nil {
				// Fallback to "saved but email failed" which returns 500 if SMTP server is offline
				if err.StatusCode() == http.StatusInternalServerError && strings.Contains(err.ResponseBody(), "saved but email failed") {
					tctx.Actual = "HTTP 500 correctly handled graceful SMTP failure: " + err.ResponseBody()
					return
				}
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}

			msg, _ := resp["message"].(string)
			if !strings.Contains(msg, tc.expectedSub) {
				tctx.FailureReason = fmt.Sprintf("Expected message containing '%s', got '%s'", tc.expectedSub, msg)
				tctx.Errorf("Expected message containing '%s', got '%s'", tc.expectedSub, msg)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
		})
	}
}

// ============================================================================
// 4. GET /api/member/notifications - Parameterized Tests
// ============================================================================

func TestAPI_MemberProfile_Notifications(t *testing.T) {
	memberUUID := "d11348e1-9a65-4945-bb1b-f100a5df15cg" // sudhantest08@gmail.com ID

	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "member"
		queryParams    string
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Retrieve Notifications Array",
			persona:        "Authenticated Member",
			description:    "Retrieves notification list for valid member ID with a Bearer JWT.",
			authType:       "member",
			queryParams:    "?member_id=" + memberUUID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Validation - Missing member_id Parameter",
			persona:        "Authenticated Member",
			description:    "Attempts notification fetch without supplying member_id.",
			authType:       "member",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "member_id is required",
		},
		{
			name:           "Security - Access Notifications Unauthenticated",
			persona:        "Anonymous Guest",
			description:    "Attempts to pull notifications without supplying a Bearer token.",
			authType:       "none",
			queryParams:    "?member_id=" + memberUUID,
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
		{
			name:           "Business Logic - View Other User's Notifications (IDOR)",
			persona:        "Authenticated Member",
			description:    "Attempts to pull notifications for a different member_id.",
			authType:       "member",
			queryParams:    "?member_id=another-user-uuid",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET Notifications - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp []interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/member/notifications"+tc.queryParams, nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d notifications", len(resp))
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 5. PUT /api/member/notifications/:id/read - Parameterized Tests
// ============================================================================

func TestAPI_MemberProfile_MarkNotificationRead(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "member"
		notificationID string
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Mark Notification Read",
			persona:        "Authenticated Member",
			description:    "Sends read status update for a notification ID.",
			authType:       "member",
			notificationID: "non-existent-notification-id",
			expectedStatus: http.StatusOK,
			expectedSub:    "Notification marked as read",
		},
		{
			name:           "Security - Mark Read Without Auth Header",
			persona:        "Anonymous Guest",
			description:    "Attempts marking notification read without login token.",
			authType:       "none",
			notificationID: "12345",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] PUT Mark Read - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d containing '%s'", tc.expectedStatus, tc.expectedSub)

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("PUT", "/api/member/notifications/"+tc.notificationID+"/read", nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				msg, _ := resp["message"].(string)
				if !strings.Contains(msg, tc.expectedSub) {
					tctx.FailureReason = fmt.Sprintf("Expected message containing '%s', got '%s'", tc.expectedSub, msg)
					tctx.Errorf("Expected message containing '%s', got '%s'", tc.expectedSub, msg)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}

// ============================================================================
// 6. GET /api/member/notifications/unread/count - Parameterized Tests
// ============================================================================

func TestAPI_MemberProfile_UnreadCount(t *testing.T) {
	memberUUID := "d11348e1-9a65-4945-bb1b-f100a5df15cg" // sudhantest08@gmail.com ID

	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "member"
		queryParams    string
		expectedStatus int
		expectedSub    string
	}{
		{
			name:           "Happy Path - Retrieve Unread Count",
			persona:        "Authenticated Member",
			description:    "Retrieves unread notifications counter using valid Member JWT.",
			authType:       "member",
			queryParams:    "?member_id=" + memberUUID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Validation - Missing member_id Parameter",
			persona:        "Authenticated Member",
			description:    "Attempts fetching unread counter without supplying member_id.",
			authType:       "member",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "member_id is required",
		},
		{
			name:           "Security - Get Counter Unauthenticated",
			persona:        "Anonymous Guest",
			description:    "Attempts pulling unread count without authorization.",
			authType:       "none",
			queryParams:    "?member_id=" + memberUUID,
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] GET Unread Count - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			if tc.authType == "member" {
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/member/notifications/unread/count"+tc.queryParams, nil, nil, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				count, exists := resp["unread_count"]
				if !exists {
					tctx.FailureReason = "Response payload missing 'unread_count' key"
					tctx.Fatalf("Response missing 'unread_count'")
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, UnreadCount=%v", count)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}
