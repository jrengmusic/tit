# Phase 6: File(s) History Handlers & Menu - KICKOFF

**Status:** 🟢 READY  
**Estimated:** 1 day (~600 lines)  
**Complexity:** MEDIUM (6 handlers, menu integration)  
**Depends On:** Phase 5 (UI rendering complete)

---

## What Phase 6 Does

Adds **keyboard navigation and menu integration** to File(s) History mode. Makes it **fully functional and user-accessible**.

### Deliverables
- Menu item in main menu
- Dispatcher function to enter File(s) History mode
- 6 keyboard handlers for navigation and pane switching
- Keyboard handler registry setup
- Visual feedback during navigation

---

## Files to Modify

### 1. `internal/app/menu.go`

Add menu item generator function:
```go
func (a *Application) menuFileHistory() []MenuItem {
    // Return menu items specific to File(s) History mode
    // Same pattern as menuHistory()
}
```

Modify `menuNormal()` or appropriate generator to include File(s) History item:
```go
{
    ID:      "file_history",
    Label:   "📁 File(s) History",
    Hint:    "Browse file changes over time",
    Enabled: a.cacheDiffs,  // Only enabled after file history cache ready
}
```

---

### 2. `internal/app/dispatchers.go`

Add dispatcher function:
```go
func (a *Application) dispatchFileHistory() (tea.Model, tea.Cmd) {
    // Check cache ready
    if !a.cacheDiffs {
        a.footerHint = "File history cache is loading..."
        return a, nil
    }
    
    // Populate fileHistoryState from cache
    // Set initial selections
    // Enter ModeFileHistory
    a.mode = ModeFileHistory
    a.footerHint = "File(s) History │ ↑↓ navigate │ TAB cycle panes │ ESC back"
    return a, nil
}
```

Add to dispatcher registry (e.g., in app initialization):
```go
"file_history": (*Application).dispatchFileHistory
```

---

### 3. `internal/app/handlers.go`

Add 6 keyboard handlers:

#### `handleFileHistoryUp()` - Navigate up or scroll
```go
func (a *Application) handleFileHistoryUp() (tea.Model, tea.Cmd) {
    // If FocusedPane == PaneCommits: move selection up in commits
    // If FocusedPane == PaneFiles: move selection up in files
    // If FocusedPane == PaneDiff: scroll diff up
    
    // Update scroll offsets as needed
    // Keep selection in bounds
    return a, nil
}
```

#### `handleFileHistoryDown()` - Navigate down or scroll
```go
func (a *Application) handleFileHistoryDown() (tea.Model, tea.Cmd) {
    // Similar to handleFileHistoryUp but in opposite direction
}
```

#### `handleFileHistoryTab()` - Cycle focus
```go
func (a *Application) handleFileHistoryTab() (tea.Model, tea.Cmd) {
    // Cycle: Commits → Files → Diff → Commits
    // Update FocusedPane
    a.fileHistoryState.FocusedPane = (a.fileHistoryState.FocusedPane + 1) % 3
    return a, nil
}
```

#### `handleFileHistoryCopy()` - Copy selection (placeholder for Phase 8)
```go
func (a *Application) handleFileHistoryCopy() (tea.Model, tea.Cmd) {
    // For now: placeholder
    // Future: copy selected lines from diff to clipboard
    return a, nil
}
```

#### `handleFileHistoryVisualMode()` - Toggle visual mode (placeholder for Phase 8)
```go
func (a *Application) handleFileHistoryVisualMode() (tea.Model, tea.Cmd) {
    // For now: placeholder
    // Future: toggle visual selection mode in diff pane
    return a, nil
}
```

#### `handleFileHistoryEsc()` - Return to menu
```go
func (a *Application) handleFileHistoryEsc() (tea.Model, tea.Cmd) {
    a.mode = ModeMenu
    a.selectedIndex = 0  // Reset menu selection
    a.menuItems = a.GenerateMenu()
    return a, nil
}
```

---

### 4. `internal/app/app.go`

Register keyboard handlers in `NewModeHandlers()`:

```go
handlers[ModeFileHistory] = map[string]KeyHandler{
    "up":       (*Application).handleFileHistoryUp,
    "down":     (*Application).handleFileHistoryDown,
    "k":        (*Application).handleFileHistoryUp,    // Vim binding
    "j":        (*Application).handleFileHistoryDown,  // Vim binding
    "tab":      (*Application).handleFileHistoryTab,
    "y":        (*Application).handleFileHistoryCopy,
    "v":        (*Application).handleFileHistoryVisualMode,
    "esc":      (*Application).handleFileHistoryEsc,
}
```

---

## Implementation Pattern

Follow the **History mode (Phase 4) pattern exactly**:

1. **Menu item** - Static definition with `Enabled` flag
2. **Dispatcher** - Check cache, populate state, set mode
3. **Handlers** - Modify state fields, return (a, nil)
4. **Registry** - Add to mode handlers map
5. **No new types** - Reuse FileHistoryState from Phase 5

---

## Key Navigation Logic

### When FocusedPane = PaneCommits
- ↑ / k: Move selection up in commits list
- ↓ / j: Move selection down in commits list
- Action: Update `SelectedCommitIdx`, adjust scroll, populate Files pane

### When FocusedPane = PaneFiles
- ↑ / k: Move selection up in files list
- ↓ / j: Move selection down in files list
- Action: Update `SelectedFileIdx`, adjust scroll, get diff from cache

### When FocusedPane = PaneDiff
- ↑ / k: Scroll diff up
- ↓ / j: Scroll diff down
- Action: Update `DiffScrollOff`

### TAB (all panes)
- Cycle: PaneCommits → PaneFiles → PaneDiff → PaneCommits
- Update `FocusedPane` field

---

## State Updates in Handlers

Each handler modifies `a.fileHistoryState` directly:

```go
// Example: handleFileHistoryDown
if a.fileHistoryState.FocusedPane == PaneCommits {
    if a.fileHistoryState.SelectedCommitIdx < len(a.fileHistoryState.Commits)-1 {
        a.fileHistoryState.SelectedCommitIdx++
        
        // Fetch files for this commit from cache
        commitHash := a.fileHistoryState.Commits[a.fileHistoryState.SelectedCommitIdx].Hash
        a.fileHistoryState.Files = a.fileHistoryDiffCache[commitHash]  // or similar
        
        // Reset file selection
        a.fileHistoryState.SelectedFileIdx = 0
    }
}
```

---

## Menu Integration

### In `menuNormal()` (or appropriate generator)

Add File(s) History item:
```go
items = append(items, MenuItem{
    ID:      "file_history",
    Label:   "📁 File(s) History",
    Hint:    "Browse file changes over time",
    Enabled: a.cacheDiffs,  // Disabled until cache ready
})
```

### In `handleMenuEnter()` (or key handler dispatcher)

Route to dispatcher:
```go
if selectedItem.ID == "file_history" {
    return a.dispatchFileHistory()
}
```

---

## Differences from History Mode (Phase 4)

| Aspect | History Mode | File(s) History Mode |
|--------|--------------|----------------------|
| Panes | 2 | 3 |
| Focus states | 2 | 3 |
| Handlers | 5 | 6 |
| Scroll offsets | 2 | 3 |
| Navigation | Commits + Details | Commits + Files + Diff |
| Cache check | `a.cacheMetadata` | `a.cacheDiffs` |

---

## Testing Checklist

- [ ] Build: `./build.sh` compiles clean
- [ ] Functional: Enter File(s) History mode from menu
- [ ] Navigation: ↑↓ navigate commits
- [ ] Selection: Commits and files change focus
- [ ] Pane switching: TAB cycles through 3 panes
- [ ] Scroll: Diff pane scrolls independently
- [ ] Escape: ESC returns to menu
- [ ] No regressions: History mode (Phase 4) still works
- [ ] Cache ready: Menu item only enabled when cache loaded

---

## What Phase 6 Does NOT Do

- ❌ No copy functionality (y key shows placeholder)
- ❌ No visual mode (v key shows placeholder)
- ❌ No real diff data (still uses placeholder from Phase 5)
- ❌ No cache population (data comes from Phase 2 cache)

These are deferred to Phase 7+ per specification.

---

## Acceptance Criteria

- [ ] Menu item visible (enabled when cache ready)
- [ ] Can select File(s) History → enters ModeFileHistory
- [ ] ↑↓ navigate commits in left pane
- [ ] Selecting commit updates file list in middle pane
- [ ] ↑↓ navigate files in middle pane
- [ ] Selecting file updates diff in right pane
- [ ] TAB cycles focus between all 3 panes
- [ ] Y / V keys are placeholders (no crash)
- [ ] ESC returns to main menu
- [ ] No regressions to Phase 3/4
- [ ] Compiles without errors
- [ ] Compiles without warnings

---

## Reference

- **Phase 4 Implementation:** PHASE-4-COMPLETION.md (same pattern)
- **Phase 5 UI:** PHASE-5-AUDIT-REPORT.md (3-pane layout)
- **Architecture:** ARCHITECTURE.md (handler pattern)
- **Spec:** SPEC.md § 11 (File(s) History behavior)
- **Implementation Plan:** HISTORY-IMPLEMENTATION-PLAN.md § Phase 6

---

## Quick Checklist for Agent

Before starting:
- [ ] Read PHASE-5-AUDIT-REPORT.md (understand the UI)
- [ ] Study PHASE-4-COMPLETION.md (follow this pattern exactly)
- [ ] Understand FileHistoryState struct
- [ ] Know the 3 focus states: PaneCommits, PaneFiles, PaneDiff
- [ ] Plan navigation logic for each pane

---

**Proceed when ready.** Phase 6 builds directly on Phase 5 (no new concepts, just handlers).

Good luck! 🚀
