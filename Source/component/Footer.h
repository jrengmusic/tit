#pragma once
#include <JuceHeader.h>

class Footer : public jam::tui::Component
{
public:
    Footer() = default;
    ~Footer() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Footer)
};
