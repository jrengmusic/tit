# Time Travel Implementation Plan

**Scope:** Incremental, testable phases. Each phase deployable independently.
**Goal:** Full time travel feature with 100% safety guarantee on work preservation.

---

## Phase Overview

| Phase | Feature | Commits | Stash | Merge | Conflicts | Lines |
|-------|---------|---------|-------|-------|-----------|-------|
| **0** | Pre-startup restoration (safety net) | ❌ | ✅ | ❌ | ❌ | 100 |
| **1** | Basic time travel (clean tree) | ✅ | ❌ | ❌ | ❌ | 200 |
| **2** | Dirty tree handling | ✅ | ✅ | ❌ | ❌ | 150 |
| **3** | Merge back (no conflicts) | ✅ | ✅ | ✅ | ❌ | 200 |
| **4** | Merge conflicts + resolution | ✅ | ✅ | ✅ | ✅ | 250 |
| **5** | Browse while time traveling | ✅ | ✅ | ✅ | ✅ | 100 |
| **6** | Return (discard changes) | ✅ | ✅ | ✅ | ✅ | 50 |
| **Total** | | | | | | ~1,050 |

---

## Phase 0: Pre-Startup Restoration (Safety Net)

**CRITICAL:** If TIT exits while time traveling, startup must restore original state.

**Code Changes:**

### 0.1 Add Startup Check

`internal/app/app.go` - in `New()` function, after git state detection but BEFORE menu generation:

```go
func New(/* ... */) *Application {
    // ... existing init ...
    
    // Detect git state
    gitState, err := git.DetectState()
    if err != nil {
        // Handle error
    }
    a.gitState = gitState
    
    // CRITICAL: Check for incomplete time travel from previous session
    if fileExists(".git/TIT_TIME_TRAVEL") {
        // User exited TIT while time traveling
        // Restore original branch immediately
        if err := a.restoreFromTimeTravel(); err != nil {
            // Show error, but don't block startup
            a.footerHint = "Warning: Could not restore time travel state"
        }
    }
    
    // Now proceed with normal startup
    // ...
}
```

### 0.2 Implement Restoration Function

`internal/app/app.go`:

```go
func (a *Application) restoreFromTimeTravel() error {
    // Read .git/TIT_TIME_TRAVEL marker
    ttInfo, err := git.LoadTimeTravelInfo()
    if err != nil {
        return err
    }
    
    buffer := ui.GetBuffer()
    buffer.Append("Restoring from incomplete time travel session...", ui.TypeStatus)
    
    // Step 1: Discard any changes made during time travel
    git.Execute("checkout", ".")
    git.Execute("clean", "-fd")
    
    // Step 2: Return to original branch
    result := git.Execute("checkout", ttInfo.OriginalBranch)
    if !result.Success {
        return fmt.Errorf("failed to checkout %s", ttInfo.OriginalBranch)
    }
    buffer.Append(fmt.Sprintf("Returned to %s", ttInfo.OriginalBranch), ui.TypeStatus)
    
    // Step 3: Restore original stashed work if any
    if ttInfo.OriginalStashID != "" {
        result := git.Execute("stash", "apply", ttInfo.OriginalStashID)
        if !result.Success {
            buffer.Append("Warning: Could not restore original work", ui.TypeStatus)
        } else {
            buffer.Append("Original work restored", ui.TypeStatus)
            git.Execute("stash", "drop", ttInfo.OriginalStashID)
        }
    }
    
    // Step 4: Clean up marker
    os.Remove(".git/TIT_TIME_TRAVEL")
    
    buffer.Append("Ready to continue. Press ESC to dismiss.", ui.TypeStatus)
    
    return nil
}
```

### 0.3 Update TimeTravelInfo Type

`internal/git/types.go`:

```go
type TimeTravelInfo struct {
    OriginalBranch    string      // Branch we came from
    OriginalHead      string      // Commit hash before time travel
    CurrentCommit     CommitInfo  // Current commit while traveling
    OriginalStashID   string      // If user had dirty tree: stash ID
}
```

**Acceptance Criteria (Phase 0):**
- [ ] On startup: check for `.git/TIT_TIME_TRAVEL` marker
- [ ] If found: automatically restore original branch
- [ ] Restore original stashed work if applicable
- [ ] Clean up marker file
- [ ] User sees status messages in console
- [ ] No menu shown until restoration complete

---

## Phase 1: Basic Time Travel (Clean Working Tree Only)

**Scope:** User on main at commit M1 (clean). Select old commit ABC. Time travel to ABC. Show menu. ESC returns to main.

**Prerequisites:** Phase 7 from SESSION-LOG already done (confirmation dialog, checkout working).

**Code Changes:**

### 1.1 Add TimeTravelInfo to State Model

`internal/git/types.go`:
```go
type TimeTravelInfo struct {
    OriginalBranch string  // e.g., "main"
    OriginalHash   string  // e.g., "abc1234..." (where we came from)
    CurrentCommit  CommitInfo  // Current commit while time traveling
}
```

### 1.2 Load TimeTravelInfo After Checkout

`internal/app/githandlers.go` - modify `handleTimeTravelCheckout()`:
```go
// After successful checkout, load time travel context
if msg.Success {
    a.gitState = state
    
    // Load time travel info from .git/TIT_TIME_TRAVEL
    ttInfo, err := git.LoadTimeTravelInfo()
    if err == nil {
        a.timeTravelInfo = ttInfo
    }
    
    // Stay in console until user ESC
    a.asyncOperationActive = false
    a.isExitAllowed = true
    
    buffer.Append("Time travel successful. Press ESC to return to menu.", ui.TypeStatus)
}
```

### 1.3 Update Header Display

`internal/ui/layout.go` - modify header rendering:
```go
func renderHeader(state *git.State, ttInfo *git.TimeTravelInfo, ...) string {
    if state.Operation == git.TimeTraveling && ttInfo != nil {
        shortHash := ttInfo.CurrentCommit.Hash[:7]
        daysAgo := calculateDaysAgo(ttInfo.CurrentCommit.Date)
        return fmt.Sprintf("🕐 TIME TRAVELING | Commit: %s (%s ago)", shortHash, daysAgo)
    }
    // ... normal header
}
```

### 1.4 Generate Time Travel Menu

`internal/app/menu.go` - add function:
```go
func (a *Application) menuTimeTraveling() []MenuItem {
    if a.gitState.Operation != git.TimeTraveling {
        return nil
    }
    
    branch := "main"  // Will get from a.timeTravelInfo.OriginalBranch
    if a.timeTravelInfo != nil {
        branch = a.timeTravelInfo.OriginalBranch
    }
    
    items := []MenuItem{
        GetMenuItem("time_travel_history"),
        GetMenuItem("time_travel_return"),
        GetMenuItem("time_travel_merge"),
    }
    
    // Customize labels with branch name
    for i := range items {
        if items[i].ID == "time_travel_return" {
            items[i].Label = fmt.Sprintf("⬅️  Return to %s", branch)
        } else if items[i].ID == "time_travel_merge" {
            items[i].Label = fmt.Sprintf("📦 Merge & return to %s", branch)
        }
    }
    
    return items
}
```

### 1.5 Register Time Travel Menu

`internal/app/menu.go` - update `GenerateMenu()`:
```go
func (a *Application) GenerateMenu() []MenuItem {
    // Priority 1: Operation state
    if a.gitState.Operation == git.TimeTraveling {
        return a.menuTimeTraveling()
    }
    // ... rest of states
}
```

### 1.6 Placeholder Handlers

`internal/app/handlers.go` - add stubs:
```go
func (a *Application) handleTimeTravelHistory(app *Application) (tea.Model, tea.Cmd) {
    // Phase 5: Browse history while time traveling
    return app, nil
}

func (a *Application) handleTimeTravelReturn(app *Application) (tea.Model, tea.Cmd) {
    // Phase 6: Return without merge
    return app, nil
}

func (a *Application) handleTimeTravelMerge(app *Application) (tea.Model, tea.Cmd) {
    // Phase 3+: Merge back to branch
    return app, nil
}
```

### 1.7 Register Handlers

`internal/app/app.go` - add to key registry:
```go
ModeMenu: NewModeHandlers().
    // ... existing ...
    On("l", a.handleMenuShortcut).  // Will route to time_travel_history
    On("r", a.handleMenuShortcut).  // Will route to time_travel_return
    On("m", a.handleMenuShortcut).  // Will route to time_travel_merge
    Build(),
```

**Acceptance Criteria (Phase 1):**
- [ ] Time travel to commit (clean tree) → shows "Time traveling..." console
- [ ] Console shows successful message
- [ ] ESC → menu appears with 3 time travel items
- [ ] Header shows 🕐 TIME TRAVELING | Commit: abc1234
- [ ] All 3 menu items have correct labels with branch name
- [ ] Binary builds clean, no warnings

---

## Phase 2: Dirty Tree Handling (Stash Before Time Travel)

**Scope:** User on main with dirty working tree. Dirty operation protocol → stash. Time travel. Menu shown.

**Code Changes:**

### 2.1 Modify Confirmation Handler

`internal/app/confirmationhandlers.go` - update `executeConfirmTimeTravel()`:
```go
func (a *Application) executeConfirmTimeTravel() (tea.Model, tea.Cmd) {
    a.confirmationDialog = nil
    commitHash := a.confirmContext["commit_hash"]
    currentBranchResult := git.Execute("rev-parse", "--abbrev-ref", "HEAD")
    if !currentBranchResult.Success {
        a.footerHint = "Failed to get current branch"
        return a, nil
    }
    originalBranch := strings.TrimSpace(currentBranchResult.Stdout)
    
    // Check working tree
    if a.gitState.WorkingTree == git.Dirty {
        // Dirty tree: show dirty operation protocol
        return a.executeTimeTravelWithDirtyProtocol(originalBranch, commitHash)
    } else {
        // Clean tree: go directly to checkout
        return a.executeTimeTravelClean(originalBranch, commitHash)
    }
}
```

### 2.2 Add Dirty Protocol Handler

`internal/app/confirmationhandlers.go` - new function:
```go
func (a *Application) executeTimeTravelWithDirtyProtocol(originalBranch, commitHash string) (tea.Model, tea.Cmd) {
    // Show dirty operation dialog
    a.mode = ModeConfirmation
    a.confirmType = "time_travel_dirty"
    a.confirmContext = map[string]string{
        "original_branch": originalBranch,
        "commit_hash": commitHash,
    }
    
    config := ui.ConfirmationConfig{
        Title: "You have uncommitted changes",
        Explanation: "Your changes will be stashed temporarily.\n" +
            "After time travel, you can merge them back or discard them.\n" +
            "Changes may conflict with the code at the target commit.",
        YesLabel: "Stash & continue",
        NoLabel: "Cancel",
        ActionID: "time_travel_dirty",
    }
    a.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &a.theme)
    
    return a, nil
}
```

### 2.3 Add Dirty Confirmation Handler

`internal/app/confirmationhandlers.go` - add to action maps:
```go
var confirmationActions = map[string]ConfirmationAction{
    // ... existing ...
    string(ConfirmTimeTravelDirty): (*Application).executeConfirmTimeTravelDirty,
}

var confirmationRejectActions = map[string]ConfirmationAction{
    // ... existing ...
    string(ConfirmTimeTravelDirty): (*Application).executeRejectTimeTravelDirty,
}
```

### 2.4 Implement Dirty Stash

`internal/app/confirmationhandlers.go`:
```go
func (a *Application) executeConfirmTimeTravelDirty() (tea.Model, tea.Cmd) {
    originalBranch := a.confirmContext["original_branch"]
    commitHash := a.confirmContext["commit_hash"]
    a.confirmationDialog = nil
    
    // Stash dirty changes
    stashResult := git.Execute("stash", "push", "-u", "-m", "TIT_TIME_TRAVEL_ORIG_WIP")
    if !stashResult.Success {
        a.footerHint = "Failed to stash changes"
        return a, nil
    }
    
    // Save original branch and stash ID for later
    stashListResult := git.Execute("stash", "list")
    if !stashListResult.Success {
        git.Execute("stash", "pop")  // Restore if stash list fails
        a.footerHint = "Failed to access stash"
        return a, nil
    }
    
    // Extract stash ID (should be stash@{0})
    stashID := ""
    lines := strings.Split(stashListResult.Stdout, "\n")
    for _, line := range lines {
        if strings.Contains(line, "TIT_TIME_TRAVEL_ORIG_WIP") {
            parts := strings.Fields(line)
            if len(parts) > 0 {
                stashID = parts[0]
                break
            }
        }
    }
    
    // Write time travel metadata
    err := git.WriteTimeTravelInfo(originalBranch, stashID)
    if err != nil {
        git.Execute("stash", "pop")  // Restore on failure
        a.footerHint = fmt.Sprintf("Failed to save time travel info: %v", err)
        return a, nil
    }
    
    // Proceed to checkout (clean tree now)
    return a.executeTimeTravelClean(originalBranch, commitHash)
}

func (a *Application) executeRejectTimeTravelDirty() (tea.Model, tea.Cmd) {
    a.confirmationDialog = nil
    return a.returnToMenu()
}
```

**Acceptance Criteria (Phase 2):**
- [ ] Dirty tree + time travel → dirty protocol dialog shown
- [ ] Cancel dirty protocol → menu returns (no stash made)
- [ ] Confirm dirty protocol → stash created, checkout proceeds
- [ ] After checkout, menu appears (tree is clean now)
- [ ] `.git/TIT_TIME_TRAVEL` contains: branch, stash@{0}
- [ ] `git stash list` shows stashed changes

---

## Phase 3: Merge Back (No Conflicts)

**Scope:** Time traveling, user at commit ABC. Select "Merge & return". Merge ABC → main. Show menu on main.

**Code Changes:**

### 3.1 Add Merge Confirmation

`internal/app/handlers.go`:
```go
func (a *Application) handleTimeTravelMerge(app *Application) (tea.Model, tea.Cmd) {
    if a.timeTravelInfo == nil {
        return app, nil
    }
    
    // Show confirmation
    app.mode = ModeConfirmation
    app.confirmType = "time_travel_merge_confirm"
    app.confirmContext = map[string]string{
        "original_branch": a.timeTravelInfo.OriginalBranch,
        "current_commit": a.timeTravelInfo.CurrentCommit.Hash,
    }
    
    config := ui.ConfirmationConfig{
        Title: fmt.Sprintf("Merge %s into %s", 
            a.timeTravelInfo.CurrentCommit.Hash[:7],
            a.timeTravelInfo.OriginalBranch),
        Explanation: "This will merge all changes from this commit into your branch.\n" +
            "If there are conflicts, you'll resolve them next.",
        YesLabel: "Merge",
        NoLabel: "Cancel",
        ActionID: "time_travel_merge_confirm",
    }
    app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)
    
    return app, nil
}
```

### 3.2 Add Merge Handler

`internal/app/confirmationhandlers.go`:
```go
var confirmationActions = map[string]ConfirmationAction{
    // ... existing ...
    "time_travel_merge_confirm": (*Application).executeTimeTravelMerge,
}

var confirmationRejectActions = map[string]ConfirmationAction{
    // ... existing ...
    "time_travel_merge_confirm": (*Application).executeRejectTimeTravelMerge,
}

func (a *Application) executeTimeTravelMerge() (tea.Model, tea.Cmd) {
    if a.timeTravelInfo == nil {
        return a.returnToMenu()
    }
    
    originalBranch := a.timeTravelInfo.OriginalBranch
    currentCommit := a.timeTravelInfo.CurrentCommit.Hash
    
    a.confirmationDialog = nil
    
    // Check for local changes (time travel changes)
    statusResult := git.Execute("status", "--short")
    hasLocalChanges := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""
    
    // Stash time travel changes if they exist
    if hasLocalChanges {
        stashResult := git.Execute("stash", "push", "-u", "-m", "TIT_TIME_TRAVEL_CHANGES")
        if !stashResult.Success {
            a.footerHint = "Failed to stash time travel changes"
            return a, nil
        }
    }
    
    // Transition to console
    a.asyncOperationActive = true
    a.asyncOperationAborted = false
    a.mode = ModeConsole
    a.outputBuffer.Clear()
    a.consoleState.Reset()
    a.footerHint = "Merging... (ESC to abort)"
    a.previousMode = ModeMenu
    a.previousMenuIndex = 0
    
    // Execute merge operation
    return a, git.ExecuteTimeTravelMerge(originalBranch, currentCommit, hasLocalChanges)
}

func (a *Application) executeRejectTimeTravelMerge() (tea.Model, tea.Cmd) {
    a.confirmationDialog = nil
    return a, nil  // Stay in time travel mode
}
```

### 3.3 Add Merge Git Operation

`internal/git/execute.go`:
```go
func ExecuteTimeTravelMerge(originalBranch, currentCommit string, hasLocalChanges bool) func() tea.Msg {
    return func() tea.Msg {
        buffer := ui.GetBuffer()
        buffer.Append("Checking out original branch...", ui.TypeStatus)
        
        // Checkout original branch
        checkoutResult := Execute("checkout", originalBranch)
        if !checkoutResult.Success {
            buffer.Append(fmt.Sprintf("Failed to checkout: %s", checkoutResult.Stderr), ui.TypeStderr)
            return TimeTravelMergeMsg{
                Success: false,
                Error: "Failed to checkout original branch",
                ConflictDetected: false,
            }
        }
        
        // Merge current commit
        buffer.Append(fmt.Sprintf("Merging %s...", currentCommit[:7]), ui.TypeStatus)
        mergeResult := Execute("merge", currentCommit, "--no-ff", "-m", 
            fmt.Sprintf("Time travel merge from %s", currentCommit[:7]))
        
        if !mergeResult.Success {
            // Check if conflict
            statusResult := Execute("status", "--short")
            hasConflicts := strings.Contains(statusResult.Stdout, "UU") || 
                           strings.Contains(statusResult.Stdout, "AA") ||
                           strings.Contains(statusResult.Stdout, "DD")
            
            if hasConflicts {
                buffer.Append("Merge conflicts detected. Resolve in conflict resolver.", ui.TypeStderr)
                return TimeTravelMergeMsg{
                    Success: false,
                    Error: "Merge conflicts",
                    ConflictDetected: true,
                }
            }
            
            buffer.Append(fmt.Sprintf("Merge failed: %s", mergeResult.Stderr), ui.TypeStderr)
            return TimeTravelMergeMsg{
                Success: false,
                Error: mergeResult.Stderr,
                ConflictDetected: false,
            }
        }
        
        // Apply stashed time travel changes
        if hasLocalChanges {
            buffer.Append("Applying your time travel changes...", ui.TypeStatus)
            applyResult := Execute("stash", "apply")
            if !applyResult.Success {
                buffer.Append("Note: Could not apply time travel changes (may be redundant)", ui.TypeStatus)
                Execute("stash", "drop")  // Clean up
            }
        }
        
        // Apply original stash if it exists
        origStashID := GetTimeTravelStashID()
        if origStashID != "" {
            buffer.Append("Restoring your original uncommitted work...", ui.TypeStatus)
            applyOrig := Execute("stash", "apply", origStashID)
            if !applyOrig.Success {
                buffer.Append("Note: Could not restore original work (may be redundant)", ui.TypeStatus)
            }
            Execute("stash", "drop", origStashID)  // Clean up
        }
        
        buffer.Append("Time travel merge complete!", ui.TypeStatus)
        return TimeTravelMergeMsg{
            Success: true,
            Error: "",
            ConflictDetected: false,
        }
    }
}
```

### 3.4 Handle Merge Result

`internal/app/githandlers.go`:
```go
func (a *Application) handleTimeTravelMerge(msg git.TimeTravelMergeMsg) (tea.Model, tea.Cmd) {
    buffer := ui.GetBuffer()
    
    if !msg.Success {
        buffer.Append("Merge aborted due to error", ui.TypeStderr)
        a.asyncOperationActive = false
        
        if msg.ConflictDetected {
            // Phase 4 will handle this
            return a.setupConflictResolver("time_travel_merge", []string{"ORIGINAL", "MERGED"})
        }
        
        // Stay in console for user to see error
        return a, nil
    }
    
    // Success: clear time travel state
    a.asyncOperationActive = false
    a.timeTravelInfo = nil
    git.ClearTimeTravelInfo()
    
    // Reload git state
    state, err := git.DetectState()
    if err != nil {
        buffer.Append("Failed to reload state", ui.TypeStderr)
        return a, nil
    }
    
    a.gitState = state
    buffer.Append("Merged successfully. Press ESC to return to menu.", ui.TypeStatus)
    
    return a, nil
}
```

**Acceptance Criteria (Phase 3):**
- [ ] "Merge & return" → confirmation dialog shown
- [ ] Cancel merge → stay in time travel mode
- [ ] Confirm merge → console shows operations, then menu appears
- [ ] After merge: tree is back on original branch
- [ ] If original had dirty work: stash applied, may show conflicts (Phase 4)
- [ ] `.git/TIT_TIME_TRAVEL` deleted after success

---

## Phase 4: Merge Conflicts & Resolution

**Scope:** Merge produces conflicts. Show ConflictResolver. User resolves. Continue merge. Apply stashes.

**Code Changes:**

### 4.1 Expand Merge Result Handler

Covered in Phase 3.4 `setupConflictResolver("time_travel_merge", ...)`

### 4.2 Handle Conflict Resolution Continue

`internal/app/conflicthandlers.go` - add case for time_travel_merge:
```go
case "time_travel_merge":
    // User resolved conflicts
    // Continue the sequence: apply time travel changes, apply original work
    return a.continueTimeTravelMerge()
```

### 4.3 Continue After Conflict

`internal/app/conflicthandlers.go`:
```go
func (a *Application) continueTimeTravelMerge() (tea.Model, tea.Cmd) {
    a.mode = ModeConsole
    a.outputBuffer.Clear()
    
    buffer := ui.GetBuffer()
    buffer.Append("Continuing merge sequence...", ui.TypeStatus)
    
    return a, git.ExecuteTimeTravelMergeContinue()
}
```

### 4.4 Continue Merge in Git

`internal/git/execute.go`:
```go
func ExecuteTimeTravelMergeContinue() func() tea.Msg {
    return func() tea.Msg {
        buffer := ui.GetBuffer()
        
        // Check for stashed time travel changes
        hasTimeTravalChanges := fileExists(".git/TIT_TIME_TRAVEL_CHANGES_STASH")
        if hasTimeTravalChanges {
            buffer.Append("Applying time travel changes...", ui.TypeStatus)
            applyResult := Execute("stash", "apply")
            if !applyResult.Success {
                // Stash apply failed - conflict
                return TimeTravelMergeMsg{
                    Success: false,
                    Error: "Conflicts applying time travel changes",
                    ConflictDetected: true,
                }
            }
            os.Remove(".git/TIT_TIME_TRAVEL_CHANGES_STASH")
        }
        
        // Check for original stash
        origStashID := GetTimeTravelStashID()
        if origStashID != "" {
            buffer.Append("Restoring original work...", ui.TypeStatus)
            applyOrig := Execute("stash", "apply", origStashID)
            if !applyOrig.Success {
                // Original stash apply failed - conflict
                return TimeTravelMergeMsg{
                    Success: false,
                    Error: "Conflicts applying original work",
                    ConflictDetected: true,
                }
            }
            Execute("stash", "drop", origStashID)
        }
        
        buffer.Append("All conflicts resolved!", ui.TypeStatus)
        ClearTimeTravelInfo()
        return TimeTravelMergeMsg{
            Success: true,
            Error: "",
            ConflictDetected: false,
        }
    }
}
```

**Acceptance Criteria (Phase 4):**
- [ ] Merge conflict → ConflictResolver shows LOCAL vs REMOTE
- [ ] Resolve conflict → continue button works
- [ ] Continue → applies time travel changes (may show 2nd conflict)
- [ ] Resolve 2nd conflict → apply original work (may show 3rd conflict)
- [ ] All conflicts resolvable in ConflictResolver
- [ ] Final ESC → menu returns, back on original branch, clean or with merged changes

---

## Phase 5: Browse While Time Traveling

**Scope:** While time traveling, user can press "Browse history" to jump to different commits.

**Code Changes:**

### 5.1 Implement Browse Handler

`internal/app/handlers.go`:
```go
func (a *Application) handleTimeTravelHistory(app *Application) (tea.Model, tea.Cmd) {
    if a.timeTravelInfo == nil {
        return app, nil
    }
    
    // Enter history mode while time traveling
    // History mode can show commits and allow jumping
    app.mode = ModeHistory
    app.historyState = &ui.HistoryState{
        SelectedIdx: 0,
        ScrollOffset: 0,
        DetailsScrollOffset: 0,
        FocusedPane: ui.PaneCommits,
    }
    
    // Load history commits
    commits, _ := git.FetchRecentCommits(30)
    app.historyState.Commits = commits
    
    return app, nil
}
```

### 5.2 Modify History ENTER Handler

`internal/app/handlers.go` - update `handleHistoryEnter()`:
```go
func (a *Application) handleHistoryEnter(app *Application) (tea.Model, tea.Cmd) {
    if a.historyState == nil || a.historyState.SelectedIdx < 0 {
        return app, nil
    }
    
    commit := a.historyState.Commits[a.historyState.SelectedIdx]
    
    // If already time traveling, just checkout different commit
    if a.timeTravelInfo != nil {
        return a.jumpToCommitWhileTimeTraveling(commit.Hash)
    }
    
    // Normal time travel entry (Phase 1-2)
    // Show confirmation dialog...
}
```

### 5.3 Jump To Different Commit

`internal/app/handlers.go`:
```go
func (a *Application) jumpToCommitWhileTimeTraveling(commitHash string) (tea.Model, tea.Cmd) {
    // Check for uncommitted changes while time traveling
    statusResult := git.Execute("status", "--short")
    hasChanges := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""
    
    if hasChanges {
        // Must stash before jumping
        a.mode = ModeConfirmation
        a.confirmContext = map[string]string{
            "new_commit": commitHash,
        }
        // Show dialog: "Save changes before jumping?"
        return a, nil
    }
    
    // Clean: jump directly
    a.mode = ModeConsole
    a.outputBuffer.Clear()
    buffer := ui.GetBuffer()
    buffer.Append(fmt.Sprintf("Jumping to %s...", commitHash[:7]), ui.TypeStatus)
    
    return a, git.ExecuteCheckoutCommit(commitHash)
}

func (a *Application) handleCheckoutCommitResult(msg git.CheckoutMsg) (tea.Model, tea.Cmd) {
    if msg.Success {
        a.timeTravelInfo.CurrentCommit = msg.CommitInfo
        buffer := ui.GetBuffer()
        buffer.Append("Jumped successfully. Press ESC to continue.", ui.TypeStatus)
    }
    return a, nil
}
```

**Acceptance Criteria (Phase 5):**
- [ ] Time traveling → can select "Browse history"
- [ ] Browse → shows commit list
- [ ] Select different commit + ENTER → jump (if clean)
- [ ] If changes exist → confirm stash first
- [ ] After jump → back in time travel mode at new commit
- [ ] ESC from history → back to time travel menu

---

## Phase 6: Return Without Merge (Discard Changes)

**Scope:** Time traveling, user selects "Return". Discard time travel changes, restore to original branch.

**Code Changes:**

### 6.1 Implement Return Handler

`internal/app/handlers.go`:
```go
func (a *Application) handleTimeTravelReturn(app *Application) (tea.Model, tea.Cmd) {
    if a.timeTravelInfo == nil {
        return app, nil
    }
    
    // Show confirmation
    app.mode = ModeConfirmation
    app.confirmType = "time_travel_return_confirm"
    
    config := ui.ConfirmationConfig{
        Title: fmt.Sprintf("Return to %s", a.timeTravelInfo.OriginalBranch),
        Explanation: "Your changes during time travel will be DISCARDED.\n" +
            "You will return to your original branch with original state.",
        YesLabel: "Discard & return",
        NoLabel: "Cancel",
        ActionID: "time_travel_return_confirm",
    }
    app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)
    
    return app, nil
}
```

### 6.2 Add Return Handler

`internal/app/confirmationhandlers.go`:
```go
var confirmationActions = map[string]ConfirmationAction{
    // ... existing ...
    "time_travel_return_confirm": (*Application).executeTimeTravelReturn,
}

var confirmationRejectActions = map[string]ConfirmationAction{
    // ... existing ...
    "time_travel_return_confirm": (*Application).executeRejectTimeTravelReturn,
}

func (a *Application) executeTimeTravelReturn() (tea.Model, tea.Cmd) {
    if a.timeTravelInfo == nil {
        return a.returnToMenu()
    }
    
    originalBranch := a.timeTravelInfo.OriginalBranch
    a.confirmationDialog = nil
    
    // Discard all changes (hard reset to HEAD of time travel commit)
    git.Execute("checkout", ".")  // Discard working tree changes
    git.Execute("clean", "-fd")   // Remove untracked files
    
    // Transition to console
    a.asyncOperationActive = true
    a.mode = ModeConsole
    a.outputBuffer.Clear()
    a.footerHint = "Returning..."
    a.previousMode = ModeMenu
    
    return a, git.ExecuteTimeTravelReturn(originalBranch)
}

func (a *Application) executeRejectTimeTravelReturn() (tea.Model, tea.Cmd) {
    a.confirmationDialog = nil
    return a, nil
}
```

### 6.3 Add Return Git Operation

`internal/git/execute.go`:
```go
func ExecuteTimeTravelReturn(originalBranch string) func() tea.Msg {
    return func() tea.Msg {
        buffer := ui.GetBuffer()
        buffer.Append(fmt.Sprintf("Returning to %s...", originalBranch), ui.TypeStatus)
        
        // Checkout original branch
        checkoutResult := Execute("checkout", originalBranch)
        if !checkoutResult.Success {
            buffer.Append("Failed to checkout original branch", ui.TypeStderr)
            return TimeTravelReturnMsg{
                Success: false,
                Error: "Checkout failed",
            }
        }
        
        // Restore original uncommitted work
        origStashID := GetTimeTravelStashID()
        if origStashID != "" {
            buffer.Append("Restoring original uncommitted work...", ui.TypeStatus)
            applyResult := Execute("stash", "apply", origStashID)
            if !applyResult.Success {
                buffer.Append("Note: Could not restore original work", ui.TypeStatus)
            }
            Execute("stash", "drop", origStashID)
        }
        
        ClearTimeTravelInfo()
        buffer.Append("Returned to original branch. Press ESC to continue.", ui.TypeStatus)
        
        return TimeTravelReturnMsg{
            Success: true,
            Error: "",
        }
    }
}
```

### 6.4 Handle Return Result

`internal/app/githandlers.go`:
```go
func (a *Application) handleTimeTravelReturn(msg git.TimeTravelReturnMsg) (tea.Model, tea.Cmd) {
    a.asyncOperationActive = false
    a.timeTravelInfo = nil
    
    if !msg.Success {
        buffer := ui.GetBuffer()
        buffer.Append(msg.Error, ui.TypeStderr)
        return a, nil
    }
    
    // Reload state
    state, _ := git.DetectState()
    a.gitState = state
    
    buffer := ui.GetBuffer()
    buffer.Append("Back on original branch. You are no longer time traveling.", ui.TypeStatus)
    
    return a, nil
}
```

**Acceptance Criteria (Phase 6):**
- [ ] "Return" → confirmation dialog
- [ ] Cancel return → stay in time travel
- [ ] Confirm return → checkout original branch
- [ ] If original had stashed work → restored
- [ ] After return: `.git/TIT_TIME_TRAVEL` deleted
- [ ] ESC from console → menu on original branch

---

## Type Definitions

Add to `internal/git/types.go`:
```go
type TimeTravelMergeMsg struct {
    Success bool
    Error string
    ConflictDetected bool
}

type TimeTravelReturnMsg struct {
    Success bool
    Error string
}

type CheckoutMsg struct {
    Success bool
    CommitInfo CommitInfo
    Error string
}
```

---

## Files Modified Summary

| File | Changes | Lines |
|------|---------|-------|
| `internal/git/types.go` | Add TimeTravelInfo, messages (Phase 0 adds OriginalStashID) | +35 |
| `internal/git/execute.go` | Merge, return, jump operations | +150 |
| `internal/app/app.go` | Pre-startup restoration, timeTravelInfo field, register handlers | +80 |
| `internal/app/handlers.go` | Time travel handlers (merge, return, browse) | +120 |
| `internal/app/confirmationhandlers.go` | Dirty protocol, merge confirm, return confirm | +180 |
| `internal/app/menu.go` | menuTimeTraveling(), customize labels | +60 |
| `internal/app/githandlers.go` | handleTimeTravelCheckout extended, merge/return handlers | +100 |
| `internal/ui/layout.go` | Time travel header display | +20 |
| **Total** | | **~745 lines** |

---

## Testing Mindset

**Each phase must test:**
1. Happy path (no issues)
2. ESC cancellation (at every prompt)
3. Git failures (checked with error messages)
4. State preservation (nothing lost on cancel)

**See TIME-TRAVEL-TESTING-CHECKLIST.md for detailed scenarios.**

---

**Status:** Ready for incremental implementation and testing.
