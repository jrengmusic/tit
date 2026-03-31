# PLAN: Branch Operations (New Branch + Merge From)

**Date:** 2026-03-31
**Status:** COMPLETE
**SPEC Reference:** Sections 8 (Branch Switching) and 9 (Merge Branch Assistance)

---

## Context

ARCHITECT's workflow is trunk-based dev with a permanent dev branch as staging gate:
- main = always shippable
- dev = where real work happens
- Merge to main = deliberate promotion when dev is stable

Two new Config menu items needed:
1. **New branch** — create and switch to new branch from current HEAD
2. **Merge from...** — merge another branch into current (visible when 2+ local branches)

---

## Feature 1: New Branch

### Step 1.1 — Menu item registration

**File:** `internal/app/menu_items.go`
- Add `"config_new_branch"` entry to `MenuItems` map
- Shortcut, emoji, label, hint following existing config item pattern

**Validation:** Entry exists in map, `GetMenuItem("config_new_branch")` does not panic.

### Step 1.2 — Config menu wiring

**File:** `internal/app/menu_render_extra.go`
- Add `config_new_branch` to `GenerateConfigMenu()`, placed before `config_switch_branch`

**Validation:** Config menu renders with new item visible.

### Step 1.3 — Dispatcher registration

**File:** `internal/app/dispatchers.go`
- Add `"config_new_branch": a.dispatchConfigNewBranch` to action dispatch table

### Step 1.4 — Dispatcher implementation

**File:** `internal/app/dispatch_dialog.go`
- New function `dispatchConfigNewBranch(app *Application) tea.Cmd`
- Transitions to `ModeInput` with:
  - `InputAction: "new_branch_name"`
  - `Prompt: "Branch name"`
  - Single-line height
- Pattern: follows existing text input dispatch (commit message, init branch name)

**Validation:** Selecting "New branch" opens text input with prompt.

### Step 1.5 — Input submit handler

**File:** `internal/app/handlers_input.go`
- Add case `"new_branch_name"` in input submit routing
- Validate: name not empty, valid git ref, branch doesn't already exist
- On valid: call `a.prepareAsyncOperation("Creating branch...")`, return `a.cmdCreateBranch(branchName)`

**Validation:** Entering a valid name triggers branch creation.

### Step 1.6 — Git command

**File:** `internal/app/handlers_git_branch.go`
- New function `cmdCreateBranch(branchName string) tea.Cmd`
- Uses `git.ExecuteWithStreaming(ctx, "checkout", "-b", branchName)`
- Returns `GitOperationMsg{Step: "branch_create", Success: bool}`
- Pattern: follows `cmdSwitchBranch`

### Step 1.7 — Operation result handler

**File:** `internal/app/git_handlers.go`
- Add case `"branch_create"` in `handleGitOperation`
- On success: refresh state, return to menu
- On failure: show error in footer

**Validation:** Branch created, switched to it, state refreshed, menu shows new branch as current.

---

## Feature 2: Merge From

### Step 2.1 — Menu item registration

**File:** `internal/app/menu_items.go`
- Add `"config_merge_branch"` entry to `MenuItems` map

**Validation:** Entry exists in map, `GetMenuItem("config_merge_branch")` does not panic.

### Step 2.2 — Config menu wiring (conditional)

**File:** `internal/app/menu_render_extra.go`
- Add `config_merge_branch` to `GenerateConfigMenu()`
- Condition: only show when 2+ local branches exist
- Placed after `config_switch_branch`
- Use same conditional pattern as remote menu items

**Validation:** Item hidden with 1 branch, visible with 2+ branches.

### Step 2.3 — Branch picker purpose flag

**File:** `internal/app/workflow_state.go` (or wherever `WorkflowState` is defined)
- Add `BranchPickerPurpose string` field to `WorkflowState`
- Constants: `"switch"`, `"merge"`, `"return"`
- Existing switch flow sets purpose to `"switch"` (or empty, backward-compatible)

**Validation:** Existing branch switch still works identically.

### Step 2.4 — Dispatcher registration

**File:** `internal/app/dispatchers.go`
- Add `"config_merge_branch": a.dispatchConfigMergeBranch` to action dispatch table

### Step 2.5 — Dispatcher implementation

**File:** `internal/app/dispatch_dialog.go`
- New function `dispatchConfigMergeBranch(app *Application) tea.Cmd`
- Loads branches via `git.ListBranchesWithDetails()`
- Filters out current branch (cannot merge current into current)
- Populates `BranchPickerState` with filtered list
- Sets `workflowState.BranchPickerPurpose = "merge"`
- Sets `mode = ModeBranchPicker`
- Footer hint: "Select branch to merge into [current_branch]"

**Validation:** Picker opens with current branch excluded from list.

### Step 2.6 — Branch picker enter handler (merge path)

**File:** `internal/app/handlers_config_branch.go`
- In `handleBranchPickerEnter`, branch on `workflowState.BranchPickerPurpose == "merge"`
- On merge path:
  - If working tree dirty: show dirty tree confirmation (Dirty Operation Protocol per SPEC Section 7)
  - If working tree clean: show merge confirmation dialog

### Step 2.7 — Confirmation dialog messages

**File:** `internal/app/messages_dialog.go`
- Add confirmation type constant `"confirm_merge_branch"`
- Add entry in `ConfirmationMessages` map:
  - Title: "Merge [source] into [current]"
  - Explanation: describes what will happen
  - YesLabel: "Merge" / NoLabel: "Cancel"

### Step 2.8 — Confirmation handler registration

**File:** `internal/app/confirm_dialog_handlers.go`
- Add `"confirm_merge_branch"` to `confirmationHandlers` map
- Confirm: `executeMergeBranch`
- Reject: return to menu

### Step 2.9 — Merge confirmation handler

**File:** `internal/app/confirm_branch.go` (or new `confirm_merge.go`)
- `executeMergeBranch()`:
  - Read source branch from dialog context
  - `prepareAsyncOperation("Merging [source] into [current]...")`
  - Return `cmdMergeBranch(sourceBranch)`

### Step 2.10 — Git merge command

**File:** `internal/app/op_merge.go`
- New function `cmdMergeBranch(sourceBranch string) tea.Cmd`
- Uses `git.ExecuteWithStreaming(ctx, "merge", sourceBranch)`
- Returns `GitOperationMsg{Step: "merge_branch", Success: bool}`
- Pattern: follows `cmdSwitchBranch` streaming pattern

**Validation:** Merge executes with streaming output visible in console.

### Step 2.11 — Operation result handler

**File:** `internal/app/git_handlers.go`
- Add case `"merge_branch"` in `handleGitOperation`
- On success: refresh state, return to menu (main is now Ahead if remote exists — push naturally appears)
- On failure (conflicts): enter conflict resolver (reuse existing `setupConflictResolver` pattern)

### Step 2.12 — Merge conflict resolution

**File:** `internal/app/op_merge.go`
- New function `cmdFinalizeBranchMerge()` — stages resolved files, commits merge
- Reuse `cmdAbortMerge()` for abort path (already exists)
- Wire into conflict resolver's finalize/abort callbacks

**Validation:** Conflicts detected, resolver opens, user can resolve or abort.

### Step 2.13 — Dirty tree handling (Dirty Operation Protocol)

**File:** `internal/app/confirm_branch.go` (or `confirm_merge.go`)
- When merge requested with dirty tree:
  - Show confirmation: "Stash changes and merge" / "Cancel"
  - On confirm: stash -> merge -> reapply stash (SPEC Section 7 pattern)
  - Reuse existing dirty operation infrastructure from pull flow

**Validation:** Dirty tree stashed before merge, reapplied after.

---

## Operation step constants to add

**File:** `internal/app/operation_steps.go`
- `OpBranchCreate = "branch_create"`
- `OpMergeBranch = "merge_branch"`
- `OpFinalizeBranchMerge = "finalize_branch_merge"`

---

## Execution Order

Feature 1 first (simpler, fewer moving parts), then Feature 2.

Within each feature: steps are sequential. Each step validated before proceeding to next.

---

## Key Files Summary

| File | Changes |
|------|---------|
| `internal/app/menu_items.go` | 2 new menu item entries |
| `internal/app/menu_render_extra.go` | Config menu additions + conditional |
| `internal/app/dispatchers.go` | 2 new dispatcher registrations |
| `internal/app/dispatch_dialog.go` | 2 new dispatcher functions |
| `internal/app/handlers_input.go` | New branch name submit case |
| `internal/app/handlers_git_branch.go` | `cmdCreateBranch` |
| `internal/app/handlers_config_branch.go` | Merge path in picker enter handler |
| `internal/app/workflow_state.go` | `BranchPickerPurpose` field |
| `internal/app/messages_dialog.go` | Merge confirmation messages |
| `internal/app/confirm_dialog_handlers.go` | Merge confirmation handler registration |
| `internal/app/confirm_branch.go` or `confirm_merge.go` | Merge confirmation + dirty tree handler |
| `internal/app/op_merge.go` | `cmdMergeBranch`, `cmdFinalizeBranchMerge` |
| `internal/app/git_handlers.go` | 2 new operation result cases |
| `internal/app/operation_steps.go` | 3 new constants |
| `internal/app/modes.go` | No new modes needed (reuses existing) |
