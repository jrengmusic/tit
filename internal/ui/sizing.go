package ui

// SSOT: Single Source of Truth for all dimensions
const (
	InterfaceWidth    = 80
	BannerHeight      = 14
	HeaderHeight      = 6
	ContentHeight     = 24
	FooterHeight      = 2
	ContentInnerWidth = InterfaceWidth - 2 // Content box minus border (76)
)

// TotalHeight is sum of all section heights
const TotalHeight = BannerHeight + HeaderHeight + ContentHeight + FooterHeight

// Sizing holds all layout dimensions
type Sizing struct {
	Width          int // Always InterfaceWidth
	Height         int // Always TotalHeight
	BannerHeight   int
	HeaderHeight   int
	ContentHeight  int
	FooterHeight   int
}

// CalculateSizing returns layout with SSOT values
func CalculateSizing() Sizing {
	return Sizing{
		Width:         InterfaceWidth,
		Height:        TotalHeight,
		BannerHeight:  BannerHeight,
		HeaderHeight:  HeaderHeight,
		ContentHeight: ContentHeight,
		FooterHeight:  FooterHeight,
	}
}
