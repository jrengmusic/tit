#pragma once
#include <JuceHeader.h>

class FileHistory : public jam::tui::Component
{
public:
    FileHistory() = default;
    ~FileHistory() override = default;

    void paint (jam::tui::Graphics& g) override;

private:
    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (FileHistory)
};
