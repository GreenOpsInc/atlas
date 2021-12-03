package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

// pipelineCreateCmd represents the pipelineCreate command
var pipelineCreateCmd = &cobra.Command{
	Use:   "create <pipeline name>",
	Short: "Create a pipeline",
	Long: `
Command to create a pipeline. Declare the name of the pipeline to be created as the argument.
Specify the team name, git repo url, and the repo path to root using flags. (These three flags are required)
Additonal flags should be set based on the type of Git access credential (open, oauth token, or username and password)
 
Example usage:
	atlas pipeline create pipeline_name --team team_name --repo git_repo --root path_to_root (No extra flags specified means open access)
 	atlas pipeline create pipeline_name --team team_name --repo git_repo --root path_to_root -t token
 	atlas pipeline create pipeline_name --team team_name --repo git_repo --root path_to_root -u username -p password`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline create -h' to see usage details")
			return
		}

		tokenFlagSet := cmd.Flags().Lookup("token").Changed
		usernameFlagSet := cmd.Flags().Lookup("username").Changed
		passwordFlagSet := cmd.Flags().Lookup("password").Changed

		if tokenFlagSet && (usernameFlagSet || passwordFlagSet) {
			fmt.Println("Invalid combination of flags. Run 'atlas pipeline create -h' to see usage details")
			return
		}

		if (usernameFlagSet && !passwordFlagSet) || (!usernameFlagSet && passwordFlagSet) {
			fmt.Println("Username must be passed in with a password. Run 'atlas pipeline create -h' to see usage details")
			return
		}

		pipelineName := args[0]
		teamName, _ := cmd.Flags().GetString("team")
		gitRepo, _ := cmd.Flags().GetString("repo")
		pathToRoot, _ := cmd.Flags().GetString("root")

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := "http://" + atlasURL + "/pipeline/" + orgName + "/" + teamName + "/" + pipelineName

		var req *http.Request

		if !(tokenFlagSet || usernameFlagSet) {
			body := GitRepoSchemaOpen{
				GitRepo:    gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredOpen{
					Type: "open",
				},
			}
			json, _ := json.Marshal(body)
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))

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
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
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
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200 {
			fmt.Println("Successfully created pipeline:", pipelineName, "for team:", teamName)
		} else if statusCode == 400 {
			fmt.Println("Pipeline creation failed because the request was invalid.\nPlease check if the team and org names are correct, a pipeline with the specified name doesn't already exist, and the Git credentials are valid.")
		} else if statusCode == 409 {
			fmt.Println("Pipeline named", pipelineName, "already exists for team:", teamName)
		} else {
			fmt.Println("Internal server error, please try again and confirm that the provided values are correct.")
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineCreateCmd)

	pipelineCreateCmd.PersistentFlags().StringP("repo", "", "", "git repo url")
	pipelineCreateCmd.PersistentFlags().StringP("root", "", "", "path to root")
	pipelineCreateCmd.MarkPersistentFlagRequired("repo")
	pipelineCreateCmd.MarkPersistentFlagRequired("root")

	pipelineCreateCmd.PersistentFlags().StringP("token", "t", "", "name of git cred token")
	pipelineCreateCmd.PersistentFlags().StringP("username", "u", "", "git username")
	pipelineCreateCmd.PersistentFlags().StringP("password", "p", "", "git password")

}
