package data

type StepData struct {
	Name                 string     `json:"name"`
	ArgoApplication      string     `json:"argo_application"`
	ArgoApplicationPath  string     `json:"argo_application_path"`
	OtherDeploymentsPath string     `json:"other_deployments_path"`
	ClusterName          string     `json:"cluster_name"`
	Tests                []TestData `json:"tests"`
	RemediationLimit     int        `json:"remediation_limit"`
	RollbackLimit        int        `json:"rollback_limit"`
	Dependencies         []string   `json:"dependencies"`
}

func CreateRootStep() *StepData {
	return &StepData{
		Name:                 RootStepName,
		ArgoApplication:      "",
		ArgoApplicationPath:  "",
		OtherDeploymentsPath: "",
		ClusterName:          "",
		Tests:                []TestData{},
		RemediationLimit:     0,
		RollbackLimit:        0,
		Dependencies:         []string{},
	}
}
