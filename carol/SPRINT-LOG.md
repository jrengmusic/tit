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

## Sprint 19: TIT Part 2 — SPEC.md Rewritten as the Contract of the C++ TIT ✅

**Date:** 2026-11-15
**Duration:** not recorded

### Agents Participated
- COUNSELOR: Claude Opus — Read the tracker canon, the old SPEC, and Parts 10 to 13. Wrote `SPEC.md` (Plan Steps 1 to 9). Validated and applied the Auditor findings. Fixed the contradictions between SPEC, Part 2, and Part 13.
- Pathfinder: Haiku — Verbatim reads of the legacy string maps, key handlers, dispatchers, menus, dialogs, LFS, header, config, and state detection.
- Auditor: Opus — One sweep, 30 findings (23 accepted and fixed, 7 rejected with citations).

### Files Modified (4 total)
- `SPEC.md` — rewritten whole. C++17 / JUCE 8 / JAM / cast stack. Nine sections, each numbered once. Invariants 7, 8, 10, 11 corrected and invariant 13 added (ruling 9). Nine Operation rows with `Pinned`. Marker table with stale rules and ruling 54. Every legacy menu item with emoji, label, shortcut, hint. Pinned, DirtyOperation (ruling 41), Conflicted (resolver, no menu) menus. Startup gate, auto-update, LFS, console, text inputs (`jam::TextEditor`). Dirty Operation Protocol with one stash label. Time travel return with two stashes in order. One history browser with five feeds (rulings 19, 39, 40). Conflict resolver labels. Pin, pull with rebase, rebase onto, interactive rebase. Header texts. Error handling table. Banner and `--help` (rulings 23, 24, 33). Command table, key table, owner-key table, dialogs, footer, error, output, and console texts. User files `user.md`, `keys.md`, `themes/<name>/theme.md` (rulings 17, 38, 43, 45 to 47). No layout drawing (ruling 7).
- `PLAN-tit-2-spec.md` — the citations of `carol/SPRINT-LOG.md` Sprints 2, 3, 5, 8, 10 replaced with legacy code lines (that log holds Sprints 14 to 18 only). Step 4 table: Conflicted and DirtyOperation rows follow ruling 41 and Part 12 items 4, 6, 8.
- `PLAN-tit-13-setup.md` — Closed by a read, item 19: the narration lines of a chain show in the footer (ruling 3; the console holds only the bytes of the tool, ruling 58). Item 22: `TimelineSyncMessages` listed in SPEC 6.6 as never shown. Step 3 item 6: eight narration rows added to `## footer`.
- `PLAN-tit-0-tracker.md` — Status row 2.

### Alignment Check
- [x] BLESSED principles followed — SPEC states behavior and data, no layout and no class (Encapsulation). One table for commands, keys, markers, dialogs (SSOT).
- [x] NAMES.md adhered — new dialog key `time_travel_return_dirty_stash` takes the family of Part 12 `timeTravelReturnDirtyStash`. No `kind`.
- [x] MANIFESTO.md principles applied — every string traced to a legacy literal, a ruling, or a part text.
- [x] ASD-STE100 — the Auditor's over-long and passive sentences rewritten.

### Problems Solved
- SPEC named Go, Bubble Tea, Lip Gloss, XML themes, and `IMPLEMENTATION_PLAN.md`. Replaced by the C++ stack and the markdown config files.
- SPEC said "no abort in menu" while its own Section 5 and the code gave aborts. Invariants corrected.
- SPEC said `ssh-keygen -t ed25519`; the code runs `rsa -b 4096` (`legacy/git/ssh.go:37`). SPEC follows the code.
- Stashes had three labels in the first draft. One label, `TIT DIRTY-OP SNAPSHOT` (Part 12 item 2).
- A dirty time travel would detect as `DirtyOperation`. `DirtyOperation` now needs HEAD on a branch (Part 12 item 3).
- Interactive rebase while pinned contradicted ruling 8. It is inactive while `Pinned`. Rewind or time travel while pinned ends the pin; SPEC states it.
- Part 2 Step 8 carried the output texts; Part 13 item 19 dropped them. Legacy reads them (78 reads in 14 files), so ruling 3 keeps them; they show in the footer.

### State for Continuation
- Part 2 changes no code. No build.
- Six SPEC choices that differ from the Part 2 plan text are recorded in that plan's Step 4 table and in this entry: DirtyOperation and Conflicted menus; dirty time travel stashes with no second question (`SPEC.md` old lines 677-681); quit timeout 2s and floor 70×20 (`PLAN-tit-10-views.md:438,443`); output texts in the footer.
- Parts 10 to 13 cite old `SPEC.md` line numbers (for example `tit/SPEC.md:350-357`). Those lines now hold the new text. Each part that executes reads its cited SPEC section by name, not by line.

### Debts Paid
- None (no DEBT.md at project root)

### Debts Deferred
- None

---

## Sprint 18: Archetype Repair + TIT Part 1 (titc Cast Project) ✅

**Date:** 2026-11-15
**Duration:** not recorded

### Agents Participated
- COUNSELOR: Claude Opus — Read the working cast projects (cast, whatdbg, scout, ling, JFS), corrected PLAN-tit-1-project.md, validated each Engineer return, filtered Auditor findings, fixed the `Application` constructor.
- Pathfinder: Haiku — Verbatim reads of the cast, whatdbg, scout, ling, ggwp, end and JFS templates, tables and installer resources.
- Engineer: Sonnet — Archetype template and INIT.md rewrite, bitmap lane removal across projects, ggwp self-containment, TIT Part 1 Steps 1–10, audit fixes.
- Auditor: Opus — One sweep, 16 findings (7 accepted and fixed, 9 rejected with citations).

### Files Modified (approx. 60 total, across 9 projects)

**archetype (`~/Documents/Poems/dev/archetype/`):**
- `cli/cast/cmake.cast` — replaced with `cast/cast/cmake.cast` verbatim. The archetype data was stale.
- `cli/cast/installer.cast`, `app/cast/installer.cast` — taken from cast's `installer.cast`. Removed `MUI_HEADERIMAGE`, `MUI_HEADERIMAGE_BITMAP` and `MUI_WELCOMEFINISHPAGE_BITMAP`.
- `cli/`, `app/` installer resources — deleted `header.bmp` and `welcome.bmp`.
- `cli/INIT.md`, `app/INIT.md`, `plugin/INIT.md`, `plugin-only/INIT.md` — template cells are now local to the project (`cmake.cast`, `installer.cast`, `signing.cast`, `installer-mac.cast`, `installer-win.cast`). Added the `juceCastPath` and `modulesInfoPath` cells to app, plugin and plugin-only.
- `cli/init.cast` — fences rebuilt from cast's `project-info.md`, `spell.md` and `signing.md`. `## macro` replaces `## define`. `## architecture` dropped. Dead `CMakeLists.txt` `## headers` record removed. `installerResourceDirectory` is local (`init.cast:99`).
- `plugin/init.cast`, `plugin-only/init.cast` — `@source` row is local `source.cast`. `app`, `plugin`, `plugin-only` `init.cast` use `:::modulesInfoPath:::` and `:::juceCastPath:::`.
- `ARCHETYPE.md:25,96-104` — artwork line and the rule that the project holds its own templates.

**cast, whatdbg, scout, ling, stamp, end:**
- `cast/installer.cast` in each — removed the three MUI bitmap lines.
- `cast/installer/resources/` in each — deleted 14 `header.bmp` / `welcome.bmp` files in total.

**ggwp:**
- `cast/cmake.cast`, `cast/installer.cast` — own copies. Project is self-contained.
- `cast/spell.md` — index rows local. `@CMakeLists` group equals `cast/cast/spell.md:43-82`. Dead `CMakeLists.txt` `## headers` record removed.
- `project-info.md` — `description` → `projectDescription`, `architecture` row added, `## architecture` table removed, `## define` → `## macro`, empty `## library` added. Release-notes separators added.

**tit:**
- Deleted `CMakeLists.txt`, `build.sh`, `install.sh`, `Source/` (old tree), `tests/`, `themes/`, `Builds/Ninja/` (Step 1).
- `project-info.md` — generated by `cast INIT.md`. Values, juce modules (scout `108-113`), 14 user modules (scout `121-134`), macro, include and binary rows set.
- `cast/cmake.cast`, `cast/installer.cast`, `cast/installer/resources/background*.png` — project-owned templates and placeholder artwork.
- `cast/signing.md` — six LING values. No `disable-library-validation` entitlement.
- `cast/identifiers.md`, `cast/text.md`, `cast/files.md`, `cast/source.cast`, `cast/spell.md` — LING-pattern tables and wiring. `@string` and `@banner` absent.
- `LICENSE` — byte-equal to ling's.
- `Source/HELP.md` — titc help body.
- `Source/Main.cpp` — `printError`, `writeVersion`, `writeHelp` (via `jam::TerminalPage`), `isTerminalOutput`, `getFlags`, `getFlagWord`, `runCommandLine`, `class Application`. `Main.cpp:104` — `Application() = default;` added after the build error (the deleted copy constructor suppressed the implicit default constructor).
- `PLAN-tit-1-project.md` — inventory, template ownership, Step 2/4/5/8/9/10 line citations, BLESSED line corrected.

### Alignment Check
- [x] BLESSED principles followed — each project owns its templates (Encapsulation). The templates are SSOT copies of the working cast project.
- [x] NAMES.md adhered — no new names. All names come from the LING family.
- [x] MANIFESTO.md principles applied — `jam::TerminalPage` used for help output. The framework API is used and not written again.

### Problems Solved
- Archetype data was stale. Its templates did not match the working projects. Replaced from cast verbatim.
- Archetype templates linked to the archetype directory and broke encapsulation. Every cell now names a project-local file. INIT copies nothing.
- The MUI bitmap lane (`header.bmp`, `welcome.bmp`) was dead. Removed from the archetype and from six projects. JFS is the installer canon.
- ggwp depended on archetype templates. It now holds its own.
- TIT had the old jam::tui tree. Replaced with a cast project that builds `titc` with `--version`, `--help` and an unknown-flag diagnostic.
- `START_JUCE_APPLICATION (Application)` failed: no default constructor. Fixed at `Source/Main.cpp:104`.

### State for Continuation
- Rebuild `tit` with `cast cast/spell.md --debug`. The last build failed at `Main.cpp:141`. The fix at `Main.cpp:104` is not yet compiled.
- Then run `titc --version`, `titc --help`, `titc`, `titc --wrong`. Expect exit 0, 0, 0, 1. A second cast run must change nothing.
- Windows: no archetype or project change is built or tested there.
- The archetype and the six projects are not rebuilt after the installer bitmap removal.
- The JFS custom-page NSIS port for the cli and app archetypes is not done. Only the bitmap lines are removed.
- `PLAN-tit-0-tracker.md` Status is not updated. Its inventory row that says `jam_tui` does not exist (near line 203) is stale. `jam/jam_tui/page/jam_TerminalPage.h` exists.
- `tit/RELEASE.md` is the INIT stub. The first cast run regenerates it.

### Debts Paid
- None (no DEBT.md at project root)

### Debts Deferred
- None

---

## Sprint 17: Foundation Reset — Cell Unit Type + LookAndFeel + View Stubs ✅

**Date:** 2026-04-22
**Duration:** ~2 hours

### Agents Participated
- COUNSELOR: Claude Opus — Planning, CONTRACT enforcement, delegation, namespace/naming violation discovery
- Engineer: Multiple — Cell migration across jam_tui (26+ call sites), LookAndFeel stub, view stub rewrites, MainComponent rewrite, Main.cpp simplification, Banner SVG asset extraction, namespace tit removal
- Pathfinder: Multiple — jam_tui theming survey, juce::LookAndFeel API surface extraction, module structure survey, bare-int Rectangle call-site audit, git/sprint-log state
- Auditor: Not invoked (mechanical migration, no complex logic)

### Files Modified

**jam_tui framework (17 files across ~/Documents/Poems/dev/jam/jam_tui/):**
- `graphics/jam_tui_rectangle.h` — `jam::literals::Cell` strong unit type (struct + arithmetic + comparison + `_cell` UDL); all Rectangle constructors, accessors, fluent builders, slicing methods take/return `literals::Cell`; `getCellWidth()`/`getCellHeight()` convenience accessors; namespace cleanup (`jam::literals::Cell` → `literals::Cell` inside `namespace jam::tui`); parameter name cleanup (`x_`/`y_` → `x`/`y`)
- `ansi/jam_tui_component.h` — `getCellBounds()`/`setCellBounds()` inline wrappers; `setLookAndFeel()`/`getLookAndFeel()` with parent-chain walk + `lookAndFeelChanged()` override
- `lookandfeel/jam_tui_look_and_feel.h` — NEW: `jam::tui::LookAndFeel` stub mirroring juce::LookAndFeel surface API (`findColour`/`setColour`/`isColourSpecified` + static default instance)
- `lookandfeel/jam_tui_look_and_feel.cpp` — NEW: implementation
- `jam_tui.h` — LookAndFeel include (before component.h for dependency order)
- `jam_tui.cpp` — LookAndFeel cpp include
- `ansi/jam_tui_graphics.cpp` — Cell migration (`.value` extraction where raw int needed)
- `ansi/jam_tui_screen.cpp` — Removed spurious `.value` on juce::Rectangle accessors
- `ansi/jam_tui_textbox.cpp` — Same
- `metrics/jam_tui_metrics.h` — Cell wrapping on `getBounds()` return
- `markdown/jam_tui_markdown_renderer.cpp:79,96,142,161` — Cell wrapping on 4 Rectangle constructions
- `primitives/jam_tui_menu.cpp:177,193,194` — Cell wrapping
- `primitives/jam_tui_listpane.cpp:155,168,193,205,222` — Cell wrapping
- `primitives/jam_tui_consolestream.cpp:239,250,265` — Cell wrapping
- `primitives/jam_tui_textpane_paint.cpp:28,76,92,106,112,126,164,174,186` — Cell wrapping (9 sites)
- `primitives/jam_tui_dialog.cpp:156,172,196,213,216` — Cell wrapping
- `primitives/jam_tui_splitpane.cpp:61,67` — Cell wrapping

**tit project (30+ files):**
- `Source/Main.cpp` — ThemeResolver removed, MainComponent construction simplified to single `juce::ValueTree` arg
- `Source/MainComponent.h` — Stub: owns `juce::ValueTree stateTree` + `jam::tui::LookAndFeel lookAndFeel`, no child components
- `Source/MainComponent.cpp` — Stub: constructor stores VT + calls `setLookAndFeel()`
- `Source/component/` — NEW directory (renamed from `Source/view/`)
- `Source/component/Banner.h` + `.cpp` — Blank stub
- `Source/component/Header.h` + `.cpp` — Blank stub
- `Source/component/Footer.h` + `.cpp` — Blank stub
- `Source/component/ConfirmDialog.h` + `.cpp` — Blank stub
- `Source/component/Menu.h` + `.cpp` — NEW (renamed from MenuView, NAMES.md Rule 2)
- `Source/component/Console.h` + `.cpp` — NEW (renamed from ConsoleView)
- `Source/component/History.h` + `.cpp` — NEW (renamed from HistoryView)
- `Source/component/FileHistory.h` + `.cpp` — NEW (renamed from FileHistoryView)
- `Source/component/ConflictResolver.h` + `.cpp` — NEW (renamed from ConflictResolverView)
- `Source/component/SetupWizard.h` + `.cpp` — NEW (renamed from SetupWizardView)
- `Assets/tit_logo.svg` — NEW (extracted from Banner.cpp inline string)
- `CMakeLists.txt` — `juce_add_binary_data(titc_assets)` for SVG asset
- 28 Source files — `namespace tit {}` wrappers and `tit::` qualifications removed

**Deleted:**
- `Source/view/` — entire directory (MenuView, ConsoleView, HistoryView, FileHistoryView, ConflictResolverView, SetupWizardView old files)

### Alignment Check
- [x] BLESSED principles followed (B: MainComponent owns VT+LAF, clear lifecycle; L: stubs are minimal; E: Cell type prevents bare-int/cell mixing at compile time, no magic values; S: LookAndFeel is SSOT for colours via parent-chain walk; E: Component LAF mirrors juce pattern — tell don't ask)
- [x] NAMES.md adhered (Rule -1: Cell, LookAndFeel, all renames ARCHITECT-approved; Rule 2: "View" suffix dropped from 6 components — type information in name)
- [x] MANIFESTO.md principles applied (JUCE-native thinking replaces ELM port patterns; ThemeResolver eliminated in favor of LookAndFeel mirroring juce::LookAndFeel; callback `std::function<void(MenuItem&)>` identified as architecture violation — removed with stub reset)

### Problems Solved
- `jam::literals::Cell` strong unit type eliminates bare-int leaking at Rectangle call sites — compile-time enforcement, no runtime cost
- `jam::tui::LookAndFeel` stub establishes JUCE-native theming pattern (set on root, children inherit) replacing ELM-ported ThemeResolver (passed by reference through every constructor)
- `namespace tit` removed — unnecessary wrapper around classes already scoped by project
- "View" suffix on 6 component names violated NAMES.md Rule 2 (type information in name)
- `Source/view/` → `Source/component/` — directory name matches what the files contain
- All view components stripped to blank stubs so ARCHITECT can rebuild foundation by hand with correct JUCE-native patterns
- `onItemSelected` callback on MainComponent identified as BLESSED violation (bypasses ValueTree) — removed with stub reset

### Debts Paid
- None (no DEBT.md at project root)

### Debts Deferred
- None

---

## Sprint 16: Phases 1–3 + UI Bootstrap + jam Migration ✅

**Date:** 2026-04-18 / 2026-04-19
**Duration:** ~1 day (multi-session)

### Agents Participated
- COUNSELOR: Claude Sonnet — Planning, SPEC/PLAN edits, delegation
- Engineer: Multiple — Phase 1–3 implementations, bug fixes, file rename, migration, bootstrap wire
- Auditor: Multiple — Per-step BLESSED/NAMES/STANDARD validation (Sprint 2.2/2.3/2.4, Sprint 3.2/3.3/3.4/3.6)
- Pathfinder: Multiple — Codebase inventory, jam vs TIT diff, commit message extraction
- Explore: Deep Go ground-truth survey (shipping behavior for SPEC alignment)

### Files Modified
**Phase 1 — jam_tui primitives (Sprint 2):**
- `modules/jam_tui/primitives/jam_tui_menu.{h,cpp}` — Menu (VT-driven, MENU/ITEM[] schema, listener)
- `modules/jam_tui/primitives/jam_tui_listpane.{h,cpp}` — ListPane (scrollable, attribute+content columns, selectedIndex SSOT)
- `modules/jam_tui/primitives/jam_tui_splitpane.{h,cpp}` — SplitPane (2-pane, Tab focus cycling)
- `modules/jam_tui/primitives/jam_tui_dialog.{h,cpp}` — Dialog (public Prop* statics, context substitution, variant-count overload)
- `modules/jam_tui/primitives/jam_tui_consolestream.{h,cpp}` — ConsoleStream (callAsync atomic writer, autoscroll)
- `modules/jam_tui/primitives/jam_tui_textpane.{h,cpp}` + `_diff.cpp` + `_paint.cpp` — TextPane (split across 3 TUs per L; DiffLineType enum SSOT; lookup-map dispatch)
- `modules/jam_tui/primitives/jam_tui_spinner.{h,cpp}` — Spinner (10 braille frames, juce::Timer)
- `modules/jam_tui/primitives/jam_tui_theme_resolver.{h,cpp}` — pure accessor, fail-loud `getInt` (isInt guard)
- `tests/TestRunner.cpp` — `juce::UnitTestRunner` driver (`titc_tests` target)
- `tests/MenuTests.cpp`, `ListPaneTests.cpp`, `SplitPaneTests.cpp`, `DialogTests.cpp`, `ConsoleStreamTests.cpp`, `TextPaneTests.cpp`, `SpinnerTests.cpp`, `ThemeResolverTests.cpp`

**Phase 2 — TitState + state:**
- `Source/TitIdentifier.h` (155 LOC, 67 `juce::Identifier` — SSOT for RFC §4.6 VT schema + themes/default.xml attributes)
- `Source/state/TitAxis.{h,cpp}` (36 + 155 LOC — 6 enum axes + 12 bridge functions, `.at()` + jassert on parse)
- `Source/state/TitState.{h,cpp}` + `TitState_flush.cpp` — APVTS-mirror (atoms + Timer, skip-unchanged, `writeIfChanged<T>` template)
- `tests/TitStateTests.cpp` — 4 tests (shape, message-thread, cross-thread, skip-unchanged)

**Phase 2 — Menu:**
- `Source/menu/MenuItems.h` — 27 `inline constexpr MenuItemDef` (verbatim from Go `menu_items.go`)
- `Source/menu/MenuBuilder.{h,cpp}` — 8 generators, `generators.at(op)` dispatch
- `tests/MenuBuilderTests.cpp` — 10 tests (determinism, coverage, per-Operation shapes)

**Phase 3 — Views:**
- `Source/view/TitScreen.{h,cpp}` — root composition
- `Source/view/Banner.{h,cpp}` — braille SVG banner
- `Source/view/Header.{h,cpp}` — REPO-observing
- `Source/view/Footer.{h,cpp}` — operation-state-observing
- `Source/view/MenuView.{h,cpp}` — MenuBuilder + MENU subtree write
- `Source/view/HistoryView.{h,cpp}`, `FileHistoryView.{h,cpp}`, `ConflictResolverView.{h,cpp}`, `ConsoleView.{h,cpp}`, `ConfirmDialog.{h,cpp}` (7 variants), `SetupWizardView.{h,cpp}` (6 phases)
- `tests/RootViewsTests.cpp`, `tests/BrowserViewsTests.cpp`

**UI bootstrap:**
- `Source/TitApp.{h,cpp}` — full bootstrap: TitState, ThemeResolver, TitScreen, `jam::tui::Screen`, `jam::tui::Input`. Theme file-first load, hardcoded fallback. Shutdown reverses construction.

**Bug fixes (framework-level, land in jam):**
- `jam::tui::Screen::renderChildren` — replaced with recursive `renderComponentTree` (was only iterating direct children)
- `Source/view/TitScreen.cpp` — CtrlC/CtrlD → `systemRequestedQuit()`

**Theme:**
- `themes/default.xml` — 13 component families × 44 color attributes (GFX default, mirrors Go `theme_gfx.go`)

**SPEC.md:**
- §6 Normal menu: Timeline tables corrected (InSync empty; Clean/Dirty variants; branch ops removed from Normal)
- §6.3 + §6.4 new (Config Menu, Preferences Menu)
- §6 TimeTraveling: dual-path Return
- §8 rewritten (Branch Picker CRUD)
- §9 compressed
- §10 subsection numbers fixed
- §13 Setup: 7 wizard phases enumerated
- §14 Keyboard §14.3–14.6 added
- §16 theme: toml → xml
- Post-§18 duplicate deleted

**File rename (pre-migration):** 16 files `jreng_ansi_*` → `jam_tui_*` (convention: filename mirrors namespace path; subdir organizational only)

**Build infrastructure:**
- `build/` → `Builds/Ninja/`
- `build.sh` ported from END (no install; MSYS2 zsh on Windows — no .bat needed)

**jam migration (2026-04-19):**
- Deleted `modules/` entirely (97 tracked + 18 untracked)
- `CMakeLists.txt` → `$HOME/Documents/Poems/dev/jam/jam_*/`
- Source + tests: `jreng::*` → `jam::*` (zero jreng refs)
- `SPEC.md`, `PLAN-tit-cpp-port.md`, `RFC.md` swept

**PLAN amendments:**
- Sprint 2 file placement corrected (`primitives/` subdir, both .h + .cpp co-located)
- Sprint 3 demo milestone → COMPONENT+STATE (dispatch deferred to Sprint 4)
- Addendum 2026-04-19 at top (jam migration note)
- Title + forward-looking text → `jam::tui`

### Alignment Check
- [x] BLESSED principles followed (multiple audit rounds per step; every early-return, shadow-state, magic-value, 30/3 violation surfaced and closed)
- [x] NAMES.md adhered (Rule -1: every new name traces to SPEC / RFC / Go source / ARCHITECT approval; SCREAMING_SNAKE for file-scope constants; PascalCase classes; camelCase members)
- [x] MANIFESTO.md principles applied (B thread ownership via std::unique_ptr + callAsync crossing; L 300/30/3 decomposed via TU splits; E zero early returns across all new code; S SSOT for VT keys via TitIdentifier.h + Dialog::Prop*; D determinism via pure-function MenuBuilder + VT-driven views)

### Problems Solved
- Byte-identity divergence in braille relaxed to visual parity (ARCHITECT 2026-04-18) — JUCE tessellator vs Go `approximateCubicBezier(20)` per-cell differs but shape preserved
- `juce::FileWatcher` hallucination corrected — forked `FigBug/Gin` as `jam::File::Watcher` nested class in `jam_core`
- `paint()` overload/shadow collision in caroline Sprint 5 masked rather than resolved — closed via RFC §6.1 explicit nullify (`override final {}`) on ansi Component
- SPEC vs shipping Go drift identified through deep codebase survey (Explore agent); corrected across 10 areas
- UI launch bugs: recursive renderComponentTree + CtrlC interception
- SSOT violation in Dialog VT keys (4 duplicated sites) → public `jam::tui::Dialog::Prop*` statics
- 5 early returns in paint methods (Sprint 2.2), 3 in keyboard dispatchers (Sprint 2.3) — all rewritten positive-nested
- TextPane 505 LOC over L → split across 3 TUs, each ≤300
- TitAxis.cpp silent-fail `.find()` fallback → `.at()` + `jassert` (fail-loud)
- Early-return in CMake `_tit_pre_generate_juce_header` closed via positive-nested rewrite
- Anonymous-namespace sweep across 8 test files (Sprint 2.3 audit) — `static` linkage replacement
- `ConfirmDialog` unused `ThemeResolver` field + unused `isSelected` shadow on `ListItem` (Sprint 2.2 + 3.6 audit)
- Stale `TIME_TRAVEL_MERGE` comment in MenuBuilder corrected against verified Go source (menu_render_extra.go)
- SPEC TIME_TRAVEL_MERGE misread corrected: Go HAS time_travel_merge as confirmation dialog type, not as menu-emitted item — comment updated, SPEC left consistent with shipping behavior

### Debts Paid
- None (all in-scope items resolved before log; no `DEBT.md` entries this session)

### Debts Deferred
- None

---

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
