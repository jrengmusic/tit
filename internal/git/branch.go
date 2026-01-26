package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"tit/internal"
)

// BranchDetails represents a single branch with all metadata
type BranchDetails struct {
	Name           string
	IsCurrent      bool
	LastCommitTime time.Time
	LastCommitHash string
	LastCommitSubj string
	Author         string
	TrackingRemote string // e.g., "origin/main", or "" if local only
	Ahead          int
	Behind         int
}

// ListBranchesWithDetails returns all local branches with metadata
func ListBranchesWithDetails() ([]BranchDetails, error) {
	cmd := exec.Command("git", "for-each-ref", "--sort=-committerdate", "refs/heads",
		"--format=%(refname:short)%09%(HEAD)%09%(committerdate:iso)%09%(objectname:short)%09%(subject)%09%(committerdate:short)%09%(authorname)%09%(upstream:short)%09%(upstream:track)")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := []BranchDetails{}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}

		isCurrent := parts[1] == "*"
		commitTime, _ := time.Parse(internal.GitTimestampFormat, parts[2])
		commitHash := parts[3]
		commitSubj := parts[4]
		author := parts[6]

		// Handle optional upstream fields (may not exist for local-only branches)
		trackingRemote := ""
		if len(parts) > 7 {
			trackingRemote = parts[7]
		}

		// Parse ahead/behind from track string (format: "[ahead 5, behind 2]")
		ahead, behind := 0, 0
		if len(parts) > 8 && parts[8] != "" {
			trackStr := parts[8]
			if strings.Contains(trackStr, "ahead") {
				fields := strings.Fields(trackStr)
				for i, f := range fields {
					if f == "ahead" && i+1 < len(fields) {
						ahead, _ = strconv.Atoi(strings.TrimSuffix(fields[i+1], ","))
					}
					if f == "behind" && i+1 < len(fields) {
						behind, _ = strconv.Atoi(fields[i+1])
					}
				}
			}
		}

		branches = append(branches, BranchDetails{
			Name:           parts[0],
			IsCurrent:      isCurrent,
			LastCommitTime: commitTime,
			LastCommitHash: commitHash,
			LastCommitSubj: commitSubj,
			Author:         author,
			TrackingRemote: trackingRemote,
			Ahead:          ahead,
			Behind:         behind,
		})
	}

	// Sort: current first, then by commit date (already sorted by git)
	currentIdx := -1
	for i, b := range branches {
		if b.IsCurrent {
			currentIdx = i
			break
		}
	}
	if currentIdx > 0 {
		current := branches[currentIdx]
		branches = append(branches[:currentIdx], branches[currentIdx+1:]...)
		branches = append([]BranchDetails{current}, branches...)
	}

	return branches, nil
}

// SwitchBranch performs git switch to target branch
func SwitchBranch(branchName string) error {
	cmd := exec.Command("git", "switch", branchName)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch to branch %s: %v", branchName, err)
	}
	return nil
}

// StashChanges stashes current changes
func StashChanges() error {
	cmd := exec.Command("git", "stash", "push", "-u")
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stash changes: %w", err)
	}
	return nil
}

// PopStash applies stashed changes
func PopStash() error {
	cmd := exec.Command("git", "stash", "pop")
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pop stash: %w", err)
	}
	return nil
}
