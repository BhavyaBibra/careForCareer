package timeutil

import "time"

// DaysUntil returns the number of full days between now and the target date.
// Returns 0 if target is in the past.
func DaysUntil(target time.Time) int {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	t := target.UTC().Truncate(24 * time.Hour)
	d := int(t.Sub(today).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

// DaysSince returns the number of full days elapsed since the start date.
func DaysSince(start time.Time) int {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	s := start.UTC().Truncate(24 * time.Hour)
	d := int(today.Sub(s).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

// TodayKey returns a string key for the current UTC date: "2026-06-07".
// Used for Redis daily counters.
func TodayKey() string {
	return time.Now().UTC().Format("2006-01-02")
}
