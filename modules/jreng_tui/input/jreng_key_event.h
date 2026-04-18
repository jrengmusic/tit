#pragma once

#include <juce_core/juce_core.h>

namespace jreng::tui
{ /*____________________________________________________________________________*/

enum class KeyType
{
    Character,
    Enter,
    Escape,
    Backspace,
    Delete,
    Tab,
    ArrowUp,
    ArrowDown,
    ArrowLeft,
    ArrowRight,
    Home,
    End,
    PageUp,
    PageDown,
    FunctionKey,
    CtrlC,
    CtrlD,
    Paste,
    Unknown
};

struct KeyEvent
{
    KeyType type { KeyType::Unknown };
    juce::juce_wchar character { 0 };
    int functionKeyNumber { 0 };
    juce::String pasteContent;
    bool ctrl { false };
    bool alt { false };
    bool shift { false };
};

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
