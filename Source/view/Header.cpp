#include <JuceHeader.h>
#include "Header.h"
#include "state/TitAxis.h"

// ---- Fallback colours — ARGB uint32 via explicit constructor (unambiguous) --
static const juce::Colour FALLBACK_CWD_TEXT       { juce::Colour { 0xffffc8be } };
static const juce::Colour FALLBACK_LABEL_TEXT      { juce::Colour { 0xffffffff } };
static const juce::Colour FALLBACK_STATUS_CLEAN    { juce::Colour { 0xff5fd75f } };
static const juce::Colour FALLBACK_STATUS_DIRTY    { juce::Colour { 0xffffaf00 } };
static const juce::Colour FALLBACK_TIMELINE_SYNC   { juce::Colour { 0xff5fd75f } };
static const juce::Colour FALLBACK_TIMELINE_AHEAD  { juce::Colour { 0xff5fafff } };
static const juce::Colour FALLBACK_TIMELINE_BEHIND { juce::Colour { 0xffffaf00 } };
static const juce::Colour FALLBACK_DIMMED_TEXT     { juce::Colour { 0xff808080 } };

namespace tit
{

Header::Header (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
{
    repoTree = stateTree.getChildWithName (ID::REPO);
    repoTree.addListener (this);
}

Header::~Header()
{
    repoTree.removeListener (this);
}

void Header::resized()
{
    repaint();
}

void Header::paint (jam::tui::Graphics& g)
{
    int row { 0 };
    const int width { getWidth() };

    paintCwdLine      (g, 0, row, width);
    paintBranchLine   (g, 0, row, width);
    paintStatusLine   (g, 0, row, width);
    paintTimelineLine (g, 0, row, width);
}

void Header::paintCwdLine (jam::tui::Graphics& g, int col, int& row, int width) const
{
    const juce::String cwd { repoTree.getProperty (ID::cwd).toString() };
    const juce::Colour colour { themeResolver.getColour (ID::cwdTextColor,
                                                          FALLBACK_CWD_TEXT) };
    g.setColour (colour);
    g.drawCellText ("CWD: " + cwd, col, row, width);
    ++row;
}

void Header::paintBranchLine (jam::tui::Graphics& g, int col, int& row, int width) const
{
    const juce::String branch { repoTree.getProperty (ID::branch).toString() };
    const juce::Colour colour { themeResolver.getColour (ID::labelTextColor,
                                                          FALLBACK_LABEL_TEXT) };
    g.setColour (colour);
    g.drawCellText ("Branch: " + branch, col, row, width);
    ++row;
}

void Header::paintStatusLine (jam::tui::Graphics& g, int col, int& row, int width) const
{
    const juce::String wtStr { repoTree.getProperty (ID::workingTree).toString() };
    const WorkingTree  wt    { parseWorkingTree (wtStr) };

    const bool isClean { wt == WorkingTree::Clean };

    const juce::Colour colour { isClean
        ? themeResolver.getColour (ID::statusClean, FALLBACK_STATUS_CLEAN)
        : themeResolver.getColour (ID::statusDirty, FALLBACK_STATUS_DIRTY) };

    const juce::String label { isClean ? "Clean" : "Modified" };

    g.setColour (colour);
    g.drawCellText (label, col, row, width);
    ++row;
}

void Header::paintTimelineLine (jam::tui::Graphics& g, int col, int& row, int width) const
{
    const juce::String tlStr { repoTree.getProperty (ID::timeline).toString() };
    const Timeline     tl    { parseTimeline (tlStr) };

    const int ahead  { static_cast<int> (repoTree.getProperty (ID::aheadCount)) };
    const int behind { static_cast<int> (repoTree.getProperty (ID::behindCount)) };

    juce::Colour colour;
    juce::String label;

    if (tl == Timeline::InSync)
    {
        colour = themeResolver.getColour (ID::timelineSynchronized, FALLBACK_TIMELINE_SYNC);
        label  = "In sync";
    }
    else if (tl == Timeline::Ahead)
    {
        colour = themeResolver.getColour (ID::timelineLocalAhead, FALLBACK_TIMELINE_AHEAD);
        label  = "Ahead " + juce::String { ahead };
    }
    else if (tl == Timeline::Behind)
    {
        colour = themeResolver.getColour (ID::timelineLocalBehind, FALLBACK_TIMELINE_BEHIND);
        label  = "Behind " + juce::String { behind };
    }
    else if (tl == Timeline::Diverged)
    {
        colour = themeResolver.getColour (ID::timelineLocalBehind, FALLBACK_TIMELINE_BEHIND);
        label  = "Diverged +" + juce::String { ahead } + " -" + juce::String { behind };
    }
    else
    {
        colour = themeResolver.getColour (ID::dimmedTextColor, FALLBACK_DIMMED_TEXT);
        label  = "No remote";
    }

    g.setColour (colour);
    g.drawCellText (label, col, row, width);
    ++row;
}

void Header::valueTreePropertyChanged (juce::ValueTree&, const juce::Identifier&)
{
    repaint();
}

} // namespace tit
