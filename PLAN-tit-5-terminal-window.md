# PLAN: TIT Part 5 — `TerminalWindow` and `TerminalComponentPeer`

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part changes JAM only, in the module `jam_gui`, beside `jam::Window`. It adds two classes.

1. **`TerminalWindow` (ruling 13).** A `juce::Component` that owns one content component. Its `createNewPeer` returns a `TerminalComponentPeer`, as `AudioProcessorEditor::createNewPeer` returns a peer that is not an OS window.
2. **`TerminalComponentPeer` (ruling 14).** A `juce::ComponentPeer` that speaks to a host terminal. One device unit is one cell. It paints the component with `handlePaint` into a `TerminalGraphicsContext`, and it hands the VT bytes of each frame to an event map, in the form of `TextEditor::bindToEvents`. It calls the vblank listeners and paints on a `juce::TimedCallback`, in the form of the JUCE Linux peer. The paint and the send of a frame run on the message thread, as in every JUCE peer. No frame byte enters a deadline thread (rulings 52 and 53).

Input and size enter the peer through JUCE API only: `handleKeyPress`, `findCurrentTextInputTarget()->insertTextAtCaret`, and `Component::setBounds`. Part 7 builds the terminal chain in the EVE form, and the chain calls them on the message thread (ruling 30). This part holds no TIT code.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Part 3 (`writePrivateMode`, the frame without a final newline) and Part 4 (the decoder that Part 7 runs).

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`. A submodule file has no include. `jam_gui.h` holds every include (`jam_gui.h:69-89`).
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization; no out-parameter; no non-owning naked pointer as a member; "NEVER use raw `new` or `delete`".
- CODING.md, "Owner first, narrowest scope": a value that one function uses lives in that function.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build: agents build only on ARCHITECT's explicit request.

## Dependency & API Inventory

JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. JAM paths are relative to `~/Documents/Poems/dev/jam/`.

### The JUCE attach path (read, vanilla JUCE 8)

| Fact | Where |
|---|---|
| `virtual ComponentPeer* createNewPeer (int styleFlags, void* nativeWindowToAttachTo);` is protected | `juce_gui_basics/components/juce_Component.h:2758-2761` |
| A JUCE class returns a peer that is not an OS window: `AudioProcessorEditor::createNewPeer` returns `juce_createUnityPeerFn (*this)` | `juce_audio_processors/processors/juce_AudioProcessorEditor.cpp:239-246` |
| `addToDesktop` calls `createNewPeer` (`:691`), `peer->updateBounds()` (`:698`), `peer->setVisible (isVisible())` (`:703`), `repaint()` (`:726`). It calls no `grabFocus`. | `juce_gui_basics/components/juce_Component.cpp:628-743` |
| `removeFromDesktop` deletes the peer: `delete peer;` | `juce_Component.cpp:758-762` |
| `Component::setBounds` writes `boundsRelativeToParent` (`:1111`), repaints when resized (`:1113-1116`), then calls `peer->updateBounds()` (`:1128-1130`) | `juce_Component.cpp:1090-1134` |
| `updateBounds` calls `setBounds (scaledScreenPosToUnscaled (component, component.getBoundsInParent()), false)` | `juce_gui_basics/windows/juce_ComponentPeer.cpp:84-87` |
| The peer constructor adds the peer to `Desktop::peers` only | `juce_ComponentPeer.cpp:41-49` |
| `handlePaint`: `Graphics g (contextToPaintTo)`; a scale only when `getBounds()` and the component size differ (`:122-131`); `component.paintEntireComponent (g, true)` (`:144`) | `juce_ComponentPeer.cpp:115-173` |
| `handleKeyPress (const KeyPress&)`: the focused component or the peer's component, then `keyPressed` up the parents | `juce_ComponentPeer.cpp:175-187,196-233` |
| `findCurrentTextInputTarget`: the focused `TextInputTarget` inside the component, when `isTextInputActive()` | `juce_ComponentPeer.cpp:291-301` |
| `handleMovedOrResized` compares `getBounds()` with the component bounds and changes nothing when they are equal | `juce_ComponentPeer.cpp:321-360` |
| `handleFocusGain` focuses the last focused component, or calls `component.grabKeyboardFocus()` | `juce_ComponentPeer.cpp:362-379` |
| `callVBlankListeners (double timestampSec)` is protected; the timestamp is "a monotonically increasing value expressed in seconds" | `juce_ComponentPeer.h:539-544,681` |
| The 26 pure virtuals: `getNativeHandle` (`:169`), `setVisible` (`:172`), `setTitle` (`:175`), `setBounds` (`:206`), `getBounds` (`:216`), `localToGlobal` (`:219`), `globalToLocal` (`:222`), `setMinimised` (`:262`), `isMinimised` (`:265`), `isShowing` (`:268`), `setFullScreen` (`:271`), `isFullScreen` (`:274`), `setIcon` (`:286`), `contains` (`:312`), `getFrameSizeIfPresent` (`:324`), `getFrameSize` (`:333`), `setAlwaysOnTop` (`:358`), `toFront` (`:361`), `toBehind` (`:364`), `isFocused` (`:373`), `grabFocus` (`:376`), `repaint` (`:422`), `performAnyPendingRepaintsNow` (`:430`), `setAlpha` (`:433`), `getAvailableRenderingEngines` (`:482`), and the private `textInputRequired` (`:703`) | `juce_gui_basics/windows/juce_ComponentPeer.h` |

### The JUCE peer form that this part follows: the Linux peer

| Piece | Where |
|---|---|
| `repaint (area)`: adds `area.getIntersection (getBounds().withZeroOrigin())` to `regionsNeedingRepaint` | `juce_gui_basics/native/juce_Windowing_linux.cpp:331-335,490-493,561` |
| `performAnyPendingRepaintsNow`: copies and clears `regionsNeedingRepaint`, builds a context, calls `peer.handlePaint (*context)`, then sends the pixels | `juce_Windowing_linux.cpp:495-554` |
| `onVBlank`: `callVBlankListeners (timestampSec)`, then the pending repaints | `juce_Windowing_linux.cpp:636-643` |
| `TimedCallback vBlankManager { [this]() { onVBlank(); } };` | `juce_Windowing_linux.cpp:661` |
| `toFront (makeActive)`: `setVisible (true)` and `grabFocus()` when `makeActive`; then `handleBroughtToFront()` | `juce_Windowing_linux.cpp:291-301` |
| `getFrameSize` from `getFrameSizeIfPresent` | `juce_Windowing_linux.cpp:170-174` |
| A capability the window does not have gets an empty body or `false`: `setAlpha`, `setAlwaysOnTop`, `textInputRequired` | `juce_Windowing_linux.cpp:370-372` |
| `using ComponentPeer::localToGlobal; using ComponentPeer::globalToLocal;` beside the float overrides | `juce_Windowing_linux.cpp:186-197` |
| `juce::TimedCallback`: a `juce::Timer` that runs a `std::function<void()>`; `startTimerHz` | `juce_events/timers/juce_TimedCallback.h:47-69` (doxygen `classjuce_1_1TimedCallback`) |

### The JAM pieces

| Piece | Where |
|---|---|
| `Window (std::unique_ptr<juce::Component> mainComponent, const juce::String& name, ...)`; the body takes ownership with `setContentOwned (mainComponent.release(), true)` | `jam_gui/windows/jam_Window.h:76`; `jam_Window.cpp:44` |
| The content member of a JUCE window: `juce::ResizableWindow::contentComponent` | `juce_gui_basics/windows/juce_ResizableWindow.h:403` |
| `TextEditor::bindToEvents (Registry::Events::mapped_type& callbacks)` sets `terminalInput = [&callbacks] (const char* bytes, int size) { callbacks.get (Id::write, static_cast<const char*> (bytes), static_cast<int> (size)); };` | `jam_gui/widgets/jam_TextEditor.cpp:102-108` |
| `std::function<void (const char*, int)> terminalInput;` "Writes the bytes of the user input to the terminal." | `jam_gui/widgets/jam_TextEditor.h:1286` |
| `Registry::Events` is `jam::HashMap<juce::Identifier, Function::Map<juce::Identifier, eventCapacity>>` | `jam_gui/registry/jam_Registry.h:45,52` |
| `Id::write` is a word of `cast/gui.md`, so only `jam_gui` names it | `cast/gui.md:99` |
| `jam_gui` depends on `jam_terminal`; `jam_terminal` does not depend on `jam_gui` | `jam_gui/jam_gui.h:17,83`; `jam_terminal/jam_terminal.h:10-17` |
| The window headers and sources of `jam_gui` | `jam_gui.h:123-127`; `jam_gui/jam_gui.cpp:20-27` |
| `explicit TerminalGraphicsContext (TerminalBuffer& terminalBufferToUse, Colours& trueColoursToUse);` It takes a write handle of `heightRows` slots and clears them in its constructor, and it holds no handle after | `jam_terminal/graphics/jam_TerminalGraphicsContext.h:84-87`; `.cpp:7-20` |
| One frame to a stream: `prepare (cols, rows, rows)`, the context, the paint, `process()`, `writeToStream (frame, laf)`; `*jam::Colours::getInstance()` | `ling/Source/Help.cpp:8-34` |
| `TerminalBuffer::prepare (newCols, newRows, transitRows)` asserts `transitRows > 0` | `jam_terminal/buffers/jam_TerminalBuffer.h:143-161`; `.cpp:7-11` |
| `TerminalBuffer::read()` takes every ready slot except the frame; the slots become free when the handle goes out of scope | `jam_TerminalBuffer.h:277-284` |
| `TerminalBuffer::writeToStream (output, laf)`: the VT lane, a newline between rows and none after the last row | `jam_TerminalBuffer.h:94-104`; Part 3 Step 2, item 8 (ruling 31) |
| `TerminalEncoder::writeEraseScreen`, `TerminalEncoder::writePrivateMode (output, mode, finalByte)` | `jam_terminal/buffers/jam_TerminalEncoder.h:48`; Part 3 Step 3, item 7 |
| `map::DEC::cursorVisible` (25), `alternateScreen` (1049), `bracketedPaste` (2004), `syncOutput` (2026); `map::CSI::setMode`, `resetMode` | `jam/cast/terminal.md:581,592-594,733-734` |
| `CpuGraphicsEngine::getRefreshRateHz()`: static; the primary display's refresh rate, or `indeterminateRefreshRateHz` (60.0) | `jam_graphics/engine/jam_CpuGraphicsEngine.h:67-71`; doxygen `classjam_1_1CpuGraphicsEngine` |
| EVE sizes its vblank work with `CpuGraphicsEngine::getRefreshRateHz()` | `eve/Source/EVETerminalProcessor.cpp:521,549` |
| `juce::Time::getHighResolutionTicks()`, `juce::Time::highResolutionTicksToSeconds (int64)` | JUCE `juce_core/time/juce_Time.h` |

### The contract with the owner (Part 7 and Part 10 build the owner)

The owner does these steps, in this order. Each step uses an API that exists today or that this part adds.

1. Construct `TerminalWindow (std::move (mainComponent), name)`.
2. Call `setLookAndFeel` on the window with a LookAndFeel that resolves the colour ids of the cells, as `printHelp` passes `styleMarkdown` (`ling/Source/Help.cpp:12,32`).
3. Call `bindToEvents (callbacks)`. The `Id::write` entry of `callbacks` receives the bytes of each frame on the message thread. The entry writes them to the terminal with `TerminalDisplay::write` (Part 3 Step 5), in the same call. This is the frame path of `printHelp`: paint, the terminal graphics context, the VT lane, stdout (`ling/Source/Help.cpp:26-34`). The entry is registered in the form of EVE's chain event (`EVETerminalProcessor.cpp:278-281`).
4. Call `setSize` with the columns and rows of `TerminalDisplay::getWinsize()`. `TerminalBuffer::prepare` asserts a height above 0.
5. Call `addToDesktop (0)` and `setVisible (true)`.
6. On the message thread, deliver input through `getPeer()->handleKeyPress (key)` and `getPeer()->findCurrentTextInputTarget()->insertTextAtCaret (text)`, and a size change through `setSize`.
7. Destroy the window before the owner of the `Id::write` entry and before `TerminalDisplay::close()`. The peer's destructor sends the bytes that leave the alternate screen through the same entry (Step 2), so the display must still be open.

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: The peer header
**Scope:** `jam_gui/windows/jam_TerminalComponentPeer.h` (new).
**Action:** Declare the class in namespace `jam`:

```cpp
class TerminalComponentPeer final : public juce::ComponentPeer
{
public:
    TerminalComponentPeer (juce::Component& component, int styleFlags, std::function<void (const char*, int)> newTerminalInput);
    ~TerminalComponentPeer() override;

    void* getNativeHandle() const override;
    void setVisible (bool shouldBeVisible) override;
    void setTitle (const juce::String& title) override;
    void setBounds (const juce::Rectangle<int>& newBounds, bool isNowFullScreen) override;
    juce::Rectangle<int> getBounds() const override;
    juce::Point<float> localToGlobal (juce::Point<float> relativePosition) override;
    juce::Point<float> globalToLocal (juce::Point<float> screenPosition) override;
    using juce::ComponentPeer::localToGlobal;
    using juce::ComponentPeer::globalToLocal;
    void setMinimised (bool shouldBeMinimised) override;
    bool isMinimised() const override;
    bool isShowing() const override;
    void setFullScreen (bool shouldBeFullScreen) override;
    bool isFullScreen() const override;
    void setIcon (const juce::Image& newIcon) override;
    bool contains (juce::Point<int> localPos, bool trueIfInAChildWindow) const override;
    OptionalBorderSize getFrameSizeIfPresent() const override;
    juce::BorderSize<int> getFrameSize() const override;
    bool setAlwaysOnTop (bool alwaysOnTop) override;
    void toFront (bool makeActive) override;
    void toBehind (juce::ComponentPeer* other) override;
    bool isFocused() const override;
    void grabFocus() override;
    void repaint (const juce::Rectangle<int>& area) override;
    void performAnyPendingRepaintsNow() override;
    void setAlpha (float newAlpha) override;
    juce::StringArray getAvailableRenderingEngines() override;

private:
    void textInputRequired (juce::Point<int> position, juce::TextInputTarget& target) override;
    void onVBlank();

    jam::TerminalBuffer terminalBuffer { true };
    juce::MemoryOutputStream frame;
    juce::RectangleList<int> regionsNeedingRepaint;
    std::function<void (const char*, int)> terminalInput;
    juce::TimedCallback vBlankManager { [this] { onVBlank(); } };

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TerminalComponentPeer)
};
```

The parameter names of the overrides are the names of `juce_ComponentPeer.h`. `vBlankManager` is the last member, so it stops before the other members go.

`terminalBuffer` turns the inline image lane on (`explicit TerminalBuffer (bool newShouldUseInlineImages = false)`, Part 3 Step 2). The inline lane is the only body of `drawImage` (ruling 28), so with the lane off, an image in the window draws nothing. TIT renders its pixels itself (ruling 24), and Part 3 names this peer the first user of the output lane (`PLAN-tit-3-terminal-display.md:16`).

**Validation:**
- The class overrides each of the 26 pure virtuals of the inventory, and no other base member.
- No member holds a pointer or a reference to another object. `terminalInput` is a copy of the window's callable.
- New names and their families: `TerminalComponentPeer` (ruling 14); `terminalBuffer` (`EVETerminalProcessor.h:363-364`); `frame` (`ling/Source/Help.cpp:8`); `regionsNeedingRepaint`, `vBlankManager`, `onVBlank` (`juce_Windowing_linux.cpp:561,661,636`); `terminalInput`, `newTerminalInput` (`jam_TextEditor.h:1286`; CODING.md, a distinct constructor parameter name).

### Step 2: The peer: construction, modes, frame, vblank
**Scope:** `jam_gui/windows/jam_TerminalComponentPeer.cpp` (new).
**Action:**
1. The constructor. It enters the full-screen modes and starts the frame clock:

```cpp
TerminalComponentPeer::TerminalComponentPeer (juce::Component& component, int styleFlags, std::function<void (const char*, int)> newTerminalInput)
    : juce::ComponentPeer (component, styleFlags),
      terminalInput (std::move (newTerminalInput))
{
    TerminalEncoder::writePrivateMode (frame, map::DEC::alternateScreen, map::CSI::setMode);
    TerminalEncoder::writePrivateMode (frame, map::DEC::cursorVisible, map::CSI::resetMode);
    TerminalEncoder::writePrivateMode (frame, map::DEC::bracketedPaste, map::CSI::setMode);
    terminalInput (static_cast<const char*> (frame.getData()), static_cast<int> (frame.getDataSize()));
    vBlankManager.startTimerHz (juce::roundToInt (CpuGraphicsEngine::getRefreshRateHz()));
}
```

2. The destructor. It leaves the modes in the reverse order:

```cpp
TerminalComponentPeer::~TerminalComponentPeer()
{
    frame.reset();
    TerminalEncoder::writePrivateMode (frame, map::DEC::bracketedPaste, map::CSI::resetMode);
    TerminalEncoder::writePrivateMode (frame, map::DEC::cursorVisible, map::CSI::setMode);
    TerminalEncoder::writePrivateMode (frame, map::DEC::alternateScreen, map::CSI::resetMode);
    terminalInput (static_cast<const char*> (frame.getData()), static_cast<int> (frame.getDataSize()));
}
```

3. `repaint (area)`: the Linux body (`juce_Windowing_linux.cpp:331-335,490-493`) with no scale factor: `regionsNeedingRepaint.add (area.getIntersection (getBounds().withZeroOrigin()));`.
4. `performAnyPendingRepaintsNow()`: the Linux order (`juce_Windowing_linux.cpp:495-547`), with the context and the frame of `printHelp` (`ling/Source/Help.cpp:23-32`):

```cpp
void TerminalComponentPeer::performAnyPendingRepaintsNow()
{
    if (not regionsNeedingRepaint.isEmpty())
    {
        regionsNeedingRepaint.clear();

        {
            TerminalGraphicsContext context { terminalBuffer, *Colours::getInstance() };
            handlePaint (context);
        }

        terminalBuffer.process();
        frame.reset();
        TerminalEncoder::writePrivateMode (frame, map::DEC::syncOutput, map::CSI::setMode);
        TerminalEncoder::writeEraseScreen (frame);
        terminalBuffer.writeToStream (frame, getComponent().getLookAndFeel());
        TerminalEncoder::writePrivateMode (frame, map::DEC::syncOutput, map::CSI::resetMode);

        {
            const auto history { terminalBuffer.read() };
        }

        terminalInput (static_cast<const char*> (frame.getData()), static_cast<int> (frame.getDataSize()));
    }
}
```

   - Each frame takes `heightRows` slots of the ring (`jam_TerminalGraphicsContext.cpp:12-13`). The scoped read frees the slots of the frame before, so the ring never fills.
   - The frame writes every row, inside synchronized output. This is the frame gate as it is (`jam_TerminalBuffer.h:94-104`; tracker hole "Damage").
   - The last row ends without a newline (ruling 31), so the host does not scroll.
5. `onVBlank()`: the Linux body (`juce_Windowing_linux.cpp:636-643`). The timestamp comes from the JUCE seconds API, so no constant is needed:

```cpp
void TerminalComponentPeer::onVBlank()
{
    callVBlankListeners (juce::Time::highResolutionTicksToSeconds (juce::Time::getHighResolutionTicks()));
    performAnyPendingRepaintsNow();
}
```

**Validation:**
- A frame has this byte order: sync set, erase screen, the rows with a newline between two rows, sync reset.
- The peer writes no frame when no repaint is pending.
- A `jam::TextEditor` with a `juce::VBlankAttachment` gets its callback, because the peer calls `callVBlankListeners` (tracker hole "Live peer"; `jam_TextEditor.cpp:98`).
- No magic number: the rate is `CpuGraphicsEngine::getRefreshRateHz()`, as in EVE (`EVETerminalProcessor.cpp:521,549`).
- New name and its family: `history` (the read handle of "the history rows", `jam_TerminalBuffer.h:277-284`).

### Step 3: The peer: bounds, focus, and the members with no capability
**Scope:** `jam_gui/windows/jam_TerminalComponentPeer.cpp`.
**Action:**
1. `getBounds()`: `return getComponent().getBounds();`. `Component::setBounds` writes the component bounds before it calls `updateBounds` (`juce_Component.cpp:1111,1128-1130`). So the peer holds no copy of the size (MANIFESTO S, SSOT), and `handlePaint` adds no scale (`juce_ComponentPeer.cpp:122-131`).
2. `setBounds (newBounds, isNowFullScreen)`: `terminalBuffer.prepare (newBounds.getWidth(), newBounds.getHeight(), newBounds.getHeight());`, the arguments of `printHelp` (`ling/Source/Help.cpp:24`). Then `handleMovedOrResized()`, as the base asks: "This should result in a callback to handleMovedOrResized()" (`juce_ComponentPeer.h:204`). The bounds are equal, so it changes nothing (`juce_ComponentPeer.cpp:329-335`). `isNowFullScreen` stays unnamed in the definition.
3. `localToGlobal (relativePosition)`: `return relativePosition + getBounds().getPosition().toFloat();`. `globalToLocal (screenPosition)`: `return screenPosition - getBounds().getPosition().toFloat();`.
4. `contains (localPos, trueIfInAChildWindow)`: `return getBounds().withZeroOrigin().contains (localPos);`. The peer has no child window. `trueIfInAChildWindow` stays unnamed in the definition.
5. `toFront (makeActive)`: the Linux body (`juce_Windowing_linux.cpp:291-301`) without the X11 call: when `makeActive`, `setVisible (true)` and `grabFocus()`; then `handleBroughtToFront()`.
6. `grabFocus()`: `handleFocusGain();` (`juce_ComponentPeer.cpp:362-379`).
7. `isFocused()`: `return true;`. The host terminal sends every key to this program, so the window always has the keyboard.
8. `isShowing()`: `return true;`. `isMinimised()`: `return false;`. `isFullScreen()`: `return false;`. A terminal window has no minimised state and no full-screen state.
9. `getFrameSizeIfPresent()`: `return OptionalBorderSize { juce::BorderSize<int> {} };`. The window has no frame. `getFrameSize()`: the Linux body (`juce_Windowing_linux.cpp:170-174`).
10. `getNativeHandle()`: `return nullptr;`. No OS window exists. The accessibility code returns early with no native client (tracker, "The JUCE attach path", row "Accessibility").
11. `getAvailableRenderingEngines()`: `return {};`. The one renderer is `TerminalGraphicsContext`, and the window offers no choice.
12. `setAlwaysOnTop`: `return false;`, the Linux body (`juce_Windowing_linux.cpp:371`). The empty bodies, in the form of the Linux `setAlpha` and `textInputRequired` (`:370,372`): `setVisible`, `setTitle`, `setMinimised`, `setFullScreen`, `setIcon`, `toBehind`, `setAlpha`, `textInputRequired`. Each parameter stays unnamed in the definition. No SPEC line and no ruling names a window title, an icon, or transparency (MANIFESTO L, YAGNI).

**Validation:**
- `getBounds()` reads the component. The peer holds no field for the size.
- A `setSize` from the owner gives `prepare` with the new size, then one full frame on the next vblank: `Component::setBounds` repaints a resized component (`juce_Component.cpp:1113-1116`).
- Every empty body has a Linux peer sibling or names a state the window does not have.

### Step 4: `TerminalWindow`
**Scope:** `jam_gui/windows/jam_TerminalWindow.h`, `jam_gui/windows/jam_TerminalWindow.cpp` (both new).
**Action:**
1. Declare the class in namespace `jam`:

```cpp
class TerminalWindow final : public juce::Component
{
public:
    TerminalWindow (std::unique_ptr<juce::Component> mainComponent, const juce::String& name);

    void bindToEvents (Registry::Events::mapped_type& callbacks);
    void resized() override;

private:
    juce::ComponentPeer* createNewPeer (int styleFlags, void* nativeWindowToAttachTo) override;

    std::unique_ptr<juce::Component> contentComponent;
    std::function<void (const char*, int)> terminalInput;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TerminalWindow)
};
```

2. The constructor: `: juce::Component (name), contentComponent (std::move (mainComponent))`, then `addAndMakeVisible (*contentComponent);`. The signature is the one of `jam::Window` without the native window options (`jam_Window.h:76`).
3. `resized()`: `contentComponent->setBounds (getLocalBounds());`.
4. `bindToEvents (callbacks)`: the body of `TextEditor::bindToEvents` (`jam_TextEditor.cpp:104-107`), verbatim.
5. `createNewPeer (styleFlags, nativeWindowToAttachTo)`:

```cpp
juce::ComponentPeer* TerminalWindow::createNewPeer (int styleFlags, void*)
{
    jassert (terminalInput != nullptr);

    return std::make_unique<TerminalComponentPeer> (*this, styleFlags, terminalInput).release();
}
```

   - The JUCE contract returns an owning raw pointer, and `removeFromDesktop` deletes it (`juce_Component.cpp:758-762`). `make_unique` and `release` hand the ownership over with no raw `new`, in the form of `setContentOwned (mainComponent.release(), true)` (`jam_Window.cpp:44`).
   - The assert is the precondition of the owner contract: `bindToEvents` runs before `addToDesktop`.

**Validation:**
- `TerminalWindow` is a plain `juce::Component`. `addToDesktop` calls its `createNewPeer` (`juce_Component.cpp:691`), and no native window exists.
- No member is a non-owning pointer. The window owns its content and the peer, through the JUCE desktop contract.
- New names and their families: `TerminalWindow` (ruling 13); `mainComponent` (`jam_Window.h:76`); `contentComponent` (`juce_ResizableWindow.h:403`); `bindToEvents`, `terminalInput` (`jam_TextEditor.cpp:102-108`; `jam_TextEditor.h:1286`).

### Step 5: Wire the module
**Scope:** `jam_gui/jam_gui.h`, `jam_gui/jam_gui.cpp`.
**Action:**
1. In `jam_gui.h`, after `#include "windows/jam_Window.h"` (`:124`), add `#include "windows/jam_TerminalComponentPeer.h"`, then `#include "windows/jam_TerminalWindow.h"`. The registry header comes before (`:97`), so `Registry::Events` is known. `jam_terminal.h` comes before (`:83`), so `TerminalBuffer`, `TerminalGraphicsContext`, `TerminalEncoder`, and `map::` are known.
2. In `jam_gui.cpp`, after `#include "windows/jam_Window.cpp"` (`:21`), add `#include "windows/jam_TerminalComponentPeer.cpp"`, then `#include "windows/jam_TerminalWindow.cpp"`.

**Validation:** Each new header and source is included once. The four new files hold no include.

### Step 6: Build and check the consumers
**Scope:** JAM through its consumers on macOS: `ling`, `cast`, `scout`. Each one links `jam_gui` (`ling/project-info.md:131`; `cast/project-info.md:134`; `scout/project-info.md:130`).
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:**
1. Build each consumer with `cast cast/spell.md --debug`.
2. Run `ling --help`, `cast --help`, and `scout --help`. Compare each output with the output saved at Part 3 Step 9.
3. Build `ling` on a Windows machine.

**Validation:**
- Each build ends with exit code 0.
- Each help output is byte-identical to its saved output. No consumer constructs a `TerminalWindow`.
- The runtime facts of this part are in the tracker list "Facts that execution measures first". Part 7 runs the window for the first time.

## BLESSED Alignment

- **B — Bound.** `TerminalWindow` owns its content. JUCE gives the window the ownership of its peer, and `removeFromDesktop` deletes it (`juce_Component.cpp:758-762`). The peer owns its buffer, its frame stream, its repaint region, and its timer.
- **L — Lean.** The peer holds only the 26 overrides of the JUCE contract, the frame, and the vblank tick. A capability that the window does not have gets the Linux form of an empty body.
- **E — Explicit.** The modes come from `map::DEC` and `map::CSI`. The frame rate comes from `CpuGraphicsEngine::getRefreshRateHz()`. The owner contract is a list of ordered steps.
- **S — SSOT.** The component holds the size; `getBounds()` reads it. The event map of the owner is the one path of the frame bytes, and it runs on the message thread.
- **S — Stateless.** The peer keeps only transient frame data and the repaint region of the Linux peer. No field tracks the state of another object.
- **E — Encapsulation.** Input and size reach the peer through JUCE API only. The peer knows no TIT type, and `jam_terminal` still names no `jam_gui` type.
- **D — Deterministic.** The same component tree and the same size give the same frame bytes.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. Three runtime facts have no run yet: the attach path with a peer that is not a native view, the record shape of Windows console input, and a live editable `jam::TextEditor` on a terminal. They are in the tracker list "Facts that execution measures first". Part 7 measures the first two (Steps 8 and 9). Part 10 measures the third, with the commit box.
