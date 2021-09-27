package cmd

import (
	"fmt"
	"os"
	"io/ioutil"
	"net/http"
	"bytes"
	"github.com/spf13/cobra"
)

// teamCreateCmd represents the teamCreate command
var teamCreateCmd = &cobra.Command{
	Use:   "create  <team name> optional: -p <parent team name> -s <path to pipeline schemas>",
	Short: "Command to create a team.",
	Long: `
Command to create a team. Specify the name of the team to be created.
The optional -p flag is used to set the parent team name, and is 'na' by default. The
filename of a JSON file with defined pipeline schemas is also optional and set with the -s flag. 
If provided, the created team will automatically have these pipelines defined.
	 
Example usage:
	atlas team create team_name (team will be created under 'na' parent team by default)
	atlas team create team_name -p parent_team
	atlas team create team_name -p parent_team -s pipeline_schemas.json`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1{
			fmt.Println("Invalid number of arguments. Run atlas team create -h to see usage details.")
			return
		}

		parentTeamName,_:=cmd.Flags().GetString("parent")
		teamName:= args[0]

		url:= "http://"+atlasURL+"/team/"+orgName+"/"+parentTeamName+"/"+teamName
		
		var req *http.Request
		var er error
				
		if cmd.Flags().Lookup("schemas").Changed{
			jsonFile, err := os.Open(args[3])
			if err != nil {
				fmt.Println("Unable to find or process pipeline schemas file")
			}
			defer jsonFile.Close()
		
			byteValue, _ := ioutil.ReadAll(jsonFile)
			req, er = http.NewRequest("POST", url, bytes.NewReader(byteValue))
			req.Header.Set("Content-Type", "application/json")
			if er != nil{
				fmt.Println("Request failed, please try again.")
			}
		} else{
			req, er = http.NewRequest("POST", url, nil)
			if er != nil{
				fmt.Println("Request failed, please try again.")
			}
		}

		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200{
			fmt.Println("Successfully created team:",teamName, "under parent team:", parentTeamName)
		} else if statusCode == 400{
			fmt.Println("Team creation failed because the request was invalid.\nPlease check if org and parent team names are correct, a team with the specified name doesn't already exist, and the format of the schema file (if provided) is valid.")			
		} else{
			fmt.Println("Internal server error: ",err)			
		}
	},
}

func init() {
	teamCmd.AddCommand(teamCreateCmd)
	teamCreateCmd.PersistentFlags().StringP("parent", "p", "na", "parent team name")
	teamCreateCmd.PersistentFlags().StringP("schemas", "s", "", "path to pipeline schemas JSON file")
	
}
