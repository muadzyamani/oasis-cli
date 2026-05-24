package components

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muadzyamani/oasis-cli/internal/engine"
	"github.com/muadzyamani/oasis-cli/internal/storage"
	"github.com/muesli/termenv"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripAnsi(str string) string {
	return ansiRegexp.ReplaceAllString(str, "")
}

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestInterpolateColor(t *testing.T) {
	c1 := RGBColor{0, 100, 200}
	c2 := RGBColor{10, 200, 100}

	res := InterpolateColor(c1, c2, 0.5)
	expected := RGBColor{5, 150, 150}

	if res != expected {
		t.Errorf("expected %v, got %v", expected, res)
	}
}

func TestRenderTimer(t *testing.T) {
	settings := storage.SettingsState{
		FocusDuration: 25,
	}
	// Initialize a timer from engine package
	timer := engine.NewTimer(settings, "", 0)
	timer.TimeRemaining = 25 * time.Minute // 25:00

	width, height := 80, 20
	rendered := RenderTimer(timer, width, height)

	// Since Lip Gloss renders ANSI codes, let's strip or check basic text content
	cleanStr := stripAnsi(rendered)

	// Verify that the status label matches "work session (ready)"
	if !strings.Contains(cleanStr, "work session (ready)") {
		t.Errorf("expected status 'work session (ready)', got: %q", cleanStr)
	}

	// Verify the percentage indicator is 0% on start
	if !strings.Contains(cleanStr, "0%") {
		t.Errorf("expected 0%% progress bar suffix, got: %q", cleanStr)
	}

	// Verify progress bar remains at 0% when time is modified in StateIdle
	timer.TimeRemaining = 20 * time.Minute
	renderedModifiedIdle := RenderTimer(timer, width, height)
	cleanModifiedIdle := stripAnsi(renderedModifiedIdle)
	if !strings.Contains(cleanModifiedIdle, "0%") {
		t.Errorf("expected 0%% progress bar suffix when idle after time modified, got: %q", cleanModifiedIdle)
	}

	// Make the session active and verify it says "running"
	timer.Start(time.Now())
	renderedActive := RenderTimer(timer, width, height)
	cleanActive := stripAnsi(renderedActive)
	if !strings.Contains(cleanActive, "work session (running)") {
		t.Errorf("expected status 'work session (running)', got: %q", cleanActive)
	}

	// Fast forward time and check progress bar is updated
	timer.TimeRemaining = 10 * time.Minute // 40% elapsed (15 mins out of 25)
	renderedProgress := RenderTimer(timer, width, height)
	cleanProgress := stripAnsi(renderedProgress)
	if !strings.Contains(cleanProgress, "60%") { // 60% elapsed
		t.Errorf("expected progress 60%%, got: %q", cleanProgress)
	}
}
