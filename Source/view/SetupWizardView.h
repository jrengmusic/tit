#pragma once
#include <JuceHeader.h>
#include <unordered_map>
#include <functional>
#include "state/TitAxis.h"
#include "TitIdentifier.h"

namespace tit
{

// ============================================================================
// SetupWizardView
// ============================================================================
//
// Multi-phase setup wizard per SPEC §11 + RFC §4.6 SETUP.
//
// Phases (SetupPhase enum from TitAxis.h):
//   EnvCheck    — checking git + SSH availability
//   SSHKeyEntry — prompt for email (key comment)
//   KeyGen      — key generation in progress
//   Display     — display public key for GitHub/GitLab/Gitea
//   GitConfig   — user.name / user.email entry
//   Done        — wizard complete
//
// Listener surface:
//   - ID::SETUP subtree — phase, email, publicKey properties.
//
// Phase dispatch via std::unordered_map<SetupPhase, PhasePainter> lookup
// (MANIFESTO L 3-branch mandate).
//
// BLESSED E: no Source/git/ imports.
// BLESSED S: no shadow phase state — SETUP subtree is the SSOT.

class SetupWizardView : public jam::tui::Component,
                        private juce::ValueTree::Listener
{
public:
    SetupWizardView (const juce::ValueTree& stateTree, jam::tui::ThemeResolver& resolver);
    ~SetupWizardView() override;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (jam::tui::Graphics& g)            override;
    void handleInput (const jam::tui::KeyEvent& event)  override;
    bool isFocusable ()                             const override { return false; }
    void resized ()                                        override {}

private:
    // -------------------------------------------------------------------------
    // Phase painter type
    // -------------------------------------------------------------------------
    using PhasePainter = std::function<void (jam::tui::Graphics&,
                                             const juce::ValueTree&,
                                             jam::tui::ThemeResolver&,
                                             int, int)>;

    using PhasePainterMap = std::unordered_map<SetupPhase, PhasePainter>;

    // -------------------------------------------------------------------------
    // Static phase painter table
    // -------------------------------------------------------------------------
    static const PhasePainterMap& phaseTable() noexcept;

    // -------------------------------------------------------------------------
    // State
    // -------------------------------------------------------------------------
    juce::ValueTree            setupTree;
    jam::tui::ThemeResolver& themeResolver;

    // -------------------------------------------------------------------------
    // ValueTree::Listener overrides
    // -------------------------------------------------------------------------
    void valueTreePropertyChanged  (juce::ValueTree& tree,
                                    const juce::Identifier& property)           override;
    void valueTreeChildAdded       (juce::ValueTree&, juce::ValueTree&)         override {}
    void valueTreeChildRemoved     (juce::ValueTree&, juce::ValueTree&, int)    override {}
    void valueTreeChildOrderChanged(juce::ValueTree&, int, int)                 override {}
    void valueTreeParentChanged    (juce::ValueTree&)                           override {}

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (SetupWizardView)
};

} // namespace tit
