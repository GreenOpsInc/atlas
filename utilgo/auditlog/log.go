package auditlog

import (
	"encoding/json"
	"log"

	"github.com/greenopsinc/util/serializerutil"
	"gitlab.com/c0b/go-ordered-json"
)

type LogStatus string

const (
	Success     LogStatus = "SUCCESS"
	Progressing LogStatus = "PROGRESSING"
	Failure     LogStatus = "FAILURE"
	Cancelled   LogStatus = "CANCELLED"
)

type Log interface {
	GetPipelineUniqueVersionNumber() string
	GetRollbackUniqueVersionNumber() string
	GetUniqueVersionInstance() int
	GetStatus() LogStatus
	IsDeploymentComplete() bool
	GetArgoApplicationName() string
	GetArgoRevisionHash() string
	GetGitCommitVersion() string
	GetBrokenTest() string
	GetBrokenTestLog() string
	SetPipelineUniqueVersionNumber(number string)
	SetRollbackUniqueVersionNumber(number string)
	SetUniqueVersionInstance(version int)
	SetStatus(status LogStatus)
	SetDeploymentComplete(complete bool)
	SetArgoApplicationName(name string)
	SetArgoRevisionHash(hash string)
	SetGitCommitVersion(version string)
	SetBrokenTest(test string)
	SetBrokenTestLog(testLog string)
}

type DeploymentLog struct {
	PipelineUniqueVersionNumber string    `json:"pipelineUniqueVersionNumber"`
	RollbackUniqueVersionNumber string    `json:"rollbackUniqueVersionNumber"`
	UniqueVersionInstance       int       `json:"uniqueVersionInstance"`
	Status                      LogStatus `json:"status"`
	DeploymentComplete          bool      `json:"deploymentComplete"`
	ArgoApplicationName         string    `json:"argoApplicationName"`
	ArgoRevisionHash            string    `json:"argoRevisionHash"`
	GitCommitVersion            string    `json:"gitCommitVersion"`
	BrokenTest                  string    `json:"brokenTest"`
	BrokenTestLog               string    `json:"brokenTestLog"`
}

func (d *DeploymentLog) GetPipelineUniqueVersionNumber() string {
	return d.PipelineUniqueVersionNumber
}

func (d *DeploymentLog) GetRollbackUniqueVersionNumber() string {
	return d.RollbackUniqueVersionNumber
}

func (d *DeploymentLog) GetUniqueVersionInstance() int {
	return d.UniqueVersionInstance
}

func (d *DeploymentLog) GetStatus() LogStatus {
	return d.Status
}

func (d *DeploymentLog) IsDeploymentComplete() bool {
	return d.DeploymentComplete
}

func (d *DeploymentLog) GetArgoApplicationName() string {
	return d.ArgoApplicationName
}

func (d *DeploymentLog) GetArgoRevisionHash() string {
	return d.ArgoRevisionHash
}

func (d *DeploymentLog) GetGitCommitVersion() string {
	return d.GitCommitVersion
}

func (d *DeploymentLog) GetBrokenTest() string {
	return d.BrokenTest
}

func (d *DeploymentLog) GetBrokenTestLog() string {
	return d.BrokenTestLog
}

func (d *DeploymentLog) SetPipelineUniqueVersionNumber(number string) {
	d.PipelineUniqueVersionNumber = number
}

func (d *DeploymentLog) SetRollbackUniqueVersionNumber(number string) {
	d.RollbackUniqueVersionNumber = number
}

func (d *DeploymentLog) SetUniqueVersionInstance(version int) {
	d.UniqueVersionInstance = version
}

func (d *DeploymentLog) SetStatus(status LogStatus) {
	d.Status = status
}

func (d *DeploymentLog) SetDeploymentComplete(complete bool) {
	d.DeploymentComplete = complete
}

func (d *DeploymentLog) SetArgoApplicationName(name string) {
	d.ArgoApplicationName = name
}

func (d *DeploymentLog) SetArgoRevisionHash(hash string) {
	d.ArgoRevisionHash = hash
}

func (d *DeploymentLog) SetGitCommitVersion(version string) {
	d.GitCommitVersion = version
}

func (d *DeploymentLog) SetBrokenTest(test string) {
	d.BrokenTest = test
}

func (d *DeploymentLog) SetBrokenTestLog(testLog string) {
	d.BrokenTestLog = testLog
}

func InitBlankDeploymentLog(pipelineUniqueVersionNumber string, status LogStatus, deploymentComplete bool, argoRevisionHash string, gitCommitVersion string) Log {
	d := DeploymentLog{}
	d.PipelineUniqueVersionNumber = pipelineUniqueVersionNumber
	d.RollbackUniqueVersionNumber = ""
	d.UniqueVersionInstance = 0
	d.Status = status
	d.DeploymentComplete = deploymentComplete
	d.ArgoApplicationName = ""
	d.ArgoRevisionHash = argoRevisionHash
	d.GitCommitVersion = gitCommitVersion
	d.BrokenTest = ""
	d.BrokenTestLog = ""
	return &d
}

type RemediationLog struct {
	PipelineUniqueVersionNumber string    `json:"pipelineUniqueVersionNumber"`
	UniqueVersionInstance       int       `json:"uniqueVersionInstance"`
	UnhealthyResources          []string  `json:"unhealthyResources"`
	RemediationStatus           LogStatus `json:"remediationStatus"`
}

func InitBlankRemediationLog(pipelineUniqueVersionNumber string, uniqueVersionInstance int, unhealthyResources []string) Log {
	d := RemediationLog{}
	d.PipelineUniqueVersionNumber = pipelineUniqueVersionNumber
	d.UniqueVersionInstance = uniqueVersionInstance
	d.UnhealthyResources = unhealthyResources
	d.RemediationStatus = Progressing
	return &d
}

func (d *RemediationLog) GetPipelineUniqueVersionNumber() string {
	return d.PipelineUniqueVersionNumber
}

func (d *RemediationLog) GetUniqueVersionInstance() int {
	return d.UniqueVersionInstance
}

func (d *RemediationLog) GetStatus() LogStatus {
	return d.RemediationStatus
}

func (d *RemediationLog) GetRollbackUniqueVersionNumber() string {
	log.Println("RollbackUniqueVersionNumber field is missing in RemediationLog, returning empty string")
	return ""
}

func (d *RemediationLog) IsDeploymentComplete() bool {
	log.Println("DeploymentComplete field is missing in RemediationLog, returning false")
	return false
}

func (d *RemediationLog) GetArgoApplicationName() string {
	log.Println("ArgoApplicationName field is missing in RemediationLog, returning empty string")
	return ""
}

func (d *RemediationLog) GetArgoRevisionHash() string {
	log.Println("ArgoRevisionHash field is missing in RemediationLog, returning empty string")
	return ""
}

func (d *RemediationLog) GetGitCommitVersion() string {
	log.Println("GitCommitVersion field is missing in RemediationLog, returning empty string")
	return ""
}

func (d *RemediationLog) GetBrokenTest() string {
	log.Println("BrokenTest field is missing in RemediationLog, returning empty string")
	return ""
}

func (d *RemediationLog) GetBrokenTestLog() string {
	log.Println("BrokenTestLog field is missing in RemediationLog, returning empty string")
	return ""
}

func (d *RemediationLog) GetUnhealthyResources() []string {
	return d.UnhealthyResources
}

func (d *RemediationLog) SetPipelineUniqueVersionNumber(number string) {
	d.PipelineUniqueVersionNumber = number
}

func (d *RemediationLog) SetRollbackUniqueVersionNumber(_ string) {
	log.Println("RollbackUniqueVersionNumber field is missing in RemediationLog")
}

func (d *RemediationLog) SetUniqueVersionInstance(version int) {
	d.UniqueVersionInstance = version
}

func (d *RemediationLog) SetStatus(status LogStatus) {
	d.RemediationStatus = status
}

func (d *RemediationLog) SetDeploymentComplete(_ bool) {
	log.Println("DeploymentComplete field is missing in RemediationLog")
}

func (d *RemediationLog) SetArgoApplicationName(_ string) {
	log.Println("ArgoApplicationName field is missing in RemediationLog")
}

func (d *RemediationLog) SetArgoRevisionHash(_ string) {
	log.Println("ArgoRevisionHash field is missing in RemediationLog")
}

func (d *RemediationLog) SetGitCommitVersion(_ string) {
	log.Println("GitCommitVersion field is missing in RemediationLog")
}

func (d *RemediationLog) SetBrokenTest(_ string) {
	log.Println("BrokenTest field is missing in RemediationLog")
}

func (d *RemediationLog) SetBrokenTestLog(_ string) {
	log.Println("BrokenTestLog field is missing in RemediationLog")
}

func (d *RemediationLog) SetUnhealthyResources(resources []string) {
	d.UnhealthyResources = resources
}

func Unmarshall(m map[string]interface{}) Log {
	logBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	if m["type"] == serializerutil.DeploymentLogType {
		var deploymentLog DeploymentLog
		_ = json.Unmarshal(logBytes, &deploymentLog)
		return &deploymentLog
	} else {
		var remediationLog RemediationLog
		_ = json.Unmarshal(logBytes, &remediationLog)
		return &remediationLog
	}
}

func UnmarshallString(str string) Log {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return Unmarshall(m)
}

func Marshal(log Log) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	switch log.(type) {
	case *RemediationLog:
		remLog := log.(*RemediationLog)
		mapObj.Set("type", serializerutil.RemediationLogType)
		mapObj.Set("pipelineUniqueVersionNumber", remLog.PipelineUniqueVersionNumber)
		mapObj.Set("uniqueVersionInstance", remLog.UniqueVersionInstance)
		mapObj.Set("unhealthyResources", remLog.UnhealthyResources)
		mapObj.Set("remediationStatus", remLog.RemediationStatus)
	default: //Deployment log
		depLog := log.(*DeploymentLog)
		mapObj.Set("type", serializerutil.DeploymentLogType)
		mapObj.Set("pipelineUniqueVersionNumber", depLog.PipelineUniqueVersionNumber)
		mapObj.Set("rollbackUniqueVersionNumber", depLog.RollbackUniqueVersionNumber)
		mapObj.Set("uniqueVersionInstance", depLog.UniqueVersionInstance)
		mapObj.Set("status", depLog.Status)
		mapObj.Set("deploymentComplete", depLog.DeploymentComplete)
		mapObj.Set("argoApplicationName", depLog.ArgoApplicationName)
		mapObj.Set("argoRevisionHash", depLog.ArgoRevisionHash)
		mapObj.Set("gitCommitVersion", depLog.GitCommitVersion)
		mapObj.Set("brokenTest", depLog.BrokenTest)
		mapObj.Set("brokenTestLog", depLog.BrokenTestLog)
	}

	return mapObj
}

func MarshalList(logList []Log) []*ordered.OrderedMap {
	var mapArr []*ordered.OrderedMap
	for _, val := range logList {
		mapArr = append(mapArr, Marshal(val))
	}
	return mapArr
}
