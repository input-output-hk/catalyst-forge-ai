package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// taskCmd represents the task command group
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task management commands",
	Long:  `Commands for creating and managing Forge AI tasks.`,
	// No Run function - this is a command group
}

func init() {
	rootCmd.AddCommand(taskCmd)
}

// generateTaskID generates a unique task ID in the format "NNN-slug"
func generateTaskID(projectRoot, title string) (string, error) {
	if title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}

	// Create slug from title
	slug := createSlug(title)

	// Find the next available number
	tasksDir := filepath.Join(projectRoot, "tasks")
	nextNum := findNextTaskNumber(tasksDir)

	// Format as NNN-slug
	return fmt.Sprintf("%03d-%s", nextNum, slug), nil
}

// createSlug converts a title to a URL-safe slug
func createSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// findNextTaskNumber finds the next available task number
func findNextTaskNumber(tasksDir string) int {
	// Default to 1 if no tasks exist
	nextNum := 1

	// Read existing task directories
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		// If tasks directory doesn't exist, return 1
		if os.IsNotExist(err) {
			return nextNum
		}
		// For other errors, we'll still try to continue with 1
		return nextNum
	}

	// Find the highest existing number
	highestNum := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Extract number from "NNN-slug" format
		if len(name) >= 3 {
			numStr := name[:3]
			if num, err := strconv.Atoi(numStr); err == nil {
				if num > highestNum {
					highestNum = num
				}
			}
		}
	}

	// Next number is highest + 1
	if highestNum > 0 {
		nextNum = highestNum + 1
	}

	return nextNum
}
