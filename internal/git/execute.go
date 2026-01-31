package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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

// ExecuteWithStreaming runs a git command and streams output to the buffer
// Use this for user-initiated actions that should display console output
// WORKER THREAD - Must be called from async operation
func ExecuteWithStreaming(ctx context.Context, args ...string) CommandResult {
	// Clean any stale git locks from interrupted operations
	cleanStaleLocks()

	// Log the command being executed
	cmdString := "git " + strings.Join(args, " ")
	Log(cmdString)

	cmd := exec.CommandContext(ctx, "git", args...)

	// CRITICAL: Disable interactive prompts - fail fast instead of hanging
	// This prevents git from waiting for SSH passphrase, HTTP auth, etc.
	// User must have SSH keys or credential helpers configured
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		Error(fmt.Sprintf("Error creating stdout pipe: %v", err))
		return CommandResult{Success: false, ExitCode: 1, Stderr: err.Error()}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		Error(fmt.Sprintf("Error creating stderr pipe: %v", err))
		return CommandResult{Success: false, ExitCode: 1, Stderr: err.Error()}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		Error(fmt.Sprintf("Error starting command: %v", err))
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
						Error(line)
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
					Error(line)
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

	// Check if context was cancelled
	if ctx.Err() == context.Canceled {
		return CommandResult{
			Stdout:   "",
			Stderr:   "aborted",
			ExitCode: 1,
			Success:  false,
		}
	}

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
		Log("Command completed successfully")
	} else {
		// Don't say "failed" - exit code 1 can be expected (e.g., merge conflicts)
		Log(fmt.Sprintf("Command exited with code %d", exitCode))
	}

	return CommandResult{
		Stdout:   "", // Output already streamed to buffer
		Stderr:   "",
		ExitCode: exitCode,
		Success:  exitCode == 0,
	}
}
