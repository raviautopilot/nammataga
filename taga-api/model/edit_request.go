package model

type EditRequest struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	MobileNumber       string `json:"mobileNumber"`
	MailId             string `json:"mailId"`
	Designation        string `json:"designation"`
	WorkingDistrict    string `json:"workingDistrict"`
	ResidentialAddress string `json:"residentialAddress"`
	PermanentAddress   string `json:"permanentAddress"`
	Remarks            string `json:"remarks"`
	Status             string `json:"status"` // pending / approved / rejected
	CreatedAt          string `json:"createdAt"`
}

// FieldEditRequest represents a single field change within a larger edit request
type FieldEditRequest struct {
	ID             string `json:"id"`
	RequestGroupID string `json:"requestGroupId"` // Links multiple fields submitted at once
	MemberID       string `json:"memberId"`
	Email          string `json:"email"`
	MemberName     string `json:"memberName"`
	Field          string `json:"field"`          // e.g., "MobileNumber"
	OldValue       string `json:"oldValue"`
	NewValue       string `json:"newValue"`
	MemberRemarks  string `json:"memberRemarks"`  // Remarks from the member
	AdminRemarks   string `json:"adminRemarks"`   // Remarks from the admin upon processing
	Status         string `json:"status"`         // pending / approved / rejected
	CreatedAt      string `json:"createdAt"`
	ProcessedAt    string `json:"processedAt,omitempty"`
}
