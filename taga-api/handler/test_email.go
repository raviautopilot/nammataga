package handler

import (
	"fmt"
	"net/http"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service/email"
	"taga-api/service"

	"github.com/gin-gonic/gin"
)

// TestEmail godoc
// @Summary Send test premium emails
// @Tags Test
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/test-email [post]
func TestEmail(c *gin.Context) {
	testEmailAddress := "sudhanop05@gmail.com"

	// 1. Room Booking Email Data
	roomData := AdminRoomBookingData{
		BookingID:     "BKG-1234567890",
		PaymentID:     "pay_TQO8llfU2ygOJG",
		OrderID:       "order_TQO7pJWoBiIuV8",
		Amount:        100, // 1.00 Rupee
		CustomerEmail: testEmailAddress,
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
	roomSubjectAdmin := "🏨 [ADMIN] TAGA Room Booking - Apex Suite A/C - ₹1.00"
	roomSubjectMember := "🏨 [MEMBER] TAGA Room Booking - Apex Suite A/C - ₹1.00"

	_ = sendEmail(testEmailAddress, roomSubjectAdmin, roomEmailBody)
	_ = sendEmail(testEmailAddress, roomSubjectMember, roomEmailBody)

	// 2. Subscription Email Data
	subData := AdminSubscriptionData{
		PaymentID:        "pay_SUB8llfU2ygOJG",
		OrderID:          "order_SUB7pJWoBiIuV8",
		Amount:           150000, // 1500.00
		CustomerEmail:    testEmailAddress,
		SubscriptionID:   "SUB-99999",
		SubscriptionName: "Annual TAGA Membership",
		MemberName:       "Sudhan",
		MemberTagaID:     "TAGA-001",
		MemberEmail:      testEmailAddress,
		PaymentType:      "subscription",
	}

	subEmailBody := buildSubscriptionEmailBody(subData)
	subSubjectAdmin := "💰 [ADMIN] TAGA Payment Received - Annual TAGA Membership - ₹1500.00"
	subSubjectMember := "💰 [MEMBER] TAGA Payment Received - Annual TAGA Membership - ₹1500.00"

	_ = sendEmail(testEmailAddress, subSubjectAdmin, subEmailBody)
	_ = sendEmail(testEmailAddress, subSubjectMember, subEmailBody)

	// 3. Welcome / Registration Email
	_ = sendSuccessEmail(testEmailAddress, "TagaPass#2026")

	// 4. Password Reset Email (from email package or via sendEmail)
	cfg := config.GetConfig()
	resetURL := fmt.Sprintf("%s/reset-password?token=sample-reset-token-12345", cfg.ResetPasswordURL)
	_ = email.SendPasswordResetEmail(testEmailAddress, resetURL)

	// 5. Official Broadcast Announcement
	announcementSubject := "📢 [MEMBER] TAGA Official Announcement - 45th Annual General Meeting"
	announcementBody := buildAnnouncementEmailContent("45th Annual General Body Meeting - July 2026", "Dear Members,\n\nWe are pleased to invite all active members to the 45th Annual General Body Meeting scheduled to be held at TAGA Towers, Chennai.\n\nKey highlights include presentation of annual accounts, executive committee elections, and new member recognition.\n\nWe look forward to your gracious presence.", "high")
	_ = sendEmail(testEmailAddress, announcementSubject, announcementBody)

	// 6. Temporary Password Email
	_ = email.SendTemporaryPasswordEmail(testEmailAddress, "Temp1234#")

	// 7. Member Edit Request Emails (Both Admin Notification and Member Processed)
	editFields := []model.FieldEditRequest{
		{
			Field: "mobile_number",
			OldValue: "9944637255",
			NewValue: "9944000000",
			Status: "approved",
			AdminRemarks: "Number verified successfully. We have updated our records.",
			MemberRemarks: "Changed my primary number.",
		},
		{
			Field: "working_district",
			OldValue: "Madurai",
			NewValue: "Kanyakumari",
			Status: "rejected",
			AdminRemarks: "Please provide valid proof of transfer before we can approve this change.",
			MemberRemarks: "I got transferred last week.",
		},
		{
			Field: "residential_address",
			OldValue: "28/a V Nagar Road no 1",
			NewValue: "42 New Street",
			Status: "approved",
			AdminRemarks: "", // Test empty remark rendering
			MemberRemarks: "",
		},
	}
	_ = service.SendAdminEditRequestEmail(testEmailAddress, "SUDHAN", editFields)
	_ = service.SendMemberRequestProcessedEmail(testEmailAddress, "SUDHAN", editFields)

	c.JSON(http.StatusOK, gin.H{
		"message": "All premium email template samples sent successfully to " + testEmailAddress,
	})
}
