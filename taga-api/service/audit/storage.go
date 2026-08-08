package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

// mu protects all audit file read-modify-write operations.
//
// NOTE: This is a process-local mutex and is sufficient for the current
// single-container Docker deployment.  If multiple backend containers are
// ever introduced that share the same filesystem, this must be replaced with
// a cross-process locking mechanism (e.g. advisory file locks via syscall.Flock).
var mu sync.Mutex

// validIDPattern whitelists characters allowed in a tagaId used as a filename.
var validIDPattern = regexp.MustCompile(`[^A-Za-z0-9_\-]`)

// save atomically appends record to the user's monthly audit file.
//
// The write sequence is:
//  1. Determine year/month directory and create it if needed.
//  2. Read existing records (empty slice if file does not yet exist).
//  3. Append the new record.
//  4. Marshal to JSON.
//  5. Write to a .tmp file.
//  6. os.Rename (.tmp → final) — atomic on Linux/POSIX.
//
// The atomic rename prevents partially-written JSON from reaching the real file.
func save(record AuditRecord) error {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	dir := filepath.Join("audit-logs", now.Format("2006"), now.Format("01"))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create audit dir: %w", err)
	}

	safeID := sanitizeID(record.UserID)
	filePath := filepath.Join(dir, fmt.Sprintf("user_%s.json", safeID))
	tmpPath := filePath + ".tmp"

	// Read existing records
	var records []AuditRecord
	if raw, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(raw, &records) // tolerate a corrupt file: start fresh
	}

	records = append(records, record)

	out, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal audit records: %w", err)
	}

	// Write to tmp then atomically rename
	if err := os.WriteFile(tmpPath, out, 0644); err != nil {
		return fmt.Errorf("write audit tmp file: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath) // clean up orphaned tmp
		return fmt.Errorf("rename audit file: %w", err)
	}

	return nil
}

// sanitizeID strips path-traversal characters and limits length so the value
// can be safely embedded in a filename.
// e.g. "T-001" → "T-001", "../../etc" → "etcpasswd" (rendered harmless)
func sanitizeID(id string) string {
	if id == "" {
		return "anonymous"
	}
	safe := validIDPattern.ReplaceAllString(id, "")
	if safe == "" {
		return "anonymous"
	}
	if len(safe) > 64 {
		safe = safe[:64]
	}
	return safe
}
