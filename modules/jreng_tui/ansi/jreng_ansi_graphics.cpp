// ============================================================================
// Graphics implementation
// ============================================================================

namespace jreng::tui
{ /*____________________________________________________________________________*/

static_assert (sizeof (Cell) == 16, "Cell must be 16 bytes");

// ============================================================================
// Construction
// ============================================================================

Graphics::Graphics (int widthCols_, int heightRows_)
    : widthCols  { widthCols_ }
    , heightRows { heightRows_ }
    , stride     { widthCols_ }
{
    cells.allocate (widthCols_ * heightRows_, true);
    cellsPtr = nullptr;
}

Graphics::Graphics (Cell*     parentCells,
                    int       parentStride,
                    int       colOffset,
                    int       rowOffset,
                    int       clipWidthCols,
                    int       clipHeightRows,
                    Graphics* rootTarget)
    : widthCols    { clipWidthCols }
    , heightRows   { clipHeightRows }
    , cellsPtr     { parentCells }
    , offsetCol    { colOffset }
    , offsetRow    { rowOffset }
    , stride       { parentStride }
    , cursorTarget { rootTarget }
{
}

// ============================================================================
// State setters
// ============================================================================

void Graphics::setColour (juce::Colour c)
{
    pen.fg.setRGB (c.getRed(), c.getGreen(), c.getBlue());
}

void Graphics::setFont (const juce::Font& f)
{
    currentFont = f;
    pen.style = 0;

    if (f.isBold())
        pen.style |= Cell::BOLD;

    if (f.isItalic())
        pen.style |= Cell::ITALIC;
}

// ============================================================================
// Cell access
// ============================================================================

Cell& Graphics::cellAt (int col, int row) noexcept
{
    jassert (col >= 0 and col < widthCols and row >= 0 and row < heightRows);
    const int absCol { col + offsetCol };
    const int absRow { row + offsetRow };

    if (cellsPtr != nullptr)
        return cellsPtr[absRow * stride + absCol];

    return cells[row * stride + col];
}

// ============================================================================
// Clip
// ============================================================================

Graphics Graphics::clip (Rectangle bounds) const
{
    const Rectangle parentBounds { 0, 0, widthCols, heightRows };
    jassert (parentBounds.contains (bounds));

    Cell* const parentCells { cellsPtr != nullptr ? cellsPtr : cells.getData() };
    Graphics*   root        { cursorTarget != nullptr ? cursorTarget : const_cast<Graphics*> (this) };

    return Graphics (parentCells,
                     stride,
                     offsetCol + bounds.getX(),
                     offsetRow + bounds.getY(),
                     bounds.getWidth(),
                     bounds.getHeight(),
                     root);
}

// ============================================================================
// Drawing
// ============================================================================

void Graphics::fillRect (Rectangle bounds)
{
    for (int row { bounds.getY() }; row < bounds.getBottom() and row < heightRows; ++row)
    {
        for (int col { bounds.getX() }; col < bounds.getRight() and col < widthCols; ++col)
        {
            Cell& c { cellAt (col, row) };
            c.codepoint = static_cast<uint32_t> (' ');
            c.fg        = pen.fg;
            c.bg        = pen.bg;
            c.style     = pen.style;
            c.width     = 1;
            c.layout    = 0;
        }
    }
}

void Graphics::drawText (const juce::String& text,
                          Rectangle           bounds,
                          juce::Justification justification,
                          bool                useEllipsis)
{
    juce::ignoreUnused (useEllipsis);

    juce::AttributedString attrStr;
    attrStr.setJustification (justification);
    attrStr.append (text, currentFont, juce::Colour (pen.fg.red, pen.fg.green, pen.fg.blue));

    drawAttributedString (attrStr, bounds);
}

void Graphics::writeTextRunIntoCells (const juce::String& runText,
                                       const Color&        fgColour,
                                       uint8_t             styleBits,
                                       int                 baselineRow,
                                       int                 startCol,
                                       int                 maxCols)
{
    if (baselineRow >= 0 and baselineRow < heightRows)
    {
        int col { startCol };
        auto iter { runText.begin() };

        while (iter != runText.end() and col < maxCols and col < widthCols)
        {
            const juce::juce_wchar cp { *iter };
            Cell& c { cellAt (col, baselineRow) };
            c.codepoint = static_cast<uint32_t> (cp);
            c.fg        = fgColour;
            c.bg        = pen.bg;
            c.style     = styleBits;
            c.width     = 1;
            c.layout    = 0;
            ++col;
            ++iter;
        }
    }
}

void Graphics::drawAttributedString (const juce::AttributedString& str,
                                      Rectangle                     bounds)
{
    const int maxCols  { bounds.getWidth() };
    const int startRow { bounds.getY() };

    int col { bounds.getX() };
    int row { startRow };

    for (int attrIdx { 0 }; attrIdx < str.getNumAttributes(); ++attrIdx)
    {
        const juce::AttributedString::Attribute& attr { str.getAttribute (attrIdx) };

        const juce::String runText { str.getText().substring (
            attr.range.getStart(), attr.range.getEnd()) };

        Color runFg;
        runFg.setRGB (attr.colour.getRed(),
                      attr.colour.getGreen(),
                      attr.colour.getBlue());

        uint8_t runStyle { 0 };

        if (attr.font.isBold())   runStyle |= Cell::BOLD;
        if (attr.font.isItalic()) runStyle |= Cell::ITALIC;

        auto iter { runText.begin() };

        while (iter != runText.end())
        {
            const juce::juce_wchar cp { *iter };

            if (cp == '\n')
            {
                ++row;
                col = bounds.getX();
            }
            else
            {
                const int relCol { col - bounds.getX() };

                if (relCol < maxCols and row < bounds.getBottom() and row >= startRow)
                {
                    Cell& c { cellAt (col, row) };
                    c.codepoint = static_cast<uint32_t> (cp);
                    c.fg        = runFg;
                    c.bg        = pen.bg;
                    c.style     = runStyle;
                    c.width     = 1;
                    c.layout    = 0;
                }

                ++col;

                if (col - bounds.getX() >= maxCols)
                {
                    ++row;
                    col = bounds.getX();
                }
            }

            ++iter;
        }
    }
}

void Graphics::drawHorizontalLine (int row, juce::Colour c)
{
    if (row >= 0 and row < heightRows)
    {
        Color lineColor;
        lineColor.setRGB (c.getRed(), c.getGreen(), c.getBlue());

        for (int col { 0 }; col < widthCols; ++col)
        {
            Cell& cell { cellAt (col, row) };
            cell.codepoint = 0x2500u;
            cell.fg        = lineColor;
            cell.bg        = pen.bg;
            cell.style     = 0;
            cell.width     = 1;
            cell.layout    = 0;
        }
    }
}

void Graphics::drawVerticalLine (int col, juce::Colour c)
{
    if (col >= 0 and col < widthCols)
    {
        Color lineColor;
        lineColor.setRGB (c.getRed(), c.getGreen(), c.getBlue());

        for (int row { 0 }; row < heightRows; ++row)
        {
            Cell& cell { cellAt (col, row) };
            cell.codepoint = 0x2502u;
            cell.fg        = lineColor;
            cell.bg        = pen.bg;
            cell.style     = 0;
            cell.width     = 1;
            cell.layout    = 0;
        }
    }
}

void Graphics::drawCellText (const juce::String& text, int startCol, int startRow, int maxCols)
{
    if (startRow >= 0 and startRow < heightRows)
    {
        int col { startCol };
        auto iter { text.begin() };

        while (iter != text.end() and col < startCol + maxCols and col < widthCols)
        {
            Cell& c { cellAt (col, startRow) };
            c.codepoint = static_cast<uint32_t> (*iter);
            c.fg        = pen.fg;
            c.bg        = pen.bg;
            c.style     = pen.style;
            c.width     = 1;
            c.layout    = 0;
            ++col;
            ++iter;
        }
    }
}

void Graphics::emitCursorMarker (int col, int row)
{
    const int absCol { col + offsetCol };
    const int absRow { row + offsetRow };

    if (cursorTarget != nullptr)
    {
        cursorTarget->cursorMarkerCol = absCol;
        cursorTarget->cursorMarkerRow = absRow;
    }
    else
    {
        cursorMarkerCol = absCol;
        cursorMarkerRow = absRow;
    }
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
