package requestdatatypes

type DeployResponse struct {
	Success      bool
	ResourceName string
	AppNamespace string
	RevisionHash string
}

type RemediationResponse struct {
	Success      bool
	ResourceName string
	AppNamespace string
	RevisionHash string
}

type ResponseEventType string

const (
	ApplicationInfraCompletionEventType ResponseEventType = "ApplicationInfraCompletionEvent"
)

func (eventType ResponseEventType) MakeResponseEvent(deployResponse *DeployResponse, deployRequest *ClientDeployRequest) ResponseEvent {
	if eventType == ApplicationInfraCompletionEventType {
		return ApplicationInfraCompletionEvent{}.GetEvent(deployResponse, deployRequest)
	}
	return nil
}

type ResponseEvent interface {
	GetEvent(deployResponse *DeployResponse, deployRequest *ClientDeployRequest) ResponseEvent
}

type ApplicationInfraCompletionEvent struct {
	Type         string `json:"type"`
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	PipelineUvn  string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	Success      bool   `json:"success"`
}

func (event ApplicationInfraCompletionEvent) GetEvent(deployResponse *DeployResponse, deployRequest *ClientDeployRequest) ResponseEvent {
	event.OrgName = deployRequest.OrgName
	event.TeamName = deployRequest.TeamName
	event.PipelineName = deployRequest.PipelineName
	event.PipelineUvn = deployRequest.PipelineUvn
	event.StepName = deployRequest.StepName
	event.Success = deployResponse.Success
	event.Type = "appinfracompletion"
	return event
}
