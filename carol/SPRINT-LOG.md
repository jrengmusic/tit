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

## Sprint 12: Config Menu Collapse + Branch Picker Actions ✅

**Date:** 2026-04-15
**Duration:** ~00:45

### Agents Participated
- **COUNSELOR** — Problem framing, config/branch menu reorg planning, footer hint dynamic design, early-return cleanup directive
- **Pathfinder** — Codebase discovery of config menu, branch picker, workflow state, dispatcher wiring, tag origin trace
- **Engineer** — Menu SSOT addition, dispatcher rewiring, picker handlers, confirmation handlers, workflow state extension, footer hint keys, single-return refactor of all picker handlers

### Files Modified (9 total)
- `internal/app/menu_items.go` — added `config_branch` SSOT entry (shortcut `b`, emoji 🌿, label "Branch", hint "Open branch picker"); retired `config_new_branch`/`config_switch_branch`/`config_merge_branch` from menu path (definitions retained as unused SSOT)
- `internal/app/menu_render_extra.go:166-195` — collapsed `GenerateConfigMenu`: removed inner separator between switch-remote and remove-remote; replaced 3 branch items with single `config_branch` entry
- `internal/app/dispatchers.go:35-42` — removed `config_switch_branch` and `config_merge_branch` from dispatcher map; added `config_branch → dispatchConfigSwitchBranch`; kept `config_new_branch` entry
- `internal/app/app_keys.go:128-137` — registered `a`, `m`, `x` key bindings on `ModeBranchPicker`
- `internal/app/handlers_config_branch.go` — added `handleBranchPickerAdd`, `handleBranchPickerMerge`, `handleBranchPickerDelete`, `refreshBranchPicker(selectName string) error` helper; refactored all 7 functions in file to single-return, positive-check control flow (CAROL contract)
- `internal/app/confirm_dialog_handlers.go` — added `"branch_delete"` confirmation pair; implemented `executeConfirmBranchDelete` (runs `git branch -D`, refreshes picker, stays in `ModeBranchPicker`) and `executeRejectBranchDelete` (hides dialog, returns to picker)
- `internal/app/workflow_state.go` — added `BranchPickerReturnAfterCreate bool` field to `WorkflowState`
- `internal/app/dispatch_dialog.go:64-73,108-115` — moved `PreviousMode = ModeConfig` assignment into `dispatchConfigNewBranch` so picker caller can override with `ModeBranchPicker`; removed hardcoded footer hint string from `dispatchConfigSwitchBranch` (footer now state-driven)
- `internal/app/handlers_git_branch.go:56-92` — removed hardcoded `PreviousMode = ModeConfig` from `handleNewBranchNameSubmit` (caller owns that field now)
- `internal/app/handlers_pull.go:83-105` — `handleBranchSwitch` now branches on `BranchPickerReturnAfterCreate`: on successful `OpBranchCreate` from picker path, refreshes picker with cursor on new branch and returns to `ModeBranchPicker` instead of `ModeConsole`
- `internal/app/footer.go:98-106` — dynamic footer hint key for `ModeBranchPicker`: `branch_picker_current` when selected row is current branch, `branch_picker_other` otherwise
- `internal/app/messages_menu.go:151-168` — replaced single `branch_picker` hint entry with two variants: current row shows `↑↓ / a / Enter / Esc`; other rows show `↑↓ / a / m / x / Enter / Esc`

### Alignment Check
- [x] BLESSED / LIFESTAR principles followed
- [x] NAMES.md adhered (shortcut `b` for Branch, `a`/`m`/`x` for add/merge/delete; identifiers semantic)
- [x] MANIFESTO principles applied (Menu = Contract preserved: `m`/`x` disabled + hidden on current branch row, no-op handlers; Explicit Encapsulation maintained)
- [x] Control flow contract: zero early returns across all 7 functions in `handlers_config_branch.go`; positive checks only; single terminal return per function

### Problems Solved
- **Config menu clutter:** three separate branch entries (`New Branch`, `Switch Branch`, `Merge from...`) consolidated into single `Branch` entry that opens the picker. Config menu now reads: remote ops / branch / preferences / back — three logical groups separated by two separators.
- **Branch picker was read-only + single-purpose:** added full CRUD surface (add/switch/merge-from/delete) with contract-aware gating — merge and delete hidden + shortcut-disabled on current-branch row.
- **Dialog return path:** new-branch creation from picker previously exited to main menu. Added `BranchPickerReturnAfterCreate` flag; on success the picker rebuilds and cursor lands on the newly created branch.
- **Static footer hint:** branch picker hint was a single fixed string. Now state-driven through the existing `FooterHintShortcuts` SSOT with two variants selected by `IsCurrent`.
- **Delete safety:** new confirmation dialog (`"Delete branch <name>?"`, YES=Delete, NO=Cancel, default=Cancel) gates the `git branch -D` call; rejection returns to picker without side effects.
- **CAROL contract compliance:** all seven branch picker handlers (including three pre-existing ones with guard-style early returns) refactored to single-return form.

### Technical Debt / Follow-up
- None. Sprint closed clean. `config_new_branch`/`config_switch_branch`/`config_merge_branch` SSOT entries and dispatcher function bodies are retained intentionally (still reachable from picker-`a` via direct `transitionTo` and from picker selection handler via `handleMergeBranchSelection`).

### Debts Paid
- None

### Debts Deferred
- None

---

## Sprint 11: Release Infrastructure + Final Audit Remediation ✅

**Date:** 2026-04-05

### Agents Participated
- **COUNSELOR** — Release planning, audit coordination, SPEC/ARCHITECTURE fixes
- **Pathfinder** — Codebase discovery, call site analysis, CAKE reference files
- **Auditor** — Final BLESSED-LANGUAGE audit (3 passes: implementation, docs/godocs, final gate)
- **Engineer** — Module migration, release files, version injection, ShortenHash SSOT, godocs, stale doc fixes
- **Librarian** — goreleaser v2 deprecation research

### Files Modified

**Release infrastructure (created):**
- `.goreleaser.yaml` — goreleaser v2 config, 6 targets, macOS signing, ldflags version injection
- `entitlements.plist` — macOS codesign entitlements
- `scripts/post-build.sh` — macOS sign + notarize hook
- `release.sh` — One-command release: commit, tag, push, goreleaser. Deletes existing release+tag if re-releasing.
- `RELEASE_NOTES.md` — GitHub release description

**Module path migration (97 files):**
- `go.mod` — `tit` -> `github.com/jrengmusic/tit`
- 96 `.go` files — all imports updated

**Version injection:**
- `internal/constants.go` — `AppVersion` from `const "v1.3.1"` to `var "dev"` (injected via ldflags)
- `build.sh` — Added `git describe --tags` version detection + ldflags injection

**Final audit remediation:**
- `SPEC.md` — Added `Rewinding` to Operation table, version placeholder in UI mockup
- `ARCHITECTURE.md` — Rewrote pre-flight blocker sections (2 CRITICAL), updated Application struct, fixed stale file references, updated method listings, fixed menu generator documentation
- `CODEBASE-MAP.md` — Removed 4 stale file references (`statusbar.go`, `input.go`, `textpane.go`)
- `internal/ui/formatters.go` — Deleted duplicate `ShortenHash` (SSOT: `git.ShortenHash` is canonical)
- `internal/ui/history.go` — Migrated to `git.ShortenHash`
- `internal/app/confirm_dialog_render.go`, `handlers_history_cache.go`, `handlers_history_copyhash.go` — Migrated to `git.ShortenHash`
- `internal/git/types.go` — Added doc comments to 4 exported types (WorkingTree, Timeline, Operation, Remote)
- `internal/ui/header.go` — Added doc comment to HeaderState
- `internal/git/init.go` — Fixed 3 non-godoc-compliant comments
- `internal/git/state.go`, `branch.go`, `app_constructor.go` — Annotated 5 silent error discards
- `.goreleaser.yaml` — Fixed `format` -> `formats` deprecation

**Deleted:**
- `PLAN-refactor-blessed.md` — Completed plan
- `RFC.md` — Consumed RFC

### Alignment Check
- [x] BLESSED principles followed
- [x] LANGUAGE.md Go addendum applied
- [x] NAMES.md adhered
- [x] MANIFESTO.md principles applied
- [x] SPEC.md updated (Rewinding, version placeholder)
- [x] ARCHITECTURE.md updated (pre-flight myth removed, struct updated)

### Problems Solved
- `ShortenHash` duplicated in `git/types.go` and `ui/formatters.go` — deleted ui copy, all callers migrated to `git.ShortenHash`
- ARCHITECTURE.md falsely claimed Conflicted/Merging/Rebasing/DirtyOperation block startup — rewritten to document actual recovery behavior
- ARCHITECTURE.md showed flat Application struct from pre-refactor era — updated to embedded state clusters
- 5 exported types missing godoc comments — added
- 3 exported functions had non-compliant godoc format — fixed
- 5 silent error discards violated LANGUAGE.md E — annotated
- goreleaser `format` deprecated in v2 — changed to `formats`
- `release.sh` didn't delete GitHub release on re-release — added `gh release delete`

### Technical Debt / Follow-up
- W5: KeyHandler dual receiver+param is systemic framework convention. Accepted.
- ARCHITECTURE.md has additional stale sections beyond what was fixed (field counts, GitEnvironment type description, Logger interface). Full rewrite recommended in a dedicated docs sprint.
- No tests for `internal/config/` and `internal/banner/` packages.

**Status:** ✅ Build passes. 56 tests pass. Vet clean. Released as v0.0.1.

---

## Sprint 10: LANGUAGE-BLESSED Production Quality Audit Remediation ✅

**Date:** 2026-04-05

### Agents Participated
- **COUNSELOR** — Full audit planning, SPEC updates, LANGUAGE.md authoring, phase execution coordination
- **Pathfinder** — Codebase discovery, accessor audit, call site analysis, line counts
- **Auditor** — Initial audit (27 findings), Phase 3 verification, Phase 5 verification, implementation validation, final gate audit (3 passes)
- **Engineer** — Phase 1-8 execution (dead code, accessors, constants, control flow, file splits, bug fixes, type safety, UI boundary)
- **Librarian** — Bubbletea architecture research
- **Oracle** — (not invoked)

### Files Modified (87 total, 1180 insertions, 4560 deletions)

**Created:**
- `~/.carol/LANGUAGE.md` — Multi-language BLESSED compliance addendum (Go + C++ sections, bubbletea framework)
- `internal/ui/menu_item.go` — MenuItem struct (moved from app, eliminates map[string]interface{})
- `internal/app/handlers_git_result.go` — Extracted git operation result handlers from git_handlers.go
- `internal/app/op_rebase.go` — Rebase conflict handling (cmdRebaseContinue/Abort, handleRebaseContinue/Abort)

**Deleted:**
- `internal/git/exec_base.go` — Dead file (TODO comment only)
- `internal/app/operations.go` — Dead file (package declaration only)
- `internal/app/part2.txt` — Dead pre-refactor artifact
- `PLAN-branch-operations.md` — Stale plan from Sprint 7

**Phase 1 — Dead Code:**
- `internal/app/app_update_msg.go` — Deleted `updateCallCount` global and increment
- `internal/app/app_init.go` — Deleted dangling `GetFooterHint` comment
- `internal/app/app_view_header.go` — Collapsed duplicate RenderStateHeader comment
- `internal/app/app_update_cmd.go` — Deleted duplicate comment block, deleted dead `GetGitState()` method

**Phase 2 — Accessor Cleanup:**
- `internal/app/ui_state.go` — Deleted 2 simple accessors, renamed `SetSize` → `Resize`
- `internal/app/navigation_state.go` — Deleted 6 simple accessors, renamed: `SetSelectedIndex` → `SelectAt`, `GetSelectedItem` → `SelectedItem`, `SetMenuItems` → `ReplaceMenu`, `GetKeyHandler` → `ResolveKeyHandler`
- `internal/app/operation_state.go` — Deleted 6 simple accessors, renamed: `SetExitAllowed` → `PermitExit`, `GetWorkflowState` → `WorkflowState`, `GetConsoleState` → `EnsureConsoleState`, `GetInputState` → `InputState`
- `internal/app/console_state.go` — Deleted 4 simple accessors, renamed `GetStateRef` → `ViewState`. Deleted `Clear()` method. `Reset()` now deterministic (buffer + scroll + autoScroll).
- `internal/app/dialog_manager.go` — Renamed: `GetDialogState` → `DialogState`, `GetDialogContext` → `DialogContext`, `GetPickerState` → `PickerState`
- `internal/app/dialog_state.go` — Deleted 6 same-package accessor methods (GetDialog, IsVisible, GetContext, SetContextValue, GetContextValue, SetContext)
- `internal/app/input_state.go` — Deleted 7 simple accessors, renamed: `SetValue` → `ReplaceValue`, `SetCursorPos` → `ClampCursorTo`, `SetPrompt` → `ConfigurePrompt`
- `internal/app/activity_state.go` — Deleted 4 simple accessors
- `internal/app/async_state.go` — Deleted `SetExitAllowed`
- `internal/app/dirty_state.go` — Renamed `SetPhase` → `AdvancePhase`. Added `DirtyPhase` and `DirtyConflictPhase` type aliases with 7 constants.
- `internal/app/environment_state.go` — Deleted 5 simple accessors
- `internal/app/picker_state.go` — Deleted 6 simple accessors
- `internal/app/time_travel_state.go` — Deleted 2 simple accessors
- `internal/app/workflow_state.go` — Deleted 2 simple accessors
- 38 call sites updated to direct field access across 16 files

**Phase 3 — SSOT Consolidation:**
- `internal/app/constants.go` — Added `PasteBurstWindow`, `PageScrollLines`, `InputActionCloneURL`, `BranchPickerPurposeMerge`
- `internal/app/messages_error.go` — Removed dead `invalid_url_format` entry (SSOT moved to ui.Validators)
- `internal/app/messages_state.go` — Added `aborting_rebase`
- `internal/app/messages_menu.go` — `GetFooterMessageText` map extracted to package-level var with `init()`
- `internal/app/app_keys.go` — Extracted `globalHandlers()` method, both callers unified
- `internal/app/operation_steps.go` — Added `OpRebase`, `OpDirtySwitch`, `OpRebaseContinue`, `OpRebaseAbort`, `OpFinalizeMergeFromMenu`, `OpAbortMergeFromMenu`
- `internal/app/app.go`, `internal/app/handlers_global_keys.go` — URL validation delegated to `ui.Validators["url"]` (SSOT)
- 9 files — Console transition consolidated: 16 triple/double patterns + 12 redundant Clear calls collapsed to `Reset()`
- 12 files — Magic strings replaced with constants (`"clone_url"`, `"merge"`, `"rebase"`, `"dirty_switch"`, etc.)

**Phase 4 — Control Flow:**
- `internal/app/handlers_global_menu.go` — Empty if-block inverted. `handleKeyESC` decomposed (144→35 line dispatcher + 6 extracted handlers: `handleEscCopyHashMode`, `handleEscAsyncAbort`, `handleEscPostAbort`, `handleEscInput`, `handleEscTimeTravelConsole`, `handleEscReturnToPrevious`). Menu nav handlers moved here from app.go.
- `internal/app/auto_update.go` — Fixed `a.startAutoUpdate()` (discarded Cmd) → `a.activityState.StartAutoUpdate()` (sets flag)
- `internal/app/conflict_handlers.go` — `handleConflictSpace` early return inverted. `handleConflictEnter` 11-branch if/else→switch. `handleConflictEsc` 6-branch if/else→switch. Added `"rebase"` cases for both.
- `internal/app/app.go` — `updateInputValidation` early return inverted to positive nesting
- `~/.carol/LANGUAGE.md` — Guard definition widened from type-based to topology-based

**Phase 5 — Lean:**
- `internal/app/app.go` — 361→252: Deleted 8 async wrapper forwarders (132 call sites updated to promoted methods via `*OperationState` embedding). Deleted dead `GetFooterHint()`.
- `internal/app/git_handlers.go` — 352→181: Extracted inline case arm bodies to handler functions, moved to `handlers_git_result.go`. Dispatch table now lean (one-liner per case).
- `internal/app/conflict_handlers.go` — 310→278: Trimmed redundant comments
- `internal/app/op_dirty_merge.go` — 308→299: Trimmed
- `internal/app/cache_manager.go` — 307→292: Trimmed

**Phase 6 — Bug Fixes:**
- `internal/app/menu.go` — `GenerateMenu()` dispatch map expanded from 3 to 8 Operation states (no more panic)
- `internal/app/menu_items.go` — Added 4 menu items: `finalize_merge`, `abort_merge`, `rebase_continue`, `rebase_abort`
- `internal/app/menu_render_extra.go` — Added `menuMerging()`, `menuRebasing()`, `menuConflicted()`, `menuDirtyOperation()`
- `internal/app/handlers_conflict.go` — Added `setupConflictResolverForRebase()`
- `internal/app/op_rebase.go` — `cmdRebaseContinue`, `cmdRebaseAbort`, `handleRebaseContinue` (conflict loop), `handleRebaseAbort`, `handleFinalizeMergeFromMenu`
- `internal/app/dispatchers.go` — Registered 4 new action IDs
- `internal/app/dispatch_git_basic.go` — Added 4 dispatch functions
- `internal/app/git_handlers.go` — Added 3 cases: `OpRebaseContinue`, `OpRebaseAbort`, `OpFinalizeMergeFromMenu`
- `internal/app/app_constructor.go` — Startup handles Merging/Rebasing states. `git.CleanStaleLocks()` moved to startup-only. `PermitExit()` replaces direct field access.
- `internal/app/op_pull.go` — Replaced raw `exec.Command` with `git.Execute()`
- `internal/git/execute.go` — Removed per-operation `cleanStaleLocks()` call
- `internal/git/exec_utils.go` — Exported `CleanStaleLocks()`
- `SPEC.md` — Section 4-5 rewritten: mid-operation recovery replaces pre-flight blocking. Added `Rebasing` to Operation table.

**Phase 7 — Type Safety:**
- `internal/app/dirty_state.go` — Added `DirtyPhase`/`DirtyConflictPhase` type aliases with 7 constants, replaced raw strings across 5 files
- `internal/ui/menu_item.go` — `MenuItem` struct defined in ui package
- `internal/app/menu.go` — `MenuItem` is now `type MenuItem = ui.MenuItem`
- `internal/ui/menu.go` — `RenderMenuWithHeight`/`RenderMenuWithBanner` accept `[]MenuItem` directly (was `interface{}`)
- `internal/app/app_view_header.go` — Deleted `menuItemsToMaps` bridge function
- `internal/app/app_view_main.go` — All calls pass `[]MenuItem` directly

**Phase 8 — Documentation:**
- `CODEBASE-MAP.md` — 12+ stale filenames updated, new files added
- `~/.carol/MANIFESTO.md` — LANGUAGE.md reference added

### Alignment Check
- [x] BLESSED principles followed
- [x] LANGUAGE.md Go addendum created and applied
- [x] NAMES.md adhered (all 17 renames follow verb-noun/semantic naming)
- [x] MANIFESTO.md principles applied
- [x] SPEC.md updated for mid-operation recovery

### Problems Solved
- `GenerateMenu()` panicked on 5 of 8 Operation states — all states now have generators
- Rebase conflicts had zero handling — full resolver loop implemented
- SPEC incorrectly blocked recoverable states at startup — rewritten to recover all
- 48 dead same-package accessors violated LANGUAGE.md E (Encapsulation) — deleted
- 8 async forwarding wrappers violated LANGUAGE.md E — deleted, 132 call sites promoted
- `ConsoleState.Clear()` was shadow of `Reset()` (S/SSOT violation) — deleted, `Reset()` deterministic
- `handleKeyESC` was 144-line cascade (L violation) — decomposed to 35-line dispatcher
- `cleanStaleLocks` ran before every git operation (TOCTOU race) — moved to startup-only
- `exec.Command` bypassed git infrastructure — replaced with `git.Execute()`
- `map[string]interface{}` at UI boundary lost type safety — `MenuItem` moved to ui package
- MANIFESTO.md had no language-specific guidance — `LANGUAGE.md` created with Go/C++ sections

### Technical Debt / Follow-up
- W5: `KeyHandler` type signature `func(a *Application) handler(app *Application)` has redundant receiver+param. Systemic across all handlers — framework convention, accepted.
- C2: Zero test coverage. Orthogonal to this refactor. Needs separate test strategy sprint.
- 5.8: 30-line function audit not completed across all files. Priority files done (handleKeyESC, conflict handlers). Full sweep is separate scope.

**Status:** ✅ Build passes. Auditor final gate: 10/10 PASS.

---

## Sprint 9: MSYS2 ARM64 Architecture Detection in build.sh ✅

**Date:** 2026-04-03
**Duration:** ~15 min

### Agents Participated
- **COUNSELOR** — Requirements counseling, investigation, trivial fix
- **Pathfinder** — Discovered build.sh architecture detection logic

### Files Modified (1 total)
- `build.sh:5-14` — Architecture detection now checks `$MSYSTEM` first (authoritative on MSYS2), falls back to `uname -m` for non-MSYS2 systems (macOS, Linux). `CLANGARM64` maps to arm64, `MINGW64`/`UCRT64`/`MSYS` maps to x64.

### Alignment Check
- [x] BLESSED principles followed (SSOT: $MSYSTEM is the single source of truth for MSYS2 arch)
- [x] No early returns
- [x] Fail-fast error handling

### Problems Solved
- `uname -m` always reports `x86_64` on MSYS2 regardless of actual architecture. On ARM64 Windows (UTM on Apple Silicon), build.sh produced `tit_x64` instead of `tit_arm64`. Fixed by checking `$MSYSTEM` environment variable first — `CLANGARM64` correctly identifies ARM64 MSYS2 environment.

### Technical Debt / Follow-up
- `build.sh` is Unix-only (bash, ln). No Windows-native build path exists.

**Status:** ✅ Build passes

---

## Sprint 8: Contextual Conflict Resolver Labels ✅

**Date:** 2026-03-31
**Duration:** ~20 min

### Agents Participated
- **COUNSELOR** — Requirements, label mapping, delegation
- **Pathfinder** — Discovered all `setupConflictResolver` call sites and available branch context
- **Engineer** — Implementation

### Files Modified (3 total)
- `internal/app/handlers_conflict.go` — Updated `setupConflictResolverForBranchSwitch`, `setupConflictResolverForBranchMerge`, `setupConflictResolverForDirtyMerge`, `setupConflictResolverForDirtySwitch` with contextual branch-name labels instead of generic LOCAL/REMOTE/yours/theirs
- `internal/app/confirm_branch.go` — Set `OriginalBranch` in dirty switch initiators (both stash and discard paths)
- `internal/app/confirm_merge.go` — Set `OriginalBranch` in dirty merge initiator

### Alignment Check
- [x] LIFESTAR principles followed (Findable: labels tell you exactly what each column represents)
- [x] NAMING-CONVENTION.md adhered
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (LOVE Empathizes: user sees branch names, not git jargon)

### Problems Solved
- Conflict resolver columns showed generic `LOCAL (yours)` / `REMOTE (theirs)` for all operations. Now shows actual branch names with context: `main (current)` / `dev (incoming)`, `main (current)` / `dev (stashed)`, etc.
- Pull/push sync labels unchanged — remote ops are already clear

### Technical Debt / Follow-up
- Startup conflict resolver (`app_constructor.go`) still uses generic labels — no branch context available at startup
- Time travel labels use commit hash/date format — separate convention, left as-is

**Status:** ✅ Build passes

---

## Sprint 7: Branch Operations — New Branch, Merge From, Dirty Switch/Merge Protocol ✅

**Date:** 2026-03-31
**Duration:** ~6 hours

### Agents Participated
- **COUNSELOR** — Requirements counseling, plan, delegation, bug investigation, root cause analysis
- **Pathfinder** — Codebase discovery (branch switch flow, dirty ops protocol, conflict detection, DetectState short-circuit)
- **Engineer** — Implementation (new branch, merge from, dirty switch protocol, dirty merge protocol, conflict detection fix)
- **Auditor** — Verified Feature 1 (2 issues caught: PreviousMode, SSOT strings), Feature 2 (2 issues caught: BranchPickerPurpose leak, dirty merge conflict gap), Dirty merge protocol (2 issues caught: finalize constant, conflictResolveState leak)

### Files Modified (21 tracked + 4 new = 25 total)
- `internal/git/state_detection.go` — Added `HasConflicts()` — direct conflict check bypassing `DetectState` DirtyOperation short-circuit
- `internal/app/app.go` — `checkForConflicts` now uses `git.HasConflicts()` instead of `reloadGitState()`
- `internal/app/menu_items.go` — Added `config_new_branch`, `config_merge_branch` SSOT entries
- `internal/app/menu_render_extra.go` — Added both items to `GenerateConfigMenu()`, merge conditional on 2+ branches
- `internal/app/dispatchers.go` — Registered `config_new_branch`, `config_merge_branch` dispatchers
- `internal/app/dispatch_dialog.go` — Added `dispatchConfigNewBranch`, `dispatchConfigMergeBranch`, cleared `BranchPickerPurpose` in switch dispatcher
- `internal/app/handlers_config_branch.go` — Added merge purpose routing in `handleBranchPickerEnter`
- `internal/app/handlers_git_branch.go` — Added `handleNewBranchNameSubmit`, `cmdCreateBranch`; removed monolithic `cmdBranchSwitchWithStash`
- `internal/app/workflow_state.go` — Added `BranchPickerPurpose` field
- `internal/app/confirm_dialog.go` — Added `ConfirmMergeBranch`, `ConfirmMergeBranchDirty` types
- `internal/app/confirm_dialog_handlers.go` — Registered merge confirmation handler pairs
- `internal/app/confirm_branch.go` — Rewrote dirty switch handlers to use Dirty Operation Protocol
- `internal/app/confirm_merge.go` — NEW: merge flow handlers, `cmdMergeBranch`, `cmdFinalizeBranchMerge`
- `internal/app/messages_dialog.go` — Added merge branch confirmation messages
- `internal/app/messages_error.go` — Added `branch_name_invalid`, `branch_already_exists`, `merge_branch_failed`
- `internal/app/messages_state.go` — Added dirty switch and dirty merge output messages
- `internal/app/operation_steps.go` — Added 13 new operation step constants (branch create, merge, dirty merge, dirty switch)
- `internal/app/dirty_state.go` — Added `MergeBranch`, `TargetBranch` fields
- `internal/app/op_dirty_merge.go` — NEW: full Dirty Operation Protocol for merge (6 phase functions)
- `internal/app/op_dirty_switch.go` — NEW: full Dirty Operation Protocol for branch switch (6 phase functions)
- `internal/app/git_handlers.go` — Added routing for all dirty merge and dirty switch phases; fixed `conflictResolveState` leak in `OpDirtyPullFinalize`
- `internal/app/handlers_pull.go` — Added `handleMergeBranchResult`, `handleFinalizeBranchMerge`
- `internal/app/handlers_conflict.go` — Added `setupConflictResolverForBranchMerge`, `setupConflictResolverForDirtyMerge`, `setupConflictResolverForDirtySwitch`
- `internal/app/conflict_handlers.go` — Added finalize/abort routing for merge, dirty merge, dirty switch
- `internal/app/op_dirty_pull_snapshot.go` — Fixed `conflictResolveState` leak in `OpDirtyPullFinalize`

### Alignment Check
- [x] LIFESTAR principles followed (SSOT: all strings in message maps, Lean: reused existing dirty op protocol)
- [x] NAMING-CONVENTION.md adhered
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (Explicit Encapsulation: dirty ops are source-agnostic)
- [ ] No early returns — pre-existing pattern in input validation handlers; not introduced by this sprint

### Problems Solved
- **New Branch**: Config menu item `n` — creates and switches to new branch from current HEAD
- **Merge From**: Config menu item `m` (visible when 2+ branches) — merge another branch into current with full confirmation flow
- **Dirty Switch Protocol**: Decomposed monolithic `cmdBranchSwitchWithStash` into 4-phase Dirty Operation Protocol identical to dirty pull (snapshot → switch → apply snapshot → finalize)
- **Dirty Merge Protocol**: Full Dirty Operation Protocol for merge with dirty tree (snapshot → merge → apply snapshot → finalize)
- **Conflict Detection Bug**: `checkForConflicts` failed during ALL dirty operations because `DetectState()` short-circuits to `DirtyOperation` when `.git/TIT_DIRTY_OP` exists, never reaching `detectOperation()` which checks for `u ` conflict lines. Fixed by adding `git.HasConflicts()` which checks index directly
- **`conflictResolveState` leak**: Pre-existing bug in `OpDirtyPullFinalize` — stale conflict state not cleared. Fixed in pull, merge, and switch finalize paths
- **`GenerateMenu` panic on Conflicted state**: Discovered via panic trace — `GenerateMenu()` doesn't handle `Conflicted` operation state, panics when ESC pressed from console after failed stash apply (bug 2, see tech debt)

### Technical Debt / Follow-up
- **Bug 2 (not fixed)**: `GenerateMenu()` panics with "Unknown git operation state: Conflicted" when git state is `Conflicted` and menu regeneration is triggered. Discovered during dirty switch testing. Needs separate fix.
- Redundant console messages in dirty pull and dirty merge flows (only dirty switch was cleaned up)
- `PLAN-branch-operations.md` in project root — can be deleted after commit

**Status:** ✅ Build passes

---

## Sprint 6: Copy Hash Mode Spacebar Page Cycling ✅

**Date:** 2026-03-30
**Duration:** ~30 min

### Agents Participated
- **COUNSELOR** — Requirements, root cause analysis, plan, delegation
- **Pathfinder** — Codebase discovery (copy-hash implementation, listpane, scroll, render pipeline)
- **Engineer** — Implementation (spacebar handler, footer hint, page derivation fix)
- **Auditor** — Verified both rounds (initial + fix), build clean

### Files Modified (3 total)
- `internal/app/handlers_history_copyhash.go:68-76` — Added spacebar handler in handleHistoryCopyHashKeypress: advances SelectedIdx by CopyHashMaxVisible, wraps to 0
- `internal/app/handlers_history_copyhash.go:84-85` — Replaced ScrollOffset-based key computation with pageStart derivation from SelectedIdx (matches renderer)
- `internal/ui/history.go:177-178` — renderHistoryListPane now derives pageStart from SelectedIdx instead of ScrollOffset for ComputeCopyHashKeys
- `internal/app/messages_menu.go:75` — Added Space hint to history_copyhash footer

### Alignment Check
- [x] LIFESTAR principles followed (Lean: no new state fields, SSOT: pageStart derivation identical in render + handler)
- [x] NAMING-CONVENTION.md adhered (pageStart, commitCount, nextIdx)
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (Explicit Encapsulation: UI computes page, app handles input)
- [x] No early returns

### Problems Solved
- Spacebar page cycling in CopyHashMode: press Space to advance labels to next 10 commits, wraps to top
- Bug: initial implementation moved SelectedIdx but labels didn't update because ComputeCopyHashKeys used ScrollOffset (stays 0 when terminal is tall enough). Fixed by deriving page boundary from SelectedIdx via integer division

### Technical Debt / Follow-up
- None

**Status:** ✅ Build passes

---

## Sprint 5: Fix Stale Time Travel Marker Misdetection ✅

**Date:** 2026-03-29
**Duration:** ~30 min

### Agents Participated
- **COUNSELOR** — Root cause analysis, plan, contract alignment, delegation
- **Pathfinder** — Codebase discovery (state detection flow, menu generation, header rendering)
- **Engineer** — Investigated end repo state, implemented fix
- **Auditor** — Verified fix, caught SSOT violation (inline os.Remove vs ClearTimeTravelInfo)

### Files Modified (1 total)
- `internal/git/state_detection.go:171-182` — Priority 2 block in detectOperation() now cross-validates TIT_TIME_TRAVEL marker against actual HEAD state via symbolic-ref; stale markers cleaned up via ClearTimeTravelInfo() (SSOT)

### Alignment Check
- [x] LIFESTAR principles followed
- [x] NAMING-CONVENTION.md adhered
- [x] ARCHITECTURAL-MANIFESTO.md principles applied (SSOT: uses existing ClearTimeTravelInfo instead of inline os.Remove)
- [ ] No early returns used — pre-existing in detectOperation(); not introduced by this fix

### Problems Solved
- Stale .git/TIT_TIME_TRAVEL marker (left from interrupted time travel return) caused detectOperation() to return TimeTraveling even when HEAD was on a branch, displaying normal repos as "DETACHED @ hash" with TimeTraveling menu

### Technical Debt / Follow-up
- detectOperation() uses early returns throughout (pre-existing) — full refactor to positive checks would be a separate effort
- No test coverage for stale marker scenario

**Status:** ✅ Build passes

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

## Sprint 4: Copy Hash Mode in History Panel ✅

**Date:** 2026-03-24
**Duration:** ~1 hour

### Agents Participated
- **COUNSELOR** — Requirements, plan, contract alignment, delegation
- **Pathfinder** — History panel rendering, key dispatch, listpane, footer patterns
- **Engineer** — Implementation

### Files Modified (14 total)
- `internal/ui/theme.go:70-71` — Added CopyHashLabelForeground/Background to Theme struct
- `internal/ui/theme_gfx.go:72-74` — Added copyHashLabel colors to GFX (default) theme
- `internal/ui/theme_seasons.go` — Added copyHashLabel colors to all 4 seasonal themes
- `internal/ui/theme_loading.go:81-82,172-173` — TOML mapping for new theme fields
- `internal/ui/history.go:21-22` — Added CopyHashMode/CopyHashFull to HistoryState
- `internal/ui/history.go:24-89` — NEW: CopyHashKey struct, ComputeCopyHashKeys algorithm, constants
- `internal/ui/history.go:130-158` — buildCommitListItems accepts copyHashKeys, disables selection in CopyHashMode
- `internal/ui/listpane.go:28-31` — Added CopyHashChar/CharPos/Fg/Bg to ListItem
- `internal/ui/listpane.go:261-276` — renderItem highlights flash char with theme fg/bg
- `internal/ui/filehistory.go` — Updated buildCommitListItems call with nil keys
- `internal/app/handlers_history_copyhash.go:1-98` — NEW: enter, enter-full, esc, keypress handlers
- `internal/app/app_keys.go:80-81,83` — Registered y, Y, rewired esc to CopyHashMode-aware handler
- `internal/app/app_update_msg.go:109-112` — CopyHashMode key intercept before normal dispatch
- `internal/app/handlers_global_menu.go:17-21` — CopyHashMode ESC check in global handleKeyESC
- `internal/app/messages_menu.go` — Added y shortcut to history_list, new history_copyhash entry
- `internal/app/footer.go:69-71` — CopyHashMode footer hint key routing
- `ARCHITECTURE.md` — Documented Copy Hash Mode section

### Alignment Check
- [x] LIFESTAR principles followed (Lean: sub-state not new AppMode, SSOT: theme colors, Explicit Encapsulation: ui computes keys, app handles clipboard)
- [x] NAMING-CONVENTION.md adhered (CopyHashMode, CopyHashKey, ComputeCopyHashKeys, CopyHashFull)
- [x] ARCHITECTURAL-MANIFESTO.md principles applied
- [x] No early returns
- [x] Fail-fast error handling

### Problems Solved
- Copy commit hash from History panel via flash/jump-style unique char labels
- y = short hash (7 chars, for communication), Y = full hash (40 chars, hidden power feature)
- Selection highlight and navigation disabled during mode — clean modal UX

### Bugs Fixed During Sprint
- GFX (default) theme missing copyHashLabel fields — labels invisible
- ESC blocked by CopyHashMode key intercept (len("esc")==3 returned true) — fixed to pass through
- Global ESC handler overrides mode-specific handler — added CopyHashMode check to handleKeyESC

### Technical Debt / Follow-up
- handleHistoryCopyHashKeypress recomputes ListPane scroll state to match renderer — coupling between key handler and render logic

**Status:** ✅ APPROVED — Build clean

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
