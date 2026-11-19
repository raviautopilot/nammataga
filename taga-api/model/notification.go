package model

import "time"

type Notification struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Priority    string    `json:"priority"` // normal, high, urgent
	SendTo      string    `json:"send_to"`  // all, paid, district
	SentBy      string    `json:"sent_by"`  // admin email
	SentAt      time.Time `json:"sent_at"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	Recipients  int       `json:"recipients"`         // number of members who received
	District    string    `json:"district,omitempty"` // if send_to is district
	IsRead      bool      `json:"is_read"`            // backward compatibility for clients expecting is_read
}
