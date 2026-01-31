package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"tit/internal"
)

// EnsureFiveThemesExist creates/regenerates all 5 themes at startup
func EnsureFiveThemesExist() error {
	configThemeDir := filepath.Join(getConfigDirectory(), "themes")
	if err := os.MkdirAll(configThemeDir, internal.StashDirPerms); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	themes := map[string]string{
		"gfx":    GfxTheme,
		"spring": SpringTheme,
		"summer": SummerTheme,
		"autumn": AutumnTheme,
		"winter": WinterTheme,
	}

	for name, content := range themes {
		themePath := filepath.Join(configThemeDir, name+".toml")
		if err := os.WriteFile(themePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s theme: %w", name, err)
		}
	}

	return nil
}

// CreateDefaultThemeIfMissing creates or regenerates all 5 themes (gfx + 4 seasons)
// SSOT: Always regenerates from GfxTheme to ensure all colors are current
func CreateDefaultThemeIfMissing() (string, error) {
	return "", EnsureFiveThemesExist()
}

// LoadDefaultTheme loads the default gfx theme
func LoadDefaultTheme() (Theme, error) {
	themeFile := filepath.Join(getConfigDirectory(), "themes", "gfx.toml")
	return LoadTheme(themeFile)
}
