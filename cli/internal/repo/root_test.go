package repo

import (
	"path/filepath"
	"testing"

	"github.com/input-output-hk/catalyst-forge-libs/fs"
	billy "github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRootFromDir(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, f fs.Filesystem) (string, string) // returns (startDir, expectedRoot)
		expectError bool
		errorMsg    string
	}{
		{
			name: "git repo at current directory",
			setupFunc: func(t *testing.T, f fs.Filesystem) (string, string) {
				root := "/repo"
				require.NoError(t, f.MkdirAll(filepath.Join(root, ".git"), 0o755))
				return root, root
			},
			expectError: false,
		},
		{
			name: "git repo one level up",
			setupFunc: func(t *testing.T, f fs.Filesystem) (string, string) {
				root := "/repo"
				require.NoError(t, f.MkdirAll(filepath.Join(root, ".git"), 0o755))
				subDir := filepath.Join(root, "subdir")
				require.NoError(t, f.MkdirAll(subDir, 0o755))
				return subDir, root
			},
			expectError: false,
		},
		{
			name: "git repo two levels up",
			setupFunc: func(t *testing.T, f fs.Filesystem) (string, string) {
				root := "/repo"
				require.NoError(t, f.MkdirAll(filepath.Join(root, ".git"), 0o755))
				subDir := filepath.Join(root, "level1", "level2")
				require.NoError(t, f.MkdirAll(subDir, 0o755))
				return subDir, root
			},
			expectError: false,
		},
		{
			name: "no git repo found",
			setupFunc: func(t *testing.T, f fs.Filesystem) (string, string) {
				subDir := "/repo/some/nested/path"
				require.NoError(t, f.MkdirAll(subDir, 0o755))
				return subDir, ""
			},
			expectError: true,
			errorMsg:    "no git repository found",
		},
		{
			name: "empty .git file instead of directory",
			setupFunc: func(t *testing.T, f fs.Filesystem) (string, string) {
				root := "/repo"
				require.NoError(t, f.MkdirAll(root, 0o755))
				require.NoError(t, f.WriteFile(filepath.Join(root, ".git"), []byte("not a directory"), 0o644))
				subDir := filepath.Join(root, "subdir")
				require.NoError(t, f.MkdirAll(subDir, 0o755))
				return subDir, ""
			},
			expectError: true,
			errorMsg:    "no git repository found",
		},
		{
			name: "git repo at filesystem root",
			setupFunc: func(t *testing.T, f fs.Filesystem) (string, string) {
				root := "/"
				require.NoError(t, f.MkdirAll(filepath.Join(root, ".git"), 0o755))
				return root, root
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := billy.NewInMemoryFS()
			startDir, expectedRoot := tt.setupFunc(t, f)

			root, err := FindRootFromDir(f, startDir)

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
	f := billy.NewInMemoryFS()
	rootPath := "/repo"
	require.NoError(t, f.MkdirAll(filepath.Join(rootPath, ".git"), 0o755))
	root, err := FindRootFromDir(f, rootPath)
	assert.NoError(t, err)
	assert.Equal(t, rootPath, root)
}

func TestIsForgeProject(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, f fs.Filesystem) string // returns projectRoot
		expected bool
	}{
		{
			name: "is forge project",
			setup: func(t *testing.T, f fs.Filesystem) string {
				root := "/proj1"
				require.NoError(t, f.MkdirAll(filepath.Join(root, ".forge", "ai"), 0o755))
				return root
			},
			expected: true,
		},
		{
			name: "not a forge project",
			setup: func(t *testing.T, f fs.Filesystem) string {
				return "/proj2"
			},
			expected: false,
		},
		{
			name: "empty .forge file instead of directory",
			setup: func(t *testing.T, f fs.Filesystem) string {
				root := "/proj3"
				require.NoError(t, f.WriteFile(filepath.Join(root, ".forge"), []byte("not a directory"), 0o644))
				return root
			},
			expected: false,
		},
		{
			name: "nested forge directory",
			setup: func(t *testing.T, f fs.Filesystem) string {
				root := "/proj4"
				require.NoError(t, f.MkdirAll(filepath.Join(root, ".forge", "ai"), 0o755))
				_ = f.MkdirAll(filepath.Join(root, "some", "nested", "path"), 0o755)
				return root
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := billy.NewInMemoryFS()
			projectRoot := tt.setup(t, f)
			result := IsForgeProject(f, projectRoot)
			assert.Equal(t, tt.expected, result)
		})
	}
}
