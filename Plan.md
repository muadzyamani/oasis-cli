# Oasis — Go TUI Architecture & Implementation Plan

This document details the architectural design, user interface layout, visual rendering systems, testing strategy, and execution roadmap for building a Go-based Terminal User Interface (TUI) version of the Oasis Pomodoro productivity application.

---

## 1. Architectural Overview

The application is structured around Charmbracelet's **Bubble Tea** (an Elm-architecture framework for Go) and **Lip Gloss** (a style definition library). To ensure maximum testability, maintainability, and clean separation of concerns, the project isolates pure mathematical and logical calculations into standard Go packages, keeping the terminal-rendering and key-binding loop as thin controllers.

### Component Layering & Z-Index Compositing

In standard web applications, overlapping dialogs (drawers, notifications) are handled via CSS absolute positioning and `z-index` layers. Since terminals are character grids, we must implement a custom **ASCII Layer Compositor**.

```
+-----------------------------------------------------------------+
| Left Nav Rail |                     OasisScene                  |
| [Stats]       |                                                 |
| [Settings]    |           +-----------------------+             |
|               |           |      Timer Widget     |             |
|               |           |         24:59         |             |
|               |           |     [Pause] [Stop]    |             |
|               |           +-----------------------+             |
|               |                                                 |
+-----------------------------------------------------------------+
```

- **Background Canvas**: The `OasisScene` renders a full-viewport text canvas comprising the sky gradient, astronomical bodies, twinkling stars, sand dunes, a ripple pool, and planted vegetation.
- **Overlay Panels**: The Timer Widget (centered float) and side drawers (Settings, Stats) are rendered as separate, bounded text rectangles.
- **Compositor Engine**: A utility function merges these text blocks by painting the top-layer characters over the background canvas at specific row/column offsets.

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
│   │   ├── growth.go            # Tier progressions, plant selectors (TDD)
│   │   ├── ambient.go           # Solar arcs, color-interpolation (TDD)
│   │   └── stats.go             # Streak calculations & daily metrics (TDD)
│   ├── storage/
│   │   └── json.go              # JSON local persistence engine (TDD)
│   └── ui/
│       ├── model.go             # Main Bubble Tea Model state
│       ├── update.go            # Bubble Tea Update event router
│       ├── view.go              # Bubble Tea View composer & Layout builder
│       ├── components/          # Styled Lip Gloss views
│       │   ├── scene.go         # Sky, dunes, pool, & plants renderer
│       │   ├── timer.go         # Floating timer widget
│       │   ├── sidebar.go       # Navigation panel
│       │   ├── settings.go      # Settings options list
│       │   └── stats.go         # Progress tracking & historic panels
│       ├── styles/
│       │   └── colors.go        # Design tokens & color functions
│       └── utils/
│           └── compositor.go    # Overlapping layout compositor utility
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

To replicate a premium visual experience in the terminal, we will build a dedicated, low-overhead ASCII rendering grid using true color (24-bit RGB) formatting via `lipgloss.Color`.

### The Sky Canvas & Gradient
- We define **14 color keyframes** across the day (similar to the web version: deep night, dawn lavender, solar noon yellow, sunset rose, twilight).
- The `ambient` engine interpolates these colors minute-by-minute.
- The Sky layer is generated row-by-row. We interpolate between `skyTopColor` and `skyBottomColor` to paint the background of each character space in that row.

### Sun, Moon, and Twinkling Stars
- **Sun (`☼`) / Moon (`☾`)**: Solar elevation $E$ follows a sinusoidal curve:
  $$E = \sin\left(\frac{t - t_{\text{sunrise}}}{t_{\text{sunset}} - t_{\text{sunrise}}} \times \pi\right)$$
  Based on $E$, we calculate a coordinate grid location $(X, Y)$ and stamp a custom multi-character glowing shape.
- **Stars**: Twinkling is simulated by generating $N$ star points seeded by coordinate hashes. A timer tick transitions the star character intensity through Lip Gloss color shades (e.g. bright white `█`, muted gray `░`, invisible) at different rates.

### The Sand Dunes
Dunes are rendered at the bottom section of the screen by plotting mathematical wave curves:
$$Y = A \sin(B \cdot X + C) + D$$
Each dune is colored in layered bands (representing ridges and shadow depth) using sand tones adjusted dynamically by the current time-of-day.

### Flora & Vegetation
Plants are stored as static multi-line ASCII shapes:
```go
var PalmSprite = []string{
	"  \\│/  ",
	" ─🌴─ ",
	"  /│\\  ",
	"   │   ",
}
```
During active focus sessions, a "ghost preview" is rendered by styling the plant with a fading opacity palette (interpolating between text color and background color) as progress approaches completion.

---

## 5. Pure Logic Engines (Test-Driven Design)

To ensure the app is robust, all logic is isolated from terminal frameworks and fully covered by unit tests.

### A. Timer Engine (`internal/engine/timer.go`)
- Controls a state machine with inputs: `Start()`, `Pause()`, `Resume()`, `Tick(delta)`, `Stop()`.
- Calculates current elapsed time and outputs state transitions.
- **Tests (`timer_test.go`)**: Mock time increments and assert that transitions between focus and breaks happen correctly, and auto-start sequences trigger as configured.

### B. Growth Engine (`internal/engine/growth.go`)
- Maintains plant thresholds and unlocks.
- Tracks round-robin cycles to choose between `palm`, `acacia`, `succulent`, and `willow`.
- Calculates random placement coordinates within designated zones, guaranteeing no overlapping collisions.
- **Tests (`growth_test.go`)**: Verify that completing focus sessions increments focus minutes, changes tiers at exact intervals (e.g. 25 min, 120 min), and returns valid coordinates.

### C. Ambient Engine (`internal/engine/ambient.go`)
- Calculates astronomical body orbits and color gradients.
- **Tests (`ambient_test.go`)**: Input specific times (12:00 PM, 6:00 AM, 12:00 AM) and verify the correct color output, solar/lunar coordinates, and moon phase calculations.

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

1. **Unit Testing Go Logic**: Every module inside `internal/engine/` and `internal/storage/` must achieve $\ge 90\%$ test coverage before writing TUI shell wrappers.
2. **Bubble Tea Update Testing**: We will test the core UI controller (`internal/ui/update.go`) using Go testing. We construct a `Model` and send it direct Bubble Tea messages (e.g. `tea.KeyMsg{Type: tea.KeySpace}`) and assert that states change as expected.
3. **Compositor Test**: Unit test the string compositing engine to verify that when text overlays are merged at specific offsets, borders and internal contents are overwritten correctly without distorting the layout.

---

## 7. Phased Implementation Roadmap

To maintain the highest quality and eliminate regressions, we divide the project into six discrete phases.

### Phase 1: Go Setup & Persistence Layer
- **Goal**: Establish project directory, Go module, and data persistence storage.
- **Tasks**:
  1. Initialize Go module.
  2. Implement database models and persistence rules in `internal/storage/json.go`.
  3. Write `json_test.go` to test database operations, atomic save safety, and directories creation.
- **Success Criteria**: `go test ./internal/storage/...` passes with $100\%$ success.

### Phase 2: Core State Engines (TDD)
- **Goal**: Implement all non-UI logical controllers.
- **Tasks**:
  1. Create `internal/engine/timer.go` and implement the Pomodoro timer state machine.
  2. Create `internal/engine/growth.go` and implement plant cycling, positioning calculations, and tier limits.
  3. Create `internal/engine/stats.go` and implement streak computation.
  4. Write comprehensive tests for all three engines.
- **Success Criteria**: All unit tests pass with high coverage.

### Phase 3: Ambient Engine & CLI Shell Integration
- **Goal**: Integrate time-of-day math and initialize the Bubble Tea terminal process.
- **Tasks**:
  1. Implement `internal/engine/ambient.go` for celestial curves and sky color mapping.
  2. Write `main.go` to bootstrap the Bubble Tea application loop.
  3. Implement window size calculations and show a resizing warning if the terminal is below 80x24 characters.
  4. Write basic key-binding handlers (Quit keys, tab navigation).
- **Success Criteria**: Running the application opens a stable terminal window with basic input controls.

### Phase 4: Layout Compositor & Canvas Rendering
- **Goal**: Build the ASCII rendering engine and visual representation.
- **Tasks**:
  1. Implement the character overlay compositor in `internal/ui/utils/compositor.go`.
  2. Implement sky, star, dune, and pool renderers.
  3. Create ASCII/Unicode representation of flora and vegetation.
  4. Implement a developer command-line flag (`-dev`) that maps specific keys (`+`/`-`) to scrub through time, and a key (`f`) to trigger instant focus completions to inspect rendering.
- **Success Criteria**: The terminal displays a beautiful animated desert landscape that reacts to time scrubbing.

### Phase 5: Panel Integration (Timer, Settings, Drawers)
- **Goal**: Connect interactive user controls and panels.
- **Tasks**:
  1. Create the Timer card component (idle expanded card, compact bottom bar when active).
  2. Connect the timer core clock to dispatch Bubble Tea ticks.
  3. Create side drawers for Settings (adjustable durations, resets) and Stats (streaks, historic grid maps).
  4. Integrate the compositor to slide/overlay drawers over the scene.
- **Success Criteria**: Timers can be started, paused, completed, and settings configurations successfully persist to local JSON files.

### Phase 6: Animation, Bells, and Final Polish
- **Goal**: Finalize animations, sound feedback, and packaging.
- **Tasks**:
  1. Implement star twinkling, particle drift, and pool ripple animations using sub-second frame ticks.
  2. Trigger terminal bell sounds (`\a`) on focus session completion.
  3. Conduct full code linting, format standardizations, and binary packaging.
- **Success Criteria**: Product is compile-ready, bug-free, and delivers a highly polished TUI experience.
