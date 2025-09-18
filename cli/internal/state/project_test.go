package state

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// writeSchemas writes minimal CUE schemas required for validation into the
// in-memory filesystem at internal/schemas/.
func writeSchemas(t *testing.T, fs *billy.FS) {
	t.Helper()
	// Ensure schema directory exists
	require.NoError(t, fs.MkdirAll(filepath.Join("internal", "schemas"), 0o755))

	projectCue := []byte(`package schemas

// Exported alias for Go type generation
Project: #Project

#Project: {
    projectName: string | *""
    template: {
        source:  string & !=""
        version: string & !=""
    }
    status: "active" | "maintenance" | "archived" | *"active"
    activeTask?: string & !~"\\s"
    tasks?: [string]: {
        path:             string & !~"\\s"
        phase:            "planning" | "implementation" | "validation"
        status:           "active" | "blocked" | "completed"
        template_version: string & !=""
        created_at?:      string
        completed_at?:    string
    }
}
`)
	require.NoError(t, fs.WriteFile(filepath.Join("internal", "schemas", "project.cue"), projectCue, 0o644))

	taskCue := []byte(`package schemas

#Step: {
    id:          string & !~"\\s" & !=""
    description: string & !=""
    ai_function: "DISCOVER" | "PLAN" | "ASSESS" | "EXECUTE" | "VALIDATE" | "REPORT"
    status:      "pending" | "in-progress" | "completed" | *"pending"
    success_criteria: [...string & !=""] | *[]
}

#Phase: {
    status:      "active" | "pending" | "completed" | *"pending"
    modifiable?: bool | *false
    steps: [...#Step] | *[]
}

// Exported alias for Go type generation
Task: #Task

#Task: {
    id:            string & !=""
    title:         string & !=""
    created_at?:   string
    current_phase: "planning" | "implementation" | "validation"
    status:        "active" | "blocked" | "completed" | *"active"
    phases: {
        planning:       #Phase
        implementation: #Phase
        validation:     #Phase
    }
}
`)
	require.NoError(t, fs.WriteFile(filepath.Join("internal", "schemas", "task.cue"), taskCue, 0o644))
}

func TestProjectState_LoadValidateSave_Primary(t *testing.T) {
	mem := billy.NewInMemoryFS()
	writeSchemas(t, mem)

	// Write minimal valid project YAML to primary path (match code's ./ prefix)
	require.NoError(t, mem.MkdirAll(filepath.Join(".", ".forge", "ai"), 0o755))
	projectYAML := []byte("" +
		"projectName: \"My Project\"\n" +
		"template:\n" +
		"  source: \"ghcr.io/example/template\"\n" +
		"  version: \"v1.0.0\"\n" +
		"status: \"active\"\n",
	)
	require.NoError(t, mem.WriteFile(filepath.Join(".", ".forge", "ai", "project.yaml"), projectYAML, 0o644))

	// Sanity check: direct unmarshal
	{
		b, err := mem.ReadFile(filepath.Join(".", ".forge", "ai", "project.yaml"))
		require.NoError(t, err)
		var tmp struct {
			ProjectName string `yaml:"projectName"`
			Template    struct {
				Source  string `yaml:"source"`
				Version string `yaml:"version"`
			} `yaml:"template"`
			Status string `yaml:"status"`
		}
		require.NoError(t, yaml.Unmarshal(b, &tmp))
		assert.Equal(t, "My Project", tmp.ProjectName)
		assert.Equal(t, "ghcr.io/example/template", tmp.Template.Source)
		assert.Equal(t, "v1.0.0", tmp.Template.Version)
		assert.Equal(t, "active", tmp.Status)
	}

	ctx := context.Background()
	ps, err := NewProjectState(ctx, mem, ".")
	require.NoError(t, err)
	assert.Equal(t, "My Project", ps.Value().ProjectName)

	// Validate and save (should create a backup since file exists)
	require.NoError(t, ps.Validate(ctx))
	require.NoError(t, ps.Save(ctx))

	// Expect a backup file to exist
	entries, err := mem.ReadDir(filepath.Join(".forge", "ai", "backups"))
	require.NoError(t, err)
	assert.NotEmpty(t, entries)
}

func TestProjectState_LegacyFallback_SaveWritesPrimary(t *testing.T) {
	mem := billy.NewInMemoryFS()
	writeSchemas(t, mem)

	// Write legacy project.yaml path only (match code's ./ prefix for root)
	require.NoError(t, mem.MkdirAll(filepath.Join(".", ".forge"), 0o755))
	legacyYAML := []byte("" +
		"projectName: \"Legacy Project\"\n" +
		"template:\n" +
		"  source: \"ghcr.io/example/template\"\n" +
		"  version: \"v1.0.0\"\n" +
		"status: \"active\"\n",
	)
	require.NoError(t, mem.WriteFile(filepath.Join(".", ".forge", "project.yaml"), legacyYAML, 0o644))

	ctx := context.Background()
	ps, err := NewProjectState(ctx, mem, ".")
	require.NoError(t, err)
	assert.Equal(t, "Legacy Project", ps.Value().ProjectName)

	// Save should write to primary path
	require.NoError(t, ps.Save(ctx))
	_, err = mem.Stat(filepath.Join(".forge", "ai", "project.yaml"))
	assert.NoError(t, err)
}
