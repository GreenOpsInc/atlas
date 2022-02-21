package commanddelegator

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/greenopsinc/util/clientrequest"
	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/tlsmanager"
	"greenops.io/workflowtrigger/apikeysmanager"
	"greenops.io/workflowtrigger/serializer"
)

const (
	RootCommit       = "ROOT_COMMIT"
	PipelineFileName = "pipeline.yaml"
	getFileExtension = "file"
)

type CommandDelegatorApi interface {
	SendNotification(orgName string, clusterName string, request clientrequest.NotificationRequestEvent) string
}

type commandDelegatorApi struct {
	serverEndpoint string
	client         httpclient.HttpClient
	apiKeysManager apikeysmanager.Manager
}

func New(serverEndpoint string, tm tlsmanager.Manager, apiKeysManager apikeysmanager.Manager) (CommandDelegatorApi, error) {
	if strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint = serverEndpoint + "notifications"
	} else {
		serverEndpoint = serverEndpoint + "/notifications"
	}
	httpClient, err := httpclient.New(tlsmanager.ClientCommandDelegator, tm)
	if err != nil {
		return nil, err
	}
	return &commandDelegatorApi{
		serverEndpoint: serverEndpoint,
		client:         httpClient,
		apiKeysManager: apiKeysManager,
	}, nil
}

func (c *commandDelegatorApi) SendNotification(orgName string, clusterName string, clientRequest clientrequest.NotificationRequestEvent) string {
	payload := []byte(serializer.Serialize(clientRequest))
	request, err := http.NewRequest("POST", c.serverEndpoint+fmt.Sprintf("/%s/%s", orgName, clusterName), bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(apikeysmanager.ApiKeyHeaderName, c.apiKeysManager.GetWorkflowTriggerApiKey())
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
