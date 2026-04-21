#pragma once
#include <JuceHeader.h>

class Header : public jam::tui::Component
{
public:
    Header() = default;
    ~Header() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Header)
};
