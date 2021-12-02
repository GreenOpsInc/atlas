package auditlog

import (
	"encoding/json"
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

func Marshal(log Log) map[string]interface{} {
	switch log.(type) {
	case *RemediationLog:
		mapObj := serializerutil.GetMapFromStruct(log)
		mapObj["type"] = serializerutil.RemediationLogType
		return mapObj
	default: //Deployment log
		mapObj := serializerutil.GetMapFromStruct(log)
		mapObj["type"] = serializerutil.DeploymentLogType
		return mapObj
	}
}

func MarshalList(logList []Log) []interface{} {
	var mapObj []interface{}
	for _, val := range logList {
		mapObj = append(mapObj, Marshal(val))
	}
	return mapObj
}
