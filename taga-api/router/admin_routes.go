package router

import (
	"taga-api/handler"
	"taga-api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAdminRoutes registers all administrative endpoints
func RegisterAdminRoutes(r *gin.Engine) {
	// Admin Login (No Auth)
	r.POST("/api/admin/login", handler.AdminLoginHandler)

	// Protected Admin Routes (JWT Required)
	admin := r.Group("/api/admin")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// Member Registration
		admin.POST("/member-registration", handler.HandleMemberRegistration)
		admin.POST("/init-password", handler.InitPassword)

		// Edit Requests Management
		admin.GET("/edit-requests", handler.GetPendingEditRequests)
		admin.POST("/edit-requests/bulk-process", handler.BulkProcessEditRequests)

		// Content Management - Resources
		admin.POST("/resources/upload", handler.UploadResource)
		admin.DELETE("/resources/:categoryId/:documentTitle", handler.DeleteResource)
		admin.GET("/resources/external-links", handler.GetExternalLinks)
		admin.POST("/resources/external-links", handler.AddExternalLink)
		admin.DELETE("/resources/external-links", handler.DeleteExternalLink)
		admin.DELETE("/resources/external-links/:title", handler.DeleteExternalLink)

		// Content Management - Events
		admin.POST("/events/create", handler.CreateEvent)
		admin.PUT("/events/:id", handler.UpdateEvent)
		admin.DELETE("/events/:id", handler.DeleteEvent)

		// Content Management - Gallery
		admin.POST("/gallery/upload", handler.UploadGalleryImage)
		admin.DELETE("/gallery/:id", handler.DeleteGalleryImage)

		// Member Management
		admin.POST("/members/add", handler.AddMember)
		admin.POST("/members/bulk-upload", handler.BulkUploadMembers)
		admin.PUT("/members/:id", handler.UpdateMember)
		admin.DELETE("/members/:id", handler.DeleteMember)

		// Reports & Export
		admin.GET("/reports/members", handler.GenerateMemberReport)
		admin.GET("/members/export", handler.ExportMembersToExcel)

		// Member Lists & Stats
		admin.GET("/members", handler.GetMembersList)
		admin.GET("/members/districts", handler.GetMemberDistricts)
		admin.GET("/members/stats", handler.GetMemberStats)

		// Announcements & Reminders
		admin.POST("/announcements/send", handler.SendAnnouncement)
		admin.POST("/send-renewal-reminders", handler.HandleSendRenewalReminders)

		// Office Bearers Management
		admin.GET("/office-bearers/districts", handler.GetAllDistrictsHandler)
		admin.GET("/office-bearers/district/:district", handler.GetDistrictBearersHandler)
		admin.PUT("/office-bearers/district/:district", handler.UpdateDistrictBearersHandler)
		admin.POST("/office-bearers/backup/restore", handler.RestoreBackupHandler)
		admin.GET("/office-bearers/backups", handler.ListBackupsHandler)
	}
}
