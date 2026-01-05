package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// IsInitializedRepo checks if current working directory is a git repository
func IsInitializedRepo() (bool, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, ""
	}

	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true, cwd
	}

	return false, ""
}

// HasParentRepo checks if any parent directory contains .git
func HasParentRepo() (bool, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, ""
	}

	parent := filepath.Dir(cwd)
	for parent != cwd {
		gitDir := filepath.Join(parent, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			return true, parent
		}

		cwd = parent
		parent = filepath.Dir(parent)
	}

	return false, ""
}

// DetectState performs full state detection: (WorkingTree, Timeline, Operation, Remote)
func DetectState() (*State, error) {
	state := &State{}

	// Detect working tree state
	workingTree, err := detectWorkingTree()
	if err != nil {
		return nil, fmt.Errorf("detecting working tree: %w", err)
	}
	state.WorkingTree = workingTree

	// Detect timeline state
	timeline, ahead, behind, err := detectTimeline()
	if err != nil {
		return nil, fmt.Errorf("detecting timeline: %w", err)
	}
	state.Timeline = timeline
	state.CommitsAhead = ahead
	state.CommitsBehind = behind

	// Detect operation state
	operation, err := detectOperation()
	if err != nil {
		return nil, fmt.Errorf("detecting operation: %w", err)
	}
	state.Operation = operation

	// Detect remote presence
	remote, err := detectRemote()
	if err != nil {
		return nil, fmt.Errorf("detecting remote: %w", err)
	}
	state.Remote = remote

	// Get current branch and commit hashes
	// Use symbolic-ref for branch name (works even with zero commits)
	branch, err := executeGitCommand("symbolic-ref", "--short", "HEAD")
	if err != nil {
		// If symbolic-ref fails (detached HEAD), try rev-parse
		branch, err = executeGitCommand("rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("detecting current branch: %w", err)
		}
	}
	state.CurrentBranch = branch

	hash, err := executeGitCommand("rev-parse", "HEAD")
	if err != nil {
		// No commits yet (empty repo after init) - this is normal
		state.CurrentHash = ""
	} else {
		state.CurrentHash = hash
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

// detectWorkingTree checks for staged/unstaged changes or untracked files
func detectWorkingTree() (WorkingTree, error) {
	cmd := exec.Command("git", "status", "--porcelain=v2")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git status failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// Lines starting with '1', '2' (changes) or '?' (untracked) indicate modifications
		if line[0] == '1' || line[0] == '2' || line[0] == '?' {
			return Modified, nil
		}
	}

	return Clean, nil
}

// detectTimeline checks relationship between local and remote branches
func detectTimeline() (Timeline, int, int, error) {
	// Check if remote exists first
	remoteStatus, _ := detectRemote()
	if remoteStatus == NoRemote {
		return TimelineNoRemote, 0, 0, nil
	}

	// Try to get upstream tracking branch
	cmd := exec.Command("git", "rev-parse", "@{u}")
	err := cmd.Run()
	if err != nil {
		// No upstream tracking - try to compare with refs/remotes/origin/[current-branch]
		// Use symbolic-ref first (works in empty repos), fall back to rev-parse
		currentBranch, err := executeGitCommand("symbolic-ref", "--short", "HEAD")
		if err != nil {
			currentBranch, err = executeGitCommand("rev-parse", "--abbrev-ref", "HEAD")
			if err != nil {
				return InSync, 0, 0, nil
			}
		}
		if currentBranch == "" || currentBranch == "HEAD" {
			return InSync, 0, 0, nil
		}

		// Use full ref path to avoid ambiguity
		remoteBranch := "refs/remotes/origin/" + currentBranch
		checkRemoteCmd := exec.Command("git", "rev-parse", remoteBranch)
		err = checkRemoteCmd.Run()
		if err != nil {
			// Remote branch doesn't exist yet (never pushed)
			cmd := exec.Command("git", "rev-list", "--count", "HEAD")
			output, err := cmd.Output()
			if err != nil {
				return InSync, 0, 0, nil
			}
			count, err := strconv.Atoi(strings.TrimSpace(string(output)))
			if err != nil {
				return InSync, 0, 0, nil
			}
			if count > 0 {
				return Ahead, count, 0, nil
			}
			return InSync, 0, 0, nil
		}

		// Remote branch exists - compare HEAD with refs/remotes/origin/[branch]
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD..."+remoteBranch)
		output, err := cmd.Output()
		if err != nil {
			return InSync, 0, 0, nil
		}
		parts := strings.Fields(strings.TrimSpace(string(output)))
		if len(parts) != 2 {
			return InSync, 0, 0, nil
		}

		ahead, err := strconv.Atoi(parts[0])
		if err != nil {
			return InSync, 0, 0, nil
		}
		behind, err := strconv.Atoi(parts[1])
		if err != nil {
			return InSync, 0, 0, nil
		}
		return determineTimeline(ahead, behind), ahead, behind, nil
	}

	// We have upstream tracking - count commits
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	output, err := cmd.Output()
	if err != nil {
		return InSync, 0, 0, nil
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return InSync, 0, 0, nil
	}

	ahead, err := strconv.Atoi(parts[0])
	if err != nil {
		return InSync, 0, 0, nil
	}
	behind, err := strconv.Atoi(parts[1])
	if err != nil {
		return InSync, 0, 0, nil
	}

	return determineTimeline(ahead, behind), ahead, behind, nil
}

// determineTimeline maps commit counts to timeline state
func determineTimeline(ahead, behind int) Timeline {
	stateMap := map[[2]bool]Timeline{
		{false, false}: InSync,   // ahead == 0 && behind == 0
		{true, false}:  Ahead,    // ahead > 0 && behind == 0
		{false, true}:  Behind,   // ahead == 0 && behind > 0
		{true, true}:   Diverged, // ahead > 0 && behind > 0
	}

	key := [2]bool{ahead > 0, behind > 0}
	if timeline, exists := stateMap[key]; exists {
		return timeline
	}
	return InSync // fallback
}

// detectOperation checks for merge/rebase/conflict/cherry-pick
func detectOperation() (Operation, error) {
	// Priority 1: Check for conflicts FIRST (highest priority)
	cmd := exec.Command("git", "status", "--porcelain=v2")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("checking for conflicts: %w", err)
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "u ") {
			return Conflicted, nil
		}
	}

	// Priority 2: Check for ongoing operations
	gitDir := ".git"

	// Check for merge in progress
	if _, err := os.Stat(filepath.Join(gitDir, "MERGE_HEAD")); err == nil {
		return Merging, nil
	}

	// Check for rebase in progress
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil {
		return Rebasing, nil
	}
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil {
		return Rebasing, nil
	}

	return Normal, nil
}

// detectRemote checks if remote exists
func detectRemote() (Remote, error) {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("detecting remote: %w", err)
	}

	if strings.TrimSpace(string(output)) == "" {
		return NoRemote, nil
	}
	return HasRemote, nil
}

// executeGitCommand runs git command and returns trimmed output or error
func executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CurrentBranchExistsOnRemote checks if current branch exists on remote
func CurrentBranchExistsOnRemote() bool {
	currentBranch, err := executeGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false // Can't determine branch, assume doesn't exist on remote
	}
	remoteBranch := "refs/remotes/origin/" + currentBranch
	cmd := exec.Command("git", "rev-parse", remoteBranch)
	return cmd.Run() == nil
}
