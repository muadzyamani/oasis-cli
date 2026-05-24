package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muadzyamani/oasis-cli/internal/engine"
)

// Update routes incoming Bubble Tea messages to mutate Model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "right", "l":
			m.ActiveTab = nextTab(m.ActiveTab)
			return m, nil

		case "shift+tab", "left", "h":
			m.ActiveTab = prevTab(m.ActiveTab)
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		// Generate astronomical stars based on terminal viewport size
		m.Stars = engine.GenerateStars(m.Width, m.Height, 30)
		return m, nil
	}

	return m, nil
}

func nextTab(current Tab) Tab {
	switch current {
	case TabOasis:
		return TabStats
	case TabStats:
		return TabSettings
	case TabSettings:
		return TabOasis
	default:
		return TabOasis
	}
}

func prevTab(current Tab) Tab {
	switch current {
	case TabOasis:
		return TabSettings
	case TabStats:
		return TabOasis
	case TabSettings:
		return TabStats
	default:
		return TabOasis
	}
}
