package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFinancialYear(t *testing.T) {
	// November 2026 -> FY 2026-2027
	tNov := time.Date(2026, time.November, 21, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, "2026-2027", FinancialYear(tNov))

	// April 1, 2026 -> FY 2026-2027
	tApr := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2026-2027", FinancialYear(tApr))

	// March 31, 2027 -> FY 2026-2027
	tMar := time.Date(2027, time.March, 31, 23, 59, 59, 0, time.UTC)
	assert.Equal(t, "2026-2027", FinancialYear(tMar))

	// April 1, 2027 -> FY 2027-2028
	tNextApr := time.Date(2027, time.April, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2027-2028", FinancialYear(tNextApr))
}

func TestMembershipYearEnd(t *testing.T) {
	tNov := time.Date(2026, time.November, 21, 12, 0, 0, 0, time.UTC)
	end := MembershipYearEnd(tNov)
	assert.Equal(t, 2027, end.Year())
	assert.Equal(t, time.March, end.Month())
	assert.Equal(t, 31, end.Day())

	tJan := time.Date(2027, time.January, 15, 12, 0, 0, 0, time.UTC)
	endJan := MembershipYearEnd(tJan)
	assert.Equal(t, 2027, endJan.Year())
	assert.Equal(t, time.March, endJan.Month())
	assert.Equal(t, 31, endJan.Day())
}

func TestIsSubscriptionValid(t *testing.T) {
	now := time.Date(2026, time.November, 21, 12, 0, 0, 0, time.UTC)

	// 1. Current financial year 2026-2027 is valid today
	assert.True(t, IsSubscriptionValid("2027-03-31T23:59:59Z", "2026-2027", now))

	// 2. Past financial year 2024-2025 is expired today
	assert.False(t, IsSubscriptionValid("2025-03-31T23:59:59Z", "2024-2025", now))

	// 3. Past financial year 2025-2026 is expired today (ended March 31, 2026)
	assert.False(t, IsSubscriptionValid("2026-03-31T23:59:59Z", "2025-2026", now))

	// 4. Only year specified (e.g. "2024-2025") -> expired
	assert.False(t, IsSubscriptionValid("", "2024-2025", now))

	// 5. Only year specified (e.g. "2026-2027") -> valid
	assert.True(t, IsSubscriptionValid("", "2026-2027", now))

	// 6. Only valid_until timestamp specified (past date) -> expired
	assert.False(t, IsSubscriptionValid("2025-03-31T23:59:59Z", "", now))

	// 7. Only valid_until timestamp specified (future date) -> valid
	assert.True(t, IsSubscriptionValid("2027-03-31T23:59:59Z", "", now))

	// 8. Next year (e.g. June 2027): FY 2026-2027 MUST be expired!
	nextYearJune := time.Date(2027, time.June, 15, 10, 0, 0, 0, time.UTC)
	assert.False(t, IsSubscriptionValid("2027-03-31T23:59:59Z", "2026-2027", nextYearJune),
		"In June 2027, FY 2026-2027 membership must be expired")

	// 9. Legacy record with empty fields -> returns true (backward compatibility)
	assert.True(t, IsSubscriptionValid("", "", now))
}
