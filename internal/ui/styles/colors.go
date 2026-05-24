package styles

import "github.com/charmbracelet/lipgloss"

// Color definitions (Harmonious palette: HSL / deep shades)
var (
	ColorGold    = lipgloss.Color("#FFB81C") // Solar Yellow
	ColorCyan    = lipgloss.Color("#00E5FF") // Pool Cyan
	ColorOrange  = lipgloss.Color("#FF7A00") // Sunset Orange
	ColorLavender = lipgloss.Color("#BB86FC") // Twilight Lavender
	ColorRed     = lipgloss.Color("#CF6679") // Error / Warning Red
	ColorGray    = lipgloss.Color("#757575") // Muted text
	ColorBgDark  = lipgloss.Color("#121212") // Surface dark
)

// Lip Gloss Style Blocks
var (
	// Headers & Labels
	StreakStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorOrange)

	// Navigation Tabs
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#3B4252")).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Padding(0, 2)

	TabSeparator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2E3440")).
			SetString(" | ")

	// Layout Containers
	HeaderContainer = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#2E3440")).
			PaddingBottom(1).
			MarginBottom(1)

	ViewportContainer = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#434C5E")).
				Padding(1, 2).
				Align(lipgloss.Center, lipgloss.Center)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true).
			PaddingTop(1)

	// Warnings
	WarningTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorRed)

	WarningBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorRed).
			Padding(1, 3).
			Align(lipgloss.Center, lipgloss.Center)

	// Settings Styles
	SettingsTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorLavender).
			MarginBottom(1)

	SettingsItemActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	SettingsItemInactive = lipgloss.NewStyle().
			Foreground(ColorGray)

	SettingsValueActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold)

	SettingsValueInactive = lipgloss.NewStyle().
			Foreground(ColorGray)

	SettingsHelp = lipgloss.NewStyle().
			Foreground(ColorGray).
			Italic(true).
			MarginTop(1)
)
