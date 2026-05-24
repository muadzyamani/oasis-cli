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

	if state.Oasis.Elements == nil {
		t.Error("expected default elements to be non-nil")
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
	customState.Oasis.Tier = 3
	customState.Oasis.TotalFocusMinutes = 120
	customState.Oasis.Elements = append(customState.Oasis.Elements, OasisElement{
		ID:        "elem-1",
		Type:      "palm",
		PlantedAt: time.Now().Round(time.Second), // Round to avoid precision diff on serialization
		SessionID: "sess-1",
		Label:     "A lovely palm",
		X:         50,
		Y:         60,
		Stage:     "mature",
	})
	customState.Sessions = append(customState.Sessions, Session{
		ID:              "sess-1",
		Type:            "focus",
		StartedAt:       time.Now().Add(-25 * time.Minute).Round(time.Second),
		CompletedAt:     time.Now().Round(time.Second),
		DurationMinutes: 25,
		Status:          "complete",
		OasisElementID:  "elem-1",
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

	if loadedState.Oasis.Tier != 3 {
		t.Errorf("expected oasis tier 3, got %d", loadedState.Oasis.Tier)
	}

	if loadedState.Settings.FocusDuration != 50 {
		t.Errorf("expected focus duration 50, got %d", loadedState.Settings.FocusDuration)
	}

	if len(loadedState.Oasis.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(loadedState.Oasis.Elements))
	}

	elem := loadedState.Oasis.Elements[0]
	if elem.ID != "elem-1" || elem.Type != "palm" || elem.Label != "A lovely palm" || elem.X != 50 || elem.Y != 60 || elem.Stage != "mature" {
		t.Errorf("element details mismatch: %+v", elem)
	}

	// Compare timestamps allowing small formatting or zone discrepancies, but they should match
	if !elem.PlantedAt.Equal(customState.Oasis.Elements[0].PlantedAt) {
		t.Errorf("expected element planted time %v, got %v", customState.Oasis.Elements[0].PlantedAt, elem.PlantedAt)
	}

	if len(loadedState.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(loadedState.Sessions))
	}

	sess := loadedState.Sessions[0]
	if sess.ID != "sess-1" || sess.Type != "focus" || sess.DurationMinutes != 25 || sess.Status != "complete" || sess.OasisElementID != "elem-1" {
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
