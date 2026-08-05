package router

import (
	"taga-api/handler"
	"taga-api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPaymentRoutes registers subscription and room booking (TAGA Towers) payment routes
func RegisterPaymentRoutes(r *gin.Engine) {
	// TAGA Towers (Room booking & payments)
	towers := r.Group("/api/towers")
	{
		towers.GET("/rooms", handler.GetRooms)
		towers.GET("/availability", handler.CheckAvailability)
		towers.POST("/bookings", handler.CreateBooking)
		towers.GET("/bookings", handler.GetBookings)
		towers.DELETE("/bookings/:id", handler.DeleteBooking)
		towers.POST("/bookings/:id/confirm-payment", handler.ConfirmPayment)
		towers.POST("/create-order", handler.CreateOrder)
		towers.POST("/verify-payment", handler.VerifyPayment)
		towers.GET("/admin/bookings", handler.GetAllBookingsAdmin)
	}

	// Subscriptions (Public listing)
	r.GET("/api/subscriptions", handler.GetSubscriptions)

	// Subscription Payments (Protected with Member Auth)
	subProtected := r.Group("/api/subscriptions")
	subProtected.Use(middleware.MemberAuthMiddleware())
	{
		subProtected.POST("/create-order", handler.CreateSubscriptionOrder)
		subProtected.POST("/verify-payment", handler.VerifySubscriptionPayment)
		subProtected.GET("/status", handler.GetMemberSubscriptionStatus)
		subProtected.GET("/member-paid", handler.GetMemberPaidSubscriptions)
	}
}
