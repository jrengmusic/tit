package app

import (
	"fmt"
	"tit/internal/git"
)

// MenuItem represents a single menu action or separator
type MenuItem struct {
	ID            string // Unique identifier for the action
	Shortcut      string // Keyboard shortcut (actual key binding, e.g., "}", "{")
	ShortcutLabel string // Display label for shortcut (e.g., "shift + ]"), empty = use Shortcut
	Emoji         string // Leading emoji
	Label         string // Action name
	Hint          string // Plain language hint shown on focus
	Enabled       bool   // Whether this item can be selected
	Separator     bool   // If true, this is a visual separator (non-selectable)
}

// MenuGenerator is a function type that generates menu items
type MenuGenerator func(*Application) []MenuItem

// GenerateMenu produces menu items based on current git state
func (a *Application) GenerateMenu() []MenuItem {
	// Priority 1: Operation State (most restrictive)
	if a.gitState == nil {
		return []MenuItem{}
	}

	menuGenerators := map[git.Operation]MenuGenerator{
		git.NotRepo:       (*Application).menuNotRepo,
		git.Normal:        (*Application).menuNormal,
		git.TimeTraveling: (*Application).menuTimeTraveling,
	}

	if generator, exists := menuGenerators[a.gitState.Operation]; exists {
		return generator(a)
	}

	// Unknown operation: fail fast with clear error
	panic(fmt.Sprintf("Unknown git operation state: %v", a.gitState.Operation))
}
