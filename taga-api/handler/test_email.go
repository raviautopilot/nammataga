package handler

import (
	"fmt"
	"net/http"
	"taga-api/config"

	"github.com/gin-gonic/gin"
)

// TestEmail godoc
// @Summary Send test premium emails
// @Tags Test
// @Accept json
// @Produce json
// @Router /api/test-email [post]
func TestEmail(c *gin.Context) {
	testEmailAddress := "raviregi@gmail.com"

	// 1. Room Booking Email Data
	roomData := AdminRoomBookingData{
		BookingID:     "BKG-1234567890",
		PaymentID:     "pay_TQO8llfU2ygOJG",
		OrderID:       "order_TQO7pJWoBiIuV8",
		Amount:        100, // 1.00 Rupee
		CustomerEmail: "customer@example.com",
		CustomerPhone: "+917604832246",
		RoomName:      "Apex Suite A/C",
		RoomNumber:    "",
		BedCount:      1,
		CheckInDate:   "2026-08-24",
		CheckOutDate:  "2026-08-25",
		BookerName:    "Sudhan",
		BookerTagaID:  "TAGA-10293",
		BookerPhone:   "+917604832246",
		BookingFor:    "guest",
		GuestDetails:  `[{"name":"John Doe","age":32,"contact":"+919876543210"},{"name":"Jane Doe","age":28,"contact":"+919876543211"}]`,
		PaymentType:   "room_booking",
	}

	roomEmailBody := buildRoomBookingEmailBody(roomData)
	roomSubjectAdmin := "🏨 [ADMIN] TEST TAGA Room Booking - Apex Suite A/C - ₹1.00"
	roomSubjectMember := "🏨 [MEMBER] TEST TAGA Room Booking - Apex Suite A/C - ₹1.00"

	err1 := sendEmail(testEmailAddress, roomSubjectAdmin, roomEmailBody)
	err2 := sendEmail(testEmailAddress, roomSubjectMember, roomEmailBody)

	// 2. Subscription Email Data
	subData := AdminSubscriptionData{
		PaymentID:        "pay_SUB8llfU2ygOJG",
		OrderID:          "order_SUB7pJWoBiIuV8",
		Amount:           150000, // 1500.00
		CustomerEmail:    testEmailAddress,
		SubscriptionID:   "SUB-99999",
		SubscriptionName: "Annual TAGA Membership",
		MemberName:       "Test Member",
		MemberTagaID:     "TAGA-001",
		MemberEmail:      testEmailAddress,
		PaymentType:      "subscription",
	}

	subEmailBody := buildSubscriptionEmailBody(subData)
	subSubjectAdmin := "💰 [ADMIN] TEST TAGA Payment Received - Annual TAGA Membership - ₹1500.00"
	subSubjectMember := "💰 [MEMBER] TEST TAGA Payment Received - Annual TAGA Membership - ₹1500.00"

	err3 := sendEmail(testEmailAddress, subSubjectAdmin, subEmailBody)
	err4 := sendEmail(testEmailAddress, subSubjectMember, subEmailBody)

	// 3. Welcome / Registration Email
	_ = sendSuccessEmail(testEmailAddress, "TagaPass#2026")

	// 4. Password Reset Email (from email package or via sendEmail)
	cfg := config.GetConfig()
	resetURL := fmt.Sprintf("%s/reset-password?token=sample-reset-token-12345", cfg.ResetPasswordURL)
	resetSubject := "🔒 [SAMPLE] TAGA Portal - Password Reset Request"
	resetBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Request</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        <div style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Account Security
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">Password Reset Request</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Secure access recovery for your TAGA Portal account</p>
        </div>
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                We received a request to reset the password for your TAGA Member account associated with <strong>%s</strong>.
            </p>
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block;">
                    Reset My Password &rarr;
                </a>
            </div>
            <div style="background: #fef2f2; border-left: 4px solid #ef4444; padding: 14px 16px; border-radius: 4px;">
                <p style="margin: 0; font-size: 13px; color: #991b1b;">
                    <strong>Didn't request this?</strong> If you did not make this request, you can safely ignore this email.
                </p>
            </div>
        </div>
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">&copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, testEmailAddress, resetURL)
	_ = sendEmail(testEmailAddress, resetSubject, resetBody)

	// 5. Official Broadcast Announcement
	announcementSubject := "📢 [SAMPLE] TAGA Official Announcement - 45th Annual General Meeting"
	announcementBody := buildAnnouncementEmailContent("45th Annual General Body Meeting - July 2026", "Dear Members,\n\nWe are pleased to invite all active members to the 45th Annual General Body Meeting scheduled to be held at TAGA Towers, Chennai.\n\nKey highlights include presentation of annual accounts, executive committee elections, and new member recognition.\n\nWe look forward to your gracious presence.", "high")
	_ = sendEmail(testEmailAddress, announcementSubject, announcementBody)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Email sending failed. Room err: %v, Sub err: %v", err1, err3),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All premium email template samples sent successfully to " + testEmailAddress,
	})
}
