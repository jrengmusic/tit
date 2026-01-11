package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateSSHKey generates a new SSH key pair
func GenerateSSHKey(email string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	sshDir := filepath.Join(home, ".ssh")
	
	// Create .ssh directory if it doesn't exist
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("could not create .ssh directory: %w", err)
	}

	// Use TIT-specific key names to avoid conflicts
	keyPath := filepath.Join(sshDir, "TIT_id_rsa")
	
	// Check if TIT key already exists
	if _, err := os.Stat(keyPath); err == nil {
		// Key already exists, skip generation
		return nil
	}

	// Generate SSH key: ssh-keygen -t rsa -b 4096 -C "email" -f keyPath -N ""
	// Note: ssh-keygen automatically creates the .pub file with the same name
	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-C", email, "-f", keyPath, "-N", "")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-keygen failed: %w", err)
	}

	return nil
}

// AddKeyToAgent adds the SSH key to ssh-agent
func AddKeyToAgent() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	keyPath := filepath.Join(home, ".ssh", "TIT_id_rsa")
	
	// Start ssh-agent if not running
	if !isSSHAgentRunning() {
		if err := startSSHAgent(); err != nil {
			return fmt.Errorf("could not start ssh-agent: %w", err)
		}
	}

	// Add key to agent: ssh-add keyPath
	cmd := exec.Command("ssh-add", keyPath)
	if err := cmd.Run(); err != nil {
		// ssh-add may fail if agent is not running or key already added
		// This is not a critical error
		return nil
	}

	return nil
}

// WriteSSHConfig writes SSH configuration
func WriteSSHConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	sshDir := filepath.Join(home, ".ssh")
	configPath := filepath.Join(sshDir, "config")
	
	// Check if config already has the required settings
	if configHasRequiredSettings(configPath) {
		return nil
	}

	// Append configuration
	configContent := `Host *
  AddKeysToAgent yes
  IdentityFile ~/.ssh/TIT_id_rsa
  IdentitiesOnly yes
`

	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not open SSH config: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(configContent); err != nil {
		return fmt.Errorf("could not write SSH config: %w", err)
	}

	return nil
}

// GetPublicKey returns the public key content
func GetPublicKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}

	pubKeyPath := filepath.Join(home, ".ssh", "TIT_id_rsa.pub")
	
	data, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return "", fmt.Errorf("could not read public key: %w", err)
	}

	return string(data), nil
}

// isSSHAgentRunning checks if ssh-agent is running
func isSSHAgentRunning() bool {
	cmd := exec.Command("ssh-add", "-l")
	return cmd.Run() == nil
}

// startSSHAgent starts ssh-agent
func startSSHAgent() error {
	cmd := exec.Command("ssh-agent", "-s")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not start ssh-agent: %w\n%s", err, string(output))
	}

	// Parse output to get agent environment variables
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "SSH_AUTH_SOCK=") || strings.HasPrefix(line, "SSH_AGENT_PID=") {
			// Set environment variable
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}

	return nil
}

// configHasRequiredSettings checks if SSH config already has the required settings
func configHasRequiredSettings(configPath string) bool {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}

	configContent := string(data)
	
	// Check if config contains the key settings
	return strings.Contains(configContent, "AddKeysToAgent yes") &&
		   strings.Contains(configContent, "IdentityFile ~/.ssh/TIT_id_rsa")
}