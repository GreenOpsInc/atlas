package cluster

type ClusterSchema struct {
	ClusterIP   string `json:"clusterIP"`
	ExposedPort int    `json:"exposedPort"`
	ClusterName string `json:"clusterName"`
}
