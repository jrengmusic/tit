package app

import (
	"fmt"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Config Menu Dispatchers
// ========================================

// dispatchConfigAddRemote starts add remote workflow from config menu
func (a *Application) dispatchConfigAddRemote(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: "New remote URL:",
		InputAction: "config_add_remote_url",
		FooterHint:  "Enter remote repository URL",
		ResetFields: []string{},
	})
	return nil
}

// dispatchConfigSwitchRemote starts switch remote workflow from config menu
func (a *Application) dispatchConfigSwitchRemote(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: "New remote URL:",
		InputAction: "config_switch_remote_url",
		FooterHint:  "Enter new remote repository URL",
		ResetFields: []string{},
	})
	return nil
}

// dispatchConfigRemoveRemote shows confirmation dialog to remove remote
func (a *Application) dispatchConfigRemoveRemote(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.dialogState.SetContext(map[string]string{})
	msg := ConfirmationMessages["remove_remote"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "config_remove_remote",
	}
	dialog := ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	app.dialogState.Show(dialog, nil)
	return nil
}

// dispatchConfigToggleAutoUpdate toggles auto-update setting
func (a *Application) dispatchConfigToggleAutoUpdate(app *Application) tea.Cmd {
	return app.cmdToggleAutoUpdate()
}

// dispatchConfigSwitchBranch enters branch picker mode
func (a *Application) dispatchConfigSwitchBranch(app *Application) tea.Cmd {
	// Load branches into the branch picker state
	branches, err := git.ListBranchesWithDetails()
	if err != nil {
		app.footerHint = fmt.Sprintf("Failed to load branches: %v", err)
		return nil
	}

	// Convert git.BranchDetails to ui.BranchInfo
	uiBranches := make([]ui.BranchInfo, len(branches))
	for i, b := range branches {
		uiBranches[i] = ui.BranchInfo{
			Name:           b.Name,
			IsCurrent:      b.IsCurrent,
			LastCommitTime: b.LastCommitTime,
			LastCommitHash: b.LastCommitHash,
			LastCommitSubj: b.LastCommitSubj,
			Author:         b.Author,
			TrackingRemote: b.TrackingRemote,
			Ahead:          b.Ahead,
			Behind:         b.Behind,
		}
	}

	// Initialize branch picker state (mirrors history state pattern: list + details pane)
	app.pickerState.BranchPicker = &ui.BranchPickerState{
		Branches:          uiBranches,
		SelectedIdx:       0,
		PaneFocused:       true, // Start with list pane focused
		ListScrollOffset:  0,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}

	// Switch to branch picker mode
	app.workflowState.PreviousMode = app.mode // Track previous mode (Config)
	app.mode = ModeBranchPicker
	app.footerHint = "↑/↓ Navigate • Tab: Switch panes • Enter: Switch branch • ESC: Cancel"
	return nil
}

// dispatchConfigPreferences enters preferences mode
func (a *Application) dispatchConfigPreferences(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode // Track previous mode (Config)
	app.mode = ModePreferences
	app.selectedIndex = 0
	app.menuItems = app.GeneratePreferencesMenu()
	app.rebuildMenuShortcuts(ModePreferences)
	return nil
}

// ========================================
// Preferences Menu Dispatchers
// ========================================

// dispatchPreferencesToggleAutoUpdate toggles auto-update ON/OFF
func (a *Application) dispatchPreferencesToggleAutoUpdate(app *Application) tea.Cmd {
	if app.appConfig != nil {
		newValue := !app.appConfig.AutoUpdate.Enabled
		app.appConfig.SetAutoUpdateEnabled(newValue)

		if newValue {
			return app.startAutoUpdate()
		}
	}
	return nil
}

// dispatchPreferencesInterval is a no-op (interval adjusted via +/- keys)
func (a *Application) dispatchPreferencesInterval(app *Application) tea.Cmd {
	// Interval is adjusted via +/- keys, not enter
	// This dispatcher exists for SSOT completeness
	return nil
}

// dispatchPreferencesCycleTheme cycles to next available theme
func (a *Application) dispatchPreferencesCycleTheme(app *Application) tea.Cmd {
	if app.appConfig != nil {
		themes := config.GetAvailableThemes()
		if len(themes) > 0 {
			currentTheme := app.appConfig.Appearance.Theme
			nextIndex := 0
			for i, t := range themes {
				if t == currentTheme {
					nextIndex = (i + 1) % len(themes)
					break
				}
			}

			app.appConfig.SetTheme(themes[nextIndex])
			if newTheme, err := ui.LoadThemeByName(themes[nextIndex]); err == nil {
				app.theme = newTheme
				// Rebuild state info with new theme colors
				a.workingTreeInfo, a.timelineInfo, a.operationInfo = BuildStateInfo(newTheme)
			}
		}
	}
	return nil
}
