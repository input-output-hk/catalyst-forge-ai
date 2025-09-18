package mcpserver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// This test exercises the composeNextFunction helper with a synthetic task state
// rather than wiring full MCP I/O. It validates selection of the first incomplete step
// and phase-complete behavior.
func TestComposeNextFunction_SelectsFirstIncompleteStep(t *testing.T) {
	tstate := &taskState{
		ID:           "001-demo",
		Title:        "Demo Task",
		CurrentPhase: "planning",
		Phases: map[string]struct {
			Status string `yaml:"status"`
			Steps  []struct {
				ID              string   `yaml:"id"`
				Description     string   `yaml:"description"`
				AIFunction      string   `yaml:"ai_function"`
				Status          string   `yaml:"status"`
				SuccessCriteria []string `yaml:"success_criteria"`
			} `yaml:"steps"`
		}{
			"planning": {
				Status: "active",
				Steps: []struct {
					ID              string   `yaml:"id"`
					Description     string   `yaml:"description"`
					AIFunction      string   `yaml:"ai_function"`
					Status          string   `yaml:"status"`
					SuccessCriteria []string `yaml:"success_criteria"`
				}{
					{ID: "discover", AIFunction: "DISCOVER", Status: "pending"},
				},
			},
		},
	}

	// Create a fake DISCOVER.ai.md in a temp .forge/ai/functions/planning directory
	tmp := t.TempDir()
	funcDir := filepath.Join(tmp, ".forge", "ai", "functions", "planning")
	require.NoError(t, os.MkdirAll(funcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(funcDir, "DISCOVER.ai.md"), []byte("# DISCOVER\nBody"), 0o644))

	// Change CWD so helper can resolve relative paths
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	require.NoError(t, os.Chdir(tmp))

	out, err := composeNextFunction(tstate)
	require.NoError(t, err)
	require.Contains(t, out, "DISCOVER")
	require.Contains(t, out, "Context: task=001-demo")
}

func TestComposeNextFunction_PhaseComplete(t *testing.T) {
	tstate := &taskState{
		ID:           "001-demo",
		Title:        "Demo Task",
		CurrentPhase: "planning",
		Phases: map[string]struct {
			Status string `yaml:"status"`
			Steps  []struct {
				ID              string   `yaml:"id"`
				Description     string   `yaml:"description"`
				AIFunction      string   `yaml:"ai_function"`
				Status          string   `yaml:"status"`
				SuccessCriteria []string `yaml:"success_criteria"`
			} `yaml:"steps"`
		}{
			"planning": {Status: "active", Steps: nil},
		},
	}

	out, err := composeNextFunction(tstate)
	require.NoError(t, err)
	require.Contains(t, out, "\"phase_complete\":true")
}
