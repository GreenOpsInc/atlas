package serializerutil

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
