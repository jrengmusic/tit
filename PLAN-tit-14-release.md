# PLAN: TIT Part 14 — Release, ARCHITECTURE.md, README.md

**RFC:** none — objective from ARCHITECT prompt (`/goplan 14`) and the tracker row 14. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE). This part writes no C++.

## Overview

Part 1 generates the signing and installer lane of the cli archetype. This part runs it.
The part does four things:

1. It writes the release notes rows of `project-info.md`.
2. It writes `ARCHITECTURE.md` from the built code.
3. It writes `README.md` with the readme-writer skill.
4. ARCHITECT runs the signed release on macOS, on Windows x64, and on Windows arm64.

The part writes no source file. It runs after Parts 1 to 13 and Part 15, because each document records the code that exists (`~/.carol/ARCHITECTURE-WRITER.md:56-65`, "Rule 0: Codebase is SSOT").

Tracker and canon: `PLAN-tit-0-tracker.md`.

## Language / Framework Constraints

- ARCHETYPE.md §1: every generated file is a build artifact. No step edits `CMakeLists.txt`, `entitlements.plist`, `gh.cmake`, `cast/installer/*`, or `RELEASE.md` by hand (`archetype/ARCHETYPE.md:50-51`). `RELEASE.md` comes from the `## release notes` rows (`archetype/cli/init.cast:375-377`; `archetype/cli/cast/installer.cast:100-126`).
- ARCHETYPE.md §1: every cast table needs a `+---` separator line between records (`archetype/ARCHETYPE.md:41-43`).
- CAROL.md, Build: agents build only on ARCHITECT's explicit request. Each release run of this part is ARCHITECT's.
- CAROL.md, Git: no agent runs a repository command for this part. `gh` runs inside the toolchain of ARCHITECT's release run (`archetype/cli/cast/cmake.cast:323-342`).
- `eve/SPEC.md:913`: *"agents never change a version."* Only ARCHITECT sets the version cells.
- CAROL.md, Writing Standard: both documents follow ASD-STE100. The README voice rule of the readme-writer skill governs its `Why` and `Architecture` sections (`~/.carol/skills/readme-writer/SKILL.md:22-27`).
- `~/.carol/skills/readme-writer/SKILL.md:61-66`: the README holds no sprint number, debt id, or agent name, and no file, class, or function name in `Why`, `Architecture`, or `What`.
- `ARCHITECTURE.md` names no PLAN, RFC, ruling number, sprint, or agent: it records what exists in the code (`~/.carol/ARCHITECTURE-WRITER.md:56-65`). CODING.md, COMMENTS, states the same rule for a comment.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise.

### The lane that Part 1 generates

| Fact | Where |
|---|---|
| One release lane for the cli, plugin, and app sets. macOS: pkgbuild, productbuild, productsign, notarize, staple. Windows: NSIS and signtool. No dmg, no `cast --pack`, no `## pack layout`. | `archetype/ARCHETYPE.md:66-68,263-269` |
| The three flows: `cast cast/spell.md` (Release, signed), `--debug` (Debug), `--no-sign` (Release, not signed) | `archetype/ARCHETYPE.md:159-163`; `archetype/cli/init.cast:159-183` |
| The default flow ends with `cmake -P gh.cmake` | `archetype/cli/init.cast:170` |
| macOS post-build order: strip, xattr, codesign, verify, pkg staging, pkgbuild, productbuild, productsign, notarize, staple, QA directory, QA copy | `archetype/cli/cast/cmake.cast:161-183` |
| codesign with the hardened runtime and the generated entitlements | `cmake.cast:240-242` |
| The pkg installs `/opt/titc/titc`. Its `postinstall` script links `~/.local/bin/titc` to it, for the user at the console. | `cmake.cast:165,252-261,263-270` |
| notarize: `xcrun notarytool submit … --keychain-profile ":::notaryProfile:::" --wait`; staple: `xcrun stapler staple` | `cmake.cast:276-282` |
| The QA copy goes to `<qaDirectory>/<date> titc v<version> <platform>` | `cmake.cast:162-164,284-290` |
| Windows: signtool from the newest Windows Kit; makensis with the binary, the resources, and `LICENSE`; the NSIS script signs the installer and the uninstaller | `cmake.cast:184-195,318-321`; `archetype/cli/cast/installer.cast:34,44-46` |
| Every build copies the binary to `installDirectory`, signed or not | `cmake.cast:198-203,292-302`; `archetype/ARCHETYPE.md:270` |
| The archive name: `<folder>/<productName> v<version> <platform>.<format>`. With the pack rows of Part 1 it is `Release/titc v<version> macOS.pkg` and `Release/titc v<version> Windows.exe`, beside `tit/`. | `cmake.cast:149`; `archetype/cli/init.cast:185-193` |
| `gh.cmake`: the tag is `v` + `versionString`. When the release exists, it uploads the archive with `--clobber`. Otherwise it creates the release with `RELEASE.md` as its notes. No archive stops the run. | `cmake.cast:323-342` |
| `RELEASE.md`: the title, the description, `## What's New` from the `note` rows, the platform table, and the install lines | `archetype/cli/cast/installer.cast:100-126` |
| The signing rows of Part 1 equal LING's: `identity`, `installerIdentity`, `entitlementsPath`, `notaryProfile`, `identityWindows`, `signArgumentsWindows` | `PLAN-tit-1-project.md` Step 5; `ling/cast/signing.md:6-16` |

### The version and the repository

| Fact | Where |
|---|---|
| The version is in the cells `versionString`, `versionNumber`, and `buildVersion` of `project-info.md`, and in the line `**Version 0.0.0**` of `Source/HELP.md`. LING carries the same line in its guide. | `PLAN-tit-1-project.md` Steps 3 and 9; `ling/Source/HELP.md:3`; `ling/project-info.md:29-30,74` |
| The repository `jrengmusic/tit` holds the tags `v0.0.1`, `v0.0.2`, `v0.0.3`, and `v0.0.4`, from the Go releases | `git log --tags --simplify-by-decoration` on the working copy, session 2026-11-14 |
| A version equal to one of these tags uploads the archive into that Go release (`--clobber`) | `cmake.cast:335-338` |
| The Go releases come from `goreleaser`, with a bare binary per platform, a macOS codesign and notarize, no staple, no pkg, and no Windows signature | `tit/___legacy___/.goreleaser.yaml:10-44`; `tit/___legacy___/scripts/post-build.sh:7-24` |

### The machines

| Platform | Machine | Where |
|---|---|---|
| macOS, universal (`x86_64` and `arm64`) | either Mac | `archetype/cli/init.cast:98-104`; `archetype/cli/cast/installer.cast:113` |
| Windows x64 | iMac 5K 2015, Windows 10, UCRT64 | `~/.config/WINDOWS-SETUP.md:9-14,20-23` |
| Windows arm64 | MBP M4, Windows 11 ARM64 in UTM, CLANGARM64 | `~/.config/WINDOWS-SETUP.md:9-14,20-23` |

### The document rules

| Rule | Where |
|---|---|
| README sections, in order: `# NAME` with the bold expansion and one paragraph; `## Why`; `## Architecture` with a link to ARCHITECTURE.md; `## What`, one `###` for each feature area; `## How` with `**Requirements:**`, a `bash` block, the config, and the platform notes; a status table only for open features; `## License`; the footer. A `---` line separates the major sections. | `~/.carol/skills/readme-writer/SKILL.md:39-59` |
| README sources: `project-info.md`, `CMakeLists.txt`, `Source/`, SPEC.md, ARCHITECTURE.md. A quoted command runs, or the script that holds it is cited. | `readme-writer/SKILL.md:29-37` |
| README voice: first person, manifesto voice, in `Why` and `Architecture` only. The words come from ARCHITECT or from SPEC.md and ARCHITECTURE.md. A claim of speed or size needs a measurement. | `readme-writer/SKILL.md:22-27` |
| The tone reference is the END README | `end/README.md:1-131` |
| ARCHITECTURE.md records what exists in the code | `~/.carol/ARCHITECTURE-WRITER.md:56-65,618-621` |
| COUNSELOR writes ARCHITECTURE.md | `~/.carol/ARCHITECTURE-WRITER.md:30-37,299-306` |
| The template sections and the adaptation for a small project | `~/.carol/templates/ARCHITECTURE.md:1-316`; `~/.carol/ARCHITECTURE-WRITER.md:589-596` |
| The completeness, accuracy, clarity, and consistency checklist | `~/.carol/ARCHITECTURE-WRITER.md:426-451` |
| The siblings of the same author: `whatdbg/ARCHITECTURE.md` (553 lines) and `end/ARCHITECTURE.md` (969 lines) | Pathfinder inventory, session 2026-11-14 |
| The license of TIT is MIT | `tit/___legacy___/LICENSE:1`; Part 1 Step 6 copies it to `tit/LICENSE` |
| The release notes form of the same lane: the first row names the lane, then one row for each feature, `**Name** -- text` | `whatdbg/project-info.md:203-231`; `whatdbg/RELEASE.md:5-18` |

## Closed by a read

1. **No row of the type `entitlement`.** The signing table of the cli template names the entitlements file (`entitlementsPath`, `archetype/cli/init.cast:390`) and holds no row of the type `entitlement` (`:355`, `:380-398`). So the generated file has an empty `dict` (`ling/entitlements.plist:19-20`). LING has none (`ling/cast/signing.md:6-16`), and LING runs child processes with `jam::Subprocess` (`ling/Source/CompilerManifest.cpp:152-159`). The hardened runtime needs no entitlement to start a child. The Go entitlements (`tit/___legacy___/entitlements.plist:5-8`) belong to the Go binary, which this part does not build. Part 1 Step 5 makes `tit/cast/signing.md` equal to LING's. Step 5 of this part measures it: the signed `titc` runs the repository tool.
2. **This part runs after Part 15.** Each document records the code that exists (ARCHITECTURE-WRITER Rule 0; readme-writer `Source`). Part 15 adds the banner. The tracker row 14 depends on Parts 1 to 13 and 15.
3. **The version is ARCHITECT's.** Step 5 starts with ARCHITECT's version. This part writes no version cell. The tags of the Go releases are facts for that choice (inventory, "The version and the repository").
4. **The Go release scripts are not carried.** The archetype lane replaces `release.sh`, `.goreleaser.yaml`, and `scripts/post-build.sh` (tracker, "it needs to conform with other project structured with manifest, build with cast"). `tit/___legacy___/` stays as it is (Part 1, "Not removed").
5. **COUNSELOR writes both documents.** ARCHITECTURE-WRITER gives ARCHITECTURE.md to COUNSELOR (`:30-37`). The readme-writer skill serves any primary (`readme-writer/SKILL.md:3`). No Engineer delegation exists for a document.
6. **The texts of `Why` and `Architecture` are in this plan.** The skill takes them from ARCHITECT's words (`readme-writer/SKILL.md:26`). Step 3 gives the text, built from the quotes of the tracker canon. Each sentence names its source in the table under it. ARCHITECT's lock of the plan approves the text, as rulings 22 and 41 approved the texts of Parts 2 and 12.
7. **The texts of `What` are in this plan.** Each sentence states a feature of a locked part. Step 3 validates each sentence against the built code. A sentence that the code contradicts takes the fact of the code (readme-writer: "Write no claim that a read or a run does not support").
8. **The pack rows stay as generated.** Ruling 49: *"just follow the working pattern"*. Part 14 changes no cell of `## pack` (Part 1 Step 4, item 6).

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: The release notes rows
**Scope:** `tit/project-info.md`, table `## release notes`.
**Action:** Write these eight rows in the `note` column, in this order. Put a `+---` separator line after each row. Widen the table borders to the longest row.

| # | note |
|---|---|
| 1 | `Release lane: a signed macOS pkg and Windows x64 and arm64 installers, uploaded to the GitHub release.` |
| 2 | `First C++ release: JUCE 8 and JAM replace Go, Bubble Tea, and Lip Gloss. The binary is titc. The Go tit and its .toml files keep working.` |
| 3 | `**State** -- the state of the repository decides the menu. An action that git would reject is not in the menu.` |
| 4 | `**Menus and keys** -- one menu for each operation state. Every key is a row of ~/.config/tit/keys.md.` |
| 5 | `**History** -- one browser: the commits, the files of a commit, the diff of a file. Search, function history, incoming and outgoing commits, and working tree changes use the same browser.` |
| 6 | `**Protocols** -- dirty operations, the conflict resolver, time travel, pin, the branch picker, and three rebase actions that never need a force push.` |
| 7 | `**Setup** -- the SSH wizard, init, clone, remotes, preferences, auto-update, themes, and Git LFS.` |
| 8 | `**Banner** -- the TIT logo in titc --help and beside the menu.` |

**Validation:**
- Row 1 is identical to `whatdbg/project-info.md`, the first `note` row: the lane is the same.
- Rows 3 to 8 hold no text that `README.md` `## What` (Step 3) does not support.
- No cell holds a `|`. The table has a separator line between records.
- No other table of `project-info.md` changes.

### Step 2: ARCHITECTURE.md
**Scope:** `tit/ARCHITECTURE.md` (new).
**Owner:** COUNSELOR (`~/.carol/ARCHITECTURE-WRITER.md:30-37`).
**Action:**
1. Delegate a read-only inventory to @Pathfinder, with the doxygen and LSP instruction of CAROL.md Doxygen. It lists, with file:line:
   - every file of `tit/Source/` and `tit/Source/generated/`, and the class or the free functions of each;
   - the members of `TITModel`: each property row and each child tree, with its type;
   - each `jam::` and `juce::` base class and each JAM call of the TIT classes;
   - each thread: the message thread, the pool of `TITGitProcessor`, the serial context of the terminal device, the file watcher;
   - the marker files and the stash names that TIT writes in a repository;
   - the cast tables of `tit/cast/` and the user files of `~/.config/tit/`.
2. Write the document with these sections, in this order. The template is `~/.carol/templates/ARCHITECTURE.md`, adapted for a small project (`ARCHITECTURE-WRITER.md:593-596`). Each fact cites the built file and line.

| Section | What it holds | The parts that build it |
|---|---|---|
| `# TIT - Architecture` | The title, then one paragraph: what TIT is, from SPEC.md Overview. | 2 |
| `## Project Overview` | `### Purpose`, `### Architecture Philosophy` (the Model holds the state, the View composes, the Processor runs the repository tool; nothing above the peer names a terminal type), `### Technology Stack` (C++17; JUCE with the `juceVersion` cell; the JAM modules of `## user module`; cast; macOS and Windows). | 1, 7, 8 |
| `## Module Structure` | `### Module Map`: a tree of the TIT classes and the JAM classes they use. `### Module Inventory`: one row for each file of `tit/Source/`: file, responsibility. | 1 to 13, 15 |
| `## Layer Separation Rules` | The application owns the Model, the window, the command manager, and the style manager. The View composes the views. The Processor writes the Model only on the message thread. The views read the Model. `jam_gui` includes `jam_terminal`, never the reverse. | 5, 7, 8, 9, 10 |
| `## Threading Model` | One row for each thread of step 1: what runs on it, and how its result reaches the message thread. | 3, 7, 8, 13 |
| `## Data Flow` | Four `###` flows: a key (terminal bytes, the device, the key lane, the command manager, the command); a frame (paint, the terminal graphics context, the VT lane, stdout); a detection (the pool job, the message thread, the `git` child, the listeners, the menu); a user file (the watcher, the load, the Model rows, the listeners). | 3 to 13 |
| `## State Management` | The tree of `TITModel`: each row and each child, its type, its writer, and its readers. | 7, 8, 11, 12, 13 |
| `## Commands and Keys` | The command ids of the cast table; the key-sheet modes; availability as rules over the rows of the Model, which an event of the Model reads to build the menu and the key sheet (rulings 55 and 56); the user key file. | 9, 10, 13 |
| `## Repository Markers` | Each marker file and stash name that TIT writes, what it means, and which command removes it. | 12 |
| `## Key Design Decisions` | One `###` for each decision in the list below: the decision, then the reason in plain words. | all |
| `## Known Limitations` | On Windows, progress lines arrive late: a read returns only when the buffer is full or the child ends (`jam_subprocess/subprocess/native/jam_ChildProcess_windows.cpp:55-89`). TIT reads no mouse input. | 8 |
| `## File Structure` | The tree of `tit/` without `Builds/` and `___legacy___/`, one line for each folder. | 1 |
| `## Glossary` | The words that a reader needs: axis, operation, marker, pin, time travel, dirty operation, key-sheet mode, cell. | 2 |
| `## References` | `SPEC.md`, `README.md`, the JAM folder, JUCE. | — |

   The decisions of `## Key Design Decisions`:
   1. The interface is a function of the repository state. An action that the repository tool would reject is inactive.
   2. The Model holds all state. No view and no worker holds a copy.
   3. Every command, callback, and key goes through the command manager. One key sheet maps keys to commands by mode.
   4. Every view is a `juce::Component` painted into the terminal through a peer. TIT runs in any terminal that reads the VT sequences it writes.
   5. Each text view and each text input is the framework's text editor.
   6. History is one browser over one shape of records.
   7. The commit cache is a Model tree that one pool job builds.
   8. A rewrite of commits never needs a force push. The command is absent when it would.
   9. Pin is a hard reset on the branch with a marker. Time travel detaches.
   10. The theme is a style sheet in markdown. The framework's style manager applies it. TIT has no theme manager.
   11. The user files are carried defaults, written to disk when absent or old. A valid disk file wins. An invalid one keeps the default and shows its error.
   12. TIT never reads or changes the `.toml` files of the Go TIT.
3. Write no sprint number, no ruling number, no agent name, and no reference to a PLAN or RFC file (Language / Framework Constraints).
4. Write no `[ARCHITECTURAL DECISION NEEDED]` mark: each decision of the list has a ruling in the tracker canon.

**Validation:**
- The checklist of `~/.carol/ARCHITECTURE-WRITER.md:426-451` holds: each file of `tit/Source/` has a row; each thread has a rule; each path and line is current.
- Each fact has a citation of the built code. No fact comes from a plan alone (Rule 0).
- A difference between the code and a locked plan goes into the sprint summary, and the document records the code (`ARCHITECTURE-WRITER.md:60-65`).
- The text follows ASD-STE100.

### Step 3: README.md
**Scope:** `tit/README.md` (new).
**Owner:** COUNSELOR, with the readme-writer skill.
**Action:** Load `~/.carol/skills/readme-writer/SKILL.md`. Read the sources of its `Source` section. Write the file with this text. Replace `<juceVersion>` with the cell of `project-info.md`.

````markdown
# TIT
**Terminal Interface for git**

I have severe git skills issues. TIT is a git tool for a user who is not a git expert. It always shows what is possible in the actual state of the repository. If it is in the menu, it works.

---

## Why

Git commands fail when you need them most. TIT's interface is a pure function of git state. If git would reject an action, the action is not in the menu. Destructive operations always ask first.

The Go TIT proved the principle. Its presentation did not. The file history was hard to read and hard to understand. Almost every text box was a hand-rolled editor with some effort at vim parity. We had the data. The rendering was limited by our past knowledge, and by the limits of Go and Lip Gloss.

This TIT keeps every feature, and fixes the design. Design by contract.

---

## Architecture

TIT is not a byte-to-byte translation. It conforms to every other project I build: a manifest, built with cast. JUCE is my expertise. A terminal interface can be built exactly like any GUI, on the JUCE native contract: every view is a component, and every look comes from the style sheet. TIT paints those components into the terminal through its own peer. TIT is terminal agnostic, and it renders its pixels itself.

There is one state machine. Every command, callback, and keyboard shortcut goes through one command manager. Every text view and every text input is one text editor, with full features. History is one browser over one source of records: from the commits at the top, down to each file of a commit. Each feed builds the same widget, and the data decides what it shows. Isn't that the whole premise of TIT?

The design is in [ARCHITECTURE.md](ARCHITECTURE.md). The behavior is in [SPEC.md](SPEC.md).

---

## What

### State

TIT reads the repository on five axes before it builds the menu: the environment, the working tree, the timeline against the remote, the operation in progress, and the remote. The menu changes when the repository changes, also from another terminal.

### Menus and keys

Each operation state has its own menu: no repository, normal, merging, rebasing, conflicted, dirty operation, time travel, and pin. One key can mean a different command in each menu. Every key is a row of `keys.md`, and you can change it.

### History

History is one browser with three levels: the commits, the files of the selected commit, and the diff of the selected file. The same browser shows a search through history, the history of one function, the commits that a pull or a push gives, and the changes of the working tree. The commits are cached, 100 by default.

### Protocols

A pull or a time travel with uncommitted changes keeps your changes, does the operation, and puts the changes back. Conflicts open in the conflict resolver. Time travel explores a past commit. Pin resets the branch to a past commit and marks it, and unpin returns. The branch picker switches, adds, merges, deletes, and rebases. Pull with rebase, rebase onto a branch, and interactive rebase work only on commits that are not pushed, so no force push is ever needed.

### Setup

A first run with no SSH key opens a wizard that makes a key and shows it. TIT can init a repository, clone one, and add a remote. Preferences set auto-update, its interval, the theme, and the count of history commits. TIT ships one theme. A theme is a folder of `~/.config/tit/themes/`. Repositories with Git LFS work: TIT fetches and checks out the LFS files.

### Banner

The TIT logo shows in `titc --help` and beside the menu.

---

## How

**Requirements:** C++17 compiler, CMake, Ninja, ccache, JUCE <juceVersion>, jam, cast. At run time: git; Git LFS for LFS repositories; `ssh-keygen` and `ssh-add` for the SSH wizard.

```bash
cast cast/spell.md --debug     # Debug build, installed to ~/.local/bin/titc
cast cast/spell.md --no-sign   # Release build, not signed
cast cast/spell.md             # Release build, signed installer, GitHub release
```

Run `titc` in a terminal, in a folder. `titc --help` prints the guide. `titc --version` prints the version.

The config lives in `~/.config/tit/`. `user.md`, `keys.md`, and `themes/gfx/theme.md` are written with their defaults on the first launch. A file with an error keeps the built-in defaults, and the footer shows the error. A change to a file takes effect at once, with no restart. TIT never reads or changes the `.toml` files of the Go TIT.

macOS: the installer is a signed and notarized pkg for `arm64` and `x86_64`. It installs `/opt/titc/titc` and links `~/.local/bin/titc` to it.

Windows: x64 and arm64 installers, signed. The installer installs `titc.exe` and adds its folder to your user PATH. TIT runs in a terminal with a Windows console, such as END.

---

## License

MIT

---

*conceived with [CAROL](https://github.com/jrengmusic/carol)*

Rock 'n Roll!

**JRENG!**
````

The sources of the text:

| Text | Source |
|---|---|
| "I have severe git skills issues." | `tit/___legacy___/README.md:7`, what the README already told the reader (`readme-writer/SKILL.md:36`) |
| "a git tool for a user who is not a git expert" | tracker, The contract of every part: *"tit as power git tools for non savvy git user"*; `PLAN-tit-2-spec.md` Step 1 |
| "always shows what is possible in the actual state"; "If it is in the menu, it works." | `tit/___legacy___/README.md:9,11`, in the README voice: *"always shows exactly what is possible based on your actual git state"*; *"If it's in the menu, it works. Period."* |
| "Git commands fail when you need them most." | `tit/___legacy___/README.md:9`: *"git commands that fail when I need them most"* |
| "pure function of git state"; "If git would reject an action"; "Destructive operations always ask first." | `tit/SPEC.md:20,26,33` (Part 2 Step 2 carries them) |
| The Go TIT paragraph | rulings 4, 5, and 6, verbatim words: *"hard to read, hard to understand"*, *"handroll text editor, with its some effort of core parity feature of vim"*, *"we have the data, but rendering probably limited by our past knowledge, and limitation of go/lipgloss"* |
| "keeps every feature, and fixes the design. Design by contract." | rulings 3 and 4: *"nothing should be dropped."*; *"kept the feature FIX THE DESIGN, DESIGN BY CONTRACT."* |
| "not a byte-to-byte translation" | ruling 42 |
| "a manifest, built with cast" | tracker, The contract of every part: *"conform with other project structured with manifest, build with cast"* |
| "JUCE is my expertise." | ruling 6: *"juce is ARCHITECT expertise"* |
| "built exactly like any GUI, on the JUCE native contract" | tracker, The contract of every part: *"tui could be build identically with any gui with juce native API contract using juce::Component base, juce::LookAndFeel (our Style etc)"* |
| "through its own peer" | rulings 13 and 14 |
| "terminal agnostic, and it renders its pixels itself" | ruling 24: *"TIT is terminal agnostic. render pixel itself then."* |
| "one state machine" | ruling 11; MANIFESTO E, "There is one state machine at each scale." |
| "Every command, callback, and keyboard shortcut goes through one command manager." | ruling 12 |
| "one text editor, with full features" | ruling 5: *"commit message text box would be MUCH BETTER using TextEditor. full feature"* |
| The history sentences and the closing question | ruling 40, in ARCHITECT's words, set in the README voice: *"we browse from top the commit, down to details per file"*; *"isnt that the whole premise of TIT?"* |
| `## What` | the tracker rows 8 to 13 and 15; SPEC.md after Part 2 |
| The menus of `### Menus and keys` | ruling 34, option text |
| "100 by default" | ruling 38 |
| `## How` commands | `archetype/ARCHETYPE.md:159-163`; `archetype/cli/init.cast:159-183` |
| ccache | `archetype/cli/cast/cmake.cast:24` (`REQUIRED`) |
| The config paragraph | Part 13 Overview, items 1 and 2; ruling 17 |
| The install sentences | `cmake.cast:165,252-261`; `archetype/cli/cast/installer.cast:113-121` |
| "a terminal with a Windows console, such as END" | tracker, Closed by a read, "Windows terminal target: ConPTY" |

**Validation:**
- The sections, their order, the `---` lines, and the footer match `readme-writer/SKILL.md:39-59`.
- `Why`, `Architecture`, and `What` name no file, class, or function. `How` names a file only where the reader opens it (`readme-writer/SKILL.md:64`).
- Each sentence of `What` and `How` holds in the built code and the built tree. A sentence that the code contradicts takes the fact of the code.
- Each command of the `bash` block is a row of the `## toolchain` table of `tit/project-info.md`.
- No claim of speed or size exists.
- `LICENSE` names MIT.

### Step 4: The macOS release
**Scope:** the `tit/` project on a Mac. Generated: `CMakeLists.txt`, `entitlements.plist`, `gh.cmake`, `RELEASE.md`, `cast/installer/*`.
**Owner:** ARCHITECT (CAROL.md Build; Language / Framework Constraints).
**Action:**
1. ARCHITECT sets the version: the cells `versionString`, `versionNumber`, and `buildVersion` of `tit/project-info.md`, and the line `**Version …**` of `tit/Source/HELP.md` (inventory, "The version and the repository").
2. From `tit/`, run `cast cast/spell.md`.
3. Run `pkgutil --check-signature "../Release/titc v<version> macOS.pkg"`.
4. Run `spctl --assess --type install -vv "../Release/titc v<version> macOS.pkg"`.
5. Run `xcrun stapler validate "../Release/titc v<version> macOS.pkg"`.
6. Open the pkg and install it.
7. Run `ls -l ~/.local/bin/titc`, then `titc --version`.
8. In END, in a repository fixture of Part 8 Step 9, run `titc`. Wait for the menu. Press Ctrl+C.

**Validation:**
- The run ends with exit code 0. The log shows `CAST: sign`, and the notary status `Accepted`.
- `RELEASE.md` holds the eight notes of Step 1 under `## What's New`.
- Item 3 prints `Status: signed by a developer certificate issued by Apple for distribution` and `Notarization: trusted by the Apple notary service`.
- Item 4 prints `accepted`, `source=Notarized Developer ID`, and an `origin=` line with the `installerIdentity` certificate.
- Item 5 prints `The validate action worked!`.
- The three expected outputs are measured on `Release/ling v0.1.0 macOS.pkg`, the pkg of the same lane, session 2026-11-14.
- `~/.local/bin/titc` is a link to `/opt/titc/titc`. `titc --version` prints `titc <version> (<commit>)`.
- Item 8 shows the menu of the fixture. This proves the signed binary runs the repository tool with no row of the type `entitlement` (Closed by a read, item 1).
- The QA folder `<qaDirectory>/<date> titc v<version> macOS` holds the pkg.
- `gh release view v<version> --repo jrengmusic/tit` lists the asset `titc v<version> macOS.pkg`.

### Step 5: The Windows releases
**Scope:** the `tit/` project on the Windows x64 machine, then on the Windows arm64 machine (inventory, "The machines").
**Owner:** ARCHITECT.
**Action:** On each machine:
1. From `tit/`, run `cast cast/spell.md`. The version is the version of Step 4.
2. Run `signtool verify /pa "..\Release\titc v<version> Windows.exe"`.
3. Run the installer. Open a new END tab.
4. Run `titc --version`.
5. In a repository fixture of Part 8 Step 9, run `titc`. Wait for the menu. Press Ctrl+C.

**Validation:**
- The run ends with exit code 0. The log shows `CAST: installer`.
- Item 2 ends with exit code 0. SignTool "returns an exit code of zero for successful execution, one for failed execution, and two for execution that completed with warnings" (learn.microsoft.com, `windows/win32/seccrypto/using-signtool-to-verify-a-file-signature`). The documentation gives no success text, so the exit code is the check.
- `titc --version` prints `titc <version> (<commit>)` from the PATH that the installer set.
- Item 5 shows the menu of the fixture.
- The GitHub release `v<version>` holds the asset of the machine. The second machine replaces the asset of the same name (`--clobber`, `cmake.cast:338`). The archive names of x64 and arm64 are the same, because the `platform` cell is `Windows` (`archetype/cli/init.cast:192`).

## BLESSED Alignment

- **B — Bound.** The toolchain owns each artifact: the archive, the QA copy, the release. No step writes a generated file by hand.
- **L — Lean.** The part adds eight table rows and two documents. It adds no script and no entitlement (YAGNI).
- **E — Explicit.** Each release check is a command with its expected output.
- **S — SSOT.** `project-info.md` and `cast/signing.md` are the one source of the lane. `RELEASE.md` is generated from the notes rows. The documents record the code, and the code is the source (Rule 0).
- **S — Stateless.** No step adds state.
- **E — Encapsulation.** The data flows one way: tables, cast, generated files, build, archive, release.
- **D — Deterministic.** The same tables and the same tree give the same archive name, the same notes, and the same documents.
- LANGUAGE.md: no override applies.

## Risks / Open Questions

None open.

The facts that the runs measure first:
- The keychain profile `notary` exists on the Mac of Step 4. The release of LING uses the same profile (`ling/CMakeLists.txt:434`). Not checked on this machine.
- The x64 and arm64 archives have the same name, and the second upload replaces the first (Step 5, Validation). The pack rows stay as the archetype generates them, as in the working siblings (`archetype/cli/init.cast:185-193`; `whatdbg/project-info.md:193-201`). Ruling 49: *"just follow the working pattern"*.
