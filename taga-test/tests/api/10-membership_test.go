package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/tests"
)

// MembershipApplication represents the payload for applying for membership
type MembershipApplication struct {
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	DateOfBirth       string `json:"dateOfBirth"`
	Gender            string `json:"gender"`
	Address           string `json:"address"`
	District          string `json:"district"`
	Qualification     string `json:"qualification"`
	GraduationYear    string `json:"graduationYear"`
	CurrentEmployment string `json:"currentEmployment"`
	Organization      string `json:"organization"`
	Experience        string `json:"experience"`
	Specialization    string `json:"specialization"`
	References        string `json:"references"`
}

// ============================================================================
// 1. POST /api/membership/apply - Parameterized Tests
// ============================================================================

func TestAPI_Membership_Apply(t *testing.T) {
	cases := []struct {
		name           string
		persona        string
		description    string
		payload        interface{}
		expectedStatus int
		expectedSub    string
	}{
		{
			name:        "Happy Path - Apply for New Membership",
			persona:     "Prospective Member",
			description: "Submits a membership application form with all required and optional details.",
			payload: &MembershipApplication{
				FirstName:         "Rajesh",
				LastName:          "Kumar",
				Email:             "rajesh.kumar@gmail.com",
				Phone:             "9876543210",
				DateOfBirth:       "1992-06-25",
				Gender:            "Male",
				Address:           "45 Nehru Street, Salem",
				District:          "Salem",
				Qualification:     "B.Sc Agriculture",
				GraduationYear:    "2014",
				CurrentEmployment: "Assistant Manager",
				Organization:      "Green Agro Ltd",
				Experience:        "5 Years",
				Specialization:    "Soil Science",
				References:        "Dr. P. Swaminathan",
			},
			expectedStatus: http.StatusOK,
			expectedSub:    "Application submitted",
		},
		{
			name:        "Validation - Malformed Request Body",
			persona:     "Buggy Client",
			description: "Submits an invalid JSON value (unparsable content) to the application endpoint.",
			payload: func() *string {
				s := "unparsable-payload"
				return &s
			}(),
			expectedStatus: http.StatusBadRequest,
			expectedSub:    "Invalid request body",
		},
		{
			name:        "Validation - Missing Required Fields",
			persona:     "Careless User",
			description: "Submits an application with empty required fields.",
			payload: &MembershipApplication{
				FirstName: "",
				Email:     "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Validation - Invalid Email Format",
			persona:     "Buggy Client",
			description: "Submits an application with malformed email.",
			payload: &MembershipApplication{
				FirstName: "Rajesh",
				Email:     "not-an-email",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Security - SQLi in Payload",
			persona:     "Malicious User",
			description: "Attempts SQL injection in address field.",
			payload: &MembershipApplication{
				FirstName: "John",
				Email:     "john@test.com",
				Address:   "1'; DROP TABLE members; --",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - Underage Applicant",
			persona:     "Applicant",
			description: "Applicant is under 18 years of age.",
			payload: &MembershipApplication{
				FirstName:         "Minor",
				LastName:          "User",
				Email:             "minor@gmail.com",
				Phone:             "9876543210",
				DateOfBirth:       "2018-06-25",
				Gender:            "Male",
				Address:           "45 Nehru Street, Salem",
				District:          "Salem",
				Qualification:     "B.Sc Agriculture",
				GraduationYear:    "2038",
				CurrentEmployment: "Student",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - Future Date of Birth",
			persona:     "Applicant",
			description: "Applicant claims to be born in the future.",
			payload: &MembershipApplication{
				FirstName:         "Time",
				LastName:          "Traveler",
				Email:             "future@gmail.com",
				Phone:             "9876543210",
				DateOfBirth:       "2050-01-01",
				Gender:            "Female",
				Address:           "45 Nehru Street, Salem",
				District:          "Salem",
				Qualification:     "B.Sc Agriculture",
				GraduationYear:    "2070",
				CurrentEmployment: "Student",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Business Logic - Graduation Before Birth",
			persona:     "Applicant",
			description: "Applicant claims to have graduated before they were born.",
			payload: &MembershipApplication{
				FirstName:         "Impossible",
				LastName:          "User",
				Email:             "impossible@gmail.com",
				Phone:             "9876543210",
				DateOfBirth:       "1990-01-01",
				Gender:            "Male",
				Address:           "45 Nehru Street, Salem",
				District:          "Salem",
				Qualification:     "B.Sc Agriculture",
				GraduationYear:    "1985",
				CurrentEmployment: "Manager",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] POST Apply - %s", tc.persona, tc.name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.expectedStatus)
		if tc.expectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.expectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.description, expectedStr, func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/membership/apply", nil, tc.payload, &resp, nil)

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
// 2. GET /api/membership/list - Parameterized Tests
// ============================================================================

func TestAPI_Membership_List(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public/Admin] GET Applications List", "Retrieves the list of all submitted membership applications.", "HTTP 200 OK containing JSON array", func(tctx *tests.TestContext) {
		var resp []interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/membership/list", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d membership applications", len(resp))
	})
}

// ============================================================================
// 3. GET /api/membership/districts - Parameterized Tests
// ============================================================================

func TestAPI_Membership_Districts(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Public] GET Districts List", "Retrieves the valid list of districts for membership applications.", "HTTP 200 OK containing districts array", func(tctx *tests.TestContext) {
		var resp []string
		err := tctx.Client.SendHttpRequest("GET", "/api/membership/districts", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}

		if len(resp) == 0 {
			tctx.FailureReason = "Districts list is empty"
			tctx.Errorf("Expected non-empty districts array")
		} else {
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d districts (Example: '%s')", len(resp), resp[0])
		}
	})
}
