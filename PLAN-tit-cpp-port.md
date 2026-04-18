# PLAN: TIT-cpp Port to C++/JUCE via `jam::tui`

**RFC:** RFC.md
**Date:** 2026-04-18
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 (reference implementation, MANIFESTO.md enforced as-written per LANGUAGE.md §C++/JUCE)

## Overview

Full-feature port of Go TIT to C++17/JUCE8 against the `jam::tui` framework. Go sources archived under `___legacy___/` for reference and continued Windows/Linux shipping. Port executes as a six-sprint arc: scaffold → primitives → state+views (demo milestone) → git → integration (macOS MVP) → Windows MSYS2 parity. No architectural invention — `SPEC.md` is the locked requirements contract, END is the architectural template, caroline provides the framework base.

## Addendum 2026-04-19

TIT migrated from local `modules/jreng_*` forks to external consumption of `~/Documents/Poems/dev/jam/jam_*/`. Zero behavioral delta (verified). Namespace: `jreng::*` → `jam::*`. Historical Sprint descriptions below preserve original `jreng::` names where they describe in-session forge work.

## Language / Framework Constraints

- **C++17, JUCE 8** — MANIFESTO.md enforced as written; no overrides.
- **`jam::tui` namespace** — `tui::Component : public juce::Component` (Path A). JUCE lingua preserved (`resized()`, `setBounds()`, `paint()`). Native JUCE types consumed wholesale (`juce::Colour`, `juce::Font`, `juce::Rectangle<int>`, `juce::AttributedString`, `juce::KeyPress`).
- **Cell vs pixel coordinates** — `juce::Rectangle<int>` means cells inside `tui::Component` context, pixels outside. Single semantic shift, same type.
- **Threading model** — message thread owns `TitState` VT + views; subprocess worker owns `juce::ChildProcess`; `callAsync` is the only cross-thread primitive. Zero locks on hot path. Mirror END.
- **Config format** — XML → `juce::ValueTree` native. Zero new parser dependencies.

## Validation Gate

Each step MUST be validated before proceeding to the next.
Validation = `@Auditor` confirms step output complies with ALL documented contracts:
- MANIFESTO.md (BLESSED: Bound, Lean 300/30/3, Explicit, SSOT, Stateless, Encapsulation, Deterministic)
- NAMES.md (naming philosophy)
- carol/JRENG-CODING-STANDARD.md (C++ coding standards)
- The locked PLAN decisions (no deviation, no scope drift)

**Test framework:** `juce::UnitTest` (JUCE-native, zero external dep, aligns with CONTRACT "use JUCE/jam as much as we can"). Fixture-based assertions for parsers, state transitions, menu generation, theme round-trip, braille output byte-identity. Test runner wired into `CMakeLists.txt` as a separate target. Coverage expectations scoped per step under the step's Validation clause.

## Steps

### SPRINT 1 — Phase 0: Scaffold + Forks + Stub Modules

#### Step 1.1: Archive Go sources under `___legacy___/`
**Scope:** clean-slate move — everything except CAROL-protocol docs and harness infra. Root survivors (ARCHITECT decision 2026-04-18): `.git/`, `.claude/`, `carol/`, `CLAUDE.md` (symlink), `SPEC.md`, `RFC.md`, `PLAN-tit-cpp-port.md`. Everything else moves: `cmd/`, `internal/`, `go.mod`, `go.sum`, `build.sh`, `release.sh`, `.goreleaser.yaml`, `entitlements.plist`, `scripts/`, `ARCHITECTURE.md`, `CODEBASE-MAP.md`, `RELEASE_NOTES.md`, `README.md`, `LICENSE`, `.gitignore`, `screenshot/`, `dist/`.
**Action:** `@Engineer` moves the tree via `git mv` where tracked, plain `mv` for untracked. Go TIT must still build via `cd ___legacy___ && go build ./cmd/tit`. No source edits — pure move. New C++ `.gitignore` and `README.md` created fresh at root in Step 1.2 and Step 5.4 respectively.
**Validation:** `@Auditor` — Go build green via `cd ___legacy___ && go build ./cmd/tit`; root contains only the seven survivors + `___legacy___/`; no source file content changed.
**Status:** ✅ Complete 2026-04-18 — 16 entries moved (14 git-mv, 2 plain mv). Go build verified.

#### Step 1.2: Scaffold JUCE CMake project
**Scope:** `CMakeLists.txt`, `.gitignore`, `Source/Main.cpp`, `Source/TitApp.h/.cpp`
**Action:** `@Engineer` writes minimal JUCE `juce_add_console_app` target. Executable name: **`titc`** (ARCHITECT decision 2026-04-18 — avoids PATH collision with Go `tit` binary during port; rename to `tit` deferred until Sprint 6 Windows parity closes and Go legacy is retired). `Main.cpp` = `JUCEApplication` entry. `TitApp` = empty shell class. Build green on macOS (Apple Silicon + Intel).
**Validation:** `@Auditor` — `cmake --build` produces `titc` executable; zero warnings at END's suppression-only flag set (`-Wno-unused-parameter -Wno-sign-conversion -Wno-shadow`, no `-Wall -Wextra`) — mirrors END's proven pattern per RFC §2.1; file topology matches RFC §4.1.

#### Step 1.3: Fork `jam_core` and `jam_data_structures` verbatim from caroline
**Scope:** `modules/jreng_core/`, `modules/jreng_data_structures/`
**Action:** `@Engineer` copies both modules verbatim from `~/Documents/Poems/dev/caroline/modules/`. Wire into `CMakeLists.txt`. Zero source edits.
**Validation:** `@Auditor` — diff against caroline source = zero delta; builds green as JUCE module dependencies of TitApp.

#### Step 1.3b: Extend `jam_core` with `jam::File::Watcher`
**Scope:** `modules/jreng_core/file/` — new nested class `jam::File::Watcher` inside existing `struct File` in `jreng_file.h`
**Action:** `@Engineer` forks `FigBug/Gin`'s `modules/gin/utilities/gin_filesystemwatcher.{h,cpp}` (BSD, deps: `juce_core` + `juce_data_structures`) into `jreng_core/file/`. Rewrap `class gin::FileSystemWatcher` as `struct File::Watcher` nested inside `jam::File`. Rewrap `FileSystemEvent` enum as `File::Watcher::Event` nested enum. Platform backends preserved verbatim: macOS FSEvents, Windows `ReadDirectoryChangesW`, Linux inotify. API: `addFolder(File)`, `removeFolder(File)`, `addListener(Listener*)`, `removeListener(Listener*)`; `Listener::folderChanged(File&)` + `Listener::fileChanged(File&, Event)`. Replace upstream JUCE include with jam_core's own aggregation header.
**Validation:** `@Auditor` — BLESSED per file (B: Watcher owns its thread/RAII; E: no early returns; S: nested under existing `jam::File` SSOT, no new top-level namespace); attribution preserved (BSD notice + Roland Rabien copyright in header); `folderChanged` and `fileChanged` callbacks proven on fixture dir mutation across all three platforms (macOS now, Win/Linux re-verified in Sprint 6).

#### Step 1.4: Fork `jam_tui` + `jam_markdown` from caroline (post-rename state)
**Scope:** `modules/jreng_tui/`, `modules/jreng_markdown/`
**Action:** `@Engineer` copies both modules verbatim from caroline (namespace already `jam::tui`, method already `paint()` per caroline Sprint 5). `jam_markdown` is forked alongside `jam_tui` because `jam_tui.h` declares `jam_markdown` as a JUCE-module dependency (markdown renderer lives in `jam_tui/markdown/AnsiMarkdownRenderer` consuming `jam_markdown`) — unused by TIT per RFC §2.2 but enforced at CMake configure time. Omission discovered 2026-04-18 during Step 1.4 execution; RFC §2.2 ground truth and RFC §4.1 scaffold missed the transitive dep. Wire both modules into `CMakeLists.txt` in dependency order: `jam_markdown` before `jam_tui`.
**Validation:** `@Auditor` — diff against caroline source = zero delta (both modules); builds green; `-Woverloaded-virtual` clean; TIT application code never imports from `jam_markdown` (verified at Sprint 5 integration).

#### Step 1.5: Implement `jam_subprocess` (caroline has stub only)
**Scope:** `modules/jreng_subprocess/`
**Action:** `@Engineer` implements per RFC-CAROLINE-00 §4.5 spec. macOS-only first (`fork`/`exec` via `juce::ChildProcess`). Streaming stdout/stderr callbacks deliver chunks on subprocess thread (not message thread). Windows impl deferred to Sprint 6.
**Validation:** `@Auditor` — BLESSED compliance (B: subprocess RAII via `std::unique_ptr<juce::ChildProcess>`; E: no early returns; S: single-writer/single-reader channel); contribution back to caroline noted for future sync.

#### Step 1.6: Implement braille renderer inside `jam_tui`
**Scope:** `modules/jreng_tui/braille/` — new subsystem inside existing jam_tui module. **No `jam_svg_braille` module.** ARCHITECT decision 2026-04-18: braille is TUI rendering, belongs inside jam_tui. Parsers (text→AST) stay generic (`jam_markdown` remains separate); renderers (AST/data→display) live inside jam_tui.
**Action:** `@Engineer` ports rasterization + braille encoding from `___legacy___/internal/banner/svg_render.go` (150 LOC) + `braille.go` (138 LOC). **Path `d` command-stream parsing delegated to `juce::Drawable::parseSVGPath()`** per CONTRACT (use JUCE/jam SSOT, don't fight the framework). `svg_paths.go` (351 LOC) is NOT ported — rasterizer rewritten to consume `juce::Path` via `juce::Path::Iterator` instead of Go's `[][]Point` subpath slices. Return type is `jam::tui::Cell` (natural fit — braille IS a TUI rendering output, no cross-module layer issue).
**Validation:** `@Auditor` — braille output **visually equivalent** to Go version against reference SVG fixture (TIT startup banner) — banner shape recognizable, colors preserved per-path (ARCHITECT decision 2026-04-18: JUCE tessellation is SSOT; byte-identity with Go's fixed `approximateCubicBezier(20)` is not a requirement — per-pixel scanline differs at cell boundaries, visual identity holds); BLESSED-S (JUCE path parser + tessellator is SSOT); no duplicate SVG parsing code; no new module in `modules/` (braille subsystem lives under `jam_tui/`).

---

### SPRINT 2 — Phase 1: `jam_tui` Extensions (8 Primitives)

#### Step 2.1: Synthetic `juce::ValueTree` fixture framework
**Scope:** `tests/fixtures/`, fixture loader utility
**Action:** `@Engineer` writes VT fixture loader (XML → VT). Hand-authored fixtures for every state tuple from SPEC §3 (GitEnvironment × WorkingTree × Timeline × Operation × Remote). View primitives in Step 2.2+ render from these, no git required.
**Validation:** `@Auditor` — fixture VT schema matches RFC §4.6; loader round-trips XML cleanly.

#### Step 2.2: Implement `Menu`, `ListPane`, `SplitPane`
**Scope:** `modules/jreng_tui/primitives/jreng_tui_menu.*`, `..._listpane.*`, `..._splitpane.*` — primitives live in `primitives/` subdirectory per existing `jam_tui` convention (both `.h` and `.cpp` co-located; see `ansi/`, `braille/`, etc.). Amended 2026-04-18 post-audit.
**Action:** `@Engineer` implements three primitives per RFC §4.3. Reference Go sources in `___legacy___/internal/ui/menu.go`, `listpane.go`, `layout.go`+`history.go`. All inherit `tui::Component`, consume JUCE native types.
**Validation:** `@Auditor` — BLESSED per-file (300/30/3); renders cleanly against fixtures; keyboard navigation matches Go behavior.

#### Step 2.3: Implement `Dialog`, `ConsoleStream`, `TextPane`
**Scope:** `modules/jreng_tui/primitives/jreng_tui_dialog.*`, `..._consolestream.*`, `..._textpane.*`
**Action:** `@Engineer` implements three primitives per RFC §4.3. References `___legacy___/internal/ui/confirmation.go`, `console.go`+`buffer.go`, `textpane_render.go`+`textpane_input.go`.
**Validation:** `@Auditor` — same criteria; `ConsoleStream` atomic-writer cross-thread pattern matches END's proven template.

#### Step 2.4: Implement `Spinner`, `ThemeResolver`
**Scope:** `modules/jreng_tui/primitives/jreng_tui_spinner.*`, `..._theme_resolver.*`
**Action:** `@Engineer` implements final two primitives per RFC §4.3. `ThemeResolver` reads `juce::ValueTree` node via `juce::Identifier` keys.
**Validation:** `@Auditor` — same criteria; ThemeResolver contract round-trips with Step 3.2 theme XML schema.

#### Step 2.5: Contribute back to caroline
**Scope:** cross-project — `~/Documents/Poems/dev/caroline/modules/jam_tui/`, `~/Documents/Poems/dev/caroline/modules/jam_core/file/`
**Action:** COUNSELOR files handoff.md to caroline's COUNSELOR documenting (a) the 8 new `jam_tui` primitives, (b) `jam::File::Watcher` addition to `jam_core`, (c) `jam_subprocess` implementation (Step 1.5), (d) `jam_svg_braille` implementation (Step 1.6). All for inheritance in caroline's next sync sprint. Not a TIT build step. Coordination only.
**Validation:** `@Auditor` — handoff documented; no TIT build impact.

---

### SPRINT 3 — Phase 2 + 3: TitState + Views (Demo Milestone)

#### Step 3.1: Draft theme XML schema + update SPEC.md
**Scope:** COUNSELOR writes theme schema. SPEC.md §16 path edit: `default.toml` → `default.xml`.
**Action:** COUNSELOR drafts XML schema with **three-level hierarchy: `Theme > LookAndFeel > Component`** (ARCHITECT decision 2026-04-18). Theme is top-level palette. LookAndFeel groups Component-family styles (JUCE `juce::LookAndFeel` pattern). Component nodes bind leaf properties (colors, paddings). Covers all 10 theme fields (RFC §6.4) as Component-level leaves. **Hot-reload is MVP** via `jam::File::Watcher` (forked in Step 1.3b) — watches `~/.config/tit/themes/` folder. On `folderChanged` callback, re-parse active theme XML → `ValueTree::copyPropertiesAndChildrenFrom` on `THEME` subtree → `ValueTree::Listener` on `THEME` repaints views. `ThemeLoader` owns the `Watcher` instance and the change handler (200 ms debounce to swallow multi-write sequences).
**Validation:** `@Auditor` — schema round-trips via `juce::XmlDocument::parse` + `juce::ValueTree::fromXml`; hierarchy matches locked Theme > LookAndFeel > Component; `jam::File::Watcher` → VT → repaint proven against fixture theme swap; debounce prevents mid-write reload; SPEC edit is text-only, no semantic change beyond file extension.

#### Step 3.2: Implement `TitIdentifier.h` + `TitAxis.h`
**Scope:** `Source/TitIdentifier.h`, `Source/state/TitAxis.h`
**Action:** `@Engineer` declares all `juce::Identifier` constants (SSOT per BLESSED-S). Enum definitions for `WorkingTree`, `Timeline`, `Operation`, `Remote` per SPEC §3.
**Validation:** `@Auditor` — single SSOT for every VT key in project; enum coverage matches SPEC.md §3 tables exactly.

#### Step 3.3: Implement `TitState` (APVTS-mirror)
**Scope:** `Source/state/TitState.h/.cpp`
**Action:** `@Engineer` implements `TitState` per RFC §3.6 — atoms for fast reads, `juce::ValueTree` for observation, `juce::Timer::flush()` copies atoms → VT on message thread. Mirror END's `Terminal::State` template verbatim in structure.
**Validation:** `@Auditor` — BLESSED (S: VT is SSOT; B: Timer owned by TitState, ChildProcess by worker thread; E: no early returns); cross-thread contract proven via fixture injection test.

#### Step 3.4: Implement `MenuBuilder` dispatch
**Scope:** `Source/menu/MenuBuilder.h/.cpp`, `Source/menu/MenuItems.h`
**Action:** `@Engineer` implements `MenuGeneratorMap<Operation, std::function<juce::Array<MenuItem>(const juce::ValueTree&)>>` per RFC §3.6. `MenuItems.h` = data SSOT for all 27 menu items from SPEC §6. Builder is pure function of VT.
**Validation:** `@Auditor` — deterministic (same VT + same operation → bit-identical menu); 3-branch rule honored via lookup map not switch chain; SPEC §6 coverage = 27/27.

#### Step 3.5: Implement root views
**Scope:** `Source/view/TitScreen.*`, `Banner.*`, `Header.*`, `Footer.*`, `MenuView.*`
**Action:** `@Engineer` implements root composition and four UI shell views per SPEC §14. Each attaches `ValueTree::Listener` to relevant subtree. Zero explicit `rebuild()` calls.
**Validation:** `@Auditor` — BLESSED-E (Encapsulation: views import no `Source/git/`); views render every fixture state from Step 2.1.

#### Step 3.6: Implement browser + dialog views
**Scope:** `HistoryView.*`, `FileHistoryView.*`, `ConflictResolverView.*`, `ConsoleView.*`, `ConfirmDialog.*`, `SetupWizardView.*`
**Action:** `@Engineer` implements remaining views per SPEC §9, §10, §11, §12.
**Validation:** `@Auditor` — BLESSED per file; `ConfirmDialog` 7 variants (rewind, time-travel, dirty, merge, push, branch, time-travel-return) all render against fixtures.

**COMPONENT+STATE MILESTONE** — At Sprint 3 close: all 11 view components built and individually tested against `juce::ValueTree` fixtures; `TitState` APVTS-mirror machine live (atoms → VT via `juce::Timer::flush`); `MenuBuilder` dispatch covering 8 Operation states × 27 items. View-to-view navigation/dispatch (e.g., `Operation==TimeTraveling` → `HistoryView` active) is Sprint 4 scope — couples with Operation FSMs + git layer. Amended 2026-04-18 (corrected from earlier "navigates every view" overclaim).

---

### SPRINT 4 — Phase 4 + 5: Git Layer + Protocol FSMs

#### Step 4.1: Implement `GitCommands.*` + command builders
**Scope:** `Source/git/GitCommands.h/.cpp`
**Action:** `@Engineer` implements typed command builders for every git call site in Go TIT: init, clone, commit, push, pull, merge, rebase, reset, checkout, branch, stash, config, log, status. Pure string builders, no execution.
**Validation:** `@Auditor` — every command corresponds 1:1 to a Go call site in `___legacy___/internal/git/`; no magic strings.

#### Step 4.2: Implement `GitRunner`
**Scope:** `Source/git/GitRunner.h/.cpp`
**Action:** `@Engineer` implements subprocess orchestrator consuming `jam_subprocess`. Worker thread runs commands, streams stdout/stderr back via `callAsync` to message thread. Owns marker-file protocol surface.
**Validation:** `@Auditor` — BLESSED-B (thread ownership clean, `callAsync` only crossing primitive); streaming delivers progress to `ConsoleStream` without locks on hot path.

#### Step 4.3: Implement parsers
**Scope:** `Source/git/parsers/PorcelainV2Parser.*`, `LogParser.*`, `UnifiedDiffParser.*`, `ConflictMarkerParser.*`
**Action:** `@Engineer` implements each parser per Go reference. `PorcelainV2Parser` consumes `git status --porcelain=v2 --branch` output. `LogParser` handles `-z` null-delimited format. `UnifiedDiffParser` parses `git show` / `git diff`. `ConflictMarkerParser` extracts `<<<<<<<` / `=======` / `>>>>>>>` blocks.
**Validation:** `@Auditor` — parser output matches Go reference against fixture repo snapshots; no early returns in parse loops.

#### Step 4.4: Implement `GitStateDetector`
**Scope:** `Source/git/GitStateDetector.h/.cpp`
**Action:** `@Engineer` implements reader-thread parser (analog to END's `Parser`). Ingests porcelain output + marker files (`MERGE_HEAD`, `REBASE_HEAD`, `.git/TIT_DIRTY_OP`, `.git/TIT_TIME_TRAVEL`), writes atoms on `TitState`. Mirrors END's `Parser` template structure.
**Validation:** `@Auditor` — state detection matches Go `DetectState()` against a matrix of fixture repos; detector runs on worker thread, atomic writes only.

#### Step 4.5: Implement protocol FSMs
**Scope:** `Source/protocols/DirtyOpProtocol.*`, `TimeTravelProtocol.*`, `ConflictProtocol.*`, `SetupWizard.*`
**Action:** `@Engineer` implements four FSMs per SPEC §7 (DirtyOp), §10 (TimeTravel), §5 (Conflict), §13 (SetupWizard). Each owns its marker file via RAII.
**Validation:** `@Auditor` — marker file RAII ownership clear (BLESSED-B); FSM transitions match SPEC exactly; abort paths restore preconditions.

---

### SPRINT 5 — Phase 6: Integration + macOS Release

#### Step 5.1: Wire `GitRunner` to `TitState`
**Scope:** `Source/TitApp.*`
**Action:** `@Engineer` completes `TitApp` — owns `TitState`, `TitScreen`, `GitRunner`. Views observe VT, GitRunner produces VT updates. Fixture framework retired for production wiring.
**Validation:** `@Auditor` — end-to-end flow test (init repo → commit → push) executes correctly; views update without explicit rebuild calls.

#### Step 5.2: Integration test matrix
**Scope:** test fixtures covering every SPEC state tuple
**Action:** `@Engineer` runs integration matrix: every flow from SPEC §6 (Normal), §7 (Dirty), §10 (Time Travel), §13.6 (Manual Detached HEAD).
**Validation:** `@Auditor` — full SPEC acceptance matrix passes; `@Auditor` produces coverage report.

#### Step 5.3: Release pipeline — macOS
**Scope:** `CMakeLists.txt` install target, code signing, notarization
**Action:** `@Engineer` adds `cmake --install` → `/usr/local/bin/titc` target. macOS codesign + notarize as a fresh rewrite for CMake (not a port of Go TIT's `scripts/post-build.sh`) — ARCHITECT decision 2026-04-18. Existing Apple Developer ID + notarization credentials reused. `entitlements.plist` authored anew against JUCE console-app requirements. No Homebrew tap at MVP.
**Validation:** `@Auditor` — signed + notarized `titc` binary runs on fresh macOS (Intel + Apple Silicon); BLESSED compliance final pass on all new code; `spctl --assess` passes.

#### Step 5.4: Documentation refresh
**Scope:** `README.md`, `ARCHITECTURE.md`
**Action:** COUNSELOR rewrites `ARCHITECTURE.md` to mirror C++ implementation (per CAROL COUNSELOR doctrine — ARCHITECTURE.md is descriptive). `README.md` reflects `go install` → `cmake --install` transition. Go binary continues shipping from `___legacy___/`.
**Validation:** `@Auditor` — docs reflect code; no stale Go references outside `___legacy___/` scope.

**macOS MVP COMPLETE** — feature parity with Go TIT shipping on macOS.

---

### SPRINT 6 — Windows MSYS2 Parity

#### Step 6.1: `jam_subprocess` Windows implementation
**Scope:** `modules/jreng_subprocess/` Windows platform guard
**Action:** `@Engineer` adds `CreateProcess` + Windows pipe handling. ConPTY byte handling for color/cursor sequences. Windows streams via `juce::ChildProcess` Windows path.
**Validation:** `@Auditor` — parity with macOS implementation; BLESSED per file; `#ifdef JUCE_WINDOWS` scope minimal.

#### Step 6.2: Path normalization
**Scope:** cross-cutting — wherever paths cross shell/git/OS boundary
**Action:** `@Engineer` audits MSYS2 `/c/...` ↔ Windows `C:\...` cases. Centralize translation in one utility.
**Validation:** `@Auditor` — SSOT utility; no duplicated translation sites; BLESSED-S.

#### Step 6.3: Windows integration test matrix
**Scope:** Windows MSYS2 test runs on ARCHITECT's 2 Windows dev machines (CLANGARM64, MINGW64)
**Action:** `@Engineer` runs SPEC acceptance matrix on Windows. MSYS2 + native-Windows `git.exe` both probed.
**Validation:** `@Auditor` — full matrix passes on both Windows targets.

#### Step 6.4: Windows release pipeline
**Scope:** `CMakeLists.txt` Windows install target
**Action:** `@Engineer` adds Windows install + packaging. Path convention matches Go TIT's existing Windows distribution.
**Validation:** `@Auditor` — Windows binary ships alongside macOS; no regression on either.

**WINDOWS PARITY COMPLETE** — `___legacy___/` Go binary can be retired for Windows (ARCHITECT call).

## BLESSED Alignment

- **B — Bound** — `TitApp` owns `TitState`, `TitScreen`, `GitRunner` via `std::unique_ptr`. `GitRunner` owns thread pool. Each `juce::ChildProcess` owned by its worker thread. Marker files owned by Protocol FSMs via RAII. Threads bound: message thread owns VT+views, subprocess workers own ChildProcess.
- **L — Lean** — 300/30/3 per file enforced file-wide. Menu dispatch is `std::unordered_map` lookup, not switch chain. State detection is atom writes, not conditional ladders. YAGNI: no `jam_git` extraction until second consumer (post-MVP).
- **E — Explicit** — Zero early returns. All parameters visible. Magic values → named constants in `TitIdentifier.h` (SSOT). `jassert` on invariants. No silent fails — every error path writes `Console::LINE{stream:stderr}` or raises setup wizard.
- **S — SSOT** — `TitState` ValueTree is SSOT for application state. `MenuItems.h` is SSOT for menu definitions. `TitIdentifier.h` is SSOT for all `juce::Identifier` keys. Theme XML is SSOT for colors. Marker files are SSOT for protocol progress.
- **S — Stateless** — `tui::Component` subclasses hold transient render state only (scroll, focus). All persistent state lives in `TitState`. Orchestrator tells, never asks.
- **E — Encapsulation** — `jam_tui` imports no TIT application header (lower-layer discipline). `Source/git/` imports no `Source/view/`. Unidirectional layer flow. Views observe VT via listeners — never poke git layer.
- **D — Deterministic** — Same VT + same generator → bit-identical menu output. Same git command + same cwd → same parser output. Emergent from BLESSE discipline.

## Contract Additions (ARCHITECT 2026-04-18)

- **Use JUCE/jam as much as possible. Don't fight the framework.** SVG path parsing delegates to `juce::Drawable::parseSVGPath` (Step 1.6). Theme uses `juce::LookAndFeel` pattern (Step 3.1). Config uses `juce::ValueTree` + `juce::XmlDocument` (no new parser deps). Subprocess uses `juce::ChildProcess` (Step 1.5). Timer flush uses `juce::Timer` (Step 3.3). This rule applies across every step — when a JUCE or jam primitive solves it, use it.

## Risks / Open Questions

- **`paint()` overload/shadow resolution** — RFC §6.1 inherits caroline's decision (favored Path A per handoff.md). Verified by caroline Sprint 5. No action needed in TIT.
- **`___legacy___/` retention policy** — Sprint 6 closes Windows parity; ARCHITECT decides when to archive or delete Go sources. Not a PLAN step.
- **Sprint parallelism** — RFC §7.4 suggests primitive forge + state layer + git layer could overlap across concurrent agent sprints post-scaffold. This master PLAN is serial for validation clarity; ARCHITECT may interleave sprints at runtime.

## Resolved Decisions

- **2026-04-18 — Scope:** Master PLAN covering all 6 sprints (Phase 0 through Windows parity). Windows MSYS2 = dedicated Sprint 6.
- **2026-04-18 — SVG (RFC §6.2):** Use `juce::Drawable::parseSVGPath` for `d` attribute parsing. Port only `svg_render.go` + `braille.go`. Rasterizer rewritten against `juce::Path::Iterator`.
- **2026-04-18 — Theme (RFC §6.4):** Hierarchy `Theme > LookAndFeel > Component`. Hot-reload MVP via `jam::File::Watcher` — forked from `FigBug/Gin` `modules/gin/utilities/gin_filesystemwatcher.{h,cpp}` (BSD) into `jreng_core/file/` as nested class of existing `struct jam::File` (Step 1.3b). Native backends: FSEvents / `ReadDirectoryChangesW` / inotify. Listener API provides both `folderChanged(File&)` and `fileChanged(File&, Event)` with event type (created/deleted/updated/renamed).
- **2026-04-18 — Binary name:** `titc` during port (Sprint 1–5). Rename to `tit` deferred to Sprint 6 close when Go legacy retires.
- **2026-04-18 — Fork mechanism:** Copy-verbatim of source files. No git submodule, no git subtree. Drift managed manually via handoff.md documentation at contribution points.
- **2026-04-18 — Test framework:** `juce::UnitTest` (JUCE-native, zero external dep).
- **2026-04-18 — macOS release (Step 5.3):** Rewrite `entitlements.plist` + CMake post-build signing/notarization hook from scratch. Do not port Go TIT's `scripts/post-build.sh`. Reuse existing Apple Developer ID + notarization credentials.
