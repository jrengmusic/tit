# Time Travel Testing Checklist

**Purpose:** Comprehensive test scenarios for each phase. Covers happy paths, edge cases, and cancellations.

---

## Test Repository Setup

**Test repo location:**

```bash
/Users/jreng/Documents/Poems/inf/t
```

This repo already has M1-M14 history. Use it for all Phase 1-6 tests.

**History summary:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
git log --oneline  # M1-M14 commits available
```

**Create branches for specific testing:**

```bash
# Branch for conflict testing
git checkout M3  # Commit with version 1.2
git checkout -b conflict-branch
echo "version 2.0" > version.txt
git add version.txt && git commit -m "conflict: Version 2.0"
git checkout main  # Back to main at M5

# Keep main on M5 for most testing
```

**Map commits to test scenarios:**

| Commit | Hash (short) | Message | Use Case |
|--------|--------------|---------|----------|
| M1 | (oldest) | Initial version | Old commit to travel to |
| M2 | | Bump to 1.1 | Merging this back → no conflict |
| M3 | | Bump to 1.2 | Merging this back → conflict (diverged) |
| M4 | | Add API | Different file, no conflict |
| M5 | (HEAD) | Finalize API | Current main |

---

## Phase 1: Basic Time Travel (Clean Working Tree)

### Test 1.1: Happy Path - Time Travel to M2

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
~/Documents/Poems/inf/tit/tit_x64
```

**Steps:**
1. Select "Commit history"
2. Navigate up to M2: Bump to 1.1
3. Press ENTER
4. See confirmation dialog
5. Press ENTER to confirm
6. See console: "Time traveling... → Time travel successful"
7. Press ESC
8. See menu with 3 items:
   - 🕐 Commit history
   - 🔙 Return to main
   - 📦 Merge & return to main

**Expected:**
- ✅ Console shows complete message
- ✅ Header shows time travel indicator (🕐)
- ✅ Menu items appear (not grayed out)
- ✅ Can read files at M2 state

**Verify Git State:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
git rev-parse HEAD  # Should show M2 hash (0000d86)
git symbolic-ref --short HEAD  # Should fail (detached)
ls -la .git/TIT_TIME_TRAVEL  # Should exist, contains "main"
```

---

### Test 1.2: ESC at Confirmation Dialog

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
~/Documents/Poems/inf/tit/tit_x64
```

**Steps:**
1. Select "Commit history"
2. Navigate to M2
3. Press ENTER (confirmation dialog appears)
4. Press ESC

**Expected:**
- ✅ Confirmation dialog closes
- ✅ Back in history mode at M2
- ✅ No checkout happened
- ✅ Still on main at M5

**Verify:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
git rev-parse HEAD  # Should show M5 (5daf3a9)
```

---

### Test 1.3: Navigate Different Commits in History

**Setup:** Time traveling at M2

**Steps:**
1. After time travel to M2, see menu
2. Select "Browse history" 
3. Navigate down to M4
4. Press ENTER (confirmation to jump)
5. See console: "Jumping to M4..."
6. ESC back to menu
7. Check current commit

**Expected:**
- ✅ Can navigate while time traveling
- ✅ ENTER jumps to new commit
- ✅ Header updates to show new commit
- ✅ Git state shows new commit

---

### Test 1.4: Return Without Changes

**Setup:** Time traveling at M2 (clean)

**Steps:**
1. In time travel menu at M2
2. Select "⬅️  Return to main"
3. See confirmation: "Return to main?"
4. Press ENTER to confirm
5. See console: "Returning to main..."
6. Press ESC
7. Check git state

**Expected:**
- ✅ Returns to main at M5
- ✅ Header shows normal state (no 🕐)
- ✅ `.git/TIT_TIME_TRAVEL` deleted
- ✅ `git rev-parse HEAD` shows M5

---

### Test 1.5: ESC at Return Confirmation

**Setup:** Time traveling at M2

**Steps:**
1. Select "⬅️  Return to main"
2. Confirmation appears
3. Press ESC

**Expected:**
- ✅ Still time traveling at M2
- ✅ No checkout happened
- ✅ Menu still shows time travel options

---

## Phase 2: Dirty Working Tree (Stash Protocol)

### Test 2.1: Happy Path - Time Travel with Dirty Tree

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
echo "wip" >> version.txt
~/Documents/Poems/inf/tit/tit_x64
```

**Steps:**
1. Select "Commit history"
2. Navigate to M2
3. Press ENTER
4. See confirmation dialog
5. Press ENTER to confirm
6. See dirty protocol dialog: "You have uncommitted changes"
7. Press ENTER (Stash & continue)
8. See console: "Stashing → Time traveling..."
9. Press ESC → menu at M2

**Expected:**
- ✅ Dirty protocol dialog shown
- ✅ Changes stashed: `git stash list` shows stash entry
- ✅ Console shows both stash + checkout
- ✅ At M2, working tree is clean
- ✅ `.git/TIT_TIME_TRAVEL` exists

**Verify:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
git stash list  # Should show stash
git status  # Should be clean
```

---

### Test 2.2: Cancel Dirty Protocol

**Setup:** Same as 2.1, at step 8

**Steps:**
1. Dirty protocol dialog shows
2. Press "Discard changes and proceed" (left button)

**Expected:**
- ✅ Stays on main at M5
- ✅ Dirty changes still there: `git status --short` shows edits
- ✅ No stash made
- ✅ Menu returns

---

### Test 2.3: ESC at Dirty Protocol

**Setup:** Same as 2.1, at step 8

**Steps:**
1. Dirty protocol dialog shows
2. Press ESC

**Expected:**
- ✅ Cancels time travel entry
- ✅ Still on main at M5
- ✅ Dirty changes preserved
- ✅ No stash made

---

### Test 2.4a: Merge with "Commit & merge"

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
git reset --hard HEAD
git clean -fd
echo "original wip" >> version.txt
~/Documents/Poems/inf/tit/tit_x64
```

**Steps:**
1. Time travel to M2 (original dirty work stashed)
2. ESC (exit console)
3. In another terminal:
   ```bash
   echo "travel feature" > travel.txt
   git add travel.txt
   git commit -m "travel: feature"
   ```
4. Back in tit, select "📦 Merge & return to main"
5. Dialog appears: "[Commit & merge] [Discard]"
6. Select "Commit & merge"
7. Prompted for message (accept or edit)
8. Press ENTER (commits travel changes, merges to main)
9. Verify back on main with both original + travel changes

**Expected:**
- ✅ Dialog appears with both options
- ✅ Travel commit merged to main
- ✅ Original dirty work restored
- ✅ Both changes on main

---

### Test 2.4b: Merge with "Discard"

**Setup:** Same as 2.4a (time travel, make changes, commit)

**Steps:**
1. Time travel to M2, make travel commit (same as 2.4a steps 1-3)
2. Back in tit, select "📦 Merge & return to main"
3. Dialog: "[Commit & merge] [Discard]"
4. Select "Discard"
5. Travel changes hard reset (lost)
6. Merges to main (original work restored)

**Expected:**
- ✅ Dialog appears
- ✅ Travel changes discarded (git hard reset)
- ✅ Returns to main with original work

---

### Test 2.4c: Return with Dirty Changes

**Setup:**
```bash
cd /Users/jreng/Documents/Poems/inf/t
git reset --hard HEAD
git clean -fd
echo "original wip" >> version.txt
~/Documents/Poems/inf/tit/tit_x64
```

**Steps:**
1. Time travel to M2 (original dirty work stashed)
2. ESC (exit console)
3. In another terminal, make uncommitted changes:
   ```bash
   echo "uncommitted travel work" > uncommitted.txt
   ```
4. Back in tit, select "🔙 Return to main"
5. Dialog appears: "[Discard & return] [Cancel]"
6. Select "Discard & return"
7. Uncommitted changes hard reset (lost)
8. Returns to main with original work restored

**Expected:**
- ✅ Dialog appears explaining changes will be discarded
- ✅ Uncommitted changes discarded (git hard reset)
- ✅ Returns to main with original work

---

## Phase 3: Merge Back (No Conflicts)

### Test 3.1: Happy Path - Merge M2 to main (No Conflict)

**Setup:**
- Clean working tree on main at M5
- M2 is ancestor of M5 (no divergence)
- Time traveling at M2

**Steps:**
1. Time travel to M2 (no local changes)
2. Select "📦 Merge & return to main"
3. Confirmation: "Merge abc1234 (M2) into main?"
4. Press ENTER
5. See console: "Checking out main → Merging M2 → Complete"
6. Press ESC

**Expected:**
- ✅ M2 merged into main (linear history, fast-forward)
- ✅ `git log --oneline` shows M5 on top
- ✅ Back on main, clean
- ✅ Menu shows normal state (no time travel)

---

### Test 3.2: Merge with Local Changes (No Conflict)

**Setup:**
- Time traveling at M4
- Made changes: `echo "api improved" > feature.txt`
- Original work S1 was clean (no stash)

**Steps:**
1. Select "📦 Merge & return to main"
2. Confirmation: "Merge M4?"
3. Press ENTER
4. See console: "Stashing changes → Checking out main → Merging M4 → Applying changes"
5. Press ESC

**Expected:**
- ✅ Changes stashed before checkout
- ✅ M4 merged clean
- ✅ Changes reapplied
- ✅ Dirty tree (with merged code + local changes)

---

### Test 3.3: Cancel Merge

**Setup:** Time traveling at M2

**Steps:**
1. Select "📦 Merge & return to main"
2. Confirmation dialog
3. Press "Cancel" (right button)

**Expected:**
- ✅ Still time traveling
- ✅ No merge executed
- ✅ Menu still shows time travel options

---

## Phase 4: Merge Conflicts & Resolution

### Test 4.1: Merge Conflict (Diverged History)

**Setup:**
- Clean working tree on main at M5
- conflict-branch (created in setup) has diverged M3
- Time travel to conflict-branch commit

**Steps:**
1. Time travel to conflict-branch commit (version 2.0)
2. Select "📦 Merge & return to main"
3. Confirmation
4. Press ENTER
5. See console: "Merging... ⚠️ CONFLICT on version.txt"
6. ESC from console → ConflictResolver appears

**Expected:**
- ✅ Conflict detected and shown
- ✅ ConflictResolver shows:
  - LOCAL: main's "version 1.2" (from M3 in main)
  - REMOTE: "version 2.0" (from conflict-branch)
- ✅ User can mark lines to keep

---

### Test 4.2: Resolve Merge Conflict, Then Conflict on Original Stash

**Setup:**
- Time travel with dirty original work
- Travel to diverged commit
- Merge creates conflict

**Steps:**
1. Time travel to conflict-branch (with dirty tree)
2. Original work S1 is stashed
3. Select "Merge & return"
4. ConflictResolver shows merge conflict
5. Resolve conflict (keep LOCAL version)
6. Press ENTER (continue)
7. Console: "Applying time travel changes..."
8. If conflict on T1: ConflictResolver shows again
9. Resolve
10. Console: "Restoring original work..."
11. If conflict on S1: ConflictResolver shows again
12. Resolve
13. ESC → menu on main

**Expected:**
- ✅ Can have 3 sequential conflict resolvers
- ✅ Each one resolvable
- ✅ No work lost at any step
- ✅ Final state on main with all changes merged

---

### Test 4.3: ESC During Conflict Resolution

**Setup:** ConflictResolver showing merge conflict

**Steps:**
1. ConflictResolver visible
2. Start resolving (mark some lines)
3. Press ESC (abort)

**Expected:**
- ✅ Merge aborted: `git merge --abort`
- ✅ Back in time travel mode at original commit
- ✅ Nothing merged
- ✅ Original work still stashed

---

## Phase 5: Browse While Time Traveling

### Test 5.1: Jump Commits While Time Traveling

**Setup:** Time traveling at M2

**Steps:**
1. Select "🕐 Browse history"
2. See commit list (M5, M4, M3, M2, M1)
3. Navigate down to M4
4. Press ENTER
5. Confirmation: "Save changes before jumping?" (if dirty)
6. ESC from history → back to time travel menu
7. Check current commit

**Expected:**
- ✅ Can browse commits
- ✅ ENTER jumps to selected
- ✅ Header updates to show new commit
- ✅ ESC returns to menu

---

### Test 5.2: Jump With Local Changes (Requires Stash)

**Setup:**
- Time traveling at M2
- Made changes: `echo "travel edit" > travel.txt`

**Steps:**
1. Select "Browse history"
2. Navigate to M4
3. Press ENTER
4. See dialog: "Save changes before jumping?"
5. Press ENTER (stash & jump)
6. See console: "Jumping to M4..."
7. ESC → menu

**Expected:**
- ✅ Changes stashed: `git stash list` shows temp stash
- ✅ Jumped to M4
- ✅ Previous changes lost (separate from original stash S1)
- ✅ Can still merge back or return (will ignore this jump's stash)

---

## Phase 6: Return Without Merge

### Test 6.1: Return with No Local Changes

**Setup:** Time traveling at M2 (clean)

**Steps:**
1. Select "⬅️  Return to main"
2. Confirmation: "Discard changes and return to main?"
3. Press ENTER
4. See console: "Returning to main..."
5. Press ESC

**Expected:**
- ✅ Back on main at M5
- ✅ Time travel state cleared
- ✅ Header shows normal state

---

### Test 6.2: Return with Local Changes (Discarded)

**Setup:**
- Time traveling at M2
- Made changes: `echo "lost" > lost.txt`

**Steps:**
1. Select "⬅️  Return to main"
2. Confirmation explains "Your changes will be DISCARDED"
3. Press ENTER
4. See console: "Discarding changes → Returning to main"
5. Press ESC

**Expected:**
- ✅ Time travel changes DISCARDED (git checkout .)
- ✅ Back on main
- ✅ `git status --short` clean
- ✅ Time travel edit is gone (expected!)

---

### Test 6.3: Return with Original Stash

**Setup:**
- Started time travel with dirty tree (S1 stashed)
- Now returning without merge

**Steps:**
1. Select "⬅️  Return to main"
2. Confirmation
3. Press ENTER
4. See console: "Returning to main → Restoring original work"
5. Press ESC

**Expected:**
- ✅ Back on main at M5
- ✅ Original stash applied: S1 restored
- ✅ `git status --short` shows original dirty changes
- ✅ Exactly where we started

---

### Test 6.4: ESC at Return Confirmation

**Setup:** Time traveling at M2

**Steps:**
1. Select "⬅️  Return to main"
2. Confirmation
3. Press ESC

**Expected:**
- ✅ Still time traveling
- ✅ No git operations executed

---

## Edge Cases & Stability

### Test E1: Very Old Commit (M1)

**Setup:** Time travel to oldest commit M1

**Steps:**
1. Navigate to M1
2. Time travel
3. Browse history, jump back to M5
4. Merge back

**Expected:**
- ✅ Works same as any other commit
- ✅ No special handling needed

---

### Test E2: Multiple ESC Sequences

**Setup:** Time traveling at M2

**Steps:**
1. ESC from menu → console
2. ESC from console → menu
3. ESC from menu → back to normal (back on main? or stay in time travel menu?)

**Expected:**
- ✅ ESC from time travel menu → nothing (already in menu)
- ✅ Need explicit "Return" or "Merge" to exit
- ✅ No accidental exits

---

### Test E3: Interrupt (Kill Terminal)

**Setup:** Time traveling, mid-operation

**Steps:**
1. Kill terminal (Ctrl+Z, kill -9)
2. Restart tit in same repo

**Expected:**
- ✅ `.git/TIT_TIME_TRAVEL` still exists
- ✅ Detects time traveling state
- ✅ Shows time travel menu
- ✅ User can continue merge/return

---

### Test E4: Concurrent Stashes

**Setup:**
- Original stash S1
- Time travel stash T1 (changes while traveling)
- Manual stash user made: S2

**Steps:**
1. Time travel to M2 from dirty tree (S1 made)
2. Make changes (T1 will be made)
3. Merge back
4. Console shows: stashing T1, merging, applying T1, restoring S1
5. Also check for S2 existence

**Expected:**
- ✅ All stashes handled correctly
- ✅ S1 and T1 applied in order
- ✅ S2 (user's stash) untouched

---

## Full Flow Tests

### Test F1: Complete Happy Path (Clean to Clean)

**Steps:**
1. On main M5, clean
2. Time travel to M2
3. Browse, jump to M4
4. Browse back to M2
5. Merge & return
6. Back on main, clean, M5

**Result:** All phases working together

---

### Test F2: Complex Path (Dirty → Stash → Travel → Edit → Merge → Conflict → Resolve → Original Stash Conflict → Final)

**Steps:**
1. On main, dirty (S1)
2. Time travel to conflict-branch (dirty protocol)
3. Edit file (T1)
4. Merge & return
5. Merge conflict on step 1 (resolve)
6. Conflict on T1 (resolve)
7. Conflict on S1 (resolve)
8. Back on main with all changes

**Result:** Maximum complexity handled

---

## Regression Tests

### Test R1: Normal Operations Still Work

**After Phase 6 complete, verify:**
- [ ] Commit works (normal branch)
- [ ] Pull works (normal branch)
- [ ] Push works (normal branch)
- [ ] History mode works
- [ ] File history works
- [ ] Conflict resolver works (on normal merge)
- [ ] No time travel menu shows when not time traveling

---

## Acceptance Criteria (All Phases)

**Before marking "COMPLETE":**

- [ ] All tests in Phase 1 pass
- [ ] All tests in Phase 2 pass
- [ ] All tests in Phase 3 pass
- [ ] All tests in Phase 4 pass
- [ ] All tests in Phase 5 pass
- [ ] All tests in Phase 6 pass
- [ ] All edge cases pass
- [ ] All full flow tests pass
- [ ] No regressions in other modes
- [ ] Binary builds clean
- [ ] No race conditions (go test -race)

---

## Test Tracking

| Phase | Tests | Status | Notes |
|-------|-------|--------|-------|
| **1** | 1.1 | ✅ | PASS - Time travel to M2 works |
| **1** | 1.2 | ✅ | PASS - ESC at confirmation cancels |
| **1** | 1.3 | ✅ | PASS - Jump between commits while traveling |
| **1** | 1.4 | ✅ | PASS - Return to main, marker deleted |
| **1** | 1.5 | ✅ | PASS - ESC at return confirmation stays traveling |
| **2** | 2.1 | ✅ | PASS - Dirty stashed, restored on return |
| **2** | 2.2 | ⊘ | SKIP - Design allows automatic stash |
| **2** | 2.3 | ✅ | PASS - ESC at dirty protocol cancels |
| **2** | 2.4a | ✅ | PASS - Merge with commit & merge |
| **2** | 2.4b | ✅ | PASS - Merge with discard |
| **2** | 2.4c | ✅ | PASS - Return with dirty changes discarded |
| **3** | 3.1 | ✅ | PASS - Merge M2 to main (no conflict) |
| **3** | 3.2 | ✅ | PASS - Merge with local changes |
| **3** | 3.3 | ✅ | PASS - Cancel merge confirmation |
| **4** | 4.1 | ✅ | PASS - Merge with conflict resolution |
| **4** | 4.2-4.3 | ⬜ | Next |
| **5** | 5.1-5.2 | ⬜ | After Phase 4 |
| **6** | 6.1 | ✅ | PASS - Return with no changes |
| **6** | 6.2 | ✅ | PASS - Return with local changes discarded |
| **6** | 6.3 | ✅ | PASS - Return with original stash restored |
| **6** | 6.4 | ✅ | PASS - ESC at return confirmation |
| **E** | E1 | ✅ | PASS - Very old commit (M1) |
| **E** | E2 | ✅ | PASS - Multiple ESC sequences |
| **E** | E3 | ✅ | PASS - Interrupt and restart |
| **E** | E4 | ✅ | PASS - Concurrent stashes |
| **F** | F1 | ✅ | PASS - Complete happy path |
| **F** | F2 | ✅ | PASS - Complex path with merge |
| **R** | R1 | ✅ | PASS - Normal operations still work |

---

**Recommended Approach:**
1. Run Phase 1 tests completely before coding Phase 2
2. Each test should be executable within 2 minutes
3. Automate setup: create `test-setup.sh` for repo creation
4. Keep test notes for debugging

