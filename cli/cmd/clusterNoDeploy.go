package cmd

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"

	// "strconv"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	remove bool
)

// clusterMarkNoDeployCmd represents the clusterMarkNoDeploy command
var clusterMarkNoDeployCmd = &cobra.Command{
	Use:   "nodeploy <cluster name> --name <name or email> --reason <\"reason for marking as no deploy\"> --namespace <optional>",
	Short: "Create a cluster",
	Long: `
Command to mark a cluster as no deploy. Specify the cluster name as the argument, and the reason, name, namespace, and "unmarking" as flags.
 
Example usage:
	atlas cluster nodeploy cluster_name --name team@company.com --reason "shutting down entire cluster"
	atlas cluster nodeploy cluster_name --name team@company.com --reason "restricting namespace for testing" --namespace dev
	atlas cluster nodeploy cluster_name --remove --name team@company.com --reason "restricting namespace for testing" --namespace dev`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas cluster nodeploy -h' to see usage details")
			return
		}

		clusterName := args[0]
		name, _ := cmd.Flags().GetString("name")
		reason, _ := cmd.Flags().GetString("reason")
		namespace, _ := cmd.Flags().GetString("namespace")

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := fmt.Sprintf("https://%s/cluster/%s/%s/noDeploy", atlasURL, orgName, clusterName)

		var req *http.Request

		body := NoDeployInfo{
			Name:      name,
			Reason:    reason,
			Namespace: namespace,
		}

		json, _ := json.Marshal(body)
		if remove {
			url = url + "/remove"
		} else {
			url = url + "/apply"
		}
		req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := getHttpClient()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200 {
			fmt.Println("Successfully updated cluster nodeploy status")
		} else {
			errBody, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error updating cluster nodeploy status: %d - %s", statusCode, string(errBody))
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterMarkNoDeployCmd)
	clusterMarkNoDeployCmd.PersistentFlags().StringP("name", "", "", "user name or email")
	clusterMarkNoDeployCmd.PersistentFlags().StringP("reason", "", "", "reason for marking as nodeploy")
	clusterMarkNoDeployCmd.PersistentFlags().StringP("namespace", "", "", "namespace to be restricted (leave blank for entire cluster)")
	clusterMarkNoDeployCmd.MarkPersistentFlagRequired("name")
	clusterMarkNoDeployCmd.MarkPersistentFlagRequired("reason")
	clusterMarkNoDeployCmd.PersistentFlags().BoolVarP(&remove, "remove", "r", false, "remove bool")
}
