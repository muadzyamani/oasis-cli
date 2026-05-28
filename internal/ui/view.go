package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muadzyamani/oasis-cli/internal/engine"
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
		tabLabel := strings.ToLower(string(t))
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
	header := styles.HeaderContainer.Width(m.Width).Render(topBar)

	// Dynamic height computation for layout components
	viewportHeight := m.Height - lipgloss.Height(header) - 5
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	viewportWidth := m.Width - 2
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
		pageContent = engine.RenderTimer(m.Timer, innerWidth, innerHeight)
	case TabStats:
		todayStr := time.Now().Format("2006-01-02")
		yesterdayStr := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

		todayMinutes := m.State.Stats.DailyRecords[todayStr]
		yesterdayMinutes := m.State.Stats.DailyRecords[yesterdayStr]

		// Add ongoing/paused focus session minutes if applicable
		if m.Timer.SessionType == "focus" && (m.Timer.State == engine.StateRunning || m.Timer.State == engine.StatePaused) {
			elapsed := m.Timer.Duration - m.Timer.TimeRemaining
			todayMinutes += int(elapsed.Minutes())
		}
		// Card 1: Today
		todayTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleToday).Render("TODAY")
		todayValStr := formatHoursMinutes(todayMinutes)
		todayValue := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleToday).Render(todayValStr + " focus")
		todayContent := lipgloss.JoinVertical(lipgloss.Center, todayTitle, todayValue)
		cardToday := styles.StatsCardToday.Render(todayContent)

		// Card 2: Yesterday
		yesterdayTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleYesterday).Render("YESTERDAY")
		yesterdayValStr := formatHoursMinutes(yesterdayMinutes)
		yesterdayValue := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleYesterday).Render(yesterdayValStr + " focus")
		yesterdayContent := lipgloss.JoinVertical(lipgloss.Center, yesterdayTitle, yesterdayValue)
		cardYesterday := styles.StatsCardYesterday.Render(yesterdayContent)

		// Card 3: Current Streak
		currStreakTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleStreak).Render("CURRENT STREAK")
		currStreakValStr := formatStreak(m.State.Stats.CurrentStreak)
		currStreakValue := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleStreak).Render(currStreakValStr + " streak")
		currStreakContent := lipgloss.JoinVertical(lipgloss.Center, currStreakTitle, currStreakValue)
		cardCurrent := styles.StatsCardCurrentStreak.Render(currStreakContent)

		// Card 4: Longest Streak
		longestStreakTitle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleLongest).Render("LONGEST STREAK")
		longestStreakValStr := formatStreak(m.State.Stats.LongestStreak)
		longestStreakValue := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPurpleLongest).Render(longestStreakValStr + " record")
		longestStreakContent := lipgloss.JoinVertical(lipgloss.Center, longestStreakTitle, longestStreakValue)
		cardLongest := styles.StatsCardLongestStreak.Render(longestStreakContent)

		// Layout Columns
		col1 := lipgloss.JoinVertical(lipgloss.Center, cardToday, cardYesterday)
		col2 := lipgloss.JoinVertical(lipgloss.Center, cardCurrent, cardLongest)
		cardsLayout := lipgloss.JoinHorizontal(lipgloss.Center, col1, "  ", col2)

		pageContent = lipgloss.JoinVertical(
			lipgloss.Center,
			styles.StatsTitle.Render("📊  Statistics"),
			cardsLayout,
		)
	case TabSettings:
		// Render settings options
		var settingsRows []string

		// Options definitions
		formatProgressBarStyle := func(style string) string {
			switch style {
			case "solid-capsule":
				return "Solid Capsule"
			case "beaded-capsule":
				return "Beaded Capsule"
			case "framed-rounded":
				return "Framed Rounded"
			default:
				return "Solid Capsule"
			}
		}

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
			{"Arabic Numerals", func() string {
				if m.State.Settings.UseArabicNumerals {
					return "[ Enabled ]"
				}
				return "[ Disabled ]"
			}()},
			{"Progress Bar Style", fmt.Sprintf("[ %s ]", formatProgressBarStyle(m.State.Settings.ProgressBarStyle))},
			{"Dev Mode", func() string {
				if m.State.Settings.DevMode {
					return "[ Enabled ]"
				}
				return "[ Disabled ]"
			}()},
			{"Hide Main Controls", func() string {
				if m.State.Settings.HideMainControls {
					return "[ Enabled ]"
				}
				return "[ Disabled ]"
			}()},
			{"Hide Timer Controls", func() string {
				if m.State.Settings.HideTimerControls {
					return "[ Enabled ]"
				}
				return "[ Disabled ]"
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

		settingsTitle := styles.SettingsTitle.Render("⚙️  Settings")
		settingsList := strings.Join(settingsRows, "\n")
		settingsHelp := styles.SettingsHelp.Render("↑/↓: Navigate  •  ←/→: Adjust  •  Space/Enter: Toggle")

		pageContent = lipgloss.JoinVertical(
			lipgloss.Center,
			settingsTitle,
			settingsList,
			settingsHelp,
		)
	}

	viewportBorder := styles.ViewportContainer.
		Width(viewportWidth).
		Height(viewportHeight).
		Render(pageContent)

	// 5. Render Footer
	var footer string
	if !m.State.Settings.HideMainControls {
		footerText := styles.FooterStyle.Render("Tab / Shift+Tab: Navigate  •  Q / Ctrl+C: Quit")
		footer = lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, footerText)
	}

	// Join entire viewport screen vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		viewportBorder,
		footer,
	)
}

func formatHoursMinutes(totalMinutes int) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func formatStreak(streak int) string {
	if streak == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", streak)
}
