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

// SetMode updates the current application mode
func (n *NavigationState) SetMode(mode AppMode) {
	n.mode = mode
}

// GetMode returns the current application mode
func (n *NavigationState) GetMode() AppMode {
	return n.mode
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

// GetSelectedIndex returns current selection index
func (n *NavigationState) GetSelectedIndex() int {
	return n.selectedIndex
}

// SetSelectedIndex sets selection to specific index
func (n *NavigationState) SetSelectedIndex(index int) {
	if index >= 0 && index < len(n.menuItems) {
		n.selectedIndex = index
	}
}

// GetSelectedItem returns currently selected menu item
func (n *NavigationState) GetSelectedItem() (MenuItem, bool) {
	if n.selectedIndex >= 0 && n.selectedIndex < len(n.menuItems) {
		return n.menuItems[n.selectedIndex], true
	}
	return MenuItem{}, false
}

// SetMenuItems updates menu items and resets selection
func (n *NavigationState) SetMenuItems(items []MenuItem) {
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

// GetMenuItems returns current menu items
func (n *NavigationState) GetMenuItems() []MenuItem {
	return n.menuItems
}

// GetKeyHandler returns handler for key in current mode
func (n *NavigationState) GetKeyHandler(key string) (KeyHandler, bool) {
	if handlers, ok := n.keyHandlers[n.mode]; ok {
		if handler, ok := handlers[key]; ok {
			return handler, true
		}
	}
	return nil, false
}

// SetKeyHandlers updates key handler map
func (n *NavigationState) SetKeyHandlers(handlers map[AppMode]map[string]KeyHandler) {
	n.keyHandlers = handlers
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

// GetQuitConfirmTime returns when quit was initiated
func (n *NavigationState) GetQuitConfirmTime() time.Time {
	return n.quitConfirmTime
}
