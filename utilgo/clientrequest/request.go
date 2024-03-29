package clientrequest

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1 "k8s.io/api/core/v1"
)

const (
	DeployArgoRequest       string = "DeployArgoRequest"
	DeployKubernetesRequest string = "DeployKubernetesRequest"
	DeployTestRequest       string = "DeployTestRequest"
	DeleteArgoRequest       string = "DeleteArgoRequest"
	DeleteKubernetesRequest string = "DeleteKubernetesRequest"
	DeleteTestRequest       string = "DeleteTestRequest"
)

type RollbackRequest struct {
	AppName    string `json:"appName"`
	RevisionId string `json:"revisionId"`
}

type WatchRequest struct {
	ClientRequestEventMetadata
	Type        string `json:"type"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	PipelineUvn string `json:"pipelineUvn"`
	TestNumber  int    `json:"testNumber"`
}

type KubernetesCreationRequest struct {
	Type           string          `json:"type"`
	Kind           string          `json:"kind"`
	ObjectName     string          `json:"objectName"`
	Namespace      string          `json:"namespace"`
	ImageName      string          `json:"imageName"`
	Command        []string        `json:"command"`
	Args           []string        `json:"args"`
	Config         string          `json:"configPayload"`
	VolumeFilename string          `json:"volumeFilename"`
	VolumeConfig   string          `json:"volumePayload"`
	Variables      []corev1.EnvVar `json:"variables"`
}

type GvkGroupRequest struct {
	ResourceList []GvkResourceInfo
}

type GvkResourceInfo struct {
	schema.GroupVersionKind
	ResourceName      string
	ResourceNamespace string
}
