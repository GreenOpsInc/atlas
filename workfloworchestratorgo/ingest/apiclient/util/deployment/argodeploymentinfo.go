package deployment

var (
	NoOpArgoDeployment = &ArgoDeploymentInfo{}
)

type ArgoDeploymentInfo struct {
	ArgoApplicationName string
	ArgoRevisionHash    string
}
