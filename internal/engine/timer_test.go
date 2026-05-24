package engine

import (
	"testing"
	"time"

	"github.com/muadzyamani/oasis-cli/internal/storage"
)

func TestNewTimer(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration:      25,
		ShortBreakDuration: 5,
		LongBreakDuration:  15,
		LongBreakInterval:  4,
	}

	// 1. Initialized after nothing (or break completed): should be focus
	timer := NewTimer(settings, "", 0)
	if timer.SessionType != "focus" {
		t.Errorf("expected session type focus, got %s", timer.SessionType)
	}
	if timer.Duration != 25*time.Minute {
		t.Errorf("expected duration 25m, got %v", timer.Duration)
	}
	if timer.State != StateIdle {
		t.Errorf("expected state idle, got %s", timer.State)
	}

	// 2. Initialized after focus completed, but count < LongBreakInterval
	timer = NewTimer(settings, "focus", 3)
	if timer.SessionType != "short-break" {
		t.Errorf("expected short-break, got %s", timer.SessionType)
	}
	if timer.Duration != 5*time.Minute {
		t.Errorf("expected duration 5m, got %v", timer.Duration)
	}

	// 3. Initialized after focus completed, count == LongBreakInterval
	timer = NewTimer(settings, "focus", 4)
	if timer.SessionType != "long-break" {
		t.Errorf("expected long-break, got %s", timer.SessionType)
	}
	if timer.Duration != 15*time.Minute {
		t.Errorf("expected duration 15m, got %v", timer.Duration)
	}
}

func TestTimerStartPauseResumeStop(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration: 25,
	}
	timer := NewTimer(settings, "", 0)
	now := time.Now()

	// Start
	session := timer.Start(now)
	if session == nil {
		t.Fatal("expected session to be created")
	}
	if timer.State != StateRunning {
		t.Errorf("expected state running, got %s", timer.State)
	}
	if session.Type != "focus" {
		t.Errorf("expected session type focus, got %s", session.Type)
	}
	if session.Status != "active" {
		t.Errorf("expected status active, got %s", session.Status)
	}
	if timer.CurrentSessionID == "" {
		t.Error("expected session ID to be populated")
	}

	// Pause
	timer.Pause()
	if timer.State != StatePaused {
		t.Errorf("expected state paused, got %s", timer.State)
	}

	// Resume
	timer.Resume()
	if timer.State != StateRunning {
		t.Errorf("expected state running, got %s", timer.State)
	}

	// Stop (abandon)
	stopTime := now.Add(10 * time.Minute)
	abandoned := timer.Stop(stopTime)
	if abandoned == nil {
		t.Fatal("expected abandoned session to be returned")
	}
	if timer.State != StateIdle {
		t.Errorf("expected state idle, got %s", timer.State)
	}
	if abandoned.Status != "abandoned" {
		t.Errorf("expected status abandoned, got %s", abandoned.Status)
	}
	if timer.CurrentSessionID != "" {
		t.Errorf("expected current session ID to be cleared, got %s", timer.CurrentSessionID)
	}
}

func TestTimerTickAndCompletion(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration:      25,
		ShortBreakDuration: 5,
		LongBreakDuration:  15,
		LongBreakInterval:  2, // triggers long break on 2nd completion
		AutoStartBreaks:    false,
		AutoStartFocus:     false,
	}

	timer := NewTimer(settings, "", 0)
	now := time.Now()

	timer.Start(now)

	// Tick some time, not completed
	completed, compSess, nextSess := timer.Tick(10*time.Minute, now.Add(10*time.Minute))
	if completed {
		t.Error("expected completed to be false")
	}
	if compSess != nil || nextSess != nil {
		t.Error("expected no sessions returned on intermediate tick")
	}
	if timer.TimeRemaining != 15*time.Minute {
		t.Errorf("expected 15m remaining, got %v", timer.TimeRemaining)
	}

	// Tick remaining time to complete
	completed, compSess, nextSess = timer.Tick(15*time.Minute, now.Add(25*time.Minute))
	if !completed {
		t.Fatal("expected completed to be true")
	}
	if compSess == nil {
		t.Fatal("expected completedSession to be returned")
	}
	if compSess.Status != "complete" {
		t.Errorf("expected complete status, got %s", compSess.Status)
	}
	if timer.FocusSessionsCompleted != 1 {
		t.Errorf("expected FocusSessionsCompleted to be 1, got %d", timer.FocusSessionsCompleted)
	}
	// Next type is short-break because LongBreakInterval is 2
	if timer.SessionType != "short-break" {
		t.Errorf("expected next session type to be short-break, got %s", timer.SessionType)
	}
	if timer.State != StateIdle {
		t.Errorf("expected timer state to revert to idle when auto-start breaks is false, got %s", timer.State)
	}
	if nextSess != nil {
		t.Error("expected no auto-started session when AutoStartBreaks is false")
	}

	// Start the break manually
	timer.Start(now.Add(25 * time.Minute))
	if timer.SessionType != "short-break" {
		t.Errorf("expected running short-break, got %s", timer.SessionType)
	}

	// Complete the break
	completed, compSess, nextSess = timer.Tick(5*time.Minute, now.Add(30*time.Minute))
	if !completed || compSess == nil {
		t.Fatal("expected break session to complete")
	}
	if compSess.Type != "short-break" {
		t.Errorf("expected completed session type short-break, got %s", compSess.Type)
	}
	if timer.SessionType != "focus" {
		t.Errorf("expected next type to cycle back to focus, got %s", timer.SessionType)
	}

	// Start second focus session
	timer.Start(now.Add(30 * time.Minute))
	// Complete second focus session
	completed, compSess, nextSess = timer.Tick(25*time.Minute, now.Add(55*time.Minute))
	if !completed {
		t.Fatal("expected focus session 2 to complete")
	}
	if timer.FocusSessionsCompleted != 2 {
		t.Errorf("expected FocusSessionsCompleted to be 2, got %d", timer.FocusSessionsCompleted)
	}
	// FocusSessionsCompleted (2) % LongBreakInterval (2) == 0 -> should trigger long break!
	if timer.SessionType != "long-break" {
		t.Errorf("expected next type to cycle to long-break, got %s", timer.SessionType)
	}
	if timer.Duration != 15*time.Minute {
		t.Errorf("expected duration to be 15m, got %v", timer.Duration)
	}
}

func TestTimerAutoStart(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration:      25,
		ShortBreakDuration: 5,
		AutoStartBreaks:    true,
		AutoStartFocus:     true,
		LongBreakInterval:  4,
	}

	timer := NewTimer(settings, "", 0)
	now := time.Now()

	// Start focus
	timer.Start(now)

	// Complete focus, should auto-start break
	completed, compSess, nextSess := timer.Tick(25*time.Minute, now.Add(25*time.Minute))
	if !completed {
		t.Fatal("expected completed")
	}
	if compSess == nil {
		t.Fatal("expected completedSession")
	}
	if nextSess == nil {
		t.Fatal("expected nextSess (auto-started break) to be populated")
	}
	if nextSess.Type != "short-break" {
		t.Errorf("expected auto-started break session, got type %s", nextSess.Type)
	}
	if timer.State != StateRunning {
		t.Errorf("expected state to remain running, got %s", timer.State)
	}
	if timer.SessionType != "short-break" {
		t.Errorf("expected active session type to be short-break, got %s", timer.SessionType)
	}

	// Complete break, should auto-start focus
	completed, compSess, nextSess = timer.Tick(5*time.Minute, now.Add(30*time.Minute))
	if !completed {
		t.Fatal("expected break to complete")
	}
	if nextSess == nil {
		t.Fatal("expected nextSess (auto-started focus) to be populated")
	}
	if nextSess.Type != "focus" {
		t.Errorf("expected auto-started focus session, got type %s", nextSess.Type)
	}
	if timer.SessionType != "focus" {
		t.Errorf("expected active session type focus, got %s", timer.SessionType)
	}
}

func TestGetCompletedFocusCountInCurrentCycle(t *testing.T) {
	sessions := []storage.Session{
		{Type: "focus", Status: "complete"},
		{Type: "short-break", Status: "complete"},
		{Type: "focus", Status: "complete"},
		{Type: "short-break", Status: "complete"},
		{Type: "focus", Status: "abandoned"}, // does not count
		{Type: "focus", Status: "complete"},
		{Type: "long-break", Status: "complete"}, // resets cycle
		{Type: "focus", Status: "complete"},
		{Type: "short-break", Status: "complete"},
		{Type: "focus", Status: "complete"},
	}

	count := GetCompletedFocusCountInCurrentCycle(sessions)
	if count != 2 {
		t.Errorf("expected 2 completed focus sessions in current cycle, got %d", count)
	}

	// Empty list
	if countZero := GetCompletedFocusCountInCurrentCycle([]storage.Session{}); countZero != 0 {
		t.Errorf("expected 0, got %d", countZero)
	}
}

func TestTimerUpdateSettings(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration: 25,
	}
	timer := NewTimer(settings, "", 0)

	newSettings := storage.SettingsState{
		FocusDuration: 45,
	}
	timer.UpdateSettings(newSettings)
	if timer.Settings.FocusDuration != 45 {
		t.Errorf("expected updated FocusDuration 45, got %d", timer.Settings.FocusDuration)
	}
	if timer.Duration != 45*time.Minute {
		t.Errorf("expected duration to update to 45m, got %v", timer.Duration)
	}
}

func TestTimerInvalidTransitions(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration: 25,
	}
	timer := NewTimer(settings, "", 0)

	// Stop when idle
	if s := timer.Stop(time.Now()); s != nil {
		t.Error("expected Stop on idle timer to return nil")
	}

	// Start running
	timer.Start(time.Now())
	// Start again while running
	if s := timer.Start(time.Now()); s != nil {
		t.Error("expected Start on already running timer to return nil")
	}
}

