// handler/admin_notify.go
package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"taga-api/config"
	"time"

	"go.uber.org/zap"
)

// AdminSubscriptionData contains data for subscription payment admin email
type AdminSubscriptionData struct {
	PaymentID        string
	OrderID          string
	Amount           int
	CustomerEmail    string
	SubscriptionID   string
	SubscriptionName string
	MemberName       string
	MemberTagaID     string
	MemberEmail      string
	PaymentType      string // "subscription"
}

// AdminRoomBookingData contains data for room booking payment admin email
type AdminRoomBookingData struct {
	PaymentID     string
	OrderID       string
	Amount        int
	CustomerEmail string
	RoomName      string
	RoomNumber    string
	BedCount      int
	CheckInDate   string
	CheckOutDate  string
	BookerName    string
	BookerTagaID  string
	BookerPhone   string
	BookingFor    string // "self" or "guest"
	GuestDetails  string // JSON string of guests
	PaymentType   string // "room_booking"
}

// sendAdminSubscriptionEmail sends a detailed HTML email to admin for subscription payments
func sendAdminSubscriptionEmail(data AdminSubscriptionData) error {
	cfg := config.GetConfig()
	adminEmail := cfg.AdminEmail

	// Log email attempt
	config.Logger.Info("📧 Attempting to send subscription admin email",
		zap.String("admin_email", adminEmail),
		zap.String("subscription", data.SubscriptionName),
		zap.String("payment_id", data.PaymentID),
	)

	if adminEmail == "" {
		config.Logger.Warn("Admin email not configured, skipping notification")
		return nil
	}

	amountInRupees := float64(data.Amount) / 100
	subject := fmt.Sprintf("💰 New Subscription Payment: %s", data.SubscriptionName)

	var body strings.Builder

	// Header
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
        .label { font-weight: bold; width: 35%%; background: #f3f4f6; }
        .value { background: white; }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: bold; background: #dbeafe; color: #1e40af; }
        h3 { color: #065f46; margin-top: 20px; margin-bottom: 10px; }
        hr { border: none; border-top: 1px solid #e5e7eb; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>TAGA Payment Notification</h2>
            <p>%s</p>
        </div>
        <div class="content">
            <div style="text-align: center; margin-bottom: 20px;">
                <span class="badge">📋 SUBSCRIPTION PAYMENT</span>
            </div>
            
            <h3>💰 Payment Summary</h3>
            <tr>
                <tr><td class="label">Payment ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Order ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Amount:</td><td class="value">₹ %.2f</td><tr>
                <tr><td class="label">Customer Email:</td><td class="value">%s</td><tr>
            云
 
            
            <h3>📋 Subscription Details</h3>
            <table>
                <tr><td class="label">Subscription Type:</td><td class="value">%s</td><tr>
                <tr><td class="label">Subscription ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Member Name:</td><td class="value">%s</td><tr>
                <tr><td class="label">TAGA ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Member Email:</td><td class="value">%s</td><tr>
            表
            
            <hr>
            <p style="font-size: 12px; color: #6b7280; text-align: center;">
                This is an automated notification from TAGA Payment System.
            </p>
        </div>
        <div class="footer">
            <p>TAGA Towers | Agriculture Complex Road, Chennai - 600017</p>
        </div>
    </div>
</body>
</html>`,
		time.Now().Format("January 02, 2006 at 03:04 PM"),
		data.PaymentID,
		data.OrderID,
		amountInRupees,
		data.CustomerEmail,
		data.SubscriptionName,
		data.SubscriptionID,
		data.MemberName,
		data.MemberTagaID,
		data.MemberEmail,
	)

	return sendEmail(adminEmail, subject, body.String())
}

// sendAdminRoomBookingEmail sends a detailed HTML email to admin for room booking payments
func sendAdminRoomBookingEmail(data AdminRoomBookingData) error {
	cfg := config.GetConfig()
	adminEmail := cfg.AdminEmail

	// Log email attempt
	config.Logger.Info("📧 Attempting to send room booking admin email",
		zap.String("admin_email", adminEmail),
		zap.String("room", data.RoomName),
		zap.String("payment_id", data.PaymentID),
	)

	if adminEmail == "" {
		config.Logger.Warn("Admin email not configured, skipping notification")
		return nil
	}

	amountInRupees := float64(data.Amount) / 100
	subject := fmt.Sprintf("🏨 New Room Booking: %s", data.RoomName)

	var body strings.Builder

	// Parse guest details if present
	var guestList []map[string]interface{}
	guestTableHTML := ""
	if data.GuestDetails != "" && data.GuestDetails != "[]" {
		if err := json.Unmarshal([]byte(data.GuestDetails), &guestList); err == nil && len(guestList) > 0 {
			guestTableHTML = `<h3>👥 Guest Details</h3>
<table>
    <thead>
        <tr style="background: #f3f4f6;">
            <th style="padding: 10px; text-align: left;">#</th>
            <th style="padding: 10px; text-align: left;">Name</th>
            <th style="padding: 10px; text-align: left;">Age</th>
            <th style="padding: 10px; text-align: left;">Contact</th>
        </tr>
    </thead>
    <tbody>`
			for i, guest := range guestList {
				name, _ := guest["name"].(string)
				age, _ := guest["age"].(float64)
				contact, _ := guest["contact"].(string)
				guestTableHTML += fmt.Sprintf(`
        <tr>
            <td style="padding: 8px; border-bottom: 1px solid #e5e7eb;">%d</td>
            <td style="padding: 8px; border-bottom: 1px solid #e5e7eb;">%s</td>
            <td style="padding: 8px; border-bottom: 1px solid #e5e7eb;">%.0f</td>
            <td style="padding: 8px; border-bottom: 1px solid #e5e7eb;">%s</td>
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

	// Build HTML Email
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
                <tr><td class="label">Payment ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Order ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Advance Amount:</td><td class="value">₹ %.2f</td><tr>
                <tr><td class="label">Customer Email:</td><td class="value">%s</td><tr>
            表
            
            <h3>🏨 Room Details</h3>
            <table>
                <tr><td class="label">Room Name:</td><td class="value">%s</td><tr>
                <tr><td class="label">Room Number:</td><td class="value">%s</td><tr>
                <tr><td class="label">Bed Count:</td><td class="value">%d bed(s)</td><tr>
                <tr><td class="label">Check-in Date:</td><td class="value">%s</td><tr>
                <tr><td class="label">Check-out Date:</td><td class="value">%s</td><tr>
            表
            
            <h3>👤 Booker Details</h3>
            <table>
                <tr><td class="label">Name:</td><td class="value">%s</td><tr>
                <tr><td class="label">TAGA ID:</td><td class="value">%s</td><tr>
                <tr><td class="label">Phone:</td><td class="value">%s</td><tr>
                <tr><td class="label">Booking For:</td><td class="value">%s</td><tr>
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

	return sendEmail(adminEmail, subject, body.String())

}
