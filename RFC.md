# RFC — TIT-cpp Port to C++/JUCE via `jreng::tui`

**Date:** 2026-04-17
**Status:** Ready for COUNSELOR handoff
**Author:** BRAINSTORMER
**Project:** TIT — Terminal Interface for git
**Path:** `~/Documents/Poems/dev/tit/`

---

## 1. Problem Statement

Go TIT ships and works. It is also a **stack fracture** in ARCHITECT's otherwise-unified `jreng_*` ecosystem:

- Go toolchain friction (proxy cache garbage on every release)
- Not BLESSED-auditable in ARCHITECT's native language
- Cannot share `jreng_*` substrate consumed by END, CAROLINE, Kuassa plugin, whatdbg
- Cannot consume `jreng_subprocess` streaming pattern that TIT would forge for the ecosystem
- bubbletea/Elm pure-update model is a worse fit than `juce::ValueTree` listener-driven observable state

**Port target:** full feature parity with Go TIT, rewritten in C++17/JUCE8 against the `jreng::tui` TUI framework forged in the same sprint. Go sources preserved under `___legacy___/` for reference and continued availability on platforms where the C++ port doesn't yet ship.

**Strategic value:**

1. **Vertical integration:** TIT-cpp converges the last daily-driver tool into the `jreng_*` substrate. One terminal, one debugger, one code editor, one substrate, native binary everywhere.
2. **Framework forge:** `jreng::tui` needs composition-level primitives (menu, list, split-pane, dialog, console, textpane, spinner, theme-resolver) that don't exist yet. TIT-cpp's feature set exercises every primitive against a shipped reference implementation (Go TIT). Caroline inherits a proven framework instead of co-developing it.
3. **Ecosystem contribution:** `jreng_subprocess` and `jreng_svg_braille` modules exist as declaration stubs only in caroline. TIT-cpp implements both, ships them back to the ecosystem first.
4. **Pre-CAROL debt paid:** Go TIT carries pre-formalization architectural artifacts (God-object `Application` struct with 21 fields + 10 extracted sub-structs). Port is the clean-origin pass under BLESSED from day zero.

---

## 2. Research Summary

### 2.1 Reference implementations available

- **Go TIT** (`~/Documents/Poems/dev/tit/`, ~23k LOC, 183 files) — requirements are fully specified in existing `SPEC.md` (1,025 lines) and `ARCHITECTURE.md`. Zero design decisions left; port is pure translation.
- **END** (`~/Documents/Poems/dev/end/`) — battle-tested architectural template. `Source/terminal/data/State.h` (APVTS-style `std::atomic<float>` + `juce::ValueTree` SSOT + timer flush), `Source/terminal/logic/Parser.h` (reader-thread byte-stream state machine with O(1) DispatchTable). Runs in production at VT520 byte rate on GPU terminal emulator.
- **CAROLINE** (`~/Documents/Poems/dev/caroline/`) — `jreng_tui` primitives forged here first (3,384 LOC implemented). Naming formalization pending (see §6 Handoff Notes).

### 2.2 Caroline module ground truth (verified 2026-04-17)

| Module | LOC | Status |
|---|---|---|
| `jreng_core` | 15,550 | **Implemented** (END verbatim fork) — utilities, concurrency (mailbox + snapshot_buffer), xml (incl. svg parser), identifier, file, string, text, value, image, context, debug, map, project_info, function_map, fuzzy_search, binary_data |
| `jreng_data_structures` | 895 | **Implemented** — `value_tree/` (ValueTree wrapper, taproot, walker) + `value_tree_json/` (JSON ↔ VT) |
| `jreng_tui` | 3,384 | **Implemented** — `ansi/` (ANSIComponent, ANSIGraphics, ANSIScreen, ANSIWriter, TextBox, color, cell, escapes), `graphics/` (Rectangle), `input/` (TerminalInput, KeyEvent), `markdown/` (AnsiMarkdownRenderer — unused by TIT), `metrics/` (TerminalMetrics) |
| `jreng_markdown` | 1,048 | Implemented (unused by TIT) |
| `jreng_subprocess` | 16 | **Stub only** — module declaration, zero implementation |
| `jreng_svg_braille` | 17 | **Stub only** — module declaration, zero implementation |

### 2.3 Non-novelty

TIT-cpp is pure translation:

- **Requirements:** locked by Go `SPEC.md` (5-axis state model, 27 menu items, 4 protocol FSMs, conflict resolver, history browser, file-history 3-pane, setup wizard, dirty-op protocol, time-travel round-trip, manual-detached-HEAD support)
- **Architecture pattern:** locked by `Terminal::State` / `Parser` template
- **Framework base:** locked by caroline's `jreng_tui` (post-rename)
- **Naming contract:** locked by ARCHITECT's `jreng::tui` decision (see §3.2)
- **Config format:** locked — XML → `juce::ValueTree` native

Zero architectural invention required.

---

## 3. Principles and Rationale

### 3.1 Vertical integration (per BLESSED — all)

TIT-cpp shares substrate with END, CAROLINE, Kuassa audio plugin, whatdbg, CAKE (future port). Bugs in shared modules surface against the strictest consumer first (Kuassa real-time audio thread, END GPU text-rendering hot path) and are fixed before reaching TIT's comparatively lenient UI-thread path. Infrastructure cost amortizes across 5+ consumers.

### 3.2 `jreng::tui` namespace contract (ARCHITECT decision 2026-04-17)

- **Namespace:** `jreng::tui` (supersedes caroline's pre-formal `jreng::Terminal`)
- **Base class:** `tui::Component : public juce::Component` — Path A per RFC-CAROLINE-00 §22 (inherits JUCE focus, mouse, keyboard, modal, bounds infrastructure)
- **JUCE lingua preserved:** `resized()`, `setBounds()`, `getBounds()`, `addChildComponent()`, `toFront()`, focus traversal — all inherited unchanged
- **JUCE native types consumed wholesale:** `juce::Colour`, `juce::Font`, `juce::Rectangle<int>`, `juce::AttributedString`, `juce::Justification`, `juce::Point<int>`, `juce::MouseEvent`, `juce::KeyPress`. No custom geometry primitives.
- **Single semantic shift:** `juce::Rectangle<int>` represents **cell coordinates** (col/row) inside `tui::Component` context; pixel coordinates outside. Same type, context-dependent meaning.

### 3.3 Layering — `jreng_tui` is View+Input only (per BLESSED E, Encapsulation)

`jreng_tui` must **not** own a Model primitive. That would force every consumer to adopt its state shape — a View framework reaching up into application architecture. END's `Terminal::State` lives in `Source/terminal/data/` for exactly this reason. TIT-cpp's `TitState` lives in `Source/state/` — domain-specific application code consuming primitives from `jreng_core` (mailbox, snapshot_buffer) and `jreng_data_structures` (ValueTree wrapper).

When a second consumer (CAROLINE's Runtime/Session/Conversation trees) demands the APVTS-style atomic-flush pattern, extract the machinery into `jreng_data_structures::ParamStore`. **YAGNI until second consumer. First consumer inlines.**

### 3.4 Dogfood build order (per BLESSED S — Stateless, D — Deterministic)

Build the UI layer against **synthetic fixture ValueTrees first, wire git last**. This enforces layering as proof, not intention:

1. If View-layer is built before git layer exists, no Component can leak git-state semantics into view code — there's nothing to leak.
2. Views render from `state.operation = "merging"` which could equally be CAKE's `"configuring"` or whatdbg's `"stepping"` — genre-generic falls out of build order.
3. Every checkpoint produces a navigable TIT demo with fixture data. Phase 3 (demo-quality TIT with zero git installed) is the defendable architectural milestone.
4. Git layer wires as ValueTree producer-observer only. If it had any direct View knowledge, Phase 3 wouldn't work — and we'd know immediately.

### 3.5 Config format — XML → ValueTree native

- JUCE-native: `juce::XmlDocument::parse()` + `juce::ValueTree::fromXml()` both first-class in `juce_core`
- Zero new module dependencies (no TOML parser, no YAML detour)
- `jreng_data_structures` already owns VT-XML round-trip (END fork inheritance)
- Theme literally becomes a ValueTree. Any future TIT config file (keybindings, preferences, forbidden-ops, setup wizard prompts) follows the same pattern. Hot reload is `FileWatcher → parse → VT::copyPropertiesAndChildrenFrom → Listeners fire` — the exact pipeline already proven by `Terminal::State::flush`.
- SPEC §16 path updates: `~/.config/tit/themes/default.toml` → `default.xml`. Text-only SPEC edit.

### 3.6 Pattern mirror — `TitState` ↔ `Terminal::State`

Full structural analogy to END:

| END | TIT-cpp |
|---|---|
| PTY reader thread | `juce::ChildProcess` on `juce::Thread` (git subprocess worker) |
| `Parser::process(bytes)` | `GitStateDetector::ingest(bytes)` — parses `git status --porcelain=v2 --branch`, `git rev-list --count`, `MERGE_HEAD`/`REBASE_HEAD`/`.git/TIT_DIRTY_OP`/`.git/TIT_TIME_TRAVEL` markers |
| `Grid` cell buffer | *n/a* — no Model-level grid (ANSI cells are View-layer inside `jreng_tui::ANSIGraphics`) |
| `Terminal::State` (atomic map + VT) | `TitState` — atoms for `workingTree`, `timeline`, `operation`, `remote`, `isTitTimeTravel`, `branch`, `aheadCount`, `behindCount`, `detachedAt` |
| `DispatchTable[state, byte] → action` | `MenuGeneratorMap[operation] → std::function<juce::Array<MenuItem>(const juce::ValueTree&)>` |
| `juce::Timer::flush()` | identical — timer polls `needsFlush`, copies atoms → VT in one pass |
| UI `ValueTree::Listener` | identical — `TitScreen` and child components attach listeners, rebuild on flush |
| `juce::Session` (orchestration) | `GitRunner` — subprocess pool, stdout streaming, marker-file protocol orchestration |

---

## 4. Scaffold

### 4.1 Project structure

```
~/Documents/Poems/dev/tit/
├── ___legacy___/                    # Go sources archived verbatim (reference + existing binary continues shipping)
│   ├── cmd/
│   ├── internal/
│   ├── go.mod, go.sum
│   ├── ARCHITECTURE.md              # referenced by RFC §2.1 — not deleted
│   ├── CODEBASE-MAP.md
│   └── ...                          # all Go-side files
├── CMakeLists.txt                   # JUCE CMake project
├── Source/                          # TIT-cpp application code
│   ├── Main.cpp                     # juce::JUCEApplication entry
│   ├── TitApp.h/.cpp                # headless app object, owns TitScreen + TitState + GitRunner
│   ├── TitIdentifier.h              # all juce::Identifier constants in one place
│   ├── state/
│   │   ├── TitState.h/.cpp          # APVTS-style model — atoms + ValueTree + timer flush
│   │   └── TitAxis.h                # enum definitions: WorkingTree, Timeline, Operation, Remote
│   ├── git/
│   │   ├── GitRunner.h/.cpp         # subprocess pool, streaming, marker protocols
│   │   ├── GitStateDetector.h/.cpp  # reader-thread parser (analog to Parser)
│   │   ├── GitCommands.h/.cpp       # command builders (init, clone, commit, push, pull, merge, rebase, reset, checkout, branch, stash, config, log, status)
│   │   └── parsers/
│   │       ├── PorcelainV2Parser.h/.cpp     # `git status --porcelain=v2 --branch`
│   │       ├── LogParser.h/.cpp             # `git log --format=... -z`
│   │       ├── UnifiedDiffParser.h/.cpp     # `git show` / `git diff` output
│   │       └── ConflictMarkerParser.h/.cpp  # `<<<<<<<` / `=======` / `>>>>>>>` block extraction
│   ├── protocols/
│   │   ├── DirtyOpProtocol.h/.cpp       # snapshot → apply → restore FSM, `.git/TIT_DIRTY_OP`
│   │   ├── TimeTravelProtocol.h/.cpp    # `.git/TIT_TIME_TRAVEL`, manual-detached detection
│   │   ├── ConflictProtocol.h/.cpp      # parse markers → resolution progress → continue/abort
│   │   └── SetupWizard.h/.cpp           # env check → ssh-keygen → ssh-add → git config → init/clone
│   ├── menu/
│   │   ├── MenuBuilder.h/.cpp           # MenuGeneratorMap dispatch (replaces TIT Go `menuGenerators`)
│   │   └── MenuItems.h                  # SSOT for all 27 menu item definitions (data, not code)
│   ├── view/
│   │   ├── TitScreen.h/.cpp             # root tui::Component, owns layout
│   │   ├── Banner.h/.cpp                # jreng_svg_braille consumer, version overlay
│   │   ├── Header.h/.cpp                # branch + state indicator
│   │   ├── Footer.h/.cpp                # context hints
│   │   ├── MenuView.h/.cpp              # composed from jreng_tui::Menu primitive
│   │   ├── HistoryView.h/.cpp           # 2-col split: commits / details (jreng_tui::SplitPane)
│   │   ├── FileHistoryView.h/.cpp       # 3-pane: commits / files / diff
│   │   ├── ConflictResolverView.h/.cpp  # N-column conflict pane with focus cycling
│   │   ├── ConsoleView.h/.cpp           # streaming git stdout (jreng_tui::ConsoleStream)
│   │   ├── ConfirmDialog.h/.cpp         # 7 variants (rewind, time-travel, dirty, merge, push, branch, time-travel-return)
│   │   └── SetupWizardView.h/.cpp       # SSH key gen flow UI
│   └── theme/
│       └── ThemeLoader.h/.cpp           # `~/.config/tit/themes/*.xml` → ValueTree
├── modules/                             # TIT-cpp's forked + implemented modules (per caroline §22 — forked, portable, isolated)
│   ├── jreng_core/                      # FORK verbatim from caroline
│   ├── jreng_data_structures/           # FORK verbatim from caroline
│   ├── jreng_tui/                       # FORK from caroline (post-rename) + EXTEND (8 new primitives)
│   ├── jreng_subprocess/                # IMPLEMENT (caroline has stub only)
│   └── jreng_svg_braille/               # IMPLEMENT (caroline has stub only; port from ___legacy___/internal/banner/)
├── tests/
│   └── fixtures/                        # canned ValueTree snapshots for every state tuple
├── RFC.md                               # this document
├── SPEC.md                              # carried from Go TIT (requirements contract unchanged)
├── CLAUDE.md -> /Users/jreng/.carol/CAROL.md
└── carol/                               # SPRINT-LOG, DEBT, etc.
```

### 4.2 Module inventory (verified)

| Module | Action | Effort |
|---|---|---|
| `jreng_core` | Fork verbatim from caroline | Mechanical |
| `jreng_data_structures` | Fork verbatim from caroline | Mechanical |
| `jreng_tui` | Fork from caroline (post-rename per handoff.md) + add 8 composition primitives | 2–3 days |
| `jreng_subprocess` | **Implement** from RFC-CAROLINE-00 §4.5 spec + add streaming stdout callbacks | 1–2 days |
| `jreng_svg_braille` | **Implement** by porting `___legacy___/internal/banner/` (649 Go LOC). May reuse `jreng_core/xml/jreng_svg` for path parsing — COUNSELOR to assess | 1–2 days |

**Ecosystem contribution:** TIT-cpp ships `jreng_subprocess` and `jreng_svg_braille` first. CAROLINE inherits completed implementations when its sprint reaches those modules.

### 4.3 `jreng_tui` extensions needed (8 primitives)

All follow the `jreng::tui::Component` base class pattern, compose with existing ANSIGraphics/ANSIScreen/TextBox infrastructure, consume JUCE native types:

| Primitive | Responsibility | Reference |
|---|---|---|
| `Menu` | Vertical list with selection, letter-jump, keyboard navigation | `___legacy___/internal/ui/menu.go` |
| `ListPane` | Scrollable list with formatters, selection callbacks — generic over row data | `___legacy___/internal/ui/listpane.go` |
| `SplitPane` | 2-col / 3-pane layout composer with focus cycling between children | `___legacy___/internal/ui/layout.go` + `history.go` |
| `Dialog` | Overlay dialog with title + body + buttons, consumes `Screen::showOverlay` | `___legacy___/internal/ui/confirmation.go` |
| `ConsoleStream` | Line-buffered streaming output with autoscroll, backed by `juce::StringArray` + atomic writer | `___legacy___/internal/ui/console.go` + `buffer.go` |
| `TextPane` | Scrollable text with diff-syntax highlighting, line numbers, selection | `___legacy___/internal/ui/textpane_render.go` + `textpane_input.go` |
| `Spinner` | Animated character cycler for async operations | `___legacy___/internal/ui/spinner.go` (trivial) |
| `ThemeResolver` | `juce::ValueTree` accessor for theme colors — `theme.getColour(ID::menuSelectedBg)` | `___legacy___/internal/ui/theme.go` |

### 4.4 Phase sequence

| Phase | Deliverable | Days |
|---|---|---|
| **0** | Scaffold — `___legacy___/` archive, CMake, fork real modules, implement `jreng_subprocess` + `jreng_svg_braille`, skeleton Main.cpp builds green | 2–3 |
| **1** | `jreng_tui` extensions (8 primitives) against synthetic ValueTree fixtures | 2–3 |
| **2** | `TitState` (APVTS-mirror), `TitIdentifier`, `MenuBuilder` dispatch, fixture framework | 1–2 |
| **3** | View composition — **demo-quality TIT with zero git installed** (milestone) | 1–2 |
| **4** | Git layer — `GitRunner`, `GitStateDetector`, parsers, commands | 1–2 |
| **5** | Protocol FSMs — DirtyOp, TimeTravel, Conflict, SetupWizard | 1 |
| **6** | Integration — real flows, error paths, macOS release binary | 1 |
| **Total MVP** | Feature parity with Go TIT on macOS | **7–11 days CAROL walltime** |

**Windows MSYS2 parity post-MVP** — estimated +3–5 days (ConPTY byte handling, `CreateProcess` in jreng_subprocess Windows impl, path normalization).

### 4.5 Threading model (mirror END)

| Thread | Owns | Crossing |
|---|---|---|
| Message (JUCE main) | `TitState` ValueTree, `TitScreen`, `MenuBuilder`, all `tui::Component` | — |
| Subprocess worker (`juce::Thread` pool via `jreng_subprocess`) | `juce::ChildProcess`, stdout/stderr capture | atomic writes to `TitState` + `callAsync` for event notifications |
| Terminal input (`jreng_tui::TerminalInput`) | stdin raw-mode, escape parser, bracketed paste | `callAsync` → message thread |
| Timer (`juce::Timer`) | `TitState::flush()` | inherited — `juce::Timer` runs on message thread |

Zero locks on hot path. `callAsync` only crossing primitive. Identical discipline to END.

### 4.6 ValueTree schema (TitState root)

```
TIT
├── ENV                             # environment check results
│   ├── gitAvailable        (bool)
│   ├── sshAvailable        (bool)
│   ├── sshKeysPresent      (bool)
│   └── setupState          (string: Ready | NeedsSetup | MissingGit | MissingSSH)
├── REPO                            # current repo state (5 axes)
│   ├── workingTree         (string: Clean | Dirty)
│   ├── timeline            (string: InSync | Ahead | Behind | Diverged | "")
│   ├── operation           (string: NotRepo | Normal | Merging | Conflicted | Rebasing | TimeTraveling | DirtyOperation | Rewinding)
│   ├── remote              (string: NoRemote | HasRemote)
│   ├── isTitTimeTravel     (bool)
│   ├── branch              (string)
│   ├── aheadCount          (int)
│   ├── behindCount         (int)
│   └── cwd                 (string)
├── HISTORY                         # commit log cache
│   └── COMMIT[] (hash, author, date, message)
├── FILES                           # files changed in selected commit
│   └── FILE[] (path, status)
├── DIFF                            # current diff content for selected file
│   └── HUNK[] (oldStart, newStart, lines)
├── MENU                            # current menu state (rebuilt on REPO change via listener)
│   └── ITEM[] (id, label, hotkey, enabled, destructive)
├── CONSOLE                         # streaming subprocess output
│   └── LINE[] (text, stream: stdout|stderr|info)
├── SELECTION                       # UI-layer state
│   ├── menuIndex           (int)
│   ├── historyIndex        (int)
│   ├── fileIndex           (int)
│   └── activePane          (string)
├── THEME                           # loaded from ~/.config/tit/themes/*.xml
│   └── ...                         # all theme.go fields as VT properties
└── SETUP                           # setup wizard state (when active)
    ├── phase               (string: EnvCheck | SSHKeyEntry | KeyGen | Display | GitConfig | Done)
    ├── email               (string)
    └── publicKey           (string)
```

Every Component attaches `ValueTree::Listener` to its relevant subtree. `TitState::flush()` fires listeners on the message thread. Menu regeneration is a listener on `REPO` — no explicit `rebuildMenu()` calls anywhere.

---

## 5. BLESSED Compliance Checklist

- [x] **Bounds** — every owned resource has clear RAII lifecycle: `TitApp` owns `TitState` + `TitScreen` + `GitRunner`; `GitRunner` owns `juce::Thread` pool; each `juce::ChildProcess` owned by its worker thread via `std::unique_ptr`; marker files owned by their respective Protocol FSM via RAII
- [x] **Lean** — 300/30/3 enforced file-wide; every file one class; menu dispatch is `unordered_map<Operation, fn>` lookup, not switch chain; state detection is atom writes, not conditional ladders
- [x] **Explicit** — zero early returns; every parameter visible in signature; magic values → named constants in `TitIdentifier.h`; `jassert` on invariant violations; no silent fails — every error path writes `Console::LINE{stream:stderr}` or raises setup wizard
- [x] **Single Source of Truth** — `TitState` ValueTree is SSOT for all application state; theme XML is SSOT for colors; `MenuItems.h` is SSOT for menu definitions; marker files are SSOT for protocol progress
- [x] **Stateless** — `tui::Component` subclasses hold transient render state only (scroll offset, focus); all persistent state lives in `TitState`; orchestrator tells, never asks
- [x] **Encapsulation** — `jreng_tui` does not import any TIT application header; `Source/git/` does not import `Source/view/`; unidirectional layer flow strictly preserved
- [x] **Deterministic** — same `TitState` + same menu generator fn = bit-identical menu output; same git command input + same cwd = same parser output; emergent D from BLESSE adherence

**Known risks to monitor:**

- `ANSIComponent`'s inheritance of `juce::Component` bounds (already noted in caroline RFC-00 §17) — cell bounds must not shadow pixel bounds
- `TitState` and `Transcript`-equivalent do not exist in TIT (TIT is stateless across runs — no session persistence) — one less risk vector than CAROLINE
- `MenuBuilder` must be pure — given same `TitState` VT, same `Array<MenuItem>` output, always. No timing dependencies.
- `jreng_subprocess` streaming callbacks must deliver chunks on the subprocess thread, NOT on message thread — `ConsoleStream`'s atomic writer handles the cross-thread delivery on flush

---

## 6. Open Questions

### 6.1 `paint()` overload/shadow resolution (caroline sprint — handoff.md §Decision point)

Resolved in caroline's rename sprint. Three options (A / B / C) surfaced in `handoff.md`; ARCHITECT direction favors A (JUCE-lingua overload). COUNSELOR + Auditor in caroline's sprint validate against `-Woverloaded-virtual`. TIT-cpp inherits whatever decision caroline ships.

### 6.2 `jreng_svg_braille` — reuse `jreng_core/xml/jreng_svg` or port-from-Go directly?

Caroline's `jreng_core` already contains `xml/jreng_svg.h/.cpp` — SVG path parsing likely present. The Go TIT port target (`internal/banner/svg_paths.go` — 351 LOC) may be partially solved upstream. COUNSELOR assesses during `jreng_svg_braille` implementation planning:

- **Option A:** Use `jreng_core::xml::svg` for path parsing; implement only `rasterizer/` + `braille/` in `jreng_svg_braille`. Likely saves 30–40% of port effort.
- **Option B:** Port all 4 Go files verbatim for clean isolation. Simpler but potentially redundant.

### 6.3 File renaming in `jreng_tui` post-rename

Caroline's current file names (`jreng_terminal_metrics.*`, `jreng_terminal_input.*`, `jreng_terminal_rectangle.*`, `jreng_ansi_*`) use "terminal" and "ansi" as **domain terms**, not namespace references. File renames are optional — namespace/method rename (`Terminal` → `tui`, `render()` → `paint()`) is the load-bearing change. Caroline's COUNSELOR decides file-rename scope during their sprint.

### 6.4 Theme XML schema

No existing theme XML schema exists. COUNSELOR drafts schema in PLAN.md — at minimum covers all 10 color fields from Go TIT's `theme.go`:

- `status_clean`, `status_modified`, `status_conflict`
- `timeline_sync`, `timeline_ahead`, `timeline_diverged`
- `menu_selected` (bg + fg)
- `border`
- Destructive-action marker color

Hot reload via `juce::FileWatcher` is desired but not required for MVP.

### 6.5 Windows MSYS2 parity — MVP or post-MVP?

**Proposed:** macOS-first MVP. Windows parity as a named post-MVP sprint. Rationale: caroline is also macOS-first (RFC-CAROLINE-00 targets macOS primary, MSYS2 secondary). `jreng_subprocess` Windows-native `CreateProcess` + anonymous pipe path is designed in RFC-CAROLINE-00 §4.5 but unbuilt. Adding Windows to TIT-cpp MVP means forging both POSIX and Windows `jreng_subprocess` paths simultaneously. Splitting into two sprints (macOS-first, then Windows) matches caroline's precedent and keeps MVP timeline realistic.

**ARCHITECT to confirm.**

### 6.6 Binary distribution — local CMake only for MVP

`cmake --install` to `/usr/local/bin/tit` (macOS) or user-local equivalent. No Homebrew tap, no releases, no notarization pipeline at MVP. Distribution polish is post-MVP sprint.

---

## 7. Handoff Notes

### 7.1 Prerequisite sprint — caroline rename

`handoff.md` written to `~/Documents/Poems/dev/caroline/handoff.md` (103 lines). Caroline's COUNSELOR owns that sprint:

1. Rename `namespace jreng::Terminal` → `namespace jreng::tui` across `modules/jreng_tui/`
2. Rename `render(Graphics&)` → `paint(...)` (A/B/C decision resolved in-sprint)
3. Update RFC-CAROLINE-00 §4.7 text references
4. Caroline builds green
5. Sprint logged

After caroline's rename sprint ships, TIT-cpp Phase 0 forks the clean `jreng_tui` module. Zero drift at fork boundary.

### 7.2 Go TIT remains shippable via `___legacy___/`

After port begins:
- Go TIT continues to build via `go build ./___legacy___/cmd/tit` — zero breakage
- Users on unsupported platforms (Windows before MSYS2 sprint, Linux never officially supported) continue using Go binary
- No behavioral regression risk — Go sources preserved intact
- When C++ TIT reaches parity, Go sources archived or deleted per ARCHITECT call

### 7.3 Reference document precedence

1. This RFC — port scope, architecture, phase sequence
2. `SPEC.md` (existing) — requirements contract, **unchanged** from Go TIT
3. `___legacy___/ARCHITECTURE.md` — Go implementation architecture, reference only
4. `___legacy___/CODEBASE-MAP.md` — Go code topology, reference only
5. `~/Documents/Poems/dev/end/Source/terminal/data/State.h` — `TitState` template
6. `~/Documents/Poems/dev/end/Source/terminal/logic/Parser.h` — `GitStateDetector` template
7. `~/Documents/Poems/dev/caroline/RFC-CAROLINE-00.md` — `jreng_subprocess` spec (§4.5), `jreng_svg_braille` spec (§4.8)

### 7.4 Estimate confidence

**7–11 day MVP** is load-bearing and defensible. Confidence rationale:

- Zero architectural invention
- Full reference implementation in Go
- Architectural template in END (proven at VT520 byte rate)
- Framework base in caroline (3,384 LOC primitive infrastructure implemented)
- Two new modules (`_subprocess`, `_svg_braille`) have spec + port source pre-specified
- CAROL parallelism — primitive forge, state layer, git layer, protocols can overlap across concurrent sprints once Phase 0 scaffolds

**Estimate floor** (7 days) assumes sustained CAROL cadence + zero Windows parity work + macOS-only subprocess impl + jreng_svg_braille reusing `jreng_core::xml::svg`.

**Estimate ceiling** (11 days) assumes full Go-port fidelity for `jreng_svg_braille` + first-primitive (`Menu` or `ListPane`) design iteration once + one conflict resolver view redesign pass.

### 7.5 Post-MVP queue

- Windows MSYS2 parity sprint (+3–5 days)
- Theme hot-reload
- Homebrew tap + CI/CD release pipeline
- `jreng_subprocess` enhancements (stdin write for rare passphrase case)
- Potential `jreng_git` extraction if CAKE-cpp / whatdbg-TUI consume similar subprocess+parser patterns

### 7.6 Sprint ownership

COUNSELOR owns PLAN.md generation from this RFC. Suggested sprint breakdown:

- **Sprint 1:** Phase 0 (scaffold + forks + implement `jreng_subprocess` + port `jreng_svg_braille`)
- **Sprint 2:** Phase 1 (`jreng_tui` extensions — 8 primitives against fixtures)
- **Sprint 3:** Phase 2 + Phase 3 (`TitState` + views — demo milestone)
- **Sprint 4:** Phase 4 + Phase 5 (git layer + protocol FSMs)
- **Sprint 5:** Phase 6 (integration + polish + release)

Each sprint ends with `log sprint` + `carol debt clear` per protocol.

---

## 8. Closing

TIT-cpp is the last daily-driver tool converging into the `jreng_*` vertical-integration stack. Shares substrate with END (daily-driver terminal), whatdbg (daily-driver debugger), CAROLINE (CAROL-native client), Kuassa (commercial audio plugin). One terminal, one debugger, one code editor, one substrate, native binary everywhere.

The port isn't "rewrite TIT." It's **use TIT's battle-tested spec as the forge for `jreng::tui`'s composition-level primitives, ship `jreng_subprocess` + `jreng_svg_braille` back to the ecosystem, and harvest a cleaner TIT in the process.**

Ready for COUNSELOR.

---

*RFC complete. Status: Ready for COUNSELOR handoff.*

**Rock 'n Roll!**
**JRENG!**
