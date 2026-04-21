#pragma once
#include <JuceHeader.h>

// ============================================================================
// MainComponent
// ============================================================================

class MainComponent : public jam::tui::Component
{
public:
    explicit MainComponent (juce::ValueTree stateTree);
    ~MainComponent() override = default;

    void paint (jam::tui::Graphics& g) override;
    void handleInput (const jam::tui::KeyEvent& event) override;
    void resized() override;

private:
    juce::ValueTree stateTree;
    jam::tui::LookAndFeel lookAndFeel;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (MainComponent)
};
