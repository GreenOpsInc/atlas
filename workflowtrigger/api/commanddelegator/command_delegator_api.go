package commanddelegator

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/greenopsinc/util/apikeys"

	"github.com/greenopsinc/util/clientrequest"
	"greenops.io/workflowtrigger/serializer"

	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/tlsmanager"
)

const (
	RootCommit       = "ROOT_COMMIT"
	PipelineFileName = "pipeline.yaml"
	getFileExtension = "file"
	apiKeySecretName = "atlas-workflow-trigger-api-key"
	apiKeyHeaderName = "api-key"
)

type CommandDelegatorApi interface {
	SendNotification(orgName string, clusterName string, request clientrequest.NotificationRequestEvent) string
}

type commandDelegatorApi struct {
	serverEndpoint string
	client         httpclient.HttpClient
	apikey         string
}

func New(serverEndpoint string, tm tlsmanager.Manager, apiKeysClient apikeys.Client) (CommandDelegatorApi, error) {
	if strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint = serverEndpoint + "notifications"
	} else {
		serverEndpoint = serverEndpoint + "/notifications"
	}
	httpClient, err := httpclient.New(tlsmanager.ClientCommandDelegator, tm)
	if err != nil {
		return nil, err
	}
	apikey, err := getApiKey(apiKeysClient)
	if err != nil {
		return nil, err
	}
	return &commandDelegatorApi{
		serverEndpoint: serverEndpoint,
		client:         httpClient,
		apikey:         apikey,
	}, nil
}

func (c *commandDelegatorApi) SendNotification(orgName string, clusterName string, clientRequest clientrequest.NotificationRequestEvent) string {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(clientRequest))
	request, err = http.NewRequest("POST", c.serverEndpoint+fmt.Sprintf("/%s/%s", orgName, clusterName), bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(apiKeyHeaderName, c.apikey)
	resp, err := c.client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	log.Printf("Send notification request returned status code %d", resp.StatusCode)
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		panic(err)
	} else if resp.StatusCode != 200 {
		panic(fmt.Sprintf("Error when making notification request to the command delegator: %s", buf.String()))
	}
	return buf.String()
}

func getApiKey(apiKeysClient apikeys.Client) (string, error) {
	apikey, err := apiKeysClient.Get(apiKeySecretName)
	if err != nil {
		return "", err
	}
	if apikey != "" {
		return apikey, nil
	}
	apikey, err = apiKeysClient.Issue(apiKeySecretName)
	if err != nil {
		return "", err
	}
	return apikey, nil
}
