#pragma once

#include <functional>
#include <atomic>

#if JUCE_WINDOWS
 #ifndef NOMINMAX
  #define NOMINMAX
 #endif
 #include <windows.h>
#else
 #include <termios.h>
#endif

namespace jreng::tui
{ /*____________________________________________________________________________*/

class Input : private juce::Thread
{
public:
    Input();
    ~Input() override;

    void start (std::function<void(KeyEvent)> onKey,
                std::function<void()> onResize);
    void stop();

private:
    void run() override;
    void enterRawMode();
    void restoreMode();
    KeyEvent parseSequence (const juce::MemoryBlock& bytes);

    struct SavedMode
    {
#if JUCE_WINDOWS
        DWORD originalMode { 0 };
#else
        struct termios originalTermios {};
#endif
    };

    SavedMode savedMode;
    std::function<void(KeyEvent)> keyCallback;
    std::function<void()> resizeCallback;
    std::atomic<bool> shouldStop { false };
    std::atomic<bool> isRunning { false };

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Input)
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
