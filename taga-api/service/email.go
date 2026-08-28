package service

import (
	"fmt"
	"net/smtp"

	"taga-api/config"
)

func SendGrievanceEmail(
	subject string,
	category string,
	priority string,
	description string,
	phone string,
	memberName string,
	memberEmail string,
	preferredResponse string,
) error {

	cfg := config.Config
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	adminURL := cfg.ResetPasswordURL + "/?page=admin-login"

	if memberName == "" {
		memberName = "A Member"
	}
	if memberEmail == "" {
		memberEmail = "Unknown Email"
	}

	priorityColor := "#3b82f6" // blue for low
	switch priority {
	case "High":
		priorityColor = "#ef4444" // red
	case "Medium":
		priorityColor = "#f59e0b" // orange
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: rgba(255, 255, 255, 0.2); margin-bottom: 12px;">
                Admin Notification
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">New Grievance Filed</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">A member has submitted a new grievance report.</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 24px 0; font-size: 15px; color: #374151;">
                Dear Admin, a new grievance has been logged in the system. Please review the details below:
            </p>

            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px; margin-bottom: 24px;">
                <h2 style="margin: 0 0 16px 0; font-size: 18px; color: #1e3a8a; border-bottom: 1px solid #e2e8f0; padding-bottom: 12px;">%s</h2>
                
                <table width="100%%" border="0" cellpadding="0" cellspacing="0" style="margin-bottom: 16px;">
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 16px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Member Name</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 16px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Member Email</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                    </tr>
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 16px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Category</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 16px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Priority</p>
                            <span style="background: %s20; color: %s; padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: 700;">%s</span>
                        </td>
                    </tr>
                    <tr>
                        <td width="50%%" valign="top" style="padding-bottom: 16px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Contact Phone</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                        <td width="50%%" valign="top" style="padding-bottom: 16px;">
                            <p style="margin: 0 0 4px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Preferred Response</p>
                            <p style="margin: 0; font-size: 14px; color: #0f172a; font-weight: 500;">%s</p>
                        </td>
                    </tr>
                </table>

                <div style="background: #ffffff; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px;">
                    <p style="margin: 0 0 8px 0; font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 700;">Description</p>
                    <p style="margin: 0; font-size: 14px; color: #334155; white-space: pre-wrap;">%s</p>
                </div>
            </div>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    View in Admin Panel &rarr;
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
</html>`, subject, memberName, memberEmail, category, priorityColor, priorityColor, priority, phone, preferredResponse, description, adminURL)

	msg := []byte("To: " + cfg.AdminEmail + "\r\n" +
		"Subject: =?UTF-8?B?8J+SoCBUQUdBIFBvcnRhbCAtIE5ldyBHcmlldmFuY2UgU3VibWl0dGVk?=\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlBody)

	return smtp.SendMail(fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.FromEmail, []string{cfg.AdminEmail}, msg)
}
