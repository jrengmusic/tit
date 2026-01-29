package internal

// Package constants holds all magic values used throughout TIT codebase.
// This provides single source of truth (SSOT) for magic values,
// making code more maintainable and self-documenting.

import "time"

// Version information (SSOT)
const (
	AppName    = "TIT"    // Application name
	AppVersion = "v1.1.1" // Application version (semantic versioning)
)

// Bit sizes for strconv parsing functions
const (
	FloatParseBitSize = 64 // ParseFloat bit size (float64)
	IntParseBitSize   = 64 // ParseInt bit size (int64)
)

// File permissions (octal)
const (
	SSHDirPerms     = 0700 // rwx------ - SSH directory (owner only)
	ConfigFilePerms = 0600 // rw------- - Config file (owner only)
	GitignorePerms  = 0644 // rw-r--r-- - .gitignore file (owner read/write, others read)
	StashDirPerms   = 0755 // rwxr-xr-x - Stash directory (owner rwx, group/others rx)
)

// Timestamp formats for git and display
const (
	GitTimestampFormat     = "2006-01-02 15:04:05 -0700"      // Git commit timestamp format
	DisplayTimestampFormat = "Mon, 2 Jan 2006 15:04:05 -0700" // Human-readable format
)

// UI constants
const (
	BezierCurveResolution = 20              // Resolution for cubic bezier curve approximation
	QuitConfirmTimeout    = 3 * time.Second // Timeout for quit/clear confirmation dialogs
)

// Git directory name
const (
	GitDirectoryName = ".git" // Git metadata directory name
)
