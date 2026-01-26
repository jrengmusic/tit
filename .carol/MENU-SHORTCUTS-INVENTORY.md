# TIT Menu Items and Keyboard Shortcuts — Complete Inventory

**Date:** 2026-01-26  
**Status:** Current state of all menus and shortcuts

---

## Menu Contexts & Items

### 1. **NotRepo Menu** (git.NotRepo)
When not in a git repository:

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `i` | 🔨 | Initialize repository | Create a new git repository |
| `c` | 📥 | Clone repository | Clone an existing repository from remote URL |

---

### 2. **Normal Menu** (git.Normal)
Main menu when repository is ready. Combines:

#### **Working Tree Actions**
Only when `WorkingTree = Dirty`:

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `m` | 📝 | Commit changes | Create a new commit with staged changes |
| `t` | 🚀 | Commit and push | Stage, commit, and push changes in one action |

*(Hidden when `WorkingTree = Clean`)*

---

#### **Timeline Actions**
Only when `Remote = HasRemote`:

**When `Timeline = InSync` AND `WorkingTree = Dirty`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `r` | 💥 | Discard all changes | 💥 DESTRUCTIVE: Discard uncommitted changes, reset to remote state |

**When `Timeline = Ahead` AND `WorkingTree = Clean`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `h` | 📤 | Push to remote | Send local commits to remote branch |
| `f` | 💥 | Force push | 💥 DESTRUCTIVE: Overwrite remote branch with local commits |

**When `Timeline = Behind` AND `WorkingTree = Dirty`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `d` | 🔺 | Pull (save changes) | Save WIP, pull remote, reapply changes (may conflict) |
| `x` | 💥 | Replace local | 💥 DESTRUCTIVE: Discard local commits, match remote exactly |

**When `Timeline = Behind` AND `WorkingTree = Clean`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `p` | 📥 | Pull (fetch + merge) | Fetch latest from remote and merge into local branch |
| `x` | 💥 | Replace local | 💥 DESTRUCTIVE: Discard local commits, match remote exactly |

**When `Timeline = Diverged` AND `WorkingTree = Dirty`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `d` | 🔺 | Pull (save changes) | Save WIP, pull remote, reapply changes (may conflict) |
| `f` | 💥 | Force push | 💥 DESTRUCTIVE: Overwrite remote branch with local commits |
| `x` | 💥 | Replace local | 💥 DESTRUCTIVE: Discard local commits, match remote exactly |

**When `Timeline = Diverged` AND `WorkingTree = Clean`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `p` | 📥 | Pull (merge) | Fetch remote and merge diverged branches |
| `f` | 💥 | Force push | 💥 DESTRUCTIVE: Overwrite remote branch with local commits |
| `x` | 💥 | Replace local | 💥 DESTRUCTIVE: Discard local commits, match remote exactly |

---

#### **History Actions** (Always shown)

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `l` | 📜 | History | View and navigate through commit history |
| `g` | 📄 | File(s) history | View history of specific files |

---

#### **Remote Setup** (Only when `Remote = NoRemote`)

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `e` | 🔗 | Add remote | Configure a remote repository URL |

---

### 3. **Time Traveling Menu** (git.TimeTraveling)
When exploring commit history (detached HEAD):

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `l` | 🕒 | History | View commit history while time traveling |
| `g` | 📄 | File(s) history | Browse file changes and diffs |
| `r` | 🔙 | Return [to branch] | Return without merging changes |

*(Note: "Merge back" was in old code but appears to be removed. Only "Return" in current code.)*

---

### 4. **Init/Clone Location Selection Menus**

#### **Initialize Location Menu**
| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `1` | 📍 | Initialize directory | Create repository here |
| `2` | 📁 | Create subdirectory | Create new folder and initialize there |

#### **Clone Location Menu**
| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `1` | 📍 | Clone to directory | Clone repository here |
| `2` | 📁 | Create subdirectory | Create new folder and clone there |

---

### 5. **Config Menu** (Accessed via ModeConfigMenu)

Dynamic based on remote state:

#### **Remote Section**
**When `Remote = NoRemote`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `a` | 🔗 | Add Remote | Configure a remote repository URL |

**When `Remote = HasRemote`:**

| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `s` | 🔗 | Switch Remote | Change the remote repository URL |
| `r` | 🗑️ | Remove Remote | Remove the configured remote repository |

#### **Branch & Preferences Section**
| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `b` | 🌿 | Switch Branch | Switch to a different local branch |
| `p` | ⚙️ | Preferences | Configure auto-update and theme settings |

#### **Navigation**
| Shortcut | Emoji | Label | Hint |
|----------|-------|-------|------|
| `esc` | 🔙 | Back | Return to main menu |

---

## Global Keyboard Shortcuts

### Navigation (All modes)
| Key | Action |
|-----|--------|
| `↑` / `k` | Move up in list |
| `↓` / `j` | Move down in list |
| `Enter` | Execute/Select |
| `Tab` | Switch pane (in multi-pane modes) |
| `Esc` | Context-dependent: Clear input, Cancel confirmation, Abort operation, Back to menu |

### Menu Navigation (Menu mode only)
| Key | Action |
|-----|--------|
| Letter key | Jump to menu item with that shortcut |

---

## Modal-Specific Shortcuts

### History Browser
| Key | Action |
|-----|--------|
| `↑↓` | Navigate commits |
| `Enter` | Enter time travel at selected commit |
| `Tab` | Switch to details pane |
| `Esc` | Return to main menu |
| `Ctrl+R` | **REWIND** to selected commit (destructive, reset --hard) |

### File History Browser
| Key | Action |
|-----|--------|
| `↑↓` | Navigate |
| `Tab` | Cycle between commits → files → diff |
| `v` | Visual selection mode |
| `y` | Yank selection (copy) |
| `Esc` | Back to menu |

### Conflict Resolver
| Key | Action |
|-----|--------|
| `↑↓` | Navigate conflicts |
| `Space` | Toggle resolution choice |
| `Tab` | Switch to diff |
| `Enter` | Confirm resolution |
| `Esc` | Back |

### Input Mode
| Key | Action |
|-----|--------|
| `Enter` | Submit input |
| `Esc` | Clear (if text entered) or Close (if empty) |

### Confirmation Dialog
| Key | Action |
|-----|--------|
| `←→` | Select option |
| `Enter` | Confirm selection |
| `Esc` | Cancel |

### Console (Running Operation)
| Key | Action |
|-----|--------|
| `↑↓` | Scroll output |
| `Esc` | Abort operation |

### Preferences
| Key | Action |
|-----|--------|
| `↑↓` | Navigate options |
| `Space` | Toggle boolean option |
| `=` / `-` | Adjust by ±1 minute |
| `+` / `_` | Adjust by ±10 minutes |
| `Esc` | Back |

---

## Shortcut Conflicts

**Current conflicts in menu items:**

| Shortcut | Conflicting Items | Context |
|----------|------------------|---------|
| `p` | `pull_merge` vs `pull_merge_diverged` | Different Timeline states (Behind vs Diverged) - **INTENTIONAL** |
| `p` | `config_preferences` | Different menus (Normal vs Config) - **INTENTIONAL** |
| `l` | `history` vs `time_travel_history` | Different Operation states (Normal vs TimeTraveling) - **INTENTIONAL** |
| `g` | `file_history` vs `time_travel_files_history` | Different Operation states (Normal vs TimeTraveling) - **INTENTIONAL** |
| `1` / `2` | init/clone location selectors | Different modes - **INTENTIONAL** |
| `r` | `reset_discard_changes` vs `time_travel_return` vs `config_remove_remote` | Different contexts - **INTENTIONAL** |
| `d` | `dirty_pull_merge` (only in dirty state) | Single context - **INTENTIONAL** |

**Analysis:** All conflicts are context-sensitive (different menus, states, or modes). No actual runtime conflicts.

---

## Shortcut Distribution

| Shortcut | Used For | Count |
|----------|----------|-------|
| `a` | Add Remote | 1 |
| `b` | Switch Branch | 1 |
| `c` | Clone | 1 |
| `d` | Dirty Pull | 1 |
| `e` | Add Remote (Normal menu) | 1 |
| `f` | Force Push | 2 contexts |
| `g` | File History | 2 contexts |
| `h` | Push | 1 |
| `i` | Initialize | 1 |
| `l` | History | 2 contexts |
| `m` | Commit / Merge back | 2 contexts |
| `p` | Pull / Preferences | 3 contexts |
| `r` | Reset / Return / Remove Remote | 3 contexts |
| `s` | Switch Remote | 1 |
| `t` | Commit and Push | 1 |
| `u` | Toggle Auto Update | 1 |
| `x` | Replace Local | 1 |
| `1`, `2` | Location selection | 4 contexts |
| `Esc` | Back | 1 |

---

## Summary Statistics

- **Total unique menu items:** 34
- **Total unique shortcuts:** 18 letters + Esc + 1-2 numbers + Ctrl+R
- **Context-sensitive conflicts:** 0 (all are intentional, state-based)
- **Global navigation keys:** ↑↓ (or kj), Tab, Enter, Esc
- **Destructive operation key:** Ctrl+R (REWIND in history)

---

**End of Inventory**

Questions for review?
- Are there shortcuts you want to change?
- Are there menu items that seem misplaced?
- Are there unclear hints?
