package git

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ExecuteTimeTravelCheckout performs a time travel checkout operation
// Creates .git/TIT_TIME_TRAVEL file with original branch info
// Returns: TimeTravelCheckoutMsg
func ExecuteTimeTravelCheckout(originalBranch, commitHash string) func() tea.Msg {
	return func() tea.Msg {
		Log(fmt.Sprintf("Time traveling to commit %s...", commitHash))

		// Get current branch (for verification)
		currentBranchResult := Execute("rev-parse", "--abbrev-ref", "HEAD")
		if !currentBranchResult.Success {
			Error(fmt.Sprintf("Error getting current branch: %s", currentBranchResult.Stderr))
			return TimeTravelCheckoutMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				CommitHash:     commitHash,
				Error:          fmt.Sprintf("Failed to get current branch: %s", currentBranchResult.Stderr),
			}
		}

		// Checkout the target commit
		checkoutResult := Execute("checkout", commitHash)
		if !checkoutResult.Success {
			Error(fmt.Sprintf("Error checking out commit: %s", checkoutResult.Stderr))
			return TimeTravelCheckoutMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				CommitHash:     commitHash,
				Error:          fmt.Sprintf("Failed to checkout commit: %s", checkoutResult.Stderr),
			}
		}

		// Write time travel info file
		err := WriteTimeTravelInfo(originalBranch, "")
		if err != nil {
			Error(fmt.Sprintf("Error writing time travel info: %v", err))
			// Try to checkout back to original branch
			Execute("checkout", originalBranch)
			return TimeTravelCheckoutMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				CommitHash:     commitHash,
				Error:          fmt.Sprintf("Failed to write time travel info: %v", err),
			}
		}

		Log("Time travel successful")
		return TimeTravelCheckoutMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			CommitHash:     commitHash,
			Error:          "",
		}
	}
}
