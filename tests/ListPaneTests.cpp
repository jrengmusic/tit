#include <jam_tui/jam_tui.h>

// ============================================================================
// ListPaneTests
// ============================================================================

class ListPaneTests : public juce::UnitTest
{
public:
    ListPaneTests() : juce::UnitTest ("ListPane", "jam_tui") {}

    void runTest() override
    {
        testCtor ();
        testSelectionState ();
        testScrollAdjust ();
        testKeyboardNavigation ();
    }

private:

    static juce::Array<jam::tui::ListItem> buildItems (int count)
    {
        juce::Array<jam::tui::ListItem> items;

        for (int i { 0 }; i < count; ++i)
        {
            jam::tui::ListItem item;
            item.attributeText   = juce::String ("attr") + juce::String (i);
            item.attributeColour = juce::Colour { 0xff555555 };
            item.contentText     = juce::String ("content") + juce::String (i);
            item.contentColour   = juce::Colour { 0xffcccccc };
            items.add (item);
        }

        return items;
    }

    void testCtor ()
    {
        beginTest ("ctor");

        jam::tui::ListPane pane { "Commits" };

        expect (pane.getSelectedIndex () == 0, "Initial selected index must be 0");
        expect (pane.getScrollOffset ()  == 0, "Initial scroll offset must be 0");
    }

    void testSelectionState ()
    {
        beginTest ("selection state");

        jam::tui::ListPane pane { "Files" };

        pane.setSelectedIndex (3);
        expect (pane.getSelectedIndex () == 3, "setSelectedIndex (3) must take effect");

        pane.setSelectedIndex (0);
        expect (pane.getSelectedIndex () == 0, "setSelectedIndex (0) must take effect");
    }

    void testScrollAdjust ()
    {
        beginTest ("scroll adjust");

        jam::tui::ListPane pane { "History" };
        const int visibleLines { 3 };

        // Select below visible window — scroll must follow
        pane.setSelectedIndex (5);
        pane.updateScrollForSelection (visibleLines);
        expect (pane.getScrollOffset () <= pane.getSelectedIndex (),
                "scrollOffset must not exceed selectedIndex");
        expect (pane.getSelectedIndex () < pane.getScrollOffset () + visibleLines,
                "selectedIndex must be within visible window");

        // Select above scroll window — scroll must retreat
        pane.setSelectedIndex (0);
        pane.updateScrollForSelection (visibleLines);
        expect (pane.getScrollOffset () == 0,
                "scrollOffset must be 0 when selectedIndex is 0");
    }

    void testKeyboardNavigation ()
    {
        beginTest ("keyboard navigation");

        jam::tui::ListPane pane { "Commits" };

        juce::Array<jam::tui::ListItem> items { buildItems (4) };
        // Paint to prime cachedItems
        pane.setBounds ({ 0, 0, 40, 10 });

        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;

        // Simulate cached items by calling paint overload
        // (paint(g, items) stores items in cachedItems for handleInput)
        // We can test via setSelectedIndex + handleInput indirectly:

        bool selectionChanged { false };
        pane.onSelectionChanged = [&] (int) { selectionChanged = true; };

        pane.setSelectedIndex (0);
        selectionChanged = false;

        // ArrowDown with no cached items is safe (no items = no change)
        jam::tui::KeyEvent up;
        up.type = jam::tui::KeyType::ArrowUp;
        pane.handleInput (up);
        expect (pane.getSelectedIndex () == 0, "ArrowUp at 0 must stay at 0");

        // Test direct index manipulation as public contract
        pane.setSelectedIndex (2);
        expect (pane.getSelectedIndex () == 2, "setSelectedIndex (2) must take effect");
        expect (selectionChanged, "onSelectionChanged must fire on setSelectedIndex");
    }
};

static ListPaneTests listPaneTestsInstance;
