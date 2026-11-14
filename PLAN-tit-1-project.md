# PLAN: TIT Part 1 — The Console Project

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-13
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part removes the stale C++ tree and makes the `titc` console project from the cli archetype.
At its end, `titc --version` and `titc --help` work, and the binary installs to `~/.local/bin`.
The part adds no state model, no terminal window, and no view. Later parts add them.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Part 0 only.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE: one object is one header and one `.cpp`.
- CODING.md: no anonymous namespace. A file-local function has `static` linkage, as `ling/Source/Main.cpp:19-262` has.
- CODING.md: no `==` on strings. A flag test is a lookup in a `jam::Function::Map` keyed by `juce::Identifier`, as `getFlags()` is (`ling/Source/Main.cpp:104-119,251`). The double-dash test is `compare (...) == 0` (`:196`).
- CODING.md: no Windows SDK macro name as an identifier.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part. The doxygen pass is separate.
- CAROL.md, Refactor-Rewrite Discipline: "Delete first, implement after". Step 1 deletes.
- ARCHETYPE.md §1: every generated file is a build artifact. No step edits `CMakeLists.txt`, `Source/generated/*`, `entitlements.plist`, `gh.cmake`, or `RELEASE.md` by hand (`archetype/ARCHETYPE.md:50-51`).
- ARCHETYPE.md §1: every cast table needs a `+---` separator line between records (`archetype/ARCHETYPE.md:41-43`).
- CAROL.md, Build: agents build only on ARCHITECT's explicit request.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/`.

### The stale tree that Step 1 removes

| File or folder | Why | Where |
|---|---|---|
| `tit/CMakeLists.txt` | It calls `jam/cmake`, `BuildSetup`, and `configure_app`, which do not exist. cast generates this file. | `tit/CMakeLists.txt:12-14,36`; `archetype/ARCHETYPE.md:202-220` |
| `tit/build.sh`, `tit/install.sh` | Hand-written cmake calls into `Builds/Ninja`. The cast toolchain rows replace them. | `tit/build.sh`; `ling/project-info.md:195-219` |
| `tit/Source/` (32 files) | Every class derives from, or uses, `jam::tui::Component` and its siblings, which do not exist: the module `jam_tui` holds only `jam::TerminalPage` (`jam/jam_tui/jam_tui.h:41-43`). `TitState` and `TitIdentifier.h` are hand-rolled (ruling 11). | `tit/Source/Main.cpp:40-107`; `tit/Source/MainComponent.h:8-20`; `tit/Source/state/TitState.h:28` |
| `tit/tests/` (13 files) | Eight use `jam::tui` classes that do not exist. Two include `view/*.h`, a folder that does not exist. No build target names them. | `tit/tests/MenuTests.cpp:1`; `tit/tests/RootViewsTests.cpp:2-5` |
| `tit/themes/default.xml` | The theme format is markdown tables (tracker, Closed by a read). | `tit/themes/default.xml:1-12` |

Not removed: `tit/Assets/tit_logo.svg` (the logo of the banner, Part 15), `tit/___legacy___/`, `tit/SPEC.md`, `tit/carol/`, `tit/.claude/`, `tit/.gitignore`, `tit/.clangd`, `tit/.nvim-standalone-config`, the `PLAN-*.md` files, `tit/RFC.md`. `tit/.clangd` points at `Builds/Ninja` (`tit/.clangd:2`); it stays, because the clangd file is MACHINIST work.

### The generator

| Need | Existing piece | Where |
|---|---|---|
| Make the skeleton | `cast INIT.md <target>`, run from the archetype directory. "Always give a target directory." | `archetype/ARCHETYPE.md:92-95` |
| What the run writes | `project-info.md`, `cast/spell.md`, `cast/signing.md`, `RELEASE.md` | `archetype/cli/INIT.md:43-55`; `archetype/ARCHETYPE.md:96-98` |
| The framework cells | JAM values; no edit is needed | `archetype/cli/INIT.md:21-41` |
| The placeholder list | `MyApp`, `myapp`, `MyCompany`, `My Company Ltd`, `com.mycompany.*`, `0.1.0` / `0x100`, `Your Name`, `TEAMID0000`, `info@example.com`, `https://example.com`, the zero thumbprint | `archetype/ARCHETYPE.md:108-116` |
| The project holds its own templates | The generated `cast/spell.md` names `cmake.cast` and `installer.cast` in the project's `cast/`; the installer artwork is `cast/installer/resources/`. Nothing points into the archetype. ARCHITECT: *"all MUST BE ENCAPSULATED"*; *"archetype is not the source, it's the cookie cutter for init"* | `archetype/cli/INIT.md:34,36`; `archetype/cli/init.cast:99`; `ling/cast/spell.md:9-10`; `ling/project-info.md:98` |
| The templates of the project | `archetype/cli/cast/cmake.cast` is `cast/cast/cmake.cast` verbatim; `archetype/cli/cast/installer.cast` is `cast/cast/installer.cast` without the three MUI bitmap lines (ARCHITECT: *"delete"*) | `archetype/cli/cast/` |
| A `LICENSE` file at the root | "the project root must hold a `LICENSE` file" | `archetype/ARCHETYPE.md:142-143` |
| The build flows | `cast cast/spell.md --debug` makes a Debug build in `Builds/Debug` | `archetype/ARCHETYPE.md:159-163`; `ling/project-info.md:208-212` |
| The install | "Post-build always: copy the binary to `installDirectory`" | `archetype/ARCHETYPE.md:269`; `ling/project-info.md:64` |
| A converged project | "a second `cast` run produces an empty diff" | `archetype/ARCHETYPE.md:478-479` |

### The nearest sibling: LING

LING is the newest cli project of the same author, and it is the only one with a `juce::JUCEApplicationBase` entry. scout, cast, and ggwp have a plain `main` (`scout/Source/Main.cpp:263`; `cast/Source/main.cpp:1622`; `ggwp/Source/Main.cpp:37`). TIT needs the message loop from Part 5 on, so its entry takes the LING form now.

| Need | Existing piece | Where |
|---|---|---|
| Project values of the same author | `companyName`, `legalCompanyName`, `productWebsite`, `companyEmail` | `ling/project-info.md:27-32` |
| The `## cmake` values of the same author | `companyCopyright`, `qaDirectory`, the `jrengmusic/<name>` repository, the `com.jrengmusic.<name>` bundle identifier | `ling/project-info.md:66,70,94,96` |
| The signing values of the same author | six rows | `ling/cast/signing.md:6-16` |
| The JUCE module rows | six rows | `scout/project-info.md:108-113` |
| The user module rows of a help printer | fourteen rows, `jam_core` to `jam_tui` | `scout/project-info.md:121-134` |
| The macro rows | thirteen rows | `ling/project-info.md:155-167` |
| The include rows | four rows | `ling/project-info.md:175-178` |
| The binary rows | `help`, `styleSheet` | `ling/project-info.md:186-187` |
| The project word tables | `cast/identifiers.md`, `cast/text.md`, `cast/files.md`, `cast/banner.md`, `cast/source.cast` | `ling/cast/` |
| The manifest wiring for those tables | index rows, header rows, output groups, the `## output index` table | `ling/cast/spell.md:21-32,45-69,140-191,193-222` |
| The entry file | `printError`, `writeVersion`, `isTerminalOutput`, `getFlags`, `getFlagWord`, `runCommandLine`, `class Application`, `START_JUCE_APPLICATION` | `ling/Source/Main.cpp:30-39,49-60,67-74,104-119,192-200,246-264,278-486,488` |
| The help printer | `jam::TerminalPage`: `TerminalPage (ResourceFetchers, const jam::Strings& styleSheets)`, `setDocument`, `writeToStream (juce::OutputStream&)`. ARCHITECT: *"jam::TerminalPage"* | `jam/jam_tui/page/jam_TerminalPage.h:11-53`; LING use `ling/Source/Main.cpp:83,86-87,91-93` |
| The help style sheet | the JAM file `jam/jam_tui/page/style.css`, embedded by the `## binary` row `styleSheet`; its name constant `files::styleSheet` is JAM's | `ling/project-info.md:187`; `jam/cast/style.md:994` |
| The help file name constant | `files::lingHelp`; the family is `<project>Help` | `ling/cast/files.md:31` |
| The license | `tit/___legacy___/LICENSE` is byte-identical to `ling/LICENSE` (compared with `cmp`) | `tit/___legacy___/LICENSE` |

### Words the framework already holds

| Word | Where |
|---|---|
| `Id::version`, `Id::diagnosticSeparator` | used at `ling/Source/Main.cpp:34,108`; rows `jam/cast/core.md:66,161` |
| `Chars::doubleDash`, `Chars::space`, `Chars::openParen`, `Chars::closeParen`, `Chars::newline` | used at `ling/Source/Main.cpp:38,53,56,58,192` |
| `jam::TerminalEncoder::writeEraseScreen` | used at `ling/Source/Main.cpp:331` |
| `jam::Format::appendUtf8`, `jam::Strings::getView`, `jam::Strings::fromView`, `jam::Function::Map` | used at `ling/Source/Main.cpp:33,37,104-119,249` |
| `Id::help` is not in JAM. Each cli project declares it in its own table. | `ling/cast/identifiers.md:38` |

### The project values of TIT

| Value | Source |
|---|---|
| Binary and product name `titc` | tracker, Closed by a read; `tit/CMakeLists.txt:37-38` |
| Bundle identifier `com.jrengmusic.titc` | `tit/CMakeLists.txt:42` |
| Version `0.0.0` | `tit/CMakeLists.txt:27,39`; `tit/Source/Main.cpp:17`. *"agents never change a version."* (`eve/SPEC.md:913`) |
| Long name "Terminal Interface for git" | `~/.carol/ARCHITECT.md`, Projects, TIT |
| Repository `jrengmusic/tit` | the remote of the working copy: `github.com:jrengmusic/tit` (commit `79e5999`, merge message) |

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: Remove the stale tree
**Scope:** `tit/CMakeLists.txt`, `tit/build.sh`, `tit/install.sh`, `tit/Source/` (whole folder), `tit/tests/` (whole folder), `tit/themes/` (whole folder), `tit/Builds/Ninja/` (whole folder).
**Action:** @Engineer deletes the three files and the four folders. `tit/Builds/Ninja/` is the build tree of the removed `jam_tui` project (ruling 27). @Engineer runs no version-control command; the files leave the working tree only.
**Validation:**
- The seven paths do not exist.
- `tit/Assets/tit_logo.svg`, `tit/___legacy___/`, `tit/SPEC.md`, and `tit/carol/` are unchanged.
- No source file under `tit/` names `jam::tui`.

### Step 2: Generate the skeleton
**Scope:** `tit/project-info.md`, `tit/cast/spell.md`, `tit/cast/signing.md`, `tit/RELEASE.md` (all new).
**Action:**
1. Go to `~/Documents/Poems/dev/archetype/cli`.
2. Run `cast INIT.md ~/Documents/Poems/dev/tit`.
3. Change no cell of `INIT.md`. The framework is JAM.
4. Put the project's own templates and artwork in `tit/cast/`: `cmake.cast` and `installer.cast` from `archetype/cli/cast/`, and `installer/resources/background.png` and `background-dark.png` from `archetype/cli/cast/installer/resources/` as placeholders (ARCHITECT: *"just use any generated png as placeholder"*).

**Validation:**
- The four generated files exist under `tit/`, and `tit/cast/cmake.cast`, `tit/cast/installer.cast`, and the two PNG files exist.
- No file under `tit/` outside the `PLAN-*.md` files names `archetype`.
- The archetype's own stubs are unchanged: `archetype/cli/project-info.md`, `cast/spell.md`, `cast/signing.md`, and `RELEASE.md`.
- `tit/SPEC.md`, `tit/CLAUDE.md`, and `tit/carol/` are unchanged.

### Step 3: Set the project values
**Scope:** `tit/project-info.md`, tables `## project info` and `## cmake`.
**Action:** Replace each placeholder. Change no other cell.

`## project info`:

| name | value | description cell |
|---|---|---|
| `projectName` | `titc` | `Terminal Interface for git` |
| `companyName` | `JRENG` | unchanged |
| `legalCompanyName` | `Jubilant Research of Eclectic Novelty Generation` | unchanged |
| `versionString` | `0.0.0` | unchanged |
| `versionNumber` | `0x0` | unchanged |
| `productWebsite` | `` `https://jrengmusic.com` `` | unchanged |
| `companyEmail` | `info@jrengmusic.com` | unchanged |
| `presetExtension` | `titc` | unchanged |
| `presetDefault` | `INIT` | unchanged |

`## cmake`:

| key | value |
|---|---|
| `projectDescription` | `Terminal Interface for git` |
| `banner` | `titc` |
| `bundleIdentifier` | `com.jrengmusic.titc` |
| `targetName` | `titc` |
| `companyCopyright` | `© 2026 JRENG` |
| `productName` | `titc` |
| `buildVersion` | `0.0.0` |
| `documentExtensions` | `titc` |
| `qaDirectory` | `` `$ENV{HOME}/Documents/Poems/dev/___builds___` `` |
| `repository` | `jrengmusic/tit` |

**Validation:**
- No placeholder of `archetype/ARCHETYPE.md:108-116` remains in `project-info.md`.
- Each value matches the LING row of the same key (`ling/project-info.md:26-34,42-100`), with the TIT value of the table "The project values of TIT" where that table gives one. The template reads `:::projectDescription:::` (`cast/cast/cmake.cast:36`).
- `juceTargetFunction` is `juce_add_console_app`, as generated.

### Step 4: Set the module, macro, include, and binary rows
**Scope:** `tit/project-info.md`, tables `## juce module`, `## user module`, `## macro`, `## include`, `## binary`.
**Action:**
1. `## juce module`: write the six rows of `scout/project-info.md:108-113`.
2. `## user module`: write the fourteen rows of `scout/project-info.md:121-134`, in that order, with those description cells: `jam_core`, `jam_document`, `jam_markdown`, `jam_freetype`, `jam_graphics`, `jam_data_structures`, `jam_animation`, `jam_web`, `jam_style`, `jam_gui`, `jam_mermaid_diagram`, `jam_markdown_graphics`, `jam_terminal`, `jam_tui`.
3. `## macro`: write the thirteen rows of `ling/project-info.md:155-167`.
4. `## include`: write the four rows of `ling/project-info.md:175-178`.
5. `## binary`: write the two rows of `ling/project-info.md:186-187` (`help`, `styleSheet`).
6. Change no row of `## toolchain`, `## source glob`, `## library`, `## pack`, `## release notes`, `## release`, or `## debug`.

**Validation:**
- Each row is identical to its cited row.
- The user module list holds no `jam_subprocess`, no `jam_pdf`, no `jam_syntax_tree`, no `jam_fuzzy`, and no `jam_profiler`. No code of this part uses them (MANIFESTO L, YAGNI). Part 8 adds `jam_subprocess`.
- Each table has a `+---` separator line between records where the generated table has one.

### Step 5: Set the signing values
**Scope:** `tit/cast/signing.md`.
**Action:** Write the six value cells of `ling/cast/signing.md:6-16` into the six rows.
**Validation:** The six value cells equal the LING cells of the same keys. LING's entitlement row `com.apple.security.cs.disable-library-validation` (`ling/cast/signing.md:12`) is not taken; no code of this part loads a library. No placeholder (`Your Name`, `TEAMID0000`, the zero thumbprint) remains.

### Step 6: Add the license file
**Scope:** `tit/LICENSE` (new).
**Action:** Copy `tit/___legacy___/LICENSE` to `tit/LICENSE` with no change.
**Validation:** `tit/LICENSE` is byte-identical to `tit/___legacy___/LICENSE` and to `ling/LICENSE`.

### Step 7: Add the project word tables
**Scope:** `tit/cast/identifiers.md`, `tit/cast/text.md`, `tit/cast/files.md`, `tit/cast/source.cast` (all new).
**Action:**
1. `cast/identifiers.md`: copy the `## index` table of `ling/cast/identifiers.md:1-7`. Write the `## identifiers` table with the brief `@brief TIT vocabulary — the words of the titc command line.` and one row: type `@id`, name `help`, value `` `help` ``, empty description.
2. `cast/text.md`: write the `## diagnostics` table with the brief `@brief titc message texts, emitted verbatim when a run fails.` and one row: name `failFlagUnknown`, value `` `unknown flag` ``, description `An argument that is not --help or --version.`
3. `cast/files.md`: copy the `## index` table of `ling/cast/files.md:1-7`. Write the `## files` table with the brief `@brief File names titc reads: help.` and one row: type `@string`, name `titcHelp`, value `` `HELP.md` ``, description `Rendered help text.`
4. No `cast/banner.md` (ruling 23). The banner of TIT is the logo of `Assets/tit_logo.svg`, shown by the components of Part 15.
5. `cast/source.cast`: copy `ling/cast/source.cast` (6 lines) with no change.

**Validation:**
- `help` follows the LING row `help`. `failFlagUnknown` follows `ling/cast/text.md` row `failFlagUnknown`. `titcHelp` follows `lingHelp` (`ling/cast/files.md:31`).
- No table holds a row that no code of this part reads (MANIFESTO L, YAGNI).
- No banner table exists in `tit/cast/`.

### Step 8: Wire the tables in the manifest
**Scope:** `tit/cast/spell.md`.
**Action:**
1. `## index`: add the rows `@source`, `@identifiers`, `@Identifiers`, `@text`, `@binaryFiles`, `@Text`, `@Files`, `@GeneratedSource`, `@Generated` of `ling/cast/spell.md:21-31`, after the `@Release` row. Do not add `@banner` (`:25`) or `@HashMaps` (`:29`), or `@string` (`:32`), which only those groups read.
2. `## headers`: add the four rows `Identifiers.h`, `Text.h`, `Files.h`, `Generated.h` of `ling/cast/spell.md:45-59,65-69`. Do not add `HashMaps.h` (`:60-64`). In each brief and description, write `titc` where LING writes `ling`, and `TIT` where LING writes `LING`.
3. `## output`: add the row groups of `ling/cast/spell.md:140-181` at the end of the table, without the `@banner:banner` group (`:158-166`). Do not add the `@HashMaps` group (`:182-191`).
4. Add the `## output index` table of `ling/cast/spell.md:193-217`, without its `HashMaps.h` group (`:218-222`).

**Validation:**
- The added rows differ from the cited LING rows only in the project words.
- The arity law holds (`archetype/ARCHETYPE.md:251-255`): no source line of the `@cmake:cmake` group is added or removed.

### Step 9: Add the help file
**Scope:** `tit/Source/HELP.md` (new).
**Action:** `jam::TerminalPage` prints the help (ARCHITECT: *"jam::TerminalPage"*; `jam/jam_tui/page/jam_TerminalPage.h:11-53`), with the JAM style sheet of the `styleSheet` binary row (Step 4). No `Help.h`, no `Help.cpp`, and no `style.css` exist in TIT. Write `Source/HELP.md` with this text.

```markdown
# TIT: Terminal Interface for git

**Version 0.0.0**

You see this guide because you ran `titc` with no argument or `titc --help`.

## Command Line

| Command          | Result                                          | Exit |
| :--------------- | :---------------------------------------------- | :--- |
| `titc`           | Prints this guide.                              | 0    |
| `titc --help`    | Prints this guide.                              | 0    |
| `titc --version` | Prints `titc 0.0.0 (<commit>)`.                 | 0    |
| `titc <other>`   | Prints `titc: <other>: unknown flag` on stderr. | 1    |

On a terminal, titc clears the screen first.
```

**Validation:**
- `HELP.md` states only the behavior that Step 10 builds.

### Step 10: Add the entry file
**Scope:** `tit/Source/Main.cpp` (new).
**Action:** Write the file from `ling/Source/Main.cpp`. Write no comment and no doxygen block.

1. Includes: lines 1-2 and 8-13 of the LING file. Do not take `LINGModel.h`, `Processor.h`, `Server.h`, `Client.h`, or `Manifest.h` (`:3-7`).
2. Take these functions with their LING bodies, unchanged: `printError` (`:30-39`), `writeVersion` (`:49-60`), `isTerminalOutput` (`:67-74`), `getFlagWord` (`:192-200`).
3. Constants: take `flagArgIndex { 0 }` and `flagOnlyArgumentCount { 1 }` (`:41,43`). Do not take `postFlagArgIndex` or `toolchainArgumentCount` (`:42,44`).
4. `writeHelp()`: a static function in the family of `writeVersion()` (`:49-60`). Its body is LING `printBannerAndHelp` (`:83,86-87,91-93`) with `files::titcHelp` in place of `files::lingHelp`: the `BinaryData::Raw` of the help, a `jam::TerminalPage` with `{ BinaryData::fetcher }` and `jam::Strings { files::styleSheet }`, a `juce::MemoryOutputStream`, `setDocument` with the parsed `jam::MarkdownDocument`, `writeToStream`, and `fwrite` to stdout. It prints no banner and no blank line before the help (ruling 23), so it does not take `:84-85,89-90`.
5. `getFlags()`: take LING `:104-119`. Its type is `jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, int ()>`, and each entry takes `const juce::Identifier&` (`:104,109,111-112`). The `Id::version` row calls `writeVersion()`. The `Id::help` row calls `writeHelp()`. Both rows return `EXIT_SUCCESS`.
6. `runCommandLine (const juce::StringArray& arguments)`: write these steps, in this order. Each return is a result return.
   - Take LING `:248-249`: the flag word of a single argument, through `getFlagWord`, and its `juce::Identifier`.
   - `getFlags().contains (flag)` gives `getFlags().get (flag)` (LING `:251`).
   - `arguments.size() >= flagOnlyArgumentCount` gives `printError (jam::Strings::getView (arguments[flagArgIndex]), text::Diagnostics::failFlagUnknown)` and `EXIT_FAILURE` (the form of LING `:256-260`).
   - Otherwise `writeHelp()` and `EXIT_SUCCESS` (the form of LING `:262-263`).
7. `class Application : public juce::JUCEApplicationBase`: take these members with their LING bodies: `getApplicationName` (`:300`), `getApplicationVersion` (`:306`), `moreThanOneInstanceAllowed` (`:312`), `shutdown` (`:364`), `anotherInstanceStarted` (`:366`), `systemRequestedQuit` (`:368`), `suspended` (`:370`), `resumed` (`:372`), `unhandledException` (`:375-378`), and `JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Application)` (`:485`).
8. `Application::initialise`: take LING `:324-336` (the Windows code page and the erase on a terminal). Then write `setApplicationReturnValue (runCommandLine (getCommandLineParameterArray()));` and `quit();` (LING `:358-359`).
9. Do not take the base `juce::ValueTree::Listener`, the constructor, the destructor, `runDaemon`, `runOperation`, `valueTreePropertyChanged`, or the members `model`, `processor`, `manifest`, `message`, `lock`, `server`, and `client`. Part 7 adds the model.
10. Do not take `printBanner`, `printBannerAndHelp`, `appendDocument`, `runDocumentNotFound`, `getDocumentFile`, `getToolchainArgument`, or `isDocument`. No code of this part calls them.
11. End the file with `START_JUCE_APPLICATION (Application)` (`:488`).

**Validation:**
- Every function name is in `ling/Source/Main.cpp`, except `writeHelp`, which joins the family of `writeVersion` (NAMES.md Rule 5).
- No `==` compares a string. No anonymous namespace exists. Each file-local function is `static`.
- `runCommandLine` has no bail-out guard. Each return carries the exit code (MANIFESTO E, result return).
- The file holds no function that nothing calls (MANIFESTO L, YAGNI).

### Step 11: Build and check on macOS
**Scope:** the `tit/` project. Generated outputs: `CMakeLists.txt`, `Source/generated/*`, `entitlements.plist`, `gh.cmake`, `cast/installer/*`.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:**
1. From `tit/`, run `cast cast/spell.md --debug`.
2. Run `titc --version`, `titc --help`, `titc`, and `titc --wrong` from `~/.local/bin`.
3. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- The first run ends with exit code 0. `Builds/Debug` holds the binary. `~/.local/bin/titc` exists.
- `titc --version` prints `titc 0.0.0 (<commit>)` and exits 0.
- `titc --help` and `titc` print the guide, and exit 0.
- `titc --wrong` prints `titc: --wrong: unknown flag` on stderr and exits 1.
- The second run changes no generated file (`archetype/ARCHETYPE.md:478-479`).
- `Source/generated/` holds `ProjectInfo.h`, `Identifiers.h`, `Text.h`, `Files.h`, `Generated.h`, `Generated.cpp`, as `ling/Source/generated/` does, without `HashMaps.h`.

### Step 12: Build and check on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 11 on a Windows machine, with the binary at the install directory that the Debug build copies to.
**Validation:** The four commands of Step 11 give the same output and the same exit codes.

## BLESSED Alignment

- **B — Bound.** The part adds no owned resource. `Application` has no member.
- **L — Lean.** The part takes only the functions, constants, table rows, and module rows that its own code uses.
- **E — Explicit.** Each value is a table row. A wrong argument fails with a message and exit code 1.
- **S — SSOT.** `project-info.md` holds each project value once. The project holds its own templates in `cast/`, and nothing points into the archetype.
- **S — Stateless.** The process holds no state. It prints and quits.
- **E — Encapsulation.** Data flows one way: tables, `cast`, generated headers, code.
- **D — Deterministic.** A second `cast` run gives no change. The same argument gives the same output and exit code.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None. Ruling 23 replaces ruling 18: Part 1 has no banner table.
