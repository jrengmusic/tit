#include <JuceHeader.h>
#include "ConfirmDialog.h"

namespace tit
{

// ============================================================================
// Static variant table
// ============================================================================

const ConfirmDialog::VariantMap& ConfirmDialog::variantTable() noexcept
{
    static const VariantMap table
    {
        {
            "rewind",
            {
                "Destructive Operation",
                "This will discard all commits after the selected commit. "
                "Any uncommitted changes will be lost.",
                "Rewind",
                "Cancel",
                "rewind"
            }
        },
        {
            "time-travel",
            {
                "Entering Time Travel Mode",
                "You are about to view a past commit in read-only mode. "
                "You can explore code, build, and test. You cannot commit. "
                "To keep changes, merge them back to your branch.",
                "Continue",
                "Cancel",
                "time_travel"
            }
        },
        {
            "dirty",
            {
                "Uncommitted Changes",
                "You have uncommitted changes. To proceed, your changes will "
                "be temporarily saved (stashed). After the operation completes "
                "they will be reapplied. This may cause conflicts.",
                "Save and proceed",
                "Cancel",
                "dirty"
            }
        },
        {
            "merge",
            {
                "Merge Branch",
                "This will merge the selected branch into the current branch. "
                "Conflicts will be handled if they occur.",
                "Merge",
                "Cancel",
                "merge"
            }
        },
        {
            "push",
            {
                "Force Push",
                "This will force push to remote. Any remote changes will be "
                "overwritten. This cannot be undone.",
                "Force push",
                "Cancel",
                "push"
            }
        },
        {
            "branch",
            {
                "Switch Branch",
                "Switch to the selected branch. Your working tree must be "
                "clean before switching.",
                "Switch",
                "Cancel",
                "branch"
            }
        },
        {
            "time-travel-return",
            {
                "Return to Branch",
                "You have uncommitted changes in time travel mode. "
                "Returning to your branch will discard these changes.",
                "Discard and return",
                "Cancel",
                "time_travel_return"
            }
        }
    };
    return table;
}

// ============================================================================
// Construction / Destruction
// ============================================================================

ConfirmDialog::ConfirmDialog (const juce::ValueTree& stateTree)
    : dialog { stateTree.getChildWithName (ID::DIALOG) }
{
    dialogTree = stateTree.getChildWithName (ID::DIALOG);
    dialogTree.addListener (this);

    dialog.onConfirmed = [this] (const juce::String& actionId)
    {
        if (onConfirmed)
            onConfirmed (actionId);
    };

    dialog.onCancelled = [this] ()
    {
        if (onCancelled)
            onCancelled ();
    };

    addAndMakeVisible (dialog);

    const juce::String kind { dialogTree.getProperty (ID::kind).toString() };
    applyVariant (kind);
}

ConfirmDialog::~ConfirmDialog()
{
    dialogTree.removeListener (this);
}

// ============================================================================
// Component overrides
// ============================================================================

void ConfirmDialog::resized()
{
    dialog.setBounds (getLocalBounds());
}

void ConfirmDialog::paint (jam::tui::Graphics&)
{
    // Dialog primitive paints itself.
}

void ConfirmDialog::handleInput (const jam::tui::KeyEvent& event)
{
    dialog.handleInput (event);
}

// ============================================================================
// ValueTree::Listener
// ============================================================================

void ConfirmDialog::valueTreePropertyChanged (juce::ValueTree& tree,
                                               const juce::Identifier& property)
{
    const bool isKindChange { tree.getType() == ID::DIALOG
                              and property == ID::kind };

    if (isKindChange)
    {
        const juce::String kind { dialogTree.getProperty (ID::kind).toString() };
        applyVariant (kind);
    }
}

// ============================================================================
// Helpers
// ============================================================================

void ConfirmDialog::applyVariant (const juce::String& kind)
{
    const VariantMap& table { variantTable() };
    const auto        it    { table.find (kind) };

    const bool knownVariant { it != table.end() };

    if (knownVariant)
    {
        const VariantConfig& cfg { it->second };

        dialogTree.setProperty (jam::tui::Dialog::PropTitle,       cfg.title,       nullptr);
        dialogTree.setProperty (jam::tui::Dialog::PropExplanation, cfg.explanation, nullptr);
        dialogTree.setProperty (jam::tui::Dialog::PropYesLabel,    cfg.yesLabel,    nullptr);
        dialogTree.setProperty (jam::tui::Dialog::PropNoLabel,     cfg.noLabel,     nullptr);
        dialogTree.setProperty (jam::tui::Dialog::PropActionId,    cfg.actionId,    nullptr);
    }
}

} // namespace tit
