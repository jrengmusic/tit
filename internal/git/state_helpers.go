package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tit/internal"
)

// IsInitializedRepo checks if current working directory is a git repository
func IsInitializedRepo() (bool, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, ""
	}

	gitDir := filepath.Join(cwd, internal.GitDirectoryName)
	if _, err := os.Stat(gitDir); err == nil {
		return true, cwd
	}

	return false, ""
}

// HasParentRepo checks if any parent directory contains .git
func HasParentRepo() (bool, string) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, ""
	}

	parent := filepath.Dir(cwd)
	for parent != cwd {
		gitDir := filepath.Join(parent, internal.GitDirectoryName)
		if _, err := os.Stat(gitDir); err == nil {
			return true, parent
		}

		cwd = parent
		parent = filepath.Dir(parent)
	}

	return false, ""
}

// detectDirtyOperation checks if a dirty operation is in progress
// by looking for the .git/TIT_DIRTY_OP snapshot file
func detectDirtyOperation() bool {
	return IsDirtyOperationActive()
}

// GetTimeTravelInfo reads the .git/TIT_TIME_TRAVEL file and returns the original branch
// Returns: originalBranch, stashID, error
func GetTimeTravelInfo() (string, string, error) {
	gitDir := internal.GitDirectoryName
	filePath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read time travel info: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 1 {
		return "", "", fmt.Errorf("invalid time travel info format")
	}

	originalBranch := strings.TrimSpace(lines[0])
	stashID := ""
	if len(lines) >= 2 {
		stashID = strings.TrimSpace(lines[1])
	}

	return originalBranch, stashID, nil
}

// WriteTimeTravelInfo writes the .git/TIT_TIME_TRAVEL file with original branch and optional stash ID
func WriteTimeTravelInfo(originalBranch, stashID string) error {
	gitDir := internal.GitDirectoryName
	filePath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	content := originalBranch + "\n"
	if stashID != "" {
		content += stashID + "\n"
	}

	err := os.WriteFile(filePath, []byte(content), internal.GitignorePerms)
	if err != nil {
		return fmt.Errorf("failed to write time travel info: %w", err)
	}

	return nil
}

// ClearTimeTravelInfo removes the .git/TIT_TIME_TRAVEL file
func ClearTimeTravelInfo() error {
	gitDir := internal.GitDirectoryName
	filePath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear time travel info: %w", err)
	}

	return nil
}

// LoadTimeTravelInfo loads time travel metadata from .git/TIT_TIME_TRAVEL
// Returns nil if marker doesn't exist (normal case)
func LoadTimeTravelInfo() (*TimeTravelInfo, error) {
	gitDir := internal.GitDirectoryName
	markerPath := filepath.Join(gitDir, "TIT_TIME_TRAVEL")

	// Check if marker exists
	if !FileExists(markerPath) {
		return nil, nil
	}

	// Read marker file as plain text (simpler format)
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read time travel marker: %w", err)
	}

	// Parse two lines: branch and optional stash ID
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("time travel marker is empty")
	}

	branch := strings.TrimSpace(lines[0])
	stashID := ""
	if len(lines) > 1 {
		stashID = strings.TrimSpace(lines[1])
	}

	// Get current commit info (we're in detached HEAD during time travel)
	currentHash, err := executeGitCommand("rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit hash: %w", err)
	}
	currentHash = strings.TrimSpace(currentHash)

	// Get current commit metadata (subject and time)
	currentSubject, err := executeGitCommand("log", "-1", "--format=%s", currentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit subject: %w", err)
	}
	currentSubject = strings.TrimSpace(currentSubject)

	currentTimeStr, err := executeGitCommand("log", "-1", "--format=%aI", currentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit time: %w", err)
	}
	currentTime, err := time.Parse(time.RFC3339, strings.TrimSpace(currentTimeStr))
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit time: %w", err)
	}

	return &TimeTravelInfo{
		OriginalBranch:  branch,
		OriginalStashID: stashID,
		CurrentCommit: CommitInfo{
			Hash:    currentHash,
			Subject: currentSubject,
			Time:    currentTime,
		},
	}, nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
