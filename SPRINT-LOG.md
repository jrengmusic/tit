# Sprint Log

## Current Session: Stale Stash Handling & Dual Conflict Resolution

**Date:** Sat Jan 31 2026
**Status:** ✅ COMPLETE

---

## Participants

- `@surgeon` (Complex Fix Specialist)
- `@oracle` (invoked for bug analysis - not needed, root cause identified)
- `@librarian` (not needed)
- `@engineer` (not needed)
- `@machinist` (not needed)
- `@auditor` (not needed)
- `@pathfinder` (not needed)
- `@researcher` (not needed)

---

## Problem Statement

TIT crashes when user manually drops a stash after time travel:
1. User time travels with dirty working tree → stash created
2. User runs `git stash drop` manually → stash gone
3. User tries to return from time travel → TIT crashes trying to apply non-existent stash

**Additional issue:** No handling for dual conflicts (merge conflict + stash apply conflict simultaneously).

---

## Root Causes Identified

1. **`FindStashRefByHash()` panic**: Function panicked on missing stash instead of returning gracefully
2. **Missing stale stash dialog**: No confirmation UI when stash was manually dropped
3. **Duplicate stash entry panic**: `AddStashEntry()` panicked if entry already existed
4. **TOML parse panic**: `loadStashList()` panicked on corrupted/empty config file

---

## Files Modified

### `internal/git/execute.go`
- **Lines 102-125**: `FindStashRefByHash()` signature changed from `string` to `(string, bool)`
- **Lines 127-132**: Added `StashExists()` wrapper function
- **Purpose**: Non-panicking stash lookup for graceful error handling

### `internal/app/confirm_handlers.go`
- **Lines 589-613**: Added `executeConfirmStaleStashContinue()` handler
- **Lines 615-619**: Added `executeRejectStaleStashContinue()` handler
- **Lines 621-652**: Added `executeConfirmStaleStashMergeContinue()` handler
- **Lines 272-279**: Added stale entry check before `AddStashEntry()` to prevent panic
- **Purpose**: Handle stale stash confirmation dialogs and prevent duplicate entry panics

### `internal/app/confirm_dialog.go`
- **Lines 101-108**: Added handler mappings for `confirm_stale_stash_continue` and `confirm_stale_stash_merge_continue`
- **Purpose**: Route stale stash confirmation responses to handlers

### `internal/config/stash.go`
- **Lines 56-70**: `loadStashList()` now resilient to corrupted files (resets instead of panicking)
- **Purpose**: Prevent crash on malformed TOML config

---

## Technical Debt / Follow-up

1. **Consistent dialog patterns**: Stale stash dialogs use different titles/explanations than other dialogs (follow up if standardization needed)
2. **Test coverage**: Consider adding integration tests for stale stash scenarios
3. **Documentation**: Add stale stash handling to user documentation

---

## Alignment Check

### ✅ LIFESTAR Principles
- **Layer Separation**: Fix respects layer boundaries (git → config → app)
- **Fail Fast**: Invalid states detected early with clear error messages
- **Single Responsibility**: Each function has one clear purpose
- **Testability**: Pure functions, injectable dependencies

### ✅ NAMING-CONVENTION
- **ActionID naming**: `confirm_stale_stash_continue` follows kebab-case pattern
- **Handler naming**: `executeConfirmStaleStashContinue()` follows verb-noun pattern
- **Consistency**: Matches existing `executeConfirmTimeTravelReturn()` pattern

### ✅ ARCHITECTURAL-MANIFESTO
- **State Machine**: UI state transitions correctly managed
- **No Magic Numbers**: All constants defined as `const` or in config
- **Error Recovery**: Graceful degradation on corrupted state

---

## Problems Solved

1. ✅ TIT no longer crashes on manually dropped stash
2. ✅ User sees confirmation dialog: "Original stash X was manually dropped. Continue without restoring stash?"
3. ✅ "Continue" button proceeds with time travel return (stash skipped)
4. ✅ "Cancel" button returns to menu
5. ✅ Dual conflict scenario tested: merge conflict + stash apply conflict
6. ✅ No panics on corrupted TOML config files
7. ✅ No panics on duplicate stash entries

---

## Test Scenarios Validated

1. **Stale Stash (Return)**: Dirty TT → drop stash → return → confirmation dialog → continue → success
2. **Stale Stash (Merge)**: Dirty TT → drop stash → merge → confirmation dialog → continue → success
3. **Dual Conflict**: Dirty TT → make conflicting commit → return → merge conflict → stash apply conflict → conflict resolver
4. **Corrupted Config**: Empty `[]` TOML file → loadStashList resets → continues
5. **Duplicate Entry**: Stale config entry → auto-cleaned before adding new entry

---

## Key Code Changes Summary

```go
// Before: Panics on missing stash
func FindStashRefByHash(targetHash string) string

// After: Returns (stashRef, found)
func FindStashRefByHash(targetHash string) (string, bool)
func StashExists(stashHash string) bool

// Before: AddStashEntry panics on duplicate
config.AddStashEntry("time_travel", hash, repoPath, branch, commit)

// After: Check and clean stale entry first
if _, exists := config.GetStashEntry("time_travel", repoPath); exists {
    config.RemoveStashEntry("time_travel", repoPath)
}
config.AddStashEntry("time_travel", hash, repoPath, branch, commit)

// Before: loadStashList panics on corrupted TOML
if _, err := toml.DecodeFile(stashFile, &list); err != nil {
    panic(fmt.Sprintf("FATAL: %v", err))
}

// After: Resets corrupted file gracefully
if _, err := toml.DecodeFile(stashFile, &list); err != nil {
    return &StashList{Stash: []StashEntry{}}
}
```

---

## Sprint [N+1]: Detached HEAD + OMP Display + Return to Branch

**Date:** Sat Jan 31 2026
**Status:** ✅ COMPLETE

---

## Participants

- `@surgeon` (Complex Fix Specialist)
- `@pathfinder` (invoked for header rendering patterns)

---

## Problem Statement

1. **Detached HEAD display issue**: TIT showed "HEAD" instead of "detached" when user manually checked out a commit
2. **OMP-style display**: Match oh-my-posh compact visual style (arrows with counts, dirty indicator)
3. **Manual detached handling**: TIT had no support for user-initiated detached HEAD (only TIT time travel)
4. **Return workflow**: No proper "return to branch" flow for manual detached state

---

## Root Causes & Solutions

### 1. Detached HEAD Detection (`internal/git/state.go`)

**Problem:** `git rev-parse --abbrev-ref HEAD` returns literal "HEAD" when detached, code used it directly

**Fix:**
- Added `IsTitTimeTravel` flag to differentiate TIT time travel vs manual detached
- Set `Operation = TimeTraveling` for both cases (correct menu behavior)
- Show `"DETACHED"` in CurrentBranch (not literal "HEAD")

### 2. Header Rendering (`internal/ui/header.go` + `internal/app/app_view.go`)

**Problem:** Empty row balancing caused CWD to disappear, layout misalignment

**Fix:**
- Simplified to 2-row layout: OPS/Branch | CWD/Remote
- Manual detached: `DETACHED` + hash with separate icons
- TIT time travel: `TIME TRAVEL` + original branch name

### 3. OMP-Style Display (`internal/app/state_info.go` + `internal/app/app_view.go`)

**Problem:** TIT showed verbose text ("Local ahead", "Local behind") instead of compact arrows

**Fix:**
- Timeline: `Local ahead` → `↑ 2`, `Local behind` → `↓ 1`, `Diverged` → `↕ 2↑ 1↓`
- WorkingTree: Added `ModifiedCount` field, display `● N` (dot + count)
- Descriptions: Simplified to compact format ("2 commit(s) ahead")

### 4. Return to Branch Flow (`internal/app/dispatchers.go` + `internal/app/handlers_config.go`)

**Problem:** No workflow for returning from manual detached HEAD

**Fix:**
- `ReturnToBranchName` field in WorkflowState for target branch
- Branch picker for multiple branches, auto-switch for single branch
- Dirty tree: Stash/Discard confirmation dialog before branch picker
- Clean tree: Direct checkout

### 5. Stash-Based Merge (`internal/app/confirm_handlers.go`)

**Problem:** Return with dirty tree needed stash → merge → apply stash flow

**Fix:**
- `executeConfirmTimeTravelMergeDirtyStash()` handler
- Stash changes, merge commit, apply stash back
- Tracked via `config.AddStashEntry()` for conflict recovery

---

## Files Modified

### `internal/git/types.go`
- **Line 79**: Added `ModifiedCount int` field
- **Line 81**: Added `IsTitTimeTravel bool` field

### `internal/git/state.go`
- **Lines 147-167**: Updated detached HEAD detection with `IsTitTimeTravel` flag
- **Lines 93-98**: `detectWorkingTree()` now returns `(WorkingTree, int, error)` with modified count

### `internal/ui/header.go`
- **Lines 62-83**: Simplified 2-row left column (CWD + Remote)
- **Lines 88-104**: 2-row right column (Operation + Branch/Hash)

### `internal/app/state_info.go`
- **Lines 47-70**: Updated Timeline emojis to OMP-style (⬆, ⬇, ↕)
- **Lines 28-35**: Updated WorkingTree emoji to `●`

### `internal/app/app_view.go`
- **Lines 212-219**: Added ModifiedCount to WorkingTreeLabel (`● N`)
- **Lines 237-243**: Added OMP-style timeline labels with arrows
- **Lines 259-276**: Manual detached HEAD handling (DETACHED + hash)

### `internal/app/messages.go`
- **Lines 550-554**: Updated descriptions to compact format

### `internal/app/workflow_state.go`
- **Lines 17-18**: Added `ReturnToBranchName`, `ReturnToBranchDirtyTree` fields

### `internal/app/dispatchers.go`
- **Lines 414-490**: `dispatchTimeTravelReturn()` with branch picker for manual detached
- **Lines 493-516**: `dispatchReturnToBranchPicker()` for return workflow

### `internal/app/handlers_config.go`
- **Lines 244-275**: `handleBranchPickerEnter()` for return-from-detached with dirty tree handling

### `internal/app/confirm_handlers.go`
- **Lines 797-849**: `executeConfirmTimeTravelMergeDirtyStash()` stash-based merge flow

### `internal/app/confirm_dialog.go`
- **Line 97-100**: Added `time_travel_return_dirty_choice` confirmation type

---

## Alignment Check

- [x] LIFESTAR principles followed (state-driven, fail-fast, SSOT)
- [x] NAMING-CONVENTION.md adhered
- [x] ARCHITECTURAL-MANIFESTO principles applied (single active branch, menu=contract)

---

## Problems Solved

1. ✅ Detached HEAD shows "DETACHED" with hash, not "HEAD"
2. ✅ Manual detached HEAD supported with same menu as time travel
3. ✅ OMP-style compact display (arrows, dot+count)
4. ✅ Return to branch workflow with stash/discard choice
5. ✅ Changes preserved through stash → merge → apply flow

---

## Technical Debt / Follow-up

1. **SPEC.md Section 13.5**: Update "Detached HEAD detected" fatal error - now TIT handles manual detached
2. **ARCHITECTURE.md**: Document new detached HEAD handling (IsTitTimeTravel flag, dual-mode display)
3. **Branch picker**: ESC handling from return workflow needs verification
4. **Stash cleanup**: Old stashes not automatically cleaned up

---

## Handoff to COUNSELOR

**For ARCHITECTURE.md updates:**
- Document `IsTitTimeTravel` flag and dual-mode detached HEAD handling (Lines 147-167 of state.go)
- Document new 2-row header layout for detached state (Lines 62-104 of header.go)
- Update TimeTravelState section to include manual detached use case
- Document OMP-style display conventions (arrows, dot+count)

**For SPEC.md updates:**
- **Section 13.5 Fatal Errors**: Remove "Detached HEAD detected" fatal error - now TIT handles manual detached
- **Section 10 Time Travel**: Extend to include manual detached HEAD handling
- **Section 3 State Model**: Add `IsTitTimeTravel` field explanation
- **Section 6 State → Menu Mapping**: Add "Manual Detached" entry with menu items (Browse history, Return to branch)
- **New section**: "Manual Detached HEAD" - TIT accepts detached HEAD (not from time travel) and provides return workflow

**Key contract changes:**
- TIT no longer shows fatal error for manual detached HEAD
- Manual detached uses same `Operation = TimeTraveling` for menu (browse + return)
- Header shows `DETACHED` with commit hash (vs `TIME TRAVEL` for TIT-initiated)
- Return workflow: branch picker → stash/discard → checkout → (optional merge if dirty)
