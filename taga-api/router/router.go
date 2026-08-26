package router

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"taga-api/config"
	"taga-api/handler"
	"taga-api/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

// SetupRouter configures and returns the Gin engine with all routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Load configuration
	appCfg := config.GetConfig()
	env := appCfg.Environment
	fmt.Println("Environment:", env)

	// CORS configuration based on environment
	isProd := env == "production" || env == "staging"
	allowOrigins := []string{}
	if !isProd {
		allowOrigins = []string{"*"}
	} else {
		allowOrigins = []string{
			"https://dev.nammataga.com",
			"https://tst.nammataga.com",
			"https://stg.nammataga.com",
			"https://nammataga.com",
			"https://www.nammataga.com",
		}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Middleware
	r.Use(middleware.GinZapLogger())

	// Static file paths
	wd, _ := os.Getwd()
	docsPath := filepath.Join(wd, "data", "docs")

	// ==================== PUBLIC WEBSITE ROUTES (No Auth Required) ====================

	// Health & Root
	r.GET("/", handler.RootHandler)
	r.GET("/health", handler.HealthHandler)

	// About & Public Info
	about := r.Group("/api/public/about")
	{
		about.GET("", handler.AboutHandler)
		about.GET("/stats", handler.AboutStatsHandler)
		about.GET("/objectives", handler.AboutObjectivesHandler)
		about.GET("/services", handler.AboutServicesHandler)
		about.GET("/contact", handler.AboutContactHandler)
	}

	// Assets, Logos & Banners
	r.GET("/api/logo", handler.GetLogo)
	r.GET("/api/member-banner", handler.GetMemberBanner)
	r.GET("/api/resources-banner", handler.GetResourcesBanner)
	r.GET("/api/grievance-banner", handler.GetGrievanceBanner)
	r.GET("/api/categories", handler.GetCategories)
	r.GET("/api/priorities", handler.GetPriorities)

	// Office Information & Bearers (Public)
	r.GET("/api/office", handler.OfficeHandler)
	r.GET("/api/office/:pathParam", handler.OfficeHandler)
	officeBearers := r.Group("/api/office-bearers")
	{
		officeBearers.GET("/state-executive", handler.GetStateExecutive)
		officeBearers.GET("/districts", handler.GetDistricts)
		officeBearers.GET("/district-office-bearers", handler.GetDistrictOfficeBearers)
	}

	// Events & Gallery (Public)
	events := r.Group("/api/events")
	{
		events.GET("/upcoming", handler.UpcomingEventsHandler)
	}

	gallery := r.Group("/api/gallery")
	{
		gallery.GET("/years", handler.GalleryYearsHandler)
		gallery.GET("", handler.GalleryImagesHandler)
	}

	// Membership Application (Public Application)
	membership := r.Group("/api/membership")
	{
		membership.POST("/apply", handler.ApplyMembershipHandler)
		membership.GET("/list", handler.GetMembershipListHandler)
		membership.GET("/districts", handler.GetMembershipDistricts)
	}

	// Subscriptions Listing (Public)
	r.GET("/api/subscriptions", handler.GetSubscriptions)

	// Razorpay Webhook for payment notifications (Verified via HMAC signature)
	r.POST("/api/webhook/razorpay", handler.WebhookHandler)

	// Member Authentication & Recovery (Public)
	r.POST("/api/admin/login", handler.AdminLoginHandler)
	r.POST("/api/member/login", handler.MemberLoginHandler)
	r.POST("/api/member/logout", handler.MemberLogoutHandler)
	r.POST("/api/member/change-password", handler.ChangeMemberPasswordHandler)
	r.POST("/api/member/edit-request", handler.CreateEditRequest)

	auth := r.Group("/api/auth")
	{
		auth.POST("/forgot-password", handler.ForgotPasswordHandler)
		auth.POST("/reset-password", handler.ResetPasswordHandler)
		auth.POST("/member-forgot-password", handler.MemberForgotPasswordHandler)
		auth.POST("/change-password", handler.ChangeMemberPasswordHandler)
	}

	// Test Email & Mock Email Test Hooks (Public for testing)
	r.POST("/api/test-email", handler.TestEmail)
	r.GET("/api/test-email", handler.TestEmail)
	r.GET("/api/admin/mock-emails", handler.GetMockEmails)
	r.DELETE("/api/admin/mock-emails", handler.ClearMockEmails)

	// ==================== PROTECTED MEMBER ROUTES (JWT Required) ====================

	// Member Profile & Notifications (JWT Required)
	memberRoutes := r.Group("/api/member")
	memberRoutes.Use(middleware.MemberAuthMiddleware())
	{
		memberRoutes.GET("/profile", handler.GetMemberProfileByToken)
		memberRoutes.PUT("/profile", handler.UpdateMemberProfileHandler)
		memberRoutes.GET("/notifications", handler.GetMemberNotifications)
		memberRoutes.PUT("/notifications/:id/read", handler.MarkNotificationRead)
		memberRoutes.GET("/notifications/unread/count", handler.GetUnreadCount)
	}

	// Resources (Protected with Member Auth)
	resourcesGroup := r.Group("/api/resources")
	resourcesGroup.Use(middleware.MemberAuthMiddleware())
	{
		resourcesGroup.GET("", handler.GetResourceCategories)
		resourcesGroup.GET("/all", handler.GetAllResources)
		resourcesGroup.GET("/external-links", handler.GetExternalLinks)
		resourcesGroup.GET("/:id", handler.GetDocumentsByCategory)
	}

	// Grievances (Protected with Member/Admin Auth)
	grievances := r.Group("/api/grievances")
	grievances.Use(middleware.MemberAuthMiddleware())
	{
		grievances.POST("", handler.CreateGrievance)
		grievances.GET("", handler.GetGrievances)
		grievances.GET("/:id", handler.GetGrievanceByID)
		grievances.PUT("/:id", handler.UpdateGrievance)
		grievances.DELETE("/:id", handler.DeleteGrievance)
	}

	// TAGA Towers (Rooms, Availability, Booking & Payment - Protected with Member/Admin Auth)
	towers := r.Group("/api/towers")
	towers.Use(middleware.MemberAuthMiddleware())
	{
		towers.GET("/rooms", handler.GetRooms)
		towers.GET("/availability", handler.CheckAvailability)
		towers.GET("/availability-range", handler.CheckAvailabilityRange)
		towers.POST("/bookings", handler.CreateBooking)
		towers.GET("/bookings", handler.GetBookings)
		towers.GET("/bookings/past", handler.GetPastBookings)
		towers.DELETE("/bookings/:id", handler.DeleteBooking)
		towers.POST("/bookings/:id/confirm-payment", handler.ConfirmPayment)
		towers.POST("/create-order", handler.CreateOrder)
		towers.POST("/verify-payment", handler.VerifyPayment)
	}

	// TAGA Towers Admin Occupancy Schedule
	r.GET("/api/towers/admin/bookings", middleware.AdminAuthMiddleware(), handler.GetAllBookingsAdmin)

	// Subscription Payments (Protected with Member Auth)
	subscriptionPaymentProtected := r.Group("/api/subscriptions")
	subscriptionPaymentProtected.Use(middleware.MemberAuthMiddleware())
	{
		subscriptionPaymentProtected.POST("/create-order", handler.CreateSubscriptionOrder)
		subscriptionPaymentProtected.POST("/verify-payment", handler.VerifySubscriptionPayment)
		subscriptionPaymentProtected.GET("/status", handler.GetMemberSubscriptionStatus)
		subscriptionPaymentProtected.GET("/member-paid", handler.GetMemberPaidSubscriptions)
	}

	// Generic Payment Routes (Protected with Member Auth)
	payment := r.Group("/api/payments")
	payment.Use(middleware.MemberAuthMiddleware())
	{
		payment.POST("/create-order", handler.CreateOrder)
		payment.POST("/verify", handler.VerifyPayment)
	}

	// ==================== ADMIN ROUTES (Admin JWT Required) ====================

	// Admin Routes (JWT Required)
	admin := r.Group("/api/admin")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		// Edit Requests Management
		admin.GET("/edit-requests", handler.GetPendingEditRequests)
		admin.POST("/edit-requests/bulk-process", handler.BulkProcessEditRequests)

		// Admin TAGA Towers Occupancy Schedule
		admin.GET("/towers/bookings", handler.GetAllBookingsAdmin)

		// Admin Member Registration
		admin.POST("/member-registration", handler.HandleMemberRegistration)
		admin.POST("/init-password", handler.InitPassword)

		// Content Management - Resources
		admin.POST("/resources/upload", handler.UploadResource)
		admin.DELETE("/resources/:categoryId/:documentTitle", handler.DeleteResource)

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

		// Reports
		admin.GET("/reports/members", handler.GenerateMemberReport)
		admin.GET("/members/export", handler.ExportMembersToExcel)

		// Member Lists & Stats
		admin.GET("/members", handler.GetMembersList)
		admin.GET("/members/districts", handler.GetMemberDistricts)
		admin.GET("/members/stats", handler.GetMemberStats)

		// Announcements
		admin.POST("/announcements/send", handler.SendAnnouncement)

		// Renewal Reminders (Manual Trigger - for testing)
		admin.POST("/send-renewal-reminders", handler.HandleSendRenewalReminders)

		// ==================== OFFICE BEARERS MANAGEMENT ====================
		// District Office Bearers CRUD Operations
		admin.GET("/office-bearers/districts", handler.GetAllDistrictsHandler)
		admin.GET("/office-bearers/district/:district", handler.GetDistrictBearersHandler)
		admin.PUT("/office-bearers/district/:district", handler.UpdateDistrictBearersHandler)
		admin.POST("/office-bearers/backup/restore", handler.RestoreBackupHandler)
		admin.GET("/office-bearers/backups", handler.ListBackupsHandler)

		// ==================== AUDIT LOG ====================
		admin.GET("/audit", handler.GetAuditLogsHandler)
		admin.GET("/audit/users", handler.GetAuditUsersHandler)
	}

	// Legacy Upload Route (Protected with Admin Auth)
	legacyAdmin := r.Group("/admin")
	legacyAdmin.Use(middleware.AdminAuthMiddleware())
	{
		legacyAdmin.POST("/upload-registration", handler.HandleMemberRegistration)
	}

	// ==================== SWAGGER ====================
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ==================== STATIC FILE SERVING ====================
	fmt.Println("✅ Serving /api/images from ./data/image")
	r.Static("/api/images", "./data/image")

	// PDF & Static document serving route
	docsHandler := func(c *gin.Context) {
		relPath := c.Param("filepath")
		fullPath := filepath.Join(docsPath, relPath)

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		if strings.ToLower(filepath.Ext(fullPath)) == ".pdf" {
			c.Writer.Header().Set("Content-Type", "application/pdf")
			c.Writer.Header().Set("Content-Disposition", "inline")
		}
		c.File(fullPath)
	}

	r.GET("/docs/*filepath", docsHandler)
	r.GET("/api/docs/*filepath", docsHandler)
	r.GET("/data/docs/*filepath", docsHandler)
	r.GET("/api/data/docs/*filepath", docsHandler)

	return r
}
