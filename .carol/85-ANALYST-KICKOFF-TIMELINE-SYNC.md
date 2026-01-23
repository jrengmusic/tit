# Session 85: Background Timeline Sync

**Role:** ANALYST
**Agent:** Amp (Claude Sonnet 4)
**Date:** 2026-01-23

**Depends on:** Session 84 (Footer Unification)

## Problem Statement

Timeline state detection (`DetectState()`) compares local HEAD against **cached local refs** (`refs/remotes/origin/<branch>`). These refs only update after `git fetch`. Current behavior:

1. App starts → `DetectState()` → Shows timeline from **stale refs**
2. Async `cmdFetchRemote()` runs in background
3. `RemoteFetchMsg` → Re-runs `DetectState()` → **Now accurate**

**Issue:** User briefly sees stale "In Sync" before it updates to "Behind" — no visual indication that sync is in progress.

## Proposed Solution

Implement **TimelineSync** — a background synchronization mechanism mirroring the existing cache building pattern:

1. **Non-blocking async fetch** — UI remains responsive
2. **Dimmed timeline display** during sync — indicates stale data
3. **Spinner animation** — visual feedback that sync is in progress
4. **Periodic refresh** — only triggers when `mode == ModeMenu`
5. **On-demand re-sync** — user can force refresh via menu or shortcut

## Design Pattern (Mirrors Cache Building)

### New Types (messages.go)

```go
// TimelineSyncMsg signals completion of background timeline sync
type TimelineSyncMsg struct {
    Success  bool
    Error    string
    Timeline git.Timeline  // Updated timeline state
    Ahead    int
    Behind   int
}

// TimelineSyncTickMsg triggers periodic sync while in menu mode
type TimelineSyncTickMsg struct{}
```

### New Application Fields (app.go)

```go
// Timeline sync state
timelineSyncInProgress bool          // True while fetch is running
timelineSyncLastUpdate time.Time     // Last successful sync timestamp
timelineSyncInterval   time.Duration // Default: 60 seconds
timelineSyncFrame      int           // Animation frame for spinner
```

### New Functions (timeline_sync.go — new file)

```go
// cmdTimelineSync runs git fetch in background and updates timeline
// CONTRACT: Only runs when HasRemote, returns immediately if NoRemote
func (a *Application) cmdTimelineSync() tea.Cmd

// cmdTimelineSyncTicker schedules periodic timeline sync
// CONTRACT: Only schedules next tick if mode == ModeMenu
func (a *Application) cmdTimelineSyncTicker() tea.Cmd

// shouldRunTimelineSync checks if sync should run
// Returns false if: NoRemote, sync in progress, or recently synced
func (a *Application) shouldRunTimelineSync() bool
```

### Sync Flow

```
App Init (HasRemote)
    │
    ├─► timelineSyncInProgress = true
    ├─► cmdTimelineSync() — async fetch
    └─► cmdTimelineSyncTicker() — schedules refresh ticks
            │
            ▼
    [Every 100ms while timelineSyncInProgress]
        │
        ├─► TimelineSyncTickMsg received
        ├─► If mode != ModeMenu → no-op (don't update UI)
        ├─► If mode == ModeMenu → increment timelineSyncFrame, regenerate header
        └─► Schedule next tick
            │
            ▼
    [Fetch completes]
        │
        ├─► TimelineSyncMsg received
        ├─► timelineSyncInProgress = false
        ├─► DetectState() — refresh git state
        ├─► timelineSyncLastUpdate = time.Now()
        └─► If mode == ModeMenu → schedule next sync after interval
```

### Periodic Sync (Menu Mode Only)

```
[In ModeMenu, after timelineSyncInterval elapsed]
    │
    ├─► shouldRunTimelineSync() == true
    ├─► timelineSyncInProgress = true
    └─► cmdTimelineSync() — async fetch
```

**Key Constraint:** Periodic sync ONLY triggers when `mode == ModeMenu`. Other modes (History, Input, Console) do not trigger syncs — they use cached state.

### Header Rendering Changes (ui/header.go)

```go
// When timelineSyncInProgress == true:
// - Timeline label shows spinner: "🔄 Syncing..." or "⏳ Checking..."
// - Timeline description dimmed or shows "Checking remote..."
// - After sync: normal timeline display

func (hs *HeaderState) TimelineLabel(syncInProgress bool, frame int) string {
    if syncInProgress {
        spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
        return spinnerFrames[frame % len(spinnerFrames)] + " Syncing"
    }
    return hs.TimelineEmoji + " " + hs.TimelineLabel
}
```

### Menu Item Integration

Timeline-dependent menu items (Push, Pull, Force Push) should reflect sync state:

```go
// In menu_items.go, when generating timeline-dependent items:
if a.timelineSyncInProgress {
    // Show items as dimmed with "(syncing...)" hint
    // Action still works (uses cached state) but hint indicates possible staleness
}
```

## Implementation Phases

### Phase 1: Core Sync Infrastructure
**Files:** `messages.go`, `app.go`, `timeline_sync.go` (new)
- Add message types and application fields
- Implement `cmdTimelineSync()` and `cmdTimelineSyncTicker()`
- Handle `TimelineSyncMsg` in Update()
- Trigger sync on Init() when HasRemote

### Phase 2: Header Visual Feedback
**Files:** `ui/header.go`, `app.go`
- Pass `timelineSyncInProgress` and `timelineSyncFrame` to header
- Render spinner when sync in progress
- Dim timeline description during sync

### Phase 3: Periodic Refresh
**Files:** `timeline_sync.go`
- Implement periodic sync scheduling (default: 60s interval)
- Only schedule when returning to ModeMenu
- Add `shouldRunTimelineSync()` guard

### Phase 4: Menu Integration (Optional)
**Files:** `menu_items.go`
- Add "(syncing...)" hint to timeline-dependent items
- Consider adding "Refresh Timeline" menu item for manual sync

## Constants (SSOT — ui/sizing.go or new constants file)

```go
const (
    TimelineSyncInterval    = 60 * time.Second  // Periodic sync interval
    TimelineSyncTickRate    = 100 * time.Millisecond  // Animation refresh rate
)
```

## Success Criteria

1. ✅ Timeline shows spinner during initial sync on startup
2. ✅ Spinner animation updates every 100ms (when in ModeMenu)
3. ✅ Timeline updates to accurate state after fetch completes
4. ✅ Periodic sync runs every 60s while in ModeMenu
5. ✅ No sync activity when in other modes (History, Input, Console)
6. ✅ No UI blocking during fetch — remains fully responsive
7. ✅ Clean build with `./build.sh`

## Files to Create

- `internal/app/timeline_sync.go` — Core sync logic

## Files to Modify

- `internal/app/messages.go` — Add `TimelineSyncMsg`, `TimelineSyncTickMsg`
- `internal/app/app.go` — Add fields, Init trigger, Update handler
- `internal/ui/header.go` — Spinner rendering, dimmed state
- `internal/ui/sizing.go` — Sync interval constants (or new constants file)

## Risk Assessment

**Low Risk:**
- Pattern mirrors existing cache building (proven approach)
- Non-blocking async (no UI freeze)
- Graceful degradation (stale data still usable)

**Medium Risk:**
- Race condition if user navigates away during sync (mitigated: sync completes, state updated on next menu entry)
- Network timeout handling (mitigated: fetch has default git timeout)

## Out of Scope

- Refresh on focus (terminal focus detection)
- Push/pull integration (sync automatically triggers before these ops)
- Offline detection (deferred to future session)

---

**End of Kickoff Plan**

Ready for SCAFFOLDER to implement Phase 1.
