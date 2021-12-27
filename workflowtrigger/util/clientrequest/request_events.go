package clientrequest

import (
	"gitlab.com/c0b/go-ordered-json"
	"greenops.io/workflowtrigger/util/serializerutil"
)

type ClientRequestPacket struct {
	RetryCount    int                `json:"retryCount"`
	Namespace     string             `json:"namespace"`
	ClientRequest ClientRequestEvent `json:"clientRequest"`
}

func MarshalRequestPacket(packet ClientRequestPacket) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("retryCount", packet.RetryCount)
	mapObj.Set("namespace", packet.Namespace)
	mapObj.Set("clientRequest", MarshalRequestEvent(packet.ClientRequest))
	return mapObj
}

type ClientRequestEvent interface {
	GetRequestEvent() string
}

type NotificationRequestEvent interface {
	GetNotificationEvent() string
	GetRequestId() string
}

type ClientRequestEventMetadata struct {
	OrgName      string `json:"orgName"`
	TeamName     string `json:"teamName"`
	PipelineName string `json:"pipelineName"`
	PipelineUvn  string `json:"pipelineUvn"`
	StepName     string `json:"stepName"`
}

// ClientDeployRequest -----
type ClientDeployRequest struct {
	ClientRequestEventMetadata
	ResponseEventType string `json:"responseEventType"`
	DeployType        string `json:"deployType"`
	RevisionHash      string `json:"revisionHash"`
	Payload           string `json:"payload"`
}

func (r ClientDeployRequest) GetRequestEvent() string {
	return serializerutil.DeployRequestEventType
}

func MarshalRequestEvent(event ClientRequestEvent) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()

	switch event.(type) {
	case ClientDeployRequest:
		mapObj.Set("orgName", event.(*ClientDeployRequest).OrgName)
		mapObj.Set("teamName", event.(*ClientDeployRequest).TeamName)
		mapObj.Set("pipelineName", event.(*ClientDeployRequest).PipelineName)
		mapObj.Set("pipelineUvn", event.(*ClientDeployRequest).PipelineUvn)
		mapObj.Set("stepName", event.(*ClientDeployRequest).StepName)
		mapObj.Set("responseEventType", event.(*ClientDeployRequest).ResponseEventType)
		mapObj.Set("deployType", event.(*ClientDeployRequest).DeployType)
		mapObj.Set("revisionHash", event.(*ClientDeployRequest).RevisionHash)
		mapObj.Set("payload", event.(*ClientDeployRequest).Payload)
		mapObj.Set("type", serializerutil.DeployRequestEventType)
	default:
		panic("Matching notification event type not found")
	}

	return mapObj
}

//*****
//*****
//Notifications Events: Need to have a request ID
//*****
//*****

type ClientNotificationEventMetadata struct {
	ClientRequestEventMetadata
	RequestId string `json:"requestId"`
}

// ClientMarkNoDeployRequest -----
type ClientMarkNoDeployRequest struct {
	ClientNotificationEventMetadata
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
	Apply       bool   `json:"apply"`
}

func (r ClientMarkNoDeployRequest) GetNotificationEvent() string {
	return serializerutil.NoDeployNotificationEventType
}

func (r ClientMarkNoDeployRequest) GetRequestId() string {
	return r.RequestId
}

func MarshalNotificationEvent(event NotificationRequestEvent) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()

	switch event.(type) {
	case ClientMarkNoDeployRequest:
		mapObj.Set("orgName", event.(*ClientMarkNoDeployRequest).OrgName)
		mapObj.Set("teamName", event.(*ClientMarkNoDeployRequest).TeamName)
		mapObj.Set("pipelineName", event.(*ClientMarkNoDeployRequest).PipelineName)
		mapObj.Set("pipelineUvn", event.(*ClientMarkNoDeployRequest).PipelineUvn)
		mapObj.Set("stepName", event.(*ClientMarkNoDeployRequest).StepName)
		mapObj.Set("requestId", event.(*ClientMarkNoDeployRequest).RequestId)
		mapObj.Set("clusterName", event.(*ClientMarkNoDeployRequest).ClusterName)
		mapObj.Set("namespace", event.(*ClientMarkNoDeployRequest).Namespace)
		mapObj.Set("apply", event.(*ClientMarkNoDeployRequest).Apply)
		mapObj.Set("type", serializerutil.NoDeployNotificationEventType)
	default:
		panic("Matching notification event type not found")
	}

	return mapObj
}
