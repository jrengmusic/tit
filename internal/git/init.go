package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"tit/internal"
)

// DefaultGitignoreContent contains common patterns to ignore
const DefaultGitignoreContent = `# macOS
.DS_Store
.AppleDouble
.LSOverride
._*

# Windows
Thumbs.db
ehthumbs.db
Desktop.ini

# Linux
*~
.directory

# IDEs
.vscode/
.idea/
*.swp
*.swo
*~

# Build artifacts (common)
*.o
*.a
*.so
*.dylib
*.exe
*.dll

# Logs
*.log
`

// InitializeRepository initializes a git repository in the given directory
// WORKER THREAD - called from git operations, must be in worker goroutine
func InitializeRepository(dirPath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, internal.StashDirPerms); err != nil {
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
// WORKER THREAD - called from git operations, must be in worker goroutine
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
// WORKER THREAD - called from git operations, must be in worker goroutine
func CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %w\n%s", branchName, err, string(output))
	}
	return nil
}

// ListBranches returns all local branches
func ListBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetRemoteDefaultBranch queries the remote HEAD symref WITHOUT fetch
// Returns the default branch name from the remote, or error if undeterminable
// This uses git ls-remote to read the symbolic ref directly from remote
// FAIL FAST: If we can't determine the branch, we fail loudly
func GetRemoteDefaultBranch() (string, error) {
	// Use git ls-remote --symref to read origin/HEAD symref directly from remote
	// Format: "ref: refs/heads/main	HEAD" or just "HEAD" without ref line
	cmd := exec.Command("git", "ls-remote", "--symref", "origin", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to query remote HEAD: %w (output: %s)", err, string(output))
	}

	// Parse the output
	// Line 1 is either "ref: refs/heads/main\tHEAD" or empty
	// Line 2 is "<hash>\tHEAD" (or the only line if no symref found)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	// FAIL FAST: Must have at least one line
	if len(lines) == 0 {
		return "", fmt.Errorf("empty response from git ls-remote --symref origin HEAD - remote may not be a git repository")
	}

	// Look for the "ref: refs/heads/XXX" line
	for _, line := range lines {
		if strings.HasPrefix(line, "ref:") {
			// Format: "ref: refs/heads/main	HEAD"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				refPath := parts[1] // "refs/heads/main"

				// Extract branch name from "refs/heads/main" → "main"
				refParts := strings.Split(refPath, "/")
				if len(refParts) >= 3 && refParts[0] == "refs" && refParts[1] == "heads" {
					branch := strings.Join(refParts[2:], "/") // Handle branch names with slashes

					if branch == "" {
						return "", fmt.Errorf("invalid remote HEAD ref: %s", refPath)
					}

					return branch, nil
				}
			}
		}
	}

	// FAIL FAST: If we get here, the remote doesn't have a symbolic HEAD
	return "", fmt.Errorf("remote HEAD is not a symbolic ref - cannot determine default branch. Output was: %s", string(output))
}

// ListRemoteBranches returns all remote branches (without remote prefix)
func ListRemoteBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "HEAD") {
			// Remove origin/ prefix
			if strings.HasPrefix(line, "origin/") {
				line = strings.TrimPrefix(line, "origin/")
			}
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// CreateDefaultGitignore creates a .gitignore file with common patterns
// Used during clone-to-cwd to ignore OS and IDE garbage files
func CreateDefaultGitignore() error {
	// Check if .gitignore already exists
	if _, err := os.Stat(".gitignore"); err == nil {
		// .gitignore exists, don't overwrite it
		return nil
	}

	// Create .gitignore with default patterns
	if err := os.WriteFile(".gitignore", []byte(DefaultGitignoreContent), internal.GitignorePerms); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}
