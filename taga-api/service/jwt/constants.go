package jwt

import "time"

// Role constants
const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Token expiry durations based on environment and role
var (
	// Admin token expiry
	AdminTokenExpiryDev   = 24 * time.Hour
	AdminTokenExpiryStage = 8 * time.Hour
	AdminTokenExpiryProd  = 4 * time.Hour

	// Member token expiry
	MemberTokenExpiryDev   = 72 * time.Hour
	MemberTokenExpiryStage = 48 * time.Hour
	MemberTokenExpiryProd  = 24 * time.Hour
)

// GetTokenExpiry returns token expiry duration based on environment and role
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
