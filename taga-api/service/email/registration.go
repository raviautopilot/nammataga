package email

import "taga-api/model"

// Called when registration fails
func SendRegistrationError(errors []model.RegistrationError) {
	// TODO: send error email to admin
	// include:
	// - member identifier
	// - error fields
}

// Called when registration succeeds
func SendRegistrationSuccess(email string, tempPassword string) {
	// TODO: send success email to member
	// include:
	// - login URL
	// - temporary password
}
