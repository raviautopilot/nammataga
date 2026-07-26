package email

import (
	"fmt"
	"net/smtp"
	"taga-api/config"
)

func SendPasswordResetEmail(to, resetToken string) error {
	cfg := config.Config
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", cfg.ResetPasswordURL, resetToken)

	subject := "Password Reset Request"
	body := fmt.Sprintf(`Hello,

We received a request to reset your password. Click the link below to reset your password:

%s

If you didn't request this, please ignore this email.

Best regards,
Taga Team`, resetURL)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.FromEmail,
		[]string{to},
		msg,
	)
	return err
}

// SendHTMLEmail sends an HTML formatted email.
func SendHTMLEmail(to, subject, htmlBody string) error {
	cfg := config.Config
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlBody)

	return smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.FromEmail,
		[]string{to},
		msg,
	)
}
