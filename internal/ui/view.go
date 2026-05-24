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

	var topBar string
	if m.State.Settings.ShowStreak {
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
		topBar = lipgloss.JoinHorizontal(
			lipgloss.Top,
			navTabs,
			strings.Repeat(" ", spaceWidth),
			streakRight,
		)
	} else {
		topBar = lipgloss.JoinHorizontal(
			lipgloss.Top,
			navTabs,
		)
	}
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
		// Render settings options
		var settingsRows []string

		// Options definitions
		opts := []struct {
			name  string
			value string
		}{
			{"Focus Duration", fmt.Sprintf("[ %d min ]", m.State.Settings.FocusDuration)},
			{"Break Duration", fmt.Sprintf("[ %d min ]", m.State.Settings.ShortBreakDuration)},
			{"Long Break Duration", fmt.Sprintf("[ %d min ]", m.State.Settings.LongBreakDuration)},
			{"Sound Toggle", func() string {
				if m.State.Settings.SoundEnabled {
					return "[ Enabled ]"
				}
				return "[ Disabled ]"
			}()},
			{"View Streak", func() string {
				if m.State.Settings.ShowStreak {
					return "[ Show ]"
				}
				return "[ Hide ]"
			}()},
		}

		for i, opt := range opts {
			var row string
			if m.SettingsCursor == i {
				itemStr := styles.SettingsItemActive.Render(fmt.Sprintf("%-22s", opt.name))
				valStr := styles.SettingsValueActive.Render(opt.value)
				row = fmt.Sprintf("> %s  %s", itemStr, valStr)
			} else {
				itemStr := styles.SettingsItemInactive.Render(fmt.Sprintf("%-22s", opt.name))
				valStr := styles.SettingsValueInactive.Render(opt.value)
				row = fmt.Sprintf("  %s  %s", itemStr, valStr)
			}
			settingsRows = append(settingsRows, row)
		}

		settingsTitle := styles.SettingsTitle.Render("⚙️  SETTINGS")
		settingsList := strings.Join(settingsRows, "\n\n")
		settingsHelp := styles.SettingsHelp.Render("↑/↓: Navigate  •  ←/→: Adjust  •  Space/Enter: Toggle")

		pageContent = lipgloss.JoinVertical(
			lipgloss.Center,
			settingsTitle,
			"",
			settingsList,
			"",
			settingsHelp,
		)
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
