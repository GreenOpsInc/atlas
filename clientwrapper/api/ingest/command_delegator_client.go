package ingest

import (
	"encoding/json"
	"fmt"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	EnvCommandDelegatorUrl     string = "COMMAND_DELEGATOR_URL"
	DefaultCommandDelegatorUrl string = "http://commanddelegator.default.svc.cluster.local:8080"
	EnvClusterName             string = "CLUSTER_NAME"
	DefaultClusterName         string = "kubernetes_local"
	EnvOrgName                 string = "ORG_NAME"
	DefaultOrgName             string = "org" //TODO: Remove this
)

const (
	EventTypeVariable string = "type"
)

type CommandDelegatorApi interface {
	GetCommands() (*[]requestdatatypes.RequestEvent, error)
	AckHeadOfRequestList() error
}

type CommandDelegatorImpl struct {
	commandDelegatorUrl string
	clusterName         string
	orgName             string
}

func Create() CommandDelegatorApi {
	commandDelegatorUrl := os.Getenv(EnvCommandDelegatorUrl)
	if commandDelegatorUrl == "" {
		commandDelegatorUrl = DefaultCommandDelegatorUrl
	}
	if strings.HasSuffix(commandDelegatorUrl, "/") {
		commandDelegatorUrl = strings.TrimSuffix(commandDelegatorUrl, "/")
	}
	clusterName := os.Getenv(EnvClusterName)
	if clusterName == "" {
		clusterName = DefaultClusterName
	}
	orgName := os.Getenv(EnvOrgName)
	if orgName == "" {
		orgName = DefaultOrgName
	}
	return CommandDelegatorImpl{commandDelegatorUrl: commandDelegatorUrl, clusterName: clusterName, orgName: orgName}
}

func (c CommandDelegatorImpl) GetCommands() (*[]requestdatatypes.RequestEvent, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", c.commandDelegatorUrl+fmt.Sprintf("/requests/%s/%s", c.orgName, c.clusterName), nil)
	if err != nil {
		log.Printf("Error while making getting commands request: %s", err)
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error while getting commands: %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading the response body: %s", err)
		return nil, err
	}
	var commands []requestdatatypes.RequestEvent
	var jsonArray []map[string]interface{}
	json.Unmarshal(bodyBytes, &jsonArray)
	for _, command := range jsonArray {
		commandType := command[EventTypeVariable].(string)

		if commandType == requestdatatypes.ClientDeployRequestType {
			var request requestdatatypes.ClientDeployRequest
			json.Unmarshal(c.getBytesFromUnstructured(command), &request)
			commands = append(commands, request)
		} else if commandType == requestdatatypes.ClientDeleteByConfigRequestType {
			var request requestdatatypes.ClientDeleteByConfigRequest
			json.Unmarshal(c.getBytesFromUnstructured(command), &request)
			commands = append(commands, request)
		} else if commandType == requestdatatypes.ClientDeleteByGvkRequestType {
			var request requestdatatypes.ClientDeleteByGvkRequest
			json.Unmarshal(c.getBytesFromUnstructured(command), &request)
			commands = append(commands, request)
		} else if commandType == requestdatatypes.ClientDeployAndWatchRequestType {
			var request requestdatatypes.ClientDeployAndWatchRequest
			json.Unmarshal(c.getBytesFromUnstructured(command), &request)
			commands = append(commands, request)
		} else if commandType == requestdatatypes.ClientRollbackAndWatchRequestType {
			var request requestdatatypes.ClientRollbackAndWatchRequest
			json.Unmarshal(c.getBytesFromUnstructured(command), &request)
			commands = append(commands, request)
		} else if commandType == requestdatatypes.ClientSelectiveSyncRequestType {
			var request requestdatatypes.ClientSelectiveSyncRequest
			json.Unmarshal(c.getBytesFromUnstructured(command), &request)
			commands = append(commands, request)
		} else {
			log.Printf("Command type %s not supported", commandType)
		}
	}

	return &commands, nil
}

func (c CommandDelegatorImpl) AckHeadOfRequestList() error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", c.commandDelegatorUrl+fmt.Sprintf("/requests/ackHead/%s/%s", c.orgName, c.clusterName), nil)
	if err != nil {
		log.Printf("Error while making ack head request: %s", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error while acking head: %s", err)
		return err
	}
	defer resp.Body.Close()
	//TODO: Check api statuses
	return nil
}

func (c CommandDelegatorImpl) getBytesFromUnstructured(jsonPayload map[string]interface{}) []byte {
	jsonString, _ := json.Marshal(jsonPayload)
	return jsonString
}