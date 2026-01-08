# History & File(s) History Implementation Plan
**Date:** 2026-01-06  
**Status:** Deep Analysis Phase

---

## Executive Summary

This document outlines the **porting/implementation of History and File(s) History features** from old-tit to the new TIT project. The implementation differs critically in one aspect: **old-tit's cherry-pick feature is replaced with Time Travel** per the new SPEC.md.

**Key Complexity Areas Identified:**
1. **Caching System** - Two separate caches (History metadata vs File(s) diffs)
2. **Mode Architecture** - Two new AppModes required (ModeHistory, ModeFileHistory)
3. **Thread Safety** - Mutex-protected pre-loading during app startup
4. **Menu State Management** - Menus disabled until cache ready
5. **State-Dependent Rendering** - Diff display changes based on WorkingTree (Clean vs Modified)
6. **Post-Operation Cache Invalidation** - Cache must refresh after commits
7. **Time Travel Integration** - Enter time travel from History mode instead of cherry-pick

---

## Phase Overview

| Phase | Task | Estimated Lines | Dependencies |
|-------|------|------------------|---|
| **1** | Infrastructure & UI Types | 500 | git/types.go |
| **2** | History Cache System | 800 | Phase 1 |
| **3** | History UI & Rendering | 600 | Phase 1, 2 |
| **4** | History Mode Handlers & Menu | 700 | Phase 1, 2, 3 |
| **5** | File(s) History UI & Rendering | 700 | Phase 1, 2, 3 |
| **6** | File(s) History Handlers & Menu | 600 | Phase 1, 2, 3, 5 |
| **7** | Time Travel (6 sub-phases) | 950 | Phase 4, 6, TIME-TRAVEL-IMPLEMENTATION-PLAN.md |
| **8** | Cache Invalidation & Integration | 400 | All phases |
| **9** | Testing & Verification | 300 | All phases |
| **Total** | | ~5,650 | Sequential phases |

---

## Detailed Analysis

### 1. What We're Porting From Old-TIT

**Files to study:**
- `internal/ui/history.go` (416 lines) - History split-pane rendering
- `internal/ui/filehistory.go` (570 lines) - File(s) History 3-pane rendering
- `internal/ui/listpane.go` (230 lines) - Shared list component
- `internal/ui/diffpane.go` (480 lines) - Shared diff component
- `internal/app/modehandlers_history.go` (147 lines) - History input handlers
- `internal/app/modehandlers_filehistory.go` (205 lines) - File(s) History input handlers
- `internal/app/cachepreload.go` (235 lines) - Cache pre-loading orchestration
- `internal/app/cache.go` (130 lines) - Cache manager abstraction

**What works in old-tit (will adapt):**
- ✅ Split-pane and multi-pane layouts (Bubble Tea + Lip Gloss)
- ✅ Thread-safe cache pre-loading with mutex protection
- ✅ Cache key scheme for diffs (hash:path:version)
- ✅ Scroll offset management per pane
- ✅ Focus cycling (TAB key) between panes
- ✅ Menu disabling until cache ready

**What we're replacing (new TIT design):**
- ❌ Cherry-pick confirmation flow → Time Travel workflow
- ❌ Mode-specific cache progress messages → Integrated into app.Update()
- ❌ Cache manager abstraction → Simpler goroutine orchestration
- ❌ ModeInput (deprecated) → Direct mode-specific handlers

---

### 2. New TIT Design Constraints

**From ARCHITECTURE.md:**
- **Single-threaded UI:** All `Application` mutations on Bubble Tea event loop
- **Worker threads:** Git commands in goroutines, results via immutable `tea.Msg`
- **OutputBuffer:** Thread-safe ring buffer for streaming git output
- **Key Handler Registry:** Built once at app init, cached: `map[AppMode]map[string]KeyHandler`
- **State Detection:** `git.DetectState()` for current state, no config file
- **Mode as SSOT:** Current `AppMode` determines rendering and keyboard routing

**From SPEC.md:**
- **State = (WorkingTree, Timeline, Operation, Remote)**
- **Operation State Priority:** If `Operation = TimeTraveling`, show ONLY time travel menu
- **Time Travel §9:** Replaces cherry-pick
   - Enter: From History mode via ENTER on commit + confirmation (dirty protocol if needed)
   - While traveling: Browse commits, view diffs via File History, make local changes
   - Exit: Merge back (with conflict handling) or return (discard changes)

---

### 3. Architecture Differences

#### Old-TIT Approach
```
Application {
    historyState          *ui.HistoryState
    historyCurrentDetails *git.CommitDetails
    fileHistoryState      *ui.FileHistoryState
    
    historyCacheMutex     sync.Mutex
    cacheMutex            sync.Mutex
    cacheProgress         chan tea.Msg    // For progress updates
}

// Cache pre-loading via separate goroutines
go a.preCacheDiffsConcurrent()
go a.preCacheHistoryMetadata()
```

#### New TIT Approach (Will Use)
```
Application {
    // Mode and state (existing)
    mode AppMode
    gitState git.State
    
    // History state (NEW - Phase 1)
    historyState *HistoryState
    fileHistoryState *FileHistoryState
    
    // Caches (NEW - Phase 2)
    historyMetadataCache map[string]*git.CommitDetails
    fileHistoryDiffCache map[string]string
    fileHistoryFilesCache map[string][]FileInfo
    
    // Cache orchestration (NEW - Phase 2)
    cacheLoadingStarted bool  // Guard against re-preloading
    cacheMetadata       bool  // Track what's been loaded
    cacheDiffs          bool
    
    // No separate channels needed; cache progress via state
}

// On app init: detectState() → if Normal/etc, preload caches
// Cache loading happens in parallel goroutines, results in app.Update()
```

---

### 4. Critical Implementation Decisions

#### Decision 1: Cache Pre-Loading Timing
**Question:** When should we start cache pre-loading?

**Answer (Decided):**
- On app initialization, after git state detection
- Only if `gitState.Operation == Normal` (not in merge/rebase/conflict)
- Only if not in detached HEAD (would mean already time traveling)
- Start goroutines to fetch commits (parallel, non-blocking)
- Menu is created/updated BEFORE caches are ready (disabled items show progress)
- Once cache ready, UI can select History menu items

**Code Location:** `internal/app/app.go` in `New()` function after `detectState()`

---

#### Decision 2: Two Separate Cache Systems
**Question:** Should History and File(s) History share one cache?

**Answer (No - Different purposes):**

**History Cache:**
- Maps: `hash → CommitDetails` (author, date, message)
- Size: ~30 commits × 1KB = ~30KB
- Used by: History mode for display
- Pre-load: Yes, on app init

**File(s) History Cache:**
- Maps: `hash → []FileInfo` (file list per commit)
- Maps: `hash:path:version → diff` (diff content, state-dependent)
- Size: ~30 commits × 10 files × 3KB diff = ~900KB
- Used by: File(s) History mode, diff rendering
- Pre-load: Yes, but selective (skip >100 file commits for perf)

**Reason for separation:**
- History is lightweight, always pre-load
- File(s) diffs are heavy, need throttling + progress feedback
- Different invalidation strategies post-commit

---

#### Decision 3: Menu Disabling Until Cache Ready
**Question:** Should History menu items be disabled until cache is ready?

**Answer (Yes, per old-tit pattern):**

**Mechanism:**
```go
type MenuItem struct {
    ID      string
    Enabled bool  // NEW: can set to false
    Label   string
    Hint    string
}

// In menu.go:
func (a *Application) menuNormal() []MenuItem {
    items := []MenuItem{
        {
            ID:      "history",
            Label:   "📜 History",
            Hint:    "Browse commit history",
            Enabled: a.cacheMetadata,  // Disabled until metadata cache ready
        },
        {
            ID:      "file_history",
            Label:   "📁 File(s) History",
            Hint:    "Browse file changes",
            Enabled: a.cacheDiffs,  // Disabled until diffs cache ready
        },
    }
    return items
}
```

**In rendering.go:** Grayed-out menu items with "(Loading...)" hint

---

#### Decision 4: State-Dependent Diff Display
**Question:** How do we handle diffs changing based on WorkingTree state?

**Answer (Per old-tit spec §15):**

When rendering File(s) History diff:
- **WorkingTree = Clean:** Show `commit vs parent` (what did this commit do?)
- **WorkingTree = Modified:** Show `commit vs working tree` (how does WIP compare?)

**Implementation:**
```go
// In filehistorystate.DiffCache key:
cacheKeySuffix := "parent"  // Default for Clean
if a.gitState.WorkingTree == git.Dirty {
    cacheKeySuffix = "wip"  // Modified: show vs working tree
}
cacheKey := fmt.Sprintf("%s:%s:%s", commit.Hash, file.Path, cacheKeySuffix)
diffContent := a.fileHistoryState.DiffCache[cacheKey]
```

**Pre-loading:** Cache BOTH versions for each file during startup

---

#### Decision 5: Cache Invalidation After Commit
**Question:** How do we refresh caches after user commits?

**Answer (Planned for Phase 8):**

```go
// In githandlers.go, after commit succeeds:
case "commit":
    if msg.Success {
        // Refresh git state
        a.gitState = git.DetectState()
        
        // Invalidate and refresh caches
        a.invalidateHistoryCaches()
        a.preloadHistoryCaches()  // Async goroutines
        
        // Return to menu (caches will be ready next update cycle)
        a.mode = ModeMenu
    }
```

---

### 5. Data Structures (Phase 1)

**New types to add to `internal/app/app.go`:**

```go
// HistoryState represents the state of the history browser
type HistoryState struct {
    Commits       []git.CommitInfo  // List of commits (hash, time, subject)
    SelectedIdx   int               // Currently selected commit (0-indexed)
    PaneFocused   bool              // true = list pane, false = details pane
    ListScrollOff int               // Scroll offset for commit list
    DetailsOff    int               // Scroll offset for details pane
}

// FileHistoryState represents the state of the file(s) history browser
type FileHistoryState struct {
    Commits              []git.CommitInfo  // List of commits
    Files                []git.FileInfo    // Files in selected commit
    SelectedCommitIdx    int               // Currently selected commit
    SelectedFileIdx      int               // Currently selected file
    FocusedPane          FileHistoryPane   // Which pane has focus
    CommitsScrollOff     int               // Scroll offset for commits list
    FilesScrollOff       int               // Scroll offset for files list
    DiffScrollOff        int               // Scroll offset for diff
}

type FileHistoryPane int
const (
    PaneCommits FileHistoryPane = iota
    PaneFiles
    PaneDiff
)

// CommitInfo: add to internal/git/types.go
type CommitInfo struct {
    Hash    string    // Full commit hash
    Subject string    // Commit message first line
    Time    time.Time // Commit author date
}

// CommitDetails: add to internal/git/types.go
type CommitDetails struct {
    Author  string
    Date    string
    Message string
}

// FileInfo: add to internal/git/types.go
type FileInfo struct {
    Path     string  // File path
    Status   string  // M, A, D, R, etc.
}
```

---

### 6. Git Command Requirements

**New git.Execute() calls needed:**

1. **Fetch recent commits (History):**
   ```bash
   git log --pretty=%H%n%s%n%ai -N  # Hash, subject, date
   ```

2. **Fetch commit metadata (History details):**
   ```bash
   git show -s --pretty=%aN%n%aD%n%B <hash>  # Author, date, message
   ```

3. **Fetch files in commit (File(s) History):**
   ```bash
   git show --name-only --pretty= <hash>  # File list
   ```

4. **Fetch diff for file in commit vs parent (File(s) History, Clean):**
   ```bash
   git diff <hash>^ <hash> -- <path>
   ```

5. **Fetch diff for file in commit vs working tree (File(s) History, Modified):**
   ```bash
   git diff <hash> -- <path>
   ```

6. **Checkout commit (Time Travel entry):**
   ```bash
   echo <original_branch> > .git/TIT_TIME_TRAVEL
   git checkout <commit_hash>
   ```

7. **Get original branch (Time Travel detection):**
   ```bash
   cat .git/TIT_TIME_TRAVEL
   ```

---

### 7. UI Components (Phase 3 & 5)

**New rendering functions needed in `internal/ui/`:**

1. **History Pane:** (`history.go` - NEW)
   - `RenderHistorySplitPane(state *HistoryState, theme *Theme, ...) string`
   - Two-pane layout: commit list (left) + details (right)
   - Status bar with keyboard hints

2. **File(s) History Pane:** (`filehistory.go` - NEW)
   - `RenderFileHistorySplitPane(state *FileHistoryState, theme *Theme, ...) string`
   - Three-pane layout: commits (top-left) + files (top-right) + diff (bottom)
   - Status bar with keyboard hints

3. **Reuse from old-tit:**
   - `ListPane` (already exists in new TIT) - used for commit & file lists
   - `DiffPane` (already exists in new TIT) - used for diff display

---

### 8. Keyboard Handlers (Phase 4 & 6)

**ModeHistory handlers:**
- `↑/k` - Navigate commit list (history pane focused)
- `↓/j` - Navigate commit list (history pane focused)
- `↑/k` - Scroll details up (details pane focused)
- `↓/j` - Scroll details down (details pane focused)
- `TAB` - Switch focus between panes
- `ENTER` - Enter time travel (if clean working tree, show confirmation)
- `ESC` - Return to menu

**ModeFileHistory handlers:**
- `↑/k` - Navigate up in focused pane (commits/files/diff)
- `↓/j` - Navigate down in focused pane
- `TAB` - Cycle focus: Commits → Files → Diff → Commits
- `SPACE/ENTER` - No action (file history is read-only view)
- `y` - Copy selected lines from diff (if diff pane focused)
- `v` - Toggle visual mode in diff (if diff pane focused)
- `ESC` - Return to menu

---

### 9. Menu Items (Phase 4 & 6)

**In menuNormal() - add to existing menu:**

```go
{
    ID:      "history",
    Label:   "📜 History",
    Hint:    "Browse commit history (time travel)",
    Enabled: a.cacheMetadata,
},
{
    ID:      "file_history",
    Label:   "📁 File(s) History",
    Hint:    "Browse file changes across commits",
    Enabled: a.cacheDiffs,
},
```

**In dispatchers.go - add new dispatchers:**
- `dispatchHistory()` - Set mode to ModeHistory, return
- `dispatchFileHistory()` - Set mode to ModeFileHistory, return

---

### 10. Time Travel Integration (Phase 7)

**Differences from cherry-pick:**

**Old-TIT Cherry-Pick Flow:**
1. User in History mode, selects commit
2. Presses ENTER → shows confirmation "Apply this commit?"
3. User confirms → `git cherry-pick <hash>`
4. If conflicts → ModeConflictResolve
5. If success → return to menu

**New TIT Time Travel Flow:**
1. User in History mode, selects commit
2. Presses ENTER → shows confirmation "Time travel to this commit?"
3. Confirmation shows read-only warning + what they can do
4. User confirms → stash if dirty + `git checkout <hash>` + save branch info
5. Mode changes to `TimeTraveling` (per SPEC.md § 9.1-9.3)
6. Menu shows ONLY time travel options (jump, view diff, merge back, return)
7. User can:
   - Make local changes, browse, test
   - Press "Merge back" to integrate into original branch
   - Press "Return" to discard and go back

**Implementation location:**
- Time travel entry: `internal/app/handlers.go` (History mode enter handler)
- Time travel menu: `internal/app/menu.go` (add `menuTimeTraveling()`)
- Time travel handlers: `internal/app/handlers.go` (merge, return, jump, etc.)

---

### 11. Testing Strategy (Phase 9)

**Manual test scenarios (no automated tests per old-tit):**

1. **History Mode Basic:**
   - [ ] History menu item appears + enabled after cache loads
   - [ ] Select History → shows 30 recent commits
   - [ ] Up/Down navigates commits
   - [ ] TAB switches between list & details panes
   - [ ] Details pane shows author, date, message
   - [ ] ESC returns to menu

2. **File(s) History Mode Basic:**
   - [ ] File(s) History menu item appears + enabled after cache loads
   - [ ] Select File(s) History → shows commits + files + diff
   - [ ] Up/Down/TAB navigate and cycle panes
   - [ ] Selecting different commit updates file list instantly (cached)
   - [ ] Selecting different file updates diff instantly (cached)
   - [ ] ESC returns to menu

3. **Time Travel Integration:**
   - [ ] In History mode, press ENTER on clean working tree → confirmation
   - [ ] Confirmation dialog appears with read-only warning
   - [ ] Confirm → enters time travel mode (detached HEAD)
   - [ ] Git state detects `Operation = TimeTraveling`
   - [ ] Menu shows ONLY time travel options
   - [ ] Can browse history while time traveling
   - [ ] Can make local changes (tracked as Modified)
   - [ ] Cannot commit (no menu option)
   - [ ] Can merge back to original branch
   - [ ] Can return without merging

4. **Cache Invalidation:**
   - [ ] Make a commit in normal mode
   - [ ] Commit succeeds → caches invalidated
   - [ ] Caches reload (History shows new commit at top)
   - [ ] File(s) History shows new commit in list

5. **Dirty Working Tree:**
   - [ ] History mode works with dirty tree (reads-only)
   - [ ] In History, try ENTER with dirty tree → shows dirty operation warning
   - [ ] Allows stash + time travel

6. **Edge Cases:**
   - [ ] History with <30 commits (cache loads all available)
   - [ ] File(s) History with >100 files in single commit (diff not cached)
   - [ ] Navigation while cache is loading
   - [ ] ESC during time travel (returns to branch)

---

### 12. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|-----------|
| **Cache mutex deadlock** | Medium | High | Careful lock scope, test concurrent ops |
| **Memory usage from large diffs** | Low | Medium | Skip caching >100-file commits |
| **Race condition during cache load** | Medium | High | Immutable cache, snapshot-based rendering |
| **Time Travel state confusion** | Medium | High | Clear menu state, strong typing |
| **Diff version mismatch** | Low | Medium | Consistent cache key format |

---

## Phase-by-Phase Breakdown

### Phase 1: Infrastructure & UI Types

**Files to modify/create:**
1. `internal/git/types.go` - Add `CommitInfo`, `CommitDetails`, `FileInfo` types
2. `internal/app/app.go` - Add `HistoryState`, `FileHistoryState` fields, init in `New()`
3. `internal/app/modes.go` - Add `ModeHistory`, `ModeFileHistory` enum values

**Acceptance Criteria:**
- [ ] New types compile
- [ ] App init doesn't break
- [ ] No new functionality visible yet

---

### Phase 2: History Cache System

**Files to create:**
1. `internal/app/historycache.go` - Cache orchestration
   - `preloadHistoryMetadata()` - Fetch 30 commits + metadata
   - `preloadFileHistoryDiffs()` - Fetch diffs for 30 commits
   - `invalidateHistoryCaches()` - Clear all caches

**Files to modify:**
1. `internal/app/app.go` - Call pre-load on init, add cache fields

**Acceptance Criteria:**
- [ ] App startup triggers cache pre-loading (no UI change yet)
- [ ] Cache fields populated after ~2-3 seconds
- [ ] No memory leaks (verified with `pprof`)
- [ ] Thread-safe (no race detector warnings)

---

### Phase 3: History UI & Rendering

**Files to create:**
1. `internal/ui/history.go` - History split-pane rendering
   - `RenderHistorySplitPane()` function
   - `HistoryState` struct definition

**Files to modify:**
1. `internal/ui/layout.go` - Add History rendering case

**Acceptance Criteria:**
- [ ] History pane renders (commits list + details)
- [ ] Focus indicator shows correctly (left vs right pane)
- [ ] Scrolling works in both panes
- [ ] Keyboard hint bar displays correctly

---

### Phase 4: History Mode Handlers & Menu

**Files to modify:**
1. `internal/app/app.go` - Add `ModeHistory` to mode registry
2. `internal/app/menu.go` - Add History menu item + `menuHistory()` function
3. `internal/app/handlers.go` - Add History mode handlers
   - `handleHistoryUp/Down`
   - `handleHistoryTab`
   - `handleHistoryEnter` (time travel entry - TBD in Phase 7)
   - `handleHistoryEsc`

**Acceptance Criteria:**
- [ ] Select History from menu → ModeHistory entered
- [ ] Arrow keys navigate commits
- [ ] TAB switches panes
- [ ] ESC returns to menu
- [ ] ENTER not yet functional (Phase 7)

---

### Phase 5: File(s) History UI & Rendering

**Files to create:**
1. `internal/ui/filehistory.go` - File(s) History multi-pane rendering
   - `RenderFileHistorySplitPane()` function
   - `FileHistoryState` struct definition

**Files to modify:**
1. `internal/ui/layout.go` - Add File(s) History rendering case

**Acceptance Criteria:**
- [ ] File(s) History pane renders (3-pane layout)
- [ ] Selecting commit updates file list (from cache, instant)
- [ ] Selecting file updates diff (from cache, instant)
- [ ] Focus indicator cycles through all 3 panes
- [ ] Scrolling works independently in all panes

---

### Phase 6: File(s) History Handlers & Menu

**Files to modify:**
1. `internal/app/app.go` - Add `ModeFileHistory` to mode registry
2. `internal/app/menu.go` - Add File(s) History menu item + `menuFileHistory()` function
3. `internal/app/handlers.go` - Add File(s) History mode handlers
   - `handleFileHistoryUp/Down`
   - `handleFileHistoryTab`
   - `handleFileHistoryCopy` (y key)
   - `handleFileHistoryVisualMode` (v key)
   - `handleFileHistoryEsc`

**Acceptance Criteria:**
- [ ] Select File(s) History from menu → ModeFileHistory entered
- [ ] Navigation works in all 3 panes
- [ ] TAB cycles through panes correctly
- [ ] Copy (y) works in diff pane
- [ ] Visual mode (v) toggles in diff pane
- [ ] ESC returns to menu

---

### Phase 7: Time Travel Integration

**Scope:** Full time travel implementation with clean/dirty tree handling, merge with conflict detection, and safe state preservation.

**Implementation:** See dedicated **TIME-TRAVEL-IMPLEMENTATION-PLAN.md** (6 incremental phases, ~950 lines).

**High-level flow:**
1. **Phase 1:** Basic time travel (clean tree) - Load `TimeTravelInfo`, show menu
2. **Phase 2:** Dirty tree handling - Dirty operation protocol + stash
3. **Phase 3:** Merge back (no conflicts) - Stash T1, merge, apply stashes in sequence
4. **Phase 4:** Merge conflicts + resolution - Sequential ConflictResolver for 3 possible conflicts
5. **Phase 5:** Browse while time traveling - Jump to different commits
6. **Phase 6:** Return without merge - Discard changes, restore original stash

**Key decisions:**
- ✅ Clean tree only for entry (dirty tree → dirty protocol)
- ✅ 3 sequential conflict points possible (ABC→main, T1, S1)
- ✅ ESC at any point restores original state
- ✅ Menu shows 3 items only: Browse history, Merge back, Return
- ✅ No "View diff" menu item (use File History instead)

**Acceptance Criteria:**
- [ ] All 6 phases implemented and tested per TIME-TRAVEL-TESTING-CHECKLIST.md
- [ ] 30+ test scenarios pass (Phase 1-6, edge cases, full flows, regressions)
- [ ] No work lost on cancel/ESC at any step
- [ ] Sequential conflicts handled correctly
- [ ] Original stash always preserved/restored

---

### Phase 8: Cache Invalidation & Integration

**Files to modify:**
1. `internal/app/historycache.go` - Add invalidation functions
2. `internal/app/githandlers.go` - Call cache invalidation after commit/merge/time-travel-merge
3. `internal/app/rendering.go` - Show "(Loading...)" hint for disabled menu items

**Acceptance Criteria:**
- [ ] After commit, caches are invalidated
- [ ] Caches re-load automatically
- [ ] History shows new commit at top after reload
- [ ] No broken state during cache reload
- [ ] Disabled menu items show progress indication

---

### Phase 9: Testing & Verification

**Manual test checklist (from Section 11):**
- [ ] All 6 test categories pass
- [ ] All 26 test items verified

**Code review checklist:**
- [ ] No circular dependencies
- [ ] Thread-safe (mutex usage correct)
- [ ] No goroutine leaks
- [ ] All modes properly initialized
- [ ] All handlers registered correctly

---

## Integration with Existing Code

### Existing Components We'll Use

1. **`internal/ui/listpane.go`** - Reuse for history/file lists
   - Already exists and works
   - No changes needed

2. **`internal/ui/diffpane.go`** - Reuse for diff display
   - Already exists and works
   - Minor: adapt for "commit vs parent" vs "commit vs WIP" selection

3. **`internal/ui/theme.go`** - Use existing theme system
   - Add any missing colors if needed

4. **`internal/git/execute.go`** - Add new git command helpers
   - `GetRecentCommits()` - Fetch commit list
   - `GetCommitDetails()` - Fetch metadata
   - `GetFilesInCommit()` - Fetch file list
   - `GetCommitDiff()` - Fetch diff content

### No Breaking Changes

- ✅ Existing menu system continues to work
- ✅ Existing git state detection unchanged
- ✅ Existing conflict resolution unaffected
- ✅ Existing Time Travel (Phase 7) is NEW, doesn't replace anything
- ✅ Old cherry-pick code doesn't exist in new TIT (clean slate)

---

## Questions to Clarify Before Implementation

1. **Cache pre-load limit:** Should we always pre-load 30 commits or make it configurable?
2. **Diff cache threshold:** Is >100 files per commit the right threshold to skip caching?
3. **Time Travel merge strategy:** When merging time-travel changes back, should we:
   - [ ] Create a merge commit (default)
   - [ ] Rebase onto original branch
   - [ ] Cherry-pick from time-travel commit
4. **History depth:** Should History show all commits or limit to N most recent?
5. **Cache invalidation timing:** After commit, should caches reload immediately or on next menu entry?

---

## Summary of Differences from Old-TIT

| Aspect | Old-TIT | New TIT |
|--------|---------|---------|
| **Cherry-Pick** | ENTER in History → cherry-pick confirmation | ❌ Removed |
| **Time Travel** | ❌ Not implemented | ✅ ENTER in History → time travel |
| **Cache Architecture** | Separate `CacheManager` abstraction | Simpler goroutine-based |
| **Progress Feedback** | Separate progress channels | Integrated into Update() |
| **Mode Registry** | Individual mode handlers | Centralized `keyHandlers` map |
| **Menu Enabling** | Explicit cache checks | `Enabled` field in MenuItem |

---

## Success Criteria

**At end of Phase 9:**

1. ✅ History mode shows last 30 commits with metadata
2. ✅ File(s) History shows commits, changed files, and diffs
3. ✅ Time Travel works from History mode (read-only exploration)
4. ✅ Can merge changes back from time-travel mode
5. ✅ Caches refresh after commits
6. ✅ All keyboard shortcuts work as documented
7. ✅ No race conditions or deadlocks
8. ✅ All manual tests pass (26 items)
9. ✅ Code follows new TIT architecture patterns
10. ✅ No existing functionality broken

---

**Document Status:** Ready for Phase 1 Implementation  
**Next Step:** Implement Phase 1 (Infrastructure & UI Types) with code review
