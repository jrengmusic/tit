#pragma once
#include <juce_core/juce_core.h>

// SPEC §3 — five state axes

    enum class GitEnvironment { Ready, NeedsSetup, MissingGit, MissingSSH };
    enum class WorkingTree    { Clean, Dirty };
    enum class Timeline       { Empty, InSync, Ahead, Behind, Diverged };
    enum class Operation      { NotRepo, Normal, Merging, Conflicted, Rebasing,
                                TimeTraveling, DirtyOperation, Rewinding };
    enum class Remote         { NoRemote, HasRemote };

    // RFC §4.6 SETUP.phase values
    enum class SetupPhase     { EnvCheck, SSHKeyEntry, KeyGen, Display, GitConfig, Done };

    // String <-> enum bridging (VT stores strings per RFC §4.6)
    juce::String toString (GitEnvironment) noexcept;
    juce::String toString (WorkingTree)    noexcept;
    juce::String toString (Timeline)       noexcept;
    juce::String toString (Operation)      noexcept;
    juce::String toString (Remote)         noexcept;
    juce::String toString (SetupPhase)     noexcept;

    GitEnvironment parseGitEnvironment (const juce::String&) noexcept;
    WorkingTree    parseWorkingTree    (const juce::String&) noexcept;
    Timeline       parseTimeline       (const juce::String&) noexcept;
    Operation      parseOperation      (const juce::String&) noexcept;
    Remote         parseRemote         (const juce::String&) noexcept;
    SetupPhase     parseSetupPhase     (const juce::String&) noexcept;
