package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/input-output-hk/catalyst-forge-libs/fs/billy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ifs "github.com/input-output-hk/catalyst-forge-ai/cli/internal/fs"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "forge-ai",
	Short: "Forge AI CLI for managing AI-assisted development workflows",
	Long: `Forge AI is a structured workflow system that guides AI agents through
software development tasks using the Model Context Protocol (MCP).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Setup context
		fs := billy.NewBaseOSFS()
		ctx := context.Background()
		ctx = ifs.With(ctx, fs)
		cmd.SetContext(ctx)

		return nil
	},
}

// Execute is called by main.main(). It only needs to happen once.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.forge-ai.yaml)")

	// Bind flags to viper
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding config flag: %v\n", err)
		os.Exit(1)
	}

	// Add subcommands
	rootCmd.AddCommand(publishCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use explicit config file
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in standard locations
		viper.SetConfigName(".forge-ai")
		viper.SetConfigType("yaml")

		// Search paths in order
		viper.AddConfigPath(".")               // Current directory
		viper.AddConfigPath("$HOME")           // Home directory
		viper.AddConfigPath("$HOME/.forge-ai") // Home directory with app prefix
		viper.AddConfigPath("/etc/forge-ai")   // System directory
	}

	// Read config but don't fail if missing
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found; use defaults and env
			// This is OK - not an error condition
			return
		}
		// Config file found but another error occurred
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
	}
}
