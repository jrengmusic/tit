package app

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ActionHandler is a function type for action dispatchers
type ActionHandler func(*Application) tea.Cmd

// isCwdEmpty checks if current working directory is empty
// Ignores macOS metadata files (.DS_Store)
// Used for smart dispatch in init/clone workflows
func isCwdEmpty() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return false
	}

	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if name != ".DS_Store" && name != ".AppleDouble" {
			count++
		}
	}

	return count == 0
}

// dispatchAction routes menu item selections to appropriate handlers
func (a *Application) dispatchAction(actionID string) tea.Cmd {
	actionDispatchers := map[string]ActionHandler{
		"init":                      a.dispatchInit,
		"clone":                     a.dispatchClone,
		"add_remote":                a.dispatchAddRemote,
		"commit":                    a.dispatchCommit,
		"commit_push":               a.dispatchCommitPush,
		"push":                      a.dispatchPush,
		"force_push":                a.dispatchForcePush,
		"pull_merge":                a.dispatchPullMerge,
		"pull_merge_diverged":       a.dispatchPullMerge,
		"dirty_pull_merge":          a.dispatchDirtyPullMerge,
		"replace_local":             a.dispatchReplaceLocal,
		"reset_discard_changes":     a.dispatchResetDiscardChanges,
		"history":                   a.dispatchHistory,
		"file_history":              a.dispatchFileHistory,
		"time_travel_history":       a.dispatchTimeTravelHistory,
		"time_travel_files_history": a.dispatchFileHistory,
		"time_travel_merge":         a.dispatchTimeTravelMerge,
		"time_travel_return":        a.dispatchTimeTravelReturn,
		// Config menu actions
		"config_add_remote":         a.dispatchConfigAddRemote,
		"config_switch_remote":      a.dispatchConfigSwitchRemote,
		"config_remove_remote":      a.dispatchConfigRemoveRemote,
		"config_toggle_auto_update": a.dispatchConfigToggleAutoUpdate,
		"config_switch_branch":      a.dispatchConfigSwitchBranch,
		"config_preferences":        a.dispatchConfigPreferences,
		// Preferences menu actions
		"preferences_auto_update": a.dispatchPreferencesToggleAutoUpdate,
		"preferences_interval":    a.dispatchPreferencesInterval,
		"preferences_theme":       a.dispatchPreferencesCycleTheme,
	}

	if handler, exists := actionDispatchers[actionID]; exists {
		return handler(a)
	}
	return nil
}

// dispatchInit starts the repository initialization workflow
func (a *Application) dispatchInit(app *Application) tea.Cmd {
	cwdEmpty := isCwdEmpty()
	if !cwdEmpty {
		// CWD not empty: ask for subdirectory name first
		return a.cmdInitSubdirectory()
	}
	// CWD empty: ask for branch name directly, init here
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: "Initial branch name:",
		InputAction: "init_branch_name",
		FooterHint:  "Enter branch name (default: main), press Enter to initialize",
	})
	app.inputState.Value = "main"
	app.inputState.CursorPosition = len("main")
	return nil
}

// dispatchClone starts the clone workflow
func (a *Application) dispatchClone(app *Application) tea.Cmd {
	cwdEmpty := isCwdEmpty()
	if cwdEmpty {
		app.transitionTo(ModeTransition{
			Mode:        ModeCloneLocation,
			ResetFields: []string{"clone"},
		})
	} else {
		app.workflowState.CloneMode = "subdir"
		app.transitionTo(ModeTransition{
			Mode:        ModeCloneURL,
			InputPrompt: InputMessages["clone_url"].Prompt,
			InputAction: "clone_url",
			FooterHint:  InputMessages["clone_url"].Hint,
			ResetFields: []string{"clone"},
		})
	}
	return nil
}

// dispatchAddRemote starts the add remote workflow
func (a *Application) dispatchAddRemote(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputMessages["remote_url"].Prompt,
		InputAction: "add_remote_url",
		FooterHint:  InputMessages["remote_url"].Hint,
		ResetFields: []string{},
	})
	return nil
}

// dispatchCommit starts the commit workflow
func (a *Application) dispatchCommit(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputMessages["commit_message"].Prompt,
		InputAction: "commit_message",
		FooterHint:  InputMessages["commit_message"].Hint,
		InputHeight: app.sizing.TerminalHeight - ui.FooterHeight,
		ResetFields: []string{},
	})
	return nil
}

// dispatchCommitPush starts commit+push workflow
func (a *Application) dispatchCommitPush(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputMessages["commit_message"].Prompt,
		InputAction: "commit_push_message",
		FooterHint:  "Enter commit message (will commit and push)",
		InputHeight: app.sizing.TerminalHeight - ui.FooterHeight,
		ResetFields: []string{},
	})
	return nil
}

// dispatchPush pushes to remote
func (a *Application) dispatchPush(app *Application) tea.Cmd {
	a.startAsyncOp()
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0
	a.mode = ModeConsole
	a.consoleState.Reset()
	return app.cmdPush()
}

// dispatchPullMerge pulls with merge strategy
func (a *Application) dispatchPullMerge(app *Application) tea.Cmd {
	confirmType := string(ConfirmPullMerge)
	if app.gitState.Timeline == git.Diverged {
		confirmType = string(ConfirmPullMergeDiverged)
	}
	app.workflowState.PreviousMode = app.mode // Track previous mode (Menu)
	app.mode = ModeConfirmation
	msg := ConfirmationMessages[confirmType]
	app.confirmationDialog = ui.NewConfirmationDialog(
		ui.ConfirmationConfig{
			Title:       msg.Title,
			Explanation: msg.Explanation,
			YesLabel:    msg.YesLabel,
			NoLabel:     msg.NoLabel,
			ActionID:    confirmType,
		},
		a.sizing.ContentInnerWidth,
		&app.theme,
	)
	app.confirmationDialog.SelectNo()
	return nil
}

// dispatchForcePush shows confirmation dialog for force push
func (a *Application) dispatchForcePush(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode // Track previous mode (Menu)
	app.mode = ModeConfirmation
	app.confirmContext = map[string]string{}
	msg := ConfirmationMessages["force_push"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "force_push",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	return nil
}

// dispatchReplaceLocal shows confirmation dialog for destructive action
func (a *Application) dispatchReplaceLocal(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.confirmContext = map[string]string{}
	msg := ConfirmationMessages["hard_reset"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "hard_reset",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	return nil
}

// dispatchResetDiscardChanges shows confirmation dialog for discarding all changes
func (a *Application) dispatchResetDiscardChanges(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.confirmContext = map[string]string{}
	msg := ConfirmationMessages["hard_reset"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "hard_reset",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	return nil
}

// dispatchHistory shows commit history
func (a *Application) dispatchHistory(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode               // Track previous mode (Menu)
	app.workflowState.PreviousMenuIndex = app.selectedIndex // Track previous selection
	app.mode = ModeHistory

	var commits []ui.CommitInfo
	for hash, details := range app.cacheManager.GetAllMetadata() {
		commitTime, _ := parseCommitDate(details.Date)
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    commitTime,
		})
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	app.historyState = &ui.HistoryState{
		Commits:           commits,
		SelectedIdx:       0,
		PaneFocused:       true,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}
	return nil
}

// dispatchFileHistory shows file(s) history
func (a *Application) dispatchFileHistory(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode               // Track previous mode (Menu)
	app.workflowState.PreviousMenuIndex = app.selectedIndex // Track previous selection
	app.mode = ModeFileHistory

	var commits []ui.CommitInfo
	for hash, details := range app.cacheManager.GetAllMetadata() {
		commitTime, _ := parseCommitDate(details.Date)
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    commitTime,
		})
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	var files []ui.FileInfo
	if len(commits) > 0 {
		firstCommitHash := commits[0].Hash
		if gitFileList, exists := app.cacheManager.GetFiles(firstCommitHash); exists {
			for _, gitFile := range gitFileList {
				files = append(files, ui.FileInfo{
					Path:   gitFile.Path,
					Status: gitFile.Status,
				})
			}
		}
	}

	app.fileHistoryState = &ui.FileHistoryState{
		Commits:           commits,
		Files:             files,
		SelectedCommitIdx: 0,
		SelectedFileIdx:   0,
		FocusedPane:       ui.PaneCommits,
		CommitsScrollOff:  0,
		FilesScrollOff:    0,
		DiffScrollOff:     0,
		DiffLineCursor:    0,
		DiffContent:       "",
		VisualModeActive:  false,
		VisualModeStart:   0,
	}
	a.updateFileHistoryDiff()
	return nil
}

func parseCommitDate(dateStr string) (time.Time, error) {
	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", dateStr)
}

// dispatchDirtyPullMerge starts the dirty pull confirmation dialog
func (a *Application) dispatchDirtyPullMerge(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.confirmContext = map[string]string{}
	msg := ConfirmationMessages["dirty_pull"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "dirty_pull",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	return nil
}

// dispatchTimeTravelHistory handles the "Browse History" action during time travel
func (a *Application) dispatchTimeTravelHistory(app *Application) tea.Cmd {
	app.mode = ModeHistory

	var commits []ui.CommitInfo
	for hash, details := range app.cacheManager.GetAllMetadata() {
		commitTime, _ := parseCommitDate(details.Date)
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    commitTime,
		})
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	app.historyState = &ui.HistoryState{
		Commits:           commits,
		SelectedIdx:       0,
		PaneFocused:       true,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}
	return nil
}

// dispatchTimeTravelMerge handles the "Merge back" action during time travel
func (a *Application) dispatchTimeTravelMerge(app *Application) tea.Cmd {
	statusResult := git.Execute("status", "--porcelain")
	hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""

	var confirmType ConfirmationType
	if hasDirtyTree {
		confirmType = ConfirmTimeTravelMergeDirty
	} else {
		confirmType = ConfirmTimeTravelMerge
	}

	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	msg := ConfirmationMessages[string(confirmType)]
	app.confirmationDialog = ui.NewConfirmationDialog(
		ui.ConfirmationConfig{
			Title:       msg.Title,
			Explanation: msg.Explanation,
			YesLabel:    msg.YesLabel,
			NoLabel:     msg.NoLabel,
			ActionID:    string(confirmType),
		},
		a.sizing.ContentInnerWidth,
		&app.theme,
	)
	app.confirmationDialog.SelectNo()
	return nil
}

// dispatchTimeTravelReturn handles the "Return without merge" action during time travel
func (a *Application) dispatchTimeTravelReturn(app *Application) tea.Cmd {
	statusResult := git.Execute("status", "--porcelain")
	hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""

	if hasDirtyTree {
		app.workflowState.PreviousMode = app.mode
		app.mode = ModeConfirmation
		app.confirmationDialog = ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       "Return to main with uncommitted changes",
				Explanation: "You have changes during time travel. Choose action:\n(Press ESC to cancel)",
				YesLabel:    "Merge changes",
				NoLabel:     "Discard changes",
				ActionID:    "time_travel_return_dirty_choice",
			},
			a.sizing.ContentInnerWidth,
			&app.theme,
		)
		app.confirmationDialog.SelectNo()
	} else {
		app.workflowState.PreviousMode = app.mode
		app.mode = ModeConfirmation
		msg := ConfirmationMessages[string(ConfirmTimeTravelReturn)]
		app.confirmationDialog = ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       msg.Title,
				Explanation: msg.Explanation,
				YesLabel:    msg.YesLabel,
				NoLabel:     msg.NoLabel,
				ActionID:    string(ConfirmTimeTravelReturn),
			},
			a.sizing.ContentInnerWidth,
			&app.theme,
		)
		app.confirmationDialog.SelectNo()
	}
	return nil
}

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
	app.confirmContext = map[string]string{}
	msg := ConfirmationMessages["remove_remote"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "config_remove_remote",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
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
	app.branchPickerState = &ui.BranchPickerState{
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
