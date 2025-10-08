package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	cacheDir = ".cache/catalyst-forge-ai"
	repoURL  = "https://github.com/input-output-hk/catalyst-forge-ai.git"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Validate dependencies before running any command
	if err := validateDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "new":
		projectPath := "."
		if len(os.Args) > 2 {
			projectPath = os.Args[2]
		}
		if err := runNew(projectPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "start":
		projectPath := "."
		if len(os.Args) > 2 {
			projectPath = os.Args[2]
		}
		if err := runStart(projectPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`catalyst-ai - Multi-agent system for building Catalyst Forge platform components

Usage:
  catalyst-ai init              Clone and cache catalyst-forge-ai repository
  catalyst-ai new [path]        Initialize AI workspace at path (default: current directory)
  catalyst-ai start [path]      Launch orchestrator for project at path (default: current directory)
  catalyst-ai help              Show this help message

Examples:
  catalyst-ai new projects/my-lib        Create .ai workspace at projects/my-lib/.ai
  catalyst-ai start projects/my-lib      Start orchestrator for projects/my-lib

Requirements:
  - git (for repository operations)
  - claude (for orchestrator and review agents)
  - cursor-agent (for code implementation)`)
}

// validateDependencies checks that required CLI tools are available
func validateDependencies() error {
	required := []string{"git", "claude", "cursor-agent"}

	for _, cmd := range required {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command '%s' not found in PATH", cmd)
		}
	}

	return nil
}

// getCacheDir returns the absolute path to the cache directory
func getCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, cacheDir), nil
}

// getGitRepoRoot finds the git repository root starting from the given path
func getGitRepoRoot(startPath string) (string, error) {
	cmd := exec.Command("git", "-C", startPath, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository or git command failed")
	}
	return strings.TrimSpace(string(output)), nil
}

func runInit() error {
	fmt.Println("Checking for catalyst-forge-ai repository...")

	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	// Check if cache directory already exists
	if _, err := os.Stat(cacheDir); err == nil {
		fmt.Println("Cache exists, updating from upstream...")

		// Pull latest from upstream
		cmd := exec.Command("git", "-C", cacheDir, "pull", "origin", "feat/prototype")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to pull latest changes: %w", err)
		}

		// Delete .ai folder in cache to ensure clean state
		aiDir := filepath.Join(cacheDir, ".ai")
		if _, err := os.Stat(aiDir); err == nil {
			if err := os.RemoveAll(aiDir); err != nil {
				return fmt.Errorf("failed to remove .ai directory: %w", err)
			}
		}

		// Get current commit
		cmd = exec.Command("git", "-C", cacheDir, "rev-parse", "--short", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get commit SHA: %w", err)
		}

		commit := strings.TrimSpace(string(output))

		fmt.Printf("✓ Updated cache at %s\n", cacheDir)
		fmt.Printf("✓ Using branch: feat/prototype\n")
		fmt.Printf("✓ Using commit: %s\n", commit)
		fmt.Println("\nYou can now use 'catalyst-ai new' to start a project.")
		return nil
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(cacheDir), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Clone repository
	fmt.Printf("Cloning repository to %s...\n", cacheDir)
	cmd := exec.Command("git", "clone", "--branch", "feat/prototype", repoURL, cacheDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Set up local tracking branch
	cmd = exec.Command("git", "-C", cacheDir, "branch", "--set-upstream-to=origin/feat/prototype", "feat/prototype")
	if err := cmd.Run(); err != nil {
		// Non-fatal, just warn
		fmt.Fprintf(os.Stderr, "Warning: failed to set upstream tracking: %v\n", err)
	}

	// Get commit SHA
	cmd = exec.Command("git", "-C", cacheDir, "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit SHA: %w", err)
	}

	commit := strings.TrimSpace(string(output))
	fmt.Printf("✓ Cloned to %s\n", cacheDir)
	fmt.Printf("✓ Using branch: feat/prototype\n")
	fmt.Printf("✓ Using commit: %s\n", commit)
	fmt.Println("\nYou can now use 'catalyst-ai new' to start a project.")

	return nil
}

func runNew(projectPath string) error {
	// Convert to absolute path
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Create project directory if it doesn't exist
	if err := os.MkdirAll(absProjectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	aiDir := filepath.Join(absProjectPath, ".ai")

	// Check if .ai/ already exists
	if _, err := os.Stat(aiDir); err == nil {
		return fmt.Errorf(".ai/ directory already exists at %s", absProjectPath)
	}

	// Get cache directory
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	// Verify cache exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return fmt.Errorf("catalyst-forge-ai repository not cached. Run 'catalyst-ai init' first")
	}

	// Get commit SHA from cache
	cmd := exec.Command("git", "-C", cacheDir, "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit SHA from cache: %w", err)
	}
	commit := strings.TrimSpace(string(output))

	// Copy templates directory
	templateSrc := filepath.Join(cacheDir, "templates", ".ai")
	if err := copyDir(templateSrc, aiDir); err != nil {
		return fmt.Errorf("failed to copy templates: %w", err)
	}

	// Process state.yml.tmpl to create state.yml
	stateTemplateFile := filepath.Join(aiDir, "state.yml.tmpl")
	stateFile := filepath.Join(aiDir, "state.yml")

	templateContent, err := os.ReadFile(stateTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read state.yml template: %w", err)
	}

	// Simple template replacement
	stateContent := string(templateContent)
	stateContent = strings.ReplaceAll(stateContent, "{{.CreatedAt}}", time.Now().UTC().Format(time.RFC3339))
	stateContent = strings.ReplaceAll(stateContent, "{{.CacheCommit}}", commit)

	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		return fmt.Errorf("failed to write state.yml: %w", err)
	}

	// Remove template file
	if err := os.Remove(stateTemplateFile); err != nil {
		return fmt.Errorf("failed to remove template file: %w", err)
	}

	fmt.Printf("✓ Created .ai/ workspace at %s\n", absProjectPath)
	fmt.Printf("✓ Copied templates from %s\n", cacheDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("1. (Optional) Add context files to %s/.ai/context/\n", absProjectPath)
	fmt.Println("2. Start the orchestrator:")
	fmt.Println()
	if projectPath == "." {
		fmt.Println("   catalyst-ai start")
	} else {
		fmt.Printf("   catalyst-ai start %s\n", projectPath)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func runStart(projectPath string) error {
	// Convert to absolute path
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Check if project directory exists
	if _, err := os.Stat(absProjectPath); os.IsNotExist(err) {
		return fmt.Errorf("project directory does not exist: %s", absProjectPath)
	}

	aiDir := filepath.Join(absProjectPath, ".ai")
	stateFile := filepath.Join(aiDir, "state.yml")
	orchestratorGuide := filepath.Join(aiDir, "guides", "ORCHESTRATOR.md")

	// Check if .ai/ exists
	if _, err := os.Stat(aiDir); os.IsNotExist(err) {
		return fmt.Errorf(".ai/ directory not found at %s. Run 'catalyst-ai new %s' first", absProjectPath, projectPath)
	}

	// Check if state.yml exists
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return fmt.Errorf(".ai/state.yml not found. Your .ai/ directory may be corrupted")
	}

	// Check if orchestrator guide exists
	if _, err := os.Stat(orchestratorGuide); os.IsNotExist(err) {
		return fmt.Errorf(".ai/guides/ORCHESTRATOR.md not found. Your .ai/ directory may be corrupted")
	}

	// Read state.yml to determine current phase
	stateContent, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state.yml: %w", err)
	}

	// Extract current phase (simple string search)
	currentPhase := "DISCOVERY" // default
	for _, line := range strings.Split(string(stateContent), "\n") {
		if strings.HasPrefix(line, "current_phase:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				currentPhase = strings.TrimSpace(parts[1])
			}
			break
		}
	}

	// Find git repository root
	repoRoot, err := getGitRepoRoot(absProjectPath)
	if err != nil {
		return fmt.Errorf("failed to find git repository root: %w", err)
	}

	// Calculate relative path from repo root to project
	relProjectPath, err := filepath.Rel(repoRoot, absProjectPath)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	fmt.Println("Launching orchestrator...")
	fmt.Printf("Repository root: %s\n", repoRoot)
	fmt.Printf("Project path: %s (relative to repo root)\n", relProjectPath)
	fmt.Printf("Current phase: %s\n\n", currentPhase)

	// Build the claude command
	// claude --system "$(cat .ai/guides/ORCHESTRATOR.md)" "Resume work based on current state in .ai/state.yml"
	orchestratorContent, err := os.ReadFile(orchestratorGuide)
	if err != nil {
		return fmt.Errorf("failed to read orchestrator guide: %w", err)
	}

	// Build initial prompt with project context
	initialPrompt := fmt.Sprintf(`Resume work based on current state in .ai/state.yml

Project context:
- Repository root: %s
- Project path (relative to repo root): %s
- Project absolute path: %s
- Current phase: %s

The .ai/ workspace is located at: %s/.ai/
`, repoRoot, relProjectPath, absProjectPath, currentPhase, absProjectPath)

	cmd := exec.Command("claude",
		"--append-system-prompt", string(orchestratorContent),
		initialPrompt)

	// Change to project directory so relative paths work correctly
	cmd.Dir = absProjectPath

	// Connect stdin/stdout/stderr so user can interact
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude command failed: %w", err)
	}

	return nil
}
