package api_tests

import (
	"fmt"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

// GrievancePayload defines the payload for creating a grievance
type GrievancePayload struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

type CreatePayloadTestCase struct {
	Name        string
	Path        string
	Payload     interface{}
	Description string
	Expected    string
	ValidateFn  func(tc *tests.TestContext)
}

// TestAPI_02_CreatePost_TableDriven is a parameterized table-driven test verifying resource creation APIs.
func TestAPI_02_CreatePost_TableDriven(t *testing.T) {
	testCases := []CreatePayloadTestCase{
		{
			Name:        "Create Grievance Payload",
			Path:        "/api/grievances",
			Description: "Verifies creating a new grievance payload with custom bearer authentication token.",
			Expected:    "HTTP 200 OK or 201 Created with successful status message",
			ValidateFn: func(tc *tests.TestContext) {
				payload := &GrievancePayload{
					Name:        "Automated Test Member",
					Email:       "testmember@taga-tn.org",
					Phone:       "9876543210",
					Category:    "General",
					Priority:    "Medium",
					Subject:     "E2E Automation Test Grievance",
					Description: "Verifying API framework creation wrapper with custom auth headers",
				}
				var response map[string]interface{}
				auth := &client.BearerTokenAuth{Token: "super-secret-e2e-token"}

				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, auth)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("Failed to create grievance: %v", err)
					tc.Fatalf("Failed to create grievance: %v", err)
				}
				tc.Actual = fmt.Sprintf("Successfully processed with response: %v", response)
			},
		},
		{
			Name:        "Create Membership Application Payload",
			Path:        "/api/membership/apply",
			Description: "Verifies applying for new membership with complete application payload.",
			Expected:    "HTTP 200 OK or 201 Created",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"tagaId":                    "TAGA_AUTO_TEST",
					"name":                      "Table Driven Application Test",
					"initial":                   "T",
					"gender":                    "male",
					"email":                     "table_driven_app@test.com",
					"mobileNumber":              "9800000999",
					"educationalQualification": "B.Sc Agriculture",
					"designation":              "Assistant Agricultural Officer",
					"workingDistrict":           "Salem",
					"nativeDistrict":            "Salem",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/membership/apply", nil, payload, &response, nil)
				if err != nil {
					tc.FailureReason = fmt.Sprintf("Failed to submit membership application: %v", err)
					tc.Fatalf("Failed to submit membership application: %v", err)
				}
				tc.Actual = fmt.Sprintf("Successfully processed with response: %v", response)
			},
		},
		{
			Name:        "Missing Required Field (Category) - Grievance",
			Path:        "/api/grievances",
			Description: "Verifies that omitting a required field fails validation.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":        "Test User",
					"email":       "test@test.com",
					"phone":       "9999999999",
					"priority":    "Medium",
					"subject":     "Missing category",
					"description": "desc",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected validation error for missing field, got 200/201"
					tc.Errorf("Expected validation error for missing field, got 200/201")
				} else if err.StatusCode() != 400 {
					tc.FailureReason = fmt.Sprintf("Expected 400, got %d", err.StatusCode())
					tc.Errorf("Expected 400, got %d", err.StatusCode())
				} else {
					tc.Actual = "HTTP 400 Bad Request for missing required field"
				}
			},
		},
		{
			Name:        "Malformed JSON Payload - Grievance",
			Path:        "/api/grievances",
			Description: "Verifies that malformed payload body is rejected.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := "invalid-json-body"
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, &payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected error for malformed JSON, got success"
					tc.Errorf("Expected error for malformed JSON, got success")
				} else if err.StatusCode() != 400 {
					tc.FailureReason = fmt.Sprintf("Expected 400, got %d", err.StatusCode())
					tc.Errorf("Expected 400, got %d", err.StatusCode())
				} else {
					tc.Actual = "HTTP 400 Bad Request for malformed JSON"
				}
			},
		},
		{
			Name:        "Missing Auth Token - Admin/Protected Route",
			Path:        "/api/admin/events/create",
			Description: "Verifies that creating a resource on protected route without auth fails.",
			Expected:    "HTTP 401 Unauthorized or 403 Forbidden",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"title": "Test Event",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/admin/events/create", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected 401 Unauthorized, got success"
					tc.Errorf("Expected 401 Unauthorized, got success")
				} else if err.StatusCode() != 401 && err.StatusCode() != 403 {
					tc.FailureReason = fmt.Sprintf("Expected 401/403, got %d", err.StatusCode())
					tc.Errorf("Expected 401/403, got %d", err.StatusCode())
				} else {
					tc.Actual = fmt.Sprintf("HTTP %d Unauthorized/Forbidden for missing auth", err.StatusCode())
				}
			},
		},
		{
			Name:        "Business Logic - Grievance Description Too Short",
			Path:        "/api/grievances",
			Description: "Verifies grievance rejected if description length is below minimum business requirement.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":        "Test User",
					"email":       "test@test.com",
					"phone":       "9999999999",
					"category":    "General",
					"priority":    "Medium",
					"subject":     "Too short",
					"description": "ab",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected validation error for short description, got success"
					tc.Errorf("Expected validation error for short description, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Business Logic - Grievance Empty Subject",
			Path:        "/api/grievances",
			Description: "Verifies grievance rejected if subject is completely empty.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":        "Test User",
					"email":       "test@test.com",
					"phone":       "9999999999",
					"category":    "General",
					"priority":    "Medium",
					"subject":     "",
					"description": "Valid description long enough",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected validation error for empty subject, got success"
					tc.Errorf("Expected validation error for empty subject, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Business Logic - Grievance Non-existent Priority Tier",
			Path:        "/api/grievances",
			Description: "Verifies grievance rejected if priority tier is not in allowed enums.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":        "Test User",
					"email":       "test@test.com",
					"phone":       "9999999999",
					"category":    "General",
					"priority":    "SuperUrgentNonExistentTier",
					"subject":     "Valid Subject",
					"description": "Valid description long enough",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected validation error for invalid enum, got success"
					tc.Errorf("Expected validation error for invalid enum, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Business Logic - Membership Age Under 18",
			Path:        "/api/membership/apply",
			Description: "Verifies application rejected if DOB indicates age under 18.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":             "Underage Applicant",
					"gender":           "male",
					"email":            "underage@test.com",
					"mobileNumber":     "9800000999",
					"dob":              "2020-01-01",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/membership/apply", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for underage applicant, got success"
					tc.Errorf("Expected rejection for underage applicant, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Business Logic - Membership Duplicate Email",
			Path:        "/api/membership/apply",
			Description: "Verifies application rejected if email is already registered.",
			Expected:    "HTTP 400 Bad Request or 409 Conflict",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":             "Duplicate Applicant",
					"gender":           "male",
					"email":            "table_driven_app@test.com",
					"mobileNumber":     "9800000999",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/membership/apply", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for duplicate email, got success"
					tc.Errorf("Expected rejection for duplicate email, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Business Logic - Membership Invalid Enum Gender",
			Path:        "/api/membership/apply",
			Description: "Verifies application rejected if Gender is not a valid enum.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":             "Invalid Gender Applicant",
					"gender":           "ALIEN",
					"email":            "alien@test.com",
					"mobileNumber":     "9800000999",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/membership/apply", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for invalid enum, got success"
					tc.Errorf("Expected rejection for invalid enum, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Extreme Boundary - Grievance 100-Year Date with Max Integer Phone",
			Path:        "/api/grievances",
			Description: "Verifies extreme boundaries on grievance creation.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":        "Test User",
					"email":       "test@test.com",
					"phone":       "92233720368547758079223372036854775807",
					"category":    "General",
					"priority":    "Medium",
					"subject":     "Extreme Bounds",
					"description": "Valid description long enough",
					"incident_date": "2126-08-12",
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected validation error for extreme boundary, got success"
					tc.Errorf("Expected validation error for extreme boundary, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Logical Paradox - Membership Application Future DOB and Negative Age",
			Path:        "/api/membership/apply",
			Description: "Verifies application rejected if DOB is in the future while claiming negative age.",
			Expected:    "HTTP 400 Bad Request",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":             "Paradox Applicant",
					"gender":           "male",
					"email":            "paradox@test.com",
					"mobileNumber":     "9800000999",
					"dob":              "2120-01-01",
					"age":              -25,
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/membership/apply", nil, payload, &response, nil)
				if err == nil {
					tc.FailureReason = "Expected rejection for logical paradox, got success"
					tc.Errorf("Expected rejection for logical paradox, got success")
				} else {
					tc.Actual = fmt.Sprintf("Rejected properly with status: %d", err.StatusCode())
				}
			},
		},
		{
			Name:        "Role Context Switching - Grievance Forcing Admin Role via Payload",
			Path:        "/api/grievances",
			Description: "Verifies that supplying role escalation fields in public grievance creation fails.",
			Expected:    "HTTP 400 Bad Request or safely ignored",
			ValidateFn: func(tc *tests.TestContext) {
				payload := map[string]interface{}{
					"name":        "Role Escalator",
					"email":       "escalate@test.com",
					"phone":       "9999999999",
					"category":    "General",
					"priority":    "Medium",
					"subject":     "Valid Subject",
					"description": "Valid description long enough",
					"role":        "admin",
					"is_admin":    true,
				}
				var response map[string]interface{}
				err := tc.Client.SendHttpRequest("POST", "/api/grievances", nil, payload, &response, nil)
				if err != nil && err.StatusCode() == 500 {
					tc.FailureReason = "Server crashed on role context switch"
					tc.Errorf("Server crashed on role context switch")
				} else {
					tc.Actual = "Handled safely without crashing"
				}
			},
		},
	}

	for _, tc := range testCases {
		testCase := tc
		tests.RunAPITestWithDetails(
			t,
			testCase.Name,
			testCase.Description,
			testCase.Expected,
			testCase.ValidateFn,
		)
	}
}
