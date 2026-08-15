package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"taga-api/config"
	"taga-api/model"
	"taga-api/service/audit"
	"taga-api/service/member"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
)

type CreateSubscriptionOrderRequest struct {
	SubscriptionID string                 `json:"subscriptionId" binding:"required"`
	Amount         int                    `json:"amount" binding:"required"`
	Email          string                 `json:"email" binding:"required"`
	Notes          map[string]interface{} `json:"notes,omitempty"`
}

type CreateSubscriptionOrderResponse struct {
	OrderID  string `json:"orderId"`
	Key      string `json:"key"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// getMembershipYearEnd returns the end date of the membership year (March 31) that contains 'now'.
// Membership year runs from April 1 to March 31.
func getMembershipYearEnd(now time.Time) time.Time {
	year := now.Year()
	// If now is in April or later, the membership year ends on March 31 of next year.
	if now.Month() >= 4 {
		return time.Date(year+1, 3, 31, 23, 59, 59, 0, now.Location())
	}
	// If now is in January to March, it's still part of the previous year's membership.
	return time.Date(year, 3, 31, 23, 59, 59, 0, now.Location())
}

// hasMemberPaidOneTime checks if a member has already paid a one‑time subscription.
// For one‑time fees, any existing record (regardless of status) means they cannot pay again.
func hasMemberPaidOneTime(subscriptionID, email string) bool {
	filePath := getMemberSubscriptionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	var subscriptions []model.MemberSubscription
	json.Unmarshal(data, &subscriptions)
	for _, sub := range subscriptions {
		if sub.MemberEmail == email && sub.SubscriptionID == subscriptionID {
			return true
		}
	}
	return false
}

// CreateSubscriptionOrder godoc
// @Summary Create subscription order
// @Description Create a Razorpay order for subscription payment
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body CreateSubscriptionOrderRequest true "Order Details"
// @Success 200 {object} CreateSubscriptionOrderResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/create-order [post]
func CreateSubscriptionOrder(c *gin.Context) {
	var req CreateSubscriptionOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		config.Logger.Warn("Failed to bind CreateSubscriptionOrder request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.Logger.Info("Creating Razorpay subscription order",
		zap.String("subscription_id", req.SubscriptionID),
		zap.Int("amount", req.Amount),
		zap.String("email", req.Email),
	)

	// Load subscription metadata to know if it's one‑time and get subscription name
	var subscriptionsMeta []map[string]interface{}
	subsFile := config.Config.Data.Config.SubscriptionType
	metaData, err := os.ReadFile(subsFile)
	if err == nil {
		json.Unmarshal(metaData, &subscriptionsMeta)
	}

	isOneTime := false
	subscriptionName := req.SubscriptionID
	for _, sub := range subscriptionsMeta {
		if id, _ := sub["id"].(string); id == req.SubscriptionID {
			if ot, ok := sub["oneTime"].(bool); ok && ot {
				isOneTime = true
			}
			if name, ok := sub["name"].(string); ok && name != "" {
				subscriptionName = name
			}
			break
		}
	}

	// Prevent duplicate payment for one‑time fees
	if isOneTime && hasMemberPaidOneTime(req.SubscriptionID, req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This one-time fee has already been paid and cannot be paid again"})
		return
	}

	// Get member details for notes
	memberName := getMemberNameByEmail(req.Email)
	memberTagaID := getMemberTagaIdByEmail(req.Email)

	// Get Razorpay credentials from environment
	razorpayKey := os.Getenv("RAZORPAY_KEY")
	razorpaySecret := os.Getenv("RAZORPAY_SECRET")

	// If payment is disabled, fallback to mock credentials if not set
	if config.Config.DisablePayment {
		if razorpayKey == "" {
			razorpayKey = "mock_key"
		}
		if razorpaySecret == "" {
			razorpaySecret = "mock_secret"
		}
	} else if razorpayKey == "" || razorpaySecret == "" {
		config.Logger.Error("Razorpay credentials missing from environment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway not configured"})
		return
	}

	// Create a short receipt ID (max 40 characters)
	shortUUID := uuid.New().String()[:30]
	receipt := fmt.Sprintf("sub_%s", shortUUID)

	// Build notes with subscription and member details
	orderNotes := map[string]interface{}{
		"subscription_id":   req.SubscriptionID,
		"subscription_name": subscriptionName,
		"subscription_type": subscriptionName,
		"member_email":      req.Email,
		"member_name":       memberName,
		"member_taga_id":    memberTagaID,
		"payment_type":      "subscription",
	}

	// Merge any additional notes from frontend
	for k, v := range req.Notes {
		orderNotes[k] = v
	}

	orderData := map[string]interface{}{
		"amount":   req.Amount,
		"currency": "INR",
		"receipt":  receipt,
		"notes":    orderNotes,
	}

	var order map[string]interface{}

	if config.Config.DisablePayment {
		mockOrderID := "mock_order_" + uuid.New().String()[:18]
		order = map[string]interface{}{
			"id":       mockOrderID,
			"amount":   req.Amount,
			"currency": "INR",
			"receipt":  receipt,
			"status":   "created",
		}
		config.Logger.Info("Bypassed Razorpay order creation, generated mock order", zap.String("order_id", mockOrderID))
	} else {
		// Initialize Razorpay client
		client := razorpay.NewClient(razorpayKey, razorpaySecret)

		config.Logger.Debug("Razorpay order payload", zap.Any("order_data", orderData))

		order, err = client.Order.Create(orderData, nil)
		if err != nil {
			config.Logger.Error("Razorpay API order creation failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order: " + err.Error()})
			return
		}

		config.Logger.Info("Razorpay subscription order created successfully", zap.Any("order_id", order["id"]))
	}

	c.JSON(http.StatusOK, CreateSubscriptionOrderResponse{
		OrderID:  order["id"].(string),
		Key:      razorpayKey,
		Amount:   req.Amount,
		Currency: "INR",
	})
}

// VerifySubscriptionPayment godoc
// @Summary Verify subscription payment
// @Description Verify Razorpay signature and update subscription status
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param verification body object true "Razorpay verification data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/subscriptions/verify-payment [post]
func VerifySubscriptionPayment(c *gin.Context) {
	var req struct {
		SubscriptionID string `json:"subscriptionId"`
		OrderID        string `json:"razorpay_order_id"`
		PaymentID      string `json:"razorpay_payment_id"`
		Signature      string `json:"razorpay_signature"`
		Email          string `json:"email"`
		Amount         int    `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify signature using environment secret
	isMock := config.Config.DisablePayment || strings.HasPrefix(req.OrderID, "mock_order_") || req.Signature == "mock_signature"
	if isMock {
		config.Logger.Info("Bypassing payment signature verification for mock payment", zap.String("order_id", req.OrderID))
	} else {
		razorpaySecret := os.Getenv("RAZORPAY_SECRET")
		data := req.OrderID + "|" + req.PaymentID
		h := hmac.New(sha256.New, []byte(razorpaySecret))
		h.Write([]byte(data))
		expectedSignature := hex.EncodeToString(h.Sum(nil))

		if expectedSignature != req.Signature {
			config.Logger.Error("Payment verification failed - invalid signature")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment verification failed"})
			return
		}
	}

	// Update transaction
	updatePaymentTransaction(req.OrderID, "captured", req.PaymentID)

	// Get member info
	memberID := getMemberTagaIdByEmail(req.Email)
	memberName := getMemberNameByEmail(req.Email)

	// Load subscription metadata to determine type (one‑time, need‑based, annual)
	var subscriptionsMeta []map[string]interface{}
	subsFile := config.Config.Data.Config.SubscriptionType
	metaData, err := os.ReadFile(subsFile)
	if err != nil {
		config.Logger.Error("Failed to read subscriptions metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment verification failed"})
		return
	}
	json.Unmarshal(metaData, &subscriptionsMeta)

	isOneTime := false
	isNeedBased := false
	for _, sub := range subscriptionsMeta {
		if id, _ := sub["id"].(string); id == req.SubscriptionID {
			if ot, ok := sub["oneTime"].(bool); ok && ot {
				isOneTime = true
			}
			if nb, ok := sub["needBased"].(bool); ok && nb {
				isNeedBased = true
			}
			break
		}
	}

	// ===== Set start/end dates based on subscription type =====
	startDate := time.Now()
	var endDate, nextDueDate time.Time
	status := "active" // default

	if isOneTime {
		// One‑time: never expires (set far future)
		endDate = time.Date(2099, 12, 31, 23, 59, 59, 0, startDate.Location())
		nextDueDate = endDate
	} else if isNeedBased {
		// Need‑based: no active status, treat as completed immediately
		endDate = startDate
		nextDueDate = startDate
		status = "completed"
	} else if req.SubscriptionID == "annual-subscription" {
		// Membership year ends on March 31; next due is April 1 of the following year
		endDate = getMembershipYearEnd(startDate)
		nextDueDate = endDate.AddDate(0, 0, 1) // April 1
	} else {
		// Fallback for other fixed one‑time (should not happen)
		endDate = startDate.AddDate(1, 0, 0)
		nextDueDate = endDate
	}

	// Get subscription name
	subscriptionName := getSubscriptionName(req.SubscriptionID)

	// ===== For annual subscriptions: expire any existing active ones =====
	if req.SubscriptionID == "annual-subscription" {
		expireExistingAnnualSubscriptions(req.Email)
	}

	// Save member subscription
	memberSub := model.MemberSubscription{
		ID:               uuid.New().String(),
		MemberID:         memberID,
		MemberEmail:      req.Email,
		MemberName:       memberName,
		SubscriptionID:   req.SubscriptionID,
		SubscriptionName: subscriptionName,
		Amount:           req.Amount,
		OrderID:          req.OrderID,
		PaymentID:        req.PaymentID,
		Status:           status,
		StartDate:        startDate,
		EndDate:          endDate,
		LastPaidDate:     startDate,
		NextDueDate:      nextDueDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	saveMemberSubscription(memberSub)

	// Update member payment status only for annual subscription
	if req.SubscriptionID == "annual-subscription" {
		if err := member.UpdateMemberPaymentStatus(req.Email, true); err != nil {
			config.Logger.Error("Failed to update payment status", zap.Error(err))
		}
	}

	// Audit subscription payment confirmation
	_ = audit.Log(c, memberID, req.Email,
		audit.ActionPaymentConfirmed, audit.ModulePayment,
		"subscription", req.SubscriptionID,
		fmt.Sprintf("Member %s paid %d for subscription '%s' (Order: %s, Payment: %s)",
			req.Email, req.Amount, subscriptionName, req.OrderID, req.PaymentID),
		nil, map[string]interface{}{
			"subscription_id":   req.SubscriptionID,
			"subscription_name": subscriptionName,
			"amount":            req.Amount,
			"order_id":          req.OrderID,
			"payment_id":        req.PaymentID,
		})

	// ========== SEND ADMIN EMAIL (DIRECT - NO WEBHOOK NEEDED) ==========
	// Check if email already sent for this payment (duplicate prevention)
	paymentID := req.PaymentID
	if !hasEmailBeenSent(paymentID) {
		// Prepare email data
		emailData := AdminSubscriptionData{
			PaymentID:        paymentID,
			OrderID:          req.OrderID,
			Amount:           req.Amount,
			CustomerEmail:    req.Email,
			SubscriptionID:   req.SubscriptionID,
			SubscriptionName: subscriptionName,
			MemberName:       memberName,
			MemberTagaID:     memberID,
			MemberEmail:      req.Email,
			PaymentType:      "subscription",
		}

		// Build email subject (professional format)
		amountInRupees := float64(req.Amount) / 100
		subject := fmt.Sprintf("💰 TAGA Payment Received - %s - ₹%.2f", subscriptionName, amountInRupees)

		// Get email body from admin_notify.go
		emailBody := buildSubscriptionEmailBody(emailData)

		// Send email with retry mechanism (2 retries, 2 sec delay)
		adminEmail := config.GetConfig().AdminEmail
		if adminEmail != "" {
			go sendEmailWithRetry(adminEmail, subject, emailBody, paymentID, "subscription", 2)
		} else {
			config.Logger.Warn("Admin email not configured, skipping notification")
		}
	} else {
		config.Logger.Info("Email already sent for this payment, skipping duplicate",
			zap.String("payment_id", paymentID))
	}
	// ========== END OF ADMIN EMAIL ==========

	c.JSON(http.StatusOK, gin.H{
		"message":     "Payment verified successfully",
		"status":      status,
		"valid_until": endDate.Format("2006-01-02"),
	})
}

// buildSubscriptionEmailBody builds the HTML email body for subscription payments
func buildSubscriptionEmailBody(data AdminSubscriptionData) string {
	amountInRupees := float64(data.Amount) / 100

	var body strings.Builder
	fmt.Fprintf(&body, `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #065f46; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 20px; border: 1px solid #e5e7eb; }
        .footer { background: #f3f4f6; padding: 15px; text-align: center; font-size: 12px; color: #6b7280; border-radius: 0 0 8px 8px; }
        table { width: 100%%; border-collapse: collapse; margin: 15px 0; }
        td { padding: 12px 10px; border-bottom: 1px solid #e5e7eb; }
        .label { font-weight: bold; width: 35%%; background: #f3f4f6; }
        .value { background: white; }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: bold; background: #dbeafe; color: #1e40af; }
        h3 { color: #065f46; margin-top: 20px; margin-bottom: 10px; }
        hr { border: none; border-top: 1px solid #e5e7eb; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>TAGA Payment Notification</h2>
            <p>%s</p>
        </div>
        <div class="content">
            <div style="text-align: center; margin-bottom: 20px;">
                <span class="badge">📋 SUBSCRIPTION PAYMENT</span>
            </div>
            
            <h3>💰 Payment Summary</h3>
            <table>
                <tr><td class="label">Payment ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Order ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Amount:</td><td class="value">₹ %.2f</td></tr>
                <tr><td class="label">Customer Email:</td><td class="value">%s</td></tr>
            }</table>
            
            <h3>📋 Subscription Details</h3>
            <table>
                <tr><td class="label">Subscription Type:</td><td class="value">%s</td></tr>
                <tr><td class="label">Subscription ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Member Name:</td><td class="value">%s</td></tr>
                <tr><td class="label">TAGA ID:</td><td class="value">%s</td></tr>
                <tr><td class="label">Member Email:</td><td class="value">%s</td></tr>
            }</table>
            
            <hr>
            <p style="font-size: 12px; color: #6b7280; text-align: center;">
                This is an automated notification from TAGA Payment System.
            </p>
        </div>
        <div class="footer">
            <p>TAGA Towers | Agriculture Complex Road, Chennai - 600017</p>
        </div>
    </div>
</body>
</html>`,
		time.Now().Format("January 02, 2006 at 03:04 PM"),
		data.PaymentID,
		data.OrderID,
		amountInRupees,
		data.CustomerEmail,
		data.SubscriptionName,
		data.SubscriptionID,
		data.MemberName,
		data.MemberTagaID,
		data.MemberEmail,
	)

	return body.String()
}

// expireExistingAnnualSubscriptions sets all active annual subscriptions for a given email to "expired".
func expireExistingAnnualSubscriptions(email string) {
	filePath := getMemberSubscriptionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var subscriptions []model.MemberSubscription
	json.Unmarshal(data, &subscriptions)

	modified := false
	for i, sub := range subscriptions {
		if sub.MemberEmail == email && sub.SubscriptionID == "annual-subscription" && sub.Status == "active" {
			subscriptions[i].Status = "expired"
			subscriptions[i].UpdatedAt = time.Now()
			modified = true
		}
	}

	if modified {
		updatedData, _ := json.MarshalIndent(subscriptions, "", "  ")
		os.WriteFile(filePath, updatedData, 0644)
	}
}

// GetMemberSubscriptionStatus godoc
// @Summary Get member subscription status
// @Description Returns the active subscription details and status of a member by email
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Param email query string true "Member Email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/subscriptions/status [get]
func GetMemberSubscriptionStatus(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	subscription, err := getActiveMemberSubscription(email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"is_paid": false,
			"status":  "inactive",
			"message": "No active subscription found",
		})
		return
	}

	if subscription.EndDate.Before(time.Now()) {
		updateSubscriptionStatus(subscription.ID, "expired")
		c.JSON(http.StatusOK, gin.H{
			"is_paid":    false,
			"status":     "expired",
			"expired_on": subscription.EndDate.Format("2006-01-02"),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_paid":         true,
		"status":          subscription.Status,
		"subscription_id": subscription.SubscriptionID,
		"valid_until":     subscription.EndDate.Format("2006-01-02"),
		"next_due_date":   subscription.NextDueDate.Format("2006-01-02"),
	})
}

// Helper functions for storage
func getMemberSubscriptionsFilePath() string {
	return filepath.Join("data", "subscriptions", "member_subscriptions.json")
}

func getPaymentTransactionsFilePath() string {
	return filepath.Join("data", "subscriptions", "payment_transactions.json")
}



func updatePaymentTransaction(orderID, status, paymentID string) {
	filePath := getPaymentTransactionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var transactions []model.PaymentTransaction
	json.Unmarshal(data, &transactions)

	for i, t := range transactions {
		if t.OrderID == orderID {
			transactions[i].Status = status
			transactions[i].PaymentID = paymentID
			break
		}
	}

	updatedData, _ := json.MarshalIndent(transactions, "", "  ")
	os.WriteFile(filePath, updatedData, 0644)
}

func saveMemberSubscription(subscription model.MemberSubscription) {
	filePath := getMemberSubscriptionsFilePath()
	os.MkdirAll(filepath.Dir(filePath), 0755)

	var subscriptions []model.MemberSubscription
	data, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(data, &subscriptions)
	}

	subscriptions = append(subscriptions, subscription)
	updatedData, _ := json.MarshalIndent(subscriptions, "", "  ")
	os.WriteFile(filePath, updatedData, 0644)
}

func getActiveMemberSubscription(email string) (*model.MemberSubscription, error) {
	filePath := getMemberSubscriptionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var subscriptions []model.MemberSubscription
	json.Unmarshal(data, &subscriptions)

	for _, s := range subscriptions {
		if s.MemberEmail == email && s.Status == "active" {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("no active subscription found")
}

func updateSubscriptionStatus(subscriptionID, status string) {
	filePath := getMemberSubscriptionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var subscriptions []model.MemberSubscription
	json.Unmarshal(data, &subscriptions)

	for i, s := range subscriptions {
		if s.ID == subscriptionID {
			subscriptions[i].Status = status
			subscriptions[i].UpdatedAt = time.Now()
			break
		}
	}

	updatedData, _ := json.MarshalIndent(subscriptions, "", "  ")
	os.WriteFile(filePath, updatedData, 0644)
}



func getMemberNameByEmail(email string) string {
	members, err := readExistingMembers()
	if err != nil {
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

func getSubscriptionName(subscriptionID string) string {
	filePath := config.Config.Data.Config.SubscriptionType
	data, err := os.ReadFile(filePath)
	if err != nil {
		return subscriptionID
	}

	var subscriptions []map[string]interface{}
	json.Unmarshal(data, &subscriptions)

	for _, s := range subscriptions {
		if id, ok := s["id"].(string); ok && id == subscriptionID {
			if name, ok := s["name"].(string); ok {
				return name
			}
		}
	}
	return subscriptionID
}

// GetMemberPaidSubscriptions godoc
// @Summary Get member paid subscriptions
// @Description Returns list of subscription IDs that the member has paid for
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Param email query string true "Member Email"
// @Success 200 {array} string
// @Failure 400 {object} map[string]string
// @Router /api/subscriptions/member-paid [get]
func GetMemberPaidSubscriptions(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	paidIDs := []string{}
	now := time.Now()

	// 1. Determine if annual subscription is paid (active or grace period)
	isAnnualPaid := false

	// Check active annual subscription (strictly paid for current year, no grace period)
	sub, err := getActiveMemberSubscription(email)
	if err == nil && sub.SubscriptionID == "annual-subscription" && sub.Status == "active" && now.Before(sub.EndDate) {
		isAnnualPaid = true
	}
	if isAnnualPaid {
		paidIDs = append(paidIDs, "annual-subscription")
	}

	// 2. Load all subscriptions metadata to identify one‑time subscription IDs
	metaFile := config.Config.Data.Config.SubscriptionType
	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		config.Logger.Warn("Failed to read subscriptions metadata for paid list", zap.Error(err))
		c.JSON(http.StatusOK, paidIDs)
		return
	}
	var subsMeta []map[string]interface{}
	json.Unmarshal(metaData, &subsMeta)

	oneTimeIDs := make(map[string]bool)
	for _, s := range subsMeta {
		if id, _ := s["id"].(string); id != "" {
			if ot, ok := s["oneTime"].(bool); ok && ot {
				oneTimeIDs[id] = true
			}
		}
	}

	// 3. Scan member's subscription records for one‑time fees (any record counts)
	filePath := getMemberSubscriptionsFilePath()
	data, err := os.ReadFile(filePath)
	if err == nil {
		var subscriptions []model.MemberSubscription
		json.Unmarshal(data, &subscriptions)
		seen := make(map[string]bool)
		for _, sub := range subscriptions {
			if sub.MemberEmail == email && oneTimeIDs[sub.SubscriptionID] && !seen[sub.SubscriptionID] {
				paidIDs = append(paidIDs, sub.SubscriptionID)
				seen[sub.SubscriptionID] = true
			}
		}
	}

	c.JSON(http.StatusOK, paidIDs)
}
