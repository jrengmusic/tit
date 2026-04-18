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
| `Rewinding` | Hard reset in progress (transient) |

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
- 📄 File(s) history (view file changes and diffs)
- 🔙 Return to [branch]

**"Return to [branch]" dual path:**
- **TIT-initiated time travel** (`.git/TIT_TIME_TRAVEL` marker present): opens confirmation dialog with Yes=merge-back / Discard=return-no-merge / Cancel. Target branch is read from the marker.
- **Manual detached HEAD** (no marker, multi-branch repo): opens **Branch Picker** — user selects target branch; then if WorkingTree = Dirty, prompts Stash/Discard before checkout. See §8 for picker behavior.
- **Manual detached HEAD, single-branch repo**: auto-selects the sole branch, proceeds as above.

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

Menu composition in order: **Working Tree → Timeline → separator (if any items above) → History → separator + Add Remote (if NoRemote)**.

#### Working Tree Actions
| State | Menu Items |
|-------|------------|
| `Clean` | *(no working tree actions — section hidden)* |
| `Dirty` | 📝 Commit changes<br>🚀 Commit and push *(only if HasRemote)*<br>💥 Discard all changes (ctrl+r) |

#### Timeline Sync Actions

Timeline section is hidden entirely when `Remote = NoRemote` (timeline is N/A without a remote).

**Timeline = InSync:** *(no items — repo is in sync, no sync action needed)*

**Timeline = Ahead:**
| WorkingTree | Menu Items |
|-------------|------------|
| `Clean` | 📤 Push to remote<br>💥 Force push (shift+]) |
| `Dirty`  | *(no items — must commit first)* |

**Timeline = Behind:**
| WorkingTree | Menu Items |
|-------------|------------|
| `Clean` | 📥 Pull (fetch + merge)<br>💥 Replace local (discard local commits) |
| `Dirty`  | 🔺 Pull (save changes) *(dirty-op protocol)*<br>💥 Replace local |

**Timeline = Diverged:**
| WorkingTree | Menu Items |
|-------------|------------|
| `Clean` | 📤 Push (auto sync)<br>📥 Pull (merge diverged)<br>💥 Force push (shift+])<br>💥 Replace local |
| `Dirty`  | 🔺 Pull (save changes) *(dirty-op protocol)*<br>💥 Force push (shift+])<br>💥 Replace local |

#### History Actions (always shown in Normal)
- 🕒 History
- 📄 File(s) history

#### Remote Setup (only when `Remote = NoRemote`)
- 🌐 Add remote

#### Branch Management / Config access

Branch operations (switch / add / delete / merge-from) are NOT direct menu items — they live inside the **Branch Picker**, reached from the **Config menu**. Press `/` in Main Menu to enter Config mode. See §8 (Branch Picker) and §6.3 (Config Menu).

### When Mode = Config

Entered via `/` keybinding from Main Menu. Exit via `Esc` or "Back" item.

**Items (order + conditional logic):**

| Item | Label | Shortcut | Condition |
|------|-------|----------|-----------|
| `config_add_remote` | 🔗 Add Remote | `r` | `Remote = NoRemote` |
| `config_switch_remote` | 🔗 Switch Remote | `s` | `Remote = HasRemote` |
| `config_remove_remote` | 🗑️ Remove Remote | `r` | `Remote = HasRemote` |
| *(separator)* | | | |
| `config_branch` | 🌿 Branch | `b` | always |
| `config_preferences` | ⚙️ Preferences | `p` | always |
| *(separator)* | | | |
| `config_back` | 🔙 Back | *(none)* | always |

- `config_add_remote` / `config_switch_remote` are **mutually exclusive** — only one shows based on Remote axis.
- `config_remove_remote` shows ONLY when `HasRemote`.
- `config_branch` opens the **Branch Picker** (§8).
- `config_preferences` enters **ModePreferences**.

### When Mode = Preferences

Navigation-only menu (no shortcut letters — `↑/↓` + `Enter`).

**Items:**

| Item | Label | Action |
|------|-------|--------|
| `preferences_auto_update` | 🔄 Auto-update | Toggle background state refresh ON/OFF |
| `preferences_interval` | ⏱️ Update Interval | Adjust via `+`/`-` keys (±1 min) or `Shift +/-` (±10 min) |
| `preferences_theme` | 🎨 Theme | Cycle through discovered themes in `~/.config/tit/themes/*.xml` |

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

## 8. Branch Picker

**Single surface for all branch management.** Add / Switch / Merge-from / Delete — all four CRUD actions live in one picker view. No separate menus for create / switch / merge.

### Entry points

- **From Config menu:** Select 🌿 **Branch** item → enters picker in default purpose (switch/manage).
- **From TimeTraveling "Return to [branch]"** on manual detached HEAD with multiple branches → enters picker with `IsReturnToBranch=true`.
- **From post-create completion** (`BranchPickerReturnAfterCreate`): after a picker-`a`-initiated branch creation finishes, returns to picker with cursor on the newly created branch.

### Layout (2-pane split)

```
┌────────────────────────┬────────────────────────┐
│ Branches               │ Details                │
│                        │                        │
│ > ● main (current)     │ main ↑0 ↓0 synced      │
│   dev                  │ Last commit: abc123    │
│   feature/auth         │ Author: …              │
│   experimental         │ Date:   …              │
│                        │ Message: …             │
└────────────────────────┴────────────────────────┘
```

Left pane: branch list. Current branch marked `●`, bold. Right pane: details for the cursor branch (hash, author, date, message, tracking status: `↑N ↓N` / `synced` / `local`).

### Keybindings

| Key | Action | Visibility |
|-----|--------|-----------|
| `↑` / `k` | Move cursor up | always |
| `↓` / `j` | Move cursor down | always |
| `Tab` | Toggle focus between list pane and details pane | always |
| `Enter` | Primary action (context-dependent — see below) | always |
| `a` | **Add branch** — open new-branch input, create from current HEAD, switch to it, return to picker with cursor on new branch | always |
| `m` | **Merge selected branch into current** — runs merge flow; dirty-op protocol if WorkingTree = Dirty | hidden on current-branch row |
| `x` | **Delete selected branch** — confirm dialog → `git branch -D <branch>` → refresh picker | hidden on current-branch row |
| `Esc` | Return to previous mode (Menu or Config) | always |

### Enter behavior (context-dependent)

| Context | Cursor on current | Cursor on other branch |
|---------|-------------------|------------------------|
| **Default** (from Config → Branch) | Return to Config menu | **Switch branch.** If Dirty → dirty-switch dialog (Stash/Discard/Cancel). Else direct checkout. |
| **Merge purpose** (`BranchPickerPurpose="merge"`) | Return to Config | Merge selected into current (dirty-op if Dirty). |
| **Return-to-branch** (`IsReturnToBranch=true`, manual detached) | *(current is detached, no such row)* | Exit manual detached — checkout target branch. If Dirty → Stash/Discard dialog first. |

### Footer hints

Two state-driven variants via `FooterHintShortcuts` SSOT:
- **`branch_picker_current`** — cursor on current branch: only navigation/ESC hints (no `a`/`m`/`x`).
- **`branch_picker_other`** — cursor on other branch: full navigation + `a` add / `m` merge / `x` delete / `Enter` switch.

### Implementation

- `git branch -D <name>` for delete (force — Go uses `-D` not `-d`).
- `git checkout -b <new>` for add.
- `git merge <selected>` for merge-from; follows dirty-op protocol when dirty.
- `git checkout <name>` for switch.

---

## 9. Merge Branch

**Not a standalone flow.** Merge-from-another-branch is the `m` keybinding of the Branch Picker (§8). When the merged branch produces conflicts, Operation transitions to `Conflicted` and the conflict resolver (§5) takes over. Abort via `git merge --abort` returns to Normal with no change.

When WorkingTree = Dirty at the time of picker `m`: the dirty operation protocol (§7) runs — snapshot → merge → reapply → finalize — preserving the uncommitted work across the merge.

---

## 10. Time Travel Specification

### 10.1 Entering Time Travel

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

### 10.2 While Time Traveling

**Status display:**
```
⚠️ TIME TRAVEL MODE (Read-only)

Viewing: commit abc123 (3 commits behind main)
Your branch: main (commit xyz789)

Timeline: ●━━━━━━━◉━━━━━━━◉━━━━━━━◉
                You      ...      main
```

**Available actions:**
- 🕒 History (jump to different commit)
- 📄 File(s) history (view file changes and diffs)
- 🔙 Return to [branch] — dual path per §6 "Return to [branch] dual path" (dialog when TIT-marker present, branch picker when manual detached multi-branch)

**Behavior:**
- Working tree changes allowed (tracked as Modified)
- CANNOT commit (no menu option for commit)
- Can build, test, experiment freely
- Changes stay local until merge-back or discard

### 10.3 Merge Changes Back to Branch (Return → Yes path)

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

### 10.4 Return to Branch (Discard Changes path)

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

**ModeSetupWizard:** Guided SSH key generation and configuration.

**Seven phases (`SetupWizardStep`):**

| Phase | Purpose |
|-------|---------|
| `SetupStepWelcome` | Intro message — explain what the wizard will do |
| `SetupStepPrerequisites` | Verify git + ssh installed; abort on `MissingGit` / `MissingSSH` |
| `SetupStepEmail` | Text input for SSH key comment (user's email) |
| `SetupStepGenerate` | Run `ssh-keygen -t ed25519 -C "<email>"`, start `ssh-agent`, add key |
| `SetupStepDisplayKey` | Show public key + provider registration URLs (GitHub / GitLab / Gitea) |
| `SetupStepComplete` | Success message — user acknowledges, GitEnvironment transitions to `Ready` |
| `SetupStepError` | Display error, allow retry or exit |

Wizard exits on `SetupStepComplete` acknowledgement, TIT proceeds to git state detection.

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
│  ⣿⣿⣿ TIT vX.Y.Z ⣿⣿⣿               │ ← Banner
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
**Letter keys:** Jump to action (shortcut on current menu item, per §6 tables)
**Space:** Alias for Enter (in ModeMenu / ModeConfig / ModePreferences)

### 14.3 Mode Entry Keys

**`/` (Main Menu only):** Enter Config mode (§6.3)

### 14.4 Branch Picker Keys (§8)

**`↑/k` / `↓/j`:** Navigate branch list
**`Tab`:** Toggle focus between branch list pane and details pane
**`Enter`:** Context-dependent action (switch / merge / return — see §8 table)
**`a`:** Add new branch
**`m`:** Merge selected branch into current (hidden on current row)
**`x`:** Delete selected branch (hidden on current row)
**`Esc`:** Return to previous mode

### 14.5 History Browser Keys (§11)

**`↑/k` / `↓/j`:** Navigate commits
**`Tab`:** Switch between Commits list and Details pane
**`Enter`:** Enter time travel mode at selected commit (safe, reversible)
**`Ctrl+R`:** REWIND to selected commit — destructive `git reset --hard` (confirmation required)
**`y`:** Copy Hash mode — show hash for copy (any mode)
**`Y`:** Copy Full Hash (in Copy Hash mode)
**`Esc`:** Exit to main menu (or exit Copy Hash mode if active)

### 14.6 Preferences Mode Keys (§6.4)

**`↑/k` / `↓/j`:** Navigate items
**`Enter`:** Execute item
**`+` / `-`:** Adjust interval by ±1 min (on interval item)
**`Shift + / -`:** Adjust interval by ±10 min (on interval item)
**`Esc`:** Return to Config menu

---

## 16. Color Theme

**Theme file:** `~/.config/tit/themes/default.xml`

**Hierarchy:** `THEME > LOOK_AND_FEEL > [component family nodes] > color attributes`. XML format, loaded via `juce::XmlDocument::parse` + `juce::ValueTree::fromXml`. Hot-reload via `jam::File::Watcher` (200 ms coalesce).

**Component families:** `SCREEN`, `TEXT`, `BORDER`, `DIALOG`, `CONFLICT_RESOLVER`, `STATUS`, `TIMELINE`, `OPERATION`, `MENU`, `SPINNER`, `DIFF`, `COPY_HASH_LABEL`, `CONSOLE_STREAM` (13 families, 44 color fields total).

**Attribute naming:** matches Go `theme_gfx.go` TOML keys 1:1 (e.g., `mainBackgroundColor`, `timelineSynchronized`, `conflictPaneFocusedBorder`) for full feature parity with Go TIT reference theme.

**Canonical default:** `themes/default.xml` at project root, installed to `~/.config/tit/themes/default.xml` on first run. Five shipped themes: `gfx` (default), `spring`, `summer`, `autumn`, `winter`.

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
