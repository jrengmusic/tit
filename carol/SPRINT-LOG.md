# SPRINT-LOG.md

**Project:** tit  
**Repository:** /Users/jreng/Documents/Poems/dev/tit  
**Started:** 2026-01-31

**Purpose:** Long-term context memory across sessions. Tracks completed work, technical debt, and unresolved issues. Written by PRIMARY agents only when user explicitly requests.

---

## 📖 Notation Reference

**[N]** = Sprint Number (e.g., `1`, `2`, `3`...)

**Sprint:** A discrete unit of work completed by one or more agents, ending with user approval ("done", "good", "commit")

---

## ⚠️ CRITICAL RULES

**AGENTS BUILD CODE FOR USER TO TEST**
- Agents build/modify code ONLY when user explicitly requests
- USER tests and provides feedback
- Agents wait for user approval before proceeding

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

## Sprint 2: Manual Detached HEAD Handling ✅

**Date:** 2026-01-31

### Agents Participated
- **SURGEON:** Implemented manual detached HEAD support with OMP-style compact display

### Files Modified (11 total)
- `internal/git/types.go:1-50` — Added `ModifiedCount`, `IsTitTimeTravel` fields
- `internal/git/state.go:1-80` — Detached detection with flag, working tree count
- `internal/ui/header.go:1-60` — 2-row layout for detached state
- `internal/app/state_info.go:1-40` — OMP-style emojis (↑, ↓, ↕, ●)
- `internal/app/app_view.go:1-100` — ModifiedCount display, detached handling
- `internal/app/messages.go:1-30` — Compact descriptions
- `internal/app/workflow_state.go:1-50` — `ReturnToBranchName`, `ReturnToBranchDirtyTree`
- `internal/app/dispatchers.go:1-80` — Return workflow, branch picker
- `internal/app/handlers_config.go:1-40` — Branch picker Enter handling
- `internal/app/confirm_handlers.go:1-60` — Stash-based merge handler
- `internal/app/confirm_dialog.go:1-50` — `time_travel_return_dirty_choice` type

### Alignment Check
- [x] LIFESTAR principles followed (SSOT for detached detection)
- [x] NAMING-CONVENTION.md adhered
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (no layer violations)
- [x] No early returns used
- [x] Fail-fast error handling implemented

### Problems Solved
- Manual detached HEAD no longer shows fatal error
- OMP-style compact display for timeline (↑, ↓, ↕) and working tree (● N)
- Return to branch workflow with stash-based merge
- Branch picker for multiple branch options

### Technical Debt / Follow-up
- ARCHITECTURE.md needs updates for dual-mode detached handling
- SPEC.md needs updates (remove Section 13.5 fatal error, add manual detached workflow)

---

## Handoff to COUNSELOR

**From:** SURGEON  
**To:** COUNSELOR  
**Context:** Manual detached HEAD feature implementation complete

### What Was Built
Manual detached HEAD handling feature with:
- `IsTitTimeTravel` flag differentiates TIT-initiated vs manual detached
- `ModifiedCount` field for working tree changes
- OMP-style compact display (↑, ↓, ↕, ● N)
- Return to branch workflow with stash-based merge
- Branch picker for multiple branch options

### What COUNSELOR Needs To Do

**ARCHITECTURE.md Updates:**
1. Document `IsTitTimeTravel` flag in TimeTraveling section (line 256+)
2. Document dual-mode detached handling (TIT vs manual)
3. Document 2-row header layout for detached state
4. Document OMP-style display conventions

**SPEC.md Updates:**
1. **Section 10 Time Travel**: Extend for manual detached HEAD handling
2. **Section 3 State Model**: Add `IsTitTimeTravel` field documentation
3. **Section 6 State → Menu Mapping**: Add "Manual Detached" entry
4. **New section**: "Manual Detached HEAD" workflow (vs TIT time travel)
5. **Remove Section 13.5**: "Detached HEAD detected" fatal error - no longer applies

### Reference
Full implementation details above in this sprint entry.

---

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
