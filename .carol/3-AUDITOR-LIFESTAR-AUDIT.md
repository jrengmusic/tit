# Sprint 3 Audit Report

**Date:** 2026-01-30  
**Scope:** Full codebase (`/Users/jreng/Documents/Poems/dev/tit/internal/`)  
**Auditor:** Amp (Claude Sonnet 4)  
**Summary:** Critical: 0, High: 4, Medium: 6, Low: 5

---

## REFACTORING OPPORTUNITIES (CRITICAL PRIORITY)

### [REF-001] Silent Error Fallbacks in State Detection
**File:** [`internal/git/state.go:93-116`](file:///Users/jreng/Documents/Poems/dev/tit/internal/git/state.go#L93-L116)  
**Principle:** LOVE: Listens (Fail Fast), LOVE: Validates  
**Issue:** `DetectState()` silently falls back to "safe" defaults when git commands fail:
```go
workingTree, err := detectWorkingTree()
if err != nil {
    state.WorkingTree = Clean  // Silent fallback - DANGEROUS
}
```
Same pattern for `detectOperation()`, `detectRemote()`, `detectTimeline()`.

**Benefits:** Prevent user from pushing broken code thinking tree is clean  
**Impact:** HIGH (incorrect state can cause data loss)  
**Effort:** Medium (add DetectionWarnings field to State struct)  
**Priority:** HIGH  
**Fix:** Add `DetectionWarnings []string` field to `State` struct. Propagate warnings to UI footer.

---

### [REF-002] Panic in Non-Critical Path
**File:** [`internal/git/execute.go:122-124`](file:///Users/jreng/Documents/Poems/dev/tit/internal/git/execute.go#L122-L124)  
**Principle:** LOVE: Validates, LIFESTAR: Accessible  
**Issue:** `FindStashRefByHash()` panics when stash not found:
```go
panic(fmt.Sprintf("FATAL: Stash with hash %s not found..."))
```
User manually dropping a stash crashes the entire application.

**Benefits:** Graceful error handling, better UX  
**Impact:** HIGH (application crash)  
**Effort:** Low (return error instead of panic)  
**Priority:** HIGH  
**Fix:** Return `(string, error)` instead of panicking. Show user-friendly error message.

---

### [REF-003] Early Returns Masking State
**File:** [`internal/git/state.go:214-300`](file:///Users/jreng/Documents/Poems/dev/tit/internal/git/state.go#L214-L300)  
**Principle:** LOVE: Listens (Fail Fast), LIFESTAR: Findable  
**Issue:** `detectTimeline()` has 10+ early returns all returning `InSync, 0, 0, nil`:
- Lines 228-229, 232-233, 245-250, 254, 273, 282, 287, 291-296

Any parsing failure silently reports "InSync" even when actually diverged.

**Benefits:** Accurate timeline state, prevent missed updates  
**Impact:** HIGH (user may miss critical remote changes)  
**Effort:** Medium (add confidence/warning tracking)  
**Priority:** HIGH  
**Fix:** Add `TimelineConfidence` field (Certain/Unknown). Show warning when Unknown.

---

### [REF-004] Global Singleton OutputBuffer
**File:** [`internal/ui/buffer.go:37-44`](file:///Users/jreng/Documents/Poems/dev/tit/internal/ui/buffer.go#L37-L44)  
**Principle:** LIFESTAR: Explicit, LIFESTAR: Testable  
**Issue:** `GetBuffer()` returns global singleton - any code can mutate shared state. Cannot test console output in isolation.

**Benefits:** Testability, explicit dependencies  
**Impact:** MEDIUM (testing difficulty, hidden coupling)  
**Effort:** Medium (inject via Application constructor)  
**Priority:** MEDIUM  
**Fix:** Inject buffer via `NewApplication()` constructor. Pass to git operations.

---

## LIFESTAR VIOLATIONS

### [AUD-001] Lean Violation: Application God Object
**File:** [`internal/app/app.go:30-94`](file:///Users/jreng/Documents/Poems/dev/tit/internal/app/app.go#L30-L94)  
**Principle:** Lean (Keep It Simple)  
**Issue:** `Application` struct has 24+ fields managing UI, git, menu, input, workflow, async, console, dialog, conflict, dirty operation, picker, time travel, environment, cache, config, and activity states.  
**Severity:** Medium  
**Impact:** Cognitive load when navigating codebase  
**Fix:** Current structure is acceptable - state is already grouped into sub-structs (PickerState, WorkflowState, etc.). Consider extracting more sub-structs if growth continues.

---

### [AUD-002] Explicit Violation: Package-Level Logger
**File:** [`internal/git/types.go:118-124`](file:///Users/jreng/Documents/Poems/dev/tit/internal/git/types.go#L118-L124)  
**Principle:** Explicit (Dependencies Visible)  
**Issue:** `SetLogger()` sets package-level global:
```go
var packageLogger Logger
func SetLogger(l Logger) { packageLogger = l }
```
Git package cannot be tested without app initialization.  
**Severity:** Medium  
**Fix:** Low priority - Logger interface allows mocking. Consider context injection if testing becomes problematic.

---

### [AUD-003] Testable Violation: Hardcoded exec.Command
**File:** [`internal/git/state.go:183`](file:///Users/jreng/Documents/Poems/dev/tit/internal/git/state.go#L183)  
**Principle:** Testable (Pure functions, DI)  
**Issue:** Git operations use hardcoded `exec.Command("git", ...)` - no way to inject test doubles.  
**Severity:** Medium  
**Fix:** Extract interface: `type GitExecutor interface { Run(args ...string) CommandResult }`. Low priority - integration tests may be sufficient for git-centric app.

---

### [AUD-004] Reviewable Violation: Redundant Handler Parameters
**File:** [`internal/app/app.go:243-284`](file:///Users/jreng/Documents/Poems/dev/tit/internal/app/app.go#L243-L284)  
**Principle:** Reviewable (Consistent Style)  
**Issue:** `handleMenuUp/Down` take redundant `app *Application` param when receiver `a` is the same:
```go
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd)
```
Pattern repeated in all handler files.  
**Severity:** Low  
**Fix:** Standardize on receiver-only pattern OR document why parameter passing is needed.

---

## ANTI-PATTERNS DETECTED

### [ANT-001] Layer Violation: UI Imports Git Types
**Files:** [`internal/ui/history.go:8-9`](file:///Users/jreng/Documents/Poems/dev/tit/internal/ui/history.go#L8-L9), [`internal/ui/filehistory.go:5-6`](file:///Users/jreng/Documents/Poems/dev/tit/internal/ui/filehistory.go#L5-L6)  
**Issue:** UI layer imports `tit/internal/git` directly:
```go
type CommitInfo = git.CommitInfo
type FileInfo = git.FileInfo
```
**Severity:** Low  
**Impact:** UI has knowledge of git layer. Codebase acknowledges this with comment "alias for git.X to avoid import cycles."  
**Fix:** Optional - move shared types to `tit/internal/types` package. Current pragmatic approach is acceptable.

---

### [ANT-002] Hidden State: determineTimeline Fallback
**File:** [`internal/git/state.go:302-316`](file:///Users/jreng/Documents/Poems/dev/tit/internal/git/state.go#L302-L316)  
**Issue:** `determineTimeline()` has unreachable fallback:
```go
if timeline, exists := stateMap[key]; exists {
    return timeline
}
return InSync // fallback - when would this ever hit?
```
The map covers all boolean combinations, so fallback is dead code.  
**Severity:** Low  
**Fix:** Remove unreachable fallback or add comment explaining defensive coding.

---

## SPEC DISCREPANCIES (DOC UPDATE RECOMMENDATIONS)

### [DOC-001] SPEC.md Accurate
**Status:** ✅ No discrepancies found  
**Note:** SPEC.md matches codebase implementation for 5-axis state model, menu generation, and operations.

### [DOC-002] ARCHITECTURE.md Accurate
**Status:** ✅ No discrepancies found  
**Note:** ARCHITECTURE.md accurately documents startup flow, state detection order, and mode handling.

---

## POSITIVE OBSERVATIONS

The codebase exhibits strong LIFESTAR compliance in several areas:

1. **SSOT for Git State:** `DetectState()` is clearly documented as single source of truth (line 51-68)
2. **Explicit State Grouping:** Application uses dedicated sub-structs (PickerState, WorkflowState, etc.)
3. **Findable Code:** Clear file naming (`op_*.go` for operations, `handlers_*.go` for handlers)
4. **Immutable Design:** State detection returns new struct, doesn't mutate globals
5. **Reviewable Comments:** Critical functions have contract documentation

---

## SUMMARY

### By Category
| Category | Count |
|----------|-------|
| Refactoring Opportunities | 4 |
| LIFESTAR Violations | 4 |
| Anti-Patterns | 2 |
| Doc Updates Needed | 0 |

### By Severity
| Severity | Count |
|----------|-------|
| CRITICAL | 0 |
| High | 4 (REF-001, REF-002, REF-003, REF-004) |
| Medium | 3 |
| Low | 5 |

### Recommended Actions

1. **[HIGH] REF-001:** Add `DetectionWarnings` to `State` struct to surface silent fallbacks
2. **[HIGH] REF-002:** Replace panic in `FindStashRefByHash()` with error return
3. **[HIGH] REF-003:** Add `TimelineConfidence` field to indicate detection reliability
4. **[MEDIUM] REF-004:** Inject OutputBuffer via constructor for testability
5. **[LOW]** Consider extracting `GitExecutor` interface for unit testing
6. **[LOW]** Standardize handler function signatures

---

**Audit Complete.**

The codebase is well-architected with strong LIFESTAR adherence. The main concerns are:
- Silent error fallbacks in state detection (can mislead users)
- Panic usage in non-critical path (crashes application)
- Early return patterns that mask state uncertainty

These are all fixable with focused refactoring sessions.

---

*AUDITOR flags issues, doesn't fix them. User decides priority.*
