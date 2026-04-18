#include <JuceHeader.h>
#include "ConflictResolverView.h"

// ---- Fallback colours -------------------------------------------------------
static const juce::Colour FALLBACK_CONFLICT_FOCUSED   { juce::Colour { 0xff44ffcc } };
static const juce::Colour FALLBACK_CONFLICT_UNFOCUSED { juce::Colour { 0xff555555 } };

namespace tit
{

ConflictResolverView::ConflictResolverView (const juce::ValueTree& stateTree,
                                             jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
    , topSplit      { &oursPane, &theirsPane }
{
    diffTree = stateTree.getChildWithName (ID::DIFF);
    diffTree.addListener (this);

    oursPane.setContent   ("(ours)",   false);
    theirsPane.setContent ("(theirs)", false);
    mergedPane.setContent ("(merged)", false);

    addAndMakeVisible (topSplit);
    addAndMakeVisible (mergedPane);

    updateFocusStyles ();
    rebuildPanes ();
}

ConflictResolverView::~ConflictResolverView()
{
    diffTree.removeListener (this);
}

// ============================================================================
// Component overrides
// ============================================================================

void ConflictResolverView::resized()
{
    const int totalH { getHeight() };
    const int topH   { totalH / TOP_ROW_FRACTION_DENOM };
    const int botH   { totalH - topH };

    topSplit.setBounds (0, 0,    getWidth(), topH);
    mergedPane.setBounds (0, topH, getWidth(), botH);
}

void ConflictResolverView::paint (jam::tui::Graphics&)
{
    // Children paint themselves.
}

void ConflictResolverView::handleInput (const jam::tui::KeyEvent& event)
{
    const bool isTab { event.type == jam::tui::KeyType::Tab };

    if (isTab)
    {
        cycleFocus ();
    }
    else
    {
        const bool mergedActive { focusedPaneIndex == PANE_MERGED };

        if (mergedActive)
        {
            mergedPane.handleInput (event);
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

void ConflictResolverView::valueTreePropertyChanged (juce::ValueTree& tree,
                                                      const juce::Identifier& property)
{
    const bool isDiffLines { tree.getType() == ID::DIFF
                             and property == ID::lines };

    if (isDiffLines)
        rebuildPanes ();
}

// ============================================================================
// Helpers
// ============================================================================

void ConflictResolverView::rebuildPanes()
{
    const juce::String rawDiff { diffTree.getProperty (ID::lines).toString() };
    const bool         hasDiff { rawDiff.isNotEmpty() };

    // Parse conflict markers from raw diff.
    // Format: <<<<<<< ours ... ======= ... >>>>>>> theirs
    juce::String oursContent;
    juce::String theirsContent;

    if (hasDiff)
    {
        enum class Region { Before, Ours, Theirs };
        Region current { Region::Before };

        const juce::StringArray rawLines { juce::StringArray::fromLines (rawDiff) };

        for (const juce::String& line : rawLines)
        {
            const bool isOursMarker   { line.startsWith ("<<<<<<<") };
            const bool isSepMarker    { line.startsWith ("=======") };
            const bool isTheirsMarker { line.startsWith (">>>>>>>") };

            if (isOursMarker)
            {
                current = Region::Ours;
            }
            else if (isSepMarker and current == Region::Ours)
            {
                current = Region::Theirs;
            }
            else if (isTheirsMarker)
            {
                current = Region::Before;
            }
            else if (current == Region::Ours)
            {
                oursContent   += line + "\n";
            }
            else if (current == Region::Theirs)
            {
                theirsContent += line + "\n";
            }
        }
    }

    oursPane.setContent   (oursContent.isNotEmpty()   ? oursContent   : "(no ours content)",   false);
    theirsPane.setContent (theirsContent.isNotEmpty() ? theirsContent : "(no theirs content)", false);
    mergedPane.setContent (oursContent + theirsContent, false);

    repaint();
}

void ConflictResolverView::cycleFocus()
{
    focusedPaneIndex = (focusedPaneIndex + 1) % PANE_COUNT;
    updateFocusStyles ();
}

void ConflictResolverView::updateFocusStyles()
{
    oursPane.setActive   (focusedPaneIndex == PANE_OURS);
    theirsPane.setActive (focusedPaneIndex == PANE_THEIRS);
    mergedPane.setActive (focusedPaneIndex == PANE_MERGED);

    const bool topPaneFocused { focusedPaneIndex == PANE_OURS
                                or focusedPaneIndex == PANE_THEIRS };

    if (topPaneFocused)
        topSplit.setFocusedChildIndex (focusedPaneIndex);
}

jam::tui::TextPane& ConflictResolverView::paneForIndex (int index) noexcept
{
    const bool isOurs   { index == PANE_OURS };
    const bool isTheirs { index == PANE_THEIRS };

    if (isOurs)   return oursPane;
    if (isTheirs) return theirsPane;
    return mergedPane;
}

} // namespace tit
