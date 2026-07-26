// handler/email_helper.go
package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"taga-api/config"

	"go.uber.org/zap"
)

// SentPayment tracks payments that have had admin emails sent
type SentPayment struct {
	PaymentID   string    `json:"payment_id"`
	PaymentType string    `json:"payment_type"` // "subscription" or "room_booking"
	SentAt      time.Time `json:"sent_at"`
}

// FailedEmail tracks email sending failures
type FailedEmail struct {
	PaymentID   string    `json:"payment_id"`
	PaymentType string    `json:"payment_type"`
	Error       string    `json:"error"`
	Attempts    int       `json:"attempts"`
	FailedAt    time.Time `json:"failed_at"`
}

var (
	sentPayments     = make(map[string]SentPayment)
	sentPaymentsLock sync.RWMutex
	sentPaymentsFile = filepath.Join("data", "emails", "sent_payments.json")

	failedEmails     = make([]FailedEmail, 0)
	failedEmailsLock sync.Mutex
	failedEmailsFile = filepath.Join("data", "emails", "failed_emails.json")
)

// safeLogInfo safely logs info if logger is initialized
func safeLogInfo(msg string, fields ...zap.Field) {
	if config.Logger != nil {
		config.Logger.Info(msg, fields...)
	}
}

// safeLogWarn safely logs warn if logger is initialized
func safeLogWarn(msg string, fields ...zap.Field) {
	if config.Logger != nil {
		config.Logger.Warn(msg, fields...)
	}
}

// safeLogError safely logs error if logger is initialized
func safeLogError(msg string, fields ...zap.Field) {
	if config.Logger != nil {
		config.Logger.Error(msg, fields...)
	}
}



// saveSentPayment saves a sent payment to JSON file
func saveSentPayment(payment SentPayment) {
	sentPaymentsLock.Lock()
	defer sentPaymentsLock.Unlock()

	sentPayments[payment.PaymentID] = payment

	// Convert map to slice for storage
	var payments []SentPayment
	for _, p := range sentPayments {
		payments = append(payments, p)
	}

	// Ensure directory exists
	dir := filepath.Dir(sentPaymentsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		safeLogError("Failed to create emails directory", zap.Error(err))
		return
	}

	data, err := json.MarshalIndent(payments, "", "  ")
	if err != nil {
		safeLogError("Failed to marshal sent payments", zap.Error(err))
		return
	}

	if err := os.WriteFile(sentPaymentsFile, data, 0644); err != nil {
		safeLogError("Failed to save sent payments", zap.Error(err))
	}
}

// hasEmailBeenSent checks if an email has already been sent for this payment
func hasEmailBeenSent(paymentID string) bool {
	sentPaymentsLock.RLock()
	defer sentPaymentsLock.RUnlock()
	_, exists := sentPayments[paymentID]
	return exists
}

// saveFailedEmail logs a failed email attempt
func saveFailedEmail(failed FailedEmail) {
	failedEmailsLock.Lock()
	defer failedEmailsLock.Unlock()

	failedEmails = append(failedEmails, failed)

	// Keep only last 100 failed emails
	if len(failedEmails) > 100 {
		failedEmails = failedEmails[len(failedEmails)-100:]
	}

	// Ensure directory exists
	dir := filepath.Dir(failedEmailsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		safeLogError("Failed to create emails directory", zap.Error(err))
		return
	}

	data, err := json.MarshalIndent(failedEmails, "", "  ")
	if err != nil {
		safeLogError("Failed to marshal failed emails", zap.Error(err))
		return
	}

	if err := os.WriteFile(failedEmailsFile, data, 0644); err != nil {
		safeLogError("Failed to save failed emails", zap.Error(err))
	}
}

// sendEmailWithRetry sends email with retry mechanism
func sendEmailWithRetry(to, subject, body string, paymentID, paymentType string, maxRetries int) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		safeLogInfo("📧 Sending email attempt",
			zap.Int("attempt", attempt),
			zap.String("payment_id", paymentID),
			zap.String("to", to),
		)

		err := sendEmail(to, subject, body)
		if err == nil {
			// Success - save to sent payments
			saveSentPayment(SentPayment{
				PaymentID:   paymentID,
				PaymentType: paymentType,
				SentAt:      time.Now(),
			})
			safeLogInfo("✅ Email sent successfully",
				zap.String("payment_id", paymentID),
				zap.Int("attempt", attempt),
			)
			return
		}

		lastErr = err
		safeLogWarn("Email send failed, retrying...",
			zap.String("payment_id", paymentID),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
			zap.Error(err),
		)

		if attempt < maxRetries {
			time.Sleep(2 * time.Second)
		}
	}

	// All retries failed - log to failed emails file
	safeLogError("❌ Email failed after all retries",
		zap.String("payment_id", paymentID),
		zap.Int("max_retries", maxRetries),
		zap.Error(lastErr),
	)

	saveFailedEmail(FailedEmail{
		PaymentID:   paymentID,
		PaymentType: paymentType,
		Error:       lastErr.Error(),
		Attempts:    maxRetries,
		FailedAt:    time.Now(),
	})
}

// NOTE: init() removed - files will be loaded when first needed
// This prevents panic because logger may not be initialized yet
