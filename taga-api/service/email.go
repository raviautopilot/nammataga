package service

import (
	"fmt"
	"net/smtp"

	"taga-api/config"
)

func SendGrievanceEmail(
	subject string,
	category string,
	priority string,
	description string,
	phone string,
) error {

	cfg := config.Config

	auth := smtp.PlainAuth(
		"",
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPHost,
	)

	message := fmt.Sprintf(
		`Subject: New Grievance Submitted

New grievance received.

Subject:
%s

Category:
%s

Priority:
%s

Phone:
%s

Description:
%s
`,
		subject,
		category,
		priority,
		phone,
		description,
	)

	return smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.FromEmail,
		[]string{cfg.AdminEmail},
		[]byte(message),
	)
}
