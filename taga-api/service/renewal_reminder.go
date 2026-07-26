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
	subsFile := filepath.Join("data", "subscriptions", "subscriptions.json")
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

// buildEmailHTML returns a beautiful HTML email for the given reminder type.
func buildEmailHTML(reminderType, memberName string) string {
	annualAmount := getAnnualSubscriptionAmount()

	// Shared header & footer
	header := `<div style="background:#0f5132;color:#fff;padding:20px;text-align:center;border-radius:8px 8px 0 0;">
                    <h1 style="margin:0;">TAGA Membership</h1>
                    <p style="margin:5px 0 0;">Tamil Nadu Agriculture Graduate Association</p>
                </div>`
	footer := `<div style="background:#f0f0f0;padding:15px;text-align:center;font-size:12px;color:#555;border-radius:0 0 8px 8px;">
                    This is an automated reminder. Please do not reply to this email.<br/>
                    For assistance, contact <a href="mailto:nammataga@gmail.com">nammataga@gmail.com</a>
                </div>`

	var body string
	switch reminderType {
	case Reminder1:
		body = fmt.Sprintf(`
            <p>Dear %s,</p>
            <p>Warm greetings from TAGA! 🌾</p>
            <p>The new membership year <strong>(April 1, %d – March 31, %d)</strong> has begun.</p>
            <p>To keep your membership active and continue enjoying all benefits, kindly renew your <strong>Annual Subscription</strong> of ₹%d at your earliest convenience.</p>
            <p>We have a grace period until <strong>May 31</strong>. After that, access to member‑only features will be restricted.</p>
            <p style="text-align:center;margin:30px 0;">
                <a href="https://www.nammataga.com/member-login" 
                   style="background:#0f5132;color:#fff;padding:12px 24px;text-decoration:none;border-radius:6px;font-weight:bold;">
                   Renew Now
                </a>
            </p>
            <p>Thank you for being a valued member of TAGA.</p>
        `, memberName, time.Now().Year(), time.Now().Year()+1, annualAmount)

	case Reminder2:
		body = fmt.Sprintf(`
            <p>Dear %s,</p>
            <p>We noticed that your Annual Subscription for the membership year <strong>(April 1, %d – March 31, %d)</strong> is still pending.</p>
            <p>This is a gentle reminder to renew at ₹%d to avoid any disruption to your membership.</p>
            <p>The grace period ends on <strong>May 31</strong>. After that, your account will become <span style="color:#c00;">inactive</span>.</p>
            <p style="text-align:center;margin:30px 0;">
                <a href="https://www.nammataga.com/member-login"
                   style="background:#0f5132;color:#fff;padding:12px 24px;text-decoration:none;border-radius:6px;font-weight:bold;">
                   Renew Now
                </a>
            </p>
            <p>We appreciate your continued support.</p>
        `, memberName, time.Now().Year(), time.Now().Year()+1, annualAmount)

	case Reminder3:
		body = fmt.Sprintf(`
            <p>Dear %s,</p>
            <p>This is the final reminder regarding your TAGA Annual Subscription for <strong>%d‑%d</strong>.</p>
            <p>After <strong>May 31</strong>, your membership will become <span style="color:#c00;">inactive</span> and you will lose access to exclusive member benefits.</p>
            <p>Please take a moment to complete the renewal today.</p>
            <p style="text-align:center;margin:30px 0;">
                <a href="https://www.nammataga.com/member-login"
                   style="background:#c00;color:#fff;padding:12px 24px;text-decoration:none;border-radius:6px;font-weight:bold;">
                   Renew Now Before It’s Too Late
                </a>
            </p>
            <p>We’d love to have you continue with us!</p>
        `, memberName, time.Now().Year(), time.Now().Year()+1)
	}

	fullHTML := fmt.Sprintf(`
    <html>
    <body style="font-family:Arial,sans-serif;background:#f4f4f4;padding:20px;">
        <div style="max-width:600px;margin:0 auto;background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,0.1);">
            %s
            <div style="padding:20px;">
                %s
            </div>
            %s
        </div>
    </body>
    </html>`, header, body, footer)

	return fullHTML
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
