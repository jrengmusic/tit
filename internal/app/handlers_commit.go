package app

import (
	"fmt"

	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleCommitPush handles OpCommit, OpPush, OpCommitPush operations
func (a *Application) handleCommitPush(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Simple operations: reload state
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		return a, nil
	}

	// CONTRACT: Rebuild cache before showing completion (commit changes history)
	cacheCmd := a.invalidateHistoryCaches()

	a.endAsyncOp()

	// NOTE: "Press ESC..." message is appended in handleCacheProgress after cache completes
	// This ensures cache messages appear before "Press ESC to return to menu"

	return a, cacheCmd
}

// handleForcePush handles OpForcePush
func (a *Application) handleForcePush(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Force push completed - reload state, stay in console
	// User presses ESC to return to menu
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		return a, nil
	}

	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.mode = ModeConsole

	return a, nil
}

// handleHardReset handles OpHardReset
func (a *Application) handleHardReset(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Hard reset completed - reload state, stay in console
	// User presses ESC to return to menu
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		return a, nil
	}

	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.mode = ModeConsole

	return a, nil
}
