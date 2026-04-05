package app

import (
	"fmt"
	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"
)

// MenuItem is defined in the ui package for cross-package type safety
type MenuItem = ui.MenuItem

// MenuGenerator is a function type that generates menu items
type MenuGenerator func(*Application) []MenuItem

// GenerateMenu produces menu items based on current git state
func (a *Application) GenerateMenu() []MenuItem {
	// Priority 1: Operation State (most restrictive)
	if a.gitState == nil {
		return []MenuItem{}
	}

	menuGenerators := map[git.Operation]MenuGenerator{
		git.NotRepo:        (*Application).menuNotRepo,
		git.Normal:         (*Application).menuNormal,
		git.TimeTraveling:  (*Application).menuTimeTraveling,
		git.Conflicted:     (*Application).menuConflicted,
		git.Merging:        (*Application).menuMerging,
		git.Rebasing:       (*Application).menuRebasing,
		git.DirtyOperation: (*Application).menuDirtyOperation,
		git.Rewinding:      (*Application).menuNormal, // Rewinding is transient — by render time it's done
	}

	if generator, exists := menuGenerators[a.gitState.Operation]; exists {
		return generator(a)
	}

	// Unknown operation: fail fast with clear error
	panic(fmt.Sprintf("Unknown git operation state: %v", a.gitState.Operation))
}
