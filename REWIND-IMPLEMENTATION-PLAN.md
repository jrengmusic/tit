# REWIND Feature Implementation Plan

**Feature:** `git reset --hard` at selected commit in Commit History
**Trigger:** Ctrl+R (vs ENTER for time travel)  
**Requirements:**
- Always available in commit history, regardless of Operation state
- Discards all commits after selected commit
- Discards all uncommitted changes
- Requires destructive confirmation

**Confirmation Message:**
```
⚠️ Destructive Operation

This will discard all commits after [HASH]. 
Any uncommitted changes will be lost.

Are you sure you want to continue?

[Rewind] [Cancel]
```

---

## Implementation Plan (14 Steps, Incrementally Testable)

### Phase 1: Type System & Constants (Step 1-2)

#### Step 1: Add REWIND to Operation types
**File:** `internal/git/types.go`

- Add `Rewinding` to Operation enum (represents active rewind operation, between reset and state refresh)
- Add documentation: "Rewinding indicates a reset --hard is in progress"

**Verify:** Code compiles, no errors

---

#### Step 2: Add REWIND command constant
**File:** `internal/app/messages.go`

- Add `RewindMsg` struct:
  ```go
  type RewindMsg struct {
      Commit    string // hash
      Success   bool
      Output    string
      Error     string
  }
  ```
- This message will be sent when `git reset --hard` completes

**Verify:** Code compiles

---

### Phase 2: UI Layer (Step 3-5)

#### Step 3: Detect Ctrl modifier in commit history keyboard handler
**File:** `internal/ui/layout.go` or relevant history rendering file

- Modify commit history footer rendering to:
  - **Default (no Ctrl):** Show "Press Enter to explore this commit in time travel mode"
  - **When Ctrl pressed:** Show "Press Ctrl+R to REWIND | Press Enter to time travel"
  
- **Challenge:** Bubble Tea doesn't report Ctrl state in `View()` (only in event handlers)
- **Solution:** Track `isCtrlPressed` boolean in Application struct, update on keyboard events
  - Set `true` in Ctrl+X handlers
  - Set `false` when non-Ctrl keys pressed
  - Reset `false` on any non-modifier key release

**Verify:** Build succeeds. Manual test: navigate to history, hold Ctrl, verify footer changes.

---

#### Step 4: Add keyboard handler for Ctrl+R in history browser
**File:** `internal/app/keyboard.go`

- Add handler entry for `ModeHistoryBrowser` + `ctrl+r`:
  ```go
  handlers[ModeHistoryBrowser]["ctrl+r"] = handleHistoryRewind
  ```

- Create `handleHistoryRewind()` function:
  - Get selected commit hash from history browser state
  - Call `a.showRewindConfirmation(commitHash)`
  - Return confirmation mode

**Verify:** Code compiles. Manual test: In history, press Ctrl+R, confirmation appears.

---

#### Step 5: Create rewind confirmation dialog
**File:** `internal/app/confirmation_handlers.go` (new section)

- Create `showRewindConfirmation(commitHash string)` function
- Render confirmation dialog with:
  - Title: "⚠️ Destructive Operation"
  - Message: "This will discard all commits after [HASH].\nAny uncommitted changes will be lost.\n\nAre you sure you want to continue?"
  - Two buttons: "Rewind" (dangerous/red) + "Cancel"
  - Store `pendingRewindCommit` in Application struct

- When user selects "Rewind":
  - Set `a.asyncOperationActive = true`
  - Set `a.mode = ModeConsole`
  - Return `executeRewindOperation()`

- When user selects "Cancel":
  - Discard `pendingRewindCommit`
  - Return to `ModeHistoryBrowser`

**Verify:** Build succeeds. Manual test: Confirm dialog appears and closes correctly.

---

### Phase 3: Git Operations (Step 6-7)

#### Step 6: Implement git reset --hard execution
**File:** `internal/git/execute.go` or new `internal/git/rewind.go`

- Create `ResetHardAtCommit(commitHash string) error` function:
  ```go
  func ResetHardAtCommit(commitHash string) (string, error) {
      cmd := exec.Command("git", "reset", "--hard", commitHash)
      // ...
  }
  ```

- Execute `git reset --hard <commit>`
- Capture stdout/stderr
- Stream output to OutputBuffer (via worker goroutine in app layer)
- Return success/error

**Verify:** Function signature correct, returns error on failure.

---

#### Step 7: Add async operation wrapper for rewind
**File:** `internal/app/handlers.go`

- Create `executeRewindOperation()` tea.Cmd:
  ```go
  func (a *Application) executeRewindOperation() tea.Cmd {
      commit := a.pendingRewindCommit
      return func() tea.Msg {
          output := ui.GetBuffer()
          output.Clear()
          
          _, err := git.ResetHardAtCommit(commit)
          
          return RewindMsg{
              Commit: commit,
              Success: err == nil,
              Error: err.Error(),
          }
      }
  }
  ```

- Implement `cmdRefreshConsole()` refresh ticker (reuse from clone feature)
- Batch both: `tea.Batch(executeRewindOperation(), cmdRefreshConsole())`

**Verify:** Build succeeds. Rewind operation doesn't hang UI.

---

### Phase 4: State Management (Step 8-9)

#### Step 8: Add RewindMsg handler in app Update()
**File:** `internal/app/app.go`

- Add case handler for `RewindMsg`:
  ```go
  case RewindMsg:
      a.asyncOperationActive = false
      if msg.Success {
          // Refresh git state
          newState, _ := git.DetectState()
          a.gitState = newState
          a.mode = ModeMenu
          a.menuItems = a.GenerateMenu()
      } else {
          // Show error in console
          output := ui.GetBuffer()
          output.Append("❌ Rewind failed: " + msg.Error)
          // Stay in ModeConsole, wait for ESC
      }
  ```

- This handler:
  - Sets `asyncOperationActive = false` (allows ESC to work)
  - Detects new git state (working tree reset, commits gone)
  - Regenerates menu
  - Returns to main menu on success

**Verify:** Build succeeds. Manual test: Rewind completes, menu regenerates.

---

#### Step 9: Handle ESC during rewind operation
**File:** `internal/app/keyboard.go` (ESC handler)

- ESC during `asyncOperationActive` already sets `asyncOperationAborted = true`
- Worker goroutine should check this flag and gracefully cancel (though reset --hard can't really be interrupted mid-operation)
- On completion, ESC handler returns to previous mode (restore from `previousMode`)

**Note:** `git reset --hard` cannot be interrupted mid-operation, but user can exit console before operation completes (will complete in background).

**Verify:** ESC works during rewind, returns to history browser.

---

### Phase 5: Edge Cases (Step 10-11)

#### Step 10: Handle rewind when Operation ≠ Normal
**File:** `internal/app/confirmation_handlers.go`

- When `showRewindConfirmation()` is called:
  - Add warning if `a.gitState.Operation != git.Normal`:
    ```
    ⚠️ You are currently [TimeTraveling/Merging/etc]
    
    This will discard all commits after [HASH] and exit your current operation.
    Any uncommitted changes will be lost.
    ```
  - Still allow rewind (always available requirement)

**Verify:** Build succeeds. Test rewind from TimeTraveling state.

---

#### Step 11: Handle rewind on detached HEAD (non-time-travel)
**File:** `internal/git/state.go`

- When rewind completes and `git rev-parse --abbrev-ref HEAD` returns `HEAD`:
  - This means we're on detached HEAD (not time travel, but after reset to non-existent branch)
  - This shouldn't happen if reset target is valid commit
  - If it does: Show warning "Rewound to detached HEAD. Please switch to a branch."

**Verify:** Code handles edge case without panicking.

---

### Phase 6: UI Polish (Step 12)

#### Step 12: Add visual feedback during rewind
**File:** `internal/ui/console.go`

- When rewind starts, show header in console:
  ```
  🔄 Rewinding to [HASH]...
  Discarding all commits after [HASH]
  
  ```
- After success:
  ```
  ✅ Rewound successfully to [HASH]
  Press ESC to return to menu
  ```

**Verify:** Console output is clear and helpful.

---

### Phase 7: Documentation (Step 13-14)

#### Step 13: Update SPEC.md
**File:** `SPEC.md`

Add new section in "State → Menu Mapping":
```markdown
### Commit History Browser — Rewind Option

**When in commit history browser (ModeHistoryBrowser):**

- **ENTER:** Enter time travel mode at selected commit (read-only exploration)
- **Ctrl+R:** REWIND to selected commit (destructive, reset --hard)

**REWIND (Ctrl+R) behavior:**
- Available on ANY commit, regardless of current Operation state
- Shows confirmation: "This will discard all commits after [HASH]. Any uncommitted changes will be lost."
- Executes `git reset --hard <commit>`
- Discards ALL commits after selected commit
- Discards ALL uncommitted changes
- Returns to main menu on success

**Example:**
```
Commits on main:
  abc1234 Feature X (current HEAD)
  def5678 Fix bug Y
  ghi9012 Initial commit ← User selects, presses Ctrl+R
  
Result: Branch resets to ghi9012, abc1234 and def5678 discarded
```

**Why Ctrl+R:**
- Distinguishes destructive reset from read-only time travel
- Ctrl modifier indicates dangerous operation
- ENTER is already bound to time travel (safe, reversible)
```

**Verify:** SPEC.md reads clearly, addition makes sense in context.

---

#### Step 14: Update ARCHITECTURE.md
**File:** `ARCHITECTURE.md`

Add section: "REWIND Operation Flow"
```markdown
### REWIND Operation (git reset --hard)

**Entry Point:** Commit history browser, Ctrl+R on selected commit

**State Transitions:**
```
ModeHistoryBrowser
  ↓ (Ctrl+R)
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
- Ctrl+R: Destructive, permanent, discards commits (reset --hard)

**Always Available:** REWIND can be initiated from any Operation state (including TimeTraveling). Confirmation dialog warns if not Normal state.

**Implementation:**
- `git.ResetHardAtCommit(commitHash)` executes reset
- OutputBuffer streams git output in real-time
- RewindMsg handler refreshes git state and regenerates menu
```

**Verify:** Addition integrates with existing architecture docs.

---

## Testing Scenarios (Verification Checklist)

After each phase, run manual tests:

### Phase 1-2 Tests
- [ ] Code compiles without errors
- [ ] RewindMsg struct matches git operation pattern

### Phase 3-5 Tests
- [ ] Navigate to commit history
- [ ] Hold Ctrl, footer changes to show "Ctrl+R REWIND"
- [ ] Release Ctrl, footer reverts to "Enter time travel"
- [ ] Press Ctrl+R on a commit
- [ ] Confirmation dialog appears with correct message
- [ ] Select "Rewind", operation executes
- [ ] Console shows git output in real-time
- [ ] After completion, menu regenerates
- [ ] Branch is now at selected commit (previous commits gone)

### Phase 6-7 Tests
- [ ] Test rewind from Normal state ✅
- [ ] Test rewind from TimeTraveling state ✅
- [ ] Test ESC during rewind operation ✅
- [ ] Test rewind with dirty working tree ✅ (discards changes as specified)
- [ ] Test rewind on non-existent commit hash ❌ (should show error)
- [ ] SPEC.md reads correctly ✅
- [ ] ARCHITECTURE.md integrates correctly ✅

### Full Integration Tests
- [ ] Build succeeds: `./build.sh`
- [ ] Binary runs: `./tit_x64`
- [ ] Rewind → return to menu → normal operations work ✅
- [ ] Rewind → ESC → returns to history browser ✅
- [ ] Multiple consecutive rewinds work ✅

---

## Implementation Order (Recommended)

**Batch 1 (Types & Messages):** Steps 1-2
- Quick compile check
- Foundation for rest of feature

**Batch 2 (UI & Handlers):** Steps 3-5
- Keyboard integration
- Visual feedback (footer change, confirmation dialog)
- Can test without git operations

**Batch 3 (Git Operations):** Steps 6-9
- Execute rewind
- State refresh
- Async handling
- Full feature functional

**Batch 4 (Polish & Edge Cases):** Steps 10-14
- Robustness
- Documentation
- Full test coverage

**Suggested user testing points:**
- After Batch 2: UI responds correctly
- After Batch 3: Rewind actually resets commits
- After Batch 4: All edge cases handled, docs complete

---

## Risk Mitigation

**Risk:** User accidentally rewinds entire branch history  
**Mitigation:** Confirmation dialog clearly states "all commits after [HASH]"

**Risk:** Rewind fails, leaves repo in inconsistent state  
**Mitigation:** `git reset --hard` is atomic; if it fails, nothing changed

**Risk:** User presses Ctrl+R thinking it's time travel  
**Mitigation:** Confirmation dialog explicitly says "destructive operation"

**Risk:** Rewind from non-Normal state confuses user  
**Mitigation:** Confirmation dialog warns if not in Normal state

---

**End of REWIND Implementation Plan**
