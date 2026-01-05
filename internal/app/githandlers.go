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

	// Handle failure
	if !msg.Success {
		buffer.Append(msg.Error, ui.TypeStderr)
		buffer.Append(GetFooterMessageText(MessageOperationFailed), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationFailed)
		a.asyncOperationActive = false
		return a, nil
	}

	// Operation succeeded
	if msg.Output != "" {
		buffer.Append(msg.Output, ui.TypeStatus)
	}

	// Handle step-specific post-processing and chaining
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
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false

	case "add_remote":
		// Chain: add_remote → fetch_remote
		buffer.Append("Fetching from remote...", ui.TypeInfo)
		return a, a.cmdFetchRemote()

	case "fetch_remote":
		// Fetch complete: set upstream tracking
		buffer.Append("Setting upstream tracking...", ui.TypeInfo)
		a.gitState, _ = git.DetectState()
		return a, a.cmdSetUpstream(a.gitState.CurrentBranch)

	case "commit", "push", "pull":
		// Simple operations: reload state
		a.gitState, _ = git.DetectState()
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false

	default:
		// Default: just cleanup
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
	}

	return a, nil
}
