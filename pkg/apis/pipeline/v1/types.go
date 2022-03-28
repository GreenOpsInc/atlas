package v1

import (
	"encoding/json"
	"strings"

	"github.com/greenopsinc/util/auditlog"
	"github.com/greenopsinc/util/serializerutil"
	"gitlab.com/c0b/go-ordered-json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	hiddenInfo         string = "Hidden cred info"
	secureGitUrlPrefix string = "https://"
)

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
	PipelineName  string        `json:"pipelineName"`
	GitRepoSchema GitRepoSchema `json:"gitRepoSchema"`
}

func UnmarshallPipelineSchema(m map[string]interface{}) PipelineSpec {
	gitRepoSchema := UnmarshallGitRepoSchema(m["gitRepoSchema"].(map[string]interface{}))
	return PipelineSpec{
		PipelineName:  m["pipelineName"].(string),
		GitRepoSchema: gitRepoSchema,
	}
}

func MarshalPipelineSchema(schema PipelineSpec) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("pipelineName", schema.PipelineName)

	repo := MarshalGitRepoSchema(schema.GitRepoSchema)
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
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`
	Items           []Pipeline `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type GitRepoSchemaInfo struct {
	GitRepoUrl string `json:"gitRepoUrl"`
	PathToRoot string `json:"pathToRoot"`
}

func (g *GitRepoSchemaInfo) GetGitRepo() string {
	return g.GitRepoUrl
}

func (g *GitRepoSchemaInfo) GetPathToRoot() string {
	return g.PathToRoot
}

type GitRepoSchema struct {
	GitRepo    string  `json:"gitRepo"`
	PathToRoot string  `json:"pathToRoot"`
	GitCred    GitCred `json:"gitCred"`
}

func New(gitRepo string, pathToRoot string, gitCred GitCred) GitRepoSchema {
	return GitRepoSchema{
		GitRepo:    gitRepo,
		PathToRoot: pathToRoot,
		GitCred:    gitCred,
	}
}

func (g *GitRepoSchema) SetGitRepo(gitRepo string) {
	g.GitRepo = gitRepo
}

func (g *GitRepoSchema) SetPathToRoot(pathToRoot string) {
	g.PathToRoot = pathToRoot
}

func (g *GitRepoSchema) SetGitCred(gitCred GitCred) {
	g.GitCred = gitCred
}

func (g *GitRepoSchema) GetGitRepo() string {
	return g.GitRepo
}

func (g *GitRepoSchema) GetPathToRoot() string {
	return g.PathToRoot
}

func (g *GitRepoSchema) GetGitCred() GitCred {
	return g.GitCred
}

func (g *GitRepoSchema) ContentsEqual(schema GitRepoSchema) bool {
	return g.GetGitRepo() == schema.GetGitRepo() && g.GetPathToRoot() == schema.GetPathToRoot()
}

func UnmarshallGitRepoSchema(m map[string]interface{}) GitRepoSchema {
	gitCred := UnmarshallGitCred(m["gitCred"].(map[string]interface{}))
	return GitRepoSchema{
		GitRepo:    m["gitRepo"].(string),
		PathToRoot: m["pathToRoot"].(string),
		GitCred:    gitCred,
	}
}

func UnmarshallGitRepoSchemaString(str string) GitRepoSchema {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallGitRepoSchema(m)
}

func UnmarshallGitRepoSchemaInfo(m map[string]interface{}) GitRepoSchemaInfo {
	return GitRepoSchemaInfo{
		GitRepoUrl: m["gitRepoUrl"].(string),
		PathToRoot: m["pathToRoot"].(string),
	}
}

func UnmarshallGitRepoSchemaInfoString(str string) GitRepoSchemaInfo {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallGitRepoSchemaInfo(m)
}

func MarshalGitRepoSchema(schema GitRepoSchema) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("gitRepo", schema.GitRepo)
	mapObj.Set("pathToRoot", schema.PathToRoot)

	cred := MarshalGitCred(schema.GetGitCred())
	mapObj.Set("gitCred", cred)

	return mapObj
}

type GitCred interface {
	Hide()
	DeepCopyGitCred() GitCred
}

type GitCredAccessible interface {
	Hide()
	ConvertGitCredToString(gitRepoLink string) string
}

// GitCredOpen Setting up GitCredOpen
type GitCredOpen struct{}

func (g *GitCredOpen) Hide() {}

func (g *GitCredOpen) ConvertGitCredToString(gitRepoLink string) string {
	return gitRepoLink
}

func (g *GitCredOpen) DeepCopyGitCred() GitCred {
	return g.DeepCopy()
}

// GitCredMachineUser Setting up GitCredMachineUser
type GitCredMachineUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (g *GitCredMachineUser) Hide() {
	g.Username = hiddenInfo
	g.Password = hiddenInfo
}

func (g *GitCredMachineUser) ConvertGitCredToString(gitRepoLink string) string {
	splitUrl := strings.Split(gitRepoLink, secureGitUrlPrefix)
	splitUrl[1] = g.Username + ":" + g.Password + "@" + splitUrl[1]
	return secureGitUrlPrefix + splitUrl[1]
}

func (g *GitCredMachineUser) DeepCopyGitCred() GitCred {
	return g.DeepCopy()
}

// GitCredToken Setting up GitCredToken
type GitCredToken struct {
	Token string `json:"token"`
}

func (g *GitCredToken) Hide() {
	g.Token = hiddenInfo
}

func (g *GitCredToken) ConvertGitCredToString(gitRepoLink string) string {
	splitUrl := strings.Split(gitRepoLink, secureGitUrlPrefix)
	splitUrl[1] = g.Token + "@" + splitUrl[1]
	return secureGitUrlPrefix + splitUrl[1]
}

func (g *GitCredToken) DeepCopyGitCred() GitCred {
	return g.DeepCopy()
}

func UnmarshallGitCred(m map[string]interface{}) GitCred {
	credType := m["type"]
	delete(m, "type")
	gitCredBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	switch credType {
	case serializerutil.GitCredMachineUserType:
		var gitCredMachineUser GitCredMachineUser
		_ = json.Unmarshal(gitCredBytes, &gitCredMachineUser)
		return &gitCredMachineUser
	case serializerutil.GitCredTokenType:
		var gitCredToken GitCredToken
		_ = json.Unmarshal(gitCredBytes, &gitCredToken)
		return &gitCredToken
	default:
		var gitCredOpen GitCredOpen
		_ = json.Unmarshal(gitCredBytes, &gitCredOpen)
		return &gitCredOpen
	}
}

func UnmarshallGitCredString(str string) GitCred {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallGitCred(m)
}

func MarshalGitCred(gitCred GitCred) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()

	switch gitCred.(type) {
	case *GitCredMachineUser:
		mapObj.Set("username", gitCred.(*GitCredMachineUser).Username)
		mapObj.Set("password", gitCred.(*GitCredMachineUser).Password)
		mapObj.Set("type", serializerutil.GitCredMachineUserType)
	case *GitCredToken:
		mapObj.Set("token", gitCred.(*GitCredToken).Token)
		mapObj.Set("type", serializerutil.GitCredTokenType)
	default: //Open cred
		mapObj.Set("type", serializerutil.GitCredOpenType)
	}

	return mapObj
}

type GetFileRequest struct {
	GitRepoSchemaInfo GitRepoSchemaInfo `json:"gitRepoSchemaInfo"`
	Filename          string            `json:"filename"`
	GitCommitHash     string            `json:"gitCommitHash"`
}
