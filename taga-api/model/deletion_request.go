package model

// DeletionRequestStatus represents the status of an account deletion request
type DeletionRequestStatus string

const (
	DeletionStatusPending  DeletionRequestStatus = "pending"
	DeletionStatusApproved DeletionRequestStatus = "approved"
	DeletionStatusRejected DeletionRequestStatus = "rejected"
)

// DeletionRequest represents a member's request to delete their account
type DeletionRequest struct {
	ID          string                `json:"id"`
	MemberID    string                `json:"member_id"`
	TagaID      string                `json:"tagaId"`
	TagaIDSnake string                `json:"taga_id,omitempty"`
	MemberName  string                `json:"member_name"`
	Email       string                `json:"email"`
	Status      DeletionRequestStatus `json:"status"`
	Reason      string                `json:"reason,omitempty"`
	Remarks     string                `json:"remarks,omitempty"`
	RequestedAt string                `json:"requested_at"`
	ProcessedAt string                `json:"processed_at,omitempty"`
	ProcessedBy string                `json:"processed_by,omitempty"`
}

// AccountDeletionRequestPayload is the payload for submitting an account deletion request
type AccountDeletionRequestPayload struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason"`
}

// RejectDeletionRequestPayload is the payload for rejecting a deletion request
type RejectDeletionRequestPayload struct {
	Remarks string `json:"remarks" binding:"required"`
}

// AccountDeletionEligibilityResponse is the response for checking account deletion eligibility
type AccountDeletionEligibilityResponse struct {
	Eligible bool   `json:"eligible"`
	Message  string `json:"message"`
	Reason   string `json:"reason,omitempty"`
}
