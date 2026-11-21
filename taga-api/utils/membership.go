package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MembershipYearEnd returns the end of the membership year (March 31) that contains 'now'.
// Membership year runs April 1 – March 31.
func MembershipYearEnd(now time.Time) time.Time {
	year := now.Year()
	if now.Month() >= 4 {
		return time.Date(year+1, 3, 31, 23, 59, 59, 0, now.Location())
	}
	return time.Date(year, 3, 31, 23, 59, 59, 0, now.Location())
}

// FinancialYear returns the financial year string (e.g. "2026-2027") for the given time.
// In India, the financial year runs from April 1 to March 31.
func FinancialYear(t time.Time) string {
	year := t.Year()
	if t.Month() >= time.April {
		return fmt.Sprintf("%d-%d", year, year+1)
	}
	return fmt.Sprintf("%d-%d", year-1, year)
}

// SubscriptionValidUntil returns the ISO RFC3339 timestamp of the end of the financial year for t.
func SubscriptionValidUntil(t time.Time) string {
	return MembershipYearEnd(t).UTC().Format(time.RFC3339)
}

// IsSubscriptionValid checks whether an offline paid subscription is valid at time 'now'.
// Rules:
// 1. If validUntilStr is provided and parseable: if now is after validUntil, subscription is expired (returns false).
// 2. If yearStr is provided (e.g. "2024-2025" or "2026-2027"): if now is after March 31 of that FY, subscription is expired (returns false).
// 3. If neither is provided, returns true for backward compatibility with legacy test records.
func IsSubscriptionValid(validUntilStr, yearStr string, now time.Time) bool {
	validUntilStr = strings.TrimSpace(validUntilStr)
	yearStr = strings.TrimSpace(yearStr)

	// Check expiration date timestamp if present
	if validUntilStr != "" {
		validUntil, err := parseDate(validUntilStr, now.Location())
		if err == nil {
			if now.After(validUntil) {
				return false
			}
		}
	}

	// Check financial year expiration if present
	if yearStr != "" {
		if endYear, ok := parseEndYearFromFY(yearStr); ok {
			fyEnd := time.Date(endYear, time.March, 31, 23, 59, 59, 0, now.Location())
			if now.After(fyEnd) {
				return false
			}
		}
	}

	return true
}

func parseDate(s string, loc *time.Location) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			if f == "2006-01-02" {
				return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, loc), nil
			}
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func parseEndYearFromFY(fy string) (int, bool) {
	var start, end int
	if _, err := fmt.Sscanf(fy, "%d-%d", &start, &end); err == nil {
		if end < 100 {
			end = 2000 + end
		}
		return end, true
	}
	if _, err := fmt.Sscanf(fy, "%d/%d", &start, &end); err == nil {
		if end < 100 {
			end = 2000 + end
		}
		return end, true
	}
	if y, err := strconv.Atoi(fy); err == nil {
		return y, true
	}
	return 0, false
}



