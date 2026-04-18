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
		return false
	}
	cmd := exec.Command("git", "config", "--get", "branch."+currentBranch+".remote")
	return cmd.Run() == nil
}
