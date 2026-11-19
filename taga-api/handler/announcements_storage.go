package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
	"taga-api/config"
	"taga-api/model"
)

const AnnouncementExpiryDuration = 14 * 24 * time.Hour

var announcementMutex sync.Mutex

func getAnnouncementsFilePath() string {
	dir := filepath.Join("data", "announcements")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "announcements.json")
}

func getArchivedAnnouncementsFilePath() string {
	dir := filepath.Join("data", "announcements")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "archived_announcements.json")
}

func getEffectivePublishedAt(n model.Notification) time.Time {
	if !n.PublishedAt.IsZero() {
		return n.PublishedAt
	}
	if !n.SentAt.IsZero() {
		return n.SentAt
	}
	if !n.CreatedAt.IsZero() {
		return n.CreatedAt
	}
	return time.Time{}
}

func atomicWriteJSON(filePath string, v interface{}) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	tmpFile := fmt.Sprintf("%s.%d.tmp", filePath, time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tmp file %s: %w", tmpFile, err)
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		// On Windows or if rename fails because dest exists, remove dest first
		_ = os.Remove(filePath)
		if err := os.Rename(tmpFile, filePath); err != nil {
			_ = os.Remove(tmpFile)
			return fmt.Errorf("failed to rename %s to %s: %w", tmpFile, filePath, err)
		}
	}
	return nil
}

// processAnnouncementExpiryLocked evaluates active announcements against the 14-day expiry window,
// moves expired items to archived_announcements.json without duplicates, and persists active announcements.
// Caller MUST hold announcementMutex.
func processAnnouncementExpiryLocked() ([]model.Notification, error) {
	filePath := getAnnouncementsFilePath()
	archivePath := getArchivedAnnouncementsFilePath()

	var announcements []model.Notification
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Notification{}, nil
		}
		return nil, fmt.Errorf("failed to read announcements file: %w", err)
	}

	if len(bytes.TrimSpace(data)) > 0 {
		if err := json.Unmarshal(data, &announcements); err != nil {
			return nil, fmt.Errorf("failed to unmarshal announcements: %w", err)
		}
	}

	now := time.Now().UTC()
	var active []model.Notification
	var expired []model.Notification

	for _, item := range announcements {
		eff := getEffectivePublishedAt(item)
		item.PublishedAt = eff
		if item.SentAt.IsZero() {
			item.SentAt = eff
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = eff
		}
		item.IsRead = false

		// Calculate expiry: published_at + 14*24*time.Hour
		if !eff.IsZero() && now.After(eff.Add(AnnouncementExpiryDuration)) {
			expired = append(expired, item)
		} else {
			active = append(active, item)
		}
	}

	if len(expired) > 0 {
		var archived []model.Notification
		archData, aErr := os.ReadFile(archivePath)
		if aErr == nil && len(bytes.TrimSpace(archData)) > 0 {
			_ = json.Unmarshal(archData, &archived)
		}

		archivedMap := make(map[string]bool)
		for _, a := range archived {
			archivedMap[a.ID] = true
		}

		var newlyArchivedIDs []string
		for _, exp := range expired {
			if !archivedMap[exp.ID] {
				archived = append(archived, exp)
				archivedMap[exp.ID] = true
				newlyArchivedIDs = append(newlyArchivedIDs, exp.ID)
			}
		}

		// Save updated archive file atomically
		if err := atomicWriteJSON(archivePath, archived); err != nil {
			if config.Logger != nil {
				config.Logger.Error("Failed to write archived announcements", zap.Error(err))
			}
		}

		// Save updated active announcements file atomically
		if err := atomicWriteJSON(filePath, active); err != nil {
			if config.Logger != nil {
				config.Logger.Error("Failed to update active announcements", zap.Error(err))
			}
			return nil, err
		}

		if config.Logger != nil {
			config.Logger.Info("Archived expired announcements",
				zap.Int("expired_count", len(expired)),
				zap.Int("newly_archived_count", len(newlyArchivedIDs)),
				zap.Strings("archived_ids", newlyArchivedIDs),
			)
		}
	}

	// Sort active announcements newest published_at first
	sort.Slice(active, func(i, j int) bool {
		return active[i].PublishedAt.After(active[j].PublishedAt)
	})

	if active == nil {
		active = []model.Notification{}
	}

	return active, nil
}

// GetActiveAnnouncements returns active announcements after running the 14-day expiry check.
func GetActiveAnnouncements() ([]model.Notification, error) {
	announcementMutex.Lock()
	defer announcementMutex.Unlock()

	return processAnnouncementExpiryLocked()
}

// GetArchivedAnnouncements returns all archived announcements.
func GetArchivedAnnouncements() ([]model.Notification, error) {
	announcementMutex.Lock()
	defer announcementMutex.Unlock()

	archivePath := getArchivedAnnouncementsFilePath()
	var archived []model.Notification
	archData, err := os.ReadFile(archivePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Notification{}, nil
		}
		return nil, err
	}
	if len(bytes.TrimSpace(archData)) > 0 {
		if err := json.Unmarshal(archData, &archived); err != nil {
			return nil, err
		}
	}
	if archived == nil {
		archived = []model.Notification{}
	}
	return archived, nil
}

// SaveAnnouncement saves a new announcement, runs expiry, and updates announcements.json.
func SaveAnnouncement(announcement model.Notification) error {
	announcementMutex.Lock()
	defer announcementMutex.Unlock()

	active, err := processAnnouncementExpiryLocked()
	if err != nil {
		active = []model.Notification{}
	}

	now := time.Now().UTC()
	if announcement.PublishedAt.IsZero() {
		announcement.PublishedAt = now
	}
	if announcement.SentAt.IsZero() {
		announcement.SentAt = announcement.PublishedAt
	}
	if announcement.CreatedAt.IsZero() {
		announcement.CreatedAt = announcement.PublishedAt
	}
	announcement.IsRead = false

	// Prepend new announcement
	active = append([]model.Notification{announcement}, active...)

	// Sort newest published_at first
	sort.Slice(active, func(i, j int) bool {
		return active[i].PublishedAt.After(active[j].PublishedAt)
	})

	return atomicWriteJSON(getAnnouncementsFilePath(), active)
}

// saveAnnouncementToFile is a helper alias for SaveAnnouncement.
func saveAnnouncementToFile(announcement model.Notification) error {
	return SaveAnnouncement(announcement)
}
