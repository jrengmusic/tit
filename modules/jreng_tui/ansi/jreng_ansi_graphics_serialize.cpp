// ============================================================================
// Graphics serialization — getLines() and UTF-8/SGR helpers
// Unity-build included by jreng_tui.cpp
// ============================================================================

namespace jreng::tui
{ /*____________________________________________________________________________*/

// ============================================================================
// UTF-8 encoding helper
// ============================================================================

static void appendCodepointAsUTF8 (std::string& out, uint32_t cp) noexcept
{
    if (cp < 0x80u)
    {
        out += static_cast<char> (cp);
    }
    else if (cp < 0x800u)
    {
        out += static_cast<char> (0xC0u | (cp >> 6));
        out += static_cast<char> (0x80u | (cp & 0x3Fu));
    }
    else if (cp < 0x10000u)
    {
        out += static_cast<char> (0xE0u | (cp >> 12));
        out += static_cast<char> (0x80u | ((cp >> 6) & 0x3Fu));
        out += static_cast<char> (0x80u | (cp & 0x3Fu));
    }
    else
    {
        out += static_cast<char> (0xF0u | (cp >> 18));
        out += static_cast<char> (0x80u | ((cp >> 12) & 0x3Fu));
        out += static_cast<char> (0x80u | ((cp >> 6)  & 0x3Fu));
        out += static_cast<char> (0x80u | (cp & 0x3Fu));
    }
}

// ============================================================================
// ANSI SGR escape builders
// ============================================================================

static std::string fgEscape (const Color& c)
{
    std::string esc { ANSI::CSI_PREFIX };
    esc += "38;2;";
    esc += std::to_string (c.red);
    esc += ";";
    esc += std::to_string (c.green);
    esc += ";";
    esc += std::to_string (c.blue);
    esc += "m";
    return esc;
}

static std::string bgEscape (const Color& c)
{
    std::string esc { ANSI::CSI_PREFIX };
    esc += "48;2;";
    esc += std::to_string (c.red);
    esc += ";";
    esc += std::to_string (c.green);
    esc += ";";
    esc += std::to_string (c.blue);
    esc += "m";
    return esc;
}

// ============================================================================
// Color equality check
// ============================================================================

static bool colorsDiffer (const Color& a, const Color& b) noexcept
{
    return a.red != b.red or a.green != b.green or a.blue != b.blue or a.mode != b.mode;
}

// ============================================================================
// Style SGR diff — emits only changed style bits
// ============================================================================

static void appendStyleDiff (std::string& out, uint8_t oldStyle, uint8_t newStyle)
{
    if ((oldStyle & Cell::BOLD) != (newStyle & Cell::BOLD))
        out += (newStyle & Cell::BOLD) ? ANSI::BOLD_ON : ANSI::BOLD_OFF;

    if ((oldStyle & Cell::ITALIC) != (newStyle & Cell::ITALIC))
        out += (newStyle & Cell::ITALIC) ? ANSI::ITALIC_ON : ANSI::ITALIC_OFF;

    if ((oldStyle & Cell::UNDERLINE) != (newStyle & Cell::UNDERLINE))
        out += (newStyle & Cell::UNDERLINE) ? ANSI::UNDERLINE_ON : ANSI::UNDERLINE_OFF;
}

// ============================================================================
// Emit SGR changes for one cell relative to running render state
// ============================================================================

static void appendCellSGR (std::string&   out,
                            const Cell&    cell,
                            Color&         currentFg,
                            Color&         currentBg,
                            uint8_t&       currentStyle,
                            bool           firstCell)
{
    if (firstCell or colorsDiffer (cell.fg, currentFg))
    {
        out += fgEscape (cell.fg);
        currentFg = cell.fg;
    }

    if (firstCell or colorsDiffer (cell.bg, currentBg))
    {
        out += bgEscape (cell.bg);
        currentBg = cell.bg;
    }

    if (firstCell or cell.style != currentStyle)
    {
        appendStyleDiff (out, currentStyle, cell.style);
        currentStyle = cell.style;
    }
}

// ============================================================================
// Emit one cell's character (or cursor marker + character)
// ============================================================================

static void appendCellCharacter (std::string& out, const Cell& cell,
                                  int col, int cursorCol, bool hasCursor)
{
    if (hasCursor and col == cursorCol)
        out += ANSI::CURSOR_MARKER;

    if (cell.codepoint == 0)
        out += ' ';
    else
        appendCodepointAsUTF8 (out, cell.codepoint);
}

// ============================================================================
// Row serializer — one row of cells → std::string of ANSI + UTF-8
// ============================================================================

static std::string serializeRow (const Cell* rowBase, int cols,
                                  int cursorCol, bool hasCursor)
{
    std::string out;
    out.reserve (static_cast<std::size_t> (cols) * 8);

    Color    currentFg;
    Color    currentBg;
    uint8_t  currentStyle { 0 };
    bool     firstCell    { true };

    for (int col { 0 }; col < cols; ++col)
    {
        const Cell& cell { rowBase[col] };

        if (not cell.isWideContinuation())
        {
            appendCellSGR (out, cell, currentFg, currentBg, currentStyle, firstCell);
            appendCellCharacter (out, cell, col, cursorCol, hasCursor);
            firstCell = false;
        }
    }

    out += ANSI::RESET;
    return out;
}

// ============================================================================
// getLines() — serialize full cell grid to StringArray
// ============================================================================

const juce::StringArray& Graphics::getLines() const
{
    serializedLines.clearQuick();
    serializedLines.ensureStorageAllocated (heightRows);

    const Cell* const rootCells { cellsPtr != nullptr ? cellsPtr : cells.getData() };

    for (int row { 0 }; row < heightRows; ++row)
    {
        const int absRow      { row + offsetRow };
        const Cell* rowBase   { rootCells + absRow * stride + offsetCol };

        const bool hasCursor  { cursorMarkerRow == absRow
                                and cursorMarkerCol >= offsetCol
                                and cursorMarkerCol < offsetCol + widthCols };
        const int  cursorCol  { hasCursor ? cursorMarkerCol - offsetCol : -1 };

        const std::string rowBytes { serializeRow (rowBase, widthCols, cursorCol, hasCursor) };

        serializedLines.add (juce::String (juce::CharPointer_UTF8 (rowBytes.c_str()),
                                            rowBytes.size()));
    }

    return serializedLines;
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
