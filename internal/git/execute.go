package git

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	
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
	io.Copy(&stdoutBuf, stdout)
	io.Copy(&stderrBuf, stderr)
	
	err := cmd.Wait()
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
		buffer.Append(fmt.Sprintf("Command failed with exit code %d", exitCode), ui.TypeStderr)
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
