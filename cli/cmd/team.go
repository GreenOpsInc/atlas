package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// teamCmd represents the team command
var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Execute team operations by specifying an action (create, read, update , delete) following the 'team' command.",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Add a subcommand action (create, read, update , delete) following the 'team' command to perform a team operation")
	},
}

func init() {
	rootCmd.AddCommand(teamCmd)
}
