package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"tit/internal"
)

// detectWorkingTree checks for staged/unstaged changes or untracked files
// Returns Clean as fallback if git status fails (system-level issue)
func detectWorkingTree() (WorkingTree, int, error) {
	cmd := exec.Command("git", "status", "--porcelain=v2")
	output, err := cmd.Output()
	if err != nil {
		return Clean, 0, nil // Graceful fallback: assume Clean on system-level failure
	}

	outputStr := string(output)

	if outputStr == "" {
		return Clean, 0, nil
	}

	lines := strings.Split(outputStr, "\n")
	modifiedCount := 0

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		// Skip untracked ignored files (.DS_Store, etc)
		if line[0] == '!' {
			continue
		}
		// Lines starting with '1', '2' (changes) or '?' (untracked) indicate modifications
		if line[0] == '1' || line[0] == '2' || line[0] == '?' {
			modifiedCount++
		}
	}

	if modifiedCount > 0 {
		return Dirty, modifiedCount, nil
	}

	return Clean, 0, nil
}

// detectTimeline checks relationship between local and remote branches
func detectTimeline() (Timeline, int, int, error) {
	// PRECONDITION: Only called when Remote = HasRemote (checked by DetectState)
	// Timeline = comparison between local branch vs remote tracking branch

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
// Returns Normal as fallback if detection fails (system-level issue)
func detectOperation() (Operation, error) {
	// Priority 1: Check for conflicts FIRST (highest priority)
	cmd := exec.Command("git", "status", "--porcelain=v2")
	output, err := cmd.Output()
	if err != nil {
		return Normal, nil // Graceful fallback: assume Normal on system-level failure
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "u ") {
			return Conflicted, nil
		}
	}

	// Priority 2: Check for time traveling (TIT-specific)
	gitDir := internal.GitDirectoryName
	if _, err := os.Stat(filepath.Join(gitDir, "TIT_TIME_TRAVEL")); err == nil {
		return TimeTraveling, nil
	}

	// Priority 3: Check for ongoing operations
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
// Returns NoRemote as fallback if detection fails (system-level issue)
func detectRemote() (Remote, error) {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return NoRemote, nil // Graceful fallback: assume NoRemote on system-level failure
	}

	if strings.TrimSpace(string(output)) == "" {
		return NoRemote, nil
	}
	return HasRemote, nil
}
