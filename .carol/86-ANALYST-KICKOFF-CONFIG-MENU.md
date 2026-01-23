# Session 86: Config Menu & Preferences

**Role:** ANALYST
**Agent:** Amp (Claude Sonnet 4)
**Date:** 2026-01-23

**Depends on:** Session 84 (Footer Unification), Session 85 (Timeline Sync)

## Overview

Add new `ModeConfig` menu accessible via `/` shortcut from main menu. Provides repository configuration (remote, branch) and user preferences (auto-update, themes).

---

## Config Menu Structure

**Shortcut:** `/` from ModeMenu

**Dynamic Menu Items (based on git state):**

| Condition | Menu Items |
|-----------|------------|
| NoRemote | Add Remote |
| HasRemote | Switch Remote |
| — | ───────── (separator) |
| HasRemote | Remove Remote |
| NoRemote | Toggle Auto Update *(disabled)* |
| HasRemote | Toggle Auto Update |
| Always | Switch Branch |
| — | ───────── (separator) |
| Always | Preferences |

---

## Component 1: Remote Operations

### Add Remote (NoRemote)
- Flow: ModeInput → prompt URL → `git remote add origin <url>` → fetch → DetectState
- Identical to existing "Add Remote" from main menu
- After success: menu shows "Switch Remote" instead

### Switch Remote (HasRemote)
- Flow: ModeInput → prompt URL → `git remote set-url origin <url>` → fetch → DetectState
- Same UI as Add Remote, different git command

### Remove Remote (HasRemote)
- Flow: Confirmation dialog → `git remote remove origin` → DetectState
- After success: menu shows "Add Remote" instead

---

## Component 2: Toggle Auto Update

**Behavior:** Single toggle action (no sub-menu)

```
Toggle Auto Update    ON   →   Toggle Auto Update    OFF
```

- `Enter` on menu item → toggle value → write config → apply immediately
- When OFF: TimelineSync disabled (no background fetch)
- When ON: TimelineSync enabled with configured interval
- Disabled when NoRemote (no remote to sync with)

---

## Component 3: Switch Branch

**New Mode:** `ModeBranchPicker`

**UI Layout:** 2-pane (identical to History)

```
┌─────────────────────────────────────────────────────────────────────┐
│  BRANCHES                                                           │
├──────────────────────────────┬──────────────────────────────────────┤
│ ● main                       │  Branch: main                        │
│   feature/config             │  Last Commit: 2 hours ago            │
│   feature/timeline-sync      │  Subject: fix: timeline sync issue   │
│   hotfix/crash               │  Author: jreng <jreng@example.com>   │
│                              │  Tracking: origin/main ↑2            │
│                              │                                      │
├──────────────────────────────┴──────────────────────────────────────┤
│  Enter select · Esc back · Ctrl+C quit                              │
└─────────────────────────────────────────────────────────────────────┘
```

**Left Pane (Branch List):**
- `●` marks current branch
- Local branches only (no remote-only branches)
- Sorted: current first, then alphabetical

**Right Pane (Branch Details):**
- Branch name
- Last commit relative time
- Last commit subject
- Author
- Tracking status (remote branch + ahead/behind, or "local only")

**Navigation:**
- `↑/↓` or `j/k` — move selection
- `Enter` — switch to selected branch
- `Esc` — back to config menu

**Switch Flow (Clean WorkingTree):**
```
Enter on branch
    │
    └─► git switch <branch> → DetectState → back to ModeMenu
```

**Switch Flow (Dirty WorkingTree):**
```
Enter on branch
    │
    └─► Prompt: "Commit changes" or "Switch anyway"
            │
            ├─► Commit → ModeInput (message) → commit → switch → DetectState
            │
            └─► Switch anyway → git stash → git switch → git stash pop
                    │
                    ├─► Success → DetectState → ModeMenu
                    └─► Conflict → ModeConflictResolver
```

**Caching (identical to History):**
- Preload branch metadata on entering ModeBranchPicker
- Cache: `branchMetadataCache map[string]*BranchDetails`
- Show loading spinner while building cache

---

## Component 4: Preferences Pane

**New Mode:** `ModePreferences`

**UI Layout:**

```
┌─────────────────────────────────────────────────────────────────────┐
│  PREFERENCES                                                        │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ▸ Auto-update Enabled      ON                                      │
│    Auto-update Interval     5 min                                   │
│    Theme                    dark                                    │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│  Space toggle · +/- interval · Esc save                             │
└─────────────────────────────────────────────────────────────────────┘
```

**Rows:**
1. **Auto-update Enabled** — `space` toggles ON/OFF
2. **Auto-update Interval** — `+`/`=` increase, `-` decrease (1-60 min)
3. **Theme** — `space` cycles through available themes

**Navigation:**
- `↑/↓` or `j/k` — move between rows
- `space` — toggle/cycle current row
- `+`/`=` — increase interval (row 2 only)
- `-` — decrease interval (row 2 only)
- `Esc` — save and return to config menu
- `Ctrl+C` — quit confirmation (reuse existing pattern)

**Behavior:**
- Changes apply immediately (hot reload)
- Changes write to config file immediately
- App reads from config (SSOT)

---

## Config File Infrastructure

**Path:** `~/.config/tit/config.toml`

**Format:** TOML (consistent with existing theme files)

**Schema:**
```toml
# TIT Configuration

[auto_update]
enabled = true
interval_minutes = 5

[appearance]
theme = "default"
```

**Startup Flow:**
```
App Init
    │
    └─► CheckConfigFile()
            │
            ├─► Exists + Valid → Load
            │
            ├─► Exists + Invalid → Log warning, create default, load
            │
            └─► Not exists → Create with defaults, load
```

**Default Values:**
```go
DefaultConfigTOML = `# TIT Configuration

[auto_update]
enabled = true
interval_minutes = 5

[appearance]
theme = "default"
`
```

**Package:** `internal/config/config.go`

**Dependencies:** `github.com/pelletier/go-toml/v2` (already in use for themes)

---

## New Types

### AppMode (modes.go)
```go
const (
    // ... existing modes
    ModeConfig        AppMode = "config"
    ModeBranchPicker  AppMode = "branch_picker"
    ModePreferences   AppMode = "preferences"
)
```

### Config Struct (config/config.go)
```go
type Config struct {
    AutoUpdate AutoUpdateConfig `toml:"auto_update"`
    Appearance AppearanceConfig `toml:"appearance"`
}

type AutoUpdateConfig struct {
    Enabled         bool `toml:"enabled"`
    IntervalMinutes int  `toml:"interval_minutes"`
}

type AppearanceConfig struct {
    Theme string `toml:"theme"`
}
```

### BranchDetails (git/branch.go — new)
```go
type BranchDetails struct {
    Name           string
    IsCurrent      bool
    LastCommitTime time.Time
    LastCommitHash string
    LastCommitSubj string
    Author         string
    TrackingRemote string    // e.g., "origin/main"
    Ahead          int
    Behind         int
}
```

### Application Fields (app.go)
```go
// Config state
appConfig         *config.Config    // Loaded from ~/.config/tit/config.yaml

// Config menu state
configMenuItems   []MenuItem
configSelectedIdx int

// Branch picker state
branchPickerState *ui.BranchPickerState

// Preferences state
preferencesState  *PreferencesState
```

---

## Implementation Phases

### Phase 1: Config Infrastructure
**Files:** `internal/config/config.go` (new), `go.mod`
- Config struct with YAML tags
- Load/Save functions
- Startup check (create default if missing)
- Add `gopkg.in/yaml.v3` dependency

### Phase 2: ModeConfig Menu
**Files:** `modes.go`, `app.go`, `handlers.go`, `menu_items.go`
- Add ModeConfig
- Generate config menu items (dynamic based on state)
- Handle `/` shortcut from ModeMenu
- Wire up navigation and item selection

### Phase 3: Remote Operations
**Files:** `operations.go`, `git_handlers.go`
- Implement Switch Remote (reuse Add Remote flow)
- Implement Remove Remote (with confirmation)
- Integrate with config menu

### Phase 4: ModePreferences
**Files:** `ui/preferences.go` (new), `app.go`, `handlers.go`
- PreferencesState struct
- Render function (3 rows with selection)
- Key handlers (space, +/-, up/down)
- Hot reload on change
- Write to config on change

### Phase 5: ModeBranchPicker
**Files:** `ui/branchpicker.go` (new), `git/branch.go` (new), `app.go`, `handlers.go`
- BranchPickerState struct (mirrors HistoryState)
- 2-pane layout (list + details)
- Branch metadata caching
- Switch flow with dirty handling

### Phase 6: Integration & Polish
**Files:** Various
- Wire up TimelineSync to respect config
- Theme switching integration
- Footer hints for all new modes
- Test all flows

---

## Files to Create

| File | Purpose |
|------|---------|
| `internal/config/config.go` | Config loading/saving, YAML schema |
| `internal/ui/preferences.go` | Preferences pane rendering |
| `internal/ui/branchpicker.go` | Branch picker 2-pane component |
| `internal/git/branch.go` | Branch listing and metadata |

## Files to Modify

| File | Changes |
|------|---------|
| `go.mod` | (no change - TOML already imported) |
| `internal/app/modes.go` | Add ModeConfig, ModeBranchPicker, ModePreferences |
| `internal/app/app.go` | Add config/state fields, Init loading, Update handlers |
| `internal/app/handlers.go` | Key handlers for new modes |
| `internal/app/menu_items.go` | Config menu generation |
| `internal/app/operations.go` | Switch/Remove remote commands |
| `internal/app/messages.go` | New message types if needed |

---

## Success Criteria

1. ✅ `/` opens config menu from main menu
2. ✅ Config menu shows correct items based on remote state
3. ✅ Add/Switch/Remove Remote work correctly
4. ✅ Toggle Auto Update toggles and persists
5. ✅ Branch picker shows all local branches with metadata
6. ✅ Branch switch works (clean and dirty)
7. ✅ Preferences pane allows editing all settings
8. ✅ Changes persist to `~/.config/tit/config.yaml`
9. ✅ Changes apply immediately (hot reload)
10. ✅ Footer shows correct hints in all modes
11. ✅ Ctrl+C confirmation works in all modes
12. ✅ Clean build with `./build.sh`

---

## Dependencies

- **Session 84 (TimelineSync):** Preferences reads `auto_update.enabled` and `interval_minutes` to control sync behavior
- **Existing theme system:** Theme switcher cycles through available themes

---

**End of Kickoff Plan**

Ready for SCAFFOLDER to implement Phase 1.
