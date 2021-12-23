package serializerutil

import "encoding/json"

const (
	//Subtypes
	DeploymentLogType             string = "deployment"
	RemediationLogType            string = "stateremediation"
	PipelineTriggerEventType      string = "pipelinetrigger"
	NoDeployNotificationEventType string = "mark_no_deploy"
	DeployRequestEventType        string = "deploy"
	GitCredOpenType               string = "open"
	GitCredMachineUserType        string = "machineuser"
	GitCredTokenType              string = "oauth"
	//Types
	NotificationType  string = "notification"
	TeamSchemaType    string = "teamschema"
	ClusterSchemaType string = "clusterschema"
	LogType           string = "log"
	PipelineInfoType  string = "pipelineinfo"
	StringListType    string = "stringlist"
)

func GetMapFromStruct(object interface{}) map[string]interface{} {
	bytes, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	var returnValue map[string]interface{}
	err = json.Unmarshal(bytes, &returnValue)
	if err != nil {
		panic(err)
	}
	return returnValue
}
