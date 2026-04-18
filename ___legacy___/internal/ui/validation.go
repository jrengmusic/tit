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

// SanitizeCommitMessage strips all characters outside the safe ASCII set.
// Allows: printable ASCII (0x20-0x7E), newline (0x0A).
// Strips: control chars, \r, zero-width Unicode, BiDi overrides,
// Private Use Area (Nerd Font icons), and all non-ASCII codepoints.
// Collapses consecutive blank lines into a single blank line.
func SanitizeCommitMessage(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	prevWasNewline := false
	prevWasBlankLine := false

	for _, r := range text {
		if r == '\n' {
			if prevWasNewline && prevWasBlankLine {
				continue // collapse consecutive blank lines
			}
			if prevWasNewline {
				prevWasBlankLine = true
			}
			prevWasNewline = true
			b.WriteRune(r)
			continue
		}

		if r >= 0x20 && r <= 0x7E {
			prevWasNewline = false
			prevWasBlankLine = false
			b.WriteRune(r)
		}
		// Everything else is silently stripped
	}

	return strings.TrimSpace(b.String())
}

// InputValidator defines a function type for input validation.
// It returns true if the input is valid, and a message if it's not.
type InputValidator func(string) (bool, string)

// Validators provides a registry of reusable validation functions.
var Validators = map[string]InputValidator{
	"url": func(s string) (bool, string) {
		if s == "" {
			return false, "Repository URL cannot be empty"
		}
		if !ValidateRemoteURL(s) {
			return false, "Invalid URL format. Try: git@github.com:user/repo.git"
		}
		return true, ""
	},
	"branch_name": func(s string) (bool, string) {
		if s == "" {
			return false, "Branch name cannot be empty"
		}
		if strings.Contains(s, " ") {
			return false, "Branch name cannot contain spaces"
		}
		return true, ""
	},
	"directory": func(s string) (bool, string) {
		if s == "" {
			return false, "Directory name cannot be empty"
		}
		return true, ""
	},
}
