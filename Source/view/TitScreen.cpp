#include <JuceHeader.h>
#include "TitScreen.h"

namespace tit
{

TitScreen::TitScreen (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver)
{
    banner   = std::make_unique<Banner>();
    header   = std::make_unique<Header>   (stateTree, resolver);
    menuView = std::make_unique<MenuView> (stateTree, resolver);
    footer   = std::make_unique<Footer>   (stateTree, resolver);

    menuView->onItemSelected = [this] (const jam::tui::MenuItem& item)
    {
        if (onItemSelected)
            onItemSelected (item);
    };

    addAndMakeVisible (*banner);
    addAndMakeVisible (*header);
    addAndMakeVisible (*menuView);
    addAndMakeVisible (*footer);
}

void TitScreen::resized()
{
    const int totalWidth  { getWidth() };
    const int totalHeight { getHeight() };

    int y { 0 };

    banner->setBounds   (0, y, totalWidth, BANNER_HEIGHT);
    y += BANNER_HEIGHT;

    header->setBounds   (0, y, totalWidth, HEADER_HEIGHT);
    y += HEADER_HEIGHT;

    const int contentHeight { totalHeight - y - FOOTER_HEIGHT };
    menuView->setBounds (0, y, totalWidth, juce::jmax (0, contentHeight));
    y += juce::jmax (0, contentHeight);

    footer->setBounds   (0, y, totalWidth, FOOTER_HEIGHT);
}

void TitScreen::paint (jam::tui::Graphics&)
{
    // Children paint themselves.
}

void TitScreen::handleInput (const jam::tui::KeyEvent& event)
{
    const bool isQuitKey { event.type == jam::tui::KeyType::CtrlC
                           or event.type == jam::tui::KeyType::CtrlD };

    if (isQuitKey)
        juce::JUCEApplication::getInstance()->systemRequestedQuit();
    else
        menuView->handleInput (event);
}

} // namespace tit
