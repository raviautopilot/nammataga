package model

// ResetPasswordRequest is used for changing password when user knows their old password
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MemberForgotPasswordHandler struct {
	MembershipID     string `json:"membershipId"`
	Email            string `json:"email"`
	SecurityQuestion string `json:"securityQuestion"`
	SecurityAnswer   string `json:"securityAnswer"`
}
type UserFull struct {
	Identifier       string `json:"identifier"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	SecurityQuestion string `json:"securityQuestion"`
	SecurityAnswer   string `json:"securityAnswer"`
}
