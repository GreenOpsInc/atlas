package cmd

import (
	"bytes"
	"encoding/json"
	// "strconv"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

// clusterCreateCmd represents the clusterCreate command
var clusterCreateCmd = &cobra.Command{
	Use:   "create <cluster name> --ip <cluster IP> --port <exposed port>",
	Short: "Create a cluster",
	Long: `
Command to create a cluster. Specify the cluster name as the argument, and cluster ip and exposed port as flags.
 
Example usage:
	atlas cluster create cluster_name --ip 192.0.2.42 --port 9376`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
			fmt.Println("Invalid number of arguments. Run 'atlas cluster create -h' to see usage details")
			return
		}

		clusterName:=args[0]
		clusterIP,_:=cmd.Flags().GetString("ip")
		exposedPort,_:=cmd.Flags().GetInt("port")
		
		
		url:= "http://"+atlasURL+"/cluster/"+orgName

		var req *http.Request		
		
		body := ClusterSchema {
			ClusterIP: clusterIP,
			ExposedPort: exposedPort,
			ClusterName: clusterName,
		}
		
		json, _:= json.Marshal(body)
		req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))

		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:",err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200{
			fmt.Println("Successfully created cluster:",clusterName, "for org:", orgName)
		} else{
			fmt.Println("Error creating cluster")			
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCreateCmd.PersistentFlags().StringP("ip", "", "", "cluster IP address")	
	clusterCreateCmd.PersistentFlags().IntP("port", "", 80, "exposed port")
	clusterCreateCmd.MarkPersistentFlagRequired("ip")
	clusterCreateCmd.MarkPersistentFlagRequired("port")
}
