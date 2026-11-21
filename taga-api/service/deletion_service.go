package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"taga-api/config"
	"taga-api/model"
	"taga-api/service/audit"
	"taga-api/service/email"
	"taga-api/utils"
)

// Active stay error constant and instance
const ActiveStayMessage = "You are currently staying at TAGA Towers. You can request account deletion after completing your checkout."
const PendingAccountMessage = "Your account deletion request is awaiting administrator review. You cannot access your account while the request is pending."

var (
	ErrActiveStay = errors.New(ActiveStayMessage)
	deletionLock  sync.Mutex
)

// HasActiveStayAtTagaTowers checks if the member has an active stay at TAGA Towers.
// Rule: payment_status == 'confirmed' AND checkInDate <= now < checkOutDate
func HasActiveStayAtTagaTowers(memberID, tagaID string) (bool, error) {
	bookingsFile := config.Config.Data.Tower.BookingsFile
	if bookingsFile == "" {
		bookingsFile = filepath.Join("data", "towers", "bookings.json")
	}

	data, err := os.ReadFile(bookingsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var bookings []model.Booking
	if err := json.Unmarshal(data, &bookings); err != nil {
		return false, err
	}

	now := time.Now()
	todayStr := now.Format("2006-01-02")

	for _, b := range bookings {
		isBooker := (memberID != "" && b.BookerID == memberID) ||
			(tagaID != "" && strings.EqualFold(b.BookerID, tagaID))

		if !isBooker {
			continue
		}

		// Payment status must be confirmed
		isConfirmed := b.PaymentStatus == model.PaymentConfirmed ||
			strings.EqualFold(string(b.PaymentStatus), "confirmed")

		if !isConfirmed {
			continue
		}

		checkInStr := b.CheckInDate.Format("2006-01-02")
		checkOutStr := b.CheckOutDate.Format("2006-01-02")

		// checkInDate <= now < checkOutDate
		// Evaluated with both datetime and calendar date boundaries
		isActive := (!now.Before(b.CheckInDate) && now.Before(b.CheckOutDate)) ||
			(checkInStr <= todayStr && todayStr < checkOutStr)

		if isActive {
			return true, nil
		}
	}

	return false, nil
}

// CancelUpcomingBookingsForMember cancels eligible upcoming bookings for a member and releases inventory.
func CancelUpcomingBookingsForMember(memberID, tagaID string) (int, error) {
	bookingsFile := config.Config.Data.Tower.BookingsFile
	if bookingsFile == "" {
		bookingsFile = filepath.Join("data", "towers", "bookings.json")
	}

	data, err := os.ReadFile(bookingsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	var bookings []model.Booking
	if err := json.Unmarshal(data, &bookings); err != nil {
		return 0, err
	}

	now := time.Now()
	todayStr := now.Format("2006-01-02")
	cancelledCount := 0

	for i := range bookings {
		b := &bookings[i]
		isBooker := (memberID != "" && b.BookerID == memberID) ||
			(tagaID != "" && strings.EqualFold(b.BookerID, tagaID))

		if !isBooker {
			continue
		}

		if b.PaymentStatus == model.PaymentCancelled || b.PaymentStatus == model.PaymentRefunded {
			continue
		}

		checkInStr := b.CheckInDate.Format("2006-01-02")
		// Eligible upcoming booking: check-in is in future or today
		if now.Before(b.CheckInDate) || checkInStr >= todayStr {
			b.PaymentStatus = model.PaymentCancelled
			cancelledCount++
		}
	}

	if cancelledCount > 0 {
		updatedData, err := json.MarshalIndent(bookings, "", "  ")
		if err != nil {
			return 0, err
		}
		if err := utils.WriteFileAtomic(bookingsFile, updatedData, 0644); err != nil {
			return 0, err
		}
	}

	return cancelledCount, nil
}

func getMemberSubscriptionsFilePath() string {
	cfg := config.GetConfig()
	if cfg.MembersFile != "" {
		subsFile := filepath.Join(filepath.Dir(cfg.MembersFile), "..", "subscriptions", "member_subscriptions.json")
		if _, err := os.Stat(subsFile); err == nil {
			return subsFile
		}
	}
	return filepath.Join("data", "subscriptions", "member_subscriptions.json")
}

// CancelActiveSubscriptionsForMember cancels any active subscription for the member.
func CancelActiveSubscriptionsForMember(memberID, tagaID, email string) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subsPath := GetMemberSubscriptionsFilePath()
	data, err := os.ReadFile(subsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var subs []map[string]interface{}
	if err := json.Unmarshal(data, &subs); err != nil {
		return err
	}

	modified := false
	nowStr := time.Now().Format(time.RFC3339)

	for i := range subs {
		s := subs[i]
		subMemberID := utils.GetStringFromMap(s, "member_id")
		subEmail := utils.GetStringFromMap(s, "member_email")

		matches := (memberID != "" && subMemberID == memberID) ||
			(tagaID != "" && strings.EqualFold(subMemberID, tagaID)) ||
			(email != "" && strings.EqualFold(subEmail, email))

		if matches {
			status := utils.GetStringFromMap(s, "status")
			if strings.EqualFold(status, "active") {
				subs[i]["status"] = "cancelled_by_account_deletion"
				subs[i]["updated_at"] = nowStr
				modified = true
			}
		}
	}

	if modified {
		updatedData, err := json.MarshalIndent(subs, "", "  ")
		if err != nil {
			return err
		}
		return utils.WriteFileAtomic(subsPath, updatedData, 0644)
	}

	return nil
}

// ExecuteAccountDeletion coordinates complete, atomic member account deletion.
// Used by both Admin Approval and Admin Direct Deletion.
func ExecuteAccountDeletion(c *gin.Context, memberID, tagaID, deletedBy, deletionReason, requestID string) error {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	// 1. Enforce active stay check
	hasActiveStay, err := HasActiveStayAtTagaTowers(memberID, tagaID)
	if err != nil {
		config.Logger.Error("Failed to check active stay status", zap.Error(err))
		return err
	}
	if hasActiveStay {
		return ErrActiveStay
	}

	cfg := config.GetConfig()
	membersFile := cfg.MembersFile
	if membersFile == "" {
		membersFile = "data/member/members.json"
	}

	data, err := os.ReadFile(membersFile)
	if err != nil {
		return fmt.Errorf("failed to read members file: %w", err)
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return fmt.Errorf("failed to parse members file: %w", err)
	}

	// 2. Find target member and build filtered list
	var targetMember map[string]interface{}
	targetIdx := -1
	cleanMemberID := strings.TrimSpace(memberID)
	cleanTagaID := strings.TrimSpace(tagaID)

	for idx, m := range members {
		id := strings.TrimSpace(utils.GetStringFromMap(m, "id"))
		mTag := strings.TrimSpace(utils.GetStringFromMap(m, "tagaId"))
		mEmail := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(m, "emailId")))
		if mEmail == "" {
			mEmail = strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(m, "email")))
		}

		matchesID := cleanMemberID != "" && id == cleanMemberID
		matchesTaga := cleanTagaID != "" && strings.EqualFold(mTag, cleanTagaID)

		if matchesID || matchesTaga {
			targetMember = m
			targetIdx = idx
			break
		}
	}

	if targetMember == nil {
		return fmt.Errorf("member not found")
	}

	memberUUID := utils.GetStringFromMap(targetMember, "id")
	memberTagaID := utils.GetStringFromMap(targetMember, "tagaId")
	memberName := utils.GetStringFromMap(targetMember, "name")
	memberEmail := utils.GetStringFromMap(targetMember, "emailId")
	if memberEmail == "" {
		memberEmail = utils.GetStringFromMap(targetMember, "email")
	}
	if memberEmail == "" {
		memberEmail = utils.GetStringFromMap(targetMember, "username")
	}

	// 3. Prepare archived copy with CRITICAL SECURITY FIX:
	// Strip all authentication secrets before archiving
	archivedMember := make(map[string]interface{})
	for k, v := range targetMember {
		archivedMember[k] = v
	}

	secretKeys := []string{
		"password",
		"salt",
		"token",
		"otp",
		"temp_password",
		"session_secret",
		"reset_token",
		"reset_password_token",
		"verification_token",
		"auth_token",
	}
	for _, sec := range secretKeys {
		delete(archivedMember, sec)
	}

	nowRFC3339 := time.Now().Format(time.RFC3339)
	archivedMember["deleted_at"] = nowRFC3339
	archivedMember["deleted_by"] = deletedBy
	if deletionReason == "" {
		deletionReason = "Account deletion"
	}
	archivedMember["deletion_reason"] = deletionReason

	// 4. Save to deleted_member.json atomically (deduplicating if previously archived)
	deletedFile := cfg.DeletedMembersFile
	if deletedFile == "" {
		deletedFile = "data/member/deleted_member.json"
	}

	var deletedMembers []map[string]interface{}
	if dData, err := os.ReadFile(deletedFile); err == nil && len(dData) > 0 {
		_ = json.Unmarshal(dData, &deletedMembers)
	}

	// Filter out any stale archived record matching this member ID, tagaId, or email
	filteredDeleted := make([]map[string]interface{}, 0, len(deletedMembers))
	for _, dm := range deletedMembers {
		dID := strings.TrimSpace(utils.GetStringFromMap(dm, "id"))
		dTaga := strings.TrimSpace(utils.GetStringFromMap(dm, "tagaId"))
		dEmail := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(dm, "emailId")))
		if dEmail == "" {
			dEmail = strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(dm, "email")))
		}

		matchesOld := (memberUUID != "" && dID == memberUUID) ||
			(memberTagaID != "" && strings.EqualFold(dTaga, memberTagaID)) ||
			(memberEmail != "" && dEmail == strings.ToLower(memberEmail))

		if !matchesOld {
			filteredDeleted = append(filteredDeleted, dm)
		}
	}
	filteredDeleted = append(filteredDeleted, archivedMember)

	deletedData, err := json.MarshalIndent(filteredDeleted, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deleted members: %w", err)
	}
	if err := utils.WriteFileAtomic(deletedFile, deletedData, 0644); err != nil {
		return fmt.Errorf("failed to write deleted members file: %w", err)
	}

	// 5. Remove member from members.json atomically
	filteredMembers := append(members[:targetIdx], members[targetIdx+1:]...)
	updatedMembersData, err := json.MarshalIndent(filteredMembers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}
	if err := utils.WriteFileAtomic(membersFile, updatedMembersData, 0644); err != nil {
		return fmt.Errorf("failed to write members file: %w", err)
	}

	// 6. Update active subscriptions to cancelled_by_account_deletion
	if err := CancelActiveSubscriptionsForMember(memberUUID, memberTagaID, memberEmail); err != nil {
		config.Logger.Warn("Failed to cancel member subscriptions during deletion", zap.Error(err))
	}

	// 7. Cancel eligible upcoming bookings at TAGA Towers and release inventory
	if _, err := CancelUpcomingBookingsForMember(memberUUID, memberTagaID); err != nil {
		config.Logger.Warn("Failed to cancel upcoming bookings during deletion", zap.Error(err))
	}

	// CRITICAL ACCOUNTING FIX:
	// DO NOT touch payment_transactions.json or processed_payments.json!
	// Existing transaction history is preserved intact.

	// 8. Update deletion request in deletion_requests.json if request exists
	_ = updateDeletionRequestStatus(requestID, memberUUID, memberTagaID, model.DeletionStatusApproved, deletedBy, "")

	// 9. Audit log
	resourceID := memberTagaID
	if resourceID == "" {
		resourceID = memberUUID
	}
	if c != nil {
		_ = audit.Log(c, "admin", deletedBy,
			audit.ActionDelete, audit.ModuleMember,
			"member", resourceID,
			fmt.Sprintf("Admin '%s' deleted member '%s' (tagaId: %s, reason: %s)", deletedBy, memberName, resourceID, deletionReason),
			audit.Sanitize(targetMember), nil)
	}

	// 10. Send confirmation email to member
	if memberEmail != "" {
		go func(to, name, taga string) {
			if err := email.SendAccountDeletionConfirmation(to, name, taga); err != nil {
				if config.Logger != nil {
					config.Logger.Warn("Failed to send account deletion confirmation email",
						zap.String("to", to), zap.Error(err))
				}
			}
		}(memberEmail, memberName, memberTagaID)
	}

	return nil
}

// ResolveMemberUID implements UID reuse during member registration.
// Before generating a new UUID, query data/member/deleted_member.json.
// Compare submitted TAGA ID and username/email:
// If BOTH match the same archived member record, reuse that archived member's original UID (id).
// If either or both do not match, generate a fresh uuid.New().String().
// If conflict/multiple match, log a warning and generate a new UUID.
// Verify that the reused UID does not conflict with any active member.
func ResolveMemberUID(tagaID, email string, activeMembers []map[string]interface{}) string {
	cleanTagaID := strings.TrimSpace(tagaID)
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	if cleanTagaID == "" || cleanEmail == "" {
		return uuid.New().String()
	}

	deletedFile := config.Config.DeletedMembersFile
	if deletedFile == "" {
		deletedFile = "data/member/deleted_member.json"
	}

	data, err := os.ReadFile(deletedFile)
	if err != nil {
		return uuid.New().String()
	}

	var archivedMembers []map[string]interface{}
	if err := json.Unmarshal(data, &archivedMembers); err != nil {
		return uuid.New().String()
	}

	var matchedUIDs []string
	for _, am := range archivedMembers {
		archivedTagaID := strings.TrimSpace(utils.GetStringFromMap(am, "tagaId"))
		archivedEmail := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "emailId")))
		if archivedEmail == "" {
			archivedEmail = strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "email")))
		}
		archivedUsername := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "username")))

		matchesTagaID := strings.EqualFold(archivedTagaID, cleanTagaID)
		matchesEmail := (archivedEmail != "" && archivedEmail == cleanEmail) ||
			(archivedUsername != "" && archivedUsername == cleanEmail)

		if matchesTagaID && matchesEmail {
			if idStr := utils.GetStringFromMap(am, "id"); idStr != "" {
				matchedUIDs = append(matchedUIDs, idStr)
			}
		}
	}

	if len(matchedUIDs) == 0 {
		return uuid.New().String()
	}

	if len(matchedUIDs) > 1 {
		if config.Logger != nil {
			config.Logger.Warn("Multiple archived records matched during UID resolution; generating fresh UUID",
				zap.String("tagaId", cleanTagaID),
				zap.String("email", cleanEmail),
				zap.Int("matches", len(matchedUIDs)),
			)
		}
		return uuid.New().String()
	}

	candidateUID := matchedUIDs[0]

	// Verify that the reused UID does not conflict with any active member
	for _, m := range activeMembers {
		if utils.GetStringFromMap(m, "id") == candidateUID {
			if config.Logger != nil {
				config.Logger.Warn("Candidate reused UID conflicts with existing active member; generating fresh UUID",
					zap.String("candidateUID", candidateUID),
					zap.String("tagaId", cleanTagaID),
				)
			}
			return uuid.New().String()
		}
	}

	if config.Logger != nil {
		config.Logger.Info("Reusing archived member UID for re-registered member",
			zap.String("reusedUID", candidateUID),
			zap.String("tagaId", cleanTagaID),
			zap.String("email", cleanEmail),
		)
	}

	return candidateUID
}

// UnarchiveMember removes a member from deleted_member.json when they are re-registered.
func UnarchiveMember(tagaID, email string) error {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	cleanTagaID := strings.TrimSpace(tagaID)
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanTagaID == "" && cleanEmail == "" {
		return nil
	}

	deletedFile := getDeletedMembersFilePath()
	data, err := os.ReadFile(deletedFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var archivedMembers []map[string]interface{}
	if err := json.Unmarshal(data, &archivedMembers); err != nil {
		return err
	}

	filtered := make([]map[string]interface{}, 0, len(archivedMembers))
	removed := false
	for _, am := range archivedMembers {
		archivedTaga := strings.TrimSpace(utils.GetStringFromMap(am, "tagaId"))
		archivedEmail := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "emailId")))
		if archivedEmail == "" {
			archivedEmail = strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "email")))
		}
		if archivedEmail == "" {
			archivedEmail = strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "username")))
		}

		matchesTaga := cleanTagaID != "" && strings.EqualFold(archivedTaga, cleanTagaID)
		matchesEmail := cleanEmail != "" && archivedEmail == cleanEmail

		if matchesTaga || matchesEmail {
			removed = true
			continue
		}
		filtered = append(filtered, am)
	}

	if removed {
		updatedData, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return err
		}
		return utils.WriteFileAtomic(deletedFile, updatedData, 0644)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Deletion Requests Storage and Management
// -----------------------------------------------------------------------------

func getDeletionRequestsFilePath() string {
	cfg := config.GetConfig()
	if cfg.DeletionRequestsFile != "" {
		return cfg.DeletionRequestsFile
	}
	return filepath.Join("data", "member", "deletion_requests.json")
}

func readAllDeletionRequestsUnsafe() ([]model.DeletionRequest, error) {
	filePath := getDeletionRequestsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.DeletionRequest{}, nil
		}
		return nil, err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return []model.DeletionRequest{}, nil
	}

	var reqs []model.DeletionRequest
	if err := json.Unmarshal(data, &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}

func saveAllDeletionRequestsUnsafe(reqs []model.DeletionRequest) error {
	filePath := getDeletionRequestsFilePath()
	data, err := json.MarshalIndent(reqs, "", "  ")
	if err != nil {
		return err
	}
	return utils.WriteFileAtomic(filePath, data, 0644)
}

func getMembersFilePath() string {
	cfg := config.GetConfig()
	if cfg.MembersFile != "" {
		return cfg.MembersFile
	}
	return "data/member/members.json"
}

func getDeletedMembersFilePath() string {
	cfg := config.GetConfig()
	if cfg.DeletedMembersFile != "" {
		return cfg.DeletedMembersFile
	}
	return "data/member/deleted_member.json"
}

func getMemberMobileFromFiles(memberID, tagaID, email string) string {
	cleanID := strings.TrimSpace(memberID)
	cleanTaga := strings.TrimSpace(tagaID)
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	// 1. Try active members.json
	membersFile := getMembersFilePath()
	if data, err := os.ReadFile(membersFile); err == nil && len(data) > 0 {
		var members []map[string]interface{}
		if err := json.Unmarshal(data, &members); err == nil {
			for _, m := range members {
				id := utils.GetStringFromMap(m, "id")
				t := utils.GetStringFromMap(m, "tagaId")
				em := strings.ToLower(utils.GetStringFromMap(m, "emailId"))
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "email"))
				}
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "username"))
				}

				if (cleanID != "" && id == cleanID) ||
					(cleanTaga != "" && strings.EqualFold(t, cleanTaga)) ||
					(cleanEmail != "" && em == cleanEmail) {
					mob := utils.GetStringFromMap(m, "mobile_number")
					if mob == "" {
						mob = utils.GetStringFromMap(m, "mobileNumber")
					}
					if mob == "" {
						mob = utils.GetStringFromMap(m, "phone")
					}
					if mob == "" {
						mob = utils.GetStringFromMap(m, "mobile")
					}
					if mob != "" {
						return mob
					}
				}
			}
		}
	}

	// 2. Try deleted_member.json
	deletedFile := getDeletedMembersFilePath()
	if data, err := os.ReadFile(deletedFile); err == nil && len(data) > 0 {
		var archived []map[string]interface{}
		if err := json.Unmarshal(data, &archived); err == nil {
			for _, m := range archived {
				id := utils.GetStringFromMap(m, "id")
				t := utils.GetStringFromMap(m, "tagaId")
				em := strings.ToLower(utils.GetStringFromMap(m, "emailId"))
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "email"))
				}
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "username"))
				}

				if (cleanID != "" && id == cleanID) ||
					(cleanTaga != "" && strings.EqualFold(t, cleanTaga)) ||
					(cleanEmail != "" && em == cleanEmail) {
					mob := utils.GetStringFromMap(m, "mobile_number")
					if mob == "" {
						mob = utils.GetStringFromMap(m, "mobileNumber")
					}
					if mob == "" {
						mob = utils.GetStringFromMap(m, "phone")
					}
					if mob == "" {
						mob = utils.GetStringFromMap(m, "mobile")
					}
					if mob != "" {
						return mob
					}
				}
			}
		}
	}

	return ""
}

func checkMemberSubscriptionStatus(memberID, tagaID, email string) (bool, string) {
	cleanID := strings.TrimSpace(memberID)
	cleanTaga := strings.TrimSpace(tagaID)
	cleanEmail := strings.ToLower(strings.TrimSpace(email))

	// 1. Check member_subscriptions.json for active subscription
	subsPath := getMemberSubscriptionsFilePath()
	if sData, err := os.ReadFile(subsPath); err == nil && len(sData) > 0 {
		var subs []map[string]interface{}
		if err := json.Unmarshal(sData, &subs); err == nil {
			for _, s := range subs {
				sMemberID := utils.GetStringFromMap(s, "member_id")
				sEmail := strings.ToLower(utils.GetStringFromMap(s, "member_email"))
				sStatus := strings.ToLower(utils.GetStringFromMap(s, "status"))

				matches := (cleanID != "" && sMemberID == cleanID) ||
					(cleanTaga != "" && strings.EqualFold(sMemberID, cleanTaga)) ||
					(cleanEmail != "" && sEmail == cleanEmail)

				if matches && sStatus == "active" {
					return true, "Paid"
				}
			}
		}
	}

	// 2. Check members.json
	membersFile := getMembersFilePath()
	if data, err := os.ReadFile(membersFile); err == nil && len(data) > 0 {
		var members []map[string]interface{}
		if err := json.Unmarshal(data, &members); err == nil {
			for _, m := range members {
				id := utils.GetStringFromMap(m, "id")
				t := utils.GetStringFromMap(m, "tagaId")
				em := strings.ToLower(utils.GetStringFromMap(m, "emailId"))
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "email"))
				}
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "username"))
				}

				matches := (cleanID != "" && id == cleanID) ||
					(cleanTaga != "" && strings.EqualFold(t, cleanTaga)) ||
					(cleanEmail != "" && em == cleanEmail)

				if matches {
					pStatus := utils.GetStringFromMap(m, "payment_status")
					mStatus := utils.GetStringFromMap(m, "membership_status")
					subActive := false
					if val, ok := m["subscription_active"].(bool); ok {
						subActive = val
					}

					if strings.EqualFold(pStatus, "Paid") || strings.EqualFold(pStatus, "active") ||
						strings.EqualFold(mStatus, "active") || subActive {
						return true, "Paid"
					}

					if pStatus != "" {
						return false, pStatus
					}
					return false, "Unpaid"
				}
			}
		}
	}

	// 3. Fallback to deleted_member.json
	deletedFile := getDeletedMembersFilePath()
	if data, err := os.ReadFile(deletedFile); err == nil && len(data) > 0 {
		var archived []map[string]interface{}
		if err := json.Unmarshal(data, &archived); err == nil {
			for _, m := range archived {
				id := utils.GetStringFromMap(m, "id")
				t := utils.GetStringFromMap(m, "tagaId")
				em := strings.ToLower(utils.GetStringFromMap(m, "emailId"))
				if em == "" {
					em = strings.ToLower(utils.GetStringFromMap(m, "email"))
				}

				matches := (cleanID != "" && id == cleanID) ||
					(cleanTaga != "" && strings.EqualFold(t, cleanTaga)) ||
					(cleanEmail != "" && em == cleanEmail)

				if matches {
					pStatus := utils.GetStringFromMap(m, "payment_status")
					if strings.EqualFold(pStatus, "Paid") {
						return true, "Paid"
					}
					if pStatus != "" {
						return false, pStatus
					}
				}
			}
		}
	}

	return false, "Unpaid"
}

func enrichDeletionRequest(r *model.DeletionRequest) {
	if r.TagaIDSnake == "" {
		r.TagaIDSnake = r.TagaID
	}

	// 1. Mobile number enrichment
	if strings.TrimSpace(r.MobileNumber) == "" {
		r.MobileNumber = getMemberMobileFromFiles(r.MemberID, r.TagaID, r.Email)
	}
	r.Mobile = r.MobileNumber

	// 2. Subscription & payment status enrichment
	isPaid, status := checkMemberSubscriptionStatus(r.MemberID, r.TagaID, r.Email)
	r.IsPaid = isPaid
	r.SubscriptionActive = isPaid
	if r.SubscriptionStatus == "" || r.SubscriptionStatus == "unknown" {
		r.SubscriptionStatus = status
	}
	if r.PaymentStatus == "" || r.PaymentStatus == "unknown" {
		r.PaymentStatus = status
	}
	if isPaid {
		r.SubscriptionStatus = "Paid"
		r.PaymentStatus = "Paid"
	}
}

func enrichDeletionRequests(reqs []model.DeletionRequest) []model.DeletionRequest {
	enriched := make([]model.DeletionRequest, len(reqs))
	for i := range reqs {
		enriched[i] = reqs[i]
		enrichDeletionRequest(&enriched[i])
	}
	return enriched
}

// ListDeletionRequests returns all deletion requests enriched with mobile number and subscription status.
func ListDeletionRequests(statusFilter string) ([]model.DeletionRequest, error) {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return nil, err
	}

	enriched := enrichDeletionRequests(reqs)

	if statusFilter == "" {
		return enriched, nil
	}

	filtered := make([]model.DeletionRequest, 0, len(enriched))
	for _, r := range enriched {
		if strings.EqualFold(string(r.Status), statusFilter) {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// GetDeletionRequestByID retrieves a deletion request by ID with enrichment.
func GetDeletionRequestByID(requestID string) (*model.DeletionRequest, error) {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return nil, err
	}

	for _, r := range reqs {
		if r.ID == requestID {
			enrichDeletionRequest(&r)
			return &r, nil
		}
	}
	return nil, fmt.Errorf("deletion request not found")
}

// GetPendingDeletionRequest returns whether a member has an active pending deletion request.
func GetPendingDeletionRequest(memberID, tagaID, email string) (bool, *model.DeletionRequest, error) {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return false, nil, err
	}

	for _, r := range reqs {
		if r.Status != model.DeletionStatusPending {
			continue
		}
		matches := (memberID != "" && r.MemberID == memberID) ||
			(tagaID != "" && strings.EqualFold(r.TagaID, tagaID)) ||
			(email != "" && strings.EqualFold(r.Email, email))

		if matches {
			enrichDeletionRequest(&r)
			return true, &r, nil
		}
	}

	return false, nil, nil
}

// GetLatestDeletionRequest returns the most recent deletion request for a member.
func GetLatestDeletionRequest(memberID, tagaID, email string) (*model.DeletionRequest, error) {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return nil, err
	}

	var latest *model.DeletionRequest
	for i := len(reqs) - 1; i >= 0; i-- {
		r := reqs[i]
		matches := (memberID != "" && r.MemberID == memberID) ||
			(tagaID != "" && strings.EqualFold(r.TagaID, tagaID)) ||
			(email != "" && strings.EqualFold(r.Email, email))

		if matches {
			enrichDeletionRequest(&r)
			latest = &r
			break
		}
	}

	return latest, nil
}

// CreateDeletionRequest creates a new pending deletion request, stores mobile number, and sends notification emails.
func CreateDeletionRequest(c *gin.Context, memberID, tagaID, memberName, memberEmail, reason string, mobileNumberOpt ...string) (*model.DeletionRequest, error) {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	// Verify active stay safety check
	hasActiveStay, err := HasActiveStayAtTagaTowers(memberID, tagaID)
	if err != nil {
		return nil, err
	}
	if hasActiveStay {
		return nil, ErrActiveStay
	}

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return nil, err
	}

	// Check if already pending
	for _, r := range reqs {
		if r.Status == model.DeletionStatusPending {
			if (memberID != "" && r.MemberID == memberID) ||
				(tagaID != "" && strings.EqualFold(r.TagaID, tagaID)) ||
				(memberEmail != "" && strings.EqualFold(r.Email, memberEmail)) {
				return nil, fmt.Errorf("account deletion request is already pending")
			}
		}
	}

	// Resolve mobile number
	mobileNumber := ""
	if len(mobileNumberOpt) > 0 {
		mobileNumber = strings.TrimSpace(mobileNumberOpt[0])
	}
	if mobileNumber == "" {
		mobileNumber = getMemberMobileFromFiles(memberID, tagaID, memberEmail)
	}

	// Resolve subscription status
	isPaid, paymentStatus := checkMemberSubscriptionStatus(memberID, tagaID, memberEmail)

	newReq := model.DeletionRequest{
		ID:                 uuid.New().String(),
		MemberID:           memberID,
		TagaID:             tagaID,
		TagaIDSnake:        tagaID,
		MemberName:         memberName,
		Email:              memberEmail,
		MobileNumber:       mobileNumber,
		Mobile:             mobileNumber,
		Status:             model.DeletionStatusPending,
		Reason:             strings.TrimSpace(reason),
		RequestedAt:        time.Now().Format(time.RFC3339),
		SubscriptionStatus: paymentStatus,
		PaymentStatus:      paymentStatus,
		IsPaid:             isPaid,
		SubscriptionActive: isPaid,
	}

	reqs = append(reqs, newReq)
	if err := saveAllDeletionRequestsUnsafe(reqs); err != nil {
		return nil, fmt.Errorf("failed to save deletion request: %w", err)
	}

	// Mark member as deletion pending in members.json
	_ = setMemberDeletionStatusInFile(memberID, tagaID, "pending")

	// Audit log
	if c != nil {
		_ = audit.Log(c, tagaID, memberEmail,
			audit.ActionCreate, audit.ModuleMember,
			"deletion_request", newReq.ID,
			fmt.Sprintf("Member %s (tagaId: %s) requested account deletion", memberName, tagaID),
			nil, newReq)
	}

	// 1. Dispatch acknowledgement email to member
	if memberEmail != "" {
		go func(to, name, taga, rsn, reqAt, mob string) {
			if err := email.SendAccountDeletionRequestAcknowledgement(to, name, taga, rsn, reqAt, mob); err != nil {
				if config.Logger != nil {
					config.Logger.Warn("Failed to send deletion request acknowledgement email to member",
						zap.String("to", to), zap.Error(err))
				}
			}
		}(memberEmail, memberName, tagaID, newReq.Reason, newReq.RequestedAt, mobileNumber)
	}

	// 2. Dispatch alert notification email to admin
	cfg := config.GetConfig()
	adminEmail := cfg.AdminEmail
	if adminEmail == "" {
		adminEmail = "nammataga@gmail.com"
	}
	if adminEmail != "" {
		go func(admEmail, name, taga, memEmail, mob, rsn, reqAt string) {
			if err := email.SendAccountDeletionRequestAdminAlert(admEmail, name, taga, memEmail, mob, rsn, reqAt); err != nil {
				if config.Logger != nil {
					config.Logger.Warn("Failed to send deletion request alert email to admin",
						zap.String("admin_email", admEmail), zap.Error(err))
				}
			}
		}(adminEmail, memberName, tagaID, memberEmail, mobileNumber, newReq.Reason, newReq.RequestedAt)
	}

	return &newReq, nil
}

// RejectDeletionRequest rejects a deletion request, unlocks member account, and sends an email.
func RejectDeletionRequest(c *gin.Context, requestID, rejectedBy, remarks string) (*model.DeletionRequest, error) {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	trimmedRemarks := strings.TrimSpace(remarks)
	if trimmedRemarks == "" {
		return nil, fmt.Errorf("remarks are required for rejecting deletion request")
	}

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return nil, err
	}

	var target *model.DeletionRequest
	for i := range reqs {
		if reqs[i].ID == requestID {
			target = &reqs[i]
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("deletion request not found")
	}

	if target.Status != model.DeletionStatusPending {
		return nil, fmt.Errorf("deletion request is already processed")
	}

	target.Status = model.DeletionStatusRejected
	target.Remarks = trimmedRemarks
	target.ProcessedAt = time.Now().Format(time.RFC3339)
	target.ProcessedBy = rejectedBy

	if err := saveAllDeletionRequestsUnsafe(reqs); err != nil {
		return nil, fmt.Errorf("failed to save updated deletion request: %w", err)
	}

	// Unlock member account by clearing deletion status in members.json
	_ = setMemberDeletionStatusInFile(target.MemberID, target.TagaID, "")

	// Audit log
	if c != nil {
		_ = audit.Log(c, "admin", rejectedBy,
			audit.ActionReject, audit.ModuleMember,
			"deletion_request", target.ID,
			fmt.Sprintf("Admin %s rejected deletion request for %s (remarks: %s)", rejectedBy, target.MemberName, trimmedRemarks),
			nil, target)
	}

	// Send notification email to member
	go func(to, name, taga, rem string) {
		if err := email.SendAccountDeletionRejection(to, name, taga, rem); err != nil {
			if config.Logger != nil {
				config.Logger.Warn("Failed to send deletion rejection email",
					zap.String("to", to), zap.Error(err))
			}
		}
	}(target.Email, target.MemberName, target.TagaID, trimmedRemarks)

	return target, nil
}

func updateDeletionRequestStatus(requestID, memberID, tagaID string, status model.DeletionRequestStatus, processedBy, remarks string) error {
	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return err
	}

	found := false
	nowStr := time.Now().Format(time.RFC3339)

	for i := range reqs {
		isTarget := false
		if requestID != "" && reqs[i].ID == requestID {
			isTarget = true
		} else if requestID == "" && reqs[i].Status == model.DeletionStatusPending {
			if (memberID != "" && reqs[i].MemberID == memberID) ||
				(tagaID != "" && strings.EqualFold(reqs[i].TagaID, tagaID)) {
				isTarget = true
			}
		}

		if isTarget {
			reqs[i].Status = status
			reqs[i].ProcessedAt = nowStr
			reqs[i].ProcessedBy = processedBy
			if remarks != "" {
				reqs[i].Remarks = remarks
			}
			found = true
			break
		}
	}

	if found {
		return saveAllDeletionRequestsUnsafe(reqs)
	}
	return nil
}

func setMemberDeletionStatusInFile(memberID, tagaID, status string) error {
	cfg := config.GetConfig()
	membersFile := cfg.MembersFile
	if membersFile == "" {
		membersFile = "data/member/members.json"
	}

	data, err := os.ReadFile(membersFile)
	if err != nil {
		return err
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}

	modified := false
	for i := range members {
		id := utils.GetStringFromMap(members[i], "id")
		t := utils.GetStringFromMap(members[i], "tagaId")
		if (memberID != "" && id == memberID) || (tagaID != "" && strings.EqualFold(t, tagaID)) {
			if status == "" {
				delete(members[i], "deletion_status")
			} else {
				members[i]["deletion_status"] = status
			}
			modified = true
			break
		}
	}

	if modified {
		updated, err := json.MarshalIndent(members, "", "  ")
		if err != nil {
			return err
		}
		return utils.WriteFileAtomic(membersFile, updated, 0644)
	}
	return nil
}

// IsMemberDeletionPending checks if member currently has a pending account deletion request.
func IsMemberDeletionPending(memberID, email string) bool {
	deletionLock.Lock()
	defer deletionLock.Unlock()

	reqs, err := readAllDeletionRequestsUnsafe()
	if err != nil {
		return false
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanID := strings.TrimSpace(memberID)

	for _, r := range reqs {
		if r.Status != model.DeletionStatusPending {
			continue
		}
		rEmail := strings.ToLower(strings.TrimSpace(r.Email))
		rTagaID := strings.ToLower(strings.TrimSpace(r.TagaID))
		rMemberID := strings.TrimSpace(r.MemberID)

		if (cleanID != "" && rMemberID == cleanID) ||
			(cleanEmail != "" && rEmail == cleanEmail) ||
			(cleanEmail != "" && rTagaID == cleanEmail) ||
			(cleanID != "" && rTagaID == strings.ToLower(cleanID)) {
			return true
		}
	}
	return false
}

// IsAccountArchivedOrDeleted checks if the identifier exists in deleted_member.json
func IsAccountArchivedOrDeleted(identifier string) bool {
	clean := strings.ToLower(strings.TrimSpace(identifier))
	if clean == "" {
		return false
	}

	deletedFile := config.Config.DeletedMembersFile
	if deletedFile == "" {
		deletedFile = "data/member/deleted_member.json"
	}

	data, err := os.ReadFile(deletedFile)
	if err != nil {
		return false
	}

	var archivedMembers []map[string]interface{}
	if err := json.Unmarshal(data, &archivedMembers); err != nil {
		return false
	}

	for _, am := range archivedMembers {
		archivedEmail := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "emailId")))
		if archivedEmail == "" {
			archivedEmail = strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "email")))
		}
		archivedUsername := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "username")))
		archivedTagaID := strings.ToLower(strings.TrimSpace(utils.GetStringFromMap(am, "tagaId")))
		archivedPhone := strings.TrimSpace(utils.GetStringFromMap(am, "mobile_number"))

		if (archivedEmail != "" && archivedEmail == clean) ||
			(archivedUsername != "" && archivedUsername == clean) ||
			(archivedTagaID != "" && archivedTagaID == clean) ||
			(archivedPhone != "" && archivedPhone == clean) {
			return true
		}
	}
	return false
}

// IsMemberDeletedOrNotFound checks if member does not exist in members.json or exists in deleted_member.json.
func IsMemberDeletedOrNotFound(memberID, email string) bool {
	cfg := config.GetConfig()
	membersFile := cfg.MembersFile
	if membersFile == "" {
		membersFile = "data/member/members.json"
	}

	data, err := os.ReadFile(membersFile)
	if err != nil {
		return false // if cannot read members file, don't falsely claim deleted
	}

	var members []map[string]interface{}
	if err := json.Unmarshal(data, &members); err != nil {
		return false
	}

	foundInActive := false
	for _, m := range members {
		id := utils.GetStringFromMap(m, "id")
		mEmail := utils.GetStringFromMap(m, "emailId")
		if mEmail == "" {
			mEmail = utils.GetStringFromMap(m, "email")
		}
		if mEmail == "" {
			mEmail = utils.GetStringFromMap(m, "username")
		}

		if (memberID != "" && id == memberID) || (email != "" && strings.EqualFold(mEmail, email)) {
			foundInActive = true
			break
		}
	}

	if !foundInActive {
		return true // not in active members!
	}

	return false
}
