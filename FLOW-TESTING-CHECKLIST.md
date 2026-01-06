# Dirty Pull Testing Framework — FLOW-TESTING-CHECKLIST

**Purpose:** Verify each dirty pull scenario against specification requirements before/during/after operation.

**Test Repo:** `/Users/jreng/Documents/Poems/inf/t`

---

## 🎯 Seven-Point Verification Matrix

For **each test case**, verify:

| # | Checkpoint | Question | Required? | Pass/Fail |
|---|------------|----------|-----------|-----------|
| 1 | **State Detection** | Is git state properly detected by TIT? | ✅ YES | [ ] |
| 2 | **Correct State Display** | Does TIT show correct state per SPEC.md (WorkingTree + Timeline + Operation)? | ✅ YES | [ ] |
| 3 | **Correct Menu** | Does TIT generate correct menu per SPEC.md for this state? | ✅ YES | [ ] |
| 4 | **Confirmation Guard** | Are destructive ops guarded with confirmation dialog? | ✅ YES | [ ] |
| 5 | **Operation Guarantee** | Can we guarantee operation ALWAYS succeeds (pre-checks automated)? | ✅ YES | [ ] |
| 6 | **Fail Prevention** | Can we prevent failed ops by not offering bad options to user at all? | ✅ YES | [ ] |
| 7 | **Abort Safety** | Can we restore exact pre-operation state if user aborts mid-operation? | ✅ YES | [ ] |

---

## 📋 Test Cases from titest.sh

```
Scenario 0: Reset to fresh state (baseline)
Scenario 1: Pull with conflicts (clean tree)
Scenario 2: Dirty pull - merge with conflicts
Scenario 3: Dirty pull - rebase with conflicts
Scenario 4: Dirty pull - stash apply conflicts after pull
Scenario 5: Dirty pull - clean pull (no conflicts)
```

---

## 🧪 Scenario 0: Reset to Fresh State

**Purpose:** Establish baseline, clean any leftover state from previous tests

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
./titest.sh
# Select: 0
```

**Expected Git State:**
```
Branch: main (tracking origin/main)
WorkingTree: Clean
Timeline: InSync (or Behind if remote has new commits)
Operation: Normal
```

### Checklist for Scenario 0:

- [ ] **#1 State Detection** 
  - [ ] `git status` shows clean
  - [ ] `git rev-parse --abbrev-ref HEAD` = `main`
  - [ ] `git branch -vv` shows `[origin/main]` tracking
  
- [ ] **#2 Correct State Display**
  - [ ] TIT header shows: `Branch: main | Clean | In sync` (or Behind)
  - [ ] No error messages in footer
  
- [ ] **#3 Correct Menu**
  - [ ] Should show Normal menu (commit, push/pull, merge, history, branches)
  - [ ] Should NOT show dirty pull option (clean tree)
  
- [ ] **#4 Confirmation Guard**
  - [ ] N/A (no destructive ops in this scenario)
  
- [ ] **#5 Operation Guarantee**
  - [ ] N/A (no operations)
  
- [ ] **#6 Fail Prevention**
  - [ ] N/A (no operations)
  
- [ ] **#7 Abort Safety**
  - [ ] N/A (no operations)

---

## 🧪 Scenario 1: Pull with Conflicts (Clean Tree)

**Purpose:** Verify conflict detection + resolution UI works with clean working tree

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
./titest.sh
# Select: 1
# Opens TIT
# Select: Pull from remote
```

**Expected Git State After Setup:**
```
Branch: main
WorkingTree: Clean
Timeline: Behind (local commit, remote ahead)
Operation: Normal
```

**Expected TIT Behavior:**
```
Menu shows:
  ✓ Pull (fetch + merge) [shortcut: p]

User selects:
  ✓ No confirmation needed (clean tree = safe)
  ✓ Merge starts
  ✓ Conflict detected
  ✓ Operation → Conflicted
  ✓ ModeConflictResolve shown
```

### Checklist for Scenario 1:

- [ ] **#1 State Detection**
  - [ ] After setup: `git status` shows "Behind"
  - [ ] `git log --oneline origin/main..HEAD` shows 0 commits ahead
  - [ ] `git log --oneline HEAD..origin/main` shows 1+ commits (behind)
  - [ ] `git diff` shows empty (working tree clean)
  
- [ ] **#2 Correct State Display**
  - [ ] TIT header: `Branch: main | Clean | Behind`
  - [ ] No git errors in footer
  
- [ ] **#3 Correct Menu**
  - [ ] Pull option visible: "Pull (fetch + merge)" [p]
  - [ ] Replace local visible: "Replace local (hard reset)" [f]
  - [ ] Dirty pull NOT visible (clean tree)
  
- [ ] **#4 Confirmation Guard**
  - [ ] Clean pull: NO confirmation needed (safe)
  - [ ] Replace local: YES confirmation (destructive)
  - [ ] User selects "Pull (merge)"
  - [ ] Should proceed directly to merge (no dialog)
  
- [ ] **#5 Operation Guarantee**
  - [ ] Merge will succeed (only conflict on merge, we handle it)
  - [ ] Or: Pre-check if conflict inevitable and warn user
  - [ ] No case where user clicks "Pull" and operation silently fails
  
- [ ] **#6 Fail Prevention**
  - [ ] Can we check for conflicts before offering pull?
  - [ ] Or: Accept that conflicts are expected, show resolver UI
  - [ ] User is NEVER left in unexpected Operation state
  
- [ ] **#7 Abort Safety**
  - [ ] Conflict resolver shows: ESC to abort
  - [ ] ESC aborts merge: `git merge --abort`
  - [ ] State returns to: Ahead (local commit still there)
  - [ ] No commits lost

---

## 🧪 Scenario 2: Dirty Pull - Merge with Conflicts

**Purpose:** Complete dirty pull flow with conflicts during merge phase

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
./titest.sh
# Select: 2
# Opens TIT
# Select: Pull (save changes)
```

**Expected Git State After Setup:**
```
Branch: main
WorkingTree: Modified (wip.txt)
Timeline: Behind
Operation: Normal
```

**Expected TIT Behavior:**
```
Menu shows:
  ✓ Pull (save changes) [shortcut: d] ← NEW dirty pull option
  ✓ Pull (fetch + merge) [p] ← regular pull
  ✓ Replace local [f]

User selects "Pull (save changes)":
  ✓ Confirmation dialog: "Save your changes? [Yes] [No] [Cancel]"
  ✓ Selects "Save changes"
  ✓ Phase 1: Stash WIP → Snapshot saved
  ✓ Phase 2a: Merge → Conflicts detected
  ✓ Operation → Conflicted
  ✓ ConflictResolver shown (dirty_pull_changeset_apply)
  ✓ User marks files + ENTER
  ✓ Phase 3: Apply snapshot → Success
  ✓ Phase 4: Finalize → Return to menu
  ✓ Result: All changes preserved, synced with remote
```

### Checklist for Scenario 2:

- [ ] **#1 State Detection**
  - [ ] After setup: `git status` shows "modified: wip.txt" (unstaged)
  - [ ] `git log --oneline HEAD..origin/main` shows 1+ commits (behind)
  - [ ] `git diff HEAD` shows wip.txt added
  - [ ] No existing stashes: `git stash list` empty
  
- [ ] **#2 Correct State Display**
  - [ ] TIT header: `Branch: main | Modified | Behind`
  - [ ] No git errors
  
- [ ] **#3 Correct Menu**
  - [ ] "Pull (save changes)" visible [d] ← NEW
  - [ ] "Pull (fetch + merge)" visible [p]
  - [ ] "Replace local" visible [f]
  - [ ] Separator between dirty pull and regular pull
  
- [ ] **#4 Confirmation Guard**
  - [ ] User selects "Pull (save changes)"
  - [ ] Dialog appears: "Save your changes?"
  - [ ] 3 buttons: "Save changes" (left, default), "Discard changes", "Cancel"
  - [ ] User clicks "Cancel" → returns to menu (no operation)
  - [ ] User clicks "Save changes" → proceeds to stash
  - [ ] User clicks "Discard changes" → proceeds without stash
  
- [ ] **#5 Operation Guarantee**
  - [ ] Pre-check: Is stash possible? (always yes, can stash anything)
  - [ ] Pre-check: Is merge possible? (check for ongoing operations)
  - [ ] If any pre-check fails: Don't show dirty pull option
  - [ ] Operation sequence is atomic: all-or-nothing
  
- [ ] **#6 Fail Prevention**
  - [ ] Before offering "Pull (save changes)", verify:
    - [ ] No ongoing merge/rebase/cherry-pick
    - [ ] Can create stash (always possible)
    - [ ] Remote configured (check timeline state)
  - [ ] Don't offer option if ANY precondition fails
  
- [ ] **#7 Abort Safety (3 abort points)**
  
  **Abort Point A: User clicks "Cancel" in confirmation dialog**
  - [ ] Return to menu, no changes
  - [ ] `git stash list` unchanged
  - [ ] `git status` unchanged
  
  **Abort Point B: User presses ESC during stash phase**
  - [ ] Read `.git/TIT_DIRTY_OP` snapshot
  - [ ] Restore: `git checkout <original_branch>`
  - [ ] Restore: `git reset --hard <original_head>`
  - [ ] Restore: `git stash apply` (if stash was created)
  - [ ] Delete: `.git/TIT_DIRTY_OP`
  - [ ] Result: Exact original state restored
  
  **Abort Point C: User presses ESC during conflict resolver (merge conflicts)**
  - [ ] Run: `git merge --abort` (to undo merge)
  - [ ] Read: `.git/TIT_DIRTY_OP` snapshot
  - [ ] Restore: `git checkout <original_branch>`
  - [ ] Restore: `git reset --hard <original_head>`
  - [ ] Restore: `git stash apply` (if stash was created)
  - [ ] Delete: `.git/TIT_DIRTY_OP`
  - [ ] Result: Exact original state restored
  
  **Abort Point D: User presses ESC during conflict resolver (snapshot apply conflicts)**
  - [ ] Same as C: Full restore to original state

---

## 🧪 Scenario 3: Dirty Pull - Rebase with Conflicts

**Purpose:** Verify dirty pull with rebase strategy (if implemented)

**Status:** Currently not wired in menu (SPEC mentions rebase not implemented)

**Skip for now** — Document as future enhancement

---

## 🧪 Scenario 4: Dirty Pull - Stash Apply Conflicts After Pull

**Purpose:** Conflicts occur during stash apply (not during merge), verify resolver handles it

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
./titest.sh
# Select: 4
# Opens TIT
# Select: Pull (save changes)
```

**Expected Git State After Setup:**
```
Branch: main
WorkingTree: Modified (conflict.txt changed)
Timeline: Behind (remote has changes to conflict.txt)
Operation: Normal
```

**Expected TIT Behavior:**
```
Menu shows:
  ✓ Pull (save changes)

User selects "Pull (save changes)":
  ✓ Confirmation: "Save your changes?"
  ✓ Phase 1: Stash WIP → Success
  ✓ Phase 2a: Merge → SUCCESS (no conflicts in merge)
  ✓ Phase 3: Apply snapshot → CONFLICTS (stash apply conflicts)
  ✓ ConflictResolver shown (dirty_pull_snapshot_reapply)
  ✓ User marks files + ENTER
  ✓ Phase 4: Finalize → Return to menu
```

### Checklist for Scenario 4:

- [ ] **#1 State Detection**
  - [ ] After setup: `git status` shows "modified: conflict.txt"
  - [ ] `git log --oneline HEAD..origin/main` shows 1+ commits (behind)
  - [ ] Merge will succeed (no remote conflicts on this file yet)
  - [ ] Stash apply will conflict (WIP + incoming remote both touch conflict.txt)
  
- [ ] **#2 Correct State Display**
  - [ ] TIT header: `Branch: main | Modified | Behind`
  
- [ ] **#3 Correct Menu**
  - [ ] "Pull (save changes)" visible [d]
  
- [ ] **#4 Confirmation Guard**
  - [ ] Dialog: "Save your changes?"
  - [ ] 3 buttons work as expected
  
- [ ] **#5 Operation Guarantee**
  - [ ] Stash will succeed (always)
  - [ ] Merge will succeed (no conflicts in our testing setup)
  - [ ] Stash apply WILL conflict (this is the test case)
  - [ ] Conflict resolver must appear automatically
  
- [ ] **#6 Fail Prevention**
  - [ ] Same pre-checks as Scenario 2
  
- [ ] **#7 Abort Safety**
  - **After merge succeeds, stash apply conflicts:**
    - [ ] Stash apply partially applied (state is dirty)
    - [ ] ESC: Abort stash apply (`git reset --hard <post-merge-head>`)
    - [ ] Stash is still available for manual cleanup
    - [ ] Return to original state before dirty pull started

---

## 🧪 Scenario 5: Dirty Pull - Clean Pull (No Conflicts)

**Purpose:** Happy path — stash, pull, apply, no conflicts, auto-return

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
./titest.sh
# Select: 5
# Opens TIT
# Select: Pull (save changes)
```

**Expected Git State After Setup:**
```
Branch: main
WorkingTree: Modified (safe_wip.txt)
Timeline: Behind
Operation: Normal
```

**Expected TIT Behavior:**
```
Menu shows:
  ✓ Pull (save changes)

User selects "Pull (save changes)":
  ✓ Confirmation: "Save your changes?"
  ✓ Phase 1: Stash → Success
  ✓ Phase 2a: Merge → Success (no conflicts)
  ✓ Phase 3: Apply snapshot → Success (no conflicts)
  ✓ Phase 4: Finalize → Cleanup
  ✓ AUTO-RETURN TO MENU (Operation=Normal, WorkingTree=Modified)
  ✓ No user interaction after ENTER
```

### Checklist for Scenario 5:

- [ ] **#1 State Detection**
  - [ ] After setup: `git status` shows "untracked: safe_wip.txt" or "modified: safe_wip.txt"
  - [ ] `git log --oneline HEAD..origin/main` shows 1+ commits (behind)
  - [ ] Merge will succeed (remote changes don't touch safe_wip.txt)
  - [ ] Stash apply will succeed (no conflicts)
  
- [ ] **#2 Correct State Display**
  - [ ] TIT header: `Branch: main | Modified | Behind`
  
- [ ] **#3 Correct Menu**
  - [ ] "Pull (save changes)" visible [d]
  
- [ ] **#4 Confirmation Guard**
  - [ ] Dialog appears and works
  
- [ ] **#5 Operation Guarantee**
  - [ ] All phases will succeed (no conflicts)
  - [ ] Operation is guaranteed to complete
  - [ ] No user interaction needed after confirmation
  
- [ ] **#6 Fail Prevention**
  - [ ] Pre-checks pass (no ongoing ops, can stash, remote exists)
  
- [ ] **#7 Abort Safety**
  - **Mid-operation abort scenarios:**
    - [ ] ESC during any phase: Full restore
    - [ ] `.git/TIT_DIRTY_OP` cleanup
    - [ ] Stash cleanup if created

---

## 📊 Testing Execution Checklist

Before running any scenarios:

- [ ] Build TIT: `./build.sh` (clean compile)
- [ ] Reset test repo: Run scenario 0
- [ ] Verify test repo state: `./titest.sh` → select `s` (show status)

For each scenario:

- [ ] Setup: Run `./titest.sh` → select scenario number
- [ ] Verify pre-condition state (git status, branch, timeline)
- [ ] Open TIT: `./tit_x64`
- [ ] Verify TIT state display matches git state
- [ ] Verify menu options match specification
- [ ] Execute operation or navigate menu
- [ ] Verify each operation phase completes
- [ ] On success: Verify final state matches expected
- [ ] Test abort flow (if applicable): ESC at each phase
- [ ] Verify restore on abort: `git status`, `git log`, `git stash list`

---

## 🔄 Abort Scenario Testing (Critical)

For each scenario with dirty pull:

**Test 1: Abort in confirmation dialog**
```
Menu → Select "Pull (save changes)" → See dialog
ESC → Confirm returns to menu
Verify: git status unchanged
```

**Test 2: Abort during Phase 1 (stash)**
```
(If stash takes time to display)
ESC → Abort stash operation
Verify: Full restore to original state
```

**Test 3: Abort during Phase 2 (merge/pull)**
```
If merge has conflicts: 
  ESC in ConflictResolver → Full abort
Verify: git merge --abort, restore stash, restore state
```

**Test 4: Abort during Phase 3 (stash apply)**
```
If stash apply conflicts:
  ESC in ConflictResolver → Full abort
Verify: git reset --hard <post-merge>, clean stash
```

---

## 📝 Notes

- **Golden Rule:** If user can abort at any point, all state must be restorable
- **SSOT:** `.git/TIT_DIRTY_OP` file is source of truth for dirty operations
- **Thread Safety:** All operations run in worker goroutine, state updates via messages
- **Confirmation:** Prevents accidental data loss, but only for initially destructive ops
- **Pre-checks:** Prevent offering options that would fail

---

**End of Checklist**
