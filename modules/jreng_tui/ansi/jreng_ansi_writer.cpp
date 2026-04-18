// ============================================================================
// Writer implementation
// ============================================================================

#include <iostream>

namespace jreng::tui
{ /*____________________________________________________________________________*/

void Writer::emit (const juce::String& seq)
{
    buffer += seq;
}

void Writer::flush()
{
    // DIAGNOSTIC — remove after debugging
    {
        juce::File diagFile { juce::File::getSpecialLocation (juce::File::userDesktopDirectory).getChildFile ("caroline_diag.bin") };
        diagFile.replaceWithData (buffer.toRawUTF8(), static_cast<size_t> (buffer.getNumBytesAsUTF8()));
    }
    // DIAGNOSTIC — remove after debugging
    {
        juce::File diagText { juce::File::getSpecialLocation (juce::File::userDesktopDirectory).getChildFile ("caroline_diag.txt") };
        juce::String hex;
        const char* raw { buffer.toRawUTF8() };
        const int len { static_cast<int> (buffer.getNumBytesAsUTF8()) };
        for (int i { 0 }; i < juce::jmin (len, 2000); ++i)
        {
            const auto byte { static_cast<unsigned char> (raw[i]) };
            hex += juce::String::toHexString (byte).paddedLeft ('0', 2) + " ";
            if ((i + 1) % 32 == 0) hex += "\n";
        }
        diagText.replaceWithText (hex);
    }
    std::cout << buffer.toRawUTF8() << std::flush;
    buffer = juce::String();
}

void Writer::beginFrame()
{
    emit (ANSI::SYNC_START);
}

void Writer::endFrame()
{
    emit (ANSI::SYNC_END);
    flush();
}

void Writer::moveUp (int rows)
{
    emit (juce::String (ANSI::CSI_PREFIX) + juce::String (rows) + "A");
}

void Writer::moveDown (int rows)
{
    emit (juce::String (ANSI::CSI_PREFIX) + juce::String (rows) + "B");
}

void Writer::moveToCol (int col)
{
    emit (juce::String (ANSI::CSI_PREFIX) + juce::String (col) + "G");
}

void Writer::moveToRow (int row)
{
    emit (juce::String (ANSI::CSI_PREFIX) + juce::String (row) + "d");
}

void Writer::moveTo (int row, int col)
{
    emit (juce::String (ANSI::CSI_PREFIX) + juce::String (row) + ";" + juce::String (col) + "H");
}

void Writer::enterAltScreen()
{
    emit (ANSI::ALT_SCREEN_ON);
}

void Writer::exitAltScreen()
{
    emit (ANSI::ALT_SCREEN_OFF);
}

void Writer::clearScreen()
{
    emit (ANSI::CLEAR_SCREEN);
}

void Writer::clearLine()
{
    emit (ANSI::CLEAR_LINE);
}

void Writer::clearToEndOfLine()
{
    emit (ANSI::CLEAR_TO_END);
}

void Writer::hideCursor()
{
    emit (ANSI::CURSOR_HIDE);
}

void Writer::showCursor()
{
    emit (ANSI::CURSOR_SHOW);
}

void Writer::writeLine (const juce::String& ansiLine)
{
    emit (ansiLine);
    emit (ANSI::LINE_SUFFIX);
}

void Writer::setFg (juce::Colour c)
{
    emit (juce::String (ANSI::CSI_PREFIX)
          + "38;2;"
          + juce::String (c.getRed())   + ";"
          + juce::String (c.getGreen()) + ";"
          + juce::String (c.getBlue())  + "m");
}

void Writer::setBg (juce::Colour c)
{
    emit (juce::String (ANSI::CSI_PREFIX)
          + "48;2;"
          + juce::String (c.getRed())   + ";"
          + juce::String (c.getGreen()) + ";"
          + juce::String (c.getBlue())  + "m");
}

void Writer::resetAttrs()
{
    emit (ANSI::RESET);
}

void Writer::setBold (bool b)
{
    if (b)
        emit (ANSI::BOLD_ON);
    else
        emit (ANSI::BOLD_OFF);
}

void Writer::setItalic (bool b)
{
    if (b)
        emit (ANSI::ITALIC_ON);
    else
        emit (ANSI::ITALIC_OFF);
}

void Writer::setUnderline (bool b)
{
    if (b)
        emit (ANSI::UNDERLINE_ON);
    else
        emit (ANSI::UNDERLINE_OFF);
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
