#pragma once
#include <JuceHeader.h>

class History : public jam::tui::Component
{
public:
    History() = default;
    ~History() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (History)
};
