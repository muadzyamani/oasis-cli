package engine

import (
	"testing"
	"time"

	"github.com/muadzyamani/oasis-cli/internal/storage"
)

func TestCalculateStreak(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)

	// 1. Nil or empty daily records
	if streak := CalculateStreak(nil, now); streak != 0 {
		t.Errorf("expected 0, got %d", streak)
	}
	if streak := CalculateStreak(map[string]int{}, now); streak != 0 {
		t.Errorf("expected 0, got %d", streak)
	}

	// 2. Completed today, nothing yesterday
	records := map[string]int{
		"2026-05-24": 25,
	}
	if streak := CalculateStreak(records, now); streak != 1 {
		t.Errorf("expected 1, got %d", streak)
	}

	// 3. Completed yesterday, nothing today
	records = map[string]int{
		"2026-05-23": 25,
	}
	if streak := CalculateStreak(records, now); streak != 1 {
		t.Errorf("expected 1, got %d", streak)
	}

	// 4. Broken streak (completed today, gap yesterday, completed day before)
	records = map[string]int{
		"2026-05-24": 25,
		"2026-05-22": 50,
	}
	if streak := CalculateStreak(records, now); streak != 1 {
		t.Errorf("expected 1 (broken streak), got %d", streak)
	}

	// 5. Multi-day active streak (5 days including today)
	records = map[string]int{
		"2026-05-24": 25,
		"2026-05-23": 50,
		"2026-05-22": 25,
		"2026-05-21": 30,
		"2026-05-20": 25,
	}
	if streak := CalculateStreak(records, now); streak != 5 {
		t.Errorf("expected 5, got %d", streak)
	}

	// 6. Multi-day active streak (4 days ending yesterday, today not done yet)
	records = map[string]int{
		"2026-05-23": 50,
		"2026-05-22": 25,
		"2026-05-21": 30,
		"2026-05-20": 25,
	}
	if streak := CalculateStreak(records, now); streak != 4 {
		t.Errorf("expected 4, got %d", streak)
	}
}

func TestUpdateStats(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	stats := &storage.StatsState{
		CurrentStreak: 0,
		LongestStreak: 3, // existing high record
		DailyRecords:  make(map[string]int),
	}

	// Day 1: complete 25 mins
	UpdateStats(stats, 25, now)
	if stats.DailyRecords["2026-05-24"] != 25 {
		t.Errorf("expected 25 mins today, got %d", stats.DailyRecords["2026-05-24"])
	}
	if stats.CurrentStreak != 1 {
		t.Errorf("expected current streak 1, got %d", stats.CurrentStreak)
	}
	if stats.LongestStreak != 3 {
		t.Errorf("expected longest streak to remain 3, got %d", stats.LongestStreak)
	}

	// Add more focus minutes today, streak should stay 1
	UpdateStats(stats, 25, now)
	if stats.DailyRecords["2026-05-24"] != 50 {
		t.Errorf("expected 50 mins today, got %d", stats.DailyRecords["2026-05-24"])
	}
	if stats.CurrentStreak != 1 {
		t.Errorf("expected current streak 1, got %d", stats.CurrentStreak)
	}

	// Advance to Day 2
	day2 := now.AddDate(0, 0, 1)
	UpdateStats(stats, 25, day2)
	if stats.CurrentStreak != 2 {
		t.Errorf("expected current streak 2, got %d", stats.CurrentStreak)
	}
	if stats.LongestStreak != 3 {
		t.Errorf("expected longest streak to remain 3, got %d", stats.LongestStreak)
	}

	// Advance to Day 3
	day3 := day2.AddDate(0, 0, 1)
	UpdateStats(stats, 25, day3)
	if stats.CurrentStreak != 3 {
		t.Errorf("expected current streak 3, got %d", stats.CurrentStreak)
	}

	// Advance to Day 4 (exceeds longest streak)
	day4 := day3.AddDate(0, 0, 1)
	UpdateStats(stats, 25, day4)
	if stats.CurrentStreak != 4 {
		t.Errorf("expected current streak 4, got %d", stats.CurrentStreak)
	}
	if stats.LongestStreak != 4 {
		t.Errorf("expected longest streak to update to 4, got %d", stats.LongestStreak)
	}
}

func TestUpdateStatsWithNilRecords(t *testing.T) {
	stats := &storage.StatsState{
		CurrentStreak: 0,
		LongestStreak: 0,
		DailyRecords:  nil, // test lazy initialization
	}
	UpdateStats(stats, 25, time.Now())
	if stats.DailyRecords == nil {
		t.Fatal("expected DailyRecords to be initialized")
	}
	if stats.CurrentStreak != 1 {
		t.Errorf("expected current streak to be 1, got %d", stats.CurrentStreak)
	}
}

