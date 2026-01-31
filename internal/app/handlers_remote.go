package app

import (
	"fmt"

	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleAddRemote handles OpAddRemote - chains to OpFetchRemote
func (a *Application) handleAddRemote(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Chain: OpAddRemote → OpFetchRemote
	buffer.Append(OutputMessages["fetching_remote"], ui.TypeInfo)
	return a, a.cmdFetchRemote()
}

// handleFetchRemote handles OpFetchRemote - sets upstream tracking
func (a *Application) handleFetchRemote(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Fetch complete: set upstream tracking
	buffer.Append(OutputMessages["setting_upstream"], ui.TypeInfo)
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		return a, nil
	}

	// SSOT: detached HEAD cannot have upstream - skip upstream setting
	if a.gitState.Detached {
		buffer.Append(ErrorMessages["cannot_set_upstream_detached"], ui.TypeInfo)
		a.endAsyncOp()
		return a, nil
	}

	return a, a.cmdSetUpstream(a.gitState.CurrentBranch)
}
