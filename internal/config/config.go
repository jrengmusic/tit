package config

import (
	"os"
	"path/filepath"

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
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "tit", "config.toml")
}

// GetConfigDir returns the directory containing the config file
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "tit")
}

// Load loads the configuration from the config file
// Returns default config if file doesn't exist or is invalid
func Load() (*Config, error) {
	configPath := GetConfigPath()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config and return it
		defaultConfig := &Config{
			AutoUpdate: AutoUpdateConfig{
				Enabled:         true,
				IntervalMinutes: 5,
			},
			Appearance: AppearanceConfig{
				Theme: "default",
			},
		}
		return defaultConfig, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return default on read error
		return createDefaultConfig()
	}

	// Parse TOML
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		// Return default on parse error
		return createDefaultConfig()
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

// createDefaultConfig creates the default config file and returns it
func createDefaultConfig() (*Config, error) {
	config := &Config{
		AutoUpdate: AutoUpdateConfig{
			Enabled:         true,
			IntervalMinutes: 5,
		},
		Appearance: AppearanceConfig{
			Theme: "gfx",
		},
	}

	// Create default config file
	if err := Save(config); err != nil {
		return config, nil // Return config anyway, file creation failed
	}

	return config, nil
}

// Save saves the configuration to the config file
func Save(config *Config) error {
	// Create directory if it doesn't exist
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// Write to file
	configPath := GetConfigPath()
	return os.WriteFile(configPath, data, 0644)
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
