package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
)

// pipelineCleanCmd represents the pipelineClean command
var pipelineCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean up stale resources for a specified pipeline ",
	Long: `Command to trigger a cleanup of stale resources for a pipeline. For each cluster-namespace combination, stale resources will be dropped. 

Example usage:
	atlas pipeline clean pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline clean -h' to see usage details")
			return
		}

		pipelineName := args[0]
		teamName, _ := cmd.Flags().GetString("team")

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		groupUrl := "https://" + atlasURL + "/combinations/" + orgName + "/" + teamName + "/" + pipelineName

		req, _ := http.NewRequest("GET", groupUrl, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := getHttpClient()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		var groups ClusterNamespaceGroups
		err = json.Unmarshal(body,
			&groups)
		if err != nil {
			fmt.Println("Unmarshall failed with the following error:", err)
			return
		}

		for _, group := range groups.Groups {
			cluster := group.ClusterName
			namespace := group.Namespace
			url := "https://" + atlasURL + "/clean/" + orgName + "/" + cluster + "/" + teamName + "/" + pipelineName + "/" + namespace
			req, _ := http.NewRequest("POST", url, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

			resp, err = client.Do(req)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}

			if resp.StatusCode == 200 {
				fmt.Println("Successfully deleted stale resources for pipeline in cluster", cluster, "and namespace", namespace)
			} else {
				errBody, _ := io.ReadAll(resp.Body)
				fmt.Printf("Error cleaning up pipeline: %s", errBody)
			}
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineCleanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pipelineCleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pipelineCleanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
