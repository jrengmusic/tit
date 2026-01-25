package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// AsyncOperation is a builder for creating asynchronous operations.
type AsyncOperation struct {
	step      string
	steps     []func() error
	onSuccess string
}

// NewAsyncOp creates a new AsyncOperation builder.
func NewAsyncOp(step string) *AsyncOperation {
	return &AsyncOperation{step: step}
}

// AddStep adds a step to the async operation.
// Each step is a function that returns an error.
// If a step returns an error, the operation is halted.
func (op *AsyncOperation) AddStep(fn func() error) *AsyncOperation {
	op.steps = append(op.steps, fn)
	return op
}

// SuccessMessage sets the message to be returned on successful completion.
func (op *AsyncOperation) SuccessMessage(msg string) *AsyncOperation {
	op.onSuccess = msg
	return op
}

// Execute returns a tea.Cmd that will run the async operation.
func (op *AsyncOperation) Execute() tea.Cmd {
	return func() tea.Msg {
		for i, step := range op.steps {
			if err := step(); err != nil {
				return GitOperationMsg{
					Step:    op.step,
					Success: false,
					Error:   fmt.Sprintf("Step %d failed: %v", i+1, err),
				}
			}
		}

		return GitOperationMsg{
			Step:    op.step,
			Success: true,
			Output:  op.onSuccess,
		}
	}
}
