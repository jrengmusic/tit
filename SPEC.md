# TIT Specification

**TIT:** Terminal Interface for git.
**Purpose:** TIT shows the state of a git repository and gives only the git actions that the state allows.
**Target End-User:** A user who is not a git expert.
**Core Workflow:** The user starts `titc` in a folder. TIT reads the repository state, shows the menu of that state, and runs the action that the user selects. After each action, TIT reads the state again.

---

## 1. Overview

TIT is a deterministic state machine over the state of git.
TIT operates on the current branch only. The user can switch branches at any time.
Each destructive action asks for confirmation.
TIT runs every git operation as a child process and shows its output.

This document states what each feature does and the exact text of each feature.
It states the command and the key of each action.
It states no layout. The build of each view sets its layout.
It states each view by the question that the view answers and the data that the view shows.

---

## 2. Technology Stack

- **Language:** C++17.
- **Frameworks:** JUCE 8 and JAM.
- **Build:** cast.
- **Platforms:** macOS (arm64 and x86_64) and Windows.
- **Git interface:** git runs as a child process only. TIT links no git library.
- **Output:** one binary, `titc`.

---

## 3. Core Principles

### 3.1 Foundational Principle

**The UI of TIT is a pure function of the git state.**

```
Git State → Allowed Actions → Menu
```

If git would reject an action, the action does not appear in the menu.

Key philosophy:

- TIT operates on the **current active branch only**.
- Users can switch branches at any time.
- TIT tracks no branch and stores no repository state.
- The state always reflects the actual git state.
- Destructive operations always require confirmation.

### 3.2 Design Invariants

1. **Menu = Contract:** If an action appears in the menu, it must succeed. TIT never shows an operation that could fail or leave a dangling state.
2. **State Machine:** The UI is a pure function of the git state.
3. **No Staging:** All changes commit together.
4. **Single Active Branch:** TIT operates on the current branch only.
5. **Branch Switching:** Users can switch branches at any time.
6. **Safe Exploration:** Time travel is read-only until merge-back.
7. **Dirty Operations:** TIT manages a dirty operation automatically. The menu of an interrupted dirty operation has an abort command. The git state is safe even if the user exits.
8. **Beautiful:** Themed colours.
9. **Guaranteed Success:** TIT never shows an operation that could fail.
10. **No stored repository state:** The state always reflects the actual git state.
11. **No Dangling States:** The `Merging`, `Rebasing`, `Conflicted`, and `DirtyOperation` states each have an abort command. The user completes the operation, aborts it, or exits TIT. TIT preserves the git state.
12. **Destructive Clarity:** The Ctrl modifier marks a destructive operation (`ctrl+r` = Rewind, Discard all changes). Enter is always safe (time travel, read-only).
13. **No force push for a rewrite:** A command that rewrites commits never needs a force push. TIT makes the command inactive when it would need one.

---

## 4. State Model

### 4.1 The axes

Every decision of TIT comes from **five axes**.

#### GitEnvironment — System prerequisites (Axis 0, checked first)

| Code | Meaning |
|------|---------|
| `Ready` | Git and SSH are available |
| `NeedsSetup` | Git is available. SSH needs configuration |
| `MissingGit` | Git is not installed |
| `MissingSSH` | SSH is not installed |

TIT checks GitEnvironment before all other state detection. If it is not `Ready`, TIT does not proceed to the repository.

#### WorkingTree — Local file changes (Axis 1)

| Code | Meaning |
|------|---------|
| `Clean` | No changes |
| `Dirty` | Has changes (staged, unstaged, or both) |

TIT does not separate staged and unstaged changes. All changes commit together.

#### Timeline — Local against remote

| Code | Meaning |
|------|---------|
| `InSync` | Local is equal to remote |
| `Ahead` | Local is ahead (unpushed commits) |
| `Behind` | Local is behind (unpulled commits) |
| `Diverged` | Both have unique commits |
| *(empty)* | **N/A**: no comparison is possible |

Timeline applies only when HEAD is on a branch and the branch has a remote tracking branch.
When `Operation` is `TimeTraveling` or `Pinned`, or `Remote` is `NoRemote`, Timeline is empty.

#### Operation — Git operation state

| Code | Meaning |
|------|---------|
| `NotRepo` | Not in a git repository |
| `Normal` | No operation in progress |
| `Merging` | Merge in progress |
| `Conflicted` | Operation stopped because of conflicts |
| `Rebasing` | Rebase in progress (can have conflicts) |
| `TimeTraveling` | Detached HEAD (exploring history, read-only) |
| `DirtyOperation` | A dirty operation of TIT stopped before its end. HEAD is on a branch. No merge, rebase, or conflict is in progress. A stash of TIT exists. |
| `Rewinding` | Hard reset in progress (transient) |
| `Pinned` | TIT reset the branch to a past commit. The marker `.git/TIT_PIN` exists. |

#### Remote — Remote repository presence

| Code | Meaning |
|------|---------|
| `NoRemote` | No remote exists |
| `HasRemote` | A remote exists |

#### IsTitTimeTravel — TIT time travel or manual detached HEAD

| Value | Meaning |
|-------|---------|
| `false` | Manual detached HEAD: the user ran `git checkout <commit>` outside TIT. No `.git/TIT_TIME_TRAVEL` marker exists. |
| `true` | TIT time travel: the user pressed Enter on a commit in History. The marker `.git/TIT_TIME_TRAVEL` exists. |

- When `Operation` is `TimeTraveling`, `IsTitTimeTravel` sets the return path.
- Both cases show the Time Travel menu.
- The return differs:
  - `IsTitTimeTravel = true`: the return reads the marker.
  - `IsTitTimeTravel = false`: the return opens the Branch Picker.

**State Tuple:** `(GitEnvironment, WorkingTree, Timeline, Operation, Remote)`.

### 4.2 LFS fields

The state has two LFS fields:

| Field | Meaning |
|---|---|
| LFS | The repository uses LFS: `.gitattributes` holds `filter=lfs`. |
| LFS ready | The `git-lfs` program is on the path, and `git config --get filter.lfs.process` gives a value. |

### 4.3 Markers

A marker is a file in `.git/`. Each marker has one writer and one remover.

| Marker | Lines | Written by | Removed by |
|---|---|---|---|
| `.git/TIT_DIRTY_OP` | branch; original HEAD | the Dirty Operation Protocol, step 1 (Section 5.6) | its finalize and its abort (Section 5.6) |
| `.git/TIT_TIME_TRAVEL` | branch; the label of the stash of the entry, or an empty line when the entry made no stash | entering time travel (Section 5.9) | the return (Section 5.9) |
| `.git/TIT_PIN` | branch; original HEAD; pinned hash | pin (Section 5.13) | unpin (Section 5.13) |

**Stale markers.** TIT removes a stale marker at state detection.

- The time travel marker is stale when HEAD is on a branch.
- The pin marker is stale when HEAD is not on the branch of the marker.
- The pin marker is stale when HEAD is not the pinned hash.
- The dirty marker is stale when no stash of TIT exists.

TIT detects the state after each chain of command lines, never inside a chain.
So a marker that a chain writes first is not stale before the chain ends.

**The end of a pin by another command.** Rewind and time travel are available while `Pinned`. Each one moves HEAD away from the pinned hash. The pin marker then becomes stale, and TIT removes it. The pin ends, and its original HEAD is no longer recorded.

**No restore at the start.** TIT discards no change of the user at its start. At its start, TIT removes a stale marker. A stash of TIT that remains gives the `DirtyOperation` menu.

### 4.4 State Priority Rules

**Pre-flight check (before any state detection): GitEnvironment**

- `MissingGit` → fatal error: Git is not installed.
- `MissingSSH` → fatal error: SSH is not installed.
- `NeedsSetup` → the setup wizard (Section 5.2).
- `Ready` → git state detection.

GitEnvironment is not part of the state tuple of the repository.

**Priority 2: Mid-Operation State Recovery**

- `Conflicted` → the conflict resolver opens (resolve or abort).
- `Merging` → the menu offers finalize merge (commit) or abort.
- `Rebasing` → the conflict resolver opens if conflicts exist. Otherwise the menu offers continue or abort.
- `DirtyOperation` → the menu offers to reapply the saved changes or to abort and restore.
- These states are recoverable. TIT has the machinery for each of them.

**Priority 3: Operation State**

- `Normal` → check the other axes.
- `TimeTraveling` → the Time Travel menu (History, File(s) history, Return).
  - Entered from: History, select a commit, Enter (TIT time travel); or `git checkout <commit>` outside TIT (manual detached).
  - It is not a menu item. TIT enters it from History, or detects it at the start.
- `Pinned` → the Pinned menu (Section 5.1).

**Priority 4: Remote Presence**

- `NoRemote` → hide the sync actions. Show "Add remote".
- `HasRemote` → show the sync actions of the Timeline.

**Priority 5: WorkingTree + Timeline**

- These two axes select the action items of the menu.

### 4.5 Mid-Operation Recovery (start and run time)

TIT recovers every mid-operation state: a state that exists at the start, and a state that a TIT operation makes.

#### Conflicted

- The conflict resolver opens (Section 5.12).
- The user selects one version for each file. TIT stages the resolved files.
- Finalize: `git commit` (merge), `git rebase --continue` (rebase), or the drop of the stash (a conflict of `stash apply`). The underlying operation selects the finalize.
- Abort of a merge: `git merge --abort`, then `git reset --hard`.
- Abort of a rebase: `git rebase --abort`.
- Abort of a conflict of `stash apply`: `git reset --hard HEAD`. The stash stays, and the `DirtyOperation` menu follows.

#### Merging (no conflicts)

- A merge is in progress and no conflict exists.
- The menu offers: Finalize merge (commit) or Abort merge.

#### Rebasing

- If conflicts exist, the conflict resolver opens.
- After the resolution: `git rebase --continue`.
- If the next commit also conflicts, the conflict resolver opens again.
- The loop continues until git replays all commits or the user aborts.
- Abort at any step: `git rebase --abort` (restores the state before the rebase).
- The menu offers: Continue rebase (if no conflicts) or Abort rebase.

#### DirtyOperation

- A dirty pull, merge, switch, rebase, pin, or unpin of TIT stopped before its end.
- `.git/TIT_DIRTY_OP` holds the original branch and the original HEAD.
- Reapply changes: apply the stash of TIT, then drop it.
- Abort and restore: restore the original state from the marker data (Section 5.6, Abort).

#### State Transitions

```
Conflicted ──→ Conflict Resolver ──→ Finalize/Abort ──→ Normal
Merging    ──→ Finalize/Abort menu ──→ Normal
Rebasing   ──→ Conflict Resolver (loop) ──→ Continue/Abort ──→ Normal
DirtyOp    ──→ Reapply/Abort menu ──→ Normal
```

---

## 5. Features

### 5.1 Menu

TIT builds the menu from the state. One menu exists for each `Operation` value.
A change of the state builds the menu again.
Each menu item has an id, an emoji, a label, a shortcut, and a hint.

#### When Operation = NotRepo

- Initialize repository: a location choice follows.
- Clone repository: a location choice follows.

**Location choice of Initialize repository:**

1. Initialize directory (current folder).
2. Create subdirectory (TIT asks for a name, then initializes there).

**Location choice of Clone repository:**

1. Clone to directory (clones into the current folder).
2. Create subdirectory (clones into a new folder).

#### When Operation = Normal

Menu order: **Working Tree → Timeline → separator → History → separator and Add remote**.

- The first separator shows only when an item is above it.
- Add remote and its separator show only when `Remote = NoRemote`.

**Working Tree actions**

| State | Menu items |
|-------|------------|
| `Clean` | *(none: TIT hides the section)* |
| `Dirty` | Commit changes<br>Commit and push *(only if HasRemote)*<br>View changes<br>Discard all changes (`ctrl + r`) |

**Timeline actions**

TIT hides the Timeline section when `Remote = NoRemote`.

`InSync`: *(no items)*.

`Ahead`:

| WorkingTree | Menu items |
|-------------|------------|
| `Clean` | Push to remote<br>Force push (`shift + ]`)<br>Incoming / outgoing |
| `Dirty` | Incoming / outgoing |

`Behind`:

| WorkingTree | Menu items |
|-------------|------------|
| `Clean` | Pull (fetch + merge)<br>Pull (rebase)<br>Replace local<br>Incoming / outgoing |
| `Dirty` | Pull (save changes) *(Dirty Operation Protocol)*<br>Replace local<br>Incoming / outgoing |

`Diverged`:

| WorkingTree | Menu items |
|-------------|------------|
| `Clean` | Push (auto sync)<br>Pull (merge)<br>Pull (rebase)<br>Force push (`shift + ]`)<br>Replace local<br>Incoming / outgoing |
| `Dirty` | Pull (save changes) *(Dirty Operation Protocol)*<br>Force push (`shift + ]`)<br>Replace local<br>Incoming / outgoing |

Pull (rebase) is absent when the tree is `Dirty`.

**History actions** (always shown in `Normal`):

- History
- File(s) history

**Remote setup** (only when `Remote = NoRemote`):

- Add remote

**Branch management.** Switch, add, delete, merge from, and rebase onto are not menu items. They are in the Branch Picker (Section 5.7). The Branch Picker opens from the Config menu. `/` opens the Config menu.

#### When Operation = TimeTraveling

- History (view other commits while time traveling)
- File(s) history (view file changes and diffs)
- Return: the label is `Return to {branch}` when the branch is known, and `Return to branch` when it is not.

The branch is known when the time travel marker names it, or when the repository has one local branch only.
The return paths are in Section 5.9.

#### When Operation = Pinned

- History
- File(s) history
- Discard all changes (`ctrl + r`): only when `Dirty`. It shows the dialog `confirm_discard_changes_local`.
- Unpin

No commit item. No Timeline item.

#### When Operation = Merging

- Finalize merge
- Abort merge

#### When Operation = Rebasing

- Continue rebase
- Abort rebase

#### When Operation = Conflicted

No menu. The conflict resolver opens (Section 5.12). Escape in the resolver aborts the operation.

#### When Operation = DirtyOperation

- Reapply changes
- Abort and restore: active when `.git/TIT_DIRTY_OP` exists.

#### History items while the cache builds

TIT builds a cache of the history at the start and after each change of HEAD.
While it builds, the History and File(s) history items are inactive, and each shows a turning spinner frame:

| Item | Label before a count exists | Label with a count |
|---|---|---|
| History | `History...` | `History {n}/{total}` (example: `History 12/100`) |
| File(s) history | `Files...` | `Files {n}/{total}` (example: `Files 12/100`) |

The spinner frames are `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`.

#### Config menu

`/` opens the Config menu from an operation menu. Escape or "Back" closes it.

| Item | Label | Shortcut | Condition |
|------|-------|----------|-----------|
| `config_add_remote` | Add Remote | `r` | `Remote = NoRemote` |
| `config_switch_remote` | Switch Remote | `s` | `Remote = HasRemote` |
| `config_remove_remote` | Remove Remote | `r` | `Remote = HasRemote` |
| *(separator)* | | | |
| `config_branch` | Branch | `b` | always |
| `config_preferences` | Preferences | `p` | always |
| *(separator)* | | | |
| `config_back` | Back | *(none)* | always |

- `config_add_remote` and `config_switch_remote` never show together.
- `config_add_remote` and `config_remove_remote` share `r`. They never show together.
- `config_branch` opens the Branch Picker.
- `config_preferences` opens the Preferences menu.

#### Preferences menu

Navigation only: no shortcut letter.

| Item | Label | Action |
|------|-------|--------|
| `preferences_auto_update` | Auto-update | Toggles the background refresh `ON` or `OFF`. |
| `preferences_interval` | Update Interval | `=` and `-` change it by 1 minute. `+` and `_` change it by 10 minutes. The value shows as `{n} min`. |
| `preferences_theme` | Theme | Cycles through the theme folders of `~/.config/tit/themes/`. |

#### The menu items

Each item of every menu. Text is exact. `—` marks an empty cell.

| id | Emoji | Label | Shortcut | Hint |
|---|---|---|---|---|
| `init` | 🔨 | Initialize repository | `i` | Create a new git repository |
| `clone` | 📥 | Clone repository | `c` | Clone an existing repository from remote URL |
| `commit` | 📝 | Commit changes | `c` | Create a new commit with staged changes |
| `commit_push` | 🚀 | Commit and push | `p` | Stage, commit, and push changes in one action |
| `view_changes` | — | View changes | `v` | View uncommitted changes file by file |
| `reset_discard_changes` | 💥 | Discard all changes | `ctrl + r` | 💥 DESTRUCTIVE: Discard uncommitted changes, reset to remote state |
| `push` | 📤 | Push to remote | `]` | Send local commits to remote branch |
| `force_push` | 💥 | Force push | `shift + ]` | 💥 DESTRUCTIVE: Overwrite remote branch with local commits |
| `dirty_pull_merge` | 🔺 | Pull (save changes) | `shift + [` | Save WIP, pull remote, reapply changes (may conflict) |
| `pull_merge` | 📥 | Pull (fetch + merge) | `[` | Fetch latest from remote and merge into local branch |
| `pull_rebase` | — | Pull (rebase) | `r` | Fetch remote and replay local commits on top |
| `replace_local` | 💥 | Replace local | `x` | 💥 DESTRUCTIVE: Discard local commits, match remote exactly |
| `pull_merge_diverged` | 📥 | Pull (merge) | `[` | Fetch remote and merge diverged branches |
| `push_auto_sync` | 📤 | Push (auto sync) | `]` | Merge remote changes then push — TIT handles conflicts |
| `incoming_outgoing` | — | Incoming / outgoing | `o` | View the commits a pull or push will transfer |
| `history` | 📜 | History | `h` | View and navigate through commit history |
| `file_history` | 📄 | File(s) history | `f` | View history of specific files |
| `add_remote` | 🔗 | Add remote | `r` | Configure a remote repository URL |
| `time_travel_history` | 🕒 | History | `h` | View commit history while time traveling |
| `time_travel_files_history` | 📄 | File(s) history | `g` | Browse file changes and diffs |
| `time_travel_return` | 🔙 | Return | `r` | Return without merging changes |
| `unpin` | — | Unpin | `u` | Return the branch to where it was before the pin |
| `init_here` | 📍 | Initialize directory | `1` | Create repository here |
| `init_subdir` | 📁 | Create subdirectory | `2` | Create new folder and initialize there |
| `clone_here` | 📍 | Clone to directory | `1` | Clone repository here |
| `clone_subdir` | 📁 | Create subdirectory | `2` | Create new folder and clone there |
| `finalize_merge` | ✅ | Finalize merge | `f` | Complete the merge by committing |
| `abort_merge` | ↩️ | Abort merge | `a` | Abort merge and return to pre-merge state |
| `rebase_continue` | ▶️ | Continue rebase | `c` | Continue rebase after resolving conflicts |
| `rebase_abort` | ↩️ | Abort rebase | `a` | Abort rebase and return to pre-rebase state |
| `dirty_continue` | ▶️ | Reapply changes | `c` | Apply your saved changes again |
| `dirty_abort` | ↩️ | Abort and restore | `a` | Restore the state before the operation |
| `config_add_remote` | 🔗 | Add Remote | `r` | Configure a remote repository URL |
| `config_switch_remote` | 🔗 | Switch Remote | `s` | Change the remote repository URL |
| `config_remove_remote` | 🗑️ | Remove Remote | `r` | Remove the configured remote repository |
| `config_branch` | 🌿 | Branch | `b` | Open branch picker |
| `config_preferences` | ⚙️ | Preferences | `p` | Configure auto-update and theme settings |
| `config_back` | 🔙 | Back | — | Return to main menu |
| `preferences_auto_update` | 🔄 | Auto-update | — | Toggle background state refresh ON/OFF |
| `preferences_interval` | ⏱️ | Update Interval | — | Adjust with +/- keys (±1 min) or Shift+/- (±10 min) |
| `preferences_theme` | 🎨 | Theme | — | Cycle through available themes |

The label of `time_travel_return` changes to `Return to {branch}` or `Return to branch` (Section 5.1, TimeTraveling).
While the history cache builds, a turning spinner frame replaces the emoji of History and File(s) history.

**Items that no menu shows.** These items have no menu. Another command does their work.

| id | Emoji | Label | Shortcut | Hint | The command that does the work |
|---|---|---|---|---|---|
| `time_travel_merge` | 📦 | Merge back | `m` | Merge changes back to original branch | `time_travel_return`, the button `Merge changes` (Section 5.9) |
| `config_toggle_auto_update` | 🔄 | Toggle Auto Update | `u` | Enable/disable background timeline sync | `preferences_auto_update` |
| `config_new_branch` | ➕ | New Branch | `n` | Create a new branch from current HEAD and switch to it | Branch Picker, `a` |
| `config_switch_branch` | 🌿 | Switch Branch | `b` | Switch to a different local branch | Branch Picker, Enter |
| `config_merge_branch` | 🔀 | Merge from... | `m` | Merge another branch into the current branch | Branch Picker, `m` |

### 5.2 Startup

#### 5.2.1 GitEnvironment check (before any git operation)

TIT checks the machine before the git state detection:

1. `git --version`. If git is not found: `MissingGit`, fatal error.
2. `ssh -V`. If ssh is not found: `MissingSSH`, fatal error.
3. TIT scans `~/.ssh` for private keys with the names `id_rsa`, `id_ed25519`, `*_rsa`, `*_ed25519`. If no key exists: `NeedsSetup`, the setup wizard opens.
4. If all checks pass: `Ready`, the git state detection starts.

#### 5.2.2 Setup wizard

The setup wizard makes an SSH key and configures it. It has seven phases:

| Phase | Purpose |
|-------|---------|
| Welcome | Explains what the wizard does. |
| Prerequisites | Verifies git and ssh. Stops on `MissingGit` or `MissingSSH`. |
| Email | Text input for the comment of the SSH key (the email of the user). |
| Generate | Runs `ssh-keygen -t rsa -b 4096 -C "<email>" -f ~/.ssh/TIT_id_rsa -N ""`. Writes the `Host *` block to `~/.ssh/config`. Adds the key to a running agent. |
| Display key | Shows the public key and the addresses of the git providers. Copies the public key to the clipboard. |
| Complete | Success message. The user acknowledges. GitEnvironment becomes `Ready`. |
| Error | Shows the error. Escape goes back to the email step. |

The wizard closes on the acknowledgement of Complete. Then the git state detection starts.
If `~/.ssh/TIT_id_rsa` exists, the Generate phase makes no new key.
TIT starts no `ssh-agent`. If no agent runs, `ssh-add` fails and the console shows the text of the tool.

**Texts of the wizard:**

| Where | Text |
|---|---|
| Welcome title | `Welcome to TIT` |
| Welcome body | `We'll set up SSH authentication for git.`<br>`This only needs to be done once per machine.` |
| Welcome button | `Continue` |
| Git found | `✓ Git found` |
| Git missing | `✗ Git not found` |
| SSH found | `✓ SSH found` |
| SSH missing | `✗ SSH not found` |
| Install hint | `Install missing tools:`<br>`  macOS:   brew install git openssh`<br>`  Linux:   sudo apt install git openssh-client`<br>`  Windows: git-scm.com/download` |
| Prerequisites button | `Re-check` |
| Email prompt | `Email (for SSH key comment):` |
| Email without `@` or `.` (footer) | `Email must hold @ and .` |
| Generate title | `Generating SSH Key` |
| Key title | `SSH Key Generated` |
| Key copied | `Your public key has been copied to clipboard` |
| Providers | `Add this key to your git provider:`<br>`  • GitHub:    github.com/settings/ssh/new`<br>`  • GitLab:    gitlab.com/-/user_settings/ssh_keys`<br>`  • Bitbucket: bitbucket.org/account/settings/ssh-keys` |
| SSH config block | `Host *`<br>`  AddKeysToAgent yes`<br>`  IdentityFile ~/.ssh/TIT_id_rsa`<br>`  IdentitiesOnly yes` |
| Config written | `✓ Configured ~/.ssh/config` |
| Config failed | `⚠ Could not write config (configure manually)` |
| Error title | `✗ Setup Error` |
| Error hint | `Press ESC to go back and try again.` |
| Error button | `Go Back` |
| Complete title | `✓ Setup Complete!` |
| Complete body | `TIT is ready. You can now init or clone repositories.` |

#### 5.2.3 Repository check

```bash
git rev-parse --git-dir
```

If it fails, `Operation = NotRepo`, and the menu shows Initialize repository and Clone repository.

**Init:**

1. The user selects the location (Section 5.1, NotRepo).
2. For a new folder, the prompt `Repository name:` asks for its name.
3. The prompt `Initial branch name:` asks for the branch name. The default text is `main`.
4. TIT runs `git init` and `git checkout -b <branch>`.

**Clone:**

1. When the current folder is empty, the user selects the location (Section 5.1, NotRepo).
2. When the current folder is not empty, TIT shows no location choice. The clone goes into a new folder.
3. The prompt `Repository URL:` asks for the address.
4. "Clone to directory": TIT removes `.DS_Store` and `.AppleDouble`. Then TIT runs `git clone --progress <url> .`.
5. "Create subdirectory": TIT runs `git clone --progress <url>`. The new folder takes its name from the address.
6. The clone checks out the default branch of the remote.

**Texts of init, clone, and the remote:**

| Where | Text |
|---|---|
| Repository name prompt | `Repository name:` |
| Empty repository name | `Directory name cannot be empty` |
| Initial branch prompt | `Initial branch name:` |
| Empty branch name | `Branch name cannot be empty` |
| Repository URL prompt | `Repository URL:` |
| Empty repository URL | `Repository URL cannot be empty` |
| Invalid repository URL | `Invalid URL format. Try: git@github.com:user/repo.git` |
| Remote URL prompt (Add) | `Remote URL:` |
| Remote URL prompt (Switch) | `New remote URL:` |
| Empty remote URL | `Remote URL cannot be empty` |
| Invalid remote URL | `Invalid format. Expected:`<br>`  • SSH: git@github.com:user/repo.git`<br>`  • HTTPS: https://github.com/user/repo.git`<br>`  • Local: /path/to/repo or ~/path/to/repo` |

A valid address starts with `git@`, `http://`, `https://`, `/`, or `~`.

#### 5.2.4 Add remote

1. The prompt `Remote URL:` asks for the address.
2. TIT runs these lines in this order:
   1. `git remote add origin <url>`
   2. `git fetch --all --progress`
   3. `git lfs fetch`, when the repository uses LFS
   4. `git remote set-head origin --auto`
   5. `git branch --set-upstream-to=origin/<branch>`
3. If a line fails, TIT runs `git push --progress -u origin <branch>`.

Switch Remote runs `git remote set-url origin <url>`, then lines 2 to 5.
Remove Remote shows the dialog `remove_remote`, then runs `git remote remove origin`.

#### 5.2.5 Branch mismatch on the first remote

TIT adds the first remote. When the local branch differs from the default branch of the remote, TIT shows the dialog `branch_mismatch` (Section 6.4).

- "Switch to {remote branch}": `git checkout -b <remote branch> origin/<remote branch>`.
- "Stay on {local branch}": the remote stays added. The user stays on the current branch.

The start of `titc` does not open this dialog.

#### 5.2.6 The startup gate

When the repository has a remote, the menu waits for the first fetch.

- The menu shows one inactive item, `Checking Remote...`, with a turning spinner.
- After the fetch, TIT detects the state again and shows the menu of the state.
- If the fetch fails, the menu shows the cached state.
- If the fetch fails, the footer shows `Remote unreachable — showing cached state`.
- During the gate, only the quit keys and Escape work.

#### 5.2.7 Auto-update

- When auto-update is `ON`, TIT fetches and detects the state again at each interval (Section 7).
- An auto-update runs only in an operation menu, and only after 30 seconds with no key.
- During a refresh, the header shows a turning spinner and `Syncing...`.
- A failed fetch shows `Remote unreachable — showing cached state` in the footer.

#### 5.2.8 Fatal errors

A manual detached HEAD is not a fatal error (Section 5.10).

`MissingGit` and `MissingSSH` are fatal. TIT shows the Prerequisites phase of the wizard with the install hint.

TIT does not check `git config user.name` and `git config user.email`. TIT does not detect a bare repository.
In both cases, a command that fails shows the text of the tool in the console.

### 5.3 Console

The console shows the output of each git command that TIT runs.

- Each command that streams progress runs with `--progress`.
- A progress line that ends with a carriage return replaces the last line of the console.
- A line that ends with a newline adds a line.
- `↑` `↓` `k` `j` `page up` `page down` scroll the console. The console editor owns these keys (Section 6.3).
- Escape while a command runs: aborts the command. The footer shows `Operation aborted by user`.
- Escape after the command ends: closes the console.

**Footer while a command runs:** `↑↓ scroll  │  Esc abort`.
**Footer after the command ends:** `↑↓ scroll  │  Esc back`.

### 5.4 LFS

TIT supports Git LFS. LFS has no menu item and no key.

- At the start, TIT runs `git lfs install` when all three conditions are true:
  1. The repository uses LFS.
  2. The LFS filters are not ready.
  3. The `git-lfs` program exists.

  Then the footer shows `LFS filters installed automatically`.
- When the repository uses LFS and `git-lfs` does not exist, the footer shows `This project uses LFS. Install git-lfs for full functionality.`
- After the fetch of a new remote: `git lfs fetch`.
- After each hard reset: `git lfs checkout`, when LFS is ready.
- The header shows `LFS` when LFS is ready.
- The header shows `LFS ⚠️` when the repository uses LFS and LFS is not ready.
- The header shows no badge when the repository does not use LFS.

### 5.5 Text input

**The text inputs:**

| Input | Prompt | Hint |
|---|---|---|
| Commit message | `Commit message:` | `Enter message and press Enter` |
| Repository name (init or clone in a new folder) | `Repository name:` | `Enter new directory name` (legacy key `subdir_name`); `Enter subdirectory name for new repository` (legacy key `init_subdir_name`) |
| Initial branch name | `Initial branch name:` | `Enter branch name (default: main), press Enter to initialize` |
| Clone address | `Repository URL:` | `Enter git repository URL (https or git+ssh)` |
| Remote address (Add) | `Remote URL:` | `Enter git repository URL and press Enter` |
| Remote address (Switch) | `New remote URL:` | `Enter git repository URL and press Enter` |
| Email of the setup wizard | `Email (for SSH key comment):` | — |
| New branch name | `New branch name:` | — |
| Reword of a commit (Section 5.14) | the old message | — |
| Search history (Section 5.11) | `Search history:` | — |
| Function history (Section 5.11) | `Function history:` | — |

**Rules:**

- Each input is a `jam::TextEditor` that opens in `insert` mode.
- The commit message and the reword are multi-line editors with the vim core.
- The vim core has the normal, insert, and visual modes.
- The vim core has the operators `d` `c` `x` `p` `P` and the insert keys `i` `a` `I` `A` `o` `O`.
- The vim core has undo `u` and redo `ctrl+r`.
- Enter in normal mode submits the text. Enter in insert mode of a multi-line editor adds a line.
- Escape in insert mode or visual mode goes to normal mode.
- A paste arrives whole. `ctrl+v` and `alt+v` insert the text of the clipboard at the caret.
- A bracketed paste inserts its text whole.
- Escape in normal mode on a text that is not empty: the footer shows `Press ESC again to clear input (2s timeout)`.
- A second Escape within 2 seconds clears the text.
- Escape in normal mode on an empty text: closes the input.
- An empty commit message shows `Commit message cannot be empty` in the footer.

**Footer of an input:** `Enter submit  │  Esc back`.

### 5.6 Dirty Operation Protocol

**Purpose:** Apply a change-set while the uncommitted work stays safe.

**Used by:**

- Pull (merge) with `WorkingTree = Dirty`.
- Merge from a branch with `WorkingTree = Dirty`.
- Switch branch with `WorkingTree = Dirty`.
- Rebase onto a branch with `WorkingTree = Dirty`.
- Pin and unpin with `WorkingTree = Dirty`.

Each operation shows its dialog first. Section 6.4 holds the dialogs.

Time travel with `WorkingTree = Dirty` saves the changes in a stash of TIT too. It does not write `.git/TIT_DIRTY_OP`. The time travel marker records that stash (Section 5.9).

**Steps:**

1. **Snapshot**

   ```bash
   ORIGINAL_BRANCH=$(git symbolic-ref --short HEAD)
   ORIGINAL_HEAD=$(git rev-parse HEAD)
   git stash push -u -m "TIT DIRTY-OP SNAPSHOT"
   echo "$ORIGINAL_BRANCH" > .git/TIT_DIRTY_OP
   echo "$ORIGINAL_HEAD" >> .git/TIT_DIRTY_OP
   ```

2. **Apply the change-set** (pull, merge, checkout, rebase, reset).
   - If conflicts occur: `Operation = Conflicted`.
   - The user resolves. The protocol continues.

3. **Apply the snapshot again**

   ```bash
   git stash apply <stash of TIT>
   ```

   - If conflicts occur: `Operation = Conflicted`.
   - The user resolves. The protocol continues.

4. **Finalize**

   ```bash
   git stash drop <stash of TIT>
   rm .git/TIT_DIRTY_OP
   ```

**Abort (at any step):**

```bash
ORIGINAL_BRANCH=$(head -n 1 .git/TIT_DIRTY_OP)
ORIGINAL_HEAD=$(tail -n 1 .git/TIT_DIRTY_OP)
git checkout $ORIGINAL_BRANCH
git reset --hard $ORIGINAL_HEAD
git stash apply <stash of TIT>
git stash drop <stash of TIT>
rm .git/TIT_DIRTY_OP
```

→ Restores the exact original dirty state.

**The stash of TIT.** Every stash of TIT has the label `TIT DIRTY-OP SNAPSHOT`. A command names that stash by its reference. TIT never applies or drops a stash of the user.

**Properties:**

- Source-agnostic: it works for each operation above.
- Conflict-first: it stops at the first conflict.
- Rollback-safe: the abort always works.
- No stash stacking: one temporary snapshot.

### 5.7 Branch Picker

**One view for all branch management.** Add, switch, merge from, delete, and rebase onto are all in the picker.

**Question it answers:** Which branches exist, and what is the state of each?

**Data:**

- The local branches. The current branch has the mark `●`.
- For the selected branch: its name, its status, its upstream, and its tip commit (hash, author, date, message). The tracking status is `↑{n} ↓{n}`, `synced`, or `local`.

**Texts of the details:**

| Where | Text |
|---|---|
| Heading | `BRANCH` |
| Name | `Name: {name}` |
| Current | `Status: ● Current` |
| Not current | `Status: Not current` |
| Upstream | `Upstream: {upstream} ({status})` |
| No upstream | `Upstream: none (local only)` |
| Tip heading | `TIP COMMIT` |
| Hash | `Hash: {hash}` |
| Author | `Author: {author}` |
| Date | `Date: {date}` |
| Tracking in sync | `synced` |
| Tracking with no upstream | `local` |
| New branch prompt | `New branch name:` |

#### Entry points

- **Config menu:** the item Branch opens the picker.
- **Time travel return** on a manual detached HEAD with more than one branch: the picker opens to select the return branch.
- **After a new branch:** the picker opens again with the new branch selected.

#### Commands

| Key | Action | Availability |
|-----|--------|-----------|
| `Tab` | Moves the focus between the branches and the details. | always |
| `Enter` | The primary action (see below). | always |
| `a` | **Add branch.** The prompt `New branch name:` asks for a name. TIT creates the branch from HEAD and switches to it. Then the picker opens with the new branch selected. | always |
| `m` | **Merge the selected branch into the current branch.** Shows the dialog `merge_branch`. With `Dirty`: the dialog `merge_branch_dirty`, then the Dirty Operation Protocol. | hidden on the current branch |
| `r` | **Rebase the current branch onto the selected branch** (Section 5.14). | hidden on the current branch |
| `x` | **Delete the selected branch.** Shows the dialog `branch_delete`, then runs `git branch -D <branch>`, then shows the picker again. | hidden on the current branch |
| `Esc` | Returns to the menu that opened the picker. | always |

#### Enter

| Context | Selected branch is the current branch | Selected branch is another branch |
|---------|-------------------|------------------------|
| **Default** (from Config, Branch) | Returns to the Config menu. | **Switch branch.** With `Dirty`: the dialog `branch_switch_dirty` (Stash changes, Discard changes; Escape cancels). With `Clean`: `git checkout <branch>`. |
| **Return to branch** (manual detached) | *(no current branch row)* | Leaves the detached HEAD for the selected branch (Section 5.10). |

#### Footer

- Current branch selected: `↑↓ navigate  │  a add  │  Tab details  │  Esc cancel`
- Another branch selected: `↑↓ navigate  │  a add  │  m merge from  │  r rebase onto  │  x delete  │  Enter switch  │  Esc cancel`

#### Command lines

- Delete: `git branch -D <name>`.
- Add: `git checkout -b <new>`. The tool checks the name (`git check-ref-format --branch`). A refused name shows the text of the tool in the console.
- Merge from: `git merge <selected>`. With `Dirty`, the Dirty Operation Protocol.
- Switch: `git checkout <name>`.
- Rebase onto: `git rebase <selected>` (Section 5.14).

### 5.8 Merge Branch

Merge from another branch is the `m` command of the Branch Picker (Section 5.7).
When the merge gives conflicts, `Operation = Conflicted`, and the conflict resolver opens.
The abort (`git merge --abort`) returns to `Normal` with no change.

With `WorkingTree = Dirty`, the Dirty Operation Protocol runs: snapshot, merge, apply again, finalize.
The uncommitted work stays safe across the merge.

### 5.9 Time Travel

#### 5.9.1 Enter time travel

**Available from:** History. Enter on a commit.

1. TIT shows the dialog `time_travel` with the short hash and the subject of the commit.
2. TIT writes the marker.
3. With `WorkingTree = Dirty`, TIT saves the changes in a stash of TIT with no other question.
4. TIT checks out the commit:

   ```bash
   ORIGINAL_BRANCH=$(git symbolic-ref --short HEAD)
   echo "$ORIGINAL_BRANCH" > .git/TIT_TIME_TRAVEL
   echo "<TIT DIRTY-OP SNAPSHOT when the tree is Dirty, or empty>" >> .git/TIT_TIME_TRAVEL
   git stash push -u -m "TIT DIRTY-OP SNAPSHOT"    # when the tree is Dirty
   git checkout <commit-hash>
   ```

5. `Operation = TimeTraveling`. HEAD is detached, so the stash of the entry does not give `DirtyOperation`.

#### 5.9.2 While time traveling

**Header:** `📌 DETACHED @ {hash}` and `Viewing commit from {date}`. The date has the form `%b %d, %Y` (example: `Dec 30, 2024`).

**Available actions:**

- History: jump to a different commit. Enter on a commit while time traveling moves the time travel to that commit. The jump makes no stash. The marker keeps its lines.
- File(s) history: view the file changes and diffs.
- Return (Section 5.9.3).

**Behavior:**

- The user can change files in the working tree (`WorkingTree = Dirty`).
- No commit: the menu has no commit item.
- The user can build, test, and experiment.
- Changes stay local until the return.

**At the start of TIT.** A time travel that TIT finds at its start stays a time travel. The menu gives the return.

#### 5.9.3 Return to the branch

The Return item has two paths:

- **TIT time travel** (the marker exists): a dialog opens. The branch comes from the marker.
- **Manual detached HEAD** (no marker): Section 5.10.

**Clean tree.** TIT shows the dialog `time_travel_return` (Return, Cancel). Return runs:

```bash
ORIGINAL_BRANCH=$(head -n 1 .git/TIT_TIME_TRAVEL)
git checkout $ORIGINAL_BRANCH
git stash apply <stash of the entry>    # when the entry made a stash
git stash drop <stash of the entry>     # when the entry made a stash
rm .git/TIT_TIME_TRAVEL
```

**Dirty tree.** TIT shows the dialog `time_travel_return_dirty` with two buttons, `Merge changes` and `Discard changes`. Escape cancels.

**Merge changes** keeps the changes of the time travel. It merges them into the original branch.
Two stashes can exist: the stash of the return (step 1) and the stash of the entry. Both have the label of a stash of TIT.

1. Save the changes of the time travel. This stash is the newest stash:

   ```bash
   git stash push -u -m "TIT DIRTY-OP SNAPSHOT"
   ```

2. Save the commit of the detached HEAD:

   ```bash
   DETACHED_COMMIT=$(git rev-parse HEAD)
   ```

3. Check out the original branch:

   ```bash
   ORIGINAL_BRANCH=$(head -n 1 .git/TIT_TIME_TRAVEL)
   git checkout $ORIGINAL_BRANCH
   ```

4. Merge the detached commit:

   ```bash
   git merge $DETACHED_COMMIT
   ```

   - If conflicts occur: `Operation = Conflicted`. The user resolves. The return continues.

5. Apply the stash of the return, then drop it:

   ```bash
   git stash apply <stash of the return>
   git stash drop <stash of the return>
   ```

   - If conflicts occur: `Operation = Conflicted`. The user resolves. The return continues.

6. When the entry made a stash, apply it, then drop it:

   ```bash
   git stash apply <stash of the entry>
   git stash drop <stash of the entry>
   ```

   - If conflicts occur: `Operation = Conflicted`. The user resolves. The return continues.

7. Remove the marker:

   ```bash
   rm .git/TIT_TIME_TRAVEL
   ```

**Abort of Merge changes.** Escape in the resolver runs the abort of the conflict type (Section 4.5).
HEAD is then on the branch, so the time travel marker is stale, and TIT removes it.
A stash of TIT that remains gives the `DirtyOperation` menu. Reapply changes applies it.

**Discard changes** discards the changes of the time travel.
Then TIT checks out the original branch, applies the stash of the entry, and removes the marker.

**New state:** `Operation = Normal`, on the original branch.

#### 5.9.4 Stash Not Found

At the return, the second line of the marker can name a stash that no longer exists.
Then TIT shows the dialog `stash_not_found` before it continues:

- Continue: the return runs with the tree as it is. A clean tree takes the return lines. A dirty tree takes the lines of Merge changes.
- Cancel: nothing changes.

### 5.10 Manual Detached HEAD

TIT supports the return from a manual detached HEAD (the user ran `git checkout <commit>` outside TIT).

#### Detection

When `Operation = TimeTraveling` and `IsTitTimeTravel = false`:

- No `.git/TIT_TIME_TRAVEL` marker exists.
- The user checked out a commit outside TIT.
- TIT shows the Time Travel menu with the Return item.

The header shows `Not on a branch. Select 'Return to branch' to continue.`

#### Menu

| Menu item | Description |
|-----------|-------------|
| **History** | Jump to other commits (as TIT time travel) |
| **File(s) history** | View file changes and diffs |
| **Return to branch** | Check out a branch |

#### Return to branch

1. Select "Return to branch".
2. The Branch Picker opens with all local branches. With one local branch only, TIT selects it with no picker.
3. Select the branch.

**Clean tree:**

- TIT checks out the branch.
- `Operation = Normal`.

**Dirty tree:** TIT shows the dialog `time_travel_return_dirty_stash` (Section 6.4). Its buttons are `Stash changes` and `Discard changes`. Escape cancels.

**Stash changes:**

1. `git stash push -u -m "TIT DIRTY-OP SNAPSHOT"`.
2. Check out the branch.
3. Merge the detached commit into the branch.
4. Apply the stash (this can conflict).
5. Drop the stash.

**Discard changes:**

1. TIT discards the changes and checks out the branch.
2. `Operation = Normal`.

#### Conflicts

1. **Merge conflict:** the detached commit changes the same lines as the branch. The conflict resolver opens. After the resolution, the stash apply runs.
2. **Stash apply conflict:** the stashed changes conflict with the merged result. The conflict resolver opens. The resolution completes the return.

Both use the conflict resolver of Section 5.12.

#### State Transitions

```
Manual Detached (clean)
    ├─ Return to branch → Normal
    └─ History → Manual Detached (different commit)

Manual Detached (dirty)
    ├─ Return to branch → Stash/Discard dialog
    │    ├─ Stash → Merge → Apply Stash → Normal (or Conflict)
    │    └─ Discard → Normal
    └─ History → Manual Detached (different commit)
```

### 5.11 History browser

**One browser of records.** History is one browser. Every reading view of TIT is this browser with a different feed.

**Questions it answers:**

- What changed in this commit?
- How big is each change?
- Which characters changed in this line?
- Which function is this hunk in?

**The three levels, from the top down:**

1. The commits. Each commit shows its date, its short hash, and its subject. The date has the form `%d-%b %H:%M` (example `30-Dec 14:23`).
2. For the selected commit: its message, author, date, and files. Each file shows its added line count (`+`) and its deleted line count (`-`).
3. For the selected file: its diff.

**Every diff in TIT** marks the changed characters inside a changed line, and shows the function name of each hunk.

**Entry:**

- History opens the browser at the commits.
- File(s) history opens the browser at the files of the selected commit.
- TIT has no blame.

**The cache.** At the start, TIT reads the last 100 commits into a cache. The user row `history count` sets the count (Section 7).
The History items are inactive while the cache builds (Section 5.1).

#### The data of each level (the reading contract)

| Question | Tool output |
|---|---|
| What changed in this commit? | `log` and `show` with `--name-status`, then the patch of a file |
| How big is each change? | `--numstat` |
| Which characters changed in this line? | `--word-diff` |
| Which function is this hunk in? | the hunk header text |
| How did this function change over time? | `log -L :<funcname>:<file>` |
| When did this text appear or vanish? | `log -S<string>` |
| What will a pull give? What does upstream lack? | `log --left-right A...B` |
| What is the state of the tree now? | `diff --numstat -p HEAD` and `ls-files --others --exclude-standard` |

#### The feeds

| Feed | Opened by | Question | Records |
|---|---|---|---|
| History | the item History; the item File(s) history | What changed in each commit? | the commits of the cache |
| Working tree changes | the item View changes (`Dirty` only) | What did I change? | One record: the changed files and their added and deleted line counts. The selected file shows its diff: the working tree against HEAD. |
| Incoming and outgoing | the item Incoming / outgoing (`Behind`, `Ahead`, or `Diverged`) | What will a pull give? What will a push give? | Two groups of records: the commits that a pull gives, and the commits that a push gives. Each commit shows its files and their diffs. |
| Search history | `/` in the browser | When did this text appear or vanish? | the commits where the typed text appeared or vanished |
| Function history | `F` in the browser, on the selected file | How did this function change? | the commits that changed the typed function, each with its diff |

- The prompt of the search is `Search history:`. The prompt of the function history is `Function history:`. Both use one text input in `insert` mode. Enter in normal mode starts the request.
- A search with no result shows `No commit matches`.

#### Commands of the browser

| Key | Action |
|---|---|
| `Enter` on a commit | Time travel to the commit (Section 5.9). |
| `ctrl+r` on a commit | Rewind to the commit. |
| `p` on a commit | Pin the branch at the commit (Section 5.13). |
| `i` | Interactive rebase of the unpushed commits (Section 5.14). |
| `y` on a commit | Copy the short hash. The footer shows `✓ Copied to clipboard`. |
| `Y` on a commit | Copy the full hash. The footer shows `✓ Copied to clipboard`. |
| `/` | Search history. |
| `F` on a file | Function history of the file. |
| `Tab` | Moves the focus between the records and the text of the selected record. |
| `v` in the text | Visual mode. `y` copies the selection. |
| `Esc` | Returns to the menu. |

The hash commands do nothing on the record of the working tree changes, because it has no hash.

**Footer of the records:** `↑↓ navigate  │  Enter time travel  │  y copy hash  │  Tab details  │  Esc back`
**Footer of the text:** `↑↓ scroll  │  v visual  │  Tab commits  │  Esc back`

#### Rewind

**Purpose:** Reset the current branch to a past commit for ever.

**Availability:** Always available in the history browser, in every Operation state.

**Behavior:**

- Discards all commits after the selected commit.
- Discards all uncommitted changes.
- Requires the destructive confirmation `rewind`.

**Command line:** `git reset --hard <commit>`, then `git lfs checkout` when LFS is ready.

### 5.12 Conflict Resolver

**Question it answers:** Which version of each conflicted file do I keep?

**Data:**

- The unmerged files. Each file shows its chosen version. A file with no chosen version shows `[ ]`.
- Three versions of the selected file: the base, the current side, and the incoming side.

**The labels of the three versions:**

| Operation | Base | Current | Incoming |
|---|---|---|---|
| Merge, pull, merge from a branch | `BASE` | `{branch} (current)` | `{name} (incoming)` |
| Rebase | `BASE` | `OURS (rebased)` | `THEIRS (incoming)` |
| Branch switch | `BASE` | `{branch} (current)` | `{branch} (incoming)` |
| Dirty merge, the reapply of the snapshot | `BASE` | `{branch} (merged)` | `stashed changes` |
| Dirty switch | `BASE` | `{target} (current)` | `{original} (stashed)` |

`{name}` of a merge is the name that the tool gives to `MERGE_HEAD` (`git name-rev --name-only MERGE_HEAD`).

A version that does not exist shows this text:

```
[FILE DELETED IN THIS VERSION]

This file was deleted in {version}.
The conflict occurred because the other side modified it.
```

**Commands:**

| Key | Action |
|---|---|
| `Tab` | Moves the focus through the files and the three versions. |
| `Space` on a version | Chooses that version for the selected file. The footer shows `Marked: {file} → {version}`. |
| `Enter` | Resolves with the chosen versions. When a file has no chosen version, the footer shows `Mark all files with SPACE before continuing`. |
| `Esc` | Aborts the operation, for every type of conflict. |

**Resolve:**

1. For each file, TIT writes the chosen version: `git checkout-index -f --stage=<n> -- <path>`, then `git add -- <path>`.
2. A chosen version that does not exist removes the file: `git rm -f -- <path>`.
3. The lines of all files are one chain. A failure stops the chain, and the console shows the text of the tool.
4. TIT finalizes by the underlying operation:
   - a merge: `git commit --no-edit`;
   - a rebase: `git rebase --continue`;
   - a stash apply: the drop of the stash.

**Footer:**

- The files have the focus: `↑↓ navigate  │  Tab versions  │  Enter resolve  │  Esc abort`
- A version has the focus: `↑↓ scroll  │  Space choose  │  Tab next  │  Enter resolve  │  Esc abort`

### 5.13 Pin

**Purpose:** Hold the branch at a past commit, with no new branch and no detached HEAD.

**Entry:** `p` on a selected commit in the history browser, beside time travel and rewind.

**The pin lines.** These lines run when the user presses a button of a pin dialog:

1. TIT writes `.git/TIT_PIN` with three lines: the branch, the original HEAD, the pinned hash. TIT writes the marker first.
2. TIT runs `git reset --hard <hash>`.
3. When LFS is ready, `git lfs checkout` follows the reset.

**Clean tree:** TIT shows the dialog `pin`. The button Pin runs the pin lines. Then `Operation = Pinned`.

**Dirty tree:** TIT shows the dialog `pin_dirty`, and no other dialog. Escape cancels.

- Save changes: the Dirty Operation Protocol runs with the pin lines as its change-set.
- Discard changes: TIT discards the changes, then runs the pin lines.

**While pinned:**

- The menu is the Pinned menu (Section 5.1).
- The header shows the branch and `PINNED @ {hash}`.
- The user can change files in the tree.
- No commit. The interactive rebase is inactive.

**Pin again** from History while pinned: the marker keeps its original HEAD and takes the new pinned hash.
Then TIT runs `git reset --hard <new hash>`.

**The unpin lines:** `git reset --hard <original HEAD>`, then `git lfs checkout` when LFS is ready, then the removal of the marker.

- Clean tree: TIT shows the dialog `unpin`. The button Unpin runs the unpin lines.
- Dirty tree: TIT shows the dialog `unpin_dirty`, and no other dialog. Escape cancels.
  - Save changes: the Dirty Operation Protocol runs with the unpin lines as its change-set.
  - Discard changes: TIT discards the changes, then runs the unpin lines.

**Conflict:** a conflict of the snapshot apply gives `Operation = Conflicted` and the conflict resolver.
**Abort:** the abort of the Dirty Operation Protocol (Section 5.6).

### 5.14 Rebase

#### Pull with rebase

- Menu item `pull_rebase`, shown with `Behind` or `Diverged` and `Clean`.
- TIT shows the dialog `pull_rebase`, then runs `git pull --rebase --progress`.
- A conflict gives `Rebasing` or `Conflicted`, and the rebase loop of Section 4.5 runs.
- Abort: `git rebase --abort`.

#### Rebase onto a branch

- A command of the Branch Picker on another branch: `r`, beside merge. The picker hides it on the current branch, as it hides merge.
- Clean tree: TIT shows the dialog `rebase_branch`, then runs `git rebase <selected>`.
- Dirty tree: TIT shows the dialog `rebase_branch_dirty`. Then TIT runs the Dirty Operation Protocol with the rebase as its change-set.
- It is active only when every commit that it replays is unpushed.
- A pushed commit to replay is a commit that HEAD and the upstream hold, and the selected branch does not hold.
- When a commit to replay is pushed, `r` writes `Pushed commits cannot be rebased` in the footer and runs nothing.
- A conflict gives the rebase loop of Section 4.5. Abort: `git rebase --abort`.

#### Interactive rebase

1. **Entry:** `i` in the history browser.
2. **Active when:** `Operation` is not `Pinned`, the tree is `Clean`, and at least one commit is ahead of the remote.
   - With no remote, every commit of the branch is in the range.
   - With a `Dirty` tree, `i` writes `Commit or discard changes first` in the footer and runs nothing.
   - With no commit ahead of the remote, `i` writes `No unpushed commits to rewrite` in the footer and runs nothing.
3. **Range:** the commits of `@{u}..HEAD`, or all commits of the branch with no upstream. The user cannot select a commit outside the range.
4. **The list:** the commits of the range, the oldest first. Each row has an action: `pick`, `squash`, `fixup`, `reword`, or `drop`. Each row starts as `pick`.
5. **The edits:**

   | Key | Edit |
   |---|---|
   | `s` | squash into the commit before |
   | `f` | fixup into the commit before |
   | `r` | reword: the commit message editor opens with the old message (Section 5.5) |
   | `d` | drop |
   | `J` | move down (later in time) |
   | `K` | move up (earlier in time) |
   | `Enter` | apply |
   | `Esc` | cancel |

   The key of the action that a row has sets the row back to `pick`.
6. **Apply:** TIT shows the dialog `rebase_interactive`, then runs the rebase with the edited list.
   - The base is `--keep-base @{u}`, or `--root` with no upstream.
   - No editor opens on the terminal.
7. **Cancel:** changes nothing.
8. **Conflict:** the rebase loop of Section 4.5. **Abort:** `git rebase --abort`.

**Footer of the list:** `s squash  │  f fixup  │  r reword  │  d drop  │  J K move  │  Enter apply  │  Esc cancel`

### 5.15 Header

**Question it answers:** Where am I, and what is the state?

**Data:**

- the current folder and the remote address;
- the operation and the branch;
- the working tree and its description;
- the timeline and its description;
- the LFS badge and the version of TIT.

**Texts:**

| Where | Text |
|---|---|
| Folder mark | `📁` |
| Remote mark | `🔗` |
| No remote | `🔌 NO REMOTE` |
| Branch mark | `🌿` |
| Not available | `N/A` |
| Separator | `─` |
| LFS ready | `LFS` |
| LFS not ready | `LFS ⚠️` |
| Clean | `✅ Clean` — `No local changes` |
| Dirty | `📝 Dirty` — `Local changes present` |
| Modified mark | `●` |
| Timeline N/A | `🔌 N/A` |
| In sync | `🔗 In sync` — `In sync with remote` |
| Ahead | `🌍 Local ahead ⬆ {n}` — `{n} commit(s) ahead` |
| Behind | `🪐 Local behind ⬇ {n}` — `{n} commit(s) behind` |
| Diverged | `💥 Diverged ⬆ {ahead} ⬇ {behind}` — `{ahead}↑ {behind}↓` |
| No remote description | `No remote configured.` |
| Normal | `🟢 READY` |
| Not a repository | `🔴 NOT REPO` |
| Conflicted | `⚡ CONFLICTED` |
| Merging | `🔀 MERGING` |
| Rebasing | `🔄 REBASING` |
| Dirty operation | `⚡ DIRTY OP` |
| Time travel | `🌀 TIME TRAVEL` |
| Detached | `DETACHED`, `📌 DETACHED @ {hash}` |
| Pinned | `PINNED @ {hash}` |
| Time travel date | `Viewing commit from {date}` |
| Manual detached | `Not on a branch. Select 'Return to branch' to continue.` |
| Refresh | `{spinner} Syncing...` |
| Local check | `Checking local state...` |
| Remote check | `Fetching remote updates...` |

**State descriptions:**

| State | Text |
|---|---|
| Clean | `No local changes` |
| Dirty | `Local changes present` |
| InSync | `In sync with remote` |
| Ahead | `{n} commit(s) ahead` |
| Behind | `{n} commit(s) behind` |
| Diverged | `{n}↑ {n}↓` |
| Normal | `Ready` |
| NotRepo | `Not a repository` |
| Conflicted | `Conflicts detected` |
| Merging | `Merge in progress` |
| Rebasing | `Rebase in progress` |
| DirtyOperation | `Operation started with local changes` |
| TimeTraveling | `Viewing commit {hash} ({date})` |

**Terminal size.** When the terminal is smaller than 70 columns by 20 rows, TIT shows this text:

```
Terminal too small.
Resize to >= 70×20.
```

### 5.16 Quit

- `ctrl+c`, first press: the footer shows `Press Ctrl+C again to quit (2s timeout)`.
- `ctrl+c`, second press within 2 seconds: TIT exits.
- `q` in a menu does the same as `ctrl+c`.
- While an operation runs: the footer shows `Exit blocked. Operation must complete or be aborted first.`

### 5.17 Error handling

| Error condition | End-user sees | System action |
|---|---|---|
| A git command fails | The text of the tool in the console. The footer shows `Failed. Press ESC to return.` | TIT stops the chain at the failed line, then detects the state again. |
| A command gives conflicts | The conflict resolver. | `Operation = Conflicted` (Section 5.12). |
| The user presses Escape while a command runs | The footer shows `Operation aborted by user`. | TIT stops the command, then detects the state again. |
| The first fetch fails | The footer shows `Remote unreachable — showing cached state`. | TIT shows the menu of the cached state. |
| An auto-update fetch fails | The footer shows `Remote unreachable — showing cached state`. | The next interval tries again. |
| An empty commit message | The footer shows `Commit message cannot be empty`. | The input stays open. |
| An empty branch name | The footer shows `Branch name cannot be empty`. | The input stays open. |
| An empty or invalid address | The text of the address row of Section 5.2.3. | The input stays open. |
| An email without `@` or `.` | The footer shows `Email must hold @ and .` | The input stays open. |
| A branch name that git refuses | The text of the tool in the console. | No branch is made. |
| A conflict with no chosen version for a file | The footer shows `Mark all files with SPACE before continuing`. | The resolver stays open. |
| The stash of a time travel entry is gone | The dialog `stash_not_found`. | Section 5.9.4. |
| An invalid user file | The footer shows the error with its file and line. | TIT uses the default of the file. |
| The terminal is too small | `Terminal too small.` and `Resize to >= 70×20.` | — |
| `git` or `ssh` is missing | The Prerequisites phase with the install hint. | TIT does not open the repository. |
| Ctrl+C while an operation runs | The footer shows `Exit blocked. Operation must complete or be aborted first.` | TIT does not exit. |

### 5.18 Banner and help

**The banner** is the TIT logo, `Assets/tit_logo.svg`.

- TIT draws the logo in two forms: braille cells, and pixels through the inline image sequence of the terminal.
- The build of the view selects the form.
- TIT sends no byte that only one host terminal reads.
- The banner shows beside the operation menus, the Config menu, and the Preferences menu.

**`titc --help`** prints the braille banner and the help guide to the terminal, then exits.

The braille form maps each cell to 2 by 4 dots. Each cell has one colour.

---

## 6. Commands and Keys

### 6.1 Availability

A command that git would reject is inactive.
While its state rule fails, the menu has no item for the command, and the key sheet has no row for it.
A change of the state builds the menu and the key sheet again.

Two commands keep their key and write a footer text in place of their work:

- Rebase onto a branch, when a commit to replay is pushed.
- Interactive rebase, when the tree is `Dirty` or no commit is unpushed.

| Command | State | Footer text |
|---|---|---|
| rebase onto | a commit to replay is pushed | `Pushed commits cannot be rebased` |
| interactive rebase | no commit is ahead of the remote | `No unpushed commits to rewrite` |
| interactive rebase | the tree is `Dirty` | `Commit or discard changes first` |

The open menu moves its own highlight on `↓`, `j`, `↑`, `k`. Return and Space run the highlighted item. Escape closes the menu. No key row moves a menu.

A key row names its command in words. The words map to the command names of Section 6.2:

| Key row | Command |
|---|---|
| `switch remote` | `config_switch_remote` |
| `branch picker` | `config_branch` |
| `preferences` | `config_preferences` |
| `add remote` in `menuConfig` | `config_add_remote`; with a remote, `config_remove_remote` |
| `history` in `menuTimeTraveling` | `time_travel_history` |
| `file history` in `menuTimeTraveling` | `time_travel_files_history` |
| `push` with `Diverged` | `push_auto_sync` |
| any other row | the command with the same words, joined by `_` |

### 6.2 Commands

| Command | Where | Active when | Description |
|---|---|---|---|
| `init` | NotRepo menu | `Operation = NotRepo` | Initialize a repository (location choice follows). |
| `clone` | NotRepo menu | `Operation = NotRepo` | Clone a repository (location choice follows). |
| `init_here` | init location menu | always | Initialize the current folder. |
| `init_subdir` | init location menu | always | Initialize a new folder. |
| `clone_here` | clone location menu | always | Clone into the current folder. |
| `clone_subdir` | clone location menu | always | Clone into a new folder. |
| `commit` | Normal menu | `Normal`, `Dirty` | Commit all changes with a typed message. |
| `commit_push` | Normal menu | `Normal`, `Dirty`, `HasRemote` | Commit all changes, then push. |
| `view_changes` | Normal menu | `Normal`, `Dirty` | Open the browser with the working tree changes. |
| `reset_discard_changes` | Normal menu, Pinned menu | (`Normal` or `Pinned`), `Dirty` | Discard all changes. Dialog `confirm_discard_changes_remote_choice` with `HasRemote` in `Normal`; dialog `confirm_discard_changes_local` otherwise. |
| `push` | Normal menu | `Normal`, `Ahead`, `Clean` | `git push --progress`. Its key `]` runs `push_auto_sync` with `Diverged`. |
| `force_push` | Normal menu | `Normal`, (`Ahead` and `Clean`) or `Diverged` | Dialog `force_push`, then `git push --progress --force`. |
| `pull_merge` | Normal menu | `Normal`, `Behind`, `Clean` | Dialog `pull_merge`, then `git pull --no-rebase --progress`. |
| `pull_merge_diverged` | Normal menu | `Normal`, `Diverged`, `Clean` | Dialog `pull_merge_diverged`, then the pull of `pull_merge`. |
| `pull_rebase` | Normal menu | `Normal`, `Behind` or `Diverged`, `Clean` | Dialog `pull_rebase`, then `git pull --rebase --progress`. |
| `dirty_pull_merge` | Normal menu | `Normal`, `Behind` or `Diverged`, `Dirty` | Dialog `dirty_pull`, then the Dirty Operation Protocol with the pull as its change-set. |
| `replace_local` | Normal menu | `Normal`, `Behind` or `Diverged` | Dialog `hard_reset`, then `git fetch --progress origin`, `git reset --hard origin/<branch>`, `git clean -fd`, and `git lfs checkout` when LFS is ready. |
| `push_auto_sync` | Normal menu | `Normal`, `Diverged`, `Clean` | `git pull --no-rebase --progress`, then `git push --progress`. A conflict opens the resolver. |
| `incoming_outgoing` | Normal menu | `Normal`, `HasRemote`, `Behind` or `Ahead` or `Diverged` | Open the browser with the incoming and outgoing commits. |
| `history` | Normal menu, Pinned menu | the cache is ready | Open the browser at the commits. |
| `file_history` | Normal menu, Pinned menu | the cache is ready | Open the browser at the files of the selected commit. |
| `add_remote` | Normal menu | `NoRemote` | Add a remote. With a remote, `r` in the Normal menu runs `pull_rebase`. |
| `time_travel_history` | Time Travel menu | the cache is ready | Open the browser at the commits. |
| `time_travel_files_history` | Time Travel menu | the cache is ready | Open the browser at the files of the selected commit. |
| `time_travel_return` | Time Travel menu | `TimeTraveling` | Return to the branch (Sections 5.9.3, 5.10). |
| `time_travel_merge` | no menu | — | Its work is the button `Merge changes` of `time_travel_return`. |
| `unpin` | Pinned menu | `Pinned` | Unpin (Section 5.13). |
| `finalize_merge` | Merging menu | `Merging` | `git commit --no-edit`. |
| `abort_merge` | Merging menu | `Merging` | `git merge --abort`. |
| `rebase_continue` | Rebasing menu | `Rebasing`, no conflict | `git rebase --continue`. |
| `rebase_abort` | Rebasing menu | `Rebasing` | `git rebase --abort`. |
| `dirty_continue` | DirtyOperation menu | `DirtyOperation` | Apply the stash of TIT, drop it, remove the marker. |
| `dirty_abort` | DirtyOperation menu | `DirtyOperation`, the marker exists | The abort of the Dirty Operation Protocol. |
| `config` | operation menus | always | Open the Config menu. |
| `config_add_remote` | Config menu | `NoRemote` | Add a remote (Section 5.2.4). With a remote, its key `r` runs `config_remove_remote`. |
| `config_switch_remote` | Config menu | `HasRemote` | Change the remote address. |
| `config_remove_remote` | Config menu | `HasRemote` | Dialog `remove_remote`, then `git remote remove origin`. |
| `config_branch` | Config menu | always | Open the Branch Picker. |
| `config_preferences` | Config menu | always | Open the Preferences menu. |
| `config_back` | Config menu | always | Return to the main menu. |
| `config_new_branch` | no menu | — | Its work is `a` of the Branch Picker. |
| `config_toggle_auto_update` | no menu | — | Its work is `preferences_auto_update`. |
| `preferences_auto_update` | Preferences menu | always | Toggle auto-update. |
| `preferences_interval` | Preferences menu | always | Change the interval with `=` `-` `+` `_`. |
| `preferences_theme` | Preferences menu | always | Select the next theme folder. |
| `preferences_increment` | Preferences menu | always | Add 1 minute to the interval (maximum 60). |
| `preferences_decrement` | Preferences menu | always | Remove 1 minute from the interval (minimum 1). |
| `preferences_increment_ten` | Preferences menu | always | Add 10 minutes to the interval (maximum 60). |
| `preferences_decrement_ten` | Preferences menu | always | Remove 10 minutes from the interval (minimum 1). |
| `time_travel` | history browser | a commit is selected | Time travel to the commit. |
| `rewind` | history browser | a commit is selected | Dialog `rewind`, then reset to the commit. |
| `pin` | history browser | a commit is selected, `Normal` or `Pinned` | Pin the branch at the commit. |
| `rebase_interactive` | history browser | not `Pinned`, `Clean`, at least one unpushed commit | Open the rebase list. |
| `copy_hash` | history browser | a commit is selected | Copy the short hash. |
| `copy_full_hash` | history browser | a commit is selected | Copy the full hash. |
| `search_history` | history browser | always | Open the prompt `Search history:`. |
| `function_history` | history browser | a file is selected | Open the prompt `Function history:`. |
| `search_submit` | search prompt | always | Start the search or the function history. |
| `history_tab` | history browser, conflict resolver, Branch Picker | always | Move the focus to the next pane. |
| `branch_picker_enter` | Branch Picker | always | The primary action of Section 5.7. |
| `branch_picker_add` | Branch Picker | always | Add a branch. |
| `branch_picker_merge` | Branch Picker | another branch is selected | Merge the branch into the current branch. |
| `branch_picker_rebase` | Branch Picker | another branch is selected | Rebase the current branch onto the branch. |
| `branch_picker_delete` | Branch Picker | another branch is selected | Delete the branch. |
| `new_branch_name_submit` | new branch prompt | always | Create the branch. |
| `conflict_space` | conflict resolver | a version has the focus | Choose the version for the file. |
| `conflict_enter` | conflict resolver | always | Resolve with the chosen versions. |
| `rebase_squash` | rebase list | always | Squash. |
| `rebase_fixup` | rebase list | always | Fixup. |
| `rebase_reword` | rebase list | always | Reword. |
| `rebase_drop` | rebase list | always | Drop. |
| `rebase_move_down` | rebase list | always | Move the row down. |
| `rebase_move_up` | rebase list | always | Move the row up. |
| `rebase_apply` | rebase list | always | Apply the list. |
| `reword_submit` | reword editor | always | Keep the message. |
| `input_submit` | text input | always | Submit the text. |
| `paste` | text input | always | Insert the text of the clipboard. |
| `cancel` | every mode | always | Stop, close, or go back (Section 6.5). |
| `quit` | every mode | always | Quit (Section 5.16). |

### 6.3 Keys

This table is the default key sheet. The user overrides it in `~/.config/tit/keys.md` (Section 7).
The columns are `key`, `type`, `value`, `mode`, `description`. `value` holds one or more key descriptions.
`MENUS` stands for these mode names: `menuNotRepo menuNormal menuMerging menuRebasing menuTimeTraveling menuDirtyOperation menuPinned`.

| key | type | value | mode | description |
|---|---|---|---|---|
| `quit` | `string` | `ctrl+c` | every mode | Quit, with confirmation. |
| `quit` | `string` | `q` | MENUS `startup menuConfig menuPreferences menuInitializeLocation menuCloneLocation setupWizard` | Quit, with confirmation. |
| `cancel` | `string` | `escape` | every mode that is not a menu | Stop, close, or go back. |
| `config` | `string` | `/` | MENUS | Open the Config menu. |
| `paste` | `string` | `ctrl+v cmd+v meta+v alt+v` | text inputs | Insert the text of the clipboard. |
| `init` | `string` | `i` | `menuNotRepo` | Initialize repository. |
| `clone` | `string` | `c` | `menuNotRepo` | Clone repository. |
| `init here` | `string` | `1` | `menuInitializeLocation` | Initialize directory. |
| `init subdir` | `string` | `2` | `menuInitializeLocation` | Create subdirectory. |
| `clone here` | `string` | `1` | `menuCloneLocation` | Clone to directory. |
| `clone subdir` | `string` | `2` | `menuCloneLocation` | Create subdirectory. |
| `commit` | `string` | `c` | `menuNormal` | Commit changes. |
| `commit push` | `string` | `p` | `menuNormal` | Commit and push. |
| `view changes` | `string` | `v` | `menuNormal` | View changes. |
| `reset discard changes` | `string` | `ctrl+r` | `menuNormal menuPinned` | Discard all changes. |
| `push` | `string` | `]` | `menuNormal` | Push to remote. With `Diverged`, push (auto sync). |
| `force push` | `string` | `}` | `menuNormal` | Force push. |
| `pull merge` | `string` | `[` | `menuNormal` | Pull. |
| `dirty pull merge` | `string` | `{` | `menuNormal` | Pull (save changes). |
| `replace local` | `string` | `x` | `menuNormal` | Replace local. |
| `incoming outgoing` | `string` | `o` | `menuNormal` | Incoming / outgoing. |
| `add remote` | `string` | `r` | `menuNormal menuConfig` | Add remote. With a remote: Pull (rebase) in `menuNormal`, Remove Remote in `menuConfig`. |
| `history` | `string` | `h` | `menuNormal menuTimeTraveling menuPinned` | History. |
| `file history` | `string` | `f` | `menuNormal menuPinned` | File(s) history. |
| `file history` | `string` | `g` | `menuTimeTraveling` | File(s) history. |
| `time travel return` | `string` | `r` | `menuTimeTraveling` | Return. |
| `unpin` | `string` | `u` | `menuPinned` | Unpin. |
| `finalize merge` | `string` | `f` | `menuMerging` | Finalize merge. |
| `abort merge` | `string` | `a` | `menuMerging` | Abort merge. |
| `rebase continue` | `string` | `c` | `menuRebasing` | Continue rebase. |
| `rebase abort` | `string` | `a` | `menuRebasing` | Abort rebase. |
| `dirty continue` | `string` | `c` | `menuDirtyOperation` | Reapply changes. |
| `dirty abort` | `string` | `a` | `menuDirtyOperation` | Abort and restore. |
| `switch remote` | `string` | `s` | `menuConfig` | Switch Remote. |
| `branch picker` | `string` | `b` | `menuConfig` | Branch. |
| `preferences` | `string` | `p` | `menuConfig` | Preferences. |
| `preferences increment` | `string` | `=` | `menuPreferences` | Add one minute to the interval. |
| `preferences decrement` | `string` | `-` | `menuPreferences` | Remove one minute from the interval. |
| `preferences increment ten` | `string` | `+` | `menuPreferences` | Add ten minutes to the interval. |
| `preferences decrement ten` | `string` | `_` | `menuPreferences` | Remove ten minutes from the interval. |
| `time travel` | `string` | `return` | `history` | Time travel to the selected commit. |
| `rewind` | `string` | `ctrl+r` | `history` | Rewind to the selected commit. |
| `pin` | `string` | `p` | `history` | Pin the branch at the selected commit. |
| `rebase interactive` | `string` | `i` | `history` | Rewrite the unpushed commits. |
| `copy hash` | `string` | `y` | `history` | Copy the short hash. |
| `copy full hash` | `string` | `Y` | `history` | Copy the full hash. |
| `search history` | `string` | `/` | `history` | Search the history. |
| `function history` | `string` | `F` | `history` | Show the history of one function of the selected file. |
| `search submit` | `string` | `return` | `search` | Start the request. |
| `history tab` | `string` | `tab` | `history conflictResolve branchPicker` | Go to the next pane. |
| `conflict space` | `string` | `spacebar` | `conflictResolve` | Choose this version for the file. |
| `conflict enter` | `string` | `return` | `conflictResolve` | Resolve with the chosen versions. |
| `branch picker enter` | `string` | `return` | `branchPicker` | Switch to the branch. |
| `branch picker add` | `string` | `a` | `branchPicker` | Add a branch. |
| `branch picker merge` | `string` | `m` | `branchPicker` | Merge the branch into the current branch. |
| `branch picker rebase` | `string` | `r` | `branchPicker` | Rebase the current branch onto the branch. |
| `branch picker delete` | `string` | `x` | `branchPicker` | Delete the branch. |
| `new branch name submit` | `string` | `return` | `branchName` | Create the branch. |
| `rebase squash` | `string` | `s` | `rebaseList` | Squash into the commit before. |
| `rebase fixup` | `string` | `f` | `rebaseList` | Fixup into the commit before. |
| `rebase reword` | `string` | `r` | `rebaseList` | Change the message. |
| `rebase drop` | `string` | `d` | `rebaseList` | Drop the commit. |
| `rebase move down` | `string` | `J` | `rebaseList` | Move the commit down. |
| `rebase move up` | `string` | `K` | `rebaseList` | Move the commit up. |
| `rebase apply` | `string` | `return` | `rebaseList` | Apply the list. |
| `reword submit` | `string` | `return` | `rebaseReword` | Keep the message. |
| `input submit` | `string` | `return` | `input` | Submit the text. |
| `initial branch name submit` | `string` | `return` | `initialBranchName` | Initialize the repository. |
| `repository name submit` | `string` | `return` | `repositoryName` | Create the folder. |
| `repository url submit` | `string` | `return` | `repositoryUrl` | Clone the repository. |
| `remote url submit` | `string` | `return` | `remoteUrl` | Set the remote. |
| `email submit` | `string` | `return` | `email` | Make the SSH key. |
| `insert before` | `string` | `i` | `normal` | Insert before the cursor. |
| `insert after` | `string` | `a` | `normal` | Insert after the cursor. |
| `insert line start` | `string` | `I` | `normal` | Insert at the start of the line. |
| `insert line end` | `string` | `A` | `normal` | Insert at the end of the line. |
| `open line below` | `string` | `o` | `normal` | Open a line below. |
| `open line above` | `string` | `O` | `normal` | Open a line above. |
| `delete action` | `string` | `d` | `normal visual` | Delete. |
| `change` | `string` | `c` | `normal visual` | Change. |
| `delete character` | `string` | `x` | `normal visual` | Delete the character. |
| `paste after` | `string` | `p` | `normal visual` | Paste after the cursor. |
| `paste before` | `string` | `P` | `normal visual` | Paste before the cursor. |
| `undo` | `string` | `u` | `normal` | Undo. |
| `redo` | `string` | `ctrl+r` | `normal` | Redo. |
| `default mode` | `string` | `escape` | `visual open` | Leave visual mode. |

The edit rows (`insert before` to `redo`) are in the sheet only while an editable text input shows.

**Keys that a component owns.** These keys have no row in the key sheet. The component that has the focus takes them.

| Keys | Where | Action | Owner |
|---|---|---|---|
| `↓` `j` | every menu | Move the highlight down. | the menu |
| `↑` `k` | every menu | Move the highlight up. | the menu |
| `return` `spacebar` | every menu | Run the highlighted item. | the menu |
| `escape` | every menu | Close the menu. | the menu |
| `←` `→` `h` `l` | dialog | Move the highlighted button. | the dialog buttons |
| `y` | dialog | Press the first button. | the dialog buttons |
| `n` | dialog | Press the second button. | the dialog buttons |
| `return` | dialog, a setup wizard step | Press the highlighted button. | the dialog buttons |
| `escape` | dialog | Cancel. | the dialog |
| `↑` `k` `↓` `j` `page up` `page down` | console, text of the history browser | Scroll. | the text editor of the pane |
| `v` | console, text of the history browser | Start visual mode. | the text editor of the pane |
| `y` | console and text of the history browser, in visual mode | Copy the selection. | the text editor of the pane |

### 6.4 Dialogs

Each dialog has a title, an explanation, and two buttons. `<br>` is a line break. `{name}` is a value that TIT fills in. Escape cancels each dialog.

| key | title | explanation | yes | no |
|---|---|---|---|---|
| `force_push` | Force Push Confirmation | This will force push to remote, overwriting remote history.<br><br>Any commits on the remote that you don't have locally will be permanently lost.<br><br>Continue? | Force push | Cancel |
| `hard_reset` | Replace Local Confirmation | This will discard all local changes and commits, resetting to match the remote exactly.<br><br>All uncommitted changes and untracked files will be permanently lost.<br><br>Continue? | Reset to remote | Cancel |
| `dirty_pull` | Save your changes? | You have uncommitted changes. To pull, they must be temporarily saved.<br><br>After the pull, we'll try to reapply them.<br>(This may cause conflicts if the changes overlap.) | Save changes | Discard changes |
| `pull_merge` | Pull from remote? | This will merge remote changes into your local branch.<br><br>If both branches modified the same files, conflicts may occur.<br>You'll be able to resolve them interactively. | Proceed | Cancel |
| `pull_merge_diverged` | Pull diverged branches? | Your branches have diverged (both have new commits).<br><br>This will merge remote changes into your local branch.<br><br>If both modified the same files, conflicts may occur. | Proceed | Cancel |
| `branch_switch_dirty` | Switch to {branch} with uncommitted changes? | You have uncommitted changes. Choose action:<br>(ESC to cancel) | Stash changes | Discard changes |
| `branch_delete` | Delete branch {branch}? | This cannot be undone. Force delete will drop unmerged commits on this branch. | Delete | Cancel |
| `time_travel` | Time Travel Confirmation | {short hash}<br><br>{subject}<br><br>Explore in read-only mode? | Time travel | Cancel |
| `time_travel_return` | Return to {branch} | Any changes you made while time traveling will be STASHED (not discarded).<br><br>Your original work (if any) will be restored.<br><br>Use 'git stash apply stash@{0}' later to restore time travel changes. | Return | Cancel |
| `time_travel_return_dirty` | Return to {branch} with uncommitted changes | You have changes during time travel. Choose action:<br>(Press ESC to cancel) | Merge changes | Discard changes |
| `time_travel_return_dirty_stash` | Return to {branch} with uncommitted changes | You have changes during time travel. Choose action:<br>(Press ESC to cancel) | Stash changes | Discard changes |
| `stash_not_found` | Stash Not Found | Original stash {stash} was manually dropped. Continue without restoring stash? | Continue | Cancel |
| `rewind` | DESTRUCTIVE OPERATION | This will discard all commits after {short hash}.<br>Any uncommitted changes will be lost.<br><br>Are you sure you want to continue? | Rewind | Cancel |
| `confirm_discard_changes_remote_choice` | Discard Changes (Remote Available) | You have uncommitted changes. How would you like to reset?<br><br>• Reset to HEAD: Discards uncommitted changes, keeps local commits.<br>• Reset to Remote: Discards EVERYTHING (commits & changes) to match remote. | Reset to HEAD | Reset to remote |
| `confirm_discard_changes_local` | Discard Changes | This will discard all uncommitted changes and untracked files.<br><br>Your local commits will be preserved.<br><br>Continue? | Discard changes | Cancel |
| `remove_remote` | Remove Remote | This will remove the origin remote from your repository.<br><br>You can add a new remote later.<br><br>Continue? | Remove | Cancel |
| `merge_branch` | Merge branch? | This will merge {selected} into {current}.<br><br>Conflicts will be handled if they occur.<br>Both branches remain intact. | Merge | Cancel |
| `merge_branch_dirty` | Uncommitted changes | You have uncommitted changes.<br><br>To merge, they must be temporarily stashed.<br>After the merge, they'll be reapplied.<br><br>(This may cause conflicts if changes overlap.) | Stash and merge | Cancel |
| `branch_mismatch` | ⚠️ Branch name mismatch | Your local branch: {local}<br>Remote default branch: {remote}<br><br>This may cause confusion. Would you like to switch? | Switch to {remote} | Stay on {local} |
| `pin` | Pin Confirmation | This will move {branch} to commit {hash}.<br><br>The commits after it stay safe. Unpin returns the branch to where it is now.<br><br>You cannot commit while pinned.<br><br>Pin here? | Pin | Cancel |
| `pin_dirty` | Save your changes? | You have uncommitted changes. To pin, they must be temporarily saved.<br><br>After the pin, we'll try to reapply them.<br>(This may cause conflicts if the changes overlap.) | Save changes | Discard changes |
| `unpin` | Unpin? | This will return {branch} to commit {hash}, where it was before the pin.<br><br>Continue? | Unpin | Cancel |
| `unpin_dirty` | Save your changes? | You have uncommitted changes. To unpin, they must be temporarily saved.<br><br>After the unpin, we'll try to reapply them.<br>(This may cause conflicts if the changes overlap.) | Save changes | Discard changes |
| `pull_rebase` | Pull with rebase? | This will replay your local commits on top of the remote changes.<br><br>If both modified the same files, conflicts may occur.<br>You'll be able to resolve them interactively. | Proceed | Cancel |
| `rebase_branch` | Rebase onto branch? | This will replay the commits of {current} on top of {selected}.<br><br>Conflicts will be handled if they occur. | Rebase | Cancel |
| `rebase_branch_dirty` | Uncommitted changes | You have uncommitted changes.<br><br>To rebase, they must be temporarily stashed.<br>After the rebase, they'll be reapplied.<br><br>(This may cause conflicts if changes overlap.) | Stash and rebase | Cancel |
| `rebase_interactive` | Rewrite commits? | This will rewrite {count} unpushed commits.<br><br>Conflicts will be handled if they occur.<br><br>Continue? | Rewrite | Cancel |
| `alert` | {title} | {explanation} | OK | — |

`time_travel_return_dirty` shows on the return of a TIT time travel (Section 5.9.3). `time_travel_return_dirty_stash` shows on the return from a manual detached HEAD (Section 5.10).
`alert` is an informational dialog with one button.

**Dialogs that legacy defines and never shows.** No legacy code opens these dialogs. TIT keeps their text as written.

| Legacy source | title | explanation | yes | no |
|---|---|---|---|---|
| `showNestedRepoWarning` | Nested Repository Detected | The directory '{path}' is inside an existing git repository. Initialize anyway? | Yes, continue | No, cancel |
| `showForcePushWarning` | Force Push Confirmation | This will force push '{branch}' to remote. Any remote changes will be overwritten. Continue? | Force push | Cancel |
| `showHardResetWarning` | Hard Reset Confirmation | This will discard all uncommitted changes. This cannot be undone. Continue? | Reset | Cancel |
| `DialogMessages`, `nested_repo` | Nested Repository Detected | The directory '{path}' is inside another git repository.<br><br>This may cause confusion. Would you like to initialize in a subdirectory instead? | — | — |
| `DialogMessages`, `force_push_nested` | Force Push Confirmation | This will force push to remote, overwriting remote history.<br><br>Any commits on the remote that you don't have locally will be permanently lost.<br><br>Continue? | — | — |
| `DialogMessages`, `hard_reset_nested` | Hard Reset Confirmation | This will discard all local changes and commits, resetting to match the remote exactly.<br><br>All uncommitted changes and untracked files will be permanently lost.<br><br>Continue? | — | — |

### 6.5 Escape

| Context | Escape |
|---|---|
| Text input, text not empty | Asks for confirmation, then clears the text. |
| Text input, text empty | Closes the input. |
| Console, a command runs | Aborts the command. |
| Console, the command ended | Closes the console. |
| History browser | Returns to the menu. |
| Conflict resolver | Aborts the operation. |
| Branch Picker | Returns to the menu that opened it. |
| Rebase list | Cancels the rebase. Nothing changes. |
| Config menu, Preferences menu, a location menu | Returns to the menu before it. |
| Dialog | Cancels the dialog. |
| Setup wizard, Error phase | Returns to the Email phase. |

### 6.6 Footer texts

**Hints by mode:**

| Mode | Text |
|---|---|
| History, the records | `↑↓ navigate  │  Enter time travel  │  y copy hash  │  Tab details  │  Esc back` |
| History, the text | `↑↓ scroll  │  v visual  │  Tab commits  │  Esc back` |
| Conflict resolver, the files | `↑↓ navigate  │  Tab versions  │  Enter resolve  │  Esc abort` |
| Conflict resolver, a version | `↑↓ scroll  │  Space choose  │  Tab next  │  Enter resolve  │  Esc abort` |
| Console, running | `↑↓ scroll  │  Esc abort` |
| Console, ended | `↑↓ scroll  │  Esc back` |
| Input | `Enter submit  │  Esc back` |
| Dialog | `←→ select  │  Enter confirm  │  Esc cancel` |
| Branch Picker, current branch | `↑↓ navigate  │  a add  │  Tab details  │  Esc cancel` |
| Branch Picker, another branch | `↑↓ navigate  │  a add  │  m merge from  │  r rebase onto  │  x delete  │  Enter switch  │  Esc cancel` |
| Rebase list | `s squash  │  f fixup  │  r reword  │  d drop  │  J K move  │  Enter apply  │  Esc cancel` |
| Preferences | `↑↓ navigate  │  Space toggle  │  =/- ±1min  │  +/_ ±10min  │  Esc back` |

**Legacy hint rows that the rows above replace.** The legacy views had more panes. Each row below has its replacement in the table above.

| Legacy mode | Legacy text | Replaced by |
|---|---|---|
| `history_copyhash` | `a-f/0-9 copy highlighted  │  Space next page  │  Esc cancel` | History, the records (`y`, `Y`) |
| `history_details` | `↑↓ scroll  │  Tab list  │  Esc back` | History, the text |
| `filehistory_commits` | `↑↓ navigate  │  Tab files  │  Esc back` | History, the records |
| `filehistory_files` | `↑↓ navigate  │  Tab diff  │  Esc back` | History, the records |
| `filehistory_diff` | `↑↓ scroll  │  v visual  │  Tab commits  │  Esc back` | History, the text |
| `filehistory_visual` | `↑↓ extend  │  y yank  │  Esc cancel` | History, the text |
| `conflict_list` | `↑↓ navigate  │  Space toggle  │  Tab diff  │  Enter resolve` | Conflict resolver, the files |
| `conflict_diff` | `↑↓ scroll  │  Tab list  │  Esc back` | Conflict resolver, a version |
| `input_filled` | `Enter submit  │  Esc clear` | Input |

**Messages:**

| Key | Text |
|---|---|
| `ctrl_c_confirm` | `Press Ctrl+C again to quit (2s timeout)` |
| `esc_clear_confirm` | `Press ESC again to clear input (2s timeout)` |
| `init` | `Initializing repository... (ESC to abort)` |
| `clone` | `Cloning repository... (ESC to abort)` |
| `commit` | `Committing changes... (ESC to abort)` |
| `push` | `Pushing to remote... (ESC to abort)` |
| `pull` | `Pulling from remote... (ESC to abort)` |
| `add_remote` | `Adding remote and fetching... (ESC to abort)` |
| `resolve_conflicts` | `Resolving conflicts...` |
| `operation_complete` | `Press ESC to return to menu` |
| `operation_failed` | `Failed. Press ESC to return.` |
| `operation_in_progress` | `Operation in progress. Please wait for completion.` |
| `operation_aborting` | `Aborting operation. Please wait...` |
| `exit_blocked` | `Exit blocked. Operation must complete or be aborted first.` |
| `aborted` | `Operation aborted by user` |
| `remote_unreachable` | `Remote unreachable — showing cached state` |
| `lfs_installed` | `LFS filters installed automatically` |
| `lfs_missing` | `This project uses LFS. Install git-lfs for full functionality.` |
| `too_small` | `Terminal too small.`<br>`Resize to >= 70×20.` |
| `email_invalid` | `Email must hold @ and .` |
| `pushed_commits` | `Pushed commits cannot be rebased` |
| `no_unpushed_commits` | `No unpushed commits to rewrite` |
| `commit_first` | `Commit or discard changes first` |
| `no_commit_matches` | `No commit matches` |
| `search_prompt` | `Search history:` |
| `function_prompt` | `Function history:` |
| `new_branch_prompt` | `New branch name:` |

**Texts that legacy defines and never shows.** No legacy code reads the map of these texts. TIT keeps their text as written.

| Key | Text |
|---|---|
| `sync_completed` | `Auto-update sync completed` |
| `sync_failed` | `Sync failed: {error}` |
| `auto_update_enabled` | `Auto-update enabled` |
| `auto_update_disabled` | `Auto-update disabled` |
| `auto_update_enabled_syncing` | `Auto-update enabled - syncing...` |

### 6.7 Error messages

| Key | Text |
|---|---|
| `cwd_read_failed` | `Failed to get current directory` |
| `operation_failed` | `Operation failed` |
| `branch_name_empty` | `Branch name cannot be empty` |
| `commit_message_empty` | `Commit message cannot be empty` |
| `remote_url_empty` | `Remote URL cannot be empty` |
| `remote_already_exists` | `Remote 'origin' already exists` |
| `failed_create_dir` | `Failed to create directory: {error}` |
| `failed_change_dir` | `Failed to change to directory: {error}` |
| `failed_detect_state` | `Failed to detect git state: {error}` |
| `failed_cd_into` | `Failed to cd into {path}: {error}` |
| `failed_checkout_branch` | `Failed to checkout branch '{branch}'` |
| `pull_conflicts` | `Merge conflicts occurred` |
| `pull_failed` | `Failed to pull` |
| `failed_stage_resolved` | `Failed to stage resolved files` |
| `failed_commit_merge` | `Failed to commit merge` |
| `failed_abort_merge` | `Failed to abort merge` |
| `failed_reset_after_abort` | `Failed to reset working tree after merge abort` |
| `failed_determine_branch` | `Error: Could not determine current branch` |
| `failed_fetch_remote` | `Failed to fetch from remote` |
| `failed_checkout_original_branch` | `Error: Failed to checkout original branch` |
| `failed_reset_to_original_head` | `Error: Failed to reset to original HEAD` |
| `stash_reapply_failed_but_restored` | `Warning: Could not reapply stash, but HEAD restored` |
| `time_travel_failed` | `Time travel failed: {error}` |
| `time_travel_merge_failed` | `Time travel merge failed: {error}` |
| `time_travel_return_failed` | `Time travel return failed: {error}` |
| `failed_detect_state_after_travel` | `Failed to detect state after time travel: {error}` |
| `failed_detect_state_after_merge` | `Failed to detect state after time travel merge: {error}` |
| `failed_detect_state_after_return` | `Failed to detect state after time travel return: {error}` |
| `failed_get_current_branch` | `Failed to get current branch` |
| `failed_stash_changes` | `Failed to stash changes` |
| `failed_get_stash_list` | `Failed to get stash list` |
| `failed_write_time_travel_info` | `Failed to write time travel info: {error}` |
| `failed_load_time_travel_info` | `Error: {error}` |
| `branch_name_invalid` | `Invalid branch name: {name}` |
| `branch_already_exists` | `Branch '{name}' already exists` |
| `merge_branch_failed` | `Failed to merge branch: {error}` |
| `rewind_commit_hash_empty` | `Commit hash cannot be empty` |
| `rewind_failed` | `Reset failed: {error}` |
| `timeline_sync_no_remote` | `no remote configured` |
| `cannot_set_upstream_detached` | `Skipped: cannot set upstream in detached HEAD state` |
| `working_tree_clean` | `Nothing to commit (working tree clean)` |

### 6.8 Output and console messages

TIT shows these texts in the footer while a chain of command lines runs and after it ends.
The console holds only the bytes of the tool, so no text of this section goes into the console.

**Output messages:**

| Key | Text |
|---|---|
| `remote_added` | `Remote added` |
| `fetch_completed` | `Fetch completed` |
| `pushed_successfully` | `Pushed successfully` |
| `pulled_successfully` | `Pulled successfully` |
| `initializing_repo` | `Initializing repository...` |
| `fetching_remote` | `Fetching from remote...` |
| `setting_upstream` | `Setting upstream tracking...` |
| `detecting_conflicts` | `Detecting conflict files...` |
| `checking_out_branch` | `Checking out branch '{branch}'...` |
| `dirty_pull_snapshot` | `Your changes have been saved` |
| `dirty_pull_snapshot_saved` | `Snapshot saved. Starting merge pull...` |
| `dirty_pull_merge_succeeded` | `Merge succeeded. Reapplying your changes...` |
| `dirty_pull_rebase_succeeded` | `Rebase succeeded. Reapplying your changes...` |
| `dirty_pull_merge_started` | `Pulling from remote (merge strategy)...` |
| `dirty_pull_reapply` | `Reapplying your saved changes...` |
| `dirty_pull_changes_reapplied` | `Changes reapplied. Finalizing...` |
| `dirty_pull_finalize` | `Finalizing dirty pull operation...` |
| `conflict_detection_error` | `Error detecting conflicts: {error}` |
| `conflict_detection_none` | `Conflict detection succeeded but no conflicts found (continuing)` |
| `conflicts_detected_count` | `Conflicts detected in {n} file(s)` |
| `mark_choices_in_resolver` | `Mark your choices in the resolver (SPACE to select, ENTER to continue)` |
| `aborting_dirty_pull` | `Aborting dirty pull...` |
| `aborting_merge` | `Aborting merge...` |
| `aborting_rebase` | `Aborting rebase...` |
| `merge_finalized` | `Merge completed successfully` |
| `merge_aborted` | `Merge aborted - state restored` |
| `abort_successful` | `Successfully aborted by user` |
| `force_push_in_progress` | `Force pushing to remote (overwriting remote history)...` |
| `fetching_latest` | `Fetching latest from remote...` |
| `removing_untracked` | `Removing untracked files and directories...` |
| `failed_clean_untracked` | `Warning: Failed to clean untracked files` |
| `saving_changes_stash` | `Saving your changes (creating stash)...` |
| `discarding_changes` | `Discarding your changes...` |
| `changes_saved_stashed` | `Changes saved (stashed)` |
| `changes_discarded` | `Changes discarded` |
| `merge_conflicts_detected` | `Merge conflicts detected` |
| `merge_completed` | `Merge completed` |
| `reapplying_changes` | `Reapplying your changes...` |
| `stash_apply_conflicts_detected` | `Conflicts detected while reapplying changes` |
| `changes_reapplied` | `Changes reapplied` |
| `dirty_pull_finalize_started` | `Finalizing dirty pull operation...` |
| `stash_drop_failed_warning` | `Warning: Failed to drop stash (manual cleanup may be needed)` |
| `dirty_pull_completed_successfully` | `Dirty pull completed successfully` |
| `dirty_pull_aborting` | `Aborting dirty pull and restoring original state...` |
| `original_state_restored` | `Original state restored` |
| `dirty_switch_snapshot_saved` | `Snapshot saved. Switching branch...` |
| `dirty_switch_started` | `Switching branch...` |
| `dirty_switch_succeeded` | `Branch switch succeeded. Reapplying your changes...` |
| `dirty_switch_finalize_started` | `Finalizing branch switch...` |
| `dirty_switch_completed` | `Branch switch completed successfully` |
| `dirty_switch_aborting` | `Aborting branch switch and restoring original state...` |
| `dirty_merge_snapshot_saved` | `Snapshot saved. Starting merge...` |
| `dirty_merge_started` | `Merging branch...` |
| `dirty_merge_succeeded` | `Merge succeeded. Reapplying your changes...` |
| `dirty_merge_finalize_started` | `Finalizing dirty merge operation...` |
| `dirty_merge_completed` | `Dirty merge completed successfully` |
| `dirty_merge_aborting` | `Aborting dirty merge and restoring original state...` |
| `merge_committed` | `Merge commit created` |
| `rewind_resetting` | `Resetting to commit {hash}...` |
| `rewind_completed` | `Rewind completed successfully` |
| `querying_default_branch` | `Querying remote default branch...` |
| `remote_url_updated` | `Remote URL updated successfully` |
| `remote_removed` | `Remote removed successfully` |
| `repository_initialized` | `Repository initialized with branch '{branch}'` |

**Console messages:**

| Key | Text |
|---|---|
| `marker_corrupted` | `Marker file corrupted. Cleaned up.` |
| `time_travel_success` | `Time travel successful. Press ESC to return to menu.` |
| `time_travel_merge_success` | `Time travel merge successful. Press ESC to return to menu.` |
| `time_travel_return_success` | `Time travel return successful. Press ESC to return to menu.` |
| `time_traveling_status` | `Time traveling... (ESC to abort)` |
| `resolve_conflicts_help` | `Resolve {n} conflicted file(s) - SPACE to mark, ENTER to continue, ESC to abort` |
| `already_marked_column` | `Already marked in this column` |
| `marked_file_column` | `Marked: {file} → {version}` |
| `mark_all_files` | `Mark all files with SPACE before continuing` |
| `error_writing_file` | `Error writing {file}: {error}` |
| `error_staging_file` | `Error staging {file}: {error}` |
| `no_commits_yet` | `No commits yet. Create an initial commit to enable sync actions.` |
| `visual_mode_active` | `-- VISUAL --` |
| `copy_success` | `✓ Copied to clipboard` |
| `copy_failed` | `✗ Copy failed` |

---

## 7. Configuration

**The user folder:** `~/.config/tit/`.

**The files:** markdown tables.

| File | Content |
|---|---|
| `~/.config/tit/user.md` | The preferences: the table `## preferences`. |
| `~/.config/tit/keys.md` | The key sheet: the table of Section 6.3. |
| `~/.config/tit/themes/<name>/theme.md` | One theme: its style sheet as tables. |

**Rules:**

- TIT carries a default of each file.
- When a file is absent or has an old version, TIT writes the default to disk.
- When the file on disk is valid, the disk wins.
- When the file on disk is invalid, TIT uses the default. The file stays on disk.
- For an invalid file, the footer shows the error with its file and line.
- A change of a file on disk makes TIT read it again.
- A change in the Preferences menu writes `user.md`.
- TIT never reads or changes a `.toml` file. The `.toml` files of an earlier TIT stay as they are.

**The preferences (`user.md`):**

| key | type | value | description |
|---|---|---|---|
| `auto update` | `bool` | `true` | Refresh the state in the background. |
| `update interval` | `int` | `5` | Minutes between two refreshes, from 1 to 60. |
| `theme` | `string` | `gfx` | The folder of the theme in ~/.config/tit/themes/. |
| `history count` | `int` | `100` | The count of commits in the history cache. |

The first line of `user.md`: `This file is written with the default values on the first run of titc. Edit a value to change titc. titc reads this file again when it changes. An invalid file gives the default values, and the footer shows the error with its file and line.`

**Auto-update:** `ON` or `OFF`. The interval is 1 to 60 minutes. The default is 5. `=` and `-` change it by 1 minute. `+` and `_` change it by 10 minutes. The value stays between 1 and 60.

**Theme:**

- TIT ships one theme, `gfx`. Its file is `~/.config/tit/themes/gfx/theme.md`.
- A user adds a theme as a folder beside it: `~/.config/tit/themes/<name>/theme.md`.
- The Preferences item Theme cycles through the folders.
- A theme gives colours to 13 families: `SCREEN`, `TEXT`, `BORDER`, `DIALOG`, `CONFLICT_RESOLVER`, `STATUS`, `TIMELINE`, `OPERATION`, `MENU`, `SPINNER`, `DIFF`, `COPY_HASH_LABEL`, `CONSOLE_STREAM`.

---

## 8. Success Criteria

An end-user can:

- [ ] Start `titc` in a folder that is not a repository. See Initialize repository and Clone repository.
- [ ] Run `titc --help` and see the braille banner and the help guide.
- [ ] See the banner beside the operation menus, the Config menu, and the Preferences menu.
- [ ] Initialize a repository in the current folder or in a new folder, with a typed branch name.
- [ ] Clone a repository into the current folder or into a new folder.
- [ ] Complete the setup wizard and get `~/.ssh/TIT_id_rsa` and the `Host *` block in `~/.ssh/config`.
- [ ] See `Email must hold @ and .` for an email without `@` or `.`.
- [ ] See the menu of each Operation state, with exactly the items of Section 5.1.
- [ ] Commit all changes with a message typed in a multi-line vim editor.
- [ ] Push, force push, and push (auto sync) from the states that Section 6.2 names.
- [ ] Pull (merge), pull (rebase), pull (save changes), and replace local from the states that Section 6.2 names.
- [ ] Not see Pull (rebase) when the tree is `Dirty`.
- [ ] Discard all changes with `ctrl+r`, after the dialog of the state.
- [ ] Add, switch, and remove the remote from the Config menu.
- [ ] See `Checking Remote...` at the start of a repository with a remote.
- [ ] See `Remote unreachable — showing cached state` when the first fetch fails.
- [ ] Open History and see the commits, and the files of a commit with their line counts.
- [ ] See the diff of a file with the changed characters marked.
- [ ] See History and File(s) history inactive, with `History 12/100` or `Files 12/100`, while the cache builds.
- [ ] Open View changes with a dirty tree and see each changed file with its diff.
- [ ] Open Incoming / outgoing and see the commits that a pull gives and the commits that a push gives.
- [ ] Search the history for a text and see `No commit matches` when no commit matches.
- [ ] See the history of one function of a file.
- [ ] Copy the short hash with `y` and the full hash with `Y`, and see `✓ Copied to clipboard`.
- [ ] Rewind to a commit with `ctrl+r` after the dialog `rewind`.
- [ ] Time travel to a commit with Enter and see `📌 DETACHED @ {hash}`.
- [ ] Return from a time travel with the changes or without them.
- [ ] See the dialog `Stash Not Found` when the stash of the entry is gone.
- [ ] Return from a manual detached HEAD through the Branch Picker.
- [ ] Pin the branch at a commit with `p` and see `PINNED @ {hash}`.
- [ ] Find no commit item while pinned, and unpin with `u`.
- [ ] Switch, add, merge from, delete, and rebase onto a branch in the Branch Picker.
- [ ] See `Pushed commits cannot be rebased` when `r` in the Branch Picker would replay a pushed commit.
- [ ] Rewrite the unpushed commits with squash, fixup, reword, drop, and move, then apply or cancel.
- [ ] See `No unpushed commits to rewrite` and `Commit or discard changes first` for `i` in those states.
- [ ] Resolve a conflict file by file with Space and Enter, and abort with Escape.
- [ ] Reapply the saved changes, or abort and restore, from the `DirtyOperation` menu.
- [ ] Change the auto-update, the interval, and the theme in the Preferences menu.
- [ ] Find each change of the Preferences menu in `~/.config/tit/user.md`.
- [ ] Change a key in `~/.config/tit/keys.md` and use the new key with no restart.
- [ ] Quit with `ctrl+c` twice within 2 seconds, or with `q` in a menu.

The system:

- [ ] Shows no menu item for a command that git would reject.
- [ ] Builds the same menu and the same active commands from the same state.
- [ ] Removes a stale marker at the state detection, and discards no change of the user at its start.
- [ ] Writes each marker before its operation and removes it after the operation.
- [ ] Never applies or drops a stash that is not labelled `TIT DIRTY-OP SNAPSHOT`.
- [ ] Never needs a force push for a rebase onto a branch or an interactive rebase.
- [ ] Runs `git lfs checkout` after each hard reset when LFS is ready.
- [ ] Runs `git lfs fetch` after the fetch of a new remote.
- [ ] Shows `LFS` or `LFS ⚠️` in the header for a repository that uses LFS.
- [ ] Replaces the last console line for a progress line that ends with a carriage return.
- [ ] Opens no editor on the terminal during a rebase.
- [ ] Never reads or changes a `.toml` file.
- [ ] Uses the default of a user file when the file on disk is invalid.
- [ ] Shows the error of an invalid user file with its file and line.

---

## 9. Architecture Constraints

- ✅ The state is a `jam::Model`.
- ✅ Every command goes through `jam::CommandManager`. The key sheet maps a key to a command.
- ✅ A change of the state is an event. The event builds the menu items and the key sheet.
- ✅ Every view is a `juce::Component` in a `TerminalWindow`.
- ✅ Text input and text output use `jam::TextEditor`.
- ✅ git runs only as a child process.
- ❌ No layout is fixed in this specification.
- ❌ No git library.
- ❌ No stored repository state outside the repository.
- ✅ Every protocol state is a marker, a stash of TIT, a merge, a rebase, or unmerged files.
- ❌ TIT reads and writes no `.toml` file.
