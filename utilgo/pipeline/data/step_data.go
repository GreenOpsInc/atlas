package data

type StepData struct {
	Name                 string     `json:"name"`
	ArgoApplication      string     `json:"argo_application"`
	ArgoApplicationPath  string     `json:"argo_application_path"`
	OtherDeploymentsPath string     `json:"other_deployments_path"`
	ClusterName          string     `json:"cluster_name"`
	Tests                []TestData `json:"tests"`
	RemediationLimit     string     `json:"remediation_limit"`
	RollbackLimit        string     `json:"rollback_limit"`
	Dependencies         []string   `json:"dependencies"`
}
