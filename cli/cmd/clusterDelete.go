package cmd

import (
	"fmt"
	"net/http"
	"github.com/spf13/cobra"
)

// clusterDeleteCmd represents the clusterDelete command
var clusterDeleteCmd = &cobra.Command{
	Use:   "delete <cluster name>",
	Short: "Delete a specified cluster",
	Long: `
Command to delete a cluster. Specify the name of the cluster to be deleted. 
		 
Example usage:
	atlas cluster delete cluster_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run atlas cluster delete -h for usage details.")
			return
		}

		clusterName:=args[0]
		

		url:= "http://"+atlasURL+"/cluster/"+orgName+"/"+clusterName
		req, err:= http.NewRequest("DELETE", url, nil)
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:",err)
			return
		}

		statusCode := resp.StatusCode
		if statusCode == 200{
			fmt.Println("Successfully deleted cluster:",clusterName)
		} else if statusCode == 400{
			fmt.Println("Cluster deletion failed. Invalid org name or cluster name provided.")			
		} else{
			fmt.Println("Internal server error, please try again.")			
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterDeleteCmd)
}
