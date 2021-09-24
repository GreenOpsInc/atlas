package generation

import (
	"bytes"
	"encoding/json"
	"greenops.io/client/atlasoperator/requestdatatypes"
	"greenops.io/client/progressionchecker/datamodel"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	WorkflowTriggerEnvVar         string = "WORKFLOW_TRIGGER_SERVER_ADDR"
	DefaultWorkflowTriggerAddress string = "http://workflowtrigger.atlas.svc.cluster.local:8080"
)

type EventGenerationApi interface {
	//TODO: Make these void methods. The only case where the generation should not work is in case of service unavailability, and we should be blocking and retrying if a service is unavaibale.
	GenerateEvent(eventInfo datamodel.EventInfo) bool
	GenerateResponseEvent(responseEvent requestdatatypes.ResponseEvent) bool
}

type EventGenerationImpl struct {
	workflowTriggerAddress string
}

func Create() EventGenerationApi {
	workflowTriggerAddress := os.Getenv(WorkflowTriggerEnvVar)
	if workflowTriggerAddress == "" {
		workflowTriggerAddress = DefaultWorkflowTriggerAddress
	}
	if strings.HasSuffix(workflowTriggerAddress, "/") {
		workflowTriggerAddress = strings.TrimSuffix(workflowTriggerAddress, "/")
	}
	return EventGenerationImpl{workflowTriggerAddress: workflowTriggerAddress}
}

func (c EventGenerationImpl) GenerateEvent(eventInfo datamodel.EventInfo) bool {
	data, err := json.Marshal(eventInfo)
	if err != nil {
		return false
	}
	resp, err := http.Post(c.workflowTriggerAddress+"/client/generateEvent", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error generating event. Error was %s\n", err)
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
	resp, err := http.Post(c.workflowTriggerAddress+"/client/generateEvent", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error generating event. Error was %s\n", err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode/100 == 2
}
