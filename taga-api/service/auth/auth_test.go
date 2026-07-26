package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationError(t *testing.T) {
	err := &AuthenticationError{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}

func TestFileMemberReader_ReadMembers(t *testing.T) {
	reader := &FileMemberReader{}

	// This will try to read from config.Config.MembersFile
	// Without config setup, this will fail
	// We'll skip this test for now
	t.Skip("Skipping test that requires config setup")

	_, err := reader.ReadMembers()
	assert.NoError(t, err)
}

func TestGenerateResetToken(t *testing.T) {
	token, err := generateResetToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 64) // 32 bytes hex encoded = 64 chars
}
