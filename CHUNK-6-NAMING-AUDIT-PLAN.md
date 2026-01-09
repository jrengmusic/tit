# Chunk 6: Handler Function Naming Audit - Detailed Plan

**Status:** In Progress  
**Date:** Session 72 (2026-01-10)  
**Scope:** Rename functions to match AGENTS.md + "How to choose your words wisely.md" standards

---

## Naming Standards Applied

From `CLAUDE.md` + AGENTS.md:
- **`handle*`** — Input/keyboard handlers (key press, menu selection)
- **`execute*`** — Logic operations (git operations, side effects)
- **`cmd*`** — Return `tea.Cmd` (async operations)
- **`dispatch*`** — Action routing (menu item → mode transition)

From "How to choose your words wisely.md" (Rule 3: Semantic over Literal):
- Use verbs that accurately describe what the function does
- `get` for pure retrieval; `getOrCreate` when function has side effects
- `execute` for running logic; `cmd` for returning Bubble Tea commands
- Boolean patterns: `is*`, `should*`, `has*`
- Avoid ambiguous or misleading verbs

---

## Functions to Rename

### Category A: `execute*` → `cmd*` (Return tea.Cmd)

These functions return `tea.Cmd` and run async background operations. Per AGENTS.md, they should use `cmd*` prefix.

| Current Name | New Name | File | Return Type | Reason |
|---|---|---|---|---|
| `executeCloneWorkflow` | `cmdCloneWorkflow` | handlers.go:455 | `tea.Cmd` | Returns Bubble Tea command, not (tea.Model, tea.Cmd) |

**Count:** 1 function

---

### Category B: `execute*` Confirmation Actions (Keep prefix, but verify semantics)

These functions return `(tea.Model, tea.Cmd)` and represent confirmation action handlers. They're not async background commands—they're event handlers that return models.

According to "How to choose your words wisely.md" Rule 1 (Word classes match role):
- If function **returns a command**, it should use `cmd*`
- If function **executes logic with side effects**, `execute*` is correct
- If function **handles user input**, `handle*` is correct

**Analysis:**
- `executeConfirm*` → These **handle confirmation choices** and return model changes
- Per Rule 1, these should be `handleConfirm*` (they handle user confirmation, return model)
- OR stay `executeConfirm*` if the focus is on the **action being executed** (semantic ambiguity here)

**Decision:** Keep `executeConfirm*` (these execute the confirmed action, not just handle the UI). Clear semantic intent: "execute the confirmed force push", not "handle the confirm button".

However, for clarity and consistency with naming conventions:

| Current Name | Semantic Intent | Category | Keep/Change |
|---|---|---|---|
| `executeConfirmNestedRepoInit` | Execute the nested repo init action (user confirmed) | Event handler + action executor | ✅ Keep |
| `executeConfirmForcePush` | Execute force push (user confirmed) | Event handler + action executor | ✅ Keep |
| `executeConfirmHardReset` | Execute hard reset (user confirmed) | Event handler + action executor | ✅ Keep |
| `executeConfirmDirtyPull` | Handle dirty pull scenario (user confirmed) | Event handler | ✅ Keep (semantic: execute the pull, not just handle UI) |
| `executeConfirmPullMerge` | Execute merge after pull (user confirmed) | Event handler + action executor | ✅ Keep |
| `executeConfirmTimeTravel` | Execute time travel checkout (user confirmed) | Event handler + action executor | ✅ Keep |
| `executeConfirmTimeTravelReturn` | Execute return from time travel (user confirmed) | Event handler + action executor | ✅ Keep |
| `executeConfirmTimeTravelMerge` | Execute merge from time travel (user confirmed) | Event handler + action executor | ✅ Keep |

**Rationale:** These differ from pure event handlers (`handle*`) because they execute business logic, not just process input. They differ from `cmd*` because they return immediately with model changes, not async operations. `execute*` is semantically correct.

**Count:** 0 renames (verify semantics only)

---

### Category C: `execute*` Time Travel Operations (Verify semantics)

These functions are time travel choreography—they run git operations and return model changes. Semantic analysis:

| Current Name | Return Type | Semantic Intent | Category | Analysis |
|---|---|---|---|---|
| `executeTimeTravelClean` | `(tea.Model, tea.Cmd)` | Checkout commit (no stash needed) | Event handler for user action | ✅ OK as `execute*` (executes time travel transition) |
| `executeTimeTravelWithDirtyTree` | `(tea.Model, tea.Cmd)` | Checkout commit (stash dirty tree) | Event handler for user action | ✅ OK as `execute*` (executes time travel transition) |

**Rationale:** These execute complex multi-step git transitions. `execute*` accurately describes the intent (execute the time travel choreography), not just "handle" the input. The semantic intent is action-focused, not input-focused.

**Count:** 0 renames (semantics correct)

---

## Summary: No Breaking Renames Needed

**Key Finding:** Current naming is already semantically correct per AGENTS.md and "How to choose your words wisely.md":

1. ✅ **Dispatcher functions** (`dispatch*`) — Route actions correctly
2. ✅ **Handler functions** (`handle*`) — Process input correctly
3. ✅ **Command functions** (`cmd*`) — Async operations correctly
4. ✅ **Execution functions** (`execute*`) — Run business logic/confirmations correctly

**Violations Found:** Only 1
- `executeCloneWorkflow` → Should be `cmdCloneWorkflow` (returns `tea.Cmd`, not `(tea.Model, tea.Cmd)`)

**Analysis:** The `executeConfirm*` and `executeTimeTravel*` functions use correct semantics:
- They "execute" confirmed actions (not just "handle" input)
- They represent action choreography, not input processing
- They return `(tea.Model, tea.Cmd)`, consistent with handler pattern
- Renaming them `handleConfirm*` would obscure the semantic meaning (action execution, not input handling)

---

## Recommended Action

**Minimal Change:**
1. Rename `executeCloneWorkflow` → `cmdCloneWorkflow` (1 rename)
2. Update 1 caller in handlers.go (line 455)
3. Update reference in `confirmationhandlers.go` (if any)
4. Verify compile + smoke test

**No other changes needed.** Current naming is semantically sound.

---

## Verification Steps

```bash
# 1. Find all usages of executeCloneWorkflow
grep -r "executeCloneWorkflow" internal/

# 2. Rename in confirmationhandlers.go
# Line 455: func (a *Application) executeCloneWorkflow() tea.Cmd {
#    →  func (a *Application) cmdCloneWorkflow() tea.Cmd {

# 3. Update handler call in handlers.go
# Find: a.executeCloneWorkflow()
#    →  a.cmdCloneWorkflow()

# 4. Build and test
./build.sh
```

---

## Files to Modify

- `internal/app/handlers.go` — Line 455 (rename function def + caller)
- Grep for usages to find all call sites

---

## Rule Application (How to choose your words wisely.md)

**Rule 1: Word classes match role**
- ✅ `cmd*` for async operations (returns `tea.Cmd`)
- ✅ `execute*` for action execution (runs business logic)
- ✅ `handle*` for input processing

**Rule 3: Semantic over literal**
- ✅ `cmdCloneWorkflow` = "Create a command that runs clone workflow" (accurate)
- ✅ `executeConfirmForcePush` = "Execute the confirmed force push action" (accurate)
- ✅ NOT `handleConfirmForcePush` which would imply input handling, not action execution

**Rule 5: Consistency**
- ✅ All `cmd*` functions return `tea.Cmd`
- ✅ All `handle*` functions return `(tea.Model, tea.Cmd)`
- ✅ All `execute*` functions return either logic result or `(tea.Model, tea.Cmd)`

---

## Decision: Proceed with 1 Rename

**Function to rename:**
```go
// Before:
func (a *Application) executeCloneWorkflow() tea.Cmd {

// After:
func (a *Application) cmdCloneWorkflow() tea.Cmd {
```

**Rationale:** Returns `tea.Cmd`, matches `cmd*` pattern for async operations per AGENTS.md.
