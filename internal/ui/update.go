package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
				} else if m.Timer.State == engine.StateRunning {
					m.Timer.Pause()
				} else if m.Timer.State == engine.StatePaused {
					m.Timer.Resume()
				}
				return m, nil

			case "up": // Up Arrow: Add 1 minute
				m.Timer.TimeRemaining += time.Minute
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
					completed, completedSession, _ := m.Timer.Tick(m.Timer.TimeRemaining, time.Now())
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
							m.State.Oasis.Tier = engine.GetTierForMinutes(m.State.Oasis.TotalFocusMinutes)
						}
						engine.UpdateStats(&m.State.Stats, completedSession.DurationMinutes, time.Now())
						_ = storage.SaveState(m.DbPath, m.State)
					}
				}
				return m, nil
			}
		}

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
