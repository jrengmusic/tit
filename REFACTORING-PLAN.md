# REFACTORING-PLAN.md

**Project:** TIT Codebase Refactoring  
**Date:** 2026-01-31  
**Approach:** Single Sprint - All violations fixed in one session  

---

## Goal

Fix all critical architectural violations in **one sprint**:
- God Object (Application: 25+ fields → <15 fields)
- 300 LOC violations (7 files → all <300 LOC)
- Massive switch statements (290 lines → registry pattern)
- Code duplication (10+ patterns → extracted helpers)

---

## Constraints

1. Clean build after each sub-step: `go build ./...`
2. Clean vet after each sub-step: `go vet ./...`
3. All features must work at end
4. No breaking changes

---

## Sprint Structure: 5 Sub-Steps

Each sub-step is independent, builds on previous, verified before proceeding.

---

## Sub-Step 1: State Cluster Extraction

**Goal:** Extract 4 cohesive state clusters from Application struct.

**Files to create:**
- `internal/app/ui_state.go` - width, height, sizing, theme, footerHint
- `internal/app/navigation_state.go` - mode, selectedIndex, menuItems, keyHandlers
- `internal/app/operation_state.go` - asyncState, workflowState, dirtyOperationState, timeTravelState, conflictResolveState
- `internal/app/dialog_manager.go` - dialogState, pickerState, inputState, consoleState

**Pattern:** Embed in Application struct for direct field access:
```go
type Application struct {
    *UIState
    *NavigationState
    *OperationState
    *DialogManager
    // ... remaining fields
}
```

**Verification:**
```bash
go build ./... && go vet ./...
```
- [ ] App launches
- [ ] Menu navigation works
- [ ] Dialogs display
- [ ] Operations run

**Estimated time:** 2 hours  
**Lines changed:** ~300 added, ~100 modified

---

## Sub-Step 2: Handler Registry Framework

**Goal:** Replace git_handlers.go switch statement with registry.

**Files to create:**
- `internal/app/handlers/registry.go` - Handler interface and registry
- `internal/app/handlers/init.go` - Init/Clone handlers
- `internal/app/handlers/remote.go` - AddRemote/FetchRemote handlers
- `internal/app/handlers/pull.go` - Pull/Merge handlers
- `internal/app/handlers/commit.go` - Commit/Push handlers
- `internal/app/handlers/timetravel.go` - Time travel handlers
- `internal/app/handlers/conflict.go` - Conflict handlers

**Pattern:**
```go
type GitOperationHandler interface {
    CanHandle(step string) bool
    Handle(app *Application, msg git.GitOperationMsg) (tea.Model, tea.Cmd)
}

var registry = &Registry{}

func init() {
    registry.Register(&InitHandler{})
    registry.Register(&PullHandler{})
    // ... etc
}
```

**Migration:** Copy each switch case to handler, remove from original.

**Verification:**
```bash
go build ./... && go vet ./...
```
- [ ] Init works
- [ ] Clone works
- [ ] Pull works
- [ ] Push works
- [ ] Time travel works
- [ ] git_handlers.go deleted

**Estimated time:** 3 hours  
**Lines changed:** ~600 added, ~700 removed

---

## Sub-Step 3: Duplicate Code Elimination

**Goal:** Extract 3 common patterns to helpers.

**Files to modify:**
- `internal/app/app.go` - add helpers
- `internal/app/confirm_handlers.go` - use helpers
- `internal/app/dispatchers.go` - use helpers
- `internal/app/handlers_config.go` - use helpers

**Helpers to add:**
```go
func (a *Application) EnterConsoleMode(hint string)
func (a *Application) ShowConfirmation(confirmType string, context map[string]string)
func (a *Application) RegenerateMenu()
```

**Pattern:** Find all 10+ occurrences, replace with single call.

**Verification:**
```bash
go build ./... && go vet ./...
```
- [ ] Console transitions work
- [ ] All confirmations work
- [ ] Menu regeneration works

**Estimated time:** 1.5 hours  
**Lines changed:** ~150 modified (net reduction: ~200 lines)

---

## Sub-Step 4: File Splitting

**Goal:** Split 3 oversized files into <300 LOC each.

**Files to split:**

**A. confirm_handlers.go (866 → <300 per file)**
```
internal/app/confirm/
├── base.go           # Common infrastructure
├── time_travel.go    # Time travel confirmations
├── branch_switch.go  # Branch switch confirmations
├── dirty_ops.go      # Dirty pull/push confirmations
└── push.go           # Force push confirmations
```

**B. git/execute.go (949 → <300 per file)**
```
internal/git/execute/
├── base.go           # Command execution infrastructure
├── clone.go          # Clone operations
├── commit.go         # Commit operations
├── push.go           # Push operations
├── pull.go           # Pull operations
└── timetravel.go     # Time travel operations
```

**C. dispatchers.go (669 → <300 per file)**
```
internal/app/dispatch/
├── registry.go       # Dispatcher registry
├── menu.go           # Menu action dispatchers
├── git.go            # Git operation dispatchers
└── dialog.go         # Dialog action dispatchers
```

**Pattern:** Move functions to new files, update imports, delete old.

**Verification:**
```bash
go build ./... && go vet ./...
```
- [ ] All files <300 LOC
- [ ] No import cycles
- [ ] All dispatchers work

**Estimated time:** 2 hours  
**Lines changed:** ~0 net (just reorganized)

---

## Sub-Step 5: Final Cleanup

**Goal:** Extract magic numbers, centralize strings, final Application facade.

**Files to modify:**
- `internal/app/constants.go` - add magic number constants
- `internal/app/messages.go` - add missing strings
- `internal/app/app.go` - final Application struct cleanup

**Constants to add:**
```go
const (
    QuitConfirmTimeout   = 2 * time.Second
    CacheRefreshInterval = 100 * time.Millisecond
    StashSearchLimit     = 10
    DefaultFilePerms     = 0755
    InputHeight          = 4
)
```

**Strings to centralize:**
- "main" → constants.DefaultBranch
- "HEAD" → constants.HEADRef
- "Skipped: cannot set upstream in detached HEAD state" → messages.CannotSetUpstreamDetached
- "Nothing to commit (working tree clean)" → messages.WorkingTreeClean

**Final Application struct:**
```go
type Application struct {
    // State (4 embedded clusters)
    *UIState
    *NavigationState
    *OperationState
    *DialogManager
    
    // Core dependencies (6 fields)
    gitState     *git.State
    appConfig    *config.Config
    cacheManager *cache.Manager
    activityState *ActivityState
    environmentState *EnvironmentState
    
    // Renderers (2 fields)
    viewRenderer *view.Renderer
}
```

**Verification:**
```bash
go build ./... && go vet ./...
```
- [ ] Application struct <15 fields
- [ ] No magic numbers
- [ ] All strings centralized
- [ ] All features work

**Estimated time:** 1.5 hours  
**Lines changed:** ~100 modified

---

## Verification Matrix

After each sub-step, verify:

| Check | Command/Action |
|-------|----------------|
| Build | `go build ./...` |
| Vet | `go vet ./...` |
| Launch | `./tit` in repo |
| Menu | Navigate up/down |
| Init | Create new repo |
| Clone | Clone existing repo |
| Commit | Stage + commit |
| Push | Push to remote |
| Pull | Pull changes |
| Time Travel | Checkout old commit |
| Return | Return to branch |
| Conflict | Resolve merge conflict |

---

## Total Estimates

| Sub-Step | Time | Lines Added | Lines Removed |
|----------|------|-------------|---------------|
| 1. State Clusters | 2h | 300 | 100 |
| 2. Handler Registry | 3h | 600 | 700 |
| 3. Duplicate Elimination | 1.5h | 50 | 250 |
| 4. File Splitting | 2h | 0 | 0 |
| 5. Final Cleanup | 1.5h | 100 | 50 |
| **TOTAL** | **10h** | **1050** | **1100** |

**Net change:** ~50 lines reduced, massive architectural improvement.

---

## Execution Order

```
Sub-Step 1 → Sub-Step 2 → Sub-Step 3 → Sub-Step 4 → Sub-Step 5
    ↓            ↓            ↓            ↓            ↓
  Verify       Verify       Verify       Verify       Verify
    ↓            ↓            ↓            ↓            ↓
  Commit       Commit       Commit       Commit       Commit
```

Each sub-step is **independent and reversible**. If issues arise, rollback to previous commit.

---

## Success Criteria

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] Application struct <15 fields (from 25+)
- [ ] All files <300 LOC
- [ ] No switch statements >50 lines
- [ ] All features work (see verification matrix)

---

## Ready to Execute

This is a **single sprint** - approximately 10 hours of focused work.

**Prerequisites:**
- Feature freeze
- Clean working directory
- Coffee ready

**Ready when you are.**

---

**End of Refactoring Plan**
