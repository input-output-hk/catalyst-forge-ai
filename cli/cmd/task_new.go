package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var taskNewTitle string

var taskNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Forge AI task",
	Long: `Create a new task with the specified title.

The command will:
- Generate a unique task ID
- Create a task directory
- Copy the default task template
- Register the task in the project

Example:
  forge-ai task new --title="Implement user authentication"`,
	RunE: runTaskNew,
}

func init() {
	taskNewCmd.Flags().StringVar(&taskNewTitle, "title", "", "Title of the new task")
	if err := taskNewCmd.MarkFlagRequired("title"); err != nil {
		panic(fmt.Sprintf("failed to mark title flag as required: %v", err))
	}
	if err := viper.BindPFlag("task.new.title", taskNewCmd.Flags().Lookup("title")); err != nil {
		panic(fmt.Sprintf("failed to bind title flag: %v", err))
	}

	taskCmd.AddCommand(taskNewCmd)
}

func runTaskNew(cmd *cobra.Command, args []string) error {
	title := viper.GetString("task.new.title")

	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Get current working directory as project root
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Verify this is a Forge AI project by checking for .forge directory
	forgeDir := filepath.Join(projectRoot, ".forge")
	if _, statErr := os.Stat(forgeDir); os.IsNotExist(statErr) {
		return fmt.Errorf("not a Forge AI project (missing .forge directory)")
	}

	// Generate task ID
	taskID, err := generateTaskID(projectRoot, title)
	if err != nil {
		return fmt.Errorf("failed to generate task ID: %w", err)
	}

	// Create task directory
	taskDir := filepath.Join(projectRoot, "tasks", taskID)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		return fmt.Errorf("failed to create task directory %s: %w", taskDir, err)
	}

	// Copy and customize task template
	if err := createTaskFromTemplate(projectRoot, taskDir, taskID, title); err != nil {
		return fmt.Errorf("failed to create task from template: %w", err)
	}

	// Register task in project.yaml
	if err := registerTaskInProject(projectRoot, taskID); err != nil {
		return fmt.Errorf("failed to register task in project: %w", err)
	}

	fmt.Printf("Created task '%s' with ID '%s'\n", title, taskID)
	return nil
}

// createTaskFromTemplate copies the default task template and customizes it
func createTaskFromTemplate(projectRoot, taskDir, taskID, title string) error {
	// Source template path
	templatePath := filepath.Join(projectRoot, ".forge", "ai", "templates", "tasks", "default.yaml")
	targetPath := filepath.Join(taskDir, "task.yaml")

	// Read template
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Replace placeholders
	taskYAML := string(templateContent)
	taskYAML = strings.ReplaceAll(taskYAML, "id: \"\"", fmt.Sprintf("id: \"%s\"", taskID))
	taskYAML = strings.ReplaceAll(taskYAML, "title: \"\"", fmt.Sprintf("title: \"%s\"", title))
	taskYAML = strings.ReplaceAll(taskYAML, "created_at: \"\"", fmt.Sprintf("created_at: \"%s\"", time.Now().Format(time.RFC3339)))

	// Write customized template
	if err := os.WriteFile(targetPath, []byte(taskYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write task.yaml: %w", err)
	}

	return nil
}

// registerTaskInProject adds the new task to the project.yaml
func registerTaskInProject(projectRoot, taskID string) error {
	projectYAMLPath := filepath.Join(projectRoot, ".forge", "project.yaml")

	// Read existing project.yaml
	content, err := os.ReadFile(projectYAMLPath)
	if err != nil {
		return fmt.Errorf("failed to read project.yaml: %w", err)
	}

	// Parse YAML
	var project map[string]interface{}
	if unmarshalErr := yaml.Unmarshal(content, &project); unmarshalErr != nil {
		return fmt.Errorf("failed to parse project.yaml: %w", unmarshalErr)
	}

	// Ensure tasks map exists
	if project["tasks"] == nil {
		project["tasks"] = make(map[string]interface{})
	}

	tasks, ok := project["tasks"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid tasks structure in project.yaml")
	}

	// Add new task entry
	tasks[taskID] = map[string]interface{}{
		"path":             fmt.Sprintf("tasks/%s", taskID),
		"phase":            "planning",
		"status":           "active",
		"template_version": "v1",
		"created_at":       time.Now().Format(time.RFC3339),
	}

	// Write back to file
	updatedContent, err := yaml.Marshal(project)
	if err != nil {
		return fmt.Errorf("failed to marshal project.yaml: %w", err)
	}

	if err := os.WriteFile(projectYAMLPath, updatedContent, 0o644); err != nil {
		return fmt.Errorf("failed to write project.yaml: %w", err)
	}

	return nil
}
