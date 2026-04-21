#pragma once
#include <JuceHeader.h>
#include <atomic>
#include "TitAxis.h"
#include "TitIdentifier.h"

// ============================================================================
// TitState
// ============================================================================
//
// APVTS-style atomic state bridge for TIT-cpp.  Mirrors END's Terminal::State
// structure verbatim, adapted for TIT's five-axis git state model.
//
// Thread ownership:
//   - Worker threads (git subprocess, detector) write exclusively to atoms
//     via set*() setters.  No ValueTree mutations, no allocations, no locks.
//   - The flush Timer (message thread) reads atoms -> writes ValueTree on
//     each tick, firing ValueTree::Listeners.
//   - Message thread reads ValueTree for UI consumption; also owns non-atom
//     fields (branch, cwd) with jassert guards on their setters.
//   - Views attach juce::ValueTree::Listener to relevant subtrees.
//
// Zero locks on the hot path.  Zero shadow state.
// Atoms ARE the state; ValueTree REFLECTS it.
//
// RFC §3.6, §4.6; BLESSED B (thread bounds), S (SSOT).

class TitState : public juce::Timer
{
public:

    // =========================================================================
    // Construction / Destruction
    // =========================================================================

    // Constructs default ValueTree per RFC §4.6 schema.
    // All nodes present; REPO/ENV properties initialised to SPEC §3 defaults.
    // MESSAGE THREAD — constructed before Timer is started via start().
    TitState();

    // Stops the Timer and destroys TitState.
    // MESSAGE THREAD — must be destroyed on the message thread.
    ~TitState() override;

    // =========================================================================
    // Lifecycle
    // =========================================================================

    // Begins the flush Timer at FLUSH_INTERVAL_MS cadence.
    // MESSAGE THREAD.
    void start() noexcept;

    // Stops the flush Timer.
    // MESSAGE THREAD.
    void stop() noexcept;

    // =========================================================================
    // ValueTree accessor (message-thread observation surface)
    // =========================================================================

    // Returns the root TIT ValueTree for view consumption.
    // Views attach juce::ValueTree::Listener to this or a subtree.
    // MESSAGE THREAD.
    juce::ValueTree getTree() const noexcept;

    // Forces an immediate flush without waiting for the next Timer tick.
    // Used by tests to drive synchronous flush on the message thread.
    // Returns true if any property was updated.
    // MESSAGE THREAD.
    bool flushNow() noexcept;

    // =========================================================================
    // Atom setters — worker-thread fast writes (any thread, lock-free)
    // =========================================================================

    void setSetupState      (GitEnvironment v) noexcept;
    void setWorkingTree     (WorkingTree    v) noexcept;
    void setTimeline        (Timeline       v) noexcept;
    void setOperation       (Operation      v) noexcept;
    void setRemote          (Remote         v) noexcept;
    void setIsTitTimeTravel (bool           v) noexcept;
    void setAheadCount      (int            v) noexcept;
    void setBehindCount     (int            v) noexcept;
    void setGitAvailable    (bool           v) noexcept;
    void setSshAvailable    (bool           v) noexcept;
    void setSshKeysPresent  (bool           v) noexcept;

    // =========================================================================
    // Atom getters — any thread, lock-free
    // =========================================================================

    GitEnvironment  getSetupState()      const noexcept;
    WorkingTree     getWorkingTree()     const noexcept;
    Timeline        getTimeline()        const noexcept;
    Operation       getOperation()       const noexcept;
    Remote          getRemote()          const noexcept;
    bool            getIsTitTimeTravel() const noexcept;
    int             getAheadCount()      const noexcept;
    int             getBehindCount()     const noexcept;
    bool            getGitAvailable()    const noexcept;
    bool            getSshAvailable()    const noexcept;
    bool            getSshKeysPresent()  const noexcept;

    // =========================================================================
    // Message-thread-only setters for non-atom fields
    // =========================================================================

    // Sets the branch string and marks flush dirty.
    // MESSAGE THREAD — jassert guards this boundary.
    void setBranch (const juce::String& v) noexcept;

    // Sets the cwd string and marks flush dirty.
    // MESSAGE THREAD — jassert guards this boundary.
    void setCwd (const juce::String& v) noexcept;

private:

    // =========================================================================
    // Timer
    // =========================================================================

    // Flush cadence: 16 ms ≈ 60 Hz.  Named constant per BLESSED-E (no magic numbers).
    static constexpr int FLUSH_INTERVAL_MS { 16 };

    // timerCallback runs on the message thread; delegates to flush().
    void timerCallback() override;

    // Copies dirty atoms -> ValueTree properties in one pass.
    // Skip-unchanged optimisation: compares against lastFlushed snapshot.
    // Returns true if any property was updated.
    // MESSAGE THREAD.
    bool flush() noexcept;

    // Per-subtree flush helpers — called exclusively from flush().
    // Each returns true if any property in its subtree was written.
    // MESSAGE THREAD.
    bool flushEnv()     noexcept;
    bool flushRepo()    noexcept;
    bool flushStrings() noexcept;

    // Compare-write-snapshot helper.  If snapshotSlot != current, writes
    // asVar to subtree[key], updates snapshotSlot, and returns true.
    // Called only from flush helpers (message thread).
    template <typename T>
    bool writeIfChanged (T&                     snapshotSlot,
                         T                      current,
                         juce::ValueTree&        subtree,
                         const juce::Identifier& key,
                         const juce::var&        asVar) noexcept;

    // =========================================================================
    // ValueTree — message-thread SSOT (observation surface)
    // =========================================================================

    juce::ValueTree tree;

    // =========================================================================
    // Atoms — cross-thread fast-write primitives
    // =========================================================================

    std::atomic<int>  setupStateAtom     { static_cast<int> (GitEnvironment::NeedsSetup) };
    std::atomic<int>  workingTreeAtom    { static_cast<int> (WorkingTree::Clean) };
    std::atomic<int>  timelineAtom       { static_cast<int> (Timeline::Empty) };
    std::atomic<int>  operationAtom      { static_cast<int> (Operation::NotRepo) };
    std::atomic<int>  remoteAtom         { static_cast<int> (Remote::NoRemote) };
    std::atomic<bool> isTitTimeTravelAtom { false };
    std::atomic<int>  aheadCountAtom     { 0 };
    std::atomic<int>  behindCountAtom    { 0 };
    std::atomic<bool> gitAvailableAtom   { false };
    std::atomic<bool> sshAvailableAtom   { false };
    std::atomic<bool> sshKeysPresentAtom { false };

    // needsFlush is set by any atom setter; read+cleared in timerCallback.
    std::atomic<bool> needsFlush { false };

    // =========================================================================
    // Non-atom message-thread fields
    // =========================================================================

    juce::String branch;
    juce::String cwd;

    // =========================================================================
    // Last-flushed snapshot (message-thread owned, skip-unchanged optimisation)
    // =========================================================================

    struct FlushSnapshot
    {
        int  setupState     { static_cast<int> (GitEnvironment::NeedsSetup) };
        int  workingTree    { static_cast<int> (WorkingTree::Clean) };
        int  timeline       { static_cast<int> (Timeline::Empty) };
        int  operation      { static_cast<int> (Operation::NotRepo) };
        int  remote         { static_cast<int> (Remote::NoRemote) };
        bool isTitTimeTravel { false };
        int  aheadCount     { 0 };
        int  behindCount    { 0 };
        bool gitAvailable   { false };
        bool sshAvailable   { false };
        bool sshKeysPresent { false };
        juce::String branch;
        juce::String cwd;
    };

    FlushSnapshot lastFlushed;
};
