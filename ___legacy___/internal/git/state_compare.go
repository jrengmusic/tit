package git

// CompareStates returns true if the two states would show different menu options
// This is used to determine if menu regeneration is necessary after a state update.
//
// Menu changes when:
// - Operation changes (affects available actions)
// - WorkingTree changes Clean<->Dirty (affects commit option)
// - Remote presence changes (affects push/pull options)
// - Timeline changes EXCEPT within same state (Ahead(n)->Ahead(m) doesn't change menu)
// - CurrentBranch changes (different branch = different state)
//
// Menu stays same when:
// - Ahead commits increase/decrease (Ahead(1)->Ahead(2) shows same menu)
// - Behind commits increase/decrease (Behind(1)->Behind(2) shows same menu)
func CompareStates(old, new *State) bool {
	// Operation changes always affect menu
	if old.Operation != new.Operation {
		return true
	}

	// WorkingTree: Clean->Dirty or Dirty->Clean affects menu
	if old.WorkingTree != new.WorkingTree {
		return true
	}

	// Remote presence affects menu
	if old.Remote != new.Remote {
		return true
	}

	// Timeline changes affect menu EXCEPT when both are Ahead or both are Behind
	if old.Timeline != new.Timeline {
		// Special case: Ahead(n) -> Ahead(m) doesn't change menu
		if old.Timeline == Ahead && new.Timeline == Ahead {
			return false
		}
		// Special case: Behind(n) -> Behind(m) doesn't change menu
		if old.Timeline == Behind && new.Timeline == Behind {
			return false
		}
		return true
	}

	// Branch changes always affect menu (different branch = different state)
	if old.CurrentBranch != new.CurrentBranch {
		return true
	}

	return false
}
