# SPRINT-LOG.md

**Project:** tit  
**Repository:** /Users/jreng/Documents/Poems/dev/tit  
**Started:** 2026-03-01

**Purpose:** Long-term context memory across sessions. Tracks completed work, technical debt, and unresolved issues. Written by PRIMARY agents only when ARCHITECT explicitly requests.

---

## 📖 Notation Reference

**[N]** = Sprint Number (e.g., `1`, `2`, `3`...)

**Sprint:** A discrete unit of work completed by one or more agents, ending with ARCHITECT approval ("done", "good", "commit")

---

## ⚠️ CRITICAL RULES

**AGENTS BUILD CODE FOR ARCHITECT TO TEST**
- Agents build/modify code ONLY when ARCHITECT explicitly requests
- ARCHITECT tests and provides feedback
- Agents wait for ARCHITECT approval before proceeding

**AGENTS NEVER RUN GIT COMMANDS**
- Write code changes without running git commands
- Agent runs git ONLY when user explicitly requests
- Never autonomous git operations
- **When committing:** Always stage ALL changes with `git add -A` before commit
  - ❌ DON'T selectively stage files (agents forget/miss files)
  - ✅ DO `git add -A` to capture every modified file

**SPRINT-LOG WRITTEN BY PRIMARY AGENTS ONLY**
- **COUNSELOR** or **SURGEON** write to SPRINT-LOG
- Only when user explicitly says: `"log sprint"`
- No intermediate summary files
- No automatic logging after every task
- Latest sprint at top, keep last 5 entries

**NAMING RULE (CODE VOCABULARY)**
- All identifiers must obey project-specific naming conventions (see NAMING-CONVENTION.md)
- Variable names: semantic + precise (not `temp`, `data`, `x`)
- Function names: verb-noun pattern (initRepository, detectCanonBranch)
- Struct fields: domain-specific terminology (not generic `value`, `item`, `entry`)
- Type names: PascalCase, clear intent (CanonBranchConfig, not BranchData)

**BEFORE CODING: ALWAYS SEARCH EXISTING PATTERNS**
- ❌ NEVER invent new states, enums, or utility functions without checking if they exist
- ✅ Always grep/search the codebase first for existing patterns
- ✅ Check types, constants, and error handling patterns before creating new ones
- **Methodology:** Read → Understand → Find SSOT → Use existing pattern

**TRUST THE LIBRARY, DON'T REINVENT**
- ❌ NEVER create custom helpers for things the library/framework already does
- ✅ Trust the library/framework - it's battle-tested

**FAIL-FAST RULE (CRITICAL)**
- ❌ NEVER silently ignore errors (no error suppression)
- ❌ NEVER use fallback values that mask failures
- ❌ NEVER return empty strings/zero values when operations fail
- ❌ NEVER use early returns
- ✅ ALWAYS check error returns explicitly
- ✅ ALWAYS return errors to caller or log + fail fast

**⚠️ NEVER REMOVE THESE RULES**
- Rules at top of SPRINT-LOG.md are immutable
- If rules need update: ADD new rules, don't erase old ones

---

## Quick Reference

### For Agents

**When user says:** `"log sprint"`

1. **Check:** Did I (PRIMARY agent) complete work this session?
2. **If YES:** Write sprint block to SPRINT-LOG.md (latest first)
3. **Include:** Files modified, changes made, alignment check, technical debt

### For User

**Activate PRIMARY:**
```
"@CAROL.md COUNSELOR: Rock 'n Roll"
"@CAROL.md SURGEON: Rock 'n Roll"
```

**Log completed work:**
```
"log sprint"
```

**Invoke subagent:**
```
"@oracle analyze this"
"@engineer scaffold that"
"@auditor verify this"
```

**Available Agents:**
- **PRIMARY:** COUNSELOR (domain specific strategic analysis), SURGEON (surgical precision problem solving)
- **Subagents:** Pathfinder, Oracle, Engineer, Auditor, Machinist, Librarian

---

<!-- SPRINT HISTORY STARTS BELOW -->
<!-- Latest sprint at top, oldest at bottom -->
<!-- Keep last 5 sprints, rotate older to git history -->

## SPRINT HISTORY

<!-- Example sprint entry (delete this after first real sprint) -->

## Sprint 1: Project Setup and Initial Planning ✅

**Date:** 2026-01-11  
**Duration:** 14:00 - 16:30 (2.5 hours)

### Agents Participated
- **COUNSELOR:** Kimi-K2 — Wrote SPEC.md and ARCHITECTURE.md
- **ENGINEER** (invoked by COUNSELOR) — Created project structure
- **AUDITOR** (invoked by COUNSELOR) — Verified spec compliance

### Files Modified (8 total)
- `SPEC.md:1-200` — Complete feature specification with all flows
- `ARCHITECTURE.md:1-150` — Initial architecture patterns documented
- `src/core/module.cpp:10-45` — Core module scaffolding with proper initialization
- `src/core/module.h:1-30` — Core module header with explicit dependencies
- `tests/core_test.cpp:1-50` — Test scaffolding following Testable principle
- `CMakeLists.txt:1-25` — Build configuration with explicit targets
- `README.md:1-20` — Project overview

### Alignment Check
- [x] LIFESTAR principles followed (Lean, Immutable, Findable, Explicit, SSOT, Testable, Accessible, Reviewable)
- [x] NAMING-CONVENTION.md adhered (semantic names, verb-noun functions, no type encoding)
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (no layer violations, explicit dependencies)
- [x] No early returns used
- [x] Fail-fast error handling implemented

### Problems Solved
- Established project foundation following domain-specific patterns
- Defined clear module boundaries preventing layer violations

### Technical Debt / Follow-up
- Error handling needs refinement in module.cpp (marked with TODO)
- Performance requirements not yet defined for real-time constraints

**Status:** ✅ APPROVED - All files compile, tests scaffold in place

---

## Sprint 3: Replace-Last-Line for Git Progress ✅

**Date:** 2026-03-24
**Duration:** ~20 min

### Agents Participated
- **COUNSELOR** — Requirements, delegation
- **Engineer** — Implementation

### Files Modified (4 total)
- `internal/ui/buffer.go:49-68` — Added ReplaceLast() method to OutputBuffer
- `internal/git/types.go:114-118` — Extended Logger interface with LogReplace, ErrorReplace
- `internal/git/types.go:143-157` — Added package-level LogReplace(), ErrorReplace() functions
- `internal/app/git_logger.go` — Implemented LogReplace/ErrorReplace on GitLogger (calls ReplaceLast)
- `internal/git/execute.go:139-207` — Both streaming goroutines now track isProgressLine; \r triggers replace, \n triggers append

### Alignment Check
- [x] LIFESTAR principles followed (Lean: minimal addition, SSOT: Logger interface is single contract, Explicit Encapsulation: git signals intent, UI handles rendering)
- [x] NAMING-CONVENTION.md adhered (ReplaceLast, LogReplace, ErrorReplace, isProgressLine)
- [x] ARCHITECTURAL-MANIFESTO.md principles applied
- [x] No early returns
- [x] Fail-fast error handling

### Problems Solved
- Git progress output (Receiving objects: X%) now updates in-place like bare terminal instead of spamming 100 lines

### Technical Debt / Follow-up
- None

**Status:** ✅ APPROVED — Build clean

---

## Sprint 2: Transparent LFS + Console UX Fixes ✅

**Date:** 2026-03-24
**Duration:** ~2 hours

### Agents Participated
- **COUNSELOR** — Requirements analysis, plan, contract alignment, delegation
- **Pathfinder** — Codebase discovery (state model, execute patterns, header rendering, cache flow, ESC handlers)
- **Researcher** — Git LFS integration patterns, edge cases, failure modes
- **Librarian** — Git progress.c source analysis, pipe buffering behavior
- **Engineer** — Implementation (6 tasks LFS, --progress flag, cache defer)
- **Auditor** — Contract compliance audit (found 5 critical, 3 high, 3 medium issues)

### Files Modified (16 total)

**LFS Feature:**
- `internal/git/lfs.go:1-48` — NEW: IsRepoLFS(), IsLFSInstalled(), IsLFSBinaryAvailable(), SetupLFSFilters(), FetchLFSObjects(), CheckoutLFSObjects()
- `internal/git/types.go:83-84` — Added LFS, LFSReady bool fields to State struct
- `internal/git/types.go:130-134` — Exported warn() to Warn() for git package logging
- `internal/git/state.go:47-51` — LFS detection in DetectState() after isRepo check
- `internal/ui/header.go:30-31` — Added LFSLabel, LFSColor to HeaderState
- `internal/ui/header.go:125-138` — LFS badge rendering next to version (independent styled pieces)
- `internal/app/app_view_header.go:119-130` — LFS indicator population from State
- `internal/app/app_constructor.go` — LFS auto-setup at startup (binary check + filter install)
- `internal/app/environment_state.go` — Clean (lfsChecked removed after audit)
- `internal/app/op_remote.go:49-77` — cmdFetchRemote chains git lfs fetch when LFS
- `internal/app/op_pull.go:91-98` — cmdHardReset chains git lfs checkout when LFS

**--progress Flag (9 files):**
- `internal/app/op_clone.go` — clone --progress
- `internal/app/op_pull.go` — pull --progress, fetch --progress
- `internal/app/op_push.go` — push --progress, force push --progress
- `internal/app/op_commit.go` — commit+push --progress
- `internal/app/op_remote.go` — fetch --all --progress
- `internal/app/op_dirty_pull_merge.go` — dirty pull --progress
- `internal/app/op_push_sync.go` — push sync --progress
- `internal/app/handlers_git_pull.go` — pull merge/rebase --progress
- `internal/app/handlers_git_workflow.go` — push workflow --progress

**Cache Defer:**
- `internal/app/handlers_commit.go` — Removed invalidateHistoryCaches(), shows completion immediately
- `internal/app/handlers_pull.go` — Removed cache invalidation from branch switch handlers
- `internal/app/handlers_timetravel.go` — Removed cache invalidation from 3 time travel handlers
- `internal/app/app_update_cmd.go` — Removed console completion message from handleCacheProgress
- `internal/app/history_cache.go` — Removed all buffer.Append() from cache goroutines (silent build)
- `internal/app/handlers_global_menu.go` — Added invalidateHistoryCaches() to ESC ModeMenu path, time-travel ESC path, and returnToMenu()

**Version:**
- `internal/constants.go:12` — v1.2.0 → v1.3.0

### Alignment Check
- [x] LIFESTAR principles followed
- [x] NAMING-CONVENTION.md adhered (IsRepoLFS, IsLFSInstalled, IsLFSBinaryAvailable)
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (Lean: no new axes/modes, SSOT: LFS state in DetectState, Explicit Encapsulation: lfs.go knows nothing about UI)
- [x] No early returns (audit caught and fixed)
- [x] Fail-fast error handling

### Problems Solved
- **LFS transparency:** Repos with LFS now auto-detect, auto-setup filters, and include LFS objects in fetch/reset. Header shows LFS status.
- **Stale console:** git suppresses progress to pipes without --progress flag. Added flag to all 16 streaming git commands.
- **Cache noise:** Cache goroutines raced with operation output in console. Moved cache rebuild to ESC-return-to-menu path. Cache builds silently.

### Technical Debt / Follow-up
- `op_pull.go:55` uses raw `exec.Command` instead of `git.Execute()` (pre-existing, not from this sprint)
- `handlers_remote.go` has pre-existing early returns in handleFetchRemote
- LFS management UI (track/untrack/prune/locks) not in scope — future feature if needed
- DirtyOperation state path skips LFS detection (state.LFS=false) — acceptable since DirtyOperation blocks all ops

**Status:** ✅ APPROVED — Build clean

---

<!-- Actual sprint entries go here, written by PRIMARY agents -->

---

**End of SPRINT-LOG.md Template**

Copy this template to your project root as `SPRINT-LOG.md` and customize:
- Project name
- Repository URL/path
- Starting date
- Add project-specific rules to CRITICAL RULES section

Rock 'n Roll!  
**JRENG!**
