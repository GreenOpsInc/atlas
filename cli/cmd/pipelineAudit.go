package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/greenopsinc/util/clientrequest"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strings"
)

// pipelineAuditCmd represents the pipelineAudit command
var pipelineAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "generate an audit for a specified pipeline ",
	Long: `Command to generate a pipeline audit. For each cluster-namespace combination, specify stale resources to drop. 

Example usage:
	atlas pipeline audit pipeline_name --team team_name`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run 'atlas pipeline audit -h' to see usage details")
			return
		}

		pipelineName := args[0]
		teamName, _ := cmd.Flags().GetString("team")

		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		groupUrl := "https://" + atlasURL + "/combinations/" + orgName + "/" + teamName + "/" + pipelineName

		req, _ := http.NewRequest("GET", groupUrl, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

		client := getHttpClient()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		var groups ClusterNamespaceGroups
		err = json.Unmarshal(body,
			&groups)
		if err != nil {
			fmt.Println("Unmarshall failed with the following error:", err)
			return
		}
		for _, group := range groups.Groups {
			cluster := group.ClusterName
			namespace := group.Namespace
			url := "https://" + atlasURL + "/aggregate/" + orgName + "/" + cluster + "/" + namespace
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}
			var aggregateResult AggregateResult
			aggregateMap := make(map[string]int)
			var resourceList []clientrequest.GvkResourceInfo

			err = json.Unmarshal(body, &aggregateResult)
			resourceNames := []string{}
			for index, resource := range aggregateResult.ResourceList {
				aggregateMap[resource.Name] = index
				resourceNames = append(resourceNames, resource.Name+" "+resource.Kind+" "+resource.Version)

				if resource.Kind == "AtlasGroup" {
					for _, subResource := range resource.ResourceList {
						aggregateMap[subResource.Name] = index
						resourceNames = append(resourceNames, subResource.Name+" "+subResource.Kind+" "+subResource.Version+" (PART OF ATLAS GROUP: "+resource.Name+")")
					}
				}
			}
			fmt.Printf("Resources for cluster: %s and namespace: %s\n", cluster, namespace)
			var selected []string
			prompt := &survey.MultiSelect{
				Message: "Select displayed resources to mark as stale (selecting an Atlas Group will mark all underlying infrastructure as stale)\n (Name Kind ApiVersion shown for each resource)",
				Options: resourceNames,
			}

			survey.AskOne(prompt, &selected)
			count := 0
			markedAtlasGroup := make(map[int]bool)
			for _, resourceStr := range selected {
				split := strings.Split(resourceStr, " ")
				name := split[0]
				kind := split[1]
				idx := aggregateMap[name]
				if _, ok := markedAtlasGroup[idx]; ok {
					continue
				}
				resource := aggregateResult.ResourceList[idx]
				if resource.Kind == "AtlasGroup" {
					if kind == "AtlasGroup" {
						markedAtlasGroup[idx] = true
						for _, subResource := range resource.ResourceList {
							count++
							gvk := schema.GroupVersionKind{Version: subResource.Version, Kind: subResource.Kind}
							resourceList = append(resourceList, clientrequest.GvkResourceInfo{GroupVersionKind: gvk, ResourceName: subResource.Name, ResourceNamespace: namespace})
						}
					} else {
						for _, subResource := range resource.ResourceList {
							if subResource.Name == name && subResource.Kind == kind {
								count++
								gvk := schema.GroupVersionKind{Version: subResource.Version, Kind: subResource.Kind}
								resourceList = append(resourceList, clientrequest.GvkResourceInfo{GroupVersionKind: gvk, ResourceName: subResource.Name, ResourceNamespace: namespace})
							}
						}
					}
				} else {
					count++
					gvk := schema.GroupVersionKind{Version: resource.Version, Kind: resource.Kind}
					resourceList = append(resourceList, clientrequest.GvkResourceInfo{GroupVersionKind: gvk, ResourceName: resource.Name, ResourceNamespace: namespace})
				}
			}

			if count == 0 {
				fmt.Printf("No resource(s) to audit in cluster %s and namespace %s\n", cluster, namespace)
				continue
			}

			url = "https://" + atlasURL + "/label/" + orgName + "/" + cluster + "/" + teamName + "/" + pipelineName
			reqBody := clientrequest.GvkGroupRequest{ResourceList: resourceList}
			json, _ := json.Marshal(reqBody)
			req, _ = http.NewRequest("POST", url, bytes.NewBuffer(json))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))

			resp, err = client.Do(req)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Request failed with the following error:", err)
				return
			}

			if resp.StatusCode == 200 {
				fmt.Printf("Marked %d resource(s) as stale in cluster %s and namespace %s\n", count, cluster, namespace)
			} else {
				errBody, _ := io.ReadAll(resp.Body)
				fmt.Printf("Error labeling resources: %s", errBody)
			}

		}

	},
}

func init() {
	pipelineCmd.AddCommand(pipelineAuditCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pipelineAuditCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pipelineAuditCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
