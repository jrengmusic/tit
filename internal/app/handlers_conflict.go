package app

import (
	"fmt"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// setupConflictResolver initializes conflict resolver UI for any conflict-resolving operation
// Parameters:
//   - operation: operation identifier (e.g., "pull_merge", "dirty_pull_changeset_apply", "cherry_pick")
//   - columnLabels: human-readable labels for the 3 columns (e.g., ["BASE", "LOCAL (yours)", "REMOTE (theirs)"])
func (a *Application) setupConflictResolver(operation string, columnLabels []string) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	buffer.Append(OutputMessages["detecting_conflicts"], ui.TypeInfo)

	// Get list of conflicted files from git status
	conflictFiles, err := git.ListConflictedFiles()
	if err != nil {
		buffer.Append(fmt.Sprintf(OutputMessages["conflict_detection_error"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.footerHint = ErrorMessages["operation_failed"]
		a.mode = ModeConsole
		return a, nil
	}

	buffer.Append(fmt.Sprintf("Found %d conflicted file(s)", len(conflictFiles)), ui.TypeInfo)

	if len(conflictFiles) == 0 {
		// No conflicts found - should not happen, but handle gracefully
		buffer.Append(OutputMessages["conflict_detection_none"], ui.TypeInfo)
		a.endAsyncOp()
		a.mode = ModeConsole
		return a, nil
	}

	// Detect which stages exist by checking the first conflicted file
	// This determines how many columns we'll show (2-way vs 3-way merge)
	// For delete/modify conflicts, we show placeholder text for deleted stages
	var stagesPresent []int
	var stagesDeleted []int
	var activeLabels []string

	if len(conflictFiles) > 0 {
		// Check which stages exist for first file (all files should have same stage structure)
		testFile := conflictFiles[0]

		// Try stage 1 (BASE)
		if _, err := git.ShowConflictVersion(testFile, 1); err == nil {
			stagesPresent = append(stagesPresent, 1)
		} else {
			stagesDeleted = append(stagesDeleted, 1)
		}

		// Try stage 2 (LOCAL)
		if _, err := git.ShowConflictVersion(testFile, 2); err == nil {
			stagesPresent = append(stagesPresent, 2)
		} else {
			stagesDeleted = append(stagesDeleted, 2)
		}

		// Try stage 3 (REMOTE)
		if _, err := git.ShowConflictVersion(testFile, 3); err == nil {
			stagesPresent = append(stagesPresent, 3)
		} else {
			stagesDeleted = append(stagesDeleted, 3)
		}

		// Build active labels for all stages (present and deleted)
		// columnLabels is indexed 0, 1, 2 for BASE, LOCAL, REMOTE
		allStages := []int{1, 2, 3}
		for _, stage := range allStages {
			labelIdx := stage - 1 // Stage 1->label[0], stage 2->label[1], stage 3->label[2]
			if labelIdx < len(columnLabels) {
				activeLabels = append(activeLabels, columnLabels[labelIdx])
			}
		}
	}

	// Always show 3 columns for standard 3-way conflicts
	// But handle cases where stages are deleted (delete/modify conflicts)
	numColumns := 3
	if len(stagesPresent) == 0 {
		// No stages could be read - this indicates a corrupt conflict state
		buffer.Append("Error: No conflict stages found. The conflict state may be corrupted.", ui.TypeStderr)
		a.endAsyncOp()
		a.footerHint = ErrorMessages["operation_failed"]
		a.mode = ModeConsole
		return a, nil
	}

	// Read versions for each conflicted file
	resolveState := &ConflictResolveState{
		Files:             make([]ui.ConflictFileGeneric, 0, len(conflictFiles)),
		SelectedFileIndex: 0,
		FocusedPane:       0,
		NumColumns:        numColumns,
		ColumnLabels:      activeLabels,
		ScrollOffsets:     make([]int, numColumns),
		LineCursors:       make([]int, numColumns),
		Operation:         operation,
	}

	for _, filePath := range conflictFiles {
		var versions []string

		// Read all 3 stages, showing placeholder for deleted stages
		for stage := 1; stage <= 3; stage++ {
			content, err := git.ShowConflictVersion(filePath, stage)
			if err != nil {
				// Stage doesn't exist (file was deleted in this version)
				content = fmt.Sprintf("[FILE DELETED IN THIS VERSION]\n\nThis file was deleted in %s.\nThe conflict occurred because the other side modified it.",
					map[int]string{1: "BASE", 2: "LOCAL", 3: "REMOTE"}[stage])
			}
			versions = append(versions, content)
		}

		// Build conflict file entry
		conflictFile := ui.ConflictFileGeneric{
			Path:     filePath,
			Versions: versions,
			Chosen:   -1, // Not yet marked
		}
		resolveState.Files = append(resolveState.Files, conflictFile)
	}

	// Store conflict state and transition to resolver UI
	a.conflictResolveState = resolveState
	a.endAsyncOp()
	a.mode = ModeConflictResolve
	a.footerHint = fmt.Sprintf(ConsoleMessages["resolve_conflicts_help"], len(conflictFiles))

	buffer.Append(fmt.Sprintf(OutputMessages["conflicts_detected_count"], len(conflictFiles)), ui.TypeInfo)
	buffer.Append(OutputMessages["mark_choices_in_resolver"], ui.TypeInfo)

	return a, nil
}

// setupConflictResolverForPull initializes conflict resolver for pull merge conflicts (convenience wrapper)
func (a *Application) setupConflictResolverForPull(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	return a.setupConflictResolver("pull_merge", []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"})
}

// setupConflictResolverForDirtyPull initializes conflict resolver for dirty pull conflicts (convenience wrapper)
func (a *Application) setupConflictResolverForDirtyPull(msg GitOperationMsg, conflictPhase string) (tea.Model, tea.Cmd) {
	operation := "dirty_pull_" + conflictPhase
	return a.setupConflictResolver(operation, []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"})
}

// setupConflictResolverForBranchSwitch initializes conflict resolver for branch switch conflicts (convenience wrapper)
// Conflicts occur when switching would overwrite local changes
func (a *Application) setupConflictResolverForBranchSwitch(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	targetBranch := msg.BranchName
	currentBranch := ""
	if a.gitState != nil {
		currentBranch = a.gitState.CurrentBranch
	}

	// Column labels: BASE, LOCAL (current branch), REMOTE (target branch)
	labels := []string{
		"BASE",
		fmt.Sprintf("LOCAL (%s)", currentBranch),
		fmt.Sprintf("REMOTE (%s)", targetBranch),
	}

	return a.setupConflictResolver("branch_switch", labels)
}
