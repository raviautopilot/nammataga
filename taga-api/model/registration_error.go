package model

type RegistrationError struct {
	Username string   `json:"username"`
	Errors   []string `json:"errors"`
}
