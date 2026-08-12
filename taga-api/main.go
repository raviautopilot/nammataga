package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"taga-api/config"
	_ "taga-api/docs"
	"taga-api/router"
	"taga-api/service"
	"taga-api/service/member"

	"go.uber.org/zap"
)

// @title Taga API
// @version 1.0
// @description This is the Taga API server
// @BasePath /

func main() {
	// Load .env file from the same directory as the executable
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize config and logger (still needed for other settings)
	config.Init()
	defer func() {
		_ = config.Logger.Sync()
	}()

	// Initialize member repository
	if err := member.InitMemberRepository(); err != nil {
		config.Logger.Error("Failed to initialize member repository", zap.Error(err))
	}

	// Log the configuration loading details
	config.Logger.Info("Configuration loaded",
		zap.String("members_file", config.Config.MembersFile),
		zap.String("environment", config.Config.Environment),
		zap.String("log_level", config.Config.LogLevel),
		zap.String("log_file", config.Config.LogFile))

	// Set Gin to release mode for production
	if config.Config.Environment == "production" {
		// gin.SetMode(gin.ReleaseMode)
	}

	// Start the daily booking cleanup scheduler
	service.StartBookingCleanupScheduler()

	// Start the renewal reminder scheduler
	service.StartScheduler()

	// Start the audit log retention cleanup scheduler
	service.StartAuditCleanupScheduler()

	// Setup router
	r := router.SetupRouter()

	config.Logger.Info("Starting server",
		zap.Int("port", config.Config.Port),
		zap.String("environment", config.Config.Environment),
	)

	port := fmt.Sprintf(":%d", config.Config.Port)
	r.Run(port)
}
