package app

import (
	"fmt"

	"tit/internal/ui"
)

func (a *Application) View() string {
	var contentText string

	// Render based on current mode
	switch a.mode {
	case ModeMenu:
		contentText = ui.RenderMenuWithBanner(a.sizing, a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme)

	case ModeConsole, ModeClone:
		// Console output (full-screen mode, footer handled by GetFooterContent)
		contentText = ui.RenderConsoleOutputFullScreen(
			a.consoleState.GetStateRef(),
			a.consoleState.GetBuffer(),
			a.theme,
			a.sizing.TerminalWidth,
			a.sizing.TerminalHeight,
			a.isAsyncActive() && !a.isAsyncAborted(),
			a.isAsyncAborted(),
			a.consoleState.IsAutoScroll(),
		)

	case ModeConfirmation:
		// Confirmation dialog (centered in content area)
		if a.dialogState.GetDialog() != nil {
			contentText = a.dialogState.GetDialog().Render(a.sizing.ContentHeight)
		} else {
			// Fallback if no dialog - return to menu
			a.mode = ModeMenu
			contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme, a.sizing.ContentHeight, a.sizing.ContentInnerWidth)
		}

	case ModeSelectBranch:
		// Dynamic menu from cloneBranches
		items := make([]map[string]interface{}, len(a.workflowState.CloneBranches))
		for i, branch := range a.workflowState.CloneBranches {
			items[i] = map[string]interface{}{
				"ID":        branch,
				"Shortcut":  "",
				"Emoji":     "🌿",
				"Label":     branch,
				"Hint":      fmt.Sprintf("Set %s as canon branch", branch),
				"Enabled":   true,
				"Separator": false,
			}
		}
		contentText = ui.RenderMenuWithHeight(items, a.selectedIndex, a.theme, a.sizing.ContentHeight, a.sizing.ContentInnerWidth)
	case ModeInput:
		textInputState := ui.TextInputState{
			Value:     a.inputState.Value,
			CursorPos: a.inputState.CursorPosition,
			Height:    a.inputState.Height,
		}

		footer := a.GetFooterContent()
		return ui.RenderTextInputFullScreen(
			a.sizing,
			a.theme,
			a.inputState.Prompt,
			textInputState,
			footer,
		)
	case ModeCloneURL:
		textInputState := ui.TextInputState{
			Value:     a.inputState.Value,
			CursorPos: a.inputState.CursorPosition,
			Height:    a.inputState.Height,
		}

		footer := a.GetFooterContent()
		return ui.RenderTextInputFullScreen(
			a.sizing,
			a.theme,
			a.inputState.Prompt,
			textInputState,
			footer,
		)
	case ModeCloneLocation:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuCloneLocation()), a.selectedIndex, a.theme, a.sizing.ContentHeight, a.sizing.ContentInnerWidth)
	case ModeInitializeLocation:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuInitializeLocation()), a.selectedIndex, a.theme, a.sizing.ContentHeight, a.sizing.ContentInnerWidth)

	case ModeHistory:
		// Render history split-pane view (footer handled by GetFooterContent)
		if a.pickerState.History == nil {
			contentText = "History state not initialized"
		} else {
			contentText = ui.RenderHistorySplitPane(
				a.pickerState.History,
				a.theme,
				a.sizing.TerminalWidth,
				a.sizing.TerminalHeight,
			)
		}
	case ModeFileHistory:
		// Render file(s) history split-pane view (footer handled by GetFooterContent)
		if a.pickerState.FileHistory == nil {
			contentText = "File history state not initialized"
		} else {
			contentText = ui.RenderFileHistorySplitPane(
				a.pickerState.FileHistory,
				a.theme,
				a.sizing.TerminalWidth,
				a.sizing.TerminalHeight,
			)
		}
	case ModeConflictResolve:
		// Render conflict resolution UI using generic N-column view (footer handled by GetFooterContent)
		if a.conflictResolveState == nil {
			contentText = "No conflict state initialized"
		} else {
			contentText = ui.RenderConflictResolveGeneric(
				a.conflictResolveState.Files,
				a.conflictResolveState.SelectedFileIndex,
				a.conflictResolveState.FocusedPane,
				a.conflictResolveState.NumColumns,
				a.conflictResolveState.ColumnLabels,
				a.conflictResolveState.ScrollOffsets,
				a.conflictResolveState.LineCursors,
				a.width,
				a.height,
				a.theme,
			)
		}
	case ModeSetupWizard:
		// Email step uses same full-screen input as ModeInput
		if a.environmentState.SetupWizardStep == SetupStepEmail {
			return a.renderSetupEmail()
		}
		// Other setup wizard steps
		contentText = a.renderSetupWizard()

	case ModeConfig:
		contentText = ui.RenderMenuWithBanner(a.sizing, a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme)

	case ModeBranchPicker:
		if a.pickerState.BranchPicker == nil {
			// Initialize branch picker state if not yet created
			a.pickerState.BranchPicker = &ui.BranchPickerState{
				Branches:          []ui.BranchInfo{},
				SelectedIdx:       0,
				PaneFocused:       true,
				ListScrollOffset:  0,
				DetailsLineCursor: 0,
				DetailsScrollOff:  0,
			}
		}
		// Render using SSOT (ListPane + TextPane) matching history pattern
		contentText = ui.RenderBranchPickerSplitPane(a.pickerState.BranchPicker, a.theme, a.sizing.TerminalWidth, a.sizing.TerminalHeight)

	case ModePreferences:
		// All menus work the same SSOT way - generate items when needed
		if len(a.menuItems) == 0 {
			a.menuItems = a.GeneratePreferencesMenu()
		}
		// Render preferences with banner (reads values directly from config)
		contentText = ui.RenderPreferencesWithBanner(a.appConfig, a.selectedIndex, a.theme, a.sizing)

	default:
		panic(fmt.Sprintf("Unknown app mode: %v", a.mode))
	}

	// Full-screen modes: skip header, show footer only
	if a.mode == ModeConsole || a.mode == ModeClone || a.mode == ModeFileHistory || a.mode == ModeHistory || a.mode == ModeConflictResolve || a.mode == ModeBranchPicker {
		footer := a.GetFooterContent()
		return contentText + "\n" + footer
	}

	// Render header using state header (or placeholder)
	header := a.RenderStateHeader()

	// Render footer content using unified footer system
	footer := a.GetFooterContent()

	// Use reactive layout
	return ui.RenderReactiveLayout(a.sizing, a.theme, header, contentText, footer)
}

// Init initializes the application
