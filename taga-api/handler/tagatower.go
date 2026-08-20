package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"

	"taga-api/config"
	"taga-api/model"
	"taga-api/service"
	"taga-api/service/audit"

	"go.uber.org/zap"
)

/* ---------------------------
   ROOMS
--------------------------- */

// GetRooms godoc
// @Summary Get all rooms
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Room
// @Router /api/towers/rooms [get]
func GetRooms(c *gin.Context) {
	rooms, err := service.ReadRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load rooms"})
		return
	}

	if rooms == nil {
		rooms = []model.Room{}
	}

	c.JSON(http.StatusOK, rooms)
}

/* ---------------------------
   BOOKINGS
--------------------------- */

// CreateBooking godoc
// @Summary Create a new booking
// @Tags TAGA Towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bookerId query string false "Booker TAGA ID (optional, extracted from token if omitted)"
// @Param request body model.CreateBookingRequest true "Booking Data"
// @Success 201 {object} model.BookingResponse
// @Router /api/towers/bookings [post]
func CreateBooking(c *gin.Context) {
	var req model.CreateBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookerID := c.GetString("bookerID")
	if bookerID == "" {
		bookerID = c.Query("bookerId")
	}

	bookerName := c.GetString("bookerName")
	if bookerName == "" && bookerID != "" {
		bookerName = getMemberNameByTagaID(bookerID)
	}
	if bookerName == "" {
		bookerName = "User"
	}

	booking, err := service.CreateBooking(req, bookerName, bookerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Audit room booking request
	_ = audit.Log(c, bookerID, bookerName,
		audit.ActionBookingCreated, audit.ModuleBooking,
		"booking", booking.ID,
		fmt.Sprintf("Member %s (ID: %s) requested room booking %s for room '%s' (%s to %s)",
			bookerName, bookerID, booking.ID, booking.RoomName, booking.CheckInDate, booking.CheckOutDate),
		nil, booking)

	c.JSON(http.StatusCreated, booking)
}

// GetBookings godoc
// @Summary Get user bookings
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Param bookerId query string true "Booker ID"
// @Success 200 {array} model.BookingResponse
// @Router /api/towers/bookings [get]
func GetBookings(c *gin.Context) {
	bookerID := c.Query("bookerId")

	bookings, err := service.GetUserBookings(bookerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bookings == nil {
		bookings = []model.BookingResponse{}
	}

	c.JSON(http.StatusOK, bookings)
}

// GetPastBookings godoc
// @Summary Get past (archived) user bookings
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Param bookerId query string true "Booker ID"
// @Param year query string false "Year (YYYY)"
// @Param month query string false "Month (MM)"
// @Success 200 {array} model.BookingResponse
// @Router /api/towers/bookings/past [get]
func GetPastBookings(c *gin.Context) {
	bookerID := c.Query("bookerId")
	if bookerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookerId is required"})
		return
	}
	
	year := c.Query("year")
	month := c.Query("month")

	bookings, err := service.GetPastUserBookings(bookerID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bookings == nil {
		bookings = []model.BookingResponse{}
	}

	c.JSON(http.StatusOK, bookings)
}

// GetAllBookingsAdmin godoc
// @Summary Get all bookings (admin only — for occupancy schedule)
// @Description Returns all bookings across all users. Used by the admin occupancy schedule tab.
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.BookingResponse
// @Router /api/towers/admin/bookings [get]
func GetAllBookingsAdmin(c *gin.Context) {
	bookings, err := service.GetAllBookingsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bookings == nil {
		bookings = []model.BookingResponse{}
	}

	c.JSON(http.StatusOK, bookings)
}

// DeleteBooking godoc
// @Summary Cancel a booking
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Router /api/towers/bookings/{id} [delete]
func DeleteBooking(c *gin.Context) {
	bookingID := c.Param("id")

	booking, err := service.GetBookingByID(bookingID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Booking not found"})
		return
	}

	err = service.CancelBooking(bookingID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Audit room booking cancellation
	_ = audit.Log(c, booking.BookerID, booking.BookerName,
		audit.ActionBookingCancelled, audit.ModuleBooking,
		"booking", booking.ID,
		fmt.Sprintf("Member %s (ID: %s) cancelled room booking %s", booking.BookerName, booking.BookerID, booking.ID),
		booking, nil)

	c.JSON(200, gin.H{"message": "Booking cancelled"})
}

/* ---------------------------
   AVAILABILITY
--------------------------- */

// CheckAvailability godoc
// @Summary Check room availability
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Param roomId query string true "Room ID"
// @Param date query string true "Date"
// @Success 200 {object} model.RoomAvailability
// @Router /api/towers/availability [get]
func CheckAvailability(c *gin.Context) {
	roomID := c.Query("roomId")
	dateStr := c.Query("date")

	if roomID == "" || dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId and date required"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	room, err := service.GetRoomByID(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room not found"})
		return
	}

	allBookings, err := service.ReadAllBookings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filtered := []model.Booking{}
	for _, b := range allBookings {
		if b.RoomID != room.ID {
			continue
		}
		if b.PaymentStatus != model.PaymentConfirmed {
			continue
		}
		if date.Before(b.CheckOutDate) && date.After(b.CheckInDate.Add(-time.Second)) {
			filtered = append(filtered, b)
		}
	}

	available, availableBeds, genderRestriction, _ :=
		service.CheckRoomAvailability(room, date, filtered)

	c.JSON(http.StatusOK, model.RoomAvailability{
		Room:              *room,
		Available:         available,
		AvailableBeds:     availableBeds,
		GenderRestriction: model.Gender(genderRestriction),
	})
}

// CheckAvailabilityRange godoc
// @Summary Check availability for all rooms over a date range (bulk)
// @Description One-shot endpoint: returns availability of every room for the entire
// check-in to check-out range. Replaces N×(checkOut-checkIn) individual /availability calls.
// @Tags TAGA Towers
// @Produce json
// @Security BearerAuth
// @Param checkIn query string true "Check-in date (YYYY-MM-DD)"
// @Param checkOut query string true "Check-out date (YYYY-MM-DD)"
// @Success 200 {object} map[string]model.RoomAvailability
// @Router /api/towers/availability-range [get]
func CheckAvailabilityRange(c *gin.Context) {
	checkInStr := c.Query("checkIn")
	checkOutStr := c.Query("checkOut")

	if checkInStr == "" || checkOutStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checkIn and checkOut are required"})
		return
	}

	checkIn, err := time.Parse("2006-01-02", checkInStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checkIn date format, use YYYY-MM-DD"})
		return
	}

	checkOut, err := time.Parse("2006-01-02", checkOutStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checkOut date format, use YYYY-MM-DD"})
		return
	}

	if !checkOut.After(checkIn) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checkOut must be after checkIn"})
		return
	}

	days := int(checkOut.Sub(checkIn).Hours() / 24)
	if days > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum booking range is 10 days"})
		return
	}

	availMap, err := service.CheckAllRoomsAvailabilityRange(checkIn, checkOut)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check availability"})
		return
	}

	c.JSON(http.StatusOK, availMap)
}

/* ---------------------------
   PAYMENT (RAZORPAY)
--------------------------- */

// CreateOrderResponse wraps Razorpay order and public key
type CreateOrderResponse struct {
	Key   string                 `json:"key"`
	Order map[string]interface{} `json:"order"`
}

// CreateOrder godoc
// @Summary Create Razorpay order
// @Description Create Razorpay order for payment
// @Tags TAGA Towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Amount in paise" example({"amount":10000})
// @Success 200 {object} CreateOrderResponse
// @Router /api/towers/create-order [post]
func CreateOrder(c *gin.Context) {
	razorpayKey := os.Getenv("RAZORPAY_KEY")
	razorpaySecret := os.Getenv("RAZORPAY_SECRET")

	if config.Config.DisablePayment {
		if razorpayKey == "" {
			razorpayKey = "mock_key"
		}
		if razorpaySecret == "" {
			razorpaySecret = "mock_secret"
		}
	} else if razorpayKey == "" || razorpaySecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Razorpay credentials not configured"})
		return
	}

	var req struct {
		Amount int                    `json:"amount"`
		Notes  map[string]interface{} `json:"notes,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	// Get member details for notes from context (set by middleware)
	bookerName := c.GetString("bookerName")
	if bookerName == "" {
		// Try to get from logged-in user
		if memberEmail, exists := c.Get("email"); exists {
			if name := getMemberNameByEmail(memberEmail.(string)); name != "" {
				bookerName = name
			}
		}
	}

	bookerID := c.GetString("bookerID")
	if bookerID == "" {
		if memberID, exists := c.Get("member_id"); exists {
			bookerID = memberID.(string)
		}
	}

	// Get booker email
	bookerEmail := ""
	if email, exists := c.Get("email"); exists {
		bookerEmail = email.(string)
	}

	// Build notes with room booking details
	orderNotes := map[string]interface{}{
		"payment_type": "room_booking",
	}

	// Merge any additional notes from frontend (contains room details, guest details, etc.)
	for k, v := range req.Notes {
		orderNotes[k] = v
	}

	// Add booker details to notes if not already present
	if _, exists := orderNotes["booker_name"]; !exists && bookerName != "" {
		orderNotes["booker_name"] = bookerName
	}
	if _, exists := orderNotes["booker_taga_id"]; !exists && bookerID != "" {
		orderNotes["booker_taga_id"] = bookerID
	}
	if _, exists := orderNotes["booker_email"]; !exists && bookerEmail != "" {
		orderNotes["booker_email"] = bookerEmail
	}

	data := map[string]interface{}{
		"amount":   req.Amount,
		"currency": "INR",
		"receipt":  "taga_booking_" + time.Now().Format("20060102150405"),
		"notes":    orderNotes,
	}

	var order map[string]interface{}
	var err error

	if config.Config.DisablePayment {
		mockOrderID := "mock_order_" + uuid.New().String()[:18]
		order = map[string]interface{}{
			"id":       mockOrderID,
			"amount":   req.Amount,
			"currency": "INR",
			"receipt":  data["receipt"].(string),
			"status":   "created",
		}
		config.Logger.Info("Bypassed Razorpay order creation, generated mock order", zap.String("order_id", mockOrderID))
	} else {
		client := razorpay.NewClient(razorpayKey, razorpaySecret)
		order, err = client.Order.Create(data, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, CreateOrderResponse{
		Key:   razorpayKey,
		Order: order,
	})
}

// ConfirmPayment godoc
// @Summary Confirm payment
// @Tags TAGA Towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Param request body object true "UPI ID"
// @Success 200 {object} map[string]string
// @Router /api/towers/bookings/{id}/confirm-payment [post]
func ConfirmPayment(c *gin.Context) {
	bookingID := c.Param("id")

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	upiID := req["upiId"]
	if upiID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upiId required"})
		return
	}

	err := service.ConfirmPayment(bookingID, upiID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Audit room booking payment confirmation via UPI
	booking, err := service.GetBookingByID(bookingID)
	if err == nil {
		_ = audit.Log(c, booking.BookerID, booking.BookerName,
			audit.ActionPaymentConfirmed, audit.ModulePayment,
			"booking", bookingID,
			fmt.Sprintf("Payment confirmed for room booking %s via UPI ID: %s (Amount: %d)",
				bookingID, upiID, booking.AdvanceAmount),
			nil, map[string]interface{}{"upi_id": upiID, "booking_id": bookingID})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed"})
}

func getMemberEmailByTagaID(tagaID string) string {
	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		return ""
	}
	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return ""
	}
	for _, m := range members {
		if id, ok := m["tagaId"].(string); ok && id == tagaID {
			if email, ok := m["emailId"].(string); ok {
				return email
			}
		}
	}
	return ""
}

func getMemberNameByTagaID(tagaID string) string {
	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		return ""
	}
	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return ""
	}
	for _, m := range members {
		if id, ok := m["tagaId"].(string); ok && id == tagaID {
			if name, ok := m["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}

func buildRoomBookingEmailBody(data AdminRoomBookingData) string {
	amountInRupees := float64(data.Amount) / 100

	var body strings.Builder

	// Parse guest details if present
	var guestList []map[string]interface{}
	guestTableHTML := ""
	if data.GuestDetails != "" && data.GuestDetails != "[]" {
		if err := json.Unmarshal([]byte(data.GuestDetails), &guestList); err == nil && len(guestList) > 0 {
			guestTableHTML = `
            <div style="margin-top: 24px;">
                <h3 style="color: #065f46; margin-bottom: 12px; font-size: 16px; border-bottom: 1px solid #e5e7eb; padding-bottom: 8px;">👥 Guest Details</h3>
                <table style="width: 100%; border-collapse: separate; border-spacing: 0; border: 1px solid #e5e7eb; border-radius: 6px; overflow: hidden;">
                    <thead>
                        <tr style="background: #f3f4f6;">
                            <th style="padding: 10px; text-align: left; border-bottom: 1px solid #e5e7eb; font-size: 13px; color: #4b5563;">#</th>
                            <th style="padding: 10px; text-align: left; border-bottom: 1px solid #e5e7eb; font-size: 13px; color: #4b5563;">Name</th>
                            <th style="padding: 10px; text-align: left; border-bottom: 1px solid #e5e7eb; font-size: 13px; color: #4b5563;">Age</th>
                            <th style="padding: 10px; text-align: left; border-bottom: 1px solid #e5e7eb; font-size: 13px; color: #4b5563;">Contact</th>
                        </tr>
                    </thead>
                    <tbody>`
			for i, guest := range guestList {
				name, _ := guest["name"].(string)
				age, _ := guest["age"].(float64)
				contact, _ := guest["contact"].(string)
				
				borderBottom := ""
				if i < len(guestList)-1 {
					borderBottom = "border-bottom: 1px solid #e5e7eb;"
				}

				guestTableHTML += fmt.Sprintf(`
                        <tr>
                            <td style="padding: 10px; %s font-size: 14px;">%d</td>
                            <td style="padding: 10px; %s font-size: 14px;">%s</td>
                            <td style="padding: 10px; %s font-size: 14px;">%.0f</td>
                            <td style="padding: 10px; %s font-size: 14px;">%s</td>
                        </tr>`, borderBottom, i+1, borderBottom, name, borderBottom, age, borderBottom, contact)
			}
			guestTableHTML += `</tbody>
                </table>
            </div>`
		}
	}

	bookingForDisplay := "Self"
	if data.BookingFor == "guest" {
		bookingForDisplay = "Guest"
	}

	fmt.Fprintf(&body, `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TAGA Towers Booking Confirmation</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);">
        
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #065f46 0%%, #047857 100%%); color: white; padding: 30px 20px; text-align: center;">
            <div style="display: inline-block; padding: 6px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; letter-spacing: 0.05em; background: rgba(255, 255, 255, 0.2); margin-bottom: 16px;">
                CONFIRMED BOOKING
            </div>
            <h2 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 700;">TAGA Towers Reservation</h2>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Booking ID: <strong>%s</strong></p>
        </div>

        <!-- Content -->
        <div style="padding: 30px 24px;">
            
            <!-- Payment Summary Card -->
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 24px;">
                <h3 style="margin: 0 0 12px 0; color: #0f172a; font-size: 15px; display: flex; align-items: center;">
                    <span style="margin-right: 8px;">💳</span> Payment Summary
                </h3>
                <table style="width: 100%%; border-collapse: collapse;">
                    <tr><td style="padding: 6px 0; color: #64748b; font-size: 14px; width: 40%%;">Advance Amount</td><td style="padding: 6px 0; font-weight: 600; color: #0f172a; text-align: right;">₹ %.2f</td></tr>
                    <tr><td style="padding: 6px 0; color: #64748b; font-size: 14px;">Payment ID</td><td style="padding: 6px 0; color: #334155; font-size: 13px; text-align: right; word-break: break-all;">%s</td></tr>
                    <tr><td style="padding: 6px 0; color: #64748b; font-size: 14px;">Order ID</td><td style="padding: 6px 0; color: #334155; font-size: 13px; text-align: right; word-break: break-all;">%s</td></tr>
                </table>
            </div>

            <!-- Room Details -->
            <h3 style="color: #065f46; margin: 0 0 16px 0; font-size: 16px; border-bottom: 1px solid #e5e7eb; padding-bottom: 8px;">🏨 Stay Details</h3>
            <table style="width: 100%%; border-collapse: collapse; margin-bottom: 24px;">
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b; width: 40%%;">Room Category</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%s</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">Beds Booked</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%d</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">Check-in</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 600; color: #065f46;">%s</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">Check-out</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 600; color: #991b1b;">%s</td></tr>
            </table>

            <!-- Guest Details -->
            <h3 style="color: #065f46; margin: 0 0 16px 0; font-size: 16px; border-bottom: 1px solid #e5e7eb; padding-bottom: 8px;">👤 Primary Guest / Booker</h3>
            <table style="width: 100%%; border-collapse: collapse;">
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b; width: 40%%;">Name</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%s</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">TAGA ID</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%s</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">Email</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%s</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">Phone</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%s</td></tr>
                <tr><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; color: #64748b;">Booking Type</td><td style="padding: 10px 0; border-bottom: 1px solid #f1f5f9; font-weight: 500;">%s</td></tr>
            </table>

            %s

        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 20px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 8px 0; font-size: 13px; color: #64748b;">This is an automated message from the TAGA Towers Booking System.</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">&copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
		data.BookingID, amountInRupees, data.PaymentID, data.OrderID,
		data.RoomName, data.BedCount, data.CheckInDate, data.CheckOutDate,
		data.BookerName, data.BookerTagaID, data.CustomerEmail, data.CustomerPhone, bookingForDisplay,
		guestTableHTML,
	)

	return body.String()
}

// VerifyPayment godoc
// @Summary Verify Razorpay payment
// @Tags TAGA Towers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Payment verification data"
// @Success 200 {object} map[string]string
// @Router /api/towers/verify-payment [post]
func VerifyPayment(c *gin.Context) {
	var req struct {
		BookingID string `json:"bookingId"`
		OrderID   string `json:"razorpay_order_id"`
		PaymentID string `json:"razorpay_payment_id"`
		Signature string `json:"razorpay_signature"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	isMock := config.Config.DisablePayment || strings.HasPrefix(req.OrderID, "mock_order_") || req.Signature == "mock_signature"
	if isMock {
		config.Logger.Info("Bypassing payment signature verification for mock payment", zap.String("order_id", req.OrderID))
	} else {
		razorpaySecret := os.Getenv("RAZORPAY_SECRET")
		data := req.OrderID + "|" + req.PaymentID
		h := hmac.New(sha256.New, []byte(razorpaySecret))
		h.Write([]byte(data))
		expected := hex.EncodeToString(h.Sum(nil))

		if expected != req.Signature {
			c.JSON(400, gin.H{"error": "Payment verification failed"})
			return
		}
	}

	err := service.ConfirmPaymentWithDetails(
		req.BookingID,
		req.OrderID,
		req.PaymentID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Audit room booking payment confirmation via Razorpay
	booking, err := service.GetBookingByID(req.BookingID)
	if err == nil {
		_ = audit.Log(c, booking.BookerID, booking.BookerName,
			audit.ActionPaymentConfirmed, audit.ModulePayment,
			"booking", req.BookingID,
			fmt.Sprintf("Razorpay payment confirmed for room booking %s (Order: %s, Payment: %s)",
				req.BookingID, req.OrderID, req.PaymentID),
			nil, map[string]interface{}{"order_id": req.OrderID, "payment_id": req.PaymentID, "booking_id": req.BookingID})
	}

	// ========== SEND ADMIN EMAIL FOR ROOM BOOKING ==========
	if err == nil && booking != nil {
		// Get room details
		room, _ := service.GetRoomByID(booking.RoomID)
		roomName := ""
		if room != nil {
			roomName = room.Name
		}

		// Convert BookingFor to string
		bookingForStr := string(booking.BookingFor)
		if bookingForStr == "" {
			bookingForStr = "self"
		}

		// Get Member to extract their email
		var customerEmail string
		customerEmail = getMemberEmailByTagaID(booking.BookerID)

		bookerName := booking.BookerName
		if (bookerName == "" || bookerName == "User") && booking.BookerID != "" {
			if realName := getMemberNameByTagaID(booking.BookerID); realName != "" {
				bookerName = realName
			}
		}

		// Prepare email data
		emailData := AdminRoomBookingData{
			BookingID:     req.BookingID,
			PaymentID:     req.PaymentID,
			OrderID:       req.OrderID,
			Amount:        booking.AdvanceAmount,
			CustomerEmail: customerEmail, 
			CustomerPhone: booking.BookerPhone,
			RoomName:      roomName,
			RoomNumber:    "", // Will be populated from frontend notes
			BedCount:      booking.BedCount,
			CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
			CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
			BookerName:    bookerName,
			BookerTagaID:  booking.BookerID,
			BookerPhone:   booking.BookerPhone,
			BookingFor:    bookingForStr,
			GuestDetails:  "",
			PaymentType:   "room_booking",
		}

		// Convert guest details to JSON if available
		if len(booking.GuestDetails) > 0 {
			guestJSON, _ := json.Marshal(booking.GuestDetails)
			emailData.GuestDetails = string(guestJSON)
		}

		// Build email subject
		amountInRupees := float64(booking.AdvanceAmount) / 100
		subject := fmt.Sprintf("🏨 TAGA Room Booking - %s - ₹%.2f", roomName, amountInRupees)

		// Get email body
		emailBody := buildRoomBookingEmailBody(emailData)

		// Send emails with retry mechanism
		paymentID := req.PaymentID
		if !hasEmailBeenSent(paymentID) {
			// 1. Send to Admin
			adminEmail := config.GetConfig().AdminEmail
			if adminEmail != "" {
				go sendEmailWithRetry(adminEmail, subject, emailBody, paymentID, "room_booking", 2)
			} else {
				config.Logger.Warn("Admin email not configured, skipping admin notification")
			}
			
			// 2. Send to Customer
			if customerEmail != "" {
				go sendEmailWithRetry(customerEmail, subject, emailBody, paymentID, "room_booking", 2)
			} else {
				config.Logger.Warn("Customer email not found, skipping customer notification")
			}
		} else {
			config.Logger.Info("Email already sent for this payment, skipping duplicate",
				zap.String("payment_id", paymentID))
		}
	} else {
		config.Logger.Error("Failed to fetch booking details for email notification",
			zap.String("booking_id", req.BookingID))
	}
	// ========== END OF ADMIN EMAIL ==========

	c.JSON(200, gin.H{
		"message": "Payment verified & booking confirmed",
	})
}


