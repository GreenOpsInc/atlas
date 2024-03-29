package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

// pipelineDeleteCmd represents the pipelineDelete command
var pipelineDeleteCmd = &cobra.Command{
	Use:   "delete <pipeline name>",
	Short: "Delete a specified pipeline",
	Long: `
Command to delete a pipeline. Specify the name of the pipeline as the argument, and team name with the flag. 

Example usage:
	atlas pipeline delete pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline delete -h' to see usage details")
			return
		}

		teamName, _ := cmd.Flags().GetString("team")
		pipelineName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/pipeline/" + orgName + "/" + teamName + "/" + pipelineName
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := getHttpClient()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}

		statusCode := resp.StatusCode
		if statusCode == 200 {
			fmt.Println("Successfully deleted pipeline:", pipelineName, "for team:", teamName)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineDeleteCmd)
}
