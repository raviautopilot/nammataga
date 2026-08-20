package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"taga-api/service/audit"

	"github.com/gin-gonic/gin"
)

// validIDRe matches safe tagaId characters (same rule as the storage layer).
var validIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)

// auditQueryParams holds the validated query parameters for the audit API.
type auditQueryParams struct {
	TagaID string // optional: "T-001", "admin", "anonymous"
	Year   string // optional: "2026"
	Month  string // optional: "08"
	Action string // optional: "LOGIN"
	Module string // optional: "MEMBER"
	Search string // optional: free text
	Page   int    // 1-indexed
	Limit  int    // max 200
}

// GetAuditLogsHandler godoc
// @Summary Get audit records
// @Description Serves audit records to authorized administrators with filtering and pagination
// @Tags Admin Audit
// @Produce json
// @Security BearerAuth
// @Param taga_id query string false "Member TAGA ID"
// @Param year query string false "Year (YYYY)"
// @Param month query string false "Month (MM)"
// @Param action query string false "Action filter"
// @Param module query string false "Module filter"
// @Param search query string false "Search keyword"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Limit per page (default 50, max 200)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/audit [get]
func GetAuditLogsHandler(c *gin.Context) {
	params, err := parseAuditQuery(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	records, err := loadAuditRecords(params)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read audit logs")
		return
	}

	// Apply action / module / search filters in-memory
	filtered := filterRecords(records, params)

	// Sort by timestamp descending (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp > filtered[j].Timestamp
	})

	total := len(filtered)
	start := (params.Page - 1) * params.Limit
	if start >= total {
		start = total
	}
	end := start + params.Limit
	if end > total {
		end = total
	}

	respondOK(c, gin.H{
		"data":  filtered[start:end],
		"page":  params.Page,
		"limit": params.Limit,
		"total": total,
	})
}

// GetAuditUsersHandler godoc
// @Summary Get audit user list
// @Description Returns the list of user IDs (taga-ids) that have audit files for the given year/month
// @Tags Admin Audit
// @Produce json
// @Security BearerAuth
// @Param year query string false "Year (YYYY)"
// @Param month query string false "Month (MM)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/admin/audit/users [get]
func GetAuditUsersHandler(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")

	if year != "" {
		if _, err := strconv.Atoi(year); err != nil || len(year) != 4 {
			respondError(c, http.StatusBadRequest, "Invalid year")
			return
		}
	}
	if month != "" {
		m, err := strconv.Atoi(month)
		if err != nil || m < 1 || m > 12 {
			respondError(c, http.StatusBadRequest, "Invalid month")
			return
		}
		month = fmt.Sprintf("%02d", m)
	}

	users := collectAuditUsers(year, month)
	respondOK(c, gin.H{"users": users})
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func parseAuditQuery(c *gin.Context) (auditQueryParams, error) {
	var p auditQueryParams

	p.TagaID = c.Query("taga_id")
	if p.TagaID != "" && !validIDRe.MatchString(p.TagaID) {
		return p, fmt.Errorf("invalid taga_id: only alphanumeric, dash, underscore allowed")
	}

	p.Year = c.Query("year")
	if p.Year != "" {
		y, err := strconv.Atoi(p.Year)
		if err != nil || y < 2000 || y > 2100 {
			return p, fmt.Errorf("invalid year")
		}
	}

	p.Month = c.Query("month")
	if p.Month != "" {
		m, err := strconv.Atoi(p.Month)
		if err != nil || m < 1 || m > 12 {
			return p, fmt.Errorf("invalid month")
		}
		p.Month = fmt.Sprintf("%02d", m)
	}

	p.Action = strings.ToUpper(c.Query("action"))
	p.Module = strings.ToUpper(c.Query("module"))
	p.Search = c.Query("search")

	p.Page = 1
	if pg := c.Query("page"); pg != "" {
		if n, err := strconv.Atoi(pg); err == nil && n > 0 {
			p.Page = n
		}
	}

	p.Limit = 50
	if lm := c.Query("limit"); lm != "" {
		if n, err := strconv.Atoi(lm); err == nil && n > 0 && n <= 200 {
			p.Limit = n
		}
	}

	return p, nil
}

// loadAuditRecords reads audit files matching the query params.
// Uses the most specific path available to avoid scanning unnecessary files.
func loadAuditRecords(p auditQueryParams) ([]audit.AuditRecord, error) {
	var files []string

	switch {
	case p.TagaID != "" && p.Year != "" && p.Month != "":
		// Most specific: single file
		f := filepath.Join("audit-logs", p.Year, p.Month, fmt.Sprintf("user_%s.json", p.TagaID))
		files = []string{f}

	case p.Year != "" && p.Month != "":
		// One month directory
		pattern := filepath.Join("audit-logs", p.Year, p.Month, "user_*.json")
		var err error
		files, err = filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

	case p.Year != "":
		// All months in a year
		pattern := filepath.Join("audit-logs", p.Year, "??", "user_*.json")
		var err error
		files, err = filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

	default:
		// All available files (last resort — bounded by retention)
		pattern := filepath.Join("audit-logs", "????", "??", "user_*.json")
		var err error
		files, err = filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
	}

	var all []audit.AuditRecord
	for _, f := range files {
		// Extra safety: ensure the resolved path stays inside audit-logs/
		clean := filepath.Clean(f)
		if !strings.HasPrefix(clean, "audit-logs"+string(os.PathSeparator)) {
			continue
		}

		raw, err := os.ReadFile(clean)
		if err != nil {
			continue // file may not exist yet; skip silently
		}

		var records []audit.AuditRecord
		if err := json.Unmarshal(raw, &records); err != nil {
			continue // corrupt file; skip
		}
		all = append(all, records...)
	}
	return all, nil
}

// filterRecords applies action, module, and free-text search filters.
func filterRecords(records []audit.AuditRecord, p auditQueryParams) []audit.AuditRecord {
	search := strings.ToLower(p.Search)

	var out []audit.AuditRecord
	for _, r := range records {
		if p.Action != "" && r.Action != p.Action {
			continue
		}
		if p.Module != "" && r.Module != p.Module {
			continue
		}
		if search != "" {
			haystack := strings.ToLower(fmt.Sprintf("%s %s %s %s %s %s",
				r.Username, r.Description, r.ResourceID, r.ResourceType, r.Module, r.Action))
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// collectAuditUsers scans the audit directory and returns unique user IDs.
func collectAuditUsers(year, month string) []string {
	var pattern string
	switch {
	case year != "" && month != "":
		pattern = filepath.Join("audit-logs", year, month, "user_*.json")
	case year != "":
		pattern = filepath.Join("audit-logs", year, "??", "user_*.json")
	default:
		pattern = filepath.Join("audit-logs", "????", "??", "user_*.json")
	}

	files, _ := filepath.Glob(pattern)
	seen := make(map[string]bool)
	for _, f := range files {
		base := filepath.Base(f)
		// "user_T-001.json" → "T-001"
		if strings.HasPrefix(base, "user_") && strings.HasSuffix(base, ".json") {
			uid := strings.TrimSuffix(strings.TrimPrefix(base, "user_"), ".json")
			seen[uid] = true
		}
	}

	users := make([]string, 0, len(seen))
	for u := range seen {
		users = append(users, u)
	}
	sort.Strings(users)
	return users
}
