package pipelinestatus

import "gitlab.com/c0b/go-ordered-json"

type FailedStep struct {
	Step             string `json:"step"`
	DeploymentFailed bool   `json:"deploymentFailed"`
	BrokenTest       string `json:"brokenTest"`
	BrokenTestLog    string `json:"brokenTestLog"`
}

func MarshallFailedStep(failedStep FailedStep) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("step", failedStep.Step)
	mapObj.Set("deploymentFailed", failedStep.DeploymentFailed)
	mapObj.Set("brokenTest", failedStep.BrokenTest)
	mapObj.Set("brokenTestLog", failedStep.BrokenTestLog)
	return mapObj
}
