// Package state provides filesystem-backed state management for Forge AI projects
// and tasks, including loading, validation, and atomic saving operations.
package state

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/input-output-hk/catalyst-forge-libs/fs"
	"gopkg.in/yaml.v3"

	cuehelpers "github.com/input-output-hk/catalyst-forge-ai/cli/internal/cue"
	"github.com/input-output-hk/catalyst-forge-ai/cli/internal/schemas"
)

// ProjectState wraps a schemas.Project and provides FS-backed operations.
type ProjectState struct {
	fs   fs.Filesystem
	root string
	val  *schemas.Project
}

// NewProjectState loads project state from root using the provided filesystem.
func NewProjectState(ctx context.Context, filesystem fs.Filesystem, root string) (*ProjectState, error) {
	// Prefer .forge/ai/project.yaml, fallback to .forge/project.yaml
	primary := filepath.Join(root, ".forge", "ai", "project.yaml")
	legacy := filepath.Join(root, ".forge", "project.yaml")

	var b []byte
	var err error
	if exists, _ := filesystem.Exists(primary); exists {
		b, err = filesystem.ReadFile(primary)
		if err != nil {
			return nil, fmt.Errorf("failed to read primary project file: %w", err)
		}
	} else if exists, _ = filesystem.Exists(legacy); exists {
		b, err = filesystem.ReadFile(legacy)
		if err != nil {
			return nil, fmt.Errorf("failed to read legacy project file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("project state not found: %s", primary)
	}

	var p schemas.Project
	// Prefer YAML->map->JSON roundtrip to respect json struct tags
	var ydoc any
	if unmarshalErr := yaml.Unmarshal(b, &ydoc); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse project.yaml: %w", unmarshalErr)
	}
	jdoc, err := json.Marshal(ydoc)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize project.yaml: %w", err)
	}
	if err := json.Unmarshal(jdoc, &p); err != nil {
		return nil, fmt.Errorf("failed to decode project.yaml: %w", err)
	}
	return &ProjectState{fs: filesystem, root: root, val: &p}, nil
}

// Value returns the current project state.
func (s *ProjectState) Value() *schemas.Project { return s.val }

// SetActiveTask updates the project's active task reference.
func (s *ProjectState) SetActiveTask(ctx context.Context, taskID string) error {
	s.val.ActiveTask = taskID
	return s.Validate(ctx)
}

// ProjectTaskEntryInput is an input struct to register/update project task entries.
type ProjectTaskEntryInput struct {
	Path            string
	Phase           string
	Status          string
	TemplateVersion string
	CreatedAt       string
	CompletedAt     string
}

// RegisterTask adds or updates a task entry in the project registry.
func (s *ProjectState) RegisterTask(ctx context.Context, id string, in *ProjectTaskEntryInput) error {
	if s.val.Tasks == nil {
		s.val.Tasks = make(map[string]struct {
			Path   string `json:"path"`
			Phase  string `json:"phase"`
			Status string `json:"status"`
			//revive:disable:var-naming
			Template_version string `json:"template_version"`
			Created_at       string `json:"created_at,omitempty"`
			Completed_at     string `json:"completed_at,omitempty"`
			//revive:enable:var-naming
		})
	}
	entry := struct {
		Path   string `json:"path"`
		Phase  string `json:"phase"`
		Status string `json:"status"`
		//revive:disable:var-naming
		Template_version string `json:"template_version"`
		Created_at       string `json:"created_at,omitempty"`
		Completed_at     string `json:"completed_at,omitempty"`
		//revive:enable:var-naming
	}{
		Path:             in.Path,
		Phase:            in.Phase,
		Status:           in.Status,
		Template_version: in.TemplateVersion,
		Created_at:       in.CreatedAt,
		Completed_at:     in.CompletedAt,
	}
	s.val.Tasks[id] = entry
	return s.Validate(ctx)
}

// UpdateTaskPhase sets a task's recorded phase in the project registry.
func (s *ProjectState) UpdateTaskPhase(ctx context.Context, id, phase string) error {
	t, ok := s.val.Tasks[id]
	if !ok {
		return fmt.Errorf("task %s not registered", id)
	}
	t.Phase = phase
	s.val.Tasks[id] = t
	return s.Validate(ctx)
}

// UpdateTaskStatus sets a task's recorded status in the project registry.
func (s *ProjectState) UpdateTaskStatus(ctx context.Context, id, status string) error {
	t, ok := s.val.Tasks[id]
	if !ok {
		return fmt.Errorf("task %s not registered", id)
	}
	t.Status = status
	s.val.Tasks[id] = t
	return s.Validate(ctx)
}

// GetActiveTask returns the active task ID and validates that the task directory exists.
func (s *ProjectState) GetActiveTask(ctx context.Context) (string, error) {
	activeTask := s.val.ActiveTask
	if activeTask == "" {
		return "", fmt.Errorf("no active task set in project")
	}

	// Verify the task directory exists
	taskDir := filepath.Join(s.root, "tasks", activeTask)
	if exists, err := s.fs.Exists(taskDir); err != nil {
		return "", fmt.Errorf("failed to check task directory: %w", err)
	} else if !exists {
		return "", fmt.Errorf("active task directory does not exist: %s", taskDir)
	}

	return activeTask, nil
}

// Validate checks the current project state against the CUE Project schema.
func (s *ProjectState) Validate(ctx context.Context) error {
	return validateProjectAgainstSchema(s.fs, s.val)
}

// Save validates and writes project atomically with a timestamped backup.
func (s *ProjectState) Save(ctx context.Context) error {
	if err := s.Validate(ctx); err != nil {
		return err
	}
	path := filepath.Join(s.root, ".forge", "ai", "project.yaml")
	// Backup
	if exists, _ := s.fs.Exists(path); exists {
		backupDir := filepath.Join(s.root, ".forge", "ai", "backups")
		if err := s.fs.MkdirAll(backupDir, 0o755); err == nil {
			backup := filepath.Join(
				backupDir,
				fmt.Sprintf("project.yaml.%s", time.Now().UTC().Format("20060102T150405Z")),
			)
			if b, err := s.fs.ReadFile(path); err == nil {
				_ = s.fs.WriteFile(backup, b, 0o644)
			}
		}
	}

	// Write temp then rename
	y, err := yaml.Marshal(s.val)
	if err != nil {
		return fmt.Errorf("failed to marshal project state: %w", err)
	}
	tmp := path + ".tmp"
	if err := s.fs.WriteFile(tmp, y, 0o644); err != nil {
		return fmt.Errorf("failed to write temporary project file: %w", err)
	}
	if err := s.fs.Rename(tmp, path); err != nil {
		return fmt.Errorf("failed to rename temporary project file: %w", err)
	}
	return nil
}

// validateProjectAgainstSchema validates the given project value against the
// exported CUE definition "Project" found under internal/schemas.
func validateProjectAgainstSchema(filesystem fs.Filesystem, p *schemas.Project) error {
	// Marshal value to JSON for compilation
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal value for validation: %w", err)
	}

	ctx := cuecontext.New()
	val := ctx.CompileBytes(data)
	if val.Err() != nil {
		return fmt.Errorf("invalid data for validation: %w", val.Err())
	}

	// Build CUE instance from schemas using helper
	base := filepath.Join("internal", "schemas")
	schemaVal, err := cuehelpers.BuildSchemaValue(ctx, filesystem, base)
	if err != nil {
		return fmt.Errorf("failed to build CUE schema instance: %w", err)
	}

	def := schemaVal.LookupPath(cue.ParsePath("Project"))
	if !def.Exists() {
		return fmt.Errorf("schema definition %q not found", "Project")
	}

	unified := def.Unify(val)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("validation failed for %s: %w", "Project", err)
	}
	return nil
}
