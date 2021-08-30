package requestdatatypes

const (
	ClientDeployRequestType                      string = "deploy"
	ClientDeleteByConfigRequestType              string = "del_config"
	ClientDeleteByGvkRequestType                 string = "del_gvk"
	ClientDeployAndWatchRequestType              string = "deploy_watch"
	ClientDeployArgoAppByNameRequestType         string = "deploy_namedargo"
	ClientDeployArgoAppByNameAndWatchRequestType string = "deploy_namedargo_watch"
	ClientRollbackAndWatchRequestType            string = "rollback"
	ClientSelectiveSyncRequestType               string = "sel_sync_watch"
)

type RequestEvent interface {
	GetEvent() string
}
type ClientEventMetadata struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	StepName     string `json:"stepName"`
}

// ClientDeployRequest -----
type ClientDeployRequest struct {
	ClientEventMetadata
	ResponseEventType ResponseEventType `json:"responseEventType"`
	DeployType        string            `json:"deployType"`
	RevisionHash      string            `json:"revisionHash"`
	Payload           string            `json:"payload"`
}

func (r ClientDeployRequest) GetEvent() string {
	return ClientDeployRequestType
}

// ClientDeleteByConfigRequest -----
type ClientDeleteByConfigRequest struct {
	ClientEventMetadata
	DeleteType    string `json:"deleteType"`
	ConfigPayload string `json:"configPayload"`
}

func (r ClientDeleteByConfigRequest) GetEvent() string {
	return ClientDeleteByConfigRequestType
}

// ClientDeleteByGvkRequest -----
type ClientDeleteByGvkRequest struct {
	ClientEventMetadata
	DeleteType        string `json:"deleteType"`
	ResourceName      string `json:"resourceName"`
	ResourceNamespace string `json:"resourceNamespace"`
	Group             string `json:"group"`
	Version           string `json:"version"`
	Kind              string `json:"kind"`
}

func (r ClientDeleteByGvkRequest) GetEvent() string {
	return ClientDeleteByGvkRequestType
}

// ClientDeployAndWatchRequest -----
type ClientDeployAndWatchRequest struct {
	ClientEventMetadata
	DeployType   string `json:"deployType"`
	RevisionHash string `json:"revisionHash"`
	Payload      string `json:"payload"`
	WatchType    string `json:"watchType"`
	TestNumber   int    `json:"testNumber"`
}

func (r ClientDeployAndWatchRequest) GetEvent() string {
	return ClientDeployAndWatchRequestType
}

// ClientRollbackAndWatchRequest -----
type ClientRollbackAndWatchRequest struct {
	ClientEventMetadata
	AppName      string `json:"appName"`
	RevisionHash string `json:"revisionHash"`
	WatchType    string `json:"watchType"`
}

func (r ClientRollbackAndWatchRequest) GetEvent() string {
	return ClientRollbackAndWatchRequestType
}

// ClientSelectiveSyncRequest -----
type ClientSelectiveSyncRequest struct {
	ClientEventMetadata
	AppName         string          `json:"appName"`
	RevisionHash    string          `json:"revisionHash"`
	GvkResourceList GvkGroupRequest `json:"resourcesGvkRequest"`
}

func (r ClientSelectiveSyncRequest) GetEvent() string {
	return ClientSelectiveSyncRequestType
}
