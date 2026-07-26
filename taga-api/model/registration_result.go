package model

type RegistrationResult struct {
	SuccessCount int                 `json:"successCount"`
	Failed       []RegistrationError `json:"failed"`
}
