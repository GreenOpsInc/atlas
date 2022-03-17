package request

type KubernetesCreationRequest struct {
	Type           string        `json:"type"`
	Kind           string        `json:"kind"`
	ObjectName     string        `json:"objectName"`
	Namespace      string        `json:"namespace"`
	ImageName      string        `json:"imageName"`
	Command        []string      `json:"command"`
	Args           []string      `json:"args"`
	ConfigPayload  string        `json:"configPayload"`
	VolumeFilename string        `json:"volumeFilename"`
	VolumePayload  string        `json:"volumePayload"`
	Variables      []interface{} `json:"variables"`
}
