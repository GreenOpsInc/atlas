package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/cluster"
	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/serializer"
	"github.com/greenopsinc/util/serializerutil"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/client/argodriver"
)

const (
	EnvCommandDelegatorUrl     string = "COMMAND_DELEGATOR_URL"
	DefaultCommandDelegatorUrl string = "http://commanddelegator.atlas.svc.cluster.local:8080"
	EnvClusterName             string = "CLUSTER_NAME"
	EnvOrgName                 string = "ORG_NAME"
	DefaultOrgName             string = "org" //TODO: Remove this
)

const (
	EventTypeVariable string = "type"
)

type CommandDelegatorApi interface {
	GetCommands() (*[]clientrequest.ClientRequestEvent, error)
	AckHeadOfRequestList() error
	AckHeadOfNotificationList() error
	RetryRequest() error
}

type CommandDelegatorImpl struct {
	commandDelegatorUrl string
	clusterName         string
	orgName             string
	argoAuthClient      argodriver.ArgoAuthClient
	client              httpclient.HttpClient
}

func Create(argoClient argodriver.ArgoAuthClient, tm tlsmanager.Manager) (CommandDelegatorApi, error) {
	commandDelegatorUrl := os.Getenv(EnvCommandDelegatorUrl)
	if commandDelegatorUrl == "" {
		commandDelegatorUrl = DefaultCommandDelegatorUrl
	}
	if strings.HasSuffix(commandDelegatorUrl, "/") {
		commandDelegatorUrl = strings.TrimSuffix(commandDelegatorUrl, "/")
	}
	clusterName := os.Getenv(EnvClusterName)
	if clusterName == "" {
		clusterName = cluster.LocalClusterName
	}
	orgName := os.Getenv(EnvOrgName)
	if orgName == "" {
		orgName = DefaultOrgName
	}
	httpClient, err := httpclient.New(tlsmanager.ClientCommandDelegator, tm)
	if err != nil {
		return nil, err
	}
	return CommandDelegatorImpl{
		commandDelegatorUrl: commandDelegatorUrl,
		clusterName:         clusterName,
		orgName:             orgName,
		argoAuthClient:      argoClient,
		client:              httpClient,
	}, nil
}

func (c CommandDelegatorImpl) GetCommands() (*[]clientrequest.ClientRequestEvent, error) {
	req, err := http.NewRequest("GET", c.commandDelegatorUrl+fmt.Sprintf("/requests/%s/%s", c.orgName, c.clusterName), nil)
	if err != nil {
		log.Printf("Error while making getting commands request: %s", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(req)
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
	if resp.StatusCode != 200 {
		log.Printf("Error while getting commands: %s", string(bodyBytes))
		return nil, errors.New(string(bodyBytes))
	}
	commands := serializer.Deserialize(string(bodyBytes), serializerutil.ClientRequestListType).([]clientrequest.ClientRequestEvent)
	return &commands, nil
}

func (c CommandDelegatorImpl) AckHeadOfRequestList() error {
	req, err := http.NewRequest("DELETE", c.commandDelegatorUrl+fmt.Sprintf("/requests/ackHead/%s/%s", c.orgName, c.clusterName), nil)
	if err != nil {
		log.Printf("Error while making ack head request: %s", err)
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error while acking head: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return errors.New(resp.Status)
	}
	return nil
}

func (c CommandDelegatorImpl) AckHeadOfNotificationList() error {
	req, err := http.NewRequest("DELETE", c.commandDelegatorUrl+fmt.Sprintf("/notifications/ackHead/%s/%s", c.orgName, c.clusterName), nil)
	if err != nil {
		log.Printf("Error while making ack head request: %s", err)
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error while acking notifications head: %s", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return errors.New(resp.Status)
	}
	return nil
}

func (c CommandDelegatorImpl) RetryRequest() error {
	req, err := http.NewRequest("DELETE", c.commandDelegatorUrl+fmt.Sprintf("/requests/retry/%s/%s", c.orgName, c.clusterName), nil)
	if err != nil {
		log.Printf("Error while making retry request: %s", err)
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error while sending retry message: %s", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return errors.New(resp.Status)
	}
	return nil
}

func (c CommandDelegatorImpl) getBytesFromUnstructured(jsonPayload map[string]interface{}) []byte {
	jsonString, _ := json.Marshal(jsonPayload)
	return jsonString
}
