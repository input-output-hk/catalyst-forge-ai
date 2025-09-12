package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindRoot finds the repository root by looking upward for a .git directory
// starting from the current working directory. Returns an error if no .git
// directory is found.
func FindRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return FindRootFromDir(cwd)
}

// FindRootFromDir finds the repository root by looking upward for a .git directory
// starting from the given directory. Returns an error if no .git directory is found.
func FindRootFromDir(startDir string) (string, error) {
	// Clean and resolve the starting directory
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %w", startDir, err)
	}

	// Walk upward from the starting directory
	for {
		gitDir := filepath.Join(dir, ".git")

		// Check if .git exists and is a directory
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)

		// If we've reached the root directory and haven't found .git, stop
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("no git repository found (no .git directory) starting from %s", startDir)
}

// IsForgeProject checks if the given directory is a Forge AI project
// by verifying the presence of the .forge/ai directory.
func IsForgeProject(projectRoot string) bool {
	forgeAIDir := filepath.Join(projectRoot, ".forge", "ai")
	if info, err := os.Stat(forgeAIDir); err == nil && info.IsDir() {
		return true
	}
	return false
}
