# Phase 7 Architecture Audit: Time Travel Implementation

**Date:** 2026-01-08  
**Status:** ARCHITECTURAL VIOLATIONS FOUND  
**Severity:** CRITICAL (2), MEDIUM (3)

---

## Executive Summary

Phase 7 implementation has **5 architectural violations** against SPEC.md and ARCHITECTURE.md:

| # | Issue | Severity | Location | Impact |
|---|-------|----------|----------|--------|
| 1 | Time travel menu items not in SSOT | CRITICAL | `menu.go:239-250` | Violates MenuItem SSOT pattern |
| 2 | DirtyOperation missing from menuGenerators | CRITICAL | `menu.go:30-37` | Missing Priority 1 menu handler |
| 3 | Time travel menu missing "View diff" option | MEDIUM | `menu.go:230-252` | Incomplete per SPEC.md:136-139 |
| 4 | Using Item() builder instead of GetMenuItem() | MEDIUM | `menu.go:239-250` | Violates ARCHITECTURE.md:201-205 |
| 5 | CurrentBranch usage may be wrong during time travel | MEDIUM | `menu.go:234-235` | Should read from TIT_TIME_TRAVEL file |

---

## Detailed Findings

### 1. ❌ CRITICAL: Time Travel Menu Items Not in SSOT

**ARCHITECTURE.md Rule (Section: Menu Item SSOT System):**
```
All menu items defined in single source of truth (internal/app/menuitems.go)
```

**Violation:**
```go
// ❌ menu.go:239-250 - Direct Item() builder usage
Item("time_travel_history").
    Label("🕒 Browse History").
    Hint("View commit history while time traveling").
    Build(),
```

**Should be:**
```go
// ✅ In menuitems.go SSOT
"time_travel_history": {
    ID: "time_travel_history",
    Label: "🕒 Browse History",
    Hint: "View commit history while time traveling",
},

// ✅ In menu.go
GetMenuItem("time_travel_history"),
GetMenuItem("time_travel_merge"),
GetMenuItem("time_travel_return"),
```

**Impact:**
- Breaks centralized menu item management
- Shortcuts not validated globally
- Makes auditing and refactoring difficult

**Fix Required:** Move all 3 time travel items to `menuitems.go` SSOT

---

### 2. ❌ CRITICAL: DirtyOperation Missing from menuGenerators

**SPEC.md Rule (Section 4: State Priority Rules):**
```
Priority 1: Operation State (Most Restrictive)
- DirtyOperation → Show ONLY dirty operation control menu
```

**ARCHITECTURE.md Rule (Section: Menu System):**
```go
menuGenerators := map[git.Operation]MenuGenerator{
    git.NotRepo:    (*Application).menuNotRepo,
    git.Conflicted: (*Application).menuConflicted,
    git.Merging:    (*Application).menuOperation,
    git.Rebasing:   (*Application).menuOperation,
    git.Normal:     (*Application).menuNormal,
    // ❌ MISSING: git.DirtyOperation
}
```

**Violation:**
`menu.go:30-37` does NOT have entry for `git.DirtyOperation`

**Expected per SPEC.md:128-131:**
```
When Operation = DirtyOperation
Show ONLY:
- 🔄 View operation status
- ⛔ Abort dirty operation (restores exact original state)
```

**Impact:**
- If DirtyOperation state is detected, GenerateMenu() will panic (line 44)
- Code comments in `git/state.go:60-61` say it returns "Conflicted" as workaround
- This is WRONG - it violates the state model

**Fix Required:**
```go
// In menu.go:30-37
menuGenerators := map[git.Operation]MenuGenerator{
    git.NotRepo:      (*Application).menuNotRepo,
    git.Conflicted:   (*Application).menuConflicted,
    git.Merging:      (*Application).menuOperation,
    git.Rebasing:     (*Application).menuOperation,
    git.DirtyOperation: (*Application).menuDirtyOperation,  // ← ADD THIS
    git.Normal:       (*Application).menuNormal,
    git.TimeTraveling: (*Application).menuTimeTraveling,
}

// Add new function in menu.go
func (a *Application) menuDirtyOperation() []MenuItem {
    return []MenuItem{
        GetMenuItem("view_operation_status"),
        GetMenuItem("abort_operation"),
    }
}
```

**Context:** `git/state.go:59-65` shows this check:
```go
if IsDirtyOperationActive() {
    return &State{
        Operation: Conflicted,  // ← WRONG: Reuses Conflicted
    }, nil
}
```

---

### 3. ⚠️ MEDIUM: Time Travel Menu Missing "View diff" Option

**SPEC.md:133-139 requires:**
```
When Operation = TimeTraveling
Show ONLY:
- 🕒 Jump to different commit
- 👁️ View diff (vs original branch)     ← MISSING
- 📦 Merge changes back to [branch]
- ⬅️ Return to [branch] (discard changes)
```

**Current Implementation (`menu.go:238-251`):**
```
✓ 🕒 Browse History
✓ 📦 Merge back to [branch]
✓ ⬅️ Return to [branch]
✗ 👁️ View diff (vs original branch)    ← MISSING
```

**Note:** "Browse History" is from History mode, not time travel menu.

**Fix Required:** Add diff view option to menuTimeTraveling()

---

### 4. ⚠️ MEDIUM: Using Item() Builder Instead of GetMenuItem()

**ARCHITECTURE.md:201-205 Rule:**
```go
// ❌ OLD
Item("commit").Shortcut("m").Label("...").Build()

// ✅ NEW
GetMenuItem("commit")
```

**Current Issue:**
`menu.go:239-250` uses:
```go
Item("time_travel_history").Label(...).Hint(...).Build()  // ❌ OLD pattern
```

**Should use:**
```go
GetMenuItem("time_travel_history")  // ✅ NEW pattern (once in SSOT)
```

---

### 5. ⚠️ MEDIUM: CurrentBranch Usage During Time Travel

**Current Code (`menu.go:234-235`):**
```go
originalBranch := "unknown"
if a.gitState != nil && a.gitState.CurrentBranch != "" {
    originalBranch = a.gitState.CurrentBranch
}
```

**Problem:** During time travel, `CurrentBranch` is the detached HEAD hash, NOT the original branch.

**Should use:** Read from `.git/TIT_TIME_TRAVEL` file via `git.GetTimeTravelInfo()`

**Example:**
```go
originalBranch := "unknown"
if origBranch, _, err := git.GetTimeTravelInfo(); err == nil {
    originalBranch = origBranch
}
```

**Why:** The git state tuple shows detached HEAD during time travel, but UI needs to show the original branch name in the labels.

---

## State Detection Order Verification

✅ **CORRECT** - `git/state.go:319-353` (detectOperation):

```
1. Conflicted (conflicts detected)         ✓ Line 320-330
2. TimeTraveling (TIT_TIME_TRAVEL exists)  ✓ Line 332-336
3. Merging/Rebasing/Cherry-pick            ✓ Line 338-350
4. Normal (fallback)                       ✓ Line 352
```

**Missing:**
- DirtyOperation check (should be Priority 0 or 1, above TimeTraveling)

---

## Menu Generator Coverage

**SPEC.md requires handlers for all Operation types:**

| Operation | Handler | Status |
|-----------|---------|--------|
| NotRepo | menuNotRepo | ✅ Exists |
| Conflicted | menuConflicted | ✅ Exists |
| Merging | menuOperation | ✅ Exists |
| Rebasing | menuOperation | ✅ Exists |
| DirtyOperation | menuDirtyOperation | ❌ **MISSING** |
| Normal | menuNormal | ✅ Exists |
| TimeTraveling | menuTimeTraveling | ✅ Exists |

---

## State Tuple During Time Travel

When `Operation = TimeTraveling`:

```go
State{
    WorkingTree: Clean or Dirty        // User can make changes
    Timeline: InSync                   // Detached HEAD (no timeline)
    Operation: TimeTraveling           // ✓ Correct
    Remote: HasRemote or NoRemote      // ✓ Correct
    CurrentBranch: "abc123def..."      // ⚠️ DETACHED HEAD HASH, NOT ORIGINAL
    CurrentHash: "abc123def..."        // ✓ Correct
}
```

**Issue:** Menu label needs original branch (from TIT_TIME_TRAVEL file), not CurrentBranch.

---

## Async Operation Pattern Compliance

✅ **CORRECT** - Time travel operations follow async pattern:

```go
// In git/execute.go
ExecuteTimeTravelCheckout() tea.Cmd  // ✅ Returns tea.Cmd
ExecuteTimeTravelMerge() tea.Cmd     // ✅ Returns tea.Cmd
ExecuteTimeTravelReturn() tea.Cmd    // ✅ Returns tea.Cmd

// Workers return immutable messages
git.TimeTravelCheckoutMsg struct{...}
git.TimeTravelMergeMsg struct{...}
git.TimeTravelReturnMsg struct{...}
```

✅ Message handlers in `app.go:309-319` dispatch to:
- `githandlers.go` methods
- Proper state reload via `DetectState()`

---

## Message Handling Chain

✅ **CORRECT** - `app.go:309-319`:

```go
case git.TimeTravelCheckoutMsg:
    return a.handleTimeTravelCheckout(msg)
case git.TimeTravelMergeMsg:
    return a.handleTimeTravelMerge(msg)
case git.TimeTravelReturnMsg:
    return a.handleTimeTravelReturn(msg)
```

✅ Handlers reload state: `DetectState() → GenerateMenu()` flow verified

---

## File I/O Operations

✅ **CORRECT** - `git/state.go:396-449`:

```go
GetTimeTravelInfo()      // ✅ Reads .git/TIT_TIME_TRAVEL
WriteTimeTravelInfo()    // ✅ Writes original branch + stash ID
ClearTimeTravelInfo()    // ✅ Cleanup on return/merge
```

All thread-safe, no race conditions (run in workers, not UI thread).

---

## Dirty Tree Handling

✅ **CORRECT** - Before time travel:

```go
ExecuteTimeTravelCheckout() {
    if workingTree == Dirty {
        git stash save                    // ✓ Stash changes
        WriteTimeTravelInfo(...stashID)   // ✓ Record stash ID
    }
}
```

✅ On return/merge:
- Stash restored automatically if stashID recorded
- Uses existing dirty operation protocol

---

## Summary of Required Fixes

### CRITICAL (Must Fix Before Testing)

1. **Add DirtyOperation to menuGenerators:**
   - Add entry in `menu.go:30-37`
   - Implement `menuDirtyOperation()` function
   - Fix `git/state.go:59-65` to return `DirtyOperation` instead of `Conflicted`

2. **Move time travel items to SSOT:**
   - Add 3 items to `menuitems.go`
   - Update `menuTimeTraveling()` to use `GetMenuItem()`

### MEDIUM (Should Fix)

3. **Add "View diff" to time travel menu:**
   - Add menu item to SSOT
   - Implement handler

4. **Fix originalBranch lookup:**
   - Use `GetTimeTravelInfo()` instead of `CurrentBranch`

---

## Files Requiring Changes

| File | Changes | Lines |
|------|---------|-------|
| `internal/app/menuitems.go` | Add 3 time travel items | ~10 |
| `internal/app/menu.go` | Fix menuTimeTraveling(), add menuDirtyOperation() | ~20 |
| `internal/git/state.go` | Return DirtyOperation instead of Conflicted | ~3 |
| `internal/app/githandlers.go` | May need dirty operation handler | ~15 |

**Total: ~50 lines of code changes**

---

## Testing Strategy After Fixes

```bash
# 1. Unit test: state detection order
- Create dirty operation scenario
- Verify DirtyOperation state returned (not Conflicted)

# 2. Unit test: menuGenerators coverage
- Call GenerateMenu() for each Operation type
- Verify no panics, all menus generate

# 3. Integration test: time travel flow
- Checkout commit → Operation=TimeTraveling
- Verify menu shows correct 4 items
- Verify originalBranch label correct
- Merge/return → Operation=Normal

# 4. UI test: dirty tree time travel
- Make changes, try time travel
- Verify stash created
- Return from time travel
- Verify changes restored
```

