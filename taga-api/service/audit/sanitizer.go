package audit

import (
	"encoding/json"
	"strings"
)

// sensitiveFields is the authoritative list of field name substrings that must
// never appear in audit logs with their real values.
var sensitiveFields = []string{
	"password",
	"passwd",
	"pass",
	"token",
	"access_token",
	"refresh_token",
	"secret",
	"client_secret",
	"api_key",
	"apikey",
	"authorization",
	"jwt",
	"otp",
	"private_key",
	"hash",
}

// Sanitize converts any value to a sanitized generic structure (via a JSON
// round-trip) and recursively replaces sensitive field values with "[REDACTED]".
//
// It is safe to pass structs, maps, slices, or nil.
func Sanitize(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	// Round-trip through JSON to get a fully generic representation.
	// This ensures struct fields with json:"-" tags are excluded automatically.
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var generic interface{}
	if err := json.Unmarshal(b, &generic); err != nil {
		return nil
	}
	return sanitizeValue(generic)
}

// sanitizeValue recursively traverses maps and slices, redacting sensitive keys.
func sanitizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(val))
		for k, vv := range val {
			if isSensitiveKey(k) {
				out[k] = "[REDACTED]"
			} else {
				out[k] = sanitizeValue(vv)
			}
		}
		return out

	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = sanitizeValue(item)
		}
		return out

	default:
		return v
	}
}

// isSensitiveKey returns true if the lowercased key contains any of the
// sensitive substrings. String comparison only — no filesystem operations.
func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, s := range sensitiveFields {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
