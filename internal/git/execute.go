package git

import (
	"bufio"
	"fmt"
	"os/exec"
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

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	exitCode := 0
	if err != nil {
		exitCode = 1
	}

	return CommandResult{
		Stdout:   outputStr,
		Stderr:   outputStr,
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
