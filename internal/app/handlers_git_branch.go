package app

import (
	"context"
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Git branch operations: switch, stash

// cmdSwitchBranch performs git switch to the target branch
// Handles conflicts if they occur during the switch (files that would be overwritten)
func (a *Application) cmdSwitchBranch(targetBranch string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Execute git switch
		result := git.ExecuteWithStreaming(ctx, "switch", targetBranch)
		if !result.Success {
			// Check if we're in a conflicted state after failed switch
			// This can happen when switching would overwrite local changes
			if msg := a.checkForConflicts("branch_switch", false); msg != nil {
				buffer.Append(fmt.Sprintf("Conflicts detected while switching to %s", targetBranch), ui.TypeWarning)
				msg.BranchName = targetBranch
				msg.Error = fmt.Sprintf("Conflicts switching to %s", targetBranch)
				return *msg
			}

			// Other failure (permissions, invalid branch, etc)
			return GitOperationMsg{
				Step:    "branch_switch",
				Success: false,
				Error:   fmt.Sprintf("Failed to switch to %s: %s", targetBranch, result.Stderr),
			}
		}

		// Success - Update() will refresh state automatically
		buffer.Append(fmt.Sprintf("Switched to branch %s", targetBranch), ui.TypeInfo)
		return GitOperationMsg{
			Step:       "branch_switch",
			Success:    true,
			Output:     fmt.Sprintf("Switched to branch %s", targetBranch),
			BranchName: targetBranch,
		}
	}
}

// handleNewBranchNameSubmit validates branch name and creates new branch
func (a *Application) handleNewBranchNameSubmit() (tea.Model, tea.Cmd) {
	inputState := a.OperationState.GetInputState()
	branchName := strings.TrimSpace(inputState.Value)

	if branchName == "" {
		a.footerHint = ErrorMessages["branch_name_empty"]
		return a, nil
	}

	// Validate branch name using git check-ref-format
	validateResult := git.Execute("check-ref-format", "--branch", branchName)
	if !validateResult.Success {
		a.footerHint = fmt.Sprintf(ErrorMessages["branch_name_invalid"], branchName)
		return a, nil
	}

	// Check if branch already exists
	existsResult := git.Execute("rev-parse", "--verify", "refs/heads/"+branchName)
	if existsResult.Success {
		a.footerHint = fmt.Sprintf(ErrorMessages["branch_already_exists"], branchName)
		return a, nil
	}

	// Set up async state for console display
	buffer := ui.GetBuffer()
	buffer.Clear()
	buffer.Append(fmt.Sprintf("Creating branch %s...", branchName), ui.TypeStatus)

	a.startAsyncOp()
	a.workflowState.PreviousMode = ModeConfig
	a.workflowState.PreviousMenuIndex = 0
	a.mode = ModeConsole
	a.consoleState.Reset()
	inputState.Value = ""

	return a, a.cmdCreateBranch(branchName)
}

// cmdCreateBranch performs git checkout -b to create and switch to new branch
func (a *Application) cmdCreateBranch(branchName string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		result := git.ExecuteWithStreaming(ctx, "checkout", "-b", branchName)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpBranchCreate,
				Success: false,
				Error:   fmt.Sprintf("Failed to create branch %s: %s", branchName, result.Stderr),
			}
		}

		buffer.Append(fmt.Sprintf("Created and switched to branch %s", branchName), ui.TypeInfo)
		return GitOperationMsg{
			Step:       OpBranchCreate,
			Success:    true,
			Output:     fmt.Sprintf("Created and switched to branch %s", branchName),
			BranchName: branchName,
		}
	}
}

