#pragma once

#if JUCE_MAC or JUCE_LINUX
 #include <sys/ioctl.h>
 #include <unistd.h>
#endif

#if JUCE_WINDOWS
 #ifndef NOMINMAX
  #define NOMINMAX
 #endif
 #include <windows.h>
#endif

namespace jreng::tui
{ /*____________________________________________________________________________*/

static constexpr int defaultCols { 80 };
static constexpr int defaultRows { 24 };

/** Returns the current terminal dimensions as a Rectangle { 0, 0, cols, rows }.
 *  Queries the platform on every call — no internal state.
 *  Falls back to 80 x 24 when the platform query fails.
 */
inline Rectangle getBounds() noexcept
{
    int cols { defaultCols };
    int rows { defaultRows };

#if JUCE_MAC or JUCE_LINUX

    struct winsize ws {};

    if (ioctl (STDOUT_FILENO, TIOCGWINSZ, &ws) == 0)
    {
        cols = ws.ws_col;
        rows = ws.ws_row;
    }

#elif JUCE_WINDOWS

    CONSOLE_SCREEN_BUFFER_INFO csbi {};

    if (GetConsoleScreenBufferInfo (GetStdHandle (STD_OUTPUT_HANDLE), &csbi) != 0)
    {
        cols = csbi.srWindow.Right  - csbi.srWindow.Left + 1;
        rows = csbi.srWindow.Bottom - csbi.srWindow.Top  + 1;
    }

#endif

    return { 0, 0, cols, rows };
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
