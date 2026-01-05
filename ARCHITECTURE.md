# TIT Architecture Guide

## Overview

TIT (Git Timeline Interface) is a state-driven terminal UI for git repository management. It follows a clean, event-driven architecture based on Bubble Tea's Model-View-Update pattern.

**Core Principle:** Git state determines UI state. Operations are always safe and abortable.

---

## Four-Axis State Model

Every moment in TIT is described by exactly 4 git axes:

```go
type State struct {
    WorkingTree      git.WorkingTree // Clean | Modified
    Timeline         git.Timeline    // InSync | Ahead | Behind | Diverged | NoRemote
    Operation        git.Operation   // NotRepo | Normal | Conflicted | Merging | Rebasing
    Remote           git.Remote      // NoRemote | HasRemote
    CurrentBranch    string          // Local branch name
    LocalBranchOnRemote bool         // Whether current branch tracked on remote
}
```

**State Detection:** `git.DetectState()` queries git commands (no config file tracking).

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

## Application Modes (AppMode)

| Mode | Purpose | Input Handler |
|------|---------|---|
| ModeMenu | Main action menu | Menu navigation (j/k/enter) |
| ModeInput | Generic text input (deprecated) | Cursor nav + character input |
| ModeInitializeLocation | Choose init location (cwd/subdir) | Menu selection |
| ModeCloneURL | Input clone URL | Single text input with validation |
| ModeCloneLocation | Choose clone location (cwd/subdir) | Menu selection |
| ModeConsole | Show streaming git output | Console scroll (↑↓/PgUp/PgDn) |
| ModeClone | Clone operation streaming output | Same as ModeConsole |
| ModeSelectBranch | Choose branch after clone | Menu selection |
| ModeConfirmation | Confirm destructive operation | left/right/h/l/y/n/enter |
| ModeConflictResolve | 3-way conflict resolution | (TBD - Phase 7) |
| ModeHistory | Commit/file history browser | (TBD - Phase 5) |

---

## Menu System

### MenuGenerator Pattern

Each git state maps to a menu generator function:

```go
type MenuGenerator func(*Application) []MenuItem

menuGenerators := map[git.Operation]MenuGenerator{
    git.NotRepo:    (*Application).menuNotRepo,
    git.Conflicted: (*Application).menuConflicted,
    git.Merging:    (*Application).menuOperation,
    git.Rebasing:   (*Application).menuOperation,
    git.Normal:     (*Application).menuNormal,
}
```

### MenuItem Builder Pattern

Clean fluent API for creating menu items:

```go
Item("commit").
    Shortcut("m").
    Emoji("📝").
    Label("Commit changes").
    Hint("Create commit from staged changes").
    When(isModified).  // Enable only when Working Tree is Modified
    Build()
```

### Generators in `internal/app/menu.go`

- `menuNotRepo()` - Init/Clone (not in repo)
- `menuConflicted()` - Resolve/Abort (conflicts detected)
- `menuOperation()` - Continue/Abort (merge/rebase in progress)
- `menuNormal()` - Full menu (normal state)
  - `menuWorkingTree()` - Commit (when Modified)
  - `menuTimeline()` - Push/Pull based on Timeline
  - `menuHistory()` - Commit history browser

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

**Fresh repository auto-setup:** When detecting a repo with no commits, TIT automatically creates and commits `.gitignore` to ensure Clean working tree state.

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

| File | Purpose |
|------|---------|
| `internal/app/app.go` | Application struct, Update() event loop, key handler registry |
| `internal/app/modes.go` | AppMode enum definition |
| `internal/app/menu.go` | Menu generators (state → []MenuItem) |
| `internal/app/menubuilder.go` | MenuItemBuilder fluent API |
| `internal/app/dispatchers.go` | Menu item → mode transitions |
| `internal/app/handlers.go` | Input handlers (enter, ESC, text input, etc) |
| `internal/app/keyboard.go` | Key handler registry construction |
| `internal/app/messages.go` | Custom tea.Msg types |
| `internal/app/confirmationhandlers.go` | Confirmation dialog system and handlers |
| `internal/git/state.go` | State detection from git commands |
| `internal/git/execute.go` | Command execution with streaming |
| `internal/ui/layout.go` | RenderLayout() main view composer |
| `internal/ui/theme.go` | Color system with semantic names |
| `internal/ui/buffer.go` | OutputBuffer thread-safe ring buffer |
| `internal/ui/console.go` | RenderConsoleOutput() component |
| `internal/ui/confirmation.go` | ConfirmationDialog component |
| `internal/ui/menu.go` | RenderMenuWithHeight() component |
| `internal/ui/validation.go` | Input validation (URLs, directory names) |

---

## Common Patterns

### Adding a New Menu Item

1. **menu.go**: Add to appropriate generator
2. **dispatchers.go**: Add dispatcher function
3. **handlers.go**: Add input handler
4. **keyboard.go**: Register key handler if needed

Example: Adding "Commit changes"

```go
// menu.go: menuWorkingTree()
Item("commit").
    Shortcut("m").
    Label("Commit changes").
    When(a.gitState.WorkingTree == git.Modified).
    Build()

// dispatchers.go
func (a *Application) dispatchCommit() (tea.Model, tea.Cmd) {
    a.mode = ModeInput
    a.inputPrompt = "Commit message:"
    a.inputAction = "commit"
    return a, nil
}

// handlers.go
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
    switch app.inputAction {
    case "commit":
        return app.handleCommitSubmit()
    }
}

func (a *Application) handleCommitSubmit(app *Application) (tea.Model, tea.Cmd) {
    message := app.inputValue
    app.asyncOperationActive = true
    app.mode = ModeConsole
    return app, app.executeCommitWorkflow(message)
}

// handlers.go: executeCommitWorkflow
func (a *Application) executeCommitWorkflow(message string) tea.Cmd {
    return func() tea.Msg {
        result := git.ExecuteWithStreaming("commit", "-m", message)
        return GitOperationMsg{
            Step: "commit",
            Success: result.Success,
            Error: "Commit failed",
        }
    }
}
```

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
            Step: "operation",
            Success: result.Success,
        }
    }
}

// 4. Handle completion
case GitOperationMsg:
    if msg.Success {
        a.asyncOperationActive = false
        a.footerHint = "Operation complete. Press ESC to continue."
    }
```

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

## Related Documentation

- `SPEC.md` - User-facing behavior specification
- `IMPLEMENTATION_PLAN.md` - Phase-by-phase feature roadmap
- `SESSION-LOG.md` - Development history with session notes
- `COLORS.md` - Theme system color reference
