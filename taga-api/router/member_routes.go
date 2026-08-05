package router

import (
	"taga-api/handler"
	"taga-api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterMemberRoutes registers all member auth, profile, and membership endpoints
func RegisterMemberRoutes(r *gin.Engine) {
	// Member Password / Auth Recovery (No JWT required)
	auth := r.Group("/api/auth")
	{
		auth.POST("/forgot-password", handler.ForgotPasswordHandler)
		auth.POST("/reset-password", handler.ResetPasswordHandler)
		auth.POST("/member-forgot-password", handler.MemberForgotPasswordHandler)
	}

	// Member Authentication & Requests (Public)
	r.POST("/api/member/login", handler.MemberLoginHandler)
	r.POST("/api/member/logout", handler.MemberLogoutHandler)
	r.POST("/api/member/change-password", handler.ChangeMemberPasswordHandler)
	r.POST("/api/member/edit-request", handler.CreateEditRequest)

	// Protected Member Operations (JWT Required)
	memberProtected := r.Group("/api/member")
	memberProtected.Use(middleware.MemberAuthMiddleware())
	{
		memberProtected.GET("/profile", handler.GetMemberProfileByToken)
		memberProtected.PUT("/profile", handler.UpdateMemberProfileHandler)
		memberProtected.GET("/notifications", handler.GetMemberNotifications)
		memberProtected.PUT("/notifications/:id/read", handler.MarkNotificationRead)
		memberProtected.GET("/notifications/unread/count", handler.GetUnreadCount)
	}

	// Membership Registration / Details
	membership := r.Group("/api/membership")
	{
		membership.POST("/apply", handler.ApplyMembershipHandler)
		membership.GET("/list", handler.GetMembershipListHandler)
		membership.GET("/districts", handler.GetMembershipDistricts)
	}
}
