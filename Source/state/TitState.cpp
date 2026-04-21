#include "TitState.h"

// ============================================================================
// Construction / Destruction
// ============================================================================

TitState::TitState()
{
    // Build ValueTree skeleton per RFC §4.6 — all nodes present, defaults per
    // SPEC §3.  Tree is owned by message thread; constructed before Timer starts.

    tree = juce::ValueTree { ID::TIT };

    // ENV node
    juce::ValueTree env { ID::ENV };
    env.setProperty (ID::gitAvailable,   false,                                  nullptr);
    env.setProperty (ID::sshAvailable,   false,                                  nullptr);
    env.setProperty (ID::sshKeysPresent, false,                                  nullptr);
    env.setProperty (ID::setupState,     toString (GitEnvironment::NeedsSetup),  nullptr);
    tree.addChild (env, -1, nullptr);

    // REPO node
    juce::ValueTree repo { ID::REPO };
    repo.setProperty (ID::workingTree,     toString (WorkingTree::Clean),         nullptr);
    repo.setProperty (ID::timeline,        toString (Timeline::Empty),            nullptr);
    repo.setProperty (ID::operation,       toString (Operation::NotRepo),         nullptr);
    repo.setProperty (ID::remote,          toString (Remote::NoRemote),           nullptr);
    repo.setProperty (ID::isTitTimeTravel, false,                                 nullptr);
    repo.setProperty (ID::branch,          juce::String{},                        nullptr);
    repo.setProperty (ID::aheadCount,      0,                                     nullptr);
    repo.setProperty (ID::behindCount,     0,                                     nullptr);
    repo.setProperty (ID::cwd,             juce::String{},                        nullptr);
    tree.addChild (repo, -1, nullptr);

    // HISTORY node
    tree.addChild (juce::ValueTree { ID::HISTORY },   -1, nullptr);

    // FILES node
    tree.addChild (juce::ValueTree { ID::FILES },     -1, nullptr);

    // DIFF node
    tree.addChild (juce::ValueTree { ID::DIFF },      -1, nullptr);

    // MENU node
    tree.addChild (juce::ValueTree { ID::MENU },      -1, nullptr);

    // CONSOLE node
    tree.addChild (juce::ValueTree { ID::CONSOLE },   -1, nullptr);

    // SELECTION node
    juce::ValueTree sel { ID::SELECTION };
    sel.setProperty (ID::menuIndex,    0,               nullptr);
    sel.setProperty (ID::historyIndex, 0,               nullptr);
    sel.setProperty (ID::fileIndex,    0,               nullptr);
    sel.setProperty (ID::activePane,   juce::String{},  nullptr);
    tree.addChild (sel, -1, nullptr);

    // THEME node
    tree.addChild (juce::ValueTree { ID::THEME }, -1, nullptr);

    // SETUP node
    juce::ValueTree setup { ID::SETUP };
    setup.setProperty (ID::phase,     toString (SetupPhase::EnvCheck), nullptr);
    setup.setProperty (ID::email,     juce::String{},                  nullptr);
    setup.setProperty (ID::publicKey, juce::String{},                  nullptr);
    tree.addChild (setup, -1, nullptr);
}

TitState::~TitState()
{
    stop();
}

// ============================================================================
// Lifecycle
// ============================================================================

void TitState::start() noexcept
{
    startTimer (FLUSH_INTERVAL_MS);
}

void TitState::stop() noexcept
{
    stopTimer();
}

// ============================================================================
// ValueTree accessor
// ============================================================================

juce::ValueTree TitState::getTree() const noexcept
{
    return tree;
}

bool TitState::flushNow() noexcept
{
    jassert (juce::MessageManager::getInstance()->isThisTheMessageThread());
    needsFlush.store (true, std::memory_order_release);
    return flush();
}

// ============================================================================
// Atom setters
// ============================================================================

void TitState::setSetupState (GitEnvironment v) noexcept
{
    setupStateAtom.store (static_cast<int> (v), std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setWorkingTree (WorkingTree v) noexcept
{
    workingTreeAtom.store (static_cast<int> (v), std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setTimeline (Timeline v) noexcept
{
    timelineAtom.store (static_cast<int> (v), std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setOperation (Operation v) noexcept
{
    operationAtom.store (static_cast<int> (v), std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setRemote (Remote v) noexcept
{
    remoteAtom.store (static_cast<int> (v), std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setIsTitTimeTravel (bool v) noexcept
{
    isTitTimeTravelAtom.store (v, std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setAheadCount (int v) noexcept
{
    aheadCountAtom.store (v, std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setBehindCount (int v) noexcept
{
    behindCountAtom.store (v, std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setGitAvailable (bool v) noexcept
{
    gitAvailableAtom.store (v, std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setSshAvailable (bool v) noexcept
{
    sshAvailableAtom.store (v, std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setSshKeysPresent (bool v) noexcept
{
    sshKeysPresentAtom.store (v, std::memory_order_relaxed);
    needsFlush.store (true, std::memory_order_release);
}

// ============================================================================
// Atom getters
// ============================================================================

GitEnvironment TitState::getSetupState() const noexcept
{
    return static_cast<GitEnvironment> (setupStateAtom.load (std::memory_order_relaxed));
}

WorkingTree TitState::getWorkingTree() const noexcept
{
    return static_cast<WorkingTree> (workingTreeAtom.load (std::memory_order_relaxed));
}

Timeline TitState::getTimeline() const noexcept
{
    return static_cast<Timeline> (timelineAtom.load (std::memory_order_relaxed));
}

Operation TitState::getOperation() const noexcept
{
    return static_cast<Operation> (operationAtom.load (std::memory_order_relaxed));
}

Remote TitState::getRemote() const noexcept
{
    return static_cast<Remote> (remoteAtom.load (std::memory_order_relaxed));
}

bool TitState::getIsTitTimeTravel() const noexcept
{
    return isTitTimeTravelAtom.load (std::memory_order_relaxed);
}

int TitState::getAheadCount() const noexcept
{
    return aheadCountAtom.load (std::memory_order_relaxed);
}

int TitState::getBehindCount() const noexcept
{
    return behindCountAtom.load (std::memory_order_relaxed);
}

bool TitState::getGitAvailable() const noexcept
{
    return gitAvailableAtom.load (std::memory_order_relaxed);
}

bool TitState::getSshAvailable() const noexcept
{
    return sshAvailableAtom.load (std::memory_order_relaxed);
}

bool TitState::getSshKeysPresent() const noexcept
{
    return sshKeysPresentAtom.load (std::memory_order_relaxed);
}

// ============================================================================
// Message-thread-only setters
// ============================================================================

void TitState::setBranch (const juce::String& v) noexcept
{
    jassert (juce::MessageManager::getInstance()->isThisTheMessageThread());
    branch = v;
    needsFlush.store (true, std::memory_order_release);
}

void TitState::setCwd (const juce::String& v) noexcept
{
    jassert (juce::MessageManager::getInstance()->isThisTheMessageThread());
    cwd = v;
    needsFlush.store (true, std::memory_order_release);
}

// ============================================================================
// Timer
// ============================================================================

void TitState::timerCallback()
{
    jassert (juce::MessageManager::getInstance()->isThisTheMessageThread());
    flush();
}

bool TitState::flush() noexcept
{
    bool changed { false };
    if (needsFlush.exchange (false, std::memory_order_acquire))
    {
        const bool envChanged     { flushEnv() };
        const bool repoChanged    { flushRepo() };
        const bool stringsChanged { flushStrings() };
        changed = envChanged or repoChanged or stringsChanged;
    }
    return changed;
}
