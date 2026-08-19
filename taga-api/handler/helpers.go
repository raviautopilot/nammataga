// handler/helpers.go
package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html"
	"math/big"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"taga-api/config"

	"github.com/gin-gonic/gin"
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
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 12

	password := make([]byte, length)
	for i := range password {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}

	return string(password)
}

var emailMockMutex sync.Mutex

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
	body.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to TAGA</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #065f46 0%%, #047857 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Membership Confirmed
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">Welcome to TAGA!</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Your member registration has been successfully completed.</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                We are delighted to welcome you to the <strong>Tamil Nadu Agricultural Graduates Association (TAGA)</strong> community. Below are your initial login credentials to access the TAGA Portal.
            </p>

            <!-- Credential Card -->
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 24px;">
                <h3 style="margin: 0 0 14px 0; color: #065f46; font-size: 16px; font-weight: 700; display: flex; align-items: center;">
                    <span style="margin-right: 8px;">🔐</span> Your Login Credentials
                </h3>
                <table style="width: 100%%; border-collapse: collapse;">
                    <tr>
                        <td style="padding: 8px 0; color: #64748b; font-size: 14px; width: 40%%; border-bottom: 1px solid #f1f5f9;">Registered Email</td>
                        <td style="padding: 8px 0; font-weight: 600; color: #0f172a; border-bottom: 1px solid #f1f5f9; font-size: 14px;">%s</td>
                    </tr>
                    <tr>
                        <td style="padding: 8px 0; color: #64748b; font-size: 14px;">Temporary Password</td>
                        <td style="padding: 8px 0; font-family: monospace; font-size: 15px; font-weight: 700; color: #047857; letter-spacing: 0.05em;">%s</td>
                    </tr>
                </table>
            </div>

            <!-- Notice Box -->
            <div style="background: #fffbeb; border-left: 4px solid #f59e0b; padding: 14px 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0; font-size: 13px; color: #92400e; line-height: 1.5;">
                    <strong>Security Notice:</strong> For your security, you will be prompted to change your temporary password immediately upon your first login.
                </p>
            </div>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0 24px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #065f46 0%%, #047857 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(6, 95, 70, 0.25);">
                    Log In to Your Account &rarr;
                </a>
            </div>

            <p style="margin: 0; font-size: 13px; color: #6b7280; text-align: center;">
                If you did not register for this account, please immediately contact our administrative team.
            </p>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, html.EscapeString(memberEmail), html.EscapeString(tempPassword), loginURL))

	// For E2E testing: Save the temporary password to a mock emails file
	mockEmailFile := "data/emails/mock_emails.json"
	
	emailMockMutex.Lock()
	var mockEmails map[string]string
	data, err := os.ReadFile(mockEmailFile)
	if err == nil {
		json.Unmarshal(data, &mockEmails)
	}
	if mockEmails == nil {
		mockEmails = make(map[string]string)
	}
	mockEmails[memberEmail] = tempPassword
	if updatedData, err := json.MarshalIndent(mockEmails, "", "  "); err == nil {
		os.WriteFile(mockEmailFile, updatedData, 0644)
	}
	emailMockMutex.Unlock()

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

// getAdminUsername retrieves the admin username from the Gin context.
// Set by AdminAuthMiddleware from the JWT claims.
func getAdminUsername(c *gin.Context) string {
	if c == nil {
		return "admin"
	}
	if val, exists := c.Get("username"); exists {
		if s, ok := val.(string); ok && s != "" {
			return s
		}
	}
	return "admin"
}
