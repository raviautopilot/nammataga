package jwt

import (
	"time"

	"taga-api/config"
)

// Role constants
const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Token expiry durations based on environment and role
var (
	// Admin token expiry (intentionally short-lived)
	AdminTokenExpiryDev   = 24 * time.Hour
	AdminTokenExpiryStage = 8 * time.Hour
	AdminTokenExpiryProd  = 4 * time.Hour

	// Member token expiry defaults (1 week)
	MemberTokenExpiryDev   = 7 * 24 * time.Hour
	MemberTokenExpiryStage = 7 * 24 * time.Hour
	MemberTokenExpiryProd  = 7 * 24 * time.Hour
)

// GetTokenExpiry returns token expiry duration based on environment and role.
// For member tokens, it reads the configurable session_duration_hours from config.json.
// If not configured (0), it falls back to the hardcoded defaults (1 week).
// Admin tokens always use the hardcoded short-lived defaults.
func GetTokenExpiry(environment, role string) time.Duration {
	switch role {
	case RoleAdmin:
		switch environment {
		case "production":
			return AdminTokenExpiryProd
		case "staging":
			return AdminTokenExpiryStage
		default:
			return AdminTokenExpiryDev
		}
	case RoleMember:
		// Use configurable duration if set
		cfg := config.GetConfig()
		if cfg.SessionDurationHours > 0 {
			return time.Duration(cfg.SessionDurationHours) * time.Hour
		}
		// Fallback to hardcoded defaults
		switch environment {
		case "production":
			return MemberTokenExpiryProd
		case "staging":
			return MemberTokenExpiryStage
		default:
			return MemberTokenExpiryDev
		}
	default:
		return 24 * time.Hour
	}
}

