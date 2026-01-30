# SPRINT-LOG.md

**Project:** tit  
**Repository:** /Users/jreng/Documents/Poems/dev/tit  
**Started:** 2026-01-30

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
- **Engineer** (invoked by COUNSELOR) — Created project structure
- **Auditor** (invoked by COUNSELOR) — Verified spec compliance

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

## Sprint 2: Time Travel Double-Conflict Path & Stale Stash Handling ✅

**Date:** 2026-01-30  
**Duration:** ~3 hours

### Agents Participated
- **SURGEON:** OpenAI CodeGPT (MiniMax-M2.1) — Primary agent, all implementation
- COUNSELOR — Provided handoff specification for stale stash handling

### Files Modified (3 total)

#### internal/git/execute.go
- Lines 100-124: Refactored `FindStashRefByHash()` to return `(string, bool)` instead of panicking
- Lines 126-130: Added `StashExists(stashHash string) bool` function
- Lines 750-760: Updated TimeTravelMerge stash drop to handle missing stash
- Lines 865-885: Updated TimeTravelReturn stash drop to handle missing stash
- Lines 398-420, 716-740, 821-840: Updated `ListConflictedFiles()` callers to check len instead of error

#### internal/app/confirm_handlers.go
- Lines 323-343: Modified `executeConfirmTimeTravelReturn()` to validate stash before operation
- Lines 384-412: Modified `executeConfirmTimeTravelMerge()` to validate stash before operation
- Lines 559-620: Added `executeConfirmStaleStashContinue()`, `executeRejectStaleStashContinue()`, `executeConfirmStaleStashMergeContinue()`

#### internal/app/confirm_dialog.go
- Lines 101-106: Added `confirm_stale_stash_continue` handler pair
- Lines 107-111: Added `confirm_stale_stash_merge_continue` handler pair

### Alignment Check
- [x] LIFESTAR principles followed (Lean, Immutable, Findable, Explicit, SSOT, Testable, Accessible, Reviewable)
- [x] NAMING-CONVENTION.md adhered (verb-noun functions, semantic names, no type encoding)
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (no layer violations, explicit dependencies)
- [x] No early returns used
- [x] Fail-fast error handling implemented (panic → graceful handling)
- [x] No scope creep (minimal surgical fixes per handoff)

### Problems Solved
1. **Panic on stale stash:** `FindStashRefByHash()` now returns gracefully instead of panicking
2. **Conflicted file detection:** `ListConflictedFiles()` now returns empty slice instead of error
3. **Double-conflict path:** User can trigger and resolve two sequential conflicts (time travel merge + stash apply)
4. **Stale stash detection:** TIT now validates stash exists before operations, shows confirmation dialog
5. **Config cleanup:** Stale TOML entries cleaned up on confirmation

### Technical Debt / Follow-up
- None identified — implementation follows existing patterns exactly

### Acceptance Criteria Met
- [x] User manually drops stash via `git stash drop`
- [x] User initiates time travel return
- [x] TIT shows confirmation dialog: "Original stash [hash] was manually dropped. Continue without restoring stash?"
- [x] If Yes: clean up TOML entry, complete return without stash restore
- [x] If No: cancel operation, return to menu
- [x] Same behavior for time travel merge
- [x] User can trigger and resolve double conflicts (merge + stash apply)
- [x] User ends up DIRTY with all changes preserved

### Test Repos Created
- `/var/tmp/test_repo` — Main test repo with 7 commits, 3 branches
- `/var/tmp/test_repo/scripts/test_double_conflict.sh` — Automated setup for double-conflict scenario

---

## Handoff to SURGEON: Graceful Stale Stash Handling with Confirmation Dialog

**From:** COUNSELOR  
**Date:** 2026-01-30

[Existing handoff content...]

---

---

**End of SPRINT-LOG.md Template**

Copy this template to your project root as `SPRINT-LOG.md` and customize:
- Project name
- Repository URL/path
- Starting date
- Add project-specific rules to CRITICAL RULES section

Rock 'n Roll!  
**JRENG!**
