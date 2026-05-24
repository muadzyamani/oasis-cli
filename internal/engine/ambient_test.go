package engine

import (
	"testing"
	"time"
)

func TestInterpolateColor(t *testing.T) {
	c1 := RGBColor{0, 100, 200}
	c2 := RGBColor{10, 150, 100}

	// 0.0 limit
	res := InterpolateColor(c1, c2, 0.0)
	if res != c1 {
		t.Errorf("expected %v, got %v", c1, res)
	}

	// 1.0 limit
	res = InterpolateColor(c1, c2, 1.0)
	if res != c2 {
		t.Errorf("expected %v, got %v", c2, res)
	}

	// Midpoint
	res = InterpolateColor(c1, c2, 0.5)
	expected := RGBColor{5, 125, 150}
	if res != expected {
		t.Errorf("expected %v, got %v", expected, res)
	}
}

func TestGetMoonPhase(t *testing.T) {
	// Let's test standard known dates
	// Jan 6, 2000 was a New Moon
	newMoonDate := time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)
	phase, name := GetMoonPhase(newMoonDate)
	if phase > 0.01 && phase < 0.99 {
		t.Errorf("expected phase near 0 or 1 for new moon date, got %f", phase)
	}
	if name != "New Moon" {
		t.Errorf("expected 'New Moon', got '%s'", name)
	}
}

func TestGetAmbientState(t *testing.T) {
	// Test Solar Noon (12:00 PM)
	noonTime := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	state := GetAmbientState(noonTime, 7, 19)

	if !state.IsDaytime {
		t.Error("expected IsDaytime = true at noon")
	}
	if state.SunX == -1 || state.SunY == -1 {
		t.Errorf("expected valid Sun position at noon, got (%d, %d)", state.SunX, state.SunY)
	}
	if state.MoonX != -1 || state.MoonY != -1 {
		t.Errorf("expected Moon position to be out of bounds at noon, got (%d, %d)", state.MoonX, state.MoonY)
	}

	// Test Solar Noon Top and Bottom colors
	if state.SkyTop != (RGBColor{26, 41, 128}) {
		t.Errorf("expected Noon Top color {26, 41, 128}, got %v", state.SkyTop)
	}

	// Test Midnight (12:00 AM)
	midnightTime := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	state = GetAmbientState(midnightTime, 7, 19)

	if state.IsDaytime {
		t.Error("expected IsDaytime = false at midnight")
	}
	if state.MoonX == -1 || state.MoonY == -1 {
		t.Errorf("expected valid Moon position at midnight, got (%d, %d)", state.MoonX, state.MoonY)
	}
	if state.SunX != -1 || state.SunY != -1 {
		t.Errorf("expected Sun position to be out of bounds at midnight, got (%d, %d)", state.SunX, state.SunY)
	}
	if state.SkyTop != (RGBColor{11, 12, 16}) {
		t.Errorf("expected Midnight Top color {11, 12, 16}, got %v", state.SkyTop)
	}
}

func TestGenerateStars(t *testing.T) {
	stars := GenerateStars(100, 30, 20)
	if len(stars) != 20 {
		t.Errorf("expected 20 stars, got %d", len(stars))
	}

	for _, s := range stars {
		if s.X < 5 || s.X > 95 {
			t.Errorf("star X coordinate %d out of bounds", s.X)
		}
		if s.Y < 2 || s.Y > 12 {
			t.Errorf("star Y coordinate %d out of bounds", s.Y)
		}
	}

	// Tiny screen height -> no stars
	emptyStars := GenerateStars(100, 10, 20)
	if len(emptyStars) != 0 {
		t.Errorf("expected 0 stars for tiny height, got %d", len(emptyStars))
	}
}
