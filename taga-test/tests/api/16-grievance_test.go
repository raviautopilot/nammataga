package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/tests"
)

type GrievancePayloadLocal struct {
	Subject           string `json:"subject"`
	Category          string `json:"category"`
	Priority          string `json:"priority"`
	Description       string `json:"description"`
	ContactPhone      string `json:"contactPhone"`
	PreferredResponse string `json:"preferredResponse"`
}

type GrievanceTestCase struct {
	Name           string
	Persona        string
	Description    string
	Payload        interface{}
	ExpectedStatus int
	ExpectedSub    string
}

// ============================================================================
// 1. POST /api/grievances - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Grievance_Submit_TableDriven(t *testing.T) {
	cases := []GrievanceTestCase{
		{
			Name:        "Happy Path - Create Member Grievance",
			Persona:     "Legitimate Member",
			Description: "Submits a new grievance issue report with valid parameters.",
			Payload: &GrievancePayloadLocal{
				Subject:           "Delay in Subscription Activation",
				Category:          "Membership",
				Priority:          "High",
				Description:       "Paid subscription 2 days ago but still shows unpaid on profile.",
				ContactPhone:      "9944637254",
				PreferredResponse: "Email",
			},
			ExpectedStatus: http.StatusOK,
			ExpectedSub:    "GRV-",
		},
		{
			Name:        "Validation - Malformed Request Body",
			Persona:     "Buggy Client",
			Description: "Submits an invalid JSON value to the endpoint.",
			Payload: func() *string {
				s := "bad-payload-format"
				return &s
			}(),
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:        "Validation - Missing Required Category",
			Persona:     "Legitimate Member",
			Description: "Submits grievance without category.",
			Payload: &GrievancePayloadLocal{
				Subject: "Test Subject",
				Description: "Test Description",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:        "Security - SQLi in Subject",
			Persona:     "Malicious User",
			Description: "Attempts SQL injection in grievance subject.",
			Payload: &GrievancePayloadLocal{
				Subject: "' OR 1=1 --",
				Category: "General",
				Description: "Test Description",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:        "Business Logic - Invalid Category",
			Persona:     "Legitimate Member",
			Description: "Submits grievance with non-existent category.",
			Payload: &GrievancePayloadLocal{
				Subject:           "Test",
				Category:          "NonExistentCategory",
				Priority:          "High",
				Description:       "Test",
				ContactPhone:      "9944637254",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:        "Business Logic - Description Too Long",
			Persona:     "Legitimate Member",
			Description: "Submits grievance with a description exceeding max length limit.",
			Payload: &GrievancePayloadLocal{
				Subject:           "Test",
				Category:          "Membership",
				Priority:          "High",
				Description:       strings.Repeat("A", 10001),
				ContactPhone:      "9944637254",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		testName := fmt.Sprintf("[%s] POST Grievance - %s", tc.Persona, tc.Name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.ExpectedStatus)
		if tc.ExpectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.ExpectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.Description, expectedStr, func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/grievances", nil, tc.Payload, &resp, nil)

			if tc.ExpectedStatus == http.StatusOK {
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				id, _ := resp["id"].(string)
				if !strings.HasPrefix(id, tc.ExpectedSub) {
					tctx.FailureReason = fmt.Sprintf("Expected grievance ID starting with '%s', got '%s'", tc.ExpectedSub, id)
					tctx.Errorf("Invalid grievance ID format: %s", id)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Grievance ID='%s'", id)
			} else {
				assertErrorStatus(tctx, err, tc.ExpectedStatus, tc.ExpectedSub)
			}
		})
	}
}

// ============================================================================
// 2. GET /api/grievances & Single Retrieval - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Grievance_ListAndGet_TableDriven(t *testing.T) {
	cases := []struct {
		Name           string
		Path           string
		Description    string
		Expected       string
		ExpectedStatus int
		ExpectedSub    string
	}{
		{
			Name:           "GET Grievances List",
			Path:           "/api/grievances",
			Description:    "Retrieves the full list of submitted grievances.",
			Expected:       "HTTP 200 OK with grievances array",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "GET Non-existent Grievance",
			Path:           "/api/grievances/GRV-nonexistent",
			Description:    "Attempts retrieving details for an invalid grievance ID.",
			Expected:       "HTTP 404 Not Found",
			ExpectedStatus: http.StatusNotFound,
			ExpectedSub:    "Grievance not found",
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Public] "+tc.Name, tc.Description, tc.Expected, func(tctx *tests.TestContext) {
			if tc.ExpectedStatus == http.StatusOK {
				var resp []interface{}
				err := tctx.Client.SendHttpRequest("GET", tc.Path, nil, nil, &resp, nil)
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d grievances", len(resp))
			} else {
				var resp map[string]interface{}
				err := tctx.Client.SendHttpRequest("GET", tc.Path, nil, nil, &resp, nil)
				assertErrorStatus(tctx, err, tc.ExpectedStatus, tc.ExpectedSub)
			}
		})
	}
}

// ============================================================================
// 3. Update & Delete Non-existent Grievances - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Grievance_UpdateAndDeleteNonExistent_TableDriven(t *testing.T) {
	cases := []struct {
		Name           string
		Method         string
		Path           string
		Payload        interface{}
		Description    string
		ExpectedStatus int
		ExpectedSub    string
	}{
		{
			Name:   "PUT Update Non-existent Grievance",
			Method: "PUT",
			Path:   "/api/grievances/GRV-nonexistent",
			Payload: &GrievancePayloadLocal{
				Subject: "Update Subject",
			},
			Description:    "Attempts updating fields of an invalid grievance ID.",
			ExpectedStatus: http.StatusNotFound,
			ExpectedSub:    "Grievance not found",
		},
		{
			Name:           "DELETE Non-existent Grievance",
			Method:         "DELETE",
			Path:           "/api/grievances/GRV-nonexistent",
			Payload:        nil,
			Description:    "Attempts deleting an invalid grievance ID.",
			ExpectedStatus: http.StatusNotFound,
			ExpectedSub:    "Grievance not found",
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Public] "+tc.Name, tc.Description, fmt.Sprintf("HTTP %d", tc.ExpectedStatus), func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest(tc.Method, tc.Path, nil, tc.Payload, &resp, nil)
			assertErrorStatus(tctx, err, tc.ExpectedStatus, tc.ExpectedSub)
		})
	}
}
