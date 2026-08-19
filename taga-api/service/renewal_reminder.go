package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"taga-api/config"
	"taga-api/model"
	"taga-api/service/email"
	"taga-api/utils"
	"time"

	"go.uber.org/zap"
)

// ReminderType identifiers for each email in the sequence.
const (
	Reminder1 = "april_1"  // Membership year started
	Reminder2 = "april_15" // Mid-month reminder
	Reminder3 = "may_1"    // Final reminder before grace ends
)

type RenewalReminderLog struct {
	MemberEmail  string    `json:"member_email"`
	ReminderType string    `json:"reminder_type"`
	SentAt       time.Time `json:"sent_at"`
	Year         int       `json:"year"` // membership year (e.g., 2027 for 2027-2028)
}

// getReminderDate returns the target date for a given reminder type in the current membership year.
func getReminderDate(reminderType string, now time.Time) time.Time {
	year := now.Year()
	// Determine membership year start: if now is before April, membership year started last year.
	if now.Month() < 4 {
		year-- // previous membership year
	}
	switch reminderType {
	case Reminder1:
		return time.Date(year, 4, 1, 0, 0, 0, 0, now.Location())
	case Reminder2:
		return time.Date(year, 4, 15, 0, 0, 0, 0, now.Location())
	case Reminder3:
		return time.Date(year, 5, 1, 0, 0, 0, 0, now.Location())
	}
	return time.Time{}
}

// isPaidForCurrentYear checks if a member has an active annual subscription for the current membership year.
func isPaidForCurrentYear(email string) bool {
	filePath := filepath.Join("data", "subscriptions", "member_subscriptions.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	var subscriptions []model.MemberSubscription
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		return false
	}

	now := time.Now()
	yearEnd := utils.MembershipYearEnd(now)

	for _, sub := range subscriptions {
		if sub.MemberEmail == email &&
			sub.SubscriptionID == "annual-subscription" &&
			sub.Status == "active" &&
			sub.EndDate.After(now) && sub.EndDate.Year() == yearEnd.Year() {
			return true
		}
	}
	return false
}

// loadRenewalLog loads the sent reminders from JSON.
func loadRenewalLog() ([]RenewalReminderLog, error) {
	logPath := filepath.Join("data", "renewal_reminders", "renewal_log.json")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	var logs []RenewalReminderLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// saveRenewalLog saves the log entry.
func saveRenewalLog(entry RenewalReminderLog) error {
	logPath := filepath.Join("data", "renewal_reminders", "renewal_log.json")
	logs, err := loadRenewalLog()
	if err != nil {
		logs = []RenewalReminderLog{} // start fresh if file missing
	}
	logs = append(logs, entry)
	data, _ := json.MarshalIndent(logs, "", "  ")
	return os.WriteFile(logPath, data, 0644)
}

// alreadySent checks if a reminder of given type has already been sent to this member for the current membership year.
func alreadySent(email, reminderType string, year int) bool {
	logs, err := loadRenewalLog()
	if err != nil {
		return false
	}
	for _, l := range logs {
		if l.MemberEmail == email && l.ReminderType == reminderType && l.Year == year {
			return true
		}
	}
	return false
}

// getUnpaidMembers returns a list of member emails who have not paid.
func getUnpaidMembers() ([]string, error) {
	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		return nil, err
	}
	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return nil, err
	}

	var unpaid []string
	for _, m := range members {
		email, ok := m["emailId"].(string)
		if !ok || email == "" {
			continue
		}
		if !isPaidForCurrentYear(email) {
			unpaid = append(unpaid, email)
		}
	}
	return unpaid, nil
}

// getAnnualSubscriptionAmount fetches the current annual subscription fee from subscriptions.json
func getAnnualSubscriptionAmount() int {
	subsFile := config.Config.Data.Config.SubscriptionType
	data, err := os.ReadFile(subsFile)
	if err != nil {
		return 3500 // Fallback default
	}
	var subscriptions []map[string]interface{}
	if err := json.Unmarshal(data, &subscriptions); err == nil {
		for _, sub := range subscriptions {
			if id, ok := sub["id"].(string); ok && id == "annual-subscription" {
				if amt, ok := sub["amount"].(float64); ok {
					return int(amt)
				}
			}
		}
	}
	return 3500
}

// buildEmailHTML returns a premium HTML email for the given reminder type.
func buildEmailHTML(reminderType, memberName string) string {
	annualAmount := getAnnualSubscriptionAmount()
	yearStart := time.Now().Year()
	yearEnd := yearStart + 1

	var badgeText, badgeColor, headerGradient, actionBtnBg, customContent string

	switch reminderType {
	case Reminder1:
		badgeText = "MEMBERSHIP RENEWAL"
		badgeColor = "rgba(255, 255, 255, 0.2)"
		headerGradient = "linear-gradient(135deg, #065f46 0%, #047857 100%)"
		actionBtnBg = "linear-gradient(135deg, #065f46 0%, #047857 100%)"
		customContent = fmt.Sprintf(`
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">Dear <strong>%s</strong>,</p>
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">Warm greetings from <strong>TAGA</strong>! 🌾</p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                The new membership year <strong>(April 1, %d – March 31, %d)</strong> has commenced. To keep your membership active and continue enjoying all association benefits and services, kindly renew your Annual Subscription.
            </p>
            
            <div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 24px;">
                <table style="width: 100%%; border-collapse: collapse;">
                    <tr>
                        <td style="padding: 6px 0; color: #64748b; font-size: 14px;">Renewal Period</td>
                        <td style="padding: 6px 0; font-weight: 600; color: #0f172a; text-align: right; font-size: 14px;">%d – %d</td>
                    </tr>
                    <tr>
                        <td style="padding: 6px 0; color: #64748b; font-size: 14px;">Annual Fee</td>
                        <td style="padding: 6px 0; font-weight: 700; color: #065f46; text-align: right; font-size: 16px;">₹ %d</td>
                    </tr>
                    <tr>
                        <td style="padding: 6px 0; color: #64748b; font-size: 14px;">Grace Period Ends</td>
                        <td style="padding: 6px 0; font-weight: 600; color: #d97706; text-align: right; font-size: 14px;">May 31, %d</td>
                    </tr>
                </table>
            </div>`, memberName, yearStart, yearEnd, yearStart, yearEnd, annualAmount, yearStart)

	case Reminder2:
		badgeText = "REMINDER: SUBSCRIPTION PENDING"
		badgeColor = "rgba(255, 255, 255, 0.25)"
		headerGradient = "linear-gradient(135deg, #d97706 0%, #b45309 100%)"
		actionBtnBg = "linear-gradient(135deg, #d97706 0%, #b45309 100%)"
		customContent = fmt.Sprintf(`
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">Dear <strong>%s</strong>,</p>
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                We noticed that your TAGA Annual Subscription for the period <strong>(April 1, %d – March 31, %d)</strong> is still pending.
            </p>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                This is a friendly reminder to complete your renewal before the grace period expires on <strong>May 31</strong> to avoid any interruption in membership privileges.
            </p>
            
            <div style="background: #fffbeb; border: 1px solid #fef3c7; border-radius: 8px; padding: 20px; margin-bottom: 24px;">
                <table style="width: 100%%; border-collapse: collapse;">
                    <tr>
                        <td style="padding: 6px 0; color: #92400e; font-size: 14px;">Annual Fee Due</td>
                        <td style="padding: 6px 0; font-weight: 700; color: #b45309; text-align: right; font-size: 16px;">₹ %d</td>
                    </tr>
                    <tr>
                        <td style="padding: 6px 0; color: #92400e; font-size: 14px;">Final Grace Date</td>
                        <td style="padding: 6px 0; font-weight: 700; color: #dc2626; text-align: right; font-size: 14px;">May 31, %d</td>
                    </tr>
                </table>
            </div>`, memberName, yearStart, yearEnd, annualAmount, yearStart)

	case Reminder3:
		badgeText = "FINAL NOTICE: ACTION REQUIRED"
		badgeColor = "rgba(255, 255, 255, 0.3)"
		headerGradient = "linear-gradient(135deg, #dc2626 0%, #991b1b 100%)"
		actionBtnBg = "linear-gradient(135deg, #dc2626 0%, #991b1b 100%)"
		customContent = fmt.Sprintf(`
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">Dear <strong>%s</strong>,</p>
            <p style="margin: 0 0 16px 0; font-size: 15px; color: #374151;">
                This is the <strong>Final Notice</strong> regarding your TAGA Annual Subscription renewal for <strong>%d‑%d</strong>.
            </p>
            <div style="background: #fef2f2; border-left: 4px solid #ef4444; padding: 16px; border-radius: 4px; margin-bottom: 24px;">
                <p style="margin: 0; font-size: 14px; color: #991b1b; line-height: 1.5;">
                    After <strong>May 31</strong>, accounts without active subscriptions will be marked as <strong style="text-decoration: underline;">Inactive</strong> and will lose access to member directory, booking privileges, and association services.
                </p>
            </div>
            <p style="margin: 0 0 20px 0; font-size: 15px; color: #374151;">
                Please take a moment right now to complete your renewal and maintain uninterrupted membership.
            </p>`, memberName, yearStart, yearEnd)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TAGA Membership Renewal</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #1f2937; background-color: #f3f4f6; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06); border: 1px solid #e5e7eb;">
        
        <!-- Header -->
        <div style="background: %s; color: white; padding: 32px 24px; text-align: center;">
            <div style="display: inline-block; padding: 6px 14px; border-radius: 20px; font-size: 12px; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; background: %s; margin-bottom: 12px;">
                %s
            </div>
            <h1 style="margin: 0 0 8px 0; font-size: 24px; font-weight: 800;">Annual Membership Renewal</h1>
            <p style="margin: 0; font-size: 14px; opacity: 0.9;">Tamil Nadu Agricultural Graduates Association</p>
        </div>

        <!-- Body -->
        <div style="padding: 32px 24px;">
            %s

            <!-- Action Button -->
            <div style="text-align: center; margin: 32px 0 24px 0;">
                <a href="https://www.nammataga.com/member-login" style="background: %s; color: #ffffff; padding: 14px 32px; font-size: 15px; font-weight: 700; text-decoration: none; border-radius: 8px; display: inline-block; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.15);">
                    Renew Membership Online &rarr;
                </a>
            </div>

            <p style="margin: 0; font-size: 13px; color: #6b7280; text-align: center;">
                We appreciate your continued dedication and support as a valued member of TAGA.
            </p>
        </div>

        <!-- Footer -->
        <div style="background: #f8fafc; padding: 24px; text-align: center; border-top: 1px solid #e2e8f0;">
            <p style="margin: 0 0 6px 0; font-size: 13px; font-weight: 600; color: #475569;">Tamil Nadu Agricultural Graduates Association</p>
            <p style="margin: 0 0 8px 0; font-size: 12px; color: #94a3b8;">TAGA Towers, Chennai &bull; Support: <a href="mailto:nammataga@gmail.com" style="color: #047857; text-decoration: none;">nammataga@gmail.com</a></p>
            <p style="margin: 0; font-size: 11px; color: #cbd5e1;">&copy; 2026 TAGA. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, headerGradient, badgeColor, badgeText, customContent, actionBtnBg)
}

// SendRemindersIfDue checks the date and sends the appropriate reminder email to all unpaid members.
func SendRemindersIfDue() error {
	now := time.Now()
	config.Logger.Info("SendRemindersIfDue called", zap.String("now", now.Format("2006-01-02 15:04:05")))

	year := now.Year()
	if now.Month() < 4 {
		year--
	}
	config.Logger.Info("Membership year for logging", zap.Int("year", year))

	var reminderType string
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	r1 := getReminderDate(Reminder1, now)
	r2 := getReminderDate(Reminder2, now)
	r3 := getReminderDate(Reminder3, now)

	config.Logger.Info("Reminder dates calculated",
		zap.String("r1", r1.Format("2006-01-02")),
		zap.String("r2", r2.Format("2006-01-02")),
		zap.String("r3", r3.Format("2006-01-02")),
		zap.String("today", today.Format("2006-01-02")),
	)

	switch {
	case today.Equal(r1):
		reminderType = Reminder1
	case today.Equal(r2):
		reminderType = Reminder2
	case today.Equal(r3):
		reminderType = Reminder3
	default:
		config.Logger.Info("Not a reminder day, exiting")
		return nil
	}

	config.Logger.Info("Reminder type determined", zap.String("type", reminderType))

	unpaid, err := getUnpaidMembers()
	if err != nil {
		config.Logger.Error("Failed to get unpaid members", zap.Error(err))
		return err
	}
	config.Logger.Info("Unpaid members count", zap.Int("count", len(unpaid)))

	sentCount := 0
	for _, emailAddr := range unpaid {
		if alreadySent(emailAddr, reminderType, year) {
			config.Logger.Info("Already sent, skipping", zap.String("email", emailAddr))
			continue
		}

		memberName := strings.Split(emailAddr, "@")[0]
		if name := getMemberNameByEmail(emailAddr); name != "" {
			memberName = name
		}

		html := buildEmailHTML(reminderType, memberName)
		subject := ""
		switch reminderType {
		case Reminder1:
			subject = "🌱 Your TAGA Membership Renewal for " + fmt.Sprintf("%d‑%d", year, year+1)
		case Reminder2:
			subject = "⏰ Reminder: Renew Your TAGA Annual Subscription"
		case Reminder3:
			subject = "⚠️ Final Notice: Renew TAGA Membership Before May 31"
		}

		config.Logger.Info("Attempting to send email", zap.String("to", emailAddr), zap.String("subject", subject))
		if err := email.SendHTMLEmail(emailAddr, subject, html); err != nil {
			config.Logger.Error("Failed to send renewal email",
				zap.String("email", emailAddr),
				zap.String("reminder", reminderType),
				zap.Error(err),
			)
			continue
		}

		logEntry := RenewalReminderLog{
			MemberEmail:  emailAddr,
			ReminderType: reminderType,
			SentAt:       time.Now(),
			Year:         year,
		}
		if err := saveRenewalLog(logEntry); err != nil {
			config.Logger.Error("Failed to save renewal log", zap.Error(err))
		}
		sentCount++
		config.Logger.Info("Renewal reminder sent",
			zap.String("email", emailAddr),
			zap.String("reminder", reminderType),
		)
	}

	config.Logger.Info("Renewal reminders completed", zap.Int("sent", sentCount))
	return nil
}

// getMemberNameByEmail looks up a member's name from members.json
func getMemberNameByEmail(email string) string {
	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		return ""
	}
	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return ""
	}
	for _, m := range members {
		if mEmail, ok := m["emailId"].(string); ok && mEmail == email {
			if name, ok := m["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}
