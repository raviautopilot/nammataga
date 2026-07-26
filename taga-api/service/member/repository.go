package member

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"taga-api/config"
	"taga-api/model"
)

var (
	members     []model.Member
	membersLock sync.RWMutex
)

// InitMemberRepository loads members into memory on startup
func InitMemberRepository() error {
	membersLock.Lock()
	defer membersLock.Unlock()

	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		if os.IsNotExist(err) {
			members = []model.Member{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &members)
}

// SaveMember saves a member to the file system
func SaveMember(member model.Member) error {
	membersLock.Lock()
	defer membersLock.Unlock()

	cfg := config.GetConfig()

	// Read current members from file
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var allMembers []model.Member
	if len(data) > 0 {
		if err := json.Unmarshal(data, &allMembers); err != nil {
			return err
		}
	}

	// Check if member exists and update
	found := false
	for i, existing := range allMembers {
		if existing.EmailId == member.EmailId {
			allMembers[i] = member
			found = true
			break
		}
	}

	// Add new member with default payment fields
	if !found {
		if member.PaymentStatus == "" {
			member.PaymentStatus = "Unpaid"
		}
		allMembers = append(allMembers, member)
	}

	// Write to file
	updatedData, err := json.MarshalIndent(allMembers, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.MembersFile, updatedData, 0644)
}



// UpdateMemberPaymentStatus updates only payment fields for a member
func UpdateMemberPaymentStatus(email string, isPaid bool) error {
	membersLock.Lock()
	defer membersLock.Unlock()

	cfg := config.GetConfig()

	// Read current members from file
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		return fmt.Errorf("failed to read members file: %w", err)
	}

	var allMembers []model.Member
	if err := json.Unmarshal(data, &allMembers); err != nil {
		return fmt.Errorf("failed to parse members: %w", err)
	}

	// Find and update only the specific member
	found := false
	for i, m := range allMembers {
		if m.EmailId == email {
			if isPaid {
				allMembers[i].PaymentStatus = "Paid"
				allMembers[i].SubscriptionActive = true
			} else {
				allMembers[i].PaymentStatus = "Unpaid"
				allMembers[i].SubscriptionActive = false
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("member not found: %s", email)
	}

	// Write back only the updated data
	updatedData, err := json.MarshalIndent(allMembers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}

	return os.WriteFile(cfg.MembersFile, updatedData, 0644)
}

// GetAllMembers returns all members
func GetAllMembers() ([]model.Member, error) {
	cfg := config.GetConfig()
	data, err := os.ReadFile(cfg.MembersFile)
	if err != nil {
		return nil, err
	}

	var allMembers []model.Member
	if err := json.Unmarshal(data, &allMembers); err != nil {
		return nil, err
	}

	return allMembers, nil
}
