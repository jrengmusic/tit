#pragma once

namespace ANSI
{ /*____________________________________________________________________________*/

    constexpr const char* SYNC_START    = "\x1b[?2026h";
    constexpr const char* SYNC_END      = "\x1b[?2026l";
    constexpr const char* ALT_SCREEN_ON  = "\x1b[?1049h";
    constexpr const char* ALT_SCREEN_OFF = "\x1b[?1049l";
    constexpr const char* CURSOR_HOME   = "\x1b[H";
    constexpr const char* CURSOR_HIDE   = "\x1b[?25l";
    constexpr const char* CURSOR_SHOW   = "\x1b[?25h";
    constexpr const char* CLEAR_SCREEN  = "\x1b[2J";
    constexpr const char* CLEAR_SCROLLBACK = "\x1b[3J";
    constexpr const char* CLEAR_LINE    = "\x1b[2K";
    constexpr const char* CLEAR_TO_END  = "\x1b[K";
    constexpr const char* RESET         = "\x1b[0m";
    constexpr const char* BOLD_ON       = "\x1b[1m";
    constexpr const char* BOLD_OFF      = "\x1b[22m";
    constexpr const char* ITALIC_ON     = "\x1b[3m";
    constexpr const char* ITALIC_OFF    = "\x1b[23m";
    constexpr const char* UNDERLINE_ON  = "\x1b[4m";
    constexpr const char* UNDERLINE_OFF = "\x1b[24m";
    constexpr const char* DIM_ON        = "\x1b[2m";
    constexpr const char* DIM_OFF       = "\x1b[22m";
    constexpr const char* CURSOR_MARKER = "\x1b_caroline:cursor\x07";
    constexpr const char* PASTE_START   = "\x1b[?2004h";
    constexpr const char* PASTE_END     = "\x1b[?2004l";
    constexpr const char* OSC_RESET     = "\x1b]8;;\x07";
    constexpr const char* LINE_SUFFIX   = "\x1b[0m\x1b]8;;\x07";
    constexpr const char* CSI_PREFIX    = "\x1b[";

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace ANSI
