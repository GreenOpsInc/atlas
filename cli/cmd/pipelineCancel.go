package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

// pipelineCancelCmd represents the pipelineCancel command
var pipelineCancelCmd = &cobra.Command{
	Use:   "cancel <pipeline name>",
	Short: "Cancel a specified pipeline",
	Long: `
Command to cancel the latest running pipeline. Specify the name of the pipeline as the argument, and team name with the flag. 

Example usage:
	atlas pipeline cancel pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline cancel -h' to see usage details")
			return
		}

		teamName, _ := cmd.Flags().GetString("team")
		pipelineName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/status/" + orgName + "/" + teamName + "/pipelineRun/" + pipelineName
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}

		statusCode := resp.StatusCode
		if statusCode == 200 {
			fmt.Println("Successfully canceled pipeline:", pipelineName, "for team:", teamName)
		} else if statusCode == 400 {
			fmt.Println("Pipeline cancellation command failed. Invalid org name, team name, or pipeline name provided.")
		} else {
			fmt.Println("Internal server error, please try again.")
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineCancelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pipelineCancelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pipelineCancelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
