#pragma once
#include <JuceHeader.h>

class Banner : public jam::tui::Component
{
public:
    Banner() = default;
    ~Banner() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Banner)
};
