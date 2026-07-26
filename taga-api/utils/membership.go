package utils

import "time"

// MembershipYearEnd returns the end of the membership year (March 31) that contains 'now'.
// Membership year runs April 1 – March 31.
func MembershipYearEnd(now time.Time) time.Time {
	year := now.Year()
	if now.Month() >= 4 {
		return time.Date(year+1, 3, 31, 23, 59, 59, 0, now.Location())
	}
	return time.Date(year, 3, 31, 23, 59, 59, 0, now.Location())
}


