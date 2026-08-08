package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	// Test basic structures
	assert.Nil(t, Sanitize(nil))

	// Test map with sensitive fields
	inputMap := map[string]interface{}{
		"username": "testuser",
		"password": "secretpassword",
		"nested": map[string]interface{}{
			"api_key":      "super-secret-key",
			"normal_field": 1234,
		},
		"slice_field": []interface{}{
			map[string]interface{}{
				"token": "sensitive-token",
			},
			"normal-string",
		},
	}

	sanitized := Sanitize(inputMap)
	m, ok := sanitized.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", m["username"])
	assert.Equal(t, "[REDACTED]", m["password"])

	nested, ok := m["nested"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "[REDACTED]", nested["api_key"])
	assert.Equal(t, float64(1234), nested["normal_field"]) // JSON unmarshals ints to float64

	slice, ok := m["slice_field"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, slice, 2)
	itemMap, ok := slice[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "[REDACTED]", itemMap["token"])
	assert.Equal(t, "normal-string", slice[1])
}

func TestSanitizeFilenameID(t *testing.T) {
	assert.Equal(t, "anonymous", sanitizeID(""))
	assert.Equal(t, "T-001", sanitizeID("T-001"))
	assert.Equal(t, "user_42", sanitizeID("user_42"))
	assert.Equal(t, "admin", sanitizeID("admin"))
	assert.Equal(t, "etcpasswd", sanitizeID("../../etc/passwd"))
	assert.Equal(t, "T-001abc", sanitizeID("T-001/../abc"))
}

func TestLogAndCleanup(t *testing.T) {
	// Temporarily redirect audit-logs directory prefix or clean up files after test
	now := time.Now()
	yearStr := now.Format("2006")
	monthStr := now.Format("01")
	testUserID := "T-TestUnit"

	targetDir := filepath.Join("audit-logs", yearStr, monthStr)
	targetFile := filepath.Join(targetDir, "user_"+testUserID+".json")

	// Ensure clean slate
	_ = os.Remove(targetFile)

	// Log an action
	err := Log(nil, testUserID, "unit-tester@test.com", ActionLogin, ModuleAuth, "test", "123", "Unit test action", nil, nil)
	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(targetFile)
	assert.NoError(t, err)

	// Read and parse file
	data, err := os.ReadFile(targetFile)
	assert.NoError(t, err)

	var records []AuditRecord
	err = json.Unmarshal(data, &records)
	assert.NoError(t, err)
	assert.Len(t, records, 1)

	assert.Equal(t, testUserID, records[0].UserID)
	assert.Equal(t, "unit-tester@test.com", records[0].Username)
	assert.Equal(t, ActionLogin, records[0].Action)
	assert.Equal(t, ModuleAuth, records[0].Module)
	assert.Equal(t, "Unit test action", records[0].Description)

	// Cleanup the test file
	_ = os.Remove(targetFile)

	// Clean up empty directories if possible
	_ = os.Remove(targetDir)
	_ = os.Remove(filepath.Join("audit-logs", yearStr))
}

func TestRetentionMonths(t *testing.T) {
	os.Setenv("AUDIT_LOG_RETENTION_MONTHS", "6")
	assert.Equal(t, 6, retentionMonths())

	os.Setenv("AUDIT_LOG_RETENTION_MONTHS", "invalid")
	assert.Equal(t, 3, retentionMonths())

	os.Unsetenv("AUDIT_LOG_RETENTION_MONTHS")
}
