#pragma once

namespace jreng::tui
{ /*____________________________________________________________________________*/

// ============================================================================
// Writer
// ============================================================================

/** Accumulates ANSI escape sequences into a string buffer and flushes
 *  atomically to stdout.
 *
 *  Every render cycle MUST be wrapped with beginFrame() / endFrame().
 *  endFrame() is the only place flush() is called — one stdout write per frame.
 */
class Writer
{
public:
    /** Emits ANSI::SYNC_START — call once at the start of every render cycle. */
    void beginFrame();

    /** Emits ANSI::SYNC_END then flushes the buffer to stdout exactly once. */
    void endFrame();

    void moveUp        (int rows);
    void moveDown      (int rows);
    void moveToCol     (int col);
    void moveToRow     (int row);
    void moveTo        (int row, int col);

    void enterAltScreen();
    void exitAltScreen();

    void clearScreen();
    void clearLine();
    void clearToEndOfLine();
    void hideCursor();
    void showCursor();

    /** Emits the line content followed by ANSI::LINE_SUFFIX. */
    void writeLine (const juce::String& ansiLine);

    /** Emits a true-colour foreground sequence: \x1b[38;2;r;g;bm */
    void setFg (juce::Colour c);

    /** Emits a true-colour background sequence: \x1b[48;2;r;g;bm */
    void setBg (juce::Colour c);

    void resetAttrs();

    void setBold      (bool b);
    void setItalic    (bool b);
    void setUnderline (bool b);

private:
    juce::String buffer;

    /** Appends seq to buffer. */
    void emit (const juce::String& seq);

    /** Writes buffer to stdout then clears it. Called once per endFrame(). */
    void flush();
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
