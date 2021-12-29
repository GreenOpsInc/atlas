package auditlog

import (
	"encoding/json"

	"gitlab.com/c0b/go-ordered-json"
	"greenops.io/workflowtrigger/util/serializerutil"
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
	GetUniqueVersionInstance() int
	GetStatus() LogStatus
	SetStatus(status LogStatus)
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

func (d *DeploymentLog) GetPipelineUniqueVersionNumber() string {
	return d.PipelineUniqueVersionNumber
}
func (d *DeploymentLog) GetUniqueVersionInstance() int {
	return d.UniqueVersionInstance
}
func (d *DeploymentLog) GetStatus() LogStatus {
	return d.Status
}
func (d *DeploymentLog) SetStatus(status LogStatus) {
	d.Status = status
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
func (d *RemediationLog) SetStatus(status LogStatus) {
	d.RemediationStatus = status
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
