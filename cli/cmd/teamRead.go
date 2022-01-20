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

// teamReadCmd represents the teamRead command
var teamReadCmd = &cobra.Command{
	Use:   "read <team name>",
	Short: "Read a specified team's information",
	Long: `
Command to read a team's information. Specify the name of the team. 
			 
Example usage:
	atlas team read example_team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run atlas team read -h for usage details.")
			return
		}

		teamName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "https://" + atlasURL + "/team/" + orgName + "/" + teamName

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
		} else if statusCode == 400 {
			fmt.Println("Team cannot be read. Invalid org name or team name provided.")
		} else {
			fmt.Println(statusCode)
			fmt.Println("Internal server error, please try again.")
		}
	},
}

func init() {
	teamCmd.AddCommand(teamReadCmd)

}
