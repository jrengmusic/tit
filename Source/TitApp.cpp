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
    quit();
}

void TitApp::shutdown()
{
}

void TitApp::anotherInstanceStarted (const juce::String&)
{
}
