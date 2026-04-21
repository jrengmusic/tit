#pragma once
#include <JuceHeader.h>

class Console : public jam::tui::Component
{
public:
    Console() = default;
    ~Console() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Console)
};
