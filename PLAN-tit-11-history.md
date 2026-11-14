# PLAN: TIT Part 11 — The History Browser

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part builds the history browser of TIT: one view over one source of records (ruling 40).

1. **One browser, three levels.** The commits; the files of the selected commit; the detail of the selected file, which is its diff. The user walks from the top down. The browser has one mode, `history`, and one set of widgets.
2. **One record shape, several feeds.** A feed is data, not a view: the repository, a search, the commits that a pull or a push gives, the working tree changes, and the commits that changed one function. The browser renders what the records hold. A command is active when the selected row holds what the command needs.
3. **The cache, instant (rulings 37 and 38).** After each detection, the pool job of Part 8 runs one child process for the last 100 commits with their files. The records go through a lane to the message thread, which builds one complete tree of the Model (ruling 57). The patch of the selected record is one run on a run chain (ruling 58): the tool prints it into a swap, and the text pane reads the swap. Each other feed is one pool job for one request.
4. **The two record levels are one menu (ruling 60):** the stock popup menu of JAM that Part 10 shows (Part 10 Steps 12 and 16). A commit is an item, and its files are the items of its submenu. `TITView` builds the menu from the records. The highlighted record becomes the selected record when a key of the browser arrives, and when the user runs the item. **One screen view, `HistoryView`,** owns the text pane and the search input, with their layout and their focus. `TITView` shows it in the modes `history` and `search`, and its commands tell it (Part 9, Overview, item 6).
5. **One text pane,** a read-only `jam::TextEditor` (ruling 5), for the detail: a `ConsoleView` of Part 10 over the run chain `patch`. It shows the bytes that the tool wrote, with the tool's own colours and word marks. The decoder of the chain reads those colours. TIT parses no diff, and no row of the Model holds a patch.
6. **The commands and keys of the browser,** and the menu rows `history`, `file_history`, `view_changes`, and `incoming_outgoing`, with the menu separator and the labels of a cache that loads.

Blame is not in TIT (ruling 39).

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 8 and 10. Step 1 adds rows to one cast table of JAM.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization, and parentheses for an initializer of the same container type; no `==` on strings; no out-parameter; no fake-carrier struct; no local struct; no anonymous namespace; no non-owning naked pointer as a member.
- CODING.md, "Uniform signature in the type" and "Consume the event payload".
- MANIFESTO L: no `if` chain longer than three branches; a longer decision is a table. No function body longer than 30 lines.
- MANIFESTO S (Stateless): the cache, the records, the selected record of each level, the focused pane, and the source of the records are Model rows. No view holds one. The highlight of the menu is rendering state of the menu, and it is no row (ruling 61).
- MANIFESTO E, Complete at Creation: the pool job parses each record one time and writes it whole into the lane. The message thread moves a complete feed into the Model in one step. No consumer splits or parses the bytes again.
- plan-writer rule: "No manual string parsing". The tool writes the colours and the word marks; `jam::AnsiDocument` reads them. The one split of this part is at creation, at separators that TIT itself puts in the output format.
- Ruling 7: a layout locks at execution. Each geometry number is a named constant. ARCHITECT changes a layout at the run of Step 13.
- ARCHETYPE.md §1: no step edits a generated file by hand. Each cast table keeps its `+---` separator lines.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build and Git: agents build only on ARCHITECT's explicit request. The fixture repositories need git commands; ARCHITECT runs them, or gives the git instruction.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`. The words of the tool are from its official documentation, `Documentation/` of its source repository (`pretty-formats.adoc`, `diff-options.adoc`, `rev-list-options.adoc`, `config/color.adoc`, `git-log.adoc`, `git-diff.adoc`, `revisions.adoc`).

### Rulings of this part, verbatim

- **Ruling 37, the cache.** ARCHITECT: *"cache is preferrable, since it could speed up everything. unless you absolutely sure we can always build everything instantly."*; *"i believe this was also legacy defect, building even 30 caches could take couple of seconds, acceptable with go/lipgloss. absolutely not with jam."*; *"jam document parser could parse huge docs in subseconds"*; *"i believe we could now make the cache building instant"*.
- **Ruling 38, the count.** ARCHITECT: *"100 commit as default. we could expose for user config"*.
- **Ruling 39, no blame.** ARCHITECT: *"i dont want blame feature for tit"*; *"i have never need blame"*.
- **Ruling 40, one history browser.** ARCHITECT: *"there is only A `history` view, an SSOT of RECORDS. - History of whole repo in general - we could also dive in to see each commit per file. so there is no multiple view, all of them build what the same SSOT widget, feed by different data, rendered according to the data, user can interact according to the data. isnt that the whole premise of TIT?"*; *"History should be designed as BROWSER. hiearchial, we browse from top the commit, down to details per file."*; *"i said i want history that's the spec."*

Ruling 36 of an earlier draft of this part is void: its question stood on a sentence of ORACLE's Part 2 draft, not on a word of ARCHITECT.

### Measured in this session (2026-11-13, this machine, arm64)

| What | Result | How |
|---|---|---|
| Legacy cache cost | one process for each commit's metadata, one for each commit's file list, two for each file's diffs | `legacy/app/history_cache.go:79,145,160-176`; `legacy/git/exec_query.go:63,90,156-159` |
| 100 commits with the counts of each file, one process | JAM 0.40 s, 4214 lines; TIT 0.08 s; JUCE 0.09 s | `log -n 100 --format=… --numstat`, wall time, second run |
| 100 commits with every patch, coloured, words marked, one process | JAM 0.61 s, 38.0 MB, 492,606 lines; TIT 0.37 s, 11.9 MB; JUCE 0.10 s, 1.6 MB | `log -n 100 -p --color=always --word-diff=color` |
| `jam::AnsiDocument::addLines` | 258 ms for 7,154,643 bytes and 100,454 lines: 27.7 MB for each second, 2.57 µs for each line. A diff of 500 lines is about 1.3 ms. | scout, Release, 11 epochs, error 0.24 %; sandbox workload in the session scratchpad |
| What that time holds | The parse, with the line-break segments of each line (`AnsiDocument::addSegments`, 50 % of the samples). The layout reads those segments (`jam_gui/widgets/jam_TextBlock.cpp:159-170`). The workload ran no layout, so the time of a layout is not in the number. | scout sampling, 1 ms, 10,418 samples; `jam_terminal/document/jam_AnsiDocument.cpp:219-232` |
| The cache command of Step 2, item 4, as written, one process | JAM 0.94 s, 42.9 MB, 100 records, 4089 files; TIT 0.21 s, 14.3 MB, 100 records, 2494 files. Each record has four fields. Each file has its count line and its parts, in order; no part is left over. | wall time, second run; a check of every record |
| The function feed, `-L :addSegments:jam_terminal/document/jam_AnsiDocument.cpp` with the format of Step 2 | 3 records, one part each, no count line | one run in `jam/` |
| A search text with a space as one argument, `-Scell unit` | the tool takes it and lists two commits | one run in `tit/` |
| The first commit of a repository | `log -p` gives its patch: 29 files in `tit/` | one run with `--max-parents=0` |
| The date form `%aI` | `2026-04-22T02:32:05+07:00`. `juce::Time::fromISO8601` reads the date, `T`, the time, and a `+` or `-` offset with a colon. | one run; JUCE `juce_core/time/juce_Time.cpp:472-564` |

So the cache holds the records, which one process gives in 0.08 s to 0.40 s, and the patch of one record is one run of the tool on a run chain.

### The sample that fixes the split

Output of `log -n 1 -p --color=always --word-diff=color --format=%x1e%H` in `tit/`, with control bytes shown as `^^` (0x1e) and `^[` (escape):

```
^^dc4ff5d47ef2d13a0cfd0e634dd2a5a2caf267e9

^[[1mdiff --git a/.nvim-standalone-config b/.nvim-standalone-config^[[m
^[[1mnew file mode 100644^[[m
^[[36m@@ -0,0 +1,4 @@^[[m
^[[32m-- Standalone Build Configuration^[[m
```

- A record starts at the byte 0x1e that the format gives ("`%x` followed by two hexadecimal digits is replaced with a byte", `pretty-formats.adoc`).
- A file of a patch starts at the line `diff --git`. This part does not split a patch: the tool prints the patch of one file when the line names the file (Step 2, item 4, `showPatch`).
- With `--numstat`, one line for each file comes before the first file: added, a tab, deleted, a tab, the path. The sample of this session shows that form: `4	0	.nvim-standalone-config`. A rename shows as `old => new`; this part turns renames off in the cache with `--no-renames`.

### What the tool gives, and what this part uses

| Reader's question (tracker, the reading contract) | Output of this part | Documentation |
|---|---|---|
| What changed in this commit? How big is each change? | `log --numstat -p` | `--numstat`: "shows number of added and deleted lines in decimal notation and pathname without abbreviation ... For binary files, outputs two `-`" |
| Which characters changed in this line? | `--word-diff=color` | "Highlight changed words using only colors. Implies `--color`." |
| Which function is this hunk in? | the hunk header, as the tool writes it | `gitattributes.adoc`, "Defining a custom hunk-header" |
| How did this function change over time? | `log -L :<funcname>:<file>` | "Trace the evolution of the line range given by `<start>,<end>`, or by the function name regex `<funcname>`, within the `<file>`." |
| When did this text appear or vanish? | `log -S<string>` | "Look for differences that change the number of occurrences of the specified `<string>`" |
| What will a pull give? What does upstream lack? | `log HEAD..<upstream>` and `log <upstream>..HEAD` | `revisions.adoc`: "commits that are reachable from r2 excluding those that are reachable from r1" |
| What is the state of the tree now? | `diff --numstat -p HEAD`, and `ls-files --others --exclude-standard` | `git-diff.adoc`: "changes you have in your working tree relative to the named `<commit>`" |

- `--color=always`: with `color.diff` set to `always`, the tool "will use color for all patches" (`config/color.adoc`). A child of TIT has a pipe, not a terminal (tracker, "The git side: child processes").
- `--word-diff=porcelain` is the row of the tracker table. It is the same data in a form "intended for script consumption". This part takes `color`, which needs no parser (Closed by a read, item 3).

### The JAM pieces

| Piece | Where |
|---|---|
| The menu: `jam::showAsync` with a parent component. The item data is `juce::PopupMenu::Item`. | Part 10 Steps 12 and 16; `jam_gui/menu/jam_PopupMenu.h:1962-1986`; JUCE `juce_gui_basics/menus/juce_PopupMenu.h:139-189` |
| An item with a submenu: `juce::PopupMenu::Item::subMenu`. The right key opens the submenu and the left key closes it. The highlight stops on an item that can run or that has a submenu. | `jam_PopupMenu.h:107-116,966-998,1637-1665,1704` |
| A run item: the menu closes, the callback of `jam::showAsync` gets the `itemID`, and the `action` of the item runs on the message thread | `jam_PopupMenu.h:861-894,1667-1671` |
| A ticked item: `juce::PopupMenu::Item::isTicked`; the LookAndFeel paints the tick. The first highlighted item of the top level: `Options::withInitiallySelectedItem`. | JUCE `juce_PopupMenu.h:139-189`; `jam_PopupMenu.h:741-755` |
| The menu gives no call when its highlight moves: `setCurrentlyHighlightedChild` tells the item component and the accessibility handler only | `jam_PopupMenu.h:1618-1633` |
| An empty menu shows nothing, and its callback is deleted | `jam_PopupMenu.h:1910,1924` |
| A top window owns its submenu window, `activeSubMenu`, and resets it when it hides. Each window is in `getActiveWindows()` from its constructor to its destructor. | `jam_PopupMenu.h:787,794-798,867,1795` |
| `jam::Model::Component<T>::state` is public: `juce::ValueTree state;`. An owner assigns it, as `TextEditor` does: `statusBar.state = state;` | `jam_data_structures/model/jam_Model.h:188,191`; `jam_gui/widgets/jam_TextEditor.cpp:34` |
| An assignment of a `juce::ValueTree` keeps the listeners of the object and calls `valueTreeRedirected` | JUCE `juce_data_structures/values/juce_ValueTree.h`, `operator=` and `Listener::valueTreeRedirected` |
| `jam::AnsiDocument` reads SGR 1 (bold), 30 to 37 and 90 to 97 (foreground), 40 to 47 (background), 38 and 48 (256 colours and RGB), 7 (inverse), 0 (reset) | `jam_terminal/document/jam_AnsiDocument.cpp:967-1044` |
| The 16 colours of SGR come from the sheet: `--Ansi--colour1ColourId: var(--preciousPersimmon);` | `eve/Source/layout/style.css:74-90`; `jam/cast/style.md:596-611` |
| `TextEditor::setLiveBlock (std::unique_ptr<Document>)`: cursor 0, mode `reader`, read-only, view at the top. A read-only editor wraps its lines. | `jam_TextEditor.cpp:1130-1155,1208`; Part 6 Step 6 |
| In `reader` mode, `j k gg G ctrl+u ctrl+d` scroll, and `v V ctrl+v` start a visual mode. `y` is a row of `normal visual`. In `reader` mode, no row of the editor binds `tab`, `return`, `/`, or `F`. | `jam/cast/terminal.md:1922-1992,2077-2113` |
| The editor has no search command | `jam/cast/terminal.md:1875-1911` |
| The run form of a job: `pool.addJob`; the pool thread stores parameters with `storeValue` | Part 8 Step 7, item 1; Part 9 Step 5 |
| A list from a processor: a lane and a count parameter; the processor builds the tree children on the message thread when the view calls it | `eve/Source/EVETerminalProcessor.cpp:31-58,168-186,205-254`; `EVETerminalProcessor.h:120-126` |
| A pane of tool output: a `ConsoleView` over a `TITRunProcessor`, which is the TTY, the decoder, the buffer, and the swap | Part 10 Steps 10 and 14 |
| `juce::Time::fromISO8601`, `juce::Time::formatted`, `juce::SystemClipboard::copyTextToClipboard`, `juce::ValueTree::sendPropertyChangeMessage`, `removeChild`, `juce::String::replace`, `startsWith`, `getHexValue32`, `charToString`, `toStdString`, `juce::StringArray::fromLines`, `fromTokens (text, breakCharacters, quoteCharacters)`, `juce::File::loadFileAsString`, `juce::Rectangle::removeFromBottom` | doxygen `classjuce_1_1Time`, `classjuce_1_1SystemClipboard`, `classjuce_1_1ValueTree`, `classjuce_1_1String`, `classjuce_1_1StringArray`, `classjuce_1_1File`, `classjuce_1_1Rectangle` |

### The legacy views that this part carries (ruling 3)

| Piece | Legacy | Where |
|---|---|---|
| A commit row: the time as `02-Jan 15:04`, dimmed, a space, the short hash | | `legacy/ui/history.go:141-143`; `legacy/ui/listpane.go:281-282` |
| The details of a commit: `Author: `, `Date:   `, a blank line, the message | | `legacy/ui/history.go:199-206` |
| The file marks: `+` added, `-` deleted | | `legacy/ui/filehistory.go:110-119` |
| The lists take the top third of the height; the diff takes the rest, at the full width | `totalPaneHeight / 3` | `legacy/ui/filehistory.go:57-77` |
| The keys of History: up, `k`, down, `j`, Tab, Enter (time travel), `y`, `Y`, `ctrl+r`, Escape | | `legacy/app/app_keys.go:78-89` |
| The keys of File history: up, `k`, down, `j`, Tab, `y`, `v`, Escape | | `legacy/app/app_keys.go:90-99` |
| The list selection does not wrap | | `legacy/app/handlers_history_cache.go:16-65` |
| Copy hash: `y` the short hash of 7 characters, `Y` the full hash; the text `✓ Copied to clipboard` | `CopyHashShortHashLen = 7` | `legacy/ui/history.go:33`; `legacy/app/handlers_history_copyhash.go:12-37,102-111`; `legacy/app/messages_state.go:115` |
| Rewind: `ctrl+r`, the dialog `rewind` with the first button selected, then `reset --hard <hash>`, and `clean -fd` when the hash is HEAD | | `legacy/app/app.go:238-252`; `legacy/app/confirm_dialog_render.go:103-114`; `legacy/ui/confirmation.go:45`; `legacy/git/exec_timetravel_reset.go:24-37` |
| The menu rows `history` (`h`, 📜) and `file_history` (`f`, 📄) | | `legacy/app/menu_items.go:119-134` |
| The menu order: working tree, timeline, a separator when an item is above, the history items | | `legacy/app/menu_render_core.go:35-59` |
| The separator: 12 spaces, then `─` for 24 cells; the selection skips it | | `legacy/ui/menu.go:34-43`; `legacy/app/handlers_global_menu.go:213,237` |
| The items of a cache that loads: inactive, the spinner glyph as the emoji, `History...`, `Files...` | first spinner frame `⠋` | `legacy/app/menu_render_extra.go:18-59`; `legacy/ui/spinner.go:5-16` |
| The footer hints of the modes | | `legacy/app/messages_menu.go:70-109` |
| The colours of a diff: added `#5A9C7A`, removed `#B07070` | `diffAddedLineColor`, `diffRemovedLineColor` | `legacy/ui/theme_gfx.go:69-70` |

### Legacy faults that this part corrects

Ruling 4: "kept the feature FIX THE DESIGN, DESIGN BY CONTRACT."

1. **The cost of the cache** (ruling 37; measured above).
2. **The author was read and dropped.** The details show the literal `Author: Unknown` (`legacy/ui/history.go:199`; `legacy/git/exec_query.go:74,79`). This part shows the author.
3. **History and File history were two views of the same records.** History showed a commit and its message. File history showed a commit, its files, and one diff (`legacy/ui/history.go:100-127`; `legacy/ui/filehistory.go:22-31,45-78`). This part has one browser with the three levels (ruling 40).
4. **A commit row had no subject** (`legacy/ui/history.go:141-143`). This part adds the subject after the short hash.
5. **The diff dropped the hunk headers, and each line that starts with `---` or `+++`** (`legacy/ui/textpane_input.go:28-51`). This part shows the bytes of the tool.
6. **The first commit of a repository had no diff** (`legacy/git/exec_query.go:156`). `log -p` gives it.
7. **Enter on an empty list read past the end** (`legacy/app/handlers_history_cache.go:84,89`). Each command of this part reads a row through `getChild`, which gives an invalid tree for a wrong index (doxygen `classjuce_1_1ValueTree`).
8. **The order of commits with equal times was random** (`legacy/app/dispatch_git_history.go:17-27`). This part keeps the order of the tool.
9. **The copy texts never showed** (Part 10, legacy fault 1). The footer of Part 10 shows them.
10. **Escape in the visual mode left the view** (`legacy/app/app_keys.go:12,88,98,158-167`). The editor takes Escape in its visual mode (Part 10 Step 7, row `default mode`).

### Words the framework and the earlier parts already hold

| Word | Where |
|---|---|
| `Id::id`, `Id::value`, `Id::text`, `Id::mode`, `Id::message` | `jam/cast/core.md`, `jam/cast/gui.md` |
| `Id::commandLines`, `Id::emoji`, `Id::shortcut`, `Id::menu`, `Id::console`, `setConfirmation`, `setMenu`, `setView`, `setFocus`, `cancel`, `events`, `commands`, `availability`, `menuKeys` | Part 10 |
| `text::Git::resetHard`, `text::Git::clean`, `text::Git::remoteBranchPrefix`, `getOutput`, `getArguments`, `addJob`, `Id::hasUpstream`, `Id::currentBranch`, `Id::currentHash` | Parts 8, 9, 10 |
| `text::Header::aheadMark` (`⬆`), `text::Header::behindMark` (`⬇`), `text::Header::separator` (`─`), `text::Header::dirtyDescription`, the `text::Footer` rows of the console, the dialog, and the input | Part 10 Step 6 |
| `Chars::newline`, `Chars::space`, `Chars::tab`, `Chars::colon`, `Chars::escapeCharacter` | `jam/cast/core.md`; `jam_terminal/document/jam_AnsiDocument.cpp:75` |

### Closed by a read

ARCHITECT can reverse each one.

1. **A set of records with a selection is a menu (ruling 60), and the menu is the stock popup menu of JAM (Part 10 Step 12).** The commits are the items of the top level, and the files of a commit are the items of its submenu. The menu holds its highlight (ruling 61), and no key row moves it. The selected record is a Model row, because it is a choice of the user (ruling 58: "State that the user edits stays rows of the view lane"). It gets its value at two points. A key that the menu does not take arrives at `TITView::keyPressed`, which writes the highlighted record to the `selectedIndex` rows before the key sheet gets the key. Return on an item runs the `action` of the item, which writes the same rows. So each command of the browser acts on the row that the user sees highlighted, as in legacy, and each command reads only Model rows. The item of the selected record has a tick. The stock menu gives no call when its highlight moves (inventory). So the text pane shows a record after Tab or Return, and not at each move of the highlight as in legacy; the stock menu has no carrier for that.
2. **The detail is one read-only `jam::TextEditor` for every feed** (ruling 5). It reads the swap of the run chain `patch`, as the console of Part 10 reads the swap of its chain (ruling 58).
3. **The diff is the output of `--word-diff=color`.** It marks the changed words inside a line with the colours of the tool, and it keeps the hunk headers. The removed words and the added words of a line show on one line, with no `+` or `-` column. The tracker row names `--word-diff=porcelain`; that form needs a hand-written parser and a hand-built document. The look of the diff locks at execution (ruling 7): a change to another form is a change of the cells of Step 2, item 4.
4. **The colours of a diff are colours 1, 2, and 6 of the sheet's `--Ansi--` set.** The tool writes SGR 32 for added words and 36 for a hunk header (the sample), and 31 for removed words (the same run of this session). The values of 1 and 2 are the legacy diff colours.
5. **The cache is one tree of the Model, `history`, and one row, `isCacheReady`.** The pool job of the detection writes its records into the lane after the detection, and the message thread builds the tree. The cache holds no patch: each record holds the command line that prints its patch, and the run chain `patch` runs it for the selected record. One run of the tool for one file is the instant cache of ruling 37 (*"unless you absolutely sure we can always build everything instantly"*); Step 13 measures it.
6. **The four labels of a cache that loads are carried (ruling 3; Part 2 Step 4, item 3).** `History...` and `Files...` show until the first commit record arrives. Then `History 12/100` and `Files 12/100` show the count, as legacy shows its progress (`legacy/app/menu_render_extra.go:18-59`). The cache is one command, and its output arrives in chunks. The chunk callback counts the record separators and stores the count in the Model row `numCommits` from its own thread (`jam::Model::storeValue`, any thread, Part 7). The total is the count that the command asks for. Both items wait for the one cache, so both show the one count.
7. **The spinner shows its first frame, `⠋`, and does not turn.** The process of the cache takes 0.94 s on JAM and 0.21 s on TIT (measured). The turning spinner needs a frame row and a timer; Part 13 builds both for the sync spinner of the header (tracker, row 13).
8. **The count of ruling 38 is the cast row `historyCount`, value `100`.** Part 13 reads the user's value from the config tables of `~/.config/tit/` (tracker, row 13).
9. **The legacy `wip` diff, a commit against the working tree** (`legacy/git/exec_query.go:159`), **is the feed `view_changes` of ruling 19.** It is the working tree against HEAD, as one record of the browser.
10. **The incoming and the outgoing commits are two groups of the records.** The incoming records come first, each with `⬇`; the outgoing records follow, each with `⬆` (the marks of the header of Part 10). Ruling 19 says "as two lists"; the place of each group on the screen is a layout (ruling 7).
11. **An untracked file is a file of the working tree record, with the mark `+` and its content as its text.** Legacy counts an untracked file as a change (Part 8, `changedRecordTypes` `12?`), and `commit` adds it.
12. **Copy hash copies the hash of the selected record.** The legacy label mode chose one of ten visible hashes with a key (`legacy/ui/history.go:41-90`). A selected record is already that choice.
13. **Rewind is in this part.** It is a command of the browser with a dialog and one command line (`tit/SPEC.md:654-675`), and Part 10 gives both pieces. Time travel (Enter), pin (`p`), and interactive rebase (`i`) are protocols; Part 12 adds their rows to the mode `history`.
14. **The prompt is one `jam::TextEditor` in `insert` mode, and Enter submits in normal mode,** as the commit box (Part 10, Closed by a read, item 15). The search and the function history use the same prompt. The footer shows `Search history:` or `Function history:`; the second text has the form of the first, which Part 2 holds.
15. **A request that the tool rejects gives no record.** `getOutputText` returns the output for each exit code, and an output with no record separator gives no record. The browser then shows `No commit matches` and the cache (Step 10). This part adds no error text.
16. **The mode names:** `history` (legacy `ModeHistory`, `legacy/app/modes.go:26`) and `search` (Part 2, "Search history prompt"). The browser has one mode. `search` is the mode of its prompt, because a typed key must go to the editor of the prompt and not to a command of the browser.
17. **The header and the footer of Part 10 stay in view.** Legacy draws no header in History (`legacy/app/app_view_main.go:180-183`), and so it skips the size test there. The shell of Part 10 gives both.
18. **The menu rows of `view_changes` and `incoming_outgoing` have an empty emoji cell.** Ruling 22 gives their label, key, and hint, and no emoji.
19. **The two legacy menu items open the one browser.** Legacy File history is a commit, its files, and one diff (`legacy/ui/filehistory.go:22-31`): the three levels of the browser. So both items open the one browser, and the right key on a commit opens its files. Both items stay (ruling 3).
20. **Function history takes a typed function name, for the selected file.** Ruling 19: "The history of one function or line range, from the file view". The tool takes the name as `-L :<funcname>:<file>`. The file is the selected file of level 2.
21. **The tracker row "How did this file change over time?" (`log --follow`) is not a feed of this part.** It is a row of ORACLE's research table, and no ruling names it. One command row and one key row add it to the browser, as each other feed.
22. **Not in this part:** the menu rows of time travel (`time_travel_history`, `time_travel_files_history`) and the pinned menu, with Part 12; the user's commit count and the turning spinner, with Part 13.

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: JAM, the modes
**Scope:** `jam/cast/gui.md`, table `## EditorMode`; then `jam/generated/` through `cast cast/spell.md` from the JAM root.
**Action:**
1. `## EditorMode`: after the `input` row of Part 10 Step 4, add `history` 18 and `search` 19.
2. Run `cast cast/spell.md` from the JAM root.

**Validation:**
- `map::EditorMode` holds `history` and `search` with the keys 18 and 19. The rows 0 to 17 keep their keys.
- This part adds no colour id. The menu takes its colours from the `--PopupMenu--` customs of Part 10 Step 8, and the text pane takes its colours from the customs of `jam::TextEditor`.
- No JAM source file changes. Each mode name is in Closed by a read, item 16.

### Step 2: TIT, the words, the command ids, the command lines
**Scope:** `tit/cast/identifiers.md`, `tit/cast/spell.md`, `tit/cast/git.md`.
**Action:**
1. `cast/identifiers.md`, table `## identifiers`: after the last row of Part 10 Step 5, add these rows of type `@id` with an empty description. The value of each row is its name in backticks. Skip a name that JAM already declares.

   | Name | Family |
   |---|---|
   | `history` | the legacy cache and the menu id (`legacy/app/cache_manager.go:16-37`; `legacy/app/menu_items.go:119`) |
   | `reading` | the tracker section "The git side: the reading contract" |
   | `commit`, `file` | legacy `CommitInfo`, `FileInfo` (`legacy/git/types.go:92-96`) |
   | `attribute` | legacy list item field `Attribute` (`legacy/ui/listpane.go:281`) |
   | `selectedIndex` | the selected item of a menu: `getInitiallySelectedItemId`, "the selected item's ID" (`jam_gui/menu/jam_PopupMenu.h:741,1960`) |
   | `isCacheReady` | legacy `metadataReady`, `diffsReady` (`legacy/app/cache_manager.go:16-37`) |
   | `numRecords` | `numGridEvents` (`eve/Source/layout/parameters.md:47`); ruling 40, "SSOT of RECORDS" |
   | `patch` | the tool word for the text of a change (`git-log.adoc`, `-p`, "Generate patch"); the type of the run chain of the text pane, and the row of a record that holds the line that prints its patch |
   | `numCommits` | legacy `metadataProgress` (`legacy/app/menu_render_extra.go:21`); `numKeyPresses` (Part 7 Step 2) |
   | `focusedPane` | legacy field `FocusedPane` (`legacy/ui/filehistory.go:11-34`) |
   | `source` | `terminalSource` (`jam_gui/widgets/jam_TextEditor.h:1282`) |
   | `path` | legacy `file.Path` |
   | `argumentPrefix`, `argumentSuffix` | the parameter `argument` of `getArguments` (Part 8 Step 7); `remoteBranchPrefix`, `remoteConfigSuffix` (Part 8 Step 3) |
   | `historyLoading`, `fileHistoryLoading` | `pullMergeDiverged` (Part 10 Step 5): the menu row of a command in one state |

2. `cast/identifiers.md`: after the table `## Remote` of Part 8 Step 2, add the table `## Pane` in the form of that step, with the brief `@brief The panes of the history browser.` The rows: `records` 0, `text` 1. Add its output group to `cast/spell.md` in the form of Part 8 Step 4, item 3.
3. `cast/identifiers.md`, table `## CommandId`: after the `inputSubmit` row, add these rows. The value of each row is its name in backticks.

   | name | key | Source of the name |
   |---|---|---|
   | `history` | `0x4a00015` | `history` (`legacy/app/menu_items.go:119`) |
   | `fileHistory` | `0x4a00016` | `file_history` (`:127`) |
   | `viewChanges` | `0x4a00017` | `view_changes` (ruling 22) |
   | `incomingOutgoing` | `0x4a00018` | `incoming_outgoing` (ruling 22) |

   | `functionHistory` | `0x4a0001c` | ruling 19 |
   | `searchHistory` | `0x4a0001d` | ruling 19 |
   | `searchSubmit` | `0x4a0001e` | `inputSubmit` (Part 10) |
   | `copyHash` | `0x4a0001f` | `handleHistoryCopyHashEnter` (`legacy/app/handlers_history_copyhash.go:12`) |
   | `copyFullHash` | `0x4a00020` | `handleHistoryCopyHashFullEnter` (`:26`) |
   | `rewind` | `0x4a00021` | `handleHistoryRewind` (`legacy/app/app.go:238`) |

   Add also `historyTab`, key `0x4a0001b`, from `handleHistoryTab` (`legacy/app/handlers_history_cache.go:68`). The keys `0x4a00019` and `0x4a0001a` are free: the menu moves its own highlight (ruling 61), so no command does it.

4. `cast/git.md`: after the last row of Part 10 Step 5, add these rows. `FORMAT` stands for the text `--format=%x1e%H%x1f%aI%x1f%h%x20%s%x1d`, and `HEADFORMAT` for `--format=Author:%x20%an%nDate:%x20%x20%x20%aD%n%n%B`, each written in full in each cell. `OPTIONS` stands for `-c core.quotePath=false`, and `COLOUR` for `--color=always --word-diff=color`, both written in full.

   | name | value | description |
   |---|---|---|
   | `historyCount` | `` `100` `` | The count of the commits that the history cache holds. |
   | `historyLog` | `` `git OPTIONS log --no-renames --numstat FORMAT -n` `` | The commits of HEAD with the count lines of their files; the count follows. |
   | `rangeLog` | `` `git OPTIONS log --no-renames --numstat FORMAT` `` | The commits of a range with the count lines of their files; the range follows. |
   | `functionLog` | `` `git OPTIONS log FORMAT -L` `` | The commits that changed one function; the function and the file follow as one argument. |
   | `showPatch` | `` `git OPTIONS show --no-renames COLOUR HEADFORMAT` `` | Prints the head text and the patch of a commit; the hash follows, then `--` and a path for one file. |
   | `functionPatch` | `` `git OPTIONS log COLOUR HEADFORMAT -n 1 -L` `` | Prints the head text and the patch of one function in one commit; the function and the file follow as one argument, then the hash. |
   | `changes` | `` `git OPTIONS diff --no-renames COLOUR HEAD` `` | Prints the patch of the working tree; `--` and a path follow for one file. |
   | `changesCounts` | `` `git OPTIONS diff --no-renames --numstat HEAD` `` | The count lines of the changed files of the working tree. |
   | `untrackedPatch` | `` `git OPTIONS diff COLOUR --no-index -- /dev/null` `` | Prints a file that is not tracked as an added file; the path follows. |
   | `pathOption` | `` `--` `` | The option before a path. |
   | `untrackedFiles` | `` `git -c core.quotePath=false ls-files --others --exclude-standard` `` | The files that are not tracked and not ignored. |
   | `searchOption` | `` `-S` `` | The option of a search text; the text follows with no space. |
   | `incomingPrefix` | `` `HEAD..` `` | The start of the range of the commits that a pull gives. |
   | `outgoingSuffix` | `` `..HEAD` `` | The end of the range of the commits that a push gives. |
   | `upstreamName` | `` `@{u}` `` | The upstream of the branch. |

   | `recordSeparator` | `` `1e` `` | The byte that starts a commit record. |
   | `fieldSeparator` | `` `1f` `` | The byte between the fields of a commit record. |
   | `headSeparator` | `` `1d` `` | The byte that ends the fields of a commit record. |

   - The fields of a record, in order: the hash, the author date in strict ISO 8601 (`%aI`), and the short hash with the subject. The head text (author, date, message) is not a field: `HEADFORMAT` prints it at the start of each patch.
   - `--no-index`: "to compare the given two paths on the filesystem" (`git-diff.adoc`). With `/dev/null` as the first path, the file shows as an added file. Step 13 measures it on macOS, and Step 14 on Windows.
   - The three separator rows hold the number of a byte as two hexadecimal digits: the bytes of `%x1e`, `%x1f`, and `%x1d` of the format. `getSeparator` of Step 7 reads each with `juce::String::getHexValue32`.
   - `text::Git::upstream` of Part 8 is the command line `rev-parse @{u}`. `upstreamName` is the name alone.
   - `functionLog` has no `--numstat`: with `-L`, "Diff stat formats (`--stat`, `--numstat`, `--shortstat`, `--dirstat`) are not currently implemented" (`git-log.adoc`). Its argument is `:<funcname>:<file>`, the function name form of `-L`.

**Validation:**
- Each placeholder is in `pretty-formats.adoc`: `%H`, `%h`, `%s`, `%an`, `%aI`, `%aD`, `%B`, `%n`, `%xNN`.
- No value holds a space inside one argument: the format writes a space as `%x20`. `jam::ChildProcess::getArguments` splits at spaces (Part 8, API facts).
- The head text has the legacy labels `Author: ` and `Date:   ` with three spaces (`legacy/ui/history.go:199-200`).

- The map `commands` of Step 11 holds each new command id. The map `availability` holds each new command that a state rule governs. A command that acts on a selected record acts only when the record has the data that it needs (ruling 40).

### Step 3: TIT, the texts
**Scope:** `tit/cast/text.md`, `tit/cast/spell.md`.
**Action:**
1. `cast/text.md`: after the table `## confirmation` of Part 10 Step 6, add the table `## reading` in the same form, with the brief `@brief Texts of the history browser.` Add its output group to `cast/spell.md` in the form of Part 10 Step 6, item 5, with `@text:reading` and `- name: Reading`. The rows:

   | name | value | Source |
   |---|---|---|
   | `dateFormat` | `%d-%b %H:%M` | `legacy/ui/history.go:141` (`02-Jan 15:04`) |
   | `addedMark` | `+` | `legacy/ui/filehistory.go:110-119` |
   | `deletedMark` | `-` | `legacy/ui/filehistory.go:110-119` |
   | `copySuccess` | `✓ Copied to clipboard` | `legacy/app/messages_state.go:115` |
   | `noCommitMatches` | `No commit matches` | Part 2, "Other text" |
   | `hashPlaceholder` | `%s` | `legacy/app/messages_dialog.go:90-95` |
   | `historyProgress` | `History %s/%s` | `legacy/app/menu_render_extra.go:34` (`History %d/%d`) |
   | `fileHistoryProgress` | `Files %s/%s` | `legacy/app/menu_render_extra.go:49` (`Files %d/%d`) |

2. `## footer`: after the `input` row of Part 10, add these rows (`legacy/app/messages_menu.go:70-109`; Part 2, "Other text"):

   | name | value |
   |---|---|
   | `history` | `↑↓ navigate  │  Enter time travel  │  y copy hash  │  Tab details  │  Esc back` |
   | `readingText` | `↑↓ scroll  │  v visual  │  Tab commits  │  Esc back` |
   | `search` | `Search history:` |
   | `functionHistory` | `Function history:` |

3. `## confirmation`: add the four rows of the dialog `rewind`, with the prefix `rewind`, from `legacy/app/messages_dialog.go:90-95`, in the form of Part 10 Step 6, item 4.

**Validation:**
- Each value equals its legacy text or its ruled text.
- `text::Confirmation::rewindExplanation` holds `%s` one time.
- `shortHashLength`, the 7 of the legacy short hash, is a constant of `TITView` (Step 11), not a text.

### Step 4: TIT, the menu rows and the key rows
**Scope:** `tit/Source/menu.md`, `tit/Source/keys.md`.
**Action:**
1. `Source/menu.md`: after the rows of Part 10 Step 7, add these rows (`legacy/app/menu_items.go:119-134`; `legacy/app/menu_render_extra.go:18-59`; Part 2, "Menu rows"):

   | key | value | shortcut | emoji | hint |
   |---|---|---|---|---|
   | `view changes` | View changes | `v` | | View uncommitted changes file by file |
   | `incoming outgoing` | Incoming / outgoing | `o` | | View the commits a pull or push will transfer |
   | `history` | History | `h` | 📜 | View and navigate through commit history |
   | `history loading` | History... | `h` | ⠋ | View and navigate through commit history |
   | `file history` | File(s) history | `f` | 📄 | View history of specific files |
   | `file history loading` | Files... | `f` | ⠋ | View history of specific files |

2. `Source/keys.md`: after the rows of Part 10 Step 7, add these rows:

   | key | type | value | mode | description |
   |---|---|---|---|---|
   | `view changes` | `string` | `v` | `menuNormal` | View changes. |
   | `incoming outgoing` | `string` | `o` | `menuNormal` | Incoming / outgoing. |
   | `history` | `string` | `h` | `menuNormal` | History. |
   | `file history` | `string` | `f` | `menuNormal` | File(s) history. |

   | `history tab` | `string` | `tab` | `history` | Go to the next pane. |
   | `copy hash` | `string` | `y` | `history` | Copy the short hash. |
   | `copy full hash` | `string` | `Y` | `history` | Copy the full hash. |
   | `rewind` | `string` | `ctrl+r` | `history` | Rewind to the selected commit. |
   | `search history` | `string` | `/` | `history` | Search the history. |
   | `function history` | `string` | `F` | `history` | Show the history of one function of the selected file. |
   | `search submit` | `string` | `return` | `search` | Start the request. |

3. `Source/keys.md`, the `cancel` row of Part 10: add `history search` to its mode cell.

- The default keys are legacy keys (`legacy/app/app_keys.go:78-99`; `legacy/app/menu_items.go:119-134`) and ruled keys (rulings 21 and 22).
- A key of these rows reaches `TITView` in two ways. When the menu shows, it is modal, and a key that `MenuWindow::keyPressed` does not take goes up to `TITView::keyPressed` (Part 10 Step 16, item 2). When the text pane has the focus, the editor is in `reader` mode and binds none of these keys, so the key goes to its parent (Part 10, inventory, console facts).
- The menu moves its highlight on the arrow keys, `j`, and `k`, opens a submenu on the right key, closes it on the left key, and runs an item on Return (Part 10 Step 12; `jam_gui/menu/jam_PopupMenu.h:956-1013`). So this part has no key row for a move. The right key opens the files of a commit, and Tab goes to the text pane.
- `time travel` (Enter), `pin` (`p`), and `rebase interactive` (`i`) are rows of Part 12 in the mode `history`.

**Validation:**
- In one mode, each key has one row.
- Each mode name is a row of `map::EditorMode` after Step 1.
- Each menu text equals its legacy text or its ruled text.

### Step 5: TIT, the colours
**Scope:** `tit/Source/style.css`.
**Action:**
1. In the `:root` block, add the palette names `diffAdded` `#5A9C7A` and `diffRemoved` `#B07070` (`legacy/ui/theme_gfx.go:69-70`).
2. In the `.appearance` block, after the customs of Part 10 Step 8, add:

   | Custom | Palette name | Legacy field |
   |---|---|---|

   | `--Ansi--colour1ColourId` | diffRemoved | `diffRemovedLineColor` |
   | `--Ansi--colour2ColourId` | diffAdded | `diffAddedLineColor` |
   | `--Ansi--colour6ColourId` | caribbeanBlue | `accentTextColor` |

   The three `--Ansi--` customs have the form of `eve/Source/layout/style.css:75-80`.

**Validation:**
- Each value is the legacy value of its field (`legacy/ui/theme_gfx.go:9-82`). The hunk header takes the accent colour: legacy did not show it (legacy fault 5).
- After Step 13, no cell of the browser shows magenta, and a console line of a run that the tool colours red or green takes the same two colours.

### Step 6: TIT, the rows
**Scope:** `tit/Source/TITModel.cpp`.
**Action:** At the end of the constructor body, after the rows of Part 10 Step 9:

```cpp
    createAndAddParameter<jam::Parameter<int>> (state, Id::isCacheReady, 0);
    createAndAddParameter<jam::Parameter<int>> (state, Id::numRecords, 0);
    state.setProperty (Id::focusedPane, static_cast<int> (map::Pane::records), nullptr);
    state.setProperty (Id::source, jam::Format::toType (Id::history).toString(), nullptr);
    auto historyRow { getOrCreateChildWithName (jam::Format::toType (Id::history)) };

    historyRow.setProperty (Id::selectedIndex, 0, nullptr);
    createAndAddParameter<jam::Parameter<int>> (historyRow, Id::numCommits, 0);
    state.appendChild (juce::ValueTree { jam::Format::toType (Id::reading),
                                         { { Id::selectedIndex, 0 },
                                           { Id::commandLines, juce::String {} },
                                           { Id::argumentPrefix, juce::String {} },
                                           { Id::argumentSuffix, juce::String {} },
                                           { Id::path, juce::String {} } } }, nullptr);
```

The one record shape, which each feed of Step 7 builds:

| Tree | Rows | Children |
|---|---|---|
| `history`, `reading` | `selectedIndex` | one `commit` for each record |
| `commit` | `id` (the hash), `attribute` (the date), `value` (the short hash and the subject), `patch` (the line that prints the patch of the commit), `selectedIndex` | one `file` for each file |
| `file` | `id` (the path), `attribute` (the counts or the mark), `value` (the path), `patch` (the line that prints the patch of the file) | none |

- `history` is the cache: the feed of the repository. `reading` is the feed of the last request.
- `isCacheReady` and `numRecords` are parameters: the processor stores them (Step 7; ruling 57). In the loop of the run chains of Part 10 Step 9, add `Id::patch` to the list: the text pane has its own chain.
- No row holds a patch or a head text. They are output of the tool, and they go from the TTY of the `patch` chain to its swap (ruling 58).
- `source` holds the type name of the tree that the browser shows: `history` or `reading`.
- `selectedIndex` of a tree is the index of its selected child (Closed by a read, item 1). A new feed starts at 0 (Step 7, `setChildren`). The highlight of the menu is not a row (ruling 61).
- `focusedPane` is `records` when the menu of the records has the keys, and `text` when the text pane has them.
- `reading` also holds the request of the prompt: the command line, the text before and after the typed text in its argument, and the path of the file.

**Validation:**
- Each row holds value data: numbers, booleans, and texts. No row holds a document or a component (MANIFESTO E, Materialisation).
- A detection of Part 8 changes none of the rows of the view lane: it stores the rows of the `git` child only.

### Step 7: `TITGitProcessor`, the cache and the feeds
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:** A feed is a list, so it goes through a lane and a count parameter, and the message thread builds the tree children from the lane (ruling 57; ruling 60; `eve/Source/EVETerminalProcessor.cpp:146-159,168-186,196-254`). The patch of a record is output of the tool, so it goes through a run chain to a swap (ruling 58; Part 10 Step 10).

1. Header, public section, after `cancel`:

```cpp
    void read (const juce::String& commandLine, const juce::String& argument, const juce::String& path);
    void readIncomingOutgoing (const juce::String& upstream);
    void readChanges();
    void readPatch (const juce::String& commandLine);
    void applyRecords();
```

   After the public member `console` of Part 10: `TITRunProcessor patch;`. The constructor initializes it with `patch (modelToUse, Id::patch)`, and `prepare()` of Part 10 Step 10 gets the statement `patch.prepare();`.

   Private section, after `applyPendingCommandLine`:

```cpp
    void registerFeedCompletion();
    void addRecord (int level, const juce::Identifier& type, const juce::NamedValueSet& rows);
    void addCommits (std::string_view output, const juce::String& attributePrefix, const juce::String& patchLine, const juce::String& path);
    void addFiles (std::string_view counts, const juce::String& patchLine);
    std::string getHistoryText (const juce::File& directory);

    static void setChildren (juce::ValueTree target, juce::ValueTree source);
    static juce::ValueTree getRecord (const juce::StringArray& names);
    static std::vector<std::string_view> getRecords (std::string_view output);
    static juce::StringArray getFields (std::string_view record);
    static std::string_view getBody (std::string_view record);
    static std::pair<int, std::string> getResult (const juce::File& directory, const jam::Strings& arguments, const std::function<void (std::string_view)>& chunkCallback);
    static std::string getOutputText (const juce::File& directory, const jam::Strings& arguments);
    static char getSeparator (const char* hexadecimal);

    static constexpr int hashField { 0 };
    static constexpr int dateField { 1 };
    static constexpr int valueField { 2 };
    static constexpr int addedField { 0 };
    static constexpr int deletedField { 1 };
    static constexpr int pathField { 2 };
    static constexpr int numCountFields { 3 };
    static constexpr int recordsCapacity { 1048576 };
    static constexpr int recordLevel { 0 };
    static constexpr int childLevel { 1 };
    static constexpr int levelName { 0 };
    static constexpr int typeName { 1 };
    static constexpr int firstRowName { 2 };
    static constexpr int namesPerRow { 2 };
```

   Members, after `pendingFallbackCommandLines` of Part 10: `jam::TerminalFIFO records;`, `juce::ValueTree pendingRecords { jam::Format::toType (Id::reading) };`, and `jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, void ()> feedCompletion;`. In the constructor body, add `records.prepare (recordsCapacity);` and `registerFeedCompletion();`.

   The two functions that make command line text from the `git` rows are public, after `getReplaceLocalLines` of Part 10 Step 10: `static juce::String getUpstream (const juce::ValueTree& git);` and `static juce::StringArray getRewindLines (const juce::ValueTree& git, const juce::String& hash);`.

2. Source: the child process of a read. `getResult` takes the body of `getOutput` of Part 8 Step 7, item 9, up to `finished.wait();`, and ends with `return { exitCode, output };`. Its chunk lambda captures `chunkCallback` by reference and becomes `{ output.append (chunk); if (chunkCallback != nullptr) chunkCallback (chunk); }`. The callback runs on the worker thread of the child process (`jam_subprocess/subprocess/jam_Subprocess.h:11-19`), while the pool thread waits. Then replace `getOutput`, and add two functions:

```cpp
std::optional<juce::String> TITGitProcessor::getOutput (const juce::File& directory, const jam::Strings& arguments)
{
    const auto [exitCode, output] { getResult (directory, arguments, nullptr) };

    if (exitCode == 0) return juce::String { output }.trim();
    return std::nullopt;
}

std::string TITGitProcessor::getOutputText (const juce::File& directory, const jam::Strings& arguments)
{
    [[maybe_unused]] auto [exitCode, output] { getResult (directory, arguments, nullptr) };

    return output;
}

char TITGitProcessor::getSeparator (const char* hexadecimal)
{
    return static_cast<char> (juce::String { hexadecimal }.getHexValue32());
}
```

   - A read that gives rows is a child with a pipe, on the pool thread, as the detection of Part 8. Its output is parsed at creation and never shown. The output that a view shows goes through a run chain (item 6).

3. Source: the lane of the records. The writer is the pool thread.

```cpp
void TITGitProcessor::addRecord (int level, const juce::Identifier& type, const juce::NamedValueSet& rows)
{
    int numBytes { static_cast<int> (juce::String { level }.getNumBytesAsUTF8() + type.toString().getNumBytesAsUTF8()) + (namesPerRow + 1) * nullTerminatorBytes };

    for (const auto& row : rows)
        numBytes += static_cast<int> (row.name.toString().getNumBytesAsUTF8() + row.value.toString().getNumBytesAsUTF8()) + namesPerRow * nullTerminatorBytes;

    jassert (numBytes <= records.getFreeSpace());

    if (numBytes <= records.getFreeSpace())
    {
        records.writeString (juce::String { level });
        records.writeString (type.toString());

        for (const auto& row : rows)
        {
            records.writeString (row.name.toString());
            records.writeString (row.value.toString());
        }

        records.writeString ({});
        records.flush();
        model.storeValue (titTag, Id::numRecords, model.getParameter<jam::Parameter<int>> (titTag, Id::numRecords)->getValue() + 1);
    }
}

std::vector<std::string_view> TITGitProcessor::getRecords (std::string_view output)
{
    const auto recordSeparator { getSeparator (text::Git::recordSeparator) };
    std::vector<std::string_view> result;

    for (auto rest { output.substr (juce::jmin (output.find (recordSeparator), output.size())) }; not rest.empty();)
    {
        const auto end { juce::jmin (rest.find (recordSeparator, sizeof (char)), rest.size()) };

        result.push_back (rest.substr (sizeof (char), end - sizeof (char)));
        rest.remove_prefix (end);
    }

    return result;
}

juce::StringArray TITGitProcessor::getFields (std::string_view record)
{
    const auto headEnd { juce::jmin (record.find (getSeparator (text::Git::headSeparator)), record.size()) };

    return juce::StringArray::fromTokens (jam::Strings::fromView (record.substr (0, headEnd)),
                                          juce::String::charToString (static_cast<juce::juce_wchar> (getSeparator (text::Git::fieldSeparator))),
                                          {});
}

std::string_view TITGitProcessor::getBody (std::string_view record)
{
    const auto headEnd { juce::jmin (record.find (getSeparator (text::Git::headSeparator)), record.size()) };

    return record.substr (juce::jmin (headEnd + sizeof (char), record.size()));
}

void TITGitProcessor::addCommits (std::string_view output, const juce::String& attributePrefix, const juce::String& patchLine, const juce::String& path)
{
    const auto space { juce::String::charToString (Chars::space) };

    for (const auto record : getRecords (output))
    {
        const auto fields (getFields (record));
        const auto commitLine { patchLine + space + fields[hashField] };

        addRecord (recordLevel, jam::Format::toType (Id::commit),
                   { { Id::id, fields[hashField] },
                     { Id::attribute, attributePrefix + juce::Time::fromISO8601 (fields[dateField]).formatted (text::Reading::dateFormat) },
                     { Id::value, fields[valueField] },
                     { Id::patch, commitLine },
                     { Id::selectedIndex, 0 } });

        if (path.isEmpty())
            addFiles (getBody (record), commitLine);
        else
            addRecord (childLevel, jam::Format::toType (Id::file), { { Id::id, path }, { Id::value, path }, { Id::patch, commitLine } });
    }
}

void TITGitProcessor::addFiles (std::string_view counts, const juce::String& patchLine)
{
    const auto space { juce::String::charToString (Chars::space) };

    for (const auto& line : juce::StringArray::fromLines (jam::Strings::fromView (counts).trim()))
    {
        const auto fields (juce::StringArray::fromTokens (line, juce::String::charToString (Chars::tab), {}));

        if (fields.size() == numCountFields)
            addRecord (childLevel, jam::Format::toType (Id::file),
                       { { Id::id, fields[pathField] },
                         { Id::attribute, text::Reading::addedMark + fields[addedField] + space + text::Reading::deletedMark + fields[deletedField] },
                         { Id::value, fields[pathField] },
                         { Id::patch, juce::StringArray { patchLine, text::Git::pathOption, fields[pathField].quoted() }.joinIntoString (space) } });
    }
}
```

   - A record in the lane is a list of texts, each with its zero byte: the level, the type, then the name and the value of each row, then one empty text. `writeString` writes a text and its zero byte (JUCE `juce_core/streams/juce_OutputStream.h`, `writeString`). One record is one `flush`, so the reader never gets a part of a record (`jam_terminal/processor/jam_TerminalFIFO.h:37-63`).
   - The rows are a `juce::NamedValueSet`, the type of the rows of a `juce::ValueTree` (JUCE `juce_data_structures/values/juce_ValueTree.h`, the constructor with properties). So each feed names its own rows, and the lane has one form for every tree of records (Part 12).
   - The level is `recordLevel` for a record of the tree and `childLevel` for a record of the last record: a file of a commit.
   - The lane has the form of EVE's `gridEvents` lane: the assert of the free space, the positive test, the flush, and the count row in the `numGridEvents` form (`EVETerminalProcessor.cpp:168-186`). `nullTerminatorBytes` is `static constexpr int nullTerminatorBytes { 1 };` at the top of the file (`EVETerminalProcessor.cpp:6`).
   - `recordsCapacity` is 1 MiB. The cache of `jam/` is 4214 lines (Measured in this session), about a quarter of that. The message thread drains the lane at each flush of the count row, while the job writes. A record that finds no room is not written, as in EVE; Step 13 measures the largest feed.
   - `getRecords` gives the records of the tool: each one starts at the record separator. `getFields` gives the fields of the head of a record: the hash, the date, and the short hash with the subject. `getBody` gives the text after the head: the count lines, each with added, a tab, deleted, a tab, and the path (inventory, the sample). The three are views and texts of one output, made at creation.
   - The row `patch` of a record is the command line that prints its patch. For a commit it is the patch line of the feed and the hash. For a file it adds `--` and the path. The path is in double quotes, because the TTY of the chain keeps a quoted group as one argument (`jam_terminal/tty/native/jam_TerminalTTY_posix.cpp:253-256`).
   - The function feed gives a path: each record then has that one file, and the body of the record, which is a patch, is not read (Step 2, item 4: `-L` gives no count lines).
   - `juce::StringArray::operator[]` returns the empty text for an index out of range (doxygen `classjuce_1_1StringArray`). So a record with a missing field gives empty rows, and no test is needed.

4. Source: the reader of the lane. The reader is the message thread: `TITView` calls `applyRecords` at the event of `numRecords` (Step 11, item 4). Copy the static function `applyNames` of `eve/Source/EVETerminalProcessor.cpp:31-58` to the top of the file, verbatim.

```cpp
void TITGitProcessor::applyRecords()
{
    juce::StringArray names;

    applyNames (records, [this, &names] (const juce::String& name)
    {
        const auto isRowName { names.size() >= firstRowName and (names.size() - firstRowName) % namesPerRow == 0 };

        if (isRowName and name.isEmpty())
        {
            const auto record { getRecord (names) };

            if (feedCompletion.contains (record.getType()))
                feedCompletion.get (record.getType());
            else if (names[levelName].getIntValue() == childLevel)
                pendingRecords.getChild (pendingRecords.getNumChildren() - 1).appendChild (record, nullptr);
            else
                pendingRecords.appendChild (record, nullptr);

            names.clear();
        }
        else
        {
            names.add (name);
        }
    });
}

juce::ValueTree TITGitProcessor::getRecord (const juce::StringArray& names)
{
    juce::ValueTree record { juce::Identifier { names[typeName] } };

    for (int index { firstRowName }; index + 1 < names.size(); index += namesPerRow)
        record.setProperty (juce::Identifier { names[index] }, names[index + 1], nullptr);

    return record;
}

void TITGitProcessor::registerFeedCompletion()
{
    feedCompletion.add (jam::Format::toType (Id::reading), [this] (const juce::Identifier& type)
    {
        setChildren (model.state.getChildWithName (type), pendingRecords);
    });

    feedCompletion.add (jam::Format::toType (Id::history), [this] (const juce::Identifier& type)
    {
        setChildren (model.state.getChildWithName (type), pendingRecords);
        model.storeValue (titTag, Id::isCacheReady, 1);
    });
}

void TITGitProcessor::setChildren (juce::ValueTree target, juce::ValueTree source)
{
    target.removeAllChildren (nullptr);

    while (source.getNumChildren() > 0)
    {
        const auto child { source.getChild (0) };

        source.removeChild (0, nullptr);
        target.appendChild (child, nullptr);
    }

    target.setProperty (Id::selectedIndex, 0, nullptr);
    target.sendPropertyChangeMessage (Id::selectedIndex);
}
```

   - `applyRecords` is the form of EVE's `applyGridEvents`: the processor reads its lane and builds tree children, on the message thread, when the view calls it (`EVETerminalProcessor.h:120-126`; `.cpp:205-254`).
   - The names of one record are: the level, the type, then pairs of a row name and a row value. An empty text at the place of a row name ends the record. An empty row value is a value.
   - The records of a feed that arrives go to `pendingRecords`, which is no part of the Model. A record of `recordLevel` goes to its end, and a record of `childLevel` goes to the last record. The last record of a feed has the type of its tree, `history` or `reading`, and no row. Its entry in `feedCompletion` moves the children to that tree and sends one change message. So the Model gets a complete feed in one step (MANIFESTO E, "A complete state replaces a complete state"), and the views react to that one message (CODING.md, "Consume the event payload").
   - A row value arrives as a text. A reader of a number row converts it: `static_cast<int>` of a `juce::var` that holds a text gives its number (JUCE `juce_core/containers/juce_Variant.cpp`, `VariantType_String::toInt`).
   - `pendingRecords` is a member in the form of EVE's `pendingInput` (`EVETerminalProcessor.h:351`). Only the message thread reads and writes it. After a move it is empty.

5. Source: the cache. Replace the body of `addJob` of Part 8 Step 7, item 1:

```cpp
void TITGitProcessor::addJob()
{
    model.storeValue (titTag, Id::isCacheReady, 0);
    model.storeValue (jam::Format::toType (Id::history), Id::numCommits, 0);
    pool.addJob ([this, directory = juce::File { model.state[Id::directory].toString() }]
    {
        setGit (directory);

        if (model.getParameter<jam::Parameter<int>> (gitTag, Id::operation)->getValue() != map::Operation::notRepo)
            addCommits (getHistoryText (directory), {}, text::Git::showPatch, {});

        addRecord (recordLevel, jam::Format::toType (Id::history), {});
    });
}

std::string TITGitProcessor::getHistoryText (const juce::File& directory)
{
    const auto recordSeparator { getSeparator (text::Git::recordSeparator) };
    int numCommits { 0 };

    [[maybe_unused]] auto [exitCode, output] { getResult (directory, getArguments (text::Git::historyLog, text::Git::historyCount), [this, recordSeparator, &numCommits] (std::string_view chunk)
    {
        numCommits += static_cast<int> (std::count (chunk.begin(), chunk.end(), recordSeparator));
        model.storeValue (jam::Format::toType (Id::history), Id::numCommits, numCommits);
    }) };

    return output;
}
```

   - The job is the detection of Part 9 Step 5, `setGit`, then the cache. The detection stores the `git` rows first, so the menu does not wait for the cache.
   - Each commit record of the tool starts with the record separator (Step 2, item 4). So the count of the separators in the chunks is the count of the commits that have arrived. `storeValue` stores an atomic from any thread (`jam_data_structures/model/jam_Model.h:1002-1019`).
   - `numCommits` is a local of the pool job. The callback runs only while `getResult` waits, so the reference is valid.
   - `isCacheReady` is a parameter (Step 6). The job clears it, and the `history` entry of `feedCompletion` sets it after the tree has the records.

6. Source: the other feeds and the patch. Each feed is one pool job.

```cpp
void TITGitProcessor::read (const juce::String& commandLine, const juce::String& argument, const juce::String& path)
{
    pool.addJob ([this, commandLine, argument, path, directory = juce::File { model.state[Id::directory].toString() }]
    {
        const auto space { juce::String::charToString (Chars::space) };

        addCommits (getOutputText (directory, getArguments (jam::Strings::getView (commandLine), argument)), {},
                    path.isEmpty() ? juce::String { text::Git::showPatch } : text::Git::functionPatch + space + argument.quoted(), path);
        addRecord (recordLevel, jam::Format::toType (Id::reading), {});
    });
}

void TITGitProcessor::readIncomingOutgoing (const juce::String& upstream)
{
    pool.addJob ([this, upstream, directory = juce::File { model.state[Id::directory].toString() }]
    {
        const auto space { juce::String::charToString (Chars::space) };

        addCommits (getOutputText (directory, getArguments (text::Git::rangeLog, text::Git::incomingPrefix + upstream)), text::Header::behindMark + space, text::Git::showPatch, {});
        addCommits (getOutputText (directory, getArguments (text::Git::rangeLog, upstream + text::Git::outgoingSuffix)), text::Header::aheadMark + space, text::Git::showPatch, {});
        addRecord (recordLevel, jam::Format::toType (Id::reading), {});
    });
}

void TITGitProcessor::readChanges()
{
    pool.addJob ([this, directory = juce::File { model.state[Id::directory].toString() }]
    {
        const auto space { juce::String::charToString (Chars::space) };

        addRecord (recordLevel, jam::Format::toType (Id::commit), { { Id::value, text::Header::dirtyDescription }, { Id::patch, text::Git::changes }, { Id::selectedIndex, 0 } });
        addFiles (getOutputText (directory, jam::ChildProcess::getArguments (text::Git::changesCounts)), text::Git::changes);

        for (const auto& path : juce::StringArray::fromLines (juce::String { getOutputText (directory, jam::ChildProcess::getArguments (text::Git::untrackedFiles)) }.trim()))
            addRecord (childLevel, jam::Format::toType (Id::file), { { Id::id, path }, { Id::attribute, text::Reading::addedMark }, { Id::value, path }, { Id::patch, text::Git::untrackedPatch + space + path.quoted() } });

        addRecord (recordLevel, jam::Format::toType (Id::reading), {});
    });
}

void TITGitProcessor::readPatch (const juce::String& commandLine)
{
    patch.clearHistory();

    if (commandLine.isNotEmpty())
        patch.run (commandLine);
}
```

```cpp
juce::String TITGitProcessor::getUpstream (const juce::ValueTree& git)
{
    if (static_cast<bool> (git.getProperty (Id::hasUpstream))) return text::Git::upstreamName;
    return text::Git::remoteBranchPrefix + git.getProperty (Id::currentBranch).toString();
}

juce::StringArray TITGitProcessor::getRewindLines (const juce::ValueTree& git, const juce::String& hash)
{
    const auto currentHash { git.getProperty (Id::currentHash).toString() };
    juce::StringArray commandLines { text::Git::resetHard + juce::String::charToString (Chars::space) + hash };

    if (currentHash.isNotEmpty() and hash.startsWith (currentHash))
        commandLines.add (text::Git::clean);

    return commandLines;
}
```

   - Each feed writes its records, then the record of its tree. The tree keeps its old records until the new feed is complete.
   - `readPatch` shows the patch of one record: it empties the `patch` chain and runs the `patch` line of the record on it. The decoder reads the colours and the word marks of the tool, and the swap holds the lines (ruling 58). The text pane of the browser reads that swap (Step 10). A record with no command line gives an empty pane.
   - The patch of a commit record is the whole commit. The patch of a file record is that file. Both start with the head text: the author, the date, and the message (Step 2, item 4, `showPatch`).
   - The working tree changes are one record with no hash. Its text in the menu is `Local changes present`, the header text of a dirty tree (Part 10 Step 6). Its files are the changed files, and each untracked file with the mark `+`.
   - `getUpstream` gives the name of the upstream for the incoming and outgoing feed. `getRewindLines` gives the legacy lines of a rewind (inventory). `currentHash` is the short hash of HEAD (Part 8 Step 7), so the test is `startsWith`. Both are static members of the processor, as `getPushLine` is (Part 9 Step 5).
   - `argument` is one argument of the child, whole: a search text or a file path can hold a space (Part 8 Step 7, item 9, `getArguments`).
   - The pool has one thread (Part 8 Step 6). So the requests run in order, after a detection that runs.

**Validation:**
- No bail-out guard. No function body is longer than 30 lines. No `==` on strings.
- No pair member is read as `.first` or `.second`.
- The pool thread writes the lane and stores parameters. It reads and writes no tree, and it calls no `juce::MessageManager::callAsync` (ruling 57). The message thread builds the tree in `applyRecords`.
- The processor makes no tree and no text row from a patch. A patch goes from the TTY to the swap (ruling 58).
- Every feed gives the one record shape of Step 6. No feed has its own tree form.
- New names and their families: `read`, `readChanges`, `readIncomingOutgoing`, `readPatch` (`run`, Part 9; the tree `reading`; the chain `patch`); `applyRecords`, `records`, `pendingRecords`, `numRecords`, `recordsCapacity` (`applyGridEvents`, `gridEvents`, `pendingInput`, `numGridEvents`, `EVETerminalProcessor.h:126,351,354`; ruling 40, "SSOT of RECORDS"); `feedCompletion`, `registerFeedCompletion` (`treeConstruction`, NAMES.md Rule 8; the word "feed" of the Overview); `addRecord`, `addCommits`, `addFiles` (`addJob`; the row types; NAMES.md Rule 1, `add`: the lane is the container); `getRecord`, `getRecords`, `getFields`, `getBody` (`getGit`, Part 8; the fields of Step 2, item 4); `recordLevel`, `childLevel`, `levelName`, `typeName`, `firstRowName`, `namesPerRow` (the three levels of the Overview; `applyNames`); `setChildren` (`setFiles`, `ling/Source/Manifest.cpp:41-46`); `getResult`, `getOutputText`, `getHistoryText` (`getOutput`, Part 8); `getSeparator` (the three separator rows); the field constants (the fields of Step 2, item 4, and of `--numstat`); `patch` (the tool word; `console`, Part 10); `getUpstream`, `getRewindLines` (`getPushLine`, Part 9 Step 5; the command `rewind`). Local names: `currentHash`, `commandLines`, `recordSeparator`, `rest`, `end`, `record`, `result`, `headEnd`, `fields`, `row`, `rows`, `names`, `isRowName`, `commitLine`, `patchLine`, `line`, `counts`, `numBytes`, `name`, `child`, `path`, `space`.

### Step 8: The text pane
**Scope:** none. No file changes in this step.
**Action:** The text pane of the browser is a `ConsoleView` of Part 10 Step 14: one `jam::TextEditor` that reads the swap of a run chain (ruling 58). Its chain is `patch` (Step 7), and its row is the `patch` child of the Model (Step 6). This part has no list component: the records show in the stock menu (Step 11, item 5).

**Validation:**
- No `ListView.h` and no `ListView.cpp` exist. No code of TIT paints a row, moves a highlight, or holds a scroll offset.
- No view of this part makes a document from a text of the tool.

### Step 9: The separator and the item that cannot run
**Scope:** none. No file changes in this step.
**Action:** The menu is the stock popup menu of JAM (Part 10 Step 12). A separator is `juce::PopupMenu::addSeparator`, and an item that cannot run has `isEnabled` false (Step 11, item 5). The LookAndFeel paints both: `drawPopupMenuItemWithOptions` gets the item with `isSeparator` and `isEnabled` (JUCE `juce_gui_basics/menus/juce_PopupMenu.h:179-186,941-1053`). TIT has no `MenuView`, no colour id for a separator or an inactive item, and no paint code.

**Validation:**
- No `MenuView.h` and no `MenuView.cpp` exist.

### Step 10: `HistoryView`
**Scope:** `tit/Source/HistoryView.h`, `tit/Source/HistoryView.cpp` (both new).
**Action:** The browser is one screen, so it is one view class with its own components, its own layout, and its own listener (Part 9, Overview, item 6; `end/Source/ui/EditorView.h:25-28,194-218`). The menu of the records is not a member: `TITView` shows it, as it shows each menu (Step 11, item 5).
1. The header, with the includes of Part 10 Step 11, then `#include "ConsoleView.h"` and `#include "InputView.h"`:

```cpp
class HistoryView : public juce::Component,
                    public jam::Model::Component<HistoryView>,
                    private juce::ValueTree::Listener
{
public:
    HistoryView (juce::ValueTree stateToUse, jam::Registry::Getters& getters);
    ~HistoryView() override = default;

    void resized() override;
    void focusGained (FocusChangeType cause) override;

    void setSearchDocument();

    juce::ValueTree getSource() const;
    juce::ValueTree getCommit() const;
    juce::ValueTree getFile() const;
    juce::String getSearchText() const;

    static bool isHistoryMode (int mode);

    static constexpr int numPanes { 2 };

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    void registerEvents();
    void setFocus();

    static constexpr int searchRows { 2 };

    jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, void (juce::ValueTree&)> events;
    ConsoleView textPane;
    InputView searchInput;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (HistoryView)
};
```

2. The constructor: the form of Part 10 Step 11, item 2, with `textPane (stateToUse.getChildWithName (jam::Format::toType (Id::patch)), getters)` in the initializer list. Then `setWantsKeyboardFocus (true);`, `addAndMakeVisible (textPane);`, `addChildComponent (searchInput);`, and `registerEvents();`. The state is the root of the Model, as for `HeaderView`. `getters` are the getters of the `patch` chain (Step 11, item 2).
3. The layout and the focus:

```cpp
void HistoryView::resized()
{
    auto content { getLocalBounds() };
    const auto isSearch { static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::search };

    searchInput.setVisible (isSearch);
    searchInput.setBounds (content.removeFromBottom (isSearch ? searchRows : 0));
    textPane.setBounds (content);
}

void HistoryView::focusGained (FocusChangeType)
{
    setFocus();
}

void HistoryView::setFocus()
{
    if (static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::search)
        searchInput.grabKeyboardFocus();
    else if (static_cast<int> (state.getProperty (Id::focusedPane)) == map::Pane::text)
        textPane.grabKeyboardFocus();
    else
        grabKeyboardFocus();
}

bool HistoryView::isHistoryMode (int mode)
{
    return mode == map::EditorMode::history or mode == map::EditorMode::search;
}
```

   - The text pane takes the content of the view, and it writes its grid to the `winsize` row of the `patch` child (Part 10 Step 14). The menu of the records is a child of `TITView`, and `MenuWindow` gives it its size and its place over the content (`jam_gui/menu/jam_PopupMenu.h:657-660,1240-1253`). ARCHITECT sets the place at the run of Step 13 (ruling 7), with the fields of `juce::PopupMenu::Options`.
   - The prompt mode `search` shows the browser with one input under it.
   - With the pane `records`, the view itself holds the focus and the menu is modal. The peer gives each key to the menu first, and then to the parents of the menu (JUCE `juce_gui_basics/windows/juce_ComponentPeer.cpp:175-222`; Part 10 Step 16, item 2).
   - `grabKeyboardFocus()` on the component that holds the focus changes nothing, so the third branch is safe inside `focusGained`.

4. The three levels:

```cpp
juce::ValueTree HistoryView::getSource() const
{
    return state.getChildWithName (juce::Identifier { state.getProperty (Id::source).toString() });
}

juce::ValueTree HistoryView::getCommit() const
{
    const auto source { getSource() };

    return source.getChild (static_cast<int> (source.getProperty (Id::selectedIndex)));
}

juce::ValueTree HistoryView::getFile() const
{
    const auto commit { getCommit() };

    return commit.getChild (static_cast<int> (commit.getProperty (Id::selectedIndex)));
}
```

   - Level 1 is the tree that `source` names: the items of the menu. Level 2 is the selected `commit` row itself: its children are its files, the items of its submenu, and its `selectedIndex` is the selected file. So each record keeps its file. Level 3 is the patch of the selected record, which the `patch` chain prints (Step 7, item 6; Step 11, item 4).
   - The browser has no code for a feed. A record with no hash, a record with one file, and a record with no file are data, and the three levels show what they hold.

5. The prompt and the events:

```cpp
void HistoryView::setSearchDocument()
{
    searchInput.setDocument();
}

juce::String HistoryView::getSearchText() const
{
    return searchInput.getText();
}

void HistoryView::registerEvents()
{
    events.add (Id::mode, [this] (const juce::Identifier&, juce::ValueTree&) { resized(); });

    events.add (Id::selectedIndex, [this] (const juce::Identifier&, juce::ValueTree& tree)
    {
        const auto isEmptyFeed { tree.hasType (jam::Format::toType (Id::reading)) and tree.getNumChildren() == 0 and getSource() == tree };

        if (isEmptyFeed and isHistoryMode (static_cast<int> (state.getProperty (Id::mode))))
        {
            state.setProperty (Id::message, text::Reading::noCommitMatches, nullptr);
            state.setProperty (Id::source, jam::Format::toType (Id::history).toString(), nullptr);
        }
    });

    events.add (Id::focusedPane, [this] (const juce::Identifier&, juce::ValueTree&)
    {
        if (hasKeyboardFocus (true))
            setFocus();
    });
}

void HistoryView::valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property)
{
    if (events.contains (property))
        events.get (property, tree);
}
```

   - A feed that gives no record shows `No commit matches` and returns to the cache (Part 2, "Other text").
   - The focus moves between the panes only when the browser holds the focus, in the form of EVE's panel (`eve/Source/EVEView.cpp:204-223`). A command sets `focusedPane` before it sets the mode (Step 11, `setBrowser`), and the hidden view then takes no focus.
   - The search input is an `InputView` (Part 10 Step 14). The view gives its text through `getSearchText()`.

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- No member holds state: one console view, one input, one table. The browser is these components for every feed.
- `HistoryView` has no branch on a feed. A feed differs only in the tree that `source` names.
- `HistoryView` names no command id, no command line, no menu, and no processor. It holds no text and makes no document. It reads Model rows, and it writes the rows `message` and `source` of an empty feed.
- `searchRows` is one row of text and the status row of an editor that is not in `reader` mode (`jam_TextEditor.cpp:902`). `numPanes` is the count of the rows of `## Pane`.
- New names and their families: `HistoryView` (the `…View` family of Part 10, with the noun of ruling 40 and of the mode `history`); `setSearchDocument`, `getSearchText` (`setDocument`, `getText`, Part 10 Step 14; the mode `search`); `setFocus` (Part 10 Step 15); `getSource`, `getCommit`, `getFile` (the rows of Step 6; `getGit`); `isHistoryMode` (`isMenuMode`, Part 10); `textPane`, `searchInput` (`map::Pane`; the mode `search`). Local names: `content`, `isSearch`, `source`, `commit`, `isEmptyFeed`.

### Step 11: `TITView`, the browser commands
**Scope:** `tit/Source/TITView.h`, `tit/Source/TITView.cpp`.
**Action:** The functions of Part 10 Step 16 stay, except where an item below says "replace".

1. The header. Add `#include "HistoryView.h"` after `"InputView.h"`. Then:
   - Private functions, after `registerViewActions`: `void registerHistoryActions();`, `void registerBrowserActions();`, `void registerSearchActions();`, `void registerCancellations();`, `void registerRecords();`, `void registerSelections();`.
   - Private functions, after `cancelInput`: `void setBrowser (const juce::Identifier& source);`, `void setPrompt (const juce::String& commandLine, const juce::String& argumentPrefix, const juce::String& argumentSuffix, const juce::String& path, const char* message);`, `void copyText (const juce::String& text);`, `void setSelection (const juce::ValueTree& record);`, `void setPatch();`.
   - Private functions, after `isLargeEnough`: `bool isRecordMode() const;`, `juce::Component& getFocusComponent();`, `juce::PopupMenu getRecordMenu (const juce::ValueTree& tree);`, `juce::ValueTree getHighlightedRecord() const;`.
   - Constants, after `confirmMilliseconds`: `static constexpr int shortHashLength { 7 };`, `static constexpr int firstItemId { 1 };`.
   - Members, after `commitMessage`: `HistoryView history;`. After `events`: `jam::Function::Map<int, jam::Function::pointerCapacity, void ()> cancellations;`, `jam::Function::Map<int, jam::Function::pointerCapacity, juce::ValueTree ()> records;`, and `jam::Function::Map<int, jam::Function::pointerCapacity, void (const juce::ValueTree&)> selections;`.

2. The constructor: in the initializer list, after `console`, add `, history (modelToUse.state, gitProcessorToUse.patch.userInterfaceGetters)`. After `addChildComponent (commitMessage);`, add `addChildComponent (history);`. After `registerActions();`, add `registerCancellations();`, `registerRecords();`, and `registerSelections();`. In `registerActions()`, add at its end `registerHistoryActions();`, `registerBrowserActions();`, and `registerSearchActions();`.

3. The view of a mode. Replace `isMenuMode` and `setFocus`. In `resized`, after `commitMessage.setBounds (content);`, add `history.setBounds (content);`. In `setView`, after the `commitMessage.setVisible` line, add `history.setVisible (isLarge and HistoryView::isHistoryMode (mode));`.

```cpp
bool TITView::isMenuMode() const
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    return mode >= map::EditorMode::menuNotRepo and mode <= map::EditorMode::menuPinned;
}

juce::Component& TITView::getFocusComponent()
{
    const jam::HashMap<int, juce::Component*> components {
        { map::EditorMode::console, &console },
        { map::EditorMode::input, &commitMessage },
        { map::EditorMode::history, &history },
        { map::EditorMode::search, &history },
    };

    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    if (components.contains (mode)) return *components.at (mode);
    return *this;
}

void TITView::setFocus()
{
    getFocusComponent().grabKeyboardFocus();
}
```

   - `TITView` gives the focus to the view of the mode. The view then gives it to its own component (`ConsoleView`, `InputView`, Part 10 Step 14; `HistoryView`, Step 10).
   - The local table of `getFocusComponent` holds addresses of members for one call. It is not a member.
   - The `Id::mode` entry of `registerEvents` of Part 10 stays: it calls `setView()`. The browser lays itself out (Step 10).

4. The events. Add one entry to `registerEvents`:

```cpp
    events.add (Id::isCacheReady, [this] (const juce::Identifier&, juce::ValueTree&)
    {
        setMenu (getGit());
        setKeys (getGit());
    });

    events.add (Id::numCommits, [this] (const juce::Identifier&, juce::ValueTree&) { setMenu (getGit()); });

    for (const auto& id : { Id::selectedIndex, Id::source })
        events.add (id, [this] (const juce::Identifier&, juce::ValueTree&)
        {
            setMenu (getGit());
            setPatch();
        });

    events.add (Id::focusedPane, [this] (const juce::Identifier&, juce::ValueTree&) { setMenu (getGit()); });
    events.add (Id::numRecords, [this] (const juce::Identifier&, juce::ValueTree&) { gitProcessor.applyRecords(); });
```

   In the `Id::mode` entry of Part 10 Step 16, item 3, add the statement `setPatch();`. Then:

```cpp
void TITView::setPatch()
{
    if (HistoryView::isHistoryMode (static_cast<int> (state.getProperty (Id::mode))))
    {
        const auto file { history.getFile() };

        gitProcessor.readPatch ((file.isValid() ? file : history.getCommit()).getProperty (Id::patch).toString());
    }
}
```

   - The first entry builds the menu and the key sheet: the rule of `history` and of `fileHistory` reads the row `isCacheReady` (item 6; Part 10 Step 16, item 3).
   - The second entry shows the count of a cache that loads: each flush of the row `numCommits` builds the menu again, and the two loading rows take the new label (item 5). No rule reads that row, so the sheet stays.
   - The loop builds the menu of the records from the rows that it shows: the one change message of a feed that arrives (Step 7, `setChildren`), a new selected record, and a new source. It has the form of the loop of Part 10 Step 16, item 3. A new focused pane builds the menu too.
   - `setPatch` tells the processor to print the patch of the selected record: the selected file of the selected commit, or the commit when it has no file (Step 7, item 6). It runs at a new selection, at a new source, and when the mode changes, and only while the browser shows. The view gives the `patch` line that the record holds; it makes none.
   - The `numRecords` event tells the processor to read its lane (Step 7, item 4), as EVE's view does for the `numGridEvents` row (`eve/Source/EVETerminalProcessor.h:120-126`).
   - `isCacheReady` is a parameter now (Step 6). Its event is the first entry, with no change.
   - `isEditMode` of Part 10 Step 16, item 5: the mode `search` shows the prompt of the browser, an editable editor. Replace its body with `static const juce::Array<int> editModes { map::EditorMode::input, map::EditorMode::search };` and `return editModes.contains (static_cast<int> (state.getProperty (Id::mode)));`.

5. The menu. Replace `getMenuCommands` and `setMenu` of Part 10 Step 16, item 4, and add `getMenu`, `getMenuItem`, `getInitialItemId`, and the functions of the record menu:

```cpp
juce::Array<juce::CommandID> TITView::getMenuCommands() const
{
    static const juce::Array<juce::CommandID> operationCommands {
        map::CommandId::commit, map::CommandId::commitPush, map::CommandId::resetDiscardChanges, map::CommandId::viewChanges,
        map::CommandId::push, map::CommandId::pullMerge, map::CommandId::forcePush, map::CommandId::replaceLocal,
        map::CommandId::incomingOutgoing, map::CommandId::finalizeMerge, map::CommandId::abortMerge, map::CommandId::rebaseContinue,
        map::CommandId::rebaseAbort, map::CommandId::history, map::CommandId::fileHistory,
    };

    return operationCommands;
}

void TITView::setMenu (const juce::ValueTree& git)
{
    auto& windows { jam::PopupMenu::MenuWindow::getActiveWindows() };

    for (int index { 0 }; index < windows.size(); ++index)
        windows.getUnchecked (index)->dismissMenu (nullptr);

    if (isShowing() and isLargeEnough() and (isMenuMode() or isRecordMode()))
    {
        auto menu { isRecordMode() ? getRecordMenu (records.get (static_cast<int> (state.getProperty (Id::mode)))) : getMenu (git) };

        jam::showAsync (menu, juce::PopupMenu::Options().withParentComponent (this).withInitiallySelectedItem (getInitialItemId (menu)), [this] (int result)
        {
            if (jam::PopupMenu::MenuWindow::getActiveWindows().isEmpty())
            {
                if (result == 0 and isRecordMode())
                    commandManager->invokeDirectly (map::CommandId::cancel, false);
                else
                    setMenu (getGit());
            }
        });
    }
}

bool TITView::isRecordMode() const
{
    return records.contains (static_cast<int> (state.getProperty (Id::mode)))
       and static_cast<int> (state.getProperty (Id::focusedPane)) == map::Pane::records;
}

juce::PopupMenu TITView::getRecordMenu (const juce::ValueTree& tree)
{
    const auto selectedIndex { static_cast<int> (tree.getProperty (Id::selectedIndex)) };
    juce::PopupMenu menu;

    for (int row { 0 }; row < tree.getNumChildren(); ++row)
    {
        const auto record { tree.getChild (row) };
        juce::PopupMenu::Item item { (record.getProperty (Id::attribute).toString() + juce::String::charToString (Chars::space) + record.getProperty (Id::value).toString()).trimStart() };

        item.itemID = row + firstItemId;
        item.isTicked = row == selectedIndex;
        item.action = [this, record]
        {
            const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

            setSelection (record);

            if (record.isAChildOf (state) and selections.contains (mode))
                selections.get (mode, record);
        };

        if (record.getChild (0).hasProperty (Id::id))
            item.subMenu = std::make_unique<juce::PopupMenu> (getRecordMenu (record));

        menu.addItem (std::move (item));
    }

    return menu;
}

void TITView::setSelection (const juce::ValueTree& record)
{
    if (record.isAChildOf (state))
        for (auto child { record }; child.getParent() != state; child = child.getParent())
            child.getParent().setProperty (Id::selectedIndex, child.getParent().indexOf (child), nullptr);
}

juce::ValueTree TITView::getHighlightedRecord() const
{
    auto record { records.get (static_cast<int> (state.getProperty (Id::mode))) };

    for (const auto* window : jam::PopupMenu::MenuWindow::getActiveWindows())
        if (const auto* highlighted = window->currentChild.getComponent())
            record = record.getChild (highlighted->item.itemID - firstItemId);

    return record;
}

bool TITView::keyPressed (const juce::KeyPress& key)
{
    if (isRecordMode())
        setSelection (getHighlightedRecord());

    return commandManager->invokeKey (key, state, false);
}

void TITView::registerRecords()
{
    records.add (map::EditorMode::history, [this] (const int&) { return history.getSource(); });
}

void TITView::registerSelections()
{
    selections.add (map::EditorMode::history, [this] (const int&, const juce::ValueTree& record)
    {
        if (record.hasType (jam::Format::toType (Id::file)))
            state.setProperty (Id::focusedPane, static_cast<int> (map::Pane::text), nullptr);
    });
}

juce::PopupMenu TITView::getMenu (const juce::ValueTree& git) const
{
    static const jam::HashMap<juce::CommandID, juce::Identifier> loadingKeys {
        { map::CommandId::history, Id::historyLoading },
        { map::CommandId::fileHistory, Id::fileHistoryLoading },
    };

    const auto isNormal { static_cast<int> (git.getProperty (Id::operation)) == map::Operation::normal };
    juce::PopupMenu menu;

    for (const auto command : getMenuCommands())
    {
        const auto isActive { isAvailable (command, git) };

        if (isActive or (isNormal and loadingKeys.contains (command)))
        {
            if (command == map::CommandId::history and menu.getNumItems() > 0)
                menu.addSeparator();

            menu.addItem (getMenuItem (isActive ? getMenuKey (command, git) : loadingKeys.at (command), command, isActive));
        }
    }

    return menu;
}

juce::PopupMenu::Item TITView::getMenuItem (const juce::Identifier& key, juce::CommandID command, bool isActive) const
{
    static const jam::HashMap<juce::CommandID, const char*> progressTexts {
        { map::CommandId::history, text::Reading::historyProgress },
        { map::CommandId::fileHistory, text::Reading::fileHistoryProgress },
    };

    const auto row { jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcMenu })->getValueTree (jam::Format::toType (Id::config)).getChildWithName (jam::Format::toType (Id::menu)).getChildWithName (jam::Format::toType (key)) };
    const auto numCommits { static_cast<int> (state.getChildWithName (jam::Format::toType (Id::history)).getProperty (Id::numCommits)) };
    const auto hasProgress { not isActive and numCommits > 0 and progressTexts.contains (command) };
    const auto label { hasProgress ? juce::String { progressTexts.at (command) }.replaceFirstOccurrenceOf (text::Reading::hashPlaceholder, juce::String { numCommits })
                                                                             .replaceFirstOccurrenceOf (text::Reading::hashPlaceholder, text::Git::historyCount)
                                   : row.getProperty (Id::value).toString() };
    juce::PopupMenu::Item item { row.getProperty (Id::emoji).toString() + juce::String::charToString (Chars::space) + label };

    item.itemID = command;
    item.commandManager = commandManager.get();
    item.shortcutKeyDescription = row.getProperty (Id::shortcut).toString();
    item.isEnabled = isActive;

    return item;
}

int TITView::getInitialItemId (const juce::PopupMenu& menu)
{
    int firstItemThatRuns { 0 };

    for (juce::PopupMenu::MenuItemIterator iterator (menu); iterator.next();)
    {
        const auto& item { iterator.getItem() };

        if (item.isTicked) return item.itemID;

        if (firstItemThatRuns == 0 and jam::PopupMenu::canBeTriggered (item))
            firstItemThatRuns = item.itemID;
    }

    return firstItemThatRuns;
}
```

   - The menu is the stock popup menu of JAM (Part 10 Step 12). A separator is `juce::PopupMenu::addSeparator`. An item that cannot run has `isEnabled` false: the menu does not highlight it and does not run it (`jam::PopupMenu::canBeTriggered`, `jam_gui/menu/jam_PopupMenu.h:99-105,1680-1714`). The LookAndFeel paints both. So TIT has no separator row, no `isActive` row, no menu index, and no move or enter command.
   - `getMenu` returns the menu, `getMenuItem` returns one item, and `setMenu` shows the menu (NAMES.md Rule 1: "`get` returns"). Declare in the header after `getMenuCommands`: `juce::PopupMenu getMenu (const juce::ValueTree& git) const;`, `juce::PopupMenu::Item getMenuItem (const juce::Identifier& key, juce::CommandID command, bool isActive) const;`, and `static int getInitialItemId (const juce::PopupMenu& menu);`. Part 13 gives the item its value text.
   - `getInitialItemId` gives the highlighted item of a new menu: the item with a tick, or else the first item that can run. It uses the iterator of JUCE that `MenuWindow` uses (`jam_PopupMenu.h:734-739`), which reads the top level only.
   - The menu of the records is the same stock menu, built from a tree of records by one function, `getRecordMenu`. A record is an item: its text is its `attribute`, a space, and its `value`. Its `itemID` is its row index plus `firstItemId`, because an `itemID` of 0 is no item (`jam_PopupMenu.h:99-105`). A record whose children are records gets them as its submenu, through the same function. A record has the row `id` (Step 6, the record shape), and `getChild` gives an invalid tree for a record with no child (doxygen `classjuce_1_1ValueTree`). So a commit shows its files on the right key, and every later tree of records shows with no new code (Part 12).
   - The selected record has a tick: `isTicked` (inventory). The LookAndFeel paints it.
   - `setSelection` writes the `selectedIndex` row of each tree from the record up to the tree under the root. An equal value gives no event (JUCE `ValueTree::setProperty`).
   - Return on an item runs its `action` (inventory). The action selects its record, and then runs the entry of the mode in `selections`: what Return does on a record of that screen. In the browser, Return on a file gives the keys to the text pane. Return on a commit selects it, and the text pane shows its head text and the patch of its selected file. Part 12 adds the legacy Enter of each of its screens to this table.
   - The action holds its record as a `juce::ValueTree`, a counted reference. A feed that arrives after the item ran takes the record out of its tree (Step 7, `setChildren`); `isAChildOf` is then false, and the action does nothing.
   - `keyPressed` replaces the body of Part 9 Step 7, item 2. A key that the menu did not take arrives there (Part 10 Step 16, item 2). When the records have the keys, the highlighted record becomes the selected record first, and then the key sheet invokes the command. So a command with a key acts on the highlighted row and reads only Model rows (Closed by a read, item 1).
   - `getHighlightedRecord` is the one place of TIT that reads the highlight of the menu. It walks the open menu windows in the order of the list: the top window, then its submenu window. Each window gives its highlighted item, `currentChild` (`jam_PopupMenu.h:1794`), and the `itemID` of the item gives the row. A window that hid has no highlighted item (`jam_PopupMenu.h:867-868`). With no highlighted item, the result is the tree itself, and `setSelection` writes nothing.
   - `records` gives the tree of records of a mode, and `isRecordMode()` is true when the menu of that tree has the keys: the mode has an entry and the focused pane is `records`. Part 12 adds its modes to the two tables.
   - The dismiss loop counts up. A top window that hides resets its submenu window, which leaves the list (inventory). A submenu window is made after its top window, so it is behind it in the list, and the loop reads no removed entry.
   - The callback gets the `itemID` of the run item, or 0 when the menu closed with no item (`jam_PopupMenu.h:877-879`). Escape closes the menu with no item (`:1003-1006`), and `setMenu` itself does so before it shows a new menu. After Escape no menu is open and the records still have the keys, so the callback invokes `cancel` (item 8). After a close by `setMenu`, a new menu is open, or the mode or the pane has changed, and the callback shows what the state gives.
   - The order of `operationCommands` is the legacy order with the two rows of ruling 22 at the end of their groups (`legacy/app/menu_render_core.go:35-59`; Part 2, "Menu rows").
   - An item of a cache that loads shows its `loading` row and cannot run. Its label is `History...` or `Files...` until the first record arrives, and then the count: `History 12/100`, `Files 12/100` (Closed by a read, item 6). The total is the count that the cache command asks for.

6. `registerAvailability()`: add these entries in the form of Part 10. An entry that reads a root row captures `this`.

   | Command | Rule |
   |---|---|
   | `history`, `fileHistory` | `operation` is `normal`, and the root row `isCacheReady` is `true` |
   | `viewChanges` | `operation` is `normal`, and `workingTree` is `dirty` |
   | `incomingOutgoing` | `operation` is `normal`, `remote` is `hasRemote`, and `timeline` is `ahead`, `behind`, or `diverged` |

   - The first row reads the root row `isCacheReady`. Its event builds the sheet (item 4).
   - `copyHash`, `copyFullHash`, `rewind`, `functionHistory`, `historyTab`, `searchHistory`, and `searchSubmit` have no entry. The mode of the key row is the gate of each one.
   - Ruling 40: "user can interact according to the data". A command of the browser acts on the selected record, and the record is its data. `copyHash`, `copyFullHash`, and `rewind` act on the hash of the selected commit. `functionHistory` acts on the path of the selected file. The record of the working tree changes has no hash, so the three hash commands do nothing on it (item 7). This is the form of an END callback, which acts on the view that exists (`end/Source/ui/ENDView.cpp:194-199`).

7. The three families of the browser. Each entry tells the processor, tells `HistoryView`, or writes a Model row (Part 9, Overview, item 6):

```cpp
void TITView::registerHistoryActions()
{
    commands.add (map::CommandId::history, [this] (const juce::CommandID&) { setBrowser (Id::history); });
    commands.add (map::CommandId::fileHistory, [this] (const juce::CommandID&) { setBrowser (Id::history); });

    commands.add (map::CommandId::viewChanges, [this] (const juce::CommandID&)
    {
        gitProcessor.readChanges();
        setBrowser (Id::reading);
    });

    commands.add (map::CommandId::incomingOutgoing, [this] (const juce::CommandID&)
    {
        gitProcessor.readIncomingOutgoing (TITGitProcessor::getUpstream (getGit()));
        setBrowser (Id::reading);
    });
}

void TITView::registerBrowserActions()
{
    commands.add (map::CommandId::historyTab, [this] (const juce::CommandID&) { state.setProperty (Id::focusedPane, (static_cast<int> (state.getProperty (Id::focusedPane)) + 1) % HistoryView::numPanes, nullptr); });

    commands.add (map::CommandId::copyHash, [this] (const juce::CommandID&) { copyText (history.getCommit().getProperty (Id::id).toString().substring (0, shortHashLength)); });
    commands.add (map::CommandId::copyFullHash, [this] (const juce::CommandID&) { copyText (history.getCommit().getProperty (Id::id).toString()); });

    commands.add (map::CommandId::rewind, [this] (const juce::CommandID&)
    {
        const auto hash { history.getCommit().getProperty (Id::id).toString() };

        if (hash.isNotEmpty())
            setConfirmation (text::Confirmation::rewindTitle,
                             juce::String { text::Confirmation::rewindExplanation }.replace (text::Reading::hashPlaceholder, hash.substring (0, shortHashLength)),
                             text::Confirmation::rewindYes, text::Confirmation::rewindNo, TITGitProcessor::getRewindLines (getGit(), hash), {});
    });
}

void TITView::registerSearchActions()
{
    commands.add (map::CommandId::searchHistory, [this] (const juce::CommandID&)
    {
        setPrompt (text::Git::historyLog + juce::String::charToString (Chars::space) + text::Git::historyCount, text::Git::searchOption, {}, {}, text::Footer::search);
    });

    commands.add (map::CommandId::functionHistory, [this] (const juce::CommandID&)
    {
        const auto colon { juce::String::charToString (Chars::colon) };
        const auto path { history.getFile().getProperty (Id::id).toString() };

        if (path.isNotEmpty())
            setPrompt (text::Git::functionLog, colon, colon + path, path, text::Footer::functionHistory);
    });

    commands.add (map::CommandId::searchSubmit, [this] (const juce::CommandID&)
    {
        const auto reading { state.getChildWithName (jam::Format::toType (Id::reading)) };
        const auto typedText { history.getSearchText() };

        if (typedText.isNotEmpty())
        {
            gitProcessor.read (reading.getProperty (Id::commandLines).toString(),
                               reading.getProperty (Id::argumentPrefix).toString() + typedText + reading.getProperty (Id::argumentSuffix).toString(),
                               reading.getProperty (Id::path).toString());
            state.setProperty (Id::source, jam::Format::toType (Id::reading).toString(), nullptr);
        }

        state.setProperty (Id::mode, static_cast<int> (map::EditorMode::history), nullptr);
    });
}

void TITView::setBrowser (const juce::Identifier& source)
{
    state.setProperty (Id::source, jam::Format::toType (source).toString(), nullptr);
    state.setProperty (Id::focusedPane, static_cast<int> (map::Pane::records), nullptr);
    state.setProperty (Id::mode, static_cast<int> (map::EditorMode::history), nullptr);
}

void TITView::setPrompt (const juce::String& commandLine, const juce::String& argumentPrefix, const juce::String& argumentSuffix, const juce::String& path, const char* message)
{
    auto reading { state.getChildWithName (jam::Format::toType (Id::reading)) };

    reading.setProperty (Id::commandLines, commandLine, nullptr);
    reading.setProperty (Id::argumentPrefix, argumentPrefix, nullptr);
    reading.setProperty (Id::argumentSuffix, argumentSuffix, nullptr);
    reading.setProperty (Id::path, path, nullptr);
    history.setSearchDocument();
    state.setProperty (Id::focusedPane, static_cast<int> (map::Pane::records), nullptr);
    state.setProperty (Id::mode, static_cast<int> (map::EditorMode::search), nullptr);
    state.setProperty (Id::message, message, nullptr);
}

void TITView::copyText (const juce::String& text)
{
    if (text.isNotEmpty())
    {
        juce::SystemClipboard::copyTextToClipboard (text);
        state.setProperty (Id::message, text::Reading::copySuccess, nullptr);
    }
}
```

   - The two legacy menu items open the one browser on the cache (Closed by a read, item 19).
   - A command of the browser acts on the selected record. `keyPressed` made it the highlighted record before the command ran (item 5; Closed by a read, item 1). `history.getCommit()` and `history.getFile()` read it from the `selectedIndex` rows. No command reads the highlight of the menu.
   - `historyTab` goes to the next pane: from the menu of the records to the text pane, and back. The text pane then shows the record that was highlighted.
   - Every other command sets the feed and nothing else: it tells the processor what to read, and it names the tree that the browser shows.
   - The two prompts are one path. The command writes the request into the `reading` rows and shows the prompt text in the footer. `searchSubmit` joins the typed text into the argument and starts the request. The search gives `-S` and the text; the function history gives `:`, the function name, `:`, and the path (Step 2, item 4).
   - Rewind has the legacy lines (inventory). The lines come from `TITGitProcessor::getRewindLines` (Step 7), and the upstream name from `TITGitProcessor::getUpstream` (Step 7).
   - The commit box and the prompt are each an `InputView` with its own `setDocument()` (Part 10 Step 14). `TITView` has no document function.

8. The cancel. Replace the `cancel` entry of `registerViewActions` of Part 10 Step 16, item 7:

```cpp
void TITView::registerCancellations()
{
    cancellations.add (map::EditorMode::input, [this] (const int&) { cancelInput(); });

    cancellations.add (map::EditorMode::console, [this] (const int&)
    {
        if (static_cast<bool> (state.getProperty (Id::isRunning)))
            gitProcessor.cancel();
        else
            setMenuMode (getGit());
    });

    cancellations.add (map::EditorMode::search, [this] (const int&) { state.setProperty (Id::mode, static_cast<int> (map::EditorMode::history), nullptr); });

    cancellations.add (map::EditorMode::history, [this] (const int&)
    {
        if (isRecordMode())
            setMenuMode (getGit());
        else
            state.setProperty (Id::focusedPane, static_cast<int> (map::Pane::records), nullptr);
    });
}
```

   ```cpp
    commands.add (map::CommandId::cancel, [this] (const juce::CommandID&)
    {
        const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

        if (cancellations.contains (mode))
            cancellations.get (mode);
        else
            setMenuMode (getGit());
    });
   ```

   Escape in the menu of the records returns to the main menu, as in legacy (`legacy/app/app_keys.go:78-89`): the menu closes, and its callback invokes `cancel` (item 5). Escape in the text pane returns to the menu of the records. Escape in the prompt returns to the browser.

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- `commands` holds 31 ids. `availability` holds the 15 ids that a state rule governs, and the 13 edit command ids. No callback and no line of `perform` tests if a command is allowed.
- A command of the browser reads only the selected record, which is its data. A record with no hash gives no copy and no rewind dialog, and a record with no file gives no function prompt.
- `TITView` has no branch on a feed, holds no menu and no editor, and makes no document and no command line. One function, `getHighlightedRecord`, reads the highlight of the menu, and `keyPressed` writes it to the Model rows. Each entry of the three families tells the processor, tells `HistoryView`, or writes a Model row.
- `TITView` reads no state of a child component, except the selected record and the prompt text through the public functions of `HistoryView`.
- `TITView` holds no patch and makes no document. It gives the `patch` line of the selected record to `readPatch`.
- Data flows one way. A feed: key, sheet, command, processor request, pool job, lane, count row, `applyRecords`, Model tree, listener, menu. A patch: selected record, `readPatch`, TTY, decoder, swap, editor.
- `shortHashLength` is the legacy `CopyHashShortHashLen` (`legacy/ui/history.go:33`).
- New names and their families: `registerHistoryActions`, `registerBrowserActions`, `registerSearchActions` (the `register…Actions()` family, `end/Source/ui/ENDView.h:121-179`; ruling 40, "BROWSER"; the mode `search`); `registerCancellations`, `cancellations` (NAMES.md Rule 8); `setBrowser` (ruling 40; `setMenuMode`); `setPrompt` (Part 2, "Search history prompt"); `copyText` (`juce::SystemClipboard::copyTextToClipboard`); `getFocusComponent` (`getViewedComponent`, `jam_TextEditor.h:204`); `getMenu`, `getMenuItem` (`getMenuKey`, Part 10; the row `menu`); `getRecordMenu`, `records`, `registerRecords`, `isRecordMode`, `getHighlightedRecord` (ruling 40, "SSOT of RECORDS"; `getMenu`; `isMenuMode`; `setCurrentlyHighlightedChild`, `jam_PopupMenu.h:1618`; NAMES.md Rule 8); `setSelection`, `selections`, `registerSelections` (`Id::selectedIndex`; `cancellations`); `setPatch` (`readPatch`, Step 7; `setMenu`); `getInitialItemId` (`getInitiallySelectedItemId`, `jam_PopupMenu.h:741`); `firstItemId` (`itemID`, JUCE); `history` (the view of Step 10). Local names: `mode`, `components`, `row`, `record`, `tree`, `child`, `window`, `highlighted`, `isNormal`, `item`, `isActive`, `index`, `result`, `firstItemThatRuns`, `hash`, `colon`, `path`, `reading`, `typedText`.

### Step 12: `FooterView`, the hints of the browser
**Scope:** `tit/Source/FooterView.cpp`.
**Action:**
1. `valueTreePropertyChanged`: add `Id::focusedPane` to `paintedProperties`.
2. Replace `getHint()` of Part 10 Step 13:

```cpp
juce::String FooterView::getHint() const
{
    static const jam::HashMap<int, const char*> hints {
        { map::EditorMode::confirmation, text::Footer::confirmation },
        { map::EditorMode::input, text::Footer::input },
    };

    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    if (mode == map::EditorMode::console) return static_cast<bool> (state.getProperty (Id::isRunning)) ? text::Footer::consoleRunning : text::Footer::consoleComplete;
    if (mode == map::EditorMode::history) return static_cast<int> (state.getProperty (Id::focusedPane)) == map::Pane::text ? text::Footer::readingText : text::Footer::history;
    if (hints.contains (mode)) return hints.at (mode);
    return {};
}
```

**Validation:**
- `getHint()` has three branches. Each return is a result return.
- Each hint equals its legacy row (`legacy/app/messages_menu.go:70-109`): `history_list` when the menu of the records has the keys, `filehistory_diff` when the text pane has them. A menu mode has no hint (Part 10 Step 13).
- In the mode `search` the footer shows the prompt text: it is the `message` row, which `getText()` of Part 10 shows before a hint.

### Step 13: Build, run, and measure on macOS
**Scope:** JAM and the `tit/` project; the fixture repositories of Part 8 Step 9.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build). The fixture commands are git commands: ARCHITECT runs them, or gives the git instruction (CAROL.md Git).
**Action:**
1. Build `ling` and EVE as in Part 10 Step 18, items 1 and 2. Step 1 changed one cast table of JAM.
2. From `tit/`, run `cast cast/spell.md --debug`.
3. Start `titc` in END in each folder below, and do the keys.

   | Folder | Keys | Expected |
   |---|---|---|
   | fixture 4 with the file committed and a second commit that changes two files | none | The menu shows a separator, then `History` and `File(s) history`. Within one second of the start, no row shows `History...`. |
   | the same | cursor down to the last row and past it | The selection skips the separator and wraps. |
   | the same | `h` | The menu of the records over the text pane. Each item shows the time, the short hash, and the subject. The first item has a tick. The text pane shows `Author: `, the author, `Date:   `, the message, then the patch of the first file with a cyan hunk header and green added words. |
   | the same | `j`, Tab, Tab, the right key, `j`, Return, `j`, `v`, `j`, `y`, Escape | The highlight moves. Tab closes the menu, and the text pane shows the head text and the first file of the second commit. Tab shows the menu again, with the tick and the highlight on the second commit. The right key shows its files with their counts. Return on the second file closes the menu; the text pane shows the patch of that file and takes the keys. The text scrolls; the visual mode marks lines; `y` copies them. Escape shows the menu again. The footer hint follows the pane. |
   | the same, from the menu | `f` | The same browser. |
   | the same | `h`, `y`, then `Y` | The footer shows `✓ Copied to clipboard`. The clipboard holds 7 characters, then 40. |
   | the same | `h`, `/`, a word of the first commit, Escape, Return | The footer shows `Search history:`. The browser then shows only the commits that added or removed the word. Escape gives the menu. |
   | the same | `h`, `/`, a text that no commit holds, Escape, Return | The footer shows `No commit matches`, and the browser shows the cache. |
   | the same, a C file with a function that two commits changed | `h`, the right key, `j` to the C file, `F`, the function name, Escape, Return | The footer shows `Function history:`. The browser shows the commits that changed the function, each with one file and the patch of the function. |
   | the same | `h`, `j`, `ctrl+r`, Return | The dialog `DESTRUCTIVE OPERATION` shows with the short hash of the second commit in its text. Return runs `REWIND`. The console shows the reset line. The history then has one commit less. |
   | fixture 4, one new file and one changed file | `v` | One record, `Local changes present`. Its files: the changed file with `+1 -0`, the new file with `+`. The detail shows the patch, and the content of the new file. `y` and `ctrl+r` do nothing. |
   | fixture 6, `Diverged` | `o` | One record with `⬇` and one record with `⬆`. Each shows its files and their patches. |
   | `jam/` | none, then `h`, then hold `j`, then Tab | The label `History...` shows for less than two seconds. The highlight moves with no visible wait. Tab shows the detail of the commit with no visible wait. |

4. Measure with the temporary log of Part 8 Step 9, item 1: write the time at the start of the cache job and at the write of `isCacheReady`, in `jam/` and in fixture 4. Remove the log lines after the run.
5. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- Each build ends with exit code 0. Each row gives its expected result.
- These facts of this part are measured in the program: the records of Step 7 equal the output of the cache command (the session measured the output alone); a path with a space as one argument of `jam::ChildProcess`, and as one quoted argument of a patch line; the colours of the detail; the patch line of a function record, `functionPatch`, prints one commit; the patch line of an untracked file, `untrackedPatch`, prints the file as added; the cache of `jam/` fits the lane of the records with no assert.
- The time from a new selection to the patch on the screen is below one frame of wait for the user in `jam/` (ruling 37: the detail is instant).
- The time of the cache job in `jam/` is below two seconds (measured in this session: 0.94 s for the process alone).
- ARCHITECT sets the layout of the browser at this run (ruling 7): the place of the menu of the records, and the form of the diff.
- The second `cast` run changes no generated file. After item 4, no log line remains.

### Step 14: Run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 13, items 2, 3, and 5, in END on a Windows machine with fixture 4 and with `jam\`.
**Validation:**
- The results of Step 13 hold.
- A path of a file row has forward slashes, as the tool writes it, and `juce::File::getChildFile` reads it.
- The clipboard holds the copied hash.

## BLESSED Alignment

- **B — Bound.** `TITView` owns `HistoryView`, and `HistoryView` owns its text pane and its search input. JAM owns the window of a menu until it closes (Part 10 Step 15). The processor owns the pool, the lane of the records, and the run chain `patch`. No tree crosses to the pool thread: the pool thread writes the lane, and the message thread builds the tree.
- **L — Lean.** One browser for every feed. One function builds the menu of every tree of records. One text pane for the detail. One record shape. No parser of a diff. No mode for a feed.
- **E — Explicit.** Each command line, format, separator, text, key, colour, and geometry number is a named row or a named constant. A request that fails shows the text of the tool.
- **S — SSOT.** The records are one shape, and the cache is one tree. The row `source` names the tree that the browser shows; no list holds a copy. `cast/git.md` is the one place of each command line, and `historyCount` is the one count.
- **S — Stateless.** The selected record is a Model row. The highlight and the scroll offset are rendering state of the stock menu, and no code of TIT holds them (ruling 61). One function reads the highlight, at a key of the user. The text pane reads the swap of its chain. The processor holds no result: the pending records wait only until their feed is complete.
- **E — Encapsulation.** Each record is complete at creation on the pool thread, and each feed enters the Model complete, in one step. The views read it in place. A patch goes from the tool to the swap and no object changes it. `TITView` tells the processor what to read and reads no answer from it.
- **D — Deterministic.** The same repository gives the same records, and the same records and selections give the same menu and the same detail.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. The items of "Closed by a read" are ORACLE's reads; ARCHITECT can reverse each one. Rulings 37 to 40 are in the inventory. The runtime facts of this part are in Steps 13 and 14.
