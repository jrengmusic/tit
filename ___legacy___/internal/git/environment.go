package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectGitEnvironment checks if the machine is ready for git+SSH operations
// This is checked BEFORE any git state detection
// Priority: MissingGit > MissingSSH > NeedsSetup > Ready
//
// For testing: Set TIT_TEST_SETUP=1 to force NeedsSetup state
func DetectGitEnvironment() GitEnvironment {
	// Testing override: force wizard mode
	if os.Getenv("TIT_TEST_SETUP") == "1" {
		return NeedsSetup
	}

	if !commandExists("git") {
		return MissingGit
	}

	if !commandExists("ssh") {
		return MissingSSH
	}

	if !sshKeyExists() {
		return NeedsSetup
	}

	return Ready
}

// commandExists checks if a command is available in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// sshKeyExists checks if any SSH private key exists in ~/.ssh
// Checks both default names (id_rsa, id_ed25519) and custom names (*_rsa, *_ed25519)
func sshKeyExists() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sshDir := filepath.Join(home, ".ssh")

	// Check if .ssh directory exists
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		return false
	}

	// Read directory and look for any private key file
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip public keys and known_hosts
		if filepath.Ext(name) == ".pub" || name == "known_hosts" || name == "known_hosts.old" || name == "config" {
			continue
		}
		// Check for common private key patterns
		// id_rsa, id_ed25519, id_ecdsa, github_rsa, bitbucket_rsa, etc.
		if strings.Contains(name, "_rsa") || strings.Contains(name, "_ed25519") || strings.Contains(name, "_ecdsa") || strings.Contains(name, "_dsa") {
			return true
		}
	}

	return false
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// GetGitVersion returns git version string or empty if not installed
func GetGitVersion() string {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

// GetSSHVersion returns ssh version string or empty if not installed
func GetSSHVersion() string {
	cmd := exec.Command("ssh", "-V")
	// ssh -V outputs to stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return string(output)
}
