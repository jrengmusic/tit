# Phase 5: File(s) History UI & Rendering - KICKOFF

**Status:** 🟢 READY  
**Estimated:** 1 day (~700 lines)  
**Complexity:** MEDIUM (3 panes instead of 2)

---

## What Phase 5 Does

Builds UI rendering for **File(s) History mode** with 3-pane split layout:
- **Left:** Commit list (same as History mode)
- **Middle:** File list for selected commit
- **Right:** Diff pane showing changes to selected file

---

## Architecture

### Layout Structure
```
┌─────────────────────────────────────────────────────────┐
│ Commits        │ Files           │ Diff                │
├────────────────┼─────────────────┼─────────────────────┤
│                │                 │                     │
│ 02-Jan 08:42   │ ✓ main.go       │ @@ -1,10 +1,12    │
│ c6d5e6f        │ [ ] README.md   │  package main      │
│ 01-Jan 15:30   │ [ ] config.yaml │                    │
│ a1b2c3d        │                 │  func main() {     │
│ 31-Dec 22:15   │                 │ +  fmt.Println()   │
│ 9f8e7d6        │                 │  }                 │
│                │                 │                    │
└────────────────┴─────────────────┴─────────────────────┘
 TAB cycles focus → Changes diff   UP/DOWN scrolls diff
```

### State Management

Use `FileHistoryState` (already defined in app.go):
```go
type FileHistoryState struct {
    Commits           []git.CommitInfo  // List of commits
    Files             []git.FileInfo    // Files in selected commit
    SelectedCommitIdx int               // Selected commit
    SelectedFileIdx   int               // Selected file
    FocusedPane       FileHistoryPane   // Which pane has focus (0=commits, 1=files, 2=diff)
    CommitsScrollOff  int               // Scroll offset for commits
    FilesScrollOff    int               // Scroll offset for files
    DiffScrollOff     int               // Scroll offset for diff
}

type FileHistoryPane int
const (
    PaneCommits FileHistoryPane = iota
    PaneFiles
    PaneDiff
)
```

---

## Files to Create/Modify

### 1. `internal/ui/filehistory.go` - NEW FILE

**Purpose:** File(s) history UI rendering (similar to history.go)

**Key functions:**
1. `RenderFileHistorySplitPane(state interface{}, theme, width, height)` - Main entry
2. `renderFileHistoryCommitsPane()` - Left (reuse ListPane)
3. `renderFileHistoryFilesPane()` - Middle (reuse ListPane)
4. `renderFileHistoryDiffPane()` - Right (use DiffPane component)

**Layout calculation:**
```go
// Divide width equally (or with small adjustments)
commitPaneWidth := (contentWidth / 3)
filesPaneWidth := (contentWidth / 3)
diffPaneWidth := (contentWidth / 3) + (contentWidth % 3)  // Handle remainder
```

**Type definitions** (mirror app.go types):
```go
type FileHistoryState struct { ... }
type FileHistoryPane int
type FileInfo struct { ... }
```

---

### 2. `internal/app/app.go` - Add File(s) History rendering case

In `View()` method, add:
```go
case ModeFileHistory:
    if a.fileHistoryState == nil {
        contentText = "File history state not initialized"
    } else {
        contentText = ui.RenderFileHistorySplitPane(
            a.fileHistoryState,
            a.theme,
            ui.ContentInnerWidth,
            ui.ContentHeight,
        )
    }
```

---

## Implementation Guidance

### Reuse Components

**ListPane** (left + middle panes):
- Commits pane: Same as History mode
- Files pane: Show file list with scroll

**DiffPane** (right pane):
- Line numbers, syntax highlighting
- Scroll independently
- Show both versions if needed (additions in green, deletions in red)

**Theme colors:**
- Same conflict resolver colors for consistency
- ConflictPaneUnfocusedBorder, ConflictPaneFocusedBorder

### Keyboard Hints

Status bar shows:
```
↑↓ navigate  │  TAB cycle panes  │  ENTER details  │  ESC back
```

---

## Differences from History Mode

| Aspect | History Mode | File(s) History Mode |
|--------|--------------|----------------------|
| Panes | 2 (commits, details) | 3 (commits, files, diff) |
| Focus | List or Details | 3 states |
| Scroll | Independent per pane | Independent (3 offsets) |
| Data source | Metadata cache | Files cache + diff cache |
| Components | ListPane + simple text | ListPane + ListPane + DiffPane |

---

## Testing

1. **Build:** `./build.sh` compiles clean
2. **Code review:** Rendering logic correct
3. **Integration:** Phase 5 builds on Phase 4 (no breaking changes)

Full functional testing deferred to Phase 6 (handlers).

---

## What Phase 5 Does NOT Do

- ❌ No keyboard handlers (Phase 6)
- ❌ No menu items (Phase 6)
- ❌ No mode transitions (Phase 6)
- ❌ No data display yet (uses placeholders OK for Phase 5)

This phase is **UI rendering only**.

---

## Acceptance Criteria

- [x] `internal/ui/filehistory.go` created
- [x] RenderFileHistorySplitPane() implemented
- [x] 3-pane layout working
- [x] ListPane + DiffPane reused correctly
- [x] Layout.go updated with ModeFileHistory case
- [x] Compiles without errors
- [x] No regressions
- [x] Code follows patterns

---

## Reference

- **PHASE-3-KICKOFF.md** - History mode pattern (similar approach)
- **HISTORY-IMPLEMENTATION-PLAN.md § Phase 5** - Full specs
- **DiffPane component** - `internal/ui/diffpane.go` (already exists)
- **ListPane component** - `internal/ui/listpane.go` (already exists)

---

**Proceed when ready.** See HISTORY-IMPLEMENTATION-PLAN.md § Phase 5 for details.

Good luck! 🚀
