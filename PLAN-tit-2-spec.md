# PLAN: TIT Part 2 — The Specification

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE). This part writes one document and no code.

## Overview

This part rewrites `tit/SPEC.md` as the contract of the C++ TIT.
It carries every feature of the Go TIT (ruling 3), corrects the lines where SPEC and the shipped code disagree, and adds the rulings of 2026-11-13: pin, three rebase actions, the reading contract, and commands and keys as tables.
A layout is not part of SPEC (ruling 7). SPEC states what each view answers and which commands it has.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Part 0 only.

## Language / Framework Constraints

- CAROL.md, Writing Standard: SPEC.md follows ASD-STE100. Exact strings stay verbatim.
- CAROL.md, Decision Gate: "Writing SPEC.md, PLAN.md, or ARCHITECTURE.md is gated." ARCHITECT's lock of this part is the gate.
- `~/.carol/SPEC-WRITER.md:312-333`: no ambiguity, complete flows, exact strings for every user-facing text.
- `~/.carol/SPEC-WRITER.md:190-308`: the section set is Overview, Technology Stack, Core Principles, Features, State Model, UI, Success Criteria. `eve/SPEC.md:5-973` has this set and adds Components, Keyboard Summary, and Architecture Constraints.
- NAMES.md Rule 6: `kind` is banned; the word is `Type`.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/`. `legacy/` stands for `tit/___legacy___/internal/`.

### Where SPEC and the shipped code disagree

Each row is corrected in the step that the last column names.

| SPEC line | The code | Step |
|---|---|---|
| "Target: Go + Bubble Tea + Lip Gloss" (`tit/SPEC.md:4,11-14`) | the target is C++17 / JUCE 8 / JAM, built with cast (tracker, The project side) | 1 |
| Invariant 7: "No manual abort in menu"; invariant 11: "Merge/Rebase/DirtyOperation has no abort menu option" (`tit/SPEC.md:1025,1029`) | `menuMerging` shows `abort_merge`; `menuRebasing` shows `rebase_abort`; `menuConflicted` and `menuDirtyOperation` show `abort_merge` (`legacy/app/menu_render_extra.go:117-147`). SPEC §5 itself gives "Finalize merge" / "Abort merge" (`tit/SPEC.md:154-157`). | 2 |
| Invariant 8: "Lip Gloss rendering" (`tit/SPEC.md:1026`) | no Lip Gloss in the C++ TIT | 2 |
| Invariant 10: "No Configuration" (`tit/SPEC.md:1028`) | a config file and a Preferences menu exist (`legacy/config/config.go:14-22,48`; `tit/SPEC.md:313-323`) | 2 |
| `ssh-keygen -t ed25519` (`tit/SPEC.md:763`) | `ssh-keygen -t rsa -b 4096` (`legacy/git/ssh.go:37`) | 5 |
| Theme file `default.xml`, hot reload through `jam::File::Watcher`, 200 ms (`tit/SPEC.md:1003-1013`) | markdown tables through cast (tracker, Closed by a read); ruling 17 for the TOML files | 9 |
| "See `IMPLEMENTATION_PLAN.md`" (`tit/SPEC.md:1036`) | no such file at the root | 1 |
| Section numbers: two sections carry `13.x`; the keyboard section is `15` with `14.x` items (`tit/SPEC.md:732,836,942-944`) | — | 1 |

### What the shipped code has and SPEC does not

| Feature | Where | Step |
|---|---|---|
| The startup gate: with a remote, the menu waits for the first fetch. Text `Checking Remote...` with a spinner; on failure the footer shows `Remote unreachable — showing cached state`. | `legacy/app/app_constructor.go:209-210`; `legacy/app/app_view_main.go:167-173`; `legacy/app/app_update_msg.go:209-222` | 5 |
| History items are disabled while the history cache builds. Labels `History...`, `History <n>/<total>`, `Files...`, `Files <n>/<total>`, with a spinner frame. | `legacy/app/menu_render_extra.go:18-59` | 4 |
| `Conflicted` and `DirtyOperation` menus | `legacy/app/menu_render_extra.go:133-147` | 4 |
| LFS: detection, filter setup at startup, `lfs fetch` after a fetch, `lfs checkout` after a hard reset, an `LFS` or `LFS ⚠` badge in the header | `legacy/git/lfs.go:10-47`; `legacy/git/state.go:50`; `legacy/app/app_constructor.go:78-90`; `legacy/app/op_remote.go:66`; `legacy/app/op_pull.go:89-95`; `legacy/app/app_view_header.go:113-124`; `legacy/ui/header.go:33-34,130-133` | 5 |
| The stash tracking list | `legacy/config/stash.go:16-17,41` | 6 |
| Progress lines replace the last console line: `\r` replaces, `\n` appends | `legacy/git/execute.go:141-214` | 5 |
| `--progress` on each streaming command | `legacy/app/handlers_git_pull.go:20,51`; `legacy/app/op_push.go:21-45`; `legacy/app/op_clone.go:26`; `legacy/app/op_pull.go:22,68` | 5 |
| The conflict resolver labels name the branches: `main (current)` / `dev (incoming)` | `legacy/app/handlers_conflict.go:157-218` | 6 |
| The nested repository dialog, the alert dialog, the stale stash dialog | `legacy/app/confirm_dialog_render.go:119-159`; `legacy/app/confirm_time_travel_return.go:38-42,97-101` | 8 |
| The stale marker rule: a time travel marker with HEAD on a branch is removed | `legacy/git/state_detection.go:188-199` | 3 |

### The exact strings that SPEC carries

`~/.carol/SPEC-WRITER.md:327-333` requires exact strings. Each set has one source.

| Set | Source |
|---|---|
| Menu item id, label, shortcut, hint (40 items) | `legacy/app/menu_items.go:8-306` |
| Confirmation title, explanation, and the two button labels (18 keys) | `legacy/app/messages_dialog.go:17-125` |
| Dialog messages | `legacy/app/messages_dialog.go:129-142` |
| Error messages | `legacy/app/messages_error.go:4-59` |
| Output messages, console messages, state descriptions | `legacy/app/messages_state.go:4,75,120` |
| Footer messages and footer hint rows by mode | `legacy/app/messages_menu.go:29-53,68-176` |
| Timeline sync messages | `legacy/app/messages_menu.go:179-185` |

### The rulings that add content

| Content | Ruling | Step |
|---|---|---|
| Pin, the pinned menu, unpin | 1, 8 | 3, 4, 7 |
| Pull with rebase, rebase onto a branch, interactive rebase | 2, 9, 10 | 4, 7 |
| The reading contract | 4, 6 | 6 |
| Text input and the commit box are `jam::TextEditor` | 5 | 6 |
| Commands and keys as tables | 12 | 8 |
| No layout in SPEC | 7 | 6, 9 |
| The new files beside the TOML files | 17 | 9 |

### The families of the new words

| New word | Family | Where |
|---|---|---|
| Operation value `Pinned` | `Conflicted`, `Normal` (adjective form of the Operation table) | `tit/SPEC.md:74-84` |
| Marker `.git/TIT_PIN` | `.git/TIT_DIRTY_OP`, `.git/TIT_TIME_TRAVEL` | `tit/SPEC.md:124,499` |
| Menu item id `pull_rebase` | `pull_merge`, `pull_merge_diverged`; the legacy step name is `pull_rebase` | `legacy/app/menu_items.go:83,101`; `legacy/app/handlers_git_pull.go:56` |
| Menu label `Pull (rebase)` | `Pull (fetch + merge)`, `Pull (merge)`, `Pull (save changes)` | `legacy/app/menu_items.go:74-101` |
| Menu item id `unpin`, label `Unpin` | ARCHITECT's word (ruling 1, option text) | tracker, Canon 1 |
| Confirmation key `pin` | `rewind`, `time_travel` | `legacy/app/confirm_dialog.go:19,24` |
| Shortcut on a menu that never shows with its twin | `config_add_remote` and `config_remove_remote` share `r` | `tit/SPEC.md:299-301` |

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

In this part, COUNSELOR writes the document. No @Engineer delegation exists, because no code changes.

## Steps

### Step 1: The frame of the document
**Scope:** `tit/SPEC.md`: the title block, Overview, Technology Stack, the section order.
**Action:**
1. Title: `# TIT Specification`. Purpose, Target End-User, and Core Workflow lines, as `~/.carol/SPEC-WRITER.md:194-201`. The end-user is ARCHITECT's sentence: a user who is not a git expert (tracker, The contract of every part, first quote).
2. Technology Stack: C++17, JUCE 8, JAM, cast; macOS (arm64 and x86_64) and Windows; the repository tool runs as a child process only, as `tit/SPEC.md:13` rules ("`os/exec` only, no libraries"); one binary, `titc`.
3. Section order: Overview, Technology Stack, Core Principles, State Model, Features, Commands and Keys, Configuration, Success Criteria, Architecture Constraints.
4. Number each section once. Remove the line that names `IMPLEMENTATION_PLAN.md`.

**Validation:** No line names Go, Bubble Tea, Lip Gloss, or `IMPLEMENTATION_PLAN.md`. Each section number is used once.

### Step 2: Core Principles
**Scope:** `tit/SPEC.md`: Core Principles.
**Action:**
1. Carry `tit/SPEC.md:18-33` (Foundational Principle) with no change of meaning.
2. Carry the twelve invariants of `tit/SPEC.md:1019-1030`, with these corrections:
   - 7 and 11: an abort command exists in the Merging, Rebasing, Conflicted, and DirtyOperation menus (`legacy/app/menu_render_extra.go:117-147`; `tit/SPEC.md:154-165`). The sentence "git state safe even if user exits" stays.
   - 8: "Beautiful: themed colours". Remove "Lip Gloss rendering".
   - 10: keep "State always reflects actual Git state". Replace the label "No Configuration" with "No stored repository state". The user configuration of Step 9 is not repository state.
3. Add one invariant from ruling 9: a rewrite of commits never needs a force push; the command is absent when it would.

**Validation:** Each corrected invariant cites nothing; SPEC is the contract. The plan citations above are the proof. Invariants 1 to 6, 9, and 12 are unchanged in meaning.

### Step 3: State Model
**Scope:** `tit/SPEC.md`: State Model.
**Action:**
1. Carry the five axes and `IsTitTimeTravel` (`tit/SPEC.md:37-105`) and the priority rules (`tit/SPEC.md:109-140`).
2. Operation table: add the row `Pinned` — "The branch is reset to a past commit by TIT. The marker `.git/TIT_PIN` exists."
3. Priority 3 (Operation State): add `Pinned` → the pinned menu of Step 4.
4. Add a marker table:

   | Marker | Lines | Written by | Removed by |
   |---|---|---|---|
   | `.git/TIT_DIRTY_OP` | branch; original HEAD | the Dirty Operation Protocol, step 1 (`tit/SPEC.md:350-357`) | its finalize and its abort (`tit/SPEC.md:370-385`) |
   | `.git/TIT_TIME_TRAVEL` | branch | entering time travel (`tit/SPEC.md:497-501`) | the return (`tit/SPEC.md:589-593,611-615`) |
   | `.git/TIT_PIN` | branch; original HEAD; pinned hash | pin (Step 7) | unpin (Step 7) |

5. Stale markers: carry the rule of `legacy/git/state_detection.go:188-199` for the time travel marker. The pin marker is stale when HEAD is not on the marker's branch, or HEAD is not the pinned hash. TIT removes a stale marker at state detection.
6. Timeline while `Pinned`: empty (not applicable), as for `TimeTraveling` (`tit/SPEC.md:72`).
7. Add the LFS fields of the state: the repository uses LFS; the LFS filters are ready (`legacy/git/types.go:88`; `legacy/git/state.go:50`).

**Validation:** The Operation table has nine rows. Each marker row matches its cited SPEC lines. No state lives outside these tables.

### Step 4: The menus
**Scope:** `tit/SPEC.md`: Feature "Menu".
**Action:**
1. Carry `tit/SPEC.md:183-323` (NotRepo, TimeTraveling, Normal, Config, Preferences).
2. Add the menus that the code has and SPEC lacks:

   | Operation | Items | Where |
   |---|---|---|
   | `Merging` | `finalize_merge`, `abort_merge` | `legacy/app/menu_render_extra.go:117-123` |
   | `Rebasing` | `rebase_continue`, `rebase_abort` | `:125-131` |
   | `Conflicted` | no menu: the conflict resolver opens, and Escape aborts | `PLAN-tit-12-protocols.md`, Closed by a read, items 6 and 8; legacy `:133-139` |
   | `DirtyOperation` | `dirty_continue` (Reapply changes), `dirty_abort` (Abort and restore) | ruling 41; `PLAN-tit-12-protocols.md`, Closed by a read, item 4; legacy `:141-147` showed `abort_merge` only (Part 12, legacy fault 1) |

3. History items while the cache builds: carry the four labels and the disabled state (`legacy/app/menu_render_extra.go:18-59`): `History...` and `Files...` before a count exists, then `History 12/100` and `Files 12/100`.
4. Timeline tables: add `pull_rebase` to `Behind` + `Clean` after `pull_merge`, and to `Diverged` + `Clean` after `pull_merge_diverged` (ruling 2). It is absent when `Dirty`.
5. Add the `Pinned` menu (ruling 8): `history`, `file_history`, `reset_discard_changes` (only when `Dirty`), `unpin`. No commit item. No timeline item.
6. Add the item rows below, and the rows `view_changes` and `incoming_outgoing` of "The exact text of the new items" with their groups:

   | id | label | shortcut | hint |
   |---|---|---|---|
   | `pull_rebase` | Pull (rebase) | `r` | Fetch remote and replay local commits on top |
   | `unpin` | Unpin | `u` | Return the branch to where it was before the pin |

   `r` does not clash: `add_remote` (`r`) shows only with no remote, and `pull_rebase` shows only with a remote, as the Config menu pairs `r` (`tit/SPEC.md:299-301`).

**Validation:** Each legacy item of `legacy/app/menu_items.go:8-306` is in a table of SPEC with its label, shortcut, and hint. The four menu rows of "The exact text of the new items" are the only new items of the main menus.

### Step 5: Startup, environment, console, LFS
**Scope:** `tit/SPEC.md`: Features "Startup" and "Console".
**Action:**
1. Carry `tit/SPEC.md:732-834` (environment check, setup wizard, repository check, branch mismatch, fatal errors), except two blocks that are not built (ruling 48; Part 13, Closed by a read, item 24): `:770-776`, the check of `user.name` and `user.email`, and `:825-832`, the bare repository. In their place write what the program does: the console shows the text of the tool.
2. Key generation: write `ssh-keygen -t rsa -b 4096 -C "<email>"`, as the code runs (`legacy/git/ssh.go:37`).
3. Add the startup gate with its two exact strings (inventory, row 1 of the second table).
4. Console: carry the ESC rules of `tit/SPEC.md:950-956`. Add: each streaming command has `--progress`; a progress line replaces the last console line.
5. Add Feature "LFS" from the inventory row. It has no menu item and no key.

**Validation:** Each added sentence has a row in the inventory tables. The setup wizard keeps its seven phases (`tit/SPEC.md:756-766`).

### Step 6: The reading contract, text input, conflict resolver
**Scope:** `tit/SPEC.md`: Features "History", "File history", "Commit", "Conflict resolver", "Branch picker", "Dirty Operation Protocol", "Time travel".
**Action:**
1. Carry with no change of meaning: the Dirty Operation Protocol (`tit/SPEC.md:327-392`), the Branch Picker (`:396-462`), time travel (`:466-625`), manual detached HEAD (`:836-912`), rewind (`:654-675`), Copy Hash (`:989-991`), mid-operation recovery (`:144-179`). Three points take the built form of Part 12 in place of the old text:
   - The return from time travel with a dirty tree is one dialog with two buttons, `Merge changes` and `Discard changes`; Escape cancels. This replaces the three buttons of `tit/SPEC.md:215` (Part 12, Closed by a read, item 12).
   - The `Stash Not Found` dialog is carried: at the return, when the stash of the entry is gone, the program asks before it continues (`legacy/app/confirm_time_travel_return.go:25-45,90-104`).
   - A time travel that the program finds at its start stays a time travel, and the menu gives the return (Part 12, Closed by a read, item 15).
2. Remove each layout drawing of `tit/SPEC.md:406-418,634-644,691-709,918-929` (ruling 7). Keep the list of what each view shows.
3. Feature "History" is one browser of records (rulings 39 and 40). It has three levels, from the top down: the commits; for the selected commit its message, author, date, and its files with added and deleted line counts; for the selected file its diff. It answers "What changed in this commit?" and "How big is each change?". The browser is the one view of every feed of this step.
4. The legacy items "History" and "File(s) history" open that same browser: the first at the commits, the second at the files of the selected commit. Legacy File history is a commit, its files, and one diff (`legacy/ui/filehistory.go:22-31`), which are the three levels. TIT has no blame (ruling 39).
5. Every diff in TIT marks the changed characters inside a changed line, and shows the function name of each hunk.
6. (Removed by ruling 39: no blame.)
7. State the data of each view as the tool output that gives it (tracker, The git side: the reading contract). SPEC names the output, not a layout.
8. Text input: the commit message, a branch name, a clone URL, and an email are typed in a `jam::TextEditor` in `insert` mode (rulings 5, 16). The commit message is a multi-line editor with the vim core. Carry the ESC rules for input of `tit/SPEC.md:951-952`.
9. Conflict resolver: carry `tit/SPEC.md:148-165`. Add the branch-name labels (inventory). Add the record of the stash of a time travel: the second line of the time travel marker holds the label of the stash of the entry (Part 12, Closed by a read, item 14). The legacy stash tracking list is not carried (ruling 17).
10. Four more feeds of the browser (ruling 19). Each states its question and the tool output of the tracker table, and no layout:
    - "Working tree changes": when the tree is `Dirty`, one record with the changed files, their added and deleted line counts, and the diff of the selected file. It is a menu item of the Working Tree group.
    - "Incoming and outgoing": when the Timeline is `Behind`, `Ahead`, or `Diverged`, the commits that a pull gives and the commits that a push gives, as two groups of records. Each commit shows its files and their diffs.
    - "Search history": a command of the browser. The user types a text; the browser shows the commits where that text appeared or vanished.
    - "Function history": a command of the browser on the selected file. The user types a function name; the browser shows the commits that changed that function, each with its diff.

**Validation:**
- Each reading question of the tracker table that this step names has one view that answers it.
- No drawing of a layout remains in SPEC.
- Ruling 3: every feature of the carried sections is present.

### Step 7: Pin and the three rebase actions
**Scope:** `tit/SPEC.md`: Features "Pin" and "Rebase".
**Action:**

**Pin** (rulings 1 and 8):
1. Entry: a command of the History view on a selected commit, beside time travel and rewind (`tit/SPEC.md:647-652`).
2. `Clean` tree: write `.git/TIT_PIN` (branch, original HEAD, pinned hash), then `reset --hard <hash>`. The marker is written first, as time travel writes its marker first (`tit/SPEC.md:497-501`).
3. `Dirty` tree: the Dirty Operation Protocol with the reset as its change-set (`tit/SPEC.md:329-334,359`).
4. With LFS: `lfs checkout` follows the reset, as it follows each hard reset (`legacy/app/op_pull.go:89-95`).
5. While `Pinned`: the menu of Step 4. The header shows the branch and the pinned hash. Changes in the tree are allowed. No commit (ruling 8).
6. Pin again from History while `Pinned`: the marker keeps its original HEAD and takes the new pinned hash, as time travel jumps to another commit (`tit/SPEC.md:239`).
7. Unpin: `reset --hard <original HEAD>`, then remove the marker. `Dirty` tree: the Dirty Operation Protocol.

**Pull with rebase** (ruling 2): `pull --rebase --progress` (`legacy/app/handlers_git_pull.go:51`). A conflict gives `Rebasing` or `Conflicted`, and the existing loop takes over (`tit/SPEC.md:159-165`).

**Rebase onto a branch** (ruling 2): a command of the Branch Picker on another branch, beside merge (`tit/SPEC.md:431`). It is hidden on the current-branch row, as merge is. It runs `rebase <selected>`. `Dirty` tree: the Dirty Operation Protocol, as merge has (`tit/SPEC.md:462`). It is active only when every commit that it replays is unpushed; otherwise it is inactive, and the footer says why (ruling 20).

**Interactive rebase** (rulings 9 and 10):
1. Entry: a command of the History view. It is active only when the tree is `Clean` and at least one commit is ahead of the remote; with no remote, every commit of the branch is in range.
2. The range is the commits of `@{u}..HEAD`. Commits outside the range cannot be selected.
3. The edits: squash, fixup, reword, drop, move up, move down. Reword opens the commit message editor of Step 6.
4. Apply runs the rebase with the edited list. Cancel changes nothing.
5. A conflict gives the existing rebase loop. Abort is `rebase --abort`.

**Validation:** Each command line is quoted from a ruling or a cited source. Each flow has its `Clean` path, its `Dirty` path, its conflict path, and its abort path (`~/.carol/SPEC-WRITER.md:321-325`).

### Step 8: Commands and keys
**Scope:** `tit/SPEC.md`: "Commands and Keys".
**Action:**
1. Replace the prose of `tit/SPEC.md:942-999` with two tables.
2. Table "commands": one row for each command: name, the views or modes where it exists, the state rule that makes it active, description. A command that the repository tool would reject is inactive: while its state rule fails, the menu has no item for it and the key sheet has no row for it (rulings 55 and 56; tracker, "Command availability"). A change of the state builds the menu and the key sheet again. Two commands keep their key and write a footer text in place of their work: the rebase onto a branch with pushed commits, and the interactive rebase with a dirty tree or with no unpushed commit (rulings 20 and 22).
3. Table "keys": the form of `eve/Source/layout/default-eve.md:44-59`: `key`, `type`, `value`, `mode`, `description`. One row per command. `value` holds one or more key descriptions.
4. Carry every binding of `legacy/app/app_keys.go:8-18,36-150` as a default row, with its mode.
5. Carry each dialog as a row set: the 18 confirmation keys with title, explanation, and labels (`legacy/app/messages_dialog.go:17-125`), plus the inline dialogs (`legacy/app/confirm_dialog_render.go:119-159`; `legacy/app/confirm_time_travel_return.go:38-42,97-101`). Four rows of the legacy text table are not written, because no legacy menu or command shows them: `time_travel_merge`, `time_travel_merge_dirty`, `time_travel_return_dirty`, `branch_switch_clean` (Part 12, Closed by a read, item 13). The `Stash Not Found` dialog is written (`legacy/app/confirm_time_travel_return.go:25-45,90-104`).
6. Add the default rows of the new commands (Risks, item 1).
7. Carry the footer hint rows by mode (`legacy/app/messages_menu.go:68-176`), and the error, output, and state strings (inventory, The exact strings).

**Validation:** Each key of `legacy/app/app_keys.go` has a row. Each of the 33 dispatchers of `legacy/app/dispatchers.go:9-46` has a command row. No binding lives in prose.

### Step 9: Configuration, Success Criteria, Architecture Constraints
**Scope:** `tit/SPEC.md`: the last three sections.
**Action:**
1. Configuration: the user folder `~/.config/tit/` (`legacy/config/config.go:48`). The files are markdown tables. TIT writes a missing file with defaults, and never reads or changes a `.toml` file (ruling 17). Carry the three preferences with their rules: auto-update on or off, interval 1 to 60 minutes with default 5, theme (`tit/SPEC.md:317-323`; `legacy/config/config.go:14-22,105-110,147-153`). State the one shipped theme and the user theme folders of ruling 43, and carry the 13 colour families (`tit/SPEC.md:1009-1013`).
2. Remove the XML sentences of `tit/SPEC.md:1005-1007`.
3. Success Criteria: one testable line for each feature, as `~/.carol/SPEC-WRITER.md:296-308`.
4. Architecture Constraints: the state is a `jam::Model` (ruling 11); commands go through `jam::CommandManager` (ruling 12); every view is a `juce::Component` in a `TerminalWindow` (ruling 13); text is `jam::TextEditor` (ruling 5); no layout is fixed in SPEC (ruling 7).

**Validation:** `~/.carol/SPEC-WRITER.md:426-437`, the output checklist, passes. Each ruling of the tracker Canon that states product behavior has a sentence in SPEC.

## BLESSED Alignment

- **B — Bound.** Each marker has one writer and one remover (Step 3 table).
- **L — Lean.** SPEC adds only what a ruling or the shipped code names. No reading view is added beyond the rulings (Risks, item 2).
- **E — Explicit.** Every user-facing string is exact. Every command has a state rule.
- **S — SSOT.** One table holds the commands, one the keys, one the markers. The legacy string maps are the one source of each carried string.
- **S — Stateless.** SPEC gives no state to a view. The state model is the five axes, the markers, and the rows of Part 7.
- **E — Encapsulation.** SPEC states behavior and data, never a layout or a class.
- **D — Deterministic.** The same state tuple gives the same menu and the same active commands.
- LANGUAGE.md: no override applies.

## Risks / Open Questions

Each item closes before this part locks.

1. **Default keys and names of the new commands.** Closed by ruling 21: the rows below stand. They follow the family of first letters that the Branch Picker uses (`a` add, `m` merge; `tit/SPEC.md:430-431`).

   | Command | View | Proposed key |
   |---|---|---|
   | pin | History | `p` |
   | interactive rebase | History | `i` |
   | squash | rebase list | `s` |
   | fixup | rebase list | `f` |
   | reword | rebase list | `r` |
   | drop | rebase list | `d` |
   | move down, move up | rebase list | `J`, `K` |
   | apply, cancel | rebase list | `enter`, `escape` |
   | rebase onto | Branch Picker | `r` |

   Pin has no `ctrl` key: unpin reverses it, and invariant 12 keeps `ctrl` for destructive commands (`tit/SPEC.md:1030`).
2. **The reading views beyond History and File history.** Closed by ruling 19: working tree changes, what a pull or push gives, search through history, and function history are in SPEC (Step 6, item 10). Moved-code marking stays out.
3. **The exact text of the new dialogs and hints.** Closed by ruling 22: the strings are in the section "The exact text of the new items" of this file.
4. **Rebase onto a branch whose commits are pushed.** Closed by ruling 20: the command is active only when every commit it replays is unpushed. Otherwise it is inactive, and the footer says why. Step 7 states this rule in the rebase onto flow.

## The exact text of the new items (ruling 22)

COUNSELOR writes these strings into SPEC verbatim. Step 4 takes the menu rows. Step 8 takes the dialogs, the footer texts, and the keys. `{name}` marks a value that TIT fills in.

### Menu rows

| id | label | shortcut | hint | Group |
|---|---|---|---|---|
| `pull_rebase` | Pull (rebase) | `r` | Fetch remote and replay local commits on top | Timeline, `Behind` or `Diverged`, `Clean` |
| `unpin` | Unpin | `u` | Return the branch to where it was before the pin | `Pinned` |
| `view_changes` | View changes | `v` | View uncommitted changes file by file | Working Tree, `Dirty` |
| `incoming_outgoing` | Incoming / outgoing | `o` | View the commits a pull or push will transfer | Timeline, `Behind`, `Ahead`, or `Diverged` |

### Dialogs

| key | title | explanation | yes | no |
|---|---|---|---|---|
| `pin` | Pin Confirmation | This will move {branch} to commit {hash}.<br><br>The commits after it stay safe. Unpin returns the branch to where it is now.<br><br>You cannot commit while pinned.<br><br>Pin here? | Pin | Cancel |
| `pin_dirty` | Save your changes? | You have uncommitted changes. To pin, they must be temporarily saved.<br><br>After the pin, we'll try to reapply them.<br>(This may cause conflicts if the changes overlap.) | Save changes | Discard changes |
| `unpin` | Unpin? | This will return {branch} to commit {hash}, where it was before the pin.<br><br>Continue? | Unpin | Cancel |
| `unpin_dirty` | Save your changes? | You have uncommitted changes. To unpin, they must be temporarily saved.<br><br>After the unpin, we'll try to reapply them.<br>(This may cause conflicts if the changes overlap.) | Save changes | Discard changes |
| `pull_rebase` | Pull with rebase? | This will replay your local commits on top of the remote changes.<br><br>If both modified the same files, conflicts may occur.<br>You'll be able to resolve them interactively. | Proceed | Cancel |
| `rebase_branch` | Rebase onto branch? | This will replay the commits of {current} on top of {selected}.<br><br>Conflicts will be handled if they occur. | Rebase | Cancel |
| `rebase_branch_dirty` | Uncommitted changes | You have uncommitted changes.<br><br>To rebase, they must be temporarily stashed.<br>After the rebase, they'll be reapplied.<br><br>(This may cause conflicts if changes overlap.) | Stash and rebase | Cancel |
| `rebase_interactive` | Rewrite commits? | This will rewrite {count} unpushed commits.<br><br>Conflicts will be handled if they occur.<br><br>Continue? | Rewrite | Cancel |

The siblings: `dirty_pull` and `pull_merge` (`legacy/app/messages_dialog.go:30-41`), `time_travel` (`:60-65`), `remove_remote`, `merge_branch`, `merge_branch_dirty` (`:108-125`).

### Footer text when a command is inactive

| Command | State | Text |
|---|---|---|
| rebase onto | a commit to replay is pushed | Pushed commits cannot be rebased |
| interactive rebase | no commit is ahead of the remote | No unpushed commits to rewrite |
| interactive rebase | the tree is `Dirty` | Commit or discard changes first |

### Other text

| Where | Text |
|---|---|
| Header while `Pinned` | PINNED @ {hash} |
| Search history prompt | Search history: |
| Function history prompt | Function history: |
| Search history with no result | No commit matches |

### Keys added by ruling 22

| Command | View | Key |
|---|---|---|
| search history | History | `/` |
| function history | History, on the selected file | `F` |
