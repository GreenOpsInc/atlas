package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/greenopsinc/util/clientrequest"
	"greenops.io/client/api/ingest"
	"greenops.io/client/argodriver"
	"greenops.io/client/progressionchecker/datamodel"

	"greenops.io/client/client"
	"greenops.io/client/kubernetesclient"
	"greenops.io/client/tlsmanager"
)

const (
	WorkflowTriggerEnvVar         string = "WORKFLOW_TRIGGER_SERVER_ADDR"
	DefaultWorkflowTriggerAddress string = "https://workflowtrigger.atlas.svc.cluster.local:8080"
)

type Notification struct {
	Successful bool
	Body       interface{}
}

type EventGenerationApi interface {
	//TODO: Make these void methods. The only case where the generation should not work is in case of service unavailability, and we should be blocking and retrying if a service is unavaibale.
	GenerateEvent(eventInfo datamodel.EventInfo) bool
	GenerateNotification(requestId string, notification Notification) bool
	GenerateResponseEvent(responseEvent clientrequest.ResponseEvent) bool
}

type EventGenerationImpl struct {
	workflowTriggerAddress string
	client                 client.HttpClient
	argoAuthClient         argodriver.ArgoAuthClient
	kclient                kubernetesclient.KubernetesClient
}

func Create(argoClient argodriver.ArgoAuthClient, kclient kubernetesclient.KubernetesClient, tm tlsmanager.Manager) (EventGenerationApi, error) {
	workflowTriggerAddress := os.Getenv(WorkflowTriggerEnvVar)
	if workflowTriggerAddress == "" {
		workflowTriggerAddress = DefaultWorkflowTriggerAddress
	}
	if strings.HasSuffix(workflowTriggerAddress, "/") {
		workflowTriggerAddress = strings.TrimSuffix(workflowTriggerAddress, "/")
	}
	httpClient, err := client.NewHttpClient(tlsmanager.ClientWorkflowTrigger, tm)
	if err != nil {
		return nil, err
	}
	return EventGenerationImpl{workflowTriggerAddress: workflowTriggerAddress, client: httpClient, argoAuthClient: argoClient, kclient: kclient}, nil
}

func (c EventGenerationImpl) GenerateEvent(eventInfo datamodel.EventInfo) bool {
	data, err := json.Marshal(eventInfo)
	if err != nil {
		return false
	}
	clusterName := os.Getenv(ingest.EnvClusterName)
	if clusterName == "" {
		clusterName = ingest.DefaultClusterName
	}
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/client/%s/%s/generateEvent", c.workflowTriggerAddress, eventInfo.GetEventOrg(), clusterName), bytes.NewBuffer(data))
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error generating event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	log.Printf("Status code for generating event was %d\n", resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		log.Printf("Error was %s", resp.Status)
	}
	return resp.StatusCode/100 == 2
}

func (c EventGenerationImpl) GenerateNotification(requestId string, notification Notification) bool {
	bodyData, err := json.Marshal(notification.Body)
	if err != nil {
		return false
	}
	notification.Body = string(bodyData)
	data, err := json.Marshal(notification)
	if err != nil {
		return false
	}
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/client/generateNotification/%s", c.workflowTriggerAddress, requestId), bytes.NewBuffer(data))
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error generating notification. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	log.Printf("Status code for generating notification event was %d\n", resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		log.Printf("Error was %s", resp.Status)
	}
	return resp.StatusCode/100 == 2
}

func (c EventGenerationImpl) GenerateResponseEvent(responseEvent clientrequest.ResponseEvent) bool {
	data, err := json.Marshal(responseEvent)
	if err != nil {
		return false
	}
	clusterName := os.Getenv(ingest.EnvClusterName)
	if clusterName == "" {
		clusterName = ingest.DefaultClusterName
	}
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/client/%s/%s/generateEvent", c.workflowTriggerAddress, responseEvent.GetEventOrg(), clusterName), bytes.NewBuffer(data))
	req.Header.Add("Authorization", "Bearer "+c.argoAuthClient.GetAuthToken())
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error generating event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	log.Printf("Status code for generating response event was %d\n", resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		log.Printf("Error was %s", resp.Status)
	}
	return resp.StatusCode/100 == 2
}
