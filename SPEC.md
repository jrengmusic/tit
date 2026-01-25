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

Every decision in TIT derives from **four axes**:

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
| `TimeTraveling` | Detached HEAD (exploring history, read-only) |
| `DirtyOperation` | Executing dirty pull/merge with stashed work |

### Remote — Remote repository presence
| Code | Meaning |
|------|---------|
| `NoRemote` | No remote configured |
| `HasRemote` | Remote exists |

**State Tuple:** `(WorkingTree, Timeline, Operation, Remote)`

---

## 4. State Priority Rules

**Pre-Flight Check (Before Any State Detection): GitEnvironment**
- Check if git/SSH properly configured
- If `MissingGit` → Fatal error: Git not installed
- If `MissingSSH` → Fatal error: SSH not installed
- If `NeedsSetup` → Enter ModeSetupWizard (SSH key generation)
- If `Ready` → Proceed to git state detection

**Note:** GitEnvironment is a separate Application field checked BEFORE git state detection. It is NOT part of the git.State tuple.

**Priority 2: Pre-Flight Checks (Git Repository State)**
- If `Conflicted` OR `Merging` OR `Rebasing` OR `DirtyOperation` → Show error screen, prevent TIT startup
- These are pre-existing abnormal states. User must resolve externally.

**Priority 3: Operation State** (For Valid Startups)
- `Normal` → Proceed to check other axes
- `TimeTraveling` → Time travel menu (Browse history, Merge back, Return)
  - **Entered via:** History mode → select commit → ENTER
  - **NOT a standalone menu item** - accessed only through History

**Priority 4: Remote Presence**
- `Remote = NoRemote` → Hide sync actions, show "Add remote"
- `Remote = HasRemote` → Enable sync menus based on Timeline

**Priority 5: WorkingTree + Timeline**
- Determines which action menus appear

---

## 5. Pre-Flight Checks (Startup)

Before showing any menu, TIT checks if git repository is in a mid-operation state:

```
Conflicted ──┐
Merging    ──┼──→ ⚠️ ERROR: Repository in mid-operation
Rebasing   ──┤
DirtyOperation ┘
```

**If detected:**
```
⚠️ Git repository is in the middle of a [operation] operation

Your repository cannot be managed by TIT while an operation is in progress.
Please complete or abort the operation using standard git commands:

  git merge --continue  (after resolving conflicts)
  git merge --abort     (to discard the merge)
  git rebase --continue / --abort
  git stash pop         (for stash operations)

[Exit TIT]
```

**Why:** These are pre-existing abnormal conditions, not states TIT manages. TIT operates on clean/normal repositories only.

---

## 6. State → Menu Mapping (Normal Operation Only)

### When Operation = NotRepo

**Purpose:** User is not in a git repository. Show initialization options.

**Smart location dispatch:**
- **If CWD is empty** → Show two options:
   - 🔨 Initialize here
   - 📥 Clone repository
- **If CWD not empty** → Skip menu, directly dispatch to:
   - 📥 Clone as subdirectory (only option for init/clone)

**Why:** Can't init in non-empty directory. No single-option menus.

**Menu items (CWD empty):**
- ✅ Initialize repository (CWD must be empty)
- 📥 Clone repository

### When Operation = TimeTraveling

**Accessed from:** History mode (select commit → ENTER)

**Show ONLY:**
- 🕒 Browse history (view other commits while time traveling)
- 📦 Merge changes back to [branch]
- ⬅️ Return to [branch] (discard changes)

**How to enter time travel:**
1. From main menu: Select "Commit history"
2. Navigate to desired commit
3. Press ENTER to confirm time travel
4. If working tree is dirty → Dirty operation protocol (stash changes)
5. On confirm → Detached HEAD at selected commit (Operation = TimeTraveling)

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
| `Dirty` | ✅ Commit changes<br>🚀 Commit and push |

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

**Detached HEAD (not from time travel):**
```
⚠️ Detached HEAD detected

You are not on a branch.
TIT requires you to be on a branch.

Please checkout a branch:
git checkout main

[Exit TIT]
```

**Bare repository:**
```
⚠️ Bare repository detected

TIT requires a working tree.

[Exit TIT]
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
