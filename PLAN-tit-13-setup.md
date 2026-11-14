# PLAN: TIT Part 13 — Setup, Configuration, and the Repository Start

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part builds the start of TIT and the user configuration.

1. **The user files** (rulings 44 to 47). `~/.config/tit/user.md` holds the preferences: auto-update, the interval, the theme, and the count of the history cache. `~/.config/tit/keys.md` holds the key sheet. Each file is the carried default, written to disk when it is absent or old. The disk file wins when it is valid. Otherwise the carried default is used, and the footer shows the error. A change in titc writes one cell of the disk file. A watcher reads the folder again after each change. This is the shape of the JFS settings, in markdown.
2. **The theme** (rulings 42 and 43). TIT ships one theme, `gfx`: its `style.css`, written as `~/.config/tit/themes/gfx/theme.md` in the END form. A user adds a theme as a folder beside it. The Preferences row cycles the folders. `jam::StyleManager::registerStyle` takes the theme, and the window gets a look-and-feel change.
3. **The Config and Preferences menus.** `/` opens Config: add, switch, or remove the remote; the branch picker; Preferences; Back.
4. **The repository start.** Init (here or in a new folder), clone (here or in a new folder), add a remote, and the branch name mismatch of `tit/SPEC.md:797-819`.
5. **The setup wizard** of `tit/SPEC.md:754-768`: welcome, prerequisites, email, the key generation, the public key, complete, error.
6. **The startup gate, the auto-update, and the turning spinner.** The first fetch of a repository with a remote shows `Checking Remote...`. A timer refreshes the state in the background. The header shows `Syncing...` with a turning spinner, and the loading rows of the menu turn.
7. **LFS.** `lfs install` at the start, `lfs fetch` after the fetch of a new remote, `lfs checkout` after each hard reset.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 8 to 12. Step 1 adds rows to one cast table of JAM.

This part replaces three statements of earlier parts:

- `PLAN-tit-2-spec.md` Step 9, item 1: "Carry the five shipped themes and the 13 colour families". Ruling 43 ships one theme. SPEC carries the theme folders of this part.
- `PLAN-tit-9-commands.md`, Closed by a read, item 10: "`getOrCreateDefaultFile` … `end/Source/config/ConfigModel.cpp:46-104`". That END code changed on 13 Nov. The form of this part is `jam::ConfigDocument::getOrCreateFileDocument` (`jam/jam_markdown/document/jam_ConfigDocument.cpp:1014-1029`).
- `PLAN-tit-11-history.md` Step 2: the cast row `historyCount` goes. The count is the user row of this part (ruling 38).
- `PLAN-tit-2-spec.md` Step 5, item 1, for `tit/SPEC.md:770-776` and `:825-832`: SPEC states that the console shows the text of the tool in these two cases (Closed by a read, item 24).

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization, and parentheses for an initializer of the same container type; no `==` on strings; no out-parameter; no fake-carrier struct; no local struct; no anonymous namespace; structured bindings for a pair, with descriptive names.
- CODING.md, "Uniform signature in the type" and "Consume the event payload".
- MANIFESTO L: no `if` chain longer than three branches; no function body longer than 30 lines. Each family of commands is its own `register…Actions()` function, in the END form (Part 9, Overview, item 6).
- MANIFESTO E, MVP: "The View composes. The Processor holds the business logic." Each function of this part that reads or writes a file, makes a folder, or makes a command line is a member of `TITGitProcessor` or of `TITModel`.
- MANIFESTO S (Stateless): "State belongs exclusively to the Model." The preferences, the key sheet, the theme, the wizard step, the clone folder, the time of the last key, and the update flag are Model rows.
- MANIFESTO E: the established pattern is the listener. A file change reaches the Model through the watcher. A Model change reaches the views through `juce::ValueTree::Listener`. This part adds no callback outside that path.
- Ruling 7: a layout locks at execution. Each geometry number is a named constant.
- ARCHETYPE.md §1: no step edits a generated file by hand. Each cast table keeps its `+---` separator lines.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build and Git: agents build only on ARCHITECT's explicit request. The fixtures need git commands; ARCHITECT runs them, or gives the git instruction. Each fixture of Step 14 runs with its own `HOME`, so the user files and `~/.ssh` of the machine do not change.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`.

### Rulings of this part, verbatim

Session 2026-11-13.

- **Ruling 42.** ARCHITECT: *"reminder we dont want byte-to-byte translation. we dont need theme manager for new tit. jam has StyleManager, we have our own robust css-md-Style as Theme, nothing need to be reinvent"*.
- **Ruling 43.** Selected: *"Ship gfx only"*. Option text: "TIT ships one theme, its style.css. A user adds a theme as a folder in ~/.config/tit/themes/<name>/ in the END form. The Preferences row cycles the folders. Spring, summer, autumn, winter are not shipped."
- **Ruling 44.** ARCHITECT: *"NO. CLI is not pluging. we dont need *.settings"*.
- **Ruling 45.** ARCHITECT: *"what if we just write it ~/.config/tit/config.md ? so default config write virgin to disk, changes on tit write to disk, disk always win (if valid)"*.
- **Ruling 46.** ARCHITECT: *"dont overcomplicate. no handroll. basically default carried config.md is identical to write on disk. the different is runtime can write to disk. carried binary serves as default tree, when disk invalid, so runtime never failed. this is the pattern isomorphic with JFS settings, only JFS use xml."*
- **Ruling 47.** ARCHITECT: *"`user.md` it is"*. The file of ruling 45 is `~/.config/tit/user.md`.
- **Ruling 38.** ARCHITECT: *"100 commit as default. we could expose for user config"*.
- **Ruling 35.** Selected: *"addRemote, calls pullRebase"*. Option text: "Row `addRemote | r | menuNormal`. With no remote it adds a remote. With a remote it calls `pullRebase`."
- **Ruling 17.** Selected: *"Leave them, write new"*. The new TIT never reads or changes a `.toml` file.

### SPEC lines that this part builds

| Subject | SPEC |
|---|---|
| The environment check and the setup wizard, seven phases | `tit/SPEC.md:734-768` |
| The repository check, init, clone | `tit/SPEC.md:778-795` |
| The branch name mismatch on the first remote | `tit/SPEC.md:797-819` |
| The Config menu and the Preferences menu | `tit/SPEC.md:284-323` |
| The startup gate, LFS | `PLAN-tit-2-spec.md` Step 5, items 3 and 5 |
| The user folder, the preferences and their rules | `PLAN-tit-2-spec.md` Step 9, item 1 |
| `lfs checkout` after a hard reset, also after a pin | `PLAN-tit-2-spec.md` Step 7, **Pin**, item 4 |

### The JFS shape, and the JAM call for each step (ruling 46)

JFS is `kuassa/jreng-filter-strip`; its settings code is in KANJUT (`kuassa/user_modules`). JAM is decoupled from KANJUT, so TIT takes the shape and no code.

| Step | JFS (XML) | TIT (markdown), the JAM call |
|---|---|---|
| The default | `DefaultSettings.xml` in BinaryData (`kuassa_PluginEditorLayout.cpp:684-693`) | `Source/user.md`, `Source/keys.md`, `Source/theme.md` in BinaryData; `BinaryData::getString (const juce::String&)` (JAM doxygen, namespace `BinaryData`) |
| The load | the disk file wins when its version is current and it is valid; an absent or old file gets the default text; an invalid file gives the default tree and stays on disk (`kuassa_PluginEditorLayout.cpp:31-54`) | `jam::ConfigDocument::getOrCreateFileDocument (file, defaultDocument, getDefaultText)`: an old version or an absent file gets the default text (`jam_ConfigDocument.cpp:1014-1029`; `Format::isVersionOld`). Then the END validation: `jam::ConfigValidator::addRules`, `addTableRules`, `jam::Document::Validator::isValid` (`end/Source/config/ConfigDirectory.cpp:71-86`). An invalid file gives the parsed default text. |
| The write | parse the disk file, set one value, write the whole document (`kuassa_SettingsDocument.cpp:33-53`); a file that does not parse is not written | `jam::ConfigDocument::parse`, `getConfigRow (tableId, key)` (`jam_ConfigDocument.h:133`), the value cell `getChildByID (Id::value)`, `jam::MarkdownWriter { defaultLineWrap, {}, { { cell, text } } }.getText (document)` (`jam_MarkdownWriter.h:22,31-33,45`), `juce::File::replaceWithText` |
| The reload | a watcher on the folder, 300 ms, then the load again (`kuassa_PluginEditor.cpp:75-78,199-219`) | `jam::File::Watcher`: `addFolder`, `coalesceEvents`, `addListener`, `fileChanged` (`end/Source/config/ConfigModel.h:55-59`; `.cpp:125-136`) |

### The theme, END form

| Fact | Where |
|---|---|
| The default text of a theme file: the BinaryData `theme.md` (front matter and one line), a newline, then the tables of the shipped style sheet | `end/Source/config/ConfigTheme.cpp:34-46`: `BinaryData::getString (fileName) + newline + jam::MarkdownWriter {}.getText (jam::ConfigDocument::toDocument (*jam::CssDocument::getOrCreate (juce::Identifier { files::themeStyleSheet })))` |
| A theme file to a style tree | `ConfigTheme.cpp:28-32`: `jam::CssDocument::parse (jam::CssWriter {}.getText (document))->getValueTree (rootType, jam::CssDocument::getSheetType (styleSheet))` |
| The tree: one child typed by the sheet name, a `FONTS` child, a `WINDOW` child, one child for each class | `jam_web/css/jam_Css.h:182-203` |
| The sheet type is the type form of the file name without its extension | `jam_Css.h:227-236` |
| `registerStyle (juce::ValueTree newConfig)` replaces every sheet, unit, and font | `jam_style/style_manager/jam_StyleManager.h:108-117`; `.cpp:21-52` |
| The constructor of TIT passes `jam::Strings { files::styleSheet }`; it builds the tree with `Css::getValueTree (Format::toType (Id::config), styleSheets)` | `jam_StyleManager.cpp:14-18`; Part 7 Step 6 |
| `StyleCustom::setAppearance (juce::StringRef lightOrDark)` applies the sheet of the style through `StyleManager::setAppearance (sheet, *this, lightOrDark)`. `TITLookAndFeel` sets `Id::light` in its constructor. | `jam_style/jam_StyleCustom.h:237-246`; `jam_StyleCustom.cpp:244-254`; Part 7 Step 4 |
| A theme change: `registerStyle`, the colours, then `sendLookAndFeelChange` on each desktop component | `end/Source/lookAndFeel/ENDLookAndFeel.cpp:507-516`; `end/Source/Main.cpp:4-10,61-65` |
| `files::theme` = `theme.md` and `files::styleSheet` = `style.css` are JAM words | `jam/cast/style.md:994-995` |
| `Extensions::md`, `Id::tokenTrue`, `Id::tokenFalse`, `Id::config`, `Id::keys`, `Id::light` are JAM words | `jam/cast/core.md:61,150,153,2125-2152`; `jam/cast/gui.md:49`; `jam/cast/style.md:34` |
| A config table with only `key`, `type`, `value`, `description` columns gives one property for each row; another column gives one child for each row. A row is named by the type form of its key. | `jam_ConfigDocument.h:145-157`; `jam_ConfigDocument.cpp:272-362` |
| END tables use the types `bool`, `int`, `string`, and `true` / `false` | `end/Source/config/display.md:16,29,76` |
| `jam_web` is a module of the cli archetype | `ling/project-info.md:129` |

### The legacy code that this part carries (ruling 3)

| Piece | Where |
|---|---|
| The environment order: the test variable, git, ssh, an SSH key, ready; a state that is not ready opens the wizard at Welcome | `legacy/git/environment.go:15-34`; `legacy/app/app_constructor.go:46-50` (Part 8) |
| The wizard steps and their texts | `legacy/app/modes.go:46-52`; `legacy/app/setup_wizard_steps.go:19-228` |
| The email check: trimmed, must hold `@` and `.` | `legacy/app/setup_wizard_handlers.go:39-57` |
| The key: `~/.ssh/TIT_id_rsa`; skip when it exists; `ssh-keygen -t rsa -b 4096 -C <email> -f <keyPath> -N ""`; `ssh-add <keyPath>` | `legacy/git/ssh.go:13-151` |
| The `~/.ssh/config` block, appended when the file lacks `AddKeysToAgent yes` and `IdentityFile ~/.ssh/TIT_id_rsa` | `legacy/git/ssh.go:88-92` |
| The public key to the clipboard; the provider addresses | `legacy/app/setup_wizard_steps.go:140-192` |
| The NotRepo menu, the two location menus, their keys, labels, hints | `legacy/app/menu_items.go:8-23,181-212` |
| The empty folder test: every entry is `.DS_Store` or `.AppleDouble` | `legacy/app/dispatch_base.go:14-33` |
| Init: the prompt `Initial branch name:` with `main`; `init`; `checkout -b <name>`; the default `.gitignore` when absent; `add .gitignore`; `commit -m "Initialize repository with .gitignore"` | `legacy/app/handlers_input.go:20-31`; `legacy/app/op_init.go:14-75`; `legacy/git/init.go:11-44,198-213` |
| Init in a new folder: the prompt `Repository name:`; `Directory name cannot be empty` | `legacy/app/handlers_input.go` |
| Clone: the prompt `Repository URL:`; `Repository URL cannot be empty`; `Invalid URL format. Try: git@github.com:user/repo.git`; the check of an address | `legacy/ui/validation.go:84-92` |
| The folder name of a clone: the address without `.git`, its last path part, the text after the last `@` | `legacy/git/exec_utils.go:68-82` |
| The clone target: `ClonePath` | `legacy/app/workflow_state.go:8` |
| Add remote: `add_remote` `r` 🔗; the prompt `Remote URL:`; `remote add origin <url>`; `fetch --all --progress`; `lfs fetch` with LFS; `branch --set-upstream-to=…` or `push -u origin <branch>` | `legacy/app/menu_items.go:137-144`; `legacy/app/handlers_config_remote.go` |
| The Config menu: its order, keys, labels, hints | `legacy/app/menu_render_extra.go:165-195`; `legacy/app/menu_items.go:249-306` |
| Switch remote: `New remote URL:`; `remote set-url origin <url>`; then the fetch and the upstream | `legacy/app/handlers_config_remote.go:90-115` |
| Remove remote: the dialog `remove_remote`, the first button selected; `remote remove origin` | `legacy/app/messages_dialog.go:108-113`; `legacy/app/dispatch_dialog.go:42-57`; `legacy/ui/confirmation.go:45`; `legacy/app/handlers_config_remote.go:118-138` |
| The Preferences menu: three rows; Enter and Space toggle or cycle; `=` `-` `+` `_` change the interval on its row only; 1 to 60 | `legacy/app/menu_items.go:332-355`; `legacy/app/app_keys.go:138-146`; `legacy/app/handlers_config_prefs.go:14-104`; `legacy/app/dispatch_dialog.go:184-226` |
| Enter on the interval row does nothing: "Interval is adjusted via +/- keys, not enter" | `legacy/app/dispatch_dialog.go:196-201` |
| Auto-update on: an update at once | `legacy/app/dispatch_dialog.go:184-194` |
| The row values: `ON` / `OFF`, `%d min`, the theme name; the columns emoji, label, value (10 cells) | `legacy/ui/preferences.go:22-60` |
| The footer of Preferences | `legacy/app/messages_menu.go:169-175` |
| Auto-update: each interval, in the menu only, not within 30 s of a key; `fetch origin` when a remote exists; then the detection | `legacy/app/activity_state.go:5-40` |
| The header while it syncs: `<frame> Syncing...`, `Checking local state...`, `Fetching remote updates...`; ten frames | `legacy/ui/header.go:36-49,146-180`; `legacy/ui/spinner.go:5-16` |
| The startup gate: `<frame> Checking Remote...`; `fetch --quiet`; on failure `Remote unreachable — showing cached state` | `legacy/app/app_view_main.go:167-173`; `legacy/app/app_constructor.go:209-210`; `legacy/app/app_update_msg.go:206-232` |
| LFS at the start: `lfs install` and `LFS filters installed automatically`, or `This project uses LFS. Install git-lfs for full functionality.` | `legacy/app/app_constructor.go:80-92`; `legacy/git/lfs.go:10-47` |
| `lfs checkout` after the hard reset of `replace_local` | `legacy/app/op_pull.go:89-96` |
| The defaults: theme `gfx`, interval 5, auto-update on | `legacy/config/config.go:14-22` |

### Legacy faults that this part corrects

Ruling 4: "kept the feature FIX THE DESIGN, DESIGN BY CONTRACT."

1. **Escape in the wizard left the wizard,** where the text said "go back and try again" (`setup_wizard_steps.go:210`). The wizard mode binds only Enter (`legacy/app/app_keys.go:121-123`), so Escape took the global rule, which goes to the mode before (`legacy/app/app_keys.go:12`; `legacy/app/handlers_global_menu.go:15-47,124-131`). In this part Escape on the error step goes back to the email step.
2. **Enter on the error step opened the generate step and ran nothing** (`legacy/app/setup_wizard_handlers.go`). In this part it opens the email step, and Enter there runs the chain.
3. **The wizard started an `ssh-agent` as a child of titc** and set its socket only for titc (`legacy/git/ssh.go`). The agent outlived titc, and no other program knew its socket. This part starts no agent. The `Host *` block holds `AddKeysToAgent yes` and `IdentityFile`: ssh reads the key from the file, and adds it to a running agent (`man ssh_config`, `AddKeysToAgent`, `IdentityFile`). `ssh-add` runs one time. With no agent it fails, and the console shows the text of the tool.
4. **The auto-update ignored a failed fetch** (MANIFESTO E, "No silent fail"). In this part each failed fetch shows `Remote unreachable — showing cached state`.
5. **`remove remote` did not run the detection** (`legacy/app/handlers_config_remote.go:118-138`). In this part `run` runs the detection after each chain (Part 9 Step 5).
6. **A clone into the current folder parsed `ls-remote --symref` by hand** (`legacy/git/init.go:125-173`). In this part titc removes `.DS_Store` and `.AppleDouble`, then runs `git clone --progress <url> .`. The tool: "Cloning into an existing directory is only allowed if the directory is empty" (`man git-clone`).
7. **The theme files were written again at each start,** so a user edit was lost (legacy inventory, item 9). In this part a theme file carries a version, and only an old or an absent file is written.
8. **The history count was the literal 30** (`legacy/app/history_cache.go:60,125`). In this part it is the user row `history count`, 100 (ruling 38).

### Words the framework and the earlier parts already hold

| Word | Where |
|---|---|
| `Id::config`, `Id::keys`, `Id::value`, `Id::id`, `Id::text`, `Id::mode`, `Id::message`, `Id::directory`, `Id::light`, `Id::tokenTrue`, `Id::tokenFalse`, `Extensions::md`, `files::theme`, `files::styleSheet`, `Chars::newline`, `Chars::space`, `Chars::colon` | JAM cast tables (table above) |
| The rows of the `git` child, `map::GitEnvironment`, `map::Operation`, `map::Remote`, `files::sshDirectory`, `files::gitDirectory`, `text::Git::remote`, `text::Git::lfsVersion`, `text::Git::publicKeyExtension`, `getOutput`, `getArguments`, `applyBranchRows`, `applyLfsRows`, `getDefaultGit` | Parts 8 and 12 |
| `run`, `isRunning`, `Id::selectedIndex`, `Id::attribute`, `Id::commandLines`, `Id::isCacheReady`, `setConfirmation`, `InputView::setDocument (text)`, `InputView::getText`, `getMenuCommands`, `getMenu`, `getMenuItem`, `setMenu`, `isRecordMode`, `menuKeys`, `getMenuKey`, `availability`, `commands`, `events`, `cancellations`, `selections`, `getFormattedText`, `TITGitProcessor::getLine`, `TITGitProcessor::lineEnding`, the command families, `firstButton`, `secondButton`, `isMenuMode`, `getFocusComponent`, `setView` | Parts 9 to 12 |
| `text::Footer::input`, `text::Footer::branchNameEmpty`, `text::Git::createBranch` (`git checkout -b`), `text::Git::add` (`git add --`), `text::Git::pushUpstream` (`git push --progress -u origin`), `text::Git::remoteBranchPrefix`, `text::Reading::hashPlaceholder` (`%s`) | Parts 9 to 12 |

### Closed by a read

ARCHITECT can reverse each one.

1. **Two user files, `user.md` and `keys.md`, in one folder.** Ruling 45 names one file for the preferences. The key sheet is a second file in the END form: END keeps `display.md` and `keys.md` in one folder (`end/cast/bimaps.md:1-33`). Both files take the one path of ruling 46. A key table in `user.md` would carry each row of `keys.md` a second time (MANIFESTO S).
2. **The folder is `~/.config/tit`,** the folder of `PLAN-tit-2-spec.md` Step 9 and `legacy/config/config.go:48`. It is a row of `cast/files.md`, not a build macro: END's macro gives a Debug build its own folder (`end/project-info.md:367`), and no TIT code reads such a difference (MANIFESTO L, YAGNI).
3. **The theme file is `theme.md` in the END form, and `gfx` is a folder on disk.** Ruling 43: "in the END form"; "The Preferences row cycles the folders". The default text comes from the shipped `style.css` (END form), so no second copy of one style is carried (ruling 46). The front matter version makes a new TIT write a new `gfx` file only when the version changes (fault 7).
4. **A file change reloads the three files.** `loadFromPath` is one complete read. The preferences go through `setValuesFrom`, so only a changed row gives an event. The key sheet and the theme are complete children of the Model, replaced whole (MANIFESTO E, "A complete state replaces a complete state").
5. **The listener of the key sheet is `TITView`, and the listener of the theme is the application.** `TITView` builds the key sheet from the state (Part 9 Step 7, item 3; ruling 56). The application owns the style manager (Part 7 Step 6). Each one reacts to the child that the event delivers (CODING.md, "Consume the event payload").
6. **The appearance is `Id::light`,** which the constructor of `TITLookAndFeel` sets (Part 7 Step 4). TIT has no appearance row. The colours of `style.css` are in its `.appearance` block (Part 10 Step 8).
7. **The wizard is a dialog with one button, in its own mode, `setupWizard`.** Each legacy step is a title, a text, and one button (`setup_wizard_steps.go`). The dialog is `jam::TerminalMessageBox::showAlert` (Part 10 Step 14): the `juce::AlertWindow` of JUCE with one button. Enter and Escape run that button (JUCE `juce_gui_basics/lookandfeel/juce_LookAndFeel_V2.cpp:402-407`), and its callback invokes `setupWizardEnter`. The email is typed in the input editor of Part 10, in the mode `email`. The generate step is a run in the console.
8. **The wizard step is a Model row, `setupStep`,** with the legacy table of steps (`legacy/app/modes.go:46-52`). The environment row of Part 8 decides `prerequisites`. A generation succeeds when the key file exists. With `TIT_TEST_SETUP=1` the detection gives `needsSetup` at each run, so the wizard starts again after Complete (Part 8 Step 9, fixture 2).
9. **One chain for a remote, with the fallback of `run`.** The lines: `remote add` (or `remote set-url`), `fetch --all --progress`, `lfs fetch` with LFS, `remote set-head origin --auto`, `branch --set-upstream-to=origin/<branch>`. The fallback: `push --progress -u origin <branch>`. The legacy test `rev-parse --verify` is the failure of the upstream line, and the first failure starts the fallback lines (Part 9 Step 5). With an empty remote, `set-head` fails, and the fallback pushes the branch.
10. **The branch mismatch reads `refs/remotes/origin/HEAD`.** `remote set-head origin --auto` sets it from the remote, and a clone sets it. The detection writes the row `remoteHead` (`symbolic-ref --short`). A new value that differs from `origin/<current branch>` opens the dialog of `tit/SPEC.md:802-819`. The start of titc does not open it: the mode is `startup` during the first detections.
11. **`r` in the Config menu.** SPEC gives `r` to Add Remote and to Remove Remote, which never show together (`tit/SPEC.md:299-301,308-309`). One key row serves one command in one mode (Part 12 Step 4, validation). So `addRemote` of ruling 35 chains once more: with a remote, in `menuConfig`, it calls `removeRemote`. `removeRemote` has no key row, as `pullRebase` has none.
12. **The clone target is the row `clonePath`** (legacy `ClonePath`): the current folder for "Clone to directory", empty for "Create subdirectory". A new folder takes the legacy name of the address.
13. **A new folder becomes the `directory` row before its chain runs.** `run` starts each line in the folder of that row (Part 9 Step 5). So init and clone in a new folder change no process folder.
14. **The auto-update and the startup gate are one request, `update()`** of `TITGitProcessor`. It fetches when the repository has a remote, then detects, in one pool job. The row `isUpdating` drives the header and the menu.
15. **The spinner frame comes from the clock:** `juce::Time::getMillisecondCounter() / spinnerMilliseconds`, modulo the ten frames. A view with a spinner repaints on a `juce::TimedCallback` while it turns (the member form of Part 5, `vBlankManager`). No object counts frames (MANIFESTO S). Part 11, Closed by a read, item 7, named "a frame row and a timer"; the timer is enough.
16. **The startup gate is the mode `startup`,** the legacy `ModeStartup` (`legacy/app/app_keys.go:148-150`). Its menu has one inactive row, `Checking Remote...`, with the spinner, when the repository has a remote.
17. **The auto-update needs 30 s with no key,** as legacy. The time of the last key is the root row `lastActivity`, written by `TITView::keyPressed`. A key that an editor takes does not reach it, and the auto-update runs only in an operation menu.
18. **`lfs checkout` follows each hard reset when LFS is ready,** as Part 2 Step 7 writes. One function, `getResetLines`, gives the reset line and the checkout line. Each hard reset of Parts 9 to 12 uses it (Step 12).
19. **The legacy narration lines of a chain show in the footer, not in the console:** `Initializing repository...`, `Querying remote default branch...`, `Fetching from remote...`, `Setting upstream tracking...`, `Remote added`, `Remote URL updated successfully`, `Remote removed successfully`, `Repository initialized with branch '%s'`, and the step lines of the key generation. Legacy shows them (78 reads of `OutputMessages` in 14 files of `legacy/app/`), so ruling 3 keeps them. The console holds only the bytes of the tool (Part 10 Step 10; ruling 58), and the text of a run that is not the tool's shows in the footer, as the command line and the abort text do (tracker, Status row 10). `tit/SPEC.md` Section 6.8 holds the texts and this rule. Step 3 item 6 adds them to `## footer`.
20. **Checks that an inactive command replaces:** `Remote 'origin' already exists` and `No remote configured to switch`. `addRemote` adds only with no remote. `switchRemote` is active only with a remote (tracker, "Command availability").
21. **`URL cannot be empty` of Switch Remote is `Remote URL cannot be empty`:** one prompt, one check, one text.
22. **Not shown:** `TimelineSyncMessages` (`legacy/app/messages_menu.go:179-185`) has no reader in legacy; `tit/SPEC.md` Section 6.6 lists them as texts that legacy never shows. The nested repository dialog has no opener (Part 2 inventory). The Config items `config_toggle_auto_update`, `config_new_branch`, `config_switch_branch`, `config_merge_branch` are in no menu (`legacy/app/menu_render_extra.go:165-195`); the branch picker of Part 12 does their work.
23. **The input prompts are footer messages,** as `New branch name:` of Part 12 Step 12. The legacy input hints are not carried, as in Part 12.
24. **Not built: `tit/SPEC.md:770-776`, the check of `user.name` and `user.email`, and `tit/SPEC.md:825-832`, the bare repository.** Legacy has no code for either (search of `legacy/`, this session), so ruling 3 carries no feature. A check of the user name needs a wizard step that sets it, and a bare dialog needs a dialog button that runs a command id: each is a new mechanism with no reader in the legacy tool (MANIFESTO L, YAGNI; MANIFESTO E, "A new pattern where one already exists is always wrong"). Neither case is silent. A commit with no `user.name` fails, and the console shows the text of the tool ("Please tell me who you are"). In a bare folder, `checkout -b` of init fails with the text of the tool ("must be run in a work tree"), and the console shows it (MANIFESTO E, "invalid state is loud").

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: JAM, the modes
**Scope:** `jam/cast/gui.md`, table `## EditorMode`; then `jam/generated/` through `cast cast/spell.md` from the JAM root.
**Action:**
1. After the `rebaseReword` row of Part 12 Step 1, add eleven rows: `startup` 25, `menuConfig` 26, `menuPreferences` 27, `menuInitializeLocation` 28, `menuCloneLocation` 29, `setupWizard` 30, `email` 31, `initialBranchName` 32, `repositoryName` 33, `repositoryUrl` 34, `remoteUrl` 35.
2. Run `cast cast/spell.md` from the JAM root.

**Validation:**
- `map::EditorMode` holds the eleven names with the keys 25 to 35. The rows 0 to 24 keep their keys. No JAM source file changes.
- The names: `startup`, `menuConfig`, `menuPreferences`, `menuInitializeLocation`, `menuCloneLocation`, `setupWizard` are the legacy modes `ModeStartup`, `ModeConfig`, `ModePreferences`, `ModeInitializeLocation`, `ModeCloneLocation`, `ModeSetupWizard` (`legacy/app/modes.go:21-40`), with the `menu` prefix of ruling 34 on the four menus. `email`, `initialBranchName`, `repositoryName`, `repositoryUrl`, `remoteUrl` are the legacy prompts `Email (for SSH key comment):`, `Initial branch name:`, `Repository name:`, `Repository URL:`, `Remote URL:`, in the family `branchName` of Part 12.
- `menuConfig` to `menuCloneLocation` have the keys 26 to 29, one block, so one range test finds them (Step 9).
- This part adds no colour id.

### Step 2: TIT, the words, the command ids, the files, the command lines
**Scope:** `tit/cast/identifiers.md`, `tit/cast/spell.md`, `tit/cast/files.md`, `tit/cast/git.md`, `tit/project-info.md`.
**Action:**
1. `cast/identifiers.md`, table `## identifiers`: after the last row of Part 12 Step 2, add these rows of type `@id` with an empty description. The value of each row is its name in backticks. Skip a name that a linked JAM module already declares; the build of Step 14 shows a second declaration.

   | Name | Family |
   |---|---|
   | `preferences` | the legacy menu `Preferences`; SPEC "Mode = Preferences" (`tit/SPEC.md:313`) |
   | `autoUpdate`, `updateInterval` | the legacy rows `preferences_auto_update`, `preferences_interval` (`legacy/app/menu_items.go:332-355`) |
   | `theme` | the legacy key `[appearance] theme`; END `display.md` row `theme` |
   | `historyCount` | the cast row of Part 11 Step 2; ruling 38 |
   | `isUpdating` | legacy `autoUpdateInProgress` (`legacy/app/activity_state.go:9`), with the `is` prefix |
   | `isRemoteReachable` | `isLfsReady` (Part 8 Step 2); the legacy text `Remote unreachable — showing cached state` (inventory) |
   | `lastActivity` | legacy `lastActivity` (`legacy/app/activity_state.go:7`) |
   | `setupStep` | legacy `SetupWizardStep` (`legacy/app/modes.go:46`) |
   | `clonePath` | legacy `ClonePath` (`legacy/app/workflow_state.go:8`) |
   | `remoteHead` | the reference `refs/remotes/origin/HEAD`; the word `head` of `headHash` (Part 12) |
   | `isLfsBinaryAvailable` | legacy `IsLFSBinaryAvailable` (`legacy/git/lfs.go:29`) |
   | `configAddRemote` | legacy id `config_add_remote` (`tit/SPEC.md:299`) |
   | `checkingRemote` | the legacy text `Checking Remote...` (`legacy/app/app_view_main.go:173`) |

2. `cast/identifiers.md`: after the table `## ConflictPane` of Part 12 Step 2, add the table `## SetupStep` in the same form, with the brief `@brief The steps of the setup wizard.` The rows: `welcome` 0, `prerequisites` 1, `email` 2, `generate` 3, `displayKey` 4, `complete` 5, `error` 6 (`legacy/app/modes.go:46-52`, without the prefix `SetupStep`). Add its output group to `cast/spell.md` in the form of Part 8 Step 4, item 3.
3. `cast/identifiers.md`, table `## CommandId`: after the `rewordSubmit` row, add these rows. The value of each row is its name in backticks.

   | name | key | Source of the name |
   |---|---|---|
   | `config` | `0x4a0003d` | the legacy mode `ModeConfig`; SPEC "Config mode" (`tit/SPEC.md:289`) |
   | `preferences` | `0x4a0003e` | `config_preferences` (`legacy/app/menu_items.go`) |
   | `switchRemote` | `0x4a0003f` | `config_switch_remote` |
   | `removeRemote` | `0x4a00040` | `config_remove_remote` |
   | `init` | `0x4a00041` | `init` (`legacy/app/menu_items.go:8`) |
   | `clone` | `0x4a00042` | `clone` (`:16`) |
   | `initHere` | `0x4a00043` | `init_here` (`:181`) |
   | `initSubdir` | `0x4a00044` | `init_subdir` (`:189`) |
   | `cloneHere` | `0x4a00045` | `clone_here` (`:197`) |
   | `cloneSubdir` | `0x4a00046` | `clone_subdir` (`:205`) |
   | `initialBranchNameSubmit` | `0x4a00047` | `newBranchNameSubmit` (Part 12) |
   | `repositoryNameSubmit` | `0x4a00048` | the same family |
   | `repositoryUrlSubmit` | `0x4a00049` | the same family |
   | `remoteUrlSubmit` | `0x4a0004a` | the same family |
   | `emailSubmit` | `0x4a0004b` | the same family |
   | `setupWizardEnter` | `0x4a0004c` | `handleSetupWizardEnter` (`legacy/app/app_keys.go:122`) |
   | `preferencesAutoUpdate` | `0x4a0004d` | `preferences_auto_update` (`legacy/app/dispatchers.go:43`) |
   | `preferencesInterval` | `0x4a0004e` | `preferences_interval` (`:44`) |
   | `preferencesTheme` | `0x4a0004f` | `preferences_theme` (`:45`) |
   | `preferencesIncrement` | `0x4a00050` | `handlePreferencesIncrement` (`legacy/app/handlers_config_prefs.go:14`) |
   | `preferencesDecrement` | `0x4a00051` | `handlePreferencesDecrement` (`:36`) |
   | `preferencesIncrementTen` | `0x4a00052` | `handlePreferencesIncrement10` (`:58`), with the number as a word |
   | `preferencesDecrementTen` | `0x4a00053` | `handlePreferencesDecrement10` (`:79`), the same |

4. `cast/files.md`, table `## files`: after the last row of Part 12 Step 2, add these rows of type `@string`:

   | name | value | description |
   |---|---|---|
   | `configDirectory` | `` `.config/tit` `` | The folder of the user files, in the home folder. |
   | `titcUser` | `` `user.md` `` | User preferences. |
   | `themesDirectory` | `` `themes` `` | The folder of the theme folders, in the folder of the user files. |
   | `sshKey` | `` `TIT_id_rsa` `` | The SSH key that the setup wizard makes, in the SSH folder. |
   | `sshConfig` | `` `config` `` | The SSH configuration file, in the SSH folder. |
   | `gitIgnore` | `` `.gitignore` `` | The ignore file of a repository. |
   | `defaultGitIgnore` | `` `gitignore` `` | The default ignore file of a new repository. |
   | `dsStore` | `` `.DS_Store` `` | A file of the macOS Finder that a clone folder can hold. |
   | `appleDouble` | `` `.AppleDouble` `` | A folder of the macOS file sharing that a clone folder can hold. |

   Change the brief to `@brief File names titc reads and writes.`
5. `project-info.md`, table `## binary`: after the `menu` row of Part 10 Step 5, add three rows in its form:

   | name | path | description |
   |---|---|---|
   | `user` | `Source/user.md` | User preferences, embedded as BinaryData |
   | `theme` | `Source/theme.md` | Theme file header, embedded as BinaryData |
   | `gitIgnore` | `Source/gitignore` | Default ignore file, embedded as BinaryData |

6. `cast/git.md`: delete the row `historyCount` of Part 11 Step 2. After the last row of Part 12 Step 2, add these rows:

   | name | value | description |
   |---|---|---|
   | `fetchQuiet` | `` `git fetch --quiet origin` `` | Fetches the branches of origin and prints nothing. |
   | `fetchAll` | `` `git fetch --all --progress` `` | Fetches every remote. |
   | `remoteHead` | `` `git symbolic-ref --short refs/remotes/origin/HEAD` `` | Prints the default branch of origin, with its prefix. |
   | `setHeadAuto` | `` `git remote set-head origin --auto` `` | Sets the default branch of origin from the remote. |
   | `remoteAdd` | `` `git remote add origin` `` | Adds the next address as origin. |
   | `remoteSetUrl` | `` `git remote set-url origin` `` | Sets the next address for origin. |
   | `removeOrigin` | `` `git remote remove origin` `` | Removes origin. |
   | `setUpstreamOrigin` | `` `git branch --set-upstream-to=origin/` `` | The start of the upstream line; the branch name follows with no space. |
   | `lfsFetch` | `` `git lfs fetch` `` | Fetches the LFS objects. |
   | `lfsCheckout` | `` `git lfs checkout` `` | Writes the content of the LFS files into the tree. |
   | `lfsInstall` | `` `git lfs install` `` | Registers the LFS filters. |
   | `init` | `` `git init` `` | Makes a repository in the folder. |
   | `commitInitial` | `` `git commit -m "Initialize repository with .gitignore"` `` | Makes the first commit of a new repository. |
   | `clone` | `` `git clone --progress` `` | Clones the next address into the folder after it. |
   | `currentFolder` | `` `.` `` | The folder of the command. |
   | `sshKeygen` | `` `ssh-keygen -t rsa -b 4096 -C` `` | Makes an SSH key; the comment follows. |
   | `keyFileOption` | `` `-f` `` | The option before the path of the key. |
   | `emptyPassphrase` | `` `-N ""` `` | A key with no passphrase. |
   | `sshAdd` | `` `ssh-add` `` | Adds the next key to the SSH agent. |

**Validation:**
- Each command line equals its legacy literal (inventory) or its documented form. `fetchQuiet`, `remoteHead`, `setHeadAuto`, and `clone` with `currentFolder` are the correction of fault 6 and the reads of Closed by a read, items 9 and 10.
- `commitInitial` holds one argument in double quotes, and `emptyPassphrase` holds one empty argument in double quotes. `jam::ChildProcess::getArguments` gives each quoted group as one argument, and an empty group as an empty argument (`jam_subprocess/subprocess/native/jam_ChildProcess_posix.cpp:265-293`; `jam_ChildProcess.h:61-74`).
- The two uses of `text::Git::historyCount` change in Step 6 and Step 9. No other code names it.
- After the build of Step 14, `Identifiers.h` holds `map::SetupStep` and the 23 new command ids.

### Step 3: TIT, the texts
**Scope:** `tit/cast/text.md`, `tit/cast/spell.md`.
**Action:**
1. `cast/text.md`: after the table `## rebase` of Part 12 Step 3, add three tables in the same form: `## setup` (`@brief Texts of the setup wizard.`), `## repository` (`@brief Texts of init, clone, and the remote.`), `## preferences` (`@brief Texts of the Preferences menu.`). Add their output groups to `cast/spell.md` in the form of Part 10 Step 6, item 5, with the names `Setup`, `Repository`, `Preferences`. No value starts or ends with a space. A newline in a value is `\n`.
2. `## setup` (`legacy/app/setup_wizard_steps.go:19-228`; `legacy/git/ssh.go:88-92`):

   | name | value |
   |---|---|
   | `welcomeTitle` | `Welcome to TIT` |
   | `welcomeBody` | `We'll set up SSH authentication for git.\nThis only needs to be done once per machine.` |
   | `continueButton` | `Continue` |
   | `gitFound` | `✓ Git found` |
   | `gitMissing` | `✗ Git not found` |
   | `sshFound` | `✓ SSH found` |
   | `sshMissing` | `✗ SSH not found` |
   | `installTools` | `Install missing tools:\n  macOS:   brew install git openssh\n  Linux:   sudo apt install git openssh-client\n  Windows: git-scm.com/download` |
   | `recheck` | `Re-check` |
   | `emailPrompt` | `Email (for SSH key comment):` |
   | `emailInvalid` | `Email must hold @ and .` |
   | `emailMark` | `@` |
   | `domainMark` | `.` |
   | `generateTitle` | `Generating SSH Key` |
   | `keyTitle` | `SSH Key Generated` |
   | `keyCopied` | `Your public key has been copied to clipboard` |
   | `providers` | `Add this key to your git provider:\n  • GitHub:    github.com/settings/ssh/new\n  • GitLab:    gitlab.com/-/user_settings/ssh_keys\n  • Bitbucket: bitbucket.org/account/settings/ssh-keys` |
   | `sshConfigBlock` | `Host *\n  AddKeysToAgent yes\n  IdentityFile ~/.ssh/TIT_id_rsa\n  IdentitiesOnly yes` |
   | `addKeysToAgent` | `AddKeysToAgent yes` |
   | `identityFile` | `IdentityFile ~/.ssh/TIT_id_rsa` |
   | `configWritten` | `✓ Configured ~/.ssh/config` |
   | `configFailed` | `⚠ Could not write config (configure manually)` |
   | `errorTitle` | `✗ Setup Error` |
   | `errorHint` | `Press ESC to go back and try again.` |
   | `goBack` | `Go Back` |
   | `completeTitle` | `✓ Setup Complete!` |
   | `completeBody` | `TIT is ready. You can now init or clone repositories.` |

3. `## repository` (`legacy/app/handlers_input.go`; `legacy/ui/validation.go:84-92`; `legacy/git/exec_utils.go:68-82`; `legacy/app/messages_error.go`):

   | name | value |
   |---|---|
   | `initialBranchName` | `Initial branch name:` |
   | `defaultBranch` | `main` |
   | `repositoryName` | `Repository name:` |
   | `directoryNameEmpty` | `Directory name cannot be empty` |
   | `repositoryUrl` | `Repository URL:` |
   | `urlEmpty` | `Repository URL cannot be empty` |
   | `urlInvalid` | `Invalid URL format. Try: git@github.com:user/repo.git` |
   | `remoteUrl` | `Remote URL:` |
   | `newRemoteUrl` | `New remote URL:` |
   | `remoteUrlEmpty` | `Remote URL cannot be empty` |
   | `remoteUrlInvalid` | `Invalid format. Expected:\n  • SSH: git@github.com:user/repo.git\n  • HTTPS: https://github.com/user/repo.git\n  • Local: /path/to/repo or ~/path/to/repo` |
   | `sshPrefix` | `git@` |
   | `httpPrefix` | `http://` |
   | `httpsPrefix` | `https://` |
   | `pathPrefix` | `/` |
   | `homePrefix` | `~` |
   | `repositorySuffix` | `.git` |
   | `pathSeparator` | `/` |
   | `userSeparator` | `@` |

4. `## preferences` (`legacy/ui/preferences.go:28-35`): `on` `ON`, `off` `OFF`, `minutes` `%s min`.
5. `## header`: after the rows of Part 12 Step 3, add `spinnerFrames` `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` (`legacy/ui/spinner.go:5-16`), `syncing` `Syncing...`, `checkingLocalState` `Checking local state...`, `fetchingRemoteUpdates` `Fetching remote updates...` (`legacy/ui/header.go:46,151,179`).
6. `## footer`: after the rows of Part 12 Step 3, add:

   | name | value | Source |
   |---|---|---|
   | `remoteUnreachable` | `Remote unreachable — showing cached state` | `legacy/app/app_update_msg.go:222` |
   | `lfsInstalled` | `LFS filters installed automatically` | `legacy/app/app_constructor.go:86` |
   | `lfsMissing` | `This project uses LFS. Install git-lfs for full functionality.` | `legacy/app/app_constructor.go:92` |
   | `preferences` | `↑↓ navigate  │  Space toggle  │  =/- ±1min  │  +/_ ±10min  │  Esc back` | `legacy/app/messages_menu.go:169-175` |
   | `initializingRepo` | `Initializing repository...` | `legacy/app/messages_state.go:9` |
   | `queryingDefaultBranch` | `Querying remote default branch...` | `legacy/app/handlers_git_workflow.go:61` |
   | `fetchingRemote` | `Fetching from remote...` | `legacy/app/messages_state.go:10` |
   | `settingUpstream` | `Setting upstream tracking...` | `legacy/app/messages_state.go:11` |
   | `remoteAdded` | `Remote added` | `legacy/app/messages_state.go:5` |
   | `remoteUrlUpdated` | `Remote URL updated successfully` | `legacy/app/handlers_config_remote.go:113` |
   | `remoteRemoved` | `Remote removed successfully` | `legacy/app/handlers_config_remote.go:135` |
   | `repositoryInitialized` | `Repository initialized with branch '%s'` | `legacy/app/op_init.go:72` |

   These rows are the narration lines of item 19 of "Closed by a read". The step lines of the key generation take their texts from `## setup`.

7. `## confirmation`: four rows for each dialog, in the form of Part 10 Step 6, item 4:

   | Prefix | Title | Explanation | Yes | No | Source |
   |---|---|---|---|---|---|
   | `removeRemote` | `Remove Remote` | `This will remove the origin remote from your repository.\n\nYou can add a new remote later.\n\nContinue?` | `Remove` | `Cancel` | `legacy/app/messages_dialog.go:108-113` |
   | `branchMismatch` | `⚠️ Branch name mismatch` | `Your local branch: %s\nRemote default branch: %s\n\nThis may cause confusion. Would you like to switch?` | `Switch to %s` | `Stay on %s` | `tit/SPEC.md:802-809` |

**Validation:**
- Each value equals its legacy text or its SPEC text. `emailInvalid` is the new text of ruling 48.
- `installTools` has no leading newline: Step 11 joins the lines with a blank line, as legacy does (`setup_wizard_steps.go:47-56`).
- `sshConfigBlock` has no last newline: Step 11 writes a newline before and after it.
- The count of `%s` of each value equals the count of the values that Steps 9 to 11 give it.

### Step 4: TIT, the user files, the menu rows, the key rows
**Scope:** `tit/Source/user.md` (new), `tit/Source/theme.md` (new), `tit/Source/gitignore` (new), `tit/Source/menu.md`, `tit/Source/keys.md`.
**Action:**
1. `Source/user.md`: the front matter of `Source/keys.md` (`version: 0.0.1`), then this line, then the table `## preferences` in the grid form of `Source/keys.md`, with the columns `key`, `type`, `value`, `description`:

   > This file is written with the default values on the first run of titc. Edit a value to change titc. titc reads this file again when it changes. An invalid file gives the default values, and the footer shows the error with its file and line.

   | key | type | value | description |
   |---|---|---|---|
   | `auto update` | `bool` | `true` | Refresh the state in the background. |
   | `update interval` | `int` | `5` | Minutes between two refreshes, from 1 to 60. |
   | `theme` | `string` | `gfx` | The folder of the theme in ~/.config/tit/themes/. |
   | `history count` | `int` | `100` | The count of commits in the history cache. |

2. `Source/theme.md`: the front matter `version: 0.0.1`, then the line `This file holds the titc theme style sheet: colours, fonts, metrics and window. An empty dark cell means the same as light.` (the END form, `end/Source/config/theme/gfx/theme.md`).
3. `Source/gitignore`: the text of `DefaultGitignoreContent` (`legacy/git/init.go:12-44`), byte for byte.
4. `Source/menu.md`: after the rows of Part 12 Step 4, add these rows (`legacy/app/menu_items.go:8-23,137-144,181-212,249-306,332-355`; `legacy/app/app_view_main.go:173`; `legacy/app/app_constructor.go:210`):

   | key | value | shortcut | emoji | hint |
   |---|---|---|---|---|
   | `init` | Initialize repository | `i` | 🔨 | Create a new git repository |
   | `clone` | Clone repository | `c` | 📥 | Clone an existing repository from remote URL |
   | `init here` | Initialize directory | `1` | 📍 | Create repository here |
   | `init subdir` | Create subdirectory | `2` | 📁 | Create new folder and initialize there |
   | `clone here` | Clone to directory | `1` | 📍 | Clone repository here |
   | `clone subdir` | Create subdirectory | `2` | 📁 | Create new folder and clone there |
   | `add remote` | Add remote | `r` | 🔗 | Configure a remote repository URL |
   | `config add remote` | Add Remote | `r` | 🔗 | Configure a remote repository URL |
   | `switch remote` | Switch Remote | `s` | 🔗 | Change the remote repository URL |
   | `remove remote` | Remove Remote | `r` | 🗑️ | Remove the configured remote repository |
   | `branch picker` | Branch | `b` | 🌿 | Open branch picker |
   | `preferences` | Preferences | `p` | ⚙️ | Configure auto-update and theme settings |
   | `cancel` | Back | | 🔙 | Return to main menu |
   | `preferences auto update` | Auto-update | | 🔄 | Toggle background state refresh ON/OFF |
   | `preferences interval` | Update Interval | | ⏱️ | Adjust with +/- keys (±1 min) or Shift+/- (±10 min) |
   | `preferences theme` | Theme | | 🎨 | Cycle through available themes |
   | `checking remote` | Checking Remote... | | ⠋ | Checking Remote... |

5. `Source/keys.md`: after the rows of Part 12 Step 4, add these rows. `MENUS` stands for the eight mode names of Part 10 Step 7, written in full.

   | key | type | value | mode | description |
   |---|---|---|---|---|
   | `config` | `string` | `/` | MENUS | Open the Config menu. |
   | `init` | `string` | `i` | `menuNotRepo` | Initialize repository. |
   | `clone` | `string` | `c` | `menuNotRepo` | Clone repository. |
   | `init here` | `string` | `1` | `menuInitializeLocation` | Initialize directory. |
   | `init subdir` | `string` | `2` | `menuInitializeLocation` | Create subdirectory. |
   | `clone here` | `string` | `1` | `menuCloneLocation` | Clone to directory. |
   | `clone subdir` | `string` | `2` | `menuCloneLocation` | Create subdirectory. |
   | `switch remote` | `string` | `s` | `menuConfig` | Switch Remote. |
   | `branch picker` | `string` | `b` | `menuConfig` | Branch. |
   | `preferences` | `string` | `p` | `menuConfig` | Preferences. |
   | `preferences increment` | `string` | `=` | `menuPreferences` | Add one minute to the interval. |
   | `preferences decrement` | `string` | `-` | `menuPreferences` | Remove one minute from the interval. |
   | `preferences increment ten` | `string` | `+` | `menuPreferences` | Add ten minutes to the interval. |
   | `preferences decrement ten` | `string` | `_` | `menuPreferences` | Remove ten minutes from the interval. |
   | `initial branch name submit` | `string` | `return` | `initialBranchName` | Initialize the repository. |
   | `repository name submit` | `string` | `return` | `repositoryName` | Create the folder. |
   | `repository url submit` | `string` | `return` | `repositoryUrl` | Clone the repository. |
   | `remote url submit` | `string` | `return` | `remoteUrl` | Set the remote. |
   | `email submit` | `string` | `return` | `email` | Make the SSH key. |

6. `Source/keys.md`, the mode cells of earlier rows:
   - `add remote` (Part 12): `menuNormal menuConfig`.
   - `quit`, the row with `q` (Part 10): add `startup menuConfig menuPreferences menuInitializeLocation menuCloneLocation setupWizard`.
   - `cancel` (Parts 10 to 12): add `menuConfig menuPreferences menuInitializeLocation menuCloneLocation setupWizard email initialBranchName repositoryName repositoryUrl remoteUrl`.

- The keys are the legacy keys (`legacy/app/menu_items.go`; `legacy/app/app_keys.go:120-150`; `legacy/app/handlers_global_keys.go:126-141`). `juce::KeyPress::createFromDescription` takes the last character of a description with no key name, so `/`, `=`, `-`, `+`, `_`, `1`, `2` are keys (JUCE `juce_gui_basics/keyboard/juce_KeyPress.cpp:183-236`).
- `remove remote` and `pull rebase` have no key row: `add remote` calls them (Closed by a read, item 11; ruling 35).
- Space in Preferences runs the highlighted item, as in legacy (`app_keys.go:139-140`): the stock menu takes return and space (`jam/jam_gui/menu/jam_PopupMenu.h:999-1002`).
- The `version` of `Source/keys.md` stays `0.0.1`: no disk file of it exists before this part. A later part that adds a key row raises the version, so that `getOrCreateFileDocument` writes the new default (`jam_ConfigDocument.cpp:1019`).

**Validation:**
- In one mode, each key has one row. `r` has one row in `menuConfig`. `1` and `2` have one row in each location menu.
- Each mode name is a row of `map::EditorMode` after Step 1.
- Each menu text equals its legacy text. The `shortcut` cell of `remove remote` shows `r`, the key of `add remote` in `menuConfig`.

### Step 5: `TITModel`, the user files
**Scope:** `tit/Source/TITModel.h`, `tit/Source/TITModel.cpp`.
**Action:**
1. Header: the class head becomes `class TITModel : public jam::Model, private jam::File::Watcher::Listener` (`ling/Source/LINGModel.h:15-18`). Public section, after `getGitTexts`:

```cpp
    static juce::ValueTree getDefaultPreferences();
    static juce::File getConfigFile (const juce::String& fileName);
    static juce::StringArray getThemeNames();

    juce::var getPreference (const juce::Identifier& key) const;
    void setPreference (const juce::Identifier& key, const juce::String& text);
    void loadFromPath();
    void startWatcher();
```

   Private section:

```cpp
    void fileChanged (const juce::File& file, jam::File::Watcher::Event) override;
    void setChild (juce::ValueTree child);

    static std::pair<juce::ValueTree, juce::String> getOrCreateConfigTree (const juce::File& file,
                                                                          const jam::ConfigDocument& defaultDocument,
                                                                          const std::function<juce::String()>& getDefaultText,
                                                                          const std::function<juce::ValueTree (const jam::ConfigDocument&)>& getTree,
                                                                          juce::ValueTree defaultTree);
    static juce::ValueTree getConfigTable (const jam::ConfigDocument& document, const juce::Identifier& tableId);
    static juce::ValueTree getStyleTree (const jam::ConfigDocument& document);
    static juce::ValueTree getDefaultStyleTree();
    static const juce::String& getOrCreateDefaultThemeText();

    static constexpr int coalesceMs { 300 };

    jam::File::Watcher watcher;
```

2. Source: at the end of the constructor body, after the rows of Part 12 Step 5:

```cpp
    createAndAddParameter<jam::Parameter<int>> (state, Id::isUpdating, 0);
    createAndAddParameter<jam::Parameter<int>> (state, Id::isRemoteReachable, 1);
    state.setProperty (Id::lastActivity, juce::int64 { 0 }, nullptr);
    state.setProperty (Id::setupStep, static_cast<int> (map::SetupStep::welcome), nullptr);
    state.setProperty (Id::clonePath, juce::String {}, nullptr);
    state.appendChild (getDefaultPreferences(), nullptr);
```

   `getGitTexts()`: after the rows of Part 12 Step 5, add `remoteHead`. `getDefaultGitValues()`: add `isLfsBinaryAvailable` with 0. `isUpdating` and `isRemoteReachable` are parameters of the root: the processor stores them (Step 6; ruling 57).

3. Source: the defaults and the paths.

```cpp
juce::ValueTree TITModel::getDefaultPreferences()
{
    return getConfigTable (*jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcUser }), Id::preferences);
}

juce::File TITModel::getConfigFile (const juce::String& fileName)
{
    return juce::File::getSpecialLocation (juce::File::userHomeDirectory).getChildFile (files::configDirectory).getChildFile (fileName);
}

juce::StringArray TITModel::getThemeNames()
{
    juce::StringArray names;

    for (const auto& folder : getConfigFile (files::themesDirectory).findChildFiles (juce::File::findDirectories, false))
        names.add (folder.getFileName());

    names.sort (true);
    return names;
}

juce::var TITModel::getPreference (const juce::Identifier& key) const
{
    return state.getChildWithName (jam::Format::toType (Id::preferences)).getProperty (jam::Format::toType (key));
}
```

   - The default child comes from the carried file, with no file read: the command-line runs of Part 1 write no user file.
   - `getOrCreate (name)` returns a pointer (`jam_ConfigDocument.h:29`).
   - A preference is a property named by the type form of its key (inventory, the theme table, "A row is named by the type form of its key").
   - `getConfigFile` with the empty text gives the folder itself (JUCE `File::getChildFile`).

4. Source: the load, in the JFS shape (inventory).

```cpp
std::pair<juce::ValueTree, juce::String> TITModel::getOrCreateConfigTree (const juce::File& file,
                                                                         const jam::ConfigDocument& defaultDocument,
                                                                         const std::function<juce::String()>& getDefaultText,
                                                                         const std::function<juce::ValueTree (const jam::ConfigDocument&)>& getTree,
                                                                         juce::ValueTree defaultTree)
{
    const auto fileDocument { jam::ConfigDocument::getOrCreateFileDocument (file, defaultDocument, getDefaultText) };
    jam::Document::Validator::Rules rules {};

    jam::ConfigValidator::addRules (rules, defaultDocument);
    jam::ConfigValidator::addTableRules (rules, defaultDocument);

    const auto result { fileDocument != nullptr ? jam::Document::Validator::isValid (rules, *fileDocument)
                                                : jam::ConfigDocument::isValid (file.loadFileAsString(), file.getFullPathName()) };

    if (result.wasOk()) return { getTree (*fileDocument), {} };
    return { defaultTree, result.getErrorMessage() };
}

juce::ValueTree TITModel::getConfigTable (const jam::ConfigDocument& document, const juce::Identifier& tableId)
{
    auto tables { document.getValueTree (jam::Format::toType (Id::config)) };
    auto table { tables.getChildWithName (jam::Format::toType (tableId)) };

    tables.removeChild (table, nullptr);
    return table;
}

juce::ValueTree TITModel::getStyleTree (const jam::ConfigDocument& document)
{
    return jam::CssDocument::parse (jam::CssWriter {}.getText (document))->getValueTree (jam::Format::toType (Id::theme), jam::CssDocument::getSheetType (files::styleSheet));
}

juce::ValueTree TITModel::getDefaultStyleTree()
{
    return jam::CssDocument::getOrCreate (juce::Identifier { files::styleSheet })->getValueTree (jam::Format::toType (Id::theme), jam::CssDocument::getSheetType (files::styleSheet));
}

const juce::String& TITModel::getOrCreateDefaultThemeText()
{
    static const juce::String defaultText { BinaryData::getString (files::theme) + juce::String::charToString (Chars::newline)
                                            + jam::MarkdownWriter {}.getText (jam::ConfigDocument::toDocument (*jam::CssDocument::getOrCreate (juce::Identifier { files::styleSheet }))) };

    return defaultText;
}

void TITModel::loadFromPath()
{
    const auto& userDocument { *jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcUser }) };
    const auto& keysDocument { *jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcKeys }) };
    const auto& themeDocument { *jam::ConfigDocument::getOrCreate (juce::Identifier { files::theme }, getOrCreateDefaultThemeText()) };
    const auto getPreferencesTable { [] (const jam::ConfigDocument& document) { return getConfigTable (document, Id::preferences); } };
    const auto getKeysTable { [] (const jam::ConfigDocument& document) { return getConfigTable (document, Id::keys); } };
    auto [preferences, userError] { getOrCreateConfigTree (getConfigFile (files::titcUser), userDocument,
                                                           [] { return BinaryData::getString (files::titcUser); },
                                                           getPreferencesTable, getPreferencesTable (userDocument)) };
    auto [keys, keysError] { getOrCreateConfigTree (getConfigFile (files::titcKeys), keysDocument,
                                                    [] { return BinaryData::getString (files::titcKeys); },
                                                    getKeysTable, getKeysTable (keysDocument)) };
    juce::ValueTree config { jam::Format::toType (Id::tit) };

    config.appendChild (preferences, nullptr);
    setValuesFrom (config);

    const auto themeFile { getConfigFile (files::themesDirectory).getChildFile (getPreference (Id::theme).toString()).getChildFile (files::theme) };
    auto [theme, themeError] { getOrCreateConfigTree (themeFile, themeDocument, getOrCreateDefaultThemeText, getStyleTree, getDefaultStyleTree()) };
    juce::StringArray errors { userError, keysError, themeError };

    setChild (keys);
    setChild (theme);
    errors.removeEmptyStrings();

    if (not errors.isEmpty())
        state.setProperty (Id::message, errors.joinIntoString (juce::String::charToString (Chars::newline)), nullptr);
}

void TITModel::setChild (juce::ValueTree child)
{
    state.removeChild (state.getChildWithName (child.getType()), nullptr);
    state.appendChild (child, nullptr);
}
```

   - `getOrCreateFileDocument` writes the default text when the file is absent or its version is old (`jam_ConfigDocument.cpp:1014-1029`). So the first run writes the three files, and a new version of TIT writes a new default over an old file.
   - A valid disk file wins. An invalid file stays on disk. The Model gets the default tree, and the footer shows the error with its file and line (ruling 46: "carried binary serves as default tree, when disk invalid, so runtime never failed"). This is the END form: an invalid file adds only its error, and the defaults stay (`end/Source/config/ConfigDirectory.cpp:79-82`; `end/Source/config/ConfigTheme.cpp:109,126,135`).
   - Each default is built once, as END builds it (MANIFESTO E, Complete at Creation: "built once, read in place"):
     1. `jam::ConfigDocument::getOrCreate` stores each default document. It parses the text only the first time (`jam_ConfigDocument.h:29,41`; `end/Source/config/ConfigModel.cpp:8-11`; `ConfigTheme.cpp:48-51`).
     2. `getOrCreateDefaultThemeText` holds the theme text in a static local, so `MarkdownWriter` runs once (`ConfigTheme.cpp:34-46`). TIT has one theme file, so one static text replaces the map of END.
     3. `getDefaultStyleTree` reads the stored shipped sheet (`jam_Css.h:179`). The default theme tree needs no `CssWriter` pass and no second parse (`ConfigTheme.cpp:53-58`).
   - A load builds the default trees from the stored documents, as END builds `getDefaultConfig` and `getDefaultThemes` at each load (`ConfigModel.cpp:94`; `ConfigTheme.cpp:109`). `getValueTree` builds a new tree at each call (`jam_Css.h:193-214`). A tree is a state update, not a second copy of a document.
   - `files::theme` is the BinaryData name of the theme header and the key of the stored theme document, as in END (`ConfigTheme.cpp:41,50`). No call uses the one-argument `getOrCreate` with this key.
   - The rules come from the default document, as END takes them (`end/Source/config/ConfigDirectory.cpp:71-77`).
   - A parse returns null only for text that is not valid, so `isValid` of that text gives its error.
   - The preferences go through `setValuesFrom`: a changed row gives one event, an equal row none (Part 8 Step 5). The key sheet and the theme replace their children whole (Closed by a read, item 4).
   - `getDefaultPreferences` and the load take the preferences table through one function, `getConfigTable`, from the one stored document (MANIFESTO S).
   - The theme name comes from the preferences that the same load wrote. A name with no folder makes `getOrCreateFileDocument` write the `gfx` text into a new folder of that name (`jam_ConfigDocument.cpp:1023-1024`). END also makes the folder of a theme name that has none (`end/Source/config/ConfigModel.cpp:61`; `ConfigTheme.cpp:87-89`).
   - `setChild` takes a root with no parent: `getConfigTable` removes the table from its tables, and `getValueTree` builds a new root.
   - `std::pair` gives two results with no out-parameter (CODING.md). The bindings have names (CODING.md, structured bindings).

5. Source: the write and the watcher.

```cpp
void TITModel::setPreference (const juce::Identifier& key, const juce::String& text)
{
    const auto file { getConfigFile (files::titcUser) };
    const auto document { jam::ConfigDocument::parse (file.loadFileAsString(), file.getFullPathName()) };

    if (document != nullptr)
    {
        if (const auto* row { document->getConfigRow (Id::preferences, key) }; row != nullptr)
            file.replaceWithText (jam::MarkdownWriter { jam::MarkdownWriter::defaultLineWrap, {}, { { row->getChildByID (Id::value), text } } }.getText (*document));
    }
}

void TITModel::startWatcher()
{
    watcher.addFolder (getConfigFile ({}));
    watcher.coalesceEvents (coalesceMs);
    watcher.addListener (this);
}

void TITModel::fileChanged (const juce::File& file, jam::File::Watcher::Event)
{
    if (file.hasFileExtension (Extensions::md))
        loadFromPath();
}
```

   - `setPreference` is the JFS write (`kuassa_SettingsDocument.cpp:33-53`): parse the disk file, set one cell, write the whole document. A file that does not parse, or has no such row, is not written, as in JFS.
   - The write changes no Model row. The watcher sees the file, `loadFromPath` reads it, and `setValuesFrom` gives the event. So the disk is the one source (ruling 45: "disk always win").
   - `getConfigRow` compares the table id and the key in their tag form (`jam_ConfigDocument.h:124-133`). The id of a cell is the header of its column (`jam_ConfigDocument.h:160-172`).
   - The watcher, the window, and the listener are the END form (`end/Source/config/ConfigModel.cpp:125-136`). The theme files are in a subfolder of the watched folder, as in END.

**Validation:**
- No bail-out guard. No function body is longer than 30 lines. No `==` on strings.
- No hand-written parse, merge, or format of a markdown file: each step is a JAM call of the inventory.
- `TITModel` holds the preferences, the key sheet, and the theme as value data: numbers, booleans, texts, and the colour numbers of the style tree (MANIFESTO E, Materialisation).
- New names and their families: `getDefaultPreferences` (`getDefaultGitValues`, Part 8); `getConfigFile` (`ConfigDirectory::Config::getPath`, `end/Source/config/ConfigDirectory.cpp:19`); `getThemeNames` (legacy `GetAvailableThemes`, `legacy/app/dispatch_dialog.go:206`); `getPreference`, `setPreference` (the table `preferences`; JFS `getSettingsValue`, `setSettingsValue`); `loadFromPath`, `startWatcher`, `fileChanged`, `coalesceMs`, `watcher` (END, verbatim); `setChild` (`setChildren`, Part 11); `getOrCreateConfigTree` (`jam::ConfigDocument::getOrCreateFileDocument`; END `getConfigTree`, `end/Source/config/ConfigModel.cpp:3`); `getConfigTable` (`jam::ConfigDocument::getConfigRow`, `jam_ConfigDocument.h:133`); `getStyleTree`, `getDefaultStyleTree` (END, verbatim, `end/Source/config/ConfigTheme.cpp:28,53`); `getOrCreateDefaultThemeText` (`getOrCreateDefaultText`, `ConfigTheme.cpp:34`). Parameter names: `defaultDocument`, `getDefaultText`, `getTree` (END `addFileTables`, `end/Source/config/ConfigDirectory.cpp:55-61`); `defaultTree`, `tableId`. Local names: `config`, `preferences`, `names`, `folder`, `fileDocument`, `rules`, `result`, `tables`, `table`, `defaultText`, `userDocument`, `keysDocument`, `themeDocument`, `getPreferencesTable`, `getKeysTable`, `userError`, `keys`, `keysError`, `themeFile`, `theme`, `themeError`, `errors`, `file`, `document`, `row`.

### Step 6: `TITGitProcessor`, the update and the rows
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`.
**Action:**
1. Header, public section, after `run`: `void update();`.
2. Source:

```cpp
void TITGitProcessor::update()
{
    model.storeValue (titTag, Id::isUpdating, 1);
    model.flush();
    pool.addJob ([this, directory = juce::File { model.state[Id::directory].toString() }]
    {
        const auto hasRemote { getOutput (directory, jam::ChildProcess::getArguments (text::Git::remote)).value_or (juce::String {}).isNotEmpty() };

        model.storeValue (titTag, Id::isRemoteReachable, not hasRemote or getOutput (directory, jam::ChildProcess::getArguments (text::Git::fetchQuiet)).has_value() ? 1 : 0);
        setGit (directory);
        model.storeValue (titTag, Id::isUpdating, 0);
    });
}
```

   - The form is `addJob` of Part 8 Step 7, item 1: one pool job that stores rows with `storeValue` (ruling 57). The fetch is the legacy startup fetch and the legacy auto-update fetch (inventory).
   - `getOutput` gives no value when the command fails (Part 8 Step 7), so a failed fetch stores `isRemoteReachable` 0. The view shows the footer text at the end of the update (Step 9, item 8; fault 4). The detection runs in both cases: the cached state shows.
   - The message thread stores `isUpdating` 1 and flushes, so the views react before `update` returns, as `run` does (Part 9 Step 5).
3. `addJob` of Part 11 Step 7, item 5: add `count = model.getPreference (Id::historyCount).toString()` to the capture list, after `directory`. Change `getHistoryText (directory)` to `getHistoryText (const juce::File& directory, const juce::String& count)`, pass `count` at its call, and replace `text::Git::historyCount` in its body with `count`.
4. `setBranchRows` of Part 12 Step 6: add, as its last statement, `storeValue (Id::remoteHead, hasRemote ? getOutput (directory, jam::ChildProcess::getArguments (text::Git::remoteHead)).value_or (juce::String {}).trim() : juce::String {});`.
5. `setLfsRows` of Part 12 Step 6: add, as its last statement, `storeValue (Id::isLfsBinaryAvailable, isUsingLfs and getOutput (directory, jam::ChildProcess::getArguments (text::Git::lfsVersion)).has_value());`.

**Validation:**
- No tree of the Model is read on the pool thread. The count is read on the message thread, as the folder is.
- The `git` child has 28 rows.
- `symbolic-ref --short refs/remotes/origin/HEAD` prints `origin/<branch>`, or fails when the reference does not exist. `trim` removes the newline of the output.
- New names: `update` (`addJob`; legacy `startAutoUpdate`); `hasRemote` (Part 12); `count` (`historyCount`).
- The processor writes no tree and calls no `juce::MessageManager::callAsync` in this step (ruling 57).

### Step 7: `HeaderView`, `FooterView`
**Scope:** `tit/Source/HeaderView.h`, `tit/Source/HeaderView.cpp`, `tit/Source/FooterView.cpp`.
**Action:**
1. `HeaderView.h`: public, after `numRows`:

```cpp
    static juce::String getSpinnerFrame();

    static constexpr int spinnerMilliseconds { 100 };
```

   Private, as the last member: `juce::TimedCallback spinner { [this] { repaint(); } };`.
2. `HeaderView.cpp`:

```cpp
juce::String HeaderView::getSpinnerFrame()
{
    const juce::String frames { text::Header::spinnerFrames };
    const auto index { static_cast<int> ((juce::Time::getMillisecondCounter() / spinnerMilliseconds) % static_cast<juce::uint32> (frames.length())) };

    return frames.substring (index, index + 1);
}
```

   In `valueTreePropertyChanged` of Part 10 Step 11, add before the repaint test:

```cpp
    if (property == Id::isUpdating)
    {
        if (static_cast<bool> (tree.getProperty (Id::isUpdating)))
            spinner.startTimer (spinnerMilliseconds);
        else
            spinner.stopTimer();

        repaint();
    }
```

   In `paint`, when the `isUpdating` row of `state` is true, the rows 3 to 6 of the table of Part 10 Step 11 show these texts, with the colours of their rows:

   | Row | Text |
   |---|---|
   | 3, left | `getSpinnerFrame()`, a space, `text::Header::syncing` |
   | 4 | `descriptionIndent` spaces, `text::Header::checkingLocalState` |
   | 5 | `getSpinnerFrame()`, a space, `text::Header::syncing` |
   | 6 | `descriptionIndent` spaces, `text::Header::fetchingRemoteUpdates` |

3. The menu has no file of TIT. The value of a Preferences item is the right-edge text of its `juce::PopupMenu::Item` (Step 9, item 3). An item of a cache that loads cannot run and shows its progress label (Part 11 Step 11, item 5). The menu shows no spinner frame: the stock menu paints an item one time, and the header shows the spinner.
5. `FooterView.cpp`, the table `hints` of `getHint()` (Part 12 Step 8): add the entries `{ map::EditorMode::menuPreferences, { text::Footer::preferences, text::Footer::preferences } }`, and `{ mode, { text::Footer::input, text::Footer::input } }` for each of `email`, `initialBranchName`, `repositoryName`, `repositoryUrl`, `remoteUrl`.

- The frame is a function of the clock (Closed by a read, item 15). The timer only repaints. The legacy period is 100 ms, `CacheRefreshInterval` (`legacy/app/constants.go:10`).

**Validation:**
- No view holds a frame, a value, or a selection. Each paint reads the Model and the clock.
- `TimedCallback` is the last member of the header, so it stops before the view goes (Part 5, `vBlankManager`).
- New names and their families: `getSpinnerFrame`, `spinnerMilliseconds`, `spinner` (legacy `TimelineSyncSpinner`, `legacy/ui/header.go:37`; `CacheRefreshInterval`, `legacy/app/constants.go:10`; `vBlankManager`); `frames`, `index` (local).

### Step 8: `TITView`, the header
**Scope:** `tit/Source/TITView.h`.
**Action:**
1. Private functions, after `registerRewordActions` of Part 12: `void registerConfigActions();`, `void registerRepositoryActions();`, `void registerNameActions();`, `void registerUrlActions();`, `void registerRemoteActions();`, `void registerSetupActions();`. After `registerSelections` of Part 12: `void registerSetups();`, `void registerSetupSteps();`, `void registerMenuValues();`.
2. Private functions, after `setReturnConfirmation` of Part 12:

```cpp
    void setSubmenuMode (int mode);
    void setInputMode (int mode, const char* prompt, const juce::String& text);
    void setSetup (const juce::ValueTree& git);
    void setSetupStep (int step);
    void setSetupConfirmation (const juce::String& title, const juce::String& explanation, const juce::String& button);
    void setPrerequisites (int environment);
    void setKeyConfirmation();
    void setUpdateTimer();
    void setInterval (int delta);
    void onUpdateTimer();
```

3. Private override, after `valueTreePropertyChanged`: `void valueTreeChildAdded (juce::ValueTree& parent, juce::ValueTree& child) override;`. Private functions, after `isDirty` of Part 12:

```cpp
    bool isSetupActive (const juce::ValueTree& git) const;
    bool isSubmenuMode() const;
    bool isInputMode() const;

    static juce::String getParagraphs (const juce::StringArray& paragraphs);
```

   `getMenuCommands`, `getMenu`, and `getMenuItem` are already declared (Part 10 Step 15; Part 11 Step 11).
4. Constants, after `shortHashLength` of Part 11:

```cpp
    static constexpr int minUpdateInterval { 1 };
    static constexpr int maxUpdateInterval { 60 };
    static constexpr int intervalStep { 1 };
    static constexpr int intervalLargeStep { 10 };
    static constexpr int millisecondsPerMinute { 60000 };
    static constexpr juce::int64 activityTimeoutMilliseconds { 30000 };
    static constexpr int paragraphNewlines { 2 };
```

5. Members, after `selections` of Part 12:

```cpp
    jam::Function::Map<juce::CommandID, jam::Function::pointerCapacity, juce::String (const juce::ValueTree&)> menuValues;
    jam::Function::Map<int, jam::Function::pointerCapacity, void ()> setups;
    jam::Function::Map<int, jam::Function::pointerCapacity, void ()> setupActions;
```

   As the last member: `juce::TimedCallback updateTimer { [this] { onUpdateTimer(); } };`.

**Validation:**
- No member holds state: three tables and one timer. The timer is the last member, so it stops before the view goes.
- `TITView` declares no function that reads or writes a file, and none that makes a command line. Those are functions of `TITGitProcessor` (Steps 10 to 12) and of `TITModel` (Step 5).
- The constants are the legacy values: 1 to 60 minutes, steps 1 and 10 (`legacy/app/handlers_config_prefs.go:14-100`); 30 s (`legacy/app/activity_state.go:17`).
- New names and their families: the six `register…Actions` (the `register…Actions()` family, `end/Source/ui/ENDView.h:121-179`; the words `config`, `repository`, `name`, `url`, `remote`, `setup` of the commands and modes of Step 2); `registerSetups`, `registerSetupSteps`, `registerMenuValues`, `menuValues`, `setups`, `setupActions` (`registerCancellations`, `cancellations`, Part 11; `menuKeys`; NAMES.md Rule 8; the table `## SetupStep`); `setSubmenuMode`, `setInputMode` (`setMenuMode`, Part 9); `setSetup`, `setSetupStep`, `setSetupConfirmation`, `setPrerequisites`, `setKeyConfirmation` (the table `## SetupStep`; `setConfirmation`; the legacy step `SetupStepPrerequisites`; the step `displayKey`); `setUpdateTimer`, `onUpdateTimer`, `updateTimer` (legacy `startAutoUpdate`; `onVBlank`, `vBlankManager`, Part 5); `setInterval` (the row `updateInterval`); `getParagraphs` (`getFormattedText`, Part 12); `isSetupActive`, `isSubmenuMode`, `isInputMode` (`isMenuMode`); the constants (legacy `IntervalMinutes`, `activityTimeout`).

### Step 9: `TITView`, the menus, Config, Preferences, the timer, the start
**Scope:** `tit/Source/TITView.cpp`.
**Action:** The functions of Parts 10 to 12 stay, except where an item below says "replace".

1. The constructor: after `registerSelections();`, add `registerSetups();`, `registerSetupSteps();`, and `registerMenuValues();`. In `registerActions()`, add at its end `registerConfigActions();`, `registerRepositoryActions();`, `registerNameActions();`, `registerUrlActions();`, `registerRemoteActions();`, `registerSetupActions();`. Replace the statement `setMenuMode (getGit());` with `state.setProperty (Id::mode, static_cast<int> (map::EditorMode::startup), nullptr);`, and add `setUpdateTimer();` after it.
2. `keyPressed`: add as its first statement `state.setProperty (Id::lastActivity, juce::Time::currentTimeMillis(), nullptr);`.
3. The menus. Replace `getMenuCommands` and `getMenu` of Part 11 Step 11, item 5, with the change of Part 12 Step 11, item 9:

```cpp
juce::Array<juce::CommandID> TITView::getMenuCommands() const
{
    static const juce::Array<juce::CommandID> operationCommands {
        map::CommandId::commit, map::CommandId::commitPush, map::CommandId::resetDiscardChanges, map::CommandId::viewChanges,
        map::CommandId::push, map::CommandId::pullMerge, map::CommandId::pullRebase, map::CommandId::dirtyPullMerge,
        map::CommandId::forcePush, map::CommandId::replaceLocal, map::CommandId::incomingOutgoing, map::CommandId::finalizeMerge,
        map::CommandId::abortMerge, map::CommandId::rebaseContinue, map::CommandId::rebaseAbort, map::CommandId::dirtyContinue,
        map::CommandId::dirtyAbort, map::CommandId::history, map::CommandId::fileHistory, map::CommandId::timeTravelReturn,
        map::CommandId::unpin, map::CommandId::init, map::CommandId::clone, map::CommandId::addRemote,
    };

    static const jam::HashMap<int, juce::Array<juce::CommandID>> submenuCommands {
        { map::EditorMode::menuConfig, { map::CommandId::addRemote, map::CommandId::switchRemote, map::CommandId::removeRemote,
                                         map::CommandId::branchPicker, map::CommandId::preferences, map::CommandId::cancel } },
        { map::EditorMode::menuPreferences, { map::CommandId::preferencesAutoUpdate, map::CommandId::preferencesInterval, map::CommandId::preferencesTheme } },
        { map::EditorMode::menuInitializeLocation, { map::CommandId::initHere, map::CommandId::initSubdir } },
        { map::EditorMode::menuCloneLocation, { map::CommandId::cloneHere, map::CommandId::cloneSubdir } },
        { map::EditorMode::startup, {} },
    };

    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    if (submenuCommands.contains (mode)) return submenuCommands.at (mode);
    return operationCommands;
}

juce::PopupMenu TITView::getMenu (const juce::ValueTree& git) const
{
    static const jam::HashMap<juce::CommandID, juce::Identifier> loadingKeys {
        { map::CommandId::history, Id::historyLoading },
        { map::CommandId::fileHistory, Id::fileHistoryLoading },
    };

    static const juce::Array<juce::CommandID> separatorCommands { map::CommandId::history, map::CommandId::addRemote, map::CommandId::branchPicker, map::CommandId::cancel };

    const auto isBrowsable { isOperation ({ map::Operation::normal, map::Operation::timeTraveling, map::Operation::pinned }) };
    const auto hasRemote { static_cast<int> (git.getProperty (Id::remote)) == map::Remote::hasRemote };
    juce::PopupMenu menu;

    for (const auto command : getMenuCommands())
    {
        const auto isActive { isAvailable (command, git) and not (command == map::CommandId::addRemote and hasRemote) };

        if (isActive or (isBrowsable and loadingKeys.contains (command)))
        {
            if (separatorCommands.contains (command) and menu.getNumItems() > 0)
                menu.addSeparator();

            menu.addItem (getMenuItem (isActive ? getMenuKey (command, git) : loadingKeys.at (command), command, isActive));
        }
    }

    if (static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::startup and hasRemote)
        menu.addItem (getMenuItem (Id::checkingRemote, map::CommandId::cancel, false));

    return menu;
}
```

   - In `getMenuItem` of Part 11 Step 11, item 5, add before its `return`: `if (menuValues.contains (command)) item.shortcutKeyDescription = menuValues.get (command, getGit());`. The value of a Preferences item shows at the right edge of the item, where JUCE shows `shortcutKeyDescription` (JUCE `juce_gui_basics/menus/juce_PopupMenu.h:166-172`).
   - In `setMenu` of Part 11 Step 11, item 5, the condition of the show becomes `isShowing() and isLargeEnough() and (isMenuMode() or isRecordMode() or isSubmenuMode() or static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::startup)`.
   - In the callback of the same `setMenu`, the condition of the `cancel` becomes `result == 0 and (isRecordMode() or isSubmenuMode())`. The menu takes Escape and closes with no item (`jam_gui/menu/jam_PopupMenu.h:1003-1006`), so Escape in a submenu reaches `cancel` through the callback, and `cancel` gives the menu above it (item 12). In a mode of the main menu the callback shows the menu again, as in Part 10.
   - `getMenuItem` keeps the table `progressTexts`, the local `numCommits`, and the progress label of Part 11 Step 11, item 5. The total of that label becomes `model.getPreference (Id::historyCount).toString()` in place of `text::Git::historyCount`: the count of the user (ruling 38).

   - `jam::ConfigDocument::getOrCreate` returns a pointer (`jam_ConfigDocument.h:29`), so the call is `->`, as in Parts 7, 10, and 11 (`jam_ConfigDocument.h:29`).
   - The order of `operationCommands` is the order of Part 12, with `init` and `clone` of the NotRepo menu (`legacy/app/menu_items.go:8-23`) and `addRemote` after a separator at the end (`tit/SPEC.md:280-285`). Each menu shows only its active commands, so `init` and `clone` show only in `NotRepo`.
   - `addRemote` shows in a menu only with no remote (`tit/SPEC.md:284-285,299,308`). With a remote its key calls `pullRebase` or `removeRemote` (Closed by a read, item 11), and the rows `pull rebase` and `remove remote` show those commands.
   - The separators of the Config menu are the legacy separators: before `Branch` and before `Back` (`legacy/app/menu_render_extra.go:165-195`).
   - The startup row cannot run: `isEnabled` is false (Part 11 Step 11, item 5).
4. `registerMenuKeys()`: add

```cpp
    menuKeys.add (map::CommandId::addRemote, [this] (const juce::CommandID&, const juce::ValueTree&)
    {
        if (static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::menuConfig) return juce::Identifier { Id::configAddRemote };
        return commandId->get (map::CommandId::addRemote);
    });
```

5. `registerMenuValues()`:

```cpp
void TITView::registerMenuValues()
{
    menuValues.add (map::CommandId::preferencesAutoUpdate, [this] (const juce::CommandID&, const juce::ValueTree&)
    {
        return juce::String { static_cast<bool> (model.getPreference (Id::autoUpdate)) ? text::Preferences::on : text::Preferences::off };
    });

    menuValues.add (map::CommandId::preferencesInterval, [this] (const juce::CommandID&, const juce::ValueTree&)
    {
        return getFormattedText (text::Preferences::minutes, { model.getPreference (Id::updateInterval).toString() });
    });

    menuValues.add (map::CommandId::preferencesTheme, [this] (const juce::CommandID&, const juce::ValueTree&) { return model.getPreference (Id::theme).toString(); });
}
```

6. The modes. Add:

```cpp
bool TITView::isSubmenuMode() const
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    return mode >= map::EditorMode::menuConfig and mode <= map::EditorMode::menuCloneLocation;
}

bool TITView::isInputMode() const
{
    static const juce::Array<int> inputModes {
        map::EditorMode::input, map::EditorMode::branchName, map::EditorMode::rebaseReword, map::EditorMode::email,
        map::EditorMode::initialBranchName, map::EditorMode::repositoryName, map::EditorMode::repositoryUrl, map::EditorMode::remoteUrl,
    };

    return inputModes.contains (static_cast<int> (state.getProperty (Id::mode)));
}

bool TITView::isEditMode() const
{
    return isInputMode() or static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::search;
}

void TITView::valueTreeChildAdded (juce::ValueTree& parent, juce::ValueTree& child)
{
    if (parent == state and child.hasType (jam::Format::toType (Id::keys)))
        setKeys (getGit());
}

void TITView::setSubmenuMode (int mode)
{
    state.setProperty (Id::mode, mode, nullptr);
}

void TITView::setInputMode (int mode, const char* prompt, const juce::String& text)
{
    commitMessage.setDocument (text);
    state.setProperty (Id::mode, mode, nullptr);
    state.setProperty (Id::message, prompt, nullptr);
}
```

   In `setView` (Parts 10 to 12), change one line:

```cpp
    commitMessage.setVisible (isLarge and isInputMode());
```

   A menu and a dialog need no line: the menu of Part 11 Step 11, item 5, and the box of Part 10 Step 14 show themselves.

   In `getFocusComponent` (Part 12 Step 11, item 7), add one entry for each of `email`, `initialBranchName`, `repositoryName`, `repositoryUrl`, `remoteUrl`, in the form of the `branchName` entry: `&commitMessage`.
   - `isMenuMode` of Part 11 stays: the operation menus. So an `operation` event and the auto-update act only there, as legacy acts in `ModeMenu` only (`legacy/app/activity_state.go`).
   - `setInputMode` is the form of `branchPickerAdd` of Part 12 Step 12: the input with its start text (`InputView::setDocument`, Part 12 Step 10), the mode, and the prompt in the footer.

7. `setMenuMode` of Part 9 Step 7, item 3: replace its last statement with

```cpp
    if (isSetupActive (git))
        setSetup (git);
    else
        state.setProperty (Id::mode, static_cast<int> (menuModes.at (static_cast<size_t> (static_cast<int> (git.getProperty (Id::operation))))), nullptr);
```

8. The events. The `Id::mode` entry of Part 12 Step 11, item 6, already builds the menu and the key sheet, and the rule of `addRemote` reads the mode (item 11). Add these entries:

```cpp
    events.add (Id::isUpdating, [this] (const juce::Identifier&, juce::ValueTree& root)
    {
        const auto isStartup { static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::startup };

        if (isStartup and not static_cast<bool> (root.getProperty (Id::isUpdating)))
        {
            setMenuMode (getGit());
            gitProcessor.setLfsFilters();
        }

        if (not static_cast<bool> (root.getProperty (Id::isUpdating)) and not static_cast<bool> (root.getProperty (Id::isRemoteReachable)))
            state.setProperty (Id::message, text::Footer::remoteUnreachable, nullptr);
    });

    events.add (Id::gitEnvironment, [this] (const juce::Identifier&, juce::ValueTree& git)
    {
        if (static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::setupWizard)
            setMenuMode (git);
    });

    events.add (jam::Format::toType (Id::autoUpdate), [this] (const juce::Identifier&, juce::ValueTree& preferences)
    {
        setUpdateTimer();
        setMenu (getGit());

        if (static_cast<bool> (preferences.getProperty (jam::Format::toType (Id::autoUpdate))))
            gitProcessor.update();
    });

    events.add (jam::Format::toType (Id::updateInterval), [this] (const juce::Identifier&, juce::ValueTree&)
    {
        setUpdateTimer();
        setMenu (getGit());
    });

    events.add (jam::Format::toType (Id::theme), [this] (const juce::Identifier&, juce::ValueTree&) { setMenu (getGit()); });
```

   - The menu of a new mode is built when the mode changes, so the Config, Preferences, and location menus show their rows.
   - The end of the first update closes the gate: the menu of the state, or the wizard. Then the LFS filters (inventory).
   - A preference row reaches `TITView` through `setValuesFrom` of `loadFromPath`, after the disk write of Step 5. So a row of a menu shows the value that the disk holds.
   - Auto-update on gives one update at once (`legacy/app/dispatch_dialog.go:189-191`).

9. The timer:

```cpp
void TITView::setUpdateTimer()
{
    const auto interval { juce::jlimit (minUpdateInterval, maxUpdateInterval, static_cast<int> (model.getPreference (Id::updateInterval))) };

    if (static_cast<bool> (model.getPreference (Id::autoUpdate)))
        updateTimer.startTimer (interval * millisecondsPerMinute);
    else
        updateTimer.stopTimer();
}

void TITView::onUpdateTimer()
{
    const auto idleMilliseconds { juce::Time::currentTimeMillis() - static_cast<juce::int64> (state.getProperty (Id::lastActivity)) };
    const auto isBusy { static_cast<bool> (state.getProperty (Id::isRunning)) or static_cast<bool> (state.getProperty (Id::isUpdating)) };

    if (isMenuMode() and not isBusy and idleMilliseconds > activityTimeoutMilliseconds)
        gitProcessor.update();
}
```

   The interval of a hand-edited file is limited to the legacy range (`legacy/config/config.go:148-153`).

10. `registerConfigActions()`:

```cpp
void TITView::registerConfigActions()
{
    commands.add (map::CommandId::config, [this] (const juce::CommandID&) { setSubmenuMode (map::EditorMode::menuConfig); });
    commands.add (map::CommandId::preferences, [this] (const juce::CommandID&) { setSubmenuMode (map::EditorMode::menuPreferences); });

    commands.add (map::CommandId::preferencesAutoUpdate, [this] (const juce::CommandID&)
    {
        model.setPreference (Id::autoUpdate, (static_cast<bool> (model.getPreference (Id::autoUpdate)) ? Id::tokenFalse : Id::tokenTrue).toString());
    });

    commands.add (map::CommandId::preferencesInterval, [] (const juce::CommandID&) {});

    commands.add (map::CommandId::preferencesTheme, [this] (const juce::CommandID&)
    {
        const auto names { TITModel::getThemeNames() };

        if (not names.isEmpty())
            model.setPreference (Id::theme, names[(names.indexOf (model.getPreference (Id::theme).toString()) + 1) % names.size()]);
    });

    commands.add (map::CommandId::preferencesIncrement, [this] (const juce::CommandID&) { setInterval (intervalStep); });
    commands.add (map::CommandId::preferencesDecrement, [this] (const juce::CommandID&) { setInterval (-intervalStep); });
    commands.add (map::CommandId::preferencesIncrementTen, [this] (const juce::CommandID&) { setInterval (intervalLargeStep); });
    commands.add (map::CommandId::preferencesDecrementTen, [this] (const juce::CommandID&) { setInterval (-intervalLargeStep); });
}

void TITView::setInterval (int delta)
{
    model.setPreference (Id::updateInterval,
                         juce::String { juce::jlimit (minUpdateInterval, maxUpdateInterval, static_cast<int> (model.getPreference (Id::updateInterval)) + delta) });
}
```

   - Enter on the interval row does nothing, as legacy (inventory).
   - The four interval keys change the interval in the mode `menuPreferences`: the mode of their key rows is the gate. The stock menu gives its highlighted item to no caller (Part 10 Step 13), so the keys do not depend on the highlighted row.
   - The theme cycle is the legacy cycle: the next folder after the current one, the first after the last, the first when the current is not found (`legacy/app/dispatch_dialog.go:204-226`). The names are sorted (Step 5).

11. `registerAvailability()`: add these entries in the form of Part 11 Step 11, item 6.

   | Command | Rule |
   |---|---|
   | `init`, `clone` | `operation` is `notRepo` |
   | `switchRemote`, `removeRemote` | `operation` is `normal`, and `remote` is `hasRemote` |

   - Replace the rule of `addRemote` of Part 12 Step 12, item 2: `operation` is `normal`, and one of: `remote` is `noRemote`; the mode is `menuConfig`; the rule of `pullRebase`. The entry captures `this` to read the mode, and the `mode` event builds the sheet (item 8).
   - `config`, `preferences`, `initHere`, `initSubdir`, `cloneHere`, `cloneSubdir`, the five submit commands, `setupWizardEnter`, the three preferences rows, and the four interval commands have no entry. The mode of the key row is the gate of each one.
   - `isEditMode` of Part 11 is replaced in item 6: an input mode of this part, or the mode `search`, shows an editable editor.

12. `registerCancellations()`: add

```cpp
    cancellations.add (map::EditorMode::menuPreferences, [this] (const int&) { setSubmenuMode (map::EditorMode::menuConfig); });
    cancellations.add (map::EditorMode::email, [this] (const int&) { setSetupStep (map::SetupStep::welcome); });
```

   Escape in `menuConfig`, in the two location menus, and in the four repository prompts has no entry: `setMenuMode`, the default of `cancel` (Part 11 Step 11, item 8).

13. `searchHistory` of Part 11 Step 11, item 7: replace `text::Git::historyCount` with `model.getPreference (Id::historyCount).toString()`. The search reads the count of the user, as the cache does (Step 6).

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- The Config menu of a repository with a remote: `Switch Remote`, `Remove Remote`, a separator, `Branch`, `Preferences`, a separator, `Back`. With no remote: `Add Remote` first (`tit/SPEC.md:297-309`).
- The Preferences rows show `ON` or `OFF`, `<n> min`, and the theme name (`legacy/ui/preferences.go:28-35`).
- Local names of the bodies: `operationCommands`, `submenuCommands`, `mode`, `loadingKeys`, `separatorCommands`, `rows`, `isBrowsable`, `hasRemote`, `items`, `command`, `isActive`, `item`, `inputModes`, `isStartup`, `root`, `preferences`, `interval`, `idleMilliseconds`, `isBusy`, `names`.

### Step 10: `TITGitProcessor` and `TITView`, init, clone, the remote, the branch mismatch
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`, `tit/Source/TITView.cpp`; `tit/cast/git.md`.
**Action:**
1. `cast/git.md`: after the rows of Step 2, add `originPrefix`, value `` `origin/` ``, description `The start of the short name of a branch of origin.`. (`remoteBranchPrefix` of Part 8 is `refs/remotes/origin/`, the full name.)
2. `TITGitProcessor`, the repository work. Each function of this item reads or writes a file, or makes command lines, so it is a function of the processor (Part 9, Overview, item 6). Header, public section, after the statics of Part 12 Step 9:

```cpp
    static juce::StringArray getRemoteLines (const juce::ValueTree& git, const juce::String& remoteLine, const juce::String& url);
    static juce::String getRepositoryName (const juce::String& url);
    static bool isValidUrl (const juce::String& url);
```

   After `moveRebaseCommit` of Part 12 Step 9:

```cpp
    void setRepositoryDirectory (const juce::String& name);
    void initRepository (const juce::String& branch);
    void cloneRepository (const juce::String& url);
    void setLfsFilters();
    bool isEmptyDirectory() const;
```

   Private constant, after `lineEnding`: `static constexpr int minHttpUrlLength { 8 };`.

```cpp
bool TITGitProcessor::isEmptyDirectory() const
{
    const auto entries (juce::File { model.state[Id::directory].toString() }.findChildFiles (juce::File::findFilesAndDirectories, false));

    return std::all_of (entries.begin(), entries.end(), [] (const juce::File& entry)
    {
        return entry.getFileName().compare (files::dsStore) == 0 or entry.getFileName().compare (files::appleDouble) == 0;
    });
}

bool TITGitProcessor::isValidUrl (const juce::String& url)
{
    if (url.startsWith (text::Repository::sshPrefix)) return url.containsChar (Chars::colon);
    if (url.startsWith (text::Repository::httpPrefix) or url.startsWith (text::Repository::httpsPrefix)) return url.length() > minHttpUrlLength;
    return url.startsWith (text::Repository::pathPrefix) or url.startsWith (text::Repository::homePrefix);
}

juce::String TITGitProcessor::getRepositoryName (const juce::String& url)
{
    const juce::String suffix { text::Repository::repositorySuffix };
    const auto address { url.endsWith (suffix) ? url.dropLastCharacters (suffix.length()) : url };

    return address.fromLastOccurrenceOf (text::Repository::pathSeparator, false, false).fromLastOccurrenceOf (text::Repository::userSeparator, false, false);
}

void TITGitProcessor::setRepositoryDirectory (const juce::String& name)
{
    const auto folder { juce::File { model.state[Id::directory].toString() }.getChildFile (name) };

    folder.createDirectory();
    setDirectory (folder);
}

void TITGitProcessor::setDirectory (const juce::File& folder)
{
    const auto path { folder.getFullPathName() };

    model.storeValue (titTag, Id::directory, path.toRawUTF8(), static_cast<int> (path.getNumBytesAsUTF8()));
    model.flush();
    addJob();
}

void TITGitProcessor::initRepository (const juce::String& branch)
{
    const auto ignoreFile { juce::File { model.state[Id::directory].toString() }.getChildFile (files::gitIgnore) };

    if (not ignoreFile.existsAsFile())
        ignoreFile.replaceWithText (BinaryData::getString (files::defaultGitIgnore), false, false, lineEnding);

    run ({ text::Git::init, getLine ({ text::Git::createBranch, branch }), getLine ({ text::Git::add, files::gitIgnore }), text::Git::commitInitial }, {});
}

void TITGitProcessor::cloneRepository (const juce::String& url)
{
    const auto clonePath { model.state.getProperty (Id::clonePath).toString() };
    const auto folder { clonePath.isNotEmpty() ? juce::File { clonePath }
                                               : juce::File { model.state[Id::directory].toString() }.getChildFile (getRepositoryName (url)) };

    folder.getChildFile (files::dsStore).deleteFile();
    folder.getChildFile (files::appleDouble).deleteRecursively();
    folder.createDirectory();
    setDirectory (folder);
    run ({ getLine ({ text::Git::clone, url, text::Git::currentFolder }) }, {});
}

juce::StringArray TITGitProcessor::getRemoteLines (const juce::ValueTree& git, const juce::String& remoteLine, const juce::String& url)
{
    juce::StringArray lines { getLine ({ remoteLine, url }), text::Git::fetchAll };

    if (static_cast<bool> (git.getProperty (Id::isUsingLfs)))
        lines.add (text::Git::lfsFetch);

    lines.add (text::Git::setHeadAuto);
    lines.add (juce::String { text::Git::setUpstreamOrigin } + git.getProperty (Id::currentBranch).toString());
    return lines;
}

void TITGitProcessor::setLfsFilters()
{
    const auto git { model.state.getChildWithName (jam::Format::toType (Id::git)) };
    const auto isMissing { static_cast<bool> (git.getProperty (Id::isUsingLfs)) and not static_cast<bool> (git.getProperty (Id::isLfsReady)) };

    if (isMissing and static_cast<bool> (git.getProperty (Id::isLfsBinaryAvailable)))
    {
        run ({ text::Git::lfsInstall }, {});
        model.storeValue (titTag, Id::message, text::Footer::lfsInstalled, static_cast<int> (juce::String { text::Footer::lfsInstalled }.getNumBytesAsUTF8()));
    }
    else if (isMissing)
    {
        model.storeValue (titTag, Id::message, text::Footer::lfsMissing, static_cast<int> (juce::String { text::Footer::lfsMissing }.getNumBytesAsUTF8()));
    }
}
```

   - `isValidUrl` is the legacy check (`legacy/ui/validation.go:84-92`): an SSH address needs a colon, a web address is longer than its scheme, a path starts with `/` or `~`. `minHttpUrlLength` is its number.
   - `getRepositoryName` is the legacy rule (`legacy/git/exec_utils.go:68-82`). `fromLastOccurrenceOf` gives the whole text when the separator is absent (doxygen `classjuce_1_1String`).
   - `setRepositoryDirectory` makes the folder of a new repository and makes it the `directory` row.
   - `setDirectory` stores the `directory` row, flushes the Model, and adds the detection job. Declare it in the private section, after `applyPendingCommandLine`: `void setDirectory (const juce::File& folder);`. The processor writes the row with `storeValue`, and a stored value gives no call to an attachment (Part 7, Model facts), so the function adds the job itself. The message thread calls it, as it calls `run`.
   - `initRepository` is the legacy chain (`legacy/app/op_init.go:14-75`). The ignore file is written before the chain, only when it is absent (`legacy/git/init.go:198-213`).
   - `cloneRepository`: the folder becomes the `directory` row, then the chain runs in it (Closed by a read, items 12 and 13). An existing folder that is not empty makes the tool refuse the clone, and the console shows its text.
   - `getRemoteLines` is the chain of Closed by a read, item 9. `remoteLine` is `remote add` or `remote set-url`.
   - `setLfsFilters` is the legacy start of LFS (`legacy/app/app_constructor.go:80-92`; `legacy/git/lfs.go:10-47`): a repository that uses LFS with no registered filters gets `lfs install` when the LFS tool runs, and the text `lfsMissing` when it does not. The rows come from the detection (Part 8 Step 7; Step 6, item 5).

3. `TITView`, the four families:

```cpp
void TITView::registerRepositoryActions()
{
    commands.add (map::CommandId::init, [this] (const juce::CommandID&) { setSubmenuMode (map::EditorMode::menuInitializeLocation); });

    commands.add (map::CommandId::clone, [this] (const juce::CommandID&)
    {
        if (gitProcessor.isEmptyDirectory())
            setSubmenuMode (map::EditorMode::menuCloneLocation);
        else
            commandManager->invokeDirectly (map::CommandId::cloneSubdir, false);
    });

    commands.add (map::CommandId::initHere, [this] (const juce::CommandID&)
    {
        setInputMode (map::EditorMode::initialBranchName, text::Repository::initialBranchName, text::Repository::defaultBranch);
    });

    commands.add (map::CommandId::initSubdir, [this] (const juce::CommandID&) { setInputMode (map::EditorMode::repositoryName, text::Repository::repositoryName, {}); });

    commands.add (map::CommandId::cloneHere, [this] (const juce::CommandID&)
    {
        state.setProperty (Id::clonePath, state.getProperty (Id::directory), nullptr);
        setInputMode (map::EditorMode::repositoryUrl, text::Repository::repositoryUrl, {});
    });

    commands.add (map::CommandId::cloneSubdir, [this] (const juce::CommandID&)
    {
        state.setProperty (Id::clonePath, juce::String {}, nullptr);
        setInputMode (map::EditorMode::repositoryUrl, text::Repository::repositoryUrl, {});
    });
}

void TITView::registerNameActions()
{
    commands.add (map::CommandId::initialBranchNameSubmit, [this] (const juce::CommandID&)
    {
        const auto branch { commitMessage.getText() };

        if (branch.isEmpty())
            state.setProperty (Id::message, text::Footer::branchNameEmpty, nullptr);
        else
            gitProcessor.initRepository (branch);
    });

    commands.add (map::CommandId::repositoryNameSubmit, [this] (const juce::CommandID&)
    {
        const auto name { commitMessage.getText() };

        if (name.isEmpty())
        {
            state.setProperty (Id::message, text::Repository::directoryNameEmpty, nullptr);
        }
        else
        {
            gitProcessor.setRepositoryDirectory (name);
            setInputMode (map::EditorMode::initialBranchName, text::Repository::initialBranchName, text::Repository::defaultBranch);
        }
    });
}

void TITView::registerUrlActions()
{
    commands.add (map::CommandId::repositoryUrlSubmit, [this] (const juce::CommandID&)
    {
        const auto url { commitMessage.getText() };

        if (url.isEmpty()) state.setProperty (Id::message, text::Repository::urlEmpty, nullptr);
        else if (not TITGitProcessor::isValidUrl (url)) state.setProperty (Id::message, text::Repository::urlInvalid, nullptr);
        else gitProcessor.cloneRepository (url);
    });

    commands.add (map::CommandId::remoteUrlSubmit, [this] (const juce::CommandID&)
    {
        const auto url { commitMessage.getText() };
        const auto git { getGit() };

        if (url.isEmpty()) state.setProperty (Id::message, text::Repository::remoteUrlEmpty, nullptr);
        else if (not TITGitProcessor::isValidUrl (url)) state.setProperty (Id::message, text::Repository::remoteUrlInvalid, nullptr);
        else gitProcessor.run (TITGitProcessor::getRemoteLines (git, state.getProperty (Id::commandLines).toString(), url),
                               { TITGitProcessor::getLine ({ text::Git::pushUpstream, git.getProperty (Id::currentBranch).toString() }) });
    });
}

void TITView::registerRemoteActions()
{
    commands.add (map::CommandId::addRemote, [this] (const juce::CommandID&)
    {
        const auto isConfig { static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::menuConfig };

        if (static_cast<int> (getGit().getProperty (Id::remote)) == map::Remote::noRemote)
        {
            state.setProperty (Id::commandLines, text::Git::remoteAdd, nullptr);
            setInputMode (map::EditorMode::remoteUrl, isConfig ? text::Repository::newRemoteUrl : text::Repository::remoteUrl, {});
        }
        else
        {
            commandManager->invokeDirectly (isConfig ? map::CommandId::removeRemote : map::CommandId::pullRebase, false);
        }
    });

    commands.add (map::CommandId::switchRemote, [this] (const juce::CommandID&)
    {
        state.setProperty (Id::commandLines, text::Git::remoteSetUrl, nullptr);
        setInputMode (map::EditorMode::remoteUrl, text::Repository::newRemoteUrl, {});
    });

    commands.add (map::CommandId::removeRemote, [this] (const juce::CommandID&)
    {
        setConfirmation (text::Confirmation::removeRemoteTitle, text::Confirmation::removeRemoteExplanation, text::Confirmation::removeRemoteYes, text::Confirmation::removeRemoteNo,
                         { text::Git::removeOrigin }, {});
    });
}
```

   Delete the `addRemote` entry of `registerPullActions` of Part 12 Step 12, item 1: the entry of `registerRemoteActions` replaces it.
   - Each family is one responsibility: the choice of a place, the submit of a name, the submit of an address, and the remote. No family is longer than 30 lines.
   - Each typed text comes from `InputView::getText` (Part 10 Step 14). The `commandLines` row holds `remote add` or `remote set-url`, as the row holds the lines of an input in Part 10 Step 16.
   - The run sets the mode `console` (Part 10, the `isRunning` event). Each submit therefore needs no mode change of its own.

4. The branch mismatch. In `registerEvents`, add:

```cpp
    events.add (Id::remoteHead, [this] (const juce::Identifier&, juce::ValueTree& git)
    {
        const auto remoteHead { git.getProperty (Id::remoteHead).toString() };
        const auto defaultBranch { remoteHead.fromFirstOccurrenceOf (text::Git::originPrefix, false, false) };
        const auto branch { git.getProperty (Id::currentBranch).toString() };
        const auto isStartup { static_cast<int> (state.getProperty (Id::mode)) == map::EditorMode::startup };

        if (not isStartup and defaultBranch.isNotEmpty() and defaultBranch.compare (branch) != 0)
            setConfirmation (text::Confirmation::branchMismatchTitle,
                             getFormattedText (text::Confirmation::branchMismatchExplanation, { branch, defaultBranch }),
                             getFormattedText (text::Confirmation::branchMismatchYes, { defaultBranch }),
                             getFormattedText (text::Confirmation::branchMismatchNo, { branch }),
                             { TITGitProcessor::getLine ({ text::Git::createBranch, defaultBranch, remoteHead }) }, {});
    });
```

   - The yes line is `git checkout -b <default> origin/<default>` (`tit/SPEC.md:812-815`). The no button runs nothing. SPEC, **Stay:** "Remote added normally"; "User stays on current branch" (`tit/SPEC.md:817-819`).
   - The second button is selected: the stay is the choice that changes nothing, as the second button of `pull_merge` (Part 10).
   - The text compare is `compare (…) != 0` (CODING.md).

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- Each command line is a cast row with a typed value. No line reads the output of another line.
- `TITView` creates no folder, deletes no file, and writes no file. It gives the typed text to the processor.
- New names and their families: `setRepositoryDirectory` (`getRepositoryDirectory`, Part 8 Step 6); `initRepository`, `cloneRepository` (legacy `InitializeRepository`, `legacy/git/init.go:48`; the command `clone`); `setLfsFilters` (legacy `SetupLFSFilters`, `legacy/git/lfs.go:36`); `getRemoteLines` (`getReturnLines`, Part 12); `getRepositoryName` (legacy `ExtractRepoName`); `isEmptyDirectory`, `isValidUrl` (legacy `isCwdEmpty`, `legacy/app/dispatch_base.go:16`; NAMES.md Rule 1); `minHttpUrlLength` (the legacy number of `legacy/ui/validation.go:84-92`). Local names: `entries`, `entry`, `suffix`, `address`, `folder`, `ignoreFile`, `clonePath`, `lines`, `git`, `isMissing`, `branch`, `name`, `url`, `isConfig`, `remoteHead`, `defaultBranch`, `isStartup`.

### Step 11: `TITGitProcessor` and `TITView`, the setup wizard
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`, `tit/Source/TITView.cpp`.
**Action:**
1. `TITGitProcessor`, the key work. Header, public section, after `isValidUrl` of Step 10: `static juce::File getSshKey();`, `static juce::String getPublicKey();`, `static bool addSshConfig();`. After `setLfsFilters`: `void addSshKey (const juce::String& email);`.

```cpp
juce::File TITGitProcessor::getSshKey()
{
    return juce::File::getSpecialLocation (juce::File::userHomeDirectory).getChildFile (files::sshDirectory).getChildFile (files::sshKey);
}

juce::String TITGitProcessor::getPublicKey()
{
    return getSshKey().withFileExtension (text::Git::publicKeyExtension).loadFileAsString().trim();
}

bool TITGitProcessor::addSshConfig()
{
    const auto config { getSshKey().getSiblingFile (files::sshConfig) };
    const auto text { config.loadFileAsString() };
    const auto newline { juce::String::charToString (Chars::newline) };

    if (text.contains (text::Setup::addKeysToAgent) and text.contains (text::Setup::identityFile)) return true;
    return config.appendText (newline + text::Setup::sshConfigBlock + newline, false, false, lineEnding);
}

void TITGitProcessor::addSshKey (const juce::String& email)
{
    const auto key { getSshKey() };
    juce::StringArray lines { getLine ({ text::Git::sshAdd, key.getFullPathName().quoted() }) };

    key.getParentDirectory().createDirectory();

    if (not key.existsAsFile())
        lines.insert (0, getLine ({ text::Git::sshKeygen, email.quoted(), text::Git::keyFileOption, key.getFullPathName().quoted(), text::Git::emptyPassphrase }));

    run (lines, {});
}
```

   - `getPublicKey` gives the text of the public key file, or the empty text when no such file exists (`juce::File::loadFileAsString`, doxygen `classjuce_1_1File`).
   - `addSshConfig` adds the block to `~/.ssh/config`: the file is the container (NAMES.md Rule 1, `add`). A second call finds the two lines and writes nothing.
   - `addSshKey` is the key chain: `ssh-keygen` only when the key file does not exist, as legacy skips it (`legacy/git/ssh.go`); then `ssh-add`. `ssh-keygen` with an existing file would ask a question on the terminal of titc, which is the stdin of the child on macOS (tracker, "The git side: child processes").
   - The key path and the email are each one argument in double quotes. A path on Windows holds backslashes; `CommandLineToArgvW` keeps a backslash that no quote follows (tracker, "The git side: child processes").

2. `TITView`, the step, the dialog, and the readers:

```cpp
bool TITView::isSetupActive (const juce::ValueTree& git) const
{
    return static_cast<int> (git.getProperty (Id::gitEnvironment)) != map::GitEnvironment::ready
        or static_cast<int> (state.getProperty (Id::setupStep)) >= map::SetupStep::generate;
}

void TITView::setSetup (const juce::ValueTree& git)
{
    const auto environment { static_cast<int> (git.getProperty (Id::gitEnvironment)) };
    const auto step { static_cast<int> (state.getProperty (Id::setupStep)) };

    if (environment == map::GitEnvironment::missingGit or environment == map::GitEnvironment::missingSsh) setPrerequisites (environment);
    else if (step == map::SetupStep::generate) setSetupStep (TITGitProcessor::getPublicKey().isNotEmpty() ? map::SetupStep::displayKey : map::SetupStep::error);
    else setups.get (step);
}

void TITView::setSetupStep (int step)
{
    state.setProperty (Id::setupStep, step, nullptr);
    setSetup (getGit());
}

void TITView::setSetupConfirmation (const juce::String& title, const juce::String& explanation, const juce::String& button)
{
    state.setProperty (Id::mode, static_cast<int> (map::EditorMode::setupWizard), nullptr);
    messageBox = jam::TerminalMessageBox::showAlert (title, explanation, button, this, [this] (int)
    {
        commandManager->invokeDirectly (map::CommandId::setupWizardEnter, false);
    });
}

void TITView::setPrerequisites (int environment)
{
    const auto newline { juce::String::charToString (Chars::newline) };

    setSetupConfirmation ({},
                          getParagraphs ({ juce::String { environment == map::GitEnvironment::missingGit ? text::Setup::gitMissing : text::Setup::gitFound }
                                               + newline + (environment == map::GitEnvironment::missingSsh ? text::Setup::sshMissing : text::Setup::sshFound),
                                           text::Setup::installTools }),
                          text::Setup::recheck);
}

void TITView::setKeyConfirmation()
{
    const auto key { TITGitProcessor::getPublicKey() };

    juce::SystemClipboard::copyTextToClipboard (key);
    setSetupConfirmation (text::Setup::keyTitle,
                          getParagraphs ({ text::Setup::keyCopied, key, text::Setup::providers, TITGitProcessor::addSshConfig() ? text::Setup::configWritten : text::Setup::configFailed }),
                          text::Setup::continueButton);
}

juce::String TITView::getParagraphs (const juce::StringArray& paragraphs)
{
    return paragraphs.joinIntoString (juce::String::repeatedString (juce::String::charToString (Chars::newline), paragraphNewlines));
}
```

   - The step after `generate` is the public key step when the public key file holds a text, and the error step when it does not. So the public key step always has a key to show, and it needs no second test of the file. Step 3 therefore holds no text for a public key that cannot be read.
   - `setKeyConfirmation` is the dialog of the public key step: the key goes to the clipboard and into the text, and the processor writes the `~/.ssh/config` block, after the key exists, as legacy writes it after the key (`legacy/git/ssh.go:88-92`).

3. The two tables of the steps:

```cpp
void TITView::registerSetups()
{
    setups.add (map::SetupStep::welcome, [this] (const int&)
    {
        setSetupConfirmation (text::Setup::welcomeTitle, text::Setup::welcomeBody, text::Setup::continueButton);
    });

    setups.add (map::SetupStep::email, [this] (const int&) { setInputMode (map::EditorMode::email, text::Setup::emailPrompt, {}); });
    setups.add (map::SetupStep::displayKey, [this] (const int&) { setKeyConfirmation(); });

    setups.add (map::SetupStep::complete, [this] (const int&)
    {
        setSetupConfirmation (text::Setup::completeTitle, text::Setup::completeBody, text::Setup::continueButton);
    });

    setups.add (map::SetupStep::error, [this] (const int&)
    {
        const auto output { state.getChildWithName (jam::Format::toType (Id::console)).getProperty (Id::text).toString().trim() };

        setSetupConfirmation (text::Setup::errorTitle, getParagraphs ({ output, text::Setup::errorHint }), text::Setup::goBack);
    });
}

void TITView::registerSetupSteps()
{
    setupActions.add (map::SetupStep::welcome, [this] (const int&) { setSetupStep (map::SetupStep::email); });
    setupActions.add (map::SetupStep::displayKey, [this] (const int&) { setSetupStep (map::SetupStep::complete); });
    setupActions.add (map::SetupStep::error, [this] (const int&) { setSetupStep (map::SetupStep::email); });

    setupActions.add (map::SetupStep::complete, [this] (const int&)
    {
        state.setProperty (Id::setupStep, static_cast<int> (map::SetupStep::welcome), nullptr);
        state.sendPropertyChangeMessage (Id::directory);
        setMenuMode (getGit());
    });
}
```

   `setups` holds what a step shows. `setupActions` holds what Enter does at a step.

4. `registerSetupActions()`:

```cpp
void TITView::registerSetupActions()
{
    commands.add (map::CommandId::setupWizardEnter, [this] (const juce::CommandID&)
    {
        const auto environment { static_cast<int> (getGit().getProperty (Id::gitEnvironment)) };

        if (environment == map::GitEnvironment::missingGit or environment == map::GitEnvironment::missingSsh)
            state.sendPropertyChangeMessage (Id::directory);
        else
            setupActions.get (static_cast<int> (state.getProperty (Id::setupStep)));
    });

    commands.add (map::CommandId::emailSubmit, [this] (const juce::CommandID&)
    {
        const auto email { commitMessage.getText() };

        if (email.contains (text::Setup::emailMark) and email.contains (text::Setup::domainMark))
        {
            state.setProperty (Id::setupStep, static_cast<int> (map::SetupStep::generate), nullptr);
            gitProcessor.addSshKey (email);
            state.setProperty (Id::message, text::Setup::generateTitle, nullptr);
        }
        else
        {
            state.setProperty (Id::message, text::Setup::emailInvalid, nullptr);
        }
    });
}
```

   The cancellation of the email prompt is in `registerCancellations()` (Step 9, item 12). A wizard dialog has no cancellation: Escape runs the one button of the box, as Enter does (Part 10 Step 14). At the error step that button goes back to the email, which is the legacy Escape of that step.
   - The steps follow `tit/SPEC.md:756-768` and `legacy/app/modes.go:46-52`. `prerequisites` is the environment row, and `generate` is the run of the console (Closed by a read, items 7 and 8).
   - Re-check runs the detection again. A new environment row reaches `setMenuMode` through the `gitEnvironment` event of Step 9.
   - After the run, Escape in the console calls `setMenuMode` (Part 11 Step 11, item 8). The step `generate` then opens the public key, or the error step when no key exists.
   - Complete runs the detection again and opens the menu of the state, as legacy (`legacy/app/setup_wizard_handlers.go`, Complete: "find repo, reload, ModeMenu").

**Validation:**
- No function has a bail-out guard. Each `if` chain has at most three branches. No function body is longer than 30 lines. No `==` on strings.
- The wizard state is the row `setupStep` and the `git` rows. `TITView` holds no step.
- `TITView` reads no file and writes no file. The key file, the public key, and `~/.ssh/config` are work of the processor.
- New names and their families: `getSshKey`, `getPublicKey` (the file `sshKey` and its public file); `addSshConfig`, `addSshKey` (NAMES.md Rule 1, `add`: the config file and the key agent are the containers; the tool word `ssh-add`); `setKeyConfirmation` (`setSetupConfirmation`). Local names: `environment`, `step`, `newline`, `config`, `text`, `key`, `lines`, `output`, `email`.

### Step 12: `TITGitProcessor`, `lfs checkout` after each hard reset
**Scope:** `tit/Source/TITGitProcessor.h`, `tit/Source/TITGitProcessor.cpp`, `tit/Source/TITView.cpp`.
**Action:**
1. Header, public section, after `getRemoteLines`: `static juce::StringArray getResetLines (const juce::ValueTree& git, const juce::String& resetLine);` and `static juce::StringArray getMergeAbortLines (const juce::ValueTree& git);`.

```cpp
juce::StringArray TITGitProcessor::getResetLines (const juce::ValueTree& git, const juce::String& resetLine)
{
    if (static_cast<bool> (git.getProperty (Id::isLfsReady))) return { resetLine, text::Git::lfsCheckout };
    return { resetLine };
}

juce::StringArray TITGitProcessor::getMergeAbortLines (const juce::ValueTree& git)
{
    juce::StringArray lines { text::Git::mergeAbort };

    lines.addArray (getResetLines (git, text::Git::resetHard));
    return lines;
}
```

2. Each hard reset of Parts 9 to 12 takes its lines from `getResetLines`. Change `getDiscardLines (changeSet)` of Part 12 Step 9 to `getDiscardLines (const juce::ValueTree& git, const juce::StringArray& changeSet)`; each of its callers passes the `git` child as the first argument.

   | Where | The lines of the reset now |
   |---|---|
   | `abortMerge`, `registerOperationActions`, Part 9 Step 7, item 6 | `gitProcessor.run (TITGitProcessor::getMergeAbortLines (model.state.getChildWithName (jam::Format::toType (Id::git))), {});` |
   | `getConflictAbortLines`, Part 12 Step 9, item 3, the merge | `return getMergeAbortLines (git);` |
   | `getConflictAbortLines`, the stash | `return getResetLines (git, text::Git::resetHardHead);` |
   | `getDiscardLines`, Part 12 Step 9, item 2 | `auto lines (getResetLines (git, text::Git::resetHardHead));`, then `lines.add (text::Git::clean);` and `lines.addArray (changeSet);` |
   | `resetDiscardChanges`, `registerWorkingTreeActions`, Part 10 Step 16, item 7, both dialogs | `TITGitProcessor::getDiscardLines (getGit(), {})` in place of `{ text::Git::resetHardHead, text::Git::clean }` |
   | `getReplaceLocalLines`, Part 10 Step 10 | `juce::StringArray lines { text::Git::fetchOrigin };`, `lines.addArray (getResetLines (git, juce::String { text::Git::resetHardOrigin } + git.getProperty (Id::currentBranch).toString()));`, `lines.add (text::Git::clean);`, `return lines;` |
   | `getRewindLines`, Part 11 Step 7 | `auto commandLines (getResetLines (git, getLine ({ text::Git::resetHard, hash })));` |
   | `getDirtyAbortLines`, Part 12 Step 9, item 2 | `juce::StringArray lines { getLine ({ text::Git::checkout, git.getProperty (Id::originalBranch).toString() }) };`, `lines.addArray (getResetLines (git, getLine ({ text::Git::resetHard, git.getProperty (Id::originalHead).toString() })));` |
   | `pin`, `registerPinActions`, Part 12 Step 12, item 1 | `const auto changeSet (TITGitProcessor::getResetLines (git, TITGitProcessor::getLine ({ text::Git::resetHard, hash })));` |
   | `unpin`, `registerPinActions`, Part 12 Step 12, item 1 | `const auto changeSet (TITGitProcessor::getResetLines (git, TITGitProcessor::getLine ({ text::Git::resetHard, original })));` |

   Each initializer of a `juce::StringArray` from a `juce::StringArray` uses parentheses (CODING.md, "initializer of the same container type").

**Validation:**
- `isLfsReady` is true only when the repository uses LFS, the LFS tool runs, and the filters are registered (Part 8 Step 7). So a repository with no LFS gets the lines of before.
- No other line of Parts 9 to 12 holds `resetHard`, `resetHardHead`, or `resetHardOrigin` outside `getResetLines`.
- `lfs checkout` after the reset of `replaceLocal` is the legacy line (`legacy/app/op_pull.go:89-96`). The other resets follow `PLAN-tit-2-spec.md` Step 7, item 4: "as it follows each hard reset".
- New names and their families: `getResetLines`, `getMergeAbortLines` (`getReturnLines`, Part 12; the command `abortMerge`).

### Step 13: The application, the user files, the theme, the start
**Scope:** `tit/Source/Main.cpp`.
**Action:**
1. `runWindow`: after the `setenv` lines of Part 9 Step 8 and Part 12 Step 13, add:

```cpp
        model.loadFromPath();
        model.startWatcher();
        applyTheme();
```

2. `TITView::setKeys` of Part 9 Step 7, item 3: replace the line `const auto config { jam::ConfigDocument::getOrCreate … };` with `const auto config { state.getChildWithName (keysSheet) };`. The sheet now comes from the `keys` child of the Model, which holds the user file (Step 5).
3. `runWindow`: after the line `model.state.setProperty (Id::directory, directory.getFullPathName(), nullptr);` of Part 8 Step 8, add `gitProcessor->update();`.
4. Add two private functions:

```cpp
    void applyTheme()
    {
        auto& desktop { juce::Desktop::getInstance() };

        styleManager.registerStyle (model.state.getChildWithName (jam::Format::toType (Id::theme)));
        lookAndFeel.setAppearance (Id::light);

        for (int index { 0 }; index < desktop.getNumComponents(); ++index)
            desktop.getComponent (index)->sendLookAndFeelChange();
    }

    void valueTreeChildAdded (juce::ValueTree& parent, juce::ValueTree& child) override
    {
        if (parent == model.state and child.hasType (jam::Format::toType (Id::theme)))
            applyTheme();
    }
```

   - The first load runs before the window, so the window paints with the user theme, and the first `setKeys` of `TITView` reads the user key sheet. The listeners are added later (Part 7 Step 7; Part 9 Step 7), so the first load gives no child event.
   - A later load replaces the two children (Step 5). The application reacts to the `theme` child, and `TITView` reacts to the `keys` child (Step 9, item 6). Each one reads the child that the event delivers.
   - `applyTheme` is the END theme event (`end/Source/lookAndFeel/ENDLookAndFeel.cpp:507-516`; `end/Source/Main.cpp:4-10`). `lookAndFeel` keeps the sheet of its construction, `StyleManager::getStyleSheet()`, which the theme tree also uses (Step 5: `getSheetType (files::styleSheet)`).
   - The update starts after the `directory` row, so its job reads the folder of the repository. Its end closes the gate of Step 9.

**Validation:**
- `runWindow` has no bail-out guard. The command-line runs of Part 1 read and write no user file: `loadFromPath` runs only in `runWindow`.
- New names and their families: `applyTheme` (`applyColours`, `end/Source/lookAndFeel/ENDLookAndFeel.cpp:474`); `desktop`, `index` (`end/Source/Main.cpp:4-10`); `parent`, `child` (JUCE `ValueTree::Listener`).

### Step 14: Build, run, and measure on macOS
**Scope:** JAM and the `tit/` project; fixture repositories and a fixture home folder outside every repository, for example in the session scratchpad.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build). The fixture commands are git commands: ARCHITECT runs them, or gives the git instruction (CAROL.md Git).
**Action:**
1. Build `ling` and EVE as in Part 10 Step 18, items 1 and 2. Step 1 changed one cast table of JAM.
2. From `tit/`, run `cast cast/spell.md --debug`.
3. Each fixture runs with `HOME` set to a new empty fixture folder, so that `~/.config/tit` and `~/.ssh` of the machine do not change. `juce::File::userHomeDirectory` reads `HOME` on macOS. The fixture numbers are those of Part 8 Step 9.

   | # | Fixture and keys | Expected |
   |---|---|---|
   | 1 | fixture 4, first start | `<HOME>/.config/tit/user.md`, `keys.md`, and `themes/gfx/theme.md` exist. Each equals its carried default. The colours equal Part 10. |
   | 2 | 1; in `user.md`, `history count` 2; start again; `h` | The history shows two commits. |
   | 3 | 1; in `user.md`, `auto update` `maybe` | Within one second the footer shows the error with the file and the line. The Preferences row shows `ON`. |
   | 4 | 1; `/`, `p`, Return | `user.md` holds `false` in the row `auto update`, in the table form of `jam::MarkdownWriter`. The row shows `OFF`. |
   | 5 | 4; `j`, `=` three times, `+`, `_` twice | The row shows `1 min` after the steps 8, 18, 8, and the limit 1. Each value is in `user.md`. On the row `Auto-update`, `=` does nothing. |
   | 6 | 1; copy `themes/gfx` to `themes/test`; in `test/theme.md`, change one colour of `HeaderView`; `/`, `p`, `j`, `j`, Return | The row shows `test`, and the header has the new colour. Return again: `gfx` and the colour of before. |
   | 7 | 1; in `keys.md`, the `commit` value `c` to `k` | Within one second `k` opens the commit box and `c` does nothing. |
   | 8 | 1; in `keys.md`, raise the version to `0.0.2` in the carried file, build, start | `keys.md` on disk is the new default. |
   | 9 | fixture 5 | The menu shows `Checking Remote...` with a turning spinner, then the menu of the state. |
   | 10 | 9; `git remote set-url origin /nonexistent` | The footer shows `Remote unreachable — showing cached state`. |
   | 11 | 9; `update interval` 1; a push to the remote from a second clone; no key for 90 s | The header shows `Syncing...`, `Checking local state...`, `Fetching remote updates...` with a turning spinner, then `Local behind`. |
   | 12 | 11 with a key every 20 s | No sync starts. |
   | 13 | 4, the history while the cache loads | The spinner of `History...` turns. |
   | 14 | an empty folder; `i`, `1`, Return | The console shows `init`, `checkout -b main`, `add -- .gitignore`, and the commit. `.gitignore` equals the carried file. The header shows `main`. |
   | 15 | an empty folder; `i`, `2`, `sub`, Return, Return | `sub/` holds the repository. The header shows the folder `sub`. |
   | 16 | an empty folder with `.DS_Store`; `c`, `1`, the path of `origin.git` of fixture 5, Return | The console shows the clone. The folder holds the repository; `.DS_Store` is gone. |
   | 17 | a folder with one file; `c` | The prompt `Repository URL:` shows at once. The path of `origin.git`: a new folder `origin` holds the clone. |
   | 18 | 17; `c`, `foo`, Return | The footer shows `Invalid URL format. Try: git@github.com:user/repo.git`. |
   | 19 | fixture 4 with no remote; `r`, the path of a new bare repository, Return | The console shows `remote add`, `fetch`, `remote set-head` with its failure, and `push -u origin`. The header shows the remote. |
   | 20 | fixture 4 on `master`; a bare remote whose default branch is `main`; `r`, its path, Return | After the run, the dialog `⚠️ Branch name mismatch` with `STAY ON MASTER` selected. `h`, Return: the branch `main` tracks `origin/main`. |
   | 21 | fixture 5; `/` | `Switch Remote`, `Remove Remote`, a separator, `Branch`, `Preferences`, a separator, `Back`. `r`, Return: no remote; the header shows `🔌 NO REMOTE`. |
   | 22 | 21 before the removal; `/`, `b` | The branch picker. Escape: the menu. |
   | 23 | fixture 19 of Part 8 with the filters not registered and `git-lfs` present | The console shows `git lfs install`; the footer shows `LFS filters installed automatically`. |
   | 24 | 23, `Behind`; `x`, `y`, Return | The console shows `lfs checkout` after the reset. |
   | 25 | an empty folder; `TIT_TEST_SETUP=1 titc` | `Welcome to TIT`. Return: the email prompt. `a`, Return: `Email must hold @ and .`. A valid address, Return: the console shows `ssh-keygen` and `ssh-add`. Escape: `SSH Key Generated` with the key; the clipboard holds it; `<HOME>/.ssh/config` holds the block. Return: `✓ Setup Complete!`. Return: the wizard starts again (Closed by a read, item 8). |
   | 26 | 25 with `<HOME>/.ssh` made a file | The console shows the failure. Escape: `✗ Setup Error` with the text of the tool. Escape: the email prompt. |
   | 27 | an empty folder; titc started with a `PATH` that holds no `git` | The dialog shows `✗ Git not found`, `✓ SSH found`, and the install text. Return runs the check again. |

4. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- Each build ends with exit code 0. Each row gives its expected result.
- These facts of this part are measured in the program: `getOrCreateFileDocument` writes an absent file; `MarkdownWriter` with one cell text writes the whole file with the new cell; the watcher sees a file of the `themes` subfolder; `symbolic-ref` after `set-head --auto`; `juce::TimedCallback` in a view on the terminal peer; `juce::SystemClipboard` with no native window.
- ARCHITECT sets the layout of the value column, of the startup row, of the wizard dialogs, and of the header during a sync at this run (ruling 7).
- The second `cast` run changes no generated file.

### Step 15: Run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 14, items 2 to 4, in END on a Windows machine with the rows 1, 4, 6, 7, 9, 14, 16, 19, and 25. Set `USERPROFILE` in place of `HOME`.
**Validation:**
- The results of Step 14 hold.
- `~/.config/tit` is `%USERPROFILE%\.config\tit`, and `getChildFile` takes the `/` of `.config/tit`.
- `ssh-keygen -f "<path>"` takes a path with backslashes as one argument, and `-N ""` as an empty argument (`CommandLineToArgvW`).
- `user.md` written by titc has `\r\n` line ends (`juce::File::replaceWithText` default), and the next load reads it.

## BLESSED Alignment

- **L — Lean, and E — Encapsulation, the layers.** The folders, the ignore file, the key files, and `~/.ssh/config` are work of `TITGitProcessor`. `TITView` gives it the typed text and tells it. Each family of commands is one function of 30 lines or fewer.
- **B — Bound.** `TITModel` owns the watcher; `TITView` owns its timer; each view owns its spinner timer. Each timer is the last member of its owner. The application owns the style manager, and the Model owns the theme tree that it registers.
- **L — Lean.** One function reads a user file in the JFS shape for the three files. One function writes a preference. One request fetches and detects for the gate and the auto-update. One function gives the reset lines. No theme manager (ruling 42), no settings file (ruling 44).
- **E — Explicit.** Each file name, command line, text, and limit is a named row or constant. An invalid file, a failed fetch, a wrong email, a wrong address each show a footer text. A command that the state does not permit is inactive.
- **S — SSOT.** The carried file is the default and the virgin disk file (ruling 46). The disk is the one source of a preference: a change goes to the disk and comes back through the watcher (ruling 45). The theme default is made from the one shipped `style.css`.
- **S — Stateless.** No view or processor holds a step, a frame, a value, or a phase. The wizard step, the clone folder, the update flag, and the last key time are Model rows.
- **E — Encapsulation.** The file reaches the Model through the watcher; the Model reaches the views and the application through listeners. `TITView` tells the processor to run or to update, and the Model tells the style manager and the command manager through the application.
- **D — Deterministic.** The same files give the same rows, the same key sheet, and the same theme. The same repository gives the same menus.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. Ruling 48 (ARCHITECT, session 2026-11-14): *"no open items"*. The three items of the draft close as follows:

1. **`tit/SPEC.md:770-776`, the check of `user.name` and `user.email`:** not built (Closed by a read, item 24).
2. **`tit/SPEC.md:825-832`, the bare repository:** not built (Closed by a read, item 24).
3. **The text `Email must hold @ and .`:** it stands (ruling 48). It is the one new text of this part. MANIFESTO E: "Prefer assert over silent fail"; legacy stays silent.

## The exact text of the new items of this part

Legacy, SPEC, and Part 2 hold no text for this item. Ruling 48: *"no open items"*.

| Where | Text |
|---|---|
| Footer, the email of the wizard has no `@` or no `.` | `Email must hold @ and .` |
