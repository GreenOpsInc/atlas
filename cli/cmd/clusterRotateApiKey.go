package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	// "strconv"
	"fmt"
	"net/http"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

type ApiKeyType string

const (
	ApiKeyTypeWorkflowTrigger ApiKeyType = "workflow-trigger"
	ApiKeyTypeClientWrapper   ApiKeyType = "client-wrapper"
)

type RotateClusterApiKeyRequest struct {
	ApiKeyType string
}

// clusterCreateCmd represents the clusterCreate command
var clusterRotateApiKeyCmd = &cobra.Command{
	Use:   "rotate <cluster name>",
	Short: "rotate cluster api key",
	Long: `
Command to rotate cluster's api key. Specify the cluster name and apikey type as arguments.
Allowed apikey types are: 'workflow-trigger' and 'client-wrapper'.

Example usage:
	atlas cluster rotate cluster_name workflow-trigger`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("Invalid number of arguments. Run 'atlas cluster read -h' to see usage details")
			return
		}

		clusterName := args[0]
		apiKeyType := args[1]
		if apiKeyType != string(ApiKeyTypeWorkflowTrigger) && apiKeyType != string(ApiKeyTypeClientWrapper) {
			fmt.Println("Invalid apikey type. Run 'atlas cluster rotate -h' to see usage details")
		}

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/cluster/" + orgName + "/" + clusterName + "/apikeys/rotate"
		reqData := &RotateClusterApiKeyRequest{ApiKeyType: apiKeyType}
		reqBytes, err := json.Marshal(reqData)
		if err != nil {
			fmt.Printf("Failed to rotate apikey: %s", err.Error())
		}
		req, _ := http.NewRequest("GET", url, bytes.NewBuffer(reqBytes))
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
			var resData CreateClusterResponse
			if err = json.Unmarshal(body, &resData); err != nil {
				fmt.Printf("Error rotating cluster apikey: %s", err.Error())
			}
			fmt.Printf("Successfully rotated apikey of type %s for cluster %s of org %s. New apikey is '%s'", apiKeyType, clusterName, orgName, resData.ApiKey)
		} else {
			fmt.Printf("Error rotating cluster apikey: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterRotateApiKeyCmd)
}
