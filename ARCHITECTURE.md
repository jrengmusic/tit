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
```

**Rendering flow:**
```
RenderStateHeader()
    ↓
Looks up WorkingTree state info via stateinfo map
    ↓
Calls Description(ahead, behind) function
    ↓
Function returns StateDescriptions[key] formatted with counts
    ↓
Display: "Branch: main | Dirty | You have 2 unsynced commits."
```

---

## Confirmation Dialog System

**Purpose:** Centralize all confirmation dialog text (titles + explanations) for safe destructive operations.

**Implementation** (`internal/app/messages.go` DialogMessages SSOT):

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
A: DiffPane is overkill for basic conflict resolution. renderGenericContentPane is simpler (line numbers + highlighting). DiffPane ready for advanced features (Phase 5).

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

✅ **Or use helper utility (Implemented in Session 59):**
```go
// RIGHT - Reusable, testable
result := ui.CenterAlignLine(text, width)
```

**Status in codebase (Session 59 - Complete):**
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

✅ **Extract to utility helper (Implemented in Session 59):**
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

**Status (Session 59 - Complete):**
- ✅ `convertGitFilesToUIFileInfo()` implemented in handlers.go (line 27-39)
- ✅ Both call sites updated (handleFileHistoryUp, handleFileHistoryDown)
- ✅ ~20 lines of duplication eliminated

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
| `internal/app/operationsteps.go` | OperationStep constants SSOT (all async operation names) |
| `internal/app/dispatchers.go` | Menu item → mode transitions |
| `internal/app/handlers.go` | Input handlers (enter, ESC, text input, etc) |
| `internal/app/keyboard.go` | Key handler registry construction |
| `internal/app/messages.go` | Custom tea.Msg types & SSOT maps (prompts, errors, dialogs) |
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

**Consolidated builder (Implemented in Session 59):**
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

**Refactored implementations (Session 59 - Complete):**
- ✅ `buildHistoryStatusBar()` (history.go:158) - Uses BuildStatusBar
- ✅ `buildFileHistoryStatusBar()` (filehistory.go:218) - Uses BuildStatusBar
- ✅ `buildDiffStatusBar()` (filehistory.go:259) - Uses BuildStatusBar (with visual mode special case)
- ✅ `buildGenericConflictStatusBar()` (conflictresolver.go:182) - Uses BuildStatusBar

**Benefits realized:**
- ~50 lines of duplication eliminated
- Consistent centering logic across all panes
- Theme color changes propagate to all status bars

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
1. Validate bounds (no nil access)
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

## Related Documentation

- `SPEC.md` - User-facing behavior specification
- `IMPLEMENTATION_PLAN.md` - Phase-by-phase feature roadmap
- `SESSION-LOG.md` - Development history with session notes
- `COLORS.md` - Theme system color reference
