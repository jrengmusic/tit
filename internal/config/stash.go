package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tit/internal"

	"github.com/BurntSushi/toml"
)

// Stash operation type constants (SSOT)
const (
	StashPrefixTimeTravel = "TIT_TIME_TRAVEL"
	StashPrefixDirtyPull  = "TIT_DIRTY_PULL"
)

// StashEntry represents a single stash tracked by TIT
type StashEntry struct {
	Operation      string    `toml:"operation"`       // "time_travel", "dirty_pull"
	StashHash      string    `toml:"stash_hash"`      // Git stash commit hash
	CreatedAt      time.Time `toml:"created_at"`      // When stash was created
	RepoPath       string    `toml:"repo_path"`       // Absolute path to repository
	OriginalBranch string    `toml:"original_branch"` // Branch user was on
	CommitHash     string    `toml:"commit_hash"`     // Commit hash (for time travel)
}

// StashList holds all tracked stashes
type StashList struct {
	Stash []StashEntry `toml:"stash"`
}

// getStashFilePath returns absolute path to stash list TOML
func getStashFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to get home directory: %v", err))
	}
	return filepath.Join(homeDir, ".config", "tit", "stash", "list.toml")
}

// ensureStashDir creates the stash directory if it doesn't exist
func ensureStashDir() {
	stashFile := getStashFilePath()
	stashDir := filepath.Dir(stashFile)

	if err := os.MkdirAll(stashDir, internal.StashDirPerms); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to create stash directory %s: %v", stashDir, err))
	}
}

// loadStashList loads the stash list from TOML file
// Returns empty list if file doesn't exist (first run)
func loadStashList() *StashList {
	stashFile := getStashFilePath()

	// If file doesn't exist, return empty list (not an error - first run)
	if _, err := os.Stat(stashFile); os.IsNotExist(err) {
		return &StashList{Stash: []StashEntry{}}
	}

	var list StashList
	if _, err := toml.DecodeFile(stashFile, &list); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to parse stash list %s: %v", stashFile, err))
	}

	return &list
}

// saveStashList writes the stash list to TOML file
func saveStashList(list *StashList) {
	ensureStashDir()
	stashFile := getStashFilePath()

	f, err := os.Create(stashFile)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to create stash file %s: %v", stashFile, err))
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(list); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to write stash file %s: %v", stashFile, err))
	}
}

// AddStashEntry adds a new stash entry to the tracking list
// Panics if an entry already exists for this operation+repo (fail fast - detect bugs early)
func AddStashEntry(operation, stashHash, repoPath, originalBranch, commitHash string) {
	list := loadStashList()

	// Check for duplicate entry (should NEVER happen - indicates bug in caller)
	for _, entry := range list.Stash {
		if entry.Operation == operation && entry.RepoPath == repoPath {
			panic(fmt.Sprintf("FATAL: Stash entry already exists for operation=%s repo=%s. This is a bug - caller must check before adding.", operation, repoPath))
		}
	}

	// Add new entry
	entry := StashEntry{
		Operation:      operation,
		StashHash:      stashHash,
		CreatedAt:      time.Now(),
		RepoPath:       repoPath,
		OriginalBranch: originalBranch,
		CommitHash:     commitHash,
	}

	list.Stash = append(list.Stash, entry)
	saveStashList(list)
}

// GetStashEntry finds a stash entry by operation and repo path
// Returns (*entry, true) if found, (nil, false) if not found
func GetStashEntry(operation, repoPath string) (*StashEntry, bool) {
	list := loadStashList()

	for _, entry := range list.Stash {
		if entry.Operation == operation && entry.RepoPath == repoPath {
			// Return copy of entry (not pointer to loop variable)
			entryCopy := entry
			return &entryCopy, true
		}
	}

	return nil, false
}

// FindStashEntry finds a stash entry by operation and repo path
// Returns (hash, true) if found, ("", false) if not found
func FindStashEntry(operation, repoPath string) (string, bool) {
	entry, found := GetStashEntry(operation, repoPath)
	if !found {
		return "", false
	}
	return entry.StashHash, true
}

// RemoveStashEntry removes a stash entry from the tracking list
// Panics if entry doesn't exist (fail fast - detect bugs early)
func RemoveStashEntry(operation, repoPath string) {
	list := loadStashList()

	found := false
	newStash := []StashEntry{}

	for _, entry := range list.Stash {
		if entry.Operation == operation && entry.RepoPath == repoPath {
			found = true
			// Skip this entry (remove it)
			continue
		}
		newStash = append(newStash, entry)
	}

	if !found {
		panic(fmt.Sprintf("FATAL: Stash entry not found for operation=%s repo=%s. This is a bug - caller must ensure entry exists before removing.", operation, repoPath))
	}

	list.Stash = newStash
	saveStashList(list)
}
