#include <JuceHeader.h>
#include "MenuView.h"

namespace tit
{

MenuView::MenuView (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
    , menu          { stateTree.getChildWithName (ID::MENU) }
{
    repoTree      = stateTree.getChildWithName (ID::REPO);
    menuTree      = stateTree.getChildWithName (ID::MENU);
    selectionTree = stateTree.getChildWithName (ID::SELECTION);

    repoTree.addListener (this);
    selectionTree.addListener (this);

    menu.onItemSelected = [this] (const jam::tui::MenuItem& item)
    {
        if (onItemSelected)
            onItemSelected (item);
    };

    addAndMakeVisible (menu);

    // Populate initial state.
    rebuildMenuTree (repoTree);
}

MenuView::~MenuView()
{
    repoTree.removeListener (this);
    selectionTree.removeListener (this);
}

void MenuView::resized()
{
    menu.setBounds (getLocalBounds());
}

void MenuView::paint (jam::tui::Graphics&)
{
    // Menu primitive paints itself; nothing to do at this level.
}

void MenuView::handleInput (const jam::tui::KeyEvent& event)
{
    menu.handleInput (event);
}

void MenuView::valueTreePropertyChanged (juce::ValueTree& tree,
                                          const juce::Identifier& property)
{
    const bool isRepoOperation { tree.getType() == ID::REPO
                                 and property == ID::operation };

    const bool isMenuIndex { tree.getType() == ID::SELECTION
                             and property == ID::menuIndex };

    if (isRepoOperation)
        rebuildMenuTree (repoTree);

    if (isMenuIndex)
    {
        const int idx { static_cast<int> (selectionTree.getProperty (ID::menuIndex)) };
        menu.setSelectedIndex (idx);
    }
}

void MenuView::rebuildMenuTree (const juce::ValueTree& repo)
{
    const juce::Array<menu::MenuItemDef> items { builder.build (repo) };

    // Remove all existing ITEM children from the MENU subtree.
    menuTree.removeAllChildren (nullptr);

    for (const menu::MenuItemDef& def : items)
    {
        juce::ValueTree item { ID::ITEM };
        item.setProperty (ID::id,          juce::String::fromUTF8 (def.id),    nullptr);
        item.setProperty (ID::label,       juce::String::fromUTF8 (def.label), nullptr);
        item.setProperty (ID::hotkey,      juce::String::charToString (
                              static_cast<juce::juce_wchar> (def.hotkey)),      nullptr);
        item.setProperty (ID::enabled,     true,                                nullptr);
        item.setProperty (ID::destructive, def.destructive,                     nullptr);
        menuTree.appendChild (item, nullptr);
    }
}

} // namespace tit
