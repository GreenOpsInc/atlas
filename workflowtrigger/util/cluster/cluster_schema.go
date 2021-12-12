package cluster

type ClusterSchema struct {
	ClusterIP   string        `json:"clusterIP"`
	ExposedPort int           `json:"exposedPort"`
	ClusterName string        `json:"clusterName"`
	NoDeploy    *NoDeployInfo `json:"noDeploy"`
}

type NoDeployInfo struct {
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Namespace string `json:"namespace"`
}
