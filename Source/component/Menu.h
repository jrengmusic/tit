#pragma once
#include <JuceHeader.h>

class Menu : public jam::tui::Component
{
public:
    Menu() = default;
    ~Menu() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Menu)
};
