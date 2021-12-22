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

// clusterDeleteCmd represents the clusterDelete command
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete <cluster name>",
	Short: "Delete a specified cluster",
	Long: `
Command to delete a cluster. Specify the name of the cluster to be deleted. 
		 
Example usage:
	atlas cluster delete cluster_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run atlas cluster delete -h for usage details.")
			return
		}

		clusterName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/cluster/" + orgName + "/" + clusterName
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
			fmt.Println("Successfully deleted cluster:", clusterName)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterDeleteCmd)
}
