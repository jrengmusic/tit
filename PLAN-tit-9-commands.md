# PLAN: TIT Part 9 — Commands and Keys

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part builds the command layer of TIT in the pattern of END and EVE (ruling 12).

1. **Static tables.** The command ids are a cast table, `## CommandId`. The default keys are the table `Source/keys.md`. Each key row has one mode and names one command id.
2. **One mode for each Operation menu (ruling 34).** JAM's `## EditorMode` gets eight rows. `TITView` holds the mode of the current Operation in its state. So `c` is a different row in each menu.
3. **A command chains.** One key maps to one command id. A command id runs one or more git command lines, or calls another command id (ruling 35; ARCHITECT, session 2026-11-13).
4. **The state builds the keys (rulings 55 and 56).** A table of rules over the `git` rows of Part 8 says which commands the state allows. A change of those rows is an event of the Model. The event builds the key sheet, and the sheet holds the row of a command only while its rule holds. So a key invokes only a command that the state allows. `perform` is the END body and reads no state. The command manager only invokes: `getCommandInfo` names the command, in the END form (tracker, "Command availability").
5. **The first user commands.** `push`, `pushAutoSync`, `finalizeMerge`, `abortMerge`, `rebaseContinue`, `rebaseAbort`. Each one needs no dialog, no text input, and no view. `TITGitProcessor` runs their git command lines and then runs the detection again.
6. **The layers (MANIFESTO E, MVP).** "The Model holds the state. The View composes. The Processor holds the business logic." `TITView` is the root command target in the END form: it lists every command id, and each entry of its `commands` map is a thin call. An entry names only a processor API, a view API, a Model row, or another command id. A function that makes a command line from the `git` rows is a member of `TITGitProcessor`. A function that reads or writes a file is a member of `TITGitProcessor` or `TITModel`. Each later part keeps this rule.

Each later part adds the command ids, the key rows, and the modes of the view that it builds, in the same tables.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 5, 6, 7, and 8. Step 1 changes JAM.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization, and parentheses for an initializer of the same container type; no `==` on strings; no non-owning naked pointer as a member; no anonymous namespace.
- CODING.md, "Uniform signature in the type": each map declares its signature.
- MANIFESTO L, YAGNI, and Part 1 Step 7: "No table holds a row that no code of this part reads". This part writes only the rows, ids, and modes that its own commands use.
- MANIFESTO S (Stateless): "State belongs exclusively to the Model." The mode and the running flag are rows of the Model.
- ARCHETYPE.md §1: no step edits a generated file by hand. Each cast table keeps its `+---` separator lines.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build and Git: agents build only on ARCHITECT's explicit request. The fixture repositories of Step 9 need git commands; ARCHITECT runs them, or gives the git instruction.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`.

### Rulings of this part, verbatim

- ARCHITECT: *"table driven static, cast manifest, mapped to CommandManager, this always the pattern. default config, user override"*.
- ARCHITECT: *"the mapping was never 1:1. one API could be one command id but one command id, could chain couple of API, or calling other commandId one keyboard shortcut can mapped into one commandId"*.
- **Ruling 34.** Question: "In the main menu, one key reaches a different git command by state. `c` is clone (NotRepo), commit (Dirty) or continue rebase (Rebasing). ... Where does that one command id come from?" Selected: *"One mode per Operation menu"*. Option text: "Each Operation menu (notRepo, normal, merging, rebasing, timeTraveling, conflicted, dirtyOperation, pinned) is its own key-sheet mode, as legacy builds one menu per Operation (menu.go:22-31). `c` is clone in notRepo, commit in normal, continue in rebasing. Inside normal, `[` `]` `r` chain by Timeline and Remote."
- **Ruling 35.** Question: "What is the command id of the `r` row?" Selected: *"addRemote, calls pullRebase"*. Option text: "Row `addRemote | r | menuNormal`. With no remote it adds a remote. With a remote it calls `pullRebase`." The question text also fixed the `]` row: "the row is `push`: it pushes when Ahead and calls `pushAutoSync` when Diverged".

### The pattern that this part follows

| Piece | Where |
|---|---|
| The command ids as a cast table: `name`, `key`, `value`; the key is the `juce::CommandID` | `eve/cast/identifiers.md:111-136`; `end/cast/bimaps.md:58-100` |
| The spell group of that table: `@code:bimap`, name `CommandId`, base `@bimapIdentifier`, `keyType: int`, `valueType: juce::Identifier` | `eve/cast/spell.md:126-136`; index row `@bimapIdentifier` = `jam::Bimap<int, juce::Identifier>` |
| The id blocks in use: `0x46…` EVE, `0x47…` END, `0x48…` `TextEditorCommandId`, `0x49…` the pane commands | `eve/cast/identifiers.md:123-135`; `end/cast/bimaps.md:70`; `jam/cast/terminal.md:1865-1911`; `jam/cast/gui.md:250-254` |
| The view lists every id and names it: `jam::CommandManager::addCommands (commands, *commandId);`, `jam::CommandManager::setInfo (result, commandId->get (commandID).toString());` | `end/Source/ui/ENDView.cpp:73-81`; `eve/Source/EVEView.cpp:440-448` |
| The members: `jam::SharedInstance<map::CommandId> commandId`, `jam::SharedInstance<jam::CommandManager> commandManager` | `end/Source/ui/ENDView.h:97,99` |
| `perform` looks the id up in a map: `const auto isPerformed { commands.contains (info.commandID) }; if (isPerformed) commands.get (info.commandID); return isPerformed;` | `ENDView.cpp:83-91` |
| The map and its entries: `jam::Function::Map<juce::CommandID, commandCapacity, void ()> commands;`, `commands.add (map::CommandId::reload, [this] (const juce::CommandID&) { … });`, in `register…Actions()` functions that `registerActions()` calls | `ENDView.h:257-259`; `ENDView.cpp:163-176,385-400` |
| One `register…Actions()` function for each family of commands: application, tab, zoom, mode, split, load plugin, close, join, swap, resize. No one of them is longer than 30 lines. | `ENDView.h:118-179`; `ENDView.cpp:163-375` |
| An entry is a thin call: it tells a child view (`tabView->join (direction)`, `tabView->getFocusedPane().close (command)`) or the owner of the logic (`config.loadFromPath()`, `nexus.createSession()`) | `ENDView.cpp:133-151,171-175,263-274,364-375` |
| A child view is its own class with its own row, its own `events` map, and its own API (`setZoom`, `setMode`, `close`) | `end/Source/ui/EditorView.h:25-28,99-114,194-218` |
| A command that calls another command: `commandManager->invokeDirectly (command->commandID, false);` | `eve/Source/EVEView.cpp:567`; `ENDView.cpp:150` |
| A command that chains by state: `close` writes the empty row when a markdown file is open; otherwise it invokes its command | `end/Source/ui/EditorView.h:110-116` |
| The view gives a key to the sheet with its own state: `return commandManager->invokeKey (key, state);` | `ENDView.cpp:55-58`; `EVEView.cpp:89-94` |
| The state of a view is a tree of the Model: `jam::Model::Component<ENDView> (… existing tree …)`; the mixin gives the member `state` | `ENDView.cpp:4`; `jam_data_structures/model/jam_Model.h:129-192` |
| A component state holds a mode while its key rows apply: `getOrCreateMarkdownState().setProperty (Id::mode, static_cast<int> (map::EditorMode::panel), nullptr);`; the row `dismiss \| cmd+w \| panel` | `EVEView.cpp:617`; `eve/Source/layout/default-eve.md:47`; `eve/SPEC.md:555-556` |
| A new mode is a row of JAM's `## EditorMode`: `panel` 5 for EVE; `insert` 6 in Part 6 | `jam/cast/gui.md:257-275`; `PLAN-tit-6-text-editor-insert.md`, Step 1 |
| One key, two command ids, kept apart by the mode: `j` is `down` in `normal visual` and `scrollDown` in `reader` | `jam/cast/terminal.md:1913,2068` |
| The next target: `return findFirstTargetParentComponent();` | `EVEView.cpp:495-498` |
| The view gives its focus to its default component | `EVEView.cpp:74-81` |
| The sheet build of TIT before this part: the JAM default rows, then the rows of `Source/keys.md`; every command target of the window registered; `setKeyPresses`. This part moves it into `TITView` (Step 7). | Part 7 Step 7, item 3 |
| The view registers itself with the manager in its constructor: `commandManager->registerAllCommandsForTarget (this);` | `end/Source/ui/ENDView.cpp:13` |
| The view registers each child target, and builds the sheet: the default rows that the config does not name, then the config rows | `eve/Source/EVEView.cpp:466-487` |
| An event of the Model replaces the key sheet: `events.add (jam::Format::toType (Id::keys), [this] (const juce::Identifier&, juce::ValueTree& tree) { commandManager->setKeyPresses (tree.getParent()); });` | `ENDView.cpp:419` |
| One event for several rows: `for (const auto& id : { … }) events.add (id, [this] (const juce::Identifier&, juce::ValueTree&) { setMouseConfig(); });` | `ENDView.cpp:408-412` |

### The key sheet facts

| Fact | Where |
|---|---|
| `setKeyPresses` splits a row value at spaces. Each token is one key description. So a description has no space. | `jam_gui/commands/jam_CommandManager.cpp:16,21` |
| A token of one or two characters matches the typed text. A longer token is a `juce::KeyPress::createFromDescription`. | `jam_CommandManager.cpp:23-24,96-109` |
| A mode cell holds names of `map::EditorMode`. A name that the map does not find adds no row. | `jam_CommandManager.cpp:31-37` |
| A row whose type matches no registered command is skipped | `jam_CommandManager.h:28-29`; `.cpp:14` |
| `getKey` returns the first row that matches the key, the mode, and the layer. `invokeKey` invokes only the command of a row of the sheet, so a command with no row has no key. | `jam_CommandManager.cpp:43-63,96-109` |
| `setKeyPresses` replaces the whole sheet. It keeps the `sequenceTimeout` child, which names no command. | `jam_CommandManager.cpp:5-41` |
| `getCommandForName (name)` gives the command info of a row type, or `nullptr` | `jam_CommandManager.h:67-75`; `.cpp:132-142` |
| A key goes to the focused component first, then to each parent | JUCE `juce_gui_basics/windows/juce_ComponentPeer.cpp:196-221` |
| The reader-mode editor gives a key to the sheet with its own `terminal` tree, then to its viewport | `jam_gui/widgets/jam_TextEditor.cpp:629-632` |
| The reader-mode default rows take `j`, `k`, `gg`, `G`, `ctrl+u`, `ctrl+d`, `v`, `V`, `ctrl+v`, `y`. They take none of `]`, `f`, `a`, `c`. | `jam/cast/terminal.md:1913-2063,2068-2113` |
| JUCE finds the target of a command from the focused component, through `getNextCommandTarget()`, to the application | JUCE `juce_gui_basics/commands/juce_ApplicationCommandManager.cpp:226-244`; `juce_ApplicationCommandTarget.cpp:94-132` |
| JUCE runs `perform` of the target for each invoke of a command whose flag is not disabled, and asserts when `perform` returns `false` for a command that the target listed | `juce_ApplicationCommandManager.cpp:192-212,239-240`; `juce_ApplicationCommandTarget.cpp:64-84` |
| `setInfo` clears the disabled flag of each command, so every listed command reaches `perform` | `jam_gui/commands/jam_CommandManager.h:78-85`; `end/Source/ui/ENDView.cpp:78-81` |

Part 7 wrote the `quit` value as `ctrl + C`. With the first fact, that is three tokens. `PLAN-tit-7-model.md` Step 2 and `PLAN-tit-4-key-input.md` Step 2 now say `ctrl+c`.

### The legacy commands of this part

| Command | Menu and state rule | Command lines | After the run | Where |
|---|---|---|---|---|
| `push`, key `]` | Normal; a remote; `Ahead`; `Clean` | `push --progress`; with no upstream `push --progress -u origin <branch>`. When the push fails: `pull --no-rebase --progress`, then the push again. | detection | `legacy/app/menu_items.go:55`; `menu_render_core.go:98-114`; `op_push.go:13-23`; `op_push_sync.go:14-21,77-91`; `handlers_commit.go:16` |
| `pushAutoSync`, key `]` | Normal; a remote; `Diverged`; `Clean` | `pull --no-rebase --progress`, then the push | detection | `menu_items.go:109`; `menu_render_core.go:142-145`; `op_push_sync.go:14-21,77-91`; `handlers_git_result.go:50` |
| `finalizeMerge`, key `f` | Merging | `commit --no-edit` | detection | `menu_items.go:215`; `menu_render_extra.go:119-121`; `dispatch_git_basic.go:126-130`; `op_rebase.go:60-61` |
| `abortMerge`, key `a` | Merging, Conflicted, DirtyOperation | `merge --abort`, then `reset --hard` | detection | `menu_items.go:223`; `menu_render_extra.go:121,137,145`; `op_merge.go:93-111`; `handlers_pull.go:58-63` |
| `rebaseContinue`, key `c` | Rebasing | `rebase --continue` | detection | `menu_items.go:231`; `menu_render_extra.go:128`; `op_rebase.go:13-14,29` |
| `rebaseAbort`, key `a` | Rebasing | `rebase --abort` | detection | `menu_items.go:239`; `menu_render_extra.go:129`; `op_rebase.go:18-19,45` |

- None of the six has a confirmation dialog or a text input (`legacy/app/dispatch_git_basic.go:11-18,46-52,126-155`).
- Legacy runs one menu for each Operation; `Rewinding` uses the Normal menu (`legacy/app/menu.go:22-31`). The menu function names are `menuNotRepo`, `menuNormal`, `menuMerging`, `menuConflicted`, `menuRebasing`, `menuTimeTraveling`, `menuDirtyOperation` (`menu_render_core.go:10,35`; `menu_render_extra.go:63,118,126,135,143`).
- Legacy sets `GIT_TERMINAL_PROMPT=0` for each command (`legacy/git/execute.go:26-27`).
- Legacy builds the menu from the state, and registers the key of an item only while the item is in the current menu (`legacy/app/app_keys.go:225-240`). No legacy code tests the state when a key arrives.

### The child process facts

| Fact | Where |
|---|---|
| `jam::Subprocess::launch` asserts that no worker exists. "Call `kill()` before a second `launch()`." `kill()` is safe with no process. | `jam_subprocess/subprocess/jam_Subprocess.h:26-30,95-104` |
| The completion callback runs on the worker thread of the child process | `jam_Subprocess.h:11-19` |
| The chunk callback is optional: "nullptr = capture-only mode" | `jam_Subprocess.h:93` |
| A child process that the pool thread starts and waits for: `getOutput` | Part 8 Step 7, item 9 |
| A processor writes a row with `storeValue`, from the thread that has the value | Part 8, "The patterns that this part follows"; tracker, ruling 57 |
| A flush on the message thread, by the processor, before the view reads: `state.flush();` | `eve/Source/EVETerminalProcessor.cpp:258-261`; `jam_data_structures/model/jam_Model.h:922-928` |
| A command line from a table cell: `jam::ChildProcess::getArguments (jam::Strings::getView (text))` | `ling/Source/CompilerManifest.cpp:64` |
| The child gets the environment of TIT. `jam::ChildProcess::start` takes no environment. On macOS the child also gets the terminal of TIT. | tracker, "The git side: child processes" |
| JUCE has `juce::SystemStats::getEnvironmentVariable` and no setter | doxygen `classjuce_1_1SystemStats`; no `setEnvironmentVariable` in `JUCE/docs/xml` |
| The detection is `setGit (directory)`, a function of the processor that runs on the pool thread | Part 8 Step 7, item 3 |
| A detection stores no row of the root | Part 8 Step 7, items 3 and 5 |

### Words the framework already holds

| Word | Where |
|---|---|
| `Id::mode`, `Id::directory` | `jam/cast/core.md:116`; Part 8 |
| `Id::git`, `Id::operation`, `Id::workingTree`, `Id::timeline`, `Id::remote`, `Id::currentBranch`, `Id::hasUpstream`; `map::Operation`, `map::WorkingTree`, `map::Timeline`, `map::Remote` | Part 8 Step 2 |
| `text::Git`, `cast/git.md` | Part 8 Steps 3, 4 |

### Closed by a read

ARCHITECT can reverse each one.

1. **This part writes only the rows that its commands use.** MANIFESTO L, YAGNI: "A parameter nobody passes yet." Part 1 Step 7: "No table holds a row that no code of this part reads". A key row with no registered command is skipped (`jam_CommandManager.h:28-29`), so it is a row with no reader. Ruling 33 says "Part 9 adds the other rows"; the rows of this part are the next rows of that sheet, and each view part adds its own. The tracker Status rows say this now.
2. **The mode names are the legacy menu function names:** `menuNotRepo`, `menuNormal`, `menuMerging`, `menuConflicted`, `menuRebasing`, `menuTimeTraveling`, `menuDirtyOperation`, and `menuPinned` for the pinned menu of Part 2 Step 4. The name `normal` is already a row of `## EditorMode`. `Rewinding` has the mode `menuNormal`, as in legacy.
3. **The state of `TITView` is the root tree of `TITModel`.** The root holds `mode`, and the `layer` and `sequence` rows that `invokeKey` writes (`jam_CommandManager.cpp:111-115`). A detection changes no root property. `Id::window` is a word of `jam/cast/web.md`, a module that TIT does not link, so this part adds no `window` child.
4. **`TITView` is the content of the window and holds the help editor of Part 7.** A key reaches the sheet with a TIT mode only through a component with that state (`ENDView.cpp:55-58`). The family is `ENDView`, `EVEView`. Part 10 puts the shell views in `TITView`.
5. **The mode follows the `operation` row.** In this part the menu is the only view. A later part that opens another view sets the mode of that view.
6. **The run keeps no output.** No code of this part reads it. Part 10 builds the console and adds the chunk path with `--progress` in view. The command lines keep the legacy text.
7. **`GIT_TERMINAL_PROMPT=0` is set once, for the process of TIT.** Each child gets the environment of TIT, and `jam::ChildProcess` takes none. The setter is `setenv` on macOS and `_putenv_s` on Windows; both come with the C library that `JuceHeader.h` already includes.
8. **A run blocks every TIT command.** The `isRunning` row is an event: while it is 1, the sheet holds no row of a command that has a rule. `run` stores the row and flushes the Model in the same call, on the message thread, so the sheet changes before the next key arrives. Legacy refuses input during an operation (`legacy/app/handlers_global_keys.go:60-63`).
9. **The normal-mode and visual-mode edit keys are not in this part.** No editable editor exists yet. Part 10 builds the first one, the commit box, and binds them (Part 6, closed item 2). The tracker rows of Parts 9 and 10 say this now.
10. **The user override of the key sheet is in Part 13**, with the other tables of `~/.config/tit/` (Part 7, Closed by a read, item 1). It is the JFS shape of rulings 44 to 47, through `jam::ConfigDocument::getOrCreateFileDocument (file, defaultDocument, getDefaultText)` (`jam_markdown/document/jam_ConfigDocument.cpp:1014-1029`): the disk file wins when it parses, and the carried default replaces a file of an older version.
11. **The command id block is `0x4a…`.** It is the next block after `0x49…`.

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: JAM, the menu modes
**Scope:** `jam/cast/gui.md`, table `## EditorMode`; then `jam/generated/` through `cast cast/spell.md` from the JAM root.
**Action:**
1. After the `insert` row of Part 6 Step 1, add eight rows in the column widths of the table: `menuNotRepo` 7, `menuNormal` 8, `menuMerging` 9, `menuConflicted` 10, `menuRebasing` 11, `menuTimeTraveling` 12, `menuDirtyOperation` 13, `menuPinned` 14.
2. Run `cast cast/spell.md` from the JAM root.

**Validation:**
- `map::EditorMode` holds the eight names with the keys 7 to 14. The rows 0 to 6 keep their keys.
- No other JAM file changes. `getKeyInputs()` and `getCarets()` get no entry: no editor takes a menu mode.
- Each name is in Closed by a read, item 2.

### Step 2: The command ids and the running row
**Scope:** `tit/cast/identifiers.md`, `tit/cast/spell.md`.
**Action:**
1. `cast/identifiers.md`, table `## identifiers`: after the last row of Part 8 Step 2, add a row of type `@id`, name `isRunning`, value `` `isRunning` ``, with an empty description (family `isRecording`, `eve/Source/EVEView.cpp:563`).
2. `cast/identifiers.md`: after the last table of Part 8, add the table `## CommandId` in the form of `eve/cast/identifiers.md:111-136`. The brief is the brief of that table with `TIT` in place of `EVE`. The rows:

   | name | key | value |
   |---|---|---|
   | `push` | `0x4a00001` | `` `push` `` |
   | `pushAutoSync` | `0x4a00002` | `` `pushAutoSync` `` |
   | `finalizeMerge` | `0x4a00003` | `` `finalizeMerge` `` |
   | `abortMerge` | `0x4a00004` | `` `abortMerge` `` |
   | `rebaseContinue` | `0x4a00005` | `` `rebaseContinue` `` |
   | `rebaseAbort` | `0x4a00006` | `` `rebaseAbort` `` |

3. `cast/spell.md`, `## index`: after the `@bimap` row of Part 8 Step 4, add `| @bimapIdentifier | jam::Bimap<int, juce::Identifier> |`, the row of the EVE index.
4. `cast/spell.md`, `## output`: after the five groups of Part 8 Step 4, item 3, add the group of `eve/cast/spell.md:126-136` with no change.

**Validation:**
- Each name is the camel-case form of its legacy id (`push`, `push_auto_sync`, `finalize_merge`, `abort_merge`, `rebase_continue`, `rebase_abort`; `legacy/app/dispatchers.go:10-45`).
- No JAM table holds a key in the block `0x4a…`.
- After the build of Step 9, `Identifiers.h` holds `map::CommandId` with six entries and an `enum value : int`.

### Step 3: The command lines
**Scope:** `tit/cast/git.md`.
**Action:** After the last row of Part 8 Step 3, add these rows:

| name | value | description |
|---|---|---|
| `push` | `` `git push --progress` `` | Sends the commits of the branch to its upstream. |
| `pushUpstream` | `` `git push --progress -u origin` `` | Sends the commits of the next branch to origin and sets its upstream. |
| `pullMerge` | `` `git pull --no-rebase --progress` `` | Fetches the upstream and merges it. |
| `commitNoEdit` | `` `git commit --no-edit` `` | Makes the merge commit with its prepared message. |
| `mergeAbort` | `` `git merge --abort` `` | Stops a merge and restores the tree. |
| `resetHard` | `` `git reset --hard` `` | Sets the tree and the index to HEAD. |
| `rebaseContinue` | `` `git rebase --continue` `` | Continues a rebase after its conflicts are resolved. |
| `rebaseAbort` | `` `git rebase --abort` `` | Stops a rebase and restores the branch. |
| `terminalPromptVariable` | `` `GIT_TERMINAL_PROMPT` `` | The environment variable that stops a credential prompt on the terminal. |
| `terminalPromptValue` | `` `0` `` | The value of terminalPromptVariable that stops the prompt. |

**Validation:** Each value equals its legacy literal (inventory, "The legacy commands of this part"; `legacy/git/execute.go:26-27`).

### Step 4: The key rows
**Scope:** `tit/Source/keys.md`.
**Action:** After the `quit` row of Part 7 Step 2, add these rows. Widen the column separators to the new cell widths.

| key | type | value | mode | description |
|---|---|---|---|---|
| `push` | `string` | `]` | `menuNormal` | Push to remote. |
| `finalize merge` | `string` | `f` | `menuMerging` | Finalize merge. |
| `abort merge` | `string` | `a` | `menuMerging menuConflicted menuDirtyOperation` | Abort merge. |
| `rebase continue` | `string` | `c` | `menuRebasing` | Continue rebase. |
| `rebase abort` | `string` | `a` | `menuRebasing` | Abort rebase. |

- A row key and a command name match as `Format::toType` forms, as `zoom in` and `zoomIn` do (`eve/Source/layout/default-eve.md:49`; `jam_CommandManager.h:28-31`).
- `pushAutoSync` has no row. `push` calls it (ruling 35, question text).
- Each description is the legacy label (`legacy/app/menu_items.go:55,215,223,231,239`).
- In one mode, each key has one row.

**Validation:** No value holds a space inside one key. Each mode name is a row of Step 1.

### Step 5: `TITModel` and `TITGitProcessor`, the run
**Scope:** `tit/Source/TITModel.cpp`, `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:**
1. `TITModel.cpp`, in the constructor body, after the `directory` row of Part 8 Step 5: `createAndAddParameter<jam::Parameter<int>> (state, Id::isRunning, 0);`. The row is a parameter of the root, as every row that a processor writes (ruling 57).
2. `TITGitProcessor.h`, public section, after `getRepositoryDirectory`: `static juce::String getPushLine (const juce::ValueTree& git);`, then `void run (const juce::StringArray& commandLines, const juce::StringArray& fallbackCommandLines);`. Private section, after the two `storeValue` functions: `static bool run (const juce::File& directory, const juce::StringArray& commandLines);`. Before the member `gitTag`: `const juce::Identifier titTag { jam::Format::toType (Id::tit) };`.
3. `TITGitProcessor.cpp`:

```cpp
void TITGitProcessor::run (const juce::StringArray& commandLines, const juce::StringArray& fallbackCommandLines)
{
    model.storeValue (titTag, Id::isRunning, 1);
    model.flush();

    pool.addJob ([this, commandLines, fallbackCommandLines, directory = juce::File { model.state[Id::directory].toString() }]
    {
        if (not run (directory, commandLines))
            run (directory, fallbackCommandLines);

        model.storeValue (titTag, Id::isRunning, 0);
        setGit (directory);
    });
}

bool TITGitProcessor::run (const juce::File& directory, const juce::StringArray& commandLines)
{
    return std::all_of (commandLines.begin(), commandLines.end(), [&directory] (const juce::String& commandLine)
    {
        return getOutput (directory, jam::ChildProcess::getArguments (jam::Strings::getView (commandLine))).has_value();
    });
}
```

   Then the push line:

```cpp
juce::String TITGitProcessor::getPushLine (const juce::ValueTree& git)
{
    if (static_cast<bool> (git.getProperty (Id::hasUpstream))) return text::Git::push;
    return text::Git::pushUpstream + juce::String { Chars::space } + git.getProperty (Id::currentBranch).toString();
}
```

   - `getPushLine` is the legacy choice of the push line (`legacy/app/op_push.go:15-23`; `op_push_sync.go:82-91`). It is a static member of the processor, as `getRepositoryDirectory` is (Part 8 Step 6): the choice of a command line from the `git` rows is business logic (Overview, item 6).
   - `Chars::space` is the JAM word of `ling/Source/CompilerManifest.cpp:174`.
   - The view tells the public `run`, on the message thread. It stores the row and flushes the Model, so the tree holds `isRunning` 1 and the views react before `run` returns (Closed by a read, item 8). The flush is the processor's flush of EVE (`eve/Source/EVETerminalProcessor.cpp:258-261`).
   - The run is one job of the pool. The pool has one thread, so a run and a detection never run at the same time.
   - The private `run` runs the lines in order on the pool thread, each one with `getOutput` of Part 8. `std::all_of` stops at the first line that fails. The first failure starts the fallback lines one time. A failure of the fallback ends the run.
   - The end of a run clears the row and runs the detection again, in the same job (Closed by a read, items 6 and 8). The pool thread writes the Model only with `storeValue`.
   - The copies of the two arrays into the job are the ownership handoff to the pool thread (MANIFESTO, Permitted copy 2).

**Validation:**
- No bail-out guard. No function body is longer than 30 lines.
- The processor keeps no run state. The pending lines live in the job of the run, and the flag is a parameter of the Model.
- The processor writes no tree: no `setProperty`, no `sendPropertyChangeMessage`, and no `juce::MessageManager::callAsync`.
- New names and their families: `run` (the legacy `Execute`, `legacy/git/execute.go:23`; the private overload is the same verb for the lines of one list); `titTag` (`gitTag`, Part 8 Step 6); `commandLines`, `fallbackCommandLines` (`commandLine`, `jam_subprocess/subprocess/jam_ChildProcess.h:61-74`); `getPushLine` (`getOpenCommand`, `eve/Source/EVEView.cpp:538`, with the noun `commandLine`).

### Step 6: `TITView`, the header
**Scope:** `tit/Source/TITView.h` (new).
**Action:** `#pragma once`, `#include <JuceHeader.h>`, `#include "generated/Generated.h"`, `#include "TITModel.h"`, `#include "TITGitProcessor.h"`. Declare:

```cpp
class TITView : public juce::Component,
                public jam::Model::Component<TITView>,
                public juce::ApplicationCommandTarget,
                private juce::ValueTree::Listener
{
public:
    TITView (TITModel& modelToUse, TITGitProcessor& gitProcessorToUse, juce::LookAndFeel& lookAndFeelToUse);
    ~TITView() override;

    void resized() override;
    void focusGained (FocusChangeType cause) override;
    bool keyPressed (const juce::KeyPress& key) override;

    juce::ApplicationCommandTarget* getNextCommandTarget() override;
    void getAllCommands (juce::Array<juce::CommandID>& commands) override;
    void getCommandInfo (juce::CommandID commandID, juce::ApplicationCommandInfo& result) override;
    bool perform (const InvocationInfo& info) override;

private:
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override;
    void setMenuMode (const juce::ValueTree& git);
    void setKeys (const juce::ValueTree& git);
    void registerAvailability();
    void registerActions();
    void registerPushActions();
    void registerOperationActions();

    bool isAvailable (juce::CommandID command, const juce::ValueTree& git) const;

    TITModel& model;
    TITGitProcessor& gitProcessor;
    jam::SharedInstance<map::CommandId> commandId { std::in_place };
    jam::SharedInstance<jam::CommandManager> commandManager;
    jam::Function::Map<juce::CommandID, jam::Function::pointerCapacity, bool (const juce::ValueTree&)> availability;
    jam::Function::Map<juce::CommandID, jam::Function::pointerCapacity, void ()> commands;
    jam::TextEditor help;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITView)
};
```

- `commandManager` takes the shared manager that the application created (Part 7 Step 6; `end/Source/ui/ENDView.h:99`).
- The two references are the form of `model` in `TITGitProcessor` (Part 8 Step 6). The application owns the three objects.
- Each entry of the two maps captures `this` only, so the capacity is `jam::Function::pointerCapacity`, as in the event map of Part 7.

**Validation:**
- New names and their families: `TITView` (`ENDView`, `EVEView`); `modelToUse`, `gitProcessorToUse`, `lookAndFeelToUse` (Part 7 Step 4; `eve/Source/EVEView.cpp:346`); `commandId`, `commandManager`, `commands`, `registerActions` (`ENDView.h:97,99,259`; `ENDView.cpp:385`); `registerPushActions`, `registerOperationActions` (the `register…Actions()` family, `ENDView.h:121-179`, with the noun of the command `push` and of the row `operation`); `availability`, `registerAvailability` (tracker, "Command availability"; NAMES.md Rule 8, a noun for what the table holds); `isAvailable` (the table `availability`; NAMES.md Rule 1, `is`); `setKeys` (`setKeyPresses`, `jam_gui/commands/jam_CommandManager.h:42`; the sheet `keys`); `setMenuMode` (`setMode`, `end/Source/ui/EditorView.h:105`, with the noun of ruling 34); `help` (Part 7 Step 7); `gitProcessor` (Part 8 Step 8).
- No member holds state. The mode and the running flag are rows of `state`.

### Step 7: `TITView`, the source
**Scope:** `tit/Source/TITView.cpp` (new).
**Action:** `#include "TITView.h"`. Then define each function.

1. The constructor and the destructor:

```cpp
TITView::TITView (TITModel& modelToUse, TITGitProcessor& gitProcessorToUse, juce::LookAndFeel& lookAndFeelToUse)
    : jam::Model::Component<TITView> (modelToUse.state)
    , model (modelToUse)
    , gitProcessor (gitProcessorToUse)
{
    const BinaryData::Raw helpText { files::titcHelp };

    setLookAndFeel (&lookAndFeelToUse);
    setWantsKeyboardFocus (true);
    addAndMakeVisible (help);
    help.setDocument (std::move (*jam::MarkdownDocument::parse (std::string_view (helpText.data, static_cast<size_t> (helpText.size)), files::titcHelp)));

    registerAvailability();
    registerActions();

    jam::Component::applyFunctionRecursively (this, [this] (juce::Component* child)
    {
        commandManager->registerAllCommandsForTarget (dynamic_cast<juce::ApplicationCommandTarget*> (child));
    });

    commandManager->registerAllCommandsForTarget (this);

    setMenuMode (model.state.getChildWithName (jam::Format::toType (Id::git)));
    setKeys (model.state.getChildWithName (jam::Format::toType (Id::git)));
    model.addListener (this);
}

TITView::~TITView()
{
    model.removeListener (this);
    setLookAndFeel (nullptr);
}
```

   - The editor has the LookAndFeel of its parent before its document, as Part 7 Step 7 requires (`ling/Source/Help.cpp:17-19`).
   - The view registers its child targets and itself, in the forms of `eve/Source/EVEView.cpp:479-484` and `end/Source/ui/ENDView.cpp:13`. The registration runs one time. The sheet is built after it, because `setKeyPresses` skips a row of a command that is not registered (`jam_CommandManager.h:28-29`).
   - The first sheet comes from the rows that the Model holds at construction. Each later sheet comes from an event (item 3).

2. The component functions, in the forms of `EVEView.cpp:74-81,495-498` and `ENDView.cpp:55-58`:

```cpp
void TITView::resized()
{
    help.setBounds (getLocalBounds());
}

void TITView::focusGained (FocusChangeType)
{
    help.grabKeyboardFocus();
}

bool TITView::keyPressed (const juce::KeyPress& key)
{
    return commandManager->invokeKey (key, state, false);
}

juce::ApplicationCommandTarget* TITView::getNextCommandTarget()
{
    return findFirstTargetParentComponent();
}
```

3. The events: the mode (ruling 34) and the keys (ruling 56):

```cpp
void TITView::valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property)
{
    const auto isGit { tree.hasType (jam::Format::toType (Id::git)) };

    if (isGit and property == Id::operation)
        setMenuMode (tree);

    if (isGit or property == Id::isRunning)
        setKeys (state.getChildWithName (jam::Format::toType (Id::git)));
}

void TITView::setKeys (const juce::ValueTree& git)
{
    const auto keysSheet { jam::Format::toType (Id::keys) };
    const auto config { jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcKeys })->getValueTree (jam::Format::toType (Id::config)).getChildWithName (keysSheet) };
    juce::ValueTree keys { keysSheet };

    for (const auto& defaultKey : jam::TextEditor::getDefaultKeys())
        if (not config.getChildWithName (jam::Format::toType (defaultKey.getType())).isValid())
            keys.appendChild (defaultKey.createCopy(), nullptr);

    for (const auto& configKey : config)
    {
        const auto* command { commandManager->getCommandForName (configKey.getType()) };

        if (command == nullptr or isAvailable (command->commandID, git))
            keys.appendChild (configKey.createCopy(), nullptr);
    }

    commandManager->setKeyPresses (keys);
}

bool TITView::isAvailable (juce::CommandID command, const juce::ValueTree& git) const
{
    return not availability.contains (command)
           or (not static_cast<bool> (state.getProperty (Id::isRunning)) and availability.get (command, git));
}

void TITView::setMenuMode (const juce::ValueTree& git)
{
    static constexpr std::array menuModes {
        map::EditorMode::menuNotRepo,
        map::EditorMode::menuNormal,
        map::EditorMode::menuMerging,
        map::EditorMode::menuConflicted,
        map::EditorMode::menuRebasing,
        map::EditorMode::menuTimeTraveling,
        map::EditorMode::menuDirtyOperation,
        map::EditorMode::menuNormal,
        map::EditorMode::menuPinned,
    };

    state.setProperty (Id::mode, static_cast<int> (menuModes.at (static_cast<size_t> (static_cast<int> (git.getProperty (Id::operation))))), nullptr);
}
```

   - The table has one entry for each key of `map::Operation`, in the order of Part 8 Step 2: 0 to 8. The entry of `rewinding` is `menuNormal` (`legacy/app/menu.go:30`). The listener reads the tree that the event delivered (CODING.md, "Consume the event payload").
   - `setKeys` is the sheet build of `eve/Source/EVEView.cpp:468-477` with one condition more: a config row of a command that the state does not allow is not copied. A row that names no command, such as `sequenceTimeout`, is always copied. A command with no rule is always allowed: the `mode` cell of its row is its gate.
   - The sheet is a complete state. Each event replaces it whole (`jam_CommandManager.cpp:5-8`; MANIFESTO E, "A state update is not mutation").
   - `isAvailable` is the one reader of the rule table. `setKeys` calls it here, and the menu build of Part 10 calls it.
   - A run is an event too: the `isRunning` row changes, and the next sheet holds no row of a command with a rule (Closed by a read, item 8).

4. The command target, in the forms of `ENDView.cpp:73-91`:

```cpp
void TITView::getAllCommands (juce::Array<juce::CommandID>& commands)
{
    jam::CommandManager::addCommands (commands, *commandId);
}

void TITView::getCommandInfo (juce::CommandID commandID, juce::ApplicationCommandInfo& result)
{
    jam::CommandManager::setInfo (result, commandId->get (commandID).toString());
}

bool TITView::perform (const InvocationInfo& info)
{
    const auto isPerformed { commands.contains (info.commandID) };

    if (isPerformed)
        commands.get (info.commandID);

    return isPerformed;
}
```

   - The three bodies are the END bodies with no change (`ENDView.cpp:73-91`). The command manager decides nothing (ruling 55), and `perform` reads no state (ruling 56).

5. `registerAvailability()`. Each entry has the form `availability.add (map::CommandId::<name>, [] (const juce::CommandID&, const juce::ValueTree& git) { return <rule>; });`. The rules, from the inventory:

   | Command | Rule |
   |---|---|
   | `push` | `remote` is `map::Remote::hasRemote`, and `workingTree` is `map::WorkingTree::clean`, and `timeline` is `map::Timeline::ahead` or `map::Timeline::diverged` |
   | `finalizeMerge` | `operation` is `map::Operation::merging` |
   | `abortMerge` | `operation` is `merging`, `conflicted`, or `dirtyOperation` |
   | `rebaseContinue` | `operation` is `map::Operation::rebasing` |
   | `rebaseAbort` | `operation` is `rebasing` |

   Each read is `static_cast<int> (git.getProperty (Id::<row>)) == map::<Table>::<value>`. The first rule covers the two states of the `]` row. `pushAutoSync` has no entry: it has no key row, and only `push` calls it.

6. `registerActions()` and its two families, in the END form (`ENDView.cpp:163-176,385-400`):

```cpp
void TITView::registerActions()
{
    registerPushActions();
    registerOperationActions();
}

void TITView::registerPushActions()
{
    commands.add (map::CommandId::push, [this] (const juce::CommandID&)
    {
        const auto git { model.state.getChildWithName (jam::Format::toType (Id::git)) };

        if (static_cast<int> (git.getProperty (Id::timeline)) == map::Timeline::ahead)
            gitProcessor.run ({ TITGitProcessor::getPushLine (git) }, { text::Git::pullMerge, TITGitProcessor::getPushLine (git) });
        else
            commandManager->invokeDirectly (map::CommandId::pushAutoSync, false);
    });

    commands.add (map::CommandId::pushAutoSync, [this] (const juce::CommandID&)
    {
        gitProcessor.run ({ text::Git::pullMerge, TITGitProcessor::getPushLine (model.state.getChildWithName (jam::Format::toType (Id::git))) }, {});
    });
}

void TITView::registerOperationActions()
{
    commands.add (map::CommandId::finalizeMerge, [this] (const juce::CommandID&) { gitProcessor.run ({ text::Git::commitNoEdit }, {}); });
    commands.add (map::CommandId::abortMerge, [this] (const juce::CommandID&) { gitProcessor.run ({ text::Git::mergeAbort, text::Git::resetHard }, {}); });
    commands.add (map::CommandId::rebaseContinue, [this] (const juce::CommandID&) { gitProcessor.run ({ text::Git::rebaseContinue }, {}); });
    commands.add (map::CommandId::rebaseAbort, [this] (const juce::CommandID&) { gitProcessor.run ({ text::Git::rebaseAbort }, {}); });
}
```

   - `push` is one command id for the key `]`. It pushes when `Ahead`, and calls `pushAutoSync` when `Diverged` (ruling 35, question text). Its fallback lines are the legacy answer to a failed push (`legacy/app/op_push.go:13-23`; `op_push_sync.go:14-21`).
   - `registerActions()` calls one function for each family of commands, and holds no entry itself. Each later part adds its families here. No family is longer than 30 lines (MANIFESTO L), as in END.
   - `registerOperationActions()` holds the commands of an operation in progress: the merge and the rebase.
   - Each entry chains rows of `text::Git` through one processor API, `run` (ARCHITECT: *"one command id, could chain couple of API, or calling other commandId"*). No entry makes a command line itself: the push line comes from `TITGitProcessor::getPushLine` (Step 5).

**Validation:**
- No function has a bail-out guard. Each early return is a result return. No `if` chain has more than three branches.
- No `==` on strings. Each compare is an Identifier compare or an integer compare.
- `perform`, `getAllCommands`, and `getCommandInfo` equal the END bodies. No callback and no line of `perform` tests if a command is allowed.
- The rule table has one reader, `isAvailable`, and `isAvailable` runs only inside the build of an event. `commands` holds six ids, and `availability` holds the five ids that have a key row.
- `getCommandInfo` sets no flag. No line of this part calls `setActive`.
- New local names: `helpText` (`helpText`, `ling/Source/Help.cpp:19`); `menuModes` (NAMES.md Rule 8, the modes of ruling 34); `git` (Part 8 Step 5); `isPerformed` (`ENDView.cpp:85`); `isGit` (the tree `git`; NAMES.md Rule 1); `keysSheet`, `config`, `keys`, `defaultKey`, `configKey` (`EVEView.cpp:468-476`); `child` (`EVEView.cpp:479`); `command` (`EVEView.cpp:562`).

### Step 8: The application
**Scope:** `tit/Source/Main.cpp`.
**Action:**
1. After `#include "TITGitProcessor.h"`, add `#include "TITView.h"`.
2. In `runWindow` of Part 7 Step 7 and Part 8 Step 8, make these changes. The other lines stay.
   - At the start of the body, add:

```cpp
        static constexpr int shouldOverwrite { 1 };

       #if JUCE_WINDOWS
        _putenv_s (text::Git::terminalPromptVariable, text::Git::terminalPromptValue);
       #else
        setenv (text::Git::terminalPromptVariable, text::Git::terminalPromptValue, shouldOverwrite);
       #endif
```

   - Move the line `gitProcessor = std::make_unique<TITGitProcessor> (model);` of Part 8 Step 8 to the line after `processor = std::make_unique<TITTerminalProcessor> (model);`.
   - Delete the line `const BinaryData::Raw help { files::titcHelp };`, the line `auto editor { std::make_unique<jam::TextEditor>() };`, and the two `editor->` lines. `TITView` does this work now.
   - Replace `window = std::make_unique<jam::TerminalWindow> (std::move (editor), ProjectInfo::projectName);` with `window = std::make_unique<jam::TerminalWindow> (std::make_unique<TITView> (model, *gitProcessor, lookAndFeel), ProjectInfo::projectName);`.
3. Delete first: `addCommands()` of Part 7 Step 7, item 3, goes, and so does its call in `runWindow`. `TITView` builds the sheet now (Step 7, item 3). Before the line `window = …`, add `commandManager->registerAllCommandsForTarget (this);`, the one statement of the old function that the application keeps. It registers `quit` (`juce_Application.cpp:66-81`) before the view builds the first sheet.

**Validation:**
- `gitProcessor` exists before `TITView` and is destroyed after the window: `window` is declared after `gitProcessor` (Part 7 Step 6; Part 8 Step 8).
- The `directory` row is still set last, so the first detection starts after the listeners exist.
- `runWindow` has no bail-out guard. The command-line runs of Part 1 do not change.
- New name and its family: `shouldOverwrite` (NAMES.md Rule 1, `shouldPromptLicense`; the third argument of `setenv`).

### Step 9: Build, run, and measure on macOS
**Scope:** the `tit/` project and JAM; the fixture repositories of Part 8 Step 9.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build). The fixture commands are git commands: ARCHITECT runs them, or gives the git instruction (CAROL.md Git).
**Action:**
1. Add the temporary log of Part 8 Step 9, item 1, with one more block in `valueTreePropertyChanged` of `Main.cpp`:

```cpp
        if (tree == model.state)
            jam::debug::Log::write (property.toString(), tree[property].toString());
```

2. Run `cast cast/spell.md --debug` from `tit/`.
3. For each row below, start `titc` in END in the fixture, press the keys, wait for the last expected line, press Ctrl+C, and read `~/titc.log`. The fixture numbers are those of Part 8 Step 9.

   | Fixture | Keys | Expected lines |
   |---|---|---|
   | 1, an empty folder | `]` | no `mode` line, because `TITView` sets `mode 7` before the Application listener exists; no `isRunning` line |
   | 4, one commit, no remote | `]` | `mode 8`; no `isRunning` line |
   | 5, `Ahead` by one commit | `]` | `mode 8`; `isRunning 1`, then `isRunning 0`; then `timeline 1` and `commitsAhead 0`. `git log origin/HEAD` in the fixture shows the commit. |
   | 6, `Diverged` | `]` | `isRunning 1`, then `isRunning 0`; then `timeline 1`. The log of the fixture shows a merge commit. |
   | 10, a merge in progress | `f` | `mode 9`; `isRunning 1`, `isRunning 0`; then `operation 1` and `mode 8` |
   | 10 again | `a` | `mode 9`; then `operation 1` and `mode 8`; the tree is as before the merge |
   | 11, a merge conflict | `f`, then `a` | `mode 10`; `f` gives no `isRunning` line; `a` gives `operation 1` and `mode 8` |
   | 12, a rebase in progress | `a` | `mode 11`; then `operation 1` and `mode 8` |
   | 13, a detached HEAD | `]`, `a`, `c`, `f` | `mode 12`; no `isRunning` line |
   | 16, a pin marker | `]` | `mode 14`; no `isRunning` line |
   | 18, a dirty operation marker | `a` | `mode 13`; `isRunning 1`, `isRunning 0` |

4. In fixture 5 with a new commit, press `]` two times with no pause. The log shows one `isRunning 1` line.
5. Remove the temporary log (Part 8 Step 9, item 4, and the block of item 1). Run `cast cast/spell.md --debug` a second time.

**Validation:**
- The two builds end with exit code 0. `Identifiers.h` holds `map::CommandId` and `Id::isRunning`. `Text.h` holds the ten new constants of `struct Git`.
- Each row gives its expected lines. The guide scrolls with its keys during each run, and Ctrl+C quits (Part 7 Step 8).
- The checks of Part 6 Step 9 on EVE still hold: the new rows of `## EditorMode` change no key of EVE.
- After item 5, `Main.cpp` holds no `jam::debug::Log` line (CODING.md, DIAGNOSTICS). The second `cast` run changes no generated file.

### Step 10: Run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 9 in END on a Windows machine with fixtures 1, 5, and 10.
**Validation:** The lines of each fixture match Step 9. `_putenv_s` builds with no new include.

## BLESSED Alignment

- **B — Bound.** The application owns the model, the git processor, and the window. The window owns `TITView`. `TITView` is destroyed before the processor and the model that it refers to. Each run is one job of the processor's pool.
- **L — Lean.** Six command ids, five key rows, eight modes: only what the commands of this part use. The run keeps no output, because no code reads it. Each rule and each chain is one table entry.
- **E — Explicit.** Each command line, each mode, and each id is a named row of a cast table. A command that git would reject has no key in the sheet.
- **S — SSOT.** `## CommandId` is the one list of the commands. `Source/keys.md` is the one list of the TIT keys. The `git` rows of Part 8 are the one source of availability.
- **S — Stateless.** `TITView` and `TITGitProcessor` hold no state. The mode and the running flag are Model rows. The processor writes its rows through the atomic lane, and the view reads the tree (ruling 57).
- **E — Encapsulation.** `TITView` tells the processor to run lines. It makes no command line and touches no file: the processor does (Overview, item 6). It reads no result from the processor: the result comes back as Model rows through the detection. Data flows one way: key, sheet, command, processor, repository, detection, Model, listener, mode and sheet. No callback asks the Model if it can run.
- **D — Deterministic.** The same `git` rows give the same mode, the same key sheet, and the same command lines.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. Rulings 34 and 35 close the keys of the menus. The eleven items of "Closed by a read" are ORACLE's reads; ARCHITECT can reverse each one. The runtime facts of this part are in Steps 9 and 10.
