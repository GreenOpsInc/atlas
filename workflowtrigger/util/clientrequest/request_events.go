package clientrequest

import "greenops.io/workflowtrigger/util/serializerutil"

type NotificationRequestEvent interface {
	GetEvent() string
}

type ClientEventMetadata struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	PipelineUvn  string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
	RequestId    string `json:"requestId"`
}

//*****
//*****
//Notifications Events: Need to have a request ID
//*****
//*****

// ClientMarkNoDeployRequest -----
type ClientMarkNoDeployRequest struct {
	ClientEventMetadata
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
	Apply       bool   `json:"apply"`
}

func (r ClientMarkNoDeployRequest) GetEvent() string {
	return serializerutil.NoDeployNotificationEventType
}

func MarshalEvent(event NotificationRequestEvent) map[string]interface{} {
	switch event.(type) {
	case ClientMarkNoDeployRequest:
		mapObj := serializerutil.GetMapFromStruct(event)
		mapObj["type"] = serializerutil.NoDeployNotificationEventType
		return mapObj
	default:
		panic("Matching notification event type not found")
	}
}
