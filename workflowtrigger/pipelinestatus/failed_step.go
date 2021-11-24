package pipelinestatus

import (
	"encoding/json"
)

type FailedStep struct {
	Step             string `json:"step"`
	DeploymentFailed bool   `json:"deploymentFailed"`
	BrokenTest       string `json:"brokenTest"`
	BrokenTestLog    string `json:"brokenTestLog"`
}

func MarshallFailedStep(failedStep FailedStep) map[string]interface{} {
	bytes, err := json.Marshal(failedStep)
	if err != nil {
		panic(err)
	}
	var mapObj map[string]interface{}
	err = json.Unmarshal(bytes, &mapObj)
	if err != nil {
		panic(err)
	}
	return mapObj
}
