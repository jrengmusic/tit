#include <JuceHeader.h>
#include "FileHistoryView.h"

// ---- Fallback colours -------------------------------------------------------
static const juce::Colour FALLBACK_DIMMED_TEXT   { juce::Colour { 0xff808080 } };
static const juce::Colour FALLBACK_CONTENT_TEXT  { juce::Colour { 0xffcccccc } };
static const juce::Colour FALLBACK_ACCENT_TEXT   { juce::Colour { 0xff44ffcc } };

namespace tit
{

FileHistoryView::FileHistoryView (const juce::ValueTree& stateTree,
                                   jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
    , commitPane    { "Commits" }
    , filePane      { "Files" }
    , topSplit      { &commitPane, &filePane, COMMIT_LIST_PANE_WIDTH }
{
    historyTree = stateTree.getChildWithName (ID::HISTORY);
    filesTree   = stateTree.getChildWithName (ID::FILES);
    diffTree    = stateTree.getChildWithName (ID::DIFF);

    historyTree.addListener (this);
    filesTree.addListener   (this);
    diffTree.addListener    (this);

    diffPane.setContent ("(no diff available)", true);

    addAndMakeVisible (topSplit);
    addAndMakeVisible (diffPane);

    rebuildCommits ();
    rebuildFiles   ();
}

FileHistoryView::~FileHistoryView()
{
    historyTree.removeListener (this);
    filesTree.removeListener   (this);
    diffTree.removeListener    (this);
}

// ============================================================================
// Component overrides
// ============================================================================

void FileHistoryView::resized()
{
    const int totalHeight { getHeight() };
    const int topRowH     { totalHeight / TOP_ROW_FRACTION_DENOM };
    const int diffH       { totalHeight - topRowH };

    topSplit.setBounds (0, 0,        getWidth(), topRowH);
    diffPane.setBounds (0, topRowH,  getWidth(), diffH);
}

void FileHistoryView::paint (jam::tui::Graphics& g)
{
    const juce::Rectangle<int> topBounds  { topSplit.getBounds() };
    const juce::Rectangle<int> diffBounds { diffPane.getBounds() };

    if (not topBounds.isEmpty())
    {
        const int leftWidth  { juce::jmin (COMMIT_LIST_PANE_WIDTH, topBounds.getWidth()) };
        const int rightWidth { topBounds.getWidth() - leftWidth };

        jam::tui::Graphics leftCtx { g.clip ({
            topBounds.getX(), topBounds.getY(), leftWidth, topBounds.getHeight() }) };
        commitPane.paint (leftCtx, buildCommitItems());

        if (rightWidth > 0)
        {
            jam::tui::Graphics rightCtx { g.clip ({
                topBounds.getX() + leftWidth, topBounds.getY(),
                rightWidth, topBounds.getHeight() }) };
            filePane.paint (rightCtx, buildFileItems());
        }
    }

    if (not diffBounds.isEmpty())
    {
        jam::tui::Graphics diffCtx { g.clip ({
            diffBounds.getX(), diffBounds.getY(),
            diffBounds.getWidth(), diffBounds.getHeight() }) };
        diffPane.paint (diffCtx);
    }
}

void FileHistoryView::handleInput (const jam::tui::KeyEvent& event)
{
    const bool isTab { event.type == jam::tui::KeyType::Tab };

    if (isTab)
    {
        cycleFocus ();
    }
    else
    {
        const bool diffActive { focusedPaneIndex == PANE_DIFF };

        if (diffActive)
        {
            diffPane.handleInput (event);
        }
        else
        {
            topSplit.handleInput (event);
        }
    }
}

// ============================================================================
// ValueTree::Listener
// ============================================================================

void FileHistoryView::valueTreePropertyChanged (juce::ValueTree& tree,
                                                 const juce::Identifier& property)
{
    const bool isDiffLines { tree.getType() == ID::DIFF
                             and property == ID::lines };

    if (isDiffLines)
        refreshDiff ();
}

void FileHistoryView::valueTreeChildAdded (juce::ValueTree& parent, juce::ValueTree&)
{
    if (parent.getType() == ID::HISTORY)
        rebuildCommits ();
    else if (parent.getType() == ID::FILES)
        rebuildFiles ();
}

void FileHistoryView::valueTreeChildRemoved (juce::ValueTree& parent, juce::ValueTree&, int)
{
    if (parent.getType() == ID::HISTORY)
        rebuildCommits ();
    else if (parent.getType() == ID::FILES)
        rebuildFiles ();
}

// ============================================================================
// Helpers
// ============================================================================

void FileHistoryView::rebuildCommits()
{
    repaint();
}

void FileHistoryView::rebuildFiles()
{
    repaint();
}

void FileHistoryView::refreshDiff()
{
    const juce::String diffContent { diffTree.getProperty (ID::lines).toString() };
    const bool         hasDiff     { diffContent.isNotEmpty() };
    diffPane.setContent (hasDiff ? diffContent : "(no diff available)", true);
    repaint();
}

void FileHistoryView::cycleFocus()
{
    const int nextIndex { (focusedPaneIndex + 1) % 3 };
    focusedPaneIndex = nextIndex;

    const bool diffActive { focusedPaneIndex == PANE_DIFF };
    diffPane.setActive (diffActive);

    if (not diffActive)
        topSplit.setFocusedChildIndex (focusedPaneIndex);
}

juce::Array<jam::tui::ListItem> FileHistoryView::buildCommitItems() const
{
    juce::Array<jam::tui::ListItem> items;
    const int count    { historyTree.getNumChildren() };
    const int selected { commitPane.getSelectedIndex() };

    const juce::Colour dimColour     { themeResolver.getColour (ID::dimmedTextColor,
                                                                  FALLBACK_DIMMED_TEXT) };
    const juce::Colour contentColour { themeResolver.getColour (ID::contentTextColor,
                                                                  FALLBACK_CONTENT_TEXT) };

    for (int i { 0 }; i < count; ++i)
    {
        const juce::ValueTree child    { historyTree.getChild (i) };
        const juce::String    hash     { child.getProperty (ID::hash).toString() };
        const juce::String    date     { child.getProperty (ID::date).toString() };
        const juce::String    shortHash{ hash.length() >= 7 ? hash.substring (0, 7) : hash };
        const juce::String    dateShort{ date.length() >= 10 ? date.substring (0, 10) : date };

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

juce::Array<jam::tui::ListItem> FileHistoryView::buildFileItems() const
{
    juce::Array<jam::tui::ListItem> items;
    const int count    { filesTree.getNumChildren() };
    const int selected { filePane.getSelectedIndex() };

    const juce::Colour dimColour     { themeResolver.getColour (ID::dimmedTextColor,
                                                                  FALLBACK_DIMMED_TEXT) };
    const juce::Colour accentColour  { themeResolver.getColour (ID::accentTextColor,
                                                                  FALLBACK_ACCENT_TEXT) };
    const juce::Colour contentColour { themeResolver.getColour (ID::contentTextColor,
                                                                  FALLBACK_CONTENT_TEXT) };

    for (int i { 0 }; i < count; ++i)
    {
        const juce::ValueTree child  { filesTree.getChild (i) };
        const juce::String    path   { child.getProperty (ID::path).toString() };
        const juce::String    status { child.getProperty (ID::status).toString() };

        jam::tui::ListItem item;
        item.attributeText   = status;
        item.attributeColour = (i == selected) ? accentColour : dimColour;
        item.contentText     = path;
        item.contentColour   = contentColour;
        item.contentBold     = (i == selected);
        items.add (item);
    }

    return items;
}

} // namespace tit
