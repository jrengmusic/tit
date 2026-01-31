package git

import (
	"fmt"
	"strings"
	"time"
	"tit/internal"
)

// FetchRecentCommits fetches the last N commits with basic info
// ref: git ref to log from (e.g., "main", "HEAD"). Empty string = HEAD (current position)
// Returns: []CommitInfo with Hash, Subject, Time populated
// Format: git log --pretty=%H%n%s%n%ai -N [ref]
// Example output: hash1\nsubject1\nISO-date1\nhash2\nsubject2\nISO-date2\n...
func FetchRecentCommits(limit int, ref string) ([]CommitInfo, error) {
	args := []string{"log", fmt.Sprintf("-%d", limit), "--pretty=%H%n%s%n%ai"}
	if ref != "" {
		args = append(args, ref)
	}
	result := Execute(args...)
	if !result.Success {
		return nil, fmt.Errorf("failed to fetch recent commits: %s", result.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	commits := make([]CommitInfo, 0)

	// Parse output: hash, subject, iso-date on consecutive lines
	for i := 0; i+2 < len(lines); i += 3 {
		hash := strings.TrimSpace(lines[i])
		subject := strings.TrimSpace(lines[i+1])
		dateStr := strings.TrimSpace(lines[i+2])

		if hash == "" {
			continue
		}

		// Parse ISO date format: YYYY-MM-DD HH:MM:SS ±HHMM
		parsedTime, err := time.Parse(internal.GitTimestampFormat, dateStr)
		if err != nil {
			// If parsing fails, use zero time but continue
			parsedTime = time.Time{}
		}

		commits = append(commits, CommitInfo{
			Hash:    hash,
			Subject: subject,
			Time:    parsedTime,
		})
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	return commits, nil
}

// GetCommitDetails fetches full metadata for a commit
// Returns: CommitDetails with Author, Date, Message
// Format: git show -s --pretty=%aN%n%aD%n%B <hash>
func GetCommitDetails(hash string) (*CommitDetails, error) {
	result := Execute("show", "-s", "--pretty=%aN%n%aD%n%B", hash)
	if !result.Success {
		return nil, fmt.Errorf("failed to get commit details: %s", result.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected commit details format")
	}

	// Parse: author name (line 0), date (line 1), message (lines 2+)
	author := strings.TrimSpace(lines[0])
	date := strings.TrimSpace(lines[1])
	message := strings.Join(lines[2:], "\n")

	return &CommitDetails{
		Author:  author,
		Date:    date,
		Message: strings.TrimSpace(message),
	}, nil
}

// GetFilesInCommit fetches list of files changed in a commit
// Returns: []FileInfo with Path and Status
// Format: git show --name-status --pretty= <hash>
// Status: M (modified), A (added), D (deleted), R (renamed), C (copied), T (type changed), U (unmerged)
func GetFilesInCommit(hash string) ([]FileInfo, error) {
	result := Execute("show", "--name-status", "--pretty=", hash)
	if !result.Success {
		return nil, fmt.Errorf("failed to get files in commit: %s", result.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	files := make([]FileInfo, 0)

	// Parse output: status\tpath (tab-separated)
	// Special case: Rename/copy has 3 fields: "R100\told\tnew"
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		status := strings.TrimSpace(parts[0])
		statusChar := string(status[0])

		// Handle rename/copy (3 fields: status, old path, new path)
		if (statusChar == "R" || statusChar == "C") && len(parts) == 3 {
			oldPath := strings.TrimSpace(parts[1])
			newPath := strings.TrimSpace(parts[2])

			// Add old path with deletion marker
			files = append(files, FileInfo{
				Path:   oldPath,
				Status: "-",
			})

			// Add new path with rename marker
			files = append(files, FileInfo{
				Path:   newPath,
				Status: "→",
			})
		} else {
			// Normal case: status + path
			path := strings.TrimSpace(parts[1])
			files = append(files, FileInfo{
				Path:   path,
				Status: statusChar,
			})
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in commit")
	}

	return files, nil
}

// GetCommitDiff fetches diff for a file in a commit
// version: "parent" (commit vs parent) or "wip" (commit vs working tree)
// Returns: unified diff content (plain text)
// Uses git diff with appropriate range based on version
func GetCommitDiff(hash, path, version string) (string, error) {
	var result CommandResult

	switch version {
	case "parent":
		// Compare commit with its parent
		result = Execute("diff", hash+"^", hash, "--", path)
	case "wip":
		// Compare commit with working tree
		result = Execute("diff", hash, "--", path)
	default:
		return "", fmt.Errorf("invalid diff version: %s (must be 'parent' or 'wip')", version)
	}

	if !result.Success {
		return "", fmt.Errorf("failed to get diff for %s: %s", path, result.Stderr)
	}

	return result.Stdout, nil
}
