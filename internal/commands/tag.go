package commands

import (
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags across multiple repositories",
	Long: `Create, push, or delete tags across multiple repositories.
Tags can be created on a specific branch and pushed to remote.`,
	Run: runTag,
}

func runTag(cmd *cobra.Command, args []string) {
	// TODO: Implement tag logic
}

func GetTagCmd() *cobra.Command {
	return tagCmd
}

