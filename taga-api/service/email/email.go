package email

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"taga-api/config"
)

// SentEmailRecord captures outgoing emails in memory for verification and testing
type SentEmailRecord struct {
	To      string    `json:"to"`
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
	SentAt  time.Time `json:"sent_at"`
}

var (
	lastSentMutex  sync.RWMutex
	lastSentEmails []SentEmailRecord
)

// GetLastSentEmails returns a copy of recently recorded sent emails
func GetLastSentEmails() []SentEmailRecord {
	lastSentMutex.RLock()
	defer lastSentMutex.RUnlock()
	copied := make([]SentEmailRecord, len(lastSentEmails))
	copy(copied, lastSentEmails)
	return copied
}

// ClearLastSentEmails clears the recorded sent emails
func ClearLastSentEmails() {
	lastSentMutex.Lock()
	defer lastSentMutex.Unlock()
	lastSentEmails = nil
}

func SendPasswordResetEmail(to, resetToken string) error {
	cfg := config.GetConfig()
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", cfg.ResetPasswordURL, resetToken)

	subject := "🔒 TAGA Portal - Password Reset Request"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Reset Request</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Account Security
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">Password Reset Request</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Secure access recovery for your TAGA Portal account</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                We received a request to reset the password for your TAGA Member account associated with <strong>%s</strong>.
            </p>
            <p style="margin: 0 0 24px 0; font-size: 15px; color: #374151;">
                Click the button below to set a new password. This reset link is secure and will expire shortly.
            </p>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    Reset My Password &rarr;
                </a>
            </div>

            <!-- Fallback URL -->
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 24px;">
                <p style="margin: 0 0 8px 0; font-size: 12px; color: #64748b; font-weight: 600; text-transform: uppercase;">Button not working? Paste this link into your browser:</p>
                <p style="margin: 0; font-size: 13px; color: #2563eb; word-break: break-all;">%s</p>
            </div>

            <!-- Security Notice -->
            <div style="background: #fef2f2; border-left: 4px solid #ef4444; padding: 14px 16px; border-radius: 4px;">
                <p style="margin: 0; font-size: 13px; color: #991b1b; line-height: 1.5;">
                    <strong>Didn't request this?</strong> If you did not make this request, you can safely ignore this email. Your current password will remain unchanged.
                </p>
            </div>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, to, resetURL, resetURL)

	return SendHTMLEmail(to, subject, htmlBody)
}

// SendHTMLEmail sends an HTML formatted email.
func SendHTMLEmail(to, subject, htmlBody string) error {
	cfg := config.GetConfig()
	from := cfg.FromEmail
	if from == "" {
		from = cfg.SMTPUsername
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("Nammataga Association <%s>", from)
	headers["To"] = to
	if cfg.AdminEmail != "" {
		headers["Reply-To"] = fmt.Sprintf("Nammataga Association <%s>", cfg.AdminEmail)
	}
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	recipients := []string{to}
	if cfg.CCEmail != "" && cfg.CCEmail != to {
		headers["Cc"] = cfg.CCEmail
		recipients = append(recipients, cfg.CCEmail)
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Ensure any embedded spaces in Gmail App Passwords are clean stripped
	smtpPass := strings.ReplaceAll(cfg.SMTPPassword, " ", "")
	smtpHost := cfg.SMTPHost
	smtpPort := cfg.SMTPPort
	smtpUser := cfg.SMTPUsername

	// Record outgoing email for test verifications
	lastSentMutex.Lock()
	lastSentEmails = append(lastSentEmails, SentEmailRecord{
		To:      to,
		Subject: subject,
		Body:    htmlBody,
		SentAt:  time.Now(),
	})
	lastSentMutex.Unlock()

	if config.Logger != nil {
		config.Logger.Info("📧 Sending email via SendHTMLEmail",
			zap.String("to", to),
			zap.String("from", from),
			zap.String("subject", subject),
			zap.String("smtp_host", smtpHost),
			zap.Int("smtp_port", smtpPort),
		)
	}

	// In test environments or when SMTP host/port/user is unconfigured, avoid network failure
	if smtpHost == "" || smtpPort == 0 || smtpUser == "" {
		if config.Logger != nil {
			config.Logger.Warn("⚠️ SMTP configuration is unconfigured or incomplete; email recorded in memory but skipping network send",
				zap.String("to", to),
				zap.String("subject", subject),
			)
		}
		return nil
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

	err := smtp.SendMail(addr, auth, from, recipients, []byte(message))
	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("❌ Failed to send email via SendHTMLEmail",
				zap.String("to", to),
				zap.String("subject", subject),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Info("✅ Email sent successfully via SendHTMLEmail",
			zap.String("to", to),
			zap.String("subject", subject),
		)
	}

	return nil
}

var tempMockMutex sync.Mutex

// SendTemporaryPasswordEmail sends a temporary password to the user and requires them to change it on login.
func SendTemporaryPasswordEmail(to, tempPassword string) error {
	cfg := config.GetConfig()
	loginURL := cfg.ResetPasswordURL + "/?page=member-login"

	subject := "🔒 TAGA Portal - Your Temporary Password"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Your Temporary Password</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Account Recovery
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">Temporary Password</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">We've generated a secure temporary password for you.</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear Member,
            </p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                As per your request, we have reset your TAGA Member account password. Please find your temporary login credentials below:
            </p>

            <!-- Credential Card -->
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 24px; text-align: center;">
                <p style="margin: 0 0 8px 0; color: #64748b; font-size: 14px; text-transform: uppercase; font-weight: 600;">Temporary Password</p>
                <p style="margin: 0; font-family: monospace; font-size: 24px; font-weight: 700; color: #2563eb; letter-spacing: 0.1em; background: #eff6ff; padding: 12px; border-radius: 6px; display: inline-block;">%s</p>
            </div>

            <!-- Mandatory Action Notice -->
            <div style="background: #fffbeb; border-left: 4px solid #f59e0b; padding: 14px 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0; font-size: 14px; color: #92400e; line-height: 1.5;">
                    <strong>Action Required:</strong> For your security, you must change this password before logging in. Please visit the Member Login page, click the <strong>Change Password</strong> button, and enter this temporary password into the 'Old Password' field to set a permanent, secure password.
                </p>
            </div>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    Go to Member Login &rarr;
                </a>
            </div>
            
            <p style="margin: 0; font-size: 14px; color: #6b7280; text-align: center;">
                If you did not request this password reset, please contact the TAGA administrative team immediately.
            </p>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, tempPassword, loginURL)

	// Save to mock emails for E2E tests
	mockEmailFile := "data/emails/mock_emails.json"
	tempMockMutex.Lock()
	var mockEmails map[string]string
	importJson := true
	if data, err := os.ReadFile(mockEmailFile); err == nil {
		if err := json.Unmarshal(data, &mockEmails); err != nil {
			importJson = false
		}
	}
	_ = importJson
	if mockEmails == nil {
		mockEmails = make(map[string]string)
	}
	mockEmails[to] = tempPassword
	if updatedData, err := json.MarshalIndent(mockEmails, "", "  "); err == nil {
		_ = os.WriteFile(mockEmailFile, updatedData, 0644)
	}
	tempMockMutex.Unlock()

	return SendHTMLEmail(to, subject, htmlBody)
}

// SendAccountDeletionRequestAcknowledgement sends an acknowledgement email to a member when their deletion request is submitted.
func SendAccountDeletionRequestAcknowledgement(to, memberName, tagaID, reason, requestedAt string) error {
	if to == "" {
		return nil
	}
	if memberName == "" {
		memberName = "Member"
	}
	if tagaID == "" {
		tagaID = "N/A"
	}
	if requestedAt == "" {
		requestedAt = time.Now().Format("02 Jan 2006, 03:04 PM")
	} else if t, err := time.Parse(time.RFC3339, requestedAt); err == nil {
		requestedAt = t.Format("02 Jan 2006, 03:04 PM")
	}

	subject := "⏳ TAGA Portal - Account Deletion Request Received"

	reasonBlock := ""
	if strings.TrimSpace(reason) != "" {
		reasonBlock = fmt.Sprintf(`
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 24px;">
                <p style="margin: 0 0 6px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em;">Stated Reason for Deletion:</p>
                <p style="margin: 0; font-size: 14px; color: #334155; line-height: 1.5; white-space: pre-wrap;">%s</p>
            </div>
        `, reason)
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Account Deletion Request Received</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Deletion Request
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 800; letter-spacing: -0.02em;">Account Deletion Request Received</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Your request has been received and is under administrator review</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear <strong>%s</strong> (TAGA ID: %s),
            </p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                We have received your formal request to delete your TAGA Member account. Your request has been queued in our administrative system and is currently pending review by an association administrator.
            </p>

            <!-- Info Card -->
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px; margin-bottom: 24px;">
                <h3 style="margin: 0 0 16px 0; font-size: 14px; color: #1e3a8a; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em; border-bottom: 1px solid #e2e8f0; padding-bottom: 10px;">Request Details</h3>
                <table width="100%%" border="0" cellpadding="0" cellspacing="0">
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 12px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Member Name</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 600;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 12px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">TAGA ID</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 600;">%s</p>
                        </td>
                    </tr>
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 4px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Registered Email</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 4px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Submitted At</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                    </tr>
                </table>
            </div>

            %s

            <!-- Security Notice -->
            <div style="background: #fffbeb; border-left: 4px solid #f59e0b; padding: 14px 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 700; color: #92400e; text-transform: uppercase; letter-spacing: 0.05em;">Account &amp; Security Notice:</p>
                <p style="margin: 0; font-size: 13px; color: #92400e; line-height: 1.5;">
                    While your request is under review, access to your member dashboard is paused. You will receive an official notification email once an administrator acts on your request. If you did not make this request, please contact the TAGA administration immediately at <a href="mailto:nammataga@gmail.com" style="color: #b45309; font-weight: 600;">nammataga@gmail.com</a>.
                </p>
            </div>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, memberName, tagaID, memberName, tagaID, to, requestedAt, reasonBlock)

	return SendHTMLEmail(to, subject, htmlBody)
}

// SendAccountDeletionRequestAdminAlert sends an alert notification email to the administrator when a member submits a deletion request.
func SendAccountDeletionRequestAdminAlert(adminEmail, memberName, tagaID, memberEmail, mobileNumber, reason, requestedAt string) error {
	if adminEmail == "" {
		cfg := config.GetConfig()
		adminEmail = cfg.AdminEmail
		if adminEmail == "" {
			adminEmail = "nammataga@gmail.com"
		}
	}
	if memberName == "" {
		memberName = "Member"
	}
	if tagaID == "" {
		tagaID = "N/A"
	}
	if mobileNumber == "" {
		mobileNumber = "N/A"
	}
	if requestedAt == "" {
		requestedAt = time.Now().Format("02 Jan 2006, 03:04 PM")
	} else if t, err := time.Parse(time.RFC3339, requestedAt); err == nil {
		requestedAt = t.Format("02 Jan 2006, 03:04 PM")
	}

	cfg := config.GetConfig()
	adminURL := cfg.ResetPasswordURL + "/?page=admin-login"

	subject := fmt.Sprintf("🔔 [ADMIN] New Account Deletion Request - %s (%s)", memberName, tagaID)

	reasonText := strings.TrimSpace(reason)
	if reasonText == "" {
		reasonText = "<i>No specific reason provided by member.</i>"
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Account Deletion Request</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #991b1b 0%%, #dc2626 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Admin Action Required
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 800; letter-spacing: -0.02em;">Account Deletion Request</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">A member has submitted a request for account deletion</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear Administrator,
            </p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                A new member account retirement/deletion request has been submitted and is awaiting your review. Please examine the member details below:
            </p>

            <!-- Member Details Card -->
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px; margin-bottom: 24px;">
                <h3 style="margin: 0 0 16px 0; font-size: 14px; color: #991b1b; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em; border-bottom: 1px solid #e2e8f0; padding-bottom: 10px;">Member Information</h3>
                
                <table width="100%%" border="0" cellpadding="0" cellspacing="0">
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 14px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Member Name</p>
                            <p style="margin: 0; font-size: 15px; color: #0f172a; font-weight: 700;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 14px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">TAGA ID</p>
                            <p style="margin: 0; font-size: 15px; color: #0f172a; font-weight: 700;">%s</p>
                        </td>
                    </tr>
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 14px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Registered Email</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 14px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Mobile Number</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                    </tr>
                    <tr>
                        <td colspan="2" valign="top">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Requested At</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                    </tr>
                </table>
            </div>

            <!-- Stated Reason Card -->
            <div style="background: #ffffff; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 24px;">
                <p style="margin: 0 0 6px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em;">Stated Reason for Deletion:</p>
                <p style="margin: 0; font-size: 14px; color: #334155; line-height: 1.5; white-space: pre-wrap;">%s</p>
            </div>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    Review in Admin Portal &rarr;
                </a>
            </div>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, memberName, tagaID, memberEmail, mobileNumber, requestedAt, reasonText, adminURL)

	return SendHTMLEmail(adminEmail, subject, htmlBody)
}

// SendAccountDeletionConfirmation sends a confirmation email when a member's account has been permanently deleted.
func SendAccountDeletionConfirmation(to, memberName, tagaID string) error {
	if to == "" {
		return nil
	}
	if memberName == "" {
		memberName = "Member"
	}
	if tagaID == "" {
		tagaID = "N/A"
	}

	subject := "🗑️ TAGA Portal - Account Deletion Confirmation"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Account Deletion Confirmation</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #991b1b 0%%, #dc2626 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Account Retired
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 800; letter-spacing: -0.02em;">Account Deletion Confirmation</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Your TAGA account has been permanently removed</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear <strong>%s</strong> (TAGA ID: %s),
            </p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                This email confirms that your TAGA Member account has been permanently deleted in accordance with your request and administrative approval.
            </p>
            <div style="background: #fef2f2; border-left: 4px solid #ef4444; padding: 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0 0 8px 0; font-size: 14px; font-weight: 600; color: #991b1b;">Summary of Actions Taken:</p>
                <ul style="margin: 0; padding-left: 20px; font-size: 13px; color: #7f1d1d; line-height: 1.6;">
                    <li>Your login access, passwords, and active session tokens have been permanently revoked.</li>
                    <li>Active annual subscriptions have been marked as cancelled.</li>
                    <li>Any upcoming room reservations at TAGA Towers have been released for other members.</li>
                    <li>Financial transaction records are preserved intact for statutory accounting and audit compliance.</li>
                </ul>
            </div>
            <p style="margin: 0; font-size: 14px; color: #6b7280; text-align: center;">
                If you believe this was done in error or if you have questions regarding past association activities, please contact the TAGA administrator at <a href="mailto:nammataga@gmail.com" style="color: #dc2626; text-decoration: underline;">nammataga@gmail.com</a>.
            </p>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, memberName, tagaID)

	return SendHTMLEmail(to, subject, htmlBody)
}

// SendAccountDeletionRejection sends an email when an account deletion request is rejected by an administrator.
func SendAccountDeletionRejection(to, memberName, tagaID, remarks string) error {
	if to == "" {
		return nil
	}
	if memberName == "" {
		memberName = "Member"
	}
	if tagaID == "" {
		tagaID = "N/A"
	}

	cfg := config.GetConfig()
	loginURL := cfg.ResetPasswordURL + "/?page=member-login"

	subject := "🛡️ TAGA Portal - Account Deletion Request Update"
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Account Deletion Request Update</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Account Active
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 800; letter-spacing: -0.02em;">Account Deletion Request Update</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">Review results for your account deletion request</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear <strong>%s</strong> (TAGA ID: %s),
            </p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                Your account deletion request has been reviewed by the association administrator and was <strong>not approved</strong>.
            </p>
            
            <!-- Mandatory Remarks Box -->
            <div style="background: #fffbeb; border-left: 4px solid #f59e0b; padding: 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 700; color: #92400e; text-transform: uppercase; letter-spacing: 0.05em;">Administrator Remarks:</p>
                <p style="margin: 0; font-size: 14px; color: #78350f; line-height: 1.5; white-space: pre-wrap;">%s</p>
            </div>
            
            <p style="margin: 0 0 24px 0; font-size: 15px; color: #374151;">
                Your account remains fully active, and you can now log in normally to access your member dashboard, subscriptions, and TAGA Towers booking services.
            </p>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    Go to Member Login &rarr;
                </a>
            </div>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; &copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, memberName, tagaID, remarks, loginURL)

	return SendHTMLEmail(to, subject, htmlBody)
}

