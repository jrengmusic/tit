# PLAN: TIT Part 3 — The Program's Own Terminal and the Image Output

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part changes JAM only, in the module `jam_terminal`. It has two jobs.

1. **The image OUTPUT (rulings 24, 28, 29).** A component draws a `juce::Image`. `TerminalGraphicsContext::drawImage` adds an output placement that holds the PNG. The VT lane writes the image to the host terminal as an OSC 1337 sequence, with the cursor rules of the tested lane. The source is JAM revision `7fbf821ba`, restored verbatim. Only the names conform: the host class is `TerminalBuffer`, and the verbs are `write…`.
2. **`TerminalDisplay` (ruling 15).** The program's own terminal: raw input mode, a read that never blocks, a write, and the window size. A size change is a readiness event, as new input is (ruling 30). Two backends: `TerminalUnixDisplay` and `TerminalWindowsDisplay`. The shape mirrors `TerminalTTY`. The part also adds `TerminalEncoder::writePrivateMode` for the four DEC modes that a full-screen program sets.

The INPUT direction does not change: `TerminalProcessor`, the numbered `TerminalBuffer::ImagePlacement`, `TerminalImageRecord`, and the Preview component (ruling 29).
One file of `jam_tui` changes: `TerminalPage::writeToStream` writes the newline after the last row (ruling 31). Each help of a consumer goes through it (`tit/Source/Main.cpp:37-46`). ARCHITECT, session 2026-11-15: *"update the plan. i have already told you this explicitly on previous session jam::TerminalPage"*. Part 5 is the first user of the output lane. Part 7 is the first user of `TerminalDisplay`, through the terminal chain in the EVE form (ruling 30).

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Part 0.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`. Templates and `constexpr` stay in the header.
- CODING.md: a submodule header or source has no include. The module header holds every include. A platform header is the exception: `jam_TerminalTTY_posix.h:53-57`, `jam_TerminalTTY_posix.cpp:17-26`, `jam_TerminalTTY_windows.h:90`.
- CODING.md: "No near-duplicate struct proliferation". "If parameterization is needed, use a template." Step 2 applies it to the two placements.
- CODING.md: no bail-out guard; use `not`, `and`, `or`; brace initialization; `.at()`; no out-parameter; no Windows SDK macro name as an identifier.
- CAROL.md, Refactor-Rewrite Discipline: "Delete first, implement after". Step 1 deletes.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part. The doxygen pass is separate, after the audit. It updates the blocks of each changed member.
- CAROL.md, Build: agents build only on ARCHITECT's explicit request.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/jam/` unless the row says otherwise. "Old" means JAM revision `7fbf821ba`; an old file is in `jam_terminal/graphics/` at that revision. COUNSELOR or @Pathfinder reads an old body with `git show 7fbf821ba:<path>` (CAROL.md Git: read-only commands for the main session and Pathfinder only) and puts it in the brief to @Engineer. The steps below quote each body that a step takes.

### What Step 1 deletes

| Code | Why | Where |
|---|---|---|
| The body of `TerminalGraphicsContext::drawImage` | Ruling 28: the inline image lane is its only body | `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:488-509` |
| `jassert (placements.isEmpty());` in the VT lane | The VT lane now writes output placements | `jam_terminal/buffers/jam_TerminalBuffer.cpp:400` |
| The out-of-class body of `ImagePlacement::operator<` | The comparison moves into the template of Step 2 | `jam_terminal/buffers/jam_TerminalBuffer.cpp:247-250` |
| The declaration `struct ImagePlacement { ... };` | Step 2 makes it an alias of the template | `jam_terminal/buffers/jam_TerminalBuffer.h:45-57` |

### The OUTPUT lane at `7fbf821ba`, and its place today

| Piece | Old body | Old place | Place today |
|---|---|---|---|
| The switch | `explicit TerminalGraphicsEngine (bool newShouldUseInlineImages = false)`; `bool isUsingInlineImages() const noexcept`; `const bool shouldUseInlineImages` | `jam_TerminalGraphicsEngine.h:41,70,406`; `.cpp:265-268,394` | `TerminalBuffer` (`jam/carol/SPRINT-LOG.md:695`); today `TerminalBuffer()` is defaulted (`jam_TerminalBuffer.h:43`; `.cpp:5`) |
| The placement | `struct ImagePlacement { jam::Cell::Rectangle destination; juce::MemoryBlock png; bool operator< ... }` | `jam_TerminalGraphicsEngine.h:386-397`; `.cpp:444-447` | `TerminalBuffer::InlineImagePlacement` (Step 2) |
| The storage | `jam::Array<jam::Array<ImagePlacement>> imagePlacements`, one list for each ring slot | `jam_TerminalGraphicsEngine.h:422`; `.cpp:333,454,517,528` | `TerminalBuffer::inlineImagePlacements` (Step 2) |
| Adding a placement | `addImagePlacement (destination, image)`: `jassert (getBounds().contains (destination))`; `auto png { ImageLoader::getPng (image) }`; for each covered row, add `{ destination.withY (jam::Cell { destination.y - row }), row == destination.y ? std::move (png) : juce::MemoryBlock {} }`, then `sort()` | `jam_TerminalGraphicsEngine.cpp:429-442` | `TerminalBuffer::addInlineImagePlacement` (Step 2) |
| The row writer | `appendFrameRow`: for each placement, the span up to `destination.x`; on the first row (`destination.y == 0`) the image, then cursor up by `height - 1` when `height > 1`; on a later row cursor forward by `width`; the column moves to the right edge. Then the rest of the row and the row end. | `jam_TerminalGraphicsEngine.cpp:533-559` | `TerminalEncoder::writeRow` (Step 3) |
| The span writer | `appendRowText (output, row, cells, ...)`: SGR at the first index of the span or at a style change | `jam_TerminalGraphicsEngine.cpp:159-189` | `TerminalEncoder::writeRowSpan` (Step 3) |
| The sequence writer | `appendInlineImage (output, png, width, height)` | `jam_TerminalGraphicsEngine.cpp:249-263` | `TerminalEncoder::writeInlineImage` (Step 3) |
| The argument writer | `template <typename Value> appendImageArgument (output, key, value)`: key, `=`, value | `jam_TerminalGraphicsEngine.cpp:222-228` | `TerminalEncoder::writeImageArgument` (Step 3) |
| The cursor writers | `appendCursorUp`, `appendCursorForward` through `appendCsiSequence`, which asserts `parameter >= 0` | `jam_TerminalGraphicsEngine.cpp:13-21,212-220` | `TerminalEncoder::writeCursorUp`, `writeCursorForward` (Step 3) |
| The lane choice | `if (engine.isUsingInlineImages() and currentState.deviceSpaceClipList.containsRectangle (getClipRectangle (destination)) and transparencyDepth == 0 and currentState.fillType.colour.isOpaque()) addImagePlacement (destination, image);` | `jam_TerminalGraphicsContext.cpp:490-501` | `TerminalGraphicsContext::drawImage` (Step 4), with no `else` (ruling 28) |

What the restore does not take, and why:

| Old piece | Why not | Where |
|---|---|---|
| `TerminalGraphicsContext::addImagePlacement`, a one-line forward to the engine | A helper called from one site, with no responsibility boundary, is relocation (MANIFESTO L). `drawImage` calls `terminalBuffer.addInlineImagePlacement` directly. | old `jam_TerminalGraphicsContext.cpp:525-528` |
| `appendOscCommand`, `appendCsiSequence` | Today's writers write the introducer in place: `Chars::csiIntroducer` (`jam_TerminalEncoder.cpp:39,134-136`) and the OSC opener of `writeLink` (`:67-68`). Each new writer takes that form. | old `jam_TerminalGraphicsEngine.cpp:5-21` |
| `inlineImageHeaderBytes`, `getNumBase64Bytes` | No caller (tracker addendum). | old `jam_TerminalGraphicsEngine.h:374-383` |
| Marker placements, shell integration marks | Not in the OUTPUT lane of ruling 24. | old `jam_TerminalGraphicsEngine.cpp:230-247,561-592` |

### `TerminalBuffer` today

| Fact | Where |
|---|---|
| The INPUT placement: `destination`, `juce::int64 image` | `jam_TerminalBuffer.h:46-57` |
| `addImagePlacement (destination, image)` grows a short row to the right edge of the destination | `jam_TerminalBuffer.cpp:183-202`, lines `191-192`; `jam/carol/SPRINT-LOG.md:897` |
| `prepare` rotates the ring lists so the frame starts at slot 0 | `jam_TerminalBuffer.cpp:16-24` |
| `prepareRows` sizes the ring lists | `jam_TerminalBuffer.cpp:37-61` |
| `clear (writer)` clears the slots of a new frame | `jam_TerminalBuffer.cpp:252-261` |
| `swap` swaps two rows and their slot data | `jam_TerminalBuffer.cpp:319-329` |
| `clear (row)` blanks one row | `jam_TerminalBuffer.cpp:331-337` |
| `writeFrame` walks the frame for both lanes; its row writer takes the stream, the row, the INPUT placements, and the output block | `jam_TerminalBuffer.h:392-409` |
| The binary lane ignores nothing; the VT lane ignores the output block with an unnamed `juce::int64` parameter | `jam_TerminalBuffer.cpp:387-390,398` |
| The ring members | `jam_TerminalBuffer.h:443,447-450` |
| `TerminalGraphicsContext` takes a frame of `heightRows` slots and clears it at construction | `jam_TerminalGraphicsContext.cpp:7-14` |
| The one external constructor call: `jam::TerminalBuffer buffer;` | `jam_tui/page/jam_TerminalPage.cpp:23` (LSP references of `jam_TerminalBuffer.h:43`) |
| The one external call of the VT lane: `buffer.writeToStream (output, editor.getLookAndFeel());` | `jam_tui/page/jam_TerminalPage.cpp:36` (LSP references of `jam_TerminalBuffer.h:104`) |

### `TerminalEncoder` today

| Fact | Where |
|---|---|
| A namespace-struct; all members static | `jam_TerminalEncoder.h:17-25` |
| `writeRow (output, row, laf)`: style at index 0 or at a style change; links through `writeLinks`; codepoints, spacer tails skipped; then `writeLinks` at `row.size()`, SGR reset, newline | `jam_TerminalEncoder.cpp:110-130` |
| `writeLinks (output, row, index, linkRunIndex)` closes a run that ends at `index`, opens a run that starts at `index`, returns the next run index | `jam_TerminalEncoder.cpp:83-108` |
| `writeEraseScreen`: `Chars::csiIntroducer`, ED 2, `Chars::csiIntroducer`, CUP | `jam_TerminalEncoder.cpp:132-137` |
| The OSC opener: `ESC`, `]`, the number, `;` | `jam_TerminalEncoder.cpp:67-68` |
| `writeRow` has one caller, the VT lane | `jam_TerminalBuffer.cpp:401`; LSP references of `jam_TerminalEncoder.h:39` |
| A blank atom writes a space: "Codepoint 0 becomes Chars::space." | `jam::AttributedString::applyToCodepoints` (doxygen `classjam_1_1AttributedString.xml`) |
| The module header includes `TerminalEncoder.h` before `TerminalBuffer.h` | `jam_terminal/jam_terminal.h:98-99` |

### Words the framework already holds

| Word | Where |
|---|---|
| `map::OSC::iterm2Image` (1337) | `jam/cast/terminal.md:619` |
| `TerminalITerm2Decoder::filePrefix`, `inlineKey`, `sizeKey`, `widthKey`, `heightKey` (each a `std::string_view`), `inlineValue` (`int`, 1) | `jam_terminal/image/jam_TerminalITerm2Decoder.h:55-84` |
| `map::DEC::cursorVisible` (25), `alternateScreen` (1049), `bracketedPaste` (2004), `syncOutput` (2026) | `jam/cast/terminal.md:581,592-594` |
| `map::CSI::cursorUp`, `cursorForward`, `setMode`, `resetMode` | `jam/cast/terminal.md:707,709,733-734` |
| The family of a DEC mode: `## PrivateMode`, `## privateModes` | `jam/cast/terminal.md:1045,1316` |
| `Chars::question`, `Chars::equals`, `Chars::colon`, `Chars::bell`, `Chars::csiIntroducer` | `jam/cast/core.md:403,356,347,340,552` |
| `jam::ImageLoader::getPng (const juce::Image&)`: "Returns the PNG bytes of an image." | `jam_graphics/image_formats/jam_ImageLoader.h:104-113` |
| `juce::Base64::convertToBase64 (OutputStream&, const void*, size_t)` | JUCE `juce_core/misc/juce_Base64.h` (used by the old writer, `:261`) |
| `jam::AttributedChar::utf8MaxBytes` | `jam_TerminalEncoder.cpp:59` |

### The seam that `TerminalDisplay` mirrors

| Fact | Where |
|---|---|
| `TerminalTTY`: an abstract class, no thread, no lock. `std::function<void (int)> onBytesReady` runs on the serial context of `open()`. `read` never blocks. `write` makes one attempt. | `jam_terminal/tty/jam_TerminalTTY.h:1-30,53-60,116-137` |
| `setWinsize (jam::TerminalWinsize)`, the consumer side of the size | `jam_TerminalTTY.h:141-151` |
| The serial context: `jam::AsyncUpdater::Context`, a serial GCD queue on macOS, a one-thread pool on Windows | `jam_core/concurrency/jam_AsyncUpdater.h:65-73` |
| macOS readiness: `DISPATCH_SOURCE_TYPE_READ` on the fd, on the context; the handler suspends on 0 bytes, `isReadSourceSuspended`, a `cancelled` semaphore that the cancel handler signals; `close` resumes a suspended source, cancels, and waits | `jam_TerminalTTY_posix.h:29-38,174-199`; `.cpp:204-241,295-306` |
| Windows readiness: `readWait`, a thread pool wait on the pool that `open()` receives | `jam_TerminalTTY_windows.h:129-136,299-300` |
| The guards of the backends: `#if JUCE_MAC or JUCE_LINUX` with readiness on macOS only; `#ifdef JUCE_WINDOWS` | `jam_TerminalTTY_posix.h:51,174`; `jam_TerminalTTY_windows.h:88` |
| The module aggregates each backend `.cpp` under its platform guard | `jam_terminal/jam_terminal.cpp:29-36` |
| `TerminalWinsize`: `cols`, `rows`, `pixelWidth`, `pixelHeight`. "Display is the sole author." | `jam_terminal/video/jam_TerminalWinsize.h:9-13,34-39` |
| The Windows console output code page is set by the program, not by JAM: `SetConsoleOutputCP (CP_UTF8)` | `ling/Source/Main.cpp:324-326`; TIT takes the same lines in Part 1 Step 10 |

### Platform facts (official documents)

| Fact | Source |
|---|---|
| "the correct method is for an application to snapshot the current terminal state using the function tcgetattr(), setting raw mode with cfmakeraw() and the subsequent tcsetattr(), and then using another tcsetattr() with the saved state to revert to the previous terminal state." | Apple `cfmakeraw(3)` |
| `TCSAFLUSH`: "The change occurs after all output written to fildes has been transmitted to the terminal." "any input that has been received but not read is discarded." | Apple `tcsetattr(3)` |
| MIN = 0 and TIME = 0: "If no characters are available, read() shall return a value of zero, having read no data." | POSIX Base Definitions, §11.1.7; Apple `termios(4)` gives the same sentence |
| `TIOCGWINSZ`: "Put the window size information associated with the terminal in the winsize structure pointed to by ws." Fields `ws_row`, `ws_col`, `ws_xpixel`, `ws_ypixel`. | Apple `tty(4)`; xnu `bsd/sys/ttycom.h` |
| A dispatch source "does not install a signal handler or otherwise alter the behavior of signal delivery." The handler of a source "does not interrupt the current thread of execution". | libdispatch `dispatch_source_create(3)` |
| The default action of SIGWINCH is "discard signal". | Apple `signal(3)` |
| `DISPATCH_SOURCE_TYPE_READ`: `dispatch_source_get_data` gives the "estimated bytes available to read". | Apple `dispatch_source_get_data` |
| Console input handle: "When the input buffer is not empty, the state of a console input buffer handle is signaled." | learn.microsoft.com `readconsoleinput` |
| `ENABLE_VIRTUAL_TERMINAL_INPUT` converts input "into Console Virtual Terminal Sequences". `ENABLE_WINDOW_INPUT`: size changes "are reported in the console's input buffer" and are read "using the ReadConsoleInput function". `ENABLE_PROCESSED_INPUT`: "CTRL+C is processed by the system and is not placed in the input buffer." `ENABLE_LINE_INPUT` and `ENABLE_ECHO_INPUT` are line modes. | learn.microsoft.com `setconsolemode` |
| `ENABLE_VIRTUAL_TERMINAL_PROCESSING`: written characters "are parsed for VT100 and similar control character sequences". "Ensure ENABLE_PROCESSED_OUTPUT is set when using this flag." | learn.microsoft.com `setconsolemode` |
| `PeekConsoleInput`: "If no data is available, the function returns immediately." `ReadConsoleInput`: "The function does not return until at least one input record has been read." | learn.microsoft.com `peekconsoleinput`, `readconsoleinput` |
| `KEY_EVENT_RECORD`: `bKeyDown` is TRUE when the key is pressed; `uChar.UnicodeChar` is the "Translated Unicode character." | learn.microsoft.com `key-event-record-str` |
| `srWindow` holds "the upper-left and lower-right corners of the display window"; cols = `Right - Left + 1`, rows = `Bottom - Top + 1` | learn.microsoft.com `console-screen-buffer-info-str`; `console-virtual-terminal-sequences` (Samples) |

The record shape of VT input under ConPTY has no Microsoft source. It is in the tracker list "Facts that execution measures first".

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: Delete
**Scope:** `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp`, `jam_terminal/buffers/jam_TerminalBuffer.h`, `jam_terminal/buffers/jam_TerminalBuffer.cpp`.
**Action:** @Engineer deletes the four pieces of the table "What Step 1 deletes".
1. Delete the body of `drawImage` (`jam_TerminalGraphicsContext.cpp:490-508`). The function stays with an empty body until Step 4.
2. Delete `jassert (placements.isEmpty());` (`jam_TerminalBuffer.cpp:400`).
3. Delete `TerminalBuffer::ImagePlacement::operator<` (`jam_TerminalBuffer.cpp:247-250`).
4. Delete `struct ImagePlacement` with its doxygen block (`jam_TerminalBuffer.h:45-57`).

**Validation:** The four pieces do not exist. The compiler errors name only `ImagePlacement`; Step 2 resolves them.

### Step 2: The two placements and the OUTPUT storage in `TerminalBuffer`
**Scope:** `jam_terminal/buffers/jam_TerminalBuffer.h`, `jam_terminal/buffers/jam_TerminalBuffer.cpp`.
**Action:**
1. At the place of the deleted struct, declare one template and two aliases. The comparison is the body of the deleted `operator<` (`.cpp:247-250`), which is the old body too (old `.cpp:444-447`).

```cpp
template <typename Content>
struct Placement
{
    jam::Cell::Rectangle destination;
    Content image;

    bool operator< (const Placement& other) const noexcept
    {
        return std::tie (destination.x, destination.y) < std::tie (other.destination.x, other.destination.y);
    }
};

using ImagePlacement = Placement<juce::int64>;
using InlineImagePlacement = Placement<juce::MemoryBlock>;
```

   `ImagePlacement` keeps its name, its fields, and its order, so the INPUT code does not change. `InlineImagePlacement` is the old placement; its PNG field takes the family field name `image`.
2. The switch, verbatim from the old engine (`.h:41,70,406`; `.cpp:265-268,394`):
   - Replace `TerminalBuffer();` (`.h:43`) with `explicit TerminalBuffer (bool newShouldUseInlineImages = false);`. Replace `TerminalBuffer::TerminalBuffer() = default;` (`.cpp:5`) with the constructor that initializes `shouldUseInlineImages (newShouldUseInlineImages)`.
   - Add the private member `const bool shouldUseInlineImages;`. `addInlineImagePlacement` tests it (item 4). No getter exposes it. ARCHITECT, session 2026-11-15, on Auditor finding 4: *"surface layers never acknowledged any of rendering context, must be rendering engine agnostic … what's the BLESSED compliance approach?"*. MANIFESTO E rule 4: "the caller says 'do this'. The caller never says 'are you in state X? then I will do Y for you.'" The old `isUsingInlineImages()` is not restored.
3. Add the private member `jam::Array<jam::Array<InlineImagePlacement>> inlineImagePlacements;` after `imagePlacements` (`.h:447`). It holds one list for each ring slot, as the old `imagePlacements` did (old `.h:422`).
4. Add the public `void addInlineImagePlacement (jam::Cell::Rectangle destination, const juce::Image& image);` after `addImagePlacement` (`.h:141`). The body is the old `addImagePlacement` (old `.cpp:429-442`), with two conformed lines:
   - The list of a row is `inlineImagePlacements.at (getSlot (jam::Cell { row }))`, as at old `.cpp:437`.
   - Before the add, grow a short row to the right edge of the destination: the two lines of today's sibling, `jam_TerminalBuffer.cpp:191-192`.
   - The PNG encode and the loop are nested inside `if (shouldUseInlineImages)`, after the two asserts (item 2).
   - The first line of the body is `jassert (screen == map::Screen::normal);`. The list lives in the normal-screen ring slot, and `swap` and `clear (row)` handle it on the normal screen only (item 5). MANIFESTO D: "An assert at that single ownership boundary is verification". Auditor finding 8.
5. Keep the ring lists together, in each place where `imagePlacements` is handled for the ring:
   - `prepare`: rotate `inlineImagePlacements` with the other lists (`.cpp:20-23`).
   - `prepareRows`: `inlineImagePlacements.resize (numSlots)`, and `ensureStorageAllocated (placementsPerRow)` for each list (`.cpp:51,55-57`).
   - `clear (writer)`: `inlineImagePlacements.at (slot).clear()` (`.cpp:256-257`; old `.cpp:454`).
   - `swap`: on the normal screen, swap the two lists of the slots (`.cpp:324-328`; old `.cpp:515-519`).
   - `clear (row)`: on the normal screen, clear the list of the slot of the row (old `.cpp:526-530`).
6. `writeFrame` (`.h:392-409`): the row writer takes the row index `jam::Cell { row }` as its second argument, after the stream (ruling 31). It takes one more argument after the INPUT placements, `inlineImagePlacements.at (getSlot (jam::Cell { row }))`.
7. The binary lane (`.cpp:387-390`): its lambda takes the row index as an unnamed `jam::Cell` and the new argument as an unnamed `const jam::Array<InlineImagePlacement>&`, in the form of the unnamed `juce::int64` of the VT lane (`.cpp:398`).
8. The VT lane (`.cpp:393-404`): its lambda takes the row index as `rowIndex`, the INPUT placements as an unnamed parameter, and the OUTPUT placements as `inlinePlacements`. When `rowIndex.value > 0`, it first writes `static_cast<char> (Chars::newline)`. It then calls `TerminalEncoder::writeRow (stream, row, inlinePlacements, laf)`. So a newline separates two rows, and the frame ends without one (ruling 31). A newline on the last row of a host terminal scrolls it up one row: JAM's own emulator scrolls in `addLine` when the cursor is on the last row of the scrolling region (`jam_TerminalProcessor.cpp:635-643`).

**Validation:**
- `ImagePlacement` has the same fields and order as before. No INPUT call site changes (`TerminalProcessor`, `TerminalSwap`, `TerminalHistory`, the binary `writeRow`).
- `writeToStream (output, laf)` writes one newline between two rows and none after the last row. The binary lane writes the same bytes as before.
- Every place that handles `imagePlacements` for a ring slot also handles `inlineImagePlacements`.
- The only external constructor call, `jam_tui/page/jam_TerminalPage.cpp:23`, compiles with no change. With the default `false`, no consumer that keeps the default writes an image (ruling 28).
- New names and their families: `Placement` (`ImagePlacement`, `placementsPerRow`, `placementHeaderBytes`, `jam_TerminalBuffer.h:46,60,63`); `InlineImagePlacement`, `inlineImagePlacements`, `addInlineImagePlacement` (`ImagePlacement`, `imagePlacements`, `addImagePlacement`, with the old words `shouldUseInlineImages`, `isUsingInlineImages`); `Content` (CODING.md: "Template parameters: Use descriptive names").

### Step 3: The writers in `TerminalEncoder`
**Scope:** `jam_terminal/buffers/jam_TerminalEncoder.h`, `jam_terminal/buffers/jam_TerminalEncoder.cpp`, `jam_terminal/jam_terminal.h`, `jam_tui/page/jam_TerminalPage.cpp`.
**Action:**
1. In `jam_terminal.h`, move `#include "buffers/jam_TerminalEncoder.h"` (`:98`) to the line after `#include "buffers/jam_TerminalBuffer.h"` (`:99`). `writeRow` now names `TerminalBuffer::InlineImagePlacement`. `jam_TerminalBuffer.h` names no `TerminalEncoder` type.
2. Change `writeRow` to `static void writeRow (juce::OutputStream& output, const jam::AttributedString& row, const jam::Array<TerminalBuffer::InlineImagePlacement>& placements, const juce::LookAndFeel& laf);`. The body is the old `appendFrameRow` (old `.cpp:533-559`) in today's form:

```cpp
void TerminalEncoder::writeRow (juce::OutputStream& output, const jam::AttributedString& row, const jam::Array<TerminalBuffer::InlineImagePlacement>& placements, const juce::LookAndFeel& laf)
{
    int column { 0 };
    int linkRunIndex { 0 };

    for (const auto& placement : placements)
    {
        linkRunIndex = writeRowSpan (output, row, { column, placement.destination.x }, linkRunIndex, laf);

        for (int index { placement.destination.x }; index < placement.destination.getRight().value; ++index)
            linkRunIndex = writeLinks (output, row, index, linkRunIndex);

        if (placement.destination.y == 0)
        {
            writeInlineImage (output, placement.image, placement.destination.width, placement.destination.height);

            if (placement.destination.height > 1)
                writeCursorUp (output, placement.destination.height - 1);
        }
        else
        {
            writeCursorForward (output, placement.destination.width);
        }

        column = placement.destination.getRight().value;
    }

    linkRunIndex = writeRowSpan (output, row, { column, row.size() }, linkRunIndex, laf);
    writeLinks (output, row, row.size(), linkRunIndex);

    output << Chars::csiIntroducer << static_cast<int> (map::SGR::reset)
           << static_cast<char> (map::CSI::selectGraphicRendition);
}
```

   The loop over the covered columns keeps the link run index in step with `writeLinks` (`.cpp:83-108`). It writes no text: the image covers those cells.
   The row ends with the SGR reset only. The VT lane writes the newline between rows (Step 2, item 8; ruling 31).
3. Add the private `static int writeRowSpan (juce::OutputStream& output, const jam::AttributedString& row, juce::Range<int> columns, int linkRunIndex, const juce::LookAndFeel& laf);`. The body is today's loop of `writeRow` (`.cpp:114-123`) over `columns`. The style test is `index == columns.getStart() or not jam::AttributedChar::isSameStyle (...)`, as at old `.cpp:181`. It returns the link run index.
4. Add the private `static void writeInlineImage (juce::OutputStream& output, const juce::MemoryBlock& png, int width, int height);`. The body is the old `appendInlineImage` (old `.cpp:249-263`) in today's form:
   - The OSC opener with `map::OSC::iterm2Image`, in the form of `writeLink` (`.cpp:67-68`).
   - `output.write (TerminalITerm2Decoder::filePrefix.data(), TerminalITerm2Decoder::filePrefix.size());`
   - `writeImageArgument` with `inlineKey` and `inlineValue`, then `Chars::semicolon`; `sizeKey` and `static_cast<int> (png.getSize())`, then `Chars::semicolon`; `widthKey` and `width`, then `Chars::semicolon`; `heightKey` and `height`.
   - `Chars::colon`, then `juce::Base64::convertToBase64 (output, png.getData(), png.getSize())`, then `Chars::bell`.
5. Add the private template, in the header (CODING.md, templates stay in the header). It is the old `appendImageArgument` (old `.cpp:222-228`):

```cpp
template <typename Value>
static void writeImageArgument (juce::OutputStream& output, std::string_view key, Value value)
{
    output.write (key.data(), key.size());
    output << static_cast<char> (Chars::equals) << value;
}
```

6. Add the private `static void writeCursorUp (juce::OutputStream& output, int numRows);` and `static void writeCursorForward (juce::OutputStream& output, int numColumns);`. Each asserts `jassert (numRows >= 0)` or `jassert (numColumns >= 0)`, as the old `appendCsiSequence` did (old `.cpp:15`). Each writes `Chars::csiIntroducer`, the count, and `static_cast<char> (map::CSI::cursorUp)` or `map::CSI::cursorForward`, in the form of `writeEraseScreen` (`.cpp:132-137`).
7. Add the public `static void writePrivateMode (juce::OutputStream& output, int mode, int finalByte);`. It asserts `finalByte == map::CSI::setMode or finalByte == map::CSI::resetMode`. It writes `Chars::csiIntroducer`, `static_cast<char> (Chars::question)`, `mode`, and `static_cast<char> (finalByte)`. A caller writes, for example, `writePrivateMode (output, map::DEC::alternateScreen, map::CSI::setMode)`.
8. In `jam_tui/page/jam_TerminalPage.cpp`, after `buffer.writeToStream (output, editor.getLookAndFeel());` (`:36`), add `output << static_cast<char> (Chars::newline);`. The help of each consumer then ends with a newline, as it did before this part (ruling 31).

**Validation:**
- A row with no placement gives the same bytes as before Step 3: one span over `{ 0, row.size() }` with the style at index 0, then the SGR reset. The newline that ended each row before Step 3 is now written between rows by the VT lane, and after the last row by `TerminalPage::writeToStream`. So the help of each consumer gives the same bytes.
- Each restored body matches its old body, except the `write…` verb, the stream form of today's writers, and `std::string_view` in place of the old `std::size` of a char array.
- New names and their families: `writeRowSpan`, `writeInlineImage`, `writeImageArgument`, `writeCursorUp`, `writeCursorForward` (the old `appendRowSpan`, `appendInlineImage`, `appendImageArgument`, `appendCursorUp`, `appendCursorForward`, with the verb of ARCHITECT's ruling, `jam/carol/SPRINT-LOG.md:672`); `writePrivateMode` (`## PrivateMode`, `## privateModes`, `jam/cast/terminal.md:1045,1316`); `numRows`, `numColumns` (`numRows` in `TerminalBuffer::scrollUp`, `jam_TerminalBuffer.h:250`).

### Step 4: The OUTPUT lane choice in `drawImage`
**Scope:** `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp`.
**Action:** Write the body of `drawImage`. It is the old body (old `jam_TerminalGraphicsContext.cpp:490-501`), with `terminalBuffer` in place of `engine`, with no `else` (ruling 28), and with no test of the buffer's switch: the buffer tests its own switch (Step 2, item 2):

```cpp
void TerminalGraphicsContext::drawImage (const juce::Image& image, const juce::AffineTransform& t)
{
    const auto destination { getQuantizedCells (image.getBounds().toFloat().transformedBy (t)) };

    if (currentState.deviceSpaceClipList.containsRectangle (getClipRectangle (destination))
        and transparencyDepth == 0
        and currentState.fillType.colour.isOpaque())
        terminalBuffer.addInlineImagePlacement (destination, image);
}
```

**Validation:**
- The body has no second lane.
- The clip test keeps the precondition of `addInlineImagePlacement`: the clip lies inside the grid (`jam_TerminalGraphicsContext.cpp:19`), so a destination inside the clip lies inside the grid.

### Step 5: `TerminalDisplay`
**Scope:** `jam_terminal/display/jam_TerminalDisplay.h` (new).
**Action:** Write the abstract class in the form of `TerminalTTY` (`jam_TerminalTTY.h:53-152`):

```cpp
class TerminalDisplay
{
public:
    virtual ~TerminalDisplay() = default;

    std::function<void (int)> onBytesReady;

    virtual bool open (AsyncUpdater::Context context, int bufferSize, double maximumPeriodMilliseconds) = 0;
    virtual void close() = 0;
    virtual int read (char* buf, int maxBytes) = 0;
    virtual int write (const char* buf, int len) = 0;
    virtual jam::TerminalWinsize getWinsize() const = 0;
};
```

The contract of each member:
- `onBytesReady (numSamples)`: the backend calls it on `context` when input is ready, and with `numSamples` 0 when the size of the terminal changes. The owner sets it before `open()`. As `TerminalTTY::onBytesReady` (`jam_TerminalTTY.h:60`). In EVE's chain this callback clocks the device (`EVETerminalProcessor.cpp:76-82`); the tick then reads the bytes and the size (ruling 30).
- `open (context, bufferSize, maximumPeriodMilliseconds)`: takes raw input mode, prepares the backend's `jam::TerminalClock` with `bufferSize` and `maximumPeriodMilliseconds`, and starts the readiness and the size change sources on `context`. It returns `false` when the input or the output is not a terminal. Both parameters are as in `TerminalTTY::open` (`jam_TerminalTTY.h:76-79,95`).
- The readiness rhythm is EVE's (`eve/SPEC.md:45-50`): "Each readiness event is one tick; its payload is the tick's numSamples … After the tick, `jam::TerminalClock` waits one period on the clock's context; the macOS backend then re-arms readiness …, and the Windows backend's next overlapped read is the re-arm." ARCHITECT, session 2026-11-15: *"why should we reinvent new pattern? why TIT must be different? what's wrong with EVE?"*. Each backend owns `jam::TerminalClock clock`, as `TerminalUnixTTY` does (`jam_TerminalTTY_posix.h:197-198`).
- `close()`: stops the sources, waits for a running callback, and restores the mode that `open()` saved. The destructor of each backend calls it, as `TerminalUnixTTY` does (`jam_TerminalTTY_posix.h:76-82`).
- `read (buf, maxBytes)`: never blocks. It returns the bytes read, 0 when none are ready, or -1 on an error. As `TerminalTTY::read` (`jam_TerminalTTY.h:115-125`).
- `write (buf, len)`: writes to the terminal output, and returns when the terminal has the bytes. It returns the bytes written, or -1 on an error. The output descriptor is a blocking descriptor, so the owner calls `write` on the message thread only, never on the tick (ruling 52; MANIFESTO, Deadline threads: no blocking call). The frame is renderer output, and the message thread paints it and sends it, as `writeHelp` writes its frame (`tit/Source/Main.cpp:37-46`) and as the JUCE Linux peer sends its pixels inside `performAnyPendingRepaintsNow` (`juce_Windowing_linux.cpp:495-554`). `read` and `getWinsize` are the two members that the tick calls.
- `getWinsize()`: returns the size now. The owner calls it for the first bounds of the window, and on each tick that `onBytesReady` starts. The owner stores the packed size into a Model row; the flush changes the row only when the value differs (`jam_Model.h:922-928`), so an equal size gives no event.

**Validation:**
- The class holds no state and no thread.
- New names and their families: `TerminalDisplay` (ruling 15); `onBytesReady`, `open`, `close`, `read`, `write` (`TerminalTTY`); `getWinsize` (`TerminalTTY::setWinsize`, `jam_TerminalTTY.h:151`).

### Step 6: `TerminalUnixDisplay`
**Scope:** `jam_terminal/display/native/jam_TerminalDisplay_posix.h`, `jam_terminal/display/native/jam_TerminalDisplay_posix.cpp` (both new).
**Action:**
1. The header: guard `#if JUCE_MAC or JUCE_LINUX`. Platform includes `<termios.h>`, and `<dispatch/dispatch.h>` under `#if JUCE_MAC`, as `jam_TerminalTTY_posix.h:51-57`. Declare `class TerminalUnixDisplay final : public TerminalDisplay` with the five overrides and `~TerminalUnixDisplay() override;`.
2. The private members:
   - `std::optional<struct termios> savedMode;` — the mode that `open()` saves. `close()` restores it only when it holds a value, then resets it. The destructor calls `close()` (item 9), and a restore with no saved mode would write a zero mode to the terminal. The sibling releases only what it holds (`jam_TerminalTTY_posix.cpp:293`).
   - Under `#if JUCE_MAC`: `readSource`, `resizeSource`, `isReadSourceSuspended`, `cancelled`, and `jam::TerminalClock clock`, with the types and deleters of `jam_TerminalTTY_posix.h:189,192,195,198`.
   - Under `#if JUCE_MAC`: the private `void addSources (jam::AsyncUpdater::Context context);`. It creates and activates the two sources of item 4. `open` calls it, as `TerminalUnixTTY::open` calls `addReadSource` (`jam_TerminalTTY_posix.cpp:269-272`). MANIFESTO §L: "30 lines per function". Auditor finding 5.
3. The source file: platform includes `<sys/ioctl.h>`, `<unistd.h>`, `<signal.h>`, at the place of the platform includes of `jam_TerminalTTY_posix.cpp:17-19`. That file includes `<fcntl.h>`, `<sys/ioctl.h>`, and `<sys/wait.h>` there; the display needs `<unistd.h>` for `read` and `write`, and `<signal.h>` for `SIGWINCH`.
4. `open (context, bufferSize)`:
   - `tcgetattr (STDIN_FILENO, &mode) == 0 and isatty (STDOUT_FILENO) != 0` gives the result: Step 5 says `open()` "returns `false` when the input or the output is not a terminal". Each step below is nested inside that positive test. The first step stores `mode` in `savedMode`.
   - Copy `savedMode`. Call `cfmakeraw` on the copy. Set the output flags of the copy back to `savedMode.c_oflag`, because the VT lane writes `Chars::newline` between two rows (Step 2, item 8) and the saved output mode moves that newline to column 0. Set `c_cc[VMIN]` and `c_cc[VTIME]` to 0, so `read()` never blocks (POSIX §11.1.7). Apply it with `tcsetattr (STDIN_FILENO, TCSAFLUSH, &raw)`.
   - Under `#if JUCE_MAC`: create `readSource` with `DISPATCH_SOURCE_TYPE_READ` on `STDIN_FILENO` and `context`. Its event handler takes `dispatch_source_get_data` as `numSamples`. The handler is the tested handler of `TerminalUnixTTY::addReadSource` (`jam_TerminalTTY_posix.cpp:210-237`): it suspends the source and sets `isReadSourceSuspended`, calls `clock.process (numSamples)`, `onBytesReady (numSamples)` and `clock.wait()`, and resumes the source only when `numSamples` is greater than 0 or the source is cancelled. Before `addSources`, `open` calls `clock.prepare (bufferSize, maximumPeriodMilliseconds)`, as `jam_TerminalTTY_posix.cpp:265-266`. So a closed input does not spin (the rule of `jam_TerminalTTY_posix.h:35-36`).
   - Under `#if JUCE_MAC`: create `resizeSource` with `DISPATCH_SOURCE_TYPE_SIGNAL`, `SIGWINCH`, and `context`. Its event handler calls `onBytesReady (0)` (ruling 30). The program changes no signal disposition: the default action of SIGWINCH is "discard signal", and a source does not "alter the behavior of signal delivery".
   - Both cancel handlers signal `cancelled`. Activate both sources. The form is `addReadSource` (`jam_TerminalTTY_posix.cpp:204-241`).
   - On Linux, `bufferSize` and `maximumPeriodMilliseconds` have no use, as the Linux note of `jam_TerminalTTY_posix.h:99-100` says for the TTY.
5. `close()`: when `readSource` is not null, do the steps of `jam_TerminalTTY_posix.cpp:296-305` for each source: cancel, resume when suspended, wait on `cancelled`, reset. Then, when `savedMode` holds a value, `tcsetattr (STDIN_FILENO, TCSAFLUSH, &savedMode.value())` and reset `savedMode`.
6. `read (buf, maxBytes)`: `::read (STDIN_FILENO, buf, maxBytes)`. A result above 0 gives the count. 0 gives 0. -1 gives -1.
7. `write (buf, len)`: `::write (STDOUT_FILENO, buf, len)`. The result gives the count, or -1. The descriptor keeps its blocking mode. The program installs no signal handler (item 4), so no signal cuts the call short.
8. `getWinsize()`: `ioctl (STDOUT_FILENO, TIOCGWINSZ, &size)` into a `struct winsize`. It returns `{ size.ws_col, size.ws_row, size.ws_xpixel, size.ws_ypixel }` as a `jam::TerminalWinsize`, the order of `jam_TerminalWinsize.h:36-39`.
9. The destructor calls `close()`.

**Validation:**
- After `close()`, the terminal mode equals the saved mode.
- No `fcntl` call changes the file status flags of stdin or of stdout, so a parent shell gets its terminal back unchanged.
- New names and their families: `TerminalUnixDisplay` (ruling 15); `readSource`, `isReadSourceSuspended`, `cancelled` (`jam_TerminalTTY_posix.h:189,192,195`); `resizeSource` (the `…Source` member of `readSource`); `addSources` (`addReadSource`, `jam_TerminalTTY_posix.h:186`); `savedMode` (Apple `cfmakeraw(3)`: "snapshot the current terminal state ... revert to the previous terminal state").

### Step 7: `TerminalWindowsDisplay`
**Scope:** `jam_terminal/display/native/jam_TerminalDisplay_windows.h`, `jam_terminal/display/native/jam_TerminalDisplay_windows.cpp` (both new).
**Action:**
1. The header: guard `#ifdef JUCE_WINDOWS`, platform include `<windows.h>`, as `jam_TerminalTTY_windows.h:88-90`. Declare `class TerminalWindowsDisplay final : public TerminalDisplay` with the five overrides and `~TerminalWindowsDisplay() override;`.
2. The private members:
   - `readWait`, with the type and deleter of the TTY's thread pool wait (`jam_TerminalTTY_windows.h:291`).
   - `jam::TerminalClock clock`, as the TTY backends own one.
   - `std::optional<DWORD> savedInputMode;` and `std::optional<DWORD> savedOutputMode;`. `close()` restores each only when it holds a value, then resets it, for the reason of Step 6 item 2.
   - `jam::Array<INPUT_RECORD> inputRecords;`, sized by `open()`.
   - No member holds a console handle. Each use calls `GetStdHandle (STD_INPUT_HANDLE)` or `GetStdHandle (STD_OUTPUT_HANDLE)`, because the handles are not owned (CODING.md, "No non-owning naked pointers as members").
3. `open (context, bufferSize)`:
   - `GetConsoleMode` on the input and on the output, into `savedInputMode` and `savedOutputMode`. Both succeed gives the result. Each step below is nested inside that positive test.
   - Input mode: `ENABLE_VIRTUAL_TERMINAL_INPUT | ENABLE_WINDOW_INPUT`. Line input, echo, and processed input are off, so each key and Ctrl+C arrive as input.
   - Output mode: `savedOutputMode | ENABLE_PROCESSED_OUTPUT | ENABLE_VIRTUAL_TERMINAL_PROCESSING`. `DISABLE_NEWLINE_AUTO_RETURN` stays as saved, so the newline between two rows moves to column 0 (Step 2, item 8).
   - `inputRecords.resize (bufferSize / jam::AttributedChar::utf8MaxBytes)`. One record gives at most `utf8MaxBytes` bytes.
   - Create `readWait` on the pool `context`, in the form that `TerminalWindowsTTY::open` creates its `readWait`: a `TP_CALLBACK_ENVIRON` bound to `context` with `SetThreadpoolCallbackPool`, then `CreateThreadpoolWait`, then `DestroyThreadpoolEnvironment` (`jam_TerminalTTY_windows.cpp:687-690,701,724`). Arm it on the console input handle with `SetThreadpoolWait`. Before it, `clock.prepare (bufferSize, maximumPeriodMilliseconds)`. Its callback takes the count of unread records from `GetNumberOfConsoleInputEvents` as `numSamples`, and calls `clock.process (numSamples)`, `onBytesReady (numSamples)` and `clock.wait()`, as the TTY readiness callback does. The callback does not arm the wait: `read()` arms it after it drains the records (item 5), because "the Windows backend's next overlapped read is the re-arm" (`eve/SPEC.md:49-50`).
4. `close()`: when `readWait` is not null, stop it and wait for a running callback: `SetThreadpoolWait (readWait.get(), nullptr, nullptr)`, `WaitForThreadpoolWaitCallbacks (readWait.get(), TRUE)`, `readWait.reset()`, the tested close form of the TTY backend. Then `SetConsoleMode` with each saved mode that holds a value.
5. `read (buf, maxBytes)`:
   - `PeekConsoleInputW` into `inputRecords`, for at most `juce::jmin (maxBytes / jam::AttributedChar::utf8MaxBytes, inputRecords.size())` records. It returns at once.
   - When the last record is a key-down record with a high surrogate, leave it unread, so a surrogate pair is read whole by the next call.
   - `ReadConsoleInputW` for that count. It does not wait, because the records exist. A failed peek or read returns -1 (Step 5).
   - A file-static `getUnicodeChar (const INPUT_RECORD&)` gives `uChar.UnicodeChar` of a key-down `KEY_EVENT` record, and 0 for any other record.
   - A key-down record with a non-zero `uChar.UnicodeChar` gives its character as UTF-8 into `buf`. A high surrogate pairs with the next record only when that record holds a low surrogate (`juce::CharacterFunctions::isLowSurrogate`); the pair gives one codepoint through `juce::CharPointer_UTF16`. One `juce::CharPointer_UTF8` writer over `buf` writes each codepoint, and its address gives the count (`juce_CharPointer_UTF8.h:75,261`). Auditor findings 2 and 6.
   - A `WINDOW_BUFFER_SIZE_EVENT` record gives no byte. The readiness callback counted it with the other records, so the owner's tick reads `getWinsize()` (ruling 30). The record's `dwSize` is the buffer size, not the window size (learn.microsoft.com `window-buffer-size-record-str`).
   - Any other record gives no byte.
   - It arms the wait on the console input handle again with `SetThreadpoolWait`, then returns the bytes written into `buf`.
6. `write (buf, len)`: `WriteFile` on the output handle. It returns the bytes written, or -1. The call is synchronous: it returns when the console has the bytes.
7. `getWinsize()`: `GetConsoleScreenBufferInfo` on the output handle. It returns `{ srWindow.Right - srWindow.Left + 1, srWindow.Bottom - srWindow.Top + 1, 0, 0 }`. The pixel fields are 0, as on the TTY backend (`jam_TerminalTTY_windows.h:249-250`).
8. The destructor calls `close()`.

**Validation:**
- After `close()`, both console modes equal the saved modes.
- No identifier is a Windows SDK macro name (CODING.md table).
- New names and their families: `TerminalWindowsDisplay` (ruling 15); `readWait` (`jam_TerminalTTY_windows.h:300`); `savedInputMode`, `savedOutputMode` (`savedMode` of Step 6); `inputRecords` (the Win32 type `INPUT_RECORD`, and `readBuffer`, `jam_TerminalTTY_windows.h:329`); `getUnicodeChar` (the `get` verb and the Win32 field `uChar.UnicodeChar`).

### Step 8: Wire the module
**Scope:** `jam_terminal/jam_terminal.h`, `jam_terminal/jam_terminal.cpp`.
**Action:**
1. In `jam_terminal.h`, after the TTY headers (`:106-109`), add `display/jam_TerminalDisplay.h`, `display/native/jam_TerminalDisplay_posix.h`, `display/native/jam_TerminalDisplay_windows.h`, in that order, as the TTY block is.
2. In `jam_terminal.cpp`, after the TTY block (`:29-36`), add the backend sources under the same guards: `display/native/jam_TerminalDisplay_windows.cpp` under `#if JUCE_WINDOWS`, `display/native/jam_TerminalDisplay_posix.cpp` under `#elif JUCE_MAC or JUCE_LINUX`.
3. In the module brief of `jam_terminal.h` (`:26-34`) and the include order note (`:35-41`), add the display after the TTY. This is the one doc text that this part changes: the module header is the include list.

**Validation:** Each new header and source is included once, under the guard of its platform. The submodule files hold no include except the platform headers of Steps 6 and 7.

### Step 9: Build and check the consumers
**Scope:** JAM through its consumers on macOS: `ling`, `cast`, `scout`.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:**
1. Before Step 1, save the outputs of `ling --help`, `cast --help`, and `scout --help` to files.
2. After Step 8, build each consumer with `cast cast/spell.md --debug`.
3. Run the three help commands again and compare with the saved outputs.
4. Do Steps 2 and 3 on a Windows machine for `ling`.

**Validation:**
- Each build ends with exit code 0.
- Each help output is byte-identical to its saved output when the help holds no image. When a help holds a diagram, its image cells are blank, because these consumers keep the switch `false` (ruling 28).
- The OUTPUT lane has no user before Part 5, and `TerminalDisplay` none before Part 7. Part 7 runs both.

## BLESSED Alignment

- **B — Bound.** `TerminalBuffer` owns the OUTPUT placements, slot by slot, with the rows. Each display backend owns its sources and its saved modes. `close()` releases them, and the destructor calls `close()`.
- **L — Lean.** The restore takes only the pieces with a caller. The display has five members and one callback. No cursor-position writer exists, because no caller needs one. `TerminalFIFO` does not change: no lane carries a frame.
- **E — Explicit.** `writePrivateMode` takes the mode and the final byte from `map::DEC` and `map::CSI`. Each restored value is a named word of the framework.
- **S — SSOT.** One template holds the shape of a placement. The PNG comes from `ImageLoader::getPng`, which keeps one encode for each image.
- **S — Stateless.** `TerminalEncoder` stays static. `TerminalDisplay` holds only the saved mode and the OS objects that it must release.
- **E — Encapsulation.** INPUT and OUTPUT stay apart (ruling 29). The OUTPUT never reads `TerminalImageRecord`. `TerminalGraphicsContext` tells `TerminalBuffer` to add a placement; it does not inspect it.
- **D — Deterministic.** A frame with the same rows and the same placements gives the same bytes.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. One fact has no Microsoft source: the record shape of VT input under ConPTY. It is in the tracker list "Facts that execution measures first", and Part 7 measures it.
