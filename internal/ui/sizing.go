package ui

// Threshold constants (SSOT) for reactive layout
const (
	MinWidth            = 69
	MinHeight           = 19
	HeaderHeight        = 9
	FooterHeight        = 1
	MinContentHeight    = 4
	HorizontalMargin    = 2
	BannerWidth         = 30
	CommitListPaneWidth = 24 // "07-Jan 02:11 957f977" = 20 chars + border + padding
)

// DynamicSizing holds computed layout dimensions
type DynamicSizing struct {
	TerminalWidth     int
	TerminalHeight    int
	ContentHeight     int
	ContentInnerWidth int
	HeaderInnerWidth  int
	FooterInnerWidth  int
	MenuColumnWidth   int // Left column width for menu (when banner shown)
	IsTooSmall        bool
}

// CalculateDynamicSizing computes all dimensions from terminal size
func CalculateDynamicSizing(termWidth, termHeight int) DynamicSizing {
	isTooSmall := termWidth < MinWidth || termHeight < MinHeight

	contentHeight := termHeight - HeaderHeight - FooterHeight
	if contentHeight < MinContentHeight {
		contentHeight = MinContentHeight
	}

	headerInnerWidth := termWidth - (HorizontalMargin * 2)
	contentInnerWidth := termWidth - (HorizontalMargin * 2)
	footerInnerWidth := termWidth - (HorizontalMargin * 2)

	// Menu column width = content width minus banner and gap
	menuColumnWidth := contentInnerWidth - BannerWidth - 2

	return DynamicSizing{
		TerminalWidth:     termWidth,
		TerminalHeight:    termHeight,
		ContentHeight:     contentHeight,
		ContentInnerWidth: contentInnerWidth,
		HeaderInnerWidth:  headerInnerWidth,
		FooterInnerWidth:  footerInnerWidth,
		MenuColumnWidth:   menuColumnWidth,
		IsTooSmall:        isTooSmall,
	}
}

// CheckIsTooSmall returns true if terminal is too small to render
func (s DynamicSizing) CheckIsTooSmall() bool {
	return s.TerminalWidth < MinWidth || s.TerminalHeight < MinHeight
}

// Legacy constants for backward compatibility (deprecated, use DynamicSizing)
const (
	ContentInnerWidth = 76
	ContentHeight     = 24
)
