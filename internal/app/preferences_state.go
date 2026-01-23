package app

// PreferencesState holds state for the preferences editor mode
type PreferencesState struct {
	SelectedRow int // 0 = auto-update enabled, 1 = auto-update interval, 2 = theme
}
