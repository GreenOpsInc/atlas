package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

var (
	pipelineRevisionHash string
	argoRevisionHash     string
)

// pipelineForceDeployCmd represents the pipelineForceDeploy command
var pipelineForceDeployCmd = &cobra.Command{
	Use:   "force-deploy <pipeline name> --step <step name> --pipelineRevisionHash <pipeline repo revision> --argoRevisionHash <argo manifest revision> --team <team name> --repo <git repo url> --root <path to root>",
	Short: "Force deploy a pipeline",
	Long: `
Command to force the deployment of an application in a step. Specify the name, step, and revision of the application to be deployed. 
Set the team name, git repo url, and path to root with the required flags. 
Optional flags should be set based on the type of Git access credential (open, oauth token, or username and password).

Example usage:
	atlas pipeline force-deploy pipeline_name --step step_name --team team_name --repo git_repo --root path_to_root (No flags specified means open access)
	atlas pipeline force-deploy pipeline_name --step step_name --team team_name --repo git_repo --root path_to_root -t token
	atlas pipeline force-deploy pipeline_name --step step_name --team team_name --repo git_repo --root path_to_root -u username -p password`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline force-deploy -h' to see usage details")
			return
		}

		tokenFlagSet := cmd.Flags().Lookup("token").Changed
		usernameFlagSet := cmd.Flags().Lookup("username").Changed
		passwordFlagSet := cmd.Flags().Lookup("password").Changed

		if tokenFlagSet && (usernameFlagSet || passwordFlagSet) {
			fmt.Println("Invalid combination of flags. Run 'atlas pipeline force-deploy -h' to see usage details")
			return
		}

		if (usernameFlagSet && !passwordFlagSet) || (!usernameFlagSet && passwordFlagSet) {
			fmt.Println("Username must be passed in with a password. Run 'atlas pipeline sync -h' to see usage details")
			return
		}

		teamName, _ := cmd.Flags().GetString("team")
		stepName, _ := cmd.Flags().GetString("step")
		gitRepo, _ := cmd.Flags().GetString("repo")
		pathToRoot, _ := cmd.Flags().GetString("root")
		pipelineName := args[0]

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		url := fmt.Sprintf("https://%s/force/%s/%s/%s/%s/%s/%s", atlasURL, orgName, teamName, pipelineName, pipelineRevisionHash, stepName, argoRevisionHash)

		var req *http.Request

		if !tokenFlagSet && !usernameFlagSet {
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
			fmt.Println("Successfully force-deployed step:", stepName, "in pipeline:", pipelineName)
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("An error occurred: %s", body)
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineForceDeployCmd)

	pipelineForceDeployCmd.PersistentFlags().StringP("repo", "", "", "git repo url")
	pipelineForceDeployCmd.PersistentFlags().StringP("root", "", "", "path to root")
	pipelineForceDeployCmd.PersistentFlags().StringP("step", "", "", "step name")
	pipelineForceDeployCmd.MarkPersistentFlagRequired("repo")
	pipelineForceDeployCmd.MarkPersistentFlagRequired("root")
	pipelineForceDeployCmd.MarkPersistentFlagRequired("step")

	pipelineForceDeployCmd.PersistentFlags().StringP("token", "t", "", "Name of git cred token")
	pipelineForceDeployCmd.PersistentFlags().StringP("username", "u", "", "Github username")
	pipelineForceDeployCmd.PersistentFlags().StringP("password", "p", "", "Github password")

	pipelineForceDeployCmd.PersistentFlags().StringVarP(&pipelineRevisionHash, "pipelineRevisionHash", "", "ROOT_COMMIT", "set pipeline repo revision")
	pipelineForceDeployCmd.PersistentFlags().StringVarP(&argoRevisionHash, "argoRevisionHash", "", "LATEST_REVISION", "set Argo manifest revision")
}
