package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

var (
	history bool
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status <pipeline name> --team <team name>",
	Short: "Fetch the status of a pipeline",
	Long: `
Command to fetch the status of a pipeline. Specify the name of the pipeline as the argument,
and the team name the pipeline is under using the --team flag. For fetching the
pipeline status, the UVN can optionally be set with the -u flag, otherwise it will be set to LATEST by default.
For fetching the status of a pipeline step, specify the name of the step with the --step argument.
The step count can optionally be set with the -c flag, otherwise it will be set to 15 by default.
For fetching the history list of pipeline UVNs, add the --history flag. The -c flag can apply to this request as well.
 
Example usage:
	atlas status pipeline_name --team team_name (No -u flag specified means uvn is LATEST)
	atlas status pipeline_name --team team_name -u LATEST
	atlas status pipeline_name --team team_name --step step_name --c 20`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas status -h' to see usage details")
			return
		}

		stepFlagSet := cmd.Flags().Lookup("step").Changed
		uvnFlagSet := cmd.Flags().Lookup("uvn").Changed

		teamName, _ := cmd.Flags().GetString("team")
		pipelineName := args[0]

		var req *http.Request
		var url string

		if (stepFlagSet && uvnFlagSet) || (stepFlagSet && history) || (uvnFlagSet && history) {
			fmt.Println("Invalid combination of flags. Run 'atlas status -h' to see usage details")
			return
		}

		if history {
			countFlagSet := cmd.Flags().Lookup("count").Changed
			var count int
			if countFlagSet {
				count, _ = cmd.Flags().GetInt("count")
			} else {
				count = 15
			}
			url = fmt.Sprintf("https://%s/status/%s/%s/pipeline/%s/history/%s", atlasURL, orgName, teamName, pipelineName, strconv.Itoa(count))
		} else if stepFlagSet {
			stepName, _ := cmd.Flags().GetString("step")
			countFlagSet := cmd.Flags().Lookup("count").Changed
			var count int
			if countFlagSet {
				count, _ = cmd.Flags().GetInt("count")
			} else {
				count = 15
			}
			url = fmt.Sprintf("https://%s/status/%s/%s/pipeline/%s/step/%s/%s", atlasURL, orgName, teamName, pipelineName, stepName, strconv.Itoa(count))
		} else {
			var uvn string
			if uvnFlagSet {
				uvn, _ = cmd.Flags().GetString("uvn")
			} else {
				uvn = "LATEST"
			}
			url = fmt.Sprintf("https://%s/status/%s/%s/pipeline/%s/%s", atlasURL, orgName, teamName, pipelineName, uvn)
		}

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))
		client := &http.Client{Timeout: 20 * time.Second}
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
		} else if statusCode == 400 {
			fmt.Println("Invalid org name, team name, pipeline name, or uvn provided.")
		} else {
			fmt.Println("Internal server error, please try again.")
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.PersistentFlags().StringP("team", "", "", "team name")
	statusCmd.PersistentFlags().StringP("uvn", "u", "LATEST", "Pipeline UVN")
	statusCmd.PersistentFlags().IntP("count", "c", 15, "count")
	statusCmd.PersistentFlags().StringP("step", "s", "", "step name")
	statusCmd.MarkPersistentFlagRequired("team")

	statusCmd.PersistentFlags().BoolVarP(&history, "history", "", false, "get previous pipeline runs' UVNs")
}
