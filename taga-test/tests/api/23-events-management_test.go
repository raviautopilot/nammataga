package api_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type EventDetail struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type EventCreateResponse struct {
	Message string      `json:"message"`
	Event   EventDetail `json:"event"`
}

func TestAPI_EventsManagement_E2E(t *testing.T) {
	var adminToken string
	var createdEventID string
	var savedClient *client.Client

	// Step A: Create Event (Admin Auth, Form Post)
	tests.RunAPITestWithDetails(t, "[Admin] POST Create Event", "Creates a new calendar event using form-urlencoded values.", "HTTP 200 OK containing event details", func(tctx *tests.TestContext) {
		savedClient = tctx.Client
		adminToken = getValidAdminToken(tctx.T, tctx.Client)

		reqUrl := tctx.Client.BaseURL + "/api/admin/events/create"
		formData := url.Values{}
		formData.Set("title", "E2E Temporary Summit")
		formData.Set("date", "2026-12-18")
		formData.Set("location", "Salem HQ")
		formData.Set("description", "Temporary summit for E2E validation")
		formData.Set("status", "upcoming")

		req, err := http.NewRequest("POST", reqUrl, strings.NewReader(formData.Encode()))
		if err != nil {
			tctx.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := tctx.Client.HTTPClient.Do(req)
		if err != nil {
			tctx.Fatalf("Failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %d", resp.StatusCode)
			tctx.Fatalf("Expected 200 OK, got: %d", resp.StatusCode)
		}

		var eventResp EventCreateResponse
		if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
			tctx.Fatalf("Failed to parse response JSON: %v", err)
		}

		createdEventID = eventResp.Event.ID
		if createdEventID == "" {
			tctx.FailureReason = "Response event did not contain a valid ID"
			tctx.Fatalf("Response event missing ID")
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Event ID='%s'", createdEventID)
	})

	// Step B: Fetch Events List (Public check)
	tests.RunAPITestWithDetails(t, "[Public] GET Events List", "Retrieves active events calendar.", "HTTP 200 OK containing events array", func(tctx *tests.TestContext) {
		var resp []interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/events/upcoming", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, retrieved %d events", len(resp))
	})

	// Step C: Update Event (Admin Auth, JSON PUT)
	if createdEventID != "" && adminToken != "" {
		tests.RunAPITestWithDetails(t, "[Admin] PUT Update Event Details", "Updates details of the created event using JSON.", "HTTP 200 OK", func(tctx *tests.TestContext) {
			auth := &client.BearerTokenAuth{Token: adminToken}
			payload := map[string]interface{}{
				"title":       "E2E Temporary Summit Updated",
				"location":    "Salem Taga Towers",
				"description": "Updated description for E2E summit",
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("PUT", "/api/admin/events/"+createdEventID, nil, &payload, &resp, auth)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			tctx.Actual = "HTTP 200 OK event details updated successfully"
		})
	}

	// Cleanup: Delete the event we created
	t.Cleanup(func() {
		if savedClient != nil && createdEventID != "" && adminToken != "" {
			auth := &client.BearerTokenAuth{Token: adminToken}
			var resp map[string]interface{}
			_ = savedClient.SendHttpRequest("DELETE", "/api/admin/events/"+createdEventID, nil, nil, &resp, auth)
		}
	})
}

func TestAPI_NegativeScenarios_EventsManagement(t *testing.T) {
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
			Description:    "Create event without token",
			Method:         "POST",
			Endpoint:       "/api/admin/events/create",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Invalid Token - Error Expected",
			Persona:        "Attacker",
			Description:    "Access with fake token",
			Method:         "POST",
			Endpoint:       "/api/admin/events/create",
			AuthType:       "invalid",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Admin",
			Description:    "Use GET instead of POST",
			Method:         "GET",
			Endpoint:       "/api/admin/events/create",
			AuthType:       "admin",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Admin",
			Description:    "SQLi attempt in event update",
			Method:         "PUT",
			Endpoint:       "/api/admin/events/123' OR '1'='1",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "XSS Payload - Error Expected",
			Persona:        "Admin",
			Description:    "XSS in event update path",
			Method:         "PUT",
			Endpoint:       "/api/admin/events/<script>alert(1)</script>",
			AuthType:       "admin",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Boundary Values / Missing Fields - Error Expected",
			Persona:        "Admin",
			Description:    "Update with empty payload",
			Method:         "PUT",
			Endpoint:       "/api/admin/events/123",
			AuthType:       "admin",
			Payload:        map[string]interface{}{},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Business Logic - Past Event Date",
			Persona:        "Admin",
			Description:    "Create an event in the past",
			Method:         "POST",
			Endpoint:       "/api/admin/events/create",
			AuthType:       "admin",
			Payload:        map[string]interface{}{
				"title": "Past Event",
				"date": "2020-01-01",
				"location": "Nowhere",
				"description": "This is in the past",
				"status": "upcoming",
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "State Machine Violation - Modify Completed Event",
			Persona:        "Admin",
			Description:    "Update an event that is already marked as completed",
			Method:         "PUT",
			Endpoint:       "/api/admin/events/completed_event_123",
			AuthType:       "admin",
			Payload:        map[string]interface{}{"status": "upcoming"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Logical Paradox - End Before Start",
			Persona:        "Admin",
			Description:    "Event dates contradict each other",
			Method:         "POST",
			Endpoint:       "/api/admin/events/create",
			AuthType:       "admin",
			Payload:        map[string]interface{}{"title": "Paradox Event", "date": "2026-12-01", "end_date": "2026-11-01"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Extreme Boundary - 100 Year Duration",
			Persona:        "Admin",
			Description:    "Event spanning a century",
			Method:         "POST",
			Endpoint:       "/api/admin/events/create",
			AuthType:       "admin",
			Payload:        map[string]interface{}{"title": "Century Event", "date": "2026-01-01", "end_date": "2126-01-01"},
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
