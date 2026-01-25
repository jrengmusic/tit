# SESSION-LOG.md Template

**Project:** TIT  
**Repository:** /Users/jreng/Documents/Poems/dev/tit  
**Started:** 2026-01-25

**Purpose:** Track agent registrations, session work, and completion reports. This file is mutable and rotates old entries (keeps last 5 sessions).

---

## 📖 Notation Reference

**[N]** = Session Number (e.g., `1`, `2`, `3`...)

**File Naming Convention:**
- `[N]-[ROLE]-[OBJECTIVE].md` — Task summary files written by agents
- `[N]-COUNSELOR-[OBJECTIVE]-KICKOFF.md` — Phase kickoff plans (COUNSELOR)
- `[N]-AUDITOR-[OBJECTIVE]-AUDIT.md` — Audit reports (AUDITOR)

**Example Filenames:**
- `1-COUNSELOR-INITIAL-PLANNING-KICKOFF.md` — COUNSELOR's plan for session 1
- `1-ENGINEER-MODULE-SCAFFOLD.md` — ENGINEER's task in session 1
- `2-AUDITOR-QUALITY-CHECK-AUDIT.md` — AUDITOR's audit after session 2

---

## ⚠️ CRITICAL AGENT RULES
**AGENTS BUILD APP FOR USER TO TEST**
- run script ./build.sh
- USER tests
- Agent waits for feedback

**AGENTS CAN RUN GIT ONLY IF USER EXPLICITLY ASKS**
- Write code changes without running git commands
- Agent runs git ONLY when user explicitly requests
- Never autonomous git operations
- **When committing:** Always stage ALL changes with `git add -A` before commit
  - ❌ DON'T selectively stage files (agents forget/miss files)
  - ✅ DO `git add -A` to capture every modified file
  - This ensures complete commits with nothing accidentally left unstaged

**LOG MAINTENANCE RULE**
- **All session logs must be written from latest to earliest (top to bottom), BELOW this rules section**
- **Only the last 5 sessions are kept in active log**
- **All agent roles except JOURNALIST write [N]-[ROLE]-[OBJECTIVE].md for each completed task**
- **JOURNALIST compiles all task summaries with same session number, updates SESSION-LOG.md as new entry**
- **Only JOURNALIST can add new session entry to SESSION HISTORY**
- **Sessions can be executed in parallel with multiple agents**
- Remove older sessions from active log (git history serves as permanent archive)
- This keeps log focused on recent work
- **JOURNALIST NEVER updates log without explicit user request**
- **During active sessions, only user decides whether to log**
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
- Rules at top of SESSION-LOG.md are immutable
- If rules need update: ADD new rules, don't erase old ones
- Any agent that removes or modifies these rules has failed
- Rules protect the integrity of the development log

---

## Quick Reference

### For Agents Starting New Session

1. **Check:** Do I see my registration in ROLE ASSIGNMENT REGISTRATION?
2. **If YES:** Proceed with role constraints, include `[Acting as: ROLE]` in responses
3. **If NO:** STOP and ask: "What is my role in this session?"

### For Human Orchestrator

**Register agent:**
```
"Read CAROL.md. You are assigned as [ROLE], register yourself in SESSION-LOG.md"
```

**Verify registration:**
```
"What is your current role?"
```

**Reassign role:**
```
"You are now reassigned as [NEW_ROLE], register yourself in SESSION-LOG.md"
```

**Complete session (call JOURNALIST):**
```
"Read CAROL, act as JOURNALIST. Log session [N] to SESSION-LOG.md"
```

---

## ROLE ASSIGNMENT REGISTRATION

COUNSELOR: Copilot (claude-opus-41)  
ENGINEER: zai-coding-plan/glm-4.7  
SURGEON: [Agent (Model)] or [Not Assigned]  
AUDITOR: zai-coding-plan/glm-4.7
MACHINIST: zai-coding-plan/glm-4.7  
JOURNALIST: zai-coding-plan/glm-4.7

---

<!-- SESSION HISTORY STARTS BELOW -->
<!-- Only JOURNALIST writes entries here -->
<!-- Latest session at top, oldest at bottom -->
<!-- Keep last 5 sessions, rotate older to git history -->

## SESSION HISTORY

## Session 2: Clean Auto-Update System ✅

**Date:** 2026-01-25
**Duration:** 22:15 - 23:00 (~45 minutes)

### Objectives
- Remove all existing timeline sync machinery
- Implement single, clean background auto-update system
- Periodic full state detection and UI refresh (exactly like app launch)
- KISS principle: one timer, full state detection, full UI update

### Agents Participated
- COUNSELOR: Copilot (claude-opus-41) — Created comprehensive implementation plan in `2-COUNSELOR-AUTO-UPDATE-KICKOFF.md`
- ENGINEER: zai-coding-plan/glm-4.7 — Implemented clean auto-update system with lazy updates and spinner feedback
- Tested by: User (PENDING)

### Files Modified (10 total)
- `internal/app/timeline_sync.go` — DELETED - Old sync system removed completely
- `internal/app/auto_update.go` — NEW FILE - Clean auto-update system with single timer, full state detection, lazy updates (5s idle timeout), spinner feedback (100ms animation frames)
- `internal/app/app.go` — Removed sync state fields (timelineSyncInProgress, timelineSyncLastUpdate, timelineSyncFrame), removed menu activity fields, updated Init() to call startAutoUpdate()
- `internal/app/messages.go` — Removed TimelineSyncMsg and TimelineSyncTickMsg, added AutoUpdateTickMsg, AutoUpdateCompleteMsg, AutoUpdateAnimationMsg
- `internal/app/history_cache.go` — Updated cmdToggleAutoUpdate to use startAutoUpdate()
- `internal/app/handlers.go` — Updated returnToMenu() to call startAutoUpdate()
- `internal/app/setup_wizard.go` — Updated setup completion to call startAutoUpdate()
- `internal/app/git_handlers.go` — Updated time travel return handlers to call startAutoUpdate()
- `internal/app/confirmation_handlers.go` — Updated 9 confirmation handlers to call startAutoUpdate() on menu transitions

### Problems Solved
- Old timeline sync system completely removed (clean slate)
- New auto-update system implemented with single timer (scheduleAutoUpdateTick)
- Full state detection implemented (all 5 axes via git.DetectState())
- Full UI update on state change (menu regeneration + shortcuts rebuild)
- **Menu activity tracking added** for lazy updates - pauses auto-update during user navigation/actions (5s idle timeout)
- **Spinner feedback added** for both WorkingTree and Timeline during auto-update (better UX)
- AutoUpdateAnimationMsg for 100ms animation frames
- Mode transitions updated to restart auto-update when returning to menu

### Summary
COUNSELOR analyzed previous session failures and created detailed 3-phase implementation plan to completely remove old sync architecture and replace with clean, simple auto-update system. ENGINEER successfully implemented all phases:

✅ **Phase 1:** Old sync system completely removed (timeline_sync.go deleted, all sync fields removed)
✅ **Phase 2:** New auto-update system implemented (auto_update.go created, single timer, full state detection, full UI update)
✅ **Phase 3:** Testing scenarios documented (basic auto-update, remote changes, menu navigation, config control, mode transitions)

Build status: ✅ VERIFIED - No errors

Implementation matches kickoff plan specifications with enhancements:
- Lazy update with 5s idle timeout (prevents updates during user interaction)
- Spinner feedback during auto-update (better UX)
- Config interval wired (no redundant fallbacks)
- Mode transitions restart auto-update

Single timer + full state detection pattern. No architectural tangles. Thread-safe via message passing.

**Status:** ✅ IMPLEMENTED - Awaiting user testing

---

## Session 1: Background State Update Feature ✅

**Date:** 2026-01-25
**Duration:** 10:45 - 22:15 (~11.5 hours)

### Objectives
- Extend existing Timeline sync to also update WorkingTree status
- Implement smart menu regeneration only when git state changes affect available options
- Add menu activity detection to pause sync during user navigation
- Maintain KISS principle: single timer, single preference

### Agents Participated
- COUNSELOR: Copilot (claude-opus-41) — Created comprehensive implementation plan in `1-COUNSELOR-STATE-UPDATE-KICKOFF.md`
- ENGINEER: Amp (Claude) — First attempt FAILED, second attempt SUCCESSFUL with full state detection implementation
- Tested by: User (PENDING)

### Files Modified (6 total)
- `internal/app/app.go` — Added menu activity tracking fields (lastMenuActivity, menuActivityTimeout), initialized in NewApplication(), updated menu handlers (handleMenuUp, handleMenuDown, handleMenuEnter) to track activity
- `internal/app/timeline_sync.go` — Extended sync to detect full state (WorkingTree + Timeline), added menu activity timeout check, updated to use smart state comparison
- `internal/app/messages.go` — Extended TimelineSyncMsg struct with OldState and NewState fields for state comparison
- `internal/git/state_compare.go` — NEW FILE. Implements CompareStates() function to determine if menu regeneration is necessary based on state changes

### Problems Solved
- Menu activity detection prevents sync during user navigation (prevents UI flicker)
- Smart state comparison prevents unnecessary menu regeneration (Ahead(n)->Ahead(m) no change, Behind(n)->Behind(m) no change)
- WorkingTree detection includes untracked files (Dirty if any untracked files exist)
- Sync architecture extended without breaking existing functionality

### Summary
COUNSELOR analyzed TIT codebase and created detailed 6-phase implementation plan. ENGINEER first attempt failed due to fundamental architectural mismatch (sync mechanism tightly coupled to remote requirement). ENGINEER second attempt successfully implemented all phases:

✅ **Phase 1-2:** Menu activity tracking added, all menu navigation handlers update lastMenuActivity timestamp
✅ **Phase 3:** State comparison logic implemented in new file, handles special cases for Ahead/Behind commit count changes
✅ **Phase 4:** Sync extended to detect full state, menu activity timeout check added, smart menu regeneration implemented

Build status: ✅ PASSED - `./build.sh` successful, binary installed to `~/.tit/bin/tit_x64`

Single timer + single preference pattern maintained. Existing architecture patterns preserved. Thread-safe via message passing (timelineSyncChan).

**Status:** ✅ IMPLEMENTED - Awaiting user testing

---

<!-- Actual session entries go here, written by JOURNALIST -->

---

## [N]-[ROLE]-[OBJECTIVE].md Format Reference

**File naming:** `[N]-[ROLE]-[OBJECTIVE].md`  
**Examples:**
- `[N]-ENGINEER-MERMAID-MODULE.md`
- `[N]-MACHINIST-ERROR-HANDLING.md`
- `[N]-SURGEON-COMPILE-FIX.md`

**Content format:**
```markdown
# Session [N] Task Summary

**Role:** [ROLE NAME]
**Agent:** [CLI Tool (Model)]
**Date:** 2026-01-25
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
3. JOURNALIST compiles all [N]-[ROLE]-[OBJECTIVE].md files into SESSION-LOG.md entry
4. JOURNALIST deletes all [N]-[ROLE]-[OBJECTIVE].md files after compilation

---

**End of SESSION-LOG.md Template**

Copy this template to your project root as `SESSION-LOG.md` and customize:
- Project name
- Repository URL/path
- Starting date
- Add project-specific rules to CRITICAL AGENT RULES section

Rock 'n Roll!  
JRENG!
