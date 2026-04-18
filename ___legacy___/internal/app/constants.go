package app

import "time"

// Application-level constants for magic numbers used throughout the codebase.
// These provide single source of truth (SSOT) for timeouts, limits, and dimensions.

const (
	// Timeouts
	CacheRefreshInterval = 100 * time.Millisecond // Cache refresh interval for UI updates
	PasteBurstWindow     = 50 * time.Millisecond  // Suppress raw key events after paste detection

	// UI Dimensions
	InputHeight = 4 // Default input height (label + 3-line box)

	// Scroll
	PageScrollLines = 10 // Lines per page scroll in console view

	// Input action identifiers
	InputActionCloneURL = "clone_url"

	// Branch picker purpose identifiers
	BranchPickerPurposeMerge = "merge"
)
