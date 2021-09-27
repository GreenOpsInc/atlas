package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/spf13/cobra"
	"bytes"
	"encoding/json"
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
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run atlas team read -h for usage details.")
			return
		}

		teamName:=args[0]

		url:= "http://"+atlasURL+"/team/"+orgName+"/"+teamName
		
		req, err:= http.NewRequest("GET", url, nil)
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:",err)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)

	
		statusCode := resp.StatusCode
		if statusCode == 200{
			var prettyJSON bytes.Buffer		
			error := json.Indent(&prettyJSON, body, "", "\t")
			if error != nil {
				fmt.Println("Request failed, please try again.")
				return
			}
		
			fmt.Println(string(prettyJSON.Bytes()))			
		} else if statusCode == 400{
			fmt.Println("Team cannot be read. Invalid org name or team name provided.")			
		} else{
			fmt.Println(statusCode)
			fmt.Println("Internal server error, please try again.")			
		}
	},
}

func init() {
	teamCmd.AddCommand(teamReadCmd)

}