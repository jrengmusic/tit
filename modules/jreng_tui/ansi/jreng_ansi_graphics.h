#pragma once

#include <string>

namespace jreng::tui
{ /*____________________________________________________________________________*/

// ============================================================================
// Graphics
// ============================================================================

/** Framebuffer writer passed to Component::paint().
 *
 *  Accumulates cell data into a flat juce::HeapBlock<Cell> grid — one Cell
 *  per terminal column/row position.  getLines() serializes the cell grid to
 *  a juce::StringArray (one ANSI-escaped string per row) for Screen diffing.
 *
 *  Mirrors the juce::Graphics API surface so Component implementations
 *  feel familiar.  Text layout walks AttributedString attributes directly —
 *  cell-native, no pixel measurement.
 *
 *  Clip contract:
 *    clip() returns a child Graphics that writes into the SAME cell grid as
 *    the parent, offset by the clip rect origin and clamped to the clip rect
 *    bounds.  No copy of the framebuffer is made.
 *
 *  Paint contract:
 *    Constructed per frame by Screen.  Must not be cached across frames.
 *    All methods are message-thread only.
 */
class Graphics
{
public:
    Graphics (int widthCols, int heightRows);

    void setColour (juce::Colour c);
    void setFont   (const juce::Font& f);

    void fillRect             (Rectangle bounds);
    void drawText             (const juce::String& text,
                               Rectangle bounds,
                               juce::Justification justification,
                               bool useEllipsis = true);
    void drawAttributedString (const juce::AttributedString& str,
                               Rectangle bounds);
    void drawHorizontalLine   (int row, juce::Colour c);
    void drawVerticalLine     (int col, juce::Colour c);

    /** Returns a child Graphics writing into the same framebuffer,
     *  offset and clamped to bounds.  jassert in debug that bounds
     *  is fully inside this graphics context's cell area.
     */
    Graphics clip (Rectangle bounds) const;

    /** Serializes the cell grid to one ANSI-escaped juce::String per row.
     *  Called by Screen only.  Return value is valid until the next non-const
     *  call on this Graphics.
     */
    const juce::StringArray& getLines() const;

    /** Records the display-column cursor position for the focused component.
     *  Injected as ANSI::CURSOR_MARKER into the serialized row by getLines().
     *  col and row are cell coordinates relative to this graphics context.
     */
    void emitCursorMarker (int col, int row);

    /** Writes text directly into cells at the given cell coordinates.
     *  Cell-native method — no TextLayout, no font measurement.
     *  Use for terminal-native rendering where cell == character.
     *  startCol and startRow are relative to this graphics context.
     */
    void drawCellText (const juce::String& text, int startCol, int startRow, int maxCols);

private:
    // -----------------------------------------------------------------------
    // Root constructor storage
    // -----------------------------------------------------------------------
    juce::HeapBlock<Cell> cells;       // owns the flat cell grid (root only)

    // -----------------------------------------------------------------------
    // Shared members (root and clip children)
    // -----------------------------------------------------------------------
    int            widthCols       { 0 };
    int            heightRows      { 0 };
    Pen            pen;            // current drawing state
    juce::Font     currentFont { juce::FontOptions{} };   // retained for AttributedString construction in drawText

    // -----------------------------------------------------------------------
    // Clip / stride support
    // -----------------------------------------------------------------------
    Cell*          cellsPtr        { nullptr }; // nullptr = owns cells; else borrows parent's
    int            offsetCol       { 0 };
    int            offsetRow       { 0 };
    int            stride          { 0 };       // row stride of the root grid

    // -----------------------------------------------------------------------
    // Cursor side-channel (mutable: written by clip children via cursorTarget)
    // -----------------------------------------------------------------------
    mutable int    cursorMarkerCol { -1 };
    mutable int    cursorMarkerRow { -1 };

    // Observer pointer to root Graphics where cursor is written back.
    // nullptr for root; points to root for clip children.
    Graphics*      cursorTarget    { nullptr };

    // -----------------------------------------------------------------------
    // Serialization cache
    // -----------------------------------------------------------------------
    mutable juce::StringArray serializedLines;

    // -----------------------------------------------------------------------
    // Clip constructor (borrows parent cell grid, propagates cursor to root)
    // -----------------------------------------------------------------------
    Graphics (Cell*     parentCells,
              int       parentStride,
              int       colOffset,
              int       rowOffset,
              int       clipWidthCols,
              int       clipHeightRows,
              Graphics* rootTarget);

    // -----------------------------------------------------------------------
    // Cell grid access
    // -----------------------------------------------------------------------
    Cell& cellAt (int col, int row) noexcept;

    // -----------------------------------------------------------------------
    // Text run helpers
    // -----------------------------------------------------------------------
    void writeTextRunIntoCells (const juce::String& runText,
                                const Color&        fg,
                                uint8_t             styleBits,
                                int                 baselineRow,
                                int                 startCol,
                                int                 maxCols);

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Graphics)
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
