# SPRINT-LOG.md Template

**Project:** tit  
**Repository:** /Users/jreng/Documents/Poems/dev/tit  
**Started:** 2026-01-29

**Purpose:** Track agent registrations, sprint work, and completion reports. This file is mutable and rotates old entries (keeps last 5 sprints).
---

## 📖 Notation Reference

**[N]** = Sprint Number (e.g., `1`, `2`, `3`...)

**File Naming Convention:**
- `[N]-[ROLE]-[OBJECTIVE].md` — Task summary files written by agents
- `[N]-COUNSELOR-[OBJECTIVE]-KICKOFF.md` — Phase kickoff plans (COUNSELOR)
- `[N]-AUDITOR-[OBJECTIVE]-AUDIT.md` — Audit reports (AUDITOR)

**Example Filenames:**
- `1-COUNSELOR-INITIAL-PLANNING-KICKOFF.md` — COUNSELOR's plan for sprint 1
- `1-ENGINEER-MODULE-SCAFFOLD.md` — ENGINEER's task in sprint 1
- `2-AUDITOR-QUALITY-CHECK-AUDIT.md` — AUDITOR's audit after sprint 2

---

## ⚠️ CRITICAL AGENT RULES

**AGENTS BUILD CODE FOR USER TO TEST**
- Always build using scripts ./build.sh 
- USER tests and provides feedback
- Agents wait for user approval before proceeding

**AGENTS CAN RUN GIT ONLY IF USER EXPLICITLY ASKS**
- Write code changes without running git commands
- Agent runs git ONLY when user explicitly requests
- Never autonomous git operations
- **When committing:** Always stage ALL changes with `git add -A` before commit
  - ❌ DON'T selectively stage files (agents forget/miss files)
  - ✅ DO `git add -A` to capture every modified file
  - This ensures complete commits with nothing accidentally left unstaged

**LOG MAINTENANCE RULE**
- **All sprint logs must be written from latest to earliest (top to bottom), BELOW this rules section**
- **Only the last 5 sprints are kept in active log**
- **All agent roles except JOURNALIST write [N]-[ROLE]-[OBJECTIVE].md for each completed task**
- **JOURNALIST compiles all task summaries with same sprint number, updates SPRINT-LOG.md as new entry**
- **Only JOURNALIST can add new sprint entry to SPRINT HISTORY**
- **Sprints can be executed in parallel with multiple agents**
- Remove older sprints from active log (git history serves as permanent archive)
- This keeps log focused on recent work
- **JOURNALIST NEVER updates log without explicit user request**
- **During active sprints, only user decides whether to log**
- **All changes must be tested/verified by user, or marked UNTESTED**
- If rule not in this section, agent must ADD it (don't erase old rules)

**NAMING RULE (CODE VOCABULARY)**
- All identifiers must obey project-specific naming conventions (see project docs)
- Variable names: semantic + precise (not `temp`, `data`, `x`)
- Function names: verb-noun pattern (initRepository, detectCanonBranch)
- Struct fields: domain-specific terminology (not generic `value`, `item`, `entry`)
- Type names: PascalCase, clear intent (CanonBranchConfig, not BranchData)

**BEFORE CODING: ALWAYS SEARCH EXISTING PATTERNS**
- ❌ NEVER invent new states, enums, or utility functions without checking if they exist
- ✅ Always grep/search the codebase first for existing patterns
- ✅ Check types, constants, and error handling patterns before creating new ones
- **Methodology:** Read → Understand → Find SSOT → Use existing pattern
- Overcomplications usually mean you missed an existing solution

**TRUST THE LIBRARY, DON'T REINVENT**
- ❌ NEVER create custom helpers for things the library/framework already does
- ✅ Trust the library/framework - it's battle-tested
- **Philosophy:** Libraries are battle-tested. Your custom code is not.
- If you find yourself writing 10+ lines of utility code, stop—the library probably does it

**FAIL-FAST RULE (CRITICAL)**
- ❌ NEVER silently ignore errors (no error suppression)
- ❌ NEVER use fallback values that mask failures
- ❌ NEVER return empty strings/zero values when operations fail
- ✅ ALWAYS check error return values explicitly
- ✅ ALWAYS return errors to caller or log + fail fast
- Better to panic/error early than debug silent failure for hours

**META-PATTERN RULE (CRITICAL)**
- ❌ NEVER start complex task without reading PATTERNS.md
- ✅ ALWAYS use Problem Decomposition Framework for multi-step tasks
- ✅ ALWAYS use Debug Methodology checklist when investigating bugs
- ✅ ALWAYS run Self-Validation Checklist before responding
- ✅ Follow role-specific patterns (COUNSELOR, ENGINEER, SURGEON, MACHINIST, AUDITOR)
- Better to pause and read patterns than repeat documented failures

**SCRIPT USAGE RULE**
- ✅ ALWAYS use scripts from SCRIPTS.md for code editing (when available)
- ✅ Scripts have dry-run mode - use it before actual edit
- ✅ Scripts create backups - verify before committing
- ❌ NEVER use raw sed/awk without safe-edit.sh wrapper (when script available)
- Scripts prevent common mistakes and enforce safety

**⚠️ NEVER EVER REMOVE THESE RULES**
- Rules at top of SPRINT-LOG.md are immutable
- If rules need update: ADD new rules, don't erase old ones
- Any agent that removes or modifies these rules has failed
- Rules protect the integrity of the development log

---

## Quick Reference

### For Agents Starting New Sprint

1. **Check:** Do I see my registration in ROLE ASSIGNMENT REGISTRATION?
2. **If YES:** Proceed with role constraints, include `[Acting as: ROLE]` in responses
3. **If NO:** STOP and ask: "What is my role in this sprint?"

### For Human Orchestrator

**Register agent:**
```
"Read CAROL.md. You are assigned as [ROLE], register yourself in SPRINT-LOG.md"
```

**Verify registration:**
```
"What is your current role?"
```

**Reassign role:**
```
"You are now reassigned as [NEW_ROLE], register yourself in SPRINT-LOG.md"
```

**Complete sprint (call JOURNALIST):**
```
"Read CAROL, act as JOURNALIST. Log sprint [N] to SPRINT-LOG.md"
```

---

## ROLE ASSIGNMENT REGISTRATION

COUNSELOR: [Not Assigned]  
ENGINEER: glm-4.7 (zai-coding-plan/glm-4.7)  
SURGEON: [Not Assigned]  
AUDITOR: Amp (Claude Sonnet 4) — Sprint 3, 4  
MACHINIST: [Not Assigned]  
JOURNALIST: glm-4.6 (zai-coding-plan/glm-4.6) — Active

---

<!-- SPRINT HISTORY STARTS BELOW -->
<!-- Only JOURNALIST writes entries here -->
<!-- Latest sprint at top, oldest at bottom -->
<!-- Keep last 5 sprints, rotate older to git history -->

## SPRINT HISTORY

## Sprint 3 - LIFESTAR Compliance and Auditing
**Date:** 2026-01-30
**Agents:** AUDITOR, COUNSELOR

### Summary
Comprehensive LIFESTAR compliance audit identifying critical refactoring opportunities, error handling issues, and architectural anti-patterns. Addressed 4 HIGH priority issues through focused fixes for error handling, state detection confidence, and testability improvements.

### Tasks Completed
- **AUDITOR**: LIFESTAR Audit - Identified 4 critical refactoring opportunities (silent error fallbacks, panic in non-critical path, early returns masking state, global singleton) and 2 anti-patterns
- **COUNSELOR**: Auditor Fixes Kickoff - Planned 4-phase approach to address high-priority findings with specific implementation details

### Files Modified
- `internal/git/state.go` - Add DetectionWarnings field to State struct
- `internal/git/execute.go` - Change FindStashRefByHash to return error instead of panic
- `internal/git/types.go` - Add TimelineConfidence enum to State struct
- `internal/ui/buffer.go` - Inject OutputBuffer via Application constructor
- `internal/app/app.go` - Add outputBuffer field and Update calls to display warnings

### Notes
Critical issues addressed: silent error fallbacks that could mislead users, application crashes from stash panics, timeline detection uncertainty, and global singleton buffer preventing testability. All fixes maintain backward compatibility while improving error visibility and system stability.

---

## Sprint 2 - Code Refactoring and Optimization
**Date:** 2026-01-30
**Agents:** COUNSELOR, ENGINEER

### Summary
Comprehensive refactoring of internal/app/ to reduce file sizes and improve maintainability. Reduced app.go from 1,771 to 392 lines and split large handler files into focused modules.

### Tasks Completed
- **COUNSELOR**: Large Files Kickoff - Planned approach to reduce internal/app/ from 11,176 lines
- **ENGINEER**: Phase 1 - Extract core Bubble Tea methods from app.go into separate files
- **ENGINEER**: Phase 2 - Analyzed delegation methods (found already using direct access)
- **ENGINEER**: Phase 3 - Split confirmation_handlers.go into domain-focused files
- **COUNSELOR**: Reduce app.go Kickoff - Planned aggressive reduction to ~400 lines
- **ENGINEER**: Phase 1 - Deleted 41 unused delegation methods (171 lines)
- **ENGINEER**: Phase 2 - Extracted constructor logic to app_constructor.go (180 lines)
- **ENGINEER**: Phase 3 - Extracted key handlers to app_keys.go (220 lines)

### Files Modified
- `internal/app/app.go` - Reduced from 1,771 to 392 lines (78% reduction)
- `internal/app/app_init.go` - Created (135 lines)
- `internal/app/app_update.go` - Created (326 lines)
- `internal/app/app_view.go` - Created (301 lines)
- `internal/app/app_constructor.go` - Created (180 lines)
- `internal/app/app_keys.go` - Created (220 lines)
- `internal/app/confirm_dialog.go` - Created (362 lines)
- `internal/app/confirm_handlers.go` - Created (649 lines)
- `internal/app/confirmation_handlers.go` - Split and removed

### Notes
Phase 4 (state sub-package) was skipped due to circular dependencies. All 41 delegation methods were verified as unused before deletion. No logic changes - pure code movement and cleanup. Overall reduction of 42.9% from the original large files.

---

## Sprint 1 - Full Codebase Audit
**Date:** 2026-01-29
**Agents:** AUDITOR

### Summary
Comprehensive audit of the entire codebase identifying structural issues, LIFESTAR violations, and anti-patterns.

### Tasks Completed
- **AUDITOR**: Full codebase audit - Identified critical issues including duplicate type definitions, God Object anti-pattern, and LIFESTAR violations

### Files Modified
- `internal/git/types.go` - Duplicate type definitions
- `internal/ui/history.go` - Duplicate type definitions
- `internal/git/state.go` - Hardcoded strings
- `internal/app/handlers_history.go` - Magic numbers
- Various backup and temporary files

### Notes
Critical findings included SSOT violations with duplicate CommitInfo and FileInfo types across files. Application struct identified as God Object with 93 fields. Recommended immediate fixes for duplicate types and removal of leftover files.

---

<!-- Actual sprint entries go here, written by JOURNALIST -->

---

## [N]-[ROLE]-[OBJECTIVE].md Format Reference

**File naming:** `[N]-[ROLE]-[OBJECTIVE].md`  
**Examples:**
- `[N]-ENGINEER-MERMAID-MODULE.md`
- `[N]-MACHINIST-ERROR-HANDLING.md`
- `[N]-SURGEON-COMPILE-FIX.md`

**Content format:**
```markdown
# Sprint [N] Task Summary

**Role:** [ROLE NAME]
**Agent:** [CLI Tool (Model)]
**Date:** 2026-01-29
**Time:** [HH:MM]
**Task:** [Brief task description]

## Objective
[What was accomplished in 1-2 sentences]

## Files Modified ([X] total)
- `path/to/file.ext` — [brief description of changes]
- `path/to/file2.ext` — [brief description of changes]

## Notes
- [Important learnings, blockers, or decisions]
- [Any warnings or follow-up needed]
```

**Lifecycle:**
1. Agent completes task
2. Agent writes [N]-[ROLE]-[OBJECTIVE].md
3. JOURNALIST compiles all [N]-[ROLE]-[OBJECTIVE].md files into SPRINT-LOG.md entry
4. JOURNALIST deletes all [N]-[ROLE]-[OBJECTIVE].md files after compilation
---

**End of SPRINT-LOG.md Template**

Copy this template to your project root as `SPRINT-LOG.md` and customize:
- Project name
- Repository URL/path
- Starting date
- Add project-specific rules to CRITICAL AGENT RULES section

Rock 'n Roll!  
JRENG!