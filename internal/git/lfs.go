package git

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// IsRepoLFS checks if the repository uses Git LFS by scanning .gitattributes for filter=lfs entries.
// Uses file read only — no subprocess. Returns false if file does not exist or cannot be read.
func IsRepoLFS() bool {
	data, err := os.ReadFile(".gitattributes")
	hasLFS := err == nil && strings.Contains(string(data), "filter=lfs")
	return hasLFS
}

// IsLFSInstalled checks if the git-lfs binary is in PATH and filters are registered in git config.
func IsLFSInstalled() bool {
	_, pathErr := exec.LookPath("git-lfs")
	binaryExists := pathErr == nil

	result := Execute("config", "--get", "filter.lfs.process")
	filtersRegistered := result.Success && result.Stdout != ""

	return binaryExists && filtersRegistered
}

// IsLFSBinaryAvailable checks if the git-lfs binary is in PATH (regardless of filter registration).
func IsLFSBinaryAvailable() bool {
	_, err := exec.LookPath("git-lfs")
	return err == nil
}

// SetupLFSFilters registers LFS smudge/clean filters by running "git lfs install".
func SetupLFSFilters() CommandResult {
	return Execute("lfs", "install")
}

// FetchLFSObjects downloads LFS objects from remote. Streams output to UI buffer.
func FetchLFSObjects(ctx context.Context) CommandResult {
	return ExecuteWithStreaming(ctx, "lfs", "fetch")
}

// CheckoutLFSObjects materializes LFS pointer files into real content. Streams output to UI buffer.
func CheckoutLFSObjects(ctx context.Context) CommandResult {
	return ExecuteWithStreaming(ctx, "lfs", "checkout")
}
