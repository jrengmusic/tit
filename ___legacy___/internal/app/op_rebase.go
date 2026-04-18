package app

import (
	"fmt"

	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdRebaseContinue continues a rebase after conflict resolution
func (a *Application) cmdRebaseContinue() tea.Cmd {
	return a.executeGitOp(OpRebaseContinue, "rebase", "--continue")
}

// cmdRebaseAbort aborts a rebase in progress
func (a *Application) cmdRebaseAbort() tea.Cmd {
	return a.executeGitOp(OpRebaseAbort, "rebase", "--abort")
}

// handleRebaseContinue handles OpRebaseContinue: checks for further conflicts or completes.
func (a *Application) handleRebaseContinue(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if git.HasConflicts() {
		// More conflicts remain in rebase — loop back into resolver
		a.conflictResolveState = nil
		return a.setupConflictResolverForRebase()
	}
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleRebaseAbort handles OpRebaseAbort: reload state and show completion.
func (a *Application) handleRebaseAbort(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleFinalizeMergeFromMenu handles OpFinalizeMergeFromMenu: reload state after merge commit.
func (a *Application) handleFinalizeMergeFromMenu(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		return a, nil
	}
	// Regenerate menu to reflect new state
	menu := a.GenerateMenu()
	a.menuItems = menu
	a.selectedIndex = 0
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.mode = ModeConsole
	return a, nil
}
