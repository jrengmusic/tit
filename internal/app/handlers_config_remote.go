package app

import (
	"context"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Remote Configuration Handlers
// ========================================

// handleConfigAddRemoteURLSubmit handles URL input from config add remote menu
func (a *Application) handleConfigAddRemoteURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	url := app.inputState.Value
	if url == "" {
		app.footerHint = "URL cannot be empty"
		return app, nil
	}

	if !ui.ValidateRemoteURL(url) {
		app.footerHint = ui.GetRemoteURLError()
		return app, nil
	}

	result := git.Execute("remote", "get-url", "origin")
	if result.Success {
		app.footerHint = ErrorMessages["remote_already_exists_validation"]
		return app, nil
	}

	app.startAsyncOp()
	app.workflowState.PreviousMode = ModeConfig
	app.workflowState.PreviousMenuIndex = app.selectedIndex
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	return app, a.cmdConfigAddRemote(url)
}

// handleConfigSwitchRemoteURLSubmit handles URL input from config switch remote menu
func (a *Application) handleConfigSwitchRemoteURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	url := app.inputState.Value
	if url == "" {
		app.footerHint = "URL cannot be empty"
		return app, nil
	}

	if !ui.ValidateRemoteURL(url) {
		app.footerHint = ui.GetRemoteURLError()
		return app, nil
	}

	result := git.Execute("remote", "get-url", "origin")
	if !result.Success {
		app.footerHint = "No remote configured to switch"
		return app, nil
	}

	app.startAsyncOp()
	app.workflowState.PreviousMode = ModeConfig
	app.workflowState.PreviousMenuIndex = app.selectedIndex
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	return app, a.cmdConfigSwitchRemote(url)
}

// cmdConfigAddRemote adds a new remote from config menu
func (a *Application) cmdConfigAddRemote(url string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "add", "origin", url)
		if !result.Success {
			return GitOperationMsg{
				Step:    "config_add_remote",
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		return GitOperationMsg{
			Step:    "config_add_remote",
			Success: true,
			Output:  "Remote added successfully",
		}
	}
}

// cmdConfigSwitchRemote updates an existing remote URL
func (a *Application) cmdConfigSwitchRemote(url string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "set-url", "origin", url)
		if !result.Success {
			return GitOperationMsg{
				Step:    "config_switch_remote",
				Success: false,
				Error:   "Failed to update remote URL",
			}
		}

		return GitOperationMsg{
			Step:    "config_switch_remote",
			Success: true,
			Output:  "Remote URL updated successfully",
		}
	}
}

// cmdConfigRemoveRemote removes the origin remote
func (a *Application) cmdConfigRemoveRemote() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "remove", "origin")
		if !result.Success {
			return GitOperationMsg{
				Step:    "config_remove_remote",
				Success: false,
				Error:   "Failed to remove remote",
			}
		}

		return GitOperationMsg{
			Step:    "config_remove_remote",
			Success: true,
			Output:  "Remote removed successfully",
		}
	}
}
