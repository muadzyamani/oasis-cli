package styles

import "github.com/charmbracelet/lipgloss"

// Color definitions (Harmonious purple theme)
var (
	ColorPurpleLight     = lipgloss.Color("#E9D5FF") // Light pastel purple
	ColorPurpleToday     = lipgloss.Color("#BB86FC") // Twilight Lavender Today card
	ColorPurpleLavender  = lipgloss.Color("#BB86FC") // Twilight Lavender (settings active item)
	ColorPurpleStreak    = lipgloss.Color("#9D4EDD") // Vibrant Purple (Current streak card)
	ColorPurpleLongest   = lipgloss.Color("#7B2CBF") // Rich Deep Purple (Longest streak card)
	ColorPurpleYesterday = lipgloss.Color("#757575") // Muted Gray for yesterday card
	ColorPurpleBgDark    = lipgloss.Color("#4C1D95") // Deep purple background for active tab
	ColorPurpleMuted     = lipgloss.Color("#7C7A90") // Muted purple-gray for inactive text/tabs/settings
	ColorPurpleBorder    = lipgloss.Color("#6D28D9") // Border purple for viewport
	ColorPurpleSeparator = lipgloss.Color("#3B2E5C") // Separator purple for header

	ColorRed             = lipgloss.Color("#CF6679") // Error / Warning Red
	ColorGray            = lipgloss.Color("#757575") // General muted text
)

// Lip Gloss Style Blocks
var (
	// Headers & Labels
	StreakStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPurpleStreak)

	// Navigation Tabs
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorPurpleBgDark).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorPurpleMuted).
			Padding(0, 2)

	TabSeparator = lipgloss.NewStyle().
			Foreground(ColorPurpleSeparator).
			SetString(" | ")

	// Layout Containers
	HeaderContainer = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorPurpleSeparator).
			PaddingBottom(1)

	ViewportContainer = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPurpleBorder).
				Padding(1, 2).
				Align(lipgloss.Center, lipgloss.Center)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorPurpleMuted).
			Italic(true)

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
			Foreground(ColorPurpleLavender).
			MarginBottom(1)

	SettingsItemActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPurpleLavender)

	SettingsItemInactive = lipgloss.NewStyle().
			Foreground(ColorPurpleMuted)

	SettingsValueActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPurpleLight)

	SettingsValueInactive = lipgloss.NewStyle().
			Foreground(ColorPurpleMuted)

	SettingsHelp = lipgloss.NewStyle().
			Foreground(ColorPurpleMuted).
			Italic(true).
			MarginTop(1)

	// Stats Styles
	StatsTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPurpleToday).
			MarginBottom(1)

	StatsCardToday = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPurpleToday).
			Width(28).
			Height(4).
			Align(lipgloss.Center, lipgloss.Center)

	StatsCardYesterday = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPurpleYesterday).
			Width(28).
			Height(4).
			Align(lipgloss.Center, lipgloss.Center)

	StatsCardCurrentStreak = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPurpleStreak).
			Width(28).
			Height(4).
			Align(lipgloss.Center, lipgloss.Center)

	StatsCardLongestStreak = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPurpleLongest).
			Width(28).
			Height(4).
			Align(lipgloss.Center, lipgloss.Center)
)
