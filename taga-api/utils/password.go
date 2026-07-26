package utils

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

func GenerateTempPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	rand.Read(b)

	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func HashPassword(pwd string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(hash)
}


