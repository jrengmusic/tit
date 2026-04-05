# TIT — Core Specification v2.0 (Single Active Branch)

**TIT:** Terminal UI for Git
**Target:** Go + Bubble Tea + Lip Gloss
**Philosophy:** Deterministic state machine. Single active branch. Zero surprises. Beautiful rendering.

---

## 1. Technology Stack

- **Language:** Go ≥ 1.21
- **Framework:** Bubble Tea (state machine) + Lip Gloss (rendering)
- **Git Interface:** `os/exec` only, no libraries
- **Output:** Single static binary

---

## 2. Foundational Principle

**TIT's UI is a pure function of Git state.**

```
Git State → Allowed Actions → Menu
```

If Git would reject an action, it must not appear in the menu.

**Key philosophy:**
- TIT operates on the **current active branch only**
- Users can switch branches anytime
- No branch tracking, no configuration files
- State always reflects actual Git state
- Destructive operations always require confirmation

---

## 3. State Model

Every decision in TIT derives from **five axes**:

### GitEnvironment — System prerequisites (Axis 0, checked first)
| Code | Meaning |
|------|---------|
| `Ready` | Git + SSH available |
| `NeedsSetup` | Git OK, SSH needs configuration |
| `MissingGit` | Git not installed |
| `MissingSSH` | SSH not installed |

**Note:** GitEnvironment is checked BEFORE all other state detection. If not Ready, TIT cannot proceed.

### WorkingTree — Local file changes (Axis 1)
| Code | Meaning |
|------|---------|
| `Clean` | No changes |
| `Dirty` | Has changes (staged, unstaged, or both) |

**Note:** TIT doesn't distinguish staging states. All changes commit together.

### Timeline — Local vs remote comparison
| Code | Meaning |
|------|---------|
| `InSync` | Local == Remote |
| `Ahead` | Local ahead (unpushed commits) |
| `Behind` | Local behind (unpulled commits) |
| `Diverged` | Both have unique commits |
| *(empty)* | **N/A** - No comparison possible (no remote OR detached HEAD) |

**Note:** Timeline is a comparison state. It only applies when:
- On a branch (not detached HEAD)
- Branch has remote tracking configured

When `Operation = TimeTraveling` or `Remote = NoRemote`, Timeline = empty (not applicable).

### Operation — Git operation state
| Code | Meaning |
|------|---------|
| `NotRepo` | Not in a git repository |
| `Normal` | No operation in progress |
| `Merging` | Merge in progress |
| `Conflicted` | Operation stopped due to conflicts |
| `Rebasing` | Rebase in progress (may have conflicts) |
| `TimeTraveling` | Detached HEAD (exploring history, read-only) |
| `DirtyOperation` | Executing dirty pull/merge with stashed work |

### Remote — Remote repository presence
| Code | Meaning |
|------|---------|
| `NoRemote` | No remote configured |
| `HasRemote` | Remote exists |

### IsTitTimeTravel — Distinguishes TIT time travel from manual detached HEAD
| Value | Meaning |
|-------|---------|
| `false` | Manual detached HEAD (user ran `git checkout <commit>` outside TIT) |
| `true` | TIT-initiated time travel (user pressed ENTER on commit in History mode) |

**Usage:**
- When `Operation = TimeTraveling`, `IsTitTimeTravel` determines menu behavior
- Both cases show Time Travel menu (Browse history, Return to branch)
- Return workflow differs:
  - `IsTitTimeTravel = true`: Uses TIT marker file and config for stash tracking
  - `IsTitTimeTravel = false`: Direct stash without TIT metadata

**State Tuple:** `(GitEnvironment, WorkingTree, Timeline, Operation, Remote)`

---

## 4. State Priority Rules

**Pre-Flight Check (Before Any State Detection): GitEnvironment**
- Check if git/SSH properly configured
- If `MissingGit` → Fatal error: Git not installed
- If `MissingSSH` → Fatal error: SSH not installed
- If `NeedsSetup` → Enter ModeSetupWizard (SSH key generation)
- If `Ready` → Proceed to git state detection

**Note:** GitEnvironment is a separate Application field checked BEFORE git state detection. It is NOT part of the git.State tuple.

**Priority 2: Mid-Operation State Recovery**
- `Conflicted` → Open conflict resolver (resolve or abort)
- `Merging` → Offer commit (finalize merge) or abort
- `Rebasing` → Open conflict resolver if conflicts, else offer continue or abort
- `DirtyOperation` → Resume dirty operation pipeline (TIT marker file `.git/TIT_DIRTY_OP` tracks state)
- These are recoverable states. TIT has the machinery to handle all of them.

**Priority 3: Operation State**
- `Normal` → Proceed to check other axes
- `TimeTraveling` → Time travel menu (Browse history, Merge back, Return)
  - **Entered via:**
    - History mode → select commit → ENTER (TIT-initiated)
    - User ran `git checkout <commit>` outside TIT (manual detached)
  - **NOT a standalone menu item** - accessed through History or detected on startup

**Priority 4: Remote Presence**
- `Remote = NoRemote` → Hide sync actions, show "Add remote"
- `Remote = HasRemote` → Enable sync menus based on Timeline

**Priority 5: WorkingTree + Timeline**
- Determines which action menus appear

---

## 5. Mid-Operation Recovery (Startup and Runtime)

TIT handles all mid-operation states — both pre-existing (detected at startup) and runtime (produced by TIT operations).

### Conflicted
- Open conflict resolver UI
- User picks file versions, stages resolved files
- Finalize: `git commit` (merge) or `git rebase --continue` (rebase) depending on underlying operation
- Abort: `git merge --abort` or `git rebase --abort`

### Merging (no conflicts)
- Merge is in progress but no conflict markers exist
- Offer: Finalize merge (commit) or abort merge
- Menu: "Finalize merge" / "Abort merge"

### Rebasing
- If conflicts exist → Open conflict resolver
- After resolution: `git rebase --continue`
- If next commit also conflicts → loop back to conflict resolver
- Loop continues until all commits replayed or user aborts
- Abort at any step: `git rebase --abort` (restores pre-rebase state)
- Menu: "Continue rebase" (if no conflicts) / "Abort rebase"

### DirtyOperation
- TIT-initiated dirty pull/merge/switch was interrupted
- `.git/TIT_DIRTY_OP` marker file tracks original branch, HEAD, and operation type
- Resume: pick up from the phase recorded in the marker
- Abort: restore original state from marker data

### State Transitions
```
Conflicted ──→ Conflict Resolver ──→ Finalize/Abort ──→ Normal
Merging    ──→ Commit/Abort menu ──→ Normal
Rebasing   ──→ Conflict Resolver (loop) ──→ Continue/Abort ──→ Normal
DirtyOp    ──→ Resume pipeline ──→ Normal
```

---

## 6. State → Menu Mapping (Normal Operation Only)

### When Operation = NotRepo

**Purpose:** User is not in a git repository. Show initialization options.

**Menu items:**
- ✅ Initialize repository (triggers location choice: Here or Subdirectory)
- 📥 Clone repository (triggers location choice: Here or Subdirectory)

**Location Choice (ModeInitializeLocation):**
When "Initialize repository" is selected, TIT prompts:
1. 📍 Initialize directory (current working directory)
2. 📁 Create subdirectory (prompts for name, then initializes there)

**Clone Choice (ModeCloneLocation):**
When "Clone repository" is selected, TIT prompts:
1. 📍 Clone to directory (clones into current working directory)
2. 📁 Create subdirectory (clones into new subdirectory)

### When Operation = TimeTraveling

**Accessed from:**
- **TIT time travel:** History mode (select commit → ENTER)
- **Manual detached:** User ran `git checkout <commit>` outside TIT

**Show ONLY:**
- 🕒 Browse history (view other commits while time traveling)
- 📦 Merge changes back to [branch]
- ⬅️ Return to [branch] (discard changes)

**Note:** Both TIT time travel and manual detached HEAD show the same menu items. The difference is in the return workflow:
- **TIT time travel:** Uses TIT marker file and config stash tracking
- **Manual detached:** Direct stash → checkout → merge → apply (no TIT metadata)

**How to enter time travel (TIT-initiated):**
1. From main menu: Select "Commit history"
2. Navigate to desired commit
3. Press ENTER to confirm time travel
4. If working tree is dirty → Dirty operation protocol (stash changes)
5. On confirm → Detached HEAD at selected commit (Operation = TimeTraveling)

**Manual detached entry:**
- User ran `git checkout <commit>` outside TIT
- TIT detects this on startup (no `.git/TIT_TIME_TRAVEL` marker file)
- Offers same Time Travel menu with return option

**While time traveling:**
- **Read-only exploration:** View code at any point in history
- **Local changes allowed:** Can make modifications (tracked as WorkingTree = Dirty)
- **Cannot commit:** Changes must be merged back to original branch
- **Can browse:** Use History mode to jump to different commits
- **Can view diffs:** Via File History mode (shows commit vs parent or vs working tree)

**To exit time travel:**
- **Merge back:** Keep local changes, merge to original branch
- **Return:** Discard changes, return to original branch

### When Operation = Normal

#### Working Tree Actions
| State | Menu Items |
|-------|------------|
| `Clean` | *(no working tree actions)* |
| `Dirty` | ✅ Commit changes<br>🚀 Commit and push<br>💥 Discard all changes |

#### Timeline Sync Actions

**When Remote = NoRemote:**
- 🌐 Add remote

**When Remote = HasRemote and Timeline = InSync:**
- 📥 Pull from remote (refresh)

**When Timeline = Ahead:**
- 📤 Push to remote
- ⚠️ Force push (overwrite remote)

**When Timeline = Behind:**
- 📥 Pull (merge)
- ⚠️ Replace local with remote (discard local commits)

**When Timeline = Diverged:**
- 🔀 Sync: Merge remote into local
- ⬇️ Keep remote (discard local commits)
- ⬆️ Keep local (overwrite remote)

#### Branch Operations (always available)
- 🔀 Switch branch (shows list of local branches)
- ➕ Create new branch
- 🔗 Merge another branch into current

#### History Actions (always available)
- 🕒 Browse commit history (optional: time travel to old commit)
- 📁 Browse file history (view file changes over time)

---

## 7. Dirty Operation Protocol

**Purpose:** Apply any change-set (pull, merge, time travel) while preserving uncommitted work.

**Used when:**
- Pull (merge) with WorkingTree = Dirty
- Merge with WorkingTree = Dirty
- Time travel with WorkingTree = Dirty

**User sees:**
```
⚠️ You have uncommitted changes

To proceed, your changes will be temporarily saved (stashed).
After the operation completes, they'll be reapplied.

This may cause conflicts if the operation changes the same files.

[Save changes and proceed] [Discard changes and proceed] [Cancel]
```

**Implementation steps:**

1. **Snapshot**
   ```bash
   ORIGINAL_BRANCH=$(git symbolic-ref --short HEAD)
   ORIGINAL_HEAD=$(git rev-parse HEAD)
   git stash push -u -m "TIT DIRTY-OP SNAPSHOT"
   echo "$ORIGINAL_BRANCH" > .git/TIT_DIRTY_OP
   echo "$ORIGINAL_HEAD" >> .git/TIT_DIRTY_OP
   ```

2. **Apply change-set** (pull, merge, checkout, etc.)
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
ORIGINAL_BRANCH=$(head -n 1 .git/TIT_DIRTY_OP)
ORIGINAL_HEAD=$(tail -n 1 .git/TIT_DIRTY_OP)
git checkout $ORIGINAL_BRANCH
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

## 8. Branch Switching

**Available from Normal state.**

### Switch to Existing Branch

**Pre-condition:** Working tree must be clean

**If WorkingTree = Modified:**
```
⚠️ Cannot switch branches

You have uncommitted changes.

Options:
[Commit changes] [Stash changes] [Cancel]
```

**If WorkingTree = Clean:**
- Show list of all local branches
- Highlight current branch
- User selects target branch
- Execute: `git checkout <branch>`

After switch:
- Menu regenerates based on new branch state
- Header shows new branch name

### Create New Branch

**Pre-condition:** Working tree must be clean

**Prompt:**
```
Create new branch

New branch will be created from current HEAD.

Branch name: [_____________]

[Create and switch] [Cancel]
```

**Validates:**
- Branch name not empty
- Branch name doesn't already exist
- Valid git ref name

**Execute:**
```bash
git checkout -b <new-branch>
```

---

## 9. Merge Branch Assistance

**Purpose:** Merge another branch into current branch with safety and clarity.

**Pre-conditions:**
- Operation = Normal
- Working tree must be clean (uses dirty protocol if Modified)

**User sees:**
```
🔀 MERGE BRANCH

Select branch to merge into current branch (main):

  > dev
    feature/auth
    experimental

This will merge the selected branch into main.
Conflicts will be handled if they occur.

[Select] [Cancel]
```

**After selection:**
```
Merge: dev → main

This will:
✓ Merge dev into main
✓ Handle conflicts if any
✓ Keep both branches intact

[Proceed] [Cancel]
```

**Implementation:**
```bash
git merge <selected-branch>
```
- If conflicts → Operation = Conflicted
- User resolves → Continue merge
- If success → Back to Normal state

**Abort (if conflicts):**
```bash
git merge --abort
```
→ Current branch unchanged

---

## 10. Time Travel Specification

### 9.1 Entering Time Travel

**Available from:** Commit History browser, press Enter on a commit.

**If WorkingTree = Modified:**
- Show Dirty Operation Protocol dialog first
- Stash changes before entering time travel

**If WorkingTree = Clean:**
- Show time travel confirmation:

```
⚠️ ENTERING TIME TRAVEL MODE

You are about to view commit abc123 from the past.

This is READ-ONLY exploration:
✓ You can view code at this point in history
✓ You can build and test this old version
✓ You can make local changes (not committed)

You CANNOT commit while exploring the past.

To keep changes, merge them back to your branch.

[Continue] [Cancel]
```

**Executes:**
```bash
ORIGINAL_BRANCH=$(git symbolic-ref --short HEAD)
echo "$ORIGINAL_BRANCH" > .git/TIT_TIME_TRAVEL
git checkout <commit-hash>
```

**New state:** Operation = TimeTraveling

### 9.2 While Time Traveling

**Status display:**
```
⚠️ TIME TRAVEL MODE (Read-only)

Viewing: commit abc123 (3 commits behind main)
Your branch: main (commit xyz789)

Timeline: ●━━━━━━━◉━━━━━━━◉━━━━━━━◉
                You      ...      main
```

**Available actions:**
- 🕒 Jump to different commit
- 📄 File(s) history (view file changes and diffs)
- 📦 Merge changes back to main
- 🔙 Return to main (discard changes)

**Behavior:**
- Working tree changes allowed (tracked as Modified)
- CANNOT commit (no menu option for commit)
- Can build, test, experiment freely
- Changes stay local until merge-back or discard

### 9.3 Merge Changes Back to Branch

**Purpose:** Keep changes made during time travel by merging them into original branch.

**Pre-conditions:**
- Currently in time travel mode
- May have Modified working tree OR Clean

**User sees:**
```
📦 MERGE TIME TRAVEL CHANGES

Merge changes from detached HEAD back to main.

Current state:
- Viewing: commit abc123
- Your branch: main
- Working tree: Modified (you have local changes)

This will:
1. Save your current changes (if any)
2. Return to main branch
3. Merge this commit + your changes into main
4. Handle conflicts if they occur

[Merge back to main] [Cancel]
```

**Implementation (using dirty op pattern):**

1. **If WorkingTree = Modified:**
   ```bash
   git stash push -u -m "TIT TIME-TRAVEL WIP"
   ```

2. **Save detached HEAD commit:**
   ```bash
   DETACHED_COMMIT=$(git rev-parse HEAD)
   ```

3. **Checkout original branch:**
   ```bash
   ORIGINAL_BRANCH=$(cat .git/TIT_TIME_TRAVEL)
   git checkout $ORIGINAL_BRANCH
   ```

4. **Merge detached commit:**
   ```bash
   git merge $DETACHED_COMMIT
   ```
   - If conflicts → Operation = Conflicted
   - User resolves → Continue

5. **If stash exists, apply back:**
   ```bash
   git stash apply
   ```
   - If conflicts → Operation = Conflicted
   - User resolves → Continue

6. **Cleanup:**
   ```bash
   git stash drop
   rm .git/TIT_TIME_TRAVEL
   ```

**Abort (ESC at any step):**
```bash
ORIGINAL_BRANCH=$(cat .git/TIT_TIME_TRAVEL)
ORIGINAL_HEAD=$(tail -n 1 .git/TIT_DIRTY_OP)
git checkout $ORIGINAL_BRANCH
git reset --hard $ORIGINAL_HEAD
git stash apply  # If stash exists
git stash drop
rm .git/TIT_TIME_TRAVEL
```

**New state:** Operation = Normal (back on original branch)

### 9.4 Return to Branch (Discard Changes)

**Simple return:**
```bash
ORIGINAL_BRANCH=$(cat .git/TIT_TIME_TRAVEL)
git checkout $ORIGINAL_BRANCH
rm .git/TIT_TIME_TRAVEL
```

**If WorkingTree = Modified:**
```
⚠️ Discard changes?

You have uncommitted changes in time travel mode.
Returning to main will DISCARD these changes.

[Discard and return] [Cancel]
```

---

## 11. Commit History Browser (2-Column)

**Purpose:** Browse commit timeline, optionally time travel to old commits.

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
- Tab: Switch between Commits list and Details pane
- **ENTER:** Enter time travel mode at selected commit (read-only exploration)
- **Ctrl+R:** REWIND to selected commit (destructive, reset --hard)
- ESC: Return to main menu

**REWIND (Ctrl+R) Feature:**

Purpose: Permanently reset the current branch to any historical commit.

Availability: Always available from commit history browser, regardless of current Operation state.

Behavior:
- Discards all commits after selected commit
- Discards all uncommitted changes
- Requires destructive confirmation

Confirmation Dialog:
```
⚠️ Destructive Operation

This will discard all commits after [HASH]. 
Any uncommitted changes will be lost.

[Rewind] [Cancel]
```

Implementation: `git reset --hard <commit>`

**Time travel flow (on ENTER):**
1. Show confirmation dialog explaining read-only nature
2. If working tree dirty → Show dirty operation protocol (stash changes)
3. On confirm → Checkout commit (dirty tree already stashed)
4. Operation state changes to `TimeTraveling`
5. Menu now shows time travel options (Browse history, Merge back, Return)

---

## 12. File(s) History Browser (3-Pane)

**Purpose:** Browse file changes over time.

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

**When WorkingTree = Dirty:**
- Diff shows: Working tree vs selected commit
- Command: `git diff <commit> -- <file>`
- Use case: "How do my current changes compare to commit X?"

**When WorkingTree = Clean:**
- Diff shows: Selected commit vs its parent
- Command: `git show <commit> -- <file>`
- Use case: "What did commit X change?"

**Navigation:**
- Tab: Cycle panes (Commits → Files → Diff)
- ↑↓: Scroll within active pane
- ESC: Return to main menu

**Note:** Cherry-pick not implemented (not wired to interface).

---

## 13. First-Time Setup

### 13.1 Check GitEnvironment (Priority 0 - Before Any Git Operations)

TIT checks machine readiness BEFORE git state detection:

**Check sequence:**
```bash
1. Check git installed: git --version
   If not found → GitEnvironmentMissingGit → Fatal error

2. Check SSH installed: ssh -V
   If not found → GitEnvironmentMissingSSH → Fatal error

3. Check SSH keys configured:
   - Scans ~/.ssh directory for private keys
   - Looks for: id_rsa, id_ed25519, *_rsa, *_ed25519 patterns
   - If no keys found → GitEnvironmentNeedsSetup → Enter ModeSetupWizard

4. If all checks pass → GitEnvironmentReady → Proceed to git state detection
```

**ModeSetupWizard:** Guided SSH key generation and configuration
- User prompted for email (key comment)
- SSH key generated: `ssh-keygen -t ed25519 -C "<email>"`
- SSH agent started and key added
- Public key displayed for user to add to GitHub/GitLab/Gitea
- Wizard exits on completion, TIT proceeds to normal startup

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

**Init:**
```bash
git init
git checkout -b main
```

**Clone:**
- Prompt for URL
- Clone to current directory or subdirectory
- Detect default branch
- Checkout default branch

### 13.4 Branch Mismatch on Remote Add

When adding first remote, if local branch ≠ remote default:

```
⚠️ Branch name mismatch

Your local branch: master
Remote default branch: main

This may cause confusion. Would you like to switch?

[Switch to main] [Stay on master]
```

**Switch:**
```bash
git checkout -b main origin/main
```

**Stay:**
- Remote added normally
- User stays on current branch

### 13.5 Fatal Errors

**NOTE:** Manual detached HEAD is NO LONGER a fatal error. TIT now supports returning from manual detached HEAD (see Section 13.6).

**Bare repository:**
```
⚠️ Bare repository detected

TIT requires a working tree.

[Exit TIT]
```

---

## 13.6 Manual Detached HEAD Support

TIT supports returning from manual detached HEAD state (user ran `git checkout <commit>` outside TIT).

### Detection

When `Operation = TimeTraveling` AND `IsTitTimeTravel = false`:
- No `.git/TIT_TIME_TRAVEL` marker file exists
- User manually checked out a commit
- TIT offers "Return to branch" option in menu

### Menu Options

| Menu Item | Description |
|-----------|-------------|
| **Browse history** | Jump to other commits (same as TIT time travel) |
| **Return to branch** | Checkout target branch |

### Return to Branch Flow

1. **Select "Return to branch"**
2. **Branch picker appears** (shows all local branches)
3. **Select target branch**

#### If working tree is clean:
- Direct checkout to target branch
- Operation changes to `Normal`

#### If working tree is dirty:
Confirmation dialog:
```
Return to <branch> with uncommitted changes

You have changes during time travel. Choose action:
(Press ESC to cancel)

[Stash changes]  [Discard changes]
```

**Stash changes flow:**
1. Stash uncommitted changes (`git stash push -u`)
2. Checkout target branch
3. Merge detached commit into target branch
4. Apply stash back (may conflict - see conflict resolution below)
5. Drop stash

**Discard changes flow:**
1. Checkout target branch (changes discarded)
2. Operation changes to `Normal`

### Conflict Resolution

Two conflict scenarios:

1. **Merge conflict:** Detached commit modifies same lines as target branch
   - Enter conflict resolver
   - User resolves, then stash apply happens

2. **Stash apply conflict:** Stashed changes conflict with merged result
   - Enter conflict resolver  
   - User resolves to complete return

Both scenarios use the same conflict resolution UI as time travel return.

### State Transitions

```
Manual Detached (clean)
    ├─ Return to branch → Normal
    └─ Browse history → Manual Detached (different commit)

Manual Detached (dirty)
    ├─ Return to branch → Stash/Discard dialog
    │    ├─ Stash → Merge → Apply Stash → Normal (or Conflict)
    │    └─ Discard → Normal
    └─ Browse history → Manual Detached (different commit)
```

---

## 14. UI Layout

```
┌────────────────────────────────────────┐
│  ⣿⣿⣿ TIT v2.0.0 ⣿⣿⣿               │ ← Banner
├────────────────────────────────────────┤
│  CWD: /path/to/repo                   │ ← Header
│  Branch: main | Clean | In sync       │
├────────────────────────────────────────┤
│           Content Area (24 lines)      │
├────────────────────────────────────────┤
│  Description / Tips                    │ ← Footer
└────────────────────────────────────────┘
```

**Minimum terminal size:** 80×30 characters

**Header format:**
- Shows current branch name
- Working tree status (Clean/Modified)
- Timeline status (In sync/Ahead/Behind/Diverged)

**All rendering via Lip Gloss:** borders, spacing, colors, alignment

---

## 15. Keyboard Shortcuts

### 14.1 Global Keys

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

### 14.2 Menu Navigation

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
```

---

## 17. Design Invariants

1. **Menu = Contract:** If action appears in menu, it must succeed. Never show operations that could fail or leave dangling state.
2. **State Machine:** UI is pure function of Git state
3. **No Staging:** All changes commit together
4. **Single Active Branch:** TIT operates on current branch only
5. **Branch Switching:** Users can switch branches anytime (when clean)
6. **Safe Exploration:** Time travel is read-only until merge-back
7. **Dirty Operations:** Automatically managed by TIT. No manual abort in menu—git state safe even if user exits.
8. **Beautiful:** Lip Gloss rendering, themed colors
9. **Guaranteed Success:** TIT never shows operations that could fail
10. **No Configuration:** State always reflects actual Git state
11. **No Dangling States:** Merge/Rebase/DirtyOperation has no abort menu option. User must complete or exit TIT (git state preserved).
12. **Destructive Clarity:** Ctrl modifier signals destructive operations (Ctrl+R = REWIND). ENTER always safe (time travel, read-only).

---

## 18. Implementation Plan

See `IMPLEMENTATION_PLAN.md` for step-by-step porting strategy from old TIT to new TIT.

---

**End of Specification**

## Commit History Browser — Rewind Option

**When in commit history browser (ModeHistoryBrowser):**

- **ENTER:** Enter time travel mode at selected commit (read-only exploration)
- **Ctrl+R:** REWIND to selected commit (destructive, reset --hard)

**REWIND (Ctrl+R) behavior:**
- Available on ANY commit, regardless of current Operation state
- Shows confirmation: "This will discard all commits after [HASH]. Any uncommitted changes will be lost."
- Executes `git reset --hard <commit>`
- Discards ALL commits after selected commit
- Discards ALL uncommitted changes
- Returns to main menu on success

**Example:**
```
Commits on main:
  abc1234 Feature X (current HEAD)
  def5678 Fix bug Y
  ghi9012 Initial commit ← User selects, presses Ctrl+R

Result: Branch resets to ghi9012, abc1234 and def5678 discarded
```

**Why Ctrl+R:**
- R for "Rewind" (semantic clarity)
- Distinguishes destructive reset from read-only time travel
- Ctrl modifier indicates dangerous operation
- ENTER remains bound to time travel (safe, reversible)
