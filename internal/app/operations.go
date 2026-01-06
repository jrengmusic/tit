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
				Step:    "init",
				Success: false,
				Error:   "Failed to initialize repository",
			}
		}

		// Create initial branch
		result = git.ExecuteWithStreaming("checkout", "-b", name)
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to create branch",
			}
		}

		// Create .gitignore and commit it so tree is clean
		if err := git.CreateDefaultGitignore(); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to create .gitignore: %v", err),
			}
		}

		result = git.ExecuteWithStreaming("add", ".gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to add .gitignore",
			}
		}

		result = git.ExecuteWithStreaming("commit", "-m", "Initialize repository with .gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to commit .gitignore",
			}
		}

		return GitOperationMsg{
			Step:    "init",
			Success: true,
			Output:  fmt.Sprintf("Repository initialized with branch '%s'", name),
		}
	}
}

// cmdClone executes `git clone` with streaming output
func (a *Application) cmdClone(url, targetPath string) tea.Cmd {
	u := url             // Capture in closure
	path := targetPath
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git clone with streaming output
		result := git.ExecuteWithStreaming("clone", u, path)
		if !result.Success {
			return GitOperationMsg{
				Step:    "clone",
				Success: false,
				Error:   "Failed to clone repository",
			}
		}

		// Change to cloned directory if needed
		if path != "." {
			if err := os.Chdir(path); err != nil {
				return GitOperationMsg{
					Step:    "clone",
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
				Step:    "clone",
				Success: false,
				Error:   fmt.Sprintf("Clone completed but .git not found"),
			}
		}

		return GitOperationMsg{
			Step:    "clone",
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
				Step:    "add_remote",
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		return GitOperationMsg{
			Step:       "add_remote",
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
				Step:    "fetch_remote",
				Success: false,
				Error:   "Failed to fetch from remote",
			}
		}

		return GitOperationMsg{
			Step:    "fetch_remote",
			Success: true,
			Output:  "Fetch completed",
		}
	}
}

// cmdSetUpstream sets upstream tracking (step 3 of 3-step chain)
// Takes branchName from previous step to avoid querying git again
func (a *Application) cmdSetUpstream(branchName string) tea.Cmd {
	branch := branchName // Capture in closure
	return func() tea.Msg {
		result := git.SetUpstreamTrackingWithBranch(branch)
		if !result.Success {
			// Non-fatal: tracking setup failed but remote is ready
			return GitOperationMsg{
				Step:    "set_upstream",
				Success: true, // Consider success because remote is added
				Output:  "Remote configured (upstream tracking could not be set - may be detached HEAD)",
			}
		}

		return GitOperationMsg{
			Step:    "set_upstream",
			Success: true,
			Output:  "Remote added, fetched, and tracking configured",
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
				Step:    "commit",
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    "commit",
				Success: false,
				Error:   "Failed to commit",
			}
		}

		return GitOperationMsg{
			Step:    "commit",
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
				Step:    "commit_push",
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    "commit_push",
				Success: false,
				Error:   "Failed to commit",
			}
		}

		// Push
		result = git.ExecuteWithStreaming("push")
		if !result.Success {
			return GitOperationMsg{
				Step:    "commit_push",
				Success: false,
				Error:   "Failed to push",
			}
		}

		return GitOperationMsg{
			Step:    "commit_push",
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
				Step:    "push",
				Success: false,
				Error:   "Failed to push",
			}
		}

		return GitOperationMsg{
			Step:    "push",
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

		// Pull (merge by default)
		result := git.ExecuteWithStreaming("pull")

		if !result.Success {
			// Check if conflict
			if strings.Contains(result.Stderr, "conflict") || strings.Contains(result.Stderr, "CONFLICT") {
				return GitOperationMsg{
					Step:    "pull",
					Success: false,
					Error:   "Merge conflicts occurred",
				}
			}
			return GitOperationMsg{
				Step:    "pull",
				Success: false,
				Error:   "Failed to pull",
			}
		}

		return GitOperationMsg{
			Step:    "pull",
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
		a.inputPrompt = InputPrompts["init_subdir_name"]
		a.inputAction = "init_subdir_name"
		a.footerHint = InputHints["init_subdir_name"]
		a.inputValue = ""
		a.inputCursorPosition = 0
		return GitOperationMsg{Step: "input_mode_set", Success: true}
	}
}

// cmdForcePush executes git push --force-with-lease (like old-tit)
func (a *Application) cmdForcePush() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()
		
		// Get current branch name
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			buffer.Append("Error: Could not determine current branch", ui.TypeStderr)
			return GitOperationMsg{
				Step:    "force_push",
				Success: false,
				Error:   "Could not determine current branch",
			}
		}
		
		branchName := strings.TrimSpace(string(output))
		buffer.Append("Force pushing to remote (overwriting remote history)...", ui.TypeInfo)
		
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
				Step:    "force_push", 
				Success: false,
				Error:   "Force push failed",
			}
		}
		
		return GitOperationMsg{
			Step:    "force_push",
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
			buffer.Append("Error: Could not determine current branch", ui.TypeStderr)
			return GitOperationMsg{
				Step:    "hard_reset",
				Success: false,
				Error:   "Could not determine current branch",
			}
		}
		
		branchName := strings.TrimSpace(string(output))
		buffer.Append("Fetching latest from remote...", ui.TypeInfo)
		
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
			buffer.Append("Failed to fetch from remote", ui.TypeStderr)
			return GitOperationMsg{
				Step:    "hard_reset",
				Success: false,
				Error:   "Fetch failed",
			}
		}
		
		buffer.Append(fmt.Sprintf("Resetting to origin/%s (discarding local state)...", branchName), ui.TypeInfo)
		
		// Second: reset to origin/<branch> (ALWAYS - regardless of timeline/worktree state)
		cmd = exec.Command("git", "reset", "--hard", fmt.Sprintf("origin/%s", branchName))
		stdout, _ = cmd.StdoutPipe()
		stderr, _ = cmd.StderrPipe()
		
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
				Step:    "hard_reset",
				Success: false,
				Error:   "Reset to remote failed",
			}
		}
		
		buffer.Append("Removing untracked files and directories...", ui.TypeInfo)
		
		// Third: clean untracked files to make LOCAL == REMOTE exactly
		cmd = exec.Command("git", "clean", "-fd")
		stdout, _ = cmd.StdoutPipe()
		stderr, _ = cmd.StderrPipe()
		
		cmd.Start()
		
		// Stream clean output
		scanner = bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			buffer.Append(line, ui.TypeStdout)
		}
		
		err = cmd.Wait()
		if err != nil {
			buffer.Append("Warning: Failed to clean untracked files", ui.TypeWarning)
		}
		
		return GitOperationMsg{
			Step:    "hard_reset",
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
			buffer.Append("Saving your changes (creating stash)...", ui.TypeInfo)
		} else {
			buffer.Append("Discarding your changes...", ui.TypeInfo)
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
			buffer.Append("Changes saved (stashed)", ui.TypeInfo)
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
			buffer.Append("Changes discarded", ui.TypeInfo)
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
		buffer.Append("Pulling from remote (merge strategy)...", ui.TypeInfo)

		result := git.ExecuteWithStreaming("pull")
		if !result.Success {
			// Check for conflict markers
			if strings.Contains(result.Stderr, "CONFLICT") || strings.Contains(result.Stderr, "conflict") {
				buffer.Append("Merge conflicts detected", ui.TypeWarning)
				return GitOperationMsg{
					Step:    "dirty_pull_merge",
					Success: false,
					Error:   "Merge conflicts detected",
				}
			}
			return GitOperationMsg{
				Step:    "dirty_pull_merge",
				Success: false,
				Error:   "Failed to pull",
			}
		}

		buffer.Append("Merge completed", ui.TypeInfo)
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
		buffer.Append("Reapplying your changes...", ui.TypeInfo)

		// Check if there's a stash to apply
		stashListResult := git.Execute("stash", "list")
		if !strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			buffer.Append("No stash to apply (changes were discarded)", ui.TypeInfo)
			return GitOperationMsg{
				Step:    "dirty_pull_apply_snapshot",
				Success: true,
				Output:  "No stashed changes to reapply",
			}
		}

		result := git.ExecuteWithStreaming("stash", "apply")
		if !result.Success {
			// Check for conflict markers
			if strings.Contains(result.Stderr, "CONFLICT") || strings.Contains(result.Stderr, "conflict") {
				buffer.Append("Conflicts detected while reapplying changes", ui.TypeWarning)
				return GitOperationMsg{
					Step:    "dirty_pull_apply_snapshot",
					Success: false,
					Error:   "Stash apply conflicts detected",
				}
			}
			return GitOperationMsg{
				Step:    "dirty_pull_apply_snapshot",
				Success: false,
				Error:   "Failed to reapply stash",
			}
		}

		buffer.Append("Changes reapplied", ui.TypeInfo)
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
		buffer.Append("Finalizing dirty pull operation...", ui.TypeInfo)

		// Drop the stash (if it exists)
		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			result := git.ExecuteWithStreaming("stash", "drop")
			if !result.Success {
				buffer.Append("Warning: Failed to drop stash (manual cleanup may be needed)", ui.TypeWarning)
				// Continue anyway - snapshot file cleanup is more important
			}
		}

		// Delete the snapshot file
		if err := git.CleanupSnapshot(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to cleanup snapshot file: %v", err), ui.TypeWarning)
			// Non-fatal, but warn user
		}

		buffer.Append("Dirty pull completed successfully", ui.TypeInfo)
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
		buffer.Append("Aborting dirty pull and restoring original state...", ui.TypeWarning)

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
			buffer.Append("Error: Failed to checkout original branch", ui.TypeStderr)
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   fmt.Sprintf("Failed to checkout %s", snapshot.OriginalBranch),
			}
		}

		// Reset to original HEAD
		result = git.ExecuteWithStreaming("reset", "--hard", snapshot.OriginalHead)
		if !result.Success {
			buffer.Append("Error: Failed to reset to original HEAD", ui.TypeStderr)
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
				buffer.Append("Warning: Could not reapply stash, but HEAD restored", ui.TypeWarning)
				// Continue - main objective (restoring HEAD) succeeded
			}

			// Drop the stash
			git.ExecuteWithStreaming("stash", "drop")
		}

		// Delete the snapshot file
		snapshot.Delete()

		buffer.Append("Original state restored", ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_abort",
			Success: true,
			Output:  "Abort completed, original state restored",
		}
	}
}


