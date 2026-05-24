package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/muadzyamani/oasis-cli/internal/storage"
)

type TimerState string

const (
	StateIdle      TimerState = "idle"
	StateRunning   TimerState = "running"
	StatePaused    TimerState = "paused"
	StateCompleted TimerState = "completed"
)

type Timer struct {
	State         TimerState
	SessionType   string // "focus" | "short-break" | "long-break"
	Duration      time.Duration
	TimeRemaining time.Duration

	CurrentSessionID string

	// Settings
	Settings storage.SettingsState

	// History / Streak / Counter tracker
	FocusSessionsCompleted int
	LongBreakInterval      int
}

// GenerateID creates a simple cryptographically secure random hexadecimal ID.
func GenerateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp in nano if random source fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GetCompletedFocusCountInCurrentCycle counts consecutive completed focus sessions since the last completed long-break.
func GetCompletedFocusCountInCurrentCycle(sessions []storage.Session) int {
	count := 0
	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		if s.Status == "complete" {
			if s.Type == "long-break" {
				break
			}
			if s.Type == "focus" {
				count++
			}
		}
	}
	return count
}

// NewTimer initializes a new Timer with given settings, previous session history context, and completed focus count.
func NewTimer(settings storage.SettingsState, lastCompletedType string, completedFocusCount int) *Timer {
	t := &Timer{
		State:                  StateIdle,
		Settings:               settings,
		FocusSessionsCompleted: completedFocusCount,
		LongBreakInterval:      settings.LongBreakInterval,
	}

	if t.LongBreakInterval <= 0 {
		t.LongBreakInterval = 4
	}

	// Determine starting SessionType based on last completed session
	if lastCompletedType == "focus" {
		if completedFocusCount > 0 && completedFocusCount%t.LongBreakInterval == 0 {
			t.SessionType = "long-break"
			t.Duration = time.Duration(settings.LongBreakDuration) * time.Minute
		} else {
			t.SessionType = "short-break"
			t.Duration = time.Duration(settings.ShortBreakDuration) * time.Minute
		}
	} else {
		t.SessionType = "focus"
		t.Duration = time.Duration(settings.FocusDuration) * time.Minute
	}

	t.TimeRemaining = t.Duration
	return t
}

// UpdateSettings updates the timer settings, adjusting active properties if idle.
func (t *Timer) UpdateSettings(settings storage.SettingsState) {
	t.Settings = settings
	t.LongBreakInterval = settings.LongBreakInterval
	if t.LongBreakInterval <= 0 {
		t.LongBreakInterval = 4
	}
	if t.State == StateIdle {
		t.Duration = t.getDurationForType(t.SessionType)
		t.TimeRemaining = t.Duration
	}
}

func (t *Timer) getDurationForType(sType string) time.Duration {
	switch sType {
	case "focus":
		return time.Duration(t.Settings.FocusDuration) * time.Minute
	case "short-break":
		return time.Duration(t.Settings.ShortBreakDuration) * time.Minute
	case "long-break":
		return time.Duration(t.Settings.LongBreakDuration) * time.Minute
	default:
		return time.Duration(t.Settings.FocusDuration) * time.Minute
	}
}

// Start starts the current timer session, transitioning to StateRunning.
func (t *Timer) Start(now time.Time) *storage.Session {
	if t.State != StateIdle && t.State != StateCompleted {
		return nil
	}

	t.State = StateRunning
	t.CurrentSessionID = GenerateID()

	return &storage.Session{
		ID:              t.CurrentSessionID,
		Type:            t.SessionType,
		StartedAt:       now,
		DurationMinutes: int(t.Duration.Minutes()),
		Status:          "active",
	}
}

// Pause pauses the running timer session, transitioning to StatePaused.
func (t *Timer) Pause() {
	if t.State == StateRunning {
		t.State = StatePaused
	}
}

// Resume resumes a paused timer session, transitioning to StateRunning.
func (t *Timer) Resume() {
	if t.State == StatePaused {
		t.State = StateRunning
	}
}

// Stop abandons the running or paused session, reverting to StateIdle.
func (t *Timer) Stop(now time.Time) *storage.Session {
	if t.State != StateRunning && t.State != StatePaused {
		return nil
	}

	abandonedSession := &storage.Session{
		ID:          t.CurrentSessionID,
		Type:        t.SessionType,
		StartedAt:   now.Add(-t.Duration + t.TimeRemaining), // approximate start
		CompletedAt: now,
		Status:      "abandoned",
	}

	t.State = StateIdle
	t.TimeRemaining = t.Duration
	t.CurrentSessionID = ""

	return abandonedSession
}

// Reset resets the timer duration and remaining time to the default value for the current session type and sets state to idle.
func (t *Timer) Reset() {
	t.Duration = t.getDurationForType(t.SessionType)
	t.TimeRemaining = t.Duration
	t.State = StateIdle
	t.CurrentSessionID = ""
}

// Tick decrements the time remaining and checks for session completion.
// Returns (completed, completedSession, autoStartedSession).
func (t *Timer) Tick(delta time.Duration, now time.Time) (bool, *storage.Session, *storage.Session) {
	if t.State != StateRunning {
		return false, nil, nil
	}

	t.TimeRemaining -= delta
	if t.TimeRemaining > 0 {
		return false, nil, nil
	}

	// Completion reached
	t.TimeRemaining = 0
	t.State = StateCompleted

	completedSession := &storage.Session{
		ID:              t.CurrentSessionID,
		Type:            t.SessionType,
		StartedAt:       now.Add(-t.Duration),
		CompletedAt:     now,
		DurationMinutes: int(t.Duration.Minutes()),
		Status:          "complete",
	}

	if t.SessionType == "focus" {
		t.FocusSessionsCompleted++
	}

	// Transition to next session type
	var nextSession *storage.Session
	var autoStart bool

	if t.SessionType == "focus" {
		if t.FocusSessionsCompleted > 0 && t.FocusSessionsCompleted%t.LongBreakInterval == 0 {
			t.SessionType = "long-break"
		} else {
			t.SessionType = "short-break"
		}
		t.Duration = t.getDurationForType(t.SessionType)
		t.TimeRemaining = t.Duration
		autoStart = t.Settings.AutoStartBreaks
	} else {
		t.SessionType = "focus"
		t.Duration = t.getDurationForType(t.SessionType)
		t.TimeRemaining = t.Duration
		autoStart = t.Settings.AutoStartFocus
	}

	if autoStart {
		t.State = StateRunning
		t.CurrentSessionID = GenerateID()
		nextSession = &storage.Session{
			ID:              t.CurrentSessionID,
			Type:            t.SessionType,
			StartedAt:       now,
			DurationMinutes: int(t.Duration.Minutes()),
			Status:          "active",
		}
	} else {
		t.State = StateIdle
		t.CurrentSessionID = ""
	}

	return true, completedSession, nextSession
}
