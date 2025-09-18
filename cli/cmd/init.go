package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	vfs "github.com/input-output-hk/catalyst-forge-libs/fs"
	ocibundle "github.com/input-output-hk/catalyst-forge-libs/oci"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
)

// ociClient defines the minimal interface we need for pulling templates
type ociClient interface {
	Pull(ctx context.Context, reference, targetDir string) error
}

// newOCIClient allows injection for tests
var newOCIClient = func(templateRef string) (ociClient, error) {
	var clientOpts []ocibundle.ClientOption
	if strings.HasPrefix(templateRef, "localhost:") || strings.HasPrefix(templateRef, "127.0.0.1:") {
		// Allow HTTP for localhost registries
		clientOpts = append(clientOpts, ocibundle.WithAllowHTTP())
	}

	real, err := ocibundle.NewWithOptions(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI client: %w", err)
	}
	return &ociWrapper{real: real}, nil
}

// ociWrapper adapts the real client to our minimal interface
type ociWrapper struct{ real *ocibundle.Client }

func (w *ociWrapper) Pull(ctx context.Context, reference, targetDir string) error {
	if err := w.real.Pull(ctx, reference, targetDir); err != nil {
		return fmt.Errorf("failed to pull OCI template: %w", err)
	}
	return nil
}

var initTemplate string

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Initialize a new Forge AI project from an OCI template",
	Long: `Initialize a Forge AI project in the current directory using a published template.

Example:
  forge-ai init my-project --template=ghcr.io/org/template:v1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().
		StringVar(&initTemplate, "template", "", "OCI template reference (e.g., ghcr.io/org/template:v1.0.0)")
	if err := initCmd.MarkFlagRequired("template"); err != nil {
		panic(fmt.Sprintf("failed to mark template flag as required: %v", err))
	}
	if err := viper.BindPFlag("init.template", initCmd.Flags().Lookup("template")); err != nil {
		panic(fmt.Sprintf("failed to bind template flag: %v", err))
	}

	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]
	templateRef := viper.GetString("init.template")

	if projectName == "" {
		return fmt.Errorf("project name is required")
	}
	if templateRef == "" {
		return fmt.Errorf("template reference is required")
	}
	if !isValidRegistryFormat(templateRef) {
		return fmt.Errorf("invalid registry format: %s", templateRef)
	}

	// Parse template reference into source and version
	source, version, err := parseTemplateRef(templateRef)
	if err != nil {
		return fmt.Errorf("invalid template reference: %w", err)
	}

	// Prepare directories
	forgeDir := ".forge"
	aiDir := filepath.Join(forgeDir, "ai")

	filesystem := ifs.From(cmd.Context())
	if filesystem == nil {
		return fmt.Errorf("no filesystem in context")
	}

	if err = ensureEmptyOrCreate(filesystem, aiDir); err != nil {
		return fmt.Errorf("failed to prepare template directory %s: %w", aiDir, err)
	}

	// Pull the template into .forge/ai
	client, err := newOCIClient(templateRef)
	if err != nil {
		return fmt.Errorf("failed to create OCI client: %w", err)
	}
	ctx := cmd.Context()
	if err := client.Pull(ctx, templateRef, aiDir); err != nil {
		return fmt.Errorf("failed to pull template: %w", err)
	}

	// Generate project.yaml
	projectYAML := fmt.Sprintf(
		"projectName: %s\nstatus: active\ntemplate:\n  source: %s\n  version: %s\n",
		projectName,
		source,
		version,
	)
	if err := filesystem.MkdirAll(forgeDir, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", forgeDir, err)
	}
	projectPath := filepath.Join(forgeDir, "project.yaml")
	if err := filesystem.WriteFile(projectPath, []byte(projectYAML), 0o644); err != nil {
		return fmt.Errorf("failed to write project.yaml: %w", err)
	}

	fmt.Printf("Initialized Forge AI project '%s' using template %s\n", projectName, templateRef)
	return nil
}

// parseTemplateRef splits an OCI reference like ghcr.io/org/repo:v1.2.3
// into source (ghcr.io/org/repo) and version (v1.2.3).
func parseTemplateRef(ref string) (string, string, error) {
	// Reject digests for init (require tag)
	if strings.Contains(ref, "@") {
		return "", "", fmt.Errorf("digest references are not supported; use a tag like v1.2.3")
	}
	lastColon := strings.LastIndex(ref, ":")
	if lastColon == -1 {
		return "", "", fmt.Errorf("missing tag (expected :vX.Y.Z)")
	}
	source := ref[:lastColon]
	version := ref[lastColon+1:]

	// Validate semver-like vX.Y.Z
	re := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	if !re.MatchString(version) {
		return "", "", fmt.Errorf("invalid version tag %q; expected vX.Y.Z", version)
	}
	return source, version, nil
}

// ensureEmptyOrCreate ensures dir exists and is empty (compatible with lib/oci Pull)
func ensureEmptyOrCreate(filesystem vfs.Filesystem, dir string) error {
	exists, err := filesystem.Exists(dir)
	if err != nil {
		return fmt.Errorf("failed to check directory existence: %w", err)
	}
	if !exists {
		if err = filesystem.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	}
	entries, err := filesystem.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("target directory is not empty")
	}
	return nil
}

// Avoid unused imports when built without using time directly elsewhere
var _ = time.Second
