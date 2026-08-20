// handler/webhook.go
package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"taga-api/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProcessedPayment tracks payments that have been processed
type ProcessedPayment struct {
	PaymentID   string    `json:"payment_id"`
	OrderID     string    `json:"order_id"`
	ProcessedAt time.Time `json:"processed_at"`
}

// WebhookPayload represents the Razorpay webhook structure
type WebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID        string                 `json:"id"`
				OrderID   string                 `json:"order_id"`
				Amount    int                    `json:"amount"`
				Currency  string                 `json:"currency"`
				Status    string                 `json:"status"`
				Email     string                 `json:"email"`
				Contact   string                 `json:"contact"`
				Notes     map[string]interface{} `json:"notes"`
				Customer  map[string]interface{} `json:"customer"`
				CreatedAt int64                  `json:"created_at"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

var (
	processedPayments     = make(map[string]ProcessedPayment)
	processedPaymentsLock sync.RWMutex
	// The path is now loaded dynamically inside functions where needed
)

func getProcessedPaymentsFile() string {
	return config.Config.ProcessedPaymentsFile
}

// loadProcessedPayments loads previously processed payments from JSON file
func loadProcessedPayments() {
	processedPaymentsLock.Lock()
	defer processedPaymentsLock.Unlock()

	data, err := os.ReadFile(getProcessedPaymentsFile())
	if err != nil {
		if !os.IsNotExist(err) {
			config.Logger.Warn("Failed to load processed payments file", zap.Error(err))
		}
		return
	}

	var payments []ProcessedPayment
	if err := json.Unmarshal(data, &payments); err != nil {
		config.Logger.Warn("Failed to parse processed payments", zap.Error(err))
		return
	}

	for _, p := range payments {
		processedPayments[p.PaymentID] = p
	}
	config.Logger.Info("Loaded processed payments", zap.Int("count", len(processedPayments)))
}

// saveProcessedPayment saves a processed payment to JSON file
func saveProcessedPayment(payment ProcessedPayment) {
	processedPaymentsLock.Lock()
	defer processedPaymentsLock.Unlock()

	processedPayments[payment.PaymentID] = payment

	// Convert map to slice for storage
	var payments []ProcessedPayment
	for _, p := range processedPayments {
		payments = append(payments, p)
	}

	// Ensure directory exists
	dir := filepath.Dir(getProcessedPaymentsFile())
	if err := os.MkdirAll(dir, 0755); err != nil {
		config.Logger.Error("Failed to create payments directory", zap.Error(err))
		return
	}

	data, err := json.MarshalIndent(payments, "", "  ")
	if err != nil {
		config.Logger.Error("Failed to marshal processed payments", zap.Error(err))
		return
	}

	if err := os.WriteFile(getProcessedPaymentsFile(), data, 0644); err != nil {
		config.Logger.Error("Failed to write processed payments file", zap.Error(err))
	}
}

// isPaymentAlreadyProcessed checks if a payment has already been handled
func isPaymentAlreadyProcessed(paymentID string) bool {
	processedPaymentsLock.RLock()
	defer processedPaymentsLock.RUnlock()
	_, exists := processedPayments[paymentID]
	return exists
}

// verifyWebhookSignature verifies that the webhook came from Razorpay
func verifyWebhookSignature(payload []byte, signature string) bool {
	razorpaySecret := os.Getenv("RAZORPAY_SECRET")
	if razorpaySecret == "" {
		config.Logger.Warn("RAZORPAY_SECRET not set, skipping webhook verification")
		return true // Skip verification in development (not recommended for production)
	}

	h := hmac.New(sha256.New, []byte(razorpaySecret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// WebhookHandler handles Razorpay webhook events
// @Summary Handle Razorpay webhook
// @Description Receives payment.captured events from Razorpay and sends admin notifications
// @Tags Webhook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/webhook/razorpay [post]
func WebhookHandler(c *gin.Context) {
	// Read raw body for signature verification
	body, err := c.GetRawData()
	if err != nil {
		config.Logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify webhook signature
	signature := c.GetHeader("X-Razorpay-Signature")
	if signature != "" && !verifyWebhookSignature(body, signature) {
		config.Logger.Error("Invalid webhook signature")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		config.Logger.Error("Failed to parse webhook payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// ADD THIS LOG - Webhook received
	config.Logger.Info("🔔 Webhook received",
		zap.String("event", payload.Event),
	)

	// Only process payment.captured events
	if payload.Event != "payment.captured" {
		c.JSON(http.StatusOK, gin.H{"message": "Ignored event"})
		return
	}

	payment := payload.Payload.Payment.Entity
	paymentID := payment.ID
	orderID := payment.OrderID

	config.Logger.Info("💰 Payment captured webhook",
		zap.String("payment_id", paymentID),
		zap.String("order_id", orderID),
		zap.Int("amount", payment.Amount),
	)

	// Check for duplicate payment
	if isPaymentAlreadyProcessed(paymentID) {
		config.Logger.Info("Payment already processed, skipping",
			zap.String("payment_id", paymentID))
		c.JSON(http.StatusOK, gin.H{"message": "Already processed"})
		return
	}

	// Extract customer email
	customerEmail := payment.Email
	if customerEmail == "" {
		if customer, ok := payment.Customer["email"].(string); ok {
			customerEmail = customer
		}
	}

	// Determine payment type from notes
	notes := payment.Notes
	paymentType := "unknown"

	if _, ok := notes["room_name"]; ok {
		paymentType = "room_booking"
	} else if _, ok := notes["subscription_type"]; ok {
		paymentType = "subscription"
	}

	config.Logger.Info("📋 Payment type detected",
		zap.String("payment_type", paymentType),
		zap.String("payment_id", paymentID),
	)

	// Process based on payment type
	switch paymentType {
	case "subscription":
		subscriptionName, _ := notes["subscription_name"].(string)
		if subscriptionName == "" {
			subscriptionName, _ = notes["subscription_type"].(string)
		}
		subscriptionID, _ := notes["subscription_id"].(string)
		memberName, _ := notes["member_name"].(string)
		memberTagaID, _ := notes["member_taga_id"].(string)
		memberEmail, _ := notes["member_email"].(string)

		emailData := AdminSubscriptionData{
			PaymentID:        paymentID,
			OrderID:          orderID,
			Amount:           payment.Amount,
			CustomerEmail:    customerEmail,
			SubscriptionID:   subscriptionID,
			SubscriptionName: subscriptionName,
			MemberName:       memberName,
			MemberTagaID:     memberTagaID,
			MemberEmail:      memberEmail,
			PaymentType:      "subscription",
		}

		// Send email asynchronously
		go func() {
			config.Logger.Info("📧 Queuing subscription admin email",
				zap.String("payment_id", paymentID),
				zap.String("subscription", subscriptionName))
			if err := sendAdminSubscriptionEmail(emailData); err != nil {
				config.Logger.Error("❌ Failed to send subscription admin email",
					zap.String("payment_id", paymentID),
					zap.Error(err))
			} else {
				config.Logger.Info("✅ Subscription admin email sent successfully",
					zap.String("payment_id", paymentID),
					zap.String("subscription", subscriptionName))
			}
		}()

	case "room_booking":
		roomName, _ := notes["room_name"].(string)
		roomNumber, _ := notes["room_number"].(string)
		bedCount, _ := notes["bed_count"].(float64)
		checkInDate, _ := notes["check_in"].(string)
		checkOutDate, _ := notes["check_out"].(string)
		bookerName, _ := notes["booker_name"].(string)
		bookerTagaID, _ := notes["booker_taga_id"].(string)
		bookerPhone, _ := notes["booker_phone"].(string)
		bookingFor, _ := notes["booking_type"].(string)
		if bookingFor == "" {
			bookingFor, _ = notes["booking_for"].(string)
		}
		guestDetailsJSON, _ := notes["guest_details"].(string)

		bookingID, _ := notes["booking_id"].(string)

		emailData := AdminRoomBookingData{
			BookingID:     bookingID,
			PaymentID:     paymentID,
			OrderID:       orderID,
			Amount:        payment.Amount,
			CustomerEmail: customerEmail,
			CustomerPhone: bookerPhone,
			RoomName:      roomName,
			RoomNumber:    roomNumber,
			BedCount:      int(bedCount),
			CheckInDate:   checkInDate,
			CheckOutDate:  checkOutDate,
			BookerName:    bookerName,
			BookerTagaID:  bookerTagaID,
			BookerPhone:   bookerPhone,
			BookingFor:    bookingFor,
			GuestDetails:  guestDetailsJSON,
			PaymentType:   "room_booking",
		}

		// Send email asynchronously
		go func() {
			config.Logger.Info("📧 Queuing room booking admin email",
				zap.String("payment_id", paymentID),
				zap.String("room", roomName))
			if err := sendAdminRoomBookingEmail(emailData); err != nil {
				config.Logger.Error("❌ Failed to send room booking admin email",
					zap.String("payment_id", paymentID),
					zap.Error(err))
			} else {
				config.Logger.Info("✅ Room booking admin email sent successfully",
					zap.String("payment_id", paymentID),
					zap.String("room", roomName))
			}
		}()

	default:
		config.Logger.Warn("⚠️ Unknown payment type, no email sent",
			zap.String("payment_id", paymentID),
			zap.Any("notes", notes))
	}

	// Mark payment as processed
	saveProcessedPayment(ProcessedPayment{
		PaymentID:   paymentID,
		OrderID:     orderID,
		ProcessedAt: time.Now(),
	})

	config.Logger.Info("✅ Webhook processed successfully",
		zap.String("payment_id", paymentID))

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}

// init function to load existing processed payments when package initializes
func init() {
	loadProcessedPayments()
}
