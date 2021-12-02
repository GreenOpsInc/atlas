package pipelinestatus

import (
	"encoding/json"
	"greenops.io/workflowtrigger/util/auditlog"
)

type PipelineStatus struct {
	ProgressingSteps []string     `json:"progressingSteps"`
	Stable           bool         `json:"stable"`
	Complete         bool         `json:"complete"`
	Cancelled        bool         `json:"cancelled"`
	FailedSteps      []FailedStep `json:"failedSteps"`
}

func New() PipelineStatus {
	return PipelineStatus{
		ProgressingSteps: make([]string, 0),
		Stable:           true,
		Complete:         true,
		Cancelled:        false,
		FailedSteps:      make([]FailedStep, 0),
	}
}

func (p *PipelineStatus) MarkCancelled() {
	p.Cancelled = true
	p.Complete = false
}

func (p *PipelineStatus) MarkIncomplete() {
	p.Complete = false
}

func (p *PipelineStatus) AddProgressingStep(stepStatus string) {
	p.ProgressingSteps = append(p.ProgressingSteps, stepStatus)
}

func (p *PipelineStatus) AddLatestLog(log auditlog.Log) {
	switch log.(type) {
	case *auditlog.DeploymentLog:
		deploymentLog := log.(*auditlog.DeploymentLog)
		if deploymentLog.GetStatus() == auditlog.Failure {
			p.Stable = false
		}
	case *auditlog.RemediationLog:
		if log.(*auditlog.RemediationLog).GetStatus() != auditlog.Failure {
			p.Stable = false
		}
	}
}

func (p *PipelineStatus) AddLatestDeploymentLog(log auditlog.DeploymentLog, step string) {
	if log.GetStatus() == auditlog.Failure {
		p.Complete = false
		p.AddFailedDeploymentLog(log, step)
	}
}

func (p *PipelineStatus) AddFailedDeploymentLog(log auditlog.DeploymentLog, step string) {
	p.FailedSteps = append(p.FailedSteps, FailedStep{
		Step:             step,
		DeploymentFailed: !log.DeploymentComplete,
		BrokenTest:       log.BrokenTest,
		BrokenTestLog:    log.BrokenTestLog,
	})
}

func MarshallPipelineStatus(status PipelineStatus) map[string]interface{} {
	bytes, err := json.Marshal(status)
	if err != nil {
		panic(err)
	}
	var mapObj map[string]interface{}
	err = json.Unmarshal(bytes, &mapObj)
	if err != nil {
		panic(err)
	}
	var failedStepsList []interface{}
	for _, val := range status.FailedSteps {
		failedStepsList = append(failedStepsList, MarshallFailedStep(val))
	}
	mapObj["failedSteps"] = failedStepsList
	return mapObj
}
