package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"taga-api/config"
	"taga-api/model"
	esrv "taga-api/service/email"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type MemberReader interface {
	ReadMembers() ([]model.Member, error)
}

type FileMemberReader struct{}

func (r *FileMemberReader) ReadMembers() ([]model.Member, error) {
	filename := config.Config.MembersFile
	if filename == "" {
		config.Logger.Error("MembersFile is empty in config")
		return nil, &AuthenticationError{"configuration error: members file path is not set"}
	}

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			config.Logger.Error("Failed to create directory", zap.String("dir", dir), zap.Error(err))
			return nil, err
		}
		// Create the file with an empty array
		emptyMembers := []model.Member{}
		data, err := json.MarshalIndent(emptyMembers, "", "  ")
		if err != nil {
			config.Logger.Error("Failed to marshal empty members", zap.Error(err))
			return nil, err
		}
		if err := os.WriteFile(filename, data, 0644); err != nil {
			config.Logger.Error("Failed to write members file", zap.String("filename", filename), zap.Error(err))
			return nil, err
		}
		config.Logger.Info("Created new members file", zap.String("filename", filename))
		return emptyMembers, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		config.Logger.Error("Failed to read members file", zap.String("filename", filename), zap.Error(err))
		return nil, err
	}

	var members []model.Member
	err = json.Unmarshal(data, &members)
	if err != nil {
		config.Logger.Error("Failed to unmarshal members data", zap.String("filename", filename), zap.Error(err))
		return nil, err
	}

	return members, nil
}

// Default reader uses the production file
var defaultReader MemberReader = &FileMemberReader{}

// Store reset tokens temporarily
var resetTokens = make(map[string]string) // token -> email
var tokenMutex = &sync.RWMutex{}



type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ForgotPassword(email string) error {
	// Check if the email exists
	members, err := defaultReader.ReadMembers()
	if err != nil {
		return err
	}

	for _, member := range members {
		if member.EmailId == email {
			// Generate reset token
			token, err := generateResetToken()
			if err != nil {
				return err
			}

			// Store token
			tokenMutex.Lock()
			resetTokens[token] = email
			tokenMutex.Unlock()

			// Send email
			err = esrv.SendPasswordResetEmail(email, token)

			if err != nil {
				return err
			}

			return nil
		}
	}

	return &AuthenticationError{"email not found"}
}

func ResetPassword(email, oldPassword, newPassword string) error {
	// Validate input parameters
	if email == "" {
		return &AuthenticationError{"email cannot be empty"}
	}

	if oldPassword == "" {
		return &AuthenticationError{"old password cannot be empty"}
	}

	if newPassword == "" {
		return &AuthenticationError{"new password cannot be empty"}
	}

	// Read members
	members, err := defaultReader.ReadMembers()
	if err != nil {
		return fmt.Errorf("failed to read members: %w", err)
	}

	// Find member by email and validate old password
	memberIndex := -1
	for i, member := range members {
		if member.EmailId == email {
			memberIndex = i
			// Verify old password
			if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(oldPassword)); err != nil {
				return &AuthenticationError{"invalid old password"}
			}
			break
		}
	}

	if memberIndex == -1 {
		return &AuthenticationError{"email not found"}
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	members[memberIndex].Password = string(hashedPassword)
	members[memberIndex].FirstLogin = false

	// Write back to file
	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}

	err = os.WriteFile(config.Config.MembersFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write members file: %w", err)
	}

	return nil
}

func GenerateAndSendTemporaryPassword(email, tbfNumber string) error {
	// Read members
	members, err := defaultReader.ReadMembers()
	if err != nil {
		return fmt.Errorf("failed to read members: %w", err)
	}

	memberIndex := -1
	for i, member := range members {
		if (email != "" && member.EmailId == email) || (tbfNumber != "" && member.TbfNumber == tbfNumber) {
			memberIndex = i
			break
		}
	}

	if memberIndex == -1 {
		return &AuthenticationError{"member not found with provided details"}
	}

	// Generate a 12-char temporary password
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tempBytes := make([]byte, 12)
	for i := range tempBytes {
		b := make([]byte, 1)
		_, _ = rand.Read(b)
		tempBytes[i] = charset[int(b[0])%len(charset)]
	}
	tempPassword := string(tempBytes)

	// Hash it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash temp password: %w", err)
	}

	// Update password and force change on next login
	members[memberIndex].Password = string(hashedPassword)
	members[memberIndex].FirstLogin = true

	// Write back to file
	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}

	err = os.WriteFile(config.Config.MembersFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write members file: %w", err)
	}

	// Send temporary password email
	return esrv.SendTemporaryPasswordEmail(members[memberIndex].EmailId, tempPassword)
}
