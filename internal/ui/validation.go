package ui

import (
	"strings"
)

// ValidateRemoteURL validates a git remote URL format
// Returns true if valid, false otherwise
// Supports SSH, HTTPS, HTTP, and local paths
func ValidateRemoteURL(url string) bool {
	url = strings.TrimSpace(url)
	if url == "" {
		return false
	}

	// SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(url, "git@") {
		return strings.Contains(url, ":")
	}

	// HTTPS format: https://github.com/user/repo.git
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return len(url) > 8
	}

	// Local paths: /path/to/repo or ~/path/to/repo
	if strings.HasPrefix(url, "/") || strings.HasPrefix(url, "~") {
		return true
	}

	return false
}

// GetRemoteURLError returns user-friendly error message for invalid remote URL
func GetRemoteURLError() string {
	return "Invalid format. Expected:\n" +
		"  • SSH: git@github.com:user/repo.git\n" +
		"  • HTTPS: https://github.com/user/repo.git\n" +
		"  • Local: /path/to/repo or ~/path/to/repo"
}
