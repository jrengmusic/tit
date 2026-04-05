package ui

// MenuItem represents a single menu action or separator
type MenuItem struct {
	ID            string
	Shortcut      string
	ShortcutLabel string
	Emoji         string
	Label         string
	Hint          string
	Enabled       bool
	Separator     bool
}
