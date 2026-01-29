package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdClone executes `git clone` with streaming output
func (a *Application) cmdClone(url, targetPath string) tea.Cmd {
	u := url // Capture in closure
	path := targetPath
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git clone with streaming output
		result := git.ExecuteWithStreaming(ctx, "clone", u, path)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpClone,
				Success: false,
				Error:   "Failed to clone repository",
			}
		}

		// Change to cloned directory if needed
		if path != "." {
			if err := os.Chdir(path); err != nil {
				return GitOperationMsg{
					Step:    OpClone,
					Success: false,
					Error:   fmt.Sprintf("Failed to change directory: %v", err),
				}
			}
		}

		// Verify .git exists
		cwd, _ := os.Getwd()
		gitDir := filepath.Join(cwd, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			return GitOperationMsg{
				Step:    OpClone,
				Success: false,
				Error:   fmt.Sprintf("Clone completed but .git not found"),
			}
		}

		return GitOperationMsg{
			Step:    OpClone,
			Success: true,
			Output:  "Repository cloned successfully",
		}
	}
}
