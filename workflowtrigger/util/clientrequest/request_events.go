package clientrequest

const (
	ClientMarkNoDeployRequestType string = "mark_no_deploy"
)

type NotificationRequestEvent interface {
	GetEvent() string
}

type ClientEventMetadata struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	PipelineUvn  string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	RequestId    string `json:"requestId"`
}

//*****
//*****
//Notifications Events: Need to have a request ID
//*****
//*****

// ClientMarkNoDeployRequest -----
type ClientMarkNoDeployRequest struct {
	ClientEventMetadata
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
}

func (r ClientMarkNoDeployRequest) GetEvent() string {
	return ClientMarkNoDeployRequestType
}
