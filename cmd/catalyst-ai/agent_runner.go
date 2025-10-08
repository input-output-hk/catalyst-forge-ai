package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// runAgent executes a specific agent with template variable substitution
func runAgent(agentName string, args map[string]string) error {
	// Default project to current directory if not provided
	projectPath := "."
	if p, ok := args["project"]; ok {
		projectPath = p
	}

	// Convert to absolute path
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	// Check if project .ai directory exists
	aiDir := filepath.Join(absProjectPath, ".ai")
	if _, err := os.Stat(aiDir); os.IsNotExist(err) {
		return fmt.Errorf(".ai/ directory not found at %s. Run 'catalyst-ai new %s' first", absProjectPath, projectPath)
	}

	// Get repository root
	repoRoot, err := getGitRepoRoot(absProjectPath)
	if err != nil {
		return fmt.Errorf("failed to find git repository root: %w", err)
	}

	// Calculate relative path
	relProjectPath, err := filepath.Rel(repoRoot, absProjectPath)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Build template data
	templateData := map[string]string{
		"ProjectPath":    absProjectPath,
		"RelProjectPath": relProjectPath,
		"RepoRoot":       repoRoot,
		"AIDir":          aiDir,
	}

	// Add all args to template data
	for k, v := range args {
		templateData[k] = v
	}

	// Determine agent guide file
	var guideFile string
	var agentCLI string
	var initialPromptTmpl string

	switch agentName {
	case "planner":
		guideFile = filepath.Join(aiDir, "guides", "PLANNER.md")
		agentCLI = "claude"
		initialPromptTmpl = "Create implementation plan based on design document at {{.AIDir}}/design/DESIGN.md"

	case "coder":
		if _, ok := args["task_id"]; !ok {
			return fmt.Errorf("coder agent requires task_id argument (e.g., task_id=001)")
		}
		guideFile = filepath.Join(aiDir, "guides", "CODER.md")
		agentCLI = "cursor-agent"
		initialPromptTmpl = "Implement task {{.task_id}} from {{.AIDir}}/planning/tasks/{{.task_id}}.md"

	case "reviewer":
		if _, ok := args["task_id"]; !ok {
			return fmt.Errorf("reviewer agent requires task_id argument (e.g., task_id=001)")
		}
		guideFile = filepath.Join(aiDir, "guides", "REVIEWER.md")
		agentCLI = "claude"
		initialPromptTmpl = "Review implementation of task {{.task_id}}"

	default:
		return fmt.Errorf("unknown agent: %s (available: planner, coder, reviewer)", agentName)
	}

	// Check if guide file exists
	if _, err := os.Stat(guideFile); os.IsNotExist(err) {
		return fmt.Errorf("agent guide not found: %s", guideFile)
	}

	// Read guide file
	guideContent, err := os.ReadFile(guideFile)
	if err != nil {
		return fmt.Errorf("failed to read agent guide: %w", err)
	}

	// Process initial prompt template
	tmpl, err := template.New("prompt").Parse(initialPromptTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, templateData); err != nil {
		return fmt.Errorf("failed to execute prompt template: %w", err)
	}
	initialPrompt := promptBuf.String()

	// Build agent-specific context
	contextPrompt := fmt.Sprintf(`Project Context:
- Repository root: %s
- Project path (relative to repo root): %s
- Project absolute path: %s
- .ai/ workspace: %s
`, repoRoot, relProjectPath, absProjectPath, aiDir)

	// Add task-specific context if applicable
	if taskID, ok := args["task_id"]; ok {
		contextPrompt += fmt.Sprintf("- Task ID: %s\n", taskID)
	}

	// Build full prompt differently for cursor-agent vs claude
	var fullPrompt string
	if agentCLI == "cursor-agent" {
		// cursor-agent: Combine guide + context + prompt into single prompt
		// (no separate system prompt support)
		fullPrompt = fmt.Sprintf(`%s

%s

%s`, string(guideContent), contextPrompt, initialPrompt)
	} else {
		// claude: Context + prompt (guide goes in system prompt)
		fullPrompt = contextPrompt + "\n" + initialPrompt
	}

	fmt.Printf("Running %s agent...\n", agentName)
	fmt.Printf("Project: %s\n", relProjectPath)
	if taskID, ok := args["task_id"]; ok {
		fmt.Printf("Task: %s\n", taskID)
	}
	fmt.Println()

	// Invoke agent CLI
	var cmd *exec.Cmd
	if agentCLI == "cursor-agent" {
		// cursor-agent only takes positional prompt argument
		// Guide content is embedded in the prompt itself
		cmd = exec.Command(agentCLI,
			"--force",
			"--print",
			"--model", "sonnet-4.5",
			fullPrompt)
	} else {
		// claude supports separate system prompt
		cmd = exec.Command(agentCLI,
			"--permission-mode", "bypassPermissions",
			"--print",
			"--append-system-prompt", string(guideContent),
			fullPrompt)
	}

	// Set working directory to project path
	cmd.Dir = absProjectPath

	// Connect stdin/stdout/stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s command failed: %w", agentCLI, err)
	}

	return nil
}
