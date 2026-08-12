package api_tests

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type OverlapBookingRequest struct {
	RoomID       string `json:"roomId"`
	CheckInDate  string `json:"checkInDate"`
	CheckOutDate string `json:"checkOutDate"`
	BookerPhone  string `json:"bookerPhone"`
	BookingFor   string `json:"bookingFor"`
	BedCount     int    `json:"bedCount"`
	Gender       string `json:"gender,omitempty"`
}

func TestAPI_Tower_BookingOverlapAndGenderRules(t *testing.T) {
	var firstBookingID string
	var savedClient *tests.TestContext

	// Step A: Create first booking for room 'kurinchi' (Capacity: 2, Female, 1 bed)
	tests.RunAPITestWithDetails(t, "[Member] POST Create Initial Booking - Kurinchi", "Creates a confirmed 1-bed female booking for Kurinchi room.", "HTTP 201 Created", func(tctx *tests.TestContext) {
		savedClient = tctx
		payload := &OverlapBookingRequest{
			RoomID:       "kurinchi",
			CheckInDate:  "2026-11-20",
			CheckOutDate: "2026-11-22",
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
			Gender:       "female",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 201 Created, got: %v", err)
			tctx.Fatalf("Expected 201 Created, got: %v", err)
		}

		firstBookingID, _ = resp["id"].(string)
		tctx.Actual = fmt.Sprintf("HTTP 201 Created, Booking ID='%s'", firstBookingID)
	})

	// Step B: Confirm payment of the first booking (necessary to block availability/gender check)
	if firstBookingID != "" {
		tests.RunAPITestWithDetails(t, "[Member] POST Confirm Initial Booking Payment", "Confirms payment to activate the booking restriction.", "HTTP 200 OK", func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			confirmPayload := map[string]string{"upiId": "test-upi-id"}
			err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings/"+firstBookingID+"/confirm-payment", nil, &confirmPayload, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			tctx.Actual = "HTTP 200 OK Payment confirmed"
		})

		// Step C: Attempt second booking in 'kurinchi' with conflicting gender (Male, 1 bed)
		tests.RunAPITestWithDetails(t, "[Member] POST Create Overlapping Booking - Gender Conflict", "Attempts booking a male guest into a room occupied by a female.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
			payload := &OverlapBookingRequest{
				RoomID:       "kurinchi",
				CheckInDate:  "2026-11-20",
				CheckOutDate: "2026-11-22",
				BookerPhone:  "9944637254",
				BookingFor:   "self",
				BedCount:     1,
				Gender:       "male",
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)

			assertErrorStatus(tctx, err, http.StatusBadRequest, "partially occupied by")
		})

		// Step D: Attempt third booking in 'kurinchi' exceeding capacity (Female, 2 beds on top of existing 1 bed)
		tests.RunAPITestWithDetails(t, "[Member] POST Create Overlapping Booking - Capacity Exceeded", "Attempts booking 2 beds in a room with only 1 bed free.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
			payload := &OverlapBookingRequest{
				RoomID:       "kurinchi",
				CheckInDate:  "2026-11-20",
				CheckOutDate: "2026-11-22",
				BookerPhone:  "9944637254",
				BookingFor:   "self",
				BedCount:     2,
				Gender:       "female",
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)

			assertErrorStatus(tctx, err, http.StatusBadRequest, "not enough beds available")
		})
	}

	// Clean up at the end: Cancel the initial booking
	t.Cleanup(func() {
		if savedClient != nil && firstBookingID != "" {
			var resp map[string]interface{}
			_ = savedClient.Client.SendHttpRequest("DELETE", "/api/towers/bookings/"+firstBookingID, nil, nil, &resp, nil)
		}
	})
}

func TestAPI_NegativeScenarios_BookingOverlap(t *testing.T) {
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
			Description:    "Create booking without token",
			Method:         "POST",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Invalid Token - Error Expected",
			Persona:        "Attacker",
			Description:    "Access with fake token",
			Method:         "POST",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "invalid",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Member",
			Description:    "Use PUT instead of POST for booking",
			Method:         "PUT",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "member",
			ExpectedStatus: http.StatusMethodNotAllowed, // or 404
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Member",
			Description:    "SQLi attempt in room ID",
			Method:         "POST",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "member",
			Payload:        map[string]interface{}{"roomId": "room' OR '1'='1", "checkInDate": "2026-11-20", "checkOutDate": "2026-11-22"},
			ExpectedStatus: http.StatusBadRequest, // Or 404
		},
		{
			Name:           "XSS Payload - Error Expected",
			Persona:        "Member",
			Description:    "XSS in booker phone",
			Method:         "POST",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "member",
			Payload:        map[string]interface{}{"roomId": "kurinchi", "checkInDate": "2026-11-20", "checkOutDate": "2026-11-22", "bookerPhone": "<script>alert(1)</script>"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Malformed JSON Payload - Error Expected",
			Persona:        "Member",
			Description:    "Send non-JSON data where expected",
			Method:         "POST",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "member",
			Payload:        "not-a-json-payload-string",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Boundary Values / Missing Fields - Error Expected",
			Persona:        "Member",
			Description:    "Create booking without dates",
			Method:         "POST",
			Endpoint:       "/api/towers/bookings",
			AuthType:       "member",
			Payload:        map[string]interface{}{"roomId": "kurinchi"},
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
