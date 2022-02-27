package serializerutil

import "encoding/json"

const (
	//Log subtypes
	DeploymentLogType  string = "deployment"
	RemediationLogType string = "stateremediation"
	//Event subtypes
	PipelineTriggerEventType string = "pipelinetrigger"
	//Client request subtypes
	ClientDeployRequestType                      string = "deploy"
	ClientDeleteByConfigRequestType              string = "del_config"
	ClientDeleteByGvkRequestType                 string = "del_gvk"
	ClientDeployAndWatchRequestType              string = "deploy_watch"
	ClientDeployArgoAppByNameRequestType         string = "deploy_namedargo"
	ClientDeployArgoAppByNameAndWatchRequestType string = "deploy_namedargo_watch"
	ClientRollbackAndWatchRequestType            string = "rollback"
	ClientSelectiveSyncRequestType               string = "sel_sync_watch"
	ClientMarkNoDeployRequestType                string = "mark_no_deploy"
	ClientLabelRequestType                       string = "label"
	ClientAggregateRequestType                   string = "aggregate"
	ClientDeleteByLabelRequestType               string = "del_label"
	//Git cred subtypes
	GitCredOpenType        string = "open"
	GitCredMachineUserType string = "machineuser"
	GitCredTokenType       string = "oauth"
	//Types
	NotificationType      string = "notification"
	TeamSchemaType        string = "teamschema"
	ClusterSchemaType     string = "clusterschema"
	LogType               string = "log"
	ClientRequestType     string = "clientrequest"
	ClientPacketType      string = "clientpacket"
	ClientRequestListType string = "clientrequestlist"
	PipelineInfoType      string = "pipelineinfo"
	StringListType        string = "stringlist"
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
