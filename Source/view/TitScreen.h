#pragma once
#include <JuceHeader.h>
#include "Banner.h"
#include "Header.h"
#include "Footer.h"
#include "MenuView.h"

namespace tit
{

// ============================================================================
// TitScreen
// ============================================================================
//
// Root composition of the four shell views.
//
// Layout per SPEC §14:
//   Banner     — top row (1 row tall, full width)
//   Header     — below banner (4 rows, full width)
//   MenuView   — content area (fills remaining rows above footer)
//   Footer     — bottom row (1 row tall, full width)
//
// Holds const reference to the VT root; passes it to children.
// Zero VT listener — composition only.
// BLESSED E: no Source/git/ imports.

class TitScreen : public jam::tui::Component
{
public:
    TitScreen (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~TitScreen() override = default;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    void resized ()                                        override;

    // -------------------------------------------------------------------------
    // Item-selected callback (forwarded from MenuView)
    // -------------------------------------------------------------------------
    std::function<void (const jam::tui::MenuItem&)> onItemSelected;

private:
    // Row heights per SPEC §14 layout.
    static constexpr int BANNER_HEIGHT { 1 };
    static constexpr int HEADER_HEIGHT { 4 };
    static constexpr int FOOTER_HEIGHT { 1 };

    std::unique_ptr<Banner>   banner;
    std::unique_ptr<Header>   header;
    std::unique_ptr<MenuView> menuView;
    std::unique_ptr<Footer>   footer;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TitScreen)
};

} // namespace tit
