#pragma once
#include <JuceHeader.h>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// FileHistoryView
// ============================================================================
//
// 3-pane file-scoped history browser.  Top row: commit list (left) + file
// list (right) via SplitPane.  Bottom row: diff TextPane.  Ported from
// ___legacy___/internal/ui/filehistory.go RenderFileHistorySplitPane.
//
// Listener surface:
//   - ID::HISTORY subtree — rebuilds commit list on child changes.
//   - ID::FILES   subtree — rebuilds file list on child changes.
//   - ID::DIFF    subtree — refreshes diff pane on lines property change.
//
// BLESSED E: no Source/git/ imports; reads VT only.
// BLESSED S: no shadow state beyond selection indices.

class FileHistoryView : public jam::tui::Component,
                        private juce::ValueTree::Listener
{
public:
    FileHistoryView (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~FileHistoryView() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    bool isFocusable ()                             const override { return true; }
    void resized ()                                        override;

private:
    // -------------------------------------------------------------------------
    // Layout constants (ported from Go CommitListPaneWidth / topRowHeight ratio)
    // -------------------------------------------------------------------------
    static constexpr int COMMIT_LIST_PANE_WIDTH { 26 };
    static constexpr int TOP_ROW_FRACTION_DENOM { 3 };

    // -------------------------------------------------------------------------
    // Focused pane index constants (matches Go FileHistoryPane enum)
    // -------------------------------------------------------------------------
    static constexpr int PANE_COMMITS { 0 };
    static constexpr int PANE_FILES   { 1 };
    static constexpr int PANE_DIFF    { 2 };

    // -------------------------------------------------------------------------
    // State
    // -------------------------------------------------------------------------
    juce::ValueTree            historyTree;
    juce::ValueTree            filesTree;
    juce::ValueTree            diffTree;
    jam::tui::ThemeResolver& themeResolver;

    jam::tui::ListPane  commitPane;
    jam::tui::ListPane  filePane;
    jam::tui::SplitPane topSplit;
    jam::tui::TextPane  diffPane;

    int focusedPaneIndex { PANE_COMMITS };

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
    void rebuildCommits ();
    void rebuildFiles   ();
    void refreshDiff    ();
    void cycleFocus     ();

    juce::Array<jam::tui::ListItem> buildCommitItems () const;
    juce::Array<jam::tui::ListItem> buildFileItems   () const;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (FileHistoryView)
};

} // namespace tit
