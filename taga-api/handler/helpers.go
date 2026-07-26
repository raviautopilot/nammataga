// handler/helpers.go
package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/smtp"
	"os"
	"strings"
	"taga-api/config"

	"go.uber.org/zap"
)

// readExistingMembers reads and returns all members from the members file
func readExistingMembers() ([]map[string]interface{}, error) {
	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok && val != nil {
		if s, ok := val.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// generateTempPassword generates a random temporary password
func generateTempPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	const length = 12

	password := make([]byte, length)
	for i := range password {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}

	return string(password)
}

// sendSuccessEmail sends welcome email with temporary password to a new member
func sendSuccessEmail(memberEmail, tempPassword string) error {
	cfg := config.GetConfig()

	if cfg.SMTPHost == "" {
		config.Logger.Warn("SMTP not configured, skipping registration email")
		return nil
	}

	subject := "Welcome to TAGA - Your Registration is Complete"
	loginURL := cfg.ResetPasswordURL + "/?page=member-login"

	var body strings.Builder
	body.WriteString("<h2>Welcome to TAGA!</h2>")
	body.WriteString("<p>Your registration has been successfully completed.</p>")
	body.WriteString("<h3>Login Credentials</h3>")
	body.WriteString(fmt.Sprintf("<p><strong>Email:</strong> %s</p>", memberEmail))
	body.WriteString(fmt.Sprintf("<p><strong>Temporary Password:</strong> %s</p>", tempPassword))
	body.WriteString("<p><strong>Important:</strong> You must change this password on your first login.</p>")
	body.WriteString(fmt.Sprintf("<p><a href='%s'>Click here to login</a></p>", loginURL))
	body.WriteString("<p>If you did not register for this account, please contact our support team.</p>")
	body.WriteString("<br><p>Best regards,<br>TAGA Team</p>")

	return sendEmail(memberEmail, subject, body.String())
}

// sendEmail sends an HTML email using SMTP configuration
func sendEmail(to, subject, body string) error {
	cfg := config.GetConfig()

	// ADD THIS LOG - Attempting to send email
	config.Logger.Info("📧 Sending email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("smtp_host", cfg.SMTPHost),
		zap.Int("smtp_port", cfg.SMTPPort),
	)

	from := cfg.FromEmail
	smtpHost := cfg.SMTPHost
	smtpPort := cfg.SMTPPort
	smtpUser := cfg.SMTPUsername
	smtpPass := cfg.SMTPPassword

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Setup authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		// ADD THIS ERROR LOG
		config.Logger.Error("❌ Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	// ADD THIS SUCCESS LOG
	config.Logger.Info("✅ Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	return nil
}

// getMemberTagaIdByEmail returns the tagaId for a given email
// If tagaId is empty, falls back to internal UUID (for old members without tagaId)
func getMemberTagaIdByEmail(email string) string {
	members, err := readExistingMembers()
	if err != nil {
		return ""
	}
	for _, m := range members {
		if mEmail, ok := m["emailId"].(string); ok && mEmail == email {
			// Try to get tagaId first (preferred)
			if tagaId, ok := m["tagaId"].(string); ok && tagaId != "" {
				return tagaId
			}
			// Fallback to internal UUID for old members
			if id, ok := m["id"].(string); ok {
				return id
			}
		}
	}
	return ""
}

// getMemberTagaIdByUUID returns the tagaId for a given internal UUID
// Used during login to convert JWT's UUID to tagaId for subscription lookup
func getMemberTagaIdByUUID(uuid string) string {
	members, err := readExistingMembers()
	if err != nil {
		return uuid // fallback to UUID if lookup fails
	}
	for _, m := range members {
		if id, ok := m["id"].(string); ok && id == uuid {
			if tagaId, ok := m["tagaId"].(string); ok && tagaId != "" {
				return tagaId
			}
			return uuid // no tagaId, use UUID
		}
	}
	return uuid
}
