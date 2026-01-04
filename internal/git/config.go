package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// DefaultRepoConfigTOML is the embedded default repo config
const DefaultRepoConfigTOML = `# TIT Repository Configuration
# This file stores repository metadata and settings

[repo]
initialized = false
repositoryPath = ""
canonBranch = "main"
lastWorkingBranch = "dev"
`

// RepoConfig represents the repository configuration
type RepoConfig struct {
	Repo struct {
		Initialized       bool   `toml:"initialized"`
		RepositoryPath    string `toml:"repositoryPath"`
		CanonBranch       string `toml:"canonBranch"`
		LastWorkingBranch string `toml:"lastWorkingBranch"`
	} `toml:"repo"`
}

// LoadRepoConfig loads the repo config from ~/.config/repo.toml
func LoadRepoConfig() (RepoConfig, error) {
	configFile := getRepoConfigPath()
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		return RepoConfig{}, fmt.Errorf("failed to read repo config: %w", err)
	}

	var config RepoConfig
	if err := toml.Unmarshal(fileData, &config); err != nil {
		return RepoConfig{}, fmt.Errorf("failed to parse repo config: %w", err)
	}

	return config, nil
}

// SaveRepoConfig saves the repo config to ~/.config/repo.toml
func SaveRepoConfig(config RepoConfig) error {
	configFile := getRepoConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal repo config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write repo config: %w", err)
	}

	return nil
}

// CreateDefaultRepoConfigIfMissing creates the default repo config on first run
func CreateDefaultRepoConfigIfMissing() (string, error) {
	configFile := getRepoConfigPath()

	if _, err := os.Stat(configFile); err == nil {
		return configFile, nil
	}

	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configFile, []byte(DefaultRepoConfigTOML), 0644); err != nil {
		return "", fmt.Errorf("failed to write default repo config: %w", err)
	}

	return configFile, nil
}

// getRepoConfigPath returns the path to ~/.config/tit/repo.toml
func getRepoConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "repo.toml"
	}
	return filepath.Join(home, ".config", "tit", "repo.toml")
}
