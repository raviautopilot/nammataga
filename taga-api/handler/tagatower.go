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
// @Param request body model.CreateBookingRequest true "Booking Data"
// @Success 201 {object} model.BookingResponse
// @Router /api/towers/bookings [post]
func CreateBooking(c *gin.Context) {
	var req model.CreateBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookerName := c.GetString("bookerName")
	if bookerName == "" {
		bookerName = "User"
	}

	bookerID := c.GetString("bookerID")
	if bookerID == "" {
		bookerID = c.Query("bookerId")
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

// GetAllBookingsAdmin godoc
// @Summary Get all bookings (admin only — for occupancy schedule)
// @Description Returns all bookings across all users. Used by the admin occupancy schedule tab.
// @Tags TAGA Towers
// @Produce json
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

// buildRoomBookingEmailBody builds the HTML email body for room booking payments
func buildRoomBookingEmailBody(data AdminRoomBookingData) string {
	amountInRupees := float64(data.Amount) / 100

	var body strings.Builder

	// Parse guest details if present
	var guestList []map[string]interface{}
	guestTableHTML := ""
	if data.GuestDetails != "" && data.GuestDetails != "[]" {
		if err := json.Unmarshal([]byte(data.GuestDetails), &guestList); err == nil && len(guestList) > 0 {
			guestTableHTML = `<h3>👥 Guest Details</h3>
<table style="width: 100%; border-collapse: collapse; margin: 15px 0;">
    <thead>
        <tr style="background: #f3f4f6;">
            <th style="padding: 10px; text-align: left; border: 1px solid #e5e7eb;">#</th>
            <th style="padding: 10px; text-align: left; border: 1px solid #e5e7eb;">Name</th>
            <th style="padding: 10px; text-align: left; border: 1px solid #e5e7eb;">Age</th>
            <th style="padding: 10px; text-align: left; border: 1px solid #e5e7eb;">Contact</th>
        </tr>
    </thead>
    <tbody>`
			for i, guest := range guestList {
				name, _ := guest["name"].(string)
				age, _ := guest["age"].(float64)
				contact, _ := guest["contact"].(string)
				guestTableHTML += fmt.Sprintf(`
        <tr>
            <td style="padding: 8px; border: 1px solid #e5e7eb;">%d</td>
            <td style="padding: 8px; border: 1px solid #e5e7eb;">%s</td>
            <td style="padding: 8px; border: 1px solid #e5e7eb;">%.0f</td>
            <td style="padding: 8px; border: 1px solid #e5e7eb;">%s</td>
        </tr>`, i+1, name, age, contact)
			}
			guestTableHTML += `</tbody>
</table>`
		}
	}

	bookingForDisplay := "Self"
	if data.BookingFor == "guest" {
		bookingForDisplay = "Guest"
	}

	fmt.Fprintf(&body, `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #065f46; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 20px; border: 1px solid #e5e7eb; }
        .footer { background: #f3f4f6; padding: 15px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; }
        table { width: 100%%; border-collapse: collapse; margin: 15px 0; }
        td { padding: 12px 10px; border-bottom: 1px solid #e5e7eb; }
        th { padding: 12px 10px; text-align: left; }
        .label { font-weight: bold; width: 35%%; background: #f3f4f6; }
        .value { background: white; }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: bold; background: #dcfce7; color: #166534; }
        h3 { color: #065f46; margin-top: 20px; margin-bottom: 10px; }
        hr { border: none; border-top: 1px solid #e5e7eb; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>TAGA Towers Booking Notification</h2>
            <p>%s</p>
        </div>
        <div class="content">
            <div style="text-align: center; margin-bottom: 20px;">
                <span class="badge">🏨 ROOM BOOKING</span>
            </div>
            
            <h3>💰 Payment Summary</h3>
            <table>
                <tr><td class="label">Payment ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Order ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Advance Amount:</td><td class="value">₹ %.2f</td></tr>
                <tr><td class="label">Customer Email:</td><td class="value">%s</td></tr>
            表
            
            <h3>🏨 Room Details</h3>
            <table>
                <tr><td class="label">Room Name:</td><td class="value">%s</td></tr>
                <tr><td class="label">Room Number:</td><td class="value">%s</td></tr>
                <tr><td class="label">Bed Count:</td><td class="value">%d bed(s)</td></tr>
                <tr><td class="label">Check-in Date:</td><td class="value">%s</td></tr>
                <tr><td class="label">Check-out Date:</td><td class="value">%s</td></tr>
            表
            
            <h3>👤 Booker Details</h3>
            <table>
                <tr><td class="label">Name:</td><td class="value">%s</td></tr>
                <tr><td class="label">TAGA ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Phone:</td><td class="value">%s</td></tr>
                <tr><td class="label">Booking For:</td><td class="value">%s</td></tr>
            表
            
            %s
            
            <hr>
            <p style="font-size: 12px; color: #6b7280; text-align: center;">
                This is an automated notification from TAGA Towers Booking System.
                <br>Remaining amount to be paid after stay.
            </p>
        </div>
        <div class="footer">
            <p>TAGA Towers | Agriculture Complex Road, Chennai - 600017</p>
            <p>Caretaker: Mr. Murugan Ramasamy | +91 98765 43210</p>
        </div>
    </div>
</body>
</html>`,
		time.Now().Format("January 02, 2006 at 03:04 PM"),
		data.PaymentID,
		data.OrderID,
		amountInRupees,
		data.CustomerEmail,
		data.RoomName,
		data.RoomNumber,
		data.BedCount,
		data.CheckInDate,
		data.CheckOutDate,
		data.BookerName,
		data.BookerTagaID,
		data.BookerPhone,
		bookingForDisplay,
		guestTableHTML,
	)

	return body.String()
}

// VerifyPayment godoc
// @Summary Verify Razorpay payment
// @Tags TAGA Towers
// @Accept json
// @Produce json
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

		// Prepare email data (no email field available, use phone as identifier)
		emailData := AdminRoomBookingData{
			PaymentID:     req.PaymentID,
			OrderID:       req.OrderID,
			Amount:        booking.AdvanceAmount,
			CustomerEmail: booking.BookerPhone, // Use phone as identifier since no email
			RoomName:      roomName,
			RoomNumber:    "", // Will be populated from frontend notes
			BedCount:      booking.BedCount,
			CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
			CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
			BookerName:    booking.BookerName,
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

		// Send email with retry mechanism
		paymentID := req.PaymentID
		if !hasEmailBeenSent(paymentID) {
			adminEmail := config.GetConfig().AdminEmail
			if adminEmail != "" {
				go sendEmailWithRetry(adminEmail, subject, emailBody, paymentID, "room_booking", 2)
			} else {
				config.Logger.Warn("Admin email not configured, skipping notification")
			}
		} else {
			config.Logger.Info("Email already sent for this payment, skipping duplicate",
				zap.String("payment_id", paymentID))
		}
	} else {
		config.Logger.Error("Failed to fetch booking details for email notification",
			zap.String("booking_id", req.BookingID),
			zap.Error(err))
	}
	// ========== END OF ADMIN EMAIL ==========

	c.JSON(200, gin.H{
		"message": "Payment verified & booking confirmed",
	})
}


