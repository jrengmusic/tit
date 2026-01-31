package git

import (
	"context"
	"fmt"
)

// ResetHardAtCommit executes git reset --hard at specified commit
// WORKER THREAD - Must be called from async operation
// Streams output to buffer for real-time display
// If resetting to HEAD, also runs git clean -fd to remove untracked files
// Returns commit hash on success or error
func ResetHardAtCommit(ctx context.Context, commitHash string) (string, error) {
	if commitHash == "" {
		// Use empty string as marker - caller will check via error
		return "", fmt.Errorf("commit hash cannot be empty")
	}

	// Show short hash in console
	Log(fmt.Sprintf("Resetting to %s...", ShortenHash(commitHash)))

	// Execute git reset --hard <commit>
	// Output streamed to buffer by ExecuteWithStreaming
	result := ExecuteWithStreaming(ctx, "reset", "--hard", commitHash)

	if !result.Success {
		return "", fmt.Errorf("reset failed: %s", result.Stderr)
	}

	// If resetting to HEAD, clean untracked files to ensure truly clean working tree
	// (per TIT design: no distinction between tracked/untracked, all must be correct)
	isHeadReset := Execute("rev-parse", commitHash)
	currentHead := Execute("rev-parse", "HEAD")
	if isHeadReset.Success && currentHead.Success && isHeadReset.Stdout == currentHead.Stdout {
		Log("Cleaning untracked files...")
		cleanResult := ExecuteWithStreaming(ctx, "clean", "-fd")
		if !cleanResult.Success {
			Error(fmt.Sprintf("Warning: clean failed: %s", cleanResult.Stderr))
		}
	}

	return commitHash, nil
}
