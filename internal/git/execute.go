package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/config"
	"tit/internal/ui"
)

// CommandResult contains the output and exit code of a git command
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Success  bool
}

// Execute runs a git command and returns the result
// Does NOT stream to output buffer (for internal git queries)
func Execute(args ...string) CommandResult {
	cmd := exec.Command("git", args...)

	// CRITICAL: Disable interactive prompts - fail fast instead of hanging
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// Capture stdout and stderr separately for better error diagnostics
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return CommandResult{
			Stdout:   "",
			Stderr:   fmt.Sprintf("Failed to create stdout pipe: %v", err),
			ExitCode: 1,
			Success:  false,
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return CommandResult{
			Stdout:   "",
			Stderr:   fmt.Sprintf("Failed to create stderr pipe: %v", err),
			ExitCode: 1,
			Success:  false,
		}
	}
	
	if err := cmd.Start(); err != nil {
		return CommandResult{
			Stdout:   "",
			Stderr:   fmt.Sprintf("Failed to start command: %v", err),
			ExitCode: 1,
			Success:  false,
		}
	}

	// Read output
	var stdoutBuf, stderrBuf strings.Builder
	if _, copyErr := io.Copy(&stdoutBuf, stdout); copyErr != nil {
		return CommandResult{
			Stdout:   "",
			Stderr:   fmt.Sprintf("Failed to read stdout: %v", copyErr),
			ExitCode: 1,
			Success:  false,
		}
	}
	if _, copyErr := io.Copy(&stderrBuf, stderr); copyErr != nil {
		return CommandResult{
			Stdout:   "",
			Stderr:   fmt.Sprintf("Failed to read stderr: %v", copyErr),
			ExitCode: 1,
			Success:  false,
		}
	}
	
	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return CommandResult{
		Stdout:   strings.TrimSpace(stdoutBuf.String()),
		Stderr:   strings.TrimSpace(stderrBuf.String()),
		ExitCode: exitCode,
		Success:  exitCode == 0,
	}
}

// FindStashRefByHash finds the stash reference (stash@{N}) for a given hash
// Returns the stash reference if found, panics if not found (fail fast)
func FindStashRefByHash(targetHash string) string {
	// Iterate through stash@{0}..{9} looking for matching hash
	// Limit to 10 stashes (reasonable - if user has more, they have bigger problems)
	for i := 0; i < 10; i++ {
		stashRef := fmt.Sprintf("stash@{%d}", i)
		result := Execute("rev-parse", stashRef)

		if !result.Success {
			// No more stashes exist at this index
			break
		}

		hash := strings.TrimSpace(result.Stdout)
		if hash == targetHash {
			return stashRef
		}
	}

	// Stash not found - this should NEVER happen (indicates bug or user manually dropped stash)
	panic(fmt.Sprintf("FATAL: Stash with hash %s not found in stash list. This indicates the stash was manually dropped or a bug in stash tracking.", targetHash))
}

// ExecuteWithStreaming runs a git command and streams output to the buffer
// Use this for user-initiated actions that should display console output
// WORKER THREAD - Must be called from async operation
func ExecuteWithStreaming(args ...string) CommandResult {
	buffer := ui.GetBuffer()

	// Log the command being executed
	cmdString := "git " + strings.Join(args, " ")
	buffer.Append(cmdString, ui.TypeCommand)

	cmd := exec.Command("git", args...)

	// CRITICAL: Disable interactive prompts - fail fast instead of hanging
	// This prevents git from waiting for SSH passphrase, HTTP auth, etc.
	// User must have SSH keys or credential helpers configured
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		buffer.Append(fmt.Sprintf("Error creating stdout pipe: %v", err), ui.TypeStderr)
		return CommandResult{Success: false, ExitCode: 1, Stderr: err.Error()}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		buffer.Append(fmt.Sprintf("Error creating stderr pipe: %v", err), ui.TypeStderr)
		return CommandResult{Success: false, ExitCode: 1, Stderr: err.Error()}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		buffer.Append(fmt.Sprintf("Error starting command: %v", err), ui.TypeStderr)
		return CommandResult{Success: false, ExitCode: 1, Stderr: err.Error()}
	}

	// WaitGroup to ensure all output is captured before completion message
	var wg sync.WaitGroup

	// Stream stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		// CRITICAL: Use io.Copy instead of scanner to avoid blocking on \r without \n
		// Git progress sends \r (carriage return) without \n (newline)
		io.Copy(io.Discard, stdout)
		// Output is captured by stderr scanner - stdout from git is usually empty for clone
	}()

	// Stream stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		// CRITICAL: Read byte-by-byte to handle \r without \n from git progress
		var currentLine strings.Builder
		oneByte := make([]byte, 1)

		for {
			n, err := stderr.Read(oneByte)
			if n > 0 {
				ch := oneByte[0]
				if ch == '\n' || ch == '\r' {
					// End of line - process what we have
					line := strings.TrimSpace(currentLine.String())
					if line != "" {
						buffer.Append(line, ui.TypeStderr)
					}
					currentLine.Reset()
				} else {
					currentLine.WriteByte(ch)
				}
			}
			if err == io.EOF {
				// Final line (no trailing \n or \r)
				line := strings.TrimSpace(currentLine.String())
				if line != "" {
					buffer.Append(line, ui.TypeStderr)
				}
				break
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()

	// Wait for all output to be read from pipes
	wg.Wait()

	// Determine exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Log completion status
	if exitCode == 0 {
		buffer.Append("Command completed successfully", ui.TypeStatus)
	} else {
		// Don't say "failed" - exit code 1 can be expected (e.g., merge conflicts)
		buffer.Append(fmt.Sprintf("Command exited with code %d", exitCode), ui.TypeInfo)
	}

	return CommandResult{
		Stdout:   "", // Output already streamed to buffer
		Stderr:   "",
		ExitCode: exitCode,
		Success:  exitCode == 0,
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
	buffer := ui.GetBuffer()
	buffer.Append("Attempting to set upstream tracking...", ui.TypeStatus)

	// Get current branch name
	result := Execute("rev-parse", "--abbrev-ref", "HEAD")
	if !result.Success || result.Stdout == "" {
		buffer.Append("DEBUG: Failed to get current branch", ui.TypeStderr)
		return CommandResult{Success: false, Stderr: "Not on a branch"}
	}

	currentBranch := strings.TrimSpace(result.Stdout)
	buffer.Append(fmt.Sprintf("DEBUG: Current branch = '%s'", currentBranch), ui.TypeStatus)

	// Use full ref path to avoid ambiguity with local branches named "origin/[branch]"
	remoteBranch := "refs/remotes/origin/" + currentBranch
	buffer.Append(fmt.Sprintf("DEBUG: Remote branch = '%s'", remoteBranch), ui.TypeStatus)

	// Try to set upstream
	result = Execute("branch", "--set-upstream-to="+remoteBranch)
	buffer.Append(fmt.Sprintf("DEBUG: git branch result success=%v, stderr=%s", result.Success, result.Stderr), ui.TypeStatus)

	return result
}

// SetUpstreamTrackingWithBranch sets upstream tracking using a provided branch name
// This avoids querying git in the worker thread context
func SetUpstreamTrackingWithBranch(branchName string) CommandResult {
	buffer := ui.GetBuffer()
	
	if branchName == "" {
		buffer.Append("WARNING: No branch name provided for upstream tracking", ui.TypeStderr)
		return CommandResult{Success: false, Stderr: "No branch name provided"}
	}

	buffer.Append(fmt.Sprintf("Setting upstream for branch '%s'...", branchName), ui.TypeStatus)

	// Use full ref path to avoid ambiguity with local branches named "origin/[branch]"
	remoteBranch := "refs/remotes/origin/" + branchName

	// Try to set upstream
	result := Execute("branch", "--set-upstream-to="+remoteBranch)
	
	if result.Success {
		buffer.Append(fmt.Sprintf("Upstream tracking set to %s", remoteBranch), ui.TypeStatus)
	} else {
		// Not fatal - may be detached HEAD or other condition
		buffer.Append(fmt.Sprintf("Could not set upstream tracking: %s", result.Stderr), ui.TypeInfo)
	}

	// Always return success - if remote branch doesn't exist, first push -u will handle it
	return CommandResult{Success: true}
}

// ListConflictedFiles returns a list of files with merge conflicts
// Uses: git status --porcelain=v2 and filters for unmerged (u) entries
func ListConflictedFiles() ([]string, error) {
	result := Execute("status", "--porcelain=v2")
	if !result.Success {
		return nil, fmt.Errorf("failed to get status: %s", result.Stderr)
	}

	var conflictedFiles []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}
		
		// Status line format: <status> <meta> <meta> ... <path>
		// Unmerged is marked with 'u' in second field
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		if parts[0] == "u" { // Unmerged status
			// Path starts at field 10 for unmerged entries (after 3 hashes)
			// Format: u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
			if len(parts) >= 11 {
				// Join fields from index 10 onwards (handles filenames with spaces)
				path := strings.Join(parts[10:], " ")
				conflictedFiles = append(conflictedFiles, path)
			}
		}
	}

	if len(conflictedFiles) == 0 {
		return nil, fmt.Errorf("no conflicted files found")
	}

	return conflictedFiles, nil
}

// ShowConflictVersion retrieves the content of a specific stage from a merge conflict
// stage: 1=base, 2=local (ours), 3=remote (theirs)
// Returns the file content as a string
func ShowConflictVersion(filePath string, stage int) (string, error) {
	if stage < 1 || stage > 3 {
		return "", fmt.Errorf("invalid stage: %d (must be 1-3)", stage)
	}

	stageRef := fmt.Sprintf(":%d:%s", stage, filePath)
	result := Execute("show", stageRef)
	
	if !result.Success {
		return "", fmt.Errorf("failed to show stage %d of %s: %s", stage, filePath, result.Stderr)
	}

	return result.Stdout, nil
}

// FetchRecentCommits fetches the last N commits with basic info
// ref: git ref to log from (e.g., "main", "HEAD"). Empty string = HEAD (current position)
// Returns: []CommitInfo with Hash, Subject, Time populated
// Format: git log --pretty=%H%n%s%n%ai -N [ref]
// Example output: hash1\nsubject1\nISO-date1\nhash2\nsubject2\nISO-date2\n...
func FetchRecentCommits(limit int, ref string) ([]CommitInfo, error) {
	args := []string{"log", fmt.Sprintf("-%d", limit), "--pretty=%H%n%s%n%ai"}
	if ref != "" {
		args = append(args, ref)
	}
	result := Execute(args...)
	if !result.Success {
		return nil, fmt.Errorf("failed to fetch recent commits: %s", result.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	commits := make([]CommitInfo, 0)

	// Parse output: hash, subject, iso-date on consecutive lines
	for i := 0; i+2 < len(lines); i += 3 {
		hash := strings.TrimSpace(lines[i])
		subject := strings.TrimSpace(lines[i+1])
		dateStr := strings.TrimSpace(lines[i+2])

		if hash == "" {
			continue
		}

		// Parse ISO date format: YYYY-MM-DD HH:MM:SS ±HHMM
		parsedTime, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			// If parsing fails, use zero time but continue
			parsedTime = time.Time{}
		}

		commits = append(commits, CommitInfo{
			Hash:    hash,
			Subject: subject,
			Time:    parsedTime,
		})
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	return commits, nil
}

// GetCommitDetails fetches full metadata for a commit
// Returns: CommitDetails with Author, Date, Message
// Format: git show -s --pretty=%aN%n%aD%n%B <hash>
func GetCommitDetails(hash string) (*CommitDetails, error) {
	result := Execute("show", "-s", "--pretty=%aN%n%aD%n%B", hash)
	if !result.Success {
		return nil, fmt.Errorf("failed to get commit details: %s", result.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected commit details format")
	}

	// Parse: author name (line 0), date (line 1), message (lines 2+)
	author := strings.TrimSpace(lines[0])
	date := strings.TrimSpace(lines[1])
	message := strings.Join(lines[2:], "\n")

	return &CommitDetails{
		Author:  author,
		Date:    date,
		Message: strings.TrimSpace(message),
	}, nil
}

// GetFilesInCommit fetches list of files changed in a commit
// Returns: []FileInfo with Path and Status
// Format: git show --name-status --pretty= <hash>
// Status: M (modified), A (added), D (deleted), R (renamed), C (copied), T (type changed), U (unmerged)
func GetFilesInCommit(hash string) ([]FileInfo, error) {
	result := Execute("show", "--name-status", "--pretty=", hash)
	if !result.Success {
		return nil, fmt.Errorf("failed to get files in commit: %s", result.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	files := make([]FileInfo, 0)

	// Parse output: status\tpath (tab-separated)
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}

		status := strings.TrimSpace(parts[0])
		path := strings.TrimSpace(parts[1])

		// Rename/copy format: "R100\told\tnew" → extract just status char
		statusChar := string(status[0])

		files = append(files, FileInfo{
			Path:   path,
			Status: statusChar,
		})
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in commit")
	}

	return files, nil
}

// GetCommitDiff fetches diff for a file in a commit
// version: "parent" (commit vs parent) or "wip" (commit vs working tree)
// Returns: unified diff content (plain text)
// Uses git diff with appropriate range based on version
func GetCommitDiff(hash, path, version string) (string, error) {
	var result CommandResult

	switch version {
	case "parent":
		// Compare commit with its parent
		result = Execute("diff", hash+"^", hash, "--", path)
	case "wip":
		// Compare commit with working tree
		result = Execute("diff", hash, "--", path)
	default:
		return "", fmt.Errorf("invalid diff version: %s (must be 'parent' or 'wip')", version)
	}

	if !result.Success {
		return "", fmt.Errorf("failed to get diff for %s: %s", path, result.Stderr)
	}

	return result.Stdout, nil
}

// ExecuteTimeTravelCheckout performs a time travel checkout operation
// Creates .git/TIT_TIME_TRAVEL file with original branch info
// Returns: TimeTravelCheckoutMsg
func ExecuteTimeTravelCheckout(originalBranch, commitHash string) func() tea.Msg {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf("Time traveling to commit %s...", commitHash), ui.TypeStatus)

		// Get current branch (for verification)
		currentBranchResult := Execute("rev-parse", "--abbrev-ref", "HEAD")
		if !currentBranchResult.Success {
			buffer.Append(fmt.Sprintf("Error getting current branch: %s", currentBranchResult.Stderr), ui.TypeStderr)
			return TimeTravelCheckoutMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				CommitHash:     commitHash,
				Error:          fmt.Sprintf("Failed to get current branch: %s", currentBranchResult.Stderr),
			}
		}

		// Checkout the target commit
		checkoutResult := Execute("checkout", commitHash)
		if !checkoutResult.Success {
			buffer.Append(fmt.Sprintf("Error checking out commit: %s", checkoutResult.Stderr), ui.TypeStderr)
			return TimeTravelCheckoutMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				CommitHash:     commitHash,
				Error:          fmt.Sprintf("Failed to checkout commit: %s", checkoutResult.Stderr),
			}
		}

		// Write time travel info file
		err := WriteTimeTravelInfo(originalBranch, "")
		if err != nil {
			buffer.Append(fmt.Sprintf("Error writing time travel info: %v", err), ui.TypeStderr)
			// Try to checkout back to original branch
			Execute("checkout", originalBranch)
			return TimeTravelCheckoutMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				CommitHash:     commitHash,
				Error:          fmt.Sprintf("Failed to write time travel info: %v", err),
			}
		}

		buffer.Append("Time travel successful", ui.TypeStatus)
		return TimeTravelCheckoutMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			CommitHash:     commitHash,
			Error:          "",
		}
	}
}

// ExecuteTimeTravelMerge performs a merge of time travel changes back to original branch
// Returns: TimeTravelMergeMsg
func ExecuteTimeTravelMerge(originalBranch, timeTravelHash string) func() tea.Msg {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf("[DEBUG] === START ExecuteTimeTravelMerge ===", originalBranch), ui.TypeStatus)
		buffer.Append(fmt.Sprintf("[DEBUG] originalBranch=%s, timeTravelHash=%s", originalBranch, timeTravelHash), ui.TypeStatus)

		// Get current working directory (repo path)
		repoPath, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
		}
		buffer.Append(fmt.Sprintf("[DEBUG] repoPath=%s", repoPath), ui.TypeStatus)

		// Get original stash hash (if any) from config tracking system
		originalStashHash, hasStash := config.FindStashEntry("time_travel", repoPath)
		if hasStash {
			buffer.Append(fmt.Sprintf("[DEBUG] Found stash in config - hash: %s", originalStashHash), ui.TypeStatus)
		} else {
			buffer.Append("[DEBUG] No stash found in config (tree was clean before time travel)", ui.TypeStatus)
		}

		// Note: Dirty tree handling is done BEFORE calling this function
		// User either committed changes or discarded them, so tree is now clean

		// Checkout original branch
		buffer.Append(fmt.Sprintf("[DEBUG] About to checkout %s...", originalBranch), ui.TypeStatus)
		checkoutResult := Execute("checkout", originalBranch)
		buffer.Append(fmt.Sprintf("[DEBUG] Checkout result: Success=%v, Stderr=%s", checkoutResult.Success, checkoutResult.Stderr), ui.TypeStatus)
		if !checkoutResult.Success {
			buffer.Append(fmt.Sprintf("[DEBUG] EARLY RETURN: Checkout failed", checkoutResult.Stderr), ui.TypeStderr)
			return TimeTravelMergeMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				TimeTravelHash: timeTravelHash,
				Error:          fmt.Sprintf("Failed to checkout original branch: %s", checkoutResult.Stderr),
			}
		}

		// =================================================================
		// PHASE 1: Merge time travel changes FIRST (before restoring stash)
		// This provides cleaner separation - merge conflicts and stash conflicts are handled separately
		// =================================================================
		buffer.Append("[DEBUG] === PHASE 1: Merge ===", ui.TypeStatus)
		buffer.Append(fmt.Sprintf("[DEBUG] About to merge %s...", timeTravelHash), ui.TypeStatus)

		mergeResult := Execute("merge", timeTravelHash)
		buffer.Append(fmt.Sprintf("[DEBUG] Merge result: Success=%v, Stdout=%s, Stderr=%s", mergeResult.Success, mergeResult.Stdout, mergeResult.Stderr), ui.TypeStatus)

		if !mergeResult.Success {
			buffer.Append(fmt.Sprintf("[DEBUG] Merge returned Success=false, checking for conflicts..."), ui.TypeInfo)

			// Check for conflicts
			conflictFiles, err := ListConflictedFiles()
			buffer.Append(fmt.Sprintf("[DEBUG] ListConflictedFiles: err=%v, files=%v", err, conflictFiles), ui.TypeStatus)

			if err != nil {
				// No conflicts detected - this is a merge error (not conflicts)
				buffer.Append("[DEBUG] BRANCH: No conflicts detected (merge error)", ui.TypeStatus)
				buffer.Append(fmt.Sprintf("[DEBUG] EARLY RETURN: Merge failed without conflicts", mergeResult.Stderr), ui.TypeStderr)
				// Don't clear marker file - operation failed
				return TimeTravelMergeMsg{
					Success:        false,
					OriginalBranch: originalBranch,
					TimeTravelHash: timeTravelHash,
					Error:          fmt.Sprintf("Merge failed: %s", mergeResult.Stderr),
				}
			} else {
				buffer.Append(fmt.Sprintf("[DEBUG] BRANCH: Conflicts detected in %d files: %v", len(conflictFiles), conflictFiles), ui.TypeStderr)
				buffer.Append("[DEBUG] EARLY RETURN: Entering conflict resolver", ui.TypeStderr)
				// Don't clear marker file yet - conflicts need to be resolved first
				return TimeTravelMergeMsg{
					Success:           false,
					OriginalBranch:    originalBranch,
					TimeTravelHash:    timeTravelHash,
					Error:             "Merge conflicts detected",
					ConflictDetected:  true,
					ConflictedFiles:   conflictFiles,
				}
			}
		}
		buffer.Append("[DEBUG] Merge succeeded, continuing to Phase 2...", ui.TypeStatus)

		// =================================================================
		// PHASE 2: Restore original stash (if any) AFTER merge succeeds
		// =================================================================
		if hasStash {
			buffer.Append("Phase 2: Restoring original work-in-progress...", ui.TypeStatus)
			buffer.Append(fmt.Sprintf("Applying stash %s...", originalStashHash), ui.TypeStatus)

			// Apply stash using hash (git stash apply accepts hash)
			applyResult := Execute("stash", "apply", originalStashHash)
			if !applyResult.Success {
				buffer.Append(fmt.Sprintf("Stash apply failed: %s", applyResult.Stderr), ui.TypeStderr)

				// Check for conflicts
				conflictFiles, err := ListConflictedFiles()
				if err != nil {
					// Stash apply failed but not due to conflicts - don't drop the stash
					buffer.Append("Stash apply failed (not conflicts) - stash will not be dropped", ui.TypeStderr)
					return TimeTravelMergeMsg{
						Success:        false,
						OriginalBranch: originalBranch,
						TimeTravelHash: timeTravelHash,
						Error:          fmt.Sprintf("Failed to restore stash: %s", applyResult.Stderr),
					}
				} else {
					// Stash apply had conflicts - enter conflict resolver
					buffer.Append(fmt.Sprintf("Stash apply conflicts detected in %d files", len(conflictFiles)), ui.TypeStderr)
					buffer.Append("Resolve conflicts to complete stash restoration", ui.TypeInfo)
					// Don't clear marker file yet - conflicts need to be resolved first
					return TimeTravelMergeMsg{
						Success:           false,
						OriginalBranch:    originalBranch,
						TimeTravelHash:    timeTravelHash,
						Error:             "Stash apply conflicts detected",
						ConflictDetected:  true,
						ConflictedFiles:   conflictFiles,
					}
				}
			}

			buffer.Append("Stash applied successfully", ui.TypeStatus)

			// Drop stash using stash@{N} reference (git stash drop requires reference, not hash)
			buffer.Append("[DEBUG] Finding stash reference by hash...", ui.TypeStatus)
			stashRef := FindStashRefByHash(originalStashHash)
			buffer.Append(fmt.Sprintf("[DEBUG] Found stash reference: %s", stashRef), ui.TypeStatus)

			buffer.Append(fmt.Sprintf("Dropping stash %s...", stashRef), ui.TypeStatus)
			dropResult := Execute("stash", "drop", stashRef)
			if !dropResult.Success {
				// Don't panic - just warn user
				buffer.Append(fmt.Sprintf("Warning: Failed to drop stash %s: %s", stashRef, dropResult.Stderr), ui.TypeStderr)
				buffer.Append("You may need to manually drop the stash later", ui.TypeInfo)
			} else {
				buffer.Append("Stash dropped successfully", ui.TypeStatus)
			}

			// Remove from config tracking system
			buffer.Append("[DEBUG] Removing stash entry from config...", ui.TypeStatus)
			config.RemoveStashEntry("time_travel", repoPath)
			buffer.Append("[DEBUG] Stash entry removed from config", ui.TypeStatus)
		}

		// =================================================================
		// PHASE 3: Clean up marker file (only after all operations succeed)
		// =================================================================
		buffer.Append("Cleaning up time travel marker...", ui.TypeStatus)
		if err := ClearTimeTravelInfo(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to clear time travel marker: %v", err), ui.TypeStderr)
		}

		// Don't append "successful" message here - handler will append it
		return TimeTravelMergeMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			TimeTravelHash: timeTravelHash,
			Error:          "",
		}
	}
}

// ExecuteTimeTravelReturn returns from time travel without merging changes
// Returns: TimeTravelReturnMsg
func ExecuteTimeTravelReturn(originalBranch string) func() tea.Msg {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf("Returning from time travel to %s...", originalBranch), ui.TypeStatus)

		// Get current working directory (repo path)
		repoPath, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
		}

		// Get original stash hash (if any) from config tracking system
		originalStashHash, hasStash := config.FindStashEntry("time_travel", repoPath)
		if hasStash {
			buffer.Append(fmt.Sprintf("[DEBUG] Found stash in config - hash: %s", originalStashHash), ui.TypeStatus)
		} else {
			buffer.Append("[DEBUG] No stash found in config (tree was clean before time travel)", ui.TypeStatus)
		}

		// Note: Dirty tree handling is done BEFORE calling this function
		// User discarded changes, so tree is now clean

		// Checkout original branch
		buffer.Append(fmt.Sprintf("Checking out %s...", originalBranch), ui.TypeStatus)
		checkoutResult := Execute("checkout", originalBranch)
		if !checkoutResult.Success {
			buffer.Append(fmt.Sprintf("Error checking out original branch: %s", checkoutResult.Stderr), ui.TypeStderr)
			return TimeTravelReturnMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				Error:          fmt.Sprintf("Failed to checkout original branch: %s", checkoutResult.Stderr),
			}
		}

		buffer.Append(fmt.Sprintf("Checked out %s successfully", originalBranch), ui.TypeStatus)

		// If we had an original stash from BEFORE time travel, restore it
		if hasStash {
			buffer.Append(fmt.Sprintf("Restoring original stash %s...", originalStashHash), ui.TypeStatus)

			// Apply stash using hash (git stash apply accepts hash)
			applyResult := Execute("stash", "apply", originalStashHash)
			if !applyResult.Success {
				buffer.Append(fmt.Sprintf("Stash apply failed: %s", applyResult.Stderr), ui.TypeStderr)

				// Check for conflicts
				conflictFiles, err := ListConflictedFiles()
				if err != nil {
					// Stash apply failed but not due to conflicts - don't drop the stash
					buffer.Append("Stash apply failed (not conflicts) - stash will not be dropped", ui.TypeStderr)
					return TimeTravelReturnMsg{
						Success:        false,
						OriginalBranch: originalBranch,
						Error:          fmt.Sprintf("Failed to restore stash: %s", applyResult.Stderr),
					}
				} else {
					// Stash apply had conflicts - enter conflict resolver
					buffer.Append(fmt.Sprintf("Stash apply conflicts detected in %d files", len(conflictFiles)), ui.TypeStderr)
					return TimeTravelReturnMsg{
						Success:           false,
						OriginalBranch:    originalBranch,
						Error:             "Stash apply conflicts detected",
						ConflictDetected:  true,
						ConflictedFiles:   conflictFiles,
					}
				}
			}

			buffer.Append("Stash applied successfully", ui.TypeStatus)

			// Drop stash using stash@{N} reference (git stash drop requires reference, not hash)
			buffer.Append("[DEBUG] Finding stash reference by hash...", ui.TypeStatus)
			stashRef := FindStashRefByHash(originalStashHash)
			buffer.Append(fmt.Sprintf("[DEBUG] Found stash reference: %s", stashRef), ui.TypeStatus)

			buffer.Append(fmt.Sprintf("Dropping stash %s...", stashRef), ui.TypeStatus)
			dropResult := Execute("stash", "drop", stashRef)
			if !dropResult.Success {
				// Don't panic - just warn user
				buffer.Append(fmt.Sprintf("Warning: Failed to drop stash %s: %s", stashRef, dropResult.Stderr), ui.TypeStderr)
				buffer.Append("You may need to manually drop the stash later", ui.TypeInfo)
			} else {
				buffer.Append("Stash dropped successfully", ui.TypeStatus)
			}

			// Remove from config tracking system
			buffer.Append("[DEBUG] Removing stash entry from config...", ui.TypeStatus)
			config.RemoveStashEntry("time_travel", repoPath)
			buffer.Append("[DEBUG] Stash entry removed from config", ui.TypeStatus)
		}

		// Clean up time travel marker file (needed for state detection)
		if err := ClearTimeTravelInfo(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to clear time travel marker: %v", err), ui.TypeStderr)
		}

		// Don't append "successful" message here - handler will append it
		return TimeTravelReturnMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			Error:          "",
		}
	}
}
