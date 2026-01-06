# Dirty Pull Implementation — Next Phases Guide

**Date:** 2026-01-06  
**Status:** Foundation complete, ready for wiring

---

## ✅ Foundation (COMPLETE)

- [x] `internal/git/dirtyop.go` — Snapshot save/load/cleanup
- [x] `internal/app/dirtystate.go` — Operation state tracking
- [x] `internal/app/operations.go` — 5 async command phases:
  - `cmdDirtyPullSnapshot()` - Snapshot + stash/discard
  - `cmdDirtyPullMerge()` - Pull with merge strategy
  - `cmdDirtyPullRebase()` - Pull with rebase strategy
  - `cmdDirtyPullApplySnapshot()` - Reapply stashed changes
  - `cmdDirtyPullFinalize()` - Cleanup stash + snapshot file
  - `cmdAbortDirtyPull()` - Restore exact original state
- [x] Code compiles cleanly ✅

---

## 🔧 PHASE 1: State Extension

**Files:** `internal/git/types.go`, `internal/git/state.go`  
**Objective:** Add DirtyOperation state detection

### 1.1 Update `internal/git/types.go`

Add `DirtyOperation` to Operation enum:

```go
type Operation string

const (
	NotRepo       Operation = "NotRepo"
	Normal        Operation = "Normal"
	Conflicted    Operation = "Conflicted"
	Merging       Operation = "Merging"
	Rebasing      Operation = "Rebasing"
	DirtyOperation Operation = "DirtyOperation"  // ← NEW
)
```

### 1.2 Update `internal/git/state.go`

Add detection function at end of file:

```go
// detectDirtyOperation checks if a dirty operation is in progress
// by looking for the .git/TIT_DIRTY_OP snapshot file
func detectDirtyOperation() bool {
	return IsDirtyOperationActive()
}
```

Then update `DetectState()` to check for dirty operation BEFORE other states:

```go
func DetectState() (State, error) {
	// ... existing code ...
	
	// PRIORITY CHECK: DirtyOperation trumps everything except NotRepo
	if !isRepo {
		return State{Operation: NotRepo}, nil
	}
	
	// Check for dirty operation in progress
	if IsDirtyOperationActive() {
		// Return Conflicted state (dirty operation blocks all menus)
		// We reuse Conflicted because it shows the conflict resolution UI
		return State{
			Operation:      Conflicted,
			WorkingTree:    detectWorkingTree(),
			// ... other fields ...
		}, nil
	}
	
	// ... rest of existing code ...
}
```

**Testing:** Run `titest.sh` scenario 2 (dirty pull merge), verify state during operation

---

## 📋 PHASE 2: Menu & Dispatcher

**Files:** `internal/app/menu.go`, `internal/app/dispatchers.go`, `internal/app/messages.go`  
**Objective:** Add dirty pull to menu when Modified + Behind

### 2.1 Update `internal/app/menu.go`

In `menuTimeline()`, update the `Behind` case:

```go
case git.Behind:
	// NEW: If Modified, show dirty pull first
	if a.gitState.WorkingTree == git.Modified {
		items = append(items,
			Item("dirty_pull_merge").
				Shortcut("d").
				Emoji("⚠️").
				Label("Pull (save changes)").
				Hint("Save WIP, pull remote, reapply changes (may conflict)").
				Build(),
		)
		// Add separator between dirty pull and clean pull
		items = append(items, Item("").Separator().Build())
	}
	
	// EXISTING: Show clean pull options
	items = append(items,
		Item("pull_merge").
			Shortcut("p").
			Emoji("📥").
			Label("Pull (fetch + merge)").
			Hint("Fetch latest from remote and merge into local branch").
			Build(),
		Item("replace_local").
			Shortcut("f").
			Emoji("💥").
			Label("Replace local (hard reset)").
			Hint("💥 DESTRUCTIVE: Discard local commits, match remote exactly").
			Build(),
	)
```

### 2.2 Update `internal/app/dispatchers.go`

Add dispatcher function:

```go
// dispatchDirtyPull starts the dirty pull confirmation dialog
func (a *Application) dispatchDirtyPull() (tea.Model, tea.Cmd) {
	a.mode = ModeConfirmation
	
	a.confirmationState = &ConfirmationState{
		Title:        "Save your changes?",
		Description:  "You have uncommitted changes. To pull, they must be temporarily saved.\n\nAfter the pull, we'll try to reapply them.\n(This may cause conflicts if the changes overlap.)",
		LeftButtonLabel:  "Save changes",
		RightButtonLabel: "Discard changes",
		CancelButtonLabel: "Cancel",
		SelectedButton:    0, // Left button selected by default
		ActionID:         "dirty_pull",
	}
	
	return a, nil
}
```

### 2.3 Update `internal/app/messages.go`

Add new message types and prompts:

```go
// In InputPrompts map:
"dirty_pull_save":    "Save and continue with dirty pull",
"dirty_pull_discard": "Discard changes and pull",

// Add to OutputMessages:
"dirty_pull_snapshot":      "Your changes have been saved",
"dirty_pull_merge_started": "Pulling from remote (merge strategy)...",
"dirty_pull_reapply":       "Reapplying your saved changes...",
```

**Testing:** Run app, navigate to menu when Modified + Behind, verify "Pull (save changes)" appears with correct shortcut

---

## 🎯 PHASE 3: Confirmation & Operation Chain

**Files:** `internal/app/confirmationhandlers.go`, `internal/app/app.go` (Update)  
**Objective:** Wire confirmation dialog to start dirty pull operation chain

### 3.1 Update `internal/app/app.go`

Add to Application struct a field to track dirty operation state:

```go
type Application struct {
	// ... existing fields ...
	
	// Dirty operation tracking
	dirtyOperationState *DirtyOperationState  // nil when no dirty op in progress
	dirtyPullStrategy   string                // "merge" or "rebase"
}
```

### 3.2 Update `internal/app/confirmationhandlers.go`

Add handler for dirty pull confirmation:

```go
// handleDirtyPullConfirm routes confirmation choice to operation start
func (a *Application) handleDirtyPullConfirm(choice string) (tea.Model, tea.Cmd) {
	if a.confirmationState == nil || a.confirmationState.ActionID != "dirty_pull" {
		return a, nil
	}

	preserveChanges := choice == "yes" // Save if yes, discard if no

	// Create operation state
	strategy := a.dirtyPullStrategy // Should be set before confirmation
	if strategy == "" {
		strategy = "merge" // Default to merge
	}
	
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_"+strategy, preserveChanges)
	a.dirtyOperationState.PullStrategy = strategy

	// Transition to console to show streaming output
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.asyncOperationActive = true
	a.previousMode = ModeMenu

	// Start the operation chain
	return a, a.cmdDirtyPullSnapshot(preserveChanges)
}
```

### 3.3 Update `internal/app/handlers.go`

In the main keyhandler for confirmation (likely `handleKeyEnter`), add routing:

```go
// When mode is ModeConfirmation, route based on ActionID
if a.mode == ModeConfirmation && a.confirmationState != nil {
	switch a.confirmationState.ActionID {
	case "dirty_pull":
		selectedChoice := "no"
		if a.confirmationState.SelectedButton == 0 { // Left button
			selectedChoice = "yes"
		}
		return a.handleDirtyPullConfirm(selectedChoice)
	
	// ... other confirmation actions ...
	}
}
```

### 3.4 Update `internal/app/githandlers.go` (or operations flow)

Add operation chain handler for GitOperationMsg:

```go
case GitOperationMsg:
	return a.handleGitOperationMsg(msg)
```

Then in `handleGitOperationMsg`, add dirty pull routing:

```go
func (a *Application) handleGitOperationMsg(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	if a.dirtyOperationState == nil {
		// Regular operation, not dirty pull
		// ... existing code ...
	}

	// Dirty operation flow
	switch a.dirtyOperationState.Phase {
	case "snapshot":
		if !msg.Success {
			a.footerHint = "Failed: " + msg.Error
			a.asyncOperationActive = false
			return a, nil
		}
		a.dirtyOperationState.SetPhase("apply_changeset")
		
		// Continue with pull
		if a.dirtyOperationState.PullStrategy == "rebase" {
			return a, a.cmdDirtyPullRebase()
		}
		return a, a.cmdDirtyPullMerge()

	case "apply_changeset":
		if !msg.Success {
			// Conflicts in changeset apply - show conflict resolver
			// TODO: Call setupConflictResolver with 3-way view
			a.dirtyOperationState.MarkConflictDetected("changeset_apply", []string{})
			a.mode = ModeConflictResolve
			return a, nil
		}
		a.dirtyOperationState.SetPhase("apply_snapshot")
		
		// Continue with snapshot reapply
		return a, a.cmdDirtyPullApplySnapshot()

	case "apply_snapshot":
		if !msg.Success {
			// Conflicts in snapshot apply - show conflict resolver
			a.dirtyOperationState.MarkConflictDetected("snapshot_reapply", []string{})
			a.mode = ModeConflictResolve
			return a, nil
		}
		a.dirtyOperationState.SetPhase("finalizing")
		
		// Finalize
		return a, a.cmdDirtyPullFinalize()

	case "finalizing":
		if !msg.Success {
			a.footerHint = "Warning: " + msg.Error
		} else {
			a.footerHint = "✓ Dirty pull completed"
		}
		a.asyncOperationActive = false
		a.dirtyOperationState = nil
		return a, nil
	}

	return a, nil
}
```

**Testing:** Run app, select "Pull (save changes)", confirm dialog, watch operation chain progress in console

---

## 🧩 PHASE 4: Conflict Integration

**Files:** `internal/app/conflicthandlers.go`, `internal/ui/conflictresolver.go` (integrate)  
**Objective:** Wire conflict resolver to dirty pull phases

### 4.1 Implement conflict file setup

When conflict is detected (either phase), populate conflict state:

```go
// In handleGitOperationMsg, when conflicts detected:
if !msg.Success && strings.Contains(msg.Error, "conflict") {
	conflictFiles, err := ReadConflictFiles()
	if err != nil {
		a.footerHint = "Error reading conflicts: " + err.Error()
		return a, nil
	}
	
	a.conflictResolveState = &ConflictResolveState{
		Operation:      "dirty_pull_" + a.dirtyOperationState.ConflictPhase,
		Files:          conflictFiles,
		NumColumns:     3, // LOCAL / REMOTE / SNAPSHOT (for dirty pull)
		ColumnLabels:   []string{"LOCAL", "REMOTE", "SNAPSHOT"},
		SelectedFileIndex: 0,
		FocusedPane:    0,
		ScrollOffsets:  make([]int, 3),
		LineCursors:    make([]int, 3),
	}
	
	a.mode = ModeConflictResolve
	a.asyncOperationActive = false
}
```

### 4.2 Update `handleConflictEnter()` in `conflicthandlers.go`

Implement the stub to route based on operation:

```go
func (a *Application) handleConflictEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve || app.conflictResolveState == nil {
		return app, nil
	}

	// Check if all files have been marked (resolution complete)
	allMarked := true
	for _, file := range app.conflictResolveState.Files {
		if file.Chosen < 0 {
			allMarked = false
			break
		}
	}

	if !allMarked {
		app.footerHint = "Mark all files with SPACE before continuing"
		return app, nil
	}

	// Apply the chosen versions
	for i, file := range app.conflictResolveState.Files {
		chosenIdx := file.Chosen
		chosenContent := file.Versions[chosenIdx]
		
		// Write the resolved content
		if err := os.WriteFile(file.Path, []byte(chosenContent), 0644); err != nil {
			app.footerHint = "Error writing resolved file: " + err.Error()
			return app, nil
		}
		
		// Stage the resolved file
		git.Execute("add", file.Path)
		
		app.conflictResolveState.Files[i].Chosen = chosenIdx // Mark as applied
	}

	// Route based on operation
	switch app.conflictResolveState.Operation {
	case "dirty_pull_changeset_apply":
		// Continue with snapshot reapply
		app.asyncOperationActive = true
		app.mode = ModeConsole
		return app, app.cmdDirtyPullApplySnapshot()
	
	case "dirty_pull_snapshot_reapply":
		// Continue to finalize
		app.asyncOperationActive = true
		app.mode = ModeConsole
		return app, app.cmdDirtyPullFinalize()
	
	default:
		// Unknown operation
		return app.returnToMenu()
	}
}
```

### 4.3 Update `handleConflictEsc()` in `conflicthandlers.go`

Implement abort:

```go
func (a *Application) handleConflictEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve {
		return app, nil
	}

	// Check if visual mode is active in diff pane
	if app.conflictResolveState != nil && 
	   app.conflictResolveState.DiffPane != nil && 
	   app.conflictResolveState.DiffPane.VisualModeActive {
		app.conflictResolveState.DiffPane.VisualModeActive = false
		return app, nil
	}

	// Route based on operation for proper abort
	if app.conflictResolveState != nil {
		switch app.conflictResolveState.Operation {
		case "dirty_pull_changeset_apply", "dirty_pull_snapshot_reapply":
			// Abort dirty pull
			app.asyncOperationActive = true
			app.mode = ModeConsole
			return app, app.cmdAbortDirtyPull()
		}
	}

	// Default: return to menu
	return app.returnToMenu()
}
```

**Testing:** Trigger conflict scenarios using titest.sh scenarios 2-4, verify conflict resolver appears, mark files, press ENTER to continue, press ESC to abort

---

## 📖 PHASE 5: Helper Functions

**File:** `internal/git/execute.go`  
**Objective:** Add conflict file reading

### 5.1 Add `ReadConflictFiles()` function

```go
// ReadConflictFiles reads git conflict markers and returns structured conflict data
// Used by conflict resolver UI to show 3-way comparison
func ReadConflictFiles() ([]ui.ConflictFileGeneric, error) {
	// Get list of conflicted files
	result := Execute("status", "--porcelain=v2")
	if !result.Success {
		return nil, fmt.Errorf("failed to list conflicted files")
	}

	var files []ui.ConflictFileGeneric
	scanner := bufio.NewScanner(strings.NewReader(result.Stdout))

	for scanner.Scan() {
		line := scanner.Text()
		// Parse status line: look for "u" in field 2 (unmerged)
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}

		if parts[1] != "u" { // Not unmerged
			continue
		}

		filePath := parts[8] // File path is last field

		// Get the three versions
		var versions []string

		// Stage 1 (base)
		base := execute("git", "show", ":1:"+filePath)
		versions = append(versions, base)

		// Stage 2 (local/ours)
		local := execute("git", "show", ":2:"+filePath)
		versions = append(versions, local)

		// Stage 3 (remote/theirs)
		remote := execute("git", "show", ":3:"+filePath)
		versions = append(versions, remote)

		files = append(files, ui.ConflictFileGeneric{
			Path:     filePath,
			Versions: versions,
			Chosen:   -1, // Not marked yet
		})
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no conflicted files found")
	}

	return files, nil
}
```

**Testing:** Manually test conflict detection works correctly

---

## 🧪 PHASE 6: Integration Testing

Use `titest.sh` to test each scenario:

```bash
cd /Users/jreng/Documents/Poems/inf/tit_test_repo
../../titest.sh

# Scenario 0: Reset to fresh state
# Scenario 5: Clean dirty pull (no conflicts)
# - Select "Pull (save changes)"
# - Confirm: "Save changes"
# - Watch operation complete
# - Return to menu automatically

# Scenario 2: Dirty pull with merge conflicts
# - See "Pull (save changes)" in menu
# - Confirm: "Save changes"
# - Hit merge conflicts
# - Conflict resolver shows 3 columns
# - Mark files, press ENTER to continue
# - Snapshot reapply succeeds
# - Operation completes

# Scenario 3: Dirty pull with stash apply conflicts
# - Similar flow but conflicts on reapply
# - Mark files, continue
# - Finalize and cleanup
```

---

## 📊 Operation Chain Flow

```
Menu (dirty_pull_merge selected)
  ↓
dispatchDirtyPull()
  ↓
ConfirmationDialog (Save? / Discard? / Cancel)
  ↓ (Yes / No)
handleDirtyPullConfirm()
  ↓
cmdDirtyPullSnapshot(preserve)
  ↓ Success
cmdDirtyPullMerge() OR cmdDirtyPullRebase()
  ↓ Conflict?
  ├─ YES → setupConflictResolver(operation="dirty_pull_changeset_apply")
  │         → ModeConflictResolve
  │         → User marks files + ENTER
  │         → handleConflictEnter() continues
  │
  └─ NO → cmdDirtyPullApplySnapshot()
          ↓ Conflict?
          ├─ YES → setupConflictResolver(operation="dirty_pull_snapshot_reapply")
          │         → User marks files + ENTER
          │         → handleConflictEnter() continues
          │
          └─ NO → cmdDirtyPullFinalize()
                  ↓
                  Return to Menu (Operation=Normal)
```

---

## ✨ Ready to Build

All components exist and compile cleanly. Next steps:
1. Implement Phase 1 (state detection)
2. Implement Phase 2 (menu + dispatcher)
3. Implement Phase 3 (confirmation + chain)
4. Implement Phase 4 (conflict integration)
5. Add Phase 5 helper (ReadConflictFiles)
6. Test Phase 6 with titest.sh scenarios

Each phase is independent and can be tested incrementally.

