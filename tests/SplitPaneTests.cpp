#include <jam_tui/jam_tui.h>

// ============================================================================
// SplitPaneTests
// ============================================================================

class SplitPaneTests : public juce::UnitTest
{
public:
    SplitPaneTests() : juce::UnitTest ("SplitPane", "jam_tui") {}

    void runTest() override
    {
        testCtorFocusDefault ();
        testFocusCycling ();
        testFocusCallback ();
        testFixedLeftWidthLayout ();
    }

private:

    // Minimal stub component used as a child placeholder in tests.
    struct StubComponent : public jam::tui::Component
    {
        void paint (jam::tui::Graphics&) override {}
    };

    void testCtorFocusDefault ()
    {
        beginTest ("ctor + focus default");

        StubComponent left;
        StubComponent right;
        jam::tui::SplitPane pane { &left, &right };

        expect (pane.getFocusedChildIndex () == 0,
                "Initial focused child must be index 0 (left)");
    }

    void testFocusCycling ()
    {
        beginTest ("focus cycling");

        StubComponent left;
        StubComponent right;
        jam::tui::SplitPane pane { &left, &right };

        jam::tui::KeyEvent tab;
        tab.type = jam::tui::KeyType::Tab;

        pane.handleInput (tab);
        expect (pane.getFocusedChildIndex () == 1,
                "Tab must advance focus to right child (index 1)");

        pane.handleInput (tab);
        expect (pane.getFocusedChildIndex () == 0,
                "Tab must wrap focus back to left child (index 0)");
    }

    void testFocusCallback ()
    {
        beginTest ("focus callback");

        StubComponent left;
        StubComponent right;
        jam::tui::SplitPane pane { &left, &right };

        int lastFocusedIndex { -1 };
        pane.onFocusChanged = [&] (int idx) { lastFocusedIndex = idx; };

        pane.setFocusedChildIndex (1);
        expect (lastFocusedIndex == 1,
                "onFocusChanged must fire with index 1 on setFocusedChildIndex (1)");

        pane.setFocusedChildIndex (0);
        expect (lastFocusedIndex == 0,
                "onFocusChanged must fire with index 0 on setFocusedChildIndex (0)");

        // Setting same index must not fire callback again
        lastFocusedIndex = -99;
        pane.setFocusedChildIndex (0);
        expect (lastFocusedIndex == -99,
                "onFocusChanged must NOT fire when index is unchanged");
    }

    void testFixedLeftWidthLayout ()
    {
        beginTest ("fixed left width layout");

        StubComponent left;
        StubComponent right;

        const int fixedWidth { 30 };
        jam::tui::SplitPane pane { &left, &right, fixedWidth };

        pane.setBounds ({ 0, 0, 80, 24 });
        pane.resized ();

        expect (left.getBounds ().getWidth ()  == fixedWidth,
                "Left child width must equal fixedLeftWidth");
        expect (right.getBounds ().getWidth () == 80 - fixedWidth,
                "Right child width must be totalWidth - fixedLeftWidth");
    }
};

static SplitPaneTests splitPaneTestsInstance;
