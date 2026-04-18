#pragma once

#include <JuceHeader.h>
#include "state/TitState.h"
#include "view/TitScreen.h"

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

private:
    jam::tui::Writer                        writer;
    juce::ValueTree                           themeTree;
    std::unique_ptr<jam::tui::ThemeResolver> themeResolver;
    std::unique_ptr<tit::TitState>            titState;
    std::unique_ptr<tit::TitScreen>           titScreen;
    std::unique_ptr<jam::tui::Screen>       screen;
    std::unique_ptr<jam::tui::Input>        input;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TitApp)
};
