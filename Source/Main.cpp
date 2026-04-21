#include <JuceHeader.h>
#include "MainComponent.h"
#include "state/TitState.h"

class TitApp : public juce::JUCEApplication
{
public:
    TitApp() = default;

    const juce::String getApplicationName() override
    {
        return "titc";
    }

    const juce::String getApplicationVersion() override
    {
        return "0.0.0";
    }

    bool moreThanOneInstanceAllowed() override
    {
        return true;
    }

    void initialise (const juce::String&) override
    {
        // -------------------------------------------------------------------------
        // 1. TitState
        // -------------------------------------------------------------------------
        titState = std::make_unique<TitState>();

        // -------------------------------------------------------------------------
        // 2. MainComponent
        // -------------------------------------------------------------------------
        mainComponent = std::make_unique<MainComponent> (titState->getTree());

        // -------------------------------------------------------------------------
        // 3. ansi::Screen — Writer is a value member; Screen holds a reference
        // -------------------------------------------------------------------------
        screen = std::make_unique<jam::tui::Screen> (writer);

        // Size Screen and MainComponent to the current terminal dimensions
        const juce::Rectangle<int> termBounds { jam::tui::getBounds().toJuce() };
        screen->setBounds (termBounds);
        mainComponent->setBounds (termBounds.withPosition (0, 0));

        screen->addAndMakeVisible (*mainComponent);

        // -------------------------------------------------------------------------
        // 4. Input — route keys to MainComponent; route resize to Screen
        // -------------------------------------------------------------------------
        input = std::make_unique<jam::tui::Input>();

        input->start (
            [this] (jam::tui::KeyEvent event)
            {
                juce::MessageManager::callAsync ([this, event]
                {
                    mainComponent->handleInput (event);
                });
            },
            [this]
            {
                juce::MessageManager::callAsync ([this]
                {
                    const juce::Rectangle<int> newBounds { jam::tui::getBounds().toJuce() };
                    screen->setBounds (newBounds);
                    mainComponent->setBounds (newBounds.withPosition (0, 0));
                    screen->onTerminalResized();
                });
            }
        );

        // -------------------------------------------------------------------------
        // 5. Start render loop and state flush timer
        // -------------------------------------------------------------------------
        screen->start();
        titState->start();
    }

    void shutdown() override
    {
        // Stop in reverse construction order (BLESSED-B).
        // Input thread first so no callbacks fire during teardown.
        if (input != nullptr)
            input->stop();

        if (titState != nullptr)
            titState->stop();

        // Unique_ptr destructors handle MainComponent, TitState, Screen.
        input.reset();
        screen.reset();
        mainComponent.reset();
        titState.reset();
    }

    void anotherInstanceStarted (const juce::String&) override
    {
    }

private:
    jam::tui::Writer                    writer;
    std::unique_ptr<TitState>           titState;
    std::unique_ptr<MainComponent>      mainComponent;
    std::unique_ptr<jam::tui::Screen>        screen;
    std::unique_ptr<jam::tui::Input>         input;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TitApp)
};

START_JUCE_APPLICATION (TitApp)
