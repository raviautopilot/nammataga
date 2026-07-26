package model

import "time"

type Notification struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	Priority   string    `json:"priority"` // normal, high, urgent
	SendTo     string    `json:"send_to"`  // all, paid, district
	SentBy     string    `json:"sent_by"`  // admin email
	SentAt     time.Time `json:"sent_at"`
	Recipients int       `json:"recipients"`         // number of members who received
	District   string    `json:"district,omitempty"` // if send_to is district
}
