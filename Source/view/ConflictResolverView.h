#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// ConflictResolverView
// ============================================================================
//
// 3-pane merge-conflict resolver.  Top row: ours (left) + theirs (right) via
// SplitPane.  Bottom row: merged TextPane.  Ported from
// ___legacy___/internal/ui/conflictresolver.go RenderConflictResolveGeneric.
//
// Listener surface:
//   - ID::DIFF subtree — rebuilds conflict regions on lines property change.
//
// Keyboard:
//   Tab       — cycle focus: ours → theirs → merged → ours.
//   Enter     — accept focused pane's content for selected hunk.
//   Left/Right— cycle which side is accepted for the focused pane.
//
// Pane border colours per ID::conflictPaneFocusedBorder /
// ID::conflictPaneUnfocusedBorder.
//
// BLESSED E: no Source/git/ imports; reads VT only.
// BLESSED S: no shadow conflict state — DIFF subtree is the SSOT.

class ConflictResolverView : public jam::tui::Component,
                             private juce::ValueTree::Listener
{
public:
    ConflictResolverView (const juce::ValueTree& stateTree,
                          jam::tui::ThemeResolver& resolver);
    ~ConflictResolverView() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    bool isFocusable ()                             const override { return true; }
    void resized ()                                        override;

private:
    // -------------------------------------------------------------------------
    // Pane focus index constants
    // -------------------------------------------------------------------------
    static constexpr int PANE_OURS   { 0 };
    static constexpr int PANE_THEIRS { 1 };
    static constexpr int PANE_MERGED { 2 };
    static constexpr int PANE_COUNT  { 3 };

    // -------------------------------------------------------------------------
    // Layout constants (ported from Go topRowHeight = totalHeight / 3)
    // -------------------------------------------------------------------------
    static constexpr int TOP_ROW_FRACTION_DENOM { 3 };

    // -------------------------------------------------------------------------
    // State
    // -------------------------------------------------------------------------
    juce::ValueTree            diffTree;
    jam::tui::ThemeResolver& themeResolver;

    jam::tui::TextPane  oursPane;
    jam::tui::TextPane  theirsPane;
    jam::tui::SplitPane topSplit;
    jam::tui::TextPane  mergedPane;

    int focusedPaneIndex { PANE_OURS };

    // -------------------------------------------------------------------------
    // ValueTree::Listener overrides
    // -------------------------------------------------------------------------
    void valueTreePropertyChanged  (juce::ValueTree& tree,
                                    const juce::Identifier& property)           override;
    void valueTreeChildAdded       (juce::ValueTree&, juce::ValueTree&)         override {}
    void valueTreeChildRemoved     (juce::ValueTree&, juce::ValueTree&, int)    override {}
    void valueTreeChildOrderChanged(juce::ValueTree&, int, int)                 override {}
    void valueTreeParentChanged    (juce::ValueTree&)                           override {}

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------
    void rebuildPanes ();
    void cycleFocus   ();
    void updateFocusStyles ();

    jam::tui::TextPane& paneForIndex (int index) noexcept;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConflictResolverView)
};

} // namespace tit
