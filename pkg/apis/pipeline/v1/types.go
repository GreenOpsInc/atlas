package v1

import (
	"github.com/greenopsinc/util/auditlog"
	"github.com/greenopsinc/util/git"
	"gitlab.com/c0b/go-ordered-json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// TODO: generate yaml file

// Pipeline is a definition of Pipeline resource.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:noStatus
// +groupName=pipeline
// +kubebuilder:resource:path=pipelines,shortName=pipeline;pipelines
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Spec              PipelineSpec   `json:"spec" protobuf:"bytes,2,opt,name=spec"`
	Status            PipelineStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// PipelineSpec represents desired pipeline state
type PipelineSpec struct {
	PipelineName string `json:"pipelineName"`
	// TODO: move schema to this file
	GitRepoSchema git.GitRepoSchema `json:"gitRepoSchema"`
}

func UnmarshallPipelineSchema(m map[string]interface{}) PipelineSpec {
	gitRepoSchema := git.UnmarshallGitRepoSchema(m["gitRepoSchema"].(map[string]interface{}))
	return PipelineSpec{
		PipelineName:  m["pipelineName"].(string),
		GitRepoSchema: gitRepoSchema,
	}
}

func MarshalPipelineSchema(schema PipelineSpec) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("pipelineName", schema.PipelineName)

	repo := git.MarshalGitRepoSchema(schema.GitRepoSchema)
	mapObj.Set("gitRepoSchema", repo)

	return mapObj
}

type PipelineStatus struct {
	ProgressingSteps []string     `json:"progressingSteps"`
	Stable           bool         `json:"stable"`
	Complete         bool         `json:"complete"`
	Cancelled        bool         `json:"cancelled"`
	FailedSteps      []FailedStep `json:"failedSteps"`
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

// PipelineWatchEvent contains information about pipeline change
type PipelineWatchEvent struct {
	Type     watch.EventType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=k8s.io/apimachinery/pkg/watch.EventType"`
	Pipeline Pipeline        `json:"pipeline" protobuf:"bytes,2,opt,name=pipeline"`
}

// PipelineList is list of Pipeline resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Pipeline `json:"items" protobuf:"bytes,2,rep,name=items"`
}
