package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"e2e-template/pkg/client"
	"e2e-template/tests"
)

type RestoreRequest struct {
	BackupFile string `json:"backup_file"`
}

func TestAPI_AdminOffice_BackupsE2E(t *testing.T) {
	tests.RunAPITestWithDetails(t, "[Admin] POST Restore Backup - Validation Error", "Submits restore request without mandatory backup_file payload field.", "HTTP 400 Bad Request", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/office-bearers/backup/restore", nil, &RestoreRequest{}, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusBadRequest, "backup_file is required")
	})

	tests.RunAPITestWithDetails(t, "[Admin] POST Restore Backup - Nonexistent File Error", "Attempts restoring office bearers using a nonexistent backup file path.", "HTTP 500 Internal Server Error", func(tctx *tests.TestContext) {
		token := getValidAdminToken(tctx.T, tctx.Client)
		auth := &client.BearerTokenAuth{Token: token}

		payload := &RestoreRequest{
			BackupFile: "nonexistent-backup-2026.json",
		}

		var resp map[string]interface{}
		err := tctx.Client.SendHttpRequest("POST", "/api/admin/office-bearers/backup/restore", nil, payload, &resp, auth)

		assertErrorStatus(tctx, err, http.StatusInternalServerError, "Failed to restore")
	})
}

func TestAPI_NegativeScenarios_OfficeBearersBackup(t *testing.T) {
	type TestCaseType struct {
		Name           string
		Persona        string
		Description    string
		Method         string
		Endpoint       string
		AuthType       string
		Payload        interface{}
		Headers        map[string]string
		ExpectedStatus int
		ExpectedSub    string
	}

	testCases := []TestCaseType{
		{
			Name:           "Missing Auth - Error Expected",
			Persona:        "Anonymous",
			Description:    "Restore backup without token",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "none",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Invalid Token - Error Expected",
			Persona:        "Attacker",
			Description:    "Access with fake token",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "invalid",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "Wrong HTTP Method - Error Expected",
			Persona:        "Admin",
			Description:    "Use GET instead of POST",
			Method:         "GET",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "admin",
			ExpectedStatus: http.StatusMethodNotAllowed,
		},
		{
			Name:           "SQLi Payload - Error Expected",
			Persona:        "Admin",
			Description:    "SQLi attempt in payload",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "admin",
			Payload:        map[string]string{"backup_file": "backup' OR '1'='1"},
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			Name:           "Path Traversal Payload - Error Expected",
			Persona:        "Admin",
			Description:    "Path traversal in backup filename",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "admin",
			Payload:        map[string]string{"backup_file": "../../../etc/passwd"},
			ExpectedStatus: http.StatusInternalServerError, // Or 400
		},
		{
			Name:           "Malformed JSON Payload - Error Expected",
			Persona:        "Admin",
			Description:    "Send non-JSON data where expected",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "admin",
			Payload:        "not-a-json-payload-string",
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "Role Context Switching - Member Restoring Backup",
			Persona:        "Member",
			Description:    "Member attempts to restore system backup",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "member",
			Payload:        map[string]string{"backup_file": "backup-2026.json"},
			ExpectedStatus: http.StatusForbidden,
		},
		{
			Name:           "Logical Paradox - Future Backup File",
			Persona:        "Admin",
			Description:    "Restore a backup that claims to be from the future",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "admin",
			Payload:        map[string]string{"backup_file": "backup-2099-01-01.json"},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name:           "State Machine Violation - Concurrent Restores",
			Persona:        "Admin",
			Description:    "Restore when a restore is supposedly already in progress",
			Method:         "POST",
			Endpoint:       "/api/admin/office-bearers/backup/restore",
			AuthType:       "admin",
			Payload:        map[string]string{"backup_file": "locked_backup_in_progress.json"},
			ExpectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		tc := tc
		testName := fmt.Sprintf("[%s] %s %s - %s", tc.Persona, tc.Method, tc.Endpoint, tc.Name)
		expectedStr := fmt.Sprintf("HTTP %d", tc.ExpectedStatus)
		if tc.ExpectedSub != "" {
			expectedStr += fmt.Sprintf(" containing '%s'", tc.ExpectedSub)
		}

		tests.RunAPITestWithDetails(t, testName, tc.Description, expectedStr, func(tctx *tests.TestContext) {
			var auth client.Authenticator
			switch tc.AuthType {
			case "admin":
				token := getValidAdminToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "member":
				token := getValidMemberToken(tctx.T, tctx.Client)
				auth = &client.BearerTokenAuth{Token: token}
			case "invalid":
				auth = &client.BearerTokenAuth{Token: "invalid-jwt-token-12345"}
			case "none":
				auth = nil
			}

			var resp map[string]interface{}
			err := tctx.Client.SendHttpRequest(tc.Method, tc.Endpoint, tc.Headers, tc.Payload, &resp, auth)
			if err == nil {
				tctx.Fatalf("Expected error for negative scenario, got none. Response: %v", resp)
			}
		})
	}
}
