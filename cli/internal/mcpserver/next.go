// Package mcpserver provides MCP (Model Context Protocol) server implementation
// for Forge AI, including tool definitions and server lifecycle management.
package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
	"github.com/input-output-hk/catalyst-forge-ai/cli/internal/state"
)

// nextTool registers the `next` tool on the provided server.
func registerNextTool(s *mcp.Server) {
	// Input has no required fields for MVP
	type input struct{}
	type output struct {
		Content string `json:"content" jsonschema:"composed AI function with context or phase message"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "next",
		Description: "Return the next AI Function with injected context or a phase transition message",
		InputSchema: nil,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, output, error) {
		// Get filesystem from context
		filesystem := ifs.From(ctx)
		if filesystem == nil {
			return nil, output{}, fmt.Errorf("no filesystem in context")
		}

		// Locate active task from project state
		proj, err := state.NewProjectState(ctx, filesystem, ".")
		if err != nil {
			return nil, output{}, fmt.Errorf("failed to read project: %w", err)
		}
		project := proj.Value()
		if project.ActiveTask == "" {
			return nil, output{Content: "no active task set in project.yaml"}, nil
		}

		// Load task state
		tstate, err := state.NewTaskState(ctx, filesystem, ".", project.ActiveTask)
		if err != nil {
			return nil, output{}, fmt.Errorf("failed to read task: %w", err)
		}
		t := tstate.Value()
		// Build a simpler view using TaskState helper
		steps, err := tstate.NormalizedPhaseSteps(t.Current_phase)
		if err != nil {
			return nil, output{}, fmt.Errorf("failed to get normalized phase steps for %s: %w", t.Current_phase, err)
		}
		adapted := &taskState{
			ID:           t.Id,
			Title:        t.Title,
			CurrentPhase: t.Current_phase,
			Phases: map[string]struct {
				Status string `yaml:"status"`
				Steps  []struct {
					ID              string   `yaml:"id"`
					Description     string   `yaml:"description"`
					AIFunction      string   `yaml:"ai_function"`
					Status          string   `yaml:"status"`
					SuccessCriteria []string `yaml:"success_criteria"`
				} `yaml:"steps"`
			}{"": {}},
		}
		// attach current phase only for composition
		var composedSteps []struct {
			ID              string   `yaml:"id"`
			Description     string   `yaml:"description"`
			AIFunction      string   `yaml:"ai_function"`
			Status          string   `yaml:"status"`
			SuccessCriteria []string `yaml:"success_criteria"`
		}
		for _, s := range steps {
			composedSteps = append(composedSteps, struct {
				ID              string   `yaml:"id"`
				Description     string   `yaml:"description"`
				AIFunction      string   `yaml:"ai_function"`
				Status          string   `yaml:"status"`
				SuccessCriteria []string `yaml:"success_criteria"`
			}{ID: s.ID, Description: s.Description, AIFunction: s.AIFunction, Status: s.Status, SuccessCriteria: s.SuccessCriteria})
		}
		adapted.Phases[t.Current_phase] = struct {
			Status string `yaml:"status"`
			Steps  []struct {
				ID              string   `yaml:"id"`
				Description     string   `yaml:"description"`
				AIFunction      string   `yaml:"ai_function"`
				Status          string   `yaml:"status"`
				SuccessCriteria []string `yaml:"success_criteria"`
			} `yaml:"steps"`
		}{Status: phaseStatus(t.Current_phase, t), Steps: composedSteps}
		content, err := composeNextFunction(adapted)
		if err != nil {
			return nil, output{}, fmt.Errorf("failed to compose next function: %w", err)
		}
		return nil, output{Content: content}, nil
	})
}

// registerPlanningTools adds planning-phase tools (e.g., step_add).
func registerPlanningTools(s *mcp.Server) {
	type addInput struct {
		Phase           string   `json:"phase"            jsonschema:"target phase to add step to"`
		ID              string   `json:"id"`
		Description     string   `json:"description"`
		AIFunction      string   `json:"ai_function"`
		SuccessCriteria []string `json:"success_criteria"`
	}
	type addOutput struct {
		OK bool `json:"ok"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "step_add",
		Description: "Add a step to a modifiable phase during planning",
		InputSchema: nil,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in addInput) (*mcp.CallToolResult, addOutput, error) {
		fs := ifs.From(ctx)
		if fs == nil {
			return nil, addOutput{}, fmt.Errorf("no filesystem in context")
		}
		proj, err := state.NewProjectState(ctx, fs, ".")
		if err != nil {
			return nil, addOutput{}, fmt.Errorf("project: %w", err)
		}
		if proj.Value().ActiveTask == "" {
			return nil, addOutput{}, fmt.Errorf("no active task set")
		}
		ts, err := state.NewTaskState(ctx, fs, ".", proj.Value().ActiveTask)
		if err != nil {
			return nil, addOutput{}, fmt.Errorf("task: %w", err)
		}
		if err := ts.AddStep(ctx, in.Phase, state.StepInput{ID: in.ID, Description: in.Description, AIFunction: in.AIFunction, SuccessCriteria: in.SuccessCriteria}); err != nil {
			return nil, addOutput{}, fmt.Errorf("failed to add step %s to phase %s: %w", in.ID, in.Phase, err)
		}
		if err := ts.Save(ctx); err != nil {
			return nil, addOutput{}, fmt.Errorf("failed to save task state after adding step: %w", err)
		}
		return nil, addOutput{OK: true}, nil
	})
}

// registerImplementationTools adds tools for implementation phase.
func registerImplementationTools(s *mcp.Server) {
	type startInput struct {
		StepID string `json:"step_id"`
	}
	type completeInput struct {
		StepID   string   `json:"step_id"`
		Evidence []string `json:"evidence"`
	}
	type implOutput struct {
		OK bool `json:"ok"`
	}

	mcp.AddTool(
		s,
		&mcp.Tool{Name: "step_start", Description: "Mark step as in-progress"},
		func(ctx context.Context, _ *mcp.CallToolRequest, in startInput) (*mcp.CallToolResult, implOutput, error) {
			fs := ifs.From(ctx)
			if fs == nil {
				return nil, implOutput{}, fmt.Errorf("no filesystem in context")
			}
			proj, err := state.NewProjectState(ctx, fs, ".")
			if err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to load project state: %w", err)
			}
			if proj.Value().ActiveTask == "" {
				return nil, implOutput{}, fmt.Errorf("no active task set")
			}
			ts, err := state.NewTaskState(ctx, fs, ".", proj.Value().ActiveTask)
			if err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to load task state: %w", err)
			}
			if err := ts.StepStart(ctx, in.StepID); err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to start step %s: %w", in.StepID, err)
			}
			if err := ts.Save(ctx); err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to save task state after starting step: %w", err)
			}
			return nil, implOutput{OK: true}, nil
		},
	)

	mcp.AddTool(
		s,
		&mcp.Tool{Name: "step_complete", Description: "Mark step as completed with evidence"},
		func(ctx context.Context, _ *mcp.CallToolRequest, in completeInput) (*mcp.CallToolResult, implOutput, error) {
			if len(in.Evidence) == 0 {
				return nil, implOutput{}, fmt.Errorf("evidence is required")
			}
			fs := ifs.From(ctx)
			if fs == nil {
				return nil, implOutput{}, fmt.Errorf("no filesystem in context")
			}
			proj, err := state.NewProjectState(ctx, fs, ".")
			if err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to load project state: %w", err)
			}
			if proj.Value().ActiveTask == "" {
				return nil, implOutput{}, fmt.Errorf("no active task set")
			}
			ts, err := state.NewTaskState(ctx, fs, ".", proj.Value().ActiveTask)
			if err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to load task state: %w", err)
			}
			if err := ts.StepComplete(ctx, in.StepID); err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to complete step %s: %w", in.StepID, err)
			}
			if err := ts.Save(ctx); err != nil {
				return nil, implOutput{}, fmt.Errorf("failed to save task state after completing step: %w", err)
			}
			return nil, implOutput{OK: true}, nil
		},
	)
}

// Minimal structures for composing output (internal-only)
type taskState struct {
	ID           string `yaml:"id"`
	Title        string `yaml:"title"`
	CurrentPhase string `yaml:"current_phase"`
	Phases       map[string]struct {
		Status string `yaml:"status"`
		Steps  []struct {
			ID              string   `yaml:"id"`
			Description     string   `yaml:"description"`
			AIFunction      string   `yaml:"ai_function"`
			Status          string   `yaml:"status"`
			SuccessCriteria []string `yaml:"success_criteria"`
		} `yaml:"steps"`
	} `yaml:"phases"`
}

func composeNextFunction(t *taskState) (string, error) {
	phase := t.CurrentPhase
	ph, ok := t.Phases[phase]
	if !ok {
		return "", fmt.Errorf("unknown current phase: %s", phase)
	}
	// Find first step with status != completed
	for _, st := range ph.Steps {
		if st.Status != "completed" {
			// Load function template from .forge/ai/functions/<phase>/<FUNC>.ai.md
			funcPath := filepath.Join(".forge", "ai", "functions", phaseDir(phase), st.AIFunction+".ai.md")
			// Use the process context FS if available for reading templates would require passing FS here too.
			// For now, read via OS path since templates are in workspace; future: plumb FS.
			b, err := os.ReadFile(funcPath)
			if err != nil {
				return "", fmt.Errorf("failed to load AI function: %w", err)
			}
			// Minimal context injection: append a small context header
			return fmt.Sprintf(
				"%s\n\n<!-- Context: task=%s title=%q phase=%s step=%s -->\n",
				string(b),
				t.ID,
				t.Title,
				phase,
				st.ID,
			), nil
		}
	}
	// No incomplete steps -> phase transition message
	return fmt.Sprintf("{\"phase_complete\":true,\"current_phase\":%q}", phase), nil
}

func phaseDir(p string) string {
	// Planning functions live under planning/, implementation under implementation/, etc.
	switch p {
	case "planning":
		return "planning"
	case "implementation":
		return "implementation"
	case "validation":
		return "validation"
	default:
		return p
	}
}

func phaseStatus(phase string, t any) string { // placeholder: keep existing behavior minimal
	return "active"
}
