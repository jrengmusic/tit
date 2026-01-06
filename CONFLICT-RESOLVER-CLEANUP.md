# Conflict Resolver Cleanup Tasks

**Date:** 2026-01-07
**Status:** Ready for delegation
**Context:** Session 43 successfully wired pull merge conflict resolution. Flow works end-to-end, but code quality issues remain.

---

## 🎯 Overview

Conflict resolver is **functionally working** for Scenario 1 (clean pull with conflicts). Testing shows:

✅ Conflicts detected correctly
✅ Resolver UI appears
✅ File list shows actual filenames
✅ Column labels show correctly (BASE, LOCAL, REMOTE)
✅ User can mark files with SPACE
✅ ENTER finalizes merge successfully
✅ Final state correct (Clean + Ahead)

However, code has **technical debt** that will cause issues when implementing:
- Scenario 2-5 (dirty pull variants)
- Branch operations with conflicts
- Cherry-pick conflicts
- Any future conflict-resolving operations

---

## 🚨 P0: Critical Issues (Breaks Abort Flow)

### Issue 1: Missing Operation Step Constants

**File:** `internal/app/operationsteps.go`

**Problem:** No constants for finalize/abort operations. Current code reuses `OpPull` for everything:

```go
// operations.go:798, 809, 817, 835, 843
return GitOperationMsg{
    Step: OpPull,  // ❌ Can't distinguish: pull success vs finalize vs abort
    Success: true,
}
```

**Impact:** `handleGitOperation()` can't route correctly. All three operations hit `case OpPull:` which expects clean pull completion.

**Fix Required:**

1. Add to `operationsteps.go`:
```go
// Pull merge conflict resolution
OpFinalizePullMerge = "finalize_pull_merge"
OpAbortMerge        = "abort_merge"
```

2. Update `operations.go` lines 798, 809, 817:
```go
return GitOperationMsg{
    Step: OpFinalizePullMerge,  // ✅ Use specific constant
    Success: true,
}
```

3. Update `operations.go` lines 835, 843:
```go
return GitOperationMsg{
    Step: OpAbortMerge,  // ✅ Use specific constant
    Success: true,
}
```

4. Add routing in `githandlers.go` after `case OpPull:`:
```go
case OpFinalizePullMerge:
    // Reload state after merge finalization
    state, err := git.DetectState()
    if err != nil {
        buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
        a.asyncOperationActive = false
        return a, nil
    }
    a.gitState = state
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
    a.footerHint = GetFooterMessageText(MessageOperationComplete)
    a.asyncOperationActive = false
    a.conflictResolveState = nil
    return a, nil

case OpAbortMerge:
    // Reload state after merge abort
    state, err := git.DetectState()
    if err != nil {
        buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
        a.asyncOperationActive = false
        return a, nil
    }
    a.gitState = state
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
    a.footerHint = GetFooterMessageText(MessageOperationComplete)
    a.asyncOperationActive = false
    a.conflictResolveState = nil
    return a, nil
```

**Test:** Run Scenario 1, test both success (ENTER) and abort (ESC) flows. Verify correct state after each.

---

## 📝 P1: SSOT Violations (Hardcoded Strings)

### Issue 2: Hardcoded Error/Success Messages

**Files:** `internal/app/operations.go`, `internal/app/conflicthandlers.go`

**Problem:** User-facing text hardcoded instead of using messages.go SSOT.

**Violations in `operations.go`:**

Line 796:
```go
buffer.Append("Failed to stage resolved files", ui.TypeStderr)
```

Line 807:
```go
buffer.Append("Failed to commit merge", ui.TypeStderr)
```

Line 815:
```go
buffer.Append("Merge completed successfully", ui.TypeInfo)
```

Line 833:
```go
buffer.Append("Failed to abort merge", ui.TypeStderr)
```

Line 841:
```go
buffer.Append("Merge aborted - state restored", ui.TypeInfo)
```

**Violations in `conflicthandlers.go`:**

Line 111:
```go
a.footerHint = "Already marked in this column"
```

Line 117:
```go
a.footerHint = fmt.Sprintf("Marked: %s → column %d", file.Path, focusedPane)
```

**Fix Required:**

1. Add to `internal/app/messages.go` ErrorMessages map:
```go
"failed_stage_resolved":  "Failed to stage resolved files",
"failed_commit_merge":    "Failed to commit merge",
"failed_abort_merge":     "Failed to abort merge",
```

2. Add to OutputMessages map:
```go
"merge_finalized":        "Merge completed successfully",
"merge_aborted":          "Merge aborted - state restored",
```

3. Add to FooterHints map:
```go
"already_marked_column":  "Already marked in this column",
"marked_file_column":     "Marked: %s → %s",  // file.Path, columnLabel
```

4. Update all hardcoded strings to use SSOT:
```go
// operations.go:796
buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)

// operations.go:815
buffer.Append(OutputMessages["merge_finalized"], ui.TypeInfo)

// conflicthandlers.go:111
a.footerHint = FooterHints["already_marked_column"]

// conflicthandlers.go:117
columnLabel := a.conflictResolveState.ColumnLabels[focusedPane]
a.footerHint = fmt.Sprintf(FooterHints["marked_file_column"], file.Path, columnLabel)
```

**Test:** Build and verify all messages display correctly.

---

## 🔧 P2: Code Quality Issues

### Issue 3: Repetitive setupConflictResolver*() Functions

**Files:** `internal/app/githandlers.go`

**Problem:** Two functions do the exact same thing:
- `setupConflictResolverForPull()` (line 241)
- `setupConflictResolverForDirtyPull()` (line 322)

**Difference:** Only operation name and column labels.

**Current duplication:**
- Both call `ListConflictedFiles()`
- Both loop through files
- Both call `ShowConflictVersion(filePath, 1/2/3)` three times
- Both build identical `ConflictResolveState`
- Both set `mode = ModeConflictResolve`

**Fix Required:**

1. Create abstracted function in `githandlers.go`:
```go
// setupConflictResolver is the generic conflict resolver setup function
// operation: "pull_merge", "dirty_pull_changeset_apply", etc.
// columnLabels: ["BASE", "LOCAL (yours)", "REMOTE (theirs)"]
func (a *Application) setupConflictResolver(
    operation string,
    columnLabels []string,
) (tea.Model, tea.Cmd) {
    buffer := ui.GetBuffer()

    buffer.Append("Detecting conflict files...", ui.TypeInfo)

    // Get list of conflicted files from git status
    conflictFiles, err := git.ListConflictedFiles()
    if err != nil {
        buffer.Append(fmt.Sprintf("Error: %v", err), ui.TypeStderr)
        buffer.Append(fmt.Sprintf(OutputMessages["conflict_detection_error"], err), ui.TypeStderr)
        a.asyncOperationActive = false
        a.footerHint = ErrorMessages["operation_failed"]
        a.mode = ModeConsole
        return a, nil
    }

    buffer.Append(fmt.Sprintf("Found %d conflicted file(s)", len(conflictFiles)), ui.TypeInfo)

    if len(conflictFiles) == 0 {
        buffer.Append(OutputMessages["conflict_detection_none"], ui.TypeInfo)
        a.asyncOperationActive = false
        a.mode = ModeConsole
        return a, nil
    }

    // Read 3-way versions for each conflicted file
    resolveState := &ConflictResolveState{
        Files:               make([]ui.ConflictFileGeneric, 0, len(conflictFiles)),
        SelectedFileIndex:   0,
        FocusedPane:         0,
        NumColumns:          len(columnLabels),
        ColumnLabels:        columnLabels,
        ScrollOffsets:       make([]int, len(columnLabels)),
        LineCursors:         make([]int, len(columnLabels)),
        Operation:           operation,
        DiffPane:            nil,
    }

    for _, filePath := range conflictFiles {
        // Build versions array
        versions := make([]string, len(columnLabels))

        // Standard 3-way merge: stage 1=base, 2=local, 3=remote
        for stage := 1; stage <= 3; stage++ {
            content, err := git.ShowConflictVersion(filePath, stage)
            if err != nil {
                content = fmt.Sprintf("Error reading stage %d: %v", stage, err)
            }
            versions[stage-1] = content
        }

        conflictFile := ui.ConflictFileGeneric{
            Path:     filePath,
            Versions: versions,
            Chosen:   -1,
        }
        resolveState.Files = append(resolveState.Files, conflictFile)
    }

    // Store conflict state and transition to resolver UI
    a.conflictResolveState = resolveState
    a.asyncOperationActive = false
    a.mode = ModeConflictResolve
    a.footerHint = fmt.Sprintf(FooterHints["resolve_conflicts_help"], len(conflictFiles))

    buffer.Append(fmt.Sprintf(OutputMessages["conflicts_detected_count"], len(conflictFiles)), ui.TypeInfo)
    buffer.Append(OutputMessages["mark_choices_in_resolver"], ui.TypeInfo)

    return a, nil
}
```

2. Replace `setupConflictResolverForPull()`:
```go
func (a *Application) setupConflictResolverForPull(msg GitOperationMsg) (tea.Model, tea.Cmd) {
    return a.setupConflictResolver(
        "pull_merge",
        []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"},
    )
}
```

3. Replace `setupConflictResolverForDirtyPull()`:
```go
func (a *Application) setupConflictResolverForDirtyPull(msg GitOperationMsg, conflictPhase string) (tea.Model, tea.Cmd) {
    return a.setupConflictResolver(
        "dirty_pull_"+conflictPhase,
        []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"},
    )
}
```

**Benefits:**
- Single source of truth for conflict resolver setup
- Future operations (cherry-pick, branch merge) just call `setupConflictResolver()`
- Easy to add 2-column mode (e.g., cherry-pick = LOCAL vs INCOMING)

**Test:** Run all scenarios, verify no regressions.

---

### Issue 4: Silent Error Fallback (Fail Fast Violation)

**File:** `internal/app/conflicthandlers.go` line 148

**Problem:**
```go
if file.Chosen < 0 || file.Chosen >= len(file.Versions) {
    continue // ❌ Skip invalid choice silently
}
```

**Why this is wrong:**
- If `file.Chosen` is invalid, it's a **programming bug**, not user error
- Silent `continue` masks the bug
- User might get partial resolution (some files applied, some skipped)
- Violates SESSION-LOG.md FAIL-FAST RULE

**Fix Required:**

Replace with fail-fast panic:
```go
if file.Chosen < 0 || file.Chosen >= len(file.Versions) {
    panic(fmt.Sprintf("BUG: Invalid choice for %s: chosen=%d, versions=%d",
        file.Path, file.Chosen, len(file.Versions)))
}
```

**Rationale:**
- `file.Chosen` is set by `handleConflictSpace()` which validates bounds
- `allMarked` check ensures all files have valid choices before reaching this code
- If we reach this with invalid choice, it's a logic error in our code
- Better to crash during development than silently produce wrong result

**Test:** Verify all test scenarios still pass. If panic occurs, fix the root cause.

---

## 📚 P3: Documentation Gaps

### Issue 5: Missing ARCHITECTURE.md Documentation

**File:** `ARCHITECTURE.md`

**Problem:** Conflict resolver integration not documented. Future agents won't know how to wire new operations.

**Missing Sections:**

1. **Conflict Resolver Flow (after line 442)**

Add after "Critical Design Decisions" section:

```markdown
---

## Conflict Resolver Integration Pattern

**Purpose:** Any git operation that can produce merge conflicts must wire into the conflict resolver UI.

**Applicable Operations:**
- Pull (merge strategy)
- Dirty pull (stash → pull → reapply)
- Cherry-pick
- Branch merge
- Rebase (not implemented yet)

### Three-Phase Model

All conflict-resolving operations follow this pattern:

```
Phase 1: START → Execute git operation
    ↓ (if conflicts detected)
Phase 2: RESOLVE → Show conflict resolver UI
    ↓ (user marks files, presses ENTER)
Phase 3: FINALIZE → Complete operation and cleanup
```

**Abort at any phase:** ESC returns to exact original state.

### Adding a New Conflict-Resolving Operation

**Example:** Wire cherry-pick with conflict resolution

**Step 1: Add operation step constants**

File: `internal/app/operationsteps.go`

```go
OpCherryPick         = "cherry_pick"
OpFinalizeCherryPick = "finalize_cherry_pick"
OpAbortCherryPick    = "abort_cherry_pick"
```

**Step 2: Create cmd*() functions**

File: `internal/app/operations.go`

```go
func (a *Application) cmdCherryPick(commitHash string) tea.Cmd {
    hash := commitHash // Capture in closure
    return func() tea.Msg {
        buffer := ui.GetBuffer()
        buffer.Clear()

        result := git.ExecuteWithStreaming("cherry-pick", hash)

        if !result.Success {
            // Check if conflicted
            state, err := git.DetectState()
            if err == nil && state.Operation == git.Conflicted {
                return GitOperationMsg{
                    Step:             OpCherryPick,
                    Success:          false,
                    ConflictDetected: true,
                    Error:            ErrorMessages["cherry_pick_conflicts"],
                }
            }
            return GitOperationMsg{
                Step:    OpCherryPick,
                Success: false,
                Error:   ErrorMessages["cherry_pick_failed"],
            }
        }

        return GitOperationMsg{
            Step:    OpCherryPick,
            Success: true,
            Output:  "Cherry-pick succeeded",
        }
    }
}

func (a *Application) cmdFinalizeCherryPick() tea.Cmd {
    return func() tea.Msg {
        buffer := ui.GetBuffer()

        // Stage all resolved files
        result := git.ExecuteWithStreaming("add", "-A")
        if !result.Success {
            return GitOperationMsg{
                Step:    OpFinalizeCherryPick,
                Success: false,
                Error:   ErrorMessages["failed_stage_resolved"],
            }
        }

        // Continue cherry-pick
        result = git.ExecuteWithStreaming("cherry-pick", "--continue")
        if !result.Success {
            return GitOperationMsg{
                Step:    OpFinalizeCherryPick,
                Success: false,
                Error:   ErrorMessages["failed_cherry_pick_continue"],
            }
        }

        return GitOperationMsg{
            Step:    OpFinalizeCherryPick,
            Success: true,
            Output:  OutputMessages["cherry_pick_finalized"],
        }
    }
}

func (a *Application) cmdAbortCherryPick() tea.Cmd {
    return func() tea.Msg {
        buffer := ui.GetBuffer()

        result := git.ExecuteWithStreaming("cherry-pick", "--abort")
        if !result.Success {
            return GitOperationMsg{
                Step:    OpAbortCherryPick,
                Success: false,
                Error:   ErrorMessages["failed_abort_cherry_pick"],
            }
        }

        return GitOperationMsg{
            Step:    OpAbortCherryPick,
            Success: true,
            Output:  OutputMessages["cherry_pick_aborted"],
        }
    }
}
```

**Step 3: Wire conflict detection**

File: `internal/app/githandlers.go`

Add to `handleGitOperation()`:

```go
case OpCherryPick:
    if msg.ConflictDetected {
        // Cherry-pick conflicts: setup resolver
        return a.setupConflictResolver(
            "cherry_pick",
            []string{"LOCAL (current)", "INCOMING (cherry-pick)"},
        )
    }
    // Success case: reload state
    state, err := git.DetectState()
    if err != nil {
        buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
        a.asyncOperationActive = false
        return a, nil
    }
    a.gitState = state
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
    a.asyncOperationActive = false
    return a, nil

case OpFinalizeCherryPick:
    // Reload state after finalization
    state, err := git.DetectState()
    if err != nil {
        buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
        a.asyncOperationActive = false
        return a, nil
    }
    a.gitState = state
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
    a.asyncOperationActive = false
    a.conflictResolveState = nil
    return a, nil

case OpAbortCherryPick:
    // Reload state after abort
    state, err := git.DetectState()
    if err != nil {
        buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
        a.asyncOperationActive = false
        return a, nil
    }
    a.gitState = state
    buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
    a.asyncOperationActive = false
    a.conflictResolveState = nil
    return a, nil
```

**Step 4: Wire conflict ENTER handler**

File: `internal/app/conflicthandlers.go`

Add to `handleConflictEnter()`:

```go
} else if app.conflictResolveState.Operation == "cherry_pick" {
    // Finalize cherry-pick
    app.asyncOperationActive = true
    app.mode = ModeConsole
    return app, app.cmdFinalizeCherryPick()
```

**Step 5: Wire conflict ESC handler**

File: `internal/app/conflicthandlers.go`

Add to `handleConflictEsc()`:

```go
} else if app.conflictResolveState.Operation == "cherry_pick" {
    // Abort cherry-pick
    app.asyncOperationActive = true
    app.mode = ModeConsole
    ui.GetBuffer().Append(OutputMessages["aborting_cherry_pick"], ui.TypeInfo)
    return app, app.cmdAbortCherryPick()
```

**Step 6: Add SSOT messages**

File: `internal/app/messages.go`

```go
// ErrorMessages
"cherry_pick_conflicts":      "Cherry-pick conflicts occurred",
"cherry_pick_failed":         "Failed to cherry-pick commit",
"failed_cherry_pick_continue": "Failed to continue cherry-pick",
"failed_abort_cherry_pick":   "Failed to abort cherry-pick",

// OutputMessages
"cherry_pick_finalized": "Cherry-pick completed successfully",
"cherry_pick_aborted":   "Cherry-pick aborted - state restored",
"aborting_cherry_pick":  "Aborting cherry-pick...",
```

**Done!** Cherry-pick now has full conflict resolution support.

---

### Pattern Summary

For any new conflict-resolving operation:

1. ✅ Add 3 constants: `Op*`, `OpFinalize*`, `OpAbort*`
2. ✅ Create 3 functions: `cmd*()`, `cmdFinalize*()`, `cmdAbort*()`
3. ✅ Wire `handleGitOperation()` with 3 cases
4. ✅ Wire `handleConflictEnter()` routing
5. ✅ Wire `handleConflictEsc()` abort routing
6. ✅ Add SSOT messages to messages.go

This pattern is now proven for pull merge (Scenario 1 working).
```

**Test:** Future agents should be able to implement cherry-pick by following this guide.

---

## ✅ Testing Checklist

After all fixes applied:

**Scenario 1: Pull with Conflicts (Clean Tree)**
- [ ] Conflicts detected correctly
- [ ] Resolver UI appears with correct file list
- [ ] Column labels show: BASE, LOCAL (yours), REMOTE (theirs)
- [ ] User can mark files with SPACE
- [ ] Footer shows semantic column label (not "column 2")
- [ ] ENTER finalizes: stages files, commits merge, returns Clean + Ahead
- [ ] ESC aborts: runs `git merge --abort`, returns Clean + Diverged
- [ ] No hardcoded strings in console output
- [ ] All messages come from messages.go SSOT

**Build Test:**
- [ ] `./build.sh` compiles cleanly (no errors, no warnings)
- [ ] All operation step constants exist in operationsteps.go
- [ ] All user-facing text in messages.go

**Code Audit:**
- [ ] No hardcoded strings in operations.go
- [ ] No hardcoded strings in conflicthandlers.go
- [ ] No hardcoded strings in githandlers.go
- [ ] All operations use Op* constants from operationsteps.go
- [ ] setupConflictResolver() abstraction eliminates duplication

---

## 📝 Notes for Future Agents

**Golden Rules:**

1. **SSOT for all text:** User-facing strings MUST be in messages.go
2. **SSOT for operation names:** Constants MUST be in operationsteps.go
3. **Fail fast:** Invalid state = panic, not silent fallback
4. **One pattern:** All conflict operations follow the 3-phase model
5. **Abstract duplication:** If two functions look the same, merge them

**Testing Strategy:**

- Fix one issue at a time
- Build after each fix
- Test Scenario 1 after each fix
- If regression occurs, revert and investigate

**Session Log:**

- Document what was fixed
- Note any surprises or edge cases discovered
- Keep log entry concise (user decides what to log, not agent)

---

**End of Document**
