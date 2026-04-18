#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// ConsoleView
// ============================================================================
//
// Wraps jam::tui::ConsoleStream and drives it from the CONSOLE VT subtree.
//
// Listener surface:
//   - ID::CONSOLE subtree — LINE children stream in via valueTreeChildAdded.
//     ConsoleStream's own listener handles VT-driven line appends.
//
// BLESSED E: no Source/git/ imports; reads VT only via ConsoleStream.
// BLESSED S: no shadow line buffer — ConsoleStream is the SSOT.

class ConsoleView : public jam::tui::Component
{
public:
    ConsoleView (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~ConsoleView() override = default;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    bool isFocusable ()                             const override { return true; }
    void resized ()                                        override;

private:
    jam::tui::ThemeResolver& themeResolver;
    jam::tui::ConsoleStream  consoleStream;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConsoleView)
};

} // namespace tit
