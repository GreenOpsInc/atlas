package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// pipelineCmd represents the pipeline command
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Execute pipeline operations by specifying an action (create, read, update , delete, sync, cancel) following the 'pipeline' command.",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Add an action (create, read, update , delete, sync, cancel) following 'pipeline' to perform a pipeline operation")
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.PersistentFlags().StringP("team", "", "", "team name")
	pipelineCmd.MarkPersistentFlagRequired("team")
	
}
