# Phase 7: Time Travel Integration - KICKOFF

**Status:** 🟢 READY  
**Estimated:** 1.5 days (~800 lines)  
**Complexity:** HIGH (git operations, state transitions, async)  
**Depends On:** Phase 4 (History mode), Phase 6 (Handlers)

---

## What Phase 7 Does

Adds **Time Travel** feature: explore commit history in read-only mode, optionally merge changes back.

Time travel allows users to:
- Checkout any historical commit
- Browse/build/test old code
- Make local changes (not committed)
- Return to original branch with/without changes

---

## Core Workflow

```
User in History Mode
    ↓
Press ENTER on commit
    ↓
Confirmation dialog
    ↓ User confirms
Git checkout <hash> (stash if dirty)
    ↓
Operation = TimeTraveling (detached HEAD)
Menu changes to time travel options only
    ↓
User can:
  - Browse history (read-only)
  - Make local changes
  - View diffs
    ↓
User selects:
  - "Merge back" → stash changes, checkout orig branch, merge
  - "Return without merge" → checkout orig branch, discard
```

---

## State Model

### Operation = TimeTraveling (NEW)

**Triggers:**
- User confirms time travel from History mode
- Successful `git checkout <hash>`
- `.git/TIT_TIME_TRAVEL` file created with original branch

**Menu:**
- Show ONLY time travel options
- Hide normal git operations
- Keep History browser accessible

**Exit:**
- User chooses "Return to branch" or "Merge back"
- Cleanup: remove `.git/TIT_TIME_TRAVEL`, checkout original branch

---

## Files to Modify

### 1. `internal/app/handlers.go`

#### Modify `handleHistoryEnter()`

**Current:** Placeholder for Phase 7

**New:** Show confirmation dialog
```go
func (a *Application) handleHistoryEnter(app *Application) (tea.Model, tea.Cmd) {
    if app.historyState == nil || app.historyState.SelectedCommitIdx < 0 {
        return app, nil
    }
    
    // Show confirmation dialog
    app.mode = ModeConfirmation
    app.confirmType = "time_travel"
    commit := app.historyState.Commits[app.historyState.SelectedCommitIdx]
    app.confirmContext = map[string]string{
        "commit_hash": commit.Hash,
        "commit_subject": commit.Subject,
    }
    
    // Create dialog (from SSOT or inline)
    app.confirmationDialog = ui.NewConfirmationDialog(config, ...)
    return app, nil
}
```

#### Add Time Travel Handlers

```go
func (a *Application) handleTimeTravelReturn(app *Application) (tea.Model, tea.Cmd)
func (a *Application) handleTimeTravelMerge(app *Application) (tea.Model, tea.Cmd)
func (a *Application) handleTimeTravelViewHistory(app *Application) (tea.Model, tea.Cmd)
```

---

### 2. `internal/app/menu.go`

Add menu generator:

```go
func (a *Application) menuTimeTraveling() []MenuItem {
    // Return time travel options
    // Show:
    // - View History (browse commits while traveling)
    // - Merge back to [branch name] (merge changes)
    // - Return without merge (discard changes)
}
```

**Menu items:**
- 🕒 Browse history (read-only, same as History mode)
- 📦 Merge back to [branch] (stash + checkout + merge)
- ⬅️ Return to [branch] (checkout, discard)

---

### 3. `internal/app/dispatchers.go`

Add confirmation handler:

```go
func (a *Application) handleTimeTravelConfirmation(app *Application) tea.Cmd {
    commitHash := app.confirmContext["commit_hash"]
    
    // Dirty operation protocol if WorkingTree = Dirty
    if app.gitState.WorkingTree == git.Dirty {
        // Show dirty operation confirmation
        return a.cmdTimeTravelDirty(commitHash)
    }
    
    // Clean: just checkout
    return a.executeTimeTravelCheckout(commitHash)
}
```

---

### 4. `internal/app/githandlers.go`

Add operation result handlers:

```go
case git.TimeTravelCheckoutMsg:
    // Checkout succeeded
    a.gitState = git.DetectState()  // Will show Operation = TimeTraveling
    a.operation = TimeTraveling
    a.mode = ModeMenu
    a.menuItems = a.GenerateMenu()  // Will show time travel options only
```

---

### 5. `internal/git/execute.go`

Add git operations:

```go
// executeTimeTravelCheckout performs: git checkout <hash>
// On success, creates .git/TIT_TIME_TRAVEL with original branch
func executeTimeTravelCheckout(hash, originalBranch string) error

// executeTimeTravelMerge performs:
// 1. git checkout <originalBranch>
// 2. git merge <hash>
func executeTimeTravelMerge(originalBranch, timeTravelHash string) error

// executeTimeTravelReturn performs:
// 1. git checkout <originalBranch>
// 2. Discard local changes (reset --hard)
func executeTimeTravelReturn(originalBranch string) error

// isTimeTraveling checks for .git/TIT_TIME_TRAVEL file
func isTimeTraveling() (bool, string, string, error)
```

---

### 6. `internal/app/app.go`

Add to `Update()` message handling:

```go
case git.TimeTravelCheckoutMsg:
    a.gitState.Operation = git.TimeTraveling
    // Save original branch
    a.timeTravelOriginalBranch = msg.OriginalBranch
    
case git.TimeTravelMergeMsg:
    a.gitState = git.DetectState()
    a.mode = ModeMenu
    a.footerHint = "Changes merged back to original branch"
    
case git.TimeTravelReturnMsg:
    a.gitState = git.DetectState()
    a.mode = ModeMenu
    a.footerHint = "Returned to original branch"
```

---

## Implementation Steps

### Step 1: Add Type Definitions
- Add `TimeTraveling` to `Operation` enum in `git/types.go`
- Add time travel message types in `app/messages.go`

### Step 2: Modify Git State Detection
- Update `DetectState()` to check for `.git/TIT_TIME_TRAVEL` file
- Return `Operation = TimeTraveling` if detected

### Step 3: Add Menu Generator
- Implement `menuTimeTraveling()` in `app/menu.go`
- Update `GenerateMenu()` to call it when `Operation = TimeTraveling`

### Step 4: Implement Handlers
- Modify `handleHistoryEnter()` to show confirmation
- Add time travel handlers
- Register in key handler registry for time travel mode

### Step 5: Add Git Operations
- Implement time travel git commands in `git/execute.go`
- Handle both clean and dirty working trees
- Proper error handling

### Step 6: Integrate with Messages
- Add message types for time travel operations
- Handle results in `Update()`

### Step 7: Confirmation Dialog
- Create time travel confirmation (reuse existing ConfirmationDialog)
- Show original branch + commit hash

---

## Key Design Decisions

### Time Travel Mode (NEW Operation State)

**When Active:**
- Git state: `Operation = TimeTraveling`
- Current branch: Detached HEAD
- Saved info: `.git/TIT_TIME_TRAVEL` file with original branch

**Menu Changes:**
- Only show time travel options
- Hide commit/push/pull options
- Keep History browser accessible

**Exit Paths:**
1. **Merge back** → Preserve local changes
2. **Return without merge** → Discard local changes

### Dirty Working Tree Protocol

**Scenario:** User has uncommitted changes, wants to time travel

**Current spec:** Apply Dirty Operation Protocol (Phase 2)
1. Stash changes
2. Checkout commit
3. User explores (optional: merge time travel changes back)
4. Return to original branch
5. Restore stashed changes

**Implementation:**
```go
if app.gitState.WorkingTree == Dirty {
    // Show dialog: save/discard/cancel
    // If save: git stash push -u -m "TIT TIME_TRAVEL"
    // Then: git checkout <hash>
    // Then: save both original branch + stash id to .git/TIT_TIME_TRAVEL
}
```

### History Access While Time Traveling

**Spec § 9:** Users can browse history while in time travel mode

**Implementation:**
- Don't add `ModeHistory` to handlers during time travel
- Keep it in menu, selectable from menu
- History mode works normally (read-only, already safe)
- Tab key can switch between time travel and history

---

## Messages & Operations (Async)

### Time Travel Checkout (Async)

```
User confirms time travel
    ↓
executeTimeTravelCheckout() spawned in goroutine
    ↓ [git checkout <hash>]
    ↓ [create .git/TIT_TIME_TRAVEL]
    ↓
TimeTravelCheckoutMsg returned
    ↓
Update() processes message
    ↓
Mode = Menu, Operation = TimeTraveling
```

### Time Travel Merge (Async)

```
User selects "Merge back"
    ↓
executeTimeTravelMerge() spawned
    ↓ [git checkout <original>]
    ↓ [git merge <time-travel-hash>]
    ↓
TimeTravelMergeMsg returned
    ↓ (If conflict: Operation = Conflicted)
    ↓ (If success: Operation = Normal)
```

---

## Error Handling

**If Time Travel Checkout Fails:**
- Show error in console mode
- Return to History mode on ESC
- Discard stashed changes (auto-cleanup)

**If Time Travel Merge Conflicts:**
- Show Conflict Resolver (same as normal merge)
- Continue or abort from conflict menu
- After resolution: return to Normal operation

**If User Manually Exits Time Travel (e.g., closes app):**
- Next app startup: detect `.git/TIT_TIME_TRAVEL`
- Show menu: "Resume time travel" or "Discard"

---

## Testing Checklist

- [ ] Enter time travel from History mode (ENTER on commit)
- [ ] Confirmation dialog shows commit hash + subject
- [ ] User can confirm or cancel
- [ ] On confirm: git checkout succeeds
- [ ] Mode changes to ModeMenu, Operation = TimeTraveling
- [ ] Menu shows only time travel options
- [ ] History still accessible from menu
- [ ] Can make local changes (not committed)
- [ ] "Merge back" option works
  - [ ] Merges time travel commit
  - [ ] Shows merge conflicts if any (Conflict Resolver)
  - [ ] Returns to Normal operation after
- [ ] "Return without merge" works
  - [ ] Returns to original branch
  - [ ] Discards local changes
  - [ ] Returns to Normal operation
- [ ] No regressions to other modes

---

## What Phase 7 Does NOT Do

- ❌ No copy/visual mode (Phase 8)
- ❌ No cache invalidation (Phase 8)
- ❌ No merge strategy options (beyond default)
- ❌ No stash management UI (beyond auto-stash)

---

## Acceptance Criteria

- [ ] Time travel entry from History mode works
- [ ] Confirmation dialog shows commit details
- [ ] Checkout executes correctly (clean working tree)
- [ ] Checkout with dirty tree uses dirty operation protocol
- [ ] Mode transitions to ModeMenu with Operation = TimeTraveling
- [ ] Menu shows time travel options only
- [ ] Can browse history while time traveling
- [ ] Can make local changes
- [ ] Merge back succeeds (merges changes)
- [ ] Return without merge succeeds (discards)
- [ ] Conflict handling works (Conflict Resolver)
- [ ] ESC/return to menu works
- [ ] No regressions
- [ ] Compiles without errors
- [ ] Compiles without warnings

---

## Reference

- **Spec:** SPEC.md § 9 (Time Travel § 9.1-9.4)
- **Implementation Plan:** HISTORY-IMPLEMENTATION-PLAN.md § Phase 7
- **Dirty Operation Protocol:** SPEC.md § 6
- **Current State:** Phase 6 COMPLETE (history navigation working)
- **Architecture:** ARCHITECTURE.md (async patterns, state machine)

---

## Git Operations Summary

| Operation | Command | Captures |
|-----------|---------|----------|
| Enter Time Travel | `git checkout <hash>` | Original branch + hash |
| While Time Traveling | Local edits OK | Not committed |
| Merge Back | `git checkout <orig> && git merge <hash>` | Returns to orig branch |
| Return Without Merge | `git checkout <orig> && git reset --hard` | Discards changes |

---

## Complexity Notes

**Why HIGH Complexity:**

1. **State Transitions:** 3 git operations, multiple confirmations
2. **Dirty Handling:** Stash/restore if working tree dirty
3. **Error Paths:** Merge conflicts, checkout failures
4. **Menu Changes:** Different menu based on Operation state
5. **Cleanup:** Must handle interruptions, cleanup files
6. **Thread Safety:** Async checkout + state detection

**Estimated Lines:** ~800
- Git execute functions: 200
- Handlers: 150
- Menu: 50
- Messages: 50
- Integration (Update handler): 100
- Error handling + cleanup: 150
- Testing: 100

---

## Quick Checklist for Agent

Before starting:
- [ ] Read SPEC.md § 9 (Time Travel requirements)
- [ ] Understand Dirty Operation Protocol (SPEC.md § 6)
- [ ] Review Phase 6 handler pattern
- [ ] Understand `.git/TIT_TIME_TRAVEL` file format
- [ ] Know 3 exit paths: merge, return, cancel
- [ ] Plan error handling for each git operation

---

**Proceed when ready.** Phase 7 is the most complex but adds the marquee "Time Travel" feature.

Good luck! 🚀
