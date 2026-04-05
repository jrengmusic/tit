# TIT Codebase Navigation Map

**Visual guide to codebase structure, dependencies, and key patterns**

---

## 📦 Package Structure

```
tit/
├── cmd/tit/
│   └── main.go                    ← Program entry point
│
├── internal/
│   ├── app/                       ← Application logic & UI coordination
│   │   ├── app.go                 ← Application struct, Update(), View()
│   │   ├── modes.go               ← AppMode enum (Menu, Console, Input, etc.)
│   │   ├── menu.go                ← Menu generation (state → menu items)
│   │   ├── menu_items.go          ← SSOT: All menu items defined here
│   │   ├── messages.go            ← String constants (prompts, errors, hints)
│   │   ├── git_handlers.go        ← Git operation result handlers
│   │   ├── conflict_handlers.go   ← Conflict resolver handlers
│   │   ├── dispatchers.go         ← Action dispatchers (menu → handler)
│   │   ├── history_cache.go       ← History metadata cache preloading
│   │   ├── menu_builders.go       ← Menu helpers (generator map, builder)
│   │   ├── dirty_state.go         ← Dirty operation state tracking
│   │   ├── conflict_state.go      ← Conflict resolver state
│   │   ├── handlers_git_result.go ← Extracted git operation result handlers
│   │   ├── op_rebase.go           ← Rebase cmd and handler functions
│   │   ├── // State struct files (extracted from Application God Object)
│   │   ├── input_state.go          ← Input field management state (7 fields)
│   │   ├── cache_manager.go        ← Cache lifecycle state (14 fields)
│   │   ├── async_state.go          ← Async operation state (3 fields)
│   │   ├── workflow_state.go       ← Workflow and clone state (7 fields)
│   │   ├── environment_state.go    ← Git environment and setup state (5 fields)
│   │   ├── picker_state.go         ← Picker UI state (3 fields)
│   │   ├── console_state.go        ← Console output state (3 fields)
│   │   ├── activity_state.go       ← Activity tracking state (4 fields)
│   │   ├── dialog_state.go         ← Dialog UI state (2 fields)
│   │   └── time_travel_state.go    ← Time travel operation state (2 fields)
│   │
│   ├── git/                       ← Git operations & state detection
│   │   ├── state.go               ← State detection (WorkingTree, Timeline, etc.)
│   │   ├── types.go               ← State enums & type definitions
│   │   ├── execute.go             ← Git command execution
│   │   ├── exec_*.go              ← Per-operation git command files
│   │   ├── init.go                ← Repository initialization
│   │   ├── dirtyop.go             ← Dirty operation (stash/restore)
│   │   └── messages.go            ← Git operation message types
│   │
│   ├── ui/                        ← User interface rendering
│   │   ├── theme.go               ← 🌟 SSOT: Colors & styling
│   │   ├── sizing.go              ← 🌟 SSOT: Terminal dimensions
│   │   ├── layout.go              ← Screen layout (banner, header, content, footer)
│   │   ├── menu.go                ← Menu rendering
│   │   ├── box.go                 ← Box drawing utilities
│   │   ├── history.go             ← History split-pane rendering
│   │   ├── filehistory.go         ← File(s) history 3-pane rendering
│   │   ├── conflictresolver.go    ← Conflict resolver N-column rendering
│   │   ├── textpane_render.go     ← Text/diff pane rendering with scrolling
│   │   ├── textpane_input.go      ← Text/diff pane input handling
│   │   ├── listpane.go            ← List pane with selection
│   │   ├── confirmation.go        ← Confirmation dialog
│   │   ├── console.go             ← Streaming output console
│   │   ├── buffer.go              ← Thread-safe output buffer
│   │   ├── formatters.go          ← Text formatting helpers
│   │   ├── textinput.go           ← Text input rendering
│   │   ├── validation.go          ← Input validation (URLs, branches)
│   │   └── assets/                ← Braille/SVG assets
│   │
│   ├── config/                    ← Configuration & persistence
│   │   └── stash.go               ← Stash list management
│   │
│   ├── banner/                    ← ASCII art banners
│   │   ├── braille.go             ← Braille character rendering
│   │   └── svg.go                 ← SVG parsing
│   │
│   └── .DS_Store
│
├── go.mod                         ← Dependencies
├── go.sum
└── README.md, *.md                ← Documentation
```

---

## 🔗 Data Flow: Core Operations

### Menu Flow (State → Menu Items → Handler)
```
git.State (WorkingTree, Timeline, Operation)
    ↓
app.GenerateMenu()
    ├─ menuGenerators[State.Operation]()
    │   ├─ menuNotRepo() / menuConflicted() / menuNormal() / ...
    │   └─ getHistoryItemsWithCacheState() [adds progress indicators]
    └─ Returns: []MenuItem
    
Menu items render via ui.RenderMenuWithSelection()
    ↓
User selects menu item
    ↓
app.handleMenuEnter()
    ├─ Find dispatcher by MenuItem.ID
    └─ dispatcher(app) → (tea.Model, tea.Cmd)
        └─ Returns: mode transition + async operation
```

### Git Operation Flow (Handler → Command → Async → Result)
```
User selects action (e.g., "Commit")
    ↓
dispatchCommit(app)
    └─ Sets mode = ModeConsole
    └─ Returns: app.cmdCommit(message) [tea.Cmd]
    
Bubble Tea executes command in worker goroutine
    ↓
cmdCommit(message)
    └─ Executes git command
    └─ Streams output to buffer
    └─ Returns: GitOperationMsg{Step: "commit", Success: bool, ...}
    
Bubble Tea delivers GitOperationMsg to Update()
    ↓
app.Update(GitOperationMsg)
    └─ handleGitOperation(msg)
    ├─ Check success/conflicts
    ├─ Rebuild cache if needed
    ├─ Detect new git state
    └─ Transition to: ModeMenu or ModeConflictResolve
    
app.View() renders new state
    └─ Cache is already updated
    └─ Menu shows correct options
```

### Time Travel Flow
```
User selects "Time travel to M2"
    ↓
dispatchTimeTravel(app)
    └─ Shows confirmation dialog
    
User confirms
    ↓
executeTimeTravelClean() / executeTimeTravelWithDirtyTree()
    └─ Sets restoreTimeTravelInitiated = true
    └─ Writes .git/TIT_TIME_TRAVEL marker file
    └─ Returns: git.TimeTravelCheckoutMsg
    
app.handleTimeTravelCheckout()
    ├─ Detects git state = TimeTraveling
    ├─ Rebuilds cache (history still available)
    └─ Transitions to ModeMenu
    
Menu generation now sees Operation = TimeTraveling
    └─ menuTimeTraveling() shows: History, View diff, Merge back, Return
    
User selects "Return to main"
    └─ executeTimeTravelReturn()
    └─ Checks for marker file
    └─ Restores stash if dirty
    └─ Merges time-traveled commits back to main
    └─ Returns git.TimeTravelReturnMsg
    
app.handleTimeTravelReturn()
    └─ Removes marker file
    └─ Resets state back to Normal
    └─ Clears time travel state
```

---

## 📊 Key Data Structures

### Git State Tuple
```go
type State struct {
    WorkingTree WorkingTreeState  // Clean | Dirty
    Timeline    TimelineState     // InSync | Ahead | Behind | Diverged | (empty)
    Operation   OperationState    // Normal | Conflicted | Merging | Rebasing | DirtyOp | TimeTraveling | NotRepo
    Remote      RemoteState       // NoRemote | HasRemote
}
```

### Application State
```go
type Application struct {
    // Core state
    gitState        git.State
    mode            AppMode
    
    // Time travel state
    timeTravelState TimeTravelState
    
    // Extracted state structs (21 fields total in Application)
    inputState      InputState      // Input field management (7 fields)
    cacheManager    CacheManager    // Cache lifecycle (14 fields)
    asyncState      AsyncState      // Async operation state (3 fields)
    workflowState   WorkflowState   // Workflow and clone state (7 fields)
    environmentState EnvironmentState // Git environment and setup state (5 fields)
    pickerState     PickerState     // Picker UI state (3 fields)
    consoleState    ConsoleState    // Console output state (3 fields)
    activityState   ActivityState   // Activity tracking state (4 fields)
    dialogState     DialogState     // Dialog UI state (2 fields)
    
    // Remaining direct fields (21 total after extraction)
    selectedMenuIndex int
    menuItems         []MenuItem
    footerHint        string
    width             int
    height            int
    // ... other UI state fields
}
```

### Theme (Single Source of Color)
```go
type Theme struct {
    // Background
    MainBackgroundColor    string
    
    // Text colors
    ContentTextColor       string
    LabelTextColor         string
    DimmedTextColor        string
    AccentTextColor        string
    HighlightTextColor     string
    
    // UI elements
    BoxBorderColor         string
    MenuSelectionBackground string
    
    // Conflict-specific
    ConflictMarkerColor    string
    ConflictResolved       string
}
```

---

## 🎯 Finding Code by Task

### "I need to add a new menu item"
1. **Define in SSOT:** `internal/app/menu_items.go` (add to `MenuItems` map)
2. **Generate in menu:** `internal/app/menu.go` (add to appropriate `menu*()` function)
3. **Dispatch action:** `internal/app/dispatchers.go` (add to `actionDispatchers` map)
4. **Handle action:** `internal/app/handlers_*.go` (add handler function)
5. **Register key:** `internal/app/keyboard.go` (add to mode handlers)

### "I need to add a new git operation"
1. **Define command:** `internal/app/op_*.go` (create `cmd*()` function in appropriate op file)
2. **Handle result:** `internal/app/handlers_git_result.go` (add case in `handleGitOperation()`)
3. **Add messages:** `internal/app/messages.go` (add error/success text)

### "I need to add a new state indicator"
1. **Enum:** `internal/git/types.go` (add constant)
2. **Detection:** `internal/git/state.go` (add to `Detect*()` function)
3. **Menu:** `internal/app/menu.go` (add menu generator)
4. **Display:** `internal/ui/layout.go` (update header/state display)

### "I need to modify colors"
1. **SSOT source:** `internal/ui/theme.go`
2. All references use `theme.FieldName` → one change everywhere

### "I need to change a message"
1. **SSOT source:** `internal/app/messages.go` (appropriate map)
2. All usages via key lookup → one change everywhere

### "I need to add error handling"
1. **Message:** `internal/app/messages.go` (ErrorMessages map)
2. **Usage:** Wherever error occurs: `buffer.Append(fmt.Sprintf(ErrorMessages[key], err), ui.TypeStderr)`

---

## 🔄 Pattern Reference

### How Menu Selection Works
```
app.menuItems[]         ← Generated from git state
                        
User presses ↓ / ↑     → app.handleMenuDown/Up() changes selectedIndex
                        
User presses Enter      → app.handleMenuEnter()
                        ├─ Find: app.menuItems[selectedIndex].ID
                        ├─ Call: actionDispatchers[id](app)
                        └─ Returns: (tea.Model, tea.Cmd)
```

### How State Drives UI
```
git.State.Operation
    ├─ NotRepo        → menuNotRepo()     → [Init, Clone]
    ├─ Normal         → menuNormal()      → [Commit, View history, ...]
    ├─ Conflicted     → menuConflicted()  → [Resolve, Abort]
    ├─ Merging        → menuOperation()   → [Continue, Abort]
    ├─ DirtyOperation → menuDirtyOperation() → [View status, Abort]
    └─ TimeTraveling  → menuTimeTraveling()  → [History, Merge back, Return]
    
app.View() renders:
    ├─ Header: current branch + state indicator
    ├─ Content: menu items specific to state
    └─ Footer: context-sensitive hints
```

### How Cache Works
```
App startup / After git-changing operation:
    └─ cmdPreloadHistoryMetadata()
    │  └─ Builds app.historyMetadataCache
    │  └─ Sends CacheProgressMsg to UI
    │  └─ UI shows "Building... 5/30"
    │  └─ On complete: menu items enabled
    │
    └─ cmdPreloadFileHistoryDiffs()
       └─ Builds app.fileHistoryDiffCache
       └─ Builds app.fileHistoryFilesCache
    
When user views history:
    └─ cache already populated
    └─ renderHistoryListPane() reads from cache
    └─ No "loading" state needed
```

---

## 🚨 Critical Files to Protect

| File | Why Critical | Change Impact |
|------|-------------|----------------|
| `app.go` | Main loop (Update, View, Init) | Breaks entire app |
| `menu_items.go` | Menu SSOT | Menu items missing/duplicated |
| `theme.go` | Color SSOT | Visual regression |
| `messages.go` | String SSOT | UX text wrong |
| `git/state.go` | State detection | Wrong menu options shown |
| `git/types.go` | State enums | Compiler errors everywhere |

---

## 📌 SSOT Locations (Single Source of Truth)

| What | Where | Update Impact |
|------|-------|---------------|
| **Menu items** | `menu_items.go` | All items affected |
| **Colors** | `ui/theme.go` | All UI colors |
| **Terminal size** | `ui/sizing.go` | All pane sizes |
| **Messages** | `app/messages.go` | All user-facing text |
| **State enums** | `git/types.go` | All state-dependent logic |
| **Status bar** | `ui/layout.go` | All status bars |

---

## 🎯 Decision Points (Where Logic Lives)

| Decision | Location | Examples |
|----------|----------|----------|
| Which menu to show | `menu.go` (menuGenerators) | If Operation=Conflicted, show resolve/abort |
| How to handle action | `*handlers.go` | If confirm fork pull, show dirty protocol |
| How to render UI | `ui/*.go` | If 2 panes, use JoinHorizontal |
| What's the git state | `git/state.go` | Detect if repo dirty/ahead/behind |
| What error message | `messages.go` | Map error key to user text |

---

## 🔍 Code Search Patterns

### Find all handlers for a feature
```bash
grep -r "handleCommit\|dispatchCommit\|cmdCommit" internal/
```

### Find where a constant is used
```bash
grep -r "ModeMenu\|ModeConsole" internal/
```

### Find all error messages
```bash
grep "ErrorMessages\[" internal/
```

### Find status bar builders
```bash
grep "func build.*StatusBar" internal/ui/
```

### Find conflicting naming
```bash
grep "func.*execute.*Command\|execute.*Workflow" internal/app/
# Should be cmd*, not execute*
```

---

## 🏗️ Extension Points (Safe to Add)

### Adding a new menu state
1. Add to `OperationState` enum in `git/types.go`
2. Add detection in `git/state.go`
3. Add menu generator in `app/menu.go`
4. Add to `menuGenerators` map

### Adding a new git operation
1. Create `cmd*()` in `op_*.go` (appropriate op file or new one)
2. Add dispatcher in `dispatchers.go`
3. Add handler in `handlers_git_result.go`
4. Add messages in `messages.go`
5. Add to menu in `menu.go`

### Adding a new rendering component
1. Create `Render*()` in `ui/*.go`
2. Use SSOT sizing from `sizing.go`
3. Use SSOT colors from `theme.go`
4. Call from `layout.go`

---

## 📈 Code Metrics

| Metric | Value |
|--------|-------|
| **App package** | ~3000 lines (23 files) |
| **Git package** | ~600 lines (6 files) |
| **UI package** | ~2200 lines (20 files) |
| **Total** | ~5800 lines |
| **Functions in app** | ~60 exported, ~80 private |
| **Structs in app** | ~15 major types |
| **Menu items (SSOT)** | 27 total |
| **Color definitions** | 20 in theme |

---

## ✅ Architecture Quality

| Aspect | Status | Notes |
|--------|--------|-------|
| **SSOT enforcement** | ✅ Good | MenuItem, Theme, Messages centralized |
| **Layer separation** | ✅ Clean | app → git → ui with clear boundaries |
| **Async pattern** | ✅ Correct | cmd* pattern throughout, no blocking |
| **Error handling** | ⚠️ Mostly good | Some inconsistency in recovery |
| **Code duplication** | ⚠️ Moderate | Status bars, shortcut styles repeated |
| **Navigation** | ⚠️ Fair | Related code split across files |
| **Naming** | ✅ Good | Mostly verb-noun, clear intent |

---

**Use this map to find code quickly and understand relationships between files.**

For detailed refactoring opportunities, see `CODEBASE-REFACTORING-AUDIT.md`.
