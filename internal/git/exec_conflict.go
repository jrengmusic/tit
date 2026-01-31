package git

import (
	"fmt"
	"strings"
)

// ListConflictedFiles returns a list of files with merge conflicts
// Uses: git status --porcelain=v2 and filters for unmerged (u) entries
func ListConflictedFiles() ([]string, error) {
	result := Execute("status", "--porcelain=v2")
	if !result.Success {
		return nil, fmt.Errorf("failed to get status: %s", result.Stderr)
	}

	var conflictedFiles []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		if line == "" {
			continue
		}

		// Status line format: <status> <meta> <meta> ... <path>
		// Unmerged is marked with 'u' in second field
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		if parts[0] == "u" { // Unmerged status
			// Path starts at field 10 for unmerged entries (after 3 hashes)
			// Format: u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
			if len(parts) >= 11 {
				// Join fields from index 10 onwards (handles filenames with spaces)
				path := strings.Join(parts[10:], " ")
				conflictedFiles = append(conflictedFiles, path)
			}
		}
	}

	if len(conflictedFiles) == 0 {
		// Return empty slice instead of error - caller can check len to determine if conflicts exist
		return []string{}, nil
	}

	return conflictedFiles, nil
}

// ShowConflictVersion retrieves the content of a specific stage from a merge conflict
// stage: 1=base, 2=local (ours), 3=remote (theirs)
// Returns the file content as a string
func ShowConflictVersion(filePath string, stage int) (string, error) {
	if stage < 1 || stage > 3 {
		return "", fmt.Errorf("invalid stage: %d (must be 1-3)", stage)
	}

	stageRef := fmt.Sprintf(":%d:%s", stage, filePath)
	result := Execute("show", stageRef)

	if !result.Success {
		return "", fmt.Errorf("failed to show stage %d of %s: %s", stage, filePath, result.Stderr)
	}

	return result.Stdout, nil
}
