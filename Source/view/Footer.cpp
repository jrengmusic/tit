#include <JuceHeader.h>
#include "Footer.h"
#include "state/TitAxis.h"

// ---- Fallback colours — ARGB uint32 via explicit constructor (unambiguous) --
static const juce::Colour FALLBACK_FOOTER_TEXT  { juce::Colour { 0xff808080 } };
static const juce::Colour FALLBACK_ACCENT_TEXT  { juce::Colour { 0xff5fd7ff } };
static const juce::Colour FALLBACK_CONTENT_TEXT { juce::Colour { 0xffffffff } };

// ---- Key-hint separator ----------------------------------------------------
static const juce::String HINT_SEP { "  |  " };

namespace tit
{

Footer::Footer (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
{
    repoTree = stateTree.getChildWithName (ID::REPO);
    repoTree.addListener (this);
}

Footer::~Footer()
{
    repoTree.removeListener (this);
}

void Footer::resized()
{
    repaint();
}

void Footer::paint (jam::tui::Graphics& g)
{
    paintHints (g, getWidth());
}

void Footer::paintHints (jam::tui::Graphics& g, int width) const
{
    const juce::String opStr { repoTree.getProperty (ID::operation).toString() };
    const Operation    op    { parseOperation (opStr) };

    // Build operation-dependent hint string per SPEC §14.1 / §14.2.
    juce::String hints;

    if (op == Operation::Normal
        or op == Operation::DirtyOperation
        or op == Operation::Rewinding)
    {
        hints = "j/k navigate" + HINT_SEP + "Enter select" + HINT_SEP + "Ctrl+C exit";
    }
    else if (op == Operation::TimeTraveling)
    {
        hints = "j/k navigate" + HINT_SEP + "Enter select" + HINT_SEP + "Esc return" + HINT_SEP + "Ctrl+C exit";
    }
    else if (op == Operation::Merging or op == Operation::Rebasing
             or op == Operation::Conflicted)
    {
        hints = "j/k navigate" + HINT_SEP + "Enter select" + HINT_SEP + "Ctrl+C exit";
    }
    else if (op == Operation::NotRepo)
    {
        hints = "j/k navigate" + HINT_SEP + "Enter select" + HINT_SEP + "Ctrl+C exit";
    }
    else
    {
        hints = "Ctrl+C exit";
    }

    const juce::Colour footerColour { themeResolver.getColour (ID::footerTextColor,
                                                                FALLBACK_FOOTER_TEXT) };
    g.setColour (footerColour);
    g.drawCellText (hints, 0, 0, width);
}

void Footer::valueTreePropertyChanged (juce::ValueTree&, const juce::Identifier&)
{
    repaint();
}

} // namespace tit
