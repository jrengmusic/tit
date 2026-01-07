# Phase 5 Audit & Acceptance Checklist

**Role:** Auditor verification document  
**Phase:** 5 - File(s) History UI & Rendering  
**Expected Output:** `internal/ui/filehistory.go` + modifications to `layout.go`  

---

## Pre-Execution Briefing

### What Phase 5 Is
- **Scope:** UI rendering only (no handlers, no menu items, no mode transitions)
- **Output:** File(s) History 3-pane split layout (commits | files | diff)
- **Dependency:** Builds on Phase 4 (uses existing cache from Phase 2)
- **Complexity:** MEDIUM (3 panes vs 2 in History mode)

### What Phase 5 Is NOT
- ❌ NOT keyboard handlers (that's Phase 6)
- ❌ NOT menu integration (that's Phase 6)
- ❌ NOT data display (can use placeholders; actual data comes in Phase 6)
- ❌ NOT mode transitions (that's Phase 6)

---

## Reference Documents (Read First)

1. **PHASE-5-KICKOFF.md** - Official specification
2. **HISTORY-IMPLEMENTATION-PLAN.md § Phase 5** - Technical details
3. **internal/ui/history.go** - Pattern to follow (2-pane layout)
4. **internal/ui/listpane.go** - Component to reuse
5. **internal/ui/diffpane.go** - Component to reuse

---

## File Checklist

### File: `internal/ui/filehistory.go` ✅ Must Create

**Size estimate:** ~400-500 lines

**Required functions:**
- [ ] `RenderFileHistorySplitPane(state interface{}, theme *Theme, width, height int) string`
  - Main entry point
  - Type assertion to `*FileHistoryState`
  - Calls 3 render functions
  - Uses `lipgloss.JoinHorizontal()` to combine panes
  
- [ ] `renderFileHistoryCommitsPane(state *FileHistoryState, theme *Theme, width, height int) string`
  - Left pane (reuse ListPane logic)
  - Shows commit list
  
- [ ] `renderFileHistoryFilesPane(state *FileHistoryState, theme *Theme, width, height int) string`
  - Middle pane (reuse ListPane logic)
  - Shows files in selected commit
  
- [ ] `renderFileHistoryDiffPane(state *FileHistoryState, theme *Theme, width, height int) string`
  - Right pane (use existing DiffPane pattern)
  - Shows diff of selected file

**Required types (mirror from app.go):**
- [ ] `type FileHistoryState struct { ... }`
- [ ] `type FileHistoryPane int` enum
- [ ] `type FileInfo struct { ... }`

**Layout math:**
```go
// Divide content width into 3 columns
commitPaneWidth := (width / 3)
filesPaneWidth := (width / 3)
diffPaneWidth := (width / 3) + (width % 3)  // Handle remainder
```

**Styling:**
- [ ] Use theme colors for pane borders
- [ ] Reuse `ConflictPaneUnfocusedBorder`, `ConflictPaneFocusedBorder` (or define new ones)
- [ ] Focus indicator: bright border on active pane, dim on others

---

### File: `internal/ui/layout.go` ✅ Must Modify

**Changes needed:**
- [ ] Add case for `ModeFileHistory` in `View()` method
- [ ] Call `ui.RenderFileHistorySplitPane()` with proper parameters

**Example:**
```go
case ModeFileHistory:
    if a.fileHistoryState == nil {
        contentText = "File history not initialized"
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

## Code Quality Checks

### Architecture Compliance (CRITICAL)
- [ ] No circular imports (filehistory.go can import types.go, app.go, theme.go)
- [ ] Type assertions with `ok` check (not bare assertion)
- [ ] No placeholder data hardcoding (use state fields)
- [ ] Error handling: return string error messages, not silent failures
- [ ] No `fmt.Sprintf()` for rendering (use `lipgloss` only)

### SSOT Compliance
- [ ] Layout constants use `ui.ContentInnerWidth`, `ui.ContentHeight` (not hardcoded)
- [ ] No duplicate color definitions (reuse `theme.go` colors)
- [ ] Pane sizing uses layout SSOT (consistent with other components)

### Reuse Compliance
- [ ] Commits pane: mirrors `renderHistoryListPane()` logic
- [ ] Files pane: reuses ListPane pattern (same borders, spacing)
- [ ] Diff pane: follows DiffPane structure
- [ ] No reinventing layout logic (use `lipgloss.JoinHorizontal()`)

### Error Handling (FAIL-FAST)
- [ ] Type assertion checked: `historyState, ok := state.(*FileHistoryState); if !ok { return "error" }`
- [ ] No `_ = ` error suppression
- [ ] Clear error messages if state invalid

---

## Testing Checklist (Auditor Verification)

### Build
- [ ] `./build.sh` compiles without errors
- [ ] `./build.sh` compiles without warnings
- [ ] Binary created: `tit_x64`
- [ ] No "undefined reference" errors

### Code Review
- [ ] File structure matches history.go pattern
- [ ] Functions properly documented (comment headers)
- [ ] No TODO comments left behind
- [ ] Variable names semantic (not `temp`, `data`, `x`)
- [ ] Line length reasonable (~100 chars max)

### Functional (Manual)
- [ ] Can call `RenderFileHistorySplitPane()` without panic
- [ ] 3-pane layout renders (can pass dummy state)
- [ ] Panes divide screen width roughly equally
- [ ] Focus indicator works (if FocusedPane field varies)
- [ ] No off-by-one errors in scroll calculations

### Integration
- [ ] `layout.go` compiles with new case statement
- [ ] App doesn't crash if `fileHistoryState` is nil
- [ ] App renders "File history not initialized" gracefully
- [ ] No regressions to History mode (Phase 4)
- [ ] No regressions to other modes

---

## Acceptance Criteria (Official)

✅ **Phase 5 is DONE if ALL below are true:**

1. [ ] `internal/ui/filehistory.go` exists
2. [ ] `RenderFileHistorySplitPane()` implemented
3. [ ] 3-pane layout working (commits | files | diff)
4. [ ] ListPane + DiffPane reused correctly
5. [ ] `layout.go` updated with `ModeFileHistory` case
6. [ ] Compiles without errors
7. [ ] Compiles without warnings
8. [ ] No regressions to existing modes
9. [ ] Code follows ARCHITECTURE.md patterns
10. [ ] Functions properly documented

---

## Known Issues to Watch For

| Issue | Severity | Watch For |
|-------|----------|-----------|
| Width calculation off-by-one | HIGH | Test with different terminal widths |
| Focus indicator not visible | MEDIUM | Check border color contrast |
| Panes not aligned horizontally | MEDIUM | Verify all panes same height |
| DiffPane scrolling broken | MEDIUM | Ensure scroll offset preserved |
| Type assertion panic | CRITICAL | Must have `ok` check |

---

## Red Flags (Reject If Found)

🚨 **REJECT Phase 5 if:**
- [ ] ❌ Compilation errors
- [ ] ❌ Compilation warnings
- [ ] ❌ Circular imports
- [ ] ❌ Bare type assertions without `ok` check
- [ ] ❌ Hardcoded layout constants (no SSOT)
- [ ] ❌ Placeholder data in rendering (use actual state fields)
- [ ] ❌ Silent error handling (missing error checks)
- [ ] ❌ Code duplication with history.go
- [ ] ❌ Regressions to Phase 3/4 functionality
- [ ] ❌ Unused imports

---

## Audit Sign-Off Template

When Phase 5 work is ready, I will verify:

```
PHASE 5 AUDIT REPORT
====================

Submission Date: ____
Agent: ____

BUILD STATUS
✓/✗ Compiles clean (no errors)
✓/✗ Compiles clean (no warnings)
✓/✗ Binary created

CODE QUALITY
✓/✗ Architecture compliant
✓/✗ SSOT compliant
✓/✗ Component reuse correct
✓/✗ Error handling correct
✓/✗ No regressions

ACCEPTANCE CRITERIA
✓/✗ All 10 criteria met

VERDICT: APPROVED / REJECTED

Comments: ____
```

---

## Integration Points to Verify

| Component | Integration Point | Verify |
|-----------|-------------------|--------|
| **layout.go** | `View()` method adds `ModeFileHistory` case | Renders without error |
| **app.go** | `fileHistoryState` field exists | No nil panic |
| **theme.go** | Border colors available | filehistory.go builds |
| **listpane.go** | Pattern borrowed | Consistent styling |
| **diffpane.go** | Pattern borrowed | Consistent styling |

---

## Handoff Notes

- **Agent executing Phase 5:** [Name/ID]
- **Specification:** PHASE-5-KICKOFF.md (final authority)
- **Technical details:** HISTORY-IMPLEMENTATION-PLAN.md § Phase 5
- **Pattern reference:** `internal/ui/history.go` (follow this structure)
- **Estimated time:** 1-2 hours
- **Estimated lines:** ~700 total (filehistory.go + layout.go changes)

---

## Questions for Agent Before Starting

Ask yourself these before beginning:

1. ✅ Have you read PHASE-5-KICKOFF.md completely?
2. ✅ Have you studied history.go to understand the pattern?
3. ✅ Do you understand the 3-pane layout structure?
4. ✅ Do you know what FileHistoryState struct contains?
5. ✅ Do you understand you're NOT implementing handlers (Phase 6)?

If ANY answer is "no", re-read the specification first.

---

**Document Status:** Ready for Phase 5 handoff  
**Auditor:** Amp (claude-code)  
**Date Created:** 2026-01-08
