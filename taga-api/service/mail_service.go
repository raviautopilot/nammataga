package service

import (
	"encoding/json"
	"fmt"
	"net/smtp"

	"taga-api/config"
	"taga-api/model"

	"go.uber.org/zap"
)

func SendEditRequestEmail(req model.EditRequest) error {

	cfg := config.Config

	config.Logger.Debug("Email configuration details",
		zap.String("host", cfg.SMTPHost),
		zap.Int("port", cfg.SMTPPort),
		zap.String("from", cfg.FromEmail),
		zap.String("admin", cfg.AdminEmail),
		zap.String("user", cfg.SMTPUsername),
	)

	auth := smtp.PlainAuth(
		"",
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPHost,
	)

	// Convert edit request into JSON
	requestJSON, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}

	message := fmt.Sprintf(
		`Subject: New Profile Edit Request

New profile edit request received.

Request Data:

%s
`,
		string(requestJSON),
	)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.FromEmail,
		[]string{cfg.AdminEmail},
		[]byte(message),
	)

	if err != nil {
		config.Logger.Error("Failed to send edit request email", zap.Error(err))
		return err
	}

	config.Logger.Info("Email sent successfully", zap.String("to", cfg.AdminEmail))

	return nil

}
