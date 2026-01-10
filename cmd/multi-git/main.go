package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexgim961101/multi-git/internal/commands"
	"github.com/spf13/cobra"
)

var (
	version    = "1.0.0"
	configPath string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "multi-git",
	Short: "Multi-Git is a CLI tool for managing multiple Git repositories",
	Long: `Multi-Git is a CLI tool that helps DevOps engineers efficiently manage multiple Git repositories.
It provides commands to clone, checkout, tag, and push across multiple repositories simultaneously.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// Root command without subcommand - show help
		cmd.Help()
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}
	defaultConfigPath := filepath.Join(homeDir, ".multi-git", "config.yaml")

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", defaultConfigPath, "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	// Register subcommands
	rootCmd.AddCommand(commands.GetCloneCmd())
	rootCmd.AddCommand(commands.GetCheckoutCmd())
	rootCmd.AddCommand(commands.GetTagCmd())
	rootCmd.AddCommand(commands.GetPushCmd())
	rootCmd.AddCommand(commands.GetExecCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
