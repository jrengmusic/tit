package app

import (
	"fmt"
	"os"

	"tit/internal/git"
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

func (a *Application) RenderStateHeader() string {
	state := a.gitState

	if state == nil || state.Operation == git.NotRepo {
		return ""
	}

	cwd, _ := os.Getwd()

	remoteURL := "🔌 NO REMOTE"
	remoteColor := a.theme.DimmedTextColor
	if state.Remote == git.HasRemote {
		url := git.GetRemoteURL()
		if url != "" {
			remoteURL = "🔗 " + url
			remoteColor = a.theme.AccentTextColor
		}
	}

	wtInfo := a.workingTreeInfo[state.WorkingTree]
	wtDesc := []string{wtInfo.Description(state.CommitsAhead, state.CommitsBehind)}

	timelineEmoji := "🔌"
	timelineLabel := "N/A"
	timelineColor := a.theme.DimmedTextColor
	timelineDesc := []string{"No remote configured."}

	if state.Operation == git.TimeTraveling {
		if a.timeTravelState.GetInfo() != nil {
			shortHash := a.timeTravelState.GetInfo().CurrentCommit.Hash
			if len(shortHash) >= 7 {
				shortHash = shortHash[:7]
			}
			timelineEmoji = "📌"
			timelineLabel = "DETACHED @ " + shortHash
			timelineColor = a.theme.OutputWarningColor
			timelineDesc = []string{"Viewing commit from " + a.timeTravelState.GetInfo().CurrentCommit.Time.Format("Jan 2, 2006")}
		}
	} else if state.Timeline != "" {
		tlInfo := a.timelineInfo[state.Timeline]
		timelineEmoji = tlInfo.Emoji
		timelineLabel = tlInfo.Label
		timelineColor = tlInfo.Color
		timelineDesc = []string{tlInfo.Description(state.CommitsAhead, state.CommitsBehind)}
	}

	// Operation status (right column top)
	opInfo := a.operationInfo[state.Operation]

	// Branch name (right column bottom)
	branchName := state.CurrentBranch
	if branchName == "" {
		branchName = "N/A"
	}

	headerState := ui.HeaderState{
		CurrentDirectory: cwd,
		RemoteURL:        remoteURL,
		RemoteColor:      remoteColor,
		OperationEmoji:   opInfo.Emoji,
		OperationLabel:   opInfo.Label,
		OperationColor:   opInfo.Color,
		BranchEmoji:      "🌿",
		BranchLabel:      branchName,
		BranchColor:      a.theme.AccentTextColor,
		WorkingTreeEmoji: wtInfo.Emoji,
		WorkingTreeLabel: wtInfo.Label,
		WorkingTreeDesc:  wtDesc,
		WorkingTreeColor: wtInfo.Color,
		TimelineEmoji:    timelineEmoji,
		TimelineLabel:    timelineLabel,
		TimelineDesc:     timelineDesc,
		TimelineColor:    timelineColor,
		SyncInProgress:   a.activityState.IsAutoUpdateInProgress(),
		SyncFrame:        a.activityState.GetFrame(),
	}

	info := ui.RenderHeaderInfo(a.sizing, a.theme, headerState)

	return ui.RenderHeader(a.sizing, a.theme, info)
}

// isInputMode checks if current mode accepts text input

func (a *Application) isInputMode() bool {
	return a.mode == ModeInput ||
		a.mode == ModeCloneURL ||
		(a.mode == ModeSetupWizard && a.environmentState.SetupWizardStep == SetupStepEmail)
}

// menuItemsToMaps converts MenuItem slice to map slice for rendering
// Note: Hint is excluded from maps (displayed in footer instead)

func (a *Application) menuItemsToMaps(items []MenuItem) []map[string]interface{} {
	maps := make([]map[string]interface{}, len(items))
	for i, item := range items {
		maps[i] = map[string]interface{}{
			"ID":            item.ID,
			"Shortcut":      item.Shortcut,
			"ShortcutLabel": item.ShortcutLabel,
			"Emoji":         item.Emoji,
			"Label":         item.Label,
			"Enabled":       item.Enabled,
			"Separator":     item.Separator,
		}
	}
	return maps
}

// buildKeyHandlers builds the complete handler registry for all modes
// Global handlers take priority and are merged into each mode

