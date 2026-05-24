package engine

import (
	"math/rand"
	"testing"
	"time"

	"github.com/muadzyamani/oasis-cli/internal/storage"
)

func TestGetTierForMinutes(t *testing.T) {
	tests := []struct {
		minutes  int
		expected int
	}{
		{0, 0},
		{24, 0},
		{25, 1},
		{119, 1},
		{120, 2},
		{299, 2},
		{300, 3},
		{599, 3},
		{600, 4},
		{1199, 4},
		{1200, 5},
		{5000, 5},
	}

	for _, tt := range tests {
		actual := GetTierForMinutes(tt.minutes)
		if actual != tt.expected {
			t.Errorf("GetTierForMinutes(%d): expected %d, got %d", tt.minutes, tt.expected, actual)
		}
	}
}

func TestGetNextPlantType(t *testing.T) {
	elements := []storage.OasisElement{}
	if pType := GetNextPlantType(elements); pType != "palm" {
		t.Errorf("expected palm, got %s", pType)
	}

	elements = append(elements, storage.OasisElement{Type: "palm"})
	if pType := GetNextPlantType(elements); pType != "acacia" {
		t.Errorf("expected acacia, got %s", pType)
	}

	elements = append(elements, storage.OasisElement{Type: "acacia"})
	if pType := GetNextPlantType(elements); pType != "succulent" {
		t.Errorf("expected succulent, got %s", pType)
	}

	elements = append(elements, storage.OasisElement{Type: "succulent"})
	if pType := GetNextPlantType(elements); pType != "willow" {
		t.Errorf("expected willow, got %s", pType)
	}

	elements = append(elements, storage.OasisElement{Type: "willow"})
	if pType := GetNextPlantType(elements); pType != "palm" {
		t.Errorf("expected palm, got %s", pType)
	}
}

func TestCalculatePlantPosition(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))

	// Test bounds on empty list
	for i := 0; i < 50; i++ {
		x, y := CalculatePlantPosition(rnd, nil, 10.0)
		if x < 5 || x > 95 {
			t.Errorf("X coordinate %d out of bounds [5, 95]", x)
		}
		if y < 65 || y > 85 {
			t.Errorf("Y coordinate %d out of bounds [65, 85]", y)
		}
	}

	// Test collision prevention
	existing := []storage.OasisElement{
		{X: 50, Y: 75},
	}

	for i := 0; i < 100; i++ {
		x, y := CalculatePlantPosition(rnd, existing, 15.0)
		// Check that distance is indeed >= 15
		dx := float64(x - 50)
		dy := float64(y - 75)
		dist := (dx * dx) + (dy * dy)
		if dist < 15.0*15.0 {
			t.Errorf("Overlap detected! Candidate (%d, %d) is too close to (50, 75). Distance square: %f", x, y, dist)
		}
	}
}

func TestPlantPreviewCompleteAbandon(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	state := &storage.OasisState{
		Name:              "Test Oasis",
		Tier:              0,
		TotalFocusMinutes: 0,
		Elements:          []storage.OasisElement{},
		CreatedAt:         time.Now(),
	}

	now := time.Now()
	sessionID := "sess-123"

	// 1. Plant Preview (Sapling)
	preview := PlantPreview(rnd, state, sessionID, "Reed planted", now)
	if preview == nil {
		t.Fatal("expected preview element to be created")
	}
	if len(state.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(state.Elements))
	}
	if state.Elements[0].Stage != "sapling" {
		t.Errorf("expected stage sapling, got %s", state.Elements[0].Stage)
	}
	if state.Elements[0].SessionID != sessionID {
		t.Errorf("expected session id %s, got %s", sessionID, state.Elements[0].SessionID)
	}
	if state.Elements[0].Type != "palm" {
		t.Errorf("expected first plant type palm, got %s", state.Elements[0].Type)
	}

	// 2. Complete Plant
	ok := CompletePlant(state, sessionID, 25)
	if !ok {
		t.Fatal("expected CompletePlant to return true")
	}
	if state.Elements[0].Stage != "mature" {
		t.Errorf("expected stage to transition to mature, got %s", state.Elements[0].Stage)
	}
	if state.TotalFocusMinutes != 25 {
		t.Errorf("expected 25 focus minutes, got %d", state.TotalFocusMinutes)
	}
	if state.Tier != 1 {
		t.Errorf("expected tier 1, got %d", state.Tier)
	}

	// 3. Plant another and then Abandon it
	sessionID2 := "sess-456"
	preview2 := PlantPreview(rnd, state, sessionID2, "Sprout planted", now)
	if preview2 == nil {
		t.Fatal("expected preview 2 to be created")
	}
	if len(state.Elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(state.Elements))
	}
	if state.Elements[1].Type != "acacia" {
		t.Errorf("expected second plant to cycle to acacia, got %s", state.Elements[1].Type)
	}

	// Abandon the second plant
	abandonOk := AbandonPlant(state, sessionID2)
	if !abandonOk {
		t.Fatal("expected AbandonPlant to return true")
	}
	if len(state.Elements) != 1 {
		t.Errorf("expected element count to revert to 1, got %d", len(state.Elements))
	}
	if state.Elements[0].SessionID != sessionID {
		t.Errorf("expected remaining element to be the first one, got session %s", state.Elements[0].SessionID)
	}
}

func TestGrowthInvalidActions(t *testing.T) {
	state := &storage.OasisState{
		Elements: []storage.OasisElement{},
	}
	rnd := rand.New(rand.NewSource(1))

	// PlantPreview with empty session ID
	if p := PlantPreview(rnd, state, "", "label", time.Now()); p != nil {
		t.Error("expected PlantPreview to return nil with empty session ID")
	}

	// AbandonPlant with non-existent session ID
	if ok := AbandonPlant(state, "non-existent"); ok {
		t.Error("expected AbandonPlant to return false with non-existent session ID")
	}
}

