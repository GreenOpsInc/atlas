package cmd

import (
	"fmt"
	"net/http"
	"github.com/spf13/cobra"
)

// pipelineDeleteCmd represents the pipelineDelete command
var pipelineDeleteCmd = &cobra.Command{
	Use:   "delete <pipeline name>",
	Short: "Delete a specified pipeline",
	Long: `
Command to delete a pipeline. Specify the name of the pipeline as the argument, and team name with the flag. 

Example usage:
	atlas pipeline delete pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline delete -h' to see usage details")
			return
		}

		teamName,_:=cmd.Flags().GetString("team")		
		pipelineName:= args[0]
		
		url:= "http://"+atlasURL+"/pipeline/"+orgName+"/"+teamName+"/"+pipelineName
		req, err:= http.NewRequest("DELETE", url, nil)
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:",err)
			return
		}

		statusCode := resp.StatusCode
		if statusCode == 200{
			fmt.Println("Successfully deleted pipeline:",pipelineName, "for team:", teamName)
		} else if statusCode == 400{
			fmt.Println("Pipeline deletion command failed. Invalid org name, team name, or pipeline name provided.")			
		} else{
			fmt.Println("Internal server error, please try again.")			
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineDeleteCmd)
}
