package model

import "time"

// MemberSubscription tracks member's subscription status
type MemberSubscription struct {
	ID               string    `json:"id"`
	MemberID         string    `json:"member_id"`
	MemberEmail      string    `json:"member_email"`
	MemberName       string    `json:"member_name"`
	SubscriptionID   string    `json:"subscription_id"`
	SubscriptionName string    `json:"subscription_name"`
	Amount           int       `json:"amount"`
	OrderID          string    `json:"order_id"`
	PaymentID        string    `json:"payment_id"`
	Status           string    `json:"status"` // pending, active, expired, cancelled
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	LastPaidDate     time.Time `json:"last_paid_date"`
	NextDueDate      time.Time `json:"next_due_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// PaymentTransaction records all payments
type PaymentTransaction struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	PaymentID      string    `json:"payment_id"`
	Signature      string    `json:"signature"`
	MemberID       string    `json:"member_id"`
	MemberEmail    string    `json:"member_email"`
	SubscriptionID string    `json:"subscription_id"`
	Amount         int       `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"` // created, captured, refunded, failed
	CreatedAt      time.Time `json:"created_at"`
}
