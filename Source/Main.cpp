#include <JuceHeader.h>
#include "generated/Generated.h"
#ifdef _WIN32
#include <io.h>
#include <windows.h>
#else
#include <unistd.h>
#endif

static void printError (std::string_view argument, std::string_view message)
{
    std::string errorLine { ProjectInfo::projectName };
    errorLine.append (jam::Strings::getView (Id::diagnosticSeparator));
    errorLine.append (argument);
    errorLine.append (jam::Strings::getView (Id::diagnosticSeparator));
    errorLine.append (message);
    jam::Format::appendUtf8 (errorLine, Chars::newline);
    fwrite (errorLine.data(), sizeof (char), errorLine.size(), stderr);
}

static constexpr int flagArgIndex { 0 };
static constexpr int flagOnlyArgumentCount { 1 };

static void writeVersion()
{
    std::string versionLine { ProjectInfo::projectName };
    jam::Format::appendUtf8 (versionLine, Chars::space);
    versionLine.append (ProjectInfo::versionString);
    jam::Format::appendUtf8 (versionLine, Chars::space);
    jam::Format::appendUtf8 (versionLine, Chars::openParen);
    versionLine.append (CAST_COMMIT);
    jam::Format::appendUtf8 (versionLine, Chars::closeParen);
    jam::Format::appendUtf8 (versionLine, Chars::newline);
    fwrite (versionLine.data(), sizeof (char), versionLine.size(), stdout);
}

static void writeHelp()
{
    const BinaryData::Raw binary { files::titcHelp };
    assert (binary.exists());
    fwrite (binary.data, sizeof (char), static_cast<size_t> (binary.size), stdout);
}

static bool isTerminalOutput() noexcept
{
#ifdef _WIN32
    return _isatty (_fileno (stdout)) != 0;
#else
    return isatty (fileno (stdout)) != 0;
#endif
}

static const jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, int ()>& getFlags()
{
    static const auto flags (
        []
        {
            jam::Function::Map<juce::Identifier, jam::Function::pointerCapacity, int ()> flags;

            flags.add (Id::version, [] (const juce::Identifier&) { writeVersion(); return EXIT_SUCCESS; });
            flags.add (Id::help, [] (const juce::Identifier&) { writeHelp(); return EXIT_SUCCESS; });

            return flags;
        }()
    );

    return flags;
}

static std::string_view getFlagWord (std::string_view argument)
{
    const std::string_view doubleDash { Chars::doubleDash };

    if (argument.compare (0, doubleDash.size(), doubleDash) == 0)
        return argument.substr (doubleDash.size());

    return {};
}

static int runCommandLine (const juce::StringArray& arguments)
{
    const auto word { arguments.size() == flagOnlyArgumentCount ? getFlagWord (jam::Strings::getView (arguments[flagArgIndex])) : std::string_view {} };
    const juce::Identifier flag { word.empty() ? juce::Identifier {} : juce::Identifier { jam::Strings::fromView (word) } };

    if (getFlags().contains (flag)) return getFlags().get (flag);

    if (arguments.size() >= flagOnlyArgumentCount)
    {
        printError (jam::Strings::getView (arguments[flagArgIndex]), text::Diagnostics::failFlagUnknown);
        return EXIT_FAILURE;
    }

    writeHelp();
    return EXIT_SUCCESS;
}

class Application : public juce::JUCEApplicationBase
{
public:
    Application() = default;

    const juce::String getApplicationName() override { return ProjectInfo::projectName; }
    const juce::String getApplicationVersion() override { return ProjectInfo::versionString; }
    bool moreThanOneInstanceAllowed() override { return true; }

    void initialise (const juce::String&) override
    {
#ifdef _WIN32
        SetConsoleOutputCP (CP_UTF8);
#endif

        if (isTerminalOutput())
        {
            const BinaryData::Raw erase { files::screenErase };
            assert (erase.exists());
            fwrite (erase.data, sizeof (char), static_cast<size_t> (erase.size), stdout);
            fflush (stdout);
        }

        setApplicationReturnValue (runCommandLine (getCommandLineParameterArray()));
        quit();
    }

    void shutdown() override {}
    void anotherInstanceStarted (const juce::String&) override {}
    void systemRequestedQuit() override { quit(); }
    void suspended() override {}
    void resumed() override {}

    void unhandledException (const std::exception*, const juce::String&, int) override
    {
        jassertfalse;
    }

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Application)
};

START_JUCE_APPLICATION (Application)
