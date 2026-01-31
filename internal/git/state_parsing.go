package git

import (
	"os/exec"
	"strings"
)

// executeGitCommand runs git command and returns trimmed output or error
func executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// CurrentBranchExistsOnRemote checks if current branch exists on remote
func CurrentBranchExistsOnRemote() bool {
	currentBranch, err := executeGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false // Can't determine branch, assume doesn't exist on remote
	}
	remoteBranch := "refs/remotes/origin/" + currentBranch
	cmd := exec.Command("git", "rev-parse", remoteBranch)
	return cmd.Run() == nil
}
