package engine

import (
	"time"

	"github.com/muadzyamani/oasis-cli/internal/storage"
)

// CalculateStreak walks backward starting from referenceTime (or yesterday if today has no session)
// and returns the count of consecutive calendar days with completed focus sessions.
func CalculateStreak(dailyRecords map[string]int, referenceTime time.Time) int {
	if dailyRecords == nil {
		return 0
	}

	todayStr := referenceTime.Format("2006-01-02")
	yesterdayStr := referenceTime.AddDate(0, 0, -1).Format("2006-01-02")

	hasToday := dailyRecords[todayStr] > 0
	hasYesterday := dailyRecords[yesterdayStr] > 0

	// If no session was completed today and none yesterday, the streak is reset.
	if !hasToday && !hasYesterday {
		return 0
	}

	// Start backward traversal.
	// If the user hasn't logged a session today but has yesterday, the streak is still active.
	currentDate := referenceTime
	if !hasToday {
		currentDate = referenceTime.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		dateStr := currentDate.Format("2006-01-02")
		if dailyRecords[dateStr] > 0 {
			streak++
			currentDate = currentDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak
}

// UpdateStats logs completed focus minutes for the current day,
// updates the current streak, and adjusts the longest streak if surpassed.
func UpdateStats(stats *storage.StatsState, focusMinutes int, now time.Time) {
	if stats.DailyRecords == nil {
		stats.DailyRecords = make(map[string]int)
	}

	dateStr := now.Format("2006-01-02")
	stats.DailyRecords[dateStr] += focusMinutes

	stats.CurrentStreak = CalculateStreak(stats.DailyRecords, now)
	if stats.CurrentStreak > stats.LongestStreak {
		stats.LongestStreak = stats.CurrentStreak
	}
}

// GetTierForMinutes maps total focus minutes to Oasis tier (0 to 5).
func GetTierForMinutes(minutes int) int {
	if minutes < 25 {
		return 0
	}
	if minutes < 120 {
		return 1
	}
	if minutes < 300 {
		return 2
	}
	if minutes < 600 {
		return 3
	}
	if minutes < 1200 {
		return 4
	}
	return 5
}
