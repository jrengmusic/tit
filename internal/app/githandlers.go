package app

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// ========================================
// Git Operation Result Handlers
// ========================================
// All functions handle GitOperationMsg returned from async operations
// They update state, handle errors, and decide next action

// handleGitOperation dispatches GitOperationMsg to the appropriate handler
func (a *Application) handleGitOperation(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Handle success/failure
	if !msg.Success {
		buffer.Append(msg.Error, ui.TypeStderr)
		buffer.Append("Press ESC to return to menu", ui.TypeInfo)
		a.asyncOperationActive = false
		return a, nil
	}

	// Operation succeeded
	if msg.Output != "" {
		buffer.Append(msg.Output, ui.TypeStatus)
	}

	// Handle step-specific post-processing
	switch msg.Step {
	case "init", "clone":
		// Init/clone: reload state and return to menu
		if msg.Path != "" {
			// Change to the path if specified
			if err := os.Chdir(msg.Path); err == nil {
				a.gitState, _ = git.DetectState()
			}
		} else {
			a.gitState, _ = git.DetectState()
		}
		buffer.Append("Press ESC to return to menu", ui.TypeInfo)
		a.asyncOperationActive = false

	case "add_remote", "commit", "push", "pull":
		// All these operations: reload state
		a.gitState, _ = git.DetectState()
		buffer.Append("Press ESC to return to menu", ui.TypeInfo)
		a.asyncOperationActive = false

	default:
		// Default: just cleanup
		buffer.Append("Press ESC to return to menu", ui.TypeInfo)
		a.asyncOperationActive = false
	}

	return a, nil
}
