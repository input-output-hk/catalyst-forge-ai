package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	oci "github.com/input-output-hk/catalyst-forge-ai/lib/oci"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	publishSource   string
	publishRegistry string
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a template to an OCI registry",
	Long: `Publish packages a template source directory and pushes it to an OCI registry.

Example:
  forge-ai publish --source=./template-source --registry=ghcr.io/org/template:v1.0.0`,
	RunE: runPublish,
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVar(&publishSource, "source", "",
		"Path to template source directory")
	publishCmd.Flags().StringVar(&publishRegistry, "registry", "",
		"OCI registry reference (e.g., ghcr.io/org/template:v1.0.0)")

	// Mark flags as required
	publishCmd.MarkFlagRequired("source")
	publishCmd.MarkFlagRequired("registry")

	// Bind flags to viper
	viper.BindPFlag("publish.source", publishCmd.Flags().Lookup("source"))
	viper.BindPFlag("publish.registry", publishCmd.Flags().Lookup("registry"))
}

// newPublishCmd creates a new publish command (used for testing)
func newPublishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a template to an OCI registry",
		Long: `Publish packages a template source directory and pushes it to an OCI registry.

Example:
  forge-ai publish --source=./template-source --registry=ghcr.io/org/template:v1.0.0`,
		RunE: runPublish,
	}

	cmd.Flags().StringVar(&publishSource, "source", "",
		"Path to template source directory")
	cmd.Flags().StringVar(&publishRegistry, "registry", "",
		"OCI registry reference (e.g., ghcr.io/org/template:v1.0.0)")

	return cmd
}

func runPublish(cmd *cobra.Command, args []string) error {
	// Get values from viper (respects flags, env, config)
	source := viper.GetString("publish.source")
	registry := viper.GetString("publish.registry")

	// Validate arguments
	if err := validatePublishArgs(source, registry); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if source directory exists
	if !sourceDirectoryExists(source) {
		return fmt.Errorf("source directory does not exist: %s", source)
	}

	// Create OCI client using lib/oci
	client, err := oci.New()
	if err != nil {
		return fmt.Errorf("failed to create OCI client: %w", err)
	}

	// Create context
	ctx := context.Background()

	// Use the client's Push method to publish the template
	fmt.Printf("Publishing template from %s to %s...\n", source, registry)

	if err := client.Push(ctx, source, registry); err != nil {
		return fmt.Errorf("failed to push template: %w", err)
	}

	fmt.Printf("Successfully published template to %s\n", registry)
	return nil
}

func validatePublishArgs(source, registry string) error {
	if source == "" {
		return fmt.Errorf("source directory is required")
	}

	if registry == "" {
		return fmt.Errorf("registry reference is required")
	}

	if !isValidRegistryFormat(registry) {
		return fmt.Errorf("invalid registry format: %s", registry)
	}

	return nil
}

func isValidRegistryFormat(registry string) bool {
	// Basic validation for OCI registry reference
	// Format: [host[:port]/]namespace/repository[:tag|@digest]

	// Must not start with http:// or https://
	if strings.HasPrefix(registry, "http://") || strings.HasPrefix(registry, "https://") {
		return false
	}

	// Must contain at least one slash (namespace/repository)
	if !strings.Contains(registry, "/") {
		return false
	}

	// Basic pattern: should have host/path structure
	// Allow localhost, domain names, and IP addresses with optional port
	pattern := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(:[0-9]+)?/[a-z0-9]+([._\-/][a-z0-9]+)*(:[a-zA-Z0-9][\w.-]*)?(@sha256:[a-f0-9]{64})?$|^localhost(:[0-9]+)?/[a-z0-9]+([._\-/][a-z0-9]+)*(:[a-zA-Z0-9][\w.-]*)?(@sha256:[a-f0-9]{64})?$`

	matched, _ := regexp.MatchString(pattern, registry)
	return matched
}

func sourceDirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
