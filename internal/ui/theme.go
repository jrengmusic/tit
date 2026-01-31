package ui

// Theme defines all semantic colors from the active theme
type Theme struct {
	// Backgrounds
	MainBackgroundColor      string
	InlineBackgroundColor    string
	SelectionBackgroundColor string

	// Text - Content & Body
	ContentTextColor   string
	LabelTextColor     string
	DimmedTextColor    string
	AccentTextColor    string
	HighlightTextColor string
	TerminalTextColor  string

	// Special Text
	CwdTextColor    string
	FooterTextColor string

	// Borders
	BoxBorderColor string
	SeparatorColor string

	// Confirmation Dialog
	ConfirmationDialogBackground string

	// Conflict Resolver - Borders
	ConflictPaneUnfocusedBorder string
	ConflictPaneFocusedBorder   string

	// Conflict Resolver - Selection
	ConflictSelectionForeground string
	ConflictSelectionBackground string

	// Conflict Resolver - Pane Headers
	ConflictPaneTitleColor string

	// Status Colors
	StatusClean string
	StatusDirty string

	// Timeline Colors
	TimelineSynchronized string
	TimelineLocalAhead   string
	TimelineLocalBehind  string

	// Operation Colors
	OperationReady      string
	OperationNotRepo    string
	OperationTimeTravel string
	OperationConflicted string
	OperationMerging    string
	OperationRebasing   string
	OperationDirtyOp    string

	// UI Elements / Buttons
	MenuSelectionBackground string
	ButtonSelectedTextColor string

	// Animation
	SpinnerColor string

	// Diff Colors
	DiffAddedLineColor   string
	DiffRemovedLineColor string

	// Console Output Colors
	OutputStdoutColor  string
	OutputStderrColor  string
	OutputStatusColor  string
	OutputWarningColor string
	OutputDebugColor   string
	OutputInfoColor    string
}
