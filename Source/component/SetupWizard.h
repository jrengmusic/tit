#pragma once
#include <JuceHeader.h>

class SetupWizard : public jam::tui::Component
{
public:
    SetupWizard() = default;
    ~SetupWizard() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (SetupWizard)
};
