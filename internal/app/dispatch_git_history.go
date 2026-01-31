package app

import (
	"sort"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// dispatchHistory shows commit history
func (a *Application) dispatchHistory(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode               // Track previous mode (Menu)
	app.workflowState.PreviousMenuIndex = app.selectedIndex // Track previous selection
	app.mode = ModeHistory

	var commits []ui.CommitInfo
	for hash, details := range app.cacheManager.GetAllMetadata() {
		commitTime, _ := parseCommitDate(details.Date)
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    commitTime,
		})
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	app.pickerState.History = &ui.HistoryState{
		Commits:           commits,
		SelectedIdx:       0,
		PaneFocused:       true,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}
	return nil
}

// dispatchFileHistory shows file(s) history
func (a *Application) dispatchFileHistory(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode               // Track previous mode (Menu)
	app.workflowState.PreviousMenuIndex = app.selectedIndex // Track previous selection
	app.mode = ModeFileHistory

	var commits []ui.CommitInfo
	for hash, details := range app.cacheManager.GetAllMetadata() {
		commitTime, _ := parseCommitDate(details.Date)
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    commitTime,
		})
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	var files []ui.FileInfo
	if len(commits) > 0 {
		firstCommitHash := commits[0].Hash
		if gitFileList, exists := app.cacheManager.GetFiles(firstCommitHash); exists {
			for _, gitFile := range gitFileList {
				files = append(files, ui.FileInfo{
					Path:   gitFile.Path,
					Status: gitFile.Status,
				})
			}
		}
	}

	app.pickerState.FileHistory = &ui.FileHistoryState{
		Commits:           commits,
		Files:             files,
		SelectedCommitIdx: 0,
		SelectedFileIdx:   0,
		FocusedPane:       ui.PaneCommits,
		CommitsScrollOff:  0,
		FilesScrollOff:    0,
		DiffScrollOff:     0,
		DiffLineCursor:    0,
		DiffContent:       "",
		VisualModeActive:  false,
		VisualModeStart:   0,
	}
	a.updateFileHistoryDiff()
	return nil
}

// dispatchTimeTravelHistory handles the "Browse History" action during time travel
func (a *Application) dispatchTimeTravelHistory(app *Application) tea.Cmd {
	app.mode = ModeHistory

	var commits []ui.CommitInfo
	for hash, details := range app.cacheManager.GetAllMetadata() {
		commitTime, _ := parseCommitDate(details.Date)
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    commitTime,
		})
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})

	app.pickerState.History = &ui.HistoryState{
		Commits:           commits,
		SelectedIdx:       0,
		PaneFocused:       true,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}
	return nil
}
