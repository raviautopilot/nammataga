package email

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"sync"
	"taga-api/config"
)

func SendPasswordResetEmail(to, resetToken string) error {
	cfg := config.Config
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
	cfg := config.Config
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

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)

	return smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		from,
		recipients,
		[]byte(message),
	)
}

var tempMockMutex sync.Mutex

// SendTemporaryPasswordEmail sends a temporary password to the user and requires them to change it on login.
func SendTemporaryPasswordEmail(to, tempPassword string) error {
	cfg := config.Config
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
