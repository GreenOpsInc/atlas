package serializerutil

import "encoding/json"

const (
	//Subtypes
	DeploymentLogType        string = "deployment"
	RemediationLogType       string = "stateremediation"
	PipelineTriggerEventType string = "pipelinetrigger"
	GitCredOpenType          string = "open"
	GitCredMachineUserType   string = "machineuser"
	GitCredTokenType         string = "oauth"
	//Types
	TeamSchemaType    string = "teamschema"
	ClusterSchemaType string = "clusterschema"
	LogType           string = "log"
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
