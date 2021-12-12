package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	WorkflowTriggerEnvVar         string = "WORKFLOW_TRIGGER_SERVER_ADDR"
	DefaultWorkflowTriggerAddress string = "http://workflowtrigger.atlas.svc.cluster.local:8080"
)


type Notification struct {
	Successful bool
	Body       interface{}
}

type EventGenerationApi interface {
	//TODO: Make these void methods. The only case where the generation should not work is in case of service unavailability, and we should be blocking and retrying if a service is unavaibale.
	GenerateEvent(eventInfo datamodel.EventInfo) bool
	GenerateNotification(requestId string, notification Notification) bool
	GenerateResponseEvent(responseEvent requestdatatypes.ResponseEvent) bool
}

type EventGenerationImpl struct {
	workflowTriggerAddress string
	client                 *http.Client
}

func Create() EventGenerationApi {
	workflowTriggerAddress := os.Getenv(WorkflowTriggerEnvVar)
	if workflowTriggerAddress == "" {
		workflowTriggerAddress = DefaultWorkflowTriggerAddress
	}
	if strings.HasSuffix(workflowTriggerAddress, "/") {
		workflowTriggerAddress = strings.TrimSuffix(workflowTriggerAddress, "/")
	}
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	return EventGenerationImpl{workflowTriggerAddress: workflowTriggerAddress, client: httpClient}
}

func (c EventGenerationImpl) GenerateEvent(eventInfo datamodel.EventInfo) bool {
	data, err := json.Marshal(eventInfo)
	if err != nil {
		return false
	}
	req, _ := http.NewRequest("POST", c.workflowTriggerAddress+"/client/generateEvent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error generating event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
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
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error generating notification. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode/100 == 2
}

func (c EventGenerationImpl) GenerateResponseEvent(responseEvent requestdatatypes.ResponseEvent) bool {
	data, err := json.Marshal(responseEvent)
	if err != nil {
		return false
	}
	req, _ := http.NewRequest("POST", c.workflowTriggerAddress+"/client/generateEvent", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Error generating event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode/100 == 2
}
