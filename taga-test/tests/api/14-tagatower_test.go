package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"e2e-template/tests"

	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
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

	// Step K: Get Past User Bookings
	tests.RunAPITestWithDetails(t, "[Member] GET Past Bookings List", "Retrieves past/archived bookings for the current member.", "HTTP 200 OK containing booking list", func(tctx *tests.TestContext) {
		var resp []interface{}
		err := tctx.Client.SendHttpRequest("GET", "/api/towers/bookings/past?bookerId=d11348e1-9a65-4945-bb1b-f100a5df15cg&year=2026", nil, nil, &resp, nil)

		if err != nil {
			tctx.FailureReason = fmt.Sprintf("Expected 200 OK, got: %v", err)
			tctx.Fatalf("Expected 200 OK, got: %v", err)
		}
		tctx.Actual = fmt.Sprintf("HTTP 200 OK, Retrieved %d past bookings for user", len(resp))
	})
}

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp dir
	tempDir, err := os.MkdirTemp("", "tagatower_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Update config
	config.Config.Data.Tower.BookingsFile = filepath.Join(tempDir, "bookings.json")
	config.Config.Data.Tower.BookingsArchiveDir = filepath.Join(tempDir, "archive")
	config.Config.Data.Config.TagaTowerRooms = filepath.Join(tempDir, "rooms.json")

	// Create dummy rooms
	rooms := []model.Room{
		{ID: "r1", Name: "Test Room 1", Capacity: 2},
	}
	roomsData, _ := json.Marshal(rooms)
	os.WriteFile(config.Config.Data.Config.TagaTowerRooms, roomsData, 0644)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return tempDir, cleanup
}

func TestBookingArchiveProcess(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Create bookings
	bActive := model.Booking{
		ID:            "BK_ACTIVE",
		RoomID:        "r1",
		BookerID:      "u1",
		CheckInDate:   now,
		CheckOutDate:  now.AddDate(0, 0, 1), // Checkout tomorrow
		PaymentStatus: model.PaymentConfirmed,
	}
	
	bCheckoutToday := model.Booking{
		ID:            "BK_TODAY",
		RoomID:        "r1",
		BookerID:      "u1",
		CheckInDate:   now.AddDate(0, 0, -2),
		CheckOutDate:  todayStart, // Checkout today
		PaymentStatus: model.PaymentConfirmed,
	}

	bPast1 := model.Booking{
		ID:            "BK_PAST_1",
		RoomID:        "r1",
		BookerID:      "u1",
		CheckInDate:   now.AddDate(0, -1, -5),
		CheckOutDate:  now.AddDate(0, -1, -3), // Checkout past month
		PaymentStatus: model.PaymentConfirmed,
	}
	
	bPast2 := model.Booking{
		ID:            "BK_PAST_2",
		RoomID:        "r1",
		BookerID:      "u2", // different user
		CheckInDate:   now.AddDate(0, 0, -5),
		CheckOutDate:  now.AddDate(0, 0, -3), // Checkout past same month
		PaymentStatus: model.PaymentConfirmed,
	}

	// Save active bookings
	initialBookings := []model.Booking{bActive, bCheckoutToday, bPast1, bPast2}
	err := service.SaveAllBookings(initialBookings)
	if err != nil {
		t.Fatalf("Failed to save initial bookings: %v", err)
	}

	// Run archive process
	service.RunBookingArchive()

	// 1. Verify active bookings
	activeBookings, _ := service.ReadAllBookings()
	if len(activeBookings) != 2 {
		t.Errorf("Expected 2 active bookings, got %d", len(activeBookings))
	}
	activeIds := map[string]bool{}
	for _, b := range activeBookings {
		activeIds[b.ID] = true
	}
	if !activeIds["BK_ACTIVE"] {
		t.Errorf("Expected BK_ACTIVE to remain active")
	}
	if !activeIds["BK_TODAY"] {
		t.Errorf("Expected BK_TODAY to remain active (checkout today should not be archived)")
	}

	// 2. Verify archived files
	year1 := bPast1.CheckOutDate.Format("2006")
	month1 := bPast1.CheckOutDate.Format("2006-01")
	file1 := filepath.Join(service.BookingsArchiveDirHelper(), year1, month1+".json")
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Errorf("Expected archive file %s to exist", file1)
	}

	year2 := bPast2.CheckOutDate.Format("2006")
	month2 := bPast2.CheckOutDate.Format("2006-01")
	file2 := filepath.Join(service.BookingsArchiveDirHelper(), year2, month2+".json")
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Errorf("Expected archive file %s to exist", file2)
	}

	// 3. Idempotency test
	service.RunBookingArchive() // Run again
	activeBookings2, _ := service.ReadAllBookings()
	if len(activeBookings2) != 2 {
		t.Errorf("Idempotency check failed: expected 2 active bookings, got %d", len(activeBookings2))
	}

	// 4. Past bookings retrieval
	pastBookingsU1, err := service.GetPastUserBookings("u1", "", "")
	if err != nil {
		t.Fatalf("GetPastUserBookings failed: %v", err)
	}
	if len(pastBookingsU1) != 1 || pastBookingsU1[0].ID != "BK_PAST_1" {
		t.Errorf("Expected past bookings for u1 to contain only BK_PAST_1, got %v", pastBookingsU1)
	}
	
	// 5. Admin occupancy behavior
	adminBookings, err := service.GetAllBookingsForAdmin()
	if err != nil {
		t.Fatalf("GetAllBookingsForAdmin failed: %v", err)
	}
	if len(adminBookings) != 2 {
		t.Errorf("Admin should only see active bookings, expected 2, got %d", len(adminBookings))
	}
}

func TestArchiveFailureHandling(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	now := time.Now()
	bPast := model.Booking{
		ID:            "BK_PAST_FAIL",
		RoomID:        "r1",
		BookerID:      "u1",
		CheckInDate:   now.AddDate(0, 0, -5),
		CheckOutDate:  now.AddDate(0, 0, -3),
		PaymentStatus: model.PaymentConfirmed,
	}
	
	service.SaveAllBookings([]model.Booking{bPast})
	
	// Create directory with bad permissions or block file creation
	year := bPast.CheckOutDate.Format("2006")
	dir := filepath.Join(service.BookingsArchiveDirHelper(), year)
	os.MkdirAll(dir, 0755)
	
	// make month file a directory to cause write failure
	month := bPast.CheckOutDate.Format("2006-01")
	badFile := filepath.Join(dir, month+".json")
	os.Mkdir(badFile, 0755) 

	// Run archive process
	service.RunBookingArchive()

	// Booking should remain in active bookings
	activeBookings, _ := service.ReadAllBookings()
	if len(activeBookings) != 1 || activeBookings[0].ID != "BK_PAST_FAIL" {
		t.Errorf("Booking should remain active when archive fails, got %v", activeBookings)
	}
	
	// Fix permission to allow cleanup
	os.RemoveAll(badFile)
}

func TestAPI_Tower_MixedGenderRulesAndAdvanceCalculation(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Seed standard rooms
	rooms := []model.Room{
		{ID: "apex-1", Name: "Apex Suite A/C", Type: model.RoomTypeApexSuite, Capacity: 3, AllowSingleBed: true},
		{ID: "kurinchi", Name: "Kurinchi", Type: model.RoomTypeACRoom, Capacity: 2, AllowSingleBed: true},
		{ID: "gents-dorm", Name: "Gents Dormitory", Type: model.RoomTypeGentsDorm, Capacity: 12, AllowSingleBed: true},
		{ID: "ladies-dorm", Name: "Ladies Dormitory", Type: model.RoomTypeLadiesDorm, Capacity: 8, AllowSingleBed: true},
	}
	roomsData, _ := json.Marshal(rooms)
	os.WriteFile(config.Config.Data.Config.TagaTowerRooms, roomsData, 0644)
	service.SaveAllBookings([]model.Booking{})

	now := time.Now().AddDate(0, 0, 1)
	checkInStr := now.Format("2006-01-02")
	checkOutStr := now.AddDate(0, 0, 2).Format("2006-01-02")

	// 1. Test Mixed Gender Couple Booking in 2-bed room (Kurinchi)
	reqKurinchi := model.CreateBookingRequest{
		RoomID:       "kurinchi",
		CheckInDate:  checkInStr,
		CheckOutDate: checkOutStr,
		BookerPhone:  "+919876543210",
		BookingFor:   model.BookingForGuest,
		BedCount:     2,
		GuestDetails: []model.GuestDetail{
			{Name: "John Doe", Age: 30, Contact: "+919876543210", Gender: model.GenderMale},
			{Name: "Jane Doe", Age: 28, Contact: "+919876543211", Gender: model.GenderFemale},
		},
	}
	resKurinchi, err := service.CreateBooking(reqKurinchi, "Member User", "M001")
	if err != nil {
		t.Fatalf("Mixed gender couple booking in 2-bed room should succeed, got: %v", err)
	}
	if resKurinchi.AdvanceAmount != 2 {
		t.Errorf("Expected advance amount for 2 beds to be 2, got %d", resKurinchi.AdvanceAmount)
	}
	if resKurinchi.Gender != model.GenderMixed {
		t.Errorf("Expected booking gender to be 'mixed', got %s", resKurinchi.Gender)
	}

	// 2. Test Mixed Gender Couple Booking in Apex Suite (2 beds out of 3)
	reqApex := model.CreateBookingRequest{
		RoomID:       "apex-1",
		CheckInDate:  checkInStr,
		CheckOutDate: checkOutStr,
		BookerPhone:  "+919876543210",
		BookingFor:   model.BookingForGuest,
		BedCount:     2,
		GuestDetails: []model.GuestDetail{
			{Name: "Alex Smith", Age: 35, Contact: "+919876543212", Gender: model.GenderMale},
			{Name: "Mary Smith", Age: 32, Contact: "+919876543213", Gender: model.GenderFemale},
		},
	}
	resApex, err := service.CreateBooking(reqApex, "Member User", "M001")
	if err != nil {
		t.Fatalf("Mixed couple booking in Apex Suite should succeed, got: %v", err)
	}
	if resApex.AdvanceAmount != 2 {
		t.Errorf("Expected advance amount for 2 beds in Apex to be 2, got %d", resApex.AdvanceAmount)
	}

	// 3. Verify Apex Suite availability: the 3rd bed must be blocked and room marked fully booked
	checkInTime, _ := time.Parse("2006-01-02", checkInStr)
	checkOutTime, _ := time.Parse("2006-01-02", checkOutStr)
	availMap, err := service.CheckAllRoomsAvailabilityRange(checkInTime, checkOutTime)
	if err != nil {
		t.Fatalf("CheckAllRoomsAvailabilityRange failed: %v", err)
	}
	apexAvail, ok := availMap["apex-1"]
	if !ok {
		t.Fatalf("Apex Suite not found in availability map")
	}
	if apexAvail.Available || apexAvail.AvailableBeds != 0 {
		t.Errorf("Expected Apex Suite 3rd bed to be blocked (0 beds available, Available=false), got Available=%v, Beds=%d",
			apexAvail.Available, apexAvail.AvailableBeds)
	}

	// 4. Test Single person attempting to book the 3rd bed in Apex on same dates -> Should be rejected
	reqApex3rdBed := model.CreateBookingRequest{
		RoomID:       "apex-1",
		CheckInDate:  checkInStr,
		CheckOutDate: checkOutStr,
		BookerPhone:  "+919876543214",
		BookingFor:   model.BookingForSelf,
		BedCount:     1,
		Gender:       model.GenderMale,
	}
	_, err = service.CreateBooking(reqApex3rdBed, "Other User", "M002")
	if err == nil {
		t.Fatalf("Expected booking attempt on 3rd bed of Apex during couple reservation to fail, but it succeeded")
	}

	// 5. Test Dormitory strict gender rule
	reqDormMixed := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  checkInStr,
		CheckOutDate: checkOutStr,
		BookerPhone:  "+919876543210",
		BookingFor:   model.BookingForGuest,
		BedCount:     2,
		GuestDetails: []model.GuestDetail{
			{Name: "Gents 1", Age: 25, Contact: "+919876543210", Gender: model.GenderMale},
			{Name: "Lady 1", Age: 24, Contact: "+919876543211", Gender: model.GenderFemale},
		},
	}
	_, err = service.CreateBooking(reqDormMixed, "Member User", "M001")
	if err == nil {
		t.Fatalf("Expected mixed booking in Gents Dorm to fail, but it succeeded")
	}

	// 6. Test Multi-bed Advance calculation (e.g. 5 beds = 5)
	reqDorm5Beds := model.CreateBookingRequest{
		RoomID:       "gents-dorm",
		CheckInDate:  now.AddDate(0, 1, 0).Format("2006-01-02"),
		CheckOutDate: now.AddDate(0, 1, 1).Format("2006-01-02"),
		BookerPhone:  "+919876543210",
		BookingFor:   model.BookingForGuest,
		BedCount:     5,
		GuestDetails: []model.GuestDetail{
			{Name: "G1", Age: 25, Contact: "+919876543210", Gender: model.GenderMale},
			{Name: "G2", Age: 26, Contact: "+919876543210", Gender: model.GenderMale},
			{Name: "G3", Age: 27, Contact: "+919876543210", Gender: model.GenderMale},
			{Name: "G4", Age: 28, Contact: "+919876543210", Gender: model.GenderMale},
			{Name: "G5", Age: 29, Contact: "+919876543210", Gender: model.GenderMale},
		},
	}
	resDorm5, err := service.CreateBooking(reqDorm5Beds, "Member User", "M001")
	if err != nil {
		t.Fatalf("5 beds booking in Gents Dorm failed: %v", err)
	}
	if resDorm5.AdvanceAmount != 5 {
		t.Errorf("Expected advance for 5 beds to be 5, got %d", resDorm5.AdvanceAmount)
	}
}
