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
	GetClientMetadata() ClientEventMetadata
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

func (r ClientDeployRequest) GetClientMetadata() ClientEventMetadata {
	return r.ClientEventMetadata
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

func (r ClientDeleteByConfigRequest) GetClientMetadata() ClientEventMetadata {
	return r.ClientEventMetadata
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

func (r ClientDeleteByGvkRequest) GetClientMetadata() ClientEventMetadata {
	return r.ClientEventMetadata
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

func (r ClientDeployAndWatchRequest) GetClientMetadata() ClientEventMetadata {
	return r.ClientEventMetadata
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

func (r ClientRollbackAndWatchRequest) GetClientMetadata() ClientEventMetadata {
	return r.ClientEventMetadata
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

func (r ClientSelectiveSyncRequest) GetClientMetadata() ClientEventMetadata {
	return r.ClientEventMetadata
}
