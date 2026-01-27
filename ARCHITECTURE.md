# TIT Architecture Guide

## Overview

TIT (Git Timeline Interface) is a state-driven terminal UI for git repository management. It follows a clean, event-driven architecture based on Bubble Tea's Model-View-Update pattern.

**Core Principle:** Git state determines UI state. Operations are always safe and abortable.

---

## Five-Axis State Model

Every moment in TIT is described by **5 axes**, checked in priority order:

**Axis 0 (Pre-flight): GitEnvironment** - System prerequisites
- Checks if git/SSH properly installed and configured BEFORE any git state detection
- States: `Ready`, `NeedsSetup`, `MissingGit`, `MissingSSH`
- If not `Ready` → Show setup wizard or fatal error

**Axes 1-4: Repository State** - Git repository state (git.State tuple)

```go
// Application struct has GitEnvironment as separate pre-flight field:
type Application struct {
    gitEnvironment   git.GitEnvironment  // Axis 0: System prerequisites (PRE-FLIGHT)
    gitState       *git.State          // Axes 1-4: Repository state tuple
}

// git.State struct contains 4 repository state axes:
type State struct {
    WorkingTree      git.WorkingTree     // Axis 1: Local file changes
    Timeline         git.Timeline        // Axis 2: Local vs remote comparison
    Operation        git.Operation       // Axis 3: Git operation state
    Remote           git.Remote          // Axis 4: Remote repository presence
}
```

**State Detection Order (Priority):**
1. **GitEnvironment** (Axis 0, PRE-FLIGHT) - Check if git/SSH properly configured
   - If not `git.Ready` → Show setup wizard or fatal error
2. **NotRepo** - Check if .git/ directory exists (Axis 3, FIRST in git state)
3. **DirtyOperation** - Check for ongoing merge/rebase/stash
4. **Operation** - Detect git operation state
5. **WorkingTree** - Detect staged/unstaged changes
6. **Remote** - Check if remote configured
7. **Timeline** - Compare local vs remote (only if Normal + HasRemote + hasCommits)

### State Semantics

**Timeline** represents the **comparison** between local branch and remote tracking branch:
- **InSync:** Local and remote point to same commit
- **Ahead:** Local has commits not on remote (includes CommitsAhead count)
- **Behind:** Remote has commits not on local (includes CommitsBehind count)
- **Diverged:** Both have unique commits (includes both counts)
- **Empty string (""):** Timeline N/A - no comparison possible when:
  - `Remote = NoRemote` (no remote configured)
  - `Operation = TimeTraveling` (detached HEAD, no tracking relationship)
  - No commits exist yet (empty repo)

**Timeline is ONLY detected when:**
```go
if state.Operation == Normal && state.Remote == HasRemote && hasCommits {
    // Detect timeline comparison (includes ahead/behind counts)
    state.Timeline = detectTimeline()
    state.CommitsAhead = ahead
    state.CommitsBehind = behind
} else {
    // Timeline N/A
    state.Timeline = ""
    state.CommitsAhead = 0
    state.CommitsBehind = 0
}
```

**Remote** is a precondition check, not a timeline status:
- `NoRemote`: No remote repository configured
- `HasRemote`: Remote exists (timeline comparison possible)

**Operation** describes the git repository state:
- `Normal`: Ready for operations
- `NotRepo`: Not a git repository (no .git/)
- `Conflicted`: Unresolved merge/rebase/cherry-pick conflicts
- `Merging`: Merge in progress (no conflicts yet)
- `Rebasing`: Rebase in progress (no conflicts yet)
- `DirtyOperation`: Operation interrupted by uncommitted changes (pre-flight check blocks startup)
- `TimeTraveling`: Detached HEAD, exploring commit history (entered via History mode)
- `Rewinding`: Performing time travel merge/return operation

---

## Git Environment & Startup Flow (Axis 0 - Highest Priority)

### GitEnvironment: System Prerequisites

**Purpose:** Detect whether git/SSH are properly installed and configured before entering normal git workflow.

**Environment States** (`git.GitEnvironment` type):
```go
const (
    GitEnvironmentReady      = "ready"        // Git + SSH available, ready for work
    GitEnvironmentNeedsSetup = "needs_setup"  // Git OK, but SSH not configured
    GitEnvironmentMissingGit = "missing_git"  // Git not installed
    GitEnvironmentMissingSSH = "missing_ssh"  // SSH not installed
)
```

**Detection at Startup:**
```
App starts → Check git availability
    ├─ Git not found → GitEnvironmentMissingGit → Show fatal error
    └─ Git found → Check SSH availability
        ├─ SSH not found → GitEnvironmentMissingSSH → Show fatal error
        └─ SSH found → Check if SSH keys configured
            ├─ No keys → GitEnvironmentNeedsSetup → Enter ModeSetupWizard
            └─ Keys found → GitEnvironmentReady → Proceed to normal startup
```

### ModeSetupWizard: First-Time SSH Configuration

**Purpose:** Guide users through SSH key generation and configuration if needed.

**Wizard Steps** (`SetupWizardStep` enum):
1. **SetupStepWelcome** - Welcome message explaining SSH setup
2. **SetupStepPrerequisites** - Verify git/ssh installed (already checked at startup)
3. **SetupStepEmail** - Input email for SSH key comment (e.g., user@example.com)
4. **SetupStepGenerate** - Generate SSH key, start agent, add key to agent
5. **SetupStepDisplayKey** - Show public key with copy button, provider URLs (GitHub/GitLab/Gitea)
6. **SetupStepComplete** - Completion message, return to normal startup

**Application Fields:**
```go
gitEnvironment  git.GitEnvironment  // Current environment state
setupWizardStep SetupWizardStep     // Current step in wizard
setupEmail      string              // Email entered by user
setupKeyCopied  bool                // Whether user copied public key
```

### Startup Flow

**Environment Check:**
```go
func (a *Application) Init() {
    // 1. Check git/SSH availability
    a.gitEnvironment = git.CheckEnvironment()
    
    // 2. If missing prerequisites, show fatal error screen
    if a.gitEnvironment == GitEnvironmentMissingGit {
        a.mode = ModeSetupWizard
        a.setupWizardStep = SetupStepFatalMissingGit
        return
    }
    
    // 3. If needs setup, enter wizard
    if a.gitEnvironment == GitEnvironmentNeedsSetup {
        a.mode = ModeSetupWizard
        a.setupWizardStep = SetupStepWelcome
        return
    }
    
    // 4. Environment ready, proceed to git state detection
    a.detectGitState()
}
```

**Git State Detection:**
```go
func (a *Application) detectGitState() {
    // 1. Check if in git repository
    isRepo, repoPath := git.IsInitializedRepo()
    if !isRepo {
        // Not in repo → Show NotRepo menu (init/clone)
        a.mode = ModeMenu
        a.gitState = &git.State{Operation: git.NotRepo}
        return
    }
    
    // 2. Check for pre-flight blockers (ongoing merge/rebase/stash)
    // If any found, show fatal error → Exit application
    
    // 3. Detect normal git state (5 axes)
    a.gitState, _ = git.DetectState()
    
    // 4. Show menu
    a.mode = ModeMenu
    a.menuItems = a.GenerateMenu()
}
```

**Initialize Caches:**
```go
func (a *Application) Init() {
    // ... earlier steps ...
    
    // Preload history caches in background (async, non-blocking)
    // Contract: History modes disabled until cache ready
    if a.gitState.Operation == git.Normal {
        a.cacheLoadingStarted = true
        go a.preloadHistoryMetadata()
        go a.preloadFileHistoryDiffs()
    }
}
```

**Complete Startup Sequence:**
```
Terminal resize → Quit confirm?
    ↓ (No, normal startup)
Create Application{width, height, theme}
    ↓
Call app.Init()
    ├─ Check GitEnvironment
    │   ├─ Missing git/SSH? → Fatal error + ModeSetupWizard
    │   └─ Needs setup? → ModeSetupWizard (wizard flow)
    │       └─ User completes setup → Return here
    ├─ Detect git state (5 axes)
    │   ├─ Conflicted/Merging/Rebasing? → Fatal error
    │   ├─ DirtyOperation? → Fatal error
    │   └─ Normal? → Proceed
    ├─ Check if in repo
    │   ├─ Not in repo → ModeMenu (NotRepo)
    │   └─ In repo → ModeMenu (normal state)
    └─ Preload caches in background (async)
        └─ Menu items disabled until ready
    ↓
Return to Bubble Tea
    ↓
View() → Render based on current mode
```

### Pre-Flight Blocker Check

**States that block startup (fatal errors):**
```go
// In git/state.go::DetectState()
state, err := git.DetectState()

// Check for blockers (before returning state)
if state.Operation == Conflicted ||
   state.Operation == Merging ||
   state.Operation == Rebasing ||
   state.Operation == DirtyOperation {
    // Startup blocked - show fatal error
    return FatalError("Unresolved merge/rebase/conflicts detected")
}
```

**User must resolve externally:**
```bash
git merge --abort / --continue
git rebase --abort / --continue
git stash pop
```

---

When `Operation = TimeTraveling`, the application tracks additional metadata:

```go
type TimeTravelInfo struct {
    OriginalBranch  string     // Branch we departed from (e.g., "main")
    OriginalHead    string     // Commit hash before time travel started
    CurrentCommit   CommitInfo // Currently checked-out commit (Hash, Subject, Time)
    OriginalStashID string     // If dirty at entry: ID of stashed work ("" if clean entry)
}
```

**Lifecycle:**
1. **Entry:** When user ENTER on commit in History mode
   - Capture `OriginalBranch` and `OriginalHead` before checkout
   - If working tree dirty: stash changes, capture `OriginalStashID`
   - Checkout target commit, set `Operation = TimeTraveling`
   - Populate `CurrentCommit` from detached HEAD state
2. **While Traveling:** User can browse history (jump to different commits) via History mode
   - `CurrentCommit` updates on each jump
   - `OriginalBranch`, `OriginalHead`, `OriginalStashID` remain unchanged
3. **Exit via Merge:** Merge time-travel changes back
   - Merge `CurrentCommit.Hash` into `OriginalBranch` (may have conflicts)
   - Apply any stashed work back (may have conflicts)
4. **Exit via Return:** Discard changes, go back
   - Checkout `OriginalBranch` (at `OriginalHead`)
   - Restore stashed work if it exists

**Loading from detached HEAD:** When TIT starts in TimeTraveling state (`.git/TIT_TIME_TRAVEL` exists), `LoadTimeTravelInfo()` reconstructs `CurrentCommit` by querying git:
- `git rev-parse HEAD` → Hash
- `git log -1 --format=%s` → Subject
- `git log -1 --format=%aI` → Time (parsed as RFC3339)

**Storage in Application:**
```go
type Application struct {
    gitState       git.State
    timeTravelInfo *TimeTravelInfo  // Non-nil only when Operation = TimeTraveling

    // Extracted state structs:
    inputState   InputState   // Input field management (7 fields)
    cacheManager CacheManager // Cache lifecycle (14 fields)
    asyncState   AsyncState   // Async operation state (3 fields)

    // ... other fields (47 total)
}
```

**Application Struct Architecture:**
- **Current state:** 47 fields  
- **Extracted structs:**
  - `InputState` (`internal/app/input_state.go`) - Input field management (7 fields, 148 lines)
  - `CacheManager` (`internal/app/cache_manager.go`) - Cache lifecycle (14 fields, 307 lines)
  - `AsyncState` (`internal/app/async_state.go`) - Async operation state (3 fields, 59 lines)

**InputState Methods:**
- Reset(), SetValue(), GetValue()
- SetCursorPos(), GetCursorPos(), MoveCursorBy()
- InsertAtCursor(), DeleteBeforeCursor(), DeleteAfterCursor()
- SetPrompt(), GetPrompt(), GetAction()
- SetValidationMessage(), ClearValidationMessage(), GetValidationMessage(), HasValidationError()
- SetClearConfirming(), IsClearConfirming(), ToggleClearConfirming()

**CacheManager Methods:**
- Status: IsLoadingStarted(), SetLoadingStarted(), IsMetadataReady(), SetMetadataReady(), IsDiffsReady(), SetDiffsReady()
- Progress: GetMetadataProgress(), SetMetadataProgress(), GetDiffsProgress(), SetDiffsProgress()
- Animation: GetAnimationFrame(), IncrementAnimationFrame()
- Cache: GetMetadata(), SetMetadata(), GetAllMetadata(), GetDiff(), SetDiff(), GetFiles(), SetFiles()
- Invalidation: Invalidate(), InvalidateMetadata(), InvalidateDiffs()
- Bulk: InitMetadataLoading(), InitDiffsLoading(), UpdateMetadataProgress(), UpdateDiffsProgress(), FinalizeMetadata(), FinalizeDiffs()

**Lock Order:** historyMutex → diffMutex (documented, enforced)

**AsyncState Methods:**
- Start(), End(), Abort(), ClearAborted()
- IsActive(), IsAborted(), CanExit(), SetExitAllowed()

**Helper Methods (Application delegates to structs):**
- Async: startAsyncOp(), endAsyncOp(), abortAsyncOp(), clearAsyncAborted(), isAsyncActive(), isAsyncAborted(), canExit(), setExitAllowed()
- Cache: All cache operations now go through a.cacheManager

**Safety invariant:** ESC at any point restores exact original state by restoring original branch and reapplying stash.

---

## Application Structure

### State Struct Extraction

The Application struct has been refactored from a God Object into focused components:

**Current Architecture:**
- **Application struct:** 47 fields (reduced from 72) 
- **Extracted structs:**
  - `InputState` (`internal/app/input_state.go`) - Input field management (7 fields, 148 lines)
  - `CacheManager` (`internal/app/cache_manager.go`) - History cache lifecycle (14 fields, 307 lines)
  - `AsyncState` (`internal/app/async_state.go`) - Async operation state (3 fields, 59 lines)

**SSOT Helper Functions (app.go):**
1. `reloadGitState()` - Centralizes state reload patterns
   - Detects git state, updates application gitState field
   - Used after any git operation
2. `checkForConflicts()` - Centralizes conflict detection
   - Checks for merge/rebase/cherry-pick conflicts
   - Returns conflict file list if found
3. `executeGitOp()` - Standardizes git command execution
   - Runs git command, checks for errors

**Layer Separation:**
- Clean git→app→ui dependency chain
- `internal/app/git_logger.go` implements `git.Logger` interface
- Git package no longer imports UI
- Implementation:
  ```go
  // git/types.go - Logger interface
  type Logger interface {
      Log(message string)
      Warn(message string)
      Error(message string)
  }

  // app/git_logger.go - Implements Logger using UI buffer
  type GitLogger struct{}

  func (l *GitLogger) Log(message string) {
      ui.GetBuffer().Append(message, ui.TypeInfo)
  }
  ```

### Operations Architecture

**Focused Operation Files:**
The operation logic is organized into 9 focused files:
- `op_init.go` - Repository initialization
- `op_clone.go` - Repository cloning
- `op_remote.go` - Remote operations
- `op_commit.go` - Commit operations
- `op_push.go` - Push operations
- `op_pull.go` - Pull operations
- `op_dirty_pull.go` - Dirty pull handling
- `op_merge.go` - Merge operations
- `op_time_travel.go` - Time travel operations

**Benefits:**
- Semantic file naming clearly indicates feature area
- Easy to locate operation logic via `grep -r "cmd" op_*.go`
- No import tracing needed (single package)
- Consistent with Go standard library patterns

---

## Cache Contract: History Always Available

### The Contract

**Principle:** History modes (Commit history, File(s) history) MUST ALWAYS show data instantly. No loading delays, no empty views, no lazy loading.

**Implementation:**

Cache precomputation is **MANDATORY** at:
1. **App startup** - Full scan of all commits before showing menu (async, non-blocking)
2. **After ANY git-changing operation** - Commit, push, pull, merge, time travel merge/return
3. **BEFORE showing completion message** - User never sees empty history after an operation

### Cache Architecture

**Three Independent Caches managed by CacheManager:**

1. **History Metadata Cache**
   - Key: commit hash
   - Value: `*git.CommitDetails` (subject, author, date, message)
   - Access: `a.cacheManager.GetMetadata(hash)`
   - Built by: `preloadHistoryMetadata()` (async goroutine)
   - Used by: ModeHistory (commit list pane)

2. **File History Files Cache**
   - Key: commit hash
   - Value: `[]git.FileInfo` (list of files changed)
   - Access: `a.cacheManager.GetFiles(hash)`
   - Built by: `preloadFileHistoryDiffs()` (async goroutine)
   - Used by: ModeFileHistory (files pane)

3. **File History Diffs Cache**
   - Key: `hash:path:version` (e.g., "abc123:main.go:parent")
   - Value: diff content string
   - Access: `a.cacheManager.GetDiff(key)`
   - Built by: `preloadFileHistoryDiffs()` (async goroutine)
   - Used by: ModeFileHistory (diff pane)

**Thread Safety:** All cache access protected by mutexes (historyMutex → diffMutex)

### Build Rules & Guards

**Startup Guard (app.go::Init):**
```go
// Preload caches only once per app instance
if !shouldRestore && a.gitEnvironment == GitEnvironmentReady {
    a.cacheManager.SetLoadingStarted(true)
    go a.preloadHistoryMetadata()
    go a.preloadFileHistoryDiffs()
}
```

**Post-Operation Guard (git_handlers.go):**
```go
// After any git operation succeeds, rebuild caches
case OpCommit, OpPush, OpPull, OpMerge, ...:
    // Reload git state
    a.gitState, _ = git.DetectState()

    // Rebuild caches via CacheManager
    a.cacheManager.SetLoadingStarted(true)
    go a.preloadHistoryMetadata()
    go a.preloadFileHistoryDiffs()

    // Show completion message
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
}
```

**Cache Status (via CacheManager methods):**
```go
a.cacheManager.IsLoadingStarted()   // True when preload started
a.cacheManager.IsMetadataReady()    // True when metadata cache populated
a.cacheManager.IsDiffsReady()       // True when diffs cache populated
a.cacheManager.GetMetadataProgress() // Current commit processed
a.cacheManager.GetMetadataTotal()    // Total commits to process
```

### UI Feedback During Cache Load

**Menu item state while cache building:**

Normal state (cache ready):
```
🕒 Commit history                        ← Enabled, selectable
📂 File history                          ← Enabled, selectable
```

Building state (cache loading):
```
⏳ Commit history [Building... 12/30]    ← Disabled, shows progress
⏳ File history [Building... 12/30]      ← Disabled, shows progress
```

No data state (cache load failed):
```
⚠️ Commit history [No commits found]     ← Disabled, error state
⚠️ File history [No commits found]       ← Disabled, error state
```

### Invariants (Guaranteed)

1. **No Empty Views:** If `dispatchHistory()` called, data always exists in cache (or disabled menu item prevents dispatch)
2. **No Lazy Loading:** No on-the-fly git queries during rendering (all data pre-cached)
3. **Consistent State:** Cache reflects current git HEAD (rebuilt after every operation)
4. **Fail Fast:** If cache load fails, menu shows disabled state with reason
5. **Read-Only:** Cache data never modified during browsing (immutable snapshots)
6. **Thread-Safe:** Cache access protected by mutexes (historyCacheMutex, diffCacheMutex)

### Cache Lifetime

```
App starts
  ↓
Check GitEnvironment (ready?)
  ↓
Detect git state
  ├─ NotRepo? → Skip cache
  └─ Normal? → Start cache preload (async)
  ↓
Menu shows (items disabled until cache ready)
  ↓
User operates (commit/push/merge/etc)
  ↓
Operation completes, git state changes
  ↓
Cache rebuilds automatically (githandlers.go)
  ↓
Menu becomes active again with fresh data
  ↓
Repeat
```

---

## Three-Layer Event Model

### 1. Input → Application Update

```
tea.KeyMsg / tea.WindowSizeMsg / CustomMsg
    ↓
app.Update(msg)
    ↓
Route to mode handler (app.keyHandlers registry)
    ↓
Mutate Application state
    ↓
Return (model, cmd)
```

**Key Handler Registry:** Built once at app init, cached in `app.keyHandlers`:
```
map[AppMode]map[string]KeyHandler
```

Global handlers (ESC, Ctrl+C) take priority and apply to all modes.

### Keyboard Input Patterns

**Critical: Bubble Tea sends actual characters, not key names**

```go
// ✅ CORRECT - Use actual character or Bubble Tea key string
On("enter", handler)     // Named key
On("tab", handler)       // Named key
On("up", handler)        // Named key
On(" ", handler)         // SPACE character, not "space"!
On("ctrl+c", handler)    // Special combo notation

// ❌ WRONG
On("space", handler)     // Bubble Tea sends " " not "space"
On("return", handler)    // Bubble Tea sends "enter" not "return"
```

**Registration pattern** (`internal/app/app.go`):
```go
ModeMenu: NewModeHandlers().
    On("j", a.handleMenuDown).
    On("k", a.handleMenuUp).
    On("enter", a.handleMenuEnter).
    Build(),

ModeConflictResolve: NewModeHandlers().
    On("up", a.handleConflictUp).
    On("down", a.handleConflictDown).
    On("tab", a.handleConflictTab).
    On(" ", a.handleConflictSpace).      // ← Space character
    On("enter", a.handleConflictEnter).
    Build(),
```

**Why this matters:**
- Bubble Tea's `msg.String()` returns the actual character (`" "`) or key name (`"enter"`, `"tab"`)
- If you register `"space"`, the handler never fires (Bubble Tea sends `" "`)
- Discovered by checking `msg.String()` in handler and comparing to registry key
- This caused SPACE key not to fire in conflict resolver until fixed

### 2. State Mutation → Async Operations

All blocking git operations run in goroutines (worker threads):

```
User presses Enter
    ↓
Handler sets asyncOperationActive = true
    ↓
Handler returns tea.Cmd that spawns goroutine
    ↓
Worker executes git command, streams output to OutputBuffer
    ↓
Worker returns GitOperationMsg{Step, Success, Path, Error}
    ↓
app.Update(GitOperationMsg) reloads git state
    ↓
View() re-renders based on new state
```

**Thread Safety Rules:**
- ❌ Never mutate Application from goroutine
- ✅ Use closures to capture state before spawning goroutine
- ✅ Return immutable messages (tea.Msg) from workers
- ✅ Use OutputBuffer for streaming (thread-safe)

### 3. State → UI Rendering

```
Current (WorkingTree, Timeline, Operation, Remote)
    ↓
GenerateMenu() → []MenuItem
    ↓
View() renders based on current AppMode
    ↓
RenderLayout() wraps with header/footer/layout
    ↓
Terminal displays result
```

---

## Application Modes (AppMode) - 14 Total

| Mode | Purpose | Input Handler | Async | Notes |
|------|---------|---|---|---|
| **ModeMenu** | Main action menu (state-driven) | Menu navigation (j/k/enter) | No | Init/Clone, Commit/Amend, Push/Pull, History browsing |
| **ModeInput** | Generic text input | Cursor nav + character input | No | **DEPRECATED** - being phased out in favor of dedicated modes |
| **ModeConsole** | Streaming git command output | Console scroll (↑↓/PgUp/PgDn), ESC abort | Yes | Shows progress indicator during async operations |
| **ModeConfirmation** | Yes/No confirmation dialog | left/right/h/l/y/n/enter | No | For destructive operations (nested repo, force push, etc) |
| **ModeHistory** | Commit history browser (2-pane) | ↑↓ nav, TAB pane, ENTER time travel, ESC menu | No | Commits (left, 24 chars) + Details (right) |
| **ModeConflictResolve** | N-column parallel conflict resolution | ↑↓ scroll, TAB cycle panes, SPACE mark, ENTER apply | No | Used for merge, dirty pull, time travel conflicts |
| **ModeInitializeLocation** | Choose init location (cwd/subdir) | Menu selection | No | First step of init flow |
| **ModeInitializeBranches** | Dual input for canon + working branch | Text input (canon pre-filled 'main') | No | Second step of init flow |
| **ModeCloneURL** | Input clone URL | Single text input with validation | No | First step of clone flow |
| **ModeCloneLocation** | Choose clone location (cwd/subdir) | Menu selection | No | Second step of clone flow |
| **ModeClone** | Clone operation streaming output | Console scroll, ESC abort | Yes | Shows `git clone` progress |
| **ModeSelectBranch** | Choose canon branch from cloned repo | Menu selection | No | Final step of clone flow |
| **ModeFileHistory** | File(s) history browser (3-pane) | ↑↓ nav, TAB cycle, V visual, Y copy, ESC | No | Commits (24 chars) + Files (remaining) + Diff |
| **ModeSetupWizard** | Git environment setup wizard | Mode-specific handlers | No | SSH key generation, agent config (runs once at startup if needed) |

**Total: 14 modes** (including deprecated ModeInput still in use)

---

## Menu System

### MenuGenerator Pattern & The Contract Principle

**CRITICAL PRINCIPLE: Menu = Contract**

If an action appears in the menu, it MUST succeed. Never show operations that could:
- Fail due to git state
- Leave repo in dangling/incomplete state (merge/rebase/stash in progress)
- Require manual user cleanup

**TIT Startup Check (Before Any Menu):**

If git state detection finds any of these:
- `Conflicted` (merge/rebase/conflict in progress)
- `Merging` (merge in progress, no conflicts)
- `Rebasing` (rebase in progress)
- `DirtyOperation` (stash mid-operation)

→ TIT shows fatal error screen and **refuses to start**. User must resolve externally:
```bash
git merge --continue / --abort
git rebase --continue / --abort
git stash pop
```

**Why:** These are pre-existing abnormal states, not states TIT creates or manages. TIT operates only on clean/normal repositories.

**Valid MenuGenerators (Only for Normal Startups):**

```go
type MenuGenerator func(*Application) []MenuItem

menuGenerators := map[git.Operation]MenuGenerator{
    git.NotRepo:        (*Application).menuNotRepo,
    git.Normal:         (*Application).menuNormal,
    git.TimeTraveling:  (*Application).menuTimeTraveling,  // Only entered via History mode
}

// These states cause startup failure, never reach menuGenerators:
// git.Conflicted, git.Merging, git.Rebasing, git.DirtyOperation
```

### MenuItem SSOT System

All menu items defined in single source of truth (`internal/app/menuitems.go`):

```go
var MenuItems = map[string]MenuItem{
    "commit": {
        ID:       "commit",
        Shortcut: "m",
        Emoji:    "📝",
        Label:    "Commit changes",
        Hint:     "Create commit from staged changes",
        Enabled:  true,
    },
    // ... all menu items defined here
}
```

**Benefits:**
- All shortcuts globally unique (no conflicts)
- Single source for labels, hints, emoji
- Hints displayed in footer (not in menu)
- Easy to audit and modify without touching menu generators

**Menu Item SSOT Guarantees:**
- Shortcut conflicts detected at build time (in `app.go` init)
- Emoji validation (no narrow emojis per SESSION-LOG.md rules)
- Hints stored in SSOT but rendered in footer hint area
- All text centralized: no hardcoded labels in menu.go

Menu generators retrieve items via `GetMenuItem(id)`:
```go
Item("commit").Shortcut("m").Label("...").Build()  // ❌ OLD
GetMenuItem("commit")                              // ✅ NEW
```

### Menu Rendering Flow

```
GenerateMenu() → []MenuItem (ID, Shortcut, Emoji, Label only)
    ↓
RenderMenu() displays with 2 columns:
    - Left: [Shortcut] emoji Label
    - Right: (empty - hints moved to footer)
    ↓
On menu selection change:
    - app.footerHint = GetMenuItem(selected).Hint
    ↓
Layout() displays footer with current hint
```

### Generators in `internal/app/menu.go`

**Valid startups only:**
- `menuNotRepo()` - Init/Clone (not in repo)
- `menuNormal()` - Full menu (normal state)
  - `menuWorkingTree()` - Commit (when Dirty)
  - `menuTimeline()` - Push/Pull based on Timeline
  - `menuHistory()` - Commit history browser (time travel entry point)
- `menuTimeTraveling()` - Browse history, Merge back, Return (only when `Operation = TimeTraveling`)

**Never called (pre-flight check blocks startup):**
- ~~`menuConflicted()`~~ - Startup prevents any menu if Conflicted
- ~~`menuOperation()`~~ - Startup prevents any menu if Merging/Rebasing
- ~~`menuDirtyOperation()`~~ - Startup prevents any menu if DirtyOperation in progress

### State Display System (StateDescriptions SSOT)

**Purpose:** Centralize all git state descriptions for consistent, translatable UI messages.

**Implementation** (`internal/app/stateinfo.go` uses `StateDescriptions` SSOT):

```go
// messages.go - State descriptions SSOT
var StateDescriptions = map[string]string{
    "working_tree_clean":  "Your files match the remote.",
    "working_tree_dirty":  "You have uncommitted changes.",
    "timeline_no_remote":  "No remote repository configured.",
    "timeline_in_sync":    "Local and remote are in sync.",
    "timeline_ahead":      "You have %d unsynced commit(s).",
    "timeline_behind":     "The remote has %d new commit(s).",
    "timeline_diverged":   "Both have new commits. Ahead %d, Behind %d.",
}

// stateinfo.go - Uses SSOT
type StateInfo struct {
    Label       string
    Emoji       string
    Color       string
    Description func(ahead, behind int) string  // Lookup from SSOT
}

BuildStateInfo(theme) returns:
- WorkingTree map: Clean/Dirty → StateInfo with description from StateDescriptions
- Timeline map: InSync/Ahead/Behind/Diverged → StateInfo with description from StateDescriptions
- Operation map: Normal/TimeTraveling/Conflicted/etc → StateInfo with description from StateDescriptions
```

**Header Rendering (9-Line Layout):**

Current header structure (RenderHeaderInfo in internal/ui/header.go):
```
─── 2-COLUMN SECTION (80/20 split) ───
Row 1: 📁 CWD (80% left)                | 🟢 OPERATION (20% right, right-aligned)
Row 2: 🔗 Remote/Status (80% left)      | 🌿 BRANCH (20% right, right-aligned)

─── FULL-WIDTH SECTION ───
Row 3: ──────────────────────── (separator line)
Row 4: ✅ WORKING TREE LABEL (bold, colored)
Row 5: Description of working tree state (indented)
Row 6: [Additional description line if needed]
Row 7: 🔗 TIMELINE LABEL (bold, colored)
Row 8: Description of timeline state (indented)
Row 9: [Additional description line if needed]
```

**Actual Height:** HeaderHeight = 9 content rows (with padding: 11 total lines including top/bottom margins)

**Normal Operation Example (Operation = Normal, Timeline = InSync):**
```
📁 /Users/jreng/Documents/Poems/tit        🟢 READY
🔗 github.com/user/repo                    🌿 main
─────────────────────────────────────────────────────────
✅ CLEAN
   Your files match the remote.
🔗 IN SYNC
   Local and remote are in sync.
```

**Time Traveling Example (Operation = TimeTraveling):**
```
📁 /Users/jreng/Documents/Poems/tit        🌀 TIME TRAVEL
🔗 github.com/user/repo                    🔀 Detached HEAD
─────────────────────────────────────────────────────────
✅ CLEAN
   Your files match the remote.
📌 COMMIT c53233c
   Mon, 7 Jan 2026 04:45:12
```

**No Remote Example (Remote = NoRemote):**
```
📁 /Users/jreng/Documents/Poems/tit        🟢 READY
🔌 NO REMOTE                               🌿 main
─────────────────────────────────────────────────────────
✅ CLEAN
   Your files match the remote.
🔌 N/A
   No remote configured.
```

**Design Details:**
- **HeaderState struct** (ui/header.go): 18 fields capturing all header display state
- **2-column section** (rows 1-2): Fixed 80/20 split with right-aligned operation/branch
- **Separator** (row 3): Full-width dashes, themed color
- **Description lines**: Indented with `EmojiColumnWidth = 3` spaces for alignment
- **Dynamic widths**: All calculations use DynamicSizing for responsive layout
- **Color system**: Operation/WorkingTree/Timeline colors from theme (SSOT in ui/theme.go)

---

---

## Responsive Layout System

### DynamicSizing: Responsive Terminal Layout

**Purpose:** Calculate exact dimensions from terminal size, enabling full-terminal responsive layout.

**Threshold Constants (SSOT in ui/sizing.go):**
```go
const (
    MinWidth         = 69      // Minimum usable terminal width
    MinHeight        = 19      // Minimum usable terminal height
    HeaderHeight     = 9       // Fixed header height
    FooterHeight     = 1       // Fixed footer height
    MinContentHeight = 4       // Minimum content area height
    HorizontalMargin = 2       // Left + right margins
    BannerWidth      = 30      // Width of optional banner
)
```

**DynamicSizing Struct:**
```go
type DynamicSizing struct {
    TerminalWidth     int  // Full terminal width
    TerminalHeight    int  // Full terminal height
    ContentHeight     int  // Available height for content (TerminalHeight - Header - Footer)
    ContentInnerWidth int  // Available width for content (TerminalWidth - 2*Margins)
    HeaderInnerWidth  int  // Available width for header
    FooterInnerWidth  int  // Available width for footer
    MenuColumnWidth   int  // Left column when banner displayed
    IsTooSmall        bool // true if width < MinWidth OR height < MinHeight
}
```

**Calculation Logic:**
```go
func CalculateDynamicSizing(termWidth, termHeight int) DynamicSizing {
    isTooSmall := termWidth < MinWidth || termHeight < MinHeight
    
    contentHeight := termHeight - HeaderHeight - FooterHeight
    if contentHeight < MinContentHeight {
        contentHeight = MinContentHeight
    }
    
    innerWidth := termWidth - (HorizontalMargin * 2)
    
    return DynamicSizing{
        TerminalWidth:     termWidth,
        TerminalHeight:    termHeight,
        ContentHeight:     contentHeight,
        ContentInnerWidth: innerWidth,
        IsTooSmall:        isTooSmall,
    }
}
```

**Too Small Handler:**
If terminal is < 69×19:
- `IsTooSmall = true`
- View renders single centered message: "Terminal too small (69×19 minimum)"
- All UI rendering blocked until terminal resized

### RenderReactiveLayout: Full-Terminal Composition

**Function:** `ui.RenderReactiveLayout(sizing, theme, header, content, footer)`

**Layout Structure:**
```
┌────────────────────────────────────────────────┐ ← Terminal top (y=0)
│ HEADER (9 lines, full width)                   │
│ - 2-column section: CWD + Remote | Op + Branch │
│ - Separator line                              │
│ - WorkingTree status + description             │
│ - Timeline status + description                │
├────────────────────────────────────────────────┤
│ CONTENT (variable height)                      │
│ - Menu, History, Input, etc. (dynamic height)  │
├────────────────────────────────────────────────┤
│ FOOTER (1 line)                                │
│ - Keyboard hints or messages (centered)        │
└────────────────────────────────────────────────┘ ← Terminal bottom (y=termHeight)
```

**Height Calculation:**
- Header: Fixed 9 lines
- Footer: Fixed 1 line
- Content: `TerminalHeight - 9 - 1 = TerminalHeight - 10`

**Assembly Pattern (lipgloss):**
```go
// 1. Render each section with exact dimensions
headerSection := lipgloss.NewStyle().
    Width(sizing.TerminalWidth).
    Height(HeaderHeight).
    Render(headerText)

contentSection := lipgloss.NewStyle().
    Width(sizing.TerminalWidth).
    Height(contentHeight).
    Render(contentText)

footerSection := lipgloss.NewStyle().
    Width(sizing.TerminalWidth).
    Height(FooterHeight).
    Render(footerText)

// 2. Join vertically
combined := lipgloss.JoinVertical(lipgloss.Left, 
    headerSection, contentSection, footerSection)

// 3. Place in exact terminal dimensions
result := lipgloss.Place(
    sizing.TerminalWidth,
    sizing.TerminalHeight,
    lipgloss.Left, lipgloss.Top,
    combined)
```

**Key Design:**
- No padding or gaps between sections (borders touch)
- Footer always sticks to bottom (via lipgloss.Place)
- Content area grows/shrinks with terminal resize
- Header and footer remain at fixed heights

### Legacy Constants (Deprecated)

```go
// Old fixed sizing - DEPRECATED, DO NOT USE
const (
    ContentInnerWidth = 76  // Legacy hardcoded width
    ContentHeight     = 24  // Legacy hardcoded height
)
```

**Status:** These constants remain for backward compatibility but should NOT be used in new code. Always use `sizing.ContentInnerWidth` and `sizing.ContentHeight` instead.

**Migration Path:**
- DynamicSizing implemented and active
- All usages migrated to DynamicSizing  
- Legacy constants remain for backward compatibility

---

## Confirmation Dialog System

```go
// Dialog messages for confirmation dialogs
var DialogMessages = map[string][2]string{
    "nested_repo": {
        "Nested Repository Detected",
        "The directory '%s' is inside another git repository...",
    },
    // Add more as needed: "force_push", "hard_reset", etc.
}

// Dialog routing in confirmationhandlers.go:
func setupConfirmation(action string) {
    if msg, ok := DialogMessages[action]; ok {
        a.confirmationTitle = msg[0]
        a.confirmationExplanation = msg[1]
    }
}
```

---

## Conflict Resolver System (ModeConflictResolve)

**The most complex and reusable UI component in TIT.** Used for:
- Dirty pull (LOCAL vs REMOTE vs INCOMING)
- Time travel conflicts (CURRENT vs PAST)
- Pull merge conflicts (LOCAL vs REMOTE)
- Any N-way file comparison + resolution

### Architecture: Generic N-Column Model

**Layout Structure:**
```
┌─────────────────────────────────────────────────────────┐
│ Top Row: N file lists (shared selection across columns)│
│ ┌────────────────┬────────────────┬────────────────┐   │
│ │   LOCAL        │   REMOTE       │   INCOMING     │   │
│ │ [✓] main.go    │ [ ] main.go    │ [ ] main.go    │   │
│ │ [ ] README.md  │ [✓] README.md  │ [ ] README.md  │   │
│ │ [✓] config.yaml│ [ ] config.yaml│ [ ] config.yaml│   │
│ └────────────────┴────────────────┴────────────────┘   │
├─────────────────────────────────────────────────────────┤
│ Bottom Row: N content panes (independent scrolling)    │
│ ┌────────────────┬────────────────┬────────────────┐   │
│ │  1 package main│  1 package main│  1 package main│   │
│ │  2             │  2             │  2             │   │
│ │  3 import "fmt"│  3 import "log"│  3 import "os" │   │
│ │  4             │  4             │  4             │   │
│ │  5 func main() │  5 func main() │  5 func main() │   │
│ └────────────────┴────────────────┴────────────────┘   │
│ ↑↓ scroll | TAB switch | SPACE mark | ENTER apply     │
└─────────────────────────────────────────────────────────┘
```

### State Management

**ConflictResolveState** (`internal/app/conflictstate.go`):
```go
type ConflictResolveState struct {
    Operation         string              // "dirty_pull", "time_travel", etc.
    Files            []ConflictFileGeneric // All conflicted files
    SelectedFileIndex int                  // Shared file selection (top row)
    FocusedPane      int                   // Which pane has focus (0...2N-1)
    NumColumns       int                   // Number of version columns (2 or 3)
    ColumnLabels     []string              // ["LOCAL", "REMOTE", "INCOMING"]
    ScrollOffsets    []int                 // Per-column scroll position (bottom row)
    LineCursors      []int                 // Per-column line cursor (bottom row)
}

type ConflictFileGeneric struct {
    Path     string   // File path
    Versions []string // Content for each column (N versions)
    Chosen   int      // Which column is chosen (0-based, radio button)
}
```

### Component Hierarchy

**RenderConflictResolveGeneric()** (`internal/ui/conflictresolver.go`)
- Top row: N × ListPane (file lists with checkboxes)
- Bottom row: N × renderGenericContentPane (code viewers)
- Status bar: buildGenericConflictStatusBar (keyboard hints)

**ListPane** (`internal/ui/listpane.go`)
- Reusable list component with:
  - Title (colorized, centered)
  - Scrollable items with checkbox + filename
  - Focus-based border color (#2C4144 → #8CC9D9)
  - Shared selection highlight across all columns

**DiffPane** (`internal/ui/diffpane.go`)
- Advanced diff viewer for future enhancement
- Features: syntax highlighting, visual mode, copy mode
- Currently not used by conflict resolver (uses simpler renderGenericContentPane)

### Navigation Model

**Pane Indexing:**
- Top row: panes 0 to N-1 (file lists)
- Bottom row: panes N to 2N-1 (content)
- Example (2-column): 0=LOCAL list, 1=REMOTE list, 2=LOCAL content, 3=REMOTE content

**Keyboard Handlers** (`internal/app/conflicthandlers.go`):

| Key | Top Row (File Lists) | Bottom Row (Content) |
|-----|----------------------|----------------------|
| ↑ | Move selection up (shared across all columns) | Scroll content up (independent per pane) |
| ↓ | Move selection down (shared) | Scroll content down (independent) |
| TAB | Cycle: pane 0 → 1 → ... → N-1 → N → ... → 2N-1 → wrap | Same |
| SPACE | Mark file in focused column (radio button - one choice per file) | No action |
| ENTER | Apply resolution choices | Apply resolution choices |
| ESC | Abort conflict resolution | Abort conflict resolution |

**Focus Feedback:**
- Border color: Unfocused (#2C4144) → Focused (#8CC9D9)
- Status bar: Shows active pane index and operation type
- Footer hint: Displays marking feedback ("Marked: file.go → column 1")

### Radio Button Marking

**Exclusive Selection:** Each file must have exactly ONE column marked.

```go
// User presses SPACE on file in column 1
if file.Chosen == focusedPane {
    // Already marked here → do nothing (show hint)
    return
}
// Mark this column, unmarks other columns automatically (Chosen field)
file.Chosen = focusedPane
```

**Visual Feedback:**
- `[✓]` = This column chosen
- `[ ]` = Other columns not chosen
- Checkbox state updates in ALL file lists simultaneously

### Width Calculation Strategy

**Problem:** N panes must fit exactly in terminal width with borders.

**Solution:**
```go
baseColumnWidth := width / numColumns
remainder := width % numColumns

// Distribute remainder to rightmost columns
for col := 0; col < numColumns; col++ {
    columnWidth := baseColumnWidth
    if col >= numColumns - remainder {
        columnWidth++ // Last columns get +1 if needed
    }
}
```

**Border Rendering:**
- Each pane draws ALL FOUR borders (lipgloss.NormalBorder)
- lipgloss.JoinHorizontal() places borders side-by-side
- Borders "touch" at seams but this is correct (not artifacts)
- Focus changes border color, making active pane clearly visible

### Theme Colors

**Operation colors** (`internal/ui/theme.go`):
```toml
operationReady = "#4ECB71"          # Emerald green (READY state)
operationNotRepo = "#FC704C"        # preciousPersimmon (NOT REPO / errors)
operationTimeTravel = "#F2AB53"     # Safflower (TIME TRAVEL - warm orange)
operationConflicted = "#FC704C"     # preciousPersimmon (conflicts)
operationMerging = "#00C8D8"        # blueBikini (merge in progress)
operationRebasing = "#00C8D8"       # blueBikini (rebase in progress)
operationDirtyOp = "#FC704C"        # preciousPersimmon (dirty operation)
```

**Conflict-specific colors** (`internal/ui/theme.go`):
```toml
conflictPaneUnfocusedBorder = "#2C4144"  # Dark teal (inactive)
conflictPaneFocusedBorder = "#8CC9D9"    # Bright cyan (active)
conflictPaneTitleColor = "#8CC9D9"       # Pane header text
conflictSelectionForeground = "#090D12"  # Checkbox text
conflictSelectionBackground = "#7EB8C5"  # Checkbox background
```

### Reusability Pattern

**Same component, different contexts:**

```go
// Dirty pull (3 versions)
state := &ConflictResolveState{
    NumColumns: 3,
    ColumnLabels: []string{"LOCAL", "REMOTE", "INCOMING"},
    Files: /* files with 3 versions each */
}

// Time travel (2 versions)
state := &ConflictResolveState{
    NumColumns: 2,
    ColumnLabels: []string{"CURRENT", "PAST"},
    Files: /* files with 2 versions each */
}

// Pull merge conflict (2 versions)
state := &ConflictResolveState{
    NumColumns: 2,
    ColumnLabels: []string{"LOCAL", "REMOTE"},
    Files: /* conflicted files */
}
```

### Integration Points

**Entry:** Menu item → `dispatchDirtyPull()` / `dispatchTimeTravel()` / etc.
- Sets `a.mode = ModeConflictResolve`
- Initializes `a.conflictResolveState` with appropriate data
- Returns to Update() → View() renders conflict UI

**Exit:** User presses ENTER → `handleConflictEnter()`
- Collects all `file.Chosen` values
- Applies resolution (copy chosen version to working tree)
- Runs git commands to complete operation
- Returns to ModeMenu with updated git state

**Abort:** User presses ESC → `handleConflictEsc()`
- Discards all choices
- Returns to ModeConsole (shows abort command output)
- User presses ESC again → Returns to ModeMenu without applying changes

### Generic Conflict Resolver Setup Pattern

**Function:** `setupConflictResolver(operation, columnLabels)` (`internal/app/githandlers.go`)

All conflict-resolving operations use the same parameterized setup function:

```go
func (a *Application) setupConflictResolver(
    operation string,              // "pull_merge", "dirty_pull_changeset_apply", "cherry_pick"
    columnLabels []string,         // ["BASE", "LOCAL (yours)", "REMOTE (theirs)"]
) (tea.Model, tea.Cmd)
```

**Usage:**

```go
// Pull merge conflicts (3-way: BASE, LOCAL, REMOTE)
return a.setupConflictResolver("pull_merge", 
    []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"})

// Dirty pull conflicts (same labels)
return a.setupConflictResolver("dirty_pull_changeset_apply",
    []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"})

// Cherry-pick conflicts (2-way: LOCAL, INCOMING)
return a.setupConflictResolver("cherry_pick",
    []string{"LOCAL (current)", "INCOMING (cherry-pick)"})
```

**Advantages:**
1. **Single source of truth** - One function handles all conflict scenarios
2. **Reduced duplication** - No copy-paste between pull/dirty-pull/cherry-pick
3. **Extensible** - New conflict operations only need to call setupConflictResolver with appropriate labels
4. **Consistent behavior** - All conflict resolvers behave identically (file detection, version loading, routing)

**Implementation Detail:**
- Detects conflicted files via `git.ListConflictedFiles()`
- Loads 3-way git versions (stages 1/2/3) for each file
- Populates `ConflictResolveState` with parameterized operation name and column labels
- Routes to `ModeConflictResolve` with handlers delegating on operation type

**Handler Routing:**

Both `handleConflictEnter()` and `handleConflictEsc()` check operation name:
```go
if app.conflictResolveState.Operation == "pull_merge" {
    return app.cmdFinalizePullMerge()  // or cmdAbortMerge()
} else if app.conflictResolveState.Operation == "dirty_pull_changeset_apply" {
    return app.cmdDirtyPullApplySnapshot()  // or cmdAbortDirtyPull()
} else if app.conflictResolveState.Operation == "cherry_pick" {
    return app.cmdFinalizeCherryPick()  // or cmdAbortCherryPick()
}
```

**Pull Merge Example:**

1. **Finalize path (ENTER):**
   - Handler routes to `cmdFinalizePullMerge()`
   - Stages all resolved files: `git add -A`
   - Commits the merge: `git commit -m "Merge commit"`
   - Returns `GitOperationMsg{Step: OpFinalizePullMerge}`
   - Handler reloads git state → displays completion message

2. **Abort path (ESC):**
   - Handler routes to `cmdAbortMerge()`
   - Aborts the merge: `git merge --abort`
   - Resets working tree: `git reset --hard`
   - Returns `GitOperationMsg{Step: OpAbortMerge}`
   - Handler reloads git state → displays abort message

3. **State Routing in `githandlers.go`:**
   ```go
   case OpFinalizePullMerge:
       state, _ := git.DetectState()
       a.gitState = state  // Reload state after merge
       buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
       a.asyncOperationActive = false
       a.conflictResolveState = nil
   
   case OpAbortMerge:
       state, _ := git.DetectState()
       a.gitState = state  // Reload state after abort
       buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
       a.asyncOperationActive = false
       a.conflictResolveState = nil
   ```

### Critical Design Decisions

**Q: Why not use lipgloss for border-free joining?**
A: Each pane needs borders for visual separation. Full borders + JoinHorizontal is the correct pattern (matches old-tit).

**Q: Why radio buttons instead of checkboxes?**
A: Conflict resolution requires choosing ONE version per file. Radio button enforces this constraint.

**Q: Why shared selection in top row but independent scrolling in bottom row?**
A: User needs to compare the SAME file across all columns. Shared selection keeps all columns synchronized. Bottom row needs independent scrolling for long files.

**Q: Why not use DiffPane for content rendering?**
A: DiffPane is overkill for basic conflict resolution. renderGenericContentPane is simpler (line numbers + highlighting). DiffPane ready for advanced features.

---

## Multi-Pane Content Component Pattern

**Purpose:** Standard pattern for building complex content views with multiple panes, context-sensitive status bars, and focus management. Used by History, Conflict Resolver, and File History modes.

### Core Pattern Overview

All multi-pane components follow this proven structure:

```
┌─────────────────────────────────────────┐
│ Top Row: One or more panes side-by-side│
├─────────────────────────────────────────┤
│ Bottom Row: One or more panes           │
├─────────────────────────────────────────┤
│ Status Bar: Context-sensitive shortcuts │
└─────────────────────────────────────────┘
```

### Height Calculations (CRITICAL - Exact Formula)

**All multi-pane components MUST use this exact calculation:**

```go
// Return height - 2 lines (wrapper will add border(2))
// Layout: topRow + bottomRow + status = height - 2
// Available for panes: (height - 2) - status(1) = height - 3
// But lipgloss adds extra padding, so reduce by 4 more
totalPaneHeight := height - 7
topRowHeight := totalPaneHeight / 3
bottomRowHeight := totalPaneHeight - topRowHeight

// Adjust: add 2 to top row, reduce from bottom row
topRowHeight += 2
bottomRowHeight -= 2
```

**Why this specific formula:**
- `height - 7`: Accounts for border(2) + status(1) + lipgloss padding(4)
- `1/3` split: Top row gets 1/3, bottom row gets 2/3
- `+2/-2` adjustment: Fine-tune to prevent gaps/overflow
- Proven in ConflictResolver, copied to FileHistory

### Width Calculations

**Two patterns based on component needs:**

#### Pattern A: Fixed + Remainder (History, FileHistory)
```go
// Commits pane: fixed 24 chars (fits "07-Jan 02:11 957f977")
commitPaneWidth := 24
detailsPaneWidth := width - commitPaneWidth  // No gap, borders touch
```

#### Pattern B: Equal Distribution (ConflictResolver)
```go
// N columns share width equally
baseColumnWidth := width / numColumns
remainder := width % numColumns

// Distribute remainder to rightmost columns
for col := 0; col < numColumns; col++ {
    columnWidth := baseColumnWidth
    if col >= numColumns - remainder {
        columnWidth++  // Last columns get +1 if needed
    }
}
```

### Assembly Pattern

**CRITICAL: Use lipgloss.JoinHorizontal + direct string concatenation**

```go
// Step 1: Render all panes
leftPane := renderLeftPane(state, theme, leftWidth, topRowHeight)
rightPane := renderRightPane(state, theme, rightWidth, topRowHeight)
bottomPane := renderBottomPane(state, theme, width, bottomRowHeight)

// Step 2: Join top row panes horizontally (borders touch)
topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

// Step 3: Build status bar (context-sensitive)
statusBar := buildStatusBar(state.FocusedPane, width, theme)

// Step 4: Assemble with direct string concatenation (NO gaps)
return topRow + "\n" + bottomPane + "\n" + statusBar
```

**Why this pattern:**
- `lipgloss.JoinHorizontal`: Handles side-by-side panes correctly
- Direct `"\n"` concatenation: No gaps, consistent with wrapper expectations
- Each pane includes ALL borders (lipgloss makes them touch seamlessly)

### Focus Management Pattern

**State tracking:**
```go
type YourState struct {
    FocusedPane  YourPaneEnum  // Which pane has focus
    // ... pane-specific scroll/cursor fields
}

type YourPaneEnum int
const (
    PaneLeft YourPaneEnum = iota
    PaneRight
    PaneBottom
)
```

**TAB key cycling:**
```go
func handleYourModeTab(app *Application) (tea.Model, tea.Cmd) {
    // Cycle through all panes
    numPanes := 3  // Or whatever your mode has
    app.yourState.FocusedPane = (app.yourState.FocusedPane + 1) % numPanes
    return app, nil
}
```

**Focus-based rendering:**
```go
// In pane renderer
isActive := (state.FocusedPane == PaneLeft)
borderColor := theme.ConflictPaneUnfocusedBorder
if isActive {
    borderColor = theme.ConflictPaneFocusedBorder
}
```

### Context-Sensitive Status Bar Pattern

**Status bar switches based on focused pane:**

```go
// In main render function
var statusBar string
if state.FocusedPane == PaneSpecial {
    statusBar = buildSpecialStatusBar(state, width, theme)
} else {
    statusBar = buildNormalStatusBar(state.FocusedPane, width, theme)
}
```

**Normal mode example:**
```go
func buildNormalStatusBar(focusedPane YourPaneEnum, width int, theme Theme) string {
    parts := []string{
        shortcutStyle.Render("↑↓") + descStyle.Render(" navigate"),
        shortcutStyle.Render("TAB") + descStyle.Render(" cycle panes"),
        shortcutStyle.Render("ESC") + descStyle.Render(" back"),
    }
    statusBar := strings.Join(parts, descStyle.Render("  "))

    // Center the status bar
    statusWidth := lipgloss.Width(statusBar)
    leftPad := (width - statusWidth) / 2
    rightPad := width - statusWidth - leftPad
    return strings.Repeat(" ", leftPad) + statusBar + strings.Repeat(" ", rightPad)
}
```

**Special mode example (FileHistory VISUAL mode):**
```go
func buildVisualStatusBar(width int, theme Theme) string {
    visualStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.MainBackgroundColor)).
        Background(lipgloss.Color(theme.AccentTextColor)).
        Bold(true)

    parts := []string{
        visualStyle.Render("VISUAL"),  // Inverted badge
        shortcutStyle.Render("↑↓") + descStyle.Render(" select"),
        shortcutStyle.Render("Y") + descStyle.Render(" copy"),
        shortcutStyle.Render("ESC") + descStyle.Render(" back"),
    }
    return strings.Join(parts, descStyle.Render("  "))  // Left-aligned, no padding
}
```

### Implementation Examples

#### Example 1: History Mode (2-Pane Side-by-Side)

**Layout:**
```
┌────────────┬─────────────────────────────────┐
│  Commits   │  Details                        │
│  List      │  (Author, Date, Message)        │
│  (24 wide) │  (Remaining width)              │
└────────────┴─────────────────────────────────┘
 ↑↓ navigate | TAB switch pane | ESC back
```

**Key characteristics:**
- Fixed 24-char commits pane (fits "07-Jan 02:11 957f977")
- Details pane takes remaining width
- Single status bar (no mode switching)
- Simple 2-pane focus cycle

#### Example 2: Conflict Resolver (N-Column)

**Layout:**
```
┌─────────────┬─────────────┬─────────────┐
│   LOCAL     │   REMOTE    │  INCOMING   │  Top row
│ [✓] file.go │ [ ] file.go │ [ ] file.go │  (file lists)
└─────────────┴─────────────┴─────────────┘
┌─────────────┬─────────────┬─────────────┐
│  1 package  │  1 package  │  1 package  │  Bottom row
│  2 main     │  2 main     │  2 main     │  (content)
└─────────────┴─────────────┴─────────────┘
 ↑↓ scroll | TAB cycle | SPACE mark | ENTER apply
```

**Key characteristics:**
- N columns distributed equally (width / numColumns)
- Top row: N file lists (shared selection)
- Bottom row: N content panes (independent scrolling)
- Focus cycles through 2N panes (0 to 2N-1)

#### Example 3: File History (3-Pane Hybrid)

**Layout:**
```
┌────────────┬─────────────────────────────────┐
│  Commits   │  Files                          │  Top row
│  List      │  (Changed files in commit)      │  (24 + remaining)
│  (24 wide) │                                 │
└────────────┴─────────────────────────────────┘
┌──────────────────────────────────────────────┐
│  Diff (full width)                           │  Bottom row
│  (Shows file changes)                        │  (full width)
└──────────────────────────────────────────────┘
 ↑↓ scroll | TAB cycle | V visual | Y copy | ESC
```

**Key characteristics:**
- Hybrid layout: 2 panes top, 1 pane bottom
- Fixed 24-char commits pane (same as History)
- Files pane takes remaining top width
- Diff pane full width on bottom
- Context-sensitive status bar (normal vs VISUAL mode)
- 3-pane focus cycle (commits → files → diff → commits)

### Registration Pattern

**Key handlers in `internal/app/app.go`:**

```go
ModeYourMode: NewModeHandlers().
    On("up", a.handleYourModeUp).
    On("down", a.handleYourModeDown).
    On("k", a.handleYourModeUp).     // Vim binding
    On("j", a.handleYourModeDown).   // Vim binding
    On("tab", a.handleYourModeTab).  // Focus cycling
    On("esc", a.handleYourModeEsc).  // Return to menu
    Build(),
```

### Navigation Handler Pattern

**Up/Down handlers route based on focused pane:**

```go
func handleYourModeUp(app *Application) (tea.Model, tea.Cmd) {
    switch app.yourState.FocusedPane {
    case PaneList:
        // Navigate list item up
        if app.yourState.SelectedIdx > 0 {
            app.yourState.SelectedIdx--
        }
    case PaneContent:
        // Scroll content up
        if app.yourState.ContentScrollOff > 0 {
            app.yourState.ContentScrollOff--
        }
    }
    return app, nil
}
```

### Common Pitfalls to Avoid

❌ **Manual line-by-line joining:**
```go
// WRONG - Manual loops, padding, trimming
for i := 0; i < maxLines; i++ {
    combinedLine := leftLine + " " + rightLine  // Gap!
    allLines = append(allLines, combinedLine)
}
```

✅ **lipgloss.JoinHorizontal:**
```go
// RIGHT - Let lipgloss handle borders
topRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
```

❌ **Adding gaps between panes:**
```go
// WRONG - Creates visible gap
filesPaneWidth := width - commitPaneWidth - 1  // -1 gap
topRow := commitLine + " " + filesLine        // Space between
```

✅ **Panes touching directly:**
```go
// RIGHT - Borders touch, no gaps
filesPaneWidth := width - commitPaneWidth
topRow := lipgloss.JoinHorizontal(lipgloss.Top, commits, files)
```

❌ **Hardcoded height calculations:**
```go
// WRONG - Magic numbers
topRowHeight := height / 3
bottomRowHeight := height * 2 / 3  // Doesn't account for status bar
```

✅ **Proven formula:**
```go
// RIGHT - Use exact formula from ConflictResolver
totalPaneHeight := height - 7
topRowHeight := totalPaneHeight / 3
bottomRowHeight := totalPaneHeight - topRowHeight
topRowHeight += 2
bottomRowHeight -= 2
```

### Padding & Text Centering

❌ **Manual string.Repeat padding (error-prone):**
```go
// WRONG - Easy to get wrong, hard to maintain
leftPad := (width - textWidth) / 2
rightPad := width - textWidth - leftPad
if leftPad < 0 { leftPad = 0 }    // Bounds checking scattered
if rightPad < 0 { rightPad = 0 }  // Easy to forget
statusBar = strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
```

✅ **Use lipgloss styling (clear intent):**
```go
// RIGHT - Clear, handles edge cases automatically
style := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
result := style.Render(text)
```

✅ **Or use helper utility:**
```go
// RIGHT - Reusable, testable
result := ui.CenterAlignLine(text, width)
```

**Implementation status:**
- ✅ statusbar.go - Unified BuildStatusBar() handles all centering
- ✅ history.go - Uses lipgloss.Width().Align(lipgloss.Center)
- ✅ filehistory.go - Uses BuildStatusBar() after refactor
- ✅ conflictresolver.go - Uses BuildStatusBar() after refactor

### Type Conversions Across Packages

❌ **Duplicate conversion code (violation of DRY):**
```go
// handlers.go - Two handlers with identical code (lines 959-968 and 1008-1017)
for _, gitFile := range gitFileList {
    convertedFiles = append(convertedFiles, ui.FileInfo{
        Path:   gitFile.Path,
        Status: gitFile.Status,
    })
}
// ... same code appears again in different handler
```

✅ **Extract to utility helper:**
```go
// handlers.go - Implemented after line 26
func convertGitFilesToUIFileInfo(gitFiles []git.FileInfo) []ui.FileInfo {
    converted := make([]ui.FileInfo, len(gitFiles))
    for i, gf := range gitFiles {
        converted[i] = ui.FileInfo{Path: gf.Path, Status: gf.Status}
    }
    return converted
}

// Both handlers now use same function:
state.Files = convertGitFilesToUIFileInfo(gitFileList)  // handleFileHistoryUp
state.Files = convertGitFilesToUIFileInfo(gitFileList)  // handleFileHistoryDown
```

**Implementation status:**
- ✅ `convertGitFilesToUIFileInfo()` implemented in handlers.go (line 27-39)
- ✅ Both call sites updated (handleFileHistoryUp, handleFileHistoryDown)
- ✅ Duplication eliminated

**Benefits realized:** 
- Single source of truth for conversion logic
- If git.FileInfo adds fields, update conversion in one place
- Easier to test the conversion logic
- Follows DRY principle

### Checklist for New Multi-Pane Components

- [ ] Use height calculation formula exactly (height - 7, split 1/3 + 2/3, adjust +2/-2)
- [ ] Choose width pattern (fixed + remainder OR equal distribution)
- [ ] Use lipgloss.JoinHorizontal for side-by-side panes
- [ ] Assemble with direct string concatenation (no gaps)
- [ ] Create focus enum (PaneLeft, PaneRight, etc.)
- [ ] Implement TAB cycling (% numPanes)
- [ ] Add focus-based border colors to each pane renderer
- [ ] Build context-sensitive status bars
- [ ] Register up/down/tab/esc key handlers
- [ ] Route up/down based on focused pane (list nav vs content scroll)
- [ ] Test with different terminal sizes (borders must not overflow)

### Files Implementing This Pattern

| Component | File | Top Row | Bottom Row | Panes | Status Modes |
|-----------|------|---------|------------|-------|--------------|
| History | `internal/ui/history.go` | Commits + Details | (none) | 2 | 1 (normal) |
| Conflict Resolver | `internal/ui/conflictresolver.go` | N file lists | N content panes | 2N | 1 (normal) |
| File History | `internal/ui/filehistory.go` | Commits + Files | Diff | 3 | 2 (normal + VISUAL) |

---

## Dispatcher Pattern (Menu Item → Mode)

Dispatchers route menu actions to appropriate modes:

```
User selects "Commit changes"
    ↓
handleMenuEnter() calls dispatchAction("commit")
    ↓
Handler in app.go: dispatchCommit()
    ↓
Set mode, prompt, action, reset state
    ↓
Return to Update() to re-render
```

**Key Dispatcher Functions:** `internal/app/dispatchers.go`
- `dispatchInit()` - ModeInitializeLocation
- `dispatchClone()` - ModeCloneURL
- `dispatchCommit()` - ModeInput with prompt
- `dispatchPush()` - Execute immediately (async)
- etc.

---

## Input Handling Lifecycle

### Text Input Mode

```
User types character
    ↓
app.Update(tea.KeyMsg)
    ↓
isInputMode() returns true
    ↓
Character handler inserts at cursor
    ↓
updateInputValidation() checks format
    ↓
View() renders updated input + validation feedback
```

### Validation Flow

```
User enters URL in clone mode
    ↓
On every character: updateInputValidation()
    ↓
ValidateRemoteURL() returns valid? + message
    ↓
If invalid: inputValidationMsg = "Invalid URL format"
    ↓
View() renders validation message below input
    ↓
User presses Enter: validateAndProceed()
    ↓
If invalid: footer shows error, don't advance mode
    ↓
If valid: proceed to next step
```

### Paste Handling

Bracketed paste (ctrl+v / cmd+v) comes as single KeyMsg with `msg.Paste = true`:

```go
if msg.Paste && a.isInputMode() {
    text := strings.TrimSpace(string(msg.Runes))
    a.insertTextAtCursor(text)
    a.updateInputValidation()
}
```

---

## Async Operation Lifecycle

### Setup Phase

```go
asyncOperationActive = true
asyncOperationAborted = false
previousMode = a.mode          // Save for ESC restore
previousMenuIndex = a.selectedIndex
mode = ModeConsole             // Show output
consoleState = NewConsoleOutState()
outputBuffer.Clear()
footerHint = "Operation in progress. (ESC to abort)"
```

### Worker Phase

```go
return func() tea.Msg {        // Closure captures state
    url := a.cloneURL          // Captured before goroutine starts
    result := git.ExecuteWithStreaming("clone", url)
    
    // Log to shared OutputBuffer (thread-safe)
    // Do NOT mutate Application directly
    
    return GitOperationMsg{
        Step: "clone",
        Success: result.Success,
        Path: clonePath,
        Error: err,
    }
}
```

### Completion Phase

```
Worker returns GitOperationMsg
    ↓
app.Update(msg) handles based on msg.Step
    ↓
If success:
    - os.Chdir(msg.Path) if path provided
    - Reload git state
    - Mark asyncOperationActive = false
    - Show success message in footer
    - Stay in ModeConsole (user dismisses with ESC)
    ↓
If error:
    - Mark asyncOperationActive = false
    - Show error message in footer
    - Stay in ModeConsole
    ↓
User presses ESC
    ↓
handleKeyESC(): restore previousMode + previousMenuIndex
    ↓
Return to ModeMenu with regenerated menu
```

---

## Confirmation Dialog System

### Purpose

Confirmation dialogs provide safe UX for destructive operations:
- Nested repository warnings
- Force push confirmations
- Hard reset warnings
- Blocking user mistakes

### Flow

```go
User initiates destructive action
    ↓
Code calls app.showNestedRepoWarning(path)
    ↓
app.confirmationDialog = NewConfirmationDialog(config, width, theme)
app.mode = ModeConfirmation
    ↓
View() renders confirmationDialog.Render()
    ↓
User presses left/right/y/n to select button
    ↓
User presses enter to confirm
    ↓
handleConfirmationEnter() → handleConfirmationResponse(confirmed)
    ↓
confirmationActions/confirmationRejectActions dispatch
    ↓
Handler executes operation or returns to menu
```

### Components

**ConfirmationDialog** (`internal/ui/confirmation.go`):
- ConfirmationConfig: title, explanation, yesLabel, noLabel, actionID
- ButtonSelection: enum (ButtonYes, ButtonNo)
- Methods: SelectYes(), SelectNo(), ToggleSelection(), GetSelectedButton()
- Render() with button styling based on selection state

**Handlers** (`internal/app/confirmationhandlers.go`):
- showConfirmation(config) - Display dialog and enter ModeConfirmation
- showNestedRepoWarning(path) - Pre-built config for nested repo warnings
- showForcePushWarning(branchName) - Pre-built config for force push
- showHardResetWarning() - Pre-built config for hard reset
- showAlert(title, explanation) - Single-button alert dialog
- confirmationActions map - YES button handlers
- confirmationRejectActions map - NO button handlers
- handleConfirmationResponse(confirmed) - Router to appropriate handler

### Keyboard Interaction

| Key | Action |
|-----|--------|
| left/h | Select Yes button |
| right/l | Select No button |
| y | Select Yes |
| n | Select No |
| enter | Confirm selection |
| esc | Cancel (global handler, dismisses dialog) |

### Styling

Dialog uses lipgloss.Place() to center both horizontally and vertically within ContentHeight.

Button colors from theme:
- Selected button: MenuSelectionBackground + HighlightTextColor
- Unselected button: InlineBackgroundColor + ContentTextColor
- Dialog border: BoxBorderColor
- Text: ContentTextColor
- Highlighted text (commit hashes): AccentTextColor

Dialog width: `ContentInnerWidth - 10` (leaves padding for visual centering)

---

## Configuration & State Persistence

### Per-Repository State

Git state is **always detected from actual git commands.** No config file, no tracking.

```go
state, err := git.DetectState()
// Queries: git status, git rev-parse, git remote, git log, etc.
```

**Single-branch model:** TIT operates on the currently checked-out branch only. No canon/working branch tracking. User can switch branches anytime with normal git commands.

**Fresh repository behavior:** Repos with no commits remain uncommitted. Timeline is N/A until the first commit exists.

### User Configuration

Theme colors: `~/.config/tit/themes/default.toml`
- Loaded once at app start
- All UI uses semantic color names (SSOT in `internal/ui/theme.go`)
- User can customize without code changes

---

## Thread Safety

### Guaranteed Safe (UI Thread Only)

- All Application mutations
- View() rendering
- menu.GenerateMenu()
- Keyboard input handling

### Shared Between Threads

- OutputBuffer (thread-safe ring buffer)
  - Worker calls `buffer.Append(line, type)`
  - UI thread calls `buffer.Lines()`
  - No locks needed (atomic operations)

### Worker Thread Rules

```go
func executeOperation() tea.Cmd {
    // UI THREAD - Capturing state
    url := a.cloneURL      // Captured value
    path := a.clonePath    // Captured value
    
    return func() tea.Msg {
        // WORKER THREAD - Never touch Application
        // ❌ a.cloneURL = ""  // Race condition!
        // ❌ a.asyncOperationActive = false  // Race condition!
        
        // ✅ Read captured values
        // ✅ Write to OutputBuffer
        // ✅ Return immutable message
        
        output := ui.GetBuffer()
        output.Append("Starting clone...", ui.TypeCommand)
        
        result := git.ExecuteWithStreaming("clone", url, path)
        
        return GitOperationMsg{
            Step: "clone",
            Success: result.Success,
            Path: path,
        }
    }
}
```

---

## Key Files & Responsibilities

### Core Application Logic (`internal/app/`)

| File | Purpose | Key Types |
|------|---------|-----------|
| `app.go` | Main Application struct, Update() event loop, View() rendering | Application (120+ fields) |
| `modes.go` | AppMode enum definition (14 modes), ModeMetadata descriptors | AppMode (enum), SetupWizardStep |
| `menu.go` | Menu generators (state → []MenuItem), menu state functions | menuNormal(), menuNotRepo(), menuTimeTraveling() |
| `menu_items.go` | MenuItems SSOT map (30+ items with labels/hints/emoji) | MenuItem struct, MenuItems map |
| `menu_builders.go` | MenuItemBuilder fluent API for separators and items | MenuBuilder type |
| `messages.go` | Custom tea.Msg types, SSOT maps (prompts, dialogs, errors) | StateDescriptions, DialogMessages, ErrorMessages maps |
| `operation_steps.go` | OperationStep constants SSOT (25+ async operation names) | Op* constants (OpInit, OpCommit, etc) |
| `keyboard.go` | Key handler registry construction, keyboard binding setup | KeyHandler type, buildKeyHandlers() |
| `handlers.go` | Input handlers (enter, ESC, text input, character input) | handleMenuEnter(), handleKeyESC(), handleTextInput() |
| `dispatchers.go` | Menu item → mode transitions, route selection to handler | dispatchInit(), dispatchCommit(), dispatchPush() |
| `state_info.go` | State info maps builder (WorkingTree/Timeline/Operation descriptions) | BuildStateInfo(), StateInfo struct |
| `confirmation_handlers.go` | Confirmation dialog system, show*Warning() functions | showConfirmation(), confirmationActions map |
| `conflict_state.go` | Conflict resolution state struct | ConflictResolveState struct |
| `conflict_handlers.go` | Conflict resolution keyboard handlers | handleConflictUp(), handleConflictSpace() |
| `git_handlers.go` | Git operation completion handlers, cache rebuild | handleGitOperationMsg() |
| `async.go` | Async command execution helpers, streaming wrappers | executeOperation(), cmdXxx functions |
| `history_cache.go` | History cache preload functions | preloadHistoryMetadata(), preloadFileHistoryDiffs() |
| `setup_wizard.go` | SSH setup wizard flow and handlers | handleSetupWizard(), stepWelcome(), etc |
| `cursor_movement.go` | Text input cursor movement helpers | moveCursorLeft(), moveCursorRight() |
| `location.go` | Clone/init location selection and path validation | promptForCloneLocation() |
| `dirty_state.go` | Dirty operation tracking (merge/rebase with uncommitted changes) | DirtyOperationState struct |
| `errors.go` | Error type definitions and handling | AppError type |
| `config.go` | Application configuration and theme loading | AppConfig struct |
| `key_builder.go` | Key handler builder pattern implementation | KeyHandlerBuilder type |

### Git Integration (`internal/git/`)

| File | Purpose | Key Functions |
|------|---------|---|
| `state.go` | State detection from git commands (5-axis system) | DetectState(), detectWorkingTree(), detectTimeline() |
| `execute.go` | Command execution with streaming, git command wrappers | executeGitCommand(), executeWithStreaming() |
| `types.go` | All git types (State, WorkingTree, Timeline, Operation, etc) | State, CommitInfo, CommitDetails, FileInfo structs |
| `init.go` | Repository initialization helpers | initRepository(), validateRepoName() |
| `ssh.go` | SSH configuration and key management | checkSSHKeys(), generateSSHKey() |
| `environment.go` | Git/SSH environment detection | CheckEnvironment(), isGitInstalled() |
| `messages.go` | Git command output message parsing | parseGitOutput(), interpretExitCode() |
| `dirtyop.go` | Dirty operation detection and state management | IsDirtyOperationActive(), captureSnapshot() |

### UI Components (`internal/ui/`)

| File | Purpose | Key Functions |
|------|---------|---|
| `layout.go` | RenderReactiveLayout() main view composer, responsive layout | RenderReactiveLayout(), renderTooSmallMessage() |
| `sizing.go` | Dynamic sizing calculations, responsive dimensions | DynamicSizing struct, CalculateDynamicSizing() |
| `header.go` | Header rendering (9-line layout), state display | RenderHeaderInfo(), RenderHeader(), HeaderState |
| `theme.go` | Color system with semantic names, theme loading | Theme struct (50+ colors), LoadTheme() |
| `menu.go` | RenderMenuWithHeight() component | RenderMenuWithHeight(), RenderMenuWithSelection() |
| `buffer.go` | OutputBuffer thread-safe ring buffer, streaming output | OutputBuffer type, Append(), Lines() |
| `console.go` | RenderConsoleOutput() component for git command output | RenderConsoleOutput(), ConsoleOutState |
| `confirmation.go` | ConfirmationDialog component, dialog rendering | ConfirmationDialog type, Render() |
| `conflictresolver.go` | N-column parallel conflict resolution UI | RenderConflictResolveGeneric(), ConflictFileGeneric |
| `history.go` | Commit history pane rendering (2-pane layout) | RenderHistory() |
| `filehistory.go` | File history mode (3-pane hybrid layout) | RenderFileHistory(), FileHistoryState |
| `listpane.go` | Reusable list pane with scrolling and selection | renderListPane(), ListPaneConfig |
| `textpane.go` | Text viewing pane (for diff, console output) | renderTextPane() |
| `input.go` | Text input box rendering | RenderInput(), InputConfig |
| `textinput.go` | Basic text input component | TextInputComponent |
| `branchinput.go` | Dual input component (canon + working branch) | RenderBranchInput() |
| `formatters.go` | Text utilities (padding, truncation, centering) | PadText(), TruncateText(), CenterAlignLine() |
| `statusbar.go` | Unified status bar builder | BuildStatusBar() |
| `validation.go` | Input validation (URLs, names, etc) | ValidateRemoteURL(), ValidateRepoName() |
| `spinner.go` | Loading spinner animation | RenderSpinner() |
| `box.go` | Border box rendering | RenderBox(), BoxConfig |

### Banner & Config (`internal/banner/`, `internal/config/`)

| File | Purpose |
|------|---------|
| `internal/banner/svg.go` | SVG to braille conversion (logo rendering) |
| `internal/banner/braille.go` | Braille character utilities |
| `internal/config/stash.go` | Stash management (loading saved state) |

---

---

## Common Patterns

### Adding a New Menu Item

**Step 1: Define in SSOT** (`menuitems.go`)
```go
var MenuItems = map[string]MenuItem{
    // ... existing items ...
    "commit": {
        ID:       "commit",
        Shortcut: "m",
        Emoji:    "📝",
        Label:    "Commit changes",
        Hint:     "Create a new commit from staged changes",
        Enabled:  true,
    },
}
```

**Step 2: Use in menu generator** (`menu.go`)
```go
// menuWorkingTree()
items = append(items, GetMenuItem("commit"))
// That's it! No inline builders
```

**Step 3: Add dispatcher** (`dispatchers.go`)
```go
func (a *Application) dispatchCommit() (tea.Model, tea.Cmd) {
    a.mode = ModeInput
    a.inputPrompt = InputPrompts["commit_message"]
    a.inputAction = "commit"
    return a, nil
}
```

**Step 4: Add handler** (`handlers.go`)
```go
func (a *Application) handleCommitSubmit() (tea.Model, tea.Cmd) {
    message := a.inputValue
    // ... execute commit ...
}
```

**Step 5: Register dispatcher** (`menu.go` generator mapping or explicit check in handlers)
- Menu selection → handleMenuEnter() → dispatchers.handleMenuAction(itemID)
- Dispatcher lookup is already automatic in menu handling

**Benefits of SSOT approach:**
- All text in one place (easy to audit, translate, change)
- Shortcuts checked for conflicts at build time
- Hints automatically in footer (no code duplication)
- Menu generators stay simple (just GetMenuItem calls)

### Async Operation with Streaming

```go
// 1. Set async state
a.asyncOperationActive = true
a.mode = ModeConsole
a.outputBuffer.Clear()

// 2. Return command that spawns worker
return a, a.executeOperation()

// 3. Worker streams output
func (a *Application) executeOperation() tea.Cmd {
    path := a.repositoryPath  // Capture before goroutine
    
    return func() tea.Msg {
        output := ui.GetBuffer()
        output.Append("Starting operation...", ui.TypeCommand)
        
        // Git command automatically streams to buffer
        result := git.ExecuteWithStreaming("status")
        
        return GitOperationMsg{
            Step: OpCommit,  // Use constant from operationsteps.go
            Success: result.Success,
        }
    }
}

// 4. Handle completion in githandlers.go
case OpCommit:
    a.gitState, _ = git.DetectState()
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
    a.footerHint = GetFooterMessageText(MessageOperationComplete)
    a.asyncOperationActive = false
```

**Operation Step Constants** (`internal/app/operationsteps.go`):
```go
// All operation step names centralized as constants
// Used in GitOperationMsg.Step field for operation routing
const (
    OpInit              = "init"
    OpClone             = "clone"
    OpCommit            = "commit"
    OpPush              = "push"
    OpPull              = "pull"
    OpAddRemote         = "add_remote"
    OpDirtyPullSnapshot = "dirty_pull_snapshot"
    // ... and 20+ more
)
```

**Why this pattern:**
- All operation names in one SSOT file (operationsteps.go)
- GitOperationMsg.Step uses constants, never hardcoded strings
- Handlers switch on constants (case OpCommit:)
- Typos caught at compile time, not at runtime

---

## Error Handling Best Practices (FAIL-FAST Rule)

**Critical Rule:** Never silently suppress errors or return fallback values. Fail early and loudly.

### Anti-Patterns ❌

```go
// WRONG: Silent error suppression
stdout, _ := cmd.StdoutPipe()  // If StdoutPipe fails, stdout is nil
scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
for scanner.Scan() { ... }  // Crashes here with confusing error

// WRONG: Using string literals for error messages
if err != nil {
    return "Operation failed"  // Doesn't tell user WHAT failed
}

// WRONG: Empty return on error
executeGitCommand(...) returns ""  // Masks why it failed
```

### Correct Patterns ✅

```go
// RIGHT: Check error immediately and return meaningful message
stdout, err := cmd.StdoutPipe()
if err != nil {
    return GitOperationMsg{
        Step: OpMyOperation,
        Success: false,
        Error: ErrorMessages["operation_failed"],  // From SSOT
    }
}

// RIGHT: Use SSOT maps for all user-facing messages
return GitOperationMsg{
    Step: OpCommit,
    Success: false,
    Error: ErrorMessages["failed_commit_merge"],  // Specific, from SSOT
}

// RIGHT: Fail fast in handlers
state, err := git.DetectState()
if err != nil {
    return nil, nil  // Return error via model state, not silently ignore
}
```

### Error Message Categories (SSOT)

```go
// messages.go - Three error categories
ErrorMessages["..."]        // Specific operation failure (git returned error)
OutputMessages["..."]       // Operation phase output (informational)
FooterHints["..."]          // User guidance (what to do next)
```

**Example flow:**
```
User selects "Commit"
    ↓
git commit fails (exit code 1)
    ↓
Handler: Check error explicitly
    ↓
Return GitOperationMsg with ErrorMessages["failed_commit_merge"]
    ↓
githandlers.go catches, displays error in console
    ↓
User sees specific reason why commit failed (not generic "Operation failed")
```

---

## Utility Functions & Helper Patterns

### Text Formatting Utilities

All text formatting helpers live in `internal/ui/formatters.go`:

- `PadText(text, width)` - Right-pad text to fixed width
- `CenterAlignLine(text, width)` - Center text within width (already exists in formatters.go)
- `TruncateText(text, width)` - Truncate to width with ellipsis

**Usage pattern:**
```go
import "tit/internal/ui"

// Right-pad to width
padded := ui.PadText("hello", 10)      // "hello     "

// Center to width
centered := ui.CenterText("hi", 10)    // "    hi    "

// Truncate with ellipsis
short := ui.TruncateText(longText, 20) // Ends with "..."
```

**Why separate?** Centralizes text calculations, makes them reusable, easier to maintain if width logic changes.

### Status Bar Building

Status bars across different modes follow a **consistent pattern** via unified builder.

**Consolidated builder:**
```go
// internal/ui/statusbar.go
type StatusBarConfig struct {
    Parts      []string       // Pre-rendered parts
    Width      int            // Terminal width
    Centered   bool           // Center or left-align
    Theme      *Theme
}

func BuildStatusBar(config StatusBarConfig) string {
    // Handles joining with separators, centering/padding
}
```

**Usage pattern across all modes:**
1. Define shortcut styles (bold, accent color)
2. Define description styles (content/dimmed colors)
3. Build parts array with styled shortcuts + descriptions
4. Call `BuildStatusBar()` with parts, width, theme

**Current implementations:**
- ✅ `buildHistoryStatusBar()` (history.go:158) - Uses BuildStatusBar
- ✅ `buildFileHistoryStatusBar()` (filehistory.go:218) - Uses BuildStatusBar
- ✅ `buildDiffStatusBar()` (filehistory.go:259) - Uses BuildStatusBar (with visual mode special case)
- ✅ `buildGenericConflictStatusBar()` (conflictresolver.go:182) - Uses BuildStatusBar

### Type Conversion Helpers

Convert between git types and UI types to avoid import cycles:

**Pattern in `internal/app/handlers.go`:**
```go
// Convert git.FileInfo to ui.FileInfo (both have Path, Status fields)
func convertGitFilesToUIFileInfo(gitFiles []git.FileInfo) []ui.FileInfo {
    converted := make([]ui.FileInfo, len(gitFiles))
    for i, gf := range gitFiles {
        converted[i] = ui.FileInfo{Path: gf.Path, Status: gf.Status}
    }
    return converted
}

// Usage in both handleFileHistoryUp and handleFileHistoryDown
app.fileHistoryState.Files = convertGitFilesToUIFileInfo(gitFileList)
```

**Why here?** 
- Avoids circular imports (ui can't import app, app can't import ui)
- Handlers are the boundary where git and UI types meet
- Other handlers can reuse if needed

### File History State Management

The `updateFileHistoryDiff()` function (handlers.go:898) exemplifies cache lookup pattern:

```go
func (a *Application) updateFileHistoryDiff() {
    // 1. Bounds check
    if len(a.fileHistoryState.Files) == 0 {
        a.fileHistoryState.DiffContent = ""
        return
    }

    // 2. Determine version based on git state
    version := "parent"  // Default
    if a.gitState.WorkingTree == git.Dirty {
        version = "wip"   // Modified: show vs working tree
    }

    // 3. Build cache key following SSOT: hash:path:version
    cacheKey := commit.Hash + ":" + file.Path + ":" + version

    // 4. Thread-safe cache lookup
    a.diffCacheMutex.Lock()
    diffContent, exists := a.fileHistoryDiffCache[cacheKey]
    a.diffCacheMutex.Unlock()

    // 5. Update state
    if exists && diffContent != "" {
        a.fileHistoryState.DiffContent = diffContent
    } else {
        a.fileHistoryState.DiffContent = ""  // Not cached yet
    }
}
```

**Pattern for similar cache operations:**
1. Validate 

---

## Type Definitions Location Map

All types in TIT are centralized in logical locations. This map helps new contributors find type definitions quickly.

### Core Git Types (`internal/git/types.go`)

**State Detection & Representation:**
- `WorkingTree` (alias string) — Clean | Dirty
- `Timeline` (alias string) — InSync | Ahead | Behind | Diverged
- `Operation` (alias string) — NotRepo | Normal | Conflicted | Merging | Rebasing | DirtyOperation | TimeTraveling
- `Remote` (alias string) — NoRemote | HasRemote
- `State` (struct) — Complete git state (WorkingTree, Timeline, Operation, Remote, CurrentBranch, LocalBranchOnRemote)

**Commit & File Information:**
- `CommitInfo` (struct) — Hash, Subject, Time
- `CommitDetails` (struct) — Extended commit info (Author, Date, etc.)
- `FileInfo` (struct) — Path, Status (for staged/unstaged files)

**Time Travel:**
- `TimeTravelInfo` (struct) — OriginalBranch, OriginalHead, CurrentCommit, OriginalStashID
  - **Related:** `git.CommitInfo` (represents currently-checked-out commit)

**Stash Management:**
- `StashEntry` (struct) — Single stash entry (ID, Subject, Time)
- `StashList` (struct) — Collection of stashes
  - **Location:** `internal/config/stash.go`

**Command Execution:**
- `CommandResult` (struct) — Exit code, stdout, stderr from git commands
  - **Location:** `internal/git/execute.go`

**Git Messages (Custom Events):**
- `TimeTravelCheckoutMsg` (struct) — Sent when time travel checkout completes
- `TimeTravelMergeMsg` (struct) — Sent when time travel merge starts
- `TimeTravelReturnMsg` (struct) — Sent when time travel return starts
- `DirtyOperationSnapshot` (struct) — Captures working tree state before operation
  - **Location:** `internal/git/dirtyop.go`

### Application Types (`internal/app/`)

**Mode Management:**
- `AppMode` (alias int) — ModeMenu, ModeHistory, ModeConsole, etc. (12 modes)
  - **Location:** `modes.go`
  - **Related:** `ModeMetadata` (documentation for each mode)
  - **See:** `GetModeMetadata()` for mode descriptions

**Menu:**
- `MenuItem` (struct) — ID, Label, Shortcut, Hint, Enabled
- `MenuGenerator` (func type) — Generates menu items based on git state
  - **Location:** `menu.go`

**Confirmation Dialogs:**
- `ConfirmationType` (alias string) — Types of confirmations
- `ConfirmationAction` (func type) — Handler for confirmation action
- `ConfirmationActionPair` (struct) — Groups confirm + reject handlers
- `ConfirmationMessage` (struct) — Title, Explanation, Confirm/Reject labels (replaces 3 maps)
  - **Location:** `confirmationhandlers.go`
  - **Usage:** `confirmationHandlers` map uses `ConfirmationActionPair`

**Input:**
- `InputMessage` (struct) — Prompt, Hint (replaces 2 maps)
  - **Location:** `messages.go`

**Core Application:**
- `Application` (struct) — Main application state container
  - **Key fields:** gitState, timeTravelInfo, mode, menuItems, caches
  - **Location:** `app.go`

**Keyboard Handling:**
- `KeyHandler` (func type) — Handler for key input
  - **Related:** `app.keyHandlers` map (keyed by AppMode → key → handler)
  - **Location:** `keyboard.go`

**Error Handling:**
- `ErrorLevel` (alias int) — ErrorInfo, ErrorWarn, ErrorFatal
- `ErrorConfig` (struct) — Standardized error configuration
  - **Location:** `errors.go`

**State Machines:**
- `ModeTransition` (struct) — Tracks mode changes for history
- `ConflictResolveState` (struct) — State while resolving merge conflicts
- `DirtyOperationState` (struct) — State during dirty tree operations
  - **Locations:** `app.go`, `conflictstate.go`, `dirtystate.go`

**Custom Messages (Bubble Tea Events):**
- `TickMsg` (alias time.Time) — Timer tick for periodic updates
- `ClearTickMsg` (alias time.Time) — Clear timer message
- `GitOperationMsg` (struct) — Async git operation completion (Step, Success, Error)
- `RestoreTimeTravelMsg` (struct) — Restore time travel state from disk
- `GitOperationCompleteMsg` (struct) — Final operation status
- `InputSubmittedMsg` (struct) — User submitted input
- `CacheProgressMsg` (struct) — Cache building progress
- `FooterMessageType` (alias int) — Categorizes footer messages
  - **Location:** `messages.go`

**Other Application Types:**
- `LocationChoiceConfig` (struct) — Configuration for location selection
- `AsyncOperation` (struct) — Tracks active async operations
- `StateInfo` (struct) — Information about current state
- `ModeHandlerBuilder` (struct) — Builder for mode handlers
- `MenuItemBuilder` (struct) — Builder for menu items
- `ActionHandler` (func type) — Async action handler
- `CursorNavigationMixin` (struct) — Embedded mixin for cursor navigation
  - **Locations:** Various app files

### UI Types (`internal/ui/`)

**History & Commit Display:**
- `CommitInfo` (struct) — Hash, Subject, Time (displayed format)
  - **Note:** Different from `git.CommitInfo` — UI representation
  - **Location:** `history.go`
- `HistoryState` (struct) — Current history view state (selected commit, offset, etc.)
  - **Location:** `history.go`

**File History:**
- `FileInfo` (struct) — Path, Status (UI representation of changed files)
  - **Location:** `filehistory.go`
  - **Note:** Convert from `git.FileInfo` using `convertGitFilesToUIFileInfo()`
- `FileHistoryPane` (alias int) — TopPane, BottomPane
- `FileHistoryState` (struct) — Selected file, diff content, pane heights
  - **Location:** `filehistory.go`

**Conflict Resolution:**
- `ConflictFile` (struct) — Ours, Theirs, Base conflict versions
- `ConflictFileGeneric` (struct) — Generic conflict file representation
  - **Location:** `conflictresolver.go`

**Input & Text:**
- `InputFieldState` (struct) — Current input text, cursor position
- `TextInputState` (struct) — Text input rendering state
- `InputValidator` (func type) — Validates input, returns (bool, error message)
  - **Location:** `input.go`, `textinput.go`, `validation.go`

**Rendering Components:**
- `ListPane` (struct) — Scrollable list rendering
- `ListItem` (struct) — Individual list item
- `DiffLine` (struct) — Single line of diff output
- `ConsoleOutState` (struct) — Console output rendering state
- `OutputLine` (struct) — Single output buffer line (content + type)
- `OutputLineType` (alias string) — TypeStdout, TypeStderr, TypeStatus, TypeInfo
- `OutputBuffer` (struct) — Thread-safe ring buffer for streaming output
  - **Location:** `buffer.go`

**Status & UI Layout:**
- `StatusBarConfig` (struct) — Configuration for status bar rendering
- `StatusBarStyles` (struct) — Pre-computed styles for status bars
- `BoxConfig` (struct) — Configuration for box rendering
- `StyledContent` (struct) — Content with style applied
- `Line` (struct) — Single rendered line with dimensions
  - **Location:** `statusbar.go`, `box.go`

**Confirmation Dialog:**
- `ConfirmationConfig` (struct) — Title, Explanation, Confirm/Reject labels
- `ButtonSelection` (alias string) — Selected button ("confirm" or "reject")
- `ConfirmationDialog` (struct) — Confirmation UI state
  - **Location:** `confirmation.go`

**Theme & Styling:**
- `ThemeDefinition` (struct) — Raw theme TOML data (from file or defaults)
- `Theme` (struct) — Loaded theme with color fields
  - **Related methods:** `ShortcutStyle()`, `DescriptionStyle()` (added for consolidation)
  - **Location:** `theme.go`

**Layout & Sizing:**
- `Sizing` (struct) — Terminal dimensions and content area calculations
  - **Location:** `sizing.go`

### Banner Types (`internal/banner/`)

**SVG/Image Rendering:**
- `Point` (struct) — X, Y coordinates
- `Color` (struct) — RGB color
- `PixelColor` (struct) — Pixel with color
- `ScanlineRange` (struct) — Range of pixels in scanline
- `Intersection` (struct) — SVG path intersection
- `BrailleChar` (struct) — Braille character representation
  - **Location:** `svg.go`, `braille.go`

---

## Type Relationships & Cross-References

**Git State Chain:**
```
git.State
  ├─ Operation → ModeMetadata (determines available UI)
  └─ TimeTravelInfo (populated when Operation == TimeTraveling)
     └─ CommitInfo → ui.CommitInfo (display format)

git.FileInfo → ui.FileInfo (converted in handlers)
```

**Application State Chain:**
```
Application.gitState (git.State)
  ├─ Operation determines AppMode
  ├─ Operation determines MenuItem generation
  └─ Operation determines Confirmation dialogs available

Application.mode (AppMode)
  ├─ Determines rendering (history, console, menu, etc.)
  └─ Determines keyboard handler registry
```

**Error Handling Chain:**
```
git command fails
  ↓
ErrorConfig captures (Level, Message, Operation)
  ↓
GitOperationMsg populated
  ↓
githandlers.go routes based on Operation
```

---

## Adding New Types

**Checklist before creating a new type:**
1. ✅ Check existing types (grep for similar names)
2. ✅ Verify location follows pattern (git types in git/types.go, app types in app/, ui types in ui/)
3. ✅ Add doc comment linking related types
4. ✅ Update this location map if visible externally
5. ✅ Use type aliases (string, int) for semantic clarity, not new named structs
6. ✅ Group related types (e.g., all confirmation types in confirmationhandlers.go)

**Example: Adding a new menu item type**
```go
// app/menuitems.go (location: same as MenuItem)
type MenuItemCategory string

const (
    CategoryWorkflow MenuItemCategory = "workflow"
    CategoryHistory  MenuItemCategory = "history"
)

// Update ARCHITECTURE.md type map to include MenuItemCategory
```

---

## File Organization by Feature

**Principle:** All files in `internal/app/` use **feature-based naming convention** within a single `package app`. This maximizes agent readability without Go package constraints.

### File Organization Map

**Initialization Feature** (UI prompts + workflow execution)
- `init_location.go` - Location choice dialog (cwd vs subdirectory)
- `init_branches.go` - Dual branch name input (canon + working)
- `init_workflow.go` - Execute init, save config

**Clone Feature** (URL input + location choice + workflow)
- `clone_url.go` - URL input validation
- `clone_location.go` - Clone location choice (cwd vs subdirectory)
- `clone_workflow.go` - Execute clone, detect branches

**History Feature** (Commit history + file history)
- `history_menu.go` - History mode menu items
- `history_cache.go` - Cache precomputation (mandatory contract)
- `history_render.go` - History view rendering (if split from UI)

**Conflict Resolution**
- `conflict_handlers.go` - Conflict handler dispatch
- `conflict_state.go` - Conflict resolution state tracking
- `conflictresolver.go` (in ui/) - Conflict resolver UI

**Time Travel Feature** (Checkout + state tracking + merge/return)
- `time_travel_handlers.go` - Time travel checkout/merge/return operations
- `dirtystate.go` - Dirty working tree handling during time travel

**Git Operations** (General git state + handler routing)
- `git_handlers.go` - State change handlers, operation routing
- `githandlers.go` (legacy naming, to be renamed)

**Core Application Structure**
- `app.go` - Application struct, main event loop
- `modes.go` - AppMode enum + metadata descriptions
- `menu.go` - Menu rendering (View layer for menu mode)
- `menu_builders.go` - Menu item generation from git state
- `menu_items.go` - MenuItem definitions + menu item registry
- `keyboard.go` - Key binding registry (cmd* → handler dispatch)
- `dispatchers.go` - Menu item ID → mode transition
- `handlers.go` - Input handlers (enter, ESC, text input)

**Git Operations**
The operation logic is organized into 9 focused files:
- `op_init.go` - Repository initialization
- `op_clone.go` - Repository cloning
- `op_remote.go` - Remote operations
- `op_commit.go` - Commit operations
- `op_push.go` - Push operations
- `op_pull.go` - Pull operations
- `op_dirty_pull.go` - Dirty pull handling
- `op_merge.go` - Merge operations
- `op_time_travel.go` - Time travel operations

**Configuration & Storage**
- `config.go` - Repo config load/save (`~/.config/tit/repo.toml`)
- `location.go` - Directory location utilities

**UI & Messages**
- `messages.go` - All user-facing messages (prompts, hints, confirmations, errors)
- `confirmationhandlers.go` - Confirmation dialog logic + action handlers
- `errors.go` - Error handling pattern + levels

**Async & State**
- `async_state.go` - Async operation state struct
- `input_state.go` - Input field management struct  
- `cache_manager.go` - History cache lifecycle struct
- `cursor_movement.go` - Menu cursor navigation

**Layer Architecture**
- `git_logger.go` - Implements git.Logger interface, removes git→ui layer violation
- SSOT helper functions:
  - `reloadGitState()` - Eliminates 19+ duplicated state reload patterns
  - `checkForConflicts()` - Eliminates 6 duplicated conflict detection patterns
  - `executeGitOp()` - Standardizes simple git command execution (2 usages)

### Why This Organization?

**For Agent Understanding:**
1. **Semantic naming** - `ls internal/app/init_*.go` shows entire init feature
2. **Grep precision** - `grep -r "func" internal/app/history_*.go` finds all history functions
3. **No import tracing** - Single package = no circular dependency investigation
4. **Go idiom** - Standard library pattern (e.g., `net/http/request.go`, `net/http/server.go`)
5. **Git history** - `git log --follow internal/app/clone_*.go` tracks clone feature evolution
6. **Refactoring** - Just rename files, no code changes needed

### Finding Code by Intent

When looking for specific functionality:
1. **"How does initialization work?"** → Read `init_*.go` files in order
2. **"What happens when user clicks menu item?"** → Start at `menu_items.go` → `dispatchers.go` → `handlers.go` → `*_workflow.go`
3. **"How is git state detected?"** → `internal/git/state.go` → `git_handlers.go` routes to handlers
4. **"How are menus generated?"** → `menu_builders.go` switches on git state → `menu_items.go` defines items
5. **"What messages exist?"** → All in `messages.go` (grouped by InputMessages, ConfirmationMessages, ErrorMessages)

---

## Navigation Tips

**Finding where a type lives:**
1. Is it about git state? → `internal/git/types.go`
2. Is it about application modes/handlers? → `internal/app/*.go`
3. Is it about rendering/UI? → `internal/ui/*.go`
4. Is it a Bubble Tea message? → `internal/app/messages.go`
5. Is it a confirmation dialog? → `internal/app/confirmationhandlers.go`
6. Is it an error? → `internal/app/errors.go`

**Example trace: Where is `State`?**
- Q: "I need to access current git state"
- A: `git.State` in `internal/git/types.go`
- Q: "I need to display the state"
- A: Use `Application.gitState` field in `app.go`
- Q: "I need to show operation in menu"
- A: `Application.gitState.Operation` → lookup `ModeMetadata` in `modes.go`bounds (no nil access)
2. Determine variant (version/type suffix for cache key)
3. Build SSOT cache key with separators
4. Lock → lookup → unlock (thread-safe)
5. Update state with result or empty fallback

---

## Testing Strategy

No automated test suite. Manual testing workflow:

1. **Build:** `./build.sh`
2. **Test scenario:** Create test git repo
3. **Execute:** `./tit_x64` (or tit_arm64)
4. **Verify:** Check menu items, execute operations, verify output
5. **Regression:** Test previous phases still work

---

## Design Decisions

### Why No Config Tracking?

Git state is always fresh from git commands. No config file to maintain or invalidate. Simpler, always correct.

### Why ModeConsole Persists After Operation?

User needs to see operation output (success/error details). They dismiss with ESC when done reading.

### Why OutputBuffer Instead of String Array?

- Thread-safe append from worker goroutines
- Circular buffer prevents unbounded memory growth
- Efficient scrolling (pre-rendered lines)

### Why Dispatcher Pattern?

Clean separation: Menu generators produce items, dispatchers route to modes, handlers consume input. Easy to add new items without touching input logic.

---

## REWIND Operation (git reset --hard)

**Entry Point:** Commit history browser, Ctrl+ENTER on selected commit

**State Transitions:**
```
ModeHistory
  ↓ (Ctrl+ENTER)
Confirmation dialog
  ↓ (User confirms)
ModeConsole + asyncOperationActive = true
  ↓ (executeRewindOperation runs in goroutine)
ModeConsole (waiting for RewindMsg)
  ↓ (git reset --hard completes)
RewindMsg received
  ↓ (handler refreshes git state)
ModeMenu + updated state
```

**Key differences from Time Travel:**
- ENTER: Safe, read-only, reversible (detached HEAD with time travel menu)
- Ctrl+ENTER: Destructive, permanent, discards commits (reset --hard)

**Always Available:** REWIND can be initiated from any Operation state (including TimeTraveling). Confirmation dialog warns if not in Normal state.

**Implementation:**
- `git.ResetHardAtCommit(commitHash)` executes reset
- OutputBuffer streams git output in real-time
- RewindMsg handler refreshes git state and regenerates menu

## Related Documentation

- `SPEC.md` - User-facing behavior specification
- `IMPLEMENTATION_PLAN.md` - Phase-by-phase feature roadmap
- `SESSION-LOG.md` - Development history with session notes
- `COLORS.md` - Theme system color reference
