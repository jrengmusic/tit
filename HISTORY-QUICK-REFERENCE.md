# History & File(s) History - Quick Reference

## File Structure (What Gets Created/Modified)

### New Files to Create
```
internal/
├── app/
│   └── historycache.go          (NEW - Cache orchestration)
├── ui/
│   ├── history.go               (NEW - History pane rendering)
│   └── filehistory.go           (NEW - File(s) History pane rendering)
└── git/
    └── (git commands added to execute.go)
```

### Files to Modify
```
internal/
├── app/
│   ├── app.go                   (Add HistoryState, FileHistoryState fields)
│   ├── modes.go                 (Add ModeHistory, ModeFileHistory)
│   ├── menu.go                  (Add menu items + generators)
│   ├── handlers.go              (Add mode handlers)
│   ├── dispatchers.go           (Add dispatchers)
│   ├── messages.go              (Add messages if needed)
│   ├── githandlers.go           (Add result handlers + cache invalidation)
│   └── rendering.go             (Add Layout cases for History/FileHistory)
├── ui/
│   └── layout.go                (Add rendering calls for new modes)
└── git/
    └── execute.go               (Add git command helpers)
```

---

## Data Types Quick Reference

### In internal/git/types.go
```go
type CommitInfo struct {
    Hash    string    // Full commit hash
    Subject string    // Commit message first line
    Time    time.Time // Commit author date
}

type CommitDetails struct {
    Author  string    // Author name
    Date    string    // Formatted date
    Message string    // Full commit message
}

type FileInfo struct {
    Path   string  // File path
    Status string  // M, A, D, R, C, T, U
}
```

### In internal/app/app.go
```go
type HistoryState struct {
    Commits       []git.CommitInfo
    SelectedIdx   int
    PaneFocused   bool  // true = list, false = details
    ListScrollOff int
    DetailsOff    int
}

type FileHistoryState struct {
    Commits           []git.CommitInfo
    Files             []git.FileInfo
    SelectedCommitIdx int
    SelectedFileIdx   int
    FocusedPane       FileHistoryPane  // 0=Commits, 1=Files, 2=Diff
    CommitsScrollOff  int
    FilesScrollOff    int
    DiffScrollOff     int
}

type FileHistoryPane int

const (
    PaneCommits FileHistoryPane = iota
    PaneFiles
    PaneDiff
)
```

---

## Keyboard Handlers by Mode

### ModeHistory
| Key | Action | Pane |
|-----|--------|------|
| `↑` or `k` | Previous commit | List focused |
| `↓` or `j` | Next commit | List focused |
| `↑` or `k` | Scroll up | Details focused |
| `↓` or `j` | Scroll down | Details focused |
| `TAB` | Switch pane (list ↔ details) | Any |
| `ENTER` | Enter time travel + confirmation | List focused |
| `ESC` | Return to menu | Any |

### ModeFileHistory
| Key | Action | Pane |
|-----|--------|------|
| `↑` or `k` | Previous item | Any |
| `↓` or `j` | Next item | Any |
| `TAB` | Cycle pane (commits → files → diff → commits) | Any |
| `y` | Copy selected lines | Diff only |
| `v` | Toggle visual mode | Diff only |
| `ENTER` | No action | Any |
| `ESC` | Return to menu | Any |

### ModeTimeTraveling (Phase 7)
| Key | Action |
|-----|--------|
| History browser same as ModeHistory | |
| `m` | Merge changes back to original branch |
| `r` | Return to original branch (discard changes) |
| Other git operations | Disabled (menu shows only time travel options) |

---

## Cache Strategy

### History Metadata Cache
- **Key:** `hash` (full commit hash)
- **Value:** `*git.CommitDetails` (author, date, message)
- **Size:** ~1KB per commit × 30 commits = ~30KB
- **Preload:** On app init if `gitState.Operation == Normal`
- **Invalidate:** After commit/merge/time-travel-merge
- **Thread-safe:** `historyCacheMutex` (but simpler approach in new TIT)

### File(s) History File Cache
- **Key:** `hash` (full commit hash)
- **Value:** `[]git.FileInfo` (list of files)
- **Size:** ~100B per file × 10 files × 30 commits = ~30KB
- **Preload:** On app init, all commits
- **Invalidate:** After commit/merge/time-travel-merge
- **Strategy:** Always cache (lightweight, instant navigation)

### File(s) History Diff Cache
- **Key:** `hash:path:version` where version = "parent" | "wip"
- **Value:** diff content (plain text)
- **Size:** ~3KB per file × 10 files × 2 versions × 30 commits = ~1.8MB
- **Preload:** On app init, but SKIP commits with >100 files
- **Invalidate:** After commit/merge/time-travel-merge
- **Strategy:** Selective caching for performance

---

## Git Commands Required

| Purpose | Command | Returns |
|---------|---------|---------|
| Recent commits | `git log --pretty=%H%n%s%n%ai -30` | Hash, Subject, ISO DateTime |
| Commit metadata | `git show -s --pretty=%aN%n%aD%n%B <hash>` | Author, Date, Message |
| Files in commit | `git show --name-only --pretty= <hash>` | File paths (one per line) |
| Diff vs parent | `git diff <hash>^ <hash> -- <path>` | Unified diff |
| Diff vs working | `git diff <hash> -- <path>` | Unified diff |
| Checkout (time travel) | `git checkout <hash>` | Detached HEAD |
| Get TIT_TIME_TRAVEL branch | `cat .git/TIT_TIME_TRAVEL` | Branch name |

---

## State Detection in New TIT

**Key field:** `a.gitState.Operation` (enum)

### Affects History Feature
```
Operation = Normal
    ↓
Allow History menu items
    ↓
Can select History or File(s) History
    ↓
Can enter time travel (if select commit + ENTER)

Operation = TimeTraveling
    ↓
Disable normal menu items
    ↓
Show ONLY time travel menu (merge/return/browse)
    ↓
Can browse history while time traveling
    ↓
Can merge changes back to original branch
```

---

## Phase Execution Checklist

### Phase 1: Infrastructure & UI Types
- [ ] Add `CommitInfo`, `CommitDetails`, `FileInfo` to git/types.go
- [ ] Add `HistoryState`, `FileHistoryState` to app/app.go
- [ ] Add `ModeHistory`, `ModeFileHistory` to modes.go
- [ ] Verify compilation

### Phase 2: History Cache System
- [ ] Create historycache.go with preload functions
- [ ] Add cache fields to Application struct
- [ ] Call preload on app init
- [ ] Test memory usage + thread safety

### Phase 3: History UI & Rendering
- [ ] Create history.go with RenderHistorySplitPane()
- [ ] Add History case to layout.go
- [ ] Verify rendering with dummy data

### Phase 4: History Mode Handlers & Menu
- [ ] Add menuHistory() generator
- [ ] Add History menu item
- [ ] Add dispatchHistory()
- [ ] Add mode handlers (up/down/tab/esc)
- [ ] Register handlers in keyHandlers
- [ ] Test navigation

### Phase 5: File(s) History UI & Rendering
- [ ] Create filehistory.go with RenderFileHistorySplitPane()
- [ ] Add FileHistory case to layout.go
- [ ] Verify rendering with dummy data

### Phase 6: File(s) History Handlers & Menu
- [ ] Add menuFileHistory() generator
- [ ] Add File(s) History menu item
- [ ] Add dispatchFileHistory()
- [ ] Add mode handlers (up/down/tab/copy/visual/esc)
- [ ] Register handlers
- [ ] Test navigation

### Phase 7: Time Travel Integration
- [ ] Add time travel menu generator
- [ ] Add ENTER handler in History mode (show confirmation)
- [ ] Add git checkout operation
- [ ] Add .git/TIT_TIME_TRAVEL file management
- [ ] Add time travel handlers (merge/return)
- [ ] Test entry and exit

### Phase 8: Cache Invalidation & Integration
- [ ] Add invalidation functions
- [ ] Call from githandlers after commit/merge
- [ ] Add progress indication for disabled menu items
- [ ] Test cache refresh

### Phase 9: Testing & Verification
- [ ] Run all 26 manual test items
- [ ] Code review for patterns
- [ ] Race condition check
- [ ] Memory leak check

---

## Important Constants & Limits

| Item | Value | Reason |
|------|-------|--------|
| Commits to preload | 30 | Reasonable depth, manageable cache size |
| File count threshold | 100 | Skip diff caching for larger commits |
| Diff cache version | 2 ("parent", "wip") | State-dependent rendering |
| Cache load timeout | ~3 seconds | Non-blocking startup |
| Scroll page size | terminal height | Standard TUI convention |

---

## Common Pitfalls to Avoid

1. **❌ Don't use separate progress channels**
   - Progress integrated into app.Update() / gitState

2. **❌ Don't block on cache load**
   - Goroutines + tea.Cmd pattern

3. **❌ Don't mutate cache from goroutine**
   - Immutable snapshot approach (or minimal mutex scope)

4. **❌ Don't forget TIT_TIME_TRAVEL file cleanup**
   - Must handle on return from time travel

5. **❌ Don't skip menu enabling/disabling**
   - Menu items disabled until cache ready (UX clarity)

6. **❌ Don't cache diffs for >100-file commits**
   - Performance risk, causes lag

7. **❌ Don't forget InvalidateHistoryCaches() on commit**
   - Caches become stale otherwise

8. **❌ Don't mix cherry-pick and time-travel logic**
   - Cherry-pick is gone; time travel is new feature

---

## References

- **Full Plan:** See `HISTORY-IMPLEMENTATION-PLAN.md` (sections 1-12)
- **SPEC.md:** Sections 7, 9 (History, Time Travel specification)
- **ARCHITECTURE.md:** Sections 2, 3, 4 (State model, event model, modes)
- **Old-TIT Code:** `/Users/jreng/Documents/Poems/inf/old-tit/`

---

## Questions Before Starting Phase 1?

Send clarifications on:
1. Cache pre-load limit (30 commits OK?)
2. Diff cache threshold (>100 files OK?)
3. Time travel merge strategy (merge commit vs rebase?)
4. History depth (recent 30 or all commits?)
5. Cache reload timing (immediate or on next menu entry?)

**Ready to proceed with Phase 1?** ✅
