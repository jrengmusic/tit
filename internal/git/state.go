package git

import (
	"os"
	"path/filepath"
	"strings"
	"tit/internal"
)

// DetectState performs comprehensive 5-axis git state detection for the current repository.
// This is the single source of truth for all git state information in TIT.
//
// The function detects:
// 1. Working Tree State: Clean, Dirty, or Untracked files
// 2. Timeline State: Current branch and commit history
// 3. Operation State: Current git operation (if any)
// 4. Remote State: Remote repository configuration
// 5. Environment State: Git installation and configuration
//
// This function is called frequently and must be fast. It uses cached results
// where possible and only executes git commands when necessary.
//
// Returns:
// - *State: Complete git state representation (always valid, never nil)
// - error: Always nil (function never fails - uses graceful fallbacks)
//
// CONTRACT: This is MANDATORY precomputation - no on-the-fly state detection
// in the application. All git state must flow through this function.
// CONTRACT: Never returns error - all system-level failures use graceful fallbacks.

func DetectState() (*State, error) {
	state := &State{}

	// PRIORITY CHECK: DirtyOperation trumps everything except NotRepo
	isRepo, _ := IsInitializedRepo()
	if !isRepo {
		return &State{Operation: NotRepo}, nil
	}

	// Check for dirty operation in progress (PRIORITY 1: Before time travel check)
	if IsDirtyOperationActive() {
		return &State{
			Operation: DirtyOperation,
		}, nil
	}

	// Check if repo has any commits yet
	hash, err := executeGitCommand("rev-parse", "HEAD")
	hasCommits := err == nil && hash != ""

	// Detect working tree state (always applicable)
	// Graceful fallback: assume Clean if git status fails
	workingTree, modifiedCount, err := detectWorkingTree()
	if err != nil {
		state.WorkingTree = Clean // Default to Clean on system-level failure
		state.ModifiedCount = 0
	} else {
		state.WorkingTree = workingTree
		state.ModifiedCount = modifiedCount
	}

	// Detect operation state (determines if timeline is applicable)
	// Graceful fallback: assume Normal if detection fails
	operation, err := detectOperation()
	if err != nil {
		state.Operation = Normal // Default to Normal on system-level failure
	} else {
		state.Operation = operation
	}

	// Detect remote presence (determines if timeline is applicable)
	// Graceful fallback: assume NoRemote if detection fails
	remote, err := detectRemote()
	if err != nil {
		state.Remote = NoRemote // Default to NoRemote on system-level failure
	} else {
		state.Remote = remote
	}

	// Detect timeline state (CONDITIONAL: only when on branch with tracking)
	// Timeline = comparison between local vs remote tracking branch
	// Not applicable when: Operation != Normal OR Remote = NoRemote
	if state.Operation == Normal && state.Remote == HasRemote && hasCommits {
		timeline, ahead, behind, err := detectTimeline()
		if err != nil {
			// Graceful fallback: assume InSync if timeline detection fails
			state.Timeline = InSync
			state.CommitsAhead = 0
			state.CommitsBehind = 0
		} else {
			state.Timeline = timeline
			state.CommitsAhead = ahead
			state.CommitsBehind = behind
		}
	} else {
		// Timeline N/A (no remote, no commits, or detached HEAD/time traveling)
		state.Timeline = ""
		state.CommitsAhead = 0
		state.CommitsBehind = 0
	}

	// Get current commit hash FIRST (needed for detached HEAD display)
	hash, _ = executeGitCommand("rev-parse", "--short", "HEAD")
	if hash != "" {
		state.CurrentHash = hash
	}

	// Get current branch
	// Use symbolic-ref for branch name (works even with zero commits)
	branch, err := executeGitCommand("symbolic-ref", "--short", "HEAD")
	if err != nil {
		// symbolic-ref fails when HEAD is detached
		state.Detached = true

		// Check if this is TIT-initiated time travel
		gitDir := internal.GitDirectoryName
		if _, statErr := os.Stat(filepath.Join(gitDir, "TIT_TIME_TRAVEL")); statErr == nil {
			state.IsTitTimeTravel = true
			// TIT time travel: get original branch from marker for display
			if data, err := os.ReadFile(filepath.Join(gitDir, "TIT_TIME_TRAVEL")); err == nil {
				lines := strings.Split(strings.TrimSpace(string(data)), "\n")
				if len(lines) > 0 && lines[0] != "" {
					state.CurrentBranch = lines[0] // Show original branch name
				} else {
					state.CurrentBranch = "DETACHED"
				}
			} else {
				state.CurrentBranch = "DETACHED"
			}
		} else {
			// Manual detached: CurrentBranch will be set in app_view.go
			state.CurrentBranch = "DETACHED"
		}

		// SSOT: detached HEAD uses TimeTraveling operation (for correct menu)
		// Menu shows browse history + return, regardless of cause
		state.Operation = TimeTraveling
	} else {
		state.CurrentBranch = branch
	}

	if state.Remote == HasRemote {
		remoteHash, err := executeGitCommand("rev-parse", "@{u}")
		if err != nil {
			// Remote tracking not set up yet (expected for new branches)
			remoteHash = ""
		}
		state.RemoteHash = remoteHash
		state.LocalBranchOnRemote = CurrentBranchExistsOnRemote()
	}

	return state, nil
}
