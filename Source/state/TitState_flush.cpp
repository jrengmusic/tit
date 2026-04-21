#include "TitState.h"

// ============================================================================
// writeIfChanged — compare / write / snapshot
// ============================================================================

template <typename T>
bool TitState::writeIfChanged (T&                     snapshotSlot,
                               T                      current,
                               juce::ValueTree&        subtree,
                               const juce::Identifier& key,
                               const juce::var&        asVar) noexcept
{
    bool didWrite { false };
    if (snapshotSlot != current)
    {
        subtree.setProperty (key, asVar, nullptr);
        snapshotSlot = current;
        didWrite     = true;
    }
    return didWrite;
}

// ============================================================================
// flushEnv
// ============================================================================

bool TitState::flushEnv() noexcept
{
    juce::ValueTree env { tree.getChildWithName (ID::ENV) };
    bool changed { false };

    const bool gitAvail { gitAvailableAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.gitAvailable, gitAvail, env, ID::gitAvailable, juce::var { gitAvail }))
        changed = true;

    const bool sshAvail { sshAvailableAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.sshAvailable, sshAvail, env, ID::sshAvailable, juce::var { sshAvail }))
        changed = true;

    const bool sshKeys { sshKeysPresentAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.sshKeysPresent, sshKeys, env, ID::sshKeysPresent, juce::var { sshKeys }))
        changed = true;

    const int setupStateVal { setupStateAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.setupState, setupStateVal, env, ID::setupState,
                        toString (static_cast<GitEnvironment> (setupStateVal))))
        changed = true;

    return changed;
}

// ============================================================================
// flushRepo
// ============================================================================

bool TitState::flushRepo() noexcept
{
    juce::ValueTree repo { tree.getChildWithName (ID::REPO) };
    bool changed { false };

    const int workingTreeVal { workingTreeAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.workingTree, workingTreeVal, repo, ID::workingTree,
                        toString (static_cast<WorkingTree> (workingTreeVal))))
        changed = true;

    const int timelineVal { timelineAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.timeline, timelineVal, repo, ID::timeline,
                        toString (static_cast<Timeline> (timelineVal))))
        changed = true;

    const int operationVal { operationAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.operation, operationVal, repo, ID::operation,
                        toString (static_cast<Operation> (operationVal))))
        changed = true;

    const int remoteVal { remoteAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.remote, remoteVal, repo, ID::remote,
                        toString (static_cast<Remote> (remoteVal))))
        changed = true;

    const bool ttVal { isTitTimeTravelAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.isTitTimeTravel, ttVal, repo, ID::isTitTimeTravel, juce::var { ttVal }))
        changed = true;

    const int aheadVal { aheadCountAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.aheadCount, aheadVal, repo, ID::aheadCount, juce::var { aheadVal }))
        changed = true;

    const int behindVal { behindCountAtom.load (std::memory_order_relaxed) };
    if (writeIfChanged (lastFlushed.behindCount, behindVal, repo, ID::behindCount, juce::var { behindVal }))
        changed = true;

    return changed;
}

// ============================================================================
// flushStrings
// ============================================================================

bool TitState::flushStrings() noexcept
{
    juce::ValueTree repo { tree.getChildWithName (ID::REPO) };
    bool changed { false };

    if (writeIfChanged (lastFlushed.branch, branch, repo, ID::branch, juce::var { branch }))
        changed = true;

    if (writeIfChanged (lastFlushed.cwd, cwd, repo, ID::cwd, juce::var { cwd }))
        changed = true;

    return changed;
}
