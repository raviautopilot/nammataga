package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taga-api/config"
	"taga-api/model"
)

func setupTestSubscriptionEnvironment(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "taga_subscription_service_test_*")
	assert.NoError(t, err)

	subsDir := filepath.Join(tempDir, "subscriptions")
	membersDir := filepath.Join(tempDir, "member")
	assert.NoError(t, os.MkdirAll(subsDir, 0755))
	assert.NoError(t, os.MkdirAll(membersDir, 0755))

	membersFile := filepath.Join(membersDir, "members.json")
	subsFile := filepath.Join(subsDir, "member_subscriptions.json")

	// Create initial empty subscriptions file
	assert.NoError(t, os.WriteFile(subsFile, []byte("[]"), 0644))

	oldMembersFile := config.Config.MembersFile
	config.Config.MembersFile = membersFile

	cleanup := func() {
		config.Config.MembersFile = oldMembersFile
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestConcurrentRenewalWrites_ZeroDataLoss(t *testing.T) {
	_, cleanup := setupTestSubscriptionEnvironment(t)
	defer cleanup()

	// Simulate 20 concurrent members paying renewal simultaneously on April 1st
	const numGoroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines)

	now := time.Now()
	nextYearEnd := time.Date(now.Year()+1, time.March, 31, 23, 59, 59, 0, time.UTC)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			memberID := fmt.Sprintf("TAGA_%04d", idx)
			email := fmt.Sprintf("member%d@nammataga.com", idx)
			name := fmt.Sprintf("Member %d", idx)

			sub := model.MemberSubscription{
				ID:               uuid.New().String(),
				MemberID:         memberID,
				MemberEmail:      email,
				MemberName:       name,
				SubscriptionID:   "annual-subscription",
				SubscriptionName: "Annual Subscription",
				Amount:           100000, // Rs. 1000 in paise
				OrderID:          fmt.Sprintf("order_%04d", idx),
				PaymentID:        fmt.Sprintf("pay_%04d", idx),
				Status:           "active",
				StartDate:        now,
				EndDate:          nextYearEnd,
				LastPaidDate:     now,
				NextDueDate:      nextYearEnd.AddDate(0, 0, 1),
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			// Perform renewal write
			if err := RenewAnnualSubscription(sub, email, memberID, memberID); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err, "No goroutine should encounter an error during renewal")
	}

	// Verify ZERO DATA LOSS: exactly 20 records must exist in member_subscriptions.json
	subs, err := ReadMemberSubscriptions()
	assert.NoError(t, err)
	assert.Equal(t, numGoroutines, len(subs), "Zero data loss: exactly 20 subscription records must exist")

	// Verify each member is represented accurately
	seenMembers := make(map[string]bool)
	for _, s := range subs {
		assert.Equal(t, "active", s.Status)
		assert.Equal(t, "annual-subscription", s.SubscriptionID)
		assert.False(t, seenMembers[s.MemberID], "Member ID should be unique in final active records: %s", s.MemberID)
		seenMembers[s.MemberID] = true
	}
	assert.Equal(t, numGoroutines, len(seenMembers))

	// Verify no stray .tmp files are left over in the subscriptions directory
	subsPath := GetMemberSubscriptionsFilePath()
	dir := filepath.Dir(subsPath)
	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, filepath.Ext(entry.Name()) == ".tmp",
			"Temporary file should not remain after atomic rename: %s", entry.Name())
	}
}

func TestConcurrentRenewalWithExistingSubscriptions_ExpiresOldAndAppendsNew(t *testing.T) {
	_, cleanup := setupTestSubscriptionEnvironment(t)
	defer cleanup()

	// Pre-populate with previous year's active subscriptions for 20 members
	const numMembers = 20
	pastDate := time.Now().AddDate(-1, 0, 0)
	initialSubs := make([]model.MemberSubscription, numMembers)
	for i := 0; i < numMembers; i++ {
		memberID := fmt.Sprintf("TAGA_%04d", i)
		email := fmt.Sprintf("member%d@nammataga.com", i)
		initialSubs[i] = model.MemberSubscription{
			ID:               uuid.New().String(),
			MemberID:         memberID,
			MemberEmail:      email,
			MemberName:       fmt.Sprintf("Member %d", i),
			SubscriptionID:   "annual-subscription",
			SubscriptionName: "Annual Subscription",
			Amount:           100000,
			OrderID:          fmt.Sprintf("old_order_%d", i),
			PaymentID:        fmt.Sprintf("old_pay_%d", i),
			Status:           "active",
			StartDate:        pastDate,
			EndDate:          pastDate.AddDate(1, 0, 0),
			LastPaidDate:     pastDate,
			CreatedAt:        pastDate,
			UpdatedAt:        pastDate,
		}
	}
	assert.NoError(t, WriteMemberSubscriptionsAtomic(initialSubs))

	// Now simulate all 20 members renewing concurrently!
	var wg sync.WaitGroup
	now := time.Now()
	nextYearEnd := time.Date(now.Year()+1, time.March, 31, 23, 59, 59, 0, time.UTC)

	for i := 0; i < numMembers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			memberID := fmt.Sprintf("TAGA_%04d", idx)
			email := fmt.Sprintf("member%d@nammataga.com", idx)
			name := fmt.Sprintf("Member %d", idx)

			newSub := model.MemberSubscription{
				ID:               uuid.New().String(),
				MemberID:         memberID,
				MemberEmail:      email,
				MemberName:       name,
				SubscriptionID:   "annual-subscription",
				SubscriptionName: "Annual Subscription",
				Amount:           100000,
				OrderID:          fmt.Sprintf("renew_order_%d", idx),
				PaymentID:        fmt.Sprintf("renew_pay_%d", idx),
				Status:           "active",
				StartDate:        now,
				EndDate:          nextYearEnd,
				LastPaidDate:     now,
				NextDueDate:      nextYearEnd.AddDate(0, 0, 1),
				CreatedAt:        now,
				UpdatedAt:        now,
			}

			err := RenewAnnualSubscription(newSub, email, memberID, memberID)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Total records should be 40: 20 expired + 20 active renewals (ZERO records lost!)
	allSubs, err := ReadMemberSubscriptions()
	assert.NoError(t, err)
	assert.Equal(t, numMembers*2, len(allSubs), "All old and new records must be preserved")

	activeCount := 0
	expiredCount := 0
	for _, s := range allSubs {
		if s.Status == "active" {
			activeCount++
		} else if s.Status == "expired" {
			expiredCount++
		}
	}

	assert.Equal(t, numMembers, activeCount, "Exactly 20 records should be active")
	assert.Equal(t, numMembers, expiredCount, "Exactly 20 previous records should be expired")
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	_, cleanup := setupTestSubscriptionEnvironment(t)
	defer cleanup()

	const writers = 15
	const readers = 15
	var wg sync.WaitGroup

	now := time.Now()
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sub := model.MemberSubscription{
				ID:             uuid.New().String(),
				MemberID:       fmt.Sprintf("TAGA_W_%d", idx),
				MemberEmail:    fmt.Sprintf("writer%d@test.com", idx),
				SubscriptionID: "annual-subscription",
				Status:         "active",
				StartDate:      now,
				EndDate:        now.AddDate(1, 0, 0),
			}
			_ = SaveMemberSubscription(sub)
		}(i)
	}

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Continuous reads should never encounter unmarshal error or panic
			for r := 0; r < 5; r++ {
				_, _ = ReadMemberSubscriptions()
				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	subs, err := ReadMemberSubscriptions()
	assert.NoError(t, err)
	assert.Equal(t, writers, len(subs), "All writer records must be safely written")
}
