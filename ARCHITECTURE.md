# TIT Architecture Guide

## Overview

TIT (Git Timeline Interface) is a state-driven terminal UI for git repository management. It follows a clean, event-driven architecture based on Bubble Tea's Model-View-Update pattern.

**Core Principle:** Git state determines UI state. Operations are always safe and abortable.

---

## Four-Axis State Model

Every moment in TIT is described by exactly 4 git axes:

```go
type State struct {
    WorkingTree      git.WorkingTree // Clean | Dirty
    Timeline         git.Timeline    // InSync | Ahead | Behind | Diverged | NoRemote
    Operation        git.Operation   // NotRepo | Normal | Conflicted | Merging | Rebasing | DirtyOperation
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
| ModeConflictResolve | N-column parallel conflict resolution | ↑↓ (nav/scroll), TAB (cycle), SPACE (mark), ENTER (apply) |
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

- `menuNotRepo()` - Init/Clone (not in repo)
- `menuConflicted()` - Resolve/Abort (conflicts detected)
- `menuOperation()` - Continue/Abort (merge/rebase in progress)
- `menuNormal()` - Full menu (normal state)
  - `menuWorkingTree()` - Commit (when Dirty)
  - `menuTimeline()` - Push/Pull based on Timeline
  - `menuHistory()` - Commit history browser

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
- Advanced diff viewer (ready for Phase 5)
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
- Returns to ModeMenu without applying changes

### Critical Design Decisions

**Q: Why not use lipgloss for border-free joining?**
A: Each pane needs borders for visual separation. Full borders + JoinHorizontal is the correct pattern (matches old-tit).

**Q: Why radio buttons instead of checkboxes?**
A: Conflict resolution requires choosing ONE version per file. Radio button enforces this constraint.

**Q: Why shared selection in top row but independent scrolling in bottom row?**
A: User needs to compare the SAME file across all columns. Shared selection keeps all columns synchronized. Bottom row needs independent scrolling for long files.

**Q: Why not use DiffPane for content rendering?**
A: DiffPane is overkill for basic conflict resolution. renderGenericContentPane is simpler (line numbers + highlighting). DiffPane ready for advanced features (Phase 5).

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
| `internal/app/menuitems.go` | MenuItems SSOT map (all menu definitions) |
| `internal/app/menubuilder.go` | MenuItemBuilder fluent API (for separators) |
| `internal/app/dispatchers.go` | Menu item → mode transitions |
| `internal/app/handlers.go` | Input handlers (enter, ESC, text input, etc) |
| `internal/app/keyboard.go` | Key handler registry construction |
| `internal/app/messages.go` | Custom tea.Msg types |
| `internal/app/confirmationhandlers.go` | Confirmation dialog system and handlers |
| `internal/app/conflictstate.go` | Conflict resolution state struct |
| `internal/app/conflicthandlers.go` | Conflict resolution keyboard handlers |
| `internal/git/state.go` | State detection from git commands |
| `internal/git/execute.go` | Command execution with streaming |
| `internal/ui/layout.go` | RenderLayout() main view composer |
| `internal/ui/theme.go` | Color system with semantic names |
| `internal/ui/buffer.go` | OutputBuffer thread-safe ring buffer |
| `internal/ui/console.go` | RenderConsoleOutput() component |
| `internal/ui/confirmation.go` | ConfirmationDialog component |
| `internal/ui/conflictresolver.go` | N-column parallel conflict resolution UI |
| `internal/ui/listpane.go` | Reusable list pane with checkboxes and scrolling |
| `internal/ui/diffpane.go` | Diff viewer with line numbers and cursor |
| `internal/ui/menu.go` | RenderMenuWithHeight() component |
| `internal/ui/validation.go` | Input validation (URLs, directory names) |

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
