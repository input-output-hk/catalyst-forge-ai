package cmd

import (
	"context"
	"fmt"

	vfs "github.com/input-output-hk/catalyst-forge-libs/fs"
	"github.com/spf13/cobra"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
	"github.com/input-output-hk/catalyst-forge-ai/cli/internal/repo"
	"github.com/input-output-hk/catalyst-forge-ai/cli/internal/state"
)

// taskPhaseCmd represents the task phase command group
var taskPhaseCmd = &cobra.Command{
	Use:   "phase",
	Short: "Phase management commands for tasks",
	Long: `Commands for managing task phases and phase transitions.

This command group provides tools to advance tasks through their lifecycle phases
(planning → implementation → validation) with proper validation and state management.

Phase transitions are human gates that ensure all prerequisites are met before
moving to the next phase.`,
	// No Run function - this is a command group
}

func init() {
	taskCmd.AddCommand(taskPhaseCmd)
}

// taskPhaseNextCmd represents the task phase next command
var taskPhaseNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Advance the current task to the next phase",
	Long: `Advance the current task to the next phase after validating prerequisites.

This command implements the human gate between task phases. It ensures that:
- All steps in the current phase are completed before transition
- The phase transition follows the allowed sequence
- Both task.yaml and project.yaml are updated consistently

Valid phase transitions:
- planning → implementation (after all planning steps complete)
- implementation → validation (after all implementation steps complete/abandoned)

Example:
  forge-ai task phase next`,
	RunE: runTaskPhaseNext,
}

func init() {
	taskPhaseCmd.AddCommand(taskPhaseNextCmd)
}

// findActiveTask is injectable for tests; defaults to finding active task from project state
var findActiveTask = func(ctx context.Context, filesystem vfs.Filesystem, projectRoot string) (string, error) {
	projectState, err := state.NewProjectState(ctx, filesystem, projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load project state: %w", err)
	}
	return projectState.GetActiveTask(ctx)
}

func runTaskPhaseNext(cmd *cobra.Command, args []string) error {
	// Find the git repository root
	filesystem := ifs.From(cmd.Context())
	if filesystem == nil {
		return fmt.Errorf("no filesystem in context")
	}
	projectRoot, err := repo.FindRoot(filesystem)
	if err != nil {
		return fmt.Errorf("failed to find git repository: %w", err)
	}

	// Verify this is a Forge AI project
	if !repo.IsForgeProject(filesystem, projectRoot) {
		return fmt.Errorf("not a Forge AI project (missing .forge/ai directory in %s)", projectRoot)
	}

	// Find the active task
	taskID, err := findActiveTask(cmd.Context(), filesystem, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to find active task: %w", err)
	}

	// Load task state
	taskState, err := state.NewTaskState(cmd.Context(), filesystem, projectRoot, taskID)
	if err != nil {
		return fmt.Errorf("failed to load task state: %w", err)
	}

	// Determine the next phase
	currentPhase := taskState.Value().Current_phase
	var nextPhase string
	switch currentPhase {
	case "planning":
		nextPhase = "implementation"
	case "implementation":
		nextPhase = "validation"
	case "validation":
		return fmt.Errorf("task is already in final phase (validation)")
	default:
		return fmt.Errorf("unknown current phase: %s", currentPhase)
	}

	// Attempt phase transition (this includes validation)
	if transitionErr := taskState.TransitionPhase(cmd.Context(), nextPhase); transitionErr != nil {
		return fmt.Errorf("cannot transition to %s phase: %w", nextPhase, transitionErr)
	}

	// Save task state
	if saveErr := taskState.Save(cmd.Context()); saveErr != nil {
		return fmt.Errorf("failed to save task state: %w", saveErr)
	}

	// Update project registry
	projectState, err := state.NewProjectState(cmd.Context(), filesystem, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project state: %w", err)
	}

	if err := projectState.UpdateTaskPhase(cmd.Context(), taskID, nextPhase); err != nil {
		return fmt.Errorf("failed to update project registry: %w", err)
	}

	if err := projectState.Save(cmd.Context()); err != nil {
		return fmt.Errorf("failed to save project state: %w", err)
	}

	fmt.Printf("✓ Advanced task '%s' from %s to %s phase\n", taskID, currentPhase, nextPhase)
	return nil
}
