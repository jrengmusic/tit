# PLAN: TIT Part 10 — The Shell Views

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part builds the shell of TIT: the views that every later part puts its content into.

1. **Two views in the form of `jam::TextEditorStatusBar`. The menu and the dialog are the stock components of JAM and JUCE.** `HeaderView` and `FooterView` are `juce::Component`s that paint the Model rows with `fillRect` and `drawText`. A menu is the popup menu of JAM, `jam::PopupMenu::MenuWindow`, shown inside the view and painted by the LookAndFeel (rulings 60, 61, and 63; Step 12). A dialog is the `juce::AlertWindow` of JUCE inside the view, shown by `jam::TerminalMessageBox` in the contract of `jam::MessageBox` (rulings 65 and 66; Step 14). The Model holds no row for the menu or for the dialog.
2. **Two editors.** The console is a read-only `jam::TextEditor` that shows the output of a run. The commit box is the first editable `jam::TextEditor` of TIT (rulings 5 and 32).
3. **`TITView` composes them.** The `mode` row of Part 9 selects the view that shows. Three new modes join the menu modes: `console`, `confirmation`, `input`.
4. **The commands of these views.** The menu keys, the dialog keys, cancel, submit, and six git commands that need a dialog, a text, or the console: `commit`, `commitPush`, `resetDiscardChanges`, `forcePush`, `replaceLocal`, `pullMerge`.
5. **Six changes of JAM** that the views need: a program run on `jam::TerminalTTY` with its exit status (ruling 59), a key description with a space in the key sheet, the default mode of an editable editor, the rows of the new modes and colour ids, the keys `j` and `k` of `jam::PopupMenu::MenuWindow`, and `jam::TerminalMessageBox`.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 5, 6, 7, 8, and 9. Steps 1 to 4, Step 12, and the first four items of Step 14 change JAM. Step 1 also changes one class of EVE.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`. A JAM submodule file has no include.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization, and parentheses for an initializer of the same container type; no `==` on strings; no out-parameter; no fake-carrier struct; no local struct; no anonymous namespace; no non-owning naked pointer as a member.
- CODING.md, "Uniform signature in the type" and "Consume the event payload".
- MANIFESTO L: no `if` chain longer than three branches; a longer decision is a table. No function body longer than 30 lines.
- MANIFESTO S (Stateless): "State belongs exclusively to the Model. View and machinery are pure." The selection of the dialog, the footer message, and the two confirm flags are Model rows. The highlight and the scroll offset of a menu are rendering state of `jam::PopupMenu::MenuWindow`, not rows (ruling 61: *"dont conflate rendering state with state machine"*). The output of a run is a swap file (ruling 58).
- MANIFESTO L, YAGNI, and Part 1 Step 7: this part writes only the rows, ids, modes, colours, and texts that its own code reads.
- Ruling 7: a layout locks at execution. Each geometry number of this part is a named constant with its legacy source. ARCHITECT changes a layout at the run of Step 18 with no change of this plan's contracts.
- ARCHETYPE.md §1: no step edits a generated file by hand. Each cast table keeps its `+---` separator lines.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build and Git: agents build only on ARCHITECT's explicit request. The fixture repositories need git commands; ARCHITECT runs them, or gives the git instruction.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`.

### What JAM has and does not have

| Need | Fact | Where |
|---|---|---|
| A list with a shortcut, a hint, and an inactive row | None. `jam::Selector` items hold `id` and `text` only. | `jam_gui/widgets/jam_Selector.h:25-29` |
| A confirmation in the terminal | None. `jam::MessageBox::showYesNoBox` is `juce::NativeMessageBox::showAsync`. `ModalWindow` and `Popup` are native windows. END, EVE, and LING hold no confirmation component. JUCE has it: a `juce::AlertWindow` with `MessageBoxOptions::withParentComponent` is a child component (Step 14). | `jam_gui/windows/jam_MessageBox.cpp:27-40`; `jam_gui/windows/jam_ModalWindow.h:35` |
| A label component | None. `jam::TextBlock` has no base. | `jam_gui/widgets/jam_TextBlock.h:11` |
| A timed message | `jam::MessageOverlay` fills with `withAlpha (0.8)`, pads by 20 units, and draws with `drawFittedText`. The terminal store drops alpha. So its paint does not fit cells. | `jam_gui/widgets/jam_MessageOverlay.cpp:12,59,77-109`; `jam_graphics/colours/jam_Colours.h:44` |
| Stock JUCE widgets | `StyleCustom` overrides `getDefaultScrollbarWidth` and `drawScrollbar` only. It overrides no `drawLabel`, `drawButtonBackground`, `drawButtonText`, or `getLabelFont`. A path or a rounded rectangle becomes half-coverage cells. | `jam_style/jam_StyleCustom.h:261,268-269` (full header read); `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:428-462` |

### The view form that this part follows: `jam::TextEditorStatusBar`

| Piece | Where |
|---|---|
| `class TextEditorStatusBar : public juce::Component, public jam::Model::Component<TextEditorStatusBar>, private juce::ValueTree::Listener` | `jam_gui/widgets/jam_TextEditorStatusBar.h:20-23` |
| The colour ids are an enum of `map::ColourId` rows: `backgroundColourId = map::ColourId::textEditorStatusBarBackgroundColourId,` | `jam_TextEditorStatusBar.h:27-36` |
| The constructor: `setOpaque (false); setInterceptsMouseClicks (false, false); state.addListener (this);` | `jam_TextEditorStatusBar.cpp:5-11` |
| `paint`: `jassert (dynamic_cast<const StyleCustom*> (&getLookAndFeel()) != nullptr);`, then the state rows, `g.fillAll (findColour (backgroundColourId));`, `g.setColour (findColour (…)); g.fillRect (area);`, `g.setFont (juce::Font { style.getMonoBoldFont() }); g.drawText (label, labelArea, juce::Justification::centred);` | `jam_TextEditorStatusBar.cpp:13-62` |
| A size in cells: `const auto cellWidth { style.getMonoMetrics().cellWidth };` and `juce::roundToInt (cellWidth * static_cast<float> (…))` | `jam_TextEditorStatusBar.cpp:40-41` |
| The listener repaints on the rows that `paint` reads: `static const jam::Array<juce::Identifier> paintedProperties { … }; if (paintedProperties.contains (property)) repaint();` | `jam_TextEditorStatusBar.cpp:64-70` |
| A table inside `paint`: `static const jam::HashMap<int, int> modeColours { { map::EditorMode::normal, normalColourId }, … };` and a text table with literals, `{ map::VisualMode::character, "VISUAL" }` | `jam_TextEditorStatusBar.cpp:19-31` |

### The terminal paint facts

| Fact | Where |
|---|---|
| `fillRect` quantizes to cells and writes the background colour id of each cell. It never replaces a character. | `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:102-117,276-300` |
| `drawSingleLineText`, `drawMultiLineText`, `drawText`, `drawFittedText`, and `AttributedString::draw` reach `drawAttributedText`, which writes cells | `jam/patch/juce-attributed-text-hook.patch` (76 lines, read whole) |
| The font size is not read for a glyph. Bold, italic, and underline are. | `jam_TerminalGraphicsContext.cpp:582-590` |
| The text colour is the fill colour of the context | `jam_TerminalGraphicsContext.cpp:739` |
| The columns of a cluster: `AttributedChar::getColumns`. A cluster with U+FE0F or U+20E3 has two columns. Any other cluster takes the width of its first codepoint from the Unicode 17.0.0 table. | `jam_core/text/jam_AttributedChar.cpp:2641-2647,3303,3319-3337` |
| A colour id reaches the sheet in three steps: a row of `## ColourId` (`textEditorStatusBarNormalColourId`, `0x4700005`, `` `text editor status bar normal colour id` ``), the enum of the component, and the custom `--TextEditorStatusBar--normalColourId: var(--paradiso);` in the `.appearance` block | `jam/cast/style.md:77-92,262`; `jam_TextEditorStatusBar.h:31`; `eve/Source/layout/style.css:68`; `ling/Source/style.css:65-72,99-101` |
| `StyleCustom` sets each `map::ColourId` entry to magenta until a sheet sets it | `jam_style/jam_StyleCustom.cpp:5-12` |
| The palette names are customs of `:root`: `--white: #FFFFFF;` | `ling/Source/style.css:17-62` |
| END adds no colour id of its own: `end/cast/` has no `style.md` | `ls end/cast` |
| The last used key of the JAM component colour ids is `0x470000a`; Part 6 adds `0x470000b` | `jam/cast/style.md:260-268`; Part 6 Step 1 |

### The panel pattern of EVE

| Piece | Where |
|---|---|
| A component over the view: `addAndMakeVisible (markdown); markdown->setBounds (getLocalBounds()); … markdown->toFront (shouldGrabFocus);`, then the state takes the mode `panel` | `eve/Source/EVEView.cpp:610-617` |
| The return: the `mode` property goes, and `if (hadFocus) grabKeyboardFocus();` | `EVEView.cpp:204-223` |
| The view gives a key to the sheet with the state that holds the mode | `EVEView.cpp:89-94` |

### The console facts

| Fact | Where |
|---|---|
| `jam::TerminalTTY` gives no exit code today: `read` returns "-1 on EOF / error", and `close()` returns `void`. Part 9's chain needs the exit code. Step 1 makes `close()` return it (ruling 59). | `jam_terminal/tty/jam_TerminalTTY.h:104,116-125`; Part 9 Step 5 |
| The decoder reads only a `jam::TerminalTTY`, and it fires `shellExited` with 1 when the read gives -1 | `jam_terminal/processor/jam_TerminalProcessor.cpp:66,78-88,122-133` |
| The chain of EVE: the tick decodes the TTY into the buffer and the swap; the async update moves the history rows into the swap; the editor getter gives the swap | `eve/Source/EVETerminalProcessor.cpp:256-273,329-345,554-564` |
| `TextEditor::bindToProcessor (getters, type, parameterID)` sets the source of the editor to the getter and pulls at each vblank | `jam_gui/widgets/jam_TextEditor.h:134-144`; `jam_TextEditor.cpp:82-100` |
| `jam::Function::Map::get` throws `std::out_of_range` for an absent key | `jam_core/function_map/jam_Function.h:311` |
| The editor takes a key by its mode. In `terminal` mode: the key sheet first, then the bytes of the key go to the input callback. In `reader` mode: the commands, then the viewport, then the parent. | `jam_gui/widgets/jam_TextEditor.cpp:604-645` |
| `addBlocks()` pulls the swap at each vblank with no test of the mode or of read-only. The view follows the end only in `terminal` mode. | `jam_TextEditor.cpp:1056-1091` |
| `jam::CpuGraphicsEngine::getRefreshRateHz()` is in `jam_graphics`, a module of TIT | `jam_graphics/engine/jam_CpuGraphicsEngine.h:71`; `PLAN-tit-1-project.md:171` |
| Legacy reads byte by byte and flushes a line at each `\r` and each `\n`; a `\r` line replaces the last line | `legacy/git/execute.go:143-186`; Part 2, "What the shipped code has", row "Progress lines" |
| `AnsiDocument::addLines (data, size)`: "Each newline in the bytes adds an `Id::line` element. The text after the last newline stays in the pen". It applies the SGR colours. Its control branch tests the newline only, so a carriage return adds nothing. | `jam_terminal/document/jam_AnsiDocument.h:43-53`; `jam_AnsiDocument.cpp:5-9,638-644,967-1044` |
| `TextEditor::setLiveBlock (std::unique_ptr<Document>)` replaces the live block. It sets the cursor and the anchor to 0, the default mode and the mode to `reader`, read-only to `true`, and the view position to the top. It is public after Part 6. | `jam_gui/widgets/jam_TextEditor.cpp:1130-1155`; Part 6 Step 6 (ruling 32) |
| The terminal path of the editor builds a new `AnsiDocument` for each live frame | `jam_TextEditor.cpp:1117-1128` |
| The view goes to the end with the command `scrollDocumentEnd` | `jam_TextEditor.cpp:557-559,1093-1096`; `jam/cast/terminal.md:1908` |
| The reader mode gives a key to the sheet, then to its viewport, then to the parent | `jam_TextEditor.cpp:629-632`; JUCE `juce_gui_basics/windows/juce_ComponentPeer.cpp:196-221` |
| `jam::TextEditor` has no hook for a text change or a Return key | `jam_gui/widgets/jam_TextEditor.h` (public section read) |

### The key sheet facts

| Fact | Where |
|---|---|
| `setKeyPresses` splits a row value with `juce::StringArray::fromTokens (…, false)`. A token longer than `sequenceLength` is a `juce::KeyPress::createFromDescription`. | `jam_gui/commands/jam_CommandManager.cpp:16,23-24` |
| The JUCE descriptions of the arrow and page keys hold a space: `cursor left`, `cursor right`, `cursor up`, `cursor down`, `page up`, `page down`. `spacebar`, `return`, `escape`, `tab`, `backspace`, `delete`, `home`, `end` hold none. | JUCE `juce_gui_basics/keyboard/juce_KeyPress.cpp:84-105` |
| `juce::StringArray::fromTokens (StringRef, bool preserveQuotedStrings)`; `juce::String::unquoted()` "Removes quotation marks from around the string" | doxygen `classjuce_1_1StringArray`, `classjuce_1_1String` |
| A `string` cell of a config table "passes the text through unchanged" | `jam_markdown/document/jam_ConfigDocument.h:91` |
| A config table with a column outside the config columns gives one child for each row, named by its key, with a `value` property and a text property for each other column | `jam_ConfigDocument.h:150-153` |
| No table of JAM, END, or EVE binds an arrow key | `jam/cast/terminal.md:1913-2066`; `end/Source/config/keys.md`; `eve/Source/layout/default-eve.md` |
| The JAM default row `defaultMode` has the keys `escape i a I A o O` in the modes `normal visual open` | `jam/cast/terminal.md:1897,1963,2099` |
| The sheet of TIT is the JAM default rows that `Source/keys.md` does not name, then the rows of `Source/keys.md`. `getKey` returns the first row that matches. | Part 7 Step 7, item 3; `jam_CommandManager.cpp:96-109` |
| The mode of a row is tested against the state that the caller passes. An editor passes its own `terminal` tree. `TITView` passes the Model root. | `jam_CommandManager.cpp:65-70`; `jam_TextEditor.cpp:611-624`; Part 9 Step 7 |
| A command with no handler in the focus chain of the focused component goes to the application | JUCE `juce_gui_basics/commands/juce_ApplicationCommandTarget.cpp:94-132` |

### The legacy shell that this part carries (ruling 3)

| Piece | Legacy | Where |
|---|---|---|
| The regions: a header of 7 rows at the top, a footer of 1 row at the bottom, the content between | `HeaderHeight = 7`, `FooterHeight = 1` | `legacy/ui/sizing.go:7-8`; `legacy/ui/layout.go:65-103` |
| The header rows | row 1: `📁 ` and the folder; the operation at the right. Row 2: the remote; the branch at the right. Row 3: a line of `─`. Row 4: the working tree; the LFS badge and the version at the right. Row 5: its description. Row 6: the timeline. Row 7: its description. With no repository the header is blank. | `legacy/ui/header.go:59-197`; `legacy/app/app_view_header.go:11-153`; `legacy/app/state_info.go:19-131` |
| The menu row: a key column of 12 cells aligned right, an emoji column of 3 cells, a label column of 21 cells; the block is centred; the selected row has the selection background on its label | `keyColWidth`, `emojiColWidth`, `labelColWidth` as literals | `legacy/ui/menu.go:16-18,58-81,102-117,138-150` |
| The menu movement wraps | `(start - 1 + n) % n`, `(start + 1) % n` | `legacy/app/handlers_global_menu.go:205-251` |
| Enter runs the selected item. Space is an alias. Escape does nothing in the menu. | | `handlers_global_menu.go:36-38,253-265`; `legacy/app/app_keys.go:187,193,199` |
| The item order of the Normal menu | working tree items, then timeline items | `legacy/app/menu_render_core.go:35-154` |
| The footer priority: the quit text, the clear text, the hint of the selected menu item, the key hints of the mode | | `legacy/app/footer.go:11-45` |
| Ctrl+C: during a run, a text; else the first press shows `Press Ctrl+C again to quit (2s timeout)` for 2 seconds, and the second press quits | `QuitConfirmTimeout` | `legacy/app/handlers_global_keys.go:52-78`; `legacy/app/messages_menu.go:35`; `legacy/constants.go:55` |
| The dialog: a title, an explanation, two buttons with upper-case labels; left, `h`, `y` select the first button; right, `l`, `n` select the second; Enter runs the selected button; Escape closes the dialog and runs nothing | | `legacy/ui/confirmation.go:134-251`; `legacy/app/handlers_config.go:14-53`; `handlers_global_menu.go:40-42,171-203` |
| The dialog width is the content width less 10 | `dialogWidth = c.Width - 10` | `legacy/ui/confirmation.go:147` |
| The console shows the command line, then the output; Escape during a run stops it and adds `Operation aborted by user`; Escape after the run returns to the menu | | `legacy/git/execute.go:101-102`; `handlers_global_menu.go:58-88,125-167` |
| The input: Escape with no text closes it; Escape with text shows `Press ESC again to clear input (2s timeout)`, and a second Escape clears the text | | `handlers_global_menu.go:90-110`; `messages_menu.go:36` |
| The smallest terminal | `MinWidth = 69`, `MinHeight = 19`; the text says `Terminal too small.` and `Resize to >= 70×20.` | `legacy/ui/sizing.go:5-6`; `legacy/ui/layout.go:137` |

### The legacy commands of this part

| Command | Key | Shown when | Dialog | Command lines | Where |
|---|---|---|---|---|---|
| `commit` | `c` | Normal, `Dirty` | none; the commit box | `add -A`, then the commit with the typed message | `legacy/app/menu_items.go:27`; `menu_render_core.go:62-89`; `op_commit.go:13-47` |
| `commitPush` | `p` | Normal, `Dirty`, a remote | none; the commit box | `add -A`, the commit, `push --progress` | `menu_items.go:35`; `menu_render_core.go:62-89`; `op_commit.go:50-94` |
| `resetDiscardChanges` | `ctrl+r` | Normal, `Dirty` | with a remote `confirm_discard_changes_remote_choice`, first button selected; else `confirm_discard_changes_local`, second button selected | first button: `reset --hard HEAD`, `clean -fd`. Second button of the remote choice: the lines of `replaceLocal`. | `menu_items.go:45`; `dispatch_git_basic.go:91-123`; `confirm_dialog_handlers.go:87-91,156-168`; `op_pull.go:107-136` |
| `forcePush` | `}` | a remote, and `Ahead` with `Clean`, or `Diverged` | `force_push`, first button selected | `push --progress --force`; with no upstream `push --progress -u origin <branch> --force` | `menu_items.go:64`; `menu_render_core.go:92-154`; `dispatch_git_basic.go:55-70`; `op_push.go:41-45` |
| `replaceLocal` | `x` | a remote, and `Behind` or `Diverged` | `hard_reset`, second button selected | `fetch --progress origin`, `reset --hard origin/<branch>`, `clean -fd` | `menu_items.go:92`; `menu_render_core.go:92-154`; `dispatch_git_basic.go:73-89`; `op_pull.go:46-100` |
| `pullMerge` | `[` | a remote, `Clean`, and `Behind` or `Diverged` | `pull_merge`, or `pull_merge_diverged` when `Diverged`; second button selected | `pull --no-rebase --progress` | `menu_items.go:84,102`; `menu_render_core.go:92-154`; `dispatch_git_basic.go:21-43`; `op_pull.go:14-22` |

- The dialog texts are in `legacy/app/messages_dialog.go:17-126`. Step 6 holds the six that this part uses.
- The menu rows are in `legacy/app/menu_items.go:8-239`. Step 7 holds the thirteen that this part uses.
- Legacy shows `push_auto_sync` in place of `push`, and `pull_merge_diverged` in place of `pull_merge`, when the timeline is `Diverged` (`menu_render_core.go:92-154`). Each pair has one key.

### Four legacy faults that this part corrects

Ruling 4: "kept the feature FIX THE DESIGN, DESIGN BY CONTRACT."

1. **The footer messages never showed.** The field `footerHint` has about 150 writes and no read (`legacy/app/footer.go:11-45`; `legacy/app/ui_state.go:11`). So each error text and each input hint was dead. This part shows a message in the footer.
2. **`q` quit inside a text input.** It is a global key, and the key lookup runs before the character insert (`legacy/app/app_keys.go:11,160-167`; `app_update_msg.go:110-117`). In this part `q` quits in the menu modes only.
3. **The size text and the size test differ.** The test is 69 by 19 and the text says 70 by 20 (`legacy/ui/sizing.go:5-6`; `layout.go:137`). This part tests what the text says.
4. **The console keys were blocked during a run**, while the footer showed `↑↓ scroll` (`legacy/app/handlers_console_ui.go:14,25,36,50`; `messages_menu.go:125-128`). In this part the console scrolls during a run.

### Words the framework already holds

| Word | Where |
|---|---|
| `Id::mode`, `Id::text`, `Id::title`, `Id::value`, `Id::id`, `Id::config`, `Id::directory` | `jam/cast/core.md` |
| `Id::message` | `jam/cast/gui.md` |
| `Id::isRunning`, `map::CommandId`, `text::Git`, the `git` rows, `map::Operation`, `map::WorkingTree`, `map::Timeline`, `map::Remote` | Parts 8 and 9 |
| `Chars::newline`, `Chars::carriageReturn`, `Chars::space` | `jam/cast/core.md:343,392,411`; `jam_subprocess/subprocess/jam_Subprocess.cpp:155` |

### Closed by a read

ARCHITECT can reverse each one.

1. **The header and the footer are TIT views in the form of `jam::TextEditorStatusBar`. The menu and the dialog are the stock components of JAM and JUCE.** ARCHITECT: *"tui could be build identically with any gui with juce native API contract using juce::Component base, juce::LookAndFeel (our Style etc)."* ARCHITECT, session 2026-11-14: *"no widget should be coupled with state machine"*; *"widget are dumb juce::Component, always styled with LookAndFeel"*; *"always being told"*. So a menu is the `jam::PopupMenu::MenuWindow` of JAM inside the view (Step 12) and a dialog is the `juce::AlertWindow` of JUCE, whose buttons are `juce::TextButton` (Step 14). The LookAndFeel paints both. The header and the footer paint Model rows, which the contract permits: "A view object that positions or paints itself from the Model's positional and appearance value-data is the materialisation of that data" (MANIFESTO E).
2. **The view names are `HeaderView` and `FooterView`.** The family is `SessionView`, `TabView`, `EditorView` (`end/Source/ui/`). TIT has no `MenuView` and no `ConfirmationView`.
3. **The colour ids of the TIT views are rows of JAM's `## ColourId`.** The sheet finds a colour id only through that table (`ling/Source/style.css:65-72`; `jam_StyleCustom.cpp:5-12`). Ruling 34 put the TIT menu modes in JAM's `## EditorMode` in the same way. The block is `0x4a…`, as the command ids of Part 9.
4. **The colour values are the legacy GFX theme** (`legacy/ui/theme_gfx.go:9-82`), written into `Source/style.css`. Part 13 adds the user theme folders and the user files; TIT ships this one theme (ruling 43).
5. **A run reads its bytes from a `jam::TerminalTTY` (ruling 59).** The decoder `jam::TerminalProcessor` reads only a TTY, and the terminal handles a carriage return and a newline itself. So `jam::Subprocess` does not change in this part, and Step 1 gives the TTY a child with no login wrapper and the exit status.
6. **No key row moves a highlight.** The stock menu moves its own highlight on the arrow keys, `j`, and `k` (ruling 61; Step 12). So no row of the sheet holds a key description with a space, and the key sheet of JAM does not change (Step 2).
7. **The mode names are `console`, `confirmation`, `input`,** the legacy modes `ModeConsole`, `ModeConfirmation`, `ModeInput` (`legacy/app/modes.go:23-25`).
8. **A run needs no rule of its own.** A run sets the mode to `console`, and no menu row applies there. So this part removes the `isRunning` term from `isAvailable` of Part 9 Step 7, and the `isRunning` event no longer builds the sheet. The `isRunning` row stays: the console cancel, the footer, and the quit read it.
9. **The console is the swap of the run chain (rulings 58 and 59).** A line of a run is one session of a `jam::TerminalTTY`. The decoder writes its bytes into the buffer and the swap, on the tick of a `jam::VirtualDevice`, and the console editor reads the swap through `bindToProcessor`. The Model has no row for the console text, and the processor sends no byte to a view. The end of a session is the rows `shellExited` and `numShellExits`; the view tells the processor at the event of the count (Step 10).
10. **The footer message is a `jam::ParameterText` of the Model root,** in the form of END and LING (`end/Source/end/ENDModel.cpp:8-9`; `ling/Source/LINGModel.cpp:10-25`). The views write it on the message thread. A message stays until the next TIT command runs.
11. **The menu texts are a config table, `Source/menu.md`,** read with `jam::ConfigDocument`, as the key sheet is (Part 7 Step 7). The view looks a row up by the command name. The `shortcut` cell is the legacy display text of the key.
12. **One menu row of a command changes with the state.** `push` shows the row `push auto sync`, and `pullMerge` shows the row `pull merge diverged`, when the timeline is `Diverged`. This is the menu side of ruling 35.
13. **The commit message goes to git through the file `.git/COMMIT_EDITMSG`.** `TITGitProcessor::run` takes command lines, and a command line splits at spaces (Part 9 Step 5), so a message cannot be one argument of a line. The file is the one that git itself uses for a commit message. Step 18 measures that the commit holds the typed text.
14. **`commitPush` has no fallback lines.** Legacy starts the pull only when the push fails (`legacy/app/op_commit.go:80-86`). `run` starts the fallback at the first line that fails, and that could be the commit. After a failed push the timeline is `Ahead` or `Diverged`, and `]` of Part 9 does the sync.
15. **In the commit box, Enter submits in normal mode and adds a line in insert mode.** Legacy: `Enter` is `submit` (`legacy/app/messages_menu.go:68-176`, rows `input_empty` and `input_filled`). Part 6 binds `return` to `splitLine` in insert mode. A key that the editor does not bind in normal mode goes to `TITView`.
16. **The input footer shows one hint row, `Enter submit │ Esc back`.** The second legacy row, `Esc clear`, needs the text of the editor as a Model row, and the clear text of the footer already states the Escape rule.
17. **The key hint rows of the footer are one text for each mode.** Legacy paints the key and its word in two colours (`legacy/ui/footer.go:98-143`). That is a presentation, and it locks at execution (ruling 7).
18. **The scroll status at the right of the console footer is not carried** (`legacy/app/footer.go:48-60`). The editor shows its own position (ruling 5).
19. **Not in this part, with the part that takes each:**
    - The banner region: Part 15. `BrailleComponent` does not exist before it.
    - `dirty_pull_merge`, the time travel menu, the pinned menu, the conflict resolver: Part 12.
    - `history`, `file_history`, their cache labels, the menu separator: Part 11.
    - `init`, `clone`, `add_remote` (ruling 35), the Config and Preferences menus, the setup wizard, the startup gate, the sync spinner of the header, `lfs checkout` after a hard reset, the user key override: Part 13 (`tit/SPEC.md:732-834`).
    - `Viewing commit from <date>` in the header: Part 12, which writes the date.

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: JAM, a program run on `jam::TerminalTTY` (ruling 59)
**Scope:** `jam/jam_terminal/tty/jam_TerminalTTY.h`; `jam/jam_terminal/tty/native/jam_TerminalTTY_posix.h`, `jam_TerminalTTY_posix.cpp`; `jam/jam_terminal/tty/native/jam_TerminalTTY_windows.h`, `jam_TerminalTTY_windows.cpp`; `eve/Source/TerminalMemoryTTY.h`, `eve/Source/TerminalMemoryTTY.cpp`.
**Action:** The decoder reads only a `jam::TerminalTTY` (`jam_terminal/processor/jam_TerminalProcessor.cpp:66,122-133`). A run of TIT needs two things that the TTY does not give today: a child with no login wrapper, and the exit status of the child.

1. `jam_TerminalTTY.h`: replace `virtual void close() = 0;` (`:104`) with `virtual int close() = 0;`.
2. `jam_TerminalTTY_posix.h`:
   - Public, before the destructor (`:82`): `explicit TerminalUnixTTY (bool newShouldUseLoginShell = true);`.
   - Replace `void close() override;` (`:125`) with `int close() override;`.
   - Private, after `pid_t childProcess { -1 };` (`:172`): `const bool shouldUseLoginShell;`.
3. `jam_TerminalTTY_posix.cpp`:
   - Before the destructor (`:173`), define the constructor:

```cpp
TerminalUnixTTY::TerminalUnixTTY (bool newShouldUseLoginShell)
    : shouldUseLoginShell (newShouldUseLoginShell)
{
}
```

   - `runChildProcess` (`:155-169`): add the last parameter `[[maybe_unused]] bool shouldUseLoginShell`. Replace the block of `:163-165` with:

```cpp
#ifdef __APPLE__
    if (shouldUseLoginShell)
        runLoginShell (shell, argv);
#endif
```

   - `open` (`:263`): the call becomes `runChildProcess (shell.toRawUTF8(), argv.data(), workingDirectory.toRawUTF8(), environment, shouldUseLoginShell);`.
   - Replace the body of `close` (`:286-316`) with:

```cpp
int TerminalUnixTTY::close()
{
    int status { 0 };
    auto reaped { childProcess > 0 ? waitpid (childProcess, &status, WNOHANG) : -1 };
    const auto isChildRunning { reaped == 0 };

    if (isChildRunning)
        kill (childProcess, SIGKILL);

    if (master >= 0)
    {
#if JUCE_MAC
        if (readSource != nullptr)
        {
            dispatch_source_cancel (readSource.get());

            if (isReadSourceSuspended.exchange (false))
                dispatch_resume (readSource.get());

            dispatch_semaphore_wait (cancelled.get(), DISPATCH_TIME_FOREVER);
            readSource.reset();
        }
#endif

        ::close (master);
        master = -1;
    }

    if (isChildRunning)
        reaped = waitpid (childProcess, &status, 0);

    childProcess = -1;

    return reaped > 0 and WIFEXITED (status) ? WEXITSTATUS (status) : EXIT_FAILURE;
}
```

4. `jam_TerminalTTY_windows.h`: the declaration of `close()` returns `int`.
5. `jam_TerminalTTY_windows.cpp`, `close` (`:748-808`):
   - The return type is `int`. After the profiler zone (`:750-752`), add `int result { EXIT_FAILURE };`.
   - Replace the statement of `:775-779` with:

```cpp
            if (GetExitCodeProcess (process.get(), &exitCode) and exitCode == STILL_ACTIVE)
            {
                TerminateProcess (process.get(), EXIT_FAILURE);
                WaitForSingleObject (process.get(), instantKillChildWaitMs);
            }

            if (GetExitCodeProcess (process.get(), &exitCode) and exitCode != STILL_ACTIVE)
                result = static_cast<int> (exitCode);
```

   - After `writePending = false;` (`:807`), add `return result;`.
6. `eve/Source/TerminalMemoryTTY.h` and `.cpp` (`:50-53`): `close()` returns `int`, and its body is `return tty->close();`.

- The login wrapper hides the exit status. Measured on macOS, session 2026-11-14 (ruling 59): `/usr/bin/login -fqlp <user> /bin/bash --noprofile --norc -c "exec -l /bin/sh -c 'echo ran; exit 7'"` prints `ran` and returns 0; the same child with no `login` returns 7.
- With `shouldUseLoginShell` false, the child is `execvp (shell, argv)` (`:167`). Windows has no wrapper: `CreateProcessW` runs `"shell" args` (`jam_TerminalTTY_windows.cpp:576-603`).
- `close()` gives the status of the call that reaps the child. The existing body already makes both calls: `waitpid` with `WNOHANG`, then the kill and the blocking `waitpid` (`jam_TerminalTTY_posix.cpp:288-313`). A child that ended by itself gives `WIFEXITED` and its own status, in the first call or in the second. A child that the signal ended is not `WIFEXITED`, so it is a failure. On Windows the code after the wait is the code of the child, or `EXIT_FAILURE` when `TerminateProcess` ended it. A second `close()` of one session is a failure too: `childProcess` is -1, and `process` is null.
- A read of a TTY with no session gives -1: `close()` sets `master` to -1 (`jam_TerminalTTY_posix.cpp:308-309`), and `read` returns -1 for the end of the stream and for each error that is not `EAGAIN` (`:318-339`).
- `close()` waits on the `cancelled` semaphore of the read source (`:296-304`). So no caller calls it inside the tick of `onBytesReady` (Step 10).
- Windows needs no sidecar file: "If the directory does not contain conpty.dll, the TerminalTTY falls back to the inbox kernel32 ConPTY" (`jam_TerminalTTY_windows.h:50-51`).

**Validation:**
- EVE and END construct `TerminalUnixTTY` with no argument, so their shell keeps the login wrapper. No caller of EVE or END reads the result of `close()`. The destructor of each backend ignores it.
- `jam::TerminalTTY` has two backends in JAM and one more class in EVE (`jam/docs/xml/classjam_1_1TerminalTTY.xml:5-6`; `eve/Source/TerminalMemoryTTY.h`). All three return `int` after this step.
- `jam::Subprocess` does not change in this part.
- No bail-out guard. Each return carries the status.
- New names and their families: `shouldUseLoginShell`, `newShouldUseLoginShell` (`const bool shouldUseInlineImages` and the constructor parameter `bool newShouldUseInlineImages = false`, Part 3, inventory, row "The switch", and Part 3 Step 2; `runLoginShell`, `jam_TerminalTTY_posix.cpp:101`); `status`, `reaped`, `result` (local).

### Step 2: JAM, the key sheet
**Scope:** none. No file changes in this step.
**Action:** `setKeyPresses` of `jam/jam_gui/commands/jam_CommandManager.cpp:5-41` stays as it is. A row value is split at spaces, so a key description with a space, for example `cursor up`, cannot be a cell value. No key row of Parts 10 to 13 has such a value: the stock menu moves its own highlight on the arrow keys (Step 12), and no row of the sheet does it (ruling 61).

**Validation:**
- No value cell of `Source/keys.md` holds a double quote or a key description with a space.
- `jam_CommandManager.cpp` has no change from this part.

### Step 3: JAM, the default mode of an editable editor
**Scope:** `jam/jam_gui/widgets/jam_TextEditor.h`, `jam/jam_gui/widgets/jam_TextEditor.cpp`.
**Action:**
1. Header: after `void setMode (int newMode);` (`:184`), declare `void setDefaultMode (int newDefaultMode);`.
2. Source: after the body of `setMode` (`:146-149`), define:

```cpp
void TextEditor::setDefaultMode (int newDefaultMode)
{
    terminal.setProperty (Id::defaultMode, newDefaultMode, nullptr);
}
```

- `setLiveBlock (std::unique_ptr<Document>)` sets `Id::defaultMode` to `reader` (`jam_TextEditor.cpp:1148`). The command `defaultMode` goes to that row (`:1051-1054`). An editable editor must return to `normal` from visual mode, so its owner sets the row after `setLiveBlock`.
- `PLAN-tit-6-text-editor-insert.md`, Step 6, owner contract, items 3 and 5, now state this call.

**Validation:**
- `setDocument<>` and `bindToProcessor` do not change. EVE calls neither `setLiveBlock (std::unique_ptr<Document>)` nor `setDefaultMode`.
- New names and their family: `setDefaultMode`, `newDefaultMode` (`setMode (int newMode)`, `setModeToDefault`, `jam_TextEditor.h:184`; `.cpp:1051`).

### Step 4: JAM, the modes and the colour ids of the views
**Scope:** `jam/cast/gui.md`, table `## EditorMode`; `jam/cast/style.md`, table `## ColourId`; then `jam/generated/` through `cast cast/spell.md` from the JAM root.
**Action:**
1. `## EditorMode`: after the `menuPinned` row of Part 9 Step 1, add `console` 15, `confirmation` 16, `input` 17.
2. `## ColourId`: after the `textEditorStatusBarInsertColourId` row of Part 6 Step 1, add 21 rows in the column widths of the table. The name of a row is the view name with a lower-case first letter, the word of the row below, and `ColourId`. The value is the same words in lower case with spaces, in backticks, as `text editor status bar normal colour id` is for `textEditorStatusBarNormalColourId` (`jam/cast/style.md:262`). The keys run from `0x4a00001` in this order:

   | View | Words, in key order | Keys |
   |---|---|---|
   | `headerView` | `Background`, `Text`, `Label`, `Dimmed`, `Accent`, `Separator`, `Clean`, `Dirty`, `InSync`, `Ahead`, `Behind`, `Ready`, `NotRepo`, `Conflicted`, `Merging`, `Rebasing`, `DirtyOperation`, `TimeTravel`, `Warning` | `0x4a00001` to `0x4a00013` |
   | `footerView` | `Background`, `Text` | `0x4a00019`, `0x4a0001a` |

   Example: `| headerViewDirtyOperationColourId | 0x4a00011 | \`header view dirty operation colour id\` |`.
   - The keys `0x4a00014` to `0x4a00018` and `0x4a0001b` to `0x4a00020` stay free. A menu has no colour id of TIT: the popup menu of JAM paints with the popup menu ids that the table already holds (`jam/cast/style.md:120-124`; Step 12). A dialog has no colour id of TIT: the `juce::AlertWindow` of Step 14 paints with the alert window ids and the text button ids that the table already holds (`jam/cast/style.md:112-115,223-225`; Step 14).
3. Run `cast cast/spell.md` from the JAM root.

**Validation:**
- `map::EditorMode` holds `console`, `confirmation`, `input` with the keys 15, 16, 17. The rows 0 to 14 keep their keys.
- `map::ColourId` holds the 21 names. No other row of `## ColourId` holds a key in the block `0x4a…` of seven hexadecimal places.
- Each colour word is a legacy theme word (`legacy/ui/theme.go:4-80`) or a state word of Part 8 Step 2.
- No JAM source file changes in this step.

### Step 5: TIT, the words, the command ids, the files, the command lines
**Scope:** `tit/cast/identifiers.md`, `tit/cast/files.md`, `tit/cast/git.md`, `tit/project-info.md`.
**Action:**
1. `cast/identifiers.md`, table `## identifiers`: after the `isRunning` row of Part 9 Step 2, add 8 rows of type `@id` with an empty description. The value of each row is its name in backticks. Skip a name that JAM already declares.

   | Name | Family |
   |---|---|
   | `console` | the type of the new child of the Model; the legacy mode and view (`legacy/app/modes.go:22-25`) |
   | `shortcut`, `emoji` | the columns of `Source/menu.md`; legacy fields `Shortcut`, `Emoji` (`legacy/ui/menu_item.go:4-13`) |
   | `commandLines` | the parameter `commandLines` of Part 9 Step 5 |
   | `isQuitConfirmActive`, `isClearConfirmActive` | legacy fields `quitConfirmActive`, `ClearConfirming` (`legacy/app/footer.go:14-25`), with the `is` prefix of a boolean (NAMES.md Rule 1) |
   | `remoteUrl` | legacy `GetRemoteURL` (`legacy/app/app_view_header.go:24-32`) |
   | `pullMergeDiverged` | legacy id `pull_merge_diverged` (`legacy/app/menu_items.go:102`) |

2. `cast/identifiers.md`, table `## CommandId`: after the `rebaseAbort` row, add these rows. The value of each row is its name in backticks.

   | name | key | Legacy source of the name |
   |---|---|---|
   | `commit` | `0x4a00007` | `commit` (`legacy/app/menu_items.go:27`) |
   | `commitPush` | `0x4a00008` | `commit_push` (`:35`) |
   | `resetDiscardChanges` | `0x4a00009` | `reset_discard_changes` (`:45`) |
   | `forcePush` | `0x4a0000a` | `force_push` (`:64`) |
   | `replaceLocal` | `0x4a0000b` | `replace_local` (`:92`) |
   | `pullMerge` | `0x4a0000c` | `pull_merge` (`:84`) |
   | `cancel` | `0x4a00013` | the footer word of Escape, `cancel` (`legacy/app/messages_menu.go:68-176`) |
   | `inputSubmit` | `0x4a00014` | `handleInputSubmit` (`legacy/app/app.go:208`) |

   - The keys `0x4a0000d` to `0x4a00012` stay free. A menu has no move command and no enter command: `jam::PopupMenu::MenuWindow` takes those keys itself (rulings 61 and 63; Step 12). A dialog has no command: the `juce::AlertWindow` of Step 14 has its own keys (rulings 65 and 66).
   - No word `confirmation`, `selectedIndex`, `explanation`, `yes`, `no`, `yesCommandLines`, or `noCommandLines` is declared. The dialog has no row, so no code reads such a word.

3. `cast/files.md`, table `## files`: after the last row of Part 8 Step 3, add two rows of type `@string`: `titcMenu`, value `` `menu.md` ``, description `Menu texts.`; `commitMessage`, value `` `COMMIT_EDITMSG` ``, description `The file in the repository data folder that holds the message of the next commit.`
4. `project-info.md`, table `## binary`: after the `keys` row of Part 7 Step 1, add `| menu | Source/menu.md | Menu texts, embedded as BinaryData |`.
5. `cast/git.md`: after the last row of Part 9 Step 3, add these rows:

   | name | value | description |
   |---|---|---|
   | `addAll` | `` `git add -A` `` | Stages every change of the tree. |
   | `commitFile` | `` `git commit -F .git/COMMIT_EDITMSG` `` | Makes a commit with the message of the file. |
   | `force` | `` `--force` `` | The push option that replaces the remote branch. |
   | `fetchOrigin` | `` `git fetch --progress origin` `` | Fetches the branches of origin. |
   | `resetHardHead` | `` `git reset --hard HEAD` `` | Sets the tree and the index to HEAD. |
   | `resetHardOrigin` | `` `git reset --hard origin/` `` | The start of the reset to the remote branch; the branch name follows with no space. |
   | `clean` | `` `git clean -fd` `` | Removes the untracked files and folders. |
   | `remoteUrl` | `` `git remote get-url origin` `` | Prints the address of origin. |

**Validation:**
- Each command line equals its legacy literal (inventory, "The legacy commands of this part"). `remoteUrl` is the line of the legacy `GetRemoteURL` (`legacy/git/exec_utils.go:86-92`). The exception is `commitFile` (Closed by a read, item 13).
- No word of JAM is declared again (`Id::text`, `Id::title`, `Id::message`, `Id::value`, `Id::id`).
- Each new word has its cited family.

### Step 6: TIT, the texts
**Scope:** `tit/cast/text.md`, `tit/cast/spell.md`.
**Action:**
1. `cast/text.md`: after the table `## diagnostics`, add three tables in its form (`ling/cast/text.md:1-9`): `## header`, `## footer`, `## confirmation`. The briefs: `@brief Texts of the header.`, `@brief Texts of the footer and of the console.`, `@brief Texts of the confirmation dialogs.` No value starts or ends with a space.
2. `## header`, in this order (`legacy/app/app_view_header.go:24-124`; `legacy/app/state_info.go:19-131`; `legacy/app/messages_state.go:120-139`; `legacy/ui/header.go:115`; Part 2, "Other text"):

   | name | value | name | value |
   |---|---|---|---|
   | `directoryMark` | `📁` | `noRemote` | `🔌 NO REMOTE` |
   | `remoteMark` | `🔗` | `branchMark` | `🌿` |
   | `notAvailable` | `N/A` | `separator` | `─` |
   | `lfs` | `LFS` | `lfsNotReady` | `LFS ⚠️` |
   | `clean` | `✅ Clean` | `cleanDescription` | `No local changes` |
   | `dirty` | `📝 Dirty` | `dirtyDescription` | `Local changes present` |
   | `modifiedMark` | `●` | `timelineNone` | `🔌 N/A` |
   | `inSync` | `🔗 In sync` | `inSyncDescription` | `In sync with remote` |
   | `ahead` | `🌍 Local ahead` | `aheadDescription` | `commit(s) ahead` |
   | `behind` | `🪐 Local behind` | `behindDescription` | `commit(s) behind` |
   | `diverged` | `💥 Diverged` | `noRemoteDescription` | `No remote configured.` |
   | `aheadMark` | `⬆` | `behindMark` | `⬇` |
   | `divergedAheadMark` | `↑` | `divergedBehindMark` | `↓` |
   | `ready` | `🟢 READY` | `notRepo` | `🔴 NOT REPO` |
   | `conflicted` | `⚡ CONFLICTED` | `merging` | `🔀 MERGING` |
   | `rebasing` | `🔄 REBASING` | `dirtyOperation` | `⚡ DIRTY OP` |
   | `timeTravel` | `🌀 TIME TRAVEL` | `detached` | `DETACHED` |
   | `detachedAt` | `DETACHED @` | `pinnedAt` | `PINNED @` |

3. `## footer`, in this order (`legacy/app/messages_menu.go:35-36,45,68-176`; `legacy/app/messages_error.go:4-59`; `legacy/app/handlers_global_menu.go:64-66`; `legacy/ui/layout.go:137`):

   | name | value |
   |---|---|
   | `quitConfirm` | `Press Ctrl+C again to quit (2s timeout)` |
   | `clearConfirm` | `Press ESC again to clear input (2s timeout)` |
   | `operationInProgress` | `Operation in progress. Please wait for completion.` |
   | `commitMessageEmpty` | `Commit message cannot be empty` |
   | `aborted` | `Operation aborted by user` |
   | `tooSmall` | `Terminal too small. Resize to >= 70×20.` |
   | `consoleRunning` | `↑↓ scroll  │  Esc abort` |
   | `consoleComplete` | `↑↓ scroll  │  Esc back` |
   | `confirmation` | `←→ select  │  Enter confirm  │  Esc cancel` |
   | `input` | `Enter submit  │  Esc back` |

4. `## confirmation`: for each of the six dialogs below, four rows named with the prefix of the dialog and the suffixes `Title`, `Explanation`, `Yes`, `No`. Each value is the Go literal of its field at the cited lines of `legacy/app/messages_dialog.go`, with each `\n` kept as `\n`.

   | Prefix | Legacy key | Lines |
   |---|---|---|
   | `forcePush` | `force_push` | `18-23` |
   | `hardReset` | `hard_reset` | `24-29` |
   | `pullMerge` | `pull_merge` | `36-41` |
   | `pullMergeDiverged` | `pull_merge_diverged` | `42-47` |
   | `discardChangesRemoteChoice` | `confirm_discard_changes_remote_choice` | `96-101` |
   | `discardChangesLocal` | `confirm_discard_changes_local` | `102-107` |

5. `cast/spell.md`, `## output`: after the `@text:diagnostics` group, add three groups. Each one is the `@text:diagnostics` group (`ling/cast/spell.md:168`) with `@text:header`, `@text:footer`, or `@text:confirmation` in its list cell, and `- name: Header`, `- name: Footer`, or `- name: Confirmation`. The file cell is `@Text`.

**Validation:**
- After the build of Step 18, `Text.h` holds `struct Header` with 36 constants, `struct Footer` with 10, and `struct Confirmation` with 24.
- `text::Confirmation::forcePushExplanation` holds four newline characters, and no backslash.
- Each value equals its legacy text. A value of the `## header` table that legacy wrote with a number in it holds only the words; Step 11 adds the number.

### Step 7: TIT, the menu texts and the key rows
**Scope:** `tit/Source/menu.md` (new), `tit/Source/keys.md`.
**Action:**
1. `Source/menu.md`: the front matter of `Source/keys.md`, then the table `## menu` with the columns `key`, `type`, `value`, `shortcut`, `emoji`, in the grid form of `Source/keys.md`. Each `type` cell is `string`. The rows, from `legacy/app/menu_items.go:27-112,216-246`:

   | key | value | shortcut | emoji |
   |---|---|---|---|
   | `commit` | Commit changes | `c` | 📝 |
   | `commit push` | Commit and push | `p` | 🚀 |
   | `reset discard changes` | Discard all changes | `ctrl + r` | 💥 |
   | `push` | Push to remote | `]` | 📤 |
   | `push auto sync` | Push (auto sync) | `]` | 📤 |
   | `pull merge` | Pull (fetch + merge) | `[` | 📥 |
   | `pull merge diverged` | Pull (merge) | `[` | 📥 |
   | `force push` | Force push | `shift + ]` | 💥 |
   | `replace local` | Replace local | `x` | 💥 |
   | `finalize merge` | Finalize merge | `f` | ✅ |
   | `abort merge` | Abort merge | `a` | ↩️ |
   | `rebase continue` | Continue rebase | `c` | ▶️ |
   | `rebase abort` | Abort rebase | `a` | ↩️ |

   A `shortcut` cell holds the text only, with no backtick.
2. `Source/keys.md`: after the last row of Part 9 Step 4, add these rows. `MENUS` stands for the eight mode names `menuNotRepo menuNormal menuMerging menuConflicted menuRebasing menuTimeTraveling menuDirtyOperation menuPinned`, written in full in each cell.

   | key | type | value | mode | description |
   |---|---|---|---|---|
   | `quit` | `string` | `q` | MENUS | Quit titc. |
   | `commit` | `string` | `c` | `menuNormal` | Commit changes. |
   | `commit push` | `string` | `p` | `menuNormal` | Commit and push. |
   | `reset discard changes` | `string` | `ctrl+r` | `menuNormal` | Discard all changes. |
   | `force push` | `string` | `}` | `menuNormal` | Force push. |
   | `replace local` | `string` | `x` | `menuNormal` | Replace local. |
   | `pull merge` | `string` | `[` | `menuNormal` | Pull. |
   | `cancel` | `string` | `escape` | `console input terminal` | Stop, close, or go back. |
   | `input submit` | `string` | `return` | `input` | Submit the text. |
   | `default mode` | `string` | `escape` | `visual open` | Leave visual mode. |
   | `insert before` | `string` | `i` | `normal` | Insert before the cursor. |
   | `insert after` | `string` | `a` | `normal` | Insert after the cursor. |
   | `insert line start` | `string` | `I` | `normal` | Insert at the start of the line. |
   | `insert line end` | `string` | `A` | `normal` | Insert at the end of the line. |
   | `open line below` | `string` | `o` | `normal` | Open a line below. |
   | `open line above` | `string` | `O` | `normal` | Open a line above. |
   | `delete action` | `string` | `d` | `normal visual` | Delete. |
   | `change` | `string` | `c` | `normal visual` | Change. |
   | `delete character` | `string` | `x` | `normal visual` | Delete the character. |
   | `paste after` | `string` | `p` | `normal visual` | Paste after the cursor. |
   | `paste before` | `string` | `P` | `normal visual` | Paste before the cursor. |
   | `undo` | `string` | `u` | `normal` | Undo. |
   | `redo` | `string` | `ctrl+r` | `normal` | Redo. |

   - No row moves a menu or runs its item. The open `jam::PopupMenu::MenuWindow` takes down, `j`, up, `k`, return, space, and Escape in its own `keyPressed` (rulings 61 and 63; Step 12). Each other key goes up to `TITView` and to this sheet.
   - No row belongs to a dialog. Enter, Escape, and the first letter of each button label are keys of the `juce::AlertWindow` buttons (Step 14, the table; JUCE `juce_gui_basics/lookandfeel/juce_LookAndFeel_V2.cpp:410-418`). So the legacy keys `y`, `n`, Enter, and Escape need no row.
   - The second `quit` row keeps the row of Part 7 (`ctrl+c`, every mode) and adds `q` in the menus only (legacy fault 2).
   - The `default mode` row takes the place of the JAM default row, because the sheet build drops a default row that the config names (Part 7 Step 7, item 3). So `escape i a I A o O` no longer run `defaultMode` in normal mode, and the six insert rows apply. A reader editor still leaves visual mode with Escape.
   - The thirteen edit rows are the tested keys of EVE's sheet (Part 6, inventory, last row of "The old bodies"; Part 6, closed item 2).
   - The sheet holds the thirteen edit rows only while the mode of TIT shows an editable editor (Step 16, item 6). A read-only editor can enter its visual mode with `v` (`jam/cast/terminal.md:1913-2063`), and the rows `d`, `c`, `x`, `p`, `P` apply in `visual`. With no edit row in the sheet, no key of a read-only view reaches an edit command (Part 6, closed item 3).
   - `menuNormal` and `normal` are two modes of two states. `c` in the menu reaches `TITView` with the Model root. `c` in the commit box reaches the editor with its `terminal` tree (inventory, key sheet facts).
   - Escape in the normal mode of the commit box, and in the reader mode of the console, has no editor row. It goes to `TITView`, which runs `cancel`. While a run runs, the console editor is in `terminal` mode: it gives the key to the key sheet with its own mode, `terminal`, before it sends bytes to the tool (`jam_gui/widgets/jam_TextEditor.cpp:634-637`). So the `cancel` row names `terminal` too, and Escape stops the run.

**Validation:**
- Each menu text equals its legacy text. Each default key equals its legacy key (`legacy/app/menu_items.go`; `legacy/app/app_keys.go:8-19,37-77`).
- In one mode, each key has one row. `return` has one row in `input`, and no row in a menu mode or in `confirmation`.
- Each mode name is a row of `map::EditorMode` after Step 4 and Part 9 Step 1.
- No value holds a space outside a quoted description.

### Step 8: TIT, the colours
**Scope:** `tit/Source/style.css`.
**Action:**
1. In the `:root` block, add each palette name below that the block does not hold, in the form `--white: #FFFFFF;` (`ling/Source/style.css:24`). The names are the legacy names of the values (`legacy/ui/theme_gfx.go:9-82`):

   | name | value | name | value |
   |---|---|---|---|
   | `bunker` | `#090D12` | `dark` | `#1B2A31` |
   | `paradiso` | `#4E8C93` | `dolphin` | `#8CC9D9` |
   | `mediterranea` | `#33535B` | `caribbeanBlue` | `#01C2D2` |
   | `lagoon` | `#519299` | `trappedDarkness` | `#112130` |
   | `preciousPersimmon` | `#FC704C` | `blueBikini` | `#00C8D8` |
   | `safflower` | `#F2AB53` | `emeraldGreen` | `#4ECB71` |
   | `brighterMutedTeal` | `#7EB8C5` | `darkText` | `#0D1418` |

2. In the `.appearance` block, after the `--TextEditor--` customs, add one custom for each colour id of Step 4, in the form `--HeaderView--backgroundColourId: var(--bunker);`:

   | View | Custom word and palette name |
   |---|---|
   | `HeaderView` | `background` bunker; `text` paradiso; `label` dolphin; `dimmed` mediterranea; `accent` caribbeanBlue; `separator` dark; `clean` caribbeanBlue; `dirty` preciousPersimmon; `inSync` caribbeanBlue; `ahead` blueBikini; `behind` safflower; `ready` emeraldGreen; `notRepo` preciousPersimmon; `conflicted` preciousPersimmon; `merging` blueBikini; `rebasing` blueBikini; `dirtyOperation` preciousPersimmon; `timeTravel` safflower; `warning` safflower |
   | `PopupMenu` | `background` bunker; `text` dolphin; `headerText` mediterranea; `highlightedBackground` brighterMutedTeal; `highlightedText` bunker |
   | `FooterView` | `background` bunker; `text` lagoon |
   | `AlertWindow` | `background` trappedDarkness; `text` paradiso |
   | `TextButton` | `button` dark; `buttonOn` brighterMutedTeal; `textColourOff` paradiso; `textColourOn` darkText |

   The four `--TextButton--` customs are `--TextButton--buttonColourId`, `--TextButton--buttonOnColourId`, `--TextButton--textColourOffId`, and `--TextButton--textColourOnId`: the id names of JUCE end in `ColourId`, `ColourOffId`, and `ColourOnId` (`jam/cast/style.md:112-115`).

3. Set the three `--TextEditor--` customs of the block to the legacy console and input colours: `backgroundColourId` bunker, `textColourId` a new palette name `terminalText` `#999999` (`theme_gfx.go:19,77`), `highlightColourId` brighterMutedTeal.

**Validation:**
- Each value is the legacy value of its theme field: `mainBackgroundColor`, `contentTextColor`, `labelTextColor`, `dimmedTextColor`, `accentTextColor`, `separatorColor`, `statusClean`, `statusDirty`, `timelineSynchronized`, `timelineLocalAhead`, `timelineLocalBehind`, `operationReady`, `operationNotRepo`, `operationConflicted`, `operationMerging`, `operationRebasing`, `operationDirtyOp`, `operationTimeTravel`, `outputWarningColor`, `menuSelectionBackground`, `footerTextColor`, `confirmationDialogBackground`, `inlineBackgroundColor`, `buttonSelectedTextColor`, `outputStdoutColor` (`legacy/ui/theme_gfx.go:9-82`).
- The spelling of each custom matches its row of Step 4, as `--TextEditorStatusBar--normalColourId` matches `textEditorStatusBarNormalColourId`. The five `--PopupMenu--` customs match the rows `popupMenuBackgroundColourId`, `popupMenuTextColourId`, `popupMenuHeaderTextColourId`, `popupMenuHighlightedBackgroundColourId`, and `popupMenuHighlightedTextColourId` of JAM (`jam/cast/style.md:120-124`). The two `--AlertWindow--` customs match `alertWindowBackgroundColourId` and `alertWindowTextColourId` (`:223-224`). The four `--TextButton--` customs match `textButtonButtonColourId`, `textButtonButtonOnColourId`, `textButtonTextColourOffId`, and `textButtonTextColourOnId` (`:112-115`).
- The shortcut text of a menu item has the text colour of the item. The LookAndFeel paints one colour for each item (Step 12). The legacy colour of the key column, `accentTextColor`, has no custom.
- After Step 18, no cell of a view shows magenta.

### Step 9: TIT, the rows of the views
**Scope:** `tit/Source/TITModel.cpp`, `tit/Source/TITGitProcessor.cpp`.
**Action:**
1. `TITModel.cpp`, at the end of the constructor body, after the `isRunning` row of Part 9 Step 5:

```cpp
    createAndAddParameter<jam::ParameterText> (state, Id::message, juce::String {}, maximumMessageLength);
    createAndAddParameter<jam::Parameter<int>> (state, Id::isQuitConfirmActive, 0);
    createAndAddParameter<jam::Parameter<int>> (state, Id::isClearConfirmActive, 0);
    state.setProperty (Id::commandLines, juce::String {}, nullptr);

    for (const auto& type : { Id::console })
    {
        auto runRow { getOrCreateChildWithName (jam::Format::toType (type)) };

        createAndAddParameter<jam::Parameter<int64_t>> (runRow, Id::winsize, int64_t { 0 });
        createAndAddParameter<jam::Parameter<int>> (runRow, Id::shellExited, 1);
        createAndAddParameter<jam::Parameter<int>> (runRow, Id::numShellExits, 0);
    }
```

   - The Model has no row for a dialog: no title, no message, no button label, no command line of a button, and no selected button (ruling 65). `TITView` gives the texts when it shows the box, the callback of the box holds what a button runs, and the buttons are `juce::TextButton`s of the `juce::AlertWindow` (ruling 66; Step 14).
   - A number, a flag, and the footer message are parameters (ruling 57; Closed by a read, item 10). `maximumMessageLength` is `static constexpr int maximumMessageLength { 4096 };` in the private section of `TITModel.h`, the name and the value of END (`end/Source/end/ENDModel.cpp:8-9`).
   - Each run chain has one child of the root with three parameters (Step 10). This part has one chain, `console`. Each later part adds the type of its chain to the list of the loop. `shellExited` and `numShellExits` are the rows that the tick of the chain stores. `shellExited` starts at 1, because no session exists before the first run. `Id::shellExited` and `Id::winsize` are words of JAM (`jam/cast/terminal.md:89,103`). Add the row of `numShellExits` to `cast/identifiers.md`, in the table of Step 5, item 1: type `@id`, name `numShellExits`, value `` `numShellExits` ``.
   - `winsize` is the grid of the pane of the chain, in the form of the `winsize` row of Part 7 Step 2. The view of the pane writes it (Step 14).
   - The Model has no row for a menu: no item list, no highlight, and no scroll offset. The items of a menu are built from the rows of `Source/menu.md` and the `git` rows at each event (Step 16), and the highlight is rendering state of `jam::PopupMenu::MenuWindow` (ruling 61; Step 12).
   - `commandLines` stays a plain row of the tree: a list with no size limit. Only views write and read it, on the message thread. A `jam::ParameterText` cuts a text at its size (`jam_ParameterText.h:49-57`), and a cut command line is a wrong command.
2. `TITModel::getGitTexts()` of Part 8 Step 5: add `Id::remoteUrl` to the table.
3. `TITGitProcessor::setRepositoryGit` of Part 8 Step 7: after the store of `currentHash`, add `storeValue (Id::remoteUrl, hasRemote ? getOutput (directory, jam::ChildProcess::getArguments (text::Git::remoteUrl)).value_or (juce::String {}) : juce::String {});`.

**Validation:**
- A detection of Part 8 stores none of the new rows of the root (Part 8 Step 7, item 3).
- Each row holds value data: numbers, flags, and texts (MANIFESTO E, Materialisation). No row holds a component.
- The `git` child has 16 rows.
- New names and their families: `numShellExits` (`numKeyPresses`, Part 7; `numCommands`, `eve/Source/layout/parameters.md:45`; the row `shellExited`); `consoleRow` (`gitRow`, Part 8 Step 5); `maximumMessageLength` (END, verbatim).

### Step 10: `TITRunProcessor`, the run chain; `TITGitProcessor`, the run and the cancel
**Scope:** `tit/Source/TITRunProcessor.h`, `tit/Source/TITRunProcessor.cpp` (both new); `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:** A run is the chain of EVE (rulings 58 and 59): a `jam::TerminalTTY` runs one command line, the decoder `jam::TerminalProcessor` writes its bytes into a `jam::TerminalBuffer`, and a `jam::TerminalSwap` holds the lines. The chain runs on the serial context of a `jam::VirtualDevice`. `TITRunProcessor` is that chain. Each of its members and functions has its sibling in `eve/Source/EVETerminalProcessor.h:30-369` and `.cpp`, with the same name. One object is one chain with one swap, so each pane that shows the output of the tool has its own object: the console in this part, and the panes of Parts 11 and 12.

1. `TITRunProcessor.h`, with `#pragma once`, `#include <JuceHeader.h>`, `#include "generated/Generated.h"`, and `#include "TITModel.h"`:

```cpp
class TITRunProcessor : private jam::AsyncUpdater,
                        private juce::AudioIODeviceCallback
{
public:
    TITRunProcessor (TITModel& modelToUse, const juce::Identifier& typeToUse);
    ~TITRunProcessor() override;

    void prepare();
    void releaseResources();

    template <typename Callback>
    void restart (Callback&& callback)
    {
        const auto isPlaying { device.isPlaying() };

        device.stop();
        callback();

        if (isPlaying)
            start();
    }

    void restart();
    void clearHistory();
    void run (const juce::String& commandLine);
    int close();

    jam::Registry::Getters userInterfaceGetters;
    jam::Registry::Events chainEvents;

private:
    bool writePendingData();
    void start();
    void process() noexcept;
    void audioDeviceIOCallbackWithContext (const float* const*, int, float* const*, int, int, const juce::AudioIODeviceCallbackContext&) override;
    void audioDeviceAboutToStart (juce::AudioIODevice*) override;
    void audioDeviceStopped() override;
    void handleAsyncUpdate() override;
    void registerEvents();
    void registerUserInterfaceGetters();
    void registerChainEvents();
    void addInput (const char* bytes, int size);
    void applyPendingInput();
    void prepare (jam::TerminalWinsize winsize);
    jam::TerminalWinsize getWinsize() const;

    static std::unique_ptr<jam::TerminalTTY> createTTY();

    TITModel& model;
    const juce::Identifier tag;
    std::unique_ptr<jam::Model::ParameterAttachment> winsizeAttachment;
    jam::TerminalProcessor::Events events;
    std::string pendingInput;
    jam::TerminalFIFO input;
    jam::VirtualDevice device { jam::VirtualDeviceType::virtualDeviceTypeName, jam::VirtualDeviceType::virtualDeviceTypeName };
    std::unique_ptr<jam::TerminalTTY> tty { createTTY() };
    jam::TerminalProcessor terminalProcessor;
    jam::TerminalBuffer terminalBuffer;
    jam::TerminalSwap swap;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITRunProcessor)
};
```

   - The member order is the order of `EVETerminalProcessor.h:347-365`. `restart` is the template of `EVETerminalProcessor.h:67-77`, verbatim.
   - `tag` is the type of the child of the Model root that holds the rows of this chain: `winsize`, `shellExited`, and `numShellExits` (Step 9).

2. `TITRunProcessor.cpp`, the life of the chain. `#include "TITRunProcessor.h"`, then `static constexpr int virtualBlockSize { 65536 };` (`EVETerminalProcessor.cpp:5`), then:

```cpp
TITRunProcessor::TITRunProcessor (TITModel& modelToUse, const juce::Identifier& typeToUse)
    : model (modelToUse)
    , tag (jam::Format::toType (typeToUse))
{
    winsizeAttachment = std::make_unique<jam::Model::ParameterAttachment> (*model.getParameter<jam::Parameter<int64_t>> (tag, Id::winsize),
                                                                          [this] (const juce::var&) { restart(); });
    registerEvents();
    registerUserInterfaceGetters();
    registerChainEvents();
    tty->onBytesReady = [this] (int numSamples) { device.clock (numSamples); };
}

TITRunProcessor::~TITRunProcessor()
{
    releaseResources();
}

std::unique_ptr<jam::TerminalTTY> TITRunProcessor::createTTY()
{
#if JUCE_WINDOWS
    return std::make_unique<jam::TerminalWindowsTTY> (juce::File::getSpecialLocation (juce::File::currentExecutableFile).getParentDirectory());
#else
    return std::make_unique<jam::TerminalUnixTTY> (false);
#endif
}

void TITRunProcessor::audioDeviceAboutToStart (juce::AudioIODevice*)
{
    input.prepare (virtualBlockSize);
    terminalProcessor.prepare (virtualBlockSize);

    if (const auto winsize { getWinsize() }; winsize.isValid())
        prepare (winsize);
}

jam::TerminalWinsize TITRunProcessor::getWinsize() const
{
    return jam::TerminalWinsize::unpack (model.getParameter<jam::Parameter<int64_t>> (tag, Id::winsize)->getValue());
}

void TITRunProcessor::prepare (jam::TerminalWinsize winsize)
{
    handleUpdateNow();
    terminalProcessor.applyWinsize (winsize, terminalBuffer);
    handleUpdateNow();

    const auto blocksPerVblank { juce::jmax (1, static_cast<int> (std::ceil (device.getCurrentSampleRate() / virtualBlockSize / jam::CpuGraphicsEngine::getRefreshRateHz()))) };

    terminalBuffer.prepare (winsize.cols, winsize.rows, blocksPerVblank * virtualBlockSize + winsize.rows);
    swap.prepare (terminalBuffer.getNumBytes());
    terminalProcessor.reset (terminalBuffer, winsize);
    swap.setCapacity (jam::TerminalSwap::linesPerIndex);
    handleUpdateNow();

    if (tty->isOpen())
        tty->setWinsize (winsize);
}

void TITRunProcessor::process() noexcept
{
    input.writePendingData (*tty);

    while (terminalProcessor.process (*tty, terminalBuffer, swap, *jam::Colours::getInstance(), 0, events)) {}
}

void TITRunProcessor::handleAsyncUpdate()
{
    if (writePendingData())
        device.requestProcess();
}
```

   These functions take the bodies of EVE, verbatim: the public `prepare()` (`EVETerminalProcessor.cpp:365-385`, without the profiler zones), `start()` (`:387-391`), `releaseResources()` (`:393-397`), `audioDeviceIOCallbackWithContext` (`:399-402`), `restart()` (`:404-407`), `audioDeviceStopped()` (`:425`), `writePendingData()` (`:329-335`), `clearHistory()` (`:566-573`), `registerChainEvents()` (`:278-281`), `addInput` (`:283-287`), and `applyPendingInput()` (`:289-301`).

   - `createTTY` makes the POSIX TTY with no login wrapper (Step 1; ruling 59). On Windows the folder has no `conpty.dll`, so the TTY takes the ConPTY of the system (Step 1, the last note).
   - `prepare (winsize)` is the body of `EVETerminalProcessor.cpp:515-552` without its profiler zones and without `openTTY`: `run` opens the TTY (item 4). The history capacity is `jam::TerminalSwap::linesPerIndex`, the count of lines of one index file (`jam_terminal/document/jam_TerminalSwap.h:29`). The test of `isOpen()` names its threat: between two runs no session exists.
   - `process()` is the body of `EVETerminalProcessor.cpp:554-564` without the bypass: it writes the input lane to the TTY, then it decodes. The input lane carries the keys of the console editor to the tool, as in EVE (`EVETerminalProcessor.h:262-269,351-352`). The editor in `terminal` mode gives the bytes of each key to its input callback (inventory, console facts), so the chain must take them. The user can then answer a prompt of the tool in the console. The background argument is 0, the start value of EVE's `background` row (`eve/Source/layout/parameters.md:73`).
   - `handleAsyncUpdate()` moves the history rows of the buffer into the swap, on the serial context of the updater (`EVETerminalProcessor.cpp:337-345`; `jam_core/concurrency/jam_AsyncUpdater.h:16-36`).
   - `clearHistory()` erases the grid inside a restart. The next write of the swap then moves its first line behind the erased lines (`jam_TerminalSwap.h:231-235`), so the pane starts empty.
   - The winsize of the chain is the row `winsize` of its child. The view of the pane writes it (Step 14), the Model carries the tree write to the parameter, and the attachment calls `restart()`, as in EVE (`EVETerminalProcessor.cpp:68-72`). The chain then prepares again for the new grid.

3. `TITRunProcessor.cpp`, the events and the getter:

```cpp
void TITRunProcessor::registerEvents()
{
    events.add<int> (Id::shellExited, [this] (const juce::Identifier& key, int value)
    {
        const auto hasExited { value == 1 and model.getParameter<jam::Parameter<int>> (tag, key)->getValue() == 0 };

        model.storeValue (tag, key, value);

        if (hasExited)
            model.storeValue (tag, Id::numShellExits, model.getParameter<jam::Parameter<int>> (tag, Id::numShellExits)->getValue() + 1);
    });

    for (const auto& id : { Id::applicationCursor, Id::bracketedPaste, Id::mouseSgr, Id::cursorVisible, Id::alternateScreen, Id::isRecording, Id::syncOutput, Id::autoWrap, Id::originMode, Id::insertMode, Id::isWrapPending, Id::mouseTracking, Id::cursorShape, Id::keyboardFlags, Id::cursor, Id::numPrompts, Id::activeLink, Id::keyboardFlagsInactive, Id::numKeyboardModesNormal, Id::numKeyboardModesAlternate, Id::savedCursorNormal, Id::savedCursorAlternate })
        events.add<int> (id, [] (const juce::Identifier&, int) {});

    for (const auto& id : { Id::command, Id::preview, Id::title, Id::foregroundProcess, Id::cwd })
        events.add<const char*, int> (id, [] (const juce::Identifier&, const char*, int) {});

    for (const auto& id : { Id::gridStart, Id::removedGridStart, Id::firstLine, Id::clipboard, Id::previewBox, Id::outputBlock, Id::scrollingRegion, Id::pen, Id::savedPenNormal, Id::savedPenAlternate, Id::keyboardModesNormal, Id::keyboardModesAlternate, Id::syncStart })
        events.add<int64_t> (id, [] (const juce::Identifier&, int64_t) {});
}

void TITRunProcessor::registerUserInterfaceGetters()
{
    userInterfaceGetters[Id::textEditor].add<jam::TerminalSwap*&> (jam::Format::toType (Id::winsize), [this] (const juce::Identifier&, jam::TerminalSwap*& output)
    {
        triggerAsyncUpdate();
        output = &swap;
        applyPendingInput();
    });
}
```

   - The decoder fires `shellExited` at the end of each drained pass, with 1 when the read of the TTY gave -1: the child ended, or no session exists (`jam_terminal/processor/jam_TerminalProcessor.cpp:78-88`; `jam_terminal/tty/jam_TerminalTTY.h:116-125`). The setter stores the row, as EVE does (`EVETerminalProcessor.cpp:143-144`). At the change from 0 to 1 it raises `numShellExits`, in the `numCommands` form (`:157`). A Model row keeps only the last value before a flush (`EVETerminalProcessor.h:241`), so the count carries each exit, and the row gives the edge. The setter runs on the tick: it reads two atomics and stores two.
   - Each other event of the decoder gets an empty setter. `jam::Function::Map::get` throws `std::out_of_range` for a key that is absent (`jam_core/function_map/jam_Function.h:311`), and the decoder and the swap call each event of EVE's lists (`EVETerminalProcessor.cpp:143-193`). TIT has no row for them.
   - The getter is the text editor getter of EVE without the flush (`EVETerminalProcessor.cpp:256-273`): `TITModel` keeps its flush timer (Part 7 Step 2). `jam::TextEditor::bindToProcessor` reads it at each vblank (`jam_gui/widgets/jam_TextEditor.cpp:82-100`). The trigger moves the new history rows into the swap.

4. `TITRunProcessor.cpp`, one session:

```cpp
void TITRunProcessor::run (const juce::String& commandLine)
{
    restart ([this, &commandLine]
    {
        const auto tokens (juce::StringArray::fromTokens (commandLine, true));
        const auto vblankMilliseconds { juce::RelativeTime::seconds (1.0 / jam::CpuGraphicsEngine::getRefreshRateHz()).inSeconds() * 1000.0 };

        tty->close();
        tty->open (device.getContext(), device.getCurrentBufferSizeSamples(), vblankMilliseconds, getWinsize().toCellRect(),
                   tokens[0].unquoted(), tokens.joinIntoString (juce::String::charToString (Chars::space), 1), model.state[Id::directory].toString());
        model.storeValue (tag, Id::shellExited, 0);
    });
}

int TITRunProcessor::close()
{
    int exitCode { EXIT_FAILURE };

    restart ([this, &exitCode] { exitCode = tty->close(); });

    return exitCode;
}
```

   - The message thread owns the life of the chain, as in EVE: each open and each close of the TTY is inside `restart`, while the clock is stopped (`EVETerminalProcessor.h:60-77`; `.cpp:404-423,515-552`). So the tick never reads a TTY that another thread opens or closes, and `close()` never runs inside the tick (Step 1).
   - `run` ends the session that exists, and starts one session for the command line, in the folder of the repository. The program is the first token and the other tokens are its arguments, in the form of `EVETerminalProcessor.cpp:451-453`. The TTY splits the argument text at spaces and keeps a quoted group as one argument (`jam_terminal/tty/native/jam_TerminalTTY_posix.cpp:253-256`). The store of `shellExited` 0 comes after the open, with the clock stopped, so the next change to 1 is the end of this session.
   - A program that does not open gives no session. The first tick reads -1, the row goes from 0 to 1, and the `close()` of that exit gives `EXIT_FAILURE` (Step 1). So a failed open is a failed line, with no branch of its own.
   - `close()` ends the session and returns the exit status of the child (Step 1).

5. `TITGitProcessor`, the header:
   - Add `#include "TITRunProcessor.h"`.
   - Public section, after `getPushLine` of Part 9 Step 5: `static juce::String getForcePushLine (const juce::ValueTree& git);` and `static juce::StringArray getReplaceLocalLines (const juce::ValueTree& git);`. After the public `run`: `void cancel();`, `void applyShellExit();`, `void setCommitMessage (const juce::String& message);`, and `void prepare();`. Then the public member `TITRunProcessor console;`.
   - Private section. Delete first: the private `run` of Part 9 Step 5 goes. Add `void applyPendingCommandLine();`.
   - Members, after `directoryAttachment`: `juce::StringArray pendingCommandLines;` and `juce::StringArray pendingFallbackCommandLines;`.
   - `console` is declared after `model`, and the constructor initializes it with `console (modelToUse, Id::console)`.

6. `TITGitProcessor`, the source. Replace the public `run` of Part 9 Step 5 with these bodies:

```cpp
void TITGitProcessor::prepare()
{
    console.prepare();
}

void TITGitProcessor::run (const juce::StringArray& commandLines, const juce::StringArray& fallbackCommandLines)
{
    pendingCommandLines = commandLines;
    pendingFallbackCommandLines = fallbackCommandLines;
    model.storeValue (titTag, Id::isRunning, 1);
    model.flush();
    console.clearHistory();
    applyPendingCommandLine();
}

void TITGitProcessor::applyShellExit()
{
    if (console.close() != 0)
        pendingCommandLines = std::exchange (pendingFallbackCommandLines, {});

    applyPendingCommandLine();
}

void TITGitProcessor::applyPendingCommandLine()
{
    if (pendingCommandLines.isEmpty())
    {
        model.storeValue (titTag, Id::isRunning, 0);
        addJob();
    }
    else
    {
        const auto commandLine { pendingCommandLines[0] };

        pendingCommandLines.remove (0);
        model.storeValue (titTag, Id::message, commandLine.toRawUTF8(), static_cast<int> (commandLine.getNumBytesAsUTF8()));
        console.run (commandLine);
    }
}

void TITGitProcessor::cancel()
{
    pendingCommandLines.clear();
    pendingFallbackCommandLines.clear();
    console.close();
    model.storeValue (titTag, Id::message, text::Footer::aborted, static_cast<int> (juce::String { text::Footer::aborted }.getNumBytesAsUTF8()));
}

void TITGitProcessor::setCommitMessage (const juce::String& message)
{
    juce::File { model.state[Id::directory].toString() }.getChildFile (files::gitDirectory).getChildFile (files::commitMessage).replaceWithText (message);
}

juce::String TITGitProcessor::getForcePushLine (const juce::ValueTree& git)
{
    return getPushLine (git) + juce::String::charToString (Chars::space) + text::Git::force;
}

juce::StringArray TITGitProcessor::getReplaceLocalLines (const juce::ValueTree& git)
{
    return { text::Git::fetchOrigin, juce::String { text::Git::resetHardOrigin } + git.getProperty (Id::currentBranch).toString(), text::Git::clean };
}
```

- The view tells `run`, `cancel`, and `applyShellExit`, on the message thread.
- A run has no thread of its own and no wait. One line is one session of the console chain. The end of the session is an event: the decoder stores the rows on the tick, the flush of the Model writes them to the tree, the event of `numShellExits` in `TITView` tells `applyShellExit` (Step 16), and the processor starts the next line. This is the path of EVE's `numCommands` row: the view calls the processor when the count changes (`EVETerminalProcessor.h:107-126`; `eve/Source/EVEView.cpp:138-144`).
- `applyShellExit` closes the session and takes its exit status. A status other than 0 replaces the pending lines with the fallback lines, one time: a failure of a fallback line finds no more fallback and ends the run. This is the rule of Part 9 Step 5.
- `applyPendingCommandLine` starts the next line, or ends the run: it clears `isRunning` and adds the detection job of Part 8.
- `cancel` drops the pending lines and closes the session, which ends the child (Step 1). The next tick reads -1, and the exit event ends the run through the same path. No line follows, because no line is pending.
- The pending lines are members of the processor, in the form of EVE's `pendingInput` (`EVETerminalProcessor.h:351`). Only the message thread reads and writes them.
- The console shows the bytes of the tool: the swap gets its lines only from the decoder, and the decoder reads only a TTY (`jam_TerminalProcessor.cpp:66,122-133`). So the command line of the line that runs, and the text `Operation aborted by user`, go to the footer through the `message` row, with the text form of `storeValue` (`EVETerminalProcessor.cpp:161-162`). Legacy writes both into the console (`legacy/git/execute.go:101-102`).
- A run starts with an empty console: `clearHistory()`.
- `setCommitMessage` writes the message file of the repository. The command line `text::Git::commitFile` of Step 5 reads it (Closed by a read, item 13). The file work is in the processor, and the view only gives the text (Part 9, Overview, item 6).
- `getForcePushLine` and `getReplaceLocalLines` make command lines from the `git` rows, as `getPushLine` does. They are static members of the processor for the same reason.

**Validation:**
- No bail-out guard. No function body is longer than 30 lines.
- No processor writes a tree or sends a byte to a view: no `setProperty`, no `sendPropertyChangeMessage`, no `juce::MessageManager::callAsync`, no `jam::Subprocess`, and no lane of output bytes (rulings 57 and 58). Each one writes the Model only with `storeValue`.
- The tick writes the input lane to the TTY, stores atomics, and decodes. It does not allocate, lock, wait, or make a blocking call (ruling 52; MANIFESTO, Deadline threads).
- No view makes a command line or writes a file.
- New names and their families: `TITRunProcessor` (`EVETerminalProcessor`, `TITTerminalProcessor`; the verb `run` of Part 9 Step 5); its members and functions are EVE's, verbatim (`EVETerminalProcessor.h:30-369`); `run`, `close` (`run`, Part 9; `jam::TerminalTTY::close`); `tag`, `typeToUse` (`terminalTag`, `EVETerminalProcessor.h:346`; `jam::Format::toType`); `applyShellExit` (`applyCommands`, `applyGridEvents`, `EVETerminalProcessor.h:118,126`; the row `shellExited`); `applyPendingCommandLine`, `pendingCommandLines`, `pendingFallbackCommandLines` (`applyPendingInput`, `pendingInput`, `EVETerminalProcessor.h:269,351`; `commandLines`, `fallbackCommandLines`, Part 9 Step 5); `numShellExits` (`numCommands`, `numGridEvents`, `eve/Source/layout/parameters.md:45,47`); `cancel` (the command of Step 5); `console` (the row and the mode `console`); `hasExited`, `tokens`, `vblankMilliseconds`, `blocksPerVblank`, `exitCode`, `commandLine` (`EVETerminalProcessor.cpp:451,521,549`; Part 8 Step 7); `setCommitMessage` (NAMES.md Rule 1, `set`: it replaces the text of the file `files::commitMessage`); `getForcePushLine`, `getReplaceLocalLines` (`getPushLine`, Part 9 Step 5; the commands `forcePush` and `replaceLocal`).

### Step 11: `HeaderView`
**Scope:** `tit/Source/HeaderView.h`, `tit/Source/HeaderView.cpp` (both new).
**Action:**
1. The header: `#pragma once`, `#include <JuceHeader.h>`, `#include "generated/Generated.h"`. Declare:

```cpp
class HeaderView : public juce::Component,
                   public jam::Model::Component<HeaderView>,
                   private juce::ValueTree::Listener
{
public:
    enum ColourIds
    {
        backgroundColourId = map::ColourId::headerViewBackgroundColourId,
        textColourId = map::ColourId::headerViewTextColourId,
        labelColourId = map::ColourId::headerViewLabelColourId,
        dimmedColourId = map::ColourId::headerViewDimmedColourId,
        accentColourId = map::ColourId::headerViewAccentColourId,
        separatorColourId = map::ColourId::headerViewSeparatorColourId,
        cleanColourId = map::ColourId::headerViewCleanColourId,
        dirtyColourId = map::ColourId::headerViewDirtyColourId,
        inSyncColourId = map::ColourId::headerViewInSyncColourId,
        aheadColourId = map::ColourId::headerViewAheadColourId,
        behindColourId = map::ColourId::headerViewBehindColourId,
        readyColourId = map::ColourId::headerViewReadyColourId,
        notRepoColourId = map::ColourId::headerViewNotRepoColourId,
        conflictedColourId = map::ColourId::headerViewConflictedColourId,
        mergingColourId = map::ColourId::headerViewMergingColourId,
        rebasingColourId = map::ColourId::headerViewRebasingColourId,
        dirtyOperationColourId = map::ColourId::headerViewDirtyOperationColourId,
        timeTravelColourId = map::ColourId::headerViewTimeTravelColourId,
        warningColourId = map::ColourId::headerViewWarningColourId,
    };

    explicit HeaderView (juce::ValueTree stateToUse);
    ~HeaderView() override = default;

    void paint (juce::Graphics& g) override;

    static constexpr int numRows { 7 };

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    void paintText (juce::Graphics& g, int row, const juce::String& text, int colourId, juce::Justification justification, bool isBold) const;

    static juce::String getOperationText (const juce::ValueTree& git);
    static int getOperationColourId (const juce::ValueTree& git);
    static juce::String getBranchText (const juce::ValueTree& git);
    static juce::String getRemoteText (const juce::ValueTree& git);
    static juce::String getWorkingTreeText (const juce::ValueTree& git);
    static juce::String getTimelineText (const juce::ValueTree& git);
    static juce::String getTimelineDescription (const juce::ValueTree& git);
    static int getTimelineColourId (const juce::ValueTree& git);

    static constexpr int descriptionIndent { 3 };

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (HeaderView)
};
```

2. The constructor: `: jam::Model::Component<HeaderView> (stateToUse)`, then the three statements of `TextEditorStatusBar::TextEditorStatusBar()` without the `terminal` line (`jam_gui/widgets/jam_TextEditorStatusBar.cpp:5-11`). The state is the root of the Model.
3. `valueTreePropertyChanged`: `if (tree.hasType (jam::Format::toType (Id::git)) or property == Id::directory) repaint();`.
4. `paintText`: the row area is `getLocalBounds().withY (row).withHeight (1)`: one unit is one cell (`ling/Source/style.css:6-9`). It sets the colour with `findColour (colourId)`, the font with `style.getMonoBoldFont()` or `style.getMonoFont()` in the form of `jam_TextEditorStatusBar.cpp:56,60`, and calls `g.drawText (text, area, justification)`.
5. `paint`: the assert of `jam_TextEditorStatusBar.cpp:15`, then `g.fillAll (findColour (backgroundColourId));`. When the `operation` row is not `map::Operation::notRepo`, it paints these rows, with `git` as the `git` child of `state`. Words join with one `Chars::space`.

   | Row | Left (`juce::Justification::centredLeft`) | Right (`juce::Justification::centredRight`) |
   |---|---|---|
   | 0 | `text::Header::directoryMark`, the `directory` row; `labelColourId`; bold | `getOperationText (git)`; `getOperationColourId (git)`; bold |
   | 1 | `getRemoteText (git)`; `accentColourId` with a remote address, else `dimmedColourId` | `getBranchText (git)`; `warningColourId` when `isTitTimeTravel`, else `accentColourId`; bold |
   | 2 | `text::Header::separator` repeated for the width; `separatorColourId` | |
   | 3 | `getWorkingTreeText (git)`; `dirtyColourId` when `dirty`, else `cleanColourId`; bold | the LFS text, two spaces, `ProjectInfo::versionString`; `warningColourId` when LFS is not ready, else `dimmedColourId` |
   | 4 | `descriptionIndent` spaces, then `dirtyDescription` or `cleanDescription`; `textColourId` | |
   | 5 | `getTimelineText (git)`; `getTimelineColourId (git)`; bold | |
   | 6 | `descriptionIndent` spaces, then `getTimelineDescription (git)`; `textColourId` | |

6. The helper functions. Each one is a lookup in a function-local `static constexpr std::array`, indexed by the key of the row value, in the form of `menuModes` (Part 9 Step 7, item 3). The entries, in key order:

   | Function | Table and rule |
   |---|---|
   | `getOperationText` | by `map::Operation` (0 to 8): `notRepo`, `ready`, `merging`, `conflicted`, `rebasing`, `timeTravel`, `dirtyOperation`, `ready`, `pinnedAt` and the `currentHash` row. When `isDetached` and not `isTitTimeTravel` and the operation is `timeTraveling`: `detached`. |
   | `getOperationColourId` | by `map::Operation`: `notRepoColourId`, `readyColourId`, `mergingColourId`, `conflictedColourId`, `rebasingColourId`, `timeTravelColourId`, `dirtyOperationColourId`, `readyColourId`, `timeTravelColourId`. |
   | `getBranchText` | `isDetached` and not `isTitTimeTravel`: the `currentHash` row. Else `branchMark` and the `currentBranch` row, or `notAvailable` when that row is empty. |
   | `getRemoteText` | the `remoteUrl` row is empty: `noRemote`. Else `remoteMark` and the row. |
   | `getWorkingTreeText` | `clean`; or `dirty`, and when `modifiedCount` is above 0, `modifiedMark` and the count. |
   | `getTimelineText` | the operation is `timeTraveling`: `detachedAt` and the `currentHash` row. Else by `map::Timeline` (0 to 4): `timelineNone`, `inSync`, `ahead` with `aheadMark` and `commitsAhead`, `behind` with `behindMark` and `commitsBehind`, `diverged` with both marks and both counts. A mark and its count show only when the count is above 0. |
   | `getTimelineDescription` | by `map::Timeline`: `noRemoteDescription`; `inSyncDescription`; `commitsAhead` and `aheadDescription`; `commitsBehind` and `behindDescription`; `commitsAhead`, `divergedAheadMark`, a space, `commitsBehind`, `divergedBehindMark`. The operation is `timeTraveling`: the empty text (Closed by a read, item 19). |
   | `getTimelineColourId` | the operation is `timeTraveling`: `warningColourId`. Else by `map::Timeline`: `dimmedColourId`, `inSyncColourId`, `aheadColourId`, `behindColourId`, `behindColourId`. |

   The LFS text is `text::Header::lfs` when `isUsingLfs` and `isLfsReady`, `lfsNotReady` when `isUsingLfs` only, and the empty text otherwise.

**Validation:**
- Each row and each state text equals legacy (`legacy/ui/header.go:59-197`; `legacy/app/app_view_header.go:24-124`; `legacy/app/state_info.go:19-131`). The `pinned` row is Part 2, "Other text". `Diverged` takes the colour of `Behind` (`state_info.go:66`).
- The view has no member that holds a value. Each paint reads the Model.
- No function has more than three branches. No `==` on strings.
- Local names of the bodies: `git`, `style`, `area`, `operation`, `timeline`, `texts`, `colourIds`, `count`.
- New names and their families: `HeaderView` (Closed by a read, item 2); `stateToUse` (`modelToUse`, Part 7 Step 4); `numRows` (`getNumRows`, `jam_gui/widgets/jam_TextEditor.h:287`; legacy `HeaderHeight`); `paintText` (`paintPage`, `jam_TextEditor.h:289`); the `get…Text`, `get…Description`, and `get…ColourId` functions (the legacy `Label`, `Description`, and `Color` fields of `state_info.go`); `descriptionIndent` (legacy "Indent (3 spaces)", `legacy/ui/header.go:145`).

### Step 12: JAM, the keys `j` and `k` of `jam::PopupMenu::MenuWindow` (rulings 60, 61, 63)
**Scope:** `jam/jam_gui/menu/jam_PopupMenu.h`.
**Action:** TIT has no `MenuView`, no `ListView`, and JAM gets no menu component. Each menu of TIT is the popup menu of JAM, `jam::PopupMenu::MenuWindow`, shown inside the view. It makes no window (ruling 61): with a parent component it is a child component.

| Need | JAM and JUCE give it | Where |
|---|---|---|
| The data of a menu | A `juce::PopupMenu` of `juce::PopupMenu::Item`: `text`, `itemID`, `commandManager`, `shortcutKeyDescription`, `isEnabled`, `isSeparator` | JUCE `juce_gui_basics/menus/juce_PopupMenu.h:139-189`; the builders of `jam_gui/menu/jam_Menu.h:17-98` |
| The show call | `jam::showAsync (juce::PopupMenu& menu, const juce::PopupMenu::Options& options, std::function<void (int)> callback)` | `jam_gui/menu/jam_PopupMenu.h:1962-1986` |
| The menu as a child component, with no desktop window | `if (auto* pc = options.getParentComponent()) pc->addChildComponent (this);`. `addToDesktop` is in the other branch only | `jam_PopupMenu.h:657-677` |
| The parent can also come from the LookAndFeel | `opts.withParentComponent (… .getParentComponentForMenuOptions (opts))` | `jam_PopupMenu.h:640` |
| The first highlighted item | `options.getInitiallySelectedItemId()` | `jam_PopupMenu.h:741-754` |
| Down and up move the highlight; return and space run the item; Escape closes the menu | `MenuWindow::keyPressed` | `jam_PopupMenu.h:956-1013` |
| The highlight wraps and stops on an item that can run | `selectNextItem` | `jam_PopupMenu.h:1680-1714` |
| A run item invokes its command | `managerOfChosenCommand->invoke (info, true)` with `InvocationInfo::fromMenu`, from the `commandManager` and the `itemID` of the item | `jam_PopupMenu.h:870-879,1852-1860` |
| The keys reach the menu | The menu is modal. The peer gives a key to the modal component, then to each parent | `jam_PopupMenu.h:1933-1934`; JUCE `juce_gui_basics/windows/juce_ComponentPeer.cpp:175-222` |
| A key that the menu does not take | `keyPressed` returns `false`, so the key goes to the parent, the view, and to its key sheet | `jam_PopupMenu.h:1007-1010`; `juce_ComponentPeer.cpp:199-220` |
| The paint | The LookAndFeel: `drawPopupMenuBackgroundWithOptions`, `drawPopupMenuItemWithOptions`, `getIdealPopupMenuItemSizeWithOptions`, `getPopupMenuFont` | `jam_PopupMenu.h:198,208,807`; JUCE `juce_PopupMenu.h:941-1053` |
| The open menus | `MenuWindow::getActiveWindows()`, and `dismissMenu (nullptr)` closes one | `jam_PopupMenu.h:914-933`; doxygen `structjam_1_1PopupMenu_1_1MenuWindow`, `getActiveWindows` |

The one thing that `MenuWindow` does not have is ruling 63: `j` moves as down and `k` moves as up.

1. `jam_PopupMenu.h`, `MenuWindow::keyPressed` (`:956-965`). Add two constants at the start of the body, and one test to each of the first two conditions:

```cpp
    bool keyPressed (const juce::KeyPress& key) override
    {
        static constexpr juce::juce_wchar nextItemCharacter { 'j' };
        static constexpr juce::juce_wchar previousItemCharacter { 'k' };

        if (key.isKeyCode (juce::KeyPress::downKey) or key.getTextCharacter() == nextItemCharacter)
        {
            selectNextItem (MenuSelectionDirection::forwards);
        }
        else if (key.isKeyCode (juce::KeyPress::upKey) or key.getTextCharacter() == previousItemCharacter)
        {
            selectNextItem (MenuSelectionDirection::backwards);
        }
```

   The other branches of the function stay as they are.
   - The test of a letter is on the text character: a typed `j` has the text character `j`, `J` has `J`, and a named key has the text character 0 (Part 4 Step 2, `getTextKeyPress`; Part 4, inventory, "How `jam::CommandManager` matches a key").
2. This step adds no component, no LookAndFeel method, and no size. The paint code of a LookAndFeel is ordinary JUCE paint code, and `jam::TerminalGraphicsContext` puts it on cells: it "lets ordinary JUCE painting code (Graphics::fillRect, Graphics::drawText, Graphics::drawImage, Path fills) target a terminal cell grid" (`jam_terminal/graphics/jam_TerminalGraphicsContext.h:5-8`).
   - The sizes come from the style sheet. In a terminal sheet "One unit is one terminal cell. The `fontSize: 1em` property sets the base font size to one cell", so "every em metric below is a count of cells" (`ling/Source/style.css:6-9`).
   - The colours are the popup menu ids of JUCE. JAM's `## ColourId` holds their rows (`jam/cast/style.md:120-124`), and the sheet of TIT sets them (Step 8).

**Validation:**
- No new class, no new function, and no new member in JAM. The change is two tests in one existing function (ruling 63).
- A GUI menu of JAM gets the same two keys. A GUI menu has no text input, so no typed `j` or `k` is lost.
- `keyPressed` keeps its branches. The two constants are in the function that uses them (CODING.md, "Narrowest scope that holds every use").
- New names and their families: `nextItemCharacter`, `previousItemCharacter` (`selectNextItem`, `jam_PopupMenu.h:1680`; `getTextCharacter`, JUCE).

### Step 13: `FooterView`
**Scope:** `tit/Source/FooterView.h`, `tit/Source/FooterView.cpp` (both new).
**Action:**
1. The header, with the includes of Step 11:

```cpp
class FooterView : public juce::Component,
                   public jam::Model::Component<FooterView>,
                   private juce::ValueTree::Listener
{
public:
    enum ColourIds
    {
        backgroundColourId = map::ColourId::footerViewBackgroundColourId,
        textColourId = map::ColourId::footerViewTextColourId,
    };

    explicit FooterView (juce::ValueTree stateToUse);
    ~FooterView() override = default;

    void paint (juce::Graphics& g) override;

    static constexpr int numRows { 1 };

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    juce::String getText() const;
    juce::String getHint() const;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (FooterView)
};
```

2. The constructor: the form of Step 11, item 2. The state is the root of the Model.
3. `valueTreePropertyChanged`: the form of `jam_TextEditorStatusBar.cpp:64-70`, with `paintedProperties { Id::isQuitConfirmActive, Id::isClearConfirmActive, Id::message, Id::mode, Id::isRunning }`.
4. `getText()` and `getHint()`:

```cpp
juce::String FooterView::getText() const
{
    if (static_cast<bool> (state.getProperty (Id::isQuitConfirmActive))) return text::Footer::quitConfirm;
    if (static_cast<bool> (state.getProperty (Id::isClearConfirmActive))) return text::Footer::clearConfirm;
    if (state.getProperty (Id::message).toString().isNotEmpty()) return state.getProperty (Id::message).toString();
    return getHint();
}

juce::String FooterView::getHint() const
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    if (mode == map::EditorMode::console) return static_cast<bool> (state.getProperty (Id::isRunning)) ? text::Footer::consoleRunning : text::Footer::consoleComplete;
    if (mode == map::EditorMode::confirmation) return text::Footer::confirmation;
    if (mode == map::EditorMode::input) return text::Footer::input;
    return juce::String {};
}
```

   - A menu mode has no footer hint. The legacy hint is the hint of the highlighted item, and `jam::PopupMenu::MenuWindow` gives its highlighted item to no caller: its only results are the id of the run item and the command that it invokes (`jam_gui/menu/jam_PopupMenu.h:870-892,1852-1860,1962-1986`). So the footer shows the empty text in a menu mode, and `Source/menu.md` has no `hint` column (Step 7).

5. `paint`: the assert and the `fillAll` of Step 11, then `g.setColour (findColour (textColourId));`, the font `style.getMonoFont()`, and `g.drawText (getText(), getLocalBounds(), juce::Justification::centred);`.

**Validation:**
- The priority equals legacy (`legacy/app/footer.go:14-42`), with the message that legacy never showed (legacy fault 1).
- Each return is a result return. No function has more than three branches.
- In a menu mode the footer shows the empty text.
- The footer reads the Model rows that it paints, and holds no copy of one (MANIFESTO E, "A view object that positions or paints itself from the Model's positional and appearance value-data is the materialisation of that data").
- New names and their families: `FooterView` (Closed by a read, item 2); `getText`, `getHint` (legacy `GetFooterContent`, `Hint`).

### Step 14: JAM, `jam::TerminalMessageBox` (rulings 65, 66); TIT, `ConsoleView`, `InputView`
**Scope:** `jam/jam_gui/windows/jam_TerminalMessageBox.h`, `jam/jam_gui/windows/jam_TerminalMessageBox.cpp` (both new); `jam/jam_gui/jam_gui.h`, `jam/jam_gui/jam_gui.cpp`; `tit/Source/ConsoleView.h`, `tit/Source/ConsoleView.cpp`, `tit/Source/InputView.h`, `tit/Source/InputView.cpp` (all new).
**Action:** TIT has no `ConfirmationView`, and JAM gets no dialog component. The dialog is the `juce::AlertWindow` of JUCE, shown inside a parent component. JUCE has each part of the legacy dialog:

| Need | JUCE gives it | Where |
|---|---|---|
| A box from a title, a message, and button labels, with a result callback | `juce::AlertWindow::showScopedAsync (const MessageBoxOptions&, std::function<void (int)>)` | JUCE `juce_gui_basics/windows/juce_AlertWindow.cpp:698-704` |
| The box as a child component, with no native window | `MessageBoxOptions::withParentComponent`: "The component that will contain the message box … This will only affect JUCE AlertWindows". The helper calls `parent->addAndMakeVisible (alert)` | `juce_gui_basics/windows/juce_MessageBoxOptions.h:99-104`; `juce_gui_basics/detail/juce_AlertWindowHelpers.h:101-107` |
| The LookAndFeel builds the box | `lf.createAlertWindow (title, message, button1, button2, button3, iconType, numButtons, associatedComponent)`, with the LookAndFeel of the associated component | `juce_AlertWindowHelpers.h:80-92` |
| Each button is a `juce::TextButton` that wants the keyboard focus (ruling 66) | `AlertWindow::addButton` | `juce_AlertWindow.cpp:125-136` |
| Enter runs the first button; Escape runs the second | `aw->addButton (button1, 1, KeyPress (KeyPress::returnKey), button1ShortCut); aw->addButton (button2, 0, KeyPress (KeyPress::escapeKey), button2ShortCut);` | `juce_gui_basics/lookandfeel/juce_LookAndFeel_V2.cpp:417-418` |
| The legacy keys `y` and `n` | The second shortcut of each button is the first letter of its label, in lower case | `juce_LookAndFeel_V2.cpp:410-413` |
| The message wraps in the box | `drawAlertBox (g, alert, textArea, textLayout)` draws a `juce::TextLayout` | `juce_LookAndFeel_V2.cpp:432-433,489-492` |
| The outline of the box | `g.drawRect (0, 0, alert.getWidth(), alert.getHeight())`. On cells a fill thinner than one cell is a run of box-drawing glyphs | `juce_LookAndFeel_V2.cpp:494-495`; `jam_terminal/graphics/jam_TerminalGraphicsContext.h:29-33` |
| No icon | `drawAlertBox` paints the icon only when the type is not `MessageBoxIconType::NoIcon` | `juce_LookAndFeel_V2.cpp:448` |
| The life of the box | `juce::ScopedMessageBox`: the owner holds it, and the box closes when the object goes. JAM uses it: "The message box. It stays open no longer than the returned object lives." | `jam_gui/windows/jam_MessageBox.h:139-149`; `jam_MessageBox.cpp:145-158` |
| The result | The return value of the pressed button: 1 for the first button, 0 for the second | `juce_LookAndFeel_V2.cpp:417-418`; `juce_AlertWindowHelpers.h:49-55` |

So `jam::TerminalMessageBox` is `jam::MessageBox` with one difference: it calls `juce::AlertWindow::showScopedAsync` with a parent component, where `jam::MessageBox` calls `juce::NativeMessageBox::showAsync` (`jam_MessageBox.cpp:18,32`).

1. `jam_gui.h`: after `#include "windows/jam_MessageBox.h"` (`:103`), add `#include "windows/jam_TerminalMessageBox.h"`. `jam_gui.cpp`: after `#include "windows/jam_MessageBox.cpp"` (`:27`), add `#include "windows/jam_TerminalMessageBox.cpp"`.
2. `jam_TerminalMessageBox.h`:

```cpp
namespace jam
{
/*____________________________________________________________________________*/

struct TerminalMessageBox
{
    [[nodiscard]] static juce::ScopedMessageBox showAlert (const juce::String& title,
                                                           const juce::String& message,
                                                           const juce::String& buttonText,
                                                           juce::Component* associatedComponent,
                                                           std::function<void (int)> callback);

    [[nodiscard]] static juce::ScopedMessageBox showYesNoBox (const juce::String& title,
                                                              const juce::String& message,
                                                              const juce::String& button1Text,
                                                              const juce::String& button2Text,
                                                              juce::Component* associatedComponent,
                                                              std::function<void (int)> callback);
};

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jam
```

3. `jam_TerminalMessageBox.cpp`:

```cpp
namespace jam
{
/*____________________________________________________________________________*/

juce::ScopedMessageBox TerminalMessageBox::showAlert (const juce::String& title,
                                                      const juce::String& message,
                                                      const juce::String& buttonText,
                                                      juce::Component* associatedComponent,
                                                      std::function<void (int)> callback)
{
    return juce::AlertWindow::showScopedAsync (juce::MessageBoxOptions::makeOptionsOk (juce::MessageBoxIconType::NoIcon, title, message, buttonText, associatedComponent)
                                                   .withParentComponent (associatedComponent),
                                               std::move (callback));
}

juce::ScopedMessageBox TerminalMessageBox::showYesNoBox (const juce::String& title,
                                                         const juce::String& message,
                                                         const juce::String& button1Text,
                                                         const juce::String& button2Text,
                                                         juce::Component* associatedComponent,
                                                         std::function<void (int)> callback)
{
    return juce::AlertWindow::showScopedAsync (juce::MessageBoxOptions::makeOptionsYesNo (juce::MessageBoxIconType::NoIcon, title, message, button1Text, button2Text, associatedComponent)
                                                   .withParentComponent (associatedComponent),
                                               std::move (callback));
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jam
```

   - The struct is the form of `jam::MessageBox`: static show functions (`jam_MessageBox.h:15-43`). The function returns the `juce::ScopedMessageBox`, the form of `MessageBox::licenseNotFound` (`jam_MessageBox.h:147-149`).
   - `makeOptionsYesNo (iconType, title, message, button1Text, button2Text, associatedComponent)` is the function of JUCE for a box with two buttons (`juce_MessageBoxOptions.h:170-179`). The two label parameters have its names.
   - `showAlert` is the box with one button, the form of `jam::MessageBox::showAlert` (`jam_MessageBox.h:27-31`), with `makeOptionsOk (iconType, title, message, buttonText, associatedComponent)` (`juce_MessageBoxOptions.h:149-157`). Enter and Escape both run its button, and the result is 0 (`juce_LookAndFeel_V2.cpp:402-407`). Part 13 uses it for the setup wizard.
   - The associated component is also the parent component. So the box is a child of the view that opens it, and it takes the LookAndFeel of that view (`juce_AlertWindowHelpers.h:80-83,101-107`).
   - No state for the widget (ruling 65). The box holds its buttons, their keys, and its focus. No Model row and no member of JAM or of TIT holds a selected button.
   - `juce::AlertWindow::showScopedAsync` shows a native box when the default LookAndFeel answers `true` to `isUsingNativeAlertWindows()` (`juce_AlertWindow.cpp:698-704`).
4. The owner of a dialog, in TIT (Step 16): it holds the returned `juce::ScopedMessageBox` as a member, and it gives each label as legacy has it.

7. `ConsoleView`, in `ConsoleView.h` and `ConsoleView.cpp`, with the includes of Step 11. Each screen of TIT is its own view class with its own row, as each pane of END is (`end/Source/ui/EditorView.h:25-28`):

```cpp
class ConsoleView : public juce::Component,
                    public jam::Model::Component<ConsoleView>
{
public:
    ConsoleView (juce::ValueTree stateToUse, jam::Registry::Getters& getters, jam::Registry::Events& events);
    ~ConsoleView() override = default;

    void resized() override;
    void focusGained (FocusChangeType cause) override;

    void setMode (int mode);

private:
    jam::TextEditor editor;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConsoleView)
};
```

   - The constructor: the form of Step 11, item 2, without the listener. Then `setWantsKeyboardFocus (true);`, `addAndMakeVisible (editor);`, `editor.bindToProcessor (getters, Id::textEditor, jam::Format::toType (Id::winsize));`, and `editor.bindToEvents (events[Id::terminalInput]);`. The state is the child of the Model that holds the rows of the chain: the `console` child for the console.
   - `setMode`: `editor.setDefaultMode (mode);` then `editor.setMode (mode);`, the pair of `InputView::setDocument` (item 8; Step 3).
   - `resized()`:

```cpp
void ConsoleView::resized()
{
    editor.setBounds (getLocalBounds());
    state.setProperty (Id::winsize, static_cast<juce::int64> (jam::TerminalWinsize { getWidth(), getHeight(), 0, 0 }.pack()), nullptr);
}
```

   - `focusGained`: `editor.grabKeyboardFocus();`. The view gives its focus to its editor, as EVE's view gives its focus to its default component (`eve/Source/EVEView.cpp:74-79`).
   - The editor reads the swap of the run chain (ruling 58): `bindToProcessor` sets the source of the editor to the getter of the processor and pulls at each vblank of the peer (`jam_gui/widgets/jam_TextEditor.h:134-144`; `jam_TextEditor.cpp:82-100`; Step 10, item 3). The keys of the getter are EVE's: the map `Id::textEditor` and the getter of the `winsize` type (`eve/Source/EVETerminalProcessor.cpp:258`). The view holds no text, makes no document, and has no listener.
   - The owner tells the view its mode. In `terminal` mode the view follows the end of the output, and a key goes to the key sheet first and then, as bytes, to the tool. In `reader` mode the view stays where the user put it, and a key goes to the commands of the editor, to the viewport, and then to the parent (inventory, console facts). The console is in `terminal` mode while a run runs, and in `reader` mode after it (Step 16). A pane of Parts 11 and 12 is in `reader` mode.
   - `bindToEvents` gives the bytes of a key to the `write` event of the chain, in the form of the window of Part 7 Step 7 (`window->bindToEvents (processor->chainEvents[Id::terminalInput])`).
   - The view writes the size of its grid to the `winsize` row of its state. One unit is one cell (Step 11, item 4), so the width and the height are the columns and the rows. The fields are those of `jam::TerminalWinsize`, and the row holds its packed form (`jam_terminal/video/jam_TerminalWinsize.h:34-48`). The chain reacts to the row (Step 10, item 2).

8. `InputView`, in `InputView.h` and `InputView.cpp`, with the includes of Step 11:

```cpp
class InputView : public juce::Component
{
public:
    InputView();
    ~InputView() override = default;

    void resized() override;
    void focusGained (FocusChangeType cause) override;

    void setDocument();
    juce::String getText() const;

private:
    jam::TextEditor editor;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (InputView)
};
```

   - The constructor: `setWantsKeyboardFocus (true);` and `addAndMakeVisible (editor);`. The view has no row of its own, so it has no Model base and no listener.
   - `resized()` and `focusGained`: the two bodies of `ConsoleView`.
   - `setDocument()` and `getText()`:

```cpp
void InputView::setDocument()
{
    const std::string text { static_cast<char> (Chars::newline) };
    auto document { std::make_unique<jam::AnsiDocument>() };

    document->addLines (text.data(), static_cast<int> (text.size()));
    editor.setLiveBlock (std::move (document));
    editor.setDefaultMode (map::EditorMode::normal);
    editor.setMode (map::EditorMode::insert);
}

juce::String InputView::getText() const
{
    return editor.getTextInRange ({ 0, editor.getTotalNumChars() }).trim();
}
```

   - `setDocument` is the owner contract of Part 6 Step 6, with the call of Step 3. The document holds one empty line.
   - `getText` reads the text of the editor with the contract of Part 6 Step 6, item 4.
   - Each editor gets its LookAndFeel from `TITView`, its parent, before its first document: a document comes only from a command or a run.
   - Each later part that needs one line of typed text uses an `InputView` (Part 11, the search prompt).

**Validation:**
- `ConsoleView` and `InputView` hold no text and no flag. `ConsoleView` reads the `console` child. `InputView` gives the text of its editor through `getText()`.
- New names and their families: `ConsoleView`, `InputView` (the `…View` family of Step 11, with the nouns of the modes `console` and `input`); `editor` (`end/Source/ui/EditorView.h:215`); `setDocument` (`jam_gui/widgets/jam_TextEditor.h:196`); `getText` (`getTextInRange`, Part 6 Step 6).
- JAM gets no component, no paint code, and no size constant for a dialog. The box, its buttons, its keys, its layout, and its paint are `juce::AlertWindow` and the LookAndFeel (CODING.md, "Use the framework API fully").
- No state for the widget: no Model row and no member holds a selected button (rulings 65 and 66).
- New names and their families: `TerminalMessageBox` (ruling 65, ARCHITECT's word; `jam::MessageBox`); `showAlert`, `showYesNoBox` (`jam_MessageBox.h:27,40`); `buttonText`, `button1Text`, `button2Text` (`juce_MessageBoxOptions.h:156,177-178`); `title`, `message`, `associatedComponent`, `callback` (`jam_MessageBox.h:40-43`).

### Step 15: `TITView`, the header
**Scope:** `tit/Source/TITView.h`.
**Action:** Delete first: remove the member `help` of Part 9 Step 6, and in Step 16 its lines of the constructor and of `resized`. The help guide stays in `titc --help`. Then add the includes `"HeaderView.h"`, `"FooterView.h"`, `"ConsoleView.h"`, `"InputView.h"` after `"TITGitProcessor.h"`, and make the class this declaration. TIT has no `MenuView.h` and no `ConfirmationView.h`: the menu and the dialog are components of JAM (Steps 12 and 14).

```cpp
class TITView : public juce::Component,
                public jam::Model::Component<TITView>,
                public juce::ApplicationCommandTarget,
                private juce::ValueTree::Listener,
                private juce::Timer
{
public:
    TITView (TITModel& modelToUse, TITGitProcessor& gitProcessorToUse, juce::LookAndFeel& lookAndFeelToUse);
    ~TITView() override;

    void paint (juce::Graphics& g) override;
    void resized() override;
    void focusGained (FocusChangeType cause) override;
    bool keyPressed (const juce::KeyPress& key) override;

    juce::ApplicationCommandTarget* getNextCommandTarget() override;
    void getAllCommands (juce::Array<juce::CommandID>& commands) override;
    void getCommandInfo (juce::CommandID commandID, juce::ApplicationCommandInfo& result) override;
    bool perform (const InvocationInfo& info) override;

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    void timerCallback() override;

    void registerEvents();
    void registerAvailability();
    void registerMenuKeys();
    void registerActions();
    void registerPushActions();
    void registerOperationActions();
    void registerPullActions();
    void registerWorkingTreeActions();
    void registerInputActions();
    void registerViewActions();

    void setMenuMode (const juce::ValueTree& git);
    void setMenu (const juce::ValueTree& git);
    void setKeys (const juce::ValueTree& git);
    void setView();
    void setFocus();
    void setInput (const juce::StringArray& commandLines);
    void setConfirmation (const juce::String& title, const juce::String& explanation, const juce::String& yes, const juce::String& no,
                          const juce::StringArray& yesCommandLines, const juce::StringArray& noCommandLines);
    void cancelInput();

    juce::ValueTree getGit() const;
    juce::Identifier getMenuKey (juce::CommandID command, const juce::ValueTree& git) const;
    juce::Array<juce::CommandID> getMenuCommands() const;
    bool isAvailable (juce::CommandID command, const juce::ValueTree& git) const;
    bool isMenuMode() const;
    bool isEditMode() const;
    bool isLargeEnough() const;

    static constexpr int minColumns { 70 };
    static constexpr int minRows { 20 };
    static constexpr int confirmMilliseconds { 2000 };

    TITModel& model;
    TITGitProcessor& gitProcessor;
    jam::SharedInstance<map::CommandId> commandId { std::in_place };
    jam::SharedInstance<jam::CommandManager> commandManager;
    jam::Function::Map<juce::CommandID, jam::Function::pointerCapacity, bool (const juce::ValueTree&)> availability;
    jam::Function::Map<juce::CommandID, jam::Function::pointerCapacity, juce::Identifier (const juce::ValueTree&)> menuKeys;
    jam::Function::Map<juce::CommandID, jam::Function::pointerCapacity, void ()> commands;
    jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, void (juce::ValueTree&)> events;
    HeaderView header;
    FooterView footer;
    ConsoleView console;
    InputView commitMessage;
    juce::ScopedMessageBox messageBox;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITView)
};
```

- `TITView` holds no menu object. `setMenu` gives a `juce::PopupMenu` to `jam::showAsync`, and JAM owns the `MenuWindow` until it closes (`jam_gui/menu/jam_PopupMenu.h:1848-1893,1905-1949`).
- `messageBox` is the `juce::ScopedMessageBox` of the open dialog, the form of `jam::MessageBox::licenseNotFound` (`jam_gui/windows/jam_MessageBox.h:139-149`). `TITView` reads no state from the menu or from the box. The box gives the pressed button as the argument of its callback (Step 14).
- The timer base ends the two confirm rows after `confirmMilliseconds`. It belongs to the view, so it stops when the view goes (MANIFESTO B). `EVEView` has the same base (`eve/Source/EVEView.h:20-26`).
- `minColumns` and `minRows` are the numbers of the legacy text (legacy fault 3). `confirmMilliseconds` is the legacy `QuitConfirmTimeout`, "Timeout for quit/clear confirmation dialogs" (`legacy/constants.go:55`).

**Validation:**
- No member holds state. The four view members and the two owned components are views; the four maps are tables.
- `TITView` composes: it owns the views, shows one view for each mode, and registers the commands. It makes no document, no command line, and no file.
- New names and their families: `registerEvents`, `events` (Part 7 Step 7); `registerMenuKeys`, `menuKeys`, `getMenuKey` (the `key` column of `Source/menu.md`; NAMES.md Rule 8); `registerPullActions`, `registerWorkingTreeActions`, `registerInputActions`, `registerViewActions` (the `register…Actions()` family of `end/Source/ui/ENDView.h:121-179`, with the nouns of the command `pullMerge`, of the row `workingTree`, and of the mode `input`); `setMenu`, `setView`, `setFocus`, `setInput`, `setConfirmation` (`setMenuMode`, Part 9); `messageBox` (`juce::ScopedMessageBox`; `jam_MessageBox.h:147`); `cancelInput` (the command `cancel` and the mode `input`); `getGit` (`getDefaultGit`, Part 8); `getMenuCommands` (`getAllCommands`, JUCE; the row `menu`); `isMenuMode`, `isEditMode`, `isLargeEnough` (NAMES.md Rule 1; legacy `CheckIsTooSmall`, as a positive question; the edit commands of Part 6); `minColumns`, `minRows` (legacy `MinWidth`, `MinHeight`); `header`, `menu`, `footer`, `confirmation`, `console`, `commitMessage` (the view and row words).

### Step 16: `TITView`, the source
**Scope:** `tit/Source/TITView.cpp`.
**Action:** The functions `getNextCommandTarget`, `getAllCommands`, `getCommandInfo`, `perform`, `keyPressed`, `setMenuMode`, `setKeys`, `registerPushActions`, and `registerOperationActions` of Part 9 Step 7 stay. Replace or add the others.

1. The constructor, the destructor, and the component functions:

```cpp
TITView::TITView (TITModel& modelToUse, TITGitProcessor& gitProcessorToUse, juce::LookAndFeel& lookAndFeelToUse)
    : jam::Model::Component<TITView> (modelToUse.state)
    , model (modelToUse)
    , gitProcessor (gitProcessorToUse)
    , header (modelToUse.state)
    , footer (modelToUse.state)
    , console (modelToUse.state.getChildWithName (jam::Format::toType (Id::console)), gitProcessorToUse.console.userInterfaceGetters, gitProcessorToUse.console.chainEvents)
{
    setLookAndFeel (&lookAndFeelToUse);
    setWantsKeyboardFocus (true);
    addAndMakeVisible (header);
    addAndMakeVisible (footer);
    addChildComponent (console);
    addChildComponent (commitMessage);

    registerEvents();
    registerAvailability();
    registerMenuKeys();
    registerActions();

    jam::Component::applyFunctionRecursively (this, [this] (juce::Component* child)
    {
        commandManager->registerAllCommandsForTarget (dynamic_cast<juce::ApplicationCommandTarget*> (child));
    });

    commandManager->registerAllCommandsForTarget (this);

    setMenuMode (getGit());
    setKeys (getGit());
    model.addListener (this);
}

TITView::~TITView()
{
    model.removeListener (this);
    setLookAndFeel (nullptr);
}

void TITView::paint (juce::Graphics& g)
{
    g.fillAll (findColour (HeaderView::backgroundColourId));

    if (not isLargeEnough())
    {
        g.setColour (findColour (HeaderView::textColourId));
        g.drawText (text::Footer::tooSmall, getLocalBounds(), juce::Justification::centred);
    }
}

void TITView::resized()
{
    auto content { getLocalBounds() };

    header.setBounds (content.removeFromTop (HeaderView::numRows));
    footer.setBounds (content.removeFromBottom (FooterView::numRows));
    console.setBounds (content);
    commitMessage.setBounds (content);

    setView();
    setMenu (getGit());
}

void TITView::focusGained (FocusChangeType)
{
    setFocus();
}
```

2. The view of a mode:

```cpp
void TITView::setView()
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };
    const auto hadFocus { hasKeyboardFocus (true) };
    const auto isLarge { isLargeEnough() };

    header.setVisible (isLarge);
    footer.setVisible (isLarge);
    console.setVisible (isLarge and mode == map::EditorMode::console);
    commitMessage.setVisible (isLarge and mode == map::EditorMode::input);

    if (hadFocus)
        setFocus();

    repaint();
}

void TITView::setFocus()
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    if (mode == map::EditorMode::console)
        console.grabKeyboardFocus();
    else if (mode == map::EditorMode::input)
        commitMessage.grabKeyboardFocus();
    else
        grabKeyboardFocus();
}

bool TITView::isMenuMode() const
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    return not (mode == map::EditorMode::console or mode == map::EditorMode::confirmation or mode == map::EditorMode::input);
}

bool TITView::isLargeEnough() const
{
    return getWidth() >= minColumns and getHeight() >= minRows;
}

juce::ValueTree TITView::getGit() const
{
    return state.getChildWithName (jam::Format::toType (Id::git));
}
```

   The focus follows the form of EVE's panel: the view that shows takes the focus when the tree had it (`eve/Source/EVEView.cpp:204-223`).
   - In a menu mode the menu is modal, so the peer gives it each key first (Step 12, the table). A key that it does not take goes up to `TITView::keyPressed` and to the key sheet.
   - In the mode `confirmation` the dialog is in a modal state (`juce_gui_basics/detail/juce_AlertWindowHelpers.h:52`), so `setFocus` has no branch for that mode.

3. The events, in the forms of Part 7 Step 7, items 1 and 2:

```cpp
void TITView::registerEvents()
{
    events.add (Id::operation, [this] (const juce::Identifier&, juce::ValueTree& git)
    {
        if (isMenuMode())
            setMenuMode (git);

        setKeys (git);
    });

    for (const auto& id : { Id::workingTree, Id::timeline, Id::remote })
        events.add (id, [this] (const juce::Identifier&, juce::ValueTree& git) { setKeys (git); });

    events.add (Id::isRunning, [this] (const juce::Identifier&, juce::ValueTree& root)
    {
        const auto isRunning { static_cast<bool> (root.getProperty (Id::isRunning)) };

        console.setMode (isRunning ? map::EditorMode::terminal : map::EditorMode::reader);

        if (isRunning)
            state.setProperty (Id::mode, static_cast<int> (map::EditorMode::console), nullptr);
    });

    events.add (Id::numShellExits, [this] (const juce::Identifier&, juce::ValueTree& tree)
    {
        if (tree.hasType (jam::Format::toType (Id::console)))
            gitProcessor.applyShellExit();
    });

    events.add (Id::mode, [this] (const juce::Identifier&, juce::ValueTree&)
    {
        setView();
        setKeys (getGit());
        setMenu (getGit());
    });
    events.add (Id::isQuitConfirmActive, [this] (const juce::Identifier&, juce::ValueTree&) { startTimer (confirmMilliseconds); });
    events.add (Id::isClearConfirmActive, [this] (const juce::Identifier&, juce::ValueTree&) { startTimer (confirmMilliseconds); });
}

void TITView::valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property)
{
    if (tree.hasType (jam::Format::toType (Id::git)))
        setMenu (tree);

    if (events.contains (property))
        events.get (property, tree);
}

void TITView::timerCallback()
{
    stopTimer();
    state.setProperty (Id::isQuitConfirmActive, false, nullptr);
    state.setProperty (Id::isClearConfirmActive, false, nullptr);
}
```

   - The console has no text row here: its editor reads the swap of the run chain (Step 14). The `isRunning` event tells the console its mode: `terminal` while a run runs, so the view follows the output, and `reader` after it, so the user scrolls with `j` and `k`.
   - The `numShellExits` event tells the processor that a session of the console chain ended (Step 10, item 6). The event of another chain has another tree, and no run waits for it. `TITView` is the object of this layer that stands on both axes, the processor and the views (`end/ARCHITECTURE.md:907-909`), so it carries the event across, as EVE's view does for the `numCommands` row (`eve/Source/EVEView.cpp:138-144`).
   - The state builds what the user can reach (ruling 56). Each row that a rule of item 6 reads has an event, and the event builds the key sheet: the four axis rows of the `git` child, and the `mode` row. The loop over the three axis rows is the form of `end/Source/ui/ENDView.cpp:408-412`. A `git` event also builds the menu, in `valueTreePropertyChanged`.
   - Delete first: the `valueTreePropertyChanged` of Part 9 Step 7, item 3, goes. The one below takes its place.
   - A confirm row that goes to `false` starts the timer one more time; its callback then writes two equal values, and an equal value gives no event (JUCE `ValueTree::setProperty`).

4. The menu (Closed by a read, items 11 and 12):

```cpp
juce::Array<juce::CommandID> TITView::getMenuCommands() const
{
    static const juce::Array<juce::CommandID> operationCommands {
        map::CommandId::commit, map::CommandId::commitPush, map::CommandId::resetDiscardChanges, map::CommandId::push,
        map::CommandId::pullMerge, map::CommandId::forcePush, map::CommandId::replaceLocal, map::CommandId::finalizeMerge,
        map::CommandId::abortMerge, map::CommandId::rebaseContinue, map::CommandId::rebaseAbort,
    };

    return operationCommands;
}

void TITView::setMenu (const juce::ValueTree& git)
{
    auto& windows { jam::PopupMenu::MenuWindow::getActiveWindows() };

    for (int index { windows.size() }; --index >= 0;)
        windows.getUnchecked (index)->dismissMenu (nullptr);

    if (isShowing() and isMenuMode() and isLargeEnough())
    {
        const auto rows { jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcMenu })->getValueTree (jam::Format::toType (Id::config)).getChildWithName (jam::Format::toType (Id::menu)) };
        juce::PopupMenu menu;
        juce::CommandID firstCommand { 0 };

        for (const auto command : getMenuCommands())
        {
            if (isAvailable (command, git))
            {
                const auto row { rows.getChildWithName (jam::Format::toType (getMenuKey (command, git))) };
                juce::PopupMenu::Item item { row.getProperty (Id::emoji).toString() + juce::String::charToString (Chars::space) + row.getProperty (Id::value).toString() };

                item.itemID = command;
                item.commandManager = commandManager.get();
                item.shortcutKeyDescription = row.getProperty (Id::shortcut).toString();
                menu.addItem (std::move (item));
                firstCommand = firstCommand == 0 ? command : firstCommand;
            }
        }

        jam::showAsync (menu, juce::PopupMenu::Options().withParentComponent (this).withInitiallySelectedItem (firstCommand), [this] (int)
        {
            if (jam::PopupMenu::MenuWindow::getActiveWindows().isEmpty())
                setMenu (getGit());
        });
    }
}

juce::Identifier TITView::getMenuKey (juce::CommandID command, const juce::ValueTree& git) const
{
    if (menuKeys.contains (command)) return menuKeys.get (command, git);
    return commandId->get (command);
}

void TITView::registerMenuKeys()
{
    menuKeys.add (map::CommandId::push, [this] (const juce::CommandID&, const juce::ValueTree& git)
    {
        return commandId->get (static_cast<int> (git.getProperty (Id::timeline)) == map::Timeline::diverged ? map::CommandId::pushAutoSync : map::CommandId::push);
    });

    menuKeys.add (map::CommandId::pullMerge, [this] (const juce::CommandID&, const juce::ValueTree& git)
    {
        return static_cast<int> (git.getProperty (Id::timeline)) == map::Timeline::diverged ? juce::Identifier { Id::pullMergeDiverged } : commandId->get (map::CommandId::pullMerge);
    });
}
```

   - `getMenuCommands()` holds the order of the menu, and `setMenu` builds the menu: two responsibilities, two functions. Each later part adds its commands to the table. Part 13 gives each submenu its own table in the same function.
   - The order of `operationCommands` is the legacy order of the items (`legacy/app/menu_render_core.go:35-154`; `menu_render_extra.go:117-131`).
   - `setMenu` shows the menu with the stock call of JAM, `jam::showAsync` (Step 12, the table). The parent component is the view, so the menu is a child component and no window is made.
   - An item is a `juce::PopupMenu::Item` (JUCE `juce_gui_basics/menus/juce_PopupMenu.h:139-189`): `text` is the emoji and the label, `shortcutKeyDescription` is the `shortcut` cell, `itemID` is the command id, and `commandManager` is the command manager. A run item invokes its command through that manager with `InvocationInfo::fromMenu` (`jam_PopupMenu.h:870-879,1852-1860`). So return on an item and the key of the item run the same command, and no code of `TITView` reads which item is highlighted. `commandManager.get()` is the accessor of `jam::SharedInstance` (doxygen `classjam_1_1SharedInstance`, `get`).
   - The first item is the highlighted item of a new menu: `withInitiallySelectedItem` (Step 12, the table). The wrap of the highlight is in the menu.
   - A menu that is open closes before the new one shows: `getActiveWindows()` and `dismissMenu (nullptr)`. The loop counts down, because a closed menu can leave the list.
   - The menu closes when an item runs and when the user presses Escape (`jam_PopupMenu.h:879-886,1003-1006`). The callback of `jam::showAsync` runs after the close, on the message thread, and after the command of the item: JUCE calls the callbacks of a modal component from an async update, the last attached one first (JUCE `juce_gui_basics/components/juce_ModalComponentManager.cpp:196-216`; `jam_PopupMenu.h:1934-1935`). The callback shows the menu again when no menu is open. `setMenu` shows nothing outside a menu mode, so a command that opened the dialog, the commit box, or the console leaves no menu.
   - `setMenu` runs at each event of the `git` rows, at each change of the mode, and at each `resized`. Before the view is on the screen `isShowing()` is false, and the first `resized` on the screen shows the first menu.
   - A config row is found by the type form of its key, as the sheet build finds a key row (Part 7 Step 7, item 3).

5. The command target:

```cpp
bool TITView::perform (const InvocationInfo& info)
{
    const auto isPerformed { commands.contains (info.commandID) };

    if (isPerformed)
    {
        state.setProperty (Id::message, juce::String {}, nullptr);
        commands.get (info.commandID);
    }

    return isPerformed;
}

bool TITView::isAvailable (juce::CommandID command, const juce::ValueTree& git) const
{
    return not availability.contains (command) or availability.get (command, git);
}

bool TITView::isEditMode() const
{
    return static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::input;
}
```

   - `getCommandInfo` of Part 9 stays: it names the command and sets no flag (ruling 55).
   - `perform` reads no state (ruling 56). It is the END body with one statement more: each command clears the footer message before it runs (Closed by a read, item 10).
   - `isAvailable` loses its `isRunning` term (Closed by a read, item 8).
   - `isEditMode` answers if the mode of TIT shows an editable editor. Each later part that adds such a mode adds it here.

6. `registerAvailability()`: the five entries of Part 9 Step 7, item 5, stay. Add these entries in the same form. Each rule of a git command also needs `operation` to be `map::Operation::normal`.

   | Command | Rule |
   |---|---|
   | `commit`, `resetDiscardChanges` | `workingTree` is `dirty` |
   | `commitPush` | `workingTree` is `dirty`, and `remote` is `hasRemote` |
   | `forcePush` | `remote` is `hasRemote`, and `timeline` is `diverged`, or `timeline` is `ahead` with `workingTree` `clean` |
   | `replaceLocal` | `remote` is `hasRemote`, and `timeline` is `behind` or `diverged` |
   | `pullMerge` | `remote` is `hasRemote`, and `workingTree` is `clean`, and `timeline` is `behind` or `diverged` |
   | the thirteen edit commands of the key rows of Step 7 | `isEditMode()` |

   - The entry of the edit commands is one loop, in the form of `end/Source/ui/ENDView.cpp:240-242`: `for (const juce::CommandID command : { map::TextEditorCommandId::insertBefore, map::TextEditorCommandId::insertAfter, map::TextEditorCommandId::insertLineStart, map::TextEditorCommandId::insertLineEnd, map::TextEditorCommandId::openLineBelow, map::TextEditorCommandId::openLineAbove, map::TextEditorCommandId::deleteAction, map::TextEditorCommandId::change, map::TextEditorCommandId::deleteCharacter, map::TextEditorCommandId::pasteAfter, map::TextEditorCommandId::pasteBefore, map::TextEditorCommandId::undo, map::TextEditorCommandId::redo }) availability.add (command, [this] (const juce::CommandID&, const juce::ValueTree&) { return isEditMode(); });`.
   - `cancel` and `inputSubmit` have no entry. The mode of the key row is the gate of each one, and `isAvailable` allows a command with no entry.

7. `registerActions()` calls one function for each family, in the END form (`end/Source/ui/ENDView.cpp:385-400`). `registerPushActions()` and `registerOperationActions()` of Part 9 Step 7 stay. Replace `registerActions()` and add the families of this part:

```cpp
void TITView::registerActions()
{
    registerPushActions();
    registerOperationActions();
    registerPullActions();
    registerWorkingTreeActions();
    registerInputActions();
    registerViewActions();
}

void TITView::registerWorkingTreeActions()
{
    commands.add (map::CommandId::commit, [this] (const juce::CommandID&) { setInput ({ text::Git::addAll, text::Git::commitFile }); });
    commands.add (map::CommandId::commitPush, [this] (const juce::CommandID&) { setInput ({ text::Git::addAll, text::Git::commitFile, TITGitProcessor::getPushLine (getGit()) }); });

    commands.add (map::CommandId::resetDiscardChanges, [this] (const juce::CommandID&)
    {
        if (static_cast<int> (getGit().getProperty (Id::remote)) == map::Remote::hasRemote)
            setConfirmation (text::Confirmation::discardChangesRemoteChoiceTitle, text::Confirmation::discardChangesRemoteChoiceExplanation, text::Confirmation::discardChangesRemoteChoiceYes, text::Confirmation::discardChangesRemoteChoiceNo,
                             { text::Git::resetHardHead, text::Git::clean }, TITGitProcessor::getReplaceLocalLines (getGit()));
        else
            setConfirmation (text::Confirmation::discardChangesLocalTitle, text::Confirmation::discardChangesLocalExplanation, text::Confirmation::discardChangesLocalYes, text::Confirmation::discardChangesLocalNo,
                             { text::Git::resetHardHead, text::Git::clean }, {});
    });
}

void TITView::registerInputActions()
{
    commands.add (map::CommandId::inputSubmit, [this] (const juce::CommandID&)
    {
        const auto message { commitMessage.getText() };

        if (message.isEmpty())
        {
            state.setProperty (Id::message, text::Footer::commitMessageEmpty, nullptr);
        }
        else
        {
            gitProcessor.setCommitMessage (message);
            setMenuMode (getGit());
            gitProcessor.run (juce::StringArray::fromLines (state.getProperty (Id::commandLines).toString()), {});
        }
    });
}

void TITView::registerViewActions()
{
    commands.add (map::CommandId::cancel, [this] (const juce::CommandID&)
    {
        const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

        if (mode == map::EditorMode::input)
            cancelInput();
        else if (mode == map::EditorMode::console and static_cast<bool> (state.getProperty (Id::isRunning)))
            gitProcessor.cancel();
        else
            setMenuMode (getGit());
    });
}

void TITView::cancelInput()
{
    if (commitMessage.getText().isEmpty())
    {
        setMenuMode (getGit());
    }
    else if (static_cast<bool> (state.getProperty (Id::isClearConfirmActive)))
    {
        state.setProperty (Id::isClearConfirmActive, false, nullptr);
        commitMessage.setDocument();
    }
    else
    {
        state.setProperty (Id::isClearConfirmActive, true, nullptr);
    }
}
```

   - Each family is one responsibility: the commands of a dirty working tree, the submit of the input, and the cancel of a view. No family is longer than 30 lines.
   - No family moves a menu or runs its highlighted item: the menu does both itself (Step 12).
   - `cancel` is one command id for Escape. Its action follows the mode, as `close` of END follows its state (`end/Source/ui/EditorView.h:110-116`).
   - Escape in a dialog is a key of the second button of the box (Step 14, the table). So `cancel` has no row in the mode `confirmation`.
   - `cancelInput` carries the legacy Escape rule of an input (inventory).
   - The text of the commit box comes from `InputView::getText` (Step 14). The message file is written by `TITGitProcessor::setCommitMessage` (Step 10). `TITView` passes the text from the one to the other.

8. The dialog commands. Add this entry at the end of `registerPushActions()` of Part 9:

```cpp
    commands.add (map::CommandId::forcePush, [this] (const juce::CommandID&)
    {
        setConfirmation (text::Confirmation::forcePushTitle, text::Confirmation::forcePushExplanation, text::Confirmation::forcePushYes, text::Confirmation::forcePushNo,
                         { TITGitProcessor::getForcePushLine (getGit()) }, {});
    });
```

   Then the pull family and the keys of the dialog:

```cpp
void TITView::registerPullActions()
{
    commands.add (map::CommandId::replaceLocal, [this] (const juce::CommandID&)
    {
        setConfirmation (text::Confirmation::hardResetTitle, text::Confirmation::hardResetExplanation, text::Confirmation::hardResetYes, text::Confirmation::hardResetNo,
                         TITGitProcessor::getReplaceLocalLines (getGit()), {});
    });

    commands.add (map::CommandId::pullMerge, [this] (const juce::CommandID&)
    {
        if (static_cast<int> (getGit().getProperty (Id::timeline)) == map::Timeline::diverged)
            setConfirmation (text::Confirmation::pullMergeDivergedTitle, text::Confirmation::pullMergeDivergedExplanation, text::Confirmation::pullMergeDivergedYes, text::Confirmation::pullMergeDivergedNo,
                             { text::Git::pullMerge }, {});
        else
            setConfirmation (text::Confirmation::pullMergeTitle, text::Confirmation::pullMergeExplanation, text::Confirmation::pullMergeYes, text::Confirmation::pullMergeNo,
                             { text::Git::pullMerge }, {});
    });
}
```

   - The force push line and the lines of `replaceLocal` come from the processor (Step 10). Each entry gives texts and command lines to `setConfirmation`, and makes none of them.
   - A button with no line closes the dialog, as `Cancel` does in legacy. Escape closes it and runs nothing.
   - The selected button of each dialog and its command lines are the legacy ones (inventory, "The legacy commands of this part").

9. The setters:

```cpp
void TITView::setConfirmation (const juce::String& title, const juce::String& explanation, const juce::String& yes, const juce::String& no,
                               const juce::StringArray& yesCommandLines, const juce::StringArray& noCommandLines)
{
    static constexpr int firstButton { 1 };

    state.setProperty (Id::mode, static_cast<int> (map::EditorMode::confirmation), nullptr);
    messageBox = jam::TerminalMessageBox::showYesNoBox (title, explanation, yes, no, this, [this, yesCommandLines, noCommandLines] (int result)
    {
        setMenuMode (getGit());
        setFocus();

        if (const auto& commandLines { result == firstButton ? yesCommandLines : noCommandLines }; not commandLines.isEmpty())
            gitProcessor.run (commandLines, {});
    });
}

void TITView::setInput (const juce::StringArray& commandLines)
{
    state.setProperty (Id::commandLines, commandLines.joinIntoString (juce::String::charToString (Chars::newline)), nullptr);
    commitMessage.setDocument();
    state.setProperty (Id::mode, static_cast<int> (map::EditorMode::input), nullptr);
}
```

   - `setInput` tells the commit box to take an empty document (`InputView::setDocument`, Step 14). The console needs no call: its editor reads the swap of the run chain (`ConsoleView`, Step 14).
   - `setConfirmation` shows the dialog with `jam::TerminalMessageBox::showYesNoBox` (Step 14). `TITView` is the associated component and the parent component, so the box is a child of the view and takes its LookAndFeel.
   - `messageBox` holds the returned `juce::ScopedMessageBox`. The box stays open no longer than that member lives (`jam_gui/windows/jam_MessageBox.h:139-149`). A new dialog replaces the member.
   - The callback holds what each button runs: the command lines of the two buttons, by value. No Model row holds them (ruling 65). The result is 1 for the first button and 0 for the second (Step 14, the table). `firstButton` names the value in the one function that reads it, the form of `contactSupportButton` (`jam_gui/windows/jam_MessageBox.cpp:125`).
   - Enter, Escape, and the first letter of each label are keys of the buttons of JUCE (Step 14, the table). So TIT has no command and no key row for a dialog.
   - The mode `confirmation` stays a row of the Model. It is state of the machine: it selects the footer text (Step 13). It is not state of a widget.

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines.
- No `==` on strings. Each compare is an Identifier compare or an integer compare.
- `commands` holds 14 ids. `availability` holds the 11 ids that a state rule governs, and the 13 edit command ids of the editor. No callback and no line of `perform` tests if a command is allowed: `isAvailable` runs only inside `setMenu` and `setKeys`, and an event calls each one.
- `HeaderView`, `FooterView`, and `ConsoleView` paint the Model rows. The menu and the dialog are shown with their data. `TITView` reads no state from a child component, except the text of the commit box through `InputView::getText`.
- No Model row holds an item, a highlight, a selected button, or a text of a dialog (rulings 61 and 65).
- `TITView` holds no `jam::TextEditor`, makes no document, makes no command line, and writes no file (Part 9, Overview, item 6).
- Data flows one way: key, sheet, command, Model row or processor, listener, view and sheet.
- Local names of the bodies: `mode`, `hadFocus`, `isLarge`, `content`, `row`, `items`, `item`, `command`, `firstCommand`, `message`, `result`, `commandLines`, `newline`, `text`, `bytes`, `document`, `root`, `git`, `block`.

### Step 17: The application, the quit
**Scope:** `tit/Source/Main.cpp`.
**Action:** In `runWindow`, after the statement that makes `gitProcessor` (Part 8 Step 8), add `gitProcessor->prepare();`. It opens and starts the clock of each run chain (Step 10), as the application prepares the terminal chain (Part 7 Step 7). The destructor of the processor stops it. Then, in the class `Application`, add the override:

```cpp
    void systemRequestedQuit() override
    {
        if (static_cast<bool> (model.state.getProperty (Id::isRunning)))
            model.state.setProperty (Id::message, text::Footer::operationInProgress, nullptr);
        else if (static_cast<bool> (model.state.getProperty (Id::isQuitConfirmActive)))
            quit();
        else
            model.state.setProperty (Id::isQuitConfirmActive, true, nullptr);
    }
```

- The JUCE command `quit` calls `systemRequestedQuit()` (Part 7, command facts). Part 7 removed the override because it equalled the default. Now it differs.
- The three branches are the legacy order: a run blocks the quit, the second press quits, the first press shows the text for `confirmMilliseconds` (`legacy/app/handlers_global_keys.go:52-78`).
- The command-line runs of Part 1 call `quit()` and never this function.

**Validation:**
- `runWindow` does not change: it still gives `model`, `*gitProcessor`, and `lookAndFeel` to `TITView` (Part 9 Step 8). `lookAndFeel` is the `TITLookAndFeel` of Part 7 Step 4 (ruling 62).
- The application holds no flag. The two rows are Model rows.

### Step 18: Build, run, and measure on macOS
**Scope:** JAM, `ling`, EVE, and the `tit/` project; the fixture repositories of Part 8 Step 9.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build). The fixture commands are git commands: ARCHITECT runs them, or gives the git instruction (CAROL.md Git).
**Action:**
1. Build `ling` with `cast cast/spell.md --debug`. Run `ling --help`, and one `ling` build of a manifest.
2. Build EVE. Press the keys of Part 6 Step 9.
3. From `tit/`, run `cast cast/spell.md --debug`.
4. Start `titc` in END in each fixture below, and do the keys. The fixture numbers are those of Part 8 Step 9.

   | Fixture | Keys | Expected |
   |---|---|---|
   | 1, an empty folder | none | A blank header, an empty menu, an empty footer. No cell is magenta. |
   | 4, one commit and one new file | none | The header shows the folder, `🔌 NO REMOTE`, `🟢 READY`, the branch, `📝 Dirty ● 1`, `Local changes present`, `🔌 N/A`, `No remote configured.` The menu shows `Commit changes` and `Discard all changes`. The footer shows the hint of the first item. |
   | 4 | cursor down, cursor up, `j`, `k` | The highlight moves and wraps. |
   | 4 | `c`, a text of two lines with Return between them, Escape, Return | The commit box opens in insert mode with a bar caret. During the run, the footer shows each command line. The console shows the output of the tool. `log -1` of the fixture shows the two lines of the text. |
   | 4 after the commit, a new file | `c`, Escape, Return | The footer shows `Commit message cannot be empty`. |
   | 4 | `c`, a text, Escape, Escape, Escape | The second Escape shows `Press ESC again to clear input (2s timeout)`. The third Escape clears the text. A fourth Escape returns to the menu. |
   | 4 | `c`, a text with `q` in it | The `q` is in the text. titc does not show the quit text. |
   | 4 | `c`, a text, Escape, `dd`, `u`, `i`, a text, Escape, `v`, Escape | Each edit command of Part 6 Step 8 works. Escape in visual mode gives normal mode. |
   | 4 | `ctrl+r`, Escape | The dialog `Discard Changes` shows inside the view. Escape closes the dialog and the file stays. |
   | 4 | `ctrl+r`, Return | The footer shows the two command lines, one after the other. Then the header shows `✅ Clean`. |
   | 5, `Ahead` by one commit | `}` , Escape; then `}`, Return | The dialog `Force Push Confirmation` shows. After Return the console shows the push with its progress lines, each on one line. Then the header shows `🔗 In sync`. |
   | 6, `Diverged` | none | The menu shows `Push (auto sync)`, `Pull (merge)`, `Force push`, `Replace local`, with the shortcuts `]`, `[`, `shift + ]`, `x`. |
   | 6 | `[`, Return | The dialog `Pull diverged branches?` shows. After Return the console shows the pull. |
   | 8, `Behind` by one commit | `x`, Return | The dialog `Replace Local Confirmation`. The footer shows three command lines, one after the other. Then `🔗 In sync`. |
   | 5 with a new commit and an address of origin that does not answer | `]`, then Escape during the run, then Escape | The footer shows `↑↓ scroll  │  Esc abort`. After the first Escape the footer shows `Operation aborted by user`. No more command line starts. The second Escape returns to the menu. |
   | 10, a merge in progress | none | The header shows `🔀 MERGING`. The menu shows `Finalize merge` and `Abort merge`. |
   | 13, a detached HEAD | none | The header shows `DETACHED`, the hash as the branch, and `DETACHED @` with the hash. The menu is empty. |
   | 16, a pin marker | none | The header shows `PINNED @` with the hash. |
   | any | Ctrl+C, then wait 3 seconds, then Ctrl+C two times | The first press shows `Press Ctrl+C again to quit (2s timeout)`. The text goes after 2 seconds. Two presses quit. |
   | any, in the menu | `q`, `q` | titc quits. |
   | any | make the END pane 60 columns wide | The window shows `Terminal too small. Resize to >= 70×20.` and no view. A wider pane shows the views again. |

5. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- Each build ends with exit code 0. `ling --help` is byte-identical to its output before this part, and a `ling` build gives the same record (Step 1). EVE keeps the behaviour of Part 6 Step 9 (Steps 2 and 3).
- Each row gives its expected result.
- These facts of the tracker list "Facts that execution measures first" are measured: a live, editable `jam::TextEditor` on a terminal.
- These facts of this part are measured: the cast emits `\n` of a value as a newline (Step 6); two rows with the key `quit` give two children of the key tree (Step 7); `.git/COMMIT_EDITMSG` gives the typed message (Closed by a read, item 13); a custom of a TIT view reaches its colour id (Step 8).
- ARCHITECT sets the layout of each view at this run (ruling 7). A change of a geometry constant is a change of its value only.
- The second `cast` run changes no generated file.

### Step 19: Run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 18, items 3 to 5, in END on a Windows machine with fixtures 1, 4, 5, and 6.
**Validation:**
- The results of Step 18 hold.
- The progress lines of a push arrive live: the child writes to the ConPTY of the run chain (ruling 59: "Progress lines arrive live on Windows (ConPTY)").

## BLESSED Alignment

- **B — Bound.** `TITView` owns its six views. `ConsoleView` and `InputView` each own one editor. Each view is destroyed with `TITView`, before the Model that it reads. The timer of the confirm rows is a base of `TITView`. `TITGitProcessor` owns the console chain, a `TITRunProcessor`, which owns its clock, its TTY, its decoder, its buffer, and its swap, in the member order of EVE. JAM owns the window of a menu and the box of a dialog.
- **L — Lean.** Four views in one form. One menu table, one key sheet, three text tables. Only the commands, rows, colours, and texts that this part reads. Each state table of the header is an array, not a branch chain.
- **E — Explicit.** Each text, key, command line, colour, and geometry number is a named row or a named constant. A command that git would reject is not in the menu and has no key in the sheet.
- **S — SSOT.** The `mode` row selects the view and gates the keys. `Source/menu.md` is the one list of the menu texts. The swap of the run chain is the one store of the console lines.
- **S — Stateless.** No view holds a selection, a text, or a flag. `TITView` and the processor hold tables and references only.
- **E — Encapsulation.** A view reads the Model and paints. `TITView` tells the processor to run or to cancel, and tells it when a session ended. It reads no result from it. The exit status comes from `jam::TerminalTTY::close()`, its producer (Step 1).
- **D — Deterministic.** The same Model rows give the same header, menu, and footer, and the same bytes of the tool give the same console lines.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. The nineteen items of "Closed by a read" are ORACLE's reads; ARCHITECT can reverse each one. The runtime facts of this part are in Steps 18 and 19.
