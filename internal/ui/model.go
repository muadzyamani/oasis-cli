package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muadzyamani/oasis-cli/internal/engine"
	"github.com/muadzyamani/oasis-cli/internal/storage"
)

type Tab string

const (
	TabOasis    Tab = "oasis"
	TabStats    Tab = "stats"
	TabSettings Tab = "settings"
)

type Model struct {
	ActiveTab Tab
	Width     int
	Height    int

	// State Database
	State  *storage.ApplicationState
	DbPath string

	// Logic Engines
	Timer   *engine.Timer

	// App State
	Ready          bool
	Err            error
	SettingsCursor int
}

// NewModel constructs a Model initialized from persistence state.
func NewModel(state *storage.ApplicationState, dbPath string) Model {
	// Initialize completed focus count in current cycle
	completedFocusCount := engine.GetCompletedFocusCountInCurrentCycle(state.Sessions)

	// Retrieve last completed session type
	lastCompletedType := ""
	for i := len(state.Sessions) - 1; i >= 0; i-- {
		if state.Sessions[i].Status == "complete" {
			lastCompletedType = state.Sessions[i].Type
			break
		}
	}

	timer := engine.NewTimer(state.Settings, lastCompletedType, completedFocusCount)

	// Recalculate streak to handle decay when app starts
	oldStreak := state.Stats.CurrentStreak
	state.Stats.CurrentStreak = engine.CalculateStreak(state.Stats.DailyRecords, time.Now())
	if oldStreak != state.Stats.CurrentStreak {
		_ = storage.SaveState(dbPath, state)
	}

	return Model{
		ActiveTab: TabOasis,
		State:     state,
		DbPath:    dbPath,
		Timer:     timer,
		Ready:     false,
	}
}

// Init initializes the Bubble Tea loop.
func (m Model) Init() tea.Cmd {
	return nil
}
