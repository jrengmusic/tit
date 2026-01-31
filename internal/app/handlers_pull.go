package app

import (
	"fmt"

	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handlePull handles OpPull - reloads state after successful pull
func (a *Application) handlePull(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Pull operation succeeded (no conflicts)
	// Reload state and return to menu
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true) // Re-enable exit on error
		return a, nil
	}

	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.setExitAllowed(true) // Re-enable exit after successful pull

	return a, nil
}

// handleFinalizePullMerge handles OpFinalizePullMerge
func (a *Application) handleFinalizePullMerge(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Merge finalization succeeded: reload state and stay in console
	// User must press ESC to return to menu (ensures merge completed before menu reachable)
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true) // Re-enable exit on error
		a.mode = ModeConsole
		return a, nil
	}

	buffer.Append(OutputMessages["merge_finalized"], ui.TypeStatus)
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.setExitAllowed(true) // Re-enable exit after successful merge finalization
	a.conflictResolveState = nil
	a.mode = ModeConsole // Stay in console, user presses ESC to return to menu

	return a, nil
}

// handleAbortMerge handles OpAbortMerge
func (a *Application) handleAbortMerge(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Merge abort succeeded: reload state and stay in console
	// User must press ESC to return to menu (ensures abort completed before menu reachable)
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true) // Re-enable exit on error
		a.mode = ModeConsole
		return a, nil
	}

	buffer.Append(OutputMessages["abort_successful"], ui.TypeStatus)
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.setExitAllowed(true) // Re-enable exit after successful abort
	a.conflictResolveState = nil
	a.mode = ModeConsole // Stay in console, user presses ESC to return to menu

	return a, nil
}

// handleBranchSwitch handles branch switch operations
func (a *Application) handleBranchSwitch(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Branch switch completed successfully (no conflicts or conflicts resolved)
	// Reload state and return to config menu
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		return a, nil
	}

	// CRITICAL: Invalidate history cache when switching branches
	// Cache was built for previous branch, needs rebuild for new branch
	a.cacheManager.SetLoadingStarted(true)
	cacheCmd := a.invalidateHistoryCaches()

	// Regenerate menu with new branch state
	menu := a.GenerateMenu()
	a.menuItems = menu
	a.selectedIndex = 0

	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.mode = ModeConsole // Stay in console so user sees the success message

	return a, cacheCmd
}

// handleFinalizeBranchSwitch handles finalize_branch_switch step
func (a *Application) handleFinalizeBranchSwitch(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Branch switch conflicts resolved and finalized
	// Reload state and return to config menu
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true)
		a.mode = ModeConsole
		return a, nil
	}

	// CRITICAL: Invalidate history cache when switching branches
	a.cacheManager.SetLoadingStarted(true)
	cacheCmd := a.invalidateHistoryCaches()

	// Regenerate menu with new branch state
	menu := a.GenerateMenu()
	a.menuItems = menu
	a.selectedIndex = 0

	buffer.Append(OutputMessages["merge_finalized"], ui.TypeStatus)
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.setExitAllowed(true)
	a.conflictResolveState = nil
	a.mode = ModeConsole // Stay in console, user presses ESC to return to menu

	return a, cacheCmd
}
