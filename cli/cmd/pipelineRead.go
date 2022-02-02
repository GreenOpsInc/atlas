package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

// pipelineReadCmd represents the pipelineRead command
var pipelineReadCmd = &cobra.Command{
	Use:   "read <pipeline name",
	Short: "Read a specified pipeline's details",
	Long: `
Command to read a pipeline. Specify the names of the pipeline as the argument, and the team name using the flag. 

Example usage:
	atlas pipeline read pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline read -h' to see usage details")
			return
		}

		teamName, _ := cmd.Flags().GetString("team")
		pipelineName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/pipeline/" + orgName + "/" + teamName + "/" + pipelineName

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := getHttpClient()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		statusCode := resp.StatusCode
		if statusCode == 200 {
			var prettyJSON bytes.Buffer
			error := json.Indent(&prettyJSON, body, "", "\t")
			if error != nil {
				fmt.Println("Request failed, please try again.")
				return
			}
			fmt.Println(string(prettyJSON.Bytes()))
		} else {
			fmt.Printf("Error: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineReadCmd)
}
