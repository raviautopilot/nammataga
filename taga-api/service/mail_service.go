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

	subject := "🔔 TAGA Portal - Pending Profile Edit Request"
	
	cardsHtml := ""
	for _, f := range fields {
		friendlyField := strings.ReplaceAll(f.Field, "_", " ")
		friendlyField = strings.Title(friendlyField)

		oldVal := f.OldValue
		if oldVal == "" {
			oldVal = "<span style='color:#cbd5e1;font-style:italic;'>Empty</span>"
		}

		memberRemarksHtml := ""
		if f.MemberRemarks != "" {
			memberRemarksHtml = fmt.Sprintf(`
				<div style="background: #fffbeb; border-left: 3px solid #f59e0b; padding: 10px 12px; border-radius: 4px; margin-top: 12px;">
                    <p style="margin: 0; font-size: 13px; color: #92400e;">
                        <strong style="text-transform: uppercase; font-size: 11px;">Member Note:</strong><br/>
						%s
                    </p>
                </div>
			`, f.MemberRemarks)
		} else {
			memberRemarksHtml = `
				<div style="background: #f8fafc; border-left: 3px solid #cbd5e1; padding: 10px 12px; border-radius: 4px; margin-top: 12px;">
                    <p style="margin: 0; font-size: 13px; color: #64748b;">
                        <strong style="text-transform: uppercase; font-size: 11px;">Member Note:</strong><br/>
						<i>None</i>
                    </p>
                </div>
			`
		}

		statusBg := "#f1f5f9"
		statusColor := "#475569"
		statusText := "PENDING"

		cardsHtml += fmt.Sprintf(`
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px; margin-bottom: 20px; box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);">
                <div style="margin-bottom: 16px;">
					<table width="100%%" border="0" cellpadding="0" cellspacing="0">
						<tr>
							<td align="left">
								<p style="margin: 0; color: #1e3a8a; font-size: 15px; text-transform: uppercase; font-weight: 800; letter-spacing: 0.05em;">%s</p>
							</td>
							<td align="right">
								<span style="background: %s; color: %s; padding: 6px 12px; border-radius: 20px; font-size: 11px; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em;">%s</span>
							</td>
						</tr>
					</table>
                </div>
                
                <div style="background: #ffffff; border: 1px solid #f1f5f9; border-radius: 8px; padding: 16px;">
					<table width="100%%" border="0" cellpadding="0" cellspacing="0">
						<tr>
							<td width="50%%" valign="top" style="padding-right: 10px; border-right: 1px solid #f1f5f9;">
								<p style="margin: 0 0 6px 0; font-size: 11px; color: #94a3b8; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em;">Previous Value</p>
                    			<p style="margin: 0; font-family: monospace; font-size: 14px; color: #64748b; text-decoration: line-through; word-break: break-all;">%s</p>
							</td>
							<td width="50%%" valign="top" style="padding-left: 14px;">
								<p style="margin: 0 0 6px 0; font-size: 11px; color: #94a3b8; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em;">Requested Value</p>
                    			<p style="margin: 0; font-family: monospace; font-size: 15px; font-weight: 700; color: #0f172a; word-break: break-all;">%s</p>
							</td>
						</tr>
					</table>
                </div>

				%s
            </div>
		`, friendlyField, statusBg, statusColor, statusText, oldVal, f.NewValue, memberRemarksHtml)
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
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">Profile Edit Request</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">A member wants to update their profile details.</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear Admin,
            </p>
            <p style="margin: 0 0 24px 0; font-size: 15px; color: #374151;">
                <strong>%s</strong> (%s) has submitted a request to edit their member profile. Please review the requested changes below:
            </p>

            %s

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    Go to Admin Dashboard &rarr;
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
</html>`, memberName, memberEmail, cardsHtml, adminURL)

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

	subject := "📝 TAGA Portal - Your Profile Update Results"
	
	cardsHtml := ""
	for _, f := range fields {
		statusColor := "#10b981"
		statusBg := "#ecfdf5"
		if f.Status == "rejected" {
			statusColor = "#ef4444"
			statusBg = "#fef2f2"
		}
		
		remarkBg := "#f0f9ff"
		remarkBorder := "#3b82f6"
		remarkText := "#1e40af"
		
		if f.Status == "rejected" {
			remarkBg = "#fef2f2"
			remarkBorder = "#ef4444"
			remarkText = "#991b1b"
		}
		
		remarkContent := f.AdminRemarks
		if remarkContent == "" {
			// Neutral grey style when empty
			remarkBg = "#f8fafc"
			remarkBorder = "#cbd5e1"
			remarkText = "#64748b"
			remarkContent = "<i>None</i>"
		}
		
		adminRemarksHtml := fmt.Sprintf(`
			<div style="background: %s; border-left: 3px solid %s; padding: 10px 12px; border-radius: 4px; margin-top: 12px;">
				<p style="margin: 0; font-size: 13px; color: %s;">
					<strong style="text-transform: uppercase; font-size: 11px;">Admin Remarks:</strong><br/>
					%s
				</p>
			</div>
		`, remarkBg, remarkBorder, remarkText, remarkContent)

		friendlyField := strings.ReplaceAll(f.Field, "_", " ")
		friendlyField = strings.Title(friendlyField)

		oldVal := f.OldValue
		if oldVal == "" {
			oldVal = "<span style='color:#cbd5e1;font-style:italic;'>Empty</span>"
		}

		cardsHtml += fmt.Sprintf(`
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 12px; padding: 20px; margin-bottom: 20px; box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);">
                <div style="margin-bottom: 16px;">
					<table width="100%%" border="0" cellpadding="0" cellspacing="0">
						<tr>
							<td align="left">
								<p style="margin: 0; color: #1e3a8a; font-size: 15px; text-transform: uppercase; font-weight: 800; letter-spacing: 0.05em;">%s</p>
							</td>
							<td align="right">
								<span style="background: %s; color: %s; padding: 6px 12px; border-radius: 20px; font-size: 11px; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em;">%s</span>
							</td>
						</tr>
					</table>
                </div>
                
                <div style="background: #ffffff; border: 1px solid #f1f5f9; border-radius: 8px; padding: 16px;">
					<table width="100%%" border="0" cellpadding="0" cellspacing="0">
						<tr>
							<td width="50%%" valign="top" style="padding-right: 10px; border-right: 1px solid #f1f5f9;">
								<p style="margin: 0 0 6px 0; font-size: 11px; color: #94a3b8; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em;">Previous Value</p>
                    			<p style="margin: 0; font-family: monospace; font-size: 14px; color: #64748b; text-decoration: line-through; word-break: break-all;">%s</p>
							</td>
							<td width="50%%" valign="top" style="padding-left: 14px;">
								<p style="margin: 0 0 6px 0; font-size: 11px; color: #94a3b8; text-transform: uppercase; font-weight: 700; letter-spacing: 0.05em;">Requested Value</p>
                    			<p style="margin: 0; font-family: monospace; font-size: 15px; font-weight: 700; color: #0f172a; word-break: break-all;">%s</p>
							</td>
						</tr>
					</table>
                </div>

				%s
            </div>
		`, friendlyField, statusBg, statusColor, f.Status, oldVal, f.NewValue, adminRemarksHtml)
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
                Profile Update
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 26px; font-weight: 800; letter-spacing: -0.02em;">Request Processed</h1>
            <p style="margin: 0; font-size: 15px; opacity: 0.9;">The administration has reviewed your profile edit request.</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                Dear <strong>%s</strong>,
            </p>
            <p style="margin: 0 0 24px 0; font-size: 15px; color: #374151;">
                Your recent request to update your TAGA member profile has been fully processed by our administrators. Please review the final decisions on each field below:
            </p>

            %s

            <!-- Mandatory Action Notice -->
            <div style="background: #fffbeb; border-left: 4px solid #f59e0b; padding: 14px 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0; font-size: 14px; color: #92400e; line-height: 1.5;">
                    <strong>Note:</strong> Approved changes are now live on your TAGA profile. If a request was rejected, please refer to the Admin Remarks for further guidance.
                </p>
            </div>

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0;">
                <a href="%s" style="background: linear-gradient(135deg, #1e3a8a 0%%, #2563eb 100%%); color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(37, 99, 235, 0.3);">
                    View Your Profile &rarr;
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
</html>`, memberName, cardsHtml, loginURL)

	msg := []byte("To: " + memberEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		htmlBody)

	return smtp.SendMail(fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort), auth, cfg.FromEmail, []string{memberEmail}, msg)
}
