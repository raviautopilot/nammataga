package model

import "time"

type Grievance struct {
	ID                string    `json:"id"`
	Subject           string    `json:"subject"`
	Category          string    `json:"category"`
	Priority          string    `json:"priority"`
	Description       string    `json:"description"`
	ContactPhone      string    `json:"contactPhone"`
	PreferredResponse string    `json:"preferredResponse"`
	Status            string    `json:"status"`
	AssignedTo        string    `json:"assignedTo"`
	SubmittedDate     time.Time `json:"submittedDate"`
	LastUpdate        time.Time `json:"lastUpdate"`
}
