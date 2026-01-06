# Dirty Pull Implementation — Component Audit

**Date:** 2026-01-06  
**Status:** Ready for Implementation

---

## ✅ EXISTING COMPONENTS (Ready to Use)

### 1. State Model (`internal/git/types.go`)
**Status:** ✅ Complete
- `WorkingTree` enum: `Clean` | `Modified` ✅
- `Timeline` enum: `InSync` | `Ahead` | `Behind` | `Diverged` | `NoRemote` ✅
- `Operation` enum: `NotRepo` | `Normal` | `Conflicted` | `Merging` | `Rebasing` ✅
- `Remote` enum: `NoRemote` | `HasRemote` ✅
- `State` struct with `CurrentBranch`, `CurrentHash` ✅

**TODO:** Add `DirtyOperation` to Operation enum (use `Conflicted` as umbrella state)

---

### 2. Conflict Resolver UI (`internal/ui/conflictresolver.go`)
**Status:** ✅ Complete, Generic N-Column Model
- Renders 2N-column layout (top row: file lists, bottom row: content)
- Supports any number of columns (tested with 2, ready for 3)
- `ConflictFileGeneric` struct: `Path`, `Versions[]`, `Chosen int` (radio button)
- Focus-based border colors, shared file selection, per-pane scrolling
- **Ready for dirty pull:** 3-column LOCAL/REMOTE/SNAPSHOT or LOCAL/REMOTE/INCOMING

---

### 3. Conflict State Struct (`internal/app/conflictstate.go`)
**Status:** ✅ Complete, Designed for Dirty Pull
```go
type ConflictResolveState struct {
    Operation      string  // "dirty_pull", "cherry_pick", "external_conflict"
    CommitHash     string  // For display
    IsRebase       bool    // Tracks pull strategy
    StashNeedsDrop bool    // Cleanup flag after snapshot apply
    
    // N-column generic model ready
    Files              []ui.ConflictFileGeneric
    SelectedFileIndex  int
    FocusedPane        int
    NumColumns         int
    ColumnLabels       []string  // Ready: ["LOCAL", "REMOTE", "INCOMING"]
    ScrollOffsets      []int     // Per-column bottom pane scrolling
    LineCursors        []int     // Per-column line cursors
}
```

---

### 4. Conflict Handlers (`internal/app/conflicthandlers.go`)
**Status:** ⚠️ PARTIALLY IMPLEMENTED
- ✅ Navigation: `handleConflictUp/Down` (file list + content scroll)
- ✅ Tab cycling: `handleConflictTab` (all panes)
- ✅ Selection: `handleConflictSpace` (radio button marking)
- ❌ Enter handler: STUB ("not yet implemented")
- ❌ ESC handler: STUB ("not yet implemented")

**TODO:** Implement `handleConflictEnter` and `handleConflictEsc`

---

### 5. Menu System (`internal/app/menu.go`)
**Status:** ✅ Complete Base Structure
- Menu generators per operation state ✅
- `menuNormal()` with working tree + timeline sections ✅
- Timeline Behind case already shows `pull_merge` + `pull_rebase` ✅
- Infrastructure ready for dirty pull items

**TODO:** Add dirty pull menu items in `menuTimeline()` when `Modified + Behind`

---

### 6. Confirmation Dialog (`internal/ui/confirmation.go`, `internal/app/confirmationhandlers.go`)
**Status:** ✅ Complete, Ready
- Dialog system with button navigation ✅
- `showConfirmation()` helper ✅
- Button handler integration ✅
- **Ready for:** "Save changes and proceed" / "Discard" / "Cancel" dialog

---

### 7. Messages & Prompts (`internal/app/messages.go`)
**Status:** ✅ Extensible
- `GitOperationMsg` struct (reusable) ✅
- `FooterMessageType` enum with centralized text ✅
- `InputPrompts` / `InputHints` / `ErrorMessages` maps ✅
- `OutputMessages` for operation feedback ✅

**TODO:** Add dirty pull message types and hints

---

### 8. Application Modes (`internal/app/modes.go`)
**Status:** ✅ Ready
- `ModeConflictResolve` already defined ✅
- Clean mode routing infrastructure ✅
- No need for separate dirty operation mode (reuse `ModeConflictResolve`)

---

## 🔴 MISSING COMPONENTS (Must Implement)

### 1. Stash Management (`internal/git/dirtyop.go`) — NEW FILE
**Priority:** HIGH

```go
package git

type DirtyOperationSnapshot struct {
    OriginalBranch string
    OriginalHead   string
}

func (s *DirtyOperationSnapshot) Save(branchName, headHash string) error
func (s *DirtyOperationSnapshot) Load() error
func (s *DirtyOperationSnapshot) Delete() error
func (s *DirtyOperationSnapshot) FilePath() string
```

**Responsibilities:**
- Write `.git/TIT_DIRTY_OP` with branch + hash
- Read snapshot data back
- Clean up on finalize/abort

---

### 2. Dirty Operation Dispatchers (`internal/app/dispatchers.go`)
**Priority:** HIGH

**New functions:**
- `dispatchDirtyPull()` → confirm dialog
- `dispatchDirtyMerge()` → confirm dialog (for future Phase)
- `dispatchDirtyTimeTravel()` → confirm dialog (for future Phase)

---

### 3. Dirty Pull Confirmation Handler (`internal/app/confirmationhandlers.go`)
**Priority:** HIGH

**New function:**
- `handleDirtyPullConfirm(choice)` 
  - "Save": Start `executeDirtyPullPreserving()`
  - "Discard": Start `executeDirtyPullDiscarding()`
  - "Cancel": Return to menu

---

### 4. Dirty Pull Execution (`internal/app/operations.go`)
**Priority:** CRITICAL

**New async commands:**
- `cmdDirtyPullSnapshot(preserveChanges)` → creates stash + saves snapshot
- `cmdDirtyPullMerge()` → `git pull origin`
- `cmdDirtyPullApplySnapshot()` → `git stash apply`
- `cmdDirtyPullFinalize()` → `git stash drop`
- `cmdAbortDirtyPull()` → `git reset --hard + git stash apply + cleanup`

**Execution chain:**
```
1. Snapshot (async cmd)
   ↓ Success → 2. Pull (async cmd)
   ↓ Error → return to menu with error
   
2. Pull (async cmd)
   ↓ No conflicts → 3. Apply Snapshot
   ↓ Conflicts → ModeConflictResolve (operation="dirty_pull_changeset")
   
3. Apply Snapshot (async cmd)
   ↓ No conflicts → 4. Finalize
   ↓ Conflicts → ModeConflictResolve (operation="dirty_pull_snapshot")
   
4. Finalize (async cmd)
   ↓ Complete → Return to menu with success
```

---

### 5. Conflict Handler Entry Points (`internal/app/conflicthandlers.go`)
**Priority:** HIGH

**Update stubs:**
- `handleConflictEnter()` → Route based on `operation` field
  - `"dirty_pull_changeset"` → continue to snapshot apply
  - `"dirty_pull_snapshot"` → finalize
  - `"cherry_pick"` → (future)

- `handleConflictEsc()` → Route abort based on `operation` field
  - `"dirty_pull_*"` → call `cmdAbortDirtyPull()`
  - Clear snapshot file
  - Return to menu

---

### 6. Menu Integration (`internal/app/menu.go`)
**Priority:** MEDIUM

**Update `menuTimeline()`:**
```go
case git.Behind:
    // If Modified: show dirty pull first
    if a.gitState.WorkingTree == git.Modified {
        items = append(items,
            Item("dirty_pull_merge").
                Shortcut("d").
                Emoji("⚠️").
                Label("Pull (save changes)").
                Hint("Save WIP, pull remote, reapply changes (may conflict)").
                Build(),
        )
    }
    // Then show clean pull options
    items = append(items,
        Item("pull_merge"). ... // existing
    )
```

---

### 7. Conflict File Reading (`internal/git/execute.go`)
**Priority:** MEDIUM

**New function needed:**
```go
func ReadConflictFiles(operation string) ([]ui.ConflictFileGeneric, error)
```

**Logic:**
- Run `git status --porcelain=v2` → find unmerged files
- For each file:
  - `git show :1:<file>` → base version
  - `git show :2:<file>` → local version
  - `git show :3:<file>` → remote version
- Format as `ConflictFileGeneric`

---

## 📋 IMPLEMENTATION CHECKLIST

### Phase 1: State Extension ⚙️
- [ ] Add `DirtyOperation` to `git.Operation` enum
- [ ] Add `DetectDirtyOperation()` to `git/state.go`
- [ ] Create `internal/git/dirtyop.go` with snapshot save/load/delete

### Phase 2: Menu & Dispatcher 📋
- [ ] Add dirty pull items to `menuTimeline()` (Behind + Modified)
- [ ] Create `dispatchDirtyPull()` in dispatchers.go
- [ ] Create confirmation dialog handler

### Phase 3: Async Operations 🔧
- [ ] Create snapshot async command (`cmdDirtyPullSnapshot()`)
- [ ] Create pull async command (`cmdDirtyPullMerge()`)
- [ ] Create apply snapshot command (`cmdDirtyPullApplySnapshot()`)
- [ ] Create finalize command (`cmdDirtyPullFinalize()`)
- [ ] Create abort command (`cmdAbortDirtyPull()`)
- [ ] Implement operation chaining with error handling

### Phase 4: Conflict Integration 🎯
- [ ] Implement `handleConflictEnter()` for dirty pull
- [ ] Implement `handleConflictEsc()` for dirty pull abort
- [ ] Create `ReadConflictFiles()` helper
- [ ] Wire conflict file detection into operation chain

### Phase 5: Testing 🧪
- [ ] Test scenario 1: Clean pull (no conflicts)
- [ ] Test scenario 2: Merge with changeset conflicts
- [ ] Test scenario 3: Merge with snapshot apply conflicts
- [ ] Test scenario 4: Abort at different phases
- [ ] Test snapshot cleanup on abort

---

## 🎯 Testing with titest.sh

Once all components are ready:

```bash
# Terminal 1: Run test scenario setup
./titest.sh
# Select scenario (1-5)

# Terminal 2: Run TIT
./tit_x64

# Test dirty pull flow:
# - See menu with dirty pull option
# - Confirm dialog appears
# - Operation snapshots changes
# - Conflict resolver opens (if conflicts)
# - Resolve conflicts
# - Snapshot reapplied
# - Success or abort
```

---

## 📊 Component Dependency Graph

```
Menu (menuTimeline) 
    ↓ dispatch
Confirmation Dialog
    ↓ confirm (Save/Discard/Cancel)
Async Ops Chain (Snapshot → Pull → Apply → Finalize)
    ↓ if conflicts
ModeConflictResolve (with operation="dirty_pull_*")
    ↓ ENTER/ESC
Conflict Handler (continue or abort)
    ↓
Back to menu or next operation
```

---

## ✨ Ready to Build

All foundational components exist. No architectural changes needed.

**Next steps:**
1. Implement missing files (dirtyop.go)
2. Wire menu + dispatchers
3. Implement async operation chain
4. Wire conflict handlers
5. Test with titest.sh scenarios

