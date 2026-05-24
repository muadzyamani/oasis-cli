package engine

import (
	"math"
	"math/rand"
	"time"

	"github.com/muadzyamani/oasis-cli/internal/storage"
)

var PlantTypes = []string{"palm", "acacia", "succulent", "willow"}

// GetTierForMinutes maps total focus minutes to Oasis tier (0 to 5).
func GetTierForMinutes(minutes int) int {
	if minutes < 25 {
		return 0
	}
	if minutes < 120 {
		return 1
	}
	if minutes < 300 {
		return 2
	}
	if minutes < 600 {
		return 3
	}
	if minutes < 1200 {
		return 4
	}
	return 5
}

// GetNextPlantType cycles through available plant types in a round-robin sequence based on existing elements.
func GetNextPlantType(elements []storage.OasisElement) string {
	return PlantTypes[len(elements)%len(PlantTypes)]
}

// CalculatePlantPosition generates a coordinate (X: 5-95, Y: 65-85) that does not collide with existing elements.
func CalculatePlantPosition(rnd *rand.Rand, existingElements []storage.OasisElement, minDistance float64) (int, int) {
	xMin, xMax := 5, 95
	yMin, yMax := 65, 85

	if len(existingElements) == 0 {
		x := rnd.Intn(xMax-xMin+1) + xMin
		y := rnd.Intn(yMax-yMin+1) + yMin
		return x, y
	}

	bestX, bestY := 0, 0
	maxAttempts := 200
	currentMinDist := minDistance

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 100 {
			currentMinDist = minDistance * 0.5
		}
		if attempt > 150 {
			currentMinDist = minDistance * 0.25
		}
		if attempt > 180 {
			currentMinDist = 0.0
		}

		x := rnd.Intn(xMax-xMin+1) + xMin
		y := rnd.Intn(yMax-yMin+1) + yMin

		tooClose := false
		for _, el := range existingElements {
			dx := float64(x - el.X)
			dy := float64(y - el.Y)
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < currentMinDist {
				tooClose = true
				break
			}
		}

		if !tooClose {
			return x, y
		}
		bestX, bestY = x, y
	}

	return bestX, bestY
}

// PlantPreview places a new plant in the "sapling" stage for the current active focus session.
func PlantPreview(rnd *rand.Rand, state *storage.OasisState, sessionID string, label string, now time.Time) *storage.OasisElement {
	if sessionID == "" {
		return nil
	}

	// Calculate collision-free position
	x, y := CalculatePlantPosition(rnd, state.Elements, 10.0)

	newElement := storage.OasisElement{
		ID:        GenerateID(),
		Type:      GetNextPlantType(state.Elements),
		PlantedAt: now,
		SessionID: sessionID,
		Label:     label,
		X:         x,
		Y:         y,
		Stage:     "sapling",
	}

	state.Elements = append(state.Elements, newElement)
	return &newElement
}

// CompletePlant transitions the plant corresponding to the given sessionID from "sapling" to "mature".
// It also updates total focus minutes and recalculates the Oasis tier.
func CompletePlant(state *storage.OasisState, sessionID string, focusMinutes int) bool {
	found := false
	for i := range state.Elements {
		if state.Elements[i].SessionID == sessionID && state.Elements[i].Stage == "sapling" {
			state.Elements[i].Stage = "mature"
			found = true
			break
		}
	}

	if found {
		state.TotalFocusMinutes += focusMinutes
		state.Tier = GetTierForMinutes(state.TotalFocusMinutes)
	}

	return found
}

// AbandonPlant removes the sapling preview element associated with the given sessionID.
func AbandonPlant(state *storage.OasisState, sessionID string) bool {
	for i, el := range state.Elements {
		if el.SessionID == sessionID && el.Stage == "sapling" {
			// Remove from slice
			state.Elements = append(state.Elements[:i], state.Elements[i+1:]...)
			return true
		}
	}
	return false
}
