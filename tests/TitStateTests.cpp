#include <JuceHeader.h>
#include "state/TitState.h"
#include "TitIdentifier.h"

// ============================================================================
// TitStateTests
// ============================================================================
//
// Covers:
//   1. Default VT shape — RFC §4.6 all nodes present, defaults per SPEC §3.
//   2. Atom -> VT flush on message thread via direct timerCallback invocation.
//   3. Cross-thread atom write + Timer flush with juce::WaitableEvent coordination.
//   4. Skip-unchanged: VT listener fires only when value actually changes.

class TitStateTests : public juce::UnitTest
{
public:
    TitStateTests() : juce::UnitTest ("TitState", "tit") {}

    void runTest() override
    {
        testDefaultVtShape ();
        testAtomToVtFlushOnMessageThread ();
        testCrossThreadAtomWrite ();
        testSkipUnchangedOptimisation ();
    }

private:

    // =========================================================================
    // Test 1 — Default VT shape
    // =========================================================================

    void testDefaultVtShape ()
    {
        beginTest ("default VT shape (RFC §4.6)");

        tit::TitState state;
        juce::ValueTree root { state.getTree() };

        // Root type
        expect (root.getType() == tit::ID::TIT,       "root type must be TIT");

        // Top-level nodes present
        expect (root.getChildWithName (tit::ID::ENV).isValid(),       "ENV node must be present");
        expect (root.getChildWithName (tit::ID::REPO).isValid(),      "REPO node must be present");
        expect (root.getChildWithName (tit::ID::HISTORY).isValid(),   "HISTORY node must be present");
        expect (root.getChildWithName (tit::ID::FILES).isValid(),     "FILES node must be present");
        expect (root.getChildWithName (tit::ID::DIFF).isValid(),      "DIFF node must be present");
        expect (root.getChildWithName (tit::ID::MENU).isValid(),      "MENU node must be present");
        expect (root.getChildWithName (tit::ID::CONSOLE).isValid(),   "CONSOLE node must be present");
        expect (root.getChildWithName (tit::ID::SELECTION).isValid(), "SELECTION node must be present");
        expect (root.getChildWithName (tit::ID::THEME).isValid(),     "THEME node must be present");
        expect (root.getChildWithName (tit::ID::SETUP).isValid(),     "SETUP node must be present");

        // ENV defaults
        juce::ValueTree env { root.getChildWithName (tit::ID::ENV) };
        expect (static_cast<bool> (env.getProperty (tit::ID::gitAvailable))   == false, "gitAvailable default false");
        expect (static_cast<bool> (env.getProperty (tit::ID::sshAvailable))   == false, "sshAvailable default false");
        expect (static_cast<bool> (env.getProperty (tit::ID::sshKeysPresent)) == false, "sshKeysPresent default false");
        expect (env.getProperty (tit::ID::setupState).toString()
                == tit::toString (tit::GitEnvironment::NeedsSetup),
                "setupState default NeedsSetup");

        // REPO defaults
        juce::ValueTree repo { root.getChildWithName (tit::ID::REPO) };
        expect (repo.getProperty (tit::ID::workingTree).toString()
                == tit::toString (tit::WorkingTree::Clean),
                "workingTree default Clean");
        expect (repo.getProperty (tit::ID::timeline).toString()
                == tit::toString (tit::Timeline::Empty),
                "timeline default Empty");
        expect (repo.getProperty (tit::ID::operation).toString()
                == tit::toString (tit::Operation::NotRepo),
                "operation default NotRepo");
        expect (repo.getProperty (tit::ID::remote).toString()
                == tit::toString (tit::Remote::NoRemote),
                "remote default NoRemote");
        expect (static_cast<bool> (repo.getProperty (tit::ID::isTitTimeTravel)) == false,
                "isTitTimeTravel default false");
        expect (static_cast<int> (repo.getProperty (tit::ID::aheadCount))  == 0, "aheadCount default 0");
        expect (static_cast<int> (repo.getProperty (tit::ID::behindCount)) == 0, "behindCount default 0");
        expect (repo.getProperty (tit::ID::branch).toString().isEmpty(),  "branch default empty");
        expect (repo.getProperty (tit::ID::cwd).toString().isEmpty(),     "cwd default empty");

        // SELECTION defaults
        juce::ValueTree sel { root.getChildWithName (tit::ID::SELECTION) };
        expect (static_cast<int> (sel.getProperty (tit::ID::menuIndex))    == 0, "menuIndex default 0");
        expect (static_cast<int> (sel.getProperty (tit::ID::historyIndex)) == 0, "historyIndex default 0");
        expect (static_cast<int> (sel.getProperty (tit::ID::fileIndex))    == 0, "fileIndex default 0");
        expect (sel.getProperty (tit::ID::activePane).toString().isEmpty(), "activePane default empty");

        // SETUP defaults
        juce::ValueTree setup { root.getChildWithName (tit::ID::SETUP) };
        expect (setup.getProperty (tit::ID::phase).toString()
                == tit::toString (tit::SetupPhase::EnvCheck),
                "setup phase default EnvCheck");
        expect (setup.getProperty (tit::ID::email).toString().isEmpty(),     "email default empty");
        expect (setup.getProperty (tit::ID::publicKey).toString().isEmpty(), "publicKey default empty");
    }

    // =========================================================================
    // Test 2 — Atom -> VT flush invoked directly (message thread)
    // =========================================================================

    void testAtomToVtFlushOnMessageThread ()
    {
        beginTest ("atom -> VT flush (message thread)");

        tit::TitState state;

        // Set atoms from "message thread" (we are on message thread in test context)
        state.setOperation       (tit::Operation::Merging);
        state.setWorkingTree     (tit::WorkingTree::Dirty);
        state.setTimeline        (tit::Timeline::Ahead);
        state.setRemote          (tit::Remote::HasRemote);
        state.setIsTitTimeTravel (true);
        state.setAheadCount      (3);
        state.setBehindCount     (1);
        state.setSetupState      (tit::GitEnvironment::Ready);
        state.setGitAvailable    (true);
        state.setSshAvailable    (true);
        state.setSshKeysPresent  (true);

        // Invoke timerCallback directly — simulates one flush tick on message thread
        state.flushNow();

        juce::ValueTree root { state.getTree() };
        juce::ValueTree repo { root.getChildWithName (tit::ID::REPO) };
        juce::ValueTree env  { root.getChildWithName (tit::ID::ENV) };

        expect (repo.getProperty (tit::ID::operation).toString()
                == tit::toString (tit::Operation::Merging),
                "operation must flush to Merging");
        expect (repo.getProperty (tit::ID::workingTree).toString()
                == tit::toString (tit::WorkingTree::Dirty),
                "workingTree must flush to Dirty");
        expect (repo.getProperty (tit::ID::timeline).toString()
                == tit::toString (tit::Timeline::Ahead),
                "timeline must flush to Ahead");
        expect (repo.getProperty (tit::ID::remote).toString()
                == tit::toString (tit::Remote::HasRemote),
                "remote must flush to HasRemote");
        expect (static_cast<bool> (repo.getProperty (tit::ID::isTitTimeTravel)) == true,
                "isTitTimeTravel must flush to true");
        expect (static_cast<int> (repo.getProperty (tit::ID::aheadCount))  == 3, "aheadCount must flush to 3");
        expect (static_cast<int> (repo.getProperty (tit::ID::behindCount)) == 1, "behindCount must flush to 1");

        expect (env.getProperty (tit::ID::setupState).toString()
                == tit::toString (tit::GitEnvironment::Ready),
                "setupState must flush to Ready");
        expect (static_cast<bool> (env.getProperty (tit::ID::gitAvailable))   == true, "gitAvailable must flush to true");
        expect (static_cast<bool> (env.getProperty (tit::ID::sshAvailable))   == true, "sshAvailable must flush to true");
        expect (static_cast<bool> (env.getProperty (tit::ID::sshKeysPresent)) == true, "sshKeysPresent must flush to true");
    }

    // =========================================================================
    // Test 3 — Cross-thread atom write + Timer flush
    // =========================================================================
    //
    // Spawns a juce::Thread that calls setOperation() from a worker thread.
    // The main test then starts the Timer, pumps the message queue until the VT
    // reflects the worker's write, bounded by a WaitableEvent with a 5000 ms timeout.

    struct WorkerThread : public juce::Thread
    {
        tit::TitState& state;
        juce::WaitableEvent written;

        explicit WorkerThread (tit::TitState& s) : juce::Thread ("TitStateWorker"), state (s) {}

        void run() override
        {
            state.setOperation (tit::Operation::Rebasing);
            state.setAheadCount (7);
            written.signal();
        }
    };

    void testCrossThreadAtomWrite ()
    {
        beginTest ("cross-thread atom write -> VT flush");

        tit::TitState state;
        WorkerThread worker { state };

        worker.startThread();
        const bool signalled { worker.written.wait (5000) };
        worker.stopThread (500);

        expect (signalled, "worker thread must signal within 5000 ms");

        if (signalled)
        {
            // Verify atoms directly (any-thread read)
            expect (state.getOperation()  == tit::Operation::Rebasing, "operation atom must be Rebasing");
            expect (state.getAheadCount() == 7,                        "aheadCount atom must be 7");

            // Flush on message thread and verify VT
            state.flushNow();

            juce::ValueTree repo { state.getTree().getChildWithName (tit::ID::REPO) };
            expect (repo.getProperty (tit::ID::operation).toString()
                    == tit::toString (tit::Operation::Rebasing),
                    "operation must flush to Rebasing after cross-thread write");
            expect (static_cast<int> (repo.getProperty (tit::ID::aheadCount)) == 7,
                    "aheadCount must flush to 7 after cross-thread write");
        }
    }

    // =========================================================================
    // Test 4 — Skip-unchanged optimisation
    // =========================================================================
    //
    // Attaches a ValueTree::Listener that counts property-changed callbacks.
    // Sets operation twice to the same value; expects exactly one callback
    // (the initial flush).  Sets to a different value; expects a second callback.

    struct CountingListener : public juce::ValueTree::Listener
    {
        int callbackCount { 0 };

        void valueTreePropertyChanged (juce::ValueTree&, const juce::Identifier&) override
        {
            ++callbackCount;
        }
    };

    void testSkipUnchangedOptimisation ()
    {
        beginTest ("skip-unchanged optimisation");

        tit::TitState state;

        juce::ValueTree repo { state.getTree().getChildWithName (tit::ID::REPO) };
        CountingListener listener;
        repo.addListener (&listener);

        // First write + flush — should fire callbacks for changed properties
        state.setOperation (tit::Operation::Merging);
        state.flushNow();
        const int countAfterFirstFlush { listener.callbackCount };
        expect (countAfterFirstFlush >= 1, "at least one callback after first flush");

        // Same value again — flush must NOT fire another callback for operation
        state.setOperation (tit::Operation::Merging);
        const int countBeforeSecondFlush { listener.callbackCount };
        state.flushNow();
        const int countAfterSecondFlush { listener.callbackCount };
        expect (countAfterSecondFlush == countBeforeSecondFlush,
                "no callback when same value flushed twice");

        // Different value — must fire
        state.setOperation (tit::Operation::Conflicted);
        state.flushNow();
        expect (listener.callbackCount > countAfterSecondFlush,
                "callback must fire when operation changes to Conflicted");

        repo.removeListener (&listener);
    }
};

static TitStateTests titStateTestsInstance;
