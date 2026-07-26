package model

import "time"

type MemberNotification struct {
	ID             string     `json:"id"`
	TagaID         string     `json:"taga_id"`
	MemberID       string     `json:"member_id"`
	MemberEmail    string     `json:"member_email"`
	NotificationID string     `json:"notification_id"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Priority       string     `json:"priority"`
	IsRead         bool       `json:"is_read"`
	CreatedAt      time.Time  `json:"created_at"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
}
