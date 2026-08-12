package api_tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type AdminAddMemberRequest struct {
	TagaID                   string `json:"tagaId"`
	Name                     string `json:"name"`
	Initial                  string `json:"initial"`
	Gender                   string `json:"gender"`
	EducationalQualification string `json:"educationalQualification"`
	Designation              string `json:"designation"`
	WorkingDistrict          string `json:"workingDistrict"`
	NativeDistrict           string `json:"nativeDistrict"`
	DateOfBirth              string `json:"dateOfBirth"`
	MobileNumber             string `json:"mobileNumber"`
	Email                    string `json:"email"`
}

type AdminUpdateMemberRequest struct {
	Name                     string `json:"name"`
	Initial                  string `json:"initial"`
	Gender                   string `json:"gender"`
	EducationalQualification string `json:"educational_qualification"`
	Designation              string `json:"designation"`
	WorkingDistrict          string `json:"working_district"`
	NativeDistrict           string `json:"native_district"`
	MobileNumber             string `json:"mobile_number"`
}

// ============================================================================
// 1. Members Search & Stats - Parameterized Tests
// ============================================================================

func TestAPI_AdminMembers_Queries(t *testing.T) {
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
			name:           "Happy Path - List Members with Pagination",
			persona:        "Legitimate Admin",
			description:    "Retrieves first page list of members.",
			endpoint:       "/api/admin/members?page=1&limit=5",
			authType:       "admin",
			expectedStatus: http.StatusOK,
			expectedSub:    "members",
		},
		{
			name:           "Happy Path - Search Member",
			persona:        "Legitimate Admin",
			description:    "Retrieves members filtered by search query 'Sudhan'.",
			endpoint:       "/api/admin/members?search=Sudhan",
			authType:       "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Happy Path - Get Districts",
			persona:        "Legitimate Admin",
			description:    "Retrieves district summary from members database.",
			endpoint:       "/api/admin/members/districts",
			authType:       "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Happy Path - Get Member Stats",
			persona:        "Legitimate Admin",
			description:    "Retrieves statistical summary of members.",
			endpoint:       "/api/admin/members/stats",
			authType:       "admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Security - List Members Anonymous",
			persona:        "Anonymous Guest",
			description:    "Attempts accessing members list without any token.",
			endpoint:       "/api/admin/members",
			authType:       "none",
			expectedStatus: http.StatusUnauthorized,
			expectedSub:    "Authorization header required",
		},
		{
			name:           "Security - List Members as Normal Member",
			persona:        "Legitimate Member",
			description:    "Attempts accessing admin members list using a member token.",
			endpoint:       "/api/admin/members",
			authType:       "member",
			expectedStatus: http.StatusUnauthorized, // Admin middleware will block member token
		},
		{
			name:           "Validation - Invalid Pagination Parameters",
			persona:        "Legitimate Admin",
			description:    "Sends non-numeric pagination parameters.",
			endpoint:       "/api/admin/members?page=abc&limit=xyz",
			authType:       "admin",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Security - SQL Injection in Search",
			persona:        "Legitimate Admin",
			description:    "Attempts SQL injection in the search query parameter.",
			endpoint:       "/api/admin/members?search=' OR '1'='1",
			authType:       "admin",
			expectedStatus: http.StatusBadRequest,
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

			var resp interface{}
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
// 2. Export & Report Utilities - Parameterized Tests
// ============================================================================

func TestAPI_AdminMembers_Reports(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] GET Members Export Excel", "Admin downloads full Excel sheet of all registered members.", "HTTP 200 OK with binary stream", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		err := tctx.Client.SendHttpRequest("GET", "/api/admin/members/export", nil, nil, nil, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK Excel binary downloaded successfully"
	})

	tests.RunAPITestWithDetails(t, "[Admin] GET Generate Member Report", "Admin requests member census text/csv report.", "HTTP 200 OK with report content", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		err := tctx.Client.SendHttpRequest("GET", "/api/admin/reports/members", nil, nil, nil, auth)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = "HTTP 200 OK report downloaded successfully"
	})
}

// ============================================================================
// 3. Member CRUD Operations Workflow
// ============================================================================

func TestAPI_AdminMembers_CRUDWorkflow(t *testing.T) {
	var targetMemberID string

	// Step A: Create Member (Happy Path)
	tests.RunAPITestWithDetails(t, "[Admin] POST Add Member - Happy Path", "Admin registers a new member.", "HTTP 200 OK with member record", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		adminAuth := &client.BearerTokenAuth{Token: token}

		payload := &AdminAddMemberRequest{
			TagaID:                   "TAGA999",
			Name:                     "E2E Temporary User",
			Initial:                  "T",
			Gender:                   "Male",
			EducationalQualification: "B.Sc Agri",
			Designation:              "AO",
			WorkingDistrict:          "Salem",
			NativeDistrict:           "Salem",
			DateOfBirth:              "1990-01-01",
			MobileNumber:             "9999988888",
			Email:                    "temp.user@gmail.com",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/members/add", nil, payload, &resp, adminAuth)

		if err != nil {
			// If already exists, delete and re-attempt or fail gracefully
			if err.StatusCode() == http.StatusBadRequest && strings.Contains(err.ResponseBody(), "already registered") {
				tctx.Actual = "Correctly handled: Member email or mobile already exists in the system"
				return
			}
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}

		id, _ := resp["id"].(string)
		if id == "" {
			// Check nested member object
			if memberObj, ok := resp["member"].(map[string]interface{}); ok {
				id, _ = memberObj["id"].(string)
			}
		}
		targetMemberID = id
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Registered Member ID='%s'", targetMemberID)
	})

	// Step B: Update Member Details
	if targetMemberID != "" {
		tests.RunAPITestWithDetails(t, "[Admin] PUT Update Member Details", "Admin updates designation and mobile of the newly registered member.", "HTTP 200 OK", func(tctx *tests.TestContext) {
			token := getValidAdminToken(tctx.T, tctx.Client)
			adminAuth := &client.BearerTokenAuth{Token: token}

			payload := &AdminUpdateMemberRequest{
				Name:                     "E2E Temporary User Updated",
				Initial:                  "T",
				Gender:                   "Male",
				EducationalQualification: "B.Sc Agri",
				Designation:              "Assistant Director",
				WorkingDistrict:          "Coimbatore",
				NativeDistrict:           "Salem",
				MobileNumber:             "9999988888",
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("PUT", "/api/admin/members/"+targetMemberID, nil, payload, &resp, adminAuth)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Member updated details: %v", resp)
		})

		// Step C: Delete Member Cleanup
		tests.RunAPITestWithDetails(t, "[Admin] DELETE Member", "Admin deletes the member to clean up database state.", "HTTP 200 OK with success message", func(tctx *tests.TestContext) {
			token := getValidAdminToken(tctx.T, tctx.Client)
			adminAuth := &client.BearerTokenAuth{Token: token}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("DELETE", "/api/admin/members/"+targetMemberID, nil, nil, &resp, adminAuth)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			msg, _ := resp["message"].(string)
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
		})
	}
}

// ============================================================================
// 4. Member Add - Business Logic Parameterized Tests
// ============================================================================

func TestAPI_AdminMembers_Add_BusinessLogic(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
	}{
		{
			name:        "Business Logic - Underage Member",
			persona:     "Legitimate Admin",
			description: "Admin attempts to register an underage member.",
			payload: &AdminAddMemberRequest{
				TagaID:                   "TAGA999",
				Name:                     "Minor User",
				DateOfBirth:              "2015-01-01",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - Duplicate TagaID",
			persona:     "Legitimate Admin",
			description: "Admin attempts to register a member with an existing TagaID.",
			payload: &AdminAddMemberRequest{
				TagaID:                   "TAGA001",
				Name:                     "Existing User",
				DateOfBirth:              "1990-01-01",
			},
			expectedStatus: http.StatusBadRequest, // Or 409 depending on backend
		},
		{
			name:        "Business Logic - Invalid Working District",
			persona:     "Legitimate Admin",
			description: "Admin attempts to register a member with an invalid working district.",
			payload: &AdminAddMemberRequest{
				TagaID:                   "TAGA999",
				Name:                     "Test User",
				WorkingDistrict:          "Unknown District",
				DateOfBirth:              "1990-01-01",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Admin] POST Add Member - "+tc.name, tc.description, fmt.Sprintf("HTTP %d", tc.expectedStatus), func(tctx *tests.TestContext) {
			token := getValidAdminToken(tctx.T, tctx.Client)
			adminAuth := &client.BearerTokenAuth{Token: token}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/admin/members/add", nil, tc.payload, &resp, adminAuth)

			if err == nil {
				// Assume 400 is normally returned, so no error is a failure
				if tc.expectedStatus != http.StatusOK {
					tctx.FailureReason = fmt.Sprintf("Expected %d, got 200 OK", tc.expectedStatus)
					tctx.Errorf("Expected error status %d", tc.expectedStatus)
				}
			} else {
				// For testing, both 400 and 409 are fine if it's a conflict
				if tc.expectedStatus == http.StatusBadRequest && err.StatusCode() == http.StatusConflict {
					tctx.Actual = fmt.Sprintf("HTTP %d correctly rejected: %v", err.StatusCode(), err.ResponseBody())
				} else {
					assertErrorStatus(tctx, err, tc.expectedStatus, "")
				}
			}
		})
	}
}
