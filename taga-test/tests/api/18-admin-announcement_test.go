package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type AdminAnnouncementRequest struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority string `json:"priority"`
	SendTo   string `json:"sendTo"`
	District string `json:"district,omitempty"`
}

// ============================================================================
// 1. POST /api/admin/announcements/send - Parameterized Tests
// ============================================================================

func TestAPI_AdminAnnouncement_Send(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		authType       string // "none", "member", "admin"
		payload        interface{}
		expectedStatus int
		expectedSub    string
	}{
		{
			name:        "Happy Path - Dispatch General Announcement",
			persona:     "Legitimate Admin",
			description: "Dispatches a high-priority announcement to all members.",
			authType:    "admin",
			payload: &AdminAnnouncementRequest{
				Title:    "Annual General Body Meeting 2026",
				Message:  "The AGM will be held on December 15th at TAGA Tower Salem. Attendance is mandatory.",
				Priority: "High",
				SendTo:   "all",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "announcement_id",
		},
		{
			name:        "Happy Path - Dispatch District Specific Announcement",
			persona:     "Legitimate Admin",
			description: "Dispatches an announcement specifically targeting members of Salem district.",
			authType:    "admin",
			payload: &AdminAnnouncementRequest{
				Title:    "Salem Members Meeting",
				Message:  "District representatives will meet this Sunday.",
				Priority: "Normal",
				SendTo:   "district",
				District: "Salem",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "announcement_id",
		},
		{
			name:        "Validation - Missing Title or Message",
			persona:     "Legitimate Admin",
			description: "Attempts dispatching an announcement without specifying the mandatory title or message field.",
			authType:    "admin",
			payload: &AdminAnnouncementRequest{
				Priority: "Low",
				SendTo:   "all",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - Dispatch Announcement Anonymous",
			persona:        "Anonymous Guest",
			description:    "Attempts posting a message dispatch without supplying an Authorization token.",
			authType:       "none",
			payload:        &AdminAnnouncementRequest{Title: "Hack", Message: "Breach", SendTo: "all"},
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
		{
			name:           "Security - Dispatch Announcement as Member",
			persona:        "Legitimate Member",
			description:    "Attempts posting a message dispatch using a standard Member Bearer token.",
			authType:       "member",
			payload:        &AdminAnnouncementRequest{Title: "Unfair Post", Message: "Intrusion", SendTo: "all"},
			expectedStatus: http.StatusUnauthorized, // Admin middleware rejects member token
		},
		{
			name:           "Validation - Missing Target Audience (SendTo)",
			persona:        "Legitimate Admin",
			description:    "Attempts to send announcement without SendTo field.",
			authType:       "admin",
			payload:        &AdminAnnouncementRequest{Title: "Test", Message: "Test message"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - HTML Injection in Message",
			persona:        "Legitimate Admin",
			description:    "Attempts to inject malicious HTML in announcement message.",
			authType:       "admin",
			payload:        &AdminAnnouncementRequest{Title: "Update", Message: "<div onclick='alert(1)'>Click me</div>", SendTo: "all"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - District Target With Missing District",
			persona:     "Legitimate Admin",
			description: "Sends announcement to 'district' but district field is empty.",
			authType:    "admin",
			payload: &AdminAnnouncementRequest{
				Title:    "Salem Members Meeting",
				Message:  "District representatives will meet this Sunday.",
				Priority: "Normal",
				SendTo:   "district",
				District: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - Invalid Priority Level",
			persona:     "Legitimate Admin",
			description: "Attempts sending an announcement with an invalid priority level.",
			authType:    "admin",
			payload: &AdminAnnouncementRequest{
				Title:    "General Announcement",
				Message:  "This is a general announcement.",
				Priority: "SuperUrgent",
				SendTo:   "all",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] POST Send Announcement - %s", tc.persona, tc.name)
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
			err := tctx.Client.SendHttpRequest("POST", "/api/admin/announcements/send", nil, tc.payload, &resp, auth)

			if tc.expectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				id, _ := resp["announcement_id"].(string)
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Announcement ID='%s'", id)
			} else {
				assertErrorStatus(tctx, err, tc.expectedStatus, tc.expectedSub)
			}
		})
	}
}
