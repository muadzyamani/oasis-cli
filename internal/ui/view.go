package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muadzyamani/oasis-cli/internal/ui/components"
	"github.com/muadzyamani/oasis-cli/internal/ui/styles"
)

// View renders the CLI user interface.
func (m Model) View() string {
	if !m.Ready {
		return "Initializing Oasis CLI..."
	}

	// 1. Terminal Size Check (Require at least 80x24)
	if m.Width < 80 || m.Height < 24 {
		warningMsg := fmt.Sprintf(
			"%s\n\nViewport: %d x %d\nRequired: 80 x 24 minimum\n\nPlease enlarge your terminal window.",
			styles.WarningTitle.Render("⚠️  TERMINAL WINDOW TOO SMALL"),
			m.Width,
			m.Height,
		)
		warningBox := styles.WarningBox.Render(warningMsg)
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, warningBox)
	}

	// 2. Render Navigation & Streak Header Row
	tabs := []string{}
	for _, t := range []Tab{TabOasis, TabStats, TabSettings} {
		tabLabel := strings.ToUpper(string(t))
		if m.ActiveTab == t {
			tabs = append(tabs, styles.ActiveTabStyle.Render(fmt.Sprintf("● %s", tabLabel)))
		} else {
			tabs = append(tabs, styles.InactiveTabStyle.Render(fmt.Sprintf("○ %s", tabLabel)))
		}
	}
	navTabs := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	streakUnit := "days"
	if m.State.Stats.CurrentStreak == 1 {
		streakUnit = "day"
	}
	streakRight := fmt.Sprintf(
		"Streak: %s",
		styles.StreakStyle.Render(fmt.Sprintf("%d %s", m.State.Stats.CurrentStreak, streakUnit)),
	)
	spaceWidth := m.Width - lipgloss.Width(navTabs) - lipgloss.Width(streakRight) - 2
	if spaceWidth < 0 {
		spaceWidth = 0
	}
	topBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		navTabs,
		strings.Repeat(" ", spaceWidth),
		streakRight,
	)
	header := styles.HeaderContainer.Render(topBar)

	// Dynamic height computation for layout components
	viewportHeight := m.Height - lipgloss.Height(header) - 5
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	viewportWidth := m.Width - 6
	if viewportWidth < 10 {
		viewportWidth = 10
	}

	// 4. Render Active Viewport Page Content
	var pageContent string
	switch m.ActiveTab {
	case TabOasis:
		innerWidth := viewportWidth - 6
		innerHeight := viewportHeight - 4
		if innerWidth < 10 {
			innerWidth = 10
		}
		if innerHeight < 3 {
			innerHeight = 3
		}
		pageContent = components.RenderTimer(m.Timer, innerWidth, innerHeight)
	case TabStats:
		pageContent = "📊  STATS & TRACKING PANEL\n\n(Streak calendars, historical focus charts, and metrics coming in Phase 5)"
	case TabSettings:
		pageContent = "⚙️  SETTINGS OPTIONS PANEL\n\n(Focus durations, sound toggles, and celestial cycles coming in Phase 5)"
	}

	viewportBorder := styles.ViewportContainer.
		Width(viewportWidth).
		Height(viewportHeight).
		Render(pageContent)

	// 5. Render Footer
	footerText := styles.FooterStyle.Render("Tab / Shift+Tab: Navigate  •  Q / Ctrl+C: Quit")
	footer := lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, footerText)

	// Join entire viewport screen vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportBorder,
		footer,
	)
}
