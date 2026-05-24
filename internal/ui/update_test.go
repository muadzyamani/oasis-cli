package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muadzyamani/oasis-cli/internal/engine"
	"github.com/muadzyamani/oasis-cli/internal/storage"
)

func TestUpdate_Navigation(t *testing.T) {
	state := storage.DefaultState()
	m := NewModel(state, "test_data/state.json")

	if m.ActiveTab != TabOasis {
		t.Fatalf("expected TabOasis, got %s", m.ActiveTab)
	}

	resModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("tab")})
	updated := resModel.(Model)
	if updated.ActiveTab != TabStats {
		t.Errorf("expected TabStats, got %s", updated.ActiveTab)
	}

	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updated = resModel.(Model)
	if updated.ActiveTab != TabOasis {
		t.Errorf("expected TabOasis, got %s", updated.ActiveTab)
	}
}

func TestUpdate_Resize(t *testing.T) {
	state := storage.DefaultState()
	m := NewModel(state, "test_data/state.json")

	resModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	updated := resModel.(Model)

	if updated.Width != 100 || updated.Height != 30 {
		t.Errorf("expected 100x30, got %dx%d", updated.Width, updated.Height)
	}
}

func TestUpdate_PomodoroKeys(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oasis_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.json")
	state := storage.DefaultState()
	m := NewModel(state, dbPath)

	// 1. Test space key starts the timer
	if m.Timer.State != engine.StateIdle {
		t.Fatalf("expected Idle state, got %s", m.Timer.State)
	}
	resModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated := resModel.(Model)
	if updated.Timer.State != engine.StateRunning {
		t.Errorf("expected Running state after space, got %s", updated.Timer.State)
	}

	// 2. Test space key pauses the timer
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated = resModel.(Model)
	if updated.Timer.State != engine.StatePaused {
		t.Errorf("expected Paused state after space, got %s", updated.Timer.State)
	}

	// 3. Test space key resumes the timer
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated = resModel.(Model)
	if updated.Timer.State != engine.StateRunning {
		t.Errorf("expected Running state after space, got %s", updated.Timer.State)
	}

	// 4. Test up arrow adds 1 minute
	initialRemaining := updated.Timer.TimeRemaining
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	updated = resModel.(Model)
	if updated.Timer.TimeRemaining != initialRemaining+time.Minute {
		t.Errorf("expected Up arrow to add 1 minute, got %v", updated.Timer.TimeRemaining)
	}

	// 4b. Test down arrow subtracts 1 minute
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	updated = resModel.(Model)
	if updated.Timer.TimeRemaining != initialRemaining {
		t.Errorf("expected Down arrow to subtract 1 minute, got %v", updated.Timer.TimeRemaining)
	}

	// 5. Test left arrow resets the timer
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("left")})
	updated = resModel.(Model)
	if updated.Timer.State != engine.StateIdle {
		t.Errorf("expected Idle state after reset (left arrow), got %s", updated.Timer.State)
	}
	if updated.Timer.TimeRemaining != 25*time.Minute {
		t.Errorf("expected reset to set remaining time to 25m, got %v", updated.Timer.TimeRemaining)
	}

	// 5b. Test down arrow clamping at 1 minute
	// Set remaining time to 1 minute
	updated.Timer.TimeRemaining = time.Minute
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	updated = resModel.(Model)
	if updated.Timer.TimeRemaining != time.Minute {
		t.Errorf("expected remaining time to clamp at 1 minute, got %v", updated.Timer.TimeRemaining)
	}

	// 5c. Test left arrow resets timer in Idle state when time was modified
	updated.Timer.TimeRemaining = 30 * time.Minute
	updated.Timer.Duration = 30 * time.Minute
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("left")})
	updated = resModel.(Model)
	if updated.Timer.TimeRemaining != 25*time.Minute {
		t.Errorf("expected reset in Idle state to restore remaining time to 25m, got %v", updated.Timer.TimeRemaining)
	}
	if updated.Timer.Duration != 25*time.Minute {
		t.Errorf("expected reset in Idle state to restore duration to 25m, got %v", updated.Timer.Duration)
	}

	// 6. Test 's' key skips focus when Idle (toggles to short break)
	if updated.Timer.SessionType != "focus" {
		t.Fatalf("expected Focus session, got %s", updated.Timer.SessionType)
	}
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated = resModel.(Model)
	if updated.Timer.SessionType != "short-break" {
		t.Errorf("expected Skip to switch focus to break when Idle, got %s", updated.Timer.SessionType)
	}

	// 6b. Toggle back to Focus when Idle
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated = resModel.(Model)
	if updated.Timer.SessionType != "focus" {
		t.Errorf("expected toggle back to focus, got %s", updated.Timer.SessionType)
	}

	// 6c. Start focus session and skip it to complete focus session
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated = resModel.(Model)
	if updated.Timer.State != engine.StateRunning || updated.Timer.SessionType != "focus" {
		t.Fatalf("expected running focus session")
	}
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated = resModel.(Model)
	if updated.State.Oasis.TotalFocusMinutes != 25 {
		t.Errorf("expected total focus minutes to be 25, got %d", updated.State.Oasis.TotalFocusMinutes)
	}

	// 7. Test 's' key skips break session during active run
	// Note: Timer completed focus and automatically set next type to short-break (since autoStart is false, state is Idle)
	if updated.Timer.SessionType != "short-break" || updated.Timer.State != engine.StateIdle {
		t.Fatalf("expected Idle short-break, got type %s state %s", updated.Timer.SessionType, updated.Timer.State)
	}
	// Start the short-break session
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated = resModel.(Model)
	if updated.Timer.State != engine.StateRunning {
		t.Fatalf("expected Running break, got %s", updated.Timer.State)
	}
	// Press 's' to skip active break (transitions back to focus and state Idle)
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	updated = resModel.(Model)
	if updated.Timer.SessionType != "focus" {
		t.Errorf("expected Skip to transition short-break to focus, got %s", updated.Timer.SessionType)
	}
	if updated.Timer.State != engine.StateIdle {
		t.Errorf("expected timer Idle after skip completion, got %s", updated.Timer.State)
	}
}

func TestUpdate_Ticking(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oasis_test_tick_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.json")
	state := storage.DefaultState()
	m := NewModel(state, dbPath)

	// Start the timer
	resModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated := resModel.(Model)
	if updated.Timer.State != engine.StateRunning {
		t.Fatalf("expected timer to be running, got %s", updated.Timer.State)
	}
	if cmd == nil {
		t.Fatal("expected tickCmd to be returned upon starting")
	}

	// Send a tickMsg
	now := time.Now()
	resModel, nextCmd := updated.Update(tickMsg(now))
	ticked := resModel.(Model)

	expectedRemaining := 25*time.Minute - time.Second
	if ticked.Timer.TimeRemaining != expectedRemaining {
		t.Errorf("expected TimeRemaining to be %v, got %v", expectedRemaining, ticked.Timer.TimeRemaining)
	}
	if nextCmd == nil {
		t.Error("expected subsequent tickCmd to be returned from tickMsg handling")
	}
}

func TestUpdate_HotkeyEToggle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oasis_test_toggle_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.json")
	state := storage.DefaultState()
	m := NewModel(state, dbPath)

	// 1. Initial type is "focus", state is StateIdle
	if m.Timer.SessionType != "focus" {
		t.Fatalf("expected initial type 'focus', got %s", m.Timer.SessionType)
	}
	if m.Timer.State != engine.StateIdle {
		t.Fatalf("expected initial state 'idle', got %s", m.Timer.State)
	}

	// 2. Press 'e': focus -> short-break
	resModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	updated := resModel.(Model)
	if updated.Timer.SessionType != "short-break" {
		t.Errorf("expected toggle to 'short-break', got %s", updated.Timer.SessionType)
	}
	if updated.Timer.TimeRemaining != time.Duration(updated.Timer.Settings.ShortBreakDuration)*time.Minute {
		t.Errorf("expected time remaining %dm, got %v", updated.Timer.Settings.ShortBreakDuration, updated.Timer.TimeRemaining)
	}

	// 3. Press 'e': short-break -> long-break
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	updated = resModel.(Model)
	if updated.Timer.SessionType != "long-break" {
		t.Errorf("expected toggle to 'long-break', got %s", updated.Timer.SessionType)
	}
	if updated.Timer.TimeRemaining != time.Duration(updated.Timer.Settings.LongBreakDuration)*time.Minute {
		t.Errorf("expected time remaining %dm, got %v", updated.Timer.Settings.LongBreakDuration, updated.Timer.TimeRemaining)
	}

	// 4. Press 'e': long-break -> focus
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	updated = resModel.(Model)
	if updated.Timer.SessionType != "focus" {
		t.Errorf("expected toggle back to 'focus', got %s", updated.Timer.SessionType)
	}

	// 5. Start focus session (state becomes StateRunning)
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated = resModel.(Model)
	if updated.Timer.State != engine.StateRunning {
		t.Fatalf("expected running state, got %s", updated.Timer.State)
	}

	// 6. Press 'e' while running: should stop/abandon current session and toggle to short-break
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	updated = resModel.(Model)
	if updated.Timer.SessionType != "short-break" {
		t.Errorf("expected toggle to 'short-break' from running focus, got %s", updated.Timer.SessionType)
	}
	if updated.Timer.State != engine.StateIdle {
		t.Errorf("expected state to reset to idle, got %s", updated.Timer.State)
	}

	// Verify that the abandoned session is in the DB
	if len(updated.State.Sessions) != 1 {
		t.Errorf("expected 1 session in DB log, got %d", len(updated.State.Sessions))
	} else if updated.State.Sessions[0].Status != "abandoned" {
		t.Errorf("expected first session status 'abandoned', got %s", updated.State.Sessions[0].Status)
	}
}

func TestUpdate_SettingsLongBreakDuration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oasis_test_settings_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.json")
	state := storage.DefaultState()
	m := NewModel(state, dbPath)

	// Switch to TabSettings
	m.ActiveTab = TabSettings
	m.SettingsCursor = 2 // Long Break Duration

	// Press left to decrement Long Break Duration
	initialLongBreak := m.State.Settings.LongBreakDuration
	resModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("left")})
	updated := resModel.(Model)
	if updated.State.Settings.LongBreakDuration != initialLongBreak-1 {
		t.Errorf("expected LongBreakDuration to decrement, got %d", updated.State.Settings.LongBreakDuration)
	}

	// Press right to increment Long Break Duration
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("right")})
	updated = resModel.(Model)
	if updated.State.Settings.LongBreakDuration != initialLongBreak {
		t.Errorf("expected LongBreakDuration to increment back, got %d", updated.State.Settings.LongBreakDuration)
	}
}

func TestUpdate_SettingsArabicNumerals(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "oasis_test_settings_arabic_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "state.json")
	state := storage.DefaultState()
	m := NewModel(state, dbPath)

	m.ActiveTab = TabSettings
	m.SettingsCursor = 5 // Arabic Numerals

	if m.State.Settings.UseArabicNumerals {
		t.Fatal("expected UseArabicNumerals to be false initially")
	}

	// Press left to toggle it
	resModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("left")})
	updated := resModel.(Model)
	if !updated.State.Settings.UseArabicNumerals {
		t.Error("expected UseArabicNumerals to toggle to true")
	}
	if !updated.Timer.Settings.UseArabicNumerals {
		t.Error("expected timer settings UseArabicNumerals to toggle to true")
	}

	// Load from disk to verify persistence
	loadedState, err := storage.LoadState(dbPath)
	if err != nil {
		t.Fatalf("failed to load state from disk: %v", err)
	}
	if !loadedState.Settings.UseArabicNumerals {
		t.Error("expected persisted UseArabicNumerals to be true")
	}

	// Press right to toggle back to false
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("right")})
	updated = resModel.(Model)
	if updated.State.Settings.UseArabicNumerals {
		t.Error("expected UseArabicNumerals to toggle back to false")
	}

	// Press space to toggle back to true
	resModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	updated = resModel.(Model)
	if !updated.State.Settings.UseArabicNumerals {
		t.Error("expected UseArabicNumerals to toggle to true on space key")
	}
}

