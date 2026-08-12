package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"e2e-template/tests"
)

type BookingGuestDetail struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Contact string `json:"contact"`
}

type TowerCreateBookingRequest struct {
	RoomID       string               `json:"roomId"`
	CheckInDate  string               `json:"checkInDate"`
	CheckOutDate string               `json:"checkOutDate"`
	BookerPhone  string               `json:"bookerPhone"`
	BookingFor   string               `json:"bookingFor"`
	BedCount     int                  `json:"bedCount"`
	Gender       string               `json:"gender,omitempty"`
	GuestDetails []BookingGuestDetail `json:"guestDetails,omitempty"`
}

// ============================================================================
// 1. Rooms & Availability - Table-Driven Parameterized Tests
// ============================================================================

func TestAPI_Tower_RoomsAndAvailability_TableDriven(t *testing.T) {
	cases := []struct {
		Name           string
		Path           string
		Description    string
		Expected       string
		ExpectedStatus int
	}{
		{
			Name:           "GET Rooms List",
			Path:           "/api/towers/rooms",
			Description:    "Retrieves all bookable rooms in TAGA Towers.",
			Expected:       "HTTP 200 OK containing rooms list",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "GET Availability",
			Path:           "/api/towers/availability?roomId=gents-dorm&date=2026-12-10",
			Description:    "Checks room availability for a specified date range.",
			Expected:       "HTTP 200 OK containing availability list",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "Validation - Missing RoomId",
			Path:           "/api/towers/availability?date=2026-12-10",
			Description:    "Checks availability without specifying roomId.",
			Expected:       "HTTP 400 Bad Request",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Security - SQL Injection in Date",
			Path:           "/api/towers/availability?roomId=gents-dorm&date=2026-12-10' OR 1=1--",
			Description:    "Attempts SQL injection in the date query parameter.",
			Expected:       "HTTP 400 Bad Request",
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		tests.RunAPITestWithDetails(t, "[Public] "+tc.Name, tc.Description, tc.Expected, func(tctx *tests.TestContext) {
			var resp interface{}
			var errResp map[string]interface{}
			
			if tc.ExpectedStatus == http.StatusOK {
				err := tctx.Client.SendHttpRequest("GET", tc.Path, nil, nil, &resp, nil)
				if err != nil {
					tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
					tctx.Fatalf("Expected 200 OK, got: %v", err)
				}
				tctx.Actual = "HTTP 200 OK, Retrieved availability data"
			} else {
				err := tctx.Client.SendHttpRequest("GET", tc.Path, nil, nil, &errResp, nil)
				assertErrorStatus(tctx, err, tc.ExpectedStatus, "")
			}
		})
	}
}

// ============================================================================
// 2. Booking Life-cycle (Create, Confirm, Cancel) - Parameterized Tests
// ============================================================================

func TestAPI_Tower_BookingWorkflow(t *testing.T) {
	var createdBookingID string

	// Step A: Create Booking (Happy Path)
	tests.RunAPITestWithDetails(t, "[Member] POST Create Booking - Happy Path", "Creates a room booking for gents-dorm.", "HTTP 201 Created with booking details", func(tctx *tests.TestContext) {
		payload := &TowerCreateBookingRequest{
			RoomID:       "gents-dorm",
			CheckInDate:  "2026-12-10",
			CheckOutDate: "2026-12-12",
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
			Gender:       "male",
			GuestDetails: []BookingGuestDetail{
				{Name: "Sudhan Guest", Age: 30, Contact: "9944637254"},
			},
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 201 Created, got: %v", err)
			tctx.Fatalf("Expected 201 Created, got: %v", err)
		}

		id, ok := resp["id"].(string)
		if !ok || id == "" {
			tctx.FailureReason = "Response did not return a valid booking ID"
			tctx.Fatalf("Response missing 'id'")
		}
		createdBookingID = id
		tctx.Actual = fmt.Sprintf("HTTP 201 Created, Booking ID='%s'", createdBookingID)
	})

	// Step B: Get User Bookings
	if createdBookingID != "" {
		tests.RunAPITestWithDetails(t, "[Member] GET User Bookings List", "Retrieves active bookings list for the current member.", "HTTP 200 OK containing booking list", func(tctx *tests.TestContext) {
			var resp []interface{}
			err := tctx.Client.SendHttpRequest("GET", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, nil, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d bookings for user", len(resp))
		})

		// Step C: Confirm Payment
		tests.RunAPITestWithDetails(t, "[Member] POST Confirm Payment", "Confirms payment for the newly created room booking.", "HTTP 200 OK", func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			confirmPayload := map[string]string{"upiId": "test-upi-id"}
			err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings/"+createdBookingID+"/confirm-payment", nil, &confirmPayload, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			tctx.Actual = "HTTP 200 OK, Payment confirmed successfully"
		})
	}

	// Step D: Admin Get All Bookings
	tests.RunAPITestWithDetails(t, "[Admin] GET Admin Bookings Catalog", "Admin lists all tower bookings.", "HTTP 200 OK with all bookings array", func(tctx *tests.TestContext) {
		var resp []interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/towers/admin/bookings", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Admin retrieved %d total bookings", len(resp))
	})

	// Step E: Cancel/Delete Booking
	if createdBookingID != "" {
		tests.RunAPITestWithDetails(t, "[Member] DELETE Cancel Booking", "Cancels/Deletes the booking reservation.", "HTTP 200 OK with success message", func(tctx *tests.TestContext) {
			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest("DELETE", "/api/towers/bookings/"+createdBookingID, nil, nil, &resp, nil)

			if err != nil {
				tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
				tctx.Fatalf("Expected 200 OK, got: %v", err)
			}
			msg, _ := resp["message"].(string)
			if !strings.Contains(strings.ToLower(msg), "cancelled") && !strings.Contains(strings.ToLower(msg), "deleted") && !strings.Contains(strings.ToLower(msg), "success") {
				tctx.FailureReason = fmt.Sprintf("Expected cancel success message, got: %s", msg)
				tctx.Errorf("Unexpected cancellation response: %s", msg)
			}
			tctx.Actual = fmt.Sprintf("HTTP 200 OK, Message='%s'", msg)
		})
	}

	// Step F: Validation Failure - Empty Booking Request
	tests.RunAPITestWithDetails(t, "[Member] POST Create Booking - Empty Payload", "Submits empty payload causing binding validation failure.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings", nil, &TowerCreateBookingRequest{}, &resp, nil)

		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	// Step G: Past Day Booking
	tests.RunAPITestWithDetails(t, "[Member] POST Create Booking - Past Date", "Attempts to book a room for a past date.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		payload := &TowerCreateBookingRequest{
			RoomID:       "gents-dorm",
			CheckInDate:  "2020-01-01",
			CheckOutDate: "2020-01-05",
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	// Step H: Check-Out Before Check-In
	tests.RunAPITestWithDetails(t, "[Member] POST Create Booking - Check-Out Before Check-In", "Attempts to book a room where checkout date is before checkin date.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		payload := &TowerCreateBookingRequest{
			RoomID:       "gents-dorm",
			CheckInDate:  "2026-12-15",
			CheckOutDate: "2026-12-10",
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	// Step I: Booking > 10 Days
	tests.RunAPITestWithDetails(t, "[Member] POST Create Booking - Duration > 10 Days", "Attempts to book a room for more than 10 days.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		payload := &TowerCreateBookingRequest{
			RoomID:       "gents-dorm",
			CheckInDate:  "2026-12-10",
			CheckOutDate: "2026-12-25", // 15 days
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
		}
		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload, &resp, nil)
		assertErrorStatus(tctx, err, http.StatusBadRequest, "")
	})

	// Step J: Overlapping Booking
	tests.RunAPITestWithDetails(t, "[Member] POST Create Booking - Overlapping", "Attempts to book a room that overlaps with an existing booking.", "HTTP 400 Bad Request or 409 Conflict", func(tctx *tests.TestContext) {
		// 1. Create a valid booking
		payload1 := &TowerCreateBookingRequest{
			RoomID:       "ladies-dorm",
			CheckInDate:  "2027-05-10",
			CheckOutDate: "2027-05-15",
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
		}
		var resp1 map[string]interface{}
		err1 := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload1, &resp1, nil)
		
		if err1 != nil {
			tctx.FailureReason = fmt.Sprintf("Failed to setup initial booking for overlap test: %v", err1)
			tctx.Fatalf("Failed to setup initial booking: %v", err1)
		}

		// 2. Attempt overlapping booking
		payload2 := &TowerCreateBookingRequest{
			RoomID:       "ladies-dorm",
			CheckInDate:  "2027-05-12",
			CheckOutDate: "2027-05-18",
			BookerPhone:  "9944637254",
			BookingFor:   "self",
			BedCount:     1,
		}
		var resp2 map[string]interface{}
		err2 := tctx.Client.SendHttpRequest("POST", "/api/towers/bookings?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg", nil, payload2, &resp2, nil)
		
		if err2 == nil {
			tctx.FailureReason = "Expected overlapping booking to fail with 400 or 409, got 201 Created"
			tctx.Errorf("Overlapping booking succeeded!")
		} else if err2.StatusCode() != http.StatusBadRequest && err2.StatusCode() != http.StatusConflict {
			tctx.FailureReason = fmt.Sprintf("Expected 400/409, got %d", err2.StatusCode())
			tctx.Errorf("Expected 400/409, got %d", err2.StatusCode())
		} else {
			tctx.Actual = fmt.Sprintf("Correctly rejected overlapping booking with %d", err2.StatusCode())
		}
	})
}
