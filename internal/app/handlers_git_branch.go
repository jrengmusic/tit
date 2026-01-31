package app

import (
	"context"
	"fmt"

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

// cmdBranchSwitchWithStash performs: stash → switch → stash apply
func (a *Application) cmdBranchSwitchWithStash(targetBranch string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Step1: Stash changes
		buffer.Append("Stashing changes...", ui.TypeStatus)
		stashResult := git.Execute("stash", "push", "-u")
		if !stashResult.Success {
			buffer.Append(fmt.Sprintf("Failed to stash: %s", stashResult.Stderr), ui.TypeStderr)
			return GitOperationMsg{
				Step:    "branch_switch",
				Success: false,
				Error:   "Failed to stash changes",
			}
		}
		buffer.Append("Changes stashed", ui.TypeStatus)

		// Step2: Switch branch
		buffer.Append(fmt.Sprintf("Switching to %s...", targetBranch), ui.TypeStatus)
		switchResult := git.ExecuteWithStreaming(ctx, "switch", targetBranch)
		if !switchResult.Success {
			buffer.Append(fmt.Sprintf("Failed to switch: %s", switchResult.Stderr), ui.TypeStderr)

			// Try to restore stash on failure
			buffer.Append("Restoring stash...", ui.TypeStatus)
			git.Execute("stash", "pop")

			return GitOperationMsg{
				Step:    "branch_switch",
				Success: false,
				Error:   fmt.Sprintf("Failed to switch to %s", targetBranch),
			}
		}
		buffer.Append(fmt.Sprintf("Switched to %s", targetBranch), ui.TypeStatus)

		// Step3: Restore stash
		buffer.Append("Restoring changes...", ui.TypeStatus)
		applyResult := git.Execute("stash", "pop")
		if !applyResult.Success {
			buffer.Append("Warning: Stash apply failed (conflicts or errors)", ui.TypeWarning)
			buffer.Append("Your changes are still in stash (use 'git stash apply')", ui.TypeInfo)
		} else {
			buffer.Append("Changes restored", ui.TypeStatus)
		}

		return GitOperationMsg{
			Step:    "branch_switch",
			Success: true,
			Output:  fmt.Sprintf("Switched to %s", targetBranch),
		}
	}
}
