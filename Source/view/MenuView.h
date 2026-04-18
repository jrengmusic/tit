#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"
#include "menu/MenuBuilder.h"

namespace tit
{

// ============================================================================
// MenuView
// ============================================================================
//
// Wraps jam::tui::Menu and drives it from VT state.
//
// Listener surface:
//   - ID::REPO subtree — on operation change: rebuild MENU subtree via
//     MenuBuilder::build(), which the contained Menu primitive observes.
//   - ID::SELECTION subtree — on menuIndex change: sync selected row.
//
// Zero explicit rebuild() calls from outside.  The REPO listener triggers
// MenuBuilder, which mutates MENU; the Menu primitive's own listener renders.
//
// BLESSED E: no Source/git/ imports; reads VT only, writes MENU subtree.
// BLESSED S: no shadow item list; Menu primitive is the single source of truth.

class MenuView : public jam::tui::Component,
                 private juce::ValueTree::Listener
{
public:
    MenuView (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~MenuView() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g) override;
    void handleInput (const jam::tui::KeyEvent& event) override;
    bool isFocusable () const override { return true; }
    void resized ()                    override;

    // -------------------------------------------------------------------------
    // Callback — forwarded from Menu primitive
    // -------------------------------------------------------------------------
    std::function<void (const jam::tui::MenuItem&)> onItemSelected;

private:
    juce::ValueTree            repoTree;
    juce::ValueTree            menuTree;
    juce::ValueTree            selectionTree;
    jam::tui::ThemeResolver& themeResolver;
    menu::MenuBuilder          builder;
    jam::tui::Menu           menu;

    // -------------------------------------------------------------------------
    // ValueTree::Listener overrides
    // -------------------------------------------------------------------------
    void valueTreePropertyChanged  (juce::ValueTree& tree, const juce::Identifier& property) override;
    void valueTreeChildAdded       (juce::ValueTree&, juce::ValueTree&)        override {}
    void valueTreeChildRemoved     (juce::ValueTree&, juce::ValueTree&, int)   override {}
    void valueTreeChildOrderChanged(juce::ValueTree&, int, int)                override {}
    void valueTreeParentChanged    (juce::ValueTree&)                          override {}

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------
    void rebuildMenuTree (const juce::ValueTree& repo);

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (MenuView)
};

} // namespace tit
