# Sprint 4: AUDITOR Fixes - Kickoff Plan

**Date:** 2026-01-31
**Objective:** Address AUDITOR's high-priority findings from Sprint 3 audit
**Scope:** 4 HIGH priority refactoring opportunities
**Estimated Duration:** 4-6 hours

---

## Pre-Flight

**Read Required:**
1. `.carol/3-AUDITOR-LIFESTAR-AUDIT.md` - Full audit report
2. `internal/git/state.go` - State detection code
3. `internal/git/execute.go` - Git execution code
4. `internal/ui/buffer.go` - OutputBuffer singleton

---

## Phase 1: Fix Silent Error Fallbacks (REF-001)

**File:** `internal/git/state.go:93-116`
**Issue:** `DetectState()` silently falls back to defaults when git commands fail
**Risk:** User may push broken code thinking tree is clean
**Estimated Time:** 90 minutes

### Implementation:

**Step 1: Add `DetectionWarnings` field to `State` struct**

```go
// State represents the complete git repository state (5 axes)
type State struct {
    WorkingTree      WorkingTree
    Timeline         Timeline
    Operation        Operation
    Remote           Remote
    CommitsAhead     int
    CommitsBehind    int
    
    // NEW: Detection warnings for failed state detection
    DetectionWarnings []string
}
```

**Step 2: Update `DetectState()` to capture warnings**

```go
func DetectState() (*State, error) {
    state := &State{
        DetectionWarnings: make([]string, 0),
    }
    
    // Detect working tree state
    workingTree, err := detectWorkingTree()
    if err != nil {
        state.DetectionWarnings = append(state.DetectionWarnings, 
            fmt.Sprintf("Working tree detection failed: %v", err))
        state.WorkingTree = Clean // Safe fallback with warning
    } else {
        state.WorkingTree = workingTree
    }
    
    // Same pattern for operation, remote, timeline...
    
    return state, nil
}
```

**Step 3: Display warnings in UI footer**

In `app.go` or `footer.go`:
```go
func (a *Application) renderFooter() string {
    // Show detection warnings if any
    if len(a.gitState.DetectionWarnings) > 0 {
        return fmt.Sprintf("⚠️ %s", a.gitState.DetectionWarnings[0])
    }
    return a.footerHint
}
```

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
go test ./internal/git/...
```

---

## Phase 2: Replace Panic with Error (REF-002)

**File:** `internal/git/execute.go:122-124`
**Issue:** `FindStashRefByHash()` panics when stash not found
**Risk:** Application crash if user manually drops stash
**Estimated Time:** 45 minutes

### Implementation:

**Step 1: Change function signature**

```go
// BEFORE:
func FindStashRefByHash(stashHash string) string {
    // ...
    panic(fmt.Sprintf("FATAL: Stash with hash %s not found..."))
}

// AFTER:
func FindStashRefByHash(stashHash string) (string, error) {
    // ...
    if stashRef == "" {
        return "", fmt.Errorf("stash with hash %s not found", stashHash)
    }
    return stashRef, nil
}
```

**Step 2: Update all call sites**

Find all callers:
```bash
grep -r "FindStashRefByHash" internal/
```

Update each call site:
```go
// BEFORE:
stashRef := FindStashRefByHash(hash)

// AFTER:
stashRef, err := FindStashRefByHash(hash)
if err != nil {
    // Handle gracefully - show error to user
    buffer.Append(fmt.Sprintf("Error: %v", err), ui.TypeStderr)
    return GitOperationMsg{Success: false, Error: err.Error()}
}
```

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Test by manually dropping a stash during time travel
```

---

## Phase 3: Fix Timeline Detection Confidence (REF-003)

**File:** `internal/git/state.go:214-300`
**Issue:** `detectTimeline()` has 10+ early returns all returning `InSync` on failure
**Risk:** User may miss critical remote changes
**Estimated Time:** 90 minutes

### Implementation:

**Step 1: Add `TimelineConfidence` enum**

```go
type TimelineConfidence int

const (
    TimelineConfidenceCertain TimelineConfidence = iota
    TimelineConfidenceUnknown
)
```

**Step 2: Add to State struct**

```go
type State struct {
    // ... existing fields ...
    TimelineConfidence TimelineConfidence
}
```

**Step 3: Update `detectTimeline()` to track confidence**

```go
func detectTimeline() (Timeline, int, int, TimelineConfidence, error) {
    // Check if we can detect timeline
    hasRemote := checkRemoteExists()
    if !hasRemote {
        return "", 0, 0, TimelineConfidenceCertain, nil
    }
    
    // Try to get ahead/behind counts
    ahead, behind, err := getAheadBehindCounts()
    if err != nil {
        // Detection failed - mark as unknown
        return "", 0, 0, TimelineConfidenceUnknown, err
    }
    
    // Parse results
    timeline, err := parseTimeline(ahead, behind)
    if err != nil {
        return "", 0, 0, TimelineConfidenceUnknown, err
    }
    
    return timeline, ahead, behind, TimelineConfidenceCertain, nil
}
```

**Step 4: Display confidence warning in UI**

```go
func (a *Application) RenderStateHeader() string {
    // ... existing code ...
    
    if a.gitState.TimelineConfidence == TimelineConfidenceUnknown {
        timelineDesc += " (detection uncertain)"
    }
    
    // ... rest of rendering ...
}
```

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Test with network disconnected to trigger detection failure
```

---

## Phase 4: Inject OutputBuffer (REF-004)

**File:** `internal/ui/buffer.go:37-44`
**Issue:** Global singleton `GetBuffer()` - cannot test in isolation
**Risk:** Testing difficulty, hidden coupling
**Estimated Time:** 120 minutes

### Implementation:

**Step 1: Add buffer to Application struct**

```go
type Application struct {
    // ... existing fields ...
    outputBuffer *ui.OutputBuffer  // NEW: Injected buffer
}
```

**Step 2: Update `NewApplication()` to create buffer**

```go
func NewApplication(sizing ui.DynamicSizing, theme ui.Theme, cfg *config.Config) *Application {
    app := &Application{
        // ... existing fields ...
        outputBuffer: ui.NewOutputBuffer(),  // NEW
    }
    // ...
}
```

**Step 3: Pass buffer to git operations**

Option A - Pass via context:
```go
func (a *Application) executeGitOp(step string, args ...string) tea.Cmd {
    ctx := context.WithValue(context.Background(), "buffer", a.outputBuffer)
    return func() tea.Msg {
        result := git.ExecuteWithStreaming(ctx, args...)
        // ...
    }
}
```

Option B - Pass as parameter (preferred):
```go
func ExecuteWithStreaming(ctx context.Context, buffer *ui.OutputBuffer, args ...string) CommandResult {
    // Use buffer directly instead of global GetBuffer()
}
```

**Step 4: Update all call sites**

Replace global `ui.GetBuffer()` calls with `a.outputBuffer` where appropriate.

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
go test ./internal/ui/...
```

---

## Summary of Changes

| Phase | File | Change | Lines |
|-------|------|--------|-------|
| 1 | `internal/git/types.go` | Add `DetectionWarnings` to State | +2 |
| 1 | `internal/git/state.go` | Capture warnings in DetectState | +20 |
| 1 | `internal/app/footer.go` | Display warnings | +10 |
| 2 | `internal/git/execute.go` | Change FindStashRefByHash signature | +3 |
| 2 | Multiple files | Update call sites | ~10 |
| 3 | `internal/git/types.go` | Add TimelineConfidence enum | +5 |
| 3 | `internal/git/state.go` | Track confidence in detectTimeline | +15 |
| 3 | `internal/app/app.go` | Display confidence warning | +5 |
| 4 | `internal/app/app.go` | Add outputBuffer field | +1 |
| 4 | `internal/ui/buffer.go` | Update ExecuteWithStreaming | +5 |
| 4 | Multiple files | Replace GetBuffer() calls | ~30 |

**Total:** ~100 lines changed across ~15 files

---

## Critical Rules

1. **Clean Build After Every Phase**
   ```bash
   go build ./...
   ```

2. **No Logic Changes**
   - Only add error handling/warnings
   - Don't change git command behavior
   - Maintain existing defaults (just warn about them)

3. **Test Edge Cases**
   - Network disconnected (timeline detection)
   - Manually drop stash during time travel
   - Corrupt git state

4. **Maintain Backward Compatibility**
   - State struct changes are additive only
   - Existing code should continue to work

---

## Rollback Plan

If any phase fails:
1. Stop immediately
2. Do not proceed to next phase
3. Report issue to user
4. User decides: fix forward or git revert

---

**End of Kickoff Plan**

Ready for ENGINEER to begin Phase 1.
