package event

import (
	"github.com/google/uuid"
	"github.com/greenopsinc/util/serializerutil"
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

func MarshalEvent(event Event) map[string]interface{} {
	switch event.(type) {
	case *PipelineTriggerEvent:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.PipelineTriggerEventType
		return mapObj
	default:
		panic("Matching event type not found")
	}
}
