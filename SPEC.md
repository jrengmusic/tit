# TIT — Core Specification v2.0 (Canon/Working Branch Architecture)

**TIT:** Terminal UI for Git
**Target:** Go + Bubble Tea + Lip Gloss
**Philosophy:** Deterministic state machine. Canon/Working branch separation. Zero surprises. Beautiful rendering.

---

## 1. Technology Stack

- **Language:** Go ≥ 1.21
- **Framework:** Bubble Tea (state machine) + Lip Gloss (rendering)
- **Git Interface:** `os/exec` only, no libraries
- **Output:** Single static binary

---

## 2. Foundational Principles

### 2.1 Core Philosophy

**TIT's UI is a pure function of Git state + Branch context.**

```
(Git State, Branch Context) → Allowed Actions → Menu
```

If Git would reject an action, it must not appear in the menu.

### 2.2 Dual-Branch Architecture

**Canon Branch (e.g., "main")**
- Read-only locally (no commits allowed)
- Clean history, production-ready
- **Allowed operations:** Push, Pull, Replace Local, Replace Remote
- **Workflow role:** Publishing channel to remote

**Working Branch (e.g., "dev")**
- Full operations (commit, merge, rebase, cherry-pick, time travel)
- Messy operations allowed (stash, amend, force push)
- **Workflow role:** Development sandbox

**Development Workflow:**
```
Work on working branch → Commit changes → Merge to canon → Push canon to remote
```

---

## 3. State Model

Every decision in TIT derives from **five axes**:

### WorkingTree — Local file changes
| Code | Meaning |
|------|---------|
| `Clean` | No changes |
| `Modified` | Has changes (staged, unstaged, or both) |

**Note:** TIT doesn't distinguish staging states. All changes commit together.

### Timeline — Local vs remote comparison
| Code | Meaning |
|------|---------|
| `InSync` | Local == Remote |
| `Ahead` | Local ahead (unpushed commits) |
| `Behind` | Local behind (unpulled commits) |
| `Diverged` | Both have unique commits |

### Operation — Git operation state
| Code | Meaning |
|------|---------|
| `Normal` | No operation in progress |
| `Merging` | Merge in progress |
| `Rebasing` | Rebase in progress |
| `Conflicted` | Operation stopped due to conflicts |
| `TimeTraveling` | Detached HEAD (exploring history) |
| `DirtyOperation` | Executing dirty pull/cherry-pick |

### Remote — Remote repository presence
| Code | Meaning |
|------|---------|
| `NoRemote` | No remote configured |
| `HasRemote` | Remote exists |

### BranchContext — Current branch classification
| Code | Meaning |
|------|---------|
| `Canon` | On canon branch (restricted operations) |
| `Working` | On working branch (full operations) |
| `Other` | On other branch (feature branches, full operations) |
| `None` | No branch context (detached HEAD, no repo) |

**State Tuple:** `(WorkingTree, Timeline, Operation, Remote, BranchContext)`

---

## 4. State Priority Rules

**Priority 1: Operation State** (Most Restrictive)
- `Conflicted` → Show ONLY conflict resolution menu
- `Merging/Rebasing` → Show ONLY operation control menu
- `DirtyOperation` → Show ONLY dirty operation control menu
- `TimeTraveling` → Show ONLY time travel menu (working branch only)
- `Normal` → Proceed to check Branch Context

**Priority 2: Branch Context** (Determines Available Operations)
- `Canon` → Show ONLY canon branch operations (push, pull, replace, switch)
- `Working` → Show ALL operations (commit, merge, rebase, history, time travel)
- `Other` → Show ALL operations (treat as working branch)
- `None` → Show error state or init/clone options

**Priority 3: Remote Presence**
- `NoRemote` → Hide sync actions, show "Add remote"
- `HasRemote` → Enable sync menus based on Timeline

**Priority 4: Timeline + WorkingTree**
- Determines which action menus appear

---

## 5. State → Menu Mapping

### When Operation = Conflicted
**Show ONLY:**
- 🧩 View conflicted files
- ✏️ Resolve conflicts externally
- ▶️ Continue operation
- ⛔ Abort operation

### When Operation = Merging/Rebasing (no conflicts)
**Show ONLY:**
- ▶️ Continue operation
- ⛔ Abort operation
- 🔄 View operation details

### When Operation = DirtyOperation
**Show ONLY:**
- 🔄 View operation status
- ⛔ Abort dirty operation (restores exact original state)

### When Operation = TimeTraveling
**Show ONLY:**
- 🕒 Jump to different commit
- 👁️ View diff (vs working branch)
- ✅ Commit changes (creates orphaned commit)
- ⚠️ **Attach HEAD as new working** (destructive)
- ⬅️ Return to working branch

**Note:** Time travel ONLY available on working/other branches, NEVER on canon.

### When Operation = Normal + BranchContext = Canon

**Canon branch operations are RESTRICTED to sync only.**

#### Timeline Sync Actions

**When Timeline = InSync:**
- 📥 Pull from remote

**When Timeline = Ahead:**
- 📤 Push to remote
- ⚠️ Replace remote with local (force push)

**When Timeline = Behind:**
- 📥 Pull from remote
- ⚠️ Replace local with remote (hard reset)

**When Timeline = Diverged:**
- ⬇️ **Keep remote** (discard local commits)
- ⬆️ **Keep local** (overwrite remote)

#### Always Available
- 🔀 Switch to working branch
- 📜 Browse commit history (read-only, no time travel)

### When Operation = Normal + BranchContext = Working

**Working branch has FULL operations.**

#### Working Tree Actions
| State | Menu Items |
|-------|------------|
| `Clean` | *(no working tree actions)* |
| `Modified` | ✅ Commit changes<br>🚀 Commit and push |

#### Timeline Sync Actions

**When Remote = NoRemote:**
- 🌐 Add remote

**When Remote = HasRemote:**
- All sync actions (push, pull, merge, rebase) based on Timeline state
- Same as old TIT spec (full operations)

#### Working → Canon Operations
- 🔀 **Merge to canon branch** (key new operation)
  - Merges working branch into canon
  - Handles conflicts
  - Optionally pushes after merge

#### Branch Management
- 🔀 Switch to canon branch

#### History Actions (full operations)
- 🕒 Commit History (browse timeline, time travel enabled)
- 📁 File(s) History (browse file changes, cherry-pick enabled)

### When Operation = Normal + BranchContext = Other

**Feature branches treated as working branches** (full operations).

Same menu as working branch, but hints clarify this is a feature branch.

---

## 6. Merge Working → Canon Operation

**Purpose:** Merge working branch changes into canon branch for publishing.

**Pre-conditions:**
- Currently on working branch (or feature branch)
- Working tree must be clean (no uncommitted changes)
- Canon branch must exist

**User sees:**
```
🔀 MERGE TO CANON

Merge: dev → main

This will:
✓ Checkout canon branch (main)
✓ Merge working branch (dev) into canon
✓ Handle conflicts if any
✓ Optionally push to remote after merge

[Proceed] [Cancel]
```

**Implementation steps:**

1. **Checkout canon branch**
   ```bash
   git checkout <canon-branch>
   ```

2. **Merge working branch**
   ```bash
   git merge <working-branch>
   ```
   - If conflicts → Operation = Conflicted
   - User resolves → Continue merge
   - If success → Proceed to step 3

3. **Post-merge (optional)**
   - Show menu on canon branch
   - User can push to remote
   - Or switch back to working branch

**Abort (if conflicts):**
```bash
git merge --abort
git checkout <working-branch>
```
→ Restores working branch, canon unchanged

**Key properties:**
- Safe (conflicts handled immediately)
- Transparent (user sees exactly what's merging)
- Flexible (push is optional after merge)

---

## 7. Canon Branch Restrictions

**What Canon Branch CANNOT Do:**

❌ Commit changes
❌ Amend commits
❌ Create branches
❌ Rebase
❌ Cherry-pick
❌ Time travel
❌ Merge other branches (except via "Merge working → canon")

**What Canon Branch CAN Do:**

✅ Pull from remote (fast-forward or merge)
✅ Push to remote
✅ Replace local with remote (hard reset)
✅ Replace remote with local (force push)
✅ Switch to working branch
✅ Browse history (read-only)

**Enforcement:**

Menu generation filters operations:
```go
if gitState.BranchContext == Canon {
    // ONLY sync operations allowed
    return menuCanonBranch()
}
```

If user somehow commits on canon (outside TIT), TIT will detect it and:
- Show Timeline = Ahead
- Show "Push to remote" option
- But NEVER show "Commit" option again

---

## 8. Dirty Operation Protocol

**Purpose:** Apply any change-set (pull, cherry-pick, time travel) while preserving uncommitted work.

**Used when:**
- Pull with WorkingTree = Modified
- Cherry-pick with WorkingTree = Modified
- Time travel with WorkingTree = Modified

**User sees:**
```
⚠️ You have uncommitted changes

Choose:
1. Save changes and [operation]
   → Your work is preserved
   → May cause conflicts

2. Discard changes and [operation]
   → Your changes will be LOST

[Save and proceed] [Discard and proceed] [Cancel]
```

**Implementation steps:**

1. **Snapshot**
   ```bash
   ORIGINAL_HEAD=$(git rev-parse HEAD)
   git stash push -u -m "TIT DIRTY-OP SNAPSHOT"
   echo "$ORIGINAL_HEAD" > .git/TIT_DIRTY_OP
   ```

2. **Apply change-set** (pull, cherry-pick, checkout, etc.)
   - If conflicts → Operation = Conflicted
   - User resolves → Continue

3. **Apply snapshot back**
   ```bash
   git stash apply
   ```
   - If conflicts → Operation = Conflicted
   - User resolves → Continue

4. **Finalize**
   ```bash
   git stash drop
   rm .git/TIT_DIRTY_OP
   ```

**Abort (at any step):**
```bash
ORIGINAL_HEAD=$(cat .git/TIT_DIRTY_OP)
git reset --hard $ORIGINAL_HEAD
git stash apply
git stash drop
rm .git/TIT_DIRTY_OP
```
→ Restores exact original dirty state

**Key properties:**
- Source-agnostic (works for any Git operation)
- Conflict-first (stops immediately on conflicts)
- Rollback-safe (abort always works)
- No stash stacking (single temporary snapshot)

---

## 9. Commit History Browser (2-Column)

**Purpose:** Browse commit timeline, optionally time travel

**Availability:**
- Canon branch: Read-only (no time travel)
- Working branch: Full access (time travel enabled)

```
┌─────────────────────────┬─────────────────────────┐
│ Commits                 │ Details                 │
│                         │                         │
│ > 2024-12-30 14:23 abc1 │ Commit: abc123f         │
│   2024-12-30 13:15 def4 │ Author: John Doe        │
│   2024-12-29 22:01 789g │ Date: Dec 30, 2024      │
│   ...                   │                         │
│                         │ fix: corrected agent    │
│                         │ logic                   │
│                         │ (full commit message)   │
└─────────────────────────┴─────────────────────────┘
```

**Navigation:**
- ↑↓: Navigate commits
- Enter: Time travel confirmation (ONLY on working/other branches)
- ESC: Return to main menu

**On Canon Branch:**
- Enter key disabled
- Footer shows: "Time travel not available on canon branch"

---

## 10. File(s) History Browser (3-Pane)

**Purpose:** Browse file changes over time, optionally cherry-pick

**Availability:**
- Canon branch: Read-only (no cherry-pick)
- Working branch: Full access (cherry-pick enabled)

```
┌─────────────────────────┬─────────────────────────┐
│ Commits                 │ Files Changed (3)       │
│                         │                         │
│ > 2024-12-30 14:23 abc1 │ > src/agent.cpp         │
│   2024-12-30 13:15 def4 │   src/state.cpp         │
│   2024-12-29 22:01 789g │   include/agent.h       │
│   ...                   │                         │
├─────────────────────────┴─────────────────────────┤
│ Diff: src/agent.cpp                               │
│                                                   │
│ --- a/src/agent.cpp                               │
│ +++ b/src/agent.cpp                               │
│ @@ -45,7 +45,8 @@                                 │
│ -    if (state == nullptr) return;                │
│ +    if (state == nullptr || !state->isValid())   │
│                                                   │
│ (scrollable diff)                                 │
└───────────────────────────────────────────────────┘
```

**State-dependent diff behavior:**

**When WorkingTree = Modified:**
- Diff shows: Working tree vs selected commit
- Command: `git diff <commit> -- <file>`
- Use case: "How do my WIP changes compare to commit X?"

**When WorkingTree = Clean:**
- Diff shows: Selected commit vs its parent
- Command: `git show <commit> -- <file>`
- Use case: "What did commit X change?"

**Navigation:**
- Tab: Cycle panes (Commits → Files → Diff)
- ↑↓: Scroll within active pane
- Enter: Cherry-pick confirmation (uses Dirty Operation Protocol if Modified)
- ESC: Return to main menu

**On Canon Branch:**
- Enter key disabled in commits/files panes
- Footer shows: "Cherry-pick not available on canon branch"

---

## 11. Time Travel Specification

### 11.1 Entering Time Travel

**ONLY available on working/other branches.**

From Commit History browser, press Enter on a commit.

**If WorkingTree = Modified:**
Show Dirty Operation Protocol dialog first.

**If WorkingTree = Clean:**
Show time travel confirmation:
```
⚠️ ENTERING TIME TRAVEL MODE

You are checking out commit abc123

You will be in DETACHED HEAD state.
Nothing is permanent until you attach HEAD to working branch.

[Continue] [Cancel]
```

Executes:
```bash
echo "abc123" > .git/TIT_TIME_TRAVEL
git checkout abc123
```

**New state:** Operation = TimeTraveling

### 11.2 While Time Traveling

**Status display:**
```
⚠️ DETACHED HEAD - TIME TRAVEL MODE

You: commit abc123 (3 commits behind working)
Working: commit xyz789 (dev)

Timeline: ●━━━━━━━◉━━━━━━━◉━━━━━━━◉
                You      ...    Working
```

**Available actions:**
- 🕒 Jump to different commit
- 👁️ View diff (vs working branch)
- ✅ Commit changes (creates orphaned commit)
- ⚠️ **Attach HEAD as new working** (destructive)
- ⬅️ Return to working branch

### 11.3 Attaching HEAD As New Working

**Requires typed confirmation: 'ATTACH'**

```
⚠️⚠️⚠️ DESTRUCTIVE ACTION ⚠️⚠️⚠️

You are about to make this commit the new working branch (dev).

What will happen:
✓ Your detached commit becomes working branch tip
✗ Later commits on dev will be ORPHANED

Type 'ATTACH' to confirm:
[_____________]

[Cancel]
```

Executes:
```bash
git checkout -B <working-branch>
rm .git/TIT_TIME_TRAVEL
```

**New state:** Operation = Normal, BranchContext = Working

### 11.4 Prevention on Canon Branch

**If user tries to time travel from canon branch history:**

```
❌ TIME TRAVEL NOT AVAILABLE

Time travel is only available on working branches.

Switch to working branch first:
[Switch to dev] [Cancel]
```

---

## 12. Branch Switching

**Available from any Normal state.**

### 12.1 Switch to Canon Branch

**Pre-condition:** Working tree must be clean

**If Modified:**
```
⚠️ Cannot switch branches

You have uncommitted changes.

Options:
[Commit changes] [Stash changes] [Cancel]
```

**If Clean:**
```bash
git checkout <canon-branch>
```

After switch:
- Menu regenerates with canon branch restrictions
- Footer shows: "On canon branch (read-only)"

### 12.2 Switch to Working Branch

**Pre-condition:** Working tree must be clean (same as above)

**If Clean:**
```bash
git checkout <working-branch>
```

After switch:
- Menu regenerates with full operations
- Footer shows: "On working branch (full operations)"

### 12.3 Keyboard Shortcut

**Proposed:** `b` key toggles between canon and working

---

## 13. First-Time Setup

### 13.1 Check Git Installation
```bash
git --version
```
If not found → Error

### 13.2 Check Git Configuration
```bash
git config user.name
git config user.email
```

If either empty → Setup wizard

### 13.3 Check Repository
```bash
git rev-parse --git-dir
```

If fails → Show init/clone options

**Init:** `git init`
- Prompt for canon branch name (default: "main")
- Prompt for working branch name (default: "dev")
- Create both branches
- Save to `~/.config/tit/repo.toml`

**Clone:** Prompt for URL, then `git clone <url> .`
- Detect remote branches
- Prompt for canon branch selection
- Prompt for working branch name (create locally)
- Save to config

### 13.4 Branch Configuration Storage

**File:** `~/.config/tit/repo.toml`

```toml
[repository]
initialized = true
repositoryPath = "/path/to/repo"
canonBranch = "main"
lastWorkingBranch = "dev"
```

### 13.5 Fatal Errors

**Detached HEAD (not from time travel):**
```
⚠️ Detached HEAD detected

Not a TIT time travel session.
Run: git checkout <branch>

[Exit TIT]
```

**Bare repository:**
```
⚠️ Bare repository detected

TIT requires a working tree.

[Exit TIT]
```

**Canon/Working branch not found:**
```
⚠️ Configuration Error

Canon branch 'main' not found.
Working branch 'dev' not found.

[Re-run init workflow] [Exit TIT]
```

---

## 14. UI Layout

```
┌────────────────────────────────────────┐
│  ⣿⣿⣿ TIT v2.0.0 ⣿⣿⣿               │ ← Banner
├────────────────────────────────────────┤
│  CWD: /path/to/repo                   │ ← Header
│  Branch: dev (working) | Clean | Sync │
├────────────────────────────────────────┤
│           Content Area (24 lines)      │
├────────────────────────────────────────┤
│  Description / Tips                    │ ← Footer
└────────────────────────────────────────┘
```

**Minimum terminal size:** 80×30 characters

**Header format:**
- Canon branch: `Branch: main (canon) | Clean | Ahead 2`
- Working branch: `Branch: dev (working) | Modified | Behind 1`
- Other branch: `Branch: feature-x (other) | Clean | Sync`

**All rendering via Lip Gloss:** borders, spacing, colors, alignment

---

## 15. Keyboard Shortcuts

### 15.1 Global Keys

**Ctrl+C (Exit)**
- First press: Show "Press Ctrl+C again to exit (3s timeout)"
- Second press: Exit TIT

**ESC (Context-dependent)**
- Text input (not empty): Confirm clear
- Text input (empty): Close input
- Console (running): Abort operation
- Console (done): Close console
- History browser: Return to menu
- Dirty operation: Abort and restore

**Tab (In browsers)**
- Cycle focus between panes

### 15.2 Branch Switching

**b key:** Toggle between canon and working branch (proposed)

### 15.3 Menu Navigation

**↑/k:** Move selection up
**↓/j:** Move selection down
**Enter:** Execute selected action
**Letter keys:** Jump to action (shortcuts)

---

## 16. Color Theme

**Theme file:** `~/.config/tit/themes/default.toml`

```toml
[colors]
status_clean = "#2ECC71"
status_modified = "#F39C12"
status_conflict = "#E74C3C"
timeline_sync = "#2ECC71"
timeline_ahead = "#3498DB"
timeline_diverged = "#E74C3C"
menu_selected = "#3498DB"
border = "#34495E"

# New: Branch context colors
branch_canon = "#27AE60"      # Green (read-only)
branch_working = "#3498DB"    # Blue (full ops)
branch_other = "#9B59B6"      # Purple (feature)
```

---

## 17. Design Invariants

1. **Menu = Contract:** If action appears, it must succeed
2. **State Machine:** UI is pure function of (Git state, Branch context)
3. **No Staging:** All changes commit together
4. **Dual Timeline:** Canon (clean) + Working (messy)
5. **Canon Read-Only:** No local commits on canon branch
6. **Safe Exploration:** Time travel non-destructive until attach (working only)
7. **Dirty Operations:** Always preservable with abort
8. **Beautiful:** Lip Gloss rendering, themed colors
9. **Guaranteed Success:** TIT never shows operations that could fail

---

## 18. Comparison: Old vs New TIT

| Feature | Old TIT (v1.0) | New TIT (v2.0) |
|---------|----------------|----------------|
| **Branch Model** | Single branch (main) | Dual branch (canon + working) |
| **Commit on Main** | ✅ Allowed | ❌ Not allowed (canon read-only) |
| **Time Travel** | On main branch | ONLY on working branches |
| **Attach HEAD** | "Attach as new main" | "Attach as new working" |
| **Workflow** | Work on main | Work on working → merge to canon |
| **State Axes** | 4 (WT, TL, Op, Remote) | 5 (WT, TL, Op, Remote, **BranchContext**) |
| **Menu Generation** | Operation-based | Operation + **Branch context** |
| **History Browser** | Full on main | Canon: read-only, Working: full |
| **Philosophy** | Single timeline | Canon (publish) + Working (develop) |

---

## 19. Workflow Examples

### 19.1 Standard Development Flow

```
1. Start on working branch (dev)
   Status: Branch: dev (working) | Clean | InSync

2. Make changes, commit
   Action: ✅ Commit changes
   Status: Branch: dev (working) | Clean | Ahead 1

3. Merge to canon
   Action: 🔀 Merge to canon branch
   → Switches to main, merges dev
   Status: Branch: main (canon) | Clean | Ahead 1

4. Push to remote
   Action: 📤 Push to remote
   Status: Branch: main (canon) | Clean | InSync

5. Switch back to working
   Action: 🔀 Switch to working branch
   Status: Branch: dev (working) | Clean | InSync
```

### 19.2 Pull Updates on Canon

```
1. On canon branch (main)
   Status: Branch: main (canon) | Clean | Behind 2

2. Pull from remote
   Action: 📥 Pull from remote
   → Fast-forward or merge
   Status: Branch: main (canon) | Clean | InSync

3. Switch to working branch
   Action: 🔀 Switch to working branch
   Status: Branch: dev (working) | Clean | Behind 2

4. Merge canon into working (optional)
   Action: (use git merge main, or pull if tracking)
```

### 19.3 Time Travel Exploration (Working Branch Only)

```
1. On working branch (dev)
   Status: Branch: dev (working) | Clean | InSync

2. Browse commit history
   Action: 📜 Browse commit history
   → Select old commit, press Enter

3. Enter time travel mode
   Status: Operation: TimeTraveling

4. Explore, make changes
   Action: ✅ Commit changes (orphaned commit)

5. Decide: Keep or discard
   Option A: ⚠️ Attach HEAD as new working (destructive)
   Option B: ⬅️ Return to working branch (discard)
```

### 19.4 Merge Conflict Resolution

```
1. On working branch (dev)
   Status: Branch: dev (working) | Clean | Ahead 1

2. Merge to canon
   Action: 🔀 Merge to canon branch
   → Conflict detected
   Status: Branch: main (canon) | Modified | Operation: Conflicted

3. Resolve conflicts
   Action: 🧩 View conflicted files
   → Edit files externally
   Action: ▶️ Continue operation

4. Merge completes
   Status: Branch: main (canon) | Clean | Ahead 1

5. Push to remote
   Action: 📤 Push to remote
```

---

## 20. Implementation Notes

### 20.1 BranchContext Detection

**In `DetectState()`:**
```go
state.IsCanonBranch = (state.CurrentBranch == state.CanonBranch)
state.IsWorkingBranch = (state.CurrentBranch == state.WorkingBranch)

if state.IsCanonBranch {
    state.BranchContext = BranchContextCanon
} else if state.IsWorkingBranch {
    state.BranchContext = BranchContextWorking
} else if state.CurrentBranch != "" {
    state.BranchContext = BranchContextOther
} else {
    state.BranchContext = BranchContextNone
}
```

### 20.2 Menu Generation Priority

```go
func GenerateMenu() []MenuItem {
    // Priority 1: Operation state
    switch gitState.Operation {
    case Conflicted: return menuConflicted()
    case Merging, Rebasing: return menuOperation()
    case DirtyOperation: return menuDirtyOperation()
    case TimeTraveling: return menuTimeTraveling()
    case Normal: break // Continue to branch context
    }

    // Priority 2: Branch context
    switch gitState.BranchContext {
    case Canon: return menuCanonBranch()
    case Working, Other: return menuWorkingBranch()
    case None: return menuNotRepo()
    }
}
```

### 20.3 Time Travel Prevention

**In `dispatchHistory()`:**
```go
if gitState.BranchContext == Canon {
    // Disable time travel on canon branch
    historyAllowTimeTravel = false
    footerHint = "Browse history (time travel not available on canon)"
} else {
    historyAllowTimeTravel = true
    footerHint = "Browse history (press Enter to time travel)"
}
```

### 20.4 State Machine Guarantees

**Every menu item maps to pre-validated state:**

| Menu Item | Pre-conditions Checked |
|-----------|------------------------|
| Commit | BranchContext != Canon, WorkingTree = Modified |
| Push | Remote = HasRemote, Timeline = Ahead |
| Merge to canon | BranchContext = Working/Other, WorkingTree = Clean |
| Switch branch | WorkingTree = Clean |
| Time travel | BranchContext != Canon, Operation = Normal |

**Result:** Menu only shows operations guaranteed to succeed.

---

**End of Specification**
