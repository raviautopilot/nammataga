package audit

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"taga-api/config"
)

// ─── Action constants ────────────────────────────────────────────────────────

const (
	ActionLogin             = "LOGIN"
	ActionLoginFailed       = "LOGIN_FAILED"
	ActionLogout            = "LOGOUT"
	ActionPasswordChanged   = "PASSWORD_CHANGED"
	ActionCreate            = "CREATE"
	ActionUpdate            = "UPDATE"
	ActionDelete            = "DELETE"
	ActionRoleChanged       = "ROLE_CHANGED"
	ActionPermissionChanged = "PERMISSION_CHANGED"
	ActionBookingCreated    = "BOOKING_CREATED"
	ActionBookingCancelled  = "BOOKING_CANCELLED"
	ActionPaymentConfirmed  = "PAYMENT_CONFIRMED"
)

// ─── Module constants ────────────────────────────────────────────────────────

const (
	ModuleAuth     = "AUTH"
	ModuleMember   = "MEMBER"
	ModuleBooking  = "BOOKING"
	ModulePayment  = "PAYMENT"
	ModuleResource = "RESOURCE"
	ModuleEvent    = "EVENT"
	ModuleGallery  = "GALLERY"
)

// ─── Record structure ────────────────────────────────────────────────────────

// AuditRecord represents a single auditable action.
type AuditRecord struct {
	AuditID      string      `json:"audit_id"`
	UserID       string      `json:"user_id"`               // tagaId (T-001), "admin", or "anonymous"
	Username     string      `json:"username"`              // email / admin username
	Action       string      `json:"action"`
	Module       string      `json:"module"`
	ResourceType string      `json:"resource_type,omitempty"`
	ResourceID   string      `json:"resource_id,omitempty"` // tagaId of affected resource
	Description  string      `json:"description"`
	OldData      interface{} `json:"old_data,omitempty"`
	NewData      interface{} `json:"new_data,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	Timestamp    string      `json:"timestamp"` // RFC3339
}

// ─── Public API ──────────────────────────────────────────────────────────────

// Log creates and persists an audit record synchronously.
//
//   - c          – Gin context (may be nil for background jobs)
//   - userID     – tagaId of the actor (e.g. "T-001", "admin", "anonymous")
//   - username   – human-readable name / email of the actor
//   - action     – one of the Action* constants above
//   - module     – one of the Module* constants above
//   - resourceType / resourceID – what was affected (may be empty)
//   - description – short human-readable sentence
//   - oldData / newData – sanitized before storage; may be nil
//
// Returns an error so callers can decide whether the business operation
// should fail when audit logging fails.  At minimum, always log failures.
func Log(
	c *gin.Context,
	userID, username,
	action, module,
	resourceType, resourceID, description string,
	oldData, newData interface{},
) error {
	if userID == "" {
		userID = "anonymous"
	}
	if username == "" {
		username = userID
	}

	ip, ua := extractRequestMeta(c)

	record := AuditRecord{
		AuditID:      uuid.New().String(),
		UserID:       sanitizeID(userID),
		Username:     username,
		Action:       action,
		Module:       module,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Description:  description,
		OldData:      Sanitize(oldData),
		NewData:      Sanitize(newData),
		IPAddress:    ip,
		UserAgent:    ua,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	if err := save(record); err != nil {
		if config.Logger != nil {
			config.Logger.Error("Failed to write audit record",
				zap.String("action", action),
				zap.String("user_id", userID),
				zap.String("module", module),
				zap.Error(err),
			)
		}
		return fmt.Errorf("audit write failed: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Debug("Audit record written",
			zap.String("action", action),
			zap.String("user_id", userID),
			zap.String("module", module),
			zap.String("description", description),
		)
	}
	return nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// extractRequestMeta pulls IP and User-Agent from the Gin context safely.
func extractRequestMeta(c *gin.Context) (ip, ua string) {
	if c == nil {
		return "", ""
	}
	return c.ClientIP(), c.GetHeader("User-Agent")
}
