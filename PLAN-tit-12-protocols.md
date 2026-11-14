# PLAN: TIT Part 12 — The Protocols

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part builds the protocols of TIT: the operations that take more than one command, and that can stop at a conflict.

1. **The state of a protocol is in the repository.** A protocol is one chain of command lines. When a chain stops, the repository holds what remains: a marker file, a stash of TIT, a merge, a rebase, unmerged files. The detection of Part 8 reads them into rows. The menu and the commands follow the rows. No object holds a phase.
2. **The Dirty Operation Protocol.** A pull, a merge, a branch switch, a rebase, a pin, or an unpin with a dirty tree saves the changes in a stash of TIT, does the operation, and applies the changes again. An interrupted protocol shows the `DirtyOperation` menu with two commands: apply the saved changes, or abort and restore.
3. **The conflict resolver.** One list of the unmerged files and three text panes: the base, the current side, the incoming side. The user chooses one version for each file. The resolver finalizes by the state: a merge commit, the next rebase step, or the drop of the stash.
4. **Time travel.** Enter on a commit of the history browser checks that commit out. The time travel menu returns to the branch, with or without the changes of the journey. A detached HEAD that TIT did not make returns through the branch picker.
5. **Pin** (rulings 1 and 8) and **unpin**.
6. **The branch picker.** One list of the branches and one details pane. Switch, add, merge from, delete, and rebase onto (ruling 2).
7. **The three rebase actions** (rulings 2, 9, 10, 20): pull with rebase, rebase onto a branch, and the interactive rebase of the unpushed commits from the history browser.
8. **The commit date in the header** during time travel.
9. **The layers** (Part 9, Overview, item 6). The processor holds each protocol: its command lines, its markers, and the files of the rebase list. The resolver and the picker are each one screen view, `ConflictView` and `BranchView`. `TITView` registers the commands in families and tells the processor and the views.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 8, 9, 10, and 11. Step 1 adds rows to one cast table of JAM.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization, and parentheses for an initializer of the same container type; no `==` on strings; no out-parameter; no fake-carrier struct; no local struct; no anonymous namespace; no non-owning naked pointer as a member.
- CODING.md, "Uniform signature in the type" and "Consume the event payload".
- MANIFESTO L: no `if` chain longer than three branches; a longer decision is a table. No function body longer than 30 lines.
- MANIFESTO S (Stateless): "State belongs exclusively to the Model." The Model rows of a protocol are a reflection of the repository. The records, the selected record of each tree, the chosen versions, and the rebase list are Model rows. The highlight of a menu is not a row (ruling 61).
- MANIFESTO E: "A manual boolean flag, a manual callback, or a helper invented where a listener pattern already exists is not a solution". A protocol continues through the detection and the listeners of Parts 8 to 11. This part adds no callback and no phase flag.
- MANIFESTO E, Complete at Creation: each pool job builds its tree complete, with the command lines of each choice.
- Ruling 7: a layout locks at execution. Each geometry number is a named constant. ARCHITECT changes a layout at the run of Step 14.
- ARCHETYPE.md §1: no step edits a generated file by hand. Each cast table keeps its `+---` separator lines.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build and Git: agents build only on ARCHITECT's explicit request. The fixture repositories need git commands; ARCHITECT runs them, or gives the git instruction.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`. The words of the tool are from its official documentation, `Documentation/` of its source repository, and from its source where a row names a `.c` file.

### Rulings of this part, verbatim

- **Ruling 1, pin.** Selected: *"Reset + pin marker"*. Option text: "`git reset --hard <hash>` on the current branch, plus a TIT marker file as `.git/TIT_TIME_TRAVEL` (SPEC.md:499). While the marker exists, the menu hides Pull and Replace local and shows Unpin."
- **Ruling 2, rebase.** Selected: *"Pull with rebase, Rebase onto branch, Interactive rebase"*.
- **Ruling 8, the pinned menu.** Selected: *"No commit while pinned"*. Option text: "Menu: History, File history, Discard changes, Unpin. Unpin with a dirty tree uses the Dirty Operation Protocol."
- **Ruling 9.** Selected: *"Unpushed commits only"*. Option text: "Only commits ahead of the remote (`@{u}..HEAD`). A branch with no remote can rewrite all its commits."
- **Ruling 10.** Selected: *"Squash / fixup, Reword, Drop, Reorder"*.
- **Ruling 20.** Selected: *"Same limit"*. Option text: "The command is active only when every commit it replays is unpushed. Otherwise it is inactive, and the footer says why."
- **Ruling 21, the keys.** History, pin `p`, interactive rebase `i`. Rebase list, squash `s`, fixup `f`, reword `r`, drop `d`, move down `J`, move up `K`, apply `enter`, cancel `escape`. Branch Picker, rebase onto `r`. Main menu, Pull (rebase) `r`, Unpin `u`.
- **Ruling 22, the texts.** The strings of `PLAN-tit-2-spec.md`, "The exact text of the new items".
- **Ruling 41, the texts of this part.** Selected: *"They stand"*. The eight texts of the section "The exact text of the new items of this part".
- **Ruling 35.** Selected: *"addRemote, calls pullRebase"*. Option text: "Row `addRemote | r | menuNormal`. With no remote it adds a remote. With a remote it calls `pullRebase`."
- ARCHITECT, session 2026-11-13: *"no manual callbacks, fully event driven. no fucking reinvention of new pattern, what is the established pattern?"*

### SPEC lines that this part builds

| Subject | SPEC |
|---|---|
| The Dirty Operation Protocol: snapshot, change-set, apply back, finalize, abort | `tit/SPEC.md:327-392` |
| An interrupted dirty operation: "Resume: pick up from the phase recorded in the marker", "Abort: restore original state from marker data" | `tit/SPEC.md:167-171` |
| Conflicted: "User picks file versions, stages resolved files"; finalize by the operation; abort | `tit/SPEC.md:148-152` |
| Rebasing: "If next commit also conflicts → loop back to conflict resolver" | `tit/SPEC.md:159-165` |
| The time travel menu and the two paths of "Return to [branch]" | `tit/SPEC.md:203-244` |
| Time travel: enter, merge back, return | `tit/SPEC.md:466-625` |
| A detached HEAD that TIT did not make | `tit/SPEC.md:836-912` |
| The branch picker | `tit/SPEC.md:396-462,973-981` |
| The menu rows of a dirty tree that is `Behind` or `Diverged`: "🔺 Pull (save changes)" | `tit/SPEC.md:268-278` |
| Pin, the three rebase actions, their dialogs and texts | `PLAN-tit-2-spec.md`, Step 7 and "The exact text of the new items" |

### The legacy code that this part carries (ruling 3)

| Piece | Legacy | Where |
|---|---|---|
| The dirty marker: two lines, the branch and the hash of HEAD | `content := fmt.Sprintf("%s\n%s\n", branchName, headHash)` | `legacy/git/dirtyop.go:19-45` |
| The chain of a dirty pull: `stash push -u -m <label>`; or `reset --hard` and `clean -fd`; `pull --no-rebase --progress`; `stash apply`; `stash drop` | | `legacy/app/op_dirty_pull_snapshot.go:31-169`; `legacy/app/op_dirty_pull_merge.go:21` |
| The chain of a dirty merge and of a dirty switch: the same, with `merge <branch>` and `switch <branch>` | | `legacy/app/op_dirty_merge.go:59,108`; `legacy/app/op_dirty_switch.go:60,110` |
| The abort of a dirty operation: `checkout <branch>`, `reset --hard <head>`, `stash apply`, `stash drop` | | `legacy/app/op_dirty_pull_snapshot.go:184-253` |
| The menu row `dirty_pull_merge`: key `{`, shown as `shift + [`, 🔺, `Pull (save changes)` | | `legacy/app/menu_items.go:74-82`; `legacy/app/menu_render_core.go:119-141` |
| The unmerged files: `status --porcelain=v2`, the records that start with `u`; the path is the text after field 10 | | `legacy/git/exec_conflict.go:10-46` |
| The three versions of a file: `show :<stage>:<path>`; stage 1 base, 2 local, 3 remote | | `legacy/git/exec_conflict.go:49-64` |
| The resolver: three columns; Space marks the column of a file; Enter needs a mark on every file, writes each chosen version, runs `add <path>`, and finalizes by the operation; Escape aborts by the operation | | `legacy/app/conflict_handlers.go:92-289`; `legacy/app/app_keys.go:100-108` |
| The column labels: `BASE`; `%s (current)`, `%s (incoming)`; `OURS (rebased)`, `THEIRS (incoming)`; `%s (merged)`, `stashed changes` | | `legacy/app/handlers_conflict.go:143-220` |
| A version that does not exist: `[FILE DELETED IN THIS VERSION]` and two more lines | | `legacy/app/handlers_conflict.go:116-117` |
| Time travel, enter: the dialog `time_travel`, first button selected; the marker with the branch; `checkout <hash>`; with a dirty tree `stash push -u -m <label>` first | | `legacy/app/handlers_history_cache.go:83-120`; `legacy/app/confirm_time_travel.go:82-242`; `legacy/git/exec_timetravel_checkout.go:29` |
| The time travel menu: `time_travel_history` (`h`), `time_travel_files_history` (`g`), `time_travel_return` (`r`) | | `legacy/app/menu_items.go:147-178`; `legacy/app/menu_render_extra.go:63-115` |
| Return, clean tree: the dialog with the title `Return to %s`, second button selected; `checkout <branch>`; then the stash of the entry: `stash apply`, `stash drop` | | `legacy/app/dispatch_git_timetravel.go:82-100`; `legacy/git/exec_timetravel_return.go:32-113` |
| Return, dirty tree: the dialog `Return to %s with uncommitted changes`, `Merge changes` / `Discard changes`, second button selected. Merge: a stash, `checkout <branch>`, `merge <hash>`, `stash apply`, `stash drop`. Discard: `reset --hard HEAD`, `clean -fd`, then the return | | `legacy/app/dispatch_git_timetravel.go:66-81`; `legacy/app/confirm_time_travel_return.go:144-232`; `legacy/git/exec_timetravel_merge.go:30-132`; `legacy/app/confirm_dialog_render.go:57-70` |
| A detached HEAD that TIT did not make: the return opens the branch picker; with a dirty tree the same dialog with `Stash changes` | | `legacy/app/dispatch_navigation.go:12-55`; `legacy/app/handlers_config_branch.go:52-64` |
| The header of a time travel: `Viewing commit from ` and the date as `Jan 2, 2006`, from `log -1 --format=%aI <hash>` | | `legacy/app/app_view_header.go:57`; `legacy/app/confirm_time_travel.go:102` |
| The header of that detached HEAD: `Not on a branch. Select 'Return to branch' to continue.` | | `legacy/app/app_view_header.go:98` |
| The branch list: `for-each-ref --sort=-committerdate refs/heads` with nine fields between tabs; the current branch first | | `legacy/git/branch.go:26-108` |
| A branch row: `● ` when current; `↑%d ↓%d`, `synced`, or `local` | | `legacy/ui/branchpicker.go:74-95` |
| The details of a branch: eleven lines | | `legacy/ui/branchpicker.go:141-173` |
| The keys of the picker: up, `k`, down, `j`, Enter, `a`, `m`, `x` | | `legacy/app/app_keys.go:128-137` |
| Switch: `switch <name>`; with a dirty tree the dialog `Switch to %s with uncommitted changes?`, `Stash changes` / `Discard changes`, second button selected | | `legacy/app/handlers_git_branch.go:25`; `legacy/app/handlers_config_branch.go:91-104` |
| Add: the prompt `New branch name:`; `check-ref-format --branch <name>`; `checkout -b <name>` | | `legacy/app/handlers_config_branch.go:115-125`; `legacy/app/handlers_git_branch.go:66,100` |
| Merge from: the dialog `merge_branch`, or `merge_branch_dirty`; second button selected; `merge <branch>` | | `legacy/app/confirm_merge.go:18-150` |
| Delete: the dialog `Delete branch %s?`, second button selected; `branch -D <name>` | | `legacy/app/handlers_config_branch.go:157-169`; `legacy/app/confirm_dialog_handlers.go:229` |
| The footer hints of the resolver and of the picker | | `legacy/app/messages_menu.go:111-122,151-166` |
| Pull with rebase: `pull --rebase --progress` | | `legacy/app/handlers_git_pull.go:51` |

Legacy has no interactive rebase, no rebase onto a branch, and no pin. Its only rebase commands are `rebase --continue` and `rebase --abort` (`legacy/app/op_rebase.go:14,19`).

### Legacy faults that this part corrects

Ruling 4: "kept the feature FIX THE DESIGN, DESIGN BY CONTRACT."

1. **An interrupted dirty operation had no way out.** The `DirtyOperation` menu showed only `Abort merge`, which runs `merge --abort` and `reset --hard`. It reads no marker, restores no stash, and removes no marker, and the detection gives `DirtyOperation` while the marker exists (`legacy/app/menu_render_extra.go:141-147`; `legacy/app/op_merge.go:93-123`; `legacy/git/state.go:41-45`). SPEC rules a resume and an abort (`tit/SPEC.md:167-171`).
2. **A failed step removed the marker and left the stash** (`legacy/app/handlers_git_result.go:13-33`).
3. **`stash apply` and `stash drop` named no stash.** They act on the newest stash, which the user can have made (`legacy/app/op_dirty_pull_snapshot.go:234-240`).
4. **The stash of a time travel was known only to a list file in the user folder** (`legacy/config/stash.go:36-42`). The marker never held it (`legacy/app/confirm_time_travel.go:85`; `legacy/git/exec_timetravel_checkout.go:41`). Ruling 17 keeps the new TIT away from that file.
5. **A second stash of a time travel made the program panic:** `AddStashEntry` panics on a second entry of one repository (`legacy/config/stash.go:92-114`; `legacy/app/confirm_time_travel_return.go:228`).
6. **A resolved conflict of a time travel merge never applied the stash** (`legacy/app/op_time_travel.go:15-79`).
7. **Escape in the resolver of a time travel conflict left the conflict** and returned to the menu (`legacy/app/conflict_handlers.go:259-288`).
8. **A version that does not exist was written into the file as the placeholder text** (`legacy/app/conflict_handlers.go:143-145`).
9. **Enter wrote the files one by one and stopped at an error,** with the earlier files already written (`legacy/app/conflict_handlers.go:138-156`).
10. **The footer of the resolver said `Esc back`** where Escape aborts the operation (`legacy/app/messages_menu.go:118-122`).
11. **The picker bound no Tab,** where SPEC and one footer text name it (`legacy/app/app_keys.go:128-137`; `tit/SPEC.md:428`).
12. **A rebase could open an editor on the terminal of TIT.** On macOS a child gets the terminal of TIT (tracker, "The git side: child processes"), and `rebase --continue` runs the commit editor.

### The facts of the tool

| Fact | Source |
|---|---|
| `stash apply`: "`<stash>` may be any commit that looks like a commit created by `stash push` or `stash create`". `stash drop` takes "A reference of the form `stash@{<revision>}`". | `git-stash.adoc` |
| `apply` exits "with status 1 when applying the stash entry resulted in conflicts, in which case the stash entry is left in place" | `git-stash.adoc` |
| `stash push -u`: "all untracked files are also stashed and then cleaned up" | `git-stash.adoc` |
| The stash list is the reflog of `refs/stash`: "older stashes are found in the reflog of this reference" | `git-stash.adoc` |
| With no stash, `log -g … refs/stash` fails: "ambiguous argument 'refs/stash': unknown revision" | one run in `tit/` and one in `jam/`, this session |
| A merge with no terminal opens no editor: the editor runs only when `isatty(0) && isatty(1)` | `builtin/merge.c`, `default_edit_option` |
| The order of the editors: "`$GIT_EDITOR`, then `core.editor` configuration value, then `$VISUAL`, then `$EDITOR`" | `git-var.adoc` |
| `sequence.editor`: "The value is meant to be interpreted by the shell when it is used." The tool appends the path of the list file, as an absolute path. | `config/sequence.adoc`; `editor.c`, `launch_specified_editor` |
| The list of an interactive rebase: "The oneline descriptions are purely for your pleasure; 'git rebase' will not look at them but at the commit names" | `git-rebase.adoc`, INTERACTIVE MODE |
| `drop`: "To drop a commit, replace the command \"pick\" with \"drop\", or just delete the matching line." | `git-rebase.adoc` |
| `squash`, `fixup`: "an editor is opened allowing you to edit the message"; for a `fixup` the message of the folded commit is dropped | `git-rebase.adoc` |
| `exec`: "launches the command in a shell ... The command is run from the root of the working tree." "The interactive rebase will stop when a command fails" | `git-rebase.adoc` |
| `reword` always opens the editor | `sequencer.c` |
| A rebase "will stop at the first problematic commit and leave conflict markers" | `git-rebase.adoc` |
| `checkout-index --stage=<number>`: "Instead of checking out unmerged entries, copy out the files from the named stage. <number> must be between 1 and 3." | `git-checkout-index.adoc` |
| The stages: "stage 1 is the common ancestor, stage 2 is the target branch's version", "stage 3 is the version from the branch which is being merged". During a rebase the two sides "may appear swapped". | `gitrevisions.adoc`; `git-checkout.adoc` |
| The unmerged record: `u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>` | `git-status.adoc` |
| `for-each-ref`: "`%xx` - where `xx` are hex digits - renders as the character with hex code `xx`"; `%(HEAD)` is "`*` if `HEAD` matches current ref"; `%(upstream:track)` shows "[ahead N, behind M]"; "Append `:track,nobracket` to show tracking information without brackets (i.e \"ahead N, behind M\")" | `git-for-each-ref.adoc`; the manual page of this machine, `git-for-each-ref.1:269-271` |
| `check-ref-format --branch`: "checks if it can be used as a valid branch name"; it "exits with a non-zero status if it is not" | `git-check-ref-format.adoc` |
| `branch -D`: "Shortcut for `--delete --force`" | `git-branch.adoc` |
| `rev-list --count`: "Print a number stating how many commits would have been listed" | `rev-list-options.adoc` |

### The JAM and JUCE pieces

| Piece | Where |
|---|---|
| `jam::ChildProcess::getArguments`: "A double-quoted or single-quoted group holds spaces"; on Windows `CommandLineToArgvW` splits the line. So one argument of a command line can hold a space when it is in double quotes. | `jam_subprocess/subprocess/jam_ChildProcess.h:61-74` |
| `run (commandLines, fallbackCommandLines)`, `cancel()`, the detection after a run, the pool job form, `getOutput`, `getOutputText`, `getArguments (commandLine, argument)` | Part 9 Step 5; Part 10 Step 10; Part 8 Step 7; Part 11 Step 7 |
| The menu of a tree of records: `getRecordMenu`, the tables `records` and `selections`, `isRecordMode`, `setSelection`, and `keyPressed`, which makes the highlighted record the selected record before a command runs | Part 11 Step 11, item 5 |
| The dialog, the input editor, the console, the footer message, `setConfirmation`, `InputView::setDocument` and `getText`, `ConsoleView`, `TITRunProcessor`, the lane of the records, `HistoryView`, `cancellations`, `getFocusComponent`, `getMenuCommands`, `setMenu`, `menuKeys`, the command families | Part 10 Steps 14 and 16; Part 11 Steps 8, 10, and 11 |
| The process environment is set once: `setenv`, `_putenv_s` | Part 9 Step 8 |
| `juce::ValueTree::moveChild`, `getChildWithProperty`, `indexOf`; `juce::File::replaceWithText`, `deleteRecursively`, `createDirectory`; `juce::String::replaceFirstOccurrenceOf`, `quoted`, `containsOnly`, `fromFirstOccurrenceOf`, `compare`; `juce::StringArray::joinIntoString (separator, startIndex)`, `addArray`, `insert`; `juce::Array::operator[]` | doxygen `classjuce_1_1ValueTree`, `classjuce_1_1File`, `classjuce_1_1String`, `classjuce_1_1StringArray`, `classjuce_1_1Array` |

### Words the framework and the earlier parts already hold

| Word | Where |
|---|---|
| `Id::id`, `Id::value`, `Id::text`, `Id::mode`, `Id::message`, `Id::directory` | `jam/cast/core.md`, `jam/cast/gui.md` |
| The rows of the `git` child, `map::Operation`, `files::dirtyOperationMarker`, `files::timeTravelMarker`, `files::pinMarker`, `branchLine`, `pinnedHashLine` | Part 8 |
| `Id::selectedIndex`, `Id::commandLines`, `Id::attribute`, `Id::commit`, `Id::file`, `Id::focusedPane`, `Id::isCacheReady`, `map::Pane` | Parts 10 and 11 |
| `text::Git::resetHard`, `resetHardHead`, `clean`, `pullMerge`, `commitNoEdit`, `mergeAbort`, `rebaseContinue`, `rebaseAbort`, `upstreamName`, `outgoingSuffix`, `FORMAT` of the history records; `text::Footer::input`, `text::Footer::commitMessageEmpty` | Parts 9, 10, 11 |
| `text::Reading::hashPlaceholder` (`%s`), `text::Reading::dateFormat` | Part 11 Step 3 |

### Closed by a read

ARCHITECT can reverse each one.

1. **The state of a protocol is in the repository, and the detection reads it.** SPEC: "Git State → Allowed Actions → Menu" (`tit/SPEC.md:23`); "State always reflects actual Git state" (invariant 10). A chain that stops leaves a stash, a marker, a merge, a rebase, or unmerged files. Each is a row of the `git` child after the next detection. So a resume is a command that the rows make active, and the processor of Part 9 does not change.
2. **Every stash of TIT has one label, `TIT DIRTY-OP SNAPSHOT`** (`tit/SPEC.md:354`). The row `titStash` holds the reference of the newest stash with that label, as the tool names it (`stash@{0}`). A command names that reference, so it never touches a stash of the user (legacy fault 3).
3. **`DirtyOperation` is: HEAD is on a branch, no conflict, no merge, no rebase, and a stash of TIT exists.** This replaces the test of Part 8, "the marker exists". The marker holds the data of the abort only: the branch and the HEAD of the start (`tit/SPEC.md:350-357,376-385`). A dirty marker with no stash of TIT is stale, and the detection removes it, as it removes a stale pin marker (Part 8 Step 7, item 7).
4. **The `DirtyOperation` menu has two rows.** "Reapply changes" applies the stash and drops it: the resume of `tit/SPEC.md:170`. "Abort and restore" is the abort of `tit/SPEC.md:376-385`; it is active when the marker exists. Their texts are new (ruling 41).
5. **The type of a conflict comes from the rows.** A merge is in progress: the finalize is `commit --no-edit`. A rebase is in progress: `rebase --continue`. Neither: the conflict came from `stash apply`, and the finalize is the drop of the stash. This is `tit/SPEC.md:151`, "depending on underlying operation".
6. **The resolver opens when the operation is `Conflicted`.** Legacy opens it after each conflict and at the start (`legacy/app/app_constructor.go:150-200`). The mode of the `Conflicted` operation is the resolver, not a menu. The mode `menuConflicted` of Part 9 gets no key row and no menu.
7. **The resolver is one menu of the files and three text panes.** Legacy drew three lists of the same files, one for each column (`legacy/ui/conflictresolver.go:50-142`). One menu with the chosen version as the attribute of a file holds the same data (ruling 60). Tab goes through the menu and the three panes. Space on a pane chooses that version for the selected file.
8. **Escape in the resolver aborts the operation,** as in legacy, for every type of conflict (legacy faults 7 and 10). The footer says `Esc abort`.
9. **A chosen version is written by the tool:** `checkout-index -f --stage=<n> -- <path>`, then `add -- <path>`. A version that does not exist is the removal of the file: `rm -f -- <path>` (legacy fault 8). The lines of all files are one chain, so a failure writes nothing more and shows the text of the tool (legacy fault 9).
10. **The label of the incoming side of a merge is the name that the tool gives to `MERGE_HEAD`:** `name-rev --name-only MERGE_HEAD`. So the labels are `<branch> (current)` and `<name> (incoming)` for every merge, also for a pull (Part 2, "What the shipped code has", row "The conflict resolver labels name the branches"). The legacy labels `LOCAL (yours)` and `REMOTE (theirs)` of a pull are not carried.
11. **Time travel with a dirty tree saves the changes with no question,** as legacy does after the dialog `time_travel` (`legacy/app/confirm_time_travel.go:137-242`). The return applies that stash.
12. **The clean return runs `checkout <branch>` and applies the stash of the entry. The dirty return has the legacy choice:** `Merge changes` or `Discard changes`. SPEC names three buttons, "Yes=merge-back / Discard=return-no-merge / Cancel" (`tit/SPEC.md:215`); the dialog of Part 10 has two buttons and Escape, as the legacy dialog has (`legacy/app/dispatch_git_timetravel.go:66-81`).
13. **The menu item `time_travel_merge` is not carried, because no legacy menu shows it.** Legacy defines the item and its dispatcher (`legacy/app/menu_items.go:163`; `legacy/app/dispatch_git_timetravel.go:14-41`), but the time travel menu adds only the history items and one return item: "Add single return option (handles both merge and discard via dialog when dirty)" (`legacy/app/menu_render_extra.go:105-113`). So the user never reaches the item or its two dialogs, `time_travel_merge` and `time_travel_merge_dirty`. The merge is the first button of the dirty return, which this part builds. The rows `time_travel_return_dirty` and `branch_switch_clean` of the legacy text table have no dialog that shows them (`legacy/app/messages_dialog.go:48,84`): the dirty return builds its dialog inline (`dispatch_git_timetravel.go:66-81`). Ruling 3 keeps every feature that the user can reach, and each of those is built.
14. **The `Stash Not Found` dialog, and the record of the stash of a time travel.** Legacy keeps the hash of the stash of the entry in a list file of the user folder (`legacy/config/stash.go:36-42`). At the return, when that stash is gone, legacy asks: "Original stash %s was manually dropped. Continue without restoring stash?", with `Continue` and `Cancel` (`legacy/app/confirm_time_travel_return.go:25-45,90-104`). Ruling 3 keeps this. In this part the record is in the repository, as every state of a protocol is (item 1): the second line of the time travel marker holds the label of the stash of the entry, or is empty when the entry made no stash. The detection reads it into the row `hasOriginalStash`. A stash of TIT is known by its label (item 2), so the return finds the stash gone when `hasOriginalStash` is true and the row `titStash` is empty. The dialog then shows, and its text names the stash by its label, because the new TIT keeps no hash of it. `Continue` runs the return of the tree as it is: the return lines when the tree is clean, and the merge lines when it is dirty (legacy `confirm_stale_stash_continue` and `confirm_stale_stash_merge_continue`).
15. **A time travel that the program finds at its start stays a time travel.** The menu of the state gives the return. Legacy has a restore at the start for one case: the marker exists, and HEAD is on a branch or the marker is not valid (`legacy/app/app_constructor.go:131-146,230-231`; `legacy/app/app_update_msg.go:70-78`). That restore runs `reset --hard HEAD`, `clean -fd`, the checkout of the first branch, and the apply and drop of the first stash (`legacy/app/app_init.go:12-98`). This part does not carry it (ruling 54): the program discards no change of the user at its start. The same case here is a stale marker, which the detection removes (Part 8 Step 7, item 7), and a stash of TIT that remains gives the `DirtyOperation` menu with "Reapply changes" (Step 12).
16. **The branch picker opens from "Return" on a detached HEAD that TIT did not make.** Its other opener is the Config menu, which is Part 13 (tracker, row 13). The command `branchPicker` of this part is what Part 13 binds.
17. **Enter on the current branch does nothing.** Legacy returns to the Config menu there (`legacy/app/handlers_config_branch.go:36-112`), and the Config menu is Part 13.
18. **A new branch name is typed in the input editor of Part 10,** in its own mode, `branchName`, because Enter must run another command. The checks of the name are the tool's: `check-ref-format --branch`, then `checkout -b`. A name that the tool refuses shows the text of the tool in the console (Part 11, Closed by a read, item 15).
19. **The two rebase commands of rulings 20 and 22 write the ruled footer text.** Rulings 20 and 22 say "inactive, and the footer says why". The footer can say why only when the key arrives. So the key rows of `branchPickerRebase` and of `rebaseInteractive` stay in the sheet, and each callback writes one Model row: the ruled text in `message`, or a dialog, or the mode of the list. The callback that writes the text runs no command line. The footer listens to `message` (Part 10 Step 13). This is the path of `inputSubmit` with an empty text (Part 10 Step 16, item 7).
20. **The interactive rebase runs with a list file that TIT writes.** The folder `.git/TIT_REBASE` holds the file `todo` and one message file for each reworded commit. The option `-c sequence.editor=…` copies `todo` over the list of the tool. The folder is replaced whole at each apply.
21. **A reword is a `pick` and an `exec` line** that amends the commit with the message file, because `reword` always opens the editor. The message is typed in the input editor of Part 10 when the user presses `r`, with the old message as its start (Part 2 Step 7: "Reword opens the commit message editor").
22. **`GIT_EDITOR` is `true` for the process of TIT,** in the form of `GIT_TERMINAL_PROMPT` of Part 9. So `rebase --continue` and a `squash` take the prepared message, and no editor opens on the terminal of TIT (legacy fault 12).
23. **In the rebase list, the key of the action that a row has sets the row back to `pick`.** Ruling 10 names the edits and no way back from one.
24. **The rebase list shows the oldest commit first,** as the list of the tool does. `J` moves a row down: later in time.
25. **The base of the interactive rebase is `--keep-base @{u}`, or `--root` with no upstream** (ruling 9: "A branch with no remote can rewrite all its commits"). `--keep-base`: "Set the starting point at which to create the new commits to the merge base of `<upstream>` and `<branch>`" (`git-rebase.adoc`; the manual page of this machine, `git-rebase.1:365-370`). So a rewrite of a `Diverged` branch does not also replay the commits on the upstream.
26. **The pushed commits of a rebase onto a branch are counted at the creation of the branch rows:** the commits that both HEAD and the upstream hold and the selected branch does not. A count above 0 gives the text of ruling 22.
27. **Pin with a dirty tree shows the dialog `pin_dirty` only.** Its two buttons are the two paths of the protocol, and Escape cancels.
28. **While `Pinned`, "Discard all changes" shows the dialog with no remote choice.** Ruling 1: "the menu hides Pull and Replace local".
29. **A marker is value data of its dialog, and it is written when a button with command lines runs.** So a dialog that the user cancels writes nothing. A marker of a chain that fails is stale, and the next detection removes it: no stash, HEAD on a branch, or HEAD not at the pinned hash.
30. **The header date is `%b %d, %Y`.** Legacy shows the day with no leading zero (`Jan 2, 2006`); `juce::Time::formatted` has no such form on both platforms.
31. **The branch name and the hash go into a command line as they are.** A branch name holds no space (`check-ref-format`). A path of the resolver goes in double quotes. A path or a name that holds a double quote makes the tool refuse the line, and the console shows it.
32. **Not in this part:** the Config menu and `add_remote` with no remote, `lfs checkout` after a hard reset, and the auto-update (Part 13).

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: JAM, the modes
**Scope:** `jam/cast/gui.md`, table `## EditorMode`; then `jam/generated/` through `cast cast/spell.md` from the JAM root.
**Action:**
1. After the `search` row of Part 11 Step 1, add `conflictResolve` 20, `branchPicker` 21, `branchName` 22, `rebaseList` 23, `rebaseReword` 24.
2. Run `cast cast/spell.md` from the JAM root.

**Validation:**
- `map::EditorMode` holds the five names with the keys 20 to 24. The rows 0 to 19 keep their keys. No JAM source file changes.
- The names: `conflictResolve` and `branchPicker` are the legacy modes `ModeConflictResolve` and `ModeBranchPicker` (`legacy/app/modes.go:21-40`). `branchName` is the legacy prompt `New branch name:`. `rebaseList` is the view name of ruling 21, "Rebase list". `rebaseReword` is the edit of ruling 10.
- This part adds no colour id. Its views are the stock menu and `jam::TextEditor`, which have their colours (Part 10 Step 8; Part 11 Step 5).

### Step 2: TIT, the words, the command ids, the files, the command lines
**Scope:** `tit/cast/identifiers.md`, `tit/cast/spell.md`, `tit/cast/files.md`, `tit/cast/git.md`.
**Action:**
1. `cast/identifiers.md`, table `## identifiers`: after the last row of Part 11 Step 2, add these rows of type `@id` with an empty description. The value of each row is its name in backticks.

   | Name | Family |
   |---|---|
   | `titStash` | the stash of the label `TIT DIRTY-OP SNAPSHOT` (`tit/SPEC.md:354`) |
   | `originalBranch`, `originalHead` | SPEC `ORIGINAL_BRANCH`, `ORIGINAL_HEAD` (`tit/SPEC.md:352-353`) |
   | `pinOriginalHead` | the second line of the pin marker, "original HEAD" (Part 2 Step 3, item 4) |
   | `headHash` | `currentHash` (Part 8 Step 2); the cast row `head` (Part 8 Step 3) |
   | `isMerging`, `isRebasing` | the Operation words `Merging`, `Rebasing` (`tit/SPEC.md:74-84`), with the `is` prefix |
   | `mergeName` | the file `MERGE_HEAD`; `name-rev` |
   | `commitDate` | legacy `CommitTime` of the time travel info (`legacy/app/app_view_header.go:57`) |
   | `conflict` | the legacy state `ConflictResolveState` (`legacy/app/conflict_state.go:5-22`) |
   | `version` | legacy `Versions` (`legacy/ui/conflictresolver.go:114-115`) |
   | `branches`, `branch` | legacy `BranchInfo` (`legacy/ui/branchpicker.go:13-23`) |
   | `isCurrent` | legacy field `IsCurrent` |
   | `pushedCount` | `modifiedCount` (Part 8 Step 2); ruling 20, "pushed commits" |
   | `rebase` | ruling 21, "Rebase list" |
   | `hasOriginalStash` | `hasUpstream` (Part 8 Step 2); `originalBranch`; the legacy text "Original stash" (`legacy/app/confirm_time_travel_return.go:39`) |
   | `base`, `ours`, `theirs` | the rows of `map::ConflictPane` (item 2); the types of the run chains of the three versions (Step 7) |
   | `timeTravelHistory`, `timeTravelFilesHistory` | legacy ids `time_travel_history`, `time_travel_files_history` (`legacy/app/menu_items.go:147-162`) |

2. `cast/identifiers.md`: after the table `## Pane` of Part 11 Step 2, add the table `## ConflictPane` in the same form, with the brief `@brief The panes of the conflict resolver.` The rows: `files` 0, `base` 1, `ours` 2, `theirs` 3. Add its output group to `cast/spell.md` in the form of Part 8 Step 4, item 3.
3. `cast/identifiers.md`, table `## CommandId`: after the `rewind` row, add these rows. The value of each row is its name in backticks.

   | name | key | Source of the name |
   |---|---|---|
   | `dirtyPullMerge` | `0x4a00022` | `dirty_pull_merge` (`legacy/app/menu_items.go:74`) |
   | `dirtyContinue` | `0x4a00023` | `rebaseContinue` (Part 9); the Operation `DirtyOperation` |
   | `dirtyAbort` | `0x4a00024` | `rebaseAbort` (Part 9) |
   | `timeTravel` | `0x4a00025` | `time_travel` (`legacy/app/confirm_dialog.go:19`) |
   | `timeTravelReturn` | `0x4a00026` | `time_travel_return` (`legacy/app/menu_items.go:171`) |
   | `pin` | `0x4a00027` | ruling 1 |
   | `unpin` | `0x4a00028` | ruling 1, "Unpin" |
   | `addRemote` | `0x4a00029` | `add_remote` (`legacy/app/menu_items.go:137`); ruling 35 |
   | `pullRebase` | `0x4a0002a` | `pull_rebase` (`legacy/app/handlers_git_pull.go:56`) |
   | `conflictSpace` | `0x4a0002b` | `handleConflictSpace` (`legacy/app/conflict_handlers.go:92`) |
   | `conflictEnter` | `0x4a0002c` | `handleConflictEnter` (`:119`) |
   | `branchPicker` | `0x4a0002d` | `config_branch`, "Open branch picker" (`legacy/app/menu_items.go`) |
   | `branchPickerEnter` | `0x4a0002e` | `handleBranchPickerEnter` (`legacy/app/handlers_config_branch.go:36`) |
   | `branchPickerAdd` | `0x4a0002f` | `handleBranchPickerAdd` (`:115`) |
   | `branchPickerMerge` | `0x4a00030` | `handleBranchPickerMerge` (`:128`) |
   | `branchPickerDelete` | `0x4a00031` | `handleBranchPickerDelete` (`:147`) |
   | `branchPickerRebase` | `0x4a00032` | ruling 2, "Rebase onto branch" |
   | `newBranchNameSubmit` | `0x4a00033` | `handleNewBranchNameSubmit` (`legacy/app/handlers_git_branch.go:56`) |
   | `rebaseInteractive` | `0x4a00034` | the dialog key `rebase_interactive` (Part 2) |
   | `rebaseSquash` | `0x4a00035` | ruling 21, squash |
   | `rebaseFixup` | `0x4a00036` | ruling 21, fixup |
   | `rebaseReword` | `0x4a00037` | ruling 21, reword |
   | `rebaseDrop` | `0x4a00038` | ruling 21, drop |
   | `rebaseMoveDown` | `0x4a00039` | ruling 21, move down |
   | `rebaseMoveUp` | `0x4a0003a` | ruling 21, move up |
   | `rebaseApply` | `0x4a0003b` | ruling 21, apply |
   | `rewordSubmit` | `0x4a0003c` | `inputSubmit` (Part 10) |

4. `cast/files.md`, table `## files`: after the last row of Part 10 Step 5, add two rows of type `@string`: `rebaseDirectory`, value `` `TIT_REBASE` ``, description `The folder in the repository data folder that holds the list and the messages of an interactive rebase.`; `rebaseTodo`, value `` `todo` ``, description `The list file of an interactive rebase.`
5. `cast/git.md`: after the last row of Part 11 Step 2, add these rows. `BRANCHFORMAT` stands for `--format=%(refname:short)%09%(HEAD)%09%(objectname:short)%09%(subject)%09%(authorname)%09%(committerdate:rfc)%09%(upstream:short)%09%(upstream:track,nobracket)`, and `REBASEFORMAT` for `--format=%x1e%H%x1f%aI%x1f%h%x20%s%x1f%B%x1d`, both written in full.

   | name | value | description |
   |---|---|---|
   | `stashLabel` | `` `TIT DIRTY-OP SNAPSHOT` `` | The message of every stash that titc makes. |
   | `stashPush` | `` `git stash push -u -m` `` | Saves the changes and the untracked files; the message follows. |
   | `stashApply` | `` `git stash apply` `` | Applies a stash and keeps it. |
   | `stashDrop` | `` `git stash drop` `` | Removes a stash. |
   | `stashLog` | `` `git log -g --first-parent -n 1 --format=%gd refs/stash` `` | Prints the reference of the newest stash that the next option selects. |
   | `grepOption` | `` `--grep=` `` | The start of the option that selects by message. |
   | `checkout` | `` `git checkout` `` | Checks out the next commit or branch. |
   | `switchBranch` | `` `git switch` `` | Switches to the next branch. |
   | `merge` | `` `git merge` `` | Merges the next commit or branch. |
   | `mergeName` | `` `git name-rev --name-only MERGE_HEAD` `` | Prints a name of the commit that a merge brings in. |
   | `commitDate` | `` `git log -1 --format=%aI` `` | Prints the author date of HEAD. |
   | `showStage` | `` `git show` `` | Prints one version of an unmerged file; the stage and the path follow as one argument. |
   | `checkoutStage` | `` `git checkout-index -f --stage=` `` | Writes one version of an unmerged file; the stage number follows with no space. |

   | `add` | `` `git add --` `` | Stages the next path. |
   | `remove` | `` `git rm -f --` `` | Removes the next path from the tree and the index. |
   | `zeroHash` | `` `0` `` | The character of the hash of a version that does not exist. |
   | `branches` | `` `git for-each-ref --sort=-committerdate refs/heads BRANCHFORMAT` `` | Prints one record for each local branch. |
   | `mergeBase` | `` `git merge-base HEAD @{u}` `` | Prints the newest commit that HEAD and its upstream share. |
   | `countCommits` | `` `git rev-list --count` `` | Prints the count of the commits of the next range. |
   | `rangeSeparator` | `` `..` `` | The text between the two ends of a range. |
   | `checkBranchName` | `` `git check-ref-format --branch` `` | Fails when the next name cannot be a branch name. |
   | `createBranch` | `` `git checkout -b` `` | Creates the next branch and switches to it. |
   | `deleteBranch` | `` `git branch -D` `` | Deletes the next branch. |
   | `rebase` | `` `git rebase` `` | Replays the commits of the branch on the next branch. |
   | `pullRebase` | `` `git pull --rebase --progress` `` | Fetches the upstream and replays the local commits on it. |
   | `rebaseLog` | `` `git log --reverse REBASEFORMAT` `` | The commits of a range, oldest first, with the message as the last field. |
   | `rebaseInteractive` | `` `git -c "sequence.editor=cp .git/TIT_REBASE/todo" rebase -i` `` | Runs a rebase with the list file of titc; the base follows. |
   | `rootOption` | `` `--root` `` | The base of a rebase of every commit of the branch. |
   | `keepBaseUpstream` | `` `--keep-base @{u}` `` | The base of a rebase that keeps the commit that HEAD and its upstream share. |
   | `headName` | `` `HEAD` `` | The name of the current commit. |
   | `amendFile` | `` `exec git commit --amend -F .git/TIT_REBASE/` `` | The line of the list that gives a commit the message of the next file. |
   | `editorVariable` | `` `GIT_EDITOR` `` | The environment variable of the editor of a commit message. |
   | `editorValue` | `` `true` `` | The value of editorVariable: a command that changes nothing and succeeds. |

   - `stashLog` with the argument `--grep=TIT DIRTY-OP SNAPSHOT` prints `stash@{<n>}` of the newest stash of TIT, and nothing when none exists. `%gd` is the reflog selector (`pretty-formats.adoc`). The detection gives the argument whole, with `getArguments (commandLine, argument)` of Part 8.
   - The fields of a branch record, in order: the name, `*` for the current branch, the short hash, the subject, the author, the date, the upstream, the tracking text.
   - The records of `rebaseLog` have the three fields of Part 11 Step 2, item 4, and the message as the fourth field. `getRecords` and `getFields` of Part 11 read them. `pathOption` is a row of Part 11.
   - `rebaseInteractive` holds one argument in double quotes (inventory, JAM pieces, first row). The tool runs `cp .git/TIT_REBASE/todo <list file>` in a shell. The child runs in the repository folder (Part 8 Step 8), so the relative path names the file that Step 12 writes.
   - `amendFile` is a line of a list file, not a command line of TIT. The tool runs it "from the root of the working tree".

**Validation:**
- Each command line equals its legacy literal or its documented form (inventory). Each new word has its cited family.
- No command line holds a space inside one argument outside double quotes.
- The map `commands` of Steps 10 to 12 holds each new command id. The map `availability` holds each new command that a state rule governs (ruling 56). A command that acts on a selected row acts only when the row has the data that it needs (ruling 40).

### Step 3: TIT, the texts
**Scope:** `tit/cast/text.md`, `tit/cast/spell.md`.
**Action:**
1. `cast/text.md`: after the table `## reading` of Part 11 Step 3, add three tables in the same form: `## conflict` (`@brief Texts of the conflict resolver.`), `## branch` (`@brief Texts of the branch picker.`), `## rebase` (`@brief The words of the list of an interactive rebase.`). Add their output groups to `cast/spell.md` in the form of Part 10 Step 6, item 5, with the names `Conflict`, `Branch`, `Rebase`. No value starts or ends with a space.
2. `## conflict` (`legacy/app/handlers_conflict.go:116-117,143-220`; `legacy/ui/conflictresolver.go:153-157`):

   | name | value |
   |---|---|
   | `base` | `BASE` |
   | `current` | `%s (current)` |
   | `incoming` | `%s (incoming)` |
   | `ours` | `OURS (rebased)` |
   | `theirs` | `THEIRS (incoming)` |
   | `merged` | `%s (merged)` |
   | `stashed` | `stashed changes` |
   | `unmarked` | `[ ]` |
   | `deleted` | `[FILE DELETED IN THIS VERSION]\n\nThis file was deleted in %s.\nThe conflict occurred because the other side modified it.` |

3. `## branch` (`legacy/ui/branchpicker.go:74-95,141-173`; `legacy/app/handlers_config_branch.go:120`):

   | name | value | name | value |
   |---|---|---|---|
   | `currentMark` | `●` | `synced` | `synced` |
   | `local` | `local` | `heading` | `BRANCH` |
   | `name` | `Name: %s` | `statusCurrent` | `Status: ● Current` |
   | `statusOther` | `Status: Not current` | `upstream` | `Upstream: %s (%s)` |
   | `upstreamNone` | `Upstream: none (local only)` | `tipHeading` | `TIP COMMIT` |
   | `hash` | `Hash: %s` | `author` | `Author: %s` |
   | `date` | `Date: %s` | `newBranchName` | `New branch name:` |

4. `## rebase`: `pick` `pick`, `squash` `squash`, `fixup` `fixup`, `reword` `reword`, `drop` `drop` (`git-rebase.adoc`, INTERACTIVE MODE).
5. `## header`: after the last row of Part 10, add `viewingCommit`, value `Viewing commit from`; `commitDateFormat`, value `%b %d, %Y`; `notOnBranch`, value `Not on a branch. Select 'Return to branch' to continue.` (`legacy/app/app_view_header.go:57,98`).
6. `## footer`: after the last row of Part 11, add:

   | name | value | Source |
   |---|---|---|
   | `marked` | `Marked: %s → %s` | `legacy/app/messages_state.go:103` |
   | `conflictFiles` | `↑↓ navigate  │  Tab versions  │  Enter resolve  │  Esc abort` | ruling 41 |
   | `conflictVersion` | `↑↓ scroll  │  Space choose  │  Tab next  │  Enter resolve  │  Esc abort` | ruling 41 |
   | `branchPickerCurrent` | `↑↓ navigate  │  a add  │  Tab details  │  Esc cancel` | `legacy/app/messages_menu.go:152-157`; ruling 41 |
   | `branchPickerOther` | `↑↓ navigate  │  a add  │  m merge from  │  r rebase onto  │  x delete  │  Enter switch  │  Esc cancel` | `legacy/app/messages_menu.go:159-166`; ruling 41 |
   | `rebaseList` | `s squash  │  f fixup  │  r reword  │  d drop  │  J K move  │  Enter apply  │  Esc cancel` | ruling 41 |
   | `pushedCommits` | `Pushed commits cannot be rebased` | Part 2, "Footer text when a command is inactive" |
   | `noUnpushedCommits` | `No unpushed commits to rewrite` | the same table |
   | `commitFirst` | `Commit or discard changes first` | the same table |
   | `branchNameEmpty` | `Branch name cannot be empty` | `legacy/app/messages_error.go:7` |

7. `## confirmation`: for each dialog below, four rows named with the prefix and the suffixes `Title`, `Explanation`, `Yes`, `No`, in the form of Part 10 Step 6, item 4. A legacy value is the Go literal of its field. A value of Part 2 is the text of its table "Dialogs", with each `<br>` as `\n` and each `{name}` as `%s`.

   | Prefix | Source |
   |---|---|
   | `dirtyPull` | `legacy/app/messages_dialog.go:30-35` |
   | `timeTravel` | `legacy/app/messages_dialog.go:60-65` |
   | `timeTravelReturn` | `legacy/app/messages_dialog.go:66-71`, with the title `Return to %s` (`legacy/app/dispatch_git_timetravel.go:86`) and the first button `Return` |
   | `stashNotFound` | title `Stash Not Found`; explanation `Original stash %s was manually dropped. Continue without restoring stash?`; `Continue`; `Cancel` (`legacy/app/confirm_time_travel_return.go:38-42`) |
   | `timeTravelReturnDirty` | title `Return to %s with uncommitted changes`; explanation `You have changes during time travel. Choose action:\n(Press ESC to cancel)`; `Merge changes`; `Discard changes` (`legacy/app/dispatch_git_timetravel.go:66-81`) |
   | `branchSwitchDirty` | title `Switch to %s with uncommitted changes?`; explanation `You have uncommitted changes. Choose action:\n(ESC to cancel)`; `Stash changes`; `Discard changes` (`legacy/app/handlers_config_branch.go:91-104`) |
   | `branchDelete` | title `Delete branch %s?`; explanation `This cannot be undone. Force delete will drop unmerged commits on this branch.`; `Delete`; `Cancel` (`legacy/app/handlers_config_branch.go:157-169`) |
   | `mergeBranch` | `legacy/app/messages_dialog.go:114-119` |
   | `mergeBranchDirty` | `legacy/app/messages_dialog.go:120-125` |
   | `pin`, `pinDirty`, `unpin`, `unpinDirty`, `pullRebase`, `rebaseBranch`, `rebaseBranchDirty`, `rebaseInteractive` | `PLAN-tit-2-spec.md`, "Dialogs", the rows `pin` to `rebase_interactive` |

   Then one more row: `timeTravelReturnDirtyStash`, value `Stash changes`: the first button of that dialog on a detached HEAD that TIT did not make (`legacy/app/handlers_config_branch.go:57`).

**Validation:**
- Each value equals its legacy text, its ruled text, or a text of ruling 41.
- The legacy title `Return to main without merge?` and the label `Return to main` name a branch that can have another name; the opener of legacy writes the title again (`legacy/app/dispatch_git_timetravel.go:86`). The label `Return` is the menu label (`legacy/app/menu_items.go:175`).
- The count of `%s` of each value equals the count of the values that Steps 7 to 12 give it.

### Step 4: TIT, the menu rows and the key rows
**Scope:** `tit/Source/menu.md`, `tit/Source/keys.md`.
**Action:**
1. `Source/menu.md`: after the rows of Part 11 Step 4, add these rows (`legacy/app/menu_items.go:74-82,147-178`; Part 2, "Menu rows"; ruling 41):

   | key | value | shortcut | emoji | hint |
   |---|---|---|---|---|
   | `dirty pull merge` | Pull (save changes) | `shift + [` | 🔺 | Save WIP, pull remote, reapply changes (may conflict) |
   | `pull rebase` | Pull (rebase) | `r` | | Fetch remote and replay local commits on top |
   | `unpin` | Unpin | `u` | | Return the branch to where it was before the pin |
   | `time travel history` | History | `h` | 🕒 | View commit history while time traveling |
   | `time travel files history` | File(s) history | `g` | 📄 | Browse file changes and diffs |
   | `time travel return` | Return | `r` | 🔙 | Return without merging changes |
   | `dirty continue` | Reapply changes | `c` | ▶️ | Apply your saved changes again |
   | `dirty abort` | Abort and restore | `a` | ↩️ | Restore the state before the operation |

2. `Source/keys.md`: after the rows of Part 11 Step 4, add these rows:

   | key | type | value | mode | description |
   |---|---|---|---|---|
   | `dirty pull merge` | `string` | `{` | `menuNormal` | Pull (save changes). |
   | `add remote` | `string` | `r` | `menuNormal` | Add remote; with a remote, pull (rebase). |
   | `unpin` | `string` | `u` | `menuPinned` | Unpin. |
   | `file history` | `string` | `g` | `menuTimeTraveling` | File(s) history. |
   | `time travel return` | `string` | `r` | `menuTimeTraveling` | Return. |
   | `dirty continue` | `string` | `c` | `menuDirtyOperation` | Reapply changes. |
   | `dirty abort` | `string` | `a` | `menuDirtyOperation` | Abort and restore. |
   | `time travel` | `string` | `return` | `history` | Time travel to the selected commit. |
   | `pin` | `string` | `p` | `history` | Pin the branch at the selected commit. |
   | `rebase interactive` | `string` | `i` | `history` | Rewrite the unpushed commits. |
   | `conflict space` | `string` | `spacebar` | `conflictResolve` | Choose this version for the file. |
   | `conflict enter` | `string` | `return` | `conflictResolve` | Resolve with the chosen versions. |
   | `branch picker enter` | `string` | `return` | `branchPicker` | Switch to the branch. |
   | `branch picker add` | `string` | `a` | `branchPicker` | Add a branch. |
   | `branch picker merge` | `string` | `m` | `branchPicker` | Merge the branch into the current branch. |
   | `branch picker delete` | `string` | `x` | `branchPicker` | Delete the branch. |
   | `branch picker rebase` | `string` | `r` | `branchPicker` | Rebase the current branch onto the branch. |
   | `new branch name submit` | `string` | `return` | `branchName` | Create the branch. |
   | `rebase squash` | `string` | `s` | `rebaseList` | Squash into the commit before. |
   | `rebase fixup` | `string` | `f` | `rebaseList` | Fixup into the commit before. |
   | `rebase reword` | `string` | `r` | `rebaseList` | Change the message. |
   | `rebase drop` | `string` | `d` | `rebaseList` | Drop the commit. |
   | `rebase move down` | `string` | `J` | `rebaseList` | Move the commit down. |
   | `rebase move up` | `string` | `K` | `rebaseList` | Move the commit up. |
   | `rebase apply` | `string` | `return` | `rebaseList` | Apply the list. |
   | `reword submit` | `string` | `return` | `rebaseReword` | Keep the message. |

3. `Source/keys.md`, the mode cells of earlier rows:
   - `history` (Part 11): `menuNormal menuTimeTraveling menuPinned`.
   - `file history`, the row with `f` (Part 11): `menuNormal menuPinned`.
   - `reset discard changes` (Part 10): `menuNormal menuPinned`.
   - `abort merge` (Part 9): `menuMerging`.
   - `history tab` (Part 11): add `conflictResolve branchPicker`.
   - `cancel` (Part 10, Part 11): add `conflictResolve branchPicker branchName rebaseList rebaseReword`.

- The keys are legacy keys (`legacy/app/app_keys.go:78-137`; `legacy/app/menu_items.go:74-178`) and ruled keys (ruling 21). `c` and `a` of the `DirtyOperation` menu are the keys of the `Rebasing` menu (Part 9 Step 4).
- `pull rebase` has no key row: `add remote` calls it (ruling 35). `branchPicker` has no key row in this part (Closed by a read, item 16).
- The second `file history` row is the form of the second `quit` row of Part 10 Step 7.
- In the modes `branchName` and `rebaseReword`, Enter reaches `TITView` from the normal mode of the editor, as in the mode `input` (Part 10, Closed by a read, item 15).
- A key of the resolver, of the picker, and of the rebase list reaches `TITView` in the two ways of Part 11 Step 4. The menu moves its own highlight, so no row of this part moves one. `J` and `K` are not keys of the menu (ruling 63), so they reach the rows of the rebase list.
- When a menu of records has the keys, the menu takes Return and runs the action of the highlighted item (`jam_gui/menu/jam_PopupMenu.h:999-1002,1667-1671`). The action runs the command of the `return` row of the mode (Step 11, item 8). The `return` rows of the modes `history`, `conflictResolve`, and `branchPicker` serve the text panes, where the key goes to the key sheet.

**Validation:**
- In one mode, each key has one row. `r` has one row in each of `menuNormal`, `menuTimeTraveling`, `branchPicker`, and `rebaseList`.
- Each mode name is a row of `map::EditorMode` after Step 1.
- Each menu text equals its legacy text, its ruled text, or a text of ruling 41.

### Step 5: TIT, the rows
**Scope:** `tit/Source/TITModel.cpp`.
**Action:**
1. At the end of the constructor body, after the rows of Part 11 Step 6:

```cpp
    state.appendChild (juce::ValueTree { jam::Format::toType (Id::conflict), { { Id::selectedIndex, 0 } } }, nullptr);
    state.appendChild (juce::ValueTree { jam::Format::toType (Id::branches), { { Id::selectedIndex, 0 } } }, nullptr);
    state.appendChild (juce::ValueTree { jam::Format::toType (Id::rebase), { { Id::selectedIndex, 0 } } }, nullptr);
```

2. The ten new rows of the `git` child, in the forms of Part 8 Step 5 and Part 10 Step 9. `TITModel::getGitTexts()`: add `titStash`, `originalBranch`, `originalHead`, `pinOriginalHead`, `headHash`, `mergeName`, and `commitDate`. `TITModel::getDefaultGitValues()`: add `isMerging`, `isRebasing`, and `hasOriginalStash` with 0. Each one is a parameter, because the detection stores it (ruling 57).

The trees:

| Tree | Rows | Children |
|---|---|---|
| `conflict` | `selectedIndex` | one `file` for each unmerged file |
| `file` of `conflict` | `id` and `value` (the path), `attribute` (the label of the chosen version, or `[ ]`), `selectedIndex` (the chosen version, or -1) | three `version`: base, ours, theirs |
| `version` | `value` (the label), `patch` (the line that prints this version), `commandLines` (the lines that write this version) | none |
| `branches` | `selectedIndex` | one `branch` for each local branch |
| `branch` | `id` and `value` (the name), `attribute` (the mark and the tracking text), `text` (the details), `isCurrent`, `pushedCount` | none |
| `rebase` | `selectedIndex` | one `commit` for each commit of the range: `id` (the hash), `value` (the short hash and the subject), `attribute` (the action), `text` (the message) |

**Validation:**
- Each row holds value data. The `git` child has 26 rows.
- The `file`, `branch`, and `commit` rows have the shape of a record of Part 11 Step 6: `id`, `attribute`, `value`, `text`. So `getRecordMenu` of Part 11 shows each tree with no change. A `version` has no `id` row, so a file of the resolver has no submenu (Part 11 Step 11, item 5).

### Step 6: `TITGitProcessor`, the rows of a protocol
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:**
1. Header, private section: change the declarations of `getOperation` and `removeStaleMarkers`, and add four functions and one constant:

```cpp
    void setBranchRows (const juce::File& directory, const std::optional<juce::String>& branch, bool hasRemote) const;
    void setCommitRows (const juce::File& directory, const std::optional<juce::String>& head) const;
    void setLfsRows (const juce::File& directory) const;
    void setProtocolRows (const juce::File& directory, const juce::String& titStash) const;

    static int getOperation (const juce::File& gitDirectory, const juce::StringArray& records, bool isOnBranch, bool hasTitStash);
    static void removeStaleMarkers (const juce::File& gitDirectory, const std::optional<juce::String>& branch, const std::optional<juce::String>& head, bool hasTitStash);

    static constexpr int originalHeadLine { 1 };
    static constexpr int stashLine { 1 };
```

2. Source: replace `setGit` and `setRepositoryGit` of Part 8 Step 7, items 3 and 5.

```cpp
void TITGitProcessor::setGit (const juce::File& directory) const
{
    const auto gitEnvironment { getGitEnvironment (directory) };
    const auto isRepository { gitEnvironment == map::GitEnvironment::ready and directory.getChildFile (files::gitDirectory).exists() };

    storeValue (Id::gitEnvironment, gitEnvironment);

    if (isRepository)
    {
        setRepositoryGit (directory);
    }
    else
    {
        storeValue (Id::operation, map::Operation::notRepo);

        for (const auto& [id, value] : TITModel::getDefaultGitValues())
            storeValue (id, value);

        for (const auto& id : TITModel::getGitTexts())
            storeValue (id, juce::String {});
    }
}

void TITGitProcessor::setRepositoryGit (const juce::File& directory) const
{
    const juce::String changedRecordTypes { text::Git::changedRecordTypes };
    const auto gitDirectory { directory.getChildFile (files::gitDirectory) };
    const auto head { getOutput (directory, jam::ChildProcess::getArguments (text::Git::head)) };
    const auto branch { getOutput (directory, jam::ChildProcess::getArguments (text::Git::branch)) };
    const auto records (juce::StringArray::fromLines (getOutput (directory, jam::ChildProcess::getArguments (text::Git::status)).value_or (juce::String {})));
    const auto hasRemote { getOutput (directory, jam::ChildProcess::getArguments (text::Git::remote)).value_or (juce::String {}).isNotEmpty() };
    const auto titStash { getOutput (directory, getArguments (text::Git::stashLog, juce::String { text::Git::grepOption } + text::Git::stashLabel)).value_or (juce::String {}) };
    const auto modifiedCount { static_cast<int> (std::count_if (records.begin(), records.end(), [&changedRecordTypes] (const juce::String& record) { return changedRecordTypes.containsChar (record[0]); })) };

    removeStaleMarkers (gitDirectory, branch, head, titStash.isNotEmpty());

    const auto operation { getOperation (gitDirectory, records, branch.has_value(), titStash.isNotEmpty()) };
    const auto hasTimeline { operation == map::Operation::normal and hasRemote and head.has_value() };
    const auto [commitsAhead, commitsBehind] { hasTimeline ? getCommitCounts (directory, *branch) : std::pair<int, int> {} };

    storeValue (Id::operation, operation);
    storeValue (Id::workingTree, modifiedCount > 0 ? map::WorkingTree::dirty : map::WorkingTree::clean);
    storeValue (Id::modifiedCount, modifiedCount);
    storeValue (Id::remote, hasRemote ? map::Remote::hasRemote : map::Remote::noRemote);
    setBranchRows (directory, branch, hasRemote);
    setCommitRows (directory, head);
    setLfsRows (directory);
    setProtocolRows (directory, titStash);
    setTimeline (commitsAhead, commitsBehind, hasTimeline);
}
```

3. Source: the four row groups. The first three hold the `storeValue` lines of Part 8 and Part 10 for their rows, with no change of a value.

   | Function | Rows | Source of each line |
   |---|---|---|
   | `setBranchRows` | `currentBranch`, `hasUpstream`, `isDetached`, `isTitTimeTravel`, `remoteUrl` | Part 8 Step 7, item 5; Part 10 Step 9, item 3. `gitDirectory` is `directory.getChildFile (files::gitDirectory)`. |
   | `setCommitRows` | `currentHash` (Part 8); `headHash`: `head.value_or (juce::String {})`; `commitDate`: `getOutput (directory, jam::ChildProcess::getArguments (text::Git::commitDate)).value_or (juce::String {})` | |
   | `setLfsRows` | `isUsingLfs`, `isLfsReady` | Part 8 Step 7, item 5 |

```cpp
void TITGitProcessor::setProtocolRows (const juce::File& directory, const juce::String& titStash) const
{
    const auto gitDirectory { directory.getChildFile (files::gitDirectory) };
    const auto dirtyMarker (juce::StringArray::fromLines (gitDirectory.getChildFile (files::dirtyOperationMarker).loadFileAsString().trim()));
    const auto pinMarker (juce::StringArray::fromLines (gitDirectory.getChildFile (files::pinMarker).loadFileAsString().trim()));
    const auto timeTravelMarker (juce::StringArray::fromLines (gitDirectory.getChildFile (files::timeTravelMarker).loadFileAsString()));
    const auto isMerging { gitDirectory.getChildFile (files::mergeHead).exists() };

    storeValue (Id::titStash, titStash);
    storeValue (Id::originalBranch, dirtyMarker[branchLine].trim());
    storeValue (Id::originalHead, dirtyMarker[originalHeadLine].trim());
    storeValue (Id::pinOriginalHead, pinMarker[originalHeadLine].trim());
    storeValue (Id::hasOriginalStash, timeTravelMarker[stashLine].trim().isNotEmpty());
    storeValue (Id::isMerging, isMerging);
    storeValue (Id::isRebasing, gitDirectory.getChildFile (files::rebaseMerge).exists() or gitDirectory.getChildFile (files::rebaseApply).exists());
    storeValue (Id::mergeName, isMerging ? getOutput (directory, jam::ChildProcess::getArguments (text::Git::mergeName)).value_or (juce::String {}) : juce::String {});
}
```

4. Source: `getOperation`. The table of Part 8 Step 7, item 6, gets one entry and one more element: `std::array<std::pair<bool, int>, 6>`. After the entry of `timeTraveling`, add `{ hasTitStash, map::Operation::dirtyOperation },`. The other entries and their order stay.
5. Source: `removeStaleMarkers`. Add as its last statement:

```cpp
    if (not hasTitStash)
        gitDirectory.getChildFile (files::dirtyOperationMarker).deleteFile();
```

- The detection stores each row with `storeValue`, on the pool thread, as in Part 8 Step 7 (ruling 57). It makes no tree.
- `setGit` has no test of the dirty marker now: `Id::operation` with `dirtyOperation` comes from the stash of TIT (Closed by a read, item 3).
- The order of the operations is: conflicted, merging, rebasing, time traveling, dirty operation, pinned, normal. A stash of TIT during a time travel is the stash of the entry; the return applies it (Step 9, `getReturnLines`).
- The marker rows are read after `removeStaleMarkers`, so a stale marker gives empty rows.
- `stashLine` is the second line of the time travel marker: the label of the stash of the entry, or an empty line (Closed by a read, item 14). The marker text is not trimmed as a whole, so an empty second line stays a line. `hasOriginalStash` is true when that line holds a text.
- A missing marker gives an empty array, and `juce::StringArray::operator[]` gives the empty text (Part 8 Step 7, item 7).
- Fixture 18 of Part 8 Step 9, an empty marker file with no stash, now gives `operation 1`, and the marker is removed.

**Validation:**
- No function body is longer than 30 lines. No bail-out guard. No `==` on strings.
- Each `set…` function replaces rows of the Model with `storeValue` (NAMES.md Rule 1, `set`; Part 8 Step 6). Each one holds one group of rows: the branch, the commit of HEAD, LFS, the protocol.
- `PLAN-tit-8-git.md` Step 9, fixture 18, and its Closed by a read, item 4, name the old rule. This step replaces both.
- New names and their families: `setBranchRows`, `setCommitRows`, `setLfsRows`, `setProtocolRows` (`setTimeline`, `setRepositoryGit`, Part 8 Step 6; the title of this part); `originalHeadLine`, `stashLine` (`branchLine`, `pinnedHashLine`, Part 8 Step 6); `timeTravelMarker` (`files::timeTravelMarker`); `titStash`, `dirtyMarker`, `pinMarker`, `isMerging`, `hasTitStash` (the rows of Step 2).

### Step 7: `TITGitProcessor`, the three trees of records
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:** Each tree is a feed of records: it goes through the lane of Part 11 Step 7, and the message thread builds it (rulings 57 and 60). The content of a version of a file is output of the tool: it goes through a run chain (ruling 58).
1. Header, public section, after `readPatch`:

```cpp
    void readConflicts (const juce::StringArray& labels);
    void readBranches();
    void readRebase (const juce::String& range);
    void readVersions (const juce::StringArray& patchLines);
```

   Public section, after the requests: `static constexpr int numVersions { 3 };` and `static constexpr int noVersion { -1 };`. `ConflictView` reads the first, and the availability rule of `conflictEnter` reads the second. After the public member `patch` of Part 11: `std::array<TITRunProcessor, numVersions> versions;`. The constructor initializes it with `versions { { { modelToUse, Id::base }, { modelToUse, Id::ours }, { modelToUse, Id::theirs } } }`, and `prepare()` of Part 10 Step 10 gets a range-for that calls `prepare()` on each element. In the loop of the run chains of Part 10 Step 9, add `Id::base`, `Id::ours`, and `Id::theirs` to the list.

   Private section, after `addFiles`:

```cpp
    void addConflictFile (const juce::String& record, const juce::StringArray& labels);
    void addBranch (const juce::File& directory, const juce::String& record, const juce::String& base);

    static juce::String getBranchDetails (const juce::StringArray& fields, bool isCurrent);

    static constexpr int firstStage { 1 };
    static constexpr int firstStageHashField { 7 };
    static constexpr int unmergedPathField { 10 };
    static constexpr int messageField { 3 };
    static constexpr int nameField { 0 };
    static constexpr int currentField { 1 };
    static constexpr int shortHashField { 2 };
    static constexpr int subjectField { 3 };
    static constexpr int authorField { 4 };
    static constexpr int commitDateField { 5 };
    static constexpr int upstreamField { 6 };
    static constexpr int trackField { 7 };
```

2. Source: the requests. Each feed runs on the message thread and adds one pool job, in the form of `read` of Part 11 Step 7.

```cpp
void TITGitProcessor::readConflicts (const juce::StringArray& labels)
{
    pool.addJob ([this, labels, directory = juce::File { model.state[Id::directory].toString() }]
    {
        for (const auto& record : juce::StringArray::fromLines (getOutput (directory, jam::ChildProcess::getArguments (text::Git::status)).value_or (juce::String {})))
            if (record.startsWith (text::Git::unmergedRecordTypes))
                addConflictFile (record, labels);

        addRecord (recordLevel, jam::Format::toType (Id::conflict), {});
    });
}

void TITGitProcessor::readBranches()
{
    pool.addJob ([this, directory = juce::File { model.state[Id::directory].toString() }]
    {
        const auto tab { juce::String::charToString (Chars::tab) };
        const auto base { getOutput (directory, jam::ChildProcess::getArguments (text::Git::mergeBase)).value_or (juce::String {}) };
        auto lines (juce::StringArray::fromLines (getOutput (directory, jam::ChildProcess::getArguments (text::Git::branches)).value_or (juce::String {})));

        std::stable_partition (lines.begin(), lines.end(), [&tab] (const juce::String& line) { return juce::StringArray::fromTokens (line, tab, {})[currentField].trim().isNotEmpty(); });

        for (const auto& line : lines)
            addBranch (directory, line, base);

        addRecord (recordLevel, jam::Format::toType (Id::branches), {});
    });
}

void TITGitProcessor::readRebase (const juce::String& range)
{
    pool.addJob ([this, range, directory = juce::File { model.state[Id::directory].toString() }]
    {
        const auto output { getOutputText (directory, getArguments (text::Git::rebaseLog, range)) };

        for (const auto record : getRecords (output))
        {
            const auto fields (getFields (record));

            addRecord (recordLevel, jam::Format::toType (Id::commit),
                       { { Id::id, fields[hashField] }, { Id::value, fields[valueField] }, { Id::attribute, text::Rebase::pick }, { Id::text, fields[messageField] } });
        }

        addRecord (recordLevel, jam::Format::toType (Id::rebase), {});
    });
}

void TITGitProcessor::readVersions (const juce::StringArray& patchLines)
{
    for (size_t index { 0 }; index < versions.size(); ++index)
    {
        versions.at (index).clearHistory();

        if (patchLines[static_cast<int> (index)].isNotEmpty())
            versions.at (index).run (patchLines[static_cast<int> (index)]);
    }
}
```

   In `registerFeedCompletion` of Part 11 Step 7, item 4, replace the `reading` entry with one entry for each type of this list, with the same body: `Id::reading`, `Id::conflict`, `Id::branches`, `Id::rebase`.

   - Each feed writes its records, then the record of its tree. The message thread moves the complete feed into the tree and sends one change message (Part 11 Step 7, item 4).
   - The current branch is the first record: `std::stable_partition` puts the line with the mark of the tool first and keeps the order of the others.
   - The records of `rebaseLog` have the fields of Part 11 Step 2, item 4, and the message as the fourth field. Each row starts with the action `pick`.
   - `readVersions` shows the three versions of one file: each chain runs the `patch` line of its version, or stays empty for a version that does not exist. The view gives the lines of the selected file (Step 11).

3. Source: one unmerged file.

```cpp
void TITGitProcessor::addConflictFile (const juce::String& record, const juce::StringArray& labels)
{
    const auto space { juce::String::charToString (Chars::space) };
    const auto colon { juce::String::charToString (Chars::colon) };
    const auto newline { juce::String::charToString (Chars::newline) };
    const auto fields (juce::StringArray::fromTokens (record, false));
    const auto path { fields.joinIntoString (space, unmergedPathField) };

    addRecord (recordLevel, jam::Format::toType (Id::file), { { Id::id, path }, { Id::value, path }, { Id::attribute, text::Conflict::unmarked }, { Id::selectedIndex, noVersion } });

    for (int version { 0 }; version < numVersions; ++version)
    {
        const juce::String stage { version + firstStage };
        const auto isPresent { not fields[firstStageHashField + version].containsOnly (text::Git::zeroHash) };

        addRecord (childLevel, jam::Format::toType (Id::version),
                   { { Id::value, isPresent ? labels[version] : juce::String { text::Conflict::deleted }.replace (text::Reading::hashPlaceholder, labels[version]) },
                     { Id::patch, isPresent ? text::Git::showStage + space + (colon + stage + colon + path).quoted() : juce::String {} },
                     { Id::commandLines, isPresent ? text::Git::checkoutStage + stage + space + text::Git::pathOption + space + path.quoted() + newline + text::Git::add + space + path.quoted()
                                                   : text::Git::remove + space + path.quoted() } });
    }
}
```

   - The fields of the record are in the inventory (the facts of the tool). The hash of a stage that does not exist holds only the character `0`.
   - The path is the text after field 10, joined with spaces, as legacy joins it (`legacy/git/exec_conflict.go:34`).
   - A file is a record of `recordLevel`, and its three versions are records of `childLevel` (Part 11 Step 7, item 3).
   - The row `value` of a version is its label, or the text of a version that does not exist. The row `patch` is the line that prints the version: the stage and the path are one quoted argument. The row `commandLines` holds the lines that write the version and stage it, or remove the file (Closed by a read, item 9). All three are complete at creation.
   - No row holds the content of a version (ruling 58).

4. Source: one branch.

```cpp
void TITGitProcessor::addBranch (const juce::File& directory, const juce::String& record, const juce::String& base)
{
    const auto space { juce::String::charToString (Chars::space) };
    const auto fields (juce::StringArray::fromTokens (record, juce::String::charToString (Chars::tab), {}));
    const auto isCurrent { fields[currentField].trim().isNotEmpty() };
    const auto tracking { fields[upstreamField].isEmpty() ? juce::String { text::Branch::local }
                                                         : fields[trackField].isEmpty() ? juce::String { text::Branch::synced } : fields[trackField] };
    const auto pushedCount { base.isEmpty() ? 0
                                            : getOutput (directory, getArguments (text::Git::countCommits, fields[nameField] + text::Git::rangeSeparator + base)).value_or (juce::String {}).getIntValue() };

    addRecord (recordLevel, jam::Format::toType (Id::branch),
               { { Id::id, fields[nameField] },
                 { Id::value, fields[nameField] },
                 { Id::attribute, isCurrent ? text::Branch::currentMark + space + tracking : tracking },
                 { Id::text, getBranchDetails (fields, isCurrent) },
                 { Id::isCurrent, isCurrent ? 1 : 0 },
                 { Id::pushedCount, pushedCount } });
}
```

   - `base` is the commit that HEAD and its upstream share. `<branch>..<base>` is the set of the commits that are pushed and that a rebase onto `<branch>` replays (Closed by a read, item 26). With no upstream, `base` is empty and the count is 0.
   - The tracking text is the text of the tool, `ahead 1, behind 2`. Legacy turned it into `↑1 ↓2` with a parser of its own (`legacy/git/branch.go:63-78`); the plan-writer rule is "No manual string parsing".
   - The rows `isCurrent` and `pushedCount` arrive as texts, `1` or `0` and a number. `static_cast<bool>` and `static_cast<int>` of the row give the flag and the count (JUCE `juce_core/containers/juce_Variant.cpp`, `VariantType_String::toBool` and `toInt`).
   - The details of a branch are a text that TIT makes from the fields, not output that a view shows as the tool wrote it. So they are a row, `text`.

5. Source: `getBranchDetails`. It returns the lines below, joined with `Chars::newline`. Each `%s` is filled with `juce::String::replaceFirstOccurrenceOf (text::Reading::hashPlaceholder, …)`.

   | Line | Text |
   |---|---|
   | 1 | `text::Branch::heading` |
   | 2 | `text::Branch::name`, with the name field |
   | 3 | `text::Branch::statusCurrent` when `isCurrent`, else `statusOther` |
   | 4 | `text::Branch::upstreamNone` when the upstream field is empty; else `text::Branch::upstream`, with the upstream field and the tracking text or `synced` |
   | 5 | empty |
   | 6 | `text::Branch::tipHeading` |
   | 7 to 9 | `hash`, `author`, `date`, each with its field |
   | 10 | empty |
   | 11 | the subject field |

**Validation:**
- No bail-out guard. No function body is longer than 30 lines. No `==` on strings.
- The pool thread writes the lane and stores parameters. It reads and writes no tree, and it calls no `juce::MessageManager::callAsync` (ruling 57).
- Each record is complete at creation: the labels, the lines, the details, and the counts.
- The lines of a branch equal legacy (`legacy/ui/branchpicker.go:141-173`), except the indent of two spaces, which is a layout (ruling 7), and the form of the tracking text.
- New names and their families: `readConflicts`, `readBranches`, `readRebase`, `readVersions` (`readChanges`, `readPatch`, Part 11); `addConflictFile`, `addBranch` (`addCommits`, `addFiles`, Part 11); `getBranchDetails` (`getFields`, Part 11); `versions` (the row type `version`; `patch`, Part 11); the field constants (the fields of Step 2, item 5, and of the unmerged record); `messageField` (the fourth field of `rebaseLog`); `numVersions`, `firstStage`, `noVersion` (the row `version`; the word "stage" of the tool). Local names: `record`, `lines`, `line`, `tab`, `base`, `output`, `space`, `colon`, `newline`, `fields`, `path`, `version`, `stage`, `isPresent`, `isCurrent`, `tracking`, `pushedCount`, `patchLines`, `index`.

### Step 8: `HeaderView`, `FooterView`
**Scope:** `tit/Source/HeaderView.cpp`, `tit/Source/FooterView.cpp`.
**Action:**
1. This part has no list component (Part 11 Step 8). `TITView` builds the menu of a tree again when a row of the tree changes (Step 11, item 8).
2. `HeaderView`: `valueTreePropertyChanged` of Part 10 Step 11 already repaints on each row of the `git` child. In `getTimelineDescription`, replace the rule "The operation is `timeTraveling`: the empty text" with:
   - `isTitTimeTravel`: `text::Header::viewingCommit`, a space, and `juce::Time::fromISO8601 (…).formatted (text::Header::commitDateFormat)` of the `commitDate` row.
   - a `timeTraveling` that TIT did not make: `text::Header::notOnBranch`.
3. `FooterView`: replace `getHint()` of Part 11 Step 12:

```cpp
juce::String FooterView::getHint() const
{
    static const jam::HashMap<int, std::array<const char*, 2>> hints {
        { map::EditorMode::console, { text::Footer::consoleRunning, text::Footer::consoleComplete } },
        { map::EditorMode::history, { text::Footer::history, text::Footer::readingText } },
        { map::EditorMode::conflictResolve, { text::Footer::conflictFiles, text::Footer::conflictVersion } },
        { map::EditorMode::branchPicker, { text::Footer::branchPickerOther, text::Footer::branchPickerCurrent } },
        { map::EditorMode::confirmation, { text::Footer::confirmation, text::Footer::confirmation } },
        { map::EditorMode::input, { text::Footer::input, text::Footer::input } },
        { map::EditorMode::branchName, { text::Footer::input, text::Footer::input } },
        { map::EditorMode::rebaseReword, { text::Footer::input, text::Footer::input } },
        { map::EditorMode::rebaseList, { text::Footer::rebaseList, text::Footer::rebaseList } },
    };

    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };
    const auto pane { static_cast<int> (state.getProperty (Id::focusedPane)) };
    const auto branches { state.getChildWithName (jam::Format::toType (Id::branches)) };
    const jam::HashMap<int, bool> alternates {
        { map::EditorMode::console, not static_cast<bool> (state.getProperty (Id::isRunning)) },
        { map::EditorMode::history, pane == map::Pane::text },
        { map::EditorMode::conflictResolve, pane != map::ConflictPane::files },
        { map::EditorMode::branchPicker, static_cast<bool> (branches.getChild (static_cast<int> (branches.getProperty (Id::selectedIndex))).getProperty (Id::isCurrent)) },
    };

    if (hints.contains (mode)) return hints.at (mode).at (alternates.contains (mode) and alternates.at (mode) ? 1 : 0);
    return {};
}
```

   The footer listens to the root, so the `selectedIndex` of the `branches` tree reaches it. Add `Id::selectedIndex` to `paintedProperties`. The hint of the picker follows the selected branch, which is the highlighted branch at the last key that the menu did not take (Part 11 Step 11, item 5).

**Validation:**
- `getHint()` has one branch. Each hint of Part 11 Step 12 keeps its text and its rule.
- No file of a list component exists. A menu mode has no hint (Part 10 Step 13).
- The header texts equal legacy (`legacy/app/app_view_header.go:57,98`), with the date form of Closed by a read, item 30.

### Step 9: `TITGitProcessor`, the protocol lines, the markers, the rebase list
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:** A protocol is business logic: which command lines run, in which order, and which marker goes with them. So each function of this step is a member of the processor (Part 9, Overview, item 6; MANIFESTO E, MVP: "The Processor holds the business logic"). `TITView` gives the values that the user selected, and takes the lines.

1. Header, public section, after `getRewindLines` of Part 11 Step 7:

```cpp
    static juce::String getLine (const juce::StringArray& words);
    static juce::StringArray getStashLines (const juce::ValueTree& git);
    static juce::StringArray getDirtyLines (const juce::StringArray& changeSet);
    static juce::StringArray getDiscardLines (const juce::StringArray& changeSet);
    static juce::StringArray getReturnLines (const juce::ValueTree& git, const juce::String& branch);
    static juce::StringArray getReturnMergeLines (const juce::ValueTree& git, const juce::String& branch);
    static juce::StringArray getDirtyAbortLines (const juce::ValueTree& git);
    static juce::StringArray getTimeTravelLines (const juce::ValueTree& git, const juce::String& hash);
    static juce::StringArray getConflictLines (const juce::ValueTree& git, const juce::ValueTree& conflict);
    static juce::StringArray getConflictAbortLines (const juce::ValueTree& git);
    static juce::String getRebaseRange (const juce::ValueTree& git);
    static juce::String getRebaseLine (const juce::ValueTree& git);
    static juce::StringArray getDirtyMarkerLines (const juce::ValueTree& git);
    static juce::StringArray getTimeTravelMarkerLines (const juce::ValueTree& git);
    static juce::StringArray getPinMarkerLines (const juce::ValueTree& git, const juce::String& hash);
```

   After `setCommitMessage` of Part 10 Step 10:

```cpp
    void setMarker (const juce::String& fileName, const juce::String& text);
    void setRebaseAction (const char* action);
    void setRebaseMessage (const juce::String& message);
    void setRebaseFiles();
    void moveRebaseCommit (int delta);
```

   Private constant, after the field constants of Part 11: `static constexpr const char* lineEnding { "\n" };`.

2. The line builders:

```cpp
juce::String TITGitProcessor::getLine (const juce::StringArray& words)
{
    return words.joinIntoString (juce::String::charToString (Chars::space));
}

juce::StringArray TITGitProcessor::getStashLines (const juce::ValueTree& git)
{
    const auto titStash { git.getProperty (Id::titStash).toString() };

    if (titStash.isNotEmpty()) return { getLine ({ text::Git::stashApply, titStash }), getLine ({ text::Git::stashDrop, titStash }) };
    return {};
}

juce::StringArray TITGitProcessor::getDirtyLines (const juce::StringArray& changeSet)
{
    juce::StringArray lines { getLine ({ text::Git::stashPush, juce::String { text::Git::stashLabel }.quoted() }) };

    lines.addArray (changeSet);
    lines.add (text::Git::stashApply);
    lines.add (text::Git::stashDrop);
    return lines;
}

juce::StringArray TITGitProcessor::getDiscardLines (const juce::StringArray& changeSet)
{
    juce::StringArray lines { text::Git::resetHardHead, text::Git::clean };

    lines.addArray (changeSet);
    return lines;
}

juce::StringArray TITGitProcessor::getReturnLines (const juce::ValueTree& git, const juce::String& branch)
{
    juce::StringArray lines { getLine ({ text::Git::checkout, branch }) };

    lines.addArray (getStashLines (git));
    return lines;
}

juce::StringArray TITGitProcessor::getReturnMergeLines (const juce::ValueTree& git, const juce::String& branch)
{
    juce::StringArray lines { getLine ({ text::Git::stashPush, juce::String { text::Git::stashLabel }.quoted() }),
                              getLine ({ text::Git::checkout, branch }),
                              getLine ({ text::Git::merge, git.getProperty (Id::headHash).toString() }),
                              text::Git::stashApply,
                              text::Git::stashDrop };

    lines.addArray (getStashLines (git));
    return lines;
}

juce::StringArray TITGitProcessor::getDirtyAbortLines (const juce::ValueTree& git)
{
    juce::StringArray lines { getLine ({ text::Git::checkout, git.getProperty (Id::originalBranch).toString() }),
                              getLine ({ text::Git::resetHard, git.getProperty (Id::originalHead).toString() }) };

    lines.addArray (getStashLines (git));
    return lines;
}

juce::StringArray TITGitProcessor::getTimeTravelLines (const juce::ValueTree& git, const juce::String& hash)
{
    const auto isTravelling { static_cast<int> (git.getProperty (Id::operation)) == map::Operation::timeTraveling };
    const auto isDirty { static_cast<int> (git.getProperty (Id::workingTree)) == map::WorkingTree::dirty };
    juce::StringArray lines { getLine ({ text::Git::checkout, hash }) };

    if (isDirty and not isTravelling)
        lines.insert (0, getLine ({ text::Git::stashPush, juce::String { text::Git::stashLabel }.quoted() }));

    return lines;
}
```

   - `getLine` joins one command line row with its arguments: a branch name, a hash, a stash reference. The view gives the words and joins none.
   - `getDirtyLines` is the protocol of `tit/SPEC.md:350-374`: the snapshot, the change-set, the apply, the drop. The stash of this chain is the newest stash from its first line to its last, so its last two lines name no stash, as SPEC writes them.
   - `getStashLines` names the stash of the row `titStash`: a stash that was there before the chain started.
   - `getDiscardLines` is the discard path of legacy (`legacy/app/op_dirty_pull_snapshot.go:76,85`).
   - `getReturnMergeLines` is the merge of `tit/SPEC.md:557-593`: stash, checkout, merge the commit of the journey, apply, drop. Then the stash of the entry, when one exists: after the drop of the new stash, the reference of the row is valid again.
   - `getDirtyAbortLines` is "Abort and restore" (`tit/SPEC.md:376-385`).
   - `getTimeTravelLines` is the legacy jump. A jump from one commit to another during a journey makes no stash.

3. The lines of the resolver and of the rebase list:

```cpp
juce::StringArray TITGitProcessor::getConflictLines (const juce::ValueTree& git, const juce::ValueTree& conflict)
{
    const auto titStash { git.getProperty (Id::titStash).toString() };
    juce::StringArray lines;

    for (const auto& file : conflict)
        lines.addArray (juce::StringArray::fromLines (file.getChild (static_cast<int> (file.getProperty (Id::selectedIndex))).getProperty (Id::commandLines).toString()));

    if (static_cast<bool> (git.getProperty (Id::isMerging)))
        lines.add (text::Git::commitNoEdit);
    else if (static_cast<bool> (git.getProperty (Id::isRebasing)))
        lines.add (text::Git::rebaseContinue);
    else if (titStash.isNotEmpty())
        lines.add (getLine ({ text::Git::stashDrop, titStash }));

    return lines;
}

juce::StringArray TITGitProcessor::getConflictAbortLines (const juce::ValueTree& git)
{
    if (static_cast<bool> (git.getProperty (Id::isMerging))) return { text::Git::mergeAbort, text::Git::resetHard };
    if (static_cast<bool> (git.getProperty (Id::isRebasing))) return { text::Git::rebaseAbort };
    return { text::Git::resetHardHead };
}

juce::String TITGitProcessor::getRebaseRange (const juce::ValueTree& git)
{
    if (static_cast<bool> (git.getProperty (Id::hasUpstream))) return juce::String { text::Git::upstreamName } + text::Git::outgoingSuffix;
    return text::Git::headName;
}

juce::String TITGitProcessor::getRebaseLine (const juce::ValueTree& git)
{
    return getLine ({ text::Git::rebaseInteractive, static_cast<bool> (git.getProperty (Id::hasUpstream)) ? text::Git::keepBaseUpstream : text::Git::rootOption });
}
```

   - `getConflictLines` joins the command lines of each chosen version, then the finalize of the conflict type (Closed by a read, item 5). All lines are one run.
   - `getConflictAbortLines` aborts by the type: the legacy lines of a merge and of a rebase (Part 9, inventory). For a conflict of `stash apply`, `reset --hard HEAD` removes the partial apply; the stash stays, and the `DirtyOperation` menu follows.
   - `getRebaseRange` is `@{u}..HEAD`, or every commit of HEAD with no upstream (rulings 9 and 25). `getRebaseLine` is the rebase line with its base.

4. The markers:

```cpp
juce::StringArray TITGitProcessor::getDirtyMarkerLines (const juce::ValueTree& git)
{
    return { git.getProperty (Id::currentBranch).toString(), git.getProperty (Id::headHash).toString() };
}

juce::StringArray TITGitProcessor::getTimeTravelMarkerLines (const juce::ValueTree& git)
{
    const auto isDirty { static_cast<int> (git.getProperty (Id::workingTree)) == map::WorkingTree::dirty };

    return { git.getProperty (Id::currentBranch).toString(), isDirty ? juce::String { text::Git::stashLabel } : juce::String {} };
}

juce::StringArray TITGitProcessor::getPinMarkerLines (const juce::ValueTree& git, const juce::String& hash)
{
    const auto isPinned { static_cast<int> (git.getProperty (Id::operation)) == map::Operation::pinned };

    return { git.getProperty (Id::currentBranch).toString(), git.getProperty (isPinned ? Id::pinOriginalHead : Id::headHash).toString(), hash };
}

void TITGitProcessor::setMarker (const juce::String& fileName, const juce::String& text)
{
    juce::File { model.state[Id::directory].toString() }.getChildFile (files::gitDirectory).getChildFile (fileName).replaceWithText (text, false, false, lineEnding);
}
```

   - The dirty marker holds the branch and the HEAD before the change-set (`tit/SPEC.md:352-356`). The time travel marker holds the branch. The pin marker holds the branch, the HEAD before the first pin, and the pinned hash (Part 2 Step 3). A pin during a pin keeps the HEAD before the first pin (Part 2 Step 7, item 6).
   - `setMarker` writes one marker file into the repository data folder. `lineEnding` is the fourth argument of `juce::File::replaceWithText`, whose default is `"\r\n"` (doxygen `classjuce_1_1File`).

5. The rebase list. The `rebase` tree is a tree of the processor (Step 7), so its edits are functions of the processor:

```cpp
void TITGitProcessor::setRebaseAction (const char* action)
{
    const auto rebase { model.state.getChildWithName (jam::Format::toType (Id::rebase)) };
    auto commit { rebase.getChild (static_cast<int> (rebase.getProperty (Id::selectedIndex))) };

    if (commit.isValid())
        commit.setProperty (Id::attribute, commit.getProperty (Id::attribute).toString().compare (action) == 0 ? text::Rebase::pick : action, nullptr);
}

void TITGitProcessor::setRebaseMessage (const juce::String& message)
{
    const auto rebase { model.state.getChildWithName (jam::Format::toType (Id::rebase)) };
    auto commit { rebase.getChild (static_cast<int> (rebase.getProperty (Id::selectedIndex))) };

    commit.setProperty (Id::text, message, nullptr);
    commit.setProperty (Id::attribute, text::Rebase::reword, nullptr);
}

void TITGitProcessor::moveRebaseCommit (int delta)
{
    auto rebase { model.state.getChildWithName (jam::Format::toType (Id::rebase)) };
    const auto index { static_cast<int> (rebase.getProperty (Id::selectedIndex)) };
    const auto target { juce::jlimit (0, juce::jmax (0, rebase.getNumChildren() - 1), index + delta) };

    if (rebase.getNumChildren() > 0)
    {
        rebase.moveChild (index, target, nullptr);
        rebase.setProperty (Id::selectedIndex, target, nullptr);
    }
}

void TITGitProcessor::setRebaseFiles()
{
    const auto newline { juce::String::charToString (Chars::newline) };
    const auto folder { juce::File { model.state[Id::directory].toString() }.getChildFile (files::gitDirectory).getChildFile (files::rebaseDirectory) };
    juce::StringArray todo;

    folder.deleteRecursively();
    folder.createDirectory();

    for (const auto& commit : model.state.getChildWithName (jam::Format::toType (Id::rebase)))
    {
        const auto action { commit.getProperty (Id::attribute).toString() };
        const auto hash { commit.getProperty (Id::id).toString() };
        const auto isReword { action.compare (text::Rebase::reword) == 0 };

        todo.add (getLine ({ isReword ? juce::String { text::Rebase::pick } : action, hash }));

        if (isReword)
        {
            folder.getChildFile (hash).replaceWithText (commit.getProperty (Id::text).toString() + newline, false, false, lineEnding);
            todo.add (text::Git::amendFile + hash);
        }
    }

    folder.getChildFile (files::rebaseTodo).replaceWithText (todo.joinIntoString (newline) + newline, false, false, lineEnding);
}
```

   - The key of the action that a row has sets the row to `pick` (Closed by a read, item 23). A row moves with `juce::ValueTree::moveChild`, and the selection follows it.
   - A reword keeps the new message in the `text` row of the commit. The list file then holds `pick <hash>` and the `exec` line with the message file, whose name is the hash (Closed by a read, items 20 and 21).
   - The folder is replaced whole before the dialog opens. A dialog that the user cancels leaves the folder; the next apply replaces it. No other code reads it.
   - The tool reads only the command and the commit name of each line (inventory, the facts of the tool). A list that the tool refuses, for example a `squash` on the first row, shows the text of the tool in the console and changes nothing.
   - These five functions run on the message thread, as `run` does. They write Model rows and files, and no pool job reads either.

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings: each text test is `compare (…) == 0`.
- Each command line is a cast row with a value of a row of the Model. No function of this step reads the repository.
- The Dirty Operation Protocol has its four steps and its abort in SPEC order (`tit/SPEC.md:350-385`).
- The processor writes three kinds of files, each in the repository data folder: the commit message of Part 10, a marker, and the folder of the rebase list. No view writes a file.
- New names and their families: `getLine` (`getPushLine`, Part 9); `getStashLines`, `getDirtyLines`, `getDiscardLines`, `getReturnLines`, `getReturnMergeLines`, `getDirtyAbortLines`, `getTimeTravelLines`, `getConflictLines`, `getConflictAbortLines` (`getReplaceLocalLines`, Part 10; the commands and the protocol words of Part 2); `getRebaseRange`, `getRebaseLine` (`getUpstream`, `getPushLine`); `getDirtyMarkerLines`, `getTimeTravelMarkerLines`, `getPinMarkerLines` (the marker files of Part 8); `setMarker` (`setCommitMessage`, Part 10); `setRebaseAction`, `setRebaseMessage`, `setRebaseFiles`, `moveRebaseCommit` (the tree `rebase`; `moveSelection`); `lineEnding` (the parameter `lineEndings` of `juce::File::replaceWithText`). Local names: `titStash`, `lines`, `isTravelling`, `isDirty`, `file`, `isPinned`, `rebase`, `commit`, `index`, `target`, `newline`, `folder`, `todo`, `action`, `hash`, `isReword`.

### Step 10: `TextView`, `ConflictView`, `BranchView`, and the text of an `InputView`
**Scope:** `tit/Source/TextView.h`, `tit/Source/TextView.cpp`, `tit/Source/ConflictView.h`, `tit/Source/ConflictView.cpp`, `tit/Source/BranchView.h`, `tit/Source/BranchView.cpp` (all new); `tit/Source/InputView.h`, `tit/Source/InputView.cpp`; `tit/Source/HistoryView.cpp`.
**Action:** The resolver and the picker are each one screen, so each is one view class in the form of `HistoryView` (Part 11 Step 10). The files of the resolver, the branches of the picker, and the commits of the rebase list show in the stock menu, which `TITView` builds from their trees (Step 11, item 8; Part 11 Step 11, item 5). The rebase list has no pane, so it has no view class.

1. `TextView`, in `TextView.h` and `TextView.cpp`, with the includes of Part 10 Step 11. It is a read-only pane for a text that TIT makes: one `jam::TextEditor` (ruling 5) with one document.

```cpp
class TextView : public juce::Component
{
public:
    TextView();
    ~TextView() override = default;

    void resized() override;
    void focusGained (FocusChangeType cause) override;

    void setText (const juce::String& text);

private:
    jam::TextEditor editor;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TextView)
};
```

   - The constructor, `resized()`, and `focusGained`: the three bodies of `InputView` (Part 10 Step 14, item 8).
   - `setText`:

```cpp
void TextView::setText (const juce::String& text)
{
    const auto bytes { (text + juce::String::charToString (Chars::newline)).toStdString() };
    auto document { std::make_unique<jam::AnsiDocument>() };

    document->addLines (bytes.data(), static_cast<int> (bytes.size()));
    editor.setLiveBlock (std::move (document));
}
```

   - The editor stays in the read-only `reader` mode of a live block (Part 6). A pane for output of the tool is a `ConsoleView` over a run chain, not this view (ruling 58; Part 10 Step 14).

2. `ConflictView`. The header, with the includes of Part 10 Step 11, then `#include "TITGitProcessor.h"` and `#include "ConsoleView.h"`:

```cpp
class ConflictView : public juce::Component,
                     public jam::Model::Component<ConflictView>,
                     private juce::ValueTree::Listener
{
public:
    ConflictView (juce::ValueTree stateToUse, std::array<TITRunProcessor, TITGitProcessor::numVersions>& chains);
    ~ConflictView() override = default;

    void resized() override;
    void focusGained (FocusChangeType cause) override;

    juce::ValueTree getFile() const;

    static constexpr int numPanes { 4 };

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    void setLabels();
    void setFocus();

    static constexpr int labelRows { 1 };

    std::array<juce::Label, TITGitProcessor::numVersions> labels;
    std::array<ConsoleView, TITGitProcessor::numVersions> versions;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConflictView)
};
```

3. The constructor: the form of Part 10 Step 11, item 2, with this initializer of `versions`. Then `setWantsKeyboardFocus (true);` and `addAndMakeVisible` for each element of `labels` and of `versions` in two range-fors. The state is the root of the Model.

```cpp
    , versions { { { stateToUse.getChildWithName (jam::Format::toType (Id::base)), chains.at (0).userInterfaceGetters },
                   { stateToUse.getChildWithName (jam::Format::toType (Id::ours)), chains.at (1).userInterfaceGetters },
                   { stateToUse.getChildWithName (jam::Format::toType (Id::theirs)), chains.at (2).userInterfaceGetters } } }
```

4. The bodies:

```cpp
void ConflictView::resized()
{
    auto content { getLocalBounds() };

    for (size_t index { 0 }; index < versions.size(); ++index)
    {
        auto column { content.removeFromLeft (content.getWidth() / static_cast<int> (versions.size() - index)) };

        labels.at (index).setBounds (column.removeFromTop (labelRows));
        versions.at (index).setBounds (column);
    }
}

void ConflictView::focusGained (FocusChangeType)
{
    setFocus();
}

void ConflictView::setFocus()
{
    const auto pane { static_cast<int> (state.getProperty (Id::focusedPane)) };

    if (pane == map::ConflictPane::files)
        grabKeyboardFocus();
    else
        versions.at (static_cast<size_t> (pane - map::ConflictPane::base)).grabKeyboardFocus();
}

juce::ValueTree ConflictView::getFile() const
{
    const auto conflict { state.getChildWithName (jam::Format::toType (Id::conflict)) };

    return conflict.getChild (static_cast<int> (conflict.getProperty (Id::selectedIndex)));
}

void ConflictView::setLabels()
{
    const auto file { getFile() };

    for (size_t index { 0 }; index < labels.size(); ++index)
        labels.at (index).setText (file.getChild (static_cast<int> (index)).getProperty (Id::value).toString(), juce::dontSendNotification);
}

void ConflictView::valueTreePropertyChanged (juce::ValueTree&, const juce::Identifier& property)
{
    const auto isShown { static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::conflictResolve };

    if (isShown and (property == Id::selectedIndex or property == Id::mode))
        setLabels();

    if (property == Id::focusedPane and hasKeyboardFocus (true))
        setFocus();
}
```

   - Each version is one column: a `juce::Label` with the label of the version, and under it a `ConsoleView` that reads the swap of the run chain of that version (Step 7). The three columns share the content, as the legacy rows do (`legacy/ui/conflictresolver.go:50-142`). Each column takes its share of the width that remains, so the last column takes every cell that is left. The menu of the files shows over the content, at the place that `MenuWindow` gives it (Part 11 Step 10, item 3). This is a layout (ruling 7).
   - The label is a stock `juce::Label`, and the LookAndFeel paints it. Its text is the row `value` of the version: the label of the side, or the text of a version that does not exist (Step 7, item 3).
   - The panes show the selected file. `TITView` tells the processor to print its three versions (Step 11, item 8), and this view sets the three labels, when the mode starts and when the selection changes. The job of Step 7 ends with one change message of `selectedIndex` (Part 11 Step 7, `setChildren`), so the panes fill when the files arrive. The highlighted file becomes the selected file at the next key that the menu does not take, for example Tab (Part 11 Step 11, item 5, `keyPressed`).
   - The row `files` of `map::ConflictPane` has the value 0, as `records` of `map::Pane` has. So `isRecordMode()` of Part 11 is true when the menu of the files has the keys.
   - The pane with the focus is a version (Closed by a read, item 7). The focus moves only when the view holds it (Part 11 Step 10, item 5).

5. `BranchView`, with the includes of Part 10 Step 11 and `#include "TextView.h"`:

```cpp
class BranchView : public juce::Component,
                   public jam::Model::Component<BranchView>,
                   private juce::ValueTree::Listener
{
public:
    explicit BranchView (juce::ValueTree stateToUse);
    ~BranchView() override = default;

    void resized() override;
    void focusGained (FocusChangeType cause) override;

    juce::ValueTree getBranch() const;

    static constexpr int numPanes { 2 };

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    void setFocus();

    TextView details;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (BranchView)
};
```

   - The constructor: the form of Part 10 Step 11, item 2. Then `setWantsKeyboardFocus (true);` and `addAndMakeVisible (details);`.
   - The bodies:

```cpp
void BranchView::resized()
{
    details.setBounds (getLocalBounds());
}

void BranchView::focusGained (FocusChangeType)
{
    setFocus();
}

void BranchView::setFocus()
{
    if (static_cast<int> (state.getProperty (Id::focusedPane)) == map::Pane::records)
        grabKeyboardFocus();
    else
        details.grabKeyboardFocus();
}

juce::ValueTree BranchView::getBranch() const
{
    const auto branches { state.getChildWithName (jam::Format::toType (Id::branches)) };

    return branches.getChild (static_cast<int> (branches.getProperty (Id::selectedIndex)));
}

void BranchView::valueTreePropertyChanged (juce::ValueTree&, const juce::Identifier& property)
{
    const auto isShown { static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::branchPicker };

    if (isShown and (property == Id::selectedIndex or property == Id::mode))
        details.setText (getBranch().getProperty (Id::text).toString());

    if (property == Id::focusedPane and hasKeyboardFocus (true))
        setFocus();
}
```

   - The details take the content, and the menu of the branches shows over it (`legacy/ui/branchpicker.go:40-66`; ruling 7). Tab goes between the menu and the details (`tit/SPEC.md:428`; legacy fault 11). The details show the selected branch, which is the highlighted branch at the last key that the menu did not take.

6. `InputView`: change `setDocument()` of Part 10 Step 14 to `void setDocument (const juce::String& text);`. Its first statement becomes `const auto bytes { (text + juce::String::charToString (Chars::newline)).toStdString() };`, and `addLines` takes `bytes.data()` and `bytes.size()`. So the input opens with the text, or with one empty line. The call sites of Parts 10 and 11 pass `{}`: `setInput` and `cancelInput` of `TITView`, and `HistoryView::setSearchDocument`. The reword of Step 12 passes the message of a commit.

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- No member holds state. Each view reads its tree and the rows `mode` and `focusedPane`.
- Neither view names a command id, a command line, a menu, or a processor function. `ConflictView` reads one constant of the processor, the count of the versions, and takes the getters of the three chains.
- `TextView` holds no text. Its owner gives the text from a Model row. No view makes a document from output of the tool.
- New names and their families: `TextView` (the `…View` family; the row `text` and `map::Pane::text`); `setText` (`juce::Label::setText`, JUCE; NAMES.md Rule 1, `set`); `ConflictView`, `BranchView` (the `…View` family; the trees `conflict` and `branches`); `versions`, `labels`, `chains` (`TITGitProcessor::numVersions`, Step 7; the row type `version`; the legacy `ColumnLabels`); `details` (the legacy details pane); `getFile`, `getBranch` (`getCommit`, `getFile` of `HistoryView`; the row types); `setLabels`, `setFocus`, `numPanes`, `labelRows` (`HistoryView`, Part 11 Step 10; `searchRows`). Local names: `content`, `column`, `index`, `pane`, `file`, `conflict`, `branches`, `isShown`.

### Step 11: `TITView`, the header and the shared pieces
**Scope:** `tit/Source/TITView.h`, `tit/Source/TITView.cpp`.
**Action:** The functions of Part 11 Step 11 stay, except where an item below says "replace".

1. The header. Add `#include "ConflictView.h"` and `#include "BranchView.h"` after `"HistoryView.h"`. Then:
   - Private functions, after `registerSearchActions` of Part 11: `void registerDirtyActions();`, `void registerTimeTravelActions();`, `void registerPinActions();`, `void registerUnpinActions();`, `void registerConflictActions();`, `void registerBranchActions();`, `void registerBranchSelectionActions();`, `void registerBranchMergeActions();`, `void registerBranchRebaseActions();`, `void registerRebaseActions();`, `void registerRebaseListActions();`, `void registerRewordActions();`.
   - Private functions, after `copyText` of Part 11:

```cpp
    void setDirtyConfirmation (const juce::String& title, const juce::String& explanation, const juce::String& yes, const juce::String& no,
                               const juce::StringArray& changeSet, const juce::StringArray& noCommandLines, juce::StringPairArray markers = {});
    void setReturnConfirmation (const juce::String& branch, const juce::String& yes);
    void setVersions();
```

   - Private functions, after `getInitialItemId` of Part 11:

```cpp
    juce::ValueTree getSelected (const juce::Identifier& tree) const;
    juce::StringArray getConflictLabels() const;
    bool isOperation (std::initializer_list<int> operations) const;
    bool isDirty() const;

    static bool isOtherBranch (const juce::ValueTree& branch);
    static juce::String getFormattedText (juce::String text, const juce::StringArray& values);
    static juce::String getMarkerText (const juce::StringArray& lines);
```

   - Members, after `history` of Part 11: `ConflictView conflict;`, `BranchView branches;`.

2. The constructor: in the initializer list, after `history`, add `, conflict (modelToUse.state, gitProcessorToUse.versions)` and `, branches (modelToUse.state)`. After `addChildComponent (history);`, add `addChildComponent` for `conflict` and `branches`. In `registerActions()`, add at its end the twelve calls of item 1, in that order.

3. The small readers and the text:

```cpp
bool TITView::isOperation (std::initializer_list<int> operations) const
{
    const auto operation { static_cast<int> (getGit().getProperty (Id::operation)) };

    return std::find (operations.begin(), operations.end(), operation) != operations.end();
}

bool TITView::isDirty() const
{
    return static_cast<int> (getGit().getProperty (Id::workingTree)) == map::WorkingTree::dirty;
}

juce::ValueTree TITView::getSelected (const juce::Identifier& tree) const
{
    const auto parent { state.getChildWithName (jam::Format::toType (tree)) };

    return parent.getChild (static_cast<int> (parent.getProperty (Id::selectedIndex)));
}

bool TITView::isOtherBranch (const juce::ValueTree& branch)
{
    return branch.isValid() and not static_cast<bool> (branch.getProperty (Id::isCurrent));
}

juce::String TITView::getFormattedText (juce::String text, const juce::StringArray& values)
{
    for (const auto& value : values)
        text = text.replaceFirstOccurrenceOf (text::Reading::hashPlaceholder, value);

    return text;
}

juce::StringArray TITView::getConflictLabels() const
{
    const auto git { getGit() };
    const juce::StringArray branch { git.getProperty (Id::currentBranch).toString() };

    if (static_cast<bool> (git.getProperty (Id::isMerging)))
        return { text::Conflict::base, getFormattedText (text::Conflict::current, branch), getFormattedText (text::Conflict::incoming, { git.getProperty (Id::mergeName).toString() }) };

    if (static_cast<bool> (git.getProperty (Id::isRebasing)))
        return { text::Conflict::base, text::Conflict::ours, text::Conflict::theirs };

    return { text::Conflict::base, getFormattedText (text::Conflict::merged, branch), text::Conflict::stashed };
}
```

   - `getFormattedText` puts values into a text that the user reads: a title, an explanation, a label. It makes no command line. The command lines come from the processor (Step 9).
   - `getSelected` reads the selected row of the `rebase` tree. The selected rows of the resolver and of the picker come from their views (`ConflictView::getFile`, `BranchView::getBranch`).
   - `isOtherBranch` answers if a row of the picker is a branch other than the current one. The four commands of the picker that act on the selected branch act only on such a row (`tit/SPEC.md:431-432`; Closed by a read, item 17). The footer shows the same two cases with `branchPickerOther` and `branchPickerCurrent` (Step 8).

4. The marker and the two dialogs:

```cpp
juce::String TITView::getMarkerText (const juce::StringArray& lines)
{
    const auto newline { juce::String::charToString (Chars::newline) };

    return lines.joinIntoString (newline) + newline;
}

void TITView::setDirtyConfirmation (const juce::String& title, const juce::String& explanation, const juce::String& yes, const juce::String& no,
                                    const juce::StringArray& changeSet, const juce::StringArray& noCommandLines, juce::StringPairArray markers)
{
    markers.set (files::dirtyOperationMarker, getMarkerText (TITGitProcessor::getDirtyMarkerLines (getGit())));
    setConfirmation (title, explanation, yes, no, TITGitProcessor::getDirtyLines (changeSet), noCommandLines, markers);
}

void TITView::setReturnConfirmation (const juce::String& branch, const juce::String& yes)
{
    const auto git { getGit() };

    setConfirmation (getFormattedText (text::Confirmation::timeTravelReturnDirtyTitle, { branch }),
                     text::Confirmation::timeTravelReturnDirtyExplanation, yes, text::Confirmation::timeTravelReturnDirtyNo,
                     TITGitProcessor::getReturnMergeLines (git, branch),
                     TITGitProcessor::getDiscardLines (TITGitProcessor::getReturnLines (git, branch)));
}
```

   - The markers of a dialog are a `juce::StringPairArray`: the file name of each marker and its text. The caller gives them to `setConfirmation`, and the callback of the box holds them by value (item 5). No Model row holds them (ruling 65). `getMarkerText` makes the text of a marker file from its lines.
   - `setReturnConfirmation` is the dialog of a return with a dirty tree. Its first button is the merge of `tit/SPEC.md:557-593`. Its second button is the legacy discard: `reset --hard HEAD`, `clean -fd`, then the return.

5. The markers of a dialog. In `setConfirmation` of Part 10 Step 16, item 9, add a last parameter `const juce::StringPairArray& markers = {}` in the declaration and `const juce::StringPairArray& markers` in the definition. Add `markers` to the capture list of the callback, and replace the last statement of the callback:

```cpp
        if (const auto& commandLines { result == firstButton ? yesCommandLines : noCommandLines }; not commandLines.isEmpty())
        {
            for (const auto& fileName : markers.getAllKeys())
                gitProcessor.setMarker (fileName, markers[fileName]);

            gitProcessor.run (commandLines, {});
        }
```

   - A dialog with no marker has an empty list, so the loop writes nothing for the dialogs of Parts 10 and 11.
   - The second button of a dirty dialog discards the changes. The dirty marker is then written with no stash; it is stale, and the detection after the run removes it (Step 6, item 5).

6. The mode of the `Conflicted` operation. In `setMenuMode` of Part 9 Step 7, item 3, change the entry of `conflicted` (the fourth entry) to `map::EditorMode::conflictResolve`. Replace the `Id::operation` and `Id::mode` entries of `registerEvents` of Part 10:

```cpp
    events.add (Id::operation, [this] (const juce::Identifier&, juce::ValueTree& git)
    {
        if (isMenuMode() or static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::conflictResolve)
            setMenuMode (git);

        setKeys (git);
    });

    events.add (Id::mode, [this] (const juce::Identifier&, juce::ValueTree&)
    {
        if (static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::conflictResolve)
            gitProcessor.readConflicts (getConflictLabels());

        setView();
        setKeys (getGit());
        setMenu (getGit());
        setPatch();
        setVersions();
    });
```

   - The resolver asks for its files when its mode starts. The views fill their own panes (Step 10).
   - Each of the two events builds the key sheet, as in Part 10 Step 16, item 3. The `mode` event builds the menu too.
   - In the loop over the axis rows of Part 10 Step 16, item 3, add `Id::originalHead`: the rule of `dirtyAbort` reads that row (Step 12, item 2).
   - `isEditMode` of Part 11 Step 11, item 4: add `map::EditorMode::branchName` and `map::EditorMode::rebaseReword` to `editModes`. Each one shows an editable editor.

7. The focus and the views. Replace `getFocusComponent` of Part 11. In `resized`, after `history.setBounds (content);`, add `conflict.setBounds (content);` and `branches.setBounds (content);`.

```cpp
juce::Component& TITView::getFocusComponent()
{
    const jam::HashMap<int, juce::Component*> components {
        { map::EditorMode::console, &console },
        { map::EditorMode::input, &commitMessage },
        { map::EditorMode::branchName, &commitMessage },
        { map::EditorMode::rebaseReword, &commitMessage },
        { map::EditorMode::history, &history },
        { map::EditorMode::search, &history },
        { map::EditorMode::conflictResolve, &conflict },
        { map::EditorMode::branchPicker, &branches },
    };

    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    if (components.contains (mode)) return *components.at (mode);
    return *this;
}
```

   In `setView`, after the line of Part 11:

```cpp
    conflict.setVisible (isLarge and mode == map::EditorMode::conflictResolve);
    branches.setVisible (isLarge and mode == map::EditorMode::branchPicker);
```

   Change the `commitMessage.setVisible` line of Part 10 to `commitMessage.setVisible (isLarge and (mode == map::EditorMode::input or mode == map::EditorMode::branchName or mode == map::EditorMode::rebaseReword));`.

   - The rebase list is the menu of the `rebase` tree, with no view under it. In its mode `TITView` holds the focus, and the menu is modal.
   - Each command that opens a view sets `focusedPane` to 0 first, so the menu of its records has the keys.

8. The menus of the records, Return, and Tab. Add three entries to `registerRecords` of Part 11 Step 11, item 5, and replace the body of `registerSelections`:

```cpp
    for (const auto& [mode, tree] : std::initializer_list<std::pair<int, juce::Identifier>> { { map::EditorMode::conflictResolve, Id::conflict },
                                                                                             { map::EditorMode::branchPicker, Id::branches },
                                                                                             { map::EditorMode::rebaseList, Id::rebase } })
        records.add (mode, [this, tree] (const int&) { return state.getChildWithName (jam::Format::toType (tree)); });
```

```cpp
void TITView::registerSelections()
{
    selections.add (map::EditorMode::history, [this] (const int&, const juce::ValueTree& record)
    {
        if (record.hasType (jam::Format::toType (Id::file)))
            state.setProperty (Id::focusedPane, static_cast<int> (map::Pane::text), nullptr);
        else
            commandManager->invokeDirectly (map::CommandId::timeTravel, false);
    });

    for (const auto& [mode, command] : std::initializer_list<std::pair<int, juce::CommandID>> { { map::EditorMode::conflictResolve, map::CommandId::conflictEnter },
                                                                                              { map::EditorMode::branchPicker, map::CommandId::branchPickerEnter },
                                                                                              { map::EditorMode::rebaseList, map::CommandId::rebaseApply } })
        selections.add (mode, [this, command] (const int&, const juce::ValueTree&) { commandManager->invokeDirectly (command, false); });
}
```

   Replace the `historyTab` entry of `registerBrowserActions` of Part 11:

```cpp
    commands.add (map::CommandId::historyTab, [this] (const juce::CommandID&)
    {
        static const jam::HashMap<int, int> paneCounts {
            { map::EditorMode::history, HistoryView::numPanes },
            { map::EditorMode::conflictResolve, ConflictView::numPanes },
            { map::EditorMode::branchPicker, BranchView::numPanes },
        };

        state.setProperty (Id::focusedPane, (static_cast<int> (state.getProperty (Id::focusedPane)) + 1) % paneCounts.at (static_cast<int> (state.getProperty (Id::mode))), nullptr);
    });
```

   In `registerEvents`, add the entry `events.add (Id::attribute, [this] (const juce::Identifier&, juce::ValueTree&) { setMenu (getGit()); });`. In the lambda of the loop of Part 11 Step 11, item 4, add the statement `setVersions();`. Then:

```cpp
void TITView::setVersions()
{
    if (static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::conflictResolve)
    {
        juce::StringArray patchLines;

        for (const auto& version : conflict.getFile())
            patchLines.add (version.getProperty (Id::patch).toString());

        gitProcessor.readVersions (patchLines);
    }
}
```

   - Each tree of records of this part is a permanent child of the root (Step 5), and `records` names it for its mode. `setMenu` of Part 11 then shows its menu with no new code (ruling 60).
   - `selections` holds what Return does on a record of a mode (Part 11 Step 11, item 5). The action of the item selects the record first. Then the entry invokes the command of the legacy Enter of the screen: time travel on a commit of the browser, the resolve of the resolver, the switch of the picker, and the apply of the rebase list. Return on a file of the browser still gives the keys to the text pane. Each command reads the selected record from the Model rows (Step 12), so Return on an item and the `return` row of a text pane run the same command.
   - No command of this part moves a highlight. The menu does it (ruling 61).
   - `historyTab` goes to the next pane of the screen that shows. The pane 0 of each screen is its menu (Step 10, item 3).
   - `setVersions` tells the processor to print the three versions of the selected file of the resolver, in the form of `setPatch` of Part 11. It runs at a new selection and when the mode changes, and only while the resolver shows. The view gives the `patch` lines that the versions hold; it makes none.
   - A new `attribute` of a record is a new item text: the chosen version of a file of the resolver, and the action of a row of the rebase list (Step 12). The event builds the menu again. A row of the rebase list that moves writes a new `selectedIndex` (Step 12), and that event builds the menu.

9. The menu. In `getMenuCommands` and `setMenu` of Part 11 Step 11, item 5:
   - `operationCommands`, in this order: `commit`, `commitPush`, `resetDiscardChanges`, `viewChanges`, `push`, `pullMerge`, `pullRebase`, `dirtyPullMerge`, `forcePush`, `replaceLocal`, `incomingOutgoing`, `finalizeMerge`, `abortMerge`, `rebaseContinue`, `rebaseAbort`, `dirtyContinue`, `dirtyAbort`, `history`, `fileHistory`, `timeTravelReturn`, `unpin`.
   - Replace `isNormal` with `const auto isBrowsable { isOperation ({ map::Operation::normal, map::Operation::timeTraveling, map::Operation::pinned }) };`.
   - `registerMenuKeys()`: add two entries. `history` gives `Id::timeTravelHistory` when the operation is `timeTraveling`, else `commandId->get (map::CommandId::history)`. `fileHistory` gives `Id::timeTravelFilesHistory` in the same way.

10. `registerAvailability()`: change these rules of Parts 9 to 11.

   | Command | New rule |
   |---|---|
   | `abortMerge` | `operation` is `merging` |
   | `history`, `fileHistory` | `operation` is `normal`, `timeTraveling`, or `pinned`, and `isCacheReady` |
   | `resetDiscardChanges` | `workingTree` is `dirty`, and `operation` is `normal` or `pinned` |

   In the `resetDiscardChanges` entry of `registerWorkingTreeActions` of Part 10 Step 16, item 7, the condition of the remote choice becomes: the remote is `hasRemote` and `isOperation ({ map::Operation::normal })` (Closed by a read, item 28).

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- The order of `operationCommands` is the legacy order of each menu (`legacy/app/menu_render_core.go:92-154`; `legacy/app/menu_render_extra.go:63-147`), with `pullRebase` after `pullMerge` (Part 2 Step 4, item 4).
- The pinned menu shows `history`, `fileHistory`, `resetDiscardChanges` when `Dirty`, and `unpin` (ruling 8). No timeline command is active while `Pinned`: each needs `normal`.
- `TITView` holds no editor and no pane table, makes no command line, and writes no file. Each pane gets its tree or its text from the Model, in its own view.
- New names and their families: the twelve `register…Actions` (the `register…Actions()` family, `end/Source/ui/ENDView.h:121-179`; the words `dirty`, `timeTravel`, `pin`, `unpin`, `conflict`, `branch`, `rebase`, `reword` of the commands of Step 2); `isOtherBranch` (`isDirty`; the text rows `branchPickerOther` and `branchPickerCurrent`); `getMarkerText`, `markers` (`getFormattedText`; `juce::StringPairArray`; the marker files of Part 8; NAMES.md Rule 1, `add`); `setDirtyConfirmation`, `setReturnConfirmation` (`setConfirmation`; the menu label `Return`); `getSelected` (`Id::selectedIndex`; `getCommit`); `getConflictLabels` (the legacy `ColumnLabels`); `isOperation`, `isDirty` (`isMenuMode`; the axis words); `getFormattedText` (`juce::String::formatted`); `conflict`, `branches` (the trees of Step 5). Local names: `tree`, `command`, `record`, `operation`, `parent`, `value`, `git`, `branch`, `newline`, `marker`, `components`, `mode`, `paneCounts`, `isBrowsable`.

### Step 12: `TITView`, the command families of the protocols
**Scope:** `tit/Source/TITView.cpp`.
**Action:** Each family is one function of 30 lines or fewer, in the END form (`end/Source/ui/ENDView.cpp:163-400`). Each entry takes its command lines from the processor (Step 9), its selected row from a view (Step 10; Part 11 Step 10), and its texts from the cast tables. It tells the processor to run, or it opens a dialog.

1. The dirty operation, time travel, and pin:

```cpp
void TITView::registerDirtyActions()
{
    commands.add (map::CommandId::dirtyPullMerge, [this] (const juce::CommandID&)
    {
        setDirtyConfirmation (text::Confirmation::dirtyPullTitle, text::Confirmation::dirtyPullExplanation, text::Confirmation::dirtyPullYes, text::Confirmation::dirtyPullNo,
                              { text::Git::pullMerge }, TITGitProcessor::getDiscardLines ({ text::Git::pullMerge }));
    });

    commands.add (map::CommandId::dirtyContinue, [this] (const juce::CommandID&) { gitProcessor.run (TITGitProcessor::getStashLines (getGit()), {}); });
    commands.add (map::CommandId::dirtyAbort, [this] (const juce::CommandID&) { gitProcessor.run (TITGitProcessor::getDirtyAbortLines (getGit()), {}); });
}

void TITView::registerTimeTravelActions()
{
    commands.add (map::CommandId::timeTravel, [this] (const juce::CommandID&)
    {
        const auto commit { history.getCommit() };
        const auto hash { commit.getProperty (Id::id).toString() };

        if (hash.isNotEmpty())
        {
            juce::StringPairArray markers;

            if (not isOperation ({ map::Operation::timeTraveling }))
                markers.set (files::timeTravelMarker, getMarkerText (TITGitProcessor::getTimeTravelMarkerLines (getGit())));

            setConfirmation (text::Confirmation::timeTravelTitle,
                             getFormattedText (text::Confirmation::timeTravelExplanation,
                                               { hash.substring (0, shortHashLength), commit.getProperty (Id::value).toString().fromFirstOccurrenceOf (juce::String::charToString (Chars::space), false, false) }),
                             text::Confirmation::timeTravelYes, text::Confirmation::timeTravelNo, TITGitProcessor::getTimeTravelLines (getGit(), hash), {}, markers);
        }
    });

    commands.add (map::CommandId::timeTravelReturn, [this] (const juce::CommandID&)
    {
        const auto git { getGit() };
        const auto branch { git.getProperty (Id::currentBranch).toString() };
        const auto isStashMissing { static_cast<bool> (git.getProperty (Id::hasOriginalStash)) and git.getProperty (Id::titStash).toString().isEmpty() };

        if (static_cast<bool> (git.getProperty (Id::isTitTimeTravel)))
        {
            if (isStashMissing)
                setConfirmation (text::Confirmation::stashNotFoundTitle, getFormattedText (text::Confirmation::stashNotFoundExplanation, { text::Git::stashLabel }),
                                 text::Confirmation::stashNotFoundYes, text::Confirmation::stashNotFoundNo,
                                 isDirty() ? TITGitProcessor::getReturnMergeLines (git, branch) : TITGitProcessor::getReturnLines (git, branch), {});
            else if (isDirty())
                setReturnConfirmation (branch, text::Confirmation::timeTravelReturnDirtyYes);
            else
                setConfirmation (getFormattedText (text::Confirmation::timeTravelReturnTitle, { branch }), text::Confirmation::timeTravelReturnExplanation,
                                 text::Confirmation::timeTravelReturnYes, text::Confirmation::timeTravelReturnNo, TITGitProcessor::getReturnLines (git, branch), {});
        }
        else
        {
            commandManager->invokeDirectly (map::CommandId::branchPicker, false);
        }
    });
}

void TITView::registerPinActions()
{
    commands.add (map::CommandId::pin, [this] (const juce::CommandID&)
    {
        const auto git { getGit() };
        const auto hash { history.getCommit().getProperty (Id::id).toString() };
        const juce::StringArray changeSet { TITGitProcessor::getLine ({ text::Git::resetHard, hash }) };

        if (hash.isNotEmpty())
        {
            juce::StringPairArray markers;

            markers.set (files::pinMarker, getMarkerText (TITGitProcessor::getPinMarkerLines (git, hash)));

            if (isDirty())
                setDirtyConfirmation (text::Confirmation::pinDirtyTitle, text::Confirmation::pinDirtyExplanation, text::Confirmation::pinDirtyYes, text::Confirmation::pinDirtyNo,
                                      changeSet, TITGitProcessor::getDiscardLines (changeSet), markers);
            else
                setConfirmation (text::Confirmation::pinTitle, getFormattedText (text::Confirmation::pinExplanation, { git.getProperty (Id::currentBranch).toString(), hash.substring (0, shortHashLength) }),
                                 text::Confirmation::pinYes, text::Confirmation::pinNo, changeSet, {}, markers);
        }
    });
}

void TITView::registerUnpinActions()
{
    commands.add (map::CommandId::unpin, [this] (const juce::CommandID&)
    {
        const auto git { getGit() };
        const auto original { git.getProperty (Id::pinOriginalHead).toString() };
        const juce::StringArray changeSet { TITGitProcessor::getLine ({ text::Git::resetHard, original }) };

        if (isDirty())
            setConfirmation (text::Confirmation::unpinDirtyTitle, text::Confirmation::unpinDirtyExplanation, text::Confirmation::unpinDirtyYes, text::Confirmation::unpinDirtyNo,
                             TITGitProcessor::getDirtyLines (changeSet), TITGitProcessor::getDiscardLines (changeSet));
        else
            setConfirmation (text::Confirmation::unpinTitle,
                             getFormattedText (text::Confirmation::unpinExplanation, { git.getProperty (Id::currentBranch).toString(), original.substring (0, shortHashLength) }),
                             text::Confirmation::unpinYes, text::Confirmation::unpinNo, changeSet, {});
    });
}
```

   Add two entries at the end of `registerPullActions()` of Part 10 Step 16, item 8:

```cpp
    commands.add (map::CommandId::addRemote, [this] (const juce::CommandID&) { commandManager->invokeDirectly (map::CommandId::pullRebase, false); });

    commands.add (map::CommandId::pullRebase, [this] (const juce::CommandID&)
    {
        setConfirmation (text::Confirmation::pullRebaseTitle, text::Confirmation::pullRebaseExplanation, text::Confirmation::pullRebaseYes, text::Confirmation::pullRebaseNo,
                         { text::Git::pullRebase }, {});
    });
```

2. `registerAvailability()`: add these entries in the form of Part 11 Step 11, item 6.

   | Command | Rule |
   |---|---|
   | `dirtyPullMerge` | `operation` is `normal`, `remote` is `hasRemote`, `workingTree` is `dirty`, and `timeline` is `behind` or `diverged` |
   | `dirtyContinue` | `operation` is `dirtyOperation` |
   | `dirtyAbort` | `operation` is `dirtyOperation`, and the `originalHead` row is not empty |
   | `timeTravel` | `operation` is `normal` or `timeTraveling` |
   | `timeTravelReturn` | `operation` is `timeTraveling` |
   | `pin` | `operation` is `normal` or `pinned` |
   | `unpin` | `operation` is `pinned` |
   | `pullRebase`, `addRemote` | `operation` is `normal`, `remote` is `hasRemote`, `workingTree` is `clean`, and `timeline` is `behind` or `diverged` |

   - Part 13 adds the branch of `addRemote` with no remote, and widens its rule (ruling 35).
   - `timeTravel` and `pin` act on the hash of the selected commit. A record with no hash gives no dialog (ruling 40; Part 11 Step 11, item 6).
   - The pin and the unpin are two families of one command each, as `registerLoadPluginActions` of END is (`end/Source/ui/ENDView.cpp:354-362`).

- **The dirty operation.** "Reapply changes" applies the stash of TIT and drops it. "Abort and restore" is `tit/SPEC.md:376-385`. Each one names the stash of the row. After the drop, no stash of TIT exists, the marker is stale, and the detection removes it and gives `Normal`.
- **A chain that stops.** `run` ends at the first line that fails, and the detection runs (Part 9 Step 5). Unmerged files give `Conflicted` and the resolver. A stash of TIT with no conflict gives `DirtyOperation` and its menu. So each stop has its commands, with no phase.
- **Time travel.** The dialog, its first button, and the lines are legacy. The marker holds the branch. A jump from one commit to another during a journey writes no marker and no stash: the branch of the marker stays.
- **The return.** With the marker: the dialog of the clean return, or `setReturnConfirmation`. With no marker: the branch picker (`tit/SPEC.md:214-216`).
- **Pin.** The detection of Part 8 gives `Pinned` only while HEAD is the pinned hash.
- **Unpin.** The reset moves HEAD from the pinned hash, so the marker is stale and the detection removes it.
- **A dirty tree during a pin or an unpin** is the protocol with the reset as its change-set (Part 2 Step 7). The unpin writes no dirty marker: its abort would reset the branch to the pinned hash after the detection has removed the pin marker, and the HEAD before the pin would be lost. So an unpin that stops has "Reapply changes" only.
- The selected button of each dialog of Part 2: the first for `pin`, `pin_dirty`, `unpin`, `unpin_dirty`, as their siblings `time_travel` and `dirty_pull`; the second for `pull_rebase`, `rebase_branch`, `rebase_branch_dirty`, `rebase_interactive`, as their siblings `pull_merge`, `merge_branch`, `merge_branch_dirty`.

3. The resolver:

```cpp
void TITView::registerConflictActions()
{
    commands.add (map::CommandId::conflictSpace, [this] (const juce::CommandID&)
    {
        const auto version { static_cast<int> (state.getProperty (Id::focusedPane)) - map::ConflictPane::base };
        auto file { conflict.getFile() };

        if (file.isValid() and static_cast<int> (state.getProperty (Id::focusedPane)) != map::ConflictPane::files)
        {
            const auto label { file.getChild (version).getProperty (Id::value).toString() };

            file.setProperty (Id::attribute, label, nullptr);
            file.setProperty (Id::selectedIndex, version, nullptr);
            state.setProperty (Id::message, getFormattedText (text::Footer::marked, { file.getProperty (Id::id).toString(), label }), nullptr);
        }
    });

    commands.add (map::CommandId::conflictEnter, [this] (const juce::CommandID&)
    {
        const auto conflictFiles { state.getChildWithName (jam::Format::toType (Id::conflict)) };
        const auto isResolved { conflictFiles.getNumChildren() > 0
                                and std::all_of (conflictFiles.begin(), conflictFiles.end(), [] (const juce::ValueTree& file)
                                {
                                    return static_cast<int> (file.getProperty (Id::selectedIndex)) != TITGitProcessor::noVersion;
                                }) };

        if (isResolved)
            gitProcessor.run (TITGitProcessor::getConflictLines (getGit(), conflictFiles), {});
    });
}
```

   In `registerCancellations()`, add:

```cpp
    cancellations.add (map::EditorMode::conflictResolve, [this] (const int&) { gitProcessor.run (TITGitProcessor::getConflictAbortLines (getGit()), {}); });
    cancellations.add (map::EditorMode::branchName, [this] (const int&) { state.setProperty (Id::mode, static_cast<int> (map::EditorMode::branchPicker), nullptr); });
    cancellations.add (map::EditorMode::rebaseList, [this] (const int&) { state.setProperty (Id::mode, static_cast<int> (map::EditorMode::history), nullptr); });
    cancellations.add (map::EditorMode::rebaseReword, [this] (const int&) { state.setProperty (Id::mode, static_cast<int> (map::EditorMode::rebaseList), nullptr); });
```

   - Space sets the chosen version of the selected file and its label in the list. The pane with the focus is the version (Closed by a read, item 7). Space acts on a file and a version pane: with the focus on the file list, it has no version to set.
   - Enter runs the lines of each chosen version and the finalize of the conflict type, as one run (Step 9, item 3). Enter acts on a complete choice: each file has a version, as in legacy (`legacy/app/conflict_handlers.go:119-136`).
   - A rebase that stops at its next commit gives `Conflicted` again after the run, and the resolver opens again with the new files: the loop of `tit/SPEC.md:159-163`.
   - After the merge commit of a conflict inside a dirty operation, the stash of TIT is still there. The detection gives `DirtyOperation`, and "Reapply changes" continues.
   - Escape aborts by the type (Step 9, item 3).
   - Escape in the picker returns to the menu through the default of `cancel` (Part 11 Step 11, item 8).

4. The picker, in four families: the picker and a new branch; the switch and the delete; the merge; the rebase. Each command of the last three acts on the selected branch, and only when `isOtherBranch` holds for it (Step 11, item 3).

```cpp
void TITView::registerBranchActions()
{
    commands.add (map::CommandId::branchPicker, [this] (const juce::CommandID&)
    {
        gitProcessor.readBranches();
        state.setProperty (Id::focusedPane, 0, nullptr);
        state.setProperty (Id::mode, static_cast<int> (map::EditorMode::branchPicker), nullptr);
    });

    commands.add (map::CommandId::branchPickerAdd, [this] (const juce::CommandID&)
    {
        commitMessage.setDocument ({});
        state.setProperty (Id::mode, static_cast<int> (map::EditorMode::branchName), nullptr);
        state.setProperty (Id::message, text::Branch::newBranchName, nullptr);
    });

    commands.add (map::CommandId::newBranchNameSubmit, [this] (const juce::CommandID&)
    {
        const auto name { commitMessage.getText() };

        if (name.isEmpty())
            state.setProperty (Id::message, text::Footer::branchNameEmpty, nullptr);
        else
            gitProcessor.run ({ TITGitProcessor::getLine ({ text::Git::checkBranchName, name }), TITGitProcessor::getLine ({ text::Git::createBranch, name }) }, {});
    });
}

void TITView::registerBranchSelectionActions()
{
    commands.add (map::CommandId::branchPickerEnter, [this] (const juce::CommandID&)
    {
        const auto name { branches.getBranch().getProperty (Id::id).toString() };
        const auto isReturn { isOperation ({ map::Operation::timeTraveling }) };
        const juce::StringArray changeSet { TITGitProcessor::getLine ({ text::Git::switchBranch, name }) };

        if (isOtherBranch (branches.getBranch()))
        {
            if (not isDirty())
                gitProcessor.run (isReturn ? TITGitProcessor::getReturnLines (getGit(), name) : changeSet, {});
            else if (isReturn)
                setReturnConfirmation (name, text::Confirmation::timeTravelReturnDirtyStash);
            else
                setDirtyConfirmation (getFormattedText (text::Confirmation::branchSwitchDirtyTitle, { name }), text::Confirmation::branchSwitchDirtyExplanation,
                                      text::Confirmation::branchSwitchDirtyYes, text::Confirmation::branchSwitchDirtyNo, changeSet, TITGitProcessor::getDiscardLines (changeSet));
        }
    });

    commands.add (map::CommandId::branchPickerDelete, [this] (const juce::CommandID&)
    {
        const auto name { branches.getBranch().getProperty (Id::id).toString() };

        if (isOtherBranch (branches.getBranch()))
            setConfirmation (getFormattedText (text::Confirmation::branchDeleteTitle, { name }), text::Confirmation::branchDeleteExplanation,
                             text::Confirmation::branchDeleteYes, text::Confirmation::branchDeleteNo, { TITGitProcessor::getLine ({ text::Git::deleteBranch, name }) }, {});
    });
}

void TITView::registerBranchMergeActions()
{
    commands.add (map::CommandId::branchPickerMerge, [this] (const juce::CommandID&)
    {
        const auto name { branches.getBranch().getProperty (Id::id).toString() };
        const juce::StringArray changeSet { TITGitProcessor::getLine ({ text::Git::merge, name }) };

        if (isOtherBranch (branches.getBranch()))
        {
            if (isDirty())
                setDirtyConfirmation (text::Confirmation::mergeBranchDirtyTitle, text::Confirmation::mergeBranchDirtyExplanation,
                                      text::Confirmation::mergeBranchDirtyYes, text::Confirmation::mergeBranchDirtyNo, changeSet, {});
            else
                setConfirmation (text::Confirmation::mergeBranchTitle,
                                 getFormattedText (text::Confirmation::mergeBranchExplanation, { name, getGit().getProperty (Id::currentBranch).toString() }),
                                 text::Confirmation::mergeBranchYes, text::Confirmation::mergeBranchNo, changeSet, {});
        }
    });
}

void TITView::registerBranchRebaseActions()
{
    commands.add (map::CommandId::branchPickerRebase, [this] (const juce::CommandID&)
    {
        const auto selected { branches.getBranch() };
        const auto name { selected.getProperty (Id::id).toString() };
        const juce::StringArray changeSet { TITGitProcessor::getLine ({ text::Git::rebase, name }) };

        if (isOtherBranch (selected))
        {
            if (static_cast<int> (selected.getProperty (Id::pushedCount)) > 0)
                state.setProperty (Id::message, text::Footer::pushedCommits, nullptr);
            else if (isDirty())
                setDirtyConfirmation (text::Confirmation::rebaseBranchDirtyTitle, text::Confirmation::rebaseBranchDirtyExplanation,
                                      text::Confirmation::rebaseBranchDirtyYes, text::Confirmation::rebaseBranchDirtyNo, changeSet, {});
            else
                setConfirmation (text::Confirmation::rebaseBranchTitle,
                                 getFormattedText (text::Confirmation::rebaseBranchExplanation, { getGit().getProperty (Id::currentBranch).toString(), name }),
                                 text::Confirmation::rebaseBranchYes, text::Confirmation::rebaseBranchNo, changeSet, {});
        }
    });
}
```

   - Enter is one command for the two contexts of `tit/SPEC.md:437-441`, and the context is the state: a detached HEAD means a return. The return lines are `checkout <branch>` and the stash of TIT, when one exists (Step 9, `getReturnLines`).
   - A new branch: the two lines are the legacy checks and the legacy command. The name comes from `InputView::getText` (Part 10 Step 14).
   - The merge, the delete, and their dialogs are legacy. The rebase onto a branch is `rebase <selected>` (Part 2 Step 7), with the limit of ruling 20 as a footer text (Closed by a read, item 19).

5. The rebase list, in three families: the range and the apply; the actions of a row; the reword.

```cpp
void TITView::registerRebaseActions()
{
    commands.add (map::CommandId::rebaseInteractive, [this] (const juce::CommandID&)
    {
        const auto git { getGit() };

        if (isDirty())
        {
            state.setProperty (Id::message, text::Footer::commitFirst, nullptr);
        }
        else if (static_cast<bool> (git.getProperty (Id::hasUpstream)) and static_cast<int> (git.getProperty (Id::commitsAhead)) == 0)
        {
            state.setProperty (Id::message, text::Footer::noUnpushedCommits, nullptr);
        }
        else
        {
            gitProcessor.readRebase (TITGitProcessor::getRebaseRange (git));
            state.setProperty (Id::focusedPane, 0, nullptr);
            state.setProperty (Id::mode, static_cast<int> (map::EditorMode::rebaseList), nullptr);
        }
    });

    commands.add (map::CommandId::rebaseApply, [this] (const juce::CommandID&)
    {
        if (const auto numCommits { state.getChildWithName (jam::Format::toType (Id::rebase)).getNumChildren() }; numCommits > 0)
        {
            gitProcessor.setRebaseFiles();
            setConfirmation (text::Confirmation::rebaseInteractiveTitle, getFormattedText (text::Confirmation::rebaseInteractiveExplanation, { juce::String { numCommits } }),
                             text::Confirmation::rebaseInteractiveYes, text::Confirmation::rebaseInteractiveNo, { TITGitProcessor::getRebaseLine (getGit()) }, {});
        }
    });
}

void TITView::registerRebaseListActions()
{
    commands.add (map::CommandId::rebaseSquash, [this] (const juce::CommandID&) { gitProcessor.setRebaseAction (text::Rebase::squash); });
    commands.add (map::CommandId::rebaseFixup, [this] (const juce::CommandID&) { gitProcessor.setRebaseAction (text::Rebase::fixup); });
    commands.add (map::CommandId::rebaseDrop, [this] (const juce::CommandID&) { gitProcessor.setRebaseAction (text::Rebase::drop); });
    commands.add (map::CommandId::rebaseMoveDown, [this] (const juce::CommandID&) { gitProcessor.moveRebaseCommit (1); });
    commands.add (map::CommandId::rebaseMoveUp, [this] (const juce::CommandID&) { gitProcessor.moveRebaseCommit (-1); });
}

void TITView::registerRewordActions()
{
    commands.add (map::CommandId::rebaseReword, [this] (const juce::CommandID&)
    {
        const auto commit { getSelected (Id::rebase) };

        if (commit.getProperty (Id::attribute).toString().compare (text::Rebase::reword) == 0)
        {
            gitProcessor.setRebaseAction (text::Rebase::reword);
        }
        else if (commit.isValid())
        {
            commitMessage.setDocument (commit.getProperty (Id::text).toString());
            state.setProperty (Id::mode, static_cast<int> (map::EditorMode::rebaseReword), nullptr);
        }
    });

    commands.add (map::CommandId::rewordSubmit, [this] (const juce::CommandID&)
    {
        const auto message { commitMessage.getText() };

        if (message.isEmpty())
        {
            state.setProperty (Id::message, text::Footer::commitMessageEmpty, nullptr);
        }
        else
        {
            gitProcessor.setRebaseMessage (message);
            state.setProperty (Id::mode, static_cast<int> (map::EditorMode::rebaseList), nullptr);
        }
    });
}
```

   - The range is `@{u}..HEAD`, or every commit of HEAD with no upstream (rulings 9 and 25). The two footer texts are the rules of Part 2 Step 7 (Closed by a read, item 19).
   - The edits of the list and its files are functions of the processor (Step 9, item 5). The view tells them.
   - A conflict during the rebase gives the resolver and its loop. A line of the list that fails gives `Rebasing` with no conflict, and the menu of Part 9 has "Continue rebase" and "Abort rebase".

6. `registerAvailability()`: add these entries.

   | Command | Rule |
   |---|---|
   | `branchPicker` | `operation` is `normal` or `timeTraveling` |
   | `branchPickerMerge`, `branchPickerDelete`, `branchPickerRebase`, `branchPickerAdd` | `operation` is `normal` |
   | `rebaseInteractive` | `operation` is `normal` |

   - Each rule reads the `operation` row, and its event builds the sheet (Step 11, item 6).
   - The picker does not open while `Pinned`: a switch would make the pin marker stale, and the HEAD before the pin would be lost.
   - The commands of a selected row have no state rule. Each one acts on its row (ruling 40; items 3 to 5):
     - `conflictSpace` and `conflictEnter`: item 3.
     - `branchPickerEnter`, `branchPickerMerge`, `branchPickerDelete`, `branchPickerRebase`: on the row of the current branch each one does nothing (`tit/SPEC.md:431-432`; Closed by a read, item 17).
     - `rebaseSquash`, `rebaseFixup`, `rebaseDrop`, `rebaseMoveDown`, `rebaseMoveUp`, `rebaseReword`, `rebaseApply`: each one acts on a commit of the list. The processor edits a commit that exists (Step 9, item 5).
   - `conflictSpace`, `conflictEnter`, `branchPickerEnter`, the seven list commands, `newBranchNameSubmit`, and `rewordSubmit` have no entry. The mode of the key row is the gate of each one.

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines: each of the twelve families is one function. No `==` on strings: each text test is `compare (…) == 0`.
- `commands` holds 60 ids. `availability` holds the ids that a state rule governs, and the 13 edit command ids. No callback and no line of `perform` tests if a command is allowed. A callback of a selected row reads only that row, which is its data.
- `TITView` joins no command line and writes no file. Each line comes from a `text::Git` row or from a function of the processor. Each file is written by the processor.
- Each flow of Part 2 Step 7 has its `Clean` path, its `Dirty` path, its conflict path, and its abort path.
- Data flows one way: key, sheet, command, Model row or processor request, repository, detection or pool job, Model, listener, view.
- Local names: `git`, `branch`, `commit`, `hash`, `changeSet`, `original`, `version`, `file`, `label`, `name`, `isReturn`, `selected`, `message`, `conflictFiles`, `isResolved`, `numCommits`.

### Step 13: The application, the editor variable
**Scope:** `tit/Source/Main.cpp`.
**Action:** In `runWindow`, after the two lines that set `text::Git::terminalPromptVariable` (Part 9 Step 8), add the same two lines for `text::Git::editorVariable` and `text::Git::editorValue`:

```cpp
       #if JUCE_WINDOWS
        _putenv_s (text::Git::editorVariable, text::Git::editorValue);
       #else
        setenv (text::Git::editorVariable, text::Git::editorValue, shouldOverwrite);
       #endif
```

**Validation:**
- Each child of TIT gets the variable (Part 9, Closed by a read, item 7). `GIT_EDITOR` is first in the order of the editors (inventory).
- No command of Parts 9 to 11 opens an editor: `commit -F`, `commit --no-edit`, `merge` with no terminal.

### Step 14: Build, run, and measure on macOS
**Scope:** JAM and the `tit/` project; fixture repositories outside every repository.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build). The fixture commands are git commands: ARCHITECT runs them, or gives the git instruction (CAROL.md Git).
**Action:**
1. Build `ling` and EVE as in Part 10 Step 18, items 1 and 2. Step 1 changed one cast table of JAM.
2. From `tit/`, run `cast cast/spell.md --debug`.
3. Start `titc` in END in each fixture below, and do the keys. The fixture numbers are those of Part 8 Step 9.

   | Fixture | Keys | Expected |
   |---|---|---|
   | 8, `Behind`, with one changed file that the remote commit does not touch | `{`, Return | The dialog `Save your changes?` with `SAVE CHANGES` selected. The console shows four lines: the stash, the pull, the apply, the drop. Then `In sync` and `Dirty`. `.git/TIT_DIRTY_OP` does not exist. |
   | 8, with a change to the line that the remote commit changes | `{`, Return, Escape | After the run, the resolver opens: one file, three panes with `BASE`, `<branch> (merged)`, `stashed changes`. |
   | the same | Tab, Tab, Space, Return | The file row shows the label of the chosen pane. The console shows the write, the add, and the drop of the stash. Then `Normal`, and the marker does not exist. |
   | the same fixture again, in the resolver | Escape | The console shows `reset --hard HEAD`. The menu then shows `Reapply changes` and `Abort and restore`. `a` gives the tree of the start: `Behind`, `Dirty`, and the change. |
   | 8, `{`, then the second button | `{`, `n`, Return | The console shows the reset, the clean, and the pull. Then `In sync` and `Clean`. No marker and no stash. |
   | 11, a merge conflict | none | The resolver opens at the start. The labels are `BASE`, `<branch> (current)`, `<other> (incoming)`. |
   | 11 | Return with no chosen version | Nothing runs. |
   | 11 | Tab, Space, Return | The console shows the lines and the merge commit. Then `Normal`. |
   | 11 again | Escape | The console shows the abort of the merge. Then `Normal` and the tree of before. |
   | a conflict where one side deleted the file | choose the side that deleted it, Return | The file is removed, and the merge commit has no such file. No placeholder text is in the tree. |
   | fixture 4 with three commits | `h`, `j`, Return, Return | The dialog `Time Travel Confirmation` with the short hash and the subject. Then the header shows `🌀 TIME TRAVEL`, the branch, `DETACHED @` with the hash, and `Viewing commit from` with the date. The menu shows `History`, `File(s) history`, `Return`. |
   | the same, in time travel | `r`, `y`, Return | The dialog `Return to <branch>` with `CANCEL` selected. Then `Normal`, and `.git/TIT_TIME_TRAVEL` does not exist. |
   | the same, with a changed file before the time travel | `h`, `j`, Return, Return, then `r`, `y`, Return | The journey starts with a clean tree. After the return, the change is in the tree again, and the stash list holds no stash of TIT. |
   | the same, with a new file made during the journey | `r` | The dialog `Return to <branch> with uncommitted changes` with `DISCARD CHANGES` selected. `y`, Return: the file is in the tree of the branch. |
   | 13, a detached HEAD | `r` | The branch picker: the branches with their tracking text, and the details of the selected branch. Tab gives the focus to the details; `j` scrolls them. |
   | 13 | `r`, Return on a branch | `Normal` on that branch. |
   | fixture 5 with two branches, from a temporary key row `branch picker`, `b`, `menuNormal` | `b`, `a`, a name, Escape, Return | The footer shows `New branch name:`. The console shows the two lines. The header shows the new branch. |
   | the same | `b`, `j`, `m`, `y`, Return | The dialog `Merge branch?`. The console shows the merge. |
   | the same | `b`, `j`, `x`, `y`, Return | The dialog `Delete branch <name>?`. The branch is gone. |
   | the same, the selected branch is behind a pushed commit of the current branch | `b`, `j`, `r` | The footer shows `Pushed commits cannot be rebased`. Nothing runs. |
   | the same, with one unpushed commit and a selected branch that has one more commit | `b`, `j`, `r`, `y`, Return | The dialog `Rebase onto branch?`. The console shows the rebase. The log shows the unpushed commit on the selected branch. |
   | 8, `Behind`, clean, with one local commit that is not pushed | `r`, `y`, Return | The dialog `Pull with rebase?`. The console shows the pull. The log shows the local commit after the remote commit, with no merge commit. |
   | fixture 5 with three unpushed commits | `h`, `p` on the second commit, Return | The dialog `Pin Confirmation`. The header shows `PINNED @` with the hash. The menu shows `History`, `File(s) history`, `Unpin`. `.git/TIT_PIN` has three lines. |
   | the same, pinned | `u`, Return | The dialog `Unpin?`. The branch has its three commits again. The marker does not exist. |
   | the same, pinned, a second pin on an older commit, then `u` | | The branch returns to the HEAD before the first pin. |
   | the same, not pinned | `h`, `i` | The rebase list: three rows, oldest first, each with `pick`. |
   | the same | `j`, `s`, `j`, `d`, `k`, `K`, `r`, a new text, Escape, Return | The rows show `squash`, `drop`, and `reword`, and one row moves up. |
   | the same | Return, `y`, Return | The dialog `Rewrite commits?` with the count 3. The console shows the rebase. The log shows the new order, the squashed commit with both messages, no dropped commit, and the new message. No editor opens. |
   | fixture 5 with no commit ahead | `h`, `i` | The footer shows `No unpushed commits to rewrite`. |
   | fixture 4 with a changed file | `h`, `i` | The footer shows `Commit or discard changes first`. |
   | 12, a rebase in progress with a conflict | none, then choose, Return | The resolver with `BASE`, `OURS (rebased)`, `THEIRS (incoming)`. The console shows `rebase --continue`, and no editor opens. |
   | any dialog of this part | Escape | No marker file is written. |

4. Remove the temporary key row of item 3. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- Each build ends with exit code 0. Each row gives its expected result.
- These facts of this part are measured in the program: `stashLog` with the `--grep` argument gives the reference of the stash of TIT; a line with an argument in double quotes is one argument of the child; the `sequence.editor` command finds the list file from the repository folder; `GIT_EDITOR=true` gives a squash its prepared message; `checkout-index --stage` and `rm -f` on an unmerged path; `name-rev` of `MERGE_HEAD`.
- These facts of the forms of rulings 57 and 58 are measured in the program: the three trees arrive whole through the lane of the records; each version pane shows the bytes of its `showStage` line, with the stage and a path with a space as one quoted argument; a version that does not exist gives an empty pane and its text in the label; the rows `isCurrent` and `pushedCount`, which arrive as texts, give the flag and the count.
- The `Stash Not Found` dialog is measured: time travel from a dirty tree, then a drop of the stash of TIT by hand, then `Return`. The dialog shows with the label of the stash in its text, and `Continue` returns to the branch.
- ARCHITECT sets the layout of the resolver, of the picker, and of the rebase list at this run (ruling 7).
- The second `cast` run changes no generated file. No temporary row remains.

### Step 15: Run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 14, items 2 to 4, in END on a Windows machine with these rows: the first dirty pull, the merge conflict of fixture 11, the time travel and its return, the pin and the unpin, and the interactive rebase with a squash and a reword.
**Validation:**
- The results of Step 14 hold.
- The tool on Windows runs `cp` of the `sequence.editor` value in its own shell, and `CommandLineToArgvW` gives the quoted argument as one argument.
- A marker file has `\n` line ends, and the detection reads it.

## BLESSED Alignment

- **B — Bound.** `TITView` owns `ConflictView` and `BranchView`. Each view owns its text views. JAM owns the window of a menu until it closes (Part 10 Step 15). Each marker has one writer, `TITGitProcessor::setMarker`, and one remover, the detection. The folder of the rebase list has one writer, which replaces it whole. Each pool job owns its tree until the message thread takes its children.
- **L — Lean.** One menu function of Part 11 and one text editor type for the three new screens. One stash label. One dirty chain for six operations. No class and no object for a protocol.
- **E — Explicit.** Each command line, text, key, and file name is a named row. A stash is named by its reference. A stop of a chain shows the commands of its state.
- **S — SSOT.** The repository is the one source of the state of a protocol. The `git` rows are its one reflection. The conflict type, the labels, and the finalize come from the same three rows.
- **S — Stateless.** No phase, no flag, and no pending step in `TITView` or in the processor. The chosen versions and the rebase list are Model rows.
- **E — Encapsulation.** The processor holds the protocols: the command lines, the markers, and the files of the rebase list (Step 9). A view composes and paints. `TITView` tells the processor to run lines or to read a list, and reads no answer from it. Each list is complete at creation, with its command lines. No listener and no callback is new: a protocol continues through the detection and the events of Parts 8 to 11.
- **D — Deterministic.** The same repository gives the same rows, the same menu, the same resolver, and the same command lines.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

1. **The exact text of the new items of this part.** Closed by ruling 41: the texts of the section below stand. None open. The items of "Closed by a read" are ORACLE's reads; ARCHITECT can reverse each one. The runtime facts of this part are in Steps 14 and 15.
2. The record of the stash of a time travel and the `Stash Not Found` dialog are written (Closed by a read, item 14; Steps 2, 3, 5, 6, 9, and 12).

## The exact text of the new items of this part

Legacy and Part 2 hold no text for these items. Ruling 41: *"They stand"*.

| Where | Text |
|---|---|
| Menu row `dirty continue`, label and hint | Reapply changes — Apply your saved changes again |
| Menu row `dirty abort`, label and hint | Abort and restore — Restore the state before the operation |
| Footer, resolver, the menu of the files has the keys | `↑↓ navigate  │  Tab versions  │  Enter resolve  │  Esc abort` |
| Footer, resolver, a version has the focus | `↑↓ scroll  │  Space choose  │  Tab next  │  Enter resolve  │  Esc abort` |
| Footer, picker, the current branch | `↑↓ navigate  │  a add  │  Tab details  │  Esc cancel` |
| Footer, picker, another branch | `↑↓ navigate  │  a add  │  m merge from  │  r rebase onto  │  x delete  │  Enter switch  │  Esc cancel` |
| Footer, rebase list | `s squash  │  f fixup  │  r reword  │  d drop  │  J K move  │  Enter apply  │  Esc cancel` |
| First button of the dialog of a clean return | Return |

"Abort and restore" is the wording of `tit/SPEC.md:956`. The picker hints are the legacy hints (`legacy/app/messages_menu.go:151-166`) with `Tab details` and `r rebase onto` added, and with no `Enter switch` on the current branch.
