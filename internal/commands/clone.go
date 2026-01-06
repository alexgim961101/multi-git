package commands

import (
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone multiple Git repositories",
	Long: `Clone multiple Git repositories defined in the configuration file.
All repositories will be cloned to the base directory specified in the config.`,
	Run: runClone,
}

func runClone(cmd *cobra.Command, args []string) {
	// TODO: Implement clone logic
}

func GetCloneCmd() *cobra.Command {
	return cloneCmd
}

