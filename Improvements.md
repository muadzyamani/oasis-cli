# Improvements Plan

This plan details the removal of the tier feature and the redesign of the top bar in the Oasis CLI.

## Proposed Changes

### Feature Cleanup (Removing Tier)
We will completely delete the "growth tier" concept from the codebase:
- **Storage**: Remove the `Tier` field from `OasisState` in `internal/storage/json.go` and clean up `DefaultState()`.
- **Engine**: Remove `GetTierForMinutes` from `internal/engine/stats.go`.
- **UI & State**: Stop calculating and setting `Tier` inside `internal/ui/update.go`.
- **Testing**: Clean up unit tests in `internal/storage/json_test.go` and `internal/ui/update_test.go` to remove assertions and assignments relating to the tier.

### Top Bar & Navigation Redesign
- Remove the separate top header row completely.
- Position the navigation tabs (`Oasis | Stats | Settings`) and the "🔥 Streak" indicator on the same row.
- Push the Streak all the way to the right side of the row.
- Clean up unused style constants (`TitleStyle`, `TierStyle`) in `internal/ui/styles/colors.go`.
- Implement a grammatical fix so that when the streak is exactly 1, the text reads "1 day" instead of "1 days".

---

## Detailed File Changes

### storage

#### [MODIFY] [json.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/storage/json.go)
- Remove `Tier int` field from `OasisState` struct.
- Remove `Tier: 0` initialization in `DefaultState()`.

#### [MODIFY] [json_test.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/storage/json_test.go)
- Remove test initialization of `customState.Oasis.Tier = 3`.
- Remove assertions verifying `loadedState.Oasis.Tier != 3`.

---

### engine

#### [MODIFY] [stats.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/engine/stats.go)
- Delete the `GetTierForMinutes` function completely.

---

### ui

#### [MODIFY] [colors.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/ui/styles/colors.go)
- Delete unused `TitleStyle` and `TierStyle` variables.

#### [MODIFY] [update.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/ui/update.go)
- Remove the code updating `m.State.Oasis.Tier` after focus minutes completed.

#### [MODIFY] [update_test.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/ui/update_test.go)
- Remove assertions checking if `updated.State.Oasis.Tier` is upgraded on focus completion.

#### [MODIFY] [view.go](file:///c:/Users/muadz/Muadz/School%20Work/NUS/Others/Year%202%20Summer%20Break/oasis-cli/internal/ui/view.go)
- Merge header and navigation tabs into a single row, pushing "🔥 Streak" to the top right next to navigation tabs.
- Add singular/plural conditional for rendering "day" vs "days" based on the streak count.

---

## Verification Plan

### Automated Tests
- Run `& "C:\Program Files\Go\bin\go.exe" test ./...` to ensure all tests pass successfully.
