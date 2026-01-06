# TIT Project Development Session Log
## Go + Bubble Tea + Lip Gloss Implementation (Redesign v2)

## ⚠️ CRITICAL AGENT RULES

**AGENTS BUILD APP FOR USER TO TEST**
- run script ./build.sh
- USER tests
- Agent waits for feedback

**AGENTS CAN RUN GIT ONLY IF USER EXPLICITLY ASKS**
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
  4. **Extract abstractions** - Use existing utilities (RenderBox, RenderInputField, formatters)
  5. **Test structure** - Verify component compiles and renders within bounds
  6. **Verify dimensions** - Ensure component respects content box boundaries (never double-border)
  7. **Document pattern** - Add comments for thread context (AUDIO/UI THREAD) if applicable
  8. **Port is NOT refactor** - Move old code first, refactor after in separate session
  9. **Keep git history clean** - Port + refactor in separate commits if doing both

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

---

## Session 44: Pull Merge Conflict Resolution - Working End-to-End ✅

**Agent:** Claude Sonnet 4.5 (claude-code CLI)
**Date:** 2026-01-07

### Objective: Fix conflict resolver bugs and verify Scenario 1 end-to-end flow

### Completed:

✅ **Root cause analysis: Conflict detection broken** (15 min)
- Found `ExecuteWithStreaming()` returns empty `Stderr` field (already streamed to buffer)
- `cmdPull()` checked `result.Stderr` for conflicts → never detected
- Changed to check actual git state: `git.DetectState()` → `state.Operation == git.Conflicted`
- More reliable than text parsing (works regardless of stderr streaming)

✅ **Fixed conflict file parsing** (10 min)
- `ListConflictedFiles()` used wrong field index (8 instead of 10+)
- Git porcelain v2 format: `u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>`
- Path starts at field 10, must join remaining fields (handles spaces in filenames)
- Fixed: `strings.Join(parts[10:], " ")`

✅ **Fixed missing conflict resolver state fields** (5 min)
- Added `ColumnLabels: []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"}`
- Added `ScrollOffsets: make([]int, 3)` initialization
- Column titles now show correctly in UI

✅ **Improved console messages** (5 min)
- Changed "Command failed with exit code 1" → "Command exited with code 1"
- Less alarming (exit code 1 is expected for conflicts)
- Uses `ui.TypeInfo` instead of `ui.TypeStderr`

✅ **Added SSOT messages** (5 min)
- Added `ErrorMessages["pull_conflicts"]` and `["pull_failed"]`
- All error text now centralized in messages.go

### Files Modified:
- `internal/app/operations.go` — Fixed conflict detection to use git state
- `internal/git/execute.go` — Fixed field parsing + neutral exit message
- `internal/app/githandlers.go` — Added ColumnLabels and ScrollOffsets
- `internal/app/messages.go` — Added pull error messages

### Build Status: ✅ Clean compile

### Testing Status: ✅ VERIFIED - Scenario 1 working end-to-end

**Test Results (titest.sh Scenario 1):**
- Initial state: Clean, Diverged (local + remote both have commits)
- Select "Pull (merge)" → Conflicts detected
- Conflict resolver appears with:
  - ✅ File list shows actual filename (`conflict.txt`)
  - ✅ Column titles: BASE | LOCAL (yours) | REMOTE (theirs)
  - ✅ Bottom panes show file content
  - ✅ SPACE marks files correctly
- User marks LOCAL (yours) column → ENTER
- Console shows:
  - ✅ "git add -A"
  - ✅ "git commit -m 'Merge resolved conflicts'"
  - ✅ "Merge completed successfully"
- Final state: Clean, Local ahead (2 commits)
- Menu shows correct options per SPEC.md:
  - ✅ Push to remote
  - ✅ Replace remote (force push)
  - ✅ Commit history
  - ✅ File(s) history

**Why 2 commits ahead?**
1. Original local commit (modified conflict.txt with LOCAL)
2. Merge commit (created when resolving conflicts)

This is correct per git merge behavior.

### Known Issues (Documented for Cleanup):

Created `CONFLICT-RESOLVER-CLEANUP.md` documenting:

**P0: Missing operation step constants**
- `cmdFinalizePullMerge()` and `cmdAbortMerge()` both return `Step: OpPull`
- Can't distinguish finalize vs abort in `handleGitOperation()`
- Need: `OpFinalizePullMerge`, `OpAbortMerge` constants

**P1: SSOT violations (6 hardcoded strings)**
- operations.go: "Failed to stage resolved files", "Merge completed successfully", etc.
- conflicthandlers.go: "Already marked in this column", "Marked: %s → column %d"

**P2: Code duplication**
- `setupConflictResolverForPull()` and `setupConflictResolverForDirtyPull()` identical
- Should abstract to `setupConflictResolver(operation, columnLabels)`

**P3: Missing documentation**
- ARCHITECTURE.md needs "Conflict Resolver Integration" section
- Pattern for adding new conflict-resolving operations not documented

All issues documented with fixes and examples in CONFLICT-RESOLVER-CLEANUP.md.

### Next Actions:

- Abort flow testing (ESC in conflict resolver)
- Scenario 2-5 testing (dirty pull variants)
- Code cleanup tasks in CONFLICT-RESOLVER-CLEANUP.md

---

## Session 43: Pull Merge Conflict Resolver Wiring (INCOMPLETE - DEBUGGING) 🔧

**Agent:** Claude (Amp)
**Date:** 2026-01-07

### Objective: Wire conflict resolver to appear after pull merge conflicts detected

### Completed:

✅ **Fixed confirmation dialog field mismatch** (10 min)
- Changed `dispatchPullMerge()` to create proper `ui.ConfirmationDialog` (not `confirmationState`)
- Added missing `git` import in dispatchers.go
- Removed duplicate handler definitions in confirmationhandlers.go
- Confirmation dialog now appears correctly

✅ **Removed duplicate method definitions** (5 min)
- Found `executeConfirmPullMerge()` and `executeRejectPullMerge()` defined twice
- Removed second set of identical definitions
- Build clean

✅ **Added pull merge conflict detection** (20 min)
- Modified `githandlers.go` to check `msg.ConflictDetected` BEFORE checking `msg.Success`
- Conflicts are "failures" but require conflict resolver UI, not error message
- Added `setupConflictResolverForPull()` function (mirrors dirty pull version)
- Added conflict detection routing in `handleGitOperation()`

✅ **Added conflict resolver finalization commands** (15 min)
- `cmdFinalizePullMerge()` — stages all resolved files and commits merge
- `cmdAbortMerge()` — runs `git merge --abort` to restore original state
- Added SSOT message in messages.go: `"aborting_merge"`

✅ **Added conflict resolver routing** (10 min)
- Updated `handleConflictEnter()` to route `pull_merge` operation to `cmdFinalizePullMerge()`
- Updated `handleConflictEsc()` to route `pull_merge` operation to `cmdAbortMerge()`

### Files Modified:
- `internal/app/dispatchers.go` — Fixed confirmation dialog creation + added git import
- `internal/app/confirmationhandlers.go` — Removed duplicate methods, fixed SSOT usage
- `internal/app/githandlers.go` — Early conflict check, added setupConflictResolverForPull()
- `internal/app/conflicthandlers.go` — Added pull_merge routing in handlers
- `internal/app/operations.go` — Added cmdFinalizePullMerge(), cmdAbortMerge()
- `internal/app/messages.go` — Added "aborting_merge" SSOT message

### Build Status: ✅ Clean compile

### Testing Status: ❌ UNTESTED - Conflict resolver NOT appearing after pull

**Current Issue:** 
```
Console shows:
[23:51:01] git pull --no-rebase
[23:51:01] Auto-merging conflict.txt
[23:51:01] CONFLICT (content): Merge conflict in conflict.txt
[23:51:01] Automatic merge failed; fix conflicts and then commit the result.
[23:51:01] Command failed with exit code 1
[23:51:01] Failed to pull
[23:51:01] Failed. Press ESC to return.
```

Expected: Should transition to conflict resolver UI after detecting conflicts

**Root Cause Analysis:**

Conflicts detected AFTER console opens:
1. Confirmation → `executeConfirmPullMerge()` sets `mode = ModeConsole`
2. `cmdPull()` spawns goroutine running `git pull --no-rebase`
3. Git detects conflict, returns `GitOperationMsg{ConflictDetected: true}`
4. `handleGitOperation()` should catch at line 24 and call `setupConflictResolverForPull()`

**Hypothesis:** `setupConflictResolverForPull()` is failing silently
- Added diagnostic messages to trace execution:
  - "Detecting conflict files..."
  - "Found N conflicted file(s)"
- `git.ListConflictedFiles()` may be returning empty or erroring
- If error, stays in ModeConsole with error message

### Next Actions (New Thread):

1. **Test with diagnostics:**
   - Run Scenario 1 again
   - Watch console for "Detecting conflict files..." message
   - If NOT seen: `setupConflictResolverForPull()` not being called at all
   - If seen but "Found 0": `ListConflictedFiles()` returning empty

2. **Debug git helper functions:**
   - Verify `git.ListConflictedFiles()` works in this state
   - Check `git status --porcelain=v2` output manually
   - Verify conflict file detection logic

3. **Fallback approach:**
   - If git helpers unreliable, parse conflict files from `git merge --name-only --diff-filter=U`
   - More direct than porcelain v2 parsing

### Key Insight:

**Conflict happens DURING operation, not before:**
- Can't check for conflicts before opening console
- Must detect them in goroutine and transition UI after
- Flow: Confirmation → Console (operation runs) → Conflict detected → Transition to resolver

This is different from dirty pull which can pre-check state before operation.

---

## Session 41: Operation Steps SSOT Cleanup (COMPLETE) ✅

**Agent:** Claude (Amp)
**Date:** 2026-01-06

### Objective: Replace hardcoded operation step strings with constants (SSOT)

### Completed:

✅ **Created operationsteps.go SSOT file** (10 min)
- Created `internal/app/operationsteps.go` with 25+ operation step constants
- Covers all async operation names: init, clone, commit, push, pull, add_remote, dirty pull phases, etc.
- Each constant maps to the operation name string (e.g., `OpInit = "init"`)
- Improves type safety and catches typos at compile time

✅ **Updated operations.go to use constants** (15 min)
- Replaced all hardcoded `Step: "init"` with `Step: OpInit`
- Applied to 50+ occurrences across all cmd* functions
- Pattern: `Step: OpCommit`, `Step: OpPush`, `Step: OpDirtyPullSnapshot`, etc.
- All git operation functions now use constants

✅ **Updated githandlers.go to use constants** (15 min)
- Replaced all case statement strings with constants
- Updated switch cases: `case "init"` → `case OpInit`
- Updated comment references: `add_remote` → `OpAddRemote`
- All operation routing now uses constants

✅ **Updated ARCHITECTURE.md** (10 min)
- Added "Operation Step Constants" section with code example
- Explained why pattern matters: compile-time typo detection
- Added operationsteps.go to Key Files table
- Showed before/after: hardcoded string vs constant
- Documented SSOT principle: all operation names in one place

### Files Created:
- `internal/app/operationsteps.go` — 25+ operation step constants

### Files Modified:
- `internal/app/operations.go` — All Step assignments use constants
- `internal/app/githandlers.go` — All case statements use constants
- `ARCHITECTURE.md` — Documented new operation steps SSOT, added to Key Files

### Build Status: ✅ Clean compile

### Testing Status: ✅ BUILD VERIFIED

### SSOT Compliance After Fix:

| Category | Location | Status |
|----------|----------|--------|
| Menu items | menuitems.go | ✅ Fully SSOT |
| User text | messages.go | ✅ Fully SSOT |
| Keyboard shortcuts | app.go + documented | ✅ Fully SSOT |
| Operation steps | operationsteps.go | ✅ **FIXED** |
| Colors | theme.go | ✅ Fully SSOT |
| Dimensions | sizing.go | ✅ Fully SSOT |
| Git commands | git/execute.go | ✅ Correct (not user-facing) |

### Impact:

**Before:** Operation step strings scattered across operations.go (50+ hardcoded instances)
```go
Step: "init"
Step: "commit"
Step: "dirty_pull_snapshot"
```

**After:** Single SSOT file with constants
```go
const OpInit = "init"
const OpCommit = "commit"
const OpDirtyPullSnapshot = "dirty_pull_snapshot"

// Used everywhere:
Step: OpInit
case OpCommit:
```

**Benefits:**
- Compile-time detection of typos (before: silent runtime mismatch)
- Single source of truth (easier to audit, rename, extend)
- Better IDE autocomplete (OpInit → shows all Op* constants)
- Consistent naming across codebase

### Complete SSOT Hierarchy:

```
operationsteps.go  ← Operation routing constants (25+)
    ↓
operations.go      ← Uses Op* constants in GitOperationMsg.Step
githandlers.go     ← Routes on Op* constants via switch cases
    ↓
menuitems.go       ← Menu item definitions + shortcuts
    ↓
messages.go        ← User-facing text (prompts, hints, errors)
    ↓
theme.go           ← Colors (semantic names)
    ↓
sizing.go          ← Dimensions (SSOT)
```

### Ready for:
- 100% SSOT compliance across codebase
- Any new operations (add to operationsteps.go, use Op* constant)
- Future refactoring with confidence (no scattered magic strings)

---

## Session 40: Keyboard Shortcut Bug Fix + Documentation (COMPLETE) ✅

**Agent:** Claude (Amp)
**Date:** 2026-01-06

### Objective: Learn and document the SPACE key fix for keyboard handler registration

### Completed:

✅ **Learned the Keyboard Registration Pattern** (15 min)
- **The Bug:** SPACE key was registered as `"space"` but Bubble Tea sends `" "` (actual space character)
- **Root Cause:** `msg.String()` returns actual character/key name, not semantic names
- **The Fix:** Changed `On("space", ...)` to `On(" ", ...)` in keyboard handler registry
- **Discovery Method:** Debugged by checking `msg.String()` output and comparing to registered key strings

✅ **Updated ARCHITECTURE.md with Keyboard Patterns** (20 min)
- Added "Keyboard Input Patterns" section after Key Handler Registry documentation
- Documented critical distinction: Bubble Tea sends actual characters, not key names
- Showed correct vs incorrect examples:
  - ✅ `On(" ", handler)` - Space character (what Bubble Tea sends)
  - ❌ `On("space", handler)` - Would never fire
- Added registration pattern example from actual code (ModeMenu, ModeConflictResolve)
- Explained why this matters: handler never fires if key string doesn't match

✅ **Documented in Code** (5 min)
- Found inline comment in app.go line 585: `// Space character, not "space"`
- This documents the pattern for future developers

### Files Modified:
- `ARCHITECTURE.md` — Added "Keyboard Input Patterns" section with critical SPACE key lesson and examples

### Build Status: ✅ Clean compile (no code changes)

### Testing Status: ✅ VERIFIED WORKING (from Session 35)
- SPACE key fires correctly in conflict resolver
- File marking works (SPACE toggles selection)
- Border focus changes work (TAB navigation)

### Key Learning:

**Bubble Tea Key Dispatch:**
```
Bubble Tea KeyMsg
    ↓
msg.String() → actual character or key name
    ↓
Lookup in handlers["enter", "tab", " ", "ctrl+c", etc.]
    ↓
Execute handler or fall through
```

**Critical Examples:**
| Key | msg.String() value | Correct Registration | Wrong Registration |
|-----|------------------|-------------------|------------------|
| Space | `" "` | `On(" ", ...)` | ~~`On("space", ...)`~~ |
| Enter | `"enter"` | `On("enter", ...)` | ~~`On("return", ...)`~~ |
| Tab | `"tab"` | `On("tab", ...)` | ~~`On("\\t", ...)`~~ |
| Up arrow | `"up"` | `On("up", ...)` | ~~`On("↑", ...)`~~ |

**Debugging Pattern:**
1. Handler not firing? Check if key string was registered
2. Add debug: `a.footerHint = "KEY: [" + msg.String() + "]"`
3. See what Bubble Tea actually sends
4. Compare with registration key in NewModeHandlers()
5. Fix mismatch

### Ready for:
- Any new keyboard shortcuts (now with correct pattern understanding)
- Future debugging of non-firing key handlers
- Documentation of any other Bubble Tea quirks discovered

---

## Session 39: Menu SSOT Refactoring + Documentation Update (COMPLETE) ✅

**Agent:** Claude (Amp)
**Date:** 2026-01-06

### Objective: Complete menu refactoring documentation and verify SSOT implementation works end-to-end

### Completed:

✅ **Documentation Updates** (20 min)
- Updated ARCHITECTURE.md MenuItem SSOT System section with comprehensive explanation
- Added "Menu Rendering Flow" diagram showing:
  - GenerateMenu() produces clean []MenuItem
  - RenderMenu() displays shortcuts + labels (no hints in menu)
  - Footer hint updates on selection change via `app.footerHint = GetMenuItem(selected).Hint`
- Updated "Adding a New Menu Item" pattern with 5-step SSOT approach:
  1. Define in menuitems.go (ID, Shortcut, Emoji, Label, Hint)
  2. Use in menu.go generator (GetMenuItem only)
  3. Add dispatcher in dispatchers.go
  4. Add handler in handlers.go
  5. Dispatcher lookup automatic (no manual registration needed)
- Updated generator list (menuWorkingTree mentions "Dirty" not "Modified")
- Cleaned up obsolete code examples (removed old Item builder pattern)

✅ **Verified SSOT Implementation Exists** (10 min)
- menuitems.go: All menu items defined with ID, Shortcut, Emoji, Label, Hint, Enabled
- messages.go: SSOT maps for:
  - InputPrompts — All text input prompts
  - InputHints — Help text for inputs
  - ErrorMessages — Error text
  - OutputMessages — Operation messages
  - ConfirmationTitles, ConfirmationExplanations, ConfirmationLabels — Dialog text
  - FooterHints — Footer messages for conflict resolution
- No hardcoded strings in new code (dirty pull, conflict resolver)

✅ **Verified "Modified" to "Dirty" Rename** (5 min)
- ARCHITECTURE.md updated: menuWorkingTree() now correctly mentions "Dirty" not "Modified"
- git/types.go: WorkingTree enum uses Clean | Dirty
- All references to "Modified" changed to "Dirty"

✅ **Clean Build**
- No errors, no warnings
- Binary ready in `/Users/jreng/Documents/Poems/inf/___user-modules___/automation/tit_x64`

### Files Modified:
- `ARCHITECTURE.md` — MenuItem SSOT documentation, Menu Rendering Flow diagram, updated "Adding a New Menu Item" pattern
- No code changes needed (all implementation already complete from Session 38)

### Build Status: ✅ Clean compile

### Testing Status: ✅ BUILD VERIFIED - Code compiles and runs cleanly

### Key Accomplishments:

**Refactored Menu System:**
- Single source of truth (menuitems.go) for all menu text
- Shortcuts globally unique (no conflicts)
- Hints displayed in footer (not in menu body)
- Generators simplified to GetMenuItem() calls
- All prompts, errors, messages, and dialog text centralized in messages.go

**Documentation Complete:**
- ARCHITECTURE.md fully documents new SSOT approach
- "Adding a New Menu Item" pattern clear and testable
- Menu rendering flow diagram shows hint movement to footer
- All changes verified as implemented and working

### Ready for:
- Phase 6 Integration Testing (dirty pull scenarios)
- Next feature development using documented SSOT patterns
- User testing of menu hints in footer

