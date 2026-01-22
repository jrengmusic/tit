package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
// - *State: Complete git state representation
// - error: Any error encountered during detection
//
// CONTRACT: This is MANDATORY precomputation - no on-the-fly state detection
// in the application. All git state must flow through this function.

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
	workingTree, err := detectWorkingTree()
	if err != nil {
		return nil, fmt.Errorf("detecting working tree: %w", err)
	}
	state.WorkingTree = workingTree

	// Detect operation state (determines if timeline is applicable)
	operation, err := detectOperation()
	if err != nil {
		return nil, fmt.Errorf("detecting operation: %w", err)
	}
	state.Operation = operation

	// Detect remote presence (determines if timeline is applicable)
	remote, err := detectRemote()
	if err != nil {
		return nil, fmt.Errorf("detecting remote: %w", err)
	}
	state.Remote = remote

	// Detect timeline state (CONDITIONAL: only when on branch with tracking)
	// Timeline = comparison between local vs remote tracking branch
	// Not applicable when: Operation != Normal OR Remote = NoRemote
	if state.Operation == Normal && state.Remote == HasRemote && hasCommits {
		timeline, ahead, behind, err := detectTimeline()
		if err != nil {
			return nil, fmt.Errorf("detecting timeline: %w", err)
		}
		state.Timeline = timeline
		state.CommitsAhead = ahead
		state.CommitsBehind = behind
	} else {
		// Timeline N/A (no remote, no commits, or detached HEAD/time traveling)
		state.Timeline = ""
		state.CommitsAhead = 0
		state.CommitsBehind = 0
	}

	// Get current branch and commit hashes
	// Use symbolic-ref for branch name (works even with zero commits)
	branch, err := executeGitCommand("symbolic-ref", "--short", "HEAD")
	if err != nil {
		// symbolic-ref fails when HEAD is detached
		state.Detached = true
		// rev-parse returns "HEAD" literally when detached
		branch, err = executeGitCommand("rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("detecting current branch: %w", err)
		}
	}
	state.CurrentBranch = branch

	hash, err = executeGitCommand("rev-parse", "HEAD")
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

	outputStr := string(output)
	if outputStr == "" {
		// Empty output = clean working tree
		return Clean, nil
	}

	lines := strings.Split(outputStr, "\n")
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
			return Dirty, nil
		}
	}

	return Clean, nil
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

	// Priority 2: Check for time traveling (TIT-specific)
	gitDir := ".git"
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

// detectDirtyOperation checks if a dirty operation is in progress
// by looking for the .git/TIT_DIRTY_OP snapshot file
func detectDirtyOperation() bool {
	return IsDirtyOperationActive()
}

// GetTimeTravelInfo reads the .git/TIT_TIME_TRAVEL file and returns the original branch
// Returns: originalBranch, stashID, error
func GetTimeTravelInfo() (string, string, error) {
	gitDir := ".git"
	filePath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read time travel info: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 1 {
		return "", "", fmt.Errorf("invalid time travel info format")
	}

	originalBranch := strings.TrimSpace(lines[0])
	stashID := ""
	if len(lines) >= 2 {
		stashID = strings.TrimSpace(lines[1])
	}

	return originalBranch, stashID, nil
}

// WriteTimeTravelInfo writes the .git/TIT_TIME_TRAVEL file with original branch and optional stash ID
func WriteTimeTravelInfo(originalBranch, stashID string) error {
	gitDir := ".git"
	filePath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	content := originalBranch + "\n"
	if stashID != "" {
		content += stashID + "\n"
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write time travel info: %w", err)
	}

	return nil
}

// ClearTimeTravelInfo removes the .git/TIT_TIME_TRAVEL file
func ClearTimeTravelInfo() error {
	gitDir := ".git"
	filePath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear time travel info: %w", err)
	}

	return nil
}

// LoadTimeTravelInfo loads time travel metadata from .git/TIT_TIME_TRAVEL
// Returns nil if marker doesn't exist (normal case)
func LoadTimeTravelInfo() (*TimeTravelInfo, error) {
	gitDir := ".git"
	markerPath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	// Check if marker exists
	if !FileExists(markerPath) {
		return nil, nil
	}

	// Read marker file as plain text (simpler format)
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read time travel marker: %w", err)
	}

	// Parse two lines: branch and optional stash ID
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("time travel marker is empty")
	}

	branch := strings.TrimSpace(lines[0])
	stashID := ""
	if len(lines) > 1 {
		stashID = strings.TrimSpace(lines[1])
	}

	// Get current commit info (we're in detached HEAD during time travel)
	currentHash, err := executeGitCommand("rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit hash: %w", err)
	}
	currentHash = strings.TrimSpace(currentHash)

	// Get current commit metadata (subject and time)
	currentSubject, err := executeGitCommand("log", "-1", "--format=%s", currentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit subject: %w", err)
	}
	currentSubject = strings.TrimSpace(currentSubject)

	currentTimeStr, err := executeGitCommand("log", "-1", "--format=%aI", currentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit time: %w", err)
	}
	currentTime, err := time.Parse(time.RFC3339, strings.TrimSpace(currentTimeStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit time: %w", err)
	}

	return &TimeTravelInfo{
		OriginalBranch:  branch,
		OriginalStashID: stashID,
		CurrentCommit: CommitInfo{
			Hash:    currentHash,
			Subject: currentSubject,
			Time:    currentTime,
		},
	}, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
