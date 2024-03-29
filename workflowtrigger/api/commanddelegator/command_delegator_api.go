package commanddelegator

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/greenopsinc/util/clientrequest"
	"greenops.io/workflowtrigger/serializer"

	"github.com/greenopsinc/util/httpclient"
	"github.com/greenopsinc/util/tlsmanager"
)

const (
	RootCommit       string = "ROOT_COMMIT"
	PipelineFileName string = "pipeline.yaml"
	getFileExtension string = "file"
)

type CommandDelegatorApi interface {
	SendNotification(orgName string, clusterName string, request clientrequest.NotificationRequestEvent) string
}

type CommandDelegatorApiImpl struct {
	serverEndpoint string
	client         httpclient.HttpClient
}

func New(serverEndpoint string, tm tlsmanager.Manager) (CommandDelegatorApi, error) {
	if strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint = serverEndpoint + "notifications"
	} else {
		serverEndpoint = serverEndpoint + "/notifications"
	}
	httpClient, err := httpclient.New(tlsmanager.ClientCommandDelegator, tm)
	if err != nil {
		return nil, err
	}
	return &CommandDelegatorApiImpl{
		serverEndpoint: serverEndpoint,
		client:         httpClient,
	}, nil
}

func (r *CommandDelegatorApiImpl) SendNotification(orgName string, clusterName string, clientRequest clientrequest.NotificationRequestEvent) string {
	var err error
	var payload []byte
	var request *http.Request
	payload = []byte(serializer.Serialize(clientRequest))
	request, err = http.NewRequest("POST", r.serverEndpoint+fmt.Sprintf("/%s/%s", orgName, clusterName), bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(request)
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
