package requestdatatypes

const (
	DeployArgoRequest       string = "DeployArgoRequest"
	DeployKubernetesRequest string = "DeployKubernetesRequest"
	DeployTestRequest       string = "DeployTestRequest"
	DeleteArgoRequest       string = "DeleteArgoRequest"
	DeleteKubernetesRequest string = "DeleteKubernetesRequest"
	DeleteTestRequest       string = "DeleteTestRequest"
)

type WatchRequest struct {
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	StepName     string `json:"stepName"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	TestNumber   int    `json:"testNumber"`
}

type KubernetesCreationRequest struct {
	Kind       string            `json:"kind"`
	ObjectName string            `json:"objectName"`
	Namespace  string            `json:"namespace"`
	ImageName  string            `json:"imageName"`
	Command    []string          `json:"command"`
	Args       []string          `json:"args"`
	Config     string            `json:"configPayload"`
	Variables  map[string]string `json:"variables"`
}
