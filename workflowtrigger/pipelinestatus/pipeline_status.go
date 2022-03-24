package pipelinestatus

import (
	"github.com/greenopsinc/util/auditlog"
	"gitlab.com/c0b/go-ordered-json"
)

// TODO: should be replaced with PipelineSpec from CRD
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

func MarshallPipelineStatus(status PipelineStatus) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()

	var failedStepsList []*ordered.OrderedMap
	failedStepsList = make([]*ordered.OrderedMap, 0)
	for _, val := range status.FailedSteps {
		step := MarshallFailedStep(val)
		failedStepsList = append(failedStepsList, step)
	}

	mapObj.Set("progressingSteps", status.ProgressingSteps)
	mapObj.Set("stable", status.Stable)
	mapObj.Set("complete", status.Complete)
	mapObj.Set("cancelled", status.Cancelled)
	mapObj.Set("failedSteps", failedStepsList)
	return mapObj
}
