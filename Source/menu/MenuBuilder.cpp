#include "MenuBuilder.h"
#include "TitIdentifier.h"

// ============================================================================
// File-scope helpers — one per Operation (>30 LOC generators extracted)
// ============================================================================
//
// Each function is a pure function of the REPO ValueTree subtree.
// No member access, no side effects, no early returns.

namespace tit::menu
{

// ---------------------------------------------------------------------------
// buildNotRepoMenu
// ---------------------------------------------------------------------------
// Go: menuNotRepo() → always [init, clone]
// SPEC §6 "When Operation = NotRepo": Initialize repository, Clone repository

static juce::Array<MenuItemDef> buildNotRepoMenu (const juce::ValueTree&)
{
    return { INIT, CLONE };
}

// ---------------------------------------------------------------------------
// buildTimeTravelingMenu
// ---------------------------------------------------------------------------
// Go: menuTimeTraveling()
// SPEC §6 "When Operation = TimeTraveling": history, file history, return.
// TIME_TRAVEL_MERGE is defined in MenuItems.h but intentionally not emitted here —
// matches Go menu_render_extra.go:menuTimeTraveling which emits only time_travel_return.
// The merge-back choice is dialog-driven (confirm_dialog.go ConfirmTimeTravelMerge),
// not a top-level menu entry. SPEC §6 was updated to match this shipping Go behavior.

static juce::Array<MenuItemDef> buildTimeTravelingMenu (const juce::ValueTree&)
{
    return { TIME_TRAVEL_HISTORY, TIME_TRAVEL_FILES_HISTORY, TIME_TRAVEL_RETURN };
}

// ---------------------------------------------------------------------------
// buildMergingMenu
// ---------------------------------------------------------------------------
// Go: menuMerging() → [finalize_merge, abort_merge]
// SPEC §5 "Merging: Offer commit (finalize merge) or abort"

static juce::Array<MenuItemDef> buildMergingMenu (const juce::ValueTree&)
{
    return { FINALIZE_MERGE, ABORT_MERGE };
}

// ---------------------------------------------------------------------------
// buildConflictedMenu
// ---------------------------------------------------------------------------
// Go: menuConflicted() → [abort_merge]
// SPEC §5 "Conflicted: open conflict resolver (resolve or abort)"

static juce::Array<MenuItemDef> buildConflictedMenu (const juce::ValueTree&)
{
    return { ABORT_MERGE };
}

// ---------------------------------------------------------------------------
// buildRebasingMenu
// ---------------------------------------------------------------------------
// Go: menuRebasing() → [rebase_continue, rebase_abort]
// SPEC §5 "Rebasing: offer continue or abort"

static juce::Array<MenuItemDef> buildRebasingMenu (const juce::ValueTree&)
{
    return { REBASE_CONTINUE, REBASE_ABORT };
}

// ---------------------------------------------------------------------------
// buildDirtyOperationMenu
// ---------------------------------------------------------------------------
// Go: menuDirtyOperation() → [abort_merge]
// SPEC §5 "DirtyOperation: resume dirty pipeline — TIT marker file tracks state"

static juce::Array<MenuItemDef> buildDirtyOperationMenu (const juce::ValueTree&)
{
    return { ABORT_MERGE };
}

// ---------------------------------------------------------------------------
// buildRewindingMenu
// ---------------------------------------------------------------------------
// Go: menuGenerators[Rewinding] = menuNormal (transient — by render time done)
// Mirrors Go: Rewinding uses the Normal generator (identical behaviour).
// Forward declaration resolved via lambda in constructor capturing buildNormalMenu.

// Declared after buildNormalMenu — see constructor.

// ---------------------------------------------------------------------------
// workingTreeItems — helpers for buildNormalMenu
// ---------------------------------------------------------------------------
//
// Go: menuWorkingTree()
// SPEC §6 Normal / Working Tree Actions
//   Clean  → no items
//   Dirty  → commit; commit_push (if HasRemote); reset_discard_changes

static juce::Array<MenuItemDef> workingTreeItems (const juce::ValueTree& repo)
{
    const juce::String wtString { repo.getProperty (ID::workingTree).toString() };
    const WorkingTree wt        { parseWorkingTree (wtString) };
    const juce::String rmString { repo.getProperty (ID::remote).toString() };
    const Remote rm             { parseRemote (rmString) };

    juce::Array<MenuItemDef> items;

    if (wt == WorkingTree::Dirty)
    {
        items.add (COMMIT);

        if (rm == Remote::HasRemote)
            items.add (COMMIT_PUSH);

        items.add (RESET_DISCARD_CHANGES);
    }

    return items;
}

// ---------------------------------------------------------------------------
// timelineItems — helpers for buildNormalMenu
// ---------------------------------------------------------------------------
//
// Go: menuTimeline()
// SPEC §6 Normal / Timeline Sync Actions

static juce::Array<MenuItemDef> timelineItems (const juce::ValueTree& repo)
{
    const juce::String rmString { repo.getProperty (ID::remote).toString() };
    const Remote rm             { parseRemote (rmString) };

    juce::Array<MenuItemDef> items;

    // Timeline operations only available when a remote exists
    // (add_remote appears at bottom of Normal menu separately)
    if (rm == Remote::HasRemote)
    {
        const juce::String tlString { repo.getProperty (ID::timeline).toString() };
        const Timeline tl           { parseTimeline (tlString) };
        const juce::String wtString { repo.getProperty (ID::workingTree).toString() };
        const WorkingTree wt        { parseWorkingTree (wtString) };

        if (tl == Timeline::InSync)
        {
            // No sync actions needed when in sync (SPEC §6 "When Timeline = InSync")
        }
        else if (tl == Timeline::Ahead)
        {
            // Push only if working tree is clean — cannot push uncommitted changes
            if (wt == WorkingTree::Clean)
            {
                items.add (PUSH);
                items.add (FORCE_PUSH);
            }
        }
        else if (tl == Timeline::Behind)
        {
            if (wt == WorkingTree::Dirty)
            {
                items.add (DIRTY_PULL_MERGE);
                items.add (REPLACE_LOCAL);
            }
            else
            {
                items.add (PULL_MERGE);
                items.add (REPLACE_LOCAL);
            }
        }
        else if (tl == Timeline::Diverged)
        {
            if (wt == WorkingTree::Dirty)
            {
                items.add (DIRTY_PULL_MERGE);
                items.add (FORCE_PUSH);
                items.add (REPLACE_LOCAL);
            }
            else
            {
                items.add (PUSH_AUTO_SYNC);
                items.add (PULL_MERGE_DIVERGED);
                items.add (FORCE_PUSH);
                items.add (REPLACE_LOCAL);
            }
        }
    }

    return items;
}

// ---------------------------------------------------------------------------
// buildNormalMenu
// ---------------------------------------------------------------------------
//
// Go: menuNormal() — assembles working tree + timeline + history + add_remote
// SPEC §6 Normal operation

static juce::Array<MenuItemDef> buildNormalMenu (const juce::ValueTree& repo)
{
    juce::Array<MenuItemDef> items;

    const juce::Array<MenuItemDef> wtItems { workingTreeItems (repo) };
    items.addArray (wtItems);

    const juce::Array<MenuItemDef> tlItems { timelineItems (repo) };
    items.addArray (tlItems);

    // History section always present
    items.add (HISTORY);
    items.add (FILE_HISTORY);

    // Add remote when no remote configured
    const juce::String rmString { repo.getProperty (ID::remote).toString() };
    const Remote rm             { parseRemote (rmString) };

    if (rm == Remote::NoRemote)
        items.add (ADD_REMOTE);

    return items;
}

// ============================================================================
// MenuBuilder — constructor + build
// ============================================================================

MenuBuilder::MenuBuilder()
{
    generators[Operation::NotRepo]        = buildNotRepoMenu;
    generators[Operation::Normal]         = buildNormalMenu;
    generators[Operation::TimeTraveling]  = buildTimeTravelingMenu;
    generators[Operation::Merging]        = buildMergingMenu;
    generators[Operation::Conflicted]     = buildConflictedMenu;
    generators[Operation::Rebasing]       = buildRebasingMenu;
    generators[Operation::DirtyOperation] = buildDirtyOperationMenu;
    // Go: Rewinding is transient — by render time it is done; reuse Normal.
    generators[Operation::Rewinding]      = buildNormalMenu;
}

juce::Array<MenuItemDef> MenuBuilder::build (const juce::ValueTree& repoSubtree) const
{
    jassert (repoSubtree.isValid());

    const juce::String operationString { repoSubtree.getProperty (ID::operation).toString() };
    const Operation op                 { parseOperation (operationString) };

    return generators.at (op) (repoSubtree);
}

} // namespace tit::menu
