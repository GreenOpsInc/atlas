package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"time"
)

// pipelineReadCmd represents the pipelineRead command
var pipelineReadCmd = &cobra.Command{
	Use:   "read <pipeline name",
	Short: "Read a specified pipeline's details",
	Long: `
Command to read a pipeline. Specify the names of the pipeline as the argument, and the team name using the flag. 

Example usage:
	atlas pipeline read pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline read -h' to see usage details")
			return
		}

		teamName,_:=cmd.Flags().GetString("team")
		pipelineName:= args[0]

		
		url:= "http://"+atlasURL+"/pipeline/"+orgName+"/"+teamName+"/"+pipelineName
		
		req, err:= http.NewRequest("GET", url, nil)
		
		client := &http.Client{Timeout: 20 * time.Second}
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
			fmt.Println("Pipeline cannot be read. Invalid org name, team name, or pipeline name provided.")			
		} else{
			fmt.Println("Internal server error, please try again.")			
		}
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineReadCmd)
}
