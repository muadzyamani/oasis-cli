package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

// ArabicDigits defines 5-line high retro-style Eastern Arabic numerals.
var ArabicDigits = map[rune][]string{
	'0': {
		"   ▄▄   ",
		" ▄████▄ ",
		"████████",
		" ▀████▀ ",
		"   ▀▀   ",
	},
	'1': {
		"  ▄██▄  ",
		"   ██   ",
		"   ██   ",
		"   ██   ",
		" ▄████▄ ",
	},
	'2': {
		"████████",
		"▀▀▀▀▀██ ",
		"    ██  ",
		"   ██   ",
		"  ████  ",
	},
	'3': {
		"██ █ ███",
		"██ █ ███",
		"▀███████",
		"     ██ ",
		"    ██  ",
	},
	'4': {
		" ▄████▄ ",
		"  ▀▀▀███",
		" ▄████▀ ",
		"  ▀▀▀███",
		" ▀████▀ ",
	},
	'5': {
		"   ▄█   ",
		" ▄████▄ ",
		"██▀  ▀██",
		"██▄  ▄██",
		" ▀████▀ ",
	},
	'6': {
		"████████",
		" ██▀▀▀▀▀",
		"  ██    ",
		"   ██   ",
		"    ██  ",
	},
	'7': {
		"██▀  ▀██",
		"██    ██",
		"▀██  ██▀",
		" ▀████▀ ",
		"   ▀▀   ",
	},
	'8': {
		"   ▄▄   ",
		" ▄████▄ ",
		"▄██  ██▄",
		"██    ██",
		"██    ██",
	},
	'9': {
		" ▄████▄ ",
		"██▀  ▀██",
		"▀███████",
		"     ██ ",
		"     ██ ",
	},
	':': {
		"   ",
		" ▄ ",
		"   ",
		" ▄ ",
		"   ",
	},
}

type RGBColor struct {
	R, G, B int
}

func InterpolateColor(c1, c2 RGBColor, t float64) RGBColor {
	if t <= 0.0 {
		return c1
	}
	if t >= 1.0 {
		return c2
	}
	return RGBColor{
		R: int(math.Round(float64(c1.R) + t*float64(c2.R-c1.R))),
		G: int(math.Round(float64(c1.G) + t*float64(c2.G-c1.G))),
		B: int(math.Round(float64(c1.B) + t*float64(c2.B-c1.B))),
	}
}

func formatHexColor(c RGBColor) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// 5-line high arcade-style block digits (8 columns wide each, colon is 3 columns wide)
var BlockDigits = map[rune][]string{
	'0': {
		" ▄████▄ ",
		"██▀  ▀██",
		"██    ██",
		"██▄  ▄██",
		" ▀████▀ ",
	},
	'1': {
		"  ▄██   ",
		" ▄▀██   ",
		"   ██   ",
		"   ██   ",
		" ▄████▄ ",
	},
	'2': {
		" ▄████▄ ",
		"▀▀   ██ ",
		"   ▄██▀ ",
		" ▄██▀   ",
		"▀██████▀",
	},
	'3': {
		" ▄████▄ ",
		"      ██",
		"  ▀████▄",
		"      ██",
		" ▀████▀ ",
	},
	'4': {
	"▄█   █▄ ",
	"██   ██ ",
	"▀██████▀",
	"     ██ ",
	"     ▀█ ",
	},
	'5': {
		" ▄████▄ ",
		"██▀▀▀▀  ",
		"▀██████▄",
		"     ▀██",
		" ▀████▀ ",
	},
	'6': {
		" ▄████▄ ",
		"██▀▀▀▀▀ ",
		"███████▄",
		"██    ██",
		" ▀████▀ ",
	},
	'7': {
		"▄██████▄",
		"▀▀   ██▀",
		"    ██▀ ",
		"   ██▀  ",
		"  ██▀   ",
	},
	'8': {
		" ▄████▄ ",
		"██▄  ▄██",
		" ▀████▀ ",
		"██▀  ▀██",
		" ▀████▀ ",
	},
	'9': {
		" ▄████▄ ",
		"██▀  ▀██",
		" ▀██████",
		"     ▄██",
		"  ▄███▀ ",
	},
	':': {
		"   ",
		" ▄ ",
		"   ",
		" ▄ ",
		"   ",
	},
}

// RenderTimer renders the centered digital Pomodoro timer inside the viewport bounds.
func RenderTimer(t *Timer, width, height int) string {
	// 1. Format Time (MM:SS)
	minutes := int(t.TimeRemaining.Minutes())
	seconds := int(t.TimeRemaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	// 2. Build Block Digit ASCII representation
	var clockLines [5]string
	for r := 0; r < 5; r++ {
		var rowParts []string
		for _, char := range timeStr {
			var block []string
			var exists bool
			if t.Settings.UseArabicNumerals {
				block, exists = ArabicDigits[char]
			} else {
				block, exists = BlockDigits[char]
			}
			if exists {
				rowParts = append(rowParts, block[r])
			} else {
				rowParts = append(rowParts, "        ")
			}
		}
		clockLines[r] = strings.Join(rowParts, "  ") // spacing between digits
	}

	// Horizontal gradient colors for the clock block: from #5D5FEF (indigo) to #BB86FC (lavender)
	clockStartColor := RGBColor{R: 93, G: 95, B: 239}   // #5D5FEF
	clockEndColor := RGBColor{R: 187, G: 134, B: 252}   // #BB86FC

	var renderedClockLines []string
	for _, line := range clockLines {
		runes := []rune(line)
		var lineBuilder strings.Builder
		numChars := len(runes)
		for col, char := range runes {
			if char == ' ' {
				lineBuilder.WriteRune(' ')
			} else {
				ratio := 0.0
				if numChars > 1 {
					ratio = float64(col) / float64(numChars-1)
				}
				interpolated := InterpolateColor(clockStartColor, clockEndColor, ratio)
				colorHex := formatHexColor(interpolated)
				charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorHex)).Bold(true)
				lineBuilder.WriteString(charStyle.Render(string(char)))
			}
		}
		renderedClockLines = append(renderedClockLines, lineBuilder.String())
	}
	clockBlock := strings.Join(renderedClockLines, "\n")

	// 3. Render Status Label
	sessionLabel := "work session"
	switch t.SessionType {
	case "short-break":
		sessionLabel = "short break"
	case "long-break":
		sessionLabel = "long break"
	}

	stateLabel := "ready"
	switch t.State {
	case StateRunning:
		stateLabel = "running"
	case StatePaused:
		stateLabel = "paused"
	case StateCompleted:
		stateLabel = "completed"
	}

	statusText := fmt.Sprintf("%s (%s)", sessionLabel, stateLabel)
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7F8C8D")).
		Italic(true)
	statusBlock := statusStyle.Render(statusText)

	// 4. Render Gradient Progress Bar
	barWidth := width - 12
	if barWidth < 20 {
		barWidth = 20
	}
	if barWidth > 60 {
		barWidth = 60
	}

	progress := 0.0
	if t.State != StateIdle {
		if t.Duration > 0 {
			progress = 1.0 - float64(t.TimeRemaining)/float64(t.Duration)
		}
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
	}

	filledCount := int(math.Round(progress * float64(barWidth)))
	if filledCount < 0 {
		filledCount = 0
	}
	if filledCount > barWidth {
		filledCount = barWidth
	}

	// Indigo to Purple-Pink gradient keyframes
	startColor := RGBColor{R: 94, G: 92, B: 230}   // #5E5CE6
	endColor := RGBColor{R: 199, G: 56, B: 216}   // #C738D8

	// Progress bar style configuration
	style := t.Settings.ProgressBarStyle
	if style == "" {
		style = "solid-capsule"
	}

	leftCap := "◖"
	rightCap := "◗"
	progressChar := "█"
	remainingChar := "░"
	useCapsules := true

	switch style {
	case "beaded-capsule":
		leftCap = "("
		rightCap = ")"
		progressChar = "●"
		remainingChar = "○"
	case "framed-rounded":
		useCapsules = false
	default: // "solid-capsule" and fallback
		leftCap = "◖"
		rightCap = "◗"
		progressChar = "█"
		remainingChar = "░"
	}

	var progressBlock string
	unfilledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2E303E"))
	percentageText := fmt.Sprintf("  %3d%%", int(progress*100))
	percentageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8C8D"))

	if useCapsules {
		var barBuilder strings.Builder
		for i := 0; i < barWidth; i++ {
			var char string
			if i == 0 {
				char = leftCap
			} else if i == barWidth-1 {
				char = rightCap
			} else {
				if i < filledCount {
					char = progressChar
				} else {
					char = remainingChar
				}
			}

			// Render with gradient if filled, otherwise with unfilledStyle
			isFilled := false
			if i == 0 {
				isFilled = filledCount > 0
			} else if i == barWidth-1 {
				isFilled = filledCount == barWidth
			} else {
				isFilled = i < filledCount
			}

			if isFilled {
				ratio := float64(i) / float64(barWidth-1) // smooth gradient scale from 0 to 1
				interpolated := InterpolateColor(startColor, endColor, ratio)
				colorHex := formatHexColor(interpolated)
				charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorHex))
				barBuilder.WriteString(charStyle.Render(char))
			} else {
				if char == " " {
					barBuilder.WriteString(" ")
				} else {
					barBuilder.WriteString(unfilledStyle.Render(char))
				}
			}
		}

		progressBlock = lipgloss.JoinHorizontal(
			lipgloss.Center,
			barBuilder.String(),
			percentageStyle.Render(percentageText),
		)
	} else {
		// Framed rounded style: 3 lines high
		innerBarWidth := barWidth - 2
		if innerBarWidth < 10 {
			innerBarWidth = 10
		}

		// Recalculate filled count for inner bar width
		innerFilledCount := int(math.Round(progress * float64(innerBarWidth)))
		if innerFilledCount < 0 {
			innerFilledCount = 0
		}
		if innerFilledCount > innerBarWidth {
			innerFilledCount = innerBarWidth
		}

		var innerBarBuilder strings.Builder
		for i := 0; i < innerBarWidth; i++ {
			if i < innerFilledCount {
				ratio := float64(i) / float64(innerBarWidth-1)
				interpolated := InterpolateColor(startColor, endColor, ratio)
				colorHex := formatHexColor(interpolated)
				charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorHex))
				innerBarBuilder.WriteString(charStyle.Render("█"))
			} else {
				innerBarBuilder.WriteString(" ")
			}
		}

		topLine := unfilledStyle.Render("╭" + strings.Repeat("─", innerBarWidth) + "╮")
		middleLine := unfilledStyle.Render("│") + innerBarBuilder.String() + unfilledStyle.Render("│")
		bottomLine := unfilledStyle.Render("╰" + strings.Repeat("─", innerBarWidth) + "╯")

		middleLineWithPercent := lipgloss.JoinHorizontal(
			lipgloss.Center,
			middleLine,
			percentageStyle.Render(percentageText),
		)

		progressBlock = lipgloss.JoinVertical(
			lipgloss.Left,
			topLine,
			middleLineWithPercent,
			bottomLine,
		)
	}

	// 5. Render Legend Guides
	var content string
	if !t.Settings.HideTimerControls {
		legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4B59"))
		var legendText string
		if t.Settings.DevMode {
			legendText = "↑ +1 minute  •  space pause/resume  •  ← reset  •  s skip  •  e toggle"
		} else {
			legendText = "↑ +1 minute  •  space pause/resume  •  ← reset  •  e toggle"
		}
		legendBlock := legendStyle.Render(legendText)

		content = lipgloss.JoinVertical(
			lipgloss.Center,
			clockBlock,
			"",
			statusBlock,
			"",
			progressBlock,
			"",
			"",
			legendBlock,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Center,
			clockBlock,
			"",
			statusBlock,
			"",
			progressBlock,
		)
	}

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

