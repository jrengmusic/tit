package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// GetRemoteDefaultBranch returns the default branch from origin
func GetRemoteDefaultBranch() string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	// Output is like "refs/remotes/origin/main"
	ref := strings.TrimSpace(string(output))
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
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
	if err := os.WriteFile(".gitignore", []byte(DefaultGitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}
