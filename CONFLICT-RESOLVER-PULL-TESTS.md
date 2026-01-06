# Pull Merge Conflict Resolution — Manual Test Cases

**Purpose:** Document exact manual git setup for each pull conflict scenario.

**Test Repo:** `/Users/jreng/Documents/Poems/inf/t`

**Philosophy:** Manual setup is more reliable than scripts. Each test documents exact git commands.

---

## Test Case 1: Clean Pull with Merge Conflicts (DIVERGED)

### Git State Setup

**Goal:** Both local and remote have commits from a common base → DIVERGED

```bash
cd /Users/jreng/Documents/Poems/inf/t

# 1. Clean any existing state
git reset --hard HEAD
git clean -fd
git merge --abort 2>/dev/null
git stash clear

# 2. Create baseline commit (if not exists)
echo "baseline content" > baseline.txt
git add baseline.txt
git commit -m "Baseline commit" 2>/dev/null || echo "Using existing HEAD"
BASELINE=$(git rev-parse HEAD)

# 3. Create LOCAL commit (modify conflict.txt)
echo "Line 1: baseline" > conflict.txt
echo "Line 2: LOCAL CHANGE" >> conflict.txt
echo "Line 3: baseline" >> conflict.txt
git add conflict.txt
git commit -m "Local: modified line 2"

# 4. Create REMOTE commit (modify same file, same line)
git checkout -b temp-remote
git reset --hard "$BASELINE"
echo "Line 1: baseline" > conflict.txt
echo "Line 2: REMOTE CHANGE" >> conflict.txt
echo "Line 3: baseline" >> conflict.txt
git add conflict.txt
git commit -m "Remote: modified line 2"
git push -f origin temp-remote:main

# 5. Reset local to have LOCAL commit
git checkout main
git reset --hard "$BASELINE"
echo "Line 1: baseline" > conflict.txt
echo "Line 2: LOCAL CHANGE" >> conflict.txt
echo "Line 3: baseline" >> conflict.txt
git add conflict.txt
git commit -m "Local: modified line 2"
git branch -D temp-remote

# 6. Fetch remote state (CRITICAL)
git fetch origin
git branch --set-upstream-to=origin/main main

# 7. Verify state
git status -sb
# Expected: ## main...origin/main [ahead 1, behind 1]
```

### Expected Git State

```
Branch: main
Tracking: origin/main
Ahead: 1 commit (local)
Behind: 1 commit (remote)
Timeline: DIVERGED
WorkingTree: Clean
Operation: Normal
```

### Expected TIT Display

**Header:**
```
Branch: main
Working Tree: Clean
Timeline: Diverged (1 ahead, 1 behind)
```

**Menu Options:**
```
✓ Pull (merge)                [p]
✓ Replace local (hard reset)  [f]
✓ Commit history              [h]
✓ File(s) history             [i]
```

### Expected Behavior — Success Path

```
1. User selects: Pull (merge) [p]
2. No confirmation (clean tree = safe)
3. Console shows:
   - git pull --no-rebase
   - Auto-merging conflict.txt
   - CONFLICT (content): Merge conflict in conflict.txt
   - Command exited with code 1
4. TIT transitions to Conflict Resolver:
   - File list: conflict.txt
   - Columns: BASE | LOCAL (yours) | REMOTE (theirs)
   - User presses TAB to focus columns
   - User presses SPACE to mark choice
   - User presses ENTER to finalize
5. Console shows:
   - git add -A
   - git commit -m "Merge resolved conflicts"
   - Merge completed successfully
   - Press ESC to return to menu
6. User presses ESC
7. Menu shows:
   - State: Clean, Ahead (2 commits)
   - Options: Push, Force Push, History
```

### Expected Behavior — Abort Path

```
1-4. Same as success path (up to conflict resolver)
5. User presses ESC in conflict resolver
6. Console shows:
   - Aborting merge...
   - git merge --abort
   - Successfully aborted by user
   - Press ESC to return to menu
7. User presses ESC
8. Menu shows:
   - State: Clean, Diverged (1 ahead, 1 behind)
   - Back to original state
```

### Compliance with FLOW-TESTING-CHECKLIST.md

- **#1 State Detection:** ✅ git status shows "ahead 1, behind 1"
- **#2 Correct State Display:** ✅ TIT shows "Diverged"
- **#3 Correct Menu:** ✅ Pull (merge) option visible
- **#4 Confirmation Guard:** ✅ Clean tree = no confirmation
- **#5 Operation Guarantee:** ✅ Conflict resolver handles merge conflicts
- **#6 Fail Prevention:** ✅ No pre-check needed (conflicts are expected)
- **#7 Abort Safety:** ✅ git merge --abort restores original state

---

## Test Case 2: Dirty Pull with Merge Conflicts (DIVERGED + Modified)

### Git State Setup

**Goal:** DIVERGED + uncommitted changes in separate file

```bash
cd /Users/jreng/Documents/Poems/inf/t

# 1. Start from Test Case 1 setup (steps 1-6)
# ... (repeat Test Case 1 steps 1-6)

# 7. Add uncommitted changes in SEPARATE file
echo "work in progress" > wip.txt

# 8. Verify state
git status -sb
# Expected: ## main...origin/main [ahead 1, behind 1]
#           ?? wip.txt
```

### Expected Git State

```
Branch: main
Tracking: origin/main
Ahead: 1 commit (local)
Behind: 1 commit (remote)
Timeline: DIVERGED
WorkingTree: Modified (wip.txt untracked)
Operation: Normal
```

### Expected TIT Display

**Header:**
```
Branch: main
Working Tree: Modified
Timeline: Diverged (1 ahead, 1 behind)
```

**Menu Options:**
```
✓ Pull (save changes)         [d]  ← NEW (dirty pull option)
─────────────────────────────
✓ Pull (merge)                [p]
✓ Replace local (hard reset)  [f]
✓ Commit                      [c]
✓ Commit history              [h]
```

### Expected Behavior — Success Path

```
1. User selects: Pull (save changes) [d]
2. Confirmation dialog:
   Title: "Save your changes?"
   Buttons: [Save changes] [Discard changes]
3. User selects: Save changes
4. Console shows Phase 1:
   - Saving your changes (creating stash)...
   - Saved as stash@{0}
   - Snapshot saved. Starting merge pull...
5. Console shows Phase 2 (merge):
   - git pull --no-rebase
   - CONFLICT (content): Merge conflict in conflict.txt
   - Command exited with code 1
6. TIT transitions to Conflict Resolver:
   - Operation: dirty_pull_changeset_apply
   - File list: conflict.txt
   - Columns: BASE | LOCAL (yours) | REMOTE (theirs)
7. User marks files and presses ENTER
8. Console shows Phase 3 (reapply stash):
   - Reapplying your changes...
   - Applied stash successfully
9. Console shows Phase 4 (finalize):
   - Finalizing dirty pull operation...
   - Dirty pull completed successfully
   - Press ESC to return to menu
10. User presses ESC
11. Menu shows:
    - State: Modified (wip.txt), Ahead (2 commits)
```

### Expected Behavior — Abort Path

```
1-6. Same as success (up to conflict resolver)
7. User presses ESC in conflict resolver
8. Console shows:
   - Aborting dirty pull and restoring original state...
   - git merge --abort
   - Checking out original branch: main
   - Resetting to original HEAD
   - Reapplying stash (your changes)
   - Original state restored
   - Press ESC to return to menu
9. User presses ESC
10. Menu shows:
    - State: Modified (wip.txt), Diverged (1 ahead, 1 behind)
    - Exact original state restored
```

### Compliance with FLOW-TESTING-CHECKLIST.md

- **#1 State Detection:** ✅ git status shows modified + diverged
- **#2 Correct State Display:** ✅ TIT shows "Modified, Diverged"
- **#3 Correct Menu:** ✅ "Pull (save changes)" visible for dirty tree
- **#4 Confirmation Guard:** ✅ Dialog asks "Save your changes?"
- **#5 Operation Guarantee:** ✅ Multi-phase with conflict handling
- **#6 Fail Prevention:** ✅ Pre-check: no ongoing operations
- **#7 Abort Safety:** ✅ Full restore including stash reapply

---

## Test Case 3: Dirty Pull with Stash Apply Conflicts (BEHIND + Modified)

### Git State Setup

**Goal:** BEHIND (no local commits) + uncommitted changes that conflict with incoming remote

```bash
cd /Users/jreng/Documents/Poems/inf/t

# 1. Clean state
git reset --hard HEAD
git clean -fd
git stash clear

# 2. Get current HEAD as baseline
BASELINE=$(git rev-parse HEAD)

# 3. Create REMOTE commit (modify conflict.txt)
git checkout -b temp-remote
echo "Line 1: baseline" > conflict.txt
echo "Line 2: REMOTE CHANGE" >> conflict.txt
echo "Line 3: baseline" >> conflict.txt
git add conflict.txt
git commit -m "Remote: modified line 2"
git push -f origin temp-remote:main

# 4. Reset local to baseline (BEHIND, no local commits)
git checkout main
git reset --hard "$BASELINE"
git branch -D temp-remote

# 5. Fetch remote
git fetch origin
git branch --set-upstream-to=origin/main main

# 6. Create uncommitted changes to SAME file
echo "Line 1: baseline" > conflict.txt
echo "Line 2: LOCAL WIP CHANGE" >> conflict.txt
echo "Line 3: baseline" >> conflict.txt

# 7. Verify state
git status -sb
# Expected: ## main...origin/main [behind 1]
#            M conflict.txt
```

### Expected Git State

```
Branch: main
Tracking: origin/main
Ahead: 0 commits
Behind: 1 commit (remote)
Timeline: BEHIND
WorkingTree: Modified (conflict.txt modified, unstaged)
Operation: Normal
```

### Expected TIT Display

**Header:**
```
Branch: main
Working Tree: Modified
Timeline: Behind (1 commit)
```

**Menu Options:**
```
✓ Pull (save changes)         [d]
─────────────────────────────
✓ Pull (merge)                [p]
✓ Replace local (hard reset)  [f]
✓ Commit                      [c]
```

### Expected Behavior — Success Path

```
1. User selects: Pull (save changes) [d]
2. Confirmation dialog: "Save your changes?"
3. User selects: Save changes
4. Console shows Phase 1 (stash):
   - Saving your changes...
   - Saved as stash@{0}
5. Console shows Phase 2 (pull):
   - git pull --no-rebase
   - Updating abc123..def456
   - Fast-forward (NO merge conflicts here)
   - Command completed successfully
6. Console shows Phase 3 (stash apply):
   - Reapplying your changes...
   - CONFLICT (content): Merge conflict in conflict.txt
7. TIT transitions to Conflict Resolver:
   - Operation: dirty_pull_snapshot_reapply
   - File list: conflict.txt
   - Columns: BASE | LOCAL (yours) | REMOTE (theirs)
8. User marks files and presses ENTER
9. Console shows Phase 4 (finalize):
   - Finalizing dirty pull operation...
   - Dirty pull completed successfully
10. Menu shows:
    - State: Modified (conflict.txt with chosen version)
```

### Expected Behavior — Abort Path

```
1-7. Same as success (up to conflict resolver)
8. User presses ESC in conflict resolver
9. Console shows:
   - Aborting dirty pull...
   - git reset --hard (to undo stash apply)
   - Checking out original branch
   - Resetting to original HEAD (before pull)
   - Reapplying stash (your WIP changes)
   - Original state restored
10. Menu shows:
    - State: Modified (conflict.txt with original WIP), Behind (1 commit)
```

### Compliance with FLOW-TESTING-CHECKLIST.md

- **#1 State Detection:** ✅ git status shows behind + modified
- **#2 Correct State Display:** ✅ TIT shows "Modified, Behind"
- **#3 Correct Menu:** ✅ Dirty pull option visible
- **#4 Confirmation Guard:** ✅ Dialog appears
- **#5 Operation Guarantee:** ✅ Pull succeeds, conflicts during stash apply
- **#6 Fail Prevention:** ✅ Pre-checks pass
- **#7 Abort Safety:** ✅ Full restore to pre-pull state

---

## Test Case 4: Dirty Pull Clean (No Conflicts) — Happy Path

### Git State Setup

**Goal:** BEHIND + uncommitted changes in separate file (no conflicts anywhere)

```bash
cd /Users/jreng/Documents/Poems/inf/t

# 1. Clean state
git reset --hard HEAD
git clean -fd
git stash clear

# 2. Get baseline
BASELINE=$(git rev-parse HEAD)

# 3. Create REMOTE commit (new file)
git checkout -b temp-remote
echo "remote file content" > remote_file.txt
git add remote_file.txt
git commit -m "Remote: added remote_file.txt"
git push -f origin temp-remote:main

# 4. Reset local to baseline
git checkout main
git reset --hard "$BASELINE"
git branch -D temp-remote

# 5. Fetch remote
git fetch origin
git branch --set-upstream-to=origin/main main

# 6. Create uncommitted changes in SEPARATE file
echo "local wip" > local_wip.txt

# 7. Verify state
git status -sb
# Expected: ## main...origin/main [behind 1]
#           ?? local_wip.txt
```

### Expected Git State

```
Branch: main
Timeline: BEHIND (1 commit)
WorkingTree: Modified (local_wip.txt untracked)
Operation: Normal
```

### Expected TIT Display

**Header:**
```
Branch: main
Working Tree: Modified
Timeline: Behind (1 commit)
```

**Menu Options:**
```
✓ Pull (save changes)         [d]
✓ Pull (merge)                [p]
✓ Replace local               [f]
✓ Commit                      [c]
```

### Expected Behavior — Success Path (Auto-Complete)

```
1. User selects: Pull (save changes) [d]
2. Confirmation: "Save your changes?"
3. User selects: Save changes
4. Console shows Phase 1:
   - Saving your changes...
   - Saved as stash@{0}
5. Console shows Phase 2:
   - git pull --no-rebase
   - Fast-forward (no conflicts)
   - Command completed successfully
6. Console shows Phase 3:
   - Reapplying your changes...
   - Applied stash successfully (no conflicts)
7. Console shows Phase 4:
   - Finalizing dirty pull operation...
   - Dirty pull completed successfully
   - Press ESC to return to menu
8. User presses ESC
9. Menu shows:
   - State: Modified (local_wip.txt), In Sync
   - NO conflict resolver shown (everything succeeded)
```

### Compliance with FLOW-TESTING-CHECKLIST.md

- **#1 State Detection:** ✅ Behind + Modified
- **#2 Correct State Display:** ✅ Accurate
- **#3 Correct Menu:** ✅ Dirty pull visible
- **#4 Confirmation Guard:** ✅ Dialog appears
- **#5 Operation Guarantee:** ✅ All phases succeed automatically
- **#6 Fail Prevention:** ✅ Pre-checks pass
- **#7 Abort Safety:** ✅ Can abort at any phase before completion

---

## Summary of Test Cases

| # | Scenario | Timeline | WorkingTree | Conflicts At | Resolver? |
|---|----------|----------|-------------|--------------|-----------|
| 1 | Clean Pull | Diverged | Clean | Merge | YES |
| 2 | Dirty Pull (merge conflicts) | Diverged | Modified | Merge | YES |
| 3 | Dirty Pull (stash conflicts) | Behind | Modified | Stash Apply | YES |
| 4 | Dirty Pull (clean) | Behind | Modified | None | NO |

---

## Testing Workflow

For each test case:

1. **Setup:** Run manual git commands (copy-paste from above)
2. **Verify Git State:** Run `git status -sb` → check output matches expected
3. **Open TIT:** `./tit_x64`
4. **Verify TIT Display:** Check header and menu match expected
5. **Test Success Path:** Follow expected behavior steps
6. **Reset:** Run setup again
7. **Test Abort Path:** Follow abort behavior steps

---

**End of Test Cases**
