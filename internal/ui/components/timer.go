package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muadzyamani/oasis-cli/internal/engine"
)

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

// // 5-line high block digits (7 columns wide each, colon is 3 columns wide)
// var BlockDigits = map[rune][]string{
// 	'0': {
// 		" ▄███▄ ",
// 		"██▀ ▀██",
// 		"██ █ ██",
// 		"██▄ ▄██",
// 		" ▀███▀ ",
// 	},
// 	'1': {
// 		"  ▄██  ",
// 		" ▀ ██  ",
// 		"   ██  ",
// 		"   ██  ",
// 		" ▄████▄",
// 	},
// 	'2': {
// 		" ▄███▄ ",
// 		"▀▀  ▀██",
// 		"   ▄██▀",
// 		" ▄██▀  ",
// 		"███████",
// 	},
// 	'3': {
// 		" █████▄",
// 		"    ▀██",
// 		" ▀████▄",
// 		"    ▀██",
// 		" █████▀",
// 	},
// 	'4': {
// 		"██  ██ ",
// 		"██  ██ ",
// 		"███████",
// 		"    ██ ",
// 		"    ██ ",
// 	},
// 	'5': {
// 		"███████",
// 		"██▀▀▀▀▀",
// 		"██████▄",
// 		"    ▀██",
// 		"██████▀",
// 	},
// 	'6': {
// 		" ▄███▄ ",
// 		"██▀▀▀▀ ",
// 		"██████▄",
// 		"██   ██",
// 		" ▀███▀ ",
// 	},
// 	'7': {
// 		"███████",
// 		"    ▄██",
// 		"   ▄██▀",
// 		"  ▄██▀ ",
// 		" ▄██▀  ",
// 	},
// 	'8': {
// 		" ▄███▄ ",
// 		"██▄ ▄██",
// 		" ▀███▀ ",
// 		"██▀ ▀██",
// 		" ▀███▀ ",
// 	},
// 	'9': {
// 		" ▄███▄ ",
// 		"██  ▄██",
// 		" ▀█████",
// 		"    ▄██",
// 		" ▀███▀ ",
// 	},
// 	':': {
// 		"   ",
// 		" ▄ ",
// 		"   ",
// 		" ▄ ",
// 		"   ",
// 	},
// }

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
		"████████",
	},
	'3': {
		" ▄████▄ ",
		"      ██",
		"  ▀████▄",
		"      ██",
		" ▀████▀ ",
	},
	'4': {
		"██   ██ ",
		"██   ██ ",
		"████████",
		"     ██ ",
		"     ██ ",
	},
	'5': {
		"████████",
		"██▀▀▀▀▀▀",
		"███████▄",
		"     ▀██",
		"███████▀",
	},
	'6': {
		" ▄████▄ ",
		"██▀▀▀▀▀ ",
		"███████▄",
		"██    ██",
		" ▀████▀ ",
	},
	'7': {
		"████████",
		"     ▄██",
		"    ▄██▀",
		"   ▄██▀ ",
		"  ▄██▀  ",
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
func RenderTimer(t *engine.Timer, width, height int) string {
	// 1. Format Time (MM:SS)
	minutes := int(t.TimeRemaining.Minutes())
	seconds := int(t.TimeRemaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	// 2. Build Block Digit ASCII representation
	var clockLines [5]string
	for r := 0; r < 5; r++ {
		var rowParts []string
		for _, char := range timeStr {
			block, exists := BlockDigits[char]
			if exists {
				rowParts = append(rowParts, block[r])
			} else {
				rowParts = append(rowParts, "        ")
			}
		}
		clockLines[r] = strings.Join(rowParts, "  ") // spacing between digits
	}

	clockColor := lipgloss.Color("#5D5FEF") // Premium Indigo/Violet
	clockStyle := lipgloss.NewStyle().Foreground(clockColor).Bold(true)

	var renderedClockLines []string
	for _, line := range clockLines {
		renderedClockLines = append(renderedClockLines, clockStyle.Render(line))
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
	case engine.StateRunning:
		stateLabel = "running"
	case engine.StatePaused:
		stateLabel = "paused"
	case engine.StateCompleted:
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
	if t.State != engine.StateIdle {
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

	var barBuilder strings.Builder
	for i := 0; i < filledCount; i++ {
		ratio := float64(i) / float64(barWidth)
		interpolated := InterpolateColor(startColor, endColor, ratio)
		colorHex := formatHexColor(interpolated)
		charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorHex))
		barBuilder.WriteString(charStyle.Render("█"))
	}

	unfilledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2E303E"))
	for i := filledCount; i < barWidth; i++ {
		barBuilder.WriteString(unfilledStyle.Render("░"))
	}

	percentageText := fmt.Sprintf("  %3d%%", int(progress*100))
	percentageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8C8D"))
	progressBlock := lipgloss.JoinHorizontal(
		lipgloss.Center,
		barBuilder.String(),
		percentageStyle.Render(percentageText),
	)

	// 5. Render Legend Guides
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4B59"))
	legendBlock := legendStyle.Render("↑ +1 minute  •  space pause/resume  •  ← reset  •  s skip")

	// Assemble page vertically and center within available space
	content := lipgloss.JoinVertical(
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

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
