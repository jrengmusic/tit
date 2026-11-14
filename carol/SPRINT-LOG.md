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

## Sprint 20: TIT Part 3 — TerminalDisplay and the Inline Image Output Lane ✅

**Date:** 2026-11-15
**Duration:** not recorded

### Agents Participated
- COUNSELOR: Claude Opus — Read the tracker and Part 3. Found the help files gone and updated the plan to `jam::TerminalPage` on ARCHITECT's instruction. Wrote each Engineer brief, validated each return, filtered the Auditor findings, and kept the plan text equal to the code.
- Pathfinder: Haiku — Verbatim old bodies at JAM `7fbf821ba`. Help baselines of `ling`, `cast`, `scout`, `titc`. Git history of `isReadSourceSuspended`.
- Librarian: Haiku — libdispatch and Win32 thread pool shutdown documents (quotes with URLs).
- Engineer: Sonnet — Plan Steps 1 to 8, the audit fixes, the TTY restore, the EVE form of the display backends.
- Auditor: Opus — One sweep, 15 findings: 6 fixed (2, 4, 5, 6, 8, 15), 9 rejected (1, 3, 7, 9, 10, 11, 12, 13, 14).

### Files Modified (15 total)
**JAM (`~/Documents/Poems/dev/jam/`):**
- `jam_terminal/buffers/jam_TerminalBuffer.h` — `template <typename Content> struct Placement` with `ImagePlacement` and `InlineImagePlacement` aliases (`:45-58`). `explicit TerminalBuffer (bool newShouldUseInlineImages = false)`. Public `addInlineImagePlacement`. Private `shouldUseInlineImages` and `inlineImagePlacements`. `writeFrame` passes the row index and the OUTPUT list of the slot. The INPUT struct and its out-of-class `operator<` are deleted (Step 1, delete first).
- `jam_terminal/buffers/jam_TerminalBuffer.cpp` — the constructor (`:5-8`). `inlineImagePlacements` handled in `prepare`, `prepareRows`, `clear (writer)`, `swap`, `clear (row)`. `addInlineImagePlacement` (`:212-234`): normal-screen assert, bounds assert, then the restored body inside `if (shouldUseInlineImages)`. The VT lane writes the newline between two rows and none after the last (ruling 31). `jassert (placements.isEmpty())` deleted.
- `jam_terminal/buffers/jam_TerminalEncoder.h`, `.cpp` — `writeRow` takes the OUTPUT placements and ends with the SGR reset only. New: `writeRowSpan`, `writeInlineImage`, `writeImageArgument` (template, header), `writeCursorUp`, `writeCursorForward`, and the public `writePrivateMode`. Bodies restored from `7fbf821ba` with the `write…` verb.
- `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:488-496` — `drawImage`: the half-block body deleted (ruling 28). It tells the buffer `addInlineImagePlacement` when the clip holds the destination, no transparency layer is open, and the fill is opaque.
- `jam_terminal/display/jam_TerminalDisplay.h` — new abstract seam in the `TerminalTTY` form: `onBytesReady`, `open (context, bufferSize, maximumPeriodMilliseconds)`, `close`, `read`, `write`, `getWinsize`.
- `jam_terminal/display/native/jam_TerminalDisplay_posix.h`, `.cpp` — `TerminalUnixDisplay`: raw mode that keeps the saved output flags, `VMIN` and `VTIME` 0, `std::optional<struct termios> savedMode`, `addSources` (read source on stdin, `SIGWINCH` source), the read handler of `TerminalUnixTTY` with `jam::TerminalClock`.
- `jam_terminal/display/native/jam_TerminalDisplay_windows.h`, `.cpp` — `TerminalWindowsDisplay`: console modes as `std::optional<DWORD>`, `readWait` with `jam::TerminalClock`, `read()` re-arms the wait after it drains, surrogate pairs only with a low surrogate, file-static `getUnicodeChar`, the tested close form.
- `jam_terminal/jam_terminal.h` — `jam_TerminalEncoder.h` after `jam_TerminalBuffer.h`; the three display headers after the TTY headers; module brief and include order note.
- `jam_terminal/jam_terminal.cpp:38-42` — display backend sources under the platform guards.
- `jam_tui/page/jam_TerminalPage.cpp:37` — one newline after the frame (ruling 31).

**TIT:**
- `PLAN-tit-3-terminal-display.md` — the help site is `jam::TerminalPage`; the audit corrections; Steps 5 to 7 in the EVE form; the switch has no getter.
- `PLAN-tit-0-tracker.md` — Part 3 depends on Part 0 only; Status row 3.

`jam_terminal/tty/native/jam_TerminalTTY_posix.h`, `.cpp` were changed and then restored. `git status --short -- jam_terminal/tty/native/` is empty.

### Alignment Check
- [x] BLESSED principles followed — Tell, don't ask: the buffer owns its switch, and no getter exists (MANIFESTO E rule 4). One template for both placements (CODING.md, near-duplicate structs). The display backends are the EVE / `TerminalTTY` form (ruling 53).
- [x] NAMES.md adhered — every new name has a family in the plan's "New names" lines: `Placement`, `InlineImagePlacement`, `inlineImagePlacements`, `addInlineImagePlacement`, the `write…` writers, `addSources` (`addReadSource`), `getUnicodeChar`, `clock`, `maximumPeriodMilliseconds`.
- [x] MANIFESTO.md principles applied — delete first; no comment and no doxygen before the doxygen pass.

### Problems Solved
- The plan named `ling/Source/Help.cpp` and `tit/Source/Help.cpp`. Neither exists; each help goes through `jam::TerminalPage`. ARCHITECT: *"update the plan. i have already told you this explicitly on previous session jam::TerminalPage"*. The newline of ruling 31 is at `jam_TerminalPage.cpp:37`.
- `drawImage` asked the buffer `isUsingInlineImages()`. ARCHITECT: *"surface layers never acknowledged any of rendering context, must be rendering engine agnostic … what's the BLESSED compliance approach?"*. The buffer tests its own switch, and the getter is deleted.
- Windows `read` returned 0 on a failed `ReadConsoleInputW`, and paired a high surrogate with any next record. Fixed (findings 6, 2).
- `TerminalUnixDisplay::open` was 52 lines; the readiness sources moved to `addSources` (finding 5).
- A restore with no saved mode would write a zero mode to the terminal. The saved modes are `std::optional`.
- COUNSELOR failure: findings 1 and 7 (races at close) had no runtime evidence. COUNSELOR accepted them, edited the tested `TerminalUnixTTY`, and proposed changes to `jam::AsyncOutputStream` and `jam::VirtualDevice`. ARCHITECT: *"why should we reinvent new pattern? why TIT must be different? what's wrong with EVE?"*. The TTY is restored to HEAD; the suspend logic dates from JAM `9e2c77a81` (2026-10-26). Both findings are rejected.
- The plan had the display backends re-arm with no `TerminalClock`, and the Windows callback re-armed itself. EVE's contract (`eve/SPEC.md:45-50`): the clock waits one period after each tick, and "the Windows backend's next overlapped read is the re-arm". The backends and Steps 5 to 7 now follow it.

### State for Continuation
- Plan Step 9 is not run: no consumer is built. The help baselines are in the session scratchpad `help-before/` (`ling.txt` 49508 bytes, `cast.txt` 209867, `scout.txt` 79958, `titc.txt` 5336; each ends `ESC[0m` LF). The scratchpad can be cleared; capture again from the installed binaries before the build if it is gone. Build `ling`, `cast`, `scout`, `tit` with `cast cast/spell.md --debug` and compare `--help` byte for byte. Build `ling` on Windows.
- `TerminalDisplay` has no caller. Part 7 calls `open (context, bufferSize, maximumPeriodMilliseconds)`.
- The doxygen pass is due. Stale blocks: `jam_TerminalBuffer.h` (constructor, VT `writeToStream` "holds no image placement (jassert)", `writeFrame` `@tparam RowWriter`), `jam_TerminalEncoder.h` (`writeRow`: "and a newline", no `@param placements`). New members and the display files have no doxygen.
- `TerminalWindowsTTY` was rewritten outside this sprint during this session: it reads through `jam::AsyncOutputStream reader` (`jam_TerminalTTY_windows.h:319`). Its `.cpp` file note (`:56`) still names `readWait`. This sprint did not touch it.
- The record shape of VT input under ConPTY is not measured (tracker, "Facts that execution measures first"). Part 7 measures it.

### Debts Paid
- None (no DEBT.md at project root)

### Debts Deferred
- None

---

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
