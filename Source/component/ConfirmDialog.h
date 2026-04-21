#pragma once
#include <JuceHeader.h>

class ConfirmDialog : public jam::tui::Component
{
public:
    ConfirmDialog() = default;
    ~ConfirmDialog() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConfirmDialog)
};
