// ============================================================================
// Input implementation
// ============================================================================

#if JUCE_MAC or JUCE_LINUX
 #include <sys/select.h>
 #include <termios.h>
 #include <unistd.h>
 #include <csignal>
#endif
#if JUCE_WINDOWS
 #include <windows.h>
#endif
#include <iostream>
#include <unordered_map>

namespace jreng::tui
{ /*____________________________________________________________________________*/

static const std::unordered_map<juce::String, KeyType> escapeTable
{
    { "\x1b[A",   KeyType::ArrowUp    },
    { "\x1b[B",   KeyType::ArrowDown  },
    { "\x1b[C",   KeyType::ArrowRight },
    { "\x1b[D",   KeyType::ArrowLeft  },
    { "\x1b[H",   KeyType::Home       },
    { "\x1b[F",   KeyType::End        },
    { "\x1b[3~",  KeyType::Delete     },
    { "\x1b[5~",  KeyType::PageUp     },
    { "\x1b[6~",  KeyType::PageDown   },
    { "\x03",     KeyType::CtrlC      },
    { "\x04",     KeyType::CtrlD      },
    { "\r",       KeyType::Enter      },
    { "\n",       KeyType::Enter      },
    { "\x7f",     KeyType::Backspace  },
};
static const juce::String pasteBegin { "\x1b[200~" };
static const juce::String pasteEnd   { "\x1b[201~" };
static constexpr int escapeTimeoutMs { 50 };

#if JUCE_MAC or JUCE_LINUX
static std::atomic<bool> sigwinchFlag { false };

static void onSigwinch (int) noexcept
{
    sigwinchFlag.store (true);
}
#endif

Input::Input()
    : juce::Thread ("tui::Input")
{
}

Input::~Input()
{
    stop();
}

void Input::enterRawMode()
{
#if JUCE_MAC or JUCE_LINUX

    tcgetattr (STDIN_FILENO, &savedMode.originalTermios);
    struct termios raw { savedMode.originalTermios };
    raw.c_lflag &= static_cast<tcflag_t> (~ (ECHO | ICANON | ISIG));
    raw.c_cc[VMIN]  = 1;
    raw.c_cc[VTIME] = 0;
    tcsetattr (STDIN_FILENO, TCSANOW, &raw);

#elif JUCE_WINDOWS

    HANDLE handle { GetStdHandle (STD_INPUT_HANDLE) };
    GetConsoleMode (handle, &savedMode.originalMode);
    DWORD newMode { savedMode.originalMode
                    & ~ static_cast<DWORD> (ENABLE_PROCESSED_INPUT
                                            | ENABLE_LINE_INPUT
                                            | ENABLE_ECHO_INPUT) };
    SetConsoleMode (handle, newMode);

#endif
}

void Input::restoreMode()
{
#if JUCE_MAC or JUCE_LINUX

    tcsetattr (STDIN_FILENO, TCSANOW, &savedMode.originalTermios);

#elif JUCE_WINDOWS

    SetConsoleMode (GetStdHandle (STD_INPUT_HANDLE), savedMode.originalMode);

#endif
}

void Input::start (std::function<void(KeyEvent)> onKey,
                   std::function<void()> onResize)
{
    if (not isRunning.load())
    {
        keyCallback    = std::move (onKey);
        resizeCallback = std::move (onResize);
        enterRawMode();
        std::cout << ANSI::PASTE_START << std::flush;

#if JUCE_MAC or JUCE_LINUX
        sigwinchFlag.store (false);
        signal (SIGWINCH, onSigwinch);
#endif

        shouldStop.store (false);
        isRunning.store (true);
        startThread();
    }
}

void Input::stop()
{
    if (isRunning.load())
    {
        shouldStop.store (true);
#if JUCE_WINDOWS
        CancelIoEx (GetStdHandle (STD_INPUT_HANDLE), nullptr);
#endif
        isRunning.store (false);
        stopThread (2000);

#if JUCE_MAC or JUCE_LINUX
        signal (SIGWINCH, SIG_DFL);
#endif

        std::cout << ANSI::PASTE_END << std::flush;
        restoreMode();
    }
}

static void dispatchEvent (const KeyEvent& event,
                           const std::function<void(KeyEvent)>& callback)
{
    juce::MessageManager::callAsync ([event, callback] { callback (event); });
}

static KeyEvent resolveTableMiss (const juce::String& seq)
{
    KeyEvent event {};

    if (seq == "\x1b")
    {
        event.type = KeyType::Escape;
    }
    else if (seq.length() == 1)
    {
        event.type      = KeyType::Character;
        event.character = static_cast<juce::juce_wchar> (seq[0]);
    }
    else
    {
        event.type = KeyType::Unknown;
    }

    return event;
}

KeyEvent Input::parseSequence (const juce::MemoryBlock& bytes)
{
    const juce::String seq { juce::CharPointer_UTF8 (
        static_cast<const char*> (bytes.getData())),
        bytes.getSize() };

    KeyEvent event {};

    if (seq.startsWith (pasteBegin) and seq.endsWith (pasteEnd))
    {
        event.type         = KeyType::Paste;
        event.pasteContent = seq.substring (
            pasteBegin.length(),
            seq.length() - pasteEnd.length());
    }
    else
    {
        const auto found { escapeTable.find (seq) };

        if (found != escapeTable.end())
            event.type = found->second;
        else
            event = resolveTableMiss (seq);
    }

    return event;
}

#if JUCE_MAC or JUCE_LINUX

static juce::MemoryBlock readEscapeSequence()
{
    juce::MemoryBlock bytes {};
    char byte { '\x1b' };
    bytes.append (&byte, 1);

    fd_set fds {};
    struct timeval timeout { 0, escapeTimeoutMs * 1000 };

    FD_ZERO (&fds);
    FD_SET (STDIN_FILENO, &fds);

    while (select (STDIN_FILENO + 1, &fds, nullptr, nullptr, &timeout) > 0)
    {
        if (read (STDIN_FILENO, &byte, 1) == 1)
        {
            bytes.append (&byte, 1);
            FD_ZERO (&fds);
            FD_SET (STDIN_FILENO, &fds);
            timeout = { 0, escapeTimeoutMs * 1000 };
        }
    }

    return bytes;
}

void Input::run()
{
    while (not shouldStop.load())
    {
        if (sigwinchFlag.exchange (false))
        {
            auto cb { resizeCallback };
            juce::MessageManager::callAsync ([cb] { cb(); });
        }

        char byte {};
        const auto bytesRead { read (STDIN_FILENO, &byte, 1) };

        if (bytesRead == 1)
        {
            juce::MemoryBlock bytes {};

            if (byte == '\x1b')
            {
                bytes = readEscapeSequence();
            }
            else
            {
                bytes.append (&byte, 1);
            }

            dispatchEvent (parseSequence (bytes), keyCallback);
        }
    }
}

#elif JUCE_WINDOWS

static juce::MemoryBlock readEscapeSequenceWin (HANDLE handle)
{
    juce::MemoryBlock bytes {};
    char byte { '\x1b' };
    bytes.append (&byte, 1);
    DWORD bytesRead { 0 };

    while (WaitForSingleObject (handle, escapeTimeoutMs) == WAIT_OBJECT_0)
    {
        if (ReadFile (handle, &byte, 1, &bytesRead, nullptr) and bytesRead == 1)
            bytes.append (&byte, 1);
    }

    return bytes;
}

void Input::run()
{
    // TODO: Windows resize detection — poll or signal from host
    HANDLE handle { GetStdHandle (STD_INPUT_HANDLE) };

    while (not shouldStop.load())
    {
        char byte {};
        DWORD bytesRead { 0 };
        if (ReadFile (handle, &byte, 1, &bytesRead, nullptr) and bytesRead == 1)
        {
            juce::MemoryBlock bytes {};

            if (byte == '\x1b')
                bytes = readEscapeSequenceWin (handle);
            else
                bytes.append (&byte, 1);

            dispatchEvent (parseSequence (bytes), keyCallback);
        }
    }
}

#endif
/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
