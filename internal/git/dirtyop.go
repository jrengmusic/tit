package git

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DirtyOperationSnapshot stores the state before a dirty operation
// Used to restore the exact original state if the operation is aborted
type DirtyOperationSnapshot struct {
	OriginalBranch string // Branch name before operation
	OriginalHead   string // Commit hash before operation
}

// FilePath returns the path to the TIT_DIRTY_OP state file in .git/
func (s *DirtyOperationSnapshot) FilePath() string {
	return filepath.Join(".git", "TIT_DIRTY_OP")
}

// Save writes the snapshot to .git/TIT_DIRTY_OP
// Format:
//   Line 1: original branch name
//   Line 2: original commit hash
func (s *DirtyOperationSnapshot) Save(branchName, headHash string) error {
	if branchName == "" || headHash == "" {
		return fmt.Errorf("branch name and head hash cannot be empty")
	}

	s.OriginalBranch = branchName
	s.OriginalHead = headHash

	filePath := s.FilePath()
	content := fmt.Sprintf("%s\n%s\n", branchName, headHash)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save dirty operation snapshot: %w", err)
	}

	return nil
}

// Load reads the snapshot from .git/TIT_DIRTY_OP
// Returns error if file doesn't exist or is malformed
func (s *DirtyOperationSnapshot) Load() error {
	filePath := s.FilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no dirty operation in progress (snapshot file not found)")
		}
		return fmt.Errorf("failed to read dirty operation snapshot: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("corrupted snapshot file: expected 2 lines, got %d", len(lines))
	}

	s.OriginalBranch = strings.TrimSpace(lines[0])
	s.OriginalHead = strings.TrimSpace(lines[1])

	if s.OriginalBranch == "" || s.OriginalHead == "" {
		return fmt.Errorf("corrupted snapshot file: branch or head hash is empty")
	}

	return nil
}

// Delete removes the snapshot file (.git/TIT_DIRTY_OP)
// Does not error if file doesn't exist
func (s *DirtyOperationSnapshot) Delete() error {
	filePath := s.FilePath()
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete dirty operation snapshot: %w", err)
	}
	return nil
}

// IsDirtyOperationActive checks if a dirty operation is currently in progress
// by looking for the snapshot file
func IsDirtyOperationActive() bool {
	filePath := filepath.Join(".git", "TIT_DIRTY_OP")
	_, err := os.Stat(filePath)
	return err == nil
}

// ReadSnapshotState loads the snapshot without creating a DirtyOperationSnapshot struct
// Useful for state detection and cleanup
func ReadSnapshotState() (branchName, headHash string, err error) {
	filePath := filepath.Join(".git", "TIT_DIRTY_OP")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("no dirty operation in progress")
		}
		return "", "", fmt.Errorf("failed to read snapshot: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	if len(lines) < 2 {
		return "", "", fmt.Errorf("corrupted snapshot file")
	}

	return lines[0], lines[1], nil
}

// CleanupSnapshot removes the snapshot file (final cleanup after successful operation)
func CleanupSnapshot() error {
	filePath := filepath.Join(".git", "TIT_DIRTY_OP")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup snapshot: %w", err)
	}
	return nil
}
