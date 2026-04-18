#pragma once

namespace jreng::tui
{ /*____________________________________________________________________________*/

// Forward declaration — Screen is referenced in OverlayHandle before the
// class definition below.
class Screen;

// ============================================================================
// Overlay
// ============================================================================

/** Registered overlay entry. component is an observer pointer — lifetime owned
 *  by the caller, not Screen. Rendered into the same Graphics framebuffer
 *  after the base tree, ordered by ascending focusOrder.
 */
struct Overlay
{
    Component* component  { nullptr };
    Rectangle  bounds     {};
    bool       hidden     { false };
    int        focusOrder { 0 };
};

// ============================================================================
// OverlayHandle
// ============================================================================

/** RAII handle returned by Screen::showOverlay(). Non-copyable, movable.
 *  Destructor calls screen->hideOverlay(id) when screen is non-null.
 *  Move transfers ownership; source screen is nulled to prevent double-hide.
 */
struct OverlayHandle
{
    int     id     { -1 };
    Screen* screen { nullptr };

    ~OverlayHandle();
    OverlayHandle() = default;
    OverlayHandle (int overlayId, Screen* ownerScreen);
    OverlayHandle (const OverlayHandle&)            = delete;
    OverlayHandle& operator= (const OverlayHandle&) = delete;
    OverlayHandle (OverlayHandle&& other) noexcept;
    OverlayHandle& operator= (OverlayHandle&& other) noexcept;
};

// ============================================================================
// Screen
// ============================================================================

/** Root TUI component. Owns the render loop and drives differential rendering
 *  to stdout via Writer.
 *
 *  Render flow:
 *    requestRender() → (debounced via renderPending) → doRender() on message
 *    thread → Graphics framebuffer built → three-strategy diff against
 *    previousFrame → Writer flush.
 *
 *  Three strategies:
 *    1. First render  — previousFrame empty, write all lines.
 *    2. Full redraw   — terminal resized or fullRedrawPending, clear + rewrite.
 *    3. Differential  — scan for changed rows, write only those.
 *
 *  Threading contract:
 *    All public methods are message-thread only, except renderPending which is
 *    atomic so any thread can observe it safely.
 *
 *  Overlay contract:
 *    Overlays render after the base tree into the same framebuffer, in ascending
 *    focusOrder. OverlayHandle RAII — destructor removes the overlay.
 */
class Screen : public Component
{
public:
    Screen (Writer& writer);

    /** Schedules the first render via requestRender(). */
    void start();

    /** Debounced render request. Posts doRender() via callAsync when not
     *  already pending. Safe to call multiple times per message loop tick.
     */
    void requestRender();

    /** Sets fullRedrawPending and calls requestRender(). */
    void requestFullRedraw();

    /** Calls getBounds() then requestFullRedraw(). Called from the
     *  resize callback posted by Input via MessageManager::callAsync.
     */
    void onTerminalResized();

    /** Component override. Called by doRender() to recurse into children.
     *  Do not call directly — entry point is doRender().
     */
    void paint  (Graphics& g) override;

    /** Registers an overlay. Returns a RAII handle — overlay removed on handle
     *  destruction. component is an observer pointer; caller owns its lifetime.
     */
    OverlayHandle showOverlay (Component* component,
                               Rectangle bounds);

    /** Removes the overlay with the given id. Calls requestRender(). */
    void hideOverlay (int id);

    /** Removes the overlay with the highest focusOrder. Calls requestRender(). */
    void hideTopOverlay();

    /** Returns true when at least one overlay is registered. */
    bool hasOverlay() const noexcept;

private:
    Writer&                   writer;
    juce::StringArray         previousFrame;
    juce::Array<Overlay>      overlays;
    int                       nextOverlayId      { 0 };
    bool                      firstRender        { true };
    bool                      fullRedrawPending   { false };
    std::atomic<bool>         renderPending       { false };

    /** Core render method. Builds framebuffer, selects strategy, writes frame. */
    void doRender();

    /** Strategy 1: write all lines without clearing (previousFrame empty). */
    void renderFirstFrame  (const juce::StringArray& newFrame);

    /** Strategy 2: clear screen then write all lines. */
    void renderFull        (const juce::StringArray& newFrame);

    /** Strategy 3: scan for changed rows, write only those. */
    void renderDifferential (const juce::StringArray& newFrame);

    /** Writes the changed row range [firstChanged, lastChanged] and clears any
     *  trailing rows when newFrame is shorter than previousFrame.
     */
    void writeDifferentialRange (const juce::StringArray& newFrame,
                                 int firstChanged,
                                 int lastChanged);

    /** Iterates JUCE child components and calls paint() on Component ones. */
    void renderChildren (Graphics& g);

    /** Renders all non-hidden overlays in ascending focusOrder into g. */
    void renderOverlays (Graphics& g);

    /** Positions the hardware cursor. No-op when pos is invalid (-1, -1). */
    void positionHardwareCursor (juce::Point<int> cursorPos);

    /** Scans frame for ANSI::CURSOR_MARKER, strips it in place, and returns
     *  the {col, row} position. Returns {-1, -1} when marker is absent.
     */
    juce::Point<int> extractCursorMarker (juce::StringArray& frame);

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Screen)
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
