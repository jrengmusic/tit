#pragma once
#include <JuceHeader.h>
#include <unordered_map>
#include <functional>
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// ConfirmDialog
// ============================================================================
//
// Wraps jam::tui::Dialog and drives it from the DIALOG VT subtree.
//
// Variant dispatch:
//   Reads ID::kind from the DIALOG subtree.  A static lookup map keyed on
//   kind string (7 variants per SPEC §12) builds the DialogConfig.  The
//   inner Dialog primitive observes the same subtree for property changes.
//
// 7 variants (SPEC §12, Go confirm_dialog.go ConfirmationType constants):
//   rewind, time-travel, dirty, merge, push, branch, time-travel-return
//
// VT surface: ID::DIALOG (title, explanation, yesLabel, noLabel, actionId,
//             kind).
// BLESSED E: no Source/git/ imports.
// BLESSED S: no shadow DialogConfig — Dialog primitive is the SSOT.

class ConfirmDialog : public jam::tui::Component,
                      private juce::ValueTree::Listener
{
public:
    explicit ConfirmDialog (const juce::ValueTree& stateTree);
    ~ConfirmDialog() override;

    // -------------------------------------------------------------------------
    // Callbacks — forwarded from Dialog primitive
    // -------------------------------------------------------------------------
    std::function<void (const juce::String& actionId)> onConfirmed;
    std::function<void ()>                             onCancelled;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    bool isFocusable ()                             const override { return true; }
    void resized ()                                        override;

private:
    // -------------------------------------------------------------------------
    // Variant builder type
    // -------------------------------------------------------------------------
    struct VariantConfig
    {
        juce::String title;
        juce::String explanation;
        juce::String yesLabel;
        juce::String noLabel;
        juce::String actionId;
    };

    using VariantMap = std::unordered_map<juce::String, VariantConfig>;

    // -------------------------------------------------------------------------
    // Static variant table
    // -------------------------------------------------------------------------
    static const VariantMap& variantTable() noexcept;

    // -------------------------------------------------------------------------
    // State
    // -------------------------------------------------------------------------
    juce::ValueTree    dialogTree;
    jam::tui::Dialog dialog;

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
    void applyVariant (const juce::String& kind);

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (ConfirmDialog)
};

} // namespace tit
