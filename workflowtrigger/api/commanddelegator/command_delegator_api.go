package commanddelegator

import (
	"bytes"
	"fmt"
	"github.com/greenopsinc/util/clientrequest"
	"greenops.io/workflowtrigger/serializer"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	RootCommit       string = "ROOT_COMMIT"
	PipelineFileName string = "pipeline.yaml"
	getFileExtension string = "file"
)

type GetFileRequest struct {
	GitRepoUrl    string
	Filename      string
	GitCommitHash string
}

type CommandDelegatorApi interface {
	SendNotification(orgName string, clusterName string, request clientrequest.NotificationRequestEvent) string
}

type CommandDelegatorApiImpl struct {
	serverEndpoint string
	client         *http.Client
}

func New(serverEndpoint string) CommandDelegatorApi {
	if strings.HasSuffix(serverEndpoint, "/") {
		serverEndpoint = serverEndpoint + "notifications"
	} else {
		serverEndpoint = serverEndpoint + "/notifications"
	}
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	return &CommandDelegatorApiImpl{
		serverEndpoint: serverEndpoint,
		client:         httpClient,
	}
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

