package app

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// LocationChoiceConfig holds configuration for the generic location choice handler.
type LocationChoiceConfig struct {
	PathSetter   func(*Application, string)             // Sets the path in the application state.
	OnCurrentDir func(*Application) (tea.Model, tea.Cmd) // Action to perform after choosing the current directory.
	SubdirPrompt string                                 // Prompt for subdirectory input, e.g., "Repository name:".
	SubdirAction string                                 // Action for subdirectory input, e.g., "init_subdir_name".
}

// handleLocationChoice is a generic handler for choosing between the current directory
// and a new subdirectory.
func (a *Application) handleLocationChoice(choice int, config LocationChoiceConfig) (tea.Model, tea.Cmd) {
	if choice == 1 { // Corresponds to "current directory"
		cwd, err := os.Getwd()
		if err != nil {
			a.footerHint = ErrorMessages["cwd_read_failed"]
			return a, nil
		}
		config.PathSetter(a, cwd)
		return config.OnCurrentDir(a)
	}

	// Corresponds to "create subdirectory"
	a.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: config.SubdirPrompt,
		InputAction: config.SubdirAction,
		FooterHint:  InputHints["subdir_name"],
	})
	return a, nil
}
