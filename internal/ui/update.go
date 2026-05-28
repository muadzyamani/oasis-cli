package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
	"github.com/muadzyamani/oasis-cli/internal/engine"
	"github.com/muadzyamani/oasis-cli/internal/storage"
)

// Update routes incoming Bubble Tea messages to mutate Model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab", "l":
			m.ActiveTab = nextTab(m.ActiveTab)
			return m, nil

		case "shift+tab", "h":
			m.ActiveTab = prevTab(m.ActiveTab)
			return m, nil
		}

		// Main Viewport Controls (only active on TabOasis)
		if m.ActiveTab == TabOasis {
			switch msg.String() {
			case " ": // Space to start/pause/resume
				if m.Timer.State == engine.StateIdle {
					session := m.Timer.Start(time.Now())
					if session != nil {
						m.State.Sessions = append(m.State.Sessions, *session)
						_ = storage.SaveState(m.DbPath, m.State)
					}
					return m, tickCmd()
				} else if m.Timer.State == engine.StateRunning {
					m.Timer.Pause()
				} else if m.Timer.State == engine.StatePaused {
					m.Timer.Resume()
					return m, tickCmd()
				}
				return m, nil

			case "up": // Up Arrow: Add 1 minute
				m.Timer.TimeRemaining += time.Minute
				if m.Timer.State == engine.StateIdle {
					m.Timer.Duration = m.Timer.TimeRemaining
				} else {
					m.Timer.Duration += time.Minute
				}
				return m, nil

			case "down": // Down Arrow: Subtract 1 minute
				if m.Timer.TimeRemaining > time.Minute {
					m.Timer.TimeRemaining -= time.Minute
					if m.Timer.State == engine.StateIdle {
						m.Timer.Duration = m.Timer.TimeRemaining
					} else {
						m.Timer.Duration -= time.Minute
					}
				}
				return m, nil

			case "left": // Left Arrow: Reset timer
				if m.Timer.State == engine.StateRunning || m.Timer.State == engine.StatePaused {
					abandonedSession := m.Timer.Stop(time.Now())
					if abandonedSession != nil {
						// Update session status in local log
						found := false
						for i, s := range m.State.Sessions {
							if s.ID == abandonedSession.ID {
								m.State.Sessions[i].Status = "abandoned"
								m.State.Sessions[i].CompletedAt = abandonedSession.CompletedAt
								found = true
								break
							}
						}
						if !found {
							m.State.Sessions = append(m.State.Sessions, *abandonedSession)
						}
						_ = storage.SaveState(m.DbPath, m.State)
					}
				}
				m.Timer.Reset()
				return m, nil

			case "e": // e key: Toggle session type (working -> break -> long break)
				if m.Timer.State == engine.StateRunning || m.Timer.State == engine.StatePaused {
					abandonedSession := m.Timer.Stop(time.Now())
					if abandonedSession != nil {
						// Update session status in local log
						found := false
						for i, s := range m.State.Sessions {
							if s.ID == abandonedSession.ID {
								m.State.Sessions[i].Status = "abandoned"
								m.State.Sessions[i].CompletedAt = abandonedSession.CompletedAt
								found = true
								break
							}
						}
						if !found {
							m.State.Sessions = append(m.State.Sessions, *abandonedSession)
						}
						_ = storage.SaveState(m.DbPath, m.State)
					}
				}

				// Cycle through focus -> short-break -> long-break -> focus
				var nextType string
				switch m.Timer.SessionType {
				case "focus":
					nextType = "short-break"
				case "short-break":
					nextType = "long-break"
				case "long-break":
					nextType = "focus"
				default:
					nextType = "focus"
				}

				m.Timer.SessionType = nextType
				m.Timer.Reset()
				return m, nil

			case "s": // s key: Skip session
				if m.Timer.State == engine.StateIdle {
					// Toggle session type using standard cycle logic
					if m.Timer.SessionType == "focus" {
						if m.Timer.FocusSessionsCompleted > 0 && m.Timer.FocusSessionsCompleted%m.Timer.LongBreakInterval == 0 {
							m.Timer.SessionType = "long-break"
						} else {
							m.Timer.SessionType = "short-break"
						}
					} else {
						m.Timer.SessionType = "focus"
					}
					var duration time.Duration
					switch m.Timer.SessionType {
					case "focus":
						duration = time.Duration(m.Timer.Settings.FocusDuration) * time.Minute
					case "short-break":
						duration = time.Duration(m.Timer.Settings.ShortBreakDuration) * time.Minute
					case "long-break":
						duration = time.Duration(m.Timer.Settings.LongBreakDuration) * time.Minute
					default:
						duration = time.Duration(m.Timer.Settings.FocusDuration) * time.Minute
					}
					m.Timer.Duration = duration
					m.Timer.TimeRemaining = duration
				} else {
					// Complete active session immediately
					completed, completedSession, nextSession := m.Timer.Tick(m.Timer.TimeRemaining, time.Now())
					if completed && completedSession != nil {
						for i, s := range m.State.Sessions {
							if s.ID == completedSession.ID {
								m.State.Sessions[i].Status = "complete"
								m.State.Sessions[i].CompletedAt = completedSession.CompletedAt
								break
							}
						}
						if completedSession.Type == "focus" {
							m.State.Oasis.TotalFocusMinutes += completedSession.DurationMinutes
						}
						engine.UpdateStats(&m.State.Stats, completedSession.DurationMinutes, time.Now())
						if nextSession != nil {
							m.State.Sessions = append(m.State.Sessions, *nextSession)
						}
						_ = storage.SaveState(m.DbPath, m.State)
						sendSessionEndNotification(completedSession.Type, m.State.Settings.SoundEnabled)
					}
					if m.Timer.State == engine.StateRunning {
						return m, tickCmd()
					}
				}
				return m, nil
			}
		} else if m.ActiveTab == TabSettings {
			switch msg.String() {
			case "up", "k":
				m.SettingsCursor--
				if m.SettingsCursor < 0 {
					m.SettingsCursor = 6
				}
				return m, nil
			case "down", "j":
				m.SettingsCursor++
				if m.SettingsCursor > 6 {
					m.SettingsCursor = 0
				}
				return m, nil
			case "left":
				switch m.SettingsCursor {
				case 0: // Focus Duration
					if m.State.Settings.FocusDuration > 1 {
						m.State.Settings.FocusDuration--
						_ = storage.SaveState(m.DbPath, m.State)
						m.Timer.UpdateSettings(m.State.Settings)
					}
				case 1: // Break Duration
					if m.State.Settings.ShortBreakDuration > 1 {
						m.State.Settings.ShortBreakDuration--
						_ = storage.SaveState(m.DbPath, m.State)
						m.Timer.UpdateSettings(m.State.Settings)
					}
				case 2: // Long Break Duration
					if m.State.Settings.LongBreakDuration > 1 {
						m.State.Settings.LongBreakDuration--
						_ = storage.SaveState(m.DbPath, m.State)
						m.Timer.UpdateSettings(m.State.Settings)
					}
				case 3: // Sound
					m.State.Settings.SoundEnabled = !m.State.Settings.SoundEnabled
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 4: // View Streak
					m.State.Settings.ShowStreak = !m.State.Settings.ShowStreak
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 5: // Arabic Numerals
					m.State.Settings.UseArabicNumerals = !m.State.Settings.UseArabicNumerals
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 6: // Progress Bar Style
					stylesList := []string{
						"solid-capsule",
						"beaded-capsule",
						"framed-rounded",
					}
					currIdx := 0
					for idx, s := range stylesList {
						if s == m.State.Settings.ProgressBarStyle {
							currIdx = idx
							break
						}
					}
					nextIdx := (currIdx - 1 + len(stylesList)) % len(stylesList)
					m.State.Settings.ProgressBarStyle = stylesList[nextIdx]
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				}
				return m, nil
			case "right":
				switch m.SettingsCursor {
				case 0: // Focus Duration
					if m.State.Settings.FocusDuration < 120 {
						m.State.Settings.FocusDuration++
						_ = storage.SaveState(m.DbPath, m.State)
						m.Timer.UpdateSettings(m.State.Settings)
					}
				case 1: // Break Duration
					if m.State.Settings.ShortBreakDuration < 60 {
						m.State.Settings.ShortBreakDuration++
						_ = storage.SaveState(m.DbPath, m.State)
						m.Timer.UpdateSettings(m.State.Settings)
					}
				case 2: // Long Break Duration
					if m.State.Settings.LongBreakDuration < 120 {
						m.State.Settings.LongBreakDuration++
						_ = storage.SaveState(m.DbPath, m.State)
						m.Timer.UpdateSettings(m.State.Settings)
					}
				case 3: // Sound
					m.State.Settings.SoundEnabled = !m.State.Settings.SoundEnabled
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 4: // View Streak
					m.State.Settings.ShowStreak = !m.State.Settings.ShowStreak
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 5: // Arabic Numerals
					m.State.Settings.UseArabicNumerals = !m.State.Settings.UseArabicNumerals
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 6: // Progress Bar Style
					stylesList := []string{
						"solid-capsule",
						"beaded-capsule",
						"framed-rounded",
					}
					currIdx := 0
					for idx, s := range stylesList {
						if s == m.State.Settings.ProgressBarStyle {
							currIdx = idx
							break
						}
					}
					nextIdx := (currIdx + 1) % len(stylesList)
					m.State.Settings.ProgressBarStyle = stylesList[nextIdx]
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				}
				return m, nil
			case " ", "enter":
				switch m.SettingsCursor {
				case 3: // Sound
					m.State.Settings.SoundEnabled = !m.State.Settings.SoundEnabled
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 4: // View Streak
					m.State.Settings.ShowStreak = !m.State.Settings.ShowStreak
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 5: // Arabic Numerals
					m.State.Settings.UseArabicNumerals = !m.State.Settings.UseArabicNumerals
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				case 6: // Progress Bar Style
					stylesList := []string{
						"solid-capsule",
						"beaded-capsule",
						"framed-rounded",
					}
					currIdx := 0
					for idx, s := range stylesList {
						if s == m.State.Settings.ProgressBarStyle {
							currIdx = idx
							break
						}
					}
					nextIdx := (currIdx + 1) % len(stylesList)
					m.State.Settings.ProgressBarStyle = stylesList[nextIdx]
					_ = storage.SaveState(m.DbPath, m.State)
					m.Timer.UpdateSettings(m.State.Settings)
				}
				return m, nil
			}
		}

	case tickMsg:
		if m.Timer.State == engine.StateRunning {
			completed, completedSession, nextSession := m.Timer.Tick(time.Second, time.Time(msg))
			if completed {
				if completedSession != nil {
					for i, s := range m.State.Sessions {
						if s.ID == completedSession.ID {
							m.State.Sessions[i].Status = "complete"
							m.State.Sessions[i].CompletedAt = completedSession.CompletedAt
							break
						}
					}
					if completedSession.Type == "focus" {
						m.State.Oasis.TotalFocusMinutes += completedSession.DurationMinutes
					}
					engine.UpdateStats(&m.State.Stats, completedSession.DurationMinutes, time.Now())
					sendSessionEndNotification(completedSession.Type, m.State.Settings.SoundEnabled)
				}
				if nextSession != nil {
					m.State.Sessions = append(m.State.Sessions, *nextSession)
				}
				_ = storage.SaveState(m.DbPath, m.State)
			}
			if m.Timer.State == engine.StateRunning {
				return m, tickCmd()
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
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

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func sendSessionEndNotification(sessionType string, soundEnabled bool) {
	// --- NOTIFICATION CONFIGURATION ---
	// You can customize the app name, title, messages, and icon path below.
	
	// AppName controls the main header of the notification popup (especially on Windows/Linux).
	beeep.AppName = "Oasis"
	
	// Default values for notifications
	var title string
	var message string
	
	// Optional: path to an app icon (.png, .ico, etc.)
	// var appIcon string = "path/to/icon.png"
	var appIcon string = ""

	// Optional: path to a custom sound file (.wav, .mp3, etc.) for future implementation
	// var customSoundPath string = "assets/completed.wav"
	var customSoundPath string = ""

	if sessionType == "focus" {
		title = "Focus Session Completed"
		message = "Great job. Time to take a break."
	} else {
		title = "Break Completed"
		message = "Let's continue."
	}
	// ----------------------------------

	if soundEnabled {
		// Play the native OS notification alert sound
		_ = beeep.Alert(title, message, appIcon)

		// Placeholder for custom sound play in the future
		if customSoundPath != "" {
			// playCustomSound(customSoundPath)
		}
	} else {
		_ = beeep.Notify(title, message, appIcon)
	}
}

// playCustomSound plays a custom sound file from the specified path.
// In the future, this can be implemented using a library like github.com/faiface/beep or github.com/hajimehoshi/oto.
// func playCustomSound(soundPath string) {
// 	// Code to open and stream/play the sound file:
// 	// f, err := os.Open(soundPath)
// 	// if err != nil {
// 	// 	return
// 	// }
// 	// defer f.Close()
// 	// ... play logic ...
// }
