package router

import (
	"taga-api/handler"
	"taga-api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers all public and general-access endpoints
func RegisterPublicRoutes(r *gin.Engine) {
	// Root & Health
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
	r.GET("/api/logo", handler.GetLogo)

	// Razorpay Webhook
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

	// Events
	events := r.Group("/api/events")
	{
		events.GET("/upcoming", handler.UpcomingEventsHandler)
	}

	// Gallery
	gallery := r.Group("/api/gallery")
	{
		gallery.GET("/years", handler.GalleryYearsHandler)
		gallery.GET("", handler.GalleryImagesHandler)
	}

	// Office Bearers
	officeBearers := r.Group("/api/office-bearers")
	{
		officeBearers.GET("/state-executive", handler.GetStateExecutive)
		officeBearers.GET("/districts", handler.GetDistricts)
		officeBearers.GET("/district-office-bearers", handler.GetDistrictOfficeBearers)
	}

	// Office Information
	r.GET("/api/office", handler.OfficeHandler)
	r.GET("/api/office/:pathParam", handler.OfficeHandler)

	// Grievances Group
	grievances := r.Group("/api/grievances")
	{
		grievances.POST("", handler.CreateGrievance)
		grievances.GET("", handler.GetGrievances)
		grievances.GET("/:id", handler.GetGrievanceByID)
		grievances.PUT("/:id", handler.UpdateGrievance)
		grievances.DELETE("/:id", handler.DeleteGrievance)
	}
	r.GET("/api/categories", handler.GetCategories)
	r.GET("/api/priorities", handler.GetPriorities)
}
