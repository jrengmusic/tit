# PLAN: TIT — Tracker and Canon

**RFC:** none — objective from ARCHITECT prompt. `RFC.md` (2026-04-17) is stale: it describes `jam_tui`, which no longer exists.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This file is the index, the status tracker, the canon, and the inventory of the TIT work.
TIT is the Terminal Interface for git: a git tool for users who are not git experts.
The numbered files are one PLAN in smaller parts. This file replaces `PLAN-tit-cpp-port.md`.

## Status

| Part | Plan | Subject | Status | Depends on |
|---|---|---|---|---|
| 0 | this file | Tracker, canon, inventory | Written. Not locked. | |
| 1 | `PLAN-tit-1-project.md` | The console project from the cli archetype. The stale tree is removed first. | Written. Not locked. No open item. | 0 |
| 2 | `PLAN-tit-2-spec.md` | `SPEC.md` rewritten: nothing dropped, pin, three rebase actions, the reading contract, commands and keys as tables | Executed and logged: Sprint 19, 2026-11-15. | 0 |
| 3 | `PLAN-tit-3-terminal-display.md` | JAM: `TerminalDisplay`, the program's own terminal. Raw mode, read, write, size, resize, on both platforms. The missing `TerminalEncoder` writers. The inline image lane, brought back verbatim (ruling 24). | Written. Not locked. No open item. One fact to measure (tracker list). | 0, 1 |
| 4 | `PLAN-tit-4-key-input.md` | JAM: input bytes to tokens, tokens to `juce::KeyPress` | Written. Not locked. No open item. | 3 |
| 5 | `PLAN-tit-5-terminal-window.md` | JAM `jam_gui`: `TerminalWindow` and `TerminalComponentPeer`, beside `jam::Window` | Written. Not locked. No open item. | 3, 4 |
| 6 | `PLAN-tit-6-text-editor-insert.md` | JAM: the `insert` mode and the edit operators of `jam::TextEditor`. The owner's key sheet gives the edit keys. `jam::TextEditor` implements `juce::TextInputTarget`, so a pasted text arrives whole (tracker, Closed by a read). Restored verbatim from `3585cb66d` | Written. Not locked. No open item. | 0 |
| 7 | `PLAN-tit-7-model.md` | `TITModel` in the LING form, with the rows that this part reads. The terminal chain in the EVE form (ruling 30): `TerminalDisplay`, a `jam::VirtualDevice`, the key lane, the `winsize` and `numKeyPresses` rows. The tick reads the keys and the size. The frame goes from the peer to `TerminalDisplay::write` on the message thread (rulings 52 and 53). The first run of `TerminalWindow` with the help guide, and the key sheet with its first row, `quit` (ruling 33) | Written. Not locked. In the form of ruling 62: Step 4 makes `TITLookAndFeel : jam::StyleTerminal`, and the application holds it as `lookAndFeel`. | 1, 2, 3, 4, 5 |
| 8 | `PLAN-tit-8-git.md` | The git layer: state detection. `TITGitProcessor` runs each git command as a `jam::Subprocess` on a one-thread pool and writes one complete `git` child of `TITModel`: the five axes, the markers, and the LFS fields, with their identifiers. The window run starts it. Three legacy detection faults are corrected by SPEC (Part 8, inventory) | Written. Not locked. No open item. | 1, 2, 7 |
| 9 | `PLAN-tit-9-commands.md` | The command layer in the END and EVE pattern: the `## CommandId` cast table, the key rows in `Source/keys.md`, `TITView` as the command target, and availability from the `git` rows of Part 8. One key-sheet mode for each Operation menu, as rows of JAM's `## EditorMode` (ruling 34). The first six user commands: `push`, `pushAutoSync`, `finalizeMerge`, `abortMerge`, `rebaseContinue`, `rebaseAbort`. `TITGitProcessor::run` runs their command lines and then the detection again. Each later part adds the command ids, the key rows, and the modes of the view that it builds (Part 9, Closed by a read, item 1) | Written. Not locked. No open item. | 5, 6, 7, 8 |
| 10 | `PLAN-tit-10-views.md` | The shell views: `HeaderView` and `FooterView` in the form of `jam::TextEditorStatusBar`; the menu and the dialog, which are the stock components of JAM and JUCE inside the view; `ConsoleView` and `InputView`, each one screen view around one `jam::TextEditor`. `TITView` composes the views and registers the commands in families, in the END form; the processor makes the command lines and writes the files (Part 9, Overview, item 6). The modes `console`, `confirmation`, `input`. The menu, dialog, cancel, and submit commands, and `commit`, `commitPush`, `resetDiscardChanges`, `forcePush`, `replaceLocal`, `pullMerge`. The console gets the output of a run. The commit box binds the edit keys of `jam::TextEditor`. The run chain of `TITGitProcessor` in the EVE form (rulings 58 and 59): TTY, decoder, buffer, swap, and the console editor bound to the swap. JAM: a program run on `jam::TerminalTTY`, the keys `j` and `k` of the menu, `TextEditor::setDefaultMode`, and the mode and colour rows | Written. Not locked. In the form of rulings 65 and 66, with no code written by hand: the dialog is the `juce::AlertWindow` of JUCE inside the view, shown by `jam::TerminalMessageBox::showYesNoBox` (Step 14), and Steps 4, 5, 7, 8, 9, 15, and 16 hold no row, no command, and no key for a dialog. The menu is the stock popup menu of JAM inside the view: `jam::showAsync` with a parent component (Steps 12 and 16); JAM gets only the keys `j` and `k` in `MenuWindow::keyPressed` (ruling 63). No component, no LookAndFeel method, no size, and no paint code is written by hand in this part. Two legacy details have no carrier in the stock API and are not carried: a dialog that opens with its second button selected, and the footer hint of the highlighted menu item. The console is in the form of rulings 57, 58, and 59 (Steps 9, 10, 14 item 7, and 16): a line of a run is one TTY session, the decoder fills the swap on the tick, the console editor reads the swap, and the rows `shellExited` and `numShellExits` carry the end of a session to the view, which tells the processor. The command line and the abort text show in the footer, because the swap holds only the bytes of the tool. Step 2 changes no file: no key row of Parts 10 to 13 has a key description with a space, so the key sheet of JAM stays as it is. | 5, 6, 7, 8, 9 |
| 11 | `PLAN-tit-11-history.md` | The history browser (ruling 40): one view over one shape of records, with three levels: the commits, the files of the selected commit, the diff of the selected file. Its feeds: the cache of 100 commits as a Model tree, built by one pool job (rulings 37 and 38); a search; the incoming and outgoing commits; the working tree changes; the history of one function. The records show in the stock popup menu of JAM (ruling 60): a commit is an item, and its files are its submenu. One function, `getRecordMenu`, builds the menu of every tree of records. One text pane, and the screen view `HistoryView` that owns it; the menu rows, the separator, and the labels of a cache that loads. Rewind and copy hash. No blame (ruling 39). JAM: the mode rows | Written. Not locked. In the form of rulings 60 and 61: no list component, no paint code, and no command that moves a highlight. The highlighted record becomes the selected record, a Model row, at each key that reaches `TITView` and at Return on an item; one function reads the highlight (`jam_gui/menu/jam_PopupMenu.h:1794`). Two details have no carrier in the stock menu and are not carried: a detail that follows each move of the highlight (the menu gives no call for it, `jam_PopupMenu.h:1618-1633`), and a row with two colours (ruling 64; `juce::PopupMenu::Item` has one `text` and one `colour`). The footer hint texts of Step 3 are the legacy texts. In the form of rulings 57 and 58: each feed goes through the lane of the records and the count row `numRecords`, and `applyRecords` builds the tree on the message thread; the cache holds the records and no patch; the patch of the selected record is one run on the run chain `patch`, and the text pane is a `ConsoleView` over its swap. No row of the Model holds output of the tool. | 8, 10 |
| 12 | `PLAN-tit-12-protocols.md` | The protocols. The state of a protocol is in the repository: a marker, a stash of TIT, a merge, a rebase, unmerged files. The detection reads it into rows of the state child, and the menu and the commands follow the rows. The Dirty Operation Protocol with its resume and its abort; the menu row `dirty_pull_merge`. The conflict resolver: the menu of the files and three text panes, as the screen view `ConflictView`. The branch picker as `BranchView`: the menu of the branches and the details. The protocol lines, the markers, and the rebase files are functions of `TITGitProcessor`. Time travel, its menu, and its return; the commit date of the header. Pin and unpin; the pinned menu. The branch picker: switch, add, merge from, delete, rebase onto. Pull with rebase. The interactive rebase of the unpushed commits. This part replaces one rule of Part 8: `DirtyOperation` is a stash of TIT on a branch, not the marker alone (Part 12, Closed by a read, item 3). JAM: the mode rows | Written. Not locked. In the form of rulings 57 and 58: the detection stores each row with `storeValue`; the three trees go through the lane of the records of Part 11; the three version panes are `ConsoleView`s over three run chains, each with a `juce::Label`. The `Stash Not Found` dialog is written: the time travel marker records the stash of the entry, and the row `hasOriginalStash` carries it (Part 12, Closed by a read, item 14). In the form of rulings 60 and 61: the files of the resolver, the branches, and the rebase list show in the stock menu through `getRecordMenu` of Part 11, and Return on an item runs the command of the legacy Enter of the screen. No list component and no move command exist. | 8, 9, 10, 11 |
| 13 | `PLAN-tit-13-setup.md` | Setup wizard, preferences, themes, auto-update, LFS. The folders, the ignore file, the key files, and `~/.ssh/config` are work of `TITGitProcessor`; `TITView` gives the typed text and registers the command families. The config tables in `~/.config/tit/` as Model rows (Part 7, Closed by a read, item 1). `init`, `clone`, `add_remote`, the Config menu, the startup gate, the sync spinner of the header, `lfs checkout` after a hard reset, the theme folders (ruling 43), the user key override (Part 10, Closed by a read, items 4 and 19). The user's count of history commits (ruling 38) and the turning spinner (Part 11, Closed by a read, items 7 and 8). The Config row that opens the branch picker, and `add_remote` with no remote (Part 12, Closed by a read, items 16 and 32). The user files `user.md` and `keys.md` in the JFS shape (rulings 44 to 47). This part replaces Part 2 Step 9 (five themes), Part 9 Closed by a read item 10 (the END citation), and the cast row `historyCount` of Part 11, and Part 2 Step 5 for `SPEC.md:770-776,825-832` (not built; the console shows the text of the tool) | Written. Not locked. In the form of ruling 57: `update`, the folder functions, and `setLfsFilters` write the Model only with `storeValue`; the rows `isUpdating` and `isRemoteReachable` are parameters, and the view shows the text of a failed fetch. In the form of ruling 62: the carried `style.css` is the default theme, the theme file on disk wins when it is valid (Step 5), and `applyTheme` applies it to `lookAndFeel`, the `TITLookAndFeel` of Part 7. The menus are the stock menu of Part 10, and Escape in a submenu reaches `cancel` through the callback of `setMenu`. The wizard dialogs are `jam::TerminalMessageBox::showAlert`. | 8, 10, 11, 12 |
| 14 | `PLAN-tit-14-release.md` | The release notes rows; `ARCHITECTURE.md` and `README.md` from the built code; ARCHITECT's signed release runs on macOS, Windows x64, and Windows arm64. The lane itself is generated in Part 1. Runs after Part 15, because each document records the code that exists (Part 14, Closed by a read, item 2) | Written. Not locked. No open item. | 1 to 13, 15 |
| 15 | `PLAN-tit-15-banner.md` | JAM: `BrailleComponent`, and the braille dot table as two cast lookup tables (ruling 50). TIT: the logo as a `juce::Image`, shown by `BrailleComponent` or by `juce::ImageComponent`, ARCHITECT's choice at execution (ruling 23); the braille banner of `titc --help`; the banner region beside the menus, Config, and Preferences (Part 10, Closed by a read, item 19). Two legacy faults corrected: the dot bits and the fill rule | Written. Not locked. No open item. | 1, 3, 5, 10, 13 |

The part list is a proposal. ARCHITECT changes it at will.
Parts 3 to 6 change JAM. Part 9 Step 1 adds rows to JAM's `## EditorMode`. Part 10 Steps 1 to 4 change JAM: `jam::Subprocess`, `jam::CommandManager`, `jam::TextEditor`, and the rows of `## EditorMode` and `## ColourId`. Part 11 Step 1 adds rows to the same two tables. Part 12 Step 1 adds rows to `## EditorMode`. Part 13 Step 1 adds rows to `## EditorMode`. Part 15 Steps 1 and 2 change JAM: two tables of `cast/graphics.md`, their rows in `cast/spell.md`, `TerminalGlyph`, and the new `BrailleComponent`. Part 6 also affects EVE, which is a consumer of `jam::TextEditor` (`eve/Source/EVEView.cpp:316`).

## The contract of every part

ARCHITECT, verbatim, session 2026-11-13:

- *"the principle of tit as power git tools for non savvy git user STANDS. but the plan must be carefully EXTRAPOLATED, not by a single stale list that behind the framework."*
- *"it need to be broken down into comprehensive multi-part, ready to exexute plan, without any open questions, unambiguously without any gaps could be filled in for COUNSELOR inference. all questions can be answered by reading citated facts, SPEC, ARCHITECTURE, agreed, locked PLAN. no assumptions, no asking answerable questions by reading."*
- *"tui could be build identically with any gui with juce native API contract using juce::Component base, juce::LookAndFeel (our Style etc)."*
- *"my point is, tit is way behind. it needs to conform with other project structured with manifest, build with cast."*
- *"ask only open question no read can answer"*

From CONTRACT, with no ruling needed:

- JAM first, then JUCE. CODING.md, Critical Rules: "Use the project's in-house framework first".
- Delete first. CAROL.md, Refactor-Rewrite Discipline: "old code is removed in step 1 and never coexists with new".
- A new domain term goes to ARCHITECT. A name from an established family needs no approval (NAMES.md Rule -1, Rule 5).

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE: one object is one header and one `.cpp`. The 300-line threshold is a smell detector for responsibility count.
- CODING.md: no include in a submodule header or source file. The topmost module header holds every include. Platform headers are the exception.
- CODING.md: no anonymous namespace, no `namespace detail`, no out-parameter, no fake-carrier struct, no local struct, no forward declaration, no bail-out guard.
- CODING.md, Platform macro names: no Windows SDK macro name as an identifier. Parts 3 and 5 touch Windows code.

## Canon: rulings, verbatim

Session 2026-11-13. A ruling by option selection quotes the selected label.

1. **Pin.** ARCHITECT: *"sometimes we need to `pin` tree at certain commit in the past without branch, without detach"*; *"with time travel we could detach, but why not with the exact same interface we could actually checkout? rewind but not detach? reset to past commit (probably can, but hidden)"*. Selected: *"Reset + pin marker"*. Option text: "`git reset --hard <hash>` on the current branch, plus a TIT marker file as `.git/TIT_TIME_TRAVEL` (SPEC.md:499). While the marker exists, the menu hides Pull and Replace local and shows Unpin."
2. **Rebase.** ARCHITECT: *"sometimes we need to rebase"*. Selected: *"Pull with rebase, Rebase onto branch, Interactive rebase"*. Option texts: "`git pull --rebase` as a menu item beside Pull (fetch + merge) when the branch is Behind or Diverged"; "In the Branch Picker, a key beside `m` merge that runs `git rebase <selected>` for the current branch. Conflicts use the existing rebase resolver loop"; "Reorder, squash, or drop commits from the History view (`git rebase -i` semantics)."
3. **Nothing is dropped.** ARCHITECT: *"nothing should be dropped."*
4. **Keep the feature, fix the design.** ARCHITECT: *"also file history is kinda useless with legacy UI, hard to read, hard to understand. my point is kept the feature FIX THE DESIGN, DESIGN BY CONTRACT."*
5. **`jam::TextEditor` is the text widget, for view and for input.** ARCHITECT: *"almost all legacy tit handroll text editor, with its some effort of core parity feature of vim."*; *"our TextEditor already have ALL. why not just use it?"*; *"we just wanted to ensure our TextEditor as juce::Component works identically on Terminal without changing its contract, never poke layers, kept encapsulated"*; *"NO, not stays read only. there are text input could use TextEditor as the Component."*; *"commit message text box would be MUCH BETTER using TextEditor. full feature"*.
6. **The reading problem is presentation.** ARCHITECT: *"so the issue only presentation. we have the data, but rendering probably limited by our past knowledge, and limitation of go/lipgloss, but i dont think we have no issue here. juce is ARCHITECT expertise"*.
7. **Views come from SPEC. A layout locks at execution.** Selected: *"Plan builds views from SPEC. but layout never locked until execution. only runtime could deterministically decide how should be the layout rendered"*.
8. **The pinned menu.** Selected: *"No commit while pinned"*. Option text: "As time travel today: changes are allowed in the tree, no Commit item (SPEC.md:235-239). Menu: History, File history, Discard changes, Unpin. Unpin with a dirty tree uses the Dirty Operation Protocol."
9. **Interactive rebase range.** Selected: *"Unpushed commits only"*. Option text: "Only commits ahead of the remote (`@{u}..HEAD`). A branch with no remote can rewrite all its commits. The action is absent when nothing is ahead, so no force push is ever needed (SPEC.md:1019)."
10. **Interactive rebase edits.** Selected: *"Squash / fixup, Reword, Drop, Reorder"*.
11. **State.** ARCHITECT: *"we also dont need to to handroll TitState, you can look how other jam project built with audio discipline architecture. now we have jam::Model, so the shape, the principle holds"*.
12. **Commands.** ARCHITECT: *"also now we manage all command, callback, keyboard shortcut with CommandManager, i believe even in a go legacy there's an effort to do such thing, but here ours is way better, BLESSED compliance"*.
13. **`TerminalWindow`.** ARCHITECT: *"that's the design. how do we design TerminalWindow crossplatform that adhere to juce::Component contract upstream? something like ComponentPeer? but speaks terminal context? so we could paint any Component into it?"* The reads confirm the design on vanilla JUCE 8 (The JUCE attach path, below).
14. **The peer class.** ARCHITECT: *"TerminalComponentPeer"*.
15. **The program's own terminal.** Selected: *"TerminalDisplay"*. Option text: "Backends `TerminalUnixDisplay`, `TerminalWindowsDisplay`. `jam_TerminalWinsize.h:9-13` already cites a `Display` as \"the sole author\" of the window size, and no such class exists."
16. **The mode row.** Selected: *"insert"*. Option text: "Row `insert` in `map::EditorMode`, status label INSERT, key sheet `## insert` as the ruled edit design names it (eve/PLAN-Grand-Scheme.md:1395-1396)."
17. **The TOML files of the Go TIT.** Selected: *"Leave them, write new"*. Option text: "The new TIT never reads or changes a `.toml` file. It writes its own markdown files beside them with getOrCreate, as END does for its config (end/Source/config/ConfigModel.cpp:142-143). Both binaries keep working."
18. **The help banner.** Selected: *"You supply the rows"*. Option text: "You give a `tit-banner.sh` or the glyph rows, as for LING. Part 1 then adds `cast/banner.md` with colour names from `## ColourNames` nearest to the banner colour, as LING Part 1 did." Replaced by ruling 23.
19. **The reading views.** Selected: *"Working tree changes, What a pull or push gives, Search through history, Function history"*. Option texts: "When the tree is Dirty: the changed files as a list with added and deleted line counts, and the diff of each"; "When Behind, Ahead, or Diverged: the incoming and the outgoing commits as two lists before the action"; "Find the commits where a text appeared or vanished, from the History view"; "The history of one function or line range, from the file view". Moved-code marking was not an option and stays out.
20. **Rebase onto a branch with pushed commits.** Selected: *"Same limit"*. Option text: "The command is active only when every commit it replays is unpushed. Otherwise it is inactive, and the footer says why. No force push is ever needed (ruling 9, invariant of Part 2 Step 2)."
21. **The default keys of the new commands.** Selected: *"They stand"*. The rows: History view, pin `p`, interactive rebase `i`. Rebase list, squash `s`, fixup `f`, reword `r`, drop `d`, move down `J`, move up `K`, apply `enter`, cancel `escape`. Branch Picker, rebase onto `r`. File view, blame `b`. Main menu, Pull (rebase) `r`, Unpin `u`.
22. **The exact text of the new items.** Selected: *"They stand"*. The strings are in `PLAN-tit-2-spec.md`, section "The exact text of the new items": four menu rows, eight dialogs, three inactive-command texts, four other texts, and the keys `/` (search history) and `F` (function history).
23. **The banner.** ARCHITECT: *"tit `banner` can use from legacy no? from svg with braille. that's THE BANNER. so we need SVG renderer to braille Component. we possibly CAN render to actual pixel too. BOTH, decide at execution. but implement both component"*. This replaces ruling 18.
24. **Pixels.** ARCHITECT: *"are you absolutely sure we cant render to pixel to terminal? we have tested this with Mermaid"*; *"ok, TIT is terminal agnostic. render pixel itself then. so this plan must explicitly bring back tested pixel renderer to our Terminal LLGC verbatim."* TIT sends no host-specific bytes; the EVE envelope is not used.
25. **The braille component.** Selected: *"BrailleComponent"*. Option text: "The `<What>Component` family of its siblings `ImageComponent` and `PreviewComponent`. File `jam_gui/widgets/jam_BrailleComponent.h` and `.cpp`."
26. **The lock.** Selected: *"All parts as one"*. Option text: "The LING form. No part executes until all 15 parts are written and you lock the whole PLAN."
27. **The old build tree.** Selected: *"Part 1 Step 1 removes it"*. Option text: "@Engineer deletes `tit/Builds/Ninja` with the rest of the stale tree. The folder is ignored by the repository. `tit/.clangd` stays as it is; the clangd file is MACHINIST work."
28. **The pixel lane of `drawImage`.** ARCHITECT, session 2026-11-13: *"you cant prove there is caller, consumer, nor usage"*. The inline image lane is the only body of `drawImage`. Part 3 deletes the body of today (`jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:488-509`) and writes no second lane.
29. **Two directions of an image.** ARCHITECT: *"in the past inline image need layout and render, current keep the layout while rendering is RENDERER responsibility. that contract STANDS. we still want SKiT only kept layout and delegate rendering. we could restore the image rendering and kept the layout."*; *"think about it as 2 direction. image placement is INPUT image rendering (including placement) is OUTPUT"*. INPUT is `TerminalProcessor`, the numbered `TerminalBuffer::ImagePlacement`, `TerminalImageRecord`, and the Preview component (`2b20e4f61`). It stays. OUTPUT is `drawImage` to the VT lane to the host terminal. Part 3 restores it verbatim from `7fbf821ba`, with its own placement that holds the PNG. The OUTPUT never reads `TerminalImageRecord`.
30. **The input path.** Selected: *"Readiness events + hop"*. Then ARCHITECT: *"no manual callbacks, fully event driven. no fucking reinvention of new pattern, what is the established pattern?"* The established pattern is EVE's terminal chain. A `jam::VirtualDevice` owns the serial context (`jam_VirtualDevice.h:19-20,99-102`). The readiness callback clocks the device (`eve/Source/EVETerminalProcessor.cpp:76-82`). The tick stores into Model rows and FIFO lanes (`:146-159`). The Model flushes on the message thread (`jam_Model.h:922-928`). Listeners react (`eve/Source/EVEView.cpp:138-151`; `EVETerminalProcessor.cpp:68-72`). Part 3 drops `onWinsizeChanged`: a size change is a readiness event. Part 7 builds the chain.
31. **The end of a frame.** Selected: *"Frame ends without newline"*. Option text: "Part 3: the VT lane writes the newline between rows, not after the last row. LING's printHelp writes one newline after writeToStream (ling/Source/Help.cpp:32-34), so `ling --help` stays byte-identical. TerminalWindow uses every row of the host." Basis: a newline on the last row of the scrolling region scrolls the screen (`jam_TerminalProcessor.cpp:635-643`).
32. **The document of an editable `jam::TextEditor`.** Selected: *"AnsiDocument, as tested"*. Option text: "Edits work on AnsiDocument lines, exactly as the restored InsertLine/RemoveLine/addLine expect. Part 6 opens one public entry for it: setLiveBlock (std::unique_ptr<Document>) becomes public, and the owner parses the text with AnsiDocument::addLines (data, size) first. setDocument<> stays for read-only MarkdownDocument views."
33. **The first run.** Question: "Part 7 is planned to run TerminalWindow for the first time. Raw mode gives Ctrl+C to TIT as a key byte, not as a signal. A key reaches a command only through the jam::CommandManager key sheet, and that sheet is Part 9 (ruling 12). How does Part 7 handle its first run?" Selected: *"Part 7 starts the key sheet"*. Option text: "Part 7 adds the command wiring with one row: quit = ctrl + C (legacy app_keys.go:10). The window shows the help guide in a read-only TextEditor (the LING Help form). Part 9 adds the other rows. No code is thrown away later."
34. **The keys of the menus.** ARCHITECT: *"table driven static, cast manifest, mapped to CommandManager, this always the pattern. default config, user override"*; *"the mapping was never 1:1. one API could be one command id but one command id, could chain couple of API, or calling other commandId one keyboard shortcut can mapped into one commandId"*. Selected: *"One mode per Operation menu"*. Option text: "Each Operation menu (notRepo, normal, merging, rebasing, timeTraveling, conflicted, dirtyOperation, pinned) is its own key-sheet mode, as legacy builds one menu per Operation (menu.go:22-31). `c` is clone in notRepo, commit in normal, continue in rebasing. Inside normal, `[` `]` `r` chain by Timeline and Remote."
35. **The `r` row and the `]` row of the normal menu.** Selected: *"addRemote, calls pullRebase"*. Option text: "Row `addRemote | r | menuNormal`. With no remote it adds a remote. With a remote it calls `pullRebase`." The question text: "For `]` the row is `push`: it pushes when Ahead and calls `pushAutoSync` when Diverged".
36. Void. It was a question of ORACLE on File history. Its premise was a sentence of ORACLE's own Part 2 draft, not a word of ARCHITECT.
37. **The history cache.** ARCHITECT: *"cache is preferrable, since it could speed up everything. unless you absolutely sure we can always build everything instantly."*; *"i believe this was also legacy defect, building even 30 caches could take couple of seconds, acceptable with go/lipgloss. absolutely not with jam."*; *"jam document parser could parse huge docs in subseconds"*; *"i believe we could now make the cache building instant"*. The measurements are in Part 11, "Measured in this session".
38. **The count of the cache.** ARCHITECT: *"100 commit as default. we could expose for user config"*.
39. **No blame.** ARCHITECT: *"i dont want blame feature for tit"*; *"i have never need blame"*. This removes the row "File view, blame `b`" of ruling 21, and the blame command of Part 2.
40. **One history browser.** ARCHITECT: *"there is only A `history` view, an SSOT of RECORDS. - History of whole repo in general - we could also dive in to see each commit per file. so there is no multiple view, all of them build what the same SSOT widget, feed by different data, rendered according to the data, user can interact according to the data. isnt that the whole premise of TIT?"*; *"History should be designed as BROWSER. hiearchial, we browse from top the commit, down to details per file."*; *"i said i want history that's the spec."*
41. **The new texts of Part 12.** Question: do the eight texts of `PLAN-tit-12-protocols.md`, section "The exact text of the new items of this part", stand? Selected: *"They stand"*. The texts: the menu rows `Reapply changes` and `Abort and restore` with their hints, the five footer hint rows of the resolver, the picker, and the rebase list, and the button `Return`.
42. **No theme manager.** ARCHITECT: *"reminder we dont want byte-to-byte translation. we dont need theme manager for new tit. jam has StyleManager, we have our own robust css-md-Style as Theme, nothing need to be reinvent"*.
43. **One shipped theme.** Selected: *"Ship gfx only"*. Option text: "TIT ships one theme, its style.css. A user adds a theme as a folder in ~/.config/tit/themes/<name>/ in the END form. The Preferences row cycles the folders. Spring, summer, autumn, winter are not shipped."
44. **No settings file.** ARCHITECT: *"NO. CLI is not pluging. we dont need *.settings"*.
45. **The user file.** ARCHITECT: *"what if we just write it ~/.config/tit/config.md ? so default config write virgin to disk, changes on tit write to disk, disk always win (if valid)"*.
46. **The JFS shape.** ARCHITECT: *"dont overcomplicate. no handroll. basically default carried config.md is identical to write on disk. the different is runtime can write to disk. carried binary serves as default tree, when disk invalid, so runtime never failed. this is the pattern isomorphic with JFS settings, only JFS use xml."*
47. **The file name.** ARCHITECT: *"`user.md` it is"*. The file of ruling 45 is `~/.config/tit/user.md`; the carried default is `Source/user.md`.
48. **Part 13 closes.** ARCHITECT, session 2026-11-14: *"no open items"*. The text `Email must hold @ and .` stands. The user-name check and the bare-repository dialog of SPEC are not built (Part 13, Closed by a read, item 24).
49. **The release lane as it works.** Question: the x64 and arm64 installers get one archive name, `titc v<version> Windows.exe`, and the second upload replaces the first. ARCHITECT, session 2026-11-14: *"just follow the working pattern"*. The `## pack` rows stay as the archetype generates them (Part 14).
50. **The braille dot table is cast data.** ARCHITECT, session 2026-11-14: *"braille dot table should not be hand-authored use LUT with cast table manifest"*. Part 15 Step 1 moves the table of `TerminalGlyph::drawBraille` to `## brailleDotColumns` and `## brailleDotRows` of `jam/cast/graphics.md`, generated as `jam::LookupTable` in the `os2UnicodeRange` form.
51. **The count of ruling 22.** Question, session 2026-11-14: the tracker said "three other texts", and `PLAN-tit-2-spec.md`, section "The exact text of the new items", holds four: the header text `PINNED @ {hash}`, the prompts `Search history:` and `Function history:`, and the no-result text `No commit matches`. Selected: *"All four stand"*.
52. **The tick is a deadline thread.** Question, session 2026-11-14: "Is the tick of the TIT terminal chain a deadline thread?" Selected: *"Yes, deadline thread"*. So the tick does not allocate, lock, wait, or make a blocking call (MANIFESTO, Deadline threads). The terminal output is a blocking descriptor, so the tick writes nothing to the terminal (Part 3 Step 5; Part 7 Step 5).
53. **The architecture.** ARCHITECT, session 2026-11-14, verbatim: *"ling is ongoing progress. tit is our first CLI. still the ISOMORPHIC ABSTRACTION holds. all of our project built with AUDIO HARD DEADLINE LOCK-FREE ARCHITECTURE where can enforce the DESIGN BY CONTRACT. exactly like plugin with sdk, we have no control over native tty just like we have no control over tick, thread, lock etc on each plugin format. but the DESIGN, the contract HOLDS. our hot path always lock free, message thread is free to run fully event driven synchronously, asynchronously, parallel on bg, etc it does not matter, concurrency is guaranteed by the design, thread safe, but must be very careful with real-time since it's not always safe. hence nothing need to be reinvent, CLI just another event loop deferred as long as the app running. at surface our CLI coded exactly like juce gui app, juce native rendering contract with juce::Component, LookAndFeel, Graphics, etc. no poking layers. our pseudo renderer LowLevelGraphicsContext is Terminal Video writer, instead of writing pixel it outputs bytes to terminal. nothing new, no new pattern, no new semantics"*. Also: *"what's wrong with EVE architecture? it works. this terminal is EVE hosted on END. why need to second guess and reinvent new pattern?"*; *"we have built TERMINAL FULL STACK end-to-end, a fully working terminal. […] we have all the building blocks, the working established pattern"*. What this fixes in the plan:
    - The hot path is the tick of the `jam::VirtualDevice`. It reads the keys and the size, and it stores into Model rows and one lane. It is lock-free, and it makes no blocking call (ruling 52).
    - The message thread lays out, paints, and sends the frame. EVE states the same split: "the clock's tick reads and writes the PTY…; the message thread lays out and paints" (`eve/SPEC.md:37-39`).
    - The renderer is `TerminalGraphicsContext`, a `juce::LowLevelGraphicsContext`. Its output is the bytes of the VT lane. The peer hands the bytes to the owner's event map, and the owner writes them to `TerminalDisplay` in the same call, on the message thread (Part 5, owner contract; Part 7 Step 5).
    - A view of TIT is a `juce::Component` with a LookAndFeel, in the form of the views of END and EVE. No part adds a machine that EVE, END, or LING does not have.
54. **No restore at the start.** Question, session 2026-11-14: legacy runs a restore at its start when the time travel marker exists and HEAD is on a branch, or the marker is not valid (`legacy/app/app_constructor.go:131-146,230-231`; `legacy/app/app_init.go:12-98`): `reset --hard HEAD`, `clean -fd`, the checkout of the first branch, the apply and drop of the first stash. "Does the new TIT run this restore at its start?" Selected: *"No restore at start"*. Option text: "The plan stands. TIT discards no change of the user at its start. The stale marker is removed, and a remaining stash of TIT shows the DirtyOperation menu." This is an exception to ruling 3 (Part 12, Closed by a read, item 15).
55. **Availability is state, not the command manager.** ARCHITECT, session 2026-11-14: *"why active or inactive has anything to do with command manager? it's state machine, we're using jam::Model, ValueTree, registered callbacks as commandId invoke command manager"*. A command is allowed by the rows of the Model. `getCommandInfo` keeps the END form. Ruling 56 states how the rows reach the commands.
56. **The state drives; no callback asks.** ARCHITECT, session 2026-11-14, verbatim: *"Availability is a test of Model rows ??? this is smell. why test model rows? it's the other way around. Model state changes drive events to build menu, set footer text, etc. it's always event driven by listener"*. And: *"legacy built with go/lipgloss/ELM architecture. i have done my best to make it as far as i can get to workaround the language and framework. but with our current design, BLESSED is non-negotiable, we are not writing byte-to-byte to translation to C++, not translating old logic business with ELM. with jam we could build anything, but conflating the logic as old design is absolutely wrong design. we can get 100% parity features, but way more performant, more thread safe, and most importantly more DETERMINISTIC if we think from CONTRACT perspective first, DESIGN BY CONTRACT to achieve the same goal. For example: - init run git to build state machine, jam::Model build - listener react - UI build - user interaction drives Model - callback driven by user input and this is not new pattern, this is how ALL JAM PROJECT build including audio plugin. even EVE as terminal is end-to-end STILL AUDIO PLUGIN, can run shell inside DAW. as fast any format it built."* What this fixes in the plan:
    - The detection writes the `git` rows. A listener reacts. The listener builds what the user can reach: the menu items and the key sheet (Part 9 Step 7, item 3; Part 10 Step 16, items 3 and 4).
    - `perform` is the END body (`end/Source/ui/ENDView.cpp:83-91`). No line of `perform` and no callback asks the Model if a command is allowed.
    - A key of the user invokes a callback. The callback writes Model rows or tells the processor. The result comes back as Model rows, and the listeners react.
    - A command that acts on a selected row acts on that row as its data (ruling 40). It is not a state rule.
    - The plan carries the features of legacy. It does not carry the control flow of legacy.
57. **MVP: the two lanes.** ARCHITECT, session 2026-11-14, verbatim: *"BLESSED. unidirectional. MVP = Model, View, Processor. Processor the orchestrator, lock free. hot path, update state always via raw value atomic, events driven from parameterChanged() Model, SSOT the state machine View, message thread, could be async always bind, listen, attached as listener, react events driven from valuePropertyChanged(), juce::Value, never atomic"*. And: *"this pattern is not hyptothetical. this is battle tested working pattern, build audio plugin, build terminal as audio plugin, build audio plugin host, everything build with the same MANIFESTO, SSOT modules, SSOT framework, SSOT code crossplatform."*; *"contract violations is failure, OTOH contract adherence guaranteed HIGH CRITICAL PERFORMANCE, concurrent"*. The code that states the same contract: `end/ARCHITECTURE.md:898-919`; `eve/SPEC.md:775-825`, with ARCHITECT's *"everything should be Parameter, no strays ana manual state management allowed"* (`:796-797`). What this fixes in the plan:
    - Each state row that a processor writes is a `jam::Parameter` or a `jam::ParameterText`, made in the constructor of `TITModel` (`end/Source/end/ENDModel.cpp:8-9`; `end/Source/plugins/Nexus.cpp:35-39`).
    - A processor writes a row only with `jam::Model::storeValue`, from the thread that has the value (`eve/Source/EVETerminalProcessor.cpp:143-144,161-162`; `jam_data_structures/model/jam_Model.h:1002-1031`). It writes no tree, and it calls no `setValuesFrom` and no `juce::MessageManager::callAsync`.
    - A processor reacts to a row through a `jam::Model::ParameterAttachment` (`EVETerminalProcessor.cpp:68-72`). It is not a `juce::ValueTree::Listener` (`eve/Source/EVETerminalProcessor.h:30-32`). Today's listener of a parameter is `jam::ParameterBase::Listener::parameterValueChanged` (`jam_data_structures/model/jam_ParameterBase.h:51-62`); END's document calls it `parameterChanged`.
    - A view writes and reads the tree on the message thread, and reacts in `valueTreePropertyChanged`. The flush of the Model carries a stored row to the tree (`jam_data_structures/model/jam_Model.cpp:29-58,291-303`), and the Model carries a tree write of a view to the parameter (`jam_Model.cpp:14-21,364-378`).
    - A list or a stream goes through a lane and a count parameter, and the message thread builds the tree children from the lane (`EVETerminalProcessor.cpp:146-159,168-186,196-254`).
    - LING's `Manifest` uses the tree listener and `callAsync` (`ling/Source/Manifest.cpp:16-46`). LING is work in progress (ruling 53), so it is not the reference for a processor.
58. **Cold state is a swap file.** ARCHITECT, session 2026-11-14, verbatim: *"instead of using Mailbox, why not store the cold state as AnsiDocument swap file? isnt this the pattern of gui TextEditor with Terminal?"*. And: *"TerminalGraphics in EVE is actually at hot path since we're using it as terminal (video writer), so with our own TERMINAL, both reader writer is hot path, lock free - with TUI, TerminalGraphics could be message thread, since there was the event loop happen. but the reader path still at Processor, still hot path, still lock free, but cold state store as swap file"*. And on the frame tick: *"so we only need to implement isomorphic seam for vblank substitution for Terminal peer lane? dont overengineer, i believe this is isomorphic on how we use virtual audio device/clock for eve."* What this fixes in the plan:
    - The output of a git run is cold state. It goes through the chain of EVE: `jam::TerminalTTY`, the decoder `jam::TerminalProcessor`, `jam::TerminalBuffer`, `jam::TerminalSwap` (`eve/Source/EVETerminalProcessor.cpp:554-564,337-345`; `jam_terminal/document/jam_TerminalSwap.h:11-24`). The chain runs on the serial context of a `jam::VirtualDevice`, and it is lock-free (`jam_audio_devices/audio_io/jam_VirtualDevice.h:19-20,92-102`).
    - A `jam::TextEditor` reads the swap through `bindToProcessor` (`jam_gui/widgets/jam_TextEditor.h:144`; `jam_TextEditor.cpp:82-100`). The editor pulls at each vblank of its peer. `jam::TerminalComponentPeer` is the source of that vblank (`juce_gui_basics/windows/juce_ComponentPeer.h:525-528,681`; Part 5 Step 2), as `jam::VirtualDevice` is the source of the tick. No new type is necessary for it.
    - A processor makes no record tree from the output of git, and it sends no raw bytes to a view. MANIFESTO: "A parse operation gives a complete AST… Each consumer downstream receives complete data"; "Published data is immutable".
    - A number, a flag, and a short text stay parameters (ruling 57). State that the user edits stays rows of the view lane: the choice of a conflict version, the action, the order, and the message of a rebase row. The highlight of a menu is not such a row (ruling 61).
    - The plan is in this form now. The run chain is `TITRunProcessor` (Part 10 Step 10): one object for each pane that shows output of the tool. The console (Part 10), the text pane of the browser (Part 11, the chain `patch`), and the three version panes of the resolver (Part 12) are each a `ConsoleView` over one chain. The lists go through the lane of the records and the count row `numRecords` (Part 11 Step 7).
    - A set of rows with a selection is not cold state. Ruling 60 states its form.
59. **The bytes of a git run come from a PTY with no login wrapper.** Question, session 2026-11-14: the decoder reads only a `jam::TerminalTTY` (`jam_terminal/processor/jam_TerminalProcessor.cpp:66,122-133`), and a run needs the exit code of each line (Part 9 Step 5). Measured on macOS in the same session: `/usr/bin/login -fqlp <user> /bin/bash --noprofile --norc -c "exec -l /bin/sh -c 'echo ran; exit 7'"` prints `ran` and returns 0; without `login` the status is 7. So the login wrapper of `TerminalUnixTTY` (`jam_terminal/tty/native/jam_TerminalTTY_posix.cpp:101-129,163-165`) hides the exit code. "Which byte source feeds the decoder for a git run?" Selected: *"PTY, no login wrapper"*. Option text: "JAM step on TerminalTTY: a program run skips the macOS login wrapper, and close() returns the exit status that both backends already ask the OS for (posix :288,:313; windows :773-775). Full EVE chain. Progress lines arrive live on Windows (ConPTY)."
60. **One menu widget for every set of rows.** ARCHITECT, session 2026-11-14, verbatim: *"what do you mean by list? menu like? for history? browser? it should not use TextEditor. but Menu, same widget component for ALL MENU"*. What this fixes in the plan:
    - A set of rows with a selection is a menu: the main menu and its submenus, the commits and the files of the history browser, the branches of the picker, the files of the resolver, and the rebase list. One widget shows each of them.
    - The plan has two widgets today: `MenuView` (Part 10 Step 12) and `ListView` (Part 11 Step 8; Part 12 Step 8). They become one widget.
    - No menu is a `jam::TextEditor`. A `jam::TextEditor` with a swap shows only the output of a git run (ruling 58).
    - The rows of a menu are children of a tree of the view lane. A processor sends them through a lane and a count parameter (ruling 57, the last item). Ruling 61 states where the highlight is.
    - A command that acts on a selected row reads that row as its data (ruling 40; ruling 56). So the answer *"cursorLine row plus a git read"* of the same session, to a question about a list in a `jam::TextEditor`, is void. The premise of that question was ORACLE's, and this ruling removes it.
61. **A menu is a component in the cell grid, in the contract of `jam::PopupMenu::MenuWindow`.** ARCHITECT, session 2026-11-14, verbatim: *"you are asking wrong question, you are still translating liploss to our terminal. you must think this like juce::Component in cell grid. why handroll? can we use juce::Label? jam::Menu? or should we write dedicated TerminalMenu in parity of jam::Menu since we dont need to create any actual window?"*. And: *"why we need selection in Model only? what is key from key sheet? your mental model still lipgloss"*. And: *"jam::PopupMenu::MenuWindows / if it doesnt fit, write the tui component context, isomoprhic contract / menu is stateless, all object are stateless. dont conflate rendering state with state machine. this is BLESSED architecture, not ELM"*. The read of `jam_gui/menu/jam_PopupMenu.h` in the same session:
    - What is the contract. The item data is `juce::PopupMenu::Item`: `text`, `itemID`, `action`, `subMenu`, `commandManager`, `shortcutKeyDescription`, `colour`, `isEnabled`, `isTicked`, `isSeparator`, `isSectionHeader` (JUCE `juce_gui_basics/menus/juce_PopupMenu.h:139-189`). Each item is a `juce::Component`, `ItemComponent` (`:155`). The size and the paint come only from the LookAndFeel: `getIdealPopupMenuItemSizeWithOptions` (`:198`), `drawPopupMenuItemWithOptions` (`:208`), `drawPopupMenuBackgroundWithOptions` (`:807`), `getPopupMenuBorderSizeWithOptions` (`:218`). The window moves its own highlight in `keyPressed`: up, down, left, right, return, space, escape (`:956-1013`); `selectNextItem` wraps and skips an item that cannot be triggered (`:1680-1714`). A trigger gives the item: its `action`, its `commandManager`, or its `itemID` (`:870-892,1667-1671`). With a parent component the window is a child and makes no desktop window (`:657-660`).
    - What does not fit a cell grid, where one unit is one cell (Part 10 Step 11, item 4). The sizes in units: `itemW { 80 }`, `itemH { 16 }` (`:181-184`); `parentArea.getHeight() - 24`, `parentArea.getWidth() - 24` (`:1247,1253`); `bufferHeight { 30 }` (`:1261`); `- 32`, `- 4`, `+ 6` (`:1286-1317`); `scrollZone { 24 }` (`:79`). The desktop: `juce::Desktop::getInstance().getDisplays().getDisplayForPoint` (`:1221`), `addGlobalMouseListener` (`:682`), a 20 Hz timer on `source.getScreenPosition()` (`:382,417-427`), and the dismiss when the process is not in the foreground (`:462-469,1195-1212`). The life: a trigger calls `exitModalState` and sets `exitingModalState` (`:879-883`), after which `windowIsStillValid()` is false (`:1107`); escape calls `setVisible (false)` (`:885-886,930`); a submenu enters the modal state (`:1659`); `setWantsKeyboardFocus (false)` (`:648`).
    - So it does not fit, and the second sentence of the ruling applies. JAM gets a menu component for the terminal, with the same contract: the same item data, the same LookAndFeel calls, an item component for each item, the highlight moved by the component's own `keyPressed`, and the trigger that gives the item. It has no desktop call, no modal state, and no size constant in pixels. ARCHITECT's word for it is `TerminalMenu`.
    - The highlight and the scroll offset are rendering state of the component, as `currentChild` and `childYOffset` are (`:1793-1794`). They are not rows of the Model. The state machine is the Model: its rows drive the event that builds the items (ruling 56: *"Model state changes drive events to build menu"*).
    - A later read of the same session corrects the third item: with a parent component the window of `MenuWindow` is a child component and makes no desktop window (`jam_gui/menu/jam_PopupMenu.h:640,657-677`), and `jam::showAsync` shows it. So the stock menu fits, and the first sentence of the ruling applies. The plan has no `TerminalMenu`: Part 10 Steps 12 and 16 show the stock menu in the view, and JAM gets only the keys `j` and `k` (ruling 63). Parts 11 and 12 show each tree of records in the same menu.
    - What leaves the plan: the parameter `selectedIndex` (Part 10 Step 9), the commands `menuUp`, `menuDown`, and the Enter command of the menu with their rows in the key sheet (Part 10 Steps 7 and 16), the hand paint of `MenuView::paint` (Part 10 Step 12; Part 11 Step 9; Part 13 Step 7) and of `ListView::paint` (Part 11 Step 8), and `ListView::moveSelection`.
62. **`TITLookAndFeel : jam::StyleTerminal`.** Question, session 2026-11-14: "Which class owns the menu item size and paint for the cell grid?" ARCHITECT: *"dont we have StyleTerminal already?"*. And: *"correct pattern / TITLookAndFeel : jam::StyleTerminal / TITLookAndFeel from css to user config markdown, disk always win"*. The reads of the same session:
    - `class StyleTerminal : public StyleCustom` sets the default colour of each indexed colour id, and `setAppearance()` applies the `terminal` sheet and then the sheet that the style owns (`jam/jam_terminal/style/jam_StyleTerminal.h:14-17`; `jam_StyleTerminal.cpp:5-9`). It has no override of a popup menu method today.
    - ARCHITECT's statement on a JAM widget in cells, the Finder: *"a juce::Component, rendered with monospace font (the exact same font with EVE's terminal)"*; *"yes at jam. it has its own ValueTree, commandId, colourId, etc. isomorphic to TextEditor. style use is StyleTerminal, one Style for Finder and all its childern"* (`eve/SPEC.md:447-450`).
    - The project form: `class ENDLookAndFeel : public jam::StyleCustom, public jam::Instance<ENDLookAndFeel>, public juce::ValueTree::Listener`, which listens to the config tree for a live theme change (`end/Source/lookAndFeel/ENDLookAndFeel.h:10-25`).
    - The plan gives TIT `jam::StyleMarkdown styleMarkdown` (Part 7 Step 4, `PLAN-tit-7-model.md:417,528,532`; Part 10 Step 17; Part 13, `PLAN-tit-13-setup.md:94,166,1719,1734`). `StyleMarkdown` derives from `StyleMermaid` and not from `StyleTerminal` (`jam/docs/xml/classjam_1_1StyleMarkdown.xml:5`).
    What this fixes in the plan:
    - The LookAndFeel of the TIT views is one TIT class, `TITLookAndFeel`, derived from `jam::StyleTerminal`.
    - The size and the paint of a menu item for the cell grid are LookAndFeel methods of JUCE (`juce_gui_basics/menus/juce_PopupMenu.h:941-1053`). `jam::StyleTerminal` is the style of a JAM widget in cells (the Finder statement), so the menu of ruling 61 takes them from it. `TITLookAndFeel` holds only what TIT changes (CODING.md: "Project code configures only the difference").
    - The values of `TITLookAndFeel` come from the carried `style.css`, then from the user's config markdown on disk. The disk wins. This is the form of rulings 45 and 46 (*"disk always win (if valid)"*; *"carried binary serves as default tree, when disk invalid, so runtime never failed"*).
    - Parts 7, 9, 10, and 13 now name `lookAndFeel`, the `TITLookAndFeel` of Part 7 Step 4, with the constructor `setSheet (jam::StyleManager::getStyleSheet()); setAppearance (Id::light);`. `printHelp` of Part 15 keeps `jam::StyleMarkdown`: it prints the help guide one time in the LING form (`ling/Source/Help.cpp:5-19`), and it is not a view.
63. **The keys of the terminal menu.** ORACLE's report, session 2026-11-14: the contract of `MenuWindow` moves the highlight on the arrow keys only (`jam_gui/menu/jam_PopupMenu.h:956-965`), and legacy moves it on `k` and `j` too. ARCHITECT: *"j-k-up-down all of them for terminal menu stands"*. So `jam::TerminalMenu::keyPressed` takes four keys for the move: down and `j`, up and `k`. A typed `j` is the key code `J` with no modifier and the text character `j`; `J` has the shift modifier and the text character `J` (Part 4 Step 2, `getTextKeyPress`). So the component tests the text character, and `J` and `K` of the rebase list (ruling 21) go up to the key sheet. No row of the key sheet moves a menu (ruling 61).
64. **A menu item with more than one colour is a `jam::AttributedString`.** ORACLE's question, session 2026-11-14: a commit row has an attribute and a subject in two colours, and `juce::PopupMenu::Item` has one `text` and one `colour` (JUCE `juce_gui_basics/menus/juce_PopupMenu.h:139-177`). ARCHITECT: *"different colour menu item can use jam::AtrributedString"*. The read of the same session: `jam::AttributedString` is the owning sequence of `jam::AttributedChar` atoms; each atom holds its `colourId`, its `bgColourId`, and its font; `append (std::string_view text, AttributedChar attributes)` adds the atoms of a text with one set of attributes; the string is move-only (`jam_core/text/jam_AttributedString.h:10-26,60-76`). The terminal context draws attributed text in cells (`jam_terminal/graphics/jam_TerminalGraphicsContext.h`, `drawAttributedText`; Part 10, inventory, "The terminal paint facts"). So the columns of a row are runs of one `jam::AttributedString`, and no row is painted column by column.
65. **No state for a widget. The dialog is `TerminalMessageBox`, in the contract of `jam::MessageBox`.** ORACLE wrote, in the same session, that rulings 60 and 61 name menus and no statement names the dialog. ARCHITECT: *"WRONG"*; *"NO FUCKING STATE FOR WIDGET. this is same contract with Menu. we need MessageBox, TerminalMessageBox"*. So the rule of ruling 61 is the rule of every widget: the rendering state of a widget is in the widget, and it is not a row of the Model. The read of the same session:
66. **The buttons of `TerminalMessageBox` are `juce::TextButton`.** ORACLE's question, session 2026-11-14: legacy opens some dialogs with the second button highlighted, and `juce::MessageBoxOptions` has no field for the first highlighted button (JUCE `juce_gui_basics/windows/juce_MessageBoxOptions.h:80-104`). ARCHITECT rejected the question and wrote: *"juce::TextButton"*. With it, in the same message: *"no widget should be coupled with state machine"*; *"widget are dumb juce::Component, always styled with LookAndFeel"*; *"always being told"*; *"this is not ELM, this is BLESSED MVP: Model View Processor"*; *"we want legacy feature parity, better UI/UX, in OUR CURRENT ARCHITECTURAL CONTRACT. not translation"*. So: each button of the box is a `juce::TextButton` child. The highlighted button is the button with the keyboard focus, which is the focus of JUCE and no row and no member. The LookAndFeel paints it: `drawButtonBackground`, `getTextButtonFont`, `getTextButtonWidthToFitText`, `drawButtonText` (JUCE `juce_gui_basics/buttons/juce_Button.h:400-412`). A press is `onClick` (`juce_Button.h:223`). A key of a button is `addShortcut (const KeyPress&)` (`juce_Button.h:270`). The JUCE box that holds buttons gives them by index, `Button* getButton (int index) const` (`juce_gui_basics/windows/juce_AlertWindow.h:111`), and adds each with its keys, `addButton (name, returnValue, shortcutKey1, shortcutKey2)` (`juce_AlertWindow.h:96-99`). The owner tells a button to take the focus. No command, no key row, and no Model row moves a button.
    - The contract of `jam::MessageBox`: `showAlert (title, message, associatedComponent, callback, iconType)` and `showYesNoBox (title, message, associatedComponent, callback)`. Each one builds a `juce::MessageBoxOptions` with `withTitle`, `withMessage`, `withButton`, and `withAssociatedComponent`, and gives the result to a `std::function<void (int)>`: the index of the pressed button (`jam_gui/windows/jam_MessageBox.h:27-43`; `jam_MessageBox.cpp:12-40,137-141`).
    - It shows through `juce::NativeMessageBox::showAsync`, a native window (`jam_MessageBox.cpp:18,32`). That does not fit a cell grid, as the desktop calls of `MenuWindow` do not (ruling 61).
    - So JAM gets `TerminalMessageBox`: a component in the cell grid with the same contract. It takes the title, the message, and the buttons, it moves its own highlighted button, and it gives the index of the pressed button to the callback.
    What leaves the plan: `ConfirmationView` with its hand paint (Part 10 Step 14); the `confirmation` child of the Model with its rows `title`, `explanation`, `yes`, `no`, `yesCommandLines`, `noCommandLines`, and `selectedIndex` (Part 10 Steps 5 and 9); the commands `confirmationYes`, `confirmationNo`, `confirmationEnter` and their key rows (Part 10 Steps 5, 7, and 16); the `confirmationView` colour rows (Part 10 Steps 4 and 8). The callback of the box holds what a button runs.

### Closed by a read, with no ruling

These are ORACLE's reads of CONTRACT and of the code. ARCHITECT can reverse each one.

- **Config and theme format: markdown tables through cast.** Basis: the ruling of conformance (*"it needs to conform with other project structured with manifest, build with cast"*); END and EVE hold keys and display rows in markdown tables (`end/Source/config/keys.md:17-21`; `eve/Source/layout/default-eve.md:44-59`). `SPEC.md:1007` (theme XML) is replaced in Part 2.
- **Command availability is state of the Model (ruling 55).** The command manager maps a key to a command id and invokes the registered callback. It never decides if a command is allowed: `getCommandInfo` of END and of EVE only names the command, and `setInfo` clears the disabled flag of every command (`end/Source/ui/ENDView.cpp:78-81`; `eve/Source/EVEView.cpp:445-448`; `jam_gui/commands/jam_CommandManager.h:78-85`). So "Git State → Allowed Actions → Menu" (`SPEC.md:23`) is a table of rules over the rows of the Model, and the direction is from the state to the commands (ruling 56). A change of a row that a rule reads is an event. The event builds the menu items and the key sheet: `setKeyPresses` with the row of a command only while its rule holds. END replaces its sheet from an event in the same way (`end/Source/ui/ENDView.cpp:419`), `invokeKey` matches only the rows of the sheet (`jam_gui/commands/jam_CommandManager.cpp:43-63,96-109`), and legacy registers the key of a menu item only while the item is in the menu (`legacy/app/app_keys.go:225-240`). `perform` is the END body and reads no state. No part sets the JUCE flag of a command. Three details are ORACLE's reads:
    - The view builds the sheet, and registers itself and its children, in the forms of `ENDView.cpp:13` and `eve/Source/EVEView.cpp:466-487`. The application keeps only the registration of `quit` (Part 9 Step 8).
    - The edit commands of `jam::TextEditor` have a rule too: their rows are in the sheet only while the mode of TIT shows an editable editor. This takes the place of a read-only test inside JAM (Part 6, closed item 3; Part 10 Step 16, item 6).
    - The two rebase commands of rulings 20 and 22 keep their key, because the footer says why only when the key arrives. Each callback writes the ruled text to the `message` row (Part 12, Closed by a read, item 19).
- **`TerminalDisplay` lives in `jam_terminal`; `TerminalWindow` and `TerminalComponentPeer` live in `jam_gui/windows/`, beside `jam::Window`.** Basis: NAMES.md Rule 5, the `Terminal` prefix of `jam_terminal` (`jam_terminal/` file list), and the `Window` family of `jam_gui` (`jam_gui.h:124`). `TerminalWindow::bindToEvents` names `Id::write`, a word of `jam/cast/gui.md:99` that only `jam_gui` holds. `jam_gui` depends on `jam_terminal` (`jam_gui.h:17`), never the reverse. This read corrects the earlier row that put both classes in `jam_terminal`.
- **Pin marker and the way back.** Marker `.git/TIT_PIN`, in the family of `.git/TIT_TIME_TRAVEL` and `.git/TIT_DIRTY_OP` (`SPEC.md:124,499`). It stores the branch and the original HEAD, as the dirty marker does (`SPEC.md:352-356`). Unpin is `git reset --hard <original HEAD>`, as the dirty abort is (`SPEC.md:378-385`). Pin with a dirty tree runs the Dirty Operation Protocol (`SPEC.md:329-334`).
- **Windows terminal target: ConPTY.** Both Windows machines run END (`~/.config/WINDOWS-SETUP.md:11,13`). END's Windows backend is ConPTY (`jam_TerminalTTY_windows.h:5-8`). mintty gives a native binary pipes, not a console; it is not a target.
- **Model name `TITModel`.** Family: `ENDModel` (`end/Source/end/ENDModel.h:19`), `LINGModel` (`ling/Source/LINGModel.h:15`).
- **Binary name `titc`.** `PLAN-tit-cpp-port.md:246`: "`titc` during port ... Rename to `tit` deferred". No ruling of this session changes it.
- **A bracketed paste reaches `jam::TextEditor` whole, through `juce::TextInputTarget`.** A JUCE peer gives text that is not a key press to `ComponentPeer::findCurrentTextInputTarget()` (`juce_gui_basics/windows/juce_ComponentPeer.h:418`), whose `insertTextAtCaret` "Inserts some text" (`juce_gui_basics/keyboard/juce_TextInputTarget.h:79-80`). `jam::TextEditor` has no such base today (`jam_gui/widgets/jam_TextEditor.h:107-112`). Legacy TIT inserts a paste whole in an input (`___legacy___/internal/app/app_update_msg.go:88-96`), and ruling 3 keeps it. A paste as single key presses would run a command on each Return. Part 6 adds the base. The terminal chain of Part 7 calls `insertTextAtCaret` through the peer's `findCurrentTextInputTarget()` (Part 5, owner contract). `ctrl+v` and `alt+v` read the clipboard, as in legacy TIT (`handlers_global_keys.go:80-96`).
- **The kitty keyboard protocol is not pushed.** The legacy encoding covers every key that TIT binds (Part 4, "The keys that TIT binds today"). A lone ESC with no more input is the Escape key, as in crossterm, termwiz, and notcurses (Part 4, Protocol facts).

## Dependency & API Inventory

Every row is an existing piece, or a hole with the read that shows it. No row is new design.

### What is stale in the TIT tree

| Fact | Where |
|---|---|
| `jam/cmake` and `configure_app` do not exist. TIT calls them. | `tit/CMakeLists.txt:12-14,36`; whatdbg left the same build: `whatdbg/carol/SPRINT-LOG.md:400` |
| `jam_tui` and `jam_mermaid` do not exist. TIT links them. | `tit/CMakeLists.txt:53-54`; JAM has `jam_mermaid_diagram` |
| Ten view stubs derive from `jam::tui::Component` | `tit/Source/component/Menu.h:4` and nine siblings |
| `TitState` is a hand-written `juce::Timer` with atoms | `tit/Source/state/TitState.h:28,161-174` |
| 112 hand-written identifiers | `tit/Source/TitIdentifier.h` |
| 13 test files; eight include `<jam_tui/jam_tui.h>`; no test target | `tit/tests/*.cpp:1`; `tit/CMakeLists.txt` |
| `SPEC.md` names Go, Bubble Tea, Lip Gloss as the target | `tit/SPEC.md:4` |

### The project side

| Need | Existing piece | Where |
|---|---|---|
| Create the project | `cast INIT.md <target>` from the cli archetype, then `cast cast/spell.md` | `archetype/ARCHETYPE.md:61,92-98,147-149` |
| Console target | `juceTargetFunction` row = `juce_add_console_app` | `ling/project-info.md:48`; `ling/CMakeLists.txt:244` |
| The project's own words | `cast/identifiers.md`, `files.md`, `text.md`, generated into `Source/generated/` | `ling/cast/spell.md:13-32`; `ling/Source/generated/Identifiers.h:11-30` |
| Application class | `class Application : public juce::JUCEApplicationBase, private juce::ValueTree::Listener` | `ling/Source/Main.cpp:276-282,398-407` |
| State model | `class LINGModel : public jam::Model, public jam::Instance<LINGModel>, private jam::File::Watcher::Listener` | `ling/Source/LINGModel.h:15-18`; `LINGModel.cpp:5-16` |
| Model API | `createAndAddParameter`, `getParameter`, `storeValue` (any thread), `flush` (message thread) | `jam_data_structures/model/jam_Model.h:928,989,1014,1053` |
| File to Model | watcher with a coalesce window, then a property change message | `ling/Source/LINGModel.cpp:14,57,61-67`; `end/Source/config/ConfigModel.cpp:125-130` |
| Child process | `jam::Subprocess::launch (command, workingDir, onComplete, onChunk)`; stdout and stderr are merged | `jam_subprocess/subprocess/jam_Subprocess.h:59-95`; use in `ling/Source/CompilerManifest.cpp:152-159` |
| Worker to Model | pool job, then `juce::MessageManager::callAsync`; "No ValueTree reaches the pool thread." | `ling/Source/Processor.cpp:584-601`; `Processor.h:11-12` |

### The command side

| Need | Existing piece | Where |
|---|---|---|
| Key to command | `jam::CommandManager::invokeKey (key, componentState)`; rows select by `Id::mode` and `Id::layer` | `jam_gui/commands/jam_CommandManager.h:65`; `.cpp:43-63` |
| Key sheet | a `juce::ValueTree`; one row per command; `Id::value` holds the keys | `jam_CommandManager.h:39-42`; `.cpp:5-41` |
| Key tokens | a row value is split at spaces, so one key description has no space (`ctrl+c`). `cursor up`, `page up`, and the other JUCE names with a space cannot be a sheet token today. Part 10 Step 2 makes a quoted description one token: `"cursor up"`. | `jam_CommandManager.cpp:16`; JUCE `juce_gui_basics/keyboard/juce_KeyPress.cpp:84-105,183-236` |
| One key, several meanings | a row for each mode; the mode is a row of `map::EditorMode`; a component state holds the mode | `jam/cast/terminal.md:1913,2068`; `eve/Source/EVEView.cpp:89-94,617` |
| User key table | `## keys`, columns `key`, `type`, `value`, `mode`, `description` | `eve/Source/layout/default-eve.md:44-59` |
| Command names | a generated id and name table; `addCommands`, `setInfo` | `jam_CommandManager.h:84-98`; `eve/Source/EVEView.cpp:440-448` |
| Command to callback | `Registry::callbacks`, one lookup in `perform` | `jam_gui/registry/jam_Registry.h:95`; `eve/Source/EVEView.cpp:353-423,500-509` |
| Rebuild on a file change | `fileChanged` runs `addCommands()` again | `eve/Source/EVEView.cpp:466-493` |

### The terminal side: what exists

| Need | Existing piece | Where |
|---|---|---|
| Paint target | `class TerminalGraphicsContext : public juce::LowLevelGraphicsContext`; one device unit is one cell | `jam_terminal/graphics/jam_TerminalGraphicsContext.h:12-13,71,84` |
| One frame to stdout | `prepare`, context, `juce::Graphics`, `paintEntireComponent`, `process`, `writeToStream`, `fwrite` | `ling/Source/Help.cpp:16-34` |
| Frame gate | a frame is written only when `frameId != publishedFrameId` | `jam_terminal/buffers/jam_TerminalBuffer.h:397-405`; `.cpp:347-350` |
| Text as cells | `drawAttributedText` overrides; the glyph-id `drawGlyphs` is unsupported | `jam_TerminalGraphicsContext.h:220-247,323-330` |
| `juce::Graphics::drawText` to cells | a JAM patch routes the four text calls and `AttributedString::draw` to `drawAttributedText` | `jam/patch/juce-attributed-text-hook.patch` |
| The cell unit for a widget | a style sheet: "One unit is one terminal cell. The `fontSize: 1em` property sets the base font size to one cell" | `ling/Source/style.css:6-9`; `jam_markdown_graphics/style/jam_StyleMarkdown.h:157-167` |
| Cell geometry | `jam::Cell`, `Cell::Point`, `Cell::Rectangle`, `fromPixel`, `toPixel` | `jam_graphics/geometry/jam_Cell.h:27,48,214,647,676` |
| Size value | `jam::TerminalWinsize` | `jam_terminal/video/jam_TerminalWinsize.h:34-63` |
| The shape of a platform seam | `TerminalTTY`: no thread, `read` never blocks, `onBytesReady` on a serial context | `jam_terminal/tty/jam_TerminalTTY.h:6,60,116-137` |
| Read readiness, macOS | `dispatch_source_create (DISPATCH_SOURCE_TYPE_READ, ...)` | `jam_terminal/tty/native/jam_TerminalTTY_posix.cpp:208-210` |
| Read readiness, Windows | overlapped read and a thread pool wait | `jam_terminal/tty/native/jam_TerminalTTY_windows.h:129-136` |
| Input grammar | tokens by `map::AnsiTokenType`: text, CSI, OSC, control; CSI by final byte through `map::csiFunctions`; a cut token is kept as `pendingBytes` | `jam_terminal/processor/jam_TerminalProcessor.h:233-238,298-299` |
| Key vocabulary | six tables, key to bytes | `jam_terminal/keyboard/jam_TerminalKeyboard.h:328-544` |
| Window shape to mirror | `Window (std::unique_ptr<juce::Component> mainComponent, const juce::String& name, ...)`; `setContentOwned` | `jam_gui/windows/jam_Window.h:63-76`; `jam_Window.cpp:44` |

### The terminal side: the holes

| Hole | The read that shows it | Part |
|---|---|---|
| Bytes to `juce::KeyPress` | `TerminalKeyboard::map` takes a `KeyPress` and returns bytes (`jam_TerminalKeyboard.h:117`); no table is a `Bimap` | 4 |
| Raw mode | no `#include <termios.h>` in JAM; the posix backend opens a pty for a child (`jam_TerminalTTY_posix.cpp:17-26,260`) | 3 |
| Own terminal size and resize | no producer of `TerminalWinsize`; its doc cites `Display::resized()`, and no `Display` exists (`jam_TerminalWinsize.h:9-13`) | 3 |
| Own stdin reader | `TerminalTTY::open` takes a shell to spawn (`jam_TerminalTTY.h:95-97`) | 3 |
| Writers for alternate screen, cursor show and hide, bracketed paste, synchronized output | `TerminalEncoder` has `writeRow` and `writeEraseScreen` only (`jam_terminal/buffers/jam_TerminalEncoder.h:39,48`). Part 3 adds one writer, `writePrivateMode`, for the four DEC modes (`jam/cast/terminal.md:581,592-594`). An absolute cursor move has no caller: a frame starts with `writeEraseScreen`, which ends at home (`jam_TerminalEncoder.cpp:132-137`). | 3 |
| Damage | a changed frame writes every row (`jam_TerminalBuffer.h:397-405`). Part 5 keeps the gate as it is: each frame is erase screen and every row, inside synchronized output (mode 2026). | 5 |
| `isatty` | local to LING (`ling/Source/Main.cpp:68-75`). In Part 3, `TerminalDisplay::open` returns `false` when `tcgetattr` or `GetConsoleMode` fails, so JAM needs no second test. | 3 |
| Live peer | `VBlankAttachment` adds a listener only when `owner->getPeer()` is not null (`juce_VBlankAttachment.cpp:105-111`); `Help.cpp` has no peer | 5 |
| Mouse bytes to an event | `TerminalMouse::map` encodes only (`jam_terminal/mouse/jam_TerminalMouse.h:164`) | none: no ruling and no SPEC line names mouse input |

### The JUCE attach path (vanilla JUCE 8, all by read)

| Step | Fact | Where |
|---|---|---|
| Override point | `virtual ComponentPeer* createNewPeer (int styleFlags, void* nativeWindowToAttachTo);` is protected | `juce_gui_basics/components/juce_Component.h:2758-2761` |
| JUCE precedent | `AudioProcessorEditor::createNewPeer` returns a peer that is not an OS window | `juce_audio_processors/processors/juce_AudioProcessorEditor.cpp:239-246` |
| Attach | `addToDesktop` sets the flag, calls `createNewPeer`, `updateBounds`, `setVisible`, `repaint` | `juce_Component.cpp:689-726` |
| Peer registration | the constructor adds to `Desktop::peers` only | `juce_gui_basics/windows/juce_ComponentPeer.cpp:41-49` |
| Paint | `handlePaint (LowLevelGraphicsContext&)` runs `component.paintEntireComponent (g, true)` | `juce_ComponentPeer.h:352`; `.cpp:117,144` |
| Keys | `handleKeyPress (const KeyPress&)` calls `target->keyPressed` | `juce_ComponentPeer.h:394`; `.cpp:196-217` |
| Repaint | `Component` calls `peer->repaint (area)` | `juce_Component.cpp:1895-1902` |
| Focus | `peer->grabFocus()`, `peer->isFocused()`; `isShowing` is `! peer->isMinimised()` | `juce_Component.cpp:602-614,2957-2959` |
| Resize | the peer calls `handleMovedOrResized()` | `juce_ComponentPeer.h:335-339`; `.cpp:321-360` |
| Frame tick | `callVBlankListeners` is protected; the Linux peer uses a timer | `juce_ComponentPeer.h:525-526,681` |
| Displays | arithmetic only; a null display returns the input | `juce_gui_basics/detail/juce_ScalingHelpers.h:48-52`; `desktop/juce_Displays.cpp:109-110` |
| Mouse source | one source exists with no device | `juce_gui_basics/detail/juce_MouseInputSourceList.h:41-50` |
| Accessibility | returns early with no client; posts the handler's own element | `native/accessibility/juce_Accessibility_mac.mm:857-869`; `juce_Accessibility_windows.cpp:163-169,214-220` |
| Loop, macOS | `[NSApp run]`; in use by LING with no window | `juce_events/native/juce_MessageManager_mac.mm:323-351`; `ling/Source/Main.cpp:407` |
| Loop, Windows | a hidden-window message pump | `juce_events/native/juce_Messaging_windows.cpp:54-58,114-131` |
| Pure virtuals of the peer | 26 | `juce_ComponentPeer.h:169-703` |
| Other desktop components | the base `createNewPeer` gives a native window, so a dialog or a menu of TIT is a child of `TerminalWindow` | `native/juce_NSViewComponentPeer_mac.mm:3036-3039`; `juce_Windowing_windows.cpp:4604-4608` |

### The text editor side

| Fact | Where |
|---|---|
| `class TextEditor : public juce::Component, public jam::Model::Component<TextEditor>, public jam::Model::Object, public juce::ApplicationCommandTarget, ...` | `jam_gui/widgets/jam_TextEditor.h:107-112` |
| Every size comes from the LookAndFeel: "The LookAndFeel must be a `StyleCustom`" | `jam_TextEditor.h:102`; `jam_style/jam_StyleCustom.h:98-121` |
| Modes today: normal, visual, terminal, open, reader | `eve/SPEC.md:739-740`; `jam/cast/gui.md:257-273` |
| "No key changes a document." | `jam_TextEditor.h:51` |
| ARCHITECT's contract: *"TextEditor is VIM core parity, functionally. Normal, Insert, Visual must work as is, that's the CONTRACT."* | `eve/SPEC.md:359-361` |
| The edit machinery existed: operators `d c y x p P`, `dd cc yy`, `o O i a I A`, undo as line-granular `juce::UndoableAction` | `eve/PLAN-Grand-Scheme.md:1376-1380` |
| It was removed for the terminal history: "The TextEditor edit machinery and insert mode are deleted." | `eve/PLAN-Grand-Scheme.md:594-595` |
| The ruled edit design: "An edit is a creation"; `InsertLine` and `RemoveLine`; sheets `## normal`, `## insert`, `## visual` | `eve/PLAN-Grand-Scheme.md:1395-1403` |
| Insert caret: *"entering insert mode it changed to pipe cursor by default following nvim"* | `eve/SPEC.md:800-801` |
| Read-only is a Model row: `setReadOnly (bool)` writes `Id::isReadOnly` | `eve/PLAN-text-editor-absorption.md:475` |
| Line edit pieces in JAM: `Document::Index::insertLine`, `removeLine`; `Document::addChild`, `removeChild` | `jam_core/document/jam_Document.h:515,524,1121,1128` |
| Per-cell runs, per-line fills, child lines as columns | `jam_TextEditor.h:25,82-84,93-95` |
| Repaint is driven by a vblank | `jam_gui/widgets/jam_TextEditor.cpp:98` |
| EVE binds `i a I A o O` in NORMAL to the return to TERMINAL, through its own key sheet | `eve/SPEC.md:349-351` |

### The git side: child processes

Part 8 runs the state detection with `jam::Subprocess`. These facts are for the part that runs the first user command (Part 8, Closed by a read, item 6).

| Fact | Where |
|---|---|
| On macOS, the child gets the stdin of TIT, which is the raw terminal, and the environment of TIT: the child sets only stdout and stderr, then calls `execvp` | `jam_subprocess/subprocess/native/jam_ChildProcess_posix.cpp:71-84` |
| On Windows, the child gets no stdin handle, no console, and the environment of TIT | `jam_subprocess/subprocess/native/jam_ChildProcess_windows.cpp:18-30` |
| `jam::ChildProcess::start` takes no environment | `jam_subprocess/subprocess/jam_ChildProcess.h:44` |
| Legacy sets `GIT_TERMINAL_PROMPT=0` for each command, and `GIT_PROGRESS_DELAY=0` for a streaming command | `___legacy___/internal/git/execute.go:26-27,106-112` |
| On Windows, a read returns only when the buffer is full or the child ends. `jam::Subprocess` reads 4096 bytes at a time. So progress lines arrive late on Windows. | `jam_ChildProcess_windows.cpp:55-89`; `jam_Subprocess.cpp:101` |
| A chunk that ends with a carriage return is a replace chunk | `jam_Subprocess.cpp:153-156` |

### The git side: the reading contract

Each row is a reader's question and the git output that answers it. Legacy TIT answers the first row only.

| Reader's question | Output | Legacy TIT |
|---|---|---|
| What changed in this commit? | `show --name-status`, then the patch of a file | `___legacy___/internal/git/exec_query.go:63,90` |
| How big is each change? | `--numstat -z` | absent |
| Which characters changed in this line? | `--word-diff=porcelain` | absent; one colour per line (`___legacy___/internal/ui/textpane_render.go:136-143`) |
| Which function is this hunk in? | the hunk header text; `-W` | not shown |
| How did this file change over time? | `log --follow -p -- <file>` | absent; the unit is the commit (`___legacy___/internal/ui/filehistory.go:22-31`) |
| How did this function change over time? | `log -L :<funcname>:<file>` | absent |
| Who last changed this line? | `blame -w -M -C --porcelain` | absent. Not in TIT (ruling 39). |
| When did this text appear or vanish? | `log -S<string>`, `-G<regex>` | absent |
| What will a pull give? What does upstream lack? | `log --left-right A...B`; `diff A...B` | counts only (`___legacy___/internal/git/state_detection.go:95,117`) |
| What is the state of the tree now? | `status --porcelain=v2 -z` | Clean or Dirty (`SPEC.md:51-57`) |
| Was this code moved? | `--color-moved` | absent |

Sources of the forms: the official manual pages `diff-options`, `log`, `blame`, `status`, `diff`, `diff-format`.
One controlled study compares unified and split views: 12 people, no significant difference, less visual strain with unified (Sandoval Alcocer et al., IEEE Access 2023, abstract only).

### The legacy inventory that Part 2 carries (ruling 3)

| Set | Count | Where |
|---|---|---|
| Main menu items | 27 | `___legacy___/internal/app/menu_items.go:8-239` |
| Config and preferences items | 13 | `menu_items.go:249-306` |
| Confirmation types, plus five inline dialogs | 18 + 5 | `___legacy___/internal/app/confirm_dialog.go:11-28`; `confirm_dialog_render.go:119-159` |
| Application modes | 18 | `___legacy___/internal/app/modes.go:21-40` |
| Key bindings by mode | 17 modes | `___legacy___/internal/app/app_keys.go:36-150` |
| Action dispatchers | 33 | `___legacy___/internal/app/dispatchers.go:9-46` |
| LFS: install, fetch, checkout | 3 calls | `___legacy___/internal/git/lfs.go:37-47`; not in `SPEC.md` |
| Stash tracking file | 1 | `___legacy___/internal/config/stash.go:41`; not in `SPEC.md` |
| Copy Hash mode | 1 | `SPEC.md:989-991` |
| Text inputs: commit message, branch name, clone URL, email | 4 | `___legacy___/internal/ui/textinput.go:10-15`; `app_keys.go:49-62` |

A contradiction that Part 2 closes: `SPEC.md:763` says `ed25519`; the code runs `ssh-keygen -t rsa -b 4096` (`___legacy___/internal/git/ssh.go:37`).

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

Each numbered part is one sprint.

## Steps

### Step 1: Write the part plans in Status order
**Scope:** `PLAN-tit-1-project.md` to `PLAN-tit-15-banner.md` at the project root.
**Action:** Write each part with the plan-writer skill. Take every dependency from the inventory of this file. Add a row to the inventory first when a part needs a piece that this file does not list, with its file:line. Give every identifier in the part: a name from an established family with its citation, or a name that Canon holds.
**Validation:** Each step of a part names its files, its functions, and the existing piece it follows. No step holds a name that is neither in Canon nor in a cited family. No part holds an open question at its lock.

### Step 2: Remove the stale plan documents
**Scope:** `PLAN-tit-cpp-port.md`, `RFC.md`.
**Action:** @Engineer deletes the two files after ARCHITECT locks this PLAN.
**Validation:** No file at the project root names `jam_tui`.

### Step 3: Track
**Scope:** this file, section Status.
**Action:** Set the status of a part when it is written, when ARCHITECT locks it, and when its sprint is logged.
**Validation:** The Status table matches the files at the project root and `carol/SPRINT-LOG.md`.

## BLESSED Alignment

- **B — Bound.** `TerminalWindow` owns its peer, as `removeFromDesktop` deletes a peer (`juce_Component.cpp:758-762`). The Model owns the state. Each marker file has one owner.
- **L — Lean.** Each part is one responsibility. No part builds a feature that Canon or SPEC does not name (YAGNI). Mouse input is in no part.
- **E — Explicit.** The keys, the config, and the theme are tables that the user reads. A command that the tool would reject is inactive, never silent.
- **S — SSOT.** The Model is the one state. `cast/identifiers.md` is the one source of the identifiers. The key sheet is the one source of the bindings.
- **S — Stateless.** A view holds no state. `jam::TextEditor` keeps its state in its component node and the Model (`jam_TextEditor.h:54-59`).
- **E — Encapsulation.** Nothing above the peer names a terminal type. `jam_gui` includes `jam_terminal`, never the reverse (`jam_gui/jam_gui.h:79-83`). `jam::TextEditor` gets its unit from the LookAndFeel.
- **D — Deterministic.** The same repository state gives the same Model rows, the same active commands, and the same menu.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

### Closed by ARCHITECT's rulings

| Item | Ruling |
|---|---|
| The name of the peer class | 14 |
| The name of the object for the program's own terminal | 15 |
| The mode name for insert | 16 |
| The user's existing TOML files | 17 |

### Closed by rulings 26 and 27

1. **The lock.** Closed by ruling 26: all parts lock as one PLAN, after all 15 are written.
2. **`Builds/Ninja`.** Closed by ruling 27: Part 1 Step 1 removes it.

### Facts that execution measures first

- `addToDesktop` with a peer that is not a native view has no run on this machine. Each line on the path is read; none needs a native handle. Part 7 Step 8 measures it.
- `JUCEApplication::initialiseApp` calls `initialiseMacMainMenu()` on macOS (JUCE `juce_gui_basics/application/juce_Application.cpp:95-107`). A console process with this base has no run. Part 7 Step 8 checks that no Dock icon and no menu bar appear.
- A native Windows binary under END gets console handles: `ResizePseudoConsole()` "delivers a `WINDOW_BUFFER_SIZE_EVENT` to the child process" (`jam_TerminalTTY_windows.h:247-248`). Not run.
- A live, editable `jam::TextEditor` on a terminal has no run. The read-only one-frame use is proven (`ling/Source/Help.cpp:16-34`). Part 7 runs a live read-only editor. Part 10 runs the first editable one, the commit box.
- Windows console input under ConPTY: with `ENABLE_VIRTUAL_TERMINAL_INPUT`, `ReadConsoleInputW` gives the VT characters as key-down `KEY_EVENT_RECORD`s, one UTF-16 unit each, and a surrogate pair as two such records in sequence. Microsoft documents VT input for `ReadFile` and `ReadConsole` only (learn.microsoft.com `setconsolemode`), and window size records for `ReadConsoleInput` only (`window-buffer-size-record-str`). "conpty's keyboard input is fundamentally backed by VT sequences" (microsoft/terminal, `doc/specs/#4999`). No Microsoft source states the record shape. Part 3 builds `TerminalWindowsDisplay::read` on it. The first Windows run of Part 7 measures it.

## Addendum 2026-11-13: the inline image lane and the banner (rulings 23 to 25)

### The inline image lane, as it was built and tested (ruling 24)

The source is JAM revision `7fbf821ba` (2026-10-25). It is the parent of `9e2c77a81` (2026-10-26), the revision that removed the lane.
The sprint record is `jam/carol/SPRINT-LOG.md:2774-2838` (Sprint 167, "OSC 1337 Pixels, Diagram as Image").
File names below are paths in `jam_terminal/graphics/` at that revision.

| Piece | Body | Where at `7fbf821ba` |
|---|---|---|
| The sequence writer | `appendInlineImage (output, png, width, height)`: the OSC number `map::OSC::iterm2Image`, `TerminalITerm2Decoder::filePrefix`, the arguments `inlineKey` = `inlineValue`, `sizeKey`, `widthKey`, `heightKey`, a colon, `juce::Base64::convertToBase64`, a bell | `jam_TerminalGraphicsEngine.cpp:249-263` |
| The switch | the constructor `TerminalGraphicsEngine (bool newShouldUseInlineImages)`, the member `const bool shouldUseInlineImages`, `isUsingInlineImages()` | `jam_TerminalGraphicsEngine.cpp:265-268,394`; `.h:37,70,406` |
| The placement | `struct ImagePlacement { jam::Cell::Rectangle destination; juce::MemoryBlock png; }` | `jam_TerminalGraphicsEngine.h:386-397` |
| Adding a placement | `addImagePlacement (destination, image)`: `ImageLoader::getPng (image)`; one placement for each covered row; the PNG on the first row only | `jam_TerminalGraphicsEngine.cpp:429-442` |
| The row writer | `appendFrameRow`: the row span up to the placement column; on the first row the image, then cursor up by height − 1; on a later row cursor forward by the width | `jam_TerminalGraphicsEngine.cpp:533-559` |
| The lane choice | `drawImage`: the inline lane when the engine uses inline images, the clip contains the whole destination, no transparency layer is open, and the fill is opaque. | `jam_TerminalGraphicsContext.cpp:490-528` |
| The size reserve | `inlineImageHeaderBytes { 128 }`, `getNumBase64Bytes` | `jam_TerminalGraphicsEngine.h:374-383` |

What differs today. Part 3 puts each body where the right column says.

| Then | Today |
|---|---|
| `TerminalGraphicsEngine` | `TerminalBuffer` (`jam/carol/SPRINT-LOG.md:695`) |
| `ImagePlacement::png`, a `juce::MemoryBlock` | Two directions (ruling 29). INPUT: `TerminalBuffer::ImagePlacement` holds the number of the image in the record (`jam_terminal/buffers/jam_TerminalBuffer.h:46-57`), since `2b20e4f61` (2026-10-30). It stays. OUTPUT: Part 3 restores the placement with its PNG as `TerminalBuffer::InlineImagePlacement`. Both are one template, `Placement<Content>` (CODING.md, "No near-duplicate struct proliferation"). |
| Writers named `append…` in the engine | Writers named `write…` in `TerminalEncoder`. ARCHITECT: *"i dont like "append" verbs"* (`jam/carol/SPRINT-LOG.md:672`) |
| `appendFrameRow` | the VT row writer of `writeToStream (output, laf)`. It receives the placements of each row and asserts that none exists (`jam_TerminalBuffer.h:95-98,392-409`; `.cpp:393-404`) |
| `drawImage` with the inline lane | `drawImage` with another body (`jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:488-509`). Part 3 deletes that body (ruling 28). |
| The size reserve, used by the engine's `getNumBytes` | No caller. The VT lane writes into a `juce::MemoryOutputStream` (`ling/Source/Help.cpp:8,32`). `TerminalBuffer::getNumBytes` sizes the binary frame only (`jam_TerminalBuffer.h:107-128`). Part 3 does not restore `inlineImageHeaderBytes` or `getNumBase64Bytes` (the basis of ruling 28). |

### The banner (rulings 23 and 25)

| Need | Existing piece | Where |
|---|---|---|
| An SVG path as a `juce::Path` with its colour | `jam::Svg::getElementPath`, `jam::Svg::parseColour`, `jam::Svg::parseStyle`, `jam::Xml::getOrCreate` | `jam_graphics/geometry/jam_Logo.h:50,193,210-213` |
| The logo | ten `<path>` elements with `fill:rgb(...)` styles | `tit/Assets/tit_logo.svg` |
| One image for a whole drawing | "Diagram is an Image": one `juce::Image`, drawn with `Graphics::drawImage` | `jam/carol/SPRINT-LOG.md:2885-2888` |
| A component that shows an image | `juce::ImageComponent`. `jam::ImageComponent` is a filmstrip display: it divides the image height by the component height | `jam_gui/widgets/jam_ImageComponent.h:118-125` |
| The pixel lane of `drawImage` | the inline image lane, after Part 3 (ruling 28) | `jam_TerminalGraphicsContext.cpp:488-509` |
| Braille | none in JAM. The legacy algorithm: 2 by 4 dots for each cell, one colour for each cell | `tit/___legacy___/internal/banner/braille.go`; `svg_render.go` |
