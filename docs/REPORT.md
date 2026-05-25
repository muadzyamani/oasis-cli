# Oasis — Focus & Grow

**A calm-focused Pomodoro productivity web application where every completed focus session grows a living desert oasis.**

- **Author:** Muadz Yamani
- **License:** MIT
- **Repository:** [github.com/muadzyamani/oasis](https://github.com/muadzyamani/oasis)
- **Status:** Phase 2 complete (see roadmap below)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Tech Stack](#2-tech-stack)
3. [Core Loop & User Experience](#3-core-loop--user-experience)
4. [Architecture](#4-architecture)
5. [Project Structure](#5-project-structure)
6. [Data Model](#6-data-model)
7. [State Management & Data Flow](#7-state-management--data-flow)
8. [Growth Engine](#8-growth-engine)
9. [Ambient Engine](#9-ambient-engine)
10. [Visual Design System](#10-visual-design-system)
11. [UI Components](#11-ui-components)
12. [Persistence](#12-persistence)
13. [Development & Build](#13-development--build)
14. [Roadmap](#14-roadmap)

---

## 1. Overview

Oasis reimagines the traditional Pomodoro timer by coupling it with a visually rich, always-present desert scene that grows and changes in response to the user's focus sessions. The core philosophy is gentle, forward-only motivation — there is no punishment, the oasis never wilts, and growth is permanent.

The application is a **client-side single-page application (SPA)** with **no backend server**. All data persists locally in the browser's `localStorage`. The app is anonymous-first by design with optional cloud sync planned for a future phase.

---

## 2. Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Language** | TypeScript 6.0 (strict mode) | Type-safe development |
| **UI Library** | React 19 | Component rendering |
| **Build Tool** | Vite 8 (SWC transforms) | Dev server & production bundling |
| **Styling** | Tailwind CSS v4 + CSS Custom Properties | Utility-first styling + design tokens |
| **Animation** | Framer Motion 12 | Spring physics, layout animations, AnimatePresence |
| **State Management** | Zustand 5 (with `persist` middleware) | Lightweight global state + localStorage persistence |
| **Linting** | ESLint 10 (typescript-eslint, react-hooks) | Code quality |
| **Formatting** | Prettier 3 | Consistent code style |
| **Fonts** | EB Garamond (display), Plus Jakarta Sans (body), DM Mono (monospace) | Typography system |

---

## 3. Core Loop & User Experience

### The Loop

1. User opens the app and sees their oasis scene — a desert landscape with sky, ground, water, and any previously grown plants.
2. User starts a **focus session** (default 25 minutes, configurable 1–120 min).
3. A timer counts down in a **Web Worker** (immune to tab-background throttling).
4. During the session, a **ghost preview** of the next plant appears in the scene, fading in gradually as time progresses (opacity 0.25 → 0.72, scale 0.45 → 0.95).
5. When the session completes, the preview disappears and a **permanent plant element** is added to the oasis with a spring animation. A toast notification announces the new plant with a poetic label (e.g., "The Reed of Monday, 12th May").
6. The user can take a short break (5 min) or long break (15 min) before starting the next focus session.
7. The oasis progresses through **6 growth tiers** based on total lifetime focus minutes — each tier unlocks new plant types and changes the environment.

### Timer States

| State | Visual | Behavior |
|-------|--------|----------|
| **idle** | Expanded glass card centered on screen | "Enter Oasis" button, session type selector, streak display |
| **active** | Compact bar pinned to bottom | Countdown, pause button, progress ring (compact) |
| **paused** | Compact bar pinned to bottom | Resume/stop buttons, time frozen |
| **complete** | Expanded glass card centered | "Well done" / "Rest complete" message, "Take a Break" button |
| **abandoned** | Returns to idle | Session logged as abandoned, no growth event triggered |

### Session Types

- **Focus** — Default 25 min timer. Only focus sessions trigger growth events.
- **Short Break** — Default 5 min. Visual atmosphere shifts to a cool blue overlay.
- **Long Break** — Default 15 min. Automatically offered after every 4 focus sessions.

### Strengths & Streaks

Daily focus activity is tracked. The streak (consecutive days with at least one completed session) is displayed prominently on the idle timer card and in the stats panel. Streaks are computed by walking backwards from the current date through daily records.

---

## 4. Architecture

### Layered Z-Index Model

The UI is structured as **CSS z-index layers** rather than separate pages:

| Layer | z-index | Component | Purpose |
|-------|---------|-----------|---------|
| 0 | `--z-scene: 0` | OasisScene (Sky, Ground, Water, Vegetation, Flora, Atmosphere layers) | The living animated environment |
| 1 | `--z-widget: 20` | TimerWidget | Floating centered timer card / compact bar |
| 2 | `--z-nav: 50` | NavigationRail | Left-edge icon sidebar (stats, settings) |
| 3 | `--z-overlay-backdrop: 30` + `--z-overlay: 40` | SettingsPanel / StatsPanel | Slide-in frosted glass drawers |
| 4 | `--z-toast: 70` | MilestoneToast | Bottom-center notification |

### Component Tree

```
<App>
  <AppShell>                            # Root fullscreen container (isolation: isolate)
    <OasisScene>                        # Always-visible background environment
      <SkyLayer />                       # css gradient sky + svg sun, moon, stars
      <GroundLayer />                    # Sand dunes with elevation shading
      <WaterLayer />                     # Animated pool with ripple rings
      <VegetationLayer />               # Static decorative flora (grass, bushes, reeds, lilies)
      <FloraLayer />                     # Planted elements + live session preview sapling
      <AtmosphereLayer />               # Float dust particles, focus warm glow, break overlay
    </OasisScene>
    
    <TimerWidget>                       # Main timer UI
      // Expanded (idle/complete):
        <TimerRing>                      # SVG circular progress ring
        <SessionControls>               # Start/Pause/Resume/Stop + type selector
      // Compact (active/paused):
        <CompactBar>                     # Bottom-pinned minimal bar
      <MilestoneToast />                 # Growth milestone notification
    </TimerWidget>
    
    <NavigationRail />                   # Stats + Settings toggle buttons
    
    <SettingsPanel>                      # Slide-in left drawer
      <TimerSection />                  # Focus/break duration steppers (1–120 min)
      <AutomationSection />             # Auto-start toggles for breaks/focus
      <AppearanceSection />             # Theme (auto/light/dark) + reduced motion
      <SundialSection />                # Sunrise/sunset hour configuration
      <OasisSection />                  # Clear all plants button
    </SettingsPanel>
    
    <StatsPanel>                         # Slide-in left drawer
      # Focus Growth card: active plant preview, progress bar, elapsed time
      # Stats grid: Streak, Total Focus hours, Plants count, Today's focus
    </StatsPanel>
    
    {DEV && <TimeDebugPanel />}          # Dev-only: sky time scrubber for testing
  </AppShell>
</App>
```

---

## 5. Project Structure

```
oasis/
├── .gitignore
├── .prettierrc
├── eslint.config.js
├── index.html                        # HTML entry point
├── package.json
├── tsconfig.json                     # Root TS config (references app + node)
├── tsconfig.app.json                 # App TypeScript config
├── tsconfig.node.json                # Node (vite.config.ts) TS config
├── vite.config.ts                    # Vite build config (SWC + Tailwind plugins, @/ alias)
│
├── dist/                             # Production build output
│
├── docs/
│   ├── implementation_plan.md        # Full product design document
│   ├── implementation_plan_phase_1.md
│   ├── implementation_plan_phase_2.md
│   └── known_bugs.md
│
└── src/
    ├── main.tsx                       # React entry point
    ├── App.tsx                        # Root component — orchestrates all layers
    │
    ├── components/
    │   ├── dev/
    │   │   └── TimeDebugPanel.tsx     # Dev sky scrubber
    │   ├── overlays/
    │   │   ├── SettingsPanel.tsx      # Main settings drawer
    │   │   ├── StatsPanel.tsx         # Statistics drawer
    │   │   └── settings/
    │   │       ├── TimerSection.tsx
    │   │       ├── AutomationSection.tsx
    │   │       ├── AppearanceSection.tsx
    │   │       ├── SundialSection.tsx
    │   │       └── OasisSection.tsx
    │   ├── scene/
    │   │   ├── OasisScene.tsx         # Scene composition root
    │   │   ├── SkyLayer.tsx           # Sun, moon, stars, sky gradient
    │   │   ├── AtmosphereLayer.tsx    # Dust particles, focus glow, break overlay
    │   │   ├── GroundLayer.tsx        # Sand dunes
    │   │   ├── WaterLayer.tsx         # Water pool with ripple animation
    │   │   ├── VegetationLayer.tsx    # Static decorative flora
    │   │   ├── FloraLayer.tsx         # Planted elements + preview sapling
    │   │   └── elements/              # Individual plant SVG components
    │   │       ├── Sprout.tsx, Flower.tsx, Reed.tsx
    │   │       ├── DatePalm.tsx, MaturePalm.tsx
    │   │       ├── Acacia.tsx, MatureAcacia.tsx
    │   │       ├── Succulent.tsx, MatureSucculent.tsx
    │   │       ├── DesertWillow.tsx, MatureDesertWillow.tsx
    │   │       ├── Lantern.tsx, Lily.tsx, Firefly.tsx
    │   ├── timer/
    │   │   ├── TimerWidget.tsx        # Main timer (expanded + compact modes)
    │   │   ├── TimerRing.tsx          # SVG circular progress ring
    │   │   ├── SessionControls.tsx    # Start/Pause/Resume/Stop buttons
    │   │   └── CompactBar.tsx         # Minimized active timer bar
    │   └── ui/
    │       ├── AppShell.tsx           # Root layout container
    │       ├── NavigationRail.tsx     # Left sidebar buttons
    │       ├── GlassPanel.tsx         # Glassmorphism surface
    │       ├── SettingsToggle.tsx     # iOS-style toggle
    │       ├── SettingsStepper.tsx    # +/- value stepper
    │       └── MilestoneToast.tsx     # Toast notification component
    │
    ├── engines/                       # Pure logic modules (zero React imports)
    │   ├── ambientEngine.ts           # Sky color math, solar geometry, lunar phase
    │   ├── growthEngine.ts            # Growth tier & element selection logic
    │   └── timerWorker.ts             # Web Worker for accurate countdowns
    │
    ├── hooks/                         # Custom React hooks
    │   ├── useTimer.ts                # Bridges Web Worker to stores, owns session pipeline
    │   ├── useAmbient.ts              # Polls clock every 30s, computes atmosphere state
    │   ├── useOasisGrowth.ts          # Reads oasis state + computes growth progress
    │   └── useStreak.ts               # Exposes streak + today's stats for components
    │
    ├── stores/                        # Zustand state stores
    │   ├── timerStore.ts              # Timer countdown, status, config (config persisted)
    │   ├── sessionStore.ts            # Session history (history persisted, current not)
    │   ├── oasisStore.ts              # Oasis elements, tier, total focus minutes (fully persisted)
    │   ├── settingsStore.ts           # User preferences (fully persisted)
    │   ├── statsStore.ts              # Aggregated stats + daily records (fully persisted)
    │   └── devStore.ts                # Dev time overrides (NOT persisted)
    │
    ├── styles/
    │   ├── tokens.css                 # Design token CSS custom properties
    │   └── globals.css                # Global styles, glassmorphism classes, font imports
    │
    ├── types/                         # TypeScript type definitions
    │   ├── oasis.types.ts             # OasisElement, OasisState, GrowthTier, GROWTH_TIERS config
    │   ├── session.types.ts           # Session, SessionConfig, SessionStatus, SessionType
    │   └── growth.types.ts            # AtmosphereState, TimeOfDay, MilestoneEvent
    │
    └── utils/
        └── formatters.ts             # Countdown display, duration formatting, element label generation
```

---

## 6. Data Model

All data is persisted in `localStorage` via Zustand's `persist` middleware. There is no backend database.

### Oasis State (`oasis.types.ts`)

```typescript
interface OasisState {
  name: string                    // "My Oasis"
  tier: GrowthTier                // 0–5
  totalFocusMinutes: number       // Cumulative lifetime focus minutes
  elements: OasisElement[]        // All planted plants, trees, and objects
  createdAt: number               // Unix timestamp (ms)
}

interface OasisElement {
  id: string                      // Unique element ID
  type: OasisElementType          // See types below
  plantedAt: number               // Unix timestamp when planted
  sessionId: string               // Links to the session that produced this element
  label: string                   // e.g. "The Reed of Monday, 12th May"
  position: { x: number; y: number }  // Percentage position (0–100)
  tier: GrowthTier                // Which growth tier unlocked this element
  stage: GrowthStage              // 'sapling' (preview/growing) | 'mature' (permanent)
}

type OasisElementType =
  | 'sprout' | 'flower' | 'reed' | 'palm'
  | 'lantern' | 'lily' | 'waterfall' | 'firefly'
  | 'acacia' | 'succulent' | 'willow'
```

### Growth Tiers

| Tier | Name | Min Minutes | Unlocked Elements | Description |
|------|------|-------------|-------------------|-------------|
| 0 | Seed | 0 | palm | A bare sandy clearing, full of potential. |
| 1 | Sprout | 25 | acacia | Your first plant takes root. A pool shimmers. |
| 2 | Bloom | 120 | succulent | Flowers open, the pool fills, first lantern flickers. |
| 3 | Grove | 480 | willow | Palms cast soft shadows. Fireflies at dusk. |
| 4 | Sanctuary | 1440 | — | Full canopy, lotus flowers, sound of water. |
| 5 | Eden | 4320 | — | A lush, breathing world. Every corner holds wonder. |

### Session (`session.types.ts`)

```typescript
interface Session {
  id: string
  type: SessionType               // 'focus' | 'short-break' | 'long-break'
  startedAt: number
  completedAt?: number
  durationMinutes: number
  status: SessionStatus           // 'idle' | 'active' | 'paused' | 'break' | 'complete' | 'abandoned'
  label?: string
  oasisElementId?: string
}

interface SessionConfig {
  focusDurationMinutes: number    // Default: 25
  shortBreakMinutes: number       // Default: 5
  longBreakMinutes: number        // Default: 15
  longBreakAfterSessions: number  // Default: 4
  autoStartBreaks: boolean        // Default: false
  autoStartFocus: boolean         // Default: false
}
```

### Daily Record (`statsStore.ts`)

```typescript
interface DailyRecord {
  date: string                    // 'YYYY-MM-DD'
  focusMinutes: number
  sessionsCompleted: number
}
```

### Settings (`settingsStore.ts`)

```typescript
interface SettingsState {
  soundEnabled: boolean           // Default: false
  ambienceVolume: number          // 0–1, Default: 0.5
  uiSoundsEnabled: boolean        // Default: true
  theme: AppTheme                 // 'auto' | 'light' | 'dark', Default: 'auto'
  showStreakOnOpen: boolean       // Default: true
  focusWindowStart: number        // Hour 0–23, Default: 8
  focusWindowEnd: number          // Hour 0–23, Default: 22
  notificationsEnabled: boolean   // Default: false
  reducedMotion: boolean          // Default: false
  sunriseHour: number             // Default: 7
  sunsetHour: number              // Default: 19
}
```

### localStorage Keys

| Key | Store | Content Persisted |
|-----|-------|-------------------|
| `oasis-timer` | timerStore | config only (not live timer state) |
| `oasis-sessions` | sessionStore | sessionHistory only (not current session) |
| `oasis-world` | oasisStore | full oasis state |
| `oasis-settings` | settingsStore | all user preferences |
| `oasis-stats` | statsStore | aggregated statistics + daily records |

---

## 7. State Management & Data Flow

### Store Architecture

The application uses **6 Zustand stores**, each responsible for a distinct domain:

```
timerWorker.ts (Web Worker — off the main thread)
    │  posts TICK / COMPLETE messages
    ▼
useTimer.ts (hook — bridges worker ↔ stores)
    │  runs the session-completion pipeline
    ▼
┌──────────────── Timer Store ────────────────┐
│  timerStore.ts                               │
│  Fields: status, sessionType,                │
│          timeRemainingSeconds, config         │
│  Persisted: config only                      │
├──────────────────────────────────────────────┤
│  sessionStore.ts                             │
│  Fields: currentSession, sessionHistory      │
│  Persisted: sessionHistory only              │
├──────────────────────────────────────────────┤
│  oasisStore.ts                               │
│  Fields: oasis (name, tier, elements,         │
│          totalFocusMinutes), derived state    │
│  Persisted: full oasis                       │
├──────────────────────────────────────────────┤
│  statsStore.ts                               │
│  Fields: totalFocusMinutes, totalSessions,   │
│          streaks, dailyRecords               │
│  Persisted: all (derived via compute)        │
├──────────────────────────────────────────────┤
│  settingsStore.ts                            │
│  Fields: all user preferences                │
│  Persisted: all                              │
├──────────────────────────────────────────────┤
│  devStore.ts                                 │
│  Fields: timeOverride                        │
│  NOT persisted                               │
└──────────────────────────────────────────────┘
```

### Session Completion Pipeline

The pipeline is orchestrated in `useTimer.ts` via an `onCompleteRef` callback that fires when the Web Worker sends a `COMPLETE` message:

1. Web Worker sends `COMPLETE` message to main thread.
2. `onCompleteRef.current()` fires.
3. If the session type was `'focus'`:
   - `sessionStore.completeSession()` — marks the session as complete, adds to history.
   - `statsStore.recordCompletedSession(session)` — updates daily records, recomputes streaks.
   - `growthEngine.resolveGrowthEvent(session, oasis)` — pure function: determines what plant type, position, and whether a tier-up occurred.
   - `oasisStore.addElement(type, sessionId, ...)` — creates a permanent element in the oasis.
   - `oasisStore.addFocusMinutes(duration)` — updates total focus minutes and checks for tier-up.
4. The `plantedElementId` state triggers `TimerWidget` to show a `MilestoneToast`.
5. `timerStore` status set to `'complete'`.

### Timer Web Worker (`timerWorker.ts`)

- Runs in a dedicated Web Worker (offloaded from main thread).
- Uses `setInterval` at 250ms with `Date.now()` delta correction.
- Messages: `START`, `PAUSE`, `RESUME`, `STOP` (incoming); `TICK`, `COMPLETE` (outgoing).
- Immune to browser tab-background throttling because it computes elapsed time via `Date.now()`.

### Custom Hooks

| Hook | Purpose | Key Responsibilities |
|------|---------|---------------------|
| `useTimer` | Bridges Web Worker to stores | Manages worker lifecycle, owns session pipeline, computes preview element |
| `useAmbient` | Polls real clock | Reads current time (or dev override), reads settings (sunrise/sunset), calls `computeAtmosphere()` every 30s |
| `useOasisGrowth` | Oasis state + progress | Reads oasis store, computes progress-to-next-tier (0–1) |
| `useStreak` | Streak + today's stats | Reads stats store, exposes current/longest streak, today's minutes/sessions |

---

## 8. Growth Engine (`src/engines/growthEngine.ts`)

A pure-function module with no side effects, responsible for determining what the oasis gains from a completed session.

### Key Functions

- **`resolveGrowthEvent(session, oasis)`** — Given a completed session and current oasis state, returns a `GrowthEvent` containing:
  - `elementType`: What plant type to grow (e.g., palm, acacia, etc.)
  - `position`: Smart placement coordinates (x%, y%) within the scene
  - `tierUp`: Whether the total focus minutes crossed a tier threshold
  - `newTier`: The resulting tier

- **`peekNextElement(oasis)`** — Pre-computes the next plant type and position without a completed session. Used to show the "ghost preview" during an active session.

### Element Selection Logic

- The 4 **growing plant types** (palm, acacia, succulent, willow) cycle round-robin based on the count of existing elements.
- Elements are unlocked by tier — only plant types whose tier threshold has been crossed are eligible.
- Each element type has pre-defined **position zones** with randomized jitter (±3% x, ±2% y) so plants never perfectly overlap.
- Non-growing decorative elements (sprout, flower, reed, lantern, lily, waterfall, firefly) are available for future milestone-based placement.

---

## 9. Ambient Engine (`src/engines/ambientEngine.ts`)

A continuous 24-hour sky system driven by minute-accurate solar math, replacing a simpler 5-bucket time-of-day approach.

### Features

- **Sky Colors:** 14 RGB keyframes across 24 hours, linearly interpolated by minute of day. Covers midnight through dawn, sunrise, morning, noon, afternoon, sunset, dusk, and night.
- **Solar Elevation:** Sinusoidal arc from sunrise to sunset. Negative values below the horizon (night).
- **Sun Position:** Left-to-right semi-circular arc (5% → 95% viewport x, arcs from horizon to zenith and back).
- **Moon Position:** Right-to-left semi-circular arc during night (opposite of sun).
- **Sun Visuals:** Size, color, and glow vary dynamically with elevation — large and orange at the horizon, small and white at zenith. Glow includes inner halo, outer diffuse atmosphere, and optional atmospheric scatter band near the horizon.
- **Moon Phase:** Real lunar phase calculated from a known new moon reference date (2000-01-06). Phase ranges 0 (new) through 0.5 (full) to 1 (new).
- **Stars:** 65 seeded pseudo-random stars with twinkle animations (varying sizes, opacities, and animation durations).
- **Brightness Overlay:** Subtle white overlay at solar noon for a realistic brightness effect.

### Key Functions

- `computeAtmosphere(hour, minute, sunriseHour, sunsetHour, sessionActive, sessionProgress, isBreak)` — Returns a complete `AtmosphereState` object.
- `interpolateSkyColors(minuteOfDay)` — Returns 5 color channels (skyTop, skyBottom, groundColor, waterColor, groundFar) as hex strings.
- `computeSolarElevation(minuteOfDay, sunriseMinute, sunsetMinute)` — Returns -1 (nadir) to 1 (zenith).
- `computeSunPosition` / `computeMoonPosition` — Returns viewport percentage coordinates or null.
- `getLunarPhase(date)` — Returns lunar phase 0–1.
- `getSunProps(elevation)` — Returns sun visual properties (size, disc color, glow color, glow radius).

---

## 10. Visual Design System

### Design Tokens (`src/styles/tokens.css`)

All visual decisions are expressed as CSS custom properties consumed by both Tailwind utility classes and component inline styles.

#### Color Palette

| Category | Purpose | Example Tokens |
|----------|---------|----------------|
| **Scene/Sand** | Desert ground tones | `--color-sand-dawn: #f5e6c8`, `--color-sand-warm: #d4a96a` |
| **Water** | Oasis pool colors | `--color-oasis-water: #2a6b7c`, `--color-oasis-water-light: #4a8fa0` |
| **Flora** | Plant greens | `--color-palm-green: #3d6b4f`, `--color-reed-green: #6b8c3d` |
| **Atmosphere** | Dawn/dusk tones | `--color-dusk-rose: #c97b5a`, `--color-dawn-lavender: #c4a8d0` |
| **Night** | Night sky colors | `--color-night-deep: #0d1b2a`, `--color-night-mid: #162436` |
| **Accent** | Primary brand | `--color-lantern-gold: #f0c060` (the single accent color) |
| **Text** | Semantic text colors | Primary, secondary, muted, on-dark variants |

#### Glassmorphism Materials

Three CSS classes simulate Apple-style frosted glass without using `backdrop-filter` (to avoid GPU compositor re-promotion flicker):

- `.glass-panel` — Primary surface with warm frost gradient, inset specular shadow, ambient lift shadow
- `.glass-panel-dark` — Dark variant for deeper overlays
- `.glass-surface` — Subtle secondary surface for nested panels and pills

#### Typography

| Role | Font | Fallback |
|------|------|----------|
| Display / Headings | EB Garamond (serif) | Palatino Linotype, Georgia |
| Body / UI | Plus Jakarta Sans (sans-serif) | Inter, system-ui |
| Monospace / Timer | DM Mono (monospace) | JetBrains Mono, Courier New |

Type scale ranges from `xs` (0.75rem) through `6xl` (3.75rem).

#### Z-Index Layers

`--z-scene: 0` → `--z-scene-top: 10` → `--z-widget: 20` → `--z-overlay-backdrop: 30` → `--z-overlay: 40` → `--z-nav: 50` → `--z-modal: 60` → `--z-toast: 70`

#### Animation Tokens

- Durations: instant (80ms) → fast (150ms) → normal (300ms) → slow (600ms) → very-slow (1200ms) → cinematic (2400ms)
- Easings: smooth (`cubic-bezier(0.4, 0, 0.2, 1)`), spring (`0.34, 1.56, 0.64, 1`), breath (`0.45, 0.05, 0.55, 0.95`)

---

## 11. UI Components

### Scene Components

| Component | File | Description |
|-----------|------|-------------|
| `OasisScene` | `scene/OasisScene.tsx` | Root scene layout — composes all 5 sub-layers with `aria-hidden="true"` |
| `SkyLayer` | `scene/SkyLayer.tsx` | Animated sky gradient, 65 twinkling stars (seeded), SVG sun with multi-layer glow, SVG moon with real lunar phase |
| `GroundLayer` | `scene/GroundLayer.tsx` | SVG sand dunes with color that shifts by time of day |
| `WaterLayer` | `scene/WaterLayer.tsx` | Animated water pool with concentric ripple rings, color shifts with time + tier |
| `VegetationLayer` | `scene/VegetationLayer.tsx` | Static decorative flora (grass tufts, desert bushes, reeds, floating lilies) |
| `FloraLayer` | `scene/FloraLayer.tsx` | All planted elements + live preview sapling; uses `AnimatePresence` for spring-based appear animations |
| `AtmosphereLayer` | `scene/AtmosphereLayer.tsx` | 10 floating dust particles (infinite horizontal scroll), warm focus glow at horizon, cool break overlay, session milestone pulse |

### Plant Elements (`scene/elements/`)

| Component | Type | Stage(s) |
|-----------|------|----------|
| `DatePalm` / `MaturePalm` | palm | sapling, mature |
| `Acacia` / `MatureAcacia` | acacia | sapling, mature |
| `Succulent` / `MatureSucculent` | succulent | sapling, mature |
| `DesertWillow` / `MatureDesertWillow` | willow | sapling, mature |
| `Sprout` | sprout | non-growing |
| `Flower` | flower | non-growing |
| `Reed` | reed | non-growing |
| `Lantern` | lantern | non-growing |
| `Lily` | lily | non-growing |
| `Firefly` | firefly | non-growing |

Each plant is an SVG component. The 4 growing types have both a "sapling" (small, understated) and "mature" (larger, more detailed) variant.

### Timer Components

| Component | File | Description |
|-----------|------|-------------|
| `TimerWidget` | `timer/TimerWidget.tsx` | Orchestrates expanded vs compact mode, milestone toasts, next-session logic |
| `TimerRing` | `timer/TimerRing.tsx` | SVG circular progress ring with gray track and gold indicator |
| `SessionControls` | `timer/SessionControls.tsx` | Start / Pause / Resume / Stop buttons + session type tabs (Focus, Short Break, Long Break) |
| `CompactBar` | `timer/CompactBar.tsx` | Bottom-pinned minimized bar showing countdown, progress ring, + pause/resume/stop buttons |

### Overlay Components

| Component | File | Description |
|-----------|------|-------------|
| `SettingsPanel` | `overlays/SettingsPanel.tsx` | Slide-in left drawer; 5 sub-sections; spring animation; backdrop click + `Escape` to dismiss |
| `StatsPanel` | `overlays/StatsPanel.tsx` | Slide-in left drawer; shows active plant preview card, growth progress, 4-stat grid (Streak, Total Focus, Plants, Today) |
| `TimerSection` | `overlays/settings/TimerSection.tsx` | Steppers for focus/break duration (1–120 min) |
| `AutomationSection` | `overlays/settings/AutomationSection.tsx` | Toggles for auto-start breaks and auto-start focus |
| `AppearanceSection` | `overlays/settings/AppearanceSection.tsx` | Theme selector (auto/light/dark) + reduced motion toggle |
| `SundialSection` | `overlays/settings/SundialSection.tsx` | Sunrise/sunset hour steppers |
| `OasisSection` | `overlays/settings/OasisSection.tsx` | "Clear the Oasis" reset button |

### UI Primitives

| Component | File | Description |
|-----------|------|-------------|
| `AppShell` | `ui/AppShell.tsx` | Root fullscreen container (`isolation: isolate`) |
| `NavigationRail` | `ui/NavigationRail.tsx` | Left-edge vertical icon bar (Stats + Settings) with hover/active states |
| `GlassPanel` | `ui/GlassPanel.tsx` | Reusable glassmorphism surface |
| `SettingsToggle` | `ui/SettingsToggle.tsx` | iOS-style toggle switch |
| `SettingsStepper` | `ui/SettingsStepper.tsx` | +/- value stepper with min/max bounds |
| `MilestoneToast` | `ui/MilestoneToast.tsx` | Bottom-center toast with spring animation for growth milestones |

---

## 12. Persistence

All persistent data lives in `localStorage`. The application writes to 5 keys:

| Key | Contents | Update Frequency |
|-----|----------|------------------|
| `oasis-timer` | `SessionConfig` (durations, auto-start preferences) | On settings change |
| `oasis-sessions` | Array of completed/abandoned `Session` objects | On session completion or abandon |
| `oasis-world` | Full `OasisState` (name, tier, total minutes, all elements) | On session completion (growth event) |
| `oasis-settings` | All user preferences | On any settings change |
| `oasis-stats` | Aggregated statistics, streaks, daily records | On session completion |

**Important:** Live timer state (current countdown, active status) is **not** persisted. Reloading the page resets the timer to idle. Session history, oasis state, and preferences survive reloads.

---

## 13. Development & Build

### Prerequisites

- Node.js (tested with recent LTS)
- npm

### Scripts

| Script | Command | Description |
|--------|---------|-------------|
| `dev` | `vite` | Start development server (hot module replacement) |
| `dev-host` | `vite --host` | Start dev server, accessible on network |
| `build` | `tsc -b && vite build` | Type-check + production build |
| `preview` | `vite preview` | Preview production build locally |
| `lint` | `eslint .` | Run ESLint across all files |
| `format` | `prettier --write src/` | Format source code with Prettier |

### Build Configuration

- **TypeScript:** Strict mode, target ES2020, JSX react-jsx, path alias `@/*` → `src/*`
- **Vite:** SWC-based React fast refresh, Tailwind CSS v4 plugin, `@/` path alias
- **ESLint:** Flat config, `typescript-eslint`, `react-hooks` plugin, no-unused-vars (ignoring `_`)
- **Prettier:** No semicolons, single quotes, trailing commas, 100 print width

---

## 14. Roadmap

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 1 | ✅ Complete | Project foundations, design system, store architecture |
| Phase 2 | ✅ Complete | Live animated Oasis scene + functional Pomodoro timer |
| Phase 3 | 🔜 Next | Audio system (Web Audio API ambience + UI sounds), onboarding flow, additional overlays |
| Phase 4 | 📋 Planned | Cloud sync (Supabase), user accounts, themes, seasonal events |
| Phase 5 | 📋 Planned | Shared focus rooms, companion features, AI-generated environments |

### Phase 3 Planned Features

- `audioManager.ts` engine — Web Audio API for procedural desert ambience (wind, water, insects) + UI sound effects
- Full onboarding flow with step-by-step introduction
- Additional scene details and animations
- Enhanced milestone celebrations

---

## Key Design Principles

1. **No Punishment Mechanics** — The oasis never wilts, dies, or loses progress. All growth is permanent and forward-only.
2. **Whisper, Don't Shout** — All feedback is gentle and atmospheric. No harsh alarms, no aggressive notifications.
3. **Layers, Not Pages** — The living scene is the constant background. UI elements float above it as transparent, frosted-glass overlays.
4. **Positive Reinforcement** — Every completed session is celebrated with a new plant, a toast message, and visible progress toward the next tier.
5. **Deep Atmosphere** — The sky and environment change realistically throughout the day, creating a sense of time, place, and presence.
