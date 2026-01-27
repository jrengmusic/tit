package app

import (
	"os/exec"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdPull pulls from remote (merge)
func (a *Application) cmdPull() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Pull with explicit --no-rebase to merge (required for diverged branches)
		result := git.ExecuteWithStreaming("pull", "--no-rebase")

		if !result.Success {
			// Check if we're in a conflicted state (more reliable than parsing stderr)
			// This detects merge conflicts by checking git state (.git/MERGE_HEAD + unmerged files)
			if msg := a.checkForConflicts(OpPull, false); msg != nil {
				return *msg
			}
			return GitOperationMsg{
				Step:    OpPull,
				Success: false,
				Error:   result.Stderr,
			}
		}

		return GitOperationMsg{
			Step:    OpPull,
			Success: true,
			Output:  "Pulled successfully",
		}
	}
}

// cmdHardReset executes git fetch + reset --hard origin/<branch> (ALWAYS get remote state)
func (a *Application) cmdHardReset() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Get current branch name
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			buffer.Append(ErrorMessages["failed_determine_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   "Could not determine current branch",
			}
		}

		branchName := strings.TrimSpace(string(output))
		buffer.Append("Fetching and resetting to remote state...", ui.TypeInfo)

		// Fetch first to ensure we have latest remote state
		fetchResult := git.ExecuteWithStreaming("fetch", "origin")
		if !fetchResult.Success {
			buffer.Append("Warning: fetch failed, using local remote refs", ui.TypeWarning)
		}

		// Reset to remote branch
		resetResult := git.ExecuteWithStreaming("reset", "--hard", "origin/"+branchName)
		if !resetResult.Success {
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   "Failed to reset to remote branch",
			}
		}

		return GitOperationMsg{
			Step:    OpHardReset,
			Success: true,
			Output:  "Reset to remote branch successfully",
		}
	}
}
