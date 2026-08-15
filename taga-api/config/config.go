package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type PaymentConfig struct {
	RazorpayKey    string `json:"razorpay_key"`
	RazorpaySecret string `json:"razorpay_secret"`
}

type FilePathsConfig struct {
	SubscriptionType string `json:"subscript_type"`
	TagaTowerRooms   string `json:"taga_tower_rooms"`
}

type RazorpayConfig struct {
	PaymentFile string `json:"payment_file"`
}

type TowerConfig struct {
}

type DataConfig struct {
	Config   FilePathsConfig `json:"config"`
	Razorpay RazorpayConfig  `json:"razorpay"`
	Tower    TowerConfig     `json:"tower"`
}

type AppConfig struct {
	Port             int        `json:"port"`
	Environment      string     `json:"environment"`
	LogLevel         string     `json:"log_level"`
	LogFile          string     `json:"log_file"`
	DisablePayment   bool       `json:"disable_payment"`
	SMTPHost         string     `json:"smtp_host"`
	SMTPPort         int        `json:"smtp_port"`
	SMTPUsername     string     `json:"smtp_username"`
	SMTPPassword     string     `json:"smtp_password"`
	ResetPasswordURL string     `json:"reset_password_url"`
	FromEmail        string     `json:"from_email"`
	AdminEmail       string     `json:"admin_email"`
	AdminPassword    string     `json:"admin_password"`
	OfficeDir        string     `json:"office_dir"`
	MembersFile           string     `json:"members_file"`
	DeletedMembersFile    string     `json:"deleted_members_file"`
	ProcessedPaymentsFile string     `json:"processed_payments_file"`
	AboutFile             string     `json:"about_file"`
	ContactFile      string     `json:"contact_file"`
	ObjectivesFile   string     `json:"objectives_file"`
	ServicesFile     string     `json:"services_file"`
	StatsFile        string     `json:"stats_file"`
	JwtSecret        string     `json:"jwt_secret"`
	AdminAPIKey      string     `json:"admin_api_key"`
	RazorpayKey      string     `json:"-"`
	RazorpaySecret   string     `json:"-"`
	Data             DataConfig `json:"data"`
}

var (
	Logger *zap.Logger
	Config AppConfig
	once   sync.Once
)

func Init() {
	once.Do(func() {
		loadEnv()
		InitConfig()
		loadRazorpay()
		InitLogger()
		ensureDataFiles()
	})
}

func loadEnv() {
	envPath := ".env"

	if _, err := os.Stat(envPath); err == nil {
		err := godotenv.Load(envPath)
		if err != nil {
			fmt.Println("❌ Failed to load .env file")
		}
	} else {
		fmt.Println("ℹ️ No .env file found, using system environment variables")
	}
}

// loadRazorpay reads Razorpay keys from environment variables only.
// No fallbacks. If keys are missing, it logs a warning but does NOT panic.
// Payment endpoints will later return an error because os.Getenv will be empty.
func loadRazorpay() {
	key := os.Getenv("RAZORPAY_KEY")
	secret := os.Getenv("RAZORPAY_SECRET")

	if key == "" || secret == "" {
		fmt.Println("⚠️ WARNING: RAZORPAY_KEY or RAZORPAY_SECRET not set in environment")
		fmt.Println("   Payment features will not work until these are provided.")
		Config.RazorpayKey = ""
		Config.RazorpaySecret = ""
		return
	}

	Config.RazorpayKey = key
	Config.RazorpaySecret = secret
	fmt.Println("✅ Razorpay loaded from ENV")
	fmt.Printf("   Key: %s...\n", key[:10])
}

func InitConfig() {
	// Try to read config.json
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		// Try parent directories for tests
		for i := 1; i <= 5; i++ {
			parentPath := strings.Repeat("../", i) + "config.json"
			configFile, err = os.ReadFile(parentPath)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		// Use default configuration
		fmt.Println("⚠️ config.json not found, using default configuration")
		Config = AppConfig{
			Port:             1801,
			Environment:      "development",
			LogLevel:         "debug",
			LogFile:          "logs/app.log",
			DisablePayment:   true,
			SMTPHost:         "smtp.gmail.com",
			SMTPPort:         587,
			SMTPUsername:     "",
			SMTPPassword:     "",
			ResetPasswordURL: "http://localhost:1801",
			FromEmail:        "",
			AdminEmail:       "",
			OfficeDir:        "data/office",
			MembersFile:      "data/member/members.json",
			AboutFile:        "data/about/about.json",
			ContactFile:      "data/about/contact.json",
			ObjectivesFile:   "data/about/objectives.json",
			ServicesFile:     "data/about/services.json",
			StatsFile:        "data/about/stats.json",
			JwtSecret:        "taga-default-secret-key-change-in-production",
			AdminAPIKey:      "taga-admin-key-change-in-production",
		}
		return
	}

	// Parse config
	err = json.Unmarshal(configFile, &Config)
	if err != nil {
		panic("Failed to parse config.json: " + err.Error())
	}

	// Set default JWT secret if not provided
	if Config.JwtSecret == "" {
		Config.JwtSecret = "taga-default-secret-key-change-in-production"
	}

	// Set default Admin API key if not provided
	if Config.AdminAPIKey == "" {
		Config.AdminAPIKey = "taga-admin-key-change-in-production"
	}

	// Override configuration with environment variables if set
	if envEnv := os.Getenv("ENVIRONMENT"); envEnv != "" {
		Config.Environment = envEnv
	}
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		var p int
		if _, err := fmt.Sscanf(portEnv, "%d", &p); err == nil {
			Config.Port = p
		}
	}
	if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		Config.LogLevel = logLevelEnv
	}
	if logFileEnv := os.Getenv("LOG_FILE"); logFileEnv != "" {
		Config.LogFile = logFileEnv
	}
	if disablePaymentEnv := os.Getenv("DISABLE_PAYMENT"); disablePaymentEnv != "" {
		Config.DisablePayment = disablePaymentEnv == "true"
	}

	// Set default LogLevel and LogFile if still empty
	if Config.LogLevel == "" {
		if Config.Environment == "development" || Config.Environment == "local" || Config.Environment == "dev" {
			Config.LogLevel = "debug"
		} else {
			Config.LogLevel = "info"
		}
	}
	if Config.LogFile == "" {
		Config.LogFile = "logs/app.log"
	}

	validatePaths()
}

func validatePaths() {
	if Config.AboutFile == "" {
		Config.AboutFile = "data/about/about.json"
	}
	if Config.ContactFile == "" {
		Config.ContactFile = "data/about/contact.json"
	}
	if Config.ObjectivesFile == "" {
		Config.ObjectivesFile = "data/about/objectives.json"
	}
	if Config.ServicesFile == "" {
		Config.ServicesFile = "data/about/services.json"
	}
	if Config.StatsFile == "" {
		Config.StatsFile = "data/about/stats.json"
	}
	if Config.Data.Config.TagaTowerRooms == "" {
		Config.Data.Config.TagaTowerRooms = "data/config/taga-tower-rooms.json"
	}
	if Config.Data.Config.SubscriptionType == "" {
		Config.Data.Config.SubscriptionType = "data/config/subscription-types.json"
	}
	if Config.Data.Razorpay.PaymentFile == "" {
		Config.Data.Razorpay.PaymentFile = "data/config/payment.json"
	}
	if Config.MembersFile == "" {
		Config.MembersFile = "data/member/members.json"
	}
	if Config.DeletedMembersFile == "" {
		Config.DeletedMembersFile = "data/member/deleted_member.json"
	}
	if Config.ProcessedPaymentsFile == "" {
		Config.ProcessedPaymentsFile = "data/payments/processed_payments.json"
	}
}

func ensureDataFiles() {
	dirs := []string{
		"data/about",
		"data/member",
		"data/config",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic("Failed to create directory: " + err.Error())
		}
	}

	files := map[string]string{
		Config.MembersFile:    `[]`,
		Config.AboutFile:      `{"id":1}`,
		Config.ContactFile:    `{}`,
		Config.ObjectivesFile: `[]`,
		Config.ServicesFile:   `[]`,
		Config.StatsFile:      `[]`,
	}

	for filePath, defaultContent := range files {
		if filePath == "" {
			continue
		}
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			dir := filepath.Dir(filePath)
			os.MkdirAll(dir, 0755)
			err := os.WriteFile(filePath, []byte(defaultContent), 0644)
			if err != nil {
				panic("Failed to create " + filePath + ": " + err.Error())
			}
			if Logger != nil {
				Logger.Info("Created file", zap.String("file", filePath))
			}
		}
	}
}

func InitLogger() {
	// Define level
	var level zapcore.Level
	switch strings.ToLower(Config.LogLevel) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		if Config.Environment == "development" || Config.Environment == "local" || Config.Environment == "dev" {
			level = zapcore.DebugLevel
		} else {
			level = zapcore.InfoLevel
		}
	}

	// Cores slice to hold all outputs
	var cores []zapcore.Core

	// 1. Console Core
	var consoleSyncer zapcore.WriteSyncer
	if Config.Environment == "development" || Config.Environment == "local" || Config.Environment == "dev" {
		consoleSyncer = zapcore.Lock(os.Stdout)
	} else {
		consoleSyncer = zapcore.Lock(os.Stderr)
	}

	var consoleEncoder zapcore.Encoder
	if Config.Environment == "development" || Config.Environment == "local" || Config.Environment == "dev" {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	cores = append(cores, zapcore.NewCore(consoleEncoder, consoleSyncer, level))

	// 2. File Core (if LogFile is specified and not empty)
	if Config.LogFile != "" {
		logDir := filepath.Dir(Config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("WARNING: Failed to create log directory for %s: %v\n", Config.LogFile, err)
		} else {
			file, err := os.OpenFile(Config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("WARNING: Failed to open log file %s: %v\n", Config.LogFile, err)
			} else {
				fileSyncer := zapcore.AddSync(file)
				var fileEncoder zapcore.Encoder
				if Config.Environment == "development" || Config.Environment == "local" || Config.Environment == "dev" {
					fileEncoderConfig := zap.NewDevelopmentEncoderConfig()
					fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
					fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
					fileEncoder = zapcore.NewConsoleEncoder(fileEncoderConfig)
				} else {
					fileEncoderConfig := zap.NewProductionEncoderConfig()
					fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
					fileEncoder = zapcore.NewJSONEncoder(fileEncoderConfig)
				}
				cores = append(cores, zapcore.NewCore(fileEncoder, fileSyncer, level))
			}
		}
	}

	// Combine cores using zapcore.NewTee
	core := zapcore.NewTee(cores...)

	// Build logger with caller info
	Logger = zap.New(core, zap.AddCaller())
}

// GetConfig returns the application configuration
func GetConfig() AppConfig {
	return Config
}
