# TIT Testing Status

**Date:** 2026-01-07
**Overall Status:** ⚠️ SCENARIO 1 VERIFIED | SCENARIOS 2-5 IMPLEMENTED BUT UNTESTED

---

## Overview

All dirty pull infrastructure is implemented and wired. Scenario 1 (pull with conflicts, clean tree) has been tested end-to-end. Scenarios 2-5 require manual testing using `titest.sh` setup.

---

## Test Scenarios

### ✅ Scenario 1: Pull with Conflicts (Clean Tree)

**Status:** VERIFIED WORKING

**Setup:** `titest.sh` option 1
- Working tree: CLEAN
- Remote ahead with diverging commits
- Both LOCAL and REMOTE modified same file

**Menu Path:** Pull from remote (p) → conflicts detected → Conflict resolver UI

**Testing Result (Session 44):**
- ✅ Conflicts detected correctly
- ✅ Resolver UI appears with correct file list
- ✅ Column labels: BASE, LOCAL (yours), REMOTE (theirs)
- ✅ User can mark files with SPACE
- ✅ ENTER finalizes: stages all, commits merge, returns Clean + Ahead
- ✅ ESC aborts: `git merge --abort` → `git reset --hard`, returns Clean + Diverged
- ✅ All messages from SSOT (no hardcoded strings)

**Code Path:**
```
cmdPull() 
  → git pull --no-rebase
  → Conflicts detected
  → setupConflictResolver("pull_merge", columns)
  → ModeConflictResolve
  → ENTER: cmdFinalizePullMerge() → git add -A → git commit
  → ESC: cmdAbortMerge() → git merge --abort → git reset --hard
```

---

### 🔧 Scenario 2: Dirty Pull - Merge Conflicts in Pull Phase

**Status:** IMPLEMENTED, NEEDS TESTING

**Setup:** `titest.sh` option 2
- Working tree: DIRTY (wip.txt - separate file, won't conflict)
- Local and Remote both modified conflict.txt differently
- Pull strategy: merge (default)
- Expected conflicts: Merge phase will conflict

**Menu Path:** Pull (save changes) (d) → Save/Discard dialog → 4-phase operation

**Expected Flow:**
```
Phase 1: cmdDirtyPullSnapshot()
  → git stash (saves wip.txt)
  → Success: continue to Phase 2

Phase 2: cmdDirtyPullMerge()
  → git pull --no-rebase
  → CONFLICTS DETECTED on conflict.txt (LOCAL vs REMOTE)
  → setupConflictResolver("dirty_pull_changeset_apply")
  → ModeConflictResolve
  → User resolves conflicts, presses ENTER
  → Continue to Phase 3

Phase 3: cmdDirtyPullApplySnapshot()
  → git stash apply
  → NO CONFLICTS (wip.txt is separate file)
  → Success: continue to Phase 4

Phase 4: cmdDirtyPullFinalize()
  → git stash drop
  → Clean up snapshot metadata
  → Return to menu
```

**To Test:**
1. Run `titest.sh` option 2
2. In TIT: Select "Pull (save changes)" (d)
3. Select "Save changes" in confirmation dialog
4. Expect conflict resolver after Phase 2 (merge)
5. Mark files and finalize conflicts
6. Watch Phase 3-4 complete automatically
7. Verify final state clean and up-to-date

---

### 🔧 Scenario 3: Dirty Pull - Stash Apply Conflicts (Pull Succeeds)

**Status:** IMPLEMENTED, NEEDS TESTING

**Setup:** `titest.sh` option 3
- Working tree: DIRTY (conflict.txt modified - overlaps with remote change)
- Remote modified conflict.txt
- Pull strategy: merge (default)
- Expected conflicts: Stash apply phase will conflict

**Menu Path:** Pull (save changes) (d) → Save/Discard dialog → 4-phase operation

**Expected Flow:**
```
Phase 1: cmdDirtyPullSnapshot()
  → git stash (saves conflict.txt changes)
  → Success: continue to Phase 2

Phase 2: cmdDirtyPullMerge()
  → git pull --no-rebase
  → NO CONFLICTS (merge succeeds cleanly)
  → Success: continue to Phase 3

Phase 3: cmdDirtyPullApplySnapshot()
  → git stash apply
  → CONFLICTS DETECTED (both sides modified conflict.txt)
  → setupConflictResolver("dirty_pull_snapshot_reapply")
  → ModeConflictResolve
  → User resolves conflicts, presses ENTER
  → Continue to Phase 4

Phase 4: cmdDirtyPullFinalize()
  → git stash drop
  → Cleanup
  → Return to menu
```

**To Test:**
1. Run `titest.sh` option 3
2. In TIT: Select "Pull (save changes)" (d)
3. Select "Save changes"
4. Watch Phase 1-2 complete without conflicts
5. Expect conflict resolver in Phase 3 (stash apply)
6. Mark files and finalize conflicts
7. Verify final state clean and up-to-date

---

### 🔧 Scenario 4: Dirty Pull - Clean (No Conflicts Anywhere)

**Status:** IMPLEMENTED, NEEDS TESTING

**Setup:** `titest.sh` option 4
- Working tree: DIRTY (safe_wip.txt - separate file, won't conflict)
- Remote has new commits but no overlap with local files
- Pull strategy: merge (default)
- Expected conflicts: NONE

**Menu Path:** Pull (save changes) (d) → Save/Discard dialog → All 4 phases auto-complete

**Expected Flow:**
```
Phase 1: cmdDirtyPullSnapshot()
  → git stash (saves safe_wip.txt)
  → Success: continue to Phase 2

Phase 2: cmdDirtyPullMerge()
  → git pull --no-rebase
  → NO CONFLICTS
  → Success: continue to Phase 3

Phase 3: cmdDirtyPullApplySnapshot()
  → git stash apply
  → NO CONFLICTS (safe_wip.txt is separate)
  → Success: continue to Phase 4

Phase 4: cmdDirtyPullFinalize()
  → git stash drop
  → Cleanup
  → AUTO-RETURN TO MENU (all phases completed without user intervention)
```

**To Test:**
1. Run `titest.sh` option 4
2. In TIT: Select "Pull (save changes)" (d)
3. Select "Save changes"
4. Watch all phases complete automatically
5. No conflict resolver should appear
6. Verify auto-return to menu with updated state

---

## Test Execution Guide

### Prerequisites
```bash
cd /path/to/test/repo
bash titest.sh
```

This repo should have:
- Remote origin (GitHub)
- main branch
- conflict.txt file (some content)
- Ability to create/delete files

### Per-Scenario Workflow

1. **Setup:** Run option 0 or specific scenario (1-5) in titest.sh
2. **Show Status:** Use option 's' to verify setup
3. **Launch TIT:** `./tit_x64` (from test repo)
4. **Execute:** Choose menu option
5. **Verify:** Check output, resolver UI, final state
6. **Cleanup:** Run option 0 when done

### What to Verify

**Console Output:**
- All messages from SSOT (no hardcoded strings)
- Phase progress messages appear
- Error messages are clear and actionable

**Resolver UI (if conflicts):**
- File list shows actual conflicted files
- Column labels correct (BASE, LOCAL, REMOTE)
- SPACE toggles file selection
- TAB cycles panes
- Footer hints are helpful

**Final State:**
- `git status` shows correct state
- Commit history shows expected commits
- No stashes left behind
- Branch tracking correct

### Known Issues / Gotchas

1. **Snapshot marker:** Dirty pull stashes use "TIT DIRTY-PULL SNAPSHOT" marker
   - Check with `git stash list` after abort
   - Should be cleaned up on successful finalize

2. **Abort safety:** ESC during any phase should restore original state
   - Test by pressing ESC in resolver and checking `git status`

3. **Race condition risk:** None identified, but test rapid succession of operations

4. **Terminal size:** Some tests may fail if terminal < 80 chars wide
   - Expand terminal if needed

---

## Summary

| Scenario | Setup | Expected Conflict Point | Status | Testing |
|----------|-------|-------------------------|--------|---------|
| 1: Clean pull (clean tree) | Diverging commits on conflict.txt | Merge phase | ✅ Verified | ✅ Done |
| 2: Dirty pull (merge conflicts) | Diverging commits on conflict.txt + WIP in separate file | Merge phase | 🟡 Implemented | ⏳ TODO |
| 3: Dirty pull (stash conflicts) | Remote modified file + WIP modifies same file | Stash apply phase | 🟡 Implemented | ⏳ TODO |
| 4: Dirty pull (clean, no conflicts) | Remote changes + WIP in separate file | NONE | 🟡 Implemented | ⏳ TODO |

**What You Just Tested (Scenario 2 as originally written):**
- ❌ Label said "merge with conflicts" but actually had NO conflicts
- ✅ Setup: Clean local → Behind remote → WIP in separate file
- ✅ Result: Stash → Fast-forward merge (no conflicts) → Stash apply (no conflicts) → Auto-return
- ℹ️ This is what Scenario 4 should test (clean, no conflicts)

**Next Priority:** Test corrected scenarios 2-4 using updated titest.sh and document results.

---

**End of Document**
