package app

import (
	"context"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Rewind Confirmation Handlers
// ========================================

// executeConfirmRewind handles "Rewind" choice
// executeConfirmRewind executes git reset --hard at pending commit
func (a *Application) executeConfirmRewind() (tea.Model, tea.Cmd) {
	if a.workflowState.PendingRewindCommit == "" {
		return a.returnToMenu()
	}

	commitHash := a.workflowState.PendingRewindCommit
	a.workflowState.PendingRewindCommit = "" // Clear after capturing

	// Set up async operation
	a.startAsyncOp()
	a.workflowState.PreviousMode = ModeHistory
	a.workflowState.PreviousMenuIndex = a.pickerState.History.SelectedIdx
	a.mode = ModeConsole
	a.consoleState.Reset()
	ui.GetBuffer().Clear()

	// Start rewind + refresh ticker
	return a, tea.Batch(
		a.executeRewindOperation(commitHash),
		a.cmdRefreshConsole(),
	)
}

// executeRejectRewind handles "Cancel" choice on rewind confirmation
func (a *Application) executeRejectRewind() (tea.Model, tea.Cmd) {
	a.workflowState.PendingRewindCommit = "" // Clear pending commit
	return a.returnToMenu()
}

// executeRewindOperation performs the actual git reset --hard in a worker goroutine
func (a *Application) executeRewindOperation(commitHash string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		_, err := git.ResetHardAtCommit(ctx, commitHash)

		return RewindMsg{
			Commit:  commitHash,
			Success: err == nil,
			Error:   errorOrEmpty(err),
		}
	}
}
