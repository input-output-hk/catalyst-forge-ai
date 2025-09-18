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

// Task phase constants
const (
	PhasePlanning       = "planning"
	PhaseImplementation = "implementation"
	PhaseValidation     = "validation"
)

// Task status constants
const (
	StatusInProgress = "in-progress"
	StatusCompleted  = "completed"
)

// TaskState wraps a schemas.Task and provides FS-backed operations.
type TaskState struct {
	fs   fs.Filesystem
	root string
	id   string
	val  *schemas.Task
}

// NewTaskState loads task state from the filesystem for the given task ID.
func NewTaskState(ctx context.Context, filesystem fs.Filesystem, root, id string) (*TaskState, error) {
	path := filepath.Join(root, ".forge", "ai", "tasks", id, "task.yaml")
	b, err := filesystem.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read task file for %s: %w", id, err)
	}
	var t schemas.Task
	var ydoc any
	if unmarshalErr := yaml.Unmarshal(b, &ydoc); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse task.yaml: %w", unmarshalErr)
	}
	jdoc, err := json.Marshal(ydoc)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize task.yaml: %w", err)
	}
	if err := json.Unmarshal(jdoc, &t); err != nil {
		return nil, fmt.Errorf("failed to decode task.yaml: %w", err)
	}
	return &TaskState{fs: filesystem, root: root, id: id, val: &t}, nil
}

// Value returns the current task state.
func (s *TaskState) Value() *schemas.Task { return s.val }

// NormalizedStep is a strongly-typed view of a step entry.
type NormalizedStep struct {
	ID              string
	Description     string
	AIFunction      string
	Status          string
	SuccessCriteria []string
}

// NormalizedPhaseSteps returns all steps for a phase as normalized structs.
func (s *TaskState) NormalizedPhaseSteps(phase string) ([]NormalizedStep, error) {
	p, err := s.getPhaseByName(phase)
	if err != nil {
		return nil, err
	}
	out := make([]NormalizedStep, 0, len(p.Steps))
	for _, it := range p.Steps {
		m, ok := it.(map[string]any)
		if !ok {
			out = append(out, NormalizedStep{})
			continue
		}
		// success_criteria list of strings
		var sc []string
		if raw, ok := m["success_criteria"]; ok {
			if arr, ok := raw.([]any); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						sc = append(sc, s)
					}
				}
			}
		}
		out = append(out, NormalizedStep{
			ID:              getStringField(m, "id"),
			Description:     getStringField(m, "description"),
			AIFunction:      getStringField(m, "ai_function"),
			Status:          getStringField(m, "status"),
			SuccessCriteria: sc,
		})
	}
	return out, nil
}

// StepInput captures required fields to create a new step.
type StepInput struct {
	ID              string
	Description     string
	AIFunction      string
	SuccessCriteria []string
}

// AddStep appends a step to the specified target phase. Allowed only while
// current phase is planning, the target phase is modifiable, and is not the
// current phase.
func (s *TaskState) AddStep(ctx context.Context, targetPhase string, step StepInput) error {
	if s.val.Current_phase != PhasePlanning {
		return fmt.Errorf("cannot add steps outside planning phase")
	}
	if targetPhase == PhasePlanning {
		return fmt.Errorf("cannot add steps to current phase during planning")
	}

	phasePtr, err := s.getPhaseByName(targetPhase)
	if err != nil {
		return err
	}
	if !phasePtr.Modifiable {
		return fmt.Errorf("target phase %q is not modifiable", targetPhase)
	}

	// Build step as a JSON/YAML-compatible map to fit generated []any type
	sc := make([]any, 0, len(step.SuccessCriteria))
	for _, v := range step.SuccessCriteria {
		sc = append(sc, v)
	}
	newStep := map[string]any{
		"id":               step.ID,
		"description":      step.Description,
		"ai_function":      step.AIFunction,
		"status":           "pending",
		"success_criteria": sc,
	}

	phasePtr.Steps = append(phasePtr.Steps, newStep)

	// Re-assign to ensure mutation sticks on value receiver fields
	s.setPhaseByName(targetPhase, *phasePtr)

	return s.Validate(ctx)
}

// StepStart marks a step in the current phase as in-progress.
func (s *TaskState) StepStart(ctx context.Context, stepID string) error {
	phase := s.currentPhaseValue()
	idx, m, err := findStepByID(phase.Steps, stepID)
	if err != nil {
		return err
	}
	status := getStringField(m, "status")
	if status == StatusInProgress || status == StatusCompleted {
		return fmt.Errorf("step %s already %s", stepID, status)
	}
	m["status"] = StatusInProgress
	phase.Steps[idx] = m
	s.setPhaseByName(s.val.Current_phase, *phase)
	return s.Validate(ctx)
}

// StepComplete marks a step in the current phase as completed.
func (s *TaskState) StepComplete(ctx context.Context, stepID string) error {
	phase := s.currentPhaseValue()
	idx, m, err := findStepByID(phase.Steps, stepID)
	if err != nil {
		return err
	}
	if getStringField(m, "status") == StatusCompleted {
		return nil
	}
	m["status"] = StatusCompleted
	phase.Steps[idx] = m
	s.setPhaseByName(s.val.Current_phase, *phase)
	return s.Validate(ctx)
}

// TransitionPhase moves current_phase forward after verifying completion
// requirements for the source phase.
func (s *TaskState) TransitionPhase(ctx context.Context, next string) error {
	current := s.val.Current_phase
	if current == next {
		return nil
	}
	// Only forward transitions allowed per MVP
	allowed := map[string]string{PhasePlanning: PhaseImplementation, PhaseImplementation: PhaseValidation}
	want, ok := allowed[current]
	if !ok || want != next {
		return fmt.Errorf("invalid phase transition: %s -> %s", current, next)
	}
	// Ensure all steps in current are completed
	if !s.allStepsCompleted(current) {
		return fmt.Errorf("cannot transition from %s: incomplete steps present", current)
	}
	s.val.Current_phase = next
	return s.Validate(ctx)
}

// Helpers
func (s *TaskState) getPhaseByName(name string) (*schemas.Phase, error) {
	switch name {
	case PhasePlanning:
		p := s.val.Phases.Planning
		return &p, nil
	case "implementation":
		p := s.val.Phases.Implementation
		return &p, nil
	case "validation":
		p := s.val.Phases.Validation
		return &p, nil
	default:
		return nil, fmt.Errorf("unknown phase: %s", name)
	}
}

func (s *TaskState) setPhaseByName(name string, p schemas.Phase) {
	switch name {
	case PhasePlanning:
		s.val.Phases.Planning = p
	case "implementation":
		s.val.Phases.Implementation = p
	case "validation":
		s.val.Phases.Validation = p
	}
}

func (s *TaskState) currentPhaseValue() *schemas.Phase {
	p, _ := s.getPhaseByName(s.val.Current_phase)
	return p
}

func (s *TaskState) allStepsCompleted(phaseName string) bool {
	p, err := s.getPhaseByName(phaseName)
	if err != nil {
		return false
	}
	for _, it := range p.Steps {
		m, ok := it.(map[string]any)
		if !ok {
			return false
		}
		if getStringField(m, "status") != StatusCompleted {
			return false
		}
	}
	return true
}

func findStepByID(steps []any, id string) (int, map[string]any, error) {
	for i, it := range steps {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if getStringField(m, "id") == id {
			return i, m, nil
		}
	}
	return -1, nil, fmt.Errorf("step %s not found", id)
}

func getStringField(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}

// Validate checks the current task state against the CUE Task schema.
func (s *TaskState) Validate(ctx context.Context) error {
	return validateTaskAgainstSchema(s.fs, s.val)
}

// Save validates and writes task state atomically with a timestamped backup.
func (s *TaskState) Save(ctx context.Context) error {
	if err := s.Validate(ctx); err != nil {
		return err
	}
	path := filepath.Join(s.root, ".forge", "ai", "tasks", s.id, "task.yaml")
	// Backup
	if exists, _ := s.fs.Exists(path); exists {
		backupDir := filepath.Join(s.root, ".forge", "ai", "backups")
		if err := s.fs.MkdirAll(backupDir, 0o755); err == nil {
			backup := filepath.Join(
				backupDir,
				fmt.Sprintf("%s.%s", "task.yaml", time.Now().UTC().Format("20060102T150405Z")),
			)
			if b, err := s.fs.ReadFile(path); err == nil {
				_ = s.fs.WriteFile(backup, b, 0o644)
			}
		}
	}

	y, err := yaml.Marshal(s.val)
	if err != nil {
		return fmt.Errorf("failed to marshal task state: %w", err)
	}
	tmp := path + ".tmp"
	if err := s.fs.WriteFile(tmp, y, 0o644); err != nil {
		return fmt.Errorf("failed to write temporary task file: %w", err)
	}
	if err := s.fs.Rename(tmp, path); err != nil {
		return fmt.Errorf("failed to rename temporary task file: %w", err)
	}
	return nil
}

// validateTaskAgainstSchema validates the given task value against the
// exported CUE definition "Task" found under internal/schemas.
func validateTaskAgainstSchema(filesystem fs.Filesystem, t *schemas.Task) error {
	// Marshal value to JSON for compilation
	data, err := json.Marshal(t)
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

	def := schemaVal.LookupPath(cue.ParsePath("Task"))
	if !def.Exists() {
		return fmt.Errorf("schema definition %q not found", "Task")
	}

	unified := def.Unify(val)
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("validation failed for %s: %w", "Task", err)
	}
	return nil
}
