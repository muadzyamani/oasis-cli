# 🌴 Oasis CLI

A beautiful, premium retro-style Pomodoro productivity application for your terminal, written in Go using the **Bubble Tea** (Elm architecture) framework and styled with **Lip Gloss**.

Oasis CLI centers a large digital clock in your terminal, tracks your sessions, keeps a daily streak, and allows you to customize your focus and break schedules.

---

## 🚀 Getting Started

### Prerequisites
* **Go** 1.26.3 or higher installed on your system.
* A terminal supporting ANSI escape codes and styled colors (e.g., Windows Terminal, iTerm2, Alacritty, GNOME Terminal).
* **Terminal Viewport Size**: The application requires a minimum terminal viewport size of **80 x 24** characters to render properly.

---

## 🛠️ Build and Install

To build the project executable locally:

### On Windows
```powershell
# Build the executable
go build -o bin/oasis.exe ./cmd/oasis

# Run the executable
.\bin\oasis.exe
```

### On macOS / Linux
```bash
# Build the executable
go build -o bin/oasis ./cmd/oasis

# Run the executable
./bin/oasis
```

---

## 🧪 Running Tests

To run the automated test suite and ensure all components function correctly:

```bash
go test ./...
```

---

## 🕹️ Keyboard Controls

Navigation and interactions are fully keyboard-driven.

### 🌐 Universal Controls
| Key | Action |
| --- | --- |
| `Tab` / `l` | Switch to the next tab (Oasis ➔ Stats ➔ Settings) |
| `Shift + Tab` / `h` | Switch to the previous tab (Settings ➔ Stats ➔ Oasis) |
| `q` / `Ctrl + C` | Quit the application |

---

### ⏱️ Oasis Tab (Timer Controls)
| Key | Action |
| --- | --- |
| `Space` | Start, pause, or resume the timer |
| `Up Arrow` | Add 1 minute to the timer |
| `Down Arrow` | Subtract 1 minute from the timer |
| `Left Arrow` | Reset the timer (marks active running sessions as *abandoned*) |
| `e` | Cycle session type: **Focus Session** ➔ **Short Break** ➔ **Long Break** |
| `s` | Skip session: In **Idle** state, cycles session type. In **Active** state, completes the session immediately. |

---

### ⚙️ Settings Tab (Configuration)
Use settings to customize durations, sound alerts, streak visibility, and number styles.

| Key | Action |
| --- | --- |
| `Up Arrow` / `k` | Move the cursor up |
| `Down Arrow` / `j` | Move the cursor down |
| `Left Arrow` | Decrease duration or disable toggle |
| `Right Arrow` | Increase duration or enable toggle |
| `Space` / `Enter` | Toggle the selected setting (Sound, View Streak, Arabic Numerals) |

---

## 💾 Storage & State Persistence

Oasis CLI automatically serializes your settings, streak metrics, and session logs to a local JSON file:
* **Path**: `data/state.json` (created in the directory from which the executable is launched).

If the file does not exist, a default configuration is generated automatically upon the first startup.