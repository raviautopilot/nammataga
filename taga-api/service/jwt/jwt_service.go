package jwt

import (
	"fmt"
	"time"

	"taga-api/config"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAdminToken generates a JWT token for admin users
// Returns: tokenString, expiresInSeconds, error
func GenerateAdminToken(username string) (string, int64, error) {
	cfg := config.GetConfig()
	secret := cfg.JwtSecret
	if secret == "" {
		secret = "ilovetaga"
	}

	expiryDuration := GetTokenExpiry(cfg.Environment, RoleAdmin)
	expirationTime := time.Now().Add(expiryDuration)
	expiresIn := int64(expiryDuration.Seconds())

	claims := AdminClaims{
		Username: username,
		UserID:   "admin-001",
		Role:     RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "taga-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresIn, nil
}

// GenerateMemberToken generates a JWT token for member users
// Parameters: memberID, email, name
// Returns: tokenString, expiresInSeconds, error
func GenerateMemberToken(memberID, email, name string) (string, int64, error) {
	cfg := config.GetConfig()
	secret := cfg.JwtSecret
	if secret == "" {
		secret = "ilovetaga"
	}

	expiryDuration := GetTokenExpiry(cfg.Environment, RoleMember)
	expirationTime := time.Now().Add(expiryDuration)
	expiresIn := int64(expiryDuration.Seconds())

	claims := MemberClaims{
		MemberID: memberID,
		Email:    email,
		Name:     name,
		Role:     RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "taga-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresIn, nil
}

// ValidateToken parses and validates any JWT token
// Returns claims and error
func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	cfg := config.GetConfig()
	secret := cfg.JwtSecret
	if secret == "" {
		secret = "ilovetaga"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check expiry explicitly
		if exp, ok := claims["exp"]; ok {
			if expFloat, ok := exp.(float64); ok {
				if int64(expFloat) < time.Now().Unix() {
					return nil, fmt.Errorf("token has expired")
				}
			}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// ValidateAdminToken validates token and checks if role is admin
func ValidateAdminToken(tokenString string) (jwt.MapClaims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	role, ok := claims["role"]
	if !ok || role != RoleAdmin {
		return nil, fmt.Errorf("admin access required")
	}

	return claims, nil
}

// ValidateMemberToken validates token and checks if role is member or admin
func ValidateMemberToken(tokenString string) (jwt.MapClaims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	role, ok := claims["role"]
	if !ok || (role != RoleMember && role != RoleAdmin) {
		return nil, fmt.Errorf("member or admin access required")
	}

	return claims, nil
}


