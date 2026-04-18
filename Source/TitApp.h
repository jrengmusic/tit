#pragma once

#include <juce_gui_basics/juce_gui_basics.h>

class TitApp : public juce::JUCEApplication
{
public:
    TitApp() = default;

    const juce::String getApplicationName() override;
    const juce::String getApplicationVersion() override;
    bool moreThanOneInstanceAllowed() override;

    void initialise (const juce::String& commandLine) override;
    void shutdown() override;
    void anotherInstanceStarted (const juce::String& commandLine) override;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TitApp)
};
