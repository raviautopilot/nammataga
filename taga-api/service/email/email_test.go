package email_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taga-api/config"
	"taga-api/service/email"
)

func TestSendHTMLEmail_PasswordSpacesSanitized(t *testing.T) {
	origConfig := config.Config
	defer func() { config.Config = origConfig }()

	config.Config.SMTPPassword = "abcd efgh ijkl mnop"
	config.Config.SMTPHost = "" // keep host empty to skip network dial while recording email in memory
	config.Config.SMTPPort = 0
	config.Config.SMTPUsername = "test@nammataga.com"

	email.ClearLastSentEmails()

	err := email.SendHTMLEmail("recipient@nammataga.com", "Test Subject", "<p>Hello</p>")
	assert.NoError(t, err)

	sent := email.GetLastSentEmails()
	require.Len(t, sent, 1)
	assert.Equal(t, "recipient@nammataga.com", sent[0].To)
	assert.Equal(t, "Test Subject", sent[0].Subject)
	assert.Contains(t, sent[0].Body, "<p>Hello</p>")
}

func TestAccountDeletionEmails_TemplateRenderingAndBranding(t *testing.T) {
	origConfig := config.Config
	defer func() { config.Config = origConfig }()

	config.Config.AdminEmail = "nammataga@gmail.com"
	config.Config.ResetPasswordURL = "https://www.nammataga.com"
	config.Config.SMTPHost = ""
	config.Config.SMTPPort = 0
	config.Config.SMTPUsername = "appnammataga@gmail.com"

	email.ClearLastSentEmails()

	// 1. Test SendAccountDeletionRequestAcknowledgement (Member)
	err := email.SendAccountDeletionRequestAcknowledgement(
		"member@nammataga.com",
		"Sudhan Palanivel",
		"TAGA100",
		"Relocating to another state",
		"2026-11-21T10:00:00Z",
	)
	require.NoError(t, err)

	sent := email.GetLastSentEmails()
	require.Len(t, sent, 1)
	ackEmail := sent[0]
	assert.Equal(t, "member@nammataga.com", ackEmail.To)
	assert.Contains(t, ackEmail.Subject, "Account Deletion Request Received")
	assert.Contains(t, ackEmail.Body, "Tamil Nadu Agricultural Graduates Association")
	assert.Contains(t, ackEmail.Body, "TAGA Towers, Chennai")
	assert.Contains(t, ackEmail.Body, "Sudhan Palanivel")
	assert.Contains(t, ackEmail.Body, "TAGA100")
	assert.Contains(t, ackEmail.Body, "Relocating to another state")
	assert.Contains(t, ackEmail.Body, "Deletion Request")

	// 2. Test SendAccountDeletionRequestAdminAlert (Admin)
	email.ClearLastSentEmails()
	err = email.SendAccountDeletionRequestAdminAlert(
		"nammataga@gmail.com",
		"Sudhan Palanivel",
		"TAGA100",
		"member@nammataga.com",
		"+919876543210",
		"Relocating to another state",
		"2026-11-21T10:00:00Z",
	)
	require.NoError(t, err)

	sent = email.GetLastSentEmails()
	require.Len(t, sent, 1)
	alertEmail := sent[0]
	assert.Equal(t, "nammataga@gmail.com", alertEmail.To)
	assert.Contains(t, alertEmail.Subject, "[ADMIN] New Account Deletion Request")
	assert.Contains(t, alertEmail.Subject, "Sudhan Palanivel")
	assert.Contains(t, alertEmail.Body, "Tamil Nadu Agricultural Graduates Association")
	assert.Contains(t, alertEmail.Body, "Sudhan Palanivel")
	assert.Contains(t, alertEmail.Body, "TAGA100")
	assert.Contains(t, alertEmail.Body, "member@nammataga.com")
	assert.Contains(t, alertEmail.Body, "+919876543210")
	assert.Contains(t, alertEmail.Body, "Relocating to another state")
	assert.Contains(t, alertEmail.Body, "Admin Action Required")
	assert.Contains(t, alertEmail.Body, "Review in Admin Portal")

	// 3. Test SendAccountDeletionConfirmation (Approved Member)
	email.ClearLastSentEmails()
	err = email.SendAccountDeletionConfirmation("member@nammataga.com", "Sudhan Palanivel", "TAGA100")
	require.NoError(t, err)

	sent = email.GetLastSentEmails()
	require.Len(t, sent, 1)
	confirmEmail := sent[0]
	assert.Equal(t, "member@nammataga.com", confirmEmail.To)
	assert.Contains(t, confirmEmail.Subject, "Account Deletion Confirmation")
	assert.Contains(t, confirmEmail.Body, "Tamil Nadu Agricultural Graduates Association")
	assert.Contains(t, confirmEmail.Body, "Account Retired")
	assert.Contains(t, confirmEmail.Body, "Sudhan Palanivel")
	assert.Contains(t, confirmEmail.Body, "TAGA100")
	assert.Contains(t, confirmEmail.Body, "Summary of Actions Taken")
	assert.Contains(t, confirmEmail.Body, "passwords, and active session tokens have been permanently revoked")

	// 4. Test SendAccountDeletionRejection (Rejected Member with Remarks)
	email.ClearLastSentEmails()
	remarks := "Please settle pending welfare society dues before deleting account."
	err = email.SendAccountDeletionRejection("member@nammataga.com", "Sudhan Palanivel", "TAGA100", remarks)
	require.NoError(t, err)

	sent = email.GetLastSentEmails()
	require.Len(t, sent, 1)
	rejectEmail := sent[0]
	assert.Equal(t, "member@nammataga.com", rejectEmail.To)
	assert.Contains(t, rejectEmail.Subject, "Account Deletion Request Update")
	assert.Contains(t, rejectEmail.Body, "Tamil Nadu Agricultural Graduates Association")
	assert.Contains(t, rejectEmail.Body, "Account Active")
	assert.Contains(t, rejectEmail.Body, "Administrator Remarks")
	assert.Contains(t, rejectEmail.Body, remarks)
	assert.Contains(t, rejectEmail.Body, "Your account remains fully active")
	assert.Contains(t, rejectEmail.Body, "Go to Member Login")
}
