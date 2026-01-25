package config

import (
	"os"
	"path/filepath"

	"tit/internal"

	"github.com/pelletier/go-toml/v2"
)

// DefaultConfigTOML is the default configuration when no config file exists
const DefaultConfigTOML = `# TIT Configuration

[auto_update]
enabled = true
interval_minutes = 5

[appearance]
theme = "gfx"
`

// Config represents the application configuration
type Config struct {
	AutoUpdate AutoUpdateConfig `toml:"auto_update"`
	Appearance AppearanceConfig `toml:"appearance"`
}

// AutoUpdateConfig contains settings for background sync
type AutoUpdateConfig struct {
	Enabled         bool `toml:"enabled"`
	IntervalMinutes int  `toml:"interval_minutes"`
}

// AppearanceConfig contains visual settings
type AppearanceConfig struct {
	Theme string `toml:"theme"`
}

// GetConfigPath returns the path to the config file
// CONTRACT: returns error if UserHomeDir fails (fail-fast)
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "tit", "config.toml"), nil
}

// GetConfigDir returns the directory containing the config file
// CONTRACT: returns error if UserHomeDir fails (fail-fast)
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "tit"), nil
}

// Load loads the configuration from the config file
// CONTRACT: fail-fast on UserHomeDir errors; create default and persist if missing
// Returns structured error on parse/read failures (fail-fast)
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err // FAIL-FAST: UserHomeDir error
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File missing: create default and attempt to save it
		defaultConfig := &Config{
			AutoUpdate: AutoUpdateConfig{
				Enabled:         true,
				IntervalMinutes: 5,
			},
			Appearance: AppearanceConfig{
				Theme: "gfx",
			},
		}
		// Attempt to save; if it fails, still return config but surface error
		if saveErr := Save(defaultConfig); saveErr != nil {
			// Log/return error but allow app to continue with in-memory default
			return defaultConfig, saveErr
		}
		return defaultConfig, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// FAIL-FAST: propagate read error
		return nil, err
	}

	// Parse TOML
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		// FAIL-FAST: propagate parse error
		return nil, err
	}

	// Apply defaults for missing fields
	if config.AutoUpdate.IntervalMinutes == 0 {
		config.AutoUpdate.IntervalMinutes = 5
	}
	if config.Appearance.Theme == "" {
		config.Appearance.Theme = "gfx"
	}

	return &config, nil
}

// Save saves the configuration to the config file
func Save(config *Config) error {
	// Create directory if it doesn't exist
	configDir, err := GetConfigDir()
	if err != nil {
		return err // FAIL-FAST: UserHomeDir error
	}
	if err := os.MkdirAll(configDir, internal.StashDirPerms); err != nil {
		return err
	}

	// Marshal to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// Write to file
	configPath, err := GetConfigPath()
	if err != nil {
		return err // FAIL-FAST: UserHomeDir error
	}
	return os.WriteFile(configPath, data, internal.GitignorePerms)
}

// SetAutoUpdateEnabled sets the auto-update enabled state and persists
func (c *Config) SetAutoUpdateEnabled(enabled bool) error {
	c.AutoUpdate.Enabled = enabled
	return Save(c)
}

// SetAutoUpdateInterval sets the auto-update interval in minutes and persists
func (c *Config) SetAutoUpdateInterval(minutes int) error {
	if minutes < 1 {
		minutes = 1
	}
	if minutes > 60 {
		minutes = 60
	}
	c.AutoUpdate.IntervalMinutes = minutes
	return Save(c)
}

// SetTheme sets the theme and persists
func (c *Config) SetTheme(theme string) error {
	c.Appearance.Theme = theme
	return Save(c)
}
