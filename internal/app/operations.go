package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// ========================================
// Async Operation Commands
// ========================================
// All cmd* functions return tea.Cmd that execute async git operations
// They capture state in closures to avoid race conditions
// Each returns a GitOperationMsg when complete

// cmdInit executes `git init`, creates initial branch, and commits .gitignore
func (a *Application) cmdInit(branchName string) tea.Cmd {
	name := branchName // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git init
		result := git.ExecuteWithStreaming("init")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to initialize repository",
			}
		}

		// Create initial branch
		result = git.ExecuteWithStreaming("checkout", "-b", name)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to create branch",
			}
		}

		// Create .gitignore and commit it so tree is clean
		if err := git.CreateDefaultGitignore(); err != nil {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   fmt.Sprintf("Failed to create .gitignore: %v", err),
			}
		}

		result = git.ExecuteWithStreaming("add", ".gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to add .gitignore",
			}
		}

		result = git.ExecuteWithStreaming("commit", "-m", "Initialize repository with .gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to commit .gitignore",
			}
		}

		return GitOperationMsg{
			Step:    OpInit,
			Success: true,
			Output:  fmt.Sprintf("Repository initialized with branch '%s'", name),
		}
	}
}

// cmdClone executes `git clone` with streaming output
func (a *Application) cmdClone(url, targetPath string) tea.Cmd {
	u := url // Capture in closure
	path := targetPath
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git clone with streaming output
		result := git.ExecuteWithStreaming("clone", u, path)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpClone,
				Success: false,
				Error:   "Failed to clone repository",
			}
		}

		// Change to cloned directory if needed
		if path != "." {
			if err := os.Chdir(path); err != nil {
				return GitOperationMsg{
					Step:    OpClone,
					Success: false,
					Error:   fmt.Sprintf("Failed to change directory: %v", err),
				}
			}
		}

		// Verify .git exists
		cwd, _ := os.Getwd()
		gitDir := filepath.Join(cwd, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			return GitOperationMsg{
				Step:    OpClone,
				Success: false,
				Error:   fmt.Sprintf("Clone completed but .git not found"),
			}
		}

		return GitOperationMsg{
			Step:    OpClone,
			Success: true,
			Output:  "Repository cloned successfully",
		}
	}
}

// cmdAddRemote adds a remote repository (step 1 of 3-step chain)
func (a *Application) cmdAddRemote(url string) tea.Cmd {
	u := url // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Get current branch BEFORE adding remote (works even with zero commits)
		// Use symbolic-ref instead of rev-parse: works when HEAD exists but has no commits
		branchResult := git.Execute("symbolic-ref", "--short", "HEAD")
		branchName := ""
		if branchResult.Success {
			branchName = strings.TrimSpace(branchResult.Stdout)
		}

		// Add remote
		result := git.ExecuteWithStreaming("remote", "add", "origin", u)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpAddRemote,
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		return GitOperationMsg{
			Step:       OpAddRemote,
			Success:    true,
			Output:     "Remote added",
			BranchName: branchName,
		}
	}
}

// cmdFetchRemote fetches from origin (step 2 of 3-step chain)
func (a *Application) cmdFetchRemote() tea.Cmd {
	return func() tea.Msg {
		result := git.ExecuteWithStreaming("fetch", "--all")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpFetchRemote,
				Success: false,
				Error:   "Failed to fetch from remote",
			}
		}

		return GitOperationMsg{
			Step:    OpFetchRemote,
			Success: true,
			Output:  "Fetch completed",
		}
	}
}

// cmdSetUpstream sets upstream tracking (step 3 of 3-step chain)
// Takes branchName from previous step to avoid querying git again
// If remote branch doesn't exist, this will push -u to create it
func (a *Application) cmdSetUpstream(branchName string) tea.Cmd {
	branch := branchName // Capture in closure
	return func() tea.Msg {
		result := git.SetUpstreamTrackingWithBranch(branch)
		if !result.Success {
			// FAIL-FAST: upstream setup failed
			return GitOperationMsg{
				Step:    OpSetUpstream,
				Success: false,
				Error:   result.Stderr,
			}
		}

		// Distinguish between "set upstream" vs "pushed to create"
		output := "Remote added and upstream configured"
		if result.Stdout == "pushed_and_upstream_set" {
			output = "Remote added and initial push completed"
		}

		return GitOperationMsg{
			Step:    OpSetUpstream,
			Success: true,
			Output:  output,
		}
	}
}

// cmdCommit stages all changes and creates a commit
func (a *Application) cmdCommit(message string) tea.Cmd {
	msg := message // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Stage all changes
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommit,
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommit,
				Success: false,
				Error:   "Failed to commit",
			}
		}

		return GitOperationMsg{
			Step:    OpCommit,
			Success: true,
			Output:  "Changes committed successfully",
		}
	}
}

// cmdCommitPush stages, commits, and pushes in one operation
func (a *Application) cmdCommitPush(message string) tea.Cmd {
	msg := message // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Stage all changes
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommitPush,
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommitPush,
				Success: false,
				Error:   "Failed to commit",
			}
		}

		// Push
		result = git.ExecuteWithStreaming("push")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommitPush,
				Success: false,
				Error:   "Failed to push",
			}
		}

		return GitOperationMsg{
			Step:    OpCommitPush,
			Success: true,
			Output:  "Committed and pushed successfully",
		}
	}
}

// cmdPush pushes current branch to remote
func (a *Application) cmdPush() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Push to upstream
		result := git.ExecuteWithStreaming("push")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpPush,
				Success: false,
				Error:   "Failed to push",
			}
		}

		return GitOperationMsg{
			Step:    OpPush,
			Success: true,
			Output:  "Pushed successfully",
		}
	}
}

// cmdPull pulls from remote (merge)
func (a *Application) cmdPull() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Pull with explicit --no-rebase to merge (required for diverged branches)
		result := git.ExecuteWithStreaming("pull", "--no-rebase")

		if !result.Success {
			// Check if we're in a conflicted state (more reliable than parsing stderr)
			// This detects merge conflicts by checking git state (.git/MERGE_HEAD + unmerged files)
			state, err := git.DetectState()
			if err == nil && state.Operation == git.Conflicted {
				return GitOperationMsg{
					Step:             OpPull,
					Success:          false,
					ConflictDetected: true,
					Error:            ErrorMessages["pull_conflicts"],
				}
			}
			return GitOperationMsg{
				Step:    OpPull,
				Success: false,
				Error:   ErrorMessages["pull_failed"],
			}
		}

		return GitOperationMsg{
			Step:    OpPull,
			Success: true,
			Output:  "Pulled successfully",
		}
	}
}

// cmdInitSubdirectory transitions directly to subdirectory input mode
// Used when CWD is not empty (can't init here, must use subdir)
func (a *Application) cmdInitSubdirectory() tea.Cmd {
	return func() tea.Msg {
		a.mode = ModeInput
		a.inputPrompt = InputMessages["init_subdir_name"].Prompt
		a.inputAction = "init_subdir_name"
		a.footerHint = InputMessages["init_subdir_name"].Hint
		a.inputValue = ""
		a.inputCursorPosition = 0
		return GitOperationMsg{Step: "input_mode_set", Success: true}
	}
}

// cmdForcePush executes git push --force-with-lease
func (a *Application) cmdForcePush() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Get current branch name
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			buffer.Append(ErrorMessages["failed_determine_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpForcePush,
				Success: false,
				Error:   "Could not determine current branch",
			}
		}

		branchName := strings.TrimSpace(string(output))
		buffer.Append(OutputMessages["force_push_in_progress"], ui.TypeInfo)

		// Fetch first to ensure --force-with-lease has current info
		buffer.Append("Fetching latest remote state...", ui.TypeInfo)
		cmd = exec.Command("git", "fetch", "origin")
		if err := cmd.Run(); err != nil {
			// Log fetch failure but continue - force push might still work
			buffer.Append("Warning: fetch failed, force push may fail with 'stale info'", ui.TypeWarning)
		}

		cmd = exec.Command("git", "push", "--force-with-lease", "origin", branchName)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		cmd.Start()

		// Stream output
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			buffer.Append(line, ui.TypeStdout)
		}

		err = cmd.Wait()
		if err != nil {
			return GitOperationMsg{
				Step:    OpForcePush,
				Success: false,
				Error:   "Force push failed",
			}
		}

		return GitOperationMsg{
			Step:    OpForcePush,
			Success: true,
		}
	}
}

// cmdHardReset executes git fetch + reset --hard origin/<branch> (ALWAYS get remote state)
func (a *Application) cmdHardReset() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Get current branch name
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			buffer.Append(ErrorMessages["failed_determine_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   "Could not determine current branch",
			}
		}

		branchName := strings.TrimSpace(string(output))
		buffer.Append(OutputMessages["fetching_latest"], ui.TypeInfo)

		// First: fetch latest from remote
		cmd = exec.Command("git", "fetch", "origin")
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		cmd.Start()

		// Stream fetch output
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			buffer.Append(line, ui.TypeStdout)
		}

		err = cmd.Wait()
		if err != nil {
			buffer.Append(ErrorMessages["failed_fetch_remote"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   "Fetch failed",
			}
		}

		buffer.Append(fmt.Sprintf("Resetting to origin/%s (discarding local state)...", branchName), ui.TypeInfo)

		// Second: reset to origin/<branch> (ALWAYS - regardless of timeline/worktree state)
		cmd = exec.Command("git", "reset", "--hard", fmt.Sprintf("origin/%s", branchName))
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["operation_failed"]),
			}
		}
		stderr, err = cmd.StderrPipe()
		if err != nil {
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["operation_failed"]),
			}
		}

		cmd.Start()

		// Stream reset output
		scanner = bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			buffer.Append(line, ui.TypeStdout)
		}

		err = cmd.Wait()
		if err != nil {
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   "Reset to remote failed",
			}
		}

		buffer.Append(OutputMessages["removing_untracked"], ui.TypeInfo)

		// Third: clean untracked files to make LOCAL == REMOTE exactly
		cmd = exec.Command("git", "clean", "-fd")
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["operation_failed"]),
			}
		}
		stderr, err = cmd.StderrPipe()
		if err != nil {
			return GitOperationMsg{
				Step:    OpHardReset,
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["operation_failed"]),
			}
		}

		cmd.Start()

		// Stream clean output
		scanner = bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			buffer.Append(line, ui.TypeStdout)
		}

		err = cmd.Wait()
		if err != nil {
			buffer.Append(OutputMessages["failed_clean_untracked"], ui.TypeWarning)
		}

		return GitOperationMsg{
			Step:    OpHardReset,
			Success: true,
		}
	}
}

// ========================================
// Dirty Pull Operations
// ========================================

// cmdDirtyPullSnapshot creates a git stash and saves the snapshot state
// Phase 1: Capture original branch/HEAD, then git stash push -u
func (a *Application) cmdDirtyPullSnapshot(preserveChanges bool) tea.Cmd {
	preserve := preserveChanges // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		if preserve {
			buffer.Append(OutputMessages["saving_changes_stash"], ui.TypeInfo)
		} else {
			buffer.Append(OutputMessages["discarding_changes"], ui.TypeInfo)
		}

		// Get current branch name
		branchResult := git.Execute("symbolic-ref", "--short", "HEAD")
		if !branchResult.Success {
			return GitOperationMsg{
				Step:    "dirty_pull_snapshot",
				Success: false,
				Error:   "Failed to get current branch",
			}
		}
		currentBranch := strings.TrimSpace(branchResult.Stdout)

		// Get current HEAD commit hash
		headResult := git.Execute("rev-parse", "HEAD")
		if !headResult.Success {
			return GitOperationMsg{
				Step:    "dirty_pull_snapshot",
				Success: false,
				Error:   "Failed to get current HEAD",
			}
		}
		currentHead := strings.TrimSpace(headResult.Stdout)

		// Save snapshot to .git/TIT_DIRTY_OP
		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Save(currentBranch, currentHead); err != nil {
			return GitOperationMsg{
				Step:    "dirty_pull_snapshot",
				Success: false,
				Error:   fmt.Sprintf("Failed to save snapshot: %v", err),
			}
		}

		// Create stash with uncommitted changes
		if preserve {
			result := git.ExecuteWithStreaming("stash", "push", "-u", "-m", "TIT DIRTY-PULL SNAPSHOT")
			if !result.Success {
				snapshot.Delete() // Cleanup on failure
				return GitOperationMsg{
					Step:    "dirty_pull_snapshot",
					Success: false,
					Error:   "Failed to stash changes",
				}
			}
			buffer.Append(OutputMessages["changes_saved_stashed"], ui.TypeInfo)
		} else {
			// Discard changes without stash
			result := git.ExecuteWithStreaming("reset", "--hard")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    "dirty_pull_snapshot",
					Success: false,
					Error:   "Failed to discard changes",
				}
			}
			result = git.ExecuteWithStreaming("clean", "-fd")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    "dirty_pull_snapshot",
					Success: false,
					Error:   "Failed to clean untracked files",
				}
			}
			buffer.Append(OutputMessages["changes_discarded"], ui.TypeInfo)
		}

		return GitOperationMsg{
			Step:    "dirty_pull_snapshot",
			Success: true,
			Output:  "Snapshot created, tree cleaned",
		}
	}
}

// cmdDirtyPullMerge pulls from remote using merge strategy
// Phase 2: After snapshot, pull remote changes
func (a *Application) cmdDirtyPullMerge() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_pull_merge_started"], ui.TypeInfo)

		result := git.ExecuteWithStreaming("pull", "--no-rebase")
		if !result.Success {
			// Check if we're in a conflicted state (more reliable than parsing stderr)
			// This detects merge conflicts by checking git state (.git/MERGE_HEAD + unmerged files)
			state, err := git.DetectState()
			if err == nil && state.Operation == git.Conflicted {
				buffer.Append(OutputMessages["merge_conflicts_detected"], ui.TypeWarning)
				return GitOperationMsg{
					Step:             "dirty_pull_merge",
					Success:          true, // Mark as success to trigger conflict resolver setup
					ConflictDetected: true,
					Error:            "Merge conflicts detected",
				}
			}
			return GitOperationMsg{
				Step:    "dirty_pull_merge",
				Success: false,
				Error:   "Failed to pull",
			}
		}

		buffer.Append(OutputMessages["merge_completed"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_merge",
			Success: true,
			Output:  "Remote changes merged",
		}
	}
}

// cmdDirtyPullApplySnapshot applies the stashed changes back to the tree
// Phase 3: After pull succeeds, reapply saved changes
func (a *Application) cmdDirtyPullApplySnapshot() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["reapplying_changes"], ui.TypeInfo)

		// Check if there's a stash to apply
		stashListResult := git.Execute("stash", "list")
		if !strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			buffer.Append("No stash to apply (changes were discarded)", ui.TypeInfo) // No SSOT entry needed - contextual message
			return GitOperationMsg{
				Step:    "dirty_pull_apply_snapshot",
				Success: true,
				Output:  "No stashed changes to reapply",
			}
		}

		result := git.ExecuteWithStreaming("stash", "apply")
		if !result.Success {
			// Check if we're in a conflicted state (more reliable than parsing stderr)
			// This detects stash apply conflicts by checking git state (unmerged files)
			state, err := git.DetectState()
			if err == nil && state.Operation == git.Conflicted {
				buffer.Append(OutputMessages["stash_apply_conflicts_detected"], ui.TypeWarning)
				return GitOperationMsg{
					Step:             "dirty_pull_apply_snapshot",
					Success:          true, // Mark as success to trigger conflict resolver setup
					ConflictDetected: true,
					Error:            "Stash apply conflicts detected",
				}
			}
			return GitOperationMsg{
				Step:    "dirty_pull_apply_snapshot",
				Success: false,
				Error:   "Failed to reapply stash",
			}
		}

		buffer.Append(OutputMessages["changes_reapplied"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_apply_snapshot",
			Success: true,
			Output:  "Stashed changes reapplied",
		}
	}
}

// cmdDirtyPullFinalize drops the stash and cleans up the snapshot file
// Phase 4: After all operations succeed, finalize
func (a *Application) cmdDirtyPullFinalize() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_pull_finalize_started"], ui.TypeInfo)

		// Drop the stash (if it exists)
		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			result := git.ExecuteWithStreaming("stash", "drop")
			if !result.Success {
				buffer.Append(OutputMessages["stash_drop_failed_warning"], ui.TypeWarning)
				// Continue anyway - snapshot file cleanup is more important
			}
		}

		// Delete the snapshot file
		if err := git.CleanupSnapshot(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to cleanup snapshot file: %v", err), ui.TypeWarning)
			// Non-fatal, but warn user
		}

		buffer.Append(OutputMessages["dirty_pull_completed_successfully"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_finalize",
			Success: true,
			Output:  "Dirty pull finalized",
		}
	}
}

// cmdAbortDirtyPull restores the exact original state by:
// 1. Checking out the original branch
// 2. Resetting to original HEAD commit
// 3. Reapplying the stash (if changes were preserved)
// 4. Dropping the stash
// 5. Deleting the snapshot file
func (a *Application) cmdAbortDirtyPull() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_pull_aborting"], ui.TypeWarning)

		// CRITICAL: Abort merge first if merge is in progress
		// This cleans up .git/MERGE_HEAD and unmerged files
		state, _ := git.DetectState()
		if state != nil && state.Operation == git.Conflicted {
			buffer.Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			result := git.ExecuteWithStreaming("merge", "--abort")
			if !result.Success {
				// Continue anyway - try to restore state even if merge abort fails
				buffer.Append("Warning: merge abort failed, continuing with restore", ui.TypeWarning)
			}
		}

		// Load snapshot
		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Load(); err != nil {
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   fmt.Sprintf("Failed to load snapshot for abort: %v", err),
			}
		}

		// Checkout original branch
		result := git.ExecuteWithStreaming("checkout", snapshot.OriginalBranch)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_checkout_original_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   fmt.Sprintf("Failed to checkout %s", snapshot.OriginalBranch),
			}
		}

		// Reset to original HEAD
		result = git.ExecuteWithStreaming("reset", "--hard", snapshot.OriginalHead)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_reset_to_original_head"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   "Failed to reset to original HEAD",
			}
		}

		// Reapply stash (if it exists)
		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			result = git.ExecuteWithStreaming("stash", "apply")
			if !result.Success {
				buffer.Append(ErrorMessages["stash_reapply_failed_but_restored"], ui.TypeWarning)
				// Continue - main objective (restoring HEAD) succeeded
			}

			// Drop the stash
			git.ExecuteWithStreaming("stash", "drop")
		}

		// Delete the snapshot file
		snapshot.Delete()

		buffer.Append(OutputMessages["original_state_restored"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_abort",
			Success: true,
			Output:  "Abort completed, original state restored",
		}
	}
}

// cmdFinalizeDirtyPullMerge finalizes the merge commit during dirty pull, then continues to stash apply
// Called after user resolves merge conflicts during dirty pull operation
// This is Phase 2b: After conflict resolution, commit merge before reapplying stash
func (a *Application) cmdFinalizeDirtyPullMerge() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files (already done in handleConflictEnter, but be safe)
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_dirty_pull_merge",
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the merge
		result = git.ExecuteWithStreaming("commit", "-m", "Merge resolved conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_dirty_pull_merge",
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append(OutputMessages["merge_finalized"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "finalize_dirty_pull_merge",
			Success: true,
			Output:  "Merge finalized, continuing to stash apply",
		}
	}
}

// cmdFinalizePullMerge finalizes a merge by committing staged changes
// Called after user resolves conflicts in conflict resolver for pull merge
func (a *Application) cmdFinalizePullMerge() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizePullMerge,
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the merge
		result = git.ExecuteWithStreaming("commit", "-m", "Merge resolved conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizePullMerge,
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append(OutputMessages["merge_finalized"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpFinalizePullMerge,
			Success: true,
			Output:  OutputMessages["merge_finalized"],
		}
	}
}

// cmdFinalizeBranchSwitch stages and commits resolved branch switch conflicts
// Called after user resolves conflicts in conflict resolver for branch switching
func (a *Application) cmdFinalizeBranchSwitch() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_branch_switch",
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the resolution (no merge commit needed, just stage the resolved files)
		// The branch switch itself is already done, we're just resolving the conflicts
		result = git.ExecuteWithStreaming("commit", "-m", "Resolved branch switch conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_branch_switch",
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append("Branch switch conflicts resolved and committed", ui.TypeInfo)
		return GitOperationMsg{
			Step:    "finalize_branch_switch",
			Success: true,
			Output:  "Branch switch completed successfully",
		}
	}
}

// cmdAbortMerge aborts an in-progress merge
// Called when user presses ESC in conflict resolver during pull merge
func (a *Application) cmdAbortMerge() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Abort the merge
		result := git.ExecuteWithStreaming("merge", "--abort")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_abort_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpAbortMerge,
				Success: false,
				Error:   ErrorMessages["failed_abort_merge"],
			}
		}

		// Reset working tree to remove conflict markers
		result = git.ExecuteWithStreaming("reset", "--hard")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_reset_after_abort"], ui.TypeWarning)
			// Non-fatal: merge state is cleared, just working tree has stale markers
		}

		buffer.Append(OutputMessages["merge_aborted"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpAbortMerge,
			Success: true,
			Output:  OutputMessages["merge_aborted"],
		}
	}
}

// cmdFinalizeTimeTravelMerge finalizes time travel merge after conflict resolution
// Commits resolved conflicts and clears time travel marker file
func (a *Application) cmdFinalizeTimeTravelMerge() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files (should already be staged, but be safe)
		buffer.Append("Staging resolved files...", ui.TypeStatus)
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			buffer.Append(fmt.Sprintf("Failed to stage resolved files: %s", result.Stderr), ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeTravelMerge,
				Success: false,
				Error:   "Failed to stage resolved files",
			}
		}

		// Check if there are any staged changes
		// (User might have chosen "keep main" for all conflicts, resulting in no changes)
		diffResult := git.Execute("diff", "--cached", "--quiet")
		hasStagedChanges := !diffResult.Success // --quiet returns non-zero if there are differences

		if hasStagedChanges {
			// Commit the merge
			buffer.Append("Committing resolved conflicts...", ui.TypeStatus)
			result = git.ExecuteWithStreaming("commit", "-m", "Merge time travel changes (conflicts resolved)")
			if !result.Success {
				buffer.Append(fmt.Sprintf("Failed to commit merge: %s", result.Stderr), ui.TypeStderr)
				return GitOperationMsg{
					Step:    OpFinalizeTravelMerge,
					Success: false,
					Error:   "Failed to commit merge",
				}
			}
		} else {
			// No changes to commit (user chose to keep main's version for all conflicts)
			buffer.Append("No changes to commit (kept current branch state)", ui.TypeStatus)

			// CRITICAL: Clean up git merge state
			// When no commit is made, MERGE_HEAD file remains and git thinks merge is incomplete
			buffer.Append("Cleaning up merge state...", ui.TypeStatus)
			os.Remove(".git/MERGE_HEAD")
			os.Remove(".git/MERGE_MODE")
			os.Remove(".git/MERGE_MSG")
			os.Remove(".git/AUTO_MERGE")
			buffer.Append("Merge state cleaned (no-op merge)", ui.TypeStatus)
		}

		// Clear time travel marker file
		buffer.Append("Cleaning up time travel marker...", ui.TypeStatus)
		if err := git.ClearTimeTravelInfo(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to clear time travel marker: %v", err), ui.TypeStderr)
		} else {
			buffer.Append("Time travel marker cleared", ui.TypeStatus)
		}

		buffer.Append("Time travel merge completed successfully!", ui.TypeStatus)
		return GitOperationMsg{
			Step:    OpFinalizeTravelMerge,
			Success: true,
			Output:  "Time travel merge finalized",
		}
	}
}

// cmdFinalizeTimeTravelReturn finalizes time travel return after conflict resolution
// Commits resolved stash conflicts and clears time travel marker file
func (a *Application) cmdFinalizeTimeTravelReturn() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files (should already be staged, but be safe)
		buffer.Append("Staging resolved files...", ui.TypeStatus)
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			buffer.Append(fmt.Sprintf("Failed to stage resolved files: %s", result.Stderr), ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeTravelReturn,
				Success: false,
				Error:   "Failed to stage resolved files",
			}
		}

		// Check if there are any staged changes
		// (User might have chosen "keep main" for all conflicts, resulting in no changes)
		diffResult := git.Execute("diff", "--cached", "--quiet")
		hasStagedChanges := !diffResult.Success // --quiet returns non-zero if there are differences

		if hasStagedChanges {
			// Commit the resolved stash
			buffer.Append("Committing resolved work-in-progress...", ui.TypeStatus)
			result = git.ExecuteWithStreaming("commit", "-m", "Restore work-in-progress (conflicts resolved)")
			if !result.Success {
				buffer.Append(fmt.Sprintf("Failed to commit: %s", result.Stderr), ui.TypeStderr)
				return GitOperationMsg{
					Step:    OpFinalizeTravelReturn,
					Success: false,
					Error:   "Failed to commit resolved work",
				}
			}
		} else {
			// No changes to commit (user chose to keep main's version for all conflicts)
			buffer.Append("No changes to commit (kept current branch state)", ui.TypeStatus)

			// CRITICAL: Clean up git merge state
			// When no commit is made, MERGE_HEAD file remains and git thinks merge is incomplete
			buffer.Append("Cleaning up merge state...", ui.TypeStatus)
			os.Remove(".git/MERGE_HEAD")
			os.Remove(".git/MERGE_MODE")
			os.Remove(".git/MERGE_MSG")
			os.Remove(".git/AUTO_MERGE")
			buffer.Append("Merge state cleaned (no-op merge)", ui.TypeStatus)
		}

		// Clear time travel marker file
		buffer.Append("Cleaning up time travel marker...", ui.TypeStatus)
		if err := git.ClearTimeTravelInfo(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to clear time travel marker: %v", err), ui.TypeStderr)
		} else {
			buffer.Append("Time travel marker cleared", ui.TypeStatus)
		}

		buffer.Append("Time travel return completed successfully!", ui.TypeStatus)
		return GitOperationMsg{
			Step:    OpFinalizeTravelReturn,
			Success: true,
			Output:  "Time travel return finalized",
		}
	}
}
