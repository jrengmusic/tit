# TIT Implementation Plan — Port Old TIT to New TIT

**Goal:** Port proven patterns from old TIT to new TIT's superior foundation.

**Philosophy:**
- Incremental: Each phase tested before moving to next
- Simple: No overengineering, no branch tracking
- Safe: Every operation abortable, state preserved
- Beautiful: Leverage new TIT's layout system

---

## Phase 1: Simplify State Model

**Remove canon/working dual-branch architecture, back to 4-axis model.**

### 1.1 Remove BranchContext from State

**Files:**
- `internal/git/types.go`
- `internal/git/state.go`

**Changes:**
1. Remove `BranchContext` enum from `types.go`
2. Remove `CanonBranch`, `WorkingBranch`, `IsCanonBranch`, `IsWorkingBranch` fields from `State` struct
3. Remove `BranchContext` detection logic from `DetectState()`
4. Keep 4-axis model: `(WorkingTree, Timeline, Operation, Remote)`

**Test:**
```bash
./build.sh
./tit_x64  # or tit_arm64
# Verify state detection works (check header display)
# Menu should show init/clone if not in repo
```

**Expected result:**
- Header shows: `Branch: main | Clean | In sync`
- No canon/working classification
- State detection compiles without errors

---

### 1.2 Remove Config File Tracking

**Files:**
- `internal/git/config.go`
- `internal/app/handlers.go`

**Changes:**
1. Remove `SaveRepoConfig()` calls from init/clone handlers
2. Remove `LoadRepoConfig()` from app startup
3. Keep config file code (might be useful later), just don't use it
4. State always reflects actual Git state (no tracking)

**Test:**
```bash
# In test repo
./tit_x64
# Check that TIT works without reading ~/.config/tit/repo.toml
# State should be detected from actual git commands
```

**Expected result:**
- TIT starts without config file
- Shows correct branch/state from Git

---

### 1.3 Update Menu Generation

**Files:**
- `internal/app/menu.go`

**Changes:**
1. Remove `menuCanonBranch()`, `menuWorkingBranch()` functions
2. Simplify `GenerateMenu()` to single branch model:
   ```go
   switch gitState.Operation {
   case git.NotRepo:
       return menuNotRepo()
   case git.Conflicted:
       return menuConflicted()
   case git.Merging, git.Rebasing:
       return menuOperation()
   case git.TimeTraveling:
       return menuTimeTraveling()
   case git.Normal:
       return menuNormal()  // Single menu for all branches
   }
   ```

**Test:**
```bash
./tit_x64
# Verify menu shows based on operation state only
# All branch operations available on any branch
```

**Expected result:**
- Menu shows commit/push options when Modified
- Menu shows sync options based on Timeline
- No branch-specific restrictions

---

## Phase 2: Core Git Operations

**Port working patterns from old TIT: commit, push, pull.**

### 2.1 Commit Operation

**Files:**
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`
- `internal/app/menu.go`

**Pattern from old TIT:**
- Show only when WorkingTree = Modified
- Commit all changes (staged + unstaged)
- Stream output to console

**Changes:**
1. Add menu item in `menuNormal()`:
   ```go
   if a.gitState.WorkingTree == git.Modified {
       items = append(items, MenuItem{
           ID: "commit",
           Shortcut: "c",
           Label: "Commit changes",
           Hint: "Commit all changes to current branch",
       })
   }
   ```

2. Add dispatcher: `dispatchCommit()`
3. Add handler: `handleCommitSubmit()` with async execution
4. Use existing console mode for streaming output

**Test:**
```bash
# Make changes in test repo
echo "test" > test.txt
./tit_x64
# Menu should show "Commit changes"
# Press 'c', enter commit message, verify commit succeeds
git log  # Verify commit exists
```

**Expected result:**
- Commit succeeds
- Console shows git output
- State updates to Clean after commit

---

### 2.2 Push Operation

**Files:**
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`
- `internal/app/menu.go`

**Pattern from old TIT:**
- Show only when Timeline = Ahead
- Simple push to upstream
- Handle auth failures gracefully

**Changes:**
1. Add menu item when `Timeline == git.Ahead`
2. Add `dispatchPush()` dispatcher
3. Execute: `git push` (async with streaming)

**Test:**
```bash
# After commit in 2.1
./tit_x64
# Menu should show "Push to remote"
# Execute push, verify in console
```

**Expected result:**
- Push succeeds
- Timeline updates to InSync
- Console shows git output

---

### 2.3 Pull Operation (Merge)

**Files:**
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`
- `internal/app/menu.go`

**Pattern from old TIT:**
- Show when Timeline = Behind or InSync
- Offer merge or rebase options
- Handle conflicts with conflict resolution mode

**Changes:**
1. Add menu items:
   - "Pull (merge)" when Behind
   - "Pull (rebase)" when Behind
2. Execute: `git pull` or `git pull --rebase`
3. If conflicts → set `Operation = Conflicted`
4. Use existing conflict detection

**Test:**
```bash
# Simulate behind state:
# 1. Clone repo twice
# 2. Commit in first clone, push
# 3. Open second clone with TIT
./tit_x64
# Should show "Pull (merge)"
# Execute, verify merge completes
```

**Expected result:**
- Pull succeeds
- Timeline updates to InSync
- If conflicts, enters conflict mode

---

## Phase 3: Branch Operations

**Add branch switching, creation, and merge assistance.**

### 3.1 List and Switch Branches

**Files:**
- `internal/app/modes.go` (add `ModeBranchSelect`)
- `internal/ui/branchselect.go` (new file, rendering)
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`

**Pattern:**
- Menu item: "Switch branch"
- Shows list of local branches (from `git branch`)
- Highlight current branch
- Select and checkout

**Changes:**
1. Add `ModeBranchSelect` to modes enum
2. Create branch selection UI (similar to clone branch selection)
3. Use `git.ListBranches()` (already exists)
4. Execute: `git checkout <branch>`
5. Reload state after switch

**Test:**
```bash
# In repo with multiple branches
git checkout -b dev
git checkout main
./tit_x64
# Select "Switch branch"
# Should show main (current), dev
# Switch to dev, verify header updates
```

**Expected result:**
- Branch list shows correctly
- Switch succeeds
- Header updates to new branch
- Menu regenerates for new branch state

---

### 3.2 Create New Branch

**Files:**
- `internal/app/modes.go` (add `ModeCreateBranch`)
- `internal/ui/textinput.go` (reuse existing)
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`

**Pattern:**
- Menu item: "Create new branch"
- Prompt for branch name
- Validate name
- Execute: `git checkout -b <name>`

**Changes:**
1. Add menu item in `menuNormal()`
2. Add `dispatchCreateBranch()` dispatcher
3. Add input validation (git ref name rules)
4. Execute branch creation (async)

**Test:**
```bash
./tit_x64
# Select "Create new branch"
# Enter "feature/test"
# Verify branch created and switched
git branch  # Should show feature/test (current)
```

**Expected result:**
- Branch created
- Automatically switched to new branch
- Header shows new branch name

---

### 3.3 Merge Branch into Current

**Files:**
- `internal/app/modes.go` (add `ModeMergeBranch`)
- `internal/ui/branchselect.go` (reuse from 3.1)
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`

**Pattern:**
- Menu item: "Merge another branch"
- Show list of branches (excluding current)
- User selects source branch
- Confirm merge: "source → current"
- Execute: `git merge <source>`
- Handle conflicts

**Changes:**
1. Add menu item in `menuNormal()`
2. Reuse branch selection UI
3. Add confirmation dialog
4. Execute merge (async)
5. If conflicts → `Operation = Conflicted`

**Test:**
```bash
# Create test scenario:
git checkout -b feature
echo "feature work" > feature.txt
git add -A && git commit -m "feature"
git checkout main
./tit_x64
# Select "Merge another branch"
# Select "feature"
# Verify merge succeeds
git log  # Should show merge commit
```

**Expected result:**
- Merge completes
- Branches intact
- If conflicts, enters conflict mode

---

## Phase 4: Dirty Operation Protocol

**Port stash-based preservation from old TIT.**

### 4.1 Dirty Pull Implementation

**Files:**
- `internal/app/handlers.go`
- `internal/git/dirty.go` (new file)

**Pattern from old TIT:**
1. Detect WorkingTree = Modified before pull
2. Show dialog: "Save changes and pull" vs "Discard and pull"
3. Stash → Pull → Stash apply
4. Handle conflicts at each step
5. Abort restores exact state

**Changes:**
1. Create `internal/git/dirty.go`:
   - `BeginDirtyOperation()` — stash + save state
   - `FinalizeDirtyOperation()` — apply stash + cleanup
   - `AbortDirtyOperation()` — restore original state
2. Modify pull handler to check WorkingTree
3. Show dirty operation dialog if Modified
4. Track operation in `.git/TIT_DIRTY_OP`

**Test:**
```bash
# Make local changes
echo "wip" >> test.txt
# Simulate remote changes (use second clone)
./tit_x64
# Select "Pull"
# Should show dirty operation dialog
# Choose "Save and pull"
# Verify stash applied after pull
```

**Expected result:**
- Stash created before pull
- Pull completes
- Stash reapplied
- Working tree shows WIP changes + pulled changes
- Can abort at any step

---

### 4.2 Dirty Merge Implementation

**Files:**
- `internal/app/handlers.go`

**Pattern:**
- Same as dirty pull
- Apply to merge operation

**Changes:**
1. Reuse `dirty.go` functions from 4.1
2. Modify merge handler to check WorkingTree
3. Show dirty operation dialog if Modified

**Test:**
```bash
# Make local changes
echo "wip" >> test.txt
./tit_x64
# Select "Merge another branch"
# Should show dirty operation dialog
# Verify same workflow as dirty pull
```

**Expected result:**
- Same safety as dirty pull
- Works for any merge source

---

## Phase 5: History Browsers

**Port 2-pane commit history and 3-pane file history from old TIT.**

### 5.1 Commit History Browser (2-Pane)

**Files:**
- `internal/app/modes.go` (add `ModeCommitHistory`)
- `internal/ui/history.go` (new file)
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`
- `internal/app/keyboard.go`

**Pattern from old TIT:**
- Two panes: Commit list (left) + Details (right)
- Navigate with ↑↓
- Press Enter to time travel (Phase 6)
- ESC returns to menu

**Changes:**
1. Add `ModeCommitHistory` mode
2. Create `ui.RenderCommitHistory()`:
   - Left pane: `git log --oneline`
   - Right pane: `git show <selected-commit>`
3. Add navigation handlers (↑↓ to scroll commits)
4. Add Enter handler (placeholder for time travel)
5. Add ESC handler (return to menu)

**Test:**
```bash
./tit_x64
# Select "Browse commit history"
# Should show 2-pane layout
# Navigate with ↑↓, verify details update
# ESC returns to menu
```

**Expected result:**
- Commit list scrolls correctly
- Details pane shows selected commit info
- Navigation smooth and responsive

---

### 5.2 File History Browser (3-Pane)

**Files:**
- `internal/app/modes.go` (add `ModeFileHistory`)
- `internal/ui/filehistory.go` (new file)
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`
- `internal/app/keyboard.go`

**Pattern from old TIT:**
- Three panes: Commits (top-left) + Files (top-right) + Diff (bottom)
- Tab to cycle focus
- ↑↓ to navigate within pane
- State-dependent diff (Modified vs Clean)

**Changes:**
1. Add `ModeFileHistory` mode
2. Create `ui.RenderFileHistory()`:
   - Commits: `git log --oneline`
   - Files: `git show --name-only <commit>`
   - Diff: `git show <commit> -- <file>` (if Clean) or `git diff <commit> -- <file>` (if Modified)
3. Add Tab handler (cycle panes)
4. Add ↑↓ handlers (scroll within pane)
5. Add ESC handler (return to menu)

**Test:**
```bash
./tit_x64
# Select "Browse file history"
# Should show 3-pane layout
# Tab to cycle focus
# Verify diff updates correctly
```

**Expected result:**
- Three panes render correctly
- Tab cycles focus
- Diff shows correct comparison based on WorkingTree state

---

## Phase 6: Time Travel

**Implement read-only time travel with merge-back option.**

### 6.1 Enter Time Travel from History

**Files:**
- `internal/app/handlers.go` (commit history)
- `internal/git/timetravel.go` (new file)

**Pattern:**
- From commit history, press Enter
- Show confirmation dialog
- Execute: `git checkout <commit-hash>`
- Set `Operation = TimeTraveling`
- Save original branch to `.git/TIT_TIME_TRAVEL`

**Changes:**
1. Create `internal/git/timetravel.go`:
   - `EnterTimeTravel(commit)` — checkout + save state
   - `ExitTimeTravel()` — return to original branch
2. Add Enter handler in commit history mode
3. Show confirmation dialog
4. Execute checkout (async)
5. Update state to TimeTraveling

**Test:**
```bash
./tit_x64
# Browse commit history
# Select old commit, press Enter
# Confirm time travel
# Verify detached HEAD state
git status  # Should show "HEAD detached at <commit>"
```

**Expected result:**
- Enters detached HEAD
- Operation = TimeTraveling
- Menu shows time travel options only

---

### 6.2 Time Travel Menu

**Files:**
- `internal/app/menu.go`
- `internal/app/dispatchers.go`

**Pattern:**
- Show ONLY time travel options when `Operation = TimeTraveling`
- "Jump to different commit" (back to history browser)
- "View diff vs original branch"
- "Merge changes back to [branch]"
- "Return to [branch]" (discard)

**Changes:**
1. Add `menuTimeTraveling()` function
2. Add dispatchers for each option
3. Implement "Jump" (reopen history browser in time travel mode)
4. Implement "View diff" (show diff vs original branch)

**Test:**
```bash
# While in time travel mode
./tit_x64
# Should show ONLY time travel menu
# Verify each option works
```

**Expected result:**
- Menu restricted to time travel options
- Can navigate to different commits
- Can view diff

---

### 6.3 Merge Time Travel Changes Back

**Files:**
- `internal/app/handlers.go`
- `internal/git/timetravel.go`

**Pattern (using dirty op protocol):**
1. If Modified: stash changes
2. Save detached HEAD commit
3. Checkout original branch
4. Merge detached commit
5. If stash exists: apply stash
6. Cleanup

**Changes:**
1. Add `MergeTimeTravelBack()` to `timetravel.go`:
   - Reuse dirty operation pattern
   - Merge detached commit into original branch
2. Add handler: `handleTimeTravelMergeBack()`
3. Show confirmation dialog
4. Handle conflicts at each step

**Test:**
```bash
# While in time travel mode
# Make changes: echo "test" > newfile.txt
./tit_x64
# Select "Merge changes back to main"
# Confirm merge
# Verify back on main with changes merged
git log  # Should show merge
```

**Expected result:**
- Returns to original branch
- Changes merged successfully
- Operation = Normal
- Can abort at any step

---

### 6.4 Return to Branch (Discard)

**Files:**
- `internal/app/handlers.go`
- `internal/git/timetravel.go`

**Pattern:**
- Simple checkout to original branch
- If Modified: warn about discarding changes

**Changes:**
1. Add `ReturnFromTimeTravel()` to `timetravel.go`
2. Check WorkingTree before checkout
3. Show warning dialog if Modified
4. Execute: `git checkout <original-branch>`
5. Cleanup `.git/TIT_TIME_TRAVEL`

**Test:**
```bash
# While in time travel mode
./tit_x64
# Select "Return to main"
# Verify returned to main
# Time travel state cleared
```

**Expected result:**
- Returns to original branch
- Working tree discarded if Modified
- Operation = Normal

---

## Phase 7: Conflict Resolution

**Integrate conflict detection and resolution from old TIT.**

### 7.1 Conflict Detection

**Files:**
- `internal/git/state.go`
- `internal/git/conflicts.go` (new file)

**Pattern from old TIT:**
- After any operation, check for conflicts
- Set `Operation = Conflicted` if found
- Show conflict resolution menu

**Changes:**
1. Create `conflicts.go`:
   - `DetectConflicts()` — check for conflict markers
   - `ListConflictedFiles()` — get list of conflicted files
   - `IsConflictResolved()` — verify all conflicts resolved
2. Modify `DetectState()` to check for conflicts
3. Update operation handlers to set Conflicted state

**Test:**
```bash
# Create merge conflict:
# 1. Create branch, modify file, commit
# 2. Checkout main, modify same file differently, commit
# 3. Try to merge branch
./tit_x64
# Select "Merge another branch"
# Should detect conflicts and show conflict menu
```

**Expected result:**
- State = Conflicted
- Menu shows conflict resolution options only

---

### 7.2 Conflict Resolution Menu

**Files:**
- `internal/app/menu.go`
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`

**Pattern from old TIT:**
- Show conflicted files list
- "Resolve conflicts externally" (opens editor)
- "Continue operation" (after resolving)
- "Abort operation" (rollback)

**Changes:**
1. Add `menuConflicted()` function
2. Add "View conflicted files" option
3. Add "Resolve externally" (opens $EDITOR)
4. Add "Continue" handler (checks if resolved)
5. Add "Abort" handler (git merge --abort or git rebase --abort)

**Test:**
```bash
# While in conflicted state
./tit_x64
# Should show conflict menu only
# Select "View conflicted files"
# Resolve manually
# Select "Continue"
# Verify operation completes
```

**Expected result:**
- Can view conflicted files
- Can open editor to resolve
- Continue succeeds only if all conflicts resolved
- Abort restores pre-operation state

---

## Phase 8: Polish and Edge Cases

**Handle edge cases and improve UX.**

### 8.1 Add Remote Flow

**Files:**
- `internal/app/dispatchers.go`
- `internal/app/handlers.go`

**Pattern:**
- Show when Timeline = NoRemote
- Prompt for remote URL
- Execute: `git remote add origin <url>`
- Optionally fetch after adding

**Changes:**
1. Add menu item when `Remote == git.NoRemote`
2. Add URL input mode (reuse existing text input)
3. Validate URL format
4. Execute remote add + fetch

**Test:**
```bash
# In repo without remote
git remote remove origin
./tit_x64
# Should show "Add remote"
# Enter URL, verify remote added
git remote -v  # Should show origin
```

**Expected result:**
- Remote added successfully
- Timeline updates after fetch

---

### 8.2 Force Operations

**Files:**
- `internal/app/menu.go`
- `internal/app/handlers.go`

**Pattern:**
- Show destructive options with ⚠️ warning
- "Force push" when Ahead
- "Replace local with remote" when Behind or Diverged
- Require confirmation

**Changes:**
1. Add force push option when Ahead
2. Add hard reset option when Behind
3. Show confirmation dialog with clear warning
4. Execute with confirmation

**Test:**
```bash
# Test force push:
# 1. Make local commit
# 2. Amend commit (creates divergence)
./tit_x64
# Should show "Force push" with warning
# Confirm and verify push
```

**Expected result:**
- Clear warning shown
- Requires explicit confirmation
- Executes correctly

---

### 8.3 Commit and Push (Combined)

**Files:**
- `internal/app/menu.go`
- `internal/app/handlers.go`

**Pattern:**
- Show when Modified + HasRemote
- Combines commit + push in one operation
- Useful for quick workflow

**Changes:**
1. Add menu item when `Modified && HasRemote`
2. Prompt for commit message
3. Execute: commit → push (sequential)
4. Handle failures at each step

**Test:**
```bash
# Make changes
echo "test" > test.txt
./tit_x64
# Should show "Commit and push"
# Execute, verify both operations succeed
```

**Expected result:**
- Commit succeeds
- Push succeeds immediately after
- Timeline = InSync

---

### 8.4 Detached HEAD Error Handling

**Files:**
- `internal/git/state.go`
- `internal/app/menu.go`

**Pattern:**
- Detect detached HEAD not from TIT
- Show error message
- Suggest fix: `git checkout <branch>`

**Changes:**
1. In `DetectState()`, check for detached HEAD
2. If `.git/TIT_TIME_TRAVEL` doesn't exist → Error state
3. Show error menu with instructions

**Test:**
```bash
# Create detached HEAD outside TIT
git checkout HEAD~1
./tit_x64
# Should show error message
# Suggest: "git checkout main"
```

**Expected result:**
- Clear error message
- Helpful recovery instructions
- TIT exits gracefully

---

## Testing Strategy

**After each phase:**

1. **Build:**
   ```bash
   ./build.sh
   ```

2. **Manual test:**
   - Create test scenario
   - Execute operation
   - Verify success path
   - Test abort/ESC at each step
   - Verify state updates correctly

3. **Edge cases:**
   - Test with empty repo
   - Test with detached HEAD
   - Test with conflicts
   - Test with no remote

4. **Regression test:**
   - Verify previous phases still work
   - Check that new changes don't break existing features

---

## Implementation Order Summary

```
Phase 1: Simplify state model (remove BranchContext) ✓
  ├─ 1.1: Remove BranchContext from types
  ├─ 1.2: Remove config file tracking
  └─ 1.3: Update menu generation

Phase 2: Core git operations ✓
  ├─ 2.1: Commit
  ├─ 2.2: Push
  └─ 2.3: Pull (merge)

Phase 3: Branch operations ✓
  ├─ 3.1: List and switch branches
  ├─ 3.2: Create new branch
  └─ 3.3: Merge branch into current

Phase 4: Dirty operation protocol ✓
  ├─ 4.1: Dirty pull
  └─ 4.2: Dirty merge

Phase 5: History browsers ✓
  ├─ 5.1: Commit history (2-pane)
  └─ 5.2: File history (3-pane)

Phase 6: Time travel ✓
  ├─ 6.1: Enter time travel
  ├─ 6.2: Time travel menu
  ├─ 6.3: Merge changes back
  └─ 6.4: Return to branch (discard)

Phase 7: Conflict resolution ✓
  ├─ 7.1: Conflict detection
  └─ 7.2: Conflict resolution menu

Phase 8: Polish and edge cases ✓
  ├─ 8.1: Add remote flow
  ├─ 8.2: Force operations
  ├─ 8.3: Commit and push (combined)
  └─ 8.4: Detached HEAD error handling
```

---

## File Reference (Old TIT → New TIT)

**Old TIT files to reference during porting:**

| Feature | Old TIT File | New TIT Target |
|---------|-------------|----------------|
| State detection | `git/state.go` | `internal/git/state.go` |
| Dirty ops | `dirty_ops.go` | `internal/git/dirty.go` |
| Commit history | `history.go` | `internal/ui/history.go` |
| File history | `file_history.go` | `internal/ui/filehistory.go` |
| Conflict resolution | `conflicts.go` | `internal/git/conflicts.go` |
| Menu generation | `menu.go` | `internal/app/menu.go` |
| Keyboard handlers | `input.go` | `internal/app/keyboard.go` |

---

**End of Implementation Plan**
