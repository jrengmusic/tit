package app

// MenuItemBuilder provides a fluent API for creating MenuItems.
type MenuItemBuilder struct {
	item MenuItem
}

// Item creates a new MenuItemBuilder with a given ID.
// By default, the item is enabled.
func Item(id string) *MenuItemBuilder {
	return &MenuItemBuilder{
		item: MenuItem{
			ID:      id,
			Enabled: true,
		},
	}
}

// Shortcut sets the keyboard shortcut for the menu item.
func (b *MenuItemBuilder) Shortcut(s string) *MenuItemBuilder {
	b.item.Shortcut = s
	return b
}

// Emoji sets the emoji for the menu item.
func (b *MenuItemBuilder) Emoji(e string) *MenuItemBuilder {
	b.item.Emoji = e
	return b
}

// Label sets the display label for the menu item.
func (b *MenuItemBuilder) Label(l string) *MenuItemBuilder {
	b.item.Label = l
	return b
}

// Hint sets the hint text displayed in the footer.
func (b *MenuItemBuilder) Hint(h string) *MenuItemBuilder {
	b.item.Hint = h
	return b
}

// When sets the enabled state of the menu item based on a condition.
func (b *MenuItemBuilder) When(condition bool) *MenuItemBuilder {
	b.item.Enabled = condition
	return b
}

// Separator marks this item as a separator.
func (b *MenuItemBuilder) Separator() *MenuItemBuilder {
	b.item.Separator = true
	return b
}

// Build returns the final MenuItem.
func (b *MenuItemBuilder) Build() MenuItem {
	return b.item
}
