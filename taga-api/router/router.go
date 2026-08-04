package router

import (
	"fmt"
	"taga-api/config"
	"taga-api/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	// Domain Route Registration
	RegisterPublicRoutes(r)
	RegisterMemberRoutes(r)
	RegisterPaymentRoutes(r)
	RegisterAdminRoutes(r)
	RegisterStaticRoutes(r)

	return r
}
