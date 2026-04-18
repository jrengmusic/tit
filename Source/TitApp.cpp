#include "TitApp.h"

const juce::String TitApp::getApplicationName()
{
    return "titc";
}

const juce::String TitApp::getApplicationVersion()
{
    return "0.0.0";
}

bool TitApp::moreThanOneInstanceAllowed()
{
    return true;
}

void TitApp::initialise (const juce::String&)
{
    // -------------------------------------------------------------------------
    // 1. TitState
    // -------------------------------------------------------------------------
    titState = std::make_unique<tit::TitState>();

    // -------------------------------------------------------------------------
    // 2. Theme — attempt file load, fall back to minimal hardcoded VT
    // -------------------------------------------------------------------------
    const juce::File themeFile { juce::File::getSpecialLocation (
        juce::File::currentExecutableFile).getSiblingFile ("themes/default.xml") };

    if (themeFile.existsAsFile())
    {
        if (auto xml = juce::XmlDocument::parse (themeFile))
            themeTree = juce::ValueTree::fromXml (*xml);
    }

    if (not themeTree.isValid())
    {
        themeTree = juce::ValueTree { "THEME" };
        themeTree.setProperty ("name", "fallback", nullptr);
    }

    // -------------------------------------------------------------------------
    // 3. ThemeResolver — pass LOOK_AND_FEEL subtree when present, else root
    // -------------------------------------------------------------------------
    juce::ValueTree themeSubtree { themeTree.getChildWithName ("LOOK_AND_FEEL") };

    if (not themeSubtree.isValid())
        themeSubtree = themeTree;

    themeResolver = std::make_unique<jam::tui::ThemeResolver> (themeSubtree);

    // -------------------------------------------------------------------------
    // 4. TitScreen
    // -------------------------------------------------------------------------
    titScreen = std::make_unique<tit::TitScreen> (titState->getTree(), *themeResolver);

    // -------------------------------------------------------------------------
    // 5. ansi::Screen — Writer is a value member; Screen holds a reference
    // -------------------------------------------------------------------------
    screen = std::make_unique<jam::tui::Screen> (writer);

    // Size Screen and TitScreen to the current terminal dimensions
    const juce::Rectangle<int> termBounds { jam::tui::getBounds().toJuce() };
    screen->setBounds (termBounds);
    titScreen->setBounds (termBounds.withPosition (0, 0));

    screen->addAndMakeVisible (*titScreen);

    // -------------------------------------------------------------------------
    // 6. Input — route keys to TitScreen; route resize to Screen
    // -------------------------------------------------------------------------
    input = std::make_unique<jam::tui::Input>();

    input->start (
        [this] (jam::tui::KeyEvent event)
        {
            juce::MessageManager::callAsync ([this, event]
            {
                titScreen->handleInput (event);
            });
        },
        [this]
        {
            juce::MessageManager::callAsync ([this]
            {
                const juce::Rectangle<int> newBounds { jam::tui::getBounds().toJuce() };
                screen->setBounds (newBounds);
                titScreen->setBounds (newBounds.withPosition (0, 0));
                screen->onTerminalResized();
            });
        }
    );

    // -------------------------------------------------------------------------
    // 7. Start render loop and state flush timer
    // -------------------------------------------------------------------------
    screen->start();
    titState->start();
}

void TitApp::shutdown()
{
    // Stop in reverse construction order (BLESSED-B).
    // Input thread first so no callbacks fire during teardown.
    if (input != nullptr)
        input->stop();

    if (titState != nullptr)
        titState->stop();

    // Unique_ptr destructors handle TitScreen, ThemeResolver, TitState, Screen.
    input.reset();
    screen.reset();
    titScreen.reset();
    themeResolver.reset();
    titState.reset();
}

void TitApp::anotherInstanceStarted (const juce::String&)
{
}
