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

	// Member Authentication (Public login & password recovery)
	r.POST("/api/member/login", handler.MemberLoginHandler)

	// Protected Member Operations (JWT Required)
	memberProtected := r.Group("/api/member")
	memberProtected.Use(middleware.MemberAuthMiddleware())
	{
		memberProtected.POST("/logout", handler.MemberLogoutHandler)
		memberProtected.POST("/change-password", handler.ChangeMemberPasswordHandler)
		memberProtected.POST("/edit-request", handler.CreateEditRequest)
		memberProtected.GET("/profile", handler.GetMemberProfileByToken)
		memberProtected.PUT("/profile", handler.UpdateMemberProfileHandler)
		memberProtected.GET("/notifications", handler.GetMemberNotifications)
		memberProtected.PUT("/notifications/:id/read", handler.MarkNotificationRead)
		memberProtected.GET("/notifications/unread/count", handler.GetUnreadCount)
	}

	// Membership Registration / Details (Protected with Member Auth)
	membership := r.Group("/api/membership")
	membership.Use(middleware.MemberAuthMiddleware())
	{
		membership.POST("/apply", handler.ApplyMembershipHandler)
		membership.GET("/list", handler.GetMembershipListHandler)
		membership.GET("/districts", handler.GetMembershipDistricts)
	}
}
