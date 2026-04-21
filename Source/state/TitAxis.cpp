#include "TitAxis.h"
#include <unordered_map>

juce::String toString (GitEnvironment v) noexcept
    {
        static const std::unordered_map<GitEnvironment, juce::String> table
        {
            { GitEnvironment::Ready,       "Ready" },
            { GitEnvironment::NeedsSetup,  "NeedsSetup" },
            { GitEnvironment::MissingGit,  "MissingGit" },
            { GitEnvironment::MissingSSH,  "MissingSSH" },
        };
        return table.at (v);
    }

    juce::String toString (WorkingTree v) noexcept
    {
        static const std::unordered_map<WorkingTree, juce::String> table
        {
            { WorkingTree::Clean, "Clean" },
            { WorkingTree::Dirty, "Dirty" },
        };
        return table.at (v);
    }

    juce::String toString (Timeline v) noexcept
    {
        static const std::unordered_map<Timeline, juce::String> table
        {
            { Timeline::Empty,    "" },
            { Timeline::InSync,   "InSync" },
            { Timeline::Ahead,    "Ahead" },
            { Timeline::Behind,   "Behind" },
            { Timeline::Diverged, "Diverged" },
        };
        return table.at (v);
    }

    juce::String toString (Operation v) noexcept
    {
        static const std::unordered_map<Operation, juce::String> table
        {
            { Operation::NotRepo,        "NotRepo" },
            { Operation::Normal,         "Normal" },
            { Operation::Merging,        "Merging" },
            { Operation::Conflicted,     "Conflicted" },
            { Operation::Rebasing,       "Rebasing" },
            { Operation::TimeTraveling,  "TimeTraveling" },
            { Operation::DirtyOperation, "DirtyOperation" },
            { Operation::Rewinding,      "Rewinding" },
        };
        return table.at (v);
    }

    juce::String toString (Remote v) noexcept
    {
        static const std::unordered_map<Remote, juce::String> table
        {
            { Remote::NoRemote,  "NoRemote" },
            { Remote::HasRemote, "HasRemote" },
        };
        return table.at (v);
    }

    juce::String toString (SetupPhase v) noexcept
    {
        static const std::unordered_map<SetupPhase, juce::String> table
        {
            { SetupPhase::EnvCheck,   "EnvCheck" },
            { SetupPhase::SSHKeyEntry,"SSHKeyEntry" },
            { SetupPhase::KeyGen,     "KeyGen" },
            { SetupPhase::Display,    "Display" },
            { SetupPhase::GitConfig,  "GitConfig" },
            { SetupPhase::Done,       "Done" },
        };
        return table.at (v);
    }

    // ---------------------------------------------------------------------------

    GitEnvironment parseGitEnvironment (const juce::String& s) noexcept
    {
        static const std::unordered_map<juce::String, GitEnvironment> table
        {
            { "Ready",      GitEnvironment::Ready },
            { "NeedsSetup", GitEnvironment::NeedsSetup },
            { "MissingGit", GitEnvironment::MissingGit },
            { "MissingSSH", GitEnvironment::MissingSSH },
        };
        jassert (table.count (s) > 0);
        return table.at (s);
    }

    WorkingTree parseWorkingTree (const juce::String& s) noexcept
    {
        static const std::unordered_map<juce::String, WorkingTree> table
        {
            { "Clean", WorkingTree::Clean },
            { "Dirty", WorkingTree::Dirty },
        };
        jassert (table.count (s) > 0);
        return table.at (s);
    }

    Timeline parseTimeline (const juce::String& s) noexcept
    {
        static const std::unordered_map<juce::String, Timeline> table
        {
            { "",         Timeline::Empty },
            { "InSync",   Timeline::InSync },
            { "Ahead",    Timeline::Ahead },
            { "Behind",   Timeline::Behind },
            { "Diverged", Timeline::Diverged },
        };
        jassert (table.count (s) > 0);
        return table.at (s);
    }

    Operation parseOperation (const juce::String& s) noexcept
    {
        static const std::unordered_map<juce::String, Operation> table
        {
            { "NotRepo",        Operation::NotRepo },
            { "Normal",         Operation::Normal },
            { "Merging",        Operation::Merging },
            { "Conflicted",     Operation::Conflicted },
            { "Rebasing",       Operation::Rebasing },
            { "TimeTraveling",  Operation::TimeTraveling },
            { "DirtyOperation", Operation::DirtyOperation },
            { "Rewinding",      Operation::Rewinding },
        };
        jassert (table.count (s) > 0);
        return table.at (s);
    }

    Remote parseRemote (const juce::String& s) noexcept
    {
        static const std::unordered_map<juce::String, Remote> table
        {
            { "NoRemote",  Remote::NoRemote },
            { "HasRemote", Remote::HasRemote },
        };
        jassert (table.count (s) > 0);
        return table.at (s);
    }

    SetupPhase parseSetupPhase (const juce::String& s) noexcept
    {
        static const std::unordered_map<juce::String, SetupPhase> table
        {
            { "EnvCheck",    SetupPhase::EnvCheck },
            { "SSHKeyEntry", SetupPhase::SSHKeyEntry },
            { "KeyGen",      SetupPhase::KeyGen },
            { "Display",     SetupPhase::Display },
            { "GitConfig",   SetupPhase::GitConfig },
            { "Done",        SetupPhase::Done },
        };
        jassert (table.count (s) > 0);
        return table.at (s);
    }
