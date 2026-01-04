package git

import (
	"fmt"
	"os"
	"os/exec"
)

// InitializeRepository initializes a git repository in the given directory
// AUDIO THREAD - Called from git operations, must be in worker goroutine
func InitializeRepository(dirPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Change to directory
	originalCwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(dirPath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	defer os.Chdir(originalCwd)

	// Run git init
	cmd := exec.Command("git", "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\n%s", err, string(output))
	}

	return nil
}

// CreateBranch creates a new git branch (or checks out if it exists as tracking)
// AUDIO THREAD - Called from git operations, must be in worker goroutine
func CreateBranch(branchName string) error {
	// Check if branch already exists
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	if err := cmd.Run(); err == nil {
		// Branch exists, just checkout
		checkoutCmd := exec.Command("git", "checkout", branchName)
		if output, err := checkoutCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w\n%s", branchName, err, string(output))
		}
		return nil
	}

	// Branch doesn't exist, create it
	createCmd := exec.Command("git", "checkout", "-b", branchName)
	if output, err := createCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch %s: %w\n%s", branchName, err, string(output))
	}

	return nil
}

// CheckoutBranch checks out an existing branch
// AUDIO THREAD - Called from git operations, must be in worker goroutine
func CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w\n%s", branchName, err, string(output))
	}
	return nil
}
