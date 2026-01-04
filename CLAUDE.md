# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Project Overview

**TIT** (Git Timeline Interface) is a terminal UI application for git repository management, built with Go, Bubble Tea (TUI framework), and Lip Gloss (styling).

**Core Philosophy:**
- **Canon branch:** Read-only locally, clean history (typically `main`)
- **Working branches:** Sandbox for messy operations (commit, cherry-pick, rebase, stash)
- **State-driven UI:** Git state determines available menu options

---

## Build & Run

```bash
# Build for current architecture (creates tit_x64 or tit_arm64)
./build.sh

# Run the binary
./tit_x64  # or ./tit_arm64 on ARM

# Build also copies binary to:
# ~/Documents/Poems/inf/___user-modules___/automation/
```

**Note:** No test suite exists. Manual testing only.

---

## Architecture

### Three-Layer State System

**1. Git State Detection (`internal/git/state.go`)**
```
State Tuple: (WorkingTree, Timeline, Operation, Remote)
- WorkingTree: Clean | Modified
- Timeline: InSync | Ahead | Behind | Diverged | NoRemote
- Operation: NotRepo | Normal | Conflicted | Merging | Rebasing
- Remote: NoRemote | HasRemote
```

**2. Application Modes (`internal/app/modes.go`)**
```
AppMode determines rendering and input handling:
- ModeMenu: Main navigation menu
- ModeInput: Generic text input (deprecated, being phased out)
- ModeInitializeLocation: Choose init location (cwd or subdir)
- ModeInitializeBranches: Dual input for canon + working branch names
- ModeCloneURL: Input clone URL
- ModeCloneLocation: Choose clone location
- ModeSelectBranch: Select canon branch from cloned branches
- ModeConsole: Streaming git command output
- ModeConflictResolve: 3-way conflict resolution
```

**3. Menu Generation (`internal/app/menu.go`)**
```
Git State → Menu Items (dynamic dispatch)
- NotRepo → Init / Clone
- Conflicted → Resolve / Abort
- Normal + Behind → Pull options
- Normal + Ahead → Push options
- Normal + Clean → Commit / Amend options
```

### Key Handler Pattern

**Registry-based dispatch** (`internal/app/keyboard.go`):
```go
// Built once on app init, cached in app.keyHandlers
map[AppMode]map[string]func(*Application) (tea.Model, tea.Cmd)

// Example:
handlers[ModeMenu]["enter"] = handleMenuEnter
handlers[ModeInitializeBranches]["enter"] = handleInitBranchesSubmit
```

**Global keys (all modes):**
- `ctrl+c`: Quit confirmation (3s timeout)
- `esc`: Context-dependent (abort async ops, clear input, return to menu)

---

## Workflow Patterns

### Initialization Flow (Dual Input Mode)

```
dispatchInit()
  ↓
ModeInitializeLocation (menu: init cwd vs subdir)
  ↓ handleInitLocationChoice1/Choice2
ModeInitializeBranches (dual text input: canon + working)
  ↓ handleInitBranchesSubmit
executeInitWorkflow() [ASYNC]
  ├─ git init <path>
  ├─ git checkout -b <canon>
  ├─ git checkout -b <working>
  └─ Save RepoConfig (~/.config/tit/repo.toml)
  ↓
GitOperationMsg("init") → Update()
  └─ DetectState() → ModeMenu
```

### Clone Flow

```
dispatchClone()
  ↓
ModeCloneURL (text input with URL validation)
  ↓ handleCloneURLSubmit
ModeCloneLocation (menu: clone to cwd vs subdir)
  ↓ handleCloneLocationChoice1/Choice2
executeCloneWorkflow() [ASYNC]
  ├─ If cwd: git init + remote add + fetch + checkout
  ├─ If subdir: git clone --progress <url> <path>
  └─ Detect branches: ListBranches()
  ↓
GitOperationMsg("clone") → Update()
  ├─ Single branch → auto-set canon, save config → ModeMenu
  └─ Multiple branches → ModeSelectBranch → user choice → ModeMenu
```

### Async Operation Pattern

**All git operations run in goroutines to prevent UI blocking.**

```go
// Start async operation
a.asyncOperationActive = true
a.previousMode = a.mode  // Save for ESC restore
a.previousMenuIndex = a.selectedIndex
a.mode = ModeConsole
return a, executeOperation()  // Returns tea.Cmd

// Worker goroutine
func executeOperation() tea.Cmd {
    return func() tea.Msg {
        // Run blocking git commands
        // Stream output to outputBuffer
        return GitOperationMsg{Step: "clone", Success: true}
    }
}

// Completion
case GitOperationMsg:
    a.asyncOperationActive = false
    if msg.Success {
        a.gitState = git.DetectState()  // Refresh
        a.mode = ModeMenu
    }
```

**User can press ESC during operation:**
```go
a.asyncOperationAborted = true  // Set flag
// Operation completes → ESC handler restores previousMode
```

---

## Threading & State Safety

**Single-threaded UI:** All `Application` struct mutations happen on UI thread (Bubble Tea event loop).

**Worker threads (goroutines):**
- Execute blocking git commands
- Read `Application` fields via closure capture (read-only)
- Return results via `tea.Msg` (immutable messages)
- **NEVER** mutate `Application` directly

**OutputBuffer (`internal/ui/buffer.go`):**
- Thread-safe ring buffer for streaming git output
- `buffer.Append(line)` called from worker goroutine
- `buffer.Lines()` called from UI thread for rendering

---

## Theme System

**All colors centralized** in `internal/ui/theme.go` and `~/.config/tit/themes/default.toml`.

**Naming rule:** Colors describe WHERE/WHAT, not appearance.
- `ContentTextColor` (body text)
- `LabelTextColor` (headers, bright UI)
- `DimmedTextColor` (disabled elements)
- `AccentTextColor` (keyboard shortcuts)
- `BoxBorderColor` (all box borders)

**Adding new colors:**
1. Add to `DefaultThemeTOML` TOML string
2. Add field to `ThemeDefinition.Palette`
3. Add field to `Theme` struct
4. Update `LoadTheme()` mapping
5. Reference as `theme.FieldName`

---

## Input Handling Patterns

### Bracketed Paste (cmd+v, ctrl+v)

```go
// Bubble Tea 1.3.10+ sends entire paste as single KeyMsg
if msg.Paste {
    text := strings.TrimSpace(string(msg.Runes))
    a.inputValue = a.inputValue[:a.cursorPos] + text + a.inputValue[a.cursorPos:]
    a.inputCursorPosition += len(text)
}
```

### ESC Behavior (Context-Dependent)

```go
// During async operation: abort
if a.asyncOperationActive {
    a.asyncOperationAborted = true
}

// Input mode with non-empty text: clear confirmation
if a.isInputMode() && a.inputValue != "" {
    if !a.clearConfirmActive {
        a.clearConfirmActive = true  // First ESC
        // Start 3s timeout
    } else {
        a.inputValue = ""  // Second ESC: clear
    }
}

// Otherwise: return to menu
return a.returnToMenu()
```

---

## Common Tasks

### Adding a New Menu Item

1. **Add to menu generator** (`internal/app/menu.go`):
```go
func (a *Application) menuNormal() []MenuItem {
    items = append(items, MenuItem{
        ID: "my_action",
        Shortcut: "m",
        Label: "My Action",
        Hint: "Does something useful",
        Enabled: true,
    })
}
```

2. **Add dispatcher** (`internal/app/dispatchers.go`):
```go
func (a *Application) dispatchMyAction() (tea.Model, tea.Cmd) {
    a.mode = ModeInput
    a.inputPrompt = "Enter value:"
    a.inputAction = "my_action"
    return a, nil
}
```

3. **Add to menu dispatcher map** (`internal/app/dispatchers.go`):
```go
"my_action": (*Application).dispatchMyAction,
```

4. **Add handler** (`internal/app/handlers.go`):
```go
func (a *Application) handleMyActionSubmit() (tea.Model, tea.Cmd) {
    value := a.inputValue
    // Do work...
    return a.returnToMenu()
}
```

5. **Register key handler** (`internal/app/keyboard.go`):
```go
handlers[ModeInput]["enter"] = func(a *Application) (tea.Model, tea.Cmd) {
    if a.inputAction == "my_action" {
        return a.handleMyActionSubmit()
    }
    // ... other actions
}
```

### Adding a Git Operation

**Pattern:** Async execution with streaming output.

```go
func (a *Application) executeMyOperation() tea.Cmd {
    url := a.someState  // Capture state in closure

    return func() tea.Msg {
        output := ui.GetBuffer()
        output.Clear()

        cmd := exec.Command("git", "my-command")

        // Stream output
        stdout, _ := cmd.StdoutPipe()
        stderr, _ := cmd.StderrPipe()

        cmd.Start()

        // Read and buffer output
        scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
        for scanner.Scan() {
            output.Append(scanner.Text())
        }

        err := cmd.Wait()

        return GitOperationMsg{
            Step: "my_operation",
            Success: err == nil,
            Output: "Done",
            Error: errorMsg,
        }
    }
}
```

---

## File Structure Quick Reference

```
cmd/tit/main.go              # Entry point, program initialization
internal/
├── app/                     # Application logic
│   ├── app.go              # Application struct, Update() event loop
│   ├── modes.go            # AppMode enum
│   ├── menu.go             # Menu generation (state → items)
│   ├── dispatchers.go      # Menu item → mode transitions
│   ├── handlers.go         # Input handlers (enter, ESC, etc)
│   ├── keyboard.go         # Key handler registry
│   ├── messages.go         # Custom tea.Msg types
│   └── config.go           # Repo config load/save
├── git/                     # Git operations
│   ├── state.go            # State detection (WorkingTree, Timeline, etc)
│   ├── execute.go          # Git command execution helpers
│   ├── init.go             # Init/clone operations
│   ├── types.go            # State type definitions
│   └── config.go           # ~/.config/tit/repo.toml
├── ui/                      # UI components & rendering
│   ├── theme.go            # Theme system
│   ├── layout.go           # View() rendering (screen layout)
│   ├── menu.go             # Menu rendering
│   ├── textinput.go        # Text input rendering
│   ├── branchinput.go      # Dual branch input rendering
│   ├── console.go          # Streaming output rendering
│   ├── buffer.go           # Thread-safe output buffer
│   ├── box.go              # Box drawing utilities
│   ├── formatters.go       # Text formatting helpers
│   ├── sizing.go           # Terminal size calculations
│   └── validation.go       # Input validation (URL, branch names)
└── banner/                  # ASCII art banners
```

---

## Critical Rules

### NEVER Modify Application State from Goroutines
```go
// ❌ WRONG
func worker() {
    a.gitState = newState  // Race condition!
}

// ✅ CORRECT
func worker() tea.Cmd {
    return func() tea.Msg {
        return StateChangedMsg{newState}  // Immutable message
    }
}
```

### ALWAYS Use OutputBuffer for Streaming
```go
// ❌ WRONG (not thread-safe)
a.consoleOutput = append(a.consoleOutput, line)

// ✅ CORRECT (thread-safe)
output := ui.GetBuffer()
output.Append(line)
```

### ALWAYS Capture State in Closures for Async Ops
```go
// ❌ WRONG (reads Application fields from goroutine)
func (a *Application) execute() tea.Cmd {
    return func() tea.Msg {
        exec.Command("git", "clone", a.cloneURL).Run()  // Race!
    }
}

// ✅ CORRECT (closure captures value)
func (a *Application) execute() tea.Cmd {
    url := a.cloneURL  // Capture before goroutine starts
    return func() tea.Msg {
        exec.Command("git", "clone", url).Run()
    }
}
```

### ALWAYS Validate Git State After Operations
```go
case GitOperationMsg:
    if msg.Success {
        // ✅ Reload state
        if newState, err := git.DetectState(); err == nil {
            a.gitState = newState
        }
        // Menu regenerates based on new state
        a.menuItems = a.GenerateMenu()
    }
```

### Repository Config (`~/.config/tit/repo.toml`)

**Stores per-repo settings:**
```toml
[repository]
initialized = true
repositoryPath = "/path/to/repo"
canonBranch = "main"
lastWorkingBranch = "dev"
```

**Load on app start, save after init/clone/branch selection.**

---

## Design Patterns in Use

### Mode-Based Rendering (View Composition)
```go
func (a *Application) View() string {
    switch a.mode {
    case ModeMenu:
        return ui.RenderMenu(...)
    case ModeInitializeBranches:
        return ui.RenderBranchInput(...)
    case ModeConsole:
        return ui.RenderConsole(...)
    }
}
```

### Dynamic Dispatch (Menu → Handler)
```go
// Menu generators: git.Operation → MenuGenerator
menuGenerators[git.NotRepo] = menuNotRepo
menuGenerators[git.Conflicted] = menuConflicted

// Dispatchers: MenuItem.ID → dispatcher func
menuDispatchers["init"] = dispatchInit
menuDispatchers["clone"] = dispatchClone
```

### Registry Pattern (Key Handlers)
```go
// Built once, cached, fast lookup
handlers := map[AppMode]map[string]func(*Application) (tea.Model, tea.Cmd)
handlers[ModeMenu]["enter"] = handleMenuEnter
handlers[ModeMenu]["j"] = handleMenuDown
```

---

## Known Quirks

1. **Clone to CWD uses `git init` approach:**
   - `git clone <url> .` fails if ANY files exist (including `.DS_Store`)
   - Solution: `git init && git remote add origin <url> && git fetch && git checkout`

2. **ModeInput is deprecated:**
   - Old generic input mode being phased out
   - New workflows use dedicated modes (ModeCloneURL, ModeInitializeBranches)

3. **No undo for async operations:**
   - ESC aborts in-progress operation but can't undo completed ones
   - User must manually revert git state if needed

4. **Config is repo-local but stored globally:**
   - `~/.config/tit/repo.toml` uses `repositoryPath` as key
   - Multiple repos = multiple config entries (future feature)

---

## Debugging Tips

**Check git state:**
```go
fmt.Printf("State: %+v\n", a.gitState)
// Shows: WorkingTree, Timeline, Operation, Remote
```

**Check current mode:**
```go
fmt.Printf("Mode: %s\n", a.mode.String())
```

**Inspect output buffer:**
```go
lines := ui.GetBuffer().Lines()
fmt.Printf("Buffer: %d lines\n", len(lines))
```

**Git detection failures:**
- `DetectState()` returns error if not in repo
- App creates `State{Operation: NotRepo}` fallback
- Menu shows Init/Clone options

---

## Related Documentation

- `INIT-WORKFLOW.md` - Complete init flow walkthrough
- `COLORS.md` - Theme system color reference
- `SESSION-LOG.md` - Development session history
- `README.md` - Basic build/run instructions
