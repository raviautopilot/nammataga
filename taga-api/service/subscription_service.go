package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"taga-api/config"
	"taga-api/model"
	"taga-api/utils"
)

// SubscriptionsLock protects all concurrent reads and writes to member_subscriptions.json.
var SubscriptionsLock sync.RWMutex

// GetMemberSubscriptionsFilePath returns the file path for member_subscriptions.json.
func GetMemberSubscriptionsFilePath() string {
	cfg := config.GetConfig()
	if cfg.MembersFile != "" {
		subsFile := filepath.Join(filepath.Dir(cfg.MembersFile), "..", "subscriptions", "member_subscriptions.json")
		if _, err := os.Stat(subsFile); err == nil {
			return subsFile
		}
	}
	return filepath.Join("data", "subscriptions", "member_subscriptions.json")
}

// ReadMemberSubscriptions reads all subscriptions safely under a read lock.
func ReadMemberSubscriptions() ([]model.MemberSubscription, error) {
	SubscriptionsLock.RLock()
	defer SubscriptionsLock.RUnlock()

	return ReadMemberSubscriptionsUnsafe()
}

// ReadMemberSubscriptionsUnsafe reads all subscriptions without acquiring lock (internal use).
func ReadMemberSubscriptionsUnsafe() ([]model.MemberSubscription, error) {
	filePath := GetMemberSubscriptionsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.MemberSubscription{}, nil
		}
		return nil, err
	}

	var subscriptions []model.MemberSubscription
	if err := json.Unmarshal(data, &subscriptions); err != nil {
		return nil, err
	}
	return subscriptions, nil
}

// WriteMemberSubscriptionsAtomic writes all subscriptions safely under a write lock using atomic rename.
func WriteMemberSubscriptionsAtomic(subscriptions []model.MemberSubscription) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	return WriteMemberSubscriptionsAtomicUnsafe(subscriptions)
}

// WriteMemberSubscriptionsAtomicUnsafe writes subscriptions atomically without acquiring lock.
func WriteMemberSubscriptionsAtomicUnsafe(subscriptions []model.MemberSubscription) error {
	filePath := GetMemberSubscriptionsFilePath()
	updatedData, err := json.MarshalIndent(subscriptions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %w", err)
	}

	return utils.WriteFileAtomic(filePath, updatedData, 0644)
}

// SaveMemberSubscription appends a new subscription under write lock and writes atomically.
func SaveMemberSubscription(subscription model.MemberSubscription) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		subscriptions = []model.MemberSubscription{}
	}

	subscriptions = append(subscriptions, subscription)
	return WriteMemberSubscriptionsAtomicUnsafe(subscriptions)
}

// RenewAnnualSubscription expires any active annual subscriptions for the member and appends
// the new subscription, performing the entire renewal in a single atomic transaction under the write lock.
func RenewAnnualSubscription(subscription model.MemberSubscription, email, tagaID, memberID string) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		subscriptions = []model.MemberSubscription{}
	}

	now := time.Now()
	// Expire existing active annual subscriptions for this member
	for i, sub := range subscriptions {
		matches := false
		if email != "" && strings.EqualFold(sub.MemberEmail, email) {
			matches = true
		} else if tagaID != "" && strings.EqualFold(sub.MemberID, tagaID) {
			matches = true
		} else if memberID != "" && sub.MemberID == memberID {
			matches = true
		}

		if matches && sub.SubscriptionID == "annual-subscription" && sub.Status == "active" {
			subscriptions[i].Status = "expired"
			subscriptions[i].UpdatedAt = now
		}
	}

	// Append the new renewal subscription
	subscriptions = append(subscriptions, subscription)
	return WriteMemberSubscriptionsAtomicUnsafe(subscriptions)
}

// ExpireExistingAnnualSubscriptions sets active annual subscriptions for email/tagaID/memberID to expired.
func ExpireExistingAnnualSubscriptions(email, tagaID, memberID string) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		return err
	}

	modified := false
	now := time.Now()
	for i, sub := range subscriptions {
		matches := false
		if email != "" && strings.EqualFold(sub.MemberEmail, email) {
			matches = true
		} else if tagaID != "" && strings.EqualFold(sub.MemberID, tagaID) {
			matches = true
		} else if memberID != "" && sub.MemberID == memberID {
			matches = true
		}

		if matches && sub.SubscriptionID == "annual-subscription" && sub.Status == "active" {
			subscriptions[i].Status = "expired"
			subscriptions[i].UpdatedAt = now
			modified = true
		}
	}

	if modified {
		return WriteMemberSubscriptionsAtomicUnsafe(subscriptions)
	}
	return nil
}

// GetActiveMemberSubscription retrieves active subscription safely under read lock.
func GetActiveMemberSubscription(email, tagaID, memberID string) (*model.MemberSubscription, error) {
	SubscriptionsLock.RLock()
	defer SubscriptionsLock.RUnlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		return nil, err
	}

	for _, s := range subscriptions {
		matches := false
		if email != "" && strings.EqualFold(s.MemberEmail, email) {
			matches = true
		} else if tagaID != "" && strings.EqualFold(s.MemberID, tagaID) {
			matches = true
		} else if memberID != "" && s.MemberID == memberID {
			matches = true
		}

		if matches && s.Status == "active" {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("no active subscription found")
}

// UpdateSubscriptionStatus updates status of a specific subscription ID safely under write lock.
func UpdateSubscriptionStatus(subscriptionID, status string) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		return err
	}

	modified := false
	for i, s := range subscriptions {
		if s.ID == subscriptionID {
			subscriptions[i].Status = status
			subscriptions[i].UpdatedAt = time.Now()
			modified = true
			break
		}
	}

	if modified {
		return WriteMemberSubscriptionsAtomicUnsafe(subscriptions)
	}
	return nil
}

// HasMemberPaidOneTime checks whether member has paid one-time fee safely under read lock.
func HasMemberPaidOneTime(subscriptionID, email string) bool {
	SubscriptionsLock.RLock()
	defer SubscriptionsLock.RUnlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		return false
	}

	for _, sub := range subscriptions {
		if strings.EqualFold(sub.MemberEmail, email) && sub.SubscriptionID == subscriptionID {
			return true
		}
	}
	return false
}

// CleanupMemberSubscriptions removes records for deleted member under write lock with atomic write.
func CleanupMemberSubscriptions(memberID, memberTagaID string) error {
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		return err
	}

	filtered := make([]model.MemberSubscription, 0, len(subscriptions))
	for _, s := range subscriptions {
		if s.MemberID != memberID && s.MemberID != memberTagaID {
			filtered = append(filtered, s)
		}
	}

	return WriteMemberSubscriptionsAtomicUnsafe(filtered)
}

// SyncMemberEmailInSubscriptions updates member_email under write lock with atomic write.
func SyncMemberEmailInSubscriptions(memberUUID, tagaID, oldEmail, newEmail string) error {
	if newEmail == "" {
		return nil
	}
	SubscriptionsLock.Lock()
	defer SubscriptionsLock.Unlock()

	subscriptions, err := ReadMemberSubscriptionsUnsafe()
	if err != nil {
		return err
	}

	modified := false
	now := time.Now()
	for i, sub := range subscriptions {
		matches := false
		if memberUUID != "" && sub.MemberID == memberUUID {
			matches = true
		} else if tagaID != "" && strings.EqualFold(sub.MemberID, tagaID) {
			matches = true
		} else if oldEmail != "" && strings.EqualFold(sub.MemberEmail, oldEmail) {
			matches = true
		}

		if matches {
			subscriptions[i].MemberEmail = newEmail
			subscriptions[i].UpdatedAt = now
			modified = true
		}
	}

	if modified {
		return WriteMemberSubscriptionsAtomicUnsafe(subscriptions)
	}
	return nil
}
