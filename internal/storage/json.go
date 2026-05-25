package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// ApplicationState represents the root state schema saved in JSON.
type ApplicationState struct {
	Oasis    OasisState    `json:"oasis"`
	Sessions []Session     `json:"sessions"`
	Settings SettingsState `json:"settings"`
	Stats    StatsState    `json:"stats"`
}

// OasisState stores the current growth progress.
type OasisState struct {
	Name              string    `json:"name"`
	TotalFocusMinutes int       `json:"total_focus_minutes"`
	CreatedAt         time.Time `json:"created_at"`
}


// Session represents a completed or active timer session.
type Session struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"` // "focus" | "short-break" | "long-break"
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	DurationMinutes int       `json:"duration_minutes"`
	Status          string    `json:"status"` // "active" | "complete" | "abandoned"
}

// SettingsState stores customizable timer configurations and visuals.
type SettingsState struct {
	SoundEnabled       bool `json:"sound_enabled"`
	FocusDuration      int  `json:"focus_duration"`
	ShortBreakDuration int  `json:"short_break_duration"`
	LongBreakDuration  int  `json:"long_break_duration"`
	LongBreakInterval  int  `json:"long_break_interval"`
	AutoStartBreaks    bool `json:"auto_start_breaks"`
	AutoStartFocus     bool `json:"auto_start_focus"`
	ShowStreak         bool `json:"show_streak"`
	UseArabicNumerals  bool `json:"use_arabic_numerals"`
}

// StatsState stores streak counters and daily focus logs.
type StatsState struct {
	CurrentStreak int            `json:"current_streak"`
	LongestStreak int            `json:"longest_streak"`
	DailyRecords  map[string]int `json:"daily_records"` // YYYY-MM-DD -> minutes
}

// GetDefaultConfigPath returns the local path to the state JSON.
func GetDefaultConfigPath() (string, error) {
	return "data/state.json", nil
}

// DefaultState returns a clean initialized ApplicationState with sensible defaults.
func DefaultState() *ApplicationState {
	return &ApplicationState{
		Oasis: OasisState{
			Name:              "My Oasis",
			TotalFocusMinutes: 0,
			CreatedAt:         time.Now(),
		},
		Sessions: []Session{},
		Settings: SettingsState{
			SoundEnabled:       true,
			FocusDuration:      25,
			ShortBreakDuration: 5,
			LongBreakDuration:  15,
			LongBreakInterval:  4,
			AutoStartBreaks:    false,
			AutoStartFocus:     false,
			ShowStreak:         true,
			UseArabicNumerals:  false,
		},
		Stats: StatsState{
			CurrentStreak: 0,
			LongestStreak: 0,
			DailyRecords:  make(map[string]int),
		},
	}
}

// LoadState reads the JSON file at the provided path and parses it.
// If the file does not exist, it returns a new default state.
func LoadState(path string) (*ApplicationState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return DefaultState(), nil
		}
		return nil, err
	}

	state := DefaultState()
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}

	// Ensure collections are initialized rather than nil
	if state.Sessions == nil {
		state.Sessions = []Session{}
	}
	if state.Stats.DailyRecords == nil {
		state.Stats.DailyRecords = make(map[string]int)
	}

	// Clean up any active sessions from a previous run
	hasActive := false
	for i, sess := range state.Sessions {
		if sess.Status == "active" {
			state.Sessions[i].Status = "abandoned"
			state.Sessions[i].CompletedAt = time.Now()
			hasActive = true
		}
	}
	if hasActive {
		if err := SaveState(path, state); err != nil {
			return nil, err
		}
	}

	return state, nil
}

// SaveState marshals and writes the state to path using an atomic temp-swap method.
// Parent directories are automatically created if they do not exist.
func SaveState(path string, state *ApplicationState) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := path + ".tmp"
	file, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		_ = os.Remove(tmpFile)
	}()

	if _, err := file.Write(data); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFile, path)
}
