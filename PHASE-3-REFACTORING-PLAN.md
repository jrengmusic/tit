# Phase 3: Incremental Refactoring Plan

**Status:** Planning  
**Scope:** File organization, type definitions, naming consistency  
**Strategy:** Independent chunks, can be done across multiple threads  
**Current Commit:** HEAD (baseline established)

---

## Overview

Phase 3 breaks into **6 independent refactoring chunks**. Each:
- Takes 1-2 sessions
- Has zero blocking dependencies
- Can be merged independently
- Includes rollback instructions if needed

**Total Effort:** ~12-16 hours spread over 6-8 threads  
**Risk:** Low (each chunk is localized, tested in isolation)  
**Interleaving:** Safe to mix with bug fixes, feature work, other PRs

---

## Chunk 1: Mode Metadata & Documentation (1 session)

**Status:** ✅ COMPLETED (Session 70)

**What:** Added `ModeMetadata` struct documenting all 12 app modes  
**Files:** `internal/app/modes.go`  
**Lines Added:** ~40  
**Breaking Changes:** None (backward-compatible)  
**Testing:** All modes render correctly

---

## Chunk 2: Error Handling Standardization (1 session)

**Status:** ✅ COMPLETED (Session 70)

**What:** Created `internal/app/errors.go` with `ErrorConfig` pattern  
**Files Modified:**
- `internal/app/errors.go` (NEW)
- `internal/app/app.go` (error handling)
- `internal/app/confirmationhandlers.go` (error logging)

**Lines Added:** ~140  
**Breaking Changes:** None (new infrastructure, old paths still work)  
**Testing:** Error paths trigger correctly, user-facing messages accurate

---

## Chunk 3: Message Organization (1-2 sessions)

**Status:** ✅ COMPLETED (Session 70)

**What:** Consolidate 11 message maps into 3-4 domain-scoped structs

**Current State:** `internal/app/messages.go`
```go
var InputPrompts map[string]string       // 11 entries
var InputHints map[string]string         // 11 entries
var ConfirmationTitles map[string]string // 10 entries
var ConfirmationExplanations map[string]string
var ConfirmationLabels map[string][2]string
// ... 6 more maps
```

**Target State:**
```go
var InputMessages map[string]InputMessage           // Combines Prompts + Hints
var ConfirmationMessages map[string]ConfirmationMessage  // Combines Title + Explanation + Labels
var ErrorMessages map[string]string  // Already standalone
// ... other domain groups
```

**Plan:**
1. Define struct types (`InputMessage`, `ConfirmationMessage`)
2. Create new maps with struct values
3. Add backwards-compatible facades (old map access still works)
4. Migrate all callers incrementally (or leave facades as permanent)
5. Test all message paths

**Files Modified:**
- `internal/app/messages.go` (restructure)
- `internal/ui/` (optional: use new API)
- `internal/app/confirmationhandlers.go` (optional: use new API)

**Lines Changed:** ~100 (net: +facades, -old maps)  
**Breaking Changes:** None (facades maintain old API)  
**Testing:** All confirmation dialogs, input prompts work correctly

**Why It Matters:**
- Clearer intent (Title + Explanation grouped = single concept)
- Easier to find related messages
- Compile-time validation (missing Title = error)

---

## Chunk 4: Confirmation Handler Pairing (1 session)

**Status:** ✅ COMPLETED (Session 70)

**What:** Replace 2 separate maps with 1 paired structure

**Current State:** `internal/app/confirmationhandlers.go`
```go
var confirmationActions = map[string]ConfirmationAction{ ... }
var confirmationRejectActions = map[string]ConfirmationAction{ ... }
```

**Target State:**
```go
var confirmationHandlers = map[string]ConfirmationActionPair{
    "action_id": {
        Confirm: ConfirmationAction{ ... },
        Reject: ConfirmationAction{ ... },
    },
}
```

**Plan:**
1. Define `ConfirmationActionPair` struct
2. Migrate all entries to paired map
3. Update handler lookup code (2-3 places)
4. Delete old maps
5. Test all confirmation paths

**Files Modified:**
- `internal/app/confirmationhandlers.go` (restructure)
- `internal/app/handlers.go` (update lookups)

**Lines Changed:** ~60 (net: -30 lines)  
**Breaking Changes:** None (internal refactor)  
**Testing:** All confirmations (init, clone, abort, etc.) work correctly

**Why It Matters:**
- Compiler ensures confirm/reject pairing (no orphaned handlers)
- Single place to modify confirmation logic
- Clearer code (pairing is explicit)

---

## Chunk 5: Type Definition Consolidation (1 session)

**Status:** ✅ COMPLETED (Session 71)

**What:** Move type definitions to canonical locations (SSOT)

**Current Scattered Types:**
- `OutputLineType` → `internal/ui/buffer.go:11-20` ✅ Already centralized
- `AppMode` → `internal/app/modes.go` ✅ Centralized
- `MenuItem` → `internal/app/menuitems.go` ✅ Centralized
- `GitState` → `internal/git/types.go` ✅ Centralized
- `ConfirmationAction` → `internal/app/confirmationhandlers.go:10-15` ✅ Local (OK)
- `InputMessage`, `ConfirmationMessage` → `internal/app/messages.go` (new, from Chunk 3)

**Assessment:** Types are already well-organized. No major moves needed.

**Remaining Opportunities:**
1. Add type aliases for clarity (e.g., `type FileInfoList = []ui.FileInfo`)
2. Document in ARCHITECTURE.md where each type lives
3. Add comments linking related types (e.g., `AppMode` ↔ `ModeMetadata`)

**Plan:**
1. Audit all type definitions (find any orphans)
2. Add ARCHITECTURE.md section "Type Definitions Location Map"
3. Add doc comments with references
4. No files move (already optimal)

**Effort:** ~1 session (mostly documentation)  
**Breaking Changes:** None  
**Testing:** Compile check only

**Why It Matters:**
- New contributors know where to find type definitions
- Prevents duplicate types in future
- Cross-references help understanding

---

## Chunk 6: Handler Function Naming Audit (1 session)

**Status:** Pending

**What:** Audit function naming against standards (handle*, execute*, cmd*, dispatch*)

**Current State:** Mixed naming patterns
```go
handleKeyCtrlC()              // ✅ handle* (key handler)
executeTimeTravelClean()      // ❌ should be cmdTimeTravelClean (returns tea.Cmd)
executeCloneWorkflow()        // ❌ should be cmdCloneWorkflow
dispatchInit()                // ✅ dispatch* (action router)
confirmConfirmNestedRepoInit() // ⚠️ awkward naming
```

**Plan:**
1. Grep all function definitions (`^func`)
2. Categorize by pattern:
   - `handle*` — Input handlers (key press, menu selection)
   - `cmd*` — Return `tea.Cmd` (async operations)
   - `dispatch*` — Route menu items
   - `execute*` — Run logic (no return value)
3. Identify violations (execute* that return tea.Cmd)
4. Rename in batches (5-10 renames per batch)
5. Test all renamed paths

**Files Modified:**
- `internal/app/operations.go` (2-3 renames)
- `internal/app/githandlers.go` (2-3 renames)
- Callers (update references)

**Effort:** ~1 session (systematic renaming)  
**Breaking Changes:** None (refactor only)  
**Testing:** Compile + smoke test

**Why It Matters:**
- Code matches documentation
- New contributors understand function intent
- Clearer API contracts

---

## Chunk 7: Directory Reorganization by Feature (2-3 sessions)

**Status:** Pending (largest chunk)

**What:** Restructure `internal/app/` and `internal/ui/` by feature domain

**Current Flat Structure:**
```
internal/app/
├── app.go
├── modes.go
├── handlers.go          (1200+ lines, mixed concerns)
├── githandlers.go       (600+ lines, mixed concerns)
├── operations.go        (800+ lines, mixed concerns)
├── menuitems.go
├── messages.go
├── confirmationhandlers.go
└── ...

internal/ui/
├── history.go
├── filehistory.go
├── conflictresolver.go
├── menu.go
├── console.go
├── ...
```

**Proposed Structure:**
```
internal/app/
├── app.go                          # Core Application
├── modes.go
├── menuitems.go
├── messages.go
├── keyboard.go                     # Input handling
├── errors.go
├── config/
│   ├── config.go                  # Config load/save
│   └── repo.go                    # Repo-specific config
├── init/                          # Init workflow
│   ├── handlers.go
│   ├── operations.go
│   └── messages.go
├── clone/                         # Clone workflow
│   ├── handlers.go
│   ├── operations.go
│   └── messages.go
├── timetraveling/                 # Time travel workflow
│   ├── handlers.go
│   ├── operations.go
│   ├── confirmations.go
│   └── messages.go
├── gitops/                        # Git operations (commit, push, pull, merge, etc.)
│   ├── commit.go
│   ├── push.go
│   ├── pull.go
│   ├── merge.go
│   └── stash.go
└── utils/
    ├── errors.go
    └── formatters.go

internal/ui/
├── theme.go
├── layout.go
├── menu/
│   ├── menu.go
│   ├── rendering.go
│   └── items.go
├── input/
│   ├── textinput.go
│   ├── branchinput.go
│   └── validation.go
├── console/
│   ├── console.go
│   ├── buffer.go
│   └── output.go
├── history/
│   ├── commit.go            (current history.go)
│   ├── file.go              (current filehistory.go)
│   └── rendering.go
├── conflict/
│   ├── resolver.go          (current conflictresolver.go)
│   └── rendering.go
└── utils/
    ├── box.go
    ├── formatting.go
    ├── sizing.go
    └── statusbar.go
```

**Phases:**
1. **Phase 7a:** Create directory structure, move files (no code changes)
2. **Phase 7b:** Update imports across codebase
3. **Phase 7c:** Verify all tests pass
4. **Phase 7d:** (Optional) Refactor large files (handlers.go → domain-specific)

**Effort:**
- Phase 7a: 1 session (mkdir, git mv)
- Phase 7b: 1-2 sessions (grep + replace imports)
- Phase 7c: 30 min (test + verify)
- Phase 7d: 2+ sessions (large refactor, defer)

**Breaking Changes:** None (same code, different location)  
**Testing:** Build + smoke test  
**Rollback:** `git revert` (atomic move commit)

**Why It Matters:**
- Related code grouped together
- Easier to navigate large packages
- Clearer responsibility boundaries
- Scales better as codebase grows

---

## Execution Order (Recommended)

**✅ Already Done:**
1. Chunk 1: Mode Metadata (Session 70)
2. Chunk 2: Error Handling Standardization (Session 70)
3. Chunk 3: Message Organization (Session 70)
4. Chunk 4: Confirmation Handler Pairing (Session 70)
5. Chunk 5: Type Definition Consolidation (Session 71)

**⏳ Next (Threads 72-75):**
6. **Thread 72:** Chunk 6 — Handler Function Naming Audit (1 session, systematic)
7. **Threads 73-75:** Chunk 7 — Directory Reorganization (3 sessions, incremental)

**Why This Order:**
- Chunks 5-6 are low-risk documentation/naming (safe to do early)
- Chunk 7 is largest, benefits from completed 1-4 (cleaner structure to move)
- Can pause between chunks without blocking features

---

## Rollback Instructions (Per Chunk)

Each chunk has a baseline commit. If needed:

```bash
# Find commit hash for chunk start
git log --oneline | grep "Phase 3 Chunk N"

# Rollback (keeps all code, reverts to pre-chunk state)
git revert --no-commit <commit-hash>
git commit -m "Revert Phase 3 Chunk N"

# Or hard reset if not pushed
git reset --hard <commit-hash>~1
```

---

## Success Criteria (Per Chunk)

| Chunk | Build | Tests | Integration | Notes |
|-------|-------|-------|-------------|-------|
| 1 | ✅ | ✅ | ✅ | Mode metadata documented |
| 2 | ✅ | ✅ | ✅ | Error paths tested |
| 3 | ✅ | ✅ | ✅ | All messages display correctly |
| 4 | ✅ | ✅ | ✅ | All confirmations work |
| 5 | ✅ | N/A | ✅ | Type location map in ARCHITECTURE.md |
| 6 | ✅ | ✅ | ✅ | Naming consistent, no orphans |
| 7 | ✅ | ✅ | ✅ | All imports updated, app runs |

---

## Interleaving with Other Work

**Safe to do in parallel:**
- Bug fixes (no overlap with reorganization)
- Feature work (in different packages)
- Testing (use latest code)

**Avoid during these chunks:**
- Large refactors of same files (wait for chunk to merge)
- Import path changes in dependent packages (coordinate timing)

---

## Documentation Updates (Per Chunk)

After each chunk:
1. Update `ARCHITECTURE.md` with new structure
2. Add migration notes to `SESSION-LOG.md`
3. Update `AGENTS.md` if patterns change

---

## Questions for User

1. **Chunk 7 scope:** Include "Phase 7d" refactoring large files (handlers.go split), or just move as-is?
2. **Backwards compat:** Keep facades for old APIs (Chunk 3), or hard-break?
3. **Parallel threads:** Can I start Chunk 5 while you work on other features?
4. **Testing strategy:** Manual testing only, or add automated tests?

---

**Ready to start Chunk 5, or prioritize differently?**
