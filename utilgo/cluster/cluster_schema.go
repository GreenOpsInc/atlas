package cluster

import "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"

const (
	LocalClusterName string = "in-cluster"
)

type ClusterSchema struct {
	ClusterIP   string        `json:"clusterIP"`
	ClusterName string        `json:"clusterName"`
	NoDeploy    *NoDeployInfo `json:"noDeploy"`
}

type NoDeployInfo struct {
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Namespace string `json:"namespace"`
}

type ClusterCreateRequest struct {
	Server string                  `json:"server"`
	Name   string                  `json:"name"`
	Config *v1alpha1.ClusterConfig `json:"config"`
}
