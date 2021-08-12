package requestdatatypes

type DeployResponse struct {
	Success      bool
	ResourceName string
	AppNamespace string
	RevisionHash string
}

type RemediationResponse struct {
	Success      bool
	ResourceName string
	AppNamespace string
	RevisionHash string
}
