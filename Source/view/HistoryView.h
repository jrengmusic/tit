#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// HistoryView
// ============================================================================
//
// 2-pane commit log browser.  Left pane: scrollable commit list.
// Right pane: details for selected commit.  Ported from
// ___legacy___/internal/ui/history.go RenderHistorySplitPane.
//
// Listener surface:
//   - ID::HISTORY subtree — rebuilds list on COMMIT child add/remove.
//   - ID::SELECTION subtree — syncs selected row on historyIndex change.
//
// On Enter (when commit list has focus): writes ID::historyIndex to
// ID::SELECTION.  Git invocation deferred to Sprint 4.
//
// BLESSED E: no Source/git/ imports; reads VT only.
// BLESSED S: no shadow commit list — HISTORY subtree is the SSOT.

class HistoryView : public jam::tui::Component,
                    private juce::ValueTree::Listener
{
public:
    HistoryView (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~HistoryView() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    bool isFocusable ()                             const override { return true; }
    void resized ()                                        override;

private:
    // -------------------------------------------------------------------------
    // Layout constant (ported from Go CommitListPaneWidth)
    // -------------------------------------------------------------------------
    static constexpr int COMMIT_LIST_PANE_WIDTH { 26 };

    // -------------------------------------------------------------------------
    // State
    // -------------------------------------------------------------------------
    juce::ValueTree            historyTree;
    juce::ValueTree            selectionTree;
    jam::tui::ThemeResolver& themeResolver;

    jam::tui::ListPane  commitPane;
    jam::tui::TextPane  detailPane;
    jam::tui::SplitPane splitPane;

    // -------------------------------------------------------------------------
    // ValueTree::Listener overrides
    // -------------------------------------------------------------------------
    void valueTreePropertyChanged  (juce::ValueTree& tree,
                                    const juce::Identifier& property)           override;
    void valueTreeChildAdded       (juce::ValueTree& parent,
                                    juce::ValueTree& child)                     override;
    void valueTreeChildRemoved     (juce::ValueTree& parent,
                                    juce::ValueTree& child, int)                override;
    void valueTreeChildOrderChanged(juce::ValueTree&, int, int)                 override {}
    void valueTreeParentChanged    (juce::ValueTree&)                           override {}

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------
    void rebuildList   ();
    void syncSelection (int index);
    void updateDetail  (int index);

    juce::Array<jam::tui::ListItem> buildItems () const;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (HistoryView)
};

} // namespace tit
