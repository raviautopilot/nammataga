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

	// ==================== PUBLIC ROUTES (No Auth) ====================

	// Health & Root
	r.GET("/", handler.RootHandler)
	r.GET("/health", handler.HealthHandler)

	// About & Public Info
	r.GET("/api/public/about", handler.AboutHandler)
	r.GET("/api/public/about/stats", handler.AboutStatsHandler)
	r.GET("/api/public/about/objectives", handler.AboutObjectivesHandler)
	r.GET("/api/public/about/services", handler.AboutServicesHandler)
	r.GET("/api/public/about/contact", handler.AboutContactHandler)
	r.GET("/api/logo", handler.GetLogo)
	r.GET("/api/member-banner", handler.GetMemberBanner)

	// Razorpay Webhook for payment notifications
	r.POST("/api/webhook/razorpay", handler.WebhookHandler)

	// Resources (Protected with Member Auth)
	resourcesGroup := r.Group("/api/resources")
	resourcesGroup.Use(middleware.MemberAuthMiddleware())
	{
		resourcesGroup.GET("", handler.GetResourceCategories)
		resourcesGroup.GET("/all", handler.GetAllResources)
		resourcesGroup.GET("/external-links", handler.GetExternalLinks)
		resourcesGroup.GET("/:id", handler.GetDocumentsByCategory)
	}
	r.GET("/api/resources-banner", handler.GetResourcesBanner)

	// Events (Public)
	events := r.Group("/api/events")
	{
		events.GET("/upcoming", handler.UpcomingEventsHandler)
	}

	// Gallery (Public)
	gallery := r.Group("/api/gallery")
	{
		gallery.GET("/years", handler.GalleryYearsHandler)
		gallery.GET("", handler.GalleryImagesHandler)
	}

	// Office Bearers (Public)
	officeBearers := r.Group("/api/office-bearers")
	{
		officeBearers.GET("/state-executive", handler.GetStateExecutive)
		officeBearers.GET("/districts", handler.GetDistricts)
		officeBearers.GET("/district-office-bearers", handler.GetDistrictOfficeBearers)
	}

	// Office Information (Public)
	r.GET("/api/office", handler.OfficeHandler)
	r.GET("/api/office/:pathParam", handler.OfficeHandler)

	// Grievances
	r.POST("/api/grievances", handler.CreateGrievance)
	r.GET("/api/grievances", handler.GetGrievances)
	r.GET("/api/grievances/:id", handler.GetGrievanceByID)
	r.PUT("/api/grievances/:id", handler.UpdateGrievance)
	r.DELETE("/api/grievances/:id", handler.DeleteGrievance)
	r.GET("/api/grievance-banner", handler.GetGrievanceBanner)
	r.GET("/api/categories", handler.GetCategories)
	r.GET("/api/priorities", handler.GetPriorities)

	// TAGA Towers (Public — member auth handled at frontend level)
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

	// ==================== MEMBER AUTHENTICATION ROUTES ====================

	// Member Auth (No JWT required)
	auth := r.Group("/api/auth")
	{
		auth.POST("/forgot-password", handler.ForgotPasswordHandler)
		auth.POST("/reset-password", handler.ResetPasswordHandler)
		auth.POST("/member-forgot-password", handler.MemberForgotPasswordHandler)
	}

	// Member Login/Logout (No JWT required)
	r.POST("/api/member/login", handler.MemberLoginHandler)
	r.POST("/api/member/logout", handler.MemberLogoutHandler)
	r.POST("/api/member/change-password", handler.ChangeMemberPasswordHandler)

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

	// Member Edit Request
	api := r.Group("/api")
	{
		api.POST("/member/edit-request", handler.CreateEditRequest)
	}

	// ==================== MEMBERSHIP ROUTES ====================
	membership := r.Group("/api/membership")
	{
		membership.POST("/apply", handler.ApplyMembershipHandler)
		membership.GET("/list", handler.GetMembershipListHandler)
		membership.GET("/districts", handler.GetMembershipDistricts)
	}

	// ==================== SUBSCRIPTION & PAYMENT ROUTES ====================

	// Subscription (Public)
	subscriptionPayment := r.Group("/api/subscriptions")
	{
		subscriptionPayment.GET("", handler.GetSubscriptions)
	}

	// Payment (Protected with Member Auth)
	payment := r.Group("/api/payments")
	payment.Use(middleware.MemberAuthMiddleware())
	{
		payment.POST("/create-order", handler.CreateOrder)
		payment.POST("/verify", handler.VerifyPayment)
	}

	// Subscription Payment (Protected with Member Auth)
	subscriptionPaymentProtected := r.Group("/api/subscriptions")
	subscriptionPaymentProtected.Use(middleware.MemberAuthMiddleware())
	{
		subscriptionPaymentProtected.POST("/create-order", handler.CreateSubscriptionOrder)
		subscriptionPaymentProtected.POST("/verify-payment", handler.VerifySubscriptionPayment)
		subscriptionPaymentProtected.GET("/status", handler.GetMemberSubscriptionStatus)
		subscriptionPaymentProtected.GET("/member-paid", handler.GetMemberPaidSubscriptions)
	}

	// ==================== ADMIN ROUTES ====================

	// Admin Login (No Auth)
	r.POST("/api/admin/login", handler.AdminLoginHandler)

	// Admin Routes (JWT Required)
	admin := r.Group("/api/admin")
	admin.Use(middleware.AdminAuthMiddleware())
	{
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
