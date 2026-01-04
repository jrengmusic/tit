# TIT Init Workflow - Complete Handler Implementation

## Flow Overview

```
App Start (NotRepo)
    ↓
User selects "Initialize"
    ↓ dispatchInit()
ModeInitializeLocation (menu with 2 choices)
    ↓
User presses "1" or "2"
    ↓
handleInitLocationChoice1/Choice2
    ├─ Choice1: Set initRepositoryPath = cwd
    └─ Choice2: Ask for directory name in ModeInput
    ↓
[If Choice2: User enters repo name]
    ↓
handleInputSubmitSubdirName
    ├─ Construct initRepositoryPath = cwd/reponame
    └─ Transition to canon branch input
    ↓
ModeInitializeCanonBranch (text input)
    ↓
User enters "main" (or custom name)
    ↓
handleCanonBranchSubmit
    ├─ Store initCanonBranch = "main"
    └─ Transition to working branch input
    ↓
ModeInitializeWorkingBranch (text input)
    ↓
User enters "dev" (or custom name)
    ↓
handleWorkingBranchSubmit
    ├─ Store initWorkingBranch = "dev"
    └─ Launch executeInitWorkflow() → tea.Cmd
    ↓
executeInitWorkflow (WORKER THREAD)
    ├─ git init <repo path>
    ├─ cd <repo path>
    ├─ git checkout -b main
    ├─ git checkout -b dev
    ├─ Save RepoConfig to ~/.config/tit/repo.toml
    └─ Return GitOperationMsg
    ↓
Update() receives GitOperationMsg("init")
    ├─ Reload git.DetectState()
    ├─ Reset init state fields
    ├─ Regenerate menu (now repo exists)
    ├─ Switch to ModeMenu
    └─ Show success message
    ↓
ModeMenu (normal operation)
```

## Handler Implementations

### Step 1: Location Choice - handleInitLocationChoice1()
**Triggered:** User presses "1" or menu Enter on "Initialize current directory"

```go
// Sets repository path to current directory
app.initRepositoryPath = cwd

// Transitions to canon branch input
app.mode = ModeInitializeCanonBranch
app.inputValue = "main"           // Default suggestion
app.inputPrompt = "Canon branch name:"
```

**Result:** Input field for canon branch with "main" pre-filled

---

### Step 2: Subdirectory Name (Optional) - handleInputSubmitSubdirName()
**Triggered:** User presses "2" at location menu, then presses Enter on directory name

```go
// User was in ModeInput with inputAction="init_subdir_name"
// Validates inputValue != ""

app.initRepositoryPath = cwd + "/" + inputValue

// Transition to canon branch (same as Choice1)
app.mode = ModeInitializeCanonBranch
app.inputValue = "main"
app.inputPrompt = "Canon branch name:"
```

**Result:** Input field for canon branch with "main" pre-filled

---

### Step 3: Canon Branch - handleCanonBranchSubmit()
**Triggered:** User presses Enter on canon branch input

```go
// Validates inputValue != ""
app.initCanonBranch = inputValue  // e.g., "main"

// Transition to working branch input
app.mode = ModeInitializeWorkingBranch
app.inputValue = "dev"             // Default suggestion
app.inputPrompt = "Working branch name:"
```

**Result:** Input field for working branch with "dev" pre-filled

---

### Step 4: Working Branch - handleWorkingBranchSubmit()
**Triggered:** User presses Enter on working branch input

```go
// Validates inputValue != ""
app.initWorkingBranch = inputValue  // e.g., "dev"

// Launch async git operations
return app, app.executeInitWorkflow()
```

**Note:** Handler returns immediately (before git ops complete)

---

### Step 5: Git Operations (WORKER) - executeInitWorkflow()
**Runs:** In separate goroutine (tea.Cmd)

```go
// WORKER THREAD - All blocking I/O here

1. git init <repo path>
   └─ Creates .git directory

2. cd <repo path> + git checkout -b <canon name>
   └─ Creates and checks out canon branch

3. git checkout -b <working name>
   └─ Creates and checks out working branch

4. Save RepoConfig to ~/.config/tit/repo.toml
   ├─ repo.initialized = true
   ├─ repo.repositoryPath = <path>
   ├─ repo.canonBranch = "main"
   └─ repo.lastWorkingBranch = "dev"

5. Return GitOperationMsg{Step:"init", Success:true, Output:"..."}
```

**Result:** App receives GitOperationMsg on UI thread

---

### Step 6: Completion - Update() GitOperationMsg Handler
**Runs:** UI thread, after worker completes

```go
case GitOperationMsg:
  if msg.Step == "init" && msg.Success {
    // Reload git state (DetectState now succeeds)
    app.gitState = git.DetectState()
    
    // Reset init workflow state
    app.initRepositoryPath = ""
    app.initCanonBranch = ""
    app.initWorkingBranch = ""
    
    // Return to menu with success message
    app.mode = ModeMenu
    app.menuItems = app.GenerateMenu()
    app.footerHint = msg.Output
  } else if !msg.Success {
    // Stay in current mode, show error
    app.footerHint = msg.Error
  }
```

**Result:** Menu now shows repo-specific options (branches exist, can push/pull, etc.)

---

## Key Design Patterns

### State Progression
- **Input Modes:** Each step stores partial state before moving to next
- **Atomic:** All git ops run together; partial failures show user error message
- **Idempotent:** Running init twice creates branches (git handles existence)

### Thread Safety
- **App state fields:** Set in UI thread only
- **executeInitWorkflow():** Pure function, captures `a.init*` fields in closure
- **Worker goroutine:** No access to Application struct (read-only closure vars)
- **GitOperationMsg:** Simple struct, safe to return from goroutine

### Error Handling
- **Validation:** Canon/working branch names checked before git ops
- **Git errors:** Caught at each step, returned as GitOperationMsg.Error
- **UI feedback:** Error messages displayed in footer, user stays in mode to retry

### Extensibility
- **GitOperationMsg.Step:** Allows multiple async operations (push, pull, etc.)
- **inputAction field:** Enables routing different input modes to different handlers
- **Modular handlers:** Each step is independent, can be reused or composed

---

## Testing Checklist

- [x] **Build:** No compilation errors
- [x] **Handlers compile:** All init handlers implemented
- [x] **State fields:** initRepositoryPath, initCanonBranch, initWorkingBranch
- [x] **Mode transitions:** Choice1 → CanonBranch → WorkingBranch → Menu
- [x] **Git ops:** init + branch creation + config save (simulated)
- [ ] **UI rendering:** ModeInitializeLocation menu displays correctly
- [ ] **Input rendering:** ModeInitializeCanonBranch shows text input with prompt
- [ ] **Cursor navigation:** Left/Right/Home/End work in input modes
- [ ] **Character input:** Typing updates inputValue + inputCursorPosition
- [ ] **Validation:** Empty names trigger error message
- [ ] **Success:** Branches created, config saved, menu regenerates
- [ ] **Failure:** Error message shown, user can retry without quitting
- [ ] **Menu state:** Post-init menu shows repo-specific options

---

## Next Steps

1. **Manual UI Test:** Run app in empty directory, walk through complete flow
2. **Edge Cases:** Test with special characters, long names, invalid paths
3. **Cleanup:** Implement "Cancel" handler (ESC) to exit init flow
4. **Config Load:** When app starts in repo, load RepoConfig and set gitState
5. **Subsequent Modes:** Implement handlers for push, pull, commit, etc.

