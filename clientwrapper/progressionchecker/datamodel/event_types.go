package datamodel

import (
	"greenops.io/client/atlasoperator/requestdatatypes"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type EventInfoType string

const (
	ClientCompletionEvent EventInfoType = "clientcompletion"
	TestCompletionEvent   EventInfoType = "testcompletion"
	FailureEvent          EventInfoType = "failureevent"
)

type EventInfo interface {
	GetEventType() EventInfoType
}

type EventInfoMetaData struct {
	Type         EventInfoType `json:"type"`
	OrgName      string        `json:"orgName"`
	TeamName     string        `json:"teamName"`
	PipelineName string        `json:"pipelineName"`
	StepName     string        `json:"stepName"`
}

type ApplicationEventInfo struct {
	EventInfoMetaData
	HealthStatus     string           `json:"healthStatus"`
	SyncStatus       string           `json:"syncStatus"`
	ResourceStatuses []ResourceStatus `json:"resourceStatuses"`
	ArgoName         string           `json:"argoName"`
	ArgoNamespace    string           `json:"argoNamespace"`
	Operation        string           `json:"operation"`
	Project          string           `json:"project"`
	Repo             string           `json:"repo"`
	RevisionHash     string           `json:"revisionId"`
}

type TestEventInfo struct {
	EventInfoMetaData
	Successful bool   `json:"successful"`
	Log        string `json:"log"`
	TestName   string `json:"testName"`
	TestNumber int    `json:"testNumber"`
}

type FailureEventInfo struct {
	EventInfoMetaData
	requestdatatypes.DeployResponse
	StatusCode string `json:"statusCode"`
	Error      string `json:"error"`
}

type ResourceStatus struct {
	schema.GroupVersionKind
	ResourceName      string `json:"resourceName"`
	ResourceNamespace string `json:"resourceNamespace"`
	HealthStatus      string `json:"healthStatus"`
	SyncStatus        string `json:"syncStatus"`
}

// **
//ApplicationEventInfo functionality
// **
func (eventInfo ApplicationEventInfo) GetEventType() EventInfoType {
	return eventInfo.Type
}

func MakeApplicationEvent(key WatchKey, appInfo ArgoAppMetricInfo, healthStatus string, syncStatus string, resourceStatuses []ResourceStatus, revisionHash string) EventInfo {
	return ApplicationEventInfo{
		EventInfoMetaData: EventInfoMetaData{
			Type:         ClientCompletionEvent,
			OrgName:      key.OrgName,
			TeamName:     key.TeamName,
			PipelineName: key.PipelineName,
			StepName:     key.StepName,
		},
		HealthStatus:     healthStatus,
		SyncStatus:       syncStatus,
		ResourceStatuses: resourceStatuses,
		ArgoName:         appInfo.Name,
		Operation:        appInfo.Operation,
		Project:          appInfo.Project,
		Repo:             appInfo.Repo,
		RevisionHash:     revisionHash,
	}
}

// **
//TestEventInfo functionality
// **
func (eventInfo TestEventInfo) GetEventType() EventInfoType {
	return eventInfo.Type
}

func MakeTestEvent(key WatchKey, successful bool, logs string) EventInfo {
	return TestEventInfo{
		EventInfoMetaData: EventInfoMetaData{
			Type:         TestCompletionEvent,
			OrgName:      key.OrgName,
			TeamName:     key.TeamName,
			PipelineName: key.PipelineName,
			StepName:     key.StepName,
		},
		Successful: successful,
		Log:        logs,
		TestName:   key.Name,
		TestNumber: key.TestNumber,
	}
}

// **
//FailureEventInfo functionality
// **
func (eventInfo FailureEventInfo) GetEventType() EventInfoType {
	return eventInfo.Type
}

func MakeFailureEventEvent(clientMetadata requestdatatypes.ClientEventMetadata, deployResponse requestdatatypes.DeployResponse, statusCode string, error string) EventInfo {
	return FailureEventInfo{
		EventInfoMetaData: EventInfoMetaData{
			Type:         FailureEvent,
			OrgName:      clientMetadata.OrgName,
			TeamName:     clientMetadata.TeamName,
			PipelineName: clientMetadata.PipelineName,
			StepName:     clientMetadata.StepName,
		},
		DeployResponse: deployResponse,
		StatusCode:     statusCode,
		Error:          error,
	}
}
