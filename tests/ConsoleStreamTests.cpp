#include <jam_tui/jam_tui.h>

// ============================================================================
// ConsoleStreamTests
// ============================================================================

class ConsoleStreamTests : public juce::UnitTest
{
public:
    ConsoleStreamTests() : juce::UnitTest ("ConsoleStream", "jam_tui") {}

    void runTest() override
    {
        testCtorAndBinding ();
        testAutoScroll ();
        testKeyboardScrolling ();
        testStreamTypeColourMapping ();
    }

private:

    static juce::ValueTree buildConsoleTree (int lineCount = 0)
    {
        juce::ValueTree console { "CONSOLE" };

        for (int i { 0 }; i < lineCount; ++i)
        {
            juce::ValueTree line { "LINE" };
            line.setProperty ("text",   "Output line " + juce::String (i), nullptr);
            line.setProperty ("stream", "stdout",                           nullptr);
            console.addChild (line, -1, nullptr);
        }

        return console;
    }

    void testCtorAndBinding ()
    {
        beginTest ("ctor + ValueTree binding");

        juce::ValueTree tree { buildConsoleTree (3) };
        jam::tui::ConsoleStream cs { tree };

        // Adding a child must not crash (VT listener wires up)
        juce::ValueTree extra { "LINE" };
        extra.setProperty ("text",   "Added line", nullptr);
        extra.setProperty ("stream", "info",       nullptr);
        tree.addChild (extra, -1, nullptr);

        expect (true, "ConsoleStream must survive VT child add");

        // Removing a child must not crash
        tree.removeChild (extra, nullptr);
        expect (true, "ConsoleStream must survive VT child remove");
    }

    void testAutoScroll ()
    {
        beginTest ("autoscroll");

        juce::ValueTree tree { buildConsoleTree (0) };
        jam::tui::ConsoleStream cs { tree };

        expect (cs.getAutoScroll (), "Autoscroll must be enabled by default");

        cs.setAutoScroll (false);
        expect (not cs.getAutoScroll (), "setAutoScroll(false) must disable autoscroll");

        cs.setAutoScroll (true);
        expect (cs.getAutoScroll (), "setAutoScroll(true) must re-enable autoscroll");
    }

    void testKeyboardScrolling ()
    {
        beginTest ("keyboard scrolling");

        juce::ValueTree tree { buildConsoleTree (0) };
        jam::tui::ConsoleStream cs { tree };

        // ArrowUp on empty stream — must not crash
        jam::tui::KeyEvent up;
        up.type = jam::tui::KeyType::ArrowUp;
        cs.handleInput (up);
        expect (true, "ArrowUp on empty ConsoleStream must not crash");

        // ArrowDown on empty stream — must not crash
        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;
        cs.handleInput (down);
        expect (true, "ArrowDown on empty ConsoleStream must not crash");

        // ArrowUp disables autoscroll
        for (int i { 0 }; i < 30; ++i)
        {
            juce::ValueTree line { "LINE" };
            line.setProperty ("text",   "Line " + juce::String (i), nullptr);
            line.setProperty ("stream", "stdout",                    nullptr);
            tree.addChild (line, -1, nullptr);
        }

        cs.setAutoScroll (true);
        cs.handleInput (up);

        expect (not cs.getAutoScroll (),
                "ArrowUp must disable autoscroll");

        // ArrowDown to bottom re-enables autoscroll
        for (int i { 0 }; i < 50; ++i)
            cs.handleInput (down);

        expect (cs.getAutoScroll (),
                "ArrowDown past last visible line must re-enable autoscroll");
    }

    void testStreamTypeColourMapping ()
    {
        beginTest ("stream type VT binding");

        juce::ValueTree tree { "CONSOLE" };

        juce::ValueTree stdoutLine { "LINE" };
        stdoutLine.setProperty ("text",   "stdout text", nullptr);
        stdoutLine.setProperty ("stream", "stdout",      nullptr);
        tree.addChild (stdoutLine, -1, nullptr);

        juce::ValueTree stderrLine { "LINE" };
        stderrLine.setProperty ("text",   "stderr text", nullptr);
        stderrLine.setProperty ("stream", "stderr",      nullptr);
        tree.addChild (stderrLine, -1, nullptr);

        juce::ValueTree infoLine { "LINE" };
        infoLine.setProperty ("text",   "info text", nullptr);
        infoLine.setProperty ("stream", "info",      nullptr);
        tree.addChild (infoLine, -1, nullptr);

        jam::tui::ConsoleStream cs { tree };

        // Smoke: construction with mixed stream types must not crash
        expect (true, "ConsoleStream with mixed stream types must not crash");
    }
};

static ConsoleStreamTests consoleStreamTestsInstance;
