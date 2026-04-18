#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// Header
// ============================================================================
//
// Renders current branch, working-tree status, and timeline (ahead/behind).
// Attaches ValueTree::Listener to the REPO subtree; redraws on any property
// change.  Ported from ___legacy___/internal/ui/header.go RenderHeaderInfo().
//
// VT surface: ID::REPO (branch, workingTree, timeline, aheadCount, behindCount,
//             cwd, remote, operation).
// BLESSED E: no Source/git/ imports; reads VT only.
// BLESSED S: no persistent state beyond the VT reference.

class Header : public jam::tui::Component,
               private juce::ValueTree::Listener
{
public:
    Header (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~Header() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint   (jam::tui::Graphics& g) override;
    void resized ()                        override;

private:
    juce::ValueTree           repoTree;
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
    // Paint helpers
    // -------------------------------------------------------------------------
    void paintCwdLine      (jam::tui::Graphics& g, int col, int& row, int width) const;
    void paintBranchLine   (jam::tui::Graphics& g, int col, int& row, int width) const;
    void paintStatusLine   (jam::tui::Graphics& g, int col, int& row, int width) const;
    void paintTimelineLine (jam::tui::Graphics& g, int col, int& row, int width) const;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Header)
};

} // namespace tit
