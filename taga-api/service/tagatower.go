package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"taga-api/config"
	"taga-api/model"
)

// bookingsLock protects all concurrent read/write operations to bookings.json
var bookingsLock sync.RWMutex

// PendingHoldDuration defines how long an unconfirmed/pending booking holds bed capacity (10 minutes)
const PendingHoldDuration = 10 * time.Minute

// isBookingActive determines if a booking is occupying room/bed inventory.
// A booking occupies inventory if:
// 1. It is PaymentConfirmed, OR
// 2. It is PaymentPending and created within PendingHoldDuration (in-flight checkout hold)
func isBookingActive(b model.Booking, now time.Time) bool {
	if b.PaymentStatus == model.PaymentConfirmed {
		return true
	}
	if b.PaymentStatus == model.PaymentPending {
		// Active hold if within 10 minutes of creation
		return now.Sub(b.CreatedAt) < PendingHoldDuration
	}
	return false
}

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

func BookingsArchiveDirHelper() string {
	path := config.Config.Data.Tower.BookingsArchiveDir
	if path == "" {
		return getFilePath("data", "towers", "archive")
	}
	return path
}

func ReadAllBookings() ([]model.Booking, error) {
	bookingsLock.RLock()
	defer bookingsLock.RUnlock()

	return readAllBookingsUnsafe()
}

func readAllBookingsUnsafe() ([]model.Booking, error) {
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
	bookingsLock.Lock()
	defer bookingsLock.Unlock()

	return saveAllBookingsUnsafe(bookings)
}

func saveAllBookingsUnsafe(bookings []model.Booking) error {
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



func isBookingMixed(b model.Booking) bool {
	if b.Gender == model.GenderMixed {
		return true
	}
	hasMale := false
	hasFemale := false
	for _, g := range b.GuestDetails {
		if g.Gender == model.GenderMale {
			hasMale = true
		} else if g.Gender == model.GenderFemale {
			hasFemale = true
		}
	}
	return hasMale && hasFemale
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

	// Determine advance amount: 1 per bed (configurable base, future: 100/200)
	advanceRatePerBed := 1
	if req.BookingFor == model.BookingForGuest {
		advanceRatePerBed = 1
	}
	effectiveBeds := req.BedCount
	if !room.AllowSingleBed {
		effectiveBeds = room.Capacity
	}
	if req.BookingFor == model.BookingForSelf {
		effectiveBeds = 1
	}
	advanceAmount := advanceRatePerBed * effectiveBeds

	// ── VALIDATION 0: Guest gender processing ────────────────────────────────
	if req.BookingFor == model.BookingForGuest && len(req.GuestDetails) > 0 {
		hasMale := false
		hasFemale := false
		for i, g := range req.GuestDetails {
			if g.Gender != model.GenderMale && g.Gender != model.GenderFemale {
				return nil, fmt.Errorf("please specify the gender for Guest %d (%s)", i+1, g.Name)
			}
			if g.Gender == model.GenderMale {
				hasMale = true
			} else if g.Gender == model.GenderFemale {
				hasFemale = true
			}
		}

		var derivedGender model.Gender
		if hasMale && hasFemale {
			derivedGender = model.GenderMixed
		} else if hasMale {
			derivedGender = model.GenderMale
		} else if hasFemale {
			derivedGender = model.GenderFemale
		}

		// Dormitories are strictly single-gender
		if room.Type == "gents-dorm" && derivedGender != model.GenderMale {
			return nil, fmt.Errorf("this room is strictly for male guests only")
		}
		if room.Type == "ladies-dorm" && derivedGender != model.GenderFemale {
			return nil, fmt.Errorf("this room is strictly for female guests only")
		}

		req.Gender = derivedGender
	}

	now := time.Now()

	// Create booking object (not saved yet)
	booking := model.Booking{
		ID:            fmt.Sprintf("BK%d", now.UnixNano()),
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
		CreatedAt:     now,
	}

	// 🔒 ATOMIC TRANSACTION: Acquire exclusive lock for reading, validating, and reserving bed/room
	bookingsLock.Lock()
	defer bookingsLock.Unlock()

	allBookingsData, err := readAllBookingsUnsafe()
	if err != nil {
		return nil, err
	}

	// ── VALIDATION 1: Bed capacity overlap check (Confirmed + Active Pending holds) ──
	for d := checkInDate; d.Before(checkOutDate); d = d.AddDate(0, 0, 1) {
		occupiedBeds := 0
		hasActiveMixedBooking := false

		for _, b := range allBookingsData {
			if b.RoomID != req.RoomID {
				continue
			}
			// Both confirmed bookings and active pending holds block capacity
			if !isBookingActive(b, now) {
				continue
			}
			// Check date overlap: existing booking overlaps day d
			if d.Before(b.CheckOutDate) && d.After(b.CheckInDate.Add(-time.Second)) {
				occupiedBeds += b.BedCount
				if room.ID == "apex-1" && (b.Gender == model.GenderMixed || isBookingMixed(b)) {
					hasActiveMixedBooking = true
				}
			}
		}

		// If Apex Suite has an active mixed couple booking, the 3rd bed is completely blocked
		if room.ID == "apex-1" && hasActiveMixedBooking {
			occupiedBeds = room.Capacity
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
			// Skip non-matching rooms and inactive bookings
			if existingBooking.RoomID != req.RoomID {
				continue
			}
			if !isBookingActive(existingBooking, now) {
				continue
			}

			// Check if the existing booking overlaps with our requested dates
			if checkInDate.Before(existingBooking.CheckOutDate) &&
				checkOutDate.After(existingBooking.CheckInDate) {

				// If existing booking is mixed in Apex, the room is fully reserved
				if room.ID == "apex-1" && (existingBooking.Gender == model.GenderMixed || isBookingMixed(existingBooking)) {
					return nil, fmt.Errorf("Apex Suite is fully booked for these dates (couple reservation)")
				}

				// If existing booking is single-gender and new booking has a different gender or is mixed
				if existingBooking.Gender != "" && existingBooking.Gender != model.GenderMixed && req.Gender != "" {
					if req.Gender != existingBooking.Gender {
						return nil, fmt.Errorf("this room is partially occupied by %s guests — only %s guests can book the remaining beds",
							existingBooking.Gender, existingBooking.Gender)
					}
				}
			}
		}
	}

	// 🔥 Append and save safely under lock
	allBookingsData = append(allBookingsData, booking)

	if err := saveAllBookingsUnsafe(allBookingsData); err != nil {
		return nil, err
	}

	// Return response (with computed status for consistency)
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
	now := time.Now()
	filtered := []model.Booking{}
	for _, b := range bookings {

		if b.RoomID != room.ID {
			continue
		}

		if !isBookingActive(b, now) {
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
	hasMixedBooking := false

	for i, b := range filtered {
		occupiedBeds += b.BedCount

		if b.Gender == model.GenderMixed || isBookingMixed(b) {
			hasMixedBooking = true
		} else if i == 0 && b.Gender != "" {
			roomGender = b.Gender
		}
	}

	// If Apex Suite has a mixed couple booking, block the 3rd bed completely
	if room.ID == "apex-1" && hasMixedBooking {
		occupiedBeds = room.Capacity
	}

	availableBeds := room.Capacity - occupiedBeds
	if availableBeds < 0 {
		availableBeds = 0
	}

	// ✅ Gender restriction (AC rooms with partial booking)
	var genderRestriction string
	if room.AllowSingleBed && occupiedBeds > 0 && availableBeds > 0 && !hasMixedBooking {
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
	bookingsLock.Lock()
	defer bookingsLock.Unlock()

	allBookings, err := readAllBookingsUnsafe()
	if err != nil {
		return err
	}

	for i := range allBookings {
		if allBookings[i].ID == bookingID {
			allBookings[i].PaymentStatus = model.PaymentCancelled
			return saveAllBookingsUnsafe(allBookings)
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
	bookingsLock.Lock()
	defer bookingsLock.Unlock()

	allBookings, err := readAllBookingsUnsafe()
	if err != nil {
		return err
	}

	for i := range allBookings {
		if allBookings[i].ID == bookingID {
			allBookings[i].PaymentStatus = model.PaymentConfirmed
			allBookings[i].UpiID = upiID

			return saveAllBookingsUnsafe(allBookings)
		}
	}

	return fmt.Errorf("booking not found")
}

func ConfirmPaymentWithDetails(bookingID, orderID, paymentID string) error {
	bookingsLock.Lock()
	defer bookingsLock.Unlock()

	allBookings, err := readAllBookingsUnsafe()
	if err != nil {
		return err
	}

	for i := range allBookings {
		if allBookings[i].ID == bookingID {
			allBookings[i].PaymentStatus = model.PaymentConfirmed
			allBookings[i].RazorpayOrderID = orderID
			allBookings[i].RazorpayPaymentID = paymentID

			return saveAllBookingsUnsafe(allBookings)
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
	  Daily Archive — Archive completed bookings
---------------------------
*/

// StartBookingArchiveScheduler runs a daily goroutine that archives bookings
// whose checkOutDate is before today.
func StartBookingArchiveScheduler() {
	go func() {
		log.Println("✅ Booking archive scheduler started (runs daily)")
		for {
			RunBookingArchive()
			// Sleep until next midnight
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
		}
	}()
}

// RunBookingArchive archives all bookings whose checkOutDate is before today.
func RunBookingArchive() {
	bookingsLock.Lock()
	defer bookingsLock.Unlock()

	allBookings, err := readAllBookingsUnsafe()
	if err != nil {
		log.Printf("❌ Booking archive: failed to read bookings: %v", err)
		return
	}

	now := time.Now()
	// Start of today. Any checkout date before this is strictly in the past.
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	kept := []model.Booking{}
	archivedCount := 0
	expiredPendingCount := 0

	for _, b := range allBookings {
		isPastDate := b.CheckOutDate.Before(todayStart)
		isCancelled := b.PaymentStatus == model.PaymentCancelled // Also archive cancelled bookings
		isExpiredPending := b.PaymentStatus == model.PaymentPending && now.Sub(b.CreatedAt) > 24*time.Hour

		if isExpiredPending {
			expiredPendingCount++
			log.Printf("🧹 Booking archive: cleaned up abandoned pending booking %s (created %s)", b.ID, b.CreatedAt.Format("2006-01-02 15:04"))
			continue
		}

		if isPastDate || isCancelled {
			if err := archiveBooking(b); err != nil {
				log.Printf("❌ Booking archive: failed to archive booking %s: %v", b.ID, err)
				kept = append(kept, b) // Keep it in active bookings so we can try again tomorrow
			} else {
				archivedCount++
				log.Printf("📦  Archive: archived completed/cancelled booking %s (checkOut: %s)", b.ID, b.CheckOutDate.Format("2006-01-02"))
			}
		} else {
			kept = append(kept, b)
		}
	}

	if archivedCount == 0 && expiredPendingCount == 0 {
		log.Println("✅ Booking archive: no completed/expired bookings to archive")
		return
	}

	if err := saveAllBookingsUnsafe(kept); err != nil {
		log.Printf("❌ Booking archive: failed to save updated active bookings: %v", err)
		return
	}

	log.Printf("✅ Booking archive: successfully archived %d booking(s), cleaned %d expired pending", archivedCount, expiredPendingCount)
}

func archiveBooking(b model.Booking) error {
	year := b.CheckOutDate.Format("2006")
	month := b.CheckOutDate.Format("2006-01")

	yearDir := filepath.Join(BookingsArchiveDirHelper(), year)
	if err := os.MkdirAll(yearDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(yearDir, month+".json")

	var archivedBookings []model.Booking
	if fileData, err := os.ReadFile(filePath); err == nil && len(fileData) > 0 {
		if err := json.Unmarshal(fileData, &archivedBookings); err != nil {
			// If file is corrupted, backup and start fresh? For now return err.
			return err
		}
	}

	// Prevent duplicates
	for _, existing := range archivedBookings {
		if existing.ID == b.ID {
			return nil // Already archived
		}
	}

	archivedBookings = append(archivedBookings, b)
	
	newData, err := json.MarshalIndent(archivedBookings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, newData, 0644)
}

/*
	---------------------------
	  Get Past Bookings
---------------------------
*/

// GetPastUserBookings retrieves archived bookings for a given member, optionally filtered by year and month.
func GetPastUserBookings(bookerID string, year string, month string) ([]model.BookingResponse, error) {
	if year != "" {
		if len(year) != 4 {
			return nil, fmt.Errorf("invalid year format")
		}
		for _, c := range year {
			if c < '0' || c > '9' {
				return nil, fmt.Errorf("invalid year format")
			}
		}
	}
	if month != "" {
		if len(month) != 2 {
			return nil, fmt.Errorf("invalid month format")
		}
		for _, c := range month {
			if c < '0' || c > '9' {
				return nil, fmt.Errorf("invalid month format")
			}
		}
	}

	var pastBookings []model.BookingResponse
	
	rooms, err := ReadRooms()
	if err != nil {
		return nil, err
	}
	roomMap := make(map[string]*model.Room)
	for i, r := range rooms {
		roomMap[r.ID] = &rooms[i]
	}
	baseDir := BookingsArchiveDirHelper()
	
	var filesToRead []string
	
	if year != "" && month != "" {
		// Specific month
		filesToRead = append(filesToRead, filepath.Join(baseDir, year, fmt.Sprintf("%s-%s.json", year, month)))
	} else if year != "" {
		// All months in a year
		dir := filepath.Join(baseDir, year)
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					filesToRead = append(filesToRead, filepath.Join(dir, e.Name()))
				}
			}
		}
	} else {
		// All years and months
		entries, err := os.ReadDir(baseDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() {
					yearDir := filepath.Join(baseDir, e.Name())
					monthEntries, err := os.ReadDir(yearDir)
					if err == nil {
						for _, me := range monthEntries {
							if !me.IsDir() && strings.HasSuffix(me.Name(), ".json") {
								filesToRead = append(filesToRead, filepath.Join(yearDir, me.Name()))
							}
						}
					}
				}
			}
		}
	}
	
	now := time.Now()
	
	for _, file := range filesToRead {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		var archivedBookings []model.Booking
		if err := json.Unmarshal(data, &archivedBookings); err != nil {
			continue
		}
		
		for _, b := range archivedBookings {
			if b.BookerID == bookerID {
				roomName := ""
				if room := roomMap[b.RoomID]; room != nil {
					roomName = room.Name
				}
				
				pastBookings = append(pastBookings, model.BookingResponse{
					ID:            b.ID,
					RoomID:        b.RoomID,
					RoomName:      roomName,
					CheckInDate:   b.CheckInDate.Format("2006-01-02"),
					CheckOutDate:  b.CheckOutDate.Format("2006-01-02"),
					BookerName:    b.BookerName,
					BookerID:      b.BookerID,
					BookingFor:    b.BookingFor,
					BedCount:      b.BedCount,
					Gender:        b.Gender,
					GuestDetails:  b.GuestDetails,
					PaymentStatus: b.PaymentStatus,
					AdvanceAmount: b.AdvanceAmount,
					BookingStatus: calculateBookingStatus(b, now),
				})
			}
		}
	}
	
	return pastBookings, nil
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

	now := time.Now()

	// Pre-filter: only confirmed bookings & active pending holds that overlap [checkIn, checkOut)
	var relevantBookings []model.Booking
	for _, b := range allBookings {
		if !isBookingActive(b, now) {
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
			hasMixedBooking := false

			for _, b := range relevantBookings {
				if b.RoomID != room.ID {
					continue
				}
				// d is within [b.checkIn, b.checkOut)
				if !d.Before(b.CheckInDate) && d.Before(b.CheckOutDate) {
					occupiedBeds += b.BedCount
					if b.Gender == model.GenderMixed || isBookingMixed(b) {
						hasMixedBooking = true
					} else if dayGender == "" && b.Gender != "" {
						dayGender = b.Gender
					}
				}
			}

			// If Apex Suite has a mixed couple booking on day d, the 3rd bed is completely blocked
			if room.ID == "apex-1" && hasMixedBooking {
				occupiedBeds = room.Capacity
			}

			availOnDay := room.Capacity - occupiedBeds
			if availOnDay < minAvailableBeds {
				minAvailableBeds = availOnDay
				if dayGender != "" && !hasMixedBooking {
					genderRestriction = dayGender
				} else if hasMixedBooking {
					genderRestriction = ""
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
