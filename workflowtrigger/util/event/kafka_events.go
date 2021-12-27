package event

import (
	"github.com/google/uuid"
	"gitlab.com/c0b/go-ordered-json"
	"greenops.io/workflowtrigger/util/serializerutil"
)

const (
	rootStepName string = "ATLAS_ROOT_DATA"
	rootCommit   string = "ROOT_COMMIT"
)

type Event interface {
	GetUvn() string
}

type PipelineTriggerEvent struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	Uvn          string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	RevisionHash string `json:"revisionHash"`
}

func NewPipelineTriggerEventRaw(orgName string, teamName string, pipelineName string, pipelineUvn string) Event {
	return &PipelineTriggerEvent{
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		Uvn:          pipelineUvn,
		StepName:     rootStepName,
	}
}

func NewPipelineTriggerEvent(orgName string, teamName string, pipelineName string) Event {
	return &PipelineTriggerEvent{
		OrgName:      orgName,
		TeamName:     teamName,
		PipelineName: pipelineName,
		Uvn:          uuid.New().String(),
		StepName:     rootStepName,
		RevisionHash: rootCommit,
	}
}

func (p *PipelineTriggerEvent) GetUvn() string {
	return p.Uvn
}

func MarshalEvent(event Event) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	switch event.(type) {
	case *PipelineTriggerEvent:
		mapObj.Set("orgName", event.(*PipelineTriggerEvent).OrgName)
		mapObj.Set("teamName", event.(*PipelineTriggerEvent).TeamName)
		mapObj.Set("pipelineName", event.(*PipelineTriggerEvent).PipelineName)
		mapObj.Set("pipelineUvn", event.(*PipelineTriggerEvent).Uvn)
		mapObj.Set("stepName", event.(*PipelineTriggerEvent).StepName)
		mapObj.Set("revisionHash", event.(*PipelineTriggerEvent).RevisionHash)
		mapObj.Set("type", serializerutil.PipelineTriggerEventType)
		return mapObj
	default:
		panic("Matching event type not found")
	}
}
