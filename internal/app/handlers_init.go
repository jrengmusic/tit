package app

import (
	"fmt"
	"os"

	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleInitCloneCheckout handles OpInit, OpClone, OpCheckout operations
// These operations initialize or switch repositories and need state reload
func (a *Application) handleInitCloneCheckout(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Change to the path if specified
	if msg.Path != "" {
		if err := os.Chdir(msg.Path); err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_cd_into"], msg.Path, err), ui.TypeStderr)
			a.endAsyncOp()
			return a, nil
		}
	}

	// Detect new state after init/clone/checkout
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		return a, nil
	}

	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.mode = ModeConsole

	return a, nil
}
