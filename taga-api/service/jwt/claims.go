package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims defines the structure of JWT claims
type CustomClaims struct {
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AdminClaims creates claims for admin users
type AdminClaims struct {
	Username string `json:"username"`
	UserID   string `json:"userID"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// MemberClaims creates claims for member users
type MemberClaims struct {
	MemberID string `json:"member_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
