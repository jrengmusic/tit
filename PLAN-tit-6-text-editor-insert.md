# PLAN: TIT Part 6 — Insert Mode and the Edit Operators of `jam::TextEditor`

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-13
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part changes JAM only: `jam_gui/widgets/jam_TextEditor`, `jam_TextEditorStatusBar`, and the cast tables. It restores the edit machinery of `jam::TextEditor` verbatim from JAM revision `3585cb66d`, the parent of `abda82478`, the revision that removed it (ruling 5; `eve/PLAN-Grand-Scheme.md:594-595`).

- **Restored verbatim:** the edit actions, their helper lambdas, the undo actions `InsertLine` and `RemoveLine`, the primitives `addLine`, `removeLine`, `setLine`, `addText`, and the `juce::UndoManager`.
- **Conformed, names and placement only:** today's command keys (`Id::…` in place of `map::EditorAction::…`), today's mode row (`getMode`, `setMode`), today's storage (the `live` `TextBlock` in place of `historyDocument`, `history` and the swap), and today's layout call (`setBlocks()`).
- **Added by rulings:** the `insert` row of `map::EditorMode` and the INSERT label (ruling 16); the editable document is an `AnsiDocument` set through a public `setLiveBlock` (ruling 32); `juce::TextInputTarget`, so a pasted text arrives whole (tracker, "Closed by a read").

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Part 0 only. It affects EVE, a consumer of `jam::TextEditor` (`eve/Source/EVEView.cpp:316`). EVE keeps its behaviour (Step 9).

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`. `InsertLine` and `RemoveLine` stay at class scope of `TextEditor`, as they were (old `jam_TextEditor.h:240-277`).
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization; "NEVER use raw `new`"; no `==` on strings.
- CAROL.md, Refactor-Rewrite Discipline: no old code coexists with new. This part adds members that do not exist today. It deletes nothing.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build: agents build only on ARCHITECT's explicit request.

## Dependency & API Inventory

JAM paths are relative to `~/Documents/Poems/dev/jam/`. "Old" is JAM revision `3585cb66d`. COUNSELOR or @Pathfinder reads an old body with `git show 3585cb66d:jam_gui/widgets/jam_TextEditor.cpp` (CAROL.md Git: read-only commands for the main session and Pathfinder only) and puts it in the brief to @Engineer.

### The removal

| Fact | Where |
|---|---|
| `abda82478` (2026-10-25), "EVE terminal on the host-plugin contract: TTY on the audio thread, TERMINAL mode, history on the JUCE audio pipeline", removed the edit machinery. Its parent `3585cb66d` (2026-10-24) is "TextEditor nvim parity: modes, buffer, swap and status row" | `git log -S "UndoableAction" -- jam_gui/widgets` |
| "The history is read-only (R7). The TextEditor edit machinery and insert mode are deleted." | `eve/PLAN-Grand-Scheme.md:594-595` |
| The edit design: operators `d c y x p P`, `dd cc yy`, `o O i a I A`, undo and redo as line-granular `juce::UndoableAction` transactions | `eve/PLAN-Grand-Scheme.md:1376-1380` |
| "The undo actions `InsertLine` and `RemoveLine` link and unlink line elements, and a replace is one transaction of both." | `eve/PLAN-Grand-Scheme.md:1402-1403` |
| Insert caret: *"entering insert mode it changed to \| pipe cursor by default following nvim"* | `eve/SPEC.md:800-801` |

### The old bodies that this part restores

| Piece | Old place |
|---|---|
| `struct InsertLine : public juce::UndoableAction` (constructor, `perform`, `undo`; members `editor`, `lineNumber`, `line`) | old `.h:240-259`; `.cpp:1035-1067` |
| `struct RemoveLine : public juce::UndoableAction` (holds an `InsertLine insert`; `perform` is `insert.undo()`, `undo` is `insert.perform()`) | old `.h:261-277`; `.cpp:1069-1082` |
| `addLine (lineNumber, cells)`, `removeLine (lineNumber)`, `setLine (lineNumber, cells)`, `addText (lineNumber, column, text)` | old `.h:284-311`; `.cpp:1084-1137` |
| `juce::UndoManager undoManager;` | old `.h:547` |
| Helper lambdas of `getActions()`: `enterInsert`, `stepHistory`, `appendCells`, `getFirstNonBlank`, `eraseSelection` with its table `erasures`, `openLine`, `put`, `paste` | old `.cpp:382-386,413-439,484-623` |
| Actions: `insertBefore`, `insertAfter`, `insertLineStart`, `insertLineEnd`, `leaveMode`, `deleteAction`, `change`, `deleteCharacter`, `pasteAfter`, `pasteBefore`, `openLineBelow`, `openLineAbove`, `splitLine`, `deleteBackward`, `undo`, `redo` | old `.cpp:625-664,694-799` |
| The typed insert of insert mode: one text character into the cursor line | old `.cpp:942-949` |
| The insert clamp of the cursor: `isInsert ? cells.size() : …` | old `.cpp:1450-1452` |
| The insert caret: `cell.withWidth (caretWidth)`, `caretWidth { 2.0f }` | old `.cpp:1534-1535`; `.h:197` |
| The keys, from EVE's sheet of that time: `## normal` (`i a I A o O`, `d c x p P`, `u`, `ctrl+r`), `## insert` (`escape`, `return`, `backspace`), `## visual` (`d c x p P`) | EVE `Source/layout/ViewLayout.md` at `df24a978d`, `:51-135` |

### Today's editor: the pieces the restore conforms to

| Piece | Where |
|---|---|
| Bases: `juce::Component`, `jam::Model::Component<TextEditor>`, `jam::Model::Object`, `juce::ApplicationCommandTarget`, `private juce::ValueTree::Listener`, `private juce::AsyncUpdater` | `jam_gui/widgets/jam_TextEditor.h:107-112` |
| `setDocument<DocumentType>` calls `owned->addLines()`, then `setLiveBlock (std::unique_ptr<Document>)` | `jam_TextEditor.h:196-202` |
| `setLiveBlock (std::unique_ptr<Document>)` (private): registers the commands when the manager has none, builds the live block and its `Document::Index`, sets `Id::defaultMode` and the mode to `reader`, `setReadOnly (true)`, lays out | `jam_TextEditor.h:484`; `.cpp:1130-1155` |
| `TextBlock`: `document`, `index`, `runs`, `graphics`; the editor's `live` block | `jam_gui/widgets/jam_TextBlock.h:14-17`; `jam_TextEditor.h:1279` |
| `getLine`, `getCells (lineNumber)` (const); `getNumLines()`, `getNumHistoryLines()` | `jam_TextEditor.h:938,940`; `.cpp:1943-1956` |
| `setCursor (lineNumber, column)`: clamps the column to `cells.size() - 1`; `getCursor()`; the rows `Id::cursorLine`, `Id::cursorColumn`, `Id::cursorWantedColumn`, `Id::anchorLine`, `Id::anchorColumn` in `state` | `jam_TextEditor.h:1050-1056`; `.cpp:1958-1978,2041-2045` |
| `getSelectedLines`, `getSelectedColumns`, `getVisualMode` | `jam_TextEditor.cpp:1980-2019` |
| `setMode`, `getMode`, `setModeToDefault`; the mode lives in the `terminal` tree | `jam_TextEditor.h:184,1207`; `.cpp:146-149,738-741,1051-1054` |
| `Motions`, `Actions`: `jam::Function::Map<juce::Identifier, …>`; an entry lambda takes `(const juce::Identifier&, TextEditor& editor, int numRepeats)`; a call is `getActions().get (action, *this, int { numRepeats })` | `jam_TextEditor.h:1142-1150`; `.cpp:249,471,842` |
| Today's `getActions()` keeps `setVisual`, `getStartColumn`, `copySelection`, `runOperator`, `yank`, `copy`, `paste` (terminal paste) | `jam_TextEditor.cpp:389-534` |
| `perform` runs an action by the name of its command | `jam_TextEditor.cpp:820-856` |
| `getCommandInfo` names a command from `textEditorCommandId` | `jam_TextEditor.cpp:755-759` |
| `getKeyInputs()`: one key input for each mode; normal and visual use `applyCommand`; terminal calls `invokeKey`, then writes bytes | `jam_TextEditor.cpp:604-667` |
| `getCarets()`: one caret for each mode; the terminal caret has `caretThickness { 2.0f }` and `getBar` inside its lambda | `jam_TextEditor.cpp:669-719` |
| `setBlocks()` (private): lays out every block, sets the content size, the image and preview bounds, and repaints | `jam_TextEditor.h:488-489`; `.cpp:1203-1217` |
| `getCellBounds (lineNumber, column)`; `getCell (juce::Point<float>)` | `jam_TextEditor.cpp:680,1112` |
| `AnsiDocument::getText (cells, columns)` | `jam_TextEditor.cpp:437` |
| The status bar label: the mode name in upper case; `modeColours` has `normal`, `visual`, `open`; the bar hides in terminal and reader | `jam_gui/widgets/jam_TextEditorStatusBar.cpp:19-36`; `jam_TextEditor.cpp:902` |
| The status bar colour ids `0x4700001` to `0x470000a` | `cast/style.md:260-268` |
| The colour values come from the consumer's stylesheet: `--TextEditorStatusBar--normalColourId` | `eve/Source/layout/style.css:66-72` |

### Words and tables of the cast

| Table | Where |
|---|---|
| `## EditorMode`: `normal 0`, `visual 1`, `terminal 2`, `open 3`, `reader 4`, `panel 5`. EVE's contract cites the numbers 0 to 4 | `cast/gui.md:257-275`; `eve/SPEC.md:739-740` |
| The command identifier rows: `\| @id \| yank \| \`yank\` \|` | `cast/gui.md:57,92,100` |
| `## TextEditorCommandId`: 33 rows, the last key `0x4800021` (`scrollHalfPageDown`) | `cast/terminal.md:1865-1911` |
| `## textEditorKeys mac`, `## textEditorKeys windows`, `## textEditorModes`, `## textEditorLayers`: one row for each command row, in its order | `cast/terminal.md:1913,1994,2068,2115` |
| `defaultMode` has the keys `escape i a I A o O` in the modes `normal visual open` | `cast/terminal.md:1963,2099` |
| The JAM vocabulary is generated with `cast cast/spell.md` from the JAM root | `jam/carol/SPRINT-LOG.md:6297`; `cast/spell.md` index |

### `CommandManager`, `AnsiDocument`, JUCE

| Fact | Where |
|---|---|
| `getKey` returns the first row of the sheet that matches the key, the mode and the layer | `jam_gui/commands/jam_CommandManager.cpp:96-109` |
| `setKeyPresses (const juce::ValueTree&)` replaces the whole sheet | `jam_CommandManager.h:25-42`; `.cpp:5-41` |
| `getCommandInfo` of the editor names a command and clears its disabled flag. `perform` finds the action by name: `const auto isAction { getActions().contains (action) };` | `jam_gui/widgets/jam_TextEditor.cpp:755-759,820-842` |
| `AnsiDocument::addLines (data, size)`: "Each newline in the bytes adds an `Id::line` element. The text after the last newline stays in the pen as the cells of the next line." | `jam_terminal/document/jam_AnsiDocument.h:43-53` |
| `AnsiDocument::addSegments (Document::Element& line)` | `jam_AnsiDocument.h:55-63` |
| `Document::addChild (parent, after, child)`, `removeChild (parent, child)`; `Document::Index::insertLine`, `removeLine` | `jam_core/document/jam_Document.h:515,524,1121,1128` |
| `juce::TextInputTarget`: 11 pure virtuals | `juce_gui_basics/keyboard/juce_TextInputTarget.h:61-107` |
| A peer gives text to `findCurrentTextInputTarget()` only when `isTextInputActive()` | `juce_gui_basics/windows/juce_ComponentPeer.cpp:291-301` |
| No JAM or EVE class implements `juce::TextInputTarget` today | doxygen `basecompoundref` search |

### Closed by a read in this part (ARCHITECT can reverse each one)

1. **`insert` is the last row, key 6.** Appending keeps the numbers 0 to 4 that EVE's contract cites (`eve/SPEC.md:739-740`).
2. **The JAM default sheet gives no key to a normal-mode or visual-mode edit command.** EVE's editor is not read-only, because `bindToProcessor` calls `setReadOnly (false)` (`jam_TextEditor.cpp:84`). EVE's history must stay read-only (`eve/SPEC.md:736-737`). A default key such as `d` would reach EVE's history. The tested keys lived in the consumer's sheet (EVE `ViewLayout.md` at `df24a978d`). TIT's sheet binds them (Part 10 Step 7). The insert-mode keys (`escape`, `return`, `backspace`) are in the JAM sheet: EVE never enters insert mode.
3. **A read-only editor gets no edit key from its owner.** JAM adds no test: `perform` and `getCommandInfo` of the editor do not change (rulings 55 and 56). The JAM default sheet gives no key to a normal-mode or visual-mode edit command (item 2), and only an edit command enters insert mode. So an edit key exists only in the sheet of the owner. The owner builds its sheet from its state, and the sheet holds the edit rows only while the owner shows an editable editor (Part 9 Step 7, item 3; Part 10 Step 16, item 6). EVE binds no edit key (Step 9).
4. **`setBlocks()` takes the place of the old one-line layout.** Old `InsertLine` laid out one line with `setAdvances` and `setLineRows` (old `.cpp:1044-1045`). Today those are private statics of `TextBlock` (`jam_TextBlock.h:64,80`). `setBlocks()` is the existing layout path of the editor.
5. **`TextInputTarget` indexes the live document as one text.** Each line gives its cells and one newline. This is the contract of a JUCE text: `getTotalNumChars`, `getTextInRange` (`juce_TextInputTarget.h:77,92`).

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: The cast words
**Scope:** `cast/gui.md`, `cast/terminal.md`, `cast/style.md`; then `generated/` through `cast cast/spell.md` from the JAM root.
**Action:**
1. `cast/gui.md`, `## EditorMode` (`:266-275`): add the row `| insert   | 6   |` after `panel`.
2. `cast/gui.md`: add one `@id` row for each new command, in the form of `:100`: `insertBefore`, `insertAfter`, `insertLineStart`, `insertLineEnd`, `openLineBelow`, `openLineAbove`, `leaveMode`, `deleteAction`, `change`, `deleteCharacter`, `pasteAfter`, `pasteBefore`, `splitLine`, `deleteBackward`, `undo`, `redo`. Skip a name that already has a row.
3. `cast/terminal.md`, `## TextEditorCommandId`: append 16 rows after `scrollHalfPageDown`, in the order of item 2, with the keys `0x4800022` to `0x4800031` and the value equal to the name.
4. Append one row for each new command, in the same order, to each of the four tables:

| Command | `textEditorKeys mac` and `windows` | `textEditorModes` | `textEditorLayers` |
|---|---|---|---|
| `insertBefore`, `insertAfter`, `insertLineStart`, `insertLineEnd`, `openLineBelow`, `openLineAbove` | `""` | `normal` | `""` |
| `leaveMode` | `` `escape` `` | `insert` | `""` |
| `deleteAction`, `change`, `deleteCharacter`, `pasteAfter`, `pasteBefore` | `""` | `normal visual` | `""` |
| `splitLine` | `` `return` `` | `insert` | `""` |
| `deleteBackward` | `` `backspace` `` | `insert` | `""` |
| `undo`, `redo` | `""` | `normal` | `""` |

5. `cast/style.md`: add `| textEditorStatusBarInsertColourId | 0x470000b | \`text editor status bar insert colour id\` |` after `:266`, in the column widths of that table.
6. Run `cast cast/spell.md` from the JAM root.

**Validation:**
- `map::EditorMode::insert` is 6. The rows 0 to 5 keep their keys.
- Each of the four tables has one row for each `TextEditorCommandId` row (the rule of `cast/terminal.md:1917-1920,2072-2073,2119-2120`).
- No default key of a new command is in mode `normal` or `visual` (closed item 2).
- New names and their families: the 16 command names (the old `map::EditorAction` names, old `.cpp:625-799`, and the old sheet names of EVE `ViewLayout.md`; `deleteAction` is the old enum name of the sheet word `delete`, a C++ keyword); `textEditorStatusBarInsertColourId` (`textEditorStatusBarNormalColourId`, `cast/style.md:262`).

### Step 2: The primitives and the undo actions
**Scope:** `jam_gui/widgets/jam_TextEditor.h`, `jam_gui/widgets/jam_TextEditor.cpp`.
**Action:**
1. Add the private member `juce::UndoManager undoManager;` (old `.h:547`).
2. Add the private structs `InsertLine` and `RemoveLine` at class scope, with the old declarations (old `.h:240-277`).
3. `InsertLine::InsertLine`: the old initializer list (old `.cpp:1035-1038`), and in the body `jassert (editor.getNumHistoryLines() == 0);`. Only a document block is edited; this one constructor is on the path of every line edit.
4. `InsertLine::perform`: the old body (old `.cpp:1040-1050`) on today's storage:
   - `editor.live.document->addChild (*editor.live.document->getRoot(), lineNumber > 0 ? editor.live.index->getElement (lineNumber - 1) : nullptr, line);`
   - `editor.live.index->insertLine (lineNumber, &line);`
   - `editor.setBlocks();` in place of the old `swap.addLine`, `setAdvances`, `setLineRows` (closed item 4).
   - `editor.setCursor (lineNumber, 0);`, `return true;`
5. `InsertLine::undo`: the old body (old `.cpp:1052-1067`) on today's storage:
   - `editor.live.document->removeChild (*editor.live.document->getRoot(), line);`
   - `editor.live.index->removeLine (lineNumber);`
   - `editor.setBlocks();` in place of the old swap calls and the old erase of the runs. `TextBlock::setRows` clears `runs` and `graphics` for the block (`jam_TextBlock.cpp:111-114`).
   - `editor.setCursor (lineNumber, 0);`, `return true;`
6. `RemoveLine`: the old bodies (old `.cpp:1069-1082`), with `*editorToUse.live.index->getElement (lineNumberToRemove)` in place of `history.index`.
7. `addLine (lineNumber, cells)`: the old body (old `.cpp:1084-1093`) with `live.document` in place of `historyDocument`. The hand-over is `undoManager.perform (std::make_unique<InsertLine> (*this, lineNumber, *line).release());`. `juce::UndoManager::perform` takes ownership of the action. `release` hands it over with no raw `new`, in the form of `jam_Window.cpp:44`.
8. `removeLine (lineNumber)`: `undoManager.perform (std::make_unique<RemoveLine> (*this, lineNumber).release());` (old `.cpp:1095-1098`).
9. `setLine`, `addText`: the old bodies (old `.cpp:1100-1137`), with `getCells` in place of `getOrCreateCells`. Every line of a live block holds its cells (`jam_TextEditor.cpp:1953-1956`).
10. `setCursor` (`.cpp:1958-1968`): the upper bound of the column is `cells.size()` in insert mode, and `cells.size() - 1` otherwise (old `.cpp:1450-1452`).

**Validation:**
- Each body equals its old body, except the storage names, the layout call, and the ownership hand-over.
- An `addLine` and its `undo` leave the document, the index, and the layout as before the `addLine`.
- New names: none. Each name is the old name.

### Step 3: The edit actions
**Scope:** `jam_gui/widgets/jam_TextEditor.cpp` (`getActions()`, `.cpp:375-560` today).
**Action:** Add the old helper lambdas and the old actions of the inventory into today's `getActions()`, verbatim, with these conformed forms only:
1. Each `actions.add<TextEditor&, int> (map::EditorAction::X, [captures] (TextEditor& editor, int count) {…})` becomes `actions.add (Id::X, [captures] (const juce::Identifier&, TextEditor& editor, int count) {…})`, the form of `.cpp:471`. `map::EditorAction::deleteAction` becomes `Id::deleteAction`.
2. Each read of the mode, `map::EditorMode::getInstance()->get (editor.state.getProperty (Id::mode).toString())`, becomes `editor.getMode()`. Each write, `editor.state.setProperty (Id::mode, …get (map::EditorMode::X), nullptr)`, becomes `editor.setMode (map::EditorMode::X)`.
3. `editor.getOrCreateCells (n)` becomes `editor.getCells (n)`. `editor.history.index->getNumLines()` becomes `editor.getNumLines()`.
4. `getMotions().get (map::EditorAction::right, editor, n)` becomes `getMotions().get (Id::right, editor, n)`; the same for `left`.
5. `deleteAction` and `change` call today's `runOperator (editor, count, Id::deleteAction, …)` and `runOperator (editor, count, Id::change, …)` (`.cpp:447-461`). `eraseSelection` calls today's `getStartColumn (editor)` (`.cpp:416-420`), which has no `isChanging` parameter. The old linewise change start, `std::numeric_limits<int>::max()` (old `.cpp:445`), becomes the `startColumn` of the change branch inside `eraseSelection`.
6. The old helper `paste` keeps its name. It is a local lambda; today's terminal action `Id::paste` (`.cpp:529-534`) is a separate entry.
7. `perform` (`.cpp:820-856`) and `getCommandInfo` (`.cpp:755-759`) do not change. `perform` finds an edit action by its name, as it finds each other action (`.cpp:828-842`).

**Validation:**
- Each restored lambda equals its old body, except the forms of items 1 to 6.
- `perform` and `getCommandInfo` equal their bodies of today. No line of the editor tests `isReadOnly()` to allow a command (closed item 3).
- New names: none.

### Step 4: Insert mode: keys, caret, status bar
**Scope:** `jam_gui/widgets/jam_TextEditor.cpp`, `jam_gui/widgets/jam_TextEditorStatusBar.h`, `jam_gui/widgets/jam_TextEditorStatusBar.cpp`.
**Action:**
1. `getKeyInputs()` (`.cpp:604-667`): add the entry for `map::EditorMode::insert`, in the form of the terminal entry (`.cpp:634-653`): `invokeKey` first; then a key with a text character of at least `Chars::space` calls `editor.insertTextAtCaret (juce::String::charToString (key.getTextCharacter()))` and returns `true` (the old typed insert, old `.cpp:873,942-949`). Any other key returns `false`.
2. `getCarets()` (`.cpp:669-719`): move `caretThickness` and `getBar` from the terminal lambda to the scope of the outer lambda, because two entries now use them (CODING.md, "narrowest scope that holds every use"). Add the insert caret: `carets.add (map::EditorMode::insert, [getCursorCell] (const int&, const TextEditor& editor) { return getBar (getCursorCell (editor)); });`. This is the old bar of width 2 (old `.cpp:1534-1535`; `.h:197`) and the pipe of `eve/SPEC.md:800-801`.
3. `TextEditorStatusBar.h` (`:27-36`): add `insertColourId = map::ColourId::textEditorStatusBarInsertColourId,` after `openColourId`.
4. `TextEditorStatusBar.cpp` (`:19-24`): add `{ map::EditorMode::insert, insertColourId },` to `modeColours`. The label is `INSERT`, from `get (mode).toUpperCase()` (`:36`; ruling 16). The bar shows in insert mode (`jam_TextEditor.cpp:902`).

**Validation:**
- In insert mode, a printable key adds its text at the caret, and `escape`, `return`, `backspace` run `leaveMode`, `splitLine`, `deleteBackward`.
- The insert caret is a bar of `caretThickness`. The terminal caret paints as before.
- New name and its family: `insertColourId` (`normalColourId`, `visualColourId`, `openColourId`, `jam_TextEditorStatusBar.h:31-33`).

### Step 5: `juce::TextInputTarget`
**Scope:** `jam_gui/widgets/jam_TextEditor.h`, `jam_gui/widgets/jam_TextEditor.cpp`.
**Action:**
1. Add `public juce::TextInputTarget` to the bases (`.h:107-112`), after `juce::ApplicationCommandTarget`.
2. Add two private helpers:
   - `int getCharIndex (juce::Point<int> cell) const;` — for each line above `cell.y`, `getCells (line).size() + 1`; then `cell.x`.
   - `juce::Point<int> getCellForCharIndex (int characterIndex) const;` — the inverse: the line where the running count passes `characterIndex`, and the column inside it.
3. The 11 overrides, public:
   - `isTextInputActive()`: `return not isReadOnly() and getMode() == map::EditorMode::insert;`. A peer gives text only to an active target (`juce_ComponentPeer.cpp:296-298`). EVE's editor is never in insert mode, so it never becomes a target.
   - `insertTextAtCaret (textToInsert)`: the non-linewise branch of the old `put` (old `.cpp:577-587`) at the cursor: `numLinesBefore`, `size`, `numLastCells { addText (cursor.y, cursor.x, textToInsert.toStdString()) }`, then `setCursor (cursor.y + getNumLines() - numLinesBefore, numLastCells - (size - cursor.x))`. A pasted text with newlines arrives whole and in one undo transaction.
   - `getCaretPosition()`: `getCharIndex (getCursor())`.
   - `getHighlightedRegion()`: `{ getCaretPosition(), getCaretPosition() }`. Insert mode has no selection.
   - `setHighlightedRegion (newRange)`: `const auto cell { getCellForCharIndex (newRange.getStart()) }; setCursor (cell.y, cell.x);`.
   - `setTemporaryUnderlining (underlinedRegions)`: an empty body. The editor paints no composition, in the form of the Linux peer's empty `textInputRequired` (`juce_Windowing_linux.cpp:372`).
   - `getTextInRange (range)`: for each line that the range covers, `AnsiDocument::getText (cells, columns)` of the covered columns, with a newline between lines.
   - `getTotalNumChars()`: `getCharIndex ({ getCells (getNumLines() - 1).size(), getNumLines() - 1 })`.
   - `getCaretRectangleForCharIndex (characterIndex)`: the cell of the index, `getLocalArea (&content, getCellBounds (cell.y, cell.x)).getSmallestIntegerContainer()`.
   - `getCharIndexForPoint (point)`: `getCharIndex (getCell (content.getLocalPoint (this, point).toFloat()))`.
   - `getTextBounds (textRange)`: a `juce::RectangleList<int>` with `getCaretRectangleForCharIndex (index)` for each index of the range.

**Validation:**
- The class implements each pure virtual of `juce_TextInputTarget.h:61-107`.
- `getTextInRange ({ 0, getTotalNumChars() })` gives the text of the document. The owner reads an edited text this way (Step 6).
- New names and their families: `getCharIndex`, `getCellForCharIndex` (JUCE `getCharIndexForPoint`, `getCaretRectangleForCharIndex`, `juce_TextInputTarget.h:89,99`; the editor's `getCell`, `.cpp:1112`).

### Step 6: The public entry of an editable document
**Scope:** `jam_gui/widgets/jam_TextEditor.h`.
**Action:** Move the declaration `void setLiveBlock (std::unique_ptr<Document> document);` (`.h:484`) to the public section, next to `setDocument` (`.h:196-202`). The body does not change (ruling 32).

The owner contract of an editable editor, for Parts 9 and 10:
1. Build `auto document { std::make_unique<AnsiDocument>() };`. Call `document->addLines (text.data(), size)` with the text and one final newline: the text after the last newline stays in the pen and is not a line (`jam_AnsiDocument.h:46-48`).
2. Call `editor.setLiveBlock (std::move (document))`. It sets the mode to `reader` and read-only to `true` (`.cpp:1148-1150`).
3. Call `editor.setReadOnly (false)`, then `editor.setDefaultMode (map::EditorMode::normal)` (Part 10 Step 3), then `editor.setMode (map::EditorMode::insert)` or `normal`.
4. Read the text with `getTextInRange ({ 0, getTotalNumChars() })` (Step 5).
5. The key sheet of the owner binds the normal-mode and visual-mode edit keys (closed item 2). It also names the `default mode` row with the key `escape` in the modes `visual open`. That row takes the place of the JAM default row, whose keys `escape i a I A o O` would run `defaultMode` in normal mode before an insert command (`cast/terminal.md:1963,2099`; Part 10 Step 7). The JAM default sheet registers only when the manager has no command (`.cpp:1132-1136`).
6. An owner that also shows a read-only editor builds its sheet from its state: the edit rows are in the sheet only while an editable editor shows (closed item 3). A read-only editor can enter its visual mode, and an edit row of the mode `visual` would apply there.

**Validation:** `setDocument<>` is unchanged. `printHelp` builds as before (`ling/Source/Help.cpp:19`).

### Step 7: Regenerate and build JAM
**Scope:** JAM through `ling`.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:** Build `ling` with `cast cast/spell.md --debug`. Run `ling --help`.
**Validation:** The build ends with exit code 0. The help output is byte-identical to its output before this part.

### Step 8: Check the edit machinery by hand
**Scope:** a run that ARCHITECT chooses. Part 10 is the first TIT user (the commit message box).
**Action:** With an editable editor from the Step 6 contract, check each restored action once: `i a I A o O`, typing, `return`, `backspace`, `escape`, `d c x` with a motion and in visual mode, `p P`, `u`, `ctrl+r`.
**Validation:** Each action gives the result of its old body. Undo restores the text and the cursor line.

### Step 9: Check EVE
**Scope:** `~/Documents/Poems/dev/eve`.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:** Build EVE. In NORMAL, press `i a I A o O`, `d c x p P u`, and `ctrl+r`.
**Validation:**
- `i a I A o O` return to TERMINAL (`eve/SPEC.md:349-351`): `defaultMode` is the first matching row (`jam_CommandManager.cpp:96-109`), and the new insert commands have no default key.
- `d c x p P u ctrl+r` change no history line (`eve/SPEC.md:736-737`): no default sheet row binds them.

## BLESSED Alignment

- **B — Bound.** The editor owns its `UndoManager`. The manager owns each action. An action refers to its editor, which outlives the manager that it owns.
- **L — Lean.** The restore takes the tested bodies. No new edit command exists beyond the old set.
- **E — Explicit.** The keys are data in the cast tables and in the owner's sheet. A read-only editor has no edit key, because no sheet that it can reach holds one (closed item 3).
- **S — SSOT.** The document of the live block is the one text. `TextInputTarget` reads it in place. The mode stays one row of the `terminal` tree.
- **S — Stateless.** The cursor, the anchor, and the mode stay rows of `state` and `terminal`. The undo history belongs to the `UndoManager`.
- **E — Encapsulation.** The owner gives text through `setLiveBlock` and `insertTextAtCaret`, and reads it through `getTextInRange`. It does not poke the document.
- **D — Deterministic.** The same keys on the same document give the same lines, cursor, and undo history.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. The five items of "Closed by a read in this part" are ORACLE's reads. ARCHITECT can reverse each one.
