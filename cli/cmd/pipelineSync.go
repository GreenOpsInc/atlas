package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"encoding/json"
	"bytes"
)

// pipelineSyncCmd represents the pipelineSync command
var pipelineSyncCmd = &cobra.Command{
	Use:   "sync <pipeline name> --team <team name> --repo <git repo url> --root <path to root>",
	Short: "Sync a pipeline",
	Long: `
Command to sync a pipeline. Specify the name of the pipeline to be synced. 
Set the team name, git repo url, and path to root with the required flags. 
Optional flags should be set based on the type of Git access credential (open, oauth token, or username and password).

Example usage:
	atlas pipeline sync pipeline_name --team team_name  --repo git_repo --root path_to_root (No flags specified means open access)
	atlas pipeline sync pipeline_name --team team_name  --repo git_repo --root path_to_root -t token
	atlas pipeline sync pipeline_name --team team_name  --repo git_repo --root path_to_root -u username -p password`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline sync -h' to see usage details")
			return
		}

		
		tokenFlagSet:=cmd.Flags().Lookup("token").Changed
		usernameFlagSet:=cmd.Flags().Lookup("username").Changed
		passwordFlagSet:=cmd.Flags().Lookup("password").Changed

		if tokenFlagSet && (usernameFlagSet || passwordFlagSet){
			fmt.Println("Invalid combination of flags. Run 'atlas pipeline sync -h' to see usage details")
			return
		}

		if (usernameFlagSet && !passwordFlagSet) || (!usernameFlagSet && passwordFlagSet){
			fmt.Println("Username must be passed in with a password. Run 'atlas pipeline sync -h' to see usage details")
			return
		}

		teamName,_:=cmd.Flags().GetString("team")
		gitRepo,_:=cmd.Flags().GetString("repo")
		pathToRoot,_:=cmd.Flags().GetString("root")
		pipelineName:= args[0]
		


		url:= "http://"+atlasURL+"/sync/"+orgName+"/"+teamName+"/"+pipelineName

		var req *http.Request		
		
		if (!tokenFlagSet && !usernameFlagSet){
			body := GitRepoSchemaOpen{
				GitRepo: gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredOpen{
					Type: "open",
				},
			}
			json, _:= json.Marshal(body)
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
			

		} else if (tokenFlagSet){
			token, _ := cmd.Flags().GetString("token")
			
			body := GitRepoSchemaToken{
				GitRepo: gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredToken{
					Type: "oauth",
					Token: token,
				},
			}
			json, _:= json.Marshal(body)
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
		} else{
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			body := GitRepoSchemaMachineUser{
				GitRepo: gitRepo,
				PathToRoot: pathToRoot,
				GitCred: GitCredMachineUser{
					Type: "machineuser",
					Username: username,
					Password: password,
				},
			}
			
			json, _:= json.Marshal(body)
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
		}

	
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:",err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200{
			fmt.Println("Successfully synced pipeline:",pipelineName, "for team:", teamName)
		} else if statusCode == 400{
			fmt.Println("Pipeline sync failed because the request was invalid.\nPlease check if the team and org names are correct, a pipeline with the specified name exists, and the Git credentials are valid.")			
		} else{
			fmt.Println("Internal server error, please try again and confirm that the provided Git credentials are correct")			
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineSyncCmd)

	pipelineSyncCmd.PersistentFlags().StringP("repo", "", "", "git repo url")	
	pipelineSyncCmd.PersistentFlags().StringP("root", "", "", "path to root")
	pipelineSyncCmd.MarkPersistentFlagRequired("repo")
	pipelineSyncCmd.MarkPersistentFlagRequired("root")
	
	pipelineSyncCmd.PersistentFlags().StringP("token", "t", "", "Name of git cred token")
	pipelineSyncCmd.PersistentFlags().StringP("username", "u", "", "Github username")
	pipelineSyncCmd.PersistentFlags().StringP("password", "p", "", "Github password")
}
