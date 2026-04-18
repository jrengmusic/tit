#include <JuceHeader.h>
#include "HistoryView.h"

// ---- Fallback colours -------------------------------------------------------
static const juce::Colour FALLBACK_DIMMED_TEXT   { juce::Colour { 0xff808080 } };
static const juce::Colour FALLBACK_CONTENT_TEXT  { juce::Colour { 0xffcccccc } };
static const juce::Colour FALLBACK_LABEL_TEXT    { juce::Colour { 0xffffffff } };

namespace tit
{

HistoryView::HistoryView (const juce::ValueTree& stateTree,
                           jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
    , commitPane    { "Commits" }
    , splitPane     { &commitPane, &detailPane, COMMIT_LIST_PANE_WIDTH }
{
    historyTree   = stateTree.getChildWithName (ID::HISTORY);
    selectionTree = stateTree.getChildWithName (ID::SELECTION);

    historyTree.addListener   (this);
    selectionTree.addListener (this);

    addAndMakeVisible (splitPane);
    rebuildList ();
}

HistoryView::~HistoryView()
{
    historyTree.removeListener   (this);
    selectionTree.removeListener (this);
}

// ============================================================================
// Component overrides
// ============================================================================

void HistoryView::resized()
{
    splitPane.setBounds (getLocalBounds());
}

void HistoryView::paint (jam::tui::Graphics&)
{
    // SplitPane paints its children.
}

void HistoryView::handleInput (const jam::tui::KeyEvent& event)
{
    const bool isEnter { event.type == jam::tui::KeyType::Enter };
    const bool listFocused { splitPane.getFocusedChildIndex() == 0 };

    if (isEnter and listFocused)
    {
        const int idx { commitPane.getSelectedIndex() };
        selectionTree.setProperty (ID::historyIndex, idx, nullptr);
    }
    else
    {
        splitPane.handleInput (event);
    }
}

// ============================================================================
// ValueTree::Listener
// ============================================================================

void HistoryView::valueTreePropertyChanged (juce::ValueTree& tree,
                                             const juce::Identifier& property)
{
    const bool isHistoryIndex { tree.getType() == ID::SELECTION
                                and property == ID::historyIndex };

    if (isHistoryIndex)
    {
        const int idx { static_cast<int> (selectionTree.getProperty (ID::historyIndex)) };
        syncSelection (idx);
    }
}

void HistoryView::valueTreeChildAdded (juce::ValueTree& parent, juce::ValueTree&)
{
    const bool isHistoryChild { parent.getType() == ID::HISTORY };

    if (isHistoryChild)
        rebuildList ();
}

void HistoryView::valueTreeChildRemoved (juce::ValueTree& parent, juce::ValueTree&, int)
{
    const bool isHistoryChild { parent.getType() == ID::HISTORY };

    if (isHistoryChild)
        rebuildList ();
}

// ============================================================================
// Helpers
// ============================================================================

void HistoryView::rebuildList()
{
    const juce::Array<jam::tui::ListItem> items { buildItems() };
    const int visibleRows { commitPane.getHeight() - 2 };
    const int safeVisible { visibleRows > 0 ? visibleRows : 1 };
    commitPane.updateScrollForSelection (safeVisible);
    repaint();
}

void HistoryView::syncSelection (int index)
{
    commitPane.setSelectedIndex (index);
    updateDetail (index);
    repaint();
}

void HistoryView::updateDetail (int index)
{
    const int count { historyTree.getNumChildren() };
    const bool inRange { index >= 0 and index < count };

    if (inRange)
    {
        const juce::ValueTree commit { historyTree.getChild (index) };
        const juce::String    hash   { commit.getProperty (ID::hash).toString() };
        const juce::String    author { commit.getProperty (ID::author).toString() };
        const juce::String    date   { commit.getProperty (ID::date).toString() };
        const juce::String    msg    { commit.getProperty (ID::message).toString() };

        const juce::String content { "Author: " + author + "\n"
                                     "Date:   " + date   + "\n\n"
                                     + msg };
        detailPane.setContent (content, false);
    }
    else
    {
        detailPane.setContent ("(no commit selected)", false);
    }
}

juce::Array<jam::tui::ListItem> HistoryView::buildItems() const
{
    juce::Array<jam::tui::ListItem> items;
    const int count { historyTree.getNumChildren() };
    const int selected { commitPane.getSelectedIndex() };

    const juce::Colour dimColour    { themeResolver.getColour (ID::dimmedTextColor,
                                                                FALLBACK_DIMMED_TEXT) };
    const juce::Colour contentColour{ themeResolver.getColour (ID::contentTextColor,
                                                                FALLBACK_CONTENT_TEXT) };

    for (int i { 0 }; i < count; ++i)
    {
        const juce::ValueTree child  { historyTree.getChild (i) };
        const juce::String    hash   { child.getProperty (ID::hash).toString() };
        const juce::String    date   { child.getProperty (ID::date).toString() };

        const juce::String shortHash { hash.length() >= 7 ? hash.substring (0, 7) : hash };
        const juce::String dateShort { date.length() >= 10 ? date.substring (0, 10) : date };

        jam::tui::ListItem item;
        item.attributeText   = dateShort;
        item.attributeColour = dimColour;
        item.contentText     = shortHash;
        item.contentColour   = contentColour;
        item.contentBold     = (i == selected);
        items.add (item);
    }

    return items;
}

} // namespace tit
