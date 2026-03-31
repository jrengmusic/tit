package app

import (
	"context"
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Branch Merge Confirmation Handlers
// ========================================

// handleMergeBranchSelection handles branch picker selection for merge
func (a *Application) handleMergeBranchSelection(sourceBranch string) (tea.Model, tea.Cmd) {
	// Determine current branch
	currentBranch := ""
	if a.gitState != nil {
		currentBranch = a.gitState.CurrentBranch
	}

	// Check if working tree is dirty
	statusResult := git.Execute("status", "--porcelain")
	hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""

	if hasDirtyTree {
		// Dirty tree: show stash confirmation
		a.mode = ModeConfirmation
		dialogContext := map[string]string{
			"sourceBranch": sourceBranch,
		}
		msg := ConfirmationMessages[string(ConfirmMergeBranchDirty)]
		dialog := ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       msg.Title,
				Explanation: msg.Explanation,
				YesLabel:    msg.YesLabel,
				NoLabel:     msg.NoLabel,
				ActionID:    string(ConfirmMergeBranchDirty),
			},
			a.sizing.ContentInnerWidth,
			&a.theme,
		)
		a.dialogState.Show(dialog, dialogContext)
		dialog.SelectNo()
		return a, nil
	}

	// Clean tree: show merge confirmation
	a.mode = ModeConfirmation
	dialogContext := map[string]string{
		"sourceBranch": sourceBranch,
	}
	msg := ConfirmationMessages[string(ConfirmMergeBranch)]
	dialog := ui.NewConfirmationDialog(
		ui.ConfirmationConfig{
			Title:       msg.Title,
			Explanation: fmt.Sprintf(msg.Explanation, sourceBranch, currentBranch),
			YesLabel:    msg.YesLabel,
			NoLabel:     msg.NoLabel,
			ActionID:    string(ConfirmMergeBranch),
		},
		a.sizing.ContentInnerWidth,
		&a.theme,
	)
	a.dialogState.Show(dialog, dialogContext)
	dialog.SelectNo()
	return a, nil
}

// executeConfirmMergeBranch handles YES response to merge branch confirmation (clean tree)
func (a *Application) executeConfirmMergeBranch() (tea.Model, tea.Cmd) {
	sourceBranch := a.dialogState.GetContext()["sourceBranch"]
	a.dialogState.Hide()

	if sourceBranch == "" {
		return a.returnToMenu()
	}

	a.prepareAsyncOperation(fmt.Sprintf("Merging %s...", sourceBranch))
	return a, a.cmdMergeBranch(sourceBranch)
}

// executeRejectMergeBranch handles NO/Cancel response to merge branch confirmation
func (a *Application) executeRejectMergeBranch() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeConfirmMergeBranchDirty handles YES response (Stash and merge via Dirty Operation Protocol)
func (a *Application) executeConfirmMergeBranchDirty() (tea.Model, tea.Cmd) {
	sourceBranch := a.dialogState.GetContext()["sourceBranch"]
	a.dialogState.Hide()

	if sourceBranch == "" {
		return a.returnToMenu()
	}

	// Initialize dirty operation state
	a.dirtyOperationState = NewDirtyOperationState("dirty_merge", true)
	a.dirtyOperationState.MergeBranch = sourceBranch
	a.dirtyOperationState.OriginalBranch = a.gitState.CurrentBranch

	a.prepareAsyncOperation(fmt.Sprintf("Saving changes and merging %s...", sourceBranch))
	return a, a.cmdDirtyMergeSnapshot(true)
}

// executeRejectMergeBranchDirty handles NO/Cancel response to dirty merge
func (a *Application) executeRejectMergeBranchDirty() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()
	return a.returnToMenu()
}

// cmdMergeBranch performs git merge on the source branch
func (a *Application) cmdMergeBranch(sourceBranch string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		result := git.ExecuteWithStreaming(ctx, "merge", sourceBranch)
		if !result.Success {
			// Check for merge conflicts
			if conflictMsg := a.checkForConflicts(OpMergeBranch, false); conflictMsg != nil {
				buffer.Append(fmt.Sprintf("Conflicts detected while merging %s", sourceBranch), ui.TypeWarning)
				conflictMsg.BranchName = sourceBranch
				conflictMsg.Error = fmt.Sprintf("Conflicts merging %s", sourceBranch)
				return *conflictMsg
			}

			return GitOperationMsg{
				Step:    OpMergeBranch,
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["merge_branch_failed"], result.Stderr),
			}
		}

		buffer.Append(fmt.Sprintf("Merged %s successfully", sourceBranch), ui.TypeInfo)
		return GitOperationMsg{
			Step:       OpMergeBranch,
			Success:    true,
			Output:     fmt.Sprintf("Merged %s successfully", sourceBranch),
			BranchName: sourceBranch,
		}
	}
}

// cmdFinalizeBranchMerge stages and commits resolved branch merge conflicts
func (a *Application) cmdFinalizeBranchMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeBranchMerge,
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the merge
		result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Merge resolved conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeBranchMerge,
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append(OutputMessages["merge_finalized"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpFinalizeBranchMerge,
			Success: true,
			Output:  OutputMessages["merge_finalized"],
		}
	}
}
