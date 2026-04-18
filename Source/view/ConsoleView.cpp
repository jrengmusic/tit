#include <JuceHeader.h>
#include "ConsoleView.h"

namespace tit
{

ConsoleView::ConsoleView (const juce::ValueTree& stateTree,
                           jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
    , consoleStream { stateTree.getChildWithName (ID::CONSOLE) }
{
    addAndMakeVisible (consoleStream);
}

void ConsoleView::resized()
{
    consoleStream.setBounds (getLocalBounds());
}

void ConsoleView::paint (jam::tui::Graphics&)
{
    // ConsoleStream paints itself.
}

void ConsoleView::handleInput (const jam::tui::KeyEvent& event)
{
    consoleStream.handleInput (event);
}

} // namespace tit
