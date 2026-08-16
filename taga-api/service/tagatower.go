package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"taga-api/config"
	"taga-api/model"
)

/*
	---------------------------
	  Helper: Get Base Path

---------------------------
*/
func getFilePath(pathParts ...string) string {
	wd, _ := os.Getwd()
	allParts := append([]string{wd}, pathParts...)
	return filepath.Join(allParts...)
}

/*
	---------------------------
	  File Paths (Dynamic)

---------------------------
*/
func getRoomsFilePath() string {
	path := config.Config.Data.Config.TagaTowerRooms
	if path == "" {
		return "data/config/taga-tower-rooms.json"
	}
	return path
}

// bookingsFilePath returns the path to the bookings JSON file.
// Reads from config at call time so VPS deployments with absolute paths in
// config.json work correctly regardless of the binary's working directory.
func bookingsFilePath() string {
	path := config.Config.Data.Tower.BookingsFile
	if path == "" {
		return getFilePath("data", "towers", "bookings.json")
	}
	return path
}

func ReadAllBookings() ([]model.Booking, error) {
	file, err := os.ReadFile(bookingsFilePath())
	if err != nil {
		return []model.Booking{}, nil
	}

	var bookings []model.Booking
	if len(file) == 0 {
		return []model.Booking{}, nil
	}

	if err := json.Unmarshal(file, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

func SaveAllBookings(bookings []model.Booking) error {
	data, err := json.MarshalIndent(bookings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(bookingsFilePath(), data, 0644)
}

/*
	---------------------------
	  Read Rooms

---------------------------
*/
func ReadRooms() ([]model.Room, error) {
	file, err := os.ReadFile(getRoomsFilePath())
	if err != nil {
		log.Printf("Error reading rooms file: %v", err)
		return nil, err
	}

	var rooms []model.Room
	if err := json.Unmarshal(file, &rooms); err != nil {
		log.Printf("Error parsing rooms JSON: %v", err)
		return nil, err
	}

	return rooms, nil
}

/*
	---------------------------
	  Get Room By ID

---------------------------
*/
func GetRoomByID(roomID string) (*model.Room, error) {
	rooms, err := ReadRooms()
	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		if room.ID == roomID {
			return &room, nil
		}
	}

	return nil, fmt.Errorf("room not found: %s", roomID)
}



/*
	---------------------------
	  Create Booking

---------------------------
*/
func CreateBooking(req model.CreateBookingRequest, bookerName, bookerID string) (*model.BookingResponse, error) {
	// Parse date
	checkInDate, err := time.Parse("2006-01-02", req.CheckInDate)
	if err != nil {
		return nil, fmt.Errorf("invalid check-in date")
	}

	checkOutDate, err := time.Parse("2006-01-02", req.CheckOutDate)
	if err != nil {
		return nil, fmt.Errorf("invalid check-out date")
	}

	// Get room
	room, err := GetRoomByID(req.RoomID)
	if err != nil {
		return nil, err
	}

	// Determine advance amount
	advanceAmount := 100
	if req.BookingFor == model.BookingForGuest {
		advanceAmount = 200
	}

	// Create booking object (not saved yet)
	booking := model.Booking{
		ID:            fmt.Sprintf("BK%d", time.Now().UnixNano()),
		RoomID:        req.RoomID,
		CheckInDate:   checkInDate,
		CheckOutDate:  checkOutDate,
		BookerName:    bookerName,
		BookerID:      bookerID,
		BookerPhone:   req.BookerPhone,
		BookingFor:    req.BookingFor,
		BedCount:      req.BedCount,
		Gender:        req.Gender,
		GuestDetails:  req.GuestDetails,
		PaymentStatus: model.PaymentPending,
		UpiID:         req.UpiID,
		AdvanceAmount: advanceAmount,
		CreatedAt:     time.Now(),
	}

	// 🔥 Read all bookings
	allBookingsData, err := ReadAllBookings()
	if err != nil {
		return nil, err
	}

	// ── VALIDATION 1: Bed capacity overlap check ──
	for d := checkInDate; d.Before(checkOutDate); d = d.AddDate(0, 0, 1) {
		occupiedBeds := 0

		for _, b := range allBookingsData {
			if b.RoomID != req.RoomID {
				continue
			}
			// Only confirmed bookings block availability
			if b.PaymentStatus != model.PaymentConfirmed {
				continue
			}
			// Check date overlap: existing booking overlaps day d
			if d.Before(b.CheckOutDate) && d.After(b.CheckInDate.Add(-time.Second)) {
				occupiedBeds += b.BedCount
			}
		}

		requestedBeds := req.BedCount
		if !room.AllowSingleBed {
			// Whole-room booking
			requestedBeds = room.Capacity
		}

		if occupiedBeds+requestedBeds > room.Capacity {
			return nil, fmt.Errorf("not enough beds available on %s (only %d bed(s) free)", d.Format("Jan 02"), room.Capacity-occupiedBeds)
		}
	}

	// ── VALIDATION 1.5: Strict gender check for Dormitories ──
	if room.Type == "gents-dorm" && req.Gender != "male" {
		return nil, fmt.Errorf("this room is strictly for male guests only")
	}
	if room.Type == "ladies-dorm" && req.Gender != "female" {
		return nil, fmt.Errorf("this room is strictly for female guests only")
	}

	// ── VALIDATION 2: Gender restriction check for AC rooms with partial booking ──
	if room.AllowSingleBed {
		// Find existing confirmed bookings for this room that overlap our date range
		for _, existingBooking := range allBookingsData {
			// Skip non-matching rooms and non-confirmed bookings
			if existingBooking.RoomID != req.RoomID {
				continue
			}
			if existingBooking.PaymentStatus != model.PaymentConfirmed {
				continue
			}

			// Check if the existing booking overlaps with our requested dates
			if checkInDate.Before(existingBooking.CheckOutDate) &&
				checkOutDate.After(existingBooking.CheckInDate) {

				// If existing booking has a gender set and it doesn't match the new request
				if existingBooking.Gender != "" && req.Gender != "" &&
					existingBooking.Gender != req.Gender {
					return nil, fmt.Errorf("this room is partially occupied by %s guests — only %s guests can book the remaining beds",
						existingBooking.Gender, existingBooking.Gender)
				}
			}
		}
	}

	// 🔥 Append and save
	allBookingsData = append(allBookingsData, booking)

	if err := SaveAllBookings(allBookingsData); err != nil {
		return nil, err
	}

	// Return response (with computed status for consistency)
	now := time.Now()
	bookingStatus := calculateBookingStatus(booking, now)

	return &model.BookingResponse{
		ID:            booking.ID,
		RoomID:        booking.RoomID,
		RoomName:      room.Name,
		CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
		CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
		BookerName:    booking.BookerName,
		BookerID:      booking.BookerID,
		BookingFor:    booking.BookingFor,
		BedCount:      booking.BedCount,
		Gender:        booking.Gender,
		GuestDetails:  booking.GuestDetails,
		PaymentStatus: booking.PaymentStatus,
		AdvanceAmount: booking.AdvanceAmount,
		BookingStatus: bookingStatus,
	}, nil
}

/*
	---------------------------
	  Check Room Availability

---------------------------
*/
func CheckRoomAvailability(room *model.Room, date time.Time, bookings []model.Booking) (bool, int, string, error) {

	filtered := []model.Booking{}
	for _, b := range bookings {

		if b.RoomID != room.ID {
			continue
		}

		if b.PaymentStatus != model.PaymentConfirmed {
			continue
		}

		// 🔥 DATE OVERLAP
		if date.Before(b.CheckOutDate) && date.After(b.CheckInDate.Add(-time.Second)) {
			filtered = append(filtered, b)
		}
	}

	// ✅ Count occupied beds
	occupiedBeds := 0
	var roomGender model.Gender

	for i, b := range filtered {
		occupiedBeds += b.BedCount

		if i == 0 && b.Gender != "" {
			roomGender = b.Gender
		}
	}

	availableBeds := room.Capacity - occupiedBeds

	// ✅ Gender restriction (AC rooms with partial booking)
	var genderRestriction string
	if room.AllowSingleBed && occupiedBeds > 0 && availableBeds > 0 {
		genderRestriction = string(roomGender)
	}

	return availableBeds > 0, availableBeds, genderRestriction, nil
}

/*
	---------------------------
	  Get Bookings for User

---------------------------
*/
func GetUserBookings(bookerID string) ([]model.BookingResponse, error) {

	allBookingsData, err := ReadAllBookings()
	if err != nil {
		return nil, err
	}

	rooms, err := ReadRooms()
	if err != nil {
		return nil, err
	}

	roomMap := make(map[string]*model.Room)
	for i, r := range rooms {
		roomMap[r.ID] = &rooms[i]
	}

	var allBookings []model.BookingResponse
	now := time.Now() // ✅ Get current time once for consistent calculations

	for _, booking := range allBookingsData {
		if booking.BookerID == bookerID {
			room := roomMap[booking.RoomID]
			roomName := ""
			if room != nil {
				roomName = room.Name
			}

			// ✅ Calculate booking status
			bookingStatus := calculateBookingStatus(booking, now)

			allBookings = append(allBookings, model.BookingResponse{
				ID:            booking.ID,
				RoomID:        booking.RoomID,
				RoomName:      roomName,
				CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
				CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
				BookerName:    booking.BookerName,
				BookerID:      booking.BookerID,
				BookingFor:    booking.BookingFor,
				BedCount:      booking.BedCount,
				Gender:        booking.Gender,
				GuestDetails:  booking.GuestDetails,
				PaymentStatus: booking.PaymentStatus,
				AdvanceAmount: booking.AdvanceAmount,
				BookingStatus: bookingStatus,
			})
		}
	}

	return allBookings, nil
}

/*
	---------------------------
	  Get ALL Bookings (Admin)
---------------------------
*/

// GetAllBookingsForAdmin returns all bookings across all users.
// Used by the admin occupancy schedule tab.
func GetAllBookingsForAdmin() ([]model.BookingResponse, error) {
	// ── CHANGE 7: New service function for admin occupancy view ──
	allBookingsData, err := ReadAllBookings()
	if err != nil {
		return nil, err
	}

	rooms, err := ReadRooms()
	if err != nil {
		return nil, err
	}

	roomMap := make(map[string]*model.Room)
	for i, r := range rooms {
		roomMap[r.ID] = &rooms[i]
	}

	var result []model.BookingResponse
	now := time.Now() // ✅ Get current time

	for _, booking := range allBookingsData {
		room := roomMap[booking.RoomID]
		roomName := ""
		if room != nil {
			roomName = room.Name
		}

		// ✅ Calculate booking status
		bookingStatus := calculateBookingStatus(booking, now)

		result = append(result, model.BookingResponse{
			ID:            booking.ID,
			RoomID:        booking.RoomID,
			RoomName:      roomName,
			CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
			CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
			BookerName:    booking.BookerName,
			BookerID:      booking.BookerID,
			BookingFor:    booking.BookingFor,
			BedCount:      booking.BedCount,
			Gender:        booking.Gender,
			GuestDetails:  booking.GuestDetails,
			PaymentStatus: booking.PaymentStatus,
			AdvanceAmount: booking.AdvanceAmount,
			BookingStatus: bookingStatus,
		})
	}

	return result, nil
}

// calculateBookingStatus determines the current status of a booking
// based on dates and payment status.
// Returns: "upcoming", "active", "completed", "cancelled"
func calculateBookingStatus(booking model.Booking, now time.Time) string {
	// Cancelled is final - no date check needed
	if booking.PaymentStatus == model.PaymentCancelled || booking.PaymentStatus == model.PaymentRefunded {
		return "cancelled"
	}

	// Normal date-based status for confirmed/pending bookings
	if now.Before(booking.CheckInDate) {
		return "upcoming"
	}

	if now.After(booking.CheckOutDate) || now.Equal(booking.CheckOutDate) {
		return "completed"
	}

	return "active"
}

/*
	---------------------------
	  Cancel Booking
---------------------------
*/

// CancelBooking marks any booking as cancelled regardless of payment status.
// ── CHANGE 6: Refund logic removed — cancellation is always a simple status update.
func CancelBooking(bookingID string) error {
	allBookings, err := ReadAllBookings()
	if err != nil {
		return err
	}

	for i := range allBookings {
		if allBookings[i].ID == bookingID {
			allBookings[i].PaymentStatus = model.PaymentCancelled
			return SaveAllBookings(allBookings)
		}
	}

	return fmt.Errorf("booking not found")
}



/*
	---------------------------
	  Confirm Payment

---------------------------
*/
func ConfirmPayment(bookingID, upiID string) error {
	allBookings, err := ReadAllBookings()
	if err != nil {
		return err
	}

	for i := range allBookings {

		if allBookings[i].ID == bookingID {

			allBookings[i].PaymentStatus = model.PaymentConfirmed
			allBookings[i].UpiID = upiID

			return SaveAllBookings(allBookings)
		}
	}

	return fmt.Errorf("booking not found")
}

func ConfirmPaymentWithDetails(bookingID, orderID, paymentID string) error {

	allBookings, err := ReadAllBookings()
	if err != nil {
		return err
	}

	for i := range allBookings {

		if allBookings[i].ID == bookingID {

			allBookings[i].PaymentStatus = model.PaymentConfirmed
			allBookings[i].RazorpayOrderID = orderID
			allBookings[i].RazorpayPaymentID = paymentID

			return SaveAllBookings(allBookings)
		}
	}

	return fmt.Errorf("booking not found")
}



func GetBookingByID(bookingID string) (*model.Booking, error) {

	allBookings, err := ReadAllBookings()
	if err != nil {
		return nil, err
	}

	for _, b := range allBookings {
		if b.ID == bookingID {
			return &b, nil
		}
	}

	return nil, fmt.Errorf("booking not found")
}

/*
	---------------------------
	  Daily Cleanup — Delete bookings older than 6 months
---------------------------
*/

// StartBookingCleanupScheduler runs a daily goroutine that removes bookings
// whose checkOutDate is more than 6 months in the past.
// ── CHANGE 5: Called from main.go on startup ──
func StartBookingCleanupScheduler() {
	go func() {
		log.Println("✅ Booking cleanup scheduler started (runs daily)")
		for {
			runBookingCleanup()
			// Sleep until next midnight
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
		}
	}()
}

// runBookingCleanup deletes all bookings whose checkOutDate is older than 6 months.
func runBookingCleanup() {
	cutoff := time.Now().AddDate(0, -6, 0) // 6 months ago

	allBookings, err := ReadAllBookings()
	if err != nil {
		log.Printf("❌ Booking cleanup: failed to read bookings: %v", err)
		return
	}

	kept := []model.Booking{}
	removed := 0

	for _, b := range allBookings {
		if b.CheckOutDate.Before(cutoff) {
			removed++
			log.Printf("🗑️  Cleanup: removing old booking %s (checkOut: %s)", b.ID, b.CheckOutDate.Format("2006-01-02"))
		} else {
			kept = append(kept, b)
		}
	}

	if removed == 0 {
		log.Println("✅ Booking cleanup: no old bookings to remove")
		return
	}

	if err := SaveAllBookings(kept); err != nil {
		log.Printf("❌ Booking cleanup: failed to save: %v", err)
		return
	}

	log.Printf("✅ Booking cleanup: removed %d old booking(s)", removed)
}

/*
	---------------------------
	  Bulk Range Availability
	  Checks ALL rooms for ALL dates in the range in one shot.
	  Returns a map: roomID → RoomAvailability (worst case across all days)
---------------------------
*/
func CheckAllRoomsAvailabilityRange(checkIn, checkOut time.Time) (map[string]model.RoomAvailability, error) {
	rooms, err := ReadRooms()
	if err != nil {
		return nil, err
	}

	allBookings, err := ReadAllBookings()
	if err != nil {
		return nil, err
	}

	// Pre-filter: only confirmed bookings that overlap [checkIn, checkOut)
	var relevantBookings []model.Booking
	for _, b := range allBookings {
		if b.PaymentStatus != model.PaymentConfirmed {
			continue
		}
		// Overlap: b.checkIn < checkOut && b.checkOut > checkIn
		if b.CheckInDate.Before(checkOut) && b.CheckOutDate.After(checkIn) {
			relevantBookings = append(relevantBookings, b)
		}
	}

	result := make(map[string]model.RoomAvailability)

	for _, room := range rooms {
		// For each day in [checkIn, checkOut), find worst-case occupied beds
		minAvailableBeds := room.Capacity
		var genderRestriction model.Gender

		for d := checkIn; d.Before(checkOut); d = d.AddDate(0, 0, 1) {
			occupiedBeds := 0
			var dayGender model.Gender

			for _, b := range relevantBookings {
				if b.RoomID != room.ID {
					continue
				}
				// d is within [b.checkIn, b.checkOut)
				if !d.Before(b.CheckInDate) && d.Before(b.CheckOutDate) {
					occupiedBeds += b.BedCount
					if dayGender == "" && b.Gender != "" {
						dayGender = b.Gender
					}
				}
			}

			availOnDay := room.Capacity - occupiedBeds
			if availOnDay < minAvailableBeds {
				minAvailableBeds = availOnDay
				if dayGender != "" {
					genderRestriction = dayGender
				}
			}
		}

		if minAvailableBeds < 0 {
			minAvailableBeds = 0
		}

		var finalGender model.Gender
		if room.AllowSingleBed && minAvailableBeds > 0 && minAvailableBeds < room.Capacity {
			finalGender = genderRestriction
		}

		result[room.ID] = model.RoomAvailability{
			Room:              room,
			Available:         minAvailableBeds > 0,
			AvailableBeds:     minAvailableBeds,
			GenderRestriction: finalGender,
		}
	}

	return result, nil
}
