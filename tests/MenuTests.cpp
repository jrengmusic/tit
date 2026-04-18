#include <jam_tui/jam_tui.h>

// ============================================================================
// MenuTests
// ============================================================================

class MenuTests : public juce::UnitTest
{
public:
    MenuTests() : juce::UnitTest ("Menu", "jam_tui") {}

    void runTest() override
    {
        testCtorAndBinding ();
        testSelectionState ();
        testKeyboardNavigation ();
        testHotkeyJump ();
    }

private:

    static juce::ValueTree buildMenuTree ()
    {
        juce::ValueTree menu { "MENU" };

        juce::ValueTree item1 { "ITEM" };
        item1.setProperty ("id",          "status",   nullptr);
        item1.setProperty ("label",       "Status",   nullptr);
        item1.setProperty ("hotkey",      "s",        nullptr);
        item1.setProperty ("enabled",     true,       nullptr);
        item1.setProperty ("destructive", false,      nullptr);
        menu.addChild (item1, -1, nullptr);

        juce::ValueTree item2 { "ITEM" };
        item2.setProperty ("id",          "commit",   nullptr);
        item2.setProperty ("label",       "Commit",   nullptr);
        item2.setProperty ("hotkey",      "c",        nullptr);
        item2.setProperty ("enabled",     true,       nullptr);
        item2.setProperty ("destructive", false,      nullptr);
        menu.addChild (item2, -1, nullptr);

        juce::ValueTree item3 { "ITEM" };
        item3.setProperty ("id",          "reset",    nullptr);
        item3.setProperty ("label",       "Reset",    nullptr);
        item3.setProperty ("hotkey",      "r",        nullptr);
        item3.setProperty ("enabled",     true,       nullptr);
        item3.setProperty ("destructive", true,       nullptr);
        menu.addChild (item3, -1, nullptr);

        return menu;
    }

    void testCtorAndBinding ()
    {
        beginTest ("ctor + ValueTree binding");

        juce::ValueTree tree { buildMenuTree () };
        jam::tui::Menu menu { tree };

        expect (menu.getSelectedIndex () == 0,
                "Initial selected index must be 0");

        // Add a child and verify the menu reacts
        juce::ValueTree extra { "ITEM" };
        extra.setProperty ("id",    "push",  nullptr);
        extra.setProperty ("label", "Push",  nullptr);
        extra.setProperty ("hotkey","p",     nullptr);
        extra.setProperty ("enabled", true,  nullptr);
        extra.setProperty ("destructive", false, nullptr);
        tree.addChild (extra, -1, nullptr);

        // selectedIndex should still be valid (not out of range)
        expect (menu.getSelectedIndex () >= 0,
                "selectedIndex must remain non-negative after tree child added");
    }

    void testSelectionState ()
    {
        beginTest ("selection state");

        juce::ValueTree tree { buildMenuTree () };
        jam::tui::Menu menu { tree };

        menu.setSelectedIndex (1);
        expect (menu.getSelectedIndex () == 1, "setSelectedIndex (1) must take effect");

        menu.setSelectedIndex (2);
        expect (menu.getSelectedIndex () == 2, "setSelectedIndex (2) must take effect");

        // Clamp beyond last item
        menu.setSelectedIndex (100);
        expect (menu.getSelectedIndex () == 2, "index must clamp to last item (2)");

        // Negative clamp
        menu.setSelectedIndex (-5);
        expect (menu.getSelectedIndex () == 0, "index must clamp to 0 for negative input");
    }

    void testKeyboardNavigation ()
    {
        beginTest ("keyboard navigation");

        juce::ValueTree tree { buildMenuTree () };
        jam::tui::Menu menu { tree };

        menu.setSelectedIndex (0);

        jam::tui::KeyEvent down;
        down.type = jam::tui::KeyType::ArrowDown;
        menu.handleInput (down);
        expect (menu.getSelectedIndex () == 1, "ArrowDown must advance selection to 1");

        menu.handleInput (down);
        expect (menu.getSelectedIndex () == 2, "ArrowDown must advance selection to 2");

        // Already at last — must not advance past end
        menu.handleInput (down);
        expect (menu.getSelectedIndex () == 2, "ArrowDown at last item must stay at 2");

        jam::tui::KeyEvent up;
        up.type = jam::tui::KeyType::ArrowUp;
        menu.handleInput (up);
        expect (menu.getSelectedIndex () == 1, "ArrowUp must retreat selection to 1");

        menu.handleInput (up);
        expect (menu.getSelectedIndex () == 0, "ArrowUp must retreat selection to 0");

        // Already at first — must not go below 0
        menu.handleInput (up);
        expect (menu.getSelectedIndex () == 0, "ArrowUp at first item must stay at 0");
    }

    void testHotkeyJump ()
    {
        beginTest ("hotkey jump");

        juce::ValueTree tree { buildMenuTree () };
        jam::tui::Menu menu { tree };

        bool callbackFired { false };
        juce::String selectedId;

        menu.onItemSelected = [&] (const jam::tui::MenuItem& item)
        {
            callbackFired = true;
            selectedId    = item.id;
        };

        // Press 'c' — should jump to Commit (index 1) and fire callback
        jam::tui::KeyEvent pressC;
        pressC.type      = jam::tui::KeyType::Character;
        pressC.character = 'c';
        menu.handleInput (pressC);

        expect (menu.getSelectedIndex () == 1,    "'c' hotkey must jump to index 1");
        expect (callbackFired,                    "onItemSelected must fire on hotkey");
        expect (selectedId == "commit",           "onItemSelected item id must be 'commit'");
    }
};

static MenuTests menuTestsInstance;
