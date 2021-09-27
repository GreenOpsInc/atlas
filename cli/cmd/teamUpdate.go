package cmd

import (
	"fmt"
	"net/http"
	"github.com/spf13/cobra"
	"bytes"
	"encoding/json"
)

// teamUpdateCmd represents the teamUpdate command
var teamUpdateCmd = &cobra.Command{
	Use:   "update <team name> -p <new parent team name> -n <new team name>",
	Short: "Update the team details",
	Long: `

Command to update a team's information. Specify the original team name as
the argument, and the new parent team the team should be under (optional),
and/or the new team name using the flags.

Example usage:
	atlas team update team_name -p new_parent_team_name -n new_team_name`,
	
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run atlas team update -h for usage details.")
			return
		}

		teamName:=args[0]
		newParentTeamName,_:=cmd.Flags().GetString("parent")
		newTeamName,_:=cmd.Flags().GetString("new-name")

		url:= "http://"+atlasURL+"/team/"+orgName+"/"+teamName
		

		body := UpdateTeamRequest{
			TeamName: newTeamName,			
			ParentTeamName: newParentTeamName,
		}
		json, _:= json.Marshal(body)
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(json))
		req.Header.Set("Content-Type", "application/json")		
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:",err)
			return
		}

	    statusCode := resp.StatusCode
		if statusCode == 200{
			fmt.Println("Successfully updated team to",newTeamName, "under parent team:", newParentTeamName)
		} else if statusCode == 400{
			fmt.Println("Team update failed because the request was invalid. Please check if the provided arguments are correct.")			
		} else{
			fmt.Println("Internal server error, please try again.")
		}

	},

}

func init() {
	teamCmd.AddCommand(teamUpdateCmd)
	teamUpdateCmd.PersistentFlags().StringP("parent", "p", "na", "new parent team name")
	teamUpdateCmd.PersistentFlags().StringP("new-name", "n", "", "new team name")
	teamUpdateCmd.MarkPersistentFlagRequired("new-name")
}