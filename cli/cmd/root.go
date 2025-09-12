package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "forge-ai",
	Short: "Forge AI CLI for managing AI-assisted development workflows",
	Long: `Forge AI is a structured workflow system that guides AI agents through
software development tasks using the Model Context Protocol (MCP).`,
	// Root command typically has no Run function
}

// Execute is called by main.main(). It only needs to happen once.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.forge-ai.yaml)")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}

func initConfig() {
	if cfgFile != "" {
		// Use explicit config file
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in standard locations
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Search paths in order
		viper.AddConfigPath(".")               // Current directory
		viper.AddConfigPath("$HOME/.forge-ai") // Home directory
		viper.AddConfigPath("/etc/forge-ai")   // System directory
	}

	// Read config but don't fail if missing
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults and env
			// This is OK - not an error condition
		} else {
			// Config file found but another error occurred
			fmt.Printf("Error reading config: %v\n", err)
		}
	}
}

// newRootCmd creates a new root command (used for testing)
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forge-ai",
		Short: "Forge AI CLI for managing AI-assisted development workflows",
		Long: `Forge AI is a structured workflow system that guides AI agents through
software development tasks using the Model Context Protocol (MCP).`,
	}

	// Add subcommands
	cmd.AddCommand(newPublishCmd())

	return cmd
}
