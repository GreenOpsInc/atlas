package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

// pipelineUpdateCmd represents the pipelineUpdate command
var pipelineUpdateCmd = &cobra.Command{
	Use:   "update <pipeline name> --team <team name> --repo <git repo url> --root <path to root>",
	Short: "Update the details for a specified pipeline",
	Long: `
Command to update a pipeline to a different upstream repo. Specify the name of the pipeline to be updated. 
Set the team name, git repo url, and path to root with the required flags. 
Optional flags should be set based on the type of Git access credential (open, oauth token, or username and password).
	
Example usage:
	atlas pipeline update org_name team_name pipeline_name git_repo path_to_root (No flags specified means open access)
	atlas pipeline update org_name team_name pipeline_name git_repo path_to_root -t token
	atlas pipeline update org_name team_name pipeline_name git_repo path_to_root -u username -p password`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline update -h' to see usage details")
			return
		}

		tokenFlagSet := cmd.Flags().Lookup("token").Changed
		usernameFlagSet := cmd.Flags().Lookup("username").Changed
		passwordFlagSet := cmd.Flags().Lookup("password").Changed

		if tokenFlagSet && (usernameFlagSet || passwordFlagSet) {
			fmt.Println("Invalid combination of flags. Run 'atlas pipeline update -h' to see usage details")
			return
		}

		if (usernameFlagSet && !passwordFlagSet) || (!usernameFlagSet && passwordFlagSet) {
			fmt.Println("Username must be passed in with a password. Run 'atlas pipeline update -h' to see usage details")
			return
		}

		teamName, _ := cmd.Flags().GetString("team")
		gitRepo, _ := cmd.Flags().GetString("repo")
		pathToRoot, _ := cmd.Flags().GetString("root")
		pipelineName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		var req *http.Request
		client := getHttpClient()
		url := "https://" + atlasURL + "/pipeline/" + orgName + "/" + teamName + "/" + pipelineName

		if !tokenFlagSet && !usernameFlagSet {
			body := GitRepoSchemaOpen{
				GitRepo:    gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredOpen{
					Type: "open",
				},
			}
			json, _ := json.Marshal(body)
			req, _ = http.NewRequest("PUT", url, bytes.NewBuffer(json))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))
		} else if tokenFlagSet {
			token, _ := cmd.Flags().GetString("token")

			body := GitRepoSchemaToken{
				GitRepo:    gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredToken{
					Type:  "oauth",
					Token: token,
				},
			}
			json, _ := json.Marshal(body)
			req, _ = http.NewRequest("PUT", url, bytes.NewBuffer(json))
		} else {
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			body := GitRepoSchemaMachineUser{
				GitRepo:    gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredMachineUser{
					Type:     "machineuser",
					Username: username,
					Password: password,
				},
			}

			json, _ := json.Marshal(body)
			req, _ = http.NewRequest("PUT", url, bytes.NewBuffer(json))
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200 {
			fmt.Println("Successfully updated pipeline:", pipelineName, "for team:", teamName)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error: %d - %s", statusCode, string(body))
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineUpdateCmd)

	pipelineUpdateCmd.PersistentFlags().StringP("repo", "", "", "git repo url")
	pipelineUpdateCmd.PersistentFlags().StringP("root", "", "", "path to root")
	pipelineUpdateCmd.MarkPersistentFlagRequired("repo")
	pipelineUpdateCmd.MarkPersistentFlagRequired("root")

	pipelineUpdateCmd.PersistentFlags().StringP("token", "t", "", "Name of git cred token")
	pipelineUpdateCmd.PersistentFlags().StringP("username", "u", "", "Github username")
	pipelineUpdateCmd.PersistentFlags().StringP("password", "p", "", "Github password")
}
