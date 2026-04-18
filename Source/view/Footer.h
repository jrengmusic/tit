#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// Footer
// ============================================================================
//
// Renders contextual key hints per SPEC §14 (§14.1, §14.2) and §15.
// Attaches ValueTree::Listener to the REPO subtree; hint set is operation-
// dependent.  Ported from ___legacy___/internal/ui/footer.go.
//
// VT surface: ID::REPO (operation property).
// BLESSED E: no Source/git/ imports; reads VT only.
// BLESSED S: no persistent state beyond the VT reference.

class Footer : public jam::tui::Component,
               private juce::ValueTree::Listener
{
public:
    Footer (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~Footer() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint   (jam::tui::Graphics& g) override;
    void resized ()                        override;

private:
    juce::ValueTree            repoTree;
    jam::tui::ThemeResolver& themeResolver;

    // -------------------------------------------------------------------------
    // ValueTree::Listener overrides
    // -------------------------------------------------------------------------
    void valueTreePropertyChanged  (juce::ValueTree&, const juce::Identifier&) override;
    void valueTreeChildAdded       (juce::ValueTree&, juce::ValueTree&)        override {}
    void valueTreeChildRemoved     (juce::ValueTree&, juce::ValueTree&, int)   override {}
    void valueTreeChildOrderChanged(juce::ValueTree&, int, int)                override {}
    void valueTreeParentChanged    (juce::ValueTree&)                          override {}

    // -------------------------------------------------------------------------
    // Paint helper
    // -------------------------------------------------------------------------
    void paintHints (jam::tui::Graphics& g, int width) const;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Footer)
};

} // namespace tit
