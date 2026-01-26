package app

import (
	"fmt"
	"tit/internal/ui"
)

// GetFooterContent returns the rendered footer for current mode.
// Priority: quitConfirm > clearConfirm > mode-specific hints
// Returns styled string ready for display.
func (a *Application) GetFooterContent() string {
	width := a.sizing.TerminalWidth

	// Priority 1: Quit confirmation (Ctrl+C)
	if a.quitConfirmActive {
		return ui.RenderFooterOverride(GetFooterMessageText(MessageCtrlCConfirm), width, &a.theme)
	}

	// Priority 2: Clear confirmation (ESC in input)
	if a.inputState.ClearConfirming {
		return ui.RenderFooterOverride(GetFooterMessageText(MessageEscClearConfirm), width, &a.theme)
	}

	// Priority 3: Mode-specific hints from SSOT
	hintKey := a.getFooterHintKey()

	// Special case: Menu uses MenuItem.Hint (plain text, not shortcuts)
	if hintKey == "menu" {
		if len(a.menuItems) > 0 && a.selectedIndex < len(a.menuItems) {
			return ui.RenderFooterOverride(a.menuItems[a.selectedIndex].Hint, width, &a.theme)
		}
		return ""
	}

	// Lookup shortcuts from SSOT
	shortcuts := FooterHintShortcuts[hintKey]

	// Special case: Console mode needs scroll status on right
	var rightContent string
	if a.mode == ModeConsole || a.mode == ModeClone {
		rightContent = a.computeConsoleScrollStatus()
	}

	return ui.RenderFooter(shortcuts, width, &a.theme, rightContent)
}

// computeConsoleScrollStatus returns the right-side scroll status for console mode
func (a *Application) computeConsoleScrollStatus() string {
	state := &a.consoleState
	atBottom := state.ScrollOffset >= state.MaxScroll
	remainingLines := state.MaxScroll - state.ScrollOffset

	if atBottom {
		return "(at bottom)"
	}
	if remainingLines > 0 {
		return fmt.Sprintf("↓ %d more", remainingLines)
	}
	return "(can scroll up)"
}

// getFooterHintKey returns the SSOT key for current mode/state
func (a *Application) getFooterHintKey() string {
	switch a.mode {
	case ModeMenu:
		return "menu"

	case ModeHistory:
		if a.historyState.PaneFocused {
			return "history_list"
		}
		return "history_details"

	case ModeFileHistory:
		return a.getFileHistoryHintKey()

	case ModeConsole, ModeClone:
		if a.asyncOperationActive {
			return "console_running"
		}
		return "console_complete"

	case ModeConflictResolve:
		return a.getConflictHintKey()

	case ModeInput, ModeCloneURL, ModeSetupWizard:
		if a.inputState.Value == "" {
			return "input_empty"
		}
		return "input_filled"

	case ModeConfirmation:
		return "confirmation"

	case ModeBranchPicker:
		return "branch_picker"

	case ModePreferences:
		return "preferences"

	case ModeConfig:
		return "menu" // Config menu uses same footer as menu

	default:
		return ""
	}
}

// getFileHistoryHintKey returns the footer hint key for file history mode
func (a *Application) getFileHistoryHintKey() string {
	if a.fileHistoryState == nil {
		return "filehistory_commits"
	}

	switch a.fileHistoryState.FocusedPane {
	case ui.PaneCommits:
		return "filehistory_commits"
	case ui.PaneFiles:
		return "filehistory_files"
	case ui.PaneDiff:
		if a.fileHistoryState.VisualModeActive {
			return "filehistory_visual"
		}
		return "filehistory_diff"
	default:
		return "filehistory_commits"
	}
}

// getConflictHintKey returns the footer hint key for conflict resolver mode
func (a *Application) getConflictHintKey() string {
	if a.conflictResolveState == nil {
		return "conflict_list"
	}

	numColumns := len(a.conflictResolveState.ColumnLabels)
	if a.conflictResolveState.FocusedPane < numColumns {
		return "conflict_list"
	}
	return "conflict_diff"
}
