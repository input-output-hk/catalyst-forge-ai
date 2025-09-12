package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRootFromDir(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string) // returns (startDir, expectedRoot)
		expectError bool
		errorMsg    string
	}{
		{
			name: "git repo at current directory",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))

				return tmpDir, tmpDir
			},
			expectError: false,
		},
		{
			name: "git repo one level up",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))

				// Create a subdirectory
				subDir := filepath.Join(tmpDir, "subdir")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				return subDir, tmpDir
			},
			expectError: false,
		},
		{
			name: "git repo two levels up",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))

				// Create nested subdirectories
				subDir := filepath.Join(tmpDir, "level1", "level2")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				return subDir, tmpDir
			},
			expectError: false,
		},
		{
			name: "no git repo found",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "some", "nested", "path")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				return subDir, ""
			},
			expectError: true,
			errorMsg:    "no git repository found",
		},
		{
			name: "empty .git file instead of directory",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()

				// Create a .git file instead of directory
				gitFile := filepath.Join(tmpDir, ".git")
				require.NoError(t, os.WriteFile(gitFile, []byte("not a directory"), 0o644))

				subDir := filepath.Join(tmpDir, "subdir")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				return subDir, ""
			},
			expectError: true,
			errorMsg:    "no git repository found",
		},
		{
			name: "git repo at filesystem root",
			setupFunc: func(t *testing.T) (string, string) {
				// This test simulates finding .git at the root level
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				require.NoError(t, os.MkdirAll(gitDir, 0o755))

				return tmpDir, tmpDir
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startDir, expectedRoot := tt.setupFunc(t)

			root, err := FindRootFromDir(startDir)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedRoot, root)
			}
		})
	}
}

func TestFindRoot(t *testing.T) {
	// Test that FindRoot uses the current working directory
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0o755))

	// Change to the temp directory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(cwd) }()

	require.NoError(t, os.Chdir(tmpDir))

	root, err := FindRoot()
	assert.NoError(t, err)

	// Resolve symlinks to handle macOS differences (/var vs /private/var)
	expected, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)
	actual, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestIsForgeProject(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string // returns projectRoot
		expected bool
	}{
		{
			name: "is forge project",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				forgeDir := filepath.Join(tmpDir, ".forge")
				require.NoError(t, os.MkdirAll(forgeDir, 0o755))
				return tmpDir
			},
			expected: true,
		},
		{
			name: "not a forge project",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			expected: false,
		},
		{
			name: "empty .forge file instead of directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				forgeFile := filepath.Join(tmpDir, ".forge")
				require.NoError(t, os.WriteFile(forgeFile, []byte("not a directory"), 0o644))
				return tmpDir
			},
			expected: false,
		},
		{
			name: "nested forge directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				forgeDir := filepath.Join(tmpDir, ".forge")
				require.NoError(t, os.MkdirAll(forgeDir, 0o755))

				// Create a subdirectory
				subDir := filepath.Join(tmpDir, "some", "nested", "path")
				require.NoError(t, os.MkdirAll(subDir, 0o755))

				return tmpDir
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := tt.setup(t)
			result := IsForgeProject(projectRoot)
			assert.Equal(t, tt.expected, result)
		})
	}
}
