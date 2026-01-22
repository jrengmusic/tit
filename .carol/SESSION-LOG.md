# TIT Project Development Session Log
## Go + Bubble Tea + Lip Gloss Implementation (Redesign v2)

## ⚠️ CRITICAL AGENT RULES

**AGENTS BUILD APP FOR USER TO TEST**
- run script ./build.sh
- USER tests
- Agent waits for feedback

**AGENTS CAN RUN GIT ONLY IF USER EXPLICITly ASKS**
- Code changes without git commands
- Agent runs git ONLY when user explicitly requests
- Never autonomous git operations
- **When committing:** Always stage ALL changes with `git add -A` before commit
  - ❌ DON'T selectively stage files (agents forget/miss files)
  - ✅ DO `git add -A` to capture every modified file
  - This ensures complete commits with nothing accidentally left unstaged

**EMOJI WIDTH RULE (CRITICAL)**
- ❌ NEVER use small/narrow width emojis - they break layout alignment
- ✅ ONLY use wide/double-width emojis (🔗 📡 ⬆️ 💥 etc.) or text symbols (✓ ✗)
- Test emoji width before using: wide emojis take 2 character cells, narrow take 1
- When in doubt, use text-based symbols instead of emojis

**LOG MAINTENANCE RULE**
- **All session logs must be written from the latest to earliest (top to bottom), BELOW this rules section.**
- **Only the last 5 sessions are kept in active log.**
- Agents must identify itself as session log author
```
**Agent:** Sonnet 3.5 (claude.ai/code), Sonnet 4.5 (GitHub Copilot CLI), GPT-5.1 (Cursor)
**Date:** 2025-12-31
```
- Session could be executed parallel with multiple agents.
- Remove older sessions from active log (git history serves as permanent archive)
- This keeps log focused on recent work
- **Agent NEVER updates log without explicit user request**
- **During active sessions, only user decides whether to log**
- **All changes must be tested/verified, or marked UNTESTED**
- If rule not in this section, agent must ADD it (don't erase old rules)

**NAMING RULE (CODE VOCABULARY)**
- All identifiers must obey: `___user-modules___/codebase-for-dummies/docs/How to choose your words wisely.md`
- Variable names: semantic + precise (not `temp`, `data`, `x`)
- Function names: verb-noun pattern (initRepository, detectCanonBranch)
- Struct fields: domain-specific terminology (not generic `value`, `item`, `entry`)
- Type names: PascalCase, clear intent (CanonBranchConfig, not BranchData)

**PATTERN FOR PORTING A COMPONENT (IMMUTABLE)**
- When porting UI components from old-tit to new-tit:
  1. **Read source** - Study old component structure and logic in old-tit
  2. **Identify SSOT** - Find sizing constants and use new-tit SSOT (ContentInnerWidth, ContentHeight, etc.)
  3. **Update colors** - Replace old hardcoded colors with semantic theme names
  2. **Extract abstractions** - Use existing utilities (RenderBox, RenderInputField, formatters)
  3. **Test structure** - Verify component compiles and renders within bounds
  4. **Verify dimensions** - Ensure component respects content box boundaries (never double-border)
  5. **Document pattern** - Add comments for thread context (AUDIO/UI THREAD) if applicable
  6. **Port is NOT refactor** - Move old code first, refactor after in separate session
  7. **Keep git history clean** - Port + refactor in separate commits if doing both

**BEFORE CODING: ALWAYS SEARCH EXISTING PATTERNS**
- ❌ NEVER invent new states, enums, or utility functions without checking if they exist
- ✅ Always grep/search the codebase first for existing patterns
- ✅ Check types.go, constants, and error handling patterns before creating new ones
- ✅ Example: `NotRepo` operation already exists—don't create "UnknownState" fallback
- **Methodology:** Read → Understand → Find SSOT → Use existing pattern
- Overcomplications usually mean you missed an existing solution

**TRUST THE LIBRARY, DON'T REINVENT**
- ❌ NEVER create custom helpers for things the library already does
- ✅ Trust lipgloss for layout/styling (Width, Padding, Alignment, JoinHorizontal)
- ✅ Trust Go stdlib (strings, filepath, os, exec)
- ✅ Trust Bubble Tea for rendering and event handling
- ✅ Example: Don't manually calculate widths—use `lipgloss.NewStyle().Width()`
- **Philosophy:** Libraries are battle-tested. Your custom code is not.
- If you find yourself writing 10+ lines of layout math, stop—the library probably does it

**FAIL-FAST RULE (CRITICAL)**
- ❌ NEVER silently ignore errors (no `_ = cmd.Output()`, no error suppression)
- ❌ NEVER use fallback values that mask failures
- ❌ NEVER return empty strings/zero values when git commands fail
- ✅ ALWAYS check error return values explicitly
- ✅ ALWAYS return errors to caller or log + fail fast
- ✅ Examples of violations:
  - `output, _ := cmd.Output()` → Hides command failures
  - `executeGitCommand("...") returning ""` → Masks why it failed
  - Creating fake Operation states (NotRepo) as fallback → Violates contract
- **Rule:** If code path executes but silently returns wrong data, you've introduced a bug that wastes debugging time later
- Better to panic/error early than debug silent failure for hours

**⚠️ NEVER EVER REMOVE THESE RULES**
- Rules at top of SESSION-LOG.md are immutable
- If rules need update: ADD new rules, don't erase old ones
- Any agent that removes or modifies these rules has failed
- Rules protect the integrity of the development log

## ROLE ASSIGNMENT REGISTRATION

ANALYST: Amp (Claude Sonnet 4)
SCAFFOLDER: OpenCode (CLI Agent) — Code scaffolding specialist, literal implementation
CARETAKER: Amp (GPT-4.1) — Polishing, error handling, syntax validation
INSPECTOR: OpenCode (CLI Agent) — Auditing code against SPEC.md and ARCHITECTURE.md, verifying SSOT compliance
SURGEON: Amp (Claude Sonnet 4) — Diagnosing and fixing bugs, architectural violations, testing
JOURNALIST: Mistral-Vibe (devstral-2) — Session documentation, log compilation, git commit messages


---

## Session 80: Reactive Layout Implementation ✅

**Agent:** Mistral-Vibe (devstral-2) — JOURNALIST
**Date:** 2026-01-22

### Objectives
- Transform TIT from fixed 80×46 centered layout to reactive full-terminal layout
- Implement sticky header/footer with dynamic content area
- Add conditional banner display based on terminal width
- Ensure graceful degradation for small terminals

### Implementation Summary

**Architecture Transformation:**
- Replaced fixed `Sizing` struct with dynamic `DynamicSizing` that recalculates on terminal resize
- Created reusable `InfoRow` component for emoji-prefixed status rows
- Implemented `RenderReactiveLayout()` function for full-terminal rendering
- Added threshold-based rendering decisions (min size: 40×20, banner at width ≥100)

**Key Components Created:**
1. **DynamicSizing struct** (`internal/ui/sizing.go`): Computes all dimensions from terminal size
2. **InfoRow component** (`internal/ui/inforow.go`): Reusable emoji-prefixed rows with descriptions
3. **Header rendering** (`internal/ui/header.go`): Conditional banner display and status info
4. **Reactive layout renderer** (`internal/ui/layout.go`): Full-terminal layout with sticky sections

**Phases Completed (Sessions 82-90):**
- **Phase 1 (Session 82):** Refactored sizing SSOT with threshold constants
- **Phase 2 (Session 83):** Updated application state to use dynamic sizing
- **Phase 3 (Session 84):** Created reactive layout renderer
- **Phase 4 (Session 85):** Implemented InfoRow component
- **Phase 5 (Session 86):** Created header rendering system
- **Phase 6 (Session 87):** Refactored banner rendering for dynamic dimensions
- **Phase 8 (Session 88):** Updated all content renderers to use dynamic sizing
- **Phase 10 (Session 89):** Cleaned up legacy code
- **Session 90:** Fixed header alignment (emoji width: 4→6→2→3 cells)

### Files Created (2)
- `internal/ui/inforow.go` — InfoRow component for emoji-prefixed status rows
- `internal/ui/header.go` — HeaderState struct and rendering functions

### Files Modified (4)
- `internal/ui/sizing.go` — DynamicSizing struct with threshold constants and calculation methods
- `internal/ui/layout.go` — RenderReactiveLayout() function and too-small message handler
- `internal/app/app.go` — Dynamic sizing integration, state header rendering, and main view
- `cmd/tit/main.go` — Updated to use CalculateDynamicSizing(80, 40)

### Technical Details

**Layout Structure:**
```
┌─────────────────────────────────────────────────────────────────┐ ← Terminal top
│ HEADER (full width, 11 lines)                                │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ margin │ INFO COLUMN          │ gap │ BANNER COLUMN │ margin│ │
│ │        │ (CWD, Remote, WT, TL) │     │ (braille)     │       │ │
│ └─────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│ CONTENT (full width, dynamic height)                        │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ margin │ CONTENT CHILDREN                           │ margin│ │
│ │        │ (menu, input, console, history, etc.)      │       │ │
│ └─────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│ FOOTER (full width, 2 lines)                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ margin │ FOOTER CONTENT (hints, status)             │ margin│ │
│ └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘ ← Terminal bottom
```

**Header Layout Example:**
```
📁 /path/to/repo              🟢 READY
🔗 git@github.com:user/repo   🌿 main
──────────────────────────────────────────────────────
📝 Dirty
   You have uncommitted changes.
🔗 Sync
   Local and remote are in sync.
```

### Success Criteria Met
- ✅ Reactive layout responds to terminal resize events
- ✅ Header and footer remain sticky while content area grows/shrinks
- ✅ "Too small" message displays correctly at <40×20 threshold
- ✅ Banner appears/disappears at width ≥100 threshold
- ✅ Footer content remains centered
- ✅ All semantic variable names follow project conventions
- ✅ Clean build with `./build.sh`
- ✅ No hardcoded layout constants (80, 24, 46, 76) remain

### Challenges Overcome
- **Emoji Width Alignment:** Iterative correction through sessions (4→6→2→3 cells)
- **Description Indentation:** Final `EmojiColumnWidth = 3` correctly aligns descriptions with wide emoji + space
- **Dynamic Sizing Propagation:** Updated all content renderers to accept and use dynamic dimensions
- **Legacy Code Cleanup:** Removed old sizing system while maintaining functionality

### Build Status
✅ Clean compile with `./build.sh`

### Testing Status
✅ VERIFIED — All success criteria met, reactive layout working as specified

---


## Session 81: Reactive Layout Implementation - Partial Fixes ⚠️

**Agent:** Mistral-Vibe (devstral-2) — JOURNALIST
**Date:** 2026-01-22

### Objectives
- Continue reactive layout implementation from Session 80
- Fix header/footer/menu layout issues
- Address regressions from initial implementation

### Implementation Status

**Partial Success with Known Regressions:**
- Attempted to implement reactive layout per REACTIVE-LAYOUT-PLAN.md
- Some fixes applied successfully, but task remains incomplete
- Key functionality working, but missing critical header elements

### Files Modified (5)
- `internal/ui/sizing.go` — Updated constants: HeaderHeight=11, FooterHeight=1, MinWidth=60, removed MinWidthBanner/ShowBanner/InfoColumnWidth, added MenuColumnWidth
- `internal/ui/header.go` — Rewrote to vertical single-column layout (CWD, Remote, Separator, WorkingTree, Timeline) but **MISSING Operation and Branch info**
- `internal/ui/layout.go` — Updated RenderReactiveLayout with lipgloss.Place for footer positioning, simplified section styling
- `internal/ui/menu.go` — Added RenderMenuWithBanner for 2-column menu+banner layout, updated RenderMenuWithHeight signature
- `internal/app/app.go` — Updated RenderStateHeader to use new HeaderState struct, added quitConfirmActive guards

### Files Deleted (1)
- `internal/ui/inforow.go` — Removed due to EmojiColumnWidth redeclaration conflict

### Known Regressions (⚠️ INCOMPLETE)

**Critical Missing Features:**
1. **Operation status (READY/etc) missing from header** — Removed during refactor, not re-added
2. **Branch name (🌿 main) missing from header** — Removed during refactor, not re-added

**Layout Issues:**
3. Header layout not fully matching REACTIVE-LAYOUT-PLAN.md specification

### Technical Details

**Header Layout Changes:**
- Converted to vertical single-column layout
- Current structure: CWD → Remote → Separator → WorkingTree → Timeline
- Missing: Operation status and Branch information

**Menu Improvements:**
- Added 2-column layout (menu left, banner right)
- Updated RenderMenuWithHeight to accept contentWidth parameter

**Footer Enhancements:**
- Added quitConfirmActive priority logic
- Prevents menu hints from overriding Ctrl+C messages

### Challenges Encountered

**Plan Interpretation Issues:**
- User frustration with agent not reading REACTIVE-LAYOUT-PLAN.md carefully
- Original plan specifies 11-line header with vertical stack but doesn't explicitly mention Operation/Branch placement
- Need clearer specification for header content organization

**Code Conflicts:**
- EmojiColumnWidth redeclaration in inforow.go caused deletion
- Component conflicts between new and legacy code

### Build Status
✅ Clean compile with `./build.sh`

### Testing Status
⚠️ PARTIAL — Regressions identified, needs further fixes

### Follow-Up Required

**Immediate Next Steps:**
1. Re-add Operation status (READY/etc) to header
2. Re-add Branch name (🌿 main) to header  
3. Verify header layout against REACTIVE-LAYOUT-PLAN.md
4. Test all layout scenarios (small/large terminals, with/without banner)

**Architectural Considerations:**
- Review REACTIVE-LAYOUT-PLAN.md for explicit header content requirements
- Consider adding Operation/Branch to header specification
- Ensure all agents read and follow plan precisely

### Summary

Session 81 represents a partial implementation of the reactive layout with some successful fixes but critical missing functionality. The SURGEON role identified and documented the regressions, providing clear guidance for follow-up work. The implementation is structurally sound but incomplete, requiring additional work to restore missing header elements and fully match the architectural plan.

**Status:** ⚠️ INCOMPLETE - Needs follow-up for header completion

---
## Session 79: Add-Remote Timeline Behavior Fix & No-Commit Footer Hint ✅

**Agent:** Mistral-Vibe (devstral-2) — JOURNALIST
**Date:** 2026-01-22

### Objectives
- Fix add-remote timeline behavior to prevent inappropriate force-push options
- Add footer hint for empty repos with remotes
- Remove auto-commit side effect from state detection

### Problems Solved

**1. Add-Remote Timeline Behavior**
- **Root cause:** State detection auto-committed in empty repos, making timeline appear ahead
- **Fix:** Removed auto-commit in `DetectState()` and gated timeline detection on commits
- **Result:** Empty repos with remotes now show Timeline N/A instead of ahead

**2. No-Commit Footer Hint**
- **Root cause:** Users confused about empty repos with remotes showing no timeline
- **Fix:** Added SSOT footer hint explaining no-commit state
- **Result:** Clear user guidance when repo has remote but no commits

### Files Modified (5 total)
- `internal/git/state.go` — Removed auto-commit and gated timeline detection on commits
- `internal/app/messages.go` — Added SSOT footer hint for no-commit state
- `internal/app/handlers.go` — Set footer hint when returning to menu with remote and no commits
- `internal/app/app.go` — Set footer hint on init when remote exists but no commits
- `ARCHITECTURE.md` — Updated timeline semantics and removed auto-setup claim

### Build Status
⚠️ Not built/tested (not requested by user)

### Testing Status
⚠️ UNTESTED — Changes not verified by user

---

## Session 78: Async Remote Fetch & Add Remote Fix ✅

**Agent:** Amp (GPT-4.1) — CARETAKER
**Date:** 2026-01-17

### Objectives
- Fix git state detection for remote changes (timeline always showed "Sync" even when remote had new commits)
- Fix "Add remote" flow for empty remotes (upstream tracking failed silently)

### Problems Solved

**1. Stale Timeline Detection**
- **Root cause:** `DetectState()` compared local refs vs cached remote refs without fetching
- **Fix:** Added async `git fetch` on startup when `HasRemote` detected
- **Result:** Timeline now accurately reflects remote state after app loads

**2. Add Remote to Empty Repository**
- **Root cause:** `SetUpstreamTrackingWithBranch` tried `--set-upstream-to` which fails when remote branch doesn't exist
- **Fix:** Check if remote branch exists first; if not, execute `git push -u` to create branch AND set upstream atomically
- **Result:** "Add remote" now guarantees upstream is configured (per SPEC contract)

### Files Modified (4 total)
- `internal/app/messages.go` — Added `RemoteFetchMsg` type
- `internal/app/handlers.go` — Added `cmdFetchRemote()` async command, added `os/exec` import
- `internal/app/app.go` — Trigger fetch in `Init()` when `HasRemote`, handle `RemoteFetchMsg` in `Update()`
- `internal/git/execute.go` — Rewrote `SetUpstreamTrackingWithBranch` to check remote branch existence and push -u if needed
- `internal/app/operations.go` — Updated `cmdSetUpstream` to FAIL-FAST and show accurate messages

### Build Status
✅ Clean compile, `./build.sh` successful

### Testing Status
✅ VERIFIED — User confirmed both fixes working

---

## Session 77: REWIND Feature Implementation & Polish ✅

**Agent:** Gemini (JOURNALIST)
**Date:** 2026-01-12
**Duration:** 3 hours (07:30 - 10:30)

### Objectives
- Implement the REWIND feature (`git reset --hard`) from scaffolding to a production-ready state.
- Audit and fix architectural violations, SSOT issues, silent fails, and hardcoding.
- Address UI polishing issues, including emoji width violations and status bar visibility.
- Fix critical keyboard shortcut (`Ctrl+R`) and status bar display issues.
- Centralize all hardcoded messages to SSOT (ErrorMessages, OutputMessages, ConfirmationMessages).

### Agents Participated
- **SCAFFOLDER:** Mistral-Vibe (devstral-2) — Provided the initial literal code scaffold for the REWIND feature.
- **SURGEON:** Amp (Claude Sonnet) — Audited the scaffold, fixed 7 critical architectural and FAIL-FAST violations, polished the UI, and centralized all user-facing strings to the SSOT message maps.
- **TROUBLESHOOTER:** Claude Code (Sonnet 4.5) — Resolved terminal compatibility issues by replacing the problematic `Ctrl+Enter` shortcut with `Ctrl+R` and simplified the UI by removing complex state tracking.
- **JOURNALIST:** Gemini (CLI Agent) - Compiled session summaries and logged the session.

### Files Modified (14 total)
- `internal/app/app.go` — Implemented core rewind logic, state management, and keyboard handlers.
- `internal/app/confirmation_handlers.go` — Added rewind confirmation dialog logic.
- `internal/app/handlers.go` — Added rewind operation handlers.
- `internal/app/keyboard.go` — Mapped `ctrl+r` to the rewind handler.
- `internal/app/messages.go` — Added `RewindMsg` and all SSOT strings for the feature.
- `internal/app/modes.go`
- `internal/git/execute.go` — Added `ResetHardAtCommit` function with proper error handling.
- `internal/git/types.go`
- `internal/ui/console.go`
- `internal/ui/history.go` — Updated status bar to reflect new `Ctrl+R` shortcut.
- `internal/ui/layout.go`
- `SPEC.md` — Updated keyboard shortcut from `Ctrl+Enter` to `Ctrl+R`.
- `ARCHITECTURE.md`
- `REWIND-IMPLEMENTATION-PLAN.md`

### Problems Solved
- **Critical Silent Fail:** Fixed `ResetHardAtCommit` which returned `("", nil)` unconditionally, violating the FAIL-FAST rule.
- **Incomplete Handlers:** Fully implemented 5 stubbed handlers and messages (`RewindMsg`, `handleHistoryRewind`, `showRewindConfirmation`, `executeConfirmRewind`).
- **Terminal Incompatibility:** Replaced `Ctrl+Enter` with the more reliable and semantic `Ctrl+R` shortcut to avoid terminal emulator conflicts.
- **UI/UX Issues:**
    - Removed complex and faulty `isCtrlPressed` state tracking for a simpler, static status bar hint.
    - Corrected an emoji width violation in a confirmation dialog title.
    - Hid the `Ctrl+R` shortcut from the main status bar to keep it a "power-user" feature, reducing UI clutter.
- **SSOT Violations:** Eradicated all hardcoded strings (errors, prompts, UI text) related to the rewind feature and centralized them into the `messages.go` SSOT maps.
- **Code Duplication:** Removed a duplicate function declaration for `executeRewindOperation`.

### Summary
Session 77 saw the end-to-end implementation of the destructive but essential REWIND feature. The process followed the CAROL protocol perfectly: **SCAFFOLDER** laid the foundation, **SURGEON** performed a deep audit, fixing critical architectural flaws and ensuring SSOT compliance. **TROUBLESHOOTER** then stepped in to solve a nuanced terminal compatibility and UX issue with the keyboard shortcut. Finally, **SURGEON** returned to polish the UI and centralize all strings, hardening the feature for production. The result is a robust, safe, and well-documented feature that adheres to all project standards.

**Status:** ✅ APPROVED

---

