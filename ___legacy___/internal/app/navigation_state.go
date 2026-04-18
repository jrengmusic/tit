package app

import "time"

// NavigationState manages menu navigation and mode state
type NavigationState struct {
	mode              AppMode
	selectedIndex     int
	menuItems         []MenuItem
	keyHandlers       map[AppMode]map[string]KeyHandler
	quitConfirmActive bool
	quitConfirmTime   time.Time
}

// SelectNext moves selection down, returns true if changed
func (n *NavigationState) SelectNext() bool {
	if len(n.menuItems) == 0 {
		return false
	}
	if n.selectedIndex < len(n.menuItems)-1 {
		n.selectedIndex++
		return true
	}
	return false
}

// SelectPrevious moves selection up, returns true if changed
func (n *NavigationState) SelectPrevious() bool {
	if n.selectedIndex > 0 {
		n.selectedIndex--
		return true
	}
	return false
}

// SelectAt sets selection to specific index with bounds checking
func (n *NavigationState) SelectAt(index int) {
	if index >= 0 && index < len(n.menuItems) {
		n.selectedIndex = index
	}
}

// SelectedItem returns currently selected menu item
func (n *NavigationState) SelectedItem() (MenuItem, bool) {
	if n.selectedIndex >= 0 && n.selectedIndex < len(n.menuItems) {
		return n.menuItems[n.selectedIndex], true
	}
	return MenuItem{}, false
}

// ReplaceMenu updates menu items and resets selection
func (n *NavigationState) ReplaceMenu(items []MenuItem) {
	n.menuItems = items
	n.selectedIndex = 0
}

// RegenerateMenu updates menu items, resets selection, and returns the hint for first item
// Returns the hint string and true if menu has items, empty string and false otherwise
// Caller is responsible for calling rebuildMenuShortcuts with the appropriate mode after this
func (n *NavigationState) RegenerateMenu(items []MenuItem) (string, bool) {
	n.menuItems = items
	n.selectedIndex = 0
	if len(items) > 0 {
		return items[0].Hint, true
	}
	return "", false
}

// ResolveKeyHandler returns handler for key in current mode via map lookup
func (n *NavigationState) ResolveKeyHandler(key string) (KeyHandler, bool) {
	if handlers, ok := n.keyHandlers[n.mode]; ok {
		if handler, ok := handlers[key]; ok {
			return handler, true
		}
	}
	return nil, false
}

// ActivateQuitConfirm enables quit confirmation with timestamp
func (n *NavigationState) ActivateQuitConfirm() {
	n.quitConfirmActive = true
	n.quitConfirmTime = time.Now()
}

// DeactivateQuitConfirm disables quit confirmation
func (n *NavigationState) DeactivateQuitConfirm() {
	n.quitConfirmActive = false
}

// IsQuitConfirmActive returns true if quit confirmation is pending
func (n *NavigationState) IsQuitConfirmActive() bool {
	return n.quitConfirmActive
}

