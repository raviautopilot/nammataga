package service

import (
	"fmt"
	"net/smtp"
	"strings"

	"taga-api/config"
	"taga-api/model"
)

func SendAdminEditRequestEmail(memberEmail, memberName string, fields []model.FieldEditRequest) error {
	cfg := config.Config
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	adminURL := cfg.ResetPasswordURL + "/?page=admin-login"

	subject := "TAGA Admin: Pending Profile Edit"
	
	rowsHtml := ""
	for _, f := range fields {
		memberRemarksHtml := ""
		if f.MemberRemarks != "" {
			memberRemarksHtml = fmt.Sprintf(`
				<tr>
					<td colspan="3" style="padding: 0 12px 16px 12px; border-bottom: 1px solid #eaeaea;">
						<div style="background-color: #fffbeb; padding: 12px; border-radius: 6px; font-size: 13px; color: #92400e; border-left: 3px solid #f59e0b;">
							<strong>Member Note:</strong> %s
						</div>
					</td>
				</tr>
			`, f.MemberRemarks)
		} else {
			memberRemarksHtml = `
				<tr>
					<td colspan="3" style="padding: 0; border-bottom: 1px solid #eaeaea; height: 16px;"></td>
				</tr>
			`
		}

		friendlyField := strings.ReplaceAll(f.Field, "_", " ")
		friendlyField = strings.Title(friendlyField)

		oldVal := f.OldValue
		if oldVal == "" {
			oldVal = "Empty"
		}

		rowsHtml += fmt.Sprintf(`
			<tr>
				<td style="padding: 16px 12px 8px 12px; color: #111827; font-weight: 500; font-size: 14px;">%s</td>
				<td style="padding: 16px 12px 8px 12px; color: #6b7280; font-family: ui-monospace, monospace; font-size: 13px; text-decoration: line-through;">%s</td>
				<td style="padding: 16px 12px 8px 12px; color: #111827; font-family: ui-monospace, monospace; font-size: 14px; font-weight: 600;">%s</td>
			</tr>
			%s
		`, friendlyField, oldVal, f.NewValue, memberRemarksHtml)
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #111827; background-color: #f9fafb; margin: 0; padding: 40px 20px;">
    <div style="max-width: 640px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; border: 1px solid #e5e7eb; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);">
        
        <div style="padding: 40px 40px 24px 40px; border-bottom: 1px solid #eaeaea;">
            <div style="display: inline-block; background-color: #fee2e2; color: #b91c1c; padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; margin-bottom: 12px; text-transform: uppercase;">Action Required</div>
            <h1 style="margin: 0; font-size: 24px; font-weight: 600; letter-spacing: -0.02em; color: #111827;">Pending Edit Request</h1>
        </div>

        <div style="padding: 32px 40px 40px 40px;">
            
            <div style="margin-bottom: 32px; padding: 16px; background-color: #f9fafb; border-radius: 8px; border: 1px solid #eaeaea;">
                <p style="margin: 0; font-size: 13px; color: #6b7280; text-transform: uppercase; font-weight: 600; letter-spacing: 0.05em;">Member Profile</p>
                <p style="margin: 4px 0 0 0; font-size: 16px; font-weight: 600; color: #111827;">%s</p>
                <p style="margin: 0; font-size: 14px; color: #4b5563;">%s</p>
            </div>

            <table style="width: 100%%; border-collapse: collapse; margin-bottom: 32px;">
                <thead>
                    <tr>
                        <th style="text-align: left; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Field</th>
                        <th style="text-align: left; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Current</th>
                        <th style="text-align: left; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Requested</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>

            <div style="margin-top: 32px;">
                <a href="%s" style="background-color: #111827; color: #ffffff; padding: 12px 24px; font-size: 14px; font-weight: 500; text-decoration: none; border-radius: 6px; display: inline-block;">
                    Open Admin Panel
                </a>
            </div>
        </div>
        
        <div style="padding: 24px 40px; background-color: #f9fafb; border-top: 1px solid #eaeaea; border-bottom-left-radius: 12px; border-bottom-right-radius: 12px;">
            <p style="margin: 0; font-size: 13px; color: #6b7280;">TAGA System Administrator Notifications</p>
        </div>
    </div>
</body>
</html>`, memberName, memberEmail, rowsHtml, adminURL)

	msg := []byte("To: " + cfg.AdminEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlBody)

	return smtp.SendMail(fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.FromEmail, []string{cfg.AdminEmail}, msg)
}

func SendMemberRequestProcessedEmail(memberEmail, memberName string, fields []model.FieldEditRequest) error {
	cfg := config.Config
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	loginURL := cfg.ResetPasswordURL + "/?page=member-login"

	subject := "TAGA Profile Update Processed"
	
	rowsHtml := ""
	for _, f := range fields {
		statusColor := "#10b981"
		statusText := "APPROVED"
		if f.Status == "rejected" {
			statusColor = "#ef4444"
			statusText = "REJECTED"
		}
		
		adminRemarksHtml := ""
		if f.AdminRemarks != "" {
			adminRemarksHtml = fmt.Sprintf(`
				<tr>
					<td colspan="4" style="padding: 0 12px 16px 12px; border-bottom: 1px solid #eaeaea;">
						<div style="background-color: #f9fafb; padding: 12px; border-radius: 6px; font-size: 13px; color: #4b5563;">
							<strong style="color: #111827;">Admin Note:</strong> %s
						</div>
					</td>
				</tr>
			`, f.AdminRemarks)
		} else {
			// If no remarks, just add a simple bottom border row
			adminRemarksHtml = `
				<tr>
					<td colspan="4" style="padding: 0; border-bottom: 1px solid #eaeaea; height: 16px;"></td>
				</tr>
			`
		}

		friendlyField := strings.ReplaceAll(f.Field, "_", " ")
		friendlyField = strings.Title(friendlyField)

		oldVal := f.OldValue
		if oldVal == "" {
			oldVal = "Empty"
		}

		rowsHtml += fmt.Sprintf(`
			<tr>
				<td style="padding: 16px 12px 8px 12px; color: #111827; font-weight: 500; font-size: 14px;">%s</td>
				<td style="padding: 16px 12px 8px 12px; color: #6b7280; font-family: ui-monospace, monospace; font-size: 13px; text-decoration: line-through;">%s</td>
				<td style="padding: 16px 12px 8px 12px; color: #111827; font-family: ui-monospace, monospace; font-size: 14px; font-weight: 600;">%s</td>
				<td style="padding: 16px 12px 8px 12px; text-align: right;">
					<span style="color: %s; font-size: 12px; font-weight: 700; letter-spacing: 0.05em;">%s</span>
				</td>
			</tr>
			%s
		`, friendlyField, oldVal, f.NewValue, statusColor, statusText, adminRemarksHtml)
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #111827; background-color: #f9fafb; margin: 0; padding: 40px 20px;">
    <div style="max-width: 640px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; border: 1px solid #e5e7eb; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);">
        
        <!-- Minimalist Header -->
        <div style="padding: 40px 40px 24px 40px; border-bottom: 1px solid #eaeaea;">
            <h1 style="margin: 0; font-size: 24px; font-weight: 600; letter-spacing: -0.02em; color: #111827;">Profile Update Processed</h1>
        </div>

        <div style="padding: 32px 40px 40px 40px;">
            <p style="margin: 0 0 24px 0; font-size: 15px; color: #4b5563;">
                Hi %s,<br><br>
                The TAGA administration has reviewed your recent profile edit request. The final decisions for your requested changes are detailed below.
            </p>

            <table style="width: 100%%; border-collapse: collapse; margin-bottom: 32px;">
                <thead>
                    <tr>
                        <th style="text-align: left; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Field</th>
                        <th style="text-align: left; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Previous</th>
                        <th style="text-align: left; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Requested</th>
                        <th style="text-align: right; padding: 0 12px 12px 12px; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Status</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>

            <!-- Simple sleek button -->
            <div style="margin-top: 32px;">
                <a href="%s" style="background-color: #111827; color: #ffffff; padding: 12px 24px; font-size: 14px; font-weight: 500; text-decoration: none; border-radius: 6px; display: inline-block;">
                    View Profile
                </a>
            </div>
        </div>
        
        <div style="padding: 24px 40px; background-color: #f9fafb; border-top: 1px solid #eaeaea; border-bottom-left-radius: 12px; border-bottom-right-radius: 12px;">
            <p style="margin: 0; font-size: 13px; color: #6b7280;">TAGA Towers, Chennai &bull; Tamil Nadu Agricultural Graduates Association</p>
        </div>
    </div>
</body>
</html>`, memberName, rowsHtml, loginURL)

	msg := []byte("To: " + memberEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlBody)

	return smtp.SendMail(fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.FromEmail, []string{memberEmail}, msg)
}
