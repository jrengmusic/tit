# PLAN: TIT Part 7 — The Model, the Terminal Chain, and the First Run

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part changes TIT only. It builds three pieces and runs them for the first time.

1. **`TITModel` (ruling 11).** The state of TIT, a `jam::Model` in the LING and END form. This part adds only the rows that its own code reads: the `terminal` row with `winsize` and `numKeyPresses`.
2. **`TITTerminalProcessor` (ruling 30).** The terminal chain in the EVE form. A `jam::VirtualDevice` owns the serial context. The readiness callback of `TerminalDisplay` clocks the device. The tick reads the display into the key lane, and stores the counter row and the size row. The tick writes nothing to the terminal. The Model flushes on the message thread, and the application reacts through its event map. The frame bytes of the window go to `TerminalDisplay::write` on the message thread, through the chain event `terminalInput` (rulings 52 and 53).
3. **The first run (ruling 33).** `titc` with no argument on a terminal opens a `jam::TerminalWindow`. The window shows the help guide in a read-only `jam::TextEditor`, in the form of `printHelp`. The key sheet starts here, with one row: `quit`, `ctrl+c`. Part 9 adds the other rows.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 1, 2, 3, 4, and 5.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`. A non-template body is in the `.cpp`.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization; no out-parameter; no non-owning naked pointer as a member; "NEVER use raw `new` or `delete`"; no anonymous namespace; a file-local function is `static`.
- CODING.md, "Uniform signature in the type": the event map declares `void (juce::ValueTree&)`.
- CODING.md, "Owner first, narrowest scope, configure only the difference": the application does not restate a default of `juce::JUCEApplication`.
- MANIFESTO, Deadline threads: "A thread with a hard deadline, such as the audio thread or a clock thread, does not allocate, lock, wait, or make a blocking call." The tick of this chain is a deadline thread (ruling 52). It allocates nothing, and it makes no blocking call: it calls `TerminalDisplay::read`, which never blocks, and `getWinsize`. `audioDeviceAboutToStart` sizes every buffer of the tick.
- ARCHITECT, ruling 53: *"our hot path always lock free, message thread is free to run fully event driven"*; *"our pseudo renderer LowLevelGraphicsContext is Terminal Video writer, instead of writing pixel it outputs bytes to terminal"*. The frame write is renderer output on the message thread.
- ARCHETYPE.md §1: no step edits a generated file by hand (`archetype/ARCHETYPE.md:50-51`). Each cast table keeps its `+---` separator lines (`:41-43`).
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build: agents build only on ARCHITECT's explicit request.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`.

### Ruling 33, verbatim

Question: "Part 7 is planned to run TerminalWindow for the first time. Raw mode gives Ctrl+C to TIT as a key byte, not as a signal. A key reaches a command only through the jam::CommandManager key sheet, and that sheet is Part 9 (ruling 12). How does Part 7 handle its first run?" Selected: *"Part 7 starts the key sheet"*. Option text: "Part 7 adds the command wiring with one row: quit = ctrl + C (legacy app_keys.go:10). The window shows the help guide in a read-only TextEditor (the LING Help form). Part 9 adds the other rows. No code is thrown away later."

### The chain that this part follows: EVE

| Piece | Where |
|---|---|
| The members: `state`, `terminalTag`, `pendingInput`, `input`, `commands`, `device`, `tty` | `eve/Source/EVETerminalProcessor.h:344-359` |
| `jam::VirtualDevice device { jam::VirtualDeviceType::virtualDeviceTypeName, jam::VirtualDeviceType::virtualDeviceTypeName };` | `EVETerminalProcessor.h:355-358` |
| `std::unique_ptr<jam::TerminalTTY> tty { createTTY() };` and the static `createTTY()` with the platform choice | `EVETerminalProcessor.h:337,359`; `.cpp:123-139` |
| The readiness callback: `tty->onBytesReady = [this] (int numSamples) { device.clock (numSamples); };` | `EVETerminalProcessor.cpp:76-82` |
| `static constexpr int virtualBlockSize { 65536 };` | `EVETerminalProcessor.cpp:5` |
| `prepare()`: `virtualSampleRate` 48000.0; `virtualBufferSize` `virtualBlockSize` on Windows, 1024 on the other systems; `device.stop()`, `device.open ({}, {}, rate, size)`, `start()` | `EVETerminalProcessor.cpp:365-385` |
| `start()`: `device.start (this); device.requestProcess();` | `EVETerminalProcessor.cpp:387-391` |
| `releaseResources()`: `device.stop(); device.close();`; the destructor calls it | `EVETerminalProcessor.cpp:85-88,393-397` |
| The tick: `audioDeviceIOCallbackWithContext` runs `process()`; `audioDeviceStopped` is empty | `EVETerminalProcessor.cpp:399-402,425` |
| `audioDeviceAboutToStart` prepares each lane with `virtualBlockSize` | `EVETerminalProcessor.cpp:409-423` |
| The open on the device context: `tty->open (device.getContext(), bufferSize, ...)`. EVE has no assert there: it stores `shellExited` when the open fails | `EVETerminalProcessor.cpp:427-469`, lines `436`, `443-444`, `466` |
| EVE's `process()` writes its input lane to the PTY, a descriptor that EVE sets non-blocking (`jam_TerminalTTY_posix.cpp:178-186,341-356`). TIT's terminal output is a blocking descriptor (Part 3 Step 6), so TIT's tick has no write. | `EVETerminalProcessor.cpp:554-558` |
| EVE's own split: "the clock's tick reads and writes the PTY…; the message thread lays out and paints"; the tick never blocks | `eve/SPEC.md:37-39,949` |
| A lane with a counter row: write, `flush()`, then `storeValue` of the count + 1 | `EVETerminalProcessor.cpp:146-159` (`commands`, `numCommands`) |
| The chain event of the window bytes: `chainEvents[Id::terminalInput].add<const char*, int> (Id::write, [this] (const juce::Identifier&, const char* bytes, int size) { addInput (bytes, size); });` | `EVETerminalProcessor.cpp:278-281` |
| The frame path that exists: paint into a `TerminalGraphicsContext`, `process()`, `writeToStream`, then one write to stdout, all on the message thread | `ling/Source/Help.cpp:23-34` |
| The message-thread reader of a lane, called by the view on the counter row: `applyCommands (const std::function<...>& callback)` | `EVETerminalProcessor.h:118`; `.cpp:196-203`; `eve/Source/EVEView.cpp:148-151` |
| A row value read from the Model: `getWinsize()` unpacks the `winsize` row | `EVETerminalProcessor.cpp:360-363` |
| The rows: `winsize` int64, packed `jam::TerminalWinsize`; `numCommands` int; `numGridEvents` int64 | `eve/Source/layout/parameters.md:45,47,57` |

### The Model facts

| Fact | Where |
|---|---|
| `jam::Model` is a `juce::Timer`. The constructor starts it at 10 Hz. `timerCallback` calls `flush()` at 120 Hz after a tick that flushed, 60 Hz otherwise. | `jam_data_structures/model/jam_Model.cpp:76-83`; `jam_Model.h:1193-1194,1212-1213` |
| `storeValue (tag, id, value)`: any thread; it stores the atomic and notifies no listener | `jam_Model.h:1002-1019`; `jam_ParameterBase.h:228-234` |
| `flush()` writes a row to the tree only when the value differs; the write blocks the reverse sync | `jam_Model.cpp:29-58`; `jam_Model.h:922-928` |
| So a stored row reaches a `juce::ValueTree::Listener` on the message thread. A `ParameterAttachment` gets no call from a flush: its listener runs only from `setValue` (`jam_ParameterBase.h:221-226`), and the reverse sync does nothing while `flushToTree` writes (`jam_Model.cpp:14-21,45-46`). The reaction to a stored row is the listener map of `EVEView` (`EVEView.cpp:138-144`), not an attachment. | read of the rows above |
| `createAndAddParameter<ParameterType> (tree, name, defaultValue)`; the group of a row is the type of its tree | `jam_Model.h:1052-1101`; `:1201-1204` |
| `getParameter<T> (tag, id)` | `jam_Model.h:988-1000` |
| The model class of a JAM application: `LINGModel : public jam::Model, ...`, constructor `jam::Model (jam::Format::toType (Id::ling))`, rows made with `getOrCreateChildWithName` and `createAndAddParameter` | `ling/Source/LINGModel.h:15-17`; `LINGModel.cpp:3-16`; `end/Source/end/ENDModel.cpp:3-10` |
| `TITModel` is the name (tracker, Closed by a read) | `PLAN-tit-0-tracker.md` |

### The device facts

| Fact | Where |
|---|---|
| `VirtualDevice::start` calls `audioDeviceAboutToStart` at once, on the calling thread | `jam_audio_devices/audio_io/jam_VirtualDevice.cpp:37-47` |
| `VirtualDevice::stop` sets the device not playing and cancels a queued `requestProcess` | `jam_VirtualDevice.cpp:49-60` |
| `clock (numSamples)` runs the tick only while the device plays | `jam_VirtualDevice.cpp:76-82` |
| `getContext()`: the serial context, a serial GCD queue on macOS, a one-thread pool on Windows | `jam_VirtualDevice.h:99-102` |
| The module `jam_audio_devices` depends on `juce_audio_devices` and `jam_core` | `jam_audio_devices/jam_audio_devices.h:10-11,28-31` |
| `juce_audio_devices` depends on `juce_audio_basics` and `juce_events`; macOS frameworks CoreAudio, CoreMIDI, AudioToolbox | JUCE `juce_audio_devices/juce_audio_devices.h:54-55` |
| The module rows of the same author: `audioDevices` = `juce_audio_devices`; `@user-module` `jam_audio_devices`, "JAM Audio Device Abstractions" | `eve/project-info.md:280,337`; `end/project-info.md:310,326` |

### The terminal facts (Parts 3, 4, 5)

| Fact | Where |
|---|---|
| `TerminalDisplay`: `onBytesReady`, `open (AsyncUpdater::Context, int bufferSize)`, `close`, `read`, `write`, `getWinsize`. A size change calls `onBytesReady (0)`. The backend destructor calls `close()`. | Part 3 Step 5 |
| `TerminalUnixDisplay`, `TerminalWindowsDisplay` | Part 3 Steps 6, 7 |
| `TerminalDisplay::write` is a blocking write. The owner calls it on the message thread only. | Part 3 Step 5 |
| `TerminalFIFO`: `prepare`, `getNumReady`, `getFreeSpace`, `write`, `flush`, `applyPendingData`; storage `jam::Array<char>` with `allocate (size, false)` and `get()` | `jam_terminal/processor/jam_TerminalFIFO.h:29-111,117`; `.cpp:5-14,60` |
| `TerminalKeyboard::applyKeyPresses (std::string_view bytes, bool hasMoreInput, keyCallback, pasteCallback)` returns the bytes used; the caller keeps the rest | Part 4 Step 2 |
| `TerminalWindow (std::unique_ptr<juce::Component>, const juce::String&)`, `bindToEvents (Registry::Events::mapped_type&)`, and the owner contract of seven steps | Part 5 Step 4; Part 5, "The contract with the owner" |
| `jam::Registry::Events` | `jam_gui/registry/jam_Registry.h:45,52` |
| `TerminalWinsize::pack()`, `unpack (int64_t)` | `jam_terminal/video/jam_TerminalWinsize.h:41-57` |

### The command facts

| Fact | Where |
|---|---|
| `JUCEApplication` registers `StandardApplicationCommandIDs::quit` and performs it with `systemRequestedQuit()`, which calls `quit()` | JUCE `juce_gui_basics/application/juce_Application.cpp:53,66-92` |
| A command that no target in the chain of the focused component handles goes to `JUCEApplication::getInstance()` | JUCE `juce_gui_basics/commands/juce_ApplicationCommandTarget.cpp:94-132` |
| The defaults of `JUCEApplication`: `moreThanOneInstanceAllowed` true, `anotherInstanceStarted` empty, `systemRequestedQuit` calls `quit()`, `suspended` and `resumed` empty, `unhandledException` `jassertfalse` | JUCE `juce_Application.cpp:47-58` |
| `JUCEApplication::initialiseApp` calls `initialiseMacMainMenu()` on macOS | JUCE `juce_Application.cpp:95-107` |
| The application of END: `class ENDApplication : public juce::JUCEApplication, private juce::ValueTree::Listener`, an `events` map, `registerEvents()`, `valueTreePropertyChanged` that dispatches through the map | `end/Source/Main.h:10-12,69-87`; `end/Source/Main.cpp:38-53` |
| The map type: `jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, void (juce::ValueTree&)>` | `end/Source/config/ConfigModel.h:62` |
| The dispatch on one row type: `if (tree.hasType (jam::Format::toType (Id::terminal)) and terminalEvents.contains (property)) terminalEvents.get (property, tree);` | `eve/Source/EVEView.cpp:138-144` |
| The reader-mode key input of `jam::TextEditor` calls `commandManager->invokeKey (key, editor.terminal, false)` on the shared `CommandManager` | `jam_gui/widgets/jam_TextEditor.cpp:604-632`; `jam_TextEditor.h:1300` |
| `setDocument` registers the editor commands and the default keys only when the shared manager has none | `jam_TextEditor.cpp:1130-1136`; `jam_TextEditor.h:196-202` |
| `jam::TextEditor::getDefaultKeys()` | `jam_TextEditor.h:217-224` |
| `CommandManager::setKeyPresses (keys)`: one child per binding; the child type names the command; `Id::value`, `Id::mode`, `Id::layer`. A row type and a command short name match as `Format::toType` forms. A row without a mode applies in every mode. | `jam_gui/commands/jam_CommandManager.h:25-42,67-75`; `.cpp:5-41` |
| The key sheet build: the default rows that the config does not name, then the config rows; every command target of the view registered; then `setKeyPresses` | `eve/Source/EVEView.cpp:466-487` |
| `jam::ConfigDocument::getOrCreate (juce::Identifier name)` reads the BinaryData resource `name`, memoized in `SharedDocuments` | `jam_markdown/document/jam_ConfigDocument.h:18-28` |
| `getValueTree (rootType)`: a table with a column outside `key`, `type`, `value`, `light`, `dark`, `description` gives one child per row, with a `value` property and a text property for each other column | `jam_ConfigDocument.h:145-157` |
| The root type of a config tree: `Id::config` | `end/Source/config/ConfigModel.cpp:3-6`; `jam/cast/core.md:61` |
| The key table form: `key`, `type`, `value`, `mode`, `description`; a binding is a `string` row | `eve/Source/layout/default-eve.md:44-59`; `end/Source/config/keys.md:17-22`; Part 2 Step 8 |
| No default key of JAM binds plain `ctrl+c` | the key tables of `jam/cast/terminal.md:1949-2063` hold `ctrl+u`, `ctrl+d`, `ctrl+v`, `ctrl+shift+c`, `ctrl+shift+v`, `ctrl+shift+j`, `ctrl+shift+k` |
| Legacy TIT quits on `ctrl+c` | `tit/___legacy___/internal/app/app_keys.go:10` |

### The help view facts

| Fact | Where |
|---|---|
| `printHelp` builds, in order: `jam::SharedInstance<jam::SharedDocuments> documents { std::in_place }`, `jam::Typeface typeface`, `const jam::GlyphArrangement glyphArrangement`, `jam::StyleManager styleManager { { BinaryData::fetcher }, jam::Strings { files::styleSheet } }`, `jam::StyleMarkdown styleMarkdown`; then `component.setLookAndFeel (&styleMarkdown)` before `setDocument (jam::MarkdownDocument::parse (helpText, files::lingHelp))` | `ling/Source/Help.cpp:5-19` |
| `StyleManager`, `GlyphArrangement`, `SharedDocuments` derive from `jam::Instance`. Two live objects of one type coexist: "the latest constructed wins (LIFO)". So `printHelp` and the application members do not collide. | doxygen `classjam_1_1StyleManager`, `classjam_1_1GlyphArrangement`, `classjam_1_1SharedDocuments`; `jam_core/utils/jam_Instance.h:13-18` |
| A component's LookAndFeel must outlive the component. `printHelp` declares `styleMarkdown` before `component`. END declares `lookAndFeel` before `window`. | `ling/Source/Help.cpp:12,16`; `end/Source/Main.h:57,68` |
| The application holds a shared document store in the END form: `jam::SharedInstance<jam::SharedDocuments> documents { std::in_place };` | `end/Source/Main.h:48` |
| Legacy TIT runs its program when `tit` has no argument | `tit/___legacy___/cmd/tit/main.go:10-40` |
| A test of a standard stream: `isTerminalOutput()`, `isatty (fileno (stdout))`, `_isatty (_fileno (stdout))` on Windows | `ling/Source/Main.cpp:68-75` |
| A run of the application in its own function: `runDaemon`, `runOperation`; the model listener is added in the run | `ling/Source/Main.cpp:341,389,420` |

### Words the framework already holds

| Word | Where |
|---|---|
| `Id::terminal`, `Id::winsize` | `jam/cast/terminal.md:94,103` |
| `Id::keys`, `Id::write` | `jam/cast/gui.md:49,99` |
| `Id::config`, `Id::mode` | `jam/cast/core.md:61,116` |

### Closed by a read

ARCHITECT can reverse each one.

1. **The rows of this part.** `TITModel` holds only the rows that this part reads. Part 1 set the rule: "No table holds a row that no code of this part reads (MANIFESTO L, YAGNI)" (Part 1 Step 7). The state axes and markers are Model rows of Part 8, which writes them. The config tables in `~/.config/tit/` belong to Part 13, which reads them. The tracker Status rows of Parts 7, 8, and 13 say this now.
2. **`jam::Instance<TITModel>` is not a base.** No code of this part calls `TITModel::getInstance()`. A later part adds the base with its first caller.
3. **The window run needs a terminal on stdin and on stdout.** `isTerminalInput()` joins `isTerminalOutput()` (NAMES.md Rule 5). With no argument off a terminal, `titc` prints the guide, as in Part 1. The chain then asserts the open of the display (Step 5, item 6; MANIFESTO E, assert at the entry point).
4. **The application derives from `juce::JUCEApplication`.** END has this base (`end/Source/Main.h:11`). It gives the `quit` command and its target with no TIT code. Part 1's `JUCEApplicationBase` had no command target.
5. **The window size comes from the `winsize` row.** `audioDeviceAboutToStart` stores the first size. The run reads the row before `addToDesktop`, in the form of EVE's `getWinsize` (`EVETerminalProcessor.cpp:360-363`), because `TerminalBuffer::prepare` asserts a height above 0 (Part 5 Step 3). Later sizes arrive through the `winsize` event.
6. **The listener starts after `addToDesktop`.** A flush can deliver `numKeyPresses` only when the peer exists, so the run adds the model listener last, in the LING form (`ling/Source/Main.cpp:341`).

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: The module and binary rows
**Scope:** `tit/project-info.md`, tables `## juce module`, `## user module`, `## binary`.
**Action:**
1. `## juce module`: after the `audioProcessors` row, add `| audioDevices | juce_audio_devices |`, in the two-column form of the table (`ling/project-info.md:104-113`).
2. `## user module`: after the `jam_terminal` row, add `| @user-module | jam_audio_devices | JAM Audio Device Abstractions |` (`eve/project-info.md:280`).
3. `## binary`: after the `styleSheet` row, add `| keys | Source/keys.md | Key sheet, embedded as BinaryData |`, in the form of the `styleSheet` row (`ling/project-info.md:184-185`).

**Validation:**
- Each new row has the cell values and the separator form of its cited row.
- No other row changes.

### Step 2: The words, the key table, and the guide
**Scope:** `tit/cast/identifiers.md`, `tit/cast/files.md`, `tit/Source/keys.md` (new), `tit/Source/HELP.md`.
**Action:**
1. `cast/identifiers.md`, table `## identifiers`, after the `help` row, three rows of type `@id` with an empty description:
   - `tit`, value `` `tit` ``: the type of the Model tree (family `ling`, `ling/cast/identifiers.md:41`).
   - `terminalInput`, value `` `terminalInput` ``: the chain event key of the window bytes (`eve/cast/identifiers.md:84`).
   - `numKeyPresses`, value `` `numKeyPresses` ``: the counter row of the key lane (family `numCommands`, `eve/cast/identifiers.md:24`).
   Change the brief to `@brief TIT vocabulary — the words of the titc command line and of the state tree.`, the brief of `ling/cast/identifiers.md:12`.
2. `cast/files.md`, table `## files`, after the `titcHelp` row: type `@string`, name `titcKeys`, value `` `keys.md` ``, description `Default key sheet.`. Change the brief to `@brief File names titc reads: help, keys.`.
3. `Source/keys.md`: write this text.

```markdown
---
version: 0.0.1
---

## keys

+------+--------+------------+------+-------------+
| key  | type   | value      | mode | description |
+======+========+============+======+=============+
| quit | string | ctrl+c     |      | Quit titc.  |
+------+--------+------------+------+-------------+
```

   The front matter is the form of `end/Source/config/keys.md:1-3`. The columns are the form of `eve/Source/layout/default-eve.md:44-46`. The `mode` column makes each row a child of the `keys` tree (`jam_ConfigDocument.h:145-153`). The empty mode applies the row in every mode (`jam_CommandManager.h:15`). The value is one token with no space: `setKeyPresses` splits a value at spaces (`jam_CommandManager.cpp:16`), and the JAM tables write `ctrl+u` (`jam/cast/terminal.md:1913`). `juce::KeyPress::createFromDescription ("ctrl+c")` gives the key that Part 4 decodes (JUCE `juce_gui_basics/keyboard/juce_KeyPress.cpp:183-236`; Part 4 Step 2, validation).
4. `Source/HELP.md`: replace the row of `titc` in the table "Command Line" with:

```markdown
| `titc`           | Opens titc on a terminal. Prints this guide otherwise. | 0    |
```

   Widen the separator and the other rows of the table to the new column width. After the line `On a terminal, titc clears the screen first.`, add the line `In titc, press Ctrl+C to quit.`

**Validation:**
- Each new word has its cited family. No word of JAM is declared again (`Id::terminal`, `Id::winsize`, `Id::keys`, `Id::write`, `Id::config`, `Id::mode`).
- `HELP.md` states only the behavior that this part builds.
- No table holds a row that no code of this part reads.

### Step 3: `TITModel`
**Scope:** `tit/Source/TITModel.h`, `tit/Source/TITModel.cpp` (both new).
**Action:**
1. The header: `#pragma once`, `#include <JuceHeader.h>`, `#include "generated/Generated.h"` (`ling/Source/LINGModel.h:1-3`). Declare:

```cpp
class TITModel : public jam::Model
{
public:
    TITModel();
    ~TITModel() override = default;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITModel)
};
```

2. The source: `#include "TITModel.h"`, then:

```cpp
TITModel::TITModel()
    : jam::Model (jam::Format::toType (Id::tit))
{
    auto terminalRow { getOrCreateChildWithName (jam::Format::toType (Id::terminal)) };

    createAndAddParameter<jam::Parameter<int64_t>> (terminalRow, Id::winsize, int64_t { 0 });
    createAndAddParameter<jam::Parameter<int64_t>> (terminalRow, Id::numKeyPresses, int64_t { 0 });
}
```

   The rows are the EVE rows `winsize` (int64) and `numGridEvents` (int64) in kind (`eve/Source/layout/parameters.md:47,57`). The model keeps its flush timer, as `LINGModel` and `ENDModel` do.

**Validation:**
- The form matches `LINGModel` (`ling/Source/LINGModel.h:15-17`; `.cpp:3-16`), without the members that this part does not read.
- New names and their families: `TITModel` (tracker, Closed by a read); `terminalRow` (`overlayRow`, `ling/Source/LINGModel.cpp:10`).

### Step 4: `TITTerminalProcessor`, the header
**Scope:** `tit/Source/TITTerminalProcessor.h` (new).
**Action:** `#pragma once`, `#include <JuceHeader.h>`, `#include "generated/Generated.h"`, `#include "TITModel.h"`. Declare:

```cpp
class TITTerminalProcessor : private juce::AudioIODeviceCallback
{
public:
    explicit TITTerminalProcessor (TITModel& modelToUse);
    ~TITTerminalProcessor() override;

    void prepare();
    void releaseResources();
    void applyKeyPresses (const std::function<void (const juce::KeyPress&)>& keyCallback,
                          const std::function<void (std::string_view)>& pasteCallback);

    jam::Registry::Events chainEvents;

private:
    void start();
    void process() noexcept;
    void audioDeviceIOCallbackWithContext (const float* const*, int, float* const*, int, int, const juce::AudioIODeviceCallbackContext&) override;
    void audioDeviceAboutToStart (juce::AudioIODevice*) override;
    void audioDeviceStopped() override;
    void registerChainEvents();

    static std::unique_ptr<jam::TerminalDisplay> createDisplay();

    TITModel& model;
    const juce::Identifier terminalTag { jam::Format::toType (Id::terminal) };
    std::string pendingKeyPresses;
    jam::TerminalFIFO keyPresses;
    jam::Array<char> readBuffer;
    jam::VirtualDevice device { jam::VirtualDeviceType::virtualDeviceTypeName, jam::VirtualDeviceType::virtualDeviceTypeName };
    std::unique_ptr<jam::TerminalDisplay> display { createDisplay() };

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITTerminalProcessor)
};
```

- `display` is the last member, after `device`. So the display closes first, and its readiness callback never clocks a destroyed device (the order of `device` and `tty`, `EVETerminalProcessor.h:355-359`).
- `model` is a reference, as EVE's `model` is (`EVETerminalProcessor.h:341`).
- No `jam::AsyncUpdater` base, no `restart`, no swap: TIT has no shell and no history rows (MANIFESTO L, YAGNI).
- No `input` lane, no `pendingInput`, no `addInput`, and no `applyPendingInput`. In EVE they carry the user input from the message thread to the tick, which writes the PTY. TIT writes no byte on the tick, so no lane goes from the message thread to the tick.

**Validation:**
- Each member and each function has its EVE sibling, except the key lane, which follows the `commands` lane.
- New names and their families: `TITTerminalProcessor` (`EVETerminalProcessor`); `modelToUse` (`EVETerminalProcessor.h:44`); `chainEvents`, `registerChainEvents`, `start`, `process`, `prepare`, `releaseResources`, `terminalTag`, `device` (EVE, verbatim); `keyPresses`, `pendingKeyPresses`, `applyKeyPresses` (the `commands`, `pendingInput`, and `applyCommands` family of EVE, `EVETerminalProcessor.h:344-359`, with the noun of `TerminalKeyboard::applyKeyPresses`, Part 4); `display`, `createDisplay` (`tty`, `createTTY`, with the class word of ruling 15); `readBuffer` (`jam_TerminalTTY_windows.h:329`).

### Step 5: `TITTerminalProcessor`, the source
**Scope:** `tit/Source/TITTerminalProcessor.cpp` (new).
**Action:** `#include "TITTerminalProcessor.h"`. Then:
1. The file constant `static constexpr int virtualBlockSize { 65536 };` (`EVETerminalProcessor.cpp:5`).
2. The constructor: `: model (modelToUse)`. Body: `registerChainEvents();`, then `display->onBytesReady = [this] (int numSamples) { device.clock (numSamples); };` (`EVETerminalProcessor.cpp:76-82`, without the profiler zone).
3. The destructor:

```cpp
TITTerminalProcessor::~TITTerminalProcessor()
{
    releaseResources();
}
```

   The form is EVE's destructor (`EVETerminalProcessor.cpp:85-88`). `device.stop()` cancels a queued tick (`jam_VirtualDevice.cpp:49-60`). The window is destroyed before the chain (Step 6, item 6), and its peer sends its last bytes, the exit from the full-screen modes, through the chain event while `display` is open (Part 5 owner contract, step 7). `display` then closes as the last member goes, and restores the saved terminal mode (Part 3 Step 6).
4. `createDisplay()`: the form of `createTTY()` (`EVETerminalProcessor.cpp:123-132`): `std::make_unique<jam::TerminalWindowsDisplay>()` under `#if JUCE_WINDOWS`, else `std::make_unique<jam::TerminalUnixDisplay>()`.
5. `prepare()`, `start()`, `releaseResources()`, `audioDeviceIOCallbackWithContext`, `audioDeviceStopped`: the bodies of `EVETerminalProcessor.cpp:365-397,399-402,425`, verbatim, without the profiler zones (`:376-378`, `:381-383`). `virtualSampleRate` and `virtualBufferSize` are constants inside `prepare()` (`:367-372`).
   `registerChainEvents()`, in the form of `EVETerminalProcessor.cpp:278-281`. The entry writes the frame to the display in the same call, on the message thread:

```cpp
void TITTerminalProcessor::registerChainEvents()
{
    chainEvents[Id::terminalInput].add<const char*, int> (Id::write, [this] (const juce::Identifier&, const char* bytes, int size)
    {
        const auto numBytes { display->write (bytes, size) };
        jassert (numBytes == size);
    });
}
```

   - The caller is the peer, inside `performAnyPendingRepaintsNow`, its constructor, and its destructor (Part 5 Step 2). Each one runs on the message thread.
   - The write returns when the terminal has the frame (Part 3 Step 5). So one frame is complete before the next paint starts, and no buffer holds a frame between the peer and the terminal (MANIFESTO, the 0 rules: no temporary container in transit).
   - The assert names a failed or short write at its entry point (MANIFESTO E: "Prefer assert over silent fail"). A blocking write to a terminal with no signal handler writes every byte (Part 3 Step 6, item 7).
   - The tick does not touch the output, and the message thread does not touch the input: `read` and `getWinsize` on the tick, `write` on the message thread. The backends share no member between the two sides (Part 3 Steps 6 and 7).
6. `audioDeviceAboutToStart`:

```cpp
void TITTerminalProcessor::audioDeviceAboutToStart (juce::AudioIODevice*)
{
    keyPresses.prepare (virtualBlockSize);
    readBuffer.allocate (virtualBlockSize, false);

    const auto isOpen { display->open (device.getContext(), device.getCurrentBufferSizeSamples()) };
    jassert (isOpen);

    if (isOpen)
        model.storeValue (terminalTag, Id::winsize, display->getWinsize().pack());
}
```

   The key lane and the read buffer are sized here, off the tick (MANIFESTO, Deadline threads). The open on the device context is the EVE form (`EVETerminalProcessor.cpp:436,466`), and the buffer size is an argument of that open. The assert with the positive test is the form of MANIFESTO E (`jassert (input.isValid()); if (input.isValid())`): Step 6 tests both streams before `runWindow`, so a failed open is a broken invariant.
7. `process()`:

```cpp
void TITTerminalProcessor::process() noexcept
{
    int numBytesWritten { 0 };

    for (auto numBytes { display->read (readBuffer.get(), keyPresses.getFreeSpace()) }; numBytes > 0;
         numBytes = display->read (readBuffer.get(), keyPresses.getFreeSpace()))
    {
        keyPresses.write (readBuffer.get(), static_cast<size_t> (numBytes));
        numBytesWritten += numBytes;
    }

    if (numBytesWritten > 0)
    {
        keyPresses.flush();
        model.storeValue (terminalTag, Id::numKeyPresses, model.getParameter<jam::Parameter<int64_t>> (terminalTag, Id::numKeyPresses)->getValue() + 1);
    }

    model.storeValue (terminalTag, Id::winsize, display->getWinsize().pack());
}
```

   - The tick has no write to the terminal (ruling 52). It makes two kinds of calls on the display: `read`, which never blocks (Part 3 Step 5), and `getWinsize`, one `ioctl` or one console query.
   - Each read asks for at most the free space of the lane. The free space is at most `virtualBlockSize`, the size of `readBuffer`. The loop ends when the display has no byte (`read` returns 0, Part 3 Step 5). So the Windows wait is armed again with no unread record (Part 3 Step 7).
   - The counter row follows the `numCommands` form (`EVETerminalProcessor.cpp:152-158`).
   - The size row is stored on each tick. A size change starts a tick (Part 3, `onBytesReady (0)`). An equal size gives no event, because the flush writes only a changed value (`jam_Model.cpp:41-48`).
   - The read into `readBuffer` is the OS boundary copy. The write into the lane is the ownership handoff copy (MANIFESTO, Permitted copies 1 and 2).
8. `applyKeyPresses` runs on the message thread:

```cpp
void TITTerminalProcessor::applyKeyPresses (const std::function<void (const juce::KeyPress&)>& keyCallback,
                                            const std::function<void (std::string_view)>& pasteCallback)
{
    keyPresses.applyPendingData ([this] (const char* data, int size) { pendingKeyPresses.append (data, static_cast<size_t> (size)); });

    const auto numBytes { jam::TerminalKeyboard::applyKeyPresses (pendingKeyPresses, keyPresses.getNumReady() > 0, keyCallback, pasteCallback) };

    pendingKeyPresses.erase (0, static_cast<size_t> (numBytes));
}
```

   - `pendingKeyPresses` keeps a cut key for the next call, so each key is a view over one contiguous buffer (MANIFESTO, Permitted copy 3; Part 4 Step 2).
   - `hasMoreInput` is true when the tick published more bytes during the call. Otherwise the tick read until the display had no byte, so a lone ESC is the Escape key (Part 4, Protocol facts).

**Validation:**
- The tick allocates nothing, takes no lock, and makes no blocking call: `read`, the lane `write`, `flush`, and `storeValue` use storage that `audioDeviceAboutToStart` sized. No line of `process()` calls `TerminalDisplay::write`.
- `process()` has no bail-out guard. Each conditional block is a positive nest.
- One writer and one reader for the key lane: the tick writes `keyPresses`, and the message thread reads it.
- `TerminalDisplay::write` has one caller, the chain event, on the message thread.
- Each verbatim body is identical to its EVE body, except the names of the members that this part renames.

### Step 6: The application, its base, and its members
**Scope:** `tit/Source/Main.cpp`.
**Action:**
1. Includes: after `#include "generated/Generated.h"`, add `#include "TITModel.h"` and `#include "TITTerminalProcessor.h"` (`ling/Source/Main.cpp:4-5` form).
2. After `isTerminalOutput()`, add `isTerminalInput()`: the body of `isTerminalOutput()` (`ling/Source/Main.cpp:68-75`) with `stdin` in place of `stdout`.
3. Change the class head to `class Application : public juce::JUCEApplication, private juce::ValueTree::Listener` (`end/Source/Main.h:10-12`).
4. Delete `moreThanOneInstanceAllowed`, `anotherInstanceStarted`, `systemRequestedQuit`, `suspended`, `resumed`, and `unhandledException`. Each body equals the `JUCEApplication` default (`juce_Application.cpp:47-58`), and project code configures only the difference (CODING.md). Keep `getApplicationName`, `getApplicationVersion`, `initialise`, and `shutdown`.
5. Add the destructor `~Application() override { model.removeListener (this); }` (`ling/Source/Main.cpp:288-291`).
6. Add the private members, in this order, before `JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Application)`:

```cpp
    jam::SharedInstance<jam::SharedDocuments> documents { std::in_place };
    jam::Typeface typeface;
    const jam::GlyphArrangement glyphArrangement;
    jam::StyleManager styleManager { { BinaryData::fetcher }, jam::Strings { files::styleSheet } };
    TITLookAndFeel lookAndFeel;
    jam::SharedInstance<jam::CommandManager> commandManager { std::in_place };
    TITModel model;
    std::unique_ptr<TITTerminalProcessor> processor;
    std::unique_ptr<jam::TerminalWindow> window;
    jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, void (juce::ValueTree&)> events;
```

   - The first five are the objects of `printHelp`, in its order (`ling/Source/Help.cpp:5-12`). `documents` has the END form (`end/Source/Main.h:48`).
   - `lookAndFeel` comes before `window`, so the LookAndFeel outlives the window and its editor (`end/Source/Main.h:57,68`). It comes after `styleManager`, because its constructor reads the sheet of the manager.
   - `TITLookAndFeel` is the one LookAndFeel of the TIT views (ruling 62: *"TITLookAndFeel : jam::StyleTerminal"*). Add `tit/Source/TITLookAndFeel.h` and `tit/Source/TITLookAndFeel.cpp` to the scope of this step, and `#include "TITLookAndFeel.h"` after `"TITTerminalProcessor.h"`. The header has `#pragma once`, `#include <JuceHeader.h>`, and `#include "generated/Generated.h"`, in the form of `end/Source/lookAndFeel/ENDLookAndFeel.h:5-7`:

```cpp
class TITLookAndFeel : public jam::StyleTerminal
{
public:
    TITLookAndFeel();
    ~TITLookAndFeel() override = default;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITLookAndFeel)
};
```

```cpp
TITLookAndFeel::TITLookAndFeel()
{
    setSheet (jam::StyleManager::getStyleSheet());
    setAppearance (Id::light);
}
```

   - `jam::StyleTerminal` sets the default colour of each indexed colour id in its constructor (`jam/jam_terminal/style/jam_StyleTerminal.h:14-29`; `jam_StyleTerminal.cpp:5-9`). `setSheet` and `setAppearance` are functions of `jam::StyleCustom`: the style takes the sheet of the manager, and `setAppearance` applies that sheet through `StyleManager::setAppearance (sheet, *this, lightOrDark)` (`jam/jam_style/jam_StyleCustom.h:237-246`; `jam_StyleCustom.cpp:244-254`). `StyleManager::getStyleSheet()` and `Id::light` are the two defaults of the constructor of `jam::StyleMarkdown` (`jam/jam_markdown_graphics/style/jam_StyleMarkdown.h:34`).
   - The class holds only what TIT changes (CODING.md: "Project code configures only the difference"). It has no paint function in this part. The sizes and the colours come from `Source/style.css`, whose unit is one cell (`ling/Source/style.css:6-9`), and Part 13 puts the user's file over it (ruling 62: *"from css to user config markdown, disk always win"*).
   - `window` comes after `processor`, so the window closes first, and its last bytes reach an open display through the chain event (Part 5 owner contract, step 7).
   - `commandManager` is the shared manager of the editor (`jam_TextEditor.h:1300`; `eve/Source/EVEView.h:408`).
7. In `initialise`, replace `setApplicationReturnValue (runCommandLine (getCommandLineParameterArray())); quit();` with:

```cpp
        const auto arguments (getCommandLineParameterArray());

        if (arguments.isEmpty() and isTerminalInput() and isTerminalOutput())
        {
            runWindow();
        }
        else
        {
            setApplicationReturnValue (runCommandLine (arguments));
            quit();
        }
```

   The form is `ling/Source/Main.cpp:336,354-358`.

**Validation:**
- The class has no override that equals a `JUCEApplication` default.
- `titc --version`, `titc --help`, `titc --wrong`, and `titc` off a terminal take the Part 1 path.
- New names and their families: `isTerminalInput` (`isTerminalOutput`); `events` (`end/Source/Main.h:71`); `documents`, `typeface`, `glyphArrangement`, `styleManager` (`ling/Source/Help.cpp:5-12`); `TITLookAndFeel`, `lookAndFeel` (ruling 62; `ENDLookAndFeel`, `end/Source/Main.h:57`); `commandManager` (`jam_TextEditor.h:1300`); `model`, `processor` (`ling/Source/Main.cpp:469-471`); `window` (`end/Source/Main.h:68`).

### Step 7: The window run, the events, and the commands
**Scope:** `tit/Source/Main.cpp`.
**Action:** Add four private functions to `Application`.
1. `registerEvents()`, in the END form (`end/Source/Main.cpp:47-53`):

```cpp
    void registerEvents()
    {
        events.add (Id::winsize, [this] (const juce::Identifier&, juce::ValueTree& terminal)
        {
            const auto winsize { jam::TerminalWinsize::unpack (static_cast<juce::int64> (terminal.getProperty (Id::winsize))) };

            window->setSize (winsize.cols, winsize.rows);
        });

        events.add (Id::numKeyPresses, [this] (const juce::Identifier&, juce::ValueTree&)
        {
            processor->applyKeyPresses ([this] (const juce::KeyPress& key) { window->getPeer()->handleKeyPress (key); },
                                        [this] (std::string_view text)
                                        {
                                            if (auto* target { window->getPeer()->findCurrentTextInputTarget() })
                                                target->insertTextAtCaret (jam::Strings::fromView (text));
                                        });
        });
    }
```

   - The `winsize` event reads the value that the event delivered (CODING.md, "Consume the event payload").
   - A key reaches the focused component through the JUCE peer (Part 5, owner contract, step 6). A paste reaches the focused text input. With no active text input, `findCurrentTextInputTarget()` returns null (`juce_ComponentPeer.cpp:291-301`), and the paste has no target.
2. `valueTreePropertyChanged`, the `EVEView` form (`eve/Source/EVEView.cpp:138-144`):

```cpp
    void valueTreePropertyChanged (juce::ValueTree& tree, const juce::Identifier& property) override
    {
        if (tree.hasType (jam::Format::toType (Id::terminal)) and events.contains (property))
            events.get (property, tree);
    }
```

3. `addCommands()`, the `EVEView::addCommands` form (`eve/Source/EVEView.cpp:466-487`), with the key table of Step 2 as the config:

```cpp
    void addCommands()
    {
        const auto keysSheet { jam::Format::toType (Id::keys) };
        const auto config { jam::ConfigDocument::getOrCreate (juce::Identifier { files::titcKeys })->getValueTree (jam::Format::toType (Id::config)).getChildWithName (keysSheet) };
        juce::ValueTree keys { keysSheet };

        for (const auto& defaultKey : jam::TextEditor::getDefaultKeys())
            if (not config.getChildWithName (jam::Format::toType (defaultKey.getType())).isValid())
                keys.appendChild (defaultKey.createCopy(), nullptr);

        for (const auto& configKey : config)
            keys.appendChild (configKey.createCopy(), nullptr);

        jam::Component::applyFunctionRecursively (window.get(), [this] (juce::Component* child)
        {
            commandManager->registerAllCommandsForTarget (dynamic_cast<juce::ApplicationCommandTarget*> (child));
        });

        commandManager->registerAllCommandsForTarget (this);
        commandManager->setKeyPresses (keys);
    }
```

   The application registers `quit` (`juce_Application.cpp:66-81`). The `quit` row matches the command short name `Quit` as a `Format::toType` form (`jam_CommandManager.h:67-75`).
4. `runWindow()`. The order follows the owner contract of Part 5, with the chain of this part:

```cpp
    void runWindow()
    {
        const BinaryData::Raw help { files::titcHelp };
        auto editor { std::make_unique<jam::TextEditor>() };

        processor = std::make_unique<TITTerminalProcessor> (model);

        editor->setLookAndFeel (&lookAndFeel);
        editor->setDocument (std::move (*jam::MarkdownDocument::parse (std::string_view (help.data, static_cast<size_t> (help.size)), files::titcHelp)));

        window = std::make_unique<jam::TerminalWindow> (std::move (editor), ProjectInfo::projectName);
        window->setLookAndFeel (&lookAndFeel);
        window->bindToEvents (processor->chainEvents[Id::terminalInput]);

        processor->prepare();

        const auto winsize { jam::TerminalWinsize::unpack (model.getParameter<jam::Parameter<int64_t>> (jam::Format::toType (Id::terminal), Id::winsize)->getValue()) };
        window->setSize (winsize.cols, winsize.rows);
        window->addToDesktop (0);
        window->setVisible (true);
        window->toFront (true);

        addCommands();
        registerEvents();
        model.addListener (this);
    }
```

   - The editor gets its LookAndFeel before its document, as in `printHelp` (`ling/Source/Help.cpp:17-19`). `setDocument` gives the read-only reader mode (`jam_TextEditor.cpp:1148-1150`).
   - The help bytes come from the binary resource, as `writeHelp()` reads them (Part 1 Step 10, item 4; `ling/Source/Main.cpp:84,90`).
   - `prepare()` opens the display and stores the first size (Step 5, item 6). The size row is read before `addToDesktop` (Closed by a read, item 5).
   - `toFront (true)` makes the peer grab the focus (Part 5 Step 3, items 5 and 6). The focus goes to the editor, as EVE's view gives its focus to its default component (`eve/Source/EVEView.cpp:74-81`).
   - The listener is added last (Closed by a read, item 6).

**Validation:**
- `runWindow` holds no bail-out guard and no raw `new`.
- No member is a non-owning pointer. The application owns the model, the chain, and the window; the window owns the editor and its peer.
- The key `ctrl + C` in the help view runs `quit`: the reader-mode key input calls `invokeKey` (`jam_TextEditor.cpp:629-632`), no editor target handles `quit`, and the command goes to the application (`juce_ApplicationCommandTarget.cpp:117-129`).
- New names and their families: `runWindow` (`runDaemon`, `runOperation`, `ling/Source/Main.cpp:389,420`); `registerEvents` (`end/Source/Main.cpp:47`); `addCommands` (`eve/Source/EVEView.cpp:466`); `keysSheet`, `config`, `keys`, `defaultKey`, `configKey` (`EVEView.cpp:468-476`); `help` (`binary`, `ling/Source/Main.cpp:84`, with the noun of `files::titcHelp`); `editor` (`EVEView.cpp:67`); `winsize`, `terminal` (`EVETerminalProcessor.cpp:362`; `EVEView.cpp:158`); `target`, `text`, `key` (JUCE `juce_ComponentPeer.cpp:291-301`; Part 4 Step 2).

### Step 8: Build and run on macOS
**Scope:** the `tit/` project.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:**
1. From `tit/`, run `cast cast/spell.md --debug`.
2. Run `titc --version`, `titc --help`, `titc --wrong`, and `titc < /dev/null`.
3. In END, run `titc`. Read the guide. Press cursor down, cursor up, page down, and page up. Resize the END pane. Paste a text. Press Ctrl+C.
4. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- The build ends with exit code 0. `Source/generated/Identifiers.h` holds `tit`, `terminalInput`, `numKeyPresses`; `Files.h` holds `titcKeys`.
- The four commands of item 2 give the outputs and exit codes of Part 1 Step 11, with the new `HELP.md` text.
- `titc` in END enters the alternate screen, hides the cursor, and shows the guide on every row of the pane, with no scroll (ruling 31).
- The guide scrolls with the four keys: reader mode gives a key with no sheet row to `viewport.keyPressed` (`jam_TextEditor.cpp:629-632`). A resize gives one frame at the new size.
- Ctrl+C quits with exit code 0. The shell gets its screen, its cursor, and its terminal mode back.
- No Dock icon and no menu bar appear (`juce_Application.cpp:95-107`).
- These facts of the tracker list "Facts that execution measures first" are measured: the attach path of a peer that is not a native view.
- The second `cast` run changes no generated file (`archetype/ARCHETYPE.md:478-479`).

### Step 9: Build and run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 8 in END on a Windows machine. Give `NUL` in place of `/dev/null`.
**Validation:**
- The results of Step 8 hold.
- The keys, the paste, and the resize work. This measures the record shape of VT input under ConPTY (tracker list "Facts that execution measures first"; Part 3 Step 7).

## BLESSED Alignment

- **B — Bound.** The application owns the style objects, the model, the chain, and the window. The declaration order sets the destruction order: the window, then the chain, then the model, then the LookAndFeel. The chain owns its device and its display; the display closes before the device goes.
- **L — Lean.** The model holds two rows. The key sheet holds one TIT row. The chain has no swap, no shell, and no restart, because TIT has none of them.
- **E — Explicit.** Each constant is named: `virtualBlockSize`, the EVE device rates. The run needs a terminal on stdin and stdout, and the application tests both.
- **S — SSOT.** The size is one Model row. The window reads it; no member keeps a copy. The `quit` command is the one of `JUCEApplication`. The key table is one file.
- **S — Stateless.** The application keeps no flag. The cut key bytes are look-ahead state of the parse, owned by the chain.
- **E — Encapsulation.** The chain knows no component. The window knows no chain type: it receives one event map. Input reaches the components through JUCE API only. Each direction flows one way. A key: display, key lane, Model row, event, peer. A frame: paint, the terminal graphics context, the VT lane, the chain event, display.
- **D — Deterministic.** The same bytes give the same key presses. The same size gives the same frame.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. Ruling 33 closes the first run. The runtime facts of this part are in Steps 8 and 9 and in the tracker list "Facts that execution measures first".
