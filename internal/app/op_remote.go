package app

import (
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdAddRemote adds a remote repository (step 1 of 3-step chain)
func (a *Application) cmdAddRemote(url string) tea.Cmd {
	u := url // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Get current branch BEFORE adding remote (works even with zero commits)
		// Use symbolic-ref instead of rev-parse: works when HEAD exists but has no commits
		branchResult := git.Execute("symbolic-ref", "--short", "HEAD")
		branchName := ""
		if branchResult.Success {
			branchName = strings.TrimSpace(branchResult.Stdout)
		}

		// Add remote
		result := git.ExecuteWithStreaming("remote", "add", "origin", u)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpAddRemote,
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		return GitOperationMsg{
			Step:       OpAddRemote,
			Success:    true,
			Output:     "Remote added",
			BranchName: branchName,
		}
	}
}

// cmdFetchRemote fetches from origin (step 2 of 3-step chain)
func (a *Application) cmdFetchRemote() tea.Cmd {
	return a.executeGitOp(OpFetchRemote, "fetch", "--all")
}

// cmdSetUpstream sets upstream tracking (step 3 of 3-step chain)
// Takes branchName from previous step to avoid querying git again
// If remote branch doesn't exist, this will push -u to create it
func (a *Application) cmdSetUpstream(branchName string) tea.Cmd {
	branch := branchName // Capture in closure
	return func() tea.Msg {
		result := git.SetUpstreamTrackingWithBranch(branch)
		if !result.Success {
			// FAIL-FAST: upstream setup failed
			return GitOperationMsg{
				Step:    OpSetUpstream,
				Success: false,
				Error:   result.Stderr,
			}
		}

		// Distinguish between "set upstream" vs "pushed to create"
		output := "Remote added and upstream configured"
		if result.Stdout == "pushed_and_upstream_set" {
			output = "Remote added and initial push completed"
		}

		return GitOperationMsg{
			Step:    OpSetUpstream,
			Success: true,
			Output:  output,
		}
	}
}
