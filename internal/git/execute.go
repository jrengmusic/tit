package git

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	tea "github.com/charmbracelet/bubbletea"
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

// ExecuteWithStreaming runs a git command and streams output to the buffer
// Use this for user-initiated actions that should display console output
// WORKER THREAD - Must be called from async operation
func ExecuteWithStreaming(args ...string) CommandResult {
	buffer := ui.GetBuffer()

	// Log the command being executed
	cmdString := "git " + strings.Join(args, " ")
	buffer.Append(cmdString, ui.TypeCommand)

	cmd := exec.Command("git", args...)

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
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Git uses \r for progress updates - take only the last segment
			if strings.Contains(line, "\r") {
				parts := strings.Split(line, "\r")
				line = parts[len(parts)-1] // Last segment is final progress state
			}
			line = strings.TrimSpace(line)
			if line != "" {
				buffer.Append(line, ui.TypeStdout)
			}
		}
	}()

	// Stream stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// Git uses \r for progress updates - take only the last segment
			if strings.Contains(line, "\r") {
				parts := strings.Split(line, "\r")
				line = parts[len(parts)-1] // Last segment is final progress state
			}
			line = strings.TrimSpace(line)
			if line != "" {
				buffer.Append(line, ui.TypeStderr)
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
// Returns: []CommitInfo with Hash, Subject, Time populated
// Format: git log --pretty=%H%n%s%n%ai -N
// Example output: hash1\nsubject1\nISO-date1\nhash2\nsubject2\nISO-date2\n...
func FetchRecentCommits(limit int) ([]CommitInfo, error) {
	result := Execute("log", fmt.Sprintf("-%d", limit), "--pretty=%H%n%s%n%ai")
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
		buffer.Append(fmt.Sprintf("Merging time travel changes back to %s...", originalBranch), ui.TypeStatus)

		// Checkout original branch
		checkoutResult := Execute("checkout", originalBranch)
		if !checkoutResult.Success {
			buffer.Append(fmt.Sprintf("Error checking out original branch: %s", checkoutResult.Stderr), ui.TypeStderr)
			return TimeTravelMergeMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				TimeTravelHash: timeTravelHash,
				Error:          fmt.Sprintf("Failed to checkout original branch: %s", checkoutResult.Stderr),
			}
		}

		// Merge the time travel commit
		mergeResult := Execute("merge", timeTravelHash)
		if !mergeResult.Success {
			buffer.Append(fmt.Sprintf("Merge completed with status: %s", mergeResult.Stderr), ui.TypeInfo)
			
			// Check for conflicts
			conflictFiles, err := ListConflictedFiles()
			if err != nil {
				buffer.Append("No conflicts detected", ui.TypeStatus)
			} else {
				buffer.Append(fmt.Sprintf("Conflicts detected in %d files", len(conflictFiles)), ui.TypeStderr)
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

		// Clear time travel info
		err := ClearTimeTravelInfo()
		if err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to clear time travel info: %v", err), ui.TypeStderr)
		}

		buffer.Append("Time travel merge successful", ui.TypeStatus)
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

		// Checkout original branch
		checkoutResult := Execute("checkout", originalBranch)
		if !checkoutResult.Success {
			buffer.Append(fmt.Sprintf("Error checking out original branch: %s", checkoutResult.Stderr), ui.TypeStderr)
			return TimeTravelReturnMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				Error:          fmt.Sprintf("Failed to checkout original branch: %s", checkoutResult.Stderr),
			}
		}

		// Clear time travel info
		err := ClearTimeTravelInfo()
		if err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to clear time travel info: %v", err), ui.TypeStderr)
		}

		buffer.Append("Time travel return successful", ui.TypeStatus)
		return TimeTravelReturnMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			Error:          "",
		}
	}
}
