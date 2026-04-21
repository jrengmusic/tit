#pragma once
#include <JuceHeader.h>

class ConflictResolver : public jam::tui::Component
{
public:
    ConflictResolver() = default;
    ~ConflictResolver() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConflictResolver)
};
