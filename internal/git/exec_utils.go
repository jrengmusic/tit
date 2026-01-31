package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tit/internal"
)

// FindStashRefByHash finds the stash reference (stash@{N}) for a given hash
// Returns (stashRef, true) if found, ("", false) if not found
// Caller should handle not-found case gracefully (don't panic)
func FindStashRefByHash(targetHash string) (string, bool) {
	// Iterate through stash@{0}..{9} looking for matching hash
	// Limit to internal.StashSearchLimit stashes (reasonable - if user has more, they have bigger problems)
	for i := 0; i < internal.StashSearchLimit; i++ {
		stashRef := fmt.Sprintf("stash@{%d}", i)
		result := Execute("rev-parse", stashRef)

		if !result.Success {
			// No more stashes exist at this index
			break
		}

		hash := strings.TrimSpace(result.Stdout)
		if hash == targetHash {
			return stashRef, true
		}
	}

	// Stash not found
	return "", false
}

// StashExists checks if a stash hash still exists in git
// Returns true if stash with given hash exists, false otherwise
func StashExists(stashHash string) bool {
	_, exists := FindStashRefByHash(stashHash)
	return exists
}

// cleanStaleLocks removes stale git lock files that can occur when git operations
// are interrupted or the git process crashes. This allows subsequent operations to proceed.
// Silent cleanup - if lock doesn't exist, that's fine.
func cleanStaleLocks() {
	repoPath, err := os.Getwd()
	if err != nil {
		return // Can't determine repo path, skip cleanup
	}

	lockPath := filepath.Join(repoPath, internal.GitDirectoryName, "index.lock")
	// Check if lock exists and delete it
	// This is safe: git creates it during operations and deletes on completion
	// If present, it means a previous operation was interrupted
	if _, err := os.Stat(lockPath); err == nil {
		if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
			Error(fmt.Sprintf("Warning: could not remove stale lock: %v", err))
		}
	}
}

// ExtractRepoName extracts repository name from a git URL
// Handles: https://github.com/user/repo.git, git@github.com:user/repo.git, etc.
// Returns just "repo" (without .git extension)
func ExtractRepoName(gitURL string) string {
	// Remove trailing .git if present
	name := strings.TrimSuffix(gitURL, ".git")

	// Get the last path component (repo name)
	name = filepath.Base(name)

	// Handle SSH URLs like git@github.com:user/repo
	if strings.Contains(name, "@") {
		parts := strings.Split(name, "@")
		name = parts[len(parts)-1]
	}

	return name
}

// GetRemoteURL returns the URL for the 'origin' remote
// Returns empty string if no remote configured
func GetRemoteURL() string {
	result := Execute("remote", "get-url", "origin")
	if result.Success {
		return strings.TrimSpace(result.Stdout)
	}
	return ""
}

// SetUpstreamTracking sets the upstream tracking branch to origin/[current-branch]
// Returns success even if remote branch doesn't exist yet (will be set on first push -u)
func SetUpstreamTracking() CommandResult {
	Log("Attempting to set upstream tracking...")

	// Get current branch name
	result := Execute("rev-parse", "--abbrev-ref", "HEAD")
	if !result.Success || result.Stdout == "" {
		return CommandResult{Success: false, Stderr: "Not on a branch"}
	}

	currentBranch := strings.TrimSpace(result.Stdout)

	// Use full ref path to avoid ambiguity with local branches named "origin/[branch]"
	remoteBranch := "refs/remotes/origin/" + currentBranch

	// Try to set upstream
	result = Execute("branch", "--set-upstream-to="+remoteBranch)

	return result
}

// SetUpstreamTrackingWithBranch sets upstream tracking using a provided branch name
// If remote branch doesn't exist (empty remote), pushes with -u to create it
func SetUpstreamTrackingWithBranch(ctx context.Context, branchName string) CommandResult {

	if branchName == "" {
		Error("WARNING: No branch name provided for upstream tracking")
		return CommandResult{Success: false, Stderr: "No branch name provided"}
	}

	// Use full ref path to avoid ambiguity with local branches named "origin/[branch]"
	remoteBranch := "refs/remotes/origin/" + branchName

	// Check if remote branch exists
	checkResult := Execute("rev-parse", "--verify", remoteBranch)

	if checkResult.Success {
		// Remote branch exists - set upstream tracking
		Log(fmt.Sprintf("Setting upstream for branch '%s'...", branchName))
		result := Execute("branch", "--set-upstream-to="+remoteBranch)
		if result.Success {
			Log(fmt.Sprintf("Upstream tracking set to %s", remoteBranch))
			return CommandResult{Success: true, Stdout: "upstream_set"}
		}
		// Failed to set upstream even though remote exists
		Error(fmt.Sprintf("Could not set upstream tracking: %s", result.Stderr))
		return CommandResult{Success: false, Stderr: result.Stderr}
	}

	// Remote branch doesn't exist - push with -u to create it and set upstream
	Log(fmt.Sprintf("Remote branch '%s' doesn't exist, pushing to create...", branchName))
	result := ExecuteWithStreaming(ctx, "push", "-u", "origin", branchName)
	if result.Success {
		Log(fmt.Sprintf("Created remote branch and set upstream tracking"))
		return CommandResult{Success: true, Stdout: "pushed_and_upstream_set"}
	}

	Error(fmt.Sprintf("Failed to push: %s", result.Stderr))
	return CommandResult{Success: false, Stderr: result.Stderr}
}
