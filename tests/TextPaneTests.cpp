#include <jam_tui/jam_tui.h>

// ============================================================================
// TextPaneTests
// ============================================================================

class TextPaneTests : public juce::UnitTest
{
public:
    TextPaneTests() : juce::UnitTest ("TextPane", "jam_tui") {}

    void runTest() override
    {
        testCtorDefaults ();
        testPlainModeNavigation ();
        testDiffParsing ();
        testDiffNavigation ();
        testVisualModeSelection ();
        testDiffRenderingDoesNotCrash ();
    }

private:

    static const char* simpleDiff ()
    {
        return
            "diff --git a/foo.go b/foo.go\n"
            "index abc1234..def5678 100644\n"
            "--- a/foo.go\n"
            "+++ b/foo.go\n"
            "@@ -1,4 +1,4 @@\n"
            " package main\n"
            "-func old() {}\n"
            "+func new() {}\n"
            " \n"
            " // end\n";
    }

    void testCtorDefaults ()
    {
        beginTest ("ctor defaults");

        jam::tui::TextPane tp;

        expect (tp.getLineCursor () == 0,        "Initial lineCursor must be 0");
        expect (not tp.getVisualModeActive (),   "Visual mode must be off by default");
        expect (tp.getVisualModeStart () == 0,   "Initial visualModeStart must be 0");
        expect (tp.getSelectedLines ().isEmpty (),"getSelectedLines must be empty by default");
    }

    void testPlainModeNavigation ()
    {
        beginTest ("plain mode navigation");

        jam::tui::TextPane tp;
        tp.setContent ("line0\nline1\nline2", false);
        tp.setActive (true);

        expect (tp.getLineCursor () == 0, "Cursor must start at 0 after setContent");

        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;
        tp.handleInput (down);
        expect (tp.getLineCursor () == 1, "ArrowDown must advance cursor to 1");

        tp.handleInput (down);
        expect (tp.getLineCursor () == 2, "ArrowDown must advance cursor to 2");

        // At last line — must not advance past end
        tp.handleInput (down);
        expect (tp.getLineCursor () == 2, "ArrowDown at last line must stay at 2");

        jam::tui::KeyEvent up;
        up.type = jam::tui::KeyType::ArrowUp;
        tp.handleInput (up);
        expect (tp.getLineCursor () == 1, "ArrowUp must retreat cursor to 1");

        tp.handleInput (up);
        expect (tp.getLineCursor () == 0, "ArrowUp must retreat cursor to 0");

        // At first line — must not go below 0
        tp.handleInput (up);
        expect (tp.getLineCursor () == 0, "ArrowUp at first line must stay at 0");
    }

    void testDiffParsing ()
    {
        beginTest ("diff parsing");

        jam::tui::TextPane tp;
        tp.setContent (simpleDiff (), true);

        // diff header lines are skipped; we expect:
        // context "package main" (1)
        // removed "func old() {}" (1)
        // added   "func new() {}" (1)
        // context ""              (1)
        // context "// end"        (1)
        // total = 5 diff lines

        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;

        // Navigate to last line (4)
        for (int i { 0 }; i < 10; ++i)
            tp.handleInput (down);

        expect (tp.getLineCursor () == 4,
                "Diff must parse to 5 lines; cursor must clamp to 4");
    }

    void testDiffNavigation ()
    {
        beginTest ("diff mode navigation");

        jam::tui::TextPane tp;
        tp.setContent (simpleDiff (), true);
        tp.setActive (true);

        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;
        tp.handleInput (down);
        expect (tp.getLineCursor () == 1, "ArrowDown in diff mode must advance cursor");

        jam::tui::KeyEvent up;
        up.type = jam::tui::KeyType::ArrowUp;
        tp.handleInput (up);
        expect (tp.getLineCursor () == 0, "ArrowUp in diff mode must retreat cursor");
    }

    void testVisualModeSelection ()
    {
        beginTest ("visual mode selection");

        jam::tui::TextPane tp;
        tp.setContent (simpleDiff (), true);
        tp.setActive (true);

        jam::tui::KeyEvent pressV;
        pressV.type      = jam::tui::KeyType::Character;
        pressV.character = 'v';

        // Activate visual mode at line 0
        tp.handleInput (pressV);
        expect (tp.getVisualModeActive (),
                "'v' must activate visual mode");
        expect (tp.getVisualModeStart () == 0,
                "visualModeStart must be 0 when 'v' pressed at cursor 0");

        // Move cursor down two lines
        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;
        tp.handleInput (down);
        tp.handleInput (down);

        const juce::StringArray selected { tp.getSelectedLines () };
        expect (selected.size () == 3,
                "getSelectedLines must return 3 lines for cursor 0..2");

        // Second 'v' deactivates visual mode
        tp.handleInput (pressV);
        expect (not tp.getVisualModeActive (),
                "Second 'v' must deactivate visual mode");
        expect (tp.getSelectedLines ().isEmpty (),
                "getSelectedLines must be empty after visual mode deactivated");
    }

    void testDiffRenderingDoesNotCrash ()
    {
        beginTest ("diff rendering smoke test");

        jam::tui::TextPane tp;
        tp.setContent (simpleDiff (), true);
        tp.setActive  (true);

        // Smoke: setContent + setActive + setShowLineNumbers must not crash
        tp.setShowLineNumbers (true);
        tp.setShowLineNumbers (false);

        // Empty content must not crash
        tp.setContent ("", true);
        tp.setContent ("", false);

        expect (true, "TextPane rendering helpers must not crash on valid/empty content");
    }
};

static TextPaneTests textPaneTestsInstance;
