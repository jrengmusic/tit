#include <JuceHeader.h>
#include "SetupWizardView.h"

// ---- Fallback colours -------------------------------------------------------
static const juce::Colour FALLBACK_CONTENT_TEXT { juce::Colour { 0xffcccccc } };
static const juce::Colour FALLBACK_LABEL_TEXT   { juce::Colour { 0xffffffff } };
static const juce::Colour FALLBACK_ACCENT_TEXT  { juce::Colour { 0xff44ffcc } };
static const juce::Colour FALLBACK_DIMMED_TEXT  { juce::Colour { 0xff808080 } };

namespace tit
{

// ============================================================================
// Phase painters (file-scope, no anonymous namespace)
// ============================================================================

static void paintEnvCheck (jam::tui::Graphics& g,
                            const juce::ValueTree& setupTree,
                            jam::tui::ThemeResolver& resolver,
                            int width, int row)
{
    const juce::Colour colour { resolver.getColour (tit::ID::labelTextColor,
                                                     FALLBACK_LABEL_TEXT) };
    g.setColour (colour);
    g.drawCellText ("Checking environment...", 0, row, width);
    (void) setupTree;
}

static void paintSSHKeyEntry (jam::tui::Graphics& g,
                               const juce::ValueTree& setupTree,
                               jam::tui::ThemeResolver& resolver,
                               int width, int row)
{
    const juce::Colour labelColour   { resolver.getColour (tit::ID::labelTextColor,
                                                            FALLBACK_LABEL_TEXT) };
    const juce::Colour contentColour { resolver.getColour (tit::ID::contentTextColor,
                                                            FALLBACK_CONTENT_TEXT) };
    const juce::String email { setupTree.getProperty (tit::ID::email).toString() };

    g.setColour (labelColour);
    g.drawCellText ("SSH Key Setup", 0, row, width);
    g.setColour (contentColour);
    g.drawCellText ("Email: " + email, 0, row + 1, width);
}

static void paintKeyGen (jam::tui::Graphics& g,
                          const juce::ValueTree& setupTree,
                          jam::tui::ThemeResolver& resolver,
                          int width, int row)
{
    const juce::Colour colour { resolver.getColour (tit::ID::accentTextColor,
                                                     FALLBACK_ACCENT_TEXT) };
    g.setColour (colour);
    g.drawCellText ("Generating SSH key...", 0, row, width);
    (void) setupTree;
}

static void paintDisplay (jam::tui::Graphics& g,
                           const juce::ValueTree& setupTree,
                           jam::tui::ThemeResolver& resolver,
                           int width, int row)
{
    const juce::Colour labelColour   { resolver.getColour (tit::ID::labelTextColor,
                                                            FALLBACK_LABEL_TEXT) };
    const juce::Colour accentColour  { resolver.getColour (tit::ID::accentTextColor,
                                                            FALLBACK_ACCENT_TEXT) };
    const juce::String publicKey { setupTree.getProperty (tit::ID::publicKey).toString() };

    g.setColour (labelColour);
    g.drawCellText ("Add this key to GitHub / GitLab / Gitea:", 0, row, width);
    g.setColour (accentColour);
    g.drawCellText (publicKey, 0, row + 1, width);
}

static void paintGitConfig (jam::tui::Graphics& g,
                              const juce::ValueTree& setupTree,
                              jam::tui::ThemeResolver& resolver,
                              int width, int row)
{
    const juce::Colour labelColour   { resolver.getColour (tit::ID::labelTextColor,
                                                            FALLBACK_LABEL_TEXT) };
    const juce::Colour contentColour { resolver.getColour (tit::ID::contentTextColor,
                                                            FALLBACK_CONTENT_TEXT) };
    const juce::String email { setupTree.getProperty (tit::ID::email).toString() };

    g.setColour (labelColour);
    g.drawCellText ("Git Configuration", 0, row, width);
    g.setColour (contentColour);
    g.drawCellText ("Email: " + email, 0, row + 1, width);
}

static void paintDone (jam::tui::Graphics& g,
                        const juce::ValueTree& setupTree,
                        jam::tui::ThemeResolver& resolver,
                        int width, int row)
{
    const juce::Colour colour { resolver.getColour (tit::ID::accentTextColor,
                                                     FALLBACK_ACCENT_TEXT) };
    g.setColour (colour);
    g.drawCellText ("Setup complete!", 0, row, width);
    (void) setupTree;
}

// ============================================================================
// Phase table
// ============================================================================

const SetupWizardView::PhasePainterMap& SetupWizardView::phaseTable() noexcept
{
    static const PhasePainterMap table
    {
        { SetupPhase::EnvCheck,    paintEnvCheck    },
        { SetupPhase::SSHKeyEntry, paintSSHKeyEntry },
        { SetupPhase::KeyGen,      paintKeyGen      },
        { SetupPhase::Display,     paintDisplay     },
        { SetupPhase::GitConfig,   paintGitConfig   },
        { SetupPhase::Done,        paintDone        },
    };
    return table;
}

// ============================================================================
// Construction / Destruction
// ============================================================================

SetupWizardView::SetupWizardView (const juce::ValueTree& stateTree,
                                   jam::tui::ThemeResolver& resolver)
    : themeResolver { resolver }
{
    setupTree = stateTree.getChildWithName (ID::SETUP);
    setupTree.addListener (this);
}

SetupWizardView::~SetupWizardView()
{
    setupTree.removeListener (this);
}

// ============================================================================
// Component overrides
// ============================================================================

void SetupWizardView::paint (jam::tui::Graphics& g)
{
    const juce::String phaseStr { setupTree.getProperty (ID::phase).toString() };
    const bool         hasPhase { phaseStr.isNotEmpty() };

    const SetupPhase phase { hasPhase ? parseSetupPhase (phaseStr) : SetupPhase::EnvCheck };

    const PhasePainterMap& table { phaseTable() };
    const auto             it    { table.find (phase) };

    const bool knownPhase { it != table.end() };

    if (knownPhase)
    {
        const int width { getWidth() };
        it->second (g, setupTree, themeResolver, width, 0);
    }
}

void SetupWizardView::handleInput (const jam::tui::KeyEvent&)
{
    // Setup wizard has no keyboard interaction at Step 3.6; keys handled by parent.
}

// ============================================================================
// ValueTree::Listener
// ============================================================================

void SetupWizardView::valueTreePropertyChanged (juce::ValueTree& tree,
                                                 const juce::Identifier& property)
{
    const bool isSetupProp { tree.getType() == ID::SETUP
                             and (property == ID::phase
                                  or property == ID::email
                                  or property == ID::publicKey) };

    if (isSetupProp)
        repaint();
}

} // namespace tit
