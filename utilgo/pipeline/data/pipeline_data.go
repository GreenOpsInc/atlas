package data

const (
	RootStepName = "ATLAS_ROOT_DATA"
)

type PipelineData struct {
	Name            string `json:"name"`
	ClusterName     string `json:"cluster_name"`
	Steps           []*StepData
	ArgoVersionLock bool                `json:"argo_version_lock"`
	StepParents     map[string][]string `json:"step_parents"`
	StepChildren    map[string][]string `json:"step_children"`
}

func NewPipelineData(name, clusterName string, argoVersionLock bool, stepData []*StepData) PipelineData {
	data := PipelineData{
		Name:            name,
		ClusterName:     clusterName,
		ArgoVersionLock: argoVersionLock,
		StepParents:     make(map[string][]string, 0),
		StepChildren:    make(map[string][]string, 0),
	}

	for _, step := range stepData {
		if step.ClusterName == "" {
			step.ClusterName = clusterName
		}
		if len(step.Dependencies) == 0 {
			parents := make([]string, 0)
			parents = append(parents, RootStepName)
			data.StepParents[step.Name] = parents
			step.Dependencies = append(step.Dependencies, RootStepName)
		} else {
			data.StepParents[step.Name] = step.Dependencies
		}

		for _, parentStep := range step.Dependencies {
			if _, ok := data.StepChildren[parentStep]; ok {
				list := data.StepChildren[parentStep]
				list = append(list, step.Name)
				data.StepChildren[parentStep] = list
			} else {
				children := make([]string, 0)
				children = append(children, step.Name)
				data.StepChildren[parentStep] = children
			}
		}
	}

	data.Steps = stepData
	return data
}

func (p *PipelineData) GetStep(stepName string) *StepData {
	for _, s := range p.Steps {
		if s.Name == stepName {
			return s
		}
	}
	return nil
}

func (p *PipelineData) GetAllSteps() []string {
	res := make([]string, len(p.Steps))
	for _, s := range p.Steps {
		res = append(res, s.Name)
	}
	return res
}

func (p *PipelineData) GetAllStepsOrdered() []string {
	orderedSteps := p.StepChildren[RootStepName]
	stepsMap := make(map[string]bool, 0)
	for _, s := range p.GetAllSteps() {
		stepsMap[s] = false
		for _, os := range orderedSteps {
			if os == s {
				stepsMap[s] = true
				break
			}
		}
	}

	for idx, _ := range orderedSteps {
		childrenList := p.StepChildren[orderedSteps[idx]]
		for _, childStep := range childrenList {
			parentsSeen := true
			for _, parentOfChildStep := range p.StepParents[childStep] {
				if _, ok := stepsMap[parentOfChildStep]; !ok {
					parentsSeen = false
					break
				}
			}

			if parentsSeen && !stepsMap[childStep] {
				orderedSteps = append(orderedSteps, childStep)
				stepsMap[childStep] = true
			}
		}
	}
	return orderedSteps
}

func (p *PipelineData) InitClusterNames() {
	for idx, val := range p.Steps {
		if val.ClusterName == "" {
			p.Steps[idx].ClusterName = p.ClusterName
		}
	}
}
