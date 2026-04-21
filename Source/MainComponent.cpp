#include <JuceHeader.h>
#include "MainComponent.h"

MainComponent::MainComponent (juce::ValueTree stateTree)
    : stateTree { stateTree }
{
    setLookAndFeel (&lookAndFeel);
}

void MainComponent::paint (jam::tui::Graphics& g)
{
}

void MainComponent::handleInput (const jam::tui::KeyEvent& event)
{
}

void MainComponent::resized()
{
}
