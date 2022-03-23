package event

import (
	"github.com/google/uuid"
	"github.com/greenopsinc/util/serializerutil"
)

const (
	rootStepName             = "ATLAS_ROOT_DATA"
	rootCommit               = "ROOT_COMMIT"
	PipelineTriggerEventName = "com.greenops.utilgo.datamodel.kafka_events.PipelineTriggerEvent"

	HealthyStatus     = "Healthy"
	ProgressingStatus = "Progressing"
	UnknownStatus     = "Unknown"
	DegradedStatus    = "Degraded"
	SuspendedStatus   = "Suspended"
	MissingStatus     = "Missing"
	SyncedStatus      = "Synced"
	OutOfSyncStatus   = "OutOfSync"
)

type Event interface {
	GetUVN() string
	GetOrgName() string
	GetTeamName() string
	GetPipelineName() string
	GetStepName() string
}

type PipelineTriggerEvent struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	UVN          string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	RevisionHash string `json:"revisionHash"`
}

func (e *PipelineTriggerEvent) GetUVN() string {
	return e.UVN
}

func (e *PipelineTriggerEvent) GetOrgName() string {
	return e.OrgName
}

func (e *PipelineTriggerEvent) GetTeamName() string {
	return e.TeamName
}

func (e *PipelineTriggerEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *PipelineTriggerEvent) GetStepName() string {
	return e.StepName
}

type ApplicationInfraCompletionEvent struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	UVN          string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	Success      bool   `json:"success"`
}

func (e *ApplicationInfraCompletionEvent) GetUVN() string {
	return e.UVN
}

func (e *ApplicationInfraCompletionEvent) GetOrgName() string {
	return e.OrgName
}

func (e *ApplicationInfraCompletionEvent) GetTeamName() string {
	return e.TeamName
}

func (e *ApplicationInfraCompletionEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *ApplicationInfraCompletionEvent) GetStepName() string {
	return e.StepName
}

type ApplicationInfraTriggerEvent struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	UVN          string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
}

func (e *ApplicationInfraTriggerEvent) GetUVN() string {
	return e.UVN
}

func (e *ApplicationInfraTriggerEvent) GetOrgName() string {
	return e.OrgName
}

func (e *ApplicationInfraTriggerEvent) GetTeamName() string {
	return e.TeamName
}

func (e *ApplicationInfraTriggerEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *ApplicationInfraTriggerEvent) GetStepName() string {
	return e.StepName
}

type ClientCompletionEvent struct {
	HealthStatus     string            `json:"healthStatus"`
	SyncStatus       string            `json:"syncStatus"`
	ResourceStatuses []*ResourceStatus `json:"resourceStatuses"`
	OrgName          string            `json:"orgName"`
	TeamName         string            `json:"teamName"`
	PipelineName     string            `json:"pipelineName"`
	UVN              string            `json:"pipelineUvn"`
	StepName         string            `json:"stepName"`
	ArgoName         string            `json:"argoName"`
	ArgoNamespace    string            `json:"argoNamespace"`
	Operation        string            `json:"operation"`
	Project          string            `json:"project"`
	Repo             string            `json:"repo"`
	RevisionHash     string            `json:"revisionHash"`
}

func (e *ClientCompletionEvent) GetUVN() string {
	return e.UVN
}

func (e *ClientCompletionEvent) GetOrgName() string {
	return e.OrgName
}

func (e *ClientCompletionEvent) GetTeamName() string {
	return e.TeamName
}

func (e *ClientCompletionEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *ClientCompletionEvent) GetStepName() string {
	return e.StepName
}

type ResourceStatus struct {
	ResourceName      string `json:"resourceName"`
	ResourceNamespace string `json:"resourceNamespace"`
	HealthStatus      string `json:"healthStatus"`
	SyncStatus        string `json:"syncStatus"`
	Group             string `json:"group"`
	Version           string `json:"version"`
	Kind              string `json:"kind"`
}

type FailureEvent struct {
	OrgName        string          `json:"orgName"`
	TeamName       string          `json:"teamName"`
	PipelineName   string          `json:"pipelineName"`
	UVN            string          `json:"pipelineUvn"`
	StepName       string          `json:"stepName"`
	DeployResponse *DeployResponse `json:"deployResponse"`
	StatusCode     string          `json:"statusCode"`
	Error          string          `json:"error"`
}

func (e *FailureEvent) GetUVN() string {
	return e.UVN
}

func (e *FailureEvent) GetOrgName() string {
	return e.OrgName
}

func (e *FailureEvent) GetTeamName() string {
	return e.TeamName
}

func (e *FailureEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *FailureEvent) GetStepName() string {
	return e.StepName
}

type DeployResponse struct {
	Success              bool   `json:"success"`
	ResourceName         string `json:"resourceName"`
	ApplicationNamespace string `json:"applicationNamespace"`
	RevisionHash         string `json:"revisionHash"`
}

type TestCompletionEvent struct {
	Successful   bool   `json:"successful"`
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	UVN          string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	Log          string `json:"log"`
	TestName     string `json:"testName"`
	TestNumber   int    `json:"testNumber"`
}

func (e *TestCompletionEvent) GetUVN() string {
	return e.UVN
}

func (e *TestCompletionEvent) GetOrgName() string {
	return e.OrgName
}

func (e *TestCompletionEvent) GetTeamName() string {
	return e.TeamName
}

func (e *TestCompletionEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *TestCompletionEvent) GetStepName() string {
	return e.StepName
}

type TriggerStepEvent struct {
	OrgName       string `json:"orgName"`
	TeamName      string `json:"teamName"`
	PipelineName  string `json:"pipelineName"`
	StepName      string `json:"stepName"`
	UVN           string `json:"pipelineUvn"`
	GitCommitHash string `json:"gitCommitHash"`
	Rollback      bool   `json:"rollback"`
}

func (e *TriggerStepEvent) GetUVN() string {
	return e.UVN
}

func (e *TriggerStepEvent) GetOrgName() string {
	return e.OrgName
}

func (e *TriggerStepEvent) GetTeamName() string {
	return e.TeamName
}

func (e *TriggerStepEvent) GetPipelineName() string {
	return e.PipelineName
}

func (e *TriggerStepEvent) GetStepName() string {
	return e.StepName
}

func NewPipelineTriggerEventRaw(orgName string, teamName string, pipelineName string, pipelineUVN string) Event {
	return &PipelineTriggerEvent{
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		UVN:          pipelineUVN,
		StepName:     rootStepName,
	}
}

func NewPipelineTriggerEvent(orgName string, teamName string, pipelineName string) Event {
	return &PipelineTriggerEvent{
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		UVN:          uuid.New().String(),
		StepName:     rootStepName,
		RevisionHash: rootCommit,
	}
}

func NewApplicationInfraCompletionEvent(orgName string, teamName string, pipelineName string, success bool) Event {
	return &ApplicationInfraCompletionEvent{
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		UVN:          uuid.New().String(),
		StepName:     rootStepName,
		Success:      success,
	}
}

func NewApplicationInfraTriggerEvent(orgName string, teamName string, pipelineName string) Event {
	return &ApplicationInfraTriggerEvent{
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		UVN:          uuid.New().String(),
		StepName:     rootStepName,
	}
}

func NewClientCompletionEvent(
	healthStatus string,
	syncStatus string,
	resourceStatuses []*ResourceStatus,
	orgName string,
	teamName string,
	pipelineName string,
	argoName string,
	argoNamespace string,
	operation string,
	project string,
	repo string,
) Event {
	return &ClientCompletionEvent{
		HealthStatus:     healthStatus,
		SyncStatus:       syncStatus,
		ResourceStatuses: resourceStatuses,
		OrgName:          orgName,
		TeamName:         teamName,
		PipelineName:     pipelineName,
		UVN:              uuid.New().String(),
		StepName:         rootStepName,
		ArgoName:         argoName,
		ArgoNamespace:    argoNamespace,
		Operation:        operation,
		Project:          project,
		Repo:             repo,
		RevisionHash:     rootCommit,
	}
}

func NewFailureEvent(
	orgName string,
	teamName string,
	pipelineName string,
	deployResponse *DeployResponse,
	statusCode string,
	err string,
) Event {
	return &FailureEvent{
		OrgName:        orgName,
		TeamName:       teamName,
		PipelineName:   pipelineName,
		UVN:            uuid.New().String(),
		StepName:       rootStepName,
		DeployResponse: deployResponse,
		StatusCode:     statusCode,
		Error:          err,
	}
}

func NewTestCompletionEvent(
	successful bool,
	orgName string,
	teamName string,
	pipelineName string,
	log string,
	testName string,
	testNumber int,
) Event {
	return &TestCompletionEvent{
		Successful:   successful,
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		UVN:          uuid.New().String(),
		StepName:     rootStepName,
		Log:          log,
		TestName:     testName,
		TestNumber:   testNumber,
	}
}

func NewTriggerStepEvent(orgName string, teamName string, pipelineName string, gitCommitHash string, rollback bool) Event {
	return &TriggerStepEvent{
		OrgName:       orgName,
		TeamName:      teamName,
		PipelineName:  pipelineName,
		UVN:           uuid.New().String(),
		StepName:      rootStepName,
		GitCommitHash: gitCommitHash,
		Rollback:      rollback,
	}
}

func MarshalEvent(event Event) map[string]interface{} {
	switch e := event.(type) {
	case *PipelineTriggerEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.PipelineTriggerEventType
		return mapObj
	case *ApplicationInfraCompletionEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.ApplicationInfraCompletionEventType
		return mapObj
	case *ApplicationInfraTriggerEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.ApplicationInfraTriggerEventType
		return mapObj
	case *ClientCompletionEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		statuses := make([]map[string]interface{}, len(e.ResourceStatuses))
		for _, s := range e.ResourceStatuses {
			statuses = append(statuses, serializerutil.GetMapFromStruct(s))
		}
		mapObj["resourceStatuses"] = statuses
		mapObj["type"] = serializerutil.ClientCompletionEventType
		return mapObj
	case *FailureEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		deployResponse := serializerutil.GetMapFromStruct(e.DeployResponse)
		mapObj["deployResponse"] = deployResponse
		mapObj["type"] = serializerutil.FailureEventType
		return mapObj
	case *TestCompletionEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.TestCompletionEventType
		return mapObj
	case *TriggerStepEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.TriggerStepEventType
		return mapObj
	default:
		panic("matching event type not found")
	}
}
