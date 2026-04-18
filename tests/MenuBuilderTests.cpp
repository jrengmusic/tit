#include <JuceHeader.h>
#include "menu/MenuBuilder.h"
#include "menu/MenuItems.h"
#include "state/TitAxis.h"
#include "TitIdentifier.h"

// ============================================================================
// MenuBuilderTests
// ============================================================================
//
// Covers:
//   1. Determinism — same repo VT + same operation → bit-identical array.
//   2. Coverage — every Operation enum value has a generator; .at() never throws.
//   3. Per-Operation shape — item counts + IDs match SPEC §6 canonical cases.
//   4. Conditional enabling — aheadCount / remote / workingTree / timeline
//      variations cause items to appear / disappear per SPEC §6 rules.

class MenuBuilderTests : public juce::UnitTest
{
public:
    MenuBuilderTests() : juce::UnitTest ("MenuBuilder", "tit") {}

    void runTest() override
    {
        testDeterminism ();
        testAllOperationsCovered ();
        testNotRepoShape ();
        testTimeTravelingShape ();
        testMergingShape ();
        testConflictedShape ();
        testRebasingShape ();
        testDirtyOperationShape ();
        testRewindingShape ();
        testNormalConditionals ();
    }

private:

    // =========================================================================
    // Fixture helpers
    // =========================================================================

    // makeRepo builds a minimal REPO ValueTree with all five axis properties.
    static juce::ValueTree makeRepo (tit::Operation    op,
                                     tit::WorkingTree  wt  = tit::WorkingTree::Clean,
                                     tit::Timeline     tl  = tit::Timeline::InSync,
                                     tit::Remote       rm  = tit::Remote::HasRemote,
                                     int               aheadCount  = 0,
                                     int               behindCount = 0)
    {
        juce::ValueTree repo { tit::ID::REPO };
        repo.setProperty (tit::ID::operation,   tit::toString (op), nullptr);
        repo.setProperty (tit::ID::workingTree, tit::toString (wt), nullptr);
        repo.setProperty (tit::ID::timeline,    tit::toString (tl), nullptr);
        repo.setProperty (tit::ID::remote,      tit::toString (rm), nullptr);
        repo.setProperty (tit::ID::aheadCount,  aheadCount,         nullptr);
        repo.setProperty (tit::ID::behindCount, behindCount,        nullptr);
        return repo;
    }

    // arrayIdsMatch checks that every element of 'result' has the expected ID
    // in order, and counts match.
    bool arrayMatchesIds (const juce::Array<tit::menu::MenuItemDef>& result,
                          std::initializer_list<const char*> expectedIds) const
    {
        if (result.size() != static_cast<int> (expectedIds.size()))
            return false;

        int index { 0 };
        for (const char* expected : expectedIds)
        {
            if (juce::String (result[index].id) != juce::String (expected))
                return false;
            ++index;
        }
        return true;
    }

    // =========================================================================
    // Test 1 — Determinism
    // =========================================================================

    void testDeterminism ()
    {
        beginTest ("determinism (same VT -> bit-identical output)");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                         tit::WorkingTree::Dirty,
                                         tit::Timeline::Ahead,
                                         tit::Remote::HasRemote, 2, 0) };

        const juce::Array<tit::menu::MenuItemDef> first  { builder.build (repo) };
        const juce::Array<tit::menu::MenuItemDef> second { builder.build (repo) };

        expect (first.size() == second.size(), "result sizes must match");

        for (int i = 0; i < first.size(); ++i)
        {
            const bool idMatch    { juce::String (first[i].id)    == juce::String (second[i].id) };
            const bool labelMatch { juce::String (first[i].label) == juce::String (second[i].label) };
            const bool hotkeyMatch { first[i].hotkey      == second[i].hotkey };
            const bool destructMatch { first[i].destructive == second[i].destructive };

            expect (idMatch,      juce::String ("id mismatch at index ")          + juce::String (i));
            expect (labelMatch,   juce::String ("label mismatch at index ")       + juce::String (i));
            expect (hotkeyMatch,  juce::String ("hotkey mismatch at index ")      + juce::String (i));
            expect (destructMatch,juce::String ("destructive mismatch at index ") + juce::String (i));
        }
    }

    // =========================================================================
    // Test 2 — All Operation values covered (no .at() throw)
    // =========================================================================

    void testAllOperationsCovered ()
    {
        beginTest ("all Operation values have registered generators");

        tit::menu::MenuBuilder builder;

        // All 8 Operation values from TitAxis.h
        const tit::Operation allOps[]
        {
            tit::Operation::NotRepo,
            tit::Operation::Normal,
            tit::Operation::Merging,
            tit::Operation::Conflicted,
            tit::Operation::Rebasing,
            tit::Operation::TimeTraveling,
            tit::Operation::DirtyOperation,
            tit::Operation::Rewinding,
        };

        for (tit::Operation op : allOps)
        {
            juce::ValueTree repo { makeRepo (op) };

            bool threw { false };

            try
            {
                const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
                expect (items.size() >= 0, "result must be non-negative sized array");
            }
            catch (...)
            {
                threw = true;
            }

            expect (not threw, juce::String ("generator threw for Operation: ")
                               + tit::toString (op));
        }
    }

    // =========================================================================
    // Test 3 — NotRepo shape
    // =========================================================================

    void testNotRepoShape ()
    {
        beginTest ("NotRepo shape: [init, clone]");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::NotRepo,
                                         tit::WorkingTree::Clean,
                                         tit::Timeline::Empty,
                                         tit::Remote::NoRemote) };

        const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };

        expect (arrayMatchesIds (items, { "init", "clone" }),
                "NotRepo must return [init, clone]");
    }

    // =========================================================================
    // Test 4 — TimeTraveling shape
    // =========================================================================

    void testTimeTravelingShape ()
    {
        beginTest ("TimeTraveling shape: [time_travel_history, time_travel_files_history, time_travel_return]");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::TimeTraveling,
                                         tit::WorkingTree::Clean,
                                         tit::Timeline::Empty,
                                         tit::Remote::HasRemote) };

        const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };

        expect (arrayMatchesIds (items,
                                 { "time_travel_history",
                                   "time_travel_files_history",
                                   "time_travel_return" }),
                "TimeTraveling must return [history, files history, return]");
    }

    // =========================================================================
    // Test 5 — Merging shape
    // =========================================================================

    void testMergingShape ()
    {
        beginTest ("Merging shape: [finalize_merge, abort_merge]");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::Merging) };

        const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };

        expect (arrayMatchesIds (items, { "finalize_merge", "abort_merge" }),
                "Merging must return [finalize_merge, abort_merge]");
    }

    // =========================================================================
    // Test 6 — Conflicted shape
    // =========================================================================

    void testConflictedShape ()
    {
        beginTest ("Conflicted shape: [abort_merge]");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::Conflicted) };

        const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };

        expect (arrayMatchesIds (items, { "abort_merge" }),
                "Conflicted must return [abort_merge]");
    }

    // =========================================================================
    // Test 7 — Rebasing shape
    // =========================================================================

    void testRebasingShape ()
    {
        beginTest ("Rebasing shape: [rebase_continue, rebase_abort]");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::Rebasing) };

        const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };

        expect (arrayMatchesIds (items, { "rebase_continue", "rebase_abort" }),
                "Rebasing must return [rebase_continue, rebase_abort]");
    }

    // =========================================================================
    // Test 8 — DirtyOperation shape
    // =========================================================================

    void testDirtyOperationShape ()
    {
        beginTest ("DirtyOperation shape: [abort_merge]");

        tit::menu::MenuBuilder builder;
        juce::ValueTree repo { makeRepo (tit::Operation::DirtyOperation) };

        const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };

        expect (arrayMatchesIds (items, { "abort_merge" }),
                "DirtyOperation must return [abort_merge]");
    }

    // =========================================================================
    // Test 9 — Rewinding shape (mirrors Normal — transient state)
    // =========================================================================

    void testRewindingShape ()
    {
        beginTest ("Rewinding shape mirrors Normal (Clean + InSync + HasRemote)");

        tit::menu::MenuBuilder builder;

        juce::ValueTree repoRewinding { makeRepo (tit::Operation::Rewinding,
                                                   tit::WorkingTree::Clean,
                                                   tit::Timeline::InSync,
                                                   tit::Remote::HasRemote) };
        juce::ValueTree repoNormal    { makeRepo (tit::Operation::Normal,
                                                   tit::WorkingTree::Clean,
                                                   tit::Timeline::InSync,
                                                   tit::Remote::HasRemote) };

        const juce::Array<tit::menu::MenuItemDef> rewindItems  { builder.build (repoRewinding) };
        const juce::Array<tit::menu::MenuItemDef> normalItems  { builder.build (repoNormal) };

        expect (rewindItems.size() == normalItems.size(),
                "Rewinding item count must equal Normal item count");

        for (int i = 0; i < rewindItems.size(); ++i)
        {
            expect (juce::String (rewindItems[i].id) == juce::String (normalItems[i].id),
                    juce::String ("id mismatch at index ") + juce::String (i));
        }
    }

    // =========================================================================
    // Test 10 — Normal conditional enabling (working tree + timeline + remote)
    // =========================================================================

    void testNormalConditionals ()
    {
        beginTest ("Normal — conditional item inclusion per SPEC §6");

        tit::menu::MenuBuilder builder;

        // ---- 10a: Clean + InSync + HasRemote → only history items
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Clean,
                                             tit::Timeline::InSync,
                                             tit::Remote::HasRemote) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items, { "history", "file_history" }),
                    "Clean+InSync+HasRemote must yield only [history, file_history]");
        }

        // ---- 10b: Clean + InSync + NoRemote → history + add_remote
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Clean,
                                             tit::Timeline::InSync,
                                             tit::Remote::NoRemote) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items, { "history", "file_history", "add_remote" }),
                    "Clean+InSync+NoRemote must yield [history, file_history, add_remote]");
        }

        // ---- 10c: Dirty + InSync + HasRemote → commit, commit_push, discard, history, file_history
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Dirty,
                                             tit::Timeline::InSync,
                                             tit::Remote::HasRemote) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "commit", "commit_push", "reset_discard_changes",
                                       "history", "file_history" }),
                    "Dirty+InSync+HasRemote must include commit, commit_push, discard, history, file_history");
        }

        // ---- 10d: Dirty + InSync + NoRemote → commit (no commit_push), discard, history, file_history, add_remote
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Dirty,
                                             tit::Timeline::InSync,
                                             tit::Remote::NoRemote) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "commit", "reset_discard_changes",
                                       "history", "file_history", "add_remote" }),
                    "Dirty+InSync+NoRemote must include commit (no commit_push), discard, history, file_history, add_remote");
        }

        // ---- 10e: Clean + Ahead + HasRemote → push, force_push, history, file_history
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Clean,
                                             tit::Timeline::Ahead,
                                             tit::Remote::HasRemote, 2, 0) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "push", "force_push", "history", "file_history" }),
                    "Clean+Ahead must yield [push, force_push, history, file_history]");
        }

        // ---- 10f: Dirty + Ahead + HasRemote → commit, commit_push, discard, history, file_history
        //           (no push — cannot push uncommitted changes)
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Dirty,
                                             tit::Timeline::Ahead,
                                             tit::Remote::HasRemote, 3, 0) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "commit", "commit_push", "reset_discard_changes",
                                       "history", "file_history" }),
                    "Dirty+Ahead must not include push items");
        }

        // ---- 10g: Clean + Behind + HasRemote → pull_merge, replace_local, history, file_history
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Clean,
                                             tit::Timeline::Behind,
                                             tit::Remote::HasRemote, 0, 3) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "pull_merge", "replace_local",
                                       "history", "file_history" }),
                    "Clean+Behind must yield [pull_merge, replace_local, history, file_history]");
        }

        // ---- 10h: Dirty + Behind + HasRemote → commit, commit_push, discard, dirty_pull_merge, replace_local, history, file_history
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Dirty,
                                             tit::Timeline::Behind,
                                             tit::Remote::HasRemote, 0, 2) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "commit", "commit_push", "reset_discard_changes",
                                       "dirty_pull_merge", "replace_local",
                                       "history", "file_history" }),
                    "Dirty+Behind must yield [commit, commit_push, discard, dirty_pull_merge, replace_local, history, file_history]");
        }

        // ---- 10i: Clean + Diverged + HasRemote → push_auto_sync, pull_merge_diverged, force_push, replace_local, history, file_history
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Clean,
                                             tit::Timeline::Diverged,
                                             tit::Remote::HasRemote, 1, 1) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "push_auto_sync", "pull_merge_diverged",
                                       "force_push", "replace_local",
                                       "history", "file_history" }),
                    "Clean+Diverged must yield full diverged set + history");
        }

        // ---- 10j: Dirty + Diverged + HasRemote → commit, commit_push, discard, dirty_pull_merge, force_push, replace_local, history, file_history
        {
            juce::ValueTree repo { makeRepo (tit::Operation::Normal,
                                             tit::WorkingTree::Dirty,
                                             tit::Timeline::Diverged,
                                             tit::Remote::HasRemote, 2, 2) };
            const juce::Array<tit::menu::MenuItemDef> items { builder.build (repo) };
            expect (arrayMatchesIds (items,
                                     { "commit", "commit_push", "reset_discard_changes",
                                       "dirty_pull_merge", "force_push", "replace_local",
                                       "history", "file_history" }),
                    "Dirty+Diverged must yield [commit, commit_push, discard, dirty_pull_merge, force_push, replace_local, history, file_history]");
        }
    }
};

static MenuBuilderTests menuBuilderTestsInstance;
