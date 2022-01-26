package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"

	// "strconv"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// clusterCreateCmd represents the clusterCreate command
var clusterReadCmd = &cobra.Command{
	Use:   "read <cluster name>",
	Short: "read a cluster",
	Long: `
Command to create a cluster. Specify the cluster name as the argument, and cluster ip and exposed port as flags.
 
Example usage:
	atlas cluster read cluster_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas cluster read -h' to see usage details")
			return
		}

		clusterName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/cluster/" + orgName + "/" + clusterName

		var req *http.Request

		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
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
			fmt.Printf("Error reading cluster: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterReadCmd)
}
