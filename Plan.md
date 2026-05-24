# Oasis — Go TUI Architecture & Implementation Plan

This document details the architectural design, user interface layout, visual rendering systems, testing strategy, and execution roadmap for building a Go-based Terminal User Interface (TUI) version of the Oasis Pomodoro productivity application.

---

## 1. Architectural Overview

The application is structured around Charmbracelet's **Bubble Tea** (an Elm-architecture framework for Go) and **Lip Gloss** (a style definition library). To ensure maximum testability, maintainability, and clean separation of concerns, the project isolates pure mathematical and logical calculations into standard Go packages, keeping the terminal-rendering and key-binding loop as thin controllers.

### Visual Layout & Rendering

The CLI TUI adopts a clean, centered, high-readability layout. Instead of dynamic environmental canvases, it centers a large digital Pomodoro clock on the screen:

```
+-----------------------------------------------------------------+
|                       🌴 OASIS POMODORO                         |
|  ● OASIS         ○ STATS         ○ SETTINGS                     |
|                                                                 |
|                         ▄███▄   ▄███▄   ▄   ▄███▄               |
|                        ██▀ ▀██ ██▀ ▀██  █  ██▀ ▀██              |
|                        ██ █ ██ ██ █ ██  ▄  ██ █ ██              |
|                        ██▄ ▄██ ██▄ ▄██  █  ██▄ ▄██              |
|                         ▀███▀   ▀███▀   ▄   ▀███▀               |
|                                                                 |
|                             focus session                       |
|                                                                 |
|        ██████████████████████████████░░░░░░░░░░░░░░░░░░░  60%   |
|                                                                 |
|       ↑ +1 min  •  space pause/resume  •  ← reset  •  s skip    |
+-----------------------------------------------------------------+
```

- **Large ASCII Digits**: The current remaining time is formatted into 5-line-high block text digits, dynamically aligned to the center.
- **Gradient Progress Bar**: A color-interpolated bar showing current session completion, fading from blue to purple-pink, with a gray-shaded unfilled track and percentage label.
- **Keyboard Guides**: Clear inline key guides for controlling the timer.

---

## 2. Directory Structure

The Go codebase will follow standard Go project layouts:

```
oasis-cli/
├── cmd/
│   └── oasis/
│       └── main.go              # CLI Entrypoint & Program initialization
├── internal/
│   ├── engine/
│   │   ├── timer.go             # Pomodoro state machine (TDD)
│   │   └── stats.go             # Streak calculations & daily metrics (TDD)
│   ├── storage/
│   │   └── json.go              # JSON local persistence engine (TDD)
│   └── ui/
│       ├── model.go             # Main Bubble Tea Model state
│       ├── update.go            # Bubble Tea Update event router
│       ├── view.go              # Bubble Tea View composer & Layout builder
│       ├── components/          # Styled Lip Gloss views
│       │   ├── timer.go         # Centered timer widget (Large digits, progress)
│       │   ├── settings.go      # Settings options list
│       │   └── stats.go         # Progress tracking & historic panels
│       └── styles/
│           └── colors.go        # Design tokens & color functions
├── go.mod
└── go.sum
```

---

## 3. Data Models & Serialization

All states are saved inside a local JSON file (`data/state.json`) within the project workspace:
- **Path**: `data/state.json`

### Database Structs (`internal/storage/json.go`)

```go
type ApplicationState struct {
	Oasis    OasisState    `json:"oasis"`
	Sessions []Session     `json:"sessions"`
	Settings SettingsState `json:"settings"`
	Stats    StatsState    `json:"stats"`
}

type OasisState struct {
	Name              string         `json:"name"`
	Tier              int            `json:"tier"`                // 0 to 5
	TotalFocusMinutes int            `json:"total_focus_minutes"`
	Elements          []OasisElement `json:"elements"`
	CreatedAt         time.Time      `json:"created_at"`
}

type OasisElement struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`       // sprout, flower, reed, palm, acacia, etc.
	PlantedAt time.Time `json:"planted_at"`
	SessionID string    `json:"session_id"`
	Label     string    `json:"label"`       // e.g. "The Reed of Monday, 12th May"
	X         int       `json:"x"`           // horizontal placement percentage (0-100)
	Y         int       `json:"y"`           // vertical placement percentage (0-100)
	Stage     string    `json:"stage"`       // "sapling" (preview) | "mature" (planted)
}

type Session struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"` // focus, short-break, long-break
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	DurationMinutes int       `json:"duration_minutes"`
	Status          string    `json:"status"` // active, complete, abandoned
	OasisElementID  string    `json:"oasis_element_id,omitempty"`
}

type SettingsState struct {
	SoundEnabled        bool `json:"sound_enabled"`
	FocusDuration       int  `json:"focus_duration"`         // default 25
	ShortBreakDuration  int  `json:"short_break_duration"`   // default 5
	LongBreakDuration   int  `json:"long_break_duration"`    // default 15
	LongBreakInterval   int  `json:"long_break_interval"`    // default 4
	AutoStartBreaks     bool `json:"auto_start_breaks"`
	AutoStartFocus      bool `json:"auto_start_focus"`
	SunriseHour         int  `json:"sunrise_hour"`           // default 7
	SunsetHour          int  `json:"sunset_hour"`            // default 19
	TwinkleStarsEnabled bool `json:"twinkle_stars_enabled"`  // default true
}

type StatsState struct {
	CurrentStreak int            `json:"current_streak"`
	LongestStreak int            `json:"longest_streak"`
	DailyRecords  map[string]int `json:"daily_records"` // Date format YYYY-MM-DD -> minutes
}
```

---

## 4. Visual Rendering System

To replicate a premium visual experience in the terminal, we will build a dedicated Pomodoro widget with the following features:

### Large ASCII/Block Digits
- Remaining time is formatted (e.g., `25:00`) and rendered using a custom 5-line-high pixel block font.
- Each digit (0-9) and the colon separator (:) are mapped to a 2D string slice.
- Centered horizontally inside the viewport panel.

### Gradient Progress Bar
- The progress bar is a smooth horizontal gradient that transitions between two custom colors:
  - Start Color (Left): `#5E5CE6` (Indigo/Purple)
  - End Color (Right): `#C738D8` (Purple/Pink)
- For each character space in the filled bar, we interpolate the color:
  - Progress ratio $P = \text{index} / \text{width}$.
  - Character color = `InterpolateColor(startColor, endColor, P)`.
- The unfilled portion is rendered with a dark grey shaded block pattern (`░` or `▒`) to give a depth effect.
- The percentage (e.g., `60%`) is displayed as a clean text suffix.

---

## 5. Pure Logic Engines (Test-Driven Design)

To ensure the app is robust, all logic is isolated from terminal frameworks and fully covered by unit tests.

### A. Timer Engine (`internal/engine/timer.go`)
- Controls a state machine with inputs: `Start()`, `Pause()`, `Resume()`, `Tick(delta)`, `Stop()`.
- Calculates current elapsed time and outputs state transitions.
- **Tests (`timer_test.go`)**: Mock time increments and assert that transitions between focus and breaks happen correctly, and auto-start sequences trigger as configured.

### B. Stats Engine (`internal/engine/stats.go`)
- Tracks user streaks, longest streaks, and daily focus logs.
- Automatically calculates streak retention rules based on calendar day changes.

---

## 6. Testing & Quality Assurance (TDD)

```
             ┌────────────────────────┐
             │       Write Tests      │
             └───────────┬────────────┘
                         │
                         ▼
             ┌────────────────────────┐
             │   Run and Watch Fail   │
             └───────────┬────────────┘
                         │
                         ▼
             ┌────────────────────────┐
             │  Implement Logic/Code  │
             └───────────┬────────────┘
                         │
                         ▼
             ┌────────────────────────┐
             │      Verify Green      │
             └────────────────────────┘
```

## 7. Phased Implementation Roadmap

To maintain the highest quality and eliminate regressions, we divide the project into six discrete phases.

### Phase 1: Go Setup & Persistence Layer
- **Goal**: Establish project directory, Go module, and data persistence storage.
- **Tasks**:
  1. Initialize Go module.
  2. Implement database models and persistence rules in `internal/storage/json.go`.
- **Success Criteria**: `go test ./internal/storage/...` passes.

### Phase 2: Core State Engines (TDD)
- **Goal**: Implement timer state machines and statistics.
- **Tasks**:
  1. Create `internal/engine/timer.go` (Pomodoro timer state machine).
  2. Create `internal/engine/stats.go` (Streak tracking).
- **Success Criteria**: All unit tests pass with high coverage.

### Phase 3: CLI Shell Integration
- **Goal**: Initialize the Bubble Tea terminal process.
- **Tasks**:
  1. Write `main.go` to bootstrap the Bubble Tea application loop.
  2. Implement window size calculations (require at least 80x24 characters).
  3. Write basic key-binding handlers (Quit keys, tab navigation).
- **Success Criteria**: Running the application opens a stable terminal window with basic input controls.

### Phase 4: Centered Pomodoro Clock Widget
- **Goal**: Build the primary focus clock rendering system.
- **Tasks**:
  1. Implement custom 5-line-high ASCII block fonts for digits 0-9 and :.
  2. Implement horizontal gradient rendering for the progress bar (indigo-pink gradient).
  3. Wire the Bubble Tea view to center the large digital clock and render the progress bar and percentage.
  4. Implement keybindings: space (pause/resume), ← (reset), ↑ (+1 min), s (skip).
- **Success Criteria**: The main Oasis page displays a beautiful, responsive centered digital Pomodoro timer with an animated progress bar.

### Phase 5: Panel Integration (Settings & Stats Drawers)
- **Goal**: Connect stats and settings sub-pages.
- **Tasks**:
  1. Connect the real-time ticker clock to dispatch Bubble Tea ticks and decrement timer state.
  2. Create full-page views or side drawers for Settings (adjustable durations, auto-start options) and Stats (streak calendars, historic grid maps).
- **Success Criteria**: Full timer cycle operates automatically, updates statistics on completion, and settings adjustments successfully persist to the state JSON.

### Phase 6: Bells and Final Polish
- **Goal**: Finalize audio cues, edge cases, and standard linting.
- **Tasks**:
  1. Trigger terminal bell sounds (`\a`) on focus session completion.
  2. Package binary and complete validation.
- **Success Criteria**: TUI Pomodoro app is compile-ready, bug-free, and delivers a premium user experience.
