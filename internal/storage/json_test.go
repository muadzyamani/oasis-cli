package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetDefaultConfigPath(t *testing.T) {
	path, err := GetDefaultConfigPath()
	if err != nil {
		t.Fatalf("unexpected error getting default config path: %v", err)
	}
	expected := "data/state.json"
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestDefaultState(t *testing.T) {
	state := DefaultState()
	if state == nil {
		t.Fatal("expected non-nil default state")
	}

	if state.Oasis.Name != "My Oasis" {
		t.Errorf("expected default oasis name 'My Oasis', got %q", state.Oasis.Name)
	}

	if state.Settings.FocusDuration != 25 {
		t.Errorf("expected default focus duration 25, got %d", state.Settings.FocusDuration)
	}

	if state.Sessions == nil {
		t.Error("expected default sessions to be non-nil")
	}

	if state.Stats.DailyRecords == nil {
		t.Error("expected default daily records to be non-nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "oasis-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	missingPath := filepath.Join(tempDir, "nonexistent.json")
	state, err := LoadState(missingPath)
	if err != nil {
		t.Fatalf("unexpected error loading missing file: %v", err)
	}

	if state == nil {
		t.Fatal("expected non-nil state returned for missing file")
	}

	if state.Oasis.Name != "My Oasis" {
		t.Errorf("expected default state loaded, name was %q", state.Oasis.Name)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "oasis-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "state.json")

	customState := DefaultState()
	customState.Oasis.Name = "Test Oasis"
	customState.Oasis.TotalFocusMinutes = 120
	customState.Sessions = append(customState.Sessions, Session{
		ID:              "sess-1",
		Type:            "focus",
		StartedAt:       time.Now().Add(-25 * time.Minute).Round(time.Second),
		CompletedAt:     time.Now().Round(time.Second),
		DurationMinutes: 25,
		Status:          "complete",
	})
	customState.Settings.FocusDuration = 50
	customState.Stats.CurrentStreak = 5
	customState.Stats.DailyRecords["2026-05-24"] = 25

	err = SaveState(path, customState)
	if err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	loadedState, err := LoadState(path)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loadedState.Oasis.Name != "Test Oasis" {
		t.Errorf("expected oasis name 'Test Oasis', got %q", loadedState.Oasis.Name)
	}



	if loadedState.Settings.FocusDuration != 50 {
		t.Errorf("expected focus duration 50, got %d", loadedState.Settings.FocusDuration)
	}

	if len(loadedState.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(loadedState.Sessions))
	}

	sess := loadedState.Sessions[0]
	if sess.ID != "sess-1" || sess.Type != "focus" || sess.DurationMinutes != 25 || sess.Status != "complete" {
		t.Errorf("session details mismatch: %+v", sess)
	}

	if loadedState.Stats.CurrentStreak != 5 {
		t.Errorf("expected current streak 5, got %d", loadedState.Stats.CurrentStreak)
	}

	if loadedState.Stats.DailyRecords["2026-05-24"] != 25 {
		t.Errorf("expected daily records for 2026-05-24 to be 25, got %d", loadedState.Stats.DailyRecords["2026-05-24"])
	}
}

func TestSaveCreatesDirectories(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "oasis-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a nested path where parent folder doesn't exist
	nestedPath := filepath.Join(tempDir, "nested", "folders", "deep", "state.json")

	state := DefaultState()
	err = SaveState(nestedPath, state)
	if err != nil {
		t.Fatalf("failed to save state in nested directory: %v", err)
	}

	// Verify the file was written
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Fatalf("expected state file to exist at %q", nestedPath)
	}

	// Load and verify
	loaded, err := LoadState(nestedPath)
	if err != nil {
		t.Fatalf("failed to load state from nested path: %v", err)
	}
	if loaded.Oasis.Name != "My Oasis" {
		t.Errorf("unexpected name from loaded state: %q", loaded.Oasis.Name)
	}
}

func TestAtomicReplace(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "oasis-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "state.json")
	tmpPath := path + ".tmp"

	state := DefaultState()
	err = SaveState(path, state)
	if err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Check that the temp file was cleaned up and is not left behind
	if _, err := os.Stat(tmpPath); err == nil {
		t.Error("expected temp file to be cleaned up/removed, but it exists")
	}

	// Verify the real file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected final state file to exist, but it is missing")
	}
}

func TestLoadCleansActiveSessions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "oasis-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "state.json")

	customState := DefaultState()
	customState.Sessions = append(customState.Sessions, Session{
		ID:              "sess-active",
		Type:            "focus",
		StartedAt:       time.Now().Add(-10 * time.Minute).Round(time.Second),
		DurationMinutes: 25,
		Status:          "active",
	})

	err = SaveState(path, customState)
	if err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Load the state, which should automatically convert the active session to abandoned
	loadedState, err := LoadState(path)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if len(loadedState.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(loadedState.Sessions))
	}

	sess := loadedState.Sessions[0]
	if sess.Status != "abandoned" {
		t.Errorf("expected loaded active session to be converted to abandoned, got %q", sess.Status)
	}

	if sess.CompletedAt.IsZero() {
		t.Error("expected CompletedAt to be set for the abandoned session, got zero time")
	}

	// Also verify it was persisted back to the file
	fileState, err := LoadState(path)
	if err != nil {
		t.Fatalf("failed to load state from disk second time: %v", err)
	}
	if fileState.Sessions[0].Status != "abandoned" {
		t.Errorf("expected session on disk to be abandoned, got %q", fileState.Sessions[0].Status)
	}
}
