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

// teamDeleteCmd represents the teamDelete command
var teamDeleteCmd = &cobra.Command{
	Use:   "delete <team name>",
	Short: "Delete a specified team",
	Long: `
Command to delete a team. Specify the name of the team to be deleted. 
		 
Example usage:
	atlas team delete team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run atlas team delete -h for usage details.")
			return
		}

		teamName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/team/" + orgName + "/" + teamName
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
			fmt.Println("Successfully deleted team:", teamName)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	teamCmd.AddCommand(teamDeleteCmd)
}
