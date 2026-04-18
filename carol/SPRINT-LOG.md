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

## Sprint 15: C++/JUCE Port Scaffold (Phase 0) ✅

**Date:** 2026-04-18
**Duration:** ~1 day (multi-session)

### Agents Participated
- **COUNSELOR** — `/goplan` against RFC.md (BRAINSTORMER handoff), PLAN-tit-cpp-port.md authoring, 6-step Sprint 1 orchestration, mid-sprint PLAN amendments (13+ decisions locked), NAMES Rule -1 gating, CONTRACT-vs-PLAN reconciliation (byte-identity → visual parity, -Wall drop to END precedent, SCREAMING_SNAKE constant convention, jreng_markdown-stays/jreng_svg_braille-absorbed module topology, paint-nullify for RFC §6.1)
- **Pathfinder** — caroline rename Sprint 5 verification (confirmed `jreng::Terminal` → `jreng::tui` shipped; prereq cleared); jreng_core SVG capability audit; jreng_tui dependency graph; tit project root enumeration; END CMake pattern extraction; FigBug/FileSystemWatcher vs Gin scope probes
- **Librarian** — JUCE 8 file-watch primitive research (no `juce::FileWatcher`; canonical is Timer + mtime poll); FigBug/Gin `FileSystemWatcher` verification (`ReadDirectoryChangesW` + per-file callback); `juce::JUCEApplicationBase` vs `juce::JUCEApplication` module boundary; `juce::Drawable::parseSVGPath` + `juce::Path::Iterator` as SVG path SSOT
- **Engineer** — all code writes across 6 steps + 2 resolution passes (Step 1.2 `-Wall` revert + juce_gui_basics addition; Step 1.4 five-point TIT-side fixes; Step 1.5 four-finding resolution + SCREAMING_SNAKE rename sweep; Step 1.6 F1/F2/F3 + sampleBraillePixelAt decomposition)
- **Auditor** — validation pass after every step (1.1 / 1.2 / 1.2-redo / 1.3 / 1.3b / 1.4 / 1.5 / 1.5-redo / 1.6 / 1.6-redo); flagged four audit findings on jreng_subprocess (L, E, workingDir silent-drop, jassert absence); flagged three findings on jreng_braille (L, E, SSOT); verified BLESSED with MANIFESTO L "smell-detector" framing on 2×4 braille cell natural shape

### Files Modified

**Project-level (authored this sprint):**
- `PLAN-tit-cpp-port.md` — new 400+ LOC master plan covering 6 sprints (Phase 0 through Windows parity). Amended repeatedly with locked decisions section + resolved-decisions audit trail
- `.gitignore` — new root C++/CMake ignore set; later fixed to remove `carol/` exclusion (SPRINT-LOG/DEBT now tracked across dev machines)
- `CMakeLists.txt` — ~360 LOC JUCE CMake scaffold; 4-probe JUCE discovery mirroring END; `juce_add_console_app` for `titc`; universal x86_64+arm64; C++17; END suppression-only flags (`-Wno-unused-parameter -Wno-sign-conversion -Wno-shadow`); modules wired in dep order: jreng_core → jreng_data_structures → jreng_markdown → jreng_tui → jreng_subprocess
- `Source/Main.cpp` — 3 LOC, `START_JUCE_APPLICATION (TitApp)`
- `Source/TitApp.h` — 19 LOC empty-shell `JUCEApplication` subclass
- `Source/TitApp.cpp` — 30 LOC, `initialise` → `quit()` stub (removed Sprint 3)

**Archived to `___legacy___/` (Step 1.1, 16 entries):**
- `cmd/`, `internal/`, `scripts/`, `screenshot/`, `dist/`, `go.mod`, `go.sum`, `build.sh`, `release.sh`, `.goreleaser.yaml`, `entitlements.plist`, `ARCHITECTURE.md`, `CODEBASE-MAP.md`, `RELEASE_NOTES.md`, `README.md`, `LICENSE`, `.gitignore` (legacy Go)
- Go TIT still builds via `cd ___legacy___ && go build ./cmd/tit`

**Forks from caroline (Step 1.3 / 1.4 — byte-identical at fork point):**
- `modules/jreng_core/` (62 files)
- `modules/jreng_data_structures/` (7 files)
- `modules/jreng_tui/` (full module post-rename; namespace `jreng::tui`, `paint()` method)
- `modules/jreng_markdown/` (co-forked; jreng_tui declares it as dep)

**TIT-forge extensions to forked modules (Step 1.3b + 1.4 fixes + 1.6):**
- `modules/jreng_core/file/jreng_file_watcher.h` — 220 LOC; `struct jreng::File::Watcher` nested; BSD attribution to Roland Rabien/FigBug/Gin preserved
- `modules/jreng_core/file/jreng_file_watcher.cpp` — 413 LOC; FSEvents/inotify/ReadDirectoryChangesW backends; `std::map::contains` → C++17-compatible `find() == end()`
- `modules/jreng_core/file/LICENSE_Gin.txt` — 29 LOC BSD attribution
- `modules/jreng_core/jreng_core.mm` — 1 LOC ObjC++ shim for FSEvents unit
- `modules/jreng_core/jreng_core.h` — `OSXFrameworks: CoreServices` added; `jreng_file_watcher.h` aggregated
- `modules/jreng_core/jreng_core.cpp` — aggregates `jreng_file_watcher.cpp` via unity include
- `modules/jreng_core/file/jreng_file.h` — `struct Watcher;` forward decl inside `struct File`
- `modules/jreng_tui/jreng_tui.h` — include order fix (`jreng_key_event.h` before `jreng_textbox.h`); added `braille/jreng_braille_grid.h` aggregation
- `modules/jreng_tui/jreng_tui.cpp` — added `braille/jreng_braille_grid.cpp` unity include
- `modules/jreng_tui/graphics/jreng_tui_rectangle.h:69,77` — removed `constexpr` from methods calling non-`constexpr` JUCE API
- `modules/jreng_tui/ansi/jreng_ansi_screen.cpp:197-199,206-209` — `juce::Graphics` rvalue → named local `clipped` binding fix
- `modules/jreng_tui/ansi/jreng_ansi_graphics.h:86` — `juce::Font` → `juce::Font { juce::FontOptions{} }` (JUCE 8 deprecation migration)
- `modules/jreng_tui/markdown/jreng_ansi_markdown_renderer.cpp:94` — same deprecation migration with `withHeight(14.0f)`
- `modules/jreng_tui/ansi/jreng_ansi_component.h:24` — explicit `void paint (juce::Graphics&) override final {}` nullify (closes RFC §6.1 concern that caroline's Sprint 5 rename only masked)
- `modules/jreng_tui/braille/jreng_braille_grid.h` — 199 LOC; `namespace jreng::braille` flat sibling to `jreng::tui`
- `modules/jreng_tui/braille/jreng_braille_grid.cpp` — 512 LOC after full decomposition pass; 14 functions including approved helpers `appendEdgeIntersection`, `collectSubpathStartPoints`, `paintScanlineRanges`, `processSvgDrawablePath`, `encodeBrailleCellAt`, `sampleBraillePixelAt`; `BraillePixelSample` POD; `computeSubpathBounds` SSOT-called (no inline duplicate)

**TIT-authored from scratch (Step 1.5):**
- `modules/jreng_subprocess/jreng_subprocess.h` — 19 LOC JUCE module declaration
- `modules/jreng_subprocess/jreng_subprocess.cpp` — 2 LOC unity include
- `modules/jreng_subprocess/subprocess/jreng_subprocess_subprocess.h` — 128 LOC `Subprocess` class + nested `Handler::Completion` / `Handler::Chunk` typedefs; SCREAMING_SNAKE constants (`BYTE_CAP`, `TRUNCATION_NOTICE`, `ENV_TERMINAL_PROMPT`, `ENV_PROGRESS_DELAY`)
- `modules/jreng_subprocess/subprocess/jreng_subprocess_subprocess.cpp` — 239 LOC post-decomposition; `Worker` private thread class owning `juce::ChildProcess`; `buildCommand` prepends `env -C <workingDir>` to argv (POSIX substitute for juce::ChildProcess's missing cwd arg); `readStream`/`appendWithCap`/`computeIsReplace`/`processChunk` helpers; `jassert` on invariants

**Fixture generator (not project code):**
- `___legacy___/braille_ref/main.go` — 42 LOC Go reference driver for visual-parity spot-check

### Alignment Check
- [x] BLESSED — MANIFESTO fully applied; L treated as smell-detector per MANIFESTO text ("natural shape" of 2×4 braille cell iteration in `encodeBrailleCellAt`); all early returns eliminated; SSOT preserved across module forks; thread ownership clean (worker owns ChildProcess; main owns Worker; main owns Subprocess)
- [x] NAMES — Rule -1 honored; all new identifiers ARCHITECT-approved pre-write (~35 names across sprint); SCREAMING_SNAKE locked as project constant convention mid-sprint
- [x] JRENG-CODING-STANDARD — braces on new line; `override` without `virtual`; brace init; `nullptr`; alternative tokens `not`/`and`/`or`; `.at()` for container access; no anonymous namespaces; no `namespace detail`; pass-by-value for small types
- [x] MANIFESTO Core Mantra ("NEVER OVERDO IT") + Contract Addition ("use JUCE/jreng as much as we can") — JUCE primitives (`Drawable::parseSVGPath`, `Path::Iterator`, `ChildProcess`, `Thread`, `Font`+`FontOptions`) consumed maximally; Gin fork reused for file-watching instead of hand-rolling; visual-parity locked over manual bezier(20) reimplementation
- [x] RFC §4.1 scaffold layout matches project tree; RFC §3.2 `tui::Component : public juce::Component` Path A locked; RFC §6.1 paint-collision resolved (not just renamed)

### Problems Solved
- **Go toolchain stack fracture** (RFC §1) — C++/JUCE scaffold stands up alongside preserved Go build; `titc` binary is the forward artifact
- **Caroline never had a downstream consumer** — TIT-cpp is the forge; 5 latent caroline bugs surfaced and fixed on first real consumption (include ordering, non-constexpr JUCE calls, rvalue binding, Font deprecation, paint-overload-shadow vs true override)
- **RFC §6.1 paint() collision** — caroline Sprint 5 rename was incomplete; explicit `override final {}` nullify closes the hide
- **`-Wall -Wextra` vs END precedent** — PLAN initially required `-Wall -Wextra`; END uses suppression-only; PLAN amended to inherit END's proven flag set
- **jreng_svg_braille module boundary** — PLAN proposed separate module; ARCHITECT clarified "renderers live inside jreng_tui, parsers stay outside"; braille absorbed into `modules/jreng_tui/braille/` with `namespace jreng::braille` flat at `jreng::` (file placement ≠ namespace hierarchy)
- **`juce::ChildProcess` missing cwd arg** — `env -C <workingDir>` POSIX prefix is the minimal CONTRACT-aligned fix (no shell escaping, no juce fork, no fork/exec bypass)
- **SVG tessellation engine mismatch** — JUCE's `Path::Iterator` yields different pixels than Go's `approximateCubicBezier(20)`; byte-identity requirement retired in favor of visual parity; JUCE becomes path-raster SSOT per CONTRACT
- **SCREAMING_SNAKE vs camelCase constant convention** — one locked project-wide; Subprocess constants retroactively renamed (`byteCap` → `BYTE_CAP`, etc., + `READ_CHUNK_SIZE`)
- **carol/ gitignored from legacy policy** — fixed mid-sprint; CAROL sprint-log + debt ledger now travel with the repo
- **jreng_markdown as jreng_tui dep** — RFC/PLAN omission discovered at Step 1.4; co-forked alongside jreng_tui (markdown parser stays separate module per parser-vs-renderer architectural rule)

### Debts Paid
- None (no `DEBT.md` entries paid this sprint — none existed at sprint start)

### Debts Deferred
- Windows parity (`CreateProcess` in `jreng_subprocess`, `ReadDirectoryChangesW` in `jreng::File::Watcher`, path normalization) — PLAN-locked to dedicated Sprint 6
- stdout/stderr stream separation in `jreng_subprocess` — JUCE `ChildProcess` merges both streams; `Handler::Completion::stderrCapture` signature reserved; platform-native pipe impl deferred to Step 4.2 (`GitRunner`) if required
- Caroline drift — all 5 TIT-forge jreng_tui fixes + extensions (Watcher, braille, Subprocess, svg_braille→absorbed, Font migration) eventually contribute back per PLAN Step 2.5 handoff pattern; caroline is downstream now, no urgency
- Apple Developer ID code-signing + notarization for `titc` — PLAN Step 5.3 scope (rewrite from scratch, not port `___legacy___/scripts/post-build.sh`)
- `jreng_ansi_markdown_renderer.cpp:94` magic `14.0f` font height — flagged pre-existing (forked from caroline), low-priority; scoped to unused-by-TIT markdown renderer

---

## Sprint 14: Startup Spinner Animation + GOBIN Install ✅

**Date:** 2026-04-15
**Duration:** ~00:45

### Agents Participated
- **COUNSELOR** — Plan, CONTRACT re-verification (BLESSED, JRENG-CODING-STANDARD, NAMES), SSOT mapping (reuse activityState frame / theme.SpinnerColor / CacheRefreshInterval / ui.GetSpinnerFrame), direct edit on build.sh, git-repo diagnostic (local-behind-origin investigation)
- **Pathfinder** — Startup remote-gate locus (`app_init.go:123-124`, `app_constructor.go:206-210`, `app_view_main.go:169`, `app_update_msg.go:206-223`), tea.Tick pattern survey (auto_update.go animation pump as template), Theme.SpinnerColor + menu.go usage sites, cake/`internal/ui/console.go:272-275` reference for themed spinner rendering, read-only git inspection of tit repo (`rev-list --count HEAD...@{upstream}` = 0/1)
- **Engineer** — Implemented StartupSpinnerMsg + schedule/handle pair mirroring AutoUpdateAnimationMsg pattern; clean `go build ./...`

### Files Modified (7 total)
- `build.sh` — rewrote: `go install` to `$(go env GOBIN)` with `$(go env GOPATH)/bin` fallback; dropped ARCH_SUFFIX detection, `$HOME/.tit/bin` install root, and `$HOME/.local/bin` symlink. Version stamping via `-ldflags -X` preserved
- `internal/app/messages.go:70-71` — added `StartupSpinnerMsg struct{}` (noun form per NAMES Rule 1, consistent with `AutoUpdateAnimationMsg`)
- `internal/app/auto_update.go:148-168` — added `scheduleStartupSpinner() tea.Cmd` (100ms `tea.Tick` → `StartupSpinnerMsg`) and `handleStartupSpinner()` (if `mode == ModeStartup`: `IncrementFrame` + reschedule; else stop). Direct clone of `scheduleAutoUpdateAnimation`/`handleAutoUpdateAnimation`
- `internal/app/app_update_msg.go:242-244` — added `case StartupSpinnerMsg` dispatching to `handleStartupSpinner()`, placed adjacent to `AutoUpdateAnimationMsg` case
- `internal/app/app_init.go:124` — batch `a.scheduleStartupSpinner()` alongside `cmdFetchRemote()` inside existing `HasRemote` gate
- `internal/app/app_view_main.go:166-174` — replaced static `"Checking remote..."` with `lipgloss`-styled braille frame (`theme.SpinnerColor`) + `" Checking Remote..."`; added `lipgloss` import
- `internal/app/app_constructor.go:210` — footer hint casing `"Checking remote..."` → `"Checking Remote..."` for consistency with content text

### Alignment Check
- [x] BLESSED — SSOT (reused `activityState.autoUpdateFrame` — ModeStartup and auto-update spinner mutually exclusive in time, single frame counter is correct), Stateless (no new state, no shadow flag), Explicit (typed message, semantic names), Encapsulation (handler mirrors established pattern — no new primitive), Bound (pump stops naturally when `RemoteFetchMsg` transitions mode out of ModeStartup)
- [x] NAMES — Rule 1 (noun `StartupSpinnerMsg`, verbs `scheduleStartupSpinner`/`handleStartupSpinner`), Rule 3 (semantic — "startup" describes role, "spinner" describes mechanism), Rule 5 (consistent with `AutoUpdateAnimationMsg`/`scheduleAutoUpdateAnimation`/`handleAutoUpdateAnimation` sibling pattern)
- [x] JRENG-CODING-STANDARD — positive nested check in handler, no early returns added; established `tea.Tick` / reschedule pattern preserved
- [x] No-Reinvention — used existing `ui.GetSpinnerFrame`, `theme.SpinnerColor`, `CacheRefreshInterval`, `activityState.IncrementFrame`; zero new helpers or constants

### Problems Solved
- **Static "Checking remote..." string during startup gate felt dead** — replaced with animated braille spinner colored from theme, matching cake's spinner pattern (`internal/ui/console.go:272-275`) and tit's own cache-build spinner (`internal/ui/menu.go:97,111,127`)
- **build.sh arch-suffix + symlink indirection redundant under Go toolchain** — `go install` already writes to `$(go env GOBIN)` and produces an arch-native binary named from the package path; removed bespoke install root, symlink, and arch detection

### Debts Paid
- None (no `DEBT.md` present at project root)

### Debts Deferred
- Local `main` is 1 commit behind `origin/main` (`a858327 v0.0.3` — mass doc-deletion from cross-machine sync). ARCHITECT to rebase manually before committing this sprint's changes
- Engineer noted spec referenced `a.activityState.Frame` but actual field is lowercase `autoUpdateFrame` (same package, unexported access valid); field naming is auto-update-specific even though now used by startup spinner too — cosmetic only, no functional impact

---

## Sprint 13: Version Fallback + Startup Remote Gate ✅

**Date:** 2026-04-15
**Duration:** ~01:30

### Agents Participated
- **COUNSELOR** — Problem framing, Go module proxy/immutability reasoning, CONTRACT re-verification (BLESSED, JRENG-CODING-STANDARD, NAMES), direct edits on approved plan
- **Librarian** — Go `debug.ReadBuildInfo()` semantics under `go install @version`, `-X` ldflags constant-initializer constraint (Go issue #64246), canonical dual-path pattern; Go module proxy immutability ground truth (`proxy.golang.org` no author-removal mechanism, `retract` signals only, `sum.golang.org` append-only)
- **Pathfinder** — Remote-status flow survey (`detectTimeline` `state_detection.go:52`, `cmdFetchRemote` `handlers_console_ui.go:63`, async fetch fired from `Init`), mode system survey (17 modes, `NavigationState.mode`, dispatch via switch + key handler map)
- **Engineer** — Local build/ldflags verification on TIT (`v0.0.3-test` injection) and CAKE (same), post-ModeStartup `go vet` + `go build` pass confirmation

### Files Modified (6 total, TIT + 1 on CAKE)
- `internal/constants.go:12-28` — replaced `var AppVersion = getVersion()` (function-call initializer that silently disabled `-X` ldflags per Go #64246) with `var AppVersion = "dev"` + `init()` fallback using `debug.ReadBuildInfo`; unified goreleaser and `go install @version` paths
- `internal/app/modes.go:39,186-191` — added `ModeStartup` enum constant and `modeDescriptions` entry (`AcceptsInput:false`, `IsAsync:true`); noun-form name per NAMES Rule 1, ARCHITECT-approved per Rule -1
- `internal/app/app_keys.go:147-149` — registered empty handler map for `ModeStartup`; existing global-merge loop wires `q`/`ctrl+c`/`esc`, no action dispatch surface
- `internal/app/app_view_main.go:165-168` — added `case ModeStartup:` rendering inline "Checking remote..." placeholder
- `internal/app/app_constructor.go:202-225` — introduced `shouldGateStartup := isRepo && app.gitState != nil && app.gitState.Remote == git.HasRemote`; on true, set `mode = ModeStartup`, footer hint "Checking remote...", skip menu gen/ReplaceMenu/rebuildMenuShortcuts; cache loading start preserved independently
- `internal/app/app_update_msg.go:206-230` — rewrote `RemoteFetchMsg` handler: on `mode == ModeStartup`, transition to `ModeMenu` regardless of `msg.Success` (never stranded), build menu from fresh timeline, rebuild shortcuts, set first-item footer hint; on failure, override footer with "Remote unreachable — showing cached state"
- `../cake/internal/constants.go:1-23` (cross-project) — same version-fallback pattern applied to CAKE (sibling Go CLI with identical goreleaser + `go install` release path)

### Alignment Check
- [x] BLESSED — Stateless (typed mode state, not a manual boolean flag), Encapsulation (mode orchestrates dispatch gating — "tell, don't ask"), Explicit (semantic `ModeStartup` name, constant string initializer for AppVersion), SSOT (single `modeDescriptions` registry extended, no shadow startup flag)
- [x] NAMES — Rule -1 (ARCHITECT approved `ModeStartup`), Rule 1 (noun form), Rule 3 (semantic — name describes role, not mechanism), Rule 5 (consistent `Mode*` prefix)
- [x] JRENG-CODING-STANDARD — positive nested checks in new init() (`if AppVersion == "dev" { if ok && ... { ... } }`), no early returns added; existing file patterns preserved
- [x] Never-Stranded invariant — `ModeStartup` always transitions on `RemoteFetchMsg` receipt (success OR failure)

### Problems Solved
- **`go install @latest` reports "dev"** — root cause identified as Go module proxy immutability (v0.0.2 SHA locked at first publish, tag retag ineffective) compounded by `var AppVersion = getVersion()` function-call initializer silently disabling `-X` ldflags; new pattern restores both goreleaser injection and `go install @version` reporting for future tags
- **Remote status check too late** — menu was generated from stale local refs at construction; user could dispatch push/pull before `git fetch` returned. `ModeStartup` gates menu construction behind fetch completion; failure mode degrades gracefully with cached state + footer warning instead of stranding user
- **Pre-existing `var AppVersion = getVersion()` latent bug** — ldflags path was broken for all future goreleaser builds on HEAD; fixed alongside

### Debts Paid
- None (no `DEBT.md` present at project root)

### Debts Deferred
- `GOROOT` misconfiguration on ARCHITECT's dev machine (`/mingw64/lib/go` unix-style breaks Windows `go` binary when invoked from bash without explicit override) — environment issue flagged by Engineer during TIT/CAKE verification, not a code concern
- Local builds now display Go-synthesized pseudo-version (`vX.Y.Z-0.TIMESTAMP-HASH+dirty`) instead of bare `"dev"` during in-tree `go build` — functional, carries debug info; ARCHITECT may tighten the init() fallback to reject pseudo-versions if bare "dev" is preferred for local builds (cosmetic only)
- `v0.0.3` tag not yet cut — fixes will not reach users via `go install @latest` until ARCHITECT runs `bash release.sh v0.0.3` (v0.0.2 proxy cache cannot be invalidated)

---

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
